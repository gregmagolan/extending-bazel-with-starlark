package ledger

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
		if !strings.HasPrefix(Prefixes[i], "ledger/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleLedgerRecord("ok", when).Fields()); err != nil {
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

func TestLedgerRecordRoundTrip(t *testing.T) {
	x := SampleLedgerRecord("ledgerrecord-1", when)
	got, err := ParseLedgerRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerRecord() = %v", err)
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

func TestLedgerRecordParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerRecord(f); err == nil {
		t.Error("ParseLedgerRecord() accepted a record with no id")
	}
	f = SampleLedgerRecord("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLedgerRecord(f); err == nil {
		t.Error("ParseLedgerRecord() accepted a non-numeric retained")
	}
}

func TestLedgerRecordCollections(t *testing.T) {
	xs := []LedgerRecord{SampleLedgerRecord("a", when), SampleLedgerRecord("bb", when), SampleLedgerRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerRecords(xs)); got != 2 {
		t.Errorf("ActiveLedgerRecords() kept %d, want 2", got)
	}
	if got, want := SumLedgerRecords(xs), 1+3; got != want {
		t.Errorf("SumLedgerRecords() = %d, want %d", got, want)
	}
	SortLedgerRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerRecords() put %q first", xs[0].ID)
	}
	if len(IndexLedgerRecords(xs)) != 3 {
		t.Errorf("IndexLedgerRecords() = %v", IndexLedgerRecords(xs))
	}
	if err := ValidateLedgerRecords(ActiveLedgerRecords(xs)); err != nil {
		t.Errorf("ValidateLedgerRecords() = %v", err)
	}
}

func TestLedgerRecordNormalizeClamps(t *testing.T) {
	x := SampleLedgerRecord("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerEntryRoundTrip(t *testing.T) {
	x := SampleLedgerEntry("ledgerentry-1", when)
	got, err := ParseLedgerEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerEntry() = %v", err)
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

func TestLedgerEntryParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerEntry(f); err == nil {
		t.Error("ParseLedgerEntry() accepted a record with no id")
	}
	f = SampleLedgerEntry("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseLedgerEntry(f); err == nil {
		t.Error("ParseLedgerEntry() accepted a non-numeric weight")
	}
}

func TestLedgerEntryCollections(t *testing.T) {
	xs := []LedgerEntry{SampleLedgerEntry("a", when), SampleLedgerEntry("bb", when), SampleLedgerEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerEntrys(xs)); got != 2 {
		t.Errorf("ActiveLedgerEntrys() kept %d, want 2", got)
	}
	if got, want := SumLedgerEntrys(xs), 1+3; got != want {
		t.Errorf("SumLedgerEntrys() = %d, want %d", got, want)
	}
	SortLedgerEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerEntrys() put %q first", xs[0].ID)
	}
	if len(IndexLedgerEntrys(xs)) != 3 {
		t.Errorf("IndexLedgerEntrys() = %v", IndexLedgerEntrys(xs))
	}
	if err := ValidateLedgerEntrys(ActiveLedgerEntrys(xs)); err != nil {
		t.Errorf("ValidateLedgerEntrys() = %v", err)
	}
}

func TestLedgerEntryNormalizeClamps(t *testing.T) {
	x := SampleLedgerEntry("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerItemRoundTrip(t *testing.T) {
	x := SampleLedgerItem("ledgeritem-1", when)
	got, err := ParseLedgerItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerItem() = %v", err)
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

func TestLedgerItemParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerItem(f); err == nil {
		t.Error("ParseLedgerItem() accepted a record with no id")
	}
	f = SampleLedgerItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseLedgerItem(f); err == nil {
		t.Error("ParseLedgerItem() accepted a non-numeric version")
	}
}

func TestLedgerItemCollections(t *testing.T) {
	xs := []LedgerItem{SampleLedgerItem("a", when), SampleLedgerItem("bb", when), SampleLedgerItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerItems(xs)); got != 2 {
		t.Errorf("ActiveLedgerItems() kept %d, want 2", got)
	}
	if got, want := SumLedgerItems(xs), 1+3; got != want {
		t.Errorf("SumLedgerItems() = %d, want %d", got, want)
	}
	SortLedgerItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerItems() put %q first", xs[0].ID)
	}
	if len(IndexLedgerItems(xs)) != 3 {
		t.Errorf("IndexLedgerItems() = %v", IndexLedgerItems(xs))
	}
	if err := ValidateLedgerItems(ActiveLedgerItems(xs)); err != nil {
		t.Errorf("ValidateLedgerItems() = %v", err)
	}
}

