package usage

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
		if !strings.HasPrefix(Prefixes[i], "usage/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleUsageRecord("ok", when).Fields()); err != nil {
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

func TestUsageRecordRoundTrip(t *testing.T) {
	x := SampleUsageRecord("usagerecord-1", when)
	got, err := ParseUsageRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageRecord() = %v", err)
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

func TestUsageRecordParseRejectsBadInput(t *testing.T) {
	f := SampleUsageRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageRecord(f); err == nil {
		t.Error("ParseUsageRecord() accepted a record with no id")
	}
	f = SampleUsageRecord("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseUsageRecord(f); err == nil {
		t.Error("ParseUsageRecord() accepted a non-numeric count")
	}
}

func TestUsageRecordCollections(t *testing.T) {
	xs := []UsageRecord{SampleUsageRecord("a", when), SampleUsageRecord("bb", when), SampleUsageRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageRecords(xs)); got != 2 {
		t.Errorf("ActiveUsageRecords() kept %d, want 2", got)
	}
	if got, want := SumUsageRecords(xs), 1+3; got != want {
		t.Errorf("SumUsageRecords() = %d, want %d", got, want)
	}
	SortUsageRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageRecords() put %q first", xs[0].ID)
	}
	if len(IndexUsageRecords(xs)) != 3 {
		t.Errorf("IndexUsageRecords() = %v", IndexUsageRecords(xs))
	}
	if err := ValidateUsageRecords(ActiveUsageRecords(xs)); err != nil {
		t.Errorf("ValidateUsageRecords() = %v", err)
	}
}

func TestUsageRecordNormalizeClamps(t *testing.T) {
	x := SampleUsageRecord("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageRecord("r1", when)
	rev, err := r.PutUsageRecord(x, 0)
	if err != nil {
		t.Fatalf("PutUsageRecord() = %v", err)
	}
	got, gotRev, err := r.GetUsageRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageRecords() = %v, %v", all, err)
	}
	if _, err := r.PutUsageRecord(UsageRecord{}, 0); err == nil {
		t.Error("PutUsageRecord() accepted an empty entity")
	}
}

func TestUsageEntryRoundTrip(t *testing.T) {
	x := SampleUsageEntry("usageentry-1", when)
	got, err := ParseUsageEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageEntry() = %v", err)
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

func TestUsageEntryParseRejectsBadInput(t *testing.T) {
	f := SampleUsageEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageEntry(f); err == nil {
		t.Error("ParseUsageEntry() accepted a record with no id")
	}
	f = SampleUsageEntry("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseUsageEntry(f); err == nil {
		t.Error("ParseUsageEntry() accepted a non-numeric capacity")
	}
}

func TestUsageEntryCollections(t *testing.T) {
	xs := []UsageEntry{SampleUsageEntry("a", when), SampleUsageEntry("bb", when), SampleUsageEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageEntrys(xs)); got != 2 {
		t.Errorf("ActiveUsageEntrys() kept %d, want 2", got)
	}
	if got, want := SumUsageEntrys(xs), 1+3; got != want {
		t.Errorf("SumUsageEntrys() = %d, want %d", got, want)
	}
	SortUsageEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageEntrys() put %q first", xs[0].ID)
	}
	if len(IndexUsageEntrys(xs)) != 3 {
		t.Errorf("IndexUsageEntrys() = %v", IndexUsageEntrys(xs))
	}
	if err := ValidateUsageEntrys(ActiveUsageEntrys(xs)); err != nil {
		t.Errorf("ValidateUsageEntrys() = %v", err)
	}
}

func TestUsageEntryNormalizeClamps(t *testing.T) {
	x := SampleUsageEntry("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageEntry("r1", when)
	rev, err := r.PutUsageEntry(x, 0)
	if err != nil {
		t.Fatalf("PutUsageEntry() = %v", err)
	}
	got, gotRev, err := r.GetUsageEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutUsageEntry(UsageEntry{}, 0); err == nil {
		t.Error("PutUsageEntry() accepted an empty entity")
	}
}

func TestUsageItemRoundTrip(t *testing.T) {
	x := SampleUsageItem("usageitem-1", when)
	got, err := ParseUsageItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageItem() = %v", err)
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

func TestUsageItemParseRejectsBadInput(t *testing.T) {
	f := SampleUsageItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageItem(f); err == nil {
		t.Error("ParseUsageItem() accepted a record with no id")
	}
	f = SampleUsageItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseUsageItem(f); err == nil {
		t.Error("ParseUsageItem() accepted a non-numeric version")
	}
}

func TestUsageItemCollections(t *testing.T) {
	xs := []UsageItem{SampleUsageItem("a", when), SampleUsageItem("bb", when), SampleUsageItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageItems(xs)); got != 2 {
		t.Errorf("ActiveUsageItems() kept %d, want 2", got)
	}
	if got, want := SumUsageItems(xs), 1+3; got != want {
		t.Errorf("SumUsageItems() = %d, want %d", got, want)
	}
	SortUsageItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageItems() put %q first", xs[0].ID)
	}
	if len(IndexUsageItems(xs)) != 3 {
		t.Errorf("IndexUsageItems() = %v", IndexUsageItems(xs))
	}
	if err := ValidateUsageItems(ActiveUsageItems(xs)); err != nil {
		t.Errorf("ValidateUsageItems() = %v", err)
	}
}

func TestUsageItemNormalizeClamps(t *testing.T) {
	x := SampleUsageItem("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageItem("r1", when)
	rev, err := r.PutUsageItem(x, 0)
	if err != nil {
		t.Fatalf("PutUsageItem() = %v", err)
	}
	got, gotRev, err := r.GetUsageItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageItems() = %v, %v", all, err)
	}
	if _, err := r.PutUsageItem(UsageItem{}, 0); err == nil {
		t.Error("PutUsageItem() accepted an empty entity")
	}
}

func TestUsageDraftRoundTrip(t *testing.T) {
	x := SampleUsageDraft("usagedraft-1", when)
	got, err := ParseUsageDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageDraft() = %v", err)
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

func TestUsageDraftParseRejectsBadInput(t *testing.T) {
	f := SampleUsageDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageDraft(f); err == nil {
		t.Error("ParseUsageDraft() accepted a record with no id")
	}
	f = SampleUsageDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseUsageDraft(f); err == nil {
		t.Error("ParseUsageDraft() accepted a non-numeric attempts")
	}
}

func TestUsageDraftCollections(t *testing.T) {
	xs := []UsageDraft{SampleUsageDraft("a", when), SampleUsageDraft("bb", when), SampleUsageDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageDrafts(xs)); got != 2 {
		t.Errorf("ActiveUsageDrafts() kept %d, want 2", got)
	}
	if got, want := SumUsageDrafts(xs), 1+3; got != want {
		t.Errorf("SumUsageDrafts() = %d, want %d", got, want)
	}
	SortUsageDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageDrafts() put %q first", xs[0].ID)
	}
	if len(IndexUsageDrafts(xs)) != 3 {
		t.Errorf("IndexUsageDrafts() = %v", IndexUsageDrafts(xs))
	}
	if err := ValidateUsageDrafts(ActiveUsageDrafts(xs)); err != nil {
		t.Errorf("ValidateUsageDrafts() = %v", err)
	}
}

func TestUsageDraftNormalizeClamps(t *testing.T) {
	x := SampleUsageDraft("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageDraft("r1", when)
	rev, err := r.PutUsageDraft(x, 0)
	if err != nil {
		t.Fatalf("PutUsageDraft() = %v", err)
	}
	got, gotRev, err := r.GetUsageDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutUsageDraft(UsageDraft{}, 0); err == nil {
		t.Error("PutUsageDraft() accepted an empty entity")
	}
}

func TestUsageRevisionRoundTrip(t *testing.T) {
	x := SampleUsageRevision("usagerevision-1", when)
	got, err := ParseUsageRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageRevision() = %v", err)
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

func TestUsageRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleUsageRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageRevision(f); err == nil {
		t.Error("ParseUsageRevision() accepted a record with no id")
	}
	f = SampleUsageRevision("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUsageRevision(f); err == nil {
		t.Error("ParseUsageRevision() accepted a non-numeric offset")
	}
}

func TestUsageRevisionCollections(t *testing.T) {
	xs := []UsageRevision{SampleUsageRevision("a", when), SampleUsageRevision("bb", when), SampleUsageRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageRevisions(xs)); got != 2 {
		t.Errorf("ActiveUsageRevisions() kept %d, want 2", got)
	}
	if got, want := SumUsageRevisions(xs), 1+3; got != want {
		t.Errorf("SumUsageRevisions() = %d, want %d", got, want)
	}
	SortUsageRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageRevisions() put %q first", xs[0].ID)
	}
	if len(IndexUsageRevisions(xs)) != 3 {
		t.Errorf("IndexUsageRevisions() = %v", IndexUsageRevisions(xs))
	}
	if err := ValidateUsageRevisions(ActiveUsageRevisions(xs)); err != nil {
		t.Errorf("ValidateUsageRevisions() = %v", err)
	}
}

func TestUsageRevisionNormalizeClamps(t *testing.T) {
	x := SampleUsageRevision("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageRevision("r1", when)
	rev, err := r.PutUsageRevision(x, 0)
	if err != nil {
		t.Fatalf("PutUsageRevision() = %v", err)
	}
	got, gotRev, err := r.GetUsageRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutUsageRevision(UsageRevision{}, 0); err == nil {
		t.Error("PutUsageRevision() accepted an empty entity")
	}
}

func TestUsageSnapshotRoundTrip(t *testing.T) {
	x := SampleUsageSnapshot("usagesnapshot-1", when)
	got, err := ParseUsageSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageSnapshot() = %v", err)
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

func TestUsageSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleUsageSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageSnapshot(f); err == nil {
		t.Error("ParseUsageSnapshot() accepted a record with no id")
	}
	f = SampleUsageSnapshot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseUsageSnapshot(f); err == nil {
		t.Error("ParseUsageSnapshot() accepted a non-numeric retained")
	}
}

func TestUsageSnapshotCollections(t *testing.T) {
	xs := []UsageSnapshot{SampleUsageSnapshot("a", when), SampleUsageSnapshot("bb", when), SampleUsageSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageSnapshots(xs)); got != 2 {
		t.Errorf("ActiveUsageSnapshots() kept %d, want 2", got)
	}
	if got, want := SumUsageSnapshots(xs), 1+3; got != want {
		t.Errorf("SumUsageSnapshots() = %d, want %d", got, want)
	}
	SortUsageSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexUsageSnapshots(xs)) != 3 {
		t.Errorf("IndexUsageSnapshots() = %v", IndexUsageSnapshots(xs))
	}
	if err := ValidateUsageSnapshots(ActiveUsageSnapshots(xs)); err != nil {
		t.Errorf("ValidateUsageSnapshots() = %v", err)
	}
}

func TestUsageSnapshotNormalizeClamps(t *testing.T) {
	x := SampleUsageSnapshot("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageSnapshot("r1", when)
	rev, err := r.PutUsageSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutUsageSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetUsageSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutUsageSnapshot(UsageSnapshot{}, 0); err == nil {
		t.Error("PutUsageSnapshot() accepted an empty entity")
	}
}

func TestUsageBatchRoundTrip(t *testing.T) {
	x := SampleUsageBatch("usagebatch-1", when)
	got, err := ParseUsageBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageBatch() = %v", err)
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

func TestUsageBatchParseRejectsBadInput(t *testing.T) {
	f := SampleUsageBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageBatch(f); err == nil {
		t.Error("ParseUsageBatch() accepted a record with no id")
	}
	f = SampleUsageBatch("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseUsageBatch(f); err == nil {
		t.Error("ParseUsageBatch() accepted a non-numeric retained")
	}
}

func TestUsageBatchCollections(t *testing.T) {
	xs := []UsageBatch{SampleUsageBatch("a", when), SampleUsageBatch("bb", when), SampleUsageBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageBatchs(xs)); got != 2 {
		t.Errorf("ActiveUsageBatchs() kept %d, want 2", got)
	}
	if got, want := SumUsageBatchs(xs), 1+3; got != want {
		t.Errorf("SumUsageBatchs() = %d, want %d", got, want)
	}
	SortUsageBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageBatchs() put %q first", xs[0].ID)
	}
	if len(IndexUsageBatchs(xs)) != 3 {
		t.Errorf("IndexUsageBatchs() = %v", IndexUsageBatchs(xs))
	}
	if err := ValidateUsageBatchs(ActiveUsageBatchs(xs)); err != nil {
		t.Errorf("ValidateUsageBatchs() = %v", err)
	}
}

func TestUsageBatchNormalizeClamps(t *testing.T) {
	x := SampleUsageBatch("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageBatch("r1", when)
	rev, err := r.PutUsageBatch(x, 0)
	if err != nil {
		t.Fatalf("PutUsageBatch() = %v", err)
	}
	got, gotRev, err := r.GetUsageBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutUsageBatch(UsageBatch{}, 0); err == nil {
		t.Error("PutUsageBatch() accepted an empty entity")
	}
}

func TestUsageRequestRoundTrip(t *testing.T) {
	x := SampleUsageRequest("usagerequest-1", when)
	got, err := ParseUsageRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageRequest() = %v", err)
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

func TestUsageRequestParseRejectsBadInput(t *testing.T) {
	f := SampleUsageRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageRequest(f); err == nil {
		t.Error("ParseUsageRequest() accepted a record with no id")
	}
	f = SampleUsageRequest("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseUsageRequest(f); err == nil {
		t.Error("ParseUsageRequest() accepted a non-numeric capacity")
	}
}

func TestUsageRequestCollections(t *testing.T) {
	xs := []UsageRequest{SampleUsageRequest("a", when), SampleUsageRequest("bb", when), SampleUsageRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageRequests(xs)); got != 2 {
		t.Errorf("ActiveUsageRequests() kept %d, want 2", got)
	}
	if got, want := SumUsageRequests(xs), 1+3; got != want {
		t.Errorf("SumUsageRequests() = %d, want %d", got, want)
	}
	SortUsageRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageRequests() put %q first", xs[0].ID)
	}
	if len(IndexUsageRequests(xs)) != 3 {
		t.Errorf("IndexUsageRequests() = %v", IndexUsageRequests(xs))
	}
	if err := ValidateUsageRequests(ActiveUsageRequests(xs)); err != nil {
		t.Errorf("ValidateUsageRequests() = %v", err)
	}
}

func TestUsageRequestNormalizeClamps(t *testing.T) {
	x := SampleUsageRequest("a", when)
	x.Label = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageRequest("r1", when)
	rev, err := r.PutUsageRequest(x, 0)
	if err != nil {
		t.Fatalf("PutUsageRequest() = %v", err)
	}
	got, gotRev, err := r.GetUsageRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageRequests() = %v, %v", all, err)
	}
	if _, err := r.PutUsageRequest(UsageRequest{}, 0); err == nil {
		t.Error("PutUsageRequest() accepted an empty entity")
	}
}

func TestUsageSummaryRoundTrip(t *testing.T) {
	x := SampleUsageSummary("usagesummary-1", when)
	got, err := ParseUsageSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageSummary() = %v", err)
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

func TestUsageSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleUsageSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageSummary(f); err == nil {
		t.Error("ParseUsageSummary() accepted a record with no id")
	}
	f = SampleUsageSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseUsageSummary(f); err == nil {
		t.Error("ParseUsageSummary() accepted a non-numeric capacity")
	}
}

func TestUsageSummaryCollections(t *testing.T) {
	xs := []UsageSummary{SampleUsageSummary("a", when), SampleUsageSummary("bb", when), SampleUsageSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageSummarys(xs)); got != 2 {
		t.Errorf("ActiveUsageSummarys() kept %d, want 2", got)
	}
	if got, want := SumUsageSummarys(xs), 1+3; got != want {
		t.Errorf("SumUsageSummarys() = %d, want %d", got, want)
	}
	SortUsageSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageSummarys() put %q first", xs[0].ID)
	}
	if len(IndexUsageSummarys(xs)) != 3 {
		t.Errorf("IndexUsageSummarys() = %v", IndexUsageSummarys(xs))
	}
	if err := ValidateUsageSummarys(ActiveUsageSummarys(xs)); err != nil {
		t.Errorf("ValidateUsageSummarys() = %v", err)
	}
}

func TestUsageSummaryNormalizeClamps(t *testing.T) {
	x := SampleUsageSummary("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageSummary("r1", when)
	rev, err := r.PutUsageSummary(x, 0)
	if err != nil {
		t.Fatalf("PutUsageSummary() = %v", err)
	}
	got, gotRev, err := r.GetUsageSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutUsageSummary(UsageSummary{}, 0); err == nil {
		t.Error("PutUsageSummary() accepted an empty entity")
	}
}

func TestUsagePolicyRoundTrip(t *testing.T) {
	x := SampleUsagePolicy("usagepolicy-1", when)
	got, err := ParseUsagePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsagePolicy() = %v", err)
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

func TestUsagePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleUsagePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsagePolicy(f); err == nil {
		t.Error("ParseUsagePolicy() accepted a record with no id")
	}
	f = SampleUsagePolicy("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseUsagePolicy(f); err == nil {
		t.Error("ParseUsagePolicy() accepted a non-numeric version")
	}
}

func TestUsagePolicyCollections(t *testing.T) {
	xs := []UsagePolicy{SampleUsagePolicy("a", when), SampleUsagePolicy("bb", when), SampleUsagePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsagePolicys(xs)); got != 2 {
		t.Errorf("ActiveUsagePolicys() kept %d, want 2", got)
	}
	if got, want := SumUsagePolicys(xs), 1+3; got != want {
		t.Errorf("SumUsagePolicys() = %d, want %d", got, want)
	}
	SortUsagePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsagePolicys() put %q first", xs[0].ID)
	}
	if len(IndexUsagePolicys(xs)) != 3 {
		t.Errorf("IndexUsagePolicys() = %v", IndexUsagePolicys(xs))
	}
	if err := ValidateUsagePolicys(ActiveUsagePolicys(xs)); err != nil {
		t.Errorf("ValidateUsagePolicys() = %v", err)
	}
}

func TestUsagePolicyNormalizeClamps(t *testing.T) {
	x := SampleUsagePolicy("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsagePolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsagePolicy("r1", when)
	rev, err := r.PutUsagePolicy(x, 0)
	if err != nil {
		t.Fatalf("PutUsagePolicy() = %v", err)
	}
	got, gotRev, err := r.GetUsagePolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsagePolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsagePolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsagePolicys() = %v, %v", all, err)
	}
	if _, err := r.PutUsagePolicy(UsagePolicy{}, 0); err == nil {
		t.Error("PutUsagePolicy() accepted an empty entity")
	}
}

func TestUsageLinkRoundTrip(t *testing.T) {
	x := SampleUsageLink("usagelink-1", when)
	got, err := ParseUsageLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageLink() = %v", err)
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

func TestUsageLinkParseRejectsBadInput(t *testing.T) {
	f := SampleUsageLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageLink(f); err == nil {
		t.Error("ParseUsageLink() accepted a record with no id")
	}
	f = SampleUsageLink("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseUsageLink(f); err == nil {
		t.Error("ParseUsageLink() accepted a non-numeric count")
	}
}

func TestUsageLinkCollections(t *testing.T) {
	xs := []UsageLink{SampleUsageLink("a", when), SampleUsageLink("bb", when), SampleUsageLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageLinks(xs)); got != 2 {
		t.Errorf("ActiveUsageLinks() kept %d, want 2", got)
	}
	if got, want := SumUsageLinks(xs), 1+3; got != want {
		t.Errorf("SumUsageLinks() = %d, want %d", got, want)
	}
	SortUsageLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageLinks() put %q first", xs[0].ID)
	}
	if len(IndexUsageLinks(xs)) != 3 {
		t.Errorf("IndexUsageLinks() = %v", IndexUsageLinks(xs))
	}
	if err := ValidateUsageLinks(ActiveUsageLinks(xs)); err != nil {
		t.Errorf("ValidateUsageLinks() = %v", err)
	}
}

func TestUsageLinkNormalizeClamps(t *testing.T) {
	x := SampleUsageLink("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageLink("r1", when)
	rev, err := r.PutUsageLink(x, 0)
	if err != nil {
		t.Fatalf("PutUsageLink() = %v", err)
	}
	got, gotRev, err := r.GetUsageLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageLinks() = %v, %v", all, err)
	}
	if _, err := r.PutUsageLink(UsageLink{}, 0); err == nil {
		t.Error("PutUsageLink() accepted an empty entity")
	}
}

func TestUsageNoteRoundTrip(t *testing.T) {
	x := SampleUsageNote("usagenote-1", when)
	got, err := ParseUsageNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageNote() = %v", err)
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

func TestUsageNoteParseRejectsBadInput(t *testing.T) {
	f := SampleUsageNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageNote(f); err == nil {
		t.Error("ParseUsageNote() accepted a record with no id")
	}
	f = SampleUsageNote("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUsageNote(f); err == nil {
		t.Error("ParseUsageNote() accepted a non-numeric offset")
	}
}

func TestUsageNoteCollections(t *testing.T) {
	xs := []UsageNote{SampleUsageNote("a", when), SampleUsageNote("bb", when), SampleUsageNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageNotes(xs)); got != 2 {
		t.Errorf("ActiveUsageNotes() kept %d, want 2", got)
	}
	if got, want := SumUsageNotes(xs), 1+3; got != want {
		t.Errorf("SumUsageNotes() = %d, want %d", got, want)
	}
	SortUsageNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageNotes() put %q first", xs[0].ID)
	}
	if len(IndexUsageNotes(xs)) != 3 {
		t.Errorf("IndexUsageNotes() = %v", IndexUsageNotes(xs))
	}
	if err := ValidateUsageNotes(ActiveUsageNotes(xs)); err != nil {
		t.Errorf("ValidateUsageNotes() = %v", err)
	}
}

func TestUsageNoteNormalizeClamps(t *testing.T) {
	x := SampleUsageNote("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageNote("r1", when)
	rev, err := r.PutUsageNote(x, 0)
	if err != nil {
		t.Fatalf("PutUsageNote() = %v", err)
	}
	got, gotRev, err := r.GetUsageNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageNotes() = %v, %v", all, err)
	}
	if _, err := r.PutUsageNote(UsageNote{}, 0); err == nil {
		t.Error("PutUsageNote() accepted an empty entity")
	}
}

func TestUsageWindowRoundTrip(t *testing.T) {
	x := SampleUsageWindow("usagewindow-1", when)
	got, err := ParseUsageWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageWindow() = %v", err)
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

func TestUsageWindowParseRejectsBadInput(t *testing.T) {
	f := SampleUsageWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageWindow(f); err == nil {
		t.Error("ParseUsageWindow() accepted a record with no id")
	}
	f = SampleUsageWindow("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseUsageWindow(f); err == nil {
		t.Error("ParseUsageWindow() accepted a non-numeric count")
	}
}

func TestUsageWindowCollections(t *testing.T) {
	xs := []UsageWindow{SampleUsageWindow("a", when), SampleUsageWindow("bb", when), SampleUsageWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageWindows(xs)); got != 2 {
		t.Errorf("ActiveUsageWindows() kept %d, want 2", got)
	}
	if got, want := SumUsageWindows(xs), 1+3; got != want {
		t.Errorf("SumUsageWindows() = %d, want %d", got, want)
	}
	SortUsageWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageWindows() put %q first", xs[0].ID)
	}
	if len(IndexUsageWindows(xs)) != 3 {
		t.Errorf("IndexUsageWindows() = %v", IndexUsageWindows(xs))
	}
	if err := ValidateUsageWindows(ActiveUsageWindows(xs)); err != nil {
		t.Errorf("ValidateUsageWindows() = %v", err)
	}
}

func TestUsageWindowNormalizeClamps(t *testing.T) {
	x := SampleUsageWindow("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageWindow("r1", when)
	rev, err := r.PutUsageWindow(x, 0)
	if err != nil {
		t.Fatalf("PutUsageWindow() = %v", err)
	}
	got, gotRev, err := r.GetUsageWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageWindows() = %v, %v", all, err)
	}
	if _, err := r.PutUsageWindow(UsageWindow{}, 0); err == nil {
		t.Error("PutUsageWindow() accepted an empty entity")
	}
}

func TestUsageQuotaRoundTrip(t *testing.T) {
	x := SampleUsageQuota("usagequota-1", when)
	got, err := ParseUsageQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageQuota() = %v", err)
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

func TestUsageQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleUsageQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageQuota(f); err == nil {
		t.Error("ParseUsageQuota() accepted a record with no id")
	}
	f = SampleUsageQuota("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseUsageQuota(f); err == nil {
		t.Error("ParseUsageQuota() accepted a non-numeric version")
	}
}

func TestUsageQuotaCollections(t *testing.T) {
	xs := []UsageQuota{SampleUsageQuota("a", when), SampleUsageQuota("bb", when), SampleUsageQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageQuotas(xs)); got != 2 {
		t.Errorf("ActiveUsageQuotas() kept %d, want 2", got)
	}
	if got, want := SumUsageQuotas(xs), 1+3; got != want {
		t.Errorf("SumUsageQuotas() = %d, want %d", got, want)
	}
	SortUsageQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageQuotas() put %q first", xs[0].ID)
	}
	if len(IndexUsageQuotas(xs)) != 3 {
		t.Errorf("IndexUsageQuotas() = %v", IndexUsageQuotas(xs))
	}
	if err := ValidateUsageQuotas(ActiveUsageQuotas(xs)); err != nil {
		t.Errorf("ValidateUsageQuotas() = %v", err)
	}
}

func TestUsageQuotaNormalizeClamps(t *testing.T) {
	x := SampleUsageQuota("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageQuota("r1", when)
	rev, err := r.PutUsageQuota(x, 0)
	if err != nil {
		t.Fatalf("PutUsageQuota() = %v", err)
	}
	got, gotRev, err := r.GetUsageQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutUsageQuota(UsageQuota{}, 0); err == nil {
		t.Error("PutUsageQuota() accepted an empty entity")
	}
}

func TestUsageBucketRoundTrip(t *testing.T) {
	x := SampleUsageBucket("usagebucket-1", when)
	got, err := ParseUsageBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageBucket() = %v", err)
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

func TestUsageBucketParseRejectsBadInput(t *testing.T) {
	f := SampleUsageBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageBucket(f); err == nil {
		t.Error("ParseUsageBucket() accepted a record with no id")
	}
	f = SampleUsageBucket("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseUsageBucket(f); err == nil {
		t.Error("ParseUsageBucket() accepted a non-numeric attempts")
	}
}

func TestUsageBucketCollections(t *testing.T) {
	xs := []UsageBucket{SampleUsageBucket("a", when), SampleUsageBucket("bb", when), SampleUsageBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageBuckets(xs)); got != 2 {
		t.Errorf("ActiveUsageBuckets() kept %d, want 2", got)
	}
	if got, want := SumUsageBuckets(xs), 1+3; got != want {
		t.Errorf("SumUsageBuckets() = %d, want %d", got, want)
	}
	SortUsageBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageBuckets() put %q first", xs[0].ID)
	}
	if len(IndexUsageBuckets(xs)) != 3 {
		t.Errorf("IndexUsageBuckets() = %v", IndexUsageBuckets(xs))
	}
	if err := ValidateUsageBuckets(ActiveUsageBuckets(xs)); err != nil {
		t.Errorf("ValidateUsageBuckets() = %v", err)
	}
}

func TestUsageBucketNormalizeClamps(t *testing.T) {
	x := SampleUsageBucket("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageBucket("r1", when)
	rev, err := r.PutUsageBucket(x, 0)
	if err != nil {
		t.Fatalf("PutUsageBucket() = %v", err)
	}
	got, gotRev, err := r.GetUsageBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutUsageBucket(UsageBucket{}, 0); err == nil {
		t.Error("PutUsageBucket() accepted an empty entity")
	}
}

func TestUsageLaneRoundTrip(t *testing.T) {
	x := SampleUsageLane("usagelane-1", when)
	got, err := ParseUsageLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageLane() = %v", err)
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

func TestUsageLaneParseRejectsBadInput(t *testing.T) {
	f := SampleUsageLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageLane(f); err == nil {
		t.Error("ParseUsageLane() accepted a record with no id")
	}
	f = SampleUsageLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseUsageLane(f); err == nil {
		t.Error("ParseUsageLane() accepted a non-numeric capacity")
	}
}

func TestUsageLaneCollections(t *testing.T) {
	xs := []UsageLane{SampleUsageLane("a", when), SampleUsageLane("bb", when), SampleUsageLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageLanes(xs)); got != 2 {
		t.Errorf("ActiveUsageLanes() kept %d, want 2", got)
	}
	if got, want := SumUsageLanes(xs), 1+3; got != want {
		t.Errorf("SumUsageLanes() = %d, want %d", got, want)
	}
	SortUsageLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageLanes() put %q first", xs[0].ID)
	}
	if len(IndexUsageLanes(xs)) != 3 {
		t.Errorf("IndexUsageLanes() = %v", IndexUsageLanes(xs))
	}
	if err := ValidateUsageLanes(ActiveUsageLanes(xs)); err != nil {
		t.Errorf("ValidateUsageLanes() = %v", err)
	}
}

func TestUsageLaneNormalizeClamps(t *testing.T) {
	x := SampleUsageLane("a", when)
	x.Label = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageLane("r1", when)
	rev, err := r.PutUsageLane(x, 0)
	if err != nil {
		t.Fatalf("PutUsageLane() = %v", err)
	}
	got, gotRev, err := r.GetUsageLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageLanes() = %v, %v", all, err)
	}
	if _, err := r.PutUsageLane(UsageLane{}, 0); err == nil {
		t.Error("PutUsageLane() accepted an empty entity")
	}
}

func TestUsageSlotRoundTrip(t *testing.T) {
	x := SampleUsageSlot("usageslot-1", when)
	got, err := ParseUsageSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageSlot() = %v", err)
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

func TestUsageSlotParseRejectsBadInput(t *testing.T) {
	f := SampleUsageSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageSlot(f); err == nil {
		t.Error("ParseUsageSlot() accepted a record with no id")
	}
	f = SampleUsageSlot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseUsageSlot(f); err == nil {
		t.Error("ParseUsageSlot() accepted a non-numeric weight")
	}
}

func TestUsageSlotCollections(t *testing.T) {
	xs := []UsageSlot{SampleUsageSlot("a", when), SampleUsageSlot("bb", when), SampleUsageSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageSlots(xs)); got != 2 {
		t.Errorf("ActiveUsageSlots() kept %d, want 2", got)
	}
	if got, want := SumUsageSlots(xs), 1+3; got != want {
		t.Errorf("SumUsageSlots() = %d, want %d", got, want)
	}
	SortUsageSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageSlots() put %q first", xs[0].ID)
	}
	if len(IndexUsageSlots(xs)) != 3 {
		t.Errorf("IndexUsageSlots() = %v", IndexUsageSlots(xs))
	}
	if err := ValidateUsageSlots(ActiveUsageSlots(xs)); err != nil {
		t.Errorf("ValidateUsageSlots() = %v", err)
	}
}

func TestUsageSlotNormalizeClamps(t *testing.T) {
	x := SampleUsageSlot("a", when)
	x.Owner = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageSlot("r1", when)
	rev, err := r.PutUsageSlot(x, 0)
	if err != nil {
		t.Fatalf("PutUsageSlot() = %v", err)
	}
	got, gotRev, err := r.GetUsageSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageSlots() = %v, %v", all, err)
	}
	if _, err := r.PutUsageSlot(UsageSlot{}, 0); err == nil {
		t.Error("PutUsageSlot() accepted an empty entity")
	}
}

func TestUsageTagRoundTrip(t *testing.T) {
	x := SampleUsageTag("usagetag-1", when)
	got, err := ParseUsageTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageTag() = %v", err)
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

func TestUsageTagParseRejectsBadInput(t *testing.T) {
	f := SampleUsageTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageTag(f); err == nil {
		t.Error("ParseUsageTag() accepted a record with no id")
	}
	f = SampleUsageTag("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseUsageTag(f); err == nil {
		t.Error("ParseUsageTag() accepted a non-numeric capacity")
	}
}

func TestUsageTagCollections(t *testing.T) {
	xs := []UsageTag{SampleUsageTag("a", when), SampleUsageTag("bb", when), SampleUsageTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageTags(xs)); got != 2 {
		t.Errorf("ActiveUsageTags() kept %d, want 2", got)
	}
	if got, want := SumUsageTags(xs), 1+3; got != want {
		t.Errorf("SumUsageTags() = %d, want %d", got, want)
	}
	SortUsageTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageTags() put %q first", xs[0].ID)
	}
	if len(IndexUsageTags(xs)) != 3 {
		t.Errorf("IndexUsageTags() = %v", IndexUsageTags(xs))
	}
	if err := ValidateUsageTags(ActiveUsageTags(xs)); err != nil {
		t.Errorf("ValidateUsageTags() = %v", err)
	}
}

func TestUsageTagNormalizeClamps(t *testing.T) {
	x := SampleUsageTag("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageTag("r1", when)
	rev, err := r.PutUsageTag(x, 0)
	if err != nil {
		t.Fatalf("PutUsageTag() = %v", err)
	}
	got, gotRev, err := r.GetUsageTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageTags() = %v, %v", all, err)
	}
	if _, err := r.PutUsageTag(UsageTag{}, 0); err == nil {
		t.Error("PutUsageTag() accepted an empty entity")
	}
}

func TestUsageRuleRoundTrip(t *testing.T) {
	x := SampleUsageRule("usagerule-1", when)
	got, err := ParseUsageRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageRule() = %v", err)
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

func TestUsageRuleParseRejectsBadInput(t *testing.T) {
	f := SampleUsageRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageRule(f); err == nil {
		t.Error("ParseUsageRule() accepted a record with no id")
	}
	f = SampleUsageRule("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseUsageRule(f); err == nil {
		t.Error("ParseUsageRule() accepted a non-numeric attempts")
	}
}

func TestUsageRuleCollections(t *testing.T) {
	xs := []UsageRule{SampleUsageRule("a", when), SampleUsageRule("bb", when), SampleUsageRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageRules(xs)); got != 2 {
		t.Errorf("ActiveUsageRules() kept %d, want 2", got)
	}
	if got, want := SumUsageRules(xs), 1+3; got != want {
		t.Errorf("SumUsageRules() = %d, want %d", got, want)
	}
	SortUsageRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageRules() put %q first", xs[0].ID)
	}
	if len(IndexUsageRules(xs)) != 3 {
		t.Errorf("IndexUsageRules() = %v", IndexUsageRules(xs))
	}
	if err := ValidateUsageRules(ActiveUsageRules(xs)); err != nil {
		t.Errorf("ValidateUsageRules() = %v", err)
	}
}

func TestUsageRuleNormalizeClamps(t *testing.T) {
	x := SampleUsageRule("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageRule("r1", when)
	rev, err := r.PutUsageRule(x, 0)
	if err != nil {
		t.Fatalf("PutUsageRule() = %v", err)
	}
	got, gotRev, err := r.GetUsageRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageRules() = %v, %v", all, err)
	}
	if _, err := r.PutUsageRule(UsageRule{}, 0); err == nil {
		t.Error("PutUsageRule() accepted an empty entity")
	}
}

func TestUsageHoldRoundTrip(t *testing.T) {
	x := SampleUsageHold("usagehold-1", when)
	got, err := ParseUsageHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageHold() = %v", err)
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

func TestUsageHoldParseRejectsBadInput(t *testing.T) {
	f := SampleUsageHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageHold(f); err == nil {
		t.Error("ParseUsageHold() accepted a record with no id")
	}
	f = SampleUsageHold("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseUsageHold(f); err == nil {
		t.Error("ParseUsageHold() accepted a non-numeric count")
	}
}

func TestUsageHoldCollections(t *testing.T) {
	xs := []UsageHold{SampleUsageHold("a", when), SampleUsageHold("bb", when), SampleUsageHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageHolds(xs)); got != 2 {
		t.Errorf("ActiveUsageHolds() kept %d, want 2", got)
	}
	if got, want := SumUsageHolds(xs), 1+3; got != want {
		t.Errorf("SumUsageHolds() = %d, want %d", got, want)
	}
	SortUsageHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageHolds() put %q first", xs[0].ID)
	}
	if len(IndexUsageHolds(xs)) != 3 {
		t.Errorf("IndexUsageHolds() = %v", IndexUsageHolds(xs))
	}
	if err := ValidateUsageHolds(ActiveUsageHolds(xs)); err != nil {
		t.Errorf("ValidateUsageHolds() = %v", err)
	}
}

func TestUsageHoldNormalizeClamps(t *testing.T) {
	x := SampleUsageHold("a", when)
	x.Source = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageHold("r1", when)
	rev, err := r.PutUsageHold(x, 0)
	if err != nil {
		t.Fatalf("PutUsageHold() = %v", err)
	}
	got, gotRev, err := r.GetUsageHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageHolds() = %v, %v", all, err)
	}
	if _, err := r.PutUsageHold(UsageHold{}, 0); err == nil {
		t.Error("PutUsageHold() accepted an empty entity")
	}
}

