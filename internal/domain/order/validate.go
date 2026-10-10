package order

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded order records against this package's schemas.
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
		return fmt.Errorf("order: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against OrderRecordSchema.
func (x OrderRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderrecord: %w", err)
	}
	if err := validate.Record(OrderRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderRecords returns the first problem found in xs.
func ValidateOrderRecords(xs []OrderRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderEntrySchema.
func (x OrderEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderentry: %w", err)
	}
	if err := validate.Record(OrderEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderEntrys returns the first problem found in xs.
func ValidateOrderEntrys(xs []OrderEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderItemSchema.
func (x OrderItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderitem: %w", err)
	}
	if err := validate.Record(OrderItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderItems returns the first problem found in xs.
func ValidateOrderItems(xs []OrderItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderDraftSchema.
func (x OrderDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderdraft: %w", err)
	}
	if err := validate.Record(OrderDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderDrafts returns the first problem found in xs.
func ValidateOrderDrafts(xs []OrderDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderRevisionSchema.
func (x OrderRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderrevision: %w", err)
	}
	if err := validate.Record(OrderRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderRevisions returns the first problem found in xs.
func ValidateOrderRevisions(xs []OrderRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderSnapshotSchema.
func (x OrderSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: ordersnapshot: %w", err)
	}
	if err := validate.Record(OrderSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: ordersnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderSnapshots returns the first problem found in xs.
func ValidateOrderSnapshots(xs []OrderSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderBatchSchema.
func (x OrderBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderbatch: %w", err)
	}
	if err := validate.Record(OrderBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderBatchs returns the first problem found in xs.
func ValidateOrderBatchs(xs []OrderBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderRequestSchema.
func (x OrderRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderrequest: %w", err)
	}
	if err := validate.Record(OrderRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderRequests returns the first problem found in xs.
func ValidateOrderRequests(xs []OrderRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderSummarySchema.
func (x OrderSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: ordersummary: %w", err)
	}
	if err := validate.Record(OrderSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: ordersummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderSummarys returns the first problem found in xs.
func ValidateOrderSummarys(xs []OrderSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderPolicySchema.
func (x OrderPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderpolicy: %w", err)
	}
	if err := validate.Record(OrderPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderPolicys returns the first problem found in xs.
func ValidateOrderPolicys(xs []OrderPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderLinkSchema.
func (x OrderLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderlink: %w", err)
	}
	if err := validate.Record(OrderLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderLinks returns the first problem found in xs.
func ValidateOrderLinks(xs []OrderLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderNoteSchema.
func (x OrderNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: ordernote: %w", err)
	}
	if err := validate.Record(OrderNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: ordernote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderNotes returns the first problem found in xs.
func ValidateOrderNotes(xs []OrderNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderWindowSchema.
func (x OrderWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderwindow: %w", err)
	}
	if err := validate.Record(OrderWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderWindows returns the first problem found in xs.
func ValidateOrderWindows(xs []OrderWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderQuotaSchema.
func (x OrderQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderquota: %w", err)
	}
	if err := validate.Record(OrderQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderQuotas returns the first problem found in xs.
func ValidateOrderQuotas(xs []OrderQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderBucketSchema.
func (x OrderBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderbucket: %w", err)
	}
	if err := validate.Record(OrderBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderBuckets returns the first problem found in xs.
func ValidateOrderBuckets(xs []OrderBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderLaneSchema.
func (x OrderLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderlane: %w", err)
	}
	if err := validate.Record(OrderLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderLanes returns the first problem found in xs.
func ValidateOrderLanes(xs []OrderLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderSlotSchema.
func (x OrderSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderslot: %w", err)
	}
	if err := validate.Record(OrderSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderSlots returns the first problem found in xs.
func ValidateOrderSlots(xs []OrderSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderTagSchema.
func (x OrderTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: ordertag: %w", err)
	}
	if err := validate.Record(OrderTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: ordertag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderTags returns the first problem found in xs.
func ValidateOrderTags(xs []OrderTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderRuleSchema.
func (x OrderRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderrule: %w", err)
	}
	if err := validate.Record(OrderRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderRules returns the first problem found in xs.
func ValidateOrderRules(xs []OrderRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderHoldSchema.
func (x OrderHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderhold: %w", err)
	}
	if err := validate.Record(OrderHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderHolds returns the first problem found in xs.
func ValidateOrderHolds(xs []OrderHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderClaimSchema.
func (x OrderClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: orderclaim: %w", err)
	}
	if err := validate.Record(OrderClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: orderclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderClaims returns the first problem found in xs.
func ValidateOrderClaims(xs []OrderClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against OrderCursorSchema.
func (x OrderCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("order: ordercursor: %w", err)
	}
	if err := validate.Record(OrderCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("order: ordercursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateOrderCursors returns the first problem found in xs.
func ValidateOrderCursors(xs []OrderCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
