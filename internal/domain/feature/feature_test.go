package feature

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
		if !strings.HasPrefix(Prefixes[i], "feature/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleFeatureRecord("ok", when).Fields()); err != nil {
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

func TestFeatureRecordRoundTrip(t *testing.T) {
	x := SampleFeatureRecord("featurerecord-1", when)
	got, err := ParseFeatureRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureRecord() = %v", err)
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

func TestFeatureRecordParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureRecord(f); err == nil {
		t.Error("ParseFeatureRecord() accepted a record with no id")
	}
	f = SampleFeatureRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseFeatureRecord(f); err == nil {
		t.Error("ParseFeatureRecord() accepted a non-numeric capacity")
	}
}

func TestFeatureRecordCollections(t *testing.T) {
	xs := []FeatureRecord{SampleFeatureRecord("a", when), SampleFeatureRecord("bb", when), SampleFeatureRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureRecords(xs)); got != 2 {
		t.Errorf("ActiveFeatureRecords() kept %d, want 2", got)
	}
	if got, want := SumFeatureRecords(xs), 1+3; got != want {
		t.Errorf("SumFeatureRecords() = %d, want %d", got, want)
	}
	SortFeatureRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureRecords() put %q first", xs[0].ID)
	}
	if len(IndexFeatureRecords(xs)) != 3 {
		t.Errorf("IndexFeatureRecords() = %v", IndexFeatureRecords(xs))
	}
	if err := ValidateFeatureRecords(ActiveFeatureRecords(xs)); err != nil {
		t.Errorf("ValidateFeatureRecords() = %v", err)
	}
}

func TestFeatureRecordNormalizeClamps(t *testing.T) {
	x := SampleFeatureRecord("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureEntryRoundTrip(t *testing.T) {
	x := SampleFeatureEntry("featureentry-1", when)
	got, err := ParseFeatureEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureEntry() = %v", err)
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

func TestFeatureEntryParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureEntry(f); err == nil {
		t.Error("ParseFeatureEntry() accepted a record with no id")
	}
	f = SampleFeatureEntry("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseFeatureEntry(f); err == nil {
		t.Error("ParseFeatureEntry() accepted a non-numeric priority")
	}
}

func TestFeatureEntryCollections(t *testing.T) {
	xs := []FeatureEntry{SampleFeatureEntry("a", when), SampleFeatureEntry("bb", when), SampleFeatureEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureEntrys(xs)); got != 2 {
		t.Errorf("ActiveFeatureEntrys() kept %d, want 2", got)
	}
	if got, want := SumFeatureEntrys(xs), 1+3; got != want {
		t.Errorf("SumFeatureEntrys() = %d, want %d", got, want)
	}
	SortFeatureEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureEntrys() put %q first", xs[0].ID)
	}
	if len(IndexFeatureEntrys(xs)) != 3 {
		t.Errorf("IndexFeatureEntrys() = %v", IndexFeatureEntrys(xs))
	}
	if err := ValidateFeatureEntrys(ActiveFeatureEntrys(xs)); err != nil {
		t.Errorf("ValidateFeatureEntrys() = %v", err)
	}
}

func TestFeatureEntryNormalizeClamps(t *testing.T) {
	x := SampleFeatureEntry("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureItemRoundTrip(t *testing.T) {
	x := SampleFeatureItem("featureitem-1", when)
	got, err := ParseFeatureItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureItem() = %v", err)
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

func TestFeatureItemParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureItem(f); err == nil {
		t.Error("ParseFeatureItem() accepted a record with no id")
	}
	f = SampleFeatureItem("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseFeatureItem(f); err == nil {
		t.Error("ParseFeatureItem() accepted a non-numeric retained")
	}
}

func TestFeatureItemCollections(t *testing.T) {
	xs := []FeatureItem{SampleFeatureItem("a", when), SampleFeatureItem("bb", when), SampleFeatureItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureItems(xs)); got != 2 {
		t.Errorf("ActiveFeatureItems() kept %d, want 2", got)
	}
	if got, want := SumFeatureItems(xs), 1+3; got != want {
		t.Errorf("SumFeatureItems() = %d, want %d", got, want)
	}
	SortFeatureItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureItems() put %q first", xs[0].ID)
	}
	if len(IndexFeatureItems(xs)) != 3 {
		t.Errorf("IndexFeatureItems() = %v", IndexFeatureItems(xs))
	}
	if err := ValidateFeatureItems(ActiveFeatureItems(xs)); err != nil {
		t.Errorf("ValidateFeatureItems() = %v", err)
	}
}

