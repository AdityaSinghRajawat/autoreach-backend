package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) Unsubscribe(
	ctx context.Context,
	messageID, signature string,
) *config.CustomError {
	message, customErr := s.unsubscribeTarget(ctx, messageID, signature)
	if customErr != nil {
		return customErr
	}

	return s.recordUnsubscribe(ctx, message)
}

func (s *OutreachService) VerifyUnsubscribeLink(
	ctx context.Context,
	messageID, signature string,
) *config.CustomError {
	_, customErr := s.unsubscribeTarget(ctx, messageID, signature)

	return customErr
}

// The opt-out is per candidate: the contact asked one sender to stop, not every AutoReach user.
func (s *OutreachService) recordUnsubscribe(
	ctx context.Context,
	message *outreachModel.Message,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()

	return s.closeConversation(ctx, message, modelHelpers.OutreachEventUnsubscribed,
		func(ctx context.Context) error {
			if err := s.messageRepo.CancelForContact(
				ctx,
				message.CandidateID,
				message.ContactID,
				now,
			); err != nil {
				return err
			}

			return s.suppressContact(ctx, &message.CandidateID, message.ContactID,
				modelHelpers.SuppressionReasonUnsubscribed)
		})
}

func (s *OutreachService) unsubscribeTarget(
	ctx context.Context,
	messageID, signature string,
) (*outreachModel.Message, *config.CustomError) {
	if !utils.VerifySignature(unsubscribeKey(), signature, messageID) {
		return nil, config.NewCustomError(
			fmt.Errorf("unsubscribe: %w", config.ErrInvalidUnsubscribeLink),
			config.GetErrCodeNotFound(),
		)
	}

	message, err := s.messageRepo.FindByID(ctx, messageID)

	return config.RequireFound("unsubscribe", message, err, config.ErrMessageNotFound)
}

func unsubscribeHeader(message *outreachModel.Message, senderAddress string) (string, bool) {
	mailto := "<mailto:" + senderAddress + "?subject=" + config.GetUnsubscribeKeyword() + ">"

	baseURL := config.GetPublicBaseURL()
	if baseURL == "" {
		return mailto, false
	}

	return "<" + unsubscribeURL(message.ID) + ">, " + mailto, true
}

func unsubscribeURL(messageID string) string {
	return config.GetPublicBaseURL() + config.GetUnsubscribePath() + messageID + "?" +
		config.GetTrackingSignatureParam() + "=" + utils.Sign(unsubscribeKey(), messageID)
}

func unsubscribeKey() []byte {
	return utils.DeriveKey(config.GetJWTSecret(), config.GetUnsubscribeKeyLabel())
}
