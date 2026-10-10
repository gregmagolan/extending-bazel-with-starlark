// Package queue is the lease-based work queue that sits between the server
// and the worker.
//
// Jobs are leased, not popped: a lease expires after a configurable interval
// and the job becomes claimable again, which is what makes the worker safe to
// kill. Durable state lives in store, so a job survives a restart.
package queue

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/clock"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/retry"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Sentinel errors callers are expected to match with errors.Is.
var (
	ErrEmpty       = errors.New("queue: no claimable job")
	ErrFull        = errors.New("queue: at depth limit")
	ErrNoSuchLease = errors.New("queue: unknown lease")
)

// KeyPrefix is the store key prefix under which jobs are persisted.
const KeyPrefix = "queue/job/"

// Job is one unit of work.
type Job struct {
	ID       string
	Kind     string
	Payload  codec.Fields
	Attempts int

	leaseUntil time.Time
	leaseID    string
	done       bool
}

// Queue holds jobs in memory and mirrors them into a store.
type Queue struct {
	mu      sync.Mutex
	jobs    map[string]*Job
	order   []string
	seq     int64
	depth   int
	lease   time.Duration
	clk     clock.Clock
	log     *log.Logger
	metrics *metrics.Registry
	store   *store.Store
}

// Options configures a Queue. Zero fields fall back to sensible defaults.
type Options struct {
	Depth int
	Lease time.Duration
	Clock clock.Clock
	Store *store.Store
}

// New returns an empty queue.
func New(l *log.Logger, reg *metrics.Registry, opts Options) *Queue {
	if opts.Depth < 1 {
		opts.Depth = 256
	}
	if opts.Lease <= 0 {
		opts.Lease = 30 * time.Second
	}
	if opts.Clock == nil {
		opts.Clock = clock.System{}
	}
	if reg == nil {
		reg = metrics.NewRegistry(opts.Clock)
	}
	return &Queue{
		jobs:    map[string]*Job{},
		depth:   opts.Depth,
		lease:   opts.Lease,
		clk:     opts.Clock,
		log:     l,
		metrics: reg,
		store:   opts.Store,
	}
}

// Len reports the number of jobs that have not been acknowledged.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := 0
	for _, j := range q.jobs {
		if !j.done {
			n++
		}
	}
	return n
}

// Submit enqueues a job of the given kind and returns its ID.
func (q *Queue) Submit(kind string, payload codec.Fields) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.order) >= q.depth {
		q.metrics.Inc("queue.submit.full", 1)
		return "", fmt.Errorf("%w: %d jobs", ErrFull, len(q.order))
	}

	q.seq++
	id := kind + "-" + strconv.FormatInt(q.seq, 10)
	job := &Job{ID: id, Kind: kind, Payload: payload.Clone()}
	q.jobs[id] = job
	q.order = append(q.order, id)
	q.metrics.Inc("queue.submit.ok", 1)

	if q.store != nil {
		if _, err := q.store.Put(KeyPrefix+id, 0, q.persist(job)); err != nil {
			q.log.Warn("job not persisted", "id", id, "err", err.Error())
		}
	}
	q.log.Debug("job submitted", "id", id, "kind", kind)
	return id, nil
}

// Claim leases the oldest claimable job. The returned lease token must be
// passed to Ack or Nack.
func (q *Queue) Claim(worker string) (Job, string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clk.Now()
	for _, id := range q.order {
		job := q.jobs[id]
		if job == nil || job.done {
			continue
		}
		if job.leaseUntil.After(now) {
			continue
		}
		q.seq++
		job.leaseID = worker + "/" + strconv.FormatInt(q.seq, 10)
		job.leaseUntil = now.Add(q.lease)
		job.Attempts++
		q.metrics.Inc("queue.claim.ok", 1)
		return q.snapshot(job), job.leaseID, nil
	}
	q.metrics.Inc("queue.claim.empty", 1)
	return Job{}, "", ErrEmpty
}

// Ack marks a leased job as finished.
func (q *Queue) Ack(id, lease string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.jobs[id]
	if !ok || job.leaseID != lease {
		return fmt.Errorf("%w: %s", ErrNoSuchLease, lease)
	}
	job.done = true
	q.metrics.Inc("queue.ack.ok", 1)
	if q.store != nil {
		if err := q.store.Delete(KeyPrefix + id); err != nil {
			q.log.Debug("job already absent from store", "id", id)
		}
	}
	q.order = removeID(q.order, id)
	delete(q.jobs, id)
	return nil
}

// Nack releases a lease so the job can be claimed again immediately.
func (q *Queue) Nack(id, lease string, cause error) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.jobs[id]
	if !ok || job.leaseID != lease {
		return fmt.Errorf("%w: %s", ErrNoSuchLease, lease)
	}
	job.leaseUntil = time.Time{}
	job.leaseID = ""
	q.metrics.Inc("queue.nack", 1)
	if cause != nil {
		q.log.Warn("job returned to the queue", "id", id, "err", cause.Error())
	}
	return nil
}

// Expired returns the IDs of jobs whose lease has lapsed while still held.
func (q *Queue) Expired() []string {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clk.Now()
	var out []string
	for id, job := range q.jobs {
		if job.done || job.leaseID == "" {
			continue
		}
		if !job.leaseUntil.After(now) {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// Drain claims and processes jobs until the queue is empty, retrying each job
// under the given policy. It returns the number of jobs acknowledged.
func (q *Queue) Drain(worker string, policy retry.Policy, handle func(Job) error) (int, error) {
	done := 0
	for {
		job, lease, err := q.Claim(worker)
		if errors.Is(err, ErrEmpty) {
			return done, nil
		}
		if err != nil {
			return done, err
		}
		err = retry.Do(q.log, policy, "job:"+job.Kind, func(int) error { return handle(job) })
		if err != nil {
			if nerr := q.Nack(job.ID, lease, err); nerr != nil {
				return done, nerr
			}
			return done, err
		}
		if err := q.Ack(job.ID, lease); err != nil {
			return done, err
		}
		done++
	}
}

// Stats renders the queue's metric snapshot.
func (q *Queue) Stats() []string { return q.metrics.Snapshot() }

// snapshot copies a job for handing to a caller. The caller must hold q.mu.
func (q *Queue) snapshot(j *Job) Job {
	out := *j
	out.Payload = j.Payload.Clone()
	return out
}

// persist renders a job as a store record. The caller must hold q.mu.
func (q *Queue) persist(j *Job) codec.Fields {
	f := j.Payload.Clone()
	f["_id"] = j.ID
	f["_kind"] = j.Kind
	f["_attempts"] = strconv.Itoa(j.Attempts)
	return f
}

func removeID(ids []string, id string) []string {
	for i, got := range ids {
		if got == id {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
}
