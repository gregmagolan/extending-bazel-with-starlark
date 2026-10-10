package inventory

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
		if !strings.HasPrefix(Prefixes[i], "inventory/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleInventoryRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

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

func TestInventoryRecordRoundTrip(t *testing.T) {
	x := SampleInventoryRecord("inventoryrecord-1", when)
	got, err := ParseInventoryRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryRecord() = %v", err)
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

func TestInventoryRecordParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryRecord(f); err == nil {
		t.Error("ParseInventoryRecord() accepted a record with no id")
	}
	f = SampleInventoryRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseInventoryRecord(f); err == nil {
		t.Error("ParseInventoryRecord() accepted a non-numeric capacity")
	}
}

func TestInventoryRecordCollections(t *testing.T) {
	xs := []InventoryRecord{SampleInventoryRecord("a", when), SampleInventoryRecord("bb", when), SampleInventoryRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryRecords(xs)); got != 2 {
		t.Errorf("ActiveInventoryRecords() kept %d, want 2", got)
	}
	if got, want := SumInventoryRecords(xs), 1+3; got != want {
		t.Errorf("SumInventoryRecords() = %d, want %d", got, want)
	}
	SortInventoryRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryRecords() put %q first", xs[0].ID)
	}
	if len(IndexInventoryRecords(xs)) != 3 {
		t.Errorf("IndexInventoryRecords() = %v", IndexInventoryRecords(xs))
	}
	if err := ValidateInventoryRecords(ActiveInventoryRecords(xs)); err != nil {
		t.Errorf("ValidateInventoryRecords() = %v", err)
	}
}

func TestInventoryRecordNormalizeClamps(t *testing.T) {
	x := SampleInventoryRecord("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryRecord("r1", when)
	rev, err := r.PutInventoryRecord(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryRecord() = %v", err)
	}
	got, gotRev, err := r.GetInventoryRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryRecords() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryRecord(InventoryRecord{}, 0); err == nil {
		t.Error("PutInventoryRecord() accepted an empty entity")
	}
}

func TestInventoryEntryRoundTrip(t *testing.T) {
	x := SampleInventoryEntry("inventoryentry-1", when)
	got, err := ParseInventoryEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryEntry() = %v", err)
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

func TestInventoryEntryParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryEntry(f); err == nil {
		t.Error("ParseInventoryEntry() accepted a record with no id")
	}
	f = SampleInventoryEntry("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInventoryEntry(f); err == nil {
		t.Error("ParseInventoryEntry() accepted a non-numeric priority")
	}
}

func TestInventoryEntryCollections(t *testing.T) {
	xs := []InventoryEntry{SampleInventoryEntry("a", when), SampleInventoryEntry("bb", when), SampleInventoryEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryEntrys(xs)); got != 2 {
		t.Errorf("ActiveInventoryEntrys() kept %d, want 2", got)
	}
	if got, want := SumInventoryEntrys(xs), 1+3; got != want {
		t.Errorf("SumInventoryEntrys() = %d, want %d", got, want)
	}
	SortInventoryEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryEntrys() put %q first", xs[0].ID)
	}
	if len(IndexInventoryEntrys(xs)) != 3 {
		t.Errorf("IndexInventoryEntrys() = %v", IndexInventoryEntrys(xs))
	}
	if err := ValidateInventoryEntrys(ActiveInventoryEntrys(xs)); err != nil {
		t.Errorf("ValidateInventoryEntrys() = %v", err)
	}
}

func TestInventoryEntryNormalizeClamps(t *testing.T) {
	x := SampleInventoryEntry("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryEntry("r1", when)
	rev, err := r.PutInventoryEntry(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryEntry() = %v", err)
	}
	got, gotRev, err := r.GetInventoryEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryEntry(InventoryEntry{}, 0); err == nil {
		t.Error("PutInventoryEntry() accepted an empty entity")
	}
}

func TestInventoryItemRoundTrip(t *testing.T) {
	x := SampleInventoryItem("inventoryitem-1", when)
	got, err := ParseInventoryItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryItem() = %v", err)
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

func TestInventoryItemParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryItem(f); err == nil {
		t.Error("ParseInventoryItem() accepted a record with no id")
	}
	f = SampleInventoryItem("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseInventoryItem(f); err == nil {
		t.Error("ParseInventoryItem() accepted a non-numeric weight")
	}
}

func TestInventoryItemCollections(t *testing.T) {
	xs := []InventoryItem{SampleInventoryItem("a", when), SampleInventoryItem("bb", when), SampleInventoryItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryItems(xs)); got != 2 {
		t.Errorf("ActiveInventoryItems() kept %d, want 2", got)
	}
	if got, want := SumInventoryItems(xs), 1+3; got != want {
		t.Errorf("SumInventoryItems() = %d, want %d", got, want)
	}
	SortInventoryItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryItems() put %q first", xs[0].ID)
	}
	if len(IndexInventoryItems(xs)) != 3 {
		t.Errorf("IndexInventoryItems() = %v", IndexInventoryItems(xs))
	}
	if err := ValidateInventoryItems(ActiveInventoryItems(xs)); err != nil {
		t.Errorf("ValidateInventoryItems() = %v", err)
	}
}

