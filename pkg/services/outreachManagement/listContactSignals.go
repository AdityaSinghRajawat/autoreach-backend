package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
)

func (s *OutreachService) ListContactSignals(
	ctx context.Context,
	contactID string,
) ([]*enrichmentModel.ContactSignal, *config.CustomError) {
	contact, err := s.companyContactRepo.FindByID(ctx, contactID)
	if _, customErr := config.RequireFound(
		"list contact signals", contact, err, config.ErrContactNotFound,
	); customErr != nil {
		return nil, customErr
	}

	signals, err := s.contactSignalRepo.ListByContact(ctx, contactID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list contact signals: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return signals, nil
}
