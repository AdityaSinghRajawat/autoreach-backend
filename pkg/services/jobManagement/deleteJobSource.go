package jobManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *JobService) DeleteJobSource(ctx context.Context, sourceID string) *config.CustomError {
	deleted, err := s.jobSourceRepo.Delete(ctx, sourceID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("delete job source: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !deleted {
		return config.NewCustomError(
			fmt.Errorf("delete job source: %w", config.ErrJobSourceNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}
