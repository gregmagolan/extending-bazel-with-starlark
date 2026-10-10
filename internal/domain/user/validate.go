package user

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded user records against this package's schemas.
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
		return fmt.Errorf("user: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against UserRecordSchema.
func (x UserRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userrecord: %w", err)
	}
	if err := validate.Record(UserRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserRecords returns the first problem found in xs.
func ValidateUserRecords(xs []UserRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserEntrySchema.
func (x UserEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userentry: %w", err)
	}
	if err := validate.Record(UserEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserEntrys returns the first problem found in xs.
func ValidateUserEntrys(xs []UserEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserItemSchema.
func (x UserItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: useritem: %w", err)
	}
	if err := validate.Record(UserItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: useritem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserItems returns the first problem found in xs.
func ValidateUserItems(xs []UserItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserDraftSchema.
func (x UserDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userdraft: %w", err)
	}
	if err := validate.Record(UserDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserDrafts returns the first problem found in xs.
func ValidateUserDrafts(xs []UserDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserRevisionSchema.
func (x UserRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userrevision: %w", err)
	}
	if err := validate.Record(UserRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserRevisions returns the first problem found in xs.
func ValidateUserRevisions(xs []UserRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserSnapshotSchema.
func (x UserSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: usersnapshot: %w", err)
	}
	if err := validate.Record(UserSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: usersnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserSnapshots returns the first problem found in xs.
func ValidateUserSnapshots(xs []UserSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserBatchSchema.
func (x UserBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userbatch: %w", err)
	}
	if err := validate.Record(UserBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserBatchs returns the first problem found in xs.
func ValidateUserBatchs(xs []UserBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserRequestSchema.
func (x UserRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userrequest: %w", err)
	}
	if err := validate.Record(UserRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserRequests returns the first problem found in xs.
func ValidateUserRequests(xs []UserRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserSummarySchema.
func (x UserSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: usersummary: %w", err)
	}
	if err := validate.Record(UserSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: usersummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserSummarys returns the first problem found in xs.
func ValidateUserSummarys(xs []UserSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserPolicySchema.
func (x UserPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userpolicy: %w", err)
	}
	if err := validate.Record(UserPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserPolicys returns the first problem found in xs.
func ValidateUserPolicys(xs []UserPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserLinkSchema.
func (x UserLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userlink: %w", err)
	}
	if err := validate.Record(UserLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserLinks returns the first problem found in xs.
func ValidateUserLinks(xs []UserLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserNoteSchema.
func (x UserNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: usernote: %w", err)
	}
	if err := validate.Record(UserNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: usernote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserNotes returns the first problem found in xs.
func ValidateUserNotes(xs []UserNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserWindowSchema.
func (x UserWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userwindow: %w", err)
	}
	if err := validate.Record(UserWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserWindows returns the first problem found in xs.
func ValidateUserWindows(xs []UserWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserQuotaSchema.
func (x UserQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userquota: %w", err)
	}
	if err := validate.Record(UserQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserQuotas returns the first problem found in xs.
func ValidateUserQuotas(xs []UserQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserBucketSchema.
func (x UserBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userbucket: %w", err)
	}
	if err := validate.Record(UserBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserBuckets returns the first problem found in xs.
func ValidateUserBuckets(xs []UserBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserLaneSchema.
func (x UserLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userlane: %w", err)
	}
	if err := validate.Record(UserLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserLanes returns the first problem found in xs.
func ValidateUserLanes(xs []UserLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserSlotSchema.
func (x UserSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userslot: %w", err)
	}
	if err := validate.Record(UserSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserSlots returns the first problem found in xs.
func ValidateUserSlots(xs []UserSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserTagSchema.
func (x UserTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: usertag: %w", err)
	}
	if err := validate.Record(UserTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: usertag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserTags returns the first problem found in xs.
func ValidateUserTags(xs []UserTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserRuleSchema.
func (x UserRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userrule: %w", err)
	}
	if err := validate.Record(UserRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserRules returns the first problem found in xs.
func ValidateUserRules(xs []UserRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserHoldSchema.
func (x UserHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userhold: %w", err)
	}
	if err := validate.Record(UserHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserHolds returns the first problem found in xs.
func ValidateUserHolds(xs []UserHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserClaimSchema.
func (x UserClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: userclaim: %w", err)
	}
	if err := validate.Record(UserClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: userclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserClaims returns the first problem found in xs.
func ValidateUserClaims(xs []UserClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against UserCursorSchema.
func (x UserCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("user: usercursor: %w", err)
	}
	if err := validate.Record(UserCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("user: usercursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateUserCursors returns the first problem found in xs.
func ValidateUserCursors(xs []UserCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
