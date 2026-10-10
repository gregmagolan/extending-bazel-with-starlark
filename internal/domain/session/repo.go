package session

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists session entities into a store.Store.
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
	l.Debug("session repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many session records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutSessionRecord writes x, creating it when revision is zero.
func (r *Repo) PutSessionRecord(x SessionRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionrecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionrecord.written", 1)
	return rev, nil
}

// GetSessionRecord reads the entity with the given id.
func (r *Repo) GetSessionRecord(id string) (SessionRecord, int64, error) {
	rec, err := r.store.Get("session/sessionrecord/" + id)
	if err != nil {
		return SessionRecord{}, 0, err
	}
	x, err := ParseSessionRecord(rec.Fields)
	if err != nil {
		return SessionRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionRecords reads every stored entity of this type, sorted.
func (r *Repo) ListSessionRecords() ([]SessionRecord, error) {
	recs := r.store.List("session/sessionrecord/")
	out := make([]SessionRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionRecords(out)
	return out, nil
}

// PutSessionEntry writes x, creating it when revision is zero.
func (r *Repo) PutSessionEntry(x SessionEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionentry.written", 1)
	return rev, nil
}

// GetSessionEntry reads the entity with the given id.
func (r *Repo) GetSessionEntry(id string) (SessionEntry, int64, error) {
	rec, err := r.store.Get("session/sessionentry/" + id)
	if err != nil {
		return SessionEntry{}, 0, err
	}
	x, err := ParseSessionEntry(rec.Fields)
	if err != nil {
		return SessionEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListSessionEntrys() ([]SessionEntry, error) {
	recs := r.store.List("session/sessionentry/")
	out := make([]SessionEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionEntrys(out)
	return out, nil
}

// PutSessionItem writes x, creating it when revision is zero.
func (r *Repo) PutSessionItem(x SessionItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionitem.written", 1)
	return rev, nil
}

// GetSessionItem reads the entity with the given id.
func (r *Repo) GetSessionItem(id string) (SessionItem, int64, error) {
	rec, err := r.store.Get("session/sessionitem/" + id)
	if err != nil {
		return SessionItem{}, 0, err
	}
	x, err := ParseSessionItem(rec.Fields)
	if err != nil {
		return SessionItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionItems reads every stored entity of this type, sorted.
func (r *Repo) ListSessionItems() ([]SessionItem, error) {
	recs := r.store.List("session/sessionitem/")
	out := make([]SessionItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionItems(out)
	return out, nil
}

// PutSessionDraft writes x, creating it when revision is zero.
func (r *Repo) PutSessionDraft(x SessionDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessiondraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessiondraft.written", 1)
	return rev, nil
}

// GetSessionDraft reads the entity with the given id.
func (r *Repo) GetSessionDraft(id string) (SessionDraft, int64, error) {
	rec, err := r.store.Get("session/sessiondraft/" + id)
	if err != nil {
		return SessionDraft{}, 0, err
	}
	x, err := ParseSessionDraft(rec.Fields)
	if err != nil {
		return SessionDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListSessionDrafts() ([]SessionDraft, error) {
	recs := r.store.List("session/sessiondraft/")
	out := make([]SessionDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionDrafts(out)
	return out, nil
}

// PutSessionRevision writes x, creating it when revision is zero.
func (r *Repo) PutSessionRevision(x SessionRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionrevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionrevision.written", 1)
	return rev, nil
}

// GetSessionRevision reads the entity with the given id.
func (r *Repo) GetSessionRevision(id string) (SessionRevision, int64, error) {
	rec, err := r.store.Get("session/sessionrevision/" + id)
	if err != nil {
		return SessionRevision{}, 0, err
	}
	x, err := ParseSessionRevision(rec.Fields)
	if err != nil {
		return SessionRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListSessionRevisions() ([]SessionRevision, error) {
	recs := r.store.List("session/sessionrevision/")
	out := make([]SessionRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionRevisions(out)
	return out, nil
}

// PutSessionSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutSessionSnapshot(x SessionSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionsnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionsnapshot.written", 1)
	return rev, nil
}

// GetSessionSnapshot reads the entity with the given id.
func (r *Repo) GetSessionSnapshot(id string) (SessionSnapshot, int64, error) {
	rec, err := r.store.Get("session/sessionsnapshot/" + id)
	if err != nil {
		return SessionSnapshot{}, 0, err
	}
	x, err := ParseSessionSnapshot(rec.Fields)
	if err != nil {
		return SessionSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListSessionSnapshots() ([]SessionSnapshot, error) {
	recs := r.store.List("session/sessionsnapshot/")
	out := make([]SessionSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionSnapshots(out)
	return out, nil
}

// PutSessionBatch writes x, creating it when revision is zero.
func (r *Repo) PutSessionBatch(x SessionBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionbatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionbatch.written", 1)
	return rev, nil
}

// GetSessionBatch reads the entity with the given id.
func (r *Repo) GetSessionBatch(id string) (SessionBatch, int64, error) {
	rec, err := r.store.Get("session/sessionbatch/" + id)
	if err != nil {
		return SessionBatch{}, 0, err
	}
	x, err := ParseSessionBatch(rec.Fields)
	if err != nil {
		return SessionBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListSessionBatchs() ([]SessionBatch, error) {
	recs := r.store.List("session/sessionbatch/")
	out := make([]SessionBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionBatchs(out)
	return out, nil
}

// PutSessionRequest writes x, creating it when revision is zero.
func (r *Repo) PutSessionRequest(x SessionRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionrequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionrequest.written", 1)
	return rev, nil
}

// GetSessionRequest reads the entity with the given id.
func (r *Repo) GetSessionRequest(id string) (SessionRequest, int64, error) {
	rec, err := r.store.Get("session/sessionrequest/" + id)
	if err != nil {
		return SessionRequest{}, 0, err
	}
	x, err := ParseSessionRequest(rec.Fields)
	if err != nil {
		return SessionRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionRequests reads every stored entity of this type, sorted.
func (r *Repo) ListSessionRequests() ([]SessionRequest, error) {
	recs := r.store.List("session/sessionrequest/")
	out := make([]SessionRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionRequests(out)
	return out, nil
}

// PutSessionSummary writes x, creating it when revision is zero.
func (r *Repo) PutSessionSummary(x SessionSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionsummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionsummary.written", 1)
	return rev, nil
}

// GetSessionSummary reads the entity with the given id.
func (r *Repo) GetSessionSummary(id string) (SessionSummary, int64, error) {
	rec, err := r.store.Get("session/sessionsummary/" + id)
	if err != nil {
		return SessionSummary{}, 0, err
	}
	x, err := ParseSessionSummary(rec.Fields)
	if err != nil {
		return SessionSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListSessionSummarys() ([]SessionSummary, error) {
	recs := r.store.List("session/sessionsummary/")
	out := make([]SessionSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionSummarys(out)
	return out, nil
}

// PutSessionPolicy writes x, creating it when revision is zero.
func (r *Repo) PutSessionPolicy(x SessionPolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionpolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionpolicy.written", 1)
	return rev, nil
}

// GetSessionPolicy reads the entity with the given id.
func (r *Repo) GetSessionPolicy(id string) (SessionPolicy, int64, error) {
	rec, err := r.store.Get("session/sessionpolicy/" + id)
	if err != nil {
		return SessionPolicy{}, 0, err
	}
	x, err := ParseSessionPolicy(rec.Fields)
	if err != nil {
		return SessionPolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionPolicys reads every stored entity of this type, sorted.
func (r *Repo) ListSessionPolicys() ([]SessionPolicy, error) {
	recs := r.store.List("session/sessionpolicy/")
	out := make([]SessionPolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionPolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionPolicys(out)
	return out, nil
}

// PutSessionLink writes x, creating it when revision is zero.
func (r *Repo) PutSessionLink(x SessionLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionlink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionlink.written", 1)
	return rev, nil
}

// GetSessionLink reads the entity with the given id.
func (r *Repo) GetSessionLink(id string) (SessionLink, int64, error) {
	rec, err := r.store.Get("session/sessionlink/" + id)
	if err != nil {
		return SessionLink{}, 0, err
	}
	x, err := ParseSessionLink(rec.Fields)
	if err != nil {
		return SessionLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionLinks reads every stored entity of this type, sorted.
func (r *Repo) ListSessionLinks() ([]SessionLink, error) {
	recs := r.store.List("session/sessionlink/")
	out := make([]SessionLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionLinks(out)
	return out, nil
}

// PutSessionNote writes x, creating it when revision is zero.
func (r *Repo) PutSessionNote(x SessionNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionnote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionnote.written", 1)
	return rev, nil
}

// GetSessionNote reads the entity with the given id.
func (r *Repo) GetSessionNote(id string) (SessionNote, int64, error) {
	rec, err := r.store.Get("session/sessionnote/" + id)
	if err != nil {
		return SessionNote{}, 0, err
	}
	x, err := ParseSessionNote(rec.Fields)
	if err != nil {
		return SessionNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionNotes reads every stored entity of this type, sorted.
func (r *Repo) ListSessionNotes() ([]SessionNote, error) {
	recs := r.store.List("session/sessionnote/")
	out := make([]SessionNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionNotes(out)
	return out, nil
}

// PutSessionWindow writes x, creating it when revision is zero.
func (r *Repo) PutSessionWindow(x SessionWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionwindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionwindow.written", 1)
	return rev, nil
}

// GetSessionWindow reads the entity with the given id.
func (r *Repo) GetSessionWindow(id string) (SessionWindow, int64, error) {
	rec, err := r.store.Get("session/sessionwindow/" + id)
	if err != nil {
		return SessionWindow{}, 0, err
	}
	x, err := ParseSessionWindow(rec.Fields)
	if err != nil {
		return SessionWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionWindows reads every stored entity of this type, sorted.
func (r *Repo) ListSessionWindows() ([]SessionWindow, error) {
	recs := r.store.List("session/sessionwindow/")
	out := make([]SessionWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionWindows(out)
	return out, nil
}

// PutSessionQuota writes x, creating it when revision is zero.
func (r *Repo) PutSessionQuota(x SessionQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionquota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionquota.written", 1)
	return rev, nil
}

// GetSessionQuota reads the entity with the given id.
func (r *Repo) GetSessionQuota(id string) (SessionQuota, int64, error) {
	rec, err := r.store.Get("session/sessionquota/" + id)
	if err != nil {
		return SessionQuota{}, 0, err
	}
	x, err := ParseSessionQuota(rec.Fields)
	if err != nil {
		return SessionQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListSessionQuotas() ([]SessionQuota, error) {
	recs := r.store.List("session/sessionquota/")
	out := make([]SessionQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionQuotas(out)
	return out, nil
}

// PutSessionBucket writes x, creating it when revision is zero.
func (r *Repo) PutSessionBucket(x SessionBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionbucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionbucket.written", 1)
	return rev, nil
}

// GetSessionBucket reads the entity with the given id.
func (r *Repo) GetSessionBucket(id string) (SessionBucket, int64, error) {
	rec, err := r.store.Get("session/sessionbucket/" + id)
	if err != nil {
		return SessionBucket{}, 0, err
	}
	x, err := ParseSessionBucket(rec.Fields)
	if err != nil {
		return SessionBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListSessionBuckets() ([]SessionBucket, error) {
	recs := r.store.List("session/sessionbucket/")
	out := make([]SessionBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionBuckets(out)
	return out, nil
}

// PutSessionLane writes x, creating it when revision is zero.
func (r *Repo) PutSessionLane(x SessionLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionlane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionlane.written", 1)
	return rev, nil
}

// GetSessionLane reads the entity with the given id.
func (r *Repo) GetSessionLane(id string) (SessionLane, int64, error) {
	rec, err := r.store.Get("session/sessionlane/" + id)
	if err != nil {
		return SessionLane{}, 0, err
	}
	x, err := ParseSessionLane(rec.Fields)
	if err != nil {
		return SessionLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionLanes reads every stored entity of this type, sorted.
func (r *Repo) ListSessionLanes() ([]SessionLane, error) {
	recs := r.store.List("session/sessionlane/")
	out := make([]SessionLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionLanes(out)
	return out, nil
}

// PutSessionSlot writes x, creating it when revision is zero.
func (r *Repo) PutSessionSlot(x SessionSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionslot.written", 1)
	return rev, nil
}

// GetSessionSlot reads the entity with the given id.
func (r *Repo) GetSessionSlot(id string) (SessionSlot, int64, error) {
	rec, err := r.store.Get("session/sessionslot/" + id)
	if err != nil {
		return SessionSlot{}, 0, err
	}
	x, err := ParseSessionSlot(rec.Fields)
	if err != nil {
		return SessionSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionSlots reads every stored entity of this type, sorted.
func (r *Repo) ListSessionSlots() ([]SessionSlot, error) {
	recs := r.store.List("session/sessionslot/")
	out := make([]SessionSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionSlots(out)
	return out, nil
}

// PutSessionTag writes x, creating it when revision is zero.
func (r *Repo) PutSessionTag(x SessionTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessiontag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessiontag.written", 1)
	return rev, nil
}

// GetSessionTag reads the entity with the given id.
func (r *Repo) GetSessionTag(id string) (SessionTag, int64, error) {
	rec, err := r.store.Get("session/sessiontag/" + id)
	if err != nil {
		return SessionTag{}, 0, err
	}
	x, err := ParseSessionTag(rec.Fields)
	if err != nil {
		return SessionTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionTags reads every stored entity of this type, sorted.
func (r *Repo) ListSessionTags() ([]SessionTag, error) {
	recs := r.store.List("session/sessiontag/")
	out := make([]SessionTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionTags(out)
	return out, nil
}

// PutSessionRule writes x, creating it when revision is zero.
func (r *Repo) PutSessionRule(x SessionRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionrule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionrule.written", 1)
	return rev, nil
}

// GetSessionRule reads the entity with the given id.
func (r *Repo) GetSessionRule(id string) (SessionRule, int64, error) {
	rec, err := r.store.Get("session/sessionrule/" + id)
	if err != nil {
		return SessionRule{}, 0, err
	}
	x, err := ParseSessionRule(rec.Fields)
	if err != nil {
		return SessionRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionRules reads every stored entity of this type, sorted.
func (r *Repo) ListSessionRules() ([]SessionRule, error) {
	recs := r.store.List("session/sessionrule/")
	out := make([]SessionRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionRules(out)
	return out, nil
}

// PutSessionHold writes x, creating it when revision is zero.
func (r *Repo) PutSessionHold(x SessionHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionhold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionhold.written", 1)
	return rev, nil
}

// GetSessionHold reads the entity with the given id.
func (r *Repo) GetSessionHold(id string) (SessionHold, int64, error) {
	rec, err := r.store.Get("session/sessionhold/" + id)
	if err != nil {
		return SessionHold{}, 0, err
	}
	x, err := ParseSessionHold(rec.Fields)
	if err != nil {
		return SessionHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionHolds reads every stored entity of this type, sorted.
func (r *Repo) ListSessionHolds() ([]SessionHold, error) {
	recs := r.store.List("session/sessionhold/")
	out := make([]SessionHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionHolds(out)
	return out, nil
}

// PutSessionClaim writes x, creating it when revision is zero.
func (r *Repo) PutSessionClaim(x SessionClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessionclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessionclaim.written", 1)
	return rev, nil
}

// GetSessionClaim reads the entity with the given id.
func (r *Repo) GetSessionClaim(id string) (SessionClaim, int64, error) {
	rec, err := r.store.Get("session/sessionclaim/" + id)
	if err != nil {
		return SessionClaim{}, 0, err
	}
	x, err := ParseSessionClaim(rec.Fields)
	if err != nil {
		return SessionClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionClaims reads every stored entity of this type, sorted.
func (r *Repo) ListSessionClaims() ([]SessionClaim, error) {
	recs := r.store.List("session/sessionclaim/")
	out := make([]SessionClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionClaims(out)
	return out, nil
}

// PutSessionCursor writes x, creating it when revision is zero.
func (r *Repo) PutSessionCursor(x SessionCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("session.sessioncursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("session write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("session.sessioncursor.written", 1)
	return rev, nil
}

// GetSessionCursor reads the entity with the given id.
func (r *Repo) GetSessionCursor(id string) (SessionCursor, int64, error) {
	rec, err := r.store.Get("session/sessioncursor/" + id)
	if err != nil {
		return SessionCursor{}, 0, err
	}
	x, err := ParseSessionCursor(rec.Fields)
	if err != nil {
		return SessionCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListSessionCursors reads every stored entity of this type, sorted.
func (r *Repo) ListSessionCursors() ([]SessionCursor, error) {
	recs := r.store.List("session/sessioncursor/")
	out := make([]SessionCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseSessionCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("session: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortSessionCursors(out)
	return out, nil
}
