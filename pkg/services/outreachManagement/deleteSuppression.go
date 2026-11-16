package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *OutreachService) DeleteSuppression(
	ctx context.Context,
	entryID string,
) *config.CustomError {
	deleted, err := s.suppressionRepo.Delete(ctx, entryID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("delete suppression: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !deleted {
		return config.NewCustomError(
			fmt.Errorf("delete suppression: %w", config.ErrSuppressionNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}