func TestFeatureItemNormalizeClamps(t *testing.T) {
	x := SampleFeatureItem("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureDraftRoundTrip(t *testing.T) {
	x := SampleFeatureDraft("featuredraft-1", when)
	got, err := ParseFeatureDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureDraft() = %v", err)
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

func TestFeatureDraftParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureDraft(f); err == nil {
		t.Error("ParseFeatureDraft() accepted a record with no id")
	}
	f = SampleFeatureDraft("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseFeatureDraft(f); err == nil {
		t.Error("ParseFeatureDraft() accepted a non-numeric retained")
	}
}

func TestFeatureDraftCollections(t *testing.T) {
	xs := []FeatureDraft{SampleFeatureDraft("a", when), SampleFeatureDraft("bb", when), SampleFeatureDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureDrafts(xs)); got != 2 {
		t.Errorf("ActiveFeatureDrafts() kept %d, want 2", got)
	}
	if got, want := SumFeatureDrafts(xs), 1+3; got != want {
		t.Errorf("SumFeatureDrafts() = %d, want %d", got, want)
	}
	SortFeatureDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureDrafts() put %q first", xs[0].ID)
	}
	if len(IndexFeatureDrafts(xs)) != 3 {
		t.Errorf("IndexFeatureDrafts() = %v", IndexFeatureDrafts(xs))
	}
	if err := ValidateFeatureDrafts(ActiveFeatureDrafts(xs)); err != nil {
		t.Errorf("ValidateFeatureDrafts() = %v", err)
	}
}

func TestFeatureDraftNormalizeClamps(t *testing.T) {
	x := SampleFeatureDraft("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureRevisionRoundTrip(t *testing.T) {
	x := SampleFeatureRevision("featurerevision-1", when)
	got, err := ParseFeatureRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureRevision() = %v", err)
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

func TestFeatureRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureRevision(f); err == nil {
		t.Error("ParseFeatureRevision() accepted a record with no id")
	}
	f = SampleFeatureRevision("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseFeatureRevision(f); err == nil {
		t.Error("ParseFeatureRevision() accepted a non-numeric capacity")
	}
}

func TestFeatureRevisionCollections(t *testing.T) {
	xs := []FeatureRevision{SampleFeatureRevision("a", when), SampleFeatureRevision("bb", when), SampleFeatureRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureRevisions(xs)); got != 2 {
		t.Errorf("ActiveFeatureRevisions() kept %d, want 2", got)
	}
	if got, want := SumFeatureRevisions(xs), 1+3; got != want {
		t.Errorf("SumFeatureRevisions() = %d, want %d", got, want)
	}
	SortFeatureRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureRevisions() put %q first", xs[0].ID)
	}
	if len(IndexFeatureRevisions(xs)) != 3 {
		t.Errorf("IndexFeatureRevisions() = %v", IndexFeatureRevisions(xs))
	}
	if err := ValidateFeatureRevisions(ActiveFeatureRevisions(xs)); err != nil {
		t.Errorf("ValidateFeatureRevisions() = %v", err)
	}
}

