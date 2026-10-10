package invoice

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
		if !strings.HasPrefix(Prefixes[i], "invoice/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleInvoiceRecord("ok", when).Fields()); err != nil {
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

func TestInvoiceRecordRoundTrip(t *testing.T) {
	x := SampleInvoiceRecord("invoicerecord-1", when)
	got, err := ParseInvoiceRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceRecord() = %v", err)
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

func TestInvoiceRecordParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceRecord(f); err == nil {
		t.Error("ParseInvoiceRecord() accepted a record with no id")
	}
	f = SampleInvoiceRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseInvoiceRecord(f); err == nil {
		t.Error("ParseInvoiceRecord() accepted a non-numeric capacity")
	}
}

func TestInvoiceRecordCollections(t *testing.T) {
	xs := []InvoiceRecord{SampleInvoiceRecord("a", when), SampleInvoiceRecord("bb", when), SampleInvoiceRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceRecords(xs)); got != 2 {
		t.Errorf("ActiveInvoiceRecords() kept %d, want 2", got)
	}
	if got, want := SumInvoiceRecords(xs), 1+3; got != want {
		t.Errorf("SumInvoiceRecords() = %d, want %d", got, want)
	}
	SortInvoiceRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceRecords() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceRecords(xs)) != 3 {
		t.Errorf("IndexInvoiceRecords() = %v", IndexInvoiceRecords(xs))
	}
	if err := ValidateInvoiceRecords(ActiveInvoiceRecords(xs)); err != nil {
		t.Errorf("ValidateInvoiceRecords() = %v", err)
	}
}

func TestInvoiceRecordNormalizeClamps(t *testing.T) {
	x := SampleInvoiceRecord("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceEntryRoundTrip(t *testing.T) {
	x := SampleInvoiceEntry("invoiceentry-1", when)
	got, err := ParseInvoiceEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceEntry() = %v", err)
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

func TestInvoiceEntryParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceEntry(f); err == nil {
		t.Error("ParseInvoiceEntry() accepted a record with no id")
	}
	f = SampleInvoiceEntry("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseInvoiceEntry(f); err == nil {
		t.Error("ParseInvoiceEntry() accepted a non-numeric attempts")
	}
}

func TestInvoiceEntryCollections(t *testing.T) {
	xs := []InvoiceEntry{SampleInvoiceEntry("a", when), SampleInvoiceEntry("bb", when), SampleInvoiceEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceEntrys(xs)); got != 2 {
		t.Errorf("ActiveInvoiceEntrys() kept %d, want 2", got)
	}
	if got, want := SumInvoiceEntrys(xs), 1+3; got != want {
		t.Errorf("SumInvoiceEntrys() = %d, want %d", got, want)
	}
	SortInvoiceEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceEntrys() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceEntrys(xs)) != 3 {
		t.Errorf("IndexInvoiceEntrys() = %v", IndexInvoiceEntrys(xs))
	}
	if err := ValidateInvoiceEntrys(ActiveInvoiceEntrys(xs)); err != nil {
		t.Errorf("ValidateInvoiceEntrys() = %v", err)
	}
}

func TestInvoiceEntryNormalizeClamps(t *testing.T) {
	x := SampleInvoiceEntry("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceItemRoundTrip(t *testing.T) {
	x := SampleInvoiceItem("invoiceitem-1", when)
	got, err := ParseInvoiceItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceItem() = %v", err)
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

func TestInvoiceItemParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceItem(f); err == nil {
		t.Error("ParseInvoiceItem() accepted a record with no id")
	}
	f = SampleInvoiceItem("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseInvoiceItem(f); err == nil {
		t.Error("ParseInvoiceItem() accepted a non-numeric retained")
	}
}

func TestInvoiceItemCollections(t *testing.T) {
	xs := []InvoiceItem{SampleInvoiceItem("a", when), SampleInvoiceItem("bb", when), SampleInvoiceItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceItems(xs)); got != 2 {
		t.Errorf("ActiveInvoiceItems() kept %d, want 2", got)
	}
	if got, want := SumInvoiceItems(xs), 1+3; got != want {
		t.Errorf("SumInvoiceItems() = %d, want %d", got, want)
	}
	SortInvoiceItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceItems() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceItems(xs)) != 3 {
		t.Errorf("IndexInvoiceItems() = %v", IndexInvoiceItems(xs))
	}
	if err := ValidateInvoiceItems(ActiveInvoiceItems(xs)); err != nil {
		t.Errorf("ValidateInvoiceItems() = %v", err)
	}
}

