package authManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) CreateAdmin(
	ctx context.Context,
	email, password string,
) (*authModel.User, *config.CustomError) {
	if len(password) < config.GetMinPasswordLength() {
		return nil, config.NewCustomError(
			fmt.Errorf(
				"create admin: %w: minimum %d characters",
				config.ErrPasswordTooShort,
				config.GetMinPasswordLength(),
			),
			config.GetErrCodeBadRequest(),
		)
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create admin: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	user := authModel.NewUser()
	user.Email = modelHelpers.NormalizeEmail(email)
	user.PasswordHash = passwordHash
	user.Roles = []string{config.RoleAdmin}

	if err := user.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create admin: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	created, err := s.userRepo.Upsert(ctx, user)
	if errors.Is(err, config.ErrDuplicateRecord) {
		return nil, config.NewCustomError(
			fmt.Errorf("create admin: %w", err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create admin: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return created, nil
}
