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
	if err != nil || candidate == nil {
		return outreachTypes.MessageContext{}, config.NewCustomError(
			fmt.Errorf(
				"load message context: candidate: %w",
				errOrNotFound(err, config.ErrCandidateNotFound),
			),
			config.GetErrCodeNotFound(),
		)
	}

	job, err := s.jobPostingRepo.FindByID(ctx, jobID)
	if err != nil || job == nil {
		return outreachTypes.MessageContext{}, config.NewCustomError(
			fmt.Errorf("load message context: job: %w", errOrNotFound(err, config.ErrJobNotFound)),
			config.GetErrCodeNotFound(),
		)
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
