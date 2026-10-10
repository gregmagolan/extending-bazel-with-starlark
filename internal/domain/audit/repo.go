package audit

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists audit entities into a store.Store.
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
	l.Debug("audit repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many audit records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutAuditRecord writes x, creating it when revision is zero.
func (r *Repo) PutAuditRecord(x AuditRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditrecord.written", 1)
	return rev, nil
}

// GetAuditRecord reads the entity with the given id.
func (r *Repo) GetAuditRecord(id string) (AuditRecord, int64, error) {
	rec, err := r.store.Get("audit/auditrecord/" + id)
	if err != nil {
		return AuditRecord{}, 0, err
	}
	x, err := ParseAuditRecord(rec.Fields)
	if err != nil {
		return AuditRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditRecords reads every stored entity of this type, sorted.
func (r *Repo) ListAuditRecords() ([]AuditRecord, error) {
	recs := r.store.List("audit/auditrecord/")
	out := make([]AuditRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditRecords(out)
	return out, nil
}

// PutAuditEntry writes x, creating it when revision is zero.
func (r *Repo) PutAuditEntry(x AuditEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditentry.written", 1)
	return rev, nil
}

// GetAuditEntry reads the entity with the given id.
func (r *Repo) GetAuditEntry(id string) (AuditEntry, int64, error) {
	rec, err := r.store.Get("audit/auditentry/" + id)
	if err != nil {
		return AuditEntry{}, 0, err
	}
	x, err := ParseAuditEntry(rec.Fields)
	if err != nil {
		return AuditEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListAuditEntrys() ([]AuditEntry, error) {
	recs := r.store.List("audit/auditentry/")
	out := make([]AuditEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditEntrys(out)
	return out, nil
}

// PutAuditItem writes x, creating it when revision is zero.
func (r *Repo) PutAuditItem(x AuditItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.audititem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.audititem.written", 1)
	return rev, nil
}

// GetAuditItem reads the entity with the given id.
func (r *Repo) GetAuditItem(id string) (AuditItem, int64, error) {
	rec, err := r.store.Get("audit/audititem/" + id)
	if err != nil {
		return AuditItem{}, 0, err
	}
	x, err := ParseAuditItem(rec.Fields)
	if err != nil {
		return AuditItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditItems reads every stored entity of this type, sorted.
func (r *Repo) ListAuditItems() ([]AuditItem, error) {
	recs := r.store.List("audit/audititem/")
	out := make([]AuditItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditItems(out)
	return out, nil
}

// PutAuditDraft writes x, creating it when revision is zero.
func (r *Repo) PutAuditDraft(x AuditDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditdraft.written", 1)
	return rev, nil
}

// GetAuditDraft reads the entity with the given id.
func (r *Repo) GetAuditDraft(id string) (AuditDraft, int64, error) {
	rec, err := r.store.Get("audit/auditdraft/" + id)
	if err != nil {
		return AuditDraft{}, 0, err
	}
	x, err := ParseAuditDraft(rec.Fields)
	if err != nil {
		return AuditDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListAuditDrafts() ([]AuditDraft, error) {
	recs := r.store.List("audit/auditdraft/")
	out := make([]AuditDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditDrafts(out)
	return out, nil
}

// PutAuditRevision writes x, creating it when revision is zero.
func (r *Repo) PutAuditRevision(x AuditRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditrevision.written", 1)
	return rev, nil
}

// GetAuditRevision reads the entity with the given id.
func (r *Repo) GetAuditRevision(id string) (AuditRevision, int64, error) {
	rec, err := r.store.Get("audit/auditrevision/" + id)
	if err != nil {
		return AuditRevision{}, 0, err
	}
	x, err := ParseAuditRevision(rec.Fields)
	if err != nil {
		return AuditRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListAuditRevisions() ([]AuditRevision, error) {
	recs := r.store.List("audit/auditrevision/")
	out := make([]AuditRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditRevisions(out)
	return out, nil
}

// PutAuditSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutAuditSnapshot(x AuditSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditsnapshot.written", 1)
	return rev, nil
}

// GetAuditSnapshot reads the entity with the given id.
func (r *Repo) GetAuditSnapshot(id string) (AuditSnapshot, int64, error) {
	rec, err := r.store.Get("audit/auditsnapshot/" + id)
	if err != nil {
		return AuditSnapshot{}, 0, err
	}
	x, err := ParseAuditSnapshot(rec.Fields)
	if err != nil {
		return AuditSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListAuditSnapshots() ([]AuditSnapshot, error) {
	recs := r.store.List("audit/auditsnapshot/")
	out := make([]AuditSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditSnapshots(out)
	return out, nil
}

// PutAuditBatch writes x, creating it when revision is zero.
func (r *Repo) PutAuditBatch(x AuditBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditbatch.written", 1)
	return rev, nil
}

// GetAuditBatch reads the entity with the given id.
func (r *Repo) GetAuditBatch(id string) (AuditBatch, int64, error) {
	rec, err := r.store.Get("audit/auditbatch/" + id)
	if err != nil {
		return AuditBatch{}, 0, err
	}
	x, err := ParseAuditBatch(rec.Fields)
	if err != nil {
		return AuditBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListAuditBatchs() ([]AuditBatch, error) {
	recs := r.store.List("audit/auditbatch/")
	out := make([]AuditBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditBatchs(out)
	return out, nil
}

// PutAuditRequest writes x, creating it when revision is zero.
func (r *Repo) PutAuditRequest(x AuditRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditrequest.written", 1)
	return rev, nil
}

// GetAuditRequest reads the entity with the given id.
func (r *Repo) GetAuditRequest(id string) (AuditRequest, int64, error) {
	rec, err := r.store.Get("audit/auditrequest/" + id)
	if err != nil {
		return AuditRequest{}, 0, err
	}
	x, err := ParseAuditRequest(rec.Fields)
	if err != nil {
		return AuditRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditRequests reads every stored entity of this type, sorted.
func (r *Repo) ListAuditRequests() ([]AuditRequest, error) {
	recs := r.store.List("audit/auditrequest/")
	out := make([]AuditRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditRequests(out)
	return out, nil
}

// PutAuditSummary writes x, creating it when revision is zero.
func (r *Repo) PutAuditSummary(x AuditSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditsummary.written", 1)
	return rev, nil
}

// GetAuditSummary reads the entity with the given id.
func (r *Repo) GetAuditSummary(id string) (AuditSummary, int64, error) {
	rec, err := r.store.Get("audit/auditsummary/" + id)
	if err != nil {
		return AuditSummary{}, 0, err
	}
	x, err := ParseAuditSummary(rec.Fields)
	if err != nil {
		return AuditSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListAuditSummarys() ([]AuditSummary, error) {
	recs := r.store.List("audit/auditsummary/")
	out := make([]AuditSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditSummarys(out)
	return out, nil
}

// PutAuditPolicy writes x, creating it when revision is zero.
func (r *Repo) PutAuditPolicy(x AuditPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditpolicy.written", 1)
	return rev, nil
}

// GetAuditPolicy reads the entity with the given id.
func (r *Repo) GetAuditPolicy(id string) (AuditPolicy, int64, error) {
	rec, err := r.store.Get("audit/auditpolicy/" + id)
	if err != nil {
		return AuditPolicy{}, 0, err
	}
	x, err := ParseAuditPolicy(rec.Fields)
	if err != nil {
		return AuditPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListAuditPolicys() ([]AuditPolicy, error) {
	recs := r.store.List("audit/auditpolicy/")
	out := make([]AuditPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditPolicys(out)
	return out, nil
}

// PutAuditLink writes x, creating it when revision is zero.
func (r *Repo) PutAuditLink(x AuditLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditlink.written", 1)
	return rev, nil
}

// GetAuditLink reads the entity with the given id.
func (r *Repo) GetAuditLink(id string) (AuditLink, int64, error) {
	rec, err := r.store.Get("audit/auditlink/" + id)
	if err != nil {
		return AuditLink{}, 0, err
	}
	x, err := ParseAuditLink(rec.Fields)
	if err != nil {
		return AuditLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditLinks reads every stored entity of this type, sorted.
func (r *Repo) ListAuditLinks() ([]AuditLink, error) {
	recs := r.store.List("audit/auditlink/")
	out := make([]AuditLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditLinks(out)
	return out, nil
}

// PutAuditNote writes x, creating it when revision is zero.
func (r *Repo) PutAuditNote(x AuditNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditnote.written", 1)
	return rev, nil
}

// GetAuditNote reads the entity with the given id.
func (r *Repo) GetAuditNote(id string) (AuditNote, int64, error) {
	rec, err := r.store.Get("audit/auditnote/" + id)
	if err != nil {
		return AuditNote{}, 0, err
	}
	x, err := ParseAuditNote(rec.Fields)
	if err != nil {
		return AuditNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditNotes reads every stored entity of this type, sorted.
func (r *Repo) ListAuditNotes() ([]AuditNote, error) {
	recs := r.store.List("audit/auditnote/")
	out := make([]AuditNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditNotes(out)
	return out, nil
}

// PutAuditWindow writes x, creating it when revision is zero.
func (r *Repo) PutAuditWindow(x AuditWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditwindow.written", 1)
	return rev, nil
}

// GetAuditWindow reads the entity with the given id.
func (r *Repo) GetAuditWindow(id string) (AuditWindow, int64, error) {
	rec, err := r.store.Get("audit/auditwindow/" + id)
	if err != nil {
		return AuditWindow{}, 0, err
	}
	x, err := ParseAuditWindow(rec.Fields)
	if err != nil {
		return AuditWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditWindows reads every stored entity of this type, sorted.
func (r *Repo) ListAuditWindows() ([]AuditWindow, error) {
	recs := r.store.List("audit/auditwindow/")
	out := make([]AuditWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditWindows(out)
	return out, nil
}

// PutAuditQuota writes x, creating it when revision is zero.
func (r *Repo) PutAuditQuota(x AuditQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditquota.written", 1)
	return rev, nil
}

// GetAuditQuota reads the entity with the given id.
func (r *Repo) GetAuditQuota(id string) (AuditQuota, int64, error) {
	rec, err := r.store.Get("audit/auditquota/" + id)
	if err != nil {
		return AuditQuota{}, 0, err
	}
	x, err := ParseAuditQuota(rec.Fields)
	if err != nil {
		return AuditQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListAuditQuotas() ([]AuditQuota, error) {
	recs := r.store.List("audit/auditquota/")
	out := make([]AuditQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditQuotas(out)
	return out, nil
}

// PutAuditBucket writes x, creating it when revision is zero.
func (r *Repo) PutAuditBucket(x AuditBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditbucket.written", 1)
	return rev, nil
}

// GetAuditBucket reads the entity with the given id.
func (r *Repo) GetAuditBucket(id string) (AuditBucket, int64, error) {
	rec, err := r.store.Get("audit/auditbucket/" + id)
	if err != nil {
		return AuditBucket{}, 0, err
	}
	x, err := ParseAuditBucket(rec.Fields)
	if err != nil {
		return AuditBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListAuditBuckets() ([]AuditBucket, error) {
	recs := r.store.List("audit/auditbucket/")
	out := make([]AuditBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditBuckets(out)
	return out, nil
}

// PutAuditLane writes x, creating it when revision is zero.
func (r *Repo) PutAuditLane(x AuditLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditlane.written", 1)
	return rev, nil
}

// GetAuditLane reads the entity with the given id.
func (r *Repo) GetAuditLane(id string) (AuditLane, int64, error) {
	rec, err := r.store.Get("audit/auditlane/" + id)
	if err != nil {
		return AuditLane{}, 0, err
	}
	x, err := ParseAuditLane(rec.Fields)
	if err != nil {
		return AuditLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditLanes reads every stored entity of this type, sorted.
func (r *Repo) ListAuditLanes() ([]AuditLane, error) {
	recs := r.store.List("audit/auditlane/")
	out := make([]AuditLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditLanes(out)
	return out, nil
}

// PutAuditSlot writes x, creating it when revision is zero.
func (r *Repo) PutAuditSlot(x AuditSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditslot.written", 1)
	return rev, nil
}

// GetAuditSlot reads the entity with the given id.
func (r *Repo) GetAuditSlot(id string) (AuditSlot, int64, error) {
	rec, err := r.store.Get("audit/auditslot/" + id)
	if err != nil {
		return AuditSlot{}, 0, err
	}
	x, err := ParseAuditSlot(rec.Fields)
	if err != nil {
		return AuditSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditSlots reads every stored entity of this type, sorted.
func (r *Repo) ListAuditSlots() ([]AuditSlot, error) {
	recs := r.store.List("audit/auditslot/")
	out := make([]AuditSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditSlots(out)
	return out, nil
}

// PutAuditTag writes x, creating it when revision is zero.
func (r *Repo) PutAuditTag(x AuditTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.audittag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.audittag.written", 1)
	return rev, nil
}

// GetAuditTag reads the entity with the given id.
func (r *Repo) GetAuditTag(id string) (AuditTag, int64, error) {
	rec, err := r.store.Get("audit/audittag/" + id)
	if err != nil {
		return AuditTag{}, 0, err
	}
	x, err := ParseAuditTag(rec.Fields)
	if err != nil {
		return AuditTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditTags reads every stored entity of this type, sorted.
func (r *Repo) ListAuditTags() ([]AuditTag, error) {
	recs := r.store.List("audit/audittag/")
	out := make([]AuditTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditTags(out)
	return out, nil
}

// PutAuditRule writes x, creating it when revision is zero.
func (r *Repo) PutAuditRule(x AuditRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditrule.written", 1)
	return rev, nil
}

// GetAuditRule reads the entity with the given id.
func (r *Repo) GetAuditRule(id string) (AuditRule, int64, error) {
	rec, err := r.store.Get("audit/auditrule/" + id)
	if err != nil {
		return AuditRule{}, 0, err
	}
	x, err := ParseAuditRule(rec.Fields)
	if err != nil {
		return AuditRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditRules reads every stored entity of this type, sorted.
func (r *Repo) ListAuditRules() ([]AuditRule, error) {
	recs := r.store.List("audit/auditrule/")
	out := make([]AuditRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditRules(out)
	return out, nil
}

// PutAuditHold writes x, creating it when revision is zero.
func (r *Repo) PutAuditHold(x AuditHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.audithold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.audithold.written", 1)
	return rev, nil
}

// GetAuditHold reads the entity with the given id.
func (r *Repo) GetAuditHold(id string) (AuditHold, int64, error) {
	rec, err := r.store.Get("audit/audithold/" + id)
	if err != nil {
		return AuditHold{}, 0, err
	}
	x, err := ParseAuditHold(rec.Fields)
	if err != nil {
		return AuditHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditHolds reads every stored entity of this type, sorted.
func (r *Repo) ListAuditHolds() ([]AuditHold, error) {
	recs := r.store.List("audit/audithold/")
	out := make([]AuditHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditHolds(out)
	return out, nil
}

// PutAuditClaim writes x, creating it when revision is zero.
func (r *Repo) PutAuditClaim(x AuditClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditclaim.written", 1)
	return rev, nil
}

// GetAuditClaim reads the entity with the given id.
func (r *Repo) GetAuditClaim(id string) (AuditClaim, int64, error) {
	rec, err := r.store.Get("audit/auditclaim/" + id)
	if err != nil {
		return AuditClaim{}, 0, err
	}
	x, err := ParseAuditClaim(rec.Fields)
	if err != nil {
		return AuditClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditClaims reads every stored entity of this type, sorted.
func (r *Repo) ListAuditClaims() ([]AuditClaim, error) {
	recs := r.store.List("audit/auditclaim/")
	out := make([]AuditClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditClaims(out)
	return out, nil
}

// PutAuditCursor writes x, creating it when revision is zero.
func (r *Repo) PutAuditCursor(x AuditCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("audit.auditcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("audit write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("audit.auditcursor.written", 1)
	return rev, nil
}

// GetAuditCursor reads the entity with the given id.
func (r *Repo) GetAuditCursor(id string) (AuditCursor, int64, error) {
	rec, err := r.store.Get("audit/auditcursor/" + id)
	if err != nil {
		return AuditCursor{}, 0, err
	}
	x, err := ParseAuditCursor(rec.Fields)
	if err != nil {
		return AuditCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListAuditCursors reads every stored entity of this type, sorted.
func (r *Repo) ListAuditCursors() ([]AuditCursor, error) {
	recs := r.store.List("audit/auditcursor/")
	out := make([]AuditCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseAuditCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("audit: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortAuditCursors(out)
	return out, nil
}
