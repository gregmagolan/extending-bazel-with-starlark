package warehouse

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded warehouse records against this package's schemas.
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
		return fmt.Errorf("warehouse: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against WarehouseRecordSchema.
func (x WarehouseRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouserecord: %w", err)
	}
	if err := validate.Record(WarehouseRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouserecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseRecords returns the first problem found in xs.
func ValidateWarehouseRecords(xs []WarehouseRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseEntrySchema.
func (x WarehouseEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouseentry: %w", err)
	}
	if err := validate.Record(WarehouseEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouseentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseEntrys returns the first problem found in xs.
func ValidateWarehouseEntrys(xs []WarehouseEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseItemSchema.
func (x WarehouseItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouseitem: %w", err)
	}
	if err := validate.Record(WarehouseItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouseitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseItems returns the first problem found in xs.
func ValidateWarehouseItems(xs []WarehouseItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseDraftSchema.
func (x WarehouseDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousedraft: %w", err)
	}
	if err := validate.Record(WarehouseDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseDrafts returns the first problem found in xs.
func ValidateWarehouseDrafts(xs []WarehouseDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseRevisionSchema.
func (x WarehouseRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouserevision: %w", err)
	}
	if err := validate.Record(WarehouseRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouserevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseRevisions returns the first problem found in xs.
func ValidateWarehouseRevisions(xs []WarehouseRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseSnapshotSchema.
func (x WarehouseSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousesnapshot: %w", err)
	}
	if err := validate.Record(WarehouseSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseSnapshots returns the first problem found in xs.
func ValidateWarehouseSnapshots(xs []WarehouseSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseBatchSchema.
func (x WarehouseBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousebatch: %w", err)
	}
	if err := validate.Record(WarehouseBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseBatchs returns the first problem found in xs.
func ValidateWarehouseBatchs(xs []WarehouseBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseRequestSchema.
func (x WarehouseRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouserequest: %w", err)
	}
	if err := validate.Record(WarehouseRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouserequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseRequests returns the first problem found in xs.
func ValidateWarehouseRequests(xs []WarehouseRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseSummarySchema.
func (x WarehouseSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousesummary: %w", err)
	}
	if err := validate.Record(WarehouseSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseSummarys returns the first problem found in xs.
func ValidateWarehouseSummarys(xs []WarehouseSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehousePolicySchema.
func (x WarehousePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousepolicy: %w", err)
	}
	if err := validate.Record(WarehousePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehousePolicys returns the first problem found in xs.
func ValidateWarehousePolicys(xs []WarehousePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseLinkSchema.
func (x WarehouseLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouselink: %w", err)
	}
	if err := validate.Record(WarehouseLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouselink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseLinks returns the first problem found in xs.
func ValidateWarehouseLinks(xs []WarehouseLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseNoteSchema.
func (x WarehouseNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousenote: %w", err)
	}
	if err := validate.Record(WarehouseNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseNotes returns the first problem found in xs.
func ValidateWarehouseNotes(xs []WarehouseNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseWindowSchema.
func (x WarehouseWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousewindow: %w", err)
	}
	if err := validate.Record(WarehouseWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseWindows returns the first problem found in xs.
func ValidateWarehouseWindows(xs []WarehouseWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseQuotaSchema.
func (x WarehouseQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousequota: %w", err)
	}
	if err := validate.Record(WarehouseQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseQuotas returns the first problem found in xs.
func ValidateWarehouseQuotas(xs []WarehouseQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseBucketSchema.
func (x WarehouseBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousebucket: %w", err)
	}
	if err := validate.Record(WarehouseBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseBuckets returns the first problem found in xs.
func ValidateWarehouseBuckets(xs []WarehouseBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseLaneSchema.
func (x WarehouseLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouselane: %w", err)
	}
	if err := validate.Record(WarehouseLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouselane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseLanes returns the first problem found in xs.
func ValidateWarehouseLanes(xs []WarehouseLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseSlotSchema.
func (x WarehouseSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouseslot: %w", err)
	}
	if err := validate.Record(WarehouseSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouseslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseSlots returns the first problem found in xs.
func ValidateWarehouseSlots(xs []WarehouseSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseTagSchema.
func (x WarehouseTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousetag: %w", err)
	}
	if err := validate.Record(WarehouseTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseTags returns the first problem found in xs.
func ValidateWarehouseTags(xs []WarehouseTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseRuleSchema.
func (x WarehouseRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouserule: %w", err)
	}
	if err := validate.Record(WarehouseRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouserule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseRules returns the first problem found in xs.
func ValidateWarehouseRules(xs []WarehouseRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseHoldSchema.
func (x WarehouseHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousehold: %w", err)
	}
	if err := validate.Record(WarehouseHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseHolds returns the first problem found in xs.
func ValidateWarehouseHolds(xs []WarehouseHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseClaimSchema.
func (x WarehouseClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehouseclaim: %w", err)
	}
	if err := validate.Record(WarehouseClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehouseclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseClaims returns the first problem found in xs.
func ValidateWarehouseClaims(xs []WarehouseClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WarehouseCursorSchema.
func (x WarehouseCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("warehouse: warehousecursor: %w", err)
	}
	if err := validate.Record(WarehouseCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("warehouse: warehousecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWarehouseCursors returns the first problem found in xs.
func ValidateWarehouseCursors(xs []WarehouseCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
