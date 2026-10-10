package shipping

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists shipping entities into a store.Store.
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
	l.Debug("shipping repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many shipping records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutShippingRecord writes x, creating it when revision is zero.
func (r *Repo) PutShippingRecord(x ShippingRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingrecord.written", 1)
	return rev, nil
}

// GetShippingRecord reads the entity with the given id.
func (r *Repo) GetShippingRecord(id string) (ShippingRecord, int64, error) {
	rec, err := r.store.Get("shipping/shippingrecord/" + id)
	if err != nil {
		return ShippingRecord{}, 0, err
	}
	x, err := ParseShippingRecord(rec.Fields)
	if err != nil {
		return ShippingRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingRecords reads every stored entity of this type, sorted.
func (r *Repo) ListShippingRecords() ([]ShippingRecord, error) {
	recs := r.store.List("shipping/shippingrecord/")
	out := make([]ShippingRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingRecords(out)
	return out, nil
}

// PutShippingEntry writes x, creating it when revision is zero.
func (r *Repo) PutShippingEntry(x ShippingEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingentry.written", 1)
	return rev, nil
}

// GetShippingEntry reads the entity with the given id.
func (r *Repo) GetShippingEntry(id string) (ShippingEntry, int64, error) {
	rec, err := r.store.Get("shipping/shippingentry/" + id)
	if err != nil {
		return ShippingEntry{}, 0, err
	}
	x, err := ParseShippingEntry(rec.Fields)
	if err != nil {
		return ShippingEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListShippingEntrys() ([]ShippingEntry, error) {
	recs := r.store.List("shipping/shippingentry/")
	out := make([]ShippingEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingEntrys(out)
	return out, nil
}

// PutShippingItem writes x, creating it when revision is zero.
func (r *Repo) PutShippingItem(x ShippingItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingitem.written", 1)
	return rev, nil
}

// GetShippingItem reads the entity with the given id.
func (r *Repo) GetShippingItem(id string) (ShippingItem, int64, error) {
	rec, err := r.store.Get("shipping/shippingitem/" + id)
	if err != nil {
		return ShippingItem{}, 0, err
	}
	x, err := ParseShippingItem(rec.Fields)
	if err != nil {
		return ShippingItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingItems reads every stored entity of this type, sorted.
func (r *Repo) ListShippingItems() ([]ShippingItem, error) {
	recs := r.store.List("shipping/shippingitem/")
	out := make([]ShippingItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingItems(out)
	return out, nil
}

// PutShippingDraft writes x, creating it when revision is zero.
func (r *Repo) PutShippingDraft(x ShippingDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingdraft.written", 1)
	return rev, nil
}

// GetShippingDraft reads the entity with the given id.
func (r *Repo) GetShippingDraft(id string) (ShippingDraft, int64, error) {
	rec, err := r.store.Get("shipping/shippingdraft/" + id)
	if err != nil {
		return ShippingDraft{}, 0, err
	}
	x, err := ParseShippingDraft(rec.Fields)
	if err != nil {
		return ShippingDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListShippingDrafts() ([]ShippingDraft, error) {
	recs := r.store.List("shipping/shippingdraft/")
	out := make([]ShippingDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingDrafts(out)
	return out, nil
}

// PutShippingRevision writes x, creating it when revision is zero.
func (r *Repo) PutShippingRevision(x ShippingRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingrevision.written", 1)
	return rev, nil
}

// GetShippingRevision reads the entity with the given id.
func (r *Repo) GetShippingRevision(id string) (ShippingRevision, int64, error) {
	rec, err := r.store.Get("shipping/shippingrevision/" + id)
	if err != nil {
		return ShippingRevision{}, 0, err
	}
	x, err := ParseShippingRevision(rec.Fields)
	if err != nil {
		return ShippingRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListShippingRevisions() ([]ShippingRevision, error) {
	recs := r.store.List("shipping/shippingrevision/")
	out := make([]ShippingRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingRevisions(out)
	return out, nil
}

// PutShippingSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutShippingSnapshot(x ShippingSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingsnapshot.written", 1)
	return rev, nil
}

// GetShippingSnapshot reads the entity with the given id.
func (r *Repo) GetShippingSnapshot(id string) (ShippingSnapshot, int64, error) {
	rec, err := r.store.Get("shipping/shippingsnapshot/" + id)
	if err != nil {
		return ShippingSnapshot{}, 0, err
	}
	x, err := ParseShippingSnapshot(rec.Fields)
	if err != nil {
		return ShippingSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListShippingSnapshots() ([]ShippingSnapshot, error) {
	recs := r.store.List("shipping/shippingsnapshot/")
	out := make([]ShippingSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingSnapshots(out)
	return out, nil
}

// PutShippingBatch writes x, creating it when revision is zero.
func (r *Repo) PutShippingBatch(x ShippingBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingbatch.written", 1)
	return rev, nil
}

// GetShippingBatch reads the entity with the given id.
func (r *Repo) GetShippingBatch(id string) (ShippingBatch, int64, error) {
	rec, err := r.store.Get("shipping/shippingbatch/" + id)
	if err != nil {
		return ShippingBatch{}, 0, err
	}
	x, err := ParseShippingBatch(rec.Fields)
	if err != nil {
		return ShippingBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListShippingBatchs() ([]ShippingBatch, error) {
	recs := r.store.List("shipping/shippingbatch/")
	out := make([]ShippingBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingBatchs(out)
	return out, nil
}

// PutShippingRequest writes x, creating it when revision is zero.
func (r *Repo) PutShippingRequest(x ShippingRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingrequest.written", 1)
	return rev, nil
}

// GetShippingRequest reads the entity with the given id.
func (r *Repo) GetShippingRequest(id string) (ShippingRequest, int64, error) {
	rec, err := r.store.Get("shipping/shippingrequest/" + id)
	if err != nil {
		return ShippingRequest{}, 0, err
	}
	x, err := ParseShippingRequest(rec.Fields)
	if err != nil {
		return ShippingRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingRequests reads every stored entity of this type, sorted.
func (r *Repo) ListShippingRequests() ([]ShippingRequest, error) {
	recs := r.store.List("shipping/shippingrequest/")
	out := make([]ShippingRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingRequests(out)
	return out, nil
}

// PutShippingSummary writes x, creating it when revision is zero.
func (r *Repo) PutShippingSummary(x ShippingSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingsummary.written", 1)
	return rev, nil
}

// GetShippingSummary reads the entity with the given id.
func (r *Repo) GetShippingSummary(id string) (ShippingSummary, int64, error) {
	rec, err := r.store.Get("shipping/shippingsummary/" + id)
	if err != nil {
		return ShippingSummary{}, 0, err
	}
	x, err := ParseShippingSummary(rec.Fields)
	if err != nil {
		return ShippingSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListShippingSummarys() ([]ShippingSummary, error) {
	recs := r.store.List("shipping/shippingsummary/")
	out := make([]ShippingSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingSummarys(out)
	return out, nil
}

// PutShippingPolicy writes x, creating it when revision is zero.
func (r *Repo) PutShippingPolicy(x ShippingPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingpolicy.written", 1)
	return rev, nil
}

// GetShippingPolicy reads the entity with the given id.
func (r *Repo) GetShippingPolicy(id string) (ShippingPolicy, int64, error) {
	rec, err := r.store.Get("shipping/shippingpolicy/" + id)
	if err != nil {
		return ShippingPolicy{}, 0, err
	}
	x, err := ParseShippingPolicy(rec.Fields)
	if err != nil {
		return ShippingPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListShippingPolicys() ([]ShippingPolicy, error) {
	recs := r.store.List("shipping/shippingpolicy/")
	out := make([]ShippingPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingPolicys(out)
	return out, nil
}

// PutShippingLink writes x, creating it when revision is zero.
func (r *Repo) PutShippingLink(x ShippingLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippinglink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippinglink.written", 1)
	return rev, nil
}

// GetShippingLink reads the entity with the given id.
func (r *Repo) GetShippingLink(id string) (ShippingLink, int64, error) {
	rec, err := r.store.Get("shipping/shippinglink/" + id)
	if err != nil {
		return ShippingLink{}, 0, err
	}
	x, err := ParseShippingLink(rec.Fields)
	if err != nil {
		return ShippingLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingLinks reads every stored entity of this type, sorted.
func (r *Repo) ListShippingLinks() ([]ShippingLink, error) {
	recs := r.store.List("shipping/shippinglink/")
	out := make([]ShippingLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingLinks(out)
	return out, nil
}

// PutShippingNote writes x, creating it when revision is zero.
func (r *Repo) PutShippingNote(x ShippingNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingnote.written", 1)
	return rev, nil
}

// GetShippingNote reads the entity with the given id.
func (r *Repo) GetShippingNote(id string) (ShippingNote, int64, error) {
	rec, err := r.store.Get("shipping/shippingnote/" + id)
	if err != nil {
		return ShippingNote{}, 0, err
	}
	x, err := ParseShippingNote(rec.Fields)
	if err != nil {
		return ShippingNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingNotes reads every stored entity of this type, sorted.
func (r *Repo) ListShippingNotes() ([]ShippingNote, error) {
	recs := r.store.List("shipping/shippingnote/")
	out := make([]ShippingNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingNotes(out)
	return out, nil
}

// PutShippingWindow writes x, creating it when revision is zero.
func (r *Repo) PutShippingWindow(x ShippingWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingwindow.written", 1)
	return rev, nil
}

// GetShippingWindow reads the entity with the given id.
func (r *Repo) GetShippingWindow(id string) (ShippingWindow, int64, error) {
	rec, err := r.store.Get("shipping/shippingwindow/" + id)
	if err != nil {
		return ShippingWindow{}, 0, err
	}
	x, err := ParseShippingWindow(rec.Fields)
	if err != nil {
		return ShippingWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingWindows reads every stored entity of this type, sorted.
func (r *Repo) ListShippingWindows() ([]ShippingWindow, error) {
	recs := r.store.List("shipping/shippingwindow/")
	out := make([]ShippingWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingWindows(out)
	return out, nil
}

// PutShippingQuota writes x, creating it when revision is zero.
func (r *Repo) PutShippingQuota(x ShippingQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingquota.written", 1)
	return rev, nil
}

// GetShippingQuota reads the entity with the given id.
func (r *Repo) GetShippingQuota(id string) (ShippingQuota, int64, error) {
	rec, err := r.store.Get("shipping/shippingquota/" + id)
	if err != nil {
		return ShippingQuota{}, 0, err
	}
	x, err := ParseShippingQuota(rec.Fields)
	if err != nil {
		return ShippingQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListShippingQuotas() ([]ShippingQuota, error) {
	recs := r.store.List("shipping/shippingquota/")
	out := make([]ShippingQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingQuotas(out)
	return out, nil
}

// PutShippingBucket writes x, creating it when revision is zero.
func (r *Repo) PutShippingBucket(x ShippingBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingbucket.written", 1)
	return rev, nil
}

// GetShippingBucket reads the entity with the given id.
func (r *Repo) GetShippingBucket(id string) (ShippingBucket, int64, error) {
	rec, err := r.store.Get("shipping/shippingbucket/" + id)
	if err != nil {
		return ShippingBucket{}, 0, err
	}
	x, err := ParseShippingBucket(rec.Fields)
	if err != nil {
		return ShippingBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListShippingBuckets() ([]ShippingBucket, error) {
	recs := r.store.List("shipping/shippingbucket/")
	out := make([]ShippingBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingBuckets(out)
	return out, nil
}

// PutShippingLane writes x, creating it when revision is zero.
func (r *Repo) PutShippingLane(x ShippingLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippinglane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippinglane.written", 1)
	return rev, nil
}

// GetShippingLane reads the entity with the given id.
func (r *Repo) GetShippingLane(id string) (ShippingLane, int64, error) {
	rec, err := r.store.Get("shipping/shippinglane/" + id)
	if err != nil {
		return ShippingLane{}, 0, err
	}
	x, err := ParseShippingLane(rec.Fields)
	if err != nil {
		return ShippingLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingLanes reads every stored entity of this type, sorted.
func (r *Repo) ListShippingLanes() ([]ShippingLane, error) {
	recs := r.store.List("shipping/shippinglane/")
	out := make([]ShippingLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingLanes(out)
	return out, nil
}

// PutShippingSlot writes x, creating it when revision is zero.
func (r *Repo) PutShippingSlot(x ShippingSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingslot.written", 1)
	return rev, nil
}

// GetShippingSlot reads the entity with the given id.
func (r *Repo) GetShippingSlot(id string) (ShippingSlot, int64, error) {
	rec, err := r.store.Get("shipping/shippingslot/" + id)
	if err != nil {
		return ShippingSlot{}, 0, err
	}
	x, err := ParseShippingSlot(rec.Fields)
	if err != nil {
		return ShippingSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingSlots reads every stored entity of this type, sorted.
func (r *Repo) ListShippingSlots() ([]ShippingSlot, error) {
	recs := r.store.List("shipping/shippingslot/")
	out := make([]ShippingSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingSlots(out)
	return out, nil
}

// PutShippingTag writes x, creating it when revision is zero.
func (r *Repo) PutShippingTag(x ShippingTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingtag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingtag.written", 1)
	return rev, nil
}

// GetShippingTag reads the entity with the given id.
func (r *Repo) GetShippingTag(id string) (ShippingTag, int64, error) {
	rec, err := r.store.Get("shipping/shippingtag/" + id)
	if err != nil {
		return ShippingTag{}, 0, err
	}
	x, err := ParseShippingTag(rec.Fields)
	if err != nil {
		return ShippingTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingTags reads every stored entity of this type, sorted.
func (r *Repo) ListShippingTags() ([]ShippingTag, error) {
	recs := r.store.List("shipping/shippingtag/")
	out := make([]ShippingTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingTags(out)
	return out, nil
}

// PutShippingRule writes x, creating it when revision is zero.
func (r *Repo) PutShippingRule(x ShippingRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingrule.written", 1)
	return rev, nil
}

// GetShippingRule reads the entity with the given id.
func (r *Repo) GetShippingRule(id string) (ShippingRule, int64, error) {
	rec, err := r.store.Get("shipping/shippingrule/" + id)
	if err != nil {
		return ShippingRule{}, 0, err
	}
	x, err := ParseShippingRule(rec.Fields)
	if err != nil {
		return ShippingRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingRules reads every stored entity of this type, sorted.
func (r *Repo) ListShippingRules() ([]ShippingRule, error) {
	recs := r.store.List("shipping/shippingrule/")
	out := make([]ShippingRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingRules(out)
	return out, nil
}

// PutShippingHold writes x, creating it when revision is zero.
func (r *Repo) PutShippingHold(x ShippingHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippinghold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippinghold.written", 1)
	return rev, nil
}

// GetShippingHold reads the entity with the given id.
func (r *Repo) GetShippingHold(id string) (ShippingHold, int64, error) {
	rec, err := r.store.Get("shipping/shippinghold/" + id)
	if err != nil {
		return ShippingHold{}, 0, err
	}
	x, err := ParseShippingHold(rec.Fields)
	if err != nil {
		return ShippingHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingHolds reads every stored entity of this type, sorted.
func (r *Repo) ListShippingHolds() ([]ShippingHold, error) {
	recs := r.store.List("shipping/shippinghold/")
	out := make([]ShippingHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingHolds(out)
	return out, nil
}

// PutShippingClaim writes x, creating it when revision is zero.
func (r *Repo) PutShippingClaim(x ShippingClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingclaim.written", 1)
	return rev, nil
}

// GetShippingClaim reads the entity with the given id.
func (r *Repo) GetShippingClaim(id string) (ShippingClaim, int64, error) {
	rec, err := r.store.Get("shipping/shippingclaim/" + id)
	if err != nil {
		return ShippingClaim{}, 0, err
	}
	x, err := ParseShippingClaim(rec.Fields)
	if err != nil {
		return ShippingClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingClaims reads every stored entity of this type, sorted.
func (r *Repo) ListShippingClaims() ([]ShippingClaim, error) {
	recs := r.store.List("shipping/shippingclaim/")
	out := make([]ShippingClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingClaims(out)
	return out, nil
}

// PutShippingCursor writes x, creating it when revision is zero.
func (r *Repo) PutShippingCursor(x ShippingCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("shipping.shippingcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("shipping write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("shipping.shippingcursor.written", 1)
	return rev, nil
}

// GetShippingCursor reads the entity with the given id.
func (r *Repo) GetShippingCursor(id string) (ShippingCursor, int64, error) {
	rec, err := r.store.Get("shipping/shippingcursor/" + id)
	if err != nil {
		return ShippingCursor{}, 0, err
	}
	x, err := ParseShippingCursor(rec.Fields)
	if err != nil {
		return ShippingCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListShippingCursors reads every stored entity of this type, sorted.
func (r *Repo) ListShippingCursors() ([]ShippingCursor, error) {
	recs := r.store.List("shipping/shippingcursor/")
	out := make([]ShippingCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseShippingCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("shipping: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortShippingCursors(out)
	return out, nil
}
