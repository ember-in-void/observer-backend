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
		_, _ = w.Write([]byte(`{"error":"missing code"}`))
		return
	}

	token, err := h.authService.CompleteGoogleLogin(r.Context(), code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot complete google login"}`))
		return
	}

	// Вариант 1: отдать JSON
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"access_token":"` + token + `"}`))

	// Вариант 2 (позже): положить в httpOnly cookie и сделать redirect на фронт
}
