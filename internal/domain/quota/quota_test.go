package quota

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
		if !strings.HasPrefix(Prefixes[i], "quota/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleQuotaRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestQuotaRecordRoundTrip(t *testing.T) {
	x := SampleQuotaRecord("quotarecord-1", when)
	got, err := ParseQuotaRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaRecord() = %v", err)
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

func TestQuotaRecordParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaRecord(f); err == nil {
		t.Error("ParseQuotaRecord() accepted a record with no id")
	}
	f = SampleQuotaRecord("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseQuotaRecord(f); err == nil {
		t.Error("ParseQuotaRecord() accepted a non-numeric version")
	}
}

func TestQuotaRecordCollections(t *testing.T) {
	xs := []QuotaRecord{SampleQuotaRecord("a", when), SampleQuotaRecord("bb", when), SampleQuotaRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaRecords(xs)); got != 2 {
		t.Errorf("ActiveQuotaRecords() kept %d, want 2", got)
	}
	if got, want := SumQuotaRecords(xs), 1+3; got != want {
		t.Errorf("SumQuotaRecords() = %d, want %d", got, want)
	}
	SortQuotaRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaRecords() put %q first", xs[0].ID)
	}
	if len(IndexQuotaRecords(xs)) != 3 {
		t.Errorf("IndexQuotaRecords() = %v", IndexQuotaRecords(xs))
	}
	if err := ValidateQuotaRecords(ActiveQuotaRecords(xs)); err != nil {
		t.Errorf("ValidateQuotaRecords() = %v", err)
	}
}

func TestQuotaRecordNormalizeClamps(t *testing.T) {
	x := SampleQuotaRecord("a", when)
	x.Label = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaEntryRoundTrip(t *testing.T) {
	x := SampleQuotaEntry("quotaentry-1", when)
	got, err := ParseQuotaEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaEntry() = %v", err)
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

func TestQuotaEntryParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaEntry(f); err == nil {
		t.Error("ParseQuotaEntry() accepted a record with no id")
	}
	f = SampleQuotaEntry("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseQuotaEntry(f); err == nil {
		t.Error("ParseQuotaEntry() accepted a non-numeric attempts")
	}
}

func TestQuotaEntryCollections(t *testing.T) {
	xs := []QuotaEntry{SampleQuotaEntry("a", when), SampleQuotaEntry("bb", when), SampleQuotaEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaEntrys(xs)); got != 2 {
		t.Errorf("ActiveQuotaEntrys() kept %d, want 2", got)
	}
	if got, want := SumQuotaEntrys(xs), 1+3; got != want {
		t.Errorf("SumQuotaEntrys() = %d, want %d", got, want)
	}
	SortQuotaEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaEntrys() put %q first", xs[0].ID)
	}
	if len(IndexQuotaEntrys(xs)) != 3 {
		t.Errorf("IndexQuotaEntrys() = %v", IndexQuotaEntrys(xs))
	}
	if err := ValidateQuotaEntrys(ActiveQuotaEntrys(xs)); err != nil {
		t.Errorf("ValidateQuotaEntrys() = %v", err)
	}
}

func TestQuotaEntryNormalizeClamps(t *testing.T) {
	x := SampleQuotaEntry("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaItemRoundTrip(t *testing.T) {
	x := SampleQuotaItem("quotaitem-1", when)
	got, err := ParseQuotaItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaItem() = %v", err)
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

func TestQuotaItemParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaItem(f); err == nil {
		t.Error("ParseQuotaItem() accepted a record with no id")
	}
	f = SampleQuotaItem("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseQuotaItem(f); err == nil {
		t.Error("ParseQuotaItem() accepted a non-numeric attempts")
	}
}

func TestQuotaItemCollections(t *testing.T) {
	xs := []QuotaItem{SampleQuotaItem("a", when), SampleQuotaItem("bb", when), SampleQuotaItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaItems(xs)); got != 2 {
		t.Errorf("ActiveQuotaItems() kept %d, want 2", got)
	}
	if got, want := SumQuotaItems(xs), 1+3; got != want {
		t.Errorf("SumQuotaItems() = %d, want %d", got, want)
	}
	SortQuotaItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaItems() put %q first", xs[0].ID)
	}
	if len(IndexQuotaItems(xs)) != 3 {
		t.Errorf("IndexQuotaItems() = %v", IndexQuotaItems(xs))
	}
	if err := ValidateQuotaItems(ActiveQuotaItems(xs)); err != nil {
		t.Errorf("ValidateQuotaItems() = %v", err)
	}
}

