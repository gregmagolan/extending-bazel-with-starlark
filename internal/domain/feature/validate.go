package feature

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded feature records against this package's schemas.
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
		return fmt.Errorf("feature: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against FeatureRecordSchema.
func (x FeatureRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurerecord: %w", err)
	}
	if err := validate.Record(FeatureRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureRecords returns the first problem found in xs.
func ValidateFeatureRecords(xs []FeatureRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureEntrySchema.
func (x FeatureEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featureentry: %w", err)
	}
	if err := validate.Record(FeatureEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featureentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureEntrys returns the first problem found in xs.
func ValidateFeatureEntrys(xs []FeatureEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureItemSchema.
func (x FeatureItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featureitem: %w", err)
	}
	if err := validate.Record(FeatureItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featureitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureItems returns the first problem found in xs.
func ValidateFeatureItems(xs []FeatureItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureDraftSchema.
func (x FeatureDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featuredraft: %w", err)
	}
	if err := validate.Record(FeatureDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featuredraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureDrafts returns the first problem found in xs.
func ValidateFeatureDrafts(xs []FeatureDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureRevisionSchema.
func (x FeatureRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurerevision: %w", err)
	}
	if err := validate.Record(FeatureRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureRevisions returns the first problem found in xs.
func ValidateFeatureRevisions(xs []FeatureRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureSnapshotSchema.
func (x FeatureSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featuresnapshot: %w", err)
	}
	if err := validate.Record(FeatureSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featuresnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureSnapshots returns the first problem found in xs.
func ValidateFeatureSnapshots(xs []FeatureSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureBatchSchema.
func (x FeatureBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurebatch: %w", err)
	}
	if err := validate.Record(FeatureBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureBatchs returns the first problem found in xs.
func ValidateFeatureBatchs(xs []FeatureBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureRequestSchema.
func (x FeatureRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurerequest: %w", err)
	}
	if err := validate.Record(FeatureRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureRequests returns the first problem found in xs.
func ValidateFeatureRequests(xs []FeatureRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureSummarySchema.
func (x FeatureSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featuresummary: %w", err)
	}
	if err := validate.Record(FeatureSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featuresummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureSummarys returns the first problem found in xs.
func ValidateFeatureSummarys(xs []FeatureSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeaturePolicySchema.
func (x FeaturePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurepolicy: %w", err)
	}
	if err := validate.Record(FeaturePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeaturePolicys returns the first problem found in xs.
func ValidateFeaturePolicys(xs []FeaturePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureLinkSchema.
func (x FeatureLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurelink: %w", err)
	}
	if err := validate.Record(FeatureLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureLinks returns the first problem found in xs.
func ValidateFeatureLinks(xs []FeatureLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureNoteSchema.
func (x FeatureNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurenote: %w", err)
	}
	if err := validate.Record(FeatureNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureNotes returns the first problem found in xs.
func ValidateFeatureNotes(xs []FeatureNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureWindowSchema.
func (x FeatureWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurewindow: %w", err)
	}
	if err := validate.Record(FeatureWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureWindows returns the first problem found in xs.
func ValidateFeatureWindows(xs []FeatureWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureQuotaSchema.
func (x FeatureQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurequota: %w", err)
	}
	if err := validate.Record(FeatureQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureQuotas returns the first problem found in xs.
func ValidateFeatureQuotas(xs []FeatureQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureBucketSchema.
func (x FeatureBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurebucket: %w", err)
	}
	if err := validate.Record(FeatureBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureBuckets returns the first problem found in xs.
func ValidateFeatureBuckets(xs []FeatureBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureLaneSchema.
func (x FeatureLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurelane: %w", err)
	}
	if err := validate.Record(FeatureLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureLanes returns the first problem found in xs.
func ValidateFeatureLanes(xs []FeatureLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureSlotSchema.
func (x FeatureSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featureslot: %w", err)
	}
	if err := validate.Record(FeatureSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featureslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureSlots returns the first problem found in xs.
func ValidateFeatureSlots(xs []FeatureSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureTagSchema.
func (x FeatureTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featuretag: %w", err)
	}
	if err := validate.Record(FeatureTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featuretag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureTags returns the first problem found in xs.
func ValidateFeatureTags(xs []FeatureTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureRuleSchema.
func (x FeatureRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurerule: %w", err)
	}
	if err := validate.Record(FeatureRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureRules returns the first problem found in xs.
func ValidateFeatureRules(xs []FeatureRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureHoldSchema.
func (x FeatureHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurehold: %w", err)
	}
	if err := validate.Record(FeatureHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureHolds returns the first problem found in xs.
func ValidateFeatureHolds(xs []FeatureHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureClaimSchema.
func (x FeatureClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featureclaim: %w", err)
	}
	if err := validate.Record(FeatureClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featureclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureClaims returns the first problem found in xs.
func ValidateFeatureClaims(xs []FeatureClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against FeatureCursorSchema.
func (x FeatureCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("feature: featurecursor: %w", err)
	}
	if err := validate.Record(FeatureCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("feature: featurecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateFeatureCursors returns the first problem found in xs.
func ValidateFeatureCursors(xs []FeatureCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
