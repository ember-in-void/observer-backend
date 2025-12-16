// internal/modules/auth/adapters/http/handlers.go
package http

import (
	"net/http"

	"steam-observer/internal/modules/auth/ports/in_ports"
)

type AuthHandler struct {
	authService in_ports.AuthService
}

func NewAuthHandler(authService in_ports.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	redirectAfter := r.URL.Query().Get("redirect")

	url, err := h.authService.BeginGoogleLogin(r.Context(), redirectAfter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot start google login"}`))
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"missing code"}`))
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"missing state"}`))
		return
	}

	token, err := h.authService.CompleteGoogleLogin(r.Context(), code, state)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"cannot complete google login"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"access_token":"` + token + `"}`))
}
