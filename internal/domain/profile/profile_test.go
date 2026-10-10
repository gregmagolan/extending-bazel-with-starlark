package profile

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
		if !strings.HasPrefix(Prefixes[i], "profile/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleProfileRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestProfileRecordRoundTrip(t *testing.T) {
	x := SampleProfileRecord("profilerecord-1", when)
	got, err := ParseProfileRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileRecord() = %v", err)
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

func TestProfileRecordParseRejectsBadInput(t *testing.T) {
	f := SampleProfileRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileRecord(f); err == nil {
		t.Error("ParseProfileRecord() accepted a record with no id")
	}
	f = SampleProfileRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseProfileRecord(f); err == nil {
		t.Error("ParseProfileRecord() accepted a non-numeric capacity")
	}
}

func TestProfileRecordCollections(t *testing.T) {
	xs := []ProfileRecord{SampleProfileRecord("a", when), SampleProfileRecord("bb", when), SampleProfileRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileRecords(xs)); got != 2 {
		t.Errorf("ActiveProfileRecords() kept %d, want 2", got)
	}
	if got, want := SumProfileRecords(xs), 1+3; got != want {
		t.Errorf("SumProfileRecords() = %d, want %d", got, want)
	}
	SortProfileRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileRecords() put %q first", xs[0].ID)
	}
	if len(IndexProfileRecords(xs)) != 3 {
		t.Errorf("IndexProfileRecords() = %v", IndexProfileRecords(xs))
	}
	if err := ValidateProfileRecords(ActiveProfileRecords(xs)); err != nil {
		t.Errorf("ValidateProfileRecords() = %v", err)
	}
}

