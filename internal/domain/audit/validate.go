package audit

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded audit records against this package's schemas.
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
		return fmt.Errorf("audit: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against AuditRecordSchema.
func (x AuditRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditrecord: %w", err)
	}
	if err := validate.Record(AuditRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditRecords returns the first problem found in xs.
func ValidateAuditRecords(xs []AuditRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditEntrySchema.
func (x AuditEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditentry: %w", err)
	}
	if err := validate.Record(AuditEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditEntrys returns the first problem found in xs.
func ValidateAuditEntrys(xs []AuditEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditItemSchema.
func (x AuditItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: audititem: %w", err)
	}
	if err := validate.Record(AuditItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: audititem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditItems returns the first problem found in xs.
func ValidateAuditItems(xs []AuditItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditDraftSchema.
func (x AuditDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditdraft: %w", err)
	}
	if err := validate.Record(AuditDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditDrafts returns the first problem found in xs.
func ValidateAuditDrafts(xs []AuditDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditRevisionSchema.
func (x AuditRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditrevision: %w", err)
	}
	if err := validate.Record(AuditRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditRevisions returns the first problem found in xs.
func ValidateAuditRevisions(xs []AuditRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditSnapshotSchema.
func (x AuditSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditsnapshot: %w", err)
	}
	if err := validate.Record(AuditSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditSnapshots returns the first problem found in xs.
func ValidateAuditSnapshots(xs []AuditSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditBatchSchema.
func (x AuditBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditbatch: %w", err)
	}
	if err := validate.Record(AuditBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditBatchs returns the first problem found in xs.
func ValidateAuditBatchs(xs []AuditBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditRequestSchema.
func (x AuditRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditrequest: %w", err)
	}
	if err := validate.Record(AuditRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditRequests returns the first problem found in xs.
func ValidateAuditRequests(xs []AuditRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditSummarySchema.
func (x AuditSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditsummary: %w", err)
	}
	if err := validate.Record(AuditSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditSummarys returns the first problem found in xs.
func ValidateAuditSummarys(xs []AuditSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditPolicySchema.
func (x AuditPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditpolicy: %w", err)
	}
	if err := validate.Record(AuditPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditPolicys returns the first problem found in xs.
func ValidateAuditPolicys(xs []AuditPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditLinkSchema.
func (x AuditLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditlink: %w", err)
	}
	if err := validate.Record(AuditLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditLinks returns the first problem found in xs.
func ValidateAuditLinks(xs []AuditLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditNoteSchema.
func (x AuditNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditnote: %w", err)
	}
	if err := validate.Record(AuditNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditNotes returns the first problem found in xs.
func ValidateAuditNotes(xs []AuditNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditWindowSchema.
func (x AuditWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditwindow: %w", err)
	}
	if err := validate.Record(AuditWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditWindows returns the first problem found in xs.
func ValidateAuditWindows(xs []AuditWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditQuotaSchema.
func (x AuditQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditquota: %w", err)
	}
	if err := validate.Record(AuditQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditQuotas returns the first problem found in xs.
func ValidateAuditQuotas(xs []AuditQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditBucketSchema.
func (x AuditBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditbucket: %w", err)
	}
	if err := validate.Record(AuditBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditBuckets returns the first problem found in xs.
func ValidateAuditBuckets(xs []AuditBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditLaneSchema.
func (x AuditLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditlane: %w", err)
	}
	if err := validate.Record(AuditLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditLanes returns the first problem found in xs.
func ValidateAuditLanes(xs []AuditLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditSlotSchema.
func (x AuditSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditslot: %w", err)
	}
	if err := validate.Record(AuditSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditSlots returns the first problem found in xs.
func ValidateAuditSlots(xs []AuditSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditTagSchema.
func (x AuditTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: audittag: %w", err)
	}
	if err := validate.Record(AuditTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: audittag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditTags returns the first problem found in xs.
func ValidateAuditTags(xs []AuditTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditRuleSchema.
func (x AuditRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditrule: %w", err)
	}
	if err := validate.Record(AuditRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditRules returns the first problem found in xs.
func ValidateAuditRules(xs []AuditRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditHoldSchema.
func (x AuditHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: audithold: %w", err)
	}
	if err := validate.Record(AuditHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: audithold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditHolds returns the first problem found in xs.
func ValidateAuditHolds(xs []AuditHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditClaimSchema.
func (x AuditClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditclaim: %w", err)
	}
	if err := validate.Record(AuditClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditClaims returns the first problem found in xs.
func ValidateAuditClaims(xs []AuditClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against AuditCursorSchema.
func (x AuditCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("audit: auditcursor: %w", err)
	}
	if err := validate.Record(AuditCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("audit: auditcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateAuditCursors returns the first problem found in xs.
func ValidateAuditCursors(xs []AuditCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
