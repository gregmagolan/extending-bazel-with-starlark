package payment

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded payment records against this package's schemas.
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
		return fmt.Errorf("payment: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against PaymentRecordSchema.
func (x PaymentRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentrecord: %w", err)
	}
	if err := validate.Record(PaymentRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentRecords returns the first problem found in xs.
func ValidatePaymentRecords(xs []PaymentRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentEntrySchema.
func (x PaymentEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymententry: %w", err)
	}
	if err := validate.Record(PaymentEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymententry %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentEntrys returns the first problem found in xs.
func ValidatePaymentEntrys(xs []PaymentEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentItemSchema.
func (x PaymentItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentitem: %w", err)
	}
	if err := validate.Record(PaymentItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentItems returns the first problem found in xs.
func ValidatePaymentItems(xs []PaymentItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentDraftSchema.
func (x PaymentDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentdraft: %w", err)
	}
	if err := validate.Record(PaymentDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentDrafts returns the first problem found in xs.
func ValidatePaymentDrafts(xs []PaymentDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentRevisionSchema.
func (x PaymentRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentrevision: %w", err)
	}
	if err := validate.Record(PaymentRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentRevisions returns the first problem found in xs.
func ValidatePaymentRevisions(xs []PaymentRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentSnapshotSchema.
func (x PaymentSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentsnapshot: %w", err)
	}
	if err := validate.Record(PaymentSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentSnapshots returns the first problem found in xs.
func ValidatePaymentSnapshots(xs []PaymentSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentBatchSchema.
func (x PaymentBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentbatch: %w", err)
	}
	if err := validate.Record(PaymentBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentBatchs returns the first problem found in xs.
func ValidatePaymentBatchs(xs []PaymentBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentRequestSchema.
func (x PaymentRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentrequest: %w", err)
	}
	if err := validate.Record(PaymentRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentRequests returns the first problem found in xs.
func ValidatePaymentRequests(xs []PaymentRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentSummarySchema.
func (x PaymentSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentsummary: %w", err)
	}
	if err := validate.Record(PaymentSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentSummarys returns the first problem found in xs.
func ValidatePaymentSummarys(xs []PaymentSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentPolicySchema.
func (x PaymentPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentpolicy: %w", err)
	}
	if err := validate.Record(PaymentPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentPolicys returns the first problem found in xs.
func ValidatePaymentPolicys(xs []PaymentPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentLinkSchema.
func (x PaymentLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentlink: %w", err)
	}
	if err := validate.Record(PaymentLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentLinks returns the first problem found in xs.
func ValidatePaymentLinks(xs []PaymentLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentNoteSchema.
func (x PaymentNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentnote: %w", err)
	}
	if err := validate.Record(PaymentNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentNotes returns the first problem found in xs.
func ValidatePaymentNotes(xs []PaymentNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentWindowSchema.
func (x PaymentWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentwindow: %w", err)
	}
	if err := validate.Record(PaymentWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentWindows returns the first problem found in xs.
func ValidatePaymentWindows(xs []PaymentWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentQuotaSchema.
func (x PaymentQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentquota: %w", err)
	}
	if err := validate.Record(PaymentQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentQuotas returns the first problem found in xs.
func ValidatePaymentQuotas(xs []PaymentQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentBucketSchema.
func (x PaymentBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentbucket: %w", err)
	}
	if err := validate.Record(PaymentBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentBuckets returns the first problem found in xs.
func ValidatePaymentBuckets(xs []PaymentBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentLaneSchema.
func (x PaymentLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentlane: %w", err)
	}
	if err := validate.Record(PaymentLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentLanes returns the first problem found in xs.
func ValidatePaymentLanes(xs []PaymentLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentSlotSchema.
func (x PaymentSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentslot: %w", err)
	}
	if err := validate.Record(PaymentSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentSlots returns the first problem found in xs.
func ValidatePaymentSlots(xs []PaymentSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentTagSchema.
func (x PaymentTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymenttag: %w", err)
	}
	if err := validate.Record(PaymentTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymenttag %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentTags returns the first problem found in xs.
func ValidatePaymentTags(xs []PaymentTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentRuleSchema.
func (x PaymentRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentrule: %w", err)
	}
	if err := validate.Record(PaymentRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentRules returns the first problem found in xs.
func ValidatePaymentRules(xs []PaymentRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentHoldSchema.
func (x PaymentHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymenthold: %w", err)
	}
	if err := validate.Record(PaymentHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymenthold %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentHolds returns the first problem found in xs.
func ValidatePaymentHolds(xs []PaymentHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentClaimSchema.
func (x PaymentClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentclaim: %w", err)
	}
	if err := validate.Record(PaymentClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentClaims returns the first problem found in xs.
func ValidatePaymentClaims(xs []PaymentClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PaymentCursorSchema.
func (x PaymentCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("payment: paymentcursor: %w", err)
	}
	if err := validate.Record(PaymentCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("payment: paymentcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePaymentCursors returns the first problem found in xs.
func ValidatePaymentCursors(xs []PaymentCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
