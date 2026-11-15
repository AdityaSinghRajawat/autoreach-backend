package assisted

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
)

// Assisted mode never touches LinkedIn, so the candidate's account stays clear of automation limits.
type AssistedProvider struct{}

func NewAssistedProvider() *AssistedProvider {
	return &AssistedProvider{}
}

func (p *AssistedProvider) Code() string {
	return config.GetLinkedinProviderAssisted()
}

func (p *AssistedProvider) SendInvite(
	_ context.Context,
	_ outreachTypes.LinkedinDelivery,
) (outreachTypes.LinkedinOutcome, error) {
	return outreachTypes.LinkedinOutcome{AwaitingCandidate: true}, nil
}

func (p *AssistedProvider) SendMessage(
	_ context.Context,
	_ outreachTypes.LinkedinDelivery,
) (outreachTypes.LinkedinOutcome, error) {
	return outreachTypes.LinkedinOutcome{AwaitingCandidate: true}, nil
}
