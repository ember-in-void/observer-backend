package in_ports

import "context"

type GoogleUserInfo struct {
	ID    string
	Email string
	Name  string
}

type AuthResult struct {
	UserID       string
	Email        string
	IsNewUser    bool
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	BeginGoogleLogin(ctx context.Context, redirectAfterLogin string) (string, error)
	CompleteGoogleLogin(ctx context.Context, code string) (string, error)
}