func TestUsageClaimRoundTrip(t *testing.T) {
	x := SampleUsageClaim("usageclaim-1", when)
	got, err := ParseUsageClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageClaim() = %v", err)
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

func TestUsageClaimParseRejectsBadInput(t *testing.T) {
	f := SampleUsageClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageClaim(f); err == nil {
		t.Error("ParseUsageClaim() accepted a record with no id")
	}
	f = SampleUsageClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseUsageClaim(f); err == nil {
		t.Error("ParseUsageClaim() accepted a non-numeric retained")
	}
}

func TestUsageClaimCollections(t *testing.T) {
	xs := []UsageClaim{SampleUsageClaim("a", when), SampleUsageClaim("bb", when), SampleUsageClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageClaims(xs)); got != 2 {
		t.Errorf("ActiveUsageClaims() kept %d, want 2", got)
	}
	if got, want := SumUsageClaims(xs), 1+3; got != want {
		t.Errorf("SumUsageClaims() = %d, want %d", got, want)
	}
	SortUsageClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageClaims() put %q first", xs[0].ID)
	}
	if len(IndexUsageClaims(xs)) != 3 {
		t.Errorf("IndexUsageClaims() = %v", IndexUsageClaims(xs))
	}
	if err := ValidateUsageClaims(ActiveUsageClaims(xs)); err != nil {
		t.Errorf("ValidateUsageClaims() = %v", err)
	}
}

