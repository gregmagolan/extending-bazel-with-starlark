package catalog

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
		if !strings.HasPrefix(Prefixes[i], "catalog/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleCatalogRecord("ok", when).Fields()); err != nil {
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

func TestCatalogRecordRoundTrip(t *testing.T) {
	x := SampleCatalogRecord("catalogrecord-1", when)
	got, err := ParseCatalogRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogRecord() = %v", err)
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

func TestCatalogRecordParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogRecord(f); err == nil {
		t.Error("ParseCatalogRecord() accepted a record with no id")
	}
	f = SampleCatalogRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCatalogRecord(f); err == nil {
		t.Error("ParseCatalogRecord() accepted a non-numeric capacity")
	}
}

func TestCatalogRecordCollections(t *testing.T) {
	xs := []CatalogRecord{SampleCatalogRecord("a", when), SampleCatalogRecord("bb", when), SampleCatalogRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogRecords(xs)); got != 2 {
		t.Errorf("ActiveCatalogRecords() kept %d, want 2", got)
	}
	if got, want := SumCatalogRecords(xs), 1+3; got != want {
		t.Errorf("SumCatalogRecords() = %d, want %d", got, want)
	}
	SortCatalogRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogRecords() put %q first", xs[0].ID)
	}
	if len(IndexCatalogRecords(xs)) != 3 {
		t.Errorf("IndexCatalogRecords() = %v", IndexCatalogRecords(xs))
	}
	if err := ValidateCatalogRecords(ActiveCatalogRecords(xs)); err != nil {
		t.Errorf("ValidateCatalogRecords() = %v", err)
	}
}

func TestCatalogRecordNormalizeClamps(t *testing.T) {
	x := SampleCatalogRecord("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogEntryRoundTrip(t *testing.T) {
	x := SampleCatalogEntry("catalogentry-1", when)
	got, err := ParseCatalogEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogEntry() = %v", err)
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

func TestCatalogEntryParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogEntry(f); err == nil {
		t.Error("ParseCatalogEntry() accepted a record with no id")
	}
	f = SampleCatalogEntry("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseCatalogEntry(f); err == nil {
		t.Error("ParseCatalogEntry() accepted a non-numeric version")
	}
}

func TestCatalogEntryCollections(t *testing.T) {
	xs := []CatalogEntry{SampleCatalogEntry("a", when), SampleCatalogEntry("bb", when), SampleCatalogEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogEntrys(xs)); got != 2 {
		t.Errorf("ActiveCatalogEntrys() kept %d, want 2", got)
	}
	if got, want := SumCatalogEntrys(xs), 1+3; got != want {
		t.Errorf("SumCatalogEntrys() = %d, want %d", got, want)
	}
	SortCatalogEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogEntrys() put %q first", xs[0].ID)
	}
	if len(IndexCatalogEntrys(xs)) != 3 {
		t.Errorf("IndexCatalogEntrys() = %v", IndexCatalogEntrys(xs))
	}
	if err := ValidateCatalogEntrys(ActiveCatalogEntrys(xs)); err != nil {
		t.Errorf("ValidateCatalogEntrys() = %v", err)
	}
}

func TestCatalogEntryNormalizeClamps(t *testing.T) {
	x := SampleCatalogEntry("a", when)
	x.Label = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogItemRoundTrip(t *testing.T) {
	x := SampleCatalogItem("catalogitem-1", when)
	got, err := ParseCatalogItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogItem() = %v", err)
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

func TestCatalogItemParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogItem(f); err == nil {
		t.Error("ParseCatalogItem() accepted a record with no id")
	}
	f = SampleCatalogItem("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCatalogItem(f); err == nil {
		t.Error("ParseCatalogItem() accepted a non-numeric capacity")
	}
}

func TestCatalogItemCollections(t *testing.T) {
	xs := []CatalogItem{SampleCatalogItem("a", when), SampleCatalogItem("bb", when), SampleCatalogItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogItems(xs)); got != 2 {
		t.Errorf("ActiveCatalogItems() kept %d, want 2", got)
	}
	if got, want := SumCatalogItems(xs), 1+3; got != want {
		t.Errorf("SumCatalogItems() = %d, want %d", got, want)
	}
	SortCatalogItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogItems() put %q first", xs[0].ID)
	}
	if len(IndexCatalogItems(xs)) != 3 {
		t.Errorf("IndexCatalogItems() = %v", IndexCatalogItems(xs))
	}
	if err := ValidateCatalogItems(ActiveCatalogItems(xs)); err != nil {
		t.Errorf("ValidateCatalogItems() = %v", err)
	}
}

