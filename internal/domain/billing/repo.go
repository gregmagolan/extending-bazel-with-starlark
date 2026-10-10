package billing

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists billing entities into a store.Store.
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
	l.Debug("billing repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many billing records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutBillingRecord writes x, creating it when revision is zero.
func (r *Repo) PutBillingRecord(x BillingRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingrecord.written", 1)
	return rev, nil
}

// GetBillingRecord reads the entity with the given id.
func (r *Repo) GetBillingRecord(id string) (BillingRecord, int64, error) {
	rec, err := r.store.Get("billing/billingrecord/" + id)
	if err != nil {
		return BillingRecord{}, 0, err
	}
	x, err := ParseBillingRecord(rec.Fields)
	if err != nil {
		return BillingRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingRecords reads every stored entity of this type, sorted.
func (r *Repo) ListBillingRecords() ([]BillingRecord, error) {
	recs := r.store.List("billing/billingrecord/")
	out := make([]BillingRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingRecords(out)
	return out, nil
}

// PutBillingEntry writes x, creating it when revision is zero.
func (r *Repo) PutBillingEntry(x BillingEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingentry.written", 1)
	return rev, nil
}

// GetBillingEntry reads the entity with the given id.
func (r *Repo) GetBillingEntry(id string) (BillingEntry, int64, error) {
	rec, err := r.store.Get("billing/billingentry/" + id)
	if err != nil {
		return BillingEntry{}, 0, err
	}
	x, err := ParseBillingEntry(rec.Fields)
	if err != nil {
		return BillingEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListBillingEntrys() ([]BillingEntry, error) {
	recs := r.store.List("billing/billingentry/")
	out := make([]BillingEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingEntrys(out)
	return out, nil
}

// PutBillingItem writes x, creating it when revision is zero.
func (r *Repo) PutBillingItem(x BillingItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingitem.written", 1)
	return rev, nil
}

// GetBillingItem reads the entity with the given id.
func (r *Repo) GetBillingItem(id string) (BillingItem, int64, error) {
	rec, err := r.store.Get("billing/billingitem/" + id)
	if err != nil {
		return BillingItem{}, 0, err
	}
	x, err := ParseBillingItem(rec.Fields)
	if err != nil {
		return BillingItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingItems reads every stored entity of this type, sorted.
func (r *Repo) ListBillingItems() ([]BillingItem, error) {
	recs := r.store.List("billing/billingitem/")
	out := make([]BillingItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingItems(out)
	return out, nil
}

// PutBillingDraft writes x, creating it when revision is zero.
func (r *Repo) PutBillingDraft(x BillingDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingdraft.written", 1)
	return rev, nil
}

// GetBillingDraft reads the entity with the given id.
func (r *Repo) GetBillingDraft(id string) (BillingDraft, int64, error) {
	rec, err := r.store.Get("billing/billingdraft/" + id)
	if err != nil {
		return BillingDraft{}, 0, err
	}
	x, err := ParseBillingDraft(rec.Fields)
	if err != nil {
		return BillingDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListBillingDrafts() ([]BillingDraft, error) {
	recs := r.store.List("billing/billingdraft/")
	out := make([]BillingDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingDrafts(out)
	return out, nil
}

// PutBillingRevision writes x, creating it when revision is zero.
func (r *Repo) PutBillingRevision(x BillingRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingrevision.written", 1)
	return rev, nil
}

// GetBillingRevision reads the entity with the given id.
func (r *Repo) GetBillingRevision(id string) (BillingRevision, int64, error) {
	rec, err := r.store.Get("billing/billingrevision/" + id)
	if err != nil {
		return BillingRevision{}, 0, err
	}
	x, err := ParseBillingRevision(rec.Fields)
	if err != nil {
		return BillingRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListBillingRevisions() ([]BillingRevision, error) {
	recs := r.store.List("billing/billingrevision/")
	out := make([]BillingRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingRevisions(out)
	return out, nil
}

// PutBillingSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutBillingSnapshot(x BillingSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingsnapshot.written", 1)
	return rev, nil
}

// GetBillingSnapshot reads the entity with the given id.
func (r *Repo) GetBillingSnapshot(id string) (BillingSnapshot, int64, error) {
	rec, err := r.store.Get("billing/billingsnapshot/" + id)
	if err != nil {
		return BillingSnapshot{}, 0, err
	}
	x, err := ParseBillingSnapshot(rec.Fields)
	if err != nil {
		return BillingSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListBillingSnapshots() ([]BillingSnapshot, error) {
	recs := r.store.List("billing/billingsnapshot/")
	out := make([]BillingSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingSnapshots(out)
	return out, nil
}

// PutBillingBatch writes x, creating it when revision is zero.
func (r *Repo) PutBillingBatch(x BillingBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingbatch.written", 1)
	return rev, nil
}

// GetBillingBatch reads the entity with the given id.
func (r *Repo) GetBillingBatch(id string) (BillingBatch, int64, error) {
	rec, err := r.store.Get("billing/billingbatch/" + id)
	if err != nil {
		return BillingBatch{}, 0, err
	}
	x, err := ParseBillingBatch(rec.Fields)
	if err != nil {
		return BillingBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListBillingBatchs() ([]BillingBatch, error) {
	recs := r.store.List("billing/billingbatch/")
	out := make([]BillingBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingBatchs(out)
	return out, nil
}

// PutBillingRequest writes x, creating it when revision is zero.
func (r *Repo) PutBillingRequest(x BillingRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingrequest.written", 1)
	return rev, nil
}

// GetBillingRequest reads the entity with the given id.
func (r *Repo) GetBillingRequest(id string) (BillingRequest, int64, error) {
	rec, err := r.store.Get("billing/billingrequest/" + id)
	if err != nil {
		return BillingRequest{}, 0, err
	}
	x, err := ParseBillingRequest(rec.Fields)
	if err != nil {
		return BillingRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingRequests reads every stored entity of this type, sorted.
func (r *Repo) ListBillingRequests() ([]BillingRequest, error) {
	recs := r.store.List("billing/billingrequest/")
	out := make([]BillingRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingRequests(out)
	return out, nil
}

// PutBillingSummary writes x, creating it when revision is zero.
func (r *Repo) PutBillingSummary(x BillingSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingsummary.written", 1)
	return rev, nil
}

// GetBillingSummary reads the entity with the given id.
func (r *Repo) GetBillingSummary(id string) (BillingSummary, int64, error) {
	rec, err := r.store.Get("billing/billingsummary/" + id)
	if err != nil {
		return BillingSummary{}, 0, err
	}
	x, err := ParseBillingSummary(rec.Fields)
	if err != nil {
		return BillingSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListBillingSummarys() ([]BillingSummary, error) {
	recs := r.store.List("billing/billingsummary/")
	out := make([]BillingSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingSummarys(out)
	return out, nil
}

// PutBillingPolicy writes x, creating it when revision is zero.
func (r *Repo) PutBillingPolicy(x BillingPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingpolicy.written", 1)
	return rev, nil
}

// GetBillingPolicy reads the entity with the given id.
func (r *Repo) GetBillingPolicy(id string) (BillingPolicy, int64, error) {
	rec, err := r.store.Get("billing/billingpolicy/" + id)
	if err != nil {
		return BillingPolicy{}, 0, err
	}
	x, err := ParseBillingPolicy(rec.Fields)
	if err != nil {
		return BillingPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListBillingPolicys() ([]BillingPolicy, error) {
	recs := r.store.List("billing/billingpolicy/")
	out := make([]BillingPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingPolicys(out)
	return out, nil
}

// PutBillingLink writes x, creating it when revision is zero.
func (r *Repo) PutBillingLink(x BillingLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billinglink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billinglink.written", 1)
	return rev, nil
}

// GetBillingLink reads the entity with the given id.
func (r *Repo) GetBillingLink(id string) (BillingLink, int64, error) {
	rec, err := r.store.Get("billing/billinglink/" + id)
	if err != nil {
		return BillingLink{}, 0, err
	}
	x, err := ParseBillingLink(rec.Fields)
	if err != nil {
		return BillingLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingLinks reads every stored entity of this type, sorted.
func (r *Repo) ListBillingLinks() ([]BillingLink, error) {
	recs := r.store.List("billing/billinglink/")
	out := make([]BillingLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingLinks(out)
	return out, nil
}

// PutBillingNote writes x, creating it when revision is zero.
func (r *Repo) PutBillingNote(x BillingNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingnote.written", 1)
	return rev, nil
}

// GetBillingNote reads the entity with the given id.
func (r *Repo) GetBillingNote(id string) (BillingNote, int64, error) {
	rec, err := r.store.Get("billing/billingnote/" + id)
	if err != nil {
		return BillingNote{}, 0, err
	}
	x, err := ParseBillingNote(rec.Fields)
	if err != nil {
		return BillingNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingNotes reads every stored entity of this type, sorted.
func (r *Repo) ListBillingNotes() ([]BillingNote, error) {
	recs := r.store.List("billing/billingnote/")
	out := make([]BillingNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingNotes(out)
	return out, nil
}

// PutBillingWindow writes x, creating it when revision is zero.
func (r *Repo) PutBillingWindow(x BillingWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingwindow.written", 1)
	return rev, nil
}

// GetBillingWindow reads the entity with the given id.
func (r *Repo) GetBillingWindow(id string) (BillingWindow, int64, error) {
	rec, err := r.store.Get("billing/billingwindow/" + id)
	if err != nil {
		return BillingWindow{}, 0, err
	}
	x, err := ParseBillingWindow(rec.Fields)
	if err != nil {
		return BillingWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingWindows reads every stored entity of this type, sorted.
func (r *Repo) ListBillingWindows() ([]BillingWindow, error) {
	recs := r.store.List("billing/billingwindow/")
	out := make([]BillingWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingWindows(out)
	return out, nil
}

// PutBillingQuota writes x, creating it when revision is zero.
func (r *Repo) PutBillingQuota(x BillingQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingquota.written", 1)
	return rev, nil
}

// GetBillingQuota reads the entity with the given id.
func (r *Repo) GetBillingQuota(id string) (BillingQuota, int64, error) {
	rec, err := r.store.Get("billing/billingquota/" + id)
	if err != nil {
		return BillingQuota{}, 0, err
	}
	x, err := ParseBillingQuota(rec.Fields)
	if err != nil {
		return BillingQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListBillingQuotas() ([]BillingQuota, error) {
	recs := r.store.List("billing/billingquota/")
	out := make([]BillingQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingQuotas(out)
	return out, nil
}

// PutBillingBucket writes x, creating it when revision is zero.
func (r *Repo) PutBillingBucket(x BillingBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingbucket.written", 1)
	return rev, nil
}

// GetBillingBucket reads the entity with the given id.
func (r *Repo) GetBillingBucket(id string) (BillingBucket, int64, error) {
	rec, err := r.store.Get("billing/billingbucket/" + id)
	if err != nil {
		return BillingBucket{}, 0, err
	}
	x, err := ParseBillingBucket(rec.Fields)
	if err != nil {
		return BillingBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListBillingBuckets() ([]BillingBucket, error) {
	recs := r.store.List("billing/billingbucket/")
	out := make([]BillingBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingBuckets(out)
	return out, nil
}

// PutBillingLane writes x, creating it when revision is zero.
func (r *Repo) PutBillingLane(x BillingLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billinglane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billinglane.written", 1)
	return rev, nil
}

// GetBillingLane reads the entity with the given id.
func (r *Repo) GetBillingLane(id string) (BillingLane, int64, error) {
	rec, err := r.store.Get("billing/billinglane/" + id)
	if err != nil {
		return BillingLane{}, 0, err
	}
	x, err := ParseBillingLane(rec.Fields)
	if err != nil {
		return BillingLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingLanes reads every stored entity of this type, sorted.
func (r *Repo) ListBillingLanes() ([]BillingLane, error) {
	recs := r.store.List("billing/billinglane/")
	out := make([]BillingLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingLanes(out)
	return out, nil
}

// PutBillingSlot writes x, creating it when revision is zero.
func (r *Repo) PutBillingSlot(x BillingSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingslot.written", 1)
	return rev, nil
}

// GetBillingSlot reads the entity with the given id.
func (r *Repo) GetBillingSlot(id string) (BillingSlot, int64, error) {
	rec, err := r.store.Get("billing/billingslot/" + id)
	if err != nil {
		return BillingSlot{}, 0, err
	}
	x, err := ParseBillingSlot(rec.Fields)
	if err != nil {
		return BillingSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingSlots reads every stored entity of this type, sorted.
func (r *Repo) ListBillingSlots() ([]BillingSlot, error) {
	recs := r.store.List("billing/billingslot/")
	out := make([]BillingSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingSlots(out)
	return out, nil
}

// PutBillingTag writes x, creating it when revision is zero.
func (r *Repo) PutBillingTag(x BillingTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingtag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingtag.written", 1)
	return rev, nil
}

// GetBillingTag reads the entity with the given id.
func (r *Repo) GetBillingTag(id string) (BillingTag, int64, error) {
	rec, err := r.store.Get("billing/billingtag/" + id)
	if err != nil {
		return BillingTag{}, 0, err
	}
	x, err := ParseBillingTag(rec.Fields)
	if err != nil {
		return BillingTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingTags reads every stored entity of this type, sorted.
func (r *Repo) ListBillingTags() ([]BillingTag, error) {
	recs := r.store.List("billing/billingtag/")
	out := make([]BillingTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingTags(out)
	return out, nil
}

// PutBillingRule writes x, creating it when revision is zero.
func (r *Repo) PutBillingRule(x BillingRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingrule.written", 1)
	return rev, nil
}

// GetBillingRule reads the entity with the given id.
func (r *Repo) GetBillingRule(id string) (BillingRule, int64, error) {
	rec, err := r.store.Get("billing/billingrule/" + id)
	if err != nil {
		return BillingRule{}, 0, err
	}
	x, err := ParseBillingRule(rec.Fields)
	if err != nil {
		return BillingRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingRules reads every stored entity of this type, sorted.
func (r *Repo) ListBillingRules() ([]BillingRule, error) {
	recs := r.store.List("billing/billingrule/")
	out := make([]BillingRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingRules(out)
	return out, nil
}

// PutBillingHold writes x, creating it when revision is zero.
func (r *Repo) PutBillingHold(x BillingHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billinghold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billinghold.written", 1)
	return rev, nil
}

// GetBillingHold reads the entity with the given id.
func (r *Repo) GetBillingHold(id string) (BillingHold, int64, error) {
	rec, err := r.store.Get("billing/billinghold/" + id)
	if err != nil {
		return BillingHold{}, 0, err
	}
	x, err := ParseBillingHold(rec.Fields)
	if err != nil {
		return BillingHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingHolds reads every stored entity of this type, sorted.
func (r *Repo) ListBillingHolds() ([]BillingHold, error) {
	recs := r.store.List("billing/billinghold/")
	out := make([]BillingHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingHolds(out)
	return out, nil
}

// PutBillingClaim writes x, creating it when revision is zero.
func (r *Repo) PutBillingClaim(x BillingClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingclaim.written", 1)
	return rev, nil
}

// GetBillingClaim reads the entity with the given id.
func (r *Repo) GetBillingClaim(id string) (BillingClaim, int64, error) {
	rec, err := r.store.Get("billing/billingclaim/" + id)
	if err != nil {
		return BillingClaim{}, 0, err
	}
	x, err := ParseBillingClaim(rec.Fields)
	if err != nil {
		return BillingClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingClaims reads every stored entity of this type, sorted.
func (r *Repo) ListBillingClaims() ([]BillingClaim, error) {
	recs := r.store.List("billing/billingclaim/")
	out := make([]BillingClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingClaims(out)
	return out, nil
}

// PutBillingCursor writes x, creating it when revision is zero.
func (r *Repo) PutBillingCursor(x BillingCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("billing.billingcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("billing write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("billing.billingcursor.written", 1)
	return rev, nil
}

// GetBillingCursor reads the entity with the given id.
func (r *Repo) GetBillingCursor(id string) (BillingCursor, int64, error) {
	rec, err := r.store.Get("billing/billingcursor/" + id)
	if err != nil {
		return BillingCursor{}, 0, err
	}
	x, err := ParseBillingCursor(rec.Fields)
	if err != nil {
		return BillingCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListBillingCursors reads every stored entity of this type, sorted.
func (r *Repo) ListBillingCursors() ([]BillingCursor, error) {
	recs := r.store.List("billing/billingcursor/")
	out := make([]BillingCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseBillingCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("billing: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortBillingCursors(out)
	return out, nil
}
