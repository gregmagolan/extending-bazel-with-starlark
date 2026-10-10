package device

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists device entities into a store.Store.
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
	l.Debug("device repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many device records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutDeviceRecord writes x, creating it when revision is zero.
func (r *Repo) PutDeviceRecord(x DeviceRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicerecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicerecord.written", 1)
	return rev, nil
}

// GetDeviceRecord reads the entity with the given id.
func (r *Repo) GetDeviceRecord(id string) (DeviceRecord, int64, error) {
	rec, err := r.store.Get("device/devicerecord/" + id)
	if err != nil {
		return DeviceRecord{}, 0, err
	}
	x, err := ParseDeviceRecord(rec.Fields)
	if err != nil {
		return DeviceRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceRecords reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceRecords() ([]DeviceRecord, error) {
	recs := r.store.List("device/devicerecord/")
	out := make([]DeviceRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceRecords(out)
	return out, nil
}

// PutDeviceEntry writes x, creating it when revision is zero.
func (r *Repo) PutDeviceEntry(x DeviceEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.deviceentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.deviceentry.written", 1)
	return rev, nil
}

// GetDeviceEntry reads the entity with the given id.
func (r *Repo) GetDeviceEntry(id string) (DeviceEntry, int64, error) {
	rec, err := r.store.Get("device/deviceentry/" + id)
	if err != nil {
		return DeviceEntry{}, 0, err
	}
	x, err := ParseDeviceEntry(rec.Fields)
	if err != nil {
		return DeviceEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceEntrys() ([]DeviceEntry, error) {
	recs := r.store.List("device/deviceentry/")
	out := make([]DeviceEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceEntrys(out)
	return out, nil
}

// PutDeviceItem writes x, creating it when revision is zero.
func (r *Repo) PutDeviceItem(x DeviceItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.deviceitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.deviceitem.written", 1)
	return rev, nil
}

// GetDeviceItem reads the entity with the given id.
func (r *Repo) GetDeviceItem(id string) (DeviceItem, int64, error) {
	rec, err := r.store.Get("device/deviceitem/" + id)
	if err != nil {
		return DeviceItem{}, 0, err
	}
	x, err := ParseDeviceItem(rec.Fields)
	if err != nil {
		return DeviceItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceItems reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceItems() ([]DeviceItem, error) {
	recs := r.store.List("device/deviceitem/")
	out := make([]DeviceItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceItems(out)
	return out, nil
}

// PutDeviceDraft writes x, creating it when revision is zero.
func (r *Repo) PutDeviceDraft(x DeviceDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicedraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicedraft.written", 1)
	return rev, nil
}

// GetDeviceDraft reads the entity with the given id.
func (r *Repo) GetDeviceDraft(id string) (DeviceDraft, int64, error) {
	rec, err := r.store.Get("device/devicedraft/" + id)
	if err != nil {
		return DeviceDraft{}, 0, err
	}
	x, err := ParseDeviceDraft(rec.Fields)
	if err != nil {
		return DeviceDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceDrafts() ([]DeviceDraft, error) {
	recs := r.store.List("device/devicedraft/")
	out := make([]DeviceDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceDrafts(out)
	return out, nil
}

// PutDeviceRevision writes x, creating it when revision is zero.
func (r *Repo) PutDeviceRevision(x DeviceRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicerevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicerevision.written", 1)
	return rev, nil
}

// GetDeviceRevision reads the entity with the given id.
func (r *Repo) GetDeviceRevision(id string) (DeviceRevision, int64, error) {
	rec, err := r.store.Get("device/devicerevision/" + id)
	if err != nil {
		return DeviceRevision{}, 0, err
	}
	x, err := ParseDeviceRevision(rec.Fields)
	if err != nil {
		return DeviceRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceRevisions() ([]DeviceRevision, error) {
	recs := r.store.List("device/devicerevision/")
	out := make([]DeviceRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceRevisions(out)
	return out, nil
}

// PutDeviceSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutDeviceSnapshot(x DeviceSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicesnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicesnapshot.written", 1)
	return rev, nil
}

// GetDeviceSnapshot reads the entity with the given id.
func (r *Repo) GetDeviceSnapshot(id string) (DeviceSnapshot, int64, error) {
	rec, err := r.store.Get("device/devicesnapshot/" + id)
	if err != nil {
		return DeviceSnapshot{}, 0, err
	}
	x, err := ParseDeviceSnapshot(rec.Fields)
	if err != nil {
		return DeviceSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceSnapshots() ([]DeviceSnapshot, error) {
	recs := r.store.List("device/devicesnapshot/")
	out := make([]DeviceSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceSnapshots(out)
	return out, nil
}

// PutDeviceBatch writes x, creating it when revision is zero.
func (r *Repo) PutDeviceBatch(x DeviceBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicebatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicebatch.written", 1)
	return rev, nil
}

// GetDeviceBatch reads the entity with the given id.
func (r *Repo) GetDeviceBatch(id string) (DeviceBatch, int64, error) {
	rec, err := r.store.Get("device/devicebatch/" + id)
	if err != nil {
		return DeviceBatch{}, 0, err
	}
	x, err := ParseDeviceBatch(rec.Fields)
	if err != nil {
		return DeviceBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceBatchs() ([]DeviceBatch, error) {
	recs := r.store.List("device/devicebatch/")
	out := make([]DeviceBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceBatchs(out)
	return out, nil
}

// PutDeviceRequest writes x, creating it when revision is zero.
func (r *Repo) PutDeviceRequest(x DeviceRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicerequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicerequest.written", 1)
	return rev, nil
}

// GetDeviceRequest reads the entity with the given id.
func (r *Repo) GetDeviceRequest(id string) (DeviceRequest, int64, error) {
	rec, err := r.store.Get("device/devicerequest/" + id)
	if err != nil {
		return DeviceRequest{}, 0, err
	}
	x, err := ParseDeviceRequest(rec.Fields)
	if err != nil {
		return DeviceRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceRequests reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceRequests() ([]DeviceRequest, error) {
	recs := r.store.List("device/devicerequest/")
	out := make([]DeviceRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceRequests(out)
	return out, nil
}

// PutDeviceSummary writes x, creating it when revision is zero.
func (r *Repo) PutDeviceSummary(x DeviceSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicesummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicesummary.written", 1)
	return rev, nil
}

// GetDeviceSummary reads the entity with the given id.
func (r *Repo) GetDeviceSummary(id string) (DeviceSummary, int64, error) {
	rec, err := r.store.Get("device/devicesummary/" + id)
	if err != nil {
		return DeviceSummary{}, 0, err
	}
	x, err := ParseDeviceSummary(rec.Fields)
	if err != nil {
		return DeviceSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceSummarys() ([]DeviceSummary, error) {
	recs := r.store.List("device/devicesummary/")
	out := make([]DeviceSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceSummarys(out)
	return out, nil
}

// PutDevicePolicy writes x, creating it when revision is zero.
func (r *Repo) PutDevicePolicy(x DevicePolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicepolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicepolicy.written", 1)
	return rev, nil
}

// GetDevicePolicy reads the entity with the given id.
func (r *Repo) GetDevicePolicy(id string) (DevicePolicy, int64, error) {
	rec, err := r.store.Get("device/devicepolicy/" + id)
	if err != nil {
		return DevicePolicy{}, 0, err
	}
	x, err := ParseDevicePolicy(rec.Fields)
	if err != nil {
		return DevicePolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDevicePolicys reads every stored entity of this type, sorted.
func (r *Repo) ListDevicePolicys() ([]DevicePolicy, error) {
	recs := r.store.List("device/devicepolicy/")
	out := make([]DevicePolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDevicePolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDevicePolicys(out)
	return out, nil
}

// PutDeviceLink writes x, creating it when revision is zero.
func (r *Repo) PutDeviceLink(x DeviceLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicelink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicelink.written", 1)
	return rev, nil
}

// GetDeviceLink reads the entity with the given id.
func (r *Repo) GetDeviceLink(id string) (DeviceLink, int64, error) {
	rec, err := r.store.Get("device/devicelink/" + id)
	if err != nil {
		return DeviceLink{}, 0, err
	}
	x, err := ParseDeviceLink(rec.Fields)
	if err != nil {
		return DeviceLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceLinks reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceLinks() ([]DeviceLink, error) {
	recs := r.store.List("device/devicelink/")
	out := make([]DeviceLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceLinks(out)
	return out, nil
}

// PutDeviceNote writes x, creating it when revision is zero.
func (r *Repo) PutDeviceNote(x DeviceNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicenote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicenote.written", 1)
	return rev, nil
}

// GetDeviceNote reads the entity with the given id.
func (r *Repo) GetDeviceNote(id string) (DeviceNote, int64, error) {
	rec, err := r.store.Get("device/devicenote/" + id)
	if err != nil {
		return DeviceNote{}, 0, err
	}
	x, err := ParseDeviceNote(rec.Fields)
	if err != nil {
		return DeviceNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceNotes reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceNotes() ([]DeviceNote, error) {
	recs := r.store.List("device/devicenote/")
	out := make([]DeviceNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceNotes(out)
	return out, nil
}

// PutDeviceWindow writes x, creating it when revision is zero.
func (r *Repo) PutDeviceWindow(x DeviceWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicewindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicewindow.written", 1)
	return rev, nil
}

// GetDeviceWindow reads the entity with the given id.
func (r *Repo) GetDeviceWindow(id string) (DeviceWindow, int64, error) {
	rec, err := r.store.Get("device/devicewindow/" + id)
	if err != nil {
		return DeviceWindow{}, 0, err
	}
	x, err := ParseDeviceWindow(rec.Fields)
	if err != nil {
		return DeviceWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceWindows reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceWindows() ([]DeviceWindow, error) {
	recs := r.store.List("device/devicewindow/")
	out := make([]DeviceWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceWindows(out)
	return out, nil
}

// PutDeviceQuota writes x, creating it when revision is zero.
func (r *Repo) PutDeviceQuota(x DeviceQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicequota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicequota.written", 1)
	return rev, nil
}

// GetDeviceQuota reads the entity with the given id.
func (r *Repo) GetDeviceQuota(id string) (DeviceQuota, int64, error) {
	rec, err := r.store.Get("device/devicequota/" + id)
	if err != nil {
		return DeviceQuota{}, 0, err
	}
	x, err := ParseDeviceQuota(rec.Fields)
	if err != nil {
		return DeviceQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceQuotas() ([]DeviceQuota, error) {
	recs := r.store.List("device/devicequota/")
	out := make([]DeviceQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceQuotas(out)
	return out, nil
}

// PutDeviceBucket writes x, creating it when revision is zero.
func (r *Repo) PutDeviceBucket(x DeviceBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicebucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicebucket.written", 1)
	return rev, nil
}

// GetDeviceBucket reads the entity with the given id.
func (r *Repo) GetDeviceBucket(id string) (DeviceBucket, int64, error) {
	rec, err := r.store.Get("device/devicebucket/" + id)
	if err != nil {
		return DeviceBucket{}, 0, err
	}
	x, err := ParseDeviceBucket(rec.Fields)
	if err != nil {
		return DeviceBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceBuckets() ([]DeviceBucket, error) {
	recs := r.store.List("device/devicebucket/")
	out := make([]DeviceBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceBuckets(out)
	return out, nil
}

// PutDeviceLane writes x, creating it when revision is zero.
func (r *Repo) PutDeviceLane(x DeviceLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicelane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicelane.written", 1)
	return rev, nil
}

// GetDeviceLane reads the entity with the given id.
func (r *Repo) GetDeviceLane(id string) (DeviceLane, int64, error) {
	rec, err := r.store.Get("device/devicelane/" + id)
	if err != nil {
		return DeviceLane{}, 0, err
	}
	x, err := ParseDeviceLane(rec.Fields)
	if err != nil {
		return DeviceLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceLanes reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceLanes() ([]DeviceLane, error) {
	recs := r.store.List("device/devicelane/")
	out := make([]DeviceLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceLanes(out)
	return out, nil
}

// PutDeviceSlot writes x, creating it when revision is zero.
func (r *Repo) PutDeviceSlot(x DeviceSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.deviceslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.deviceslot.written", 1)
	return rev, nil
}

// GetDeviceSlot reads the entity with the given id.
func (r *Repo) GetDeviceSlot(id string) (DeviceSlot, int64, error) {
	rec, err := r.store.Get("device/deviceslot/" + id)
	if err != nil {
		return DeviceSlot{}, 0, err
	}
	x, err := ParseDeviceSlot(rec.Fields)
	if err != nil {
		return DeviceSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceSlots reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceSlots() ([]DeviceSlot, error) {
	recs := r.store.List("device/deviceslot/")
	out := make([]DeviceSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceSlots(out)
	return out, nil
}

// PutDeviceTag writes x, creating it when revision is zero.
func (r *Repo) PutDeviceTag(x DeviceTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicetag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicetag.written", 1)
	return rev, nil
}

// GetDeviceTag reads the entity with the given id.
func (r *Repo) GetDeviceTag(id string) (DeviceTag, int64, error) {
	rec, err := r.store.Get("device/devicetag/" + id)
	if err != nil {
		return DeviceTag{}, 0, err
	}
	x, err := ParseDeviceTag(rec.Fields)
	if err != nil {
		return DeviceTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceTags reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceTags() ([]DeviceTag, error) {
	recs := r.store.List("device/devicetag/")
	out := make([]DeviceTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceTags(out)
	return out, nil
}

// PutDeviceRule writes x, creating it when revision is zero.
func (r *Repo) PutDeviceRule(x DeviceRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicerule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicerule.written", 1)
	return rev, nil
}

// GetDeviceRule reads the entity with the given id.
func (r *Repo) GetDeviceRule(id string) (DeviceRule, int64, error) {
	rec, err := r.store.Get("device/devicerule/" + id)
	if err != nil {
		return DeviceRule{}, 0, err
	}
	x, err := ParseDeviceRule(rec.Fields)
	if err != nil {
		return DeviceRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceRules reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceRules() ([]DeviceRule, error) {
	recs := r.store.List("device/devicerule/")
	out := make([]DeviceRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceRules(out)
	return out, nil
}

// PutDeviceHold writes x, creating it when revision is zero.
func (r *Repo) PutDeviceHold(x DeviceHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicehold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicehold.written", 1)
	return rev, nil
}

// GetDeviceHold reads the entity with the given id.
func (r *Repo) GetDeviceHold(id string) (DeviceHold, int64, error) {
	rec, err := r.store.Get("device/devicehold/" + id)
	if err != nil {
		return DeviceHold{}, 0, err
	}
	x, err := ParseDeviceHold(rec.Fields)
	if err != nil {
		return DeviceHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceHolds reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceHolds() ([]DeviceHold, error) {
	recs := r.store.List("device/devicehold/")
	out := make([]DeviceHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceHolds(out)
	return out, nil
}

// PutDeviceClaim writes x, creating it when revision is zero.
func (r *Repo) PutDeviceClaim(x DeviceClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.deviceclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.deviceclaim.written", 1)
	return rev, nil
}

// GetDeviceClaim reads the entity with the given id.
func (r *Repo) GetDeviceClaim(id string) (DeviceClaim, int64, error) {
	rec, err := r.store.Get("device/deviceclaim/" + id)
	if err != nil {
		return DeviceClaim{}, 0, err
	}
	x, err := ParseDeviceClaim(rec.Fields)
	if err != nil {
		return DeviceClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceClaims reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceClaims() ([]DeviceClaim, error) {
	recs := r.store.List("device/deviceclaim/")
	out := make([]DeviceClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceClaims(out)
	return out, nil
}

// PutDeviceCursor writes x, creating it when revision is zero.
func (r *Repo) PutDeviceCursor(x DeviceCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("device.devicecursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("device write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("device.devicecursor.written", 1)
	return rev, nil
}

// GetDeviceCursor reads the entity with the given id.
func (r *Repo) GetDeviceCursor(id string) (DeviceCursor, int64, error) {
	rec, err := r.store.Get("device/devicecursor/" + id)
	if err != nil {
		return DeviceCursor{}, 0, err
	}
	x, err := ParseDeviceCursor(rec.Fields)
	if err != nil {
		return DeviceCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListDeviceCursors reads every stored entity of this type, sorted.
func (r *Repo) ListDeviceCursors() ([]DeviceCursor, error) {
	recs := r.store.List("device/devicecursor/")
	out := make([]DeviceCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseDeviceCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("device: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortDeviceCursors(out)
	return out, nil
}
