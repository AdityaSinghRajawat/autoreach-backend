package contactout

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	contactoutIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/contactout"
)

type ContactOutProvider struct {
	integration *contactoutIntegration.ContactOutIntegration
}

func NewContactOutProvider(
	integration *contactoutIntegration.ContactOutIntegration,
) *ContactOutProvider {
	return &ContactOutProvider{integration: integration}
}

func (p *ContactOutProvider) Code() string {
	return config.ProviderContactOut
}
