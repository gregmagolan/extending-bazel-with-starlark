package settlement

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
		if !strings.HasPrefix(Prefixes[i], "settlement/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleSettlementRecord("ok", when).Fields()); err != nil {
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

func TestSettlementRecordRoundTrip(t *testing.T) {
	x := SampleSettlementRecord("settlementrecord-1", when)
	got, err := ParseSettlementRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementRecord() = %v", err)
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

func TestSettlementRecordParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementRecord(f); err == nil {
		t.Error("ParseSettlementRecord() accepted a record with no id")
	}
	f = SampleSettlementRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSettlementRecord(f); err == nil {
		t.Error("ParseSettlementRecord() accepted a non-numeric capacity")
	}
}

func TestSettlementRecordCollections(t *testing.T) {
	xs := []SettlementRecord{SampleSettlementRecord("a", when), SampleSettlementRecord("bb", when), SampleSettlementRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementRecords(xs)); got != 2 {
		t.Errorf("ActiveSettlementRecords() kept %d, want 2", got)
	}
	if got, want := SumSettlementRecords(xs), 1+3; got != want {
		t.Errorf("SumSettlementRecords() = %d, want %d", got, want)
	}
	SortSettlementRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementRecords() put %q first", xs[0].ID)
	}
	if len(IndexSettlementRecords(xs)) != 3 {
		t.Errorf("IndexSettlementRecords() = %v", IndexSettlementRecords(xs))
	}
	if err := ValidateSettlementRecords(ActiveSettlementRecords(xs)); err != nil {
		t.Errorf("ValidateSettlementRecords() = %v", err)
	}
}

func TestSettlementRecordNormalizeClamps(t *testing.T) {
	x := SampleSettlementRecord("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementRecord("r1", when)
	rev, err := r.PutSettlementRecord(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementRecord() = %v", err)
	}
	got, gotRev, err := r.GetSettlementRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementRecords() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementRecord(SettlementRecord{}, 0); err == nil {
		t.Error("PutSettlementRecord() accepted an empty entity")
	}
}

func TestSettlementEntryRoundTrip(t *testing.T) {
	x := SampleSettlementEntry("settlemententry-1", when)
	got, err := ParseSettlementEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementEntry() = %v", err)
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

func TestSettlementEntryParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementEntry(f); err == nil {
		t.Error("ParseSettlementEntry() accepted a record with no id")
	}
	f = SampleSettlementEntry("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSettlementEntry(f); err == nil {
		t.Error("ParseSettlementEntry() accepted a non-numeric count")
	}
}

func TestSettlementEntryCollections(t *testing.T) {
	xs := []SettlementEntry{SampleSettlementEntry("a", when), SampleSettlementEntry("bb", when), SampleSettlementEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementEntrys(xs)); got != 2 {
		t.Errorf("ActiveSettlementEntrys() kept %d, want 2", got)
	}
	if got, want := SumSettlementEntrys(xs), 1+3; got != want {
		t.Errorf("SumSettlementEntrys() = %d, want %d", got, want)
	}
	SortSettlementEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementEntrys() put %q first", xs[0].ID)
	}
	if len(IndexSettlementEntrys(xs)) != 3 {
		t.Errorf("IndexSettlementEntrys() = %v", IndexSettlementEntrys(xs))
	}
	if err := ValidateSettlementEntrys(ActiveSettlementEntrys(xs)); err != nil {
		t.Errorf("ValidateSettlementEntrys() = %v", err)
	}
}

func TestSettlementEntryNormalizeClamps(t *testing.T) {
	x := SampleSettlementEntry("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementEntry("r1", when)
	rev, err := r.PutSettlementEntry(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementEntry() = %v", err)
	}
	got, gotRev, err := r.GetSettlementEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementEntry(SettlementEntry{}, 0); err == nil {
		t.Error("PutSettlementEntry() accepted an empty entity")
	}
}

func TestSettlementItemRoundTrip(t *testing.T) {
	x := SampleSettlementItem("settlementitem-1", when)
	got, err := ParseSettlementItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementItem() = %v", err)
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

func TestSettlementItemParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementItem(f); err == nil {
		t.Error("ParseSettlementItem() accepted a record with no id")
	}
	f = SampleSettlementItem("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSettlementItem(f); err == nil {
		t.Error("ParseSettlementItem() accepted a non-numeric priority")
	}
}

func TestSettlementItemCollections(t *testing.T) {
	xs := []SettlementItem{SampleSettlementItem("a", when), SampleSettlementItem("bb", when), SampleSettlementItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementItems(xs)); got != 2 {
		t.Errorf("ActiveSettlementItems() kept %d, want 2", got)
	}
	if got, want := SumSettlementItems(xs), 1+3; got != want {
		t.Errorf("SumSettlementItems() = %d, want %d", got, want)
	}
	SortSettlementItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementItems() put %q first", xs[0].ID)
	}
	if len(IndexSettlementItems(xs)) != 3 {
		t.Errorf("IndexSettlementItems() = %v", IndexSettlementItems(xs))
	}
	if err := ValidateSettlementItems(ActiveSettlementItems(xs)); err != nil {
		t.Errorf("ValidateSettlementItems() = %v", err)
	}
}

