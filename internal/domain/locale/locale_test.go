package locale

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
		if !strings.HasPrefix(Prefixes[i], "locale/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleLocaleRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
	}
}

func TestLocaleRecordRoundTrip(t *testing.T) {
	x := SampleLocaleRecord("localerecord-1", when)
	got, err := ParseLocaleRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleRecord() = %v", err)
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

func TestLocaleRecordParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleRecord(f); err == nil {
		t.Error("ParseLocaleRecord() accepted a record with no id")
	}
	f = SampleLocaleRecord("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLocaleRecord(f); err == nil {
		t.Error("ParseLocaleRecord() accepted a non-numeric attempts")
	}
}

func TestLocaleRecordCollections(t *testing.T) {
	xs := []LocaleRecord{SampleLocaleRecord("a", when), SampleLocaleRecord("bb", when), SampleLocaleRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleRecords(xs)); got != 2 {
		t.Errorf("ActiveLocaleRecords() kept %d, want 2", got)
	}
	if got, want := SumLocaleRecords(xs), 1+3; got != want {
		t.Errorf("SumLocaleRecords() = %d, want %d", got, want)
	}
	SortLocaleRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleRecords() put %q first", xs[0].ID)
	}
	if len(IndexLocaleRecords(xs)) != 3 {
		t.Errorf("IndexLocaleRecords() = %v", IndexLocaleRecords(xs))
	}
	if err := ValidateLocaleRecords(ActiveLocaleRecords(xs)); err != nil {
		t.Errorf("ValidateLocaleRecords() = %v", err)
	}
}

func TestLocaleRecordNormalizeClamps(t *testing.T) {
	x := SampleLocaleRecord("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleEntryRoundTrip(t *testing.T) {
	x := SampleLocaleEntry("localeentry-1", when)
	got, err := ParseLocaleEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleEntry() = %v", err)
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

func TestLocaleEntryParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleEntry(f); err == nil {
		t.Error("ParseLocaleEntry() accepted a record with no id")
	}
	f = SampleLocaleEntry("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseLocaleEntry(f); err == nil {
		t.Error("ParseLocaleEntry() accepted a non-numeric capacity")
	}
}

func TestLocaleEntryCollections(t *testing.T) {
	xs := []LocaleEntry{SampleLocaleEntry("a", when), SampleLocaleEntry("bb", when), SampleLocaleEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleEntrys(xs)); got != 2 {
		t.Errorf("ActiveLocaleEntrys() kept %d, want 2", got)
	}
	if got, want := SumLocaleEntrys(xs), 1+3; got != want {
		t.Errorf("SumLocaleEntrys() = %d, want %d", got, want)
	}
	SortLocaleEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleEntrys() put %q first", xs[0].ID)
	}
	if len(IndexLocaleEntrys(xs)) != 3 {
		t.Errorf("IndexLocaleEntrys() = %v", IndexLocaleEntrys(xs))
	}
	if err := ValidateLocaleEntrys(ActiveLocaleEntrys(xs)); err != nil {
		t.Errorf("ValidateLocaleEntrys() = %v", err)
	}
}

func TestLocaleEntryNormalizeClamps(t *testing.T) {
	x := SampleLocaleEntry("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleItemRoundTrip(t *testing.T) {
	x := SampleLocaleItem("localeitem-1", when)
	got, err := ParseLocaleItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleItem() = %v", err)
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

func TestLocaleItemParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleItem(f); err == nil {
		t.Error("ParseLocaleItem() accepted a record with no id")
	}
	f = SampleLocaleItem("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseLocaleItem(f); err == nil {
		t.Error("ParseLocaleItem() accepted a non-numeric capacity")
	}
}

func TestLocaleItemCollections(t *testing.T) {
	xs := []LocaleItem{SampleLocaleItem("a", when), SampleLocaleItem("bb", when), SampleLocaleItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleItems(xs)); got != 2 {
		t.Errorf("ActiveLocaleItems() kept %d, want 2", got)
	}
	if got, want := SumLocaleItems(xs), 1+3; got != want {
		t.Errorf("SumLocaleItems() = %d, want %d", got, want)
	}
	SortLocaleItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleItems() put %q first", xs[0].ID)
	}
	if len(IndexLocaleItems(xs)) != 3 {
		t.Errorf("IndexLocaleItems() = %v", IndexLocaleItems(xs))
	}
	if err := ValidateLocaleItems(ActiveLocaleItems(xs)); err != nil {
		t.Errorf("ValidateLocaleItems() = %v", err)
	}
}

