package session

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded session records against this package's schemas.
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
		return fmt.Errorf("session: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against SessionRecordSchema.
func (x SessionRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionrecord: %w", err)
	}
	if err := validate.Record(SessionRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionRecords returns the first problem found in xs.
func ValidateSessionRecords(xs []SessionRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionEntrySchema.
func (x SessionEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionentry: %w", err)
	}
	if err := validate.Record(SessionEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionEntrys returns the first problem found in xs.
func ValidateSessionEntrys(xs []SessionEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionItemSchema.
func (x SessionItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionitem: %w", err)
	}
	if err := validate.Record(SessionItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionItems returns the first problem found in xs.
func ValidateSessionItems(xs []SessionItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionDraftSchema.
func (x SessionDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessiondraft: %w", err)
	}
	if err := validate.Record(SessionDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessiondraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionDrafts returns the first problem found in xs.
func ValidateSessionDrafts(xs []SessionDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionRevisionSchema.
func (x SessionRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionrevision: %w", err)
	}
	if err := validate.Record(SessionRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionRevisions returns the first problem found in xs.
func ValidateSessionRevisions(xs []SessionRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionSnapshotSchema.
func (x SessionSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionsnapshot: %w", err)
	}
	if err := validate.Record(SessionSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionSnapshots returns the first problem found in xs.
func ValidateSessionSnapshots(xs []SessionSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionBatchSchema.
func (x SessionBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionbatch: %w", err)
	}
	if err := validate.Record(SessionBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionBatchs returns the first problem found in xs.
func ValidateSessionBatchs(xs []SessionBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionRequestSchema.
func (x SessionRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionrequest: %w", err)
	}
	if err := validate.Record(SessionRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionRequests returns the first problem found in xs.
func ValidateSessionRequests(xs []SessionRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionSummarySchema.
func (x SessionSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionsummary: %w", err)
	}
	if err := validate.Record(SessionSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionSummarys returns the first problem found in xs.
func ValidateSessionSummarys(xs []SessionSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionPolicySchema.
func (x SessionPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionpolicy: %w", err)
	}
	if err := validate.Record(SessionPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionPolicys returns the first problem found in xs.
func ValidateSessionPolicys(xs []SessionPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionLinkSchema.
func (x SessionLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionlink: %w", err)
	}
	if err := validate.Record(SessionLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionLinks returns the first problem found in xs.
func ValidateSessionLinks(xs []SessionLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionNoteSchema.
func (x SessionNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionnote: %w", err)
	}
	if err := validate.Record(SessionNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionNotes returns the first problem found in xs.
func ValidateSessionNotes(xs []SessionNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionWindowSchema.
func (x SessionWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionwindow: %w", err)
	}
	if err := validate.Record(SessionWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionWindows returns the first problem found in xs.
func ValidateSessionWindows(xs []SessionWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionQuotaSchema.
func (x SessionQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionquota: %w", err)
	}
	if err := validate.Record(SessionQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionQuotas returns the first problem found in xs.
func ValidateSessionQuotas(xs []SessionQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionBucketSchema.
func (x SessionBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionbucket: %w", err)
	}
	if err := validate.Record(SessionBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionBuckets returns the first problem found in xs.
func ValidateSessionBuckets(xs []SessionBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionLaneSchema.
func (x SessionLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionlane: %w", err)
	}
	if err := validate.Record(SessionLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionLanes returns the first problem found in xs.
func ValidateSessionLanes(xs []SessionLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionSlotSchema.
func (x SessionSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionslot: %w", err)
	}
	if err := validate.Record(SessionSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionSlots returns the first problem found in xs.
func ValidateSessionSlots(xs []SessionSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionTagSchema.
func (x SessionTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessiontag: %w", err)
	}
	if err := validate.Record(SessionTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessiontag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionTags returns the first problem found in xs.
func ValidateSessionTags(xs []SessionTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionRuleSchema.
func (x SessionRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionrule: %w", err)
	}
	if err := validate.Record(SessionRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionRules returns the first problem found in xs.
func ValidateSessionRules(xs []SessionRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionHoldSchema.
func (x SessionHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionhold: %w", err)
	}
	if err := validate.Record(SessionHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionHolds returns the first problem found in xs.
func ValidateSessionHolds(xs []SessionHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionClaimSchema.
func (x SessionClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessionclaim: %w", err)
	}
	if err := validate.Record(SessionClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessionclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionClaims returns the first problem found in xs.
func ValidateSessionClaims(xs []SessionClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against SessionCursorSchema.
func (x SessionCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("session: sessioncursor: %w", err)
	}
	if err := validate.Record(SessionCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("session: sessioncursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateSessionCursors returns the first problem found in xs.
func ValidateSessionCursors(xs []SessionCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
