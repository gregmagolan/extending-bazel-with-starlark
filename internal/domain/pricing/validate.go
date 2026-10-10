package pricing

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded pricing records against this package's schemas.
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
		return fmt.Errorf("pricing: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against PricingRecordSchema.
func (x PricingRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingrecord: %w", err)
	}
	if err := validate.Record(PricingRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingRecords returns the first problem found in xs.
func ValidatePricingRecords(xs []PricingRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingEntrySchema.
func (x PricingEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingentry: %w", err)
	}
	if err := validate.Record(PricingEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingEntrys returns the first problem found in xs.
func ValidatePricingEntrys(xs []PricingEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingItemSchema.
func (x PricingItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingitem: %w", err)
	}
	if err := validate.Record(PricingItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingItems returns the first problem found in xs.
func ValidatePricingItems(xs []PricingItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingDraftSchema.
func (x PricingDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingdraft: %w", err)
	}
	if err := validate.Record(PricingDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingDrafts returns the first problem found in xs.
func ValidatePricingDrafts(xs []PricingDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingRevisionSchema.
func (x PricingRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingrevision: %w", err)
	}
	if err := validate.Record(PricingRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingRevisions returns the first problem found in xs.
func ValidatePricingRevisions(xs []PricingRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingSnapshotSchema.
func (x PricingSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingsnapshot: %w", err)
	}
	if err := validate.Record(PricingSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingSnapshots returns the first problem found in xs.
func ValidatePricingSnapshots(xs []PricingSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingBatchSchema.
func (x PricingBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingbatch: %w", err)
	}
	if err := validate.Record(PricingBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingBatchs returns the first problem found in xs.
func ValidatePricingBatchs(xs []PricingBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingRequestSchema.
func (x PricingRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingrequest: %w", err)
	}
	if err := validate.Record(PricingRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingRequests returns the first problem found in xs.
func ValidatePricingRequests(xs []PricingRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingSummarySchema.
func (x PricingSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingsummary: %w", err)
	}
	if err := validate.Record(PricingSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingSummarys returns the first problem found in xs.
func ValidatePricingSummarys(xs []PricingSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingPolicySchema.
func (x PricingPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingpolicy: %w", err)
	}
	if err := validate.Record(PricingPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingPolicys returns the first problem found in xs.
func ValidatePricingPolicys(xs []PricingPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingLinkSchema.
func (x PricingLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricinglink: %w", err)
	}
	if err := validate.Record(PricingLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricinglink %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingLinks returns the first problem found in xs.
func ValidatePricingLinks(xs []PricingLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingNoteSchema.
func (x PricingNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingnote: %w", err)
	}
	if err := validate.Record(PricingNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingNotes returns the first problem found in xs.
func ValidatePricingNotes(xs []PricingNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingWindowSchema.
func (x PricingWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingwindow: %w", err)
	}
	if err := validate.Record(PricingWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingWindows returns the first problem found in xs.
func ValidatePricingWindows(xs []PricingWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingQuotaSchema.
func (x PricingQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingquota: %w", err)
	}
	if err := validate.Record(PricingQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingQuotas returns the first problem found in xs.
func ValidatePricingQuotas(xs []PricingQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingBucketSchema.
func (x PricingBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingbucket: %w", err)
	}
	if err := validate.Record(PricingBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingBuckets returns the first problem found in xs.
func ValidatePricingBuckets(xs []PricingBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingLaneSchema.
func (x PricingLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricinglane: %w", err)
	}
	if err := validate.Record(PricingLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricinglane %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingLanes returns the first problem found in xs.
func ValidatePricingLanes(xs []PricingLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingSlotSchema.
func (x PricingSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingslot: %w", err)
	}
	if err := validate.Record(PricingSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingSlots returns the first problem found in xs.
func ValidatePricingSlots(xs []PricingSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingTagSchema.
func (x PricingTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingtag: %w", err)
	}
	if err := validate.Record(PricingTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingTags returns the first problem found in xs.
func ValidatePricingTags(xs []PricingTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingRuleSchema.
func (x PricingRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingrule: %w", err)
	}
	if err := validate.Record(PricingRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingRules returns the first problem found in xs.
func ValidatePricingRules(xs []PricingRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingHoldSchema.
func (x PricingHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricinghold: %w", err)
	}
	if err := validate.Record(PricingHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricinghold %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingHolds returns the first problem found in xs.
func ValidatePricingHolds(xs []PricingHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingClaimSchema.
func (x PricingClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingclaim: %w", err)
	}
	if err := validate.Record(PricingClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingClaims returns the first problem found in xs.
func ValidatePricingClaims(xs []PricingClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PricingCursorSchema.
func (x PricingCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("pricing: pricingcursor: %w", err)
	}
	if err := validate.Record(PricingCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("pricing: pricingcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePricingCursors returns the first problem found in xs.
func ValidatePricingCursors(xs []PricingCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
