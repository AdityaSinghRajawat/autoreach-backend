package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) VerifyEmail(
	ctx context.Context,
	request authTypes.VerifyEmailRequest,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()

	userID, valid := utils.AccountTokenUserID(request.Token, now)
	if !valid ||
		!utils.IsAccountTokenSigned(config.GetEmailVerificationKeyLabel(), request.Token, "") {
		return config.NewCustomError(
			fmt.Errorf("verify email: %w", config.ErrVerificationLinkInvalid),
			config.GetErrCodeBadRequest(),
		)
	}

	verified, err := s.userRepo.MarkEmailVerified(ctx, userID, now)

	return config.RequireAffected("verify email", verified, err, config.ErrVerificationLinkInvalid)
}
