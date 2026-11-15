package outreachManagement

import (
	"context"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) RecordOpen(ctx context.Context, token string) {
	if message := s.trackedMessage(ctx, token); message != nil {
		if customErr := s.recordEvent(
			ctx,
			message.ID,
			modelHelpers.OutreachEventOpened,
		); customErr != nil {
			utils.LogError(ctx, customErr)
		}
	}
}

func (s *OutreachService) trackedMessage(ctx context.Context, token string) *outreachModel.Message {
	message, err := s.messageRepo.FindByTrackingToken(ctx, token)
	if err != nil {
		utils.LogError(ctx, err)
	}

	return message
}
