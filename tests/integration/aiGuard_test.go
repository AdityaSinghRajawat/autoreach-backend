//go:build integration

package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/ai"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/tests/setup"
)

type quotaAI struct {
	err   error
	calls int
}

func (a *quotaAI) CreateEmbeddings(_ context.Context, inputs []string) ([][]float32, error) {
	a.calls++
	if a.err != nil {
		return nil, a.err
	}

	return make([][]float32, len(inputs)), nil
}

func (a *quotaAI) CreateChatCompletion(
	context.Context,
	sharedTypes.ChatCompletionRequest,
) (string, error) {
	return "", nil
}

func (a *quotaAI) Model() string     { return "embedding-model" }
func (a *quotaAI) ChatModel() string { return "chat-model" }

func TestAIGuardPacesTokensThroughRedisAndTellsTheCallerHowLongToWait(t *testing.T) {
	deps := setup.Dependencies(t)
	inner := &quotaAI{}
	tenTokensPerSecond := config.RateLimitPolicy{
		Name: config.RateLimitPolicyAI, Limit: 600, Period: time.Minute, Burst: 100,
	}
	guard := ai.NewGuard(inner, deps.RateLimiter, deps.RedisUtil, tenTokensPerSecond)
	hundredTokens := []string{strings.Repeat("x", 400)}

	_, err := guard.CreateEmbeddings(context.Background(), hundredTokens)
	require.NoError(t, err, "the first request fits the burst")

	_, err = guard.CreateEmbeddings(context.Background(), hundredTokens)

	wait, deferred := utils.RetryAfterOf(err)
	require.True(t, deferred, "the next request has to wait for the budget to refill")
	assert.ErrorIs(t, err, config.ErrAIRateLimited)
	assert.InDelta(t, float64(10*time.Second), float64(wait), float64(2*time.Second))
	assert.Equal(t, 1, inner.calls, "the held-back request never reached the provider")
}

func TestAIGuardPauseSurvivesAcrossCallersAndEndsWhenTheProviderIsReady(t *testing.T) {
	deps := setup.Dependencies(t)
	quota := &utils.APIError{
		StatusCode: http.StatusTooManyRequests,
		RetryAfter: time.Second,
	}
	inner := &quotaAI{err: quota}
	disabledPacing := config.RateLimitPolicy{Name: config.RateLimitPolicyAI}
	guard := ai.NewGuard(inner, deps.RateLimiter, deps.RedisUtil, disabledPacing)

	_, err := guard.CreateEmbeddings(context.Background(), []string{"a job"})
	require.Error(t, err)

	left, err := deps.RedisUtil.GetTTL(
		context.Background(),
		utils.BuildKey(
			config.GetKeyDelimiter(),
			config.GetKeyPrefix(),
			config.GetCooldownNamespace(),
			config.GetAIEmbeddingsOperation(),
		),
	)
	require.NoError(t, err)
	assert.Positive(t, left)
	assert.LessOrEqual(t, left, time.Second)

	otherProcess := ai.NewGuard(&quotaAI{}, deps.RateLimiter, deps.RedisUtil, disabledPacing)

	_, err = otherProcess.CreateEmbeddings(context.Background(), []string{"another job"})
	assert.ErrorIs(t, err, config.ErrAIPaused, "the pause is shared, not per process")

	time.Sleep(1200 * time.Millisecond)

	_, err = otherProcess.CreateEmbeddings(context.Background(), []string{"another job"})
	assert.NoError(t, err, "once the provider's wait is over, calls go through again")
}
