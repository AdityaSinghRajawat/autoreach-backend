package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) SweepMailbox(ctx context.Context, accountID string) *config.CustomError {
	account, err := s.channelAccountRepo.FindByID(ctx, accountID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("sweep mailbox: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if account == nil || account.Status != modelHelpers.ChannelAccountStatusActive ||
		account.IMAPUIDValidity == nil || account.IMAPLastUID == nil {
		return nil
	}

	emailAccount, customErr := s.toEmailAccount(account)
	if customErr != nil {
		return customErr
	}

	inbound, cursor, err := s.emailUtil.FetchInboxSince(
		ctx,
		emailAccount,
		sharedTypes.MailboxCursor{
			UIDValidity: uint32(*account.IMAPUIDValidity),
			LastUID:     uint32(*account.IMAPLastUID),
		},
	)
	if err != nil {
		utils.LogWarn(
			ctx,
			"mailbox sweep failed",
			"channel_account_id",
			accountID,
			"error",
			err.Error(),
		)

		return nil
	}

	for _, email := range inbound {
		if customErr := s.handleInbound(ctx, account, email); customErr != nil {
			return customErr
		}
	}

	if err := s.channelAccountRepo.AdvanceCursor(
		ctx,
		account.ID,
		int64(cursor.UIDValidity),
		int64(cursor.LastUID),
		utils.GetCurrTimeStamp(),
	); err != nil {
		return config.NewCustomError(
			fmt.Errorf("sweep mailbox: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}

func (s *OutreachService) handleInbound(
	ctx context.Context,
	account *outreachModel.ChannelAccount,
	email sharedTypes.InboundEmail,
) *config.CustomError {
	references, address := email.ReferencedIDs, email.FromAddress
	if email.IsBounce {
		references, address = email.BouncedMessageIDs, utils.FirstOrZero(email.BouncedAddresses)
	}

	message, err := s.matchOutreachMessage(ctx, account, references, address)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("handle inbound mail: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if message == nil {
		return nil
	}

	switch {
	case email.IsBounce:
		return s.recordBounce(ctx, message)
	case email.IsUnsubscribeRequest:
		return s.recordUnsubscribe(ctx, message)
	case email.IsAutoReply:
		return s.recordEvent(ctx, message.ID, modelHelpers.OutreachEventAutoReplied)
	default:
		return s.recordReply(ctx, message, email.FromAddress)
	}
}

func (s *OutreachService) matchOutreachMessage(
	ctx context.Context,
	account *outreachModel.ChannelAccount,
	references []string,
	address string,
) (*outreachModel.Message, error) {
	if len(references) > 0 {
		message, err := s.messageRepo.FindByProviderMessageIDs(ctx, references)
		if err != nil || (message != nil && message.CandidateID == account.CandidateID) {
			return message, err
		}
	}

	if address == "" {
		return nil, nil
	}

	return s.messageRepo.FindLatestSentToAddress(ctx, account.ID, address)
}
