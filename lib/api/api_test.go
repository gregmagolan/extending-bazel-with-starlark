package api

import (
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/auth"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/ratelimit"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

const secret = "test-secret"

func newServer(t *testing.T, cfgText string) *Server {
	t.Helper()
	cfg, err := config.Parse(cfgText)
	if err != nil {
		t.Fatal(err)
	}
	l := log.Discard()
	reg := metrics.NewRegistry(nil)
	st, err := store.Open(l, cfg, reg)
	if err != nil {
		t.Fatal(err)
	}
	v, err := auth.NewVerifier(l, cfg, secret)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(Deps{Log: l, Metrics: reg, Config: cfg, Store: st, Verifier: v})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func token(t *testing.T, scopes ...string) string {
	t.Helper()
	cfg, _ := config.Parse("")
	v, err := auth.NewVerifier(log.Discard(), cfg, secret)
	if err != nil {
		t.Fatal(err)
	}
	return v.Mint(auth.Identity{Subject: "u-1", Tenant: "acme", Scopes: scopes})
}

func TestNewRequiresDeps(t *testing.T) {
	if _, err := New(Deps{}); err == nil {
		t.Error("New accepted empty deps")
	}
}

func TestRoutesAndSchemasAreRegistered(t *testing.T) {
	s := newServer(t, "")
	if got := len(s.Routes()); got != 7 {
		t.Errorf("Routes() = %v", s.Routes())
	}
	if got := len(s.Schemas()); got < 70 {
		t.Errorf("Schemas() registered only %d schemas", got)
	}
}

func TestHealthNeedsNoScope(t *testing.T) {
	s := newServer(t, "")
	resp := s.Handle(Request{Method: "GET", Path: "/healthz", Token: token(t)})
	if resp.Status != StatusOK || resp.Body["status"] != "ok" {
		t.Fatalf("Handle(/healthz) = %+v", resp)
	}
}

func TestUnknownRouteIsNotFound(t *testing.T) {
	s := newServer(t, "")
	if resp := s.Handle(Request{Method: "GET", Path: "/nope", Token: token(t)}); resp.Status != StatusNotFound {
		t.Fatalf("Handle(/nope) = %+v", resp)
	}
}

func TestBadTokenIsUnauthorized(t *testing.T) {
	s := newServer(t, "")
	if resp := s.Handle(Request{Method: "GET", Path: "/healthz", Token: "bogus"}); resp.Status != StatusUnauthorized {
		t.Fatalf("Handle(bad token) = %+v", resp)
	}
}

func TestScopeIsEnforced(t *testing.T) {
	s := newServer(t, "")
	resp := s.Handle(Request{Method: "GET", Path: "/metrics", Token: token(t, "read")})
	if resp.Status != StatusForbidden {
		t.Fatalf("Handle(/metrics) without admin = %+v", resp)
	}
	resp = s.Handle(Request{Method: "GET", Path: "/metrics", Token: token(t, "admin")})
	if resp.Status != StatusOK || resp.Body["snapshot"] == "" {
		t.Fatalf("Handle(/metrics) with admin = %+v", resp)
	}
}

func TestRecordLifecycle(t *testing.T) {
	s := newServer(t, "")
	tok := token(t, "read", "write")

	put := s.Handle(Request{Method: "PUT", Path: "/records/notes/1", Token: tok, Body: codec.Fields{"body": "hello"}})
	if put.Status != StatusOK || put.Body["_revision"] != "1" {
		t.Fatalf("PUT = %+v", put)
	}

	get := s.Handle(Request{Method: "GET", Path: "/records/notes/1", Token: tok})
	if get.Status != StatusOK || get.Body["body"] != "hello" || get.Body["tenant"] != "acme" {
		t.Fatalf("GET = %+v", get)
	}

	stale := s.Handle(Request{Method: "PUT", Path: "/records/notes/1", Token: tok, Body: codec.Fields{"_revision": "7"}})
	if stale.Status != StatusConflict {
		t.Fatalf("stale PUT = %+v", stale)
	}

	if del := s.Handle(Request{Method: "DELETE", Path: "/records/notes/1", Token: tok}); del.Status != StatusOK {
		t.Fatalf("DELETE = %+v", del)
	}
	if get := s.Handle(Request{Method: "GET", Path: "/records/notes/1", Token: tok}); get.Status != StatusNotFound {
		t.Fatalf("GET after DELETE = %+v", get)
	}
}

func TestPutRejectsBadRevision(t *testing.T) {
	s := newServer(t, "")
	resp := s.Handle(Request{Method: "PUT", Path: "/records/notes/1", Token: token(t, "write"),
		Body: codec.Fields{"_revision": "soon"}})
	if resp.Status != StatusBadRequest {
		t.Fatalf("PUT with a bad revision = %+v", resp)
	}
}

func TestThrottling(t *testing.T) {
	s := newServer(t, "queue.depth=2")
	s.limiter = ratelimit.New(log.Discard(), nil, nil, 2, 0.0001)
	tok := token(t, "read")
	for i := 0; i < 2; i++ {
		if resp := s.Handle(Request{Method: "GET", Path: "/healthz", Token: tok, Caller: "c"}); resp.Status != StatusOK {
			t.Fatalf("call %d = %+v", i, resp)
		}
	}
	if resp := s.Handle(Request{Method: "GET", Path: "/healthz", Token: tok, Caller: "c"}); resp.Status != StatusThrottled {
		t.Fatalf("third call = %+v", resp)
	}
}

func TestSummaryCoversEveryReadOnlyDomain(t *testing.T) {
	s := newServer(t, "")
	resp := s.Handle(Request{Method: "GET", Path: "/summary", Token: token(t, "read")})
	if resp.Status != StatusOK || resp.Body["domains"] != "7" {
		t.Fatalf("GET /summary = %+v", resp)
	}
}

func TestSeed(t *testing.T) {
	s := newServer(t, "")
	n, err := s.Seed()
	if err != nil || n != 7 {
		t.Fatalf("Seed() = %d, %v", n, err)
	}
	if s.store.Len() != 7 {
		t.Errorf("store holds %d records after Seed()", s.store.Len())
	}
	if got := zeroTime().Year(); got != 2000 {
		t.Errorf("zeroTime() = %v", got)
	}
	_ = time.Now
}
