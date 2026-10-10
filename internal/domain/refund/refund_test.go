package refund

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
		if !strings.HasPrefix(Prefixes[i], "refund/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleRefundRecord("ok", when).Fields()); err != nil {
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

func TestRefundRecordRoundTrip(t *testing.T) {
	x := SampleRefundRecord("refundrecord-1", when)
	got, err := ParseRefundRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundRecord() = %v", err)
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

func TestRefundRecordParseRejectsBadInput(t *testing.T) {
	f := SampleRefundRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundRecord(f); err == nil {
		t.Error("ParseRefundRecord() accepted a record with no id")
	}
	f = SampleRefundRecord("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseRefundRecord(f); err == nil {
		t.Error("ParseRefundRecord() accepted a non-numeric version")
	}
}

func TestRefundRecordCollections(t *testing.T) {
	xs := []RefundRecord{SampleRefundRecord("a", when), SampleRefundRecord("bb", when), SampleRefundRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundRecords(xs)); got != 2 {
		t.Errorf("ActiveRefundRecords() kept %d, want 2", got)
	}
	if got, want := SumRefundRecords(xs), 1+3; got != want {
		t.Errorf("SumRefundRecords() = %d, want %d", got, want)
	}
	SortRefundRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundRecords() put %q first", xs[0].ID)
	}
	if len(IndexRefundRecords(xs)) != 3 {
		t.Errorf("IndexRefundRecords() = %v", IndexRefundRecords(xs))
	}
	if err := ValidateRefundRecords(ActiveRefundRecords(xs)); err != nil {
		t.Errorf("ValidateRefundRecords() = %v", err)
	}
}

func TestRefundRecordNormalizeClamps(t *testing.T) {
	x := SampleRefundRecord("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundRecord("r1", when)
	rev, err := r.PutRefundRecord(x, 0)
	if err != nil {
		t.Fatalf("PutRefundRecord() = %v", err)
	}
	got, gotRev, err := r.GetRefundRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundRecords() = %v, %v", all, err)
	}
	if _, err := r.PutRefundRecord(RefundRecord{}, 0); err == nil {
		t.Error("PutRefundRecord() accepted an empty entity")
	}
}

func TestRefundEntryRoundTrip(t *testing.T) {
	x := SampleRefundEntry("refundentry-1", when)
	got, err := ParseRefundEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundEntry() = %v", err)
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

func TestRefundEntryParseRejectsBadInput(t *testing.T) {
	f := SampleRefundEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundEntry(f); err == nil {
		t.Error("ParseRefundEntry() accepted a record with no id")
	}
	f = SampleRefundEntry("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRefundEntry(f); err == nil {
		t.Error("ParseRefundEntry() accepted a non-numeric priority")
	}
}

func TestRefundEntryCollections(t *testing.T) {
	xs := []RefundEntry{SampleRefundEntry("a", when), SampleRefundEntry("bb", when), SampleRefundEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundEntrys(xs)); got != 2 {
		t.Errorf("ActiveRefundEntrys() kept %d, want 2", got)
	}
	if got, want := SumRefundEntrys(xs), 1+3; got != want {
		t.Errorf("SumRefundEntrys() = %d, want %d", got, want)
	}
	SortRefundEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundEntrys() put %q first", xs[0].ID)
	}
	if len(IndexRefundEntrys(xs)) != 3 {
		t.Errorf("IndexRefundEntrys() = %v", IndexRefundEntrys(xs))
	}
	if err := ValidateRefundEntrys(ActiveRefundEntrys(xs)); err != nil {
		t.Errorf("ValidateRefundEntrys() = %v", err)
	}
}

func TestRefundEntryNormalizeClamps(t *testing.T) {
	x := SampleRefundEntry("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundEntry("r1", when)
	rev, err := r.PutRefundEntry(x, 0)
	if err != nil {
		t.Fatalf("PutRefundEntry() = %v", err)
	}
	got, gotRev, err := r.GetRefundEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutRefundEntry(RefundEntry{}, 0); err == nil {
		t.Error("PutRefundEntry() accepted an empty entity")
	}
}

func TestRefundItemRoundTrip(t *testing.T) {
	x := SampleRefundItem("refunditem-1", when)
	got, err := ParseRefundItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundItem() = %v", err)
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

func TestRefundItemParseRejectsBadInput(t *testing.T) {
	f := SampleRefundItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundItem(f); err == nil {
		t.Error("ParseRefundItem() accepted a record with no id")
	}
	f = SampleRefundItem("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRefundItem(f); err == nil {
		t.Error("ParseRefundItem() accepted a non-numeric offset")
	}
}

func TestRefundItemCollections(t *testing.T) {
	xs := []RefundItem{SampleRefundItem("a", when), SampleRefundItem("bb", when), SampleRefundItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundItems(xs)); got != 2 {
		t.Errorf("ActiveRefundItems() kept %d, want 2", got)
	}
	if got, want := SumRefundItems(xs), 1+3; got != want {
		t.Errorf("SumRefundItems() = %d, want %d", got, want)
	}
	SortRefundItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundItems() put %q first", xs[0].ID)
	}
	if len(IndexRefundItems(xs)) != 3 {
		t.Errorf("IndexRefundItems() = %v", IndexRefundItems(xs))
	}
	if err := ValidateRefundItems(ActiveRefundItems(xs)); err != nil {
		t.Errorf("ValidateRefundItems() = %v", err)
	}
}

func TestRefundItemNormalizeClamps(t *testing.T) {
	x := SampleRefundItem("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundItem("r1", when)
	rev, err := r.PutRefundItem(x, 0)
	if err != nil {
		t.Fatalf("PutRefundItem() = %v", err)
	}
	got, gotRev, err := r.GetRefundItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundItems() = %v, %v", all, err)
	}
	if _, err := r.PutRefundItem(RefundItem{}, 0); err == nil {
		t.Error("PutRefundItem() accepted an empty entity")
	}
}

func TestRefundDraftRoundTrip(t *testing.T) {
	x := SampleRefundDraft("refunddraft-1", when)
	got, err := ParseRefundDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundDraft() = %v", err)
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

func TestRefundDraftParseRejectsBadInput(t *testing.T) {
	f := SampleRefundDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundDraft(f); err == nil {
		t.Error("ParseRefundDraft() accepted a record with no id")
	}
	f = SampleRefundDraft("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRefundDraft(f); err == nil {
		t.Error("ParseRefundDraft() accepted a non-numeric capacity")
	}
}

func TestRefundDraftCollections(t *testing.T) {
	xs := []RefundDraft{SampleRefundDraft("a", when), SampleRefundDraft("bb", when), SampleRefundDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundDrafts(xs)); got != 2 {
		t.Errorf("ActiveRefundDrafts() kept %d, want 2", got)
	}
	if got, want := SumRefundDrafts(xs), 1+3; got != want {
		t.Errorf("SumRefundDrafts() = %d, want %d", got, want)
	}
	SortRefundDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundDrafts() put %q first", xs[0].ID)
	}
	if len(IndexRefundDrafts(xs)) != 3 {
		t.Errorf("IndexRefundDrafts() = %v", IndexRefundDrafts(xs))
	}
	if err := ValidateRefundDrafts(ActiveRefundDrafts(xs)); err != nil {
		t.Errorf("ValidateRefundDrafts() = %v", err)
	}
}

func TestRefundDraftNormalizeClamps(t *testing.T) {
	x := SampleRefundDraft("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundDraft("r1", when)
	rev, err := r.PutRefundDraft(x, 0)
	if err != nil {
		t.Fatalf("PutRefundDraft() = %v", err)
	}
	got, gotRev, err := r.GetRefundDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutRefundDraft(RefundDraft{}, 0); err == nil {
		t.Error("PutRefundDraft() accepted an empty entity")
	}
}

func TestRefundRevisionRoundTrip(t *testing.T) {
	x := SampleRefundRevision("refundrevision-1", when)
	got, err := ParseRefundRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundRevision() = %v", err)
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

func TestRefundRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleRefundRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundRevision(f); err == nil {
		t.Error("ParseRefundRevision() accepted a record with no id")
	}
	f = SampleRefundRevision("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRefundRevision(f); err == nil {
		t.Error("ParseRefundRevision() accepted a non-numeric retained")
	}
}

func TestRefundRevisionCollections(t *testing.T) {
	xs := []RefundRevision{SampleRefundRevision("a", when), SampleRefundRevision("bb", when), SampleRefundRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundRevisions(xs)); got != 2 {
		t.Errorf("ActiveRefundRevisions() kept %d, want 2", got)
	}
	if got, want := SumRefundRevisions(xs), 1+3; got != want {
		t.Errorf("SumRefundRevisions() = %d, want %d", got, want)
	}
	SortRefundRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundRevisions() put %q first", xs[0].ID)
	}
	if len(IndexRefundRevisions(xs)) != 3 {
		t.Errorf("IndexRefundRevisions() = %v", IndexRefundRevisions(xs))
	}
	if err := ValidateRefundRevisions(ActiveRefundRevisions(xs)); err != nil {
		t.Errorf("ValidateRefundRevisions() = %v", err)
	}
}

func TestRefundRevisionNormalizeClamps(t *testing.T) {
	x := SampleRefundRevision("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundRevision("r1", when)
	rev, err := r.PutRefundRevision(x, 0)
	if err != nil {
		t.Fatalf("PutRefundRevision() = %v", err)
	}
	got, gotRev, err := r.GetRefundRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutRefundRevision(RefundRevision{}, 0); err == nil {
		t.Error("PutRefundRevision() accepted an empty entity")
	}
}

func TestRefundSnapshotRoundTrip(t *testing.T) {
	x := SampleRefundSnapshot("refundsnapshot-1", when)
	got, err := ParseRefundSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundSnapshot() = %v", err)
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

func TestRefundSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleRefundSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundSnapshot(f); err == nil {
		t.Error("ParseRefundSnapshot() accepted a record with no id")
	}
	f = SampleRefundSnapshot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRefundSnapshot(f); err == nil {
		t.Error("ParseRefundSnapshot() accepted a non-numeric offset")
	}
}

func TestRefundSnapshotCollections(t *testing.T) {
	xs := []RefundSnapshot{SampleRefundSnapshot("a", when), SampleRefundSnapshot("bb", when), SampleRefundSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundSnapshots(xs)); got != 2 {
		t.Errorf("ActiveRefundSnapshots() kept %d, want 2", got)
	}
	if got, want := SumRefundSnapshots(xs), 1+3; got != want {
		t.Errorf("SumRefundSnapshots() = %d, want %d", got, want)
	}
	SortRefundSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexRefundSnapshots(xs)) != 3 {
		t.Errorf("IndexRefundSnapshots() = %v", IndexRefundSnapshots(xs))
	}
	if err := ValidateRefundSnapshots(ActiveRefundSnapshots(xs)); err != nil {
		t.Errorf("ValidateRefundSnapshots() = %v", err)
	}
}

func TestRefundSnapshotNormalizeClamps(t *testing.T) {
	x := SampleRefundSnapshot("a", when)
	x.Reference = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundSnapshot("r1", when)
	rev, err := r.PutRefundSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutRefundSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetRefundSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutRefundSnapshot(RefundSnapshot{}, 0); err == nil {
		t.Error("PutRefundSnapshot() accepted an empty entity")
	}
}

func TestRefundBatchRoundTrip(t *testing.T) {
	x := SampleRefundBatch("refundbatch-1", when)
	got, err := ParseRefundBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundBatch() = %v", err)
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

func TestRefundBatchParseRejectsBadInput(t *testing.T) {
	f := SampleRefundBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundBatch(f); err == nil {
		t.Error("ParseRefundBatch() accepted a record with no id")
	}
	f = SampleRefundBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseRefundBatch(f); err == nil {
		t.Error("ParseRefundBatch() accepted a non-numeric version")
	}
}

func TestRefundBatchCollections(t *testing.T) {
	xs := []RefundBatch{SampleRefundBatch("a", when), SampleRefundBatch("bb", when), SampleRefundBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundBatchs(xs)); got != 2 {
		t.Errorf("ActiveRefundBatchs() kept %d, want 2", got)
	}
	if got, want := SumRefundBatchs(xs), 1+3; got != want {
		t.Errorf("SumRefundBatchs() = %d, want %d", got, want)
	}
	SortRefundBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundBatchs() put %q first", xs[0].ID)
	}
	if len(IndexRefundBatchs(xs)) != 3 {
		t.Errorf("IndexRefundBatchs() = %v", IndexRefundBatchs(xs))
	}
	if err := ValidateRefundBatchs(ActiveRefundBatchs(xs)); err != nil {
		t.Errorf("ValidateRefundBatchs() = %v", err)
	}
}

func TestRefundBatchNormalizeClamps(t *testing.T) {
	x := SampleRefundBatch("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundBatch("r1", when)
	rev, err := r.PutRefundBatch(x, 0)
	if err != nil {
		t.Fatalf("PutRefundBatch() = %v", err)
	}
	got, gotRev, err := r.GetRefundBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutRefundBatch(RefundBatch{}, 0); err == nil {
		t.Error("PutRefundBatch() accepted an empty entity")
	}
}

func TestRefundRequestRoundTrip(t *testing.T) {
	x := SampleRefundRequest("refundrequest-1", when)
	got, err := ParseRefundRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundRequest() = %v", err)
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

func TestRefundRequestParseRejectsBadInput(t *testing.T) {
	f := SampleRefundRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundRequest(f); err == nil {
		t.Error("ParseRefundRequest() accepted a record with no id")
	}
	f = SampleRefundRequest("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRefundRequest(f); err == nil {
		t.Error("ParseRefundRequest() accepted a non-numeric capacity")
	}
}

func TestRefundRequestCollections(t *testing.T) {
	xs := []RefundRequest{SampleRefundRequest("a", when), SampleRefundRequest("bb", when), SampleRefundRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundRequests(xs)); got != 2 {
		t.Errorf("ActiveRefundRequests() kept %d, want 2", got)
	}
	if got, want := SumRefundRequests(xs), 1+3; got != want {
		t.Errorf("SumRefundRequests() = %d, want %d", got, want)
	}
	SortRefundRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundRequests() put %q first", xs[0].ID)
	}
	if len(IndexRefundRequests(xs)) != 3 {
		t.Errorf("IndexRefundRequests() = %v", IndexRefundRequests(xs))
	}
	if err := ValidateRefundRequests(ActiveRefundRequests(xs)); err != nil {
		t.Errorf("ValidateRefundRequests() = %v", err)
	}
}

func TestRefundRequestNormalizeClamps(t *testing.T) {
	x := SampleRefundRequest("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundRequest("r1", when)
	rev, err := r.PutRefundRequest(x, 0)
	if err != nil {
		t.Fatalf("PutRefundRequest() = %v", err)
	}
	got, gotRev, err := r.GetRefundRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundRequests() = %v, %v", all, err)
	}
	if _, err := r.PutRefundRequest(RefundRequest{}, 0); err == nil {
		t.Error("PutRefundRequest() accepted an empty entity")
	}
}

func TestRefundSummaryRoundTrip(t *testing.T) {
	x := SampleRefundSummary("refundsummary-1", when)
	got, err := ParseRefundSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundSummary() = %v", err)
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

func TestRefundSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleRefundSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundSummary(f); err == nil {
		t.Error("ParseRefundSummary() accepted a record with no id")
	}
	f = SampleRefundSummary("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRefundSummary(f); err == nil {
		t.Error("ParseRefundSummary() accepted a non-numeric retained")
	}
}

func TestRefundSummaryCollections(t *testing.T) {
	xs := []RefundSummary{SampleRefundSummary("a", when), SampleRefundSummary("bb", when), SampleRefundSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundSummarys(xs)); got != 2 {
		t.Errorf("ActiveRefundSummarys() kept %d, want 2", got)
	}
	if got, want := SumRefundSummarys(xs), 1+3; got != want {
		t.Errorf("SumRefundSummarys() = %d, want %d", got, want)
	}
	SortRefundSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundSummarys() put %q first", xs[0].ID)
	}
	if len(IndexRefundSummarys(xs)) != 3 {
		t.Errorf("IndexRefundSummarys() = %v", IndexRefundSummarys(xs))
	}
	if err := ValidateRefundSummarys(ActiveRefundSummarys(xs)); err != nil {
		t.Errorf("ValidateRefundSummarys() = %v", err)
	}
}

func TestRefundSummaryNormalizeClamps(t *testing.T) {
	x := SampleRefundSummary("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundSummary("r1", when)
	rev, err := r.PutRefundSummary(x, 0)
	if err != nil {
		t.Fatalf("PutRefundSummary() = %v", err)
	}
	got, gotRev, err := r.GetRefundSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutRefundSummary(RefundSummary{}, 0); err == nil {
		t.Error("PutRefundSummary() accepted an empty entity")
	}
}

func TestRefundPolicyRoundTrip(t *testing.T) {
	x := SampleRefundPolicy("refundpolicy-1", when)
	got, err := ParseRefundPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundPolicy() = %v", err)
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

func TestRefundPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleRefundPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundPolicy(f); err == nil {
		t.Error("ParseRefundPolicy() accepted a record with no id")
	}
	f = SampleRefundPolicy("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRefundPolicy(f); err == nil {
		t.Error("ParseRefundPolicy() accepted a non-numeric count")
	}
}

func TestRefundPolicyCollections(t *testing.T) {
	xs := []RefundPolicy{SampleRefundPolicy("a", when), SampleRefundPolicy("bb", when), SampleRefundPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundPolicys(xs)); got != 2 {
		t.Errorf("ActiveRefundPolicys() kept %d, want 2", got)
	}
	if got, want := SumRefundPolicys(xs), 1+3; got != want {
		t.Errorf("SumRefundPolicys() = %d, want %d", got, want)
	}
	SortRefundPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundPolicys() put %q first", xs[0].ID)
	}
	if len(IndexRefundPolicys(xs)) != 3 {
		t.Errorf("IndexRefundPolicys() = %v", IndexRefundPolicys(xs))
	}
	if err := ValidateRefundPolicys(ActiveRefundPolicys(xs)); err != nil {
		t.Errorf("ValidateRefundPolicys() = %v", err)
	}
}

func TestRefundPolicyNormalizeClamps(t *testing.T) {
	x := SampleRefundPolicy("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundPolicy("r1", when)
	rev, err := r.PutRefundPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutRefundPolicy() = %v", err)
	}
	got, gotRev, err := r.GetRefundPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutRefundPolicy(RefundPolicy{}, 0); err == nil {
		t.Error("PutRefundPolicy() accepted an empty entity")
	}
}

func TestRefundLinkRoundTrip(t *testing.T) {
	x := SampleRefundLink("refundlink-1", when)
	got, err := ParseRefundLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundLink() = %v", err)
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

func TestRefundLinkParseRejectsBadInput(t *testing.T) {
	f := SampleRefundLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundLink(f); err == nil {
		t.Error("ParseRefundLink() accepted a record with no id")
	}
	f = SampleRefundLink("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRefundLink(f); err == nil {
		t.Error("ParseRefundLink() accepted a non-numeric weight")
	}
}

func TestRefundLinkCollections(t *testing.T) {
	xs := []RefundLink{SampleRefundLink("a", when), SampleRefundLink("bb", when), SampleRefundLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundLinks(xs)); got != 2 {
		t.Errorf("ActiveRefundLinks() kept %d, want 2", got)
	}
	if got, want := SumRefundLinks(xs), 1+3; got != want {
		t.Errorf("SumRefundLinks() = %d, want %d", got, want)
	}
	SortRefundLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundLinks() put %q first", xs[0].ID)
	}
	if len(IndexRefundLinks(xs)) != 3 {
		t.Errorf("IndexRefundLinks() = %v", IndexRefundLinks(xs))
	}
	if err := ValidateRefundLinks(ActiveRefundLinks(xs)); err != nil {
		t.Errorf("ValidateRefundLinks() = %v", err)
	}
}

func TestRefundLinkNormalizeClamps(t *testing.T) {
	x := SampleRefundLink("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundLink("r1", when)
	rev, err := r.PutRefundLink(x, 0)
	if err != nil {
		t.Fatalf("PutRefundLink() = %v", err)
	}
	got, gotRev, err := r.GetRefundLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundLinks() = %v, %v", all, err)
	}
	if _, err := r.PutRefundLink(RefundLink{}, 0); err == nil {
		t.Error("PutRefundLink() accepted an empty entity")
	}
}

func TestRefundNoteRoundTrip(t *testing.T) {
	x := SampleRefundNote("refundnote-1", when)
	got, err := ParseRefundNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundNote() = %v", err)
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

func TestRefundNoteParseRejectsBadInput(t *testing.T) {
	f := SampleRefundNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundNote(f); err == nil {
		t.Error("ParseRefundNote() accepted a record with no id")
	}
	f = SampleRefundNote("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseRefundNote(f); err == nil {
		t.Error("ParseRefundNote() accepted a non-numeric attempts")
	}
}

func TestRefundNoteCollections(t *testing.T) {
	xs := []RefundNote{SampleRefundNote("a", when), SampleRefundNote("bb", when), SampleRefundNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundNotes(xs)); got != 2 {
		t.Errorf("ActiveRefundNotes() kept %d, want 2", got)
	}
	if got, want := SumRefundNotes(xs), 1+3; got != want {
		t.Errorf("SumRefundNotes() = %d, want %d", got, want)
	}
	SortRefundNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundNotes() put %q first", xs[0].ID)
	}
	if len(IndexRefundNotes(xs)) != 3 {
		t.Errorf("IndexRefundNotes() = %v", IndexRefundNotes(xs))
	}
	if err := ValidateRefundNotes(ActiveRefundNotes(xs)); err != nil {
		t.Errorf("ValidateRefundNotes() = %v", err)
	}
}

func TestRefundNoteNormalizeClamps(t *testing.T) {
	x := SampleRefundNote("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundNote("r1", when)
	rev, err := r.PutRefundNote(x, 0)
	if err != nil {
		t.Fatalf("PutRefundNote() = %v", err)
	}
	got, gotRev, err := r.GetRefundNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundNotes() = %v, %v", all, err)
	}
	if _, err := r.PutRefundNote(RefundNote{}, 0); err == nil {
		t.Error("PutRefundNote() accepted an empty entity")
	}
}

func TestRefundWindowRoundTrip(t *testing.T) {
	x := SampleRefundWindow("refundwindow-1", when)
	got, err := ParseRefundWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundWindow() = %v", err)
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

func TestRefundWindowParseRejectsBadInput(t *testing.T) {
	f := SampleRefundWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundWindow(f); err == nil {
		t.Error("ParseRefundWindow() accepted a record with no id")
	}
	f = SampleRefundWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRefundWindow(f); err == nil {
		t.Error("ParseRefundWindow() accepted a non-numeric priority")
	}
}

func TestRefundWindowCollections(t *testing.T) {
	xs := []RefundWindow{SampleRefundWindow("a", when), SampleRefundWindow("bb", when), SampleRefundWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundWindows(xs)); got != 2 {
		t.Errorf("ActiveRefundWindows() kept %d, want 2", got)
	}
	if got, want := SumRefundWindows(xs), 1+3; got != want {
		t.Errorf("SumRefundWindows() = %d, want %d", got, want)
	}
	SortRefundWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundWindows() put %q first", xs[0].ID)
	}
	if len(IndexRefundWindows(xs)) != 3 {
		t.Errorf("IndexRefundWindows() = %v", IndexRefundWindows(xs))
	}
	if err := ValidateRefundWindows(ActiveRefundWindows(xs)); err != nil {
		t.Errorf("ValidateRefundWindows() = %v", err)
	}
}

func TestRefundWindowNormalizeClamps(t *testing.T) {
	x := SampleRefundWindow("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundWindow("r1", when)
	rev, err := r.PutRefundWindow(x, 0)
	if err != nil {
		t.Fatalf("PutRefundWindow() = %v", err)
	}
	got, gotRev, err := r.GetRefundWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundWindows() = %v, %v", all, err)
	}
	if _, err := r.PutRefundWindow(RefundWindow{}, 0); err == nil {
		t.Error("PutRefundWindow() accepted an empty entity")
	}
}

func TestRefundQuotaRoundTrip(t *testing.T) {
	x := SampleRefundQuota("refundquota-1", when)
	got, err := ParseRefundQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundQuota() = %v", err)
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

func TestRefundQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleRefundQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundQuota(f); err == nil {
		t.Error("ParseRefundQuota() accepted a record with no id")
	}
	f = SampleRefundQuota("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRefundQuota(f); err == nil {
		t.Error("ParseRefundQuota() accepted a non-numeric capacity")
	}
}

func TestRefundQuotaCollections(t *testing.T) {
	xs := []RefundQuota{SampleRefundQuota("a", when), SampleRefundQuota("bb", when), SampleRefundQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundQuotas(xs)); got != 2 {
		t.Errorf("ActiveRefundQuotas() kept %d, want 2", got)
	}
	if got, want := SumRefundQuotas(xs), 1+3; got != want {
		t.Errorf("SumRefundQuotas() = %d, want %d", got, want)
	}
	SortRefundQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundQuotas() put %q first", xs[0].ID)
	}
	if len(IndexRefundQuotas(xs)) != 3 {
		t.Errorf("IndexRefundQuotas() = %v", IndexRefundQuotas(xs))
	}
	if err := ValidateRefundQuotas(ActiveRefundQuotas(xs)); err != nil {
		t.Errorf("ValidateRefundQuotas() = %v", err)
	}
}

func TestRefundQuotaNormalizeClamps(t *testing.T) {
	x := SampleRefundQuota("a", when)
	x.Label = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundQuota("r1", when)
	rev, err := r.PutRefundQuota(x, 0)
	if err != nil {
		t.Fatalf("PutRefundQuota() = %v", err)
	}
	got, gotRev, err := r.GetRefundQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutRefundQuota(RefundQuota{}, 0); err == nil {
		t.Error("PutRefundQuota() accepted an empty entity")
	}
}

func TestRefundBucketRoundTrip(t *testing.T) {
	x := SampleRefundBucket("refundbucket-1", when)
	got, err := ParseRefundBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundBucket() = %v", err)
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

func TestRefundBucketParseRejectsBadInput(t *testing.T) {
	f := SampleRefundBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundBucket(f); err == nil {
		t.Error("ParseRefundBucket() accepted a record with no id")
	}
	f = SampleRefundBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRefundBucket(f); err == nil {
		t.Error("ParseRefundBucket() accepted a non-numeric retained")
	}
}

func TestRefundBucketCollections(t *testing.T) {
	xs := []RefundBucket{SampleRefundBucket("a", when), SampleRefundBucket("bb", when), SampleRefundBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundBuckets(xs)); got != 2 {
		t.Errorf("ActiveRefundBuckets() kept %d, want 2", got)
	}
	if got, want := SumRefundBuckets(xs), 1+3; got != want {
		t.Errorf("SumRefundBuckets() = %d, want %d", got, want)
	}
	SortRefundBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundBuckets() put %q first", xs[0].ID)
	}
	if len(IndexRefundBuckets(xs)) != 3 {
		t.Errorf("IndexRefundBuckets() = %v", IndexRefundBuckets(xs))
	}
	if err := ValidateRefundBuckets(ActiveRefundBuckets(xs)); err != nil {
		t.Errorf("ValidateRefundBuckets() = %v", err)
	}
}

func TestRefundBucketNormalizeClamps(t *testing.T) {
	x := SampleRefundBucket("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundBucket("r1", when)
	rev, err := r.PutRefundBucket(x, 0)
	if err != nil {
		t.Fatalf("PutRefundBucket() = %v", err)
	}
	got, gotRev, err := r.GetRefundBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutRefundBucket(RefundBucket{}, 0); err == nil {
		t.Error("PutRefundBucket() accepted an empty entity")
	}
}

func TestRefundLaneRoundTrip(t *testing.T) {
	x := SampleRefundLane("refundlane-1", when)
	got, err := ParseRefundLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundLane() = %v", err)
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

func TestRefundLaneParseRejectsBadInput(t *testing.T) {
	f := SampleRefundLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundLane(f); err == nil {
		t.Error("ParseRefundLane() accepted a record with no id")
	}
	f = SampleRefundLane("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRefundLane(f); err == nil {
		t.Error("ParseRefundLane() accepted a non-numeric retained")
	}
}

func TestRefundLaneCollections(t *testing.T) {
	xs := []RefundLane{SampleRefundLane("a", when), SampleRefundLane("bb", when), SampleRefundLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundLanes(xs)); got != 2 {
		t.Errorf("ActiveRefundLanes() kept %d, want 2", got)
	}
	if got, want := SumRefundLanes(xs), 1+3; got != want {
		t.Errorf("SumRefundLanes() = %d, want %d", got, want)
	}
	SortRefundLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundLanes() put %q first", xs[0].ID)
	}
	if len(IndexRefundLanes(xs)) != 3 {
		t.Errorf("IndexRefundLanes() = %v", IndexRefundLanes(xs))
	}
	if err := ValidateRefundLanes(ActiveRefundLanes(xs)); err != nil {
		t.Errorf("ValidateRefundLanes() = %v", err)
	}
}

func TestRefundLaneNormalizeClamps(t *testing.T) {
	x := SampleRefundLane("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundLane("r1", when)
	rev, err := r.PutRefundLane(x, 0)
	if err != nil {
		t.Fatalf("PutRefundLane() = %v", err)
	}
	got, gotRev, err := r.GetRefundLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundLanes() = %v, %v", all, err)
	}
	if _, err := r.PutRefundLane(RefundLane{}, 0); err == nil {
		t.Error("PutRefundLane() accepted an empty entity")
	}
}

func TestRefundSlotRoundTrip(t *testing.T) {
	x := SampleRefundSlot("refundslot-1", when)
	got, err := ParseRefundSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundSlot() = %v", err)
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

func TestRefundSlotParseRejectsBadInput(t *testing.T) {
	f := SampleRefundSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundSlot(f); err == nil {
		t.Error("ParseRefundSlot() accepted a record with no id")
	}
	f = SampleRefundSlot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRefundSlot(f); err == nil {
		t.Error("ParseRefundSlot() accepted a non-numeric offset")
	}
}

func TestRefundSlotCollections(t *testing.T) {
	xs := []RefundSlot{SampleRefundSlot("a", when), SampleRefundSlot("bb", when), SampleRefundSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundSlots(xs)); got != 2 {
		t.Errorf("ActiveRefundSlots() kept %d, want 2", got)
	}
	if got, want := SumRefundSlots(xs), 1+3; got != want {
		t.Errorf("SumRefundSlots() = %d, want %d", got, want)
	}
	SortRefundSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundSlots() put %q first", xs[0].ID)
	}
	if len(IndexRefundSlots(xs)) != 3 {
		t.Errorf("IndexRefundSlots() = %v", IndexRefundSlots(xs))
	}
	if err := ValidateRefundSlots(ActiveRefundSlots(xs)); err != nil {
		t.Errorf("ValidateRefundSlots() = %v", err)
	}
}

func TestRefundSlotNormalizeClamps(t *testing.T) {
	x := SampleRefundSlot("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundSlot("r1", when)
	rev, err := r.PutRefundSlot(x, 0)
	if err != nil {
		t.Fatalf("PutRefundSlot() = %v", err)
	}
	got, gotRev, err := r.GetRefundSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundSlots() = %v, %v", all, err)
	}
	if _, err := r.PutRefundSlot(RefundSlot{}, 0); err == nil {
		t.Error("PutRefundSlot() accepted an empty entity")
	}
}

func TestRefundTagRoundTrip(t *testing.T) {
	x := SampleRefundTag("refundtag-1", when)
	got, err := ParseRefundTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundTag() = %v", err)
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

func TestRefundTagParseRejectsBadInput(t *testing.T) {
	f := SampleRefundTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundTag(f); err == nil {
		t.Error("ParseRefundTag() accepted a record with no id")
	}
	f = SampleRefundTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRefundTag(f); err == nil {
		t.Error("ParseRefundTag() accepted a non-numeric count")
	}
}

func TestRefundTagCollections(t *testing.T) {
	xs := []RefundTag{SampleRefundTag("a", when), SampleRefundTag("bb", when), SampleRefundTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundTags(xs)); got != 2 {
		t.Errorf("ActiveRefundTags() kept %d, want 2", got)
	}
	if got, want := SumRefundTags(xs), 1+3; got != want {
		t.Errorf("SumRefundTags() = %d, want %d", got, want)
	}
	SortRefundTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundTags() put %q first", xs[0].ID)
	}
	if len(IndexRefundTags(xs)) != 3 {
		t.Errorf("IndexRefundTags() = %v", IndexRefundTags(xs))
	}
	if err := ValidateRefundTags(ActiveRefundTags(xs)); err != nil {
		t.Errorf("ValidateRefundTags() = %v", err)
	}
}

func TestRefundTagNormalizeClamps(t *testing.T) {
	x := SampleRefundTag("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundTag("r1", when)
	rev, err := r.PutRefundTag(x, 0)
	if err != nil {
		t.Fatalf("PutRefundTag() = %v", err)
	}
	got, gotRev, err := r.GetRefundTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundTags() = %v, %v", all, err)
	}
	if _, err := r.PutRefundTag(RefundTag{}, 0); err == nil {
		t.Error("PutRefundTag() accepted an empty entity")
	}
}

func TestRefundRuleRoundTrip(t *testing.T) {
	x := SampleRefundRule("refundrule-1", when)
	got, err := ParseRefundRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundRule() = %v", err)
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

func TestRefundRuleParseRejectsBadInput(t *testing.T) {
	f := SampleRefundRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundRule(f); err == nil {
		t.Error("ParseRefundRule() accepted a record with no id")
	}
	f = SampleRefundRule("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRefundRule(f); err == nil {
		t.Error("ParseRefundRule() accepted a non-numeric weight")
	}
}

func TestRefundRuleCollections(t *testing.T) {
	xs := []RefundRule{SampleRefundRule("a", when), SampleRefundRule("bb", when), SampleRefundRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundRules(xs)); got != 2 {
		t.Errorf("ActiveRefundRules() kept %d, want 2", got)
	}
	if got, want := SumRefundRules(xs), 1+3; got != want {
		t.Errorf("SumRefundRules() = %d, want %d", got, want)
	}
	SortRefundRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundRules() put %q first", xs[0].ID)
	}
	if len(IndexRefundRules(xs)) != 3 {
		t.Errorf("IndexRefundRules() = %v", IndexRefundRules(xs))
	}
	if err := ValidateRefundRules(ActiveRefundRules(xs)); err != nil {
		t.Errorf("ValidateRefundRules() = %v", err)
	}
}

func TestRefundRuleNormalizeClamps(t *testing.T) {
	x := SampleRefundRule("a", when)
	x.Group = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundRule("r1", when)
	rev, err := r.PutRefundRule(x, 0)
	if err != nil {
		t.Fatalf("PutRefundRule() = %v", err)
	}
	got, gotRev, err := r.GetRefundRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundRules() = %v, %v", all, err)
	}
	if _, err := r.PutRefundRule(RefundRule{}, 0); err == nil {
		t.Error("PutRefundRule() accepted an empty entity")
	}
}

func TestRefundHoldRoundTrip(t *testing.T) {
	x := SampleRefundHold("refundhold-1", when)
	got, err := ParseRefundHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundHold() = %v", err)
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

func TestRefundHoldParseRejectsBadInput(t *testing.T) {
	f := SampleRefundHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundHold(f); err == nil {
		t.Error("ParseRefundHold() accepted a record with no id")
	}
	f = SampleRefundHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRefundHold(f); err == nil {
		t.Error("ParseRefundHold() accepted a non-numeric weight")
	}
}

func TestRefundHoldCollections(t *testing.T) {
	xs := []RefundHold{SampleRefundHold("a", when), SampleRefundHold("bb", when), SampleRefundHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundHolds(xs)); got != 2 {
		t.Errorf("ActiveRefundHolds() kept %d, want 2", got)
	}
	if got, want := SumRefundHolds(xs), 1+3; got != want {
		t.Errorf("SumRefundHolds() = %d, want %d", got, want)
	}
	SortRefundHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundHolds() put %q first", xs[0].ID)
	}
	if len(IndexRefundHolds(xs)) != 3 {
		t.Errorf("IndexRefundHolds() = %v", IndexRefundHolds(xs))
	}
	if err := ValidateRefundHolds(ActiveRefundHolds(xs)); err != nil {
		t.Errorf("ValidateRefundHolds() = %v", err)
	}
}

func TestRefundHoldNormalizeClamps(t *testing.T) {
	x := SampleRefundHold("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundHold("r1", when)
	rev, err := r.PutRefundHold(x, 0)
	if err != nil {
		t.Fatalf("PutRefundHold() = %v", err)
	}
	got, gotRev, err := r.GetRefundHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundHolds() = %v, %v", all, err)
	}
	if _, err := r.PutRefundHold(RefundHold{}, 0); err == nil {
		t.Error("PutRefundHold() accepted an empty entity")
	}
}

func TestRefundClaimRoundTrip(t *testing.T) {
	x := SampleRefundClaim("refundclaim-1", when)
	got, err := ParseRefundClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundClaim() = %v", err)
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

func TestRefundClaimParseRejectsBadInput(t *testing.T) {
	f := SampleRefundClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundClaim(f); err == nil {
		t.Error("ParseRefundClaim() accepted a record with no id")
	}
	f = SampleRefundClaim("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRefundClaim(f); err == nil {
		t.Error("ParseRefundClaim() accepted a non-numeric count")
	}
}

func TestRefundClaimCollections(t *testing.T) {
	xs := []RefundClaim{SampleRefundClaim("a", when), SampleRefundClaim("bb", when), SampleRefundClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundClaims(xs)); got != 2 {
		t.Errorf("ActiveRefundClaims() kept %d, want 2", got)
	}
	if got, want := SumRefundClaims(xs), 1+3; got != want {
		t.Errorf("SumRefundClaims() = %d, want %d", got, want)
	}
	SortRefundClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundClaims() put %q first", xs[0].ID)
	}
	if len(IndexRefundClaims(xs)) != 3 {
		t.Errorf("IndexRefundClaims() = %v", IndexRefundClaims(xs))
	}
	if err := ValidateRefundClaims(ActiveRefundClaims(xs)); err != nil {
		t.Errorf("ValidateRefundClaims() = %v", err)
	}
}

func TestRefundClaimNormalizeClamps(t *testing.T) {
	x := SampleRefundClaim("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundClaim("r1", when)
	rev, err := r.PutRefundClaim(x, 0)
	if err != nil {
		t.Fatalf("PutRefundClaim() = %v", err)
	}
	got, gotRev, err := r.GetRefundClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundClaims() = %v, %v", all, err)
	}
	if _, err := r.PutRefundClaim(RefundClaim{}, 0); err == nil {
		t.Error("PutRefundClaim() accepted an empty entity")
	}
}

func TestRefundCursorRoundTrip(t *testing.T) {
	x := SampleRefundCursor("refundcursor-1", when)
	got, err := ParseRefundCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseRefundCursor() = %v", err)
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

func TestRefundCursorParseRejectsBadInput(t *testing.T) {
	f := SampleRefundCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRefundCursor(f); err == nil {
		t.Error("ParseRefundCursor() accepted a record with no id")
	}
	f = SampleRefundCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRefundCursor(f); err == nil {
		t.Error("ParseRefundCursor() accepted a non-numeric weight")
	}
}

func TestRefundCursorCollections(t *testing.T) {
	xs := []RefundCursor{SampleRefundCursor("a", when), SampleRefundCursor("bb", when), SampleRefundCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRefundCursors(xs)); got != 2 {
		t.Errorf("ActiveRefundCursors() kept %d, want 2", got)
	}
	if got, want := SumRefundCursors(xs), 1+3; got != want {
		t.Errorf("SumRefundCursors() = %d, want %d", got, want)
	}
	SortRefundCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRefundCursors() put %q first", xs[0].ID)
	}
	if len(IndexRefundCursors(xs)) != 3 {
		t.Errorf("IndexRefundCursors() = %v", IndexRefundCursors(xs))
	}
	if err := ValidateRefundCursors(ActiveRefundCursors(xs)); err != nil {
		t.Errorf("ValidateRefundCursors() = %v", err)
	}
}

func TestRefundCursorNormalizeClamps(t *testing.T) {
	x := SampleRefundCursor("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRefundCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleRefundCursor("r1", when)
	rev, err := r.PutRefundCursor(x, 0)
	if err != nil {
		t.Fatalf("PutRefundCursor() = %v", err)
	}
	got, gotRev, err := r.GetRefundCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetRefundCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListRefundCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListRefundCursors() = %v, %v", all, err)
	}
	if _, err := r.PutRefundCursor(RefundCursor{}, 0); err == nil {
		t.Error("PutRefundCursor() accepted an empty entity")
	}
}
