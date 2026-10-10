package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
)

func testPolicy(attempts int, slept *[]time.Duration) Policy {
	return Policy{
		Attempts: attempts,
		Base:     10 * time.Millisecond,
		Max:      40 * time.Millisecond,
		Sleep:    func(d time.Duration) { *slept = append(*slept, d) },
	}
}

func TestSucceedsOnSecondAttempt(t *testing.T) {
	var slept []time.Duration
	calls := 0
	err := Do(log.Discard(), testPolicy(3, &slept), "op", func(int) error {
		calls++
		if calls < 2 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Do() = %v", err)
	}
	if calls != 2 || len(slept) != 1 || slept[0] != 10*time.Millisecond {
		t.Errorf("calls=%d slept=%v", calls, slept)
	}
}

func TestExhaustsAttempts(t *testing.T) {
	var slept []time.Duration
	calls := 0
	err := Do(log.Discard(), testPolicy(3, &slept), "op", func(int) error {
		calls++
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("Do() succeeded, want error")
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestPermanentStopsImmediately(t *testing.T) {
	var slept []time.Duration
	calls := 0
	sentinel := errors.New("nope")
	err := Do(log.Discard(), testPolicy(5, &slept), "op", func(int) error {
		calls++
		return Permanent(sentinel)
	})
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
	if !IsPermanent(err) || !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}
}

func TestDelayIsCapped(t *testing.T) {
	p := testPolicy(10, nil)
	for n, want := range map[int]time.Duration{1: 0, 2: 10 * time.Millisecond, 3: 20 * time.Millisecond, 4: 40 * time.Millisecond, 9: 40 * time.Millisecond} {
		if got := p.Delay(n); got != want {
			t.Errorf("Delay(%d) = %s, want %s", n, got, want)
		}
	}
}
