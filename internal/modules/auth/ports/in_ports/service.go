package in_ports

import "context"

type AuthService interface {
	BeginGoogleLogin(ctx context.Context, redirectAfterLogin string) (string, error)
}