func TestSettlementItemNormalizeClamps(t *testing.T) {
	x := SampleSettlementItem("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementItem("r1", when)
	rev, err := r.PutSettlementItem(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementItem() = %v", err)
	}
	got, gotRev, err := r.GetSettlementItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementItems() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementItem(SettlementItem{}, 0); err == nil {
		t.Error("PutSettlementItem() accepted an empty entity")
	}
}

func TestSettlementDraftRoundTrip(t *testing.T) {
	x := SampleSettlementDraft("settlementdraft-1", when)
	got, err := ParseSettlementDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementDraft() = %v", err)
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

func TestSettlementDraftParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementDraft(f); err == nil {
		t.Error("ParseSettlementDraft() accepted a record with no id")
	}
	f = SampleSettlementDraft("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseSettlementDraft(f); err == nil {
		t.Error("ParseSettlementDraft() accepted a non-numeric retained")
	}
}

func TestSettlementDraftCollections(t *testing.T) {
	xs := []SettlementDraft{SampleSettlementDraft("a", when), SampleSettlementDraft("bb", when), SampleSettlementDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementDrafts(xs)); got != 2 {
		t.Errorf("ActiveSettlementDrafts() kept %d, want 2", got)
	}
	if got, want := SumSettlementDrafts(xs), 1+3; got != want {
		t.Errorf("SumSettlementDrafts() = %d, want %d", got, want)
	}
	SortSettlementDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementDrafts() put %q first", xs[0].ID)
	}
	if len(IndexSettlementDrafts(xs)) != 3 {
		t.Errorf("IndexSettlementDrafts() = %v", IndexSettlementDrafts(xs))
	}
	if err := ValidateSettlementDrafts(ActiveSettlementDrafts(xs)); err != nil {
		t.Errorf("ValidateSettlementDrafts() = %v", err)
	}
}

func TestSettlementDraftNormalizeClamps(t *testing.T) {
	x := SampleSettlementDraft("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementDraft("r1", when)
	rev, err := r.PutSettlementDraft(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementDraft() = %v", err)
	}
	got, gotRev, err := r.GetSettlementDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementDraft(SettlementDraft{}, 0); err == nil {
		t.Error("PutSettlementDraft() accepted an empty entity")
	}
}

func TestSettlementRevisionRoundTrip(t *testing.T) {
	x := SampleSettlementRevision("settlementrevision-1", when)
	got, err := ParseSettlementRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementRevision() = %v", err)
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

func TestSettlementRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementRevision(f); err == nil {
		t.Error("ParseSettlementRevision() accepted a record with no id")
	}
	f = SampleSettlementRevision("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSettlementRevision(f); err == nil {
		t.Error("ParseSettlementRevision() accepted a non-numeric version")
	}
}

func TestSettlementRevisionCollections(t *testing.T) {
	xs := []SettlementRevision{SampleSettlementRevision("a", when), SampleSettlementRevision("bb", when), SampleSettlementRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementRevisions(xs)); got != 2 {
		t.Errorf("ActiveSettlementRevisions() kept %d, want 2", got)
	}
	if got, want := SumSettlementRevisions(xs), 1+3; got != want {
		t.Errorf("SumSettlementRevisions() = %d, want %d", got, want)
	}
	SortSettlementRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementRevisions() put %q first", xs[0].ID)
	}
	if len(IndexSettlementRevisions(xs)) != 3 {
		t.Errorf("IndexSettlementRevisions() = %v", IndexSettlementRevisions(xs))
	}
	if err := ValidateSettlementRevisions(ActiveSettlementRevisions(xs)); err != nil {
		t.Errorf("ValidateSettlementRevisions() = %v", err)
	}
}

func TestSettlementRevisionNormalizeClamps(t *testing.T) {
	x := SampleSettlementRevision("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementRevision("r1", when)
	rev, err := r.PutSettlementRevision(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementRevision() = %v", err)
	}
	got, gotRev, err := r.GetSettlementRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementRevision(SettlementRevision{}, 0); err == nil {
		t.Error("PutSettlementRevision() accepted an empty entity")
	}
}

func TestSettlementSnapshotRoundTrip(t *testing.T) {
	x := SampleSettlementSnapshot("settlementsnapshot-1", when)
	got, err := ParseSettlementSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementSnapshot() = %v", err)
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

func TestSettlementSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementSnapshot(f); err == nil {
		t.Error("ParseSettlementSnapshot() accepted a record with no id")
	}
	f = SampleSettlementSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSettlementSnapshot(f); err == nil {
		t.Error("ParseSettlementSnapshot() accepted a non-numeric priority")
	}
}

func TestSettlementSnapshotCollections(t *testing.T) {
	xs := []SettlementSnapshot{SampleSettlementSnapshot("a", when), SampleSettlementSnapshot("bb", when), SampleSettlementSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementSnapshots(xs)); got != 2 {
		t.Errorf("ActiveSettlementSnapshots() kept %d, want 2", got)
	}
	if got, want := SumSettlementSnapshots(xs), 1+3; got != want {
		t.Errorf("SumSettlementSnapshots() = %d, want %d", got, want)
	}
	SortSettlementSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexSettlementSnapshots(xs)) != 3 {
		t.Errorf("IndexSettlementSnapshots() = %v", IndexSettlementSnapshots(xs))
	}
	if err := ValidateSettlementSnapshots(ActiveSettlementSnapshots(xs)); err != nil {
		t.Errorf("ValidateSettlementSnapshots() = %v", err)
	}
}

