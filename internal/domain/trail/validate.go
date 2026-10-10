package trail

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded trail records against this package's schemas.
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
		return fmt.Errorf("trail: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against TrailRecordSchema.
func (x TrailRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailrecord: %w", err)
	}
	if err := validate.Record(TrailRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailRecords returns the first problem found in xs.
func ValidateTrailRecords(xs []TrailRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailEntrySchema.
func (x TrailEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailentry: %w", err)
	}
	if err := validate.Record(TrailEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailEntrys returns the first problem found in xs.
func ValidateTrailEntrys(xs []TrailEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailItemSchema.
func (x TrailItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailitem: %w", err)
	}
	if err := validate.Record(TrailItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailItems returns the first problem found in xs.
func ValidateTrailItems(xs []TrailItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailDraftSchema.
func (x TrailDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: traildraft: %w", err)
	}
	if err := validate.Record(TrailDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: traildraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailDrafts returns the first problem found in xs.
func ValidateTrailDrafts(xs []TrailDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailRevisionSchema.
func (x TrailRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailrevision: %w", err)
	}
	if err := validate.Record(TrailRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailRevisions returns the first problem found in xs.
func ValidateTrailRevisions(xs []TrailRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailSnapshotSchema.
func (x TrailSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailsnapshot: %w", err)
	}
	if err := validate.Record(TrailSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailSnapshots returns the first problem found in xs.
func ValidateTrailSnapshots(xs []TrailSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailBatchSchema.
func (x TrailBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailbatch: %w", err)
	}
	if err := validate.Record(TrailBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailBatchs returns the first problem found in xs.
func ValidateTrailBatchs(xs []TrailBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailRequestSchema.
func (x TrailRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailrequest: %w", err)
	}
	if err := validate.Record(TrailRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailRequests returns the first problem found in xs.
func ValidateTrailRequests(xs []TrailRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailSummarySchema.
func (x TrailSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailsummary: %w", err)
	}
	if err := validate.Record(TrailSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailSummarys returns the first problem found in xs.
func ValidateTrailSummarys(xs []TrailSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailPolicySchema.
func (x TrailPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailpolicy: %w", err)
	}
	if err := validate.Record(TrailPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailPolicys returns the first problem found in xs.
func ValidateTrailPolicys(xs []TrailPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailLinkSchema.
func (x TrailLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: traillink: %w", err)
	}
	if err := validate.Record(TrailLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: traillink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailLinks returns the first problem found in xs.
func ValidateTrailLinks(xs []TrailLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailNoteSchema.
func (x TrailNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailnote: %w", err)
	}
	if err := validate.Record(TrailNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailNotes returns the first problem found in xs.
func ValidateTrailNotes(xs []TrailNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailWindowSchema.
func (x TrailWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailwindow: %w", err)
	}
	if err := validate.Record(TrailWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailWindows returns the first problem found in xs.
func ValidateTrailWindows(xs []TrailWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailQuotaSchema.
func (x TrailQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailquota: %w", err)
	}
	if err := validate.Record(TrailQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailQuotas returns the first problem found in xs.
func ValidateTrailQuotas(xs []TrailQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailBucketSchema.
func (x TrailBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailbucket: %w", err)
	}
	if err := validate.Record(TrailBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailBuckets returns the first problem found in xs.
func ValidateTrailBuckets(xs []TrailBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailLaneSchema.
func (x TrailLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: traillane: %w", err)
	}
	if err := validate.Record(TrailLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: traillane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailLanes returns the first problem found in xs.
func ValidateTrailLanes(xs []TrailLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailSlotSchema.
func (x TrailSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailslot: %w", err)
	}
	if err := validate.Record(TrailSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailSlots returns the first problem found in xs.
func ValidateTrailSlots(xs []TrailSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailTagSchema.
func (x TrailTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailtag: %w", err)
	}
	if err := validate.Record(TrailTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailTags returns the first problem found in xs.
func ValidateTrailTags(xs []TrailTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailRuleSchema.
func (x TrailRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailrule: %w", err)
	}
	if err := validate.Record(TrailRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailRules returns the first problem found in xs.
func ValidateTrailRules(xs []TrailRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailHoldSchema.
func (x TrailHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailhold: %w", err)
	}
	if err := validate.Record(TrailHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailHolds returns the first problem found in xs.
func ValidateTrailHolds(xs []TrailHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailClaimSchema.
func (x TrailClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailclaim: %w", err)
	}
	if err := validate.Record(TrailClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailClaims returns the first problem found in xs.
func ValidateTrailClaims(xs []TrailClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TrailCursorSchema.
func (x TrailCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("trail: trailcursor: %w", err)
	}
	if err := validate.Record(TrailCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("trail: trailcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTrailCursors returns the first problem found in xs.
func ValidateTrailCursors(xs []TrailCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
