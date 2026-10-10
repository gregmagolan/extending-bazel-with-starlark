package delivery

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists delivery entities into a store.Store.
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
	l.Debug("delivery repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many delivery records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutDeliveryRecord writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryRecord(x DeliveryRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryrecord.written", 1)
	return rev, nil
}

// GetDeliveryRecord reads the entity with the given id.
func (r *Repo) GetDeliveryRecord(id string) (DeliveryRecord, int64, error) {
	rec, err := r.store.Get("delivery/deliveryrecord/" + id)
	if err != nil {
		return DeliveryRecord{}, 0, err
	}
	x, err := ParseDeliveryRecord(rec.Fields)
	if err != nil {
		return DeliveryRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryRecords reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryRecords() ([]DeliveryRecord, error) {
	recs := r.store.List("delivery/deliveryrecord/")
	out := make([]DeliveryRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryRecords(out)
	return out, nil
}

// PutDeliveryEntry writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryEntry(x DeliveryEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryentry.written", 1)
	return rev, nil
}

// GetDeliveryEntry reads the entity with the given id.
func (r *Repo) GetDeliveryEntry(id string) (DeliveryEntry, int64, error) {
	rec, err := r.store.Get("delivery/deliveryentry/" + id)
	if err != nil {
		return DeliveryEntry{}, 0, err
	}
	x, err := ParseDeliveryEntry(rec.Fields)
	if err != nil {
		return DeliveryEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryEntrys() ([]DeliveryEntry, error) {
	recs := r.store.List("delivery/deliveryentry/")
	out := make([]DeliveryEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryEntrys(out)
	return out, nil
}

// PutDeliveryItem writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryItem(x DeliveryItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryitem.written", 1)
	return rev, nil
}

// GetDeliveryItem reads the entity with the given id.
func (r *Repo) GetDeliveryItem(id string) (DeliveryItem, int64, error) {
	rec, err := r.store.Get("delivery/deliveryitem/" + id)
	if err != nil {
		return DeliveryItem{}, 0, err
	}
	x, err := ParseDeliveryItem(rec.Fields)
	if err != nil {
		return DeliveryItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryItems reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryItems() ([]DeliveryItem, error) {
	recs := r.store.List("delivery/deliveryitem/")
	out := make([]DeliveryItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryItems(out)
	return out, nil
}

// PutDeliveryDraft writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryDraft(x DeliveryDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverydraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverydraft.written", 1)
	return rev, nil
}

// GetDeliveryDraft reads the entity with the given id.
func (r *Repo) GetDeliveryDraft(id string) (DeliveryDraft, int64, error) {
	rec, err := r.store.Get("delivery/deliverydraft/" + id)
	if err != nil {
		return DeliveryDraft{}, 0, err
	}
	x, err := ParseDeliveryDraft(rec.Fields)
	if err != nil {
		return DeliveryDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryDrafts() ([]DeliveryDraft, error) {
	recs := r.store.List("delivery/deliverydraft/")
	out := make([]DeliveryDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryDrafts(out)
	return out, nil
}

// PutDeliveryRevision writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryRevision(x DeliveryRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryrevision.written", 1)
	return rev, nil
}

// GetDeliveryRevision reads the entity with the given id.
func (r *Repo) GetDeliveryRevision(id string) (DeliveryRevision, int64, error) {
	rec, err := r.store.Get("delivery/deliveryrevision/" + id)
	if err != nil {
		return DeliveryRevision{}, 0, err
	}
	x, err := ParseDeliveryRevision(rec.Fields)
	if err != nil {
		return DeliveryRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryRevisions() ([]DeliveryRevision, error) {
	recs := r.store.List("delivery/deliveryrevision/")
	out := make([]DeliveryRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryRevisions(out)
	return out, nil
}

// PutDeliverySnapshot writes x, creating it when revision is zero.
func (r *Repo) PutDeliverySnapshot(x DeliverySnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverysnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverysnapshot.written", 1)
	return rev, nil
}

// GetDeliverySnapshot reads the entity with the given id.
func (r *Repo) GetDeliverySnapshot(id string) (DeliverySnapshot, int64, error) {
	rec, err := r.store.Get("delivery/deliverysnapshot/" + id)
	if err != nil {
		return DeliverySnapshot{}, 0, err
	}
	x, err := ParseDeliverySnapshot(rec.Fields)
	if err != nil {
		return DeliverySnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliverySnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListDeliverySnapshots() ([]DeliverySnapshot, error) {
	recs := r.store.List("delivery/deliverysnapshot/")
	out := make([]DeliverySnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliverySnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliverySnapshots(out)
	return out, nil
}

// PutDeliveryBatch writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryBatch(x DeliveryBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverybatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverybatch.written", 1)
	return rev, nil
}

// GetDeliveryBatch reads the entity with the given id.
func (r *Repo) GetDeliveryBatch(id string) (DeliveryBatch, int64, error) {
	rec, err := r.store.Get("delivery/deliverybatch/" + id)
	if err != nil {
		return DeliveryBatch{}, 0, err
	}
	x, err := ParseDeliveryBatch(rec.Fields)
	if err != nil {
		return DeliveryBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryBatchs() ([]DeliveryBatch, error) {
	recs := r.store.List("delivery/deliverybatch/")
	out := make([]DeliveryBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryBatchs(out)
	return out, nil
}

// PutDeliveryRequest writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryRequest(x DeliveryRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryrequest.written", 1)
	return rev, nil
}

// GetDeliveryRequest reads the entity with the given id.
func (r *Repo) GetDeliveryRequest(id string) (DeliveryRequest, int64, error) {
	rec, err := r.store.Get("delivery/deliveryrequest/" + id)
	if err != nil {
		return DeliveryRequest{}, 0, err
	}
	x, err := ParseDeliveryRequest(rec.Fields)
	if err != nil {
		return DeliveryRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryRequests reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryRequests() ([]DeliveryRequest, error) {
	recs := r.store.List("delivery/deliveryrequest/")
	out := make([]DeliveryRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryRequests(out)
	return out, nil
}

// PutDeliverySummary writes x, creating it when revision is zero.
func (r *Repo) PutDeliverySummary(x DeliverySummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverysummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverysummary.written", 1)
	return rev, nil
}

// GetDeliverySummary reads the entity with the given id.
func (r *Repo) GetDeliverySummary(id string) (DeliverySummary, int64, error) {
	rec, err := r.store.Get("delivery/deliverysummary/" + id)
	if err != nil {
		return DeliverySummary{}, 0, err
	}
	x, err := ParseDeliverySummary(rec.Fields)
	if err != nil {
		return DeliverySummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliverySummarys reads every stored entity of this type, sorted.
func (r *Repo) ListDeliverySummarys() ([]DeliverySummary, error) {
	recs := r.store.List("delivery/deliverysummary/")
	out := make([]DeliverySummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliverySummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliverySummarys(out)
	return out, nil
}

// PutDeliveryPolicy writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryPolicy(x DeliveryPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverypolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverypolicy.written", 1)
	return rev, nil
}

// GetDeliveryPolicy reads the entity with the given id.
func (r *Repo) GetDeliveryPolicy(id string) (DeliveryPolicy, int64, error) {
	rec, err := r.store.Get("delivery/deliverypolicy/" + id)
	if err != nil {
		return DeliveryPolicy{}, 0, err
	}
	x, err := ParseDeliveryPolicy(rec.Fields)
	if err != nil {
		return DeliveryPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryPolicys() ([]DeliveryPolicy, error) {
	recs := r.store.List("delivery/deliverypolicy/")
	out := make([]DeliveryPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryPolicys(out)
	return out, nil
}

// PutDeliveryLink writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryLink(x DeliveryLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverylink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverylink.written", 1)
	return rev, nil
}

// GetDeliveryLink reads the entity with the given id.
func (r *Repo) GetDeliveryLink(id string) (DeliveryLink, int64, error) {
	rec, err := r.store.Get("delivery/deliverylink/" + id)
	if err != nil {
		return DeliveryLink{}, 0, err
	}
	x, err := ParseDeliveryLink(rec.Fields)
	if err != nil {
		return DeliveryLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryLinks reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryLinks() ([]DeliveryLink, error) {
	recs := r.store.List("delivery/deliverylink/")
	out := make([]DeliveryLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryLinks(out)
	return out, nil
}

// PutDeliveryNote writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryNote(x DeliveryNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverynote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverynote.written", 1)
	return rev, nil
}

// GetDeliveryNote reads the entity with the given id.
func (r *Repo) GetDeliveryNote(id string) (DeliveryNote, int64, error) {
	rec, err := r.store.Get("delivery/deliverynote/" + id)
	if err != nil {
		return DeliveryNote{}, 0, err
	}
	x, err := ParseDeliveryNote(rec.Fields)
	if err != nil {
		return DeliveryNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryNotes reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryNotes() ([]DeliveryNote, error) {
	recs := r.store.List("delivery/deliverynote/")
	out := make([]DeliveryNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryNotes(out)
	return out, nil
}

// PutDeliveryWindow writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryWindow(x DeliveryWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverywindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverywindow.written", 1)
	return rev, nil
}

// GetDeliveryWindow reads the entity with the given id.
func (r *Repo) GetDeliveryWindow(id string) (DeliveryWindow, int64, error) {
	rec, err := r.store.Get("delivery/deliverywindow/" + id)
	if err != nil {
		return DeliveryWindow{}, 0, err
	}
	x, err := ParseDeliveryWindow(rec.Fields)
	if err != nil {
		return DeliveryWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryWindows reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryWindows() ([]DeliveryWindow, error) {
	recs := r.store.List("delivery/deliverywindow/")
	out := make([]DeliveryWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryWindows(out)
	return out, nil
}

// PutDeliveryQuota writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryQuota(x DeliveryQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryquota.written", 1)
	return rev, nil
}

// GetDeliveryQuota reads the entity with the given id.
func (r *Repo) GetDeliveryQuota(id string) (DeliveryQuota, int64, error) {
	rec, err := r.store.Get("delivery/deliveryquota/" + id)
	if err != nil {
		return DeliveryQuota{}, 0, err
	}
	x, err := ParseDeliveryQuota(rec.Fields)
	if err != nil {
		return DeliveryQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryQuotas() ([]DeliveryQuota, error) {
	recs := r.store.List("delivery/deliveryquota/")
	out := make([]DeliveryQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryQuotas(out)
	return out, nil
}

// PutDeliveryBucket writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryBucket(x DeliveryBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverybucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverybucket.written", 1)
	return rev, nil
}

// GetDeliveryBucket reads the entity with the given id.
func (r *Repo) GetDeliveryBucket(id string) (DeliveryBucket, int64, error) {
	rec, err := r.store.Get("delivery/deliverybucket/" + id)
	if err != nil {
		return DeliveryBucket{}, 0, err
	}
	x, err := ParseDeliveryBucket(rec.Fields)
	if err != nil {
		return DeliveryBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryBuckets() ([]DeliveryBucket, error) {
	recs := r.store.List("delivery/deliverybucket/")
	out := make([]DeliveryBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryBuckets(out)
	return out, nil
}

// PutDeliveryLane writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryLane(x DeliveryLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverylane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverylane.written", 1)
	return rev, nil
}

// GetDeliveryLane reads the entity with the given id.
func (r *Repo) GetDeliveryLane(id string) (DeliveryLane, int64, error) {
	rec, err := r.store.Get("delivery/deliverylane/" + id)
	if err != nil {
		return DeliveryLane{}, 0, err
	}
	x, err := ParseDeliveryLane(rec.Fields)
	if err != nil {
		return DeliveryLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryLanes reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryLanes() ([]DeliveryLane, error) {
	recs := r.store.List("delivery/deliverylane/")
	out := make([]DeliveryLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryLanes(out)
	return out, nil
}

// PutDeliverySlot writes x, creating it when revision is zero.
func (r *Repo) PutDeliverySlot(x DeliverySlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryslot.written", 1)
	return rev, nil
}

// GetDeliverySlot reads the entity with the given id.
func (r *Repo) GetDeliverySlot(id string) (DeliverySlot, int64, error) {
	rec, err := r.store.Get("delivery/deliveryslot/" + id)
	if err != nil {
		return DeliverySlot{}, 0, err
	}
	x, err := ParseDeliverySlot(rec.Fields)
	if err != nil {
		return DeliverySlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliverySlots reads every stored entity of this type, sorted.
func (r *Repo) ListDeliverySlots() ([]DeliverySlot, error) {
	recs := r.store.List("delivery/deliveryslot/")
	out := make([]DeliverySlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliverySlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliverySlots(out)
	return out, nil
}

// PutDeliveryTag writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryTag(x DeliveryTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverytag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverytag.written", 1)
	return rev, nil
}

// GetDeliveryTag reads the entity with the given id.
func (r *Repo) GetDeliveryTag(id string) (DeliveryTag, int64, error) {
	rec, err := r.store.Get("delivery/deliverytag/" + id)
	if err != nil {
		return DeliveryTag{}, 0, err
	}
	x, err := ParseDeliveryTag(rec.Fields)
	if err != nil {
		return DeliveryTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryTags reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryTags() ([]DeliveryTag, error) {
	recs := r.store.List("delivery/deliverytag/")
	out := make([]DeliveryTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryTags(out)
	return out, nil
}

// PutDeliveryRule writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryRule(x DeliveryRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryrule.written", 1)
	return rev, nil
}

// GetDeliveryRule reads the entity with the given id.
func (r *Repo) GetDeliveryRule(id string) (DeliveryRule, int64, error) {
	rec, err := r.store.Get("delivery/deliveryrule/" + id)
	if err != nil {
		return DeliveryRule{}, 0, err
	}
	x, err := ParseDeliveryRule(rec.Fields)
	if err != nil {
		return DeliveryRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryRules reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryRules() ([]DeliveryRule, error) {
	recs := r.store.List("delivery/deliveryrule/")
	out := make([]DeliveryRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryRules(out)
	return out, nil
}

// PutDeliveryHold writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryHold(x DeliveryHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryhold.written", 1)
	return rev, nil
}

// GetDeliveryHold reads the entity with the given id.
func (r *Repo) GetDeliveryHold(id string) (DeliveryHold, int64, error) {
	rec, err := r.store.Get("delivery/deliveryhold/" + id)
	if err != nil {
		return DeliveryHold{}, 0, err
	}
	x, err := ParseDeliveryHold(rec.Fields)
	if err != nil {
		return DeliveryHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryHolds reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryHolds() ([]DeliveryHold, error) {
	recs := r.store.List("delivery/deliveryhold/")
	out := make([]DeliveryHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryHolds(out)
	return out, nil
}

// PutDeliveryClaim writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryClaim(x DeliveryClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliveryclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliveryclaim.written", 1)
	return rev, nil
}

// GetDeliveryClaim reads the entity with the given id.
func (r *Repo) GetDeliveryClaim(id string) (DeliveryClaim, int64, error) {
	rec, err := r.store.Get("delivery/deliveryclaim/" + id)
	if err != nil {
		return DeliveryClaim{}, 0, err
	}
	x, err := ParseDeliveryClaim(rec.Fields)
	if err != nil {
		return DeliveryClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryClaims reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryClaims() ([]DeliveryClaim, error) {
	recs := r.store.List("delivery/deliveryclaim/")
	out := make([]DeliveryClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryClaims(out)
	return out, nil
}

// PutDeliveryCursor writes x, creating it when revision is zero.
func (r *Repo) PutDeliveryCursor(x DeliveryCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("delivery.deliverycursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("delivery write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("delivery.deliverycursor.written", 1)
	return rev, nil
}

// GetDeliveryCursor reads the entity with the given id.
func (r *Repo) GetDeliveryCursor(id string) (DeliveryCursor, int64, error) {
	rec, err := r.store.Get("delivery/deliverycursor/" + id)
	if err != nil {
		return DeliveryCursor{}, 0, err
	}
	x, err := ParseDeliveryCursor(rec.Fields)
	if err != nil {
		return DeliveryCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeliveryCursors reads every stored entity of this type, sorted.
func (r *Repo) ListDeliveryCursors() ([]DeliveryCursor, error) {
	recs := r.store.List("delivery/deliverycursor/")
	out := make([]DeliveryCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeliveryCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("delivery: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeliveryCursors(out)
	return out, nil
}
