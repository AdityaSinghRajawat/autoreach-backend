package userManagement

import (
	"context"
	"fmt"
	"net/url"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	userTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// An address already registered gets the same answer and no email.
func (s *UserService) RequestEmailChange(
	ctx context.Context,
	principal sharedTypes.Principal,
	request userTypes.ChangeEmailRequest,
) *config.CustomError {
	user, customErr := s.confirmPassword(ctx, principal.Subject, request.Password)
	if customErr != nil {
		return customErr
	}

	newEmail := request.NormalizedNewEmail()
	if newEmail == user.Email {
		return config.NewCustomError(
			fmt.Errorf("request email change: %w", config.ErrEmailUnchanged),
			config.GetErrCodeBadRequest(),
		)
	}

	existing, err := s.userRepo.FindByEmail(ctx, newEmail)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("request email change: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if existing != nil {
		return nil
	}

	ttl := config.GetEmailVerificationTTL()
	token := utils.NewAccountToken(
		config.GetEmailChangeKeyLabel(),
		user.ID,
		utils.GetCurrTimeStamp().Add(ttl),
		emailChangeBoundState(user.Email, newEmail),
	)

	link, err := utils.AccountLink(config.GetEmailChangeURL(), token,
		url.Values{config.GetEmailChangeEmailParam(): {newEmail}})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("request email change: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	email, err := emailTemplates.GetEmailChangeEmail(emailTemplates.EmailChangeEmailData{
		ConfirmURL:    link,
		ValidForHours: int(ttl.Hours()),
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("request email change: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	utils.SendAccountEmailInBackground(ctx, s.systemMailer, user.ID, newEmail, email)

	return nil
}

func emailChangeBoundState(currentEmail, newEmail string) string {
	return currentEmail + config.GetEmailChangeBoundSeparator() + newEmail
}
