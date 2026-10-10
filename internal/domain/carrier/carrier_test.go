package carrier

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
		if !strings.HasPrefix(Prefixes[i], "carrier/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleCarrierRecord("ok", when).Fields()); err != nil {
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

func TestCarrierRecordRoundTrip(t *testing.T) {
	x := SampleCarrierRecord("carrierrecord-1", when)
	got, err := ParseCarrierRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierRecord() = %v", err)
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

func TestCarrierRecordParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierRecord(f); err == nil {
		t.Error("ParseCarrierRecord() accepted a record with no id")
	}
	f = SampleCarrierRecord("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseCarrierRecord(f); err == nil {
		t.Error("ParseCarrierRecord() accepted a non-numeric priority")
	}
}

func TestCarrierRecordCollections(t *testing.T) {
	xs := []CarrierRecord{SampleCarrierRecord("a", when), SampleCarrierRecord("bb", when), SampleCarrierRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierRecords(xs)); got != 2 {
		t.Errorf("ActiveCarrierRecords() kept %d, want 2", got)
	}
	if got, want := SumCarrierRecords(xs), 1+3; got != want {
		t.Errorf("SumCarrierRecords() = %d, want %d", got, want)
	}
	SortCarrierRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierRecords() put %q first", xs[0].ID)
	}
	if len(IndexCarrierRecords(xs)) != 3 {
		t.Errorf("IndexCarrierRecords() = %v", IndexCarrierRecords(xs))
	}
	if err := ValidateCarrierRecords(ActiveCarrierRecords(xs)); err != nil {
		t.Errorf("ValidateCarrierRecords() = %v", err)
	}
}

func TestCarrierRecordNormalizeClamps(t *testing.T) {
	x := SampleCarrierRecord("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierRecord("r1", when)
	rev, err := r.PutCarrierRecord(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierRecord() = %v", err)
	}
	got, gotRev, err := r.GetCarrierRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierRecords() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierRecord(CarrierRecord{}, 0); err == nil {
		t.Error("PutCarrierRecord() accepted an empty entity")
	}
}

func TestCarrierEntryRoundTrip(t *testing.T) {
	x := SampleCarrierEntry("carrierentry-1", when)
	got, err := ParseCarrierEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierEntry() = %v", err)
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

func TestCarrierEntryParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierEntry(f); err == nil {
		t.Error("ParseCarrierEntry() accepted a record with no id")
	}
	f = SampleCarrierEntry("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCarrierEntry(f); err == nil {
		t.Error("ParseCarrierEntry() accepted a non-numeric capacity")
	}
}

func TestCarrierEntryCollections(t *testing.T) {
	xs := []CarrierEntry{SampleCarrierEntry("a", when), SampleCarrierEntry("bb", when), SampleCarrierEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierEntrys(xs)); got != 2 {
		t.Errorf("ActiveCarrierEntrys() kept %d, want 2", got)
	}
	if got, want := SumCarrierEntrys(xs), 1+3; got != want {
		t.Errorf("SumCarrierEntrys() = %d, want %d", got, want)
	}
	SortCarrierEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierEntrys() put %q first", xs[0].ID)
	}
	if len(IndexCarrierEntrys(xs)) != 3 {
		t.Errorf("IndexCarrierEntrys() = %v", IndexCarrierEntrys(xs))
	}
	if err := ValidateCarrierEntrys(ActiveCarrierEntrys(xs)); err != nil {
		t.Errorf("ValidateCarrierEntrys() = %v", err)
	}
}

func TestCarrierEntryNormalizeClamps(t *testing.T) {
	x := SampleCarrierEntry("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierEntry("r1", when)
	rev, err := r.PutCarrierEntry(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierEntry() = %v", err)
	}
	got, gotRev, err := r.GetCarrierEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierEntry(CarrierEntry{}, 0); err == nil {
		t.Error("PutCarrierEntry() accepted an empty entity")
	}
}

func TestCarrierItemRoundTrip(t *testing.T) {
	x := SampleCarrierItem("carrieritem-1", when)
	got, err := ParseCarrierItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierItem() = %v", err)
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

func TestCarrierItemParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierItem(f); err == nil {
		t.Error("ParseCarrierItem() accepted a record with no id")
	}
	f = SampleCarrierItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseCarrierItem(f); err == nil {
		t.Error("ParseCarrierItem() accepted a non-numeric version")
	}
}

func TestCarrierItemCollections(t *testing.T) {
	xs := []CarrierItem{SampleCarrierItem("a", when), SampleCarrierItem("bb", when), SampleCarrierItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierItems(xs)); got != 2 {
		t.Errorf("ActiveCarrierItems() kept %d, want 2", got)
	}
	if got, want := SumCarrierItems(xs), 1+3; got != want {
		t.Errorf("SumCarrierItems() = %d, want %d", got, want)
	}
	SortCarrierItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierItems() put %q first", xs[0].ID)
	}
	if len(IndexCarrierItems(xs)) != 3 {
		t.Errorf("IndexCarrierItems() = %v", IndexCarrierItems(xs))
	}
	if err := ValidateCarrierItems(ActiveCarrierItems(xs)); err != nil {
		t.Errorf("ValidateCarrierItems() = %v", err)
	}
}

