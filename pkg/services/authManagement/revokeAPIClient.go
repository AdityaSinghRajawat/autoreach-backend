package authManagement

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) RevokeAPIClient(ctx context.Context, clientID string) *config.CustomError {
	if _, err := uuid.Parse(clientID); err != nil {
		return config.NewCustomError(
			fmt.Errorf("revoke api client: %w", config.ErrAPIClientNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	revoked, err := s.apiClientRepo.MarkRevoked(ctx, clientID, utils.GetCurrTimeStamp())
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("revoke api client: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !revoked {
		return config.NewCustomError(
			fmt.Errorf("revoke api client: %w", config.ErrAPIClientNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}
