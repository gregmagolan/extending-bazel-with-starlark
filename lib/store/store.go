// Package store is the in-memory record store the services read and write.
//
// It is not a database. It is a map with the behaviours that matter for the
// rest of the repo: schema validation on write, a capacity limit taken from
// config, revision numbers so queue can detect lost updates, and counters so
// the timing table in the course has something to show.
package store

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Sentinel errors callers are expected to match with errors.Is.
var (
	ErrNotFound = errors.New("store: not found")
	ErrConflict = errors.New("store: revision conflict")
	ErrFull     = errors.New("store: at capacity")
)

// Record is a stored value plus the bookkeeping the store adds.
type Record struct {
	Key      string
	Revision int64
	Fields   codec.Fields
}

// Clone returns a deep copy. Callers always get a copy so that a reader
// cannot mutate the stored map.
func (r Record) Clone() Record {
	out := r
	out.Fields = r.Fields.Clone()
	return out
}

// Store holds records in a single table, keyed by string.
type Store struct {
	mu      sync.RWMutex
	records map[string]Record
	max     int
	schemas map[string]schema.Schema
	log     *log.Logger
	metrics *metrics.Registry
}

// Open returns an empty store sized from "store.max_records".
func Open(l *log.Logger, cfg *config.Config, reg *metrics.Registry) (*Store, error) {
	max, err := cfg.Int("store.max_records", 10000)
	if err != nil {
		return nil, err
	}
	if max < 1 {
		return nil, fmt.Errorf("store: store.max_records %d must be positive", max)
	}
	if reg == nil {
		reg = metrics.NewRegistry(nil)
	}
	l.Debug("store opened", "path", cfg.String("store.path", ""), "max", fmt.Sprint(max))
	return &Store{
		records: make(map[string]Record, 64),
		max:     max,
		schemas: map[string]schema.Schema{},
		log:     l,
		metrics: reg,
	}, nil
}

// Register associates a schema with a key prefix. Writes to keys under that
// prefix are validated against it.
func (s *Store) Register(prefix string, sch schema.Schema) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schemas[prefix] = sch
}

// Len reports the number of stored records.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}

// Put writes fields at key. A non-zero expect must match the current
// revision, otherwise ErrConflict is returned. The new revision is returned.
func (s *Store) Put(key string, expect int64, fields codec.Fields) (int64, error) {
	defer s.metrics.Time("store.put")()
	if err := validate.NonEmpty("key", key); err != nil {
		return 0, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if sch, ok := s.schemaFor(key); ok {
		if err := validate.Record(sch, fields).Err(); err != nil {
			s.metrics.Inc("store.put.invalid", 1)
			return 0, fmt.Errorf("store: %s: %w", key, err)
		}
	}

	cur, exists := s.records[key]
	if !exists && len(s.records) >= s.max {
		s.metrics.Inc("store.put.full", 1)
		return 0, fmt.Errorf("%w: %d records", ErrFull, len(s.records))
	}
	if expect != 0 && cur.Revision != expect {
		s.metrics.Inc("store.put.conflict", 1)
		return 0, fmt.Errorf("%w: %s is at revision %d, expected %d", ErrConflict, key, cur.Revision, expect)
	}

	next := cur.Revision + 1
	s.records[key] = Record{Key: key, Revision: next, Fields: fields.Clone()}
	s.metrics.Inc("store.put.ok", 1)
	return next, nil
}

// Get returns a copy of the record at key.
func (s *Store) Get(key string) (Record, error) {
	defer s.metrics.Time("store.get")()
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.records[key]
	if !ok {
		s.metrics.Inc("store.get.miss", 1)
		return Record{}, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	s.metrics.Inc("store.get.hit", 1)
	return r.Clone(), nil
}

// Delete removes the record at key.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[key]; !ok {
		return fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	delete(s.records, key)
	s.metrics.Inc("store.delete.ok", 1)
	return nil
}

// List returns every record whose key starts with prefix, sorted by key.
func (s *Store) List(prefix string) []Record {
	defer s.metrics.Time("store.list")()
	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []Record
	for k, r := range s.records {
		if strings.HasPrefix(k, prefix) {
			out = append(out, r.Clone())
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// Keys returns every key, sorted.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.records))
	for k := range s.records {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Stats renders the store's metric snapshot.
func (s *Store) Stats() []string { return s.metrics.Snapshot() }

// schemaFor finds the longest registered prefix matching key. The caller must
// hold at least a read lock.
func (s *Store) schemaFor(key string) (schema.Schema, bool) {
	best := ""
	var found schema.Schema
	ok := false
	for prefix, sch := range s.schemas {
		if strings.HasPrefix(key, prefix) && len(prefix) >= len(best) {
			best, found, ok = prefix, sch, true
		}
	}
	return found, ok
}
