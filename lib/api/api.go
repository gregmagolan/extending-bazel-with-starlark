// Package api is the request layer the server binary exposes.
//
// There is no net/http here on purpose: a Request is a struct and a Handler
// returns a Response, so the whole surface is testable without a socket. The
// pipeline per request is: rate limit, authenticate, authorize, validate,
// dispatch.
package api

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/catalog"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/feature"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/invoice"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/order"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/report"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/tenant"
	"github.com/gregmagolan/extending-bazel-with-starlark/internal/domain/user"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/auth"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/metrics"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/ratelimit"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

// Status mirrors the HTTP status codes this layer can produce.
type Status int

// The statuses a Handler may return.
const (
	StatusOK           Status = 200
	StatusBadRequest   Status = 400
	StatusUnauthorized Status = 401
	StatusForbidden    Status = 403
	StatusNotFound     Status = 404
	StatusConflict     Status = 409
	StatusThrottled    Status = 429
	StatusError        Status = 500
)

// Request is one inbound call.
type Request struct {
	Method string
	Path   string
	Token  string
	Caller string
	Body   codec.Fields
}

// Response is what a Handler produced.
type Response struct {
	Status Status
	Body   codec.Fields
	Detail string
}

// Handler serves one route for a verified identity.
type Handler func(Request, auth.Identity) Response

// route is one registered method+path pair.
type route struct {
	method  string
	path    string
	scope   string
	handler Handler
}

// Server dispatches requests to routes.
type Server struct {
	routes   []route
	verifier *auth.Verifier
	limiter  *ratelimit.Limiter
	store    *store.Store
	log      *log.Logger
	metrics  *metrics.Registry
	schemas  map[string]schema.Schema
}

// Deps is everything a Server needs from the process that hosts it.
type Deps struct {
	Log      *log.Logger
	Metrics  *metrics.Registry
	Config   *config.Config
	Store    *store.Store
	Verifier *auth.Verifier
	Limiter  *ratelimit.Limiter
}

// New builds a Server with the standard route table registered.
func New(d Deps) (*Server, error) {
	if d.Log == nil || d.Store == nil || d.Verifier == nil {
		return nil, errors.New("api: log, store and verifier are required")
	}
	if d.Metrics == nil {
		d.Metrics = metrics.NewRegistry(nil)
	}
	if d.Limiter == nil {
		burst, err := d.Config.Int("queue.depth", 256)
		if err != nil {
			return nil, err
		}
		d.Limiter = ratelimit.New(d.Log, d.Metrics, nil, burst, float64(burst))
	}

	s := &Server{
		verifier: d.Verifier,
		limiter:  d.Limiter,
		store:    d.Store,
		log:      d.Log,
		metrics:  d.Metrics,
		schemas:  map[string]schema.Schema{},
	}
	s.registerSchemas()
	s.registerRoutes()
	return s, nil
}

// registerSchemas pulls the schemas the read-only domain packages declare into
// one lookup table, and teaches the store the key prefixes they live under.
func (s *Server) registerSchemas() {
	groups := []struct {
		schemas  []schema.Schema
		prefixes []string
	}{
		{user.Schemas, user.Prefixes},
		{order.Schemas, order.Prefixes},
		{invoice.Schemas, invoice.Prefixes},
		{catalog.Schemas, catalog.Prefixes},
		{tenant.Schemas, tenant.Prefixes},
		{feature.Schemas, feature.Prefixes},
		{report.Schemas, report.Prefixes},
	}
	for _, g := range groups {
		for i, sch := range g.schemas {
			s.schemas[sch.Name] = sch
			s.store.Register(g.prefixes[i], sch)
		}
	}
}

