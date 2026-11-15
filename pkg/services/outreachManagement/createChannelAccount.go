package outreachManagement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) CreateChannelAccount(
	ctx context.Context,
	account *outreachModel.ChannelAccount,
	credentials outreachTypes.ChannelAccountCredentials,
) (*outreachModel.ChannelAccount, *config.CustomError) {
	if account.UsesInsecureConnection() && !config.IsInsecureMailAllowed() {
		return nil, config.NewValidationError(
			fmt.Errorf("create channel account: %w", config.ErrInsecureMailNotAllowed),
			[]config.ErrorDetail{
				{Field: "smtpSecurity", Issue: config.ErrInsecureMailNotAllowed.Error()},
			},
		)
	}

	if s.credentialCipher == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create channel account: %w", config.ErrCredentialEncryptionUnavailable),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	plaintext, err := json.Marshal(credentials)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create channel account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	account.EncryptedCredentials, account.KeyVersion, err = s.credentialCipher.Encrypt(
		plaintext,
		[]byte(account.ID),
	)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create channel account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	account.UpdateTimestamp()

	if err := account.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create channel account: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	created, err := s.channelAccountRepo.Upsert(ctx, account)
	if errors.Is(err, config.ErrDuplicateRecord) {
		return nil, config.NewCustomError(
			fmt.Errorf("create channel account: %w", err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create channel account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return created, nil
}
