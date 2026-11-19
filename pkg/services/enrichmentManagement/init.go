package enrichmentManagement

import (
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

type EnrichmentServiceDependencies struct {
	JobPostingRepo        *jobRepo.JobPostingRepo
	JobMatchRepo          *matchingRepo.JobMatchRepo
	CompanyContactRepo    *enrichmentRepo.CompanyContactRepo
	EnrichmentRequestRepo *enrichmentRepo.EnrichmentRequestRepo
	CompanyDomainRepo     *enrichmentRepo.CompanyDomainRepo
	TransactionManager    db.TransactionManager
	WorkerClient          workerConfigs.IWorkerClient
	RateLimiter           sharedTypes.Limiter
	RedisUtil             *utils.RedisUtil
	EmailVerifier         *utils.EmailVerifier
	Providers             []providers.Provider
}

type EnrichmentService struct {
	jobPostingRepo        *jobRepo.JobPostingRepo
	jobMatchRepo          *matchingRepo.JobMatchRepo
	companyContactRepo    *enrichmentRepo.CompanyContactRepo
	enrichmentRequestRepo *enrichmentRepo.EnrichmentRequestRepo
	companyDomainRepo     *enrichmentRepo.CompanyDomainRepo
	transactionManager    db.TransactionManager
	workerClient          workerConfigs.IWorkerClient
	rateLimiter           sharedTypes.Limiter
	redisUtil             *utils.RedisUtil
	emailVerifier         *utils.EmailVerifier
	providers             []providers.Provider
	circuitBreakers       map[string]*utils.CircuitBreaker
}

func NewEnrichmentService(dependencies EnrichmentServiceDependencies) (*EnrichmentService, error) {
	if dependencies.JobPostingRepo == nil ||
		dependencies.JobMatchRepo == nil ||
		dependencies.CompanyContactRepo == nil ||
		dependencies.EnrichmentRequestRepo == nil ||
		dependencies.CompanyDomainRepo == nil ||
		dependencies.TransactionManager == nil ||
		dependencies.WorkerClient == nil ||
		dependencies.RateLimiter == nil ||
		dependencies.RedisUtil == nil ||
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
		companyDomainRepo:     dependencies.CompanyDomainRepo,
		transactionManager:    dependencies.TransactionManager,
		workerClient:          dependencies.WorkerClient,
		rateLimiter:           dependencies.RateLimiter,
		redisUtil:             dependencies.RedisUtil,
		emailVerifier:         dependencies.EmailVerifier,
		providers:             dependencies.Providers,
		circuitBreakers:       circuitBreakers,
	}, nil
}
