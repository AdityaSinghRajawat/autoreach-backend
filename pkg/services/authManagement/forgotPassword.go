package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// ForgotPassword emails a reset link to an active account and does nothing otherwise; the answer
// is the same either way, so it cannot be used to find out who is registered.
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

	if user == nil || !user.IsActive || s.systemMailer == nil {
		return nil
	}

	ttl := config.GetPasswordResetTTL()
	token := accountToken(config.GetPasswordResetKeyLabel(), user.ID,
		utils.GetCurrTimeStamp().Add(ttl), user.PasswordHash)

	link, err := accountLink(config.GetPasswordResetURL(), token)
	if err != nil {
		utils.LogError(ctx, err)

		return nil
	}

	email, err := emailTemplates.GetPasswordResetEmail(emailTemplates.PasswordResetEmailData{
		ResetURL:        link,
		ValidForMinutes: int(ttl.Minutes()),
	})
	if err != nil {
		utils.LogError(ctx, err)

		return nil
	}

	address, userID := user.Email, user.ID

	utils.RunInBackground(ctx, "send password reset email", func(ctx context.Context) {
		if err := s.systemMailer.Send(ctx, address, email); err != nil {
			utils.LogWarn(
				ctx,
				"password reset email not sent",
				"user_id",
				userID,
				"error",
				err.Error(),
			)
		}
	})

	return nil
}
