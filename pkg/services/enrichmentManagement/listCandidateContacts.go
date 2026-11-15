package enrichmentManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *EnrichmentService) ListCandidateContacts(
	ctx context.Context,
	candidateID string,
	params enrichmentTypes.ListContactsParams,
) ([]*enrichmentModel.CompanyContact, *sharedTypes.PaginationMeta, *config.CustomError) {
	page := params.Normalize()

	var (
		contacts []*enrichmentModel.CompanyContact
		total    int
	)

	err := utils.RunConcurrently(
		ctx,
		utils.ConcurrentTask{
			Name: "list candidate contacts",
			Run: func(ctx context.Context) error {
				var err error

				contacts, err = s.companyContactRepo.ListByCandidate(ctx, candidateID, page)

				return err
			},
		},
		utils.ConcurrentTask{
			Name: "count candidate contacts",
			Run: func(ctx context.Context) error {
				var err error

				total, err = s.companyContactRepo.CountByCandidate(ctx, candidateID)

				return err
			},
		},
	)
	if err != nil {
		return nil, nil, config.NewCustomError(
			fmt.Errorf("list candidate contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	meta := sharedTypes.NewPaginationMeta(page, total)

	return contacts, &meta, nil
}
