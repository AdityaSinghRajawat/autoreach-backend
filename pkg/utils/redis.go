package utils

import (
	"context"
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

var releaseLockScriptI = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

type RedisLock struct {
	key   string
	token string
}

type RedisUtil struct {
	client *redis.Client
}

func NewRedisUtil(client *redis.Client) (*RedisUtil, error) {
	if client == nil {
		return nil, fmt.Errorf("new redis util: %w", config.ErrMissingDependency)
	}

	return &RedisUtil{client: client}, nil
}

func (r *RedisUtil) AcquireLock(
	ctx context.Context,
	name string,
	ttl time.Duration,
) (*RedisLock, bool, error) {
	lock := &RedisLock{
		key:   BuildKey(config.GetLockNamespace(), name),
		token: uuid.NewString(),
	}

	acquired, err := r.client.SetNX(ctx, lock.key, lock.token, ttl).Result()
	if err != nil {
		return nil, false, fmt.Errorf("acquire lock %s: %w", name, err)
	}

	if !acquired {
		return nil, false, nil
	}

	return lock, true, nil
}

func (r *RedisUtil) ReleaseLock(ctx context.Context, lock *RedisLock) error {
	if lock == nil {
		return nil
	}

	if err := releaseLockScriptI.Run(ctx, r.client, []string{lock.key}, lock.token).
		Err(); err != nil {
		return fmt.Errorf("release lock: %w", err)
	}

	return nil
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
