package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) recordReply(
	ctx context.Context,
	message *outreachModel.Message,
) *config.CustomError {
	if message.Status != modelHelpers.OutreachStatusSent {
		return nil
	}

	now := utils.GetCurrTimeStamp()
	message.Status = modelHelpers.OutreachStatusReplied
	message.RepliedAt = &now
	message.UpdateTimestamp()

	return s.closeConversation(
		ctx,
		message,
		modelHelpers.OutreachEventReplied,
		func(ctx context.Context) error {
			return s.messageRepo.PauseForContact(ctx, message.CandidateID, message.ContactID, now)
		},
	)
}
