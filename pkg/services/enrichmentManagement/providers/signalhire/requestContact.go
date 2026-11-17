package signalhire

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
)

func (p *SignalHireProvider) RequestContact(
	ctx context.Context,
	person enrichmentTypes.Person,
) (string, error) {
	identifier := person.LookupIdentifier(config.ProviderSignalHire)
	if identifier == "" {
		return "", fmt.Errorf("signalhire request contact: %w", config.ErrMalformedProviderPayload)
	}

	return p.integration.RequestPerson(ctx, identifier, p.callbackURL)
}
