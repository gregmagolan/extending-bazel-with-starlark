package preference

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded preference records against this package's schemas.
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
		return fmt.Errorf("preference: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against PreferenceRecordSchema.
func (x PreferenceRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencerecord: %w", err)
	}
	if err := validate.Record(PreferenceRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceRecords returns the first problem found in xs.
func ValidatePreferenceRecords(xs []PreferenceRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceEntrySchema.
func (x PreferenceEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferenceentry: %w", err)
	}
	if err := validate.Record(PreferenceEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferenceentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceEntrys returns the first problem found in xs.
func ValidatePreferenceEntrys(xs []PreferenceEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceItemSchema.
func (x PreferenceItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferenceitem: %w", err)
	}
	if err := validate.Record(PreferenceItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferenceitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceItems returns the first problem found in xs.
func ValidatePreferenceItems(xs []PreferenceItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceDraftSchema.
func (x PreferenceDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencedraft: %w", err)
	}
	if err := validate.Record(PreferenceDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceDrafts returns the first problem found in xs.
func ValidatePreferenceDrafts(xs []PreferenceDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceRevisionSchema.
func (x PreferenceRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencerevision: %w", err)
	}
	if err := validate.Record(PreferenceRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceRevisions returns the first problem found in xs.
func ValidatePreferenceRevisions(xs []PreferenceRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceSnapshotSchema.
func (x PreferenceSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencesnapshot: %w", err)
	}
	if err := validate.Record(PreferenceSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceSnapshots returns the first problem found in xs.
func ValidatePreferenceSnapshots(xs []PreferenceSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceBatchSchema.
func (x PreferenceBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencebatch: %w", err)
	}
	if err := validate.Record(PreferenceBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceBatchs returns the first problem found in xs.
func ValidatePreferenceBatchs(xs []PreferenceBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceRequestSchema.
func (x PreferenceRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencerequest: %w", err)
	}
	if err := validate.Record(PreferenceRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceRequests returns the first problem found in xs.
func ValidatePreferenceRequests(xs []PreferenceRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceSummarySchema.
func (x PreferenceSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencesummary: %w", err)
	}
	if err := validate.Record(PreferenceSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceSummarys returns the first problem found in xs.
func ValidatePreferenceSummarys(xs []PreferenceSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferencePolicySchema.
func (x PreferencePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencepolicy: %w", err)
	}
	if err := validate.Record(PreferencePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferencePolicys returns the first problem found in xs.
func ValidatePreferencePolicys(xs []PreferencePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceLinkSchema.
func (x PreferenceLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencelink: %w", err)
	}
	if err := validate.Record(PreferenceLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceLinks returns the first problem found in xs.
func ValidatePreferenceLinks(xs []PreferenceLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceNoteSchema.
func (x PreferenceNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencenote: %w", err)
	}
	if err := validate.Record(PreferenceNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceNotes returns the first problem found in xs.
func ValidatePreferenceNotes(xs []PreferenceNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceWindowSchema.
func (x PreferenceWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencewindow: %w", err)
	}
	if err := validate.Record(PreferenceWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceWindows returns the first problem found in xs.
func ValidatePreferenceWindows(xs []PreferenceWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceQuotaSchema.
func (x PreferenceQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencequota: %w", err)
	}
	if err := validate.Record(PreferenceQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceQuotas returns the first problem found in xs.
func ValidatePreferenceQuotas(xs []PreferenceQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceBucketSchema.
func (x PreferenceBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencebucket: %w", err)
	}
	if err := validate.Record(PreferenceBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceBuckets returns the first problem found in xs.
func ValidatePreferenceBuckets(xs []PreferenceBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceLaneSchema.
func (x PreferenceLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencelane: %w", err)
	}
	if err := validate.Record(PreferenceLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceLanes returns the first problem found in xs.
func ValidatePreferenceLanes(xs []PreferenceLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceSlotSchema.
func (x PreferenceSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferenceslot: %w", err)
	}
	if err := validate.Record(PreferenceSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferenceslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceSlots returns the first problem found in xs.
func ValidatePreferenceSlots(xs []PreferenceSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceTagSchema.
func (x PreferenceTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencetag: %w", err)
	}
	if err := validate.Record(PreferenceTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceTags returns the first problem found in xs.
func ValidatePreferenceTags(xs []PreferenceTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceRuleSchema.
func (x PreferenceRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencerule: %w", err)
	}
	if err := validate.Record(PreferenceRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceRules returns the first problem found in xs.
func ValidatePreferenceRules(xs []PreferenceRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceHoldSchema.
func (x PreferenceHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencehold: %w", err)
	}
	if err := validate.Record(PreferenceHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceHolds returns the first problem found in xs.
func ValidatePreferenceHolds(xs []PreferenceHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceClaimSchema.
func (x PreferenceClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferenceclaim: %w", err)
	}
	if err := validate.Record(PreferenceClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferenceclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceClaims returns the first problem found in xs.
func ValidatePreferenceClaims(xs []PreferenceClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against PreferenceCursorSchema.
func (x PreferenceCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("preference: preferencecursor: %w", err)
	}
	if err := validate.Record(PreferenceCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("preference: preferencecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidatePreferenceCursors returns the first problem found in xs.
func ValidatePreferenceCursors(xs []PreferenceCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
