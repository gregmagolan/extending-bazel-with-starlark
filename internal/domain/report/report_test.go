package report

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
		if !strings.HasPrefix(Prefixes[i], "report/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleReportRecord("ok", when).Fields()); err != nil {
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

func TestReportRecordRoundTrip(t *testing.T) {
	x := SampleReportRecord("reportrecord-1", when)
	got, err := ParseReportRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportRecord() = %v", err)
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

func TestReportRecordParseRejectsBadInput(t *testing.T) {
	f := SampleReportRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportRecord(f); err == nil {
		t.Error("ParseReportRecord() accepted a record with no id")
	}
	f = SampleReportRecord("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseReportRecord(f); err == nil {
		t.Error("ParseReportRecord() accepted a non-numeric offset")
	}
}

func TestReportRecordCollections(t *testing.T) {
	xs := []ReportRecord{SampleReportRecord("a", when), SampleReportRecord("bb", when), SampleReportRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportRecords(xs)); got != 2 {
		t.Errorf("ActiveReportRecords() kept %d, want 2", got)
	}
	if got, want := SumReportRecords(xs), 1+3; got != want {
		t.Errorf("SumReportRecords() = %d, want %d", got, want)
	}
	SortReportRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportRecords() put %q first", xs[0].ID)
	}
	if len(IndexReportRecords(xs)) != 3 {
		t.Errorf("IndexReportRecords() = %v", IndexReportRecords(xs))
	}
	if err := ValidateReportRecords(ActiveReportRecords(xs)); err != nil {
		t.Errorf("ValidateReportRecords() = %v", err)
	}
}

func TestReportRecordNormalizeClamps(t *testing.T) {
	x := SampleReportRecord("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportEntryRoundTrip(t *testing.T) {
	x := SampleReportEntry("reportentry-1", when)
	got, err := ParseReportEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportEntry() = %v", err)
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

func TestReportEntryParseRejectsBadInput(t *testing.T) {
	f := SampleReportEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportEntry(f); err == nil {
		t.Error("ParseReportEntry() accepted a record with no id")
	}
	f = SampleReportEntry("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseReportEntry(f); err == nil {
		t.Error("ParseReportEntry() accepted a non-numeric version")
	}
}

func TestReportEntryCollections(t *testing.T) {
	xs := []ReportEntry{SampleReportEntry("a", when), SampleReportEntry("bb", when), SampleReportEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportEntrys(xs)); got != 2 {
		t.Errorf("ActiveReportEntrys() kept %d, want 2", got)
	}
	if got, want := SumReportEntrys(xs), 1+3; got != want {
		t.Errorf("SumReportEntrys() = %d, want %d", got, want)
	}
	SortReportEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportEntrys() put %q first", xs[0].ID)
	}
	if len(IndexReportEntrys(xs)) != 3 {
		t.Errorf("IndexReportEntrys() = %v", IndexReportEntrys(xs))
	}
	if err := ValidateReportEntrys(ActiveReportEntrys(xs)); err != nil {
		t.Errorf("ValidateReportEntrys() = %v", err)
	}
}

func TestReportEntryNormalizeClamps(t *testing.T) {
	x := SampleReportEntry("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportItemRoundTrip(t *testing.T) {
	x := SampleReportItem("reportitem-1", when)
	got, err := ParseReportItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportItem() = %v", err)
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

func TestReportItemParseRejectsBadInput(t *testing.T) {
	f := SampleReportItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportItem(f); err == nil {
		t.Error("ParseReportItem() accepted a record with no id")
	}
	f = SampleReportItem("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseReportItem(f); err == nil {
		t.Error("ParseReportItem() accepted a non-numeric attempts")
	}
}

func TestReportItemCollections(t *testing.T) {
	xs := []ReportItem{SampleReportItem("a", when), SampleReportItem("bb", when), SampleReportItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportItems(xs)); got != 2 {
		t.Errorf("ActiveReportItems() kept %d, want 2", got)
	}
	if got, want := SumReportItems(xs), 1+3; got != want {
		t.Errorf("SumReportItems() = %d, want %d", got, want)
	}
	SortReportItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportItems() put %q first", xs[0].ID)
	}
	if len(IndexReportItems(xs)) != 3 {
		t.Errorf("IndexReportItems() = %v", IndexReportItems(xs))
	}
	if err := ValidateReportItems(ActiveReportItems(xs)); err != nil {
		t.Errorf("ValidateReportItems() = %v", err)
	}
}

func TestReportItemNormalizeClamps(t *testing.T) {
	x := SampleReportItem("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportDraftRoundTrip(t *testing.T) {
	x := SampleReportDraft("reportdraft-1", when)
	got, err := ParseReportDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportDraft() = %v", err)
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

func TestReportDraftParseRejectsBadInput(t *testing.T) {
	f := SampleReportDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportDraft(f); err == nil {
		t.Error("ParseReportDraft() accepted a record with no id")
	}
	f = SampleReportDraft("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseReportDraft(f); err == nil {
		t.Error("ParseReportDraft() accepted a non-numeric offset")
	}
}

func TestReportDraftCollections(t *testing.T) {
	xs := []ReportDraft{SampleReportDraft("a", when), SampleReportDraft("bb", when), SampleReportDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportDrafts(xs)); got != 2 {
		t.Errorf("ActiveReportDrafts() kept %d, want 2", got)
	}
	if got, want := SumReportDrafts(xs), 1+3; got != want {
		t.Errorf("SumReportDrafts() = %d, want %d", got, want)
	}
	SortReportDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportDrafts() put %q first", xs[0].ID)
	}
	if len(IndexReportDrafts(xs)) != 3 {
		t.Errorf("IndexReportDrafts() = %v", IndexReportDrafts(xs))
	}
	if err := ValidateReportDrafts(ActiveReportDrafts(xs)); err != nil {
		t.Errorf("ValidateReportDrafts() = %v", err)
	}
}

func TestReportDraftNormalizeClamps(t *testing.T) {
	x := SampleReportDraft("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportRevisionRoundTrip(t *testing.T) {
	x := SampleReportRevision("reportrevision-1", when)
	got, err := ParseReportRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportRevision() = %v", err)
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

func TestReportRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleReportRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportRevision(f); err == nil {
		t.Error("ParseReportRevision() accepted a record with no id")
	}
	f = SampleReportRevision("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseReportRevision(f); err == nil {
		t.Error("ParseReportRevision() accepted a non-numeric count")
	}
}

func TestReportRevisionCollections(t *testing.T) {
	xs := []ReportRevision{SampleReportRevision("a", when), SampleReportRevision("bb", when), SampleReportRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportRevisions(xs)); got != 2 {
		t.Errorf("ActiveReportRevisions() kept %d, want 2", got)
	}
	if got, want := SumReportRevisions(xs), 1+3; got != want {
		t.Errorf("SumReportRevisions() = %d, want %d", got, want)
	}
	SortReportRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportRevisions() put %q first", xs[0].ID)
	}
	if len(IndexReportRevisions(xs)) != 3 {
		t.Errorf("IndexReportRevisions() = %v", IndexReportRevisions(xs))
	}
	if err := ValidateReportRevisions(ActiveReportRevisions(xs)); err != nil {
		t.Errorf("ValidateReportRevisions() = %v", err)
	}
}

func TestReportRevisionNormalizeClamps(t *testing.T) {
	x := SampleReportRevision("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportSnapshotRoundTrip(t *testing.T) {
	x := SampleReportSnapshot("reportsnapshot-1", when)
	got, err := ParseReportSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportSnapshot() = %v", err)
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

func TestReportSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleReportSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportSnapshot(f); err == nil {
		t.Error("ParseReportSnapshot() accepted a record with no id")
	}
	f = SampleReportSnapshot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseReportSnapshot(f); err == nil {
		t.Error("ParseReportSnapshot() accepted a non-numeric retained")
	}
}

func TestReportSnapshotCollections(t *testing.T) {
	xs := []ReportSnapshot{SampleReportSnapshot("a", when), SampleReportSnapshot("bb", when), SampleReportSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportSnapshots(xs)); got != 2 {
		t.Errorf("ActiveReportSnapshots() kept %d, want 2", got)
	}
	if got, want := SumReportSnapshots(xs), 1+3; got != want {
		t.Errorf("SumReportSnapshots() = %d, want %d", got, want)
	}
	SortReportSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexReportSnapshots(xs)) != 3 {
		t.Errorf("IndexReportSnapshots() = %v", IndexReportSnapshots(xs))
	}
	if err := ValidateReportSnapshots(ActiveReportSnapshots(xs)); err != nil {
		t.Errorf("ValidateReportSnapshots() = %v", err)
	}
}

func TestReportSnapshotNormalizeClamps(t *testing.T) {
	x := SampleReportSnapshot("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportBatchRoundTrip(t *testing.T) {
	x := SampleReportBatch("reportbatch-1", when)
	got, err := ParseReportBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportBatch() = %v", err)
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

func TestReportBatchParseRejectsBadInput(t *testing.T) {
	f := SampleReportBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportBatch(f); err == nil {
		t.Error("ParseReportBatch() accepted a record with no id")
	}
	f = SampleReportBatch("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseReportBatch(f); err == nil {
		t.Error("ParseReportBatch() accepted a non-numeric count")
	}
}

func TestReportBatchCollections(t *testing.T) {
	xs := []ReportBatch{SampleReportBatch("a", when), SampleReportBatch("bb", when), SampleReportBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportBatchs(xs)); got != 2 {
		t.Errorf("ActiveReportBatchs() kept %d, want 2", got)
	}
	if got, want := SumReportBatchs(xs), 1+3; got != want {
		t.Errorf("SumReportBatchs() = %d, want %d", got, want)
	}
	SortReportBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportBatchs() put %q first", xs[0].ID)
	}
	if len(IndexReportBatchs(xs)) != 3 {
		t.Errorf("IndexReportBatchs() = %v", IndexReportBatchs(xs))
	}
	if err := ValidateReportBatchs(ActiveReportBatchs(xs)); err != nil {
		t.Errorf("ValidateReportBatchs() = %v", err)
	}
}

func TestReportBatchNormalizeClamps(t *testing.T) {
	x := SampleReportBatch("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportRequestRoundTrip(t *testing.T) {
	x := SampleReportRequest("reportrequest-1", when)
	got, err := ParseReportRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportRequest() = %v", err)
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

func TestReportRequestParseRejectsBadInput(t *testing.T) {
	f := SampleReportRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportRequest(f); err == nil {
		t.Error("ParseReportRequest() accepted a record with no id")
	}
	f = SampleReportRequest("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseReportRequest(f); err == nil {
		t.Error("ParseReportRequest() accepted a non-numeric attempts")
	}
}

func TestReportRequestCollections(t *testing.T) {
	xs := []ReportRequest{SampleReportRequest("a", when), SampleReportRequest("bb", when), SampleReportRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportRequests(xs)); got != 2 {
		t.Errorf("ActiveReportRequests() kept %d, want 2", got)
	}
	if got, want := SumReportRequests(xs), 1+3; got != want {
		t.Errorf("SumReportRequests() = %d, want %d", got, want)
	}
	SortReportRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportRequests() put %q first", xs[0].ID)
	}
	if len(IndexReportRequests(xs)) != 3 {
		t.Errorf("IndexReportRequests() = %v", IndexReportRequests(xs))
	}
	if err := ValidateReportRequests(ActiveReportRequests(xs)); err != nil {
		t.Errorf("ValidateReportRequests() = %v", err)
	}
}

func TestReportRequestNormalizeClamps(t *testing.T) {
	x := SampleReportRequest("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportSummaryRoundTrip(t *testing.T) {
	x := SampleReportSummary("reportsummary-1", when)
	got, err := ParseReportSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportSummary() = %v", err)
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

func TestReportSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleReportSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportSummary(f); err == nil {
		t.Error("ParseReportSummary() accepted a record with no id")
	}
	f = SampleReportSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseReportSummary(f); err == nil {
		t.Error("ParseReportSummary() accepted a non-numeric capacity")
	}
}

func TestReportSummaryCollections(t *testing.T) {
	xs := []ReportSummary{SampleReportSummary("a", when), SampleReportSummary("bb", when), SampleReportSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportSummarys(xs)); got != 2 {
		t.Errorf("ActiveReportSummarys() kept %d, want 2", got)
	}
	if got, want := SumReportSummarys(xs), 1+3; got != want {
		t.Errorf("SumReportSummarys() = %d, want %d", got, want)
	}
	SortReportSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportSummarys() put %q first", xs[0].ID)
	}
	if len(IndexReportSummarys(xs)) != 3 {
		t.Errorf("IndexReportSummarys() = %v", IndexReportSummarys(xs))
	}
	if err := ValidateReportSummarys(ActiveReportSummarys(xs)); err != nil {
		t.Errorf("ValidateReportSummarys() = %v", err)
	}
}

func TestReportSummaryNormalizeClamps(t *testing.T) {
	x := SampleReportSummary("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportPolicyRoundTrip(t *testing.T) {
	x := SampleReportPolicy("reportpolicy-1", when)
	got, err := ParseReportPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportPolicy() = %v", err)
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

func TestReportPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleReportPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportPolicy(f); err == nil {
		t.Error("ParseReportPolicy() accepted a record with no id")
	}
	f = SampleReportPolicy("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseReportPolicy(f); err == nil {
		t.Error("ParseReportPolicy() accepted a non-numeric priority")
	}
}

func TestReportPolicyCollections(t *testing.T) {
	xs := []ReportPolicy{SampleReportPolicy("a", when), SampleReportPolicy("bb", when), SampleReportPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportPolicys(xs)); got != 2 {
		t.Errorf("ActiveReportPolicys() kept %d, want 2", got)
	}
	if got, want := SumReportPolicys(xs), 1+3; got != want {
		t.Errorf("SumReportPolicys() = %d, want %d", got, want)
	}
	SortReportPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportPolicys() put %q first", xs[0].ID)
	}
	if len(IndexReportPolicys(xs)) != 3 {
		t.Errorf("IndexReportPolicys() = %v", IndexReportPolicys(xs))
	}
	if err := ValidateReportPolicys(ActiveReportPolicys(xs)); err != nil {
		t.Errorf("ValidateReportPolicys() = %v", err)
	}
}

func TestReportPolicyNormalizeClamps(t *testing.T) {
	x := SampleReportPolicy("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportLinkRoundTrip(t *testing.T) {
	x := SampleReportLink("reportlink-1", when)
	got, err := ParseReportLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportLink() = %v", err)
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

func TestReportLinkParseRejectsBadInput(t *testing.T) {
	f := SampleReportLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportLink(f); err == nil {
		t.Error("ParseReportLink() accepted a record with no id")
	}
	f = SampleReportLink("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseReportLink(f); err == nil {
		t.Error("ParseReportLink() accepted a non-numeric offset")
	}
}

func TestReportLinkCollections(t *testing.T) {
	xs := []ReportLink{SampleReportLink("a", when), SampleReportLink("bb", when), SampleReportLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportLinks(xs)); got != 2 {
		t.Errorf("ActiveReportLinks() kept %d, want 2", got)
	}
	if got, want := SumReportLinks(xs), 1+3; got != want {
		t.Errorf("SumReportLinks() = %d, want %d", got, want)
	}
	SortReportLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportLinks() put %q first", xs[0].ID)
	}
	if len(IndexReportLinks(xs)) != 3 {
		t.Errorf("IndexReportLinks() = %v", IndexReportLinks(xs))
	}
	if err := ValidateReportLinks(ActiveReportLinks(xs)); err != nil {
		t.Errorf("ValidateReportLinks() = %v", err)
	}
}

func TestReportLinkNormalizeClamps(t *testing.T) {
	x := SampleReportLink("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportNoteRoundTrip(t *testing.T) {
	x := SampleReportNote("reportnote-1", when)
	got, err := ParseReportNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportNote() = %v", err)
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

func TestReportNoteParseRejectsBadInput(t *testing.T) {
	f := SampleReportNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportNote(f); err == nil {
		t.Error("ParseReportNote() accepted a record with no id")
	}
	f = SampleReportNote("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseReportNote(f); err == nil {
		t.Error("ParseReportNote() accepted a non-numeric count")
	}
}

func TestReportNoteCollections(t *testing.T) {
	xs := []ReportNote{SampleReportNote("a", when), SampleReportNote("bb", when), SampleReportNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportNotes(xs)); got != 2 {
		t.Errorf("ActiveReportNotes() kept %d, want 2", got)
	}
	if got, want := SumReportNotes(xs), 1+3; got != want {
		t.Errorf("SumReportNotes() = %d, want %d", got, want)
	}
	SortReportNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportNotes() put %q first", xs[0].ID)
	}
	if len(IndexReportNotes(xs)) != 3 {
		t.Errorf("IndexReportNotes() = %v", IndexReportNotes(xs))
	}
	if err := ValidateReportNotes(ActiveReportNotes(xs)); err != nil {
		t.Errorf("ValidateReportNotes() = %v", err)
	}
}

func TestReportNoteNormalizeClamps(t *testing.T) {
	x := SampleReportNote("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportWindowRoundTrip(t *testing.T) {
	x := SampleReportWindow("reportwindow-1", when)
	got, err := ParseReportWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportWindow() = %v", err)
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

func TestReportWindowParseRejectsBadInput(t *testing.T) {
	f := SampleReportWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportWindow(f); err == nil {
		t.Error("ParseReportWindow() accepted a record with no id")
	}
	f = SampleReportWindow("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseReportWindow(f); err == nil {
		t.Error("ParseReportWindow() accepted a non-numeric attempts")
	}
}

func TestReportWindowCollections(t *testing.T) {
	xs := []ReportWindow{SampleReportWindow("a", when), SampleReportWindow("bb", when), SampleReportWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportWindows(xs)); got != 2 {
		t.Errorf("ActiveReportWindows() kept %d, want 2", got)
	}
	if got, want := SumReportWindows(xs), 1+3; got != want {
		t.Errorf("SumReportWindows() = %d, want %d", got, want)
	}
	SortReportWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportWindows() put %q first", xs[0].ID)
	}
	if len(IndexReportWindows(xs)) != 3 {
		t.Errorf("IndexReportWindows() = %v", IndexReportWindows(xs))
	}
	if err := ValidateReportWindows(ActiveReportWindows(xs)); err != nil {
		t.Errorf("ValidateReportWindows() = %v", err)
	}
}

func TestReportWindowNormalizeClamps(t *testing.T) {
	x := SampleReportWindow("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportQuotaRoundTrip(t *testing.T) {
	x := SampleReportQuota("reportquota-1", when)
	got, err := ParseReportQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportQuota() = %v", err)
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

func TestReportQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleReportQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportQuota(f); err == nil {
		t.Error("ParseReportQuota() accepted a record with no id")
	}
	f = SampleReportQuota("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseReportQuota(f); err == nil {
		t.Error("ParseReportQuota() accepted a non-numeric count")
	}
}

func TestReportQuotaCollections(t *testing.T) {
	xs := []ReportQuota{SampleReportQuota("a", when), SampleReportQuota("bb", when), SampleReportQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportQuotas(xs)); got != 2 {
		t.Errorf("ActiveReportQuotas() kept %d, want 2", got)
	}
	if got, want := SumReportQuotas(xs), 1+3; got != want {
		t.Errorf("SumReportQuotas() = %d, want %d", got, want)
	}
	SortReportQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportQuotas() put %q first", xs[0].ID)
	}
	if len(IndexReportQuotas(xs)) != 3 {
		t.Errorf("IndexReportQuotas() = %v", IndexReportQuotas(xs))
	}
	if err := ValidateReportQuotas(ActiveReportQuotas(xs)); err != nil {
		t.Errorf("ValidateReportQuotas() = %v", err)
	}
}

func TestReportQuotaNormalizeClamps(t *testing.T) {
	x := SampleReportQuota("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportBucketRoundTrip(t *testing.T) {
	x := SampleReportBucket("reportbucket-1", when)
	got, err := ParseReportBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportBucket() = %v", err)
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

func TestReportBucketParseRejectsBadInput(t *testing.T) {
	f := SampleReportBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportBucket(f); err == nil {
		t.Error("ParseReportBucket() accepted a record with no id")
	}
	f = SampleReportBucket("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseReportBucket(f); err == nil {
		t.Error("ParseReportBucket() accepted a non-numeric capacity")
	}
}

func TestReportBucketCollections(t *testing.T) {
	xs := []ReportBucket{SampleReportBucket("a", when), SampleReportBucket("bb", when), SampleReportBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportBuckets(xs)); got != 2 {
		t.Errorf("ActiveReportBuckets() kept %d, want 2", got)
	}
	if got, want := SumReportBuckets(xs), 1+3; got != want {
		t.Errorf("SumReportBuckets() = %d, want %d", got, want)
	}
	SortReportBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportBuckets() put %q first", xs[0].ID)
	}
	if len(IndexReportBuckets(xs)) != 3 {
		t.Errorf("IndexReportBuckets() = %v", IndexReportBuckets(xs))
	}
	if err := ValidateReportBuckets(ActiveReportBuckets(xs)); err != nil {
		t.Errorf("ValidateReportBuckets() = %v", err)
	}
}

func TestReportBucketNormalizeClamps(t *testing.T) {
	x := SampleReportBucket("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportLaneRoundTrip(t *testing.T) {
	x := SampleReportLane("reportlane-1", when)
	got, err := ParseReportLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportLane() = %v", err)
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

func TestReportLaneParseRejectsBadInput(t *testing.T) {
	f := SampleReportLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportLane(f); err == nil {
		t.Error("ParseReportLane() accepted a record with no id")
	}
	f = SampleReportLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseReportLane(f); err == nil {
		t.Error("ParseReportLane() accepted a non-numeric priority")
	}
}

func TestReportLaneCollections(t *testing.T) {
	xs := []ReportLane{SampleReportLane("a", when), SampleReportLane("bb", when), SampleReportLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportLanes(xs)); got != 2 {
		t.Errorf("ActiveReportLanes() kept %d, want 2", got)
	}
	if got, want := SumReportLanes(xs), 1+3; got != want {
		t.Errorf("SumReportLanes() = %d, want %d", got, want)
	}
	SortReportLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportLanes() put %q first", xs[0].ID)
	}
	if len(IndexReportLanes(xs)) != 3 {
		t.Errorf("IndexReportLanes() = %v", IndexReportLanes(xs))
	}
	if err := ValidateReportLanes(ActiveReportLanes(xs)); err != nil {
		t.Errorf("ValidateReportLanes() = %v", err)
	}
}

func TestReportLaneNormalizeClamps(t *testing.T) {
	x := SampleReportLane("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportSlotRoundTrip(t *testing.T) {
	x := SampleReportSlot("reportslot-1", when)
	got, err := ParseReportSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportSlot() = %v", err)
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

func TestReportSlotParseRejectsBadInput(t *testing.T) {
	f := SampleReportSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportSlot(f); err == nil {
		t.Error("ParseReportSlot() accepted a record with no id")
	}
	f = SampleReportSlot("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseReportSlot(f); err == nil {
		t.Error("ParseReportSlot() accepted a non-numeric version")
	}
}

func TestReportSlotCollections(t *testing.T) {
	xs := []ReportSlot{SampleReportSlot("a", when), SampleReportSlot("bb", when), SampleReportSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportSlots(xs)); got != 2 {
		t.Errorf("ActiveReportSlots() kept %d, want 2", got)
	}
	if got, want := SumReportSlots(xs), 1+3; got != want {
		t.Errorf("SumReportSlots() = %d, want %d", got, want)
	}
	SortReportSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportSlots() put %q first", xs[0].ID)
	}
	if len(IndexReportSlots(xs)) != 3 {
		t.Errorf("IndexReportSlots() = %v", IndexReportSlots(xs))
	}
	if err := ValidateReportSlots(ActiveReportSlots(xs)); err != nil {
		t.Errorf("ValidateReportSlots() = %v", err)
	}
}

func TestReportSlotNormalizeClamps(t *testing.T) {
	x := SampleReportSlot("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportTagRoundTrip(t *testing.T) {
	x := SampleReportTag("reporttag-1", when)
	got, err := ParseReportTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportTag() = %v", err)
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

func TestReportTagParseRejectsBadInput(t *testing.T) {
	f := SampleReportTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportTag(f); err == nil {
		t.Error("ParseReportTag() accepted a record with no id")
	}
	f = SampleReportTag("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseReportTag(f); err == nil {
		t.Error("ParseReportTag() accepted a non-numeric attempts")
	}
}

func TestReportTagCollections(t *testing.T) {
	xs := []ReportTag{SampleReportTag("a", when), SampleReportTag("bb", when), SampleReportTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportTags(xs)); got != 2 {
		t.Errorf("ActiveReportTags() kept %d, want 2", got)
	}
	if got, want := SumReportTags(xs), 1+3; got != want {
		t.Errorf("SumReportTags() = %d, want %d", got, want)
	}
	SortReportTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportTags() put %q first", xs[0].ID)
	}
	if len(IndexReportTags(xs)) != 3 {
		t.Errorf("IndexReportTags() = %v", IndexReportTags(xs))
	}
	if err := ValidateReportTags(ActiveReportTags(xs)); err != nil {
		t.Errorf("ValidateReportTags() = %v", err)
	}
}

func TestReportTagNormalizeClamps(t *testing.T) {
	x := SampleReportTag("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportRuleRoundTrip(t *testing.T) {
	x := SampleReportRule("reportrule-1", when)
	got, err := ParseReportRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportRule() = %v", err)
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

func TestReportRuleParseRejectsBadInput(t *testing.T) {
	f := SampleReportRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportRule(f); err == nil {
		t.Error("ParseReportRule() accepted a record with no id")
	}
	f = SampleReportRule("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseReportRule(f); err == nil {
		t.Error("ParseReportRule() accepted a non-numeric capacity")
	}
}

func TestReportRuleCollections(t *testing.T) {
	xs := []ReportRule{SampleReportRule("a", when), SampleReportRule("bb", when), SampleReportRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportRules(xs)); got != 2 {
		t.Errorf("ActiveReportRules() kept %d, want 2", got)
	}
	if got, want := SumReportRules(xs), 1+3; got != want {
		t.Errorf("SumReportRules() = %d, want %d", got, want)
	}
	SortReportRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportRules() put %q first", xs[0].ID)
	}
	if len(IndexReportRules(xs)) != 3 {
		t.Errorf("IndexReportRules() = %v", IndexReportRules(xs))
	}
	if err := ValidateReportRules(ActiveReportRules(xs)); err != nil {
		t.Errorf("ValidateReportRules() = %v", err)
	}
}

func TestReportRuleNormalizeClamps(t *testing.T) {
	x := SampleReportRule("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportHoldRoundTrip(t *testing.T) {
	x := SampleReportHold("reporthold-1", when)
	got, err := ParseReportHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportHold() = %v", err)
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

func TestReportHoldParseRejectsBadInput(t *testing.T) {
	f := SampleReportHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportHold(f); err == nil {
		t.Error("ParseReportHold() accepted a record with no id")
	}
	f = SampleReportHold("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseReportHold(f); err == nil {
		t.Error("ParseReportHold() accepted a non-numeric count")
	}
}

func TestReportHoldCollections(t *testing.T) {
	xs := []ReportHold{SampleReportHold("a", when), SampleReportHold("bb", when), SampleReportHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportHolds(xs)); got != 2 {
		t.Errorf("ActiveReportHolds() kept %d, want 2", got)
	}
	if got, want := SumReportHolds(xs), 1+3; got != want {
		t.Errorf("SumReportHolds() = %d, want %d", got, want)
	}
	SortReportHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportHolds() put %q first", xs[0].ID)
	}
	if len(IndexReportHolds(xs)) != 3 {
		t.Errorf("IndexReportHolds() = %v", IndexReportHolds(xs))
	}
	if err := ValidateReportHolds(ActiveReportHolds(xs)); err != nil {
		t.Errorf("ValidateReportHolds() = %v", err)
	}
}

func TestReportHoldNormalizeClamps(t *testing.T) {
	x := SampleReportHold("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportClaimRoundTrip(t *testing.T) {
	x := SampleReportClaim("reportclaim-1", when)
	got, err := ParseReportClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportClaim() = %v", err)
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

func TestReportClaimParseRejectsBadInput(t *testing.T) {
	f := SampleReportClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportClaim(f); err == nil {
		t.Error("ParseReportClaim() accepted a record with no id")
	}
	f = SampleReportClaim("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseReportClaim(f); err == nil {
		t.Error("ParseReportClaim() accepted a non-numeric weight")
	}
}

func TestReportClaimCollections(t *testing.T) {
	xs := []ReportClaim{SampleReportClaim("a", when), SampleReportClaim("bb", when), SampleReportClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportClaims(xs)); got != 2 {
		t.Errorf("ActiveReportClaims() kept %d, want 2", got)
	}
	if got, want := SumReportClaims(xs), 1+3; got != want {
		t.Errorf("SumReportClaims() = %d, want %d", got, want)
	}
	SortReportClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportClaims() put %q first", xs[0].ID)
	}
	if len(IndexReportClaims(xs)) != 3 {
		t.Errorf("IndexReportClaims() = %v", IndexReportClaims(xs))
	}
	if err := ValidateReportClaims(ActiveReportClaims(xs)); err != nil {
		t.Errorf("ValidateReportClaims() = %v", err)
	}
}

func TestReportClaimNormalizeClamps(t *testing.T) {
	x := SampleReportClaim("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestReportCursorRoundTrip(t *testing.T) {
	x := SampleReportCursor("reportcursor-1", when)
	got, err := ParseReportCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseReportCursor() = %v", err)
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

func TestReportCursorParseRejectsBadInput(t *testing.T) {
	f := SampleReportCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseReportCursor(f); err == nil {
		t.Error("ParseReportCursor() accepted a record with no id")
	}
	f = SampleReportCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseReportCursor(f); err == nil {
		t.Error("ParseReportCursor() accepted a non-numeric weight")
	}
}

func TestReportCursorCollections(t *testing.T) {
	xs := []ReportCursor{SampleReportCursor("a", when), SampleReportCursor("bb", when), SampleReportCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveReportCursors(xs)); got != 2 {
		t.Errorf("ActiveReportCursors() kept %d, want 2", got)
	}
	if got, want := SumReportCursors(xs), 1+3; got != want {
		t.Errorf("SumReportCursors() = %d, want %d", got, want)
	}
	SortReportCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortReportCursors() put %q first", xs[0].ID)
	}
	if len(IndexReportCursors(xs)) != 3 {
		t.Errorf("IndexReportCursors() = %v", IndexReportCursors(xs))
	}
	if err := ValidateReportCursors(ActiveReportCursors(xs)); err != nil {
		t.Errorf("ValidateReportCursors() = %v", err)
	}
}

func TestReportCursorNormalizeClamps(t *testing.T) {
	x := SampleReportCursor("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
