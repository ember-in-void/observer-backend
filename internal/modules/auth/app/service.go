package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"steam-observer/internal/modules/auth/ports/in_ports"
	"steam-observer/internal/modules/auth/ports/out_ports"
	"steam-observer/internal/shared/config"
	"steam-observer/internal/shared/logger"
)

type AuthService interface {
	in_ports.AuthService
}

type authServiceImpl struct {
	cfg           config.GoogleOAuthConfig
	userRepo      out_ports.UserRepository
	oauthClient   out_ports.GoogleOAuthClient
	tokenProvider out_ports.TokenProvider
	stateStore    StateStore
	logger        logger.Logger
}

func NewAuthService(
	googleCfg config.GoogleOAuthConfig,
	userRepo out_ports.UserRepository,
	oauthClient out_ports.GoogleOAuthClient,
	tokenProvider out_ports.TokenProvider,
	stateStore StateStore,
	log logger.Logger,
) AuthService {
	return &authServiceImpl{
		cfg:           googleCfg,
		userRepo:      userRepo,
		oauthClient:   oauthClient,
		tokenProvider: tokenProvider,
		stateStore:    stateStore,
		logger:        log,
	}
}

// BeginGoogleLogin - начинает OAuth flow с CSRF protection
func (s *authServiceImpl) BeginGoogleLogin(ctx context.Context, redirectAfterLogin string) (string, error) {
	// ========================================
	// 1. Генерируем secure random state
	// ========================================

	// generateSecureState() использует crypto/rand для генерации 256-bit random string
	// Этот state будет:
	// 1. Сохранён в backend (с redirect URL)
	// 2. Отправлен в Google OAuth URL
	// 3. Вернётся обратно в callback
	// 4. Проверен на соответствие → защита от CSRF
	state, err := generateSecureState()
	if err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}

	// ========================================
	// 2. Сохраняем state с metadata
	// ========================================

	// TTL 10 минут - баланс между:
	// - Security: короткий TTL → меньше окно для атаки
	// - UX: длинный TTL → пользователь не торопится
	//
	// Почему redirectAfterLogin важен:
	// - Пользователь был на /dashboard
	// - Не авторизован → редирект на /auth/google/login?redirect=/dashboard
	// - После OAuth flow → вернуть на /dashboard (а не на /)
	if err := s.stateStore.Save(ctx, state, redirectAfterLogin, 10*time.Minute); err != nil {
		return "", fmt.Errorf("save state: %w", err)
	}

	// ========================================
	// 3. Строим Google OAuth URL
	// ========================================

	// url.Values - тип map[string][]string для query parameters
	// Методы:
	// - Set(key, value) - устанавливает один value
	// - Add(key, value) - добавляет value (для multiple values)
	// - Get(key) - получает первый value
	// - Encode() - конвертирует в query string с URL encoding
	params := url.Values{}
	params.Set("client_id", s.cfg.ClientID)       // OAuth Client ID из Google Console
	params.Set("redirect_uri", s.cfg.RedirectURL) // Куда Google редиректит после login
	params.Set("response_type", "code")           // OAuth 2.0 Authorization Code flow
	params.Set("scope", "openid email profile")   // Запрашиваемые permissions
	params.Set("access_type", "offline")          // Для получения refresh_token (опционально)
	params.Set("state", state)                    // ✅ CSRF protection token

	// url.URL - структура для безопасного построения URL
	// Автоматически экранирует специальные символы в query parameters
	u := url.URL{
		Scheme:   "https",
		Host:     "accounts.google.com",
		Path:     "/o/oauth2/v2/auth",
		RawQuery: params.Encode(), // Encode() делает URL encoding
	}

	// Результат:
	// https://accounts.google.com/o/oauth2/v2/auth?
	//   client_id=xxx&
	//   redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Fgoogle%2Fcallback&
	//   response_type=code&
	//   scope=openid+email+profile&
	//   access_type=offline&
	//   state=xYz123...
	return u.String(), nil
}

