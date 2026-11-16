package enrichmentManagement

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/enrichmentManagement/providers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *EnrichmentService) callProvider(
	ctx context.Context,
	provider providers.Provider,
	credits int,
	call func(ctx context.Context) error,
) error {
	code := provider.Code()
	policy, _ := config.GetEnrichmentProviderPolicy(code)
	circuitBreaker := s.circuitBreakers[code]

	if deadline, bounded := ctx.Deadline(); bounded &&
		time.Until(deadline) < policy.Timeout+config.GetEnrichmentSaveReserve() {
		return fmt.Errorf("%s: %w", code, config.ErrProviderOutOfTime)
	}

	if !circuitBreaker.Allow() {
		return fmt.Errorf("%s: %w", code, config.ErrProviderUnavailable)
	}

	decision, err := s.rateLimiter.Allow(ctx, policy.RateLimitPolicy(), code, 1)
	if err == nil && !decision.Allowed {
		return fmt.Errorf("%s: %w", code, config.ErrProviderRateLimited)
	}

	if !s.reserveCredits(ctx, code, credits, policy.DailyCredits) {
		return fmt.Errorf("%s: %w", code, config.ErrProviderBudgetExhausted)
	}

	err = call(ctx)
	circuitBreaker.Record(classifyProviderFailure(err))

	return err
}

func (s *EnrichmentService) reserveCredits(
	ctx context.Context,
	code string,
	credits, dailyCredits int,
) bool {
	if credits <= 0 || dailyCredits <= 0 {
		return true
	}

	used, err := s.creditCounter.IncrementDaily(
		ctx,
		config.GetEnrichmentCreditNamespace(),
		code,
		int64(credits),
	)
	if err != nil {
		utils.LogWarn(ctx, "credit budget unavailable, allowing call", "provider", code)

		return true
	}

	return used <= int64(dailyCredits)
}

func classifyProviderFailure(err error) (failed, tripImmediately bool) {
	if err == nil || errors.Is(err, config.ErrMissingCredential) {
		return false, false
	}

	// A rejected key or exhausted plan stops calls until the cooldown, then a probe checks again.
	var apiErr *utils.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= http.StatusInternalServerError,
			apiErr.StatusCode == http.StatusTooManyRequests ||
				apiErr.StatusCode == http.StatusUnauthorized ||
				apiErr.StatusCode == http.StatusForbidden
	}

	return true, false
}

func isProviderSkipped(err error) bool {
	return errors.Is(err, config.ErrProviderUnavailable) ||
		errors.Is(err, config.ErrProviderRateLimited) ||
		errors.Is(err, config.ErrProviderBudgetExhausted) ||
		errors.Is(err, config.ErrProviderOutOfTime)
}
