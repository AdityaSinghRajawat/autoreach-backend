package candidateManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type FileStore interface {
	Save(ctx context.Context, key string, content []byte) error
	Read(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type CandidateServiceDependencies struct {
	CandidateRepo      *candidateRepo.CandidateRepo
	ResumeRepo         *candidateRepo.ResumeRepo
	TransactionManager db.TransactionManager
	WorkerClient       workerConfigs.IWorkerClient
	FileStore          FileStore
	Embedder           sharedTypes.Embedder
}

type CandidateService struct {
	candidateRepo      *candidateRepo.CandidateRepo
	resumeRepo         *candidateRepo.ResumeRepo
	transactionManager db.TransactionManager
	workerClient       workerConfigs.IWorkerClient
	fileStore          FileStore
	embedder           sharedTypes.Embedder
}

func NewCandidateService(dependencies CandidateServiceDependencies) (*CandidateService, error) {
	if dependencies.CandidateRepo == nil ||
		dependencies.ResumeRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil ||
		dependencies.FileStore == nil ||
		dependencies.Embedder == nil {
		return nil, fmt.Errorf("new candidate service: %w", config.ErrMissingDependency)
	}

	return &CandidateService{
		candidateRepo:      dependencies.CandidateRepo,
		resumeRepo:         dependencies.ResumeRepo,
		transactionManager: dependencies.TransactionManager,
		workerClient:       dependencies.WorkerClient,
		fileStore:          dependencies.FileStore,
		embedder:           dependencies.Embedder,
	}, nil
}
