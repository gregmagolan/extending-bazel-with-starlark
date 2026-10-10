package delivery

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
		if !strings.HasPrefix(Prefixes[i], "delivery/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleDeliveryRecord("ok", when).Fields()); err != nil {
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

func TestDeliveryRecordRoundTrip(t *testing.T) {
	x := SampleDeliveryRecord("deliveryrecord-1", when)
	got, err := ParseDeliveryRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryRecord() = %v", err)
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

func TestDeliveryRecordParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryRecord(f); err == nil {
		t.Error("ParseDeliveryRecord() accepted a record with no id")
	}
	f = SampleDeliveryRecord("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeliveryRecord(f); err == nil {
		t.Error("ParseDeliveryRecord() accepted a non-numeric priority")
	}
}

func TestDeliveryRecordCollections(t *testing.T) {
	xs := []DeliveryRecord{SampleDeliveryRecord("a", when), SampleDeliveryRecord("bb", when), SampleDeliveryRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryRecords(xs)); got != 2 {
		t.Errorf("ActiveDeliveryRecords() kept %d, want 2", got)
	}
	if got, want := SumDeliveryRecords(xs), 1+3; got != want {
		t.Errorf("SumDeliveryRecords() = %d, want %d", got, want)
	}
	SortDeliveryRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryRecords() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryRecords(xs)) != 3 {
		t.Errorf("IndexDeliveryRecords() = %v", IndexDeliveryRecords(xs))
	}
	if err := ValidateDeliveryRecords(ActiveDeliveryRecords(xs)); err != nil {
		t.Errorf("ValidateDeliveryRecords() = %v", err)
	}
}