func TestLocaleItemNormalizeClamps(t *testing.T) {
	x := SampleLocaleItem("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleDraftRoundTrip(t *testing.T) {
	x := SampleLocaleDraft("localedraft-1", when)
	got, err := ParseLocaleDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleDraft() = %v", err)
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

func TestLocaleDraftParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleDraft(f); err == nil {
		t.Error("ParseLocaleDraft() accepted a record with no id")
	}
	f = SampleLocaleDraft("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseLocaleDraft(f); err == nil {
		t.Error("ParseLocaleDraft() accepted a non-numeric offset")
	}
}

func TestLocaleDraftCollections(t *testing.T) {
	xs := []LocaleDraft{SampleLocaleDraft("a", when), SampleLocaleDraft("bb", when), SampleLocaleDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleDrafts(xs)); got != 2 {
		t.Errorf("ActiveLocaleDrafts() kept %d, want 2", got)
	}
	if got, want := SumLocaleDrafts(xs), 1+3; got != want {
		t.Errorf("SumLocaleDrafts() = %d, want %d", got, want)
	}
	SortLocaleDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleDrafts() put %q first", xs[0].ID)
	}
	if len(IndexLocaleDrafts(xs)) != 3 {
		t.Errorf("IndexLocaleDrafts() = %v", IndexLocaleDrafts(xs))
	}
	if err := ValidateLocaleDrafts(ActiveLocaleDrafts(xs)); err != nil {
		t.Errorf("ValidateLocaleDrafts() = %v", err)
	}
}

func TestLocaleDraftNormalizeClamps(t *testing.T) {
	x := SampleLocaleDraft("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleRevisionRoundTrip(t *testing.T) {
	x := SampleLocaleRevision("localerevision-1", when)
	got, err := ParseLocaleRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleRevision() = %v", err)
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

func TestLocaleRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleRevision(f); err == nil {
		t.Error("ParseLocaleRevision() accepted a record with no id")
	}
	f = SampleLocaleRevision("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLocaleRevision(f); err == nil {
		t.Error("ParseLocaleRevision() accepted a non-numeric attempts")
	}
}

func TestLocaleRevisionCollections(t *testing.T) {
	xs := []LocaleRevision{SampleLocaleRevision("a", when), SampleLocaleRevision("bb", when), SampleLocaleRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleRevisions(xs)); got != 2 {
		t.Errorf("ActiveLocaleRevisions() kept %d, want 2", got)
	}
	if got, want := SumLocaleRevisions(xs), 1+3; got != want {
		t.Errorf("SumLocaleRevisions() = %d, want %d", got, want)
	}
	SortLocaleRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleRevisions() put %q first", xs[0].ID)
	}
	if len(IndexLocaleRevisions(xs)) != 3 {
		t.Errorf("IndexLocaleRevisions() = %v", IndexLocaleRevisions(xs))
	}
	if err := ValidateLocaleRevisions(ActiveLocaleRevisions(xs)); err != nil {
		t.Errorf("ValidateLocaleRevisions() = %v", err)
	}
}