func TestLedgerItemNormalizeClamps(t *testing.T) {
	x := SampleLedgerItem("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerDraftRoundTrip(t *testing.T) {
	x := SampleLedgerDraft("ledgerdraft-1", when)
	got, err := ParseLedgerDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerDraft() = %v", err)
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

func TestLedgerDraftParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerDraft(f); err == nil {
		t.Error("ParseLedgerDraft() accepted a record with no id")
	}
	f = SampleLedgerDraft("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseLedgerDraft(f); err == nil {
		t.Error("ParseLedgerDraft() accepted a non-numeric weight")
	}
}

func TestLedgerDraftCollections(t *testing.T) {
	xs := []LedgerDraft{SampleLedgerDraft("a", when), SampleLedgerDraft("bb", when), SampleLedgerDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerDrafts(xs)); got != 2 {
		t.Errorf("ActiveLedgerDrafts() kept %d, want 2", got)
	}
	if got, want := SumLedgerDrafts(xs), 1+3; got != want {
		t.Errorf("SumLedgerDrafts() = %d, want %d", got, want)
	}
	SortLedgerDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerDrafts() put %q first", xs[0].ID)
	}
	if len(IndexLedgerDrafts(xs)) != 3 {
		t.Errorf("IndexLedgerDrafts() = %v", IndexLedgerDrafts(xs))
	}
	if err := ValidateLedgerDrafts(ActiveLedgerDrafts(xs)); err != nil {
		t.Errorf("ValidateLedgerDrafts() = %v", err)
	}
}

func TestLedgerDraftNormalizeClamps(t *testing.T) {
	x := SampleLedgerDraft("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerRevisionRoundTrip(t *testing.T) {
	x := SampleLedgerRevision("ledgerrevision-1", when)
	got, err := ParseLedgerRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerRevision() = %v", err)
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

func TestLedgerRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerRevision(f); err == nil {
		t.Error("ParseLedgerRevision() accepted a record with no id")
	}
	f = SampleLedgerRevision("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseLedgerRevision(f); err == nil {
		t.Error("ParseLedgerRevision() accepted a non-numeric priority")
	}
}

func TestLedgerRevisionCollections(t *testing.T) {
	xs := []LedgerRevision{SampleLedgerRevision("a", when), SampleLedgerRevision("bb", when), SampleLedgerRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerRevisions(xs)); got != 2 {
		t.Errorf("ActiveLedgerRevisions() kept %d, want 2", got)
	}
	if got, want := SumLedgerRevisions(xs), 1+3; got != want {
		t.Errorf("SumLedgerRevisions() = %d, want %d", got, want)
	}
	SortLedgerRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerRevisions() put %q first", xs[0].ID)
	}
	if len(IndexLedgerRevisions(xs)) != 3 {
		t.Errorf("IndexLedgerRevisions() = %v", IndexLedgerRevisions(xs))
	}
	if err := ValidateLedgerRevisions(ActiveLedgerRevisions(xs)); err != nil {
		t.Errorf("ValidateLedgerRevisions() = %v", err)
	}
}

