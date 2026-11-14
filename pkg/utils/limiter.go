package utils

import (
	"context"
	"time"
)

type Limiter interface {
	Allow(
		ctx context.Context,
		key string,
		limit int,
		window time.Duration,
	) (RateLimitDecision, error)
}
