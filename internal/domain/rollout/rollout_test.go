package rollout

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
		if !strings.HasPrefix(Prefixes[i], "rollout/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleRolloutRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestRolloutRecordRoundTrip(t *testing.T) {
	x := SampleRolloutRecord("rolloutrecord-1", when)
	got, err := ParseRolloutRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutRecord() = %v", err)
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

func TestRolloutRecordParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutRecord(f); err == nil {
		t.Error("ParseRolloutRecord() accepted a record with no id")
	}
	f = SampleRolloutRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRolloutRecord(f); err == nil {
		t.Error("ParseRolloutRecord() accepted a non-numeric capacity")
	}
}

func TestRolloutRecordCollections(t *testing.T) {
	xs := []RolloutRecord{SampleRolloutRecord("a", when), SampleRolloutRecord("bb", when), SampleRolloutRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutRecords(xs)); got != 2 {
		t.Errorf("ActiveRolloutRecords() kept %d, want 2", got)
	}
	if got, want := SumRolloutRecords(xs), 1+3; got != want {
		t.Errorf("SumRolloutRecords() = %d, want %d", got, want)
	}
	SortRolloutRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutRecords() put %q first", xs[0].ID)
	}
	if len(IndexRolloutRecords(xs)) != 3 {
		t.Errorf("IndexRolloutRecords() = %v", IndexRolloutRecords(xs))
	}
	if err := ValidateRolloutRecords(ActiveRolloutRecords(xs)); err != nil {
		t.Errorf("ValidateRolloutRecords() = %v", err)
	}
}

func TestRolloutRecordNormalizeClamps(t *testing.T) {
	x := SampleRolloutRecord("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutEntryRoundTrip(t *testing.T) {
	x := SampleRolloutEntry("rolloutentry-1", when)
	got, err := ParseRolloutEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutEntry() = %v", err)
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

func TestRolloutEntryParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutEntry(f); err == nil {
		t.Error("ParseRolloutEntry() accepted a record with no id")
	}
	f = SampleRolloutEntry("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseRolloutEntry(f); err == nil {
		t.Error("ParseRolloutEntry() accepted a non-numeric attempts")
	}
}

func TestRolloutEntryCollections(t *testing.T) {
	xs := []RolloutEntry{SampleRolloutEntry("a", when), SampleRolloutEntry("bb", when), SampleRolloutEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutEntrys(xs)); got != 2 {
		t.Errorf("ActiveRolloutEntrys() kept %d, want 2", got)
	}
	if got, want := SumRolloutEntrys(xs), 1+3; got != want {
		t.Errorf("SumRolloutEntrys() = %d, want %d", got, want)
	}
	SortRolloutEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutEntrys() put %q first", xs[0].ID)
	}
	if len(IndexRolloutEntrys(xs)) != 3 {
		t.Errorf("IndexRolloutEntrys() = %v", IndexRolloutEntrys(xs))
	}
	if err := ValidateRolloutEntrys(ActiveRolloutEntrys(xs)); err != nil {
		t.Errorf("ValidateRolloutEntrys() = %v", err)
	}
}

func TestRolloutEntryNormalizeClamps(t *testing.T) {
	x := SampleRolloutEntry("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutItemRoundTrip(t *testing.T) {
	x := SampleRolloutItem("rolloutitem-1", when)
	got, err := ParseRolloutItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutItem() = %v", err)
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

func TestRolloutItemParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutItem(f); err == nil {
		t.Error("ParseRolloutItem() accepted a record with no id")
	}
	f = SampleRolloutItem("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRolloutItem(f); err == nil {
		t.Error("ParseRolloutItem() accepted a non-numeric count")
	}
}

func TestRolloutItemCollections(t *testing.T) {
	xs := []RolloutItem{SampleRolloutItem("a", when), SampleRolloutItem("bb", when), SampleRolloutItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutItems(xs)); got != 2 {
		t.Errorf("ActiveRolloutItems() kept %d, want 2", got)
	}
	if got, want := SumRolloutItems(xs), 1+3; got != want {
		t.Errorf("SumRolloutItems() = %d, want %d", got, want)
	}
	SortRolloutItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutItems() put %q first", xs[0].ID)
	}
	if len(IndexRolloutItems(xs)) != 3 {
		t.Errorf("IndexRolloutItems() = %v", IndexRolloutItems(xs))
	}
	if err := ValidateRolloutItems(ActiveRolloutItems(xs)); err != nil {
		t.Errorf("ValidateRolloutItems() = %v", err)
	}
}

