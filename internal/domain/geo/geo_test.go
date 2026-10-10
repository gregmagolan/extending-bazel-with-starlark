package geo

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
		if !strings.HasPrefix(Prefixes[i], "geo/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleGeoRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestGeoRecordRoundTrip(t *testing.T) {
	x := SampleGeoRecord("georecord-1", when)
	got, err := ParseGeoRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoRecord() = %v", err)
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

func TestGeoRecordParseRejectsBadInput(t *testing.T) {
	f := SampleGeoRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoRecord(f); err == nil {
		t.Error("ParseGeoRecord() accepted a record with no id")
	}
	f = SampleGeoRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseGeoRecord(f); err == nil {
		t.Error("ParseGeoRecord() accepted a non-numeric capacity")
	}
}

func TestGeoRecordCollections(t *testing.T) {
	xs := []GeoRecord{SampleGeoRecord("a", when), SampleGeoRecord("bb", when), SampleGeoRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoRecords(xs)); got != 2 {
		t.Errorf("ActiveGeoRecords() kept %d, want 2", got)
	}
	if got, want := SumGeoRecords(xs), 1+3; got != want {
		t.Errorf("SumGeoRecords() = %d, want %d", got, want)
	}
	SortGeoRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoRecords() put %q first", xs[0].ID)
	}
	if len(IndexGeoRecords(xs)) != 3 {
		t.Errorf("IndexGeoRecords() = %v", IndexGeoRecords(xs))
	}
	if err := ValidateGeoRecords(ActiveGeoRecords(xs)); err != nil {
		t.Errorf("ValidateGeoRecords() = %v", err)
	}
}