func TestCatalogItemNormalizeClamps(t *testing.T) {
	x := SampleCatalogItem("a", when)
	x.Comment = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogDraftRoundTrip(t *testing.T) {
	x := SampleCatalogDraft("catalogdraft-1", when)
	got, err := ParseCatalogDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogDraft() = %v", err)
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

func TestCatalogDraftParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogDraft(f); err == nil {
		t.Error("ParseCatalogDraft() accepted a record with no id")
	}
	f = SampleCatalogDraft("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCatalogDraft(f); err == nil {
		t.Error("ParseCatalogDraft() accepted a non-numeric count")
	}
}

func TestCatalogDraftCollections(t *testing.T) {
	xs := []CatalogDraft{SampleCatalogDraft("a", when), SampleCatalogDraft("bb", when), SampleCatalogDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogDrafts(xs)); got != 2 {
		t.Errorf("ActiveCatalogDrafts() kept %d, want 2", got)
	}
	if got, want := SumCatalogDrafts(xs), 1+3; got != want {
		t.Errorf("SumCatalogDrafts() = %d, want %d", got, want)
	}
	SortCatalogDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogDrafts() put %q first", xs[0].ID)
	}
	if len(IndexCatalogDrafts(xs)) != 3 {
		t.Errorf("IndexCatalogDrafts() = %v", IndexCatalogDrafts(xs))
	}
	if err := ValidateCatalogDrafts(ActiveCatalogDrafts(xs)); err != nil {
		t.Errorf("ValidateCatalogDrafts() = %v", err)
	}
}

func TestCatalogDraftNormalizeClamps(t *testing.T) {
	x := SampleCatalogDraft("a", when)
	x.Source = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogRevisionRoundTrip(t *testing.T) {
	x := SampleCatalogRevision("catalogrevision-1", when)
	got, err := ParseCatalogRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogRevision() = %v", err)
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

func TestCatalogRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogRevision(f); err == nil {
		t.Error("ParseCatalogRevision() accepted a record with no id")
	}
	f = SampleCatalogRevision("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCatalogRevision(f); err == nil {
		t.Error("ParseCatalogRevision() accepted a non-numeric capacity")
	}
}

func TestCatalogRevisionCollections(t *testing.T) {
	xs := []CatalogRevision{SampleCatalogRevision("a", when), SampleCatalogRevision("bb", when), SampleCatalogRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogRevisions(xs)); got != 2 {
		t.Errorf("ActiveCatalogRevisions() kept %d, want 2", got)
	}
	if got, want := SumCatalogRevisions(xs), 1+3; got != want {
		t.Errorf("SumCatalogRevisions() = %d, want %d", got, want)
	}
	SortCatalogRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogRevisions() put %q first", xs[0].ID)
	}
	if len(IndexCatalogRevisions(xs)) != 3 {
		t.Errorf("IndexCatalogRevisions() = %v", IndexCatalogRevisions(xs))
	}
	if err := ValidateCatalogRevisions(ActiveCatalogRevisions(xs)); err != nil {
		t.Errorf("ValidateCatalogRevisions() = %v", err)
	}
}

