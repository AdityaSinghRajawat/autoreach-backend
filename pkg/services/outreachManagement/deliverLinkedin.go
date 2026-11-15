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

func (s *OutreachService) deliverLinkedin(
	ctx context.Context,
	messageID string,
	channel modelHelpers.OutreachChannel,
) *config.CustomError {
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("deliver linkedin message: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if message == nil || message.Channel != channel ||
		message.Status != modelHelpers.OutreachStatusSending || message.SentAt != nil {
		return nil
	}

	contact, closeReason := s.linkedinRecipient(ctx, message)
	if closeReason != nil {
		return s.finish(ctx, message, closeReason)
	}

	delivery := outreachTypes.LinkedinDelivery{
		ProfileURL: *contact.LinkedinURL,
		Note:       valueOrEmpty(message.LinkedinNote),
		Body:       message.BodyContent,
	}

	if channel == modelHelpers.OutreachChannelLinkedinDM {
		return s.deliverLinkedinMessage(ctx, message, delivery)
	}

	return s.deliverLinkedinInvite(ctx, message, delivery)
}

func (s *OutreachService) deliverLinkedinInvite(
	ctx context.Context,
	message *outreachModel.Message,
	delivery outreachTypes.LinkedinDelivery,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()

	var reservation outreachTypes.SendReservation

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		reservation, err = s.messageRepo.ReserveLinkedinInvite(
			ctx,
			message,
			config.GetLinkedinDailyInviteLimit(),
			now,
		)

		return err
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("deliver linkedin invite: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if reservation.CapReached {
		message.Reschedule(withJitter(now.Add(config.GetLinkedinCapRetryDelay())))

		return s.finish(ctx, message, nil)
	}

	if !reservation.Reserved {
		return nil
	}

	message.SentAt = &now
	outcome, sendErr := s.linkedinProvider.SendInvite(ctx, delivery)

	return s.recordLinkedinOutcome(ctx, message, outcome, sendErr)
}

// A message to a contact who got an invite waits until the candidate confirms it was accepted.
func (s *OutreachService) deliverLinkedinMessage(
	ctx context.Context,
	message *outreachModel.Message,
	delivery outreachTypes.LinkedinDelivery,
) *config.CustomError {
	invite, err := s.messageRepo.FindLinkedinInvite(ctx, message.CandidateID, message.ContactID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("deliver linkedin message: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	now := utils.GetCurrTimeStamp()

	if invite != nil && invite.Status != modelHelpers.OutreachStatusAccepted {
		if now.Sub(invite.CreatedAt) > config.GetLinkedinAcceptanceWait() {
			return s.finish(ctx, message, &outreachTypes.CloseOutcome{
				Status: modelHelpers.OutreachStatusCancelled,
				Reason: config.ErrConnectionNotAccepted,
			})
		}

		message.Reschedule(withJitter(now.Add(config.GetLinkedinAcceptanceRecheck())))

		return s.finish(ctx, message, nil)
	}

	message.SentAt = &now
	outcome, sendErr := s.linkedinProvider.SendMessage(ctx, delivery)

	return s.recordLinkedinOutcome(ctx, message, outcome, sendErr)
}

func (s *OutreachService) recordLinkedinOutcome(
	ctx context.Context,
	message *outreachModel.Message,
	outcome outreachTypes.LinkedinOutcome,
	sendErr error,
) *config.CustomError {
	switch {
	case sendErr != nil && message.AttemptCount < config.GetMaxSendAttempts():
		message.Reschedule(withJitter(utils.GetCurrTimeStamp().Add(config.GetSendRetryDelay())))

		return s.finish(ctx, message, nil)
	case sendErr != nil:
		message.SentAt = nil

		return s.finish(ctx, message, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusFailed,
			Reason: sendErr,
		})
	case outcome.AwaitingCandidate:
		message.AwaitCandidate()

		return s.finish(ctx, message, nil)
	default:
		message.MarkSent()

		if customErr := s.finish(ctx, message, nil); customErr != nil {
			return customErr
		}

		return s.recordEvent(ctx, message.ID, modelHelpers.OutreachEventSent)
	}
}

func (s *OutreachService) linkedinRecipient(
	ctx context.Context,
	message *outreachModel.Message,
) (*enrichmentModel.CompanyContact, *outreachTypes.CloseOutcome) {
	contact, err := s.companyContactRepo.FindByID(ctx, message.ContactID)
	if err != nil || contact == nil || contact.LinkedinURL == nil {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusCancelled,
			Reason: config.ErrContactHasNoLinkedin,
		}
	}

	if contact.WorkEmail != nil {
		suppressed, err := s.suppressionRepo.IsSuppressed(
			ctx,
			message.CandidateID,
			*contact.WorkEmail,
		)
		if err != nil || suppressed {
			return nil, &outreachTypes.CloseOutcome{
				Status: modelHelpers.OutreachStatusCancelled,
				Reason: config.ErrContactSuppressed,
			}
		}
	}

	ended, err := s.messageRepo.HasConversationEnded(ctx, message.CandidateID, message.ContactID)
	if err != nil || ended {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusPaused,
			Reason: config.ErrConversationEnded,
		}
	}

	return contact, nil
}