func TestLedgerRevisionNormalizeClamps(t *testing.T) {
	x := SampleLedgerRevision("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerSnapshotRoundTrip(t *testing.T) {
	x := SampleLedgerSnapshot("ledgersnapshot-1", when)
	got, err := ParseLedgerSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerSnapshot() = %v", err)
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

func TestLedgerSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerSnapshot(f); err == nil {
		t.Error("ParseLedgerSnapshot() accepted a record with no id")
	}
	f = SampleLedgerSnapshot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLedgerSnapshot(f); err == nil {
		t.Error("ParseLedgerSnapshot() accepted a non-numeric retained")
	}
}

func TestLedgerSnapshotCollections(t *testing.T) {
	xs := []LedgerSnapshot{SampleLedgerSnapshot("a", when), SampleLedgerSnapshot("bb", when), SampleLedgerSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerSnapshots(xs)); got != 2 {
		t.Errorf("ActiveLedgerSnapshots() kept %d, want 2", got)
	}
	if got, want := SumLedgerSnapshots(xs), 1+3; got != want {
		t.Errorf("SumLedgerSnapshots() = %d, want %d", got, want)
	}
	SortLedgerSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexLedgerSnapshots(xs)) != 3 {
		t.Errorf("IndexLedgerSnapshots() = %v", IndexLedgerSnapshots(xs))
	}
	if err := ValidateLedgerSnapshots(ActiveLedgerSnapshots(xs)); err != nil {
		t.Errorf("ValidateLedgerSnapshots() = %v", err)
	}
}

func TestLedgerSnapshotNormalizeClamps(t *testing.T) {
	x := SampleLedgerSnapshot("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerBatchRoundTrip(t *testing.T) {
	x := SampleLedgerBatch("ledgerbatch-1", when)
	got, err := ParseLedgerBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerBatch() = %v", err)
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

func TestLedgerBatchParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerBatch(f); err == nil {
		t.Error("ParseLedgerBatch() accepted a record with no id")
	}
	f = SampleLedgerBatch("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLedgerBatch(f); err == nil {
		t.Error("ParseLedgerBatch() accepted a non-numeric retained")
	}
}

func TestLedgerBatchCollections(t *testing.T) {
	xs := []LedgerBatch{SampleLedgerBatch("a", when), SampleLedgerBatch("bb", when), SampleLedgerBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerBatchs(xs)); got != 2 {
		t.Errorf("ActiveLedgerBatchs() kept %d, want 2", got)
	}
	if got, want := SumLedgerBatchs(xs), 1+3; got != want {
		t.Errorf("SumLedgerBatchs() = %d, want %d", got, want)
	}
	SortLedgerBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerBatchs() put %q first", xs[0].ID)
	}
	if len(IndexLedgerBatchs(xs)) != 3 {
		t.Errorf("IndexLedgerBatchs() = %v", IndexLedgerBatchs(xs))
	}
	if err := ValidateLedgerBatchs(ActiveLedgerBatchs(xs)); err != nil {
		t.Errorf("ValidateLedgerBatchs() = %v", err)
	}
}

func TestLedgerBatchNormalizeClamps(t *testing.T) {
	x := SampleLedgerBatch("a", when)
	x.Reference = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerRequestRoundTrip(t *testing.T) {
	x := SampleLedgerRequest("ledgerrequest-1", when)
	got, err := ParseLedgerRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerRequest() = %v", err)
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

func TestLedgerRequestParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerRequest(f); err == nil {
		t.Error("ParseLedgerRequest() accepted a record with no id")
	}
	f = SampleLedgerRequest("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseLedgerRequest(f); err == nil {
		t.Error("ParseLedgerRequest() accepted a non-numeric count")
	}
}

func TestLedgerRequestCollections(t *testing.T) {
	xs := []LedgerRequest{SampleLedgerRequest("a", when), SampleLedgerRequest("bb", when), SampleLedgerRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerRequests(xs)); got != 2 {
		t.Errorf("ActiveLedgerRequests() kept %d, want 2", got)
	}
	if got, want := SumLedgerRequests(xs), 1+3; got != want {
		t.Errorf("SumLedgerRequests() = %d, want %d", got, want)
	}
	SortLedgerRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerRequests() put %q first", xs[0].ID)
	}
	if len(IndexLedgerRequests(xs)) != 3 {
		t.Errorf("IndexLedgerRequests() = %v", IndexLedgerRequests(xs))
	}
	if err := ValidateLedgerRequests(ActiveLedgerRequests(xs)); err != nil {
		t.Errorf("ValidateLedgerRequests() = %v", err)
	}
}

