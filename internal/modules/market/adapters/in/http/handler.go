package http

import (
	"encoding/json"
	"net/http"

	"steam-observer/internal/modules/market/ports/in_ports"
	mw "steam-observer/internal/shared/http/middleware"
)

type Handler struct {
	service in_ports.MarketService
}

func NewHandler(service in_ports.MarketService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListTracked(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"missing user in context"}`))
		return
	}

	items, err := h.service.ListTrackedItems(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"failed to list items"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}