func TestRolloutItemNormalizeClamps(t *testing.T) {
	x := SampleRolloutItem("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutDraftRoundTrip(t *testing.T) {
	x := SampleRolloutDraft("rolloutdraft-1", when)
	got, err := ParseRolloutDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutDraft() = %v", err)
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

func TestRolloutDraftParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutDraft(f); err == nil {
		t.Error("ParseRolloutDraft() accepted a record with no id")
	}
	f = SampleRolloutDraft("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRolloutDraft(f); err == nil {
		t.Error("ParseRolloutDraft() accepted a non-numeric offset")
	}
}

func TestRolloutDraftCollections(t *testing.T) {
	xs := []RolloutDraft{SampleRolloutDraft("a", when), SampleRolloutDraft("bb", when), SampleRolloutDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutDrafts(xs)); got != 2 {
		t.Errorf("ActiveRolloutDrafts() kept %d, want 2", got)
	}
	if got, want := SumRolloutDrafts(xs), 1+3; got != want {
		t.Errorf("SumRolloutDrafts() = %d, want %d", got, want)
	}
	SortRolloutDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutDrafts() put %q first", xs[0].ID)
	}
	if len(IndexRolloutDrafts(xs)) != 3 {
		t.Errorf("IndexRolloutDrafts() = %v", IndexRolloutDrafts(xs))
	}
	if err := ValidateRolloutDrafts(ActiveRolloutDrafts(xs)); err != nil {
		t.Errorf("ValidateRolloutDrafts() = %v", err)
	}
}

func TestRolloutDraftNormalizeClamps(t *testing.T) {
	x := SampleRolloutDraft("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutRevisionRoundTrip(t *testing.T) {
	x := SampleRolloutRevision("rolloutrevision-1", when)
	got, err := ParseRolloutRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutRevision() = %v", err)
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

func TestRolloutRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutRevision(f); err == nil {
		t.Error("ParseRolloutRevision() accepted a record with no id")
	}
	f = SampleRolloutRevision("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRolloutRevision(f); err == nil {
		t.Error("ParseRolloutRevision() accepted a non-numeric priority")
	}
}

func TestRolloutRevisionCollections(t *testing.T) {
	xs := []RolloutRevision{SampleRolloutRevision("a", when), SampleRolloutRevision("bb", when), SampleRolloutRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutRevisions(xs)); got != 2 {
		t.Errorf("ActiveRolloutRevisions() kept %d, want 2", got)
	}
	if got, want := SumRolloutRevisions(xs), 1+3; got != want {
		t.Errorf("SumRolloutRevisions() = %d, want %d", got, want)
	}
	SortRolloutRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutRevisions() put %q first", xs[0].ID)
	}
	if len(IndexRolloutRevisions(xs)) != 3 {
		t.Errorf("IndexRolloutRevisions() = %v", IndexRolloutRevisions(xs))
	}
	if err := ValidateRolloutRevisions(ActiveRolloutRevisions(xs)); err != nil {
		t.Errorf("ValidateRolloutRevisions() = %v", err)
	}
}

