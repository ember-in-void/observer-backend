package domain

type SectionType string

const (
	SectionMarket  SectionType = "market"
	SectionParser  SectionType = "parser"
	SectionRoutine SectionType = "routine"
)

type DashboardSection struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Type        SectionType `json:"type"`
	Icon        string      `json:"icon"`
	IsEnabled   bool        `json:"is_enabled"`
}

type Dashboard struct {
	Sections []DashboardSection `json:"sections"`
}
