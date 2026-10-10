package route

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Repo persists route entities into a store.Store.
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
	l.Debug("route repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many route records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}

// PutRouteRecord writes x, creating it when revision is zero.
func (r *Repo) PutRouteRecord(x RouteRecord, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routerecord.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routerecord.written", 1)
	return rev, nil
}

// GetRouteRecord reads the entity with the given id.
func (r *Repo) GetRouteRecord(id string) (RouteRecord, int64, error) {
	rec, err := r.store.Get("route/routerecord/" + id)
	if err != nil {
		return RouteRecord{}, 0, err
	}
	x, err := ParseRouteRecord(rec.Fields)
	if err != nil {
		return RouteRecord{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteRecords reads every stored entity of this type, sorted.
func (r *Repo) ListRouteRecords() ([]RouteRecord, error) {
	recs := r.store.List("route/routerecord/")
	out := make([]RouteRecord, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteRecord(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteRecords(out)
	return out, nil
}

// PutRouteEntry writes x, creating it when revision is zero.
func (r *Repo) PutRouteEntry(x RouteEntry, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routeentry.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routeentry.written", 1)
	return rev, nil
}

// GetRouteEntry reads the entity with the given id.
func (r *Repo) GetRouteEntry(id string) (RouteEntry, int64, error) {
	rec, err := r.store.Get("route/routeentry/" + id)
	if err != nil {
		return RouteEntry{}, 0, err
	}
	x, err := ParseRouteEntry(rec.Fields)
	if err != nil {
		return RouteEntry{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteEntrys reads every stored entity of this type, sorted.
func (r *Repo) ListRouteEntrys() ([]RouteEntry, error) {
	recs := r.store.List("route/routeentry/")
	out := make([]RouteEntry, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteEntry(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteEntrys(out)
	return out, nil
}

// PutRouteItem writes x, creating it when revision is zero.
func (r *Repo) PutRouteItem(x RouteItem, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routeitem.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routeitem.written", 1)
	return rev, nil
}

// GetRouteItem reads the entity with the given id.
func (r *Repo) GetRouteItem(id string) (RouteItem, int64, error) {
	rec, err := r.store.Get("route/routeitem/" + id)
	if err != nil {
		return RouteItem{}, 0, err
	}
	x, err := ParseRouteItem(rec.Fields)
	if err != nil {
		return RouteItem{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteItems reads every stored entity of this type, sorted.
func (r *Repo) ListRouteItems() ([]RouteItem, error) {
	recs := r.store.List("route/routeitem/")
	out := make([]RouteItem, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteItem(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteItems(out)
	return out, nil
}

// PutRouteDraft writes x, creating it when revision is zero.
func (r *Repo) PutRouteDraft(x RouteDraft, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routedraft.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routedraft.written", 1)
	return rev, nil
}

// GetRouteDraft reads the entity with the given id.
func (r *Repo) GetRouteDraft(id string) (RouteDraft, int64, error) {
	rec, err := r.store.Get("route/routedraft/" + id)
	if err != nil {
		return RouteDraft{}, 0, err
	}
	x, err := ParseRouteDraft(rec.Fields)
	if err != nil {
		return RouteDraft{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteDrafts reads every stored entity of this type, sorted.
func (r *Repo) ListRouteDrafts() ([]RouteDraft, error) {
	recs := r.store.List("route/routedraft/")
	out := make([]RouteDraft, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteDraft(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteDrafts(out)
	return out, nil
}

// PutRouteRevision writes x, creating it when revision is zero.
func (r *Repo) PutRouteRevision(x RouteRevision, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routerevision.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routerevision.written", 1)
	return rev, nil
}

// GetRouteRevision reads the entity with the given id.
func (r *Repo) GetRouteRevision(id string) (RouteRevision, int64, error) {
	rec, err := r.store.Get("route/routerevision/" + id)
	if err != nil {
		return RouteRevision{}, 0, err
	}
	x, err := ParseRouteRevision(rec.Fields)
	if err != nil {
		return RouteRevision{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteRevisions reads every stored entity of this type, sorted.
func (r *Repo) ListRouteRevisions() ([]RouteRevision, error) {
	recs := r.store.List("route/routerevision/")
	out := make([]RouteRevision, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteRevision(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteRevisions(out)
	return out, nil
}

// PutRouteSnapshot writes x, creating it when revision is zero.
func (r *Repo) PutRouteSnapshot(x RouteSnapshot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routesnapshot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routesnapshot.written", 1)
	return rev, nil
}

// GetRouteSnapshot reads the entity with the given id.
func (r *Repo) GetRouteSnapshot(id string) (RouteSnapshot, int64, error) {
	rec, err := r.store.Get("route/routesnapshot/" + id)
	if err != nil {
		return RouteSnapshot{}, 0, err
	}
	x, err := ParseRouteSnapshot(rec.Fields)
	if err != nil {
		return RouteSnapshot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteSnapshots reads every stored entity of this type, sorted.
func (r *Repo) ListRouteSnapshots() ([]RouteSnapshot, error) {
	recs := r.store.List("route/routesnapshot/")
	out := make([]RouteSnapshot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteSnapshot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteSnapshots(out)
	return out, nil
}

// PutRouteBatch writes x, creating it when revision is zero.
func (r *Repo) PutRouteBatch(x RouteBatch, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routebatch.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routebatch.written", 1)
	return rev, nil
}

// GetRouteBatch reads the entity with the given id.
func (r *Repo) GetRouteBatch(id string) (RouteBatch, int64, error) {
	rec, err := r.store.Get("route/routebatch/" + id)
	if err != nil {
		return RouteBatch{}, 0, err
	}
	x, err := ParseRouteBatch(rec.Fields)
	if err != nil {
		return RouteBatch{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteBatchs reads every stored entity of this type, sorted.
func (r *Repo) ListRouteBatchs() ([]RouteBatch, error) {
	recs := r.store.List("route/routebatch/")
	out := make([]RouteBatch, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteBatch(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteBatchs(out)
	return out, nil
}

// PutRouteRequest writes x, creating it when revision is zero.
func (r *Repo) PutRouteRequest(x RouteRequest, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routerequest.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routerequest.written", 1)
	return rev, nil
}

// GetRouteRequest reads the entity with the given id.
func (r *Repo) GetRouteRequest(id string) (RouteRequest, int64, error) {
	rec, err := r.store.Get("route/routerequest/" + id)
	if err != nil {
		return RouteRequest{}, 0, err
	}
	x, err := ParseRouteRequest(rec.Fields)
	if err != nil {
		return RouteRequest{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteRequests reads every stored entity of this type, sorted.
func (r *Repo) ListRouteRequests() ([]RouteRequest, error) {
	recs := r.store.List("route/routerequest/")
	out := make([]RouteRequest, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteRequest(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteRequests(out)
	return out, nil
}

// PutRouteSummary writes x, creating it when revision is zero.
func (r *Repo) PutRouteSummary(x RouteSummary, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routesummary.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routesummary.written", 1)
	return rev, nil
}

// GetRouteSummary reads the entity with the given id.
func (r *Repo) GetRouteSummary(id string) (RouteSummary, int64, error) {
	rec, err := r.store.Get("route/routesummary/" + id)
	if err != nil {
		return RouteSummary{}, 0, err
	}
	x, err := ParseRouteSummary(rec.Fields)
	if err != nil {
		return RouteSummary{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteSummarys reads every stored entity of this type, sorted.
func (r *Repo) ListRouteSummarys() ([]RouteSummary, error) {
	recs := r.store.List("route/routesummary/")
	out := make([]RouteSummary, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteSummary(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteSummarys(out)
	return out, nil
}

// PutRoutePolicy writes x, creating it when revision is zero.
func (r *Repo) PutRoutePolicy(x RoutePolicy, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routepolicy.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routepolicy.written", 1)
	return rev, nil
}

// GetRoutePolicy reads the entity with the given id.
func (r *Repo) GetRoutePolicy(id string) (RoutePolicy, int64, error) {
	rec, err := r.store.Get("route/routepolicy/" + id)
	if err != nil {
		return RoutePolicy{}, 0, err
	}
	x, err := ParseRoutePolicy(rec.Fields)
	if err != nil {
		return RoutePolicy{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRoutePolicys reads every stored entity of this type, sorted.
func (r *Repo) ListRoutePolicys() ([]RoutePolicy, error) {
	recs := r.store.List("route/routepolicy/")
	out := make([]RoutePolicy, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRoutePolicy(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRoutePolicys(out)
	return out, nil
}

// PutRouteLink writes x, creating it when revision is zero.
func (r *Repo) PutRouteLink(x RouteLink, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routelink.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routelink.written", 1)
	return rev, nil
}

// GetRouteLink reads the entity with the given id.
func (r *Repo) GetRouteLink(id string) (RouteLink, int64, error) {
	rec, err := r.store.Get("route/routelink/" + id)
	if err != nil {
		return RouteLink{}, 0, err
	}
	x, err := ParseRouteLink(rec.Fields)
	if err != nil {
		return RouteLink{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteLinks reads every stored entity of this type, sorted.
func (r *Repo) ListRouteLinks() ([]RouteLink, error) {
	recs := r.store.List("route/routelink/")
	out := make([]RouteLink, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteLink(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteLinks(out)
	return out, nil
}

// PutRouteNote writes x, creating it when revision is zero.
func (r *Repo) PutRouteNote(x RouteNote, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routenote.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routenote.written", 1)
	return rev, nil
}

// GetRouteNote reads the entity with the given id.
func (r *Repo) GetRouteNote(id string) (RouteNote, int64, error) {
	rec, err := r.store.Get("route/routenote/" + id)
	if err != nil {
		return RouteNote{}, 0, err
	}
	x, err := ParseRouteNote(rec.Fields)
	if err != nil {
		return RouteNote{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteNotes reads every stored entity of this type, sorted.
func (r *Repo) ListRouteNotes() ([]RouteNote, error) {
	recs := r.store.List("route/routenote/")
	out := make([]RouteNote, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteNote(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteNotes(out)
	return out, nil
}

// PutRouteWindow writes x, creating it when revision is zero.
func (r *Repo) PutRouteWindow(x RouteWindow, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routewindow.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routewindow.written", 1)
	return rev, nil
}

// GetRouteWindow reads the entity with the given id.
func (r *Repo) GetRouteWindow(id string) (RouteWindow, int64, error) {
	rec, err := r.store.Get("route/routewindow/" + id)
	if err != nil {
		return RouteWindow{}, 0, err
	}
	x, err := ParseRouteWindow(rec.Fields)
	if err != nil {
		return RouteWindow{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteWindows reads every stored entity of this type, sorted.
func (r *Repo) ListRouteWindows() ([]RouteWindow, error) {
	recs := r.store.List("route/routewindow/")
	out := make([]RouteWindow, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteWindow(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteWindows(out)
	return out, nil
}

// PutRouteQuota writes x, creating it when revision is zero.
func (r *Repo) PutRouteQuota(x RouteQuota, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routequota.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routequota.written", 1)
	return rev, nil
}

// GetRouteQuota reads the entity with the given id.
func (r *Repo) GetRouteQuota(id string) (RouteQuota, int64, error) {
	rec, err := r.store.Get("route/routequota/" + id)
	if err != nil {
		return RouteQuota{}, 0, err
	}
	x, err := ParseRouteQuota(rec.Fields)
	if err != nil {
		return RouteQuota{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteQuotas reads every stored entity of this type, sorted.
func (r *Repo) ListRouteQuotas() ([]RouteQuota, error) {
	recs := r.store.List("route/routequota/")
	out := make([]RouteQuota, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteQuota(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteQuotas(out)
	return out, nil
}

// PutRouteBucket writes x, creating it when revision is zero.
func (r *Repo) PutRouteBucket(x RouteBucket, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routebucket.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routebucket.written", 1)
	return rev, nil
}

// GetRouteBucket reads the entity with the given id.
func (r *Repo) GetRouteBucket(id string) (RouteBucket, int64, error) {
	rec, err := r.store.Get("route/routebucket/" + id)
	if err != nil {
		return RouteBucket{}, 0, err
	}
	x, err := ParseRouteBucket(rec.Fields)
	if err != nil {
		return RouteBucket{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteBuckets reads every stored entity of this type, sorted.
func (r *Repo) ListRouteBuckets() ([]RouteBucket, error) {
	recs := r.store.List("route/routebucket/")
	out := make([]RouteBucket, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteBucket(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteBuckets(out)
	return out, nil
}

// PutRouteLane writes x, creating it when revision is zero.
func (r *Repo) PutRouteLane(x RouteLane, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routelane.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routelane.written", 1)
	return rev, nil
}

// GetRouteLane reads the entity with the given id.
func (r *Repo) GetRouteLane(id string) (RouteLane, int64, error) {
	rec, err := r.store.Get("route/routelane/" + id)
	if err != nil {
		return RouteLane{}, 0, err
	}
	x, err := ParseRouteLane(rec.Fields)
	if err != nil {
		return RouteLane{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteLanes reads every stored entity of this type, sorted.
func (r *Repo) ListRouteLanes() ([]RouteLane, error) {
	recs := r.store.List("route/routelane/")
	out := make([]RouteLane, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteLane(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteLanes(out)
	return out, nil
}

// PutRouteSlot writes x, creating it when revision is zero.
func (r *Repo) PutRouteSlot(x RouteSlot, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routeslot.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routeslot.written", 1)
	return rev, nil
}

// GetRouteSlot reads the entity with the given id.
func (r *Repo) GetRouteSlot(id string) (RouteSlot, int64, error) {
	rec, err := r.store.Get("route/routeslot/" + id)
	if err != nil {
		return RouteSlot{}, 0, err
	}
	x, err := ParseRouteSlot(rec.Fields)
	if err != nil {
		return RouteSlot{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteSlots reads every stored entity of this type, sorted.
func (r *Repo) ListRouteSlots() ([]RouteSlot, error) {
	recs := r.store.List("route/routeslot/")
	out := make([]RouteSlot, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteSlot(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteSlots(out)
	return out, nil
}

// PutRouteTag writes x, creating it when revision is zero.
func (r *Repo) PutRouteTag(x RouteTag, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routetag.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routetag.written", 1)
	return rev, nil
}

// GetRouteTag reads the entity with the given id.
func (r *Repo) GetRouteTag(id string) (RouteTag, int64, error) {
	rec, err := r.store.Get("route/routetag/" + id)
	if err != nil {
		return RouteTag{}, 0, err
	}
	x, err := ParseRouteTag(rec.Fields)
	if err != nil {
		return RouteTag{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteTags reads every stored entity of this type, sorted.
func (r *Repo) ListRouteTags() ([]RouteTag, error) {
	recs := r.store.List("route/routetag/")
	out := make([]RouteTag, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteTag(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteTags(out)
	return out, nil
}

// PutRouteRule writes x, creating it when revision is zero.
func (r *Repo) PutRouteRule(x RouteRule, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routerule.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routerule.written", 1)
	return rev, nil
}

// GetRouteRule reads the entity with the given id.
func (r *Repo) GetRouteRule(id string) (RouteRule, int64, error) {
	rec, err := r.store.Get("route/routerule/" + id)
	if err != nil {
		return RouteRule{}, 0, err
	}
	x, err := ParseRouteRule(rec.Fields)
	if err != nil {
		return RouteRule{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteRules reads every stored entity of this type, sorted.
func (r *Repo) ListRouteRules() ([]RouteRule, error) {
	recs := r.store.List("route/routerule/")
	out := make([]RouteRule, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteRule(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteRules(out)
	return out, nil
}

// PutRouteHold writes x, creating it when revision is zero.
func (r *Repo) PutRouteHold(x RouteHold, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routehold.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routehold.written", 1)
	return rev, nil
}

// GetRouteHold reads the entity with the given id.
func (r *Repo) GetRouteHold(id string) (RouteHold, int64, error) {
	rec, err := r.store.Get("route/routehold/" + id)
	if err != nil {
		return RouteHold{}, 0, err
	}
	x, err := ParseRouteHold(rec.Fields)
	if err != nil {
		return RouteHold{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteHolds reads every stored entity of this type, sorted.
func (r *Repo) ListRouteHolds() ([]RouteHold, error) {
	recs := r.store.List("route/routehold/")
	out := make([]RouteHold, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteHold(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteHolds(out)
	return out, nil
}

// PutRouteClaim writes x, creating it when revision is zero.
func (r *Repo) PutRouteClaim(x RouteClaim, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routeclaim.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routeclaim.written", 1)
	return rev, nil
}

// GetRouteClaim reads the entity with the given id.
func (r *Repo) GetRouteClaim(id string) (RouteClaim, int64, error) {
	rec, err := r.store.Get("route/routeclaim/" + id)
	if err != nil {
		return RouteClaim{}, 0, err
	}
	x, err := ParseRouteClaim(rec.Fields)
	if err != nil {
		return RouteClaim{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteClaims reads every stored entity of this type, sorted.
func (r *Repo) ListRouteClaims() ([]RouteClaim, error) {
	recs := r.store.List("route/routeclaim/")
	out := make([]RouteClaim, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteClaim(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteClaims(out)
	return out, nil
}

// PutRouteCursor writes x, creating it when revision is zero.
func (r *Repo) PutRouteCursor(x RouteCursor, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("route.routecursor.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("route write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("route.routecursor.written", 1)
	return rev, nil
}

// GetRouteCursor reads the entity with the given id.
func (r *Repo) GetRouteCursor(id string) (RouteCursor, int64, error) {
	rec, err := r.store.Get("route/routecursor/" + id)
	if err != nil {
		return RouteCursor{}, 0, err
	}
	x, err := ParseRouteCursor(rec.Fields)
	if err != nil {
		return RouteCursor{}, 0, err
	}
	return x, rec.Revision, nil
}

// ListRouteCursors reads every stored entity of this type, sorted.
func (r *Repo) ListRouteCursors() ([]RouteCursor, error) {
	recs := r.store.List("route/routecursor/")
	out := make([]RouteCursor, 0, len(recs))
	for _, rec := range recs {
		x, err := ParseRouteCursor(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("route: %s: %w", rec.Key, err)
		}
		out = append(out, x)
	}
	SortRouteCursors(out)
	return out, nil
}
