// internal/modules/auth/app/service.go
package app

import (
	"context"
	"net/url"

	"steam-observer/internal/modules/auth/ports/in_ports"
	"steam-observer/internal/shared/config"
)

type AuthService interface {
	in_ports.AuthService
}

type authServiceImpl struct {
	cfg config.GoogleOAuthConfig
}

func NewAuthService(googleCfg config.GoogleOAuthConfig) AuthService {
	return &authServiceImpl{
		cfg: googleCfg,
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
