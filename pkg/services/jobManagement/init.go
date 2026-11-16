package jobManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement/jobBoards"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type JobServiceDependencies struct {
	JobPostingRepo     *jobRepo.JobPostingRepo
	TransactionManager db.TransactionManager
	WorkerClient       workerConfigs.IWorkerClient
	Embedder           sharedTypes.Embedder
	JobSourceRepo      *jobRepo.JobSourceRepo
	RateLimiter        sharedTypes.Limiter
	JobBoards          []jobBoards.JobBoard
	DomainResolver     sharedTypes.CompanyDomainResolver
}

type JobService struct {
	jobPostingRepo     *jobRepo.JobPostingRepo
	transactionManager db.TransactionManager
	workerClient       workerConfigs.IWorkerClient
	embedder           sharedTypes.Embedder
	jobSourceRepo      *jobRepo.JobSourceRepo
	rateLimiter        sharedTypes.Limiter
	jobBoards          []jobBoards.JobBoard
	domainResolver     sharedTypes.CompanyDomainResolver
}

func NewJobService(dependencies JobServiceDependencies) (*JobService, error) {
	if dependencies.JobPostingRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil ||
		dependencies.Embedder == nil ||
		dependencies.JobSourceRepo == nil ||
		dependencies.RateLimiter == nil ||
		dependencies.DomainResolver == nil {
		return nil, fmt.Errorf("new job service: %w", config.ErrMissingDependency)
	}

	return &JobService{
		jobPostingRepo:     dependencies.JobPostingRepo,
		transactionManager: dependencies.TransactionManager,
		workerClient:       dependencies.WorkerClient,
		embedder:           dependencies.Embedder,
		jobSourceRepo:      dependencies.JobSourceRepo,
		rateLimiter:        dependencies.RateLimiter,
		jobBoards:          dependencies.JobBoards,
		domainResolver:     dependencies.DomainResolver,
	}, nil
}