func TestSettlementSnapshotNormalizeClamps(t *testing.T) {
	x := SampleSettlementSnapshot("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementSnapshot("r1", when)
	rev, err := r.PutSettlementSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetSettlementSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementSnapshot(SettlementSnapshot{}, 0); err == nil {
		t.Error("PutSettlementSnapshot() accepted an empty entity")
	}
}

func TestSettlementBatchRoundTrip(t *testing.T) {
	x := SampleSettlementBatch("settlementbatch-1", when)
	got, err := ParseSettlementBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementBatch() = %v", err)
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

func TestSettlementBatchParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementBatch(f); err == nil {
		t.Error("ParseSettlementBatch() accepted a record with no id")
	}
	f = SampleSettlementBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSettlementBatch(f); err == nil {
		t.Error("ParseSettlementBatch() accepted a non-numeric version")
	}
}

func TestSettlementBatchCollections(t *testing.T) {
	xs := []SettlementBatch{SampleSettlementBatch("a", when), SampleSettlementBatch("bb", when), SampleSettlementBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementBatchs(xs)); got != 2 {
		t.Errorf("ActiveSettlementBatchs() kept %d, want 2", got)
	}
	if got, want := SumSettlementBatchs(xs), 1+3; got != want {
		t.Errorf("SumSettlementBatchs() = %d, want %d", got, want)
	}
	SortSettlementBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementBatchs() put %q first", xs[0].ID)
	}
	if len(IndexSettlementBatchs(xs)) != 3 {
		t.Errorf("IndexSettlementBatchs() = %v", IndexSettlementBatchs(xs))
	}
	if err := ValidateSettlementBatchs(ActiveSettlementBatchs(xs)); err != nil {
		t.Errorf("ValidateSettlementBatchs() = %v", err)
	}
}

func TestSettlementBatchNormalizeClamps(t *testing.T) {
	x := SampleSettlementBatch("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementBatch("r1", when)
	rev, err := r.PutSettlementBatch(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementBatch() = %v", err)
	}
	got, gotRev, err := r.GetSettlementBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementBatch(SettlementBatch{}, 0); err == nil {
		t.Error("PutSettlementBatch() accepted an empty entity")
	}
}

func TestSettlementRequestRoundTrip(t *testing.T) {
	x := SampleSettlementRequest("settlementrequest-1", when)
	got, err := ParseSettlementRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementRequest() = %v", err)
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

func TestSettlementRequestParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementRequest(f); err == nil {
		t.Error("ParseSettlementRequest() accepted a record with no id")
	}
	f = SampleSettlementRequest("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseSettlementRequest(f); err == nil {
		t.Error("ParseSettlementRequest() accepted a non-numeric attempts")
	}
}

func TestSettlementRequestCollections(t *testing.T) {
	xs := []SettlementRequest{SampleSettlementRequest("a", when), SampleSettlementRequest("bb", when), SampleSettlementRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementRequests(xs)); got != 2 {
		t.Errorf("ActiveSettlementRequests() kept %d, want 2", got)
	}
	if got, want := SumSettlementRequests(xs), 1+3; got != want {
		t.Errorf("SumSettlementRequests() = %d, want %d", got, want)
	}
	SortSettlementRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementRequests() put %q first", xs[0].ID)
	}
	if len(IndexSettlementRequests(xs)) != 3 {
		t.Errorf("IndexSettlementRequests() = %v", IndexSettlementRequests(xs))
	}
	if err := ValidateSettlementRequests(ActiveSettlementRequests(xs)); err != nil {
		t.Errorf("ValidateSettlementRequests() = %v", err)
	}
}

