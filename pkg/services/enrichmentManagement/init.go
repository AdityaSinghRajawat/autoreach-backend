package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	enrichmentRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/enrichmentManagement"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	matchingRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/matchingManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

type CreditCounter interface {
	IncrementDaily(ctx context.Context, namespace, name string, amount int64) (int64, error)
}

type EnrichmentServiceDependencies struct {
	JobPostingRepo        *jobRepo.JobPostingRepo
	JobMatchRepo          *matchingRepo.JobMatchRepo
	CompanyContactRepo    *enrichmentRepo.CompanyContactRepo
	EnrichmentRequestRepo *enrichmentRepo.EnrichmentRequestRepo
	TransactionManager    db.TransactionManager
	WorkerClient          workerConfigs.IWorkerClient
	RateLimiter           sharedTypes.Limiter
	CreditCounter         CreditCounter
	EmailVerifier         *utils.EmailVerifier
	Providers             []providers.Provider
}

type EnrichmentService struct {
	jobPostingRepo        *jobRepo.JobPostingRepo
	jobMatchRepo          *matchingRepo.JobMatchRepo
	companyContactRepo    *enrichmentRepo.CompanyContactRepo
	enrichmentRequestRepo *enrichmentRepo.EnrichmentRequestRepo
	transactionManager    db.TransactionManager
	workerClient          workerConfigs.IWorkerClient
	rateLimiter           sharedTypes.Limiter
	creditCounter         CreditCounter
	emailVerifier         *utils.EmailVerifier
	providers             []providers.Provider
	circuitBreakers       map[string]*utils.CircuitBreaker
}

func NewEnrichmentService(dependencies EnrichmentServiceDependencies) (*EnrichmentService, error) {
	if dependencies.JobPostingRepo == nil ||
		dependencies.JobMatchRepo == nil ||
		dependencies.CompanyContactRepo == nil ||
		dependencies.EnrichmentRequestRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil ||
		dependencies.RateLimiter == nil ||
		dependencies.CreditCounter == nil ||
		dependencies.EmailVerifier == nil {
		return nil, fmt.Errorf("new enrichment service: %w", config.ErrMissingDependency)
	}

	circuitBreakers := make(map[string]*utils.CircuitBreaker, len(dependencies.Providers))
	for _, provider := range dependencies.Providers {
		circuitBreakers[provider.Code()] = utils.NewCircuitBreaker()
	}

	return &EnrichmentService{
		jobPostingRepo:        dependencies.JobPostingRepo,
		jobMatchRepo:          dependencies.JobMatchRepo,
		companyContactRepo:    dependencies.CompanyContactRepo,
		enrichmentRequestRepo: dependencies.EnrichmentRequestRepo,
		transactionManager:    dependencies.TransactionManager,
		workerClient:          dependencies.WorkerClient,
		rateLimiter:           dependencies.RateLimiter,
		creditCounter:         dependencies.CreditCounter,
		emailVerifier:         dependencies.EmailVerifier,
		providers:             dependencies.Providers,
		circuitBreakers:       circuitBreakers,
	}, nil
}
