package segment

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
		if !strings.HasPrefix(Prefixes[i], "segment/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleSegmentRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestSegmentRecordRoundTrip(t *testing.T) {
	x := SampleSegmentRecord("segmentrecord-1", when)
	got, err := ParseSegmentRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentRecord() = %v", err)
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

func TestSegmentRecordParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentRecord(f); err == nil {
		t.Error("ParseSegmentRecord() accepted a record with no id")
	}
	f = SampleSegmentRecord("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSegmentRecord(f); err == nil {
		t.Error("ParseSegmentRecord() accepted a non-numeric count")
	}
}

func TestSegmentRecordCollections(t *testing.T) {
	xs := []SegmentRecord{SampleSegmentRecord("a", when), SampleSegmentRecord("bb", when), SampleSegmentRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentRecords(xs)); got != 2 {
		t.Errorf("ActiveSegmentRecords() kept %d, want 2", got)
	}
	if got, want := SumSegmentRecords(xs), 1+3; got != want {
		t.Errorf("SumSegmentRecords() = %d, want %d", got, want)
	}
	SortSegmentRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentRecords() put %q first", xs[0].ID)
	}
	if len(IndexSegmentRecords(xs)) != 3 {
		t.Errorf("IndexSegmentRecords() = %v", IndexSegmentRecords(xs))
	}
	if err := ValidateSegmentRecords(ActiveSegmentRecords(xs)); err != nil {
		t.Errorf("ValidateSegmentRecords() = %v", err)
	}
}

func TestSegmentRecordNormalizeClamps(t *testing.T) {
	x := SampleSegmentRecord("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentEntryRoundTrip(t *testing.T) {
	x := SampleSegmentEntry("segmententry-1", when)
	got, err := ParseSegmentEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentEntry() = %v", err)
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

func TestSegmentEntryParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentEntry(f); err == nil {
		t.Error("ParseSegmentEntry() accepted a record with no id")
	}
	f = SampleSegmentEntry("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseSegmentEntry(f); err == nil {
		t.Error("ParseSegmentEntry() accepted a non-numeric attempts")
	}
}

func TestSegmentEntryCollections(t *testing.T) {
	xs := []SegmentEntry{SampleSegmentEntry("a", when), SampleSegmentEntry("bb", when), SampleSegmentEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentEntrys(xs)); got != 2 {
		t.Errorf("ActiveSegmentEntrys() kept %d, want 2", got)
	}
	if got, want := SumSegmentEntrys(xs), 1+3; got != want {
		t.Errorf("SumSegmentEntrys() = %d, want %d", got, want)
	}
	SortSegmentEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentEntrys() put %q first", xs[0].ID)
	}
	if len(IndexSegmentEntrys(xs)) != 3 {
		t.Errorf("IndexSegmentEntrys() = %v", IndexSegmentEntrys(xs))
	}
	if err := ValidateSegmentEntrys(ActiveSegmentEntrys(xs)); err != nil {
		t.Errorf("ValidateSegmentEntrys() = %v", err)
	}
}

func TestSegmentEntryNormalizeClamps(t *testing.T) {
	x := SampleSegmentEntry("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentItemRoundTrip(t *testing.T) {
	x := SampleSegmentItem("segmentitem-1", when)
	got, err := ParseSegmentItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentItem() = %v", err)
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

func TestSegmentItemParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentItem(f); err == nil {
		t.Error("ParseSegmentItem() accepted a record with no id")
	}
	f = SampleSegmentItem("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSegmentItem(f); err == nil {
		t.Error("ParseSegmentItem() accepted a non-numeric priority")
	}
}

func TestSegmentItemCollections(t *testing.T) {
	xs := []SegmentItem{SampleSegmentItem("a", when), SampleSegmentItem("bb", when), SampleSegmentItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentItems(xs)); got != 2 {
		t.Errorf("ActiveSegmentItems() kept %d, want 2", got)
	}
	if got, want := SumSegmentItems(xs), 1+3; got != want {
		t.Errorf("SumSegmentItems() = %d, want %d", got, want)
	}
	SortSegmentItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentItems() put %q first", xs[0].ID)
	}
	if len(IndexSegmentItems(xs)) != 3 {
		t.Errorf("IndexSegmentItems() = %v", IndexSegmentItems(xs))
	}
	if err := ValidateSegmentItems(ActiveSegmentItems(xs)); err != nil {
		t.Errorf("ValidateSegmentItems() = %v", err)
	}
}

