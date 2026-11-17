package utils

import (
	"sync"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type breakerEvent struct {
	at     time.Time
	failed bool
}

// After the cooldown one probe decides whether the circuit closes.
type CircuitBreaker struct {
	mutex          sync.Mutex
	window         time.Duration
	cooldown       time.Duration
	minRequests    int
	failureRate    float64
	events         []breakerEvent
	openedAt       time.Time
	probeStartedAt time.Time
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		window:      config.GetBreakerWindow(),
		cooldown:    config.GetBreakerCooldown(),
		minRequests: config.GetBreakerMinRequests(),
		failureRate: config.GetBreakerFailureRate(),
	}
}

func (b *CircuitBreaker) Allow() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.openedAt.IsZero() {
		return true
	}

	now := time.Now()
	if now.Sub(b.openedAt) < b.cooldown {
		return false
	}

	// A probe that never reported back (its caller gave up first) must not block recovery.
	if !b.probeStartedAt.IsZero() && now.Sub(b.probeStartedAt) < b.cooldown {
		return false
	}

	b.probeStartedAt = now

	return true
}

func (b *CircuitBreaker) Record(failed, tripImmediately bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	now := time.Now()

	if !b.probeStartedAt.IsZero() {
		b.probeStartedAt = time.Time{}
		b.events = nil

		if failed || tripImmediately {
			b.openedAt = now
		} else {
			b.openedAt = time.Time{}
		}

		return
	}

	if tripImmediately {
		b.openedAt = now

		return
	}

	b.events = append(b.prune(now), breakerEvent{at: now, failed: failed})
	if len(b.events) < b.minRequests {
		return
	}

	failures := 0

	for _, event := range b.events {
		if event.failed {
			failures++
		}
	}

	if float64(failures)/float64(len(b.events)) > b.failureRate {
		b.openedAt = now
	}
}

func (b *CircuitBreaker) prune(now time.Time) []breakerEvent {
	kept := b.events[:0]

	for _, event := range b.events {
		if now.Sub(event.at) <= b.window {
			kept = append(kept, event)
		}
	}

	return kept
}
