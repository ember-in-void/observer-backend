package out_ports

import (
	"context"

	"steam-observer/internal/modules/auth/domain"
)

type UserRepository interface {
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
}
