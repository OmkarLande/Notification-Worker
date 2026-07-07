package entities

import "time"

// App represents an external application registered in the apps table.
// Each App has its own provider implementation registered in the provider factory.
type App struct {
	ID               int
	Name             string
	BaseURL          string
	ConnectionString string
	DatabaseName     string
	MaintenanceMode  bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
