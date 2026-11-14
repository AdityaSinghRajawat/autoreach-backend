package dependencies

import (
	authService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/authManagement"
	candidateService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/candidateManagement"
	enrichmentService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	matchingService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/matchingManagement"
)

type Services struct {
	Auth       *authService.AuthService
	Candidate  *candidateService.CandidateService
	Job        *jobService.JobService
	Matching   *matchingService.MatchingService
	Enrichment *enrichmentService.EnrichmentService
}

func (d *Dependencies) initServices() error {
	auth, err := authService.NewAuthService(authService.AuthServiceDependencies{
		UserRepo:           d.Repositories.User,
		RefreshTokenRepo:   d.Repositories.RefreshToken,
		APIClientRepo:      d.Repositories.APIClient,
		CandidateRepo:      d.Repositories.Candidate,
		TransactionManager: d.TransactionManager,
	})
	if err != nil {
		return err
	}

	candidate, err := candidateService.NewCandidateService(
		candidateService.CandidateServiceDependencies{
			CandidateRepo:      d.Repositories.Candidate,
			ResumeRepo:         d.Repositories.Resume,
			TransactionManager: d.TransactionManager,
			WorkerClient:       d.WorkerClient,
			FileStore:          d.FileStore,
			Embedder:           d.Embedder,
		},
	)
	if err != nil {
		return err
	}

	job, err := jobService.NewJobService(jobService.JobServiceDependencies{
		JobPostingRepo:     d.Repositories.JobPosting,
		TransactionManager: d.TransactionManager,
		WorkerClient:       d.WorkerClient,
		Embedder:           d.Embedder,
	})
	if err != nil {
		return err
	}

	matching, err := matchingService.NewMatchingService(
		matchingService.MatchingServiceDependencies{
			CandidateRepo:      d.Repositories.Candidate,
			ResumeRepo:         d.Repositories.Resume,
			JobPostingRepo:     d.Repositories.JobPosting,
			JobMatchRepo:       d.Repositories.JobMatch,
			TransactionManager: d.TransactionManager,
			WorkerClient:       d.WorkerClient,
		},
	)
	if err != nil {
		return err
	}

	enrichment, err := enrichmentService.NewEnrichmentService(
		enrichmentService.EnrichmentServiceDependencies{
			JobPostingRepo:        d.Repositories.JobPosting,
			JobMatchRepo:          d.Repositories.JobMatch,
			CompanyContactRepo:    d.Repositories.CompanyContact,
			EnrichmentRequestRepo: d.Repositories.EnrichmentRequest,
			TransactionManager:    d.TransactionManager,
			WorkerClient:          d.WorkerClient,
			RateLimiter:           d.RateLimiter,
			CreditCounter:         d.RedisUtil,
			EmailVerifier:         d.EmailVerifier,
			Providers:             d.EnrichmentProviders,
		},
	)
	if err != nil {
		return err
	}

	d.Services = Services{
		Enrichment: enrichment,
		Auth:       auth,
		Candidate:  candidate,
		Job:        job,
		Matching:   matching,
	}

	return nil
}