func TestFeatureRevisionNormalizeClamps(t *testing.T) {
	x := SampleFeatureRevision("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureSnapshotRoundTrip(t *testing.T) {
	x := SampleFeatureSnapshot("featuresnapshot-1", when)
	got, err := ParseFeatureSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureSnapshot() = %v", err)
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

func TestFeatureSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureSnapshot(f); err == nil {
		t.Error("ParseFeatureSnapshot() accepted a record with no id")
	}
	f = SampleFeatureSnapshot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseFeatureSnapshot(f); err == nil {
		t.Error("ParseFeatureSnapshot() accepted a non-numeric retained")
	}
}

func TestFeatureSnapshotCollections(t *testing.T) {
	xs := []FeatureSnapshot{SampleFeatureSnapshot("a", when), SampleFeatureSnapshot("bb", when), SampleFeatureSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureSnapshots(xs)); got != 2 {
		t.Errorf("ActiveFeatureSnapshots() kept %d, want 2", got)
	}
	if got, want := SumFeatureSnapshots(xs), 1+3; got != want {
		t.Errorf("SumFeatureSnapshots() = %d, want %d", got, want)
	}
	SortFeatureSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexFeatureSnapshots(xs)) != 3 {
		t.Errorf("IndexFeatureSnapshots() = %v", IndexFeatureSnapshots(xs))
	}
	if err := ValidateFeatureSnapshots(ActiveFeatureSnapshots(xs)); err != nil {
		t.Errorf("ValidateFeatureSnapshots() = %v", err)
	}
}

func TestFeatureSnapshotNormalizeClamps(t *testing.T) {
	x := SampleFeatureSnapshot("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureBatchRoundTrip(t *testing.T) {
	x := SampleFeatureBatch("featurebatch-1", when)
	got, err := ParseFeatureBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureBatch() = %v", err)
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

func TestFeatureBatchParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureBatch(f); err == nil {
		t.Error("ParseFeatureBatch() accepted a record with no id")
	}
	f = SampleFeatureBatch("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseFeatureBatch(f); err == nil {
		t.Error("ParseFeatureBatch() accepted a non-numeric offset")
	}
}

func TestFeatureBatchCollections(t *testing.T) {
	xs := []FeatureBatch{SampleFeatureBatch("a", when), SampleFeatureBatch("bb", when), SampleFeatureBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureBatchs(xs)); got != 2 {
		t.Errorf("ActiveFeatureBatchs() kept %d, want 2", got)
	}
	if got, want := SumFeatureBatchs(xs), 1+3; got != want {
		t.Errorf("SumFeatureBatchs() = %d, want %d", got, want)
	}
	SortFeatureBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureBatchs() put %q first", xs[0].ID)
	}
	if len(IndexFeatureBatchs(xs)) != 3 {
		t.Errorf("IndexFeatureBatchs() = %v", IndexFeatureBatchs(xs))
	}
	if err := ValidateFeatureBatchs(ActiveFeatureBatchs(xs)); err != nil {
		t.Errorf("ValidateFeatureBatchs() = %v", err)
	}
}

func TestFeatureBatchNormalizeClamps(t *testing.T) {
	x := SampleFeatureBatch("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureRequestRoundTrip(t *testing.T) {
	x := SampleFeatureRequest("featurerequest-1", when)
	got, err := ParseFeatureRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureRequest() = %v", err)
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

func TestFeatureRequestParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureRequest(f); err == nil {
		t.Error("ParseFeatureRequest() accepted a record with no id")
	}
	f = SampleFeatureRequest("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseFeatureRequest(f); err == nil {
		t.Error("ParseFeatureRequest() accepted a non-numeric offset")
	}
}

func TestFeatureRequestCollections(t *testing.T) {
	xs := []FeatureRequest{SampleFeatureRequest("a", when), SampleFeatureRequest("bb", when), SampleFeatureRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureRequests(xs)); got != 2 {
		t.Errorf("ActiveFeatureRequests() kept %d, want 2", got)
	}
	if got, want := SumFeatureRequests(xs), 1+3; got != want {
		t.Errorf("SumFeatureRequests() = %d, want %d", got, want)
	}
	SortFeatureRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureRequests() put %q first", xs[0].ID)
	}
	if len(IndexFeatureRequests(xs)) != 3 {
		t.Errorf("IndexFeatureRequests() = %v", IndexFeatureRequests(xs))
	}
	if err := ValidateFeatureRequests(ActiveFeatureRequests(xs)); err != nil {
		t.Errorf("ValidateFeatureRequests() = %v", err)
	}
}

