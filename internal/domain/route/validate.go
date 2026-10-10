package route

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded route records against this package's schemas.
//
// The service layer calls Check before a write so that a bad request fails
// with every problem listed at once, rather than one problem per round trip.
type Validator struct{ strict bool }

// NewValidator returns a Validator. In strict mode an unknown schema name is
// an error instead of a pass.
func NewValidator(strict bool) Validator { return Validator{strict: strict} }

// Check validates f against the named schema.
func (v Validator) Check(name string, f codec.Fields) error {
	for _, s := range Schemas {
		if s.Name == name {
			return validate.Record(s, f).Err()
		}
	}
	if v.strict {
		return fmt.Errorf("route: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against RouteRecordSchema.
func (x RouteRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routerecord: %w", err)
	}
	if err := validate.Record(RouteRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteRecords returns the first problem found in xs.
func ValidateRouteRecords(xs []RouteRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteEntrySchema.
func (x RouteEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routeentry: %w", err)
	}
	if err := validate.Record(RouteEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routeentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteEntrys returns the first problem found in xs.
func ValidateRouteEntrys(xs []RouteEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteItemSchema.
func (x RouteItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routeitem: %w", err)
	}
	if err := validate.Record(RouteItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routeitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteItems returns the first problem found in xs.
func ValidateRouteItems(xs []RouteItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteDraftSchema.
func (x RouteDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routedraft: %w", err)
	}
	if err := validate.Record(RouteDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteDrafts returns the first problem found in xs.
func ValidateRouteDrafts(xs []RouteDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteRevisionSchema.
func (x RouteRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routerevision: %w", err)
	}
	if err := validate.Record(RouteRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteRevisions returns the first problem found in xs.
func ValidateRouteRevisions(xs []RouteRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteSnapshotSchema.
func (x RouteSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routesnapshot: %w", err)
	}
	if err := validate.Record(RouteSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteSnapshots returns the first problem found in xs.
func ValidateRouteSnapshots(xs []RouteSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteBatchSchema.
func (x RouteBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routebatch: %w", err)
	}
	if err := validate.Record(RouteBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteBatchs returns the first problem found in xs.
func ValidateRouteBatchs(xs []RouteBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteRequestSchema.
func (x RouteRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routerequest: %w", err)
	}
	if err := validate.Record(RouteRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteRequests returns the first problem found in xs.
func ValidateRouteRequests(xs []RouteRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteSummarySchema.
func (x RouteSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routesummary: %w", err)
	}
	if err := validate.Record(RouteSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteSummarys returns the first problem found in xs.
func ValidateRouteSummarys(xs []RouteSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RoutePolicySchema.
func (x RoutePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routepolicy: %w", err)
	}
	if err := validate.Record(RoutePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRoutePolicys returns the first problem found in xs.
func ValidateRoutePolicys(xs []RoutePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteLinkSchema.
func (x RouteLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routelink: %w", err)
	}
	if err := validate.Record(RouteLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteLinks returns the first problem found in xs.
func ValidateRouteLinks(xs []RouteLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteNoteSchema.
func (x RouteNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routenote: %w", err)
	}
	if err := validate.Record(RouteNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteNotes returns the first problem found in xs.
func ValidateRouteNotes(xs []RouteNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteWindowSchema.
func (x RouteWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routewindow: %w", err)
	}
	if err := validate.Record(RouteWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteWindows returns the first problem found in xs.
func ValidateRouteWindows(xs []RouteWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteQuotaSchema.
func (x RouteQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routequota: %w", err)
	}
	if err := validate.Record(RouteQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteQuotas returns the first problem found in xs.
func ValidateRouteQuotas(xs []RouteQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteBucketSchema.
func (x RouteBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routebucket: %w", err)
	}
	if err := validate.Record(RouteBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteBuckets returns the first problem found in xs.
func ValidateRouteBuckets(xs []RouteBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteLaneSchema.
func (x RouteLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routelane: %w", err)
	}
	if err := validate.Record(RouteLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteLanes returns the first problem found in xs.
func ValidateRouteLanes(xs []RouteLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteSlotSchema.
func (x RouteSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routeslot: %w", err)
	}
	if err := validate.Record(RouteSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routeslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteSlots returns the first problem found in xs.
func ValidateRouteSlots(xs []RouteSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteTagSchema.
func (x RouteTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routetag: %w", err)
	}
	if err := validate.Record(RouteTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteTags returns the first problem found in xs.
func ValidateRouteTags(xs []RouteTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteRuleSchema.
func (x RouteRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routerule: %w", err)
	}
	if err := validate.Record(RouteRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteRules returns the first problem found in xs.
func ValidateRouteRules(xs []RouteRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteHoldSchema.
func (x RouteHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routehold: %w", err)
	}
	if err := validate.Record(RouteHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteHolds returns the first problem found in xs.
func ValidateRouteHolds(xs []RouteHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteClaimSchema.
func (x RouteClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routeclaim: %w", err)
	}
	if err := validate.Record(RouteClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routeclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteClaims returns the first problem found in xs.
func ValidateRouteClaims(xs []RouteClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RouteCursorSchema.
func (x RouteCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("route: routecursor: %w", err)
	}
	if err := validate.Record(RouteCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("route: routecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRouteCursors returns the first problem found in xs.
func ValidateRouteCursors(xs []RouteCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