func TestInventoryItemNormalizeClamps(t *testing.T) {
	x := SampleInventoryItem("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryItem("r1", when)
	rev, err := r.PutInventoryItem(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryItem() = %v", err)
	}
	got, gotRev, err := r.GetInventoryItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryItems() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryItem(InventoryItem{}, 0); err == nil {
		t.Error("PutInventoryItem() accepted an empty entity")
	}
}

func TestInventoryDraftRoundTrip(t *testing.T) {
	x := SampleInventoryDraft("inventorydraft-1", when)
	got, err := ParseInventoryDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryDraft() = %v", err)
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

func TestInventoryDraftParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryDraft(f); err == nil {
		t.Error("ParseInventoryDraft() accepted a record with no id")
	}
	f = SampleInventoryDraft("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseInventoryDraft(f); err == nil {
		t.Error("ParseInventoryDraft() accepted a non-numeric offset")
	}
}

func TestInventoryDraftCollections(t *testing.T) {
	xs := []InventoryDraft{SampleInventoryDraft("a", when), SampleInventoryDraft("bb", when), SampleInventoryDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryDrafts(xs)); got != 2 {
		t.Errorf("ActiveInventoryDrafts() kept %d, want 2", got)
	}
	if got, want := SumInventoryDrafts(xs), 1+3; got != want {
		t.Errorf("SumInventoryDrafts() = %d, want %d", got, want)
	}
	SortInventoryDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryDrafts() put %q first", xs[0].ID)
	}
	if len(IndexInventoryDrafts(xs)) != 3 {
		t.Errorf("IndexInventoryDrafts() = %v", IndexInventoryDrafts(xs))
	}
	if err := ValidateInventoryDrafts(ActiveInventoryDrafts(xs)); err != nil {
		t.Errorf("ValidateInventoryDrafts() = %v", err)
	}
}

func TestInventoryDraftNormalizeClamps(t *testing.T) {
	x := SampleInventoryDraft("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryDraft("r1", when)
	rev, err := r.PutInventoryDraft(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryDraft() = %v", err)
	}
	got, gotRev, err := r.GetInventoryDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryDraft(InventoryDraft{}, 0); err == nil {
		t.Error("PutInventoryDraft() accepted an empty entity")
	}
}

func TestInventoryRevisionRoundTrip(t *testing.T) {
	x := SampleInventoryRevision("inventoryrevision-1", when)
	got, err := ParseInventoryRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryRevision() = %v", err)
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

func TestInventoryRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryRevision(f); err == nil {
		t.Error("ParseInventoryRevision() accepted a record with no id")
	}
	f = SampleInventoryRevision("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseInventoryRevision(f); err == nil {
		t.Error("ParseInventoryRevision() accepted a non-numeric offset")
	}
}

func TestInventoryRevisionCollections(t *testing.T) {
	xs := []InventoryRevision{SampleInventoryRevision("a", when), SampleInventoryRevision("bb", when), SampleInventoryRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryRevisions(xs)); got != 2 {
		t.Errorf("ActiveInventoryRevisions() kept %d, want 2", got)
	}
	if got, want := SumInventoryRevisions(xs), 1+3; got != want {
		t.Errorf("SumInventoryRevisions() = %d, want %d", got, want)
	}
	SortInventoryRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryRevisions() put %q first", xs[0].ID)
	}
	if len(IndexInventoryRevisions(xs)) != 3 {
		t.Errorf("IndexInventoryRevisions() = %v", IndexInventoryRevisions(xs))
	}
	if err := ValidateInventoryRevisions(ActiveInventoryRevisions(xs)); err != nil {
		t.Errorf("ValidateInventoryRevisions() = %v", err)
	}
}

func TestInventoryRevisionNormalizeClamps(t *testing.T) {
	x := SampleInventoryRevision("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryRevision("r1", when)
	rev, err := r.PutInventoryRevision(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryRevision() = %v", err)
	}
	got, gotRev, err := r.GetInventoryRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryRevision(InventoryRevision{}, 0); err == nil {
		t.Error("PutInventoryRevision() accepted an empty entity")
	}
}

func TestInventorySnapshotRoundTrip(t *testing.T) {
	x := SampleInventorySnapshot("inventorysnapshot-1", when)
	got, err := ParseInventorySnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventorySnapshot() = %v", err)
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

func TestInventorySnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleInventorySnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventorySnapshot(f); err == nil {
		t.Error("ParseInventorySnapshot() accepted a record with no id")
	}
	f = SampleInventorySnapshot("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseInventorySnapshot(f); err == nil {
		t.Error("ParseInventorySnapshot() accepted a non-numeric attempts")
	}
}

func TestInventorySnapshotCollections(t *testing.T) {
	xs := []InventorySnapshot{SampleInventorySnapshot("a", when), SampleInventorySnapshot("bb", when), SampleInventorySnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventorySnapshots(xs)); got != 2 {
		t.Errorf("ActiveInventorySnapshots() kept %d, want 2", got)
	}
	if got, want := SumInventorySnapshots(xs), 1+3; got != want {
		t.Errorf("SumInventorySnapshots() = %d, want %d", got, want)
	}
	SortInventorySnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventorySnapshots() put %q first", xs[0].ID)
	}
	if len(IndexInventorySnapshots(xs)) != 3 {
		t.Errorf("IndexInventorySnapshots() = %v", IndexInventorySnapshots(xs))
	}
	if err := ValidateInventorySnapshots(ActiveInventorySnapshots(xs)); err != nil {
		t.Errorf("ValidateInventorySnapshots() = %v", err)
	}
}