func TestSettlementRequestNormalizeClamps(t *testing.T) {
	x := SampleSettlementRequest("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementRequest("r1", when)
	rev, err := r.PutSettlementRequest(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementRequest() = %v", err)
	}
	got, gotRev, err := r.GetSettlementRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementRequests() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementRequest(SettlementRequest{}, 0); err == nil {
		t.Error("PutSettlementRequest() accepted an empty entity")
	}
}

func TestSettlementSummaryRoundTrip(t *testing.T) {
	x := SampleSettlementSummary("settlementsummary-1", when)
	got, err := ParseSettlementSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementSummary() = %v", err)
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

func TestSettlementSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementSummary(f); err == nil {
		t.Error("ParseSettlementSummary() accepted a record with no id")
	}
	f = SampleSettlementSummary("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSettlementSummary(f); err == nil {
		t.Error("ParseSettlementSummary() accepted a non-numeric version")
	}
}

func TestSettlementSummaryCollections(t *testing.T) {
	xs := []SettlementSummary{SampleSettlementSummary("a", when), SampleSettlementSummary("bb", when), SampleSettlementSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementSummarys(xs)); got != 2 {
		t.Errorf("ActiveSettlementSummarys() kept %d, want 2", got)
	}
	if got, want := SumSettlementSummarys(xs), 1+3; got != want {
		t.Errorf("SumSettlementSummarys() = %d, want %d", got, want)
	}
	SortSettlementSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementSummarys() put %q first", xs[0].ID)
	}
	if len(IndexSettlementSummarys(xs)) != 3 {
		t.Errorf("IndexSettlementSummarys() = %v", IndexSettlementSummarys(xs))
	}
	if err := ValidateSettlementSummarys(ActiveSettlementSummarys(xs)); err != nil {
		t.Errorf("ValidateSettlementSummarys() = %v", err)
	}
}

func TestSettlementSummaryNormalizeClamps(t *testing.T) {
	x := SampleSettlementSummary("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementSummary("r1", when)
	rev, err := r.PutSettlementSummary(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementSummary() = %v", err)
	}
	got, gotRev, err := r.GetSettlementSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementSummary(SettlementSummary{}, 0); err == nil {
		t.Error("PutSettlementSummary() accepted an empty entity")
	}
}

func TestSettlementPolicyRoundTrip(t *testing.T) {
	x := SampleSettlementPolicy("settlementpolicy-1", when)
	got, err := ParseSettlementPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementPolicy() = %v", err)
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

func TestSettlementPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementPolicy(f); err == nil {
		t.Error("ParseSettlementPolicy() accepted a record with no id")
	}
	f = SampleSettlementPolicy("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseSettlementPolicy(f); err == nil {
		t.Error("ParseSettlementPolicy() accepted a non-numeric retained")
	}
}

func TestSettlementPolicyCollections(t *testing.T) {
	xs := []SettlementPolicy{SampleSettlementPolicy("a", when), SampleSettlementPolicy("bb", when), SampleSettlementPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementPolicys(xs)); got != 2 {
		t.Errorf("ActiveSettlementPolicys() kept %d, want 2", got)
	}
	if got, want := SumSettlementPolicys(xs), 1+3; got != want {
		t.Errorf("SumSettlementPolicys() = %d, want %d", got, want)
	}
	SortSettlementPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementPolicys() put %q first", xs[0].ID)
	}
	if len(IndexSettlementPolicys(xs)) != 3 {
		t.Errorf("IndexSettlementPolicys() = %v", IndexSettlementPolicys(xs))
	}
	if err := ValidateSettlementPolicys(ActiveSettlementPolicys(xs)); err != nil {
		t.Errorf("ValidateSettlementPolicys() = %v", err)
	}
}

func TestSettlementPolicyNormalizeClamps(t *testing.T) {
	x := SampleSettlementPolicy("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementPolicy("r1", when)
	rev, err := r.PutSettlementPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementPolicy() = %v", err)
	}
	got, gotRev, err := r.GetSettlementPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementPolicy(SettlementPolicy{}, 0); err == nil {
		t.Error("PutSettlementPolicy() accepted an empty entity")
	}
}

func TestSettlementLinkRoundTrip(t *testing.T) {
	x := SampleSettlementLink("settlementlink-1", when)
	got, err := ParseSettlementLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementLink() = %v", err)
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

func TestSettlementLinkParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementLink(f); err == nil {
		t.Error("ParseSettlementLink() accepted a record with no id")
	}
	f = SampleSettlementLink("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSettlementLink(f); err == nil {
		t.Error("ParseSettlementLink() accepted a non-numeric offset")
	}
}

func TestSettlementLinkCollections(t *testing.T) {
	xs := []SettlementLink{SampleSettlementLink("a", when), SampleSettlementLink("bb", when), SampleSettlementLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementLinks(xs)); got != 2 {
		t.Errorf("ActiveSettlementLinks() kept %d, want 2", got)
	}
	if got, want := SumSettlementLinks(xs), 1+3; got != want {
		t.Errorf("SumSettlementLinks() = %d, want %d", got, want)
	}
	SortSettlementLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementLinks() put %q first", xs[0].ID)
	}
	if len(IndexSettlementLinks(xs)) != 3 {
		t.Errorf("IndexSettlementLinks() = %v", IndexSettlementLinks(xs))
	}
	if err := ValidateSettlementLinks(ActiveSettlementLinks(xs)); err != nil {
		t.Errorf("ValidateSettlementLinks() = %v", err)
	}
}

