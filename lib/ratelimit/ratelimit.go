// Package ratelimit is a token-bucket limiter keyed by caller.
//
// Buckets refill lazily from the injected clock, so there is no background
// goroutine and tests can advance time explicitly.
package ratelimit

import (
	"sync"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/clock"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
)

// Limiter hands out tokens to named callers.
type Limiter struct {
	mu      sync.Mutex
	clk     clock.Clock
	buckets map[string]*bucket
	burst   float64
	rate    float64 // tokens per second
	log     *log.Logger
	metrics *metrics.Registry
}

type bucket struct {
	tokens float64
	last   time.Time
}

// New returns a limiter allowing burst requests immediately and rate requests
// per second thereafter.
func New(l *log.Logger, reg *metrics.Registry, clk clock.Clock, burst int, rate float64) *Limiter {
	if clk == nil {
		clk = clock.System{}
	}
	if reg == nil {
		reg = metrics.NewRegistry(clk)
	}
	if burst < 1 {
		burst = 1
	}
	if rate <= 0 {
		rate = 1
	}
	return &Limiter{
		clk:     clk,
		buckets: map[string]*bucket{},
		burst:   float64(burst),
		rate:    rate,
		log:     l,
		metrics: reg,
	}
}

// Allow consumes one token for key, reporting whether the call may proceed.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.refill(key)
	if b.tokens < 1 {
		l.metrics.Inc("ratelimit.denied", 1)
		l.log.Debug("rate limited", "key", key)
		return false
	}
	b.tokens--
	l.metrics.Inc("ratelimit.allowed", 1)
	return true
}

// Tokens reports the tokens currently available to key. Reading refills the
// bucket first, so the answer reflects the clock rather than the last Allow.
func (l *Limiter) Tokens(key string) float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.refill(key).tokens
}

// refill brings key's bucket up to date with the clock, creating it if needed.
// The caller must hold l.mu.
func (l *Limiter) refill(key string) *bucket {
	now := l.clk.Now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}
	if elapsed := now.Sub(b.last); elapsed > 0 {
		b.tokens += elapsed.Seconds() * l.rate
		if b.tokens > l.burst {
			b.tokens = l.burst
		}
		b.last = now
	}
	return b
}

// Forget drops the bucket for key.
func (l *Limiter) Forget(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// Tracked reports how many distinct keys the limiter is holding state for.
func (l *Limiter) Tracked() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}
