package outreachManagement

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) getChannelAccount(
	ctx context.Context,
	candidateID, accountID string,
) (*outreachModel.ChannelAccount, *config.CustomError) {
	account, err := s.channelAccountRepo.FindByCandidate(ctx, candidateID, accountID)

	return config.RequireFound(
		"get channel account",
		account,
		err,
		config.ErrChannelAccountNotFound,
	)
}

func (s *OutreachService) toEmailAccount(
	account *outreachModel.ChannelAccount,
) (sharedTypes.EmailAccount, *config.CustomError) {
	if s.credentialCipher == nil {
		return sharedTypes.EmailAccount{}, config.NewCustomError(
			fmt.Errorf("open channel account: %w", config.ErrCredentialEncryptionUnavailable),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	plaintext, err := s.credentialCipher.Decrypt(
		account.EncryptedCredentials,
		account.KeyVersion,
		[]byte(account.ID),
	)
	if err != nil {
		return sharedTypes.EmailAccount{}, config.NewCustomError(
			fmt.Errorf("open channel account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	credentials := outreachTypes.ChannelAccountCredentials{}
	if err := json.Unmarshal(plaintext, &credentials); err != nil {
		return sharedTypes.EmailAccount{}, config.NewCustomError(
			fmt.Errorf("open channel account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return sharedTypes.EmailAccount{
		Address:  account.EmailAddress,
		Username: credentials.Username,
		Password: credentials.Password,
		SMTP: sharedTypes.EmailServer{
			Host:     account.SMTPHost,
			Port:     account.SMTPPort,
			Security: string(account.SMTPSecurity),
		},
		IMAP: sharedTypes.EmailServer{
			Host:     account.IMAPHost,
			Port:     account.IMAPPort,
			Security: string(account.IMAPSecurity),
		},
	}, nil
}
