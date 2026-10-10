package order

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
		if !strings.HasPrefix(Prefixes[i], "order/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleOrderRecord("ok", when).Fields()); err != nil {
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

func TestOrderRecordRoundTrip(t *testing.T) {
	x := SampleOrderRecord("orderrecord-1", when)
	got, err := ParseOrderRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderRecord() = %v", err)
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

func TestOrderRecordParseRejectsBadInput(t *testing.T) {
	f := SampleOrderRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderRecord(f); err == nil {
		t.Error("ParseOrderRecord() accepted a record with no id")
	}
	f = SampleOrderRecord("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseOrderRecord(f); err == nil {
		t.Error("ParseOrderRecord() accepted a non-numeric attempts")
	}
}

func TestOrderRecordCollections(t *testing.T) {
	xs := []OrderRecord{SampleOrderRecord("a", when), SampleOrderRecord("bb", when), SampleOrderRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderRecords(xs)); got != 2 {
		t.Errorf("ActiveOrderRecords() kept %d, want 2", got)
	}
	if got, want := SumOrderRecords(xs), 1+3; got != want {
		t.Errorf("SumOrderRecords() = %d, want %d", got, want)
	}
	SortOrderRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderRecords() put %q first", xs[0].ID)
	}
	if len(IndexOrderRecords(xs)) != 3 {
		t.Errorf("IndexOrderRecords() = %v", IndexOrderRecords(xs))
	}
	if err := ValidateOrderRecords(ActiveOrderRecords(xs)); err != nil {
		t.Errorf("ValidateOrderRecords() = %v", err)
	}
}

func TestOrderRecordNormalizeClamps(t *testing.T) {
	x := SampleOrderRecord("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderEntryRoundTrip(t *testing.T) {
	x := SampleOrderEntry("orderentry-1", when)
	got, err := ParseOrderEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderEntry() = %v", err)
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

func TestOrderEntryParseRejectsBadInput(t *testing.T) {
	f := SampleOrderEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderEntry(f); err == nil {
		t.Error("ParseOrderEntry() accepted a record with no id")
	}
	f = SampleOrderEntry("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseOrderEntry(f); err == nil {
		t.Error("ParseOrderEntry() accepted a non-numeric retained")
	}
}

func TestOrderEntryCollections(t *testing.T) {
	xs := []OrderEntry{SampleOrderEntry("a", when), SampleOrderEntry("bb", when), SampleOrderEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderEntrys(xs)); got != 2 {
		t.Errorf("ActiveOrderEntrys() kept %d, want 2", got)
	}
	if got, want := SumOrderEntrys(xs), 1+3; got != want {
		t.Errorf("SumOrderEntrys() = %d, want %d", got, want)
	}
	SortOrderEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderEntrys() put %q first", xs[0].ID)
	}
	if len(IndexOrderEntrys(xs)) != 3 {
		t.Errorf("IndexOrderEntrys() = %v", IndexOrderEntrys(xs))
	}
	if err := ValidateOrderEntrys(ActiveOrderEntrys(xs)); err != nil {
		t.Errorf("ValidateOrderEntrys() = %v", err)
	}
}

func TestOrderEntryNormalizeClamps(t *testing.T) {
	x := SampleOrderEntry("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderItemRoundTrip(t *testing.T) {
	x := SampleOrderItem("orderitem-1", when)
	got, err := ParseOrderItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderItem() = %v", err)
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

func TestOrderItemParseRejectsBadInput(t *testing.T) {
	f := SampleOrderItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderItem(f); err == nil {
		t.Error("ParseOrderItem() accepted a record with no id")
	}
	f = SampleOrderItem("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseOrderItem(f); err == nil {
		t.Error("ParseOrderItem() accepted a non-numeric priority")
	}
}

func TestOrderItemCollections(t *testing.T) {
	xs := []OrderItem{SampleOrderItem("a", when), SampleOrderItem("bb", when), SampleOrderItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderItems(xs)); got != 2 {
		t.Errorf("ActiveOrderItems() kept %d, want 2", got)
	}
	if got, want := SumOrderItems(xs), 1+3; got != want {
		t.Errorf("SumOrderItems() = %d, want %d", got, want)
	}
	SortOrderItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderItems() put %q first", xs[0].ID)
	}
	if len(IndexOrderItems(xs)) != 3 {
		t.Errorf("IndexOrderItems() = %v", IndexOrderItems(xs))
	}
	if err := ValidateOrderItems(ActiveOrderItems(xs)); err != nil {
		t.Errorf("ValidateOrderItems() = %v", err)
	}
}

func TestOrderItemNormalizeClamps(t *testing.T) {
	x := SampleOrderItem("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderDraftRoundTrip(t *testing.T) {
	x := SampleOrderDraft("orderdraft-1", when)
	got, err := ParseOrderDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderDraft() = %v", err)
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

func TestOrderDraftParseRejectsBadInput(t *testing.T) {
	f := SampleOrderDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderDraft(f); err == nil {
		t.Error("ParseOrderDraft() accepted a record with no id")
	}
	f = SampleOrderDraft("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseOrderDraft(f); err == nil {
		t.Error("ParseOrderDraft() accepted a non-numeric priority")
	}
}

func TestOrderDraftCollections(t *testing.T) {
	xs := []OrderDraft{SampleOrderDraft("a", when), SampleOrderDraft("bb", when), SampleOrderDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderDrafts(xs)); got != 2 {
		t.Errorf("ActiveOrderDrafts() kept %d, want 2", got)
	}
	if got, want := SumOrderDrafts(xs), 1+3; got != want {
		t.Errorf("SumOrderDrafts() = %d, want %d", got, want)
	}
	SortOrderDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderDrafts() put %q first", xs[0].ID)
	}
	if len(IndexOrderDrafts(xs)) != 3 {
		t.Errorf("IndexOrderDrafts() = %v", IndexOrderDrafts(xs))
	}
	if err := ValidateOrderDrafts(ActiveOrderDrafts(xs)); err != nil {
		t.Errorf("ValidateOrderDrafts() = %v", err)
	}
}

func TestOrderDraftNormalizeClamps(t *testing.T) {
	x := SampleOrderDraft("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderRevisionRoundTrip(t *testing.T) {
	x := SampleOrderRevision("orderrevision-1", when)
	got, err := ParseOrderRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderRevision() = %v", err)
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

func TestOrderRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleOrderRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderRevision(f); err == nil {
		t.Error("ParseOrderRevision() accepted a record with no id")
	}
	f = SampleOrderRevision("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseOrderRevision(f); err == nil {
		t.Error("ParseOrderRevision() accepted a non-numeric count")
	}
}

func TestOrderRevisionCollections(t *testing.T) {
	xs := []OrderRevision{SampleOrderRevision("a", when), SampleOrderRevision("bb", when), SampleOrderRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderRevisions(xs)); got != 2 {
		t.Errorf("ActiveOrderRevisions() kept %d, want 2", got)
	}
	if got, want := SumOrderRevisions(xs), 1+3; got != want {
		t.Errorf("SumOrderRevisions() = %d, want %d", got, want)
	}
	SortOrderRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderRevisions() put %q first", xs[0].ID)
	}
	if len(IndexOrderRevisions(xs)) != 3 {
		t.Errorf("IndexOrderRevisions() = %v", IndexOrderRevisions(xs))
	}
	if err := ValidateOrderRevisions(ActiveOrderRevisions(xs)); err != nil {
		t.Errorf("ValidateOrderRevisions() = %v", err)
	}
}

func TestOrderRevisionNormalizeClamps(t *testing.T) {
	x := SampleOrderRevision("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderSnapshotRoundTrip(t *testing.T) {
	x := SampleOrderSnapshot("ordersnapshot-1", when)
	got, err := ParseOrderSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderSnapshot() = %v", err)
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

func TestOrderSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleOrderSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderSnapshot(f); err == nil {
		t.Error("ParseOrderSnapshot() accepted a record with no id")
	}
	f = SampleOrderSnapshot("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseOrderSnapshot(f); err == nil {
		t.Error("ParseOrderSnapshot() accepted a non-numeric retained")
	}
}

func TestOrderSnapshotCollections(t *testing.T) {
	xs := []OrderSnapshot{SampleOrderSnapshot("a", when), SampleOrderSnapshot("bb", when), SampleOrderSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderSnapshots(xs)); got != 2 {
		t.Errorf("ActiveOrderSnapshots() kept %d, want 2", got)
	}
	if got, want := SumOrderSnapshots(xs), 1+3; got != want {
		t.Errorf("SumOrderSnapshots() = %d, want %d", got, want)
	}
	SortOrderSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexOrderSnapshots(xs)) != 3 {
		t.Errorf("IndexOrderSnapshots() = %v", IndexOrderSnapshots(xs))
	}
	if err := ValidateOrderSnapshots(ActiveOrderSnapshots(xs)); err != nil {
		t.Errorf("ValidateOrderSnapshots() = %v", err)
	}
}

func TestOrderSnapshotNormalizeClamps(t *testing.T) {
	x := SampleOrderSnapshot("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderBatchRoundTrip(t *testing.T) {
	x := SampleOrderBatch("orderbatch-1", when)
	got, err := ParseOrderBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderBatch() = %v", err)
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

func TestOrderBatchParseRejectsBadInput(t *testing.T) {
	f := SampleOrderBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderBatch(f); err == nil {
		t.Error("ParseOrderBatch() accepted a record with no id")
	}
	f = SampleOrderBatch("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseOrderBatch(f); err == nil {
		t.Error("ParseOrderBatch() accepted a non-numeric offset")
	}
}

func TestOrderBatchCollections(t *testing.T) {
	xs := []OrderBatch{SampleOrderBatch("a", when), SampleOrderBatch("bb", when), SampleOrderBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderBatchs(xs)); got != 2 {
		t.Errorf("ActiveOrderBatchs() kept %d, want 2", got)
	}
	if got, want := SumOrderBatchs(xs), 1+3; got != want {
		t.Errorf("SumOrderBatchs() = %d, want %d", got, want)
	}
	SortOrderBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderBatchs() put %q first", xs[0].ID)
	}
	if len(IndexOrderBatchs(xs)) != 3 {
		t.Errorf("IndexOrderBatchs() = %v", IndexOrderBatchs(xs))
	}
	if err := ValidateOrderBatchs(ActiveOrderBatchs(xs)); err != nil {
		t.Errorf("ValidateOrderBatchs() = %v", err)
	}
}

func TestOrderBatchNormalizeClamps(t *testing.T) {
	x := SampleOrderBatch("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderRequestRoundTrip(t *testing.T) {
	x := SampleOrderRequest("orderrequest-1", when)
	got, err := ParseOrderRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderRequest() = %v", err)
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

func TestOrderRequestParseRejectsBadInput(t *testing.T) {
	f := SampleOrderRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderRequest(f); err == nil {
		t.Error("ParseOrderRequest() accepted a record with no id")
	}
	f = SampleOrderRequest("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseOrderRequest(f); err == nil {
		t.Error("ParseOrderRequest() accepted a non-numeric version")
	}
}

func TestOrderRequestCollections(t *testing.T) {
	xs := []OrderRequest{SampleOrderRequest("a", when), SampleOrderRequest("bb", when), SampleOrderRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderRequests(xs)); got != 2 {
		t.Errorf("ActiveOrderRequests() kept %d, want 2", got)
	}
	if got, want := SumOrderRequests(xs), 1+3; got != want {
		t.Errorf("SumOrderRequests() = %d, want %d", got, want)
	}
	SortOrderRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderRequests() put %q first", xs[0].ID)
	}
	if len(IndexOrderRequests(xs)) != 3 {
		t.Errorf("IndexOrderRequests() = %v", IndexOrderRequests(xs))
	}
	if err := ValidateOrderRequests(ActiveOrderRequests(xs)); err != nil {
		t.Errorf("ValidateOrderRequests() = %v", err)
	}
}

func TestOrderRequestNormalizeClamps(t *testing.T) {
	x := SampleOrderRequest("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderSummaryRoundTrip(t *testing.T) {
	x := SampleOrderSummary("ordersummary-1", when)
	got, err := ParseOrderSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderSummary() = %v", err)
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

func TestOrderSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleOrderSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderSummary(f); err == nil {
		t.Error("ParseOrderSummary() accepted a record with no id")
	}
	f = SampleOrderSummary("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseOrderSummary(f); err == nil {
		t.Error("ParseOrderSummary() accepted a non-numeric count")
	}
}

func TestOrderSummaryCollections(t *testing.T) {
	xs := []OrderSummary{SampleOrderSummary("a", when), SampleOrderSummary("bb", when), SampleOrderSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderSummarys(xs)); got != 2 {
		t.Errorf("ActiveOrderSummarys() kept %d, want 2", got)
	}
	if got, want := SumOrderSummarys(xs), 1+3; got != want {
		t.Errorf("SumOrderSummarys() = %d, want %d", got, want)
	}
	SortOrderSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderSummarys() put %q first", xs[0].ID)
	}
	if len(IndexOrderSummarys(xs)) != 3 {
		t.Errorf("IndexOrderSummarys() = %v", IndexOrderSummarys(xs))
	}
	if err := ValidateOrderSummarys(ActiveOrderSummarys(xs)); err != nil {
		t.Errorf("ValidateOrderSummarys() = %v", err)
	}
}

func TestOrderSummaryNormalizeClamps(t *testing.T) {
	x := SampleOrderSummary("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderPolicyRoundTrip(t *testing.T) {
	x := SampleOrderPolicy("orderpolicy-1", when)
	got, err := ParseOrderPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderPolicy() = %v", err)
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

func TestOrderPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleOrderPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderPolicy(f); err == nil {
		t.Error("ParseOrderPolicy() accepted a record with no id")
	}
	f = SampleOrderPolicy("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseOrderPolicy(f); err == nil {
		t.Error("ParseOrderPolicy() accepted a non-numeric offset")
	}
}

func TestOrderPolicyCollections(t *testing.T) {
	xs := []OrderPolicy{SampleOrderPolicy("a", when), SampleOrderPolicy("bb", when), SampleOrderPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderPolicys(xs)); got != 2 {
		t.Errorf("ActiveOrderPolicys() kept %d, want 2", got)
	}
	if got, want := SumOrderPolicys(xs), 1+3; got != want {
		t.Errorf("SumOrderPolicys() = %d, want %d", got, want)
	}
	SortOrderPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderPolicys() put %q first", xs[0].ID)
	}
	if len(IndexOrderPolicys(xs)) != 3 {
		t.Errorf("IndexOrderPolicys() = %v", IndexOrderPolicys(xs))
	}
	if err := ValidateOrderPolicys(ActiveOrderPolicys(xs)); err != nil {
		t.Errorf("ValidateOrderPolicys() = %v", err)
	}
}

func TestOrderPolicyNormalizeClamps(t *testing.T) {
	x := SampleOrderPolicy("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderLinkRoundTrip(t *testing.T) {
	x := SampleOrderLink("orderlink-1", when)
	got, err := ParseOrderLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderLink() = %v", err)
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

func TestOrderLinkParseRejectsBadInput(t *testing.T) {
	f := SampleOrderLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderLink(f); err == nil {
		t.Error("ParseOrderLink() accepted a record with no id")
	}
	f = SampleOrderLink("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseOrderLink(f); err == nil {
		t.Error("ParseOrderLink() accepted a non-numeric capacity")
	}
}

func TestOrderLinkCollections(t *testing.T) {
	xs := []OrderLink{SampleOrderLink("a", when), SampleOrderLink("bb", when), SampleOrderLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderLinks(xs)); got != 2 {
		t.Errorf("ActiveOrderLinks() kept %d, want 2", got)
	}
	if got, want := SumOrderLinks(xs), 1+3; got != want {
		t.Errorf("SumOrderLinks() = %d, want %d", got, want)
	}
	SortOrderLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderLinks() put %q first", xs[0].ID)
	}
	if len(IndexOrderLinks(xs)) != 3 {
		t.Errorf("IndexOrderLinks() = %v", IndexOrderLinks(xs))
	}
	if err := ValidateOrderLinks(ActiveOrderLinks(xs)); err != nil {
		t.Errorf("ValidateOrderLinks() = %v", err)
	}
}

func TestOrderLinkNormalizeClamps(t *testing.T) {
	x := SampleOrderLink("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderNoteRoundTrip(t *testing.T) {
	x := SampleOrderNote("ordernote-1", when)
	got, err := ParseOrderNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderNote() = %v", err)
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

func TestOrderNoteParseRejectsBadInput(t *testing.T) {
	f := SampleOrderNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderNote(f); err == nil {
		t.Error("ParseOrderNote() accepted a record with no id")
	}
	f = SampleOrderNote("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseOrderNote(f); err == nil {
		t.Error("ParseOrderNote() accepted a non-numeric capacity")
	}
}

func TestOrderNoteCollections(t *testing.T) {
	xs := []OrderNote{SampleOrderNote("a", when), SampleOrderNote("bb", when), SampleOrderNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderNotes(xs)); got != 2 {
		t.Errorf("ActiveOrderNotes() kept %d, want 2", got)
	}
	if got, want := SumOrderNotes(xs), 1+3; got != want {
		t.Errorf("SumOrderNotes() = %d, want %d", got, want)
	}
	SortOrderNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderNotes() put %q first", xs[0].ID)
	}
	if len(IndexOrderNotes(xs)) != 3 {
		t.Errorf("IndexOrderNotes() = %v", IndexOrderNotes(xs))
	}
	if err := ValidateOrderNotes(ActiveOrderNotes(xs)); err != nil {
		t.Errorf("ValidateOrderNotes() = %v", err)
	}
}

func TestOrderNoteNormalizeClamps(t *testing.T) {
	x := SampleOrderNote("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderWindowRoundTrip(t *testing.T) {
	x := SampleOrderWindow("orderwindow-1", when)
	got, err := ParseOrderWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderWindow() = %v", err)
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

func TestOrderWindowParseRejectsBadInput(t *testing.T) {
	f := SampleOrderWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderWindow(f); err == nil {
		t.Error("ParseOrderWindow() accepted a record with no id")
	}
	f = SampleOrderWindow("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseOrderWindow(f); err == nil {
		t.Error("ParseOrderWindow() accepted a non-numeric attempts")
	}
}

func TestOrderWindowCollections(t *testing.T) {
	xs := []OrderWindow{SampleOrderWindow("a", when), SampleOrderWindow("bb", when), SampleOrderWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderWindows(xs)); got != 2 {
		t.Errorf("ActiveOrderWindows() kept %d, want 2", got)
	}
	if got, want := SumOrderWindows(xs), 1+3; got != want {
		t.Errorf("SumOrderWindows() = %d, want %d", got, want)
	}
	SortOrderWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderWindows() put %q first", xs[0].ID)
	}
	if len(IndexOrderWindows(xs)) != 3 {
		t.Errorf("IndexOrderWindows() = %v", IndexOrderWindows(xs))
	}
	if err := ValidateOrderWindows(ActiveOrderWindows(xs)); err != nil {
		t.Errorf("ValidateOrderWindows() = %v", err)
	}
}

func TestOrderWindowNormalizeClamps(t *testing.T) {
	x := SampleOrderWindow("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderQuotaRoundTrip(t *testing.T) {
	x := SampleOrderQuota("orderquota-1", when)
	got, err := ParseOrderQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderQuota() = %v", err)
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

func TestOrderQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleOrderQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderQuota(f); err == nil {
		t.Error("ParseOrderQuota() accepted a record with no id")
	}
	f = SampleOrderQuota("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseOrderQuota(f); err == nil {
		t.Error("ParseOrderQuota() accepted a non-numeric priority")
	}
}

func TestOrderQuotaCollections(t *testing.T) {
	xs := []OrderQuota{SampleOrderQuota("a", when), SampleOrderQuota("bb", when), SampleOrderQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderQuotas(xs)); got != 2 {
		t.Errorf("ActiveOrderQuotas() kept %d, want 2", got)
	}
	if got, want := SumOrderQuotas(xs), 1+3; got != want {
		t.Errorf("SumOrderQuotas() = %d, want %d", got, want)
	}
	SortOrderQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderQuotas() put %q first", xs[0].ID)
	}
	if len(IndexOrderQuotas(xs)) != 3 {
		t.Errorf("IndexOrderQuotas() = %v", IndexOrderQuotas(xs))
	}
	if err := ValidateOrderQuotas(ActiveOrderQuotas(xs)); err != nil {
		t.Errorf("ValidateOrderQuotas() = %v", err)
	}
}

func TestOrderQuotaNormalizeClamps(t *testing.T) {
	x := SampleOrderQuota("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderBucketRoundTrip(t *testing.T) {
	x := SampleOrderBucket("orderbucket-1", when)
	got, err := ParseOrderBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderBucket() = %v", err)
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

func TestOrderBucketParseRejectsBadInput(t *testing.T) {
	f := SampleOrderBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderBucket(f); err == nil {
		t.Error("ParseOrderBucket() accepted a record with no id")
	}
	f = SampleOrderBucket("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseOrderBucket(f); err == nil {
		t.Error("ParseOrderBucket() accepted a non-numeric offset")
	}
}

func TestOrderBucketCollections(t *testing.T) {
	xs := []OrderBucket{SampleOrderBucket("a", when), SampleOrderBucket("bb", when), SampleOrderBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderBuckets(xs)); got != 2 {
		t.Errorf("ActiveOrderBuckets() kept %d, want 2", got)
	}
	if got, want := SumOrderBuckets(xs), 1+3; got != want {
		t.Errorf("SumOrderBuckets() = %d, want %d", got, want)
	}
	SortOrderBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderBuckets() put %q first", xs[0].ID)
	}
	if len(IndexOrderBuckets(xs)) != 3 {
		t.Errorf("IndexOrderBuckets() = %v", IndexOrderBuckets(xs))
	}
	if err := ValidateOrderBuckets(ActiveOrderBuckets(xs)); err != nil {
		t.Errorf("ValidateOrderBuckets() = %v", err)
	}
}

func TestOrderBucketNormalizeClamps(t *testing.T) {
	x := SampleOrderBucket("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderLaneRoundTrip(t *testing.T) {
	x := SampleOrderLane("orderlane-1", when)
	got, err := ParseOrderLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderLane() = %v", err)
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

func TestOrderLaneParseRejectsBadInput(t *testing.T) {
	f := SampleOrderLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderLane(f); err == nil {
		t.Error("ParseOrderLane() accepted a record with no id")
	}
	f = SampleOrderLane("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseOrderLane(f); err == nil {
		t.Error("ParseOrderLane() accepted a non-numeric offset")
	}
}

func TestOrderLaneCollections(t *testing.T) {
	xs := []OrderLane{SampleOrderLane("a", when), SampleOrderLane("bb", when), SampleOrderLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderLanes(xs)); got != 2 {
		t.Errorf("ActiveOrderLanes() kept %d, want 2", got)
	}
	if got, want := SumOrderLanes(xs), 1+3; got != want {
		t.Errorf("SumOrderLanes() = %d, want %d", got, want)
	}
	SortOrderLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderLanes() put %q first", xs[0].ID)
	}
	if len(IndexOrderLanes(xs)) != 3 {
		t.Errorf("IndexOrderLanes() = %v", IndexOrderLanes(xs))
	}
	if err := ValidateOrderLanes(ActiveOrderLanes(xs)); err != nil {
		t.Errorf("ValidateOrderLanes() = %v", err)
	}
}

func TestOrderLaneNormalizeClamps(t *testing.T) {
	x := SampleOrderLane("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderSlotRoundTrip(t *testing.T) {
	x := SampleOrderSlot("orderslot-1", when)
	got, err := ParseOrderSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderSlot() = %v", err)
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

func TestOrderSlotParseRejectsBadInput(t *testing.T) {
	f := SampleOrderSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderSlot(f); err == nil {
		t.Error("ParseOrderSlot() accepted a record with no id")
	}
	f = SampleOrderSlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseOrderSlot(f); err == nil {
		t.Error("ParseOrderSlot() accepted a non-numeric count")
	}
}

func TestOrderSlotCollections(t *testing.T) {
	xs := []OrderSlot{SampleOrderSlot("a", when), SampleOrderSlot("bb", when), SampleOrderSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderSlots(xs)); got != 2 {
		t.Errorf("ActiveOrderSlots() kept %d, want 2", got)
	}
	if got, want := SumOrderSlots(xs), 1+3; got != want {
		t.Errorf("SumOrderSlots() = %d, want %d", got, want)
	}
	SortOrderSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderSlots() put %q first", xs[0].ID)
	}
	if len(IndexOrderSlots(xs)) != 3 {
		t.Errorf("IndexOrderSlots() = %v", IndexOrderSlots(xs))
	}
	if err := ValidateOrderSlots(ActiveOrderSlots(xs)); err != nil {
		t.Errorf("ValidateOrderSlots() = %v", err)
	}
}

func TestOrderSlotNormalizeClamps(t *testing.T) {
	x := SampleOrderSlot("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderTagRoundTrip(t *testing.T) {
	x := SampleOrderTag("ordertag-1", when)
	got, err := ParseOrderTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderTag() = %v", err)
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

func TestOrderTagParseRejectsBadInput(t *testing.T) {
	f := SampleOrderTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderTag(f); err == nil {
		t.Error("ParseOrderTag() accepted a record with no id")
	}
	f = SampleOrderTag("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseOrderTag(f); err == nil {
		t.Error("ParseOrderTag() accepted a non-numeric capacity")
	}
}

func TestOrderTagCollections(t *testing.T) {
	xs := []OrderTag{SampleOrderTag("a", when), SampleOrderTag("bb", when), SampleOrderTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderTags(xs)); got != 2 {
		t.Errorf("ActiveOrderTags() kept %d, want 2", got)
	}
	if got, want := SumOrderTags(xs), 1+3; got != want {
		t.Errorf("SumOrderTags() = %d, want %d", got, want)
	}
	SortOrderTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderTags() put %q first", xs[0].ID)
	}
	if len(IndexOrderTags(xs)) != 3 {
		t.Errorf("IndexOrderTags() = %v", IndexOrderTags(xs))
	}
	if err := ValidateOrderTags(ActiveOrderTags(xs)); err != nil {
		t.Errorf("ValidateOrderTags() = %v", err)
	}
}

func TestOrderTagNormalizeClamps(t *testing.T) {
	x := SampleOrderTag("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderRuleRoundTrip(t *testing.T) {
	x := SampleOrderRule("orderrule-1", when)
	got, err := ParseOrderRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderRule() = %v", err)
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

func TestOrderRuleParseRejectsBadInput(t *testing.T) {
	f := SampleOrderRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderRule(f); err == nil {
		t.Error("ParseOrderRule() accepted a record with no id")
	}
	f = SampleOrderRule("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseOrderRule(f); err == nil {
		t.Error("ParseOrderRule() accepted a non-numeric weight")
	}
}

func TestOrderRuleCollections(t *testing.T) {
	xs := []OrderRule{SampleOrderRule("a", when), SampleOrderRule("bb", when), SampleOrderRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderRules(xs)); got != 2 {
		t.Errorf("ActiveOrderRules() kept %d, want 2", got)
	}
	if got, want := SumOrderRules(xs), 1+3; got != want {
		t.Errorf("SumOrderRules() = %d, want %d", got, want)
	}
	SortOrderRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderRules() put %q first", xs[0].ID)
	}
	if len(IndexOrderRules(xs)) != 3 {
		t.Errorf("IndexOrderRules() = %v", IndexOrderRules(xs))
	}
	if err := ValidateOrderRules(ActiveOrderRules(xs)); err != nil {
		t.Errorf("ValidateOrderRules() = %v", err)
	}
}

func TestOrderRuleNormalizeClamps(t *testing.T) {
	x := SampleOrderRule("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderHoldRoundTrip(t *testing.T) {
	x := SampleOrderHold("orderhold-1", when)
	got, err := ParseOrderHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderHold() = %v", err)
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

func TestOrderHoldParseRejectsBadInput(t *testing.T) {
	f := SampleOrderHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderHold(f); err == nil {
		t.Error("ParseOrderHold() accepted a record with no id")
	}
	f = SampleOrderHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseOrderHold(f); err == nil {
		t.Error("ParseOrderHold() accepted a non-numeric weight")
	}
}

func TestOrderHoldCollections(t *testing.T) {
	xs := []OrderHold{SampleOrderHold("a", when), SampleOrderHold("bb", when), SampleOrderHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderHolds(xs)); got != 2 {
		t.Errorf("ActiveOrderHolds() kept %d, want 2", got)
	}
	if got, want := SumOrderHolds(xs), 1+3; got != want {
		t.Errorf("SumOrderHolds() = %d, want %d", got, want)
	}
	SortOrderHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderHolds() put %q first", xs[0].ID)
	}
	if len(IndexOrderHolds(xs)) != 3 {
		t.Errorf("IndexOrderHolds() = %v", IndexOrderHolds(xs))
	}
	if err := ValidateOrderHolds(ActiveOrderHolds(xs)); err != nil {
		t.Errorf("ValidateOrderHolds() = %v", err)
	}
}

func TestOrderHoldNormalizeClamps(t *testing.T) {
	x := SampleOrderHold("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderClaimRoundTrip(t *testing.T) {
	x := SampleOrderClaim("orderclaim-1", when)
	got, err := ParseOrderClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderClaim() = %v", err)
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

func TestOrderClaimParseRejectsBadInput(t *testing.T) {
	f := SampleOrderClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderClaim(f); err == nil {
		t.Error("ParseOrderClaim() accepted a record with no id")
	}
	f = SampleOrderClaim("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseOrderClaim(f); err == nil {
		t.Error("ParseOrderClaim() accepted a non-numeric capacity")
	}
}

func TestOrderClaimCollections(t *testing.T) {
	xs := []OrderClaim{SampleOrderClaim("a", when), SampleOrderClaim("bb", when), SampleOrderClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderClaims(xs)); got != 2 {
		t.Errorf("ActiveOrderClaims() kept %d, want 2", got)
	}
	if got, want := SumOrderClaims(xs), 1+3; got != want {
		t.Errorf("SumOrderClaims() = %d, want %d", got, want)
	}
	SortOrderClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderClaims() put %q first", xs[0].ID)
	}
	if len(IndexOrderClaims(xs)) != 3 {
		t.Errorf("IndexOrderClaims() = %v", IndexOrderClaims(xs))
	}
	if err := ValidateOrderClaims(ActiveOrderClaims(xs)); err != nil {
		t.Errorf("ValidateOrderClaims() = %v", err)
	}
}

func TestOrderClaimNormalizeClamps(t *testing.T) {
	x := SampleOrderClaim("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestOrderCursorRoundTrip(t *testing.T) {
	x := SampleOrderCursor("ordercursor-1", when)
	got, err := ParseOrderCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseOrderCursor() = %v", err)
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

func TestOrderCursorParseRejectsBadInput(t *testing.T) {
	f := SampleOrderCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseOrderCursor(f); err == nil {
		t.Error("ParseOrderCursor() accepted a record with no id")
	}
	f = SampleOrderCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseOrderCursor(f); err == nil {
		t.Error("ParseOrderCursor() accepted a non-numeric version")
	}
}

func TestOrderCursorCollections(t *testing.T) {
	xs := []OrderCursor{SampleOrderCursor("a", when), SampleOrderCursor("bb", when), SampleOrderCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveOrderCursors(xs)); got != 2 {
		t.Errorf("ActiveOrderCursors() kept %d, want 2", got)
	}
	if got, want := SumOrderCursors(xs), 1+3; got != want {
		t.Errorf("SumOrderCursors() = %d, want %d", got, want)
	}
	SortOrderCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortOrderCursors() put %q first", xs[0].ID)
	}
	if len(IndexOrderCursors(xs)) != 3 {
		t.Errorf("IndexOrderCursors() = %v", IndexOrderCursors(xs))
	}
	if err := ValidateOrderCursors(ActiveOrderCursors(xs)); err != nil {
		t.Errorf("ValidateOrderCursors() = %v", err)
	}
}

func TestOrderCursorNormalizeClamps(t *testing.T) {
	x := SampleOrderCursor("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
