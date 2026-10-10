package main

import (
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// storeHandle is an alias kept so that main.go reads the same whether the
// store is opened directly or wrapped by a future caching layer.
type storeHandle = store.Store

func newStore(l *log.Logger, cfg *config.Config, reg *metrics.Registry) (*storeHandle, error) {
	return store.Open(l, cfg, reg)
}