func TestInventorySnapshotNormalizeClamps(t *testing.T) {
	x := SampleInventorySnapshot("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventorySnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventorySnapshot("r1", when)
	rev, err := r.PutInventorySnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutInventorySnapshot() = %v", err)
	}
	got, gotRev, err := r.GetInventorySnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventorySnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventorySnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventorySnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutInventorySnapshot(InventorySnapshot{}, 0); err == nil {
		t.Error("PutInventorySnapshot() accepted an empty entity")
	}
}

func TestInventoryBatchRoundTrip(t *testing.T) {
	x := SampleInventoryBatch("inventorybatch-1", when)
	got, err := ParseInventoryBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryBatch() = %v", err)
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

func TestInventoryBatchParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryBatch(f); err == nil {
		t.Error("ParseInventoryBatch() accepted a record with no id")
	}
	f = SampleInventoryBatch("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseInventoryBatch(f); err == nil {
		t.Error("ParseInventoryBatch() accepted a non-numeric attempts")
	}
}

func TestInventoryBatchCollections(t *testing.T) {
	xs := []InventoryBatch{SampleInventoryBatch("a", when), SampleInventoryBatch("bb", when), SampleInventoryBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryBatchs(xs)); got != 2 {
		t.Errorf("ActiveInventoryBatchs() kept %d, want 2", got)
	}
	if got, want := SumInventoryBatchs(xs), 1+3; got != want {
		t.Errorf("SumInventoryBatchs() = %d, want %d", got, want)
	}
	SortInventoryBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryBatchs() put %q first", xs[0].ID)
	}
	if len(IndexInventoryBatchs(xs)) != 3 {
		t.Errorf("IndexInventoryBatchs() = %v", IndexInventoryBatchs(xs))
	}
	if err := ValidateInventoryBatchs(ActiveInventoryBatchs(xs)); err != nil {
		t.Errorf("ValidateInventoryBatchs() = %v", err)
	}
}

func TestInventoryBatchNormalizeClamps(t *testing.T) {
	x := SampleInventoryBatch("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryBatch("r1", when)
	rev, err := r.PutInventoryBatch(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryBatch() = %v", err)
	}
	got, gotRev, err := r.GetInventoryBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryBatch(InventoryBatch{}, 0); err == nil {
		t.Error("PutInventoryBatch() accepted an empty entity")
	}
}

func TestInventoryRequestRoundTrip(t *testing.T) {
	x := SampleInventoryRequest("inventoryrequest-1", when)
	got, err := ParseInventoryRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryRequest() = %v", err)
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

func TestInventoryRequestParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryRequest(f); err == nil {
		t.Error("ParseInventoryRequest() accepted a record with no id")
	}
	f = SampleInventoryRequest("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseInventoryRequest(f); err == nil {
		t.Error("ParseInventoryRequest() accepted a non-numeric count")
	}
}

func TestInventoryRequestCollections(t *testing.T) {
	xs := []InventoryRequest{SampleInventoryRequest("a", when), SampleInventoryRequest("bb", when), SampleInventoryRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryRequests(xs)); got != 2 {
		t.Errorf("ActiveInventoryRequests() kept %d, want 2", got)
	}
	if got, want := SumInventoryRequests(xs), 1+3; got != want {
		t.Errorf("SumInventoryRequests() = %d, want %d", got, want)
	}
	SortInventoryRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryRequests() put %q first", xs[0].ID)
	}
	if len(IndexInventoryRequests(xs)) != 3 {
		t.Errorf("IndexInventoryRequests() = %v", IndexInventoryRequests(xs))
	}
	if err := ValidateInventoryRequests(ActiveInventoryRequests(xs)); err != nil {
		t.Errorf("ValidateInventoryRequests() = %v", err)
	}
}