func TestSettlementLinkNormalizeClamps(t *testing.T) {
	x := SampleSettlementLink("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementLink("r1", when)
	rev, err := r.PutSettlementLink(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementLink() = %v", err)
	}
	got, gotRev, err := r.GetSettlementLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementLinks() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementLink(SettlementLink{}, 0); err == nil {
		t.Error("PutSettlementLink() accepted an empty entity")
	}
}

func TestSettlementNoteRoundTrip(t *testing.T) {
	x := SampleSettlementNote("settlementnote-1", when)
	got, err := ParseSettlementNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementNote() = %v", err)
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

func TestSettlementNoteParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementNote(f); err == nil {
		t.Error("ParseSettlementNote() accepted a record with no id")
	}
	f = SampleSettlementNote("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSettlementNote(f); err == nil {
		t.Error("ParseSettlementNote() accepted a non-numeric count")
	}
}

func TestSettlementNoteCollections(t *testing.T) {
	xs := []SettlementNote{SampleSettlementNote("a", when), SampleSettlementNote("bb", when), SampleSettlementNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementNotes(xs)); got != 2 {
		t.Errorf("ActiveSettlementNotes() kept %d, want 2", got)
	}
	if got, want := SumSettlementNotes(xs), 1+3; got != want {
		t.Errorf("SumSettlementNotes() = %d, want %d", got, want)
	}
	SortSettlementNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementNotes() put %q first", xs[0].ID)
	}
	if len(IndexSettlementNotes(xs)) != 3 {
		t.Errorf("IndexSettlementNotes() = %v", IndexSettlementNotes(xs))
	}
	if err := ValidateSettlementNotes(ActiveSettlementNotes(xs)); err != nil {
		t.Errorf("ValidateSettlementNotes() = %v", err)
	}
}

func TestSettlementNoteNormalizeClamps(t *testing.T) {
	x := SampleSettlementNote("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementNote("r1", when)
	rev, err := r.PutSettlementNote(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementNote() = %v", err)
	}
	got, gotRev, err := r.GetSettlementNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementNotes() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementNote(SettlementNote{}, 0); err == nil {
		t.Error("PutSettlementNote() accepted an empty entity")
	}
}

func TestSettlementWindowRoundTrip(t *testing.T) {
	x := SampleSettlementWindow("settlementwindow-1", when)
	got, err := ParseSettlementWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementWindow() = %v", err)
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

func TestSettlementWindowParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementWindow(f); err == nil {
		t.Error("ParseSettlementWindow() accepted a record with no id")
	}
	f = SampleSettlementWindow("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseSettlementWindow(f); err == nil {
		t.Error("ParseSettlementWindow() accepted a non-numeric attempts")
	}
}

func TestSettlementWindowCollections(t *testing.T) {
	xs := []SettlementWindow{SampleSettlementWindow("a", when), SampleSettlementWindow("bb", when), SampleSettlementWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementWindows(xs)); got != 2 {
		t.Errorf("ActiveSettlementWindows() kept %d, want 2", got)
	}
	if got, want := SumSettlementWindows(xs), 1+3; got != want {
		t.Errorf("SumSettlementWindows() = %d, want %d", got, want)
	}
	SortSettlementWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementWindows() put %q first", xs[0].ID)
	}
	if len(IndexSettlementWindows(xs)) != 3 {
		t.Errorf("IndexSettlementWindows() = %v", IndexSettlementWindows(xs))
	}
	if err := ValidateSettlementWindows(ActiveSettlementWindows(xs)); err != nil {
		t.Errorf("ValidateSettlementWindows() = %v", err)
	}
}

func TestSettlementWindowNormalizeClamps(t *testing.T) {
	x := SampleSettlementWindow("a", when)
	x.Reference = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementWindow("r1", when)
	rev, err := r.PutSettlementWindow(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementWindow() = %v", err)
	}
	got, gotRev, err := r.GetSettlementWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementWindows() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementWindow(SettlementWindow{}, 0); err == nil {
		t.Error("PutSettlementWindow() accepted an empty entity")
	}
}

func TestSettlementQuotaRoundTrip(t *testing.T) {
	x := SampleSettlementQuota("settlementquota-1", when)
	got, err := ParseSettlementQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementQuota() = %v", err)
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

func TestSettlementQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementQuota(f); err == nil {
		t.Error("ParseSettlementQuota() accepted a record with no id")
	}
	f = SampleSettlementQuota("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseSettlementQuota(f); err == nil {
		t.Error("ParseSettlementQuota() accepted a non-numeric weight")
	}
}

func TestSettlementQuotaCollections(t *testing.T) {
	xs := []SettlementQuota{SampleSettlementQuota("a", when), SampleSettlementQuota("bb", when), SampleSettlementQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementQuotas(xs)); got != 2 {
		t.Errorf("ActiveSettlementQuotas() kept %d, want 2", got)
	}
	if got, want := SumSettlementQuotas(xs), 1+3; got != want {
		t.Errorf("SumSettlementQuotas() = %d, want %d", got, want)
	}
	SortSettlementQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementQuotas() put %q first", xs[0].ID)
	}
	if len(IndexSettlementQuotas(xs)) != 3 {
		t.Errorf("IndexSettlementQuotas() = %v", IndexSettlementQuotas(xs))
	}
	if err := ValidateSettlementQuotas(ActiveSettlementQuotas(xs)); err != nil {
		t.Errorf("ValidateSettlementQuotas() = %v", err)
	}
}

