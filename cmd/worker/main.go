// Command worker drains the queue, applying each job to the domain
// repositories that own the affected records.
//
// Like the server, it runs a scripted batch and exits rather than looping
// forever. Its dependency set is deliberately different from the server's:
// the worker owns the write-side domain packages and never touches lib/api.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/audit"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/billing"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/inventory"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/notification"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/payment"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/session"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/shipping"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/webhook"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/queue"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/retry"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

func main() {
	settings := flag.String("config", "", "inline key=value configuration")
	jobs := flag.Int("jobs", 8, "number of scripted jobs to enqueue and drain")
	flag.Parse()

	if err := run(*settings, *jobs); err != nil {
		fmt.Fprintln(os.Stderr, "worker:", err)
		os.Exit(1)
	}
}

func run(settings string, jobs int) error {
	cfg, err := config.Load(settings)
	if err != nil {
		return err
	}
	if err := cfg.Check(); err != nil {
		return err
	}

	level, err := cfg.LogLevel()
	if err != nil {
		return err
	}
	l := log.New(os.Stderr, level).With("service", "worker")

	reg := metrics.NewRegistry(nil)
	st, err := store.Open(l, cfg, reg)
	if err != nil {
		return err
	}

	depth, err := cfg.Int("queue.depth", 256)
	if err != nil {
		return err
	}
	lease, err := cfg.Duration("queue.lease", 0)
	if err != nil {
		return err
	}
	q := queue.New(l, reg, queue.Options{Depth: depth, Lease: lease, Store: st})

	repos := newRepos(l, reg, st)
	l.Info("worker ready", "kinds", strings.Join(repos.kinds(), ","))

	for i := 0; i < jobs; i++ {
		kind := repos.kinds()[i%len(repos.kinds())]
		if _, err := q.Submit(kind, codec.Fields{"id": fmt.Sprintf("job-%d", i)}); err != nil {
			return err
		}
	}

	done, err := q.Drain("worker-1", retry.Default(), repos.handle)
	if err != nil {
		return fmt.Errorf("draining: %w", err)
	}

	l.Info("drained", "jobs", fmt.Sprint(done), "remaining", fmt.Sprint(q.Len()))
	l.Info("records",
		"payment", fmt.Sprint(repos.payment.Count()),
		"inventory", fmt.Sprint(repos.inventory.Count()),
		"shipping", fmt.Sprint(repos.shipping.Count()))
	l.Info("metrics", "snapshot", strings.Join(reg.Snapshot(), " "))
	return nil
}

// repos is the set of write-side repositories the worker owns.
type repos struct {
	payment      *payment.Repo
	inventory    *inventory.Repo
	shipping     *shipping.Repo
	notification *notification.Repo
	audit        *audit.Repo
	billing      *billing.Repo
	session      *session.Repo
	webhook      *webhook.Repo
}

func newRepos(l *log.Logger, reg *metrics.Registry, st *store.Store) *repos {
	return &repos{
		payment:      payment.NewRepo(l, reg, st),
		inventory:    inventory.NewRepo(l, reg, st),
		shipping:     shipping.NewRepo(l, reg, st),
		notification: notification.NewRepo(l, reg, st),
		audit:        audit.NewRepo(l, reg, st),
		billing:      billing.NewRepo(l, reg, st),
		session:      session.NewRepo(l, reg, st),
		webhook:      webhook.NewRepo(l, reg, st),
	}
}

// kinds lists the job kinds this worker knows how to handle, in a stable
// order.
func (r *repos) kinds() []string {
	return []string{"payment", "inventory", "shipping", "notification",
		"audit", "billing", "session", "webhook"}
}
