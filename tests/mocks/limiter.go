package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type Limiter struct {
	mutex    sync.Mutex
	Decision utils.RateLimitDecision
	Err      error
	Keys     []string
}

var _ utils.Limiter = (*Limiter)(nil)

func NewAllowingLimiter() *Limiter {
	return &Limiter{Decision: utils.RateLimitDecision{Allowed: true}}
}

func (l *Limiter) Allow(
	_ context.Context,
	key string,
	_ int,
	_ time.Duration,
) (utils.RateLimitDecision, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.Keys = append(l.Keys, key)

	return l.Decision, l.Err
}