func TestCatalogRevisionNormalizeClamps(t *testing.T) {
	x := SampleCatalogRevision("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogSnapshotRoundTrip(t *testing.T) {
	x := SampleCatalogSnapshot("catalogsnapshot-1", when)
	got, err := ParseCatalogSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogSnapshot() = %v", err)
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

func TestCatalogSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogSnapshot(f); err == nil {
		t.Error("ParseCatalogSnapshot() accepted a record with no id")
	}
	f = SampleCatalogSnapshot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCatalogSnapshot(f); err == nil {
		t.Error("ParseCatalogSnapshot() accepted a non-numeric count")
	}
}

func TestCatalogSnapshotCollections(t *testing.T) {
	xs := []CatalogSnapshot{SampleCatalogSnapshot("a", when), SampleCatalogSnapshot("bb", when), SampleCatalogSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogSnapshots(xs)); got != 2 {
		t.Errorf("ActiveCatalogSnapshots() kept %d, want 2", got)
	}
	if got, want := SumCatalogSnapshots(xs), 1+3; got != want {
		t.Errorf("SumCatalogSnapshots() = %d, want %d", got, want)
	}
	SortCatalogSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexCatalogSnapshots(xs)) != 3 {
		t.Errorf("IndexCatalogSnapshots() = %v", IndexCatalogSnapshots(xs))
	}
	if err := ValidateCatalogSnapshots(ActiveCatalogSnapshots(xs)); err != nil {
		t.Errorf("ValidateCatalogSnapshots() = %v", err)
	}
}

func TestCatalogSnapshotNormalizeClamps(t *testing.T) {
	x := SampleCatalogSnapshot("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogBatchRoundTrip(t *testing.T) {
	x := SampleCatalogBatch("catalogbatch-1", when)
	got, err := ParseCatalogBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogBatch() = %v", err)
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

func TestCatalogBatchParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogBatch(f); err == nil {
		t.Error("ParseCatalogBatch() accepted a record with no id")
	}
	f = SampleCatalogBatch("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseCatalogBatch(f); err == nil {
		t.Error("ParseCatalogBatch() accepted a non-numeric offset")
	}
}

func TestCatalogBatchCollections(t *testing.T) {
	xs := []CatalogBatch{SampleCatalogBatch("a", when), SampleCatalogBatch("bb", when), SampleCatalogBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogBatchs(xs)); got != 2 {
		t.Errorf("ActiveCatalogBatchs() kept %d, want 2", got)
	}
	if got, want := SumCatalogBatchs(xs), 1+3; got != want {
		t.Errorf("SumCatalogBatchs() = %d, want %d", got, want)
	}
	SortCatalogBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogBatchs() put %q first", xs[0].ID)
	}
	if len(IndexCatalogBatchs(xs)) != 3 {
		t.Errorf("IndexCatalogBatchs() = %v", IndexCatalogBatchs(xs))
	}
	if err := ValidateCatalogBatchs(ActiveCatalogBatchs(xs)); err != nil {
		t.Errorf("ValidateCatalogBatchs() = %v", err)
	}
}

func TestCatalogBatchNormalizeClamps(t *testing.T) {
	x := SampleCatalogBatch("a", when)
	x.Source = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogRequestRoundTrip(t *testing.T) {
	x := SampleCatalogRequest("catalogrequest-1", when)
	got, err := ParseCatalogRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogRequest() = %v", err)
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

func TestCatalogRequestParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogRequest(f); err == nil {
		t.Error("ParseCatalogRequest() accepted a record with no id")
	}
	f = SampleCatalogRequest("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseCatalogRequest(f); err == nil {
		t.Error("ParseCatalogRequest() accepted a non-numeric version")
	}
}

func TestCatalogRequestCollections(t *testing.T) {
	xs := []CatalogRequest{SampleCatalogRequest("a", when), SampleCatalogRequest("bb", when), SampleCatalogRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogRequests(xs)); got != 2 {
		t.Errorf("ActiveCatalogRequests() kept %d, want 2", got)
	}
	if got, want := SumCatalogRequests(xs), 1+3; got != want {
		t.Errorf("SumCatalogRequests() = %d, want %d", got, want)
	}
	SortCatalogRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogRequests() put %q first", xs[0].ID)
	}
	if len(IndexCatalogRequests(xs)) != 3 {
		t.Errorf("IndexCatalogRequests() = %v", IndexCatalogRequests(xs))
	}
	if err := ValidateCatalogRequests(ActiveCatalogRequests(xs)); err != nil {
		t.Errorf("ValidateCatalogRequests() = %v", err)
	}
}

