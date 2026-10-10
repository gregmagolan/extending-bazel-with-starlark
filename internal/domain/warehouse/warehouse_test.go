package warehouse

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
		if !strings.HasPrefix(Prefixes[i], "warehouse/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleWarehouseRecord("ok", when).Fields()); err != nil {
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

func TestWarehouseRecordRoundTrip(t *testing.T) {
	x := SampleWarehouseRecord("warehouserecord-1", when)
	got, err := ParseWarehouseRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseRecord() = %v", err)
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

func TestWarehouseRecordParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseRecord(f); err == nil {
		t.Error("ParseWarehouseRecord() accepted a record with no id")
	}
	f = SampleWarehouseRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseWarehouseRecord(f); err == nil {
		t.Error("ParseWarehouseRecord() accepted a non-numeric capacity")
	}
}

func TestWarehouseRecordCollections(t *testing.T) {
	xs := []WarehouseRecord{SampleWarehouseRecord("a", when), SampleWarehouseRecord("bb", when), SampleWarehouseRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseRecords(xs)); got != 2 {
		t.Errorf("ActiveWarehouseRecords() kept %d, want 2", got)
	}
	if got, want := SumWarehouseRecords(xs), 1+3; got != want {
		t.Errorf("SumWarehouseRecords() = %d, want %d", got, want)
	}
	SortWarehouseRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseRecords() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseRecords(xs)) != 3 {
		t.Errorf("IndexWarehouseRecords() = %v", IndexWarehouseRecords(xs))
	}
	if err := ValidateWarehouseRecords(ActiveWarehouseRecords(xs)); err != nil {
		t.Errorf("ValidateWarehouseRecords() = %v", err)
	}
}

