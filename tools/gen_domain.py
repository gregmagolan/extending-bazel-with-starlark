#!/usr/bin/env python3
"""Regenerate the internal/domain/* packages.

    tools/gen_domain.py [entities-per-package]   # default 22

This script is repo maintenance, not part of the build: nothing in Bazel runs
it, and its output is checked in. It exists so that the ~230k lines under
internal/domain are reproducible rather than mysterious.

The generated code is deliberately boring and repetitive. Its job is to give
the build a realistic amount of work (about 25s cold on an M-series Mac) and
the dependency graph a realistic width, not to be read closely. Field names
per entity come from a per-package seeded RNG, so regenerating with the same
entity count produces byte-identical output.

After running this, run `bazel run //:gazelle` to refresh the BUILD files.
"""
import os
import random
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
MOD = "github.com/gregmagolan/extending-bazel-with-starlark"

# Each entry is (package, wired, deps).
#
# "wired" packages talk to lib/log, lib/metrics and lib/store and get a Repo;
# "pure" packages are model-only and never import lib/log, which is what makes
# the reverse-dependency query on //lib/log interesting.
#
# "deps" are other domain packages this one links to. They give the graph depth
# and mean every package has a consumer: lib/api imports the top of the pure
# chain and cmd/worker imports the top of the wired chain, and everything else
# is reached transitively.
PACKAGES = [
    # pure, leaves
    ("pricing", False, []),
    ("geo", False, []),
    ("locale", False, []),
    ("forecast", False, []),
    ("taxonomy", False, []),
    ("rollout", False, []),
    ("segment", False, []),
    ("quota", False, []),
    ("profile", False, []),
    ("preference", False, []),
    # pure, linked
    ("ledger", False, ["pricing"]),
    ("user", False, ["profile", "preference"]),
    ("catalog", False, ["pricing", "taxonomy"]),
    ("tenant", False, ["locale", "quota"]),
    ("feature", False, ["rollout", "segment"]),
    ("order", False, ["user", "catalog"]),
    ("invoice", False, ["order", "pricing", "ledger"]),
    ("report", False, ["ledger", "geo", "locale", "forecast"]),
    # wired, leaves
    ("settlement", True, []),
    ("refund", True, []),
    ("warehouse", True, []),
    ("stocktake", True, []),
    ("carrier", True, []),
    ("route", True, []),
    ("template", True, []),
    ("digest", True, []),
    ("trail", True, []),
    ("dunning", True, []),
    ("usage", True, []),
    ("device", True, []),
    ("delivery", True, []),
    # wired, linked
    ("payment", True, ["settlement", "refund"]),
    ("inventory", True, ["warehouse", "stocktake"]),
    ("shipping", True, ["carrier", "route"]),
    ("notification", True, ["template", "digest"]),
    ("audit", True, ["trail"]),
    ("billing", True, ["dunning", "usage"]),
    ("session", True, ["device"]),
    ("webhook", True, ["delivery"]),
]

NOUNS = ["Record", "Entry", "Item", "Draft", "Revision", "Snapshot", "Batch",
         "Request", "Summary", "Policy", "Link", "Note", "Window", "Quota",
         "Bucket", "Lane", "Slot", "Tag", "Rule", "Hold", "Claim", "Cursor",
         "Chunk", "Segment", "Marker", "Ticket", "Plan", "Target", "Stage",
         "Channel", "Outcome", "Attempt", "Label", "Scope", "Bundle", "Group"]

STRFIELDS = ["Name", "Label", "Owner", "Region", "Channel", "Reference",
             "Category", "Source", "Comment", "Group"]
INTFIELDS = ["Count", "Weight", "Priority", "Attempts", "Version", "Offset",
             "Capacity", "Retained"]


