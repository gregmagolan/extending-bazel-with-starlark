package digest

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
		if !strings.HasPrefix(Prefixes[i], "digest/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleDigestRecord("ok", when).Fields()); err != nil {
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

func TestDigestRecordRoundTrip(t *testing.T) {
	x := SampleDigestRecord("digestrecord-1", when)
	got, err := ParseDigestRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestRecord() = %v", err)
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

func TestDigestRecordParseRejectsBadInput(t *testing.T) {
	f := SampleDigestRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestRecord(f); err == nil {
		t.Error("ParseDigestRecord() accepted a record with no id")
	}
	f = SampleDigestRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDigestRecord(f); err == nil {
		t.Error("ParseDigestRecord() accepted a non-numeric capacity")
	}
}

func TestDigestRecordCollections(t *testing.T) {
	xs := []DigestRecord{SampleDigestRecord("a", when), SampleDigestRecord("bb", when), SampleDigestRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestRecords(xs)); got != 2 {
		t.Errorf("ActiveDigestRecords() kept %d, want 2", got)
	}
	if got, want := SumDigestRecords(xs), 1+3; got != want {
		t.Errorf("SumDigestRecords() = %d, want %d", got, want)
	}
	SortDigestRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestRecords() put %q first", xs[0].ID)
	}
	if len(IndexDigestRecords(xs)) != 3 {
		t.Errorf("IndexDigestRecords() = %v", IndexDigestRecords(xs))
	}
	if err := ValidateDigestRecords(ActiveDigestRecords(xs)); err != nil {
		t.Errorf("ValidateDigestRecords() = %v", err)
	}
}

func TestDigestRecordNormalizeClamps(t *testing.T) {
	x := SampleDigestRecord("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestRecord("r1", when)
	rev, err := r.PutDigestRecord(x, 0)
	if err != nil {
		t.Fatalf("PutDigestRecord() = %v", err)
	}
	got, gotRev, err := r.GetDigestRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestRecords() = %v, %v", all, err)
	}
	if _, err := r.PutDigestRecord(DigestRecord{}, 0); err == nil {
		t.Error("PutDigestRecord() accepted an empty entity")
	}
}

func TestDigestEntryRoundTrip(t *testing.T) {
	x := SampleDigestEntry("digestentry-1", when)
	got, err := ParseDigestEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestEntry() = %v", err)
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

func TestDigestEntryParseRejectsBadInput(t *testing.T) {
	f := SampleDigestEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestEntry(f); err == nil {
		t.Error("ParseDigestEntry() accepted a record with no id")
	}
	f = SampleDigestEntry("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDigestEntry(f); err == nil {
		t.Error("ParseDigestEntry() accepted a non-numeric priority")
	}
}

func TestDigestEntryCollections(t *testing.T) {
	xs := []DigestEntry{SampleDigestEntry("a", when), SampleDigestEntry("bb", when), SampleDigestEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestEntrys(xs)); got != 2 {
		t.Errorf("ActiveDigestEntrys() kept %d, want 2", got)
	}
	if got, want := SumDigestEntrys(xs), 1+3; got != want {
		t.Errorf("SumDigestEntrys() = %d, want %d", got, want)
	}
	SortDigestEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestEntrys() put %q first", xs[0].ID)
	}
	if len(IndexDigestEntrys(xs)) != 3 {
		t.Errorf("IndexDigestEntrys() = %v", IndexDigestEntrys(xs))
	}
	if err := ValidateDigestEntrys(ActiveDigestEntrys(xs)); err != nil {
		t.Errorf("ValidateDigestEntrys() = %v", err)
	}
}

func TestDigestEntryNormalizeClamps(t *testing.T) {
	x := SampleDigestEntry("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestEntry("r1", when)
	rev, err := r.PutDigestEntry(x, 0)
	if err != nil {
		t.Fatalf("PutDigestEntry() = %v", err)
	}
	got, gotRev, err := r.GetDigestEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutDigestEntry(DigestEntry{}, 0); err == nil {
		t.Error("PutDigestEntry() accepted an empty entity")
	}
}

func TestDigestItemRoundTrip(t *testing.T) {
	x := SampleDigestItem("digestitem-1", when)
	got, err := ParseDigestItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestItem() = %v", err)
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

func TestDigestItemParseRejectsBadInput(t *testing.T) {
	f := SampleDigestItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestItem(f); err == nil {
		t.Error("ParseDigestItem() accepted a record with no id")
	}
	f = SampleDigestItem("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDigestItem(f); err == nil {
		t.Error("ParseDigestItem() accepted a non-numeric weight")
	}
}

func TestDigestItemCollections(t *testing.T) {
	xs := []DigestItem{SampleDigestItem("a", when), SampleDigestItem("bb", when), SampleDigestItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestItems(xs)); got != 2 {
		t.Errorf("ActiveDigestItems() kept %d, want 2", got)
	}
	if got, want := SumDigestItems(xs), 1+3; got != want {
		t.Errorf("SumDigestItems() = %d, want %d", got, want)
	}
	SortDigestItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestItems() put %q first", xs[0].ID)
	}
	if len(IndexDigestItems(xs)) != 3 {
		t.Errorf("IndexDigestItems() = %v", IndexDigestItems(xs))
	}
	if err := ValidateDigestItems(ActiveDigestItems(xs)); err != nil {
		t.Errorf("ValidateDigestItems() = %v", err)
	}
}

func TestDigestItemNormalizeClamps(t *testing.T) {
	x := SampleDigestItem("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestItem("r1", when)
	rev, err := r.PutDigestItem(x, 0)
	if err != nil {
		t.Fatalf("PutDigestItem() = %v", err)
	}
	got, gotRev, err := r.GetDigestItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestItems() = %v, %v", all, err)
	}
	if _, err := r.PutDigestItem(DigestItem{}, 0); err == nil {
		t.Error("PutDigestItem() accepted an empty entity")
	}
}

func TestDigestDraftRoundTrip(t *testing.T) {
	x := SampleDigestDraft("digestdraft-1", when)
	got, err := ParseDigestDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestDraft() = %v", err)
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

func TestDigestDraftParseRejectsBadInput(t *testing.T) {
	f := SampleDigestDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestDraft(f); err == nil {
		t.Error("ParseDigestDraft() accepted a record with no id")
	}
	f = SampleDigestDraft("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDigestDraft(f); err == nil {
		t.Error("ParseDigestDraft() accepted a non-numeric weight")
	}
}

func TestDigestDraftCollections(t *testing.T) {
	xs := []DigestDraft{SampleDigestDraft("a", when), SampleDigestDraft("bb", when), SampleDigestDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestDrafts(xs)); got != 2 {
		t.Errorf("ActiveDigestDrafts() kept %d, want 2", got)
	}
	if got, want := SumDigestDrafts(xs), 1+3; got != want {
		t.Errorf("SumDigestDrafts() = %d, want %d", got, want)
	}
	SortDigestDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestDrafts() put %q first", xs[0].ID)
	}
	if len(IndexDigestDrafts(xs)) != 3 {
		t.Errorf("IndexDigestDrafts() = %v", IndexDigestDrafts(xs))
	}
	if err := ValidateDigestDrafts(ActiveDigestDrafts(xs)); err != nil {
		t.Errorf("ValidateDigestDrafts() = %v", err)
	}
}

func TestDigestDraftNormalizeClamps(t *testing.T) {
	x := SampleDigestDraft("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestDraft("r1", when)
	rev, err := r.PutDigestDraft(x, 0)
	if err != nil {
		t.Fatalf("PutDigestDraft() = %v", err)
	}
	got, gotRev, err := r.GetDigestDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutDigestDraft(DigestDraft{}, 0); err == nil {
		t.Error("PutDigestDraft() accepted an empty entity")
	}
}

func TestDigestRevisionRoundTrip(t *testing.T) {
	x := SampleDigestRevision("digestrevision-1", when)
	got, err := ParseDigestRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestRevision() = %v", err)
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

func TestDigestRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleDigestRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestRevision(f); err == nil {
		t.Error("ParseDigestRevision() accepted a record with no id")
	}
	f = SampleDigestRevision("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDigestRevision(f); err == nil {
		t.Error("ParseDigestRevision() accepted a non-numeric weight")
	}
}

func TestDigestRevisionCollections(t *testing.T) {
	xs := []DigestRevision{SampleDigestRevision("a", when), SampleDigestRevision("bb", when), SampleDigestRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestRevisions(xs)); got != 2 {
		t.Errorf("ActiveDigestRevisions() kept %d, want 2", got)
	}
	if got, want := SumDigestRevisions(xs), 1+3; got != want {
		t.Errorf("SumDigestRevisions() = %d, want %d", got, want)
	}
	SortDigestRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestRevisions() put %q first", xs[0].ID)
	}
	if len(IndexDigestRevisions(xs)) != 3 {
		t.Errorf("IndexDigestRevisions() = %v", IndexDigestRevisions(xs))
	}
	if err := ValidateDigestRevisions(ActiveDigestRevisions(xs)); err != nil {
		t.Errorf("ValidateDigestRevisions() = %v", err)
	}
}

func TestDigestRevisionNormalizeClamps(t *testing.T) {
	x := SampleDigestRevision("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestRevision("r1", when)
	rev, err := r.PutDigestRevision(x, 0)
	if err != nil {
		t.Fatalf("PutDigestRevision() = %v", err)
	}
	got, gotRev, err := r.GetDigestRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutDigestRevision(DigestRevision{}, 0); err == nil {
		t.Error("PutDigestRevision() accepted an empty entity")
	}
}

func TestDigestSnapshotRoundTrip(t *testing.T) {
	x := SampleDigestSnapshot("digestsnapshot-1", when)
	got, err := ParseDigestSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestSnapshot() = %v", err)
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

func TestDigestSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleDigestSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestSnapshot(f); err == nil {
		t.Error("ParseDigestSnapshot() accepted a record with no id")
	}
	f = SampleDigestSnapshot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDigestSnapshot(f); err == nil {
		t.Error("ParseDigestSnapshot() accepted a non-numeric weight")
	}
}

func TestDigestSnapshotCollections(t *testing.T) {
	xs := []DigestSnapshot{SampleDigestSnapshot("a", when), SampleDigestSnapshot("bb", when), SampleDigestSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestSnapshots(xs)); got != 2 {
		t.Errorf("ActiveDigestSnapshots() kept %d, want 2", got)
	}
	if got, want := SumDigestSnapshots(xs), 1+3; got != want {
		t.Errorf("SumDigestSnapshots() = %d, want %d", got, want)
	}
	SortDigestSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexDigestSnapshots(xs)) != 3 {
		t.Errorf("IndexDigestSnapshots() = %v", IndexDigestSnapshots(xs))
	}
	if err := ValidateDigestSnapshots(ActiveDigestSnapshots(xs)); err != nil {
		t.Errorf("ValidateDigestSnapshots() = %v", err)
	}
}

func TestDigestSnapshotNormalizeClamps(t *testing.T) {
	x := SampleDigestSnapshot("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestSnapshot("r1", when)
	rev, err := r.PutDigestSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutDigestSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetDigestSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutDigestSnapshot(DigestSnapshot{}, 0); err == nil {
		t.Error("PutDigestSnapshot() accepted an empty entity")
	}
}

func TestDigestBatchRoundTrip(t *testing.T) {
	x := SampleDigestBatch("digestbatch-1", when)
	got, err := ParseDigestBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestBatch() = %v", err)
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

func TestDigestBatchParseRejectsBadInput(t *testing.T) {
	f := SampleDigestBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestBatch(f); err == nil {
		t.Error("ParseDigestBatch() accepted a record with no id")
	}
	f = SampleDigestBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDigestBatch(f); err == nil {
		t.Error("ParseDigestBatch() accepted a non-numeric version")
	}
}

func TestDigestBatchCollections(t *testing.T) {
	xs := []DigestBatch{SampleDigestBatch("a", when), SampleDigestBatch("bb", when), SampleDigestBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestBatchs(xs)); got != 2 {
		t.Errorf("ActiveDigestBatchs() kept %d, want 2", got)
	}
	if got, want := SumDigestBatchs(xs), 1+3; got != want {
		t.Errorf("SumDigestBatchs() = %d, want %d", got, want)
	}
	SortDigestBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestBatchs() put %q first", xs[0].ID)
	}
	if len(IndexDigestBatchs(xs)) != 3 {
		t.Errorf("IndexDigestBatchs() = %v", IndexDigestBatchs(xs))
	}
	if err := ValidateDigestBatchs(ActiveDigestBatchs(xs)); err != nil {
		t.Errorf("ValidateDigestBatchs() = %v", err)
	}
}

func TestDigestBatchNormalizeClamps(t *testing.T) {
	x := SampleDigestBatch("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestBatch("r1", when)
	rev, err := r.PutDigestBatch(x, 0)
	if err != nil {
		t.Fatalf("PutDigestBatch() = %v", err)
	}
	got, gotRev, err := r.GetDigestBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutDigestBatch(DigestBatch{}, 0); err == nil {
		t.Error("PutDigestBatch() accepted an empty entity")
	}
}

func TestDigestRequestRoundTrip(t *testing.T) {
	x := SampleDigestRequest("digestrequest-1", when)
	got, err := ParseDigestRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestRequest() = %v", err)
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

func TestDigestRequestParseRejectsBadInput(t *testing.T) {
	f := SampleDigestRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestRequest(f); err == nil {
		t.Error("ParseDigestRequest() accepted a record with no id")
	}
	f = SampleDigestRequest("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDigestRequest(f); err == nil {
		t.Error("ParseDigestRequest() accepted a non-numeric capacity")
	}
}

func TestDigestRequestCollections(t *testing.T) {
	xs := []DigestRequest{SampleDigestRequest("a", when), SampleDigestRequest("bb", when), SampleDigestRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestRequests(xs)); got != 2 {
		t.Errorf("ActiveDigestRequests() kept %d, want 2", got)
	}
	if got, want := SumDigestRequests(xs), 1+3; got != want {
		t.Errorf("SumDigestRequests() = %d, want %d", got, want)
	}
	SortDigestRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestRequests() put %q first", xs[0].ID)
	}
	if len(IndexDigestRequests(xs)) != 3 {
		t.Errorf("IndexDigestRequests() = %v", IndexDigestRequests(xs))
	}
	if err := ValidateDigestRequests(ActiveDigestRequests(xs)); err != nil {
		t.Errorf("ValidateDigestRequests() = %v", err)
	}
}

func TestDigestRequestNormalizeClamps(t *testing.T) {
	x := SampleDigestRequest("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestRequest("r1", when)
	rev, err := r.PutDigestRequest(x, 0)
	if err != nil {
		t.Fatalf("PutDigestRequest() = %v", err)
	}
	got, gotRev, err := r.GetDigestRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestRequests() = %v, %v", all, err)
	}
	if _, err := r.PutDigestRequest(DigestRequest{}, 0); err == nil {
		t.Error("PutDigestRequest() accepted an empty entity")
	}
}

func TestDigestSummaryRoundTrip(t *testing.T) {
	x := SampleDigestSummary("digestsummary-1", when)
	got, err := ParseDigestSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestSummary() = %v", err)
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

func TestDigestSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleDigestSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestSummary(f); err == nil {
		t.Error("ParseDigestSummary() accepted a record with no id")
	}
	f = SampleDigestSummary("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDigestSummary(f); err == nil {
		t.Error("ParseDigestSummary() accepted a non-numeric attempts")
	}
}

func TestDigestSummaryCollections(t *testing.T) {
	xs := []DigestSummary{SampleDigestSummary("a", when), SampleDigestSummary("bb", when), SampleDigestSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestSummarys(xs)); got != 2 {
		t.Errorf("ActiveDigestSummarys() kept %d, want 2", got)
	}
	if got, want := SumDigestSummarys(xs), 1+3; got != want {
		t.Errorf("SumDigestSummarys() = %d, want %d", got, want)
	}
	SortDigestSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestSummarys() put %q first", xs[0].ID)
	}
	if len(IndexDigestSummarys(xs)) != 3 {
		t.Errorf("IndexDigestSummarys() = %v", IndexDigestSummarys(xs))
	}
	if err := ValidateDigestSummarys(ActiveDigestSummarys(xs)); err != nil {
		t.Errorf("ValidateDigestSummarys() = %v", err)
	}
}

func TestDigestSummaryNormalizeClamps(t *testing.T) {
	x := SampleDigestSummary("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestSummary("r1", when)
	rev, err := r.PutDigestSummary(x, 0)
	if err != nil {
		t.Fatalf("PutDigestSummary() = %v", err)
	}
	got, gotRev, err := r.GetDigestSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutDigestSummary(DigestSummary{}, 0); err == nil {
		t.Error("PutDigestSummary() accepted an empty entity")
	}
}

func TestDigestPolicyRoundTrip(t *testing.T) {
	x := SampleDigestPolicy("digestpolicy-1", when)
	got, err := ParseDigestPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestPolicy() = %v", err)
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

func TestDigestPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleDigestPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestPolicy(f); err == nil {
		t.Error("ParseDigestPolicy() accepted a record with no id")
	}
	f = SampleDigestPolicy("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDigestPolicy(f); err == nil {
		t.Error("ParseDigestPolicy() accepted a non-numeric weight")
	}
}

func TestDigestPolicyCollections(t *testing.T) {
	xs := []DigestPolicy{SampleDigestPolicy("a", when), SampleDigestPolicy("bb", when), SampleDigestPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestPolicys(xs)); got != 2 {
		t.Errorf("ActiveDigestPolicys() kept %d, want 2", got)
	}
	if got, want := SumDigestPolicys(xs), 1+3; got != want {
		t.Errorf("SumDigestPolicys() = %d, want %d", got, want)
	}
	SortDigestPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestPolicys() put %q first", xs[0].ID)
	}
	if len(IndexDigestPolicys(xs)) != 3 {
		t.Errorf("IndexDigestPolicys() = %v", IndexDigestPolicys(xs))
	}
	if err := ValidateDigestPolicys(ActiveDigestPolicys(xs)); err != nil {
		t.Errorf("ValidateDigestPolicys() = %v", err)
	}
}

func TestDigestPolicyNormalizeClamps(t *testing.T) {
	x := SampleDigestPolicy("a", when)
	x.Owner = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestPolicy("r1", when)
	rev, err := r.PutDigestPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutDigestPolicy() = %v", err)
	}
	got, gotRev, err := r.GetDigestPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutDigestPolicy(DigestPolicy{}, 0); err == nil {
		t.Error("PutDigestPolicy() accepted an empty entity")
	}
}

func TestDigestLinkRoundTrip(t *testing.T) {
	x := SampleDigestLink("digestlink-1", when)
	got, err := ParseDigestLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestLink() = %v", err)
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

func TestDigestLinkParseRejectsBadInput(t *testing.T) {
	f := SampleDigestLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestLink(f); err == nil {
		t.Error("ParseDigestLink() accepted a record with no id")
	}
	f = SampleDigestLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDigestLink(f); err == nil {
		t.Error("ParseDigestLink() accepted a non-numeric retained")
	}
}

func TestDigestLinkCollections(t *testing.T) {
	xs := []DigestLink{SampleDigestLink("a", when), SampleDigestLink("bb", when), SampleDigestLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestLinks(xs)); got != 2 {
		t.Errorf("ActiveDigestLinks() kept %d, want 2", got)
	}
	if got, want := SumDigestLinks(xs), 1+3; got != want {
		t.Errorf("SumDigestLinks() = %d, want %d", got, want)
	}
	SortDigestLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestLinks() put %q first", xs[0].ID)
	}
	if len(IndexDigestLinks(xs)) != 3 {
		t.Errorf("IndexDigestLinks() = %v", IndexDigestLinks(xs))
	}
	if err := ValidateDigestLinks(ActiveDigestLinks(xs)); err != nil {
		t.Errorf("ValidateDigestLinks() = %v", err)
	}
}

func TestDigestLinkNormalizeClamps(t *testing.T) {
	x := SampleDigestLink("a", when)
	x.Reference = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestLink("r1", when)
	rev, err := r.PutDigestLink(x, 0)
	if err != nil {
		t.Fatalf("PutDigestLink() = %v", err)
	}
	got, gotRev, err := r.GetDigestLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestLinks() = %v, %v", all, err)
	}
	if _, err := r.PutDigestLink(DigestLink{}, 0); err == nil {
		t.Error("PutDigestLink() accepted an empty entity")
	}
}

func TestDigestNoteRoundTrip(t *testing.T) {
	x := SampleDigestNote("digestnote-1", when)
	got, err := ParseDigestNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestNote() = %v", err)
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

func TestDigestNoteParseRejectsBadInput(t *testing.T) {
	f := SampleDigestNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestNote(f); err == nil {
		t.Error("ParseDigestNote() accepted a record with no id")
	}
	f = SampleDigestNote("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDigestNote(f); err == nil {
		t.Error("ParseDigestNote() accepted a non-numeric capacity")
	}
}

func TestDigestNoteCollections(t *testing.T) {
	xs := []DigestNote{SampleDigestNote("a", when), SampleDigestNote("bb", when), SampleDigestNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestNotes(xs)); got != 2 {
		t.Errorf("ActiveDigestNotes() kept %d, want 2", got)
	}
	if got, want := SumDigestNotes(xs), 1+3; got != want {
		t.Errorf("SumDigestNotes() = %d, want %d", got, want)
	}
	SortDigestNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestNotes() put %q first", xs[0].ID)
	}
	if len(IndexDigestNotes(xs)) != 3 {
		t.Errorf("IndexDigestNotes() = %v", IndexDigestNotes(xs))
	}
	if err := ValidateDigestNotes(ActiveDigestNotes(xs)); err != nil {
		t.Errorf("ValidateDigestNotes() = %v", err)
	}
}

func TestDigestNoteNormalizeClamps(t *testing.T) {
	x := SampleDigestNote("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestNote("r1", when)
	rev, err := r.PutDigestNote(x, 0)
	if err != nil {
		t.Fatalf("PutDigestNote() = %v", err)
	}
	got, gotRev, err := r.GetDigestNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestNotes() = %v, %v", all, err)
	}
	if _, err := r.PutDigestNote(DigestNote{}, 0); err == nil {
		t.Error("PutDigestNote() accepted an empty entity")
	}
}

func TestDigestWindowRoundTrip(t *testing.T) {
	x := SampleDigestWindow("digestwindow-1", when)
	got, err := ParseDigestWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestWindow() = %v", err)
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

func TestDigestWindowParseRejectsBadInput(t *testing.T) {
	f := SampleDigestWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestWindow(f); err == nil {
		t.Error("ParseDigestWindow() accepted a record with no id")
	}
	f = SampleDigestWindow("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDigestWindow(f); err == nil {
		t.Error("ParseDigestWindow() accepted a non-numeric retained")
	}
}

func TestDigestWindowCollections(t *testing.T) {
	xs := []DigestWindow{SampleDigestWindow("a", when), SampleDigestWindow("bb", when), SampleDigestWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestWindows(xs)); got != 2 {
		t.Errorf("ActiveDigestWindows() kept %d, want 2", got)
	}
	if got, want := SumDigestWindows(xs), 1+3; got != want {
		t.Errorf("SumDigestWindows() = %d, want %d", got, want)
	}
	SortDigestWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestWindows() put %q first", xs[0].ID)
	}
	if len(IndexDigestWindows(xs)) != 3 {
		t.Errorf("IndexDigestWindows() = %v", IndexDigestWindows(xs))
	}
	if err := ValidateDigestWindows(ActiveDigestWindows(xs)); err != nil {
		t.Errorf("ValidateDigestWindows() = %v", err)
	}
}

func TestDigestWindowNormalizeClamps(t *testing.T) {
	x := SampleDigestWindow("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestWindow("r1", when)
	rev, err := r.PutDigestWindow(x, 0)
	if err != nil {
		t.Fatalf("PutDigestWindow() = %v", err)
	}
	got, gotRev, err := r.GetDigestWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestWindows() = %v, %v", all, err)
	}
	if _, err := r.PutDigestWindow(DigestWindow{}, 0); err == nil {
		t.Error("PutDigestWindow() accepted an empty entity")
	}
}

func TestDigestQuotaRoundTrip(t *testing.T) {
	x := SampleDigestQuota("digestquota-1", when)
	got, err := ParseDigestQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestQuota() = %v", err)
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

func TestDigestQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleDigestQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestQuota(f); err == nil {
		t.Error("ParseDigestQuota() accepted a record with no id")
	}
	f = SampleDigestQuota("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDigestQuota(f); err == nil {
		t.Error("ParseDigestQuota() accepted a non-numeric count")
	}
}

func TestDigestQuotaCollections(t *testing.T) {
	xs := []DigestQuota{SampleDigestQuota("a", when), SampleDigestQuota("bb", when), SampleDigestQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestQuotas(xs)); got != 2 {
		t.Errorf("ActiveDigestQuotas() kept %d, want 2", got)
	}
	if got, want := SumDigestQuotas(xs), 1+3; got != want {
		t.Errorf("SumDigestQuotas() = %d, want %d", got, want)
	}
	SortDigestQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestQuotas() put %q first", xs[0].ID)
	}
	if len(IndexDigestQuotas(xs)) != 3 {
		t.Errorf("IndexDigestQuotas() = %v", IndexDigestQuotas(xs))
	}
	if err := ValidateDigestQuotas(ActiveDigestQuotas(xs)); err != nil {
		t.Errorf("ValidateDigestQuotas() = %v", err)
	}
}

