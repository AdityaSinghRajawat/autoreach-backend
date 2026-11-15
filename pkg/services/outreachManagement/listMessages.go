package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) ListMessages(
	ctx context.Context,
	candidateID string,
	params outreachTypes.ListMessagesParams,
) ([]*outreachModel.Message, *sharedTypes.PaginationMeta, *config.CustomError) {
	page := params.Normalize()

	messages, total, err := s.messageRepo.ListByCandidate(
		ctx,
		candidateID,
		params.MessageFilter,
		page,
	)
	if err != nil {
		return nil, nil, config.NewCustomError(
			fmt.Errorf("list outreach messages: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	meta := sharedTypes.NewPaginationMeta(page, total)

	return messages, &meta, nil
}
