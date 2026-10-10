package notification

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded notification records against this package's schemas.
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
		return fmt.Errorf("notification: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against NotificationRecordSchema.
func (x NotificationRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationrecord: %w", err)
	}
	if err := validate.Record(NotificationRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationRecords returns the first problem found in xs.
func ValidateNotificationRecords(xs []NotificationRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationEntrySchema.
func (x NotificationEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationentry: %w", err)
	}
	if err := validate.Record(NotificationEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationEntrys returns the first problem found in xs.
func ValidateNotificationEntrys(xs []NotificationEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationItemSchema.
func (x NotificationItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationitem: %w", err)
	}
	if err := validate.Record(NotificationItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationItems returns the first problem found in xs.
func ValidateNotificationItems(xs []NotificationItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationDraftSchema.
func (x NotificationDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationdraft: %w", err)
	}
	if err := validate.Record(NotificationDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationDrafts returns the first problem found in xs.
func ValidateNotificationDrafts(xs []NotificationDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationRevisionSchema.
func (x NotificationRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationrevision: %w", err)
	}
	if err := validate.Record(NotificationRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationRevisions returns the first problem found in xs.
func ValidateNotificationRevisions(xs []NotificationRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationSnapshotSchema.
func (x NotificationSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationsnapshot: %w", err)
	}
	if err := validate.Record(NotificationSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationsnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationSnapshots returns the first problem found in xs.
func ValidateNotificationSnapshots(xs []NotificationSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationBatchSchema.
func (x NotificationBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationbatch: %w", err)
	}
	if err := validate.Record(NotificationBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationBatchs returns the first problem found in xs.
func ValidateNotificationBatchs(xs []NotificationBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationRequestSchema.
func (x NotificationRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationrequest: %w", err)
	}
	if err := validate.Record(NotificationRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationRequests returns the first problem found in xs.
func ValidateNotificationRequests(xs []NotificationRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationSummarySchema.
func (x NotificationSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationsummary: %w", err)
	}
	if err := validate.Record(NotificationSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationsummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationSummarys returns the first problem found in xs.
func ValidateNotificationSummarys(xs []NotificationSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationPolicySchema.
func (x NotificationPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationpolicy: %w", err)
	}
	if err := validate.Record(NotificationPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationPolicys returns the first problem found in xs.
func ValidateNotificationPolicys(xs []NotificationPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationLinkSchema.
func (x NotificationLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationlink: %w", err)
	}
	if err := validate.Record(NotificationLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationlink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationLinks returns the first problem found in xs.
func ValidateNotificationLinks(xs []NotificationLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationNoteSchema.
func (x NotificationNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationnote: %w", err)
	}
	if err := validate.Record(NotificationNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationnote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationNotes returns the first problem found in xs.
func ValidateNotificationNotes(xs []NotificationNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationWindowSchema.
func (x NotificationWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationwindow: %w", err)
	}
	if err := validate.Record(NotificationWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationWindows returns the first problem found in xs.
func ValidateNotificationWindows(xs []NotificationWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationQuotaSchema.
func (x NotificationQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationquota: %w", err)
	}
	if err := validate.Record(NotificationQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationQuotas returns the first problem found in xs.
func ValidateNotificationQuotas(xs []NotificationQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationBucketSchema.
func (x NotificationBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationbucket: %w", err)
	}
	if err := validate.Record(NotificationBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationBuckets returns the first problem found in xs.
func ValidateNotificationBuckets(xs []NotificationBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationLaneSchema.
func (x NotificationLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationlane: %w", err)
	}
	if err := validate.Record(NotificationLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationlane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationLanes returns the first problem found in xs.
func ValidateNotificationLanes(xs []NotificationLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationSlotSchema.
func (x NotificationSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationslot: %w", err)
	}
	if err := validate.Record(NotificationSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationSlots returns the first problem found in xs.
func ValidateNotificationSlots(xs []NotificationSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationTagSchema.
func (x NotificationTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationtag: %w", err)
	}
	if err := validate.Record(NotificationTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationtag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationTags returns the first problem found in xs.
func ValidateNotificationTags(xs []NotificationTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationRuleSchema.
func (x NotificationRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationrule: %w", err)
	}
	if err := validate.Record(NotificationRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationRules returns the first problem found in xs.
func ValidateNotificationRules(xs []NotificationRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationHoldSchema.
func (x NotificationHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationhold: %w", err)
	}
	if err := validate.Record(NotificationHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationHolds returns the first problem found in xs.
func ValidateNotificationHolds(xs []NotificationHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationClaimSchema.
func (x NotificationClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationclaim: %w", err)
	}
	if err := validate.Record(NotificationClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationClaims returns the first problem found in xs.
func ValidateNotificationClaims(xs []NotificationClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against NotificationCursorSchema.
func (x NotificationCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("notification: notificationcursor: %w", err)
	}
	if err := validate.Record(NotificationCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("notification: notificationcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateNotificationCursors returns the first problem found in xs.
func ValidateNotificationCursors(xs []NotificationCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