func TestInvoiceItemNormalizeClamps(t *testing.T) {
	x := SampleInvoiceItem("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceDraftRoundTrip(t *testing.T) {
	x := SampleInvoiceDraft("invoicedraft-1", when)
	got, err := ParseInvoiceDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceDraft() = %v", err)
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

func TestInvoiceDraftParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceDraft(f); err == nil {
		t.Error("ParseInvoiceDraft() accepted a record with no id")
	}
	f = SampleInvoiceDraft("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseInvoiceDraft(f); err == nil {
		t.Error("ParseInvoiceDraft() accepted a non-numeric offset")
	}
}

func TestInvoiceDraftCollections(t *testing.T) {
	xs := []InvoiceDraft{SampleInvoiceDraft("a", when), SampleInvoiceDraft("bb", when), SampleInvoiceDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceDrafts(xs)); got != 2 {
		t.Errorf("ActiveInvoiceDrafts() kept %d, want 2", got)
	}
	if got, want := SumInvoiceDrafts(xs), 1+3; got != want {
		t.Errorf("SumInvoiceDrafts() = %d, want %d", got, want)
	}
	SortInvoiceDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceDrafts() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceDrafts(xs)) != 3 {
		t.Errorf("IndexInvoiceDrafts() = %v", IndexInvoiceDrafts(xs))
	}
	if err := ValidateInvoiceDrafts(ActiveInvoiceDrafts(xs)); err != nil {
		t.Errorf("ValidateInvoiceDrafts() = %v", err)
	}
}

func TestInvoiceDraftNormalizeClamps(t *testing.T) {
	x := SampleInvoiceDraft("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceRevisionRoundTrip(t *testing.T) {
	x := SampleInvoiceRevision("invoicerevision-1", when)
	got, err := ParseInvoiceRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceRevision() = %v", err)
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

func TestInvoiceRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceRevision(f); err == nil {
		t.Error("ParseInvoiceRevision() accepted a record with no id")
	}
	f = SampleInvoiceRevision("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseInvoiceRevision(f); err == nil {
		t.Error("ParseInvoiceRevision() accepted a non-numeric weight")
	}
}

func TestInvoiceRevisionCollections(t *testing.T) {
	xs := []InvoiceRevision{SampleInvoiceRevision("a", when), SampleInvoiceRevision("bb", when), SampleInvoiceRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceRevisions(xs)); got != 2 {
		t.Errorf("ActiveInvoiceRevisions() kept %d, want 2", got)
	}
	if got, want := SumInvoiceRevisions(xs), 1+3; got != want {
		t.Errorf("SumInvoiceRevisions() = %d, want %d", got, want)
	}
	SortInvoiceRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceRevisions() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceRevisions(xs)) != 3 {
		t.Errorf("IndexInvoiceRevisions() = %v", IndexInvoiceRevisions(xs))
	}
	if err := ValidateInvoiceRevisions(ActiveInvoiceRevisions(xs)); err != nil {
		t.Errorf("ValidateInvoiceRevisions() = %v", err)
	}
}