func TestSegmentItemNormalizeClamps(t *testing.T) {
	x := SampleSegmentItem("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentDraftRoundTrip(t *testing.T) {
	x := SampleSegmentDraft("segmentdraft-1", when)
	got, err := ParseSegmentDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentDraft() = %v", err)
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

func TestSegmentDraftParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentDraft(f); err == nil {
		t.Error("ParseSegmentDraft() accepted a record with no id")
	}
	f = SampleSegmentDraft("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSegmentDraft(f); err == nil {
		t.Error("ParseSegmentDraft() accepted a non-numeric count")
	}
}

func TestSegmentDraftCollections(t *testing.T) {
	xs := []SegmentDraft{SampleSegmentDraft("a", when), SampleSegmentDraft("bb", when), SampleSegmentDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentDrafts(xs)); got != 2 {
		t.Errorf("ActiveSegmentDrafts() kept %d, want 2", got)
	}
	if got, want := SumSegmentDrafts(xs), 1+3; got != want {
		t.Errorf("SumSegmentDrafts() = %d, want %d", got, want)
	}
	SortSegmentDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentDrafts() put %q first", xs[0].ID)
	}
	if len(IndexSegmentDrafts(xs)) != 3 {
		t.Errorf("IndexSegmentDrafts() = %v", IndexSegmentDrafts(xs))
	}
	if err := ValidateSegmentDrafts(ActiveSegmentDrafts(xs)); err != nil {
		t.Errorf("ValidateSegmentDrafts() = %v", err)
	}
}

func TestSegmentDraftNormalizeClamps(t *testing.T) {
	x := SampleSegmentDraft("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentRevisionRoundTrip(t *testing.T) {
	x := SampleSegmentRevision("segmentrevision-1", when)
	got, err := ParseSegmentRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentRevision() = %v", err)
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

func TestSegmentRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentRevision(f); err == nil {
		t.Error("ParseSegmentRevision() accepted a record with no id")
	}
	f = SampleSegmentRevision("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSegmentRevision(f); err == nil {
		t.Error("ParseSegmentRevision() accepted a non-numeric version")
	}
}

func TestSegmentRevisionCollections(t *testing.T) {
	xs := []SegmentRevision{SampleSegmentRevision("a", when), SampleSegmentRevision("bb", when), SampleSegmentRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentRevisions(xs)); got != 2 {
		t.Errorf("ActiveSegmentRevisions() kept %d, want 2", got)
	}
	if got, want := SumSegmentRevisions(xs), 1+3; got != want {
		t.Errorf("SumSegmentRevisions() = %d, want %d", got, want)
	}
	SortSegmentRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentRevisions() put %q first", xs[0].ID)
	}
	if len(IndexSegmentRevisions(xs)) != 3 {
		t.Errorf("IndexSegmentRevisions() = %v", IndexSegmentRevisions(xs))
	}
	if err := ValidateSegmentRevisions(ActiveSegmentRevisions(xs)); err != nil {
		t.Errorf("ValidateSegmentRevisions() = %v", err)
	}
}

func TestSegmentRevisionNormalizeClamps(t *testing.T) {
	x := SampleSegmentRevision("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentSnapshotRoundTrip(t *testing.T) {
	x := SampleSegmentSnapshot("segmentsnapshot-1", when)
	got, err := ParseSegmentSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentSnapshot() = %v", err)
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

func TestSegmentSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentSnapshot(f); err == nil {
		t.Error("ParseSegmentSnapshot() accepted a record with no id")
	}
	f = SampleSegmentSnapshot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSegmentSnapshot(f); err == nil {
		t.Error("ParseSegmentSnapshot() accepted a non-numeric count")
	}
}

func TestSegmentSnapshotCollections(t *testing.T) {
	xs := []SegmentSnapshot{SampleSegmentSnapshot("a", when), SampleSegmentSnapshot("bb", when), SampleSegmentSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentSnapshots(xs)); got != 2 {
		t.Errorf("ActiveSegmentSnapshots() kept %d, want 2", got)
	}
	if got, want := SumSegmentSnapshots(xs), 1+3; got != want {
		t.Errorf("SumSegmentSnapshots() = %d, want %d", got, want)
	}
	SortSegmentSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexSegmentSnapshots(xs)) != 3 {
		t.Errorf("IndexSegmentSnapshots() = %v", IndexSegmentSnapshots(xs))
	}
	if err := ValidateSegmentSnapshots(ActiveSegmentSnapshots(xs)); err != nil {
		t.Errorf("ValidateSegmentSnapshots() = %v", err)
	}
}