func TestFeatureRequestNormalizeClamps(t *testing.T) {
	x := SampleFeatureRequest("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureSummaryRoundTrip(t *testing.T) {
	x := SampleFeatureSummary("featuresummary-1", when)
	got, err := ParseFeatureSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureSummary() = %v", err)
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

func TestFeatureSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureSummary(f); err == nil {
		t.Error("ParseFeatureSummary() accepted a record with no id")
	}
	f = SampleFeatureSummary("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseFeatureSummary(f); err == nil {
		t.Error("ParseFeatureSummary() accepted a non-numeric version")
	}
}

func TestFeatureSummaryCollections(t *testing.T) {
	xs := []FeatureSummary{SampleFeatureSummary("a", when), SampleFeatureSummary("bb", when), SampleFeatureSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureSummarys(xs)); got != 2 {
		t.Errorf("ActiveFeatureSummarys() kept %d, want 2", got)
	}
	if got, want := SumFeatureSummarys(xs), 1+3; got != want {
		t.Errorf("SumFeatureSummarys() = %d, want %d", got, want)
	}
	SortFeatureSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureSummarys() put %q first", xs[0].ID)
	}
	if len(IndexFeatureSummarys(xs)) != 3 {
		t.Errorf("IndexFeatureSummarys() = %v", IndexFeatureSummarys(xs))
	}
	if err := ValidateFeatureSummarys(ActiveFeatureSummarys(xs)); err != nil {
		t.Errorf("ValidateFeatureSummarys() = %v", err)
	}
}

func TestFeatureSummaryNormalizeClamps(t *testing.T) {
	x := SampleFeatureSummary("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeaturePolicyRoundTrip(t *testing.T) {
	x := SampleFeaturePolicy("featurepolicy-1", when)
	got, err := ParseFeaturePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeaturePolicy() = %v", err)
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

func TestFeaturePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleFeaturePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeaturePolicy(f); err == nil {
		t.Error("ParseFeaturePolicy() accepted a record with no id")
	}
	f = SampleFeaturePolicy("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseFeaturePolicy(f); err == nil {
		t.Error("ParseFeaturePolicy() accepted a non-numeric attempts")
	}
}

func TestFeaturePolicyCollections(t *testing.T) {
	xs := []FeaturePolicy{SampleFeaturePolicy("a", when), SampleFeaturePolicy("bb", when), SampleFeaturePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeaturePolicys(xs)); got != 2 {
		t.Errorf("ActiveFeaturePolicys() kept %d, want 2", got)
	}
	if got, want := SumFeaturePolicys(xs), 1+3; got != want {
		t.Errorf("SumFeaturePolicys() = %d, want %d", got, want)
	}
	SortFeaturePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeaturePolicys() put %q first", xs[0].ID)
	}
	if len(IndexFeaturePolicys(xs)) != 3 {
		t.Errorf("IndexFeaturePolicys() = %v", IndexFeaturePolicys(xs))
	}
	if err := ValidateFeaturePolicys(ActiveFeaturePolicys(xs)); err != nil {
		t.Errorf("ValidateFeaturePolicys() = %v", err)
	}
}

func TestFeaturePolicyNormalizeClamps(t *testing.T) {
	x := SampleFeaturePolicy("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureLinkRoundTrip(t *testing.T) {
	x := SampleFeatureLink("featurelink-1", when)
	got, err := ParseFeatureLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureLink() = %v", err)
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

func TestFeatureLinkParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureLink(f); err == nil {
		t.Error("ParseFeatureLink() accepted a record with no id")
	}
	f = SampleFeatureLink("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseFeatureLink(f); err == nil {
		t.Error("ParseFeatureLink() accepted a non-numeric count")
	}
}

func TestFeatureLinkCollections(t *testing.T) {
	xs := []FeatureLink{SampleFeatureLink("a", when), SampleFeatureLink("bb", when), SampleFeatureLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureLinks(xs)); got != 2 {
		t.Errorf("ActiveFeatureLinks() kept %d, want 2", got)
	}
	if got, want := SumFeatureLinks(xs), 1+3; got != want {
		t.Errorf("SumFeatureLinks() = %d, want %d", got, want)
	}
	SortFeatureLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureLinks() put %q first", xs[0].ID)
	}
	if len(IndexFeatureLinks(xs)) != 3 {
		t.Errorf("IndexFeatureLinks() = %v", IndexFeatureLinks(xs))
	}
	if err := ValidateFeatureLinks(ActiveFeatureLinks(xs)); err != nil {
		t.Errorf("ValidateFeatureLinks() = %v", err)
	}
}

