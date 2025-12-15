package app

import (
	"net/http"

	authhttp "steam-observer/internal/modules/auth/adapters/in/http"
	"steam-observer/internal/shared/http/middleware"
)

func RegisterRoutes(mux *http.ServeMux, c *Container) {
	// Health check
	mux.HandleFunc("/health", handleHealth)

	// Auth routes
	authHandler := authhttp.NewHandler(c.AuthService)
	mux.HandleFunc("/auth/google/login", authHandler.GoogleLogin)
	mux.HandleFunc("/auth/google/callback", authHandler.GoogleCallback)

	authMW := middleware.Auth()
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
