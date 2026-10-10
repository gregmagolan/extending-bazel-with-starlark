package notification

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists notification entities into a store.Store.
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
	l.Debug("notification repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many notification records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutNotificationRecord writes x, creating it when revision is zero.
func (r *Repo) PutNotificationRecord(x NotificationRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationrecord.written", 1)
	return rev, nil
}

// GetNotificationRecord reads the entity with the given id.
func (r *Repo) GetNotificationRecord(id string) (NotificationRecord, int64, error) {
	rec, err := r.store.Get("notification/notificationrecord/" + id)
	if err != nil {
		return NotificationRecord{}, 0, err
	}
	x, err := ParseNotificationRecord(rec.Fields)
	if err != nil {
		return NotificationRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationRecords reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationRecords() ([]NotificationRecord, error) {
	recs := r.store.List("notification/notificationrecord/")
	out := make([]NotificationRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationRecords(out)
	return out, nil
}

// PutNotificationEntry writes x, creating it when revision is zero.
func (r *Repo) PutNotificationEntry(x NotificationEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationentry.written", 1)
	return rev, nil
}

// GetNotificationEntry reads the entity with the given id.
func (r *Repo) GetNotificationEntry(id string) (NotificationEntry, int64, error) {
	rec, err := r.store.Get("notification/notificationentry/" + id)
	if err != nil {
		return NotificationEntry{}, 0, err
	}
	x, err := ParseNotificationEntry(rec.Fields)
	if err != nil {
		return NotificationEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationEntrys() ([]NotificationEntry, error) {
	recs := r.store.List("notification/notificationentry/")
	out := make([]NotificationEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationEntrys(out)
	return out, nil
}

// PutNotificationItem writes x, creating it when revision is zero.
func (r *Repo) PutNotificationItem(x NotificationItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationitem.written", 1)
	return rev, nil
}

// GetNotificationItem reads the entity with the given id.
func (r *Repo) GetNotificationItem(id string) (NotificationItem, int64, error) {
	rec, err := r.store.Get("notification/notificationitem/" + id)
	if err != nil {
		return NotificationItem{}, 0, err
	}
	x, err := ParseNotificationItem(rec.Fields)
	if err != nil {
		return NotificationItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationItems reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationItems() ([]NotificationItem, error) {
	recs := r.store.List("notification/notificationitem/")
	out := make([]NotificationItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationItems(out)
	return out, nil
}

// PutNotificationDraft writes x, creating it when revision is zero.
func (r *Repo) PutNotificationDraft(x NotificationDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationdraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationdraft.written", 1)
	return rev, nil
}

// GetNotificationDraft reads the entity with the given id.
func (r *Repo) GetNotificationDraft(id string) (NotificationDraft, int64, error) {
	rec, err := r.store.Get("notification/notificationdraft/" + id)
	if err != nil {
		return NotificationDraft{}, 0, err
	}
	x, err := ParseNotificationDraft(rec.Fields)
	if err != nil {
		return NotificationDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationDrafts() ([]NotificationDraft, error) {
	recs := r.store.List("notification/notificationdraft/")
	out := make([]NotificationDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationDrafts(out)
	return out, nil
}

// PutNotificationRevision writes x, creating it when revision is zero.
func (r *Repo) PutNotificationRevision(x NotificationRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationrevision.written", 1)
	return rev, nil
}

// GetNotificationRevision reads the entity with the given id.
func (r *Repo) GetNotificationRevision(id string) (NotificationRevision, int64, error) {
	rec, err := r.store.Get("notification/notificationrevision/" + id)
	if err != nil {
		return NotificationRevision{}, 0, err
	}
	x, err := ParseNotificationRevision(rec.Fields)
	if err != nil {
		return NotificationRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationRevisions() ([]NotificationRevision, error) {
	recs := r.store.List("notification/notificationrevision/")
	out := make([]NotificationRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationRevisions(out)
	return out, nil
}

// PutNotificationSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutNotificationSnapshot(x NotificationSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationsnapshot.written", 1)
	return rev, nil
}

// GetNotificationSnapshot reads the entity with the given id.
func (r *Repo) GetNotificationSnapshot(id string) (NotificationSnapshot, int64, error) {
	rec, err := r.store.Get("notification/notificationsnapshot/" + id)
	if err != nil {
		return NotificationSnapshot{}, 0, err
	}
	x, err := ParseNotificationSnapshot(rec.Fields)
	if err != nil {
		return NotificationSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationSnapshots() ([]NotificationSnapshot, error) {
	recs := r.store.List("notification/notificationsnapshot/")
	out := make([]NotificationSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationSnapshots(out)
	return out, nil
}

// PutNotificationBatch writes x, creating it when revision is zero.
func (r *Repo) PutNotificationBatch(x NotificationBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationbatch.written", 1)
	return rev, nil
}

// GetNotificationBatch reads the entity with the given id.
func (r *Repo) GetNotificationBatch(id string) (NotificationBatch, int64, error) {
	rec, err := r.store.Get("notification/notificationbatch/" + id)
	if err != nil {
		return NotificationBatch{}, 0, err
	}
	x, err := ParseNotificationBatch(rec.Fields)
	if err != nil {
		return NotificationBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationBatchs() ([]NotificationBatch, error) {
	recs := r.store.List("notification/notificationbatch/")
	out := make([]NotificationBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationBatchs(out)
	return out, nil
}

// PutNotificationRequest writes x, creating it when revision is zero.
func (r *Repo) PutNotificationRequest(x NotificationRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationrequest.written", 1)
	return rev, nil
}

// GetNotificationRequest reads the entity with the given id.
func (r *Repo) GetNotificationRequest(id string) (NotificationRequest, int64, error) {
	rec, err := r.store.Get("notification/notificationrequest/" + id)
	if err != nil {
		return NotificationRequest{}, 0, err
	}
	x, err := ParseNotificationRequest(rec.Fields)
	if err != nil {
		return NotificationRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationRequests reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationRequests() ([]NotificationRequest, error) {
	recs := r.store.List("notification/notificationrequest/")
	out := make([]NotificationRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationRequests(out)
	return out, nil
}

// PutNotificationSummary writes x, creating it when revision is zero.
func (r *Repo) PutNotificationSummary(x NotificationSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationsummary.written", 1)
	return rev, nil
}

// GetNotificationSummary reads the entity with the given id.
func (r *Repo) GetNotificationSummary(id string) (NotificationSummary, int64, error) {
	rec, err := r.store.Get("notification/notificationsummary/" + id)
	if err != nil {
		return NotificationSummary{}, 0, err
	}
	x, err := ParseNotificationSummary(rec.Fields)
	if err != nil {
		return NotificationSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationSummarys() ([]NotificationSummary, error) {
	recs := r.store.List("notification/notificationsummary/")
	out := make([]NotificationSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationSummarys(out)
	return out, nil
}

// PutNotificationPolicy writes x, creating it when revision is zero.
func (r *Repo) PutNotificationPolicy(x NotificationPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationpolicy.written", 1)
	return rev, nil
}

// GetNotificationPolicy reads the entity with the given id.
func (r *Repo) GetNotificationPolicy(id string) (NotificationPolicy, int64, error) {
	rec, err := r.store.Get("notification/notificationpolicy/" + id)
	if err != nil {
		return NotificationPolicy{}, 0, err
	}
	x, err := ParseNotificationPolicy(rec.Fields)
	if err != nil {
		return NotificationPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationPolicys() ([]NotificationPolicy, error) {
	recs := r.store.List("notification/notificationpolicy/")
	out := make([]NotificationPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationPolicys(out)
	return out, nil
}

// PutNotificationLink writes x, creating it when revision is zero.
func (r *Repo) PutNotificationLink(x NotificationLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationlink.written", 1)
	return rev, nil
}

// GetNotificationLink reads the entity with the given id.
func (r *Repo) GetNotificationLink(id string) (NotificationLink, int64, error) {
	rec, err := r.store.Get("notification/notificationlink/" + id)
	if err != nil {
		return NotificationLink{}, 0, err
	}
	x, err := ParseNotificationLink(rec.Fields)
	if err != nil {
		return NotificationLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationLinks reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationLinks() ([]NotificationLink, error) {
	recs := r.store.List("notification/notificationlink/")
	out := make([]NotificationLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationLinks(out)
	return out, nil
}

// PutNotificationNote writes x, creating it when revision is zero.
func (r *Repo) PutNotificationNote(x NotificationNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationnote.written", 1)
	return rev, nil
}

// GetNotificationNote reads the entity with the given id.
func (r *Repo) GetNotificationNote(id string) (NotificationNote, int64, error) {
	rec, err := r.store.Get("notification/notificationnote/" + id)
	if err != nil {
		return NotificationNote{}, 0, err
	}
	x, err := ParseNotificationNote(rec.Fields)
	if err != nil {
		return NotificationNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationNotes reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationNotes() ([]NotificationNote, error) {
	recs := r.store.List("notification/notificationnote/")
	out := make([]NotificationNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationNotes(out)
	return out, nil
}

// PutNotificationWindow writes x, creating it when revision is zero.
func (r *Repo) PutNotificationWindow(x NotificationWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationwindow.written", 1)
	return rev, nil
}

// GetNotificationWindow reads the entity with the given id.
func (r *Repo) GetNotificationWindow(id string) (NotificationWindow, int64, error) {
	rec, err := r.store.Get("notification/notificationwindow/" + id)
	if err != nil {
		return NotificationWindow{}, 0, err
	}
	x, err := ParseNotificationWindow(rec.Fields)
	if err != nil {
		return NotificationWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationWindows reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationWindows() ([]NotificationWindow, error) {
	recs := r.store.List("notification/notificationwindow/")
	out := make([]NotificationWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationWindows(out)
	return out, nil
}

// PutNotificationQuota writes x, creating it when revision is zero.
func (r *Repo) PutNotificationQuota(x NotificationQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationquota.written", 1)
	return rev, nil
}

// GetNotificationQuota reads the entity with the given id.
func (r *Repo) GetNotificationQuota(id string) (NotificationQuota, int64, error) {
	rec, err := r.store.Get("notification/notificationquota/" + id)
	if err != nil {
		return NotificationQuota{}, 0, err
	}
	x, err := ParseNotificationQuota(rec.Fields)
	if err != nil {
		return NotificationQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationQuotas() ([]NotificationQuota, error) {
	recs := r.store.List("notification/notificationquota/")
	out := make([]NotificationQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationQuotas(out)
	return out, nil
}

// PutNotificationBucket writes x, creating it when revision is zero.
func (r *Repo) PutNotificationBucket(x NotificationBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationbucket.written", 1)
	return rev, nil
}

// GetNotificationBucket reads the entity with the given id.
func (r *Repo) GetNotificationBucket(id string) (NotificationBucket, int64, error) {
	rec, err := r.store.Get("notification/notificationbucket/" + id)
	if err != nil {
		return NotificationBucket{}, 0, err
	}
	x, err := ParseNotificationBucket(rec.Fields)
	if err != nil {
		return NotificationBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationBuckets() ([]NotificationBucket, error) {
	recs := r.store.List("notification/notificationbucket/")
	out := make([]NotificationBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationBuckets(out)
	return out, nil
}

// PutNotificationLane writes x, creating it when revision is zero.
func (r *Repo) PutNotificationLane(x NotificationLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationlane.written", 1)
	return rev, nil
}

// GetNotificationLane reads the entity with the given id.
func (r *Repo) GetNotificationLane(id string) (NotificationLane, int64, error) {
	rec, err := r.store.Get("notification/notificationlane/" + id)
	if err != nil {
		return NotificationLane{}, 0, err
	}
	x, err := ParseNotificationLane(rec.Fields)
	if err != nil {
		return NotificationLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationLanes reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationLanes() ([]NotificationLane, error) {
	recs := r.store.List("notification/notificationlane/")
	out := make([]NotificationLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationLanes(out)
	return out, nil
}

// PutNotificationSlot writes x, creating it when revision is zero.
func (r *Repo) PutNotificationSlot(x NotificationSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationslot.written", 1)
	return rev, nil
}

// GetNotificationSlot reads the entity with the given id.
func (r *Repo) GetNotificationSlot(id string) (NotificationSlot, int64, error) {
	rec, err := r.store.Get("notification/notificationslot/" + id)
	if err != nil {
		return NotificationSlot{}, 0, err
	}
	x, err := ParseNotificationSlot(rec.Fields)
	if err != nil {
		return NotificationSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationSlots reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationSlots() ([]NotificationSlot, error) {
	recs := r.store.List("notification/notificationslot/")
	out := make([]NotificationSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationSlots(out)
	return out, nil
}

// PutNotificationTag writes x, creating it when revision is zero.
func (r *Repo) PutNotificationTag(x NotificationTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationtag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationtag.written", 1)
	return rev, nil
}

// GetNotificationTag reads the entity with the given id.
func (r *Repo) GetNotificationTag(id string) (NotificationTag, int64, error) {
	rec, err := r.store.Get("notification/notificationtag/" + id)
	if err != nil {
		return NotificationTag{}, 0, err
	}
	x, err := ParseNotificationTag(rec.Fields)
	if err != nil {
		return NotificationTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationTags reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationTags() ([]NotificationTag, error) {
	recs := r.store.List("notification/notificationtag/")
	out := make([]NotificationTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationTags(out)
	return out, nil
}

// PutNotificationRule writes x, creating it when revision is zero.
func (r *Repo) PutNotificationRule(x NotificationRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationrule.written", 1)
	return rev, nil
}

// GetNotificationRule reads the entity with the given id.
func (r *Repo) GetNotificationRule(id string) (NotificationRule, int64, error) {
	rec, err := r.store.Get("notification/notificationrule/" + id)
	if err != nil {
		return NotificationRule{}, 0, err
	}
	x, err := ParseNotificationRule(rec.Fields)
	if err != nil {
		return NotificationRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationRules reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationRules() ([]NotificationRule, error) {
	recs := r.store.List("notification/notificationrule/")
	out := make([]NotificationRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationRules(out)
	return out, nil
}

// PutNotificationHold writes x, creating it when revision is zero.
func (r *Repo) PutNotificationHold(x NotificationHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationhold.written", 1)
	return rev, nil
}

// GetNotificationHold reads the entity with the given id.
func (r *Repo) GetNotificationHold(id string) (NotificationHold, int64, error) {
	rec, err := r.store.Get("notification/notificationhold/" + id)
	if err != nil {
		return NotificationHold{}, 0, err
	}
	x, err := ParseNotificationHold(rec.Fields)
	if err != nil {
		return NotificationHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationHolds reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationHolds() ([]NotificationHold, error) {
	recs := r.store.List("notification/notificationhold/")
	out := make([]NotificationHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationHolds(out)
	return out, nil
}

// PutNotificationClaim writes x, creating it when revision is zero.
func (r *Repo) PutNotificationClaim(x NotificationClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationclaim.written", 1)
	return rev, nil
}

// GetNotificationClaim reads the entity with the given id.
func (r *Repo) GetNotificationClaim(id string) (NotificationClaim, int64, error) {
	rec, err := r.store.Get("notification/notificationclaim/" + id)
	if err != nil {
		return NotificationClaim{}, 0, err
	}
	x, err := ParseNotificationClaim(rec.Fields)
	if err != nil {
		return NotificationClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationClaims reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationClaims() ([]NotificationClaim, error) {
	recs := r.store.List("notification/notificationclaim/")
	out := make([]NotificationClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationClaims(out)
	return out, nil
}

// PutNotificationCursor writes x, creating it when revision is zero.
func (r *Repo) PutNotificationCursor(x NotificationCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("notification.notificationcursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("notification write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("notification.notificationcursor.written", 1)
	return rev, nil
}

// GetNotificationCursor reads the entity with the given id.
func (r *Repo) GetNotificationCursor(id string) (NotificationCursor, int64, error) {
	rec, err := r.store.Get("notification/notificationcursor/" + id)
	if err != nil {
		return NotificationCursor{}, 0, err
	}
	x, err := ParseNotificationCursor(rec.Fields)
	if err != nil {
		return NotificationCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListNotificationCursors reads every stored entity of this type, sorted.
func (r *Repo) ListNotificationCursors() ([]NotificationCursor, error) {
	recs := r.store.List("notification/notificationcursor/")
	out := make([]NotificationCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseNotificationCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("notification: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortNotificationCursors(out)
	return out, nil
}