func TestSegmentSnapshotNormalizeClamps(t *testing.T) {
	x := SampleSegmentSnapshot("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentBatchRoundTrip(t *testing.T) {
	x := SampleSegmentBatch("segmentbatch-1", when)
	got, err := ParseSegmentBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentBatch() = %v", err)
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

func TestSegmentBatchParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentBatch(f); err == nil {
		t.Error("ParseSegmentBatch() accepted a record with no id")
	}
	f = SampleSegmentBatch("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSegmentBatch(f); err == nil {
		t.Error("ParseSegmentBatch() accepted a non-numeric count")
	}
}

func TestSegmentBatchCollections(t *testing.T) {
	xs := []SegmentBatch{SampleSegmentBatch("a", when), SampleSegmentBatch("bb", when), SampleSegmentBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentBatchs(xs)); got != 2 {
		t.Errorf("ActiveSegmentBatchs() kept %d, want 2", got)
	}
	if got, want := SumSegmentBatchs(xs), 1+3; got != want {
		t.Errorf("SumSegmentBatchs() = %d, want %d", got, want)
	}
	SortSegmentBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentBatchs() put %q first", xs[0].ID)
	}
	if len(IndexSegmentBatchs(xs)) != 3 {
		t.Errorf("IndexSegmentBatchs() = %v", IndexSegmentBatchs(xs))
	}
	if err := ValidateSegmentBatchs(ActiveSegmentBatchs(xs)); err != nil {
		t.Errorf("ValidateSegmentBatchs() = %v", err)
	}
}

func TestSegmentBatchNormalizeClamps(t *testing.T) {
	x := SampleSegmentBatch("a", when)
	x.Source = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentRequestRoundTrip(t *testing.T) {
	x := SampleSegmentRequest("segmentrequest-1", when)
	got, err := ParseSegmentRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentRequest() = %v", err)
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

func TestSegmentRequestParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentRequest(f); err == nil {
		t.Error("ParseSegmentRequest() accepted a record with no id")
	}
	f = SampleSegmentRequest("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSegmentRequest(f); err == nil {
		t.Error("ParseSegmentRequest() accepted a non-numeric count")
	}
}

func TestSegmentRequestCollections(t *testing.T) {
	xs := []SegmentRequest{SampleSegmentRequest("a", when), SampleSegmentRequest("bb", when), SampleSegmentRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentRequests(xs)); got != 2 {
		t.Errorf("ActiveSegmentRequests() kept %d, want 2", got)
	}
	if got, want := SumSegmentRequests(xs), 1+3; got != want {
		t.Errorf("SumSegmentRequests() = %d, want %d", got, want)
	}
	SortSegmentRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentRequests() put %q first", xs[0].ID)
	}
	if len(IndexSegmentRequests(xs)) != 3 {
		t.Errorf("IndexSegmentRequests() = %v", IndexSegmentRequests(xs))
	}
	if err := ValidateSegmentRequests(ActiveSegmentRequests(xs)); err != nil {
		t.Errorf("ValidateSegmentRequests() = %v", err)
	}
}

