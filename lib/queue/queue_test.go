package queue

import (
	"errors"
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/retry"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

type manualClock struct{ now time.Time }

func (c *manualClock) Now() time.Time          { return c.now }
func (c *manualClock) advance(d time.Duration) { c.now = c.now.Add(d) }

func newQueue(t *testing.T, opts Options) (*Queue, *manualClock) {
	t.Helper()
	clk := &manualClock{now: time.Unix(1_700_000_000, 0)}
	if opts.Clock == nil {
		opts.Clock = clk
	}
	return New(log.Discard(), nil, opts), clk
}

func newStore(t *testing.T) *store.Store {
	t.Helper()
	cfg, err := config.Parse("")
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(log.Discard(), cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSubmitClaimAck(t *testing.T) {
	q, _ := newQueue(t, Options{})
	id, err := q.Submit("resize", codec.Fields{"path": "/a.png"})
	if err != nil {
		t.Fatalf("Submit() = %v", err)
	}
	if q.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", q.Len())
	}

	job, lease, err := q.Claim("w1")
	if err != nil {
		t.Fatalf("Claim() = %v", err)
	}
	if job.ID != id || job.Attempts != 1 || job.Payload["path"] != "/a.png" {
		t.Fatalf("Claim() = %+v", job)
	}
	if _, _, err := q.Claim("w2"); !errors.Is(err, ErrEmpty) {
		t.Fatalf("second Claim() = %v, want ErrEmpty", err)
	}
	if err := q.Ack(id, lease); err != nil {
		t.Fatalf("Ack() = %v", err)
	}
	if q.Len() != 0 {
		t.Errorf("Len() after Ack = %d, want 0", q.Len())
	}
}

func TestLeaseExpiryMakesJobClaimableAgain(t *testing.T) {
	q, clk := newQueue(t, Options{Lease: time.Minute})
	if _, err := q.Submit("resize", nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := q.Claim("w1"); err != nil {
		t.Fatal(err)
	}
	if got := q.Expired(); len(got) != 0 {
		t.Fatalf("Expired() = %v, want none", got)
	}

	clk.advance(2 * time.Minute)
	if got := q.Expired(); len(got) != 1 {
		t.Fatalf("Expired() = %v, want one", got)
	}
	job, _, err := q.Claim("w2")
	if err != nil {
		t.Fatalf("Claim() after expiry = %v", err)
	}
	if job.Attempts != 2 {
		t.Errorf("Attempts = %d, want 2", job.Attempts)
	}
}

func TestNackReleasesImmediately(t *testing.T) {
	q, _ := newQueue(t, Options{Lease: time.Hour})
	id, _ := q.Submit("resize", nil)
	_, lease, err := q.Claim("w1")
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Nack(id, lease, errors.New("downstream 503")); err != nil {
		t.Fatalf("Nack() = %v", err)
	}
	if _, _, err := q.Claim("w1"); err != nil {
		t.Fatalf("Claim() after Nack = %v", err)
	}
	if err := q.Ack(id, "stale-lease"); !errors.Is(err, ErrNoSuchLease) {
		t.Errorf("Ack(stale) = %v", err)
	}
}

func TestDepthLimit(t *testing.T) {
	q, _ := newQueue(t, Options{Depth: 2})
	for i := 0; i < 2; i++ {
		if _, err := q.Submit("resize", nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := q.Submit("resize", nil); !errors.Is(err, ErrFull) {
		t.Fatalf("Submit() over depth = %v, want ErrFull", err)
	}
}

func TestSubmitMirrorsIntoStore(t *testing.T) {
	s := newStore(t)
	q, _ := newQueue(t, Options{Store: s})
	id, err := q.Submit("resize", codec.Fields{"path": "/a.png"})
	if err != nil {
		t.Fatal(err)
	}
	rec, err := s.Get(KeyPrefix + id)
	if err != nil {
		t.Fatalf("store.Get() = %v", err)
	}
	if rec.Fields["_kind"] != "resize" || rec.Fields["path"] != "/a.png" {
		t.Fatalf("persisted record = %+v", rec.Fields)
	}

	_, lease, err := q.Claim("w1")
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Ack(id, lease); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(KeyPrefix + id); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("record survived Ack: %v", err)
	}
}

func TestDrainRetriesThenAcks(t *testing.T) {
	q, _ := newQueue(t, Options{})
	for i := 0; i < 3; i++ {
		if _, err := q.Submit("resize", nil); err != nil {
			t.Fatal(err)
		}
	}
	policy := retry.Policy{Attempts: 3, Base: time.Millisecond, Sleep: func(time.Duration) {}}

	seen := map[string]int{}
	done, err := q.Drain("w1", policy, func(j Job) error {
		seen[j.ID]++
		if seen[j.ID] < 2 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Drain() = %v", err)
	}
	if done != 3 || q.Len() != 0 {
		t.Fatalf("Drain() acked %d, Len = %d", done, q.Len())
	}
}

func TestDrainSurfacesPermanentFailure(t *testing.T) {
	q, _ := newQueue(t, Options{})
	if _, err := q.Submit("resize", nil); err != nil {
		t.Fatal(err)
	}
	policy := retry.Policy{Attempts: 2, Base: time.Millisecond, Sleep: func(time.Duration) {}}
	if _, err := q.Drain("w1", policy, func(Job) error {
		return retry.Permanent(errors.New("bad payload"))
	}); err == nil {
		t.Fatal("Drain() hid a permanent failure")
	}
	if q.Len() != 1 {
		t.Errorf("Len() = %d, want the job back on the queue", q.Len())
	}
}

// TestLeaseRenewalRace exercises the window between a lease expiring and the
// next worker picking the job up.
//
// A lease that has expired must be claimable by a second worker.
func TestLeaseRenewalRace(t *testing.T) {
	q, clk := newQueue(t, Options{Lease: 50 * time.Millisecond})
	id, err := q.Submit("resize", codec.Fields{"path": "/a.png"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := q.Claim("w1"); err != nil {
		t.Fatal(err)
	}
	clk.advance(60 * time.Millisecond)

	job, _, err := q.Claim("w2")
	if err != nil {
		t.Fatalf("Claim() after expiry = %v", err)
	}
	if job.ID != id {
		t.Errorf("Claim() = %s, want %s", job.ID, id)
	}
}