// CompleteGoogleLogin - завершает OAuth flow, создаёт/обновляет пользователя
func (s *authServiceImpl) CompleteGoogleLogin(ctx context.Context, code, state string) (string, error) {
	// ========================================
	// 0. Валидируем state (CSRF protection)
	// ========================================

	// Get проверяет state и удаляет его (one-time use)
	// Если state не найден или истёк - это потенциальная CSRF атака
	_, err := s.stateStore.Get(ctx, state)
	if err != nil {
		s.logger.Warnf("invalid state: %v", err)
		return "", fmt.Errorf("invalid state: %w", err)
	}

	// ========================================
	// 1. Обменять authorization code на токены
	// ========================================

	// ExchangeCode делает POST запрос к https://oauth2.googleapis.com/token
	// Параметры: code, client_id, client_secret, redirect_uri, grant_type
	tokens, err := s.oauthClient.ExchangeCode(ctx, code)
	if err != nil {
		s.logger.Errorf("failed to exchange code: %v", err)
		return "", fmt.Errorf("exchange authorization code: %w", err)
	}

	// ========================================
	// 2. Получить информацию о пользователе от Google
	// ========================================

	// GetUserInfo делает GET запрос к https://www.googleapis.com/oauth2/v2/userinfo
	// Header: Authorization: Bearer {access_token}
	// Возвращает: sub (Google ID), email, name, picture и т.д.
	googleUserInfo, err := s.oauthClient.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		s.logger.Errorf("failed to get user info: %v", err)
		return "", fmt.Errorf("get google user info: %w", err)
	}

	// ========================================
	// 3. Найти или создать пользователя в БД
	// ========================================

	// Пытаемся найти существующего пользователя
	user, err := s.userRepo.FindByGoogleID(ctx, googleUserInfo.Sub)
	if err != nil {
		// errors.Is проверяет ошибку и все wrapped ошибки
		// Работает благодаря %w в fmt.Errorf
		if errors.Is(err, out_ports.ErrNotFound) {
			// Пользователь не найден - создаём нового
			s.logger.Infof("creating new user with google_id=%s", googleUserInfo.Sub)

			// ToUser() конвертирует Google данные в domain.User
			user = googleUserInfo.ToUser()

			// Create сохраняет в БД и обновляет user.ID, user.CreatedAt
			if err := s.userRepo.Create(ctx, user); err != nil {
				s.logger.Errorf("failed to create user: %v", err)
				return "", fmt.Errorf("create user: %w", err)
			}

			s.logger.Infof("user created successfully, id=%s", user.ID)
		} else {
			// Другая ошибка (БД недоступна, timeout и т.д.)
			s.logger.Errorf("repository error: %v", err)
			return "", fmt.Errorf("find user by google_id: %w", err)
		}
	} else {
		// Пользователь найден - обновляем email если изменился
		s.logger.Infof("found existing user, id=%s", user.ID)

		if googleUserInfo.Email != "" && (user.Email == nil || *user.Email != googleUserInfo.Email) {
			s.logger.Infof("updating user email: %s -> %s",
				safeDeref(user.Email), googleUserInfo.Email)

			user.UpdateEmail(googleUserInfo.Email)

			if err := s.userRepo.Update(ctx, user); err != nil {
				// Не критично - можем продолжить login
				s.logger.Warnf("failed to update user email: %v", err)
			}
		}
	}

	// ========================================
	// 4. Сгенерировать JWT токен
	// ========================================

	// GenerateAccessToken создаёт JWT с claims:
	// - user_id: string
	// - email: *string
	// - exp: время истечения (now + TTL)
	// - iat: время создания
	// - iss: "steam-observer"
	token, err := s.tokenProvider.GenerateAccessToken(ctx, string(user.ID), user.Email)
	if err != nil {
		s.logger.Errorf("failed to generate token: %v", err)
		return "", fmt.Errorf("generate access token: %w", err)
	}

	s.logger.Infof("login successful, user_id=%s", user.ID)

	return token, nil
}

// safeDeref - безопасное разыменование указателя для логирования
func safeDeref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
