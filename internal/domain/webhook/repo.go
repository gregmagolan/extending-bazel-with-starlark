package webhook

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists webhook entities into a store.Store.
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
	l.Debug("webhook repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many webhook records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutWebhookRecord writes x, creating it when revision is zero.
func (r *Repo) PutWebhookRecord(x WebhookRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookrecord.written", 1)
	return rev, nil
}

// GetWebhookRecord reads the entity with the given id.
func (r *Repo) GetWebhookRecord(id string) (WebhookRecord, int64, error) {
	rec, err := r.store.Get("webhook/webhookrecord/" + id)
	if err != nil {
		return WebhookRecord{}, 0, err
	}
	x, err := ParseWebhookRecord(rec.Fields)
	if err != nil {
		return WebhookRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookRecords reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookRecords() ([]WebhookRecord, error) {
	recs := r.store.List("webhook/webhookrecord/")
	out := make([]WebhookRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookRecords(out)
	return out, nil
}

// PutWebhookEntry writes x, creating it when revision is zero.
func (r *Repo) PutWebhookEntry(x WebhookEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookentry.written", 1)
	return rev, nil
}

// GetWebhookEntry reads the entity with the given id.
func (r *Repo) GetWebhookEntry(id string) (WebhookEntry, int64, error) {
	rec, err := r.store.Get("webhook/webhookentry/" + id)
	if err != nil {
		return WebhookEntry{}, 0, err
	}
	x, err := ParseWebhookEntry(rec.Fields)
	if err != nil {
		return WebhookEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookEntrys() ([]WebhookEntry, error) {
	recs := r.store.List("webhook/webhookentry/")
	out := make([]WebhookEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookEntrys(out)
	return out, nil
}

// PutWebhookItem writes x, creating it when revision is zero.
func (r *Repo) PutWebhookItem(x WebhookItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookitem.written", 1)
	return rev, nil
}

// GetWebhookItem reads the entity with the given id.
func (r *Repo) GetWebhookItem(id string) (WebhookItem, int64, error) {
	rec, err := r.store.Get("webhook/webhookitem/" + id)
	if err != nil {
		return WebhookItem{}, 0, err
	}
	x, err := ParseWebhookItem(rec.Fields)
	if err != nil {
		return WebhookItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookItems reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookItems() ([]WebhookItem, error) {
	recs := r.store.List("webhook/webhookitem/")
	out := make([]WebhookItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookItems(out)
	return out, nil
}

// PutWebhookDraft writes x, creating it when revision is zero.
func (r *Repo) PutWebhookDraft(x WebhookDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookdraft.written", 1)
	return rev, nil
}

// GetWebhookDraft reads the entity with the given id.
func (r *Repo) GetWebhookDraft(id string) (WebhookDraft, int64, error) {
	rec, err := r.store.Get("webhook/webhookdraft/" + id)
	if err != nil {
		return WebhookDraft{}, 0, err
	}
	x, err := ParseWebhookDraft(rec.Fields)
	if err != nil {
		return WebhookDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookDrafts() ([]WebhookDraft, error) {
	recs := r.store.List("webhook/webhookdraft/")
	out := make([]WebhookDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookDrafts(out)
	return out, nil
}

// PutWebhookRevision writes x, creating it when revision is zero.
func (r *Repo) PutWebhookRevision(x WebhookRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookrevision.written", 1)
	return rev, nil
}

// GetWebhookRevision reads the entity with the given id.
func (r *Repo) GetWebhookRevision(id string) (WebhookRevision, int64, error) {
	rec, err := r.store.Get("webhook/webhookrevision/" + id)
	if err != nil {
		return WebhookRevision{}, 0, err
	}
	x, err := ParseWebhookRevision(rec.Fields)
	if err != nil {
		return WebhookRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookRevisions() ([]WebhookRevision, error) {
	recs := r.store.List("webhook/webhookrevision/")
	out := make([]WebhookRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookRevisions(out)
	return out, nil
}

// PutWebhookSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutWebhookSnapshot(x WebhookSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhooksnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhooksnapshot.written", 1)
	return rev, nil
}

// GetWebhookSnapshot reads the entity with the given id.
func (r *Repo) GetWebhookSnapshot(id string) (WebhookSnapshot, int64, error) {
	rec, err := r.store.Get("webhook/webhooksnapshot/" + id)
	if err != nil {
		return WebhookSnapshot{}, 0, err
	}
	x, err := ParseWebhookSnapshot(rec.Fields)
	if err != nil {
		return WebhookSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookSnapshots() ([]WebhookSnapshot, error) {
	recs := r.store.List("webhook/webhooksnapshot/")
	out := make([]WebhookSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookSnapshots(out)
	return out, nil
}

// PutWebhookBatch writes x, creating it when revision is zero.
func (r *Repo) PutWebhookBatch(x WebhookBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookbatch.written", 1)
	return rev, nil
}

// GetWebhookBatch reads the entity with the given id.
func (r *Repo) GetWebhookBatch(id string) (WebhookBatch, int64, error) {
	rec, err := r.store.Get("webhook/webhookbatch/" + id)
	if err != nil {
		return WebhookBatch{}, 0, err
	}
	x, err := ParseWebhookBatch(rec.Fields)
	if err != nil {
		return WebhookBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookBatchs() ([]WebhookBatch, error) {
	recs := r.store.List("webhook/webhookbatch/")
	out := make([]WebhookBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookBatchs(out)
	return out, nil
}

// PutWebhookRequest writes x, creating it when revision is zero.
func (r *Repo) PutWebhookRequest(x WebhookRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookrequest.written", 1)
	return rev, nil
}

// GetWebhookRequest reads the entity with the given id.
func (r *Repo) GetWebhookRequest(id string) (WebhookRequest, int64, error) {
	rec, err := r.store.Get("webhook/webhookrequest/" + id)
	if err != nil {
		return WebhookRequest{}, 0, err
	}
	x, err := ParseWebhookRequest(rec.Fields)
	if err != nil {
		return WebhookRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookRequests reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookRequests() ([]WebhookRequest, error) {
	recs := r.store.List("webhook/webhookrequest/")
	out := make([]WebhookRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookRequests(out)
	return out, nil
}

// PutWebhookSummary writes x, creating it when revision is zero.
func (r *Repo) PutWebhookSummary(x WebhookSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhooksummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhooksummary.written", 1)
	return rev, nil
}

// GetWebhookSummary reads the entity with the given id.
func (r *Repo) GetWebhookSummary(id string) (WebhookSummary, int64, error) {
	rec, err := r.store.Get("webhook/webhooksummary/" + id)
	if err != nil {
		return WebhookSummary{}, 0, err
	}
	x, err := ParseWebhookSummary(rec.Fields)
	if err != nil {
		return WebhookSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookSummarys() ([]WebhookSummary, error) {
	recs := r.store.List("webhook/webhooksummary/")
	out := make([]WebhookSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookSummarys(out)
	return out, nil
}

// PutWebhookPolicy writes x, creating it when revision is zero.
func (r *Repo) PutWebhookPolicy(x WebhookPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookpolicy.written", 1)
	return rev, nil
}

// GetWebhookPolicy reads the entity with the given id.
func (r *Repo) GetWebhookPolicy(id string) (WebhookPolicy, int64, error) {
	rec, err := r.store.Get("webhook/webhookpolicy/" + id)
	if err != nil {
		return WebhookPolicy{}, 0, err
	}
	x, err := ParseWebhookPolicy(rec.Fields)
	if err != nil {
		return WebhookPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookPolicys() ([]WebhookPolicy, error) {
	recs := r.store.List("webhook/webhookpolicy/")
	out := make([]WebhookPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookPolicys(out)
	return out, nil
}

// PutWebhookLink writes x, creating it when revision is zero.
func (r *Repo) PutWebhookLink(x WebhookLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhooklink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhooklink.written", 1)
	return rev, nil
}

// GetWebhookLink reads the entity with the given id.
func (r *Repo) GetWebhookLink(id string) (WebhookLink, int64, error) {
	rec, err := r.store.Get("webhook/webhooklink/" + id)
	if err != nil {
		return WebhookLink{}, 0, err
	}
	x, err := ParseWebhookLink(rec.Fields)
	if err != nil {
		return WebhookLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookLinks reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookLinks() ([]WebhookLink, error) {
	recs := r.store.List("webhook/webhooklink/")
	out := make([]WebhookLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookLinks(out)
	return out, nil
}

// PutWebhookNote writes x, creating it when revision is zero.
func (r *Repo) PutWebhookNote(x WebhookNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhooknote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhooknote.written", 1)
	return rev, nil
}

// GetWebhookNote reads the entity with the given id.
func (r *Repo) GetWebhookNote(id string) (WebhookNote, int64, error) {
	rec, err := r.store.Get("webhook/webhooknote/" + id)
	if err != nil {
		return WebhookNote{}, 0, err
	}
	x, err := ParseWebhookNote(rec.Fields)
	if err != nil {
		return WebhookNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookNotes reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookNotes() ([]WebhookNote, error) {
	recs := r.store.List("webhook/webhooknote/")
	out := make([]WebhookNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookNotes(out)
	return out, nil
}

// PutWebhookWindow writes x, creating it when revision is zero.
func (r *Repo) PutWebhookWindow(x WebhookWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookwindow.written", 1)
	return rev, nil
}

// GetWebhookWindow reads the entity with the given id.
func (r *Repo) GetWebhookWindow(id string) (WebhookWindow, int64, error) {
	rec, err := r.store.Get("webhook/webhookwindow/" + id)
	if err != nil {
		return WebhookWindow{}, 0, err
	}
	x, err := ParseWebhookWindow(rec.Fields)
	if err != nil {
		return WebhookWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookWindows reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookWindows() ([]WebhookWindow, error) {
	recs := r.store.List("webhook/webhookwindow/")
	out := make([]WebhookWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookWindows(out)
	return out, nil
}

// PutWebhookQuota writes x, creating it when revision is zero.
func (r *Repo) PutWebhookQuota(x WebhookQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookquota.written", 1)
	return rev, nil
}

// GetWebhookQuota reads the entity with the given id.
func (r *Repo) GetWebhookQuota(id string) (WebhookQuota, int64, error) {
	rec, err := r.store.Get("webhook/webhookquota/" + id)
	if err != nil {
		return WebhookQuota{}, 0, err
	}
	x, err := ParseWebhookQuota(rec.Fields)
	if err != nil {
		return WebhookQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookQuotas() ([]WebhookQuota, error) {
	recs := r.store.List("webhook/webhookquota/")
	out := make([]WebhookQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookQuotas(out)
	return out, nil
}

// PutWebhookBucket writes x, creating it when revision is zero.
func (r *Repo) PutWebhookBucket(x WebhookBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookbucket.written", 1)
	return rev, nil
}

// GetWebhookBucket reads the entity with the given id.
func (r *Repo) GetWebhookBucket(id string) (WebhookBucket, int64, error) {
	rec, err := r.store.Get("webhook/webhookbucket/" + id)
	if err != nil {
		return WebhookBucket{}, 0, err
	}
	x, err := ParseWebhookBucket(rec.Fields)
	if err != nil {
		return WebhookBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookBuckets() ([]WebhookBucket, error) {
	recs := r.store.List("webhook/webhookbucket/")
	out := make([]WebhookBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookBuckets(out)
	return out, nil
}

// PutWebhookLane writes x, creating it when revision is zero.
func (r *Repo) PutWebhookLane(x WebhookLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhooklane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhooklane.written", 1)
	return rev, nil
}

// GetWebhookLane reads the entity with the given id.
func (r *Repo) GetWebhookLane(id string) (WebhookLane, int64, error) {
	rec, err := r.store.Get("webhook/webhooklane/" + id)
	if err != nil {
		return WebhookLane{}, 0, err
	}
	x, err := ParseWebhookLane(rec.Fields)
	if err != nil {
		return WebhookLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookLanes reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookLanes() ([]WebhookLane, error) {
	recs := r.store.List("webhook/webhooklane/")
	out := make([]WebhookLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookLanes(out)
	return out, nil
}

// PutWebhookSlot writes x, creating it when revision is zero.
func (r *Repo) PutWebhookSlot(x WebhookSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookslot.written", 1)
	return rev, nil
}

// GetWebhookSlot reads the entity with the given id.
func (r *Repo) GetWebhookSlot(id string) (WebhookSlot, int64, error) {
	rec, err := r.store.Get("webhook/webhookslot/" + id)
	if err != nil {
		return WebhookSlot{}, 0, err
	}
	x, err := ParseWebhookSlot(rec.Fields)
	if err != nil {
		return WebhookSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookSlots reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookSlots() ([]WebhookSlot, error) {
	recs := r.store.List("webhook/webhookslot/")
	out := make([]WebhookSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookSlots(out)
	return out, nil
}

// PutWebhookTag writes x, creating it when revision is zero.
func (r *Repo) PutWebhookTag(x WebhookTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhooktag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhooktag.written", 1)
	return rev, nil
}

// GetWebhookTag reads the entity with the given id.
func (r *Repo) GetWebhookTag(id string) (WebhookTag, int64, error) {
	rec, err := r.store.Get("webhook/webhooktag/" + id)
	if err != nil {
		return WebhookTag{}, 0, err
	}
	x, err := ParseWebhookTag(rec.Fields)
	if err != nil {
		return WebhookTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookTags reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookTags() ([]WebhookTag, error) {
	recs := r.store.List("webhook/webhooktag/")
	out := make([]WebhookTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookTags(out)
	return out, nil
}

// PutWebhookRule writes x, creating it when revision is zero.
func (r *Repo) PutWebhookRule(x WebhookRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookrule.written", 1)
	return rev, nil
}

// GetWebhookRule reads the entity with the given id.
func (r *Repo) GetWebhookRule(id string) (WebhookRule, int64, error) {
	rec, err := r.store.Get("webhook/webhookrule/" + id)
	if err != nil {
		return WebhookRule{}, 0, err
	}
	x, err := ParseWebhookRule(rec.Fields)
	if err != nil {
		return WebhookRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookRules reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookRules() ([]WebhookRule, error) {
	recs := r.store.List("webhook/webhookrule/")
	out := make([]WebhookRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookRules(out)
	return out, nil
}

// PutWebhookHold writes x, creating it when revision is zero.
func (r *Repo) PutWebhookHold(x WebhookHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookhold.written", 1)
	return rev, nil
}

// GetWebhookHold reads the entity with the given id.
func (r *Repo) GetWebhookHold(id string) (WebhookHold, int64, error) {
	rec, err := r.store.Get("webhook/webhookhold/" + id)
	if err != nil {
		return WebhookHold{}, 0, err
	}
	x, err := ParseWebhookHold(rec.Fields)
	if err != nil {
		return WebhookHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookHolds reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookHolds() ([]WebhookHold, error) {
	recs := r.store.List("webhook/webhookhold/")
	out := make([]WebhookHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookHolds(out)
	return out, nil
}

// PutWebhookClaim writes x, creating it when revision is zero.
func (r *Repo) PutWebhookClaim(x WebhookClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookclaim.written", 1)
	return rev, nil
}

// GetWebhookClaim reads the entity with the given id.
func (r *Repo) GetWebhookClaim(id string) (WebhookClaim, int64, error) {
	rec, err := r.store.Get("webhook/webhookclaim/" + id)
	if err != nil {
		return WebhookClaim{}, 0, err
	}
	x, err := ParseWebhookClaim(rec.Fields)
	if err != nil {
		return WebhookClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookClaims reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookClaims() ([]WebhookClaim, error) {
	recs := r.store.List("webhook/webhookclaim/")
	out := make([]WebhookClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookClaims(out)
	return out, nil
}

// PutWebhookCursor writes x, creating it when revision is zero.
func (r *Repo) PutWebhookCursor(x WebhookCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("webhook.webhookcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("webhook write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("webhook.webhookcursor.written", 1)
	return rev, nil
}

// GetWebhookCursor reads the entity with the given id.
func (r *Repo) GetWebhookCursor(id string) (WebhookCursor, int64, error) {
	rec, err := r.store.Get("webhook/webhookcursor/" + id)
	if err != nil {
		return WebhookCursor{}, 0, err
	}
	x, err := ParseWebhookCursor(rec.Fields)
	if err != nil {
		return WebhookCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWebhookCursors reads every stored entity of this type, sorted.
func (r *Repo) ListWebhookCursors() ([]WebhookCursor, error) {
	recs := r.store.List("webhook/webhookcursor/")
	out := make([]WebhookCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWebhookCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("webhook: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWebhookCursors(out)
	return out, nil
}
