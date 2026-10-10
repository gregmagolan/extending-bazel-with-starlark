package segment

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded segment records against this package's schemas.
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
		return fmt.Errorf("segment: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against SegmentRecordSchema.
func (x SegmentRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentrecord: %w", err)
	}
	if err := validate.Record(SegmentRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentRecords returns the first problem found in xs.
func ValidateSegmentRecords(xs []SegmentRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentEntrySchema.
func (x SegmentEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmententry: %w", err)
	}
	if err := validate.Record(SegmentEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmententry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentEntrys returns the first problem found in xs.
func ValidateSegmentEntrys(xs []SegmentEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentItemSchema.
func (x SegmentItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentitem: %w", err)
	}
	if err := validate.Record(SegmentItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentItems returns the first problem found in xs.
func ValidateSegmentItems(xs []SegmentItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentDraftSchema.
func (x SegmentDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentdraft: %w", err)
	}
	if err := validate.Record(SegmentDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentDrafts returns the first problem found in xs.
func ValidateSegmentDrafts(xs []SegmentDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentRevisionSchema.
func (x SegmentRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentrevision: %w", err)
	}
	if err := validate.Record(SegmentRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentRevisions returns the first problem found in xs.
func ValidateSegmentRevisions(xs []SegmentRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentSnapshotSchema.
func (x SegmentSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentsnapshot: %w", err)
	}
	if err := validate.Record(SegmentSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentSnapshots returns the first problem found in xs.
func ValidateSegmentSnapshots(xs []SegmentSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentBatchSchema.
func (x SegmentBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentbatch: %w", err)
	}
	if err := validate.Record(SegmentBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentBatchs returns the first problem found in xs.
func ValidateSegmentBatchs(xs []SegmentBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentRequestSchema.
func (x SegmentRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentrequest: %w", err)
	}
	if err := validate.Record(SegmentRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentRequests returns the first problem found in xs.
func ValidateSegmentRequests(xs []SegmentRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentSummarySchema.
func (x SegmentSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentsummary: %w", err)
	}
	if err := validate.Record(SegmentSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentSummarys returns the first problem found in xs.
func ValidateSegmentSummarys(xs []SegmentSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentPolicySchema.
func (x SegmentPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentpolicy: %w", err)
	}
	if err := validate.Record(SegmentPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentPolicys returns the first problem found in xs.
func ValidateSegmentPolicys(xs []SegmentPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentLinkSchema.
func (x SegmentLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentlink: %w", err)
	}
	if err := validate.Record(SegmentLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentLinks returns the first problem found in xs.
func ValidateSegmentLinks(xs []SegmentLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentNoteSchema.
func (x SegmentNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentnote: %w", err)
	}
	if err := validate.Record(SegmentNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentNotes returns the first problem found in xs.
func ValidateSegmentNotes(xs []SegmentNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentWindowSchema.
func (x SegmentWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentwindow: %w", err)
	}
	if err := validate.Record(SegmentWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentWindows returns the first problem found in xs.
func ValidateSegmentWindows(xs []SegmentWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentQuotaSchema.
func (x SegmentQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentquota: %w", err)
	}
	if err := validate.Record(SegmentQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentQuotas returns the first problem found in xs.
func ValidateSegmentQuotas(xs []SegmentQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentBucketSchema.
func (x SegmentBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentbucket: %w", err)
	}
	if err := validate.Record(SegmentBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentBuckets returns the first problem found in xs.
func ValidateSegmentBuckets(xs []SegmentBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentLaneSchema.
func (x SegmentLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentlane: %w", err)
	}
	if err := validate.Record(SegmentLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentLanes returns the first problem found in xs.
func ValidateSegmentLanes(xs []SegmentLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentSlotSchema.
func (x SegmentSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentslot: %w", err)
	}
	if err := validate.Record(SegmentSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentSlots returns the first problem found in xs.
func ValidateSegmentSlots(xs []SegmentSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentTagSchema.
func (x SegmentTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmenttag: %w", err)
	}
	if err := validate.Record(SegmentTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmenttag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentTags returns the first problem found in xs.
func ValidateSegmentTags(xs []SegmentTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentRuleSchema.
func (x SegmentRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentrule: %w", err)
	}
	if err := validate.Record(SegmentRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentRules returns the first problem found in xs.
func ValidateSegmentRules(xs []SegmentRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentHoldSchema.
func (x SegmentHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmenthold: %w", err)
	}
	if err := validate.Record(SegmentHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmenthold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentHolds returns the first problem found in xs.
func ValidateSegmentHolds(xs []SegmentHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentClaimSchema.
func (x SegmentClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentclaim: %w", err)
	}
	if err := validate.Record(SegmentClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentClaims returns the first problem found in xs.
func ValidateSegmentClaims(xs []SegmentClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SegmentCursorSchema.
func (x SegmentCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("segment: segmentcursor: %w", err)
	}
	if err := validate.Record(SegmentCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("segment: segmentcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSegmentCursors returns the first problem found in xs.
func ValidateSegmentCursors(xs []SegmentCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
