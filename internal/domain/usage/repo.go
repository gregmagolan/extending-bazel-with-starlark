package usage

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists usage entities into a store.Store.
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
	l.Debug("usage repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many usage records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutUsageRecord writes x, creating it when revision is zero.
func (r *Repo) PutUsageRecord(x UsageRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagerecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagerecord.written", 1)
	return rev, nil
}

// GetUsageRecord reads the entity with the given id.
func (r *Repo) GetUsageRecord(id string) (UsageRecord, int64, error) {
	rec, err := r.store.Get("usage/usagerecord/" + id)
	if err != nil {
		return UsageRecord{}, 0, err
	}
	x, err := ParseUsageRecord(rec.Fields)
	if err != nil {
		return UsageRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageRecords reads every stored entity of this type, sorted.
func (r *Repo) ListUsageRecords() ([]UsageRecord, error) {
	recs := r.store.List("usage/usagerecord/")
	out := make([]UsageRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageRecords(out)
	return out, nil
}

// PutUsageEntry writes x, creating it when revision is zero.
func (r *Repo) PutUsageEntry(x UsageEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usageentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usageentry.written", 1)
	return rev, nil
}

// GetUsageEntry reads the entity with the given id.
func (r *Repo) GetUsageEntry(id string) (UsageEntry, int64, error) {
	rec, err := r.store.Get("usage/usageentry/" + id)
	if err != nil {
		return UsageEntry{}, 0, err
	}
	x, err := ParseUsageEntry(rec.Fields)
	if err != nil {
		return UsageEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListUsageEntrys() ([]UsageEntry, error) {
	recs := r.store.List("usage/usageentry/")
	out := make([]UsageEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageEntrys(out)
	return out, nil
}

// PutUsageItem writes x, creating it when revision is zero.
func (r *Repo) PutUsageItem(x UsageItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usageitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usageitem.written", 1)
	return rev, nil
}

// GetUsageItem reads the entity with the given id.
func (r *Repo) GetUsageItem(id string) (UsageItem, int64, error) {
	rec, err := r.store.Get("usage/usageitem/" + id)
	if err != nil {
		return UsageItem{}, 0, err
	}
	x, err := ParseUsageItem(rec.Fields)
	if err != nil {
		return UsageItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageItems reads every stored entity of this type, sorted.
func (r *Repo) ListUsageItems() ([]UsageItem, error) {
	recs := r.store.List("usage/usageitem/")
	out := make([]UsageItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageItems(out)
	return out, nil
}

// PutUsageDraft writes x, creating it when revision is zero.
func (r *Repo) PutUsageDraft(x UsageDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagedraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagedraft.written", 1)
	return rev, nil
}

// GetUsageDraft reads the entity with the given id.
func (r *Repo) GetUsageDraft(id string) (UsageDraft, int64, error) {
	rec, err := r.store.Get("usage/usagedraft/" + id)
	if err != nil {
		return UsageDraft{}, 0, err
	}
	x, err := ParseUsageDraft(rec.Fields)
	if err != nil {
		return UsageDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListUsageDrafts() ([]UsageDraft, error) {
	recs := r.store.List("usage/usagedraft/")
	out := make([]UsageDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageDrafts(out)
	return out, nil
}

// PutUsageRevision writes x, creating it when revision is zero.
func (r *Repo) PutUsageRevision(x UsageRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagerevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagerevision.written", 1)
	return rev, nil
}

// GetUsageRevision reads the entity with the given id.
func (r *Repo) GetUsageRevision(id string) (UsageRevision, int64, error) {
	rec, err := r.store.Get("usage/usagerevision/" + id)
	if err != nil {
		return UsageRevision{}, 0, err
	}
	x, err := ParseUsageRevision(rec.Fields)
	if err != nil {
		return UsageRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListUsageRevisions() ([]UsageRevision, error) {
	recs := r.store.List("usage/usagerevision/")
	out := make([]UsageRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageRevisions(out)
	return out, nil
}

// PutUsageSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutUsageSnapshot(x UsageSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagesnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagesnapshot.written", 1)
	return rev, nil
}

// GetUsageSnapshot reads the entity with the given id.
func (r *Repo) GetUsageSnapshot(id string) (UsageSnapshot, int64, error) {
	rec, err := r.store.Get("usage/usagesnapshot/" + id)
	if err != nil {
		return UsageSnapshot{}, 0, err
	}
	x, err := ParseUsageSnapshot(rec.Fields)
	if err != nil {
		return UsageSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListUsageSnapshots() ([]UsageSnapshot, error) {
	recs := r.store.List("usage/usagesnapshot/")
	out := make([]UsageSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageSnapshots(out)
	return out, nil
}

// PutUsageBatch writes x, creating it when revision is zero.
func (r *Repo) PutUsageBatch(x UsageBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagebatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagebatch.written", 1)
	return rev, nil
}

// GetUsageBatch reads the entity with the given id.
func (r *Repo) GetUsageBatch(id string) (UsageBatch, int64, error) {
	rec, err := r.store.Get("usage/usagebatch/" + id)
	if err != nil {
		return UsageBatch{}, 0, err
	}
	x, err := ParseUsageBatch(rec.Fields)
	if err != nil {
		return UsageBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListUsageBatchs() ([]UsageBatch, error) {
	recs := r.store.List("usage/usagebatch/")
	out := make([]UsageBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageBatchs(out)
	return out, nil
}

// PutUsageRequest writes x, creating it when revision is zero.
func (r *Repo) PutUsageRequest(x UsageRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagerequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagerequest.written", 1)
	return rev, nil
}

// GetUsageRequest reads the entity with the given id.
func (r *Repo) GetUsageRequest(id string) (UsageRequest, int64, error) {
	rec, err := r.store.Get("usage/usagerequest/" + id)
	if err != nil {
		return UsageRequest{}, 0, err
	}
	x, err := ParseUsageRequest(rec.Fields)
	if err != nil {
		return UsageRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageRequests reads every stored entity of this type, sorted.
func (r *Repo) ListUsageRequests() ([]UsageRequest, error) {
	recs := r.store.List("usage/usagerequest/")
	out := make([]UsageRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageRequests(out)
	return out, nil
}

// PutUsageSummary writes x, creating it when revision is zero.
func (r *Repo) PutUsageSummary(x UsageSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagesummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagesummary.written", 1)
	return rev, nil
}

// GetUsageSummary reads the entity with the given id.
func (r *Repo) GetUsageSummary(id string) (UsageSummary, int64, error) {
	rec, err := r.store.Get("usage/usagesummary/" + id)
	if err != nil {
		return UsageSummary{}, 0, err
	}
	x, err := ParseUsageSummary(rec.Fields)
	if err != nil {
		return UsageSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListUsageSummarys() ([]UsageSummary, error) {
	recs := r.store.List("usage/usagesummary/")
	out := make([]UsageSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageSummarys(out)
	return out, nil
}

// PutUsagePolicy writes x, creating it when revision is zero.
func (r *Repo) PutUsagePolicy(x UsagePolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagepolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagepolicy.written", 1)
	return rev, nil
}

// GetUsagePolicy reads the entity with the given id.
func (r *Repo) GetUsagePolicy(id string) (UsagePolicy, int64, error) {
	rec, err := r.store.Get("usage/usagepolicy/" + id)
	if err != nil {
		return UsagePolicy{}, 0, err
	}
	x, err := ParseUsagePolicy(rec.Fields)
	if err != nil {
		return UsagePolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsagePolicys reads every stored entity of this type, sorted.
func (r *Repo) ListUsagePolicys() ([]UsagePolicy, error) {
	recs := r.store.List("usage/usagepolicy/")
	out := make([]UsagePolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsagePolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsagePolicys(out)
	return out, nil
}

// PutUsageLink writes x, creating it when revision is zero.
func (r *Repo) PutUsageLink(x UsageLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagelink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagelink.written", 1)
	return rev, nil
}

// GetUsageLink reads the entity with the given id.
func (r *Repo) GetUsageLink(id string) (UsageLink, int64, error) {
	rec, err := r.store.Get("usage/usagelink/" + id)
	if err != nil {
		return UsageLink{}, 0, err
	}
	x, err := ParseUsageLink(rec.Fields)
	if err != nil {
		return UsageLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageLinks reads every stored entity of this type, sorted.
func (r *Repo) ListUsageLinks() ([]UsageLink, error) {
	recs := r.store.List("usage/usagelink/")
	out := make([]UsageLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageLinks(out)
	return out, nil
}

// PutUsageNote writes x, creating it when revision is zero.
func (r *Repo) PutUsageNote(x UsageNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagenote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagenote.written", 1)
	return rev, nil
}

// GetUsageNote reads the entity with the given id.
func (r *Repo) GetUsageNote(id string) (UsageNote, int64, error) {
	rec, err := r.store.Get("usage/usagenote/" + id)
	if err != nil {
		return UsageNote{}, 0, err
	}
	x, err := ParseUsageNote(rec.Fields)
	if err != nil {
		return UsageNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageNotes reads every stored entity of this type, sorted.
func (r *Repo) ListUsageNotes() ([]UsageNote, error) {
	recs := r.store.List("usage/usagenote/")
	out := make([]UsageNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageNotes(out)
	return out, nil
}

// PutUsageWindow writes x, creating it when revision is zero.
func (r *Repo) PutUsageWindow(x UsageWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagewindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagewindow.written", 1)
	return rev, nil
}

// GetUsageWindow reads the entity with the given id.
func (r *Repo) GetUsageWindow(id string) (UsageWindow, int64, error) {
	rec, err := r.store.Get("usage/usagewindow/" + id)
	if err != nil {
		return UsageWindow{}, 0, err
	}
	x, err := ParseUsageWindow(rec.Fields)
	if err != nil {
		return UsageWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageWindows reads every stored entity of this type, sorted.
func (r *Repo) ListUsageWindows() ([]UsageWindow, error) {
	recs := r.store.List("usage/usagewindow/")
	out := make([]UsageWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageWindows(out)
	return out, nil
}

// PutUsageQuota writes x, creating it when revision is zero.
func (r *Repo) PutUsageQuota(x UsageQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagequota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagequota.written", 1)
	return rev, nil
}

// GetUsageQuota reads the entity with the given id.
func (r *Repo) GetUsageQuota(id string) (UsageQuota, int64, error) {
	rec, err := r.store.Get("usage/usagequota/" + id)
	if err != nil {
		return UsageQuota{}, 0, err
	}
	x, err := ParseUsageQuota(rec.Fields)
	if err != nil {
		return UsageQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListUsageQuotas() ([]UsageQuota, error) {
	recs := r.store.List("usage/usagequota/")
	out := make([]UsageQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageQuotas(out)
	return out, nil
}

// PutUsageBucket writes x, creating it when revision is zero.
func (r *Repo) PutUsageBucket(x UsageBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagebucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagebucket.written", 1)
	return rev, nil
}

// GetUsageBucket reads the entity with the given id.
func (r *Repo) GetUsageBucket(id string) (UsageBucket, int64, error) {
	rec, err := r.store.Get("usage/usagebucket/" + id)
	if err != nil {
		return UsageBucket{}, 0, err
	}
	x, err := ParseUsageBucket(rec.Fields)
	if err != nil {
		return UsageBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListUsageBuckets() ([]UsageBucket, error) {
	recs := r.store.List("usage/usagebucket/")
	out := make([]UsageBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageBuckets(out)
	return out, nil
}

// PutUsageLane writes x, creating it when revision is zero.
func (r *Repo) PutUsageLane(x UsageLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagelane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagelane.written", 1)
	return rev, nil
}

// GetUsageLane reads the entity with the given id.
func (r *Repo) GetUsageLane(id string) (UsageLane, int64, error) {
	rec, err := r.store.Get("usage/usagelane/" + id)
	if err != nil {
		return UsageLane{}, 0, err
	}
	x, err := ParseUsageLane(rec.Fields)
	if err != nil {
		return UsageLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageLanes reads every stored entity of this type, sorted.
func (r *Repo) ListUsageLanes() ([]UsageLane, error) {
	recs := r.store.List("usage/usagelane/")
	out := make([]UsageLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageLanes(out)
	return out, nil
}

// PutUsageSlot writes x, creating it when revision is zero.
func (r *Repo) PutUsageSlot(x UsageSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usageslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usageslot.written", 1)
	return rev, nil
}

// GetUsageSlot reads the entity with the given id.
func (r *Repo) GetUsageSlot(id string) (UsageSlot, int64, error) {
	rec, err := r.store.Get("usage/usageslot/" + id)
	if err != nil {
		return UsageSlot{}, 0, err
	}
	x, err := ParseUsageSlot(rec.Fields)
	if err != nil {
		return UsageSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageSlots reads every stored entity of this type, sorted.
func (r *Repo) ListUsageSlots() ([]UsageSlot, error) {
	recs := r.store.List("usage/usageslot/")
	out := make([]UsageSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageSlots(out)
	return out, nil
}

// PutUsageTag writes x, creating it when revision is zero.
func (r *Repo) PutUsageTag(x UsageTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagetag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagetag.written", 1)
	return rev, nil
}

// GetUsageTag reads the entity with the given id.
func (r *Repo) GetUsageTag(id string) (UsageTag, int64, error) {
	rec, err := r.store.Get("usage/usagetag/" + id)
	if err != nil {
		return UsageTag{}, 0, err
	}
	x, err := ParseUsageTag(rec.Fields)
	if err != nil {
		return UsageTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageTags reads every stored entity of this type, sorted.
func (r *Repo) ListUsageTags() ([]UsageTag, error) {
	recs := r.store.List("usage/usagetag/")
	out := make([]UsageTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageTags(out)
	return out, nil
}

// PutUsageRule writes x, creating it when revision is zero.
func (r *Repo) PutUsageRule(x UsageRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagerule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagerule.written", 1)
	return rev, nil
}

// GetUsageRule reads the entity with the given id.
func (r *Repo) GetUsageRule(id string) (UsageRule, int64, error) {
	rec, err := r.store.Get("usage/usagerule/" + id)
	if err != nil {
		return UsageRule{}, 0, err
	}
	x, err := ParseUsageRule(rec.Fields)
	if err != nil {
		return UsageRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageRules reads every stored entity of this type, sorted.
func (r *Repo) ListUsageRules() ([]UsageRule, error) {
	recs := r.store.List("usage/usagerule/")
	out := make([]UsageRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageRules(out)
	return out, nil
}

// PutUsageHold writes x, creating it when revision is zero.
func (r *Repo) PutUsageHold(x UsageHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagehold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagehold.written", 1)
	return rev, nil
}

// GetUsageHold reads the entity with the given id.
func (r *Repo) GetUsageHold(id string) (UsageHold, int64, error) {
	rec, err := r.store.Get("usage/usagehold/" + id)
	if err != nil {
		return UsageHold{}, 0, err
	}
	x, err := ParseUsageHold(rec.Fields)
	if err != nil {
		return UsageHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageHolds reads every stored entity of this type, sorted.
func (r *Repo) ListUsageHolds() ([]UsageHold, error) {
	recs := r.store.List("usage/usagehold/")
	out := make([]UsageHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageHolds(out)
	return out, nil
}

// PutUsageClaim writes x, creating it when revision is zero.
func (r *Repo) PutUsageClaim(x UsageClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usageclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usageclaim.written", 1)
	return rev, nil
}

// GetUsageClaim reads the entity with the given id.
func (r *Repo) GetUsageClaim(id string) (UsageClaim, int64, error) {
	rec, err := r.store.Get("usage/usageclaim/" + id)
	if err != nil {
		return UsageClaim{}, 0, err
	}
	x, err := ParseUsageClaim(rec.Fields)
	if err != nil {
		return UsageClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageClaims reads every stored entity of this type, sorted.
func (r *Repo) ListUsageClaims() ([]UsageClaim, error) {
	recs := r.store.List("usage/usageclaim/")
	out := make([]UsageClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageClaims(out)
	return out, nil
}

// PutUsageCursor writes x, creating it when revision is zero.
func (r *Repo) PutUsageCursor(x UsageCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("usage.usagecursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("usage write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("usage.usagecursor.written", 1)
	return rev, nil
}

// GetUsageCursor reads the entity with the given id.
func (r *Repo) GetUsageCursor(id string) (UsageCursor, int64, error) {
	rec, err := r.store.Get("usage/usagecursor/" + id)
	if err != nil {
		return UsageCursor{}, 0, err
	}
	x, err := ParseUsageCursor(rec.Fields)
	if err != nil {
		return UsageCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListUsageCursors reads every stored entity of this type, sorted.
func (r *Repo) ListUsageCursors() ([]UsageCursor, error) {
	recs := r.store.List("usage/usagecursor/")
	out := make([]UsageCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseUsageCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("usage: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortUsageCursors(out)
	return out, nil
}