func TestProfileRecordNormalizeClamps(t *testing.T) {
	x := SampleProfileRecord("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileEntryRoundTrip(t *testing.T) {
	x := SampleProfileEntry("profileentry-1", when)
	got, err := ParseProfileEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileEntry() = %v", err)
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

func TestProfileEntryParseRejectsBadInput(t *testing.T) {
	f := SampleProfileEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileEntry(f); err == nil {
		t.Error("ParseProfileEntry() accepted a record with no id")
	}
	f = SampleProfileEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseProfileEntry(f); err == nil {
		t.Error("ParseProfileEntry() accepted a non-numeric offset")
	}
}

func TestProfileEntryCollections(t *testing.T) {
	xs := []ProfileEntry{SampleProfileEntry("a", when), SampleProfileEntry("bb", when), SampleProfileEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileEntrys(xs)); got != 2 {
		t.Errorf("ActiveProfileEntrys() kept %d, want 2", got)
	}
	if got, want := SumProfileEntrys(xs), 1+3; got != want {
		t.Errorf("SumProfileEntrys() = %d, want %d", got, want)
	}
	SortProfileEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileEntrys() put %q first", xs[0].ID)
	}
	if len(IndexProfileEntrys(xs)) != 3 {
		t.Errorf("IndexProfileEntrys() = %v", IndexProfileEntrys(xs))
	}
	if err := ValidateProfileEntrys(ActiveProfileEntrys(xs)); err != nil {
		t.Errorf("ValidateProfileEntrys() = %v", err)
	}
}

func TestProfileEntryNormalizeClamps(t *testing.T) {
	x := SampleProfileEntry("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileItemRoundTrip(t *testing.T) {
	x := SampleProfileItem("profileitem-1", when)
	got, err := ParseProfileItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileItem() = %v", err)
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

func TestProfileItemParseRejectsBadInput(t *testing.T) {
	f := SampleProfileItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileItem(f); err == nil {
		t.Error("ParseProfileItem() accepted a record with no id")
	}
	f = SampleProfileItem("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseProfileItem(f); err == nil {
		t.Error("ParseProfileItem() accepted a non-numeric offset")
	}
}

func TestProfileItemCollections(t *testing.T) {
	xs := []ProfileItem{SampleProfileItem("a", when), SampleProfileItem("bb", when), SampleProfileItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileItems(xs)); got != 2 {
		t.Errorf("ActiveProfileItems() kept %d, want 2", got)
	}
	if got, want := SumProfileItems(xs), 1+3; got != want {
		t.Errorf("SumProfileItems() = %d, want %d", got, want)
	}
	SortProfileItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileItems() put %q first", xs[0].ID)
	}
	if len(IndexProfileItems(xs)) != 3 {
		t.Errorf("IndexProfileItems() = %v", IndexProfileItems(xs))
	}
	if err := ValidateProfileItems(ActiveProfileItems(xs)); err != nil {
		t.Errorf("ValidateProfileItems() = %v", err)
	}
}

func TestProfileItemNormalizeClamps(t *testing.T) {
	x := SampleProfileItem("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileDraftRoundTrip(t *testing.T) {
	x := SampleProfileDraft("profiledraft-1", when)
	got, err := ParseProfileDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileDraft() = %v", err)
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

func TestProfileDraftParseRejectsBadInput(t *testing.T) {
	f := SampleProfileDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileDraft(f); err == nil {
		t.Error("ParseProfileDraft() accepted a record with no id")
	}
	f = SampleProfileDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseProfileDraft(f); err == nil {
		t.Error("ParseProfileDraft() accepted a non-numeric attempts")
	}
}

func TestProfileDraftCollections(t *testing.T) {
	xs := []ProfileDraft{SampleProfileDraft("a", when), SampleProfileDraft("bb", when), SampleProfileDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileDrafts(xs)); got != 2 {
		t.Errorf("ActiveProfileDrafts() kept %d, want 2", got)
	}
	if got, want := SumProfileDrafts(xs), 1+3; got != want {
		t.Errorf("SumProfileDrafts() = %d, want %d", got, want)
	}
	SortProfileDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileDrafts() put %q first", xs[0].ID)
	}
	if len(IndexProfileDrafts(xs)) != 3 {
		t.Errorf("IndexProfileDrafts() = %v", IndexProfileDrafts(xs))
	}
	if err := ValidateProfileDrafts(ActiveProfileDrafts(xs)); err != nil {
		t.Errorf("ValidateProfileDrafts() = %v", err)
	}
}

func TestProfileDraftNormalizeClamps(t *testing.T) {
	x := SampleProfileDraft("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileRevisionRoundTrip(t *testing.T) {
	x := SampleProfileRevision("profilerevision-1", when)
	got, err := ParseProfileRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileRevision() = %v", err)
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

func TestProfileRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleProfileRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileRevision(f); err == nil {
		t.Error("ParseProfileRevision() accepted a record with no id")
	}
	f = SampleProfileRevision("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseProfileRevision(f); err == nil {
		t.Error("ParseProfileRevision() accepted a non-numeric retained")
	}
}

func TestProfileRevisionCollections(t *testing.T) {
	xs := []ProfileRevision{SampleProfileRevision("a", when), SampleProfileRevision("bb", when), SampleProfileRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileRevisions(xs)); got != 2 {
		t.Errorf("ActiveProfileRevisions() kept %d, want 2", got)
	}
	if got, want := SumProfileRevisions(xs), 1+3; got != want {
		t.Errorf("SumProfileRevisions() = %d, want %d", got, want)
	}
	SortProfileRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileRevisions() put %q first", xs[0].ID)
	}
	if len(IndexProfileRevisions(xs)) != 3 {
		t.Errorf("IndexProfileRevisions() = %v", IndexProfileRevisions(xs))
	}
	if err := ValidateProfileRevisions(ActiveProfileRevisions(xs)); err != nil {
		t.Errorf("ValidateProfileRevisions() = %v", err)
	}
}

func TestProfileRevisionNormalizeClamps(t *testing.T) {
	x := SampleProfileRevision("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileSnapshotRoundTrip(t *testing.T) {
	x := SampleProfileSnapshot("profilesnapshot-1", when)
	got, err := ParseProfileSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileSnapshot() = %v", err)
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

func TestProfileSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleProfileSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileSnapshot(f); err == nil {
		t.Error("ParseProfileSnapshot() accepted a record with no id")
	}
	f = SampleProfileSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseProfileSnapshot(f); err == nil {
		t.Error("ParseProfileSnapshot() accepted a non-numeric priority")
	}
}

func TestProfileSnapshotCollections(t *testing.T) {
	xs := []ProfileSnapshot{SampleProfileSnapshot("a", when), SampleProfileSnapshot("bb", when), SampleProfileSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileSnapshots(xs)); got != 2 {
		t.Errorf("ActiveProfileSnapshots() kept %d, want 2", got)
	}
	if got, want := SumProfileSnapshots(xs), 1+3; got != want {
		t.Errorf("SumProfileSnapshots() = %d, want %d", got, want)
	}
	SortProfileSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexProfileSnapshots(xs)) != 3 {
		t.Errorf("IndexProfileSnapshots() = %v", IndexProfileSnapshots(xs))
	}
	if err := ValidateProfileSnapshots(ActiveProfileSnapshots(xs)); err != nil {
		t.Errorf("ValidateProfileSnapshots() = %v", err)
	}
}

func TestProfileSnapshotNormalizeClamps(t *testing.T) {
	x := SampleProfileSnapshot("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileBatchRoundTrip(t *testing.T) {
	x := SampleProfileBatch("profilebatch-1", when)
	got, err := ParseProfileBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileBatch() = %v", err)
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

func TestProfileBatchParseRejectsBadInput(t *testing.T) {
	f := SampleProfileBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileBatch(f); err == nil {
		t.Error("ParseProfileBatch() accepted a record with no id")
	}
	f = SampleProfileBatch("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseProfileBatch(f); err == nil {
		t.Error("ParseProfileBatch() accepted a non-numeric weight")
	}
}

func TestProfileBatchCollections(t *testing.T) {
	xs := []ProfileBatch{SampleProfileBatch("a", when), SampleProfileBatch("bb", when), SampleProfileBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileBatchs(xs)); got != 2 {
		t.Errorf("ActiveProfileBatchs() kept %d, want 2", got)
	}
	if got, want := SumProfileBatchs(xs), 1+3; got != want {
		t.Errorf("SumProfileBatchs() = %d, want %d", got, want)
	}
	SortProfileBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileBatchs() put %q first", xs[0].ID)
	}
	if len(IndexProfileBatchs(xs)) != 3 {
		t.Errorf("IndexProfileBatchs() = %v", IndexProfileBatchs(xs))
	}
	if err := ValidateProfileBatchs(ActiveProfileBatchs(xs)); err != nil {
		t.Errorf("ValidateProfileBatchs() = %v", err)
	}
}

func TestProfileBatchNormalizeClamps(t *testing.T) {
	x := SampleProfileBatch("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileRequestRoundTrip(t *testing.T) {
	x := SampleProfileRequest("profilerequest-1", when)
	got, err := ParseProfileRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileRequest() = %v", err)
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

func TestProfileRequestParseRejectsBadInput(t *testing.T) {
	f := SampleProfileRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileRequest(f); err == nil {
		t.Error("ParseProfileRequest() accepted a record with no id")
	}
	f = SampleProfileRequest("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseProfileRequest(f); err == nil {
		t.Error("ParseProfileRequest() accepted a non-numeric offset")
	}
}

func TestProfileRequestCollections(t *testing.T) {
	xs := []ProfileRequest{SampleProfileRequest("a", when), SampleProfileRequest("bb", when), SampleProfileRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileRequests(xs)); got != 2 {
		t.Errorf("ActiveProfileRequests() kept %d, want 2", got)
	}
	if got, want := SumProfileRequests(xs), 1+3; got != want {
		t.Errorf("SumProfileRequests() = %d, want %d", got, want)
	}
	SortProfileRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileRequests() put %q first", xs[0].ID)
	}
	if len(IndexProfileRequests(xs)) != 3 {
		t.Errorf("IndexProfileRequests() = %v", IndexProfileRequests(xs))
	}
	if err := ValidateProfileRequests(ActiveProfileRequests(xs)); err != nil {
		t.Errorf("ValidateProfileRequests() = %v", err)
	}
}

func TestProfileRequestNormalizeClamps(t *testing.T) {
	x := SampleProfileRequest("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileSummaryRoundTrip(t *testing.T) {
	x := SampleProfileSummary("profilesummary-1", when)
	got, err := ParseProfileSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileSummary() = %v", err)
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

func TestProfileSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleProfileSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileSummary(f); err == nil {
		t.Error("ParseProfileSummary() accepted a record with no id")
	}
	f = SampleProfileSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseProfileSummary(f); err == nil {
		t.Error("ParseProfileSummary() accepted a non-numeric capacity")
	}
}

func TestProfileSummaryCollections(t *testing.T) {
	xs := []ProfileSummary{SampleProfileSummary("a", when), SampleProfileSummary("bb", when), SampleProfileSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileSummarys(xs)); got != 2 {
		t.Errorf("ActiveProfileSummarys() kept %d, want 2", got)
	}
	if got, want := SumProfileSummarys(xs), 1+3; got != want {
		t.Errorf("SumProfileSummarys() = %d, want %d", got, want)
	}
	SortProfileSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileSummarys() put %q first", xs[0].ID)
	}
	if len(IndexProfileSummarys(xs)) != 3 {
		t.Errorf("IndexProfileSummarys() = %v", IndexProfileSummarys(xs))
	}
	if err := ValidateProfileSummarys(ActiveProfileSummarys(xs)); err != nil {
		t.Errorf("ValidateProfileSummarys() = %v", err)
	}
}

func TestProfileSummaryNormalizeClamps(t *testing.T) {
	x := SampleProfileSummary("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfilePolicyRoundTrip(t *testing.T) {
	x := SampleProfilePolicy("profilepolicy-1", when)
	got, err := ParseProfilePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfilePolicy() = %v", err)
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

func TestProfilePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleProfilePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfilePolicy(f); err == nil {
		t.Error("ParseProfilePolicy() accepted a record with no id")
	}
	f = SampleProfilePolicy("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseProfilePolicy(f); err == nil {
		t.Error("ParseProfilePolicy() accepted a non-numeric offset")
	}
}

func TestProfilePolicyCollections(t *testing.T) {
	xs := []ProfilePolicy{SampleProfilePolicy("a", when), SampleProfilePolicy("bb", when), SampleProfilePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfilePolicys(xs)); got != 2 {
		t.Errorf("ActiveProfilePolicys() kept %d, want 2", got)
	}
	if got, want := SumProfilePolicys(xs), 1+3; got != want {
		t.Errorf("SumProfilePolicys() = %d, want %d", got, want)
	}
	SortProfilePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfilePolicys() put %q first", xs[0].ID)
	}
	if len(IndexProfilePolicys(xs)) != 3 {
		t.Errorf("IndexProfilePolicys() = %v", IndexProfilePolicys(xs))
	}
	if err := ValidateProfilePolicys(ActiveProfilePolicys(xs)); err != nil {
		t.Errorf("ValidateProfilePolicys() = %v", err)
	}
}

func TestProfilePolicyNormalizeClamps(t *testing.T) {
	x := SampleProfilePolicy("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileLinkRoundTrip(t *testing.T) {
	x := SampleProfileLink("profilelink-1", when)
	got, err := ParseProfileLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileLink() = %v", err)
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

func TestProfileLinkParseRejectsBadInput(t *testing.T) {
	f := SampleProfileLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileLink(f); err == nil {
		t.Error("ParseProfileLink() accepted a record with no id")
	}
	f = SampleProfileLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseProfileLink(f); err == nil {
		t.Error("ParseProfileLink() accepted a non-numeric retained")
	}
}

func TestProfileLinkCollections(t *testing.T) {
	xs := []ProfileLink{SampleProfileLink("a", when), SampleProfileLink("bb", when), SampleProfileLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileLinks(xs)); got != 2 {
		t.Errorf("ActiveProfileLinks() kept %d, want 2", got)
	}
	if got, want := SumProfileLinks(xs), 1+3; got != want {
		t.Errorf("SumProfileLinks() = %d, want %d", got, want)
	}
	SortProfileLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileLinks() put %q first", xs[0].ID)
	}
	if len(IndexProfileLinks(xs)) != 3 {
		t.Errorf("IndexProfileLinks() = %v", IndexProfileLinks(xs))
	}
	if err := ValidateProfileLinks(ActiveProfileLinks(xs)); err != nil {
		t.Errorf("ValidateProfileLinks() = %v", err)
	}
}

func TestProfileLinkNormalizeClamps(t *testing.T) {
	x := SampleProfileLink("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileNoteRoundTrip(t *testing.T) {
	x := SampleProfileNote("profilenote-1", when)
	got, err := ParseProfileNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileNote() = %v", err)
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

func TestProfileNoteParseRejectsBadInput(t *testing.T) {
	f := SampleProfileNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileNote(f); err == nil {
		t.Error("ParseProfileNote() accepted a record with no id")
	}
	f = SampleProfileNote("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseProfileNote(f); err == nil {
		t.Error("ParseProfileNote() accepted a non-numeric attempts")
	}
}

func TestProfileNoteCollections(t *testing.T) {
	xs := []ProfileNote{SampleProfileNote("a", when), SampleProfileNote("bb", when), SampleProfileNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileNotes(xs)); got != 2 {
		t.Errorf("ActiveProfileNotes() kept %d, want 2", got)
	}
	if got, want := SumProfileNotes(xs), 1+3; got != want {
		t.Errorf("SumProfileNotes() = %d, want %d", got, want)
	}
	SortProfileNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileNotes() put %q first", xs[0].ID)
	}
	if len(IndexProfileNotes(xs)) != 3 {
		t.Errorf("IndexProfileNotes() = %v", IndexProfileNotes(xs))
	}
	if err := ValidateProfileNotes(ActiveProfileNotes(xs)); err != nil {
		t.Errorf("ValidateProfileNotes() = %v", err)
	}
}

func TestProfileNoteNormalizeClamps(t *testing.T) {
	x := SampleProfileNote("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileWindowRoundTrip(t *testing.T) {
	x := SampleProfileWindow("profilewindow-1", when)
	got, err := ParseProfileWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileWindow() = %v", err)
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

func TestProfileWindowParseRejectsBadInput(t *testing.T) {
	f := SampleProfileWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileWindow(f); err == nil {
		t.Error("ParseProfileWindow() accepted a record with no id")
	}
	f = SampleProfileWindow("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseProfileWindow(f); err == nil {
		t.Error("ParseProfileWindow() accepted a non-numeric offset")
	}
}

func TestProfileWindowCollections(t *testing.T) {
	xs := []ProfileWindow{SampleProfileWindow("a", when), SampleProfileWindow("bb", when), SampleProfileWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileWindows(xs)); got != 2 {
		t.Errorf("ActiveProfileWindows() kept %d, want 2", got)
	}
	if got, want := SumProfileWindows(xs), 1+3; got != want {
		t.Errorf("SumProfileWindows() = %d, want %d", got, want)
	}
	SortProfileWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileWindows() put %q first", xs[0].ID)
	}
	if len(IndexProfileWindows(xs)) != 3 {
		t.Errorf("IndexProfileWindows() = %v", IndexProfileWindows(xs))
	}
	if err := ValidateProfileWindows(ActiveProfileWindows(xs)); err != nil {
		t.Errorf("ValidateProfileWindows() = %v", err)
	}
}

func TestProfileWindowNormalizeClamps(t *testing.T) {
	x := SampleProfileWindow("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileQuotaRoundTrip(t *testing.T) {
	x := SampleProfileQuota("profilequota-1", when)
	got, err := ParseProfileQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileQuota() = %v", err)
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

func TestProfileQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleProfileQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileQuota(f); err == nil {
		t.Error("ParseProfileQuota() accepted a record with no id")
	}
	f = SampleProfileQuota("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseProfileQuota(f); err == nil {
		t.Error("ParseProfileQuota() accepted a non-numeric retained")
	}
}

func TestProfileQuotaCollections(t *testing.T) {
	xs := []ProfileQuota{SampleProfileQuota("a", when), SampleProfileQuota("bb", when), SampleProfileQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileQuotas(xs)); got != 2 {
		t.Errorf("ActiveProfileQuotas() kept %d, want 2", got)
	}
	if got, want := SumProfileQuotas(xs), 1+3; got != want {
		t.Errorf("SumProfileQuotas() = %d, want %d", got, want)
	}
	SortProfileQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileQuotas() put %q first", xs[0].ID)
	}
	if len(IndexProfileQuotas(xs)) != 3 {
		t.Errorf("IndexProfileQuotas() = %v", IndexProfileQuotas(xs))
	}
	if err := ValidateProfileQuotas(ActiveProfileQuotas(xs)); err != nil {
		t.Errorf("ValidateProfileQuotas() = %v", err)
	}
}

func TestProfileQuotaNormalizeClamps(t *testing.T) {
	x := SampleProfileQuota("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileBucketRoundTrip(t *testing.T) {
	x := SampleProfileBucket("profilebucket-1", when)
	got, err := ParseProfileBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileBucket() = %v", err)
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

func TestProfileBucketParseRejectsBadInput(t *testing.T) {
	f := SampleProfileBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileBucket(f); err == nil {
		t.Error("ParseProfileBucket() accepted a record with no id")
	}
	f = SampleProfileBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseProfileBucket(f); err == nil {
		t.Error("ParseProfileBucket() accepted a non-numeric retained")
	}
}

func TestProfileBucketCollections(t *testing.T) {
	xs := []ProfileBucket{SampleProfileBucket("a", when), SampleProfileBucket("bb", when), SampleProfileBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileBuckets(xs)); got != 2 {
		t.Errorf("ActiveProfileBuckets() kept %d, want 2", got)
	}
	if got, want := SumProfileBuckets(xs), 1+3; got != want {
		t.Errorf("SumProfileBuckets() = %d, want %d", got, want)
	}
	SortProfileBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileBuckets() put %q first", xs[0].ID)
	}
	if len(IndexProfileBuckets(xs)) != 3 {
		t.Errorf("IndexProfileBuckets() = %v", IndexProfileBuckets(xs))
	}
	if err := ValidateProfileBuckets(ActiveProfileBuckets(xs)); err != nil {
		t.Errorf("ValidateProfileBuckets() = %v", err)
	}
}

func TestProfileBucketNormalizeClamps(t *testing.T) {
	x := SampleProfileBucket("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileLaneRoundTrip(t *testing.T) {
	x := SampleProfileLane("profilelane-1", when)
	got, err := ParseProfileLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileLane() = %v", err)
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

func TestProfileLaneParseRejectsBadInput(t *testing.T) {
	f := SampleProfileLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileLane(f); err == nil {
		t.Error("ParseProfileLane() accepted a record with no id")
	}
	f = SampleProfileLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseProfileLane(f); err == nil {
		t.Error("ParseProfileLane() accepted a non-numeric attempts")
	}
}

func TestProfileLaneCollections(t *testing.T) {
	xs := []ProfileLane{SampleProfileLane("a", when), SampleProfileLane("bb", when), SampleProfileLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileLanes(xs)); got != 2 {
		t.Errorf("ActiveProfileLanes() kept %d, want 2", got)
	}
	if got, want := SumProfileLanes(xs), 1+3; got != want {
		t.Errorf("SumProfileLanes() = %d, want %d", got, want)
	}
	SortProfileLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileLanes() put %q first", xs[0].ID)
	}
	if len(IndexProfileLanes(xs)) != 3 {
		t.Errorf("IndexProfileLanes() = %v", IndexProfileLanes(xs))
	}
	if err := ValidateProfileLanes(ActiveProfileLanes(xs)); err != nil {
		t.Errorf("ValidateProfileLanes() = %v", err)
	}
}

func TestProfileLaneNormalizeClamps(t *testing.T) {
	x := SampleProfileLane("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileSlotRoundTrip(t *testing.T) {
	x := SampleProfileSlot("profileslot-1", when)
	got, err := ParseProfileSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileSlot() = %v", err)
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

func TestProfileSlotParseRejectsBadInput(t *testing.T) {
	f := SampleProfileSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileSlot(f); err == nil {
		t.Error("ParseProfileSlot() accepted a record with no id")
	}
	f = SampleProfileSlot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseProfileSlot(f); err == nil {
		t.Error("ParseProfileSlot() accepted a non-numeric retained")
	}
}

func TestProfileSlotCollections(t *testing.T) {
	xs := []ProfileSlot{SampleProfileSlot("a", when), SampleProfileSlot("bb", when), SampleProfileSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileSlots(xs)); got != 2 {
		t.Errorf("ActiveProfileSlots() kept %d, want 2", got)
	}
	if got, want := SumProfileSlots(xs), 1+3; got != want {
		t.Errorf("SumProfileSlots() = %d, want %d", got, want)
	}
	SortProfileSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileSlots() put %q first", xs[0].ID)
	}
	if len(IndexProfileSlots(xs)) != 3 {
		t.Errorf("IndexProfileSlots() = %v", IndexProfileSlots(xs))
	}
	if err := ValidateProfileSlots(ActiveProfileSlots(xs)); err != nil {
		t.Errorf("ValidateProfileSlots() = %v", err)
	}
}

func TestProfileSlotNormalizeClamps(t *testing.T) {
	x := SampleProfileSlot("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileTagRoundTrip(t *testing.T) {
	x := SampleProfileTag("profiletag-1", when)
	got, err := ParseProfileTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileTag() = %v", err)
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

func TestProfileTagParseRejectsBadInput(t *testing.T) {
	f := SampleProfileTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileTag(f); err == nil {
		t.Error("ParseProfileTag() accepted a record with no id")
	}
	f = SampleProfileTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseProfileTag(f); err == nil {
		t.Error("ParseProfileTag() accepted a non-numeric count")
	}
}

func TestProfileTagCollections(t *testing.T) {
	xs := []ProfileTag{SampleProfileTag("a", when), SampleProfileTag("bb", when), SampleProfileTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileTags(xs)); got != 2 {
		t.Errorf("ActiveProfileTags() kept %d, want 2", got)
	}
	if got, want := SumProfileTags(xs), 1+3; got != want {
		t.Errorf("SumProfileTags() = %d, want %d", got, want)
	}
	SortProfileTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileTags() put %q first", xs[0].ID)
	}
	if len(IndexProfileTags(xs)) != 3 {
		t.Errorf("IndexProfileTags() = %v", IndexProfileTags(xs))
	}
	if err := ValidateProfileTags(ActiveProfileTags(xs)); err != nil {
		t.Errorf("ValidateProfileTags() = %v", err)
	}
}

func TestProfileTagNormalizeClamps(t *testing.T) {
	x := SampleProfileTag("a", when)
	x.Source = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileRuleRoundTrip(t *testing.T) {
	x := SampleProfileRule("profilerule-1", when)
	got, err := ParseProfileRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileRule() = %v", err)
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

func TestProfileRuleParseRejectsBadInput(t *testing.T) {
	f := SampleProfileRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileRule(f); err == nil {
		t.Error("ParseProfileRule() accepted a record with no id")
	}
	f = SampleProfileRule("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseProfileRule(f); err == nil {
		t.Error("ParseProfileRule() accepted a non-numeric weight")
	}
}

func TestProfileRuleCollections(t *testing.T) {
	xs := []ProfileRule{SampleProfileRule("a", when), SampleProfileRule("bb", when), SampleProfileRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileRules(xs)); got != 2 {
		t.Errorf("ActiveProfileRules() kept %d, want 2", got)
	}
	if got, want := SumProfileRules(xs), 1+3; got != want {
		t.Errorf("SumProfileRules() = %d, want %d", got, want)
	}
	SortProfileRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileRules() put %q first", xs[0].ID)
	}
	if len(IndexProfileRules(xs)) != 3 {
		t.Errorf("IndexProfileRules() = %v", IndexProfileRules(xs))
	}
	if err := ValidateProfileRules(ActiveProfileRules(xs)); err != nil {
		t.Errorf("ValidateProfileRules() = %v", err)
	}
}

func TestProfileRuleNormalizeClamps(t *testing.T) {
	x := SampleProfileRule("a", when)
	x.Group = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileHoldRoundTrip(t *testing.T) {
	x := SampleProfileHold("profilehold-1", when)
	got, err := ParseProfileHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileHold() = %v", err)
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

func TestProfileHoldParseRejectsBadInput(t *testing.T) {
	f := SampleProfileHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileHold(f); err == nil {
		t.Error("ParseProfileHold() accepted a record with no id")
	}
	f = SampleProfileHold("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseProfileHold(f); err == nil {
		t.Error("ParseProfileHold() accepted a non-numeric retained")
	}
}

func TestProfileHoldCollections(t *testing.T) {
	xs := []ProfileHold{SampleProfileHold("a", when), SampleProfileHold("bb", when), SampleProfileHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileHolds(xs)); got != 2 {
		t.Errorf("ActiveProfileHolds() kept %d, want 2", got)
	}
	if got, want := SumProfileHolds(xs), 1+3; got != want {
		t.Errorf("SumProfileHolds() = %d, want %d", got, want)
	}
	SortProfileHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileHolds() put %q first", xs[0].ID)
	}
	if len(IndexProfileHolds(xs)) != 3 {
		t.Errorf("IndexProfileHolds() = %v", IndexProfileHolds(xs))
	}
	if err := ValidateProfileHolds(ActiveProfileHolds(xs)); err != nil {
		t.Errorf("ValidateProfileHolds() = %v", err)
	}
}

func TestProfileHoldNormalizeClamps(t *testing.T) {
	x := SampleProfileHold("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileClaimRoundTrip(t *testing.T) {
	x := SampleProfileClaim("profileclaim-1", when)
	got, err := ParseProfileClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileClaim() = %v", err)
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

func TestProfileClaimParseRejectsBadInput(t *testing.T) {
	f := SampleProfileClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileClaim(f); err == nil {
		t.Error("ParseProfileClaim() accepted a record with no id")
	}
	f = SampleProfileClaim("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseProfileClaim(f); err == nil {
		t.Error("ParseProfileClaim() accepted a non-numeric version")
	}
}

func TestProfileClaimCollections(t *testing.T) {
	xs := []ProfileClaim{SampleProfileClaim("a", when), SampleProfileClaim("bb", when), SampleProfileClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileClaims(xs)); got != 2 {
		t.Errorf("ActiveProfileClaims() kept %d, want 2", got)
	}
	if got, want := SumProfileClaims(xs), 1+3; got != want {
		t.Errorf("SumProfileClaims() = %d, want %d", got, want)
	}
	SortProfileClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileClaims() put %q first", xs[0].ID)
	}
	if len(IndexProfileClaims(xs)) != 3 {
		t.Errorf("IndexProfileClaims() = %v", IndexProfileClaims(xs))
	}
	if err := ValidateProfileClaims(ActiveProfileClaims(xs)); err != nil {
		t.Errorf("ValidateProfileClaims() = %v", err)
	}
}

func TestProfileClaimNormalizeClamps(t *testing.T) {
	x := SampleProfileClaim("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestProfileCursorRoundTrip(t *testing.T) {
	x := SampleProfileCursor("profilecursor-1", when)
	got, err := ParseProfileCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseProfileCursor() = %v", err)
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

func TestProfileCursorParseRejectsBadInput(t *testing.T) {
	f := SampleProfileCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseProfileCursor(f); err == nil {
		t.Error("ParseProfileCursor() accepted a record with no id")
	}
	f = SampleProfileCursor("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseProfileCursor(f); err == nil {
		t.Error("ParseProfileCursor() accepted a non-numeric capacity")
	}
}

func TestProfileCursorCollections(t *testing.T) {
	xs := []ProfileCursor{SampleProfileCursor("a", when), SampleProfileCursor("bb", when), SampleProfileCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveProfileCursors(xs)); got != 2 {
		t.Errorf("ActiveProfileCursors() kept %d, want 2", got)
	}
	if got, want := SumProfileCursors(xs), 1+3; got != want {
		t.Errorf("SumProfileCursors() = %d, want %d", got, want)
	}
	SortProfileCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortProfileCursors() put %q first", xs[0].ID)
	}
	if len(IndexProfileCursors(xs)) != 3 {
		t.Errorf("IndexProfileCursors() = %v", IndexProfileCursors(xs))
	}
	if err := ValidateProfileCursors(ActiveProfileCursors(xs)); err != nil {
		t.Errorf("ValidateProfileCursors() = %v", err)
	}
}

func TestProfileCursorNormalizeClamps(t *testing.T) {
	x := SampleProfileCursor("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
