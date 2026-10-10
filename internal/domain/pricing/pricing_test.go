package pricing

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
		if !strings.HasPrefix(Prefixes[i], "pricing/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SamplePricingRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestPricingRecordRoundTrip(t *testing.T) {
	x := SamplePricingRecord("pricingrecord-1", when)
	got, err := ParsePricingRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingRecord() = %v", err)
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

func TestPricingRecordParseRejectsBadInput(t *testing.T) {
	f := SamplePricingRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingRecord(f); err == nil {
		t.Error("ParsePricingRecord() accepted a record with no id")
	}
	f = SamplePricingRecord("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePricingRecord(f); err == nil {
		t.Error("ParsePricingRecord() accepted a non-numeric retained")
	}
}

func TestPricingRecordCollections(t *testing.T) {
	xs := []PricingRecord{SamplePricingRecord("a", when), SamplePricingRecord("bb", when), SamplePricingRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingRecords(xs)); got != 2 {
		t.Errorf("ActivePricingRecords() kept %d, want 2", got)
	}
	if got, want := SumPricingRecords(xs), 1+3; got != want {
		t.Errorf("SumPricingRecords() = %d, want %d", got, want)
	}
	SortPricingRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingRecords() put %q first", xs[0].ID)
	}
	if len(IndexPricingRecords(xs)) != 3 {
		t.Errorf("IndexPricingRecords() = %v", IndexPricingRecords(xs))
	}
	if err := ValidatePricingRecords(ActivePricingRecords(xs)); err != nil {
		t.Errorf("ValidatePricingRecords() = %v", err)
	}
}