func TestFeatureLinkNormalizeClamps(t *testing.T) {
	x := SampleFeatureLink("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureNoteRoundTrip(t *testing.T) {
	x := SampleFeatureNote("featurenote-1", when)
	got, err := ParseFeatureNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureNote() = %v", err)
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

func TestFeatureNoteParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureNote(f); err == nil {
		t.Error("ParseFeatureNote() accepted a record with no id")
	}
	f = SampleFeatureNote("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseFeatureNote(f); err == nil {
		t.Error("ParseFeatureNote() accepted a non-numeric version")
	}
}

func TestFeatureNoteCollections(t *testing.T) {
	xs := []FeatureNote{SampleFeatureNote("a", when), SampleFeatureNote("bb", when), SampleFeatureNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureNotes(xs)); got != 2 {
		t.Errorf("ActiveFeatureNotes() kept %d, want 2", got)
	}
	if got, want := SumFeatureNotes(xs), 1+3; got != want {
		t.Errorf("SumFeatureNotes() = %d, want %d", got, want)
	}
	SortFeatureNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureNotes() put %q first", xs[0].ID)
	}
	if len(IndexFeatureNotes(xs)) != 3 {
		t.Errorf("IndexFeatureNotes() = %v", IndexFeatureNotes(xs))
	}
	if err := ValidateFeatureNotes(ActiveFeatureNotes(xs)); err != nil {
		t.Errorf("ValidateFeatureNotes() = %v", err)
	}
}

func TestFeatureNoteNormalizeClamps(t *testing.T) {
	x := SampleFeatureNote("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureWindowRoundTrip(t *testing.T) {
	x := SampleFeatureWindow("featurewindow-1", when)
	got, err := ParseFeatureWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureWindow() = %v", err)
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

func TestFeatureWindowParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureWindow(f); err == nil {
		t.Error("ParseFeatureWindow() accepted a record with no id")
	}
	f = SampleFeatureWindow("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseFeatureWindow(f); err == nil {
		t.Error("ParseFeatureWindow() accepted a non-numeric capacity")
	}
}

func TestFeatureWindowCollections(t *testing.T) {
	xs := []FeatureWindow{SampleFeatureWindow("a", when), SampleFeatureWindow("bb", when), SampleFeatureWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureWindows(xs)); got != 2 {
		t.Errorf("ActiveFeatureWindows() kept %d, want 2", got)
	}
	if got, want := SumFeatureWindows(xs), 1+3; got != want {
		t.Errorf("SumFeatureWindows() = %d, want %d", got, want)
	}
	SortFeatureWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureWindows() put %q first", xs[0].ID)
	}
	if len(IndexFeatureWindows(xs)) != 3 {
		t.Errorf("IndexFeatureWindows() = %v", IndexFeatureWindows(xs))
	}
	if err := ValidateFeatureWindows(ActiveFeatureWindows(xs)); err != nil {
		t.Errorf("ValidateFeatureWindows() = %v", err)
	}
}

func TestFeatureWindowNormalizeClamps(t *testing.T) {
	x := SampleFeatureWindow("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureQuotaRoundTrip(t *testing.T) {
	x := SampleFeatureQuota("featurequota-1", when)
	got, err := ParseFeatureQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureQuota() = %v", err)
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

func TestFeatureQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureQuota(f); err == nil {
		t.Error("ParseFeatureQuota() accepted a record with no id")
	}
	f = SampleFeatureQuota("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseFeatureQuota(f); err == nil {
		t.Error("ParseFeatureQuota() accepted a non-numeric count")
	}
}

func TestFeatureQuotaCollections(t *testing.T) {
	xs := []FeatureQuota{SampleFeatureQuota("a", when), SampleFeatureQuota("bb", when), SampleFeatureQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureQuotas(xs)); got != 2 {
		t.Errorf("ActiveFeatureQuotas() kept %d, want 2", got)
	}
	if got, want := SumFeatureQuotas(xs), 1+3; got != want {
		t.Errorf("SumFeatureQuotas() = %d, want %d", got, want)
	}
	SortFeatureQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureQuotas() put %q first", xs[0].ID)
	}
	if len(IndexFeatureQuotas(xs)) != 3 {
		t.Errorf("IndexFeatureQuotas() = %v", IndexFeatureQuotas(xs))
	}
	if err := ValidateFeatureQuotas(ActiveFeatureQuotas(xs)); err != nil {
		t.Errorf("ValidateFeatureQuotas() = %v", err)
	}
}

