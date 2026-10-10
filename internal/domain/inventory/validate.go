package inventory

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded inventory records against this package's schemas.
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
		return fmt.Errorf("inventory: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against InventoryRecordSchema.
func (x InventoryRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryrecord: %w", err)
	}
	if err := validate.Record(InventoryRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryRecords returns the first problem found in xs.
func ValidateInventoryRecords(xs []InventoryRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryEntrySchema.
func (x InventoryEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryentry: %w", err)
	}
	if err := validate.Record(InventoryEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryEntrys returns the first problem found in xs.
func ValidateInventoryEntrys(xs []InventoryEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryItemSchema.
func (x InventoryItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryitem: %w", err)
	}
	if err := validate.Record(InventoryItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryItems returns the first problem found in xs.
func ValidateInventoryItems(xs []InventoryItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryDraftSchema.
func (x InventoryDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorydraft: %w", err)
	}
	if err := validate.Record(InventoryDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorydraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryDrafts returns the first problem found in xs.
func ValidateInventoryDrafts(xs []InventoryDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryRevisionSchema.
func (x InventoryRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryrevision: %w", err)
	}
	if err := validate.Record(InventoryRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryRevisions returns the first problem found in xs.
func ValidateInventoryRevisions(xs []InventoryRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventorySnapshotSchema.
func (x InventorySnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorysnapshot: %w", err)
	}
	if err := validate.Record(InventorySnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorysnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventorySnapshots returns the first problem found in xs.
func ValidateInventorySnapshots(xs []InventorySnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryBatchSchema.
func (x InventoryBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorybatch: %w", err)
	}
	if err := validate.Record(InventoryBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorybatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryBatchs returns the first problem found in xs.
func ValidateInventoryBatchs(xs []InventoryBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryRequestSchema.
func (x InventoryRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryrequest: %w", err)
	}
	if err := validate.Record(InventoryRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryRequests returns the first problem found in xs.
func ValidateInventoryRequests(xs []InventoryRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventorySummarySchema.
func (x InventorySummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorysummary: %w", err)
	}
	if err := validate.Record(InventorySummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorysummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventorySummarys returns the first problem found in xs.
func ValidateInventorySummarys(xs []InventorySummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryPolicySchema.
func (x InventoryPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorypolicy: %w", err)
	}
	if err := validate.Record(InventoryPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorypolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryPolicys returns the first problem found in xs.
func ValidateInventoryPolicys(xs []InventoryPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryLinkSchema.
func (x InventoryLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorylink: %w", err)
	}
	if err := validate.Record(InventoryLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorylink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryLinks returns the first problem found in xs.
func ValidateInventoryLinks(xs []InventoryLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryNoteSchema.
func (x InventoryNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorynote: %w", err)
	}
	if err := validate.Record(InventoryNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorynote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryNotes returns the first problem found in xs.
func ValidateInventoryNotes(xs []InventoryNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryWindowSchema.
func (x InventoryWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorywindow: %w", err)
	}
	if err := validate.Record(InventoryWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorywindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryWindows returns the first problem found in xs.
func ValidateInventoryWindows(xs []InventoryWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryQuotaSchema.
func (x InventoryQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryquota: %w", err)
	}
	if err := validate.Record(InventoryQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryQuotas returns the first problem found in xs.
func ValidateInventoryQuotas(xs []InventoryQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryBucketSchema.
func (x InventoryBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorybucket: %w", err)
	}
	if err := validate.Record(InventoryBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorybucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryBuckets returns the first problem found in xs.
func ValidateInventoryBuckets(xs []InventoryBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryLaneSchema.
func (x InventoryLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorylane: %w", err)
	}
	if err := validate.Record(InventoryLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorylane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryLanes returns the first problem found in xs.
func ValidateInventoryLanes(xs []InventoryLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventorySlotSchema.
func (x InventorySlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryslot: %w", err)
	}
	if err := validate.Record(InventorySlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventorySlots returns the first problem found in xs.
func ValidateInventorySlots(xs []InventorySlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryTagSchema.
func (x InventoryTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorytag: %w", err)
	}
	if err := validate.Record(InventoryTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorytag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryTags returns the first problem found in xs.
func ValidateInventoryTags(xs []InventoryTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryRuleSchema.
func (x InventoryRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryrule: %w", err)
	}
	if err := validate.Record(InventoryRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryRules returns the first problem found in xs.
func ValidateInventoryRules(xs []InventoryRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryHoldSchema.
func (x InventoryHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryhold: %w", err)
	}
	if err := validate.Record(InventoryHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryHolds returns the first problem found in xs.
func ValidateInventoryHolds(xs []InventoryHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryClaimSchema.
func (x InventoryClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventoryclaim: %w", err)
	}
	if err := validate.Record(InventoryClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventoryclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryClaims returns the first problem found in xs.
func ValidateInventoryClaims(xs []InventoryClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InventoryCursorSchema.
func (x InventoryCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("inventory: inventorycursor: %w", err)
	}
	if err := validate.Record(InventoryCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("inventory: inventorycursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInventoryCursors returns the first problem found in xs.
func ValidateInventoryCursors(xs []InventoryCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
