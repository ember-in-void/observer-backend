package google

import (
	"context"

	"steam-observer/internal/modules/auth/ports/out_ports"
)

type StubGoogleClient struct{}

func NewStubClient() out_ports.GoogleOAuthClient {
	return &StubGoogleClient{}
}

func (c *StubGoogleClient) ExchangeCode(ctx context.Context, code string) (*out_ports.GoogleTokens, error) {
	// Пока игнорируем code и возвращаем фиктивные токены
	return &out_ports.GoogleTokens{
		AccessToken:  "stub-access-token",
		RefreshToken: "stub-refresh-token",
		IDToken:      "stub-id-token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}