func TestGeoRecordNormalizeClamps(t *testing.T) {
	x := SampleGeoRecord("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoEntryRoundTrip(t *testing.T) {
	x := SampleGeoEntry("geoentry-1", when)
	got, err := ParseGeoEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoEntry() = %v", err)
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

func TestGeoEntryParseRejectsBadInput(t *testing.T) {
	f := SampleGeoEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoEntry(f); err == nil {
		t.Error("ParseGeoEntry() accepted a record with no id")
	}
	f = SampleGeoEntry("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseGeoEntry(f); err == nil {
		t.Error("ParseGeoEntry() accepted a non-numeric capacity")
	}
}

func TestGeoEntryCollections(t *testing.T) {
	xs := []GeoEntry{SampleGeoEntry("a", when), SampleGeoEntry("bb", when), SampleGeoEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoEntrys(xs)); got != 2 {
		t.Errorf("ActiveGeoEntrys() kept %d, want 2", got)
	}
	if got, want := SumGeoEntrys(xs), 1+3; got != want {
		t.Errorf("SumGeoEntrys() = %d, want %d", got, want)
	}
	SortGeoEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoEntrys() put %q first", xs[0].ID)
	}
	if len(IndexGeoEntrys(xs)) != 3 {
		t.Errorf("IndexGeoEntrys() = %v", IndexGeoEntrys(xs))
	}
	if err := ValidateGeoEntrys(ActiveGeoEntrys(xs)); err != nil {
		t.Errorf("ValidateGeoEntrys() = %v", err)
	}
}

func TestGeoEntryNormalizeClamps(t *testing.T) {
	x := SampleGeoEntry("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoItemRoundTrip(t *testing.T) {
	x := SampleGeoItem("geoitem-1", when)
	got, err := ParseGeoItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoItem() = %v", err)
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

func TestGeoItemParseRejectsBadInput(t *testing.T) {
	f := SampleGeoItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoItem(f); err == nil {
		t.Error("ParseGeoItem() accepted a record with no id")
	}
	f = SampleGeoItem("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseGeoItem(f); err == nil {
		t.Error("ParseGeoItem() accepted a non-numeric retained")
	}
}

func TestGeoItemCollections(t *testing.T) {
	xs := []GeoItem{SampleGeoItem("a", when), SampleGeoItem("bb", when), SampleGeoItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoItems(xs)); got != 2 {
		t.Errorf("ActiveGeoItems() kept %d, want 2", got)
	}
	if got, want := SumGeoItems(xs), 1+3; got != want {
		t.Errorf("SumGeoItems() = %d, want %d", got, want)
	}
	SortGeoItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoItems() put %q first", xs[0].ID)
	}
	if len(IndexGeoItems(xs)) != 3 {
		t.Errorf("IndexGeoItems() = %v", IndexGeoItems(xs))
	}
	if err := ValidateGeoItems(ActiveGeoItems(xs)); err != nil {
		t.Errorf("ValidateGeoItems() = %v", err)
	}
}

func TestGeoItemNormalizeClamps(t *testing.T) {
	x := SampleGeoItem("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoDraftRoundTrip(t *testing.T) {
	x := SampleGeoDraft("geodraft-1", when)
	got, err := ParseGeoDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoDraft() = %v", err)
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

func TestGeoDraftParseRejectsBadInput(t *testing.T) {
	f := SampleGeoDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoDraft(f); err == nil {
		t.Error("ParseGeoDraft() accepted a record with no id")
	}
	f = SampleGeoDraft("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseGeoDraft(f); err == nil {
		t.Error("ParseGeoDraft() accepted a non-numeric count")
	}
}

func TestGeoDraftCollections(t *testing.T) {
	xs := []GeoDraft{SampleGeoDraft("a", when), SampleGeoDraft("bb", when), SampleGeoDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoDrafts(xs)); got != 2 {
		t.Errorf("ActiveGeoDrafts() kept %d, want 2", got)
	}
	if got, want := SumGeoDrafts(xs), 1+3; got != want {
		t.Errorf("SumGeoDrafts() = %d, want %d", got, want)
	}
	SortGeoDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoDrafts() put %q first", xs[0].ID)
	}
	if len(IndexGeoDrafts(xs)) != 3 {
		t.Errorf("IndexGeoDrafts() = %v", IndexGeoDrafts(xs))
	}
	if err := ValidateGeoDrafts(ActiveGeoDrafts(xs)); err != nil {
		t.Errorf("ValidateGeoDrafts() = %v", err)
	}
}

func TestGeoDraftNormalizeClamps(t *testing.T) {
	x := SampleGeoDraft("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoRevisionRoundTrip(t *testing.T) {
	x := SampleGeoRevision("georevision-1", when)
	got, err := ParseGeoRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoRevision() = %v", err)
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

func TestGeoRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleGeoRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoRevision(f); err == nil {
		t.Error("ParseGeoRevision() accepted a record with no id")
	}
	f = SampleGeoRevision("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseGeoRevision(f); err == nil {
		t.Error("ParseGeoRevision() accepted a non-numeric capacity")
	}
}

func TestGeoRevisionCollections(t *testing.T) {
	xs := []GeoRevision{SampleGeoRevision("a", when), SampleGeoRevision("bb", when), SampleGeoRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoRevisions(xs)); got != 2 {
		t.Errorf("ActiveGeoRevisions() kept %d, want 2", got)
	}
	if got, want := SumGeoRevisions(xs), 1+3; got != want {
		t.Errorf("SumGeoRevisions() = %d, want %d", got, want)
	}
	SortGeoRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoRevisions() put %q first", xs[0].ID)
	}
	if len(IndexGeoRevisions(xs)) != 3 {
		t.Errorf("IndexGeoRevisions() = %v", IndexGeoRevisions(xs))
	}
	if err := ValidateGeoRevisions(ActiveGeoRevisions(xs)); err != nil {
		t.Errorf("ValidateGeoRevisions() = %v", err)
	}
}

func TestGeoRevisionNormalizeClamps(t *testing.T) {
	x := SampleGeoRevision("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoSnapshotRoundTrip(t *testing.T) {
	x := SampleGeoSnapshot("geosnapshot-1", when)
	got, err := ParseGeoSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoSnapshot() = %v", err)
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

func TestGeoSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleGeoSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoSnapshot(f); err == nil {
		t.Error("ParseGeoSnapshot() accepted a record with no id")
	}
	f = SampleGeoSnapshot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseGeoSnapshot(f); err == nil {
		t.Error("ParseGeoSnapshot() accepted a non-numeric weight")
	}
}

func TestGeoSnapshotCollections(t *testing.T) {
	xs := []GeoSnapshot{SampleGeoSnapshot("a", when), SampleGeoSnapshot("bb", when), SampleGeoSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoSnapshots(xs)); got != 2 {
		t.Errorf("ActiveGeoSnapshots() kept %d, want 2", got)
	}
	if got, want := SumGeoSnapshots(xs), 1+3; got != want {
		t.Errorf("SumGeoSnapshots() = %d, want %d", got, want)
	}
	SortGeoSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexGeoSnapshots(xs)) != 3 {
		t.Errorf("IndexGeoSnapshots() = %v", IndexGeoSnapshots(xs))
	}
	if err := ValidateGeoSnapshots(ActiveGeoSnapshots(xs)); err != nil {
		t.Errorf("ValidateGeoSnapshots() = %v", err)
	}
}

func TestGeoSnapshotNormalizeClamps(t *testing.T) {
	x := SampleGeoSnapshot("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoBatchRoundTrip(t *testing.T) {
	x := SampleGeoBatch("geobatch-1", when)
	got, err := ParseGeoBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoBatch() = %v", err)
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

func TestGeoBatchParseRejectsBadInput(t *testing.T) {
	f := SampleGeoBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoBatch(f); err == nil {
		t.Error("ParseGeoBatch() accepted a record with no id")
	}
	f = SampleGeoBatch("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseGeoBatch(f); err == nil {
		t.Error("ParseGeoBatch() accepted a non-numeric count")
	}
}

func TestGeoBatchCollections(t *testing.T) {
	xs := []GeoBatch{SampleGeoBatch("a", when), SampleGeoBatch("bb", when), SampleGeoBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoBatchs(xs)); got != 2 {
		t.Errorf("ActiveGeoBatchs() kept %d, want 2", got)
	}
	if got, want := SumGeoBatchs(xs), 1+3; got != want {
		t.Errorf("SumGeoBatchs() = %d, want %d", got, want)
	}
	SortGeoBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoBatchs() put %q first", xs[0].ID)
	}
	if len(IndexGeoBatchs(xs)) != 3 {
		t.Errorf("IndexGeoBatchs() = %v", IndexGeoBatchs(xs))
	}
	if err := ValidateGeoBatchs(ActiveGeoBatchs(xs)); err != nil {
		t.Errorf("ValidateGeoBatchs() = %v", err)
	}
}

func TestGeoBatchNormalizeClamps(t *testing.T) {
	x := SampleGeoBatch("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoRequestRoundTrip(t *testing.T) {
	x := SampleGeoRequest("georequest-1", when)
	got, err := ParseGeoRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoRequest() = %v", err)
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

func TestGeoRequestParseRejectsBadInput(t *testing.T) {
	f := SampleGeoRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoRequest(f); err == nil {
		t.Error("ParseGeoRequest() accepted a record with no id")
	}
	f = SampleGeoRequest("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseGeoRequest(f); err == nil {
		t.Error("ParseGeoRequest() accepted a non-numeric retained")
	}
}

func TestGeoRequestCollections(t *testing.T) {
	xs := []GeoRequest{SampleGeoRequest("a", when), SampleGeoRequest("bb", when), SampleGeoRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoRequests(xs)); got != 2 {
		t.Errorf("ActiveGeoRequests() kept %d, want 2", got)
	}
	if got, want := SumGeoRequests(xs), 1+3; got != want {
		t.Errorf("SumGeoRequests() = %d, want %d", got, want)
	}
	SortGeoRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoRequests() put %q first", xs[0].ID)
	}
	if len(IndexGeoRequests(xs)) != 3 {
		t.Errorf("IndexGeoRequests() = %v", IndexGeoRequests(xs))
	}
	if err := ValidateGeoRequests(ActiveGeoRequests(xs)); err != nil {
		t.Errorf("ValidateGeoRequests() = %v", err)
	}
}

func TestGeoRequestNormalizeClamps(t *testing.T) {
	x := SampleGeoRequest("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoSummaryRoundTrip(t *testing.T) {
	x := SampleGeoSummary("geosummary-1", when)
	got, err := ParseGeoSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoSummary() = %v", err)
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

func TestGeoSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleGeoSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoSummary(f); err == nil {
		t.Error("ParseGeoSummary() accepted a record with no id")
	}
	f = SampleGeoSummary("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseGeoSummary(f); err == nil {
		t.Error("ParseGeoSummary() accepted a non-numeric retained")
	}
}

func TestGeoSummaryCollections(t *testing.T) {
	xs := []GeoSummary{SampleGeoSummary("a", when), SampleGeoSummary("bb", when), SampleGeoSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoSummarys(xs)); got != 2 {
		t.Errorf("ActiveGeoSummarys() kept %d, want 2", got)
	}
	if got, want := SumGeoSummarys(xs), 1+3; got != want {
		t.Errorf("SumGeoSummarys() = %d, want %d", got, want)
	}
	SortGeoSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoSummarys() put %q first", xs[0].ID)
	}
	if len(IndexGeoSummarys(xs)) != 3 {
		t.Errorf("IndexGeoSummarys() = %v", IndexGeoSummarys(xs))
	}
	if err := ValidateGeoSummarys(ActiveGeoSummarys(xs)); err != nil {
		t.Errorf("ValidateGeoSummarys() = %v", err)
	}
}

func TestGeoSummaryNormalizeClamps(t *testing.T) {
	x := SampleGeoSummary("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoPolicyRoundTrip(t *testing.T) {
	x := SampleGeoPolicy("geopolicy-1", when)
	got, err := ParseGeoPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoPolicy() = %v", err)
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

func TestGeoPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleGeoPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoPolicy(f); err == nil {
		t.Error("ParseGeoPolicy() accepted a record with no id")
	}
	f = SampleGeoPolicy("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseGeoPolicy(f); err == nil {
		t.Error("ParseGeoPolicy() accepted a non-numeric version")
	}
}

func TestGeoPolicyCollections(t *testing.T) {
	xs := []GeoPolicy{SampleGeoPolicy("a", when), SampleGeoPolicy("bb", when), SampleGeoPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoPolicys(xs)); got != 2 {
		t.Errorf("ActiveGeoPolicys() kept %d, want 2", got)
	}
	if got, want := SumGeoPolicys(xs), 1+3; got != want {
		t.Errorf("SumGeoPolicys() = %d, want %d", got, want)
	}
	SortGeoPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoPolicys() put %q first", xs[0].ID)
	}
	if len(IndexGeoPolicys(xs)) != 3 {
		t.Errorf("IndexGeoPolicys() = %v", IndexGeoPolicys(xs))
	}
	if err := ValidateGeoPolicys(ActiveGeoPolicys(xs)); err != nil {
		t.Errorf("ValidateGeoPolicys() = %v", err)
	}
}

func TestGeoPolicyNormalizeClamps(t *testing.T) {
	x := SampleGeoPolicy("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoLinkRoundTrip(t *testing.T) {
	x := SampleGeoLink("geolink-1", when)
	got, err := ParseGeoLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoLink() = %v", err)
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

func TestGeoLinkParseRejectsBadInput(t *testing.T) {
	f := SampleGeoLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoLink(f); err == nil {
		t.Error("ParseGeoLink() accepted a record with no id")
	}
	f = SampleGeoLink("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseGeoLink(f); err == nil {
		t.Error("ParseGeoLink() accepted a non-numeric offset")
	}
}

func TestGeoLinkCollections(t *testing.T) {
	xs := []GeoLink{SampleGeoLink("a", when), SampleGeoLink("bb", when), SampleGeoLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoLinks(xs)); got != 2 {
		t.Errorf("ActiveGeoLinks() kept %d, want 2", got)
	}
	if got, want := SumGeoLinks(xs), 1+3; got != want {
		t.Errorf("SumGeoLinks() = %d, want %d", got, want)
	}
	SortGeoLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoLinks() put %q first", xs[0].ID)
	}
	if len(IndexGeoLinks(xs)) != 3 {
		t.Errorf("IndexGeoLinks() = %v", IndexGeoLinks(xs))
	}
	if err := ValidateGeoLinks(ActiveGeoLinks(xs)); err != nil {
		t.Errorf("ValidateGeoLinks() = %v", err)
	}
}

func TestGeoLinkNormalizeClamps(t *testing.T) {
	x := SampleGeoLink("a", when)
	x.Reference = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoNoteRoundTrip(t *testing.T) {
	x := SampleGeoNote("geonote-1", when)
	got, err := ParseGeoNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoNote() = %v", err)
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

func TestGeoNoteParseRejectsBadInput(t *testing.T) {
	f := SampleGeoNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoNote(f); err == nil {
		t.Error("ParseGeoNote() accepted a record with no id")
	}
	f = SampleGeoNote("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseGeoNote(f); err == nil {
		t.Error("ParseGeoNote() accepted a non-numeric version")
	}
}

func TestGeoNoteCollections(t *testing.T) {
	xs := []GeoNote{SampleGeoNote("a", when), SampleGeoNote("bb", when), SampleGeoNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoNotes(xs)); got != 2 {
		t.Errorf("ActiveGeoNotes() kept %d, want 2", got)
	}
	if got, want := SumGeoNotes(xs), 1+3; got != want {
		t.Errorf("SumGeoNotes() = %d, want %d", got, want)
	}
	SortGeoNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoNotes() put %q first", xs[0].ID)
	}
	if len(IndexGeoNotes(xs)) != 3 {
		t.Errorf("IndexGeoNotes() = %v", IndexGeoNotes(xs))
	}
	if err := ValidateGeoNotes(ActiveGeoNotes(xs)); err != nil {
		t.Errorf("ValidateGeoNotes() = %v", err)
	}
}

func TestGeoNoteNormalizeClamps(t *testing.T) {
	x := SampleGeoNote("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoWindowRoundTrip(t *testing.T) {
	x := SampleGeoWindow("geowindow-1", when)
	got, err := ParseGeoWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoWindow() = %v", err)
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

func TestGeoWindowParseRejectsBadInput(t *testing.T) {
	f := SampleGeoWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoWindow(f); err == nil {
		t.Error("ParseGeoWindow() accepted a record with no id")
	}
	f = SampleGeoWindow("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseGeoWindow(f); err == nil {
		t.Error("ParseGeoWindow() accepted a non-numeric capacity")
	}
}

func TestGeoWindowCollections(t *testing.T) {
	xs := []GeoWindow{SampleGeoWindow("a", when), SampleGeoWindow("bb", when), SampleGeoWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoWindows(xs)); got != 2 {
		t.Errorf("ActiveGeoWindows() kept %d, want 2", got)
	}
	if got, want := SumGeoWindows(xs), 1+3; got != want {
		t.Errorf("SumGeoWindows() = %d, want %d", got, want)
	}
	SortGeoWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoWindows() put %q first", xs[0].ID)
	}
	if len(IndexGeoWindows(xs)) != 3 {
		t.Errorf("IndexGeoWindows() = %v", IndexGeoWindows(xs))
	}
	if err := ValidateGeoWindows(ActiveGeoWindows(xs)); err != nil {
		t.Errorf("ValidateGeoWindows() = %v", err)
	}
}

func TestGeoWindowNormalizeClamps(t *testing.T) {
	x := SampleGeoWindow("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoQuotaRoundTrip(t *testing.T) {
	x := SampleGeoQuota("geoquota-1", when)
	got, err := ParseGeoQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoQuota() = %v", err)
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

func TestGeoQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleGeoQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoQuota(f); err == nil {
		t.Error("ParseGeoQuota() accepted a record with no id")
	}
	f = SampleGeoQuota("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseGeoQuota(f); err == nil {
		t.Error("ParseGeoQuota() accepted a non-numeric version")
	}
}

func TestGeoQuotaCollections(t *testing.T) {
	xs := []GeoQuota{SampleGeoQuota("a", when), SampleGeoQuota("bb", when), SampleGeoQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoQuotas(xs)); got != 2 {
		t.Errorf("ActiveGeoQuotas() kept %d, want 2", got)
	}
	if got, want := SumGeoQuotas(xs), 1+3; got != want {
		t.Errorf("SumGeoQuotas() = %d, want %d", got, want)
	}
	SortGeoQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoQuotas() put %q first", xs[0].ID)
	}
	if len(IndexGeoQuotas(xs)) != 3 {
		t.Errorf("IndexGeoQuotas() = %v", IndexGeoQuotas(xs))
	}
	if err := ValidateGeoQuotas(ActiveGeoQuotas(xs)); err != nil {
		t.Errorf("ValidateGeoQuotas() = %v", err)
	}
}

func TestGeoQuotaNormalizeClamps(t *testing.T) {
	x := SampleGeoQuota("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoBucketRoundTrip(t *testing.T) {
	x := SampleGeoBucket("geobucket-1", when)
	got, err := ParseGeoBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoBucket() = %v", err)
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

func TestGeoBucketParseRejectsBadInput(t *testing.T) {
	f := SampleGeoBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoBucket(f); err == nil {
		t.Error("ParseGeoBucket() accepted a record with no id")
	}
	f = SampleGeoBucket("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseGeoBucket(f); err == nil {
		t.Error("ParseGeoBucket() accepted a non-numeric count")
	}
}

func TestGeoBucketCollections(t *testing.T) {
	xs := []GeoBucket{SampleGeoBucket("a", when), SampleGeoBucket("bb", when), SampleGeoBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoBuckets(xs)); got != 2 {
		t.Errorf("ActiveGeoBuckets() kept %d, want 2", got)
	}
	if got, want := SumGeoBuckets(xs), 1+3; got != want {
		t.Errorf("SumGeoBuckets() = %d, want %d", got, want)
	}
	SortGeoBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoBuckets() put %q first", xs[0].ID)
	}
	if len(IndexGeoBuckets(xs)) != 3 {
		t.Errorf("IndexGeoBuckets() = %v", IndexGeoBuckets(xs))
	}
	if err := ValidateGeoBuckets(ActiveGeoBuckets(xs)); err != nil {
		t.Errorf("ValidateGeoBuckets() = %v", err)
	}
}

func TestGeoBucketNormalizeClamps(t *testing.T) {
	x := SampleGeoBucket("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoLaneRoundTrip(t *testing.T) {
	x := SampleGeoLane("geolane-1", when)
	got, err := ParseGeoLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoLane() = %v", err)
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

func TestGeoLaneParseRejectsBadInput(t *testing.T) {
	f := SampleGeoLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoLane(f); err == nil {
		t.Error("ParseGeoLane() accepted a record with no id")
	}
	f = SampleGeoLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseGeoLane(f); err == nil {
		t.Error("ParseGeoLane() accepted a non-numeric attempts")
	}
}

func TestGeoLaneCollections(t *testing.T) {
	xs := []GeoLane{SampleGeoLane("a", when), SampleGeoLane("bb", when), SampleGeoLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoLanes(xs)); got != 2 {
		t.Errorf("ActiveGeoLanes() kept %d, want 2", got)
	}
	if got, want := SumGeoLanes(xs), 1+3; got != want {
		t.Errorf("SumGeoLanes() = %d, want %d", got, want)
	}
	SortGeoLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoLanes() put %q first", xs[0].ID)
	}
	if len(IndexGeoLanes(xs)) != 3 {
		t.Errorf("IndexGeoLanes() = %v", IndexGeoLanes(xs))
	}
	if err := ValidateGeoLanes(ActiveGeoLanes(xs)); err != nil {
		t.Errorf("ValidateGeoLanes() = %v", err)
	}
}

func TestGeoLaneNormalizeClamps(t *testing.T) {
	x := SampleGeoLane("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoSlotRoundTrip(t *testing.T) {
	x := SampleGeoSlot("geoslot-1", when)
	got, err := ParseGeoSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoSlot() = %v", err)
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

func TestGeoSlotParseRejectsBadInput(t *testing.T) {
	f := SampleGeoSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoSlot(f); err == nil {
		t.Error("ParseGeoSlot() accepted a record with no id")
	}
	f = SampleGeoSlot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseGeoSlot(f); err == nil {
		t.Error("ParseGeoSlot() accepted a non-numeric offset")
	}
}

func TestGeoSlotCollections(t *testing.T) {
	xs := []GeoSlot{SampleGeoSlot("a", when), SampleGeoSlot("bb", when), SampleGeoSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoSlots(xs)); got != 2 {
		t.Errorf("ActiveGeoSlots() kept %d, want 2", got)
	}
	if got, want := SumGeoSlots(xs), 1+3; got != want {
		t.Errorf("SumGeoSlots() = %d, want %d", got, want)
	}
	SortGeoSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoSlots() put %q first", xs[0].ID)
	}
	if len(IndexGeoSlots(xs)) != 3 {
		t.Errorf("IndexGeoSlots() = %v", IndexGeoSlots(xs))
	}
	if err := ValidateGeoSlots(ActiveGeoSlots(xs)); err != nil {
		t.Errorf("ValidateGeoSlots() = %v", err)
	}
}

func TestGeoSlotNormalizeClamps(t *testing.T) {
	x := SampleGeoSlot("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoTagRoundTrip(t *testing.T) {
	x := SampleGeoTag("geotag-1", when)
	got, err := ParseGeoTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoTag() = %v", err)
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

func TestGeoTagParseRejectsBadInput(t *testing.T) {
	f := SampleGeoTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoTag(f); err == nil {
		t.Error("ParseGeoTag() accepted a record with no id")
	}
	f = SampleGeoTag("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseGeoTag(f); err == nil {
		t.Error("ParseGeoTag() accepted a non-numeric weight")
	}
}

func TestGeoTagCollections(t *testing.T) {
	xs := []GeoTag{SampleGeoTag("a", when), SampleGeoTag("bb", when), SampleGeoTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoTags(xs)); got != 2 {
		t.Errorf("ActiveGeoTags() kept %d, want 2", got)
	}
	if got, want := SumGeoTags(xs), 1+3; got != want {
		t.Errorf("SumGeoTags() = %d, want %d", got, want)
	}
	SortGeoTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoTags() put %q first", xs[0].ID)
	}
	if len(IndexGeoTags(xs)) != 3 {
		t.Errorf("IndexGeoTags() = %v", IndexGeoTags(xs))
	}
	if err := ValidateGeoTags(ActiveGeoTags(xs)); err != nil {
		t.Errorf("ValidateGeoTags() = %v", err)
	}
}

func TestGeoTagNormalizeClamps(t *testing.T) {
	x := SampleGeoTag("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoRuleRoundTrip(t *testing.T) {
	x := SampleGeoRule("georule-1", when)
	got, err := ParseGeoRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoRule() = %v", err)
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

func TestGeoRuleParseRejectsBadInput(t *testing.T) {
	f := SampleGeoRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoRule(f); err == nil {
		t.Error("ParseGeoRule() accepted a record with no id")
	}
	f = SampleGeoRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseGeoRule(f); err == nil {
		t.Error("ParseGeoRule() accepted a non-numeric offset")
	}
}

func TestGeoRuleCollections(t *testing.T) {
	xs := []GeoRule{SampleGeoRule("a", when), SampleGeoRule("bb", when), SampleGeoRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoRules(xs)); got != 2 {
		t.Errorf("ActiveGeoRules() kept %d, want 2", got)
	}
	if got, want := SumGeoRules(xs), 1+3; got != want {
		t.Errorf("SumGeoRules() = %d, want %d", got, want)
	}
	SortGeoRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoRules() put %q first", xs[0].ID)
	}
	if len(IndexGeoRules(xs)) != 3 {
		t.Errorf("IndexGeoRules() = %v", IndexGeoRules(xs))
	}
	if err := ValidateGeoRules(ActiveGeoRules(xs)); err != nil {
		t.Errorf("ValidateGeoRules() = %v", err)
	}
}

func TestGeoRuleNormalizeClamps(t *testing.T) {
	x := SampleGeoRule("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoHoldRoundTrip(t *testing.T) {
	x := SampleGeoHold("geohold-1", when)
	got, err := ParseGeoHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoHold() = %v", err)
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

func TestGeoHoldParseRejectsBadInput(t *testing.T) {
	f := SampleGeoHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoHold(f); err == nil {
		t.Error("ParseGeoHold() accepted a record with no id")
	}
	f = SampleGeoHold("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseGeoHold(f); err == nil {
		t.Error("ParseGeoHold() accepted a non-numeric capacity")
	}
}

func TestGeoHoldCollections(t *testing.T) {
	xs := []GeoHold{SampleGeoHold("a", when), SampleGeoHold("bb", when), SampleGeoHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoHolds(xs)); got != 2 {
		t.Errorf("ActiveGeoHolds() kept %d, want 2", got)
	}
	if got, want := SumGeoHolds(xs), 1+3; got != want {
		t.Errorf("SumGeoHolds() = %d, want %d", got, want)
	}
	SortGeoHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoHolds() put %q first", xs[0].ID)
	}
	if len(IndexGeoHolds(xs)) != 3 {
		t.Errorf("IndexGeoHolds() = %v", IndexGeoHolds(xs))
	}
	if err := ValidateGeoHolds(ActiveGeoHolds(xs)); err != nil {
		t.Errorf("ValidateGeoHolds() = %v", err)
	}
}

func TestGeoHoldNormalizeClamps(t *testing.T) {
	x := SampleGeoHold("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoClaimRoundTrip(t *testing.T) {
	x := SampleGeoClaim("geoclaim-1", when)
	got, err := ParseGeoClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoClaim() = %v", err)
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

func TestGeoClaimParseRejectsBadInput(t *testing.T) {
	f := SampleGeoClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoClaim(f); err == nil {
		t.Error("ParseGeoClaim() accepted a record with no id")
	}
	f = SampleGeoClaim("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseGeoClaim(f); err == nil {
		t.Error("ParseGeoClaim() accepted a non-numeric version")
	}
}

func TestGeoClaimCollections(t *testing.T) {
	xs := []GeoClaim{SampleGeoClaim("a", when), SampleGeoClaim("bb", when), SampleGeoClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoClaims(xs)); got != 2 {
		t.Errorf("ActiveGeoClaims() kept %d, want 2", got)
	}
	if got, want := SumGeoClaims(xs), 1+3; got != want {
		t.Errorf("SumGeoClaims() = %d, want %d", got, want)
	}
	SortGeoClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoClaims() put %q first", xs[0].ID)
	}
	if len(IndexGeoClaims(xs)) != 3 {
		t.Errorf("IndexGeoClaims() = %v", IndexGeoClaims(xs))
	}
	if err := ValidateGeoClaims(ActiveGeoClaims(xs)); err != nil {
		t.Errorf("ValidateGeoClaims() = %v", err)
	}
}

func TestGeoClaimNormalizeClamps(t *testing.T) {
	x := SampleGeoClaim("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestGeoCursorRoundTrip(t *testing.T) {
	x := SampleGeoCursor("geocursor-1", when)
	got, err := ParseGeoCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseGeoCursor() = %v", err)
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

func TestGeoCursorParseRejectsBadInput(t *testing.T) {
	f := SampleGeoCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseGeoCursor(f); err == nil {
		t.Error("ParseGeoCursor() accepted a record with no id")
	}
	f = SampleGeoCursor("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseGeoCursor(f); err == nil {
		t.Error("ParseGeoCursor() accepted a non-numeric count")
	}
}

func TestGeoCursorCollections(t *testing.T) {
	xs := []GeoCursor{SampleGeoCursor("a", when), SampleGeoCursor("bb", when), SampleGeoCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveGeoCursors(xs)); got != 2 {
		t.Errorf("ActiveGeoCursors() kept %d, want 2", got)
	}
	if got, want := SumGeoCursors(xs), 1+3; got != want {
		t.Errorf("SumGeoCursors() = %d, want %d", got, want)
	}
	SortGeoCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortGeoCursors() put %q first", xs[0].ID)
	}
	if len(IndexGeoCursors(xs)) != 3 {
		t.Errorf("IndexGeoCursors() = %v", IndexGeoCursors(xs))
	}
	if err := ValidateGeoCursors(ActiveGeoCursors(xs)); err != nil {
		t.Errorf("ValidateGeoCursors() = %v", err)
	}
}

func TestGeoCursorNormalizeClamps(t *testing.T) {
	x := SampleGeoCursor("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