func TestUsageClaimNormalizeClamps(t *testing.T) {
	x := SampleUsageClaim("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageClaim("r1", when)
	rev, err := r.PutUsageClaim(x, 0)
	if err != nil {
		t.Fatalf("PutUsageClaim() = %v", err)
	}
	got, gotRev, err := r.GetUsageClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageClaims() = %v, %v", all, err)
	}
	if _, err := r.PutUsageClaim(UsageClaim{}, 0); err == nil {
		t.Error("PutUsageClaim() accepted an empty entity")
	}
}

func TestUsageCursorRoundTrip(t *testing.T) {
	x := SampleUsageCursor("usagecursor-1", when)
	got, err := ParseUsageCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseUsageCursor() = %v", err)
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

func TestUsageCursorParseRejectsBadInput(t *testing.T) {
	f := SampleUsageCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUsageCursor(f); err == nil {
		t.Error("ParseUsageCursor() accepted a record with no id")
	}
	f = SampleUsageCursor("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseUsageCursor(f); err == nil {
		t.Error("ParseUsageCursor() accepted a non-numeric retained")
	}
}

func TestUsageCursorCollections(t *testing.T) {
	xs := []UsageCursor{SampleUsageCursor("a", when), SampleUsageCursor("bb", when), SampleUsageCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUsageCursors(xs)); got != 2 {
		t.Errorf("ActiveUsageCursors() kept %d, want 2", got)
	}
	if got, want := SumUsageCursors(xs), 1+3; got != want {
		t.Errorf("SumUsageCursors() = %d, want %d", got, want)
	}
	SortUsageCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUsageCursors() put %q first", xs[0].ID)
	}
	if len(IndexUsageCursors(xs)) != 3 {
		t.Errorf("IndexUsageCursors() = %v", IndexUsageCursors(xs))
	}
	if err := ValidateUsageCursors(ActiveUsageCursors(xs)); err != nil {
		t.Errorf("ValidateUsageCursors() = %v", err)
	}
}

func TestUsageCursorNormalizeClamps(t *testing.T) {
	x := SampleUsageCursor("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUsageCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleUsageCursor("r1", when)
	rev, err := r.PutUsageCursor(x, 0)
	if err != nil {
		t.Fatalf("PutUsageCursor() = %v", err)
	}
	got, gotRev, err := r.GetUsageCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetUsageCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListUsageCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListUsageCursors() = %v, %v", all, err)
	}
	if _, err := r.PutUsageCursor(UsageCursor{}, 0); err == nil {
		t.Error("PutUsageCursor() accepted an empty entity")
	}
}
