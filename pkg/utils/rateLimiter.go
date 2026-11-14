package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

// GCRA: stores each subject's theoretical arrival time (ms, Redis clock shared by all replicas).
var gcraScriptI = redis.NewScript(`
local limit = tonumber(ARGV[1])
local period = tonumber(ARGV[2])
local burst = tonumber(ARGV[3])
local cost = tonumber(ARGV[4])

local clock = redis.call("TIME")
local now = tonumber(clock[1]) * 1000 + tonumber(clock[2]) / 1000
local interval = period / limit
local burst_offset = interval * burst

local tat = tonumber(redis.call("GET", KEYS[1])) or now
if tat < now then
	tat = now
end

local new_tat = tat + interval * math.max(cost, 1)
local allow_at = new_tat - burst_offset
if allow_at > now then
	return {0, 0, math.ceil(allow_at - now)}
end

if cost > 0 then
	tat = new_tat
	redis.call("SET", KEYS[1], string.format("%.3f", tat), "PX", math.ceil(tat - now))
end

local next_unit = 0
if tat > now then
	next_unit = (tat - now) % interval
	if next_unit == 0 then
		next_unit = interval
	end
end

return {1, math.floor((now - tat + burst_offset) / interval), math.ceil(next_unit)}
`)

type RateLimiter struct {
	client *redis.Client
}

var _ sharedTypes.Limiter = (*RateLimiter)(nil)

func NewRateLimiter(client *redis.Client) (*RateLimiter, error) {
	if client == nil {
		return nil, fmt.Errorf("new rate limiter: %w", config.ErrMissingDependency)
	}

	return &RateLimiter{client: client}, nil
}

func (r *RateLimiter) Allow(
	ctx context.Context,
	policy config.RateLimitPolicy,
	subject string,
	cost int,
) (sharedTypes.RateLimitDecision, error) {
	if !policy.IsEnabled() {
		return sharedTypes.RateLimitDecision{Allowed: true, Remaining: policy.Limit}, nil
	}

	values, err := gcraScriptI.Run(
		ctx,
		r.client,
		[]string{BuildKey(config.GetRateLimitNamespace(), string(policy.Name), subject)},
		policy.Limit,
		policy.Period.Milliseconds(),
		max(policy.Burst, 1),
		max(cost, 0),
	).Int64Slice()
	if err != nil {
		return sharedTypes.RateLimitDecision{}, fmt.Errorf("run rate limit script: %w", err)
	}

	if len(values) != 3 {
		return sharedTypes.RateLimitDecision{}, fmt.Errorf(
			"run rate limit script: %w",
			config.ErrUnexpectedRedisReply,
		)
	}

	return sharedTypes.RateLimitDecision{
		Allowed:    values[0] == 1,
		Remaining:  int(values[1]),
		NextUnitIn: time.Duration(values[2]) * time.Millisecond,
	}, nil
}