func TestSegmentRequestNormalizeClamps(t *testing.T) {
	x := SampleSegmentRequest("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentSummaryRoundTrip(t *testing.T) {
	x := SampleSegmentSummary("segmentsummary-1", when)
	got, err := ParseSegmentSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentSummary() = %v", err)
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

func TestSegmentSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentSummary(f); err == nil {
		t.Error("ParseSegmentSummary() accepted a record with no id")
	}
	f = SampleSegmentSummary("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSegmentSummary(f); err == nil {
		t.Error("ParseSegmentSummary() accepted a non-numeric version")
	}
}

func TestSegmentSummaryCollections(t *testing.T) {
	xs := []SegmentSummary{SampleSegmentSummary("a", when), SampleSegmentSummary("bb", when), SampleSegmentSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentSummarys(xs)); got != 2 {
		t.Errorf("ActiveSegmentSummarys() kept %d, want 2", got)
	}
	if got, want := SumSegmentSummarys(xs), 1+3; got != want {
		t.Errorf("SumSegmentSummarys() = %d, want %d", got, want)
	}
	SortSegmentSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentSummarys() put %q first", xs[0].ID)
	}
	if len(IndexSegmentSummarys(xs)) != 3 {
		t.Errorf("IndexSegmentSummarys() = %v", IndexSegmentSummarys(xs))
	}
	if err := ValidateSegmentSummarys(ActiveSegmentSummarys(xs)); err != nil {
		t.Errorf("ValidateSegmentSummarys() = %v", err)
	}
}

func TestSegmentSummaryNormalizeClamps(t *testing.T) {
	x := SampleSegmentSummary("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentPolicyRoundTrip(t *testing.T) {
	x := SampleSegmentPolicy("segmentpolicy-1", when)
	got, err := ParseSegmentPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentPolicy() = %v", err)
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

func TestSegmentPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentPolicy(f); err == nil {
		t.Error("ParseSegmentPolicy() accepted a record with no id")
	}
	f = SampleSegmentPolicy("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSegmentPolicy(f); err == nil {
		t.Error("ParseSegmentPolicy() accepted a non-numeric priority")
	}
}

func TestSegmentPolicyCollections(t *testing.T) {
	xs := []SegmentPolicy{SampleSegmentPolicy("a", when), SampleSegmentPolicy("bb", when), SampleSegmentPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentPolicys(xs)); got != 2 {
		t.Errorf("ActiveSegmentPolicys() kept %d, want 2", got)
	}
	if got, want := SumSegmentPolicys(xs), 1+3; got != want {
		t.Errorf("SumSegmentPolicys() = %d, want %d", got, want)
	}
	SortSegmentPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentPolicys() put %q first", xs[0].ID)
	}
	if len(IndexSegmentPolicys(xs)) != 3 {
		t.Errorf("IndexSegmentPolicys() = %v", IndexSegmentPolicys(xs))
	}
	if err := ValidateSegmentPolicys(ActiveSegmentPolicys(xs)); err != nil {
		t.Errorf("ValidateSegmentPolicys() = %v", err)
	}
}

func TestSegmentPolicyNormalizeClamps(t *testing.T) {
	x := SampleSegmentPolicy("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentLinkRoundTrip(t *testing.T) {
	x := SampleSegmentLink("segmentlink-1", when)
	got, err := ParseSegmentLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentLink() = %v", err)
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

func TestSegmentLinkParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentLink(f); err == nil {
		t.Error("ParseSegmentLink() accepted a record with no id")
	}
	f = SampleSegmentLink("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSegmentLink(f); err == nil {
		t.Error("ParseSegmentLink() accepted a non-numeric priority")
	}
}

func TestSegmentLinkCollections(t *testing.T) {
	xs := []SegmentLink{SampleSegmentLink("a", when), SampleSegmentLink("bb", when), SampleSegmentLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentLinks(xs)); got != 2 {
		t.Errorf("ActiveSegmentLinks() kept %d, want 2", got)
	}
	if got, want := SumSegmentLinks(xs), 1+3; got != want {
		t.Errorf("SumSegmentLinks() = %d, want %d", got, want)
	}
	SortSegmentLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentLinks() put %q first", xs[0].ID)
	}
	if len(IndexSegmentLinks(xs)) != 3 {
		t.Errorf("IndexSegmentLinks() = %v", IndexSegmentLinks(xs))
	}
	if err := ValidateSegmentLinks(ActiveSegmentLinks(xs)); err != nil {
		t.Errorf("ValidateSegmentLinks() = %v", err)
	}
}