func TestInvoiceRevisionNormalizeClamps(t *testing.T) {
	x := SampleInvoiceRevision("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceSnapshotRoundTrip(t *testing.T) {
	x := SampleInvoiceSnapshot("invoicesnapshot-1", when)
	got, err := ParseInvoiceSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceSnapshot() = %v", err)
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

func TestInvoiceSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceSnapshot(f); err == nil {
		t.Error("ParseInvoiceSnapshot() accepted a record with no id")
	}
	f = SampleInvoiceSnapshot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseInvoiceSnapshot(f); err == nil {
		t.Error("ParseInvoiceSnapshot() accepted a non-numeric weight")
	}
}

func TestInvoiceSnapshotCollections(t *testing.T) {
	xs := []InvoiceSnapshot{SampleInvoiceSnapshot("a", when), SampleInvoiceSnapshot("bb", when), SampleInvoiceSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceSnapshots(xs)); got != 2 {
		t.Errorf("ActiveInvoiceSnapshots() kept %d, want 2", got)
	}
	if got, want := SumInvoiceSnapshots(xs), 1+3; got != want {
		t.Errorf("SumInvoiceSnapshots() = %d, want %d", got, want)
	}
	SortInvoiceSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceSnapshots(xs)) != 3 {
		t.Errorf("IndexInvoiceSnapshots() = %v", IndexInvoiceSnapshots(xs))
	}
	if err := ValidateInvoiceSnapshots(ActiveInvoiceSnapshots(xs)); err != nil {
		t.Errorf("ValidateInvoiceSnapshots() = %v", err)
	}
}

func TestInvoiceSnapshotNormalizeClamps(t *testing.T) {
	x := SampleInvoiceSnapshot("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceBatchRoundTrip(t *testing.T) {
	x := SampleInvoiceBatch("invoicebatch-1", when)
	got, err := ParseInvoiceBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceBatch() = %v", err)
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

func TestInvoiceBatchParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceBatch(f); err == nil {
		t.Error("ParseInvoiceBatch() accepted a record with no id")
	}
	f = SampleInvoiceBatch("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseInvoiceBatch(f); err == nil {
		t.Error("ParseInvoiceBatch() accepted a non-numeric attempts")
	}
}

func TestInvoiceBatchCollections(t *testing.T) {
	xs := []InvoiceBatch{SampleInvoiceBatch("a", when), SampleInvoiceBatch("bb", when), SampleInvoiceBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceBatchs(xs)); got != 2 {
		t.Errorf("ActiveInvoiceBatchs() kept %d, want 2", got)
	}
	if got, want := SumInvoiceBatchs(xs), 1+3; got != want {
		t.Errorf("SumInvoiceBatchs() = %d, want %d", got, want)
	}
	SortInvoiceBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceBatchs() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceBatchs(xs)) != 3 {
		t.Errorf("IndexInvoiceBatchs() = %v", IndexInvoiceBatchs(xs))
	}
	if err := ValidateInvoiceBatchs(ActiveInvoiceBatchs(xs)); err != nil {
		t.Errorf("ValidateInvoiceBatchs() = %v", err)
	}
}

func TestInvoiceBatchNormalizeClamps(t *testing.T) {
	x := SampleInvoiceBatch("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceRequestRoundTrip(t *testing.T) {
	x := SampleInvoiceRequest("invoicerequest-1", when)
	got, err := ParseInvoiceRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceRequest() = %v", err)
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

func TestInvoiceRequestParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceRequest(f); err == nil {
		t.Error("ParseInvoiceRequest() accepted a record with no id")
	}
	f = SampleInvoiceRequest("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInvoiceRequest(f); err == nil {
		t.Error("ParseInvoiceRequest() accepted a non-numeric priority")
	}
}

func TestInvoiceRequestCollections(t *testing.T) {
	xs := []InvoiceRequest{SampleInvoiceRequest("a", when), SampleInvoiceRequest("bb", when), SampleInvoiceRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceRequests(xs)); got != 2 {
		t.Errorf("ActiveInvoiceRequests() kept %d, want 2", got)
	}
	if got, want := SumInvoiceRequests(xs), 1+3; got != want {
		t.Errorf("SumInvoiceRequests() = %d, want %d", got, want)
	}
	SortInvoiceRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceRequests() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceRequests(xs)) != 3 {
		t.Errorf("IndexInvoiceRequests() = %v", IndexInvoiceRequests(xs))
	}
	if err := ValidateInvoiceRequests(ActiveInvoiceRequests(xs)); err != nil {
		t.Errorf("ValidateInvoiceRequests() = %v", err)
	}
}

