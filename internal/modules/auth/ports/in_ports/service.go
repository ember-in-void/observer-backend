package in_ports

import "context"

type AuthService interface {
	BeginGoogleLogin(ctx context.Context, redirectAfterLogin string) (string, error)
	CompleteGoogleLogin(ctx context.Context, code, state string) (string, error)
}
