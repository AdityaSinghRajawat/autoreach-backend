package authManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	userModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/userManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) sendSignupEmail(ctx context.Context, user *userModel.User) {
	email, err := signupEmail(user)
	if err != nil {
		utils.LogError(ctx, err)

		return
	}

	utils.SendAccountEmailInBackground(ctx, s.systemMailer, user.ID, user.Email, email)
}

func signupEmail(user *userModel.User) (sharedTypes.AccountEmail, error) {
	if user.EmailVerifiedAt != nil {
		return emailTemplates.GetAccountExistsEmail()
	}

	ttl := config.GetEmailVerificationTTL()
	token := utils.NewAccountToken(
		config.GetEmailVerificationKeyLabel(),
		user.ID,
		utils.GetCurrTimeStamp().Add(ttl),
		"",
	)

	link, err := utils.AccountLink(config.GetEmailVerificationURL(), token, nil)
	if err != nil {
		return sharedTypes.AccountEmail{}, err
	}

	return emailTemplates.GetVerificationEmail(emailTemplates.VerificationEmailData{
		VerifyURL:     link,
		ValidForHours: int(ttl.Hours()),
	})
}
