package adminManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	dashboardRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/dashboardManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type CreditReader interface {
	GetDaily(ctx context.Context, namespace, name string) (int64, error)
}

type AdminServiceDependencies struct {
	WorkerInspector workerConfigs.IWorkerInspector
	CreditReader    CreditReader
	MetricsRepo     *dashboardRepo.MetricsRepo
}

type AdminService struct {
	workerInspector workerConfigs.IWorkerInspector
	creditReader    CreditReader
	metricsRepo     *dashboardRepo.MetricsRepo
}

func NewAdminService(dependencies AdminServiceDependencies) (*AdminService, error) {
	if dependencies.WorkerInspector == nil ||
		dependencies.CreditReader == nil ||
		dependencies.MetricsRepo == nil {
		return nil, fmt.Errorf("new admin service: %w", config.ErrMissingDependency)
	}

	return &AdminService{
		workerInspector: dependencies.WorkerInspector,
		creditReader:    dependencies.CreditReader,
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
