package out_ports

import "context"

type TokenProvider interface {
	GenerateAccessToken(ctx context.Context, userID string, email *string) (string, error)
	ParseAccessToken(ctx context.Context, token string) (userID string, email *string, err error)
}
