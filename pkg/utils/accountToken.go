package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

// NewAccountToken is userID.expiry.signature; signing boundState kills the link once that state
// changes.
func NewAccountToken(keyLabel, userID string, expiresAt time.Time, boundState string) string {
	expiry := strconv.FormatInt(expiresAt.Unix(), 10)
	signature := Sign(DeriveKey(config.GetJWTSecret(), keyLabel), userID, expiry, boundState)

	return strings.Join([]string{userID, expiry, signature}, config.GetAccountTokenSeparator())
}

func AccountTokenUserID(token string, now time.Time) (string, bool) {
	parts := strings.Split(token, config.GetAccountTokenSeparator())
	if len(parts) != config.GetAccountTokenParts() {
		return "", false
	}

	expiry, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || now.Unix() > expiry {
		return "", false
	}

	return parts[0], true
}

func IsAccountTokenSigned(keyLabel, token, boundState string) bool {
	parts := strings.Split(token, config.GetAccountTokenSeparator())

	return len(parts) == config.GetAccountTokenParts() && VerifySignature(
		DeriveKey(config.GetJWTSecret(), keyLabel),
		parts[2],
		parts[0],
		parts[1],
		boundState,
	)
}

func AccountLink(baseURL, token string, extraParams url.Values) (string, error) {
	link, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("account link: %w", err)
	}

	query := link.Query()
	query.Set(config.GetAccountTokenParam(), token)

	for name, values := range extraParams {
		query[name] = values
	}

	link.RawQuery = query.Encode()

	return link.String(), nil
}
