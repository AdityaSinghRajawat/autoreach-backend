package dependencies

import (
	adminService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/adminManagement"
	applicationService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement"
	authService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/authManagement"
	candidateService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/candidateManagement"
	dashboardService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/dashboardManagement"
	enrichmentService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement"
	jobService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement"
	matchingService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/matchingManagement"
	outreachService "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement"
)

type Services struct {
	Admin       *adminService.AdminService
	Dashboard   *dashboardService.DashboardService
	Application *applicationService.ApplicationService
	Auth        *authService.AuthService
	Candidate   *candidateService.CandidateService
	Job         *jobService.JobService
	Matching    *matchingService.MatchingService
	Enrichment  *enrichmentService.EnrichmentService
	Outreach    *outreachService.OutreachService
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
			ChatCompleter:      d.ChatCompleter,
		},
	)
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
			CompanyDomainRepo:     d.Repositories.CompanyDomain,
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

	job, err := jobService.NewJobService(jobService.JobServiceDependencies{
		JobPostingRepo:     d.Repositories.JobPosting,
		TransactionManager: d.TransactionManager,
		WorkerClient:       d.WorkerClient,
		Embedder:           d.Embedder,
		JobSourceRepo:      d.Repositories.JobSource,
		RateLimiter:        d.RateLimiter,
		JobBoards:          d.JobBoards,
		DomainResolver:     enrichment,
	})
	if err != nil {
		return err
	}

	outreach, err := outreachService.NewOutreachService(outreachService.OutreachServiceDependencies{
		ChannelAccountRepo: d.Repositories.ChannelAccount,
		TemplateRepo:       d.Repositories.Template,
		CampaignRepo:       d.Repositories.Campaign,
		MessageRepo:        d.Repositories.OutreachMessage,
		EventRepo:          d.Repositories.OutreachEvent,
		SuppressionRepo:    d.Repositories.Suppression,
		CandidateRepo:      d.Repositories.Candidate,
		ResumeRepo:         d.Repositories.Resume,
		CompanyContactRepo: d.Repositories.CompanyContact,
		JobPostingRepo:     d.Repositories.JobPosting,
		TransactionManager: d.TransactionManager,
		WorkerClient:       d.WorkerClient,
		CredentialCipher:   d.CredentialCipher,
		EmailUtil:          d.EmailUtil,
		LinkedinProvider:   d.LinkedinProvider,
	})
	if err != nil {
		return err
	}

	application, err := applicationService.NewApplicationService(
		applicationService.ApplicationServiceDependencies{
			ApplicationRepo:    d.Repositories.Application,
			CandidateRepo:      d.Repositories.Candidate,
			ResumeRepo:         d.Repositories.Resume,
			JobPostingRepo:     d.Repositories.JobPosting,
			TransactionManager: d.TransactionManager,
			WorkerClient:       d.WorkerClient,
			FileStore:          d.FileStore,
			RateLimiter:        d.RateLimiter,
			Appliers:           d.Appliers,
		},
	)
	if err != nil {
		return err
	}

	dashboard, err := dashboardService.NewDashboardService(
		dashboardService.DashboardServiceDependencies{
			MetricsRepo:       d.Repositories.Metrics,
			ActivityEventRepo: d.Repositories.ActivityEvent,
		},
	)
	if err != nil {
		return err
	}

	admin, err := adminService.NewAdminService(adminService.AdminServiceDependencies{
		WorkerInspector: d.WorkerInspector,
		CreditReader:    d.RedisUtil,
		MetricsRepo:     d.Repositories.Metrics,
	})
	if err != nil {
		return err
	}

	d.Services = Services{
		Admin:       admin,
		Dashboard:   dashboard,
		Application: application,
		Outreach:    outreach,
		Enrichment:  enrichment,
		Auth:        auth,
		Candidate:   candidate,
		Job:         job,
		Matching:    matching,
	}

	return nil
}
