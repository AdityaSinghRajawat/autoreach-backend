package apollo

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	apolloIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/apollo"
)

type ApolloProvider struct {
	integration *apolloIntegration.ApolloIntegration
}

func NewApolloProvider(integration *apolloIntegration.ApolloIntegration) *ApolloProvider {
	return &ApolloProvider{integration: integration}
}

func (p *ApolloProvider) Code() string {
	return config.ProviderApollo
}