func TestLocaleRevisionNormalizeClamps(t *testing.T) {
	x := SampleLocaleRevision("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleSnapshotRoundTrip(t *testing.T) {
	x := SampleLocaleSnapshot("localesnapshot-1", when)
	got, err := ParseLocaleSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleSnapshot() = %v", err)
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

func TestLocaleSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleSnapshot(f); err == nil {
		t.Error("ParseLocaleSnapshot() accepted a record with no id")
	}
	f = SampleLocaleSnapshot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLocaleSnapshot(f); err == nil {
		t.Error("ParseLocaleSnapshot() accepted a non-numeric retained")
	}
}

func TestLocaleSnapshotCollections(t *testing.T) {
	xs := []LocaleSnapshot{SampleLocaleSnapshot("a", when), SampleLocaleSnapshot("bb", when), SampleLocaleSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleSnapshots(xs)); got != 2 {
		t.Errorf("ActiveLocaleSnapshots() kept %d, want 2", got)
	}
	if got, want := SumLocaleSnapshots(xs), 1+3; got != want {
		t.Errorf("SumLocaleSnapshots() = %d, want %d", got, want)
	}
	SortLocaleSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexLocaleSnapshots(xs)) != 3 {
		t.Errorf("IndexLocaleSnapshots() = %v", IndexLocaleSnapshots(xs))
	}
	if err := ValidateLocaleSnapshots(ActiveLocaleSnapshots(xs)); err != nil {
		t.Errorf("ValidateLocaleSnapshots() = %v", err)
	}
}

func TestLocaleSnapshotNormalizeClamps(t *testing.T) {
	x := SampleLocaleSnapshot("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleBatchRoundTrip(t *testing.T) {
	x := SampleLocaleBatch("localebatch-1", when)
	got, err := ParseLocaleBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleBatch() = %v", err)
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

func TestLocaleBatchParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleBatch(f); err == nil {
		t.Error("ParseLocaleBatch() accepted a record with no id")
	}
	f = SampleLocaleBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseLocaleBatch(f); err == nil {
		t.Error("ParseLocaleBatch() accepted a non-numeric capacity")
	}
}

func TestLocaleBatchCollections(t *testing.T) {
	xs := []LocaleBatch{SampleLocaleBatch("a", when), SampleLocaleBatch("bb", when), SampleLocaleBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleBatchs(xs)); got != 2 {
		t.Errorf("ActiveLocaleBatchs() kept %d, want 2", got)
	}
	if got, want := SumLocaleBatchs(xs), 1+3; got != want {
		t.Errorf("SumLocaleBatchs() = %d, want %d", got, want)
	}
	SortLocaleBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleBatchs() put %q first", xs[0].ID)
	}
	if len(IndexLocaleBatchs(xs)) != 3 {
		t.Errorf("IndexLocaleBatchs() = %v", IndexLocaleBatchs(xs))
	}
	if err := ValidateLocaleBatchs(ActiveLocaleBatchs(xs)); err != nil {
		t.Errorf("ValidateLocaleBatchs() = %v", err)
	}
}

func TestLocaleBatchNormalizeClamps(t *testing.T) {
	x := SampleLocaleBatch("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleRequestRoundTrip(t *testing.T) {
	x := SampleLocaleRequest("localerequest-1", when)
	got, err := ParseLocaleRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleRequest() = %v", err)
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

func TestLocaleRequestParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleRequest(f); err == nil {
		t.Error("ParseLocaleRequest() accepted a record with no id")
	}
	f = SampleLocaleRequest("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseLocaleRequest(f); err == nil {
		t.Error("ParseLocaleRequest() accepted a non-numeric offset")
	}
}

func TestLocaleRequestCollections(t *testing.T) {
	xs := []LocaleRequest{SampleLocaleRequest("a", when), SampleLocaleRequest("bb", when), SampleLocaleRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleRequests(xs)); got != 2 {
		t.Errorf("ActiveLocaleRequests() kept %d, want 2", got)
	}
	if got, want := SumLocaleRequests(xs), 1+3; got != want {
		t.Errorf("SumLocaleRequests() = %d, want %d", got, want)
	}
	SortLocaleRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleRequests() put %q first", xs[0].ID)
	}
	if len(IndexLocaleRequests(xs)) != 3 {
		t.Errorf("IndexLocaleRequests() = %v", IndexLocaleRequests(xs))
	}
	if err := ValidateLocaleRequests(ActiveLocaleRequests(xs)); err != nil {
		t.Errorf("ValidateLocaleRequests() = %v", err)
	}
}

