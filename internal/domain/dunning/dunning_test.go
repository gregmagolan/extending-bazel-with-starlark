package dunning

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
		if !strings.HasPrefix(Prefixes[i], "dunning/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleDunningRecord("ok", when).Fields()); err != nil {
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

func TestDunningRecordRoundTrip(t *testing.T) {
	x := SampleDunningRecord("dunningrecord-1", when)
	got, err := ParseDunningRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningRecord() = %v", err)
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

func TestDunningRecordParseRejectsBadInput(t *testing.T) {
	f := SampleDunningRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningRecord(f); err == nil {
		t.Error("ParseDunningRecord() accepted a record with no id")
	}
	f = SampleDunningRecord("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDunningRecord(f); err == nil {
		t.Error("ParseDunningRecord() accepted a non-numeric count")
	}
}

func TestDunningRecordCollections(t *testing.T) {
	xs := []DunningRecord{SampleDunningRecord("a", when), SampleDunningRecord("bb", when), SampleDunningRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningRecords(xs)); got != 2 {
		t.Errorf("ActiveDunningRecords() kept %d, want 2", got)
	}
	if got, want := SumDunningRecords(xs), 1+3; got != want {
		t.Errorf("SumDunningRecords() = %d, want %d", got, want)
	}
	SortDunningRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningRecords() put %q first", xs[0].ID)
	}
	if len(IndexDunningRecords(xs)) != 3 {
		t.Errorf("IndexDunningRecords() = %v", IndexDunningRecords(xs))
	}
	if err := ValidateDunningRecords(ActiveDunningRecords(xs)); err != nil {
		t.Errorf("ValidateDunningRecords() = %v", err)
	}
}

func TestDunningRecordNormalizeClamps(t *testing.T) {
	x := SampleDunningRecord("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningRecord("r1", when)
	rev, err := r.PutDunningRecord(x, 0)
	if err != nil {
		t.Fatalf("PutDunningRecord() = %v", err)
	}
	got, gotRev, err := r.GetDunningRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningRecords() = %v, %v", all, err)
	}
	if _, err := r.PutDunningRecord(DunningRecord{}, 0); err == nil {
		t.Error("PutDunningRecord() accepted an empty entity")
	}
}

func TestDunningEntryRoundTrip(t *testing.T) {
	x := SampleDunningEntry("dunningentry-1", when)
	got, err := ParseDunningEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningEntry() = %v", err)
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

func TestDunningEntryParseRejectsBadInput(t *testing.T) {
	f := SampleDunningEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningEntry(f); err == nil {
		t.Error("ParseDunningEntry() accepted a record with no id")
	}
	f = SampleDunningEntry("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDunningEntry(f); err == nil {
		t.Error("ParseDunningEntry() accepted a non-numeric retained")
	}
}

func TestDunningEntryCollections(t *testing.T) {
	xs := []DunningEntry{SampleDunningEntry("a", when), SampleDunningEntry("bb", when), SampleDunningEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningEntrys(xs)); got != 2 {
		t.Errorf("ActiveDunningEntrys() kept %d, want 2", got)
	}
	if got, want := SumDunningEntrys(xs), 1+3; got != want {
		t.Errorf("SumDunningEntrys() = %d, want %d", got, want)
	}
	SortDunningEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningEntrys() put %q first", xs[0].ID)
	}
	if len(IndexDunningEntrys(xs)) != 3 {
		t.Errorf("IndexDunningEntrys() = %v", IndexDunningEntrys(xs))
	}
	if err := ValidateDunningEntrys(ActiveDunningEntrys(xs)); err != nil {
		t.Errorf("ValidateDunningEntrys() = %v", err)
	}
}

func TestDunningEntryNormalizeClamps(t *testing.T) {
	x := SampleDunningEntry("a", when)
	x.Reference = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningEntry("r1", when)
	rev, err := r.PutDunningEntry(x, 0)
	if err != nil {
		t.Fatalf("PutDunningEntry() = %v", err)
	}
	got, gotRev, err := r.GetDunningEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutDunningEntry(DunningEntry{}, 0); err == nil {
		t.Error("PutDunningEntry() accepted an empty entity")
	}
}

func TestDunningItemRoundTrip(t *testing.T) {
	x := SampleDunningItem("dunningitem-1", when)
	got, err := ParseDunningItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningItem() = %v", err)
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

func TestDunningItemParseRejectsBadInput(t *testing.T) {
	f := SampleDunningItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningItem(f); err == nil {
		t.Error("ParseDunningItem() accepted a record with no id")
	}
	f = SampleDunningItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDunningItem(f); err == nil {
		t.Error("ParseDunningItem() accepted a non-numeric version")
	}
}

func TestDunningItemCollections(t *testing.T) {
	xs := []DunningItem{SampleDunningItem("a", when), SampleDunningItem("bb", when), SampleDunningItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningItems(xs)); got != 2 {
		t.Errorf("ActiveDunningItems() kept %d, want 2", got)
	}
	if got, want := SumDunningItems(xs), 1+3; got != want {
		t.Errorf("SumDunningItems() = %d, want %d", got, want)
	}
	SortDunningItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningItems() put %q first", xs[0].ID)
	}
	if len(IndexDunningItems(xs)) != 3 {
		t.Errorf("IndexDunningItems() = %v", IndexDunningItems(xs))
	}
	if err := ValidateDunningItems(ActiveDunningItems(xs)); err != nil {
		t.Errorf("ValidateDunningItems() = %v", err)
	}
}

func TestDunningItemNormalizeClamps(t *testing.T) {
	x := SampleDunningItem("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningItem("r1", when)
	rev, err := r.PutDunningItem(x, 0)
	if err != nil {
		t.Fatalf("PutDunningItem() = %v", err)
	}
	got, gotRev, err := r.GetDunningItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningItems() = %v, %v", all, err)
	}
	if _, err := r.PutDunningItem(DunningItem{}, 0); err == nil {
		t.Error("PutDunningItem() accepted an empty entity")
	}
}

func TestDunningDraftRoundTrip(t *testing.T) {
	x := SampleDunningDraft("dunningdraft-1", when)
	got, err := ParseDunningDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningDraft() = %v", err)
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

func TestDunningDraftParseRejectsBadInput(t *testing.T) {
	f := SampleDunningDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningDraft(f); err == nil {
		t.Error("ParseDunningDraft() accepted a record with no id")
	}
	f = SampleDunningDraft("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDunningDraft(f); err == nil {
		t.Error("ParseDunningDraft() accepted a non-numeric capacity")
	}
}

func TestDunningDraftCollections(t *testing.T) {
	xs := []DunningDraft{SampleDunningDraft("a", when), SampleDunningDraft("bb", when), SampleDunningDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningDrafts(xs)); got != 2 {
		t.Errorf("ActiveDunningDrafts() kept %d, want 2", got)
	}
	if got, want := SumDunningDrafts(xs), 1+3; got != want {
		t.Errorf("SumDunningDrafts() = %d, want %d", got, want)
	}
	SortDunningDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningDrafts() put %q first", xs[0].ID)
	}
	if len(IndexDunningDrafts(xs)) != 3 {
		t.Errorf("IndexDunningDrafts() = %v", IndexDunningDrafts(xs))
	}
	if err := ValidateDunningDrafts(ActiveDunningDrafts(xs)); err != nil {
		t.Errorf("ValidateDunningDrafts() = %v", err)
	}
}

func TestDunningDraftNormalizeClamps(t *testing.T) {
	x := SampleDunningDraft("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningDraft("r1", when)
	rev, err := r.PutDunningDraft(x, 0)
	if err != nil {
		t.Fatalf("PutDunningDraft() = %v", err)
	}
	got, gotRev, err := r.GetDunningDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutDunningDraft(DunningDraft{}, 0); err == nil {
		t.Error("PutDunningDraft() accepted an empty entity")
	}
}

func TestDunningRevisionRoundTrip(t *testing.T) {
	x := SampleDunningRevision("dunningrevision-1", when)
	got, err := ParseDunningRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningRevision() = %v", err)
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

func TestDunningRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleDunningRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningRevision(f); err == nil {
		t.Error("ParseDunningRevision() accepted a record with no id")
	}
	f = SampleDunningRevision("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDunningRevision(f); err == nil {
		t.Error("ParseDunningRevision() accepted a non-numeric capacity")
	}
}

func TestDunningRevisionCollections(t *testing.T) {
	xs := []DunningRevision{SampleDunningRevision("a", when), SampleDunningRevision("bb", when), SampleDunningRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningRevisions(xs)); got != 2 {
		t.Errorf("ActiveDunningRevisions() kept %d, want 2", got)
	}
	if got, want := SumDunningRevisions(xs), 1+3; got != want {
		t.Errorf("SumDunningRevisions() = %d, want %d", got, want)
	}
	SortDunningRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningRevisions() put %q first", xs[0].ID)
	}
	if len(IndexDunningRevisions(xs)) != 3 {
		t.Errorf("IndexDunningRevisions() = %v", IndexDunningRevisions(xs))
	}
	if err := ValidateDunningRevisions(ActiveDunningRevisions(xs)); err != nil {
		t.Errorf("ValidateDunningRevisions() = %v", err)
	}
}

func TestDunningRevisionNormalizeClamps(t *testing.T) {
	x := SampleDunningRevision("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningRevision("r1", when)
	rev, err := r.PutDunningRevision(x, 0)
	if err != nil {
		t.Fatalf("PutDunningRevision() = %v", err)
	}
	got, gotRev, err := r.GetDunningRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutDunningRevision(DunningRevision{}, 0); err == nil {
		t.Error("PutDunningRevision() accepted an empty entity")
	}
}

func TestDunningSnapshotRoundTrip(t *testing.T) {
	x := SampleDunningSnapshot("dunningsnapshot-1", when)
	got, err := ParseDunningSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningSnapshot() = %v", err)
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

func TestDunningSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleDunningSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningSnapshot(f); err == nil {
		t.Error("ParseDunningSnapshot() accepted a record with no id")
	}
	f = SampleDunningSnapshot("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDunningSnapshot(f); err == nil {
		t.Error("ParseDunningSnapshot() accepted a non-numeric version")
	}
}

func TestDunningSnapshotCollections(t *testing.T) {
	xs := []DunningSnapshot{SampleDunningSnapshot("a", when), SampleDunningSnapshot("bb", when), SampleDunningSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningSnapshots(xs)); got != 2 {
		t.Errorf("ActiveDunningSnapshots() kept %d, want 2", got)
	}
	if got, want := SumDunningSnapshots(xs), 1+3; got != want {
		t.Errorf("SumDunningSnapshots() = %d, want %d", got, want)
	}
	SortDunningSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexDunningSnapshots(xs)) != 3 {
		t.Errorf("IndexDunningSnapshots() = %v", IndexDunningSnapshots(xs))
	}
	if err := ValidateDunningSnapshots(ActiveDunningSnapshots(xs)); err != nil {
		t.Errorf("ValidateDunningSnapshots() = %v", err)
	}
}

func TestDunningSnapshotNormalizeClamps(t *testing.T) {
	x := SampleDunningSnapshot("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningSnapshot("r1", when)
	rev, err := r.PutDunningSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutDunningSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetDunningSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutDunningSnapshot(DunningSnapshot{}, 0); err == nil {
		t.Error("PutDunningSnapshot() accepted an empty entity")
	}
}

func TestDunningBatchRoundTrip(t *testing.T) {
	x := SampleDunningBatch("dunningbatch-1", when)
	got, err := ParseDunningBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningBatch() = %v", err)
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

func TestDunningBatchParseRejectsBadInput(t *testing.T) {
	f := SampleDunningBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningBatch(f); err == nil {
		t.Error("ParseDunningBatch() accepted a record with no id")
	}
	f = SampleDunningBatch("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDunningBatch(f); err == nil {
		t.Error("ParseDunningBatch() accepted a non-numeric offset")
	}
}

func TestDunningBatchCollections(t *testing.T) {
	xs := []DunningBatch{SampleDunningBatch("a", when), SampleDunningBatch("bb", when), SampleDunningBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningBatchs(xs)); got != 2 {
		t.Errorf("ActiveDunningBatchs() kept %d, want 2", got)
	}
	if got, want := SumDunningBatchs(xs), 1+3; got != want {
		t.Errorf("SumDunningBatchs() = %d, want %d", got, want)
	}
	SortDunningBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningBatchs() put %q first", xs[0].ID)
	}
	if len(IndexDunningBatchs(xs)) != 3 {
		t.Errorf("IndexDunningBatchs() = %v", IndexDunningBatchs(xs))
	}
	if err := ValidateDunningBatchs(ActiveDunningBatchs(xs)); err != nil {
		t.Errorf("ValidateDunningBatchs() = %v", err)
	}
}

func TestDunningBatchNormalizeClamps(t *testing.T) {
	x := SampleDunningBatch("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningBatch("r1", when)
	rev, err := r.PutDunningBatch(x, 0)
	if err != nil {
		t.Fatalf("PutDunningBatch() = %v", err)
	}
	got, gotRev, err := r.GetDunningBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutDunningBatch(DunningBatch{}, 0); err == nil {
		t.Error("PutDunningBatch() accepted an empty entity")
	}
}

func TestDunningRequestRoundTrip(t *testing.T) {
	x := SampleDunningRequest("dunningrequest-1", when)
	got, err := ParseDunningRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningRequest() = %v", err)
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

func TestDunningRequestParseRejectsBadInput(t *testing.T) {
	f := SampleDunningRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningRequest(f); err == nil {
		t.Error("ParseDunningRequest() accepted a record with no id")
	}
	f = SampleDunningRequest("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDunningRequest(f); err == nil {
		t.Error("ParseDunningRequest() accepted a non-numeric version")
	}
}

func TestDunningRequestCollections(t *testing.T) {
	xs := []DunningRequest{SampleDunningRequest("a", when), SampleDunningRequest("bb", when), SampleDunningRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningRequests(xs)); got != 2 {
		t.Errorf("ActiveDunningRequests() kept %d, want 2", got)
	}
	if got, want := SumDunningRequests(xs), 1+3; got != want {
		t.Errorf("SumDunningRequests() = %d, want %d", got, want)
	}
	SortDunningRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningRequests() put %q first", xs[0].ID)
	}
	if len(IndexDunningRequests(xs)) != 3 {
		t.Errorf("IndexDunningRequests() = %v", IndexDunningRequests(xs))
	}
	if err := ValidateDunningRequests(ActiveDunningRequests(xs)); err != nil {
		t.Errorf("ValidateDunningRequests() = %v", err)
	}
}

func TestDunningRequestNormalizeClamps(t *testing.T) {
	x := SampleDunningRequest("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningRequest("r1", when)
	rev, err := r.PutDunningRequest(x, 0)
	if err != nil {
		t.Fatalf("PutDunningRequest() = %v", err)
	}
	got, gotRev, err := r.GetDunningRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningRequests() = %v, %v", all, err)
	}
	if _, err := r.PutDunningRequest(DunningRequest{}, 0); err == nil {
		t.Error("PutDunningRequest() accepted an empty entity")
	}
}

func TestDunningSummaryRoundTrip(t *testing.T) {
	x := SampleDunningSummary("dunningsummary-1", when)
	got, err := ParseDunningSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningSummary() = %v", err)
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

func TestDunningSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleDunningSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningSummary(f); err == nil {
		t.Error("ParseDunningSummary() accepted a record with no id")
	}
	f = SampleDunningSummary("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDunningSummary(f); err == nil {
		t.Error("ParseDunningSummary() accepted a non-numeric weight")
	}
}

func TestDunningSummaryCollections(t *testing.T) {
	xs := []DunningSummary{SampleDunningSummary("a", when), SampleDunningSummary("bb", when), SampleDunningSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningSummarys(xs)); got != 2 {
		t.Errorf("ActiveDunningSummarys() kept %d, want 2", got)
	}
	if got, want := SumDunningSummarys(xs), 1+3; got != want {
		t.Errorf("SumDunningSummarys() = %d, want %d", got, want)
	}
	SortDunningSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningSummarys() put %q first", xs[0].ID)
	}
	if len(IndexDunningSummarys(xs)) != 3 {
		t.Errorf("IndexDunningSummarys() = %v", IndexDunningSummarys(xs))
	}
	if err := ValidateDunningSummarys(ActiveDunningSummarys(xs)); err != nil {
		t.Errorf("ValidateDunningSummarys() = %v", err)
	}
}

func TestDunningSummaryNormalizeClamps(t *testing.T) {
	x := SampleDunningSummary("a", when)
	x.Owner = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningSummary("r1", when)
	rev, err := r.PutDunningSummary(x, 0)
	if err != nil {
		t.Fatalf("PutDunningSummary() = %v", err)
	}
	got, gotRev, err := r.GetDunningSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutDunningSummary(DunningSummary{}, 0); err == nil {
		t.Error("PutDunningSummary() accepted an empty entity")
	}
}

func TestDunningPolicyRoundTrip(t *testing.T) {
	x := SampleDunningPolicy("dunningpolicy-1", when)
	got, err := ParseDunningPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningPolicy() = %v", err)
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

func TestDunningPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleDunningPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningPolicy(f); err == nil {
		t.Error("ParseDunningPolicy() accepted a record with no id")
	}
	f = SampleDunningPolicy("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDunningPolicy(f); err == nil {
		t.Error("ParseDunningPolicy() accepted a non-numeric attempts")
	}
}

func TestDunningPolicyCollections(t *testing.T) {
	xs := []DunningPolicy{SampleDunningPolicy("a", when), SampleDunningPolicy("bb", when), SampleDunningPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningPolicys(xs)); got != 2 {
		t.Errorf("ActiveDunningPolicys() kept %d, want 2", got)
	}
	if got, want := SumDunningPolicys(xs), 1+3; got != want {
		t.Errorf("SumDunningPolicys() = %d, want %d", got, want)
	}
	SortDunningPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningPolicys() put %q first", xs[0].ID)
	}
	if len(IndexDunningPolicys(xs)) != 3 {
		t.Errorf("IndexDunningPolicys() = %v", IndexDunningPolicys(xs))
	}
	if err := ValidateDunningPolicys(ActiveDunningPolicys(xs)); err != nil {
		t.Errorf("ValidateDunningPolicys() = %v", err)
	}
}

func TestDunningPolicyNormalizeClamps(t *testing.T) {
	x := SampleDunningPolicy("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningPolicy("r1", when)
	rev, err := r.PutDunningPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutDunningPolicy() = %v", err)
	}
	got, gotRev, err := r.GetDunningPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutDunningPolicy(DunningPolicy{}, 0); err == nil {
		t.Error("PutDunningPolicy() accepted an empty entity")
	}
}

func TestDunningLinkRoundTrip(t *testing.T) {
	x := SampleDunningLink("dunninglink-1", when)
	got, err := ParseDunningLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningLink() = %v", err)
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

func TestDunningLinkParseRejectsBadInput(t *testing.T) {
	f := SampleDunningLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningLink(f); err == nil {
		t.Error("ParseDunningLink() accepted a record with no id")
	}
	f = SampleDunningLink("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDunningLink(f); err == nil {
		t.Error("ParseDunningLink() accepted a non-numeric priority")
	}
}

func TestDunningLinkCollections(t *testing.T) {
	xs := []DunningLink{SampleDunningLink("a", when), SampleDunningLink("bb", when), SampleDunningLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningLinks(xs)); got != 2 {
		t.Errorf("ActiveDunningLinks() kept %d, want 2", got)
	}
	if got, want := SumDunningLinks(xs), 1+3; got != want {
		t.Errorf("SumDunningLinks() = %d, want %d", got, want)
	}
	SortDunningLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningLinks() put %q first", xs[0].ID)
	}
	if len(IndexDunningLinks(xs)) != 3 {
		t.Errorf("IndexDunningLinks() = %v", IndexDunningLinks(xs))
	}
	if err := ValidateDunningLinks(ActiveDunningLinks(xs)); err != nil {
		t.Errorf("ValidateDunningLinks() = %v", err)
	}
}

func TestDunningLinkNormalizeClamps(t *testing.T) {
	x := SampleDunningLink("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningLink("r1", when)
	rev, err := r.PutDunningLink(x, 0)
	if err != nil {
		t.Fatalf("PutDunningLink() = %v", err)
	}
	got, gotRev, err := r.GetDunningLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningLinks() = %v, %v", all, err)
	}
	if _, err := r.PutDunningLink(DunningLink{}, 0); err == nil {
		t.Error("PutDunningLink() accepted an empty entity")
	}
}

func TestDunningNoteRoundTrip(t *testing.T) {
	x := SampleDunningNote("dunningnote-1", when)
	got, err := ParseDunningNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningNote() = %v", err)
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

func TestDunningNoteParseRejectsBadInput(t *testing.T) {
	f := SampleDunningNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningNote(f); err == nil {
		t.Error("ParseDunningNote() accepted a record with no id")
	}
	f = SampleDunningNote("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDunningNote(f); err == nil {
		t.Error("ParseDunningNote() accepted a non-numeric weight")
	}
}

func TestDunningNoteCollections(t *testing.T) {
	xs := []DunningNote{SampleDunningNote("a", when), SampleDunningNote("bb", when), SampleDunningNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningNotes(xs)); got != 2 {
		t.Errorf("ActiveDunningNotes() kept %d, want 2", got)
	}
	if got, want := SumDunningNotes(xs), 1+3; got != want {
		t.Errorf("SumDunningNotes() = %d, want %d", got, want)
	}
	SortDunningNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningNotes() put %q first", xs[0].ID)
	}
	if len(IndexDunningNotes(xs)) != 3 {
		t.Errorf("IndexDunningNotes() = %v", IndexDunningNotes(xs))
	}
	if err := ValidateDunningNotes(ActiveDunningNotes(xs)); err != nil {
		t.Errorf("ValidateDunningNotes() = %v", err)
	}
}

func TestDunningNoteNormalizeClamps(t *testing.T) {
	x := SampleDunningNote("a", when)
	x.Comment = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningNote("r1", when)
	rev, err := r.PutDunningNote(x, 0)
	if err != nil {
		t.Fatalf("PutDunningNote() = %v", err)
	}
	got, gotRev, err := r.GetDunningNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningNotes() = %v, %v", all, err)
	}
	if _, err := r.PutDunningNote(DunningNote{}, 0); err == nil {
		t.Error("PutDunningNote() accepted an empty entity")
	}
}

