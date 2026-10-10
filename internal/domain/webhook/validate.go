package webhook

import (
	"fmt"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/validate"
)

// Validator checks encoded webhook records against this package's schemas.
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
		return fmt.Errorf("webhook: unknown schema %q", name)
	}
	return nil
}

// Validate checks x against WebhookRecordSchema.
func (x WebhookRecord) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookrecord: %w", err)
	}
	if err := validate.Record(WebhookRecordSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookrecord %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookRecords returns the first problem found in xs.
func ValidateWebhookRecords(xs []WebhookRecord) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookEntrySchema.
func (x WebhookEntry) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookentry: %w", err)
	}
	if err := validate.Record(WebhookEntrySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookentry %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookEntrys returns the first problem found in xs.
func ValidateWebhookEntrys(xs []WebhookEntry) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookItemSchema.
func (x WebhookItem) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookitem: %w", err)
	}
	if err := validate.Record(WebhookItemSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookitem %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookItems returns the first problem found in xs.
func ValidateWebhookItems(xs []WebhookItem) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookDraftSchema.
func (x WebhookDraft) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookdraft: %w", err)
	}
	if err := validate.Record(WebhookDraftSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookdraft %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookDrafts returns the first problem found in xs.
func ValidateWebhookDrafts(xs []WebhookDraft) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookRevisionSchema.
func (x WebhookRevision) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookrevision: %w", err)
	}
	if err := validate.Record(WebhookRevisionSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookrevision %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookRevisions returns the first problem found in xs.
func ValidateWebhookRevisions(xs []WebhookRevision) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookSnapshotSchema.
func (x WebhookSnapshot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhooksnapshot: %w", err)
	}
	if err := validate.Record(WebhookSnapshotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhooksnapshot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookSnapshots returns the first problem found in xs.
func ValidateWebhookSnapshots(xs []WebhookSnapshot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookBatchSchema.
func (x WebhookBatch) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookbatch: %w", err)
	}
	if err := validate.Record(WebhookBatchSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookbatch %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookBatchs returns the first problem found in xs.
func ValidateWebhookBatchs(xs []WebhookBatch) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookRequestSchema.
func (x WebhookRequest) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookrequest: %w", err)
	}
	if err := validate.Record(WebhookRequestSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookrequest %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookRequests returns the first problem found in xs.
func ValidateWebhookRequests(xs []WebhookRequest) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookSummarySchema.
func (x WebhookSummary) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhooksummary: %w", err)
	}
	if err := validate.Record(WebhookSummarySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhooksummary %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookSummarys returns the first problem found in xs.
func ValidateWebhookSummarys(xs []WebhookSummary) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookPolicySchema.
func (x WebhookPolicy) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookpolicy: %w", err)
	}
	if err := validate.Record(WebhookPolicySchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookpolicy %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookPolicys returns the first problem found in xs.
func ValidateWebhookPolicys(xs []WebhookPolicy) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookLinkSchema.
func (x WebhookLink) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhooklink: %w", err)
	}
	if err := validate.Record(WebhookLinkSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhooklink %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookLinks returns the first problem found in xs.
func ValidateWebhookLinks(xs []WebhookLink) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookNoteSchema.
func (x WebhookNote) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhooknote: %w", err)
	}
	if err := validate.Record(WebhookNoteSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhooknote %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookNotes returns the first problem found in xs.
func ValidateWebhookNotes(xs []WebhookNote) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookWindowSchema.
func (x WebhookWindow) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookwindow: %w", err)
	}
	if err := validate.Record(WebhookWindowSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookwindow %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookWindows returns the first problem found in xs.
func ValidateWebhookWindows(xs []WebhookWindow) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookQuotaSchema.
func (x WebhookQuota) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookquota: %w", err)
	}
	if err := validate.Record(WebhookQuotaSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookquota %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookQuotas returns the first problem found in xs.
func ValidateWebhookQuotas(xs []WebhookQuota) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookBucketSchema.
func (x WebhookBucket) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookbucket: %w", err)
	}
	if err := validate.Record(WebhookBucketSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookbucket %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookBuckets returns the first problem found in xs.
func ValidateWebhookBuckets(xs []WebhookBucket) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookLaneSchema.
func (x WebhookLane) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhooklane: %w", err)
	}
	if err := validate.Record(WebhookLaneSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhooklane %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookLanes returns the first problem found in xs.
func ValidateWebhookLanes(xs []WebhookLane) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookSlotSchema.
func (x WebhookSlot) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookslot: %w", err)
	}
	if err := validate.Record(WebhookSlotSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookslot %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookSlots returns the first problem found in xs.
func ValidateWebhookSlots(xs []WebhookSlot) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookTagSchema.
func (x WebhookTag) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhooktag: %w", err)
	}
	if err := validate.Record(WebhookTagSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhooktag %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookTags returns the first problem found in xs.
func ValidateWebhookTags(xs []WebhookTag) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookRuleSchema.
func (x WebhookRule) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookrule: %w", err)
	}
	if err := validate.Record(WebhookRuleSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookrule %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookRules returns the first problem found in xs.
func ValidateWebhookRules(xs []WebhookRule) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookHoldSchema.
func (x WebhookHold) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookhold: %w", err)
	}
	if err := validate.Record(WebhookHoldSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookhold %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookHolds returns the first problem found in xs.
func ValidateWebhookHolds(xs []WebhookHold) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookClaimSchema.
func (x WebhookClaim) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookclaim: %w", err)
	}
	if err := validate.Record(WebhookClaimSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookclaim %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookClaims returns the first problem found in xs.
func ValidateWebhookClaims(xs []WebhookClaim) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks x against WebhookCursorSchema.
func (x WebhookCursor) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("webhook: webhookcursor: %w", err)
	}
	if err := validate.Record(WebhookCursorSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("webhook: webhookcursor %s: %w", x.ID, err)
	}
	return nil
}

// ValidateWebhookCursors returns the first problem found in xs.
func ValidateWebhookCursors(xs []WebhookCursor) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}
