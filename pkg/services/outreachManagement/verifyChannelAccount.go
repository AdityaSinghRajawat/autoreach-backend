package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) VerifyChannelAccount(
	ctx context.Context,
	candidateID, accountID string,
) (*outreachModel.ChannelAccount, *config.CustomError) {
	account, customErr := s.getChannelAccount(ctx, candidateID, accountID)
	if customErr != nil {
		return nil, customErr
	}

	emailAccount, customErr := s.toEmailAccount(account)
	if customErr != nil {
		return nil, customErr
	}

	cursor, err := s.emailUtil.VerifyEmailAccount(ctx, emailAccount)
	if err != nil {
		account.MarkFailed(err.Error())
	} else {
		account.MarkVerified(cursor)
	}

	account.UpdateTimestamp()

	if _, err := s.channelAccountRepo.Upsert(ctx, account); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("verify channel account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return account, nil
}