func TestDunningWindowRoundTrip(t *testing.T) {
	x := SampleDunningWindow("dunningwindow-1", when)
	got, err := ParseDunningWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningWindow() = %v", err)
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

func TestDunningWindowParseRejectsBadInput(t *testing.T) {
	f := SampleDunningWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningWindow(f); err == nil {
		t.Error("ParseDunningWindow() accepted a record with no id")
	}
	f = SampleDunningWindow("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDunningWindow(f); err == nil {
		t.Error("ParseDunningWindow() accepted a non-numeric count")
	}
}

func TestDunningWindowCollections(t *testing.T) {
	xs := []DunningWindow{SampleDunningWindow("a", when), SampleDunningWindow("bb", when), SampleDunningWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningWindows(xs)); got != 2 {
		t.Errorf("ActiveDunningWindows() kept %d, want 2", got)
	}
	if got, want := SumDunningWindows(xs), 1+3; got != want {
		t.Errorf("SumDunningWindows() = %d, want %d", got, want)
	}
	SortDunningWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningWindows() put %q first", xs[0].ID)
	}
	if len(IndexDunningWindows(xs)) != 3 {
		t.Errorf("IndexDunningWindows() = %v", IndexDunningWindows(xs))
	}
	if err := ValidateDunningWindows(ActiveDunningWindows(xs)); err != nil {
		t.Errorf("ValidateDunningWindows() = %v", err)
	}
}

