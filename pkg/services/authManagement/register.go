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

// Register returns no session when email verification is required: the account works once the
// emailed link is opened.
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

	if s.requireVerification {
		return nil, s.registerPendingVerification(ctx, user, candidate)
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

// registerPendingVerification answers alike whether or not the address is registered: a new
// address gets a verification link, an unverified one gets a fresh link, and a verified one gets a
// notice. The existing account is never changed, so signing up cannot take it over.
func (s *AuthService) registerPendingVerification(
	ctx context.Context,
	user *authModel.User,
	candidate *candidateModel.Candidate,
) *config.CustomError {
	existing, err := s.userRepo.FindByEmail(ctx, user.Email)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("register: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if existing != nil {
		s.sendSignupEmail(ctx, existing, existing.EmailVerifiedAt != nil)

		return nil
	}

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.userRepo.Upsert(ctx, user); err != nil {
			return err
		}

		_, err := s.candidateRepo.Upsert(ctx, candidate)

		return err
	})
	if errors.Is(err, config.ErrDuplicateRecord) {
		return nil
	}

	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("register: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	s.sendSignupEmail(ctx, user, false)

	return nil
}
