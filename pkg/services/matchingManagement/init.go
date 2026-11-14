package matchingManagement

import (
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	matchingRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/matchingManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type MatchingServiceDependencies struct {
	CandidateRepo      *candidateRepo.CandidateRepo
	ResumeRepo         *candidateRepo.ResumeRepo
	JobPostingRepo     *jobRepo.JobPostingRepo
	JobMatchRepo       *matchingRepo.JobMatchRepo
	TransactionManager db.TransactionManager
	WorkerClient       workerConfigs.IWorkerClient
}

type MatchingService struct {
	candidateRepo      *candidateRepo.CandidateRepo
	resumeRepo         *candidateRepo.ResumeRepo
	jobPostingRepo     *jobRepo.JobPostingRepo
	jobMatchRepo       *matchingRepo.JobMatchRepo
	transactionManager db.TransactionManager
	workerClient       workerConfigs.IWorkerClient
}

func NewMatchingService(dependencies MatchingServiceDependencies) (*MatchingService, error) {
	if dependencies.CandidateRepo == nil ||
		dependencies.ResumeRepo == nil ||
		dependencies.JobPostingRepo == nil ||
		dependencies.JobMatchRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil {
		return nil, fmt.Errorf("new matching service: %w", config.ErrMissingDependency)
	}

	return &MatchingService{
		candidateRepo:      dependencies.CandidateRepo,
		resumeRepo:         dependencies.ResumeRepo,
		jobPostingRepo:     dependencies.JobPostingRepo,
		jobMatchRepo:       dependencies.JobMatchRepo,
		transactionManager: dependencies.TransactionManager,
		workerClient:       dependencies.WorkerClient,
	}, nil
}
