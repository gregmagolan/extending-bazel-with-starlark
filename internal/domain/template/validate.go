package template

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded template records against this package's schemas.
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
		return fmt.Errorf("template: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against TemplateRecordSchema.
func (x TemplateRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templaterecord: %w", err)
	}
	if err := validate.Record(TemplateRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templaterecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateRecords returns the first problem found in xs.
func ValidateTemplateRecords(xs []TemplateRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateEntrySchema.
func (x TemplateEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templateentry: %w", err)
	}
	if err := validate.Record(TemplateEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templateentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateEntrys returns the first problem found in xs.
func ValidateTemplateEntrys(xs []TemplateEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateItemSchema.
func (x TemplateItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templateitem: %w", err)
	}
	if err := validate.Record(TemplateItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templateitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateItems returns the first problem found in xs.
func ValidateTemplateItems(xs []TemplateItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateDraftSchema.
func (x TemplateDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatedraft: %w", err)
	}
	if err := validate.Record(TemplateDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateDrafts returns the first problem found in xs.
func ValidateTemplateDrafts(xs []TemplateDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateRevisionSchema.
func (x TemplateRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templaterevision: %w", err)
	}
	if err := validate.Record(TemplateRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templaterevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateRevisions returns the first problem found in xs.
func ValidateTemplateRevisions(xs []TemplateRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateSnapshotSchema.
func (x TemplateSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatesnapshot: %w", err)
	}
	if err := validate.Record(TemplateSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateSnapshots returns the first problem found in xs.
func ValidateTemplateSnapshots(xs []TemplateSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateBatchSchema.
func (x TemplateBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatebatch: %w", err)
	}
	if err := validate.Record(TemplateBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateBatchs returns the first problem found in xs.
func ValidateTemplateBatchs(xs []TemplateBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateRequestSchema.
func (x TemplateRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templaterequest: %w", err)
	}
	if err := validate.Record(TemplateRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templaterequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateRequests returns the first problem found in xs.
func ValidateTemplateRequests(xs []TemplateRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateSummarySchema.
func (x TemplateSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatesummary: %w", err)
	}
	if err := validate.Record(TemplateSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateSummarys returns the first problem found in xs.
func ValidateTemplateSummarys(xs []TemplateSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplatePolicySchema.
func (x TemplatePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatepolicy: %w", err)
	}
	if err := validate.Record(TemplatePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplatePolicys returns the first problem found in xs.
func ValidateTemplatePolicys(xs []TemplatePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateLinkSchema.
func (x TemplateLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatelink: %w", err)
	}
	if err := validate.Record(TemplateLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateLinks returns the first problem found in xs.
func ValidateTemplateLinks(xs []TemplateLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateNoteSchema.
func (x TemplateNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatenote: %w", err)
	}
	if err := validate.Record(TemplateNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateNotes returns the first problem found in xs.
func ValidateTemplateNotes(xs []TemplateNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateWindowSchema.
func (x TemplateWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatewindow: %w", err)
	}
	if err := validate.Record(TemplateWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateWindows returns the first problem found in xs.
func ValidateTemplateWindows(xs []TemplateWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateQuotaSchema.
func (x TemplateQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatequota: %w", err)
	}
	if err := validate.Record(TemplateQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateQuotas returns the first problem found in xs.
func ValidateTemplateQuotas(xs []TemplateQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateBucketSchema.
func (x TemplateBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatebucket: %w", err)
	}
	if err := validate.Record(TemplateBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateBuckets returns the first problem found in xs.
func ValidateTemplateBuckets(xs []TemplateBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateLaneSchema.
func (x TemplateLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatelane: %w", err)
	}
	if err := validate.Record(TemplateLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateLanes returns the first problem found in xs.
func ValidateTemplateLanes(xs []TemplateLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateSlotSchema.
func (x TemplateSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templateslot: %w", err)
	}
	if err := validate.Record(TemplateSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templateslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateSlots returns the first problem found in xs.
func ValidateTemplateSlots(xs []TemplateSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateTagSchema.
func (x TemplateTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatetag: %w", err)
	}
	if err := validate.Record(TemplateTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateTags returns the first problem found in xs.
func ValidateTemplateTags(xs []TemplateTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateRuleSchema.
func (x TemplateRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templaterule: %w", err)
	}
	if err := validate.Record(TemplateRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templaterule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateRules returns the first problem found in xs.
func ValidateTemplateRules(xs []TemplateRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateHoldSchema.
func (x TemplateHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatehold: %w", err)
	}
	if err := validate.Record(TemplateHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateHolds returns the first problem found in xs.
func ValidateTemplateHolds(xs []TemplateHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateClaimSchema.
func (x TemplateClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templateclaim: %w", err)
	}
	if err := validate.Record(TemplateClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templateclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateClaims returns the first problem found in xs.
func ValidateTemplateClaims(xs []TemplateClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TemplateCursorSchema.
func (x TemplateCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("template: templatecursor: %w", err)
	}
	if err := validate.Record(TemplateCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("template: templatecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTemplateCursors returns the first problem found in xs.
func ValidateTemplateCursors(xs []TemplateCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