func TestLocaleRequestNormalizeClamps(t *testing.T) {
	x := SampleLocaleRequest("a", when)
	x.Source = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleSummaryRoundTrip(t *testing.T) {
	x := SampleLocaleSummary("localesummary-1", when)
	got, err := ParseLocaleSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleSummary() = %v", err)
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

func TestLocaleSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleSummary(f); err == nil {
		t.Error("ParseLocaleSummary() accepted a record with no id")
	}
	f = SampleLocaleSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseLocaleSummary(f); err == nil {
		t.Error("ParseLocaleSummary() accepted a non-numeric capacity")
	}
}

func TestLocaleSummaryCollections(t *testing.T) {
	xs := []LocaleSummary{SampleLocaleSummary("a", when), SampleLocaleSummary("bb", when), SampleLocaleSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleSummarys(xs)); got != 2 {
		t.Errorf("ActiveLocaleSummarys() kept %d, want 2", got)
	}
	if got, want := SumLocaleSummarys(xs), 1+3; got != want {
		t.Errorf("SumLocaleSummarys() = %d, want %d", got, want)
	}
	SortLocaleSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleSummarys() put %q first", xs[0].ID)
	}
	if len(IndexLocaleSummarys(xs)) != 3 {
		t.Errorf("IndexLocaleSummarys() = %v", IndexLocaleSummarys(xs))
	}
	if err := ValidateLocaleSummarys(ActiveLocaleSummarys(xs)); err != nil {
		t.Errorf("ValidateLocaleSummarys() = %v", err)
	}
}

func TestLocaleSummaryNormalizeClamps(t *testing.T) {
	x := SampleLocaleSummary("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocalePolicyRoundTrip(t *testing.T) {
	x := SampleLocalePolicy("localepolicy-1", when)
	got, err := ParseLocalePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocalePolicy() = %v", err)
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

func TestLocalePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleLocalePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocalePolicy(f); err == nil {
		t.Error("ParseLocalePolicy() accepted a record with no id")
	}
	f = SampleLocalePolicy("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseLocalePolicy(f); err == nil {
		t.Error("ParseLocalePolicy() accepted a non-numeric offset")
	}
}

func TestLocalePolicyCollections(t *testing.T) {
	xs := []LocalePolicy{SampleLocalePolicy("a", when), SampleLocalePolicy("bb", when), SampleLocalePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocalePolicys(xs)); got != 2 {
		t.Errorf("ActiveLocalePolicys() kept %d, want 2", got)
	}
	if got, want := SumLocalePolicys(xs), 1+3; got != want {
		t.Errorf("SumLocalePolicys() = %d, want %d", got, want)
	}
	SortLocalePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocalePolicys() put %q first", xs[0].ID)
	}
	if len(IndexLocalePolicys(xs)) != 3 {
		t.Errorf("IndexLocalePolicys() = %v", IndexLocalePolicys(xs))
	}
	if err := ValidateLocalePolicys(ActiveLocalePolicys(xs)); err != nil {
		t.Errorf("ValidateLocalePolicys() = %v", err)
	}
}

func TestLocalePolicyNormalizeClamps(t *testing.T) {
	x := SampleLocalePolicy("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleLinkRoundTrip(t *testing.T) {
	x := SampleLocaleLink("localelink-1", when)
	got, err := ParseLocaleLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleLink() = %v", err)
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

func TestLocaleLinkParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleLink(f); err == nil {
		t.Error("ParseLocaleLink() accepted a record with no id")
	}
	f = SampleLocaleLink("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseLocaleLink(f); err == nil {
		t.Error("ParseLocaleLink() accepted a non-numeric offset")
	}
}

func TestLocaleLinkCollections(t *testing.T) {
	xs := []LocaleLink{SampleLocaleLink("a", when), SampleLocaleLink("bb", when), SampleLocaleLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleLinks(xs)); got != 2 {
		t.Errorf("ActiveLocaleLinks() kept %d, want 2", got)
	}
	if got, want := SumLocaleLinks(xs), 1+3; got != want {
		t.Errorf("SumLocaleLinks() = %d, want %d", got, want)
	}
	SortLocaleLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleLinks() put %q first", xs[0].ID)
	}
	if len(IndexLocaleLinks(xs)) != 3 {
		t.Errorf("IndexLocaleLinks() = %v", IndexLocaleLinks(xs))
	}
	if err := ValidateLocaleLinks(ActiveLocaleLinks(xs)); err != nil {
		t.Errorf("ValidateLocaleLinks() = %v", err)
	}
}