func TestCarrierItemNormalizeClamps(t *testing.T) {
	x := SampleCarrierItem("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierItem("r1", when)
	rev, err := r.PutCarrierItem(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierItem() = %v", err)
	}
	got, gotRev, err := r.GetCarrierItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierItems() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierItem(CarrierItem{}, 0); err == nil {
		t.Error("PutCarrierItem() accepted an empty entity")
	}
}

func TestCarrierDraftRoundTrip(t *testing.T) {
	x := SampleCarrierDraft("carrierdraft-1", when)
	got, err := ParseCarrierDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierDraft() = %v", err)
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

func TestCarrierDraftParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierDraft(f); err == nil {
		t.Error("ParseCarrierDraft() accepted a record with no id")
	}
	f = SampleCarrierDraft("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseCarrierDraft(f); err == nil {
		t.Error("ParseCarrierDraft() accepted a non-numeric version")
	}
}

func TestCarrierDraftCollections(t *testing.T) {
	xs := []CarrierDraft{SampleCarrierDraft("a", when), SampleCarrierDraft("bb", when), SampleCarrierDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierDrafts(xs)); got != 2 {
		t.Errorf("ActiveCarrierDrafts() kept %d, want 2", got)
	}
	if got, want := SumCarrierDrafts(xs), 1+3; got != want {
		t.Errorf("SumCarrierDrafts() = %d, want %d", got, want)
	}
	SortCarrierDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierDrafts() put %q first", xs[0].ID)
	}
	if len(IndexCarrierDrafts(xs)) != 3 {
		t.Errorf("IndexCarrierDrafts() = %v", IndexCarrierDrafts(xs))
	}
	if err := ValidateCarrierDrafts(ActiveCarrierDrafts(xs)); err != nil {
		t.Errorf("ValidateCarrierDrafts() = %v", err)
	}
}

func TestCarrierDraftNormalizeClamps(t *testing.T) {
	x := SampleCarrierDraft("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierDraft("r1", when)
	rev, err := r.PutCarrierDraft(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierDraft() = %v", err)
	}
	got, gotRev, err := r.GetCarrierDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierDraft(CarrierDraft{}, 0); err == nil {
		t.Error("PutCarrierDraft() accepted an empty entity")
	}
}

func TestCarrierRevisionRoundTrip(t *testing.T) {
	x := SampleCarrierRevision("carrierrevision-1", when)
	got, err := ParseCarrierRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierRevision() = %v", err)
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

func TestCarrierRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierRevision(f); err == nil {
		t.Error("ParseCarrierRevision() accepted a record with no id")
	}
	f = SampleCarrierRevision("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCarrierRevision(f); err == nil {
		t.Error("ParseCarrierRevision() accepted a non-numeric count")
	}
}

func TestCarrierRevisionCollections(t *testing.T) {
	xs := []CarrierRevision{SampleCarrierRevision("a", when), SampleCarrierRevision("bb", when), SampleCarrierRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierRevisions(xs)); got != 2 {
		t.Errorf("ActiveCarrierRevisions() kept %d, want 2", got)
	}
	if got, want := SumCarrierRevisions(xs), 1+3; got != want {
		t.Errorf("SumCarrierRevisions() = %d, want %d", got, want)
	}
	SortCarrierRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierRevisions() put %q first", xs[0].ID)
	}
	if len(IndexCarrierRevisions(xs)) != 3 {
		t.Errorf("IndexCarrierRevisions() = %v", IndexCarrierRevisions(xs))
	}
	if err := ValidateCarrierRevisions(ActiveCarrierRevisions(xs)); err != nil {
		t.Errorf("ValidateCarrierRevisions() = %v", err)
	}
}

func TestCarrierRevisionNormalizeClamps(t *testing.T) {
	x := SampleCarrierRevision("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierRevision("r1", when)
	rev, err := r.PutCarrierRevision(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierRevision() = %v", err)
	}
	got, gotRev, err := r.GetCarrierRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierRevision(CarrierRevision{}, 0); err == nil {
		t.Error("PutCarrierRevision() accepted an empty entity")
	}
}

func TestCarrierSnapshotRoundTrip(t *testing.T) {
	x := SampleCarrierSnapshot("carriersnapshot-1", when)
	got, err := ParseCarrierSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierSnapshot() = %v", err)
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

func TestCarrierSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierSnapshot(f); err == nil {
		t.Error("ParseCarrierSnapshot() accepted a record with no id")
	}
	f = SampleCarrierSnapshot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCarrierSnapshot(f); err == nil {
		t.Error("ParseCarrierSnapshot() accepted a non-numeric count")
	}
}

func TestCarrierSnapshotCollections(t *testing.T) {
	xs := []CarrierSnapshot{SampleCarrierSnapshot("a", when), SampleCarrierSnapshot("bb", when), SampleCarrierSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierSnapshots(xs)); got != 2 {
		t.Errorf("ActiveCarrierSnapshots() kept %d, want 2", got)
	}
	if got, want := SumCarrierSnapshots(xs), 1+3; got != want {
		t.Errorf("SumCarrierSnapshots() = %d, want %d", got, want)
	}
	SortCarrierSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexCarrierSnapshots(xs)) != 3 {
		t.Errorf("IndexCarrierSnapshots() = %v", IndexCarrierSnapshots(xs))
	}
	if err := ValidateCarrierSnapshots(ActiveCarrierSnapshots(xs)); err != nil {
		t.Errorf("ValidateCarrierSnapshots() = %v", err)
	}
}