func TestCatalogRequestNormalizeClamps(t *testing.T) {
	x := SampleCatalogRequest("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogSummaryRoundTrip(t *testing.T) {
	x := SampleCatalogSummary("catalogsummary-1", when)
	got, err := ParseCatalogSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogSummary() = %v", err)
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

func TestCatalogSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogSummary(f); err == nil {
		t.Error("ParseCatalogSummary() accepted a record with no id")
	}
	f = SampleCatalogSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCatalogSummary(f); err == nil {
		t.Error("ParseCatalogSummary() accepted a non-numeric capacity")
	}
}

func TestCatalogSummaryCollections(t *testing.T) {
	xs := []CatalogSummary{SampleCatalogSummary("a", when), SampleCatalogSummary("bb", when), SampleCatalogSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogSummarys(xs)); got != 2 {
		t.Errorf("ActiveCatalogSummarys() kept %d, want 2", got)
	}
	if got, want := SumCatalogSummarys(xs), 1+3; got != want {
		t.Errorf("SumCatalogSummarys() = %d, want %d", got, want)
	}
	SortCatalogSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogSummarys() put %q first", xs[0].ID)
	}
	if len(IndexCatalogSummarys(xs)) != 3 {
		t.Errorf("IndexCatalogSummarys() = %v", IndexCatalogSummarys(xs))
	}
	if err := ValidateCatalogSummarys(ActiveCatalogSummarys(xs)); err != nil {
		t.Errorf("ValidateCatalogSummarys() = %v", err)
	}
}

func TestCatalogSummaryNormalizeClamps(t *testing.T) {
	x := SampleCatalogSummary("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogPolicyRoundTrip(t *testing.T) {
	x := SampleCatalogPolicy("catalogpolicy-1", when)
	got, err := ParseCatalogPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogPolicy() = %v", err)
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

func TestCatalogPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogPolicy(f); err == nil {
		t.Error("ParseCatalogPolicy() accepted a record with no id")
	}
	f = SampleCatalogPolicy("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseCatalogPolicy(f); err == nil {
		t.Error("ParseCatalogPolicy() accepted a non-numeric attempts")
	}
}

func TestCatalogPolicyCollections(t *testing.T) {
	xs := []CatalogPolicy{SampleCatalogPolicy("a", when), SampleCatalogPolicy("bb", when), SampleCatalogPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogPolicys(xs)); got != 2 {
		t.Errorf("ActiveCatalogPolicys() kept %d, want 2", got)
	}
	if got, want := SumCatalogPolicys(xs), 1+3; got != want {
		t.Errorf("SumCatalogPolicys() = %d, want %d", got, want)
	}
	SortCatalogPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogPolicys() put %q first", xs[0].ID)
	}
	if len(IndexCatalogPolicys(xs)) != 3 {
		t.Errorf("IndexCatalogPolicys() = %v", IndexCatalogPolicys(xs))
	}
	if err := ValidateCatalogPolicys(ActiveCatalogPolicys(xs)); err != nil {
		t.Errorf("ValidateCatalogPolicys() = %v", err)
	}
}

func TestCatalogPolicyNormalizeClamps(t *testing.T) {
	x := SampleCatalogPolicy("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogLinkRoundTrip(t *testing.T) {
	x := SampleCatalogLink("cataloglink-1", when)
	got, err := ParseCatalogLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogLink() = %v", err)
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

func TestCatalogLinkParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogLink(f); err == nil {
		t.Error("ParseCatalogLink() accepted a record with no id")
	}
	f = SampleCatalogLink("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseCatalogLink(f); err == nil {
		t.Error("ParseCatalogLink() accepted a non-numeric version")
	}
}

func TestCatalogLinkCollections(t *testing.T) {
	xs := []CatalogLink{SampleCatalogLink("a", when), SampleCatalogLink("bb", when), SampleCatalogLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogLinks(xs)); got != 2 {
		t.Errorf("ActiveCatalogLinks() kept %d, want 2", got)
	}
	if got, want := SumCatalogLinks(xs), 1+3; got != want {
		t.Errorf("SumCatalogLinks() = %d, want %d", got, want)
	}
	SortCatalogLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogLinks() put %q first", xs[0].ID)
	}
	if len(IndexCatalogLinks(xs)) != 3 {
		t.Errorf("IndexCatalogLinks() = %v", IndexCatalogLinks(xs))
	}
	if err := ValidateCatalogLinks(ActiveCatalogLinks(xs)); err != nil {
		t.Errorf("ValidateCatalogLinks() = %v", err)
	}
}