func TestInvoiceRequestNormalizeClamps(t *testing.T) {
	x := SampleInvoiceRequest("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceSummaryRoundTrip(t *testing.T) {
	x := SampleInvoiceSummary("invoicesummary-1", when)
	got, err := ParseInvoiceSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceSummary() = %v", err)
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

func TestInvoiceSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceSummary(f); err == nil {
		t.Error("ParseInvoiceSummary() accepted a record with no id")
	}
	f = SampleInvoiceSummary("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseInvoiceSummary(f); err == nil {
		t.Error("ParseInvoiceSummary() accepted a non-numeric weight")
	}
}

func TestInvoiceSummaryCollections(t *testing.T) {
	xs := []InvoiceSummary{SampleInvoiceSummary("a", when), SampleInvoiceSummary("bb", when), SampleInvoiceSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceSummarys(xs)); got != 2 {
		t.Errorf("ActiveInvoiceSummarys() kept %d, want 2", got)
	}
	if got, want := SumInvoiceSummarys(xs), 1+3; got != want {
		t.Errorf("SumInvoiceSummarys() = %d, want %d", got, want)
	}
	SortInvoiceSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceSummarys() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceSummarys(xs)) != 3 {
		t.Errorf("IndexInvoiceSummarys() = %v", IndexInvoiceSummarys(xs))
	}
	if err := ValidateInvoiceSummarys(ActiveInvoiceSummarys(xs)); err != nil {
		t.Errorf("ValidateInvoiceSummarys() = %v", err)
	}
}

func TestInvoiceSummaryNormalizeClamps(t *testing.T) {
	x := SampleInvoiceSummary("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoicePolicyRoundTrip(t *testing.T) {
	x := SampleInvoicePolicy("invoicepolicy-1", when)
	got, err := ParseInvoicePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoicePolicy() = %v", err)
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

func TestInvoicePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleInvoicePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoicePolicy(f); err == nil {
		t.Error("ParseInvoicePolicy() accepted a record with no id")
	}
	f = SampleInvoicePolicy("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInvoicePolicy(f); err == nil {
		t.Error("ParseInvoicePolicy() accepted a non-numeric priority")
	}
}

func TestInvoicePolicyCollections(t *testing.T) {
	xs := []InvoicePolicy{SampleInvoicePolicy("a", when), SampleInvoicePolicy("bb", when), SampleInvoicePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoicePolicys(xs)); got != 2 {
		t.Errorf("ActiveInvoicePolicys() kept %d, want 2", got)
	}
	if got, want := SumInvoicePolicys(xs), 1+3; got != want {
		t.Errorf("SumInvoicePolicys() = %d, want %d", got, want)
	}
	SortInvoicePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoicePolicys() put %q first", xs[0].ID)
	}
	if len(IndexInvoicePolicys(xs)) != 3 {
		t.Errorf("IndexInvoicePolicys() = %v", IndexInvoicePolicys(xs))
	}
	if err := ValidateInvoicePolicys(ActiveInvoicePolicys(xs)); err != nil {
		t.Errorf("ValidateInvoicePolicys() = %v", err)
	}
}

func TestInvoicePolicyNormalizeClamps(t *testing.T) {
	x := SampleInvoicePolicy("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceLinkRoundTrip(t *testing.T) {
	x := SampleInvoiceLink("invoicelink-1", when)
	got, err := ParseInvoiceLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceLink() = %v", err)
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

func TestInvoiceLinkParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceLink(f); err == nil {
		t.Error("ParseInvoiceLink() accepted a record with no id")
	}
	f = SampleInvoiceLink("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInvoiceLink(f); err == nil {
		t.Error("ParseInvoiceLink() accepted a non-numeric priority")
	}
}

func TestInvoiceLinkCollections(t *testing.T) {
	xs := []InvoiceLink{SampleInvoiceLink("a", when), SampleInvoiceLink("bb", when), SampleInvoiceLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceLinks(xs)); got != 2 {
		t.Errorf("ActiveInvoiceLinks() kept %d, want 2", got)
	}
	if got, want := SumInvoiceLinks(xs), 1+3; got != want {
		t.Errorf("SumInvoiceLinks() = %d, want %d", got, want)
	}
	SortInvoiceLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceLinks() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceLinks(xs)) != 3 {
		t.Errorf("IndexInvoiceLinks() = %v", IndexInvoiceLinks(xs))
	}
	if err := ValidateInvoiceLinks(ActiveInvoiceLinks(xs)); err != nil {
		t.Errorf("ValidateInvoiceLinks() = %v", err)
	}
}

