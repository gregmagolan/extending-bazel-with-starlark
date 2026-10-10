package carrier

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded carrier records against this package's schemas.
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
		return fmt.Errorf("carrier: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against CarrierRecordSchema.
func (x CarrierRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierrecord: %w", err)
	}
	if err := validate.Record(CarrierRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierRecords returns the first problem found in xs.
func ValidateCarrierRecords(xs []CarrierRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierEntrySchema.
func (x CarrierEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierentry: %w", err)
	}
	if err := validate.Record(CarrierEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierEntrys returns the first problem found in xs.
func ValidateCarrierEntrys(xs []CarrierEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierItemSchema.
func (x CarrierItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrieritem: %w", err)
	}
	if err := validate.Record(CarrierItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrieritem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierItems returns the first problem found in xs.
func ValidateCarrierItems(xs []CarrierItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierDraftSchema.
func (x CarrierDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierdraft: %w", err)
	}
	if err := validate.Record(CarrierDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierDrafts returns the first problem found in xs.
func ValidateCarrierDrafts(xs []CarrierDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierRevisionSchema.
func (x CarrierRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierrevision: %w", err)
	}
	if err := validate.Record(CarrierRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierRevisions returns the first problem found in xs.
func ValidateCarrierRevisions(xs []CarrierRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierSnapshotSchema.
func (x CarrierSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carriersnapshot: %w", err)
	}
	if err := validate.Record(CarrierSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carriersnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierSnapshots returns the first problem found in xs.
func ValidateCarrierSnapshots(xs []CarrierSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierBatchSchema.
func (x CarrierBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierbatch: %w", err)
	}
	if err := validate.Record(CarrierBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierBatchs returns the first problem found in xs.
func ValidateCarrierBatchs(xs []CarrierBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierRequestSchema.
func (x CarrierRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierrequest: %w", err)
	}
	if err := validate.Record(CarrierRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierRequests returns the first problem found in xs.
func ValidateCarrierRequests(xs []CarrierRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierSummarySchema.
func (x CarrierSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carriersummary: %w", err)
	}
	if err := validate.Record(CarrierSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carriersummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierSummarys returns the first problem found in xs.
func ValidateCarrierSummarys(xs []CarrierSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierPolicySchema.
func (x CarrierPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierpolicy: %w", err)
	}
	if err := validate.Record(CarrierPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierPolicys returns the first problem found in xs.
func ValidateCarrierPolicys(xs []CarrierPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierLinkSchema.
func (x CarrierLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierlink: %w", err)
	}
	if err := validate.Record(CarrierLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierLinks returns the first problem found in xs.
func ValidateCarrierLinks(xs []CarrierLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierNoteSchema.
func (x CarrierNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carriernote: %w", err)
	}
	if err := validate.Record(CarrierNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carriernote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierNotes returns the first problem found in xs.
func ValidateCarrierNotes(xs []CarrierNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierWindowSchema.
func (x CarrierWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierwindow: %w", err)
	}
	if err := validate.Record(CarrierWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierWindows returns the first problem found in xs.
func ValidateCarrierWindows(xs []CarrierWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierQuotaSchema.
func (x CarrierQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierquota: %w", err)
	}
	if err := validate.Record(CarrierQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierQuotas returns the first problem found in xs.
func ValidateCarrierQuotas(xs []CarrierQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierBucketSchema.
func (x CarrierBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierbucket: %w", err)
	}
	if err := validate.Record(CarrierBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierBuckets returns the first problem found in xs.
func ValidateCarrierBuckets(xs []CarrierBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierLaneSchema.
func (x CarrierLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierlane: %w", err)
	}
	if err := validate.Record(CarrierLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierLanes returns the first problem found in xs.
func ValidateCarrierLanes(xs []CarrierLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierSlotSchema.
func (x CarrierSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierslot: %w", err)
	}
	if err := validate.Record(CarrierSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierSlots returns the first problem found in xs.
func ValidateCarrierSlots(xs []CarrierSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierTagSchema.
func (x CarrierTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carriertag: %w", err)
	}
	if err := validate.Record(CarrierTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carriertag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierTags returns the first problem found in xs.
func ValidateCarrierTags(xs []CarrierTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierRuleSchema.
func (x CarrierRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierrule: %w", err)
	}
	if err := validate.Record(CarrierRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierRules returns the first problem found in xs.
func ValidateCarrierRules(xs []CarrierRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierHoldSchema.
func (x CarrierHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierhold: %w", err)
	}
	if err := validate.Record(CarrierHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierHolds returns the first problem found in xs.
func ValidateCarrierHolds(xs []CarrierHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierClaimSchema.
func (x CarrierClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carrierclaim: %w", err)
	}
	if err := validate.Record(CarrierClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carrierclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierClaims returns the first problem found in xs.
func ValidateCarrierClaims(xs []CarrierClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CarrierCursorSchema.
func (x CarrierCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("carrier: carriercursor: %w", err)
	}
	if err := validate.Record(CarrierCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("carrier: carriercursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCarrierCursors returns the first problem found in xs.
func ValidateCarrierCursors(xs []CarrierCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