func TestLocaleLinkNormalizeClamps(t *testing.T) {
	x := SampleLocaleLink("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleNoteRoundTrip(t *testing.T) {
	x := SampleLocaleNote("localenote-1", when)
	got, err := ParseLocaleNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleNote() = %v", err)
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

func TestLocaleNoteParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleNote(f); err == nil {
		t.Error("ParseLocaleNote() accepted a record with no id")
	}
	f = SampleLocaleNote("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseLocaleNote(f); err == nil {
		t.Error("ParseLocaleNote() accepted a non-numeric priority")
	}
}

func TestLocaleNoteCollections(t *testing.T) {
	xs := []LocaleNote{SampleLocaleNote("a", when), SampleLocaleNote("bb", when), SampleLocaleNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleNotes(xs)); got != 2 {
		t.Errorf("ActiveLocaleNotes() kept %d, want 2", got)
	}
	if got, want := SumLocaleNotes(xs), 1+3; got != want {
		t.Errorf("SumLocaleNotes() = %d, want %d", got, want)
	}
	SortLocaleNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleNotes() put %q first", xs[0].ID)
	}
	if len(IndexLocaleNotes(xs)) != 3 {
		t.Errorf("IndexLocaleNotes() = %v", IndexLocaleNotes(xs))
	}
	if err := ValidateLocaleNotes(ActiveLocaleNotes(xs)); err != nil {
		t.Errorf("ValidateLocaleNotes() = %v", err)
	}
}

func TestLocaleNoteNormalizeClamps(t *testing.T) {
	x := SampleLocaleNote("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleWindowRoundTrip(t *testing.T) {
	x := SampleLocaleWindow("localewindow-1", when)
	got, err := ParseLocaleWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleWindow() = %v", err)
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

func TestLocaleWindowParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleWindow(f); err == nil {
		t.Error("ParseLocaleWindow() accepted a record with no id")
	}
	f = SampleLocaleWindow("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLocaleWindow(f); err == nil {
		t.Error("ParseLocaleWindow() accepted a non-numeric retained")
	}
}

func TestLocaleWindowCollections(t *testing.T) {
	xs := []LocaleWindow{SampleLocaleWindow("a", when), SampleLocaleWindow("bb", when), SampleLocaleWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleWindows(xs)); got != 2 {
		t.Errorf("ActiveLocaleWindows() kept %d, want 2", got)
	}
	if got, want := SumLocaleWindows(xs), 1+3; got != want {
		t.Errorf("SumLocaleWindows() = %d, want %d", got, want)
	}
	SortLocaleWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleWindows() put %q first", xs[0].ID)
	}
	if len(IndexLocaleWindows(xs)) != 3 {
		t.Errorf("IndexLocaleWindows() = %v", IndexLocaleWindows(xs))
	}
	if err := ValidateLocaleWindows(ActiveLocaleWindows(xs)); err != nil {
		t.Errorf("ValidateLocaleWindows() = %v", err)
	}
}

func TestLocaleWindowNormalizeClamps(t *testing.T) {
	x := SampleLocaleWindow("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleQuotaRoundTrip(t *testing.T) {
	x := SampleLocaleQuota("localequota-1", when)
	got, err := ParseLocaleQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleQuota() = %v", err)
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

func TestLocaleQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleQuota(f); err == nil {
		t.Error("ParseLocaleQuota() accepted a record with no id")
	}
	f = SampleLocaleQuota("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLocaleQuota(f); err == nil {
		t.Error("ParseLocaleQuota() accepted a non-numeric attempts")
	}
}

func TestLocaleQuotaCollections(t *testing.T) {
	xs := []LocaleQuota{SampleLocaleQuota("a", when), SampleLocaleQuota("bb", when), SampleLocaleQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleQuotas(xs)); got != 2 {
		t.Errorf("ActiveLocaleQuotas() kept %d, want 2", got)
	}
	if got, want := SumLocaleQuotas(xs), 1+3; got != want {
		t.Errorf("SumLocaleQuotas() = %d, want %d", got, want)
	}
	SortLocaleQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleQuotas() put %q first", xs[0].ID)
	}
	if len(IndexLocaleQuotas(xs)) != 3 {
		t.Errorf("IndexLocaleQuotas() = %v", IndexLocaleQuotas(xs))
	}
	if err := ValidateLocaleQuotas(ActiveLocaleQuotas(xs)); err != nil {
		t.Errorf("ValidateLocaleQuotas() = %v", err)
	}
}