func TestInvoiceLinkNormalizeClamps(t *testing.T) {
	x := SampleInvoiceLink("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceNoteRoundTrip(t *testing.T) {
	x := SampleInvoiceNote("invoicenote-1", when)
	got, err := ParseInvoiceNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceNote() = %v", err)
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

func TestInvoiceNoteParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceNote(f); err == nil {
		t.Error("ParseInvoiceNote() accepted a record with no id")
	}
	f = SampleInvoiceNote("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseInvoiceNote(f); err == nil {
		t.Error("ParseInvoiceNote() accepted a non-numeric retained")
	}
}

func TestInvoiceNoteCollections(t *testing.T) {
	xs := []InvoiceNote{SampleInvoiceNote("a", when), SampleInvoiceNote("bb", when), SampleInvoiceNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceNotes(xs)); got != 2 {
		t.Errorf("ActiveInvoiceNotes() kept %d, want 2", got)
	}
	if got, want := SumInvoiceNotes(xs), 1+3; got != want {
		t.Errorf("SumInvoiceNotes() = %d, want %d", got, want)
	}
	SortInvoiceNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceNotes() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceNotes(xs)) != 3 {
		t.Errorf("IndexInvoiceNotes() = %v", IndexInvoiceNotes(xs))
	}
	if err := ValidateInvoiceNotes(ActiveInvoiceNotes(xs)); err != nil {
		t.Errorf("ValidateInvoiceNotes() = %v", err)
	}
}

func TestInvoiceNoteNormalizeClamps(t *testing.T) {
	x := SampleInvoiceNote("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceWindowRoundTrip(t *testing.T) {
	x := SampleInvoiceWindow("invoicewindow-1", when)
	got, err := ParseInvoiceWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceWindow() = %v", err)
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

func TestInvoiceWindowParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceWindow(f); err == nil {
		t.Error("ParseInvoiceWindow() accepted a record with no id")
	}
	f = SampleInvoiceWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInvoiceWindow(f); err == nil {
		t.Error("ParseInvoiceWindow() accepted a non-numeric priority")
	}
}

func TestInvoiceWindowCollections(t *testing.T) {
	xs := []InvoiceWindow{SampleInvoiceWindow("a", when), SampleInvoiceWindow("bb", when), SampleInvoiceWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceWindows(xs)); got != 2 {
		t.Errorf("ActiveInvoiceWindows() kept %d, want 2", got)
	}
	if got, want := SumInvoiceWindows(xs), 1+3; got != want {
		t.Errorf("SumInvoiceWindows() = %d, want %d", got, want)
	}
	SortInvoiceWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceWindows() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceWindows(xs)) != 3 {
		t.Errorf("IndexInvoiceWindows() = %v", IndexInvoiceWindows(xs))
	}
	if err := ValidateInvoiceWindows(ActiveInvoiceWindows(xs)); err != nil {
		t.Errorf("ValidateInvoiceWindows() = %v", err)
	}
}

func TestInvoiceWindowNormalizeClamps(t *testing.T) {
	x := SampleInvoiceWindow("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceQuotaRoundTrip(t *testing.T) {
	x := SampleInvoiceQuota("invoicequota-1", when)
	got, err := ParseInvoiceQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceQuota() = %v", err)
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

func TestInvoiceQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceQuota(f); err == nil {
		t.Error("ParseInvoiceQuota() accepted a record with no id")
	}
	f = SampleInvoiceQuota("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseInvoiceQuota(f); err == nil {
		t.Error("ParseInvoiceQuota() accepted a non-numeric weight")
	}
}

func TestInvoiceQuotaCollections(t *testing.T) {
	xs := []InvoiceQuota{SampleInvoiceQuota("a", when), SampleInvoiceQuota("bb", when), SampleInvoiceQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceQuotas(xs)); got != 2 {
		t.Errorf("ActiveInvoiceQuotas() kept %d, want 2", got)
	}
	if got, want := SumInvoiceQuotas(xs), 1+3; got != want {
		t.Errorf("SumInvoiceQuotas() = %d, want %d", got, want)
	}
	SortInvoiceQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceQuotas() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceQuotas(xs)) != 3 {
		t.Errorf("IndexInvoiceQuotas() = %v", IndexInvoiceQuotas(xs))
	}
	if err := ValidateInvoiceQuotas(ActiveInvoiceQuotas(xs)); err != nil {
		t.Errorf("ValidateInvoiceQuotas() = %v", err)
	}
}

