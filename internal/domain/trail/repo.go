package trail

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists trail entities into a store.Store.
//
// Every method is the same three steps -- validate, encode, write -- repeated
// per entity type. The repetition is deliberate: it is what gives the
// reverse-dependency queries in the course a non-trivial answer.
type Repo struct {
	store   *store.Store
	log     *log.Logger
	metrics *metrics.Registry
}

// NewRepo wires a Repo and registers every schema this package declares.
func NewRepo(l *log.Logger, reg *metrics.Registry, s *store.Store) *Repo {
	if reg == nil {
		reg = metrics.NewRegistry(nil)
	}
	for i, sch := range Schemas {
		s.Register(Prefixes[i], sch)
	}
	l.Debug("trail repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many trail records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutTrailRecord writes x, creating it when revision is zero.
func (r *Repo) PutTrailRecord(x TrailRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailrecord.written", 1)
	return rev, nil
}

// GetTrailRecord reads the entity with the given id.
func (r *Repo) GetTrailRecord(id string) (TrailRecord, int64, error) {
	rec, err := r.store.Get("trail/trailrecord/" + id)
	if err != nil {
		return TrailRecord{}, 0, err
	}
	x, err := ParseTrailRecord(rec.Fields)
	if err != nil {
		return TrailRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailRecords reads every stored entity of this type, sorted.
func (r *Repo) ListTrailRecords() ([]TrailRecord, error) {
	recs := r.store.List("trail/trailrecord/")
	out := make([]TrailRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailRecords(out)
	return out, nil
}

// PutTrailEntry writes x, creating it when revision is zero.
func (r *Repo) PutTrailEntry(x TrailEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailentry.written", 1)
	return rev, nil
}

// GetTrailEntry reads the entity with the given id.
func (r *Repo) GetTrailEntry(id string) (TrailEntry, int64, error) {
	rec, err := r.store.Get("trail/trailentry/" + id)
	if err != nil {
		return TrailEntry{}, 0, err
	}
	x, err := ParseTrailEntry(rec.Fields)
	if err != nil {
		return TrailEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListTrailEntrys() ([]TrailEntry, error) {
	recs := r.store.List("trail/trailentry/")
	out := make([]TrailEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailEntrys(out)
	return out, nil
}

// PutTrailItem writes x, creating it when revision is zero.
func (r *Repo) PutTrailItem(x TrailItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailitem.written", 1)
	return rev, nil
}

// GetTrailItem reads the entity with the given id.
func (r *Repo) GetTrailItem(id string) (TrailItem, int64, error) {
	rec, err := r.store.Get("trail/trailitem/" + id)
	if err != nil {
		return TrailItem{}, 0, err
	}
	x, err := ParseTrailItem(rec.Fields)
	if err != nil {
		return TrailItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailItems reads every stored entity of this type, sorted.
func (r *Repo) ListTrailItems() ([]TrailItem, error) {
	recs := r.store.List("trail/trailitem/")
	out := make([]TrailItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailItems(out)
	return out, nil
}

// PutTrailDraft writes x, creating it when revision is zero.
func (r *Repo) PutTrailDraft(x TrailDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.traildraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.traildraft.written", 1)
	return rev, nil
}

// GetTrailDraft reads the entity with the given id.
func (r *Repo) GetTrailDraft(id string) (TrailDraft, int64, error) {
	rec, err := r.store.Get("trail/traildraft/" + id)
	if err != nil {
		return TrailDraft{}, 0, err
	}
	x, err := ParseTrailDraft(rec.Fields)
	if err != nil {
		return TrailDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListTrailDrafts() ([]TrailDraft, error) {
	recs := r.store.List("trail/traildraft/")
	out := make([]TrailDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailDrafts(out)
	return out, nil
}

// PutTrailRevision writes x, creating it when revision is zero.
func (r *Repo) PutTrailRevision(x TrailRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailrevision.written", 1)
	return rev, nil
}

// GetTrailRevision reads the entity with the given id.
func (r *Repo) GetTrailRevision(id string) (TrailRevision, int64, error) {
	rec, err := r.store.Get("trail/trailrevision/" + id)
	if err != nil {
		return TrailRevision{}, 0, err
	}
	x, err := ParseTrailRevision(rec.Fields)
	if err != nil {
		return TrailRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListTrailRevisions() ([]TrailRevision, error) {
	recs := r.store.List("trail/trailrevision/")
	out := make([]TrailRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailRevisions(out)
	return out, nil
}

// PutTrailSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutTrailSnapshot(x TrailSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailsnapshot.written", 1)
	return rev, nil
}

// GetTrailSnapshot reads the entity with the given id.
func (r *Repo) GetTrailSnapshot(id string) (TrailSnapshot, int64, error) {
	rec, err := r.store.Get("trail/trailsnapshot/" + id)
	if err != nil {
		return TrailSnapshot{}, 0, err
	}
	x, err := ParseTrailSnapshot(rec.Fields)
	if err != nil {
		return TrailSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListTrailSnapshots() ([]TrailSnapshot, error) {
	recs := r.store.List("trail/trailsnapshot/")
	out := make([]TrailSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailSnapshots(out)
	return out, nil
}

// PutTrailBatch writes x, creating it when revision is zero.
func (r *Repo) PutTrailBatch(x TrailBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailbatch.written", 1)
	return rev, nil
}

// GetTrailBatch reads the entity with the given id.
func (r *Repo) GetTrailBatch(id string) (TrailBatch, int64, error) {
	rec, err := r.store.Get("trail/trailbatch/" + id)
	if err != nil {
		return TrailBatch{}, 0, err
	}
	x, err := ParseTrailBatch(rec.Fields)
	if err != nil {
		return TrailBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListTrailBatchs() ([]TrailBatch, error) {
	recs := r.store.List("trail/trailbatch/")
	out := make([]TrailBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailBatchs(out)
	return out, nil
}

// PutTrailRequest writes x, creating it when revision is zero.
func (r *Repo) PutTrailRequest(x TrailRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailrequest.written", 1)
	return rev, nil
}

// GetTrailRequest reads the entity with the given id.
func (r *Repo) GetTrailRequest(id string) (TrailRequest, int64, error) {
	rec, err := r.store.Get("trail/trailrequest/" + id)
	if err != nil {
		return TrailRequest{}, 0, err
	}
	x, err := ParseTrailRequest(rec.Fields)
	if err != nil {
		return TrailRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailRequests reads every stored entity of this type, sorted.
func (r *Repo) ListTrailRequests() ([]TrailRequest, error) {
	recs := r.store.List("trail/trailrequest/")
	out := make([]TrailRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailRequests(out)
	return out, nil
}

// PutTrailSummary writes x, creating it when revision is zero.
func (r *Repo) PutTrailSummary(x TrailSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailsummary.written", 1)
	return rev, nil
}

// GetTrailSummary reads the entity with the given id.
func (r *Repo) GetTrailSummary(id string) (TrailSummary, int64, error) {
	rec, err := r.store.Get("trail/trailsummary/" + id)
	if err != nil {
		return TrailSummary{}, 0, err
	}
	x, err := ParseTrailSummary(rec.Fields)
	if err != nil {
		return TrailSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListTrailSummarys() ([]TrailSummary, error) {
	recs := r.store.List("trail/trailsummary/")
	out := make([]TrailSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailSummarys(out)
	return out, nil
}

// PutTrailPolicy writes x, creating it when revision is zero.
func (r *Repo) PutTrailPolicy(x TrailPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailpolicy.written", 1)
	return rev, nil
}

// GetTrailPolicy reads the entity with the given id.
func (r *Repo) GetTrailPolicy(id string) (TrailPolicy, int64, error) {
	rec, err := r.store.Get("trail/trailpolicy/" + id)
	if err != nil {
		return TrailPolicy{}, 0, err
	}
	x, err := ParseTrailPolicy(rec.Fields)
	if err != nil {
		return TrailPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListTrailPolicys() ([]TrailPolicy, error) {
	recs := r.store.List("trail/trailpolicy/")
	out := make([]TrailPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailPolicys(out)
	return out, nil
}

// PutTrailLink writes x, creating it when revision is zero.
func (r *Repo) PutTrailLink(x TrailLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.traillink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.traillink.written", 1)
	return rev, nil
}

// GetTrailLink reads the entity with the given id.
func (r *Repo) GetTrailLink(id string) (TrailLink, int64, error) {
	rec, err := r.store.Get("trail/traillink/" + id)
	if err != nil {
		return TrailLink{}, 0, err
	}
	x, err := ParseTrailLink(rec.Fields)
	if err != nil {
		return TrailLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailLinks reads every stored entity of this type, sorted.
func (r *Repo) ListTrailLinks() ([]TrailLink, error) {
	recs := r.store.List("trail/traillink/")
	out := make([]TrailLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailLinks(out)
	return out, nil
}

// PutTrailNote writes x, creating it when revision is zero.
func (r *Repo) PutTrailNote(x TrailNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailnote.written", 1)
	return rev, nil
}

// GetTrailNote reads the entity with the given id.
func (r *Repo) GetTrailNote(id string) (TrailNote, int64, error) {
	rec, err := r.store.Get("trail/trailnote/" + id)
	if err != nil {
		return TrailNote{}, 0, err
	}
	x, err := ParseTrailNote(rec.Fields)
	if err != nil {
		return TrailNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailNotes reads every stored entity of this type, sorted.
func (r *Repo) ListTrailNotes() ([]TrailNote, error) {
	recs := r.store.List("trail/trailnote/")
	out := make([]TrailNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailNotes(out)
	return out, nil
}

// PutTrailWindow writes x, creating it when revision is zero.
func (r *Repo) PutTrailWindow(x TrailWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailwindow.written", 1)
	return rev, nil
}

// GetTrailWindow reads the entity with the given id.
func (r *Repo) GetTrailWindow(id string) (TrailWindow, int64, error) {
	rec, err := r.store.Get("trail/trailwindow/" + id)
	if err != nil {
		return TrailWindow{}, 0, err
	}
	x, err := ParseTrailWindow(rec.Fields)
	if err != nil {
		return TrailWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailWindows reads every stored entity of this type, sorted.
func (r *Repo) ListTrailWindows() ([]TrailWindow, error) {
	recs := r.store.List("trail/trailwindow/")
	out := make([]TrailWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailWindows(out)
	return out, nil
}

// PutTrailQuota writes x, creating it when revision is zero.
func (r *Repo) PutTrailQuota(x TrailQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailquota.written", 1)
	return rev, nil
}

// GetTrailQuota reads the entity with the given id.
func (r *Repo) GetTrailQuota(id string) (TrailQuota, int64, error) {
	rec, err := r.store.Get("trail/trailquota/" + id)
	if err != nil {
		return TrailQuota{}, 0, err
	}
	x, err := ParseTrailQuota(rec.Fields)
	if err != nil {
		return TrailQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListTrailQuotas() ([]TrailQuota, error) {
	recs := r.store.List("trail/trailquota/")
	out := make([]TrailQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailQuotas(out)
	return out, nil
}

// PutTrailBucket writes x, creating it when revision is zero.
func (r *Repo) PutTrailBucket(x TrailBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailbucket.written", 1)
	return rev, nil
}

// GetTrailBucket reads the entity with the given id.
func (r *Repo) GetTrailBucket(id string) (TrailBucket, int64, error) {
	rec, err := r.store.Get("trail/trailbucket/" + id)
	if err != nil {
		return TrailBucket{}, 0, err
	}
	x, err := ParseTrailBucket(rec.Fields)
	if err != nil {
		return TrailBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListTrailBuckets() ([]TrailBucket, error) {
	recs := r.store.List("trail/trailbucket/")
	out := make([]TrailBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailBuckets(out)
	return out, nil
}

// PutTrailLane writes x, creating it when revision is zero.
func (r *Repo) PutTrailLane(x TrailLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.traillane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.traillane.written", 1)
	return rev, nil
}

// GetTrailLane reads the entity with the given id.
func (r *Repo) GetTrailLane(id string) (TrailLane, int64, error) {
	rec, err := r.store.Get("trail/traillane/" + id)
	if err != nil {
		return TrailLane{}, 0, err
	}
	x, err := ParseTrailLane(rec.Fields)
	if err != nil {
		return TrailLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailLanes reads every stored entity of this type, sorted.
func (r *Repo) ListTrailLanes() ([]TrailLane, error) {
	recs := r.store.List("trail/traillane/")
	out := make([]TrailLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailLanes(out)
	return out, nil
}

// PutTrailSlot writes x, creating it when revision is zero.
func (r *Repo) PutTrailSlot(x TrailSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailslot.written", 1)
	return rev, nil
}

// GetTrailSlot reads the entity with the given id.
func (r *Repo) GetTrailSlot(id string) (TrailSlot, int64, error) {
	rec, err := r.store.Get("trail/trailslot/" + id)
	if err != nil {
		return TrailSlot{}, 0, err
	}
	x, err := ParseTrailSlot(rec.Fields)
	if err != nil {
		return TrailSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailSlots reads every stored entity of this type, sorted.
func (r *Repo) ListTrailSlots() ([]TrailSlot, error) {
	recs := r.store.List("trail/trailslot/")
	out := make([]TrailSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailSlots(out)
	return out, nil
}

// PutTrailTag writes x, creating it when revision is zero.
func (r *Repo) PutTrailTag(x TrailTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailtag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailtag.written", 1)
	return rev, nil
}

// GetTrailTag reads the entity with the given id.
func (r *Repo) GetTrailTag(id string) (TrailTag, int64, error) {
	rec, err := r.store.Get("trail/trailtag/" + id)
	if err != nil {
		return TrailTag{}, 0, err
	}
	x, err := ParseTrailTag(rec.Fields)
	if err != nil {
		return TrailTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailTags reads every stored entity of this type, sorted.
func (r *Repo) ListTrailTags() ([]TrailTag, error) {
	recs := r.store.List("trail/trailtag/")
	out := make([]TrailTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailTags(out)
	return out, nil
}

// PutTrailRule writes x, creating it when revision is zero.
func (r *Repo) PutTrailRule(x TrailRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailrule.written", 1)
	return rev, nil
}

// GetTrailRule reads the entity with the given id.
func (r *Repo) GetTrailRule(id string) (TrailRule, int64, error) {
	rec, err := r.store.Get("trail/trailrule/" + id)
	if err != nil {
		return TrailRule{}, 0, err
	}
	x, err := ParseTrailRule(rec.Fields)
	if err != nil {
		return TrailRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailRules reads every stored entity of this type, sorted.
func (r *Repo) ListTrailRules() ([]TrailRule, error) {
	recs := r.store.List("trail/trailrule/")
	out := make([]TrailRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailRules(out)
	return out, nil
}

// PutTrailHold writes x, creating it when revision is zero.
func (r *Repo) PutTrailHold(x TrailHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailhold.written", 1)
	return rev, nil
}

// GetTrailHold reads the entity with the given id.
func (r *Repo) GetTrailHold(id string) (TrailHold, int64, error) {
	rec, err := r.store.Get("trail/trailhold/" + id)
	if err != nil {
		return TrailHold{}, 0, err
	}
	x, err := ParseTrailHold(rec.Fields)
	if err != nil {
		return TrailHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailHolds reads every stored entity of this type, sorted.
func (r *Repo) ListTrailHolds() ([]TrailHold, error) {
	recs := r.store.List("trail/trailhold/")
	out := make([]TrailHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailHolds(out)
	return out, nil
}

// PutTrailClaim writes x, creating it when revision is zero.
func (r *Repo) PutTrailClaim(x TrailClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailclaim.written", 1)
	return rev, nil
}

// GetTrailClaim reads the entity with the given id.
func (r *Repo) GetTrailClaim(id string) (TrailClaim, int64, error) {
	rec, err := r.store.Get("trail/trailclaim/" + id)
	if err != nil {
		return TrailClaim{}, 0, err
	}
	x, err := ParseTrailClaim(rec.Fields)
	if err != nil {
		return TrailClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailClaims reads every stored entity of this type, sorted.
func (r *Repo) ListTrailClaims() ([]TrailClaim, error) {
	recs := r.store.List("trail/trailclaim/")
	out := make([]TrailClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailClaims(out)
	return out, nil
}

// PutTrailCursor writes x, creating it when revision is zero.
func (r *Repo) PutTrailCursor(x TrailCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("trail.trailcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("trail write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("trail.trailcursor.written", 1)
	return rev, nil
}

// GetTrailCursor reads the entity with the given id.
func (r *Repo) GetTrailCursor(id string) (TrailCursor, int64, error) {
	rec, err := r.store.Get("trail/trailcursor/" + id)
	if err != nil {
		return TrailCursor{}, 0, err
	}
	x, err := ParseTrailCursor(rec.Fields)
	if err != nil {
		return TrailCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTrailCursors reads every stored entity of this type, sorted.
func (r *Repo) ListTrailCursors() ([]TrailCursor, error) {
	recs := r.store.List("trail/trailcursor/")
	out := make([]TrailCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTrailCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("trail: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTrailCursors(out)
	return out, nil
}
