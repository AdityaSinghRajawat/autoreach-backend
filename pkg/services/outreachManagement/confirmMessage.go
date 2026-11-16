package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// REPLIED pauses the remaining steps to the contact, as an email reply does.
func (s *OutreachService) ConfirmMessage(
	ctx context.Context,
	candidateID, messageID string,
	request outreachTypes.ConfirmMessageRequest,
) (*outreachModel.Message, *config.CustomError) {
	message, customErr := s.GetMessage(ctx, candidateID, messageID)
	if customErr != nil {
		return nil, customErr
	}

	target := modelHelpers.OutreachStatus(request.Status)
	if !message.IsLinkedin() || !confirmationAllowed(message, target) {
		return nil, config.NewCustomError(
			fmt.Errorf("confirm outreach message: %w", config.ErrConfirmationNotAllowed),
			config.GetErrCodeConflict(),
		)
	}

	now := utils.GetCurrTimeStamp()
	if message.SentAt == nil {
		message.SentAt = &now
	}

	switch target {
	case modelHelpers.OutreachStatusReplied:
		message.Status = modelHelpers.OutreachStatusSent

		if customErr := s.recordReply(ctx, message, ""); customErr != nil {
			return nil, customErr
		}

		return message, nil
	case modelHelpers.OutreachStatusAccepted:
		message.Status = target
		message.UpdateTimestamp()

		return message, s.closeConversation(ctx, message, modelHelpers.OutreachEventAccepted,
			func(ctx context.Context) error {
				return s.updateContact(
					ctx,
					message.ContactID,
					func(contact *enrichmentModel.CompanyContact) {
						contact.RecordLinkedinResponse(now)
					},
				)
			})
	default:
		message.MarkSent()
		message.UpdateTimestamp()

		return message, s.closeConversation(ctx, message, modelHelpers.OutreachEventSent,
			func(context.Context) error { return nil })
	}
}

func confirmationAllowed(message *outreachModel.Message, target modelHelpers.OutreachStatus) bool {
	awaiting := message.Status == modelHelpers.OutreachStatusActionRequired
	sent := message.Status == modelHelpers.OutreachStatusSent
	accepted := message.Status == modelHelpers.OutreachStatusAccepted

	switch target {
	case modelHelpers.OutreachStatusSent:
		return awaiting
	case modelHelpers.OutreachStatusAccepted:
		return message.Channel == modelHelpers.OutreachChannelLinkedinConnect && (awaiting || sent)
	default:
		return awaiting || sent || accepted
	}
}
