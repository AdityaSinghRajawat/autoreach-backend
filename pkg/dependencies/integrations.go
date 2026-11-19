package dependencies

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	aiIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ai"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement/appliers"
	enrichmentProviders "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/jobManagement/jobBoards"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement/linkedinProviders"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (d *Dependencies) initIntegrations() error {
	if d.Embedder == nil || d.ChatCompleter == nil {
		aiClient := aiIntegration.NewGuard(
			aiIntegration.NewAIIntegration(config.GetAIBaseURL()),
			d.RateLimiter,
			d.RedisUtil,
			config.GetRateLimitPolicy(config.RateLimitPolicyAI),
		)
		if d.Embedder == nil {
			d.Embedder = aiClient
		}

		if d.ChatCompleter == nil {
			d.ChatCompleter = aiClient
		}
	}

	if d.EnrichmentProviders == nil {
		d.EnrichmentProviders = enrichmentProviders.NewProviderRegistry(
			config.GetEnrichmentProviderPolicies(),
			config.GetSignalHireCallbackURL(),
		)
	}

	if d.JobBoards == nil {
		d.JobBoards = jobBoards.NewJobBoardRegistry(jobBoards.RegistryOptions{
			GreenhouseAPIURL:      config.GetGreenhouseBoardsAPIURL(),
			LeverAPIURL:           config.GetLeverPostingsAPIURL(),
			LeverEUAPIURL:         config.GetLeverEUPostingsAPIURL(),
			AshbyAPIURL:           config.GetAshbyJobBoardAPIURL(),
			SmartRecruitersAPIURL: config.GetSmartRecruitersAPIURL(),
			WorkdayHostURLFormat:  config.GetWorkdayHostURLFormat(),
			AdzunaAPIURL:          config.GetAdzunaAPIURL(),
			AdzunaAppID:           config.GetAdzunaAppID(),
			AdzunaAppKey:          config.GetAdzunaAppKey(),
		})
	}

	if d.Appliers == nil {
		d.Appliers = appliers.NewApplierRegistry(appliers.RegistryOptions{
			GreenhouseAPIURL: config.GetGreenhouseBoardsAPIURL(),
			LeverAPIURL:      config.GetLeverPostingsAPIURL(),
			LeverEUAPIURL:    config.GetLeverEUPostingsAPIURL(),
			GreenhouseKeys:   config.GetGreenhouseBoardKeys(),
			LeverKeys:        config.GetLeverSiteKeys(),
		})
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

	if d.SystemMailer == nil {
		systemMailer, err := utils.NewSystemMailer(d.EmailUtil)
		if err != nil {
			return err
		}

		d.SystemMailer = systemMailer
		d.RequireEmailVerification = config.IsEmailVerificationRequired()
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