func TestInventoryRequestNormalizeClamps(t *testing.T) {
	x := SampleInventoryRequest("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryRequest("r1", when)
	rev, err := r.PutInventoryRequest(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryRequest() = %v", err)
	}
	got, gotRev, err := r.GetInventoryRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryRequests() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryRequest(InventoryRequest{}, 0); err == nil {
		t.Error("PutInventoryRequest() accepted an empty entity")
	}
}

func TestInventorySummaryRoundTrip(t *testing.T) {
	x := SampleInventorySummary("inventorysummary-1", when)
	got, err := ParseInventorySummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventorySummary() = %v", err)
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

func TestInventorySummaryParseRejectsBadInput(t *testing.T) {
	f := SampleInventorySummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventorySummary(f); err == nil {
		t.Error("ParseInventorySummary() accepted a record with no id")
	}
	f = SampleInventorySummary("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseInventorySummary(f); err == nil {
		t.Error("ParseInventorySummary() accepted a non-numeric attempts")
	}
}

func TestInventorySummaryCollections(t *testing.T) {
	xs := []InventorySummary{SampleInventorySummary("a", when), SampleInventorySummary("bb", when), SampleInventorySummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventorySummarys(xs)); got != 2 {
		t.Errorf("ActiveInventorySummarys() kept %d, want 2", got)
	}
	if got, want := SumInventorySummarys(xs), 1+3; got != want {
		t.Errorf("SumInventorySummarys() = %d, want %d", got, want)
	}
	SortInventorySummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventorySummarys() put %q first", xs[0].ID)
	}
	if len(IndexInventorySummarys(xs)) != 3 {
		t.Errorf("IndexInventorySummarys() = %v", IndexInventorySummarys(xs))
	}
	if err := ValidateInventorySummarys(ActiveInventorySummarys(xs)); err != nil {
		t.Errorf("ValidateInventorySummarys() = %v", err)
	}
}

func TestInventorySummaryNormalizeClamps(t *testing.T) {
	x := SampleInventorySummary("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventorySummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventorySummary("r1", when)
	rev, err := r.PutInventorySummary(x, 0)
	if err != nil {
		t.Fatalf("PutInventorySummary() = %v", err)
	}
	got, gotRev, err := r.GetInventorySummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventorySummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventorySummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventorySummarys() = %v, %v", all, err)
	}
	if _, err := r.PutInventorySummary(InventorySummary{}, 0); err == nil {
		t.Error("PutInventorySummary() accepted an empty entity")
	}
}

func TestInventoryPolicyRoundTrip(t *testing.T) {
	x := SampleInventoryPolicy("inventorypolicy-1", when)
	got, err := ParseInventoryPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryPolicy() = %v", err)
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

func TestInventoryPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryPolicy(f); err == nil {
		t.Error("ParseInventoryPolicy() accepted a record with no id")
	}
	f = SampleInventoryPolicy("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseInventoryPolicy(f); err == nil {
		t.Error("ParseInventoryPolicy() accepted a non-numeric weight")
	}
}

func TestInventoryPolicyCollections(t *testing.T) {
	xs := []InventoryPolicy{SampleInventoryPolicy("a", when), SampleInventoryPolicy("bb", when), SampleInventoryPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryPolicys(xs)); got != 2 {
		t.Errorf("ActiveInventoryPolicys() kept %d, want 2", got)
	}
	if got, want := SumInventoryPolicys(xs), 1+3; got != want {
		t.Errorf("SumInventoryPolicys() = %d, want %d", got, want)
	}
	SortInventoryPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryPolicys() put %q first", xs[0].ID)
	}
	if len(IndexInventoryPolicys(xs)) != 3 {
		t.Errorf("IndexInventoryPolicys() = %v", IndexInventoryPolicys(xs))
	}
	if err := ValidateInventoryPolicys(ActiveInventoryPolicys(xs)); err != nil {
		t.Errorf("ValidateInventoryPolicys() = %v", err)
	}
}

func TestInventoryPolicyNormalizeClamps(t *testing.T) {
	x := SampleInventoryPolicy("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryPolicy("r1", when)
	rev, err := r.PutInventoryPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryPolicy() = %v", err)
	}
	got, gotRev, err := r.GetInventoryPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryPolicy(InventoryPolicy{}, 0); err == nil {
		t.Error("PutInventoryPolicy() accepted an empty entity")
	}
}

func TestInventoryLinkRoundTrip(t *testing.T) {
	x := SampleInventoryLink("inventorylink-1", when)
	got, err := ParseInventoryLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryLink() = %v", err)
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

func TestInventoryLinkParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryLink(f); err == nil {
		t.Error("ParseInventoryLink() accepted a record with no id")
	}
	f = SampleInventoryLink("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseInventoryLink(f); err == nil {
		t.Error("ParseInventoryLink() accepted a non-numeric version")
	}
}

func TestInventoryLinkCollections(t *testing.T) {
	xs := []InventoryLink{SampleInventoryLink("a", when), SampleInventoryLink("bb", when), SampleInventoryLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryLinks(xs)); got != 2 {
		t.Errorf("ActiveInventoryLinks() kept %d, want 2", got)
	}
	if got, want := SumInventoryLinks(xs), 1+3; got != want {
		t.Errorf("SumInventoryLinks() = %d, want %d", got, want)
	}
	SortInventoryLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryLinks() put %q first", xs[0].ID)
	}
	if len(IndexInventoryLinks(xs)) != 3 {
		t.Errorf("IndexInventoryLinks() = %v", IndexInventoryLinks(xs))
	}
	if err := ValidateInventoryLinks(ActiveInventoryLinks(xs)); err != nil {
		t.Errorf("ValidateInventoryLinks() = %v", err)
	}
}

