package unit

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/routes"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/mocks"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

const testClientIP = "192.0.2.1"

func rateLimitedEngine(
	t *testing.T,
	limiter sharedTypes.Limiter,
	name config.RateLimitPolicyName,
	principal *sharedTypes.Principal,
) *gin.Engine {
	t.Helper()

	engine := newEngine(t)
	engine.GET("/limited", func(c *gin.Context) {
		if principal != nil {
			utils.SetPrincipal(c, *principal)
		}

		c.Next()
	}, middlewares.RateLimit(limiter, name), func(c *gin.Context) {
		utils.BuildAndSendResponse(c, gin.H{"ok": true}, http.StatusOK)
	})

	return engine
}

func TestRateLimitRejectsWithRetryAfterAndReportsTheRemainingAllowance(t *testing.T) {
	setup.Config(t)

	blocking := &mocks.Limiter{Decision: sharedTypes.RateLimitDecision{
		NextUnitIn: 1500 * time.Millisecond,
	}}
	blocked := helpers.Do(t, rateLimitedEngine(t, blocking, config.RateLimitPolicyIP, nil),
		http.MethodGet, "/limited", nil, nil)

	require.Equal(t, http.StatusTooManyRequests, blocked.Recorder.Code)
	assert.Equal(t, "2", blocked.Recorder.Header().Get(config.GetRetryAfterHeader()))
	assert.Equal(
		t,
		string(config.GetErrCodeRateLimited()),
		helpers.DecodeEnvelope(t, blocked).Error.Code,
	)
	assert.Equal(
		t,
		fmt.Sprintf(`"ip";q=%d;w=60`, config.GetRateLimitPolicy(config.RateLimitPolicyIP).Limit),
		blocked.Recorder.Header().Get(config.GetRateLimitPolicyHeader()),
	)
	assert.Equal(t, `"ip";r=0;t=2`, blocked.Recorder.Header().Get(config.GetRateLimitHeader()))

	allowing := &mocks.Limiter{
		Decision: sharedTypes.RateLimitDecision{Allowed: true, Remaining: 41},
	}
	allowed := helpers.Do(t, rateLimitedEngine(t, allowing, config.RateLimitPolicyIP, nil),
		http.MethodGet, "/limited", nil, nil)

	require.Equal(t, http.StatusOK, allowed.Recorder.Code)
	assert.Empty(t, allowed.Recorder.Header().Get(config.GetRetryAfterHeader()))
	assert.Equal(t, `"ip";r=41`, allowed.Recorder.Header().Get(config.GetRateLimitHeader()))
}

func TestRateLimitFailsOpenOrClosedByPolicyWhenRedisIsDown(t *testing.T) {
	setup.Config(t)

	failing := &mocks.Limiter{Err: errors.New("redis unavailable")}

	open := helpers.Do(t, rateLimitedEngine(t, failing, config.RateLimitPolicyIP, nil),
		http.MethodGet, "/limited", nil, nil)
	closed := helpers.Do(t, rateLimitedEngine(t, failing, config.RateLimitPolicyAuth, nil),
		http.MethodGet, "/limited", nil, nil)

	assert.Equal(t, http.StatusOK, open.Recorder.Code, "general traffic keeps flowing")
	assert.Equal(
		t,
		http.StatusServiceUnavailable,
		closed.Recorder.Code,
		"sign-in is not left unprotected",
	)
}

func TestRateLimitKeysByPrincipalAndFallsBackToTheClientIP(t *testing.T) {
	setup.Config(t)

	principal := sharedTypes.NewUserPrincipal("user-1", "", []string{config.RoleCandidate})
	limiter := mocks.NewAllowingLimiter()

	helpers.Do(t, rateLimitedEngine(t, limiter, config.RateLimitPolicyUser, &principal),
		http.MethodGet, "/limited", nil, nil)
	helpers.Do(t, rateLimitedEngine(t, limiter, config.RateLimitPolicyUser, nil),
		http.MethodGet, "/limited", nil, nil)

	assert.Equal(t, []mocks.LimiterCall{
		{Policy: config.RateLimitPolicyUser, Subject: "user:user-1", Cost: 1},
		{Policy: config.RateLimitPolicyUser, Subject: testClientIP, Cost: 1},
	}, limiter.Calls)
}

func TestForwardedForIsIgnoredUnlessTheProxyIsTrusted(t *testing.T) {
	logger := setup.Config(t)
	limiter := mocks.NewAllowingLimiter()
	router := routes.NewRoutes(&dependencies.Dependencies{Logger: logger, RateLimiter: limiter})

	helpers.Do(t, router, http.MethodGet, config.GetAPIBasePath()+"/jobs/unknown", nil,
		map[string]string{"X-Forwarded-For": "203.0.113.7"})

	require.NotEmpty(t, limiter.Calls)
	assert.Equal(t, config.RateLimitPolicyIP, limiter.Calls[0].Policy)
	assert.Equal(
		t,
		testClientIP,
		limiter.Calls[0].Subject,
		"a spoofed header must not change the key",
	)
}

func TestOutcomeCountedPoliciesPublishOnlyRetryAfter(t *testing.T) {
	setup.Config(t)

	blocking := &mocks.Limiter{
		Decision: sharedTypes.RateLimitDecision{NextUnitIn: 90 * time.Second},
	}
	exchange := helpers.Do(
		t,
		rateLimitedEngine(t, blocking, config.RateLimitPolicyLoginAccount, nil),
		http.MethodGet,
		"/limited",
		nil,
		nil,
	)

	require.Equal(t, http.StatusTooManyRequests, exchange.Recorder.Code)
	assert.Equal(t, "90", exchange.Recorder.Header().Get(config.GetRetryAfterHeader()))
	assert.Empty(t, exchange.Recorder.Header().Values(config.GetRateLimitHeader()),
		"another account's failures must not be readable")
	assert.Empty(t, exchange.Recorder.Header().Values(config.GetRateLimitPolicyHeader()))
	assert.Equal(t, []mocks.LimiterCall{
		{Policy: config.RateLimitPolicyLoginAccount, Subject: testClientIP, Cost: 1},
	}, blocking.Calls, "the attempt is reserved before the handler runs")
	assert.Empty(t, blocking.Refunds, "a rejected attempt spent nothing, so nothing is refunded")
}
