package in_ports

import (
	"context"

	"steam-observer/internal/modules/market/domain"
)

type MarketService interface {
	ListTrackedItems(ctx context.Context, userID string) ([]domain.TrackedItem, error)
}
