package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

func BuildKey(parts ...string) string {
	segments := make([]string, 0, len(parts)+1)
	segments = append(segments, config.GetKeyPrefix())
	segments = append(segments, parts...)

	return strings.Join(segments, config.GetKeyDelimiter())
}

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

// ClaimSlot takes a key that nobody else holds for the next ttl. It is never released, so the
// claim covers the whole slot even after the holder's work has finished.
func (r *RedisUtil) ClaimSlot(ctx context.Context, name string, ttl time.Duration) (bool, error) {
	claimed, err := r.client.SetNX(ctx, BuildKey(config.GetLockNamespace(), name), uuid.NewString(), ttl).
		Result()
	if err != nil {
		return false, fmt.Errorf("claim slot %s: %w", name, err)
	}

	return claimed, nil
}

func (r *RedisUtil) IncrementDaily(
	ctx context.Context,
	namespace, name string,
	amount int64,
) (int64, error) {
	key := BuildKey(namespace, name, GetCurrTimeStamp().Format(time.DateOnly))

	pipeline := r.client.TxPipeline()
	counter := pipeline.IncrBy(ctx, key, amount)
	pipeline.Expire(ctx, key, config.GetDailyCounterTTL())

	if _, err := pipeline.Exec(ctx); err != nil {
		return 0, fmt.Errorf("increment daily counter %s: %w", name, err)
	}

	return counter.Val(), nil
}

func (r *RedisUtil) GetDaily(ctx context.Context, namespace, name string) (int64, error) {
	value, err := r.client.Get(ctx, BuildKey(namespace, name, GetCurrTimeStamp().Format(time.DateOnly))).
		Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("read daily counter %s: %w", name, err)
	}

	return value, nil
}

// TakeSeat holds one of limit seats under name for at most ttl, and returns a release func. Seats
// expire on their own, so a process that dies holding one cannot keep it.
func (r *RedisUtil) TakeSeat(
	ctx context.Context,
	name string,
	limit int,
	ttl time.Duration,
) (bool, func(), error) {
	key := BuildKey(config.GetSeatNamespace(), name)
	seat := uuid.NewString()
	now := GetCurrTimeStamp()

	taken, err := takeSeatScriptI.Run(ctx, r.client, []string{key},
		now.UnixMilli(), limit, now.Add(ttl).UnixMilli(), seat, ttl.Milliseconds()).Int()
	if err != nil {
		return false, nil, fmt.Errorf("take seat %s: %w", name, err)
	}

	if taken == 0 {
		return false, nil, nil
	}

	return true, func() { _ = r.client.ZRem(context.WithoutCancel(ctx), key, seat).Err() }, nil
}