func TestInvoiceQuotaNormalizeClamps(t *testing.T) {
	x := SampleInvoiceQuota("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceBucketRoundTrip(t *testing.T) {
	x := SampleInvoiceBucket("invoicebucket-1", when)
	got, err := ParseInvoiceBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceBucket() = %v", err)
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

func TestInvoiceBucketParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceBucket(f); err == nil {
		t.Error("ParseInvoiceBucket() accepted a record with no id")
	}
	f = SampleInvoiceBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseInvoiceBucket(f); err == nil {
		t.Error("ParseInvoiceBucket() accepted a non-numeric retained")
	}
}

func TestInvoiceBucketCollections(t *testing.T) {
	xs := []InvoiceBucket{SampleInvoiceBucket("a", when), SampleInvoiceBucket("bb", when), SampleInvoiceBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceBuckets(xs)); got != 2 {
		t.Errorf("ActiveInvoiceBuckets() kept %d, want 2", got)
	}
	if got, want := SumInvoiceBuckets(xs), 1+3; got != want {
		t.Errorf("SumInvoiceBuckets() = %d, want %d", got, want)
	}
	SortInvoiceBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceBuckets() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceBuckets(xs)) != 3 {
		t.Errorf("IndexInvoiceBuckets() = %v", IndexInvoiceBuckets(xs))
	}
	if err := ValidateInvoiceBuckets(ActiveInvoiceBuckets(xs)); err != nil {
		t.Errorf("ValidateInvoiceBuckets() = %v", err)
	}
}

func TestInvoiceBucketNormalizeClamps(t *testing.T) {
	x := SampleInvoiceBucket("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceLaneRoundTrip(t *testing.T) {
	x := SampleInvoiceLane("invoicelane-1", when)
	got, err := ParseInvoiceLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceLane() = %v", err)
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

func TestInvoiceLaneParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceLane(f); err == nil {
		t.Error("ParseInvoiceLane() accepted a record with no id")
	}
	f = SampleInvoiceLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInvoiceLane(f); err == nil {
		t.Error("ParseInvoiceLane() accepted a non-numeric priority")
	}
}

func TestInvoiceLaneCollections(t *testing.T) {
	xs := []InvoiceLane{SampleInvoiceLane("a", when), SampleInvoiceLane("bb", when), SampleInvoiceLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceLanes(xs)); got != 2 {
		t.Errorf("ActiveInvoiceLanes() kept %d, want 2", got)
	}
	if got, want := SumInvoiceLanes(xs), 1+3; got != want {
		t.Errorf("SumInvoiceLanes() = %d, want %d", got, want)
	}
	SortInvoiceLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceLanes() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceLanes(xs)) != 3 {
		t.Errorf("IndexInvoiceLanes() = %v", IndexInvoiceLanes(xs))
	}
	if err := ValidateInvoiceLanes(ActiveInvoiceLanes(xs)); err != nil {
		t.Errorf("ValidateInvoiceLanes() = %v", err)
	}
}

func TestInvoiceLaneNormalizeClamps(t *testing.T) {
	x := SampleInvoiceLane("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceSlotRoundTrip(t *testing.T) {
	x := SampleInvoiceSlot("invoiceslot-1", when)
	got, err := ParseInvoiceSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceSlot() = %v", err)
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

func TestInvoiceSlotParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceSlot(f); err == nil {
		t.Error("ParseInvoiceSlot() accepted a record with no id")
	}
	f = SampleInvoiceSlot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseInvoiceSlot(f); err == nil {
		t.Error("ParseInvoiceSlot() accepted a non-numeric retained")
	}
}

func TestInvoiceSlotCollections(t *testing.T) {
	xs := []InvoiceSlot{SampleInvoiceSlot("a", when), SampleInvoiceSlot("bb", when), SampleInvoiceSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceSlots(xs)); got != 2 {
		t.Errorf("ActiveInvoiceSlots() kept %d, want 2", got)
	}
	if got, want := SumInvoiceSlots(xs), 1+3; got != want {
		t.Errorf("SumInvoiceSlots() = %d, want %d", got, want)
	}
	SortInvoiceSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceSlots() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceSlots(xs)) != 3 {
		t.Errorf("IndexInvoiceSlots() = %v", IndexInvoiceSlots(xs))
	}
	if err := ValidateInvoiceSlots(ActiveInvoiceSlots(xs)); err != nil {
		t.Errorf("ValidateInvoiceSlots() = %v", err)
	}
}

