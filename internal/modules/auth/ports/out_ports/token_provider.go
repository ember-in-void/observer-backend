package out_ports

import "context"

// TokenClaims - данные извлечённые из токена
type TokenClaims struct {
	UserID string
	Email  *string
}

// TokenProvider - интерфейс для работы с JWT токенами
type TokenProvider interface {
	// GenerateAccessToken - генерирует access token
	GenerateAccessToken(ctx context.Context, userID string, email *string) (string, error)

	// ValidateToken - валидирует токен и возвращает claims
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)

	// ParseAccessToken - парсит токен и возвращает userID + email
	ParseAccessToken(ctx context.Context, token string) (string, *string, error)
}
