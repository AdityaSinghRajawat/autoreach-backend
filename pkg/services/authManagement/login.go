package authManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) Login(
	ctx context.Context,
	request authTypes.LoginRequest,
) (*authTypes.SessionResponse, *config.CustomError) {
	user, err := s.userRepo.FindByEmail(ctx, request.NormalizedEmail())
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("login: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	passwordHash := s.decoyPasswordHash
	if user != nil {
		passwordHash = user.PasswordHash
	}

	matches, err := utils.VerifyPassword(request.Password, passwordHash)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("login: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if user == nil || !matches || !user.IsActive {
		return nil, config.NewCustomError(
			fmt.Errorf("login: %w", config.ErrInvalidCredentials),
			config.GetErrCodeUnauthorized(),
		)
	}

	if s.requireVerification && user.EmailVerifiedAt == nil {
		return nil, config.NewCustomError(
			fmt.Errorf("login: %w", config.ErrEmailNotVerified),
			config.GetErrCodeForbidden(),
		)
	}

	candidateID, err := s.findCandidateID(ctx, user.ID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("login: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	session, err := s.issueSession(ctx, user, candidateID, "")
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("login: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return session, nil
}