func TestCarrierSnapshotNormalizeClamps(t *testing.T) {
	x := SampleCarrierSnapshot("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierSnapshot("r1", when)
	rev, err := r.PutCarrierSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetCarrierSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierSnapshot(CarrierSnapshot{}, 0); err == nil {
		t.Error("PutCarrierSnapshot() accepted an empty entity")
	}
}

func TestCarrierBatchRoundTrip(t *testing.T) {
	x := SampleCarrierBatch("carrierbatch-1", when)
	got, err := ParseCarrierBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierBatch() = %v", err)
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

func TestCarrierBatchParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierBatch(f); err == nil {
		t.Error("ParseCarrierBatch() accepted a record with no id")
	}
	f = SampleCarrierBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCarrierBatch(f); err == nil {
		t.Error("ParseCarrierBatch() accepted a non-numeric capacity")
	}
}

func TestCarrierBatchCollections(t *testing.T) {
	xs := []CarrierBatch{SampleCarrierBatch("a", when), SampleCarrierBatch("bb", when), SampleCarrierBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierBatchs(xs)); got != 2 {
		t.Errorf("ActiveCarrierBatchs() kept %d, want 2", got)
	}
	if got, want := SumCarrierBatchs(xs), 1+3; got != want {
		t.Errorf("SumCarrierBatchs() = %d, want %d", got, want)
	}
	SortCarrierBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierBatchs() put %q first", xs[0].ID)
	}
	if len(IndexCarrierBatchs(xs)) != 3 {
		t.Errorf("IndexCarrierBatchs() = %v", IndexCarrierBatchs(xs))
	}
	if err := ValidateCarrierBatchs(ActiveCarrierBatchs(xs)); err != nil {
		t.Errorf("ValidateCarrierBatchs() = %v", err)
	}
}

func TestCarrierBatchNormalizeClamps(t *testing.T) {
	x := SampleCarrierBatch("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierBatch("r1", when)
	rev, err := r.PutCarrierBatch(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierBatch() = %v", err)
	}
	got, gotRev, err := r.GetCarrierBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierBatch(CarrierBatch{}, 0); err == nil {
		t.Error("PutCarrierBatch() accepted an empty entity")
	}
}

func TestCarrierRequestRoundTrip(t *testing.T) {
	x := SampleCarrierRequest("carrierrequest-1", when)
	got, err := ParseCarrierRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierRequest() = %v", err)
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

func TestCarrierRequestParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierRequest(f); err == nil {
		t.Error("ParseCarrierRequest() accepted a record with no id")
	}
	f = SampleCarrierRequest("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseCarrierRequest(f); err == nil {
		t.Error("ParseCarrierRequest() accepted a non-numeric version")
	}
}

func TestCarrierRequestCollections(t *testing.T) {
	xs := []CarrierRequest{SampleCarrierRequest("a", when), SampleCarrierRequest("bb", when), SampleCarrierRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierRequests(xs)); got != 2 {
		t.Errorf("ActiveCarrierRequests() kept %d, want 2", got)
	}
	if got, want := SumCarrierRequests(xs), 1+3; got != want {
		t.Errorf("SumCarrierRequests() = %d, want %d", got, want)
	}
	SortCarrierRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierRequests() put %q first", xs[0].ID)
	}
	if len(IndexCarrierRequests(xs)) != 3 {
		t.Errorf("IndexCarrierRequests() = %v", IndexCarrierRequests(xs))
	}
	if err := ValidateCarrierRequests(ActiveCarrierRequests(xs)); err != nil {
		t.Errorf("ValidateCarrierRequests() = %v", err)
	}
}

