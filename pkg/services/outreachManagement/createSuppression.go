package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) CreateSuppression(
	ctx context.Context,
	entry *outreachModel.SuppressionEntry,
) (*outreachModel.SuppressionEntry, *config.CustomError) {
	if err := entry.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create suppression: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	inserted, err := s.suppressionRepo.InsertIfAbsent(ctx, entry)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create suppression: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !inserted {
		return nil, config.NewCustomError(
			fmt.Errorf("create suppression: %w", config.ErrDuplicateRecord),
			config.GetErrCodeConflict(),
		)
	}

	return entry, nil
}
