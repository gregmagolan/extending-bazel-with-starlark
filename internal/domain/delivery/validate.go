package delivery

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded delivery records against this package's schemas.
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
		return fmt.Errorf("delivery: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against DeliveryRecordSchema.
func (x DeliveryRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryrecord: %w", err)
	}
	if err := validate.Record(DeliveryRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryRecords returns the first problem found in xs.
func ValidateDeliveryRecords(xs []DeliveryRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryEntrySchema.
func (x DeliveryEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryentry: %w", err)
	}
	if err := validate.Record(DeliveryEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryEntrys returns the first problem found in xs.
func ValidateDeliveryEntrys(xs []DeliveryEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryItemSchema.
func (x DeliveryItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryitem: %w", err)
	}
	if err := validate.Record(DeliveryItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryItems returns the first problem found in xs.
func ValidateDeliveryItems(xs []DeliveryItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryDraftSchema.
func (x DeliveryDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverydraft: %w", err)
	}
	if err := validate.Record(DeliveryDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverydraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryDrafts returns the first problem found in xs.
func ValidateDeliveryDrafts(xs []DeliveryDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryRevisionSchema.
func (x DeliveryRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryrevision: %w", err)
	}
	if err := validate.Record(DeliveryRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryRevisions returns the first problem found in xs.
func ValidateDeliveryRevisions(xs []DeliveryRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliverySnapshotSchema.
func (x DeliverySnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverysnapshot: %w", err)
	}
	if err := validate.Record(DeliverySnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverysnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliverySnapshots returns the first problem found in xs.
func ValidateDeliverySnapshots(xs []DeliverySnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryBatchSchema.
func (x DeliveryBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverybatch: %w", err)
	}
	if err := validate.Record(DeliveryBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverybatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryBatchs returns the first problem found in xs.
func ValidateDeliveryBatchs(xs []DeliveryBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryRequestSchema.
func (x DeliveryRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryrequest: %w", err)
	}
	if err := validate.Record(DeliveryRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryRequests returns the first problem found in xs.
func ValidateDeliveryRequests(xs []DeliveryRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliverySummarySchema.
func (x DeliverySummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverysummary: %w", err)
	}
	if err := validate.Record(DeliverySummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverysummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliverySummarys returns the first problem found in xs.
func ValidateDeliverySummarys(xs []DeliverySummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryPolicySchema.
func (x DeliveryPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverypolicy: %w", err)
	}
	if err := validate.Record(DeliveryPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverypolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryPolicys returns the first problem found in xs.
func ValidateDeliveryPolicys(xs []DeliveryPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryLinkSchema.
func (x DeliveryLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverylink: %w", err)
	}
	if err := validate.Record(DeliveryLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverylink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryLinks returns the first problem found in xs.
func ValidateDeliveryLinks(xs []DeliveryLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryNoteSchema.
func (x DeliveryNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverynote: %w", err)
	}
	if err := validate.Record(DeliveryNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverynote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryNotes returns the first problem found in xs.
func ValidateDeliveryNotes(xs []DeliveryNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryWindowSchema.
func (x DeliveryWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverywindow: %w", err)
	}
	if err := validate.Record(DeliveryWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverywindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryWindows returns the first problem found in xs.
func ValidateDeliveryWindows(xs []DeliveryWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryQuotaSchema.
func (x DeliveryQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryquota: %w", err)
	}
	if err := validate.Record(DeliveryQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryQuotas returns the first problem found in xs.
func ValidateDeliveryQuotas(xs []DeliveryQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryBucketSchema.
func (x DeliveryBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverybucket: %w", err)
	}
	if err := validate.Record(DeliveryBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverybucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryBuckets returns the first problem found in xs.
func ValidateDeliveryBuckets(xs []DeliveryBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryLaneSchema.
func (x DeliveryLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverylane: %w", err)
	}
	if err := validate.Record(DeliveryLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverylane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryLanes returns the first problem found in xs.
func ValidateDeliveryLanes(xs []DeliveryLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliverySlotSchema.
func (x DeliverySlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryslot: %w", err)
	}
	if err := validate.Record(DeliverySlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliverySlots returns the first problem found in xs.
func ValidateDeliverySlots(xs []DeliverySlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryTagSchema.
func (x DeliveryTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverytag: %w", err)
	}
	if err := validate.Record(DeliveryTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverytag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryTags returns the first problem found in xs.
func ValidateDeliveryTags(xs []DeliveryTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryRuleSchema.
func (x DeliveryRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryrule: %w", err)
	}
	if err := validate.Record(DeliveryRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryRules returns the first problem found in xs.
func ValidateDeliveryRules(xs []DeliveryRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryHoldSchema.
func (x DeliveryHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryhold: %w", err)
	}
	if err := validate.Record(DeliveryHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryHolds returns the first problem found in xs.
func ValidateDeliveryHolds(xs []DeliveryHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryClaimSchema.
func (x DeliveryClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliveryclaim: %w", err)
	}
	if err := validate.Record(DeliveryClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliveryclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryClaims returns the first problem found in xs.
func ValidateDeliveryClaims(xs []DeliveryClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeliveryCursorSchema.
func (x DeliveryCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("delivery: deliverycursor: %w", err)
	}
	if err := validate.Record(DeliveryCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("delivery: deliverycursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeliveryCursors returns the first problem found in xs.
func ValidateDeliveryCursors(xs []DeliveryCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
