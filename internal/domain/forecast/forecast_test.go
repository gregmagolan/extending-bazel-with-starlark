package forecast

import (
	"math/rand/v2"
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
		if !strings.HasPrefix(Prefixes[i], "forecast/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleForecastRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestForecastRecordRoundTrip(t *testing.T) {
	x := SampleForecastRecord("forecastrecord-1", when)
	got, err := ParseForecastRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastRecord() = %v", err)
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

func TestForecastRecordParseRejectsBadInput(t *testing.T) {
	f := SampleForecastRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastRecord(f); err == nil {
		t.Error("ParseForecastRecord() accepted a record with no id")
	}
	f = SampleForecastRecord("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseForecastRecord(f); err == nil {
		t.Error("ParseForecastRecord() accepted a non-numeric offset")
	}
}

func TestForecastRecordCollections(t *testing.T) {
	xs := []ForecastRecord{SampleForecastRecord("a", when), SampleForecastRecord("bb", when), SampleForecastRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastRecords(xs)); got != 2 {
		t.Errorf("ActiveForecastRecords() kept %d, want 2", got)
	}
	if got, want := SumForecastRecords(xs), 1+3; got != want {
		t.Errorf("SumForecastRecords() = %d, want %d", got, want)
	}
	SortForecastRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastRecords() put %q first", xs[0].ID)
	}
	if len(IndexForecastRecords(xs)) != 3 {
		t.Errorf("IndexForecastRecords() = %v", IndexForecastRecords(xs))
	}
	if err := ValidateForecastRecords(ActiveForecastRecords(xs)); err != nil {
		t.Errorf("ValidateForecastRecords() = %v", err)
	}
}

func TestForecastRecordNormalizeClamps(t *testing.T) {
	x := SampleForecastRecord("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastEntryRoundTrip(t *testing.T) {
	x := SampleForecastEntry("forecastentry-1", when)
	got, err := ParseForecastEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastEntry() = %v", err)
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

func TestForecastEntryParseRejectsBadInput(t *testing.T) {
	f := SampleForecastEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastEntry(f); err == nil {
		t.Error("ParseForecastEntry() accepted a record with no id")
	}
	f = SampleForecastEntry("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseForecastEntry(f); err == nil {
		t.Error("ParseForecastEntry() accepted a non-numeric capacity")
	}
}

func TestForecastEntryCollections(t *testing.T) {
	xs := []ForecastEntry{SampleForecastEntry("a", when), SampleForecastEntry("bb", when), SampleForecastEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastEntrys(xs)); got != 2 {
		t.Errorf("ActiveForecastEntrys() kept %d, want 2", got)
	}
	if got, want := SumForecastEntrys(xs), 1+3; got != want {
		t.Errorf("SumForecastEntrys() = %d, want %d", got, want)
	}
	SortForecastEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastEntrys() put %q first", xs[0].ID)
	}
	if len(IndexForecastEntrys(xs)) != 3 {
		t.Errorf("IndexForecastEntrys() = %v", IndexForecastEntrys(xs))
	}
	if err := ValidateForecastEntrys(ActiveForecastEntrys(xs)); err != nil {
		t.Errorf("ValidateForecastEntrys() = %v", err)
	}
}

func TestForecastEntryNormalizeClamps(t *testing.T) {
	x := SampleForecastEntry("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastItemRoundTrip(t *testing.T) {
	x := SampleForecastItem("forecastitem-1", when)
	got, err := ParseForecastItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastItem() = %v", err)
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

func TestForecastItemParseRejectsBadInput(t *testing.T) {
	f := SampleForecastItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastItem(f); err == nil {
		t.Error("ParseForecastItem() accepted a record with no id")
	}
	f = SampleForecastItem("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseForecastItem(f); err == nil {
		t.Error("ParseForecastItem() accepted a non-numeric weight")
	}
}

func TestForecastItemCollections(t *testing.T) {
	xs := []ForecastItem{SampleForecastItem("a", when), SampleForecastItem("bb", when), SampleForecastItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastItems(xs)); got != 2 {
		t.Errorf("ActiveForecastItems() kept %d, want 2", got)
	}
	if got, want := SumForecastItems(xs), 1+3; got != want {
		t.Errorf("SumForecastItems() = %d, want %d", got, want)
	}
	SortForecastItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastItems() put %q first", xs[0].ID)
	}
	if len(IndexForecastItems(xs)) != 3 {
		t.Errorf("IndexForecastItems() = %v", IndexForecastItems(xs))
	}
	if err := ValidateForecastItems(ActiveForecastItems(xs)); err != nil {
		t.Errorf("ValidateForecastItems() = %v", err)
	}
}

func TestForecastItemNormalizeClamps(t *testing.T) {
	x := SampleForecastItem("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastDraftRoundTrip(t *testing.T) {
	x := SampleForecastDraft("forecastdraft-1", when)
	got, err := ParseForecastDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastDraft() = %v", err)
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

func TestForecastDraftParseRejectsBadInput(t *testing.T) {
	f := SampleForecastDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastDraft(f); err == nil {
		t.Error("ParseForecastDraft() accepted a record with no id")
	}
	f = SampleForecastDraft("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseForecastDraft(f); err == nil {
		t.Error("ParseForecastDraft() accepted a non-numeric retained")
	}
}

func TestForecastDraftCollections(t *testing.T) {
	xs := []ForecastDraft{SampleForecastDraft("a", when), SampleForecastDraft("bb", when), SampleForecastDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastDrafts(xs)); got != 2 {
		t.Errorf("ActiveForecastDrafts() kept %d, want 2", got)
	}
	if got, want := SumForecastDrafts(xs), 1+3; got != want {
		t.Errorf("SumForecastDrafts() = %d, want %d", got, want)
	}
	SortForecastDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastDrafts() put %q first", xs[0].ID)
	}
	if len(IndexForecastDrafts(xs)) != 3 {
		t.Errorf("IndexForecastDrafts() = %v", IndexForecastDrafts(xs))
	}
	if err := ValidateForecastDrafts(ActiveForecastDrafts(xs)); err != nil {
		t.Errorf("ValidateForecastDrafts() = %v", err)
	}
}

func TestForecastDraftNormalizeClamps(t *testing.T) {
	x := SampleForecastDraft("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastRevisionRoundTrip(t *testing.T) {
	x := SampleForecastRevision("forecastrevision-1", when)
	got, err := ParseForecastRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastRevision() = %v", err)
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

func TestForecastRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleForecastRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastRevision(f); err == nil {
		t.Error("ParseForecastRevision() accepted a record with no id")
	}
	f = SampleForecastRevision("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseForecastRevision(f); err == nil {
		t.Error("ParseForecastRevision() accepted a non-numeric attempts")
	}
}

func TestForecastRevisionCollections(t *testing.T) {
	xs := []ForecastRevision{SampleForecastRevision("a", when), SampleForecastRevision("bb", when), SampleForecastRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastRevisions(xs)); got != 2 {
		t.Errorf("ActiveForecastRevisions() kept %d, want 2", got)
	}
	if got, want := SumForecastRevisions(xs), 1+3; got != want {
		t.Errorf("SumForecastRevisions() = %d, want %d", got, want)
	}
	SortForecastRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastRevisions() put %q first", xs[0].ID)
	}
	if len(IndexForecastRevisions(xs)) != 3 {
		t.Errorf("IndexForecastRevisions() = %v", IndexForecastRevisions(xs))
	}
	if err := ValidateForecastRevisions(ActiveForecastRevisions(xs)); err != nil {
		t.Errorf("ValidateForecastRevisions() = %v", err)
	}
}

func TestForecastRevisionNormalizeClamps(t *testing.T) {
	x := SampleForecastRevision("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastSnapshotRoundTrip(t *testing.T) {
	x := SampleForecastSnapshot("forecastsnapshot-1", when)
	got, err := ParseForecastSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastSnapshot() = %v", err)
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

func TestForecastSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleForecastSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastSnapshot(f); err == nil {
		t.Error("ParseForecastSnapshot() accepted a record with no id")
	}
	f = SampleForecastSnapshot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseForecastSnapshot(f); err == nil {
		t.Error("ParseForecastSnapshot() accepted a non-numeric weight")
	}
}

func TestForecastSnapshotCollections(t *testing.T) {
	xs := []ForecastSnapshot{SampleForecastSnapshot("a", when), SampleForecastSnapshot("bb", when), SampleForecastSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastSnapshots(xs)); got != 2 {
		t.Errorf("ActiveForecastSnapshots() kept %d, want 2", got)
	}
	if got, want := SumForecastSnapshots(xs), 1+3; got != want {
		t.Errorf("SumForecastSnapshots() = %d, want %d", got, want)
	}
	SortForecastSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexForecastSnapshots(xs)) != 3 {
		t.Errorf("IndexForecastSnapshots() = %v", IndexForecastSnapshots(xs))
	}
	if err := ValidateForecastSnapshots(ActiveForecastSnapshots(xs)); err != nil {
		t.Errorf("ValidateForecastSnapshots() = %v", err)
	}
}

func TestForecastSnapshotNormalizeClamps(t *testing.T) {
	x := SampleForecastSnapshot("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastBatchRoundTrip(t *testing.T) {
	x := SampleForecastBatch("forecastbatch-1", when)
	got, err := ParseForecastBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastBatch() = %v", err)
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

func TestForecastBatchParseRejectsBadInput(t *testing.T) {
	f := SampleForecastBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastBatch(f); err == nil {
		t.Error("ParseForecastBatch() accepted a record with no id")
	}
	f = SampleForecastBatch("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseForecastBatch(f); err == nil {
		t.Error("ParseForecastBatch() accepted a non-numeric priority")
	}
}

func TestForecastBatchCollections(t *testing.T) {
	xs := []ForecastBatch{SampleForecastBatch("a", when), SampleForecastBatch("bb", when), SampleForecastBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastBatchs(xs)); got != 2 {
		t.Errorf("ActiveForecastBatchs() kept %d, want 2", got)
	}
	if got, want := SumForecastBatchs(xs), 1+3; got != want {
		t.Errorf("SumForecastBatchs() = %d, want %d", got, want)
	}
	SortForecastBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastBatchs() put %q first", xs[0].ID)
	}
	if len(IndexForecastBatchs(xs)) != 3 {
		t.Errorf("IndexForecastBatchs() = %v", IndexForecastBatchs(xs))
	}
	if err := ValidateForecastBatchs(ActiveForecastBatchs(xs)); err != nil {
		t.Errorf("ValidateForecastBatchs() = %v", err)
	}
}

func TestForecastBatchNormalizeClamps(t *testing.T) {
	x := SampleForecastBatch("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastRequestRoundTrip(t *testing.T) {
	x := SampleForecastRequest("forecastrequest-1", when)
	got, err := ParseForecastRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastRequest() = %v", err)
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

func TestForecastRequestParseRejectsBadInput(t *testing.T) {
	f := SampleForecastRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastRequest(f); err == nil {
		t.Error("ParseForecastRequest() accepted a record with no id")
	}
	f = SampleForecastRequest("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseForecastRequest(f); err == nil {
		t.Error("ParseForecastRequest() accepted a non-numeric weight")
	}
}

func TestForecastRequestCollections(t *testing.T) {
	xs := []ForecastRequest{SampleForecastRequest("a", when), SampleForecastRequest("bb", when), SampleForecastRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastRequests(xs)); got != 2 {
		t.Errorf("ActiveForecastRequests() kept %d, want 2", got)
	}
	if got, want := SumForecastRequests(xs), 1+3; got != want {
		t.Errorf("SumForecastRequests() = %d, want %d", got, want)
	}
	SortForecastRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastRequests() put %q first", xs[0].ID)
	}
	if len(IndexForecastRequests(xs)) != 3 {
		t.Errorf("IndexForecastRequests() = %v", IndexForecastRequests(xs))
	}
	if err := ValidateForecastRequests(ActiveForecastRequests(xs)); err != nil {
		t.Errorf("ValidateForecastRequests() = %v", err)
	}
}

func TestForecastRequestNormalizeClamps(t *testing.T) {
	x := SampleForecastRequest("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastSummaryRoundTrip(t *testing.T) {
	x := SampleForecastSummary("forecastsummary-1", when)
	got, err := ParseForecastSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastSummary() = %v", err)
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

func TestForecastSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleForecastSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastSummary(f); err == nil {
		t.Error("ParseForecastSummary() accepted a record with no id")
	}
	f = SampleForecastSummary("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseForecastSummary(f); err == nil {
		t.Error("ParseForecastSummary() accepted a non-numeric offset")
	}
}

func TestForecastSummaryCollections(t *testing.T) {
	xs := []ForecastSummary{SampleForecastSummary("a", when), SampleForecastSummary("bb", when), SampleForecastSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastSummarys(xs)); got != 2 {
		t.Errorf("ActiveForecastSummarys() kept %d, want 2", got)
	}
	if got, want := SumForecastSummarys(xs), 1+3; got != want {
		t.Errorf("SumForecastSummarys() = %d, want %d", got, want)
	}
	SortForecastSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastSummarys() put %q first", xs[0].ID)
	}
	if len(IndexForecastSummarys(xs)) != 3 {
		t.Errorf("IndexForecastSummarys() = %v", IndexForecastSummarys(xs))
	}
	if err := ValidateForecastSummarys(ActiveForecastSummarys(xs)); err != nil {
		t.Errorf("ValidateForecastSummarys() = %v", err)
	}
}

func TestForecastSummaryNormalizeClamps(t *testing.T) {
	x := SampleForecastSummary("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastPolicyRoundTrip(t *testing.T) {
	x := SampleForecastPolicy("forecastpolicy-1", when)
	got, err := ParseForecastPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastPolicy() = %v", err)
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

func TestForecastPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleForecastPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastPolicy(f); err == nil {
		t.Error("ParseForecastPolicy() accepted a record with no id")
	}
	f = SampleForecastPolicy("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseForecastPolicy(f); err == nil {
		t.Error("ParseForecastPolicy() accepted a non-numeric capacity")
	}
}

func TestForecastPolicyCollections(t *testing.T) {
	xs := []ForecastPolicy{SampleForecastPolicy("a", when), SampleForecastPolicy("bb", when), SampleForecastPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastPolicys(xs)); got != 2 {
		t.Errorf("ActiveForecastPolicys() kept %d, want 2", got)
	}
	if got, want := SumForecastPolicys(xs), 1+3; got != want {
		t.Errorf("SumForecastPolicys() = %d, want %d", got, want)
	}
	SortForecastPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastPolicys() put %q first", xs[0].ID)
	}
	if len(IndexForecastPolicys(xs)) != 3 {
		t.Errorf("IndexForecastPolicys() = %v", IndexForecastPolicys(xs))
	}
	if err := ValidateForecastPolicys(ActiveForecastPolicys(xs)); err != nil {
		t.Errorf("ValidateForecastPolicys() = %v", err)
	}
}

func TestForecastPolicyNormalizeClamps(t *testing.T) {
	x := SampleForecastPolicy("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastLinkRoundTrip(t *testing.T) {
	x := SampleForecastLink("forecastlink-1", when)
	got, err := ParseForecastLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastLink() = %v", err)
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

func TestForecastLinkParseRejectsBadInput(t *testing.T) {
	f := SampleForecastLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastLink(f); err == nil {
		t.Error("ParseForecastLink() accepted a record with no id")
	}
	f = SampleForecastLink("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseForecastLink(f); err == nil {
		t.Error("ParseForecastLink() accepted a non-numeric priority")
	}
}

func TestForecastLinkCollections(t *testing.T) {
	xs := []ForecastLink{SampleForecastLink("a", when), SampleForecastLink("bb", when), SampleForecastLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastLinks(xs)); got != 2 {
		t.Errorf("ActiveForecastLinks() kept %d, want 2", got)
	}
	if got, want := SumForecastLinks(xs), 1+3; got != want {
		t.Errorf("SumForecastLinks() = %d, want %d", got, want)
	}
	SortForecastLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastLinks() put %q first", xs[0].ID)
	}
	if len(IndexForecastLinks(xs)) != 3 {
		t.Errorf("IndexForecastLinks() = %v", IndexForecastLinks(xs))
	}
	if err := ValidateForecastLinks(ActiveForecastLinks(xs)); err != nil {
		t.Errorf("ValidateForecastLinks() = %v", err)
	}
}

func TestForecastLinkNormalizeClamps(t *testing.T) {
	x := SampleForecastLink("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastNoteRoundTrip(t *testing.T) {
	x := SampleForecastNote("forecastnote-1", when)
	got, err := ParseForecastNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastNote() = %v", err)
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

func TestForecastNoteParseRejectsBadInput(t *testing.T) {
	f := SampleForecastNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastNote(f); err == nil {
		t.Error("ParseForecastNote() accepted a record with no id")
	}
	f = SampleForecastNote("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseForecastNote(f); err == nil {
		t.Error("ParseForecastNote() accepted a non-numeric offset")
	}
}

func TestForecastNoteCollections(t *testing.T) {
	xs := []ForecastNote{SampleForecastNote("a", when), SampleForecastNote("bb", when), SampleForecastNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastNotes(xs)); got != 2 {
		t.Errorf("ActiveForecastNotes() kept %d, want 2", got)
	}
	if got, want := SumForecastNotes(xs), 1+3; got != want {
		t.Errorf("SumForecastNotes() = %d, want %d", got, want)
	}
	SortForecastNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastNotes() put %q first", xs[0].ID)
	}
	if len(IndexForecastNotes(xs)) != 3 {
		t.Errorf("IndexForecastNotes() = %v", IndexForecastNotes(xs))
	}
	if err := ValidateForecastNotes(ActiveForecastNotes(xs)); err != nil {
		t.Errorf("ValidateForecastNotes() = %v", err)
	}
}

func TestForecastNoteNormalizeClamps(t *testing.T) {
	x := SampleForecastNote("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastWindowRoundTrip(t *testing.T) {
	x := SampleForecastWindow("forecastwindow-1", when)
	got, err := ParseForecastWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastWindow() = %v", err)
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

func TestForecastWindowParseRejectsBadInput(t *testing.T) {
	f := SampleForecastWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastWindow(f); err == nil {
		t.Error("ParseForecastWindow() accepted a record with no id")
	}
	f = SampleForecastWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseForecastWindow(f); err == nil {
		t.Error("ParseForecastWindow() accepted a non-numeric priority")
	}
}

func TestForecastWindowCollections(t *testing.T) {
	xs := []ForecastWindow{SampleForecastWindow("a", when), SampleForecastWindow("bb", when), SampleForecastWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastWindows(xs)); got != 2 {
		t.Errorf("ActiveForecastWindows() kept %d, want 2", got)
	}
	if got, want := SumForecastWindows(xs), 1+3; got != want {
		t.Errorf("SumForecastWindows() = %d, want %d", got, want)
	}
	SortForecastWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastWindows() put %q first", xs[0].ID)
	}
	if len(IndexForecastWindows(xs)) != 3 {
		t.Errorf("IndexForecastWindows() = %v", IndexForecastWindows(xs))
	}
	if err := ValidateForecastWindows(ActiveForecastWindows(xs)); err != nil {
		t.Errorf("ValidateForecastWindows() = %v", err)
	}
}

func TestForecastWindowNormalizeClamps(t *testing.T) {
	x := SampleForecastWindow("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastQuotaRoundTrip(t *testing.T) {
	x := SampleForecastQuota("forecastquota-1", when)
	got, err := ParseForecastQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastQuota() = %v", err)
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

func TestForecastQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleForecastQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastQuota(f); err == nil {
		t.Error("ParseForecastQuota() accepted a record with no id")
	}
	f = SampleForecastQuota("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseForecastQuota(f); err == nil {
		t.Error("ParseForecastQuota() accepted a non-numeric retained")
	}
}

func TestForecastQuotaCollections(t *testing.T) {
	xs := []ForecastQuota{SampleForecastQuota("a", when), SampleForecastQuota("bb", when), SampleForecastQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastQuotas(xs)); got != 2 {
		t.Errorf("ActiveForecastQuotas() kept %d, want 2", got)
	}
	if got, want := SumForecastQuotas(xs), 1+3; got != want {
		t.Errorf("SumForecastQuotas() = %d, want %d", got, want)
	}
	SortForecastQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastQuotas() put %q first", xs[0].ID)
	}
	if len(IndexForecastQuotas(xs)) != 3 {
		t.Errorf("IndexForecastQuotas() = %v", IndexForecastQuotas(xs))
	}
	if err := ValidateForecastQuotas(ActiveForecastQuotas(xs)); err != nil {
		t.Errorf("ValidateForecastQuotas() = %v", err)
	}
}

func TestForecastQuotaNormalizeClamps(t *testing.T) {
	x := SampleForecastQuota("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastBucketRoundTrip(t *testing.T) {
	x := SampleForecastBucket("forecastbucket-1", when)
	got, err := ParseForecastBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastBucket() = %v", err)
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

func TestForecastBucketParseRejectsBadInput(t *testing.T) {
	f := SampleForecastBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastBucket(f); err == nil {
		t.Error("ParseForecastBucket() accepted a record with no id")
	}
	f = SampleForecastBucket("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseForecastBucket(f); err == nil {
		t.Error("ParseForecastBucket() accepted a non-numeric capacity")
	}
}

func TestForecastBucketCollections(t *testing.T) {
	xs := []ForecastBucket{SampleForecastBucket("a", when), SampleForecastBucket("bb", when), SampleForecastBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastBuckets(xs)); got != 2 {
		t.Errorf("ActiveForecastBuckets() kept %d, want 2", got)
	}
	if got, want := SumForecastBuckets(xs), 1+3; got != want {
		t.Errorf("SumForecastBuckets() = %d, want %d", got, want)
	}
	SortForecastBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastBuckets() put %q first", xs[0].ID)
	}
	if len(IndexForecastBuckets(xs)) != 3 {
		t.Errorf("IndexForecastBuckets() = %v", IndexForecastBuckets(xs))
	}
	if err := ValidateForecastBuckets(ActiveForecastBuckets(xs)); err != nil {
		t.Errorf("ValidateForecastBuckets() = %v", err)
	}
}

func TestForecastBucketNormalizeClamps(t *testing.T) {
	x := SampleForecastBucket("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastLaneRoundTrip(t *testing.T) {
	x := SampleForecastLane("forecastlane-1", when)
	got, err := ParseForecastLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastLane() = %v", err)
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

func TestForecastLaneParseRejectsBadInput(t *testing.T) {
	f := SampleForecastLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastLane(f); err == nil {
		t.Error("ParseForecastLane() accepted a record with no id")
	}
	f = SampleForecastLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseForecastLane(f); err == nil {
		t.Error("ParseForecastLane() accepted a non-numeric capacity")
	}
}

func TestForecastLaneCollections(t *testing.T) {
	xs := []ForecastLane{SampleForecastLane("a", when), SampleForecastLane("bb", when), SampleForecastLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastLanes(xs)); got != 2 {
		t.Errorf("ActiveForecastLanes() kept %d, want 2", got)
	}
	if got, want := SumForecastLanes(xs), 1+3; got != want {
		t.Errorf("SumForecastLanes() = %d, want %d", got, want)
	}
	SortForecastLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastLanes() put %q first", xs[0].ID)
	}
	if len(IndexForecastLanes(xs)) != 3 {
		t.Errorf("IndexForecastLanes() = %v", IndexForecastLanes(xs))
	}
	if err := ValidateForecastLanes(ActiveForecastLanes(xs)); err != nil {
		t.Errorf("ValidateForecastLanes() = %v", err)
	}
}

func TestForecastLaneNormalizeClamps(t *testing.T) {
	x := SampleForecastLane("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastSlotRoundTrip(t *testing.T) {
	x := SampleForecastSlot("forecastslot-1", when)
	got, err := ParseForecastSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastSlot() = %v", err)
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

func TestForecastSlotParseRejectsBadInput(t *testing.T) {
	f := SampleForecastSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastSlot(f); err == nil {
		t.Error("ParseForecastSlot() accepted a record with no id")
	}
	f = SampleForecastSlot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseForecastSlot(f); err == nil {
		t.Error("ParseForecastSlot() accepted a non-numeric retained")
	}
}

func TestForecastSlotCollections(t *testing.T) {
	xs := []ForecastSlot{SampleForecastSlot("a", when), SampleForecastSlot("bb", when), SampleForecastSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastSlots(xs)); got != 2 {
		t.Errorf("ActiveForecastSlots() kept %d, want 2", got)
	}
	if got, want := SumForecastSlots(xs), 1+3; got != want {
		t.Errorf("SumForecastSlots() = %d, want %d", got, want)
	}
	SortForecastSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastSlots() put %q first", xs[0].ID)
	}
	if len(IndexForecastSlots(xs)) != 3 {
		t.Errorf("IndexForecastSlots() = %v", IndexForecastSlots(xs))
	}
	if err := ValidateForecastSlots(ActiveForecastSlots(xs)); err != nil {
		t.Errorf("ValidateForecastSlots() = %v", err)
	}
}

func TestForecastSlotNormalizeClamps(t *testing.T) {
	x := SampleForecastSlot("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastTagRoundTrip(t *testing.T) {
	x := SampleForecastTag("forecasttag-1", when)
	got, err := ParseForecastTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastTag() = %v", err)
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

func TestForecastTagParseRejectsBadInput(t *testing.T) {
	f := SampleForecastTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastTag(f); err == nil {
		t.Error("ParseForecastTag() accepted a record with no id")
	}
	f = SampleForecastTag("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseForecastTag(f); err == nil {
		t.Error("ParseForecastTag() accepted a non-numeric priority")
	}
}

func TestForecastTagCollections(t *testing.T) {
	xs := []ForecastTag{SampleForecastTag("a", when), SampleForecastTag("bb", when), SampleForecastTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastTags(xs)); got != 2 {
		t.Errorf("ActiveForecastTags() kept %d, want 2", got)
	}
	if got, want := SumForecastTags(xs), 1+3; got != want {
		t.Errorf("SumForecastTags() = %d, want %d", got, want)
	}
	SortForecastTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastTags() put %q first", xs[0].ID)
	}
	if len(IndexForecastTags(xs)) != 3 {
		t.Errorf("IndexForecastTags() = %v", IndexForecastTags(xs))
	}
	if err := ValidateForecastTags(ActiveForecastTags(xs)); err != nil {
		t.Errorf("ValidateForecastTags() = %v", err)
	}
}

func TestForecastTagNormalizeClamps(t *testing.T) {
	x := SampleForecastTag("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastRuleRoundTrip(t *testing.T) {
	x := SampleForecastRule("forecastrule-1", when)
	got, err := ParseForecastRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastRule() = %v", err)
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

func TestForecastRuleParseRejectsBadInput(t *testing.T) {
	f := SampleForecastRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastRule(f); err == nil {
		t.Error("ParseForecastRule() accepted a record with no id")
	}
	f = SampleForecastRule("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseForecastRule(f); err == nil {
		t.Error("ParseForecastRule() accepted a non-numeric priority")
	}
}

func TestForecastRuleCollections(t *testing.T) {
	xs := []ForecastRule{SampleForecastRule("a", when), SampleForecastRule("bb", when), SampleForecastRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastRules(xs)); got != 2 {
		t.Errorf("ActiveForecastRules() kept %d, want 2", got)
	}
	if got, want := SumForecastRules(xs), 1+3; got != want {
		t.Errorf("SumForecastRules() = %d, want %d", got, want)
	}
	SortForecastRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastRules() put %q first", xs[0].ID)
	}
	if len(IndexForecastRules(xs)) != 3 {
		t.Errorf("IndexForecastRules() = %v", IndexForecastRules(xs))
	}
	if err := ValidateForecastRules(ActiveForecastRules(xs)); err != nil {
		t.Errorf("ValidateForecastRules() = %v", err)
	}
}

func TestForecastRuleNormalizeClamps(t *testing.T) {
	x := SampleForecastRule("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastHoldRoundTrip(t *testing.T) {
	x := SampleForecastHold("forecasthold-1", when)
	got, err := ParseForecastHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastHold() = %v", err)
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

func TestForecastHoldParseRejectsBadInput(t *testing.T) {
	f := SampleForecastHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastHold(f); err == nil {
		t.Error("ParseForecastHold() accepted a record with no id")
	}
	f = SampleForecastHold("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseForecastHold(f); err == nil {
		t.Error("ParseForecastHold() accepted a non-numeric retained")
	}
}

func TestForecastHoldCollections(t *testing.T) {
	xs := []ForecastHold{SampleForecastHold("a", when), SampleForecastHold("bb", when), SampleForecastHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastHolds(xs)); got != 2 {
		t.Errorf("ActiveForecastHolds() kept %d, want 2", got)
	}
	if got, want := SumForecastHolds(xs), 1+3; got != want {
		t.Errorf("SumForecastHolds() = %d, want %d", got, want)
	}
	SortForecastHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastHolds() put %q first", xs[0].ID)
	}
	if len(IndexForecastHolds(xs)) != 3 {
		t.Errorf("IndexForecastHolds() = %v", IndexForecastHolds(xs))
	}
	if err := ValidateForecastHolds(ActiveForecastHolds(xs)); err != nil {
		t.Errorf("ValidateForecastHolds() = %v", err)
	}
}

func TestForecastHoldNormalizeClamps(t *testing.T) {
	x := SampleForecastHold("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastClaimRoundTrip(t *testing.T) {
	x := SampleForecastClaim("forecastclaim-1", when)
	got, err := ParseForecastClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastClaim() = %v", err)
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

func TestForecastClaimParseRejectsBadInput(t *testing.T) {
	f := SampleForecastClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastClaim(f); err == nil {
		t.Error("ParseForecastClaim() accepted a record with no id")
	}
	f = SampleForecastClaim("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseForecastClaim(f); err == nil {
		t.Error("ParseForecastClaim() accepted a non-numeric offset")
	}
}

func TestForecastClaimCollections(t *testing.T) {
	xs := []ForecastClaim{SampleForecastClaim("a", when), SampleForecastClaim("bb", when), SampleForecastClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastClaims(xs)); got != 2 {
		t.Errorf("ActiveForecastClaims() kept %d, want 2", got)
	}
	if got, want := SumForecastClaims(xs), 1+3; got != want {
		t.Errorf("SumForecastClaims() = %d, want %d", got, want)
	}
	SortForecastClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastClaims() put %q first", xs[0].ID)
	}
	if len(IndexForecastClaims(xs)) != 3 {
		t.Errorf("IndexForecastClaims() = %v", IndexForecastClaims(xs))
	}
	if err := ValidateForecastClaims(ActiveForecastClaims(xs)); err != nil {
		t.Errorf("ValidateForecastClaims() = %v", err)
	}
}

func TestForecastClaimNormalizeClamps(t *testing.T) {
	x := SampleForecastClaim("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestForecastCursorRoundTrip(t *testing.T) {
	x := SampleForecastCursor("forecastcursor-1", when)
	got, err := ParseForecastCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseForecastCursor() = %v", err)
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

func TestForecastCursorParseRejectsBadInput(t *testing.T) {
	f := SampleForecastCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseForecastCursor(f); err == nil {
		t.Error("ParseForecastCursor() accepted a record with no id")
	}
	f = SampleForecastCursor("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseForecastCursor(f); err == nil {
		t.Error("ParseForecastCursor() accepted a non-numeric attempts")
	}
}

func TestForecastCursorCollections(t *testing.T) {
	xs := []ForecastCursor{SampleForecastCursor("a", when), SampleForecastCursor("bb", when), SampleForecastCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveForecastCursors(xs)); got != 2 {
		t.Errorf("ActiveForecastCursors() kept %d, want 2", got)
	}
	if got, want := SumForecastCursors(xs), 1+3; got != want {
		t.Errorf("SumForecastCursors() = %d, want %d", got, want)
	}
	SortForecastCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortForecastCursors() put %q first", xs[0].ID)
	}
	if len(IndexForecastCursors(xs)) != 3 {
		t.Errorf("IndexForecastCursors() = %v", IndexForecastCursors(xs))
	}
	if err := ValidateForecastCursors(ActiveForecastCursors(xs)); err != nil {
		t.Errorf("ValidateForecastCursors() = %v", err)
	}
}

func TestForecastCursorNormalizeClamps(t *testing.T) {
	x := SampleForecastCursor("a", when)
	x.Category = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

// Deliberately flaky — the course uses this in the flake-detection step. Not a bug.
//
// Fails on close to one run in three. The draw is unbiased; deriving it from
// time.Now().UnixNano() instead skews high on macOS, where the clock is
// quantized to a microsecond and process start times are not independent.
//
// It lives here, in a leaf package, rather than somewhere the earlier steps
// demo against — a flake in lib/queue would turn those steps red about half
// the time.
func TestHorizonRollover(t *testing.T) {
	if rand.IntN(3) == 0 {
		t.Fatalf("horizon rolled over before the window closed")
	}
	if len(Schemas) == 0 {
		t.Fatal("no schemas")
	}
}
