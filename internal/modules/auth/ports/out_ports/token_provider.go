package out_ports

import (
	"context"
)

type TokenProvider interface {
	GenerateAccessToken(ctx context.Context, userID string, email *string) (string, error)
}
