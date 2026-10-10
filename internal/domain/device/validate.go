package device

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded device records against this package's schemas.
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
		return fmt.Errorf("device: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against DeviceRecordSchema.
func (x DeviceRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicerecord: %w", err)
	}
	if err := validate.Record(DeviceRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicerecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceRecords returns the first problem found in xs.
func ValidateDeviceRecords(xs []DeviceRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceEntrySchema.
func (x DeviceEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: deviceentry: %w", err)
	}
	if err := validate.Record(DeviceEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: deviceentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceEntrys returns the first problem found in xs.
func ValidateDeviceEntrys(xs []DeviceEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceItemSchema.
func (x DeviceItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: deviceitem: %w", err)
	}
	if err := validate.Record(DeviceItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: deviceitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceItems returns the first problem found in xs.
func ValidateDeviceItems(xs []DeviceItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceDraftSchema.
func (x DeviceDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicedraft: %w", err)
	}
	if err := validate.Record(DeviceDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicedraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceDrafts returns the first problem found in xs.
func ValidateDeviceDrafts(xs []DeviceDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceRevisionSchema.
func (x DeviceRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicerevision: %w", err)
	}
	if err := validate.Record(DeviceRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicerevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceRevisions returns the first problem found in xs.
func ValidateDeviceRevisions(xs []DeviceRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceSnapshotSchema.
func (x DeviceSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicesnapshot: %w", err)
	}
	if err := validate.Record(DeviceSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicesnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceSnapshots returns the first problem found in xs.
func ValidateDeviceSnapshots(xs []DeviceSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceBatchSchema.
func (x DeviceBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicebatch: %w", err)
	}
	if err := validate.Record(DeviceBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicebatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceBatchs returns the first problem found in xs.
func ValidateDeviceBatchs(xs []DeviceBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceRequestSchema.
func (x DeviceRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicerequest: %w", err)
	}
	if err := validate.Record(DeviceRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicerequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceRequests returns the first problem found in xs.
func ValidateDeviceRequests(xs []DeviceRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceSummarySchema.
func (x DeviceSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicesummary: %w", err)
	}
	if err := validate.Record(DeviceSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicesummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceSummarys returns the first problem found in xs.
func ValidateDeviceSummarys(xs []DeviceSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DevicePolicySchema.
func (x DevicePolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicepolicy: %w", err)
	}
	if err := validate.Record(DevicePolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicepolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDevicePolicys returns the first problem found in xs.
func ValidateDevicePolicys(xs []DevicePolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceLinkSchema.
func (x DeviceLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicelink: %w", err)
	}
	if err := validate.Record(DeviceLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicelink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceLinks returns the first problem found in xs.
func ValidateDeviceLinks(xs []DeviceLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceNoteSchema.
func (x DeviceNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicenote: %w", err)
	}
	if err := validate.Record(DeviceNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicenote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceNotes returns the first problem found in xs.
func ValidateDeviceNotes(xs []DeviceNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceWindowSchema.
func (x DeviceWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicewindow: %w", err)
	}
	if err := validate.Record(DeviceWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicewindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceWindows returns the first problem found in xs.
func ValidateDeviceWindows(xs []DeviceWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceQuotaSchema.
func (x DeviceQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicequota: %w", err)
	}
	if err := validate.Record(DeviceQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicequota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceQuotas returns the first problem found in xs.
func ValidateDeviceQuotas(xs []DeviceQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceBucketSchema.
func (x DeviceBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicebucket: %w", err)
	}
	if err := validate.Record(DeviceBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicebucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceBuckets returns the first problem found in xs.
func ValidateDeviceBuckets(xs []DeviceBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceLaneSchema.
func (x DeviceLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicelane: %w", err)
	}
	if err := validate.Record(DeviceLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicelane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceLanes returns the first problem found in xs.
func ValidateDeviceLanes(xs []DeviceLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceSlotSchema.
func (x DeviceSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: deviceslot: %w", err)
	}
	if err := validate.Record(DeviceSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: deviceslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceSlots returns the first problem found in xs.
func ValidateDeviceSlots(xs []DeviceSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceTagSchema.
func (x DeviceTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicetag: %w", err)
	}
	if err := validate.Record(DeviceTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicetag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceTags returns the first problem found in xs.
func ValidateDeviceTags(xs []DeviceTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceRuleSchema.
func (x DeviceRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicerule: %w", err)
	}
	if err := validate.Record(DeviceRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicerule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceRules returns the first problem found in xs.
func ValidateDeviceRules(xs []DeviceRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceHoldSchema.
func (x DeviceHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicehold: %w", err)
	}
	if err := validate.Record(DeviceHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicehold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceHolds returns the first problem found in xs.
func ValidateDeviceHolds(xs []DeviceHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceClaimSchema.
func (x DeviceClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: deviceclaim: %w", err)
	}
	if err := validate.Record(DeviceClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: deviceclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceClaims returns the first problem found in xs.
func ValidateDeviceClaims(xs []DeviceClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against DeviceCursorSchema.
func (x DeviceCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("device: devicecursor: %w", err)
	}
	if err := validate.Record(DeviceCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("device: devicecursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateDeviceCursors returns the first problem found in xs.
func ValidateDeviceCursors(xs []DeviceCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
