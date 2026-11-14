package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type RedisConn struct {
	client *redis.Client
}

func ConnectRedis(ctx context.Context) (*RedisConn, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        config.GetRedisAddr(),
		Password:    config.GetRedisPassword(),
		DB:          config.GetRedisDB(),
		DialTimeout: config.GetRedisConnectTimeout(),
	})

	conn := &RedisConn{client: client}
	if err := conn.Ping(ctx); err != nil {
		_ = client.Close()

		return nil, err
	}

	return conn, nil
}

func (c *RedisConn) GetClient() *redis.Client {
	return c.client
}

func (c *RedisConn) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, config.GetRedisConnectTimeout())
	defer cancel()

	if err := c.client.Ping(pingCtx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}

	return nil
}

func (c *RedisConn) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("close redis: %w", err)
	}

	return nil
}
