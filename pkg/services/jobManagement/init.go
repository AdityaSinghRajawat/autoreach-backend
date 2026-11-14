package jobManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type JobServiceDependencies struct {
	JobPostingRepo     *jobRepo.JobPostingRepo
	TransactionManager db.TransactionManager
	WorkerClient       workerConfigs.IWorkerClient
	Embedder           sharedTypes.Embedder
}

type JobService struct {
	jobPostingRepo     *jobRepo.JobPostingRepo
	transactionManager db.TransactionManager
	workerClient       workerConfigs.IWorkerClient
	embedder           sharedTypes.Embedder
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