func TestSegmentLinkNormalizeClamps(t *testing.T) {
	x := SampleSegmentLink("a", when)
	x.Region = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentNoteRoundTrip(t *testing.T) {
	x := SampleSegmentNote("segmentnote-1", when)
	got, err := ParseSegmentNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentNote() = %v", err)
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

func TestSegmentNoteParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentNote(f); err == nil {
		t.Error("ParseSegmentNote() accepted a record with no id")
	}
	f = SampleSegmentNote("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSegmentNote(f); err == nil {
		t.Error("ParseSegmentNote() accepted a non-numeric priority")
	}
}

func TestSegmentNoteCollections(t *testing.T) {
	xs := []SegmentNote{SampleSegmentNote("a", when), SampleSegmentNote("bb", when), SampleSegmentNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentNotes(xs)); got != 2 {
		t.Errorf("ActiveSegmentNotes() kept %d, want 2", got)
	}
	if got, want := SumSegmentNotes(xs), 1+3; got != want {
		t.Errorf("SumSegmentNotes() = %d, want %d", got, want)
	}
	SortSegmentNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentNotes() put %q first", xs[0].ID)
	}
	if len(IndexSegmentNotes(xs)) != 3 {
		t.Errorf("IndexSegmentNotes() = %v", IndexSegmentNotes(xs))
	}
	if err := ValidateSegmentNotes(ActiveSegmentNotes(xs)); err != nil {
		t.Errorf("ValidateSegmentNotes() = %v", err)
	}
}

func TestSegmentNoteNormalizeClamps(t *testing.T) {
	x := SampleSegmentNote("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentWindowRoundTrip(t *testing.T) {
	x := SampleSegmentWindow("segmentwindow-1", when)
	got, err := ParseSegmentWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentWindow() = %v", err)
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

func TestSegmentWindowParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentWindow(f); err == nil {
		t.Error("ParseSegmentWindow() accepted a record with no id")
	}
	f = SampleSegmentWindow("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSegmentWindow(f); err == nil {
		t.Error("ParseSegmentWindow() accepted a non-numeric offset")
	}
}

func TestSegmentWindowCollections(t *testing.T) {
	xs := []SegmentWindow{SampleSegmentWindow("a", when), SampleSegmentWindow("bb", when), SampleSegmentWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentWindows(xs)); got != 2 {
		t.Errorf("ActiveSegmentWindows() kept %d, want 2", got)
	}
	if got, want := SumSegmentWindows(xs), 1+3; got != want {
		t.Errorf("SumSegmentWindows() = %d, want %d", got, want)
	}
	SortSegmentWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentWindows() put %q first", xs[0].ID)
	}
	if len(IndexSegmentWindows(xs)) != 3 {
		t.Errorf("IndexSegmentWindows() = %v", IndexSegmentWindows(xs))
	}
	if err := ValidateSegmentWindows(ActiveSegmentWindows(xs)); err != nil {
		t.Errorf("ValidateSegmentWindows() = %v", err)
	}
}

func TestSegmentWindowNormalizeClamps(t *testing.T) {
	x := SampleSegmentWindow("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentQuotaRoundTrip(t *testing.T) {
	x := SampleSegmentQuota("segmentquota-1", when)
	got, err := ParseSegmentQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentQuota() = %v", err)
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

func TestSegmentQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentQuota(f); err == nil {
		t.Error("ParseSegmentQuota() accepted a record with no id")
	}
	f = SampleSegmentQuota("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSegmentQuota(f); err == nil {
		t.Error("ParseSegmentQuota() accepted a non-numeric capacity")
	}
}

func TestSegmentQuotaCollections(t *testing.T) {
	xs := []SegmentQuota{SampleSegmentQuota("a", when), SampleSegmentQuota("bb", when), SampleSegmentQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentQuotas(xs)); got != 2 {
		t.Errorf("ActiveSegmentQuotas() kept %d, want 2", got)
	}
	if got, want := SumSegmentQuotas(xs), 1+3; got != want {
		t.Errorf("SumSegmentQuotas() = %d, want %d", got, want)
	}
	SortSegmentQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentQuotas() put %q first", xs[0].ID)
	}
	if len(IndexSegmentQuotas(xs)) != 3 {
		t.Errorf("IndexSegmentQuotas() = %v", IndexSegmentQuotas(xs))
	}
	if err := ValidateSegmentQuotas(ActiveSegmentQuotas(xs)); err != nil {
		t.Errorf("ValidateSegmentQuotas() = %v", err)
	}
}