func TestSettlementQuotaNormalizeClamps(t *testing.T) {
	x := SampleSettlementQuota("a", when)
	x.Comment = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementQuota("r1", when)
	rev, err := r.PutSettlementQuota(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementQuota() = %v", err)
	}
	got, gotRev, err := r.GetSettlementQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementQuota(SettlementQuota{}, 0); err == nil {
		t.Error("PutSettlementQuota() accepted an empty entity")
	}
}

func TestSettlementBucketRoundTrip(t *testing.T) {
	x := SampleSettlementBucket("settlementbucket-1", when)
	got, err := ParseSettlementBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementBucket() = %v", err)
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

func TestSettlementBucketParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementBucket(f); err == nil {
		t.Error("ParseSettlementBucket() accepted a record with no id")
	}
	f = SampleSettlementBucket("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSettlementBucket(f); err == nil {
		t.Error("ParseSettlementBucket() accepted a non-numeric offset")
	}
}

func TestSettlementBucketCollections(t *testing.T) {
	xs := []SettlementBucket{SampleSettlementBucket("a", when), SampleSettlementBucket("bb", when), SampleSettlementBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementBuckets(xs)); got != 2 {
		t.Errorf("ActiveSettlementBuckets() kept %d, want 2", got)
	}
	if got, want := SumSettlementBuckets(xs), 1+3; got != want {
		t.Errorf("SumSettlementBuckets() = %d, want %d", got, want)
	}
	SortSettlementBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementBuckets() put %q first", xs[0].ID)
	}
	if len(IndexSettlementBuckets(xs)) != 3 {
		t.Errorf("IndexSettlementBuckets() = %v", IndexSettlementBuckets(xs))
	}
	if err := ValidateSettlementBuckets(ActiveSettlementBuckets(xs)); err != nil {
		t.Errorf("ValidateSettlementBuckets() = %v", err)
	}
}

func TestSettlementBucketNormalizeClamps(t *testing.T) {
	x := SampleSettlementBucket("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementBucket("r1", when)
	rev, err := r.PutSettlementBucket(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementBucket() = %v", err)
	}
	got, gotRev, err := r.GetSettlementBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementBucket(SettlementBucket{}, 0); err == nil {
		t.Error("PutSettlementBucket() accepted an empty entity")
	}
}

func TestSettlementLaneRoundTrip(t *testing.T) {
	x := SampleSettlementLane("settlementlane-1", when)
	got, err := ParseSettlementLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementLane() = %v", err)
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

func TestSettlementLaneParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementLane(f); err == nil {
		t.Error("ParseSettlementLane() accepted a record with no id")
	}
	f = SampleSettlementLane("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseSettlementLane(f); err == nil {
		t.Error("ParseSettlementLane() accepted a non-numeric weight")
	}
}

func TestSettlementLaneCollections(t *testing.T) {
	xs := []SettlementLane{SampleSettlementLane("a", when), SampleSettlementLane("bb", when), SampleSettlementLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementLanes(xs)); got != 2 {
		t.Errorf("ActiveSettlementLanes() kept %d, want 2", got)
	}
	if got, want := SumSettlementLanes(xs), 1+3; got != want {
		t.Errorf("SumSettlementLanes() = %d, want %d", got, want)
	}
	SortSettlementLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementLanes() put %q first", xs[0].ID)
	}
	if len(IndexSettlementLanes(xs)) != 3 {
		t.Errorf("IndexSettlementLanes() = %v", IndexSettlementLanes(xs))
	}
	if err := ValidateSettlementLanes(ActiveSettlementLanes(xs)); err != nil {
		t.Errorf("ValidateSettlementLanes() = %v", err)
	}
}

func TestSettlementLaneNormalizeClamps(t *testing.T) {
	x := SampleSettlementLane("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementLane("r1", when)
	rev, err := r.PutSettlementLane(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementLane() = %v", err)
	}
	got, gotRev, err := r.GetSettlementLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementLanes() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementLane(SettlementLane{}, 0); err == nil {
		t.Error("PutSettlementLane() accepted an empty entity")
	}
}

func TestSettlementSlotRoundTrip(t *testing.T) {
	x := SampleSettlementSlot("settlementslot-1", when)
	got, err := ParseSettlementSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementSlot() = %v", err)
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

func TestSettlementSlotParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementSlot(f); err == nil {
		t.Error("ParseSettlementSlot() accepted a record with no id")
	}
	f = SampleSettlementSlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSettlementSlot(f); err == nil {
		t.Error("ParseSettlementSlot() accepted a non-numeric count")
	}
}

func TestSettlementSlotCollections(t *testing.T) {
	xs := []SettlementSlot{SampleSettlementSlot("a", when), SampleSettlementSlot("bb", when), SampleSettlementSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementSlots(xs)); got != 2 {
		t.Errorf("ActiveSettlementSlots() kept %d, want 2", got)
	}
	if got, want := SumSettlementSlots(xs), 1+3; got != want {
		t.Errorf("SumSettlementSlots() = %d, want %d", got, want)
	}
	SortSettlementSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementSlots() put %q first", xs[0].ID)
	}
	if len(IndexSettlementSlots(xs)) != 3 {
		t.Errorf("IndexSettlementSlots() = %v", IndexSettlementSlots(xs))
	}
	if err := ValidateSettlementSlots(ActiveSettlementSlots(xs)); err != nil {
		t.Errorf("ValidateSettlementSlots() = %v", err)
	}
}

