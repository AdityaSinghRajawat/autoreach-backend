package authManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func (s *AuthService) GetCurrentUser(
	ctx context.Context,
	principal sharedTypes.Principal,
) (*authTypes.UserResponse, *config.CustomError) {
	user, candidateID, err := s.findActiveUserWithCandidate(ctx, principal.UserID)
	if errors.Is(err, config.ErrUserInactive) {
		return nil, config.NewCustomError(
			fmt.Errorf("get current user: %w", err),
			config.GetErrCodeUnauthorized(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get current user: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	response := authTypes.NewUserResponse(user, candidateID)

	return &response, nil
}
