package app

import (
	"context"

	"steam-observer/internal/modules/dashboard/domain"
	"steam-observer/internal/modules/dashboard/ports/in_ports"
)

type dashboardServiceImpl struct{}

func NewDashboardService() in_ports.DashboardService {
	return &dashboardServiceImpl{}
}

func (s *dashboardServiceImpl) GetDashboard(ctx context.Context, userID string) (*domain.Dashboard, error) {
	// ==========================================
	// Заглушечные данные для дашборда
	// ==========================================
	return &domain.Dashboard{
		Sections: []domain.DashboardSection{
			{
				ID:          "market-1",
				Title:       "Market Observer",
				Description: "Отслеживание цен и предметов Steam",
				Type:        domain.SectionMarket,
				Icon:        "shopping_cart",
				IsEnabled:   true,
			},
			{
				ID:          "parser-1",
				Title:       "Media Parser",
				Description: "Парсинг контента и аналитика",
				Type:        domain.SectionParser,
				Icon:        "analytics",
				IsEnabled:   false, // Пока не реализовано
			},
			{
				ID:          "routine-1",
				Title:       "Routine Tasks",
				Description: "Автоматизация рутинных действий",
				Type:        domain.SectionRoutine,
				Icon:        "assignment",
				IsEnabled:   false, // Пока не реализовано
			},
		},
	}, nil
}