func TestLocaleQuotaNormalizeClamps(t *testing.T) {
	x := SampleLocaleQuota("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleBucketRoundTrip(t *testing.T) {
	x := SampleLocaleBucket("localebucket-1", when)
	got, err := ParseLocaleBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleBucket() = %v", err)
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

func TestLocaleBucketParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleBucket(f); err == nil {
		t.Error("ParseLocaleBucket() accepted a record with no id")
	}
	f = SampleLocaleBucket("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLocaleBucket(f); err == nil {
		t.Error("ParseLocaleBucket() accepted a non-numeric attempts")
	}
}

func TestLocaleBucketCollections(t *testing.T) {
	xs := []LocaleBucket{SampleLocaleBucket("a", when), SampleLocaleBucket("bb", when), SampleLocaleBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleBuckets(xs)); got != 2 {
		t.Errorf("ActiveLocaleBuckets() kept %d, want 2", got)
	}
	if got, want := SumLocaleBuckets(xs), 1+3; got != want {
		t.Errorf("SumLocaleBuckets() = %d, want %d", got, want)
	}
	SortLocaleBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleBuckets() put %q first", xs[0].ID)
	}
	if len(IndexLocaleBuckets(xs)) != 3 {
		t.Errorf("IndexLocaleBuckets() = %v", IndexLocaleBuckets(xs))
	}
	if err := ValidateLocaleBuckets(ActiveLocaleBuckets(xs)); err != nil {
		t.Errorf("ValidateLocaleBuckets() = %v", err)
	}
}

func TestLocaleBucketNormalizeClamps(t *testing.T) {
	x := SampleLocaleBucket("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleLaneRoundTrip(t *testing.T) {
	x := SampleLocaleLane("localelane-1", when)
	got, err := ParseLocaleLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleLane() = %v", err)
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

func TestLocaleLaneParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleLane(f); err == nil {
		t.Error("ParseLocaleLane() accepted a record with no id")
	}
	f = SampleLocaleLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseLocaleLane(f); err == nil {
		t.Error("ParseLocaleLane() accepted a non-numeric attempts")
	}
}

func TestLocaleLaneCollections(t *testing.T) {
	xs := []LocaleLane{SampleLocaleLane("a", when), SampleLocaleLane("bb", when), SampleLocaleLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleLanes(xs)); got != 2 {
		t.Errorf("ActiveLocaleLanes() kept %d, want 2", got)
	}
	if got, want := SumLocaleLanes(xs), 1+3; got != want {
		t.Errorf("SumLocaleLanes() = %d, want %d", got, want)
	}
	SortLocaleLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleLanes() put %q first", xs[0].ID)
	}
	if len(IndexLocaleLanes(xs)) != 3 {
		t.Errorf("IndexLocaleLanes() = %v", IndexLocaleLanes(xs))
	}
	if err := ValidateLocaleLanes(ActiveLocaleLanes(xs)); err != nil {
		t.Errorf("ValidateLocaleLanes() = %v", err)
	}
}