// Schemas returns the names of every schema the server will accept, sorted.
func (s *Server) Schemas() []string {
	out := make([]string, 0, len(s.schemas))
	for name := range s.schemas {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Routes returns "METHOD path" for each registered route, sorted.
func (s *Server) Routes() []string {
	out := make([]string, 0, len(s.routes))
	for _, r := range s.routes {
		out = append(out, r.method+" "+r.path)
	}
	sort.Strings(out)
	return out
}

// Handle runs the full pipeline for req.
func (s *Server) Handle(req Request) Response {
	defer s.metrics.Time("api.request")()
	s.metrics.Inc("api.requests", 1)

	caller := req.Caller
	if caller == "" {
		caller = "anonymous"
	}
	if !s.limiter.Allow(caller) {
		s.metrics.Inc("api.throttled", 1)
		return Response{Status: StatusThrottled, Detail: "too many requests"}
	}

	id, err := s.verifier.Verify(req.Token)
	if err != nil {
		s.metrics.Inc("api.unauthorized", 1)
		return Response{Status: StatusUnauthorized, Detail: "unauthorized"}
	}

	r, ok := s.lookup(req.Method, req.Path)
	if !ok {
		s.metrics.Inc("api.not_found", 1)
		return Response{Status: StatusNotFound, Detail: "no such route"}
	}
	if r.scope != "" && !id.HasScope(r.scope) {
		s.metrics.Inc("api.forbidden", 1)
		return Response{Status: StatusForbidden, Detail: "scope " + r.scope + " required"}
	}

	resp := r.handler(req, id)
	s.metrics.Inc(fmt.Sprintf("api.status.%d", int(resp.Status)), 1)
	s.log.Debug("request served", "method", req.Method, "path", req.Path,
		"status", fmt.Sprint(int(resp.Status)), "sub", id.Subject)
	return resp
}

// lookup finds the route for a method and path. A route path ending in "/"
// matches any longer path, which is how the record routes take an id.
func (s *Server) lookup(method, path string) (route, bool) {
	for _, r := range s.routes {
		if r.method != method {
			continue
		}
		if r.path == path {
			return r, true
		}
		if strings.HasSuffix(r.path, "/") && strings.HasPrefix(path, r.path) && len(path) > len(r.path) {
			return r, true
		}
	}
	return route{}, false
}

func (s *Server) registerRoutes() {
	s.routes = []route{
		{"GET", "/healthz", "", s.handleHealth},
		{"GET", "/schemas", "", s.handleSchemas},
		{"GET", "/metrics", "admin", s.handleMetrics},
		{"GET", "/records/", "read", s.handleGet},
		{"PUT", "/records/", "write", s.handlePut},
		{"DELETE", "/records/", "write", s.handleDelete},
		{"GET", "/summary", "read", s.handleSummary},
	}
}

func (s *Server) handleHealth(Request, auth.Identity) Response {
	return Response{Status: StatusOK, Body: codec.Fields{
		"status":  "ok",
		"records": fmt.Sprint(s.store.Len()),
	}}
}

func (s *Server) handleSchemas(Request, auth.Identity) Response {
	return Response{Status: StatusOK, Body: codec.Fields{
		"schemas": strings.Join(s.Schemas(), ","),
	}}
}

func (s *Server) handleMetrics(Request, auth.Identity) Response {
	return Response{Status: StatusOK, Body: codec.Fields{
		"snapshot": strings.Join(s.metrics.Snapshot(), " "),
	}}
}

func (s *Server) handleGet(req Request, _ auth.Identity) Response {
	key := strings.TrimPrefix(req.Path, "/records/")
	rec, err := s.store.Get(key)
	if errors.Is(err, store.ErrNotFound) {
		return Response{Status: StatusNotFound, Detail: err.Error()}
	}
	if err != nil {
		return Response{Status: StatusError, Detail: err.Error()}
	}
	body := rec.Fields.Clone()
	body["_revision"] = fmt.Sprint(rec.Revision)
	return Response{Status: StatusOK, Body: body}
}

func (s *Server) handlePut(req Request, id auth.Identity) Response {
	key := strings.TrimPrefix(req.Path, "/records/")
	fields := req.Body.Clone()
	if fields == nil {
		fields = codec.Fields{}
	}

	expect, err := fields.Int("_revision", 0)
	if err != nil {
		return Response{Status: StatusBadRequest, Detail: err.Error()}
	}
	delete(fields, "_revision")
	if id.Tenant != "" {
		fields["tenant"] = id.Tenant
	}

	rev, err := s.store.Put(key, int64(expect), fields)
	switch {
	case errors.Is(err, store.ErrConflict):
		return Response{Status: StatusConflict, Detail: err.Error()}
	case errors.Is(err, store.ErrFull):
		return Response{Status: StatusError, Detail: err.Error()}
	case err != nil:
		return Response{Status: StatusBadRequest, Detail: err.Error()}
	}
	return Response{Status: StatusOK, Body: codec.Fields{"_revision": fmt.Sprint(rev)}}
}

func (s *Server) handleDelete(req Request, _ auth.Identity) Response {
	key := strings.TrimPrefix(req.Path, "/records/")
	if err := s.store.Delete(key); err != nil {
		return Response{Status: StatusNotFound, Detail: err.Error()}
	}
	return Response{Status: StatusOK}
}

// handleSummary reports what each read-only domain package manages. It is the
// only place that pulls every domain Summary together.
func (s *Server) handleSummary(Request, auth.Identity) Response {
	summaries := []string{
		user.Summary(), order.Summary(), invoice.Summary(), catalog.Summary(),
		tenant.Summary(), feature.Summary(), report.Summary(),
	}
	sort.Strings(summaries)
	return Response{Status: StatusOK, Body: codec.Fields{
		"domains": fmt.Sprint(len(summaries)),
		"detail":  strings.Join(summaries, " | "),
	}}
}

// Seed writes one sample record per read-only domain package. The cron binary
// uses it to give a fresh deployment something to serve.
func (s *Server) Seed() (int, error) {
	samples := []struct {
		key    string
		fields codec.Fields
	}{
		{user.SampleUserRecord("seed", zeroTime()).Key(), user.SampleUserRecord("seed", zeroTime()).Fields()},
		{order.SampleOrderRecord("seed", zeroTime()).Key(), order.SampleOrderRecord("seed", zeroTime()).Fields()},
		{invoice.SampleInvoiceRecord("seed", zeroTime()).Key(), invoice.SampleInvoiceRecord("seed", zeroTime()).Fields()},
		{catalog.SampleCatalogRecord("seed", zeroTime()).Key(), catalog.SampleCatalogRecord("seed", zeroTime()).Fields()},
		{tenant.SampleTenantRecord("seed", zeroTime()).Key(), tenant.SampleTenantRecord("seed", zeroTime()).Fields()},
		{feature.SampleFeatureRecord("seed", zeroTime()).Key(), feature.SampleFeatureRecord("seed", zeroTime()).Fields()},
		{report.SampleReportRecord("seed", zeroTime()).Key(), report.SampleReportRecord("seed", zeroTime()).Fields()},
	}
	n := 0
	for _, sample := range samples {
		if _, err := s.store.Put(sample.key, 0, sample.fields); err != nil {
			return n, fmt.Errorf("api: seed %s: %w", sample.key, err)
		}
		n++
	}
	return n, nil
}