func TestCatalogLinkNormalizeClamps(t *testing.T) {
	x := SampleCatalogLink("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogNoteRoundTrip(t *testing.T) {
	x := SampleCatalogNote("catalognote-1", when)
	got, err := ParseCatalogNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogNote() = %v", err)
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

func TestCatalogNoteParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogNote(f); err == nil {
		t.Error("ParseCatalogNote() accepted a record with no id")
	}
	f = SampleCatalogNote("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseCatalogNote(f); err == nil {
		t.Error("ParseCatalogNote() accepted a non-numeric weight")
	}
}

func TestCatalogNoteCollections(t *testing.T) {
	xs := []CatalogNote{SampleCatalogNote("a", when), SampleCatalogNote("bb", when), SampleCatalogNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogNotes(xs)); got != 2 {
		t.Errorf("ActiveCatalogNotes() kept %d, want 2", got)
	}
	if got, want := SumCatalogNotes(xs), 1+3; got != want {
		t.Errorf("SumCatalogNotes() = %d, want %d", got, want)
	}
	SortCatalogNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogNotes() put %q first", xs[0].ID)
	}
	if len(IndexCatalogNotes(xs)) != 3 {
		t.Errorf("IndexCatalogNotes() = %v", IndexCatalogNotes(xs))
	}
	if err := ValidateCatalogNotes(ActiveCatalogNotes(xs)); err != nil {
		t.Errorf("ValidateCatalogNotes() = %v", err)
	}
}

func TestCatalogNoteNormalizeClamps(t *testing.T) {
	x := SampleCatalogNote("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogWindowRoundTrip(t *testing.T) {
	x := SampleCatalogWindow("catalogwindow-1", when)
	got, err := ParseCatalogWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogWindow() = %v", err)
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

func TestCatalogWindowParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogWindow(f); err == nil {
		t.Error("ParseCatalogWindow() accepted a record with no id")
	}
	f = SampleCatalogWindow("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseCatalogWindow(f); err == nil {
		t.Error("ParseCatalogWindow() accepted a non-numeric offset")
	}
}

func TestCatalogWindowCollections(t *testing.T) {
	xs := []CatalogWindow{SampleCatalogWindow("a", when), SampleCatalogWindow("bb", when), SampleCatalogWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogWindows(xs)); got != 2 {
		t.Errorf("ActiveCatalogWindows() kept %d, want 2", got)
	}
	if got, want := SumCatalogWindows(xs), 1+3; got != want {
		t.Errorf("SumCatalogWindows() = %d, want %d", got, want)
	}
	SortCatalogWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogWindows() put %q first", xs[0].ID)
	}
	if len(IndexCatalogWindows(xs)) != 3 {
		t.Errorf("IndexCatalogWindows() = %v", IndexCatalogWindows(xs))
	}
	if err := ValidateCatalogWindows(ActiveCatalogWindows(xs)); err != nil {
		t.Errorf("ValidateCatalogWindows() = %v", err)
	}
}

func TestCatalogWindowNormalizeClamps(t *testing.T) {
	x := SampleCatalogWindow("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogQuotaRoundTrip(t *testing.T) {
	x := SampleCatalogQuota("catalogquota-1", when)
	got, err := ParseCatalogQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogQuota() = %v", err)
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

func TestCatalogQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogQuota(f); err == nil {
		t.Error("ParseCatalogQuota() accepted a record with no id")
	}
	f = SampleCatalogQuota("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseCatalogQuota(f); err == nil {
		t.Error("ParseCatalogQuota() accepted a non-numeric attempts")
	}
}

func TestCatalogQuotaCollections(t *testing.T) {
	xs := []CatalogQuota{SampleCatalogQuota("a", when), SampleCatalogQuota("bb", when), SampleCatalogQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogQuotas(xs)); got != 2 {
		t.Errorf("ActiveCatalogQuotas() kept %d, want 2", got)
	}
	if got, want := SumCatalogQuotas(xs), 1+3; got != want {
		t.Errorf("SumCatalogQuotas() = %d, want %d", got, want)
	}
	SortCatalogQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogQuotas() put %q first", xs[0].ID)
	}
	if len(IndexCatalogQuotas(xs)) != 3 {
		t.Errorf("IndexCatalogQuotas() = %v", IndexCatalogQuotas(xs))
	}
	if err := ValidateCatalogQuotas(ActiveCatalogQuotas(xs)); err != nil {
		t.Errorf("ValidateCatalogQuotas() = %v", err)
	}
}

