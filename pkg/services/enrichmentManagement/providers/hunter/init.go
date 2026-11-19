package hunter

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	hunterIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/hunter"
)

type HunterProvider struct {
	integration *hunterIntegration.HunterIntegration
}

func NewHunterProvider(integration *hunterIntegration.HunterIntegration) *HunterProvider {
	return &HunterProvider{integration: integration}
}

func (p *HunterProvider) Code() string {
	return config.ProviderHunter
}
