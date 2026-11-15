package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *EnrichmentService) requireMatchedJob(
	ctx context.Context,
	candidateID, jobID string,
) *config.CustomError {
	if _, err := uuid.Parse(jobID); err != nil {
		return config.NewCustomError(
			fmt.Errorf("require matched job: %w", config.ErrJobNotMatched),
			config.GetErrCodeNotFound(),
		)
	}

	matched, err := s.jobMatchRepo.Exists(ctx, candidateID, jobID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("require matched job: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !matched {
		return config.NewCustomError(
			fmt.Errorf("require matched job: %w", config.ErrJobNotMatched),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}
