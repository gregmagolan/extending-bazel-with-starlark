package main

import (
	"testing"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/queue"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/retry"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

func newTestRepos(t *testing.T) *repos {
	t.Helper()
	cfg, err := config.Parse("")
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(log.Discard(), cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return newRepos(log.Discard(), nil, st)
}

func TestRunDrainsEveryJob(t *testing.T) {
	if err := run("log.level=error", 8); err != nil {
		t.Fatalf("run() = %v", err)
	}
}

func TestRunRejectsBadConfig(t *testing.T) {
	if err := run("queue.depth=0", 1); err == nil {
		t.Error("run() accepted a zero queue depth")
	}
}

func TestHandleCoversEveryKind(t *testing.T) {
	r := newTestRepos(t)
	for i, kind := range r.kinds() {
		job := queue.Job{ID: kind + "-1", Kind: kind, Payload: codec.Fields{"id": "x"}}
		if err := r.handle(job); err != nil {
			t.Errorf("handle(%s) = %v", kind, err)
		}
		if i == 0 && r.payment.Count() != 1 {
			t.Errorf("payment.Count() = %d, want 1", r.payment.Count())
		}
	}
}

func TestHandleRejectsUnknownKindPermanently(t *testing.T) {
	r := newTestRepos(t)
	err := r.handle(queue.Job{ID: "x", Kind: "nope", Payload: codec.Fields{"id": "x"}})
	if !retry.IsPermanent(err) {
		t.Errorf("handle(unknown kind) = %v, want a permanent error", err)
	}
	err = r.handle(queue.Job{ID: "x", Kind: "payment"})
	if !retry.IsPermanent(err) {
		t.Errorf("handle(no id) = %v, want a permanent error", err)
	}
}