func TestDigestQuotaNormalizeClamps(t *testing.T) {
	x := SampleDigestQuota("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestQuota("r1", when)
	rev, err := r.PutDigestQuota(x, 0)
	if err != nil {
		t.Fatalf("PutDigestQuota() = %v", err)
	}
	got, gotRev, err := r.GetDigestQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutDigestQuota(DigestQuota{}, 0); err == nil {
		t.Error("PutDigestQuota() accepted an empty entity")
	}
}

func TestDigestBucketRoundTrip(t *testing.T) {
	x := SampleDigestBucket("digestbucket-1", when)
	got, err := ParseDigestBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestBucket() = %v", err)
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

func TestDigestBucketParseRejectsBadInput(t *testing.T) {
	f := SampleDigestBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestBucket(f); err == nil {
		t.Error("ParseDigestBucket() accepted a record with no id")
	}
	f = SampleDigestBucket("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDigestBucket(f); err == nil {
		t.Error("ParseDigestBucket() accepted a non-numeric priority")
	}
}

func TestDigestBucketCollections(t *testing.T) {
	xs := []DigestBucket{SampleDigestBucket("a", when), SampleDigestBucket("bb", when), SampleDigestBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestBuckets(xs)); got != 2 {
		t.Errorf("ActiveDigestBuckets() kept %d, want 2", got)
	}
	if got, want := SumDigestBuckets(xs), 1+3; got != want {
		t.Errorf("SumDigestBuckets() = %d, want %d", got, want)
	}
	SortDigestBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestBuckets() put %q first", xs[0].ID)
	}
	if len(IndexDigestBuckets(xs)) != 3 {
		t.Errorf("IndexDigestBuckets() = %v", IndexDigestBuckets(xs))
	}
	if err := ValidateDigestBuckets(ActiveDigestBuckets(xs)); err != nil {
		t.Errorf("ValidateDigestBuckets() = %v", err)
	}
}

func TestDigestBucketNormalizeClamps(t *testing.T) {
	x := SampleDigestBucket("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestBucket("r1", when)
	rev, err := r.PutDigestBucket(x, 0)
	if err != nil {
		t.Fatalf("PutDigestBucket() = %v", err)
	}
	got, gotRev, err := r.GetDigestBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutDigestBucket(DigestBucket{}, 0); err == nil {
		t.Error("PutDigestBucket() accepted an empty entity")
	}
}

func TestDigestLaneRoundTrip(t *testing.T) {
	x := SampleDigestLane("digestlane-1", when)
	got, err := ParseDigestLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestLane() = %v", err)
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

func TestDigestLaneParseRejectsBadInput(t *testing.T) {
	f := SampleDigestLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestLane(f); err == nil {
		t.Error("ParseDigestLane() accepted a record with no id")
	}
	f = SampleDigestLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDigestLane(f); err == nil {
		t.Error("ParseDigestLane() accepted a non-numeric priority")
	}
}

func TestDigestLaneCollections(t *testing.T) {
	xs := []DigestLane{SampleDigestLane("a", when), SampleDigestLane("bb", when), SampleDigestLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestLanes(xs)); got != 2 {
		t.Errorf("ActiveDigestLanes() kept %d, want 2", got)
	}
	if got, want := SumDigestLanes(xs), 1+3; got != want {
		t.Errorf("SumDigestLanes() = %d, want %d", got, want)
	}
	SortDigestLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestLanes() put %q first", xs[0].ID)
	}
	if len(IndexDigestLanes(xs)) != 3 {
		t.Errorf("IndexDigestLanes() = %v", IndexDigestLanes(xs))
	}
	if err := ValidateDigestLanes(ActiveDigestLanes(xs)); err != nil {
		t.Errorf("ValidateDigestLanes() = %v", err)
	}
}

func TestDigestLaneNormalizeClamps(t *testing.T) {
	x := SampleDigestLane("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestLane("r1", when)
	rev, err := r.PutDigestLane(x, 0)
	if err != nil {
		t.Fatalf("PutDigestLane() = %v", err)
	}
	got, gotRev, err := r.GetDigestLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestLanes() = %v, %v", all, err)
	}
	if _, err := r.PutDigestLane(DigestLane{}, 0); err == nil {
		t.Error("PutDigestLane() accepted an empty entity")
	}
}

func TestDigestSlotRoundTrip(t *testing.T) {
	x := SampleDigestSlot("digestslot-1", when)
	got, err := ParseDigestSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestSlot() = %v", err)
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

func TestDigestSlotParseRejectsBadInput(t *testing.T) {
	f := SampleDigestSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestSlot(f); err == nil {
		t.Error("ParseDigestSlot() accepted a record with no id")
	}
	f = SampleDigestSlot("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDigestSlot(f); err == nil {
		t.Error("ParseDigestSlot() accepted a non-numeric attempts")
	}
}

func TestDigestSlotCollections(t *testing.T) {
	xs := []DigestSlot{SampleDigestSlot("a", when), SampleDigestSlot("bb", when), SampleDigestSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestSlots(xs)); got != 2 {
		t.Errorf("ActiveDigestSlots() kept %d, want 2", got)
	}
	if got, want := SumDigestSlots(xs), 1+3; got != want {
		t.Errorf("SumDigestSlots() = %d, want %d", got, want)
	}
	SortDigestSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestSlots() put %q first", xs[0].ID)
	}
	if len(IndexDigestSlots(xs)) != 3 {
		t.Errorf("IndexDigestSlots() = %v", IndexDigestSlots(xs))
	}
	if err := ValidateDigestSlots(ActiveDigestSlots(xs)); err != nil {
		t.Errorf("ValidateDigestSlots() = %v", err)
	}
}

func TestDigestSlotNormalizeClamps(t *testing.T) {
	x := SampleDigestSlot("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestSlot("r1", when)
	rev, err := r.PutDigestSlot(x, 0)
	if err != nil {
		t.Fatalf("PutDigestSlot() = %v", err)
	}
	got, gotRev, err := r.GetDigestSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestSlots() = %v, %v", all, err)
	}
	if _, err := r.PutDigestSlot(DigestSlot{}, 0); err == nil {
		t.Error("PutDigestSlot() accepted an empty entity")
	}
}

func TestDigestTagRoundTrip(t *testing.T) {
	x := SampleDigestTag("digesttag-1", when)
	got, err := ParseDigestTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestTag() = %v", err)
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

func TestDigestTagParseRejectsBadInput(t *testing.T) {
	f := SampleDigestTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestTag(f); err == nil {
		t.Error("ParseDigestTag() accepted a record with no id")
	}
	f = SampleDigestTag("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDigestTag(f); err == nil {
		t.Error("ParseDigestTag() accepted a non-numeric offset")
	}
}

func TestDigestTagCollections(t *testing.T) {
	xs := []DigestTag{SampleDigestTag("a", when), SampleDigestTag("bb", when), SampleDigestTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestTags(xs)); got != 2 {
		t.Errorf("ActiveDigestTags() kept %d, want 2", got)
	}
	if got, want := SumDigestTags(xs), 1+3; got != want {
		t.Errorf("SumDigestTags() = %d, want %d", got, want)
	}
	SortDigestTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestTags() put %q first", xs[0].ID)
	}
	if len(IndexDigestTags(xs)) != 3 {
		t.Errorf("IndexDigestTags() = %v", IndexDigestTags(xs))
	}
	if err := ValidateDigestTags(ActiveDigestTags(xs)); err != nil {
		t.Errorf("ValidateDigestTags() = %v", err)
	}
}

func TestDigestTagNormalizeClamps(t *testing.T) {
	x := SampleDigestTag("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestTag("r1", when)
	rev, err := r.PutDigestTag(x, 0)
	if err != nil {
		t.Fatalf("PutDigestTag() = %v", err)
	}
	got, gotRev, err := r.GetDigestTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestTags() = %v, %v", all, err)
	}
	if _, err := r.PutDigestTag(DigestTag{}, 0); err == nil {
		t.Error("PutDigestTag() accepted an empty entity")
	}
}

func TestDigestRuleRoundTrip(t *testing.T) {
	x := SampleDigestRule("digestrule-1", when)
	got, err := ParseDigestRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestRule() = %v", err)
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

func TestDigestRuleParseRejectsBadInput(t *testing.T) {
	f := SampleDigestRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestRule(f); err == nil {
		t.Error("ParseDigestRule() accepted a record with no id")
	}
	f = SampleDigestRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDigestRule(f); err == nil {
		t.Error("ParseDigestRule() accepted a non-numeric offset")
	}
}

func TestDigestRuleCollections(t *testing.T) {
	xs := []DigestRule{SampleDigestRule("a", when), SampleDigestRule("bb", when), SampleDigestRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestRules(xs)); got != 2 {
		t.Errorf("ActiveDigestRules() kept %d, want 2", got)
	}
	if got, want := SumDigestRules(xs), 1+3; got != want {
		t.Errorf("SumDigestRules() = %d, want %d", got, want)
	}
	SortDigestRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestRules() put %q first", xs[0].ID)
	}
	if len(IndexDigestRules(xs)) != 3 {
		t.Errorf("IndexDigestRules() = %v", IndexDigestRules(xs))
	}
	if err := ValidateDigestRules(ActiveDigestRules(xs)); err != nil {
		t.Errorf("ValidateDigestRules() = %v", err)
	}
}

func TestDigestRuleNormalizeClamps(t *testing.T) {
	x := SampleDigestRule("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestRule("r1", when)
	rev, err := r.PutDigestRule(x, 0)
	if err != nil {
		t.Fatalf("PutDigestRule() = %v", err)
	}
	got, gotRev, err := r.GetDigestRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestRules() = %v, %v", all, err)
	}
	if _, err := r.PutDigestRule(DigestRule{}, 0); err == nil {
		t.Error("PutDigestRule() accepted an empty entity")
	}
}

func TestDigestHoldRoundTrip(t *testing.T) {
	x := SampleDigestHold("digesthold-1", when)
	got, err := ParseDigestHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestHold() = %v", err)
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

func TestDigestHoldParseRejectsBadInput(t *testing.T) {
	f := SampleDigestHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestHold(f); err == nil {
		t.Error("ParseDigestHold() accepted a record with no id")
	}
	f = SampleDigestHold("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDigestHold(f); err == nil {
		t.Error("ParseDigestHold() accepted a non-numeric attempts")
	}
}

func TestDigestHoldCollections(t *testing.T) {
	xs := []DigestHold{SampleDigestHold("a", when), SampleDigestHold("bb", when), SampleDigestHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestHolds(xs)); got != 2 {
		t.Errorf("ActiveDigestHolds() kept %d, want 2", got)
	}
	if got, want := SumDigestHolds(xs), 1+3; got != want {
		t.Errorf("SumDigestHolds() = %d, want %d", got, want)
	}
	SortDigestHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestHolds() put %q first", xs[0].ID)
	}
	if len(IndexDigestHolds(xs)) != 3 {
		t.Errorf("IndexDigestHolds() = %v", IndexDigestHolds(xs))
	}
	if err := ValidateDigestHolds(ActiveDigestHolds(xs)); err != nil {
		t.Errorf("ValidateDigestHolds() = %v", err)
	}
}

func TestDigestHoldNormalizeClamps(t *testing.T) {
	x := SampleDigestHold("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestHold("r1", when)
	rev, err := r.PutDigestHold(x, 0)
	if err != nil {
		t.Fatalf("PutDigestHold() = %v", err)
	}
	got, gotRev, err := r.GetDigestHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestHolds() = %v, %v", all, err)
	}
	if _, err := r.PutDigestHold(DigestHold{}, 0); err == nil {
		t.Error("PutDigestHold() accepted an empty entity")
	}
}

func TestDigestClaimRoundTrip(t *testing.T) {
	x := SampleDigestClaim("digestclaim-1", when)
	got, err := ParseDigestClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestClaim() = %v", err)
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

func TestDigestClaimParseRejectsBadInput(t *testing.T) {
	f := SampleDigestClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestClaim(f); err == nil {
		t.Error("ParseDigestClaim() accepted a record with no id")
	}
	f = SampleDigestClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDigestClaim(f); err == nil {
		t.Error("ParseDigestClaim() accepted a non-numeric retained")
	}
}

func TestDigestClaimCollections(t *testing.T) {
	xs := []DigestClaim{SampleDigestClaim("a", when), SampleDigestClaim("bb", when), SampleDigestClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestClaims(xs)); got != 2 {
		t.Errorf("ActiveDigestClaims() kept %d, want 2", got)
	}
	if got, want := SumDigestClaims(xs), 1+3; got != want {
		t.Errorf("SumDigestClaims() = %d, want %d", got, want)
	}
	SortDigestClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestClaims() put %q first", xs[0].ID)
	}
	if len(IndexDigestClaims(xs)) != 3 {
		t.Errorf("IndexDigestClaims() = %v", IndexDigestClaims(xs))
	}
	if err := ValidateDigestClaims(ActiveDigestClaims(xs)); err != nil {
		t.Errorf("ValidateDigestClaims() = %v", err)
	}
}

func TestDigestClaimNormalizeClamps(t *testing.T) {
	x := SampleDigestClaim("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestClaim("r1", when)
	rev, err := r.PutDigestClaim(x, 0)
	if err != nil {
		t.Fatalf("PutDigestClaim() = %v", err)
	}
	got, gotRev, err := r.GetDigestClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestClaims() = %v, %v", all, err)
	}
	if _, err := r.PutDigestClaim(DigestClaim{}, 0); err == nil {
		t.Error("PutDigestClaim() accepted an empty entity")
	}
}

func TestDigestCursorRoundTrip(t *testing.T) {
	x := SampleDigestCursor("digestcursor-1", when)
	got, err := ParseDigestCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseDigestCursor() = %v", err)
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

func TestDigestCursorParseRejectsBadInput(t *testing.T) {
	f := SampleDigestCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDigestCursor(f); err == nil {
		t.Error("ParseDigestCursor() accepted a record with no id")
	}
	f = SampleDigestCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDigestCursor(f); err == nil {
		t.Error("ParseDigestCursor() accepted a non-numeric version")
	}
}

func TestDigestCursorCollections(t *testing.T) {
	xs := []DigestCursor{SampleDigestCursor("a", when), SampleDigestCursor("bb", when), SampleDigestCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDigestCursors(xs)); got != 2 {
		t.Errorf("ActiveDigestCursors() kept %d, want 2", got)
	}
	if got, want := SumDigestCursors(xs), 1+3; got != want {
		t.Errorf("SumDigestCursors() = %d, want %d", got, want)
	}
	SortDigestCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDigestCursors() put %q first", xs[0].ID)
	}
	if len(IndexDigestCursors(xs)) != 3 {
		t.Errorf("IndexDigestCursors() = %v", IndexDigestCursors(xs))
	}
	if err := ValidateDigestCursors(ActiveDigestCursors(xs)); err != nil {
		t.Errorf("ValidateDigestCursors() = %v", err)
	}
}

func TestDigestCursorNormalizeClamps(t *testing.T) {
	x := SampleDigestCursor("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDigestCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDigestCursor("r1", when)
	rev, err := r.PutDigestCursor(x, 0)
	if err != nil {
		t.Fatalf("PutDigestCursor() = %v", err)
	}
	got, gotRev, err := r.GetDigestCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDigestCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDigestCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDigestCursors() = %v, %v", all, err)
	}
	if _, err := r.PutDigestCursor(DigestCursor{}, 0); err == nil {
		t.Error("PutDigestCursor() accepted an empty entity")
	}
}