func TestCatalogQuotaNormalizeClamps(t *testing.T) {
	x := SampleCatalogQuota("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogBucketRoundTrip(t *testing.T) {
	x := SampleCatalogBucket("catalogbucket-1", when)
	got, err := ParseCatalogBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogBucket() = %v", err)
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

func TestCatalogBucketParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogBucket(f); err == nil {
		t.Error("ParseCatalogBucket() accepted a record with no id")
	}
	f = SampleCatalogBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseCatalogBucket(f); err == nil {
		t.Error("ParseCatalogBucket() accepted a non-numeric retained")
	}
}

func TestCatalogBucketCollections(t *testing.T) {
	xs := []CatalogBucket{SampleCatalogBucket("a", when), SampleCatalogBucket("bb", when), SampleCatalogBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogBuckets(xs)); got != 2 {
		t.Errorf("ActiveCatalogBuckets() kept %d, want 2", got)
	}
	if got, want := SumCatalogBuckets(xs), 1+3; got != want {
		t.Errorf("SumCatalogBuckets() = %d, want %d", got, want)
	}
	SortCatalogBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogBuckets() put %q first", xs[0].ID)
	}
	if len(IndexCatalogBuckets(xs)) != 3 {
		t.Errorf("IndexCatalogBuckets() = %v", IndexCatalogBuckets(xs))
	}
	if err := ValidateCatalogBuckets(ActiveCatalogBuckets(xs)); err != nil {
		t.Errorf("ValidateCatalogBuckets() = %v", err)
	}
}

func TestCatalogBucketNormalizeClamps(t *testing.T) {
	x := SampleCatalogBucket("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogLaneRoundTrip(t *testing.T) {
	x := SampleCatalogLane("cataloglane-1", when)
	got, err := ParseCatalogLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogLane() = %v", err)
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

func TestCatalogLaneParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogLane(f); err == nil {
		t.Error("ParseCatalogLane() accepted a record with no id")
	}
	f = SampleCatalogLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCatalogLane(f); err == nil {
		t.Error("ParseCatalogLane() accepted a non-numeric capacity")
	}
}

func TestCatalogLaneCollections(t *testing.T) {
	xs := []CatalogLane{SampleCatalogLane("a", when), SampleCatalogLane("bb", when), SampleCatalogLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogLanes(xs)); got != 2 {
		t.Errorf("ActiveCatalogLanes() kept %d, want 2", got)
	}
	if got, want := SumCatalogLanes(xs), 1+3; got != want {
		t.Errorf("SumCatalogLanes() = %d, want %d", got, want)
	}
	SortCatalogLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogLanes() put %q first", xs[0].ID)
	}
	if len(IndexCatalogLanes(xs)) != 3 {
		t.Errorf("IndexCatalogLanes() = %v", IndexCatalogLanes(xs))
	}
	if err := ValidateCatalogLanes(ActiveCatalogLanes(xs)); err != nil {
		t.Errorf("ValidateCatalogLanes() = %v", err)
	}
}

