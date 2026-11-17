package authManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// sendSignupEmail mails the verification link, or a notice when the address already has a
// verified account. It runs in the background so both cases answer the request equally fast.
func (s *AuthService) sendSignupEmail(ctx context.Context, user *authModel.User, verified bool) {
	email, err := s.signupEmail(user, verified)
	if err != nil {
		utils.LogError(ctx, err)

		return
	}

	address := user.Email

	utils.RunInBackground(ctx, "send signup email", func(ctx context.Context) {
		if err := s.systemMailer.Send(ctx, address, email); err != nil {
			utils.LogWarn(ctx, "signup email not sent", "user_id", user.ID, "error", err.Error())
		}
	})
}

func (s *AuthService) signupEmail(
	user *authModel.User,
	verified bool,
) (sharedTypes.AccountEmail, error) {
	if verified {
		return emailTemplates.GetAccountExistsEmail()
	}

	ttl := config.GetEmailVerificationTTL()

	link, err := accountLink(config.GetEmailVerificationURL(), accountToken(
		config.GetEmailVerificationKeyLabel(), user.ID, utils.GetCurrTimeStamp().Add(ttl), "",
	))
	if err != nil {
		return sharedTypes.AccountEmail{}, err
	}

	return emailTemplates.GetVerificationEmail(emailTemplates.VerificationEmailData{
		VerifyURL:     link,
		ValidForHours: int(ttl.Hours()),
	})
}
