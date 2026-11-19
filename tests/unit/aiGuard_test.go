package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
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

type scriptedAI struct {
	embedErr   error
	embedCalls int
	chatCalls  int
}

func (a *scriptedAI) CreateEmbeddings(_ context.Context, inputs []string) ([][]float32, error) {
	a.embedCalls++
	if a.embedErr != nil {
		return nil, a.embedErr
	}

	return make([][]float32, len(inputs)), nil
}

func (a *scriptedAI) CreateChatCompletion(
	_ context.Context,
	_ sharedTypes.ChatCompletionRequest,
) (string, error) {
	a.chatCalls++

	return "reply", nil
}

func (a *scriptedAI) Model() string     { return "embedding-model" }
func (a *scriptedAI) ChatModel() string { return "chat-model" }

type scriptedLimiter struct {
	decision sharedTypes.RateLimitDecision
	costs    []int
}

func (l *scriptedLimiter) Allow(
	_ context.Context,
	_ config.RateLimitPolicy,
	_ string,
	cost int,
) (sharedTypes.RateLimitDecision, error) {
	l.costs = append(l.costs, cost)

	return l.decision, nil
}

func (l *scriptedLimiter) Refund(context.Context, config.RateLimitPolicy, string, int) error {
	return nil
}

type memoryCooldowns struct {
	left map[string]time.Duration
}

func (m *memoryCooldowns) SetKey(
	_ context.Context,
	key string,
	_ any,
	expiration time.Duration,
) error {
	m.left[key] = expiration

	return nil
}

func (m *memoryCooldowns) GetTTL(_ context.Context, key string) (time.Duration, error) {
	return m.left[key], nil
}

func embeddingsCooldownKey() string {
	return utils.BuildKey(
		config.GetKeyDelimiter(),
		config.GetKeyPrefix(),
		config.GetCooldownNamespace(),
		config.GetAIEmbeddingsOperation(),
	)
}

func newGuard(
	inner *scriptedAI,
	limiter *scriptedLimiter,
	cooldowns *memoryCooldowns,
) *ai.Guard {
	policy := config.RateLimitPolicy{
		Name:   config.RateLimitPolicyAI,
		Limit:  1000,
		Period: time.Minute,
		Burst:  100,
	}

	return ai.NewGuard(inner, limiter, cooldowns, policy)
}

func TestRetryAfterIsReadFromTheHeaderOrTheProviderMessage(t *testing.T) {
	setup.Config(t)

	assert.Equal(t, 7*time.Second, utils.ParseRetryAfter("7", ""))
	assert.InDelta(
		t,
		90*time.Second,
		utils.ParseRetryAfter(time.Now().Add(90*time.Second).UTC().Format(http.TimeFormat), ""),
		float64(2*time.Second),
	)
	assert.Equal(
		t,
		2*time.Hour+6*time.Minute+56*time.Second+510*time.Millisecond,
		utils.ParseRetryAfter(
			"",
			"Embed content requests per day limit exceeded. Please retry in 2h6m56.51s.",
		),
		"the sentence Gemini puts in its quota errors",
	)
	assert.Equal(
		t,
		7580*time.Second,
		utils.ParseRetryAfter("", `{"details":[{"retryDelay": "7580s"}]}`),
	)
	assert.Zero(t, utils.ParseRetryAfter("soon", "no hint here"))
}

func TestAPIErrorCarriesTheProvidersRetryAfter(t *testing.T) {
	setup.Config(t)

	server := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Retry-After", "12")
			writer.WriteHeader(http.StatusTooManyRequests)
		}),
	)
	t.Cleanup(server.Close)

	err := utils.NewHTTPClient(server.URL, time.Second).Get(context.Background(), "/", nil, nil)

	var apiErr *utils.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 12*time.Second, apiErr.RetryAfter)
}

