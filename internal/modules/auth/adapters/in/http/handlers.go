// internal/modules/auth/adapters/http/handlers.go
package http

import (
	"net/http"

	"steam-observer/internal/modules/auth/ports/in_ports"
	"steam-observer/internal/shared/logger"
)

type AuthHandler struct {
	authService in_ports.AuthService
	logger      logger.Logger
}

func NewAuthHandler(authService in_ports.AuthService, log logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      log,
	}
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	redirectAfter := r.URL.Query().Get("redirect")

	h.logger.Infof("starting google login, redirect_after=%s", redirectAfter)

	url, err := h.authService.BeginGoogleLogin(r.Context(), redirectAfter)
	if err != nil {
		h.logger.Errorf("cannot start google login: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot start google login"}`))
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		h.logger.Warn("google callback: missing code")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"missing code"}`))
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		h.logger.Warn("google callback: missing state")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"missing state"}`))
		return
	}

	h.logger.Info("processing google callback")

	token, redirectAfter, err := h.authService.CompleteGoogleLogin(r.Context(), code, state)
	if err != nil {
		h.logger.Errorf("cannot complete google login: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot complete google login"}`))
		return
	}

	h.logger.Info("google login completed successfully")

	// ==========================================
	// Формируем URL для редиректа на фронтенд
	// ==========================================
	// По умолчанию редиректим на корень фронтенда
	frontendURL := "http://localhost:3000"
	if redirectAfter != "" {
		// Если был передан конкретный путь (напр. /dashboard), можно добавить его
		// Но пока для простоты — всегда на корень с токеном
		// frontendURL = frontendURL + redirectAfter
	}

	targetURL := frontendURL + "/?token=" + token
	http.Redirect(w, r, targetURL, http.StatusFound)
}
