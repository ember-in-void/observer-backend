// internal/app/di.go
package app

import (
	"context"
	"log"

	"steam-observer/internal/modules/auth/adapters/in/google"
	authpg "steam-observer/internal/modules/auth/adapters/out/postgres"
	authapp "steam-observer/internal/modules/auth/app"
	"steam-observer/internal/shared/config"
	"steam-observer/internal/shared/db"
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
	oauthClient := google.NewStubClient()       // твой адаптер
	tokenProvider := token.NewProvider(cfg.JWT) // твой адаптер

	authService := authapp.NewAuthService(cfg.Google, userRepo, oauthClient, tokenProvider)

	return &Container{
		Config:      cfg,
		DB:          pg,
		AuthService: authService,
	}
}
