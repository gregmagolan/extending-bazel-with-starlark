package catalog

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded catalog records against this package's schemas.
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
		return fmt.Errorf("catalog: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against CatalogRecordSchema.
func (x CatalogRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogrecord: %w", err)
	}
	if err := validate.Record(CatalogRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogRecords returns the first problem found in xs.
func ValidateCatalogRecords(xs []CatalogRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogEntrySchema.
func (x CatalogEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogentry: %w", err)
	}
	if err := validate.Record(CatalogEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogEntrys returns the first problem found in xs.
func ValidateCatalogEntrys(xs []CatalogEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogItemSchema.
func (x CatalogItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogitem: %w", err)
	}
	if err := validate.Record(CatalogItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogItems returns the first problem found in xs.
func ValidateCatalogItems(xs []CatalogItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogDraftSchema.
func (x CatalogDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogdraft: %w", err)
	}
	if err := validate.Record(CatalogDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogDrafts returns the first problem found in xs.
func ValidateCatalogDrafts(xs []CatalogDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogRevisionSchema.
func (x CatalogRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogrevision: %w", err)
	}
	if err := validate.Record(CatalogRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogRevisions returns the first problem found in xs.
func ValidateCatalogRevisions(xs []CatalogRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogSnapshotSchema.
func (x CatalogSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogsnapshot: %w", err)
	}
	if err := validate.Record(CatalogSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogSnapshots returns the first problem found in xs.
func ValidateCatalogSnapshots(xs []CatalogSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogBatchSchema.
func (x CatalogBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogbatch: %w", err)
	}
	if err := validate.Record(CatalogBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogBatchs returns the first problem found in xs.
func ValidateCatalogBatchs(xs []CatalogBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogRequestSchema.
func (x CatalogRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogrequest: %w", err)
	}
	if err := validate.Record(CatalogRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogRequests returns the first problem found in xs.
func ValidateCatalogRequests(xs []CatalogRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogSummarySchema.
func (x CatalogSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogsummary: %w", err)
	}
	if err := validate.Record(CatalogSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogSummarys returns the first problem found in xs.
func ValidateCatalogSummarys(xs []CatalogSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogPolicySchema.
func (x CatalogPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogpolicy: %w", err)
	}
	if err := validate.Record(CatalogPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogPolicys returns the first problem found in xs.
func ValidateCatalogPolicys(xs []CatalogPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogLinkSchema.
func (x CatalogLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: cataloglink: %w", err)
	}
	if err := validate.Record(CatalogLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: cataloglink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogLinks returns the first problem found in xs.
func ValidateCatalogLinks(xs []CatalogLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogNoteSchema.
func (x CatalogNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalognote: %w", err)
	}
	if err := validate.Record(CatalogNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalognote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogNotes returns the first problem found in xs.
func ValidateCatalogNotes(xs []CatalogNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogWindowSchema.
func (x CatalogWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogwindow: %w", err)
	}
	if err := validate.Record(CatalogWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogWindows returns the first problem found in xs.
func ValidateCatalogWindows(xs []CatalogWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogQuotaSchema.
func (x CatalogQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogquota: %w", err)
	}
	if err := validate.Record(CatalogQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogQuotas returns the first problem found in xs.
func ValidateCatalogQuotas(xs []CatalogQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogBucketSchema.
func (x CatalogBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogbucket: %w", err)
	}
	if err := validate.Record(CatalogBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogBuckets returns the first problem found in xs.
func ValidateCatalogBuckets(xs []CatalogBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogLaneSchema.
func (x CatalogLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: cataloglane: %w", err)
	}
	if err := validate.Record(CatalogLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: cataloglane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogLanes returns the first problem found in xs.
func ValidateCatalogLanes(xs []CatalogLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogSlotSchema.
func (x CatalogSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogslot: %w", err)
	}
	if err := validate.Record(CatalogSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogSlots returns the first problem found in xs.
func ValidateCatalogSlots(xs []CatalogSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogTagSchema.
func (x CatalogTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogtag: %w", err)
	}
	if err := validate.Record(CatalogTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogTags returns the first problem found in xs.
func ValidateCatalogTags(xs []CatalogTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogRuleSchema.
func (x CatalogRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogrule: %w", err)
	}
	if err := validate.Record(CatalogRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogRules returns the first problem found in xs.
func ValidateCatalogRules(xs []CatalogRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogHoldSchema.
func (x CatalogHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: cataloghold: %w", err)
	}
	if err := validate.Record(CatalogHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: cataloghold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogHolds returns the first problem found in xs.
func ValidateCatalogHolds(xs []CatalogHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogClaimSchema.
func (x CatalogClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogclaim: %w", err)
	}
	if err := validate.Record(CatalogClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogClaims returns the first problem found in xs.
func ValidateCatalogClaims(xs []CatalogClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against CatalogCursorSchema.
func (x CatalogCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("catalog: catalogcursor: %w", err)
	}
	if err := validate.Record(CatalogCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("catalog: catalogcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateCatalogCursors returns the first problem found in xs.
func ValidateCatalogCursors(xs []CatalogCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
