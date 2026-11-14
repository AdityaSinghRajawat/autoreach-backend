package candidateManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *CandidateService) DeleteResume(
	ctx context.Context,
	candidateID, resumeID string,
) *config.CustomError {
	resume, customErr := s.GetResume(ctx, candidateID, resumeID)
	if customErr != nil {
		return customErr
	}

	deleted, err := s.resumeRepo.Delete(ctx, candidateID, resumeID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("delete resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !deleted {
		return config.NewCustomError(
			fmt.Errorf("delete resume: %w", config.ErrResumeNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	s.deleteStoredFile(ctx, resume.StorageKey)

	return nil
}