func TestCarrierRequestNormalizeClamps(t *testing.T) {
	x := SampleCarrierRequest("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierRequest("r1", when)
	rev, err := r.PutCarrierRequest(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierRequest() = %v", err)
	}
	got, gotRev, err := r.GetCarrierRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierRequests() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierRequest(CarrierRequest{}, 0); err == nil {
		t.Error("PutCarrierRequest() accepted an empty entity")
	}
}

func TestCarrierSummaryRoundTrip(t *testing.T) {
	x := SampleCarrierSummary("carriersummary-1", when)
	got, err := ParseCarrierSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierSummary() = %v", err)
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

func TestCarrierSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierSummary(f); err == nil {
		t.Error("ParseCarrierSummary() accepted a record with no id")
	}
	f = SampleCarrierSummary("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseCarrierSummary(f); err == nil {
		t.Error("ParseCarrierSummary() accepted a non-numeric offset")
	}
}

func TestCarrierSummaryCollections(t *testing.T) {
	xs := []CarrierSummary{SampleCarrierSummary("a", when), SampleCarrierSummary("bb", when), SampleCarrierSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierSummarys(xs)); got != 2 {
		t.Errorf("ActiveCarrierSummarys() kept %d, want 2", got)
	}
	if got, want := SumCarrierSummarys(xs), 1+3; got != want {
		t.Errorf("SumCarrierSummarys() = %d, want %d", got, want)
	}
	SortCarrierSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierSummarys() put %q first", xs[0].ID)
	}
	if len(IndexCarrierSummarys(xs)) != 3 {
		t.Errorf("IndexCarrierSummarys() = %v", IndexCarrierSummarys(xs))
	}
	if err := ValidateCarrierSummarys(ActiveCarrierSummarys(xs)); err != nil {
		t.Errorf("ValidateCarrierSummarys() = %v", err)
	}
}

func TestCarrierSummaryNormalizeClamps(t *testing.T) {
	x := SampleCarrierSummary("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierSummary("r1", when)
	rev, err := r.PutCarrierSummary(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierSummary() = %v", err)
	}
	got, gotRev, err := r.GetCarrierSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierSummary(CarrierSummary{}, 0); err == nil {
		t.Error("PutCarrierSummary() accepted an empty entity")
	}
}

func TestCarrierPolicyRoundTrip(t *testing.T) {
	x := SampleCarrierPolicy("carrierpolicy-1", when)
	got, err := ParseCarrierPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierPolicy() = %v", err)
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

func TestCarrierPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierPolicy(f); err == nil {
		t.Error("ParseCarrierPolicy() accepted a record with no id")
	}
	f = SampleCarrierPolicy("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseCarrierPolicy(f); err == nil {
		t.Error("ParseCarrierPolicy() accepted a non-numeric weight")
	}
}

func TestCarrierPolicyCollections(t *testing.T) {
	xs := []CarrierPolicy{SampleCarrierPolicy("a", when), SampleCarrierPolicy("bb", when), SampleCarrierPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierPolicys(xs)); got != 2 {
		t.Errorf("ActiveCarrierPolicys() kept %d, want 2", got)
	}
	if got, want := SumCarrierPolicys(xs), 1+3; got != want {
		t.Errorf("SumCarrierPolicys() = %d, want %d", got, want)
	}
	SortCarrierPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierPolicys() put %q first", xs[0].ID)
	}
	if len(IndexCarrierPolicys(xs)) != 3 {
		t.Errorf("IndexCarrierPolicys() = %v", IndexCarrierPolicys(xs))
	}
	if err := ValidateCarrierPolicys(ActiveCarrierPolicys(xs)); err != nil {
		t.Errorf("ValidateCarrierPolicys() = %v", err)
	}
}

func TestCarrierPolicyNormalizeClamps(t *testing.T) {
	x := SampleCarrierPolicy("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierPolicy("r1", when)
	rev, err := r.PutCarrierPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierPolicy() = %v", err)
	}
	got, gotRev, err := r.GetCarrierPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierPolicy(CarrierPolicy{}, 0); err == nil {
		t.Error("PutCarrierPolicy() accepted an empty entity")
	}
}

func TestCarrierLinkRoundTrip(t *testing.T) {
	x := SampleCarrierLink("carrierlink-1", when)
	got, err := ParseCarrierLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierLink() = %v", err)
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

func TestCarrierLinkParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierLink(f); err == nil {
		t.Error("ParseCarrierLink() accepted a record with no id")
	}
	f = SampleCarrierLink("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseCarrierLink(f); err == nil {
		t.Error("ParseCarrierLink() accepted a non-numeric priority")
	}
}

func TestCarrierLinkCollections(t *testing.T) {
	xs := []CarrierLink{SampleCarrierLink("a", when), SampleCarrierLink("bb", when), SampleCarrierLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierLinks(xs)); got != 2 {
		t.Errorf("ActiveCarrierLinks() kept %d, want 2", got)
	}
	if got, want := SumCarrierLinks(xs), 1+3; got != want {
		t.Errorf("SumCarrierLinks() = %d, want %d", got, want)
	}
	SortCarrierLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierLinks() put %q first", xs[0].ID)
	}
	if len(IndexCarrierLinks(xs)) != 3 {
		t.Errorf("IndexCarrierLinks() = %v", IndexCarrierLinks(xs))
	}
	if err := ValidateCarrierLinks(ActiveCarrierLinks(xs)); err != nil {
		t.Errorf("ValidateCarrierLinks() = %v", err)
	}
}

