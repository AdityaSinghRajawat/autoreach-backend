package outreachManagement

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/textproto"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// At most once: outcomes are stored on the row, so the task itself never retries a send.
func (s *OutreachService) SendColdEmail(ctx context.Context, messageID string) *config.CustomError {
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("send email: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if message == nil || message.Status != modelHelpers.OutreachStatusSending ||
		message.SentAt != nil {
		return nil
	}

	delivery, closeReason := s.prepareDelivery(ctx, message)
	if closeReason != nil {
		return s.finish(ctx, message, closeReason)
	}

	now := utils.GetCurrTimeStamp()
	if sendAt := utils.NextSendTime(now, delivery.Location, config.GetSendWindowStartHour(),
		config.GetSendWindowEndHour(), config.IsSendingOnWeekends()); sendAt.After(now) {
		message.Reschedule(withJitter(sendAt))

		return s.finish(ctx, message, nil)
	}

	htmlBody, err := s.trackedHTMLBody(message, delivery.Candidate)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("send email: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	reservation, customErr := s.reserve(ctx, message, delivery.Account, now)
	if customErr != nil || !reservation.Reserved {
		if reservation.CapReached {
			return s.deferPastCap(ctx, message, delivery.Account.ID, now)
		}

		return customErr
	}

	email := sharedTypes.OutgoingEmail{
		FromAddress: delivery.Account.EmailAddress,
		FromName:    delivery.Candidate.FullName,
		ToAddress:   *delivery.Contact.WorkEmail,
		ToName:      delivery.Contact.FullName,
		Subject:     *message.Subject,
		Body:        message.BodyContent,
		MessageID:   *message.ProviderMessageID,
		HTMLBody:    htmlBody,
		InReplyTo:   valueOrEmpty(message.InReplyTo),
		Date:        now,
	}

	email.ListUnsubscribe, email.OneClickUnsubscribe = unsubscribeHeader(
		message,
		delivery.Account.EmailAddress,
	)

	return s.recordSendOutcome(
		ctx,
		message,
		s.emailUtil.SendEmail(ctx, delivery.EmailAccount, email),
	)
}

func (s *OutreachService) prepareDelivery(
	ctx context.Context,
	message *outreachModel.Message,
) (*outreachTypes.Delivery, *outreachTypes.CloseOutcome) {
	if message.ChannelAccountID == nil {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusCancelled,
			Reason: config.ErrNoActiveMailbox,
		}
	}

	account, err := s.channelAccountRepo.FindByID(ctx, *message.ChannelAccountID)
	if err != nil || account == nil || account.Status != modelHelpers.ChannelAccountStatusActive {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusFailed,
			Reason: config.ErrNoActiveMailbox,
		}
	}

	contact, err := s.companyContactRepo.FindByID(ctx, message.ContactID)
	if err != nil || contact == nil || contact.WorkEmail == nil {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusCancelled,
			Reason: config.ErrContactNotReachable,
		}
	}

	if suppressed, err := s.suppressionRepo.IsSuppressed(
		ctx,
		message.CandidateID,
		*contact.WorkEmail,
	); err != nil ||
		suppressed {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusCancelled,
			Reason: config.ErrContactSuppressed,
		}
	}

	if ended, err := s.messageRepo.HasConversationEnded(
		ctx,
		message.CandidateID,
		message.ContactID,
	); err != nil ||
		ended {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusPaused,
			Reason: config.ErrConversationEnded,
		}
	}

	candidate, err := s.candidateRepo.FindByID(ctx, message.CandidateID)
	if err != nil || candidate == nil {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusCancelled,
			Reason: config.ErrCandidateNotFound,
		}
	}

	location, err := time.LoadLocation(candidate.Timezone)
	if err != nil {
		location = time.UTC
	}

	emailAccount, customErr := s.toEmailAccount(account)
	if customErr != nil {
		return nil, &outreachTypes.CloseOutcome{
			Status: modelHelpers.OutreachStatusFailed,
			Reason: customErr,
		}
	}

	return &outreachTypes.Delivery{
		Account:      account,
		EmailAccount: emailAccount,
		Contact:      contact,
		Candidate:    candidate,
		Location:     location,
	}, nil
}

func (s *OutreachService) reserve(
	ctx context.Context,
	message *outreachModel.Message,
	account *outreachModel.ChannelAccount,
	now time.Time,
) (outreachTypes.SendReservation, *config.CustomError) {
	if root, err := s.messageRepo.FindThreadRoot(ctx, message); err == nil && root != nil {
		message.InReplyTo = root.ProviderMessageID
	}

	providerMessageID := utils.NewEmailMessageID(account.EmailAddress)
	message.ProviderMessageID = &providerMessageID

	var reservation outreachTypes.SendReservation

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		var err error

		reservation, err = s.messageRepo.ReserveSend(ctx, message, account.DailyLimit, now)

		return err
	})
	if err != nil {
		return reservation, config.NewCustomError(
			fmt.Errorf("send email: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	message.SentAt = &now

	return reservation, nil
}

func (s *OutreachService) deferPastCap(
	ctx context.Context,
	message *outreachModel.Message,
	accountID string,
	now time.Time,
) *config.CustomError {
	oldest, err := s.messageRepo.OldestSendSince(
		ctx,
		accountID,
		now.Add(-config.GetSendAttemptWindow()),
	)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("send email: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	message.ProviderMessageID = nil
	message.InReplyTo = nil
	message.Reschedule(withJitter(oldest.Add(config.GetSendAttemptWindow())))

	return s.finish(ctx, message, nil)
}

func (s *OutreachService) recordSendOutcome(
	ctx context.Context,
	message *outreachModel.Message,
	sendErr error,
) *config.CustomError {
	if sendErr == nil {
		message.MarkSent()

		if customErr := s.finish(ctx, message, nil); customErr != nil {
			return customErr
		}

		return s.recordEvent(ctx, message.ID, modelHelpers.OutreachEventSent)
	}

	utils.LogWarn(ctx, "email send failed", "message_id", message.ID, "error", sendErr.Error())

	if isTransientSendError(sendErr) && message.AttemptCount < config.GetMaxSendAttempts() {
		message.ProviderMessageID = nil
		message.Reschedule(withJitter(utils.GetCurrTimeStamp().Add(
			time.Duration(message.AttemptCount) * config.GetSendRetryDelay(),
		)))

		return s.finish(ctx, message, nil)
	}

	message.SentAt = nil

	return s.finish(
		ctx,
		message,
		&outreachTypes.CloseOutcome{Status: modelHelpers.OutreachStatusFailed, Reason: sendErr},
	)
}

func (s *OutreachService) finish(
	ctx context.Context,
	message *outreachModel.Message,
	outcome *outreachTypes.CloseOutcome,
) *config.CustomError {
	if outcome != nil {
		message.Close(outcome.Status, outcome.Reason.Error())
	}

	_, customErr := s.saveMessage(ctx, message, "send email")

	return customErr
}

func isTransientSendError(err error) bool {
	if errors.Is(err, config.ErrSendOutcomeUnknown) {
		return false
	}

	var protocolErr *textproto.Error
	if errors.As(err, &protocolErr) {
		return protocolErr.Code >= 400 && protocolErr.Code < 500
	}

	return true
}

func withJitter(at time.Time) time.Time {
	jitter := config.GetSendJitter()
	if jitter <= 0 {
		return at
	}

	return at.Add(time.Duration(rand.Int64N(int64(jitter))))
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