func TestQuotaItemNormalizeClamps(t *testing.T) {
	x := SampleQuotaItem("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaDraftRoundTrip(t *testing.T) {
	x := SampleQuotaDraft("quotadraft-1", when)
	got, err := ParseQuotaDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaDraft() = %v", err)
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

func TestQuotaDraftParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaDraft(f); err == nil {
		t.Error("ParseQuotaDraft() accepted a record with no id")
	}
	f = SampleQuotaDraft("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseQuotaDraft(f); err == nil {
		t.Error("ParseQuotaDraft() accepted a non-numeric version")
	}
}

func TestQuotaDraftCollections(t *testing.T) {
	xs := []QuotaDraft{SampleQuotaDraft("a", when), SampleQuotaDraft("bb", when), SampleQuotaDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaDrafts(xs)); got != 2 {
		t.Errorf("ActiveQuotaDrafts() kept %d, want 2", got)
	}
	if got, want := SumQuotaDrafts(xs), 1+3; got != want {
		t.Errorf("SumQuotaDrafts() = %d, want %d", got, want)
	}
	SortQuotaDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaDrafts() put %q first", xs[0].ID)
	}
	if len(IndexQuotaDrafts(xs)) != 3 {
		t.Errorf("IndexQuotaDrafts() = %v", IndexQuotaDrafts(xs))
	}
	if err := ValidateQuotaDrafts(ActiveQuotaDrafts(xs)); err != nil {
		t.Errorf("ValidateQuotaDrafts() = %v", err)
	}
}

func TestQuotaDraftNormalizeClamps(t *testing.T) {
	x := SampleQuotaDraft("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaRevisionRoundTrip(t *testing.T) {
	x := SampleQuotaRevision("quotarevision-1", when)
	got, err := ParseQuotaRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaRevision() = %v", err)
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

func TestQuotaRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaRevision(f); err == nil {
		t.Error("ParseQuotaRevision() accepted a record with no id")
	}
	f = SampleQuotaRevision("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseQuotaRevision(f); err == nil {
		t.Error("ParseQuotaRevision() accepted a non-numeric priority")
	}
}

func TestQuotaRevisionCollections(t *testing.T) {
	xs := []QuotaRevision{SampleQuotaRevision("a", when), SampleQuotaRevision("bb", when), SampleQuotaRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaRevisions(xs)); got != 2 {
		t.Errorf("ActiveQuotaRevisions() kept %d, want 2", got)
	}
	if got, want := SumQuotaRevisions(xs), 1+3; got != want {
		t.Errorf("SumQuotaRevisions() = %d, want %d", got, want)
	}
	SortQuotaRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaRevisions() put %q first", xs[0].ID)
	}
	if len(IndexQuotaRevisions(xs)) != 3 {
		t.Errorf("IndexQuotaRevisions() = %v", IndexQuotaRevisions(xs))
	}
	if err := ValidateQuotaRevisions(ActiveQuotaRevisions(xs)); err != nil {
		t.Errorf("ValidateQuotaRevisions() = %v", err)
	}
}

