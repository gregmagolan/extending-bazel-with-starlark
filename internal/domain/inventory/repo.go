package inventory

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists inventory entities into a store.Store.
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
	l.Debug("inventory repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many inventory records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutInventoryRecord writes x, creating it when revision is zero.
func (r *Repo) PutInventoryRecord(x InventoryRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryrecord.written", 1)
	return rev, nil
}

// GetInventoryRecord reads the entity with the given id.
func (r *Repo) GetInventoryRecord(id string) (InventoryRecord, int64, error) {
	rec, err := r.store.Get("inventory/inventoryrecord/" + id)
	if err != nil {
		return InventoryRecord{}, 0, err
	}
	x, err := ParseInventoryRecord(rec.Fields)
	if err != nil {
		return InventoryRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryRecords reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryRecords() ([]InventoryRecord, error) {
	recs := r.store.List("inventory/inventoryrecord/")
	out := make([]InventoryRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryRecords(out)
	return out, nil
}

// PutInventoryEntry writes x, creating it when revision is zero.
func (r *Repo) PutInventoryEntry(x InventoryEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryentry.written", 1)
	return rev, nil
}

// GetInventoryEntry reads the entity with the given id.
func (r *Repo) GetInventoryEntry(id string) (InventoryEntry, int64, error) {
	rec, err := r.store.Get("inventory/inventoryentry/" + id)
	if err != nil {
		return InventoryEntry{}, 0, err
	}
	x, err := ParseInventoryEntry(rec.Fields)
	if err != nil {
		return InventoryEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryEntrys() ([]InventoryEntry, error) {
	recs := r.store.List("inventory/inventoryentry/")
	out := make([]InventoryEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryEntrys(out)
	return out, nil
}

// PutInventoryItem writes x, creating it when revision is zero.
func (r *Repo) PutInventoryItem(x InventoryItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryitem.written", 1)
	return rev, nil
}

// GetInventoryItem reads the entity with the given id.
func (r *Repo) GetInventoryItem(id string) (InventoryItem, int64, error) {
	rec, err := r.store.Get("inventory/inventoryitem/" + id)
	if err != nil {
		return InventoryItem{}, 0, err
	}
	x, err := ParseInventoryItem(rec.Fields)
	if err != nil {
		return InventoryItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryItems reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryItems() ([]InventoryItem, error) {
	recs := r.store.List("inventory/inventoryitem/")
	out := make([]InventoryItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryItems(out)
	return out, nil
}

// PutInventoryDraft writes x, creating it when revision is zero.
func (r *Repo) PutInventoryDraft(x InventoryDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorydraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorydraft.written", 1)
	return rev, nil
}

// GetInventoryDraft reads the entity with the given id.
func (r *Repo) GetInventoryDraft(id string) (InventoryDraft, int64, error) {
	rec, err := r.store.Get("inventory/inventorydraft/" + id)
	if err != nil {
		return InventoryDraft{}, 0, err
	}
	x, err := ParseInventoryDraft(rec.Fields)
	if err != nil {
		return InventoryDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryDrafts() ([]InventoryDraft, error) {
	recs := r.store.List("inventory/inventorydraft/")
	out := make([]InventoryDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryDrafts(out)
	return out, nil
}

// PutInventoryRevision writes x, creating it when revision is zero.
func (r *Repo) PutInventoryRevision(x InventoryRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryrevision.written", 1)
	return rev, nil
}

// GetInventoryRevision reads the entity with the given id.
func (r *Repo) GetInventoryRevision(id string) (InventoryRevision, int64, error) {
	rec, err := r.store.Get("inventory/inventoryrevision/" + id)
	if err != nil {
		return InventoryRevision{}, 0, err
	}
	x, err := ParseInventoryRevision(rec.Fields)
	if err != nil {
		return InventoryRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryRevisions() ([]InventoryRevision, error) {
	recs := r.store.List("inventory/inventoryrevision/")
	out := make([]InventoryRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryRevisions(out)
	return out, nil
}

// PutInventorySnapshot writes x, creating it when revision is zero.
func (r *Repo) PutInventorySnapshot(x InventorySnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorysnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorysnapshot.written", 1)
	return rev, nil
}

// GetInventorySnapshot reads the entity with the given id.
func (r *Repo) GetInventorySnapshot(id string) (InventorySnapshot, int64, error) {
	rec, err := r.store.Get("inventory/inventorysnapshot/" + id)
	if err != nil {
		return InventorySnapshot{}, 0, err
	}
	x, err := ParseInventorySnapshot(rec.Fields)
	if err != nil {
		return InventorySnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventorySnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListInventorySnapshots() ([]InventorySnapshot, error) {
	recs := r.store.List("inventory/inventorysnapshot/")
	out := make([]InventorySnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventorySnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventorySnapshots(out)
	return out, nil
}

// PutInventoryBatch writes x, creating it when revision is zero.
func (r *Repo) PutInventoryBatch(x InventoryBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorybatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorybatch.written", 1)
	return rev, nil
}

// GetInventoryBatch reads the entity with the given id.
func (r *Repo) GetInventoryBatch(id string) (InventoryBatch, int64, error) {
	rec, err := r.store.Get("inventory/inventorybatch/" + id)
	if err != nil {
		return InventoryBatch{}, 0, err
	}
	x, err := ParseInventoryBatch(rec.Fields)
	if err != nil {
		return InventoryBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryBatchs() ([]InventoryBatch, error) {
	recs := r.store.List("inventory/inventorybatch/")
	out := make([]InventoryBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryBatchs(out)
	return out, nil
}

// PutInventoryRequest writes x, creating it when revision is zero.
func (r *Repo) PutInventoryRequest(x InventoryRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryrequest.written", 1)
	return rev, nil
}

// GetInventoryRequest reads the entity with the given id.
func (r *Repo) GetInventoryRequest(id string) (InventoryRequest, int64, error) {
	rec, err := r.store.Get("inventory/inventoryrequest/" + id)
	if err != nil {
		return InventoryRequest{}, 0, err
	}
	x, err := ParseInventoryRequest(rec.Fields)
	if err != nil {
		return InventoryRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryRequests reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryRequests() ([]InventoryRequest, error) {
	recs := r.store.List("inventory/inventoryrequest/")
	out := make([]InventoryRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryRequests(out)
	return out, nil
}

// PutInventorySummary writes x, creating it when revision is zero.
func (r *Repo) PutInventorySummary(x InventorySummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorysummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorysummary.written", 1)
	return rev, nil
}

// GetInventorySummary reads the entity with the given id.
func (r *Repo) GetInventorySummary(id string) (InventorySummary, int64, error) {
	rec, err := r.store.Get("inventory/inventorysummary/" + id)
	if err != nil {
		return InventorySummary{}, 0, err
	}
	x, err := ParseInventorySummary(rec.Fields)
	if err != nil {
		return InventorySummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventorySummarys reads every stored entity of this type, sorted.
func (r *Repo) ListInventorySummarys() ([]InventorySummary, error) {
	recs := r.store.List("inventory/inventorysummary/")
	out := make([]InventorySummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventorySummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventorySummarys(out)
	return out, nil
}

// PutInventoryPolicy writes x, creating it when revision is zero.
func (r *Repo) PutInventoryPolicy(x InventoryPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorypolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorypolicy.written", 1)
	return rev, nil
}

// GetInventoryPolicy reads the entity with the given id.
func (r *Repo) GetInventoryPolicy(id string) (InventoryPolicy, int64, error) {
	rec, err := r.store.Get("inventory/inventorypolicy/" + id)
	if err != nil {
		return InventoryPolicy{}, 0, err
	}
	x, err := ParseInventoryPolicy(rec.Fields)
	if err != nil {
		return InventoryPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryPolicys() ([]InventoryPolicy, error) {
	recs := r.store.List("inventory/inventorypolicy/")
	out := make([]InventoryPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryPolicys(out)
	return out, nil
}

// PutInventoryLink writes x, creating it when revision is zero.
func (r *Repo) PutInventoryLink(x InventoryLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorylink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorylink.written", 1)
	return rev, nil
}

// GetInventoryLink reads the entity with the given id.
func (r *Repo) GetInventoryLink(id string) (InventoryLink, int64, error) {
	rec, err := r.store.Get("inventory/inventorylink/" + id)
	if err != nil {
		return InventoryLink{}, 0, err
	}
	x, err := ParseInventoryLink(rec.Fields)
	if err != nil {
		return InventoryLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryLinks reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryLinks() ([]InventoryLink, error) {
	recs := r.store.List("inventory/inventorylink/")
	out := make([]InventoryLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryLinks(out)
	return out, nil
}

// PutInventoryNote writes x, creating it when revision is zero.
func (r *Repo) PutInventoryNote(x InventoryNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorynote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorynote.written", 1)
	return rev, nil
}

// GetInventoryNote reads the entity with the given id.
func (r *Repo) GetInventoryNote(id string) (InventoryNote, int64, error) {
	rec, err := r.store.Get("inventory/inventorynote/" + id)
	if err != nil {
		return InventoryNote{}, 0, err
	}
	x, err := ParseInventoryNote(rec.Fields)
	if err != nil {
		return InventoryNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryNotes reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryNotes() ([]InventoryNote, error) {
	recs := r.store.List("inventory/inventorynote/")
	out := make([]InventoryNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryNotes(out)
	return out, nil
}

// PutInventoryWindow writes x, creating it when revision is zero.
func (r *Repo) PutInventoryWindow(x InventoryWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorywindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorywindow.written", 1)
	return rev, nil
}

// GetInventoryWindow reads the entity with the given id.
func (r *Repo) GetInventoryWindow(id string) (InventoryWindow, int64, error) {
	rec, err := r.store.Get("inventory/inventorywindow/" + id)
	if err != nil {
		return InventoryWindow{}, 0, err
	}
	x, err := ParseInventoryWindow(rec.Fields)
	if err != nil {
		return InventoryWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryWindows reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryWindows() ([]InventoryWindow, error) {
	recs := r.store.List("inventory/inventorywindow/")
	out := make([]InventoryWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryWindows(out)
	return out, nil
}

// PutInventoryQuota writes x, creating it when revision is zero.
func (r *Repo) PutInventoryQuota(x InventoryQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryquota.written", 1)
	return rev, nil
}

// GetInventoryQuota reads the entity with the given id.
func (r *Repo) GetInventoryQuota(id string) (InventoryQuota, int64, error) {
	rec, err := r.store.Get("inventory/inventoryquota/" + id)
	if err != nil {
		return InventoryQuota{}, 0, err
	}
	x, err := ParseInventoryQuota(rec.Fields)
	if err != nil {
		return InventoryQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryQuotas() ([]InventoryQuota, error) {
	recs := r.store.List("inventory/inventoryquota/")
	out := make([]InventoryQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryQuotas(out)
	return out, nil
}

// PutInventoryBucket writes x, creating it when revision is zero.
func (r *Repo) PutInventoryBucket(x InventoryBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorybucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorybucket.written", 1)
	return rev, nil
}

// GetInventoryBucket reads the entity with the given id.
func (r *Repo) GetInventoryBucket(id string) (InventoryBucket, int64, error) {
	rec, err := r.store.Get("inventory/inventorybucket/" + id)
	if err != nil {
		return InventoryBucket{}, 0, err
	}
	x, err := ParseInventoryBucket(rec.Fields)
	if err != nil {
		return InventoryBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryBuckets() ([]InventoryBucket, error) {
	recs := r.store.List("inventory/inventorybucket/")
	out := make([]InventoryBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryBuckets(out)
	return out, nil
}

// PutInventoryLane writes x, creating it when revision is zero.
func (r *Repo) PutInventoryLane(x InventoryLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorylane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorylane.written", 1)
	return rev, nil
}

// GetInventoryLane reads the entity with the given id.
func (r *Repo) GetInventoryLane(id string) (InventoryLane, int64, error) {
	rec, err := r.store.Get("inventory/inventorylane/" + id)
	if err != nil {
		return InventoryLane{}, 0, err
	}
	x, err := ParseInventoryLane(rec.Fields)
	if err != nil {
		return InventoryLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryLanes reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryLanes() ([]InventoryLane, error) {
	recs := r.store.List("inventory/inventorylane/")
	out := make([]InventoryLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryLanes(out)
	return out, nil
}

// PutInventorySlot writes x, creating it when revision is zero.
func (r *Repo) PutInventorySlot(x InventorySlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryslot.written", 1)
	return rev, nil
}

// GetInventorySlot reads the entity with the given id.
func (r *Repo) GetInventorySlot(id string) (InventorySlot, int64, error) {
	rec, err := r.store.Get("inventory/inventoryslot/" + id)
	if err != nil {
		return InventorySlot{}, 0, err
	}
	x, err := ParseInventorySlot(rec.Fields)
	if err != nil {
		return InventorySlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventorySlots reads every stored entity of this type, sorted.
func (r *Repo) ListInventorySlots() ([]InventorySlot, error) {
	recs := r.store.List("inventory/inventoryslot/")
	out := make([]InventorySlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventorySlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventorySlots(out)
	return out, nil
}

// PutInventoryTag writes x, creating it when revision is zero.
func (r *Repo) PutInventoryTag(x InventoryTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorytag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorytag.written", 1)
	return rev, nil
}

// GetInventoryTag reads the entity with the given id.
func (r *Repo) GetInventoryTag(id string) (InventoryTag, int64, error) {
	rec, err := r.store.Get("inventory/inventorytag/" + id)
	if err != nil {
		return InventoryTag{}, 0, err
	}
	x, err := ParseInventoryTag(rec.Fields)
	if err != nil {
		return InventoryTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryTags reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryTags() ([]InventoryTag, error) {
	recs := r.store.List("inventory/inventorytag/")
	out := make([]InventoryTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryTags(out)
	return out, nil
}

// PutInventoryRule writes x, creating it when revision is zero.
func (r *Repo) PutInventoryRule(x InventoryRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryrule.written", 1)
	return rev, nil
}

// GetInventoryRule reads the entity with the given id.
func (r *Repo) GetInventoryRule(id string) (InventoryRule, int64, error) {
	rec, err := r.store.Get("inventory/inventoryrule/" + id)
	if err != nil {
		return InventoryRule{}, 0, err
	}
	x, err := ParseInventoryRule(rec.Fields)
	if err != nil {
		return InventoryRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryRules reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryRules() ([]InventoryRule, error) {
	recs := r.store.List("inventory/inventoryrule/")
	out := make([]InventoryRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryRules(out)
	return out, nil
}

// PutInventoryHold writes x, creating it when revision is zero.
func (r *Repo) PutInventoryHold(x InventoryHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryhold.written", 1)
	return rev, nil
}

// GetInventoryHold reads the entity with the given id.
func (r *Repo) GetInventoryHold(id string) (InventoryHold, int64, error) {
	rec, err := r.store.Get("inventory/inventoryhold/" + id)
	if err != nil {
		return InventoryHold{}, 0, err
	}
	x, err := ParseInventoryHold(rec.Fields)
	if err != nil {
		return InventoryHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryHolds reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryHolds() ([]InventoryHold, error) {
	recs := r.store.List("inventory/inventoryhold/")
	out := make([]InventoryHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryHolds(out)
	return out, nil
}

// PutInventoryClaim writes x, creating it when revision is zero.
func (r *Repo) PutInventoryClaim(x InventoryClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventoryclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventoryclaim.written", 1)
	return rev, nil
}

// GetInventoryClaim reads the entity with the given id.
func (r *Repo) GetInventoryClaim(id string) (InventoryClaim, int64, error) {
	rec, err := r.store.Get("inventory/inventoryclaim/" + id)
	if err != nil {
		return InventoryClaim{}, 0, err
	}
	x, err := ParseInventoryClaim(rec.Fields)
	if err != nil {
		return InventoryClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryClaims reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryClaims() ([]InventoryClaim, error) {
	recs := r.store.List("inventory/inventoryclaim/")
	out := make([]InventoryClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryClaims(out)
	return out, nil
}

// PutInventoryCursor writes x, creating it when revision is zero.
func (r *Repo) PutInventoryCursor(x InventoryCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("inventory.inventorycursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("inventory write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("inventory.inventorycursor.written", 1)
	return rev, nil
}

// GetInventoryCursor reads the entity with the given id.
func (r *Repo) GetInventoryCursor(id string) (InventoryCursor, int64, error) {
	rec, err := r.store.Get("inventory/inventorycursor/" + id)
	if err != nil {
		return InventoryCursor{}, 0, err
	}
	x, err := ParseInventoryCursor(rec.Fields)
	if err != nil {
		return InventoryCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListInventoryCursors reads every stored entity of this type, sorted.
func (r *Repo) ListInventoryCursors() ([]InventoryCursor, error) {
	recs := r.store.List("inventory/inventorycursor/")
	out := make([]InventoryCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseInventoryCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("inventory: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortInventoryCursors(out)
	return out, nil
}