func TestWarehouseRecordNormalizeClamps(t *testing.T) {
	x := SampleWarehouseRecord("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseRecord("r1", when)
	rev, err := r.PutWarehouseRecord(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseRecord() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseRecords() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseRecord(WarehouseRecord{}, 0); err == nil {
		t.Error("PutWarehouseRecord() accepted an empty entity")
	}
}

func TestWarehouseEntryRoundTrip(t *testing.T) {
	x := SampleWarehouseEntry("warehouseentry-1", when)
	got, err := ParseWarehouseEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseEntry() = %v", err)
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

func TestWarehouseEntryParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseEntry(f); err == nil {
		t.Error("ParseWarehouseEntry() accepted a record with no id")
	}
	f = SampleWarehouseEntry("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseWarehouseEntry(f); err == nil {
		t.Error("ParseWarehouseEntry() accepted a non-numeric weight")
	}
}

func TestWarehouseEntryCollections(t *testing.T) {
	xs := []WarehouseEntry{SampleWarehouseEntry("a", when), SampleWarehouseEntry("bb", when), SampleWarehouseEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseEntrys(xs)); got != 2 {
		t.Errorf("ActiveWarehouseEntrys() kept %d, want 2", got)
	}
	if got, want := SumWarehouseEntrys(xs), 1+3; got != want {
		t.Errorf("SumWarehouseEntrys() = %d, want %d", got, want)
	}
	SortWarehouseEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseEntrys() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseEntrys(xs)) != 3 {
		t.Errorf("IndexWarehouseEntrys() = %v", IndexWarehouseEntrys(xs))
	}
	if err := ValidateWarehouseEntrys(ActiveWarehouseEntrys(xs)); err != nil {
		t.Errorf("ValidateWarehouseEntrys() = %v", err)
	}
}

func TestWarehouseEntryNormalizeClamps(t *testing.T) {
	x := SampleWarehouseEntry("a", when)
	x.Comment = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseEntry("r1", when)
	rev, err := r.PutWarehouseEntry(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseEntry() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseEntry(WarehouseEntry{}, 0); err == nil {
		t.Error("PutWarehouseEntry() accepted an empty entity")
	}
}

func TestWarehouseItemRoundTrip(t *testing.T) {
	x := SampleWarehouseItem("warehouseitem-1", when)
	got, err := ParseWarehouseItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseItem() = %v", err)
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

func TestWarehouseItemParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseItem(f); err == nil {
		t.Error("ParseWarehouseItem() accepted a record with no id")
	}
	f = SampleWarehouseItem("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseWarehouseItem(f); err == nil {
		t.Error("ParseWarehouseItem() accepted a non-numeric offset")
	}
}

func TestWarehouseItemCollections(t *testing.T) {
	xs := []WarehouseItem{SampleWarehouseItem("a", when), SampleWarehouseItem("bb", when), SampleWarehouseItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseItems(xs)); got != 2 {
		t.Errorf("ActiveWarehouseItems() kept %d, want 2", got)
	}
	if got, want := SumWarehouseItems(xs), 1+3; got != want {
		t.Errorf("SumWarehouseItems() = %d, want %d", got, want)
	}
	SortWarehouseItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseItems() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseItems(xs)) != 3 {
		t.Errorf("IndexWarehouseItems() = %v", IndexWarehouseItems(xs))
	}
	if err := ValidateWarehouseItems(ActiveWarehouseItems(xs)); err != nil {
		t.Errorf("ValidateWarehouseItems() = %v", err)
	}
}

func TestWarehouseItemNormalizeClamps(t *testing.T) {
	x := SampleWarehouseItem("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseItem("r1", when)
	rev, err := r.PutWarehouseItem(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseItem() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseItems() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseItem(WarehouseItem{}, 0); err == nil {
		t.Error("PutWarehouseItem() accepted an empty entity")
	}
}

func TestWarehouseDraftRoundTrip(t *testing.T) {
	x := SampleWarehouseDraft("warehousedraft-1", when)
	got, err := ParseWarehouseDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseDraft() = %v", err)
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

func TestWarehouseDraftParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseDraft(f); err == nil {
		t.Error("ParseWarehouseDraft() accepted a record with no id")
	}
	f = SampleWarehouseDraft("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWarehouseDraft(f); err == nil {
		t.Error("ParseWarehouseDraft() accepted a non-numeric priority")
	}
}

func TestWarehouseDraftCollections(t *testing.T) {
	xs := []WarehouseDraft{SampleWarehouseDraft("a", when), SampleWarehouseDraft("bb", when), SampleWarehouseDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseDrafts(xs)); got != 2 {
		t.Errorf("ActiveWarehouseDrafts() kept %d, want 2", got)
	}
	if got, want := SumWarehouseDrafts(xs), 1+3; got != want {
		t.Errorf("SumWarehouseDrafts() = %d, want %d", got, want)
	}
	SortWarehouseDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseDrafts() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseDrafts(xs)) != 3 {
		t.Errorf("IndexWarehouseDrafts() = %v", IndexWarehouseDrafts(xs))
	}
	if err := ValidateWarehouseDrafts(ActiveWarehouseDrafts(xs)); err != nil {
		t.Errorf("ValidateWarehouseDrafts() = %v", err)
	}
}

func TestWarehouseDraftNormalizeClamps(t *testing.T) {
	x := SampleWarehouseDraft("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseDraft("r1", when)
	rev, err := r.PutWarehouseDraft(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseDraft() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseDraft(WarehouseDraft{}, 0); err == nil {
		t.Error("PutWarehouseDraft() accepted an empty entity")
	}
}

func TestWarehouseRevisionRoundTrip(t *testing.T) {
	x := SampleWarehouseRevision("warehouserevision-1", when)
	got, err := ParseWarehouseRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseRevision() = %v", err)
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

func TestWarehouseRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseRevision(f); err == nil {
		t.Error("ParseWarehouseRevision() accepted a record with no id")
	}
	f = SampleWarehouseRevision("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseWarehouseRevision(f); err == nil {
		t.Error("ParseWarehouseRevision() accepted a non-numeric weight")
	}
}

func TestWarehouseRevisionCollections(t *testing.T) {
	xs := []WarehouseRevision{SampleWarehouseRevision("a", when), SampleWarehouseRevision("bb", when), SampleWarehouseRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseRevisions(xs)); got != 2 {
		t.Errorf("ActiveWarehouseRevisions() kept %d, want 2", got)
	}
	if got, want := SumWarehouseRevisions(xs), 1+3; got != want {
		t.Errorf("SumWarehouseRevisions() = %d, want %d", got, want)
	}
	SortWarehouseRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseRevisions() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseRevisions(xs)) != 3 {
		t.Errorf("IndexWarehouseRevisions() = %v", IndexWarehouseRevisions(xs))
	}
	if err := ValidateWarehouseRevisions(ActiveWarehouseRevisions(xs)); err != nil {
		t.Errorf("ValidateWarehouseRevisions() = %v", err)
	}
}

func TestWarehouseRevisionNormalizeClamps(t *testing.T) {
	x := SampleWarehouseRevision("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseRevision("r1", when)
	rev, err := r.PutWarehouseRevision(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseRevision() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseRevision(WarehouseRevision{}, 0); err == nil {
		t.Error("PutWarehouseRevision() accepted an empty entity")
	}
}

func TestWarehouseSnapshotRoundTrip(t *testing.T) {
	x := SampleWarehouseSnapshot("warehousesnapshot-1", when)
	got, err := ParseWarehouseSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseSnapshot() = %v", err)
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

func TestWarehouseSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseSnapshot(f); err == nil {
		t.Error("ParseWarehouseSnapshot() accepted a record with no id")
	}
	f = SampleWarehouseSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWarehouseSnapshot(f); err == nil {
		t.Error("ParseWarehouseSnapshot() accepted a non-numeric priority")
	}
}

func TestWarehouseSnapshotCollections(t *testing.T) {
	xs := []WarehouseSnapshot{SampleWarehouseSnapshot("a", when), SampleWarehouseSnapshot("bb", when), SampleWarehouseSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseSnapshots(xs)); got != 2 {
		t.Errorf("ActiveWarehouseSnapshots() kept %d, want 2", got)
	}
	if got, want := SumWarehouseSnapshots(xs), 1+3; got != want {
		t.Errorf("SumWarehouseSnapshots() = %d, want %d", got, want)
	}
	SortWarehouseSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseSnapshots(xs)) != 3 {
		t.Errorf("IndexWarehouseSnapshots() = %v", IndexWarehouseSnapshots(xs))
	}
	if err := ValidateWarehouseSnapshots(ActiveWarehouseSnapshots(xs)); err != nil {
		t.Errorf("ValidateWarehouseSnapshots() = %v", err)
	}
}

func TestWarehouseSnapshotNormalizeClamps(t *testing.T) {
	x := SampleWarehouseSnapshot("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseSnapshot("r1", when)
	rev, err := r.PutWarehouseSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseSnapshot(WarehouseSnapshot{}, 0); err == nil {
		t.Error("PutWarehouseSnapshot() accepted an empty entity")
	}
}

func TestWarehouseBatchRoundTrip(t *testing.T) {
	x := SampleWarehouseBatch("warehousebatch-1", when)
	got, err := ParseWarehouseBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseBatch() = %v", err)
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

func TestWarehouseBatchParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseBatch(f); err == nil {
		t.Error("ParseWarehouseBatch() accepted a record with no id")
	}
	f = SampleWarehouseBatch("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseWarehouseBatch(f); err == nil {
		t.Error("ParseWarehouseBatch() accepted a non-numeric count")
	}
}

func TestWarehouseBatchCollections(t *testing.T) {
	xs := []WarehouseBatch{SampleWarehouseBatch("a", when), SampleWarehouseBatch("bb", when), SampleWarehouseBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseBatchs(xs)); got != 2 {
		t.Errorf("ActiveWarehouseBatchs() kept %d, want 2", got)
	}
	if got, want := SumWarehouseBatchs(xs), 1+3; got != want {
		t.Errorf("SumWarehouseBatchs() = %d, want %d", got, want)
	}
	SortWarehouseBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseBatchs() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseBatchs(xs)) != 3 {
		t.Errorf("IndexWarehouseBatchs() = %v", IndexWarehouseBatchs(xs))
	}
	if err := ValidateWarehouseBatchs(ActiveWarehouseBatchs(xs)); err != nil {
		t.Errorf("ValidateWarehouseBatchs() = %v", err)
	}
}

func TestWarehouseBatchNormalizeClamps(t *testing.T) {
	x := SampleWarehouseBatch("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseBatch("r1", when)
	rev, err := r.PutWarehouseBatch(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseBatch() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseBatch(WarehouseBatch{}, 0); err == nil {
		t.Error("PutWarehouseBatch() accepted an empty entity")
	}
}

func TestWarehouseRequestRoundTrip(t *testing.T) {
	x := SampleWarehouseRequest("warehouserequest-1", when)
	got, err := ParseWarehouseRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseRequest() = %v", err)
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

func TestWarehouseRequestParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseRequest(f); err == nil {
		t.Error("ParseWarehouseRequest() accepted a record with no id")
	}
	f = SampleWarehouseRequest("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseWarehouseRequest(f); err == nil {
		t.Error("ParseWarehouseRequest() accepted a non-numeric attempts")
	}
}

func TestWarehouseRequestCollections(t *testing.T) {
	xs := []WarehouseRequest{SampleWarehouseRequest("a", when), SampleWarehouseRequest("bb", when), SampleWarehouseRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseRequests(xs)); got != 2 {
		t.Errorf("ActiveWarehouseRequests() kept %d, want 2", got)
	}
	if got, want := SumWarehouseRequests(xs), 1+3; got != want {
		t.Errorf("SumWarehouseRequests() = %d, want %d", got, want)
	}
	SortWarehouseRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseRequests() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseRequests(xs)) != 3 {
		t.Errorf("IndexWarehouseRequests() = %v", IndexWarehouseRequests(xs))
	}
	if err := ValidateWarehouseRequests(ActiveWarehouseRequests(xs)); err != nil {
		t.Errorf("ValidateWarehouseRequests() = %v", err)
	}
}

func TestWarehouseRequestNormalizeClamps(t *testing.T) {
	x := SampleWarehouseRequest("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseRequest("r1", when)
	rev, err := r.PutWarehouseRequest(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseRequest() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseRequests() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseRequest(WarehouseRequest{}, 0); err == nil {
		t.Error("PutWarehouseRequest() accepted an empty entity")
	}
}

func TestWarehouseSummaryRoundTrip(t *testing.T) {
	x := SampleWarehouseSummary("warehousesummary-1", when)
	got, err := ParseWarehouseSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseSummary() = %v", err)
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

func TestWarehouseSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseSummary(f); err == nil {
		t.Error("ParseWarehouseSummary() accepted a record with no id")
	}
	f = SampleWarehouseSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseWarehouseSummary(f); err == nil {
		t.Error("ParseWarehouseSummary() accepted a non-numeric capacity")
	}
}

func TestWarehouseSummaryCollections(t *testing.T) {
	xs := []WarehouseSummary{SampleWarehouseSummary("a", when), SampleWarehouseSummary("bb", when), SampleWarehouseSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseSummarys(xs)); got != 2 {
		t.Errorf("ActiveWarehouseSummarys() kept %d, want 2", got)
	}
	if got, want := SumWarehouseSummarys(xs), 1+3; got != want {
		t.Errorf("SumWarehouseSummarys() = %d, want %d", got, want)
	}
	SortWarehouseSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseSummarys() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseSummarys(xs)) != 3 {
		t.Errorf("IndexWarehouseSummarys() = %v", IndexWarehouseSummarys(xs))
	}
	if err := ValidateWarehouseSummarys(ActiveWarehouseSummarys(xs)); err != nil {
		t.Errorf("ValidateWarehouseSummarys() = %v", err)
	}
}

func TestWarehouseSummaryNormalizeClamps(t *testing.T) {
	x := SampleWarehouseSummary("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseSummary("r1", when)
	rev, err := r.PutWarehouseSummary(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseSummary() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseSummary(WarehouseSummary{}, 0); err == nil {
		t.Error("PutWarehouseSummary() accepted an empty entity")
	}
}

func TestWarehousePolicyRoundTrip(t *testing.T) {
	x := SampleWarehousePolicy("warehousepolicy-1", when)
	got, err := ParseWarehousePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehousePolicy() = %v", err)
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

func TestWarehousePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleWarehousePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehousePolicy(f); err == nil {
		t.Error("ParseWarehousePolicy() accepted a record with no id")
	}
	f = SampleWarehousePolicy("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseWarehousePolicy(f); err == nil {
		t.Error("ParseWarehousePolicy() accepted a non-numeric version")
	}
}

func TestWarehousePolicyCollections(t *testing.T) {
	xs := []WarehousePolicy{SampleWarehousePolicy("a", when), SampleWarehousePolicy("bb", when), SampleWarehousePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehousePolicys(xs)); got != 2 {
		t.Errorf("ActiveWarehousePolicys() kept %d, want 2", got)
	}
	if got, want := SumWarehousePolicys(xs), 1+3; got != want {
		t.Errorf("SumWarehousePolicys() = %d, want %d", got, want)
	}
	SortWarehousePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehousePolicys() put %q first", xs[0].ID)
	}
	if len(IndexWarehousePolicys(xs)) != 3 {
		t.Errorf("IndexWarehousePolicys() = %v", IndexWarehousePolicys(xs))
	}
	if err := ValidateWarehousePolicys(ActiveWarehousePolicys(xs)); err != nil {
		t.Errorf("ValidateWarehousePolicys() = %v", err)
	}
}

func TestWarehousePolicyNormalizeClamps(t *testing.T) {
	x := SampleWarehousePolicy("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehousePolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehousePolicy("r1", when)
	rev, err := r.PutWarehousePolicy(x, 0)
	if err != nil {
		t.Fatalf("PutWarehousePolicy() = %v", err)
	}
	got, gotRev, err := r.GetWarehousePolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehousePolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehousePolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehousePolicys() = %v, %v", all, err)
	}
	if _, err := r.PutWarehousePolicy(WarehousePolicy{}, 0); err == nil {
		t.Error("PutWarehousePolicy() accepted an empty entity")
	}
}

func TestWarehouseLinkRoundTrip(t *testing.T) {
	x := SampleWarehouseLink("warehouselink-1", when)
	got, err := ParseWarehouseLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseLink() = %v", err)
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

func TestWarehouseLinkParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseLink(f); err == nil {
		t.Error("ParseWarehouseLink() accepted a record with no id")
	}
	f = SampleWarehouseLink("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseWarehouseLink(f); err == nil {
		t.Error("ParseWarehouseLink() accepted a non-numeric attempts")
	}
}

func TestWarehouseLinkCollections(t *testing.T) {
	xs := []WarehouseLink{SampleWarehouseLink("a", when), SampleWarehouseLink("bb", when), SampleWarehouseLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseLinks(xs)); got != 2 {
		t.Errorf("ActiveWarehouseLinks() kept %d, want 2", got)
	}
	if got, want := SumWarehouseLinks(xs), 1+3; got != want {
		t.Errorf("SumWarehouseLinks() = %d, want %d", got, want)
	}
	SortWarehouseLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseLinks() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseLinks(xs)) != 3 {
		t.Errorf("IndexWarehouseLinks() = %v", IndexWarehouseLinks(xs))
	}
	if err := ValidateWarehouseLinks(ActiveWarehouseLinks(xs)); err != nil {
		t.Errorf("ValidateWarehouseLinks() = %v", err)
	}
}

func TestWarehouseLinkNormalizeClamps(t *testing.T) {
	x := SampleWarehouseLink("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseLink("r1", when)
	rev, err := r.PutWarehouseLink(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseLink() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseLinks() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseLink(WarehouseLink{}, 0); err == nil {
		t.Error("PutWarehouseLink() accepted an empty entity")
	}
}

func TestWarehouseNoteRoundTrip(t *testing.T) {
	x := SampleWarehouseNote("warehousenote-1", when)
	got, err := ParseWarehouseNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseNote() = %v", err)
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

func TestWarehouseNoteParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseNote(f); err == nil {
		t.Error("ParseWarehouseNote() accepted a record with no id")
	}
	f = SampleWarehouseNote("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseWarehouseNote(f); err == nil {
		t.Error("ParseWarehouseNote() accepted a non-numeric weight")
	}
}

func TestWarehouseNoteCollections(t *testing.T) {
	xs := []WarehouseNote{SampleWarehouseNote("a", when), SampleWarehouseNote("bb", when), SampleWarehouseNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseNotes(xs)); got != 2 {
		t.Errorf("ActiveWarehouseNotes() kept %d, want 2", got)
	}
	if got, want := SumWarehouseNotes(xs), 1+3; got != want {
		t.Errorf("SumWarehouseNotes() = %d, want %d", got, want)
	}
	SortWarehouseNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseNotes() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseNotes(xs)) != 3 {
		t.Errorf("IndexWarehouseNotes() = %v", IndexWarehouseNotes(xs))
	}
	if err := ValidateWarehouseNotes(ActiveWarehouseNotes(xs)); err != nil {
		t.Errorf("ValidateWarehouseNotes() = %v", err)
	}
}

func TestWarehouseNoteNormalizeClamps(t *testing.T) {
	x := SampleWarehouseNote("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseNote("r1", when)
	rev, err := r.PutWarehouseNote(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseNote() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseNotes() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseNote(WarehouseNote{}, 0); err == nil {
		t.Error("PutWarehouseNote() accepted an empty entity")
	}
}

func TestWarehouseWindowRoundTrip(t *testing.T) {
	x := SampleWarehouseWindow("warehousewindow-1", when)
	got, err := ParseWarehouseWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseWindow() = %v", err)
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

func TestWarehouseWindowParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseWindow(f); err == nil {
		t.Error("ParseWarehouseWindow() accepted a record with no id")
	}
	f = SampleWarehouseWindow("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseWarehouseWindow(f); err == nil {
		t.Error("ParseWarehouseWindow() accepted a non-numeric attempts")
	}
}

func TestWarehouseWindowCollections(t *testing.T) {
	xs := []WarehouseWindow{SampleWarehouseWindow("a", when), SampleWarehouseWindow("bb", when), SampleWarehouseWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseWindows(xs)); got != 2 {
		t.Errorf("ActiveWarehouseWindows() kept %d, want 2", got)
	}
	if got, want := SumWarehouseWindows(xs), 1+3; got != want {
		t.Errorf("SumWarehouseWindows() = %d, want %d", got, want)
	}
	SortWarehouseWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseWindows() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseWindows(xs)) != 3 {
		t.Errorf("IndexWarehouseWindows() = %v", IndexWarehouseWindows(xs))
	}
	if err := ValidateWarehouseWindows(ActiveWarehouseWindows(xs)); err != nil {
		t.Errorf("ValidateWarehouseWindows() = %v", err)
	}
}

func TestWarehouseWindowNormalizeClamps(t *testing.T) {
	x := SampleWarehouseWindow("a", when)
	x.Reference = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseWindow("r1", when)
	rev, err := r.PutWarehouseWindow(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseWindow() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseWindows() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseWindow(WarehouseWindow{}, 0); err == nil {
		t.Error("PutWarehouseWindow() accepted an empty entity")
	}
}

func TestWarehouseQuotaRoundTrip(t *testing.T) {
	x := SampleWarehouseQuota("warehousequota-1", when)
	got, err := ParseWarehouseQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseQuota() = %v", err)
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

func TestWarehouseQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseQuota(f); err == nil {
		t.Error("ParseWarehouseQuota() accepted a record with no id")
	}
	f = SampleWarehouseQuota("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseWarehouseQuota(f); err == nil {
		t.Error("ParseWarehouseQuota() accepted a non-numeric count")
	}
}

func TestWarehouseQuotaCollections(t *testing.T) {
	xs := []WarehouseQuota{SampleWarehouseQuota("a", when), SampleWarehouseQuota("bb", when), SampleWarehouseQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseQuotas(xs)); got != 2 {
		t.Errorf("ActiveWarehouseQuotas() kept %d, want 2", got)
	}
	if got, want := SumWarehouseQuotas(xs), 1+3; got != want {
		t.Errorf("SumWarehouseQuotas() = %d, want %d", got, want)
	}
	SortWarehouseQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseQuotas() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseQuotas(xs)) != 3 {
		t.Errorf("IndexWarehouseQuotas() = %v", IndexWarehouseQuotas(xs))
	}
	if err := ValidateWarehouseQuotas(ActiveWarehouseQuotas(xs)); err != nil {
		t.Errorf("ValidateWarehouseQuotas() = %v", err)
	}
}

func TestWarehouseQuotaNormalizeClamps(t *testing.T) {
	x := SampleWarehouseQuota("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseQuota("r1", when)
	rev, err := r.PutWarehouseQuota(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseQuota() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseQuota(WarehouseQuota{}, 0); err == nil {
		t.Error("PutWarehouseQuota() accepted an empty entity")
	}
}

func TestWarehouseBucketRoundTrip(t *testing.T) {
	x := SampleWarehouseBucket("warehousebucket-1", when)
	got, err := ParseWarehouseBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseBucket() = %v", err)
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

func TestWarehouseBucketParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseBucket(f); err == nil {
		t.Error("ParseWarehouseBucket() accepted a record with no id")
	}
	f = SampleWarehouseBucket("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseWarehouseBucket(f); err == nil {
		t.Error("ParseWarehouseBucket() accepted a non-numeric capacity")
	}
}

func TestWarehouseBucketCollections(t *testing.T) {
	xs := []WarehouseBucket{SampleWarehouseBucket("a", when), SampleWarehouseBucket("bb", when), SampleWarehouseBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseBuckets(xs)); got != 2 {
		t.Errorf("ActiveWarehouseBuckets() kept %d, want 2", got)
	}
	if got, want := SumWarehouseBuckets(xs), 1+3; got != want {
		t.Errorf("SumWarehouseBuckets() = %d, want %d", got, want)
	}
	SortWarehouseBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseBuckets() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseBuckets(xs)) != 3 {
		t.Errorf("IndexWarehouseBuckets() = %v", IndexWarehouseBuckets(xs))
	}
	if err := ValidateWarehouseBuckets(ActiveWarehouseBuckets(xs)); err != nil {
		t.Errorf("ValidateWarehouseBuckets() = %v", err)
	}
}

func TestWarehouseBucketNormalizeClamps(t *testing.T) {
	x := SampleWarehouseBucket("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseBucket("r1", when)
	rev, err := r.PutWarehouseBucket(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseBucket() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseBucket(WarehouseBucket{}, 0); err == nil {
		t.Error("PutWarehouseBucket() accepted an empty entity")
	}
}

func TestWarehouseLaneRoundTrip(t *testing.T) {
	x := SampleWarehouseLane("warehouselane-1", when)
	got, err := ParseWarehouseLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseLane() = %v", err)
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

func TestWarehouseLaneParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseLane(f); err == nil {
		t.Error("ParseWarehouseLane() accepted a record with no id")
	}
	f = SampleWarehouseLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseWarehouseLane(f); err == nil {
		t.Error("ParseWarehouseLane() accepted a non-numeric attempts")
	}
}

func TestWarehouseLaneCollections(t *testing.T) {
	xs := []WarehouseLane{SampleWarehouseLane("a", when), SampleWarehouseLane("bb", when), SampleWarehouseLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseLanes(xs)); got != 2 {
		t.Errorf("ActiveWarehouseLanes() kept %d, want 2", got)
	}
	if got, want := SumWarehouseLanes(xs), 1+3; got != want {
		t.Errorf("SumWarehouseLanes() = %d, want %d", got, want)
	}
	SortWarehouseLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseLanes() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseLanes(xs)) != 3 {
		t.Errorf("IndexWarehouseLanes() = %v", IndexWarehouseLanes(xs))
	}
	if err := ValidateWarehouseLanes(ActiveWarehouseLanes(xs)); err != nil {
		t.Errorf("ValidateWarehouseLanes() = %v", err)
	}
}

func TestWarehouseLaneNormalizeClamps(t *testing.T) {
	x := SampleWarehouseLane("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseLane("r1", when)
	rev, err := r.PutWarehouseLane(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseLane() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseLanes() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseLane(WarehouseLane{}, 0); err == nil {
		t.Error("PutWarehouseLane() accepted an empty entity")
	}
}

func TestWarehouseSlotRoundTrip(t *testing.T) {
	x := SampleWarehouseSlot("warehouseslot-1", when)
	got, err := ParseWarehouseSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseSlot() = %v", err)
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

func TestWarehouseSlotParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseSlot(f); err == nil {
		t.Error("ParseWarehouseSlot() accepted a record with no id")
	}
	f = SampleWarehouseSlot("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseWarehouseSlot(f); err == nil {
		t.Error("ParseWarehouseSlot() accepted a non-numeric attempts")
	}
}

func TestWarehouseSlotCollections(t *testing.T) {
	xs := []WarehouseSlot{SampleWarehouseSlot("a", when), SampleWarehouseSlot("bb", when), SampleWarehouseSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseSlots(xs)); got != 2 {
		t.Errorf("ActiveWarehouseSlots() kept %d, want 2", got)
	}
	if got, want := SumWarehouseSlots(xs), 1+3; got != want {
		t.Errorf("SumWarehouseSlots() = %d, want %d", got, want)
	}
	SortWarehouseSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseSlots() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseSlots(xs)) != 3 {
		t.Errorf("IndexWarehouseSlots() = %v", IndexWarehouseSlots(xs))
	}
	if err := ValidateWarehouseSlots(ActiveWarehouseSlots(xs)); err != nil {
		t.Errorf("ValidateWarehouseSlots() = %v", err)
	}
}

func TestWarehouseSlotNormalizeClamps(t *testing.T) {
	x := SampleWarehouseSlot("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseSlot("r1", when)
	rev, err := r.PutWarehouseSlot(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseSlot() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseSlots() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseSlot(WarehouseSlot{}, 0); err == nil {
		t.Error("PutWarehouseSlot() accepted an empty entity")
	}
}

func TestWarehouseTagRoundTrip(t *testing.T) {
	x := SampleWarehouseTag("warehousetag-1", when)
	got, err := ParseWarehouseTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseTag() = %v", err)
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

func TestWarehouseTagParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseTag(f); err == nil {
		t.Error("ParseWarehouseTag() accepted a record with no id")
	}
	f = SampleWarehouseTag("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseWarehouseTag(f); err == nil {
		t.Error("ParseWarehouseTag() accepted a non-numeric version")
	}
}

func TestWarehouseTagCollections(t *testing.T) {
	xs := []WarehouseTag{SampleWarehouseTag("a", when), SampleWarehouseTag("bb", when), SampleWarehouseTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseTags(xs)); got != 2 {
		t.Errorf("ActiveWarehouseTags() kept %d, want 2", got)
	}
	if got, want := SumWarehouseTags(xs), 1+3; got != want {
		t.Errorf("SumWarehouseTags() = %d, want %d", got, want)
	}
	SortWarehouseTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseTags() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseTags(xs)) != 3 {
		t.Errorf("IndexWarehouseTags() = %v", IndexWarehouseTags(xs))
	}
	if err := ValidateWarehouseTags(ActiveWarehouseTags(xs)); err != nil {
		t.Errorf("ValidateWarehouseTags() = %v", err)
	}
}

func TestWarehouseTagNormalizeClamps(t *testing.T) {
	x := SampleWarehouseTag("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseTag("r1", when)
	rev, err := r.PutWarehouseTag(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseTag() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseTags() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseTag(WarehouseTag{}, 0); err == nil {
		t.Error("PutWarehouseTag() accepted an empty entity")
	}
}

func TestWarehouseRuleRoundTrip(t *testing.T) {
	x := SampleWarehouseRule("warehouserule-1", when)
	got, err := ParseWarehouseRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseRule() = %v", err)
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

func TestWarehouseRuleParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseRule(f); err == nil {
		t.Error("ParseWarehouseRule() accepted a record with no id")
	}
	f = SampleWarehouseRule("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWarehouseRule(f); err == nil {
		t.Error("ParseWarehouseRule() accepted a non-numeric priority")
	}
}

func TestWarehouseRuleCollections(t *testing.T) {
	xs := []WarehouseRule{SampleWarehouseRule("a", when), SampleWarehouseRule("bb", when), SampleWarehouseRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseRules(xs)); got != 2 {
		t.Errorf("ActiveWarehouseRules() kept %d, want 2", got)
	}
	if got, want := SumWarehouseRules(xs), 1+3; got != want {
		t.Errorf("SumWarehouseRules() = %d, want %d", got, want)
	}
	SortWarehouseRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseRules() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseRules(xs)) != 3 {
		t.Errorf("IndexWarehouseRules() = %v", IndexWarehouseRules(xs))
	}
	if err := ValidateWarehouseRules(ActiveWarehouseRules(xs)); err != nil {
		t.Errorf("ValidateWarehouseRules() = %v", err)
	}
}

func TestWarehouseRuleNormalizeClamps(t *testing.T) {
	x := SampleWarehouseRule("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseRule("r1", when)
	rev, err := r.PutWarehouseRule(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseRule() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseRules() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseRule(WarehouseRule{}, 0); err == nil {
		t.Error("PutWarehouseRule() accepted an empty entity")
	}
}

func TestWarehouseHoldRoundTrip(t *testing.T) {
	x := SampleWarehouseHold("warehousehold-1", when)
	got, err := ParseWarehouseHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseHold() = %v", err)
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

func TestWarehouseHoldParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseHold(f); err == nil {
		t.Error("ParseWarehouseHold() accepted a record with no id")
	}
	f = SampleWarehouseHold("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWarehouseHold(f); err == nil {
		t.Error("ParseWarehouseHold() accepted a non-numeric priority")
	}
}

func TestWarehouseHoldCollections(t *testing.T) {
	xs := []WarehouseHold{SampleWarehouseHold("a", when), SampleWarehouseHold("bb", when), SampleWarehouseHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseHolds(xs)); got != 2 {
		t.Errorf("ActiveWarehouseHolds() kept %d, want 2", got)
	}
	if got, want := SumWarehouseHolds(xs), 1+3; got != want {
		t.Errorf("SumWarehouseHolds() = %d, want %d", got, want)
	}
	SortWarehouseHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseHolds() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseHolds(xs)) != 3 {
		t.Errorf("IndexWarehouseHolds() = %v", IndexWarehouseHolds(xs))
	}
	if err := ValidateWarehouseHolds(ActiveWarehouseHolds(xs)); err != nil {
		t.Errorf("ValidateWarehouseHolds() = %v", err)
	}
}

func TestWarehouseHoldNormalizeClamps(t *testing.T) {
	x := SampleWarehouseHold("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseHold("r1", when)
	rev, err := r.PutWarehouseHold(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseHold() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseHolds() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseHold(WarehouseHold{}, 0); err == nil {
		t.Error("PutWarehouseHold() accepted an empty entity")
	}
}

func TestWarehouseClaimRoundTrip(t *testing.T) {
	x := SampleWarehouseClaim("warehouseclaim-1", when)
	got, err := ParseWarehouseClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseClaim() = %v", err)
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

func TestWarehouseClaimParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseClaim(f); err == nil {
		t.Error("ParseWarehouseClaim() accepted a record with no id")
	}
	f = SampleWarehouseClaim("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseWarehouseClaim(f); err == nil {
		t.Error("ParseWarehouseClaim() accepted a non-numeric offset")
	}
}

func TestWarehouseClaimCollections(t *testing.T) {
	xs := []WarehouseClaim{SampleWarehouseClaim("a", when), SampleWarehouseClaim("bb", when), SampleWarehouseClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseClaims(xs)); got != 2 {
		t.Errorf("ActiveWarehouseClaims() kept %d, want 2", got)
	}
	if got, want := SumWarehouseClaims(xs), 1+3; got != want {
		t.Errorf("SumWarehouseClaims() = %d, want %d", got, want)
	}
	SortWarehouseClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseClaims() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseClaims(xs)) != 3 {
		t.Errorf("IndexWarehouseClaims() = %v", IndexWarehouseClaims(xs))
	}
	if err := ValidateWarehouseClaims(ActiveWarehouseClaims(xs)); err != nil {
		t.Errorf("ValidateWarehouseClaims() = %v", err)
	}
}

func TestWarehouseClaimNormalizeClamps(t *testing.T) {
	x := SampleWarehouseClaim("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseClaim("r1", when)
	rev, err := r.PutWarehouseClaim(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseClaim() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseClaims() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseClaim(WarehouseClaim{}, 0); err == nil {
		t.Error("PutWarehouseClaim() accepted an empty entity")
	}
}

func TestWarehouseCursorRoundTrip(t *testing.T) {
	x := SampleWarehouseCursor("warehousecursor-1", when)
	got, err := ParseWarehouseCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseWarehouseCursor() = %v", err)
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

func TestWarehouseCursorParseRejectsBadInput(t *testing.T) {
	f := SampleWarehouseCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWarehouseCursor(f); err == nil {
		t.Error("ParseWarehouseCursor() accepted a record with no id")
	}
	f = SampleWarehouseCursor("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWarehouseCursor(f); err == nil {
		t.Error("ParseWarehouseCursor() accepted a non-numeric priority")
	}
}

func TestWarehouseCursorCollections(t *testing.T) {
	xs := []WarehouseCursor{SampleWarehouseCursor("a", when), SampleWarehouseCursor("bb", when), SampleWarehouseCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWarehouseCursors(xs)); got != 2 {
		t.Errorf("ActiveWarehouseCursors() kept %d, want 2", got)
	}
	if got, want := SumWarehouseCursors(xs), 1+3; got != want {
		t.Errorf("SumWarehouseCursors() = %d, want %d", got, want)
	}
	SortWarehouseCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWarehouseCursors() put %q first", xs[0].ID)
	}
	if len(IndexWarehouseCursors(xs)) != 3 {
		t.Errorf("IndexWarehouseCursors() = %v", IndexWarehouseCursors(xs))
	}
	if err := ValidateWarehouseCursors(ActiveWarehouseCursors(xs)); err != nil {
		t.Errorf("ValidateWarehouseCursors() = %v", err)
	}
}

func TestWarehouseCursorNormalizeClamps(t *testing.T) {
	x := SampleWarehouseCursor("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWarehouseCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWarehouseCursor("r1", when)
	rev, err := r.PutWarehouseCursor(x, 0)
	if err != nil {
		t.Fatalf("PutWarehouseCursor() = %v", err)
	}
	got, gotRev, err := r.GetWarehouseCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWarehouseCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWarehouseCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWarehouseCursors() = %v, %v", all, err)
	}
	if _, err := r.PutWarehouseCursor(WarehouseCursor{}, 0); err == nil {
		t.Error("PutWarehouseCursor() accepted an empty entity")
	}
}