func TestQuotaRevisionNormalizeClamps(t *testing.T) {
	x := SampleQuotaRevision("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaSnapshotRoundTrip(t *testing.T) {
	x := SampleQuotaSnapshot("quotasnapshot-1", when)
	got, err := ParseQuotaSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaSnapshot() = %v", err)
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

func TestQuotaSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaSnapshot(f); err == nil {
		t.Error("ParseQuotaSnapshot() accepted a record with no id")
	}
	f = SampleQuotaSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseQuotaSnapshot(f); err == nil {
		t.Error("ParseQuotaSnapshot() accepted a non-numeric priority")
	}
}

func TestQuotaSnapshotCollections(t *testing.T) {
	xs := []QuotaSnapshot{SampleQuotaSnapshot("a", when), SampleQuotaSnapshot("bb", when), SampleQuotaSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaSnapshots(xs)); got != 2 {
		t.Errorf("ActiveQuotaSnapshots() kept %d, want 2", got)
	}
	if got, want := SumQuotaSnapshots(xs), 1+3; got != want {
		t.Errorf("SumQuotaSnapshots() = %d, want %d", got, want)
	}
	SortQuotaSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexQuotaSnapshots(xs)) != 3 {
		t.Errorf("IndexQuotaSnapshots() = %v", IndexQuotaSnapshots(xs))
	}
	if err := ValidateQuotaSnapshots(ActiveQuotaSnapshots(xs)); err != nil {
		t.Errorf("ValidateQuotaSnapshots() = %v", err)
	}
}

func TestQuotaSnapshotNormalizeClamps(t *testing.T) {
	x := SampleQuotaSnapshot("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaBatchRoundTrip(t *testing.T) {
	x := SampleQuotaBatch("quotabatch-1", when)
	got, err := ParseQuotaBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaBatch() = %v", err)
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

func TestQuotaBatchParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaBatch(f); err == nil {
		t.Error("ParseQuotaBatch() accepted a record with no id")
	}
	f = SampleQuotaBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseQuotaBatch(f); err == nil {
		t.Error("ParseQuotaBatch() accepted a non-numeric capacity")
	}
}

func TestQuotaBatchCollections(t *testing.T) {
	xs := []QuotaBatch{SampleQuotaBatch("a", when), SampleQuotaBatch("bb", when), SampleQuotaBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaBatchs(xs)); got != 2 {
		t.Errorf("ActiveQuotaBatchs() kept %d, want 2", got)
	}
	if got, want := SumQuotaBatchs(xs), 1+3; got != want {
		t.Errorf("SumQuotaBatchs() = %d, want %d", got, want)
	}
	SortQuotaBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaBatchs() put %q first", xs[0].ID)
	}
	if len(IndexQuotaBatchs(xs)) != 3 {
		t.Errorf("IndexQuotaBatchs() = %v", IndexQuotaBatchs(xs))
	}
	if err := ValidateQuotaBatchs(ActiveQuotaBatchs(xs)); err != nil {
		t.Errorf("ValidateQuotaBatchs() = %v", err)
	}
}

func TestQuotaBatchNormalizeClamps(t *testing.T) {
	x := SampleQuotaBatch("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaRequestRoundTrip(t *testing.T) {
	x := SampleQuotaRequest("quotarequest-1", when)
	got, err := ParseQuotaRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaRequest() = %v", err)
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

func TestQuotaRequestParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaRequest(f); err == nil {
		t.Error("ParseQuotaRequest() accepted a record with no id")
	}
	f = SampleQuotaRequest("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseQuotaRequest(f); err == nil {
		t.Error("ParseQuotaRequest() accepted a non-numeric capacity")
	}
}

func TestQuotaRequestCollections(t *testing.T) {
	xs := []QuotaRequest{SampleQuotaRequest("a", when), SampleQuotaRequest("bb", when), SampleQuotaRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaRequests(xs)); got != 2 {
		t.Errorf("ActiveQuotaRequests() kept %d, want 2", got)
	}
	if got, want := SumQuotaRequests(xs), 1+3; got != want {
		t.Errorf("SumQuotaRequests() = %d, want %d", got, want)
	}
	SortQuotaRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaRequests() put %q first", xs[0].ID)
	}
	if len(IndexQuotaRequests(xs)) != 3 {
		t.Errorf("IndexQuotaRequests() = %v", IndexQuotaRequests(xs))
	}
	if err := ValidateQuotaRequests(ActiveQuotaRequests(xs)); err != nil {
		t.Errorf("ValidateQuotaRequests() = %v", err)
	}
}

