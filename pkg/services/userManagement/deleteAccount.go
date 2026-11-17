package userManagement

import (
	"context"
	"fmt"
	"slices"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Shared contacts and global suppressions stay: they protect the people contacted.
func (s *UserService) DeleteAccount(
	ctx context.Context,
	principal sharedTypes.Principal,
	request userTypes.DeleteAccountRequest,
) *config.CustomError {
	user, customErr := s.confirmPassword(ctx, principal.Subject, request.Password)
	if customErr != nil {
		return customErr
	}

	if slices.Contains(user.Roles, config.RoleAdmin) {
		return config.NewCustomError(
			fmt.Errorf("delete account: %w", config.ErrAdminSelfDeletion),
			config.GetErrCodeConflict(),
		)
	}

	candidateID, err := s.findCandidateID(ctx, user.ID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("delete account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	storageKeys, err := s.resumeRepo.ListStorageKeys(ctx, candidateID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("delete account: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	var deleted bool

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.campaignRepo.DeleteByCandidate(ctx, candidateID); err != nil {
			return err
		}

		deleted, err = s.userRepo.Delete(ctx, user.ID)

		return err
	})
	if customErr := config.RequireAffected(
		"delete account",
		deleted,
		err,
		config.ErrUserInactive,
	); customErr != nil {
		return customErr
	}

	s.deleteStoredFiles(ctx, storageKeys)
	utils.LogInfo(ctx, "account deleted on request", "user_id", user.ID)

	return nil
}

func (s *UserService) deleteStoredFiles(ctx context.Context, storageKeys []string) {
	utils.RunInBackground(ctx, "delete account files", func(ctx context.Context) {
		tasks := make([]utils.ConcurrentTask, 0, len(storageKeys))
		for _, storageKey := range storageKeys {
			tasks = append(tasks, utils.ConcurrentTask{
				Name: "delete stored file",
				Run: func(ctx context.Context) error {
					if err := s.fileStore.Delete(ctx, storageKey); err != nil {
						utils.LogWarn(ctx, "stored file delete failed", "error", err.Error())
					}

					return nil
				},
			})
		}

		_ = utils.RunConcurrently(ctx, tasks...)
	})
}