func TestLedgerRequestNormalizeClamps(t *testing.T) {
	x := SampleLedgerRequest("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerSummaryRoundTrip(t *testing.T) {
	x := SampleLedgerSummary("ledgersummary-1", when)
	got, err := ParseLedgerSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerSummary() = %v", err)
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

func TestLedgerSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerSummary(f); err == nil {
		t.Error("ParseLedgerSummary() accepted a record with no id")
	}
	f = SampleLedgerSummary("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseLedgerSummary(f); err == nil {
		t.Error("ParseLedgerSummary() accepted a non-numeric offset")
	}
}

func TestLedgerSummaryCollections(t *testing.T) {
	xs := []LedgerSummary{SampleLedgerSummary("a", when), SampleLedgerSummary("bb", when), SampleLedgerSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerSummarys(xs)); got != 2 {
		t.Errorf("ActiveLedgerSummarys() kept %d, want 2", got)
	}
	if got, want := SumLedgerSummarys(xs), 1+3; got != want {
		t.Errorf("SumLedgerSummarys() = %d, want %d", got, want)
	}
	SortLedgerSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerSummarys() put %q first", xs[0].ID)
	}
	if len(IndexLedgerSummarys(xs)) != 3 {
		t.Errorf("IndexLedgerSummarys() = %v", IndexLedgerSummarys(xs))
	}
	if err := ValidateLedgerSummarys(ActiveLedgerSummarys(xs)); err != nil {
		t.Errorf("ValidateLedgerSummarys() = %v", err)
	}
}

func TestLedgerSummaryNormalizeClamps(t *testing.T) {
	x := SampleLedgerSummary("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerPolicyRoundTrip(t *testing.T) {
	x := SampleLedgerPolicy("ledgerpolicy-1", when)
	got, err := ParseLedgerPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerPolicy() = %v", err)
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

func TestLedgerPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerPolicy(f); err == nil {
		t.Error("ParseLedgerPolicy() accepted a record with no id")
	}
	f = SampleLedgerPolicy("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLedgerPolicy(f); err == nil {
		t.Error("ParseLedgerPolicy() accepted a non-numeric attempts")
	}
}

func TestLedgerPolicyCollections(t *testing.T) {
	xs := []LedgerPolicy{SampleLedgerPolicy("a", when), SampleLedgerPolicy("bb", when), SampleLedgerPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerPolicys(xs)); got != 2 {
		t.Errorf("ActiveLedgerPolicys() kept %d, want 2", got)
	}
	if got, want := SumLedgerPolicys(xs), 1+3; got != want {
		t.Errorf("SumLedgerPolicys() = %d, want %d", got, want)
	}
	SortLedgerPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerPolicys() put %q first", xs[0].ID)
	}
	if len(IndexLedgerPolicys(xs)) != 3 {
		t.Errorf("IndexLedgerPolicys() = %v", IndexLedgerPolicys(xs))
	}
	if err := ValidateLedgerPolicys(ActiveLedgerPolicys(xs)); err != nil {
		t.Errorf("ValidateLedgerPolicys() = %v", err)
	}
}

func TestLedgerPolicyNormalizeClamps(t *testing.T) {
	x := SampleLedgerPolicy("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerLinkRoundTrip(t *testing.T) {
	x := SampleLedgerLink("ledgerlink-1", when)
	got, err := ParseLedgerLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerLink() = %v", err)
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

func TestLedgerLinkParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerLink(f); err == nil {
		t.Error("ParseLedgerLink() accepted a record with no id")
	}
	f = SampleLedgerLink("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseLedgerLink(f); err == nil {
		t.Error("ParseLedgerLink() accepted a non-numeric count")
	}
}

func TestLedgerLinkCollections(t *testing.T) {
	xs := []LedgerLink{SampleLedgerLink("a", when), SampleLedgerLink("bb", when), SampleLedgerLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerLinks(xs)); got != 2 {
		t.Errorf("ActiveLedgerLinks() kept %d, want 2", got)
	}
	if got, want := SumLedgerLinks(xs), 1+3; got != want {
		t.Errorf("SumLedgerLinks() = %d, want %d", got, want)
	}
	SortLedgerLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerLinks() put %q first", xs[0].ID)
	}
	if len(IndexLedgerLinks(xs)) != 3 {
		t.Errorf("IndexLedgerLinks() = %v", IndexLedgerLinks(xs))
	}
	if err := ValidateLedgerLinks(ActiveLedgerLinks(xs)); err != nil {
		t.Errorf("ValidateLedgerLinks() = %v", err)
	}
}