func TestLocaleLaneNormalizeClamps(t *testing.T) {
	x := SampleLocaleLane("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleSlotRoundTrip(t *testing.T) {
	x := SampleLocaleSlot("localeslot-1", when)
	got, err := ParseLocaleSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleSlot() = %v", err)
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

func TestLocaleSlotParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleSlot(f); err == nil {
		t.Error("ParseLocaleSlot() accepted a record with no id")
	}
	f = SampleLocaleSlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseLocaleSlot(f); err == nil {
		t.Error("ParseLocaleSlot() accepted a non-numeric count")
	}
}

func TestLocaleSlotCollections(t *testing.T) {
	xs := []LocaleSlot{SampleLocaleSlot("a", when), SampleLocaleSlot("bb", when), SampleLocaleSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleSlots(xs)); got != 2 {
		t.Errorf("ActiveLocaleSlots() kept %d, want 2", got)
	}
	if got, want := SumLocaleSlots(xs), 1+3; got != want {
		t.Errorf("SumLocaleSlots() = %d, want %d", got, want)
	}
	SortLocaleSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleSlots() put %q first", xs[0].ID)
	}
	if len(IndexLocaleSlots(xs)) != 3 {
		t.Errorf("IndexLocaleSlots() = %v", IndexLocaleSlots(xs))
	}
	if err := ValidateLocaleSlots(ActiveLocaleSlots(xs)); err != nil {
		t.Errorf("ValidateLocaleSlots() = %v", err)
	}
}

func TestLocaleSlotNormalizeClamps(t *testing.T) {
	x := SampleLocaleSlot("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleTagRoundTrip(t *testing.T) {
	x := SampleLocaleTag("localetag-1", when)
	got, err := ParseLocaleTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleTag() = %v", err)
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

func TestLocaleTagParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleTag(f); err == nil {
		t.Error("ParseLocaleTag() accepted a record with no id")
	}
	f = SampleLocaleTag("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseLocaleTag(f); err == nil {
		t.Error("ParseLocaleTag() accepted a non-numeric priority")
	}
}

func TestLocaleTagCollections(t *testing.T) {
	xs := []LocaleTag{SampleLocaleTag("a", when), SampleLocaleTag("bb", when), SampleLocaleTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleTags(xs)); got != 2 {
		t.Errorf("ActiveLocaleTags() kept %d, want 2", got)
	}
	if got, want := SumLocaleTags(xs), 1+3; got != want {
		t.Errorf("SumLocaleTags() = %d, want %d", got, want)
	}
	SortLocaleTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleTags() put %q first", xs[0].ID)
	}
	if len(IndexLocaleTags(xs)) != 3 {
		t.Errorf("IndexLocaleTags() = %v", IndexLocaleTags(xs))
	}
	if err := ValidateLocaleTags(ActiveLocaleTags(xs)); err != nil {
		t.Errorf("ValidateLocaleTags() = %v", err)
	}
}

func TestLocaleTagNormalizeClamps(t *testing.T) {
	x := SampleLocaleTag("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleRuleRoundTrip(t *testing.T) {
	x := SampleLocaleRule("localerule-1", when)
	got, err := ParseLocaleRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleRule() = %v", err)
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

func TestLocaleRuleParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleRule(f); err == nil {
		t.Error("ParseLocaleRule() accepted a record with no id")
	}
	f = SampleLocaleRule("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseLocaleRule(f); err == nil {
		t.Error("ParseLocaleRule() accepted a non-numeric version")
	}
}

func TestLocaleRuleCollections(t *testing.T) {
	xs := []LocaleRule{SampleLocaleRule("a", when), SampleLocaleRule("bb", when), SampleLocaleRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleRules(xs)); got != 2 {
		t.Errorf("ActiveLocaleRules() kept %d, want 2", got)
	}
	if got, want := SumLocaleRules(xs), 1+3; got != want {
		t.Errorf("SumLocaleRules() = %d, want %d", got, want)
	}
	SortLocaleRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleRules() put %q first", xs[0].ID)
	}
	if len(IndexLocaleRules(xs)) != 3 {
		t.Errorf("IndexLocaleRules() = %v", IndexLocaleRules(xs))
	}
	if err := ValidateLocaleRules(ActiveLocaleRules(xs)); err != nil {
		t.Errorf("ValidateLocaleRules() = %v", err)
	}
}

