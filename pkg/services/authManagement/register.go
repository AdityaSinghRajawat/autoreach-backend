package authManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) Register(
	ctx context.Context,
	user *authModel.User,
	candidate *candidateModel.Candidate,
	password string,
) (*authTypes.SessionResponse, *config.CustomError) {
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("register: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	user.PasswordHash = passwordHash
	user.UpdateTimestamp()
	candidate.UpdateTimestamp()

	if err := errors.Join(user.Validate(), candidate.Validate()); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("register: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	var session *authTypes.SessionResponse

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.userRepo.Upsert(ctx, user); err != nil {
			return err
		}

		if _, err := s.candidateRepo.Upsert(ctx, candidate); err != nil {
			return err
		}

		session, err = s.issueSession(ctx, user, candidate.ID, "")

		return err
	})
	if errors.Is(err, config.ErrDuplicateRecord) {
		return nil, config.NewCustomError(
			fmt.Errorf("register: %w", err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("register: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return session, nil
}