func TestCarrierLinkNormalizeClamps(t *testing.T) {
	x := SampleCarrierLink("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierLink("r1", when)
	rev, err := r.PutCarrierLink(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierLink() = %v", err)
	}
	got, gotRev, err := r.GetCarrierLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierLinks() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierLink(CarrierLink{}, 0); err == nil {
		t.Error("PutCarrierLink() accepted an empty entity")
	}
}

func TestCarrierNoteRoundTrip(t *testing.T) {
	x := SampleCarrierNote("carriernote-1", when)
	got, err := ParseCarrierNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierNote() = %v", err)
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

func TestCarrierNoteParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierNote(f); err == nil {
		t.Error("ParseCarrierNote() accepted a record with no id")
	}
	f = SampleCarrierNote("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseCarrierNote(f); err == nil {
		t.Error("ParseCarrierNote() accepted a non-numeric priority")
	}
}

func TestCarrierNoteCollections(t *testing.T) {
	xs := []CarrierNote{SampleCarrierNote("a", when), SampleCarrierNote("bb", when), SampleCarrierNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierNotes(xs)); got != 2 {
		t.Errorf("ActiveCarrierNotes() kept %d, want 2", got)
	}
	if got, want := SumCarrierNotes(xs), 1+3; got != want {
		t.Errorf("SumCarrierNotes() = %d, want %d", got, want)
	}
	SortCarrierNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierNotes() put %q first", xs[0].ID)
	}
	if len(IndexCarrierNotes(xs)) != 3 {
		t.Errorf("IndexCarrierNotes() = %v", IndexCarrierNotes(xs))
	}
	if err := ValidateCarrierNotes(ActiveCarrierNotes(xs)); err != nil {
		t.Errorf("ValidateCarrierNotes() = %v", err)
	}
}

func TestCarrierNoteNormalizeClamps(t *testing.T) {
	x := SampleCarrierNote("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierNote("r1", when)
	rev, err := r.PutCarrierNote(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierNote() = %v", err)
	}
	got, gotRev, err := r.GetCarrierNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierNotes() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierNote(CarrierNote{}, 0); err == nil {
		t.Error("PutCarrierNote() accepted an empty entity")
	}
}

func TestCarrierWindowRoundTrip(t *testing.T) {
	x := SampleCarrierWindow("carrierwindow-1", when)
	got, err := ParseCarrierWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierWindow() = %v", err)
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

func TestCarrierWindowParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierWindow(f); err == nil {
		t.Error("ParseCarrierWindow() accepted a record with no id")
	}
	f = SampleCarrierWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseCarrierWindow(f); err == nil {
		t.Error("ParseCarrierWindow() accepted a non-numeric priority")
	}
}

func TestCarrierWindowCollections(t *testing.T) {
	xs := []CarrierWindow{SampleCarrierWindow("a", when), SampleCarrierWindow("bb", when), SampleCarrierWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierWindows(xs)); got != 2 {
		t.Errorf("ActiveCarrierWindows() kept %d, want 2", got)
	}
	if got, want := SumCarrierWindows(xs), 1+3; got != want {
		t.Errorf("SumCarrierWindows() = %d, want %d", got, want)
	}
	SortCarrierWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierWindows() put %q first", xs[0].ID)
	}
	if len(IndexCarrierWindows(xs)) != 3 {
		t.Errorf("IndexCarrierWindows() = %v", IndexCarrierWindows(xs))
	}
	if err := ValidateCarrierWindows(ActiveCarrierWindows(xs)); err != nil {
		t.Errorf("ValidateCarrierWindows() = %v", err)
	}
}

func TestCarrierWindowNormalizeClamps(t *testing.T) {
	x := SampleCarrierWindow("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierWindow("r1", when)
	rev, err := r.PutCarrierWindow(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierWindow() = %v", err)
	}
	got, gotRev, err := r.GetCarrierWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierWindows() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierWindow(CarrierWindow{}, 0); err == nil {
		t.Error("PutCarrierWindow() accepted an empty entity")
	}
}

func TestCarrierQuotaRoundTrip(t *testing.T) {
	x := SampleCarrierQuota("carrierquota-1", when)
	got, err := ParseCarrierQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierQuota() = %v", err)
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

func TestCarrierQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierQuota(f); err == nil {
		t.Error("ParseCarrierQuota() accepted a record with no id")
	}
	f = SampleCarrierQuota("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseCarrierQuota(f); err == nil {
		t.Error("ParseCarrierQuota() accepted a non-numeric offset")
	}
}

func TestCarrierQuotaCollections(t *testing.T) {
	xs := []CarrierQuota{SampleCarrierQuota("a", when), SampleCarrierQuota("bb", when), SampleCarrierQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierQuotas(xs)); got != 2 {
		t.Errorf("ActiveCarrierQuotas() kept %d, want 2", got)
	}
	if got, want := SumCarrierQuotas(xs), 1+3; got != want {
		t.Errorf("SumCarrierQuotas() = %d, want %d", got, want)
	}
	SortCarrierQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierQuotas() put %q first", xs[0].ID)
	}
	if len(IndexCarrierQuotas(xs)) != 3 {
		t.Errorf("IndexCarrierQuotas() = %v", IndexCarrierQuotas(xs))
	}
	if err := ValidateCarrierQuotas(ActiveCarrierQuotas(xs)); err != nil {
		t.Errorf("ValidateCarrierQuotas() = %v", err)
	}
}

