package utils

import (
	"context"
	"net"
	"net/mail"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

type MXResolver interface {
	LookupMX(ctx context.Context, domain string) ([]*net.MX, error)
}

type EmailVerifier struct {
	resolver MXResolver
}

func NewEmailVerifier(resolver MXResolver) *EmailVerifier {
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	return &EmailVerifier{resolver: resolver}
}

func (v *EmailVerifier) Verify(
	ctx context.Context,
	email string,
	providerVerified bool,
) modelHelpers.EmailVerificationStatus {
	localPart, domain, valid := splitEmail(email)
	if !valid || config.IsDisposableDomain(domain) {
		return modelHelpers.EmailStatusInvalid
	}

	lookupCtx, cancel := context.WithTimeout(ctx, config.GetMXLookupTimeout())
	defer cancel()

	records, err := v.resolver.LookupMX(lookupCtx, domain)
	if err != nil || len(records) == 0 {
		return modelHelpers.EmailStatusInvalid
	}

	if config.IsRoleLocalPart(localPart) || !providerVerified {
		return modelHelpers.EmailStatusRisky
	}

	return modelHelpers.EmailStatusVerified
}

func splitEmail(email string) (localPart, domain string, valid bool) {
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", "", false
	}

	localPart, domain, found := strings.Cut(email, "@")
	if !found || localPart == "" || !strings.Contains(domain, ".") {
		return "", "", false
	}

	return localPart, strings.ToLower(domain), true
}