func TestLedgerLinkNormalizeClamps(t *testing.T) {
	x := SampleLedgerLink("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerNoteRoundTrip(t *testing.T) {
	x := SampleLedgerNote("ledgernote-1", when)
	got, err := ParseLedgerNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerNote() = %v", err)
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

func TestLedgerNoteParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerNote(f); err == nil {
		t.Error("ParseLedgerNote() accepted a record with no id")
	}
	f = SampleLedgerNote("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLedgerNote(f); err == nil {
		t.Error("ParseLedgerNote() accepted a non-numeric attempts")
	}
}

func TestLedgerNoteCollections(t *testing.T) {
	xs := []LedgerNote{SampleLedgerNote("a", when), SampleLedgerNote("bb", when), SampleLedgerNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerNotes(xs)); got != 2 {
		t.Errorf("ActiveLedgerNotes() kept %d, want 2", got)
	}
	if got, want := SumLedgerNotes(xs), 1+3; got != want {
		t.Errorf("SumLedgerNotes() = %d, want %d", got, want)
	}
	SortLedgerNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerNotes() put %q first", xs[0].ID)
	}
	if len(IndexLedgerNotes(xs)) != 3 {
		t.Errorf("IndexLedgerNotes() = %v", IndexLedgerNotes(xs))
	}
	if err := ValidateLedgerNotes(ActiveLedgerNotes(xs)); err != nil {
		t.Errorf("ValidateLedgerNotes() = %v", err)
	}
}

func TestLedgerNoteNormalizeClamps(t *testing.T) {
	x := SampleLedgerNote("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerWindowRoundTrip(t *testing.T) {
	x := SampleLedgerWindow("ledgerwindow-1", when)
	got, err := ParseLedgerWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerWindow() = %v", err)
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

func TestLedgerWindowParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerWindow(f); err == nil {
		t.Error("ParseLedgerWindow() accepted a record with no id")
	}
	f = SampleLedgerWindow("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseLedgerWindow(f); err == nil {
		t.Error("ParseLedgerWindow() accepted a non-numeric count")
	}
}

func TestLedgerWindowCollections(t *testing.T) {
	xs := []LedgerWindow{SampleLedgerWindow("a", when), SampleLedgerWindow("bb", when), SampleLedgerWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerWindows(xs)); got != 2 {
		t.Errorf("ActiveLedgerWindows() kept %d, want 2", got)
	}
	if got, want := SumLedgerWindows(xs), 1+3; got != want {
		t.Errorf("SumLedgerWindows() = %d, want %d", got, want)
	}
	SortLedgerWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerWindows() put %q first", xs[0].ID)
	}
	if len(IndexLedgerWindows(xs)) != 3 {
		t.Errorf("IndexLedgerWindows() = %v", IndexLedgerWindows(xs))
	}
	if err := ValidateLedgerWindows(ActiveLedgerWindows(xs)); err != nil {
		t.Errorf("ValidateLedgerWindows() = %v", err)
	}
}

func TestLedgerWindowNormalizeClamps(t *testing.T) {
	x := SampleLedgerWindow("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerQuotaRoundTrip(t *testing.T) {
	x := SampleLedgerQuota("ledgerquota-1", when)
	got, err := ParseLedgerQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerQuota() = %v", err)
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

func TestLedgerQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerQuota(f); err == nil {
		t.Error("ParseLedgerQuota() accepted a record with no id")
	}
	f = SampleLedgerQuota("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseLedgerQuota(f); err == nil {
		t.Error("ParseLedgerQuota() accepted a non-numeric capacity")
	}
}

func TestLedgerQuotaCollections(t *testing.T) {
	xs := []LedgerQuota{SampleLedgerQuota("a", when), SampleLedgerQuota("bb", when), SampleLedgerQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerQuotas(xs)); got != 2 {
		t.Errorf("ActiveLedgerQuotas() kept %d, want 2", got)
	}
	if got, want := SumLedgerQuotas(xs), 1+3; got != want {
		t.Errorf("SumLedgerQuotas() = %d, want %d", got, want)
	}
	SortLedgerQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerQuotas() put %q first", xs[0].ID)
	}
	if len(IndexLedgerQuotas(xs)) != 3 {
		t.Errorf("IndexLedgerQuotas() = %v", IndexLedgerQuotas(xs))
	}
	if err := ValidateLedgerQuotas(ActiveLedgerQuotas(xs)); err != nil {
		t.Errorf("ValidateLedgerQuotas() = %v", err)
	}
}