func TestInvoiceSlotNormalizeClamps(t *testing.T) {
	x := SampleInvoiceSlot("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceTagRoundTrip(t *testing.T) {
	x := SampleInvoiceTag("invoicetag-1", when)
	got, err := ParseInvoiceTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceTag() = %v", err)
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

func TestInvoiceTagParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceTag(f); err == nil {
		t.Error("ParseInvoiceTag() accepted a record with no id")
	}
	f = SampleInvoiceTag("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseInvoiceTag(f); err == nil {
		t.Error("ParseInvoiceTag() accepted a non-numeric capacity")
	}
}

func TestInvoiceTagCollections(t *testing.T) {
	xs := []InvoiceTag{SampleInvoiceTag("a", when), SampleInvoiceTag("bb", when), SampleInvoiceTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceTags(xs)); got != 2 {
		t.Errorf("ActiveInvoiceTags() kept %d, want 2", got)
	}
	if got, want := SumInvoiceTags(xs), 1+3; got != want {
		t.Errorf("SumInvoiceTags() = %d, want %d", got, want)
	}
	SortInvoiceTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceTags() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceTags(xs)) != 3 {
		t.Errorf("IndexInvoiceTags() = %v", IndexInvoiceTags(xs))
	}
	if err := ValidateInvoiceTags(ActiveInvoiceTags(xs)); err != nil {
		t.Errorf("ValidateInvoiceTags() = %v", err)
	}
}

func TestInvoiceTagNormalizeClamps(t *testing.T) {
	x := SampleInvoiceTag("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceRuleRoundTrip(t *testing.T) {
	x := SampleInvoiceRule("invoicerule-1", when)
	got, err := ParseInvoiceRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceRule() = %v", err)
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

func TestInvoiceRuleParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceRule(f); err == nil {
		t.Error("ParseInvoiceRule() accepted a record with no id")
	}
	f = SampleInvoiceRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseInvoiceRule(f); err == nil {
		t.Error("ParseInvoiceRule() accepted a non-numeric offset")
	}
}

func TestInvoiceRuleCollections(t *testing.T) {
	xs := []InvoiceRule{SampleInvoiceRule("a", when), SampleInvoiceRule("bb", when), SampleInvoiceRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceRules(xs)); got != 2 {
		t.Errorf("ActiveInvoiceRules() kept %d, want 2", got)
	}
	if got, want := SumInvoiceRules(xs), 1+3; got != want {
		t.Errorf("SumInvoiceRules() = %d, want %d", got, want)
	}
	SortInvoiceRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceRules() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceRules(xs)) != 3 {
		t.Errorf("IndexInvoiceRules() = %v", IndexInvoiceRules(xs))
	}
	if err := ValidateInvoiceRules(ActiveInvoiceRules(xs)); err != nil {
		t.Errorf("ValidateInvoiceRules() = %v", err)
	}
}

