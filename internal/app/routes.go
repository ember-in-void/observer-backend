package app

import (
	"net/http"

	authhttp "steam-observer/internal/modules/auth/adapters/in/http"
	dashboardhttp "steam-observer/internal/modules/dashboard/adapters/in/http"
	markethttp "steam-observer/internal/modules/market/adapters/in/http"
	"steam-observer/internal/shared/http/middleware"
)

func RegisterRoutes(mux *http.ServeMux, c *Container) {
	// Health check
	mux.HandleFunc("/health", handleHealth)

	// Auth routes
	authHandler := authhttp.NewAuthHandler(c.AuthService, c.Logger.WithField("handler", "auth"), c.Config.FrontendURL)
	mux.HandleFunc("/auth/google/login", authHandler.GoogleLogin)
	mux.HandleFunc("/auth/google/callback", authHandler.GoogleCallback)

	authMW := middleware.Auth(c.TokenProvider, c.Logger.WithField("middleware", "auth"))

	// Dashboard routes
	dashboardHandler := dashboardhttp.NewDashboardHandler(c.DashboardService)
	mux.Handle("/dashboard", authMW(http.HandlerFunc(dashboardHandler.GetDashboard)))

	// Protected routes
	marketHandler := markethttp.NewMarketHandler(c.MarketService)
	mux.Handle("/market/tracked", authMW(http.HandlerFunc(marketHandler.ListTracked)))

	c.Logger.Info("routes registered successfully")
}

func NewRoutesHandler(mux *http.ServeMux, corsOrigins []string) http.Handler {
	// Применяем CORS middleware ко всему
	return middleware.CORS(corsOrigins)(mux)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