func TestCarrierQuotaNormalizeClamps(t *testing.T) {
	x := SampleCarrierQuota("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierQuota("r1", when)
	rev, err := r.PutCarrierQuota(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierQuota() = %v", err)
	}
	got, gotRev, err := r.GetCarrierQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierQuota(CarrierQuota{}, 0); err == nil {
		t.Error("PutCarrierQuota() accepted an empty entity")
	}
}

func TestCarrierBucketRoundTrip(t *testing.T) {
	x := SampleCarrierBucket("carrierbucket-1", when)
	got, err := ParseCarrierBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierBucket() = %v", err)
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

func TestCarrierBucketParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierBucket(f); err == nil {
		t.Error("ParseCarrierBucket() accepted a record with no id")
	}
	f = SampleCarrierBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseCarrierBucket(f); err == nil {
		t.Error("ParseCarrierBucket() accepted a non-numeric retained")
	}
}

func TestCarrierBucketCollections(t *testing.T) {
	xs := []CarrierBucket{SampleCarrierBucket("a", when), SampleCarrierBucket("bb", when), SampleCarrierBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierBuckets(xs)); got != 2 {
		t.Errorf("ActiveCarrierBuckets() kept %d, want 2", got)
	}
	if got, want := SumCarrierBuckets(xs), 1+3; got != want {
		t.Errorf("SumCarrierBuckets() = %d, want %d", got, want)
	}
	SortCarrierBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierBuckets() put %q first", xs[0].ID)
	}
	if len(IndexCarrierBuckets(xs)) != 3 {
		t.Errorf("IndexCarrierBuckets() = %v", IndexCarrierBuckets(xs))
	}
	if err := ValidateCarrierBuckets(ActiveCarrierBuckets(xs)); err != nil {
		t.Errorf("ValidateCarrierBuckets() = %v", err)
	}
}

func TestCarrierBucketNormalizeClamps(t *testing.T) {
	x := SampleCarrierBucket("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierBucket("r1", when)
	rev, err := r.PutCarrierBucket(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierBucket() = %v", err)
	}
	got, gotRev, err := r.GetCarrierBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierBucket(CarrierBucket{}, 0); err == nil {
		t.Error("PutCarrierBucket() accepted an empty entity")
	}
}

func TestCarrierLaneRoundTrip(t *testing.T) {
	x := SampleCarrierLane("carrierlane-1", when)
	got, err := ParseCarrierLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierLane() = %v", err)
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

func TestCarrierLaneParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierLane(f); err == nil {
		t.Error("ParseCarrierLane() accepted a record with no id")
	}
	f = SampleCarrierLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCarrierLane(f); err == nil {
		t.Error("ParseCarrierLane() accepted a non-numeric capacity")
	}
}

func TestCarrierLaneCollections(t *testing.T) {
	xs := []CarrierLane{SampleCarrierLane("a", when), SampleCarrierLane("bb", when), SampleCarrierLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierLanes(xs)); got != 2 {
		t.Errorf("ActiveCarrierLanes() kept %d, want 2", got)
	}
	if got, want := SumCarrierLanes(xs), 1+3; got != want {
		t.Errorf("SumCarrierLanes() = %d, want %d", got, want)
	}
	SortCarrierLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierLanes() put %q first", xs[0].ID)
	}
	if len(IndexCarrierLanes(xs)) != 3 {
		t.Errorf("IndexCarrierLanes() = %v", IndexCarrierLanes(xs))
	}
	if err := ValidateCarrierLanes(ActiveCarrierLanes(xs)); err != nil {
		t.Errorf("ValidateCarrierLanes() = %v", err)
	}
}

func TestCarrierLaneNormalizeClamps(t *testing.T) {
	x := SampleCarrierLane("a", when)
	x.Comment = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierLane("r1", when)
	rev, err := r.PutCarrierLane(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierLane() = %v", err)
	}
	got, gotRev, err := r.GetCarrierLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierLanes() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierLane(CarrierLane{}, 0); err == nil {
		t.Error("PutCarrierLane() accepted an empty entity")
	}
}

func TestCarrierSlotRoundTrip(t *testing.T) {
	x := SampleCarrierSlot("carrierslot-1", when)
	got, err := ParseCarrierSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierSlot() = %v", err)
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

func TestCarrierSlotParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierSlot(f); err == nil {
		t.Error("ParseCarrierSlot() accepted a record with no id")
	}
	f = SampleCarrierSlot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseCarrierSlot(f); err == nil {
		t.Error("ParseCarrierSlot() accepted a non-numeric offset")
	}
}

func TestCarrierSlotCollections(t *testing.T) {
	xs := []CarrierSlot{SampleCarrierSlot("a", when), SampleCarrierSlot("bb", when), SampleCarrierSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierSlots(xs)); got != 2 {
		t.Errorf("ActiveCarrierSlots() kept %d, want 2", got)
	}
	if got, want := SumCarrierSlots(xs), 1+3; got != want {
		t.Errorf("SumCarrierSlots() = %d, want %d", got, want)
	}
	SortCarrierSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierSlots() put %q first", xs[0].ID)
	}
	if len(IndexCarrierSlots(xs)) != 3 {
		t.Errorf("IndexCarrierSlots() = %v", IndexCarrierSlots(xs))
	}
	if err := ValidateCarrierSlots(ActiveCarrierSlots(xs)); err != nil {
		t.Errorf("ValidateCarrierSlots() = %v", err)
	}
}