func TestRolloutRevisionNormalizeClamps(t *testing.T) {
	x := SampleRolloutRevision("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutSnapshotRoundTrip(t *testing.T) {
	x := SampleRolloutSnapshot("rolloutsnapshot-1", when)
	got, err := ParseRolloutSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutSnapshot() = %v", err)
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

func TestRolloutSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutSnapshot(f); err == nil {
		t.Error("ParseRolloutSnapshot() accepted a record with no id")
	}
	f = SampleRolloutSnapshot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRolloutSnapshot(f); err == nil {
		t.Error("ParseRolloutSnapshot() accepted a non-numeric offset")
	}
}

func TestRolloutSnapshotCollections(t *testing.T) {
	xs := []RolloutSnapshot{SampleRolloutSnapshot("a", when), SampleRolloutSnapshot("bb", when), SampleRolloutSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutSnapshots(xs)); got != 2 {
		t.Errorf("ActiveRolloutSnapshots() kept %d, want 2", got)
	}
	if got, want := SumRolloutSnapshots(xs), 1+3; got != want {
		t.Errorf("SumRolloutSnapshots() = %d, want %d", got, want)
	}
	SortRolloutSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexRolloutSnapshots(xs)) != 3 {
		t.Errorf("IndexRolloutSnapshots() = %v", IndexRolloutSnapshots(xs))
	}
	if err := ValidateRolloutSnapshots(ActiveRolloutSnapshots(xs)); err != nil {
		t.Errorf("ValidateRolloutSnapshots() = %v", err)
	}
}

func TestRolloutSnapshotNormalizeClamps(t *testing.T) {
	x := SampleRolloutSnapshot("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutBatchRoundTrip(t *testing.T) {
	x := SampleRolloutBatch("rolloutbatch-1", when)
	got, err := ParseRolloutBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutBatch() = %v", err)
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

func TestRolloutBatchParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutBatch(f); err == nil {
		t.Error("ParseRolloutBatch() accepted a record with no id")
	}
	f = SampleRolloutBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRolloutBatch(f); err == nil {
		t.Error("ParseRolloutBatch() accepted a non-numeric capacity")
	}
}

func TestRolloutBatchCollections(t *testing.T) {
	xs := []RolloutBatch{SampleRolloutBatch("a", when), SampleRolloutBatch("bb", when), SampleRolloutBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutBatchs(xs)); got != 2 {
		t.Errorf("ActiveRolloutBatchs() kept %d, want 2", got)
	}
	if got, want := SumRolloutBatchs(xs), 1+3; got != want {
		t.Errorf("SumRolloutBatchs() = %d, want %d", got, want)
	}
	SortRolloutBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutBatchs() put %q first", xs[0].ID)
	}
	if len(IndexRolloutBatchs(xs)) != 3 {
		t.Errorf("IndexRolloutBatchs() = %v", IndexRolloutBatchs(xs))
	}
	if err := ValidateRolloutBatchs(ActiveRolloutBatchs(xs)); err != nil {
		t.Errorf("ValidateRolloutBatchs() = %v", err)
	}
}

func TestRolloutBatchNormalizeClamps(t *testing.T) {
	x := SampleRolloutBatch("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutRequestRoundTrip(t *testing.T) {
	x := SampleRolloutRequest("rolloutrequest-1", when)
	got, err := ParseRolloutRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutRequest() = %v", err)
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

func TestRolloutRequestParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutRequest(f); err == nil {
		t.Error("ParseRolloutRequest() accepted a record with no id")
	}
	f = SampleRolloutRequest("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRolloutRequest(f); err == nil {
		t.Error("ParseRolloutRequest() accepted a non-numeric retained")
	}
}

func TestRolloutRequestCollections(t *testing.T) {
	xs := []RolloutRequest{SampleRolloutRequest("a", when), SampleRolloutRequest("bb", when), SampleRolloutRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutRequests(xs)); got != 2 {
		t.Errorf("ActiveRolloutRequests() kept %d, want 2", got)
	}
	if got, want := SumRolloutRequests(xs), 1+3; got != want {
		t.Errorf("SumRolloutRequests() = %d, want %d", got, want)
	}
	SortRolloutRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutRequests() put %q first", xs[0].ID)
	}
	if len(IndexRolloutRequests(xs)) != 3 {
		t.Errorf("IndexRolloutRequests() = %v", IndexRolloutRequests(xs))
	}
	if err := ValidateRolloutRequests(ActiveRolloutRequests(xs)); err != nil {
		t.Errorf("ValidateRolloutRequests() = %v", err)
	}
}