func TestSegmentQuotaNormalizeClamps(t *testing.T) {
	x := SampleSegmentQuota("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentBucketRoundTrip(t *testing.T) {
	x := SampleSegmentBucket("segmentbucket-1", when)
	got, err := ParseSegmentBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentBucket() = %v", err)
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

func TestSegmentBucketParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentBucket(f); err == nil {
		t.Error("ParseSegmentBucket() accepted a record with no id")
	}
	f = SampleSegmentBucket("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSegmentBucket(f); err == nil {
		t.Error("ParseSegmentBucket() accepted a non-numeric priority")
	}
}

func TestSegmentBucketCollections(t *testing.T) {
	xs := []SegmentBucket{SampleSegmentBucket("a", when), SampleSegmentBucket("bb", when), SampleSegmentBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentBuckets(xs)); got != 2 {
		t.Errorf("ActiveSegmentBuckets() kept %d, want 2", got)
	}
	if got, want := SumSegmentBuckets(xs), 1+3; got != want {
		t.Errorf("SumSegmentBuckets() = %d, want %d", got, want)
	}
	SortSegmentBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentBuckets() put %q first", xs[0].ID)
	}
	if len(IndexSegmentBuckets(xs)) != 3 {
		t.Errorf("IndexSegmentBuckets() = %v", IndexSegmentBuckets(xs))
	}
	if err := ValidateSegmentBuckets(ActiveSegmentBuckets(xs)); err != nil {
		t.Errorf("ValidateSegmentBuckets() = %v", err)
	}
}

func TestSegmentBucketNormalizeClamps(t *testing.T) {
	x := SampleSegmentBucket("a", when)
	x.Category = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentLaneRoundTrip(t *testing.T) {
	x := SampleSegmentLane("segmentlane-1", when)
	got, err := ParseSegmentLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentLane() = %v", err)
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

func TestSegmentLaneParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentLane(f); err == nil {
		t.Error("ParseSegmentLane() accepted a record with no id")
	}
	f = SampleSegmentLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseSegmentLane(f); err == nil {
		t.Error("ParseSegmentLane() accepted a non-numeric attempts")
	}
}

func TestSegmentLaneCollections(t *testing.T) {
	xs := []SegmentLane{SampleSegmentLane("a", when), SampleSegmentLane("bb", when), SampleSegmentLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentLanes(xs)); got != 2 {
		t.Errorf("ActiveSegmentLanes() kept %d, want 2", got)
	}
	if got, want := SumSegmentLanes(xs), 1+3; got != want {
		t.Errorf("SumSegmentLanes() = %d, want %d", got, want)
	}
	SortSegmentLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentLanes() put %q first", xs[0].ID)
	}
	if len(IndexSegmentLanes(xs)) != 3 {
		t.Errorf("IndexSegmentLanes() = %v", IndexSegmentLanes(xs))
	}
	if err := ValidateSegmentLanes(ActiveSegmentLanes(xs)); err != nil {
		t.Errorf("ValidateSegmentLanes() = %v", err)
	}
}

func TestSegmentLaneNormalizeClamps(t *testing.T) {
	x := SampleSegmentLane("a", when)
	x.Reference = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentSlotRoundTrip(t *testing.T) {
	x := SampleSegmentSlot("segmentslot-1", when)
	got, err := ParseSegmentSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentSlot() = %v", err)
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

func TestSegmentSlotParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentSlot(f); err == nil {
		t.Error("ParseSegmentSlot() accepted a record with no id")
	}
	f = SampleSegmentSlot("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSegmentSlot(f); err == nil {
		t.Error("ParseSegmentSlot() accepted a non-numeric capacity")
	}
}

func TestSegmentSlotCollections(t *testing.T) {
	xs := []SegmentSlot{SampleSegmentSlot("a", when), SampleSegmentSlot("bb", when), SampleSegmentSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentSlots(xs)); got != 2 {
		t.Errorf("ActiveSegmentSlots() kept %d, want 2", got)
	}
	if got, want := SumSegmentSlots(xs), 1+3; got != want {
		t.Errorf("SumSegmentSlots() = %d, want %d", got, want)
	}
	SortSegmentSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentSlots() put %q first", xs[0].ID)
	}
	if len(IndexSegmentSlots(xs)) != 3 {
		t.Errorf("IndexSegmentSlots() = %v", IndexSegmentSlots(xs))
	}
	if err := ValidateSegmentSlots(ActiveSegmentSlots(xs)); err != nil {
		t.Errorf("ValidateSegmentSlots() = %v", err)
	}
}

