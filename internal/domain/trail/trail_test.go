package trail

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
		if !strings.HasPrefix(Prefixes[i], "trail/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleTrailRecord("ok", when).Fields()); err != nil {
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

func TestTrailRecordRoundTrip(t *testing.T) {
	x := SampleTrailRecord("trailrecord-1", when)
	got, err := ParseTrailRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailRecord() = %v", err)
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

func TestTrailRecordParseRejectsBadInput(t *testing.T) {
	f := SampleTrailRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailRecord(f); err == nil {
		t.Error("ParseTrailRecord() accepted a record with no id")
	}
	f = SampleTrailRecord("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTrailRecord(f); err == nil {
		t.Error("ParseTrailRecord() accepted a non-numeric offset")
	}
}

func TestTrailRecordCollections(t *testing.T) {
	xs := []TrailRecord{SampleTrailRecord("a", when), SampleTrailRecord("bb", when), SampleTrailRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailRecords(xs)); got != 2 {
		t.Errorf("ActiveTrailRecords() kept %d, want 2", got)
	}
	if got, want := SumTrailRecords(xs), 1+3; got != want {
		t.Errorf("SumTrailRecords() = %d, want %d", got, want)
	}
	SortTrailRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailRecords() put %q first", xs[0].ID)
	}
	if len(IndexTrailRecords(xs)) != 3 {
		t.Errorf("IndexTrailRecords() = %v", IndexTrailRecords(xs))
	}
	if err := ValidateTrailRecords(ActiveTrailRecords(xs)); err != nil {
		t.Errorf("ValidateTrailRecords() = %v", err)
	}
}

func TestTrailRecordNormalizeClamps(t *testing.T) {
	x := SampleTrailRecord("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailRecord("r1", when)
	rev, err := r.PutTrailRecord(x, 0)
	if err != nil {
		t.Fatalf("PutTrailRecord() = %v", err)
	}
	got, gotRev, err := r.GetTrailRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailRecords() = %v, %v", all, err)
	}
	if _, err := r.PutTrailRecord(TrailRecord{}, 0); err == nil {
		t.Error("PutTrailRecord() accepted an empty entity")
	}
}

func TestTrailEntryRoundTrip(t *testing.T) {
	x := SampleTrailEntry("trailentry-1", when)
	got, err := ParseTrailEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailEntry() = %v", err)
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

func TestTrailEntryParseRejectsBadInput(t *testing.T) {
	f := SampleTrailEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailEntry(f); err == nil {
		t.Error("ParseTrailEntry() accepted a record with no id")
	}
	f = SampleTrailEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTrailEntry(f); err == nil {
		t.Error("ParseTrailEntry() accepted a non-numeric offset")
	}
}

func TestTrailEntryCollections(t *testing.T) {
	xs := []TrailEntry{SampleTrailEntry("a", when), SampleTrailEntry("bb", when), SampleTrailEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailEntrys(xs)); got != 2 {
		t.Errorf("ActiveTrailEntrys() kept %d, want 2", got)
	}
	if got, want := SumTrailEntrys(xs), 1+3; got != want {
		t.Errorf("SumTrailEntrys() = %d, want %d", got, want)
	}
	SortTrailEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailEntrys() put %q first", xs[0].ID)
	}
	if len(IndexTrailEntrys(xs)) != 3 {
		t.Errorf("IndexTrailEntrys() = %v", IndexTrailEntrys(xs))
	}
	if err := ValidateTrailEntrys(ActiveTrailEntrys(xs)); err != nil {
		t.Errorf("ValidateTrailEntrys() = %v", err)
	}
}

func TestTrailEntryNormalizeClamps(t *testing.T) {
	x := SampleTrailEntry("a", when)
	x.Reference = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailEntry("r1", when)
	rev, err := r.PutTrailEntry(x, 0)
	if err != nil {
		t.Fatalf("PutTrailEntry() = %v", err)
	}
	got, gotRev, err := r.GetTrailEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutTrailEntry(TrailEntry{}, 0); err == nil {
		t.Error("PutTrailEntry() accepted an empty entity")
	}
}

func TestTrailItemRoundTrip(t *testing.T) {
	x := SampleTrailItem("trailitem-1", when)
	got, err := ParseTrailItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailItem() = %v", err)
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

func TestTrailItemParseRejectsBadInput(t *testing.T) {
	f := SampleTrailItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailItem(f); err == nil {
		t.Error("ParseTrailItem() accepted a record with no id")
	}
	f = SampleTrailItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTrailItem(f); err == nil {
		t.Error("ParseTrailItem() accepted a non-numeric version")
	}
}

func TestTrailItemCollections(t *testing.T) {
	xs := []TrailItem{SampleTrailItem("a", when), SampleTrailItem("bb", when), SampleTrailItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailItems(xs)); got != 2 {
		t.Errorf("ActiveTrailItems() kept %d, want 2", got)
	}
	if got, want := SumTrailItems(xs), 1+3; got != want {
		t.Errorf("SumTrailItems() = %d, want %d", got, want)
	}
	SortTrailItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailItems() put %q first", xs[0].ID)
	}
	if len(IndexTrailItems(xs)) != 3 {
		t.Errorf("IndexTrailItems() = %v", IndexTrailItems(xs))
	}
	if err := ValidateTrailItems(ActiveTrailItems(xs)); err != nil {
		t.Errorf("ValidateTrailItems() = %v", err)
	}
}

func TestTrailItemNormalizeClamps(t *testing.T) {
	x := SampleTrailItem("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailItem("r1", when)
	rev, err := r.PutTrailItem(x, 0)
	if err != nil {
		t.Fatalf("PutTrailItem() = %v", err)
	}
	got, gotRev, err := r.GetTrailItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailItems() = %v, %v", all, err)
	}
	if _, err := r.PutTrailItem(TrailItem{}, 0); err == nil {
		t.Error("PutTrailItem() accepted an empty entity")
	}
}

func TestTrailDraftRoundTrip(t *testing.T) {
	x := SampleTrailDraft("traildraft-1", when)
	got, err := ParseTrailDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailDraft() = %v", err)
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

func TestTrailDraftParseRejectsBadInput(t *testing.T) {
	f := SampleTrailDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailDraft(f); err == nil {
		t.Error("ParseTrailDraft() accepted a record with no id")
	}
	f = SampleTrailDraft("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTrailDraft(f); err == nil {
		t.Error("ParseTrailDraft() accepted a non-numeric capacity")
	}
}

func TestTrailDraftCollections(t *testing.T) {
	xs := []TrailDraft{SampleTrailDraft("a", when), SampleTrailDraft("bb", when), SampleTrailDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailDrafts(xs)); got != 2 {
		t.Errorf("ActiveTrailDrafts() kept %d, want 2", got)
	}
	if got, want := SumTrailDrafts(xs), 1+3; got != want {
		t.Errorf("SumTrailDrafts() = %d, want %d", got, want)
	}
	SortTrailDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailDrafts() put %q first", xs[0].ID)
	}
	if len(IndexTrailDrafts(xs)) != 3 {
		t.Errorf("IndexTrailDrafts() = %v", IndexTrailDrafts(xs))
	}
	if err := ValidateTrailDrafts(ActiveTrailDrafts(xs)); err != nil {
		t.Errorf("ValidateTrailDrafts() = %v", err)
	}
}

func TestTrailDraftNormalizeClamps(t *testing.T) {
	x := SampleTrailDraft("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailDraft("r1", when)
	rev, err := r.PutTrailDraft(x, 0)
	if err != nil {
		t.Fatalf("PutTrailDraft() = %v", err)
	}
	got, gotRev, err := r.GetTrailDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutTrailDraft(TrailDraft{}, 0); err == nil {
		t.Error("PutTrailDraft() accepted an empty entity")
	}
}

func TestTrailRevisionRoundTrip(t *testing.T) {
	x := SampleTrailRevision("trailrevision-1", when)
	got, err := ParseTrailRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailRevision() = %v", err)
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

func TestTrailRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleTrailRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailRevision(f); err == nil {
		t.Error("ParseTrailRevision() accepted a record with no id")
	}
	f = SampleTrailRevision("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTrailRevision(f); err == nil {
		t.Error("ParseTrailRevision() accepted a non-numeric version")
	}
}

func TestTrailRevisionCollections(t *testing.T) {
	xs := []TrailRevision{SampleTrailRevision("a", when), SampleTrailRevision("bb", when), SampleTrailRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailRevisions(xs)); got != 2 {
		t.Errorf("ActiveTrailRevisions() kept %d, want 2", got)
	}
	if got, want := SumTrailRevisions(xs), 1+3; got != want {
		t.Errorf("SumTrailRevisions() = %d, want %d", got, want)
	}
	SortTrailRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailRevisions() put %q first", xs[0].ID)
	}
	if len(IndexTrailRevisions(xs)) != 3 {
		t.Errorf("IndexTrailRevisions() = %v", IndexTrailRevisions(xs))
	}
	if err := ValidateTrailRevisions(ActiveTrailRevisions(xs)); err != nil {
		t.Errorf("ValidateTrailRevisions() = %v", err)
	}
}

func TestTrailRevisionNormalizeClamps(t *testing.T) {
	x := SampleTrailRevision("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailRevision("r1", when)
	rev, err := r.PutTrailRevision(x, 0)
	if err != nil {
		t.Fatalf("PutTrailRevision() = %v", err)
	}
	got, gotRev, err := r.GetTrailRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutTrailRevision(TrailRevision{}, 0); err == nil {
		t.Error("PutTrailRevision() accepted an empty entity")
	}
}

func TestTrailSnapshotRoundTrip(t *testing.T) {
	x := SampleTrailSnapshot("trailsnapshot-1", when)
	got, err := ParseTrailSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailSnapshot() = %v", err)
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

func TestTrailSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleTrailSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailSnapshot(f); err == nil {
		t.Error("ParseTrailSnapshot() accepted a record with no id")
	}
	f = SampleTrailSnapshot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTrailSnapshot(f); err == nil {
		t.Error("ParseTrailSnapshot() accepted a non-numeric offset")
	}
}

func TestTrailSnapshotCollections(t *testing.T) {
	xs := []TrailSnapshot{SampleTrailSnapshot("a", when), SampleTrailSnapshot("bb", when), SampleTrailSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailSnapshots(xs)); got != 2 {
		t.Errorf("ActiveTrailSnapshots() kept %d, want 2", got)
	}
	if got, want := SumTrailSnapshots(xs), 1+3; got != want {
		t.Errorf("SumTrailSnapshots() = %d, want %d", got, want)
	}
	SortTrailSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexTrailSnapshots(xs)) != 3 {
		t.Errorf("IndexTrailSnapshots() = %v", IndexTrailSnapshots(xs))
	}
	if err := ValidateTrailSnapshots(ActiveTrailSnapshots(xs)); err != nil {
		t.Errorf("ValidateTrailSnapshots() = %v", err)
	}
}

func TestTrailSnapshotNormalizeClamps(t *testing.T) {
	x := SampleTrailSnapshot("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailSnapshot("r1", when)
	rev, err := r.PutTrailSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutTrailSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetTrailSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutTrailSnapshot(TrailSnapshot{}, 0); err == nil {
		t.Error("PutTrailSnapshot() accepted an empty entity")
	}
}

func TestTrailBatchRoundTrip(t *testing.T) {
	x := SampleTrailBatch("trailbatch-1", when)
	got, err := ParseTrailBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailBatch() = %v", err)
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

func TestTrailBatchParseRejectsBadInput(t *testing.T) {
	f := SampleTrailBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailBatch(f); err == nil {
		t.Error("ParseTrailBatch() accepted a record with no id")
	}
	f = SampleTrailBatch("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTrailBatch(f); err == nil {
		t.Error("ParseTrailBatch() accepted a non-numeric retained")
	}
}

func TestTrailBatchCollections(t *testing.T) {
	xs := []TrailBatch{SampleTrailBatch("a", when), SampleTrailBatch("bb", when), SampleTrailBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailBatchs(xs)); got != 2 {
		t.Errorf("ActiveTrailBatchs() kept %d, want 2", got)
	}
	if got, want := SumTrailBatchs(xs), 1+3; got != want {
		t.Errorf("SumTrailBatchs() = %d, want %d", got, want)
	}
	SortTrailBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailBatchs() put %q first", xs[0].ID)
	}
	if len(IndexTrailBatchs(xs)) != 3 {
		t.Errorf("IndexTrailBatchs() = %v", IndexTrailBatchs(xs))
	}
	if err := ValidateTrailBatchs(ActiveTrailBatchs(xs)); err != nil {
		t.Errorf("ValidateTrailBatchs() = %v", err)
	}
}

func TestTrailBatchNormalizeClamps(t *testing.T) {
	x := SampleTrailBatch("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailBatch("r1", when)
	rev, err := r.PutTrailBatch(x, 0)
	if err != nil {
		t.Fatalf("PutTrailBatch() = %v", err)
	}
	got, gotRev, err := r.GetTrailBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutTrailBatch(TrailBatch{}, 0); err == nil {
		t.Error("PutTrailBatch() accepted an empty entity")
	}
}

func TestTrailRequestRoundTrip(t *testing.T) {
	x := SampleTrailRequest("trailrequest-1", when)
	got, err := ParseTrailRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailRequest() = %v", err)
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

func TestTrailRequestParseRejectsBadInput(t *testing.T) {
	f := SampleTrailRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailRequest(f); err == nil {
		t.Error("ParseTrailRequest() accepted a record with no id")
	}
	f = SampleTrailRequest("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTrailRequest(f); err == nil {
		t.Error("ParseTrailRequest() accepted a non-numeric weight")
	}
}

func TestTrailRequestCollections(t *testing.T) {
	xs := []TrailRequest{SampleTrailRequest("a", when), SampleTrailRequest("bb", when), SampleTrailRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailRequests(xs)); got != 2 {
		t.Errorf("ActiveTrailRequests() kept %d, want 2", got)
	}
	if got, want := SumTrailRequests(xs), 1+3; got != want {
		t.Errorf("SumTrailRequests() = %d, want %d", got, want)
	}
	SortTrailRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailRequests() put %q first", xs[0].ID)
	}
	if len(IndexTrailRequests(xs)) != 3 {
		t.Errorf("IndexTrailRequests() = %v", IndexTrailRequests(xs))
	}
	if err := ValidateTrailRequests(ActiveTrailRequests(xs)); err != nil {
		t.Errorf("ValidateTrailRequests() = %v", err)
	}
}

func TestTrailRequestNormalizeClamps(t *testing.T) {
	x := SampleTrailRequest("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailRequest("r1", when)
	rev, err := r.PutTrailRequest(x, 0)
	if err != nil {
		t.Fatalf("PutTrailRequest() = %v", err)
	}
	got, gotRev, err := r.GetTrailRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailRequests() = %v, %v", all, err)
	}
	if _, err := r.PutTrailRequest(TrailRequest{}, 0); err == nil {
		t.Error("PutTrailRequest() accepted an empty entity")
	}
}

func TestTrailSummaryRoundTrip(t *testing.T) {
	x := SampleTrailSummary("trailsummary-1", when)
	got, err := ParseTrailSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailSummary() = %v", err)
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

func TestTrailSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleTrailSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailSummary(f); err == nil {
		t.Error("ParseTrailSummary() accepted a record with no id")
	}
	f = SampleTrailSummary("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTrailSummary(f); err == nil {
		t.Error("ParseTrailSummary() accepted a non-numeric version")
	}
}

func TestTrailSummaryCollections(t *testing.T) {
	xs := []TrailSummary{SampleTrailSummary("a", when), SampleTrailSummary("bb", when), SampleTrailSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailSummarys(xs)); got != 2 {
		t.Errorf("ActiveTrailSummarys() kept %d, want 2", got)
	}
	if got, want := SumTrailSummarys(xs), 1+3; got != want {
		t.Errorf("SumTrailSummarys() = %d, want %d", got, want)
	}
	SortTrailSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailSummarys() put %q first", xs[0].ID)
	}
	if len(IndexTrailSummarys(xs)) != 3 {
		t.Errorf("IndexTrailSummarys() = %v", IndexTrailSummarys(xs))
	}
	if err := ValidateTrailSummarys(ActiveTrailSummarys(xs)); err != nil {
		t.Errorf("ValidateTrailSummarys() = %v", err)
	}
}

func TestTrailSummaryNormalizeClamps(t *testing.T) {
	x := SampleTrailSummary("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailSummary("r1", when)
	rev, err := r.PutTrailSummary(x, 0)
	if err != nil {
		t.Fatalf("PutTrailSummary() = %v", err)
	}
	got, gotRev, err := r.GetTrailSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutTrailSummary(TrailSummary{}, 0); err == nil {
		t.Error("PutTrailSummary() accepted an empty entity")
	}
}

func TestTrailPolicyRoundTrip(t *testing.T) {
	x := SampleTrailPolicy("trailpolicy-1", when)
	got, err := ParseTrailPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailPolicy() = %v", err)
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

func TestTrailPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleTrailPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailPolicy(f); err == nil {
		t.Error("ParseTrailPolicy() accepted a record with no id")
	}
	f = SampleTrailPolicy("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTrailPolicy(f); err == nil {
		t.Error("ParseTrailPolicy() accepted a non-numeric offset")
	}
}

func TestTrailPolicyCollections(t *testing.T) {
	xs := []TrailPolicy{SampleTrailPolicy("a", when), SampleTrailPolicy("bb", when), SampleTrailPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailPolicys(xs)); got != 2 {
		t.Errorf("ActiveTrailPolicys() kept %d, want 2", got)
	}
	if got, want := SumTrailPolicys(xs), 1+3; got != want {
		t.Errorf("SumTrailPolicys() = %d, want %d", got, want)
	}
	SortTrailPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailPolicys() put %q first", xs[0].ID)
	}
	if len(IndexTrailPolicys(xs)) != 3 {
		t.Errorf("IndexTrailPolicys() = %v", IndexTrailPolicys(xs))
	}
	if err := ValidateTrailPolicys(ActiveTrailPolicys(xs)); err != nil {
		t.Errorf("ValidateTrailPolicys() = %v", err)
	}
}

func TestTrailPolicyNormalizeClamps(t *testing.T) {
	x := SampleTrailPolicy("a", when)
	x.Reference = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailPolicy("r1", when)
	rev, err := r.PutTrailPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutTrailPolicy() = %v", err)
	}
	got, gotRev, err := r.GetTrailPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutTrailPolicy(TrailPolicy{}, 0); err == nil {
		t.Error("PutTrailPolicy() accepted an empty entity")
	}
}

func TestTrailLinkRoundTrip(t *testing.T) {
	x := SampleTrailLink("traillink-1", when)
	got, err := ParseTrailLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailLink() = %v", err)
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

func TestTrailLinkParseRejectsBadInput(t *testing.T) {
	f := SampleTrailLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailLink(f); err == nil {
		t.Error("ParseTrailLink() accepted a record with no id")
	}
	f = SampleTrailLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTrailLink(f); err == nil {
		t.Error("ParseTrailLink() accepted a non-numeric retained")
	}
}

func TestTrailLinkCollections(t *testing.T) {
	xs := []TrailLink{SampleTrailLink("a", when), SampleTrailLink("bb", when), SampleTrailLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailLinks(xs)); got != 2 {
		t.Errorf("ActiveTrailLinks() kept %d, want 2", got)
	}
	if got, want := SumTrailLinks(xs), 1+3; got != want {
		t.Errorf("SumTrailLinks() = %d, want %d", got, want)
	}
	SortTrailLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailLinks() put %q first", xs[0].ID)
	}
	if len(IndexTrailLinks(xs)) != 3 {
		t.Errorf("IndexTrailLinks() = %v", IndexTrailLinks(xs))
	}
	if err := ValidateTrailLinks(ActiveTrailLinks(xs)); err != nil {
		t.Errorf("ValidateTrailLinks() = %v", err)
	}
}

func TestTrailLinkNormalizeClamps(t *testing.T) {
	x := SampleTrailLink("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailLink("r1", when)
	rev, err := r.PutTrailLink(x, 0)
	if err != nil {
		t.Fatalf("PutTrailLink() = %v", err)
	}
	got, gotRev, err := r.GetTrailLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailLinks() = %v, %v", all, err)
	}
	if _, err := r.PutTrailLink(TrailLink{}, 0); err == nil {
		t.Error("PutTrailLink() accepted an empty entity")
	}
}

func TestTrailNoteRoundTrip(t *testing.T) {
	x := SampleTrailNote("trailnote-1", when)
	got, err := ParseTrailNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailNote() = %v", err)
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

func TestTrailNoteParseRejectsBadInput(t *testing.T) {
	f := SampleTrailNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailNote(f); err == nil {
		t.Error("ParseTrailNote() accepted a record with no id")
	}
	f = SampleTrailNote("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTrailNote(f); err == nil {
		t.Error("ParseTrailNote() accepted a non-numeric count")
	}
}

func TestTrailNoteCollections(t *testing.T) {
	xs := []TrailNote{SampleTrailNote("a", when), SampleTrailNote("bb", when), SampleTrailNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailNotes(xs)); got != 2 {
		t.Errorf("ActiveTrailNotes() kept %d, want 2", got)
	}
	if got, want := SumTrailNotes(xs), 1+3; got != want {
		t.Errorf("SumTrailNotes() = %d, want %d", got, want)
	}
	SortTrailNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailNotes() put %q first", xs[0].ID)
	}
	if len(IndexTrailNotes(xs)) != 3 {
		t.Errorf("IndexTrailNotes() = %v", IndexTrailNotes(xs))
	}
	if err := ValidateTrailNotes(ActiveTrailNotes(xs)); err != nil {
		t.Errorf("ValidateTrailNotes() = %v", err)
	}
}

func TestTrailNoteNormalizeClamps(t *testing.T) {
	x := SampleTrailNote("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailNote("r1", when)
	rev, err := r.PutTrailNote(x, 0)
	if err != nil {
		t.Fatalf("PutTrailNote() = %v", err)
	}
	got, gotRev, err := r.GetTrailNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailNotes() = %v, %v", all, err)
	}
	if _, err := r.PutTrailNote(TrailNote{}, 0); err == nil {
		t.Error("PutTrailNote() accepted an empty entity")
	}
}

func TestTrailWindowRoundTrip(t *testing.T) {
	x := SampleTrailWindow("trailwindow-1", when)
	got, err := ParseTrailWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailWindow() = %v", err)
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

func TestTrailWindowParseRejectsBadInput(t *testing.T) {
	f := SampleTrailWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailWindow(f); err == nil {
		t.Error("ParseTrailWindow() accepted a record with no id")
	}
	f = SampleTrailWindow("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTrailWindow(f); err == nil {
		t.Error("ParseTrailWindow() accepted a non-numeric attempts")
	}
}

func TestTrailWindowCollections(t *testing.T) {
	xs := []TrailWindow{SampleTrailWindow("a", when), SampleTrailWindow("bb", when), SampleTrailWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailWindows(xs)); got != 2 {
		t.Errorf("ActiveTrailWindows() kept %d, want 2", got)
	}
	if got, want := SumTrailWindows(xs), 1+3; got != want {
		t.Errorf("SumTrailWindows() = %d, want %d", got, want)
	}
	SortTrailWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailWindows() put %q first", xs[0].ID)
	}
	if len(IndexTrailWindows(xs)) != 3 {
		t.Errorf("IndexTrailWindows() = %v", IndexTrailWindows(xs))
	}
	if err := ValidateTrailWindows(ActiveTrailWindows(xs)); err != nil {
		t.Errorf("ValidateTrailWindows() = %v", err)
	}
}

func TestTrailWindowNormalizeClamps(t *testing.T) {
	x := SampleTrailWindow("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailWindow("r1", when)
	rev, err := r.PutTrailWindow(x, 0)
	if err != nil {
		t.Fatalf("PutTrailWindow() = %v", err)
	}
	got, gotRev, err := r.GetTrailWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailWindows() = %v, %v", all, err)
	}
	if _, err := r.PutTrailWindow(TrailWindow{}, 0); err == nil {
		t.Error("PutTrailWindow() accepted an empty entity")
	}
}

func TestTrailQuotaRoundTrip(t *testing.T) {
	x := SampleTrailQuota("trailquota-1", when)
	got, err := ParseTrailQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailQuota() = %v", err)
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

func TestTrailQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleTrailQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailQuota(f); err == nil {
		t.Error("ParseTrailQuota() accepted a record with no id")
	}
	f = SampleTrailQuota("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTrailQuota(f); err == nil {
		t.Error("ParseTrailQuota() accepted a non-numeric capacity")
	}
}

func TestTrailQuotaCollections(t *testing.T) {
	xs := []TrailQuota{SampleTrailQuota("a", when), SampleTrailQuota("bb", when), SampleTrailQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailQuotas(xs)); got != 2 {
		t.Errorf("ActiveTrailQuotas() kept %d, want 2", got)
	}
	if got, want := SumTrailQuotas(xs), 1+3; got != want {
		t.Errorf("SumTrailQuotas() = %d, want %d", got, want)
	}
	SortTrailQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailQuotas() put %q first", xs[0].ID)
	}
	if len(IndexTrailQuotas(xs)) != 3 {
		t.Errorf("IndexTrailQuotas() = %v", IndexTrailQuotas(xs))
	}
	if err := ValidateTrailQuotas(ActiveTrailQuotas(xs)); err != nil {
		t.Errorf("ValidateTrailQuotas() = %v", err)
	}
}