func TestInventoryLinkNormalizeClamps(t *testing.T) {
	x := SampleInventoryLink("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryLink("r1", when)
	rev, err := r.PutInventoryLink(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryLink() = %v", err)
	}
	got, gotRev, err := r.GetInventoryLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryLinks() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryLink(InventoryLink{}, 0); err == nil {
		t.Error("PutInventoryLink() accepted an empty entity")
	}
}

func TestInventoryNoteRoundTrip(t *testing.T) {
	x := SampleInventoryNote("inventorynote-1", when)
	got, err := ParseInventoryNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryNote() = %v", err)
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

func TestInventoryNoteParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryNote(f); err == nil {
		t.Error("ParseInventoryNote() accepted a record with no id")
	}
	f = SampleInventoryNote("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseInventoryNote(f); err == nil {
		t.Error("ParseInventoryNote() accepted a non-numeric version")
	}
}

func TestInventoryNoteCollections(t *testing.T) {
	xs := []InventoryNote{SampleInventoryNote("a", when), SampleInventoryNote("bb", when), SampleInventoryNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryNotes(xs)); got != 2 {
		t.Errorf("ActiveInventoryNotes() kept %d, want 2", got)
	}
	if got, want := SumInventoryNotes(xs), 1+3; got != want {
		t.Errorf("SumInventoryNotes() = %d, want %d", got, want)
	}
	SortInventoryNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryNotes() put %q first", xs[0].ID)
	}
	if len(IndexInventoryNotes(xs)) != 3 {
		t.Errorf("IndexInventoryNotes() = %v", IndexInventoryNotes(xs))
	}
	if err := ValidateInventoryNotes(ActiveInventoryNotes(xs)); err != nil {
		t.Errorf("ValidateInventoryNotes() = %v", err)
	}
}

func TestInventoryNoteNormalizeClamps(t *testing.T) {
	x := SampleInventoryNote("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryNote("r1", when)
	rev, err := r.PutInventoryNote(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryNote() = %v", err)
	}
	got, gotRev, err := r.GetInventoryNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryNotes() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryNote(InventoryNote{}, 0); err == nil {
		t.Error("PutInventoryNote() accepted an empty entity")
	}
}

func TestInventoryWindowRoundTrip(t *testing.T) {
	x := SampleInventoryWindow("inventorywindow-1", when)
	got, err := ParseInventoryWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryWindow() = %v", err)
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

func TestInventoryWindowParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryWindow(f); err == nil {
		t.Error("ParseInventoryWindow() accepted a record with no id")
	}
	f = SampleInventoryWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInventoryWindow(f); err == nil {
		t.Error("ParseInventoryWindow() accepted a non-numeric priority")
	}
}

func TestInventoryWindowCollections(t *testing.T) {
	xs := []InventoryWindow{SampleInventoryWindow("a", when), SampleInventoryWindow("bb", when), SampleInventoryWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryWindows(xs)); got != 2 {
		t.Errorf("ActiveInventoryWindows() kept %d, want 2", got)
	}
	if got, want := SumInventoryWindows(xs), 1+3; got != want {
		t.Errorf("SumInventoryWindows() = %d, want %d", got, want)
	}
	SortInventoryWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryWindows() put %q first", xs[0].ID)
	}
	if len(IndexInventoryWindows(xs)) != 3 {
		t.Errorf("IndexInventoryWindows() = %v", IndexInventoryWindows(xs))
	}
	if err := ValidateInventoryWindows(ActiveInventoryWindows(xs)); err != nil {
		t.Errorf("ValidateInventoryWindows() = %v", err)
	}
}

func TestInventoryWindowNormalizeClamps(t *testing.T) {
	x := SampleInventoryWindow("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryWindow("r1", when)
	rev, err := r.PutInventoryWindow(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryWindow() = %v", err)
	}
	got, gotRev, err := r.GetInventoryWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryWindows() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryWindow(InventoryWindow{}, 0); err == nil {
		t.Error("PutInventoryWindow() accepted an empty entity")
	}
}

func TestInventoryQuotaRoundTrip(t *testing.T) {
	x := SampleInventoryQuota("inventoryquota-1", when)
	got, err := ParseInventoryQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryQuota() = %v", err)
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

func TestInventoryQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryQuota(f); err == nil {
		t.Error("ParseInventoryQuota() accepted a record with no id")
	}
	f = SampleInventoryQuota("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseInventoryQuota(f); err == nil {
		t.Error("ParseInventoryQuota() accepted a non-numeric count")
	}
}

func TestInventoryQuotaCollections(t *testing.T) {
	xs := []InventoryQuota{SampleInventoryQuota("a", when), SampleInventoryQuota("bb", when), SampleInventoryQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryQuotas(xs)); got != 2 {
		t.Errorf("ActiveInventoryQuotas() kept %d, want 2", got)
	}
	if got, want := SumInventoryQuotas(xs), 1+3; got != want {
		t.Errorf("SumInventoryQuotas() = %d, want %d", got, want)
	}
	SortInventoryQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryQuotas() put %q first", xs[0].ID)
	}
	if len(IndexInventoryQuotas(xs)) != 3 {
		t.Errorf("IndexInventoryQuotas() = %v", IndexInventoryQuotas(xs))
	}
	if err := ValidateInventoryQuotas(ActiveInventoryQuotas(xs)); err != nil {
		t.Errorf("ValidateInventoryQuotas() = %v", err)
	}
}

