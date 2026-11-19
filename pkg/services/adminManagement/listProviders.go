package adminManagement

import (
	"context"
	"fmt"
	"math"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	adminTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/adminManagement"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// ListProviders reports each enrichment provider's configuration and today's credit use, and the
// cold-email bounce rate over the deliverability window.
func (s *AdminService) ListProviders(
	ctx context.Context,
) (*adminTypes.ProvidersResponse, *config.CustomError) {
	response := &adminTypes.ProvidersResponse{}

	for _, policy := range config.GetEnrichmentProviderPolicies() {
		creditsUsed, err := s.redisUtil.GetInt(ctx, enrichmentTypes.DailyCreditsKey(policy.Code))
		if err != nil {
			utils.LogWarn(
				ctx,
				"provider credits unavailable",
				"provider",
				policy.Code,
				"error",
				err.Error(),
			)
		}

		response.Enrichment = append(response.Enrichment, adminTypes.EnrichmentProviderStatus{
			Code:              policy.Code,
			Configured:        policy.APIKey != "",
			DailyCreditLimit:  policy.DailyCredits,
			CreditsUsedToday:  creditsUsed,
			RateLimit:         policy.RateLimit,
			RateWindowSeconds: int(policy.RateWindow.Seconds()),
		})
	}

	window := config.GetDeliverabilityWindow()

	sent, bounced, err := s.metricsRepo.CountEmailDeliveries(
		ctx,
		utils.GetCurrTimeStamp().Add(-window),
	)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list providers: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	response.Email = adminTypes.EmailDeliverability{
		WindowDays: int(window.Hours() / 24),
		Sent:       sent,
		Bounced:    bounced,
	}

	if sent > 0 {
		delivered := math.Round(
			float64(sent-bounced)/float64(sent)*config.GetPercentageScale()*100,
		) / 100
		response.Email.DeliverabilityPercentage = &delivered
	}

	return response, nil
}
