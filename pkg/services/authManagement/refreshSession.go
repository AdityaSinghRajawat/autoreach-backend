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
		if err := s.refreshTokenRepo.MarkFamilyRevoked(ctx, stored.FamilyID, now); err != nil {
			return nil, config.NewCustomError(
				fmt.Errorf("refresh session: %w", err),
				config.GetErrCodeInternalServerError(),
			)
		}

		return nil, unauthorizedRefresh(config.ErrRefreshTokenReused)
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
		return nil, unauthorizedRefresh(err)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("refresh session: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return session, nil
}

func unauthorizedRefresh(reason error) *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("refresh session: %w", reason),
		config.GetErrCodeUnauthorized(),
	)
}