func TestInventoryQuotaNormalizeClamps(t *testing.T) {
	x := SampleInventoryQuota("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryQuota("r1", when)
	rev, err := r.PutInventoryQuota(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryQuota() = %v", err)
	}
	got, gotRev, err := r.GetInventoryQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryQuota(InventoryQuota{}, 0); err == nil {
		t.Error("PutInventoryQuota() accepted an empty entity")
	}
}

func TestInventoryBucketRoundTrip(t *testing.T) {
	x := SampleInventoryBucket("inventorybucket-1", when)
	got, err := ParseInventoryBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryBucket() = %v", err)
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

func TestInventoryBucketParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryBucket(f); err == nil {
		t.Error("ParseInventoryBucket() accepted a record with no id")
	}
	f = SampleInventoryBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseInventoryBucket(f); err == nil {
		t.Error("ParseInventoryBucket() accepted a non-numeric retained")
	}
}

func TestInventoryBucketCollections(t *testing.T) {
	xs := []InventoryBucket{SampleInventoryBucket("a", when), SampleInventoryBucket("bb", when), SampleInventoryBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryBuckets(xs)); got != 2 {
		t.Errorf("ActiveInventoryBuckets() kept %d, want 2", got)
	}
	if got, want := SumInventoryBuckets(xs), 1+3; got != want {
		t.Errorf("SumInventoryBuckets() = %d, want %d", got, want)
	}
	SortInventoryBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryBuckets() put %q first", xs[0].ID)
	}
	if len(IndexInventoryBuckets(xs)) != 3 {
		t.Errorf("IndexInventoryBuckets() = %v", IndexInventoryBuckets(xs))
	}
	if err := ValidateInventoryBuckets(ActiveInventoryBuckets(xs)); err != nil {
		t.Errorf("ValidateInventoryBuckets() = %v", err)
	}
}

func TestInventoryBucketNormalizeClamps(t *testing.T) {
	x := SampleInventoryBucket("a", when)
	x.Reference = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryBucket("r1", when)
	rev, err := r.PutInventoryBucket(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryBucket() = %v", err)
	}
	got, gotRev, err := r.GetInventoryBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryBucket(InventoryBucket{}, 0); err == nil {
		t.Error("PutInventoryBucket() accepted an empty entity")
	}
}

func TestInventoryLaneRoundTrip(t *testing.T) {
	x := SampleInventoryLane("inventorylane-1", when)
	got, err := ParseInventoryLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryLane() = %v", err)
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

func TestInventoryLaneParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryLane(f); err == nil {
		t.Error("ParseInventoryLane() accepted a record with no id")
	}
	f = SampleInventoryLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseInventoryLane(f); err == nil {
		t.Error("ParseInventoryLane() accepted a non-numeric capacity")
	}
}

func TestInventoryLaneCollections(t *testing.T) {
	xs := []InventoryLane{SampleInventoryLane("a", when), SampleInventoryLane("bb", when), SampleInventoryLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryLanes(xs)); got != 2 {
		t.Errorf("ActiveInventoryLanes() kept %d, want 2", got)
	}
	if got, want := SumInventoryLanes(xs), 1+3; got != want {
		t.Errorf("SumInventoryLanes() = %d, want %d", got, want)
	}
	SortInventoryLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryLanes() put %q first", xs[0].ID)
	}
	if len(IndexInventoryLanes(xs)) != 3 {
		t.Errorf("IndexInventoryLanes() = %v", IndexInventoryLanes(xs))
	}
	if err := ValidateInventoryLanes(ActiveInventoryLanes(xs)); err != nil {
		t.Errorf("ValidateInventoryLanes() = %v", err)
	}
}

