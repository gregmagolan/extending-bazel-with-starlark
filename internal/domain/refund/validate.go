package refund

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded refund records against this package's schemas.
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
		return fmt.Errorf("refund: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against RefundRecordSchema.
func (x RefundRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundrecord: %w", err)
	}
	if err := validate.Record(RefundRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundRecords returns the first problem found in xs.
func ValidateRefundRecords(xs []RefundRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundEntrySchema.
func (x RefundEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundentry: %w", err)
	}
	if err := validate.Record(RefundEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundEntrys returns the first problem found in xs.
func ValidateRefundEntrys(xs []RefundEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundItemSchema.
func (x RefundItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refunditem: %w", err)
	}
	if err := validate.Record(RefundItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refunditem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundItems returns the first problem found in xs.
func ValidateRefundItems(xs []RefundItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundDraftSchema.
func (x RefundDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refunddraft: %w", err)
	}
	if err := validate.Record(RefundDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refunddraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundDrafts returns the first problem found in xs.
func ValidateRefundDrafts(xs []RefundDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundRevisionSchema.
func (x RefundRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundrevision: %w", err)
	}
	if err := validate.Record(RefundRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundRevisions returns the first problem found in xs.
func ValidateRefundRevisions(xs []RefundRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundSnapshotSchema.
func (x RefundSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundsnapshot: %w", err)
	}
	if err := validate.Record(RefundSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundSnapshots returns the first problem found in xs.
func ValidateRefundSnapshots(xs []RefundSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundBatchSchema.
func (x RefundBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundbatch: %w", err)
	}
	if err := validate.Record(RefundBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundBatchs returns the first problem found in xs.
func ValidateRefundBatchs(xs []RefundBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundRequestSchema.
func (x RefundRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundrequest: %w", err)
	}
	if err := validate.Record(RefundRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundRequests returns the first problem found in xs.
func ValidateRefundRequests(xs []RefundRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundSummarySchema.
func (x RefundSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundsummary: %w", err)
	}
	if err := validate.Record(RefundSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundSummarys returns the first problem found in xs.
func ValidateRefundSummarys(xs []RefundSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundPolicySchema.
func (x RefundPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundpolicy: %w", err)
	}
	if err := validate.Record(RefundPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundPolicys returns the first problem found in xs.
func ValidateRefundPolicys(xs []RefundPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundLinkSchema.
func (x RefundLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundlink: %w", err)
	}
	if err := validate.Record(RefundLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundLinks returns the first problem found in xs.
func ValidateRefundLinks(xs []RefundLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundNoteSchema.
func (x RefundNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundnote: %w", err)
	}
	if err := validate.Record(RefundNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundNotes returns the first problem found in xs.
func ValidateRefundNotes(xs []RefundNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundWindowSchema.
func (x RefundWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundwindow: %w", err)
	}
	if err := validate.Record(RefundWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundWindows returns the first problem found in xs.
func ValidateRefundWindows(xs []RefundWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundQuotaSchema.
func (x RefundQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundquota: %w", err)
	}
	if err := validate.Record(RefundQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundQuotas returns the first problem found in xs.
func ValidateRefundQuotas(xs []RefundQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundBucketSchema.
func (x RefundBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundbucket: %w", err)
	}
	if err := validate.Record(RefundBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundBuckets returns the first problem found in xs.
func ValidateRefundBuckets(xs []RefundBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundLaneSchema.
func (x RefundLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundlane: %w", err)
	}
	if err := validate.Record(RefundLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundLanes returns the first problem found in xs.
func ValidateRefundLanes(xs []RefundLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundSlotSchema.
func (x RefundSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundslot: %w", err)
	}
	if err := validate.Record(RefundSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundSlots returns the first problem found in xs.
func ValidateRefundSlots(xs []RefundSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundTagSchema.
func (x RefundTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundtag: %w", err)
	}
	if err := validate.Record(RefundTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundTags returns the first problem found in xs.
func ValidateRefundTags(xs []RefundTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundRuleSchema.
func (x RefundRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundrule: %w", err)
	}
	if err := validate.Record(RefundRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundRules returns the first problem found in xs.
func ValidateRefundRules(xs []RefundRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundHoldSchema.
func (x RefundHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundhold: %w", err)
	}
	if err := validate.Record(RefundHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundHolds returns the first problem found in xs.
func ValidateRefundHolds(xs []RefundHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundClaimSchema.
func (x RefundClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundclaim: %w", err)
	}
	if err := validate.Record(RefundClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundClaims returns the first problem found in xs.
func ValidateRefundClaims(xs []RefundClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against RefundCursorSchema.
func (x RefundCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("refund: refundcursor: %w", err)
	}
	if err := validate.Record(RefundCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("refund: refundcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateRefundCursors returns the first problem found in xs.
func ValidateRefundCursors(xs []RefundCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