func TestDunningWindowNormalizeClamps(t *testing.T) {
	x := SampleDunningWindow("a", when)
	x.Source = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningWindow("r1", when)
	rev, err := r.PutDunningWindow(x, 0)
	if err != nil {
		t.Fatalf("PutDunningWindow() = %v", err)
	}
	got, gotRev, err := r.GetDunningWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningWindows() = %v, %v", all, err)
	}
	if _, err := r.PutDunningWindow(DunningWindow{}, 0); err == nil {
		t.Error("PutDunningWindow() accepted an empty entity")
	}
}

func TestDunningQuotaRoundTrip(t *testing.T) {
	x := SampleDunningQuota("dunningquota-1", when)
	got, err := ParseDunningQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningQuota() = %v", err)
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

func TestDunningQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleDunningQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningQuota(f); err == nil {
		t.Error("ParseDunningQuota() accepted a record with no id")
	}
	f = SampleDunningQuota("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDunningQuota(f); err == nil {
		t.Error("ParseDunningQuota() accepted a non-numeric offset")
	}
}

func TestDunningQuotaCollections(t *testing.T) {
	xs := []DunningQuota{SampleDunningQuota("a", when), SampleDunningQuota("bb", when), SampleDunningQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningQuotas(xs)); got != 2 {
		t.Errorf("ActiveDunningQuotas() kept %d, want 2", got)
	}
	if got, want := SumDunningQuotas(xs), 1+3; got != want {
		t.Errorf("SumDunningQuotas() = %d, want %d", got, want)
	}
	SortDunningQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningQuotas() put %q first", xs[0].ID)
	}
	if len(IndexDunningQuotas(xs)) != 3 {
		t.Errorf("IndexDunningQuotas() = %v", IndexDunningQuotas(xs))
	}
	if err := ValidateDunningQuotas(ActiveDunningQuotas(xs)); err != nil {
		t.Errorf("ValidateDunningQuotas() = %v", err)
	}
}

