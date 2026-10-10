package stocktake

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded stocktake records against this package's schemas.
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
		return fmt.Errorf("stocktake: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against StocktakeRecordSchema.
func (x StocktakeRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakerecord: %w", err)
	}
	if err := validate.Record(StocktakeRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeRecords returns the first problem found in xs.
func ValidateStocktakeRecords(xs []StocktakeRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeEntrySchema.
func (x StocktakeEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakeentry: %w", err)
	}
	if err := validate.Record(StocktakeEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakeentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeEntrys returns the first problem found in xs.
func ValidateStocktakeEntrys(xs []StocktakeEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeItemSchema.
func (x StocktakeItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakeitem: %w", err)
	}
	if err := validate.Record(StocktakeItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakeitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeItems returns the first problem found in xs.
func ValidateStocktakeItems(xs []StocktakeItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeDraftSchema.
func (x StocktakeDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakedraft: %w", err)
	}
	if err := validate.Record(StocktakeDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeDrafts returns the first problem found in xs.
func ValidateStocktakeDrafts(xs []StocktakeDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeRevisionSchema.
func (x StocktakeRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakerevision: %w", err)
	}
	if err := validate.Record(StocktakeRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeRevisions returns the first problem found in xs.
func ValidateStocktakeRevisions(xs []StocktakeRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeSnapshotSchema.
func (x StocktakeSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakesnapshot: %w", err)
	}
	if err := validate.Record(StocktakeSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeSnapshots returns the first problem found in xs.
func ValidateStocktakeSnapshots(xs []StocktakeSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeBatchSchema.
func (x StocktakeBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakebatch: %w", err)
	}
	if err := validate.Record(StocktakeBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeBatchs returns the first problem found in xs.
func ValidateStocktakeBatchs(xs []StocktakeBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeRequestSchema.
func (x StocktakeRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakerequest: %w", err)
	}
	if err := validate.Record(StocktakeRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeRequests returns the first problem found in xs.
func ValidateStocktakeRequests(xs []StocktakeRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeSummarySchema.
func (x StocktakeSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakesummary: %w", err)
	}
	if err := validate.Record(StocktakeSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeSummarys returns the first problem found in xs.
func ValidateStocktakeSummarys(xs []StocktakeSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakePolicySchema.
func (x StocktakePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakepolicy: %w", err)
	}
	if err := validate.Record(StocktakePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakePolicys returns the first problem found in xs.
func ValidateStocktakePolicys(xs []StocktakePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeLinkSchema.
func (x StocktakeLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakelink: %w", err)
	}
	if err := validate.Record(StocktakeLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeLinks returns the first problem found in xs.
func ValidateStocktakeLinks(xs []StocktakeLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeNoteSchema.
func (x StocktakeNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakenote: %w", err)
	}
	if err := validate.Record(StocktakeNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeNotes returns the first problem found in xs.
func ValidateStocktakeNotes(xs []StocktakeNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeWindowSchema.
func (x StocktakeWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakewindow: %w", err)
	}
	if err := validate.Record(StocktakeWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeWindows returns the first problem found in xs.
func ValidateStocktakeWindows(xs []StocktakeWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeQuotaSchema.
func (x StocktakeQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakequota: %w", err)
	}
	if err := validate.Record(StocktakeQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeQuotas returns the first problem found in xs.
func ValidateStocktakeQuotas(xs []StocktakeQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeBucketSchema.
func (x StocktakeBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakebucket: %w", err)
	}
	if err := validate.Record(StocktakeBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeBuckets returns the first problem found in xs.
func ValidateStocktakeBuckets(xs []StocktakeBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeLaneSchema.
func (x StocktakeLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakelane: %w", err)
	}
	if err := validate.Record(StocktakeLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeLanes returns the first problem found in xs.
func ValidateStocktakeLanes(xs []StocktakeLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeSlotSchema.
func (x StocktakeSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakeslot: %w", err)
	}
	if err := validate.Record(StocktakeSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakeslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeSlots returns the first problem found in xs.
func ValidateStocktakeSlots(xs []StocktakeSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeTagSchema.
func (x StocktakeTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktaketag: %w", err)
	}
	if err := validate.Record(StocktakeTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktaketag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeTags returns the first problem found in xs.
func ValidateStocktakeTags(xs []StocktakeTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeRuleSchema.
func (x StocktakeRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakerule: %w", err)
	}
	if err := validate.Record(StocktakeRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeRules returns the first problem found in xs.
func ValidateStocktakeRules(xs []StocktakeRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeHoldSchema.
func (x StocktakeHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakehold: %w", err)
	}
	if err := validate.Record(StocktakeHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeHolds returns the first problem found in xs.
func ValidateStocktakeHolds(xs []StocktakeHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeClaimSchema.
func (x StocktakeClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakeclaim: %w", err)
	}
	if err := validate.Record(StocktakeClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakeclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeClaims returns the first problem found in xs.
func ValidateStocktakeClaims(xs []StocktakeClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against StocktakeCursorSchema.
func (x StocktakeCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("stocktake: stocktakecursor: %w", err)
	}
	if err := validate.Record(StocktakeCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("stocktake: stocktakecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateStocktakeCursors returns the first problem found in xs.
func ValidateStocktakeCursors(xs []StocktakeCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
