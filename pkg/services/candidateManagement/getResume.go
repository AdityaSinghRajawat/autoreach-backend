package candidateManagement

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
)

func (s *CandidateService) GetResume(
	ctx context.Context,
	candidateID, resumeID string,
) (*candidateModel.Resume, *config.CustomError) {
	if _, err := uuid.Parse(resumeID); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get resume: %w", config.ErrResumeNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	resume, err := s.resumeRepo.FindByID(ctx, resumeID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if resume == nil || resume.CandidateID != candidateID {
		return nil, config.NewCustomError(
			fmt.Errorf("get resume: %w", config.ErrResumeNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return resume, nil
}
