package main

import (
	"fmt"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/audit"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/billing"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/inventory"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/notification"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/payment"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/session"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/shipping"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/webhook"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/queue"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/retry"
)

// seedTime is the timestamp written onto every record the worker creates from
// a scripted job. Scripted work is not real work, so it does not get the
// current time.
var seedTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

// handle applies one job to the repository that owns its kind.
//
// An unknown kind is a permanent failure: retrying it cannot help, and the
// course's flake-detection step relies on retries only happening for things
// that could actually succeed on a second attempt.
func (r *repos) handle(job queue.Job) error {
	id := job.Payload["id"]
	if id == "" {
		return retry.Permanent(fmt.Errorf("worker: job %s has no id", job.ID))
	}

	var err error
	switch job.Kind {
	case "payment":
		_, err = r.payment.PutPaymentRecord(payment.SamplePaymentRecord(id, seedTime), 0)
	case "inventory":
		_, err = r.inventory.PutInventoryRecord(inventory.SampleInventoryRecord(id, seedTime), 0)
	case "shipping":
		_, err = r.shipping.PutShippingRecord(shipping.SampleShippingRecord(id, seedTime), 0)
	case "notification":
		_, err = r.notification.PutNotificationRecord(notification.SampleNotificationRecord(id, seedTime), 0)
	case "audit":
		_, err = r.audit.PutAuditRecord(audit.SampleAuditRecord(id, seedTime), 0)
	case "billing":
		_, err = r.billing.PutBillingRecord(billing.SampleBillingRecord(id, seedTime), 0)
	case "session":
		_, err = r.session.PutSessionRecord(session.SampleSessionRecord(id, seedTime), 0)
	case "webhook":
		_, err = r.webhook.PutWebhookRecord(webhook.SampleWebhookRecord(id, seedTime), 0)
	default:
		return retry.Permanent(fmt.Errorf("worker: unknown job kind %q", job.Kind))
	}
	if err != nil {
		return fmt.Errorf("worker: %s: %w", job.Kind, err)
	}
	return nil
}