class Spec:
    def __init__(self, pkg, index, rnd):
        noun = NOUNS[index % len(NOUNS)]
        suffix = "" if index < len(NOUNS) else str(index // len(NOUNS))
        self.name = pkg.capitalize() + noun + suffix
        self.noun = noun.lower()
        self.lower = self.name.lower()
        self.prefix = "%s/%s/" % (pkg, self.lower)
        self.s = rnd.sample(STRFIELDS, 3)
        self.i = rnd.sample(INTFIELDS, 2)

    @property
    def sk(self):
        return [f.lower() for f in self.s]

    @property
    def ik(self):
        return [f.lower() for f in self.i]


def model_file(pkg, specs):
    parts = ['''// Package %s holds the %s domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package %s

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"%s/lib/codec"
	"%s/lib/schema"
)
''' % (pkg, pkg, pkg, MOD, MOD)]

    for sp in specs:
        n, s, sk, i, ik = sp.name, sp.s, sp.sk, sp.i, sp.ik
        parts.append('''
// %(n)sSchema is the stored shape of a %(n)s.
var %(n)sSchema = schema.New("%(pkg)s.%(low)s",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "%(sk0)s", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "%(sk1)s", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "%(sk2)s", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "%(ik0)s", Kind: schema.KindInt},
	schema.Field{Name: "%(ik1)s", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// %(n)s is one %(pkg)s %(noun)s.
type %(n)s struct {
	ID       string
	%(s0)s string
	%(s1)s string
	%(s2)s string
	%(i0)s int
	%(i1)s int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x %(n)s) Key() string { return "%(pkg)s/%(low)s/" + x.ID }

// Fields encodes x for storage.
func (x %(n)s) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"%(sk0)s":  x.%(s0)s,
		"%(sk1)s":  x.%(s1)s,
		"%(sk2)s":  x.%(s2)s,
		"%(ik0)s":  strconv.Itoa(x.%(i0)s),
		"%(ik1)s":  strconv.Itoa(x.%(i1)s),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// Parse%(n)s decodes a stored record.
func Parse%(n)s(f codec.Fields) (%(n)s, error) {
	var x %(n)s
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("%(pkg)s: %(low)s: missing id")
	}
	x.%(s0)s = f["%(sk0)s"]
	x.%(s1)s = f["%(sk1)s"]
	x.%(s2)s = f["%(sk2)s"]
	var err error
	if x.%(i0)s, err = f.Int("%(ik0)s", 0); err != nil {
		return x, fmt.Errorf("%(pkg)s: %(low)s %%s: %%w", x.ID, err)
	}
	if x.%(i1)s, err = f.Int("%(ik1)s", 0); err != nil {
		return x, fmt.Errorf("%(pkg)s: %(low)s %%s: %%w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("%(pkg)s: %(low)s %%s: %%w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("%(pkg)s: %(low)s %%s: updated: %%w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x %(n)s) String() string {
	return fmt.Sprintf("%(n)s(%%s, %(sk0)s=%%q, %(ik0)s=%%d, archived=%%t)", x.ID, x.%(s0)s, x.%(i0)s, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x %(n)s) Active() bool { return !x.Archived && x.%(i0)s >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x %(n)s) Normalize() %(n)s {
	x.%(s0)s = strings.TrimSpace(x.%(s0)s)
	x.%(s1)s = strings.TrimSpace(x.%(s1)s)
	x.%(s2)s = strings.TrimSpace(x.%(s2)s)
	if x.%(i0)s < 0 {
		x.%(i0)s = 0
	}
	if x.%(i1)s < 0 {
		x.%(i1)s = 0
	}
	return x
}

// Sort%(n)ss orders by %(ik0)s descending, then by ID.
func Sort%(n)ss(xs []%(n)s) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].%(i0)s != xs[j].%(i0)s {
			return xs[i].%(i0)s > xs[j].%(i0)s
		}
		return xs[i].ID < xs[j].ID
	})
}

// Active%(n)ss drops the archived entries, preserving order.
func Active%(n)ss(xs []%(n)s) []%(n)s {
	out := make([]%(n)s, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// Sum%(n)ss adds up the %(ik0)s of every active entry.
func Sum%(n)ss(xs []%(n)s) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.%(i0)s
		}
	}
	return total
}

// Index%(n)ss keys the entries by ID, last write winning.
func Index%(n)ss(xs []%(n)s) map[string]%(n)s {
	out := make(map[string]%(n)s, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// Sample%(n)s builds a valid %(n)s, for tests and for the seed data the cron
// binary writes.
func Sample%(n)s(id string, at time.Time) %(n)s {
	return %(n)s{
		ID:       id,
		%(s0)s: "%(sk0)s-" + id,
		%(s1)s: "%(sk1)s-" + id,
		%(s2)s: "seeded by %(pkg)s.Sample%(n)s",
		%(i0)s: len(id),
		%(i1)s: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}
''' % dict(n=n, pkg=pkg, low=sp.lower, noun=sp.noun,
           s0=s[0], s1=s[1], s2=s[2], i0=i[0], i1=i[1],
           sk0=sk[0], sk1=sk[1], sk2=sk[2], ik0=ik[0], ik1=ik[1]))

    parts.append('''
// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
%s
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
%s
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "%s: " + strings.Join(names, ", ")
}
''' % ("\n".join("\t%sSchema," % sp.name for sp in specs),
       "\n".join('\t"%s",' % sp.prefix for sp in specs),
       pkg))
    return "".join(parts)


def validate_file(pkg, specs):
    parts = ['''package %s

import (
	"fmt"

	"%s/lib/codec"
	"%s/lib/validate"
)

// Validator checks encoded %s records against this package's schemas.
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
		return fmt.Errorf("%s: unknown schema %%q", name)
	}
	return nil
}
''' % (pkg, MOD, MOD, pkg, pkg)]

    for sp in specs:
        parts.append('''
// Validate checks x against %(n)sSchema.
func (x %(n)s) Validate() error {
	if err := validate.NonEmpty("id", x.ID); err != nil {
		return fmt.Errorf("%(pkg)s: %(low)s: %%w", err)
	}
	if err := validate.Record(%(n)sSchema, x.Fields()).Err(); err != nil {
		return fmt.Errorf("%(pkg)s: %(low)s %%s: %%w", x.ID, err)
	}
	return nil
}

// Validate%(n)ss returns the first problem found in xs.
func Validate%(n)ss(xs []%(n)s) error {
	for i, x := range xs {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("index %%d: %%w", i, err)
		}
	}
	return nil
}
''' % dict(n=sp.name, pkg=pkg, low=sp.lower))
    return "".join(parts)


def repo_file(pkg, specs):
    parts = ['''package %s

import (
	"fmt"

	"%s/lib/log"
	"%s/lib/metrics"
	"%s/lib/store"
)

// Repo persists %s entities into a store.Store.
//
// Every method is the same three steps -- validate, encode, write -- repeated
// per entity type. The repetition is deliberate: it is what gives the
// reverse-dependency queries in the course a non-trivial answer.
type Repo struct {
	store   *store.Store
	log     *log.Logger
	metrics *metrics.Registry
}

// NewRepo wires a Repo and registers every schema this package declares.
func NewRepo(l *log.Logger, reg *metrics.Registry, s *store.Store) *Repo {
	if reg == nil {
		reg = metrics.NewRegistry(nil)
	}
	for i, sch := range Schemas {
		s.Register(Prefixes[i], sch)
	}
	l.Debug("%s repo ready", "schemas", fmt.Sprint(len(Schemas)))
	return &Repo{store: s, log: l, metrics: reg}
}

// Count reports how many %s records are currently held.
func (r *Repo) Count() int {
	n := 0
	for _, prefix := range Prefixes {
		n += len(r.store.List(prefix))
	}
	return n
}
''' % (pkg, MOD, MOD, MOD, pkg, pkg, pkg)]

    for sp in specs:
        parts.append('''
// Put%(n)s writes x, creating it when revision is zero.
func (r *Repo) Put%(n)s(x %(n)s, revision int64) (int64, error) {
	x = x.Normalize()
	if err := x.Validate(); err != nil {
		r.metrics.Inc("%(pkg)s.%(low)s.invalid", 1)
		return 0, err
	}
	rev, err := r.store.Put(x.Key(), revision, x.Fields())
	if err != nil {
		r.log.Warn("%(pkg)s write failed", "key", x.Key(), "err", err.Error())
		return 0, err
	}
	r.metrics.Inc("%(pkg)s.%(low)s.written", 1)
	return rev, nil
}

// Get%(n)s reads the entity with the given id.
func (r *Repo) Get%(n)s(id string) (%(n)s, int64, error) {
	rec, err := r.store.Get("%(prefix)s" + id)
	if err != nil {
		return %(n)s{}, 0, err
	}
	x, err := Parse%(n)s(rec.Fields)
	if err != nil {
		return %(n)s{}, 0, err
	}
	return x, rec.Revision, nil
}

// List%(n)ss reads every stored entity of this type, sorted.
func (r *Repo) List%(n)ss() ([]%(n)s, error) {
	recs := r.store.List("%(prefix)s")
	out := make([]%(n)s, 0, len(recs))
	for _, rec := range recs {
		x, err := Parse%(n)s(rec.Fields)
		if err != nil {
			return nil, fmt.Errorf("%(pkg)s: %%s: %%w", rec.Key, err)
		}
		out = append(out, x)
	}
	Sort%(n)ss(out)
	return out, nil
}
''' % dict(n=sp.name, pkg=pkg, low=sp.lower, prefix=sp.prefix))
    return "".join(parts)


def links_file(pkg, deps):
    """A package that links to other domain packages exposes their summaries
    and schema counts through one function, so the import is load-bearing."""
    imports = "\n".join('\t"%s/internal/domain/%s"' % (MOD, d) for d in deps)
    calls = "\n".join('\t\t"%s: " + %s.Summary(),' % (d, d) for d in deps)
    counts = "\n".join('\tn += len(%s.Schemas)' % d for d in deps)
    return '''package %(pkg)s

import (
	"sort"
	"strings"

%(imports)s
)

// Linked returns the summary of every domain package %(pkg)s builds on, plus
// its own, sorted. The api and worker layers print it at startup so that a
// deployment logs the shape of the model it is serving.
func Linked() []string {
	out := []string{
		"%(pkg)s: " + Summary(),
%(calls)s
	}
	sort.Strings(out)
	return out
}

// LinkedSchemas counts the schemas reachable from this package, including its
// own.
func LinkedSchemas() int {
	n := len(Schemas)
%(counts)s
	return n
}

// LinkedNames lists the domain packages %(pkg)s links to.
func LinkedNames() []string {
	return strings.Split("%(names)s", ",")
}
''' % dict(pkg=pkg, imports=imports, calls=calls, counts=counts,
           names=",".join(deps))


def test_file(pkg, wired, specs, deps):
    imports = ['\t"%s/lib/codec"' % MOD]
    if wired:
        imports += ['\t"%s/lib/config"' % MOD, '\t"%s/lib/log"' % MOD,
                    '\t"%s/lib/store"' % MOD]
    parts = ['''package %s

import (
	"strings"
	"testing"
	"time"

%s
)

var when = time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)

func TestSchemasAndPrefixesLineUp(t *testing.T) {
	if len(Schemas) != len(Prefixes) {
		t.Fatalf("%%d schemas but %%d prefixes", len(Schemas), len(Prefixes))
	}
	for i, s := range Schemas {
		if s.Fingerprint() == "" {
			t.Errorf("schema %%d has an empty fingerprint", i)
		}
		if !strings.HasPrefix(Prefixes[i], "%s/") {
			t.Errorf("prefix %%d = %%q", i, Prefixes[i])
		}
	}
}

func TestSummaryNamesEverySchema(t *testing.T) {
	got := Summary()
	for _, s := range Schemas {
		if !strings.Contains(got, s.Name) {
			t.Errorf("Summary() is missing %%q", s.Name)
		}
	}
}

func TestValidatorStrictness(t *testing.T) {
	if err := NewValidator(false).Check("nope", codec.Fields{}); err != nil {
		t.Errorf("lenient Check() = %%v", err)
	}
	if err := NewValidator(true).Check("nope", codec.Fields{}); err == nil {
		t.Error("strict Check() accepted an unknown schema")
	}
	if err := NewValidator(true).Check(Schemas[0].Name, Sample%s("ok", when).Fields()); err != nil {
		t.Errorf("Check(%%q) = %%v", Schemas[0].Name, err)
	}
}
''' % (pkg, "\n".join(imports), pkg, specs[0].name)]

    if deps:
        parts.append('''
func TestLinkedCoversEveryDependency(t *testing.T) {
	got := Linked()
	if len(got) != len(LinkedNames())+1 {
		t.Fatalf("Linked() = %v", got)
	}
	for _, name := range LinkedNames() {
		found := false
		for _, line := range got {
			if strings.HasPrefix(line, name+":") {
				found = true
			}
		}
		if !found {
			t.Errorf("Linked() is missing %q", name)
		}
	}
	if LinkedSchemas() <= len(Schemas) {
		t.Errorf("LinkedSchemas() = %d, want more than the %d local schemas", LinkedSchemas(), len(Schemas))
	}
}
''')

    if wired:
        parts.append('''
func newRepo(t *testing.T) *Repo {
	t.Helper()
	cfg, err := config.Parse("")
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(log.Discard(), cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return NewRepo(log.Discard(), nil, s)
}

func TestRepoStartsEmpty(t *testing.T) {
	if got := newRepo(t).Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
}
''')

    for sp in specs:
        parts.append('''
func Test%(n)sRoundTrip(t *testing.T) {
	x := Sample%(n)s("%(low)s-1", when)
	got, err := Parse%(n)s(x.Fields())
	if err != nil {
		t.Fatalf("Parse%(n)s() = %%v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.%(s0)s != x.%(s0)s {
		t.Fatalf("round trip = %%+v, want %%+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %%v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %%q", got.String())
	}
}

func Test%(n)sParseRejectsBadInput(t *testing.T) {
	f := Sample%(n)s("x", when).Fields()
	delete(f, "id")
	if _, err := Parse%(n)s(f); err == nil {
		t.Error("Parse%(n)s() accepted a record with no id")
	}
	f = Sample%(n)s("x", when).Fields()
	f["%(ik0)s"] = "not-a-number"
	if _, err := Parse%(n)s(f); err == nil {
		t.Error("Parse%(n)s() accepted a non-numeric %(ik0)s")
	}
}

func Test%(n)sCollections(t *testing.T) {
	xs := []%(n)s{Sample%(n)s("a", when), Sample%(n)s("bb", when), Sample%(n)s("ccc", when)}
	xs[1].Archived = true
	if got := len(Active%(n)ss(xs)); got != 2 {
		t.Errorf("Active%(n)ss() kept %%d, want 2", got)
	}
	if got, want := Sum%(n)ss(xs), 1+3; got != want {
		t.Errorf("Sum%(n)ss() = %%d, want %%d", got, want)
	}
	Sort%(n)ss(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("Sort%(n)ss() put %%q first", xs[0].ID)
	}
	if len(Index%(n)ss(xs)) != 3 {
		t.Errorf("Index%(n)ss() = %%v", Index%(n)ss(xs))
	}
	if err := Validate%(n)ss(Active%(n)ss(xs)); err != nil {
		t.Errorf("Validate%(n)ss() = %%v", err)
	}
}

func Test%(n)sNormalizeClamps(t *testing.T) {
	x := Sample%(n)s("a", when)
	x.%(s0)s = "  padded  "
	x.%(i0)s = -5
	got := x.Normalize()
	if got.%(s0)s != "padded" || got.%(i0)s != 0 {
		t.Errorf("Normalize() = %%+v", got)
	}
}
''' % dict(n=sp.name, low=sp.lower, s0=sp.s[0], i0=sp.i[0], ik0=sp.ik[0]))
        if wired:
            parts.append('''
func Test%(n)sRepo(t *testing.T) {
	r := newRepo(t)
	x := Sample%(n)s("r1", when)
	rev, err := r.Put%(n)s(x, 0)
	if err != nil {
		t.Fatalf("Put%(n)s() = %%v", err)
	}
	got, gotRev, err := r.Get%(n)s(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("Get%(n)s() = %%+v, %%d, %%v", got, gotRev, err)
	}
	all, err := r.List%(n)ss()
	if err != nil || len(all) != 1 {
		t.Fatalf("List%(n)ss() = %%v, %%v", all, err)
	}
	if _, err := r.Put%(n)s(%(n)s{}, 0); err == nil {
		t.Error("Put%(n)s() accepted an empty entity")
	}
}
''' % dict(n=sp.name))
    return "".join(parts)


def main():
    entities = int(sys.argv[1]) if len(sys.argv) > 1 else 22
    for pkg, wired, deps in PACKAGES:
        d = os.path.join(ROOT, "internal", "domain", pkg)
        # Drop the previous .go files but keep anything else in the directory,
        # notably the BUILD.bazel that Gazelle generated.
        os.makedirs(d, exist_ok=True)
        for name in os.listdir(d):
            if name.endswith(".go"):
                os.remove(os.path.join(d, name))
        rnd = random.Random("%s-seed" % pkg)
        specs = [Spec(pkg, i, rnd) for i in range(entities)]
        with open(os.path.join(d, "model.go"), "w") as f:
            f.write(model_file(pkg, specs))
        with open(os.path.join(d, "validate.go"), "w") as f:
            f.write(validate_file(pkg, specs))
        if wired:
            with open(os.path.join(d, "repo.go"), "w") as f:
                f.write(repo_file(pkg, specs))
        if deps:
            with open(os.path.join(d, "links.go"), "w") as f:
                f.write(links_file(pkg, deps))
        with open(os.path.join(d, "%s_test.go" % pkg), "w") as f:
            f.write(test_file(pkg, wired, specs, deps))
    # The templates do not try to get struct-field and map-literal alignment
    # right; gofmt does it for us.
    subprocess.run(["gofmt", "-w", os.path.join(ROOT, "internal")], check=True)
    print("generated %d packages x %d entities" % (len(PACKAGES), entities))


if __name__ == "__main__":
    main()