func TestQuotaRequestNormalizeClamps(t *testing.T) {
	x := SampleQuotaRequest("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaSummaryRoundTrip(t *testing.T) {
	x := SampleQuotaSummary("quotasummary-1", when)
	got, err := ParseQuotaSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaSummary() = %v", err)
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

func TestQuotaSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaSummary(f); err == nil {
		t.Error("ParseQuotaSummary() accepted a record with no id")
	}
	f = SampleQuotaSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseQuotaSummary(f); err == nil {
		t.Error("ParseQuotaSummary() accepted a non-numeric capacity")
	}
}

func TestQuotaSummaryCollections(t *testing.T) {
	xs := []QuotaSummary{SampleQuotaSummary("a", when), SampleQuotaSummary("bb", when), SampleQuotaSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaSummarys(xs)); got != 2 {
		t.Errorf("ActiveQuotaSummarys() kept %d, want 2", got)
	}
	if got, want := SumQuotaSummarys(xs), 1+3; got != want {
		t.Errorf("SumQuotaSummarys() = %d, want %d", got, want)
	}
	SortQuotaSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaSummarys() put %q first", xs[0].ID)
	}
	if len(IndexQuotaSummarys(xs)) != 3 {
		t.Errorf("IndexQuotaSummarys() = %v", IndexQuotaSummarys(xs))
	}
	if err := ValidateQuotaSummarys(ActiveQuotaSummarys(xs)); err != nil {
		t.Errorf("ValidateQuotaSummarys() = %v", err)
	}
}

func TestQuotaSummaryNormalizeClamps(t *testing.T) {
	x := SampleQuotaSummary("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaPolicyRoundTrip(t *testing.T) {
	x := SampleQuotaPolicy("quotapolicy-1", when)
	got, err := ParseQuotaPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaPolicy() = %v", err)
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

func TestQuotaPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaPolicy(f); err == nil {
		t.Error("ParseQuotaPolicy() accepted a record with no id")
	}
	f = SampleQuotaPolicy("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseQuotaPolicy(f); err == nil {
		t.Error("ParseQuotaPolicy() accepted a non-numeric offset")
	}
}

func TestQuotaPolicyCollections(t *testing.T) {
	xs := []QuotaPolicy{SampleQuotaPolicy("a", when), SampleQuotaPolicy("bb", when), SampleQuotaPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaPolicys(xs)); got != 2 {
		t.Errorf("ActiveQuotaPolicys() kept %d, want 2", got)
	}
	if got, want := SumQuotaPolicys(xs), 1+3; got != want {
		t.Errorf("SumQuotaPolicys() = %d, want %d", got, want)
	}
	SortQuotaPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaPolicys() put %q first", xs[0].ID)
	}
	if len(IndexQuotaPolicys(xs)) != 3 {
		t.Errorf("IndexQuotaPolicys() = %v", IndexQuotaPolicys(xs))
	}
	if err := ValidateQuotaPolicys(ActiveQuotaPolicys(xs)); err != nil {
		t.Errorf("ValidateQuotaPolicys() = %v", err)
	}
}

func TestQuotaPolicyNormalizeClamps(t *testing.T) {
	x := SampleQuotaPolicy("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaLinkRoundTrip(t *testing.T) {
	x := SampleQuotaLink("quotalink-1", when)
	got, err := ParseQuotaLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaLink() = %v", err)
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

func TestQuotaLinkParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaLink(f); err == nil {
		t.Error("ParseQuotaLink() accepted a record with no id")
	}
	f = SampleQuotaLink("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseQuotaLink(f); err == nil {
		t.Error("ParseQuotaLink() accepted a non-numeric attempts")
	}
}

func TestQuotaLinkCollections(t *testing.T) {
	xs := []QuotaLink{SampleQuotaLink("a", when), SampleQuotaLink("bb", when), SampleQuotaLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaLinks(xs)); got != 2 {
		t.Errorf("ActiveQuotaLinks() kept %d, want 2", got)
	}
	if got, want := SumQuotaLinks(xs), 1+3; got != want {
		t.Errorf("SumQuotaLinks() = %d, want %d", got, want)
	}
	SortQuotaLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaLinks() put %q first", xs[0].ID)
	}
	if len(IndexQuotaLinks(xs)) != 3 {
		t.Errorf("IndexQuotaLinks() = %v", IndexQuotaLinks(xs))
	}
	if err := ValidateQuotaLinks(ActiveQuotaLinks(xs)); err != nil {
		t.Errorf("ValidateQuotaLinks() = %v", err)
	}
}

