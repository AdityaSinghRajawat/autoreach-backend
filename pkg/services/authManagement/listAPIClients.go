package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
)

func (s *AuthService) ListAPIClients(
	ctx context.Context,
) ([]*authModel.APIClient, *config.CustomError) {
	clients, err := s.apiClientRepo.List(ctx)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list api clients: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return clients, nil
}
