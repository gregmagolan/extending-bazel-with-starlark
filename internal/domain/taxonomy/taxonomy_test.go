package taxonomy

import (
	"strings"
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
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
		if !strings.HasPrefix(Prefixes[i], "taxonomy/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleTaxonomyRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestTaxonomyRecordRoundTrip(t *testing.T) {
	x := SampleTaxonomyRecord("taxonomyrecord-1", when)
	got, err := ParseTaxonomyRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyRecord() = %v", err)
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

func TestTaxonomyRecordParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyRecord(f); err == nil {
		t.Error("ParseTaxonomyRecord() accepted a record with no id")
	}
	f = SampleTaxonomyRecord("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTaxonomyRecord(f); err == nil {
		t.Error("ParseTaxonomyRecord() accepted a non-numeric attempts")
	}
}

func TestTaxonomyRecordCollections(t *testing.T) {
	xs := []TaxonomyRecord{SampleTaxonomyRecord("a", when), SampleTaxonomyRecord("bb", when), SampleTaxonomyRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyRecords(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyRecords() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyRecords(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyRecords() = %d, want %d", got, want)
	}
	SortTaxonomyRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyRecords() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyRecords(xs)) != 3 {
		t.Errorf("IndexTaxonomyRecords() = %v", IndexTaxonomyRecords(xs))
	}
	if err := ValidateTaxonomyRecords(ActiveTaxonomyRecords(xs)); err != nil {
		t.Errorf("ValidateTaxonomyRecords() = %v", err)
	}
}

func TestTaxonomyRecordNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyRecord("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyEntryRoundTrip(t *testing.T) {
	x := SampleTaxonomyEntry("taxonomyentry-1", when)
	got, err := ParseTaxonomyEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyEntry() = %v", err)
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

func TestTaxonomyEntryParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyEntry(f); err == nil {
		t.Error("ParseTaxonomyEntry() accepted a record with no id")
	}
	f = SampleTaxonomyEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTaxonomyEntry(f); err == nil {
		t.Error("ParseTaxonomyEntry() accepted a non-numeric offset")
	}
}

func TestTaxonomyEntryCollections(t *testing.T) {
	xs := []TaxonomyEntry{SampleTaxonomyEntry("a", when), SampleTaxonomyEntry("bb", when), SampleTaxonomyEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyEntrys(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyEntrys() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyEntrys(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyEntrys() = %d, want %d", got, want)
	}
	SortTaxonomyEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyEntrys() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyEntrys(xs)) != 3 {
		t.Errorf("IndexTaxonomyEntrys() = %v", IndexTaxonomyEntrys(xs))
	}
	if err := ValidateTaxonomyEntrys(ActiveTaxonomyEntrys(xs)); err != nil {
		t.Errorf("ValidateTaxonomyEntrys() = %v", err)
	}
}

func TestTaxonomyEntryNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyEntry("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyItemRoundTrip(t *testing.T) {
	x := SampleTaxonomyItem("taxonomyitem-1", when)
	got, err := ParseTaxonomyItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyItem() = %v", err)
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

func TestTaxonomyItemParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyItem(f); err == nil {
		t.Error("ParseTaxonomyItem() accepted a record with no id")
	}
	f = SampleTaxonomyItem("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTaxonomyItem(f); err == nil {
		t.Error("ParseTaxonomyItem() accepted a non-numeric priority")
	}
}

func TestTaxonomyItemCollections(t *testing.T) {
	xs := []TaxonomyItem{SampleTaxonomyItem("a", when), SampleTaxonomyItem("bb", when), SampleTaxonomyItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyItems(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyItems() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyItems(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyItems() = %d, want %d", got, want)
	}
	SortTaxonomyItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyItems() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyItems(xs)) != 3 {
		t.Errorf("IndexTaxonomyItems() = %v", IndexTaxonomyItems(xs))
	}
	if err := ValidateTaxonomyItems(ActiveTaxonomyItems(xs)); err != nil {
		t.Errorf("ValidateTaxonomyItems() = %v", err)
	}
}

func TestTaxonomyItemNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyItem("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyDraftRoundTrip(t *testing.T) {
	x := SampleTaxonomyDraft("taxonomydraft-1", when)
	got, err := ParseTaxonomyDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyDraft() = %v", err)
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

func TestTaxonomyDraftParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyDraft(f); err == nil {
		t.Error("ParseTaxonomyDraft() accepted a record with no id")
	}
	f = SampleTaxonomyDraft("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTaxonomyDraft(f); err == nil {
		t.Error("ParseTaxonomyDraft() accepted a non-numeric offset")
	}
}

func TestTaxonomyDraftCollections(t *testing.T) {
	xs := []TaxonomyDraft{SampleTaxonomyDraft("a", when), SampleTaxonomyDraft("bb", when), SampleTaxonomyDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyDrafts(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyDrafts() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyDrafts(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyDrafts() = %d, want %d", got, want)
	}
	SortTaxonomyDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyDrafts() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyDrafts(xs)) != 3 {
		t.Errorf("IndexTaxonomyDrafts() = %v", IndexTaxonomyDrafts(xs))
	}
	if err := ValidateTaxonomyDrafts(ActiveTaxonomyDrafts(xs)); err != nil {
		t.Errorf("ValidateTaxonomyDrafts() = %v", err)
	}
}

func TestTaxonomyDraftNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyDraft("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyRevisionRoundTrip(t *testing.T) {
	x := SampleTaxonomyRevision("taxonomyrevision-1", when)
	got, err := ParseTaxonomyRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyRevision() = %v", err)
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

func TestTaxonomyRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyRevision(f); err == nil {
		t.Error("ParseTaxonomyRevision() accepted a record with no id")
	}
	f = SampleTaxonomyRevision("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTaxonomyRevision(f); err == nil {
		t.Error("ParseTaxonomyRevision() accepted a non-numeric retained")
	}
}

func TestTaxonomyRevisionCollections(t *testing.T) {
	xs := []TaxonomyRevision{SampleTaxonomyRevision("a", when), SampleTaxonomyRevision("bb", when), SampleTaxonomyRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyRevisions(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyRevisions() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyRevisions(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyRevisions() = %d, want %d", got, want)
	}
	SortTaxonomyRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyRevisions() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyRevisions(xs)) != 3 {
		t.Errorf("IndexTaxonomyRevisions() = %v", IndexTaxonomyRevisions(xs))
	}
	if err := ValidateTaxonomyRevisions(ActiveTaxonomyRevisions(xs)); err != nil {
		t.Errorf("ValidateTaxonomyRevisions() = %v", err)
	}
}

func TestTaxonomyRevisionNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyRevision("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomySnapshotRoundTrip(t *testing.T) {
	x := SampleTaxonomySnapshot("taxonomysnapshot-1", when)
	got, err := ParseTaxonomySnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomySnapshot() = %v", err)
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

func TestTaxonomySnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomySnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomySnapshot(f); err == nil {
		t.Error("ParseTaxonomySnapshot() accepted a record with no id")
	}
	f = SampleTaxonomySnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTaxonomySnapshot(f); err == nil {
		t.Error("ParseTaxonomySnapshot() accepted a non-numeric priority")
	}
}

func TestTaxonomySnapshotCollections(t *testing.T) {
	xs := []TaxonomySnapshot{SampleTaxonomySnapshot("a", when), SampleTaxonomySnapshot("bb", when), SampleTaxonomySnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomySnapshots(xs)); got != 2 {
		t.Errorf("ActiveTaxonomySnapshots() kept %d, want 2", got)
	}
	if got, want := SumTaxonomySnapshots(xs), 1+3; got != want {
		t.Errorf("SumTaxonomySnapshots() = %d, want %d", got, want)
	}
	SortTaxonomySnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomySnapshots() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomySnapshots(xs)) != 3 {
		t.Errorf("IndexTaxonomySnapshots() = %v", IndexTaxonomySnapshots(xs))
	}
	if err := ValidateTaxonomySnapshots(ActiveTaxonomySnapshots(xs)); err != nil {
		t.Errorf("ValidateTaxonomySnapshots() = %v", err)
	}
}

func TestTaxonomySnapshotNormalizeClamps(t *testing.T) {
	x := SampleTaxonomySnapshot("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyBatchRoundTrip(t *testing.T) {
	x := SampleTaxonomyBatch("taxonomybatch-1", when)
	got, err := ParseTaxonomyBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyBatch() = %v", err)
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

func TestTaxonomyBatchParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyBatch(f); err == nil {
		t.Error("ParseTaxonomyBatch() accepted a record with no id")
	}
	f = SampleTaxonomyBatch("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTaxonomyBatch(f); err == nil {
		t.Error("ParseTaxonomyBatch() accepted a non-numeric priority")
	}
}

func TestTaxonomyBatchCollections(t *testing.T) {
	xs := []TaxonomyBatch{SampleTaxonomyBatch("a", when), SampleTaxonomyBatch("bb", when), SampleTaxonomyBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyBatchs(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyBatchs() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyBatchs(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyBatchs() = %d, want %d", got, want)
	}
	SortTaxonomyBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyBatchs() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyBatchs(xs)) != 3 {
		t.Errorf("IndexTaxonomyBatchs() = %v", IndexTaxonomyBatchs(xs))
	}
	if err := ValidateTaxonomyBatchs(ActiveTaxonomyBatchs(xs)); err != nil {
		t.Errorf("ValidateTaxonomyBatchs() = %v", err)
	}
}

func TestTaxonomyBatchNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyBatch("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyRequestRoundTrip(t *testing.T) {
	x := SampleTaxonomyRequest("taxonomyrequest-1", when)
	got, err := ParseTaxonomyRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyRequest() = %v", err)
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

func TestTaxonomyRequestParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyRequest(f); err == nil {
		t.Error("ParseTaxonomyRequest() accepted a record with no id")
	}
	f = SampleTaxonomyRequest("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTaxonomyRequest(f); err == nil {
		t.Error("ParseTaxonomyRequest() accepted a non-numeric capacity")
	}
}

func TestTaxonomyRequestCollections(t *testing.T) {
	xs := []TaxonomyRequest{SampleTaxonomyRequest("a", when), SampleTaxonomyRequest("bb", when), SampleTaxonomyRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyRequests(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyRequests() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyRequests(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyRequests() = %d, want %d", got, want)
	}
	SortTaxonomyRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyRequests() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyRequests(xs)) != 3 {
		t.Errorf("IndexTaxonomyRequests() = %v", IndexTaxonomyRequests(xs))
	}
	if err := ValidateTaxonomyRequests(ActiveTaxonomyRequests(xs)); err != nil {
		t.Errorf("ValidateTaxonomyRequests() = %v", err)
	}
}

func TestTaxonomyRequestNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyRequest("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomySummaryRoundTrip(t *testing.T) {
	x := SampleTaxonomySummary("taxonomysummary-1", when)
	got, err := ParseTaxonomySummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomySummary() = %v", err)
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

func TestTaxonomySummaryParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomySummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomySummary(f); err == nil {
		t.Error("ParseTaxonomySummary() accepted a record with no id")
	}
	f = SampleTaxonomySummary("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTaxonomySummary(f); err == nil {
		t.Error("ParseTaxonomySummary() accepted a non-numeric offset")
	}
}

func TestTaxonomySummaryCollections(t *testing.T) {
	xs := []TaxonomySummary{SampleTaxonomySummary("a", when), SampleTaxonomySummary("bb", when), SampleTaxonomySummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomySummarys(xs)); got != 2 {
		t.Errorf("ActiveTaxonomySummarys() kept %d, want 2", got)
	}
	if got, want := SumTaxonomySummarys(xs), 1+3; got != want {
		t.Errorf("SumTaxonomySummarys() = %d, want %d", got, want)
	}
	SortTaxonomySummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomySummarys() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomySummarys(xs)) != 3 {
		t.Errorf("IndexTaxonomySummarys() = %v", IndexTaxonomySummarys(xs))
	}
	if err := ValidateTaxonomySummarys(ActiveTaxonomySummarys(xs)); err != nil {
		t.Errorf("ValidateTaxonomySummarys() = %v", err)
	}
}

func TestTaxonomySummaryNormalizeClamps(t *testing.T) {
	x := SampleTaxonomySummary("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyPolicyRoundTrip(t *testing.T) {
	x := SampleTaxonomyPolicy("taxonomypolicy-1", when)
	got, err := ParseTaxonomyPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyPolicy() = %v", err)
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

func TestTaxonomyPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyPolicy(f); err == nil {
		t.Error("ParseTaxonomyPolicy() accepted a record with no id")
	}
	f = SampleTaxonomyPolicy("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTaxonomyPolicy(f); err == nil {
		t.Error("ParseTaxonomyPolicy() accepted a non-numeric attempts")
	}
}

func TestTaxonomyPolicyCollections(t *testing.T) {
	xs := []TaxonomyPolicy{SampleTaxonomyPolicy("a", when), SampleTaxonomyPolicy("bb", when), SampleTaxonomyPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyPolicys(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyPolicys() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyPolicys(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyPolicys() = %d, want %d", got, want)
	}
	SortTaxonomyPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyPolicys() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyPolicys(xs)) != 3 {
		t.Errorf("IndexTaxonomyPolicys() = %v", IndexTaxonomyPolicys(xs))
	}
	if err := ValidateTaxonomyPolicys(ActiveTaxonomyPolicys(xs)); err != nil {
		t.Errorf("ValidateTaxonomyPolicys() = %v", err)
	}
}

func TestTaxonomyPolicyNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyPolicy("a", when)
	x.Reference = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyLinkRoundTrip(t *testing.T) {
	x := SampleTaxonomyLink("taxonomylink-1", when)
	got, err := ParseTaxonomyLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyLink() = %v", err)
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

func TestTaxonomyLinkParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyLink(f); err == nil {
		t.Error("ParseTaxonomyLink() accepted a record with no id")
	}
	f = SampleTaxonomyLink("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTaxonomyLink(f); err == nil {
		t.Error("ParseTaxonomyLink() accepted a non-numeric weight")
	}
}

func TestTaxonomyLinkCollections(t *testing.T) {
	xs := []TaxonomyLink{SampleTaxonomyLink("a", when), SampleTaxonomyLink("bb", when), SampleTaxonomyLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyLinks(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyLinks() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyLinks(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyLinks() = %d, want %d", got, want)
	}
	SortTaxonomyLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyLinks() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyLinks(xs)) != 3 {
		t.Errorf("IndexTaxonomyLinks() = %v", IndexTaxonomyLinks(xs))
	}
	if err := ValidateTaxonomyLinks(ActiveTaxonomyLinks(xs)); err != nil {
		t.Errorf("ValidateTaxonomyLinks() = %v", err)
	}
}

func TestTaxonomyLinkNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyLink("a", when)
	x.Group = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyNoteRoundTrip(t *testing.T) {
	x := SampleTaxonomyNote("taxonomynote-1", when)
	got, err := ParseTaxonomyNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyNote() = %v", err)
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

func TestTaxonomyNoteParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyNote(f); err == nil {
		t.Error("ParseTaxonomyNote() accepted a record with no id")
	}
	f = SampleTaxonomyNote("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTaxonomyNote(f); err == nil {
		t.Error("ParseTaxonomyNote() accepted a non-numeric count")
	}
}

func TestTaxonomyNoteCollections(t *testing.T) {
	xs := []TaxonomyNote{SampleTaxonomyNote("a", when), SampleTaxonomyNote("bb", when), SampleTaxonomyNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyNotes(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyNotes() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyNotes(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyNotes() = %d, want %d", got, want)
	}
	SortTaxonomyNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyNotes() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyNotes(xs)) != 3 {
		t.Errorf("IndexTaxonomyNotes() = %v", IndexTaxonomyNotes(xs))
	}
	if err := ValidateTaxonomyNotes(ActiveTaxonomyNotes(xs)); err != nil {
		t.Errorf("ValidateTaxonomyNotes() = %v", err)
	}
}

func TestTaxonomyNoteNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyNote("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyWindowRoundTrip(t *testing.T) {
	x := SampleTaxonomyWindow("taxonomywindow-1", when)
	got, err := ParseTaxonomyWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyWindow() = %v", err)
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

func TestTaxonomyWindowParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyWindow(f); err == nil {
		t.Error("ParseTaxonomyWindow() accepted a record with no id")
	}
	f = SampleTaxonomyWindow("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTaxonomyWindow(f); err == nil {
		t.Error("ParseTaxonomyWindow() accepted a non-numeric capacity")
	}
}

func TestTaxonomyWindowCollections(t *testing.T) {
	xs := []TaxonomyWindow{SampleTaxonomyWindow("a", when), SampleTaxonomyWindow("bb", when), SampleTaxonomyWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyWindows(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyWindows() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyWindows(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyWindows() = %d, want %d", got, want)
	}
	SortTaxonomyWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyWindows() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyWindows(xs)) != 3 {
		t.Errorf("IndexTaxonomyWindows() = %v", IndexTaxonomyWindows(xs))
	}
	if err := ValidateTaxonomyWindows(ActiveTaxonomyWindows(xs)); err != nil {
		t.Errorf("ValidateTaxonomyWindows() = %v", err)
	}
}

func TestTaxonomyWindowNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyWindow("a", when)
	x.Label = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyQuotaRoundTrip(t *testing.T) {
	x := SampleTaxonomyQuota("taxonomyquota-1", when)
	got, err := ParseTaxonomyQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyQuota() = %v", err)
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

func TestTaxonomyQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyQuota(f); err == nil {
		t.Error("ParseTaxonomyQuota() accepted a record with no id")
	}
	f = SampleTaxonomyQuota("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTaxonomyQuota(f); err == nil {
		t.Error("ParseTaxonomyQuota() accepted a non-numeric count")
	}
}

func TestTaxonomyQuotaCollections(t *testing.T) {
	xs := []TaxonomyQuota{SampleTaxonomyQuota("a", when), SampleTaxonomyQuota("bb", when), SampleTaxonomyQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyQuotas(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyQuotas() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyQuotas(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyQuotas() = %d, want %d", got, want)
	}
	SortTaxonomyQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyQuotas() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyQuotas(xs)) != 3 {
		t.Errorf("IndexTaxonomyQuotas() = %v", IndexTaxonomyQuotas(xs))
	}
	if err := ValidateTaxonomyQuotas(ActiveTaxonomyQuotas(xs)); err != nil {
		t.Errorf("ValidateTaxonomyQuotas() = %v", err)
	}
}

func TestTaxonomyQuotaNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyQuota("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyBucketRoundTrip(t *testing.T) {
	x := SampleTaxonomyBucket("taxonomybucket-1", when)
	got, err := ParseTaxonomyBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyBucket() = %v", err)
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

func TestTaxonomyBucketParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyBucket(f); err == nil {
		t.Error("ParseTaxonomyBucket() accepted a record with no id")
	}
	f = SampleTaxonomyBucket("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTaxonomyBucket(f); err == nil {
		t.Error("ParseTaxonomyBucket() accepted a non-numeric priority")
	}
}

func TestTaxonomyBucketCollections(t *testing.T) {
	xs := []TaxonomyBucket{SampleTaxonomyBucket("a", when), SampleTaxonomyBucket("bb", when), SampleTaxonomyBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyBuckets(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyBuckets() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyBuckets(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyBuckets() = %d, want %d", got, want)
	}
	SortTaxonomyBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyBuckets() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyBuckets(xs)) != 3 {
		t.Errorf("IndexTaxonomyBuckets() = %v", IndexTaxonomyBuckets(xs))
	}
	if err := ValidateTaxonomyBuckets(ActiveTaxonomyBuckets(xs)); err != nil {
		t.Errorf("ValidateTaxonomyBuckets() = %v", err)
	}
}

func TestTaxonomyBucketNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyBucket("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyLaneRoundTrip(t *testing.T) {
	x := SampleTaxonomyLane("taxonomylane-1", when)
	got, err := ParseTaxonomyLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyLane() = %v", err)
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

func TestTaxonomyLaneParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyLane(f); err == nil {
		t.Error("ParseTaxonomyLane() accepted a record with no id")
	}
	f = SampleTaxonomyLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTaxonomyLane(f); err == nil {
		t.Error("ParseTaxonomyLane() accepted a non-numeric priority")
	}
}

func TestTaxonomyLaneCollections(t *testing.T) {
	xs := []TaxonomyLane{SampleTaxonomyLane("a", when), SampleTaxonomyLane("bb", when), SampleTaxonomyLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyLanes(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyLanes() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyLanes(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyLanes() = %d, want %d", got, want)
	}
	SortTaxonomyLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyLanes() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyLanes(xs)) != 3 {
		t.Errorf("IndexTaxonomyLanes() = %v", IndexTaxonomyLanes(xs))
	}
	if err := ValidateTaxonomyLanes(ActiveTaxonomyLanes(xs)); err != nil {
		t.Errorf("ValidateTaxonomyLanes() = %v", err)
	}
}

func TestTaxonomyLaneNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyLane("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomySlotRoundTrip(t *testing.T) {
	x := SampleTaxonomySlot("taxonomyslot-1", when)
	got, err := ParseTaxonomySlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomySlot() = %v", err)
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

func TestTaxonomySlotParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomySlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomySlot(f); err == nil {
		t.Error("ParseTaxonomySlot() accepted a record with no id")
	}
	f = SampleTaxonomySlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTaxonomySlot(f); err == nil {
		t.Error("ParseTaxonomySlot() accepted a non-numeric count")
	}
}

func TestTaxonomySlotCollections(t *testing.T) {
	xs := []TaxonomySlot{SampleTaxonomySlot("a", when), SampleTaxonomySlot("bb", when), SampleTaxonomySlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomySlots(xs)); got != 2 {
		t.Errorf("ActiveTaxonomySlots() kept %d, want 2", got)
	}
	if got, want := SumTaxonomySlots(xs), 1+3; got != want {
		t.Errorf("SumTaxonomySlots() = %d, want %d", got, want)
	}
	SortTaxonomySlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomySlots() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomySlots(xs)) != 3 {
		t.Errorf("IndexTaxonomySlots() = %v", IndexTaxonomySlots(xs))
	}
	if err := ValidateTaxonomySlots(ActiveTaxonomySlots(xs)); err != nil {
		t.Errorf("ValidateTaxonomySlots() = %v", err)
	}
}

func TestTaxonomySlotNormalizeClamps(t *testing.T) {
	x := SampleTaxonomySlot("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyTagRoundTrip(t *testing.T) {
	x := SampleTaxonomyTag("taxonomytag-1", when)
	got, err := ParseTaxonomyTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyTag() = %v", err)
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

func TestTaxonomyTagParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyTag(f); err == nil {
		t.Error("ParseTaxonomyTag() accepted a record with no id")
	}
	f = SampleTaxonomyTag("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTaxonomyTag(f); err == nil {
		t.Error("ParseTaxonomyTag() accepted a non-numeric retained")
	}
}

func TestTaxonomyTagCollections(t *testing.T) {
	xs := []TaxonomyTag{SampleTaxonomyTag("a", when), SampleTaxonomyTag("bb", when), SampleTaxonomyTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyTags(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyTags() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyTags(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyTags() = %d, want %d", got, want)
	}
	SortTaxonomyTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyTags() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyTags(xs)) != 3 {
		t.Errorf("IndexTaxonomyTags() = %v", IndexTaxonomyTags(xs))
	}
	if err := ValidateTaxonomyTags(ActiveTaxonomyTags(xs)); err != nil {
		t.Errorf("ValidateTaxonomyTags() = %v", err)
	}
}

func TestTaxonomyTagNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyTag("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyRuleRoundTrip(t *testing.T) {
	x := SampleTaxonomyRule("taxonomyrule-1", when)
	got, err := ParseTaxonomyRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyRule() = %v", err)
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

func TestTaxonomyRuleParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyRule(f); err == nil {
		t.Error("ParseTaxonomyRule() accepted a record with no id")
	}
	f = SampleTaxonomyRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTaxonomyRule(f); err == nil {
		t.Error("ParseTaxonomyRule() accepted a non-numeric offset")
	}
}

func TestTaxonomyRuleCollections(t *testing.T) {
	xs := []TaxonomyRule{SampleTaxonomyRule("a", when), SampleTaxonomyRule("bb", when), SampleTaxonomyRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyRules(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyRules() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyRules(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyRules() = %d, want %d", got, want)
	}
	SortTaxonomyRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyRules() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyRules(xs)) != 3 {
		t.Errorf("IndexTaxonomyRules() = %v", IndexTaxonomyRules(xs))
	}
	if err := ValidateTaxonomyRules(ActiveTaxonomyRules(xs)); err != nil {
		t.Errorf("ValidateTaxonomyRules() = %v", err)
	}
}

func TestTaxonomyRuleNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyRule("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyHoldRoundTrip(t *testing.T) {
	x := SampleTaxonomyHold("taxonomyhold-1", when)
	got, err := ParseTaxonomyHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyHold() = %v", err)
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

func TestTaxonomyHoldParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyHold(f); err == nil {
		t.Error("ParseTaxonomyHold() accepted a record with no id")
	}
	f = SampleTaxonomyHold("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTaxonomyHold(f); err == nil {
		t.Error("ParseTaxonomyHold() accepted a non-numeric count")
	}
}

func TestTaxonomyHoldCollections(t *testing.T) {
	xs := []TaxonomyHold{SampleTaxonomyHold("a", when), SampleTaxonomyHold("bb", when), SampleTaxonomyHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyHolds(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyHolds() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyHolds(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyHolds() = %d, want %d", got, want)
	}
	SortTaxonomyHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyHolds() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyHolds(xs)) != 3 {
		t.Errorf("IndexTaxonomyHolds() = %v", IndexTaxonomyHolds(xs))
	}
	if err := ValidateTaxonomyHolds(ActiveTaxonomyHolds(xs)); err != nil {
		t.Errorf("ValidateTaxonomyHolds() = %v", err)
	}
}

func TestTaxonomyHoldNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyHold("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyClaimRoundTrip(t *testing.T) {
	x := SampleTaxonomyClaim("taxonomyclaim-1", when)
	got, err := ParseTaxonomyClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyClaim() = %v", err)
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

func TestTaxonomyClaimParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyClaim(f); err == nil {
		t.Error("ParseTaxonomyClaim() accepted a record with no id")
	}
	f = SampleTaxonomyClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTaxonomyClaim(f); err == nil {
		t.Error("ParseTaxonomyClaim() accepted a non-numeric retained")
	}
}

func TestTaxonomyClaimCollections(t *testing.T) {
	xs := []TaxonomyClaim{SampleTaxonomyClaim("a", when), SampleTaxonomyClaim("bb", when), SampleTaxonomyClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyClaims(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyClaims() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyClaims(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyClaims() = %d, want %d", got, want)
	}
	SortTaxonomyClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyClaims() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyClaims(xs)) != 3 {
		t.Errorf("IndexTaxonomyClaims() = %v", IndexTaxonomyClaims(xs))
	}
	if err := ValidateTaxonomyClaims(ActiveTaxonomyClaims(xs)); err != nil {
		t.Errorf("ValidateTaxonomyClaims() = %v", err)
	}
}

func TestTaxonomyClaimNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyClaim("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTaxonomyCursorRoundTrip(t *testing.T) {
	x := SampleTaxonomyCursor("taxonomycursor-1", when)
	got, err := ParseTaxonomyCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseTaxonomyCursor() = %v", err)
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

func TestTaxonomyCursorParseRejectsBadInput(t *testing.T) {
	f := SampleTaxonomyCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTaxonomyCursor(f); err == nil {
		t.Error("ParseTaxonomyCursor() accepted a record with no id")
	}
	f = SampleTaxonomyCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTaxonomyCursor(f); err == nil {
		t.Error("ParseTaxonomyCursor() accepted a non-numeric weight")
	}
}

func TestTaxonomyCursorCollections(t *testing.T) {
	xs := []TaxonomyCursor{SampleTaxonomyCursor("a", when), SampleTaxonomyCursor("bb", when), SampleTaxonomyCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTaxonomyCursors(xs)); got != 2 {
		t.Errorf("ActiveTaxonomyCursors() kept %d, want 2", got)
	}
	if got, want := SumTaxonomyCursors(xs), 1+3; got != want {
		t.Errorf("SumTaxonomyCursors() = %d, want %d", got, want)
	}
	SortTaxonomyCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTaxonomyCursors() put %q first", xs[0].ID)
	}
	if len(IndexTaxonomyCursors(xs)) != 3 {
		t.Errorf("IndexTaxonomyCursors() = %v", IndexTaxonomyCursors(xs))
	}
	if err := ValidateTaxonomyCursors(ActiveTaxonomyCursors(xs)); err != nil {
		t.Errorf("ValidateTaxonomyCursors() = %v", err)
	}
}

func TestTaxonomyCursorNormalizeClamps(t *testing.T) {
	x := SampleTaxonomyCursor("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