func TestQuotaLinkNormalizeClamps(t *testing.T) {
	x := SampleQuotaLink("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaNoteRoundTrip(t *testing.T) {
	x := SampleQuotaNote("quotanote-1", when)
	got, err := ParseQuotaNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaNote() = %v", err)
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

func TestQuotaNoteParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaNote(f); err == nil {
		t.Error("ParseQuotaNote() accepted a record with no id")
	}
	f = SampleQuotaNote("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseQuotaNote(f); err == nil {
		t.Error("ParseQuotaNote() accepted a non-numeric attempts")
	}
}

func TestQuotaNoteCollections(t *testing.T) {
	xs := []QuotaNote{SampleQuotaNote("a", when), SampleQuotaNote("bb", when), SampleQuotaNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaNotes(xs)); got != 2 {
		t.Errorf("ActiveQuotaNotes() kept %d, want 2", got)
	}
	if got, want := SumQuotaNotes(xs), 1+3; got != want {
		t.Errorf("SumQuotaNotes() = %d, want %d", got, want)
	}
	SortQuotaNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaNotes() put %q first", xs[0].ID)
	}
	if len(IndexQuotaNotes(xs)) != 3 {
		t.Errorf("IndexQuotaNotes() = %v", IndexQuotaNotes(xs))
	}
	if err := ValidateQuotaNotes(ActiveQuotaNotes(xs)); err != nil {
		t.Errorf("ValidateQuotaNotes() = %v", err)
	}
}

func TestQuotaNoteNormalizeClamps(t *testing.T) {
	x := SampleQuotaNote("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaWindowRoundTrip(t *testing.T) {
	x := SampleQuotaWindow("quotawindow-1", when)
	got, err := ParseQuotaWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaWindow() = %v", err)
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

func TestQuotaWindowParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaWindow(f); err == nil {
		t.Error("ParseQuotaWindow() accepted a record with no id")
	}
	f = SampleQuotaWindow("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseQuotaWindow(f); err == nil {
		t.Error("ParseQuotaWindow() accepted a non-numeric attempts")
	}
}

func TestQuotaWindowCollections(t *testing.T) {
	xs := []QuotaWindow{SampleQuotaWindow("a", when), SampleQuotaWindow("bb", when), SampleQuotaWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaWindows(xs)); got != 2 {
		t.Errorf("ActiveQuotaWindows() kept %d, want 2", got)
	}
	if got, want := SumQuotaWindows(xs), 1+3; got != want {
		t.Errorf("SumQuotaWindows() = %d, want %d", got, want)
	}
	SortQuotaWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaWindows() put %q first", xs[0].ID)
	}
	if len(IndexQuotaWindows(xs)) != 3 {
		t.Errorf("IndexQuotaWindows() = %v", IndexQuotaWindows(xs))
	}
	if err := ValidateQuotaWindows(ActiveQuotaWindows(xs)); err != nil {
		t.Errorf("ValidateQuotaWindows() = %v", err)
	}
}

func TestQuotaWindowNormalizeClamps(t *testing.T) {
	x := SampleQuotaWindow("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaQuotaRoundTrip(t *testing.T) {
	x := SampleQuotaQuota("quotaquota-1", when)
	got, err := ParseQuotaQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaQuota() = %v", err)
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

func TestQuotaQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaQuota(f); err == nil {
		t.Error("ParseQuotaQuota() accepted a record with no id")
	}
	f = SampleQuotaQuota("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseQuotaQuota(f); err == nil {
		t.Error("ParseQuotaQuota() accepted a non-numeric attempts")
	}
}

func TestQuotaQuotaCollections(t *testing.T) {
	xs := []QuotaQuota{SampleQuotaQuota("a", when), SampleQuotaQuota("bb", when), SampleQuotaQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaQuotas(xs)); got != 2 {
		t.Errorf("ActiveQuotaQuotas() kept %d, want 2", got)
	}
	if got, want := SumQuotaQuotas(xs), 1+3; got != want {
		t.Errorf("SumQuotaQuotas() = %d, want %d", got, want)
	}
	SortQuotaQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaQuotas() put %q first", xs[0].ID)
	}
	if len(IndexQuotaQuotas(xs)) != 3 {
		t.Errorf("IndexQuotaQuotas() = %v", IndexQuotaQuotas(xs))
	}
	if err := ValidateQuotaQuotas(ActiveQuotaQuotas(xs)); err != nil {
		t.Errorf("ValidateQuotaQuotas() = %v", err)
	}
}

