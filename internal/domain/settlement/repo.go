package settlement

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists settlement entities into a store.Store.
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
	l.Debug("settlement repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many settlement records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutSettlementRecord writes x, creating it when revision is zero.
func (r *Repo) PutSettlementRecord(x SettlementRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementrecord.written", 1)
	return rev, nil
}

// GetSettlementRecord reads the entity with the given id.
func (r *Repo) GetSettlementRecord(id string) (SettlementRecord, int64, error) {
	rec, err := r.store.Get("settlement/settlementrecord/" + id)
	if err != nil {
		return SettlementRecord{}, 0, err
	}
	x, err := ParseSettlementRecord(rec.Fields)
	if err != nil {
		return SettlementRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementRecords reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementRecords() ([]SettlementRecord, error) {
	recs := r.store.List("settlement/settlementrecord/")
	out := make([]SettlementRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementRecords(out)
	return out, nil
}

// PutSettlementEntry writes x, creating it when revision is zero.
func (r *Repo) PutSettlementEntry(x SettlementEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlemententry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlemententry.written", 1)
	return rev, nil
}

// GetSettlementEntry reads the entity with the given id.
func (r *Repo) GetSettlementEntry(id string) (SettlementEntry, int64, error) {
	rec, err := r.store.Get("settlement/settlemententry/" + id)
	if err != nil {
		return SettlementEntry{}, 0, err
	}
	x, err := ParseSettlementEntry(rec.Fields)
	if err != nil {
		return SettlementEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementEntrys() ([]SettlementEntry, error) {
	recs := r.store.List("settlement/settlemententry/")
	out := make([]SettlementEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementEntrys(out)
	return out, nil
}

// PutSettlementItem writes x, creating it when revision is zero.
func (r *Repo) PutSettlementItem(x SettlementItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementitem.written", 1)
	return rev, nil
}

// GetSettlementItem reads the entity with the given id.
func (r *Repo) GetSettlementItem(id string) (SettlementItem, int64, error) {
	rec, err := r.store.Get("settlement/settlementitem/" + id)
	if err != nil {
		return SettlementItem{}, 0, err
	}
	x, err := ParseSettlementItem(rec.Fields)
	if err != nil {
		return SettlementItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementItems reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementItems() ([]SettlementItem, error) {
	recs := r.store.List("settlement/settlementitem/")
	out := make([]SettlementItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementItems(out)
	return out, nil
}

// PutSettlementDraft writes x, creating it when revision is zero.
func (r *Repo) PutSettlementDraft(x SettlementDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementdraft.written", 1)
	return rev, nil
}

// GetSettlementDraft reads the entity with the given id.
func (r *Repo) GetSettlementDraft(id string) (SettlementDraft, int64, error) {
	rec, err := r.store.Get("settlement/settlementdraft/" + id)
	if err != nil {
		return SettlementDraft{}, 0, err
	}
	x, err := ParseSettlementDraft(rec.Fields)
	if err != nil {
		return SettlementDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementDrafts() ([]SettlementDraft, error) {
	recs := r.store.List("settlement/settlementdraft/")
	out := make([]SettlementDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementDrafts(out)
	return out, nil
}

// PutSettlementRevision writes x, creating it when revision is zero.
func (r *Repo) PutSettlementRevision(x SettlementRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementrevision.written", 1)
	return rev, nil
}

// GetSettlementRevision reads the entity with the given id.
func (r *Repo) GetSettlementRevision(id string) (SettlementRevision, int64, error) {
	rec, err := r.store.Get("settlement/settlementrevision/" + id)
	if err != nil {
		return SettlementRevision{}, 0, err
	}
	x, err := ParseSettlementRevision(rec.Fields)
	if err != nil {
		return SettlementRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementRevisions() ([]SettlementRevision, error) {
	recs := r.store.List("settlement/settlementrevision/")
	out := make([]SettlementRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementRevisions(out)
	return out, nil
}

// PutSettlementSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutSettlementSnapshot(x SettlementSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementsnapshot.written", 1)
	return rev, nil
}

// GetSettlementSnapshot reads the entity with the given id.
func (r *Repo) GetSettlementSnapshot(id string) (SettlementSnapshot, int64, error) {
	rec, err := r.store.Get("settlement/settlementsnapshot/" + id)
	if err != nil {
		return SettlementSnapshot{}, 0, err
	}
	x, err := ParseSettlementSnapshot(rec.Fields)
	if err != nil {
		return SettlementSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementSnapshots() ([]SettlementSnapshot, error) {
	recs := r.store.List("settlement/settlementsnapshot/")
	out := make([]SettlementSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementSnapshots(out)
	return out, nil
}

// PutSettlementBatch writes x, creating it when revision is zero.
func (r *Repo) PutSettlementBatch(x SettlementBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementbatch.written", 1)
	return rev, nil
}

// GetSettlementBatch reads the entity with the given id.
func (r *Repo) GetSettlementBatch(id string) (SettlementBatch, int64, error) {
	rec, err := r.store.Get("settlement/settlementbatch/" + id)
	if err != nil {
		return SettlementBatch{}, 0, err
	}
	x, err := ParseSettlementBatch(rec.Fields)
	if err != nil {
		return SettlementBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementBatchs() ([]SettlementBatch, error) {
	recs := r.store.List("settlement/settlementbatch/")
	out := make([]SettlementBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementBatchs(out)
	return out, nil
}

// PutSettlementRequest writes x, creating it when revision is zero.
func (r *Repo) PutSettlementRequest(x SettlementRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementrequest.written", 1)
	return rev, nil
}

// GetSettlementRequest reads the entity with the given id.
func (r *Repo) GetSettlementRequest(id string) (SettlementRequest, int64, error) {
	rec, err := r.store.Get("settlement/settlementrequest/" + id)
	if err != nil {
		return SettlementRequest{}, 0, err
	}
	x, err := ParseSettlementRequest(rec.Fields)
	if err != nil {
		return SettlementRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementRequests reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementRequests() ([]SettlementRequest, error) {
	recs := r.store.List("settlement/settlementrequest/")
	out := make([]SettlementRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementRequests(out)
	return out, nil
}

// PutSettlementSummary writes x, creating it when revision is zero.
func (r *Repo) PutSettlementSummary(x SettlementSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementsummary.written", 1)
	return rev, nil
}

// GetSettlementSummary reads the entity with the given id.
func (r *Repo) GetSettlementSummary(id string) (SettlementSummary, int64, error) {
	rec, err := r.store.Get("settlement/settlementsummary/" + id)
	if err != nil {
		return SettlementSummary{}, 0, err
	}
	x, err := ParseSettlementSummary(rec.Fields)
	if err != nil {
		return SettlementSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementSummarys() ([]SettlementSummary, error) {
	recs := r.store.List("settlement/settlementsummary/")
	out := make([]SettlementSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementSummarys(out)
	return out, nil
}

// PutSettlementPolicy writes x, creating it when revision is zero.
func (r *Repo) PutSettlementPolicy(x SettlementPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementpolicy.written", 1)
	return rev, nil
}

// GetSettlementPolicy reads the entity with the given id.
func (r *Repo) GetSettlementPolicy(id string) (SettlementPolicy, int64, error) {
	rec, err := r.store.Get("settlement/settlementpolicy/" + id)
	if err != nil {
		return SettlementPolicy{}, 0, err
	}
	x, err := ParseSettlementPolicy(rec.Fields)
	if err != nil {
		return SettlementPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementPolicys() ([]SettlementPolicy, error) {
	recs := r.store.List("settlement/settlementpolicy/")
	out := make([]SettlementPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementPolicys(out)
	return out, nil
}

// PutSettlementLink writes x, creating it when revision is zero.
func (r *Repo) PutSettlementLink(x SettlementLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementlink.written", 1)
	return rev, nil
}

// GetSettlementLink reads the entity with the given id.
func (r *Repo) GetSettlementLink(id string) (SettlementLink, int64, error) {
	rec, err := r.store.Get("settlement/settlementlink/" + id)
	if err != nil {
		return SettlementLink{}, 0, err
	}
	x, err := ParseSettlementLink(rec.Fields)
	if err != nil {
		return SettlementLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementLinks reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementLinks() ([]SettlementLink, error) {
	recs := r.store.List("settlement/settlementlink/")
	out := make([]SettlementLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementLinks(out)
	return out, nil
}

// PutSettlementNote writes x, creating it when revision is zero.
func (r *Repo) PutSettlementNote(x SettlementNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementnote.written", 1)
	return rev, nil
}

// GetSettlementNote reads the entity with the given id.
func (r *Repo) GetSettlementNote(id string) (SettlementNote, int64, error) {
	rec, err := r.store.Get("settlement/settlementnote/" + id)
	if err != nil {
		return SettlementNote{}, 0, err
	}
	x, err := ParseSettlementNote(rec.Fields)
	if err != nil {
		return SettlementNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementNotes reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementNotes() ([]SettlementNote, error) {
	recs := r.store.List("settlement/settlementnote/")
	out := make([]SettlementNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementNotes(out)
	return out, nil
}

// PutSettlementWindow writes x, creating it when revision is zero.
func (r *Repo) PutSettlementWindow(x SettlementWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementwindow.written", 1)
	return rev, nil
}

// GetSettlementWindow reads the entity with the given id.
func (r *Repo) GetSettlementWindow(id string) (SettlementWindow, int64, error) {
	rec, err := r.store.Get("settlement/settlementwindow/" + id)
	if err != nil {
		return SettlementWindow{}, 0, err
	}
	x, err := ParseSettlementWindow(rec.Fields)
	if err != nil {
		return SettlementWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementWindows reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementWindows() ([]SettlementWindow, error) {
	recs := r.store.List("settlement/settlementwindow/")
	out := make([]SettlementWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementWindows(out)
	return out, nil
}

// PutSettlementQuota writes x, creating it when revision is zero.
func (r *Repo) PutSettlementQuota(x SettlementQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementquota.written", 1)
	return rev, nil
}

// GetSettlementQuota reads the entity with the given id.
func (r *Repo) GetSettlementQuota(id string) (SettlementQuota, int64, error) {
	rec, err := r.store.Get("settlement/settlementquota/" + id)
	if err != nil {
		return SettlementQuota{}, 0, err
	}
	x, err := ParseSettlementQuota(rec.Fields)
	if err != nil {
		return SettlementQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementQuotas() ([]SettlementQuota, error) {
	recs := r.store.List("settlement/settlementquota/")
	out := make([]SettlementQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementQuotas(out)
	return out, nil
}

// PutSettlementBucket writes x, creating it when revision is zero.
func (r *Repo) PutSettlementBucket(x SettlementBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementbucket.written", 1)
	return rev, nil
}

// GetSettlementBucket reads the entity with the given id.
func (r *Repo) GetSettlementBucket(id string) (SettlementBucket, int64, error) {
	rec, err := r.store.Get("settlement/settlementbucket/" + id)
	if err != nil {
		return SettlementBucket{}, 0, err
	}
	x, err := ParseSettlementBucket(rec.Fields)
	if err != nil {
		return SettlementBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementBuckets() ([]SettlementBucket, error) {
	recs := r.store.List("settlement/settlementbucket/")
	out := make([]SettlementBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementBuckets(out)
	return out, nil
}

// PutSettlementLane writes x, creating it when revision is zero.
func (r *Repo) PutSettlementLane(x SettlementLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementlane.written", 1)
	return rev, nil
}

// GetSettlementLane reads the entity with the given id.
func (r *Repo) GetSettlementLane(id string) (SettlementLane, int64, error) {
	rec, err := r.store.Get("settlement/settlementlane/" + id)
	if err != nil {
		return SettlementLane{}, 0, err
	}
	x, err := ParseSettlementLane(rec.Fields)
	if err != nil {
		return SettlementLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementLanes reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementLanes() ([]SettlementLane, error) {
	recs := r.store.List("settlement/settlementlane/")
	out := make([]SettlementLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementLanes(out)
	return out, nil
}

// PutSettlementSlot writes x, creating it when revision is zero.
func (r *Repo) PutSettlementSlot(x SettlementSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementslot.written", 1)
	return rev, nil
}

// GetSettlementSlot reads the entity with the given id.
func (r *Repo) GetSettlementSlot(id string) (SettlementSlot, int64, error) {
	rec, err := r.store.Get("settlement/settlementslot/" + id)
	if err != nil {
		return SettlementSlot{}, 0, err
	}
	x, err := ParseSettlementSlot(rec.Fields)
	if err != nil {
		return SettlementSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementSlots reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementSlots() ([]SettlementSlot, error) {
	recs := r.store.List("settlement/settlementslot/")
	out := make([]SettlementSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementSlots(out)
	return out, nil
}

// PutSettlementTag writes x, creating it when revision is zero.
func (r *Repo) PutSettlementTag(x SettlementTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementtag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementtag.written", 1)
	return rev, nil
}

// GetSettlementTag reads the entity with the given id.
func (r *Repo) GetSettlementTag(id string) (SettlementTag, int64, error) {
	rec, err := r.store.Get("settlement/settlementtag/" + id)
	if err != nil {
		return SettlementTag{}, 0, err
	}
	x, err := ParseSettlementTag(rec.Fields)
	if err != nil {
		return SettlementTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementTags reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementTags() ([]SettlementTag, error) {
	recs := r.store.List("settlement/settlementtag/")
	out := make([]SettlementTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementTags(out)
	return out, nil
}

// PutSettlementRule writes x, creating it when revision is zero.
func (r *Repo) PutSettlementRule(x SettlementRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementrule.written", 1)
	return rev, nil
}

// GetSettlementRule reads the entity with the given id.
func (r *Repo) GetSettlementRule(id string) (SettlementRule, int64, error) {
	rec, err := r.store.Get("settlement/settlementrule/" + id)
	if err != nil {
		return SettlementRule{}, 0, err
	}
	x, err := ParseSettlementRule(rec.Fields)
	if err != nil {
		return SettlementRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementRules reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementRules() ([]SettlementRule, error) {
	recs := r.store.List("settlement/settlementrule/")
	out := make([]SettlementRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementRules(out)
	return out, nil
}

// PutSettlementHold writes x, creating it when revision is zero.
func (r *Repo) PutSettlementHold(x SettlementHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementhold.written", 1)
	return rev, nil
}

// GetSettlementHold reads the entity with the given id.
func (r *Repo) GetSettlementHold(id string) (SettlementHold, int64, error) {
	rec, err := r.store.Get("settlement/settlementhold/" + id)
	if err != nil {
		return SettlementHold{}, 0, err
	}
	x, err := ParseSettlementHold(rec.Fields)
	if err != nil {
		return SettlementHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementHolds reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementHolds() ([]SettlementHold, error) {
	recs := r.store.List("settlement/settlementhold/")
	out := make([]SettlementHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementHolds(out)
	return out, nil
}

// PutSettlementClaim writes x, creating it when revision is zero.
func (r *Repo) PutSettlementClaim(x SettlementClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementclaim.written", 1)
	return rev, nil
}

// GetSettlementClaim reads the entity with the given id.
func (r *Repo) GetSettlementClaim(id string) (SettlementClaim, int64, error) {
	rec, err := r.store.Get("settlement/settlementclaim/" + id)
	if err != nil {
		return SettlementClaim{}, 0, err
	}
	x, err := ParseSettlementClaim(rec.Fields)
	if err != nil {
		return SettlementClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementClaims reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementClaims() ([]SettlementClaim, error) {
	recs := r.store.List("settlement/settlementclaim/")
	out := make([]SettlementClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementClaims(out)
	return out, nil
}

// PutSettlementCursor writes x, creating it when revision is zero.
func (r *Repo) PutSettlementCursor(x SettlementCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("settlement.settlementcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("settlement write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("settlement.settlementcursor.written", 1)
	return rev, nil
}

// GetSettlementCursor reads the entity with the given id.
func (r *Repo) GetSettlementCursor(id string) (SettlementCursor, int64, error) {
	rec, err := r.store.Get("settlement/settlementcursor/" + id)
	if err != nil {
		return SettlementCursor{}, 0, err
	}
	x, err := ParseSettlementCursor(rec.Fields)
	if err != nil {
		return SettlementCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSettlementCursors reads every stored entity of this type, sorted.
func (r *Repo) ListSettlementCursors() ([]SettlementCursor, error) {
	recs := r.store.List("settlement/settlementcursor/")
	out := make([]SettlementCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSettlementCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("settlement: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSettlementCursors(out)
	return out, nil
}
