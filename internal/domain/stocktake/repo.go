package stocktake

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists stocktake entities into a store.Store.
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
	l.Debug("stocktake repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many stocktake records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutStocktakeRecord writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeRecord(x StocktakeRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakerecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakerecord.written", 1)
	return rev, nil
}

// GetStocktakeRecord reads the entity with the given id.
func (r *Repo) GetStocktakeRecord(id string) (StocktakeRecord, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakerecord/" + id)
	if err != nil {
		return StocktakeRecord{}, 0, err
	}
	x, err := ParseStocktakeRecord(rec.Fields)
	if err != nil {
		return StocktakeRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeRecords reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeRecords() ([]StocktakeRecord, error) {
	recs := r.store.List("stocktake/stocktakerecord/")
	out := make([]StocktakeRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeRecords(out)
	return out, nil
}

// PutStocktakeEntry writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeEntry(x StocktakeEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakeentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakeentry.written", 1)
	return rev, nil
}

// GetStocktakeEntry reads the entity with the given id.
func (r *Repo) GetStocktakeEntry(id string) (StocktakeEntry, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakeentry/" + id)
	if err != nil {
		return StocktakeEntry{}, 0, err
	}
	x, err := ParseStocktakeEntry(rec.Fields)
	if err != nil {
		return StocktakeEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeEntrys() ([]StocktakeEntry, error) {
	recs := r.store.List("stocktake/stocktakeentry/")
	out := make([]StocktakeEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeEntrys(out)
	return out, nil
}

// PutStocktakeItem writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeItem(x StocktakeItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakeitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakeitem.written", 1)
	return rev, nil
}

// GetStocktakeItem reads the entity with the given id.
func (r *Repo) GetStocktakeItem(id string) (StocktakeItem, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakeitem/" + id)
	if err != nil {
		return StocktakeItem{}, 0, err
	}
	x, err := ParseStocktakeItem(rec.Fields)
	if err != nil {
		return StocktakeItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeItems reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeItems() ([]StocktakeItem, error) {
	recs := r.store.List("stocktake/stocktakeitem/")
	out := make([]StocktakeItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeItems(out)
	return out, nil
}

// PutStocktakeDraft writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeDraft(x StocktakeDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakedraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakedraft.written", 1)
	return rev, nil
}

// GetStocktakeDraft reads the entity with the given id.
func (r *Repo) GetStocktakeDraft(id string) (StocktakeDraft, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakedraft/" + id)
	if err != nil {
		return StocktakeDraft{}, 0, err
	}
	x, err := ParseStocktakeDraft(rec.Fields)
	if err != nil {
		return StocktakeDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeDrafts() ([]StocktakeDraft, error) {
	recs := r.store.List("stocktake/stocktakedraft/")
	out := make([]StocktakeDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeDrafts(out)
	return out, nil
}

// PutStocktakeRevision writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeRevision(x StocktakeRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakerevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakerevision.written", 1)
	return rev, nil
}

// GetStocktakeRevision reads the entity with the given id.
func (r *Repo) GetStocktakeRevision(id string) (StocktakeRevision, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakerevision/" + id)
	if err != nil {
		return StocktakeRevision{}, 0, err
	}
	x, err := ParseStocktakeRevision(rec.Fields)
	if err != nil {
		return StocktakeRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeRevisions() ([]StocktakeRevision, error) {
	recs := r.store.List("stocktake/stocktakerevision/")
	out := make([]StocktakeRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeRevisions(out)
	return out, nil
}

// PutStocktakeSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeSnapshot(x StocktakeSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakesnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakesnapshot.written", 1)
	return rev, nil
}

// GetStocktakeSnapshot reads the entity with the given id.
func (r *Repo) GetStocktakeSnapshot(id string) (StocktakeSnapshot, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakesnapshot/" + id)
	if err != nil {
		return StocktakeSnapshot{}, 0, err
	}
	x, err := ParseStocktakeSnapshot(rec.Fields)
	if err != nil {
		return StocktakeSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeSnapshots() ([]StocktakeSnapshot, error) {
	recs := r.store.List("stocktake/stocktakesnapshot/")
	out := make([]StocktakeSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeSnapshots(out)
	return out, nil
}

// PutStocktakeBatch writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeBatch(x StocktakeBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakebatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakebatch.written", 1)
	return rev, nil
}

// GetStocktakeBatch reads the entity with the given id.
func (r *Repo) GetStocktakeBatch(id string) (StocktakeBatch, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakebatch/" + id)
	if err != nil {
		return StocktakeBatch{}, 0, err
	}
	x, err := ParseStocktakeBatch(rec.Fields)
	if err != nil {
		return StocktakeBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeBatchs() ([]StocktakeBatch, error) {
	recs := r.store.List("stocktake/stocktakebatch/")
	out := make([]StocktakeBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeBatchs(out)
	return out, nil
}

// PutStocktakeRequest writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeRequest(x StocktakeRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakerequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakerequest.written", 1)
	return rev, nil
}

// GetStocktakeRequest reads the entity with the given id.
func (r *Repo) GetStocktakeRequest(id string) (StocktakeRequest, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakerequest/" + id)
	if err != nil {
		return StocktakeRequest{}, 0, err
	}
	x, err := ParseStocktakeRequest(rec.Fields)
	if err != nil {
		return StocktakeRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeRequests reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeRequests() ([]StocktakeRequest, error) {
	recs := r.store.List("stocktake/stocktakerequest/")
	out := make([]StocktakeRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeRequests(out)
	return out, nil
}

// PutStocktakeSummary writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeSummary(x StocktakeSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakesummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakesummary.written", 1)
	return rev, nil
}

// GetStocktakeSummary reads the entity with the given id.
func (r *Repo) GetStocktakeSummary(id string) (StocktakeSummary, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakesummary/" + id)
	if err != nil {
		return StocktakeSummary{}, 0, err
	}
	x, err := ParseStocktakeSummary(rec.Fields)
	if err != nil {
		return StocktakeSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeSummarys() ([]StocktakeSummary, error) {
	recs := r.store.List("stocktake/stocktakesummary/")
	out := make([]StocktakeSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeSummarys(out)
	return out, nil
}

// PutStocktakePolicy writes x, creating it when revision is zero.
func (r *Repo) PutStocktakePolicy(x StocktakePolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakepolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakepolicy.written", 1)
	return rev, nil
}

// GetStocktakePolicy reads the entity with the given id.
func (r *Repo) GetStocktakePolicy(id string) (StocktakePolicy, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakepolicy/" + id)
	if err != nil {
		return StocktakePolicy{}, 0, err
	}
	x, err := ParseStocktakePolicy(rec.Fields)
	if err != nil {
		return StocktakePolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakePolicys reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakePolicys() ([]StocktakePolicy, error) {
	recs := r.store.List("stocktake/stocktakepolicy/")
	out := make([]StocktakePolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakePolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakePolicys(out)
	return out, nil
}

// PutStocktakeLink writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeLink(x StocktakeLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakelink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakelink.written", 1)
	return rev, nil
}

// GetStocktakeLink reads the entity with the given id.
func (r *Repo) GetStocktakeLink(id string) (StocktakeLink, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakelink/" + id)
	if err != nil {
		return StocktakeLink{}, 0, err
	}
	x, err := ParseStocktakeLink(rec.Fields)
	if err != nil {
		return StocktakeLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeLinks reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeLinks() ([]StocktakeLink, error) {
	recs := r.store.List("stocktake/stocktakelink/")
	out := make([]StocktakeLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeLinks(out)
	return out, nil
}

// PutStocktakeNote writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeNote(x StocktakeNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakenote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakenote.written", 1)
	return rev, nil
}

// GetStocktakeNote reads the entity with the given id.
func (r *Repo) GetStocktakeNote(id string) (StocktakeNote, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakenote/" + id)
	if err != nil {
		return StocktakeNote{}, 0, err
	}
	x, err := ParseStocktakeNote(rec.Fields)
	if err != nil {
		return StocktakeNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeNotes reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeNotes() ([]StocktakeNote, error) {
	recs := r.store.List("stocktake/stocktakenote/")
	out := make([]StocktakeNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeNotes(out)
	return out, nil
}

// PutStocktakeWindow writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeWindow(x StocktakeWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakewindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakewindow.written", 1)
	return rev, nil
}

// GetStocktakeWindow reads the entity with the given id.
func (r *Repo) GetStocktakeWindow(id string) (StocktakeWindow, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakewindow/" + id)
	if err != nil {
		return StocktakeWindow{}, 0, err
	}
	x, err := ParseStocktakeWindow(rec.Fields)
	if err != nil {
		return StocktakeWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeWindows reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeWindows() ([]StocktakeWindow, error) {
	recs := r.store.List("stocktake/stocktakewindow/")
	out := make([]StocktakeWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeWindows(out)
	return out, nil
}

// PutStocktakeQuota writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeQuota(x StocktakeQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakequota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakequota.written", 1)
	return rev, nil
}

// GetStocktakeQuota reads the entity with the given id.
func (r *Repo) GetStocktakeQuota(id string) (StocktakeQuota, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakequota/" + id)
	if err != nil {
		return StocktakeQuota{}, 0, err
	}
	x, err := ParseStocktakeQuota(rec.Fields)
	if err != nil {
		return StocktakeQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeQuotas() ([]StocktakeQuota, error) {
	recs := r.store.List("stocktake/stocktakequota/")
	out := make([]StocktakeQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeQuotas(out)
	return out, nil
}

// PutStocktakeBucket writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeBucket(x StocktakeBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakebucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakebucket.written", 1)
	return rev, nil
}

// GetStocktakeBucket reads the entity with the given id.
func (r *Repo) GetStocktakeBucket(id string) (StocktakeBucket, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakebucket/" + id)
	if err != nil {
		return StocktakeBucket{}, 0, err
	}
	x, err := ParseStocktakeBucket(rec.Fields)
	if err != nil {
		return StocktakeBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeBuckets() ([]StocktakeBucket, error) {
	recs := r.store.List("stocktake/stocktakebucket/")
	out := make([]StocktakeBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeBuckets(out)
	return out, nil
}

// PutStocktakeLane writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeLane(x StocktakeLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakelane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakelane.written", 1)
	return rev, nil
}

// GetStocktakeLane reads the entity with the given id.
func (r *Repo) GetStocktakeLane(id string) (StocktakeLane, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakelane/" + id)
	if err != nil {
		return StocktakeLane{}, 0, err
	}
	x, err := ParseStocktakeLane(rec.Fields)
	if err != nil {
		return StocktakeLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeLanes reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeLanes() ([]StocktakeLane, error) {
	recs := r.store.List("stocktake/stocktakelane/")
	out := make([]StocktakeLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeLanes(out)
	return out, nil
}

// PutStocktakeSlot writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeSlot(x StocktakeSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakeslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakeslot.written", 1)
	return rev, nil
}

// GetStocktakeSlot reads the entity with the given id.
func (r *Repo) GetStocktakeSlot(id string) (StocktakeSlot, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakeslot/" + id)
	if err != nil {
		return StocktakeSlot{}, 0, err
	}
	x, err := ParseStocktakeSlot(rec.Fields)
	if err != nil {
		return StocktakeSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeSlots reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeSlots() ([]StocktakeSlot, error) {
	recs := r.store.List("stocktake/stocktakeslot/")
	out := make([]StocktakeSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeSlots(out)
	return out, nil
}

// PutStocktakeTag writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeTag(x StocktakeTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktaketag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktaketag.written", 1)
	return rev, nil
}

// GetStocktakeTag reads the entity with the given id.
func (r *Repo) GetStocktakeTag(id string) (StocktakeTag, int64, error) {
	rec, err := r.store.Get("stocktake/stocktaketag/" + id)
	if err != nil {
		return StocktakeTag{}, 0, err
	}
	x, err := ParseStocktakeTag(rec.Fields)
	if err != nil {
		return StocktakeTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeTags reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeTags() ([]StocktakeTag, error) {
	recs := r.store.List("stocktake/stocktaketag/")
	out := make([]StocktakeTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeTags(out)
	return out, nil
}

// PutStocktakeRule writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeRule(x StocktakeRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakerule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakerule.written", 1)
	return rev, nil
}

// GetStocktakeRule reads the entity with the given id.
func (r *Repo) GetStocktakeRule(id string) (StocktakeRule, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakerule/" + id)
	if err != nil {
		return StocktakeRule{}, 0, err
	}
	x, err := ParseStocktakeRule(rec.Fields)
	if err != nil {
		return StocktakeRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeRules reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeRules() ([]StocktakeRule, error) {
	recs := r.store.List("stocktake/stocktakerule/")
	out := make([]StocktakeRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeRules(out)
	return out, nil
}

// PutStocktakeHold writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeHold(x StocktakeHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakehold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakehold.written", 1)
	return rev, nil
}

// GetStocktakeHold reads the entity with the given id.
func (r *Repo) GetStocktakeHold(id string) (StocktakeHold, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakehold/" + id)
	if err != nil {
		return StocktakeHold{}, 0, err
	}
	x, err := ParseStocktakeHold(rec.Fields)
	if err != nil {
		return StocktakeHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeHolds reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeHolds() ([]StocktakeHold, error) {
	recs := r.store.List("stocktake/stocktakehold/")
	out := make([]StocktakeHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeHolds(out)
	return out, nil
}

// PutStocktakeClaim writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeClaim(x StocktakeClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakeclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakeclaim.written", 1)
	return rev, nil
}

// GetStocktakeClaim reads the entity with the given id.
func (r *Repo) GetStocktakeClaim(id string) (StocktakeClaim, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakeclaim/" + id)
	if err != nil {
		return StocktakeClaim{}, 0, err
	}
	x, err := ParseStocktakeClaim(rec.Fields)
	if err != nil {
		return StocktakeClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeClaims reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeClaims() ([]StocktakeClaim, error) {
	recs := r.store.List("stocktake/stocktakeclaim/")
	out := make([]StocktakeClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeClaims(out)
	return out, nil
}

// PutStocktakeCursor writes x, creating it when revision is zero.
func (r *Repo) PutStocktakeCursor(x StocktakeCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("stocktake.stocktakecursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("stocktake write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("stocktake.stocktakecursor.written", 1)
	return rev, nil
}

// GetStocktakeCursor reads the entity with the given id.
func (r *Repo) GetStocktakeCursor(id string) (StocktakeCursor, int64, error) {
	rec, err := r.store.Get("stocktake/stocktakecursor/" + id)
	if err != nil {
		return StocktakeCursor{}, 0, err
	}
	x, err := ParseStocktakeCursor(rec.Fields)
	if err != nil {
		return StocktakeCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListStocktakeCursors reads every stored entity of this type, sorted.
func (r *Repo) ListStocktakeCursors() ([]StocktakeCursor, error) {
	recs := r.store.List("stocktake/stocktakecursor/")
	out := make([]StocktakeCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseStocktakeCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("stocktake: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortStocktakeCursors(out)
	return out, nil
}
