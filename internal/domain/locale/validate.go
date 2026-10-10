package locale

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded locale records against this package's schemas.
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
		return fmt.Errorf("locale: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against LocaleRecordSchema.
func (x LocaleRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localerecord: %w", err)
	}
	if err := validate.Record(LocaleRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleRecords returns the first problem found in xs.
func ValidateLocaleRecords(xs []LocaleRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleEntrySchema.
func (x LocaleEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localeentry: %w", err)
	}
	if err := validate.Record(LocaleEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localeentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleEntrys returns the first problem found in xs.
func ValidateLocaleEntrys(xs []LocaleEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleItemSchema.
func (x LocaleItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localeitem: %w", err)
	}
	if err := validate.Record(LocaleItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localeitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleItems returns the first problem found in xs.
func ValidateLocaleItems(xs []LocaleItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleDraftSchema.
func (x LocaleDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localedraft: %w", err)
	}
	if err := validate.Record(LocaleDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleDrafts returns the first problem found in xs.
func ValidateLocaleDrafts(xs []LocaleDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleRevisionSchema.
func (x LocaleRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localerevision: %w", err)
	}
	if err := validate.Record(LocaleRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleRevisions returns the first problem found in xs.
func ValidateLocaleRevisions(xs []LocaleRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleSnapshotSchema.
func (x LocaleSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localesnapshot: %w", err)
	}
	if err := validate.Record(LocaleSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleSnapshots returns the first problem found in xs.
func ValidateLocaleSnapshots(xs []LocaleSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleBatchSchema.
func (x LocaleBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localebatch: %w", err)
	}
	if err := validate.Record(LocaleBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleBatchs returns the first problem found in xs.
func ValidateLocaleBatchs(xs []LocaleBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleRequestSchema.
func (x LocaleRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localerequest: %w", err)
	}
	if err := validate.Record(LocaleRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleRequests returns the first problem found in xs.
func ValidateLocaleRequests(xs []LocaleRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleSummarySchema.
func (x LocaleSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localesummary: %w", err)
	}
	if err := validate.Record(LocaleSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleSummarys returns the first problem found in xs.
func ValidateLocaleSummarys(xs []LocaleSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocalePolicySchema.
func (x LocalePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localepolicy: %w", err)
	}
	if err := validate.Record(LocalePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocalePolicys returns the first problem found in xs.
func ValidateLocalePolicys(xs []LocalePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleLinkSchema.
func (x LocaleLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localelink: %w", err)
	}
	if err := validate.Record(LocaleLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleLinks returns the first problem found in xs.
func ValidateLocaleLinks(xs []LocaleLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleNoteSchema.
func (x LocaleNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localenote: %w", err)
	}
	if err := validate.Record(LocaleNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleNotes returns the first problem found in xs.
func ValidateLocaleNotes(xs []LocaleNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleWindowSchema.
func (x LocaleWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localewindow: %w", err)
	}
	if err := validate.Record(LocaleWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleWindows returns the first problem found in xs.
func ValidateLocaleWindows(xs []LocaleWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleQuotaSchema.
func (x LocaleQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localequota: %w", err)
	}
	if err := validate.Record(LocaleQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleQuotas returns the first problem found in xs.
func ValidateLocaleQuotas(xs []LocaleQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleBucketSchema.
func (x LocaleBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localebucket: %w", err)
	}
	if err := validate.Record(LocaleBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleBuckets returns the first problem found in xs.
func ValidateLocaleBuckets(xs []LocaleBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleLaneSchema.
func (x LocaleLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localelane: %w", err)
	}
	if err := validate.Record(LocaleLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleLanes returns the first problem found in xs.
func ValidateLocaleLanes(xs []LocaleLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleSlotSchema.
func (x LocaleSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localeslot: %w", err)
	}
	if err := validate.Record(LocaleSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localeslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleSlots returns the first problem found in xs.
func ValidateLocaleSlots(xs []LocaleSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleTagSchema.
func (x LocaleTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localetag: %w", err)
	}
	if err := validate.Record(LocaleTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleTags returns the first problem found in xs.
func ValidateLocaleTags(xs []LocaleTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleRuleSchema.
func (x LocaleRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localerule: %w", err)
	}
	if err := validate.Record(LocaleRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleRules returns the first problem found in xs.
func ValidateLocaleRules(xs []LocaleRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleHoldSchema.
func (x LocaleHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localehold: %w", err)
	}
	if err := validate.Record(LocaleHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleHolds returns the first problem found in xs.
func ValidateLocaleHolds(xs []LocaleHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleClaimSchema.
func (x LocaleClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localeclaim: %w", err)
	}
	if err := validate.Record(LocaleClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localeclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleClaims returns the first problem found in xs.
func ValidateLocaleClaims(xs []LocaleClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against LocaleCursorSchema.
func (x LocaleCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("locale: localecursor: %w", err)
	}
	if err := validate.Record(LocaleCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("locale: localecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateLocaleCursors returns the first problem found in xs.
func ValidateLocaleCursors(xs []LocaleCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
