package outreachManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func (s *OutreachService) DeleteTemplate(
	ctx context.Context,
	candidateID, templateID string,
) *config.CustomError {
	deleted, err := s.templateRepo.Delete(ctx, candidateID, templateID)
	if errors.Is(err, config.ErrTemplateInUse) {
		return config.NewCustomError(
			fmt.Errorf("delete template: %w", err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("delete template: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !deleted {
		return config.NewCustomError(
			fmt.Errorf("delete template: %w", config.ErrTemplateNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}
