package dependencies

import (
	authRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/authManagement"
	candidateRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/candidateManagement"
	enrichmentRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/enrichmentManagement"
	jobRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/jobManagement"
	matchingRepo "github.com/AdityaSinghRajawat/autoreach-backend/pkg/repositories/matchingManagement"
)

type Repositories struct {
	User              *authRepo.UserRepo
	RefreshToken      *authRepo.RefreshTokenRepo
	APIClient         *authRepo.APIClientRepo
	Candidate         *candidateRepo.CandidateRepo
	Resume            *candidateRepo.ResumeRepo
	JobPosting        *jobRepo.JobPostingRepo
	JobMatch          *matchingRepo.JobMatchRepo
	CompanyContact    *enrichmentRepo.CompanyContactRepo
	EnrichmentRequest *enrichmentRepo.EnrichmentRequestRepo
}

func (d *Dependencies) initRepositories() error {
	userRepo, err := authRepo.NewUserRepo(d.Postgres)
	if err != nil {
		return err
	}

	refreshTokenRepo, err := authRepo.NewRefreshTokenRepo(d.Postgres)
	if err != nil {
		return err
	}

	apiClientRepo, err := authRepo.NewAPIClientRepo(d.Postgres)
	if err != nil {
		return err
	}

	candidateRepository, err := candidateRepo.NewCandidateRepo(d.Postgres)
	if err != nil {
		return err
	}

	resumeRepo, err := candidateRepo.NewResumeRepo(d.Postgres)
	if err != nil {
		return err
	}

	jobPostingRepo, err := jobRepo.NewJobPostingRepo(d.Postgres)
	if err != nil {
		return err
	}

	jobMatchRepo, err := matchingRepo.NewJobMatchRepo(d.Postgres)
	if err != nil {
		return err
	}

	companyContactRepo, err := enrichmentRepo.NewCompanyContactRepo(d.Postgres)
	if err != nil {
		return err
	}

	enrichmentRequestRepo, err := enrichmentRepo.NewEnrichmentRequestRepo(d.Postgres)
	if err != nil {
		return err
	}

	d.Repositories = Repositories{
		CompanyContact:    companyContactRepo,
		EnrichmentRequest: enrichmentRequestRepo,
		User:              userRepo,
		RefreshToken:      refreshTokenRepo,
		APIClient:         apiClientRepo,
		Candidate:         candidateRepository,
		Resume:            resumeRepo,
		JobPosting:        jobPostingRepo,
		JobMatch:          jobMatchRepo,
	}

	return nil
}
