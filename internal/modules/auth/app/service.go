// internal/modules/auth/app/service.go
package app

import (
	"context"
	"errors"
	"net/url"

	"steam-observer/internal/modules/auth/domain"
	"steam-observer/internal/modules/auth/ports/in_ports"
	"steam-observer/internal/modules/auth/ports/out_ports"
	"steam-observer/internal/shared/config"
)

type AuthService interface {
	in_ports.AuthService
}

type authServiceImpl struct {
	cfg           config.GoogleOAuthConfig
	userRepo      out_ports.UserRepository
	oauthClient   out_ports.GoogleOAuthClient
	tokenProvider out_ports.TokenProvider
}

func NewAuthService(
	googleCfg config.GoogleOAuthConfig,
	userRepo out_ports.UserRepository,
	oauthClient out_ports.GoogleOAuthClient,
	tokenProvider out_ports.TokenProvider,
) AuthService {
	return &authServiceImpl{
		cfg:           googleCfg,
		userRepo:      userRepo,
		oauthClient:   oauthClient,
		tokenProvider: tokenProvider,
	}
}

func (s *authServiceImpl) BeginGoogleLogin(ctx context.Context, redirectAfterLogin string) (string, error) {
	// redirectAfterLogin пока можно игнорировать или добавить как state позже

	params := url.Values{}
	params.Set("client_id", s.cfg.ClientID)
	params.Set("redirect_uri", s.cfg.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("access_type", "offline")
	// state добавим позже

	u := url.URL{
		Scheme:   "https",
		Host:     "accounts.google.com",
		Path:     "/o/oauth2/v2/auth",
		RawQuery: params.Encode(),
	}

	return u.String(), nil
}

func (s *authServiceImpl) CompleteGoogleLogin(ctx context.Context, code string) (string, error) {
	// 1. Обмен code -> tokens у Google
	tokens, err := s.oauthClient.ExchangeCode(ctx, code)
	if err != nil {
		return "", err
	}

	// 2. Распарсить и провалидировать id_token (подпись, iss, aud, exp)
	// тут можно сделать внутри oauthClient или отдельным helper'ом
	claims, err := ParseAndValidateIDToken(tokens.IDToken, s.cfg.ClientID)
	if err != nil {
		return "", err
	}

	googleID := claims.Sub
	email := claims.Email

	// 3. Найти пользователя по google_id или email
	user, err := s.userRepo.FindByGoogleID(ctx, googleID)
	if err != nil {
		// если not found — создаём
		if errors.Is(err, out_ports.ErrNotFound) {
			user = &domain.User{
				Email:    &email,
				GoogleID: &googleID,
				// CreatedAt/UpdatedAt заполнятся в репозитории
			}
			if err := s.userRepo.Create(ctx, user); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		// опционально обновить email, если изменился
	}

	// 4. Сгенерировать свой JWT по user.ID
	token, err := s.tokenProvider.GenerateAccessToken(ctx, string(user.ID), user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}
