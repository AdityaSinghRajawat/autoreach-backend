package outreachManagement

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
)

func (s *OutreachService) messageContext(
	ctx context.Context,
	candidateID, jobID string,
	contact *enrichmentModel.CompanyContact,
) (outreachTypes.MessageContext, *config.CustomError) {
	candidate, err := s.candidateRepo.FindByID(ctx, candidateID)

	candidate, customErr := config.RequireFound(
		"load message context", candidate, err, config.ErrCandidateNotFound,
	)
	if customErr != nil {
		return outreachTypes.MessageContext{}, customErr
	}

	job, err := s.jobPostingRepo.FindByID(ctx, jobID)

	job, customErr = config.RequireFound("load message context", job, err, config.ErrJobNotFound)
	if customErr != nil {
		return outreachTypes.MessageContext{}, customErr
	}

	skills, err := s.activeResumeSkills(ctx, candidateID)
	if err != nil {
		return outreachTypes.MessageContext{}, config.NewCustomError(
			fmt.Errorf("load message context: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return outreachTypes.MessageContext{
		Candidate: candidate,
		Job:       job,
		Contact:   contact,
		Skills:    skills,
	}, nil
}

func (s *OutreachService) activeResumeSkills(
	ctx context.Context,
	candidateID string,
) ([]string, error) {
	resumes, err := s.resumeRepo.ListByCandidate(ctx, candidateID)
	if err != nil {
		return nil, err
	}

	for _, resume := range resumes {
		if !resume.IsActive {
			continue
		}

		metadata := candidateTypes.ResumeMetadata{}
		if err := json.Unmarshal(resume.ParsedMetadata, &metadata); err != nil {
			return nil, err
		}

		return metadata.Skills, nil
	}

	return nil, nil
}
