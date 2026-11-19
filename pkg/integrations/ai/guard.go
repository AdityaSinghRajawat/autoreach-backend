package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Guard keeps calls to the AI provider inside its limits: calls are paced by estimated tokens,
// and a 429 pauses that operation for everyone until the provider says it will accept calls again.
type Guard struct {
	inner     sharedTypes.AIClient
	limiter   sharedTypes.Limiter
	cooldowns sharedTypes.ExpiringKeyStore
	policy    config.RateLimitPolicy
}

var (
	_ sharedTypes.AIClient      = (*Guard)(nil)
	_ sharedTypes.EmbeddingGate = (*Guard)(nil)
)

func NewGuard(
	inner sharedTypes.AIClient,
	limiter sharedTypes.Limiter,
	cooldowns sharedTypes.ExpiringKeyStore,
	policy config.RateLimitPolicy,
) *Guard {
	return &Guard{inner: inner, limiter: limiter, cooldowns: cooldowns, policy: policy}
}

func (g *Guard) Model() string {
	return g.inner.Model()
}

func (g *Guard) ChatModel() string {
	return g.inner.ChatModel()
}

func (g *Guard) CreateEmbeddings(ctx context.Context, inputs []string) ([][]float32, error) {
	var vectors [][]float32

	err := g.guard(ctx, config.GetAIEmbeddingsOperation(), estimateTokens(inputs...), func() error {
		var callErr error

		vectors, callErr = g.inner.CreateEmbeddings(ctx, inputs)

		return callErr
	})

	return vectors, err
}

func (g *Guard) CreateChatCompletion(
	ctx context.Context,
	request sharedTypes.ChatCompletionRequest,
) (string, error) {
	texts := make([]string, 0, len(request.Messages))
	for _, message := range request.Messages {
		texts = append(texts, message.Content)
	}

	var reply string

	err := g.guard(ctx, config.GetAIChatOperation(), estimateTokens(texts...), func() error {
		var callErr error

		reply, callErr = g.inner.CreateChatCompletion(ctx, request)

		return callErr
	})

	return reply, err
}

func (g *Guard) EmbeddingsPausedFor(ctx context.Context) time.Duration {
	return g.pausedFor(ctx, config.GetAIEmbeddingsOperation())
}

func (g *Guard) guard(
	ctx context.Context,
	operation string,
	tokens int,
	call func() error,
) error {
	if left := g.pausedFor(ctx, operation); left > 0 {
		return &utils.RetryAfterError{
			Err:   fmt.Errorf("%s: %w", operation, config.ErrAIPaused),
			After: left,
		}
	}

	if g.policy.IsEnabled() {
		// A request larger than the whole budget could never be allowed, so it is capped to it.
		decision, err := g.limiter.Allow(ctx, g.policy, operation, min(tokens, g.policy.Burst))

		switch {
		case err != nil:
			utils.LogWarn(ctx, "ai rate limiter unavailable, allowing call", "error", err.Error())
		case !decision.Allowed:
			return &utils.RetryAfterError{
				Err:   fmt.Errorf("%s: %w", operation, config.ErrAIRateLimited),
				After: decision.NextUnitIn,
			}
		}
	}

	err := call()

	var apiErr *utils.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTooManyRequests {
		return err
	}

	pause := apiErr.RetryAfter
	if pause <= 0 {
		pause = config.GetAICooldownDefault()
	}

	pause = min(pause, config.GetAICooldownMax())

	if setErr := g.cooldowns.SetKey(ctx, cooldownKey(operation), "1", pause); setErr != nil {
		utils.LogWarn(ctx, "ai cooldown not recorded", "error", setErr.Error())
	} else {
		utils.LogWarn(ctx, "ai provider quota reached, pausing calls",
			"operation", operation, "for", pause.Round(time.Second).String())
	}

	return &utils.RetryAfterError{Err: err, After: pause}
}

func (g *Guard) pausedFor(ctx context.Context, operation string) time.Duration {
	left, err := g.cooldowns.GetTTL(ctx, cooldownKey(operation))
	if err != nil {
		utils.LogWarn(ctx, "ai cooldown unavailable, allowing call", "error", err.Error())

		return 0
	}

	return left
}

func cooldownKey(operation string) string {
	return utils.BuildKey(
		config.GetKeyDelimiter(),
		config.GetKeyPrefix(),
		config.GetCooldownNamespace(),
		operation,
	)
}

func estimateTokens(texts ...string) int {
	characters := 0
	for _, text := range texts {
		characters += utf8.RuneCountInString(text)
	}

	perToken := config.GetAICharsPerToken()

	return max((characters+perToken-1)/perToken, 1)
}
