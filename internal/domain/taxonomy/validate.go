package taxonomy

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded taxonomy records against this package's schemas.
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
		return fmt.Errorf("taxonomy: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against TaxonomyRecordSchema.
func (x TaxonomyRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrecord: %w", err)
	}
	if err := validate.Record(TaxonomyRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyRecords returns the first problem found in xs.
func ValidateTaxonomyRecords(xs []TaxonomyRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyEntrySchema.
func (x TaxonomyEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyentry: %w", err)
	}
	if err := validate.Record(TaxonomyEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyEntrys returns the first problem found in xs.
func ValidateTaxonomyEntrys(xs []TaxonomyEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyItemSchema.
func (x TaxonomyItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyitem: %w", err)
	}
	if err := validate.Record(TaxonomyItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyItems returns the first problem found in xs.
func ValidateTaxonomyItems(xs []TaxonomyItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyDraftSchema.
func (x TaxonomyDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomydraft: %w", err)
	}
	if err := validate.Record(TaxonomyDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomydraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyDrafts returns the first problem found in xs.
func ValidateTaxonomyDrafts(xs []TaxonomyDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyRevisionSchema.
func (x TaxonomyRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrevision: %w", err)
	}
	if err := validate.Record(TaxonomyRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyRevisions returns the first problem found in xs.
func ValidateTaxonomyRevisions(xs []TaxonomyRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomySnapshotSchema.
func (x TaxonomySnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomysnapshot: %w", err)
	}
	if err := validate.Record(TaxonomySnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomysnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomySnapshots returns the first problem found in xs.
func ValidateTaxonomySnapshots(xs []TaxonomySnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyBatchSchema.
func (x TaxonomyBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomybatch: %w", err)
	}
	if err := validate.Record(TaxonomyBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomybatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyBatchs returns the first problem found in xs.
func ValidateTaxonomyBatchs(xs []TaxonomyBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyRequestSchema.
func (x TaxonomyRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrequest: %w", err)
	}
	if err := validate.Record(TaxonomyRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyRequests returns the first problem found in xs.
func ValidateTaxonomyRequests(xs []TaxonomyRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomySummarySchema.
func (x TaxonomySummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomysummary: %w", err)
	}
	if err := validate.Record(TaxonomySummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomysummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomySummarys returns the first problem found in xs.
func ValidateTaxonomySummarys(xs []TaxonomySummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyPolicySchema.
func (x TaxonomyPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomypolicy: %w", err)
	}
	if err := validate.Record(TaxonomyPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomypolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyPolicys returns the first problem found in xs.
func ValidateTaxonomyPolicys(xs []TaxonomyPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyLinkSchema.
func (x TaxonomyLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomylink: %w", err)
	}
	if err := validate.Record(TaxonomyLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomylink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyLinks returns the first problem found in xs.
func ValidateTaxonomyLinks(xs []TaxonomyLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyNoteSchema.
func (x TaxonomyNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomynote: %w", err)
	}
	if err := validate.Record(TaxonomyNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomynote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyNotes returns the first problem found in xs.
func ValidateTaxonomyNotes(xs []TaxonomyNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyWindowSchema.
func (x TaxonomyWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomywindow: %w", err)
	}
	if err := validate.Record(TaxonomyWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomywindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyWindows returns the first problem found in xs.
func ValidateTaxonomyWindows(xs []TaxonomyWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyQuotaSchema.
func (x TaxonomyQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyquota: %w", err)
	}
	if err := validate.Record(TaxonomyQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyQuotas returns the first problem found in xs.
func ValidateTaxonomyQuotas(xs []TaxonomyQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyBucketSchema.
func (x TaxonomyBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomybucket: %w", err)
	}
	if err := validate.Record(TaxonomyBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomybucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyBuckets returns the first problem found in xs.
func ValidateTaxonomyBuckets(xs []TaxonomyBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyLaneSchema.
func (x TaxonomyLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomylane: %w", err)
	}
	if err := validate.Record(TaxonomyLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomylane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyLanes returns the first problem found in xs.
func ValidateTaxonomyLanes(xs []TaxonomyLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomySlotSchema.
func (x TaxonomySlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyslot: %w", err)
	}
	if err := validate.Record(TaxonomySlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomySlots returns the first problem found in xs.
func ValidateTaxonomySlots(xs []TaxonomySlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyTagSchema.
func (x TaxonomyTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomytag: %w", err)
	}
	if err := validate.Record(TaxonomyTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomytag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyTags returns the first problem found in xs.
func ValidateTaxonomyTags(xs []TaxonomyTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyRuleSchema.
func (x TaxonomyRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrule: %w", err)
	}
	if err := validate.Record(TaxonomyRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyRules returns the first problem found in xs.
func ValidateTaxonomyRules(xs []TaxonomyRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyHoldSchema.
func (x TaxonomyHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyhold: %w", err)
	}
	if err := validate.Record(TaxonomyHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyHolds returns the first problem found in xs.
func ValidateTaxonomyHolds(xs []TaxonomyHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyClaimSchema.
func (x TaxonomyClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomyclaim: %w", err)
	}
	if err := validate.Record(TaxonomyClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomyclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyClaims returns the first problem found in xs.
func ValidateTaxonomyClaims(xs []TaxonomyClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TaxonomyCursorSchema.
func (x TaxonomyCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("taxonomy: taxonomycursor: %w", err)
	}
	if err := validate.Record(TaxonomyCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("taxonomy: taxonomycursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTaxonomyCursors returns the first problem found in xs.
func ValidateTaxonomyCursors(xs []TaxonomyCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
