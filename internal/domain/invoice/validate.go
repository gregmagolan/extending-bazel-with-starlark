package invoice

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded invoice records against this package's schemas.
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
		return fmt.Errorf("invoice: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against InvoiceRecordSchema.
func (x InvoiceRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicerecord: %w", err)
	}
	if err := validate.Record(InvoiceRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceRecords returns the first problem found in xs.
func ValidateInvoiceRecords(xs []InvoiceRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceEntrySchema.
func (x InvoiceEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoiceentry: %w", err)
	}
	if err := validate.Record(InvoiceEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoiceentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceEntrys returns the first problem found in xs.
func ValidateInvoiceEntrys(xs []InvoiceEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceItemSchema.
func (x InvoiceItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoiceitem: %w", err)
	}
	if err := validate.Record(InvoiceItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoiceitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceItems returns the first problem found in xs.
func ValidateInvoiceItems(xs []InvoiceItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceDraftSchema.
func (x InvoiceDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicedraft: %w", err)
	}
	if err := validate.Record(InvoiceDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceDrafts returns the first problem found in xs.
func ValidateInvoiceDrafts(xs []InvoiceDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceRevisionSchema.
func (x InvoiceRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicerevision: %w", err)
	}
	if err := validate.Record(InvoiceRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceRevisions returns the first problem found in xs.
func ValidateInvoiceRevisions(xs []InvoiceRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceSnapshotSchema.
func (x InvoiceSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicesnapshot: %w", err)
	}
	if err := validate.Record(InvoiceSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceSnapshots returns the first problem found in xs.
func ValidateInvoiceSnapshots(xs []InvoiceSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceBatchSchema.
func (x InvoiceBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicebatch: %w", err)
	}
	if err := validate.Record(InvoiceBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceBatchs returns the first problem found in xs.
func ValidateInvoiceBatchs(xs []InvoiceBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceRequestSchema.
func (x InvoiceRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicerequest: %w", err)
	}
	if err := validate.Record(InvoiceRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceRequests returns the first problem found in xs.
func ValidateInvoiceRequests(xs []InvoiceRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceSummarySchema.
func (x InvoiceSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicesummary: %w", err)
	}
	if err := validate.Record(InvoiceSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceSummarys returns the first problem found in xs.
func ValidateInvoiceSummarys(xs []InvoiceSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoicePolicySchema.
func (x InvoicePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicepolicy: %w", err)
	}
	if err := validate.Record(InvoicePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoicePolicys returns the first problem found in xs.
func ValidateInvoicePolicys(xs []InvoicePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceLinkSchema.
func (x InvoiceLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicelink: %w", err)
	}
	if err := validate.Record(InvoiceLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceLinks returns the first problem found in xs.
func ValidateInvoiceLinks(xs []InvoiceLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceNoteSchema.
func (x InvoiceNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicenote: %w", err)
	}
	if err := validate.Record(InvoiceNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceNotes returns the first problem found in xs.
func ValidateInvoiceNotes(xs []InvoiceNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceWindowSchema.
func (x InvoiceWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicewindow: %w", err)
	}
	if err := validate.Record(InvoiceWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceWindows returns the first problem found in xs.
func ValidateInvoiceWindows(xs []InvoiceWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceQuotaSchema.
func (x InvoiceQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicequota: %w", err)
	}
	if err := validate.Record(InvoiceQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceQuotas returns the first problem found in xs.
func ValidateInvoiceQuotas(xs []InvoiceQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceBucketSchema.
func (x InvoiceBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicebucket: %w", err)
	}
	if err := validate.Record(InvoiceBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceBuckets returns the first problem found in xs.
func ValidateInvoiceBuckets(xs []InvoiceBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceLaneSchema.
func (x InvoiceLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicelane: %w", err)
	}
	if err := validate.Record(InvoiceLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceLanes returns the first problem found in xs.
func ValidateInvoiceLanes(xs []InvoiceLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceSlotSchema.
func (x InvoiceSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoiceslot: %w", err)
	}
	if err := validate.Record(InvoiceSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoiceslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceSlots returns the first problem found in xs.
func ValidateInvoiceSlots(xs []InvoiceSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceTagSchema.
func (x InvoiceTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicetag: %w", err)
	}
	if err := validate.Record(InvoiceTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceTags returns the first problem found in xs.
func ValidateInvoiceTags(xs []InvoiceTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceRuleSchema.
func (x InvoiceRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicerule: %w", err)
	}
	if err := validate.Record(InvoiceRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceRules returns the first problem found in xs.
func ValidateInvoiceRules(xs []InvoiceRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceHoldSchema.
func (x InvoiceHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicehold: %w", err)
	}
	if err := validate.Record(InvoiceHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceHolds returns the first problem found in xs.
func ValidateInvoiceHolds(xs []InvoiceHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceClaimSchema.
func (x InvoiceClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoiceclaim: %w", err)
	}
	if err := validate.Record(InvoiceClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoiceclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceClaims returns the first problem found in xs.
func ValidateInvoiceClaims(xs []InvoiceClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against InvoiceCursorSchema.
func (x InvoiceCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("invoice: invoicecursor: %w", err)
	}
	if err := validate.Record(InvoiceCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("invoice: invoicecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateInvoiceCursors returns the first problem found in xs.
func ValidateInvoiceCursors(xs []InvoiceCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