func TestDunningQuotaNormalizeClamps(t *testing.T) {
	x := SampleDunningQuota("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningQuota("r1", when)
	rev, err := r.PutDunningQuota(x, 0)
	if err != nil {
		t.Fatalf("PutDunningQuota() = %v", err)
	}
	got, gotRev, err := r.GetDunningQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutDunningQuota(DunningQuota{}, 0); err == nil {
		t.Error("PutDunningQuota() accepted an empty entity")
	}
}

func TestDunningBucketRoundTrip(t *testing.T) {
	x := SampleDunningBucket("dunningbucket-1", when)
	got, err := ParseDunningBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningBucket() = %v", err)
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

func TestDunningBucketParseRejectsBadInput(t *testing.T) {
	f := SampleDunningBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningBucket(f); err == nil {
		t.Error("ParseDunningBucket() accepted a record with no id")
	}
	f = SampleDunningBucket("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDunningBucket(f); err == nil {
		t.Error("ParseDunningBucket() accepted a non-numeric offset")
	}
}

func TestDunningBucketCollections(t *testing.T) {
	xs := []DunningBucket{SampleDunningBucket("a", when), SampleDunningBucket("bb", when), SampleDunningBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningBuckets(xs)); got != 2 {
		t.Errorf("ActiveDunningBuckets() kept %d, want 2", got)
	}
	if got, want := SumDunningBuckets(xs), 1+3; got != want {
		t.Errorf("SumDunningBuckets() = %d, want %d", got, want)
	}
	SortDunningBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningBuckets() put %q first", xs[0].ID)
	}
	if len(IndexDunningBuckets(xs)) != 3 {
		t.Errorf("IndexDunningBuckets() = %v", IndexDunningBuckets(xs))
	}
	if err := ValidateDunningBuckets(ActiveDunningBuckets(xs)); err != nil {
		t.Errorf("ValidateDunningBuckets() = %v", err)
	}
}

func TestDunningBucketNormalizeClamps(t *testing.T) {
	x := SampleDunningBucket("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningBucket("r1", when)
	rev, err := r.PutDunningBucket(x, 0)
	if err != nil {
		t.Fatalf("PutDunningBucket() = %v", err)
	}
	got, gotRev, err := r.GetDunningBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutDunningBucket(DunningBucket{}, 0); err == nil {
		t.Error("PutDunningBucket() accepted an empty entity")
	}
}

func TestDunningLaneRoundTrip(t *testing.T) {
	x := SampleDunningLane("dunninglane-1", when)
	got, err := ParseDunningLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningLane() = %v", err)
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

func TestDunningLaneParseRejectsBadInput(t *testing.T) {
	f := SampleDunningLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningLane(f); err == nil {
		t.Error("ParseDunningLane() accepted a record with no id")
	}
	f = SampleDunningLane("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDunningLane(f); err == nil {
		t.Error("ParseDunningLane() accepted a non-numeric version")
	}
}

func TestDunningLaneCollections(t *testing.T) {
	xs := []DunningLane{SampleDunningLane("a", when), SampleDunningLane("bb", when), SampleDunningLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningLanes(xs)); got != 2 {
		t.Errorf("ActiveDunningLanes() kept %d, want 2", got)
	}
	if got, want := SumDunningLanes(xs), 1+3; got != want {
		t.Errorf("SumDunningLanes() = %d, want %d", got, want)
	}
	SortDunningLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningLanes() put %q first", xs[0].ID)
	}
	if len(IndexDunningLanes(xs)) != 3 {
		t.Errorf("IndexDunningLanes() = %v", IndexDunningLanes(xs))
	}
	if err := ValidateDunningLanes(ActiveDunningLanes(xs)); err != nil {
		t.Errorf("ValidateDunningLanes() = %v", err)
	}
}

func TestDunningLaneNormalizeClamps(t *testing.T) {
	x := SampleDunningLane("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningLane("r1", when)
	rev, err := r.PutDunningLane(x, 0)
	if err != nil {
		t.Fatalf("PutDunningLane() = %v", err)
	}
	got, gotRev, err := r.GetDunningLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningLanes() = %v, %v", all, err)
	}
	if _, err := r.PutDunningLane(DunningLane{}, 0); err == nil {
		t.Error("PutDunningLane() accepted an empty entity")
	}
}

func TestDunningSlotRoundTrip(t *testing.T) {
	x := SampleDunningSlot("dunningslot-1", when)
	got, err := ParseDunningSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningSlot() = %v", err)
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

func TestDunningSlotParseRejectsBadInput(t *testing.T) {
	f := SampleDunningSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningSlot(f); err == nil {
		t.Error("ParseDunningSlot() accepted a record with no id")
	}
	f = SampleDunningSlot("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDunningSlot(f); err == nil {
		t.Error("ParseDunningSlot() accepted a non-numeric attempts")
	}
}

func TestDunningSlotCollections(t *testing.T) {
	xs := []DunningSlot{SampleDunningSlot("a", when), SampleDunningSlot("bb", when), SampleDunningSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningSlots(xs)); got != 2 {
		t.Errorf("ActiveDunningSlots() kept %d, want 2", got)
	}
	if got, want := SumDunningSlots(xs), 1+3; got != want {
		t.Errorf("SumDunningSlots() = %d, want %d", got, want)
	}
	SortDunningSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningSlots() put %q first", xs[0].ID)
	}
	if len(IndexDunningSlots(xs)) != 3 {
		t.Errorf("IndexDunningSlots() = %v", IndexDunningSlots(xs))
	}
	if err := ValidateDunningSlots(ActiveDunningSlots(xs)); err != nil {
		t.Errorf("ValidateDunningSlots() = %v", err)
	}
}

func TestDunningSlotNormalizeClamps(t *testing.T) {
	x := SampleDunningSlot("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningSlot("r1", when)
	rev, err := r.PutDunningSlot(x, 0)
	if err != nil {
		t.Fatalf("PutDunningSlot() = %v", err)
	}
	got, gotRev, err := r.GetDunningSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningSlots() = %v, %v", all, err)
	}
	if _, err := r.PutDunningSlot(DunningSlot{}, 0); err == nil {
		t.Error("PutDunningSlot() accepted an empty entity")
	}
}

func TestDunningTagRoundTrip(t *testing.T) {
	x := SampleDunningTag("dunningtag-1", when)
	got, err := ParseDunningTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningTag() = %v", err)
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

func TestDunningTagParseRejectsBadInput(t *testing.T) {
	f := SampleDunningTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningTag(f); err == nil {
		t.Error("ParseDunningTag() accepted a record with no id")
	}
	f = SampleDunningTag("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDunningTag(f); err == nil {
		t.Error("ParseDunningTag() accepted a non-numeric priority")
	}
}

func TestDunningTagCollections(t *testing.T) {
	xs := []DunningTag{SampleDunningTag("a", when), SampleDunningTag("bb", when), SampleDunningTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningTags(xs)); got != 2 {
		t.Errorf("ActiveDunningTags() kept %d, want 2", got)
	}
	if got, want := SumDunningTags(xs), 1+3; got != want {
		t.Errorf("SumDunningTags() = %d, want %d", got, want)
	}
	SortDunningTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningTags() put %q first", xs[0].ID)
	}
	if len(IndexDunningTags(xs)) != 3 {
		t.Errorf("IndexDunningTags() = %v", IndexDunningTags(xs))
	}
	if err := ValidateDunningTags(ActiveDunningTags(xs)); err != nil {
		t.Errorf("ValidateDunningTags() = %v", err)
	}
}

func TestDunningTagNormalizeClamps(t *testing.T) {
	x := SampleDunningTag("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningTag("r1", when)
	rev, err := r.PutDunningTag(x, 0)
	if err != nil {
		t.Fatalf("PutDunningTag() = %v", err)
	}
	got, gotRev, err := r.GetDunningTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningTags() = %v, %v", all, err)
	}
	if _, err := r.PutDunningTag(DunningTag{}, 0); err == nil {
		t.Error("PutDunningTag() accepted an empty entity")
	}
}

func TestDunningRuleRoundTrip(t *testing.T) {
	x := SampleDunningRule("dunningrule-1", when)
	got, err := ParseDunningRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningRule() = %v", err)
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

func TestDunningRuleParseRejectsBadInput(t *testing.T) {
	f := SampleDunningRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningRule(f); err == nil {
		t.Error("ParseDunningRule() accepted a record with no id")
	}
	f = SampleDunningRule("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDunningRule(f); err == nil {
		t.Error("ParseDunningRule() accepted a non-numeric count")
	}
}

func TestDunningRuleCollections(t *testing.T) {
	xs := []DunningRule{SampleDunningRule("a", when), SampleDunningRule("bb", when), SampleDunningRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningRules(xs)); got != 2 {
		t.Errorf("ActiveDunningRules() kept %d, want 2", got)
	}
	if got, want := SumDunningRules(xs), 1+3; got != want {
		t.Errorf("SumDunningRules() = %d, want %d", got, want)
	}
	SortDunningRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningRules() put %q first", xs[0].ID)
	}
	if len(IndexDunningRules(xs)) != 3 {
		t.Errorf("IndexDunningRules() = %v", IndexDunningRules(xs))
	}
	if err := ValidateDunningRules(ActiveDunningRules(xs)); err != nil {
		t.Errorf("ValidateDunningRules() = %v", err)
	}
}

func TestDunningRuleNormalizeClamps(t *testing.T) {
	x := SampleDunningRule("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningRule("r1", when)
	rev, err := r.PutDunningRule(x, 0)
	if err != nil {
		t.Fatalf("PutDunningRule() = %v", err)
	}
	got, gotRev, err := r.GetDunningRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningRules() = %v, %v", all, err)
	}
	if _, err := r.PutDunningRule(DunningRule{}, 0); err == nil {
		t.Error("PutDunningRule() accepted an empty entity")
	}
}

func TestDunningHoldRoundTrip(t *testing.T) {
	x := SampleDunningHold("dunninghold-1", when)
	got, err := ParseDunningHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningHold() = %v", err)
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

func TestDunningHoldParseRejectsBadInput(t *testing.T) {
	f := SampleDunningHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningHold(f); err == nil {
		t.Error("ParseDunningHold() accepted a record with no id")
	}
	f = SampleDunningHold("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDunningHold(f); err == nil {
		t.Error("ParseDunningHold() accepted a non-numeric offset")
	}
}

func TestDunningHoldCollections(t *testing.T) {
	xs := []DunningHold{SampleDunningHold("a", when), SampleDunningHold("bb", when), SampleDunningHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningHolds(xs)); got != 2 {
		t.Errorf("ActiveDunningHolds() kept %d, want 2", got)
	}
	if got, want := SumDunningHolds(xs), 1+3; got != want {
		t.Errorf("SumDunningHolds() = %d, want %d", got, want)
	}
	SortDunningHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningHolds() put %q first", xs[0].ID)
	}
	if len(IndexDunningHolds(xs)) != 3 {
		t.Errorf("IndexDunningHolds() = %v", IndexDunningHolds(xs))
	}
	if err := ValidateDunningHolds(ActiveDunningHolds(xs)); err != nil {
		t.Errorf("ValidateDunningHolds() = %v", err)
	}
}

func TestDunningHoldNormalizeClamps(t *testing.T) {
	x := SampleDunningHold("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningHold("r1", when)
	rev, err := r.PutDunningHold(x, 0)
	if err != nil {
		t.Fatalf("PutDunningHold() = %v", err)
	}
	got, gotRev, err := r.GetDunningHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningHolds() = %v, %v", all, err)
	}
	if _, err := r.PutDunningHold(DunningHold{}, 0); err == nil {
		t.Error("PutDunningHold() accepted an empty entity")
	}
}

func TestDunningClaimRoundTrip(t *testing.T) {
	x := SampleDunningClaim("dunningclaim-1", when)
	got, err := ParseDunningClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningClaim() = %v", err)
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

func TestDunningClaimParseRejectsBadInput(t *testing.T) {
	f := SampleDunningClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningClaim(f); err == nil {
		t.Error("ParseDunningClaim() accepted a record with no id")
	}
	f = SampleDunningClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDunningClaim(f); err == nil {
		t.Error("ParseDunningClaim() accepted a non-numeric retained")
	}
}

func TestDunningClaimCollections(t *testing.T) {
	xs := []DunningClaim{SampleDunningClaim("a", when), SampleDunningClaim("bb", when), SampleDunningClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningClaims(xs)); got != 2 {
		t.Errorf("ActiveDunningClaims() kept %d, want 2", got)
	}
	if got, want := SumDunningClaims(xs), 1+3; got != want {
		t.Errorf("SumDunningClaims() = %d, want %d", got, want)
	}
	SortDunningClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningClaims() put %q first", xs[0].ID)
	}
	if len(IndexDunningClaims(xs)) != 3 {
		t.Errorf("IndexDunningClaims() = %v", IndexDunningClaims(xs))
	}
	if err := ValidateDunningClaims(ActiveDunningClaims(xs)); err != nil {
		t.Errorf("ValidateDunningClaims() = %v", err)
	}
}

func TestDunningClaimNormalizeClamps(t *testing.T) {
	x := SampleDunningClaim("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningClaim("r1", when)
	rev, err := r.PutDunningClaim(x, 0)
	if err != nil {
		t.Fatalf("PutDunningClaim() = %v", err)
	}
	got, gotRev, err := r.GetDunningClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningClaims() = %v, %v", all, err)
	}
	if _, err := r.PutDunningClaim(DunningClaim{}, 0); err == nil {
		t.Error("PutDunningClaim() accepted an empty entity")
	}
}

func TestDunningCursorRoundTrip(t *testing.T) {
	x := SampleDunningCursor("dunningcursor-1", when)
	got, err := ParseDunningCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseDunningCursor() = %v", err)
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

func TestDunningCursorParseRejectsBadInput(t *testing.T) {
	f := SampleDunningCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDunningCursor(f); err == nil {
		t.Error("ParseDunningCursor() accepted a record with no id")
	}
	f = SampleDunningCursor("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDunningCursor(f); err == nil {
		t.Error("ParseDunningCursor() accepted a non-numeric retained")
	}
}

func TestDunningCursorCollections(t *testing.T) {
	xs := []DunningCursor{SampleDunningCursor("a", when), SampleDunningCursor("bb", when), SampleDunningCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDunningCursors(xs)); got != 2 {
		t.Errorf("ActiveDunningCursors() kept %d, want 2", got)
	}
	if got, want := SumDunningCursors(xs), 1+3; got != want {
		t.Errorf("SumDunningCursors() = %d, want %d", got, want)
	}
	SortDunningCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDunningCursors() put %q first", xs[0].ID)
	}
	if len(IndexDunningCursors(xs)) != 3 {
		t.Errorf("IndexDunningCursors() = %v", IndexDunningCursors(xs))
	}
	if err := ValidateDunningCursors(ActiveDunningCursors(xs)); err != nil {
		t.Errorf("ValidateDunningCursors() = %v", err)
	}
}

func TestDunningCursorNormalizeClamps(t *testing.T) {
	x := SampleDunningCursor("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDunningCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDunningCursor("r1", when)
	rev, err := r.PutDunningCursor(x, 0)
	if err != nil {
		t.Fatalf("PutDunningCursor() = %v", err)
	}
	got, gotRev, err := r.GetDunningCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDunningCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDunningCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDunningCursors() = %v, %v", all, err)
	}
	if _, err := r.PutDunningCursor(DunningCursor{}, 0); err == nil {
		t.Error("PutDunningCursor() accepted an empty entity")
	}
}
