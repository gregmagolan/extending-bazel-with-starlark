package dunning

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded dunning records against this package's schemas.
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
		return fmt.Errorf("dunning: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against DunningRecordSchema.
func (x DunningRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningrecord: %w", err)
	}
	if err := validate.Record(DunningRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningRecords returns the first problem found in xs.
func ValidateDunningRecords(xs []DunningRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningEntrySchema.
func (x DunningEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningentry: %w", err)
	}
	if err := validate.Record(DunningEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningEntrys returns the first problem found in xs.
func ValidateDunningEntrys(xs []DunningEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningItemSchema.
func (x DunningItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningitem: %w", err)
	}
	if err := validate.Record(DunningItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningItems returns the first problem found in xs.
func ValidateDunningItems(xs []DunningItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningDraftSchema.
func (x DunningDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningdraft: %w", err)
	}
	if err := validate.Record(DunningDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningDrafts returns the first problem found in xs.
func ValidateDunningDrafts(xs []DunningDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningRevisionSchema.
func (x DunningRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningrevision: %w", err)
	}
	if err := validate.Record(DunningRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningRevisions returns the first problem found in xs.
func ValidateDunningRevisions(xs []DunningRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningSnapshotSchema.
func (x DunningSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningsnapshot: %w", err)
	}
	if err := validate.Record(DunningSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningSnapshots returns the first problem found in xs.
func ValidateDunningSnapshots(xs []DunningSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningBatchSchema.
func (x DunningBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningbatch: %w", err)
	}
	if err := validate.Record(DunningBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningBatchs returns the first problem found in xs.
func ValidateDunningBatchs(xs []DunningBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningRequestSchema.
func (x DunningRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningrequest: %w", err)
	}
	if err := validate.Record(DunningRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningRequests returns the first problem found in xs.
func ValidateDunningRequests(xs []DunningRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningSummarySchema.
func (x DunningSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningsummary: %w", err)
	}
	if err := validate.Record(DunningSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningSummarys returns the first problem found in xs.
func ValidateDunningSummarys(xs []DunningSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningPolicySchema.
func (x DunningPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningpolicy: %w", err)
	}
	if err := validate.Record(DunningPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningPolicys returns the first problem found in xs.
func ValidateDunningPolicys(xs []DunningPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningLinkSchema.
func (x DunningLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunninglink: %w", err)
	}
	if err := validate.Record(DunningLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunninglink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningLinks returns the first problem found in xs.
func ValidateDunningLinks(xs []DunningLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningNoteSchema.
func (x DunningNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningnote: %w", err)
	}
	if err := validate.Record(DunningNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningNotes returns the first problem found in xs.
func ValidateDunningNotes(xs []DunningNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningWindowSchema.
func (x DunningWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningwindow: %w", err)
	}
	if err := validate.Record(DunningWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningWindows returns the first problem found in xs.
func ValidateDunningWindows(xs []DunningWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningQuotaSchema.
func (x DunningQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningquota: %w", err)
	}
	if err := validate.Record(DunningQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningQuotas returns the first problem found in xs.
func ValidateDunningQuotas(xs []DunningQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningBucketSchema.
func (x DunningBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningbucket: %w", err)
	}
	if err := validate.Record(DunningBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningBuckets returns the first problem found in xs.
func ValidateDunningBuckets(xs []DunningBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningLaneSchema.
func (x DunningLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunninglane: %w", err)
	}
	if err := validate.Record(DunningLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunninglane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningLanes returns the first problem found in xs.
func ValidateDunningLanes(xs []DunningLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningSlotSchema.
func (x DunningSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningslot: %w", err)
	}
	if err := validate.Record(DunningSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningSlots returns the first problem found in xs.
func ValidateDunningSlots(xs []DunningSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningTagSchema.
func (x DunningTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningtag: %w", err)
	}
	if err := validate.Record(DunningTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningTags returns the first problem found in xs.
func ValidateDunningTags(xs []DunningTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningRuleSchema.
func (x DunningRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningrule: %w", err)
	}
	if err := validate.Record(DunningRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningRules returns the first problem found in xs.
func ValidateDunningRules(xs []DunningRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningHoldSchema.
func (x DunningHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunninghold: %w", err)
	}
	if err := validate.Record(DunningHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunninghold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningHolds returns the first problem found in xs.
func ValidateDunningHolds(xs []DunningHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningClaimSchema.
func (x DunningClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningclaim: %w", err)
	}
	if err := validate.Record(DunningClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningClaims returns the first problem found in xs.
func ValidateDunningClaims(xs []DunningClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DunningCursorSchema.
func (x DunningCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("dunning: dunningcursor: %w", err)
	}
	if err := validate.Record(DunningCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("dunning: dunningcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDunningCursors returns the first problem found in xs.
func ValidateDunningCursors(xs []DunningCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
