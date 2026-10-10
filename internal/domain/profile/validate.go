package profile

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded profile records against this package's schemas.
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
		return fmt.Errorf("profile: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against ProfileRecordSchema.
func (x ProfileRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilerecord: %w", err)
	}
	if err := validate.Record(ProfileRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileRecords returns the first problem found in xs.
func ValidateProfileRecords(xs []ProfileRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileEntrySchema.
func (x ProfileEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profileentry: %w", err)
	}
	if err := validate.Record(ProfileEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profileentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileEntrys returns the first problem found in xs.
func ValidateProfileEntrys(xs []ProfileEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileItemSchema.
func (x ProfileItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profileitem: %w", err)
	}
	if err := validate.Record(ProfileItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profileitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileItems returns the first problem found in xs.
func ValidateProfileItems(xs []ProfileItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileDraftSchema.
func (x ProfileDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profiledraft: %w", err)
	}
	if err := validate.Record(ProfileDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profiledraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileDrafts returns the first problem found in xs.
func ValidateProfileDrafts(xs []ProfileDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileRevisionSchema.
func (x ProfileRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilerevision: %w", err)
	}
	if err := validate.Record(ProfileRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileRevisions returns the first problem found in xs.
func ValidateProfileRevisions(xs []ProfileRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileSnapshotSchema.
func (x ProfileSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilesnapshot: %w", err)
	}
	if err := validate.Record(ProfileSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileSnapshots returns the first problem found in xs.
func ValidateProfileSnapshots(xs []ProfileSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileBatchSchema.
func (x ProfileBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilebatch: %w", err)
	}
	if err := validate.Record(ProfileBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileBatchs returns the first problem found in xs.
func ValidateProfileBatchs(xs []ProfileBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileRequestSchema.
func (x ProfileRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilerequest: %w", err)
	}
	if err := validate.Record(ProfileRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileRequests returns the first problem found in xs.
func ValidateProfileRequests(xs []ProfileRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileSummarySchema.
func (x ProfileSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilesummary: %w", err)
	}
	if err := validate.Record(ProfileSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileSummarys returns the first problem found in xs.
func ValidateProfileSummarys(xs []ProfileSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfilePolicySchema.
func (x ProfilePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilepolicy: %w", err)
	}
	if err := validate.Record(ProfilePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfilePolicys returns the first problem found in xs.
func ValidateProfilePolicys(xs []ProfilePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileLinkSchema.
func (x ProfileLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilelink: %w", err)
	}
	if err := validate.Record(ProfileLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileLinks returns the first problem found in xs.
func ValidateProfileLinks(xs []ProfileLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileNoteSchema.
func (x ProfileNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilenote: %w", err)
	}
	if err := validate.Record(ProfileNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileNotes returns the first problem found in xs.
func ValidateProfileNotes(xs []ProfileNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileWindowSchema.
func (x ProfileWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilewindow: %w", err)
	}
	if err := validate.Record(ProfileWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileWindows returns the first problem found in xs.
func ValidateProfileWindows(xs []ProfileWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileQuotaSchema.
func (x ProfileQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilequota: %w", err)
	}
	if err := validate.Record(ProfileQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileQuotas returns the first problem found in xs.
func ValidateProfileQuotas(xs []ProfileQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileBucketSchema.
func (x ProfileBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilebucket: %w", err)
	}
	if err := validate.Record(ProfileBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileBuckets returns the first problem found in xs.
func ValidateProfileBuckets(xs []ProfileBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileLaneSchema.
func (x ProfileLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilelane: %w", err)
	}
	if err := validate.Record(ProfileLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileLanes returns the first problem found in xs.
func ValidateProfileLanes(xs []ProfileLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileSlotSchema.
func (x ProfileSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profileslot: %w", err)
	}
	if err := validate.Record(ProfileSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profileslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileSlots returns the first problem found in xs.
func ValidateProfileSlots(xs []ProfileSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileTagSchema.
func (x ProfileTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profiletag: %w", err)
	}
	if err := validate.Record(ProfileTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profiletag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileTags returns the first problem found in xs.
func ValidateProfileTags(xs []ProfileTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileRuleSchema.
func (x ProfileRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilerule: %w", err)
	}
	if err := validate.Record(ProfileRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileRules returns the first problem found in xs.
func ValidateProfileRules(xs []ProfileRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileHoldSchema.
func (x ProfileHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilehold: %w", err)
	}
	if err := validate.Record(ProfileHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileHolds returns the first problem found in xs.
func ValidateProfileHolds(xs []ProfileHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileClaimSchema.
func (x ProfileClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profileclaim: %w", err)
	}
	if err := validate.Record(ProfileClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profileclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileClaims returns the first problem found in xs.
func ValidateProfileClaims(xs []ProfileClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ProfileCursorSchema.
func (x ProfileCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("profile: profilecursor: %w", err)
	}
	if err := validate.Record(ProfileCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("profile: profilecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateProfileCursors returns the first problem found in xs.
func ValidateProfileCursors(xs []ProfileCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
