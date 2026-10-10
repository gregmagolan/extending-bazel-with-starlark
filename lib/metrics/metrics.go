// Package metrics is an in-process counter and timer registry.
//
// There is no exporter: the registry is snapshotted and logged, which is all
// the course needs. Timings are recorded against an injected clock so tests
// are not time-dependent.
package metrics

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/clock"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

// Registry holds named counters and timers. It is safe for concurrent use.
type Registry struct {
	mu       sync.Mutex
	clk      clock.Clock
	counters map[string]int64
	timers   map[string]*timer
}

type timer struct {
	count int64
	total time.Duration
	max   time.Duration
}

// NewRegistry returns an empty registry driven by clk. A nil clk means the
// system clock.
func NewRegistry(clk clock.Clock) *Registry {
	if clk == nil {
		clk = clock.System{}
	}
	return &Registry{
		clk:      clk,
		counters: map[string]int64{},
		timers:   map[string]*timer{},
	}
}

// Inc adds delta to the named counter.
func (r *Registry) Inc(name string, delta int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name] += delta
}

// Counter reads the named counter.
func (r *Registry) Counter(name string) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counters[name]
}

// Observe records a single duration against the named timer.
func (r *Registry) Observe(name string, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.timers[name]
	if !ok {
		t = &timer{}
		r.timers[name] = t
	}
	t.count++
	t.total += d
	if d > t.max {
		t.max = d
	}
}

// Time returns a function that, when called, records the elapsed time against
// the named timer. The intended use is `defer r.Time("handler")()`.
func (r *Registry) Time(name string) func() {
	start := r.clk.Now()
	return func() { r.Observe(name, r.clk.Now().Sub(start)) }
}

// Mean returns the average observed duration, or zero if nothing was observed.
func (r *Registry) Mean(name string) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.timers[name]
	if !ok || t.count == 0 {
		return 0
	}
	return t.total / time.Duration(t.count)
}

// Snapshot renders the registry as sorted "name=value" strings.
func (r *Registry) Snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]string, 0, len(r.counters)+len(r.timers))
	for name, v := range r.counters {
		out = append(out, fmt.Sprintf("%s=%d", name, v))
	}
	for name, t := range r.timers {
		mean := time.Duration(0)
		if t.count > 0 {
			mean = t.total / time.Duration(t.count)
		}
		out = append(out, fmt.Sprintf("%s.count=%d", name, t.count))
		out = append(out, fmt.Sprintf("%s.mean=%s", name, mean))
		out = append(out, fmt.Sprintf("%s.max=%s", name, t.max))
	}
	sort.Strings(out)
	return out
}

// Dump writes the snapshot to the logger, one field per metric.
func (r *Registry) Dump(l *log.Logger) {
	kv := make([]string, 0, 16)
	for _, s := range r.Snapshot() {
		for i := 0; i < len(s); i++ {
			if s[i] == '=' {
				kv = append(kv, s[:i], s[i+1:])
				break
			}
		}
	}
	l.Info("metrics", kv...)
}
