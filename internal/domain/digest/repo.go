package digest

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists digest entities into a store.Store.
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
	l.Debug("digest repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many digest records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutDigestRecord writes x, creating it when revision is zero.
func (r *Repo) PutDigestRecord(x DigestRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestrecord.written", 1)
	return rev, nil
}

// GetDigestRecord reads the entity with the given id.
func (r *Repo) GetDigestRecord(id string) (DigestRecord, int64, error) {
	rec, err := r.store.Get("digest/digestrecord/" + id)
	if err != nil {
		return DigestRecord{}, 0, err
	}
	x, err := ParseDigestRecord(rec.Fields)
	if err != nil {
		return DigestRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestRecords reads every stored entity of this type, sorted.
func (r *Repo) ListDigestRecords() ([]DigestRecord, error) {
	recs := r.store.List("digest/digestrecord/")
	out := make([]DigestRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestRecords(out)
	return out, nil
}

// PutDigestEntry writes x, creating it when revision is zero.
func (r *Repo) PutDigestEntry(x DigestEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestentry.written", 1)
	return rev, nil
}

// GetDigestEntry reads the entity with the given id.
func (r *Repo) GetDigestEntry(id string) (DigestEntry, int64, error) {
	rec, err := r.store.Get("digest/digestentry/" + id)
	if err != nil {
		return DigestEntry{}, 0, err
	}
	x, err := ParseDigestEntry(rec.Fields)
	if err != nil {
		return DigestEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListDigestEntrys() ([]DigestEntry, error) {
	recs := r.store.List("digest/digestentry/")
	out := make([]DigestEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestEntrys(out)
	return out, nil
}

// PutDigestItem writes x, creating it when revision is zero.
func (r *Repo) PutDigestItem(x DigestItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestitem.written", 1)
	return rev, nil
}

// GetDigestItem reads the entity with the given id.
func (r *Repo) GetDigestItem(id string) (DigestItem, int64, error) {
	rec, err := r.store.Get("digest/digestitem/" + id)
	if err != nil {
		return DigestItem{}, 0, err
	}
	x, err := ParseDigestItem(rec.Fields)
	if err != nil {
		return DigestItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestItems reads every stored entity of this type, sorted.
func (r *Repo) ListDigestItems() ([]DigestItem, error) {
	recs := r.store.List("digest/digestitem/")
	out := make([]DigestItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestItems(out)
	return out, nil
}

// PutDigestDraft writes x, creating it when revision is zero.
func (r *Repo) PutDigestDraft(x DigestDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestdraft.written", 1)
	return rev, nil
}

// GetDigestDraft reads the entity with the given id.
func (r *Repo) GetDigestDraft(id string) (DigestDraft, int64, error) {
	rec, err := r.store.Get("digest/digestdraft/" + id)
	if err != nil {
		return DigestDraft{}, 0, err
	}
	x, err := ParseDigestDraft(rec.Fields)
	if err != nil {
		return DigestDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListDigestDrafts() ([]DigestDraft, error) {
	recs := r.store.List("digest/digestdraft/")
	out := make([]DigestDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestDrafts(out)
	return out, nil
}

// PutDigestRevision writes x, creating it when revision is zero.
func (r *Repo) PutDigestRevision(x DigestRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestrevision.written", 1)
	return rev, nil
}

// GetDigestRevision reads the entity with the given id.
func (r *Repo) GetDigestRevision(id string) (DigestRevision, int64, error) {
	rec, err := r.store.Get("digest/digestrevision/" + id)
	if err != nil {
		return DigestRevision{}, 0, err
	}
	x, err := ParseDigestRevision(rec.Fields)
	if err != nil {
		return DigestRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListDigestRevisions() ([]DigestRevision, error) {
	recs := r.store.List("digest/digestrevision/")
	out := make([]DigestRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestRevisions(out)
	return out, nil
}

// PutDigestSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutDigestSnapshot(x DigestSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestsnapshot.written", 1)
	return rev, nil
}

// GetDigestSnapshot reads the entity with the given id.
func (r *Repo) GetDigestSnapshot(id string) (DigestSnapshot, int64, error) {
	rec, err := r.store.Get("digest/digestsnapshot/" + id)
	if err != nil {
		return DigestSnapshot{}, 0, err
	}
	x, err := ParseDigestSnapshot(rec.Fields)
	if err != nil {
		return DigestSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListDigestSnapshots() ([]DigestSnapshot, error) {
	recs := r.store.List("digest/digestsnapshot/")
	out := make([]DigestSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestSnapshots(out)
	return out, nil
}

// PutDigestBatch writes x, creating it when revision is zero.
func (r *Repo) PutDigestBatch(x DigestBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestbatch.written", 1)
	return rev, nil
}

// GetDigestBatch reads the entity with the given id.
func (r *Repo) GetDigestBatch(id string) (DigestBatch, int64, error) {
	rec, err := r.store.Get("digest/digestbatch/" + id)
	if err != nil {
		return DigestBatch{}, 0, err
	}
	x, err := ParseDigestBatch(rec.Fields)
	if err != nil {
		return DigestBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListDigestBatchs() ([]DigestBatch, error) {
	recs := r.store.List("digest/digestbatch/")
	out := make([]DigestBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestBatchs(out)
	return out, nil
}

// PutDigestRequest writes x, creating it when revision is zero.
func (r *Repo) PutDigestRequest(x DigestRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestrequest.written", 1)
	return rev, nil
}

// GetDigestRequest reads the entity with the given id.
func (r *Repo) GetDigestRequest(id string) (DigestRequest, int64, error) {
	rec, err := r.store.Get("digest/digestrequest/" + id)
	if err != nil {
		return DigestRequest{}, 0, err
	}
	x, err := ParseDigestRequest(rec.Fields)
	if err != nil {
		return DigestRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestRequests reads every stored entity of this type, sorted.
func (r *Repo) ListDigestRequests() ([]DigestRequest, error) {
	recs := r.store.List("digest/digestrequest/")
	out := make([]DigestRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestRequests(out)
	return out, nil
}

// PutDigestSummary writes x, creating it when revision is zero.
func (r *Repo) PutDigestSummary(x DigestSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestsummary.written", 1)
	return rev, nil
}

// GetDigestSummary reads the entity with the given id.
func (r *Repo) GetDigestSummary(id string) (DigestSummary, int64, error) {
	rec, err := r.store.Get("digest/digestsummary/" + id)
	if err != nil {
		return DigestSummary{}, 0, err
	}
	x, err := ParseDigestSummary(rec.Fields)
	if err != nil {
		return DigestSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListDigestSummarys() ([]DigestSummary, error) {
	recs := r.store.List("digest/digestsummary/")
	out := make([]DigestSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestSummarys(out)
	return out, nil
}

// PutDigestPolicy writes x, creating it when revision is zero.
func (r *Repo) PutDigestPolicy(x DigestPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestpolicy.written", 1)
	return rev, nil
}

// GetDigestPolicy reads the entity with the given id.
func (r *Repo) GetDigestPolicy(id string) (DigestPolicy, int64, error) {
	rec, err := r.store.Get("digest/digestpolicy/" + id)
	if err != nil {
		return DigestPolicy{}, 0, err
	}
	x, err := ParseDigestPolicy(rec.Fields)
	if err != nil {
		return DigestPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListDigestPolicys() ([]DigestPolicy, error) {
	recs := r.store.List("digest/digestpolicy/")
	out := make([]DigestPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestPolicys(out)
	return out, nil
}

// PutDigestLink writes x, creating it when revision is zero.
func (r *Repo) PutDigestLink(x DigestLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestlink.written", 1)
	return rev, nil
}

// GetDigestLink reads the entity with the given id.
func (r *Repo) GetDigestLink(id string) (DigestLink, int64, error) {
	rec, err := r.store.Get("digest/digestlink/" + id)
	if err != nil {
		return DigestLink{}, 0, err
	}
	x, err := ParseDigestLink(rec.Fields)
	if err != nil {
		return DigestLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestLinks reads every stored entity of this type, sorted.
func (r *Repo) ListDigestLinks() ([]DigestLink, error) {
	recs := r.store.List("digest/digestlink/")
	out := make([]DigestLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestLinks(out)
	return out, nil
}

// PutDigestNote writes x, creating it when revision is zero.
func (r *Repo) PutDigestNote(x DigestNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestnote.written", 1)
	return rev, nil
}

// GetDigestNote reads the entity with the given id.
func (r *Repo) GetDigestNote(id string) (DigestNote, int64, error) {
	rec, err := r.store.Get("digest/digestnote/" + id)
	if err != nil {
		return DigestNote{}, 0, err
	}
	x, err := ParseDigestNote(rec.Fields)
	if err != nil {
		return DigestNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestNotes reads every stored entity of this type, sorted.
func (r *Repo) ListDigestNotes() ([]DigestNote, error) {
	recs := r.store.List("digest/digestnote/")
	out := make([]DigestNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestNotes(out)
	return out, nil
}

// PutDigestWindow writes x, creating it when revision is zero.
func (r *Repo) PutDigestWindow(x DigestWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestwindow.written", 1)
	return rev, nil
}

// GetDigestWindow reads the entity with the given id.
func (r *Repo) GetDigestWindow(id string) (DigestWindow, int64, error) {
	rec, err := r.store.Get("digest/digestwindow/" + id)
	if err != nil {
		return DigestWindow{}, 0, err
	}
	x, err := ParseDigestWindow(rec.Fields)
	if err != nil {
		return DigestWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestWindows reads every stored entity of this type, sorted.
func (r *Repo) ListDigestWindows() ([]DigestWindow, error) {
	recs := r.store.List("digest/digestwindow/")
	out := make([]DigestWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestWindows(out)
	return out, nil
}

// PutDigestQuota writes x, creating it when revision is zero.
func (r *Repo) PutDigestQuota(x DigestQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestquota.written", 1)
	return rev, nil
}

// GetDigestQuota reads the entity with the given id.
func (r *Repo) GetDigestQuota(id string) (DigestQuota, int64, error) {
	rec, err := r.store.Get("digest/digestquota/" + id)
	if err != nil {
		return DigestQuota{}, 0, err
	}
	x, err := ParseDigestQuota(rec.Fields)
	if err != nil {
		return DigestQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListDigestQuotas() ([]DigestQuota, error) {
	recs := r.store.List("digest/digestquota/")
	out := make([]DigestQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestQuotas(out)
	return out, nil
}

// PutDigestBucket writes x, creating it when revision is zero.
func (r *Repo) PutDigestBucket(x DigestBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestbucket.written", 1)
	return rev, nil
}

// GetDigestBucket reads the entity with the given id.
func (r *Repo) GetDigestBucket(id string) (DigestBucket, int64, error) {
	rec, err := r.store.Get("digest/digestbucket/" + id)
	if err != nil {
		return DigestBucket{}, 0, err
	}
	x, err := ParseDigestBucket(rec.Fields)
	if err != nil {
		return DigestBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListDigestBuckets() ([]DigestBucket, error) {
	recs := r.store.List("digest/digestbucket/")
	out := make([]DigestBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestBuckets(out)
	return out, nil
}

// PutDigestLane writes x, creating it when revision is zero.
func (r *Repo) PutDigestLane(x DigestLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestlane.written", 1)
	return rev, nil
}

// GetDigestLane reads the entity with the given id.
func (r *Repo) GetDigestLane(id string) (DigestLane, int64, error) {
	rec, err := r.store.Get("digest/digestlane/" + id)
	if err != nil {
		return DigestLane{}, 0, err
	}
	x, err := ParseDigestLane(rec.Fields)
	if err != nil {
		return DigestLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestLanes reads every stored entity of this type, sorted.
func (r *Repo) ListDigestLanes() ([]DigestLane, error) {
	recs := r.store.List("digest/digestlane/")
	out := make([]DigestLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestLanes(out)
	return out, nil
}

// PutDigestSlot writes x, creating it when revision is zero.
func (r *Repo) PutDigestSlot(x DigestSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestslot.written", 1)
	return rev, nil
}

// GetDigestSlot reads the entity with the given id.
func (r *Repo) GetDigestSlot(id string) (DigestSlot, int64, error) {
	rec, err := r.store.Get("digest/digestslot/" + id)
	if err != nil {
		return DigestSlot{}, 0, err
	}
	x, err := ParseDigestSlot(rec.Fields)
	if err != nil {
		return DigestSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestSlots reads every stored entity of this type, sorted.
func (r *Repo) ListDigestSlots() ([]DigestSlot, error) {
	recs := r.store.List("digest/digestslot/")
	out := make([]DigestSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestSlots(out)
	return out, nil
}

// PutDigestTag writes x, creating it when revision is zero.
func (r *Repo) PutDigestTag(x DigestTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digesttag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digesttag.written", 1)
	return rev, nil
}

// GetDigestTag reads the entity with the given id.
func (r *Repo) GetDigestTag(id string) (DigestTag, int64, error) {
	rec, err := r.store.Get("digest/digesttag/" + id)
	if err != nil {
		return DigestTag{}, 0, err
	}
	x, err := ParseDigestTag(rec.Fields)
	if err != nil {
		return DigestTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestTags reads every stored entity of this type, sorted.
func (r *Repo) ListDigestTags() ([]DigestTag, error) {
	recs := r.store.List("digest/digesttag/")
	out := make([]DigestTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestTags(out)
	return out, nil
}

// PutDigestRule writes x, creating it when revision is zero.
func (r *Repo) PutDigestRule(x DigestRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestrule.written", 1)
	return rev, nil
}

// GetDigestRule reads the entity with the given id.
func (r *Repo) GetDigestRule(id string) (DigestRule, int64, error) {
	rec, err := r.store.Get("digest/digestrule/" + id)
	if err != nil {
		return DigestRule{}, 0, err
	}
	x, err := ParseDigestRule(rec.Fields)
	if err != nil {
		return DigestRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestRules reads every stored entity of this type, sorted.
func (r *Repo) ListDigestRules() ([]DigestRule, error) {
	recs := r.store.List("digest/digestrule/")
	out := make([]DigestRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestRules(out)
	return out, nil
}

// PutDigestHold writes x, creating it when revision is zero.
func (r *Repo) PutDigestHold(x DigestHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digesthold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digesthold.written", 1)
	return rev, nil
}

// GetDigestHold reads the entity with the given id.
func (r *Repo) GetDigestHold(id string) (DigestHold, int64, error) {
	rec, err := r.store.Get("digest/digesthold/" + id)
	if err != nil {
		return DigestHold{}, 0, err
	}
	x, err := ParseDigestHold(rec.Fields)
	if err != nil {
		return DigestHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestHolds reads every stored entity of this type, sorted.
func (r *Repo) ListDigestHolds() ([]DigestHold, error) {
	recs := r.store.List("digest/digesthold/")
	out := make([]DigestHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestHolds(out)
	return out, nil
}

// PutDigestClaim writes x, creating it when revision is zero.
func (r *Repo) PutDigestClaim(x DigestClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestclaim.written", 1)
	return rev, nil
}

// GetDigestClaim reads the entity with the given id.
func (r *Repo) GetDigestClaim(id string) (DigestClaim, int64, error) {
	rec, err := r.store.Get("digest/digestclaim/" + id)
	if err != nil {
		return DigestClaim{}, 0, err
	}
	x, err := ParseDigestClaim(rec.Fields)
	if err != nil {
		return DigestClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestClaims reads every stored entity of this type, sorted.
func (r *Repo) ListDigestClaims() ([]DigestClaim, error) {
	recs := r.store.List("digest/digestclaim/")
	out := make([]DigestClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestClaims(out)
	return out, nil
}

// PutDigestCursor writes x, creating it when revision is zero.
func (r *Repo) PutDigestCursor(x DigestCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("digest.digestcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("digest write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("digest.digestcursor.written", 1)
	return rev, nil
}

// GetDigestCursor reads the entity with the given id.
func (r *Repo) GetDigestCursor(id string) (DigestCursor, int64, error) {
	rec, err := r.store.Get("digest/digestcursor/" + id)
	if err != nil {
		return DigestCursor{}, 0, err
	}
	x, err := ParseDigestCursor(rec.Fields)
	if err != nil {
		return DigestCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDigestCursors reads every stored entity of this type, sorted.
func (r *Repo) ListDigestCursors() ([]DigestCursor, error) {
	recs := r.store.List("digest/digestcursor/")
	out := make([]DigestCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDigestCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("digest: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDigestCursors(out)
	return out, nil
}
