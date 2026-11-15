package dependencies

import (
	applicationHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/applicationManagement"
	authHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/authManagement"
	candidateHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/candidateManagement"
	enrichmentHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/enrichmentManagement"
	jobHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/jobManagement"
	matchingHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/matchingManagement"
	outreachHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/outreachManagement"
)

type Handlers struct {
	Application *applicationHandler.ApplicationHandler
	Auth        *authHandler.AuthHandler
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

	d.Handlers = Handlers{
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
