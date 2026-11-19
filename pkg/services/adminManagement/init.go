package adminManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/dashboardManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type AdminServiceDependencies struct {
	WorkerInspector workerConfigs.IWorkerInspector
	RedisUtil       *utils.RedisUtil
	MetricsRepo     *dashboardRepo.MetricsRepo
}

type AdminService struct {
	workerInspector workerConfigs.IWorkerInspector
	redisUtil       *utils.RedisUtil
	metricsRepo     *dashboardRepo.MetricsRepo
}

func NewAdminService(dependencies AdminServiceDependencies) (*AdminService, error) {
	if dependencies.WorkerInspector == nil ||
		dependencies.RedisUtil == nil ||
		dependencies.MetricsRepo == nil {
		return nil, fmt.Errorf("new admin service: %w", config.ErrMissingDependency)
	}

	return &AdminService{
		workerInspector: dependencies.WorkerInspector,
		redisUtil:       dependencies.RedisUtil,
		metricsRepo:     dependencies.MetricsRepo,
	}, nil
}

func requireKnownQueue(queue string) *config.CustomError {
	if workerConfigs.IsKnownQueue(queue) {
		return nil
	}

	return config.NewCustomError(
		fmt.Errorf("queue %s: %w", queue, config.ErrQueueNotFound),
		config.GetErrCodeNotFound(),
	)
}