func TestCarrierSlotNormalizeClamps(t *testing.T) {
	x := SampleCarrierSlot("a", when)
	x.Source = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierSlot("r1", when)
	rev, err := r.PutCarrierSlot(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierSlot() = %v", err)
	}
	got, gotRev, err := r.GetCarrierSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierSlots() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierSlot(CarrierSlot{}, 0); err == nil {
		t.Error("PutCarrierSlot() accepted an empty entity")
	}
}

func TestCarrierTagRoundTrip(t *testing.T) {
	x := SampleCarrierTag("carriertag-1", when)
	got, err := ParseCarrierTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierTag() = %v", err)
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

func TestCarrierTagParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierTag(f); err == nil {
		t.Error("ParseCarrierTag() accepted a record with no id")
	}
	f = SampleCarrierTag("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseCarrierTag(f); err == nil {
		t.Error("ParseCarrierTag() accepted a non-numeric priority")
	}
}

func TestCarrierTagCollections(t *testing.T) {
	xs := []CarrierTag{SampleCarrierTag("a", when), SampleCarrierTag("bb", when), SampleCarrierTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierTags(xs)); got != 2 {
		t.Errorf("ActiveCarrierTags() kept %d, want 2", got)
	}
	if got, want := SumCarrierTags(xs), 1+3; got != want {
		t.Errorf("SumCarrierTags() = %d, want %d", got, want)
	}
	SortCarrierTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierTags() put %q first", xs[0].ID)
	}
	if len(IndexCarrierTags(xs)) != 3 {
		t.Errorf("IndexCarrierTags() = %v", IndexCarrierTags(xs))
	}
	if err := ValidateCarrierTags(ActiveCarrierTags(xs)); err != nil {
		t.Errorf("ValidateCarrierTags() = %v", err)
	}
}

func TestCarrierTagNormalizeClamps(t *testing.T) {
	x := SampleCarrierTag("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierTag("r1", when)
	rev, err := r.PutCarrierTag(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierTag() = %v", err)
	}
	got, gotRev, err := r.GetCarrierTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierTags() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierTag(CarrierTag{}, 0); err == nil {
		t.Error("PutCarrierTag() accepted an empty entity")
	}
}

func TestCarrierRuleRoundTrip(t *testing.T) {
	x := SampleCarrierRule("carrierrule-1", when)
	got, err := ParseCarrierRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierRule() = %v", err)
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

func TestCarrierRuleParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierRule(f); err == nil {
		t.Error("ParseCarrierRule() accepted a record with no id")
	}
	f = SampleCarrierRule("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseCarrierRule(f); err == nil {
		t.Error("ParseCarrierRule() accepted a non-numeric weight")
	}
}

func TestCarrierRuleCollections(t *testing.T) {
	xs := []CarrierRule{SampleCarrierRule("a", when), SampleCarrierRule("bb", when), SampleCarrierRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierRules(xs)); got != 2 {
		t.Errorf("ActiveCarrierRules() kept %d, want 2", got)
	}
	if got, want := SumCarrierRules(xs), 1+3; got != want {
		t.Errorf("SumCarrierRules() = %d, want %d", got, want)
	}
	SortCarrierRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierRules() put %q first", xs[0].ID)
	}
	if len(IndexCarrierRules(xs)) != 3 {
		t.Errorf("IndexCarrierRules() = %v", IndexCarrierRules(xs))
	}
	if err := ValidateCarrierRules(ActiveCarrierRules(xs)); err != nil {
		t.Errorf("ValidateCarrierRules() = %v", err)
	}
}

func TestCarrierRuleNormalizeClamps(t *testing.T) {
	x := SampleCarrierRule("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierRule("r1", when)
	rev, err := r.PutCarrierRule(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierRule() = %v", err)
	}
	got, gotRev, err := r.GetCarrierRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierRules() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierRule(CarrierRule{}, 0); err == nil {
		t.Error("PutCarrierRule() accepted an empty entity")
	}
}

func TestCarrierHoldRoundTrip(t *testing.T) {
	x := SampleCarrierHold("carrierhold-1", when)
	got, err := ParseCarrierHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierHold() = %v", err)
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

func TestCarrierHoldParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierHold(f); err == nil {
		t.Error("ParseCarrierHold() accepted a record with no id")
	}
	f = SampleCarrierHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseCarrierHold(f); err == nil {
		t.Error("ParseCarrierHold() accepted a non-numeric weight")
	}
}

func TestCarrierHoldCollections(t *testing.T) {
	xs := []CarrierHold{SampleCarrierHold("a", when), SampleCarrierHold("bb", when), SampleCarrierHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierHolds(xs)); got != 2 {
		t.Errorf("ActiveCarrierHolds() kept %d, want 2", got)
	}
	if got, want := SumCarrierHolds(xs), 1+3; got != want {
		t.Errorf("SumCarrierHolds() = %d, want %d", got, want)
	}
	SortCarrierHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierHolds() put %q first", xs[0].ID)
	}
	if len(IndexCarrierHolds(xs)) != 3 {
		t.Errorf("IndexCarrierHolds() = %v", IndexCarrierHolds(xs))
	}
	if err := ValidateCarrierHolds(ActiveCarrierHolds(xs)); err != nil {
		t.Errorf("ValidateCarrierHolds() = %v", err)
	}
}

