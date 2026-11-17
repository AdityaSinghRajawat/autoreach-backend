package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func RateLimit(limiter sharedTypes.Limiter, name config.RateLimitPolicyName) gin.HandlerFunc {
	policy := config.GetRateLimitPolicy(name)
	if !policy.IsEnabled() {
		return func(c *gin.Context) { c.Next() }
	}

	countsOutcome := policy.CountOnlyStatus != 0

	// Reserve first so concurrent attempts cannot all pass; refund when the outcome does not count.
	return func(c *gin.Context) {
		subject := rateLimitSubject(c, policy.Key)
		if !admit(c, limiter, policy, subject, 1) {
			return
		}

		c.Next()

		if countsOutcome && responseStatus(c) != policy.CountOnlyStatus {
			if err := limiter.Refund(c.Request.Context(), policy, subject, 1); err != nil {
				utils.LogWarn(c.Request.Context(), "rate limiter unavailable, unit not refunded",
					"policy", string(policy.Name), "error", err.Error())
			}
		}
	}
}

func admit(
	c *gin.Context,
	limiter sharedTypes.Limiter,
	policy config.RateLimitPolicy,
	subject string,
	cost int,
) bool {
	decision, err := limiter.Allow(c.Request.Context(), policy, subject, cost)
	if err != nil {
		if policy.FailClosed {
			utils.HandleCustomError(c, config.NewCustomError(
				fmt.Errorf(
					"rate limit %s: %w: %w",
					policy.Name,
					config.ErrRateLimiterUnavailable,
					err,
				),
				config.GetErrCodeServiceUnavailable(),
			))

			return false
		}

		utils.LogWarn(c.Request.Context(), "rate limiter unavailable, allowing request",
			"policy", string(policy.Name), "error", err.Error())

		return true
	}

	// Quotas spent by failures would reveal another account's failed attempts (draft §8.2).
	if policy.CountOnlyStatus == 0 {
		writeRateLimitFields(c, policy, decision)
	}

	if decision.Allowed {
		return true
	}

	c.Header(config.GetRetryAfterHeader(), strconv.FormatInt(ceilSeconds(decision.NextUnitIn), 10))
	utils.HandleCustomError(c, config.NewCustomError(
		fmt.Errorf("rate limit %s: %w", policy.Name, config.ErrRateLimitExceeded),
		config.GetErrCodeRateLimited(),
	))

	return false
}

func writeRateLimitFields(
	c *gin.Context,
	policy config.RateLimitPolicy,
	decision sharedTypes.RateLimitDecision,
) {
	limit := fmt.Sprintf(`"%s";r=%d`, policy.Name, decision.Remaining)
	if decision.NextUnitIn > 0 {
		limit += fmt.Sprintf(";t=%d", ceilSeconds(decision.NextUnitIn))
	}

	header := c.Writer.Header()
	header.Add(config.GetRateLimitPolicyHeader(),
		fmt.Sprintf(`"%s";q=%d;w=%d`, policy.Name, policy.Limit, ceilSeconds(policy.Period)))
	header.Add(config.GetRateLimitHeader(), limit)
}

func rateLimitSubject(c *gin.Context, key config.RateLimitKey) string {
	switch key {
	case config.RateLimitKeyPrincipal:
		if principal, found := utils.GetPrincipal(c); found {
			return string(principal.Kind) + config.GetKeyDelimiter() + principal.Subject
		}
	case config.RateLimitKeyAccount:
		if email := peekLoginEmail(c); email != "" {
			return utils.Fingerprint(email)
		}
	}

	return c.ClientIP()
}

func peekLoginEmail(c *gin.Context) string {
	body, err := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	request := authTypes.LoginRequest{}
	if err != nil || json.Unmarshal(body, &request) != nil {
		return ""
	}

	return request.NormalizedEmail()
}

// ErrorHandler writes handler errors after the chain returns, so read the status from them.
func responseStatus(c *gin.Context) int {
	if len(c.Errors) > 0 && !c.Writer.Written() {
		return utils.GetHTTPStatusForErrCode(toCustomError(c.Errors.Last().Err).ErrCode)
	}

	return c.Writer.Status()
}

func ceilSeconds(duration time.Duration) int64 {
	return int64((duration + time.Second - 1) / time.Second)
}
