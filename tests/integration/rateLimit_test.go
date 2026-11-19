//go:build integration

package integration

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

var twoPerMinute = config.RateLimitPolicy{
	Name:   "integration",
	Limit:  2,
	Period: time.Minute,
	Burst:  2,
}

func TestGCRAAllowsTheBurstThenSpacesRequestsEvenly(t *testing.T) {
	deps := setup.Dependencies(t)
	ctx := context.Background()

	first, err := deps.RateLimiter.Allow(ctx, twoPerMinute, "client", 1)
	require.NoError(t, err)
	second, err := deps.RateLimiter.Allow(ctx, twoPerMinute, "client", 1)
	require.NoError(t, err)
	third, err := deps.RateLimiter.Allow(ctx, twoPerMinute, "client", 1)
	require.NoError(t, err)
	other, err := deps.RateLimiter.Allow(ctx, twoPerMinute, "other", 1)
	require.NoError(t, err)

	assert.True(t, first.Allowed)
	assert.Equal(t, 1, first.Remaining)
	assert.True(t, second.Allowed)
	assert.Equal(t, 0, second.Remaining)
	assert.False(t, third.Allowed)
	assert.Greater(t, third.NextUnitIn, 25*time.Second, "the next slot opens one interval later")
	assert.LessOrEqual(t, third.NextUnitIn, 30*time.Second, "not at the end of a fixed window")
	assert.InDelta(t, 30*time.Second, second.NextUnitIn, float64(time.Second),
		"an allowed decision reports when one more unit returns")
	assert.True(t, other.Allowed, "limits are tracked per subject")
}

func TestGCRAPeeksWithoutSpendingAndChargesCost(t *testing.T) {
	deps := setup.Dependencies(t)
	ctx := context.Background()

	for range 3 {
		peek, err := deps.RateLimiter.Allow(ctx, twoPerMinute, "peek", 0)
		require.NoError(t, err)
		assert.True(t, peek.Allowed)
		assert.Equal(t, 2, peek.Remaining)
	}

	expensive, err := deps.RateLimiter.Allow(ctx, twoPerMinute, "peek", 2)
	require.NoError(t, err)
	assert.True(t, expensive.Allowed)
	assert.Equal(t, 0, expensive.Remaining)

	after, err := deps.RateLimiter.Allow(ctx, twoPerMinute, "peek", 0)
	require.NoError(t, err)
	assert.False(t, after.Allowed)
}

func TestFailedLoginsAreLimitedPerAccountEvenWithTheRightPassword(t *testing.T) {
	router := newRouter(t)
	register(t, router, "target@example.com")
	register(t, router, "bystander@example.com")

	burst := config.GetRateLimitPolicy(config.RateLimitPolicyLoginAccount).Burst
	for range burst {
		failed := helpers.Do(
			t,
			router,
			http.MethodPost,
			loginPath,
			authTypes.LoginRequest{
				Email:    "target@example.com",
				Password: "wrong-password-here",
			},
			nil,
		)
		require.Equal(t, http.StatusUnauthorized, failed.Recorder.Code)
	}

	locked := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "Target@Example.com", Password: testPassword}, nil)
	require.Equal(t, http.StatusTooManyRequests, locked.Recorder.Code)
	helpers.AssertMatchesSpec(t, locked)
	assert.NotEmpty(t, locked.Recorder.Header().Get(config.GetRetryAfterHeader()))

	bystander := helpers.Do(t, router, http.MethodPost, loginPath,
		authTypes.LoginRequest{Email: "bystander@example.com", Password: testPassword}, nil)
	assert.Equal(t, http.StatusOK, bystander.Recorder.Code, "other accounts are unaffected")
}

func TestTrackingAndUnsubscribeLinksAreRateLimitedPerIP(t *testing.T) {
	router := newRouter(t)

	for _, path := range []string{
		config.GetTrackingOpenPath() + "unknown-token",
		config.GetUnsubscribePath() + utils.GenerateUUID(),
	} {
		exchange := helpers.Do(t, router, http.MethodGet, path, nil, nil)
		assert.Contains(
			t,
			exchange.Recorder.Header().Get(config.GetRateLimitPolicyHeader()),
			string(config.RateLimitPolicyPublicLink),
			path,
		)
	}
}

func TestConcurrentWrongPasswordsCannotExceedTheAccountLimit(t *testing.T) {
	router := newRouter(t)
	register(t, router, "target@example.com")

	burst := config.GetRateLimitPolicy(config.RateLimitPolicyLoginAccount).Burst
	attempts := burst * 4
	rejected := atomic.Int32{}
	waiting := sync.WaitGroup{}

	for range attempts {
		waiting.Go(func() {
			attempt := helpers.Do(t, router, http.MethodPost, loginPath, authTypes.LoginRequest{
				Email:    "target@example.com",
				Password: "wrong-password-here",
			}, nil)
			if attempt.Recorder.Code == http.StatusUnauthorized {
				rejected.Add(1)
			}
		})
	}

	waiting.Wait()

	assert.LessOrEqual(t, int(rejected.Load()), burst,
		"only the burst may reach the password check, however many arrive at once")
}

func TestSuccessfulLoginsDoNotSpendTheFailedLoginLimit(t *testing.T) {
	router := newRouter(t)
	register(t, router, "target@example.com")

	burst := config.GetRateLimitPolicy(config.RateLimitPolicyLoginAccount).Burst
	for range burst + 2 {
		login := helpers.Do(t, router, http.MethodPost, loginPath,
			authTypes.LoginRequest{Email: "target@example.com", Password: testPassword}, nil)
		require.Equal(t, http.StatusOK, login.Recorder.Code, string(login.Body))
	}
}
