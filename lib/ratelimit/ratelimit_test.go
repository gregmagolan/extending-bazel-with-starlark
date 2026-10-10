package ratelimit

import (
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

type manualClock struct{ now time.Time }

func (c *manualClock) Now() time.Time          { return c.now }
func (c *manualClock) advance(d time.Duration) { c.now = c.now.Add(d) }

func TestBurstThenDeny(t *testing.T) {
	clk := &manualClock{now: time.Unix(1000, 0)}
	l := New(log.Discard(), nil, clk, 3, 1)
	for i := 0; i < 3; i++ {
		if !l.Allow("a") {
			t.Fatalf("call %d denied inside the burst", i)
		}
	}
	if l.Allow("a") {
		t.Fatal("call past the burst was allowed")
	}
}

func TestRefill(t *testing.T) {
	clk := &manualClock{now: time.Unix(1000, 0)}
	l := New(log.Discard(), nil, clk, 1, 2)
	if !l.Allow("a") || l.Allow("a") {
		t.Fatal("burst of 1 behaved unexpectedly")
	}
	clk.advance(500 * time.Millisecond)
	if !l.Allow("a") {
		t.Fatal("token did not refill after 500ms at 2/s")
	}
}

func TestRefillIsCappedAtBurst(t *testing.T) {
	clk := &manualClock{now: time.Unix(1000, 0)}
	l := New(log.Discard(), nil, clk, 2, 100)
	l.Allow("a")
	clk.advance(time.Hour)
	if got := l.Tokens("a"); got != 2 {
		t.Fatalf("Tokens() = %v, want 2", got)
	}
}

func TestKeysAreIndependent(t *testing.T) {
	clk := &manualClock{now: time.Unix(1000, 0)}
	l := New(log.Discard(), nil, clk, 1, 1)
	if !l.Allow("a") || !l.Allow("b") {
		t.Fatal("separate keys shared a bucket")
	}
	if l.Tracked() != 2 {
		t.Errorf("Tracked() = %d, want 2", l.Tracked())
	}
	l.Forget("a")
	if l.Tracked() != 1 {
		t.Errorf("Tracked() after Forget = %d, want 1", l.Tracked())
	}
}

func TestDegenerateArgumentsAreClamped(t *testing.T) {
	l := New(log.Discard(), nil, nil, 0, 0)
	if !l.Allow("a") {
		t.Error("a zero burst did not clamp to 1")
	}
}
