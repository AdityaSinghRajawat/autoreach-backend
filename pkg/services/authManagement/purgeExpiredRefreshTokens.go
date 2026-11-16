package authManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) PurgeExpiredRefreshTokens(
	ctx context.Context,
	retention time.Duration,
) (int, *config.CustomError) {
	deleted, err := s.refreshTokenRepo.DeleteExpiredBefore(
		ctx,
		utils.GetCurrTimeStamp().Add(-retention),
	)
	if err != nil {
		return 0, config.NewCustomError(
			fmt.Errorf("purge expired refresh tokens: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return deleted, nil
}
