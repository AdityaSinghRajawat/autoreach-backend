package dashboardManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/dashboardManagement"
)

type DashboardHandler struct {
	dashboardService *dashboardService.DashboardService
}

func NewDashboardHandler(service *dashboardService.DashboardService) (*DashboardHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("new dashboard handler: %w", config.ErrMissingDependency)
	}

	return &DashboardHandler{dashboardService: service}, nil
}
