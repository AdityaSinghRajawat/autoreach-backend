package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) GetMessageDetail(
	ctx context.Context,
	candidateID, messageID string,
) (*outreachTypes.MessageDetail, *config.CustomError) {
	message, customErr := s.GetMessage(ctx, candidateID, messageID)
	if customErr != nil {
		return nil, customErr
	}

	details, customErr := s.withContacts(ctx, []*outreachModel.Message{message})
	if customErr != nil {
		return nil, customErr
	}

	return details[0], nil
}
