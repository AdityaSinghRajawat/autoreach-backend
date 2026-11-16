package dashboardManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type DashboardServiceDependencies struct {
	MetricsRepo       *dashboardRepo.MetricsRepo
	ActivityEventRepo *dashboardRepo.ActivityEventRepo
	RedisUtil         *utils.RedisUtil
}

type DashboardService struct {
	metricsRepo       *dashboardRepo.MetricsRepo
	activityEventRepo *dashboardRepo.ActivityEventRepo
	redisUtil         *utils.RedisUtil
}

func NewDashboardService(dependencies DashboardServiceDependencies) (*DashboardService, error) {
	if dependencies.MetricsRepo == nil || dependencies.ActivityEventRepo == nil ||
		dependencies.RedisUtil == nil {
		return nil, fmt.Errorf("new dashboard service: %w", config.ErrMissingDependency)
	}

	return &DashboardService{
		metricsRepo:       dependencies.MetricsRepo,
		activityEventRepo: dependencies.ActivityEventRepo,
		redisUtil:         dependencies.RedisUtil,
	}, nil
}