func TestFeatureQuotaNormalizeClamps(t *testing.T) {
	x := SampleFeatureQuota("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureBucketRoundTrip(t *testing.T) {
	x := SampleFeatureBucket("featurebucket-1", when)
	got, err := ParseFeatureBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureBucket() = %v", err)
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

func TestFeatureBucketParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureBucket(f); err == nil {
		t.Error("ParseFeatureBucket() accepted a record with no id")
	}
	f = SampleFeatureBucket("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseFeatureBucket(f); err == nil {
		t.Error("ParseFeatureBucket() accepted a non-numeric attempts")
	}
}

func TestFeatureBucketCollections(t *testing.T) {
	xs := []FeatureBucket{SampleFeatureBucket("a", when), SampleFeatureBucket("bb", when), SampleFeatureBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureBuckets(xs)); got != 2 {
		t.Errorf("ActiveFeatureBuckets() kept %d, want 2", got)
	}
	if got, want := SumFeatureBuckets(xs), 1+3; got != want {
		t.Errorf("SumFeatureBuckets() = %d, want %d", got, want)
	}
	SortFeatureBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureBuckets() put %q first", xs[0].ID)
	}
	if len(IndexFeatureBuckets(xs)) != 3 {
		t.Errorf("IndexFeatureBuckets() = %v", IndexFeatureBuckets(xs))
	}
	if err := ValidateFeatureBuckets(ActiveFeatureBuckets(xs)); err != nil {
		t.Errorf("ValidateFeatureBuckets() = %v", err)
	}
}

func TestFeatureBucketNormalizeClamps(t *testing.T) {
	x := SampleFeatureBucket("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureLaneRoundTrip(t *testing.T) {
	x := SampleFeatureLane("featurelane-1", when)
	got, err := ParseFeatureLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureLane() = %v", err)
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

func TestFeatureLaneParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureLane(f); err == nil {
		t.Error("ParseFeatureLane() accepted a record with no id")
	}
	f = SampleFeatureLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseFeatureLane(f); err == nil {
		t.Error("ParseFeatureLane() accepted a non-numeric attempts")
	}
}

func TestFeatureLaneCollections(t *testing.T) {
	xs := []FeatureLane{SampleFeatureLane("a", when), SampleFeatureLane("bb", when), SampleFeatureLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureLanes(xs)); got != 2 {
		t.Errorf("ActiveFeatureLanes() kept %d, want 2", got)
	}
	if got, want := SumFeatureLanes(xs), 1+3; got != want {
		t.Errorf("SumFeatureLanes() = %d, want %d", got, want)
	}
	SortFeatureLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureLanes() put %q first", xs[0].ID)
	}
	if len(IndexFeatureLanes(xs)) != 3 {
		t.Errorf("IndexFeatureLanes() = %v", IndexFeatureLanes(xs))
	}
	if err := ValidateFeatureLanes(ActiveFeatureLanes(xs)); err != nil {
		t.Errorf("ValidateFeatureLanes() = %v", err)
	}
}

func TestFeatureLaneNormalizeClamps(t *testing.T) {
	x := SampleFeatureLane("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureSlotRoundTrip(t *testing.T) {
	x := SampleFeatureSlot("featureslot-1", when)
	got, err := ParseFeatureSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureSlot() = %v", err)
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

func TestFeatureSlotParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureSlot(f); err == nil {
		t.Error("ParseFeatureSlot() accepted a record with no id")
	}
	f = SampleFeatureSlot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseFeatureSlot(f); err == nil {
		t.Error("ParseFeatureSlot() accepted a non-numeric priority")
	}
}

func TestFeatureSlotCollections(t *testing.T) {
	xs := []FeatureSlot{SampleFeatureSlot("a", when), SampleFeatureSlot("bb", when), SampleFeatureSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureSlots(xs)); got != 2 {
		t.Errorf("ActiveFeatureSlots() kept %d, want 2", got)
	}
	if got, want := SumFeatureSlots(xs), 1+3; got != want {
		t.Errorf("SumFeatureSlots() = %d, want %d", got, want)
	}
	SortFeatureSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureSlots() put %q first", xs[0].ID)
	}
	if len(IndexFeatureSlots(xs)) != 3 {
		t.Errorf("IndexFeatureSlots() = %v", IndexFeatureSlots(xs))
	}
	if err := ValidateFeatureSlots(ActiveFeatureSlots(xs)); err != nil {
		t.Errorf("ValidateFeatureSlots() = %v", err)
	}
}

