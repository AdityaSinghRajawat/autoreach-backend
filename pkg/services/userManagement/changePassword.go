package userManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Ends every session, so whoever knew the old password is signed out; the caller signs in again.
func (s *UserService) ChangePassword(
	ctx context.Context,
	principal sharedTypes.Principal,
	request userTypes.ChangePasswordRequest,
) *config.CustomError {
	user, customErr := s.confirmPassword(ctx, principal.Subject, request.CurrentPassword)
	if customErr != nil {
		return customErr
	}

	passwordHash, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("change password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	now := utils.GetCurrTimeStamp()

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.userRepo.UpdatePassword(ctx, user.ID, passwordHash, now); err != nil {
			return err
		}

		return s.refreshTokenRepo.MarkUserRevoked(ctx, user.ID, now)
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("change password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}
