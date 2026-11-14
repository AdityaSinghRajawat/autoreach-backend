package dependencies

import (
	authHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/authManagement"
	candidateHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/candidateManagement"
	jobHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/jobManagement"
	matchingHandler "github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers/matchingManagement"
)

type Handlers struct {
	Auth      *authHandler.AuthHandler
	Candidate *candidateHandler.CandidateHandler
	Job       *jobHandler.JobHandler
	Matching  *matchingHandler.MatchingHandler
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

	d.Handlers = Handlers{
		Auth:      auth,
		Candidate: candidate,
		Job:       job,
		Matching:  matching,
	}

	return nil
}
