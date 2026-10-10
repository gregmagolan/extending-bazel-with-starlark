package stocktake

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
		if !strings.HasPrefix(Prefixes[i], "stocktake/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleStocktakeRecord("ok", when).Fields()); err != nil {
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

func TestStocktakeRecordRoundTrip(t *testing.T) {
	x := SampleStocktakeRecord("stocktakerecord-1", when)
	got, err := ParseStocktakeRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeRecord() = %v", err)
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

func TestStocktakeRecordParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeRecord(f); err == nil {
		t.Error("ParseStocktakeRecord() accepted a record with no id")
	}
	f = SampleStocktakeRecord("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseStocktakeRecord(f); err == nil {
		t.Error("ParseStocktakeRecord() accepted a non-numeric version")
	}
}

func TestStocktakeRecordCollections(t *testing.T) {
	xs := []StocktakeRecord{SampleStocktakeRecord("a", when), SampleStocktakeRecord("bb", when), SampleStocktakeRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeRecords(xs)); got != 2 {
		t.Errorf("ActiveStocktakeRecords() kept %d, want 2", got)
	}
	if got, want := SumStocktakeRecords(xs), 1+3; got != want {
		t.Errorf("SumStocktakeRecords() = %d, want %d", got, want)
	}
	SortStocktakeRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeRecords() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeRecords(xs)) != 3 {
		t.Errorf("IndexStocktakeRecords() = %v", IndexStocktakeRecords(xs))
	}
	if err := ValidateStocktakeRecords(ActiveStocktakeRecords(xs)); err != nil {
		t.Errorf("ValidateStocktakeRecords() = %v", err)
	}
}

