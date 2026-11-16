package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) ListSuppressions(
	ctx context.Context,
	params outreachTypes.ListSuppressionsParams,
) ([]*outreachModel.SuppressionEntry, *sharedTypes.PaginationMeta, *config.CustomError) {
	page := params.Normalize()

	entries, total, err := s.suppressionRepo.List(
		ctx,
		params.Query,
		params.CandidateID,
		page.PerPage,
		page.Offset(),
	)
	if err != nil {
		return nil, nil, config.NewCustomError(
			fmt.Errorf("list suppressions: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	meta := sharedTypes.NewPaginationMeta(page, total)

	return entries, &meta, nil
}
