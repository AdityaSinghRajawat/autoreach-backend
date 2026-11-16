package authManagement

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func verificationKey() []byte {
	return utils.DeriveKey(config.GetJWTSecret(), config.GetEmailVerificationKeyLabel())
}

// verificationToken carries userID.expiry.signature, so nothing is stored to check a link.
func verificationToken(userID string, expiresAt time.Time) string {
	expiry := strconv.FormatInt(expiresAt.Unix(), 10)

	return strings.Join(
		[]string{userID, expiry, utils.Sign(verificationKey(), userID, expiry)},
		config.GetVerificationTokenSeparator(),
	)
}

func parseVerificationToken(token string, now time.Time) (string, bool) {
	parts := strings.Split(token, config.GetVerificationTokenSeparator())
	if len(parts) != config.GetVerificationTokenParts() ||
		!utils.VerifySignature(verificationKey(), parts[2], parts[0], parts[1]) {
		return "", false
	}

	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || now.Unix() > expiry {
		return "", false
	}

	return parts[0], true
}

func verificationLink(token string) (string, error) {
	link, err := url.Parse(config.GetEmailVerificationURL())
	if err != nil {
		return "", fmt.Errorf("verification link: %w", err)
	}

	query := link.Query()
	query.Set(config.GetVerificationTokenParam(), token)
	link.RawQuery = query.Encode()

	return link.String(), nil
}

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

	link, err := verificationLink(verificationToken(user.ID, utils.GetCurrTimeStamp().Add(ttl)))
	if err != nil {
		return sharedTypes.AccountEmail{}, err
	}

	return emailTemplates.GetVerificationEmail(emailTemplates.VerificationEmailData{
		VerifyURL:     link,
		ValidForHours: int(ttl.Hours()),
	})
}
