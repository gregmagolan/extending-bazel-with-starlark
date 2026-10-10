package shipping

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded shipping records against this package's schemas.
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
		return fmt.Errorf("shipping: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against ShippingRecordSchema.
func (x ShippingRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingrecord: %w", err)
	}
	if err := validate.Record(ShippingRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingRecords returns the first problem found in xs.
func ValidateShippingRecords(xs []ShippingRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingEntrySchema.
func (x ShippingEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingentry: %w", err)
	}
	if err := validate.Record(ShippingEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingEntrys returns the first problem found in xs.
func ValidateShippingEntrys(xs []ShippingEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingItemSchema.
func (x ShippingItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingitem: %w", err)
	}
	if err := validate.Record(ShippingItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingItems returns the first problem found in xs.
func ValidateShippingItems(xs []ShippingItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingDraftSchema.
func (x ShippingDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingdraft: %w", err)
	}
	if err := validate.Record(ShippingDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingDrafts returns the first problem found in xs.
func ValidateShippingDrafts(xs []ShippingDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingRevisionSchema.
func (x ShippingRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingrevision: %w", err)
	}
	if err := validate.Record(ShippingRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingRevisions returns the first problem found in xs.
func ValidateShippingRevisions(xs []ShippingRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingSnapshotSchema.
func (x ShippingSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingsnapshot: %w", err)
	}
	if err := validate.Record(ShippingSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingSnapshots returns the first problem found in xs.
func ValidateShippingSnapshots(xs []ShippingSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingBatchSchema.
func (x ShippingBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingbatch: %w", err)
	}
	if err := validate.Record(ShippingBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingBatchs returns the first problem found in xs.
func ValidateShippingBatchs(xs []ShippingBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingRequestSchema.
func (x ShippingRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingrequest: %w", err)
	}
	if err := validate.Record(ShippingRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingRequests returns the first problem found in xs.
func ValidateShippingRequests(xs []ShippingRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingSummarySchema.
func (x ShippingSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingsummary: %w", err)
	}
	if err := validate.Record(ShippingSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingSummarys returns the first problem found in xs.
func ValidateShippingSummarys(xs []ShippingSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingPolicySchema.
func (x ShippingPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingpolicy: %w", err)
	}
	if err := validate.Record(ShippingPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingPolicys returns the first problem found in xs.
func ValidateShippingPolicys(xs []ShippingPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingLinkSchema.
func (x ShippingLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippinglink: %w", err)
	}
	if err := validate.Record(ShippingLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippinglink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingLinks returns the first problem found in xs.
func ValidateShippingLinks(xs []ShippingLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingNoteSchema.
func (x ShippingNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingnote: %w", err)
	}
	if err := validate.Record(ShippingNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingNotes returns the first problem found in xs.
func ValidateShippingNotes(xs []ShippingNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingWindowSchema.
func (x ShippingWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingwindow: %w", err)
	}
	if err := validate.Record(ShippingWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingWindows returns the first problem found in xs.
func ValidateShippingWindows(xs []ShippingWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingQuotaSchema.
func (x ShippingQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingquota: %w", err)
	}
	if err := validate.Record(ShippingQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingQuotas returns the first problem found in xs.
func ValidateShippingQuotas(xs []ShippingQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingBucketSchema.
func (x ShippingBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingbucket: %w", err)
	}
	if err := validate.Record(ShippingBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingBuckets returns the first problem found in xs.
func ValidateShippingBuckets(xs []ShippingBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingLaneSchema.
func (x ShippingLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippinglane: %w", err)
	}
	if err := validate.Record(ShippingLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippinglane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingLanes returns the first problem found in xs.
func ValidateShippingLanes(xs []ShippingLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingSlotSchema.
func (x ShippingSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingslot: %w", err)
	}
	if err := validate.Record(ShippingSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingSlots returns the first problem found in xs.
func ValidateShippingSlots(xs []ShippingSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingTagSchema.
func (x ShippingTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingtag: %w", err)
	}
	if err := validate.Record(ShippingTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingTags returns the first problem found in xs.
func ValidateShippingTags(xs []ShippingTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingRuleSchema.
func (x ShippingRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingrule: %w", err)
	}
	if err := validate.Record(ShippingRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingRules returns the first problem found in xs.
func ValidateShippingRules(xs []ShippingRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingHoldSchema.
func (x ShippingHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippinghold: %w", err)
	}
	if err := validate.Record(ShippingHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippinghold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingHolds returns the first problem found in xs.
func ValidateShippingHolds(xs []ShippingHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingClaimSchema.
func (x ShippingClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingclaim: %w", err)
	}
	if err := validate.Record(ShippingClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingClaims returns the first problem found in xs.
func ValidateShippingClaims(xs []ShippingClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ShippingCursorSchema.
func (x ShippingCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("shipping: shippingcursor: %w", err)
	}
	if err := validate.Record(ShippingCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("shipping: shippingcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateShippingCursors returns the first problem found in xs.
func ValidateShippingCursors(xs []ShippingCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
