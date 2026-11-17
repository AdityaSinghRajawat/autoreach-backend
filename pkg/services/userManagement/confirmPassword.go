package userManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	userModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Guards sensitive changes, so a stolen session alone cannot make them.
func (s *UserService) confirmPassword(
	ctx context.Context,
	userID, password string,
) (*userModel.User, *config.CustomError) {
	user, err := s.userRepo.FindByID(ctx, userID)

	user, customErr := config.RequireFound("confirm password", user, err, config.ErrUserInactive)
	if customErr != nil {
		return nil, customErr
	}

	matches, err := utils.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("confirm password: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !matches || !user.IsActive {
		return nil, config.NewCustomError(
			fmt.Errorf("confirm password: %w", config.ErrCurrentPasswordWrong),
			config.GetErrCodeForbidden(),
		)
	}

	return user, nil
}

func (s *UserService) findCandidateID(ctx context.Context, userID string) (string, error) {
	candidate, err := s.candidateRepo.FindByUserID(ctx, userID)
	if err != nil || candidate == nil {
		return "", err
	}

	return candidate.ID, nil
}
