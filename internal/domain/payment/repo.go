package payment

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists payment entities into a store.Store.
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
	l.Debug("payment repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many payment records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutPaymentRecord writes x, creating it when revision is zero.
func (r *Repo) PutPaymentRecord(x PaymentRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentrecord.written", 1)
	return rev, nil
}

// GetPaymentRecord reads the entity with the given id.
func (r *Repo) GetPaymentRecord(id string) (PaymentRecord, int64, error) {
	rec, err := r.store.Get("payment/paymentrecord/" + id)
	if err != nil {
		return PaymentRecord{}, 0, err
	}
	x, err := ParsePaymentRecord(rec.Fields)
	if err != nil {
		return PaymentRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentRecords reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentRecords() ([]PaymentRecord, error) {
	recs := r.store.List("payment/paymentrecord/")
	out := make([]PaymentRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentRecords(out)
	return out, nil
}

// PutPaymentEntry writes x, creating it when revision is zero.
func (r *Repo) PutPaymentEntry(x PaymentEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymententry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymententry.written", 1)
	return rev, nil
}

// GetPaymentEntry reads the entity with the given id.
func (r *Repo) GetPaymentEntry(id string) (PaymentEntry, int64, error) {
	rec, err := r.store.Get("payment/paymententry/" + id)
	if err != nil {
		return PaymentEntry{}, 0, err
	}
	x, err := ParsePaymentEntry(rec.Fields)
	if err != nil {
		return PaymentEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentEntrys() ([]PaymentEntry, error) {
	recs := r.store.List("payment/paymententry/")
	out := make([]PaymentEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentEntrys(out)
	return out, nil
}

// PutPaymentItem writes x, creating it when revision is zero.
func (r *Repo) PutPaymentItem(x PaymentItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentitem.written", 1)
	return rev, nil
}

// GetPaymentItem reads the entity with the given id.
func (r *Repo) GetPaymentItem(id string) (PaymentItem, int64, error) {
	rec, err := r.store.Get("payment/paymentitem/" + id)
	if err != nil {
		return PaymentItem{}, 0, err
	}
	x, err := ParsePaymentItem(rec.Fields)
	if err != nil {
		return PaymentItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentItems reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentItems() ([]PaymentItem, error) {
	recs := r.store.List("payment/paymentitem/")
	out := make([]PaymentItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentItems(out)
	return out, nil
}

// PutPaymentDraft writes x, creating it when revision is zero.
func (r *Repo) PutPaymentDraft(x PaymentDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentdraft.written", 1)
	return rev, nil
}

// GetPaymentDraft reads the entity with the given id.
func (r *Repo) GetPaymentDraft(id string) (PaymentDraft, int64, error) {
	rec, err := r.store.Get("payment/paymentdraft/" + id)
	if err != nil {
		return PaymentDraft{}, 0, err
	}
	x, err := ParsePaymentDraft(rec.Fields)
	if err != nil {
		return PaymentDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentDrafts() ([]PaymentDraft, error) {
	recs := r.store.List("payment/paymentdraft/")
	out := make([]PaymentDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentDrafts(out)
	return out, nil
}

// PutPaymentRevision writes x, creating it when revision is zero.
func (r *Repo) PutPaymentRevision(x PaymentRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentrevision.written", 1)
	return rev, nil
}

// GetPaymentRevision reads the entity with the given id.
func (r *Repo) GetPaymentRevision(id string) (PaymentRevision, int64, error) {
	rec, err := r.store.Get("payment/paymentrevision/" + id)
	if err != nil {
		return PaymentRevision{}, 0, err
	}
	x, err := ParsePaymentRevision(rec.Fields)
	if err != nil {
		return PaymentRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentRevisions() ([]PaymentRevision, error) {
	recs := r.store.List("payment/paymentrevision/")
	out := make([]PaymentRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentRevisions(out)
	return out, nil
}

// PutPaymentSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutPaymentSnapshot(x PaymentSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentsnapshot.written", 1)
	return rev, nil
}

// GetPaymentSnapshot reads the entity with the given id.
func (r *Repo) GetPaymentSnapshot(id string) (PaymentSnapshot, int64, error) {
	rec, err := r.store.Get("payment/paymentsnapshot/" + id)
	if err != nil {
		return PaymentSnapshot{}, 0, err
	}
	x, err := ParsePaymentSnapshot(rec.Fields)
	if err != nil {
		return PaymentSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentSnapshots() ([]PaymentSnapshot, error) {
	recs := r.store.List("payment/paymentsnapshot/")
	out := make([]PaymentSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentSnapshots(out)
	return out, nil
}

// PutPaymentBatch writes x, creating it when revision is zero.
func (r *Repo) PutPaymentBatch(x PaymentBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentbatch.written", 1)
	return rev, nil
}

// GetPaymentBatch reads the entity with the given id.
func (r *Repo) GetPaymentBatch(id string) (PaymentBatch, int64, error) {
	rec, err := r.store.Get("payment/paymentbatch/" + id)
	if err != nil {
		return PaymentBatch{}, 0, err
	}
	x, err := ParsePaymentBatch(rec.Fields)
	if err != nil {
		return PaymentBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentBatchs() ([]PaymentBatch, error) {
	recs := r.store.List("payment/paymentbatch/")
	out := make([]PaymentBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentBatchs(out)
	return out, nil
}

// PutPaymentRequest writes x, creating it when revision is zero.
func (r *Repo) PutPaymentRequest(x PaymentRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentrequest.written", 1)
	return rev, nil
}

// GetPaymentRequest reads the entity with the given id.
func (r *Repo) GetPaymentRequest(id string) (PaymentRequest, int64, error) {
	rec, err := r.store.Get("payment/paymentrequest/" + id)
	if err != nil {
		return PaymentRequest{}, 0, err
	}
	x, err := ParsePaymentRequest(rec.Fields)
	if err != nil {
		return PaymentRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentRequests reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentRequests() ([]PaymentRequest, error) {
	recs := r.store.List("payment/paymentrequest/")
	out := make([]PaymentRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentRequests(out)
	return out, nil
}

// PutPaymentSummary writes x, creating it when revision is zero.
func (r *Repo) PutPaymentSummary(x PaymentSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentsummary.written", 1)
	return rev, nil
}

// GetPaymentSummary reads the entity with the given id.
func (r *Repo) GetPaymentSummary(id string) (PaymentSummary, int64, error) {
	rec, err := r.store.Get("payment/paymentsummary/" + id)
	if err != nil {
		return PaymentSummary{}, 0, err
	}
	x, err := ParsePaymentSummary(rec.Fields)
	if err != nil {
		return PaymentSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentSummarys() ([]PaymentSummary, error) {
	recs := r.store.List("payment/paymentsummary/")
	out := make([]PaymentSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentSummarys(out)
	return out, nil
}

// PutPaymentPolicy writes x, creating it when revision is zero.
func (r *Repo) PutPaymentPolicy(x PaymentPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentpolicy.written", 1)
	return rev, nil
}

// GetPaymentPolicy reads the entity with the given id.
func (r *Repo) GetPaymentPolicy(id string) (PaymentPolicy, int64, error) {
	rec, err := r.store.Get("payment/paymentpolicy/" + id)
	if err != nil {
		return PaymentPolicy{}, 0, err
	}
	x, err := ParsePaymentPolicy(rec.Fields)
	if err != nil {
		return PaymentPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentPolicys() ([]PaymentPolicy, error) {
	recs := r.store.List("payment/paymentpolicy/")
	out := make([]PaymentPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentPolicys(out)
	return out, nil
}

// PutPaymentLink writes x, creating it when revision is zero.
func (r *Repo) PutPaymentLink(x PaymentLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentlink.written", 1)
	return rev, nil
}

// GetPaymentLink reads the entity with the given id.
func (r *Repo) GetPaymentLink(id string) (PaymentLink, int64, error) {
	rec, err := r.store.Get("payment/paymentlink/" + id)
	if err != nil {
		return PaymentLink{}, 0, err
	}
	x, err := ParsePaymentLink(rec.Fields)
	if err != nil {
		return PaymentLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentLinks reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentLinks() ([]PaymentLink, error) {
	recs := r.store.List("payment/paymentlink/")
	out := make([]PaymentLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentLinks(out)
	return out, nil
}

// PutPaymentNote writes x, creating it when revision is zero.
func (r *Repo) PutPaymentNote(x PaymentNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentnote.written", 1)
	return rev, nil
}

// GetPaymentNote reads the entity with the given id.
func (r *Repo) GetPaymentNote(id string) (PaymentNote, int64, error) {
	rec, err := r.store.Get("payment/paymentnote/" + id)
	if err != nil {
		return PaymentNote{}, 0, err
	}
	x, err := ParsePaymentNote(rec.Fields)
	if err != nil {
		return PaymentNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentNotes reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentNotes() ([]PaymentNote, error) {
	recs := r.store.List("payment/paymentnote/")
	out := make([]PaymentNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentNotes(out)
	return out, nil
}

// PutPaymentWindow writes x, creating it when revision is zero.
func (r *Repo) PutPaymentWindow(x PaymentWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentwindow.written", 1)
	return rev, nil
}

// GetPaymentWindow reads the entity with the given id.
func (r *Repo) GetPaymentWindow(id string) (PaymentWindow, int64, error) {
	rec, err := r.store.Get("payment/paymentwindow/" + id)
	if err != nil {
		return PaymentWindow{}, 0, err
	}
	x, err := ParsePaymentWindow(rec.Fields)
	if err != nil {
		return PaymentWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentWindows reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentWindows() ([]PaymentWindow, error) {
	recs := r.store.List("payment/paymentwindow/")
	out := make([]PaymentWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentWindows(out)
	return out, nil
}

// PutPaymentQuota writes x, creating it when revision is zero.
func (r *Repo) PutPaymentQuota(x PaymentQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentquota.written", 1)
	return rev, nil
}

// GetPaymentQuota reads the entity with the given id.
func (r *Repo) GetPaymentQuota(id string) (PaymentQuota, int64, error) {
	rec, err := r.store.Get("payment/paymentquota/" + id)
	if err != nil {
		return PaymentQuota{}, 0, err
	}
	x, err := ParsePaymentQuota(rec.Fields)
	if err != nil {
		return PaymentQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentQuotas() ([]PaymentQuota, error) {
	recs := r.store.List("payment/paymentquota/")
	out := make([]PaymentQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentQuotas(out)
	return out, nil
}

// PutPaymentBucket writes x, creating it when revision is zero.
func (r *Repo) PutPaymentBucket(x PaymentBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentbucket.written", 1)
	return rev, nil
}

// GetPaymentBucket reads the entity with the given id.
func (r *Repo) GetPaymentBucket(id string) (PaymentBucket, int64, error) {
	rec, err := r.store.Get("payment/paymentbucket/" + id)
	if err != nil {
		return PaymentBucket{}, 0, err
	}
	x, err := ParsePaymentBucket(rec.Fields)
	if err != nil {
		return PaymentBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentBuckets() ([]PaymentBucket, error) {
	recs := r.store.List("payment/paymentbucket/")
	out := make([]PaymentBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentBuckets(out)
	return out, nil
}

// PutPaymentLane writes x, creating it when revision is zero.
func (r *Repo) PutPaymentLane(x PaymentLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentlane.written", 1)
	return rev, nil
}

// GetPaymentLane reads the entity with the given id.
func (r *Repo) GetPaymentLane(id string) (PaymentLane, int64, error) {
	rec, err := r.store.Get("payment/paymentlane/" + id)
	if err != nil {
		return PaymentLane{}, 0, err
	}
	x, err := ParsePaymentLane(rec.Fields)
	if err != nil {
		return PaymentLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentLanes reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentLanes() ([]PaymentLane, error) {
	recs := r.store.List("payment/paymentlane/")
	out := make([]PaymentLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentLanes(out)
	return out, nil
}

// PutPaymentSlot writes x, creating it when revision is zero.
func (r *Repo) PutPaymentSlot(x PaymentSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentslot.written", 1)
	return rev, nil
}

// GetPaymentSlot reads the entity with the given id.
func (r *Repo) GetPaymentSlot(id string) (PaymentSlot, int64, error) {
	rec, err := r.store.Get("payment/paymentslot/" + id)
	if err != nil {
		return PaymentSlot{}, 0, err
	}
	x, err := ParsePaymentSlot(rec.Fields)
	if err != nil {
		return PaymentSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentSlots reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentSlots() ([]PaymentSlot, error) {
	recs := r.store.List("payment/paymentslot/")
	out := make([]PaymentSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentSlots(out)
	return out, nil
}

// PutPaymentTag writes x, creating it when revision is zero.
func (r *Repo) PutPaymentTag(x PaymentTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymenttag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymenttag.written", 1)
	return rev, nil
}

// GetPaymentTag reads the entity with the given id.
func (r *Repo) GetPaymentTag(id string) (PaymentTag, int64, error) {
	rec, err := r.store.Get("payment/paymenttag/" + id)
	if err != nil {
		return PaymentTag{}, 0, err
	}
	x, err := ParsePaymentTag(rec.Fields)
	if err != nil {
		return PaymentTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentTags reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentTags() ([]PaymentTag, error) {
	recs := r.store.List("payment/paymenttag/")
	out := make([]PaymentTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentTags(out)
	return out, nil
}

// PutPaymentRule writes x, creating it when revision is zero.
func (r *Repo) PutPaymentRule(x PaymentRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentrule.written", 1)
	return rev, nil
}

// GetPaymentRule reads the entity with the given id.
func (r *Repo) GetPaymentRule(id string) (PaymentRule, int64, error) {
	rec, err := r.store.Get("payment/paymentrule/" + id)
	if err != nil {
		return PaymentRule{}, 0, err
	}
	x, err := ParsePaymentRule(rec.Fields)
	if err != nil {
		return PaymentRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentRules reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentRules() ([]PaymentRule, error) {
	recs := r.store.List("payment/paymentrule/")
	out := make([]PaymentRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentRules(out)
	return out, nil
}

// PutPaymentHold writes x, creating it when revision is zero.
func (r *Repo) PutPaymentHold(x PaymentHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymenthold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymenthold.written", 1)
	return rev, nil
}

// GetPaymentHold reads the entity with the given id.
func (r *Repo) GetPaymentHold(id string) (PaymentHold, int64, error) {
	rec, err := r.store.Get("payment/paymenthold/" + id)
	if err != nil {
		return PaymentHold{}, 0, err
	}
	x, err := ParsePaymentHold(rec.Fields)
	if err != nil {
		return PaymentHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentHolds reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentHolds() ([]PaymentHold, error) {
	recs := r.store.List("payment/paymenthold/")
	out := make([]PaymentHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentHolds(out)
	return out, nil
}

// PutPaymentClaim writes x, creating it when revision is zero.
func (r *Repo) PutPaymentClaim(x PaymentClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentclaim.written", 1)
	return rev, nil
}

// GetPaymentClaim reads the entity with the given id.
func (r *Repo) GetPaymentClaim(id string) (PaymentClaim, int64, error) {
	rec, err := r.store.Get("payment/paymentclaim/" + id)
	if err != nil {
		return PaymentClaim{}, 0, err
	}
	x, err := ParsePaymentClaim(rec.Fields)
	if err != nil {
		return PaymentClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentClaims reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentClaims() ([]PaymentClaim, error) {
	recs := r.store.List("payment/paymentclaim/")
	out := make([]PaymentClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentClaims(out)
	return out, nil
}

// PutPaymentCursor writes x, creating it when revision is zero.
func (r *Repo) PutPaymentCursor(x PaymentCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("payment.paymentcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("payment write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("payment.paymentcursor.written", 1)
	return rev, nil
}

// GetPaymentCursor reads the entity with the given id.
func (r *Repo) GetPaymentCursor(id string) (PaymentCursor, int64, error) {
	rec, err := r.store.Get("payment/paymentcursor/" + id)
	if err != nil {
		return PaymentCursor{}, 0, err
	}
	x, err := ParsePaymentCursor(rec.Fields)
	if err != nil {
		return PaymentCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListPaymentCursors reads every stored entity of this type, sorted.
func (r *Repo) ListPaymentCursors() ([]PaymentCursor, error) {
	recs := r.store.List("payment/paymentcursor/")
	out := make([]PaymentCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParsePaymentCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("payment: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortPaymentCursors(out)
	return out, nil
}
