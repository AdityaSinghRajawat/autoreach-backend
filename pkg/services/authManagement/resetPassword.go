package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Signed for the old password hash, so the link works once.
func (s *AuthService) ResetPassword(
	ctx context.Context,
	request authTypes.ResetPasswordRequest,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()

	userID, valid := utils.AccountTokenUserID(request.Token, now)
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
		!utils.IsAccountTokenSigned(
			config.GetPasswordResetKeyLabel(),
			request.Token,
			user.PasswordHash,
		) {
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
