package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
)

func (s *EnrichmentService) ListJobContacts(
	ctx context.Context,
	candidateID, jobID string,
) ([]*enrichmentModel.CompanyContact, *config.CustomError) {
	if customErr := s.requireMatchedJob(ctx, candidateID, jobID); customErr != nil {
		return nil, customErr
	}

	contacts, err := s.companyContactRepo.ListByJob(ctx, jobID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list job contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return contacts, nil
}