func TestInventoryLaneNormalizeClamps(t *testing.T) {
	x := SampleInventoryLane("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryLane("r1", when)
	rev, err := r.PutInventoryLane(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryLane() = %v", err)
	}
	got, gotRev, err := r.GetInventoryLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryLanes() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryLane(InventoryLane{}, 0); err == nil {
		t.Error("PutInventoryLane() accepted an empty entity")
	}
}

func TestInventorySlotRoundTrip(t *testing.T) {
	x := SampleInventorySlot("inventoryslot-1", when)
	got, err := ParseInventorySlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventorySlot() = %v", err)
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

func TestInventorySlotParseRejectsBadInput(t *testing.T) {
	f := SampleInventorySlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventorySlot(f); err == nil {
		t.Error("ParseInventorySlot() accepted a record with no id")
	}
	f = SampleInventorySlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseInventorySlot(f); err == nil {
		t.Error("ParseInventorySlot() accepted a non-numeric count")
	}
}

func TestInventorySlotCollections(t *testing.T) {
	xs := []InventorySlot{SampleInventorySlot("a", when), SampleInventorySlot("bb", when), SampleInventorySlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventorySlots(xs)); got != 2 {
		t.Errorf("ActiveInventorySlots() kept %d, want 2", got)
	}
	if got, want := SumInventorySlots(xs), 1+3; got != want {
		t.Errorf("SumInventorySlots() = %d, want %d", got, want)
	}
	SortInventorySlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventorySlots() put %q first", xs[0].ID)
	}
	if len(IndexInventorySlots(xs)) != 3 {
		t.Errorf("IndexInventorySlots() = %v", IndexInventorySlots(xs))
	}
	if err := ValidateInventorySlots(ActiveInventorySlots(xs)); err != nil {
		t.Errorf("ValidateInventorySlots() = %v", err)
	}
}

func TestInventorySlotNormalizeClamps(t *testing.T) {
	x := SampleInventorySlot("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventorySlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventorySlot("r1", when)
	rev, err := r.PutInventorySlot(x, 0)
	if err != nil {
		t.Fatalf("PutInventorySlot() = %v", err)
	}
	got, gotRev, err := r.GetInventorySlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventorySlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventorySlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventorySlots() = %v, %v", all, err)
	}
	if _, err := r.PutInventorySlot(InventorySlot{}, 0); err == nil {
		t.Error("PutInventorySlot() accepted an empty entity")
	}
}

func TestInventoryTagRoundTrip(t *testing.T) {
	x := SampleInventoryTag("inventorytag-1", when)
	got, err := ParseInventoryTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryTag() = %v", err)
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

func TestInventoryTagParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryTag(f); err == nil {
		t.Error("ParseInventoryTag() accepted a record with no id")
	}
	f = SampleInventoryTag("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInventoryTag(f); err == nil {
		t.Error("ParseInventoryTag() accepted a non-numeric priority")
	}
}

func TestInventoryTagCollections(t *testing.T) {
	xs := []InventoryTag{SampleInventoryTag("a", when), SampleInventoryTag("bb", when), SampleInventoryTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryTags(xs)); got != 2 {
		t.Errorf("ActiveInventoryTags() kept %d, want 2", got)
	}
	if got, want := SumInventoryTags(xs), 1+3; got != want {
		t.Errorf("SumInventoryTags() = %d, want %d", got, want)
	}
	SortInventoryTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryTags() put %q first", xs[0].ID)
	}
	if len(IndexInventoryTags(xs)) != 3 {
		t.Errorf("IndexInventoryTags() = %v", IndexInventoryTags(xs))
	}
	if err := ValidateInventoryTags(ActiveInventoryTags(xs)); err != nil {
		t.Errorf("ValidateInventoryTags() = %v", err)
	}
}

func TestInventoryTagNormalizeClamps(t *testing.T) {
	x := SampleInventoryTag("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryTag("r1", when)
	rev, err := r.PutInventoryTag(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryTag() = %v", err)
	}
	got, gotRev, err := r.GetInventoryTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryTags() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryTag(InventoryTag{}, 0); err == nil {
		t.Error("PutInventoryTag() accepted an empty entity")
	}
}

func TestInventoryRuleRoundTrip(t *testing.T) {
	x := SampleInventoryRule("inventoryrule-1", when)
	got, err := ParseInventoryRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryRule() = %v", err)
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

func TestInventoryRuleParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryRule(f); err == nil {
		t.Error("ParseInventoryRule() accepted a record with no id")
	}
	f = SampleInventoryRule("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInventoryRule(f); err == nil {
		t.Error("ParseInventoryRule() accepted a non-numeric priority")
	}
}

func TestInventoryRuleCollections(t *testing.T) {
	xs := []InventoryRule{SampleInventoryRule("a", when), SampleInventoryRule("bb", when), SampleInventoryRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryRules(xs)); got != 2 {
		t.Errorf("ActiveInventoryRules() kept %d, want 2", got)
	}
	if got, want := SumInventoryRules(xs), 1+3; got != want {
		t.Errorf("SumInventoryRules() = %d, want %d", got, want)
	}
	SortInventoryRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryRules() put %q first", xs[0].ID)
	}
	if len(IndexInventoryRules(xs)) != 3 {
		t.Errorf("IndexInventoryRules() = %v", IndexInventoryRules(xs))
	}
	if err := ValidateInventoryRules(ActiveInventoryRules(xs)); err != nil {
		t.Errorf("ValidateInventoryRules() = %v", err)
	}
}

func TestInventoryRuleNormalizeClamps(t *testing.T) {
	x := SampleInventoryRule("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryRule("r1", when)
	rev, err := r.PutInventoryRule(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryRule() = %v", err)
	}
	got, gotRev, err := r.GetInventoryRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryRules() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryRule(InventoryRule{}, 0); err == nil {
		t.Error("PutInventoryRule() accepted an empty entity")
	}
}

func TestInventoryHoldRoundTrip(t *testing.T) {
	x := SampleInventoryHold("inventoryhold-1", when)
	got, err := ParseInventoryHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryHold() = %v", err)
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

func TestInventoryHoldParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryHold(f); err == nil {
		t.Error("ParseInventoryHold() accepted a record with no id")
	}
	f = SampleInventoryHold("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseInventoryHold(f); err == nil {
		t.Error("ParseInventoryHold() accepted a non-numeric count")
	}
}

func TestInventoryHoldCollections(t *testing.T) {
	xs := []InventoryHold{SampleInventoryHold("a", when), SampleInventoryHold("bb", when), SampleInventoryHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryHolds(xs)); got != 2 {
		t.Errorf("ActiveInventoryHolds() kept %d, want 2", got)
	}
	if got, want := SumInventoryHolds(xs), 1+3; got != want {
		t.Errorf("SumInventoryHolds() = %d, want %d", got, want)
	}
	SortInventoryHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryHolds() put %q first", xs[0].ID)
	}
	if len(IndexInventoryHolds(xs)) != 3 {
		t.Errorf("IndexInventoryHolds() = %v", IndexInventoryHolds(xs))
	}
	if err := ValidateInventoryHolds(ActiveInventoryHolds(xs)); err != nil {
		t.Errorf("ValidateInventoryHolds() = %v", err)
	}
}

