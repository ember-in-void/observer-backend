package out_ports

import "context"

type GoogleTokens struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int64
	TokenType    string
}

type GoogleOAuthClient interface {
	ExchangeCode(ctx context.Context, code string) (*GoogleTokens, error)
	// позже тут может быть метод для получения userinfo
}
