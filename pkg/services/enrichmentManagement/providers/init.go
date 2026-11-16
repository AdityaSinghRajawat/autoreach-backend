package providers

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	apolloIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/apollo"
	contactoutIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/contactout"
	signalhireIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/signalhire"
	apolloProvider "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers/apollo"
	contactoutProvider "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers/contactout"
	signalhireProvider "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers/signalhire"
)

type Provider interface {
	Code() string
}

type Discoverer interface {
	Provider
	DiscoverPeople(
		ctx context.Context,
		query enrichmentTypes.DiscoveryQuery,
	) ([]enrichmentTypes.Person, error)
}

type ContactFinder interface {
	Provider
	FindContact(
		ctx context.Context,
		person enrichmentTypes.Person,
	) (*enrichmentTypes.ContactDetails, error)
}

type DomainResolver interface {
	Provider
	ResolveCompanyDomain(ctx context.Context, companyName string) (string, error)
}

type AsyncContactFinder interface {
	Provider
	RequestContact(ctx context.Context, person enrichmentTypes.Person) (string, error)
	ParseCallback(payload []byte) ([]enrichmentTypes.CallbackResult, error)
}

func NewProviderRegistry(
	policies []config.ProviderPolicy,
	signalHireCallbackURL string,
) []Provider {
	registry := make([]Provider, 0, len(policies))

	for _, policy := range policies {
		if provider := newProvider(policy, signalHireCallbackURL); provider != nil {
			registry = append(registry, provider)
		}
	}

	return registry
}

func newProvider(policy config.ProviderPolicy, signalHireCallbackURL string) Provider {
	if policy.APIKey == "" {
		return nil
	}

	switch policy.Code {
	case config.ProviderApollo:
		return apolloProvider.NewApolloProvider(apolloIntegration.NewApolloIntegration(policy))
	case config.ProviderSignalHire:
		if signalHireCallbackURL == "" {
			return nil
		}

		return signalhireProvider.NewSignalHireProvider(
			signalhireIntegration.NewSignalHireIntegration(policy),
			signalHireCallbackURL,
		)
	case config.ProviderContactOut:
		return contactoutProvider.NewContactOutProvider(
			contactoutIntegration.NewContactOutIntegration(policy),
		)
	default:
		return nil
	}
}
