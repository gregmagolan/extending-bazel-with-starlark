// Package retry runs an operation again when it fails in a way worth retrying.
//
// The backoff is deterministic (no jitter) so that the worker's behaviour is
// reproducible in tests; the sleep function is injectable for the same reason.
package retry

import (
	"errors"
	"fmt"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/clock"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

// Policy describes how many attempts to make and how long to wait between them.
type Policy struct {
	// Attempts is the total number of tries, including the first one.
	Attempts int
	// Base is the delay before the second attempt.
	Base time.Duration
	// Max caps the delay between attempts.
	Max time.Duration
	// Sleep defaults to time.Sleep; tests replace it.
	Sleep func(time.Duration)
	// Clock is used only to stamp log lines.
	Clock clock.Clock
}

// Default returns a policy suitable for a background worker.
func Default() Policy {
	return Policy{Attempts: 4, Base: 50 * time.Millisecond, Max: time.Second}
}

// Permanent wraps err to mark it as not worth retrying.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return permanent{err}
}

type permanent struct{ err error }

func (p permanent) Error() string { return "permanent: " + p.err.Error() }
func (p permanent) Unwrap() error { return p.err }

// IsPermanent reports whether err was marked with Permanent.
func IsPermanent(err error) bool {
	var p permanent
	return errors.As(err, &p)
}

// Delay returns the wait before attempt n, where n is 1-based. Attempt 1 has
// no delay; later attempts double the base, clamped to Max.
func (p Policy) Delay(n int) time.Duration {
	if n <= 1 {
		return 0
	}
	d := p.Base
	for i := 2; i < n; i++ {
		d *= 2
		if p.Max > 0 && d >= p.Max {
			return p.Max
		}
	}
	if p.Max > 0 && d > p.Max {
		return p.Max
	}
	return d
}

// Do calls op until it succeeds, until the policy runs out of attempts, or
// until op returns a Permanent error. The returned error wraps the last
// failure and names the operation.
func Do(l *log.Logger, p Policy, name string, op func(attempt int) error) error {
	if p.Attempts < 1 {
		p.Attempts = 1
	}
	sleep := p.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}

	var last error
	for attempt := 1; attempt <= p.Attempts; attempt++ {
		if d := p.Delay(attempt); d > 0 {
			sleep(d)
		}
		err := op(attempt)
		if err == nil {
			if attempt > 1 && l != nil {
				l.Info("retry succeeded", "op", name, "attempt", fmt.Sprint(attempt))
			}
			return nil
		}
		last = err
		if IsPermanent(err) {
			if l != nil {
				l.Error("retry abandoned", "op", name, "err", err.Error())
			}
			return err
		}
		if l != nil {
			l.Warn("retry failed", "op", name, "attempt", fmt.Sprint(attempt), "err", err.Error())
		}
	}
	return fmt.Errorf("retry: %s failed after %d attempts: %w", name, p.Attempts, last)
}