func TestCatalogLaneNormalizeClamps(t *testing.T) {
	x := SampleCatalogLane("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogSlotRoundTrip(t *testing.T) {
	x := SampleCatalogSlot("catalogslot-1", when)
	got, err := ParseCatalogSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogSlot() = %v", err)
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

func TestCatalogSlotParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogSlot(f); err == nil {
		t.Error("ParseCatalogSlot() accepted a record with no id")
	}
	f = SampleCatalogSlot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseCatalogSlot(f); err == nil {
		t.Error("ParseCatalogSlot() accepted a non-numeric retained")
	}
}

func TestCatalogSlotCollections(t *testing.T) {
	xs := []CatalogSlot{SampleCatalogSlot("a", when), SampleCatalogSlot("bb", when), SampleCatalogSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogSlots(xs)); got != 2 {
		t.Errorf("ActiveCatalogSlots() kept %d, want 2", got)
	}
	if got, want := SumCatalogSlots(xs), 1+3; got != want {
		t.Errorf("SumCatalogSlots() = %d, want %d", got, want)
	}
	SortCatalogSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogSlots() put %q first", xs[0].ID)
	}
	if len(IndexCatalogSlots(xs)) != 3 {
		t.Errorf("IndexCatalogSlots() = %v", IndexCatalogSlots(xs))
	}
	if err := ValidateCatalogSlots(ActiveCatalogSlots(xs)); err != nil {
		t.Errorf("ValidateCatalogSlots() = %v", err)
	}
}

func TestCatalogSlotNormalizeClamps(t *testing.T) {
	x := SampleCatalogSlot("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogTagRoundTrip(t *testing.T) {
	x := SampleCatalogTag("catalogtag-1", when)
	got, err := ParseCatalogTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogTag() = %v", err)
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

func TestCatalogTagParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogTag(f); err == nil {
		t.Error("ParseCatalogTag() accepted a record with no id")
	}
	f = SampleCatalogTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCatalogTag(f); err == nil {
		t.Error("ParseCatalogTag() accepted a non-numeric count")
	}
}

func TestCatalogTagCollections(t *testing.T) {
	xs := []CatalogTag{SampleCatalogTag("a", when), SampleCatalogTag("bb", when), SampleCatalogTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogTags(xs)); got != 2 {
		t.Errorf("ActiveCatalogTags() kept %d, want 2", got)
	}
	if got, want := SumCatalogTags(xs), 1+3; got != want {
		t.Errorf("SumCatalogTags() = %d, want %d", got, want)
	}
	SortCatalogTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogTags() put %q first", xs[0].ID)
	}
	if len(IndexCatalogTags(xs)) != 3 {
		t.Errorf("IndexCatalogTags() = %v", IndexCatalogTags(xs))
	}
	if err := ValidateCatalogTags(ActiveCatalogTags(xs)); err != nil {
		t.Errorf("ValidateCatalogTags() = %v", err)
	}
}

func TestCatalogTagNormalizeClamps(t *testing.T) {
	x := SampleCatalogTag("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogRuleRoundTrip(t *testing.T) {
	x := SampleCatalogRule("catalogrule-1", when)
	got, err := ParseCatalogRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogRule() = %v", err)
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

func TestCatalogRuleParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogRule(f); err == nil {
		t.Error("ParseCatalogRule() accepted a record with no id")
	}
	f = SampleCatalogRule("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCatalogRule(f); err == nil {
		t.Error("ParseCatalogRule() accepted a non-numeric count")
	}
}

func TestCatalogRuleCollections(t *testing.T) {
	xs := []CatalogRule{SampleCatalogRule("a", when), SampleCatalogRule("bb", when), SampleCatalogRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogRules(xs)); got != 2 {
		t.Errorf("ActiveCatalogRules() kept %d, want 2", got)
	}
	if got, want := SumCatalogRules(xs), 1+3; got != want {
		t.Errorf("SumCatalogRules() = %d, want %d", got, want)
	}
	SortCatalogRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogRules() put %q first", xs[0].ID)
	}
	if len(IndexCatalogRules(xs)) != 3 {
		t.Errorf("IndexCatalogRules() = %v", IndexCatalogRules(xs))
	}
	if err := ValidateCatalogRules(ActiveCatalogRules(xs)); err != nil {
		t.Errorf("ValidateCatalogRules() = %v", err)
	}
}

