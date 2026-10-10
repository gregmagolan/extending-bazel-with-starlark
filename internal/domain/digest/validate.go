package digest

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded digest records against this package's schemas.
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
		return fmt.Errorf("digest: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against DigestRecordSchema.
func (x DigestRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestrecord: %w", err)
	}
	if err := validate.Record(DigestRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestRecords returns the first problem found in xs.
func ValidateDigestRecords(xs []DigestRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestEntrySchema.
func (x DigestEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestentry: %w", err)
	}
	if err := validate.Record(DigestEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestEntrys returns the first problem found in xs.
func ValidateDigestEntrys(xs []DigestEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestItemSchema.
func (x DigestItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestitem: %w", err)
	}
	if err := validate.Record(DigestItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestItems returns the first problem found in xs.
func ValidateDigestItems(xs []DigestItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestDraftSchema.
func (x DigestDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestdraft: %w", err)
	}
	if err := validate.Record(DigestDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestDrafts returns the first problem found in xs.
func ValidateDigestDrafts(xs []DigestDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestRevisionSchema.
func (x DigestRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestrevision: %w", err)
	}
	if err := validate.Record(DigestRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestRevisions returns the first problem found in xs.
func ValidateDigestRevisions(xs []DigestRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestSnapshotSchema.
func (x DigestSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestsnapshot: %w", err)
	}
	if err := validate.Record(DigestSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestSnapshots returns the first problem found in xs.
func ValidateDigestSnapshots(xs []DigestSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestBatchSchema.
func (x DigestBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestbatch: %w", err)
	}
	if err := validate.Record(DigestBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestBatchs returns the first problem found in xs.
func ValidateDigestBatchs(xs []DigestBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestRequestSchema.
func (x DigestRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestrequest: %w", err)
	}
	if err := validate.Record(DigestRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestRequests returns the first problem found in xs.
func ValidateDigestRequests(xs []DigestRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestSummarySchema.
func (x DigestSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestsummary: %w", err)
	}
	if err := validate.Record(DigestSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestSummarys returns the first problem found in xs.
func ValidateDigestSummarys(xs []DigestSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestPolicySchema.
func (x DigestPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestpolicy: %w", err)
	}
	if err := validate.Record(DigestPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestPolicys returns the first problem found in xs.
func ValidateDigestPolicys(xs []DigestPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestLinkSchema.
func (x DigestLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestlink: %w", err)
	}
	if err := validate.Record(DigestLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestLinks returns the first problem found in xs.
func ValidateDigestLinks(xs []DigestLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestNoteSchema.
func (x DigestNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestnote: %w", err)
	}
	if err := validate.Record(DigestNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestNotes returns the first problem found in xs.
func ValidateDigestNotes(xs []DigestNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestWindowSchema.
func (x DigestWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestwindow: %w", err)
	}
	if err := validate.Record(DigestWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestWindows returns the first problem found in xs.
func ValidateDigestWindows(xs []DigestWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestQuotaSchema.
func (x DigestQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestquota: %w", err)
	}
	if err := validate.Record(DigestQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestQuotas returns the first problem found in xs.
func ValidateDigestQuotas(xs []DigestQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestBucketSchema.
func (x DigestBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestbucket: %w", err)
	}
	if err := validate.Record(DigestBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestBuckets returns the first problem found in xs.
func ValidateDigestBuckets(xs []DigestBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestLaneSchema.
func (x DigestLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestlane: %w", err)
	}
	if err := validate.Record(DigestLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestLanes returns the first problem found in xs.
func ValidateDigestLanes(xs []DigestLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestSlotSchema.
func (x DigestSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestslot: %w", err)
	}
	if err := validate.Record(DigestSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestSlots returns the first problem found in xs.
func ValidateDigestSlots(xs []DigestSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestTagSchema.
func (x DigestTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digesttag: %w", err)
	}
	if err := validate.Record(DigestTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digesttag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestTags returns the first problem found in xs.
func ValidateDigestTags(xs []DigestTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestRuleSchema.
func (x DigestRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestrule: %w", err)
	}
	if err := validate.Record(DigestRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestRules returns the first problem found in xs.
func ValidateDigestRules(xs []DigestRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestHoldSchema.
func (x DigestHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digesthold: %w", err)
	}
	if err := validate.Record(DigestHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digesthold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestHolds returns the first problem found in xs.
func ValidateDigestHolds(xs []DigestHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestClaimSchema.
func (x DigestClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestclaim: %w", err)
	}
	if err := validate.Record(DigestClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestClaims returns the first problem found in xs.
func ValidateDigestClaims(xs []DigestClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DigestCursorSchema.
func (x DigestCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("digest: digestcursor: %w", err)
	}
	if err := validate.Record(DigestCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("digest: digestcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDigestCursors returns the first problem found in xs.
func ValidateDigestCursors(xs []DigestCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
