package preference

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
		if !strings.HasPrefix(Prefixes[i], "preference/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SamplePreferenceRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestPreferenceRecordRoundTrip(t *testing.T) {
	x := SamplePreferenceRecord("preferencerecord-1", when)
	got, err := ParsePreferenceRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceRecord() = %v", err)
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

func TestPreferenceRecordParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceRecord(f); err == nil {
		t.Error("ParsePreferenceRecord() accepted a record with no id")
	}
	f = SamplePreferenceRecord("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePreferenceRecord(f); err == nil {
		t.Error("ParsePreferenceRecord() accepted a non-numeric offset")
	}
}

func TestPreferenceRecordCollections(t *testing.T) {
	xs := []PreferenceRecord{SamplePreferenceRecord("a", when), SamplePreferenceRecord("bb", when), SamplePreferenceRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceRecords(xs)); got != 2 {
		t.Errorf("ActivePreferenceRecords() kept %d, want 2", got)
	}
	if got, want := SumPreferenceRecords(xs), 1+3; got != want {
		t.Errorf("SumPreferenceRecords() = %d, want %d", got, want)
	}
	SortPreferenceRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceRecords() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceRecords(xs)) != 3 {
		t.Errorf("IndexPreferenceRecords() = %v", IndexPreferenceRecords(xs))
	}
	if err := ValidatePreferenceRecords(ActivePreferenceRecords(xs)); err != nil {
		t.Errorf("ValidatePreferenceRecords() = %v", err)
	}
}

func TestPreferenceRecordNormalizeClamps(t *testing.T) {
	x := SamplePreferenceRecord("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceEntryRoundTrip(t *testing.T) {
	x := SamplePreferenceEntry("preferenceentry-1", when)
	got, err := ParsePreferenceEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceEntry() = %v", err)
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

func TestPreferenceEntryParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceEntry(f); err == nil {
		t.Error("ParsePreferenceEntry() accepted a record with no id")
	}
	f = SamplePreferenceEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePreferenceEntry(f); err == nil {
		t.Error("ParsePreferenceEntry() accepted a non-numeric offset")
	}
}

func TestPreferenceEntryCollections(t *testing.T) {
	xs := []PreferenceEntry{SamplePreferenceEntry("a", when), SamplePreferenceEntry("bb", when), SamplePreferenceEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceEntrys(xs)); got != 2 {
		t.Errorf("ActivePreferenceEntrys() kept %d, want 2", got)
	}
	if got, want := SumPreferenceEntrys(xs), 1+3; got != want {
		t.Errorf("SumPreferenceEntrys() = %d, want %d", got, want)
	}
	SortPreferenceEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceEntrys() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceEntrys(xs)) != 3 {
		t.Errorf("IndexPreferenceEntrys() = %v", IndexPreferenceEntrys(xs))
	}
	if err := ValidatePreferenceEntrys(ActivePreferenceEntrys(xs)); err != nil {
		t.Errorf("ValidatePreferenceEntrys() = %v", err)
	}
}

