package report

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded report records against this package's schemas.
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
		return fmt.Errorf("report: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against ReportRecordSchema.
func (x ReportRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportrecord: %w", err)
	}
	if err := validate.Record(ReportRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportRecords returns the first problem found in xs.
func ValidateReportRecords(xs []ReportRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportEntrySchema.
func (x ReportEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportentry: %w", err)
	}
	if err := validate.Record(ReportEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportEntrys returns the first problem found in xs.
func ValidateReportEntrys(xs []ReportEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportItemSchema.
func (x ReportItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportitem: %w", err)
	}
	if err := validate.Record(ReportItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportItems returns the first problem found in xs.
func ValidateReportItems(xs []ReportItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportDraftSchema.
func (x ReportDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportdraft: %w", err)
	}
	if err := validate.Record(ReportDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportDrafts returns the first problem found in xs.
func ValidateReportDrafts(xs []ReportDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportRevisionSchema.
func (x ReportRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportrevision: %w", err)
	}
	if err := validate.Record(ReportRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportRevisions returns the first problem found in xs.
func ValidateReportRevisions(xs []ReportRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportSnapshotSchema.
func (x ReportSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportsnapshot: %w", err)
	}
	if err := validate.Record(ReportSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportSnapshots returns the first problem found in xs.
func ValidateReportSnapshots(xs []ReportSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportBatchSchema.
func (x ReportBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportbatch: %w", err)
	}
	if err := validate.Record(ReportBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportBatchs returns the first problem found in xs.
func ValidateReportBatchs(xs []ReportBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportRequestSchema.
func (x ReportRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportrequest: %w", err)
	}
	if err := validate.Record(ReportRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportRequests returns the first problem found in xs.
func ValidateReportRequests(xs []ReportRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportSummarySchema.
func (x ReportSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportsummary: %w", err)
	}
	if err := validate.Record(ReportSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportSummarys returns the first problem found in xs.
func ValidateReportSummarys(xs []ReportSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportPolicySchema.
func (x ReportPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportpolicy: %w", err)
	}
	if err := validate.Record(ReportPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportPolicys returns the first problem found in xs.
func ValidateReportPolicys(xs []ReportPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportLinkSchema.
func (x ReportLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportlink: %w", err)
	}
	if err := validate.Record(ReportLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportLinks returns the first problem found in xs.
func ValidateReportLinks(xs []ReportLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportNoteSchema.
func (x ReportNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportnote: %w", err)
	}
	if err := validate.Record(ReportNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportNotes returns the first problem found in xs.
func ValidateReportNotes(xs []ReportNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportWindowSchema.
func (x ReportWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportwindow: %w", err)
	}
	if err := validate.Record(ReportWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportWindows returns the first problem found in xs.
func ValidateReportWindows(xs []ReportWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportQuotaSchema.
func (x ReportQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportquota: %w", err)
	}
	if err := validate.Record(ReportQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportQuotas returns the first problem found in xs.
func ValidateReportQuotas(xs []ReportQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportBucketSchema.
func (x ReportBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportbucket: %w", err)
	}
	if err := validate.Record(ReportBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportBuckets returns the first problem found in xs.
func ValidateReportBuckets(xs []ReportBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportLaneSchema.
func (x ReportLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportlane: %w", err)
	}
	if err := validate.Record(ReportLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportLanes returns the first problem found in xs.
func ValidateReportLanes(xs []ReportLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportSlotSchema.
func (x ReportSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportslot: %w", err)
	}
	if err := validate.Record(ReportSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportSlots returns the first problem found in xs.
func ValidateReportSlots(xs []ReportSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportTagSchema.
func (x ReportTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reporttag: %w", err)
	}
	if err := validate.Record(ReportTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reporttag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportTags returns the first problem found in xs.
func ValidateReportTags(xs []ReportTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportRuleSchema.
func (x ReportRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportrule: %w", err)
	}
	if err := validate.Record(ReportRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportRules returns the first problem found in xs.
func ValidateReportRules(xs []ReportRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportHoldSchema.
func (x ReportHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reporthold: %w", err)
	}
	if err := validate.Record(ReportHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reporthold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportHolds returns the first problem found in xs.
func ValidateReportHolds(xs []ReportHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportClaimSchema.
func (x ReportClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportclaim: %w", err)
	}
	if err := validate.Record(ReportClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportClaims returns the first problem found in xs.
func ValidateReportClaims(xs []ReportClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ReportCursorSchema.
func (x ReportCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("report: reportcursor: %w", err)
	}
	if err := validate.Record(ReportCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("report: reportcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateReportCursors returns the first problem found in xs.
func ValidateReportCursors(xs []ReportCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
