package forecast

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded forecast records against this package's schemas.
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
		return fmt.Errorf("forecast: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against ForecastRecordSchema.
func (x ForecastRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastrecord: %w", err)
	}
	if err := validate.Record(ForecastRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastRecords returns the first problem found in xs.
func ValidateForecastRecords(xs []ForecastRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastEntrySchema.
func (x ForecastEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastentry: %w", err)
	}
	if err := validate.Record(ForecastEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastEntrys returns the first problem found in xs.
func ValidateForecastEntrys(xs []ForecastEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastItemSchema.
func (x ForecastItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastitem: %w", err)
	}
	if err := validate.Record(ForecastItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastItems returns the first problem found in xs.
func ValidateForecastItems(xs []ForecastItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastDraftSchema.
func (x ForecastDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastdraft: %w", err)
	}
	if err := validate.Record(ForecastDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastDrafts returns the first problem found in xs.
func ValidateForecastDrafts(xs []ForecastDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastRevisionSchema.
func (x ForecastRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastrevision: %w", err)
	}
	if err := validate.Record(ForecastRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastRevisions returns the first problem found in xs.
func ValidateForecastRevisions(xs []ForecastRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastSnapshotSchema.
func (x ForecastSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastsnapshot: %w", err)
	}
	if err := validate.Record(ForecastSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastSnapshots returns the first problem found in xs.
func ValidateForecastSnapshots(xs []ForecastSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastBatchSchema.
func (x ForecastBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastbatch: %w", err)
	}
	if err := validate.Record(ForecastBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastBatchs returns the first problem found in xs.
func ValidateForecastBatchs(xs []ForecastBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastRequestSchema.
func (x ForecastRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastrequest: %w", err)
	}
	if err := validate.Record(ForecastRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastRequests returns the first problem found in xs.
func ValidateForecastRequests(xs []ForecastRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastSummarySchema.
func (x ForecastSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastsummary: %w", err)
	}
	if err := validate.Record(ForecastSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastSummarys returns the first problem found in xs.
func ValidateForecastSummarys(xs []ForecastSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastPolicySchema.
func (x ForecastPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastpolicy: %w", err)
	}
	if err := validate.Record(ForecastPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastPolicys returns the first problem found in xs.
func ValidateForecastPolicys(xs []ForecastPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastLinkSchema.
func (x ForecastLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastlink: %w", err)
	}
	if err := validate.Record(ForecastLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastLinks returns the first problem found in xs.
func ValidateForecastLinks(xs []ForecastLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastNoteSchema.
func (x ForecastNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastnote: %w", err)
	}
	if err := validate.Record(ForecastNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastNotes returns the first problem found in xs.
func ValidateForecastNotes(xs []ForecastNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastWindowSchema.
func (x ForecastWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastwindow: %w", err)
	}
	if err := validate.Record(ForecastWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastWindows returns the first problem found in xs.
func ValidateForecastWindows(xs []ForecastWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastQuotaSchema.
func (x ForecastQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastquota: %w", err)
	}
	if err := validate.Record(ForecastQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastQuotas returns the first problem found in xs.
func ValidateForecastQuotas(xs []ForecastQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastBucketSchema.
func (x ForecastBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastbucket: %w", err)
	}
	if err := validate.Record(ForecastBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastBuckets returns the first problem found in xs.
func ValidateForecastBuckets(xs []ForecastBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastLaneSchema.
func (x ForecastLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastlane: %w", err)
	}
	if err := validate.Record(ForecastLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastLanes returns the first problem found in xs.
func ValidateForecastLanes(xs []ForecastLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastSlotSchema.
func (x ForecastSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastslot: %w", err)
	}
	if err := validate.Record(ForecastSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastSlots returns the first problem found in xs.
func ValidateForecastSlots(xs []ForecastSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastTagSchema.
func (x ForecastTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecasttag: %w", err)
	}
	if err := validate.Record(ForecastTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecasttag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastTags returns the first problem found in xs.
func ValidateForecastTags(xs []ForecastTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastRuleSchema.
func (x ForecastRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastrule: %w", err)
	}
	if err := validate.Record(ForecastRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastRules returns the first problem found in xs.
func ValidateForecastRules(xs []ForecastRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastHoldSchema.
func (x ForecastHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecasthold: %w", err)
	}
	if err := validate.Record(ForecastHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecasthold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastHolds returns the first problem found in xs.
func ValidateForecastHolds(xs []ForecastHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastClaimSchema.
func (x ForecastClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastclaim: %w", err)
	}
	if err := validate.Record(ForecastClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastClaims returns the first problem found in xs.
func ValidateForecastClaims(xs []ForecastClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against ForecastCursorSchema.
func (x ForecastCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("forecast: forecastcursor: %w", err)
	}
	if err := validate.Record(ForecastCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("forecast: forecastcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateForecastCursors returns the first problem found in xs.
func ValidateForecastCursors(xs []ForecastCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
