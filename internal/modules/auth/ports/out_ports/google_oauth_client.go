package out_ports

import (
	"context"

	"steam-observer/internal/modules/auth/domain"
)

// OAuthTokens - токены полученные от Google
type OAuthTokens struct {
	AccessToken  string // Токен для доступа к Google API
	RefreshToken string // Токен для обновления (опционально)
	IDToken      string // JWT токен с информацией о пользователе
	ExpiresIn    int    // Время жизни токена в секундах
}

// GoogleOAuthClient - интерфейс для работы с Google OAuth
type GoogleOAuthClient interface {
	// ExchangeCode - обменивает authorization code на токены
	ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error)

	// GetUserInfo - получает информацию о пользователе
	GetUserInfo(ctx context.Context, accessToken string) (*domain.GoogleUserInfo, error)
}
