package userManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
)

func (s *UserService) GetCurrentUser(
	ctx context.Context,
	principal sharedTypes.Principal,
) (*userTypes.UserResponse, *config.CustomError) {
	user, err := s.userRepo.FindByID(ctx, principal.Subject)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get current user: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if user == nil || !user.IsActive {
		return nil, config.NewCustomError(
			fmt.Errorf("get current user: %w", config.ErrUserInactive),
			config.GetErrCodeUnauthorized(),
		)
	}

	candidateID, err := s.findCandidateID(ctx, user.ID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("get current user: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	response := userTypes.NewUserResponse(user, candidateID)

	return &response, nil
}
