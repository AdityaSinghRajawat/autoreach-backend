package dependencies

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	openaiIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/openai"
	enrichmentProviders "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement/linkedinProviders"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (d *Dependencies) initIntegrations() error {
	if d.Embedder == nil || d.ChatCompleter == nil {
		openAI := openaiIntegration.NewOpenAIIntegration(config.GetOpenAIBaseURL())
		if d.Embedder == nil {
			d.Embedder = openAI
		}

		if d.ChatCompleter == nil {
			d.ChatCompleter = openAI
		}
	}

	if d.EnrichmentProviders == nil {
		d.EnrichmentProviders = enrichmentProviders.NewProviderRegistry(
			config.GetEnrichmentProviderPolicies(),
			config.GetSignalHireCallbackURL(),
		)
	}

	if d.LinkedinProvider == nil {
		provider, err := linkedinProviders.NewProvider(config.GetLinkedinProvider())
		if err != nil {
			return err
		}

		d.LinkedinProvider = provider
	}

	if d.EmailUtil == nil {
		d.EmailUtil = utils.NewEmailUtil()
	}

	if d.CredentialCipher == nil && len(config.GetCredentialKeys()) > 0 {
		credentialCipher, err := utils.NewCredentialCipher(
			config.GetCredentialKeys(),
			config.GetActiveCredentialKeyVersion(),
		)
		if err != nil {
			return err
		}

		d.CredentialCipher = credentialCipher
	}

	if d.EmailVerifier == nil {
		d.EmailVerifier = utils.NewEmailVerifier(nil)
	}

	return nil
}
