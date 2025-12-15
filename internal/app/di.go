// internal/app/di.go
package app

import (
	"context"
	"log"

	"steam-observer/internal/modules/auth/adapters/in/google"
	"steam-observer/internal/modules/auth/adapters/out/jwt_provider"
	"steam-observer/internal/shared/config"
	"steam-observer/internal/shared/db"

	authpg "steam-observer/internal/modules/auth/adapters/out/postgres"

	authapp "steam-observer/internal/modules/auth/app"
)

type Container struct {
	Config      *config.Config
	DB          *db.Postgres
	AuthService authapp.AuthService
}

func NewContainer(cfg *config.Config) *Container {
	ctx := context.Background()

	pg, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	userRepo := authpg.NewUserRepository(pg.Pool)
	oauthClient := google.NewStubClient()
	tokenProvider := jwt_provider.NewJWTProvider(cfg.JWT)

	authService := authapp.NewAuthService(cfg.Google, userRepo, oauthClient, tokenProvider)

	return &Container{
		Config:      cfg,
		DB:          pg,
		AuthService: authService,
	}
}