func TestPreferenceEntryNormalizeClamps(t *testing.T) {
	x := SamplePreferenceEntry("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceItemRoundTrip(t *testing.T) {
	x := SamplePreferenceItem("preferenceitem-1", when)
	got, err := ParsePreferenceItem(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceItem() = %v", err)
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

func TestPreferenceItemParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceItem(f); err == nil {
		t.Error("ParsePreferenceItem() accepted a record with no id")
	}
	f = SamplePreferenceItem("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePreferenceItem(f); err == nil {
		t.Error("ParsePreferenceItem() accepted a non-numeric attempts")
	}
}

func TestPreferenceItemCollections(t *testing.T) {
	xs := []PreferenceItem{SamplePreferenceItem("a", when), SamplePreferenceItem("bb", when), SamplePreferenceItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceItems(xs)); got != 2 {
		t.Errorf("ActivePreferenceItems() kept %d, want 2", got)
	}
	if got, want := SumPreferenceItems(xs), 1+3; got != want {
		t.Errorf("SumPreferenceItems() = %d, want %d", got, want)
	}
	SortPreferenceItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceItems() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceItems(xs)) != 3 {
		t.Errorf("IndexPreferenceItems() = %v", IndexPreferenceItems(xs))
	}
	if err := ValidatePreferenceItems(ActivePreferenceItems(xs)); err != nil {
		t.Errorf("ValidatePreferenceItems() = %v", err)
	}
}

func TestPreferenceItemNormalizeClamps(t *testing.T) {
	x := SamplePreferenceItem("a", when)
	x.Category = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceDraftRoundTrip(t *testing.T) {
	x := SamplePreferenceDraft("preferencedraft-1", when)
	got, err := ParsePreferenceDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceDraft() = %v", err)
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

func TestPreferenceDraftParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceDraft(f); err == nil {
		t.Error("ParsePreferenceDraft() accepted a record with no id")
	}
	f = SamplePreferenceDraft("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParsePreferenceDraft(f); err == nil {
		t.Error("ParsePreferenceDraft() accepted a non-numeric weight")
	}
}

func TestPreferenceDraftCollections(t *testing.T) {
	xs := []PreferenceDraft{SamplePreferenceDraft("a", when), SamplePreferenceDraft("bb", when), SamplePreferenceDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceDrafts(xs)); got != 2 {
		t.Errorf("ActivePreferenceDrafts() kept %d, want 2", got)
	}
	if got, want := SumPreferenceDrafts(xs), 1+3; got != want {
		t.Errorf("SumPreferenceDrafts() = %d, want %d", got, want)
	}
	SortPreferenceDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceDrafts() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceDrafts(xs)) != 3 {
		t.Errorf("IndexPreferenceDrafts() = %v", IndexPreferenceDrafts(xs))
	}
	if err := ValidatePreferenceDrafts(ActivePreferenceDrafts(xs)); err != nil {
		t.Errorf("ValidatePreferenceDrafts() = %v", err)
	}
}

func TestPreferenceDraftNormalizeClamps(t *testing.T) {
	x := SamplePreferenceDraft("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceRevisionRoundTrip(t *testing.T) {
	x := SamplePreferenceRevision("preferencerevision-1", when)
	got, err := ParsePreferenceRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceRevision() = %v", err)
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

func TestPreferenceRevisionParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceRevision(f); err == nil {
		t.Error("ParsePreferenceRevision() accepted a record with no id")
	}
	f = SamplePreferenceRevision("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePreferenceRevision(f); err == nil {
		t.Error("ParsePreferenceRevision() accepted a non-numeric attempts")
	}
}

func TestPreferenceRevisionCollections(t *testing.T) {
	xs := []PreferenceRevision{SamplePreferenceRevision("a", when), SamplePreferenceRevision("bb", when), SamplePreferenceRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceRevisions(xs)); got != 2 {
		t.Errorf("ActivePreferenceRevisions() kept %d, want 2", got)
	}
	if got, want := SumPreferenceRevisions(xs), 1+3; got != want {
		t.Errorf("SumPreferenceRevisions() = %d, want %d", got, want)
	}
	SortPreferenceRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceRevisions() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceRevisions(xs)) != 3 {
		t.Errorf("IndexPreferenceRevisions() = %v", IndexPreferenceRevisions(xs))
	}
	if err := ValidatePreferenceRevisions(ActivePreferenceRevisions(xs)); err != nil {
		t.Errorf("ValidatePreferenceRevisions() = %v", err)
	}
}

func TestPreferenceRevisionNormalizeClamps(t *testing.T) {
	x := SamplePreferenceRevision("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceSnapshotRoundTrip(t *testing.T) {
	x := SamplePreferenceSnapshot("preferencesnapshot-1", when)
	got, err := ParsePreferenceSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceSnapshot() = %v", err)
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

func TestPreferenceSnapshotParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceSnapshot(f); err == nil {
		t.Error("ParsePreferenceSnapshot() accepted a record with no id")
	}
	f = SamplePreferenceSnapshot("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePreferenceSnapshot(f); err == nil {
		t.Error("ParsePreferenceSnapshot() accepted a non-numeric capacity")
	}
}

func TestPreferenceSnapshotCollections(t *testing.T) {
	xs := []PreferenceSnapshot{SamplePreferenceSnapshot("a", when), SamplePreferenceSnapshot("bb", when), SamplePreferenceSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceSnapshots(xs)); got != 2 {
		t.Errorf("ActivePreferenceSnapshots() kept %d, want 2", got)
	}
	if got, want := SumPreferenceSnapshots(xs), 1+3; got != want {
		t.Errorf("SumPreferenceSnapshots() = %d, want %d", got, want)
	}
	SortPreferenceSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceSnapshots(xs)) != 3 {
		t.Errorf("IndexPreferenceSnapshots() = %v", IndexPreferenceSnapshots(xs))
	}
	if err := ValidatePreferenceSnapshots(ActivePreferenceSnapshots(xs)); err != nil {
		t.Errorf("ValidatePreferenceSnapshots() = %v", err)
	}
}

func TestPreferenceSnapshotNormalizeClamps(t *testing.T) {
	x := SamplePreferenceSnapshot("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceBatchRoundTrip(t *testing.T) {
	x := SamplePreferenceBatch("preferencebatch-1", when)
	got, err := ParsePreferenceBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceBatch() = %v", err)
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

func TestPreferenceBatchParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceBatch(f); err == nil {
		t.Error("ParsePreferenceBatch() accepted a record with no id")
	}
	f = SamplePreferenceBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePreferenceBatch(f); err == nil {
		t.Error("ParsePreferenceBatch() accepted a non-numeric capacity")
	}
}

func TestPreferenceBatchCollections(t *testing.T) {
	xs := []PreferenceBatch{SamplePreferenceBatch("a", when), SamplePreferenceBatch("bb", when), SamplePreferenceBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceBatchs(xs)); got != 2 {
		t.Errorf("ActivePreferenceBatchs() kept %d, want 2", got)
	}
	if got, want := SumPreferenceBatchs(xs), 1+3; got != want {
		t.Errorf("SumPreferenceBatchs() = %d, want %d", got, want)
	}
	SortPreferenceBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceBatchs() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceBatchs(xs)) != 3 {
		t.Errorf("IndexPreferenceBatchs() = %v", IndexPreferenceBatchs(xs))
	}
	if err := ValidatePreferenceBatchs(ActivePreferenceBatchs(xs)); err != nil {
		t.Errorf("ValidatePreferenceBatchs() = %v", err)
	}
}

func TestPreferenceBatchNormalizeClamps(t *testing.T) {
	x := SamplePreferenceBatch("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceRequestRoundTrip(t *testing.T) {
	x := SamplePreferenceRequest("preferencerequest-1", when)
	got, err := ParsePreferenceRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceRequest() = %v", err)
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

func TestPreferenceRequestParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceRequest(f); err == nil {
		t.Error("ParsePreferenceRequest() accepted a record with no id")
	}
	f = SamplePreferenceRequest("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePreferenceRequest(f); err == nil {
		t.Error("ParsePreferenceRequest() accepted a non-numeric retained")
	}
}

func TestPreferenceRequestCollections(t *testing.T) {
	xs := []PreferenceRequest{SamplePreferenceRequest("a", when), SamplePreferenceRequest("bb", when), SamplePreferenceRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceRequests(xs)); got != 2 {
		t.Errorf("ActivePreferenceRequests() kept %d, want 2", got)
	}
	if got, want := SumPreferenceRequests(xs), 1+3; got != want {
		t.Errorf("SumPreferenceRequests() = %d, want %d", got, want)
	}
	SortPreferenceRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceRequests() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceRequests(xs)) != 3 {
		t.Errorf("IndexPreferenceRequests() = %v", IndexPreferenceRequests(xs))
	}
	if err := ValidatePreferenceRequests(ActivePreferenceRequests(xs)); err != nil {
		t.Errorf("ValidatePreferenceRequests() = %v", err)
	}
}

func TestPreferenceRequestNormalizeClamps(t *testing.T) {
	x := SamplePreferenceRequest("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceSummaryRoundTrip(t *testing.T) {
	x := SamplePreferenceSummary("preferencesummary-1", when)
	got, err := ParsePreferenceSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceSummary() = %v", err)
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

func TestPreferenceSummaryParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceSummary(f); err == nil {
		t.Error("ParsePreferenceSummary() accepted a record with no id")
	}
	f = SamplePreferenceSummary("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePreferenceSummary(f); err == nil {
		t.Error("ParsePreferenceSummary() accepted a non-numeric count")
	}
}

func TestPreferenceSummaryCollections(t *testing.T) {
	xs := []PreferenceSummary{SamplePreferenceSummary("a", when), SamplePreferenceSummary("bb", when), SamplePreferenceSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceSummarys(xs)); got != 2 {
		t.Errorf("ActivePreferenceSummarys() kept %d, want 2", got)
	}
	if got, want := SumPreferenceSummarys(xs), 1+3; got != want {
		t.Errorf("SumPreferenceSummarys() = %d, want %d", got, want)
	}
	SortPreferenceSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceSummarys() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceSummarys(xs)) != 3 {
		t.Errorf("IndexPreferenceSummarys() = %v", IndexPreferenceSummarys(xs))
	}
	if err := ValidatePreferenceSummarys(ActivePreferenceSummarys(xs)); err != nil {
		t.Errorf("ValidatePreferenceSummarys() = %v", err)
	}
}

func TestPreferenceSummaryNormalizeClamps(t *testing.T) {
	x := SamplePreferenceSummary("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferencePolicyRoundTrip(t *testing.T) {
	x := SamplePreferencePolicy("preferencepolicy-1", when)
	got, err := ParsePreferencePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferencePolicy() = %v", err)
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

func TestPreferencePolicyParseRejectsBadInput(t *testing.T) {
	f := SamplePreferencePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferencePolicy(f); err == nil {
		t.Error("ParsePreferencePolicy() accepted a record with no id")
	}
	f = SamplePreferencePolicy("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePreferencePolicy(f); err == nil {
		t.Error("ParsePreferencePolicy() accepted a non-numeric retained")
	}
}

func TestPreferencePolicyCollections(t *testing.T) {
	xs := []PreferencePolicy{SamplePreferencePolicy("a", when), SamplePreferencePolicy("bb", when), SamplePreferencePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferencePolicys(xs)); got != 2 {
		t.Errorf("ActivePreferencePolicys() kept %d, want 2", got)
	}
	if got, want := SumPreferencePolicys(xs), 1+3; got != want {
		t.Errorf("SumPreferencePolicys() = %d, want %d", got, want)
	}
	SortPreferencePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferencePolicys() put %q first", xs[0].ID)
	}
	if len(IndexPreferencePolicys(xs)) != 3 {
		t.Errorf("IndexPreferencePolicys() = %v", IndexPreferencePolicys(xs))
	}
	if err := ValidatePreferencePolicys(ActivePreferencePolicys(xs)); err != nil {
		t.Errorf("ValidatePreferencePolicys() = %v", err)
	}
}

func TestPreferencePolicyNormalizeClamps(t *testing.T) {
	x := SamplePreferencePolicy("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceLinkRoundTrip(t *testing.T) {
	x := SamplePreferenceLink("preferencelink-1", when)
	got, err := ParsePreferenceLink(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceLink() = %v", err)
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

func TestPreferenceLinkParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceLink(f); err == nil {
		t.Error("ParsePreferenceLink() accepted a record with no id")
	}
	f = SamplePreferenceLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePreferenceLink(f); err == nil {
		t.Error("ParsePreferenceLink() accepted a non-numeric retained")
	}
}

func TestPreferenceLinkCollections(t *testing.T) {
	xs := []PreferenceLink{SamplePreferenceLink("a", when), SamplePreferenceLink("bb", when), SamplePreferenceLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceLinks(xs)); got != 2 {
		t.Errorf("ActivePreferenceLinks() kept %d, want 2", got)
	}
	if got, want := SumPreferenceLinks(xs), 1+3; got != want {
		t.Errorf("SumPreferenceLinks() = %d, want %d", got, want)
	}
	SortPreferenceLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceLinks() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceLinks(xs)) != 3 {
		t.Errorf("IndexPreferenceLinks() = %v", IndexPreferenceLinks(xs))
	}
	if err := ValidatePreferenceLinks(ActivePreferenceLinks(xs)); err != nil {
		t.Errorf("ValidatePreferenceLinks() = %v", err)
	}
}

func TestPreferenceLinkNormalizeClamps(t *testing.T) {
	x := SamplePreferenceLink("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceNoteRoundTrip(t *testing.T) {
	x := SamplePreferenceNote("preferencenote-1", when)
	got, err := ParsePreferenceNote(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceNote() = %v", err)
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

func TestPreferenceNoteParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceNote(f); err == nil {
		t.Error("ParsePreferenceNote() accepted a record with no id")
	}
	f = SamplePreferenceNote("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePreferenceNote(f); err == nil {
		t.Error("ParsePreferenceNote() accepted a non-numeric offset")
	}
}

func TestPreferenceNoteCollections(t *testing.T) {
	xs := []PreferenceNote{SamplePreferenceNote("a", when), SamplePreferenceNote("bb", when), SamplePreferenceNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceNotes(xs)); got != 2 {
		t.Errorf("ActivePreferenceNotes() kept %d, want 2", got)
	}
	if got, want := SumPreferenceNotes(xs), 1+3; got != want {
		t.Errorf("SumPreferenceNotes() = %d, want %d", got, want)
	}
	SortPreferenceNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceNotes() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceNotes(xs)) != 3 {
		t.Errorf("IndexPreferenceNotes() = %v", IndexPreferenceNotes(xs))
	}
	if err := ValidatePreferenceNotes(ActivePreferenceNotes(xs)); err != nil {
		t.Errorf("ValidatePreferenceNotes() = %v", err)
	}
}

func TestPreferenceNoteNormalizeClamps(t *testing.T) {
	x := SamplePreferenceNote("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceWindowRoundTrip(t *testing.T) {
	x := SamplePreferenceWindow("preferencewindow-1", when)
	got, err := ParsePreferenceWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceWindow() = %v", err)
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

func TestPreferenceWindowParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceWindow(f); err == nil {
		t.Error("ParsePreferenceWindow() accepted a record with no id")
	}
	f = SamplePreferenceWindow("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePreferenceWindow(f); err == nil {
		t.Error("ParsePreferenceWindow() accepted a non-numeric offset")
	}
}

func TestPreferenceWindowCollections(t *testing.T) {
	xs := []PreferenceWindow{SamplePreferenceWindow("a", when), SamplePreferenceWindow("bb", when), SamplePreferenceWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceWindows(xs)); got != 2 {
		t.Errorf("ActivePreferenceWindows() kept %d, want 2", got)
	}
	if got, want := SumPreferenceWindows(xs), 1+3; got != want {
		t.Errorf("SumPreferenceWindows() = %d, want %d", got, want)
	}
	SortPreferenceWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceWindows() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceWindows(xs)) != 3 {
		t.Errorf("IndexPreferenceWindows() = %v", IndexPreferenceWindows(xs))
	}
	if err := ValidatePreferenceWindows(ActivePreferenceWindows(xs)); err != nil {
		t.Errorf("ValidatePreferenceWindows() = %v", err)
	}
}

func TestPreferenceWindowNormalizeClamps(t *testing.T) {
	x := SamplePreferenceWindow("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceQuotaRoundTrip(t *testing.T) {
	x := SamplePreferenceQuota("preferencequota-1", when)
	got, err := ParsePreferenceQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceQuota() = %v", err)
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

func TestPreferenceQuotaParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceQuota(f); err == nil {
		t.Error("ParsePreferenceQuota() accepted a record with no id")
	}
	f = SamplePreferenceQuota("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePreferenceQuota(f); err == nil {
		t.Error("ParsePreferenceQuota() accepted a non-numeric attempts")
	}
}

func TestPreferenceQuotaCollections(t *testing.T) {
	xs := []PreferenceQuota{SamplePreferenceQuota("a", when), SamplePreferenceQuota("bb", when), SamplePreferenceQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceQuotas(xs)); got != 2 {
		t.Errorf("ActivePreferenceQuotas() kept %d, want 2", got)
	}
	if got, want := SumPreferenceQuotas(xs), 1+3; got != want {
		t.Errorf("SumPreferenceQuotas() = %d, want %d", got, want)
	}
	SortPreferenceQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceQuotas() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceQuotas(xs)) != 3 {
		t.Errorf("IndexPreferenceQuotas() = %v", IndexPreferenceQuotas(xs))
	}
	if err := ValidatePreferenceQuotas(ActivePreferenceQuotas(xs)); err != nil {
		t.Errorf("ValidatePreferenceQuotas() = %v", err)
	}
}

func TestPreferenceQuotaNormalizeClamps(t *testing.T) {
	x := SamplePreferenceQuota("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceBucketRoundTrip(t *testing.T) {
	x := SamplePreferenceBucket("preferencebucket-1", when)
	got, err := ParsePreferenceBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceBucket() = %v", err)
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

func TestPreferenceBucketParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceBucket(f); err == nil {
		t.Error("ParsePreferenceBucket() accepted a record with no id")
	}
	f = SamplePreferenceBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePreferenceBucket(f); err == nil {
		t.Error("ParsePreferenceBucket() accepted a non-numeric retained")
	}
}

func TestPreferenceBucketCollections(t *testing.T) {
	xs := []PreferenceBucket{SamplePreferenceBucket("a", when), SamplePreferenceBucket("bb", when), SamplePreferenceBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceBuckets(xs)); got != 2 {
		t.Errorf("ActivePreferenceBuckets() kept %d, want 2", got)
	}
	if got, want := SumPreferenceBuckets(xs), 1+3; got != want {
		t.Errorf("SumPreferenceBuckets() = %d, want %d", got, want)
	}
	SortPreferenceBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceBuckets() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceBuckets(xs)) != 3 {
		t.Errorf("IndexPreferenceBuckets() = %v", IndexPreferenceBuckets(xs))
	}
	if err := ValidatePreferenceBuckets(ActivePreferenceBuckets(xs)); err != nil {
		t.Errorf("ValidatePreferenceBuckets() = %v", err)
	}
}

func TestPreferenceBucketNormalizeClamps(t *testing.T) {
	x := SamplePreferenceBucket("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceLaneRoundTrip(t *testing.T) {
	x := SamplePreferenceLane("preferencelane-1", when)
	got, err := ParsePreferenceLane(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceLane() = %v", err)
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

func TestPreferenceLaneParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceLane(f); err == nil {
		t.Error("ParsePreferenceLane() accepted a record with no id")
	}
	f = SamplePreferenceLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePreferenceLane(f); err == nil {
		t.Error("ParsePreferenceLane() accepted a non-numeric attempts")
	}
}

func TestPreferenceLaneCollections(t *testing.T) {
	xs := []PreferenceLane{SamplePreferenceLane("a", when), SamplePreferenceLane("bb", when), SamplePreferenceLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceLanes(xs)); got != 2 {
		t.Errorf("ActivePreferenceLanes() kept %d, want 2", got)
	}
	if got, want := SumPreferenceLanes(xs), 1+3; got != want {
		t.Errorf("SumPreferenceLanes() = %d, want %d", got, want)
	}
	SortPreferenceLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceLanes() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceLanes(xs)) != 3 {
		t.Errorf("IndexPreferenceLanes() = %v", IndexPreferenceLanes(xs))
	}
	if err := ValidatePreferenceLanes(ActivePreferenceLanes(xs)); err != nil {
		t.Errorf("ValidatePreferenceLanes() = %v", err)
	}
}

func TestPreferenceLaneNormalizeClamps(t *testing.T) {
	x := SamplePreferenceLane("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceSlotRoundTrip(t *testing.T) {
	x := SamplePreferenceSlot("preferenceslot-1", when)
	got, err := ParsePreferenceSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceSlot() = %v", err)
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

func TestPreferenceSlotParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceSlot(f); err == nil {
		t.Error("ParsePreferenceSlot() accepted a record with no id")
	}
	f = SamplePreferenceSlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePreferenceSlot(f); err == nil {
		t.Error("ParsePreferenceSlot() accepted a non-numeric count")
	}
}

func TestPreferenceSlotCollections(t *testing.T) {
	xs := []PreferenceSlot{SamplePreferenceSlot("a", when), SamplePreferenceSlot("bb", when), SamplePreferenceSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceSlots(xs)); got != 2 {
		t.Errorf("ActivePreferenceSlots() kept %d, want 2", got)
	}
	if got, want := SumPreferenceSlots(xs), 1+3; got != want {
		t.Errorf("SumPreferenceSlots() = %d, want %d", got, want)
	}
	SortPreferenceSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceSlots() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceSlots(xs)) != 3 {
		t.Errorf("IndexPreferenceSlots() = %v", IndexPreferenceSlots(xs))
	}
	if err := ValidatePreferenceSlots(ActivePreferenceSlots(xs)); err != nil {
		t.Errorf("ValidatePreferenceSlots() = %v", err)
	}
}

func TestPreferenceSlotNormalizeClamps(t *testing.T) {
	x := SamplePreferenceSlot("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceTagRoundTrip(t *testing.T) {
	x := SamplePreferenceTag("preferencetag-1", when)
	got, err := ParsePreferenceTag(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceTag() = %v", err)
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

func TestPreferenceTagParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceTag(f); err == nil {
		t.Error("ParsePreferenceTag() accepted a record with no id")
	}
	f = SamplePreferenceTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePreferenceTag(f); err == nil {
		t.Error("ParsePreferenceTag() accepted a non-numeric count")
	}
}

func TestPreferenceTagCollections(t *testing.T) {
	xs := []PreferenceTag{SamplePreferenceTag("a", when), SamplePreferenceTag("bb", when), SamplePreferenceTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceTags(xs)); got != 2 {
		t.Errorf("ActivePreferenceTags() kept %d, want 2", got)
	}
	if got, want := SumPreferenceTags(xs), 1+3; got != want {
		t.Errorf("SumPreferenceTags() = %d, want %d", got, want)
	}
	SortPreferenceTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceTags() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceTags(xs)) != 3 {
		t.Errorf("IndexPreferenceTags() = %v", IndexPreferenceTags(xs))
	}
	if err := ValidatePreferenceTags(ActivePreferenceTags(xs)); err != nil {
		t.Errorf("ValidatePreferenceTags() = %v", err)
	}
}

func TestPreferenceTagNormalizeClamps(t *testing.T) {
	x := SamplePreferenceTag("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceRuleRoundTrip(t *testing.T) {
	x := SamplePreferenceRule("preferencerule-1", when)
	got, err := ParsePreferenceRule(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceRule() = %v", err)
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

func TestPreferenceRuleParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceRule(f); err == nil {
		t.Error("ParsePreferenceRule() accepted a record with no id")
	}
	f = SamplePreferenceRule("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParsePreferenceRule(f); err == nil {
		t.Error("ParsePreferenceRule() accepted a non-numeric priority")
	}
}

func TestPreferenceRuleCollections(t *testing.T) {
	xs := []PreferenceRule{SamplePreferenceRule("a", when), SamplePreferenceRule("bb", when), SamplePreferenceRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceRules(xs)); got != 2 {
		t.Errorf("ActivePreferenceRules() kept %d, want 2", got)
	}
	if got, want := SumPreferenceRules(xs), 1+3; got != want {
		t.Errorf("SumPreferenceRules() = %d, want %d", got, want)
	}
	SortPreferenceRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceRules() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceRules(xs)) != 3 {
		t.Errorf("IndexPreferenceRules() = %v", IndexPreferenceRules(xs))
	}
	if err := ValidatePreferenceRules(ActivePreferenceRules(xs)); err != nil {
		t.Errorf("ValidatePreferenceRules() = %v", err)
	}
}

func TestPreferenceRuleNormalizeClamps(t *testing.T) {
	x := SamplePreferenceRule("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceHoldRoundTrip(t *testing.T) {
	x := SamplePreferenceHold("preferencehold-1", when)
	got, err := ParsePreferenceHold(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceHold() = %v", err)
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

func TestPreferenceHoldParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceHold(f); err == nil {
		t.Error("ParsePreferenceHold() accepted a record with no id")
	}
	f = SamplePreferenceHold("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePreferenceHold(f); err == nil {
		t.Error("ParsePreferenceHold() accepted a non-numeric attempts")
	}
}

func TestPreferenceHoldCollections(t *testing.T) {
	xs := []PreferenceHold{SamplePreferenceHold("a", when), SamplePreferenceHold("bb", when), SamplePreferenceHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceHolds(xs)); got != 2 {
		t.Errorf("ActivePreferenceHolds() kept %d, want 2", got)
	}
	if got, want := SumPreferenceHolds(xs), 1+3; got != want {
		t.Errorf("SumPreferenceHolds() = %d, want %d", got, want)
	}
	SortPreferenceHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceHolds() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceHolds(xs)) != 3 {
		t.Errorf("IndexPreferenceHolds() = %v", IndexPreferenceHolds(xs))
	}
	if err := ValidatePreferenceHolds(ActivePreferenceHolds(xs)); err != nil {
		t.Errorf("ValidatePreferenceHolds() = %v", err)
	}
}

func TestPreferenceHoldNormalizeClamps(t *testing.T) {
	x := SamplePreferenceHold("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceClaimRoundTrip(t *testing.T) {
	x := SamplePreferenceClaim("preferenceclaim-1", when)
	got, err := ParsePreferenceClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceClaim() = %v", err)
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

func TestPreferenceClaimParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceClaim(f); err == nil {
		t.Error("ParsePreferenceClaim() accepted a record with no id")
	}
	f = SamplePreferenceClaim("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePreferenceClaim(f); err == nil {
		t.Error("ParsePreferenceClaim() accepted a non-numeric version")
	}
}

func TestPreferenceClaimCollections(t *testing.T) {
	xs := []PreferenceClaim{SamplePreferenceClaim("a", when), SamplePreferenceClaim("bb", when), SamplePreferenceClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceClaims(xs)); got != 2 {
		t.Errorf("ActivePreferenceClaims() kept %d, want 2", got)
	}
	if got, want := SumPreferenceClaims(xs), 1+3; got != want {
		t.Errorf("SumPreferenceClaims() = %d, want %d", got, want)
	}
	SortPreferenceClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceClaims() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceClaims(xs)) != 3 {
		t.Errorf("IndexPreferenceClaims() = %v", IndexPreferenceClaims(xs))
	}
	if err := ValidatePreferenceClaims(ActivePreferenceClaims(xs)); err != nil {
		t.Errorf("ValidatePreferenceClaims() = %v", err)
	}
}

func TestPreferenceClaimNormalizeClamps(t *testing.T) {
	x := SamplePreferenceClaim("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPreferenceCursorRoundTrip(t *testing.T) {
	x := SamplePreferenceCursor("preferencecursor-1", when)
	got, err := ParsePreferenceCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParsePreferenceCursor() = %v", err)
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

func TestPreferenceCursorParseRejectsBadInput(t *testing.T) {
	f := SamplePreferenceCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePreferenceCursor(f); err == nil {
		t.Error("ParsePreferenceCursor() accepted a record with no id")
	}
	f = SamplePreferenceCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParsePreferenceCursor(f); err == nil {
		t.Error("ParsePreferenceCursor() accepted a non-numeric weight")
	}
}

func TestPreferenceCursorCollections(t *testing.T) {
	xs := []PreferenceCursor{SamplePreferenceCursor("a", when), SamplePreferenceCursor("bb", when), SamplePreferenceCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePreferenceCursors(xs)); got != 2 {
		t.Errorf("ActivePreferenceCursors() kept %d, want 2", got)
	}
	if got, want := SumPreferenceCursors(xs), 1+3; got != want {
		t.Errorf("SumPreferenceCursors() = %d, want %d", got, want)
	}
	SortPreferenceCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPreferenceCursors() put %q first", xs[0].ID)
	}
	if len(IndexPreferenceCursors(xs)) != 3 {
		t.Errorf("IndexPreferenceCursors() = %v", IndexPreferenceCursors(xs))
	}
	if err := ValidatePreferenceCursors(ActivePreferenceCursors(xs)); err != nil {
		t.Errorf("ValidatePreferenceCursors() = %v", err)
	}
}

func TestPreferenceCursorNormalizeClamps(t *testing.T) {
	x := SamplePreferenceCursor("a", when)
	x.Region = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
