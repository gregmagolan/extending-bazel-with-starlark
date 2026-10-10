package rollout

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded rollout records against this package's schemas.
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
		return fmt.Errorf("rollout: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against RolloutRecordSchema.
func (x RolloutRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutrecord: %w", err)
	}
	if err := validate.Record(RolloutRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutRecords returns the first problem found in xs.
func ValidateRolloutRecords(xs []RolloutRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutEntrySchema.
func (x RolloutEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutentry: %w", err)
	}
	if err := validate.Record(RolloutEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutEntrys returns the first problem found in xs.
func ValidateRolloutEntrys(xs []RolloutEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutItemSchema.
func (x RolloutItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutitem: %w", err)
	}
	if err := validate.Record(RolloutItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutItems returns the first problem found in xs.
func ValidateRolloutItems(xs []RolloutItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutDraftSchema.
func (x RolloutDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutdraft: %w", err)
	}
	if err := validate.Record(RolloutDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutDrafts returns the first problem found in xs.
func ValidateRolloutDrafts(xs []RolloutDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutRevisionSchema.
func (x RolloutRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutrevision: %w", err)
	}
	if err := validate.Record(RolloutRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutRevisions returns the first problem found in xs.
func ValidateRolloutRevisions(xs []RolloutRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutSnapshotSchema.
func (x RolloutSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutsnapshot: %w", err)
	}
	if err := validate.Record(RolloutSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutSnapshots returns the first problem found in xs.
func ValidateRolloutSnapshots(xs []RolloutSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutBatchSchema.
func (x RolloutBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutbatch: %w", err)
	}
	if err := validate.Record(RolloutBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutBatchs returns the first problem found in xs.
func ValidateRolloutBatchs(xs []RolloutBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutRequestSchema.
func (x RolloutRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutrequest: %w", err)
	}
	if err := validate.Record(RolloutRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutRequests returns the first problem found in xs.
func ValidateRolloutRequests(xs []RolloutRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutSummarySchema.
func (x RolloutSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutsummary: %w", err)
	}
	if err := validate.Record(RolloutSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutSummarys returns the first problem found in xs.
func ValidateRolloutSummarys(xs []RolloutSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutPolicySchema.
func (x RolloutPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutpolicy: %w", err)
	}
	if err := validate.Record(RolloutPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutPolicys returns the first problem found in xs.
func ValidateRolloutPolicys(xs []RolloutPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutLinkSchema.
func (x RolloutLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutlink: %w", err)
	}
	if err := validate.Record(RolloutLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutLinks returns the first problem found in xs.
func ValidateRolloutLinks(xs []RolloutLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutNoteSchema.
func (x RolloutNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutnote: %w", err)
	}
	if err := validate.Record(RolloutNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutNotes returns the first problem found in xs.
func ValidateRolloutNotes(xs []RolloutNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutWindowSchema.
func (x RolloutWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutwindow: %w", err)
	}
	if err := validate.Record(RolloutWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutWindows returns the first problem found in xs.
func ValidateRolloutWindows(xs []RolloutWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutQuotaSchema.
func (x RolloutQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutquota: %w", err)
	}
	if err := validate.Record(RolloutQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutQuotas returns the first problem found in xs.
func ValidateRolloutQuotas(xs []RolloutQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutBucketSchema.
func (x RolloutBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutbucket: %w", err)
	}
	if err := validate.Record(RolloutBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutBuckets returns the first problem found in xs.
func ValidateRolloutBuckets(xs []RolloutBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutLaneSchema.
func (x RolloutLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutlane: %w", err)
	}
	if err := validate.Record(RolloutLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutLanes returns the first problem found in xs.
func ValidateRolloutLanes(xs []RolloutLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutSlotSchema.
func (x RolloutSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutslot: %w", err)
	}
	if err := validate.Record(RolloutSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutSlots returns the first problem found in xs.
func ValidateRolloutSlots(xs []RolloutSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutTagSchema.
func (x RolloutTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rollouttag: %w", err)
	}
	if err := validate.Record(RolloutTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rollouttag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutTags returns the first problem found in xs.
func ValidateRolloutTags(xs []RolloutTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutRuleSchema.
func (x RolloutRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutrule: %w", err)
	}
	if err := validate.Record(RolloutRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutRules returns the first problem found in xs.
func ValidateRolloutRules(xs []RolloutRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutHoldSchema.
func (x RolloutHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rollouthold: %w", err)
	}
	if err := validate.Record(RolloutHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rollouthold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutHolds returns the first problem found in xs.
func ValidateRolloutHolds(xs []RolloutHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutClaimSchema.
func (x RolloutClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutclaim: %w", err)
	}
	if err := validate.Record(RolloutClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutClaims returns the first problem found in xs.
func ValidateRolloutClaims(xs []RolloutClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RolloutCursorSchema.
func (x RolloutCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("rollout: rolloutcursor: %w", err)
	}
	if err := validate.Record(RolloutCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("rollout: rolloutcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRolloutCursors returns the first problem found in xs.
func ValidateRolloutCursors(xs []RolloutCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
