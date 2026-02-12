package in_ports

import (
	"context"

	"steam-observer/internal/modules/dashboard/domain"
)

type DashboardService interface {
	GetDashboard(ctx context.Context, userID string) (*domain.Dashboard, error)
}
