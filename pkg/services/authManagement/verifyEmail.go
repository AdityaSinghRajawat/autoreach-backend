package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// VerifyEmail confirms the address from an emailed link; the user then signs in as usual.
func (s *AuthService) VerifyEmail(
	ctx context.Context,
	request authTypes.VerifyEmailRequest,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()

	userID, valid := accountTokenUser(request.Token, now)
	if !valid || !verifyAccountToken(config.GetEmailVerificationKeyLabel(), request.Token, "") {
		return config.NewCustomError(
			fmt.Errorf("verify email: %w", config.ErrVerificationLinkInvalid),
			config.GetErrCodeBadRequest(),
		)
	}

	verified, err := s.userRepo.MarkEmailVerified(ctx, userID, now)

	return config.RequireAffected("verify email", verified, err, config.ErrVerificationLinkInvalid)
}