func TestSettlementSlotNormalizeClamps(t *testing.T) {
	x := SampleSettlementSlot("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementSlot("r1", when)
	rev, err := r.PutSettlementSlot(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementSlot() = %v", err)
	}
	got, gotRev, err := r.GetSettlementSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementSlots() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementSlot(SettlementSlot{}, 0); err == nil {
		t.Error("PutSettlementSlot() accepted an empty entity")
	}
}

func TestSettlementTagRoundTrip(t *testing.T) {
	x := SampleSettlementTag("settlementtag-1", when)
	got, err := ParseSettlementTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementTag() = %v", err)
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

func TestSettlementTagParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementTag(f); err == nil {
		t.Error("ParseSettlementTag() accepted a record with no id")
	}
	f = SampleSettlementTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSettlementTag(f); err == nil {
		t.Error("ParseSettlementTag() accepted a non-numeric count")
	}
}

func TestSettlementTagCollections(t *testing.T) {
	xs := []SettlementTag{SampleSettlementTag("a", when), SampleSettlementTag("bb", when), SampleSettlementTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementTags(xs)); got != 2 {
		t.Errorf("ActiveSettlementTags() kept %d, want 2", got)
	}
	if got, want := SumSettlementTags(xs), 1+3; got != want {
		t.Errorf("SumSettlementTags() = %d, want %d", got, want)
	}
	SortSettlementTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementTags() put %q first", xs[0].ID)
	}
	if len(IndexSettlementTags(xs)) != 3 {
		t.Errorf("IndexSettlementTags() = %v", IndexSettlementTags(xs))
	}
	if err := ValidateSettlementTags(ActiveSettlementTags(xs)); err != nil {
		t.Errorf("ValidateSettlementTags() = %v", err)
	}
}

func TestSettlementTagNormalizeClamps(t *testing.T) {
	x := SampleSettlementTag("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementTag("r1", when)
	rev, err := r.PutSettlementTag(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementTag() = %v", err)
	}
	got, gotRev, err := r.GetSettlementTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementTags() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementTag(SettlementTag{}, 0); err == nil {
		t.Error("PutSettlementTag() accepted an empty entity")
	}
}

func TestSettlementRuleRoundTrip(t *testing.T) {
	x := SampleSettlementRule("settlementrule-1", when)
	got, err := ParseSettlementRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementRule() = %v", err)
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

func TestSettlementRuleParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementRule(f); err == nil {
		t.Error("ParseSettlementRule() accepted a record with no id")
	}
	f = SampleSettlementRule("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSettlementRule(f); err == nil {
		t.Error("ParseSettlementRule() accepted a non-numeric count")
	}
}

func TestSettlementRuleCollections(t *testing.T) {
	xs := []SettlementRule{SampleSettlementRule("a", when), SampleSettlementRule("bb", when), SampleSettlementRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementRules(xs)); got != 2 {
		t.Errorf("ActiveSettlementRules() kept %d, want 2", got)
	}
	if got, want := SumSettlementRules(xs), 1+3; got != want {
		t.Errorf("SumSettlementRules() = %d, want %d", got, want)
	}
	SortSettlementRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementRules() put %q first", xs[0].ID)
	}
	if len(IndexSettlementRules(xs)) != 3 {
		t.Errorf("IndexSettlementRules() = %v", IndexSettlementRules(xs))
	}
	if err := ValidateSettlementRules(ActiveSettlementRules(xs)); err != nil {
		t.Errorf("ValidateSettlementRules() = %v", err)
	}
}

func TestSettlementRuleNormalizeClamps(t *testing.T) {
	x := SampleSettlementRule("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementRule("r1", when)
	rev, err := r.PutSettlementRule(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementRule() = %v", err)
	}
	got, gotRev, err := r.GetSettlementRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementRules() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementRule(SettlementRule{}, 0); err == nil {
		t.Error("PutSettlementRule() accepted an empty entity")
	}
}

func TestSettlementHoldRoundTrip(t *testing.T) {
	x := SampleSettlementHold("settlementhold-1", when)
	got, err := ParseSettlementHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementHold() = %v", err)
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

func TestSettlementHoldParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementHold(f); err == nil {
		t.Error("ParseSettlementHold() accepted a record with no id")
	}
	f = SampleSettlementHold("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSettlementHold(f); err == nil {
		t.Error("ParseSettlementHold() accepted a non-numeric capacity")
	}
}

func TestSettlementHoldCollections(t *testing.T) {
	xs := []SettlementHold{SampleSettlementHold("a", when), SampleSettlementHold("bb", when), SampleSettlementHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementHolds(xs)); got != 2 {
		t.Errorf("ActiveSettlementHolds() kept %d, want 2", got)
	}
	if got, want := SumSettlementHolds(xs), 1+3; got != want {
		t.Errorf("SumSettlementHolds() = %d, want %d", got, want)
	}
	SortSettlementHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementHolds() put %q first", xs[0].ID)
	}
	if len(IndexSettlementHolds(xs)) != 3 {
		t.Errorf("IndexSettlementHolds() = %v", IndexSettlementHolds(xs))
	}
	if err := ValidateSettlementHolds(ActiveSettlementHolds(xs)); err != nil {
		t.Errorf("ValidateSettlementHolds() = %v", err)
	}
}

