package dashboardManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/dashboardManagement"
)

type DashboardServiceDependencies struct {
	MetricsRepo       *dashboardRepo.MetricsRepo
	ActivityEventRepo *dashboardRepo.ActivityEventRepo
}

type DashboardService struct {
	metricsRepo       *dashboardRepo.MetricsRepo
	activityEventRepo *dashboardRepo.ActivityEventRepo
}

func NewDashboardService(dependencies DashboardServiceDependencies) (*DashboardService, error) {
	if dependencies.MetricsRepo == nil || dependencies.ActivityEventRepo == nil {
		return nil, fmt.Errorf("new dashboard service: %w", config.ErrMissingDependency)
	}

	return &DashboardService{
		metricsRepo:       dependencies.MetricsRepo,
		activityEventRepo: dependencies.ActivityEventRepo,
	}, nil
}
