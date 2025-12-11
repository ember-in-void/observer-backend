// internal/modules/auth/adapters/http/handlers.go
package http

import (
	"net/http"

	"steam-observer/internal/modules/auth/ports/in_ports"
)

type Handler struct {
	authService in_ports.AuthService
}

func NewHandler(authService in_ports.AuthService) *Handler {
	return &Handler{authService: authService}
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	redirectAfter := r.URL.Query().Get("redirect")

	url, err := h.authService.BeginGoogleLogin(r.Context(), redirectAfter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot start google login"}`))
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("missing code"))
		return
	}

	// Пока просто показываем заглушку, чтобы видеть, что всё работает
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Google callback OK, code received"))
}
