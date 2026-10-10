package carrier

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists carrier entities into a store.Store.
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
	l.Debug("carrier repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many carrier records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutCarrierRecord writes x, creating it when revision is zero.
func (r *Repo) PutCarrierRecord(x CarrierRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierrecord.written", 1)
	return rev, nil
}

// GetCarrierRecord reads the entity with the given id.
func (r *Repo) GetCarrierRecord(id string) (CarrierRecord, int64, error) {
	rec, err := r.store.Get("carrier/carrierrecord/" + id)
	if err != nil {
		return CarrierRecord{}, 0, err
	}
	x, err := ParseCarrierRecord(rec.Fields)
	if err != nil {
		return CarrierRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierRecords reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierRecords() ([]CarrierRecord, error) {
	recs := r.store.List("carrier/carrierrecord/")
	out := make([]CarrierRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierRecords(out)
	return out, nil
}

// PutCarrierEntry writes x, creating it when revision is zero.
func (r *Repo) PutCarrierEntry(x CarrierEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierentry.written", 1)
	return rev, nil
}

// GetCarrierEntry reads the entity with the given id.
func (r *Repo) GetCarrierEntry(id string) (CarrierEntry, int64, error) {
	rec, err := r.store.Get("carrier/carrierentry/" + id)
	if err != nil {
		return CarrierEntry{}, 0, err
	}
	x, err := ParseCarrierEntry(rec.Fields)
	if err != nil {
		return CarrierEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierEntrys() ([]CarrierEntry, error) {
	recs := r.store.List("carrier/carrierentry/")
	out := make([]CarrierEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierEntrys(out)
	return out, nil
}

// PutCarrierItem writes x, creating it when revision is zero.
func (r *Repo) PutCarrierItem(x CarrierItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrieritem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrieritem.written", 1)
	return rev, nil
}

// GetCarrierItem reads the entity with the given id.
func (r *Repo) GetCarrierItem(id string) (CarrierItem, int64, error) {
	rec, err := r.store.Get("carrier/carrieritem/" + id)
	if err != nil {
		return CarrierItem{}, 0, err
	}
	x, err := ParseCarrierItem(rec.Fields)
	if err != nil {
		return CarrierItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierItems reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierItems() ([]CarrierItem, error) {
	recs := r.store.List("carrier/carrieritem/")
	out := make([]CarrierItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierItems(out)
	return out, nil
}

// PutCarrierDraft writes x, creating it when revision is zero.
func (r *Repo) PutCarrierDraft(x CarrierDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierdraft.written", 1)
	return rev, nil
}

// GetCarrierDraft reads the entity with the given id.
func (r *Repo) GetCarrierDraft(id string) (CarrierDraft, int64, error) {
	rec, err := r.store.Get("carrier/carrierdraft/" + id)
	if err != nil {
		return CarrierDraft{}, 0, err
	}
	x, err := ParseCarrierDraft(rec.Fields)
	if err != nil {
		return CarrierDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierDrafts() ([]CarrierDraft, error) {
	recs := r.store.List("carrier/carrierdraft/")
	out := make([]CarrierDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierDrafts(out)
	return out, nil
}

// PutCarrierRevision writes x, creating it when revision is zero.
func (r *Repo) PutCarrierRevision(x CarrierRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierrevision.written", 1)
	return rev, nil
}

// GetCarrierRevision reads the entity with the given id.
func (r *Repo) GetCarrierRevision(id string) (CarrierRevision, int64, error) {
	rec, err := r.store.Get("carrier/carrierrevision/" + id)
	if err != nil {
		return CarrierRevision{}, 0, err
	}
	x, err := ParseCarrierRevision(rec.Fields)
	if err != nil {
		return CarrierRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierRevisions() ([]CarrierRevision, error) {
	recs := r.store.List("carrier/carrierrevision/")
	out := make([]CarrierRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierRevisions(out)
	return out, nil
}

// PutCarrierSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutCarrierSnapshot(x CarrierSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carriersnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carriersnapshot.written", 1)
	return rev, nil
}

// GetCarrierSnapshot reads the entity with the given id.
func (r *Repo) GetCarrierSnapshot(id string) (CarrierSnapshot, int64, error) {
	rec, err := r.store.Get("carrier/carriersnapshot/" + id)
	if err != nil {
		return CarrierSnapshot{}, 0, err
	}
	x, err := ParseCarrierSnapshot(rec.Fields)
	if err != nil {
		return CarrierSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierSnapshots() ([]CarrierSnapshot, error) {
	recs := r.store.List("carrier/carriersnapshot/")
	out := make([]CarrierSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierSnapshots(out)
	return out, nil
}

// PutCarrierBatch writes x, creating it when revision is zero.
func (r *Repo) PutCarrierBatch(x CarrierBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierbatch.written", 1)
	return rev, nil
}

// GetCarrierBatch reads the entity with the given id.
func (r *Repo) GetCarrierBatch(id string) (CarrierBatch, int64, error) {
	rec, err := r.store.Get("carrier/carrierbatch/" + id)
	if err != nil {
		return CarrierBatch{}, 0, err
	}
	x, err := ParseCarrierBatch(rec.Fields)
	if err != nil {
		return CarrierBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierBatchs() ([]CarrierBatch, error) {
	recs := r.store.List("carrier/carrierbatch/")
	out := make([]CarrierBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierBatchs(out)
	return out, nil
}

// PutCarrierRequest writes x, creating it when revision is zero.
func (r *Repo) PutCarrierRequest(x CarrierRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierrequest.written", 1)
	return rev, nil
}

// GetCarrierRequest reads the entity with the given id.
func (r *Repo) GetCarrierRequest(id string) (CarrierRequest, int64, error) {
	rec, err := r.store.Get("carrier/carrierrequest/" + id)
	if err != nil {
		return CarrierRequest{}, 0, err
	}
	x, err := ParseCarrierRequest(rec.Fields)
	if err != nil {
		return CarrierRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierRequests reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierRequests() ([]CarrierRequest, error) {
	recs := r.store.List("carrier/carrierrequest/")
	out := make([]CarrierRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierRequests(out)
	return out, nil
}

// PutCarrierSummary writes x, creating it when revision is zero.
func (r *Repo) PutCarrierSummary(x CarrierSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carriersummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carriersummary.written", 1)
	return rev, nil
}

// GetCarrierSummary reads the entity with the given id.
func (r *Repo) GetCarrierSummary(id string) (CarrierSummary, int64, error) {
	rec, err := r.store.Get("carrier/carriersummary/" + id)
	if err != nil {
		return CarrierSummary{}, 0, err
	}
	x, err := ParseCarrierSummary(rec.Fields)
	if err != nil {
		return CarrierSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierSummarys() ([]CarrierSummary, error) {
	recs := r.store.List("carrier/carriersummary/")
	out := make([]CarrierSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierSummarys(out)
	return out, nil
}

// PutCarrierPolicy writes x, creating it when revision is zero.
func (r *Repo) PutCarrierPolicy(x CarrierPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierpolicy.written", 1)
	return rev, nil
}

// GetCarrierPolicy reads the entity with the given id.
func (r *Repo) GetCarrierPolicy(id string) (CarrierPolicy, int64, error) {
	rec, err := r.store.Get("carrier/carrierpolicy/" + id)
	if err != nil {
		return CarrierPolicy{}, 0, err
	}
	x, err := ParseCarrierPolicy(rec.Fields)
	if err != nil {
		return CarrierPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierPolicys() ([]CarrierPolicy, error) {
	recs := r.store.List("carrier/carrierpolicy/")
	out := make([]CarrierPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierPolicys(out)
	return out, nil
}

// PutCarrierLink writes x, creating it when revision is zero.
func (r *Repo) PutCarrierLink(x CarrierLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierlink.written", 1)
	return rev, nil
}

// GetCarrierLink reads the entity with the given id.
func (r *Repo) GetCarrierLink(id string) (CarrierLink, int64, error) {
	rec, err := r.store.Get("carrier/carrierlink/" + id)
	if err != nil {
		return CarrierLink{}, 0, err
	}
	x, err := ParseCarrierLink(rec.Fields)
	if err != nil {
		return CarrierLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierLinks reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierLinks() ([]CarrierLink, error) {
	recs := r.store.List("carrier/carrierlink/")
	out := make([]CarrierLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierLinks(out)
	return out, nil
}

// PutCarrierNote writes x, creating it when revision is zero.
func (r *Repo) PutCarrierNote(x CarrierNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carriernote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carriernote.written", 1)
	return rev, nil
}

// GetCarrierNote reads the entity with the given id.
func (r *Repo) GetCarrierNote(id string) (CarrierNote, int64, error) {
	rec, err := r.store.Get("carrier/carriernote/" + id)
	if err != nil {
		return CarrierNote{}, 0, err
	}
	x, err := ParseCarrierNote(rec.Fields)
	if err != nil {
		return CarrierNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierNotes reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierNotes() ([]CarrierNote, error) {
	recs := r.store.List("carrier/carriernote/")
	out := make([]CarrierNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierNotes(out)
	return out, nil
}

// PutCarrierWindow writes x, creating it when revision is zero.
func (r *Repo) PutCarrierWindow(x CarrierWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierwindow.written", 1)
	return rev, nil
}

// GetCarrierWindow reads the entity with the given id.
func (r *Repo) GetCarrierWindow(id string) (CarrierWindow, int64, error) {
	rec, err := r.store.Get("carrier/carrierwindow/" + id)
	if err != nil {
		return CarrierWindow{}, 0, err
	}
	x, err := ParseCarrierWindow(rec.Fields)
	if err != nil {
		return CarrierWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierWindows reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierWindows() ([]CarrierWindow, error) {
	recs := r.store.List("carrier/carrierwindow/")
	out := make([]CarrierWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierWindows(out)
	return out, nil
}

// PutCarrierQuota writes x, creating it when revision is zero.
func (r *Repo) PutCarrierQuota(x CarrierQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierquota.written", 1)
	return rev, nil
}

// GetCarrierQuota reads the entity with the given id.
func (r *Repo) GetCarrierQuota(id string) (CarrierQuota, int64, error) {
	rec, err := r.store.Get("carrier/carrierquota/" + id)
	if err != nil {
		return CarrierQuota{}, 0, err
	}
	x, err := ParseCarrierQuota(rec.Fields)
	if err != nil {
		return CarrierQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierQuotas() ([]CarrierQuota, error) {
	recs := r.store.List("carrier/carrierquota/")
	out := make([]CarrierQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierQuotas(out)
	return out, nil
}

// PutCarrierBucket writes x, creating it when revision is zero.
func (r *Repo) PutCarrierBucket(x CarrierBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierbucket.written", 1)
	return rev, nil
}

// GetCarrierBucket reads the entity with the given id.
func (r *Repo) GetCarrierBucket(id string) (CarrierBucket, int64, error) {
	rec, err := r.store.Get("carrier/carrierbucket/" + id)
	if err != nil {
		return CarrierBucket{}, 0, err
	}
	x, err := ParseCarrierBucket(rec.Fields)
	if err != nil {
		return CarrierBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierBuckets() ([]CarrierBucket, error) {
	recs := r.store.List("carrier/carrierbucket/")
	out := make([]CarrierBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierBuckets(out)
	return out, nil
}

// PutCarrierLane writes x, creating it when revision is zero.
func (r *Repo) PutCarrierLane(x CarrierLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierlane.written", 1)
	return rev, nil
}

// GetCarrierLane reads the entity with the given id.
func (r *Repo) GetCarrierLane(id string) (CarrierLane, int64, error) {
	rec, err := r.store.Get("carrier/carrierlane/" + id)
	if err != nil {
		return CarrierLane{}, 0, err
	}
	x, err := ParseCarrierLane(rec.Fields)
	if err != nil {
		return CarrierLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierLanes reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierLanes() ([]CarrierLane, error) {
	recs := r.store.List("carrier/carrierlane/")
	out := make([]CarrierLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierLanes(out)
	return out, nil
}

// PutCarrierSlot writes x, creating it when revision is zero.
func (r *Repo) PutCarrierSlot(x CarrierSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierslot.written", 1)
	return rev, nil
}

// GetCarrierSlot reads the entity with the given id.
func (r *Repo) GetCarrierSlot(id string) (CarrierSlot, int64, error) {
	rec, err := r.store.Get("carrier/carrierslot/" + id)
	if err != nil {
		return CarrierSlot{}, 0, err
	}
	x, err := ParseCarrierSlot(rec.Fields)
	if err != nil {
		return CarrierSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierSlots reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierSlots() ([]CarrierSlot, error) {
	recs := r.store.List("carrier/carrierslot/")
	out := make([]CarrierSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierSlots(out)
	return out, nil
}

// PutCarrierTag writes x, creating it when revision is zero.
func (r *Repo) PutCarrierTag(x CarrierTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carriertag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carriertag.written", 1)
	return rev, nil
}

// GetCarrierTag reads the entity with the given id.
func (r *Repo) GetCarrierTag(id string) (CarrierTag, int64, error) {
	rec, err := r.store.Get("carrier/carriertag/" + id)
	if err != nil {
		return CarrierTag{}, 0, err
	}
	x, err := ParseCarrierTag(rec.Fields)
	if err != nil {
		return CarrierTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierTags reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierTags() ([]CarrierTag, error) {
	recs := r.store.List("carrier/carriertag/")
	out := make([]CarrierTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierTags(out)
	return out, nil
}

// PutCarrierRule writes x, creating it when revision is zero.
func (r *Repo) PutCarrierRule(x CarrierRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierrule.written", 1)
	return rev, nil
}

// GetCarrierRule reads the entity with the given id.
func (r *Repo) GetCarrierRule(id string) (CarrierRule, int64, error) {
	rec, err := r.store.Get("carrier/carrierrule/" + id)
	if err != nil {
		return CarrierRule{}, 0, err
	}
	x, err := ParseCarrierRule(rec.Fields)
	if err != nil {
		return CarrierRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierRules reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierRules() ([]CarrierRule, error) {
	recs := r.store.List("carrier/carrierrule/")
	out := make([]CarrierRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierRules(out)
	return out, nil
}

// PutCarrierHold writes x, creating it when revision is zero.
func (r *Repo) PutCarrierHold(x CarrierHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierhold.written", 1)
	return rev, nil
}

// GetCarrierHold reads the entity with the given id.
func (r *Repo) GetCarrierHold(id string) (CarrierHold, int64, error) {
	rec, err := r.store.Get("carrier/carrierhold/" + id)
	if err != nil {
		return CarrierHold{}, 0, err
	}
	x, err := ParseCarrierHold(rec.Fields)
	if err != nil {
		return CarrierHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierHolds reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierHolds() ([]CarrierHold, error) {
	recs := r.store.List("carrier/carrierhold/")
	out := make([]CarrierHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierHolds(out)
	return out, nil
}

// PutCarrierClaim writes x, creating it when revision is zero.
func (r *Repo) PutCarrierClaim(x CarrierClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carrierclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carrierclaim.written", 1)
	return rev, nil
}

// GetCarrierClaim reads the entity with the given id.
func (r *Repo) GetCarrierClaim(id string) (CarrierClaim, int64, error) {
	rec, err := r.store.Get("carrier/carrierclaim/" + id)
	if err != nil {
		return CarrierClaim{}, 0, err
	}
	x, err := ParseCarrierClaim(rec.Fields)
	if err != nil {
		return CarrierClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierClaims reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierClaims() ([]CarrierClaim, error) {
	recs := r.store.List("carrier/carrierclaim/")
	out := make([]CarrierClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierClaims(out)
	return out, nil
}

// PutCarrierCursor writes x, creating it when revision is zero.
func (r *Repo) PutCarrierCursor(x CarrierCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("carrier.carriercursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("carrier write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("carrier.carriercursor.written", 1)
	return rev, nil
}

// GetCarrierCursor reads the entity with the given id.
func (r *Repo) GetCarrierCursor(id string) (CarrierCursor, int64, error) {
	rec, err := r.store.Get("carrier/carriercursor/" + id)
	if err != nil {
		return CarrierCursor{}, 0, err
	}
	x, err := ParseCarrierCursor(rec.Fields)
	if err != nil {
		return CarrierCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListCarrierCursors reads every stored entity of this type, sorted.
func (r *Repo) ListCarrierCursors() ([]CarrierCursor, error) {
	recs := r.store.List("carrier/carriercursor/")
	out := make([]CarrierCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseCarrierCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("carrier: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortCarrierCursors(out)
	return out, nil
}
