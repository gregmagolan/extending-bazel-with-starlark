package billing

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded billing records against this package's schemas.
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
		return fmt.Errorf("billing: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against BillingRecordSchema.
func (x BillingRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingrecord: %w", err)
	}
	if err := validate.Record(BillingRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingRecords returns the first problem found in xs.
func ValidateBillingRecords(xs []BillingRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingEntrySchema.
func (x BillingEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingentry: %w", err)
	}
	if err := validate.Record(BillingEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingEntrys returns the first problem found in xs.
func ValidateBillingEntrys(xs []BillingEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingItemSchema.
func (x BillingItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingitem: %w", err)
	}
	if err := validate.Record(BillingItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingItems returns the first problem found in xs.
func ValidateBillingItems(xs []BillingItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingDraftSchema.
func (x BillingDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingdraft: %w", err)
	}
	if err := validate.Record(BillingDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingDrafts returns the first problem found in xs.
func ValidateBillingDrafts(xs []BillingDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingRevisionSchema.
func (x BillingRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingrevision: %w", err)
	}
	if err := validate.Record(BillingRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingRevisions returns the first problem found in xs.
func ValidateBillingRevisions(xs []BillingRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingSnapshotSchema.
func (x BillingSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingsnapshot: %w", err)
	}
	if err := validate.Record(BillingSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingSnapshots returns the first problem found in xs.
func ValidateBillingSnapshots(xs []BillingSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingBatchSchema.
func (x BillingBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingbatch: %w", err)
	}
	if err := validate.Record(BillingBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingBatchs returns the first problem found in xs.
func ValidateBillingBatchs(xs []BillingBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingRequestSchema.
func (x BillingRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingrequest: %w", err)
	}
	if err := validate.Record(BillingRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingRequests returns the first problem found in xs.
func ValidateBillingRequests(xs []BillingRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingSummarySchema.
func (x BillingSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingsummary: %w", err)
	}
	if err := validate.Record(BillingSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingSummarys returns the first problem found in xs.
func ValidateBillingSummarys(xs []BillingSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingPolicySchema.
func (x BillingPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingpolicy: %w", err)
	}
	if err := validate.Record(BillingPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingPolicys returns the first problem found in xs.
func ValidateBillingPolicys(xs []BillingPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingLinkSchema.
func (x BillingLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billinglink: %w", err)
	}
	if err := validate.Record(BillingLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billinglink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingLinks returns the first problem found in xs.
func ValidateBillingLinks(xs []BillingLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingNoteSchema.
func (x BillingNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingnote: %w", err)
	}
	if err := validate.Record(BillingNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingNotes returns the first problem found in xs.
func ValidateBillingNotes(xs []BillingNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingWindowSchema.
func (x BillingWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingwindow: %w", err)
	}
	if err := validate.Record(BillingWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingWindows returns the first problem found in xs.
func ValidateBillingWindows(xs []BillingWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingQuotaSchema.
func (x BillingQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingquota: %w", err)
	}
	if err := validate.Record(BillingQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingQuotas returns the first problem found in xs.
func ValidateBillingQuotas(xs []BillingQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingBucketSchema.
func (x BillingBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingbucket: %w", err)
	}
	if err := validate.Record(BillingBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingBuckets returns the first problem found in xs.
func ValidateBillingBuckets(xs []BillingBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingLaneSchema.
func (x BillingLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billinglane: %w", err)
	}
	if err := validate.Record(BillingLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billinglane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingLanes returns the first problem found in xs.
func ValidateBillingLanes(xs []BillingLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingSlotSchema.
func (x BillingSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingslot: %w", err)
	}
	if err := validate.Record(BillingSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingSlots returns the first problem found in xs.
func ValidateBillingSlots(xs []BillingSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingTagSchema.
func (x BillingTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingtag: %w", err)
	}
	if err := validate.Record(BillingTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingTags returns the first problem found in xs.
func ValidateBillingTags(xs []BillingTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingRuleSchema.
func (x BillingRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingrule: %w", err)
	}
	if err := validate.Record(BillingRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingRules returns the first problem found in xs.
func ValidateBillingRules(xs []BillingRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingHoldSchema.
func (x BillingHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billinghold: %w", err)
	}
	if err := validate.Record(BillingHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billinghold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingHolds returns the first problem found in xs.
func ValidateBillingHolds(xs []BillingHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingClaimSchema.
func (x BillingClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingclaim: %w", err)
	}
	if err := validate.Record(BillingClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingClaims returns the first problem found in xs.
func ValidateBillingClaims(xs []BillingClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against BillingCursorSchema.
func (x BillingCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("billing: billingcursor: %w", err)
	}
	if err := validate.Record(BillingCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("billing: billingcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateBillingCursors returns the first problem found in xs.
func ValidateBillingCursors(xs []BillingCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