func TestRolloutRequestNormalizeClamps(t *testing.T) {
	x := SampleRolloutRequest("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutSummaryRoundTrip(t *testing.T) {
	x := SampleRolloutSummary("rolloutsummary-1", when)
	got, err := ParseRolloutSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutSummary() = %v", err)
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

func TestRolloutSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutSummary(f); err == nil {
		t.Error("ParseRolloutSummary() accepted a record with no id")
	}
	f = SampleRolloutSummary("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRolloutSummary(f); err == nil {
		t.Error("ParseRolloutSummary() accepted a non-numeric retained")
	}
}

func TestRolloutSummaryCollections(t *testing.T) {
	xs := []RolloutSummary{SampleRolloutSummary("a", when), SampleRolloutSummary("bb", when), SampleRolloutSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutSummarys(xs)); got != 2 {
		t.Errorf("ActiveRolloutSummarys() kept %d, want 2", got)
	}
	if got, want := SumRolloutSummarys(xs), 1+3; got != want {
		t.Errorf("SumRolloutSummarys() = %d, want %d", got, want)
	}
	SortRolloutSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutSummarys() put %q first", xs[0].ID)
	}
	if len(IndexRolloutSummarys(xs)) != 3 {
		t.Errorf("IndexRolloutSummarys() = %v", IndexRolloutSummarys(xs))
	}
	if err := ValidateRolloutSummarys(ActiveRolloutSummarys(xs)); err != nil {
		t.Errorf("ValidateRolloutSummarys() = %v", err)
	}
}

func TestRolloutSummaryNormalizeClamps(t *testing.T) {
	x := SampleRolloutSummary("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutPolicyRoundTrip(t *testing.T) {
	x := SampleRolloutPolicy("rolloutpolicy-1", when)
	got, err := ParseRolloutPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutPolicy() = %v", err)
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

func TestRolloutPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutPolicy(f); err == nil {
		t.Error("ParseRolloutPolicy() accepted a record with no id")
	}
	f = SampleRolloutPolicy("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRolloutPolicy(f); err == nil {
		t.Error("ParseRolloutPolicy() accepted a non-numeric capacity")
	}
}

func TestRolloutPolicyCollections(t *testing.T) {
	xs := []RolloutPolicy{SampleRolloutPolicy("a", when), SampleRolloutPolicy("bb", when), SampleRolloutPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutPolicys(xs)); got != 2 {
		t.Errorf("ActiveRolloutPolicys() kept %d, want 2", got)
	}
	if got, want := SumRolloutPolicys(xs), 1+3; got != want {
		t.Errorf("SumRolloutPolicys() = %d, want %d", got, want)
	}
	SortRolloutPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutPolicys() put %q first", xs[0].ID)
	}
	if len(IndexRolloutPolicys(xs)) != 3 {
		t.Errorf("IndexRolloutPolicys() = %v", IndexRolloutPolicys(xs))
	}
	if err := ValidateRolloutPolicys(ActiveRolloutPolicys(xs)); err != nil {
		t.Errorf("ValidateRolloutPolicys() = %v", err)
	}
}

func TestRolloutPolicyNormalizeClamps(t *testing.T) {
	x := SampleRolloutPolicy("a", when)
	x.Label = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutLinkRoundTrip(t *testing.T) {
	x := SampleRolloutLink("rolloutlink-1", when)
	got, err := ParseRolloutLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutLink() = %v", err)
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

func TestRolloutLinkParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutLink(f); err == nil {
		t.Error("ParseRolloutLink() accepted a record with no id")
	}
	f = SampleRolloutLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRolloutLink(f); err == nil {
		t.Error("ParseRolloutLink() accepted a non-numeric retained")
	}
}

func TestRolloutLinkCollections(t *testing.T) {
	xs := []RolloutLink{SampleRolloutLink("a", when), SampleRolloutLink("bb", when), SampleRolloutLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutLinks(xs)); got != 2 {
		t.Errorf("ActiveRolloutLinks() kept %d, want 2", got)
	}
	if got, want := SumRolloutLinks(xs), 1+3; got != want {
		t.Errorf("SumRolloutLinks() = %d, want %d", got, want)
	}
	SortRolloutLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutLinks() put %q first", xs[0].ID)
	}
	if len(IndexRolloutLinks(xs)) != 3 {
		t.Errorf("IndexRolloutLinks() = %v", IndexRolloutLinks(xs))
	}
	if err := ValidateRolloutLinks(ActiveRolloutLinks(xs)); err != nil {
		t.Errorf("ValidateRolloutLinks() = %v", err)
	}
}