func TestPricingRecordNormalizeClamps(t *testing.T) {
	x := SamplePricingRecord("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingEntryRoundTrip(t *testing.T) {
	x := SamplePricingEntry("pricingentry-1", when)
	got, err := ParsePricingEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingEntry() = %v", err)
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

func TestPricingEntryParseRejectsBadInput(t *testing.T) {
	f := SamplePricingEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingEntry(f); err == nil {
		t.Error("ParsePricingEntry() accepted a record with no id")
	}
	f = SamplePricingEntry("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePricingEntry(f); err == nil {
		t.Error("ParsePricingEntry() accepted a non-numeric version")
	}
}

func TestPricingEntryCollections(t *testing.T) {
	xs := []PricingEntry{SamplePricingEntry("a", when), SamplePricingEntry("bb", when), SamplePricingEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingEntrys(xs)); got != 2 {
		t.Errorf("ActivePricingEntrys() kept %d, want 2", got)
	}
	if got, want := SumPricingEntrys(xs), 1+3; got != want {
		t.Errorf("SumPricingEntrys() = %d, want %d", got, want)
	}
	SortPricingEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingEntrys() put %q first", xs[0].ID)
	}
	if len(IndexPricingEntrys(xs)) != 3 {
		t.Errorf("IndexPricingEntrys() = %v", IndexPricingEntrys(xs))
	}
	if err := ValidatePricingEntrys(ActivePricingEntrys(xs)); err != nil {
		t.Errorf("ValidatePricingEntrys() = %v", err)
	}
}

func TestPricingEntryNormalizeClamps(t *testing.T) {
	x := SamplePricingEntry("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingItemRoundTrip(t *testing.T) {
	x := SamplePricingItem("pricingitem-1", when)
	got, err := ParsePricingItem(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingItem() = %v", err)
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

func TestPricingItemParseRejectsBadInput(t *testing.T) {
	f := SamplePricingItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingItem(f); err == nil {
		t.Error("ParsePricingItem() accepted a record with no id")
	}
	f = SamplePricingItem("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePricingItem(f); err == nil {
		t.Error("ParsePricingItem() accepted a non-numeric retained")
	}
}

func TestPricingItemCollections(t *testing.T) {
	xs := []PricingItem{SamplePricingItem("a", when), SamplePricingItem("bb", when), SamplePricingItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingItems(xs)); got != 2 {
		t.Errorf("ActivePricingItems() kept %d, want 2", got)
	}
	if got, want := SumPricingItems(xs), 1+3; got != want {
		t.Errorf("SumPricingItems() = %d, want %d", got, want)
	}
	SortPricingItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingItems() put %q first", xs[0].ID)
	}
	if len(IndexPricingItems(xs)) != 3 {
		t.Errorf("IndexPricingItems() = %v", IndexPricingItems(xs))
	}
	if err := ValidatePricingItems(ActivePricingItems(xs)); err != nil {
		t.Errorf("ValidatePricingItems() = %v", err)
	}
}

func TestPricingItemNormalizeClamps(t *testing.T) {
	x := SamplePricingItem("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingDraftRoundTrip(t *testing.T) {
	x := SamplePricingDraft("pricingdraft-1", when)
	got, err := ParsePricingDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingDraft() = %v", err)
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

func TestPricingDraftParseRejectsBadInput(t *testing.T) {
	f := SamplePricingDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingDraft(f); err == nil {
		t.Error("ParsePricingDraft() accepted a record with no id")
	}
	f = SamplePricingDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePricingDraft(f); err == nil {
		t.Error("ParsePricingDraft() accepted a non-numeric attempts")
	}
}

func TestPricingDraftCollections(t *testing.T) {
	xs := []PricingDraft{SamplePricingDraft("a", when), SamplePricingDraft("bb", when), SamplePricingDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingDrafts(xs)); got != 2 {
		t.Errorf("ActivePricingDrafts() kept %d, want 2", got)
	}
	if got, want := SumPricingDrafts(xs), 1+3; got != want {
		t.Errorf("SumPricingDrafts() = %d, want %d", got, want)
	}
	SortPricingDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingDrafts() put %q first", xs[0].ID)
	}
	if len(IndexPricingDrafts(xs)) != 3 {
		t.Errorf("IndexPricingDrafts() = %v", IndexPricingDrafts(xs))
	}
	if err := ValidatePricingDrafts(ActivePricingDrafts(xs)); err != nil {
		t.Errorf("ValidatePricingDrafts() = %v", err)
	}
}

func TestPricingDraftNormalizeClamps(t *testing.T) {
	x := SamplePricingDraft("a", when)
	x.Category = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingRevisionRoundTrip(t *testing.T) {
	x := SamplePricingRevision("pricingrevision-1", when)
	got, err := ParsePricingRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingRevision() = %v", err)
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

func TestPricingRevisionParseRejectsBadInput(t *testing.T) {
	f := SamplePricingRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingRevision(f); err == nil {
		t.Error("ParsePricingRevision() accepted a record with no id")
	}
	f = SamplePricingRevision("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParsePricingRevision(f); err == nil {
		t.Error("ParsePricingRevision() accepted a non-numeric weight")
	}
}

func TestPricingRevisionCollections(t *testing.T) {
	xs := []PricingRevision{SamplePricingRevision("a", when), SamplePricingRevision("bb", when), SamplePricingRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingRevisions(xs)); got != 2 {
		t.Errorf("ActivePricingRevisions() kept %d, want 2", got)
	}
	if got, want := SumPricingRevisions(xs), 1+3; got != want {
		t.Errorf("SumPricingRevisions() = %d, want %d", got, want)
	}
	SortPricingRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingRevisions() put %q first", xs[0].ID)
	}
	if len(IndexPricingRevisions(xs)) != 3 {
		t.Errorf("IndexPricingRevisions() = %v", IndexPricingRevisions(xs))
	}
	if err := ValidatePricingRevisions(ActivePricingRevisions(xs)); err != nil {
		t.Errorf("ValidatePricingRevisions() = %v", err)
	}
}

func TestPricingRevisionNormalizeClamps(t *testing.T) {
	x := SamplePricingRevision("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingSnapshotRoundTrip(t *testing.T) {
	x := SamplePricingSnapshot("pricingsnapshot-1", when)
	got, err := ParsePricingSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingSnapshot() = %v", err)
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

func TestPricingSnapshotParseRejectsBadInput(t *testing.T) {
	f := SamplePricingSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingSnapshot(f); err == nil {
		t.Error("ParsePricingSnapshot() accepted a record with no id")
	}
	f = SamplePricingSnapshot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePricingSnapshot(f); err == nil {
		t.Error("ParsePricingSnapshot() accepted a non-numeric offset")
	}
}

func TestPricingSnapshotCollections(t *testing.T) {
	xs := []PricingSnapshot{SamplePricingSnapshot("a", when), SamplePricingSnapshot("bb", when), SamplePricingSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingSnapshots(xs)); got != 2 {
		t.Errorf("ActivePricingSnapshots() kept %d, want 2", got)
	}
	if got, want := SumPricingSnapshots(xs), 1+3; got != want {
		t.Errorf("SumPricingSnapshots() = %d, want %d", got, want)
	}
	SortPricingSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexPricingSnapshots(xs)) != 3 {
		t.Errorf("IndexPricingSnapshots() = %v", IndexPricingSnapshots(xs))
	}
	if err := ValidatePricingSnapshots(ActivePricingSnapshots(xs)); err != nil {
		t.Errorf("ValidatePricingSnapshots() = %v", err)
	}
}

func TestPricingSnapshotNormalizeClamps(t *testing.T) {
	x := SamplePricingSnapshot("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingBatchRoundTrip(t *testing.T) {
	x := SamplePricingBatch("pricingbatch-1", when)
	got, err := ParsePricingBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingBatch() = %v", err)
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

func TestPricingBatchParseRejectsBadInput(t *testing.T) {
	f := SamplePricingBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingBatch(f); err == nil {
		t.Error("ParsePricingBatch() accepted a record with no id")
	}
	f = SamplePricingBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePricingBatch(f); err == nil {
		t.Error("ParsePricingBatch() accepted a non-numeric version")
	}
}

func TestPricingBatchCollections(t *testing.T) {
	xs := []PricingBatch{SamplePricingBatch("a", when), SamplePricingBatch("bb", when), SamplePricingBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingBatchs(xs)); got != 2 {
		t.Errorf("ActivePricingBatchs() kept %d, want 2", got)
	}
	if got, want := SumPricingBatchs(xs), 1+3; got != want {
		t.Errorf("SumPricingBatchs() = %d, want %d", got, want)
	}
	SortPricingBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingBatchs() put %q first", xs[0].ID)
	}
	if len(IndexPricingBatchs(xs)) != 3 {
		t.Errorf("IndexPricingBatchs() = %v", IndexPricingBatchs(xs))
	}
	if err := ValidatePricingBatchs(ActivePricingBatchs(xs)); err != nil {
		t.Errorf("ValidatePricingBatchs() = %v", err)
	}
}

func TestPricingBatchNormalizeClamps(t *testing.T) {
	x := SamplePricingBatch("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingRequestRoundTrip(t *testing.T) {
	x := SamplePricingRequest("pricingrequest-1", when)
	got, err := ParsePricingRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingRequest() = %v", err)
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

func TestPricingRequestParseRejectsBadInput(t *testing.T) {
	f := SamplePricingRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingRequest(f); err == nil {
		t.Error("ParsePricingRequest() accepted a record with no id")
	}
	f = SamplePricingRequest("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePricingRequest(f); err == nil {
		t.Error("ParsePricingRequest() accepted a non-numeric retained")
	}
}

func TestPricingRequestCollections(t *testing.T) {
	xs := []PricingRequest{SamplePricingRequest("a", when), SamplePricingRequest("bb", when), SamplePricingRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingRequests(xs)); got != 2 {
		t.Errorf("ActivePricingRequests() kept %d, want 2", got)
	}
	if got, want := SumPricingRequests(xs), 1+3; got != want {
		t.Errorf("SumPricingRequests() = %d, want %d", got, want)
	}
	SortPricingRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingRequests() put %q first", xs[0].ID)
	}
	if len(IndexPricingRequests(xs)) != 3 {
		t.Errorf("IndexPricingRequests() = %v", IndexPricingRequests(xs))
	}
	if err := ValidatePricingRequests(ActivePricingRequests(xs)); err != nil {
		t.Errorf("ValidatePricingRequests() = %v", err)
	}
}

func TestPricingRequestNormalizeClamps(t *testing.T) {
	x := SamplePricingRequest("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingSummaryRoundTrip(t *testing.T) {
	x := SamplePricingSummary("pricingsummary-1", when)
	got, err := ParsePricingSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingSummary() = %v", err)
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

func TestPricingSummaryParseRejectsBadInput(t *testing.T) {
	f := SamplePricingSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingSummary(f); err == nil {
		t.Error("ParsePricingSummary() accepted a record with no id")
	}
	f = SamplePricingSummary("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePricingSummary(f); err == nil {
		t.Error("ParsePricingSummary() accepted a non-numeric count")
	}
}

func TestPricingSummaryCollections(t *testing.T) {
	xs := []PricingSummary{SamplePricingSummary("a", when), SamplePricingSummary("bb", when), SamplePricingSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingSummarys(xs)); got != 2 {
		t.Errorf("ActivePricingSummarys() kept %d, want 2", got)
	}
	if got, want := SumPricingSummarys(xs), 1+3; got != want {
		t.Errorf("SumPricingSummarys() = %d, want %d", got, want)
	}
	SortPricingSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingSummarys() put %q first", xs[0].ID)
	}
	if len(IndexPricingSummarys(xs)) != 3 {
		t.Errorf("IndexPricingSummarys() = %v", IndexPricingSummarys(xs))
	}
	if err := ValidatePricingSummarys(ActivePricingSummarys(xs)); err != nil {
		t.Errorf("ValidatePricingSummarys() = %v", err)
	}
}

func TestPricingSummaryNormalizeClamps(t *testing.T) {
	x := SamplePricingSummary("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingPolicyRoundTrip(t *testing.T) {
	x := SamplePricingPolicy("pricingpolicy-1", when)
	got, err := ParsePricingPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingPolicy() = %v", err)
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

func TestPricingPolicyParseRejectsBadInput(t *testing.T) {
	f := SamplePricingPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingPolicy(f); err == nil {
		t.Error("ParsePricingPolicy() accepted a record with no id")
	}
	f = SamplePricingPolicy("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePricingPolicy(f); err == nil {
		t.Error("ParsePricingPolicy() accepted a non-numeric capacity")
	}
}

func TestPricingPolicyCollections(t *testing.T) {
	xs := []PricingPolicy{SamplePricingPolicy("a", when), SamplePricingPolicy("bb", when), SamplePricingPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingPolicys(xs)); got != 2 {
		t.Errorf("ActivePricingPolicys() kept %d, want 2", got)
	}
	if got, want := SumPricingPolicys(xs), 1+3; got != want {
		t.Errorf("SumPricingPolicys() = %d, want %d", got, want)
	}
	SortPricingPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingPolicys() put %q first", xs[0].ID)
	}
	if len(IndexPricingPolicys(xs)) != 3 {
		t.Errorf("IndexPricingPolicys() = %v", IndexPricingPolicys(xs))
	}
	if err := ValidatePricingPolicys(ActivePricingPolicys(xs)); err != nil {
		t.Errorf("ValidatePricingPolicys() = %v", err)
	}
}

func TestPricingPolicyNormalizeClamps(t *testing.T) {
	x := SamplePricingPolicy("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingLinkRoundTrip(t *testing.T) {
	x := SamplePricingLink("pricinglink-1", when)
	got, err := ParsePricingLink(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingLink() = %v", err)
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

func TestPricingLinkParseRejectsBadInput(t *testing.T) {
	f := SamplePricingLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingLink(f); err == nil {
		t.Error("ParsePricingLink() accepted a record with no id")
	}
	f = SamplePricingLink("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePricingLink(f); err == nil {
		t.Error("ParsePricingLink() accepted a non-numeric attempts")
	}
}

func TestPricingLinkCollections(t *testing.T) {
	xs := []PricingLink{SamplePricingLink("a", when), SamplePricingLink("bb", when), SamplePricingLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingLinks(xs)); got != 2 {
		t.Errorf("ActivePricingLinks() kept %d, want 2", got)
	}
	if got, want := SumPricingLinks(xs), 1+3; got != want {
		t.Errorf("SumPricingLinks() = %d, want %d", got, want)
	}
	SortPricingLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingLinks() put %q first", xs[0].ID)
	}
	if len(IndexPricingLinks(xs)) != 3 {
		t.Errorf("IndexPricingLinks() = %v", IndexPricingLinks(xs))
	}
	if err := ValidatePricingLinks(ActivePricingLinks(xs)); err != nil {
		t.Errorf("ValidatePricingLinks() = %v", err)
	}
}

func TestPricingLinkNormalizeClamps(t *testing.T) {
	x := SamplePricingLink("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingNoteRoundTrip(t *testing.T) {
	x := SamplePricingNote("pricingnote-1", when)
	got, err := ParsePricingNote(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingNote() = %v", err)
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

func TestPricingNoteParseRejectsBadInput(t *testing.T) {
	f := SamplePricingNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingNote(f); err == nil {
		t.Error("ParsePricingNote() accepted a record with no id")
	}
	f = SamplePricingNote("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePricingNote(f); err == nil {
		t.Error("ParsePricingNote() accepted a non-numeric capacity")
	}
}

func TestPricingNoteCollections(t *testing.T) {
	xs := []PricingNote{SamplePricingNote("a", when), SamplePricingNote("bb", when), SamplePricingNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingNotes(xs)); got != 2 {
		t.Errorf("ActivePricingNotes() kept %d, want 2", got)
	}
	if got, want := SumPricingNotes(xs), 1+3; got != want {
		t.Errorf("SumPricingNotes() = %d, want %d", got, want)
	}
	SortPricingNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingNotes() put %q first", xs[0].ID)
	}
	if len(IndexPricingNotes(xs)) != 3 {
		t.Errorf("IndexPricingNotes() = %v", IndexPricingNotes(xs))
	}
	if err := ValidatePricingNotes(ActivePricingNotes(xs)); err != nil {
		t.Errorf("ValidatePricingNotes() = %v", err)
	}
}

func TestPricingNoteNormalizeClamps(t *testing.T) {
	x := SamplePricingNote("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingWindowRoundTrip(t *testing.T) {
	x := SamplePricingWindow("pricingwindow-1", when)
	got, err := ParsePricingWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingWindow() = %v", err)
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

func TestPricingWindowParseRejectsBadInput(t *testing.T) {
	f := SamplePricingWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingWindow(f); err == nil {
		t.Error("ParsePricingWindow() accepted a record with no id")
	}
	f = SamplePricingWindow("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePricingWindow(f); err == nil {
		t.Error("ParsePricingWindow() accepted a non-numeric offset")
	}
}

func TestPricingWindowCollections(t *testing.T) {
	xs := []PricingWindow{SamplePricingWindow("a", when), SamplePricingWindow("bb", when), SamplePricingWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingWindows(xs)); got != 2 {
		t.Errorf("ActivePricingWindows() kept %d, want 2", got)
	}
	if got, want := SumPricingWindows(xs), 1+3; got != want {
		t.Errorf("SumPricingWindows() = %d, want %d", got, want)
	}
	SortPricingWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingWindows() put %q first", xs[0].ID)
	}
	if len(IndexPricingWindows(xs)) != 3 {
		t.Errorf("IndexPricingWindows() = %v", IndexPricingWindows(xs))
	}
	if err := ValidatePricingWindows(ActivePricingWindows(xs)); err != nil {
		t.Errorf("ValidatePricingWindows() = %v", err)
	}
}

func TestPricingWindowNormalizeClamps(t *testing.T) {
	x := SamplePricingWindow("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingQuotaRoundTrip(t *testing.T) {
	x := SamplePricingQuota("pricingquota-1", when)
	got, err := ParsePricingQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingQuota() = %v", err)
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

func TestPricingQuotaParseRejectsBadInput(t *testing.T) {
	f := SamplePricingQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingQuota(f); err == nil {
		t.Error("ParsePricingQuota() accepted a record with no id")
	}
	f = SamplePricingQuota("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePricingQuota(f); err == nil {
		t.Error("ParsePricingQuota() accepted a non-numeric attempts")
	}
}

func TestPricingQuotaCollections(t *testing.T) {
	xs := []PricingQuota{SamplePricingQuota("a", when), SamplePricingQuota("bb", when), SamplePricingQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingQuotas(xs)); got != 2 {
		t.Errorf("ActivePricingQuotas() kept %d, want 2", got)
	}
	if got, want := SumPricingQuotas(xs), 1+3; got != want {
		t.Errorf("SumPricingQuotas() = %d, want %d", got, want)
	}
	SortPricingQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingQuotas() put %q first", xs[0].ID)
	}
	if len(IndexPricingQuotas(xs)) != 3 {
		t.Errorf("IndexPricingQuotas() = %v", IndexPricingQuotas(xs))
	}
	if err := ValidatePricingQuotas(ActivePricingQuotas(xs)); err != nil {
		t.Errorf("ValidatePricingQuotas() = %v", err)
	}
}

func TestPricingQuotaNormalizeClamps(t *testing.T) {
	x := SamplePricingQuota("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingBucketRoundTrip(t *testing.T) {
	x := SamplePricingBucket("pricingbucket-1", when)
	got, err := ParsePricingBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingBucket() = %v", err)
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

func TestPricingBucketParseRejectsBadInput(t *testing.T) {
	f := SamplePricingBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingBucket(f); err == nil {
		t.Error("ParsePricingBucket() accepted a record with no id")
	}
	f = SamplePricingBucket("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePricingBucket(f); err == nil {
		t.Error("ParsePricingBucket() accepted a non-numeric capacity")
	}
}

func TestPricingBucketCollections(t *testing.T) {
	xs := []PricingBucket{SamplePricingBucket("a", when), SamplePricingBucket("bb", when), SamplePricingBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingBuckets(xs)); got != 2 {
		t.Errorf("ActivePricingBuckets() kept %d, want 2", got)
	}
	if got, want := SumPricingBuckets(xs), 1+3; got != want {
		t.Errorf("SumPricingBuckets() = %d, want %d", got, want)
	}
	SortPricingBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingBuckets() put %q first", xs[0].ID)
	}
	if len(IndexPricingBuckets(xs)) != 3 {
		t.Errorf("IndexPricingBuckets() = %v", IndexPricingBuckets(xs))
	}
	if err := ValidatePricingBuckets(ActivePricingBuckets(xs)); err != nil {
		t.Errorf("ValidatePricingBuckets() = %v", err)
	}
}

func TestPricingBucketNormalizeClamps(t *testing.T) {
	x := SamplePricingBucket("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingLaneRoundTrip(t *testing.T) {
	x := SamplePricingLane("pricinglane-1", when)
	got, err := ParsePricingLane(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingLane() = %v", err)
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

func TestPricingLaneParseRejectsBadInput(t *testing.T) {
	f := SamplePricingLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingLane(f); err == nil {
		t.Error("ParsePricingLane() accepted a record with no id")
	}
	f = SamplePricingLane("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePricingLane(f); err == nil {
		t.Error("ParsePricingLane() accepted a non-numeric offset")
	}
}

func TestPricingLaneCollections(t *testing.T) {
	xs := []PricingLane{SamplePricingLane("a", when), SamplePricingLane("bb", when), SamplePricingLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingLanes(xs)); got != 2 {
		t.Errorf("ActivePricingLanes() kept %d, want 2", got)
	}
	if got, want := SumPricingLanes(xs), 1+3; got != want {
		t.Errorf("SumPricingLanes() = %d, want %d", got, want)
	}
	SortPricingLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingLanes() put %q first", xs[0].ID)
	}
	if len(IndexPricingLanes(xs)) != 3 {
		t.Errorf("IndexPricingLanes() = %v", IndexPricingLanes(xs))
	}
	if err := ValidatePricingLanes(ActivePricingLanes(xs)); err != nil {
		t.Errorf("ValidatePricingLanes() = %v", err)
	}
}

func TestPricingLaneNormalizeClamps(t *testing.T) {
	x := SamplePricingLane("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingSlotRoundTrip(t *testing.T) {
	x := SamplePricingSlot("pricingslot-1", when)
	got, err := ParsePricingSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingSlot() = %v", err)
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

func TestPricingSlotParseRejectsBadInput(t *testing.T) {
	f := SamplePricingSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingSlot(f); err == nil {
		t.Error("ParsePricingSlot() accepted a record with no id")
	}
	f = SamplePricingSlot("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePricingSlot(f); err == nil {
		t.Error("ParsePricingSlot() accepted a non-numeric capacity")
	}
}

func TestPricingSlotCollections(t *testing.T) {
	xs := []PricingSlot{SamplePricingSlot("a", when), SamplePricingSlot("bb", when), SamplePricingSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingSlots(xs)); got != 2 {
		t.Errorf("ActivePricingSlots() kept %d, want 2", got)
	}
	if got, want := SumPricingSlots(xs), 1+3; got != want {
		t.Errorf("SumPricingSlots() = %d, want %d", got, want)
	}
	SortPricingSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingSlots() put %q first", xs[0].ID)
	}
	if len(IndexPricingSlots(xs)) != 3 {
		t.Errorf("IndexPricingSlots() = %v", IndexPricingSlots(xs))
	}
	if err := ValidatePricingSlots(ActivePricingSlots(xs)); err != nil {
		t.Errorf("ValidatePricingSlots() = %v", err)
	}
}

func TestPricingSlotNormalizeClamps(t *testing.T) {
	x := SamplePricingSlot("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingTagRoundTrip(t *testing.T) {
	x := SamplePricingTag("pricingtag-1", when)
	got, err := ParsePricingTag(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingTag() = %v", err)
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

func TestPricingTagParseRejectsBadInput(t *testing.T) {
	f := SamplePricingTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingTag(f); err == nil {
		t.Error("ParsePricingTag() accepted a record with no id")
	}
	f = SamplePricingTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePricingTag(f); err == nil {
		t.Error("ParsePricingTag() accepted a non-numeric count")
	}
}

func TestPricingTagCollections(t *testing.T) {
	xs := []PricingTag{SamplePricingTag("a", when), SamplePricingTag("bb", when), SamplePricingTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingTags(xs)); got != 2 {
		t.Errorf("ActivePricingTags() kept %d, want 2", got)
	}
	if got, want := SumPricingTags(xs), 1+3; got != want {
		t.Errorf("SumPricingTags() = %d, want %d", got, want)
	}
	SortPricingTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingTags() put %q first", xs[0].ID)
	}
	if len(IndexPricingTags(xs)) != 3 {
		t.Errorf("IndexPricingTags() = %v", IndexPricingTags(xs))
	}
	if err := ValidatePricingTags(ActivePricingTags(xs)); err != nil {
		t.Errorf("ValidatePricingTags() = %v", err)
	}
}

func TestPricingTagNormalizeClamps(t *testing.T) {
	x := SamplePricingTag("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingRuleRoundTrip(t *testing.T) {
	x := SamplePricingRule("pricingrule-1", when)
	got, err := ParsePricingRule(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingRule() = %v", err)
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

func TestPricingRuleParseRejectsBadInput(t *testing.T) {
	f := SamplePricingRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingRule(f); err == nil {
		t.Error("ParsePricingRule() accepted a record with no id")
	}
	f = SamplePricingRule("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePricingRule(f); err == nil {
		t.Error("ParsePricingRule() accepted a non-numeric attempts")
	}
}

func TestPricingRuleCollections(t *testing.T) {
	xs := []PricingRule{SamplePricingRule("a", when), SamplePricingRule("bb", when), SamplePricingRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingRules(xs)); got != 2 {
		t.Errorf("ActivePricingRules() kept %d, want 2", got)
	}
	if got, want := SumPricingRules(xs), 1+3; got != want {
		t.Errorf("SumPricingRules() = %d, want %d", got, want)
	}
	SortPricingRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingRules() put %q first", xs[0].ID)
	}
	if len(IndexPricingRules(xs)) != 3 {
		t.Errorf("IndexPricingRules() = %v", IndexPricingRules(xs))
	}
	if err := ValidatePricingRules(ActivePricingRules(xs)); err != nil {
		t.Errorf("ValidatePricingRules() = %v", err)
	}
}

func TestPricingRuleNormalizeClamps(t *testing.T) {
	x := SamplePricingRule("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingHoldRoundTrip(t *testing.T) {
	x := SamplePricingHold("pricinghold-1", when)
	got, err := ParsePricingHold(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingHold() = %v", err)
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

func TestPricingHoldParseRejectsBadInput(t *testing.T) {
	f := SamplePricingHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingHold(f); err == nil {
		t.Error("ParsePricingHold() accepted a record with no id")
	}
	f = SamplePricingHold("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePricingHold(f); err == nil {
		t.Error("ParsePricingHold() accepted a non-numeric count")
	}
}

func TestPricingHoldCollections(t *testing.T) {
	xs := []PricingHold{SamplePricingHold("a", when), SamplePricingHold("bb", when), SamplePricingHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingHolds(xs)); got != 2 {
		t.Errorf("ActivePricingHolds() kept %d, want 2", got)
	}
	if got, want := SumPricingHolds(xs), 1+3; got != want {
		t.Errorf("SumPricingHolds() = %d, want %d", got, want)
	}
	SortPricingHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingHolds() put %q first", xs[0].ID)
	}
	if len(IndexPricingHolds(xs)) != 3 {
		t.Errorf("IndexPricingHolds() = %v", IndexPricingHolds(xs))
	}
	if err := ValidatePricingHolds(ActivePricingHolds(xs)); err != nil {
		t.Errorf("ValidatePricingHolds() = %v", err)
	}
}

func TestPricingHoldNormalizeClamps(t *testing.T) {
	x := SamplePricingHold("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingClaimRoundTrip(t *testing.T) {
	x := SamplePricingClaim("pricingclaim-1", when)
	got, err := ParsePricingClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingClaim() = %v", err)
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

func TestPricingClaimParseRejectsBadInput(t *testing.T) {
	f := SamplePricingClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingClaim(f); err == nil {
		t.Error("ParsePricingClaim() accepted a record with no id")
	}
	f = SamplePricingClaim("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePricingClaim(f); err == nil {
		t.Error("ParsePricingClaim() accepted a non-numeric version")
	}
}

func TestPricingClaimCollections(t *testing.T) {
	xs := []PricingClaim{SamplePricingClaim("a", when), SamplePricingClaim("bb", when), SamplePricingClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingClaims(xs)); got != 2 {
		t.Errorf("ActivePricingClaims() kept %d, want 2", got)
	}
	if got, want := SumPricingClaims(xs), 1+3; got != want {
		t.Errorf("SumPricingClaims() = %d, want %d", got, want)
	}
	SortPricingClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingClaims() put %q first", xs[0].ID)
	}
	if len(IndexPricingClaims(xs)) != 3 {
		t.Errorf("IndexPricingClaims() = %v", IndexPricingClaims(xs))
	}
	if err := ValidatePricingClaims(ActivePricingClaims(xs)); err != nil {
		t.Errorf("ValidatePricingClaims() = %v", err)
	}
}

func TestPricingClaimNormalizeClamps(t *testing.T) {
	x := SamplePricingClaim("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPricingCursorRoundTrip(t *testing.T) {
	x := SamplePricingCursor("pricingcursor-1", when)
	got, err := ParsePricingCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParsePricingCursor() = %v", err)
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

func TestPricingCursorParseRejectsBadInput(t *testing.T) {
	f := SamplePricingCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePricingCursor(f); err == nil {
		t.Error("ParsePricingCursor() accepted a record with no id")
	}
	f = SamplePricingCursor("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePricingCursor(f); err == nil {
		t.Error("ParsePricingCursor() accepted a non-numeric count")
	}
}

func TestPricingCursorCollections(t *testing.T) {
	xs := []PricingCursor{SamplePricingCursor("a", when), SamplePricingCursor("bb", when), SamplePricingCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePricingCursors(xs)); got != 2 {
		t.Errorf("ActivePricingCursors() kept %d, want 2", got)
	}
	if got, want := SumPricingCursors(xs), 1+3; got != want {
		t.Errorf("SumPricingCursors() = %d, want %d", got, want)
	}
	SortPricingCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPricingCursors() put %q first", xs[0].ID)
	}
	if len(IndexPricingCursors(xs)) != 3 {
		t.Errorf("IndexPricingCursors() = %v", IndexPricingCursors(xs))
	}
	if err := ValidatePricingCursors(ActivePricingCursors(xs)); err != nil {
		t.Errorf("ValidatePricingCursors() = %v", err)
	}
}

func TestPricingCursorNormalizeClamps(t *testing.T) {
	x := SamplePricingCursor("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
