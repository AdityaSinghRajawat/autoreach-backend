package authManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) RevokeAPIClient(ctx context.Context, clientID string) *config.CustomError {
	revoked, err := s.apiClientRepo.MarkRevoked(ctx, clientID, utils.GetCurrTimeStamp())

	return config.RequireAffected("revoke api client", revoked, err, config.ErrAPIClientNotFound)
}
