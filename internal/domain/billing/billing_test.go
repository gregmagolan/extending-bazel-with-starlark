package billing

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
		if !strings.HasPrefix(Prefixes[i], "billing/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleBillingRecord("ok", when).Fields()); err != nil {
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

func TestBillingRecordRoundTrip(t *testing.T) {
	x := SampleBillingRecord("billingrecord-1", when)
	got, err := ParseBillingRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingRecord() = %v", err)
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

func TestBillingRecordParseRejectsBadInput(t *testing.T) {
	f := SampleBillingRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingRecord(f); err == nil {
		t.Error("ParseBillingRecord() accepted a record with no id")
	}
	f = SampleBillingRecord("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseBillingRecord(f); err == nil {
		t.Error("ParseBillingRecord() accepted a non-numeric count")
	}
}

func TestBillingRecordCollections(t *testing.T) {
	xs := []BillingRecord{SampleBillingRecord("a", when), SampleBillingRecord("bb", when), SampleBillingRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingRecords(xs)); got != 2 {
		t.Errorf("ActiveBillingRecords() kept %d, want 2", got)
	}
	if got, want := SumBillingRecords(xs), 1+3; got != want {
		t.Errorf("SumBillingRecords() = %d, want %d", got, want)
	}
	SortBillingRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingRecords() put %q first", xs[0].ID)
	}
	if len(IndexBillingRecords(xs)) != 3 {
		t.Errorf("IndexBillingRecords() = %v", IndexBillingRecords(xs))
	}
	if err := ValidateBillingRecords(ActiveBillingRecords(xs)); err != nil {
		t.Errorf("ValidateBillingRecords() = %v", err)
	}
}

func TestBillingRecordNormalizeClamps(t *testing.T) {
	x := SampleBillingRecord("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingRecord("r1", when)
	rev, err := r.PutBillingRecord(x, 0)
	if err != nil {
		t.Fatalf("PutBillingRecord() = %v", err)
	}
	got, gotRev, err := r.GetBillingRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingRecords() = %v, %v", all, err)
	}
	if _, err := r.PutBillingRecord(BillingRecord{}, 0); err == nil {
		t.Error("PutBillingRecord() accepted an empty entity")
	}
}

func TestBillingEntryRoundTrip(t *testing.T) {
	x := SampleBillingEntry("billingentry-1", when)
	got, err := ParseBillingEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingEntry() = %v", err)
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

func TestBillingEntryParseRejectsBadInput(t *testing.T) {
	f := SampleBillingEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingEntry(f); err == nil {
		t.Error("ParseBillingEntry() accepted a record with no id")
	}
	f = SampleBillingEntry("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseBillingEntry(f); err == nil {
		t.Error("ParseBillingEntry() accepted a non-numeric version")
	}
}

func TestBillingEntryCollections(t *testing.T) {
	xs := []BillingEntry{SampleBillingEntry("a", when), SampleBillingEntry("bb", when), SampleBillingEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingEntrys(xs)); got != 2 {
		t.Errorf("ActiveBillingEntrys() kept %d, want 2", got)
	}
	if got, want := SumBillingEntrys(xs), 1+3; got != want {
		t.Errorf("SumBillingEntrys() = %d, want %d", got, want)
	}
	SortBillingEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingEntrys() put %q first", xs[0].ID)
	}
	if len(IndexBillingEntrys(xs)) != 3 {
		t.Errorf("IndexBillingEntrys() = %v", IndexBillingEntrys(xs))
	}
	if err := ValidateBillingEntrys(ActiveBillingEntrys(xs)); err != nil {
		t.Errorf("ValidateBillingEntrys() = %v", err)
	}
}

func TestBillingEntryNormalizeClamps(t *testing.T) {
	x := SampleBillingEntry("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingEntry("r1", when)
	rev, err := r.PutBillingEntry(x, 0)
	if err != nil {
		t.Fatalf("PutBillingEntry() = %v", err)
	}
	got, gotRev, err := r.GetBillingEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutBillingEntry(BillingEntry{}, 0); err == nil {
		t.Error("PutBillingEntry() accepted an empty entity")
	}
}

func TestBillingItemRoundTrip(t *testing.T) {
	x := SampleBillingItem("billingitem-1", when)
	got, err := ParseBillingItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingItem() = %v", err)
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

func TestBillingItemParseRejectsBadInput(t *testing.T) {
	f := SampleBillingItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingItem(f); err == nil {
		t.Error("ParseBillingItem() accepted a record with no id")
	}
	f = SampleBillingItem("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseBillingItem(f); err == nil {
		t.Error("ParseBillingItem() accepted a non-numeric retained")
	}
}

func TestBillingItemCollections(t *testing.T) {
	xs := []BillingItem{SampleBillingItem("a", when), SampleBillingItem("bb", when), SampleBillingItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingItems(xs)); got != 2 {
		t.Errorf("ActiveBillingItems() kept %d, want 2", got)
	}
	if got, want := SumBillingItems(xs), 1+3; got != want {
		t.Errorf("SumBillingItems() = %d, want %d", got, want)
	}
	SortBillingItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingItems() put %q first", xs[0].ID)
	}
	if len(IndexBillingItems(xs)) != 3 {
		t.Errorf("IndexBillingItems() = %v", IndexBillingItems(xs))
	}
	if err := ValidateBillingItems(ActiveBillingItems(xs)); err != nil {
		t.Errorf("ValidateBillingItems() = %v", err)
	}
}

func TestBillingItemNormalizeClamps(t *testing.T) {
	x := SampleBillingItem("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingItem("r1", when)
	rev, err := r.PutBillingItem(x, 0)
	if err != nil {
		t.Fatalf("PutBillingItem() = %v", err)
	}
	got, gotRev, err := r.GetBillingItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingItems() = %v, %v", all, err)
	}
	if _, err := r.PutBillingItem(BillingItem{}, 0); err == nil {
		t.Error("PutBillingItem() accepted an empty entity")
	}
}

func TestBillingDraftRoundTrip(t *testing.T) {
	x := SampleBillingDraft("billingdraft-1", when)
	got, err := ParseBillingDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingDraft() = %v", err)
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

func TestBillingDraftParseRejectsBadInput(t *testing.T) {
	f := SampleBillingDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingDraft(f); err == nil {
		t.Error("ParseBillingDraft() accepted a record with no id")
	}
	f = SampleBillingDraft("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseBillingDraft(f); err == nil {
		t.Error("ParseBillingDraft() accepted a non-numeric version")
	}
}

func TestBillingDraftCollections(t *testing.T) {
	xs := []BillingDraft{SampleBillingDraft("a", when), SampleBillingDraft("bb", when), SampleBillingDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingDrafts(xs)); got != 2 {
		t.Errorf("ActiveBillingDrafts() kept %d, want 2", got)
	}
	if got, want := SumBillingDrafts(xs), 1+3; got != want {
		t.Errorf("SumBillingDrafts() = %d, want %d", got, want)
	}
	SortBillingDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingDrafts() put %q first", xs[0].ID)
	}
	if len(IndexBillingDrafts(xs)) != 3 {
		t.Errorf("IndexBillingDrafts() = %v", IndexBillingDrafts(xs))
	}
	if err := ValidateBillingDrafts(ActiveBillingDrafts(xs)); err != nil {
		t.Errorf("ValidateBillingDrafts() = %v", err)
	}
}

func TestBillingDraftNormalizeClamps(t *testing.T) {
	x := SampleBillingDraft("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingDraft("r1", when)
	rev, err := r.PutBillingDraft(x, 0)
	if err != nil {
		t.Fatalf("PutBillingDraft() = %v", err)
	}
	got, gotRev, err := r.GetBillingDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutBillingDraft(BillingDraft{}, 0); err == nil {
		t.Error("PutBillingDraft() accepted an empty entity")
	}
}

func TestBillingRevisionRoundTrip(t *testing.T) {
	x := SampleBillingRevision("billingrevision-1", when)
	got, err := ParseBillingRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingRevision() = %v", err)
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

func TestBillingRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleBillingRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingRevision(f); err == nil {
		t.Error("ParseBillingRevision() accepted a record with no id")
	}
	f = SampleBillingRevision("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseBillingRevision(f); err == nil {
		t.Error("ParseBillingRevision() accepted a non-numeric capacity")
	}
}

func TestBillingRevisionCollections(t *testing.T) {
	xs := []BillingRevision{SampleBillingRevision("a", when), SampleBillingRevision("bb", when), SampleBillingRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingRevisions(xs)); got != 2 {
		t.Errorf("ActiveBillingRevisions() kept %d, want 2", got)
	}
	if got, want := SumBillingRevisions(xs), 1+3; got != want {
		t.Errorf("SumBillingRevisions() = %d, want %d", got, want)
	}
	SortBillingRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingRevisions() put %q first", xs[0].ID)
	}
	if len(IndexBillingRevisions(xs)) != 3 {
		t.Errorf("IndexBillingRevisions() = %v", IndexBillingRevisions(xs))
	}
	if err := ValidateBillingRevisions(ActiveBillingRevisions(xs)); err != nil {
		t.Errorf("ValidateBillingRevisions() = %v", err)
	}
}

func TestBillingRevisionNormalizeClamps(t *testing.T) {
	x := SampleBillingRevision("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingRevision("r1", when)
	rev, err := r.PutBillingRevision(x, 0)
	if err != nil {
		t.Fatalf("PutBillingRevision() = %v", err)
	}
	got, gotRev, err := r.GetBillingRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutBillingRevision(BillingRevision{}, 0); err == nil {
		t.Error("PutBillingRevision() accepted an empty entity")
	}
}

func TestBillingSnapshotRoundTrip(t *testing.T) {
	x := SampleBillingSnapshot("billingsnapshot-1", when)
	got, err := ParseBillingSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingSnapshot() = %v", err)
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

func TestBillingSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleBillingSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingSnapshot(f); err == nil {
		t.Error("ParseBillingSnapshot() accepted a record with no id")
	}
	f = SampleBillingSnapshot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseBillingSnapshot(f); err == nil {
		t.Error("ParseBillingSnapshot() accepted a non-numeric retained")
	}
}

func TestBillingSnapshotCollections(t *testing.T) {
	xs := []BillingSnapshot{SampleBillingSnapshot("a", when), SampleBillingSnapshot("bb", when), SampleBillingSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingSnapshots(xs)); got != 2 {
		t.Errorf("ActiveBillingSnapshots() kept %d, want 2", got)
	}
	if got, want := SumBillingSnapshots(xs), 1+3; got != want {
		t.Errorf("SumBillingSnapshots() = %d, want %d", got, want)
	}
	SortBillingSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexBillingSnapshots(xs)) != 3 {
		t.Errorf("IndexBillingSnapshots() = %v", IndexBillingSnapshots(xs))
	}
	if err := ValidateBillingSnapshots(ActiveBillingSnapshots(xs)); err != nil {
		t.Errorf("ValidateBillingSnapshots() = %v", err)
	}
}

func TestBillingSnapshotNormalizeClamps(t *testing.T) {
	x := SampleBillingSnapshot("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingSnapshot("r1", when)
	rev, err := r.PutBillingSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutBillingSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetBillingSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutBillingSnapshot(BillingSnapshot{}, 0); err == nil {
		t.Error("PutBillingSnapshot() accepted an empty entity")
	}
}

func TestBillingBatchRoundTrip(t *testing.T) {
	x := SampleBillingBatch("billingbatch-1", when)
	got, err := ParseBillingBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingBatch() = %v", err)
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

func TestBillingBatchParseRejectsBadInput(t *testing.T) {
	f := SampleBillingBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingBatch(f); err == nil {
		t.Error("ParseBillingBatch() accepted a record with no id")
	}
	f = SampleBillingBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseBillingBatch(f); err == nil {
		t.Error("ParseBillingBatch() accepted a non-numeric version")
	}
}

func TestBillingBatchCollections(t *testing.T) {
	xs := []BillingBatch{SampleBillingBatch("a", when), SampleBillingBatch("bb", when), SampleBillingBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingBatchs(xs)); got != 2 {
		t.Errorf("ActiveBillingBatchs() kept %d, want 2", got)
	}
	if got, want := SumBillingBatchs(xs), 1+3; got != want {
		t.Errorf("SumBillingBatchs() = %d, want %d", got, want)
	}
	SortBillingBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingBatchs() put %q first", xs[0].ID)
	}
	if len(IndexBillingBatchs(xs)) != 3 {
		t.Errorf("IndexBillingBatchs() = %v", IndexBillingBatchs(xs))
	}
	if err := ValidateBillingBatchs(ActiveBillingBatchs(xs)); err != nil {
		t.Errorf("ValidateBillingBatchs() = %v", err)
	}
}

func TestBillingBatchNormalizeClamps(t *testing.T) {
	x := SampleBillingBatch("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingBatch("r1", when)
	rev, err := r.PutBillingBatch(x, 0)
	if err != nil {
		t.Fatalf("PutBillingBatch() = %v", err)
	}
	got, gotRev, err := r.GetBillingBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutBillingBatch(BillingBatch{}, 0); err == nil {
		t.Error("PutBillingBatch() accepted an empty entity")
	}
}

func TestBillingRequestRoundTrip(t *testing.T) {
	x := SampleBillingRequest("billingrequest-1", when)
	got, err := ParseBillingRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingRequest() = %v", err)
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

func TestBillingRequestParseRejectsBadInput(t *testing.T) {
	f := SampleBillingRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingRequest(f); err == nil {
		t.Error("ParseBillingRequest() accepted a record with no id")
	}
	f = SampleBillingRequest("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseBillingRequest(f); err == nil {
		t.Error("ParseBillingRequest() accepted a non-numeric weight")
	}
}

func TestBillingRequestCollections(t *testing.T) {
	xs := []BillingRequest{SampleBillingRequest("a", when), SampleBillingRequest("bb", when), SampleBillingRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingRequests(xs)); got != 2 {
		t.Errorf("ActiveBillingRequests() kept %d, want 2", got)
	}
	if got, want := SumBillingRequests(xs), 1+3; got != want {
		t.Errorf("SumBillingRequests() = %d, want %d", got, want)
	}
	SortBillingRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingRequests() put %q first", xs[0].ID)
	}
	if len(IndexBillingRequests(xs)) != 3 {
		t.Errorf("IndexBillingRequests() = %v", IndexBillingRequests(xs))
	}
	if err := ValidateBillingRequests(ActiveBillingRequests(xs)); err != nil {
		t.Errorf("ValidateBillingRequests() = %v", err)
	}
}

func TestBillingRequestNormalizeClamps(t *testing.T) {
	x := SampleBillingRequest("a", when)
	x.Group = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingRequest("r1", when)
	rev, err := r.PutBillingRequest(x, 0)
	if err != nil {
		t.Fatalf("PutBillingRequest() = %v", err)
	}
	got, gotRev, err := r.GetBillingRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingRequests() = %v, %v", all, err)
	}
	if _, err := r.PutBillingRequest(BillingRequest{}, 0); err == nil {
		t.Error("PutBillingRequest() accepted an empty entity")
	}
}

func TestBillingSummaryRoundTrip(t *testing.T) {
	x := SampleBillingSummary("billingsummary-1", when)
	got, err := ParseBillingSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingSummary() = %v", err)
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

func TestBillingSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleBillingSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingSummary(f); err == nil {
		t.Error("ParseBillingSummary() accepted a record with no id")
	}
	f = SampleBillingSummary("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseBillingSummary(f); err == nil {
		t.Error("ParseBillingSummary() accepted a non-numeric retained")
	}
}

func TestBillingSummaryCollections(t *testing.T) {
	xs := []BillingSummary{SampleBillingSummary("a", when), SampleBillingSummary("bb", when), SampleBillingSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingSummarys(xs)); got != 2 {
		t.Errorf("ActiveBillingSummarys() kept %d, want 2", got)
	}
	if got, want := SumBillingSummarys(xs), 1+3; got != want {
		t.Errorf("SumBillingSummarys() = %d, want %d", got, want)
	}
	SortBillingSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingSummarys() put %q first", xs[0].ID)
	}
	if len(IndexBillingSummarys(xs)) != 3 {
		t.Errorf("IndexBillingSummarys() = %v", IndexBillingSummarys(xs))
	}
	if err := ValidateBillingSummarys(ActiveBillingSummarys(xs)); err != nil {
		t.Errorf("ValidateBillingSummarys() = %v", err)
	}
}

func TestBillingSummaryNormalizeClamps(t *testing.T) {
	x := SampleBillingSummary("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingSummary("r1", when)
	rev, err := r.PutBillingSummary(x, 0)
	if err != nil {
		t.Fatalf("PutBillingSummary() = %v", err)
	}
	got, gotRev, err := r.GetBillingSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutBillingSummary(BillingSummary{}, 0); err == nil {
		t.Error("PutBillingSummary() accepted an empty entity")
	}
}

func TestBillingPolicyRoundTrip(t *testing.T) {
	x := SampleBillingPolicy("billingpolicy-1", when)
	got, err := ParseBillingPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingPolicy() = %v", err)
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

func TestBillingPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleBillingPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingPolicy(f); err == nil {
		t.Error("ParseBillingPolicy() accepted a record with no id")
	}
	f = SampleBillingPolicy("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseBillingPolicy(f); err == nil {
		t.Error("ParseBillingPolicy() accepted a non-numeric priority")
	}
}

func TestBillingPolicyCollections(t *testing.T) {
	xs := []BillingPolicy{SampleBillingPolicy("a", when), SampleBillingPolicy("bb", when), SampleBillingPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingPolicys(xs)); got != 2 {
		t.Errorf("ActiveBillingPolicys() kept %d, want 2", got)
	}
	if got, want := SumBillingPolicys(xs), 1+3; got != want {
		t.Errorf("SumBillingPolicys() = %d, want %d", got, want)
	}
	SortBillingPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingPolicys() put %q first", xs[0].ID)
	}
	if len(IndexBillingPolicys(xs)) != 3 {
		t.Errorf("IndexBillingPolicys() = %v", IndexBillingPolicys(xs))
	}
	if err := ValidateBillingPolicys(ActiveBillingPolicys(xs)); err != nil {
		t.Errorf("ValidateBillingPolicys() = %v", err)
	}
}

func TestBillingPolicyNormalizeClamps(t *testing.T) {
	x := SampleBillingPolicy("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingPolicy("r1", when)
	rev, err := r.PutBillingPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutBillingPolicy() = %v", err)
	}
	got, gotRev, err := r.GetBillingPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutBillingPolicy(BillingPolicy{}, 0); err == nil {
		t.Error("PutBillingPolicy() accepted an empty entity")
	}
}

func TestBillingLinkRoundTrip(t *testing.T) {
	x := SampleBillingLink("billinglink-1", when)
	got, err := ParseBillingLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingLink() = %v", err)
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

func TestBillingLinkParseRejectsBadInput(t *testing.T) {
	f := SampleBillingLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingLink(f); err == nil {
		t.Error("ParseBillingLink() accepted a record with no id")
	}
	f = SampleBillingLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseBillingLink(f); err == nil {
		t.Error("ParseBillingLink() accepted a non-numeric retained")
	}
}

func TestBillingLinkCollections(t *testing.T) {
	xs := []BillingLink{SampleBillingLink("a", when), SampleBillingLink("bb", when), SampleBillingLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingLinks(xs)); got != 2 {
		t.Errorf("ActiveBillingLinks() kept %d, want 2", got)
	}
	if got, want := SumBillingLinks(xs), 1+3; got != want {
		t.Errorf("SumBillingLinks() = %d, want %d", got, want)
	}
	SortBillingLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingLinks() put %q first", xs[0].ID)
	}
	if len(IndexBillingLinks(xs)) != 3 {
		t.Errorf("IndexBillingLinks() = %v", IndexBillingLinks(xs))
	}
	if err := ValidateBillingLinks(ActiveBillingLinks(xs)); err != nil {
		t.Errorf("ValidateBillingLinks() = %v", err)
	}
}

func TestBillingLinkNormalizeClamps(t *testing.T) {
	x := SampleBillingLink("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingLink("r1", when)
	rev, err := r.PutBillingLink(x, 0)
	if err != nil {
		t.Fatalf("PutBillingLink() = %v", err)
	}
	got, gotRev, err := r.GetBillingLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingLinks() = %v, %v", all, err)
	}
	if _, err := r.PutBillingLink(BillingLink{}, 0); err == nil {
		t.Error("PutBillingLink() accepted an empty entity")
	}
}

func TestBillingNoteRoundTrip(t *testing.T) {
	x := SampleBillingNote("billingnote-1", when)
	got, err := ParseBillingNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingNote() = %v", err)
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

func TestBillingNoteParseRejectsBadInput(t *testing.T) {
	f := SampleBillingNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingNote(f); err == nil {
		t.Error("ParseBillingNote() accepted a record with no id")
	}
	f = SampleBillingNote("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseBillingNote(f); err == nil {
		t.Error("ParseBillingNote() accepted a non-numeric offset")
	}
}

func TestBillingNoteCollections(t *testing.T) {
	xs := []BillingNote{SampleBillingNote("a", when), SampleBillingNote("bb", when), SampleBillingNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingNotes(xs)); got != 2 {
		t.Errorf("ActiveBillingNotes() kept %d, want 2", got)
	}
	if got, want := SumBillingNotes(xs), 1+3; got != want {
		t.Errorf("SumBillingNotes() = %d, want %d", got, want)
	}
	SortBillingNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingNotes() put %q first", xs[0].ID)
	}
	if len(IndexBillingNotes(xs)) != 3 {
		t.Errorf("IndexBillingNotes() = %v", IndexBillingNotes(xs))
	}
	if err := ValidateBillingNotes(ActiveBillingNotes(xs)); err != nil {
		t.Errorf("ValidateBillingNotes() = %v", err)
	}
}

func TestBillingNoteNormalizeClamps(t *testing.T) {
	x := SampleBillingNote("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingNote("r1", when)
	rev, err := r.PutBillingNote(x, 0)
	if err != nil {
		t.Fatalf("PutBillingNote() = %v", err)
	}
	got, gotRev, err := r.GetBillingNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingNotes() = %v, %v", all, err)
	}
	if _, err := r.PutBillingNote(BillingNote{}, 0); err == nil {
		t.Error("PutBillingNote() accepted an empty entity")
	}
}

func TestBillingWindowRoundTrip(t *testing.T) {
	x := SampleBillingWindow("billingwindow-1", when)
	got, err := ParseBillingWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingWindow() = %v", err)
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

func TestBillingWindowParseRejectsBadInput(t *testing.T) {
	f := SampleBillingWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingWindow(f); err == nil {
		t.Error("ParseBillingWindow() accepted a record with no id")
	}
	f = SampleBillingWindow("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseBillingWindow(f); err == nil {
		t.Error("ParseBillingWindow() accepted a non-numeric offset")
	}
}

func TestBillingWindowCollections(t *testing.T) {
	xs := []BillingWindow{SampleBillingWindow("a", when), SampleBillingWindow("bb", when), SampleBillingWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingWindows(xs)); got != 2 {
		t.Errorf("ActiveBillingWindows() kept %d, want 2", got)
	}
	if got, want := SumBillingWindows(xs), 1+3; got != want {
		t.Errorf("SumBillingWindows() = %d, want %d", got, want)
	}
	SortBillingWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingWindows() put %q first", xs[0].ID)
	}
	if len(IndexBillingWindows(xs)) != 3 {
		t.Errorf("IndexBillingWindows() = %v", IndexBillingWindows(xs))
	}
	if err := ValidateBillingWindows(ActiveBillingWindows(xs)); err != nil {
		t.Errorf("ValidateBillingWindows() = %v", err)
	}
}

func TestBillingWindowNormalizeClamps(t *testing.T) {
	x := SampleBillingWindow("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingWindow("r1", when)
	rev, err := r.PutBillingWindow(x, 0)
	if err != nil {
		t.Fatalf("PutBillingWindow() = %v", err)
	}
	got, gotRev, err := r.GetBillingWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingWindows() = %v, %v", all, err)
	}
	if _, err := r.PutBillingWindow(BillingWindow{}, 0); err == nil {
		t.Error("PutBillingWindow() accepted an empty entity")
	}
}

func TestBillingQuotaRoundTrip(t *testing.T) {
	x := SampleBillingQuota("billingquota-1", when)
	got, err := ParseBillingQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingQuota() = %v", err)
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

func TestBillingQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleBillingQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingQuota(f); err == nil {
		t.Error("ParseBillingQuota() accepted a record with no id")
	}
	f = SampleBillingQuota("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseBillingQuota(f); err == nil {
		t.Error("ParseBillingQuota() accepted a non-numeric version")
	}
}

func TestBillingQuotaCollections(t *testing.T) {
	xs := []BillingQuota{SampleBillingQuota("a", when), SampleBillingQuota("bb", when), SampleBillingQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingQuotas(xs)); got != 2 {
		t.Errorf("ActiveBillingQuotas() kept %d, want 2", got)
	}
	if got, want := SumBillingQuotas(xs), 1+3; got != want {
		t.Errorf("SumBillingQuotas() = %d, want %d", got, want)
	}
	SortBillingQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingQuotas() put %q first", xs[0].ID)
	}
	if len(IndexBillingQuotas(xs)) != 3 {
		t.Errorf("IndexBillingQuotas() = %v", IndexBillingQuotas(xs))
	}
	if err := ValidateBillingQuotas(ActiveBillingQuotas(xs)); err != nil {
		t.Errorf("ValidateBillingQuotas() = %v", err)
	}
}

func TestBillingQuotaNormalizeClamps(t *testing.T) {
	x := SampleBillingQuota("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingQuota("r1", when)
	rev, err := r.PutBillingQuota(x, 0)
	if err != nil {
		t.Fatalf("PutBillingQuota() = %v", err)
	}
	got, gotRev, err := r.GetBillingQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutBillingQuota(BillingQuota{}, 0); err == nil {
		t.Error("PutBillingQuota() accepted an empty entity")
	}
}

func TestBillingBucketRoundTrip(t *testing.T) {
	x := SampleBillingBucket("billingbucket-1", when)
	got, err := ParseBillingBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingBucket() = %v", err)
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

func TestBillingBucketParseRejectsBadInput(t *testing.T) {
	f := SampleBillingBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingBucket(f); err == nil {
		t.Error("ParseBillingBucket() accepted a record with no id")
	}
	f = SampleBillingBucket("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseBillingBucket(f); err == nil {
		t.Error("ParseBillingBucket() accepted a non-numeric attempts")
	}
}

func TestBillingBucketCollections(t *testing.T) {
	xs := []BillingBucket{SampleBillingBucket("a", when), SampleBillingBucket("bb", when), SampleBillingBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingBuckets(xs)); got != 2 {
		t.Errorf("ActiveBillingBuckets() kept %d, want 2", got)
	}
	if got, want := SumBillingBuckets(xs), 1+3; got != want {
		t.Errorf("SumBillingBuckets() = %d, want %d", got, want)
	}
	SortBillingBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingBuckets() put %q first", xs[0].ID)
	}
	if len(IndexBillingBuckets(xs)) != 3 {
		t.Errorf("IndexBillingBuckets() = %v", IndexBillingBuckets(xs))
	}
	if err := ValidateBillingBuckets(ActiveBillingBuckets(xs)); err != nil {
		t.Errorf("ValidateBillingBuckets() = %v", err)
	}
}

func TestBillingBucketNormalizeClamps(t *testing.T) {
	x := SampleBillingBucket("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingBucket("r1", when)
	rev, err := r.PutBillingBucket(x, 0)
	if err != nil {
		t.Fatalf("PutBillingBucket() = %v", err)
	}
	got, gotRev, err := r.GetBillingBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutBillingBucket(BillingBucket{}, 0); err == nil {
		t.Error("PutBillingBucket() accepted an empty entity")
	}
}

func TestBillingLaneRoundTrip(t *testing.T) {
	x := SampleBillingLane("billinglane-1", when)
	got, err := ParseBillingLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingLane() = %v", err)
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

func TestBillingLaneParseRejectsBadInput(t *testing.T) {
	f := SampleBillingLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingLane(f); err == nil {
		t.Error("ParseBillingLane() accepted a record with no id")
	}
	f = SampleBillingLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseBillingLane(f); err == nil {
		t.Error("ParseBillingLane() accepted a non-numeric priority")
	}
}

func TestBillingLaneCollections(t *testing.T) {
	xs := []BillingLane{SampleBillingLane("a", when), SampleBillingLane("bb", when), SampleBillingLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingLanes(xs)); got != 2 {
		t.Errorf("ActiveBillingLanes() kept %d, want 2", got)
	}
	if got, want := SumBillingLanes(xs), 1+3; got != want {
		t.Errorf("SumBillingLanes() = %d, want %d", got, want)
	}
	SortBillingLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingLanes() put %q first", xs[0].ID)
	}
	if len(IndexBillingLanes(xs)) != 3 {
		t.Errorf("IndexBillingLanes() = %v", IndexBillingLanes(xs))
	}
	if err := ValidateBillingLanes(ActiveBillingLanes(xs)); err != nil {
		t.Errorf("ValidateBillingLanes() = %v", err)
	}
}

func TestBillingLaneNormalizeClamps(t *testing.T) {
	x := SampleBillingLane("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingLane("r1", when)
	rev, err := r.PutBillingLane(x, 0)
	if err != nil {
		t.Fatalf("PutBillingLane() = %v", err)
	}
	got, gotRev, err := r.GetBillingLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingLanes() = %v, %v", all, err)
	}
	if _, err := r.PutBillingLane(BillingLane{}, 0); err == nil {
		t.Error("PutBillingLane() accepted an empty entity")
	}
}

func TestBillingSlotRoundTrip(t *testing.T) {
	x := SampleBillingSlot("billingslot-1", when)
	got, err := ParseBillingSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingSlot() = %v", err)
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

func TestBillingSlotParseRejectsBadInput(t *testing.T) {
	f := SampleBillingSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingSlot(f); err == nil {
		t.Error("ParseBillingSlot() accepted a record with no id")
	}
	f = SampleBillingSlot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseBillingSlot(f); err == nil {
		t.Error("ParseBillingSlot() accepted a non-numeric priority")
	}
}

func TestBillingSlotCollections(t *testing.T) {
	xs := []BillingSlot{SampleBillingSlot("a", when), SampleBillingSlot("bb", when), SampleBillingSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingSlots(xs)); got != 2 {
		t.Errorf("ActiveBillingSlots() kept %d, want 2", got)
	}
	if got, want := SumBillingSlots(xs), 1+3; got != want {
		t.Errorf("SumBillingSlots() = %d, want %d", got, want)
	}
	SortBillingSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingSlots() put %q first", xs[0].ID)
	}
	if len(IndexBillingSlots(xs)) != 3 {
		t.Errorf("IndexBillingSlots() = %v", IndexBillingSlots(xs))
	}
	if err := ValidateBillingSlots(ActiveBillingSlots(xs)); err != nil {
		t.Errorf("ValidateBillingSlots() = %v", err)
	}
}

func TestBillingSlotNormalizeClamps(t *testing.T) {
	x := SampleBillingSlot("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingSlot("r1", when)
	rev, err := r.PutBillingSlot(x, 0)
	if err != nil {
		t.Fatalf("PutBillingSlot() = %v", err)
	}
	got, gotRev, err := r.GetBillingSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingSlots() = %v, %v", all, err)
	}
	if _, err := r.PutBillingSlot(BillingSlot{}, 0); err == nil {
		t.Error("PutBillingSlot() accepted an empty entity")
	}
}

func TestBillingTagRoundTrip(t *testing.T) {
	x := SampleBillingTag("billingtag-1", when)
	got, err := ParseBillingTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingTag() = %v", err)
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

func TestBillingTagParseRejectsBadInput(t *testing.T) {
	f := SampleBillingTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingTag(f); err == nil {
		t.Error("ParseBillingTag() accepted a record with no id")
	}
	f = SampleBillingTag("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseBillingTag(f); err == nil {
		t.Error("ParseBillingTag() accepted a non-numeric weight")
	}
}

func TestBillingTagCollections(t *testing.T) {
	xs := []BillingTag{SampleBillingTag("a", when), SampleBillingTag("bb", when), SampleBillingTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingTags(xs)); got != 2 {
		t.Errorf("ActiveBillingTags() kept %d, want 2", got)
	}
	if got, want := SumBillingTags(xs), 1+3; got != want {
		t.Errorf("SumBillingTags() = %d, want %d", got, want)
	}
	SortBillingTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingTags() put %q first", xs[0].ID)
	}
	if len(IndexBillingTags(xs)) != 3 {
		t.Errorf("IndexBillingTags() = %v", IndexBillingTags(xs))
	}
	if err := ValidateBillingTags(ActiveBillingTags(xs)); err != nil {
		t.Errorf("ValidateBillingTags() = %v", err)
	}
}

func TestBillingTagNormalizeClamps(t *testing.T) {
	x := SampleBillingTag("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingTag("r1", when)
	rev, err := r.PutBillingTag(x, 0)
	if err != nil {
		t.Fatalf("PutBillingTag() = %v", err)
	}
	got, gotRev, err := r.GetBillingTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingTags() = %v, %v", all, err)
	}
	if _, err := r.PutBillingTag(BillingTag{}, 0); err == nil {
		t.Error("PutBillingTag() accepted an empty entity")
	}
}

func TestBillingRuleRoundTrip(t *testing.T) {
	x := SampleBillingRule("billingrule-1", when)
	got, err := ParseBillingRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingRule() = %v", err)
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

func TestBillingRuleParseRejectsBadInput(t *testing.T) {
	f := SampleBillingRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingRule(f); err == nil {
		t.Error("ParseBillingRule() accepted a record with no id")
	}
	f = SampleBillingRule("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseBillingRule(f); err == nil {
		t.Error("ParseBillingRule() accepted a non-numeric version")
	}
}

func TestBillingRuleCollections(t *testing.T) {
	xs := []BillingRule{SampleBillingRule("a", when), SampleBillingRule("bb", when), SampleBillingRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingRules(xs)); got != 2 {
		t.Errorf("ActiveBillingRules() kept %d, want 2", got)
	}
	if got, want := SumBillingRules(xs), 1+3; got != want {
		t.Errorf("SumBillingRules() = %d, want %d", got, want)
	}
	SortBillingRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingRules() put %q first", xs[0].ID)
	}
	if len(IndexBillingRules(xs)) != 3 {
		t.Errorf("IndexBillingRules() = %v", IndexBillingRules(xs))
	}
	if err := ValidateBillingRules(ActiveBillingRules(xs)); err != nil {
		t.Errorf("ValidateBillingRules() = %v", err)
	}
}

func TestBillingRuleNormalizeClamps(t *testing.T) {
	x := SampleBillingRule("a", when)
	x.Label = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingRule("r1", when)
	rev, err := r.PutBillingRule(x, 0)
	if err != nil {
		t.Fatalf("PutBillingRule() = %v", err)
	}
	got, gotRev, err := r.GetBillingRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingRules() = %v, %v", all, err)
	}
	if _, err := r.PutBillingRule(BillingRule{}, 0); err == nil {
		t.Error("PutBillingRule() accepted an empty entity")
	}
}

func TestBillingHoldRoundTrip(t *testing.T) {
	x := SampleBillingHold("billinghold-1", when)
	got, err := ParseBillingHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingHold() = %v", err)
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

func TestBillingHoldParseRejectsBadInput(t *testing.T) {
	f := SampleBillingHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingHold(f); err == nil {
		t.Error("ParseBillingHold() accepted a record with no id")
	}
	f = SampleBillingHold("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseBillingHold(f); err == nil {
		t.Error("ParseBillingHold() accepted a non-numeric attempts")
	}
}

func TestBillingHoldCollections(t *testing.T) {
	xs := []BillingHold{SampleBillingHold("a", when), SampleBillingHold("bb", when), SampleBillingHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingHolds(xs)); got != 2 {
		t.Errorf("ActiveBillingHolds() kept %d, want 2", got)
	}
	if got, want := SumBillingHolds(xs), 1+3; got != want {
		t.Errorf("SumBillingHolds() = %d, want %d", got, want)
	}
	SortBillingHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingHolds() put %q first", xs[0].ID)
	}
	if len(IndexBillingHolds(xs)) != 3 {
		t.Errorf("IndexBillingHolds() = %v", IndexBillingHolds(xs))
	}
	if err := ValidateBillingHolds(ActiveBillingHolds(xs)); err != nil {
		t.Errorf("ValidateBillingHolds() = %v", err)
	}
}

func TestBillingHoldNormalizeClamps(t *testing.T) {
	x := SampleBillingHold("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingHold("r1", when)
	rev, err := r.PutBillingHold(x, 0)
	if err != nil {
		t.Fatalf("PutBillingHold() = %v", err)
	}
	got, gotRev, err := r.GetBillingHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingHolds() = %v, %v", all, err)
	}
	if _, err := r.PutBillingHold(BillingHold{}, 0); err == nil {
		t.Error("PutBillingHold() accepted an empty entity")
	}
}

func TestBillingClaimRoundTrip(t *testing.T) {
	x := SampleBillingClaim("billingclaim-1", when)
	got, err := ParseBillingClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingClaim() = %v", err)
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

func TestBillingClaimParseRejectsBadInput(t *testing.T) {
	f := SampleBillingClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingClaim(f); err == nil {
		t.Error("ParseBillingClaim() accepted a record with no id")
	}
	f = SampleBillingClaim("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseBillingClaim(f); err == nil {
		t.Error("ParseBillingClaim() accepted a non-numeric count")
	}
}

func TestBillingClaimCollections(t *testing.T) {
	xs := []BillingClaim{SampleBillingClaim("a", when), SampleBillingClaim("bb", when), SampleBillingClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingClaims(xs)); got != 2 {
		t.Errorf("ActiveBillingClaims() kept %d, want 2", got)
	}
	if got, want := SumBillingClaims(xs), 1+3; got != want {
		t.Errorf("SumBillingClaims() = %d, want %d", got, want)
	}
	SortBillingClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingClaims() put %q first", xs[0].ID)
	}
	if len(IndexBillingClaims(xs)) != 3 {
		t.Errorf("IndexBillingClaims() = %v", IndexBillingClaims(xs))
	}
	if err := ValidateBillingClaims(ActiveBillingClaims(xs)); err != nil {
		t.Errorf("ValidateBillingClaims() = %v", err)
	}
}

func TestBillingClaimNormalizeClamps(t *testing.T) {
	x := SampleBillingClaim("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingClaim("r1", when)
	rev, err := r.PutBillingClaim(x, 0)
	if err != nil {
		t.Fatalf("PutBillingClaim() = %v", err)
	}
	got, gotRev, err := r.GetBillingClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingClaims() = %v, %v", all, err)
	}
	if _, err := r.PutBillingClaim(BillingClaim{}, 0); err == nil {
		t.Error("PutBillingClaim() accepted an empty entity")
	}
}

func TestBillingCursorRoundTrip(t *testing.T) {
	x := SampleBillingCursor("billingcursor-1", when)
	got, err := ParseBillingCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseBillingCursor() = %v", err)
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

func TestBillingCursorParseRejectsBadInput(t *testing.T) {
	f := SampleBillingCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseBillingCursor(f); err == nil {
		t.Error("ParseBillingCursor() accepted a record with no id")
	}
	f = SampleBillingCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseBillingCursor(f); err == nil {
		t.Error("ParseBillingCursor() accepted a non-numeric weight")
	}
}

func TestBillingCursorCollections(t *testing.T) {
	xs := []BillingCursor{SampleBillingCursor("a", when), SampleBillingCursor("bb", when), SampleBillingCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveBillingCursors(xs)); got != 2 {
		t.Errorf("ActiveBillingCursors() kept %d, want 2", got)
	}
	if got, want := SumBillingCursors(xs), 1+3; got != want {
		t.Errorf("SumBillingCursors() = %d, want %d", got, want)
	}
	SortBillingCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortBillingCursors() put %q first", xs[0].ID)
	}
	if len(IndexBillingCursors(xs)) != 3 {
		t.Errorf("IndexBillingCursors() = %v", IndexBillingCursors(xs))
	}
	if err := ValidateBillingCursors(ActiveBillingCursors(xs)); err != nil {
		t.Errorf("ValidateBillingCursors() = %v", err)
	}
}

func TestBillingCursorNormalizeClamps(t *testing.T) {
	x := SampleBillingCursor("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestBillingCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleBillingCursor("r1", when)
	rev, err := r.PutBillingCursor(x, 0)
	if err != nil {
		t.Fatalf("PutBillingCursor() = %v", err)
	}
	got, gotRev, err := r.GetBillingCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetBillingCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListBillingCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListBillingCursors() = %v, %v", all, err)
	}
	if _, err := r.PutBillingCursor(BillingCursor{}, 0); err == nil {
		t.Error("PutBillingCursor() accepted an empty entity")
	}
}
