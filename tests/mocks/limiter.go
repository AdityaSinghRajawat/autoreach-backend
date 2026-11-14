package mocks

import (
	"context"
	"sync"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

type Limiter struct {
	mutex    sync.Mutex
	Decision sharedTypes.RateLimitDecision
	Err      error
	Calls    []LimiterCall
}

type LimiterCall struct {
	Policy  config.RateLimitPolicyName
	Subject string
	Cost    int
}

var _ sharedTypes.Limiter = (*Limiter)(nil)

func NewAllowingLimiter() *Limiter {
	return &Limiter{Decision: sharedTypes.RateLimitDecision{Allowed: true}}
}

func (l *Limiter) Allow(
	_ context.Context,
	policy config.RateLimitPolicy,
	subject string,
	cost int,
) (sharedTypes.RateLimitDecision, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.Calls = append(l.Calls, LimiterCall{Policy: policy.Name, Subject: subject, Cost: cost})

	return l.Decision, l.Err
}
