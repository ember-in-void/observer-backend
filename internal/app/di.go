// internal/app/di.go
package app

import (
	"context"
	"log"

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

	authService := authapp.NewAuthService(cfg.Google /* позже сюда добавим repo, tokenProvider, etc. */)

	return &Container{
		Config:      cfg,
		DB:          pg,
		AuthService: authService,
	}
}
