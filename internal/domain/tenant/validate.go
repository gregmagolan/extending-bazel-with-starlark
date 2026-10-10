package tenant

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded tenant records against this package's schemas.
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
		return fmt.Errorf("tenant: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against TenantRecordSchema.
func (x TenantRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantrecord: %w", err)
	}
	if err := validate.Record(TenantRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantRecords returns the first problem found in xs.
func ValidateTenantRecords(xs []TenantRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantEntrySchema.
func (x TenantEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantentry: %w", err)
	}
	if err := validate.Record(TenantEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantEntrys returns the first problem found in xs.
func ValidateTenantEntrys(xs []TenantEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantItemSchema.
func (x TenantItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantitem: %w", err)
	}
	if err := validate.Record(TenantItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantItems returns the first problem found in xs.
func ValidateTenantItems(xs []TenantItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantDraftSchema.
func (x TenantDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantdraft: %w", err)
	}
	if err := validate.Record(TenantDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantDrafts returns the first problem found in xs.
func ValidateTenantDrafts(xs []TenantDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantRevisionSchema.
func (x TenantRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantrevision: %w", err)
	}
	if err := validate.Record(TenantRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantRevisions returns the first problem found in xs.
func ValidateTenantRevisions(xs []TenantRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantSnapshotSchema.
func (x TenantSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantsnapshot: %w", err)
	}
	if err := validate.Record(TenantSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantSnapshots returns the first problem found in xs.
func ValidateTenantSnapshots(xs []TenantSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantBatchSchema.
func (x TenantBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantbatch: %w", err)
	}
	if err := validate.Record(TenantBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantBatchs returns the first problem found in xs.
func ValidateTenantBatchs(xs []TenantBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantRequestSchema.
func (x TenantRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantrequest: %w", err)
	}
	if err := validate.Record(TenantRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantRequests returns the first problem found in xs.
func ValidateTenantRequests(xs []TenantRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantSummarySchema.
func (x TenantSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantsummary: %w", err)
	}
	if err := validate.Record(TenantSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantSummarys returns the first problem found in xs.
func ValidateTenantSummarys(xs []TenantSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantPolicySchema.
func (x TenantPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantpolicy: %w", err)
	}
	if err := validate.Record(TenantPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantPolicys returns the first problem found in xs.
func ValidateTenantPolicys(xs []TenantPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantLinkSchema.
func (x TenantLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantlink: %w", err)
	}
	if err := validate.Record(TenantLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantLinks returns the first problem found in xs.
func ValidateTenantLinks(xs []TenantLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantNoteSchema.
func (x TenantNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantnote: %w", err)
	}
	if err := validate.Record(TenantNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantNotes returns the first problem found in xs.
func ValidateTenantNotes(xs []TenantNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantWindowSchema.
func (x TenantWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantwindow: %w", err)
	}
	if err := validate.Record(TenantWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantWindows returns the first problem found in xs.
func ValidateTenantWindows(xs []TenantWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantQuotaSchema.
func (x TenantQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantquota: %w", err)
	}
	if err := validate.Record(TenantQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantQuotas returns the first problem found in xs.
func ValidateTenantQuotas(xs []TenantQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantBucketSchema.
func (x TenantBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantbucket: %w", err)
	}
	if err := validate.Record(TenantBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantBuckets returns the first problem found in xs.
func ValidateTenantBuckets(xs []TenantBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantLaneSchema.
func (x TenantLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantlane: %w", err)
	}
	if err := validate.Record(TenantLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantLanes returns the first problem found in xs.
func ValidateTenantLanes(xs []TenantLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantSlotSchema.
func (x TenantSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantslot: %w", err)
	}
	if err := validate.Record(TenantSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantSlots returns the first problem found in xs.
func ValidateTenantSlots(xs []TenantSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantTagSchema.
func (x TenantTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenanttag: %w", err)
	}
	if err := validate.Record(TenantTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenanttag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantTags returns the first problem found in xs.
func ValidateTenantTags(xs []TenantTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantRuleSchema.
func (x TenantRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantrule: %w", err)
	}
	if err := validate.Record(TenantRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantRules returns the first problem found in xs.
func ValidateTenantRules(xs []TenantRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantHoldSchema.
func (x TenantHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenanthold: %w", err)
	}
	if err := validate.Record(TenantHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenanthold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantHolds returns the first problem found in xs.
func ValidateTenantHolds(xs []TenantHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantClaimSchema.
func (x TenantClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantclaim: %w", err)
	}
	if err := validate.Record(TenantClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantClaims returns the first problem found in xs.
func ValidateTenantClaims(xs []TenantClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against TenantCursorSchema.
func (x TenantCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("tenant: tenantcursor: %w", err)
	}
	if err := validate.Record(TenantCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("tenant: tenantcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateTenantCursors returns the first problem found in xs.
func ValidateTenantCursors(xs []TenantCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