func TestLedgerQuotaNormalizeClamps(t *testing.T) {
	x := SampleLedgerQuota("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerBucketRoundTrip(t *testing.T) {
	x := SampleLedgerBucket("ledgerbucket-1", when)
	got, err := ParseLedgerBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerBucket() = %v", err)
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

func TestLedgerBucketParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerBucket(f); err == nil {
		t.Error("ParseLedgerBucket() accepted a record with no id")
	}
	f = SampleLedgerBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLedgerBucket(f); err == nil {
		t.Error("ParseLedgerBucket() accepted a non-numeric retained")
	}
}

func TestLedgerBucketCollections(t *testing.T) {
	xs := []LedgerBucket{SampleLedgerBucket("a", when), SampleLedgerBucket("bb", when), SampleLedgerBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerBuckets(xs)); got != 2 {
		t.Errorf("ActiveLedgerBuckets() kept %d, want 2", got)
	}
	if got, want := SumLedgerBuckets(xs), 1+3; got != want {
		t.Errorf("SumLedgerBuckets() = %d, want %d", got, want)
	}
	SortLedgerBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerBuckets() put %q first", xs[0].ID)
	}
	if len(IndexLedgerBuckets(xs)) != 3 {
		t.Errorf("IndexLedgerBuckets() = %v", IndexLedgerBuckets(xs))
	}
	if err := ValidateLedgerBuckets(ActiveLedgerBuckets(xs)); err != nil {
		t.Errorf("ValidateLedgerBuckets() = %v", err)
	}
}

func TestLedgerBucketNormalizeClamps(t *testing.T) {
	x := SampleLedgerBucket("a", when)
	x.Reference = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerLaneRoundTrip(t *testing.T) {
	x := SampleLedgerLane("ledgerlane-1", when)
	got, err := ParseLedgerLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerLane() = %v", err)
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

func TestLedgerLaneParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerLane(f); err == nil {
		t.Error("ParseLedgerLane() accepted a record with no id")
	}
	f = SampleLedgerLane("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLedgerLane(f); err == nil {
		t.Error("ParseLedgerLane() accepted a non-numeric retained")
	}
}

func TestLedgerLaneCollections(t *testing.T) {
	xs := []LedgerLane{SampleLedgerLane("a", when), SampleLedgerLane("bb", when), SampleLedgerLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerLanes(xs)); got != 2 {
		t.Errorf("ActiveLedgerLanes() kept %d, want 2", got)
	}
	if got, want := SumLedgerLanes(xs), 1+3; got != want {
		t.Errorf("SumLedgerLanes() = %d, want %d", got, want)
	}
	SortLedgerLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerLanes() put %q first", xs[0].ID)
	}
	if len(IndexLedgerLanes(xs)) != 3 {
		t.Errorf("IndexLedgerLanes() = %v", IndexLedgerLanes(xs))
	}
	if err := ValidateLedgerLanes(ActiveLedgerLanes(xs)); err != nil {
		t.Errorf("ValidateLedgerLanes() = %v", err)
	}
}