func TestRolloutLinkNormalizeClamps(t *testing.T) {
	x := SampleRolloutLink("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutNoteRoundTrip(t *testing.T) {
	x := SampleRolloutNote("rolloutnote-1", when)
	got, err := ParseRolloutNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutNote() = %v", err)
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

func TestRolloutNoteParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutNote(f); err == nil {
		t.Error("ParseRolloutNote() accepted a record with no id")
	}
	f = SampleRolloutNote("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseRolloutNote(f); err == nil {
		t.Error("ParseRolloutNote() accepted a non-numeric capacity")
	}
}

func TestRolloutNoteCollections(t *testing.T) {
	xs := []RolloutNote{SampleRolloutNote("a", when), SampleRolloutNote("bb", when), SampleRolloutNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutNotes(xs)); got != 2 {
		t.Errorf("ActiveRolloutNotes() kept %d, want 2", got)
	}
	if got, want := SumRolloutNotes(xs), 1+3; got != want {
		t.Errorf("SumRolloutNotes() = %d, want %d", got, want)
	}
	SortRolloutNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutNotes() put %q first", xs[0].ID)
	}
	if len(IndexRolloutNotes(xs)) != 3 {
		t.Errorf("IndexRolloutNotes() = %v", IndexRolloutNotes(xs))
	}
	if err := ValidateRolloutNotes(ActiveRolloutNotes(xs)); err != nil {
		t.Errorf("ValidateRolloutNotes() = %v", err)
	}
}

func TestRolloutNoteNormalizeClamps(t *testing.T) {
	x := SampleRolloutNote("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutWindowRoundTrip(t *testing.T) {
	x := SampleRolloutWindow("rolloutwindow-1", when)
	got, err := ParseRolloutWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutWindow() = %v", err)
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

func TestRolloutWindowParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutWindow(f); err == nil {
		t.Error("ParseRolloutWindow() accepted a record with no id")
	}
	f = SampleRolloutWindow("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseRolloutWindow(f); err == nil {
		t.Error("ParseRolloutWindow() accepted a non-numeric count")
	}
}

func TestRolloutWindowCollections(t *testing.T) {
	xs := []RolloutWindow{SampleRolloutWindow("a", when), SampleRolloutWindow("bb", when), SampleRolloutWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutWindows(xs)); got != 2 {
		t.Errorf("ActiveRolloutWindows() kept %d, want 2", got)
	}
	if got, want := SumRolloutWindows(xs), 1+3; got != want {
		t.Errorf("SumRolloutWindows() = %d, want %d", got, want)
	}
	SortRolloutWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutWindows() put %q first", xs[0].ID)
	}
	if len(IndexRolloutWindows(xs)) != 3 {
		t.Errorf("IndexRolloutWindows() = %v", IndexRolloutWindows(xs))
	}
	if err := ValidateRolloutWindows(ActiveRolloutWindows(xs)); err != nil {
		t.Errorf("ValidateRolloutWindows() = %v", err)
	}
}

func TestRolloutWindowNormalizeClamps(t *testing.T) {
	x := SampleRolloutWindow("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutQuotaRoundTrip(t *testing.T) {
	x := SampleRolloutQuota("rolloutquota-1", when)
	got, err := ParseRolloutQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutQuota() = %v", err)
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

func TestRolloutQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutQuota(f); err == nil {
		t.Error("ParseRolloutQuota() accepted a record with no id")
	}
	f = SampleRolloutQuota("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRolloutQuota(f); err == nil {
		t.Error("ParseRolloutQuota() accepted a non-numeric retained")
	}
}

func TestRolloutQuotaCollections(t *testing.T) {
	xs := []RolloutQuota{SampleRolloutQuota("a", when), SampleRolloutQuota("bb", when), SampleRolloutQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutQuotas(xs)); got != 2 {
		t.Errorf("ActiveRolloutQuotas() kept %d, want 2", got)
	}
	if got, want := SumRolloutQuotas(xs), 1+3; got != want {
		t.Errorf("SumRolloutQuotas() = %d, want %d", got, want)
	}
	SortRolloutQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutQuotas() put %q first", xs[0].ID)
	}
	if len(IndexRolloutQuotas(xs)) != 3 {
		t.Errorf("IndexRolloutQuotas() = %v", IndexRolloutQuotas(xs))
	}
	if err := ValidateRolloutQuotas(ActiveRolloutQuotas(xs)); err != nil {
		t.Errorf("ValidateRolloutQuotas() = %v", err)
	}
}