func TestCarrierHoldNormalizeClamps(t *testing.T) {
	x := SampleCarrierHold("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierHold("r1", when)
	rev, err := r.PutCarrierHold(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierHold() = %v", err)
	}
	got, gotRev, err := r.GetCarrierHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierHolds() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierHold(CarrierHold{}, 0); err == nil {
		t.Error("PutCarrierHold() accepted an empty entity")
	}
}

func TestCarrierClaimRoundTrip(t *testing.T) {
	x := SampleCarrierClaim("carrierclaim-1", when)
	got, err := ParseCarrierClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierClaim() = %v", err)
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

func TestCarrierClaimParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierClaim(f); err == nil {
		t.Error("ParseCarrierClaim() accepted a record with no id")
	}
	f = SampleCarrierClaim("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCarrierClaim(f); err == nil {
		t.Error("ParseCarrierClaim() accepted a non-numeric count")
	}
}

func TestCarrierClaimCollections(t *testing.T) {
	xs := []CarrierClaim{SampleCarrierClaim("a", when), SampleCarrierClaim("bb", when), SampleCarrierClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierClaims(xs)); got != 2 {
		t.Errorf("ActiveCarrierClaims() kept %d, want 2", got)
	}
	if got, want := SumCarrierClaims(xs), 1+3; got != want {
		t.Errorf("SumCarrierClaims() = %d, want %d", got, want)
	}
	SortCarrierClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierClaims() put %q first", xs[0].ID)
	}
	if len(IndexCarrierClaims(xs)) != 3 {
		t.Errorf("IndexCarrierClaims() = %v", IndexCarrierClaims(xs))
	}
	if err := ValidateCarrierClaims(ActiveCarrierClaims(xs)); err != nil {
		t.Errorf("ValidateCarrierClaims() = %v", err)
	}
}

func TestCarrierClaimNormalizeClamps(t *testing.T) {
	x := SampleCarrierClaim("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierClaim("r1", when)
	rev, err := r.PutCarrierClaim(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierClaim() = %v", err)
	}
	got, gotRev, err := r.GetCarrierClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierClaims() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierClaim(CarrierClaim{}, 0); err == nil {
		t.Error("PutCarrierClaim() accepted an empty entity")
	}
}

func TestCarrierCursorRoundTrip(t *testing.T) {
	x := SampleCarrierCursor("carriercursor-1", when)
	got, err := ParseCarrierCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseCarrierCursor() = %v", err)
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

func TestCarrierCursorParseRejectsBadInput(t *testing.T) {
	f := SampleCarrierCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCarrierCursor(f); err == nil {
		t.Error("ParseCarrierCursor() accepted a record with no id")
	}
	f = SampleCarrierCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseCarrierCursor(f); err == nil {
		t.Error("ParseCarrierCursor() accepted a non-numeric weight")
	}
}

func TestCarrierCursorCollections(t *testing.T) {
	xs := []CarrierCursor{SampleCarrierCursor("a", when), SampleCarrierCursor("bb", when), SampleCarrierCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCarrierCursors(xs)); got != 2 {
		t.Errorf("ActiveCarrierCursors() kept %d, want 2", got)
	}
	if got, want := SumCarrierCursors(xs), 1+3; got != want {
		t.Errorf("SumCarrierCursors() = %d, want %d", got, want)
	}
	SortCarrierCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCarrierCursors() put %q first", xs[0].ID)
	}
	if len(IndexCarrierCursors(xs)) != 3 {
		t.Errorf("IndexCarrierCursors() = %v", IndexCarrierCursors(xs))
	}
	if err := ValidateCarrierCursors(ActiveCarrierCursors(xs)); err != nil {
		t.Errorf("ValidateCarrierCursors() = %v", err)
	}
}

func TestCarrierCursorNormalizeClamps(t *testing.T) {
	x := SampleCarrierCursor("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCarrierCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleCarrierCursor("r1", when)
	rev, err := r.PutCarrierCursor(x, 0)
	if err != nil {
		t.Fatalf("PutCarrierCursor() = %v", err)
	}
	got, gotRev, err := r.GetCarrierCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetCarrierCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListCarrierCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListCarrierCursors() = %v, %v", all, err)
	}
	if _, err := r.PutCarrierCursor(CarrierCursor{}, 0); err == nil {
		t.Error("PutCarrierCursor() accepted an empty entity")
	}
}
