package dunning

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists dunning entities into a store.Store.
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
	l.Debug("dunning repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many dunning records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutDunningRecord writes x, creating it when revision is zero.
func (r *Repo) PutDunningRecord(x DunningRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningrecord.written", 1)
	return rev, nil
}

// GetDunningRecord reads the entity with the given id.
func (r *Repo) GetDunningRecord(id string) (DunningRecord, int64, error) {
	rec, err := r.store.Get("dunning/dunningrecord/" + id)
	if err != nil {
		return DunningRecord{}, 0, err
	}
	x, err := ParseDunningRecord(rec.Fields)
	if err != nil {
		return DunningRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningRecords reads every stored entity of this type, sorted.
func (r *Repo) ListDunningRecords() ([]DunningRecord, error) {
	recs := r.store.List("dunning/dunningrecord/")
	out := make([]DunningRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningRecords(out)
	return out, nil
}

// PutDunningEntry writes x, creating it when revision is zero.
func (r *Repo) PutDunningEntry(x DunningEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningentry.written", 1)
	return rev, nil
}

// GetDunningEntry reads the entity with the given id.
func (r *Repo) GetDunningEntry(id string) (DunningEntry, int64, error) {
	rec, err := r.store.Get("dunning/dunningentry/" + id)
	if err != nil {
		return DunningEntry{}, 0, err
	}
	x, err := ParseDunningEntry(rec.Fields)
	if err != nil {
		return DunningEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListDunningEntrys() ([]DunningEntry, error) {
	recs := r.store.List("dunning/dunningentry/")
	out := make([]DunningEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningEntrys(out)
	return out, nil
}

// PutDunningItem writes x, creating it when revision is zero.
func (r *Repo) PutDunningItem(x DunningItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningitem.written", 1)
	return rev, nil
}

// GetDunningItem reads the entity with the given id.
func (r *Repo) GetDunningItem(id string) (DunningItem, int64, error) {
	rec, err := r.store.Get("dunning/dunningitem/" + id)
	if err != nil {
		return DunningItem{}, 0, err
	}
	x, err := ParseDunningItem(rec.Fields)
	if err != nil {
		return DunningItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningItems reads every stored entity of this type, sorted.
func (r *Repo) ListDunningItems() ([]DunningItem, error) {
	recs := r.store.List("dunning/dunningitem/")
	out := make([]DunningItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningItems(out)
	return out, nil
}

// PutDunningDraft writes x, creating it when revision is zero.
func (r *Repo) PutDunningDraft(x DunningDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningdraft.written", 1)
	return rev, nil
}

// GetDunningDraft reads the entity with the given id.
func (r *Repo) GetDunningDraft(id string) (DunningDraft, int64, error) {
	rec, err := r.store.Get("dunning/dunningdraft/" + id)
	if err != nil {
		return DunningDraft{}, 0, err
	}
	x, err := ParseDunningDraft(rec.Fields)
	if err != nil {
		return DunningDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListDunningDrafts() ([]DunningDraft, error) {
	recs := r.store.List("dunning/dunningdraft/")
	out := make([]DunningDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningDrafts(out)
	return out, nil
}

// PutDunningRevision writes x, creating it when revision is zero.
func (r *Repo) PutDunningRevision(x DunningRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningrevision.written", 1)
	return rev, nil
}

// GetDunningRevision reads the entity with the given id.
func (r *Repo) GetDunningRevision(id string) (DunningRevision, int64, error) {
	rec, err := r.store.Get("dunning/dunningrevision/" + id)
	if err != nil {
		return DunningRevision{}, 0, err
	}
	x, err := ParseDunningRevision(rec.Fields)
	if err != nil {
		return DunningRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListDunningRevisions() ([]DunningRevision, error) {
	recs := r.store.List("dunning/dunningrevision/")
	out := make([]DunningRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningRevisions(out)
	return out, nil
}

// PutDunningSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutDunningSnapshot(x DunningSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningsnapshot.written", 1)
	return rev, nil
}

// GetDunningSnapshot reads the entity with the given id.
func (r *Repo) GetDunningSnapshot(id string) (DunningSnapshot, int64, error) {
	rec, err := r.store.Get("dunning/dunningsnapshot/" + id)
	if err != nil {
		return DunningSnapshot{}, 0, err
	}
	x, err := ParseDunningSnapshot(rec.Fields)
	if err != nil {
		return DunningSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListDunningSnapshots() ([]DunningSnapshot, error) {
	recs := r.store.List("dunning/dunningsnapshot/")
	out := make([]DunningSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningSnapshots(out)
	return out, nil
}

// PutDunningBatch writes x, creating it when revision is zero.
func (r *Repo) PutDunningBatch(x DunningBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningbatch.written", 1)
	return rev, nil
}

// GetDunningBatch reads the entity with the given id.
func (r *Repo) GetDunningBatch(id string) (DunningBatch, int64, error) {
	rec, err := r.store.Get("dunning/dunningbatch/" + id)
	if err != nil {
		return DunningBatch{}, 0, err
	}
	x, err := ParseDunningBatch(rec.Fields)
	if err != nil {
		return DunningBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListDunningBatchs() ([]DunningBatch, error) {
	recs := r.store.List("dunning/dunningbatch/")
	out := make([]DunningBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningBatchs(out)
	return out, nil
}

// PutDunningRequest writes x, creating it when revision is zero.
func (r *Repo) PutDunningRequest(x DunningRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningrequest.written", 1)
	return rev, nil
}

// GetDunningRequest reads the entity with the given id.
func (r *Repo) GetDunningRequest(id string) (DunningRequest, int64, error) {
	rec, err := r.store.Get("dunning/dunningrequest/" + id)
	if err != nil {
		return DunningRequest{}, 0, err
	}
	x, err := ParseDunningRequest(rec.Fields)
	if err != nil {
		return DunningRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningRequests reads every stored entity of this type, sorted.
func (r *Repo) ListDunningRequests() ([]DunningRequest, error) {
	recs := r.store.List("dunning/dunningrequest/")
	out := make([]DunningRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningRequests(out)
	return out, nil
}

// PutDunningSummary writes x, creating it when revision is zero.
func (r *Repo) PutDunningSummary(x DunningSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningsummary.written", 1)
	return rev, nil
}

// GetDunningSummary reads the entity with the given id.
func (r *Repo) GetDunningSummary(id string) (DunningSummary, int64, error) {
	rec, err := r.store.Get("dunning/dunningsummary/" + id)
	if err != nil {
		return DunningSummary{}, 0, err
	}
	x, err := ParseDunningSummary(rec.Fields)
	if err != nil {
		return DunningSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListDunningSummarys() ([]DunningSummary, error) {
	recs := r.store.List("dunning/dunningsummary/")
	out := make([]DunningSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningSummarys(out)
	return out, nil
}

// PutDunningPolicy writes x, creating it when revision is zero.
func (r *Repo) PutDunningPolicy(x DunningPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningpolicy.written", 1)
	return rev, nil
}

// GetDunningPolicy reads the entity with the given id.
func (r *Repo) GetDunningPolicy(id string) (DunningPolicy, int64, error) {
	rec, err := r.store.Get("dunning/dunningpolicy/" + id)
	if err != nil {
		return DunningPolicy{}, 0, err
	}
	x, err := ParseDunningPolicy(rec.Fields)
	if err != nil {
		return DunningPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListDunningPolicys() ([]DunningPolicy, error) {
	recs := r.store.List("dunning/dunningpolicy/")
	out := make([]DunningPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningPolicys(out)
	return out, nil
}

// PutDunningLink writes x, creating it when revision is zero.
func (r *Repo) PutDunningLink(x DunningLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunninglink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunninglink.written", 1)
	return rev, nil
}

// GetDunningLink reads the entity with the given id.
func (r *Repo) GetDunningLink(id string) (DunningLink, int64, error) {
	rec, err := r.store.Get("dunning/dunninglink/" + id)
	if err != nil {
		return DunningLink{}, 0, err
	}
	x, err := ParseDunningLink(rec.Fields)
	if err != nil {
		return DunningLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningLinks reads every stored entity of this type, sorted.
func (r *Repo) ListDunningLinks() ([]DunningLink, error) {
	recs := r.store.List("dunning/dunninglink/")
	out := make([]DunningLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningLinks(out)
	return out, nil
}

// PutDunningNote writes x, creating it when revision is zero.
func (r *Repo) PutDunningNote(x DunningNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningnote.written", 1)
	return rev, nil
}

// GetDunningNote reads the entity with the given id.
func (r *Repo) GetDunningNote(id string) (DunningNote, int64, error) {
	rec, err := r.store.Get("dunning/dunningnote/" + id)
	if err != nil {
		return DunningNote{}, 0, err
	}
	x, err := ParseDunningNote(rec.Fields)
	if err != nil {
		return DunningNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningNotes reads every stored entity of this type, sorted.
func (r *Repo) ListDunningNotes() ([]DunningNote, error) {
	recs := r.store.List("dunning/dunningnote/")
	out := make([]DunningNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningNotes(out)
	return out, nil
}

// PutDunningWindow writes x, creating it when revision is zero.
func (r *Repo) PutDunningWindow(x DunningWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningwindow.written", 1)
	return rev, nil
}

// GetDunningWindow reads the entity with the given id.
func (r *Repo) GetDunningWindow(id string) (DunningWindow, int64, error) {
	rec, err := r.store.Get("dunning/dunningwindow/" + id)
	if err != nil {
		return DunningWindow{}, 0, err
	}
	x, err := ParseDunningWindow(rec.Fields)
	if err != nil {
		return DunningWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningWindows reads every stored entity of this type, sorted.
func (r *Repo) ListDunningWindows() ([]DunningWindow, error) {
	recs := r.store.List("dunning/dunningwindow/")
	out := make([]DunningWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningWindows(out)
	return out, nil
}

// PutDunningQuota writes x, creating it when revision is zero.
func (r *Repo) PutDunningQuota(x DunningQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningquota.written", 1)
	return rev, nil
}

// GetDunningQuota reads the entity with the given id.
func (r *Repo) GetDunningQuota(id string) (DunningQuota, int64, error) {
	rec, err := r.store.Get("dunning/dunningquota/" + id)
	if err != nil {
		return DunningQuota{}, 0, err
	}
	x, err := ParseDunningQuota(rec.Fields)
	if err != nil {
		return DunningQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListDunningQuotas() ([]DunningQuota, error) {
	recs := r.store.List("dunning/dunningquota/")
	out := make([]DunningQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningQuotas(out)
	return out, nil
}

// PutDunningBucket writes x, creating it when revision is zero.
func (r *Repo) PutDunningBucket(x DunningBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningbucket.written", 1)
	return rev, nil
}

// GetDunningBucket reads the entity with the given id.
func (r *Repo) GetDunningBucket(id string) (DunningBucket, int64, error) {
	rec, err := r.store.Get("dunning/dunningbucket/" + id)
	if err != nil {
		return DunningBucket{}, 0, err
	}
	x, err := ParseDunningBucket(rec.Fields)
	if err != nil {
		return DunningBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListDunningBuckets() ([]DunningBucket, error) {
	recs := r.store.List("dunning/dunningbucket/")
	out := make([]DunningBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningBuckets(out)
	return out, nil
}

// PutDunningLane writes x, creating it when revision is zero.
func (r *Repo) PutDunningLane(x DunningLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunninglane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunninglane.written", 1)
	return rev, nil
}

// GetDunningLane reads the entity with the given id.
func (r *Repo) GetDunningLane(id string) (DunningLane, int64, error) {
	rec, err := r.store.Get("dunning/dunninglane/" + id)
	if err != nil {
		return DunningLane{}, 0, err
	}
	x, err := ParseDunningLane(rec.Fields)
	if err != nil {
		return DunningLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningLanes reads every stored entity of this type, sorted.
func (r *Repo) ListDunningLanes() ([]DunningLane, error) {
	recs := r.store.List("dunning/dunninglane/")
	out := make([]DunningLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningLanes(out)
	return out, nil
}

// PutDunningSlot writes x, creating it when revision is zero.
func (r *Repo) PutDunningSlot(x DunningSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningslot.written", 1)
	return rev, nil
}

// GetDunningSlot reads the entity with the given id.
func (r *Repo) GetDunningSlot(id string) (DunningSlot, int64, error) {
	rec, err := r.store.Get("dunning/dunningslot/" + id)
	if err != nil {
		return DunningSlot{}, 0, err
	}
	x, err := ParseDunningSlot(rec.Fields)
	if err != nil {
		return DunningSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningSlots reads every stored entity of this type, sorted.
func (r *Repo) ListDunningSlots() ([]DunningSlot, error) {
	recs := r.store.List("dunning/dunningslot/")
	out := make([]DunningSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningSlots(out)
	return out, nil
}

// PutDunningTag writes x, creating it when revision is zero.
func (r *Repo) PutDunningTag(x DunningTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningtag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningtag.written", 1)
	return rev, nil
}

// GetDunningTag reads the entity with the given id.
func (r *Repo) GetDunningTag(id string) (DunningTag, int64, error) {
	rec, err := r.store.Get("dunning/dunningtag/" + id)
	if err != nil {
		return DunningTag{}, 0, err
	}
	x, err := ParseDunningTag(rec.Fields)
	if err != nil {
		return DunningTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningTags reads every stored entity of this type, sorted.
func (r *Repo) ListDunningTags() ([]DunningTag, error) {
	recs := r.store.List("dunning/dunningtag/")
	out := make([]DunningTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningTags(out)
	return out, nil
}

// PutDunningRule writes x, creating it when revision is zero.
func (r *Repo) PutDunningRule(x DunningRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningrule.written", 1)
	return rev, nil
}

// GetDunningRule reads the entity with the given id.
func (r *Repo) GetDunningRule(id string) (DunningRule, int64, error) {
	rec, err := r.store.Get("dunning/dunningrule/" + id)
	if err != nil {
		return DunningRule{}, 0, err
	}
	x, err := ParseDunningRule(rec.Fields)
	if err != nil {
		return DunningRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningRules reads every stored entity of this type, sorted.
func (r *Repo) ListDunningRules() ([]DunningRule, error) {
	recs := r.store.List("dunning/dunningrule/")
	out := make([]DunningRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningRules(out)
	return out, nil
}

// PutDunningHold writes x, creating it when revision is zero.
func (r *Repo) PutDunningHold(x DunningHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunninghold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunninghold.written", 1)
	return rev, nil
}

// GetDunningHold reads the entity with the given id.
func (r *Repo) GetDunningHold(id string) (DunningHold, int64, error) {
	rec, err := r.store.Get("dunning/dunninghold/" + id)
	if err != nil {
		return DunningHold{}, 0, err
	}
	x, err := ParseDunningHold(rec.Fields)
	if err != nil {
		return DunningHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningHolds reads every stored entity of this type, sorted.
func (r *Repo) ListDunningHolds() ([]DunningHold, error) {
	recs := r.store.List("dunning/dunninghold/")
	out := make([]DunningHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningHolds(out)
	return out, nil
}

// PutDunningClaim writes x, creating it when revision is zero.
func (r *Repo) PutDunningClaim(x DunningClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningclaim.written", 1)
	return rev, nil
}

// GetDunningClaim reads the entity with the given id.
func (r *Repo) GetDunningClaim(id string) (DunningClaim, int64, error) {
	rec, err := r.store.Get("dunning/dunningclaim/" + id)
	if err != nil {
		return DunningClaim{}, 0, err
	}
	x, err := ParseDunningClaim(rec.Fields)
	if err != nil {
		return DunningClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningClaims reads every stored entity of this type, sorted.
func (r *Repo) ListDunningClaims() ([]DunningClaim, error) {
	recs := r.store.List("dunning/dunningclaim/")
	out := make([]DunningClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningClaims(out)
	return out, nil
}

// PutDunningCursor writes x, creating it when revision is zero.
func (r *Repo) PutDunningCursor(x DunningCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("dunning.dunningcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("dunning write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("dunning.dunningcursor.written", 1)
	return rev, nil
}

// GetDunningCursor reads the entity with the given id.
func (r *Repo) GetDunningCursor(id string) (DunningCursor, int64, error) {
	rec, err := r.store.Get("dunning/dunningcursor/" + id)
	if err != nil {
		return DunningCursor{}, 0, err
	}
	x, err := ParseDunningCursor(rec.Fields)
	if err != nil {
		return DunningCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDunningCursors reads every stored entity of this type, sorted.
func (r *Repo) ListDunningCursors() ([]DunningCursor, error) {
	recs := r.store.List("dunning/dunningcursor/")
	out := make([]DunningCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDunningCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("dunning: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDunningCursors(out)
	return out, nil
}
