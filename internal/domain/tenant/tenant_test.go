package tenant

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
		if !strings.HasPrefix(Prefixes[i], "tenant/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleTenantRecord("ok", when).Fields()); err != nil {
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

func TestTenantRecordRoundTrip(t *testing.T) {
	x := SampleTenantRecord("tenantrecord-1", when)
	got, err := ParseTenantRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantRecord() = %v", err)
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

func TestTenantRecordParseRejectsBadInput(t *testing.T) {
	f := SampleTenantRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantRecord(f); err == nil {
		t.Error("ParseTenantRecord() accepted a record with no id")
	}
	f = SampleTenantRecord("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTenantRecord(f); err == nil {
		t.Error("ParseTenantRecord() accepted a non-numeric offset")
	}
}

func TestTenantRecordCollections(t *testing.T) {
	xs := []TenantRecord{SampleTenantRecord("a", when), SampleTenantRecord("bb", when), SampleTenantRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantRecords(xs)); got != 2 {
		t.Errorf("ActiveTenantRecords() kept %d, want 2", got)
	}
	if got, want := SumTenantRecords(xs), 1+3; got != want {
		t.Errorf("SumTenantRecords() = %d, want %d", got, want)
	}
	SortTenantRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantRecords() put %q first", xs[0].ID)
	}
	if len(IndexTenantRecords(xs)) != 3 {
		t.Errorf("IndexTenantRecords() = %v", IndexTenantRecords(xs))
	}
	if err := ValidateTenantRecords(ActiveTenantRecords(xs)); err != nil {
		t.Errorf("ValidateTenantRecords() = %v", err)
	}
}

func TestTenantRecordNormalizeClamps(t *testing.T) {
	x := SampleTenantRecord("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantEntryRoundTrip(t *testing.T) {
	x := SampleTenantEntry("tenantentry-1", when)
	got, err := ParseTenantEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantEntry() = %v", err)
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

func TestTenantEntryParseRejectsBadInput(t *testing.T) {
	f := SampleTenantEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantEntry(f); err == nil {
		t.Error("ParseTenantEntry() accepted a record with no id")
	}
	f = SampleTenantEntry("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTenantEntry(f); err == nil {
		t.Error("ParseTenantEntry() accepted a non-numeric retained")
	}
}

func TestTenantEntryCollections(t *testing.T) {
	xs := []TenantEntry{SampleTenantEntry("a", when), SampleTenantEntry("bb", when), SampleTenantEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantEntrys(xs)); got != 2 {
		t.Errorf("ActiveTenantEntrys() kept %d, want 2", got)
	}
	if got, want := SumTenantEntrys(xs), 1+3; got != want {
		t.Errorf("SumTenantEntrys() = %d, want %d", got, want)
	}
	SortTenantEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantEntrys() put %q first", xs[0].ID)
	}
	if len(IndexTenantEntrys(xs)) != 3 {
		t.Errorf("IndexTenantEntrys() = %v", IndexTenantEntrys(xs))
	}
	if err := ValidateTenantEntrys(ActiveTenantEntrys(xs)); err != nil {
		t.Errorf("ValidateTenantEntrys() = %v", err)
	}
}

func TestTenantEntryNormalizeClamps(t *testing.T) {
	x := SampleTenantEntry("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantItemRoundTrip(t *testing.T) {
	x := SampleTenantItem("tenantitem-1", when)
	got, err := ParseTenantItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantItem() = %v", err)
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

func TestTenantItemParseRejectsBadInput(t *testing.T) {
	f := SampleTenantItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantItem(f); err == nil {
		t.Error("ParseTenantItem() accepted a record with no id")
	}
	f = SampleTenantItem("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTenantItem(f); err == nil {
		t.Error("ParseTenantItem() accepted a non-numeric priority")
	}
}

func TestTenantItemCollections(t *testing.T) {
	xs := []TenantItem{SampleTenantItem("a", when), SampleTenantItem("bb", when), SampleTenantItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantItems(xs)); got != 2 {
		t.Errorf("ActiveTenantItems() kept %d, want 2", got)
	}
	if got, want := SumTenantItems(xs), 1+3; got != want {
		t.Errorf("SumTenantItems() = %d, want %d", got, want)
	}
	SortTenantItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantItems() put %q first", xs[0].ID)
	}
	if len(IndexTenantItems(xs)) != 3 {
		t.Errorf("IndexTenantItems() = %v", IndexTenantItems(xs))
	}
	if err := ValidateTenantItems(ActiveTenantItems(xs)); err != nil {
		t.Errorf("ValidateTenantItems() = %v", err)
	}
}

func TestTenantItemNormalizeClamps(t *testing.T) {
	x := SampleTenantItem("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantDraftRoundTrip(t *testing.T) {
	x := SampleTenantDraft("tenantdraft-1", when)
	got, err := ParseTenantDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantDraft() = %v", err)
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

func TestTenantDraftParseRejectsBadInput(t *testing.T) {
	f := SampleTenantDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantDraft(f); err == nil {
		t.Error("ParseTenantDraft() accepted a record with no id")
	}
	f = SampleTenantDraft("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTenantDraft(f); err == nil {
		t.Error("ParseTenantDraft() accepted a non-numeric capacity")
	}
}

func TestTenantDraftCollections(t *testing.T) {
	xs := []TenantDraft{SampleTenantDraft("a", when), SampleTenantDraft("bb", when), SampleTenantDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantDrafts(xs)); got != 2 {
		t.Errorf("ActiveTenantDrafts() kept %d, want 2", got)
	}
	if got, want := SumTenantDrafts(xs), 1+3; got != want {
		t.Errorf("SumTenantDrafts() = %d, want %d", got, want)
	}
	SortTenantDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantDrafts() put %q first", xs[0].ID)
	}
	if len(IndexTenantDrafts(xs)) != 3 {
		t.Errorf("IndexTenantDrafts() = %v", IndexTenantDrafts(xs))
	}
	if err := ValidateTenantDrafts(ActiveTenantDrafts(xs)); err != nil {
		t.Errorf("ValidateTenantDrafts() = %v", err)
	}
}

func TestTenantDraftNormalizeClamps(t *testing.T) {
	x := SampleTenantDraft("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantRevisionRoundTrip(t *testing.T) {
	x := SampleTenantRevision("tenantrevision-1", when)
	got, err := ParseTenantRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantRevision() = %v", err)
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

func TestTenantRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleTenantRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantRevision(f); err == nil {
		t.Error("ParseTenantRevision() accepted a record with no id")
	}
	f = SampleTenantRevision("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTenantRevision(f); err == nil {
		t.Error("ParseTenantRevision() accepted a non-numeric offset")
	}
}

func TestTenantRevisionCollections(t *testing.T) {
	xs := []TenantRevision{SampleTenantRevision("a", when), SampleTenantRevision("bb", when), SampleTenantRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantRevisions(xs)); got != 2 {
		t.Errorf("ActiveTenantRevisions() kept %d, want 2", got)
	}
	if got, want := SumTenantRevisions(xs), 1+3; got != want {
		t.Errorf("SumTenantRevisions() = %d, want %d", got, want)
	}
	SortTenantRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantRevisions() put %q first", xs[0].ID)
	}
	if len(IndexTenantRevisions(xs)) != 3 {
		t.Errorf("IndexTenantRevisions() = %v", IndexTenantRevisions(xs))
	}
	if err := ValidateTenantRevisions(ActiveTenantRevisions(xs)); err != nil {
		t.Errorf("ValidateTenantRevisions() = %v", err)
	}
}

func TestTenantRevisionNormalizeClamps(t *testing.T) {
	x := SampleTenantRevision("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantSnapshotRoundTrip(t *testing.T) {
	x := SampleTenantSnapshot("tenantsnapshot-1", when)
	got, err := ParseTenantSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantSnapshot() = %v", err)
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

func TestTenantSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleTenantSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantSnapshot(f); err == nil {
		t.Error("ParseTenantSnapshot() accepted a record with no id")
	}
	f = SampleTenantSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTenantSnapshot(f); err == nil {
		t.Error("ParseTenantSnapshot() accepted a non-numeric priority")
	}
}

func TestTenantSnapshotCollections(t *testing.T) {
	xs := []TenantSnapshot{SampleTenantSnapshot("a", when), SampleTenantSnapshot("bb", when), SampleTenantSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantSnapshots(xs)); got != 2 {
		t.Errorf("ActiveTenantSnapshots() kept %d, want 2", got)
	}
	if got, want := SumTenantSnapshots(xs), 1+3; got != want {
		t.Errorf("SumTenantSnapshots() = %d, want %d", got, want)
	}
	SortTenantSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexTenantSnapshots(xs)) != 3 {
		t.Errorf("IndexTenantSnapshots() = %v", IndexTenantSnapshots(xs))
	}
	if err := ValidateTenantSnapshots(ActiveTenantSnapshots(xs)); err != nil {
		t.Errorf("ValidateTenantSnapshots() = %v", err)
	}
}

func TestTenantSnapshotNormalizeClamps(t *testing.T) {
	x := SampleTenantSnapshot("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantBatchRoundTrip(t *testing.T) {
	x := SampleTenantBatch("tenantbatch-1", when)
	got, err := ParseTenantBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantBatch() = %v", err)
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

func TestTenantBatchParseRejectsBadInput(t *testing.T) {
	f := SampleTenantBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantBatch(f); err == nil {
		t.Error("ParseTenantBatch() accepted a record with no id")
	}
	f = SampleTenantBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTenantBatch(f); err == nil {
		t.Error("ParseTenantBatch() accepted a non-numeric version")
	}
}

func TestTenantBatchCollections(t *testing.T) {
	xs := []TenantBatch{SampleTenantBatch("a", when), SampleTenantBatch("bb", when), SampleTenantBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantBatchs(xs)); got != 2 {
		t.Errorf("ActiveTenantBatchs() kept %d, want 2", got)
	}
	if got, want := SumTenantBatchs(xs), 1+3; got != want {
		t.Errorf("SumTenantBatchs() = %d, want %d", got, want)
	}
	SortTenantBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantBatchs() put %q first", xs[0].ID)
	}
	if len(IndexTenantBatchs(xs)) != 3 {
		t.Errorf("IndexTenantBatchs() = %v", IndexTenantBatchs(xs))
	}
	if err := ValidateTenantBatchs(ActiveTenantBatchs(xs)); err != nil {
		t.Errorf("ValidateTenantBatchs() = %v", err)
	}
}

func TestTenantBatchNormalizeClamps(t *testing.T) {
	x := SampleTenantBatch("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantRequestRoundTrip(t *testing.T) {
	x := SampleTenantRequest("tenantrequest-1", when)
	got, err := ParseTenantRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantRequest() = %v", err)
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

func TestTenantRequestParseRejectsBadInput(t *testing.T) {
	f := SampleTenantRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantRequest(f); err == nil {
		t.Error("ParseTenantRequest() accepted a record with no id")
	}
	f = SampleTenantRequest("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTenantRequest(f); err == nil {
		t.Error("ParseTenantRequest() accepted a non-numeric attempts")
	}
}

func TestTenantRequestCollections(t *testing.T) {
	xs := []TenantRequest{SampleTenantRequest("a", when), SampleTenantRequest("bb", when), SampleTenantRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantRequests(xs)); got != 2 {
		t.Errorf("ActiveTenantRequests() kept %d, want 2", got)
	}
	if got, want := SumTenantRequests(xs), 1+3; got != want {
		t.Errorf("SumTenantRequests() = %d, want %d", got, want)
	}
	SortTenantRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantRequests() put %q first", xs[0].ID)
	}
	if len(IndexTenantRequests(xs)) != 3 {
		t.Errorf("IndexTenantRequests() = %v", IndexTenantRequests(xs))
	}
	if err := ValidateTenantRequests(ActiveTenantRequests(xs)); err != nil {
		t.Errorf("ValidateTenantRequests() = %v", err)
	}
}

func TestTenantRequestNormalizeClamps(t *testing.T) {
	x := SampleTenantRequest("a", when)
	x.Reference = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantSummaryRoundTrip(t *testing.T) {
	x := SampleTenantSummary("tenantsummary-1", when)
	got, err := ParseTenantSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantSummary() = %v", err)
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

func TestTenantSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleTenantSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantSummary(f); err == nil {
		t.Error("ParseTenantSummary() accepted a record with no id")
	}
	f = SampleTenantSummary("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTenantSummary(f); err == nil {
		t.Error("ParseTenantSummary() accepted a non-numeric offset")
	}
}

func TestTenantSummaryCollections(t *testing.T) {
	xs := []TenantSummary{SampleTenantSummary("a", when), SampleTenantSummary("bb", when), SampleTenantSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantSummarys(xs)); got != 2 {
		t.Errorf("ActiveTenantSummarys() kept %d, want 2", got)
	}
	if got, want := SumTenantSummarys(xs), 1+3; got != want {
		t.Errorf("SumTenantSummarys() = %d, want %d", got, want)
	}
	SortTenantSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantSummarys() put %q first", xs[0].ID)
	}
	if len(IndexTenantSummarys(xs)) != 3 {
		t.Errorf("IndexTenantSummarys() = %v", IndexTenantSummarys(xs))
	}
	if err := ValidateTenantSummarys(ActiveTenantSummarys(xs)); err != nil {
		t.Errorf("ValidateTenantSummarys() = %v", err)
	}
}

func TestTenantSummaryNormalizeClamps(t *testing.T) {
	x := SampleTenantSummary("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantPolicyRoundTrip(t *testing.T) {
	x := SampleTenantPolicy("tenantpolicy-1", when)
	got, err := ParseTenantPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantPolicy() = %v", err)
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

func TestTenantPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleTenantPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantPolicy(f); err == nil {
		t.Error("ParseTenantPolicy() accepted a record with no id")
	}
	f = SampleTenantPolicy("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTenantPolicy(f); err == nil {
		t.Error("ParseTenantPolicy() accepted a non-numeric count")
	}
}

func TestTenantPolicyCollections(t *testing.T) {
	xs := []TenantPolicy{SampleTenantPolicy("a", when), SampleTenantPolicy("bb", when), SampleTenantPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantPolicys(xs)); got != 2 {
		t.Errorf("ActiveTenantPolicys() kept %d, want 2", got)
	}
	if got, want := SumTenantPolicys(xs), 1+3; got != want {
		t.Errorf("SumTenantPolicys() = %d, want %d", got, want)
	}
	SortTenantPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantPolicys() put %q first", xs[0].ID)
	}
	if len(IndexTenantPolicys(xs)) != 3 {
		t.Errorf("IndexTenantPolicys() = %v", IndexTenantPolicys(xs))
	}
	if err := ValidateTenantPolicys(ActiveTenantPolicys(xs)); err != nil {
		t.Errorf("ValidateTenantPolicys() = %v", err)
	}
}

func TestTenantPolicyNormalizeClamps(t *testing.T) {
	x := SampleTenantPolicy("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantLinkRoundTrip(t *testing.T) {
	x := SampleTenantLink("tenantlink-1", when)
	got, err := ParseTenantLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantLink() = %v", err)
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

func TestTenantLinkParseRejectsBadInput(t *testing.T) {
	f := SampleTenantLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantLink(f); err == nil {
		t.Error("ParseTenantLink() accepted a record with no id")
	}
	f = SampleTenantLink("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTenantLink(f); err == nil {
		t.Error("ParseTenantLink() accepted a non-numeric priority")
	}
}

func TestTenantLinkCollections(t *testing.T) {
	xs := []TenantLink{SampleTenantLink("a", when), SampleTenantLink("bb", when), SampleTenantLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantLinks(xs)); got != 2 {
		t.Errorf("ActiveTenantLinks() kept %d, want 2", got)
	}
	if got, want := SumTenantLinks(xs), 1+3; got != want {
		t.Errorf("SumTenantLinks() = %d, want %d", got, want)
	}
	SortTenantLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantLinks() put %q first", xs[0].ID)
	}
	if len(IndexTenantLinks(xs)) != 3 {
		t.Errorf("IndexTenantLinks() = %v", IndexTenantLinks(xs))
	}
	if err := ValidateTenantLinks(ActiveTenantLinks(xs)); err != nil {
		t.Errorf("ValidateTenantLinks() = %v", err)
	}
}

func TestTenantLinkNormalizeClamps(t *testing.T) {
	x := SampleTenantLink("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantNoteRoundTrip(t *testing.T) {
	x := SampleTenantNote("tenantnote-1", when)
	got, err := ParseTenantNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantNote() = %v", err)
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

func TestTenantNoteParseRejectsBadInput(t *testing.T) {
	f := SampleTenantNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantNote(f); err == nil {
		t.Error("ParseTenantNote() accepted a record with no id")
	}
	f = SampleTenantNote("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTenantNote(f); err == nil {
		t.Error("ParseTenantNote() accepted a non-numeric count")
	}
}

func TestTenantNoteCollections(t *testing.T) {
	xs := []TenantNote{SampleTenantNote("a", when), SampleTenantNote("bb", when), SampleTenantNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantNotes(xs)); got != 2 {
		t.Errorf("ActiveTenantNotes() kept %d, want 2", got)
	}
	if got, want := SumTenantNotes(xs), 1+3; got != want {
		t.Errorf("SumTenantNotes() = %d, want %d", got, want)
	}
	SortTenantNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantNotes() put %q first", xs[0].ID)
	}
	if len(IndexTenantNotes(xs)) != 3 {
		t.Errorf("IndexTenantNotes() = %v", IndexTenantNotes(xs))
	}
	if err := ValidateTenantNotes(ActiveTenantNotes(xs)); err != nil {
		t.Errorf("ValidateTenantNotes() = %v", err)
	}
}

func TestTenantNoteNormalizeClamps(t *testing.T) {
	x := SampleTenantNote("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantWindowRoundTrip(t *testing.T) {
	x := SampleTenantWindow("tenantwindow-1", when)
	got, err := ParseTenantWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantWindow() = %v", err)
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

func TestTenantWindowParseRejectsBadInput(t *testing.T) {
	f := SampleTenantWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantWindow(f); err == nil {
		t.Error("ParseTenantWindow() accepted a record with no id")
	}
	f = SampleTenantWindow("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTenantWindow(f); err == nil {
		t.Error("ParseTenantWindow() accepted a non-numeric attempts")
	}
}

func TestTenantWindowCollections(t *testing.T) {
	xs := []TenantWindow{SampleTenantWindow("a", when), SampleTenantWindow("bb", when), SampleTenantWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantWindows(xs)); got != 2 {
		t.Errorf("ActiveTenantWindows() kept %d, want 2", got)
	}
	if got, want := SumTenantWindows(xs), 1+3; got != want {
		t.Errorf("SumTenantWindows() = %d, want %d", got, want)
	}
	SortTenantWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantWindows() put %q first", xs[0].ID)
	}
	if len(IndexTenantWindows(xs)) != 3 {
		t.Errorf("IndexTenantWindows() = %v", IndexTenantWindows(xs))
	}
	if err := ValidateTenantWindows(ActiveTenantWindows(xs)); err != nil {
		t.Errorf("ValidateTenantWindows() = %v", err)
	}
}

func TestTenantWindowNormalizeClamps(t *testing.T) {
	x := SampleTenantWindow("a", when)
	x.Category = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantQuotaRoundTrip(t *testing.T) {
	x := SampleTenantQuota("tenantquota-1", when)
	got, err := ParseTenantQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantQuota() = %v", err)
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

func TestTenantQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleTenantQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantQuota(f); err == nil {
		t.Error("ParseTenantQuota() accepted a record with no id")
	}
	f = SampleTenantQuota("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTenantQuota(f); err == nil {
		t.Error("ParseTenantQuota() accepted a non-numeric version")
	}
}

func TestTenantQuotaCollections(t *testing.T) {
	xs := []TenantQuota{SampleTenantQuota("a", when), SampleTenantQuota("bb", when), SampleTenantQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantQuotas(xs)); got != 2 {
		t.Errorf("ActiveTenantQuotas() kept %d, want 2", got)
	}
	if got, want := SumTenantQuotas(xs), 1+3; got != want {
		t.Errorf("SumTenantQuotas() = %d, want %d", got, want)
	}
	SortTenantQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantQuotas() put %q first", xs[0].ID)
	}
	if len(IndexTenantQuotas(xs)) != 3 {
		t.Errorf("IndexTenantQuotas() = %v", IndexTenantQuotas(xs))
	}
	if err := ValidateTenantQuotas(ActiveTenantQuotas(xs)); err != nil {
		t.Errorf("ValidateTenantQuotas() = %v", err)
	}
}

func TestTenantQuotaNormalizeClamps(t *testing.T) {
	x := SampleTenantQuota("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantBucketRoundTrip(t *testing.T) {
	x := SampleTenantBucket("tenantbucket-1", when)
	got, err := ParseTenantBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantBucket() = %v", err)
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

func TestTenantBucketParseRejectsBadInput(t *testing.T) {
	f := SampleTenantBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantBucket(f); err == nil {
		t.Error("ParseTenantBucket() accepted a record with no id")
	}
	f = SampleTenantBucket("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTenantBucket(f); err == nil {
		t.Error("ParseTenantBucket() accepted a non-numeric capacity")
	}
}

func TestTenantBucketCollections(t *testing.T) {
	xs := []TenantBucket{SampleTenantBucket("a", when), SampleTenantBucket("bb", when), SampleTenantBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantBuckets(xs)); got != 2 {
		t.Errorf("ActiveTenantBuckets() kept %d, want 2", got)
	}
	if got, want := SumTenantBuckets(xs), 1+3; got != want {
		t.Errorf("SumTenantBuckets() = %d, want %d", got, want)
	}
	SortTenantBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantBuckets() put %q first", xs[0].ID)
	}
	if len(IndexTenantBuckets(xs)) != 3 {
		t.Errorf("IndexTenantBuckets() = %v", IndexTenantBuckets(xs))
	}
	if err := ValidateTenantBuckets(ActiveTenantBuckets(xs)); err != nil {
		t.Errorf("ValidateTenantBuckets() = %v", err)
	}
}

func TestTenantBucketNormalizeClamps(t *testing.T) {
	x := SampleTenantBucket("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantLaneRoundTrip(t *testing.T) {
	x := SampleTenantLane("tenantlane-1", when)
	got, err := ParseTenantLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantLane() = %v", err)
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

func TestTenantLaneParseRejectsBadInput(t *testing.T) {
	f := SampleTenantLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantLane(f); err == nil {
		t.Error("ParseTenantLane() accepted a record with no id")
	}
	f = SampleTenantLane("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTenantLane(f); err == nil {
		t.Error("ParseTenantLane() accepted a non-numeric capacity")
	}
}

func TestTenantLaneCollections(t *testing.T) {
	xs := []TenantLane{SampleTenantLane("a", when), SampleTenantLane("bb", when), SampleTenantLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantLanes(xs)); got != 2 {
		t.Errorf("ActiveTenantLanes() kept %d, want 2", got)
	}
	if got, want := SumTenantLanes(xs), 1+3; got != want {
		t.Errorf("SumTenantLanes() = %d, want %d", got, want)
	}
	SortTenantLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantLanes() put %q first", xs[0].ID)
	}
	if len(IndexTenantLanes(xs)) != 3 {
		t.Errorf("IndexTenantLanes() = %v", IndexTenantLanes(xs))
	}
	if err := ValidateTenantLanes(ActiveTenantLanes(xs)); err != nil {
		t.Errorf("ValidateTenantLanes() = %v", err)
	}
}

func TestTenantLaneNormalizeClamps(t *testing.T) {
	x := SampleTenantLane("a", when)
	x.Label = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantSlotRoundTrip(t *testing.T) {
	x := SampleTenantSlot("tenantslot-1", when)
	got, err := ParseTenantSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantSlot() = %v", err)
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

func TestTenantSlotParseRejectsBadInput(t *testing.T) {
	f := SampleTenantSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantSlot(f); err == nil {
		t.Error("ParseTenantSlot() accepted a record with no id")
	}
	f = SampleTenantSlot("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTenantSlot(f); err == nil {
		t.Error("ParseTenantSlot() accepted a non-numeric attempts")
	}
}

func TestTenantSlotCollections(t *testing.T) {
	xs := []TenantSlot{SampleTenantSlot("a", when), SampleTenantSlot("bb", when), SampleTenantSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantSlots(xs)); got != 2 {
		t.Errorf("ActiveTenantSlots() kept %d, want 2", got)
	}
	if got, want := SumTenantSlots(xs), 1+3; got != want {
		t.Errorf("SumTenantSlots() = %d, want %d", got, want)
	}
	SortTenantSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantSlots() put %q first", xs[0].ID)
	}
	if len(IndexTenantSlots(xs)) != 3 {
		t.Errorf("IndexTenantSlots() = %v", IndexTenantSlots(xs))
	}
	if err := ValidateTenantSlots(ActiveTenantSlots(xs)); err != nil {
		t.Errorf("ValidateTenantSlots() = %v", err)
	}
}

func TestTenantSlotNormalizeClamps(t *testing.T) {
	x := SampleTenantSlot("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantTagRoundTrip(t *testing.T) {
	x := SampleTenantTag("tenanttag-1", when)
	got, err := ParseTenantTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantTag() = %v", err)
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

func TestTenantTagParseRejectsBadInput(t *testing.T) {
	f := SampleTenantTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantTag(f); err == nil {
		t.Error("ParseTenantTag() accepted a record with no id")
	}
	f = SampleTenantTag("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTenantTag(f); err == nil {
		t.Error("ParseTenantTag() accepted a non-numeric weight")
	}
}

func TestTenantTagCollections(t *testing.T) {
	xs := []TenantTag{SampleTenantTag("a", when), SampleTenantTag("bb", when), SampleTenantTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantTags(xs)); got != 2 {
		t.Errorf("ActiveTenantTags() kept %d, want 2", got)
	}
	if got, want := SumTenantTags(xs), 1+3; got != want {
		t.Errorf("SumTenantTags() = %d, want %d", got, want)
	}
	SortTenantTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantTags() put %q first", xs[0].ID)
	}
	if len(IndexTenantTags(xs)) != 3 {
		t.Errorf("IndexTenantTags() = %v", IndexTenantTags(xs))
	}
	if err := ValidateTenantTags(ActiveTenantTags(xs)); err != nil {
		t.Errorf("ValidateTenantTags() = %v", err)
	}
}

func TestTenantTagNormalizeClamps(t *testing.T) {
	x := SampleTenantTag("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantRuleRoundTrip(t *testing.T) {
	x := SampleTenantRule("tenantrule-1", when)
	got, err := ParseTenantRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantRule() = %v", err)
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

func TestTenantRuleParseRejectsBadInput(t *testing.T) {
	f := SampleTenantRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantRule(f); err == nil {
		t.Error("ParseTenantRule() accepted a record with no id")
	}
	f = SampleTenantRule("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTenantRule(f); err == nil {
		t.Error("ParseTenantRule() accepted a non-numeric attempts")
	}
}

func TestTenantRuleCollections(t *testing.T) {
	xs := []TenantRule{SampleTenantRule("a", when), SampleTenantRule("bb", when), SampleTenantRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantRules(xs)); got != 2 {
		t.Errorf("ActiveTenantRules() kept %d, want 2", got)
	}
	if got, want := SumTenantRules(xs), 1+3; got != want {
		t.Errorf("SumTenantRules() = %d, want %d", got, want)
	}
	SortTenantRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantRules() put %q first", xs[0].ID)
	}
	if len(IndexTenantRules(xs)) != 3 {
		t.Errorf("IndexTenantRules() = %v", IndexTenantRules(xs))
	}
	if err := ValidateTenantRules(ActiveTenantRules(xs)); err != nil {
		t.Errorf("ValidateTenantRules() = %v", err)
	}
}

func TestTenantRuleNormalizeClamps(t *testing.T) {
	x := SampleTenantRule("a", when)
	x.Reference = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantHoldRoundTrip(t *testing.T) {
	x := SampleTenantHold("tenanthold-1", when)
	got, err := ParseTenantHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantHold() = %v", err)
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

func TestTenantHoldParseRejectsBadInput(t *testing.T) {
	f := SampleTenantHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantHold(f); err == nil {
		t.Error("ParseTenantHold() accepted a record with no id")
	}
	f = SampleTenantHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTenantHold(f); err == nil {
		t.Error("ParseTenantHold() accepted a non-numeric weight")
	}
}

func TestTenantHoldCollections(t *testing.T) {
	xs := []TenantHold{SampleTenantHold("a", when), SampleTenantHold("bb", when), SampleTenantHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantHolds(xs)); got != 2 {
		t.Errorf("ActiveTenantHolds() kept %d, want 2", got)
	}
	if got, want := SumTenantHolds(xs), 1+3; got != want {
		t.Errorf("SumTenantHolds() = %d, want %d", got, want)
	}
	SortTenantHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantHolds() put %q first", xs[0].ID)
	}
	if len(IndexTenantHolds(xs)) != 3 {
		t.Errorf("IndexTenantHolds() = %v", IndexTenantHolds(xs))
	}
	if err := ValidateTenantHolds(ActiveTenantHolds(xs)); err != nil {
		t.Errorf("ValidateTenantHolds() = %v", err)
	}
}

func TestTenantHoldNormalizeClamps(t *testing.T) {
	x := SampleTenantHold("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantClaimRoundTrip(t *testing.T) {
	x := SampleTenantClaim("tenantclaim-1", when)
	got, err := ParseTenantClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantClaim() = %v", err)
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

func TestTenantClaimParseRejectsBadInput(t *testing.T) {
	f := SampleTenantClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantClaim(f); err == nil {
		t.Error("ParseTenantClaim() accepted a record with no id")
	}
	f = SampleTenantClaim("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTenantClaim(f); err == nil {
		t.Error("ParseTenantClaim() accepted a non-numeric offset")
	}
}

func TestTenantClaimCollections(t *testing.T) {
	xs := []TenantClaim{SampleTenantClaim("a", when), SampleTenantClaim("bb", when), SampleTenantClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantClaims(xs)); got != 2 {
		t.Errorf("ActiveTenantClaims() kept %d, want 2", got)
	}
	if got, want := SumTenantClaims(xs), 1+3; got != want {
		t.Errorf("SumTenantClaims() = %d, want %d", got, want)
	}
	SortTenantClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantClaims() put %q first", xs[0].ID)
	}
	if len(IndexTenantClaims(xs)) != 3 {
		t.Errorf("IndexTenantClaims() = %v", IndexTenantClaims(xs))
	}
	if err := ValidateTenantClaims(ActiveTenantClaims(xs)); err != nil {
		t.Errorf("ValidateTenantClaims() = %v", err)
	}
}

func TestTenantClaimNormalizeClamps(t *testing.T) {
	x := SampleTenantClaim("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTenantCursorRoundTrip(t *testing.T) {
	x := SampleTenantCursor("tenantcursor-1", when)
	got, err := ParseTenantCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseTenantCursor() = %v", err)
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

func TestTenantCursorParseRejectsBadInput(t *testing.T) {
	f := SampleTenantCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTenantCursor(f); err == nil {
		t.Error("ParseTenantCursor() accepted a record with no id")
	}
	f = SampleTenantCursor("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTenantCursor(f); err == nil {
		t.Error("ParseTenantCursor() accepted a non-numeric priority")
	}
}

func TestTenantCursorCollections(t *testing.T) {
	xs := []TenantCursor{SampleTenantCursor("a", when), SampleTenantCursor("bb", when), SampleTenantCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTenantCursors(xs)); got != 2 {
		t.Errorf("ActiveTenantCursors() kept %d, want 2", got)
	}
	if got, want := SumTenantCursors(xs), 1+3; got != want {
		t.Errorf("SumTenantCursors() = %d, want %d", got, want)
	}
	SortTenantCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTenantCursors() put %q first", xs[0].ID)
	}
	if len(IndexTenantCursors(xs)) != 3 {
		t.Errorf("IndexTenantCursors() = %v", IndexTenantCursors(xs))
	}
	if err := ValidateTenantCursors(ActiveTenantCursors(xs)); err != nil {
		t.Errorf("ValidateTenantCursors() = %v", err)
	}
}

func TestTenantCursorNormalizeClamps(t *testing.T) {
	x := SampleTenantCursor("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
