package quota

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded quota records against this package's schemas.
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
		return fmt.Errorf("quota: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against QuotaRecordSchema.
func (x QuotaRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotarecord: %w", err)
	}
	if err := validate.Record(QuotaRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotarecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaRecords returns the first problem found in xs.
func ValidateQuotaRecords(xs []QuotaRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaEntrySchema.
func (x QuotaEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotaentry: %w", err)
	}
	if err := validate.Record(QuotaEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotaentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaEntrys returns the first problem found in xs.
func ValidateQuotaEntrys(xs []QuotaEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaItemSchema.
func (x QuotaItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotaitem: %w", err)
	}
	if err := validate.Record(QuotaItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotaitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaItems returns the first problem found in xs.
func ValidateQuotaItems(xs []QuotaItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaDraftSchema.
func (x QuotaDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotadraft: %w", err)
	}
	if err := validate.Record(QuotaDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotadraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaDrafts returns the first problem found in xs.
func ValidateQuotaDrafts(xs []QuotaDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaRevisionSchema.
func (x QuotaRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotarevision: %w", err)
	}
	if err := validate.Record(QuotaRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotarevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaRevisions returns the first problem found in xs.
func ValidateQuotaRevisions(xs []QuotaRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaSnapshotSchema.
func (x QuotaSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotasnapshot: %w", err)
	}
	if err := validate.Record(QuotaSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotasnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaSnapshots returns the first problem found in xs.
func ValidateQuotaSnapshots(xs []QuotaSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaBatchSchema.
func (x QuotaBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotabatch: %w", err)
	}
	if err := validate.Record(QuotaBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotabatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaBatchs returns the first problem found in xs.
func ValidateQuotaBatchs(xs []QuotaBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaRequestSchema.
func (x QuotaRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotarequest: %w", err)
	}
	if err := validate.Record(QuotaRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotarequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaRequests returns the first problem found in xs.
func ValidateQuotaRequests(xs []QuotaRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaSummarySchema.
func (x QuotaSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotasummary: %w", err)
	}
	if err := validate.Record(QuotaSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotasummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaSummarys returns the first problem found in xs.
func ValidateQuotaSummarys(xs []QuotaSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaPolicySchema.
func (x QuotaPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotapolicy: %w", err)
	}
	if err := validate.Record(QuotaPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotapolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaPolicys returns the first problem found in xs.
func ValidateQuotaPolicys(xs []QuotaPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaLinkSchema.
func (x QuotaLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotalink: %w", err)
	}
	if err := validate.Record(QuotaLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotalink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaLinks returns the first problem found in xs.
func ValidateQuotaLinks(xs []QuotaLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaNoteSchema.
func (x QuotaNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotanote: %w", err)
	}
	if err := validate.Record(QuotaNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotanote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaNotes returns the first problem found in xs.
func ValidateQuotaNotes(xs []QuotaNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaWindowSchema.
func (x QuotaWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotawindow: %w", err)
	}
	if err := validate.Record(QuotaWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotawindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaWindows returns the first problem found in xs.
func ValidateQuotaWindows(xs []QuotaWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaQuotaSchema.
func (x QuotaQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotaquota: %w", err)
	}
	if err := validate.Record(QuotaQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotaquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaQuotas returns the first problem found in xs.
func ValidateQuotaQuotas(xs []QuotaQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaBucketSchema.
func (x QuotaBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotabucket: %w", err)
	}
	if err := validate.Record(QuotaBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotabucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaBuckets returns the first problem found in xs.
func ValidateQuotaBuckets(xs []QuotaBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaLaneSchema.
func (x QuotaLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotalane: %w", err)
	}
	if err := validate.Record(QuotaLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotalane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaLanes returns the first problem found in xs.
func ValidateQuotaLanes(xs []QuotaLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaSlotSchema.
func (x QuotaSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotaslot: %w", err)
	}
	if err := validate.Record(QuotaSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotaslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaSlots returns the first problem found in xs.
func ValidateQuotaSlots(xs []QuotaSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaTagSchema.
func (x QuotaTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotatag: %w", err)
	}
	if err := validate.Record(QuotaTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotatag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaTags returns the first problem found in xs.
func ValidateQuotaTags(xs []QuotaTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaRuleSchema.
func (x QuotaRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotarule: %w", err)
	}
	if err := validate.Record(QuotaRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotarule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaRules returns the first problem found in xs.
func ValidateQuotaRules(xs []QuotaRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaHoldSchema.
func (x QuotaHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotahold: %w", err)
	}
	if err := validate.Record(QuotaHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotahold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaHolds returns the first problem found in xs.
func ValidateQuotaHolds(xs []QuotaHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaClaimSchema.
func (x QuotaClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotaclaim: %w", err)
	}
	if err := validate.Record(QuotaClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotaclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaClaims returns the first problem found in xs.
func ValidateQuotaClaims(xs []QuotaClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against QuotaCursorSchema.
func (x QuotaCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("quota: quotacursor: %w", err)
	}
	if err := validate.Record(QuotaCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("quota: quotacursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateQuotaCursors returns the first problem found in xs.
func ValidateQuotaCursors(xs []QuotaCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