func TestSegmentSlotNormalizeClamps(t *testing.T) {
	x := SampleSegmentSlot("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentTagRoundTrip(t *testing.T) {
	x := SampleSegmentTag("segmenttag-1", when)
	got, err := ParseSegmentTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentTag() = %v", err)
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

func TestSegmentTagParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentTag(f); err == nil {
		t.Error("ParseSegmentTag() accepted a record with no id")
	}
	f = SampleSegmentTag("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseSegmentTag(f); err == nil {
		t.Error("ParseSegmentTag() accepted a non-numeric retained")
	}
}

func TestSegmentTagCollections(t *testing.T) {
	xs := []SegmentTag{SampleSegmentTag("a", when), SampleSegmentTag("bb", when), SampleSegmentTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentTags(xs)); got != 2 {
		t.Errorf("ActiveSegmentTags() kept %d, want 2", got)
	}
	if got, want := SumSegmentTags(xs), 1+3; got != want {
		t.Errorf("SumSegmentTags() = %d, want %d", got, want)
	}
	SortSegmentTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentTags() put %q first", xs[0].ID)
	}
	if len(IndexSegmentTags(xs)) != 3 {
		t.Errorf("IndexSegmentTags() = %v", IndexSegmentTags(xs))
	}
	if err := ValidateSegmentTags(ActiveSegmentTags(xs)); err != nil {
		t.Errorf("ValidateSegmentTags() = %v", err)
	}
}

func TestSegmentTagNormalizeClamps(t *testing.T) {
	x := SampleSegmentTag("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentRuleRoundTrip(t *testing.T) {
	x := SampleSegmentRule("segmentrule-1", when)
	got, err := ParseSegmentRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentRule() = %v", err)
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

func TestSegmentRuleParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentRule(f); err == nil {
		t.Error("ParseSegmentRule() accepted a record with no id")
	}
	f = SampleSegmentRule("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseSegmentRule(f); err == nil {
		t.Error("ParseSegmentRule() accepted a non-numeric attempts")
	}
}

func TestSegmentRuleCollections(t *testing.T) {
	xs := []SegmentRule{SampleSegmentRule("a", when), SampleSegmentRule("bb", when), SampleSegmentRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentRules(xs)); got != 2 {
		t.Errorf("ActiveSegmentRules() kept %d, want 2", got)
	}
	if got, want := SumSegmentRules(xs), 1+3; got != want {
		t.Errorf("SumSegmentRules() = %d, want %d", got, want)
	}
	SortSegmentRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentRules() put %q first", xs[0].ID)
	}
	if len(IndexSegmentRules(xs)) != 3 {
		t.Errorf("IndexSegmentRules() = %v", IndexSegmentRules(xs))
	}
	if err := ValidateSegmentRules(ActiveSegmentRules(xs)); err != nil {
		t.Errorf("ValidateSegmentRules() = %v", err)
	}
}

