package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// ResetPassword sets a new password from an emailed link and ends every session, so whoever knew
// the old password is signed out. The link is bound to the old password, so it works once.
func (s *AuthService) ResetPassword(
	ctx context.Context,
	request authTypes.ResetPasswordRequest,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()

	userID, valid := accountTokenUser(request.Token, now)
	if !valid {
		return invalidResetLink()
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("reset password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if user == nil || !user.IsActive ||
		!verifyAccountToken(config.GetPasswordResetKeyLabel(), request.Token, user.PasswordHash) {
		return invalidResetLink()
	}

	passwordHash, err := utils.HashPassword(request.Password)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("reset password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.userRepo.ResetPassword(ctx, user.ID, passwordHash, now); err != nil {
			return err
		}

		return s.refreshTokenRepo.MarkUserRevoked(ctx, user.ID, now)
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("reset password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}

func invalidResetLink() *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("reset password: %w", config.ErrResetLinkInvalid),
		config.GetErrCodeBadRequest(),
	)
}