func TestQuotaQuotaNormalizeClamps(t *testing.T) {
	x := SampleQuotaQuota("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaBucketRoundTrip(t *testing.T) {
	x := SampleQuotaBucket("quotabucket-1", when)
	got, err := ParseQuotaBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaBucket() = %v", err)
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

func TestQuotaBucketParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaBucket(f); err == nil {
		t.Error("ParseQuotaBucket() accepted a record with no id")
	}
	f = SampleQuotaBucket("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseQuotaBucket(f); err == nil {
		t.Error("ParseQuotaBucket() accepted a non-numeric weight")
	}
}

func TestQuotaBucketCollections(t *testing.T) {
	xs := []QuotaBucket{SampleQuotaBucket("a", when), SampleQuotaBucket("bb", when), SampleQuotaBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaBuckets(xs)); got != 2 {
		t.Errorf("ActiveQuotaBuckets() kept %d, want 2", got)
	}
	if got, want := SumQuotaBuckets(xs), 1+3; got != want {
		t.Errorf("SumQuotaBuckets() = %d, want %d", got, want)
	}
	SortQuotaBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaBuckets() put %q first", xs[0].ID)
	}
	if len(IndexQuotaBuckets(xs)) != 3 {
		t.Errorf("IndexQuotaBuckets() = %v", IndexQuotaBuckets(xs))
	}
	if err := ValidateQuotaBuckets(ActiveQuotaBuckets(xs)); err != nil {
		t.Errorf("ValidateQuotaBuckets() = %v", err)
	}
}

func TestQuotaBucketNormalizeClamps(t *testing.T) {
	x := SampleQuotaBucket("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaLaneRoundTrip(t *testing.T) {
	x := SampleQuotaLane("quotalane-1", when)
	got, err := ParseQuotaLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaLane() = %v", err)
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

func TestQuotaLaneParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaLane(f); err == nil {
		t.Error("ParseQuotaLane() accepted a record with no id")
	}
	f = SampleQuotaLane("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseQuotaLane(f); err == nil {
		t.Error("ParseQuotaLane() accepted a non-numeric weight")
	}
}

func TestQuotaLaneCollections(t *testing.T) {
	xs := []QuotaLane{SampleQuotaLane("a", when), SampleQuotaLane("bb", when), SampleQuotaLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaLanes(xs)); got != 2 {
		t.Errorf("ActiveQuotaLanes() kept %d, want 2", got)
	}
	if got, want := SumQuotaLanes(xs), 1+3; got != want {
		t.Errorf("SumQuotaLanes() = %d, want %d", got, want)
	}
	SortQuotaLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaLanes() put %q first", xs[0].ID)
	}
	if len(IndexQuotaLanes(xs)) != 3 {
		t.Errorf("IndexQuotaLanes() = %v", IndexQuotaLanes(xs))
	}
	if err := ValidateQuotaLanes(ActiveQuotaLanes(xs)); err != nil {
		t.Errorf("ValidateQuotaLanes() = %v", err)
	}
}

func TestQuotaLaneNormalizeClamps(t *testing.T) {
	x := SampleQuotaLane("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaSlotRoundTrip(t *testing.T) {
	x := SampleQuotaSlot("quotaslot-1", when)
	got, err := ParseQuotaSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaSlot() = %v", err)
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

func TestQuotaSlotParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaSlot(f); err == nil {
		t.Error("ParseQuotaSlot() accepted a record with no id")
	}
	f = SampleQuotaSlot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseQuotaSlot(f); err == nil {
		t.Error("ParseQuotaSlot() accepted a non-numeric retained")
	}
}

func TestQuotaSlotCollections(t *testing.T) {
	xs := []QuotaSlot{SampleQuotaSlot("a", when), SampleQuotaSlot("bb", when), SampleQuotaSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaSlots(xs)); got != 2 {
		t.Errorf("ActiveQuotaSlots() kept %d, want 2", got)
	}
	if got, want := SumQuotaSlots(xs), 1+3; got != want {
		t.Errorf("SumQuotaSlots() = %d, want %d", got, want)
	}
	SortQuotaSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaSlots() put %q first", xs[0].ID)
	}
	if len(IndexQuotaSlots(xs)) != 3 {
		t.Errorf("IndexQuotaSlots() = %v", IndexQuotaSlots(xs))
	}
	if err := ValidateQuotaSlots(ActiveQuotaSlots(xs)); err != nil {
		t.Errorf("ValidateQuotaSlots() = %v", err)
	}
}

