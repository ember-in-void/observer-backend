// internal/modules/auth/adapters/in/google/client.go
package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"steam-observer/internal/modules/auth/domain"
	"steam-observer/internal/modules/auth/ports/out_ports"
	"steam-observer/internal/shared/config"
)

type client struct {
	cfg        config.GoogleOAuthConfig
	httpClient *http.Client
}

// NewClient - создаёт реальный Google OAuth клиент
func NewClient(cfg config.GoogleOAuthConfig) out_ports.GoogleOAuthClient {
	return &client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// tokenResponse - ответ от Google Token API
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
}

// ExchangeCode - обменивает authorization code на токены
func (c *client) ExchangeCode(ctx context.Context, code string) (*out_ports.OAuthTokens, error) {
	// Подготовка запроса к Google Token API
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", c.cfg.ClientID)
	data.Set("client_secret", c.cfg.ClientSecret)
	data.Set("redirect_uri", c.cfg.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = data.Encode()

	// Выполнение запроса
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google token api error: %s (status: %d)", body, resp.StatusCode)
	}

	// Парсинг ответа
	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	return &out_ports.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// GetUserInfo - получает информацию о пользователе от Google
func (c *client) GetUserInfo(ctx context.Context, accessToken string) (*domain.GoogleUserInfo, error) {
	// Запрос к Google UserInfo API (OpenID Connect endpoint)
	// v3 endpoint возвращает 'sub', v2 возвращает 'id'
	req, err := http.NewRequestWithContext(ctx, "GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google userinfo api error: %s (status: %d)", body, resp.StatusCode)
	}

	// Парсинг ответа
	var userInfo domain.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}

	// Валидация обязательных полей
	if userInfo.Sub == "" {
		return nil, errors.New("google user info missing 'sub' field")
	}

	return &userInfo, nil
}