func TestRolloutQuotaNormalizeClamps(t *testing.T) {
	x := SampleRolloutQuota("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutBucketRoundTrip(t *testing.T) {
	x := SampleRolloutBucket("rolloutbucket-1", when)
	got, err := ParseRolloutBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutBucket() = %v", err)
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

func TestRolloutBucketParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutBucket(f); err == nil {
		t.Error("ParseRolloutBucket() accepted a record with no id")
	}
	f = SampleRolloutBucket("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRolloutBucket(f); err == nil {
		t.Error("ParseRolloutBucket() accepted a non-numeric weight")
	}
}

func TestRolloutBucketCollections(t *testing.T) {
	xs := []RolloutBucket{SampleRolloutBucket("a", when), SampleRolloutBucket("bb", when), SampleRolloutBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutBuckets(xs)); got != 2 {
		t.Errorf("ActiveRolloutBuckets() kept %d, want 2", got)
	}
	if got, want := SumRolloutBuckets(xs), 1+3; got != want {
		t.Errorf("SumRolloutBuckets() = %d, want %d", got, want)
	}
	SortRolloutBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutBuckets() put %q first", xs[0].ID)
	}
	if len(IndexRolloutBuckets(xs)) != 3 {
		t.Errorf("IndexRolloutBuckets() = %v", IndexRolloutBuckets(xs))
	}
	if err := ValidateRolloutBuckets(ActiveRolloutBuckets(xs)); err != nil {
		t.Errorf("ValidateRolloutBuckets() = %v", err)
	}
}

func TestRolloutBucketNormalizeClamps(t *testing.T) {
	x := SampleRolloutBucket("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutLaneRoundTrip(t *testing.T) {
	x := SampleRolloutLane("rolloutlane-1", when)
	got, err := ParseRolloutLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutLane() = %v", err)
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

func TestRolloutLaneParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutLane(f); err == nil {
		t.Error("ParseRolloutLane() accepted a record with no id")
	}
	f = SampleRolloutLane("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseRolloutLane(f); err == nil {
		t.Error("ParseRolloutLane() accepted a non-numeric weight")
	}
}

func TestRolloutLaneCollections(t *testing.T) {
	xs := []RolloutLane{SampleRolloutLane("a", when), SampleRolloutLane("bb", when), SampleRolloutLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutLanes(xs)); got != 2 {
		t.Errorf("ActiveRolloutLanes() kept %d, want 2", got)
	}
	if got, want := SumRolloutLanes(xs), 1+3; got != want {
		t.Errorf("SumRolloutLanes() = %d, want %d", got, want)
	}
	SortRolloutLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutLanes() put %q first", xs[0].ID)
	}
	if len(IndexRolloutLanes(xs)) != 3 {
		t.Errorf("IndexRolloutLanes() = %v", IndexRolloutLanes(xs))
	}
	if err := ValidateRolloutLanes(ActiveRolloutLanes(xs)); err != nil {
		t.Errorf("ValidateRolloutLanes() = %v", err)
	}
}

func TestRolloutLaneNormalizeClamps(t *testing.T) {
	x := SampleRolloutLane("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutSlotRoundTrip(t *testing.T) {
	x := SampleRolloutSlot("rolloutslot-1", when)
	got, err := ParseRolloutSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutSlot() = %v", err)
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

func TestRolloutSlotParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutSlot(f); err == nil {
		t.Error("ParseRolloutSlot() accepted a record with no id")
	}
	f = SampleRolloutSlot("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseRolloutSlot(f); err == nil {
		t.Error("ParseRolloutSlot() accepted a non-numeric version")
	}
}

func TestRolloutSlotCollections(t *testing.T) {
	xs := []RolloutSlot{SampleRolloutSlot("a", when), SampleRolloutSlot("bb", when), SampleRolloutSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutSlots(xs)); got != 2 {
		t.Errorf("ActiveRolloutSlots() kept %d, want 2", got)
	}
	if got, want := SumRolloutSlots(xs), 1+3; got != want {
		t.Errorf("SumRolloutSlots() = %d, want %d", got, want)
	}
	SortRolloutSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutSlots() put %q first", xs[0].ID)
	}
	if len(IndexRolloutSlots(xs)) != 3 {
		t.Errorf("IndexRolloutSlots() = %v", IndexRolloutSlots(xs))
	}
	if err := ValidateRolloutSlots(ActiveRolloutSlots(xs)); err != nil {
		t.Errorf("ValidateRolloutSlots() = %v", err)
	}
}