func TestQuotaSlotNormalizeClamps(t *testing.T) {
	x := SampleQuotaSlot("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaTagRoundTrip(t *testing.T) {
	x := SampleQuotaTag("quotatag-1", when)
	got, err := ParseQuotaTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaTag() = %v", err)
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

func TestQuotaTagParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaTag(f); err == nil {
		t.Error("ParseQuotaTag() accepted a record with no id")
	}
	f = SampleQuotaTag("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseQuotaTag(f); err == nil {
		t.Error("ParseQuotaTag() accepted a non-numeric offset")
	}
}

func TestQuotaTagCollections(t *testing.T) {
	xs := []QuotaTag{SampleQuotaTag("a", when), SampleQuotaTag("bb", when), SampleQuotaTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaTags(xs)); got != 2 {
		t.Errorf("ActiveQuotaTags() kept %d, want 2", got)
	}
	if got, want := SumQuotaTags(xs), 1+3; got != want {
		t.Errorf("SumQuotaTags() = %d, want %d", got, want)
	}
	SortQuotaTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaTags() put %q first", xs[0].ID)
	}
	if len(IndexQuotaTags(xs)) != 3 {
		t.Errorf("IndexQuotaTags() = %v", IndexQuotaTags(xs))
	}
	if err := ValidateQuotaTags(ActiveQuotaTags(xs)); err != nil {
		t.Errorf("ValidateQuotaTags() = %v", err)
	}
}

func TestQuotaTagNormalizeClamps(t *testing.T) {
	x := SampleQuotaTag("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaRuleRoundTrip(t *testing.T) {
	x := SampleQuotaRule("quotarule-1", when)
	got, err := ParseQuotaRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaRule() = %v", err)
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

func TestQuotaRuleParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaRule(f); err == nil {
		t.Error("ParseQuotaRule() accepted a record with no id")
	}
	f = SampleQuotaRule("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseQuotaRule(f); err == nil {
		t.Error("ParseQuotaRule() accepted a non-numeric version")
	}
}

func TestQuotaRuleCollections(t *testing.T) {
	xs := []QuotaRule{SampleQuotaRule("a", when), SampleQuotaRule("bb", when), SampleQuotaRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaRules(xs)); got != 2 {
		t.Errorf("ActiveQuotaRules() kept %d, want 2", got)
	}
	if got, want := SumQuotaRules(xs), 1+3; got != want {
		t.Errorf("SumQuotaRules() = %d, want %d", got, want)
	}
	SortQuotaRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaRules() put %q first", xs[0].ID)
	}
	if len(IndexQuotaRules(xs)) != 3 {
		t.Errorf("IndexQuotaRules() = %v", IndexQuotaRules(xs))
	}
	if err := ValidateQuotaRules(ActiveQuotaRules(xs)); err != nil {
		t.Errorf("ValidateQuotaRules() = %v", err)
	}
}

