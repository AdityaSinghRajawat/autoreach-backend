package authManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) RefreshSession(
	ctx context.Context,
	request authTypes.RefreshRequest,
) (*authTypes.SessionResponse, *config.CustomError) {
	now := utils.GetCurrTimeStamp()

	stored, err := s.refreshTokenRepo.FindByHash(ctx, utils.Fingerprint(request.RefreshToken))
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("refresh session: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if stored == nil {
		return nil, unauthorizedRefresh(config.ErrRefreshTokenUnknown)
	}

	if stored.IsRevoked() {
		return nil, s.revokeReusedFamily(ctx, stored.FamilyID)
	}

	if stored.IsExpired(now) {
		return nil, unauthorizedRefresh(config.ErrRefreshTokenExpired)
	}

	user, candidateID, err := s.findActiveUserWithCandidate(ctx, stored.UserID)
	if errors.Is(err, config.ErrUserInactive) {
		return nil, unauthorizedRefresh(err)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("refresh session: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	var session *authTypes.SessionResponse

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		revoked, err := s.refreshTokenRepo.MarkRevoked(ctx, stored.ID, now)
		if err != nil {
			return err
		}

		if !revoked {
			return config.ErrRefreshTokenReused
		}

		session, err = s.issueSession(ctx, user, candidateID, stored.FamilyID)

		return err
	})
	if errors.Is(err, config.ErrRefreshTokenReused) {
		return nil, s.revokeReusedFamily(ctx, stored.FamilyID)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("refresh session: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return session, nil
}

// revokeReusedFamily ends every session descended from a token presented twice, whether the reuse
// came later or raced the first refresh, since either can mean the token was stolen.
func (s *AuthService) revokeReusedFamily(ctx context.Context, familyID string) *config.CustomError {
	if err := s.refreshTokenRepo.MarkFamilyRevoked(
		ctx,
		familyID,
		utils.GetCurrTimeStamp(),
	); err != nil {
		return config.NewCustomError(
			fmt.Errorf("refresh session: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return unauthorizedRefresh(config.ErrRefreshTokenReused)
}

func unauthorizedRefresh(reason error) *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("refresh session: %w", reason),
		config.GetErrCodeUnauthorized(),
	)
}