func TestRolloutSlotNormalizeClamps(t *testing.T) {
	x := SampleRolloutSlot("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutTagRoundTrip(t *testing.T) {
	x := SampleRolloutTag("rollouttag-1", when)
	got, err := ParseRolloutTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutTag() = %v", err)
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

func TestRolloutTagParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutTag(f); err == nil {
		t.Error("ParseRolloutTag() accepted a record with no id")
	}
	f = SampleRolloutTag("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRolloutTag(f); err == nil {
		t.Error("ParseRolloutTag() accepted a non-numeric offset")
	}
}

func TestRolloutTagCollections(t *testing.T) {
	xs := []RolloutTag{SampleRolloutTag("a", when), SampleRolloutTag("bb", when), SampleRolloutTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutTags(xs)); got != 2 {
		t.Errorf("ActiveRolloutTags() kept %d, want 2", got)
	}
	if got, want := SumRolloutTags(xs), 1+3; got != want {
		t.Errorf("SumRolloutTags() = %d, want %d", got, want)
	}
	SortRolloutTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutTags() put %q first", xs[0].ID)
	}
	if len(IndexRolloutTags(xs)) != 3 {
		t.Errorf("IndexRolloutTags() = %v", IndexRolloutTags(xs))
	}
	if err := ValidateRolloutTags(ActiveRolloutTags(xs)); err != nil {
		t.Errorf("ValidateRolloutTags() = %v", err)
	}
}

func TestRolloutTagNormalizeClamps(t *testing.T) {
	x := SampleRolloutTag("a", when)
	x.Source = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutRuleRoundTrip(t *testing.T) {
	x := SampleRolloutRule("rolloutrule-1", when)
	got, err := ParseRolloutRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutRule() = %v", err)
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

func TestRolloutRuleParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutRule(f); err == nil {
		t.Error("ParseRolloutRule() accepted a record with no id")
	}
	f = SampleRolloutRule("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseRolloutRule(f); err == nil {
		t.Error("ParseRolloutRule() accepted a non-numeric priority")
	}
}

func TestRolloutRuleCollections(t *testing.T) {
	xs := []RolloutRule{SampleRolloutRule("a", when), SampleRolloutRule("bb", when), SampleRolloutRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutRules(xs)); got != 2 {
		t.Errorf("ActiveRolloutRules() kept %d, want 2", got)
	}
	if got, want := SumRolloutRules(xs), 1+3; got != want {
		t.Errorf("SumRolloutRules() = %d, want %d", got, want)
	}
	SortRolloutRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutRules() put %q first", xs[0].ID)
	}
	if len(IndexRolloutRules(xs)) != 3 {
		t.Errorf("IndexRolloutRules() = %v", IndexRolloutRules(xs))
	}
	if err := ValidateRolloutRules(ActiveRolloutRules(xs)); err != nil {
		t.Errorf("ValidateRolloutRules() = %v", err)
	}
}

