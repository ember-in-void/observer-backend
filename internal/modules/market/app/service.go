package app

import (
	"context"

	"steam-observer/internal/modules/market/domain"
	"steam-observer/internal/modules/market/ports/in_ports"
)

// Пока без out_ports — вернём заглушечные данные.
type MarketService interface {
	in_ports.MarketService
}

type marketServiceImpl struct{}

func NewMarketService() MarketService {
	return &marketServiceImpl{}
}

func (s *marketServiceImpl) ListTrackedItems(ctx context.Context, userID string) ([]domain.TrackedItem, error) {
	// Пока просто заглушка: один фейковый элемент с привязкой к userID
	return []domain.TrackedItem{
		{
			ID:     "item-1",
			Name:   "Stub AK-47 | Redline",
			UserID: userID,
		},
	}, nil
}
