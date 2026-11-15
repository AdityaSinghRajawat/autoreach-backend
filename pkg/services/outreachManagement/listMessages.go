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
) ([]*outreachTypes.MessageDetail, *sharedTypes.PaginationMeta, *config.CustomError) {
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

	details, customErr := s.withContacts(ctx, messages)
	if customErr != nil {
		return nil, nil, customErr
	}

	meta := sharedTypes.NewPaginationMeta(page, total)

	return details, &meta, nil
}

func (s *OutreachService) withContacts(
	ctx context.Context,
	messages []*outreachModel.Message,
) ([]*outreachTypes.MessageDetail, *config.CustomError) {
	contacts, err := s.companyContactRepo.ListByIDs(ctx, outreachTypes.ContactIDs(messages))
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("load message contacts: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return outreachTypes.NewMessageDetails(messages, contacts), nil
}
