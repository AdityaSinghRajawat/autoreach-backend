package applicationManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	applicationRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/applicationManagement"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement/appliers"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type ApplicationServiceDependencies struct {
	ApplicationRepo    *applicationRepo.ApplicationRepo
	CandidateRepo      *candidateRepo.CandidateRepo
	ResumeRepo         *candidateRepo.ResumeRepo
	JobPostingRepo     *jobRepo.JobPostingRepo
	TransactionManager db.TransactionManager
	WorkerClient       workerConfigs.IWorkerClient
	FileStore          sharedTypes.FileStore
	RateLimiter        sharedTypes.Limiter
	Appliers           []appliers.Applier
}

type ApplicationService struct {
	applicationRepo    *applicationRepo.ApplicationRepo
	candidateRepo      *candidateRepo.CandidateRepo
	resumeRepo         *candidateRepo.ResumeRepo
	jobPostingRepo     *jobRepo.JobPostingRepo
	transactionManager db.TransactionManager
	workerClient       workerConfigs.IWorkerClient
	fileStore          sharedTypes.FileStore
	rateLimiter        sharedTypes.Limiter
	appliers           []appliers.Applier
}

func NewApplicationService(
	dependencies ApplicationServiceDependencies,
) (*ApplicationService, error) {
	if dependencies.ApplicationRepo == nil ||
		dependencies.CandidateRepo == nil ||
		dependencies.ResumeRepo == nil ||
		dependencies.JobPostingRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil ||
		dependencies.FileStore == nil ||
		dependencies.RateLimiter == nil {
		return nil, fmt.Errorf("new application service: %w", config.ErrMissingDependency)
	}

	return &ApplicationService{
		applicationRepo:    dependencies.ApplicationRepo,
		candidateRepo:      dependencies.CandidateRepo,
		resumeRepo:         dependencies.ResumeRepo,
		jobPostingRepo:     dependencies.JobPostingRepo,
		transactionManager: dependencies.TransactionManager,
		workerClient:       dependencies.WorkerClient,
		fileStore:          dependencies.FileStore,
		rateLimiter:        dependencies.RateLimiter,
		appliers:           dependencies.Appliers,
	}, nil
}
