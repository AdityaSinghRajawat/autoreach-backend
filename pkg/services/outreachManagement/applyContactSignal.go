package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

// Applying an already applied signal changes nothing.
func (s *OutreachService) ApplyContactSignal(
	ctx context.Context,
	contactID, signalID string,
) (*enrichmentModel.ContactSignal, *config.CustomError) {
	signal, err := s.contactSignalRepo.FindByContact(ctx, contactID, signalID)

	signal, customErr := config.RequireFound(
		"apply contact signal", signal, err, config.ErrContactSignalNotFound,
	)
	if customErr != nil || signal.AppliedAt != nil {
		return signal, customErr
	}

	if err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		return s.applyContactSignal(ctx, signal)
	}); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("apply contact signal: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if signal.SignalType == modelHelpers.ContactSignalBounced {
		s.publishContactRefresh(ctx, contactID)
	}

	signal, err = s.contactSignalRepo.FindByContact(ctx, contactID, signalID)

	return config.RequireFound("apply contact signal", signal, err, config.ErrContactSignalNotFound)
}
