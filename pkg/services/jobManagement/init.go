package jobManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type Embedder interface {
	CreateEmbeddings(ctx context.Context, inputs []string) ([][]float32, error)
	Model() string
}

type JobServiceDependencies struct {
	JobPostingRepo     *jobRepo.JobPostingRepo
	TransactionManager db.TransactionManager
	WorkerClient       workerConfigs.IWorkerClient
	Embedder           Embedder
}

type JobService struct {
	jobPostingRepo     *jobRepo.JobPostingRepo
	transactionManager db.TransactionManager
	workerClient       workerConfigs.IWorkerClient
	embedder           Embedder
}

func NewJobService(dependencies JobServiceDependencies) (*JobService, error) {
	if dependencies.JobPostingRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil ||
		dependencies.Embedder == nil {
		return nil, fmt.Errorf("new job service: %w", config.ErrMissingDependency)
	}

	return &JobService{
		jobPostingRepo:     dependencies.JobPostingRepo,
		transactionManager: dependencies.TransactionManager,
		workerClient:       dependencies.WorkerClient,
		embedder:           dependencies.Embedder,
	}, nil
}