func TestStocktakeRecordNormalizeClamps(t *testing.T) {
	x := SampleStocktakeRecord("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeRecord("r1", when)
	rev, err := r.PutStocktakeRecord(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeRecord() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeRecords() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeRecord(StocktakeRecord{}, 0); err == nil {
		t.Error("PutStocktakeRecord() accepted an empty entity")
	}
}

func TestStocktakeEntryRoundTrip(t *testing.T) {
	x := SampleStocktakeEntry("stocktakeentry-1", when)
	got, err := ParseStocktakeEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeEntry() = %v", err)
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

func TestStocktakeEntryParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeEntry(f); err == nil {
		t.Error("ParseStocktakeEntry() accepted a record with no id")
	}
	f = SampleStocktakeEntry("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseStocktakeEntry(f); err == nil {
		t.Error("ParseStocktakeEntry() accepted a non-numeric count")
	}
}

func TestStocktakeEntryCollections(t *testing.T) {
	xs := []StocktakeEntry{SampleStocktakeEntry("a", when), SampleStocktakeEntry("bb", when), SampleStocktakeEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeEntrys(xs)); got != 2 {
		t.Errorf("ActiveStocktakeEntrys() kept %d, want 2", got)
	}
	if got, want := SumStocktakeEntrys(xs), 1+3; got != want {
		t.Errorf("SumStocktakeEntrys() = %d, want %d", got, want)
	}
	SortStocktakeEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeEntrys() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeEntrys(xs)) != 3 {
		t.Errorf("IndexStocktakeEntrys() = %v", IndexStocktakeEntrys(xs))
	}
	if err := ValidateStocktakeEntrys(ActiveStocktakeEntrys(xs)); err != nil {
		t.Errorf("ValidateStocktakeEntrys() = %v", err)
	}
}

func TestStocktakeEntryNormalizeClamps(t *testing.T) {
	x := SampleStocktakeEntry("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeEntry("r1", when)
	rev, err := r.PutStocktakeEntry(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeEntry() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeEntry(StocktakeEntry{}, 0); err == nil {
		t.Error("PutStocktakeEntry() accepted an empty entity")
	}
}

func TestStocktakeItemRoundTrip(t *testing.T) {
	x := SampleStocktakeItem("stocktakeitem-1", when)
	got, err := ParseStocktakeItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeItem() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Label != x.Label {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestStocktakeItemParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeItem(f); err == nil {
		t.Error("ParseStocktakeItem() accepted a record with no id")
	}
	f = SampleStocktakeItem("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseStocktakeItem(f); err == nil {
		t.Error("ParseStocktakeItem() accepted a non-numeric retained")
	}
}

func TestStocktakeItemCollections(t *testing.T) {
	xs := []StocktakeItem{SampleStocktakeItem("a", when), SampleStocktakeItem("bb", when), SampleStocktakeItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeItems(xs)); got != 2 {
		t.Errorf("ActiveStocktakeItems() kept %d, want 2", got)
	}
	if got, want := SumStocktakeItems(xs), 1+3; got != want {
		t.Errorf("SumStocktakeItems() = %d, want %d", got, want)
	}
	SortStocktakeItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeItems() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeItems(xs)) != 3 {
		t.Errorf("IndexStocktakeItems() = %v", IndexStocktakeItems(xs))
	}
	if err := ValidateStocktakeItems(ActiveStocktakeItems(xs)); err != nil {
		t.Errorf("ValidateStocktakeItems() = %v", err)
	}
}

func TestStocktakeItemNormalizeClamps(t *testing.T) {
	x := SampleStocktakeItem("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeItem("r1", when)
	rev, err := r.PutStocktakeItem(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeItem() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeItems() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeItem(StocktakeItem{}, 0); err == nil {
		t.Error("PutStocktakeItem() accepted an empty entity")
	}
}

func TestStocktakeDraftRoundTrip(t *testing.T) {
	x := SampleStocktakeDraft("stocktakedraft-1", when)
	got, err := ParseStocktakeDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeDraft() = %v", err)
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

func TestStocktakeDraftParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeDraft(f); err == nil {
		t.Error("ParseStocktakeDraft() accepted a record with no id")
	}
	f = SampleStocktakeDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseStocktakeDraft(f); err == nil {
		t.Error("ParseStocktakeDraft() accepted a non-numeric attempts")
	}
}

func TestStocktakeDraftCollections(t *testing.T) {
	xs := []StocktakeDraft{SampleStocktakeDraft("a", when), SampleStocktakeDraft("bb", when), SampleStocktakeDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeDrafts(xs)); got != 2 {
		t.Errorf("ActiveStocktakeDrafts() kept %d, want 2", got)
	}
	if got, want := SumStocktakeDrafts(xs), 1+3; got != want {
		t.Errorf("SumStocktakeDrafts() = %d, want %d", got, want)
	}
	SortStocktakeDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeDrafts() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeDrafts(xs)) != 3 {
		t.Errorf("IndexStocktakeDrafts() = %v", IndexStocktakeDrafts(xs))
	}
	if err := ValidateStocktakeDrafts(ActiveStocktakeDrafts(xs)); err != nil {
		t.Errorf("ValidateStocktakeDrafts() = %v", err)
	}
}

func TestStocktakeDraftNormalizeClamps(t *testing.T) {
	x := SampleStocktakeDraft("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeDraft("r1", when)
	rev, err := r.PutStocktakeDraft(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeDraft() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeDraft(StocktakeDraft{}, 0); err == nil {
		t.Error("PutStocktakeDraft() accepted an empty entity")
	}
}

func TestStocktakeRevisionRoundTrip(t *testing.T) {
	x := SampleStocktakeRevision("stocktakerevision-1", when)
	got, err := ParseStocktakeRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeRevision() = %v", err)
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

func TestStocktakeRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeRevision(f); err == nil {
		t.Error("ParseStocktakeRevision() accepted a record with no id")
	}
	f = SampleStocktakeRevision("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseStocktakeRevision(f); err == nil {
		t.Error("ParseStocktakeRevision() accepted a non-numeric priority")
	}
}

func TestStocktakeRevisionCollections(t *testing.T) {
	xs := []StocktakeRevision{SampleStocktakeRevision("a", when), SampleStocktakeRevision("bb", when), SampleStocktakeRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeRevisions(xs)); got != 2 {
		t.Errorf("ActiveStocktakeRevisions() kept %d, want 2", got)
	}
	if got, want := SumStocktakeRevisions(xs), 1+3; got != want {
		t.Errorf("SumStocktakeRevisions() = %d, want %d", got, want)
	}
	SortStocktakeRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeRevisions() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeRevisions(xs)) != 3 {
		t.Errorf("IndexStocktakeRevisions() = %v", IndexStocktakeRevisions(xs))
	}
	if err := ValidateStocktakeRevisions(ActiveStocktakeRevisions(xs)); err != nil {
		t.Errorf("ValidateStocktakeRevisions() = %v", err)
	}
}

func TestStocktakeRevisionNormalizeClamps(t *testing.T) {
	x := SampleStocktakeRevision("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeRevision("r1", when)
	rev, err := r.PutStocktakeRevision(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeRevision() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeRevision(StocktakeRevision{}, 0); err == nil {
		t.Error("PutStocktakeRevision() accepted an empty entity")
	}
}

func TestStocktakeSnapshotRoundTrip(t *testing.T) {
	x := SampleStocktakeSnapshot("stocktakesnapshot-1", when)
	got, err := ParseStocktakeSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeSnapshot() = %v", err)
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

func TestStocktakeSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeSnapshot(f); err == nil {
		t.Error("ParseStocktakeSnapshot() accepted a record with no id")
	}
	f = SampleStocktakeSnapshot("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseStocktakeSnapshot(f); err == nil {
		t.Error("ParseStocktakeSnapshot() accepted a non-numeric capacity")
	}
}

func TestStocktakeSnapshotCollections(t *testing.T) {
	xs := []StocktakeSnapshot{SampleStocktakeSnapshot("a", when), SampleStocktakeSnapshot("bb", when), SampleStocktakeSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeSnapshots(xs)); got != 2 {
		t.Errorf("ActiveStocktakeSnapshots() kept %d, want 2", got)
	}
	if got, want := SumStocktakeSnapshots(xs), 1+3; got != want {
		t.Errorf("SumStocktakeSnapshots() = %d, want %d", got, want)
	}
	SortStocktakeSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeSnapshots(xs)) != 3 {
		t.Errorf("IndexStocktakeSnapshots() = %v", IndexStocktakeSnapshots(xs))
	}
	if err := ValidateStocktakeSnapshots(ActiveStocktakeSnapshots(xs)); err != nil {
		t.Errorf("ValidateStocktakeSnapshots() = %v", err)
	}
}

func TestStocktakeSnapshotNormalizeClamps(t *testing.T) {
	x := SampleStocktakeSnapshot("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeSnapshot("r1", when)
	rev, err := r.PutStocktakeSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeSnapshot(StocktakeSnapshot{}, 0); err == nil {
		t.Error("PutStocktakeSnapshot() accepted an empty entity")
	}
}

func TestStocktakeBatchRoundTrip(t *testing.T) {
	x := SampleStocktakeBatch("stocktakebatch-1", when)
	got, err := ParseStocktakeBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeBatch() = %v", err)
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

func TestStocktakeBatchParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeBatch(f); err == nil {
		t.Error("ParseStocktakeBatch() accepted a record with no id")
	}
	f = SampleStocktakeBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseStocktakeBatch(f); err == nil {
		t.Error("ParseStocktakeBatch() accepted a non-numeric capacity")
	}
}

func TestStocktakeBatchCollections(t *testing.T) {
	xs := []StocktakeBatch{SampleStocktakeBatch("a", when), SampleStocktakeBatch("bb", when), SampleStocktakeBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeBatchs(xs)); got != 2 {
		t.Errorf("ActiveStocktakeBatchs() kept %d, want 2", got)
	}
	if got, want := SumStocktakeBatchs(xs), 1+3; got != want {
		t.Errorf("SumStocktakeBatchs() = %d, want %d", got, want)
	}
	SortStocktakeBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeBatchs() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeBatchs(xs)) != 3 {
		t.Errorf("IndexStocktakeBatchs() = %v", IndexStocktakeBatchs(xs))
	}
	if err := ValidateStocktakeBatchs(ActiveStocktakeBatchs(xs)); err != nil {
		t.Errorf("ValidateStocktakeBatchs() = %v", err)
	}
}

func TestStocktakeBatchNormalizeClamps(t *testing.T) {
	x := SampleStocktakeBatch("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeBatch("r1", when)
	rev, err := r.PutStocktakeBatch(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeBatch() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeBatch(StocktakeBatch{}, 0); err == nil {
		t.Error("PutStocktakeBatch() accepted an empty entity")
	}
}

func TestStocktakeRequestRoundTrip(t *testing.T) {
	x := SampleStocktakeRequest("stocktakerequest-1", when)
	got, err := ParseStocktakeRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeRequest() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Label != x.Label {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestStocktakeRequestParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeRequest(f); err == nil {
		t.Error("ParseStocktakeRequest() accepted a record with no id")
	}
	f = SampleStocktakeRequest("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseStocktakeRequest(f); err == nil {
		t.Error("ParseStocktakeRequest() accepted a non-numeric attempts")
	}
}

func TestStocktakeRequestCollections(t *testing.T) {
	xs := []StocktakeRequest{SampleStocktakeRequest("a", when), SampleStocktakeRequest("bb", when), SampleStocktakeRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeRequests(xs)); got != 2 {
		t.Errorf("ActiveStocktakeRequests() kept %d, want 2", got)
	}
	if got, want := SumStocktakeRequests(xs), 1+3; got != want {
		t.Errorf("SumStocktakeRequests() = %d, want %d", got, want)
	}
	SortStocktakeRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeRequests() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeRequests(xs)) != 3 {
		t.Errorf("IndexStocktakeRequests() = %v", IndexStocktakeRequests(xs))
	}
	if err := ValidateStocktakeRequests(ActiveStocktakeRequests(xs)); err != nil {
		t.Errorf("ValidateStocktakeRequests() = %v", err)
	}
}

func TestStocktakeRequestNormalizeClamps(t *testing.T) {
	x := SampleStocktakeRequest("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeRequest("r1", when)
	rev, err := r.PutStocktakeRequest(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeRequest() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeRequests() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeRequest(StocktakeRequest{}, 0); err == nil {
		t.Error("PutStocktakeRequest() accepted an empty entity")
	}
}

func TestStocktakeSummaryRoundTrip(t *testing.T) {
	x := SampleStocktakeSummary("stocktakesummary-1", when)
	got, err := ParseStocktakeSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeSummary() = %v", err)
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

func TestStocktakeSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeSummary(f); err == nil {
		t.Error("ParseStocktakeSummary() accepted a record with no id")
	}
	f = SampleStocktakeSummary("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseStocktakeSummary(f); err == nil {
		t.Error("ParseStocktakeSummary() accepted a non-numeric priority")
	}
}

func TestStocktakeSummaryCollections(t *testing.T) {
	xs := []StocktakeSummary{SampleStocktakeSummary("a", when), SampleStocktakeSummary("bb", when), SampleStocktakeSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeSummarys(xs)); got != 2 {
		t.Errorf("ActiveStocktakeSummarys() kept %d, want 2", got)
	}
	if got, want := SumStocktakeSummarys(xs), 1+3; got != want {
		t.Errorf("SumStocktakeSummarys() = %d, want %d", got, want)
	}
	SortStocktakeSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeSummarys() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeSummarys(xs)) != 3 {
		t.Errorf("IndexStocktakeSummarys() = %v", IndexStocktakeSummarys(xs))
	}
	if err := ValidateStocktakeSummarys(ActiveStocktakeSummarys(xs)); err != nil {
		t.Errorf("ValidateStocktakeSummarys() = %v", err)
	}
}

func TestStocktakeSummaryNormalizeClamps(t *testing.T) {
	x := SampleStocktakeSummary("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeSummary("r1", when)
	rev, err := r.PutStocktakeSummary(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeSummary() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeSummary(StocktakeSummary{}, 0); err == nil {
		t.Error("PutStocktakeSummary() accepted an empty entity")
	}
}

func TestStocktakePolicyRoundTrip(t *testing.T) {
	x := SampleStocktakePolicy("stocktakepolicy-1", when)
	got, err := ParseStocktakePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakePolicy() = %v", err)
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

func TestStocktakePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakePolicy(f); err == nil {
		t.Error("ParseStocktakePolicy() accepted a record with no id")
	}
	f = SampleStocktakePolicy("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseStocktakePolicy(f); err == nil {
		t.Error("ParseStocktakePolicy() accepted a non-numeric count")
	}
}

func TestStocktakePolicyCollections(t *testing.T) {
	xs := []StocktakePolicy{SampleStocktakePolicy("a", when), SampleStocktakePolicy("bb", when), SampleStocktakePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakePolicys(xs)); got != 2 {
		t.Errorf("ActiveStocktakePolicys() kept %d, want 2", got)
	}
	if got, want := SumStocktakePolicys(xs), 1+3; got != want {
		t.Errorf("SumStocktakePolicys() = %d, want %d", got, want)
	}
	SortStocktakePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakePolicys() put %q first", xs[0].ID)
	}
	if len(IndexStocktakePolicys(xs)) != 3 {
		t.Errorf("IndexStocktakePolicys() = %v", IndexStocktakePolicys(xs))
	}
	if err := ValidateStocktakePolicys(ActiveStocktakePolicys(xs)); err != nil {
		t.Errorf("ValidateStocktakePolicys() = %v", err)
	}
}

func TestStocktakePolicyNormalizeClamps(t *testing.T) {
	x := SampleStocktakePolicy("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakePolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakePolicy("r1", when)
	rev, err := r.PutStocktakePolicy(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakePolicy() = %v", err)
	}
	got, gotRev, err := r.GetStocktakePolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakePolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakePolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakePolicys() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakePolicy(StocktakePolicy{}, 0); err == nil {
		t.Error("PutStocktakePolicy() accepted an empty entity")
	}
}

func TestStocktakeLinkRoundTrip(t *testing.T) {
	x := SampleStocktakeLink("stocktakelink-1", when)
	got, err := ParseStocktakeLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeLink() = %v", err)
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

func TestStocktakeLinkParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeLink(f); err == nil {
		t.Error("ParseStocktakeLink() accepted a record with no id")
	}
	f = SampleStocktakeLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseStocktakeLink(f); err == nil {
		t.Error("ParseStocktakeLink() accepted a non-numeric retained")
	}
}

func TestStocktakeLinkCollections(t *testing.T) {
	xs := []StocktakeLink{SampleStocktakeLink("a", when), SampleStocktakeLink("bb", when), SampleStocktakeLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeLinks(xs)); got != 2 {
		t.Errorf("ActiveStocktakeLinks() kept %d, want 2", got)
	}
	if got, want := SumStocktakeLinks(xs), 1+3; got != want {
		t.Errorf("SumStocktakeLinks() = %d, want %d", got, want)
	}
	SortStocktakeLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeLinks() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeLinks(xs)) != 3 {
		t.Errorf("IndexStocktakeLinks() = %v", IndexStocktakeLinks(xs))
	}
	if err := ValidateStocktakeLinks(ActiveStocktakeLinks(xs)); err != nil {
		t.Errorf("ValidateStocktakeLinks() = %v", err)
	}
}

func TestStocktakeLinkNormalizeClamps(t *testing.T) {
	x := SampleStocktakeLink("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeLink("r1", when)
	rev, err := r.PutStocktakeLink(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeLink() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeLinks() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeLink(StocktakeLink{}, 0); err == nil {
		t.Error("PutStocktakeLink() accepted an empty entity")
	}
}

func TestStocktakeNoteRoundTrip(t *testing.T) {
	x := SampleStocktakeNote("stocktakenote-1", when)
	got, err := ParseStocktakeNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeNote() = %v", err)
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

func TestStocktakeNoteParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeNote(f); err == nil {
		t.Error("ParseStocktakeNote() accepted a record with no id")
	}
	f = SampleStocktakeNote("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseStocktakeNote(f); err == nil {
		t.Error("ParseStocktakeNote() accepted a non-numeric version")
	}
}

func TestStocktakeNoteCollections(t *testing.T) {
	xs := []StocktakeNote{SampleStocktakeNote("a", when), SampleStocktakeNote("bb", when), SampleStocktakeNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeNotes(xs)); got != 2 {
		t.Errorf("ActiveStocktakeNotes() kept %d, want 2", got)
	}
	if got, want := SumStocktakeNotes(xs), 1+3; got != want {
		t.Errorf("SumStocktakeNotes() = %d, want %d", got, want)
	}
	SortStocktakeNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeNotes() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeNotes(xs)) != 3 {
		t.Errorf("IndexStocktakeNotes() = %v", IndexStocktakeNotes(xs))
	}
	if err := ValidateStocktakeNotes(ActiveStocktakeNotes(xs)); err != nil {
		t.Errorf("ValidateStocktakeNotes() = %v", err)
	}
}

func TestStocktakeNoteNormalizeClamps(t *testing.T) {
	x := SampleStocktakeNote("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeNote("r1", when)
	rev, err := r.PutStocktakeNote(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeNote() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeNotes() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeNote(StocktakeNote{}, 0); err == nil {
		t.Error("PutStocktakeNote() accepted an empty entity")
	}
}

func TestStocktakeWindowRoundTrip(t *testing.T) {
	x := SampleStocktakeWindow("stocktakewindow-1", when)
	got, err := ParseStocktakeWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeWindow() = %v", err)
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

func TestStocktakeWindowParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeWindow(f); err == nil {
		t.Error("ParseStocktakeWindow() accepted a record with no id")
	}
	f = SampleStocktakeWindow("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseStocktakeWindow(f); err == nil {
		t.Error("ParseStocktakeWindow() accepted a non-numeric capacity")
	}
}

func TestStocktakeWindowCollections(t *testing.T) {
	xs := []StocktakeWindow{SampleStocktakeWindow("a", when), SampleStocktakeWindow("bb", when), SampleStocktakeWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeWindows(xs)); got != 2 {
		t.Errorf("ActiveStocktakeWindows() kept %d, want 2", got)
	}
	if got, want := SumStocktakeWindows(xs), 1+3; got != want {
		t.Errorf("SumStocktakeWindows() = %d, want %d", got, want)
	}
	SortStocktakeWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeWindows() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeWindows(xs)) != 3 {
		t.Errorf("IndexStocktakeWindows() = %v", IndexStocktakeWindows(xs))
	}
	if err := ValidateStocktakeWindows(ActiveStocktakeWindows(xs)); err != nil {
		t.Errorf("ValidateStocktakeWindows() = %v", err)
	}
}

func TestStocktakeWindowNormalizeClamps(t *testing.T) {
	x := SampleStocktakeWindow("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeWindow("r1", when)
	rev, err := r.PutStocktakeWindow(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeWindow() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeWindows() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeWindow(StocktakeWindow{}, 0); err == nil {
		t.Error("PutStocktakeWindow() accepted an empty entity")
	}
}

func TestStocktakeQuotaRoundTrip(t *testing.T) {
	x := SampleStocktakeQuota("stocktakequota-1", when)
	got, err := ParseStocktakeQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeQuota() = %v", err)
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

func TestStocktakeQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeQuota(f); err == nil {
		t.Error("ParseStocktakeQuota() accepted a record with no id")
	}
	f = SampleStocktakeQuota("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseStocktakeQuota(f); err == nil {
		t.Error("ParseStocktakeQuota() accepted a non-numeric weight")
	}
}

func TestStocktakeQuotaCollections(t *testing.T) {
	xs := []StocktakeQuota{SampleStocktakeQuota("a", when), SampleStocktakeQuota("bb", when), SampleStocktakeQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeQuotas(xs)); got != 2 {
		t.Errorf("ActiveStocktakeQuotas() kept %d, want 2", got)
	}
	if got, want := SumStocktakeQuotas(xs), 1+3; got != want {
		t.Errorf("SumStocktakeQuotas() = %d, want %d", got, want)
	}
	SortStocktakeQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeQuotas() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeQuotas(xs)) != 3 {
		t.Errorf("IndexStocktakeQuotas() = %v", IndexStocktakeQuotas(xs))
	}
	if err := ValidateStocktakeQuotas(ActiveStocktakeQuotas(xs)); err != nil {
		t.Errorf("ValidateStocktakeQuotas() = %v", err)
	}
}

func TestStocktakeQuotaNormalizeClamps(t *testing.T) {
	x := SampleStocktakeQuota("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeQuota("r1", when)
	rev, err := r.PutStocktakeQuota(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeQuota() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeQuota(StocktakeQuota{}, 0); err == nil {
		t.Error("PutStocktakeQuota() accepted an empty entity")
	}
}

func TestStocktakeBucketRoundTrip(t *testing.T) {
	x := SampleStocktakeBucket("stocktakebucket-1", when)
	got, err := ParseStocktakeBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeBucket() = %v", err)
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

func TestStocktakeBucketParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeBucket(f); err == nil {
		t.Error("ParseStocktakeBucket() accepted a record with no id")
	}
	f = SampleStocktakeBucket("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseStocktakeBucket(f); err == nil {
		t.Error("ParseStocktakeBucket() accepted a non-numeric priority")
	}
}

func TestStocktakeBucketCollections(t *testing.T) {
	xs := []StocktakeBucket{SampleStocktakeBucket("a", when), SampleStocktakeBucket("bb", when), SampleStocktakeBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeBuckets(xs)); got != 2 {
		t.Errorf("ActiveStocktakeBuckets() kept %d, want 2", got)
	}
	if got, want := SumStocktakeBuckets(xs), 1+3; got != want {
		t.Errorf("SumStocktakeBuckets() = %d, want %d", got, want)
	}
	SortStocktakeBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeBuckets() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeBuckets(xs)) != 3 {
		t.Errorf("IndexStocktakeBuckets() = %v", IndexStocktakeBuckets(xs))
	}
	if err := ValidateStocktakeBuckets(ActiveStocktakeBuckets(xs)); err != nil {
		t.Errorf("ValidateStocktakeBuckets() = %v", err)
	}
}

func TestStocktakeBucketNormalizeClamps(t *testing.T) {
	x := SampleStocktakeBucket("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeBucket("r1", when)
	rev, err := r.PutStocktakeBucket(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeBucket() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeBucket(StocktakeBucket{}, 0); err == nil {
		t.Error("PutStocktakeBucket() accepted an empty entity")
	}
}

func TestStocktakeLaneRoundTrip(t *testing.T) {
	x := SampleStocktakeLane("stocktakelane-1", when)
	got, err := ParseStocktakeLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeLane() = %v", err)
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

func TestStocktakeLaneParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeLane(f); err == nil {
		t.Error("ParseStocktakeLane() accepted a record with no id")
	}
	f = SampleStocktakeLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseStocktakeLane(f); err == nil {
		t.Error("ParseStocktakeLane() accepted a non-numeric capacity")
	}
}

func TestStocktakeLaneCollections(t *testing.T) {
	xs := []StocktakeLane{SampleStocktakeLane("a", when), SampleStocktakeLane("bb", when), SampleStocktakeLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeLanes(xs)); got != 2 {
		t.Errorf("ActiveStocktakeLanes() kept %d, want 2", got)
	}
	if got, want := SumStocktakeLanes(xs), 1+3; got != want {
		t.Errorf("SumStocktakeLanes() = %d, want %d", got, want)
	}
	SortStocktakeLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeLanes() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeLanes(xs)) != 3 {
		t.Errorf("IndexStocktakeLanes() = %v", IndexStocktakeLanes(xs))
	}
	if err := ValidateStocktakeLanes(ActiveStocktakeLanes(xs)); err != nil {
		t.Errorf("ValidateStocktakeLanes() = %v", err)
	}
}

func TestStocktakeLaneNormalizeClamps(t *testing.T) {
	x := SampleStocktakeLane("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeLane("r1", when)
	rev, err := r.PutStocktakeLane(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeLane() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeLanes() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeLane(StocktakeLane{}, 0); err == nil {
		t.Error("PutStocktakeLane() accepted an empty entity")
	}
}

func TestStocktakeSlotRoundTrip(t *testing.T) {
	x := SampleStocktakeSlot("stocktakeslot-1", when)
	got, err := ParseStocktakeSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeSlot() = %v", err)
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

func TestStocktakeSlotParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeSlot(f); err == nil {
		t.Error("ParseStocktakeSlot() accepted a record with no id")
	}
	f = SampleStocktakeSlot("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseStocktakeSlot(f); err == nil {
		t.Error("ParseStocktakeSlot() accepted a non-numeric capacity")
	}
}

func TestStocktakeSlotCollections(t *testing.T) {
	xs := []StocktakeSlot{SampleStocktakeSlot("a", when), SampleStocktakeSlot("bb", when), SampleStocktakeSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeSlots(xs)); got != 2 {
		t.Errorf("ActiveStocktakeSlots() kept %d, want 2", got)
	}
	if got, want := SumStocktakeSlots(xs), 1+3; got != want {
		t.Errorf("SumStocktakeSlots() = %d, want %d", got, want)
	}
	SortStocktakeSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeSlots() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeSlots(xs)) != 3 {
		t.Errorf("IndexStocktakeSlots() = %v", IndexStocktakeSlots(xs))
	}
	if err := ValidateStocktakeSlots(ActiveStocktakeSlots(xs)); err != nil {
		t.Errorf("ValidateStocktakeSlots() = %v", err)
	}
}

func TestStocktakeSlotNormalizeClamps(t *testing.T) {
	x := SampleStocktakeSlot("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeSlot("r1", when)
	rev, err := r.PutStocktakeSlot(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeSlot() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeSlots() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeSlot(StocktakeSlot{}, 0); err == nil {
		t.Error("PutStocktakeSlot() accepted an empty entity")
	}
}

func TestStocktakeTagRoundTrip(t *testing.T) {
	x := SampleStocktakeTag("stocktaketag-1", when)
	got, err := ParseStocktakeTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeTag() = %v", err)
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

func TestStocktakeTagParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeTag(f); err == nil {
		t.Error("ParseStocktakeTag() accepted a record with no id")
	}
	f = SampleStocktakeTag("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseStocktakeTag(f); err == nil {
		t.Error("ParseStocktakeTag() accepted a non-numeric capacity")
	}
}

func TestStocktakeTagCollections(t *testing.T) {
	xs := []StocktakeTag{SampleStocktakeTag("a", when), SampleStocktakeTag("bb", when), SampleStocktakeTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeTags(xs)); got != 2 {
		t.Errorf("ActiveStocktakeTags() kept %d, want 2", got)
	}
	if got, want := SumStocktakeTags(xs), 1+3; got != want {
		t.Errorf("SumStocktakeTags() = %d, want %d", got, want)
	}
	SortStocktakeTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeTags() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeTags(xs)) != 3 {
		t.Errorf("IndexStocktakeTags() = %v", IndexStocktakeTags(xs))
	}
	if err := ValidateStocktakeTags(ActiveStocktakeTags(xs)); err != nil {
		t.Errorf("ValidateStocktakeTags() = %v", err)
	}
}

func TestStocktakeTagNormalizeClamps(t *testing.T) {
	x := SampleStocktakeTag("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeTag("r1", when)
	rev, err := r.PutStocktakeTag(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeTag() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeTags() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeTag(StocktakeTag{}, 0); err == nil {
		t.Error("PutStocktakeTag() accepted an empty entity")
	}
}

func TestStocktakeRuleRoundTrip(t *testing.T) {
	x := SampleStocktakeRule("stocktakerule-1", when)
	got, err := ParseStocktakeRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeRule() = %v", err)
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

func TestStocktakeRuleParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeRule(f); err == nil {
		t.Error("ParseStocktakeRule() accepted a record with no id")
	}
	f = SampleStocktakeRule("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseStocktakeRule(f); err == nil {
		t.Error("ParseStocktakeRule() accepted a non-numeric attempts")
	}
}

func TestStocktakeRuleCollections(t *testing.T) {
	xs := []StocktakeRule{SampleStocktakeRule("a", when), SampleStocktakeRule("bb", when), SampleStocktakeRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeRules(xs)); got != 2 {
		t.Errorf("ActiveStocktakeRules() kept %d, want 2", got)
	}
	if got, want := SumStocktakeRules(xs), 1+3; got != want {
		t.Errorf("SumStocktakeRules() = %d, want %d", got, want)
	}
	SortStocktakeRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeRules() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeRules(xs)) != 3 {
		t.Errorf("IndexStocktakeRules() = %v", IndexStocktakeRules(xs))
	}
	if err := ValidateStocktakeRules(ActiveStocktakeRules(xs)); err != nil {
		t.Errorf("ValidateStocktakeRules() = %v", err)
	}
}

func TestStocktakeRuleNormalizeClamps(t *testing.T) {
	x := SampleStocktakeRule("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeRule("r1", when)
	rev, err := r.PutStocktakeRule(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeRule() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeRules() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeRule(StocktakeRule{}, 0); err == nil {
		t.Error("PutStocktakeRule() accepted an empty entity")
	}
}

func TestStocktakeHoldRoundTrip(t *testing.T) {
	x := SampleStocktakeHold("stocktakehold-1", when)
	got, err := ParseStocktakeHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeHold() = %v", err)
	}
	if got.ID != x.ID || !got.Updated.Equal(x.Updated) || got.Label != x.Label {
		t.Fatalf("round trip = %+v, want %+v", got, x)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("Validate() = %v", err)
	}
	if !strings.Contains(got.String(), got.ID) {
		t.Errorf("String() = %q", got.String())
	}
}

func TestStocktakeHoldParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeHold(f); err == nil {
		t.Error("ParseStocktakeHold() accepted a record with no id")
	}
	f = SampleStocktakeHold("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseStocktakeHold(f); err == nil {
		t.Error("ParseStocktakeHold() accepted a non-numeric attempts")
	}
}

func TestStocktakeHoldCollections(t *testing.T) {
	xs := []StocktakeHold{SampleStocktakeHold("a", when), SampleStocktakeHold("bb", when), SampleStocktakeHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeHolds(xs)); got != 2 {
		t.Errorf("ActiveStocktakeHolds() kept %d, want 2", got)
	}
	if got, want := SumStocktakeHolds(xs), 1+3; got != want {
		t.Errorf("SumStocktakeHolds() = %d, want %d", got, want)
	}
	SortStocktakeHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeHolds() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeHolds(xs)) != 3 {
		t.Errorf("IndexStocktakeHolds() = %v", IndexStocktakeHolds(xs))
	}
	if err := ValidateStocktakeHolds(ActiveStocktakeHolds(xs)); err != nil {
		t.Errorf("ValidateStocktakeHolds() = %v", err)
	}
}

func TestStocktakeHoldNormalizeClamps(t *testing.T) {
	x := SampleStocktakeHold("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeHold("r1", when)
	rev, err := r.PutStocktakeHold(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeHold() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeHolds() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeHold(StocktakeHold{}, 0); err == nil {
		t.Error("PutStocktakeHold() accepted an empty entity")
	}
}

func TestStocktakeClaimRoundTrip(t *testing.T) {
	x := SampleStocktakeClaim("stocktakeclaim-1", when)
	got, err := ParseStocktakeClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeClaim() = %v", err)
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

func TestStocktakeClaimParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeClaim(f); err == nil {
		t.Error("ParseStocktakeClaim() accepted a record with no id")
	}
	f = SampleStocktakeClaim("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseStocktakeClaim(f); err == nil {
		t.Error("ParseStocktakeClaim() accepted a non-numeric attempts")
	}
}

func TestStocktakeClaimCollections(t *testing.T) {
	xs := []StocktakeClaim{SampleStocktakeClaim("a", when), SampleStocktakeClaim("bb", when), SampleStocktakeClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeClaims(xs)); got != 2 {
		t.Errorf("ActiveStocktakeClaims() kept %d, want 2", got)
	}
	if got, want := SumStocktakeClaims(xs), 1+3; got != want {
		t.Errorf("SumStocktakeClaims() = %d, want %d", got, want)
	}
	SortStocktakeClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeClaims() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeClaims(xs)) != 3 {
		t.Errorf("IndexStocktakeClaims() = %v", IndexStocktakeClaims(xs))
	}
	if err := ValidateStocktakeClaims(ActiveStocktakeClaims(xs)); err != nil {
		t.Errorf("ValidateStocktakeClaims() = %v", err)
	}
}

func TestStocktakeClaimNormalizeClamps(t *testing.T) {
	x := SampleStocktakeClaim("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeClaim("r1", when)
	rev, err := r.PutStocktakeClaim(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeClaim() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeClaims() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeClaim(StocktakeClaim{}, 0); err == nil {
		t.Error("PutStocktakeClaim() accepted an empty entity")
	}
}

func TestStocktakeCursorRoundTrip(t *testing.T) {
	x := SampleStocktakeCursor("stocktakecursor-1", when)
	got, err := ParseStocktakeCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseStocktakeCursor() = %v", err)
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

func TestStocktakeCursorParseRejectsBadInput(t *testing.T) {
	f := SampleStocktakeCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseStocktakeCursor(f); err == nil {
		t.Error("ParseStocktakeCursor() accepted a record with no id")
	}
	f = SampleStocktakeCursor("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseStocktakeCursor(f); err == nil {
		t.Error("ParseStocktakeCursor() accepted a non-numeric retained")
	}
}

func TestStocktakeCursorCollections(t *testing.T) {
	xs := []StocktakeCursor{SampleStocktakeCursor("a", when), SampleStocktakeCursor("bb", when), SampleStocktakeCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveStocktakeCursors(xs)); got != 2 {
		t.Errorf("ActiveStocktakeCursors() kept %d, want 2", got)
	}
	if got, want := SumStocktakeCursors(xs), 1+3; got != want {
		t.Errorf("SumStocktakeCursors() = %d, want %d", got, want)
	}
	SortStocktakeCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortStocktakeCursors() put %q first", xs[0].ID)
	}
	if len(IndexStocktakeCursors(xs)) != 3 {
		t.Errorf("IndexStocktakeCursors() = %v", IndexStocktakeCursors(xs))
	}
	if err := ValidateStocktakeCursors(ActiveStocktakeCursors(xs)); err != nil {
		t.Errorf("ValidateStocktakeCursors() = %v", err)
	}
}

func TestStocktakeCursorNormalizeClamps(t *testing.T) {
	x := SampleStocktakeCursor("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestStocktakeCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleStocktakeCursor("r1", when)
	rev, err := r.PutStocktakeCursor(x, 0)
	if err != nil {
		t.Fatalf("PutStocktakeCursor() = %v", err)
	}
	got, gotRev, err := r.GetStocktakeCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetStocktakeCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListStocktakeCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListStocktakeCursors() = %v, %v", all, err)
	}
	if _, err := r.PutStocktakeCursor(StocktakeCursor{}, 0); err == nil {
		t.Error("PutStocktakeCursor() accepted an empty entity")
	}
}
