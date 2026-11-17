package authManagement

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// accountToken carries userID.expiry.signature, so nothing is stored to check an emailed link.
// bound ties the signature to state that must not change, so a reset link stops working once the
// password it was issued for has changed.
func accountToken(keyLabel, userID string, expiresAt time.Time, bound string) string {
	expiry := strconv.FormatInt(expiresAt.Unix(), 10)
	signature := utils.Sign(
		utils.DeriveKey(config.GetJWTSecret(), keyLabel),
		userID,
		expiry,
		bound,
	)

	return strings.Join([]string{userID, expiry, signature}, config.GetVerificationTokenSeparator())
}

// accountTokenUser returns the user a well-formed, unexpired token names. Its signature is checked
// by verifyAccountToken once the bound state has been loaded.
func accountTokenUser(token string, now time.Time) (string, bool) {
	parts := strings.Split(token, config.GetVerificationTokenSeparator())
	if len(parts) != config.GetVerificationTokenParts() {
		return "", false
	}

	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || now.Unix() > expiry {
		return "", false
	}

	return parts[0], true
}

func verifyAccountToken(keyLabel, token, bound string) bool {
	parts := strings.Split(token, config.GetVerificationTokenSeparator())

	return len(parts) == config.GetVerificationTokenParts() && utils.VerifySignature(
		utils.DeriveKey(config.GetJWTSecret(), keyLabel),
		parts[2],
		parts[0],
		parts[1],
		bound,
	)
}

func accountLink(baseURL, token string) (string, error) {
	link, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("account link: %w", err)
	}

	query := link.Query()
	query.Set(config.GetVerificationTokenParam(), token)
	link.RawQuery = query.Encode()

	return link.String(), nil
}
