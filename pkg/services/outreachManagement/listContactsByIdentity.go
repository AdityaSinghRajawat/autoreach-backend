package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
)

func (s *OutreachService) ListContactsByIdentity(
	ctx context.Context,
	params outreachTypes.ContactLookupParams,
) ([]*enrichmentModel.CompanyContact, *config.CustomError) {
	if params.IsEmpty() {
		return nil, config.NewCustomError(
			fmt.Errorf("list contacts: %w", config.ErrNoFieldsSupplied),
			config.GetErrCodeBadRequest(),
		)
	}

	contacts, err := s.companyContactRepo.ListByIdentity(
		ctx,
		params.NormalizedEmail(),
		params.NormalizedLinkedinURL(),
	)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return contacts, nil
}
