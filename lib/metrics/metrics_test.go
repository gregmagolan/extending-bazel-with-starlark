package metrics

import (
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

type stepClock struct {
	now  time.Time
	step time.Duration
}

func (c *stepClock) Now() time.Time {
	c.now = c.now.Add(c.step)
	return c.now
}

func TestCounters(t *testing.T) {
	r := NewRegistry(nil)
	r.Inc("requests", 1)
	r.Inc("requests", 4)
	if got := r.Counter("requests"); got != 5 {
		t.Fatalf("Counter() = %d, want 5", got)
	}
	if got := r.Counter("missing"); got != 0 {
		t.Errorf("Counter(missing) = %d", got)
	}
}

func TestTimeUsesInjectedClock(t *testing.T) {
	r := NewRegistry(&stepClock{step: 10 * time.Millisecond})
	r.Time("handler")()
	if got := r.Mean("handler"); got != 10*time.Millisecond {
		t.Fatalf("Mean() = %s, want 10ms", got)
	}
}

func TestSnapshotIsSorted(t *testing.T) {
	r := NewRegistry(nil)
	r.Inc("zebra", 1)
	r.Inc("apple", 2)
	got := r.Snapshot()
	if len(got) != 2 || got[0] != "apple=2" || got[1] != "zebra=1" {
		t.Fatalf("Snapshot() = %v", got)
	}
}

func TestObserveTracksMax(t *testing.T) {
	r := NewRegistry(nil)
	r.Observe("x", time.Second)
	r.Observe("x", 3*time.Second)
	if got := r.Mean("x"); got != 2*time.Second {
		t.Errorf("Mean() = %s, want 2s", got)
	}
	r.Dump(log.Discard())
}
