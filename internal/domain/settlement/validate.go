package settlement

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded settlement records against this package's schemas.
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
		return fmt.Errorf("settlement: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against SettlementRecordSchema.
func (x SettlementRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementrecord: %w", err)
	}
	if err := validate.Record(SettlementRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementRecords returns the first problem found in xs.
func ValidateSettlementRecords(xs []SettlementRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementEntrySchema.
func (x SettlementEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlemententry: %w", err)
	}
	if err := validate.Record(SettlementEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlemententry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementEntrys returns the first problem found in xs.
func ValidateSettlementEntrys(xs []SettlementEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementItemSchema.
func (x SettlementItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementitem: %w", err)
	}
	if err := validate.Record(SettlementItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementItems returns the first problem found in xs.
func ValidateSettlementItems(xs []SettlementItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementDraftSchema.
func (x SettlementDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementdraft: %w", err)
	}
	if err := validate.Record(SettlementDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementDrafts returns the first problem found in xs.
func ValidateSettlementDrafts(xs []SettlementDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementRevisionSchema.
func (x SettlementRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementrevision: %w", err)
	}
	if err := validate.Record(SettlementRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementRevisions returns the first problem found in xs.
func ValidateSettlementRevisions(xs []SettlementRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementSnapshotSchema.
func (x SettlementSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementsnapshot: %w", err)
	}
	if err := validate.Record(SettlementSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementSnapshots returns the first problem found in xs.
func ValidateSettlementSnapshots(xs []SettlementSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementBatchSchema.
func (x SettlementBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementbatch: %w", err)
	}
	if err := validate.Record(SettlementBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementBatchs returns the first problem found in xs.
func ValidateSettlementBatchs(xs []SettlementBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementRequestSchema.
func (x SettlementRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementrequest: %w", err)
	}
	if err := validate.Record(SettlementRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementRequests returns the first problem found in xs.
func ValidateSettlementRequests(xs []SettlementRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementSummarySchema.
func (x SettlementSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementsummary: %w", err)
	}
	if err := validate.Record(SettlementSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementSummarys returns the first problem found in xs.
func ValidateSettlementSummarys(xs []SettlementSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementPolicySchema.
func (x SettlementPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementpolicy: %w", err)
	}
	if err := validate.Record(SettlementPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementPolicys returns the first problem found in xs.
func ValidateSettlementPolicys(xs []SettlementPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementLinkSchema.
func (x SettlementLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementlink: %w", err)
	}
	if err := validate.Record(SettlementLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementLinks returns the first problem found in xs.
func ValidateSettlementLinks(xs []SettlementLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementNoteSchema.
func (x SettlementNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementnote: %w", err)
	}
	if err := validate.Record(SettlementNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementNotes returns the first problem found in xs.
func ValidateSettlementNotes(xs []SettlementNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementWindowSchema.
func (x SettlementWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementwindow: %w", err)
	}
	if err := validate.Record(SettlementWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementWindows returns the first problem found in xs.
func ValidateSettlementWindows(xs []SettlementWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementQuotaSchema.
func (x SettlementQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementquota: %w", err)
	}
	if err := validate.Record(SettlementQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementQuotas returns the first problem found in xs.
func ValidateSettlementQuotas(xs []SettlementQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementBucketSchema.
func (x SettlementBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementbucket: %w", err)
	}
	if err := validate.Record(SettlementBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementBuckets returns the first problem found in xs.
func ValidateSettlementBuckets(xs []SettlementBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementLaneSchema.
func (x SettlementLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementlane: %w", err)
	}
	if err := validate.Record(SettlementLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementLanes returns the first problem found in xs.
func ValidateSettlementLanes(xs []SettlementLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementSlotSchema.
func (x SettlementSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementslot: %w", err)
	}
	if err := validate.Record(SettlementSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementSlots returns the first problem found in xs.
func ValidateSettlementSlots(xs []SettlementSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementTagSchema.
func (x SettlementTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementtag: %w", err)
	}
	if err := validate.Record(SettlementTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementTags returns the first problem found in xs.
func ValidateSettlementTags(xs []SettlementTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementRuleSchema.
func (x SettlementRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementrule: %w", err)
	}
	if err := validate.Record(SettlementRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementRules returns the first problem found in xs.
func ValidateSettlementRules(xs []SettlementRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementHoldSchema.
func (x SettlementHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementhold: %w", err)
	}
	if err := validate.Record(SettlementHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementHolds returns the first problem found in xs.
func ValidateSettlementHolds(xs []SettlementHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementClaimSchema.
func (x SettlementClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementclaim: %w", err)
	}
	if err := validate.Record(SettlementClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementClaims returns the first problem found in xs.
func ValidateSettlementClaims(xs []SettlementClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SettlementCursorSchema.
func (x SettlementCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("settlement: settlementcursor: %w", err)
	}
	if err := validate.Record(SettlementCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("settlement: settlementcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSettlementCursors returns the first problem found in xs.
func ValidateSettlementCursors(xs []SettlementCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