func TestAIGuardPausesEveryoneAfterAProviderQuotaErrorWithoutCallingItAgain(t *testing.T) {
	setup.Config(t)

	quota := &utils.APIError{
		StatusCode: http.StatusTooManyRequests,
		Body:       "limit exceeded. Please retry in 2h6m56s.",
		RetryAfter: 2*time.Hour + 6*time.Minute + 56*time.Second,
	}
	inner := &scriptedAI{embedErr: quota}
	cooldowns := &memoryCooldowns{left: map[string]time.Duration{}}
	guard := newGuard(
		inner,
		&scriptedLimiter{decision: sharedTypes.RateLimitDecision{Allowed: true}},
		cooldowns,
	)

	_, err := guard.CreateEmbeddings(context.Background(), []string{"a job"})

	wait, deferred := utils.RetryAfterOf(err)
	require.True(t, deferred, "a quota error must tell the worker when to come back")
	assert.Equal(t, quota.RetryAfter, wait)
	require.ErrorAs(t, err, new(*utils.APIError), "the provider's own error stays in the chain")
	assert.Equal(t, quota.RetryAfter, cooldowns.left[embeddingsCooldownKey()])
	assert.Equal(t, 1, inner.embedCalls)

	_, err = guard.CreateEmbeddings(context.Background(), []string{"another job"})

	assert.ErrorIs(t, err, config.ErrAIPaused)
	assert.Equal(
		t,
		1,
		inner.embedCalls,
		"while paused the provider is not called, so no quota is wasted",
	)
	assert.Positive(t, guard.EmbeddingsPausedFor(context.Background()))

	reply, err := guard.CreateChatCompletion(
		context.Background(),
		sharedTypes.ChatCompletionRequest{},
	)
	require.NoError(t, err)
	assert.Equal(t, "reply", reply, "chat has its own quota and is not paused by embeddings")
}

func TestAIGuardUsesADefaultPauseAndCapsALongOne(t *testing.T) {
	setup.Config(t)

	allowed := &scriptedLimiter{decision: sharedTypes.RateLimitDecision{Allowed: true}}

	noHint := &scriptedAI{embedErr: &utils.APIError{StatusCode: http.StatusTooManyRequests}}
	cooldowns := &memoryCooldowns{left: map[string]time.Duration{}}
	_, _ = newGuard(
		noHint,
		allowed,
		cooldowns,
	).CreateEmbeddings(context.Background(), []string{"x"})
	assert.Equal(t, config.GetAICooldownDefault(), cooldowns.left[embeddingsCooldownKey()])

	tooLong := &scriptedAI{
		embedErr: &utils.APIError{
			StatusCode: http.StatusTooManyRequests,
			RetryAfter: 100 * 24 * time.Hour,
		},
	}
	cooldowns = &memoryCooldowns{left: map[string]time.Duration{}}
	_, _ = newGuard(
		tooLong,
		allowed,
		cooldowns,
	).CreateEmbeddings(context.Background(), []string{"x"})
	assert.Equal(t, config.GetAICooldownMax(), cooldowns.left[embeddingsCooldownKey()])
}

func TestAIGuardPacesByEstimatedTokensAndNeverAsksForMoreThanTheBudget(t *testing.T) {
	setup.Config(t)

	inner := &scriptedAI{}
	limiter := &scriptedLimiter{decision: sharedTypes.RateLimitDecision{Allowed: true}}
	guard := newGuard(inner, limiter, &memoryCooldowns{left: map[string]time.Duration{}})

	_, err := guard.CreateEmbeddings(context.Background(), []string{strings.Repeat("x", 80)})
	require.NoError(t, err)
	assert.Equal(t, []int{20}, limiter.costs, "80 characters are about 20 tokens")

	_, err = guard.CreateEmbeddings(context.Background(), []string{strings.Repeat("x", 40_000)})
	require.NoError(t, err)
	assert.Equal(t, 100, limiter.costs[1], "a request bigger than the whole budget is capped to it")

	limiter.decision = sharedTypes.RateLimitDecision{NextUnitIn: 7 * time.Second}

	_, err = guard.CreateEmbeddings(context.Background(), []string{"held back"})

	wait, deferred := utils.RetryAfterOf(err)
	require.True(t, deferred)
	assert.Equal(t, 7*time.Second, wait)
	assert.ErrorIs(t, err, config.ErrAIRateLimited)
	assert.Equal(t, 2, inner.embedCalls, "a request over budget never reaches the provider")
}

func TestOtherProviderErrorsPassThroughWithoutPausingAnything(t *testing.T) {
	setup.Config(t)

	rejected := &utils.APIError{StatusCode: http.StatusBadRequest, Body: "bad input"}
	cooldowns := &memoryCooldowns{left: map[string]time.Duration{}}
	guard := newGuard(
		&scriptedAI{embedErr: rejected},
		&scriptedLimiter{decision: sharedTypes.RateLimitDecision{Allowed: true}},
		cooldowns,
	)

	_, err := guard.CreateEmbeddings(context.Background(), []string{"x"})

	assert.Same(t, rejected, err)
	_, deferred := utils.RetryAfterOf(err)
	assert.False(t, deferred)
	assert.Empty(t, cooldowns.left)
}
