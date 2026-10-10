package template

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists template entities into a store.Store.
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
	l.Debug("template repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many template records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutTemplateRecord writes x, creating it when revision is zero.
func (r *Repo) PutTemplateRecord(x TemplateRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templaterecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templaterecord.written", 1)
	return rev, nil
}

// GetTemplateRecord reads the entity with the given id.
func (r *Repo) GetTemplateRecord(id string) (TemplateRecord, int64, error) {
	rec, err := r.store.Get("template/templaterecord/" + id)
	if err != nil {
		return TemplateRecord{}, 0, err
	}
	x, err := ParseTemplateRecord(rec.Fields)
	if err != nil {
		return TemplateRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateRecords reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateRecords() ([]TemplateRecord, error) {
	recs := r.store.List("template/templaterecord/")
	out := make([]TemplateRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateRecords(out)
	return out, nil
}

// PutTemplateEntry writes x, creating it when revision is zero.
func (r *Repo) PutTemplateEntry(x TemplateEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templateentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templateentry.written", 1)
	return rev, nil
}

// GetTemplateEntry reads the entity with the given id.
func (r *Repo) GetTemplateEntry(id string) (TemplateEntry, int64, error) {
	rec, err := r.store.Get("template/templateentry/" + id)
	if err != nil {
		return TemplateEntry{}, 0, err
	}
	x, err := ParseTemplateEntry(rec.Fields)
	if err != nil {
		return TemplateEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateEntrys() ([]TemplateEntry, error) {
	recs := r.store.List("template/templateentry/")
	out := make([]TemplateEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateEntrys(out)
	return out, nil
}

// PutTemplateItem writes x, creating it when revision is zero.
func (r *Repo) PutTemplateItem(x TemplateItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templateitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templateitem.written", 1)
	return rev, nil
}

// GetTemplateItem reads the entity with the given id.
func (r *Repo) GetTemplateItem(id string) (TemplateItem, int64, error) {
	rec, err := r.store.Get("template/templateitem/" + id)
	if err != nil {
		return TemplateItem{}, 0, err
	}
	x, err := ParseTemplateItem(rec.Fields)
	if err != nil {
		return TemplateItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateItems reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateItems() ([]TemplateItem, error) {
	recs := r.store.List("template/templateitem/")
	out := make([]TemplateItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateItems(out)
	return out, nil
}

// PutTemplateDraft writes x, creating it when revision is zero.
func (r *Repo) PutTemplateDraft(x TemplateDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatedraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatedraft.written", 1)
	return rev, nil
}

// GetTemplateDraft reads the entity with the given id.
func (r *Repo) GetTemplateDraft(id string) (TemplateDraft, int64, error) {
	rec, err := r.store.Get("template/templatedraft/" + id)
	if err != nil {
		return TemplateDraft{}, 0, err
	}
	x, err := ParseTemplateDraft(rec.Fields)
	if err != nil {
		return TemplateDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateDrafts() ([]TemplateDraft, error) {
	recs := r.store.List("template/templatedraft/")
	out := make([]TemplateDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateDrafts(out)
	return out, nil
}

// PutTemplateRevision writes x, creating it when revision is zero.
func (r *Repo) PutTemplateRevision(x TemplateRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templaterevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templaterevision.written", 1)
	return rev, nil
}

// GetTemplateRevision reads the entity with the given id.
func (r *Repo) GetTemplateRevision(id string) (TemplateRevision, int64, error) {
	rec, err := r.store.Get("template/templaterevision/" + id)
	if err != nil {
		return TemplateRevision{}, 0, err
	}
	x, err := ParseTemplateRevision(rec.Fields)
	if err != nil {
		return TemplateRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateRevisions() ([]TemplateRevision, error) {
	recs := r.store.List("template/templaterevision/")
	out := make([]TemplateRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateRevisions(out)
	return out, nil
}

// PutTemplateSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutTemplateSnapshot(x TemplateSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatesnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatesnapshot.written", 1)
	return rev, nil
}

// GetTemplateSnapshot reads the entity with the given id.
func (r *Repo) GetTemplateSnapshot(id string) (TemplateSnapshot, int64, error) {
	rec, err := r.store.Get("template/templatesnapshot/" + id)
	if err != nil {
		return TemplateSnapshot{}, 0, err
	}
	x, err := ParseTemplateSnapshot(rec.Fields)
	if err != nil {
		return TemplateSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateSnapshots() ([]TemplateSnapshot, error) {
	recs := r.store.List("template/templatesnapshot/")
	out := make([]TemplateSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateSnapshots(out)
	return out, nil
}

// PutTemplateBatch writes x, creating it when revision is zero.
func (r *Repo) PutTemplateBatch(x TemplateBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatebatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatebatch.written", 1)
	return rev, nil
}

// GetTemplateBatch reads the entity with the given id.
func (r *Repo) GetTemplateBatch(id string) (TemplateBatch, int64, error) {
	rec, err := r.store.Get("template/templatebatch/" + id)
	if err != nil {
		return TemplateBatch{}, 0, err
	}
	x, err := ParseTemplateBatch(rec.Fields)
	if err != nil {
		return TemplateBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateBatchs() ([]TemplateBatch, error) {
	recs := r.store.List("template/templatebatch/")
	out := make([]TemplateBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateBatchs(out)
	return out, nil
}

// PutTemplateRequest writes x, creating it when revision is zero.
func (r *Repo) PutTemplateRequest(x TemplateRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templaterequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templaterequest.written", 1)
	return rev, nil
}

// GetTemplateRequest reads the entity with the given id.
func (r *Repo) GetTemplateRequest(id string) (TemplateRequest, int64, error) {
	rec, err := r.store.Get("template/templaterequest/" + id)
	if err != nil {
		return TemplateRequest{}, 0, err
	}
	x, err := ParseTemplateRequest(rec.Fields)
	if err != nil {
		return TemplateRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateRequests reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateRequests() ([]TemplateRequest, error) {
	recs := r.store.List("template/templaterequest/")
	out := make([]TemplateRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateRequests(out)
	return out, nil
}

// PutTemplateSummary writes x, creating it when revision is zero.
func (r *Repo) PutTemplateSummary(x TemplateSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatesummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatesummary.written", 1)
	return rev, nil
}

// GetTemplateSummary reads the entity with the given id.
func (r *Repo) GetTemplateSummary(id string) (TemplateSummary, int64, error) {
	rec, err := r.store.Get("template/templatesummary/" + id)
	if err != nil {
		return TemplateSummary{}, 0, err
	}
	x, err := ParseTemplateSummary(rec.Fields)
	if err != nil {
		return TemplateSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateSummarys() ([]TemplateSummary, error) {
	recs := r.store.List("template/templatesummary/")
	out := make([]TemplateSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateSummarys(out)
	return out, nil
}

// PutTemplatePolicy writes x, creating it when revision is zero.
func (r *Repo) PutTemplatePolicy(x TemplatePolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatepolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatepolicy.written", 1)
	return rev, nil
}

// GetTemplatePolicy reads the entity with the given id.
func (r *Repo) GetTemplatePolicy(id string) (TemplatePolicy, int64, error) {
	rec, err := r.store.Get("template/templatepolicy/" + id)
	if err != nil {
		return TemplatePolicy{}, 0, err
	}
	x, err := ParseTemplatePolicy(rec.Fields)
	if err != nil {
		return TemplatePolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplatePolicys reads every stored entity of this type, sorted.
func (r *Repo) ListTemplatePolicys() ([]TemplatePolicy, error) {
	recs := r.store.List("template/templatepolicy/")
	out := make([]TemplatePolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplatePolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplatePolicys(out)
	return out, nil
}

// PutTemplateLink writes x, creating it when revision is zero.
func (r *Repo) PutTemplateLink(x TemplateLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatelink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatelink.written", 1)
	return rev, nil
}

// GetTemplateLink reads the entity with the given id.
func (r *Repo) GetTemplateLink(id string) (TemplateLink, int64, error) {
	rec, err := r.store.Get("template/templatelink/" + id)
	if err != nil {
		return TemplateLink{}, 0, err
	}
	x, err := ParseTemplateLink(rec.Fields)
	if err != nil {
		return TemplateLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateLinks reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateLinks() ([]TemplateLink, error) {
	recs := r.store.List("template/templatelink/")
	out := make([]TemplateLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateLinks(out)
	return out, nil
}

// PutTemplateNote writes x, creating it when revision is zero.
func (r *Repo) PutTemplateNote(x TemplateNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatenote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatenote.written", 1)
	return rev, nil
}

// GetTemplateNote reads the entity with the given id.
func (r *Repo) GetTemplateNote(id string) (TemplateNote, int64, error) {
	rec, err := r.store.Get("template/templatenote/" + id)
	if err != nil {
		return TemplateNote{}, 0, err
	}
	x, err := ParseTemplateNote(rec.Fields)
	if err != nil {
		return TemplateNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateNotes reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateNotes() ([]TemplateNote, error) {
	recs := r.store.List("template/templatenote/")
	out := make([]TemplateNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateNotes(out)
	return out, nil
}

// PutTemplateWindow writes x, creating it when revision is zero.
func (r *Repo) PutTemplateWindow(x TemplateWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatewindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatewindow.written", 1)
	return rev, nil
}

// GetTemplateWindow reads the entity with the given id.
func (r *Repo) GetTemplateWindow(id string) (TemplateWindow, int64, error) {
	rec, err := r.store.Get("template/templatewindow/" + id)
	if err != nil {
		return TemplateWindow{}, 0, err
	}
	x, err := ParseTemplateWindow(rec.Fields)
	if err != nil {
		return TemplateWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateWindows reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateWindows() ([]TemplateWindow, error) {
	recs := r.store.List("template/templatewindow/")
	out := make([]TemplateWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateWindows(out)
	return out, nil
}

// PutTemplateQuota writes x, creating it when revision is zero.
func (r *Repo) PutTemplateQuota(x TemplateQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatequota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatequota.written", 1)
	return rev, nil
}

// GetTemplateQuota reads the entity with the given id.
func (r *Repo) GetTemplateQuota(id string) (TemplateQuota, int64, error) {
	rec, err := r.store.Get("template/templatequota/" + id)
	if err != nil {
		return TemplateQuota{}, 0, err
	}
	x, err := ParseTemplateQuota(rec.Fields)
	if err != nil {
		return TemplateQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateQuotas() ([]TemplateQuota, error) {
	recs := r.store.List("template/templatequota/")
	out := make([]TemplateQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateQuotas(out)
	return out, nil
}

// PutTemplateBucket writes x, creating it when revision is zero.
func (r *Repo) PutTemplateBucket(x TemplateBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatebucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatebucket.written", 1)
	return rev, nil
}

// GetTemplateBucket reads the entity with the given id.
func (r *Repo) GetTemplateBucket(id string) (TemplateBucket, int64, error) {
	rec, err := r.store.Get("template/templatebucket/" + id)
	if err != nil {
		return TemplateBucket{}, 0, err
	}
	x, err := ParseTemplateBucket(rec.Fields)
	if err != nil {
		return TemplateBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateBuckets() ([]TemplateBucket, error) {
	recs := r.store.List("template/templatebucket/")
	out := make([]TemplateBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateBuckets(out)
	return out, nil
}

// PutTemplateLane writes x, creating it when revision is zero.
func (r *Repo) PutTemplateLane(x TemplateLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatelane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatelane.written", 1)
	return rev, nil
}

// GetTemplateLane reads the entity with the given id.
func (r *Repo) GetTemplateLane(id string) (TemplateLane, int64, error) {
	rec, err := r.store.Get("template/templatelane/" + id)
	if err != nil {
		return TemplateLane{}, 0, err
	}
	x, err := ParseTemplateLane(rec.Fields)
	if err != nil {
		return TemplateLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateLanes reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateLanes() ([]TemplateLane, error) {
	recs := r.store.List("template/templatelane/")
	out := make([]TemplateLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateLanes(out)
	return out, nil
}

// PutTemplateSlot writes x, creating it when revision is zero.
func (r *Repo) PutTemplateSlot(x TemplateSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templateslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templateslot.written", 1)
	return rev, nil
}

// GetTemplateSlot reads the entity with the given id.
func (r *Repo) GetTemplateSlot(id string) (TemplateSlot, int64, error) {
	rec, err := r.store.Get("template/templateslot/" + id)
	if err != nil {
		return TemplateSlot{}, 0, err
	}
	x, err := ParseTemplateSlot(rec.Fields)
	if err != nil {
		return TemplateSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateSlots reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateSlots() ([]TemplateSlot, error) {
	recs := r.store.List("template/templateslot/")
	out := make([]TemplateSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateSlots(out)
	return out, nil
}

// PutTemplateTag writes x, creating it when revision is zero.
func (r *Repo) PutTemplateTag(x TemplateTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatetag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatetag.written", 1)
	return rev, nil
}

// GetTemplateTag reads the entity with the given id.
func (r *Repo) GetTemplateTag(id string) (TemplateTag, int64, error) {
	rec, err := r.store.Get("template/templatetag/" + id)
	if err != nil {
		return TemplateTag{}, 0, err
	}
	x, err := ParseTemplateTag(rec.Fields)
	if err != nil {
		return TemplateTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateTags reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateTags() ([]TemplateTag, error) {
	recs := r.store.List("template/templatetag/")
	out := make([]TemplateTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateTags(out)
	return out, nil
}

// PutTemplateRule writes x, creating it when revision is zero.
func (r *Repo) PutTemplateRule(x TemplateRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templaterule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templaterule.written", 1)
	return rev, nil
}

// GetTemplateRule reads the entity with the given id.
func (r *Repo) GetTemplateRule(id string) (TemplateRule, int64, error) {
	rec, err := r.store.Get("template/templaterule/" + id)
	if err != nil {
		return TemplateRule{}, 0, err
	}
	x, err := ParseTemplateRule(rec.Fields)
	if err != nil {
		return TemplateRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateRules reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateRules() ([]TemplateRule, error) {
	recs := r.store.List("template/templaterule/")
	out := make([]TemplateRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateRules(out)
	return out, nil
}

// PutTemplateHold writes x, creating it when revision is zero.
func (r *Repo) PutTemplateHold(x TemplateHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatehold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatehold.written", 1)
	return rev, nil
}

// GetTemplateHold reads the entity with the given id.
func (r *Repo) GetTemplateHold(id string) (TemplateHold, int64, error) {
	rec, err := r.store.Get("template/templatehold/" + id)
	if err != nil {
		return TemplateHold{}, 0, err
	}
	x, err := ParseTemplateHold(rec.Fields)
	if err != nil {
		return TemplateHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateHolds reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateHolds() ([]TemplateHold, error) {
	recs := r.store.List("template/templatehold/")
	out := make([]TemplateHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateHolds(out)
	return out, nil
}

// PutTemplateClaim writes x, creating it when revision is zero.
func (r *Repo) PutTemplateClaim(x TemplateClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templateclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templateclaim.written", 1)
	return rev, nil
}

// GetTemplateClaim reads the entity with the given id.
func (r *Repo) GetTemplateClaim(id string) (TemplateClaim, int64, error) {
	rec, err := r.store.Get("template/templateclaim/" + id)
	if err != nil {
		return TemplateClaim{}, 0, err
	}
	x, err := ParseTemplateClaim(rec.Fields)
	if err != nil {
		return TemplateClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateClaims reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateClaims() ([]TemplateClaim, error) {
	recs := r.store.List("template/templateclaim/")
	out := make([]TemplateClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateClaims(out)
	return out, nil
}

// PutTemplateCursor writes x, creating it when revision is zero.
func (r *Repo) PutTemplateCursor(x TemplateCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("template.templatecursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("template write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("template.templatecursor.written", 1)
	return rev, nil
}

// GetTemplateCursor reads the entity with the given id.
func (r *Repo) GetTemplateCursor(id string) (TemplateCursor, int64, error) {
	rec, err := r.store.Get("template/templatecursor/" + id)
	if err != nil {
		return TemplateCursor{}, 0, err
	}
	x, err := ParseTemplateCursor(rec.Fields)
	if err != nil {
		return TemplateCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListTemplateCursors reads every stored entity of this type, sorted.
func (r *Repo) ListTemplateCursors() ([]TemplateCursor, error) {
	recs := r.store.List("template/templatecursor/")
	out := make([]TemplateCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseTemplateCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("template: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortTemplateCursors(out)
	return out, nil
}
