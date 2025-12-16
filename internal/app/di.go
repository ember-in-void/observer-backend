package app

import (
	"context"
	"log"

	"steam-observer/internal/modules/auth/adapters/in/google"
	"steam-observer/internal/modules/auth/adapters/out/jwt_provider"
	authpg "steam-observer/internal/modules/auth/adapters/out/postgres"
	authapp "steam-observer/internal/modules/auth/app"
	"steam-observer/internal/modules/auth/ports/out_ports"
	marketapp "steam-observer/internal/modules/market/app"
	"steam-observer/internal/shared/config"
	"steam-observer/internal/shared/db"
)

type Container struct {
	Config        *config.Config
	DB            *db.Postgres
	AuthService   authapp.AuthService
	TokenProvider out_ports.TokenProvider
	MarketService marketapp.MarketService
}

func NewContainer(cfg *config.Config) *Container {
	ctx := context.Background()

	// 1. Infrastructure: Database
	pg, err := db.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	// 2. Adapters Out: Repositories & Clients
	userRepo := authpg.NewUserRepository(pg.Pool)
	oauthClient := google.NewClient(cfg.Google)
	tokenProvider := jwt_provider.NewJWTProvider(cfg.JWT)

	// 3. Application: State Store (NEW!)
	stateStore := authapp.NewInMemoryStateStore()

	// 4. Application: Services
	authService := authapp.NewAuthService(
		cfg.Google,
		userRepo,
		oauthClient,
		tokenProvider,
		stateStore,
	)

	marketService := marketapp.NewMarketService()

	return &Container{
		Config:        cfg,
		DB:            pg,
		AuthService:   authService,
		TokenProvider: tokenProvider,
		MarketService: marketService,
	}
}
