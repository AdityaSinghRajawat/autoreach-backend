package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) UpdateMessage(
	ctx context.Context,
	candidateID, messageID string,
	request outreachTypes.UpdateMessageRequest,
) (*outreachModel.Message, *config.CustomError) {
	message, customErr := s.GetMessage(ctx, candidateID, messageID)
	if customErr != nil {
		return nil, customErr
	}

	if !message.IsCancellable() {
		return nil, config.NewCustomError(
			fmt.Errorf("cancel outreach message: %w", config.ErrMessageNotCancellable),
			config.GetErrCodeConflict(),
		)
	}

	message.Status = modelHelpers.OutreachStatus(request.Status)

	return s.saveMessage(ctx, message, "cancel outreach message")
}
