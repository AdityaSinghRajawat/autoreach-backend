package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

var takeSeatScriptI = redis.NewScript(`
local now = tonumber(ARGV[1])
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now)
if redis.call("ZCARD", KEYS[1]) >= tonumber(ARGV[2]) then
	return 0
end
redis.call("ZADD", KEYS[1], ARGV[3], ARGV[4])
redis.call("PEXPIRE", KEYS[1], ARGV[5])
return 1
`)

type RedisUtil struct {
	client *redis.Client
}

func NewRedisUtil(client *redis.Client) (*RedisUtil, error) {
	if client == nil {
		return nil, fmt.Errorf("new redis util: %w", config.ErrMissingDependency)
	}

	return &RedisUtil{client: client}, nil
}

func (r *RedisUtil) SetKey(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("set key %s: %w", key, err)
	}

	return nil
}

func (r *RedisUtil) SetIfNotExists(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) (bool, error) {
	created, err := r.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("set key %s if not exists: %w", key, err)
	}

	return created, nil
}

// A missing key reads as an empty string.
func (r *RedisUtil) GetKey(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("get key %s: %w", key, err)
	}

	return value, nil
}

// A missing key reads as 0.
func (r *RedisUtil) GetInt(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("get integer key %s: %w", key, err)
	}

	return value, nil
}

// A missing key, or one that never expires, has no time left.
func (r *RedisUtil) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	left, err := r.client.PTTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("get ttl of key %s: %w", key, err)
	}

	return max(left, 0), nil
}

func (r *RedisUtil) DeleteKey(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete key %s: %w", key, err)
	}

	return nil
}

// The expiry is set in the same transaction, so a counter never outlives its window.
func (r *RedisUtil) IncrementKey(
	ctx context.Context,
	key string,
	amount int64,
	expiration time.Duration,
) (int64, error) {
	pipeline := r.client.TxPipeline()
	counter := pipeline.IncrBy(ctx, key, amount)

	if expiration > 0 {
		pipeline.Expire(ctx, key, expiration)
	}

	if _, err := pipeline.Exec(ctx); err != nil {
		return 0, fmt.Errorf("increment key %s: %w", key, err)
	}

	return counter.Val(), nil
}

// Seats expire on their own, so a crashed process cannot keep one.
func (r *RedisUtil) TakeSeat(
	ctx context.Context,
	key string,
	limit int,
	ttl time.Duration,
) (bool, func(), error) {
	seat := GenerateUUID()
	now := GetCurrTimeStamp()

	taken, err := takeSeatScriptI.Run(ctx, r.client, []string{key},
		now.UnixMilli(), limit, now.Add(ttl).UnixMilli(), seat, ttl.Milliseconds()).Int()
	if err != nil {
		return false, nil, fmt.Errorf("take seat %s: %w", key, err)
	}

	if taken == 0 {
		return false, nil, nil
	}

	return true, func() { _ = r.client.ZRem(context.WithoutCancel(ctx), key, seat).Err() }, nil
}
