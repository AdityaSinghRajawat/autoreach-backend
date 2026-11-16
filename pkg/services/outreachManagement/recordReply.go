package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// recordReply closes the conversation and updates the contact; replyAddress is the sender of an
// email reply and empty for a reply confirmed on LinkedIn.
func (s *OutreachService) recordReply(
	ctx context.Context,
	message *outreachModel.Message,
	replyAddress string,
) *config.CustomError {
	if message.Status != modelHelpers.OutreachStatusSent &&
		message.Status != modelHelpers.OutreachStatusAccepted {
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
			if err := s.messageRepo.PauseForContact(
				ctx,
				message.CandidateID,
				message.ContactID,
				now,
			); err != nil {
				return err
			}

			return s.updateContact(
				ctx,
				message.ContactID,
				func(contact *enrichmentModel.CompanyContact) {
					if message.IsLinkedin() {
						contact.RecordLinkedinResponse(now)

						return
					}

					contact.RecordReply(replyAddress, now)
				},
			)
		},
	)
}
