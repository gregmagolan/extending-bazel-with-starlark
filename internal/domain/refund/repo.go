package refund

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists refund entities into a store.Store.
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
	l.Debug("refund repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many refund records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutRefundRecord writes x, creating it when revision is zero.
func (r *Repo) PutRefundRecord(x RefundRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundrecord.written", 1)
	return rev, nil
}

// GetRefundRecord reads the entity with the given id.
func (r *Repo) GetRefundRecord(id string) (RefundRecord, int64, error) {
	rec, err := r.store.Get("refund/refundrecord/" + id)
	if err != nil {
		return RefundRecord{}, 0, err
	}
	x, err := ParseRefundRecord(rec.Fields)
	if err != nil {
		return RefundRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundRecords reads every stored entity of this type, sorted.
func (r *Repo) ListRefundRecords() ([]RefundRecord, error) {
	recs := r.store.List("refund/refundrecord/")
	out := make([]RefundRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundRecords(out)
	return out, nil
}

// PutRefundEntry writes x, creating it when revision is zero.
func (r *Repo) PutRefundEntry(x RefundEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundentry.written", 1)
	return rev, nil
}

// GetRefundEntry reads the entity with the given id.
func (r *Repo) GetRefundEntry(id string) (RefundEntry, int64, error) {
	rec, err := r.store.Get("refund/refundentry/" + id)
	if err != nil {
		return RefundEntry{}, 0, err
	}
	x, err := ParseRefundEntry(rec.Fields)
	if err != nil {
		return RefundEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListRefundEntrys() ([]RefundEntry, error) {
	recs := r.store.List("refund/refundentry/")
	out := make([]RefundEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundEntrys(out)
	return out, nil
}

// PutRefundItem writes x, creating it when revision is zero.
func (r *Repo) PutRefundItem(x RefundItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refunditem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refunditem.written", 1)
	return rev, nil
}

// GetRefundItem reads the entity with the given id.
func (r *Repo) GetRefundItem(id string) (RefundItem, int64, error) {
	rec, err := r.store.Get("refund/refunditem/" + id)
	if err != nil {
		return RefundItem{}, 0, err
	}
	x, err := ParseRefundItem(rec.Fields)
	if err != nil {
		return RefundItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundItems reads every stored entity of this type, sorted.
func (r *Repo) ListRefundItems() ([]RefundItem, error) {
	recs := r.store.List("refund/refunditem/")
	out := make([]RefundItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundItems(out)
	return out, nil
}

// PutRefundDraft writes x, creating it when revision is zero.
func (r *Repo) PutRefundDraft(x RefundDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refunddraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refunddraft.written", 1)
	return rev, nil
}

// GetRefundDraft reads the entity with the given id.
func (r *Repo) GetRefundDraft(id string) (RefundDraft, int64, error) {
	rec, err := r.store.Get("refund/refunddraft/" + id)
	if err != nil {
		return RefundDraft{}, 0, err
	}
	x, err := ParseRefundDraft(rec.Fields)
	if err != nil {
		return RefundDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListRefundDrafts() ([]RefundDraft, error) {
	recs := r.store.List("refund/refunddraft/")
	out := make([]RefundDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundDrafts(out)
	return out, nil
}

// PutRefundRevision writes x, creating it when revision is zero.
func (r *Repo) PutRefundRevision(x RefundRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundrevision.written", 1)
	return rev, nil
}

// GetRefundRevision reads the entity with the given id.
func (r *Repo) GetRefundRevision(id string) (RefundRevision, int64, error) {
	rec, err := r.store.Get("refund/refundrevision/" + id)
	if err != nil {
		return RefundRevision{}, 0, err
	}
	x, err := ParseRefundRevision(rec.Fields)
	if err != nil {
		return RefundRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListRefundRevisions() ([]RefundRevision, error) {
	recs := r.store.List("refund/refundrevision/")
	out := make([]RefundRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundRevisions(out)
	return out, nil
}

// PutRefundSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutRefundSnapshot(x RefundSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundsnapshot.written", 1)
	return rev, nil
}

// GetRefundSnapshot reads the entity with the given id.
func (r *Repo) GetRefundSnapshot(id string) (RefundSnapshot, int64, error) {
	rec, err := r.store.Get("refund/refundsnapshot/" + id)
	if err != nil {
		return RefundSnapshot{}, 0, err
	}
	x, err := ParseRefundSnapshot(rec.Fields)
	if err != nil {
		return RefundSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListRefundSnapshots() ([]RefundSnapshot, error) {
	recs := r.store.List("refund/refundsnapshot/")
	out := make([]RefundSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundSnapshots(out)
	return out, nil
}

// PutRefundBatch writes x, creating it when revision is zero.
func (r *Repo) PutRefundBatch(x RefundBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundbatch.written", 1)
	return rev, nil
}

// GetRefundBatch reads the entity with the given id.
func (r *Repo) GetRefundBatch(id string) (RefundBatch, int64, error) {
	rec, err := r.store.Get("refund/refundbatch/" + id)
	if err != nil {
		return RefundBatch{}, 0, err
	}
	x, err := ParseRefundBatch(rec.Fields)
	if err != nil {
		return RefundBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListRefundBatchs() ([]RefundBatch, error) {
	recs := r.store.List("refund/refundbatch/")
	out := make([]RefundBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundBatchs(out)
	return out, nil
}

// PutRefundRequest writes x, creating it when revision is zero.
func (r *Repo) PutRefundRequest(x RefundRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundrequest.written", 1)
	return rev, nil
}

// GetRefundRequest reads the entity with the given id.
func (r *Repo) GetRefundRequest(id string) (RefundRequest, int64, error) {
	rec, err := r.store.Get("refund/refundrequest/" + id)
	if err != nil {
		return RefundRequest{}, 0, err
	}
	x, err := ParseRefundRequest(rec.Fields)
	if err != nil {
		return RefundRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundRequests reads every stored entity of this type, sorted.
func (r *Repo) ListRefundRequests() ([]RefundRequest, error) {
	recs := r.store.List("refund/refundrequest/")
	out := make([]RefundRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundRequests(out)
	return out, nil
}

// PutRefundSummary writes x, creating it when revision is zero.
func (r *Repo) PutRefundSummary(x RefundSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundsummary.written", 1)
	return rev, nil
}

// GetRefundSummary reads the entity with the given id.
func (r *Repo) GetRefundSummary(id string) (RefundSummary, int64, error) {
	rec, err := r.store.Get("refund/refundsummary/" + id)
	if err != nil {
		return RefundSummary{}, 0, err
	}
	x, err := ParseRefundSummary(rec.Fields)
	if err != nil {
		return RefundSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListRefundSummarys() ([]RefundSummary, error) {
	recs := r.store.List("refund/refundsummary/")
	out := make([]RefundSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundSummarys(out)
	return out, nil
}

// PutRefundPolicy writes x, creating it when revision is zero.
func (r *Repo) PutRefundPolicy(x RefundPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundpolicy.written", 1)
	return rev, nil
}

// GetRefundPolicy reads the entity with the given id.
func (r *Repo) GetRefundPolicy(id string) (RefundPolicy, int64, error) {
	rec, err := r.store.Get("refund/refundpolicy/" + id)
	if err != nil {
		return RefundPolicy{}, 0, err
	}
	x, err := ParseRefundPolicy(rec.Fields)
	if err != nil {
		return RefundPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListRefundPolicys() ([]RefundPolicy, error) {
	recs := r.store.List("refund/refundpolicy/")
	out := make([]RefundPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundPolicys(out)
	return out, nil
}

// PutRefundLink writes x, creating it when revision is zero.
func (r *Repo) PutRefundLink(x RefundLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundlink.written", 1)
	return rev, nil
}

// GetRefundLink reads the entity with the given id.
func (r *Repo) GetRefundLink(id string) (RefundLink, int64, error) {
	rec, err := r.store.Get("refund/refundlink/" + id)
	if err != nil {
		return RefundLink{}, 0, err
	}
	x, err := ParseRefundLink(rec.Fields)
	if err != nil {
		return RefundLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundLinks reads every stored entity of this type, sorted.
func (r *Repo) ListRefundLinks() ([]RefundLink, error) {
	recs := r.store.List("refund/refundlink/")
	out := make([]RefundLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundLinks(out)
	return out, nil
}

// PutRefundNote writes x, creating it when revision is zero.
func (r *Repo) PutRefundNote(x RefundNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundnote.written", 1)
	return rev, nil
}

// GetRefundNote reads the entity with the given id.
func (r *Repo) GetRefundNote(id string) (RefundNote, int64, error) {
	rec, err := r.store.Get("refund/refundnote/" + id)
	if err != nil {
		return RefundNote{}, 0, err
	}
	x, err := ParseRefundNote(rec.Fields)
	if err != nil {
		return RefundNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundNotes reads every stored entity of this type, sorted.
func (r *Repo) ListRefundNotes() ([]RefundNote, error) {
	recs := r.store.List("refund/refundnote/")
	out := make([]RefundNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundNotes(out)
	return out, nil
}

// PutRefundWindow writes x, creating it when revision is zero.
func (r *Repo) PutRefundWindow(x RefundWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundwindow.written", 1)
	return rev, nil
}

// GetRefundWindow reads the entity with the given id.
func (r *Repo) GetRefundWindow(id string) (RefundWindow, int64, error) {
	rec, err := r.store.Get("refund/refundwindow/" + id)
	if err != nil {
		return RefundWindow{}, 0, err
	}
	x, err := ParseRefundWindow(rec.Fields)
	if err != nil {
		return RefundWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundWindows reads every stored entity of this type, sorted.
func (r *Repo) ListRefundWindows() ([]RefundWindow, error) {
	recs := r.store.List("refund/refundwindow/")
	out := make([]RefundWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundWindows(out)
	return out, nil
}

// PutRefundQuota writes x, creating it when revision is zero.
func (r *Repo) PutRefundQuota(x RefundQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundquota.written", 1)
	return rev, nil
}

// GetRefundQuota reads the entity with the given id.
func (r *Repo) GetRefundQuota(id string) (RefundQuota, int64, error) {
	rec, err := r.store.Get("refund/refundquota/" + id)
	if err != nil {
		return RefundQuota{}, 0, err
	}
	x, err := ParseRefundQuota(rec.Fields)
	if err != nil {
		return RefundQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListRefundQuotas() ([]RefundQuota, error) {
	recs := r.store.List("refund/refundquota/")
	out := make([]RefundQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundQuotas(out)
	return out, nil
}

// PutRefundBucket writes x, creating it when revision is zero.
func (r *Repo) PutRefundBucket(x RefundBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundbucket.written", 1)
	return rev, nil
}

// GetRefundBucket reads the entity with the given id.
func (r *Repo) GetRefundBucket(id string) (RefundBucket, int64, error) {
	rec, err := r.store.Get("refund/refundbucket/" + id)
	if err != nil {
		return RefundBucket{}, 0, err
	}
	x, err := ParseRefundBucket(rec.Fields)
	if err != nil {
		return RefundBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListRefundBuckets() ([]RefundBucket, error) {
	recs := r.store.List("refund/refundbucket/")
	out := make([]RefundBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundBuckets(out)
	return out, nil
}

// PutRefundLane writes x, creating it when revision is zero.
func (r *Repo) PutRefundLane(x RefundLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundlane.written", 1)
	return rev, nil
}

// GetRefundLane reads the entity with the given id.
func (r *Repo) GetRefundLane(id string) (RefundLane, int64, error) {
	rec, err := r.store.Get("refund/refundlane/" + id)
	if err != nil {
		return RefundLane{}, 0, err
	}
	x, err := ParseRefundLane(rec.Fields)
	if err != nil {
		return RefundLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundLanes reads every stored entity of this type, sorted.
func (r *Repo) ListRefundLanes() ([]RefundLane, error) {
	recs := r.store.List("refund/refundlane/")
	out := make([]RefundLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundLanes(out)
	return out, nil
}

// PutRefundSlot writes x, creating it when revision is zero.
func (r *Repo) PutRefundSlot(x RefundSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundslot.written", 1)
	return rev, nil
}

// GetRefundSlot reads the entity with the given id.
func (r *Repo) GetRefundSlot(id string) (RefundSlot, int64, error) {
	rec, err := r.store.Get("refund/refundslot/" + id)
	if err != nil {
		return RefundSlot{}, 0, err
	}
	x, err := ParseRefundSlot(rec.Fields)
	if err != nil {
		return RefundSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundSlots reads every stored entity of this type, sorted.
func (r *Repo) ListRefundSlots() ([]RefundSlot, error) {
	recs := r.store.List("refund/refundslot/")
	out := make([]RefundSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundSlots(out)
	return out, nil
}

// PutRefundTag writes x, creating it when revision is zero.
func (r *Repo) PutRefundTag(x RefundTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundtag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundtag.written", 1)
	return rev, nil
}

// GetRefundTag reads the entity with the given id.
func (r *Repo) GetRefundTag(id string) (RefundTag, int64, error) {
	rec, err := r.store.Get("refund/refundtag/" + id)
	if err != nil {
		return RefundTag{}, 0, err
	}
	x, err := ParseRefundTag(rec.Fields)
	if err != nil {
		return RefundTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundTags reads every stored entity of this type, sorted.
func (r *Repo) ListRefundTags() ([]RefundTag, error) {
	recs := r.store.List("refund/refundtag/")
	out := make([]RefundTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundTags(out)
	return out, nil
}

// PutRefundRule writes x, creating it when revision is zero.
func (r *Repo) PutRefundRule(x RefundRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundrule.written", 1)
	return rev, nil
}

// GetRefundRule reads the entity with the given id.
func (r *Repo) GetRefundRule(id string) (RefundRule, int64, error) {
	rec, err := r.store.Get("refund/refundrule/" + id)
	if err != nil {
		return RefundRule{}, 0, err
	}
	x, err := ParseRefundRule(rec.Fields)
	if err != nil {
		return RefundRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundRules reads every stored entity of this type, sorted.
func (r *Repo) ListRefundRules() ([]RefundRule, error) {
	recs := r.store.List("refund/refundrule/")
	out := make([]RefundRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundRules(out)
	return out, nil
}

// PutRefundHold writes x, creating it when revision is zero.
func (r *Repo) PutRefundHold(x RefundHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundhold.written", 1)
	return rev, nil
}

// GetRefundHold reads the entity with the given id.
func (r *Repo) GetRefundHold(id string) (RefundHold, int64, error) {
	rec, err := r.store.Get("refund/refundhold/" + id)
	if err != nil {
		return RefundHold{}, 0, err
	}
	x, err := ParseRefundHold(rec.Fields)
	if err != nil {
		return RefundHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundHolds reads every stored entity of this type, sorted.
func (r *Repo) ListRefundHolds() ([]RefundHold, error) {
	recs := r.store.List("refund/refundhold/")
	out := make([]RefundHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundHolds(out)
	return out, nil
}

// PutRefundClaim writes x, creating it when revision is zero.
func (r *Repo) PutRefundClaim(x RefundClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundclaim.written", 1)
	return rev, nil
}

// GetRefundClaim reads the entity with the given id.
func (r *Repo) GetRefundClaim(id string) (RefundClaim, int64, error) {
	rec, err := r.store.Get("refund/refundclaim/" + id)
	if err != nil {
		return RefundClaim{}, 0, err
	}
	x, err := ParseRefundClaim(rec.Fields)
	if err != nil {
		return RefundClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundClaims reads every stored entity of this type, sorted.
func (r *Repo) ListRefundClaims() ([]RefundClaim, error) {
	recs := r.store.List("refund/refundclaim/")
	out := make([]RefundClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundClaims(out)
	return out, nil
}

// PutRefundCursor writes x, creating it when revision is zero.
func (r *Repo) PutRefundCursor(x RefundCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("refund.refundcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("refund write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("refund.refundcursor.written", 1)
	return rev, nil
}

// GetRefundCursor reads the entity with the given id.
func (r *Repo) GetRefundCursor(id string) (RefundCursor, int64, error) {
	rec, err := r.store.Get("refund/refundcursor/" + id)
	if err != nil {
		return RefundCursor{}, 0, err
	}
	x, err := ParseRefundCursor(rec.Fields)
	if err != nil {
		return RefundCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRefundCursors reads every stored entity of this type, sorted.
func (r *Repo) ListRefundCursors() ([]RefundCursor, error) {
	recs := r.store.List("refund/refundcursor/")
	out := make([]RefundCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRefundCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("refund: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRefundCursors(out)
	return out, nil
}
