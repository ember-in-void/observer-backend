package http

import (
	"encoding/json"
	"net/http"

	"steam-observer/internal/modules/dashboard/ports/in_ports"
	"steam-observer/internal/shared/http/middleware"
)

type DashboardHandler struct {
	service in_ports.DashboardService
}

func NewDashboardHandler(service in_ports.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: service,
	}
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Извлекаем userID из контекста (добавляется middleware)
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	dashboard, err := h.service.GetDashboard(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dashboard)
}