func TestCatalogRuleNormalizeClamps(t *testing.T) {
	x := SampleCatalogRule("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogHoldRoundTrip(t *testing.T) {
	x := SampleCatalogHold("cataloghold-1", when)
	got, err := ParseCatalogHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogHold() = %v", err)
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

func TestCatalogHoldParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogHold(f); err == nil {
		t.Error("ParseCatalogHold() accepted a record with no id")
	}
	f = SampleCatalogHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseCatalogHold(f); err == nil {
		t.Error("ParseCatalogHold() accepted a non-numeric weight")
	}
}

func TestCatalogHoldCollections(t *testing.T) {
	xs := []CatalogHold{SampleCatalogHold("a", when), SampleCatalogHold("bb", when), SampleCatalogHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogHolds(xs)); got != 2 {
		t.Errorf("ActiveCatalogHolds() kept %d, want 2", got)
	}
	if got, want := SumCatalogHolds(xs), 1+3; got != want {
		t.Errorf("SumCatalogHolds() = %d, want %d", got, want)
	}
	SortCatalogHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogHolds() put %q first", xs[0].ID)
	}
	if len(IndexCatalogHolds(xs)) != 3 {
		t.Errorf("IndexCatalogHolds() = %v", IndexCatalogHolds(xs))
	}
	if err := ValidateCatalogHolds(ActiveCatalogHolds(xs)); err != nil {
		t.Errorf("ValidateCatalogHolds() = %v", err)
	}
}

func TestCatalogHoldNormalizeClamps(t *testing.T) {
	x := SampleCatalogHold("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogClaimRoundTrip(t *testing.T) {
	x := SampleCatalogClaim("catalogclaim-1", when)
	got, err := ParseCatalogClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogClaim() = %v", err)
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

func TestCatalogClaimParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogClaim(f); err == nil {
		t.Error("ParseCatalogClaim() accepted a record with no id")
	}
	f = SampleCatalogClaim("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseCatalogClaim(f); err == nil {
		t.Error("ParseCatalogClaim() accepted a non-numeric capacity")
	}
}

func TestCatalogClaimCollections(t *testing.T) {
	xs := []CatalogClaim{SampleCatalogClaim("a", when), SampleCatalogClaim("bb", when), SampleCatalogClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogClaims(xs)); got != 2 {
		t.Errorf("ActiveCatalogClaims() kept %d, want 2", got)
	}
	if got, want := SumCatalogClaims(xs), 1+3; got != want {
		t.Errorf("SumCatalogClaims() = %d, want %d", got, want)
	}
	SortCatalogClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogClaims() put %q first", xs[0].ID)
	}
	if len(IndexCatalogClaims(xs)) != 3 {
		t.Errorf("IndexCatalogClaims() = %v", IndexCatalogClaims(xs))
	}
	if err := ValidateCatalogClaims(ActiveCatalogClaims(xs)); err != nil {
		t.Errorf("ValidateCatalogClaims() = %v", err)
	}
}

func TestCatalogClaimNormalizeClamps(t *testing.T) {
	x := SampleCatalogClaim("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestCatalogCursorRoundTrip(t *testing.T) {
	x := SampleCatalogCursor("catalogcursor-1", when)
	got, err := ParseCatalogCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseCatalogCursor() = %v", err)
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

func TestCatalogCursorParseRejectsBadInput(t *testing.T) {
	f := SampleCatalogCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseCatalogCursor(f); err == nil {
		t.Error("ParseCatalogCursor() accepted a record with no id")
	}
	f = SampleCatalogCursor("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseCatalogCursor(f); err == nil {
		t.Error("ParseCatalogCursor() accepted a non-numeric count")
	}
}

func TestCatalogCursorCollections(t *testing.T) {
	xs := []CatalogCursor{SampleCatalogCursor("a", when), SampleCatalogCursor("bb", when), SampleCatalogCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveCatalogCursors(xs)); got != 2 {
		t.Errorf("ActiveCatalogCursors() kept %d, want 2", got)
	}
	if got, want := SumCatalogCursors(xs), 1+3; got != want {
		t.Errorf("SumCatalogCursors() = %d, want %d", got, want)
	}
	SortCatalogCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortCatalogCursors() put %q first", xs[0].ID)
	}
	if len(IndexCatalogCursors(xs)) != 3 {
		t.Errorf("IndexCatalogCursors() = %v", IndexCatalogCursors(xs))
	}
	if err := ValidateCatalogCursors(ActiveCatalogCursors(xs)); err != nil {
		t.Errorf("ValidateCatalogCursors() = %v", err)
	}
}

func TestCatalogCursorNormalizeClamps(t *testing.T) {
	x := SampleCatalogCursor("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
