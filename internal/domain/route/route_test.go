package route

import (
	"strings"
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
)

var when = time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)

func TestSchemasAndPrefixesLineUp(t *testing.T) {
	if len(Schemas) != len(Prefixes) {
		t.Fatalf("%d schemas but %d prefixes", len(Schemas), len(Prefixes))
	}
	for i, s := range Schemas {
		if s.Fingerprint() == "" {
			t.Errorf("schema %d has an empty fingerprint", i)
		}
		if !strings.HasPrefix(Prefixes[i], "route/") {
			t.Errorf("prefix %d = %q", i, Prefixes[i])
		}
	}
}

func TestSummaryNamesEverySchema(t *testing.T) {
	got := Summary()
	for _, s := range Schemas {
		if !strings.Contains(got, s.Name) {
			t.Errorf("Summary() is missing %q", s.Name)
		}
	}
}

func TestValidatorStrictness(t *testing.T) {
	if err := NewValidator(false).Check("nope", codec.Fields{}); err != nil {
		t.Errorf("lenient Check() = %v", err)
	}
	if err := NewValidator(true).Check("nope", codec.Fields{}); err == nil {
		t.Error("strict Check() accepted an unknown schema")
	}
	if err := NewValidator(true).Check(Schemas[0].Name, SampleRouteRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

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

func TestRouteRecordRoundTrip(t *testing.T) {
	x := SampleRouteRecord("routerecord-1", when)
	got, err := ParseRouteRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteRecord() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Category != x.Category {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteRecordParseRejectsBadInput(t *testing.T) {
	f := SampleRouteRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteRecord(f); err == nil {
		t.Error("ParseRouteRecord() accepted a record with no id")
	}
	f = SampleRouteRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRouteRecord(f); err == nil {
		t.Error("ParseRouteRecord() accepted a non-numeric capacity")
	}
}

func TestRouteRecordCollections(t *testing.T) {
	xs := []RouteRecord{SampleRouteRecord("a", when), SampleRouteRecord("bb", when), SampleRouteRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteRecords(xs)); got != 2 {
		t.Errorf("ActiveRouteRecords() kept %d, want 2", got)
	}
	if got, want := SumRouteRecords(xs), 1+3; got != want {
		t.Errorf("SumRouteRecords() = %d, want %d", got, want)
	}
	SortRouteRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteRecords() put %q first", xs[0].ID)
	}
	if len(IndexRouteRecords(xs)) != 3 {
		t.Errorf("IndexRouteRecords() = %v", IndexRouteRecords(xs))
	}
	if err := ValidateRouteRecords(ActiveRouteRecords(xs)); err != nil {
		t.Errorf("ValidateRouteRecords() = %v", err)
	}
}

func TestRouteRecordNormalizeClamps(t *testing.T) {
	x := SampleRouteRecord("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteRecord("r1", when)
	rev, err := r.PutRouteRecord(x, 0)
	if err != nil {
		t.Fatalf("PutRouteRecord() = %v", err)
	}
	got, gotRev, err := r.GetRouteRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteRecords() = %v, %v", all, err)
	}
	if _, err := r.PutRouteRecord(RouteRecord{}, 0); err == nil {
		t.Error("PutRouteRecord() accepted an empty entity")
	}
}

func TestRouteEntryRoundTrip(t *testing.T) {
	x := SampleRouteEntry("routeentry-1", when)
	got, err := ParseRouteEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteEntry() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Region != x.Region {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteEntryParseRejectsBadInput(t *testing.T) {
	f := SampleRouteEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteEntry(f); err == nil {
		t.Error("ParseRouteEntry() accepted a record with no id")
	}
	f = SampleRouteEntry("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRouteEntry(f); err == nil {
		t.Error("ParseRouteEntry() accepted a non-numeric retained")
	}
}

func TestRouteEntryCollections(t *testing.T) {
	xs := []RouteEntry{SampleRouteEntry("a", when), SampleRouteEntry("bb", when), SampleRouteEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteEntrys(xs)); got != 2 {
		t.Errorf("ActiveRouteEntrys() kept %d, want 2", got)
	}
	if got, want := SumRouteEntrys(xs), 1+3; got != want {
		t.Errorf("SumRouteEntrys() = %d, want %d", got, want)
	}
	SortRouteEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteEntrys() put %q first", xs[0].ID)
	}
	if len(IndexRouteEntrys(xs)) != 3 {
		t.Errorf("IndexRouteEntrys() = %v", IndexRouteEntrys(xs))
	}
	if err := ValidateRouteEntrys(ActiveRouteEntrys(xs)); err != nil {
		t.Errorf("ValidateRouteEntrys() = %v", err)
	}
}

func TestRouteEntryNormalizeClamps(t *testing.T) {
	x := SampleRouteEntry("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteEntry("r1", when)
	rev, err := r.PutRouteEntry(x, 0)
	if err != nil {
		t.Fatalf("PutRouteEntry() = %v", err)
	}
	got, gotRev, err := r.GetRouteEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutRouteEntry(RouteEntry{}, 0); err == nil {
		t.Error("PutRouteEntry() accepted an empty entity")
	}
}

func TestRouteItemRoundTrip(t *testing.T) {
	x := SampleRouteItem("routeitem-1", when)
	got, err := ParseRouteItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteItem() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Owner != x.Owner {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteItemParseRejectsBadInput(t *testing.T) {
	f := SampleRouteItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteItem(f); err == nil {
		t.Error("ParseRouteItem() accepted a record with no id")
	}
	f = SampleRouteItem("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseRouteItem(f); err == nil {
		t.Error("ParseRouteItem() accepted a non-numeric attempts")
	}
}

func TestRouteItemCollections(t *testing.T) {
	xs := []RouteItem{SampleRouteItem("a", when), SampleRouteItem("bb", when), SampleRouteItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteItems(xs)); got != 2 {
		t.Errorf("ActiveRouteItems() kept %d, want 2", got)
	}
	if got, want := SumRouteItems(xs), 1+3; got != want {
		t.Errorf("SumRouteItems() = %d, want %d", got, want)
	}
	SortRouteItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteItems() put %q first", xs[0].ID)
	}
	if len(IndexRouteItems(xs)) != 3 {
		t.Errorf("IndexRouteItems() = %v", IndexRouteItems(xs))
	}
	if err := ValidateRouteItems(ActiveRouteItems(xs)); err != nil {
		t.Errorf("ValidateRouteItems() = %v", err)
	}
}

func TestRouteItemNormalizeClamps(t *testing.T) {
	x := SampleRouteItem("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteItem("r1", when)
	rev, err := r.PutRouteItem(x, 0)
	if err != nil {
		t.Fatalf("PutRouteItem() = %v", err)
	}
	got, gotRev, err := r.GetRouteItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteItems() = %v, %v", all, err)
	}
	if _, err := r.PutRouteItem(RouteItem{}, 0); err == nil {
		t.Error("PutRouteItem() accepted an empty entity")
	}
}

func TestRouteDraftRoundTrip(t *testing.T) {
	x := SampleRouteDraft("routedraft-1", when)
	got, err := ParseRouteDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteDraft() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Reference != x.Reference {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteDraftParseRejectsBadInput(t *testing.T) {
	f := SampleRouteDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteDraft(f); err == nil {
		t.Error("ParseRouteDraft() accepted a record with no id")
	}
	f = SampleRouteDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseRouteDraft(f); err == nil {
		t.Error("ParseRouteDraft() accepted a non-numeric attempts")
	}
}

func TestRouteDraftCollections(t *testing.T) {
	xs := []RouteDraft{SampleRouteDraft("a", when), SampleRouteDraft("bb", when), SampleRouteDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteDrafts(xs)); got != 2 {
		t.Errorf("ActiveRouteDrafts() kept %d, want 2", got)
	}
	if got, want := SumRouteDrafts(xs), 1+3; got != want {
		t.Errorf("SumRouteDrafts() = %d, want %d", got, want)
	}
	SortRouteDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteDrafts() put %q first", xs[0].ID)
	}
	if len(IndexRouteDrafts(xs)) != 3 {
		t.Errorf("IndexRouteDrafts() = %v", IndexRouteDrafts(xs))
	}
	if err := ValidateRouteDrafts(ActiveRouteDrafts(xs)); err != nil {
		t.Errorf("ValidateRouteDrafts() = %v", err)
	}
}

func TestRouteDraftNormalizeClamps(t *testing.T) {
	x := SampleRouteDraft("a", when)
	x.Reference = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteDraft("r1", when)
	rev, err := r.PutRouteDraft(x, 0)
	if err != nil {
		t.Fatalf("PutRouteDraft() = %v", err)
	}
	got, gotRev, err := r.GetRouteDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutRouteDraft(RouteDraft{}, 0); err == nil {
		t.Error("PutRouteDraft() accepted an empty entity")
	}
}

func TestRouteRevisionRoundTrip(t *testing.T) {
	x := SampleRouteRevision("routerevision-1", when)
	got, err := ParseRouteRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteRevision() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Reference != x.Reference {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleRouteRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteRevision(f); err == nil {
		t.Error("ParseRouteRevision() accepted a record with no id")
	}
	f = SampleRouteRevision("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRouteRevision(f); err == nil {
		t.Error("ParseRouteRevision() accepted a non-numeric count")
	}
}

func TestRouteRevisionCollections(t *testing.T) {
	xs := []RouteRevision{SampleRouteRevision("a", when), SampleRouteRevision("bb", when), SampleRouteRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteRevisions(xs)); got != 2 {
		t.Errorf("ActiveRouteRevisions() kept %d, want 2", got)
	}
	if got, want := SumRouteRevisions(xs), 1+3; got != want {
		t.Errorf("SumRouteRevisions() = %d, want %d", got, want)
	}
	SortRouteRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteRevisions() put %q first", xs[0].ID)
	}
	if len(IndexRouteRevisions(xs)) != 3 {
		t.Errorf("IndexRouteRevisions() = %v", IndexRouteRevisions(xs))
	}
	if err := ValidateRouteRevisions(ActiveRouteRevisions(xs)); err != nil {
		t.Errorf("ValidateRouteRevisions() = %v", err)
	}
}

func TestRouteRevisionNormalizeClamps(t *testing.T) {
	x := SampleRouteRevision("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteRevision("r1", when)
	rev, err := r.PutRouteRevision(x, 0)
	if err != nil {
		t.Fatalf("PutRouteRevision() = %v", err)
	}
	got, gotRev, err := r.GetRouteRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutRouteRevision(RouteRevision{}, 0); err == nil {
		t.Error("PutRouteRevision() accepted an empty entity")
	}
}

func TestRouteSnapshotRoundTrip(t *testing.T) {
	x := SampleRouteSnapshot("routesnapshot-1", when)
	got, err := ParseRouteSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteSnapshot() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Category != x.Category {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleRouteSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteSnapshot(f); err == nil {
		t.Error("ParseRouteSnapshot() accepted a record with no id")
	}
	f = SampleRouteSnapshot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRouteSnapshot(f); err == nil {
		t.Error("ParseRouteSnapshot() accepted a non-numeric count")
	}
}

func TestRouteSnapshotCollections(t *testing.T) {
	xs := []RouteSnapshot{SampleRouteSnapshot("a", when), SampleRouteSnapshot("bb", when), SampleRouteSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteSnapshots(xs)); got != 2 {
		t.Errorf("ActiveRouteSnapshots() kept %d, want 2", got)
	}
	if got, want := SumRouteSnapshots(xs), 1+3; got != want {
		t.Errorf("SumRouteSnapshots() = %d, want %d", got, want)
	}
	SortRouteSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexRouteSnapshots(xs)) != 3 {
		t.Errorf("IndexRouteSnapshots() = %v", IndexRouteSnapshots(xs))
	}
	if err := ValidateRouteSnapshots(ActiveRouteSnapshots(xs)); err != nil {
		t.Errorf("ValidateRouteSnapshots() = %v", err)
	}
}

func TestRouteSnapshotNormalizeClamps(t *testing.T) {
	x := SampleRouteSnapshot("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteSnapshot("r1", when)
	rev, err := r.PutRouteSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutRouteSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetRouteSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutRouteSnapshot(RouteSnapshot{}, 0); err == nil {
		t.Error("PutRouteSnapshot() accepted an empty entity")
	}
}

func TestRouteBatchRoundTrip(t *testing.T) {
	x := SampleRouteBatch("routebatch-1", when)
	got, err := ParseRouteBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteBatch() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Group != x.Group {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteBatchParseRejectsBadInput(t *testing.T) {
	f := SampleRouteBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteBatch(f); err == nil {
		t.Error("ParseRouteBatch() accepted a record with no id")
	}
	f = SampleRouteBatch("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRouteBatch(f); err == nil {
		t.Error("ParseRouteBatch() accepted a non-numeric offset")
	}
}

func TestRouteBatchCollections(t *testing.T) {
	xs := []RouteBatch{SampleRouteBatch("a", when), SampleRouteBatch("bb", when), SampleRouteBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteBatchs(xs)); got != 2 {
		t.Errorf("ActiveRouteBatchs() kept %d, want 2", got)
	}
	if got, want := SumRouteBatchs(xs), 1+3; got != want {
		t.Errorf("SumRouteBatchs() = %d, want %d", got, want)
	}
	SortRouteBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteBatchs() put %q first", xs[0].ID)
	}
	if len(IndexRouteBatchs(xs)) != 3 {
		t.Errorf("IndexRouteBatchs() = %v", IndexRouteBatchs(xs))
	}
	if err := ValidateRouteBatchs(ActiveRouteBatchs(xs)); err != nil {
		t.Errorf("ValidateRouteBatchs() = %v", err)
	}
}

func TestRouteBatchNormalizeClamps(t *testing.T) {
	x := SampleRouteBatch("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteBatch("r1", when)
	rev, err := r.PutRouteBatch(x, 0)
	if err != nil {
		t.Fatalf("PutRouteBatch() = %v", err)
	}
	got, gotRev, err := r.GetRouteBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutRouteBatch(RouteBatch{}, 0); err == nil {
		t.Error("PutRouteBatch() accepted an empty entity")
	}
}

func TestRouteRequestRoundTrip(t *testing.T) {
	x := SampleRouteRequest("routerequest-1", when)
	got, err := ParseRouteRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteRequest() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Name != x.Name {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteRequestParseRejectsBadInput(t *testing.T) {
	f := SampleRouteRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteRequest(f); err == nil {
		t.Error("ParseRouteRequest() accepted a record with no id")
	}
	f = SampleRouteRequest("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRouteRequest(f); err == nil {
		t.Error("ParseRouteRequest() accepted a non-numeric retained")
	}
}

func TestRouteRequestCollections(t *testing.T) {
	xs := []RouteRequest{SampleRouteRequest("a", when), SampleRouteRequest("bb", when), SampleRouteRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteRequests(xs)); got != 2 {
		t.Errorf("ActiveRouteRequests() kept %d, want 2", got)
	}
	if got, want := SumRouteRequests(xs), 1+3; got != want {
		t.Errorf("SumRouteRequests() = %d, want %d", got, want)
	}
	SortRouteRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteRequests() put %q first", xs[0].ID)
	}
	if len(IndexRouteRequests(xs)) != 3 {
		t.Errorf("IndexRouteRequests() = %v", IndexRouteRequests(xs))
	}
	if err := ValidateRouteRequests(ActiveRouteRequests(xs)); err != nil {
		t.Errorf("ValidateRouteRequests() = %v", err)
	}
}

func TestRouteRequestNormalizeClamps(t *testing.T) {
	x := SampleRouteRequest("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteRequest("r1", when)
	rev, err := r.PutRouteRequest(x, 0)
	if err != nil {
		t.Fatalf("PutRouteRequest() = %v", err)
	}
	got, gotRev, err := r.GetRouteRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteRequests() = %v, %v", all, err)
	}
	if _, err := r.PutRouteRequest(RouteRequest{}, 0); err == nil {
		t.Error("PutRouteRequest() accepted an empty entity")
	}
}

func TestRouteSummaryRoundTrip(t *testing.T) {
	x := SampleRouteSummary("routesummary-1", when)
	got, err := ParseRouteSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteSummary() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Category != x.Category {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleRouteSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteSummary(f); err == nil {
		t.Error("ParseRouteSummary() accepted a record with no id")
	}
	f = SampleRouteSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRouteSummary(f); err == nil {
		t.Error("ParseRouteSummary() accepted a non-numeric capacity")
	}
}

func TestRouteSummaryCollections(t *testing.T) {
	xs := []RouteSummary{SampleRouteSummary("a", when), SampleRouteSummary("bb", when), SampleRouteSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteSummarys(xs)); got != 2 {
		t.Errorf("ActiveRouteSummarys() kept %d, want 2", got)
	}
	if got, want := SumRouteSummarys(xs), 1+3; got != want {
		t.Errorf("SumRouteSummarys() = %d, want %d", got, want)
	}
	SortRouteSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteSummarys() put %q first", xs[0].ID)
	}
	if len(IndexRouteSummarys(xs)) != 3 {
		t.Errorf("IndexRouteSummarys() = %v", IndexRouteSummarys(xs))
	}
	if err := ValidateRouteSummarys(ActiveRouteSummarys(xs)); err != nil {
		t.Errorf("ValidateRouteSummarys() = %v", err)
	}
}

func TestRouteSummaryNormalizeClamps(t *testing.T) {
	x := SampleRouteSummary("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteSummary("r1", when)
	rev, err := r.PutRouteSummary(x, 0)
	if err != nil {
		t.Fatalf("PutRouteSummary() = %v", err)
	}
	got, gotRev, err := r.GetRouteSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutRouteSummary(RouteSummary{}, 0); err == nil {
		t.Error("PutRouteSummary() accepted an empty entity")
	}
}

func TestRoutePolicyRoundTrip(t *testing.T) {
	x := SampleRoutePolicy("routepolicy-1", when)
	got, err := ParseRoutePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseRoutePolicy() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Channel != x.Channel {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRoutePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleRoutePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRoutePolicy(f); err == nil {
		t.Error("ParseRoutePolicy() accepted a record with no id")
	}
	f = SampleRoutePolicy("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseRoutePolicy(f); err == nil {
		t.Error("ParseRoutePolicy() accepted a non-numeric version")
	}
}

func TestRoutePolicyCollections(t *testing.T) {
	xs := []RoutePolicy{SampleRoutePolicy("a", when), SampleRoutePolicy("bb", when), SampleRoutePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRoutePolicys(xs)); got != 2 {
		t.Errorf("ActiveRoutePolicys() kept %d, want 2", got)
	}
	if got, want := SumRoutePolicys(xs), 1+3; got != want {
		t.Errorf("SumRoutePolicys() = %d, want %d", got, want)
	}
	SortRoutePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRoutePolicys() put %q first", xs[0].ID)
	}
	if len(IndexRoutePolicys(xs)) != 3 {
		t.Errorf("IndexRoutePolicys() = %v", IndexRoutePolicys(xs))
	}
	if err := ValidateRoutePolicys(ActiveRoutePolicys(xs)); err != nil {
		t.Errorf("ValidateRoutePolicys() = %v", err)
	}
}

func TestRoutePolicyNormalizeClamps(t *testing.T) {
	x := SampleRoutePolicy("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRoutePolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRoutePolicy("r1", when)
	rev, err := r.PutRoutePolicy(x, 0)
	if err != nil {
		t.Fatalf("PutRoutePolicy() = %v", err)
	}
	got, gotRev, err := r.GetRoutePolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRoutePolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRoutePolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRoutePolicys() = %v, %v", all, err)
	}
	if _, err := r.PutRoutePolicy(RoutePolicy{}, 0); err == nil {
		t.Error("PutRoutePolicy() accepted an empty entity")
	}
}

func TestRouteLinkRoundTrip(t *testing.T) {
	x := SampleRouteLink("routelink-1", when)
	got, err := ParseRouteLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteLink() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Region != x.Region {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteLinkParseRejectsBadInput(t *testing.T) {
	f := SampleRouteLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteLink(f); err == nil {
		t.Error("ParseRouteLink() accepted a record with no id")
	}
	f = SampleRouteLink("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRouteLink(f); err == nil {
		t.Error("ParseRouteLink() accepted a non-numeric offset")
	}
}

func TestRouteLinkCollections(t *testing.T) {
	xs := []RouteLink{SampleRouteLink("a", when), SampleRouteLink("bb", when), SampleRouteLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteLinks(xs)); got != 2 {
		t.Errorf("ActiveRouteLinks() kept %d, want 2", got)
	}
	if got, want := SumRouteLinks(xs), 1+3; got != want {
		t.Errorf("SumRouteLinks() = %d, want %d", got, want)
	}
	SortRouteLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteLinks() put %q first", xs[0].ID)
	}
	if len(IndexRouteLinks(xs)) != 3 {
		t.Errorf("IndexRouteLinks() = %v", IndexRouteLinks(xs))
	}
	if err := ValidateRouteLinks(ActiveRouteLinks(xs)); err != nil {
		t.Errorf("ValidateRouteLinks() = %v", err)
	}
}

func TestRouteLinkNormalizeClamps(t *testing.T) {
	x := SampleRouteLink("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteLink("r1", when)
	rev, err := r.PutRouteLink(x, 0)
	if err != nil {
		t.Fatalf("PutRouteLink() = %v", err)
	}
	got, gotRev, err := r.GetRouteLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteLinks() = %v, %v", all, err)
	}
	if _, err := r.PutRouteLink(RouteLink{}, 0); err == nil {
		t.Error("PutRouteLink() accepted an empty entity")
	}
}

func TestRouteNoteRoundTrip(t *testing.T) {
	x := SampleRouteNote("routenote-1", when)
	got, err := ParseRouteNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteNote() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Channel != x.Channel {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteNoteParseRejectsBadInput(t *testing.T) {
	f := SampleRouteNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteNote(f); err == nil {
		t.Error("ParseRouteNote() accepted a record with no id")
	}
	f = SampleRouteNote("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRouteNote(f); err == nil {
		t.Error("ParseRouteNote() accepted a non-numeric priority")
	}
}

func TestRouteNoteCollections(t *testing.T) {
	xs := []RouteNote{SampleRouteNote("a", when), SampleRouteNote("bb", when), SampleRouteNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteNotes(xs)); got != 2 {
		t.Errorf("ActiveRouteNotes() kept %d, want 2", got)
	}
	if got, want := SumRouteNotes(xs), 1+3; got != want {
		t.Errorf("SumRouteNotes() = %d, want %d", got, want)
	}
	SortRouteNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteNotes() put %q first", xs[0].ID)
	}
	if len(IndexRouteNotes(xs)) != 3 {
		t.Errorf("IndexRouteNotes() = %v", IndexRouteNotes(xs))
	}
	if err := ValidateRouteNotes(ActiveRouteNotes(xs)); err != nil {
		t.Errorf("ValidateRouteNotes() = %v", err)
	}
}

func TestRouteNoteNormalizeClamps(t *testing.T) {
	x := SampleRouteNote("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteNote("r1", when)
	rev, err := r.PutRouteNote(x, 0)
	if err != nil {
		t.Fatalf("PutRouteNote() = %v", err)
	}
	got, gotRev, err := r.GetRouteNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteNotes() = %v, %v", all, err)
	}
	if _, err := r.PutRouteNote(RouteNote{}, 0); err == nil {
		t.Error("PutRouteNote() accepted an empty entity")
	}
}

func TestRouteWindowRoundTrip(t *testing.T) {
	x := SampleRouteWindow("routewindow-1", when)
	got, err := ParseRouteWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteWindow() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Channel != x.Channel {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteWindowParseRejectsBadInput(t *testing.T) {
	f := SampleRouteWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteWindow(f); err == nil {
		t.Error("ParseRouteWindow() accepted a record with no id")
	}
	f = SampleRouteWindow("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRouteWindow(f); err == nil {
		t.Error("ParseRouteWindow() accepted a non-numeric count")
	}
}

func TestRouteWindowCollections(t *testing.T) {
	xs := []RouteWindow{SampleRouteWindow("a", when), SampleRouteWindow("bb", when), SampleRouteWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteWindows(xs)); got != 2 {
		t.Errorf("ActiveRouteWindows() kept %d, want 2", got)
	}
	if got, want := SumRouteWindows(xs), 1+3; got != want {
		t.Errorf("SumRouteWindows() = %d, want %d", got, want)
	}
	SortRouteWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteWindows() put %q first", xs[0].ID)
	}
	if len(IndexRouteWindows(xs)) != 3 {
		t.Errorf("IndexRouteWindows() = %v", IndexRouteWindows(xs))
	}
	if err := ValidateRouteWindows(ActiveRouteWindows(xs)); err != nil {
		t.Errorf("ValidateRouteWindows() = %v", err)
	}
}

func TestRouteWindowNormalizeClamps(t *testing.T) {
	x := SampleRouteWindow("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteWindow("r1", when)
	rev, err := r.PutRouteWindow(x, 0)
	if err != nil {
		t.Fatalf("PutRouteWindow() = %v", err)
	}
	got, gotRev, err := r.GetRouteWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteWindows() = %v, %v", all, err)
	}
	if _, err := r.PutRouteWindow(RouteWindow{}, 0); err == nil {
		t.Error("PutRouteWindow() accepted an empty entity")
	}
}

func TestRouteQuotaRoundTrip(t *testing.T) {
	x := SampleRouteQuota("routequota-1", when)
	got, err := ParseRouteQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteQuota() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Category != x.Category {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleRouteQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteQuota(f); err == nil {
		t.Error("ParseRouteQuota() accepted a record with no id")
	}
	f = SampleRouteQuota("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRouteQuota(f); err == nil {
		t.Error("ParseRouteQuota() accepted a non-numeric priority")
	}
}

func TestRouteQuotaCollections(t *testing.T) {
	xs := []RouteQuota{SampleRouteQuota("a", when), SampleRouteQuota("bb", when), SampleRouteQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteQuotas(xs)); got != 2 {
		t.Errorf("ActiveRouteQuotas() kept %d, want 2", got)
	}
	if got, want := SumRouteQuotas(xs), 1+3; got != want {
		t.Errorf("SumRouteQuotas() = %d, want %d", got, want)
	}
	SortRouteQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteQuotas() put %q first", xs[0].ID)
	}
	if len(IndexRouteQuotas(xs)) != 3 {
		t.Errorf("IndexRouteQuotas() = %v", IndexRouteQuotas(xs))
	}
	if err := ValidateRouteQuotas(ActiveRouteQuotas(xs)); err != nil {
		t.Errorf("ValidateRouteQuotas() = %v", err)
	}
}

func TestRouteQuotaNormalizeClamps(t *testing.T) {
	x := SampleRouteQuota("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteQuota("r1", when)
	rev, err := r.PutRouteQuota(x, 0)
	if err != nil {
		t.Fatalf("PutRouteQuota() = %v", err)
	}
	got, gotRev, err := r.GetRouteQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutRouteQuota(RouteQuota{}, 0); err == nil {
		t.Error("PutRouteQuota() accepted an empty entity")
	}
}

func TestRouteBucketRoundTrip(t *testing.T) {
	x := SampleRouteBucket("routebucket-1", when)
	got, err := ParseRouteBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteBucket() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Source != x.Source {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteBucketParseRejectsBadInput(t *testing.T) {
	f := SampleRouteBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteBucket(f); err == nil {
		t.Error("ParseRouteBucket() accepted a record with no id")
	}
	f = SampleRouteBucket("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseRouteBucket(f); err == nil {
		t.Error("ParseRouteBucket() accepted a non-numeric attempts")
	}
}

func TestRouteBucketCollections(t *testing.T) {
	xs := []RouteBucket{SampleRouteBucket("a", when), SampleRouteBucket("bb", when), SampleRouteBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteBuckets(xs)); got != 2 {
		t.Errorf("ActiveRouteBuckets() kept %d, want 2", got)
	}
	if got, want := SumRouteBuckets(xs), 1+3; got != want {
		t.Errorf("SumRouteBuckets() = %d, want %d", got, want)
	}
	SortRouteBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteBuckets() put %q first", xs[0].ID)
	}
	if len(IndexRouteBuckets(xs)) != 3 {
		t.Errorf("IndexRouteBuckets() = %v", IndexRouteBuckets(xs))
	}
	if err := ValidateRouteBuckets(ActiveRouteBuckets(xs)); err != nil {
		t.Errorf("ValidateRouteBuckets() = %v", err)
	}
}

func TestRouteBucketNormalizeClamps(t *testing.T) {
	x := SampleRouteBucket("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteBucket("r1", when)
	rev, err := r.PutRouteBucket(x, 0)
	if err != nil {
		t.Fatalf("PutRouteBucket() = %v", err)
	}
	got, gotRev, err := r.GetRouteBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutRouteBucket(RouteBucket{}, 0); err == nil {
		t.Error("PutRouteBucket() accepted an empty entity")
	}
}

func TestRouteLaneRoundTrip(t *testing.T) {
	x := SampleRouteLane("routelane-1", when)
	got, err := ParseRouteLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteLane() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Region != x.Region {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteLaneParseRejectsBadInput(t *testing.T) {
	f := SampleRouteLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteLane(f); err == nil {
		t.Error("ParseRouteLane() accepted a record with no id")
	}
	f = SampleRouteLane("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRouteLane(f); err == nil {
		t.Error("ParseRouteLane() accepted a non-numeric weight")
	}
}

func TestRouteLaneCollections(t *testing.T) {
	xs := []RouteLane{SampleRouteLane("a", when), SampleRouteLane("bb", when), SampleRouteLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteLanes(xs)); got != 2 {
		t.Errorf("ActiveRouteLanes() kept %d, want 2", got)
	}
	if got, want := SumRouteLanes(xs), 1+3; got != want {
		t.Errorf("SumRouteLanes() = %d, want %d", got, want)
	}
	SortRouteLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteLanes() put %q first", xs[0].ID)
	}
	if len(IndexRouteLanes(xs)) != 3 {
		t.Errorf("IndexRouteLanes() = %v", IndexRouteLanes(xs))
	}
	if err := ValidateRouteLanes(ActiveRouteLanes(xs)); err != nil {
		t.Errorf("ValidateRouteLanes() = %v", err)
	}
}

func TestRouteLaneNormalizeClamps(t *testing.T) {
	x := SampleRouteLane("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteLane("r1", when)
	rev, err := r.PutRouteLane(x, 0)
	if err != nil {
		t.Fatalf("PutRouteLane() = %v", err)
	}
	got, gotRev, err := r.GetRouteLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteLanes() = %v, %v", all, err)
	}
	if _, err := r.PutRouteLane(RouteLane{}, 0); err == nil {
		t.Error("PutRouteLane() accepted an empty entity")
	}
}

func TestRouteSlotRoundTrip(t *testing.T) {
	x := SampleRouteSlot("routeslot-1", when)
	got, err := ParseRouteSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteSlot() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Category != x.Category {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteSlotParseRejectsBadInput(t *testing.T) {
	f := SampleRouteSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteSlot(f); err == nil {
		t.Error("ParseRouteSlot() accepted a record with no id")
	}
	f = SampleRouteSlot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRouteSlot(f); err == nil {
		t.Error("ParseRouteSlot() accepted a non-numeric offset")
	}
}

func TestRouteSlotCollections(t *testing.T) {
	xs := []RouteSlot{SampleRouteSlot("a", when), SampleRouteSlot("bb", when), SampleRouteSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteSlots(xs)); got != 2 {
		t.Errorf("ActiveRouteSlots() kept %d, want 2", got)
	}
	if got, want := SumRouteSlots(xs), 1+3; got != want {
		t.Errorf("SumRouteSlots() = %d, want %d", got, want)
	}
	SortRouteSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteSlots() put %q first", xs[0].ID)
	}
	if len(IndexRouteSlots(xs)) != 3 {
		t.Errorf("IndexRouteSlots() = %v", IndexRouteSlots(xs))
	}
	if err := ValidateRouteSlots(ActiveRouteSlots(xs)); err != nil {
		t.Errorf("ValidateRouteSlots() = %v", err)
	}
}

func TestRouteSlotNormalizeClamps(t *testing.T) {
	x := SampleRouteSlot("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteSlot("r1", when)
	rev, err := r.PutRouteSlot(x, 0)
	if err != nil {
		t.Fatalf("PutRouteSlot() = %v", err)
	}
	got, gotRev, err := r.GetRouteSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteSlots() = %v, %v", all, err)
	}
	if _, err := r.PutRouteSlot(RouteSlot{}, 0); err == nil {
		t.Error("PutRouteSlot() accepted an empty entity")
	}
}

func TestRouteTagRoundTrip(t *testing.T) {
	x := SampleRouteTag("routetag-1", when)
	got, err := ParseRouteTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteTag() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Group != x.Group {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteTagParseRejectsBadInput(t *testing.T) {
	f := SampleRouteTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteTag(f); err == nil {
		t.Error("ParseRouteTag() accepted a record with no id")
	}
	f = SampleRouteTag("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRouteTag(f); err == nil {
		t.Error("ParseRouteTag() accepted a non-numeric retained")
	}
}

func TestRouteTagCollections(t *testing.T) {
	xs := []RouteTag{SampleRouteTag("a", when), SampleRouteTag("bb", when), SampleRouteTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteTags(xs)); got != 2 {
		t.Errorf("ActiveRouteTags() kept %d, want 2", got)
	}
	if got, want := SumRouteTags(xs), 1+3; got != want {
		t.Errorf("SumRouteTags() = %d, want %d", got, want)
	}
	SortRouteTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteTags() put %q first", xs[0].ID)
	}
	if len(IndexRouteTags(xs)) != 3 {
		t.Errorf("IndexRouteTags() = %v", IndexRouteTags(xs))
	}
	if err := ValidateRouteTags(ActiveRouteTags(xs)); err != nil {
		t.Errorf("ValidateRouteTags() = %v", err)
	}
}

func TestRouteTagNormalizeClamps(t *testing.T) {
	x := SampleRouteTag("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteTag("r1", when)
	rev, err := r.PutRouteTag(x, 0)
	if err != nil {
		t.Fatalf("PutRouteTag() = %v", err)
	}
	got, gotRev, err := r.GetRouteTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteTags() = %v, %v", all, err)
	}
	if _, err := r.PutRouteTag(RouteTag{}, 0); err == nil {
		t.Error("PutRouteTag() accepted an empty entity")
	}
}

func TestRouteRuleRoundTrip(t *testing.T) {
	x := SampleRouteRule("routerule-1", when)
	got, err := ParseRouteRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteRule() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Source != x.Source {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteRuleParseRejectsBadInput(t *testing.T) {
	f := SampleRouteRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteRule(f); err == nil {
		t.Error("ParseRouteRule() accepted a record with no id")
	}
	f = SampleRouteRule("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRouteRule(f); err == nil {
		t.Error("ParseRouteRule() accepted a non-numeric priority")
	}
}

func TestRouteRuleCollections(t *testing.T) {
	xs := []RouteRule{SampleRouteRule("a", when), SampleRouteRule("bb", when), SampleRouteRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteRules(xs)); got != 2 {
		t.Errorf("ActiveRouteRules() kept %d, want 2", got)
	}
	if got, want := SumRouteRules(xs), 1+3; got != want {
		t.Errorf("SumRouteRules() = %d, want %d", got, want)
	}
	SortRouteRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteRules() put %q first", xs[0].ID)
	}
	if len(IndexRouteRules(xs)) != 3 {
		t.Errorf("IndexRouteRules() = %v", IndexRouteRules(xs))
	}
	if err := ValidateRouteRules(ActiveRouteRules(xs)); err != nil {
		t.Errorf("ValidateRouteRules() = %v", err)
	}
}

func TestRouteRuleNormalizeClamps(t *testing.T) {
	x := SampleRouteRule("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteRule("r1", when)
	rev, err := r.PutRouteRule(x, 0)
	if err != nil {
		t.Fatalf("PutRouteRule() = %v", err)
	}
	got, gotRev, err := r.GetRouteRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteRules() = %v, %v", all, err)
	}
	if _, err := r.PutRouteRule(RouteRule{}, 0); err == nil {
		t.Error("PutRouteRule() accepted an empty entity")
	}
}

func TestRouteHoldRoundTrip(t *testing.T) {
	x := SampleRouteHold("routehold-1", when)
	got, err := ParseRouteHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteHold() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Name != x.Name {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteHoldParseRejectsBadInput(t *testing.T) {
	f := SampleRouteHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteHold(f); err == nil {
		t.Error("ParseRouteHold() accepted a record with no id")
	}
	f = SampleRouteHold("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRouteHold(f); err == nil {
		t.Error("ParseRouteHold() accepted a non-numeric retained")
	}
}

func TestRouteHoldCollections(t *testing.T) {
	xs := []RouteHold{SampleRouteHold("a", when), SampleRouteHold("bb", when), SampleRouteHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteHolds(xs)); got != 2 {
		t.Errorf("ActiveRouteHolds() kept %d, want 2", got)
	}
	if got, want := SumRouteHolds(xs), 1+3; got != want {
		t.Errorf("SumRouteHolds() = %d, want %d", got, want)
	}
	SortRouteHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteHolds() put %q first", xs[0].ID)
	}
	if len(IndexRouteHolds(xs)) != 3 {
		t.Errorf("IndexRouteHolds() = %v", IndexRouteHolds(xs))
	}
	if err := ValidateRouteHolds(ActiveRouteHolds(xs)); err != nil {
		t.Errorf("ValidateRouteHolds() = %v", err)
	}
}

func TestRouteHoldNormalizeClamps(t *testing.T) {
	x := SampleRouteHold("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteHold("r1", when)
	rev, err := r.PutRouteHold(x, 0)
	if err != nil {
		t.Fatalf("PutRouteHold() = %v", err)
	}
	got, gotRev, err := r.GetRouteHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteHolds() = %v, %v", all, err)
	}
	if _, err := r.PutRouteHold(RouteHold{}, 0); err == nil {
		t.Error("PutRouteHold() accepted an empty entity")
	}
}

func TestRouteClaimRoundTrip(t *testing.T) {
	x := SampleRouteClaim("routeclaim-1", when)
	got, err := ParseRouteClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteClaim() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Category != x.Category {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteClaimParseRejectsBadInput(t *testing.T) {
	f := SampleRouteClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteClaim(f); err == nil {
		t.Error("ParseRouteClaim() accepted a record with no id")
	}
	f = SampleRouteClaim("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRouteClaim(f); err == nil {
		t.Error("ParseRouteClaim() accepted a non-numeric priority")
	}
}

func TestRouteClaimCollections(t *testing.T) {
	xs := []RouteClaim{SampleRouteClaim("a", when), SampleRouteClaim("bb", when), SampleRouteClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteClaims(xs)); got != 2 {
		t.Errorf("ActiveRouteClaims() kept %d, want 2", got)
	}
	if got, want := SumRouteClaims(xs), 1+3; got != want {
		t.Errorf("SumRouteClaims() = %d, want %d", got, want)
	}
	SortRouteClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteClaims() put %q first", xs[0].ID)
	}
	if len(IndexRouteClaims(xs)) != 3 {
		t.Errorf("IndexRouteClaims() = %v", IndexRouteClaims(xs))
	}
	if err := ValidateRouteClaims(ActiveRouteClaims(xs)); err != nil {
		t.Errorf("ValidateRouteClaims() = %v", err)
	}
}

func TestRouteClaimNormalizeClamps(t *testing.T) {
	x := SampleRouteClaim("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteClaim("r1", when)
	rev, err := r.PutRouteClaim(x, 0)
	if err != nil {
		t.Fatalf("PutRouteClaim() = %v", err)
	}
	got, gotRev, err := r.GetRouteClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteClaims() = %v, %v", all, err)
	}
	if _, err := r.PutRouteClaim(RouteClaim{}, 0); err == nil {
		t.Error("PutRouteClaim() accepted an empty entity")
	}
}

func TestRouteCursorRoundTrip(t *testing.T) {
	x := SampleRouteCursor("routecursor-1", when)
	got, err := ParseRouteCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseRouteCursor() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Comment != x.Comment {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestRouteCursorParseRejectsBadInput(t *testing.T) {
	f := SampleRouteCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRouteCursor(f); err == nil {
		t.Error("ParseRouteCursor() accepted a record with no id")
	}
	f = SampleRouteCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRouteCursor(f); err == nil {
		t.Error("ParseRouteCursor() accepted a non-numeric weight")
	}
}

func TestRouteCursorCollections(t *testing.T) {
	xs := []RouteCursor{SampleRouteCursor("a", when), SampleRouteCursor("bb", when), SampleRouteCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRouteCursors(xs)); got != 2 {
		t.Errorf("ActiveRouteCursors() kept %d, want 2", got)
	}
	if got, want := SumRouteCursors(xs), 1+3; got != want {
		t.Errorf("SumRouteCursors() = %d, want %d", got, want)
	}
	SortRouteCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRouteCursors() put %q first", xs[0].ID)
	}
	if len(IndexRouteCursors(xs)) != 3 {
		t.Errorf("IndexRouteCursors() = %v", IndexRouteCursors(xs))
	}
	if err := ValidateRouteCursors(ActiveRouteCursors(xs)); err != nil {
		t.Errorf("ValidateRouteCursors() = %v", err)
	}
}

func TestRouteCursorNormalizeClamps(t *testing.T) {
	x := SampleRouteCursor("a", when)
	x.Comment = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRouteCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRouteCursor("r1", when)
	rev, err := r.PutRouteCursor(x, 0)
	if err != nil {
		t.Fatalf("PutRouteCursor() = %v", err)
	}
	got, gotRev, err := r.GetRouteCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRouteCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRouteCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRouteCursors() = %v, %v", all, err)
	}
	if _, err := r.PutRouteCursor(RouteCursor{}, 0); err == nil {
		t.Error("PutRouteCursor() accepted an empty entity")
	}
}
