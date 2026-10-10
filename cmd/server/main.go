// Command server hosts the api layer and accepts work for the queue.
//
// It does not listen on a socket: the course cares about the build graph, not
// about the network, so main wires the dependencies, serves a short scripted
// batch of requests and exits. The wiring is the point -- it is what makes
// //cmd/server depend on most of lib/.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/api"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/auth"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/queue"
)

func main() {
	settings := flag.String("config", "", "inline key=value configuration")
	secret := flag.String("secret", "local-dev-secret", "token signing secret")
	requests := flag.Int("requests", 3, "number of scripted requests to serve")
	flag.Parse()

	if err := run(*settings, *secret, *requests); err != nil {
		fmt.Fprintln(os.Stderr, "server:", err)
		os.Exit(1)
	}
}

func run(settings, secret string, requests int) error {
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
	l := log.New(os.Stderr, level).With("service", "server")

	reg := metrics.NewRegistry(nil)
	st, err := openStore(l, cfg, reg)
	if err != nil {
		return err
	}

	verifier, err := auth.NewVerifier(l, cfg, secret)
	if err != nil {
		return err
	}

	srv, err := api.New(api.Deps{
		Log:      l,
		Metrics:  reg,
		Config:   cfg,
		Store:    st,
		Verifier: verifier,
	})
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

	port, err := cfg.Int("http.port", 8080)
	if err != nil {
		return err
	}
	l.Info("server ready",
		"port", fmt.Sprint(port),
		"routes", fmt.Sprint(len(srv.Routes())),
		"schemas", fmt.Sprint(len(srv.Schemas())))

	token := verifier.Mint(auth.Identity{Subject: "scripted", Tenant: "local", Scopes: []string{"admin"}})
	for i := 0; i < requests; i++ {
		key := fmt.Sprintf("notes/%d", i)
		resp := srv.Handle(api.Request{
			Method: "PUT",
			Path:   "/records/" + key,
			Token:  token,
			Caller: "scripted",
			Body:   codec.Fields{"body": fmt.Sprintf("scripted request %d", i)},
		})
		if resp.Status != api.StatusOK {
			return fmt.Errorf("PUT %s: %d %s", key, int(resp.Status), resp.Detail)
		}
		if _, err := q.Submit("index", codec.Fields{"key": key}); err != nil {
			return err
		}
	}

	health := srv.Handle(api.Request{Method: "GET", Path: "/healthz", Token: token, Caller: "scripted"})
	l.Info("health", "status", fmt.Sprint(int(health.Status)), "records", health.Body["records"])
	l.Info("queued", "depth", fmt.Sprint(q.Len()))
	l.Info("metrics", "snapshot", strings.Join(reg.Snapshot(), " "))
	return nil
}

// openStore is split out so the error message names the configured path,
// which is the thing that is usually wrong.
func openStore(l *log.Logger, cfg *config.Config, reg *metrics.Registry) (*storeHandle, error) {
	st, err := newStore(l, cfg, reg)
	if err != nil {
		return nil, fmt.Errorf("opening store at %s: %w", cfg.String("store.path", "?"), err)
	}
	return st, nil
}