func TestLocaleRuleNormalizeClamps(t *testing.T) {
	x := SampleLocaleRule("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleHoldRoundTrip(t *testing.T) {
	x := SampleLocaleHold("localehold-1", when)
	got, err := ParseLocaleHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleHold() = %v", err)
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

func TestLocaleHoldParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleHold(f); err == nil {
		t.Error("ParseLocaleHold() accepted a record with no id")
	}
	f = SampleLocaleHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseLocaleHold(f); err == nil {
		t.Error("ParseLocaleHold() accepted a non-numeric weight")
	}
}

func TestLocaleHoldCollections(t *testing.T) {
	xs := []LocaleHold{SampleLocaleHold("a", when), SampleLocaleHold("bb", when), SampleLocaleHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleHolds(xs)); got != 2 {
		t.Errorf("ActiveLocaleHolds() kept %d, want 2", got)
	}
	if got, want := SumLocaleHolds(xs), 1+3; got != want {
		t.Errorf("SumLocaleHolds() = %d, want %d", got, want)
	}
	SortLocaleHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleHolds() put %q first", xs[0].ID)
	}
	if len(IndexLocaleHolds(xs)) != 3 {
		t.Errorf("IndexLocaleHolds() = %v", IndexLocaleHolds(xs))
	}
	if err := ValidateLocaleHolds(ActiveLocaleHolds(xs)); err != nil {
		t.Errorf("ValidateLocaleHolds() = %v", err)
	}
}

func TestLocaleHoldNormalizeClamps(t *testing.T) {
	x := SampleLocaleHold("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleClaimRoundTrip(t *testing.T) {
	x := SampleLocaleClaim("localeclaim-1", when)
	got, err := ParseLocaleClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleClaim() = %v", err)
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

func TestLocaleClaimParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleClaim(f); err == nil {
		t.Error("ParseLocaleClaim() accepted a record with no id")
	}
	f = SampleLocaleClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseLocaleClaim(f); err == nil {
		t.Error("ParseLocaleClaim() accepted a non-numeric retained")
	}
}

func TestLocaleClaimCollections(t *testing.T) {
	xs := []LocaleClaim{SampleLocaleClaim("a", when), SampleLocaleClaim("bb", when), SampleLocaleClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleClaims(xs)); got != 2 {
		t.Errorf("ActiveLocaleClaims() kept %d, want 2", got)
	}
	if got, want := SumLocaleClaims(xs), 1+3; got != want {
		t.Errorf("SumLocaleClaims() = %d, want %d", got, want)
	}
	SortLocaleClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleClaims() put %q first", xs[0].ID)
	}
	if len(IndexLocaleClaims(xs)) != 3 {
		t.Errorf("IndexLocaleClaims() = %v", IndexLocaleClaims(xs))
	}
	if err := ValidateLocaleClaims(ActiveLocaleClaims(xs)); err != nil {
		t.Errorf("ValidateLocaleClaims() = %v", err)
	}
}

func TestLocaleClaimNormalizeClamps(t *testing.T) {
	x := SampleLocaleClaim("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestLocaleCursorRoundTrip(t *testing.T) {
	x := SampleLocaleCursor("localecursor-1", when)
	got, err := ParseLocaleCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseLocaleCursor() = %v", err)
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

func TestLocaleCursorParseRejectsBadInput(t *testing.T) {
	f := SampleLocaleCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseLocaleCursor(f); err == nil {
		t.Error("ParseLocaleCursor() accepted a record with no id")
	}
	f = SampleLocaleCursor("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseLocaleCursor(f); err == nil {
		t.Error("ParseLocaleCursor() accepted a non-numeric offset")
	}
}

func TestLocaleCursorCollections(t *testing.T) {
	xs := []LocaleCursor{SampleLocaleCursor("a", when), SampleLocaleCursor("bb", when), SampleLocaleCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveLocaleCursors(xs)); got != 2 {
		t.Errorf("ActiveLocaleCursors() kept %d, want 2", got)
	}
	if got, want := SumLocaleCursors(xs), 1+3; got != want {
		t.Errorf("SumLocaleCursors() = %d, want %d", got, want)
	}
	SortLocaleCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortLocaleCursors() put %q first", xs[0].ID)
	}
	if len(IndexLocaleCursors(xs)) != 3 {
		t.Errorf("IndexLocaleCursors() = %v", IndexLocaleCursors(xs))
	}
	if err := ValidateLocaleCursors(ActiveLocaleCursors(xs)); err != nil {
		t.Errorf("ValidateLocaleCursors() = %v", err)
	}
}

func TestLocaleCursorNormalizeClamps(t *testing.T) {
	x := SampleLocaleCursor("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