func TestLedgerLaneNormalizeClamps(t *testing.T) {
	x := SampleLedgerLane("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerSlotRoundTrip(t *testing.T) {
	x := SampleLedgerSlot("ledgerslot-1", when)
	got, err := ParseLedgerSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerSlot() = %v", err)
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

func TestLedgerSlotParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerSlot(f); err == nil {
		t.Error("ParseLedgerSlot() accepted a record with no id")
	}
	f = SampleLedgerSlot("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseLedgerSlot(f); err == nil {
		t.Error("ParseLedgerSlot() accepted a non-numeric version")
	}
}

func TestLedgerSlotCollections(t *testing.T) {
	xs := []LedgerSlot{SampleLedgerSlot("a", when), SampleLedgerSlot("bb", when), SampleLedgerSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerSlots(xs)); got != 2 {
		t.Errorf("ActiveLedgerSlots() kept %d, want 2", got)
	}
	if got, want := SumLedgerSlots(xs), 1+3; got != want {
		t.Errorf("SumLedgerSlots() = %d, want %d", got, want)
	}
	SortLedgerSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerSlots() put %q first", xs[0].ID)
	}
	if len(IndexLedgerSlots(xs)) != 3 {
		t.Errorf("IndexLedgerSlots() = %v", IndexLedgerSlots(xs))
	}
	if err := ValidateLedgerSlots(ActiveLedgerSlots(xs)); err != nil {
		t.Errorf("ValidateLedgerSlots() = %v", err)
	}
}

func TestLedgerSlotNormalizeClamps(t *testing.T) {
	x := SampleLedgerSlot("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerTagRoundTrip(t *testing.T) {
	x := SampleLedgerTag("ledgertag-1", when)
	got, err := ParseLedgerTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerTag() = %v", err)
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

func TestLedgerTagParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerTag(f); err == nil {
		t.Error("ParseLedgerTag() accepted a record with no id")
	}
	f = SampleLedgerTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseLedgerTag(f); err == nil {
		t.Error("ParseLedgerTag() accepted a non-numeric count")
	}
}

func TestLedgerTagCollections(t *testing.T) {
	xs := []LedgerTag{SampleLedgerTag("a", when), SampleLedgerTag("bb", when), SampleLedgerTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerTags(xs)); got != 2 {
		t.Errorf("ActiveLedgerTags() kept %d, want 2", got)
	}
	if got, want := SumLedgerTags(xs), 1+3; got != want {
		t.Errorf("SumLedgerTags() = %d, want %d", got, want)
	}
	SortLedgerTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerTags() put %q first", xs[0].ID)
	}
	if len(IndexLedgerTags(xs)) != 3 {
		t.Errorf("IndexLedgerTags() = %v", IndexLedgerTags(xs))
	}
	if err := ValidateLedgerTags(ActiveLedgerTags(xs)); err != nil {
		t.Errorf("ValidateLedgerTags() = %v", err)
	}
}

func TestLedgerTagNormalizeClamps(t *testing.T) {
	x := SampleLedgerTag("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerRuleRoundTrip(t *testing.T) {
	x := SampleLedgerRule("ledgerrule-1", when)
	got, err := ParseLedgerRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerRule() = %v", err)
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

func TestLedgerRuleParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerRule(f); err == nil {
		t.Error("ParseLedgerRule() accepted a record with no id")
	}
	f = SampleLedgerRule("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLedgerRule(f); err == nil {
		t.Error("ParseLedgerRule() accepted a non-numeric attempts")
	}
}

func TestLedgerRuleCollections(t *testing.T) {
	xs := []LedgerRule{SampleLedgerRule("a", when), SampleLedgerRule("bb", when), SampleLedgerRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerRules(xs)); got != 2 {
		t.Errorf("ActiveLedgerRules() kept %d, want 2", got)
	}
	if got, want := SumLedgerRules(xs), 1+3; got != want {
		t.Errorf("SumLedgerRules() = %d, want %d", got, want)
	}
	SortLedgerRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerRules() put %q first", xs[0].ID)
	}
	if len(IndexLedgerRules(xs)) != 3 {
		t.Errorf("IndexLedgerRules() = %v", IndexLedgerRules(xs))
	}
	if err := ValidateLedgerRules(ActiveLedgerRules(xs)); err != nil {
		t.Errorf("ValidateLedgerRules() = %v", err)
	}
}