func TestInvoiceRuleNormalizeClamps(t *testing.T) {
	x := SampleInvoiceRule("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceHoldRoundTrip(t *testing.T) {
	x := SampleInvoiceHold("invoicehold-1", when)
	got, err := ParseInvoiceHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceHold() = %v", err)
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

func TestInvoiceHoldParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceHold(f); err == nil {
		t.Error("ParseInvoiceHold() accepted a record with no id")
	}
	f = SampleInvoiceHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseInvoiceHold(f); err == nil {
		t.Error("ParseInvoiceHold() accepted a non-numeric weight")
	}
}

func TestInvoiceHoldCollections(t *testing.T) {
	xs := []InvoiceHold{SampleInvoiceHold("a", when), SampleInvoiceHold("bb", when), SampleInvoiceHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceHolds(xs)); got != 2 {
		t.Errorf("ActiveInvoiceHolds() kept %d, want 2", got)
	}
	if got, want := SumInvoiceHolds(xs), 1+3; got != want {
		t.Errorf("SumInvoiceHolds() = %d, want %d", got, want)
	}
	SortInvoiceHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceHolds() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceHolds(xs)) != 3 {
		t.Errorf("IndexInvoiceHolds() = %v", IndexInvoiceHolds(xs))
	}
	if err := ValidateInvoiceHolds(ActiveInvoiceHolds(xs)); err != nil {
		t.Errorf("ValidateInvoiceHolds() = %v", err)
	}
}

func TestInvoiceHoldNormalizeClamps(t *testing.T) {
	x := SampleInvoiceHold("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceClaimRoundTrip(t *testing.T) {
	x := SampleInvoiceClaim("invoiceclaim-1", when)
	got, err := ParseInvoiceClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceClaim() = %v", err)
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

func TestInvoiceClaimParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceClaim(f); err == nil {
		t.Error("ParseInvoiceClaim() accepted a record with no id")
	}
	f = SampleInvoiceClaim("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseInvoiceClaim(f); err == nil {
		t.Error("ParseInvoiceClaim() accepted a non-numeric attempts")
	}
}

func TestInvoiceClaimCollections(t *testing.T) {
	xs := []InvoiceClaim{SampleInvoiceClaim("a", when), SampleInvoiceClaim("bb", when), SampleInvoiceClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceClaims(xs)); got != 2 {
		t.Errorf("ActiveInvoiceClaims() kept %d, want 2", got)
	}
	if got, want := SumInvoiceClaims(xs), 1+3; got != want {
		t.Errorf("SumInvoiceClaims() = %d, want %d", got, want)
	}
	SortInvoiceClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceClaims() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceClaims(xs)) != 3 {
		t.Errorf("IndexInvoiceClaims() = %v", IndexInvoiceClaims(xs))
	}
	if err := ValidateInvoiceClaims(ActiveInvoiceClaims(xs)); err != nil {
		t.Errorf("ValidateInvoiceClaims() = %v", err)
	}
}

func TestInvoiceClaimNormalizeClamps(t *testing.T) {
	x := SampleInvoiceClaim("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestInvoiceCursorRoundTrip(t *testing.T) {
	x := SampleInvoiceCursor("invoicecursor-1", when)
	got, err := ParseInvoiceCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseInvoiceCursor() = %v", err)
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

func TestInvoiceCursorParseRejectsBadInput(t *testing.T) {
	f := SampleInvoiceCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseInvoiceCursor(f); err == nil {
		t.Error("ParseInvoiceCursor() accepted a record with no id")
	}
	f = SampleInvoiceCursor("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseInvoiceCursor(f); err == nil {
		t.Error("ParseInvoiceCursor() accepted a non-numeric priority")
	}
}

func TestInvoiceCursorCollections(t *testing.T) {
	xs := []InvoiceCursor{SampleInvoiceCursor("a", when), SampleInvoiceCursor("bb", when), SampleInvoiceCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveInvoiceCursors(xs)); got != 2 {
		t.Errorf("ActiveInvoiceCursors() kept %d, want 2", got)
	}
	if got, want := SumInvoiceCursors(xs), 1+3; got != want {
		t.Errorf("SumInvoiceCursors() = %d, want %d", got, want)
	}
	SortInvoiceCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortInvoiceCursors() put %q first", xs[0].ID)
	}
	if len(IndexInvoiceCursors(xs)) != 3 {
		t.Errorf("IndexInvoiceCursors() = %v", IndexInvoiceCursors(xs))
	}
	if err := ValidateInvoiceCursors(ActiveInvoiceCursors(xs)); err != nil {
		t.Errorf("ValidateInvoiceCursors() = %v", err)
	}
}

func TestInvoiceCursorNormalizeClamps(t *testing.T) {
	x := SampleInvoiceCursor("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
