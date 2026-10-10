package warehouse

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists warehouse entities into a store.Store.
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
	l.Debug("warehouse repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many warehouse records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutWarehouseRecord writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseRecord(x WarehouseRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouserecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouserecord.written", 1)
	return rev, nil
}

// GetWarehouseRecord reads the entity with the given id.
func (r *Repo) GetWarehouseRecord(id string) (WarehouseRecord, int64, error) {
	rec, err := r.store.Get("warehouse/warehouserecord/" + id)
	if err != nil {
		return WarehouseRecord{}, 0, err
	}
	x, err := ParseWarehouseRecord(rec.Fields)
	if err != nil {
		return WarehouseRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseRecords reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseRecords() ([]WarehouseRecord, error) {
	recs := r.store.List("warehouse/warehouserecord/")
	out := make([]WarehouseRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseRecords(out)
	return out, nil
}

// PutWarehouseEntry writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseEntry(x WarehouseEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouseentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouseentry.written", 1)
	return rev, nil
}

// GetWarehouseEntry reads the entity with the given id.
func (r *Repo) GetWarehouseEntry(id string) (WarehouseEntry, int64, error) {
	rec, err := r.store.Get("warehouse/warehouseentry/" + id)
	if err != nil {
		return WarehouseEntry{}, 0, err
	}
	x, err := ParseWarehouseEntry(rec.Fields)
	if err != nil {
		return WarehouseEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseEntrys() ([]WarehouseEntry, error) {
	recs := r.store.List("warehouse/warehouseentry/")
	out := make([]WarehouseEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseEntrys(out)
	return out, nil
}

// PutWarehouseItem writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseItem(x WarehouseItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouseitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouseitem.written", 1)
	return rev, nil
}

// GetWarehouseItem reads the entity with the given id.
func (r *Repo) GetWarehouseItem(id string) (WarehouseItem, int64, error) {
	rec, err := r.store.Get("warehouse/warehouseitem/" + id)
	if err != nil {
		return WarehouseItem{}, 0, err
	}
	x, err := ParseWarehouseItem(rec.Fields)
	if err != nil {
		return WarehouseItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseItems reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseItems() ([]WarehouseItem, error) {
	recs := r.store.List("warehouse/warehouseitem/")
	out := make([]WarehouseItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseItems(out)
	return out, nil
}

// PutWarehouseDraft writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseDraft(x WarehouseDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousedraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousedraft.written", 1)
	return rev, nil
}

// GetWarehouseDraft reads the entity with the given id.
func (r *Repo) GetWarehouseDraft(id string) (WarehouseDraft, int64, error) {
	rec, err := r.store.Get("warehouse/warehousedraft/" + id)
	if err != nil {
		return WarehouseDraft{}, 0, err
	}
	x, err := ParseWarehouseDraft(rec.Fields)
	if err != nil {
		return WarehouseDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseDrafts() ([]WarehouseDraft, error) {
	recs := r.store.List("warehouse/warehousedraft/")
	out := make([]WarehouseDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseDrafts(out)
	return out, nil
}

// PutWarehouseRevision writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseRevision(x WarehouseRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouserevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouserevision.written", 1)
	return rev, nil
}

// GetWarehouseRevision reads the entity with the given id.
func (r *Repo) GetWarehouseRevision(id string) (WarehouseRevision, int64, error) {
	rec, err := r.store.Get("warehouse/warehouserevision/" + id)
	if err != nil {
		return WarehouseRevision{}, 0, err
	}
	x, err := ParseWarehouseRevision(rec.Fields)
	if err != nil {
		return WarehouseRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseRevisions() ([]WarehouseRevision, error) {
	recs := r.store.List("warehouse/warehouserevision/")
	out := make([]WarehouseRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseRevisions(out)
	return out, nil
}

// PutWarehouseSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseSnapshot(x WarehouseSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousesnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousesnapshot.written", 1)
	return rev, nil
}

// GetWarehouseSnapshot reads the entity with the given id.
func (r *Repo) GetWarehouseSnapshot(id string) (WarehouseSnapshot, int64, error) {
	rec, err := r.store.Get("warehouse/warehousesnapshot/" + id)
	if err != nil {
		return WarehouseSnapshot{}, 0, err
	}
	x, err := ParseWarehouseSnapshot(rec.Fields)
	if err != nil {
		return WarehouseSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseSnapshots() ([]WarehouseSnapshot, error) {
	recs := r.store.List("warehouse/warehousesnapshot/")
	out := make([]WarehouseSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseSnapshots(out)
	return out, nil
}

// PutWarehouseBatch writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseBatch(x WarehouseBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousebatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousebatch.written", 1)
	return rev, nil
}

// GetWarehouseBatch reads the entity with the given id.
func (r *Repo) GetWarehouseBatch(id string) (WarehouseBatch, int64, error) {
	rec, err := r.store.Get("warehouse/warehousebatch/" + id)
	if err != nil {
		return WarehouseBatch{}, 0, err
	}
	x, err := ParseWarehouseBatch(rec.Fields)
	if err != nil {
		return WarehouseBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseBatchs() ([]WarehouseBatch, error) {
	recs := r.store.List("warehouse/warehousebatch/")
	out := make([]WarehouseBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseBatchs(out)
	return out, nil
}

// PutWarehouseRequest writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseRequest(x WarehouseRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouserequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouserequest.written", 1)
	return rev, nil
}

// GetWarehouseRequest reads the entity with the given id.
func (r *Repo) GetWarehouseRequest(id string) (WarehouseRequest, int64, error) {
	rec, err := r.store.Get("warehouse/warehouserequest/" + id)
	if err != nil {
		return WarehouseRequest{}, 0, err
	}
	x, err := ParseWarehouseRequest(rec.Fields)
	if err != nil {
		return WarehouseRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseRequests reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseRequests() ([]WarehouseRequest, error) {
	recs := r.store.List("warehouse/warehouserequest/")
	out := make([]WarehouseRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseRequests(out)
	return out, nil
}

// PutWarehouseSummary writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseSummary(x WarehouseSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousesummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousesummary.written", 1)
	return rev, nil
}

// GetWarehouseSummary reads the entity with the given id.
func (r *Repo) GetWarehouseSummary(id string) (WarehouseSummary, int64, error) {
	rec, err := r.store.Get("warehouse/warehousesummary/" + id)
	if err != nil {
		return WarehouseSummary{}, 0, err
	}
	x, err := ParseWarehouseSummary(rec.Fields)
	if err != nil {
		return WarehouseSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseSummarys() ([]WarehouseSummary, error) {
	recs := r.store.List("warehouse/warehousesummary/")
	out := make([]WarehouseSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseSummarys(out)
	return out, nil
}

// PutWarehousePolicy writes x, creating it when revision is zero.
func (r *Repo) PutWarehousePolicy(x WarehousePolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousepolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousepolicy.written", 1)
	return rev, nil
}

// GetWarehousePolicy reads the entity with the given id.
func (r *Repo) GetWarehousePolicy(id string) (WarehousePolicy, int64, error) {
	rec, err := r.store.Get("warehouse/warehousepolicy/" + id)
	if err != nil {
		return WarehousePolicy{}, 0, err
	}
	x, err := ParseWarehousePolicy(rec.Fields)
	if err != nil {
		return WarehousePolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehousePolicys reads every stored entity of this type, sorted.
func (r *Repo) ListWarehousePolicys() ([]WarehousePolicy, error) {
	recs := r.store.List("warehouse/warehousepolicy/")
	out := make([]WarehousePolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehousePolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehousePolicys(out)
	return out, nil
}

// PutWarehouseLink writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseLink(x WarehouseLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouselink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouselink.written", 1)
	return rev, nil
}

// GetWarehouseLink reads the entity with the given id.
func (r *Repo) GetWarehouseLink(id string) (WarehouseLink, int64, error) {
	rec, err := r.store.Get("warehouse/warehouselink/" + id)
	if err != nil {
		return WarehouseLink{}, 0, err
	}
	x, err := ParseWarehouseLink(rec.Fields)
	if err != nil {
		return WarehouseLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseLinks reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseLinks() ([]WarehouseLink, error) {
	recs := r.store.List("warehouse/warehouselink/")
	out := make([]WarehouseLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseLinks(out)
	return out, nil
}

// PutWarehouseNote writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseNote(x WarehouseNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousenote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousenote.written", 1)
	return rev, nil
}

// GetWarehouseNote reads the entity with the given id.
func (r *Repo) GetWarehouseNote(id string) (WarehouseNote, int64, error) {
	rec, err := r.store.Get("warehouse/warehousenote/" + id)
	if err != nil {
		return WarehouseNote{}, 0, err
	}
	x, err := ParseWarehouseNote(rec.Fields)
	if err != nil {
		return WarehouseNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseNotes reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseNotes() ([]WarehouseNote, error) {
	recs := r.store.List("warehouse/warehousenote/")
	out := make([]WarehouseNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseNotes(out)
	return out, nil
}

// PutWarehouseWindow writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseWindow(x WarehouseWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousewindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousewindow.written", 1)
	return rev, nil
}

// GetWarehouseWindow reads the entity with the given id.
func (r *Repo) GetWarehouseWindow(id string) (WarehouseWindow, int64, error) {
	rec, err := r.store.Get("warehouse/warehousewindow/" + id)
	if err != nil {
		return WarehouseWindow{}, 0, err
	}
	x, err := ParseWarehouseWindow(rec.Fields)
	if err != nil {
		return WarehouseWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseWindows reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseWindows() ([]WarehouseWindow, error) {
	recs := r.store.List("warehouse/warehousewindow/")
	out := make([]WarehouseWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseWindows(out)
	return out, nil
}

// PutWarehouseQuota writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseQuota(x WarehouseQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousequota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousequota.written", 1)
	return rev, nil
}

// GetWarehouseQuota reads the entity with the given id.
func (r *Repo) GetWarehouseQuota(id string) (WarehouseQuota, int64, error) {
	rec, err := r.store.Get("warehouse/warehousequota/" + id)
	if err != nil {
		return WarehouseQuota{}, 0, err
	}
	x, err := ParseWarehouseQuota(rec.Fields)
	if err != nil {
		return WarehouseQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseQuotas() ([]WarehouseQuota, error) {
	recs := r.store.List("warehouse/warehousequota/")
	out := make([]WarehouseQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseQuotas(out)
	return out, nil
}

// PutWarehouseBucket writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseBucket(x WarehouseBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousebucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousebucket.written", 1)
	return rev, nil
}

// GetWarehouseBucket reads the entity with the given id.
func (r *Repo) GetWarehouseBucket(id string) (WarehouseBucket, int64, error) {
	rec, err := r.store.Get("warehouse/warehousebucket/" + id)
	if err != nil {
		return WarehouseBucket{}, 0, err
	}
	x, err := ParseWarehouseBucket(rec.Fields)
	if err != nil {
		return WarehouseBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseBuckets() ([]WarehouseBucket, error) {
	recs := r.store.List("warehouse/warehousebucket/")
	out := make([]WarehouseBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseBuckets(out)
	return out, nil
}

// PutWarehouseLane writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseLane(x WarehouseLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouselane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouselane.written", 1)
	return rev, nil
}

// GetWarehouseLane reads the entity with the given id.
func (r *Repo) GetWarehouseLane(id string) (WarehouseLane, int64, error) {
	rec, err := r.store.Get("warehouse/warehouselane/" + id)
	if err != nil {
		return WarehouseLane{}, 0, err
	}
	x, err := ParseWarehouseLane(rec.Fields)
	if err != nil {
		return WarehouseLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseLanes reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseLanes() ([]WarehouseLane, error) {
	recs := r.store.List("warehouse/warehouselane/")
	out := make([]WarehouseLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseLanes(out)
	return out, nil
}

// PutWarehouseSlot writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseSlot(x WarehouseSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouseslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouseslot.written", 1)
	return rev, nil
}

// GetWarehouseSlot reads the entity with the given id.
func (r *Repo) GetWarehouseSlot(id string) (WarehouseSlot, int64, error) {
	rec, err := r.store.Get("warehouse/warehouseslot/" + id)
	if err != nil {
		return WarehouseSlot{}, 0, err
	}
	x, err := ParseWarehouseSlot(rec.Fields)
	if err != nil {
		return WarehouseSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseSlots reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseSlots() ([]WarehouseSlot, error) {
	recs := r.store.List("warehouse/warehouseslot/")
	out := make([]WarehouseSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseSlots(out)
	return out, nil
}

// PutWarehouseTag writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseTag(x WarehouseTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousetag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousetag.written", 1)
	return rev, nil
}

// GetWarehouseTag reads the entity with the given id.
func (r *Repo) GetWarehouseTag(id string) (WarehouseTag, int64, error) {
	rec, err := r.store.Get("warehouse/warehousetag/" + id)
	if err != nil {
		return WarehouseTag{}, 0, err
	}
	x, err := ParseWarehouseTag(rec.Fields)
	if err != nil {
		return WarehouseTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseTags reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseTags() ([]WarehouseTag, error) {
	recs := r.store.List("warehouse/warehousetag/")
	out := make([]WarehouseTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseTags(out)
	return out, nil
}

// PutWarehouseRule writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseRule(x WarehouseRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouserule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouserule.written", 1)
	return rev, nil
}

// GetWarehouseRule reads the entity with the given id.
func (r *Repo) GetWarehouseRule(id string) (WarehouseRule, int64, error) {
	rec, err := r.store.Get("warehouse/warehouserule/" + id)
	if err != nil {
		return WarehouseRule{}, 0, err
	}
	x, err := ParseWarehouseRule(rec.Fields)
	if err != nil {
		return WarehouseRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseRules reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseRules() ([]WarehouseRule, error) {
	recs := r.store.List("warehouse/warehouserule/")
	out := make([]WarehouseRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseRules(out)
	return out, nil
}

// PutWarehouseHold writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseHold(x WarehouseHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousehold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousehold.written", 1)
	return rev, nil
}

// GetWarehouseHold reads the entity with the given id.
func (r *Repo) GetWarehouseHold(id string) (WarehouseHold, int64, error) {
	rec, err := r.store.Get("warehouse/warehousehold/" + id)
	if err != nil {
		return WarehouseHold{}, 0, err
	}
	x, err := ParseWarehouseHold(rec.Fields)
	if err != nil {
		return WarehouseHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseHolds reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseHolds() ([]WarehouseHold, error) {
	recs := r.store.List("warehouse/warehousehold/")
	out := make([]WarehouseHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseHolds(out)
	return out, nil
}

// PutWarehouseClaim writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseClaim(x WarehouseClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehouseclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehouseclaim.written", 1)
	return rev, nil
}

// GetWarehouseClaim reads the entity with the given id.
func (r *Repo) GetWarehouseClaim(id string) (WarehouseClaim, int64, error) {
	rec, err := r.store.Get("warehouse/warehouseclaim/" + id)
	if err != nil {
		return WarehouseClaim{}, 0, err
	}
	x, err := ParseWarehouseClaim(rec.Fields)
	if err != nil {
		return WarehouseClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseClaims reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseClaims() ([]WarehouseClaim, error) {
	recs := r.store.List("warehouse/warehouseclaim/")
	out := make([]WarehouseClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseClaims(out)
	return out, nil
}

// PutWarehouseCursor writes x, creating it when revision is zero.
func (r *Repo) PutWarehouseCursor(x WarehouseCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("warehouse.warehousecursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("warehouse write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("warehouse.warehousecursor.written", 1)
	return rev, nil
}

// GetWarehouseCursor reads the entity with the given id.
func (r *Repo) GetWarehouseCursor(id string) (WarehouseCursor, int64, error) {
	rec, err := r.store.Get("warehouse/warehousecursor/" + id)
	if err != nil {
		return WarehouseCursor{}, 0, err
	}
	x, err := ParseWarehouseCursor(rec.Fields)
	if err != nil {
		return WarehouseCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListWarehouseCursors reads every stored entity of this type, sorted.
func (r *Repo) ListWarehouseCursors() ([]WarehouseCursor, error) {
	recs := r.store.List("warehouse/warehousecursor/")
	out := make([]WarehouseCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseWarehouseCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("warehouse: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortWarehouseCursors(out)
	return out, nil
}