func TestSettlementHoldNormalizeClamps(t *testing.T) {
	x := SampleSettlementHold("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementHold("r1", when)
	rev, err := r.PutSettlementHold(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementHold() = %v", err)
	}
	got, gotRev, err := r.GetSettlementHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementHolds() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementHold(SettlementHold{}, 0); err == nil {
		t.Error("PutSettlementHold() accepted an empty entity")
	}
}

func TestSettlementClaimRoundTrip(t *testing.T) {
	x := SampleSettlementClaim("settlementclaim-1", when)
	got, err := ParseSettlementClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementClaim() = %v", err)
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

func TestSettlementClaimParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementClaim(f); err == nil {
		t.Error("ParseSettlementClaim() accepted a record with no id")
	}
	f = SampleSettlementClaim("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSettlementClaim(f); err == nil {
		t.Error("ParseSettlementClaim() accepted a non-numeric version")
	}
}

func TestSettlementClaimCollections(t *testing.T) {
	xs := []SettlementClaim{SampleSettlementClaim("a", when), SampleSettlementClaim("bb", when), SampleSettlementClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementClaims(xs)); got != 2 {
		t.Errorf("ActiveSettlementClaims() kept %d, want 2", got)
	}
	if got, want := SumSettlementClaims(xs), 1+3; got != want {
		t.Errorf("SumSettlementClaims() = %d, want %d", got, want)
	}
	SortSettlementClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementClaims() put %q first", xs[0].ID)
	}
	if len(IndexSettlementClaims(xs)) != 3 {
		t.Errorf("IndexSettlementClaims() = %v", IndexSettlementClaims(xs))
	}
	if err := ValidateSettlementClaims(ActiveSettlementClaims(xs)); err != nil {
		t.Errorf("ValidateSettlementClaims() = %v", err)
	}
}

func TestSettlementClaimNormalizeClamps(t *testing.T) {
	x := SampleSettlementClaim("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementClaim("r1", when)
	rev, err := r.PutSettlementClaim(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementClaim() = %v", err)
	}
	got, gotRev, err := r.GetSettlementClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementClaims() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementClaim(SettlementClaim{}, 0); err == nil {
		t.Error("PutSettlementClaim() accepted an empty entity")
	}
}

func TestSettlementCursorRoundTrip(t *testing.T) {
	x := SampleSettlementCursor("settlementcursor-1", when)
	got, err := ParseSettlementCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseSettlementCursor() = %v", err)
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

func TestSettlementCursorParseRejectsBadInput(t *testing.T) {
	f := SampleSettlementCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSettlementCursor(f); err == nil {
		t.Error("ParseSettlementCursor() accepted a record with no id")
	}
	f = SampleSettlementCursor("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSettlementCursor(f); err == nil {
		t.Error("ParseSettlementCursor() accepted a non-numeric count")
	}
}

func TestSettlementCursorCollections(t *testing.T) {
	xs := []SettlementCursor{SampleSettlementCursor("a", when), SampleSettlementCursor("bb", when), SampleSettlementCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSettlementCursors(xs)); got != 2 {
		t.Errorf("ActiveSettlementCursors() kept %d, want 2", got)
	}
	if got, want := SumSettlementCursors(xs), 1+3; got != want {
		t.Errorf("SumSettlementCursors() = %d, want %d", got, want)
	}
	SortSettlementCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSettlementCursors() put %q first", xs[0].ID)
	}
	if len(IndexSettlementCursors(xs)) != 3 {
		t.Errorf("IndexSettlementCursors() = %v", IndexSettlementCursors(xs))
	}
	if err := ValidateSettlementCursors(ActiveSettlementCursors(xs)); err != nil {
		t.Errorf("ValidateSettlementCursors() = %v", err)
	}
}

func TestSettlementCursorNormalizeClamps(t *testing.T) {
	x := SampleSettlementCursor("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSettlementCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSettlementCursor("r1", when)
	rev, err := r.PutSettlementCursor(x, 0)
	if err != nil {
		t.Fatalf("PutSettlementCursor() = %v", err)
	}
	got, gotRev, err := r.GetSettlementCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSettlementCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSettlementCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSettlementCursors() = %v, %v", all, err)
	}
	if _, err := r.PutSettlementCursor(SettlementCursor{}, 0); err == nil {
		t.Error("PutSettlementCursor() accepted an empty entity")
	}
}