func TestRolloutRuleNormalizeClamps(t *testing.T) {
	x := SampleRolloutRule("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutHoldRoundTrip(t *testing.T) {
	x := SampleRolloutHold("rollouthold-1", when)
	got, err := ParseRolloutHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutHold() = %v", err)
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

func TestRolloutHoldParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutHold(f); err == nil {
		t.Error("ParseRolloutHold() accepted a record with no id")
	}
	f = SampleRolloutHold("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseRolloutHold(f); err == nil {
		t.Error("ParseRolloutHold() accepted a non-numeric attempts")
	}
}

func TestRolloutHoldCollections(t *testing.T) {
	xs := []RolloutHold{SampleRolloutHold("a", when), SampleRolloutHold("bb", when), SampleRolloutHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutHolds(xs)); got != 2 {
		t.Errorf("ActiveRolloutHolds() kept %d, want 2", got)
	}
	if got, want := SumRolloutHolds(xs), 1+3; got != want {
		t.Errorf("SumRolloutHolds() = %d, want %d", got, want)
	}
	SortRolloutHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutHolds() put %q first", xs[0].ID)
	}
	if len(IndexRolloutHolds(xs)) != 3 {
		t.Errorf("IndexRolloutHolds() = %v", IndexRolloutHolds(xs))
	}
	if err := ValidateRolloutHolds(ActiveRolloutHolds(xs)); err != nil {
		t.Errorf("ValidateRolloutHolds() = %v", err)
	}
}

func TestRolloutHoldNormalizeClamps(t *testing.T) {
	x := SampleRolloutHold("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutClaimRoundTrip(t *testing.T) {
	x := SampleRolloutClaim("rolloutclaim-1", when)
	got, err := ParseRolloutClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutClaim() = %v", err)
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

func TestRolloutClaimParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutClaim(f); err == nil {
		t.Error("ParseRolloutClaim() accepted a record with no id")
	}
	f = SampleRolloutClaim("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseRolloutClaim(f); err == nil {
		t.Error("ParseRolloutClaim() accepted a non-numeric offset")
	}
}

func TestRolloutClaimCollections(t *testing.T) {
	xs := []RolloutClaim{SampleRolloutClaim("a", when), SampleRolloutClaim("bb", when), SampleRolloutClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutClaims(xs)); got != 2 {
		t.Errorf("ActiveRolloutClaims() kept %d, want 2", got)
	}
	if got, want := SumRolloutClaims(xs), 1+3; got != want {
		t.Errorf("SumRolloutClaims() = %d, want %d", got, want)
	}
	SortRolloutClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutClaims() put %q first", xs[0].ID)
	}
	if len(IndexRolloutClaims(xs)) != 3 {
		t.Errorf("IndexRolloutClaims() = %v", IndexRolloutClaims(xs))
	}
	if err := ValidateRolloutClaims(ActiveRolloutClaims(xs)); err != nil {
		t.Errorf("ValidateRolloutClaims() = %v", err)
	}
}

func TestRolloutClaimNormalizeClamps(t *testing.T) {
	x := SampleRolloutClaim("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestRolloutCursorRoundTrip(t *testing.T) {
	x := SampleRolloutCursor("rolloutcursor-1", when)
	got, err := ParseRolloutCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseRolloutCursor() = %v", err)
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

func TestRolloutCursorParseRejectsBadInput(t *testing.T) {
	f := SampleRolloutCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseRolloutCursor(f); err == nil {
		t.Error("ParseRolloutCursor() accepted a record with no id")
	}
	f = SampleRolloutCursor("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseRolloutCursor(f); err == nil {
		t.Error("ParseRolloutCursor() accepted a non-numeric retained")
	}
}

func TestRolloutCursorCollections(t *testing.T) {
	xs := []RolloutCursor{SampleRolloutCursor("a", when), SampleRolloutCursor("bb", when), SampleRolloutCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveRolloutCursors(xs)); got != 2 {
		t.Errorf("ActiveRolloutCursors() kept %d, want 2", got)
	}
	if got, want := SumRolloutCursors(xs), 1+3; got != want {
		t.Errorf("SumRolloutCursors() = %d, want %d", got, want)
	}
	SortRolloutCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortRolloutCursors() put %q first", xs[0].ID)
	}
	if len(IndexRolloutCursors(xs)) != 3 {
		t.Errorf("IndexRolloutCursors() = %v", IndexRolloutCursors(xs))
	}
	if err := ValidateRolloutCursors(ActiveRolloutCursors(xs)); err != nil {
		t.Errorf("ValidateRolloutCursors() = %v", err)
	}
}

func TestRolloutCursorNormalizeClamps(t *testing.T) {
	x := SampleRolloutCursor("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
