package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Unknown addresses get the same answer and no email.
func (s *AuthService) ForgotPassword(
	ctx context.Context,
	request authTypes.ForgotPasswordRequest,
) *config.CustomError {
	user, err := s.userRepo.FindByEmail(ctx, request.NormalizedEmail())
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("forgot password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if user == nil || !user.IsActive {
		return nil
	}

	ttl := config.GetPasswordResetTTL()
	token := utils.NewAccountToken(
		config.GetPasswordResetKeyLabel(),
		user.ID,
		utils.GetCurrTimeStamp().Add(ttl),
		user.PasswordHash,
	)

	link, err := utils.AccountLink(config.GetPasswordResetURL(), token, nil)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("forgot password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	email, err := emailTemplates.GetPasswordResetEmail(emailTemplates.PasswordResetEmailData{
		ResetURL:        link,
		ValidForMinutes: int(ttl.Minutes()),
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("forgot password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	utils.SendAccountEmailInBackground(ctx, s.systemMailer, user.ID, user.Email, email)

	return nil
}