func TestDeliveryRecordNormalizeClamps(t *testing.T) {
	x := SampleDeliveryRecord("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryRecord("r1", when)
	rev, err := r.PutDeliveryRecord(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryRecord() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryRecords() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryRecord(DeliveryRecord{}, 0); err == nil {
		t.Error("PutDeliveryRecord() accepted an empty entity")
	}
}

func TestDeliveryEntryRoundTrip(t *testing.T) {
	x := SampleDeliveryEntry("deliveryentry-1", when)
	got, err := ParseDeliveryEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryEntry() = %v", err)
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

func TestDeliveryEntryParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryEntry(f); err == nil {
		t.Error("ParseDeliveryEntry() accepted a record with no id")
	}
	f = SampleDeliveryEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDeliveryEntry(f); err == nil {
		t.Error("ParseDeliveryEntry() accepted a non-numeric offset")
	}
}

func TestDeliveryEntryCollections(t *testing.T) {
	xs := []DeliveryEntry{SampleDeliveryEntry("a", when), SampleDeliveryEntry("bb", when), SampleDeliveryEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryEntrys(xs)); got != 2 {
		t.Errorf("ActiveDeliveryEntrys() kept %d, want 2", got)
	}
	if got, want := SumDeliveryEntrys(xs), 1+3; got != want {
		t.Errorf("SumDeliveryEntrys() = %d, want %d", got, want)
	}
	SortDeliveryEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryEntrys() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryEntrys(xs)) != 3 {
		t.Errorf("IndexDeliveryEntrys() = %v", IndexDeliveryEntrys(xs))
	}
	if err := ValidateDeliveryEntrys(ActiveDeliveryEntrys(xs)); err != nil {
		t.Errorf("ValidateDeliveryEntrys() = %v", err)
	}
}

func TestDeliveryEntryNormalizeClamps(t *testing.T) {
	x := SampleDeliveryEntry("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryEntry("r1", when)
	rev, err := r.PutDeliveryEntry(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryEntry() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryEntry(DeliveryEntry{}, 0); err == nil {
		t.Error("PutDeliveryEntry() accepted an empty entity")
	}
}

func TestDeliveryItemRoundTrip(t *testing.T) {
	x := SampleDeliveryItem("deliveryitem-1", when)
	got, err := ParseDeliveryItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryItem() = %v", err)
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

func TestDeliveryItemParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryItem(f); err == nil {
		t.Error("ParseDeliveryItem() accepted a record with no id")
	}
	f = SampleDeliveryItem("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDeliveryItem(f); err == nil {
		t.Error("ParseDeliveryItem() accepted a non-numeric attempts")
	}
}

func TestDeliveryItemCollections(t *testing.T) {
	xs := []DeliveryItem{SampleDeliveryItem("a", when), SampleDeliveryItem("bb", when), SampleDeliveryItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryItems(xs)); got != 2 {
		t.Errorf("ActiveDeliveryItems() kept %d, want 2", got)
	}
	if got, want := SumDeliveryItems(xs), 1+3; got != want {
		t.Errorf("SumDeliveryItems() = %d, want %d", got, want)
	}
	SortDeliveryItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryItems() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryItems(xs)) != 3 {
		t.Errorf("IndexDeliveryItems() = %v", IndexDeliveryItems(xs))
	}
	if err := ValidateDeliveryItems(ActiveDeliveryItems(xs)); err != nil {
		t.Errorf("ValidateDeliveryItems() = %v", err)
	}
}

func TestDeliveryItemNormalizeClamps(t *testing.T) {
	x := SampleDeliveryItem("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryItem("r1", when)
	rev, err := r.PutDeliveryItem(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryItem() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryItems() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryItem(DeliveryItem{}, 0); err == nil {
		t.Error("PutDeliveryItem() accepted an empty entity")
	}
}

func TestDeliveryDraftRoundTrip(t *testing.T) {
	x := SampleDeliveryDraft("deliverydraft-1", when)
	got, err := ParseDeliveryDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryDraft() = %v", err)
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

func TestDeliveryDraftParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryDraft(f); err == nil {
		t.Error("ParseDeliveryDraft() accepted a record with no id")
	}
	f = SampleDeliveryDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDeliveryDraft(f); err == nil {
		t.Error("ParseDeliveryDraft() accepted a non-numeric attempts")
	}
}

func TestDeliveryDraftCollections(t *testing.T) {
	xs := []DeliveryDraft{SampleDeliveryDraft("a", when), SampleDeliveryDraft("bb", when), SampleDeliveryDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryDrafts(xs)); got != 2 {
		t.Errorf("ActiveDeliveryDrafts() kept %d, want 2", got)
	}
	if got, want := SumDeliveryDrafts(xs), 1+3; got != want {
		t.Errorf("SumDeliveryDrafts() = %d, want %d", got, want)
	}
	SortDeliveryDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryDrafts() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryDrafts(xs)) != 3 {
		t.Errorf("IndexDeliveryDrafts() = %v", IndexDeliveryDrafts(xs))
	}
	if err := ValidateDeliveryDrafts(ActiveDeliveryDrafts(xs)); err != nil {
		t.Errorf("ValidateDeliveryDrafts() = %v", err)
	}
}

func TestDeliveryDraftNormalizeClamps(t *testing.T) {
	x := SampleDeliveryDraft("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryDraft("r1", when)
	rev, err := r.PutDeliveryDraft(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryDraft() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryDraft(DeliveryDraft{}, 0); err == nil {
		t.Error("PutDeliveryDraft() accepted an empty entity")
	}
}

func TestDeliveryRevisionRoundTrip(t *testing.T) {
	x := SampleDeliveryRevision("deliveryrevision-1", when)
	got, err := ParseDeliveryRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryRevision() = %v", err)
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

func TestDeliveryRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryRevision(f); err == nil {
		t.Error("ParseDeliveryRevision() accepted a record with no id")
	}
	f = SampleDeliveryRevision("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDeliveryRevision(f); err == nil {
		t.Error("ParseDeliveryRevision() accepted a non-numeric offset")
	}
}

func TestDeliveryRevisionCollections(t *testing.T) {
	xs := []DeliveryRevision{SampleDeliveryRevision("a", when), SampleDeliveryRevision("bb", when), SampleDeliveryRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryRevisions(xs)); got != 2 {
		t.Errorf("ActiveDeliveryRevisions() kept %d, want 2", got)
	}
	if got, want := SumDeliveryRevisions(xs), 1+3; got != want {
		t.Errorf("SumDeliveryRevisions() = %d, want %d", got, want)
	}
	SortDeliveryRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryRevisions() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryRevisions(xs)) != 3 {
		t.Errorf("IndexDeliveryRevisions() = %v", IndexDeliveryRevisions(xs))
	}
	if err := ValidateDeliveryRevisions(ActiveDeliveryRevisions(xs)); err != nil {
		t.Errorf("ValidateDeliveryRevisions() = %v", err)
	}
}

func TestDeliveryRevisionNormalizeClamps(t *testing.T) {
	x := SampleDeliveryRevision("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryRevision("r1", when)
	rev, err := r.PutDeliveryRevision(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryRevision() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryRevision(DeliveryRevision{}, 0); err == nil {
		t.Error("PutDeliveryRevision() accepted an empty entity")
	}
}

func TestDeliverySnapshotRoundTrip(t *testing.T) {
	x := SampleDeliverySnapshot("deliverysnapshot-1", when)
	got, err := ParseDeliverySnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliverySnapshot() = %v", err)
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

func TestDeliverySnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleDeliverySnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliverySnapshot(f); err == nil {
		t.Error("ParseDeliverySnapshot() accepted a record with no id")
	}
	f = SampleDeliverySnapshot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDeliverySnapshot(f); err == nil {
		t.Error("ParseDeliverySnapshot() accepted a non-numeric offset")
	}
}

func TestDeliverySnapshotCollections(t *testing.T) {
	xs := []DeliverySnapshot{SampleDeliverySnapshot("a", when), SampleDeliverySnapshot("bb", when), SampleDeliverySnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliverySnapshots(xs)); got != 2 {
		t.Errorf("ActiveDeliverySnapshots() kept %d, want 2", got)
	}
	if got, want := SumDeliverySnapshots(xs), 1+3; got != want {
		t.Errorf("SumDeliverySnapshots() = %d, want %d", got, want)
	}
	SortDeliverySnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliverySnapshots() put %q first", xs[0].ID)
	}
	if len(IndexDeliverySnapshots(xs)) != 3 {
		t.Errorf("IndexDeliverySnapshots() = %v", IndexDeliverySnapshots(xs))
	}
	if err := ValidateDeliverySnapshots(ActiveDeliverySnapshots(xs)); err != nil {
		t.Errorf("ValidateDeliverySnapshots() = %v", err)
	}
}

func TestDeliverySnapshotNormalizeClamps(t *testing.T) {
	x := SampleDeliverySnapshot("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliverySnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliverySnapshot("r1", when)
	rev, err := r.PutDeliverySnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutDeliverySnapshot() = %v", err)
	}
	got, gotRev, err := r.GetDeliverySnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliverySnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliverySnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliverySnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutDeliverySnapshot(DeliverySnapshot{}, 0); err == nil {
		t.Error("PutDeliverySnapshot() accepted an empty entity")
	}
}

func TestDeliveryBatchRoundTrip(t *testing.T) {
	x := SampleDeliveryBatch("deliverybatch-1", when)
	got, err := ParseDeliveryBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryBatch() = %v", err)
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

func TestDeliveryBatchParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryBatch(f); err == nil {
		t.Error("ParseDeliveryBatch() accepted a record with no id")
	}
	f = SampleDeliveryBatch("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeliveryBatch(f); err == nil {
		t.Error("ParseDeliveryBatch() accepted a non-numeric priority")
	}
}

func TestDeliveryBatchCollections(t *testing.T) {
	xs := []DeliveryBatch{SampleDeliveryBatch("a", when), SampleDeliveryBatch("bb", when), SampleDeliveryBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryBatchs(xs)); got != 2 {
		t.Errorf("ActiveDeliveryBatchs() kept %d, want 2", got)
	}
	if got, want := SumDeliveryBatchs(xs), 1+3; got != want {
		t.Errorf("SumDeliveryBatchs() = %d, want %d", got, want)
	}
	SortDeliveryBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryBatchs() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryBatchs(xs)) != 3 {
		t.Errorf("IndexDeliveryBatchs() = %v", IndexDeliveryBatchs(xs))
	}
	if err := ValidateDeliveryBatchs(ActiveDeliveryBatchs(xs)); err != nil {
		t.Errorf("ValidateDeliveryBatchs() = %v", err)
	}
}

func TestDeliveryBatchNormalizeClamps(t *testing.T) {
	x := SampleDeliveryBatch("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryBatch("r1", when)
	rev, err := r.PutDeliveryBatch(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryBatch() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryBatch(DeliveryBatch{}, 0); err == nil {
		t.Error("PutDeliveryBatch() accepted an empty entity")
	}
}

func TestDeliveryRequestRoundTrip(t *testing.T) {
	x := SampleDeliveryRequest("deliveryrequest-1", when)
	got, err := ParseDeliveryRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryRequest() = %v", err)
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

func TestDeliveryRequestParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryRequest(f); err == nil {
		t.Error("ParseDeliveryRequest() accepted a record with no id")
	}
	f = SampleDeliveryRequest("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeliveryRequest(f); err == nil {
		t.Error("ParseDeliveryRequest() accepted a non-numeric priority")
	}
}

func TestDeliveryRequestCollections(t *testing.T) {
	xs := []DeliveryRequest{SampleDeliveryRequest("a", when), SampleDeliveryRequest("bb", when), SampleDeliveryRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryRequests(xs)); got != 2 {
		t.Errorf("ActiveDeliveryRequests() kept %d, want 2", got)
	}
	if got, want := SumDeliveryRequests(xs), 1+3; got != want {
		t.Errorf("SumDeliveryRequests() = %d, want %d", got, want)
	}
	SortDeliveryRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryRequests() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryRequests(xs)) != 3 {
		t.Errorf("IndexDeliveryRequests() = %v", IndexDeliveryRequests(xs))
	}
	if err := ValidateDeliveryRequests(ActiveDeliveryRequests(xs)); err != nil {
		t.Errorf("ValidateDeliveryRequests() = %v", err)
	}
}

func TestDeliveryRequestNormalizeClamps(t *testing.T) {
	x := SampleDeliveryRequest("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryRequest("r1", when)
	rev, err := r.PutDeliveryRequest(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryRequest() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryRequests() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryRequest(DeliveryRequest{}, 0); err == nil {
		t.Error("PutDeliveryRequest() accepted an empty entity")
	}
}

func TestDeliverySummaryRoundTrip(t *testing.T) {
	x := SampleDeliverySummary("deliverysummary-1", when)
	got, err := ParseDeliverySummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliverySummary() = %v", err)
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

func TestDeliverySummaryParseRejectsBadInput(t *testing.T) {
	f := SampleDeliverySummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliverySummary(f); err == nil {
		t.Error("ParseDeliverySummary() accepted a record with no id")
	}
	f = SampleDeliverySummary("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDeliverySummary(f); err == nil {
		t.Error("ParseDeliverySummary() accepted a non-numeric weight")
	}
}

func TestDeliverySummaryCollections(t *testing.T) {
	xs := []DeliverySummary{SampleDeliverySummary("a", when), SampleDeliverySummary("bb", when), SampleDeliverySummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliverySummarys(xs)); got != 2 {
		t.Errorf("ActiveDeliverySummarys() kept %d, want 2", got)
	}
	if got, want := SumDeliverySummarys(xs), 1+3; got != want {
		t.Errorf("SumDeliverySummarys() = %d, want %d", got, want)
	}
	SortDeliverySummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliverySummarys() put %q first", xs[0].ID)
	}
	if len(IndexDeliverySummarys(xs)) != 3 {
		t.Errorf("IndexDeliverySummarys() = %v", IndexDeliverySummarys(xs))
	}
	if err := ValidateDeliverySummarys(ActiveDeliverySummarys(xs)); err != nil {
		t.Errorf("ValidateDeliverySummarys() = %v", err)
	}
}

func TestDeliverySummaryNormalizeClamps(t *testing.T) {
	x := SampleDeliverySummary("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliverySummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliverySummary("r1", when)
	rev, err := r.PutDeliverySummary(x, 0)
	if err != nil {
		t.Fatalf("PutDeliverySummary() = %v", err)
	}
	got, gotRev, err := r.GetDeliverySummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliverySummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliverySummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliverySummarys() = %v, %v", all, err)
	}
	if _, err := r.PutDeliverySummary(DeliverySummary{}, 0); err == nil {
		t.Error("PutDeliverySummary() accepted an empty entity")
	}
}

func TestDeliveryPolicyRoundTrip(t *testing.T) {
	x := SampleDeliveryPolicy("deliverypolicy-1", when)
	got, err := ParseDeliveryPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryPolicy() = %v", err)
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

func TestDeliveryPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryPolicy(f); err == nil {
		t.Error("ParseDeliveryPolicy() accepted a record with no id")
	}
	f = SampleDeliveryPolicy("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDeliveryPolicy(f); err == nil {
		t.Error("ParseDeliveryPolicy() accepted a non-numeric retained")
	}
}

func TestDeliveryPolicyCollections(t *testing.T) {
	xs := []DeliveryPolicy{SampleDeliveryPolicy("a", when), SampleDeliveryPolicy("bb", when), SampleDeliveryPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryPolicys(xs)); got != 2 {
		t.Errorf("ActiveDeliveryPolicys() kept %d, want 2", got)
	}
	if got, want := SumDeliveryPolicys(xs), 1+3; got != want {
		t.Errorf("SumDeliveryPolicys() = %d, want %d", got, want)
	}
	SortDeliveryPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryPolicys() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryPolicys(xs)) != 3 {
		t.Errorf("IndexDeliveryPolicys() = %v", IndexDeliveryPolicys(xs))
	}
	if err := ValidateDeliveryPolicys(ActiveDeliveryPolicys(xs)); err != nil {
		t.Errorf("ValidateDeliveryPolicys() = %v", err)
	}
}

func TestDeliveryPolicyNormalizeClamps(t *testing.T) {
	x := SampleDeliveryPolicy("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryPolicy("r1", when)
	rev, err := r.PutDeliveryPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryPolicy() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryPolicy(DeliveryPolicy{}, 0); err == nil {
		t.Error("PutDeliveryPolicy() accepted an empty entity")
	}
}

func TestDeliveryLinkRoundTrip(t *testing.T) {
	x := SampleDeliveryLink("deliverylink-1", when)
	got, err := ParseDeliveryLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryLink() = %v", err)
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

func TestDeliveryLinkParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryLink(f); err == nil {
		t.Error("ParseDeliveryLink() accepted a record with no id")
	}
	f = SampleDeliveryLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDeliveryLink(f); err == nil {
		t.Error("ParseDeliveryLink() accepted a non-numeric retained")
	}
}

func TestDeliveryLinkCollections(t *testing.T) {
	xs := []DeliveryLink{SampleDeliveryLink("a", when), SampleDeliveryLink("bb", when), SampleDeliveryLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryLinks(xs)); got != 2 {
		t.Errorf("ActiveDeliveryLinks() kept %d, want 2", got)
	}
	if got, want := SumDeliveryLinks(xs), 1+3; got != want {
		t.Errorf("SumDeliveryLinks() = %d, want %d", got, want)
	}
	SortDeliveryLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryLinks() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryLinks(xs)) != 3 {
		t.Errorf("IndexDeliveryLinks() = %v", IndexDeliveryLinks(xs))
	}
	if err := ValidateDeliveryLinks(ActiveDeliveryLinks(xs)); err != nil {
		t.Errorf("ValidateDeliveryLinks() = %v", err)
	}
}

func TestDeliveryLinkNormalizeClamps(t *testing.T) {
	x := SampleDeliveryLink("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryLink("r1", when)
	rev, err := r.PutDeliveryLink(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryLink() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryLinks() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryLink(DeliveryLink{}, 0); err == nil {
		t.Error("PutDeliveryLink() accepted an empty entity")
	}
}

func TestDeliveryNoteRoundTrip(t *testing.T) {
	x := SampleDeliveryNote("deliverynote-1", when)
	got, err := ParseDeliveryNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryNote() = %v", err)
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

func TestDeliveryNoteParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryNote(f); err == nil {
		t.Error("ParseDeliveryNote() accepted a record with no id")
	}
	f = SampleDeliveryNote("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeliveryNote(f); err == nil {
		t.Error("ParseDeliveryNote() accepted a non-numeric version")
	}
}

func TestDeliveryNoteCollections(t *testing.T) {
	xs := []DeliveryNote{SampleDeliveryNote("a", when), SampleDeliveryNote("bb", when), SampleDeliveryNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryNotes(xs)); got != 2 {
		t.Errorf("ActiveDeliveryNotes() kept %d, want 2", got)
	}
	if got, want := SumDeliveryNotes(xs), 1+3; got != want {
		t.Errorf("SumDeliveryNotes() = %d, want %d", got, want)
	}
	SortDeliveryNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryNotes() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryNotes(xs)) != 3 {
		t.Errorf("IndexDeliveryNotes() = %v", IndexDeliveryNotes(xs))
	}
	if err := ValidateDeliveryNotes(ActiveDeliveryNotes(xs)); err != nil {
		t.Errorf("ValidateDeliveryNotes() = %v", err)
	}
}

func TestDeliveryNoteNormalizeClamps(t *testing.T) {
	x := SampleDeliveryNote("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryNote("r1", when)
	rev, err := r.PutDeliveryNote(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryNote() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryNotes() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryNote(DeliveryNote{}, 0); err == nil {
		t.Error("PutDeliveryNote() accepted an empty entity")
	}
}

func TestDeliveryWindowRoundTrip(t *testing.T) {
	x := SampleDeliveryWindow("deliverywindow-1", when)
	got, err := ParseDeliveryWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryWindow() = %v", err)
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

func TestDeliveryWindowParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryWindow(f); err == nil {
		t.Error("ParseDeliveryWindow() accepted a record with no id")
	}
	f = SampleDeliveryWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeliveryWindow(f); err == nil {
		t.Error("ParseDeliveryWindow() accepted a non-numeric priority")
	}
}

func TestDeliveryWindowCollections(t *testing.T) {
	xs := []DeliveryWindow{SampleDeliveryWindow("a", when), SampleDeliveryWindow("bb", when), SampleDeliveryWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryWindows(xs)); got != 2 {
		t.Errorf("ActiveDeliveryWindows() kept %d, want 2", got)
	}
	if got, want := SumDeliveryWindows(xs), 1+3; got != want {
		t.Errorf("SumDeliveryWindows() = %d, want %d", got, want)
	}
	SortDeliveryWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryWindows() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryWindows(xs)) != 3 {
		t.Errorf("IndexDeliveryWindows() = %v", IndexDeliveryWindows(xs))
	}
	if err := ValidateDeliveryWindows(ActiveDeliveryWindows(xs)); err != nil {
		t.Errorf("ValidateDeliveryWindows() = %v", err)
	}
}

func TestDeliveryWindowNormalizeClamps(t *testing.T) {
	x := SampleDeliveryWindow("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryWindow("r1", when)
	rev, err := r.PutDeliveryWindow(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryWindow() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryWindows() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryWindow(DeliveryWindow{}, 0); err == nil {
		t.Error("PutDeliveryWindow() accepted an empty entity")
	}
}

func TestDeliveryQuotaRoundTrip(t *testing.T) {
	x := SampleDeliveryQuota("deliveryquota-1", when)
	got, err := ParseDeliveryQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryQuota() = %v", err)
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

func TestDeliveryQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryQuota(f); err == nil {
		t.Error("ParseDeliveryQuota() accepted a record with no id")
	}
	f = SampleDeliveryQuota("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeliveryQuota(f); err == nil {
		t.Error("ParseDeliveryQuota() accepted a non-numeric priority")
	}
}

func TestDeliveryQuotaCollections(t *testing.T) {
	xs := []DeliveryQuota{SampleDeliveryQuota("a", when), SampleDeliveryQuota("bb", when), SampleDeliveryQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryQuotas(xs)); got != 2 {
		t.Errorf("ActiveDeliveryQuotas() kept %d, want 2", got)
	}
	if got, want := SumDeliveryQuotas(xs), 1+3; got != want {
		t.Errorf("SumDeliveryQuotas() = %d, want %d", got, want)
	}
	SortDeliveryQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryQuotas() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryQuotas(xs)) != 3 {
		t.Errorf("IndexDeliveryQuotas() = %v", IndexDeliveryQuotas(xs))
	}
	if err := ValidateDeliveryQuotas(ActiveDeliveryQuotas(xs)); err != nil {
		t.Errorf("ValidateDeliveryQuotas() = %v", err)
	}
}

func TestDeliveryQuotaNormalizeClamps(t *testing.T) {
	x := SampleDeliveryQuota("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryQuota("r1", when)
	rev, err := r.PutDeliveryQuota(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryQuota() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryQuota(DeliveryQuota{}, 0); err == nil {
		t.Error("PutDeliveryQuota() accepted an empty entity")
	}
}

func TestDeliveryBucketRoundTrip(t *testing.T) {
	x := SampleDeliveryBucket("deliverybucket-1", when)
	got, err := ParseDeliveryBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryBucket() = %v", err)
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

func TestDeliveryBucketParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryBucket(f); err == nil {
		t.Error("ParseDeliveryBucket() accepted a record with no id")
	}
	f = SampleDeliveryBucket("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDeliveryBucket(f); err == nil {
		t.Error("ParseDeliveryBucket() accepted a non-numeric count")
	}
}

func TestDeliveryBucketCollections(t *testing.T) {
	xs := []DeliveryBucket{SampleDeliveryBucket("a", when), SampleDeliveryBucket("bb", when), SampleDeliveryBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryBuckets(xs)); got != 2 {
		t.Errorf("ActiveDeliveryBuckets() kept %d, want 2", got)
	}
	if got, want := SumDeliveryBuckets(xs), 1+3; got != want {
		t.Errorf("SumDeliveryBuckets() = %d, want %d", got, want)
	}
	SortDeliveryBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryBuckets() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryBuckets(xs)) != 3 {
		t.Errorf("IndexDeliveryBuckets() = %v", IndexDeliveryBuckets(xs))
	}
	if err := ValidateDeliveryBuckets(ActiveDeliveryBuckets(xs)); err != nil {
		t.Errorf("ValidateDeliveryBuckets() = %v", err)
	}
}

func TestDeliveryBucketNormalizeClamps(t *testing.T) {
	x := SampleDeliveryBucket("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryBucket("r1", when)
	rev, err := r.PutDeliveryBucket(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryBucket() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryBucket(DeliveryBucket{}, 0); err == nil {
		t.Error("PutDeliveryBucket() accepted an empty entity")
	}
}

func TestDeliveryLaneRoundTrip(t *testing.T) {
	x := SampleDeliveryLane("deliverylane-1", when)
	got, err := ParseDeliveryLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryLane() = %v", err)
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

func TestDeliveryLaneParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryLane(f); err == nil {
		t.Error("ParseDeliveryLane() accepted a record with no id")
	}
	f = SampleDeliveryLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDeliveryLane(f); err == nil {
		t.Error("ParseDeliveryLane() accepted a non-numeric capacity")
	}
}

func TestDeliveryLaneCollections(t *testing.T) {
	xs := []DeliveryLane{SampleDeliveryLane("a", when), SampleDeliveryLane("bb", when), SampleDeliveryLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryLanes(xs)); got != 2 {
		t.Errorf("ActiveDeliveryLanes() kept %d, want 2", got)
	}
	if got, want := SumDeliveryLanes(xs), 1+3; got != want {
		t.Errorf("SumDeliveryLanes() = %d, want %d", got, want)
	}
	SortDeliveryLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryLanes() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryLanes(xs)) != 3 {
		t.Errorf("IndexDeliveryLanes() = %v", IndexDeliveryLanes(xs))
	}
	if err := ValidateDeliveryLanes(ActiveDeliveryLanes(xs)); err != nil {
		t.Errorf("ValidateDeliveryLanes() = %v", err)
	}
}

func TestDeliveryLaneNormalizeClamps(t *testing.T) {
	x := SampleDeliveryLane("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryLane("r1", when)
	rev, err := r.PutDeliveryLane(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryLane() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryLanes() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryLane(DeliveryLane{}, 0); err == nil {
		t.Error("PutDeliveryLane() accepted an empty entity")
	}
}

func TestDeliverySlotRoundTrip(t *testing.T) {
	x := SampleDeliverySlot("deliveryslot-1", when)
	got, err := ParseDeliverySlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliverySlot() = %v", err)
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

func TestDeliverySlotParseRejectsBadInput(t *testing.T) {
	f := SampleDeliverySlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliverySlot(f); err == nil {
		t.Error("ParseDeliverySlot() accepted a record with no id")
	}
	f = SampleDeliverySlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDeliverySlot(f); err == nil {
		t.Error("ParseDeliverySlot() accepted a non-numeric count")
	}
}

func TestDeliverySlotCollections(t *testing.T) {
	xs := []DeliverySlot{SampleDeliverySlot("a", when), SampleDeliverySlot("bb", when), SampleDeliverySlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliverySlots(xs)); got != 2 {
		t.Errorf("ActiveDeliverySlots() kept %d, want 2", got)
	}
	if got, want := SumDeliverySlots(xs), 1+3; got != want {
		t.Errorf("SumDeliverySlots() = %d, want %d", got, want)
	}
	SortDeliverySlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliverySlots() put %q first", xs[0].ID)
	}
	if len(IndexDeliverySlots(xs)) != 3 {
		t.Errorf("IndexDeliverySlots() = %v", IndexDeliverySlots(xs))
	}
	if err := ValidateDeliverySlots(ActiveDeliverySlots(xs)); err != nil {
		t.Errorf("ValidateDeliverySlots() = %v", err)
	}
}

func TestDeliverySlotNormalizeClamps(t *testing.T) {
	x := SampleDeliverySlot("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliverySlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliverySlot("r1", when)
	rev, err := r.PutDeliverySlot(x, 0)
	if err != nil {
		t.Fatalf("PutDeliverySlot() = %v", err)
	}
	got, gotRev, err := r.GetDeliverySlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliverySlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliverySlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliverySlots() = %v, %v", all, err)
	}
	if _, err := r.PutDeliverySlot(DeliverySlot{}, 0); err == nil {
		t.Error("PutDeliverySlot() accepted an empty entity")
	}
}

func TestDeliveryTagRoundTrip(t *testing.T) {
	x := SampleDeliveryTag("deliverytag-1", when)
	got, err := ParseDeliveryTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryTag() = %v", err)
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

func TestDeliveryTagParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryTag(f); err == nil {
		t.Error("ParseDeliveryTag() accepted a record with no id")
	}
	f = SampleDeliveryTag("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeliveryTag(f); err == nil {
		t.Error("ParseDeliveryTag() accepted a non-numeric version")
	}
}

func TestDeliveryTagCollections(t *testing.T) {
	xs := []DeliveryTag{SampleDeliveryTag("a", when), SampleDeliveryTag("bb", when), SampleDeliveryTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryTags(xs)); got != 2 {
		t.Errorf("ActiveDeliveryTags() kept %d, want 2", got)
	}
	if got, want := SumDeliveryTags(xs), 1+3; got != want {
		t.Errorf("SumDeliveryTags() = %d, want %d", got, want)
	}
	SortDeliveryTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryTags() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryTags(xs)) != 3 {
		t.Errorf("IndexDeliveryTags() = %v", IndexDeliveryTags(xs))
	}
	if err := ValidateDeliveryTags(ActiveDeliveryTags(xs)); err != nil {
		t.Errorf("ValidateDeliveryTags() = %v", err)
	}
}

func TestDeliveryTagNormalizeClamps(t *testing.T) {
	x := SampleDeliveryTag("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryTag("r1", when)
	rev, err := r.PutDeliveryTag(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryTag() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryTags() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryTag(DeliveryTag{}, 0); err == nil {
		t.Error("PutDeliveryTag() accepted an empty entity")
	}
}

func TestDeliveryRuleRoundTrip(t *testing.T) {
	x := SampleDeliveryRule("deliveryrule-1", when)
	got, err := ParseDeliveryRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryRule() = %v", err)
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

func TestDeliveryRuleParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryRule(f); err == nil {
		t.Error("ParseDeliveryRule() accepted a record with no id")
	}
	f = SampleDeliveryRule("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeliveryRule(f); err == nil {
		t.Error("ParseDeliveryRule() accepted a non-numeric version")
	}
}

func TestDeliveryRuleCollections(t *testing.T) {
	xs := []DeliveryRule{SampleDeliveryRule("a", when), SampleDeliveryRule("bb", when), SampleDeliveryRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryRules(xs)); got != 2 {
		t.Errorf("ActiveDeliveryRules() kept %d, want 2", got)
	}
	if got, want := SumDeliveryRules(xs), 1+3; got != want {
		t.Errorf("SumDeliveryRules() = %d, want %d", got, want)
	}
	SortDeliveryRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryRules() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryRules(xs)) != 3 {
		t.Errorf("IndexDeliveryRules() = %v", IndexDeliveryRules(xs))
	}
	if err := ValidateDeliveryRules(ActiveDeliveryRules(xs)); err != nil {
		t.Errorf("ValidateDeliveryRules() = %v", err)
	}
}

func TestDeliveryRuleNormalizeClamps(t *testing.T) {
	x := SampleDeliveryRule("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryRule("r1", when)
	rev, err := r.PutDeliveryRule(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryRule() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryRules() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryRule(DeliveryRule{}, 0); err == nil {
		t.Error("PutDeliveryRule() accepted an empty entity")
	}
}

func TestDeliveryHoldRoundTrip(t *testing.T) {
	x := SampleDeliveryHold("deliveryhold-1", when)
	got, err := ParseDeliveryHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryHold() = %v", err)
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

func TestDeliveryHoldParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryHold(f); err == nil {
		t.Error("ParseDeliveryHold() accepted a record with no id")
	}
	f = SampleDeliveryHold("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeliveryHold(f); err == nil {
		t.Error("ParseDeliveryHold() accepted a non-numeric version")
	}
}

func TestDeliveryHoldCollections(t *testing.T) {
	xs := []DeliveryHold{SampleDeliveryHold("a", when), SampleDeliveryHold("bb", when), SampleDeliveryHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryHolds(xs)); got != 2 {
		t.Errorf("ActiveDeliveryHolds() kept %d, want 2", got)
	}
	if got, want := SumDeliveryHolds(xs), 1+3; got != want {
		t.Errorf("SumDeliveryHolds() = %d, want %d", got, want)
	}
	SortDeliveryHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryHolds() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryHolds(xs)) != 3 {
		t.Errorf("IndexDeliveryHolds() = %v", IndexDeliveryHolds(xs))
	}
	if err := ValidateDeliveryHolds(ActiveDeliveryHolds(xs)); err != nil {
		t.Errorf("ValidateDeliveryHolds() = %v", err)
	}
}

func TestDeliveryHoldNormalizeClamps(t *testing.T) {
	x := SampleDeliveryHold("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryHold("r1", when)
	rev, err := r.PutDeliveryHold(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryHold() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryHolds() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryHold(DeliveryHold{}, 0); err == nil {
		t.Error("PutDeliveryHold() accepted an empty entity")
	}
}

func TestDeliveryClaimRoundTrip(t *testing.T) {
	x := SampleDeliveryClaim("deliveryclaim-1", when)
	got, err := ParseDeliveryClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryClaim() = %v", err)
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

func TestDeliveryClaimParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryClaim(f); err == nil {
		t.Error("ParseDeliveryClaim() accepted a record with no id")
	}
	f = SampleDeliveryClaim("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeliveryClaim(f); err == nil {
		t.Error("ParseDeliveryClaim() accepted a non-numeric priority")
	}
}

func TestDeliveryClaimCollections(t *testing.T) {
	xs := []DeliveryClaim{SampleDeliveryClaim("a", when), SampleDeliveryClaim("bb", when), SampleDeliveryClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryClaims(xs)); got != 2 {
		t.Errorf("ActiveDeliveryClaims() kept %d, want 2", got)
	}
	if got, want := SumDeliveryClaims(xs), 1+3; got != want {
		t.Errorf("SumDeliveryClaims() = %d, want %d", got, want)
	}
	SortDeliveryClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryClaims() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryClaims(xs)) != 3 {
		t.Errorf("IndexDeliveryClaims() = %v", IndexDeliveryClaims(xs))
	}
	if err := ValidateDeliveryClaims(ActiveDeliveryClaims(xs)); err != nil {
		t.Errorf("ValidateDeliveryClaims() = %v", err)
	}
}

func TestDeliveryClaimNormalizeClamps(t *testing.T) {
	x := SampleDeliveryClaim("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryClaim("r1", when)
	rev, err := r.PutDeliveryClaim(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryClaim() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryClaims() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryClaim(DeliveryClaim{}, 0); err == nil {
		t.Error("PutDeliveryClaim() accepted an empty entity")
	}
}

func TestDeliveryCursorRoundTrip(t *testing.T) {
	x := SampleDeliveryCursor("deliverycursor-1", when)
	got, err := ParseDeliveryCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeliveryCursor() = %v", err)
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

func TestDeliveryCursorParseRejectsBadInput(t *testing.T) {
	f := SampleDeliveryCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeliveryCursor(f); err == nil {
		t.Error("ParseDeliveryCursor() accepted a record with no id")
	}
	f = SampleDeliveryCursor("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDeliveryCursor(f); err == nil {
		t.Error("ParseDeliveryCursor() accepted a non-numeric count")
	}
}

func TestDeliveryCursorCollections(t *testing.T) {
	xs := []DeliveryCursor{SampleDeliveryCursor("a", when), SampleDeliveryCursor("bb", when), SampleDeliveryCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeliveryCursors(xs)); got != 2 {
		t.Errorf("ActiveDeliveryCursors() kept %d, want 2", got)
	}
	if got, want := SumDeliveryCursors(xs), 1+3; got != want {
		t.Errorf("SumDeliveryCursors() = %d, want %d", got, want)
	}
	SortDeliveryCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeliveryCursors() put %q first", xs[0].ID)
	}
	if len(IndexDeliveryCursors(xs)) != 3 {
		t.Errorf("IndexDeliveryCursors() = %v", IndexDeliveryCursors(xs))
	}
	if err := ValidateDeliveryCursors(ActiveDeliveryCursors(xs)); err != nil {
		t.Errorf("ValidateDeliveryCursors() = %v", err)
	}
}

func TestDeliveryCursorNormalizeClamps(t *testing.T) {
	x := SampleDeliveryCursor("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeliveryCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeliveryCursor("r1", when)
	rev, err := r.PutDeliveryCursor(x, 0)
	if err != nil {
		t.Fatalf("PutDeliveryCursor() = %v", err)
	}
	got, gotRev, err := r.GetDeliveryCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeliveryCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeliveryCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeliveryCursors() = %v, %v", all, err)
	}
	if _, err := r.PutDeliveryCursor(DeliveryCursor{}, 0); err == nil {
		t.Error("PutDeliveryCursor() accepted an empty entity")
	}
}
