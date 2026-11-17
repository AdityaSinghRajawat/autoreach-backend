package dependencies

import (
	adminHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/adminManagement"
	applicationHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/applicationManagement"
	authHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/authManagement"
	candidateHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/candidateManagement"
	dashboardHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/dashboardManagement"
	enrichmentHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/enrichmentManagement"
	jobHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/jobManagement"
	matchingHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/matchingManagement"
	outreachHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/outreachManagement"
	userHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/userManagement"
)

type Handlers struct {
	Admin       *adminHandler.AdminHandler
	Dashboard   *dashboardHandler.DashboardHandler
	Application *applicationHandler.ApplicationHandler
	Auth        *authHandler.AuthHandler
	User        *userHandler.UserHandler
	Candidate   *candidateHandler.CandidateHandler
	Job         *jobHandler.JobHandler
	Matching    *matchingHandler.MatchingHandler
	Enrichment  *enrichmentHandler.EnrichmentHandler
	Outreach    *outreachHandler.OutreachHandler
}

func (d *Dependencies) initHandlers() error {
	auth, err := authHandler.NewAuthHandler(d.Services.Auth)
	if err != nil {
		return err
	}

	user, err := userHandler.NewUserHandler(d.Services.User)
	if err != nil {
		return err
	}

	candidate, err := candidateHandler.NewCandidateHandler(d.Services.Candidate)
	if err != nil {
		return err
	}

	job, err := jobHandler.NewJobHandler(d.Services.Job)
	if err != nil {
		return err
	}

	matching, err := matchingHandler.NewMatchingHandler(d.Services.Matching)
	if err != nil {
		return err
	}

	enrichment, err := enrichmentHandler.NewEnrichmentHandler(d.Services.Enrichment)
	if err != nil {
		return err
	}

	outreach, err := outreachHandler.NewOutreachHandler(d.Services.Outreach)
	if err != nil {
		return err
	}

	application, err := applicationHandler.NewApplicationHandler(d.Services.Application)
	if err != nil {
		return err
	}

	dashboard, err := dashboardHandler.NewDashboardHandler(d.Services.Dashboard)
	if err != nil {
		return err
	}

	admin, err := adminHandler.NewAdminHandler(d.Services.Admin)
	if err != nil {
		return err
	}

	d.Handlers = Handlers{
		Admin:       admin,
		Dashboard:   dashboard,
		Application: application,
		Outreach:    outreach,
		Enrichment:  enrichment,
		Auth:        auth,
		User:        user,
		Candidate:   candidate,
		Job:         job,
		Matching:    matching,
	}

	return nil
}