func TestTrailQuotaNormalizeClamps(t *testing.T) {
	x := SampleTrailQuota("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailQuota("r1", when)
	rev, err := r.PutTrailQuota(x, 0)
	if err != nil {
		t.Fatalf("PutTrailQuota() = %v", err)
	}
	got, gotRev, err := r.GetTrailQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutTrailQuota(TrailQuota{}, 0); err == nil {
		t.Error("PutTrailQuota() accepted an empty entity")
	}
}

func TestTrailBucketRoundTrip(t *testing.T) {
	x := SampleTrailBucket("trailbucket-1", when)
	got, err := ParseTrailBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailBucket() = %v", err)
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

func TestTrailBucketParseRejectsBadInput(t *testing.T) {
	f := SampleTrailBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailBucket(f); err == nil {
		t.Error("ParseTrailBucket() accepted a record with no id")
	}
	f = SampleTrailBucket("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTrailBucket(f); err == nil {
		t.Error("ParseTrailBucket() accepted a non-numeric count")
	}
}

func TestTrailBucketCollections(t *testing.T) {
	xs := []TrailBucket{SampleTrailBucket("a", when), SampleTrailBucket("bb", when), SampleTrailBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailBuckets(xs)); got != 2 {
		t.Errorf("ActiveTrailBuckets() kept %d, want 2", got)
	}
	if got, want := SumTrailBuckets(xs), 1+3; got != want {
		t.Errorf("SumTrailBuckets() = %d, want %d", got, want)
	}
	SortTrailBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailBuckets() put %q first", xs[0].ID)
	}
	if len(IndexTrailBuckets(xs)) != 3 {
		t.Errorf("IndexTrailBuckets() = %v", IndexTrailBuckets(xs))
	}
	if err := ValidateTrailBuckets(ActiveTrailBuckets(xs)); err != nil {
		t.Errorf("ValidateTrailBuckets() = %v", err)
	}
}

func TestTrailBucketNormalizeClamps(t *testing.T) {
	x := SampleTrailBucket("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailBucket("r1", when)
	rev, err := r.PutTrailBucket(x, 0)
	if err != nil {
		t.Fatalf("PutTrailBucket() = %v", err)
	}
	got, gotRev, err := r.GetTrailBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutTrailBucket(TrailBucket{}, 0); err == nil {
		t.Error("PutTrailBucket() accepted an empty entity")
	}
}

func TestTrailLaneRoundTrip(t *testing.T) {
	x := SampleTrailLane("traillane-1", when)
	got, err := ParseTrailLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailLane() = %v", err)
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

func TestTrailLaneParseRejectsBadInput(t *testing.T) {
	f := SampleTrailLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailLane(f); err == nil {
		t.Error("ParseTrailLane() accepted a record with no id")
	}
	f = SampleTrailLane("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTrailLane(f); err == nil {
		t.Error("ParseTrailLane() accepted a non-numeric retained")
	}
}

func TestTrailLaneCollections(t *testing.T) {
	xs := []TrailLane{SampleTrailLane("a", when), SampleTrailLane("bb", when), SampleTrailLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailLanes(xs)); got != 2 {
		t.Errorf("ActiveTrailLanes() kept %d, want 2", got)
	}
	if got, want := SumTrailLanes(xs), 1+3; got != want {
		t.Errorf("SumTrailLanes() = %d, want %d", got, want)
	}
	SortTrailLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailLanes() put %q first", xs[0].ID)
	}
	if len(IndexTrailLanes(xs)) != 3 {
		t.Errorf("IndexTrailLanes() = %v", IndexTrailLanes(xs))
	}
	if err := ValidateTrailLanes(ActiveTrailLanes(xs)); err != nil {
		t.Errorf("ValidateTrailLanes() = %v", err)
	}
}

func TestTrailLaneNormalizeClamps(t *testing.T) {
	x := SampleTrailLane("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailLane("r1", when)
	rev, err := r.PutTrailLane(x, 0)
	if err != nil {
		t.Fatalf("PutTrailLane() = %v", err)
	}
	got, gotRev, err := r.GetTrailLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailLanes() = %v, %v", all, err)
	}
	if _, err := r.PutTrailLane(TrailLane{}, 0); err == nil {
		t.Error("PutTrailLane() accepted an empty entity")
	}
}

func TestTrailSlotRoundTrip(t *testing.T) {
	x := SampleTrailSlot("trailslot-1", when)
	got, err := ParseTrailSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailSlot() = %v", err)
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

func TestTrailSlotParseRejectsBadInput(t *testing.T) {
	f := SampleTrailSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailSlot(f); err == nil {
		t.Error("ParseTrailSlot() accepted a record with no id")
	}
	f = SampleTrailSlot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTrailSlot(f); err == nil {
		t.Error("ParseTrailSlot() accepted a non-numeric weight")
	}
}

func TestTrailSlotCollections(t *testing.T) {
	xs := []TrailSlot{SampleTrailSlot("a", when), SampleTrailSlot("bb", when), SampleTrailSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailSlots(xs)); got != 2 {
		t.Errorf("ActiveTrailSlots() kept %d, want 2", got)
	}
	if got, want := SumTrailSlots(xs), 1+3; got != want {
		t.Errorf("SumTrailSlots() = %d, want %d", got, want)
	}
	SortTrailSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailSlots() put %q first", xs[0].ID)
	}
	if len(IndexTrailSlots(xs)) != 3 {
		t.Errorf("IndexTrailSlots() = %v", IndexTrailSlots(xs))
	}
	if err := ValidateTrailSlots(ActiveTrailSlots(xs)); err != nil {
		t.Errorf("ValidateTrailSlots() = %v", err)
	}
}

func TestTrailSlotNormalizeClamps(t *testing.T) {
	x := SampleTrailSlot("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailSlot("r1", when)
	rev, err := r.PutTrailSlot(x, 0)
	if err != nil {
		t.Fatalf("PutTrailSlot() = %v", err)
	}
	got, gotRev, err := r.GetTrailSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailSlots() = %v, %v", all, err)
	}
	if _, err := r.PutTrailSlot(TrailSlot{}, 0); err == nil {
		t.Error("PutTrailSlot() accepted an empty entity")
	}
}

func TestTrailTagRoundTrip(t *testing.T) {
	x := SampleTrailTag("trailtag-1", when)
	got, err := ParseTrailTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailTag() = %v", err)
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

func TestTrailTagParseRejectsBadInput(t *testing.T) {
	f := SampleTrailTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailTag(f); err == nil {
		t.Error("ParseTrailTag() accepted a record with no id")
	}
	f = SampleTrailTag("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTrailTag(f); err == nil {
		t.Error("ParseTrailTag() accepted a non-numeric retained")
	}
}

func TestTrailTagCollections(t *testing.T) {
	xs := []TrailTag{SampleTrailTag("a", when), SampleTrailTag("bb", when), SampleTrailTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailTags(xs)); got != 2 {
		t.Errorf("ActiveTrailTags() kept %d, want 2", got)
	}
	if got, want := SumTrailTags(xs), 1+3; got != want {
		t.Errorf("SumTrailTags() = %d, want %d", got, want)
	}
	SortTrailTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailTags() put %q first", xs[0].ID)
	}
	if len(IndexTrailTags(xs)) != 3 {
		t.Errorf("IndexTrailTags() = %v", IndexTrailTags(xs))
	}
	if err := ValidateTrailTags(ActiveTrailTags(xs)); err != nil {
		t.Errorf("ValidateTrailTags() = %v", err)
	}
}

func TestTrailTagNormalizeClamps(t *testing.T) {
	x := SampleTrailTag("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailTag("r1", when)
	rev, err := r.PutTrailTag(x, 0)
	if err != nil {
		t.Fatalf("PutTrailTag() = %v", err)
	}
	got, gotRev, err := r.GetTrailTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailTags() = %v, %v", all, err)
	}
	if _, err := r.PutTrailTag(TrailTag{}, 0); err == nil {
		t.Error("PutTrailTag() accepted an empty entity")
	}
}

func TestTrailRuleRoundTrip(t *testing.T) {
	x := SampleTrailRule("trailrule-1", when)
	got, err := ParseTrailRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailRule() = %v", err)
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

func TestTrailRuleParseRejectsBadInput(t *testing.T) {
	f := SampleTrailRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailRule(f); err == nil {
		t.Error("ParseTrailRule() accepted a record with no id")
	}
	f = SampleTrailRule("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTrailRule(f); err == nil {
		t.Error("ParseTrailRule() accepted a non-numeric retained")
	}
}

func TestTrailRuleCollections(t *testing.T) {
	xs := []TrailRule{SampleTrailRule("a", when), SampleTrailRule("bb", when), SampleTrailRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailRules(xs)); got != 2 {
		t.Errorf("ActiveTrailRules() kept %d, want 2", got)
	}
	if got, want := SumTrailRules(xs), 1+3; got != want {
		t.Errorf("SumTrailRules() = %d, want %d", got, want)
	}
	SortTrailRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailRules() put %q first", xs[0].ID)
	}
	if len(IndexTrailRules(xs)) != 3 {
		t.Errorf("IndexTrailRules() = %v", IndexTrailRules(xs))
	}
	if err := ValidateTrailRules(ActiveTrailRules(xs)); err != nil {
		t.Errorf("ValidateTrailRules() = %v", err)
	}
}

func TestTrailRuleNormalizeClamps(t *testing.T) {
	x := SampleTrailRule("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailRule("r1", when)
	rev, err := r.PutTrailRule(x, 0)
	if err != nil {
		t.Fatalf("PutTrailRule() = %v", err)
	}
	got, gotRev, err := r.GetTrailRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailRules() = %v, %v", all, err)
	}
	if _, err := r.PutTrailRule(TrailRule{}, 0); err == nil {
		t.Error("PutTrailRule() accepted an empty entity")
	}
}

func TestTrailHoldRoundTrip(t *testing.T) {
	x := SampleTrailHold("trailhold-1", when)
	got, err := ParseTrailHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailHold() = %v", err)
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

func TestTrailHoldParseRejectsBadInput(t *testing.T) {
	f := SampleTrailHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailHold(f); err == nil {
		t.Error("ParseTrailHold() accepted a record with no id")
	}
	f = SampleTrailHold("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTrailHold(f); err == nil {
		t.Error("ParseTrailHold() accepted a non-numeric priority")
	}
}

func TestTrailHoldCollections(t *testing.T) {
	xs := []TrailHold{SampleTrailHold("a", when), SampleTrailHold("bb", when), SampleTrailHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailHolds(xs)); got != 2 {
		t.Errorf("ActiveTrailHolds() kept %d, want 2", got)
	}
	if got, want := SumTrailHolds(xs), 1+3; got != want {
		t.Errorf("SumTrailHolds() = %d, want %d", got, want)
	}
	SortTrailHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailHolds() put %q first", xs[0].ID)
	}
	if len(IndexTrailHolds(xs)) != 3 {
		t.Errorf("IndexTrailHolds() = %v", IndexTrailHolds(xs))
	}
	if err := ValidateTrailHolds(ActiveTrailHolds(xs)); err != nil {
		t.Errorf("ValidateTrailHolds() = %v", err)
	}
}

func TestTrailHoldNormalizeClamps(t *testing.T) {
	x := SampleTrailHold("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailHold("r1", when)
	rev, err := r.PutTrailHold(x, 0)
	if err != nil {
		t.Fatalf("PutTrailHold() = %v", err)
	}
	got, gotRev, err := r.GetTrailHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailHolds() = %v, %v", all, err)
	}
	if _, err := r.PutTrailHold(TrailHold{}, 0); err == nil {
		t.Error("PutTrailHold() accepted an empty entity")
	}
}

func TestTrailClaimRoundTrip(t *testing.T) {
	x := SampleTrailClaim("trailclaim-1", when)
	got, err := ParseTrailClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailClaim() = %v", err)
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

func TestTrailClaimParseRejectsBadInput(t *testing.T) {
	f := SampleTrailClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailClaim(f); err == nil {
		t.Error("ParseTrailClaim() accepted a record with no id")
	}
	f = SampleTrailClaim("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTrailClaim(f); err == nil {
		t.Error("ParseTrailClaim() accepted a non-numeric count")
	}
}

func TestTrailClaimCollections(t *testing.T) {
	xs := []TrailClaim{SampleTrailClaim("a", when), SampleTrailClaim("bb", when), SampleTrailClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailClaims(xs)); got != 2 {
		t.Errorf("ActiveTrailClaims() kept %d, want 2", got)
	}
	if got, want := SumTrailClaims(xs), 1+3; got != want {
		t.Errorf("SumTrailClaims() = %d, want %d", got, want)
	}
	SortTrailClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailClaims() put %q first", xs[0].ID)
	}
	if len(IndexTrailClaims(xs)) != 3 {
		t.Errorf("IndexTrailClaims() = %v", IndexTrailClaims(xs))
	}
	if err := ValidateTrailClaims(ActiveTrailClaims(xs)); err != nil {
		t.Errorf("ValidateTrailClaims() = %v", err)
	}
}

func TestTrailClaimNormalizeClamps(t *testing.T) {
	x := SampleTrailClaim("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailClaim("r1", when)
	rev, err := r.PutTrailClaim(x, 0)
	if err != nil {
		t.Fatalf("PutTrailClaim() = %v", err)
	}
	got, gotRev, err := r.GetTrailClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailClaims() = %v, %v", all, err)
	}
	if _, err := r.PutTrailClaim(TrailClaim{}, 0); err == nil {
		t.Error("PutTrailClaim() accepted an empty entity")
	}
}

func TestTrailCursorRoundTrip(t *testing.T) {
	x := SampleTrailCursor("trailcursor-1", when)
	got, err := ParseTrailCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseTrailCursor() = %v", err)
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

func TestTrailCursorParseRejectsBadInput(t *testing.T) {
	f := SampleTrailCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTrailCursor(f); err == nil {
		t.Error("ParseTrailCursor() accepted a record with no id")
	}
	f = SampleTrailCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTrailCursor(f); err == nil {
		t.Error("ParseTrailCursor() accepted a non-numeric version")
	}
}

func TestTrailCursorCollections(t *testing.T) {
	xs := []TrailCursor{SampleTrailCursor("a", when), SampleTrailCursor("bb", when), SampleTrailCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTrailCursors(xs)); got != 2 {
		t.Errorf("ActiveTrailCursors() kept %d, want 2", got)
	}
	if got, want := SumTrailCursors(xs), 1+3; got != want {
		t.Errorf("SumTrailCursors() = %d, want %d", got, want)
	}
	SortTrailCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTrailCursors() put %q first", xs[0].ID)
	}
	if len(IndexTrailCursors(xs)) != 3 {
		t.Errorf("IndexTrailCursors() = %v", IndexTrailCursors(xs))
	}
	if err := ValidateTrailCursors(ActiveTrailCursors(xs)); err != nil {
		t.Errorf("ValidateTrailCursors() = %v", err)
	}
}

func TestTrailCursorNormalizeClamps(t *testing.T) {
	x := SampleTrailCursor("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTrailCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTrailCursor("r1", when)
	rev, err := r.PutTrailCursor(x, 0)
	if err != nil {
		t.Fatalf("PutTrailCursor() = %v", err)
	}
	got, gotRev, err := r.GetTrailCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTrailCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTrailCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTrailCursors() = %v, %v", all, err)
	}
	if _, err := r.PutTrailCursor(TrailCursor{}, 0); err == nil {
		t.Error("PutTrailCursor() accepted an empty entity")
	}
}
