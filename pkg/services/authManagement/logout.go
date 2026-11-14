package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) Logout(
	ctx context.Context,
	request authTypes.LogoutRequest,
) *config.CustomError {
	stored, err := s.refreshTokenRepo.FindByHash(ctx, utils.Fingerprint(request.RefreshToken))
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("logout: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if stored == nil {
		return nil
	}

	if err := s.refreshTokenRepo.MarkFamilyRevoked(
		ctx,
		stored.FamilyID,
		utils.GetCurrTimeStamp(),
	); err != nil {
		return config.NewCustomError(
			fmt.Errorf("logout: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}
