package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// A conditional update, so it loses with 409 to a send already in progress.
func (s *OutreachService) UpdateMessage(
	ctx context.Context,
	candidateID, messageID string,
	_ outreachTypes.UpdateMessageRequest,
) (*outreachModel.Message, *config.CustomError) {
	if _, customErr := s.GetMessage(ctx, candidateID, messageID); customErr != nil {
		return nil, customErr
	}

	cancelled, err := s.messageRepo.Cancel(ctx, candidateID, messageID, utils.GetCurrTimeStamp())
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("cancel outreach message: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !cancelled {
		return nil, config.NewCustomError(
			fmt.Errorf("cancel outreach message: %w", config.ErrMessageNotCancellable),
			config.GetErrCodeConflict(),
		)
	}

	return s.GetMessage(ctx, candidateID, messageID)
}
