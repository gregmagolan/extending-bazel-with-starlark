package shipping

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
		if !strings.HasPrefix(Prefixes[i], "shipping/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleShippingRecord("ok", when).Fields()); err != nil {
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

func TestShippingRecordRoundTrip(t *testing.T) {
	x := SampleShippingRecord("shippingrecord-1", when)
	got, err := ParseShippingRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingRecord() = %v", err)
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

func TestShippingRecordParseRejectsBadInput(t *testing.T) {
	f := SampleShippingRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingRecord(f); err == nil {
		t.Error("ParseShippingRecord() accepted a record with no id")
	}
	f = SampleShippingRecord("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseShippingRecord(f); err == nil {
		t.Error("ParseShippingRecord() accepted a non-numeric attempts")
	}
}

func TestShippingRecordCollections(t *testing.T) {
	xs := []ShippingRecord{SampleShippingRecord("a", when), SampleShippingRecord("bb", when), SampleShippingRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingRecords(xs)); got != 2 {
		t.Errorf("ActiveShippingRecords() kept %d, want 2", got)
	}
	if got, want := SumShippingRecords(xs), 1+3; got != want {
		t.Errorf("SumShippingRecords() = %d, want %d", got, want)
	}
	SortShippingRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingRecords() put %q first", xs[0].ID)
	}
	if len(IndexShippingRecords(xs)) != 3 {
		t.Errorf("IndexShippingRecords() = %v", IndexShippingRecords(xs))
	}
	if err := ValidateShippingRecords(ActiveShippingRecords(xs)); err != nil {
		t.Errorf("ValidateShippingRecords() = %v", err)
	}
}

func TestShippingRecordNormalizeClamps(t *testing.T) {
	x := SampleShippingRecord("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingRecord("r1", when)
	rev, err := r.PutShippingRecord(x, 0)
	if err != nil {
		t.Fatalf("PutShippingRecord() = %v", err)
	}
	got, gotRev, err := r.GetShippingRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingRecords() = %v, %v", all, err)
	}
	if _, err := r.PutShippingRecord(ShippingRecord{}, 0); err == nil {
		t.Error("PutShippingRecord() accepted an empty entity")
	}
}

func TestShippingEntryRoundTrip(t *testing.T) {
	x := SampleShippingEntry("shippingentry-1", when)
	got, err := ParseShippingEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingEntry() = %v", err)
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

func TestShippingEntryParseRejectsBadInput(t *testing.T) {
	f := SampleShippingEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingEntry(f); err == nil {
		t.Error("ParseShippingEntry() accepted a record with no id")
	}
	f = SampleShippingEntry("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseShippingEntry(f); err == nil {
		t.Error("ParseShippingEntry() accepted a non-numeric weight")
	}
}

func TestShippingEntryCollections(t *testing.T) {
	xs := []ShippingEntry{SampleShippingEntry("a", when), SampleShippingEntry("bb", when), SampleShippingEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingEntrys(xs)); got != 2 {
		t.Errorf("ActiveShippingEntrys() kept %d, want 2", got)
	}
	if got, want := SumShippingEntrys(xs), 1+3; got != want {
		t.Errorf("SumShippingEntrys() = %d, want %d", got, want)
	}
	SortShippingEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingEntrys() put %q first", xs[0].ID)
	}
	if len(IndexShippingEntrys(xs)) != 3 {
		t.Errorf("IndexShippingEntrys() = %v", IndexShippingEntrys(xs))
	}
	if err := ValidateShippingEntrys(ActiveShippingEntrys(xs)); err != nil {
		t.Errorf("ValidateShippingEntrys() = %v", err)
	}
}

func TestShippingEntryNormalizeClamps(t *testing.T) {
	x := SampleShippingEntry("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingEntry("r1", when)
	rev, err := r.PutShippingEntry(x, 0)
	if err != nil {
		t.Fatalf("PutShippingEntry() = %v", err)
	}
	got, gotRev, err := r.GetShippingEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutShippingEntry(ShippingEntry{}, 0); err == nil {
		t.Error("PutShippingEntry() accepted an empty entity")
	}
}

func TestShippingItemRoundTrip(t *testing.T) {
	x := SampleShippingItem("shippingitem-1", when)
	got, err := ParseShippingItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingItem() = %v", err)
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

func TestShippingItemParseRejectsBadInput(t *testing.T) {
	f := SampleShippingItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingItem(f); err == nil {
		t.Error("ParseShippingItem() accepted a record with no id")
	}
	f = SampleShippingItem("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseShippingItem(f); err == nil {
		t.Error("ParseShippingItem() accepted a non-numeric capacity")
	}
}

func TestShippingItemCollections(t *testing.T) {
	xs := []ShippingItem{SampleShippingItem("a", when), SampleShippingItem("bb", when), SampleShippingItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingItems(xs)); got != 2 {
		t.Errorf("ActiveShippingItems() kept %d, want 2", got)
	}
	if got, want := SumShippingItems(xs), 1+3; got != want {
		t.Errorf("SumShippingItems() = %d, want %d", got, want)
	}
	SortShippingItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingItems() put %q first", xs[0].ID)
	}
	if len(IndexShippingItems(xs)) != 3 {
		t.Errorf("IndexShippingItems() = %v", IndexShippingItems(xs))
	}
	if err := ValidateShippingItems(ActiveShippingItems(xs)); err != nil {
		t.Errorf("ValidateShippingItems() = %v", err)
	}
}

func TestShippingItemNormalizeClamps(t *testing.T) {
	x := SampleShippingItem("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingItem("r1", when)
	rev, err := r.PutShippingItem(x, 0)
	if err != nil {
		t.Fatalf("PutShippingItem() = %v", err)
	}
	got, gotRev, err := r.GetShippingItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingItems() = %v, %v", all, err)
	}
	if _, err := r.PutShippingItem(ShippingItem{}, 0); err == nil {
		t.Error("PutShippingItem() accepted an empty entity")
	}
}

func TestShippingDraftRoundTrip(t *testing.T) {
	x := SampleShippingDraft("shippingdraft-1", when)
	got, err := ParseShippingDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingDraft() = %v", err)
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

func TestShippingDraftParseRejectsBadInput(t *testing.T) {
	f := SampleShippingDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingDraft(f); err == nil {
		t.Error("ParseShippingDraft() accepted a record with no id")
	}
	f = SampleShippingDraft("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseShippingDraft(f); err == nil {
		t.Error("ParseShippingDraft() accepted a non-numeric capacity")
	}
}

func TestShippingDraftCollections(t *testing.T) {
	xs := []ShippingDraft{SampleShippingDraft("a", when), SampleShippingDraft("bb", when), SampleShippingDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingDrafts(xs)); got != 2 {
		t.Errorf("ActiveShippingDrafts() kept %d, want 2", got)
	}
	if got, want := SumShippingDrafts(xs), 1+3; got != want {
		t.Errorf("SumShippingDrafts() = %d, want %d", got, want)
	}
	SortShippingDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingDrafts() put %q first", xs[0].ID)
	}
	if len(IndexShippingDrafts(xs)) != 3 {
		t.Errorf("IndexShippingDrafts() = %v", IndexShippingDrafts(xs))
	}
	if err := ValidateShippingDrafts(ActiveShippingDrafts(xs)); err != nil {
		t.Errorf("ValidateShippingDrafts() = %v", err)
	}
}

func TestShippingDraftNormalizeClamps(t *testing.T) {
	x := SampleShippingDraft("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingDraft("r1", when)
	rev, err := r.PutShippingDraft(x, 0)
	if err != nil {
		t.Fatalf("PutShippingDraft() = %v", err)
	}
	got, gotRev, err := r.GetShippingDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutShippingDraft(ShippingDraft{}, 0); err == nil {
		t.Error("PutShippingDraft() accepted an empty entity")
	}
}

func TestShippingRevisionRoundTrip(t *testing.T) {
	x := SampleShippingRevision("shippingrevision-1", when)
	got, err := ParseShippingRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingRevision() = %v", err)
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

func TestShippingRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleShippingRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingRevision(f); err == nil {
		t.Error("ParseShippingRevision() accepted a record with no id")
	}
	f = SampleShippingRevision("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseShippingRevision(f); err == nil {
		t.Error("ParseShippingRevision() accepted a non-numeric weight")
	}
}

func TestShippingRevisionCollections(t *testing.T) {
	xs := []ShippingRevision{SampleShippingRevision("a", when), SampleShippingRevision("bb", when), SampleShippingRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingRevisions(xs)); got != 2 {
		t.Errorf("ActiveShippingRevisions() kept %d, want 2", got)
	}
	if got, want := SumShippingRevisions(xs), 1+3; got != want {
		t.Errorf("SumShippingRevisions() = %d, want %d", got, want)
	}
	SortShippingRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingRevisions() put %q first", xs[0].ID)
	}
	if len(IndexShippingRevisions(xs)) != 3 {
		t.Errorf("IndexShippingRevisions() = %v", IndexShippingRevisions(xs))
	}
	if err := ValidateShippingRevisions(ActiveShippingRevisions(xs)); err != nil {
		t.Errorf("ValidateShippingRevisions() = %v", err)
	}
}

func TestShippingRevisionNormalizeClamps(t *testing.T) {
	x := SampleShippingRevision("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingRevision("r1", when)
	rev, err := r.PutShippingRevision(x, 0)
	if err != nil {
		t.Fatalf("PutShippingRevision() = %v", err)
	}
	got, gotRev, err := r.GetShippingRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutShippingRevision(ShippingRevision{}, 0); err == nil {
		t.Error("PutShippingRevision() accepted an empty entity")
	}
}

func TestShippingSnapshotRoundTrip(t *testing.T) {
	x := SampleShippingSnapshot("shippingsnapshot-1", when)
	got, err := ParseShippingSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingSnapshot() = %v", err)
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

func TestShippingSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleShippingSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingSnapshot(f); err == nil {
		t.Error("ParseShippingSnapshot() accepted a record with no id")
	}
	f = SampleShippingSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseShippingSnapshot(f); err == nil {
		t.Error("ParseShippingSnapshot() accepted a non-numeric priority")
	}
}

func TestShippingSnapshotCollections(t *testing.T) {
	xs := []ShippingSnapshot{SampleShippingSnapshot("a", when), SampleShippingSnapshot("bb", when), SampleShippingSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingSnapshots(xs)); got != 2 {
		t.Errorf("ActiveShippingSnapshots() kept %d, want 2", got)
	}
	if got, want := SumShippingSnapshots(xs), 1+3; got != want {
		t.Errorf("SumShippingSnapshots() = %d, want %d", got, want)
	}
	SortShippingSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexShippingSnapshots(xs)) != 3 {
		t.Errorf("IndexShippingSnapshots() = %v", IndexShippingSnapshots(xs))
	}
	if err := ValidateShippingSnapshots(ActiveShippingSnapshots(xs)); err != nil {
		t.Errorf("ValidateShippingSnapshots() = %v", err)
	}
}

func TestShippingSnapshotNormalizeClamps(t *testing.T) {
	x := SampleShippingSnapshot("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingSnapshot("r1", when)
	rev, err := r.PutShippingSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutShippingSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetShippingSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutShippingSnapshot(ShippingSnapshot{}, 0); err == nil {
		t.Error("PutShippingSnapshot() accepted an empty entity")
	}
}

func TestShippingBatchRoundTrip(t *testing.T) {
	x := SampleShippingBatch("shippingbatch-1", when)
	got, err := ParseShippingBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingBatch() = %v", err)
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

func TestShippingBatchParseRejectsBadInput(t *testing.T) {
	f := SampleShippingBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingBatch(f); err == nil {
		t.Error("ParseShippingBatch() accepted a record with no id")
	}
	f = SampleShippingBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseShippingBatch(f); err == nil {
		t.Error("ParseShippingBatch() accepted a non-numeric version")
	}
}

func TestShippingBatchCollections(t *testing.T) {
	xs := []ShippingBatch{SampleShippingBatch("a", when), SampleShippingBatch("bb", when), SampleShippingBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingBatchs(xs)); got != 2 {
		t.Errorf("ActiveShippingBatchs() kept %d, want 2", got)
	}
	if got, want := SumShippingBatchs(xs), 1+3; got != want {
		t.Errorf("SumShippingBatchs() = %d, want %d", got, want)
	}
	SortShippingBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingBatchs() put %q first", xs[0].ID)
	}
	if len(IndexShippingBatchs(xs)) != 3 {
		t.Errorf("IndexShippingBatchs() = %v", IndexShippingBatchs(xs))
	}
	if err := ValidateShippingBatchs(ActiveShippingBatchs(xs)); err != nil {
		t.Errorf("ValidateShippingBatchs() = %v", err)
	}
}

func TestShippingBatchNormalizeClamps(t *testing.T) {
	x := SampleShippingBatch("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingBatch("r1", when)
	rev, err := r.PutShippingBatch(x, 0)
	if err != nil {
		t.Fatalf("PutShippingBatch() = %v", err)
	}
	got, gotRev, err := r.GetShippingBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutShippingBatch(ShippingBatch{}, 0); err == nil {
		t.Error("PutShippingBatch() accepted an empty entity")
	}
}

func TestShippingRequestRoundTrip(t *testing.T) {
	x := SampleShippingRequest("shippingrequest-1", when)
	got, err := ParseShippingRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingRequest() = %v", err)
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

func TestShippingRequestParseRejectsBadInput(t *testing.T) {
	f := SampleShippingRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingRequest(f); err == nil {
		t.Error("ParseShippingRequest() accepted a record with no id")
	}
	f = SampleShippingRequest("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseShippingRequest(f); err == nil {
		t.Error("ParseShippingRequest() accepted a non-numeric version")
	}
}

func TestShippingRequestCollections(t *testing.T) {
	xs := []ShippingRequest{SampleShippingRequest("a", when), SampleShippingRequest("bb", when), SampleShippingRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingRequests(xs)); got != 2 {
		t.Errorf("ActiveShippingRequests() kept %d, want 2", got)
	}
	if got, want := SumShippingRequests(xs), 1+3; got != want {
		t.Errorf("SumShippingRequests() = %d, want %d", got, want)
	}
	SortShippingRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingRequests() put %q first", xs[0].ID)
	}
	if len(IndexShippingRequests(xs)) != 3 {
		t.Errorf("IndexShippingRequests() = %v", IndexShippingRequests(xs))
	}
	if err := ValidateShippingRequests(ActiveShippingRequests(xs)); err != nil {
		t.Errorf("ValidateShippingRequests() = %v", err)
	}
}

func TestShippingRequestNormalizeClamps(t *testing.T) {
	x := SampleShippingRequest("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingRequest("r1", when)
	rev, err := r.PutShippingRequest(x, 0)
	if err != nil {
		t.Fatalf("PutShippingRequest() = %v", err)
	}
	got, gotRev, err := r.GetShippingRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingRequests() = %v, %v", all, err)
	}
	if _, err := r.PutShippingRequest(ShippingRequest{}, 0); err == nil {
		t.Error("PutShippingRequest() accepted an empty entity")
	}
}

func TestShippingSummaryRoundTrip(t *testing.T) {
	x := SampleShippingSummary("shippingsummary-1", when)
	got, err := ParseShippingSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingSummary() = %v", err)
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

func TestShippingSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleShippingSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingSummary(f); err == nil {
		t.Error("ParseShippingSummary() accepted a record with no id")
	}
	f = SampleShippingSummary("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseShippingSummary(f); err == nil {
		t.Error("ParseShippingSummary() accepted a non-numeric retained")
	}
}

func TestShippingSummaryCollections(t *testing.T) {
	xs := []ShippingSummary{SampleShippingSummary("a", when), SampleShippingSummary("bb", when), SampleShippingSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingSummarys(xs)); got != 2 {
		t.Errorf("ActiveShippingSummarys() kept %d, want 2", got)
	}
	if got, want := SumShippingSummarys(xs), 1+3; got != want {
		t.Errorf("SumShippingSummarys() = %d, want %d", got, want)
	}
	SortShippingSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingSummarys() put %q first", xs[0].ID)
	}
	if len(IndexShippingSummarys(xs)) != 3 {
		t.Errorf("IndexShippingSummarys() = %v", IndexShippingSummarys(xs))
	}
	if err := ValidateShippingSummarys(ActiveShippingSummarys(xs)); err != nil {
		t.Errorf("ValidateShippingSummarys() = %v", err)
	}
}

func TestShippingSummaryNormalizeClamps(t *testing.T) {
	x := SampleShippingSummary("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingSummary("r1", when)
	rev, err := r.PutShippingSummary(x, 0)
	if err != nil {
		t.Fatalf("PutShippingSummary() = %v", err)
	}
	got, gotRev, err := r.GetShippingSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutShippingSummary(ShippingSummary{}, 0); err == nil {
		t.Error("PutShippingSummary() accepted an empty entity")
	}
}

func TestShippingPolicyRoundTrip(t *testing.T) {
	x := SampleShippingPolicy("shippingpolicy-1", when)
	got, err := ParseShippingPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingPolicy() = %v", err)
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

func TestShippingPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleShippingPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingPolicy(f); err == nil {
		t.Error("ParseShippingPolicy() accepted a record with no id")
	}
	f = SampleShippingPolicy("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseShippingPolicy(f); err == nil {
		t.Error("ParseShippingPolicy() accepted a non-numeric weight")
	}
}

func TestShippingPolicyCollections(t *testing.T) {
	xs := []ShippingPolicy{SampleShippingPolicy("a", when), SampleShippingPolicy("bb", when), SampleShippingPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingPolicys(xs)); got != 2 {
		t.Errorf("ActiveShippingPolicys() kept %d, want 2", got)
	}
	if got, want := SumShippingPolicys(xs), 1+3; got != want {
		t.Errorf("SumShippingPolicys() = %d, want %d", got, want)
	}
	SortShippingPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingPolicys() put %q first", xs[0].ID)
	}
	if len(IndexShippingPolicys(xs)) != 3 {
		t.Errorf("IndexShippingPolicys() = %v", IndexShippingPolicys(xs))
	}
	if err := ValidateShippingPolicys(ActiveShippingPolicys(xs)); err != nil {
		t.Errorf("ValidateShippingPolicys() = %v", err)
	}
}

func TestShippingPolicyNormalizeClamps(t *testing.T) {
	x := SampleShippingPolicy("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingPolicy("r1", when)
	rev, err := r.PutShippingPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutShippingPolicy() = %v", err)
	}
	got, gotRev, err := r.GetShippingPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutShippingPolicy(ShippingPolicy{}, 0); err == nil {
		t.Error("PutShippingPolicy() accepted an empty entity")
	}
}

func TestShippingLinkRoundTrip(t *testing.T) {
	x := SampleShippingLink("shippinglink-1", when)
	got, err := ParseShippingLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingLink() = %v", err)
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

func TestShippingLinkParseRejectsBadInput(t *testing.T) {
	f := SampleShippingLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingLink(f); err == nil {
		t.Error("ParseShippingLink() accepted a record with no id")
	}
	f = SampleShippingLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseShippingLink(f); err == nil {
		t.Error("ParseShippingLink() accepted a non-numeric retained")
	}
}

func TestShippingLinkCollections(t *testing.T) {
	xs := []ShippingLink{SampleShippingLink("a", when), SampleShippingLink("bb", when), SampleShippingLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingLinks(xs)); got != 2 {
		t.Errorf("ActiveShippingLinks() kept %d, want 2", got)
	}
	if got, want := SumShippingLinks(xs), 1+3; got != want {
		t.Errorf("SumShippingLinks() = %d, want %d", got, want)
	}
	SortShippingLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingLinks() put %q first", xs[0].ID)
	}
	if len(IndexShippingLinks(xs)) != 3 {
		t.Errorf("IndexShippingLinks() = %v", IndexShippingLinks(xs))
	}
	if err := ValidateShippingLinks(ActiveShippingLinks(xs)); err != nil {
		t.Errorf("ValidateShippingLinks() = %v", err)
	}
}

func TestShippingLinkNormalizeClamps(t *testing.T) {
	x := SampleShippingLink("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingLink("r1", when)
	rev, err := r.PutShippingLink(x, 0)
	if err != nil {
		t.Fatalf("PutShippingLink() = %v", err)
	}
	got, gotRev, err := r.GetShippingLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingLinks() = %v, %v", all, err)
	}
	if _, err := r.PutShippingLink(ShippingLink{}, 0); err == nil {
		t.Error("PutShippingLink() accepted an empty entity")
	}
}

func TestShippingNoteRoundTrip(t *testing.T) {
	x := SampleShippingNote("shippingnote-1", when)
	got, err := ParseShippingNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingNote() = %v", err)
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

func TestShippingNoteParseRejectsBadInput(t *testing.T) {
	f := SampleShippingNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingNote(f); err == nil {
		t.Error("ParseShippingNote() accepted a record with no id")
	}
	f = SampleShippingNote("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseShippingNote(f); err == nil {
		t.Error("ParseShippingNote() accepted a non-numeric count")
	}
}

func TestShippingNoteCollections(t *testing.T) {
	xs := []ShippingNote{SampleShippingNote("a", when), SampleShippingNote("bb", when), SampleShippingNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingNotes(xs)); got != 2 {
		t.Errorf("ActiveShippingNotes() kept %d, want 2", got)
	}
	if got, want := SumShippingNotes(xs), 1+3; got != want {
		t.Errorf("SumShippingNotes() = %d, want %d", got, want)
	}
	SortShippingNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingNotes() put %q first", xs[0].ID)
	}
	if len(IndexShippingNotes(xs)) != 3 {
		t.Errorf("IndexShippingNotes() = %v", IndexShippingNotes(xs))
	}
	if err := ValidateShippingNotes(ActiveShippingNotes(xs)); err != nil {
		t.Errorf("ValidateShippingNotes() = %v", err)
	}
}

func TestShippingNoteNormalizeClamps(t *testing.T) {
	x := SampleShippingNote("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingNote("r1", when)
	rev, err := r.PutShippingNote(x, 0)
	if err != nil {
		t.Fatalf("PutShippingNote() = %v", err)
	}
	got, gotRev, err := r.GetShippingNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingNotes() = %v, %v", all, err)
	}
	if _, err := r.PutShippingNote(ShippingNote{}, 0); err == nil {
		t.Error("PutShippingNote() accepted an empty entity")
	}
}

func TestShippingWindowRoundTrip(t *testing.T) {
	x := SampleShippingWindow("shippingwindow-1", when)
	got, err := ParseShippingWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingWindow() = %v", err)
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

func TestShippingWindowParseRejectsBadInput(t *testing.T) {
	f := SampleShippingWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingWindow(f); err == nil {
		t.Error("ParseShippingWindow() accepted a record with no id")
	}
	f = SampleShippingWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseShippingWindow(f); err == nil {
		t.Error("ParseShippingWindow() accepted a non-numeric priority")
	}
}

func TestShippingWindowCollections(t *testing.T) {
	xs := []ShippingWindow{SampleShippingWindow("a", when), SampleShippingWindow("bb", when), SampleShippingWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingWindows(xs)); got != 2 {
		t.Errorf("ActiveShippingWindows() kept %d, want 2", got)
	}
	if got, want := SumShippingWindows(xs), 1+3; got != want {
		t.Errorf("SumShippingWindows() = %d, want %d", got, want)
	}
	SortShippingWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingWindows() put %q first", xs[0].ID)
	}
	if len(IndexShippingWindows(xs)) != 3 {
		t.Errorf("IndexShippingWindows() = %v", IndexShippingWindows(xs))
	}
	if err := ValidateShippingWindows(ActiveShippingWindows(xs)); err != nil {
		t.Errorf("ValidateShippingWindows() = %v", err)
	}
}

func TestShippingWindowNormalizeClamps(t *testing.T) {
	x := SampleShippingWindow("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingWindow("r1", when)
	rev, err := r.PutShippingWindow(x, 0)
	if err != nil {
		t.Fatalf("PutShippingWindow() = %v", err)
	}
	got, gotRev, err := r.GetShippingWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingWindows() = %v, %v", all, err)
	}
	if _, err := r.PutShippingWindow(ShippingWindow{}, 0); err == nil {
		t.Error("PutShippingWindow() accepted an empty entity")
	}
}

func TestShippingQuotaRoundTrip(t *testing.T) {
	x := SampleShippingQuota("shippingquota-1", when)
	got, err := ParseShippingQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingQuota() = %v", err)
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

func TestShippingQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleShippingQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingQuota(f); err == nil {
		t.Error("ParseShippingQuota() accepted a record with no id")
	}
	f = SampleShippingQuota("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseShippingQuota(f); err == nil {
		t.Error("ParseShippingQuota() accepted a non-numeric weight")
	}
}

func TestShippingQuotaCollections(t *testing.T) {
	xs := []ShippingQuota{SampleShippingQuota("a", when), SampleShippingQuota("bb", when), SampleShippingQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingQuotas(xs)); got != 2 {
		t.Errorf("ActiveShippingQuotas() kept %d, want 2", got)
	}
	if got, want := SumShippingQuotas(xs), 1+3; got != want {
		t.Errorf("SumShippingQuotas() = %d, want %d", got, want)
	}
	SortShippingQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingQuotas() put %q first", xs[0].ID)
	}
	if len(IndexShippingQuotas(xs)) != 3 {
		t.Errorf("IndexShippingQuotas() = %v", IndexShippingQuotas(xs))
	}
	if err := ValidateShippingQuotas(ActiveShippingQuotas(xs)); err != nil {
		t.Errorf("ValidateShippingQuotas() = %v", err)
	}
}

func TestShippingQuotaNormalizeClamps(t *testing.T) {
	x := SampleShippingQuota("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingQuota("r1", when)
	rev, err := r.PutShippingQuota(x, 0)
	if err != nil {
		t.Fatalf("PutShippingQuota() = %v", err)
	}
	got, gotRev, err := r.GetShippingQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutShippingQuota(ShippingQuota{}, 0); err == nil {
		t.Error("PutShippingQuota() accepted an empty entity")
	}
}

func TestShippingBucketRoundTrip(t *testing.T) {
	x := SampleShippingBucket("shippingbucket-1", when)
	got, err := ParseShippingBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingBucket() = %v", err)
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

func TestShippingBucketParseRejectsBadInput(t *testing.T) {
	f := SampleShippingBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingBucket(f); err == nil {
		t.Error("ParseShippingBucket() accepted a record with no id")
	}
	f = SampleShippingBucket("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseShippingBucket(f); err == nil {
		t.Error("ParseShippingBucket() accepted a non-numeric priority")
	}
}

func TestShippingBucketCollections(t *testing.T) {
	xs := []ShippingBucket{SampleShippingBucket("a", when), SampleShippingBucket("bb", when), SampleShippingBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingBuckets(xs)); got != 2 {
		t.Errorf("ActiveShippingBuckets() kept %d, want 2", got)
	}
	if got, want := SumShippingBuckets(xs), 1+3; got != want {
		t.Errorf("SumShippingBuckets() = %d, want %d", got, want)
	}
	SortShippingBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingBuckets() put %q first", xs[0].ID)
	}
	if len(IndexShippingBuckets(xs)) != 3 {
		t.Errorf("IndexShippingBuckets() = %v", IndexShippingBuckets(xs))
	}
	if err := ValidateShippingBuckets(ActiveShippingBuckets(xs)); err != nil {
		t.Errorf("ValidateShippingBuckets() = %v", err)
	}
}

func TestShippingBucketNormalizeClamps(t *testing.T) {
	x := SampleShippingBucket("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingBucket("r1", when)
	rev, err := r.PutShippingBucket(x, 0)
	if err != nil {
		t.Fatalf("PutShippingBucket() = %v", err)
	}
	got, gotRev, err := r.GetShippingBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutShippingBucket(ShippingBucket{}, 0); err == nil {
		t.Error("PutShippingBucket() accepted an empty entity")
	}
}

func TestShippingLaneRoundTrip(t *testing.T) {
	x := SampleShippingLane("shippinglane-1", when)
	got, err := ParseShippingLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingLane() = %v", err)
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

func TestShippingLaneParseRejectsBadInput(t *testing.T) {
	f := SampleShippingLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingLane(f); err == nil {
		t.Error("ParseShippingLane() accepted a record with no id")
	}
	f = SampleShippingLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseShippingLane(f); err == nil {
		t.Error("ParseShippingLane() accepted a non-numeric priority")
	}
}

func TestShippingLaneCollections(t *testing.T) {
	xs := []ShippingLane{SampleShippingLane("a", when), SampleShippingLane("bb", when), SampleShippingLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingLanes(xs)); got != 2 {
		t.Errorf("ActiveShippingLanes() kept %d, want 2", got)
	}
	if got, want := SumShippingLanes(xs), 1+3; got != want {
		t.Errorf("SumShippingLanes() = %d, want %d", got, want)
	}
	SortShippingLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingLanes() put %q first", xs[0].ID)
	}
	if len(IndexShippingLanes(xs)) != 3 {
		t.Errorf("IndexShippingLanes() = %v", IndexShippingLanes(xs))
	}
	if err := ValidateShippingLanes(ActiveShippingLanes(xs)); err != nil {
		t.Errorf("ValidateShippingLanes() = %v", err)
	}
}

func TestShippingLaneNormalizeClamps(t *testing.T) {
	x := SampleShippingLane("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingLane("r1", when)
	rev, err := r.PutShippingLane(x, 0)
	if err != nil {
		t.Fatalf("PutShippingLane() = %v", err)
	}
	got, gotRev, err := r.GetShippingLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingLanes() = %v, %v", all, err)
	}
	if _, err := r.PutShippingLane(ShippingLane{}, 0); err == nil {
		t.Error("PutShippingLane() accepted an empty entity")
	}
}

func TestShippingSlotRoundTrip(t *testing.T) {
	x := SampleShippingSlot("shippingslot-1", when)
	got, err := ParseShippingSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingSlot() = %v", err)
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

func TestShippingSlotParseRejectsBadInput(t *testing.T) {
	f := SampleShippingSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingSlot(f); err == nil {
		t.Error("ParseShippingSlot() accepted a record with no id")
	}
	f = SampleShippingSlot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseShippingSlot(f); err == nil {
		t.Error("ParseShippingSlot() accepted a non-numeric offset")
	}
}

func TestShippingSlotCollections(t *testing.T) {
	xs := []ShippingSlot{SampleShippingSlot("a", when), SampleShippingSlot("bb", when), SampleShippingSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingSlots(xs)); got != 2 {
		t.Errorf("ActiveShippingSlots() kept %d, want 2", got)
	}
	if got, want := SumShippingSlots(xs), 1+3; got != want {
		t.Errorf("SumShippingSlots() = %d, want %d", got, want)
	}
	SortShippingSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingSlots() put %q first", xs[0].ID)
	}
	if len(IndexShippingSlots(xs)) != 3 {
		t.Errorf("IndexShippingSlots() = %v", IndexShippingSlots(xs))
	}
	if err := ValidateShippingSlots(ActiveShippingSlots(xs)); err != nil {
		t.Errorf("ValidateShippingSlots() = %v", err)
	}
}

func TestShippingSlotNormalizeClamps(t *testing.T) {
	x := SampleShippingSlot("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingSlot("r1", when)
	rev, err := r.PutShippingSlot(x, 0)
	if err != nil {
		t.Fatalf("PutShippingSlot() = %v", err)
	}
	got, gotRev, err := r.GetShippingSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingSlots() = %v, %v", all, err)
	}
	if _, err := r.PutShippingSlot(ShippingSlot{}, 0); err == nil {
		t.Error("PutShippingSlot() accepted an empty entity")
	}
}

func TestShippingTagRoundTrip(t *testing.T) {
	x := SampleShippingTag("shippingtag-1", when)
	got, err := ParseShippingTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingTag() = %v", err)
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

func TestShippingTagParseRejectsBadInput(t *testing.T) {
	f := SampleShippingTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingTag(f); err == nil {
		t.Error("ParseShippingTag() accepted a record with no id")
	}
	f = SampleShippingTag("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseShippingTag(f); err == nil {
		t.Error("ParseShippingTag() accepted a non-numeric priority")
	}
}

func TestShippingTagCollections(t *testing.T) {
	xs := []ShippingTag{SampleShippingTag("a", when), SampleShippingTag("bb", when), SampleShippingTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingTags(xs)); got != 2 {
		t.Errorf("ActiveShippingTags() kept %d, want 2", got)
	}
	if got, want := SumShippingTags(xs), 1+3; got != want {
		t.Errorf("SumShippingTags() = %d, want %d", got, want)
	}
	SortShippingTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingTags() put %q first", xs[0].ID)
	}
	if len(IndexShippingTags(xs)) != 3 {
		t.Errorf("IndexShippingTags() = %v", IndexShippingTags(xs))
	}
	if err := ValidateShippingTags(ActiveShippingTags(xs)); err != nil {
		t.Errorf("ValidateShippingTags() = %v", err)
	}
}

func TestShippingTagNormalizeClamps(t *testing.T) {
	x := SampleShippingTag("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingTag("r1", when)
	rev, err := r.PutShippingTag(x, 0)
	if err != nil {
		t.Fatalf("PutShippingTag() = %v", err)
	}
	got, gotRev, err := r.GetShippingTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingTags() = %v, %v", all, err)
	}
	if _, err := r.PutShippingTag(ShippingTag{}, 0); err == nil {
		t.Error("PutShippingTag() accepted an empty entity")
	}
}

func TestShippingRuleRoundTrip(t *testing.T) {
	x := SampleShippingRule("shippingrule-1", when)
	got, err := ParseShippingRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingRule() = %v", err)
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

func TestShippingRuleParseRejectsBadInput(t *testing.T) {
	f := SampleShippingRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingRule(f); err == nil {
		t.Error("ParseShippingRule() accepted a record with no id")
	}
	f = SampleShippingRule("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseShippingRule(f); err == nil {
		t.Error("ParseShippingRule() accepted a non-numeric capacity")
	}
}

func TestShippingRuleCollections(t *testing.T) {
	xs := []ShippingRule{SampleShippingRule("a", when), SampleShippingRule("bb", when), SampleShippingRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingRules(xs)); got != 2 {
		t.Errorf("ActiveShippingRules() kept %d, want 2", got)
	}
	if got, want := SumShippingRules(xs), 1+3; got != want {
		t.Errorf("SumShippingRules() = %d, want %d", got, want)
	}
	SortShippingRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingRules() put %q first", xs[0].ID)
	}
	if len(IndexShippingRules(xs)) != 3 {
		t.Errorf("IndexShippingRules() = %v", IndexShippingRules(xs))
	}
	if err := ValidateShippingRules(ActiveShippingRules(xs)); err != nil {
		t.Errorf("ValidateShippingRules() = %v", err)
	}
}

func TestShippingRuleNormalizeClamps(t *testing.T) {
	x := SampleShippingRule("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingRule("r1", when)
	rev, err := r.PutShippingRule(x, 0)
	if err != nil {
		t.Fatalf("PutShippingRule() = %v", err)
	}
	got, gotRev, err := r.GetShippingRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingRules() = %v, %v", all, err)
	}
	if _, err := r.PutShippingRule(ShippingRule{}, 0); err == nil {
		t.Error("PutShippingRule() accepted an empty entity")
	}
}

func TestShippingHoldRoundTrip(t *testing.T) {
	x := SampleShippingHold("shippinghold-1", when)
	got, err := ParseShippingHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingHold() = %v", err)
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

func TestShippingHoldParseRejectsBadInput(t *testing.T) {
	f := SampleShippingHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingHold(f); err == nil {
		t.Error("ParseShippingHold() accepted a record with no id")
	}
	f = SampleShippingHold("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseShippingHold(f); err == nil {
		t.Error("ParseShippingHold() accepted a non-numeric priority")
	}
}

func TestShippingHoldCollections(t *testing.T) {
	xs := []ShippingHold{SampleShippingHold("a", when), SampleShippingHold("bb", when), SampleShippingHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingHolds(xs)); got != 2 {
		t.Errorf("ActiveShippingHolds() kept %d, want 2", got)
	}
	if got, want := SumShippingHolds(xs), 1+3; got != want {
		t.Errorf("SumShippingHolds() = %d, want %d", got, want)
	}
	SortShippingHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingHolds() put %q first", xs[0].ID)
	}
	if len(IndexShippingHolds(xs)) != 3 {
		t.Errorf("IndexShippingHolds() = %v", IndexShippingHolds(xs))
	}
	if err := ValidateShippingHolds(ActiveShippingHolds(xs)); err != nil {
		t.Errorf("ValidateShippingHolds() = %v", err)
	}
}

func TestShippingHoldNormalizeClamps(t *testing.T) {
	x := SampleShippingHold("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingHold("r1", when)
	rev, err := r.PutShippingHold(x, 0)
	if err != nil {
		t.Fatalf("PutShippingHold() = %v", err)
	}
	got, gotRev, err := r.GetShippingHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingHolds() = %v, %v", all, err)
	}
	if _, err := r.PutShippingHold(ShippingHold{}, 0); err == nil {
		t.Error("PutShippingHold() accepted an empty entity")
	}
}

func TestShippingClaimRoundTrip(t *testing.T) {
	x := SampleShippingClaim("shippingclaim-1", when)
	got, err := ParseShippingClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingClaim() = %v", err)
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

func TestShippingClaimParseRejectsBadInput(t *testing.T) {
	f := SampleShippingClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingClaim(f); err == nil {
		t.Error("ParseShippingClaim() accepted a record with no id")
	}
	f = SampleShippingClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseShippingClaim(f); err == nil {
		t.Error("ParseShippingClaim() accepted a non-numeric retained")
	}
}

func TestShippingClaimCollections(t *testing.T) {
	xs := []ShippingClaim{SampleShippingClaim("a", when), SampleShippingClaim("bb", when), SampleShippingClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingClaims(xs)); got != 2 {
		t.Errorf("ActiveShippingClaims() kept %d, want 2", got)
	}
	if got, want := SumShippingClaims(xs), 1+3; got != want {
		t.Errorf("SumShippingClaims() = %d, want %d", got, want)
	}
	SortShippingClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingClaims() put %q first", xs[0].ID)
	}
	if len(IndexShippingClaims(xs)) != 3 {
		t.Errorf("IndexShippingClaims() = %v", IndexShippingClaims(xs))
	}
	if err := ValidateShippingClaims(ActiveShippingClaims(xs)); err != nil {
		t.Errorf("ValidateShippingClaims() = %v", err)
	}
}

func TestShippingClaimNormalizeClamps(t *testing.T) {
	x := SampleShippingClaim("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingClaim("r1", when)
	rev, err := r.PutShippingClaim(x, 0)
	if err != nil {
		t.Fatalf("PutShippingClaim() = %v", err)
	}
	got, gotRev, err := r.GetShippingClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingClaims() = %v, %v", all, err)
	}
	if _, err := r.PutShippingClaim(ShippingClaim{}, 0); err == nil {
		t.Error("PutShippingClaim() accepted an empty entity")
	}
}

func TestShippingCursorRoundTrip(t *testing.T) {
	x := SampleShippingCursor("shippingcursor-1", when)
	got, err := ParseShippingCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseShippingCursor() = %v", err)
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

func TestShippingCursorParseRejectsBadInput(t *testing.T) {
	f := SampleShippingCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseShippingCursor(f); err == nil {
		t.Error("ParseShippingCursor() accepted a record with no id")
	}
	f = SampleShippingCursor("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseShippingCursor(f); err == nil {
		t.Error("ParseShippingCursor() accepted a non-numeric count")
	}
}

func TestShippingCursorCollections(t *testing.T) {
	xs := []ShippingCursor{SampleShippingCursor("a", when), SampleShippingCursor("bb", when), SampleShippingCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveShippingCursors(xs)); got != 2 {
		t.Errorf("ActiveShippingCursors() kept %d, want 2", got)
	}
	if got, want := SumShippingCursors(xs), 1+3; got != want {
		t.Errorf("SumShippingCursors() = %d, want %d", got, want)
	}
	SortShippingCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortShippingCursors() put %q first", xs[0].ID)
	}
	if len(IndexShippingCursors(xs)) != 3 {
		t.Errorf("IndexShippingCursors() = %v", IndexShippingCursors(xs))
	}
	if err := ValidateShippingCursors(ActiveShippingCursors(xs)); err != nil {
		t.Errorf("ValidateShippingCursors() = %v", err)
	}
}

func TestShippingCursorNormalizeClamps(t *testing.T) {
	x := SampleShippingCursor("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestShippingCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleShippingCursor("r1", when)
	rev, err := r.PutShippingCursor(x, 0)
	if err != nil {
		t.Fatalf("PutShippingCursor() = %v", err)
	}
	got, gotRev, err := r.GetShippingCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetShippingCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListShippingCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListShippingCursors() = %v, %v", all, err)
	}
	if _, err := r.PutShippingCursor(ShippingCursor{}, 0); err == nil {
		t.Error("PutShippingCursor() accepted an empty entity")
	}
}