func TestInventoryHoldNormalizeClamps(t *testing.T) {
	x := SampleInventoryHold("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryHold("r1", when)
	rev, err := r.PutInventoryHold(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryHold() = %v", err)
	}
	got, gotRev, err := r.GetInventoryHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryHolds() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryHold(InventoryHold{}, 0); err == nil {
		t.Error("PutInventoryHold() accepted an empty entity")
	}
}

func TestInventoryClaimRoundTrip(t *testing.T) {
	x := SampleInventoryClaim("inventoryclaim-1", when)
	got, err := ParseInventoryClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryClaim() = %v", err)
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

func TestInventoryClaimParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryClaim(f); err == nil {
		t.Error("ParseInventoryClaim() accepted a record with no id")
	}
	f = SampleInventoryClaim("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseInventoryClaim(f); err == nil {
		t.Error("ParseInventoryClaim() accepted a non-numeric version")
	}
}

func TestInventoryClaimCollections(t *testing.T) {
	xs := []InventoryClaim{SampleInventoryClaim("a", when), SampleInventoryClaim("bb", when), SampleInventoryClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryClaims(xs)); got != 2 {
		t.Errorf("ActiveInventoryClaims() kept %d, want 2", got)
	}
	if got, want := SumInventoryClaims(xs), 1+3; got != want {
		t.Errorf("SumInventoryClaims() = %d, want %d", got, want)
	}
	SortInventoryClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryClaims() put %q first", xs[0].ID)
	}
	if len(IndexInventoryClaims(xs)) != 3 {
		t.Errorf("IndexInventoryClaims() = %v", IndexInventoryClaims(xs))
	}
	if err := ValidateInventoryClaims(ActiveInventoryClaims(xs)); err != nil {
		t.Errorf("ValidateInventoryClaims() = %v", err)
	}
}

func TestInventoryClaimNormalizeClamps(t *testing.T) {
	x := SampleInventoryClaim("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryClaim("r1", when)
	rev, err := r.PutInventoryClaim(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryClaim() = %v", err)
	}
	got, gotRev, err := r.GetInventoryClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryClaims() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryClaim(InventoryClaim{}, 0); err == nil {
		t.Error("PutInventoryClaim() accepted an empty entity")
	}
}

func TestInventoryCursorRoundTrip(t *testing.T) {
	x := SampleInventoryCursor("inventorycursor-1", when)
	got, err := ParseInventoryCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseInventoryCursor() = %v", err)
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

func TestInventoryCursorParseRejectsBadInput(t *testing.T) {
	f := SampleInventoryCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInventoryCursor(f); err == nil {
		t.Error("ParseInventoryCursor() accepted a record with no id")
	}
	f = SampleInventoryCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseInventoryCursor(f); err == nil {
		t.Error("ParseInventoryCursor() accepted a non-numeric version")
	}
}

func TestInventoryCursorCollections(t *testing.T) {
	xs := []InventoryCursor{SampleInventoryCursor("a", when), SampleInventoryCursor("bb", when), SampleInventoryCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInventoryCursors(xs)); got != 2 {
		t.Errorf("ActiveInventoryCursors() kept %d, want 2", got)
	}
	if got, want := SumInventoryCursors(xs), 1+3; got != want {
		t.Errorf("SumInventoryCursors() = %d, want %d", got, want)
	}
	SortInventoryCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInventoryCursors() put %q first", xs[0].ID)
	}
	if len(IndexInventoryCursors(xs)) != 3 {
		t.Errorf("IndexInventoryCursors() = %v", IndexInventoryCursors(xs))
	}
	if err := ValidateInventoryCursors(ActiveInventoryCursors(xs)); err != nil {
		t.Errorf("ValidateInventoryCursors() = %v", err)
	}
}

func TestInventoryCursorNormalizeClamps(t *testing.T) {
	x := SampleInventoryCursor("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInventoryCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleInventoryCursor("r1", when)
	rev, err := r.PutInventoryCursor(x, 0)
	if err != nil {
		t.Fatalf("PutInventoryCursor() = %v", err)
	}
	got, gotRev, err := r.GetInventoryCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetInventoryCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListInventoryCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListInventoryCursors() = %v, %v", all, err)
	}
	if _, err := r.PutInventoryCursor(InventoryCursor{}, 0); err == nil {
		t.Error("PutInventoryCursor() accepted an empty entity")
	}
}