func TestFeatureSlotNormalizeClamps(t *testing.T) {
	x := SampleFeatureSlot("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureTagRoundTrip(t *testing.T) {
	x := SampleFeatureTag("featuretag-1", when)
	got, err := ParseFeatureTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureTag() = %v", err)
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

func TestFeatureTagParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureTag(f); err == nil {
		t.Error("ParseFeatureTag() accepted a record with no id")
	}
	f = SampleFeatureTag("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseFeatureTag(f); err == nil {
		t.Error("ParseFeatureTag() accepted a non-numeric offset")
	}
}

func TestFeatureTagCollections(t *testing.T) {
	xs := []FeatureTag{SampleFeatureTag("a", when), SampleFeatureTag("bb", when), SampleFeatureTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureTags(xs)); got != 2 {
		t.Errorf("ActiveFeatureTags() kept %d, want 2", got)
	}
	if got, want := SumFeatureTags(xs), 1+3; got != want {
		t.Errorf("SumFeatureTags() = %d, want %d", got, want)
	}
	SortFeatureTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureTags() put %q first", xs[0].ID)
	}
	if len(IndexFeatureTags(xs)) != 3 {
		t.Errorf("IndexFeatureTags() = %v", IndexFeatureTags(xs))
	}
	if err := ValidateFeatureTags(ActiveFeatureTags(xs)); err != nil {
		t.Errorf("ValidateFeatureTags() = %v", err)
	}
}

func TestFeatureTagNormalizeClamps(t *testing.T) {
	x := SampleFeatureTag("a", when)
	x.Reference = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureRuleRoundTrip(t *testing.T) {
	x := SampleFeatureRule("featurerule-1", when)
	got, err := ParseFeatureRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureRule() = %v", err)
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

func TestFeatureRuleParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureRule(f); err == nil {
		t.Error("ParseFeatureRule() accepted a record with no id")
	}
	f = SampleFeatureRule("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseFeatureRule(f); err == nil {
		t.Error("ParseFeatureRule() accepted a non-numeric count")
	}
}

func TestFeatureRuleCollections(t *testing.T) {
	xs := []FeatureRule{SampleFeatureRule("a", when), SampleFeatureRule("bb", when), SampleFeatureRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureRules(xs)); got != 2 {
		t.Errorf("ActiveFeatureRules() kept %d, want 2", got)
	}
	if got, want := SumFeatureRules(xs), 1+3; got != want {
		t.Errorf("SumFeatureRules() = %d, want %d", got, want)
	}
	SortFeatureRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureRules() put %q first", xs[0].ID)
	}
	if len(IndexFeatureRules(xs)) != 3 {
		t.Errorf("IndexFeatureRules() = %v", IndexFeatureRules(xs))
	}
	if err := ValidateFeatureRules(ActiveFeatureRules(xs)); err != nil {
		t.Errorf("ValidateFeatureRules() = %v", err)
	}
}