func TestLedgerRuleNormalizeClamps(t *testing.T) {
	x := SampleLedgerRule("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerHoldRoundTrip(t *testing.T) {
	x := SampleLedgerHold("ledgerhold-1", when)
	got, err := ParseLedgerHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerHold() = %v", err)
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

func TestLedgerHoldParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerHold(f); err == nil {
		t.Error("ParseLedgerHold() accepted a record with no id")
	}
	f = SampleLedgerHold("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseLedgerHold(f); err == nil {
		t.Error("ParseLedgerHold() accepted a non-numeric count")
	}
}

func TestLedgerHoldCollections(t *testing.T) {
	xs := []LedgerHold{SampleLedgerHold("a", when), SampleLedgerHold("bb", when), SampleLedgerHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerHolds(xs)); got != 2 {
		t.Errorf("ActiveLedgerHolds() kept %d, want 2", got)
	}
	if got, want := SumLedgerHolds(xs), 1+3; got != want {
		t.Errorf("SumLedgerHolds() = %d, want %d", got, want)
	}
	SortLedgerHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerHolds() put %q first", xs[0].ID)
	}
	if len(IndexLedgerHolds(xs)) != 3 {
		t.Errorf("IndexLedgerHolds() = %v", IndexLedgerHolds(xs))
	}
	if err := ValidateLedgerHolds(ActiveLedgerHolds(xs)); err != nil {
		t.Errorf("ValidateLedgerHolds() = %v", err)
	}
}

func TestLedgerHoldNormalizeClamps(t *testing.T) {
	x := SampleLedgerHold("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerClaimRoundTrip(t *testing.T) {
	x := SampleLedgerClaim("ledgerclaim-1", when)
	got, err := ParseLedgerClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerClaim() = %v", err)
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

func TestLedgerClaimParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerClaim(f); err == nil {
		t.Error("ParseLedgerClaim() accepted a record with no id")
	}
	f = SampleLedgerClaim("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseLedgerClaim(f); err == nil {
		t.Error("ParseLedgerClaim() accepted a non-numeric weight")
	}
}

func TestLedgerClaimCollections(t *testing.T) {
	xs := []LedgerClaim{SampleLedgerClaim("a", when), SampleLedgerClaim("bb", when), SampleLedgerClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerClaims(xs)); got != 2 {
		t.Errorf("ActiveLedgerClaims() kept %d, want 2", got)
	}
	if got, want := SumLedgerClaims(xs), 1+3; got != want {
		t.Errorf("SumLedgerClaims() = %d, want %d", got, want)
	}
	SortLedgerClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerClaims() put %q first", xs[0].ID)
	}
	if len(IndexLedgerClaims(xs)) != 3 {
		t.Errorf("IndexLedgerClaims() = %v", IndexLedgerClaims(xs))
	}
	if err := ValidateLedgerClaims(ActiveLedgerClaims(xs)); err != nil {
		t.Errorf("ValidateLedgerClaims() = %v", err)
	}
}

func TestLedgerClaimNormalizeClamps(t *testing.T) {
	x := SampleLedgerClaim("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLedgerCursorRoundTrip(t *testing.T) {
	x := SampleLedgerCursor("ledgercursor-1", when)
	got, err := ParseLedgerCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseLedgerCursor() = %v", err)
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

func TestLedgerCursorParseRejectsBadInput(t *testing.T) {
	f := SampleLedgerCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLedgerCursor(f); err == nil {
		t.Error("ParseLedgerCursor() accepted a record with no id")
	}
	f = SampleLedgerCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseLedgerCursor(f); err == nil {
		t.Error("ParseLedgerCursor() accepted a non-numeric version")
	}
}

func TestLedgerCursorCollections(t *testing.T) {
	xs := []LedgerCursor{SampleLedgerCursor("a", when), SampleLedgerCursor("bb", when), SampleLedgerCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLedgerCursors(xs)); got != 2 {
		t.Errorf("ActiveLedgerCursors() kept %d, want 2", got)
	}
	if got, want := SumLedgerCursors(xs), 1+3; got != want {
		t.Errorf("SumLedgerCursors() = %d, want %d", got, want)
	}
	SortLedgerCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLedgerCursors() put %q first", xs[0].ID)
	}
	if len(IndexLedgerCursors(xs)) != 3 {
		t.Errorf("IndexLedgerCursors() = %v", IndexLedgerCursors(xs))
	}
	if err := ValidateLedgerCursors(ActiveLedgerCursors(xs)); err != nil {
		t.Errorf("ValidateLedgerCursors() = %v", err)
	}
}

func TestLedgerCursorNormalizeClamps(t *testing.T) {
	x := SampleLedgerCursor("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
