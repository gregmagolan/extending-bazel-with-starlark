package ledger

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded ledger records against this package's schemas.
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
		return fmt.Errorf("ledger: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against LedgerRecordSchema.
func (x LedgerRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerrecord: %w", err)
	}
	if err := validate.Record(LedgerRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerRecords returns the first problem found in xs.
func ValidateLedgerRecords(xs []LedgerRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerEntrySchema.
func (x LedgerEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerentry: %w", err)
	}
	if err := validate.Record(LedgerEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerEntrys returns the first problem found in xs.
func ValidateLedgerEntrys(xs []LedgerEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerItemSchema.
func (x LedgerItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgeritem: %w", err)
	}
	if err := validate.Record(LedgerItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgeritem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerItems returns the first problem found in xs.
func ValidateLedgerItems(xs []LedgerItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerDraftSchema.
func (x LedgerDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerdraft: %w", err)
	}
	if err := validate.Record(LedgerDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerDrafts returns the first problem found in xs.
func ValidateLedgerDrafts(xs []LedgerDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerRevisionSchema.
func (x LedgerRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerrevision: %w", err)
	}
	if err := validate.Record(LedgerRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerRevisions returns the first problem found in xs.
func ValidateLedgerRevisions(xs []LedgerRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerSnapshotSchema.
func (x LedgerSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgersnapshot: %w", err)
	}
	if err := validate.Record(LedgerSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgersnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerSnapshots returns the first problem found in xs.
func ValidateLedgerSnapshots(xs []LedgerSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerBatchSchema.
func (x LedgerBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerbatch: %w", err)
	}
	if err := validate.Record(LedgerBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerBatchs returns the first problem found in xs.
func ValidateLedgerBatchs(xs []LedgerBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerRequestSchema.
func (x LedgerRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerrequest: %w", err)
	}
	if err := validate.Record(LedgerRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerRequests returns the first problem found in xs.
func ValidateLedgerRequests(xs []LedgerRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerSummarySchema.
func (x LedgerSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgersummary: %w", err)
	}
	if err := validate.Record(LedgerSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgersummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerSummarys returns the first problem found in xs.
func ValidateLedgerSummarys(xs []LedgerSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerPolicySchema.
func (x LedgerPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerpolicy: %w", err)
	}
	if err := validate.Record(LedgerPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerPolicys returns the first problem found in xs.
func ValidateLedgerPolicys(xs []LedgerPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerLinkSchema.
func (x LedgerLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerlink: %w", err)
	}
	if err := validate.Record(LedgerLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerLinks returns the first problem found in xs.
func ValidateLedgerLinks(xs []LedgerLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerNoteSchema.
func (x LedgerNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgernote: %w", err)
	}
	if err := validate.Record(LedgerNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgernote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerNotes returns the first problem found in xs.
func ValidateLedgerNotes(xs []LedgerNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerWindowSchema.
func (x LedgerWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerwindow: %w", err)
	}
	if err := validate.Record(LedgerWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerWindows returns the first problem found in xs.
func ValidateLedgerWindows(xs []LedgerWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerQuotaSchema.
func (x LedgerQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerquota: %w", err)
	}
	if err := validate.Record(LedgerQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerQuotas returns the first problem found in xs.
func ValidateLedgerQuotas(xs []LedgerQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerBucketSchema.
func (x LedgerBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerbucket: %w", err)
	}
	if err := validate.Record(LedgerBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerBuckets returns the first problem found in xs.
func ValidateLedgerBuckets(xs []LedgerBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerLaneSchema.
func (x LedgerLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerlane: %w", err)
	}
	if err := validate.Record(LedgerLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerLanes returns the first problem found in xs.
func ValidateLedgerLanes(xs []LedgerLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerSlotSchema.
func (x LedgerSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerslot: %w", err)
	}
	if err := validate.Record(LedgerSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerSlots returns the first problem found in xs.
func ValidateLedgerSlots(xs []LedgerSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerTagSchema.
func (x LedgerTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgertag: %w", err)
	}
	if err := validate.Record(LedgerTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgertag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerTags returns the first problem found in xs.
func ValidateLedgerTags(xs []LedgerTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerRuleSchema.
func (x LedgerRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerrule: %w", err)
	}
	if err := validate.Record(LedgerRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerRules returns the first problem found in xs.
func ValidateLedgerRules(xs []LedgerRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerHoldSchema.
func (x LedgerHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerhold: %w", err)
	}
	if err := validate.Record(LedgerHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerHolds returns the first problem found in xs.
func ValidateLedgerHolds(xs []LedgerHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerClaimSchema.
func (x LedgerClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgerclaim: %w", err)
	}
	if err := validate.Record(LedgerClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgerclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerClaims returns the first problem found in xs.
func ValidateLedgerClaims(xs []LedgerClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LedgerCursorSchema.
func (x LedgerCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("ledger: ledgercursor: %w", err)
	}
	if err := validate.Record(LedgerCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("ledger: ledgercursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLedgerCursors returns the first problem found in xs.
func ValidateLedgerCursors(xs []LedgerCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
