package signalhire

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	signalhireIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/signalhire"
)

type SignalHireProvider struct {
	integration *signalhireIntegration.SignalHireIntegration
	callbackURL string
}

func NewSignalHireProvider(
	integration *signalhireIntegration.SignalHireIntegration,
	callbackURL string,
) *SignalHireProvider {
	return &SignalHireProvider{integration: integration, callbackURL: callbackURL}
}

func (p *SignalHireProvider) Code() string {
	return config.ProviderSignalHire
}