func TestFeatureRuleNormalizeClamps(t *testing.T) {
	x := SampleFeatureRule("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureHoldRoundTrip(t *testing.T) {
	x := SampleFeatureHold("featurehold-1", when)
	got, err := ParseFeatureHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureHold() = %v", err)
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

func TestFeatureHoldParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureHold(f); err == nil {
		t.Error("ParseFeatureHold() accepted a record with no id")
	}
	f = SampleFeatureHold("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseFeatureHold(f); err == nil {
		t.Error("ParseFeatureHold() accepted a non-numeric priority")
	}
}

func TestFeatureHoldCollections(t *testing.T) {
	xs := []FeatureHold{SampleFeatureHold("a", when), SampleFeatureHold("bb", when), SampleFeatureHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureHolds(xs)); got != 2 {
		t.Errorf("ActiveFeatureHolds() kept %d, want 2", got)
	}
	if got, want := SumFeatureHolds(xs), 1+3; got != want {
		t.Errorf("SumFeatureHolds() = %d, want %d", got, want)
	}
	SortFeatureHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureHolds() put %q first", xs[0].ID)
	}
	if len(IndexFeatureHolds(xs)) != 3 {
		t.Errorf("IndexFeatureHolds() = %v", IndexFeatureHolds(xs))
	}
	if err := ValidateFeatureHolds(ActiveFeatureHolds(xs)); err != nil {
		t.Errorf("ValidateFeatureHolds() = %v", err)
	}
}

func TestFeatureHoldNormalizeClamps(t *testing.T) {
	x := SampleFeatureHold("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureClaimRoundTrip(t *testing.T) {
	x := SampleFeatureClaim("featureclaim-1", when)
	got, err := ParseFeatureClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureClaim() = %v", err)
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

func TestFeatureClaimParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureClaim(f); err == nil {
		t.Error("ParseFeatureClaim() accepted a record with no id")
	}
	f = SampleFeatureClaim("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseFeatureClaim(f); err == nil {
		t.Error("ParseFeatureClaim() accepted a non-numeric weight")
	}
}

func TestFeatureClaimCollections(t *testing.T) {
	xs := []FeatureClaim{SampleFeatureClaim("a", when), SampleFeatureClaim("bb", when), SampleFeatureClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureClaims(xs)); got != 2 {
		t.Errorf("ActiveFeatureClaims() kept %d, want 2", got)
	}
	if got, want := SumFeatureClaims(xs), 1+3; got != want {
		t.Errorf("SumFeatureClaims() = %d, want %d", got, want)
	}
	SortFeatureClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureClaims() put %q first", xs[0].ID)
	}
	if len(IndexFeatureClaims(xs)) != 3 {
		t.Errorf("IndexFeatureClaims() = %v", IndexFeatureClaims(xs))
	}
	if err := ValidateFeatureClaims(ActiveFeatureClaims(xs)); err != nil {
		t.Errorf("ValidateFeatureClaims() = %v", err)
	}
}

func TestFeatureClaimNormalizeClamps(t *testing.T) {
	x := SampleFeatureClaim("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestFeatureCursorRoundTrip(t *testing.T) {
	x := SampleFeatureCursor("featurecursor-1", when)
	got, err := ParseFeatureCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseFeatureCursor() = %v", err)
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

func TestFeatureCursorParseRejectsBadInput(t *testing.T) {
	f := SampleFeatureCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseFeatureCursor(f); err == nil {
		t.Error("ParseFeatureCursor() accepted a record with no id")
	}
	f = SampleFeatureCursor("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseFeatureCursor(f); err == nil {
		t.Error("ParseFeatureCursor() accepted a non-numeric count")
	}
}

func TestFeatureCursorCollections(t *testing.T) {
	xs := []FeatureCursor{SampleFeatureCursor("a", when), SampleFeatureCursor("bb", when), SampleFeatureCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveFeatureCursors(xs)); got != 2 {
		t.Errorf("ActiveFeatureCursors() kept %d, want 2", got)
	}
	if got, want := SumFeatureCursors(xs), 1+3; got != want {
		t.Errorf("SumFeatureCursors() = %d, want %d", got, want)
	}
	SortFeatureCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortFeatureCursors() put %q first", xs[0].ID)
	}
	if len(IndexFeatureCursors(xs)) != 3 {
		t.Errorf("IndexFeatureCursors() = %v", IndexFeatureCursors(xs))
	}
	if err := ValidateFeatureCursors(ActiveFeatureCursors(xs)); err != nil {
		t.Errorf("ValidateFeatureCursors() = %v", err)
	}
}

func TestFeatureCursorNormalizeClamps(t *testing.T) {
	x := SampleFeatureCursor("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
