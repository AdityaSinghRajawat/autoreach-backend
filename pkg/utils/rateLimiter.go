package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

var fixedWindowScriptI = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
	redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
local ttl = redis.call("PTTL", KEYS[1])
return {current, ttl}
`)

type RateLimitDecision struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) (*RateLimiter, error) {
	if client == nil {
		return nil, fmt.Errorf("new rate limiter: %w", config.ErrMissingDependency)
	}

	return &RateLimiter{client: client}, nil
}

func (r *RateLimiter) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (RateLimitDecision, error) {
	if limit <= 0 {
		return RateLimitDecision{Allowed: true}, nil
	}

	redisKey := BuildKey(config.GetRateLimitNamespace(), key)

	values, err := fixedWindowScriptI.Run(ctx, r.client, []string{redisKey}, window.Milliseconds()).
		Int64Slice()
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("run rate limit script: %w", err)
	}

	if len(values) != 2 {
		return RateLimitDecision{}, fmt.Errorf(
			"run rate limit script: %w",
			config.ErrUnexpectedRedisReply,
		)
	}

	current := int(values[0])

	remaining := limit - current
	if remaining < 0 {
		remaining = 0
	}

	return RateLimitDecision{
		Allowed:    current <= limit,
		Remaining:  remaining,
		RetryAfter: time.Duration(values[1]) * time.Millisecond,
	}, nil
}
