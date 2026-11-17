package userManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Works once and only for the signed-in user: the link is signed for that user's old address.
func (s *UserService) ConfirmEmailChange(
	ctx context.Context,
	principal sharedTypes.Principal,
	request userTypes.ConfirmEmailChangeRequest,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()
	newEmail := request.NormalizedEmail()

	userID, valid := utils.AccountTokenUserID(request.Token, now)
	if !valid || userID != principal.Subject {
		return invalidEmailChangeLink()
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("confirm email change: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if user == nil || !user.IsActive || !utils.IsAccountTokenSigned(
		config.GetEmailChangeKeyLabel(),
		request.Token,
		emailChangeBoundState(user.Email, newEmail),
	) {
		return invalidEmailChangeLink()
	}

	err = s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.userRepo.UpdateEmail(ctx, user.ID, newEmail, now); err != nil {
			return err
		}

		return s.candidateRepo.UpdateEmailForUser(ctx, user.ID, newEmail, now)
	})
	if errors.Is(err, config.ErrDuplicateRecord) {
		return config.NewCustomError(
			fmt.Errorf("confirm email change: %w", err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("confirm email change: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	notice, err := emailTemplates.GetEmailChangedEmail()
	if err != nil {
		utils.LogError(ctx, err)

		return nil
	}

	utils.SendAccountEmailInBackground(ctx, s.systemMailer, user.ID, user.Email, notice)

	return nil
}

func invalidEmailChangeLink() *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("confirm email change: %w", config.ErrEmailChangeLinkInvalid),
		config.GetErrCodeBadRequest(),
	)
}
