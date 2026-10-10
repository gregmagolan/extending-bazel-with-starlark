package usage

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded usage records against this package's schemas.
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
		return fmt.Errorf("usage: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against UsageRecordSchema.
func (x UsageRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagerecord: %w", err)
	}
	if err := validate.Record(UsageRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageRecords returns the first problem found in xs.
func ValidateUsageRecords(xs []UsageRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageEntrySchema.
func (x UsageEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usageentry: %w", err)
	}
	if err := validate.Record(UsageEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usageentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageEntrys returns the first problem found in xs.
func ValidateUsageEntrys(xs []UsageEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageItemSchema.
func (x UsageItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usageitem: %w", err)
	}
	if err := validate.Record(UsageItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usageitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageItems returns the first problem found in xs.
func ValidateUsageItems(xs []UsageItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageDraftSchema.
func (x UsageDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagedraft: %w", err)
	}
	if err := validate.Record(UsageDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageDrafts returns the first problem found in xs.
func ValidateUsageDrafts(xs []UsageDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageRevisionSchema.
func (x UsageRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagerevision: %w", err)
	}
	if err := validate.Record(UsageRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageRevisions returns the first problem found in xs.
func ValidateUsageRevisions(xs []UsageRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageSnapshotSchema.
func (x UsageSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagesnapshot: %w", err)
	}
	if err := validate.Record(UsageSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageSnapshots returns the first problem found in xs.
func ValidateUsageSnapshots(xs []UsageSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageBatchSchema.
func (x UsageBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagebatch: %w", err)
	}
	if err := validate.Record(UsageBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageBatchs returns the first problem found in xs.
func ValidateUsageBatchs(xs []UsageBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageRequestSchema.
func (x UsageRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagerequest: %w", err)
	}
	if err := validate.Record(UsageRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageRequests returns the first problem found in xs.
func ValidateUsageRequests(xs []UsageRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageSummarySchema.
func (x UsageSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagesummary: %w", err)
	}
	if err := validate.Record(UsageSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageSummarys returns the first problem found in xs.
func ValidateUsageSummarys(xs []UsageSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsagePolicySchema.
func (x UsagePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagepolicy: %w", err)
	}
	if err := validate.Record(UsagePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsagePolicys returns the first problem found in xs.
func ValidateUsagePolicys(xs []UsagePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageLinkSchema.
func (x UsageLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagelink: %w", err)
	}
	if err := validate.Record(UsageLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageLinks returns the first problem found in xs.
func ValidateUsageLinks(xs []UsageLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageNoteSchema.
func (x UsageNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagenote: %w", err)
	}
	if err := validate.Record(UsageNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageNotes returns the first problem found in xs.
func ValidateUsageNotes(xs []UsageNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageWindowSchema.
func (x UsageWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagewindow: %w", err)
	}
	if err := validate.Record(UsageWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageWindows returns the first problem found in xs.
func ValidateUsageWindows(xs []UsageWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageQuotaSchema.
func (x UsageQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagequota: %w", err)
	}
	if err := validate.Record(UsageQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageQuotas returns the first problem found in xs.
func ValidateUsageQuotas(xs []UsageQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageBucketSchema.
func (x UsageBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagebucket: %w", err)
	}
	if err := validate.Record(UsageBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageBuckets returns the first problem found in xs.
func ValidateUsageBuckets(xs []UsageBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageLaneSchema.
func (x UsageLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagelane: %w", err)
	}
	if err := validate.Record(UsageLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageLanes returns the first problem found in xs.
func ValidateUsageLanes(xs []UsageLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageSlotSchema.
func (x UsageSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usageslot: %w", err)
	}
	if err := validate.Record(UsageSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usageslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageSlots returns the first problem found in xs.
func ValidateUsageSlots(xs []UsageSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageTagSchema.
func (x UsageTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagetag: %w", err)
	}
	if err := validate.Record(UsageTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageTags returns the first problem found in xs.
func ValidateUsageTags(xs []UsageTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageRuleSchema.
func (x UsageRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagerule: %w", err)
	}
	if err := validate.Record(UsageRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageRules returns the first problem found in xs.
func ValidateUsageRules(xs []UsageRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageHoldSchema.
func (x UsageHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagehold: %w", err)
	}
	if err := validate.Record(UsageHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageHolds returns the first problem found in xs.
func ValidateUsageHolds(xs []UsageHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageClaimSchema.
func (x UsageClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usageclaim: %w", err)
	}
	if err := validate.Record(UsageClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usageclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageClaims returns the first problem found in xs.
func ValidateUsageClaims(xs []UsageClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UsageCursorSchema.
func (x UsageCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("usage: usagecursor: %w", err)
	}
	if err := validate.Record(UsageCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("usage: usagecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUsageCursors returns the first problem found in xs.
func ValidateUsageCursors(xs []UsageCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