func TestQuotaRuleNormalizeClamps(t *testing.T) {
	x := SampleQuotaRule("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaHoldRoundTrip(t *testing.T) {
	x := SampleQuotaHold("quotahold-1", when)
	got, err := ParseQuotaHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaHold() = %v", err)
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

func TestQuotaHoldParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaHold(f); err == nil {
		t.Error("ParseQuotaHold() accepted a record with no id")
	}
	f = SampleQuotaHold("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseQuotaHold(f); err == nil {
		t.Error("ParseQuotaHold() accepted a non-numeric capacity")
	}
}

func TestQuotaHoldCollections(t *testing.T) {
	xs := []QuotaHold{SampleQuotaHold("a", when), SampleQuotaHold("bb", when), SampleQuotaHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaHolds(xs)); got != 2 {
		t.Errorf("ActiveQuotaHolds() kept %d, want 2", got)
	}
	if got, want := SumQuotaHolds(xs), 1+3; got != want {
		t.Errorf("SumQuotaHolds() = %d, want %d", got, want)
	}
	SortQuotaHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaHolds() put %q first", xs[0].ID)
	}
	if len(IndexQuotaHolds(xs)) != 3 {
		t.Errorf("IndexQuotaHolds() = %v", IndexQuotaHolds(xs))
	}
	if err := ValidateQuotaHolds(ActiveQuotaHolds(xs)); err != nil {
		t.Errorf("ValidateQuotaHolds() = %v", err)
	}
}

func TestQuotaHoldNormalizeClamps(t *testing.T) {
	x := SampleQuotaHold("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaClaimRoundTrip(t *testing.T) {
	x := SampleQuotaClaim("quotaclaim-1", when)
	got, err := ParseQuotaClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaClaim() = %v", err)
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

func TestQuotaClaimParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaClaim(f); err == nil {
		t.Error("ParseQuotaClaim() accepted a record with no id")
	}
	f = SampleQuotaClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseQuotaClaim(f); err == nil {
		t.Error("ParseQuotaClaim() accepted a non-numeric retained")
	}
}

func TestQuotaClaimCollections(t *testing.T) {
	xs := []QuotaClaim{SampleQuotaClaim("a", when), SampleQuotaClaim("bb", when), SampleQuotaClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaClaims(xs)); got != 2 {
		t.Errorf("ActiveQuotaClaims() kept %d, want 2", got)
	}
	if got, want := SumQuotaClaims(xs), 1+3; got != want {
		t.Errorf("SumQuotaClaims() = %d, want %d", got, want)
	}
	SortQuotaClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaClaims() put %q first", xs[0].ID)
	}
	if len(IndexQuotaClaims(xs)) != 3 {
		t.Errorf("IndexQuotaClaims() = %v", IndexQuotaClaims(xs))
	}
	if err := ValidateQuotaClaims(ActiveQuotaClaims(xs)); err != nil {
		t.Errorf("ValidateQuotaClaims() = %v", err)
	}
}

func TestQuotaClaimNormalizeClamps(t *testing.T) {
	x := SampleQuotaClaim("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestQuotaCursorRoundTrip(t *testing.T) {
	x := SampleQuotaCursor("quotacursor-1", when)
	got, err := ParseQuotaCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseQuotaCursor() = %v", err)
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

func TestQuotaCursorParseRejectsBadInput(t *testing.T) {
	f := SampleQuotaCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseQuotaCursor(f); err == nil {
		t.Error("ParseQuotaCursor() accepted a record with no id")
	}
	f = SampleQuotaCursor("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseQuotaCursor(f); err == nil {
		t.Error("ParseQuotaCursor() accepted a non-numeric priority")
	}
}

func TestQuotaCursorCollections(t *testing.T) {
	xs := []QuotaCursor{SampleQuotaCursor("a", when), SampleQuotaCursor("bb", when), SampleQuotaCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveQuotaCursors(xs)); got != 2 {
		t.Errorf("ActiveQuotaCursors() kept %d, want 2", got)
	}
	if got, want := SumQuotaCursors(xs), 1+3; got != want {
		t.Errorf("SumQuotaCursors() = %d, want %d", got, want)
	}
	SortQuotaCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortQuotaCursors() put %q first", xs[0].ID)
	}
	if len(IndexQuotaCursors(xs)) != 3 {
		t.Errorf("IndexQuotaCursors() = %v", IndexQuotaCursors(xs))
	}
	if err := ValidateQuotaCursors(ActiveQuotaCursors(xs)); err != nil {
		t.Errorf("ValidateQuotaCursors() = %v", err)
	}
}

func TestQuotaCursorNormalizeClamps(t *testing.T) {
	x := SampleQuotaCursor("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