func TestSegmentRuleNormalizeClamps(t *testing.T) {
	x := SampleSegmentRule("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentHoldRoundTrip(t *testing.T) {
	x := SampleSegmentHold("segmenthold-1", when)
	got, err := ParseSegmentHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentHold() = %v", err)
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

func TestSegmentHoldParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentHold(f); err == nil {
		t.Error("ParseSegmentHold() accepted a record with no id")
	}
	f = SampleSegmentHold("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSegmentHold(f); err == nil {
		t.Error("ParseSegmentHold() accepted a non-numeric version")
	}
}

func TestSegmentHoldCollections(t *testing.T) {
	xs := []SegmentHold{SampleSegmentHold("a", when), SampleSegmentHold("bb", when), SampleSegmentHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentHolds(xs)); got != 2 {
		t.Errorf("ActiveSegmentHolds() kept %d, want 2", got)
	}
	if got, want := SumSegmentHolds(xs), 1+3; got != want {
		t.Errorf("SumSegmentHolds() = %d, want %d", got, want)
	}
	SortSegmentHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentHolds() put %q first", xs[0].ID)
	}
	if len(IndexSegmentHolds(xs)) != 3 {
		t.Errorf("IndexSegmentHolds() = %v", IndexSegmentHolds(xs))
	}
	if err := ValidateSegmentHolds(ActiveSegmentHolds(xs)); err != nil {
		t.Errorf("ValidateSegmentHolds() = %v", err)
	}
}

func TestSegmentHoldNormalizeClamps(t *testing.T) {
	x := SampleSegmentHold("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentClaimRoundTrip(t *testing.T) {
	x := SampleSegmentClaim("segmentclaim-1", when)
	got, err := ParseSegmentClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentClaim() = %v", err)
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

func TestSegmentClaimParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentClaim(f); err == nil {
		t.Error("ParseSegmentClaim() accepted a record with no id")
	}
	f = SampleSegmentClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseSegmentClaim(f); err == nil {
		t.Error("ParseSegmentClaim() accepted a non-numeric retained")
	}
}

func TestSegmentClaimCollections(t *testing.T) {
	xs := []SegmentClaim{SampleSegmentClaim("a", when), SampleSegmentClaim("bb", when), SampleSegmentClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentClaims(xs)); got != 2 {
		t.Errorf("ActiveSegmentClaims() kept %d, want 2", got)
	}
	if got, want := SumSegmentClaims(xs), 1+3; got != want {
		t.Errorf("SumSegmentClaims() = %d, want %d", got, want)
	}
	SortSegmentClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentClaims() put %q first", xs[0].ID)
	}
	if len(IndexSegmentClaims(xs)) != 3 {
		t.Errorf("IndexSegmentClaims() = %v", IndexSegmentClaims(xs))
	}
	if err := ValidateSegmentClaims(ActiveSegmentClaims(xs)); err != nil {
		t.Errorf("ValidateSegmentClaims() = %v", err)
	}
}

func TestSegmentClaimNormalizeClamps(t *testing.T) {
	x := SampleSegmentClaim("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSegmentCursorRoundTrip(t *testing.T) {
	x := SampleSegmentCursor("segmentcursor-1", when)
	got, err := ParseSegmentCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseSegmentCursor() = %v", err)
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

func TestSegmentCursorParseRejectsBadInput(t *testing.T) {
	f := SampleSegmentCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSegmentCursor(f); err == nil {
		t.Error("ParseSegmentCursor() accepted a record with no id")
	}
	f = SampleSegmentCursor("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSegmentCursor(f); err == nil {
		t.Error("ParseSegmentCursor() accepted a non-numeric priority")
	}
}

func TestSegmentCursorCollections(t *testing.T) {
	xs := []SegmentCursor{SampleSegmentCursor("a", when), SampleSegmentCursor("bb", when), SampleSegmentCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSegmentCursors(xs)); got != 2 {
		t.Errorf("ActiveSegmentCursors() kept %d, want 2", got)
	}
	if got, want := SumSegmentCursors(xs), 1+3; got != want {
		t.Errorf("SumSegmentCursors() = %d, want %d", got, want)
	}
	SortSegmentCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSegmentCursors() put %q first", xs[0].ID)
	}
	if len(IndexSegmentCursors(xs)) != 3 {
		t.Errorf("IndexSegmentCursors() = %v", IndexSegmentCursors(xs))
	}
	if err := ValidateSegmentCursors(ActiveSegmentCursors(xs)); err != nil {
		t.Errorf("ValidateSegmentCursors() = %v", err)
	}
}

func TestSegmentCursorNormalizeClamps(t *testing.T) {
	x := SampleSegmentCursor("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
