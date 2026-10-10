package payment

import (
	"strings"
	"testing"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/config"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/log"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/store"
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
		if !strings.HasPrefix(Prefixes[i], "payment/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SamplePaymentRecord("ok", when).Fields()); err != nil {
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

func newRepo(t *testing.T) *Repo {
	t.Helper()
	cfg, err := config.Parse("")
	if err != nil {
		t.Fatal(err)
	}
	s, err := store.Open(log.Discard(), cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return NewRepo(log.Discard(), nil, s)
}

func TestRepoStartsEmpty(t *testing.T) {
	if got := newRepo(t).Count(); got != 0 {
		t.Errorf("Count() = %d, want 0", got)
	}
}

func TestPaymentRecordRoundTrip(t *testing.T) {
	x := SamplePaymentRecord("paymentrecord-1", when)
	got, err := ParsePaymentRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentRecord() = %v", err)
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

func TestPaymentRecordParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentRecord(f); err == nil {
		t.Error("ParsePaymentRecord() accepted a record with no id")
	}
	f = SamplePaymentRecord("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePaymentRecord(f); err == nil {
		t.Error("ParsePaymentRecord() accepted a non-numeric version")
	}
}

func TestPaymentRecordCollections(t *testing.T) {
	xs := []PaymentRecord{SamplePaymentRecord("a", when), SamplePaymentRecord("bb", when), SamplePaymentRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentRecords(xs)); got != 2 {
		t.Errorf("ActivePaymentRecords() kept %d, want 2", got)
	}
	if got, want := SumPaymentRecords(xs), 1+3; got != want {
		t.Errorf("SumPaymentRecords() = %d, want %d", got, want)
	}
	SortPaymentRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentRecords() put %q first", xs[0].ID)
	}
	if len(IndexPaymentRecords(xs)) != 3 {
		t.Errorf("IndexPaymentRecords() = %v", IndexPaymentRecords(xs))
	}
	if err := ValidatePaymentRecords(ActivePaymentRecords(xs)); err != nil {
		t.Errorf("ValidatePaymentRecords() = %v", err)
	}
}

func TestPaymentRecordNormalizeClamps(t *testing.T) {
	x := SamplePaymentRecord("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentRecord("r1", when)
	rev, err := r.PutPaymentRecord(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentRecord() = %v", err)
	}
	got, gotRev, err := r.GetPaymentRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentRecords() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentRecord(PaymentRecord{}, 0); err == nil {
		t.Error("PutPaymentRecord() accepted an empty entity")
	}
}

func TestPaymentEntryRoundTrip(t *testing.T) {
	x := SamplePaymentEntry("paymententry-1", when)
	got, err := ParsePaymentEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentEntry() = %v", err)
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

func TestPaymentEntryParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentEntry(f); err == nil {
		t.Error("ParsePaymentEntry() accepted a record with no id")
	}
	f = SamplePaymentEntry("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePaymentEntry(f); err == nil {
		t.Error("ParsePaymentEntry() accepted a non-numeric capacity")
	}
}

func TestPaymentEntryCollections(t *testing.T) {
	xs := []PaymentEntry{SamplePaymentEntry("a", when), SamplePaymentEntry("bb", when), SamplePaymentEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentEntrys(xs)); got != 2 {
		t.Errorf("ActivePaymentEntrys() kept %d, want 2", got)
	}
	if got, want := SumPaymentEntrys(xs), 1+3; got != want {
		t.Errorf("SumPaymentEntrys() = %d, want %d", got, want)
	}
	SortPaymentEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentEntrys() put %q first", xs[0].ID)
	}
	if len(IndexPaymentEntrys(xs)) != 3 {
		t.Errorf("IndexPaymentEntrys() = %v", IndexPaymentEntrys(xs))
	}
	if err := ValidatePaymentEntrys(ActivePaymentEntrys(xs)); err != nil {
		t.Errorf("ValidatePaymentEntrys() = %v", err)
	}
}

func TestPaymentEntryNormalizeClamps(t *testing.T) {
	x := SamplePaymentEntry("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentEntry("r1", when)
	rev, err := r.PutPaymentEntry(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentEntry() = %v", err)
	}
	got, gotRev, err := r.GetPaymentEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentEntry(PaymentEntry{}, 0); err == nil {
		t.Error("PutPaymentEntry() accepted an empty entity")
	}
}

func TestPaymentItemRoundTrip(t *testing.T) {
	x := SamplePaymentItem("paymentitem-1", when)
	got, err := ParsePaymentItem(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentItem() = %v", err)
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

func TestPaymentItemParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentItem(f); err == nil {
		t.Error("ParsePaymentItem() accepted a record with no id")
	}
	f = SamplePaymentItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePaymentItem(f); err == nil {
		t.Error("ParsePaymentItem() accepted a non-numeric version")
	}
}

func TestPaymentItemCollections(t *testing.T) {
	xs := []PaymentItem{SamplePaymentItem("a", when), SamplePaymentItem("bb", when), SamplePaymentItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentItems(xs)); got != 2 {
		t.Errorf("ActivePaymentItems() kept %d, want 2", got)
	}
	if got, want := SumPaymentItems(xs), 1+3; got != want {
		t.Errorf("SumPaymentItems() = %d, want %d", got, want)
	}
	SortPaymentItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentItems() put %q first", xs[0].ID)
	}
	if len(IndexPaymentItems(xs)) != 3 {
		t.Errorf("IndexPaymentItems() = %v", IndexPaymentItems(xs))
	}
	if err := ValidatePaymentItems(ActivePaymentItems(xs)); err != nil {
		t.Errorf("ValidatePaymentItems() = %v", err)
	}
}

func TestPaymentItemNormalizeClamps(t *testing.T) {
	x := SamplePaymentItem("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentItem("r1", when)
	rev, err := r.PutPaymentItem(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentItem() = %v", err)
	}
	got, gotRev, err := r.GetPaymentItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentItems() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentItem(PaymentItem{}, 0); err == nil {
		t.Error("PutPaymentItem() accepted an empty entity")
	}
}

func TestPaymentDraftRoundTrip(t *testing.T) {
	x := SamplePaymentDraft("paymentdraft-1", when)
	got, err := ParsePaymentDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentDraft() = %v", err)
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

func TestPaymentDraftParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentDraft(f); err == nil {
		t.Error("ParsePaymentDraft() accepted a record with no id")
	}
	f = SamplePaymentDraft("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePaymentDraft(f); err == nil {
		t.Error("ParsePaymentDraft() accepted a non-numeric offset")
	}
}

func TestPaymentDraftCollections(t *testing.T) {
	xs := []PaymentDraft{SamplePaymentDraft("a", when), SamplePaymentDraft("bb", when), SamplePaymentDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentDrafts(xs)); got != 2 {
		t.Errorf("ActivePaymentDrafts() kept %d, want 2", got)
	}
	if got, want := SumPaymentDrafts(xs), 1+3; got != want {
		t.Errorf("SumPaymentDrafts() = %d, want %d", got, want)
	}
	SortPaymentDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentDrafts() put %q first", xs[0].ID)
	}
	if len(IndexPaymentDrafts(xs)) != 3 {
		t.Errorf("IndexPaymentDrafts() = %v", IndexPaymentDrafts(xs))
	}
	if err := ValidatePaymentDrafts(ActivePaymentDrafts(xs)); err != nil {
		t.Errorf("ValidatePaymentDrafts() = %v", err)
	}
}

func TestPaymentDraftNormalizeClamps(t *testing.T) {
	x := SamplePaymentDraft("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentDraft("r1", when)
	rev, err := r.PutPaymentDraft(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentDraft() = %v", err)
	}
	got, gotRev, err := r.GetPaymentDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentDraft(PaymentDraft{}, 0); err == nil {
		t.Error("PutPaymentDraft() accepted an empty entity")
	}
}

func TestPaymentRevisionRoundTrip(t *testing.T) {
	x := SamplePaymentRevision("paymentrevision-1", when)
	got, err := ParsePaymentRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentRevision() = %v", err)
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

func TestPaymentRevisionParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentRevision(f); err == nil {
		t.Error("ParsePaymentRevision() accepted a record with no id")
	}
	f = SamplePaymentRevision("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParsePaymentRevision(f); err == nil {
		t.Error("ParsePaymentRevision() accepted a non-numeric attempts")
	}
}

func TestPaymentRevisionCollections(t *testing.T) {
	xs := []PaymentRevision{SamplePaymentRevision("a", when), SamplePaymentRevision("bb", when), SamplePaymentRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentRevisions(xs)); got != 2 {
		t.Errorf("ActivePaymentRevisions() kept %d, want 2", got)
	}
	if got, want := SumPaymentRevisions(xs), 1+3; got != want {
		t.Errorf("SumPaymentRevisions() = %d, want %d", got, want)
	}
	SortPaymentRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentRevisions() put %q first", xs[0].ID)
	}
	if len(IndexPaymentRevisions(xs)) != 3 {
		t.Errorf("IndexPaymentRevisions() = %v", IndexPaymentRevisions(xs))
	}
	if err := ValidatePaymentRevisions(ActivePaymentRevisions(xs)); err != nil {
		t.Errorf("ValidatePaymentRevisions() = %v", err)
	}
}

func TestPaymentRevisionNormalizeClamps(t *testing.T) {
	x := SamplePaymentRevision("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentRevision("r1", when)
	rev, err := r.PutPaymentRevision(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentRevision() = %v", err)
	}
	got, gotRev, err := r.GetPaymentRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentRevision(PaymentRevision{}, 0); err == nil {
		t.Error("PutPaymentRevision() accepted an empty entity")
	}
}

func TestPaymentSnapshotRoundTrip(t *testing.T) {
	x := SamplePaymentSnapshot("paymentsnapshot-1", when)
	got, err := ParsePaymentSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentSnapshot() = %v", err)
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

func TestPaymentSnapshotParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentSnapshot(f); err == nil {
		t.Error("ParsePaymentSnapshot() accepted a record with no id")
	}
	f = SamplePaymentSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParsePaymentSnapshot(f); err == nil {
		t.Error("ParsePaymentSnapshot() accepted a non-numeric priority")
	}
}

func TestPaymentSnapshotCollections(t *testing.T) {
	xs := []PaymentSnapshot{SamplePaymentSnapshot("a", when), SamplePaymentSnapshot("bb", when), SamplePaymentSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentSnapshots(xs)); got != 2 {
		t.Errorf("ActivePaymentSnapshots() kept %d, want 2", got)
	}
	if got, want := SumPaymentSnapshots(xs), 1+3; got != want {
		t.Errorf("SumPaymentSnapshots() = %d, want %d", got, want)
	}
	SortPaymentSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexPaymentSnapshots(xs)) != 3 {
		t.Errorf("IndexPaymentSnapshots() = %v", IndexPaymentSnapshots(xs))
	}
	if err := ValidatePaymentSnapshots(ActivePaymentSnapshots(xs)); err != nil {
		t.Errorf("ValidatePaymentSnapshots() = %v", err)
	}
}

func TestPaymentSnapshotNormalizeClamps(t *testing.T) {
	x := SamplePaymentSnapshot("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentSnapshot("r1", when)
	rev, err := r.PutPaymentSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetPaymentSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentSnapshot(PaymentSnapshot{}, 0); err == nil {
		t.Error("PutPaymentSnapshot() accepted an empty entity")
	}
}

func TestPaymentBatchRoundTrip(t *testing.T) {
	x := SamplePaymentBatch("paymentbatch-1", when)
	got, err := ParsePaymentBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentBatch() = %v", err)
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

func TestPaymentBatchParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentBatch(f); err == nil {
		t.Error("ParsePaymentBatch() accepted a record with no id")
	}
	f = SamplePaymentBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePaymentBatch(f); err == nil {
		t.Error("ParsePaymentBatch() accepted a non-numeric capacity")
	}
}

func TestPaymentBatchCollections(t *testing.T) {
	xs := []PaymentBatch{SamplePaymentBatch("a", when), SamplePaymentBatch("bb", when), SamplePaymentBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentBatchs(xs)); got != 2 {
		t.Errorf("ActivePaymentBatchs() kept %d, want 2", got)
	}
	if got, want := SumPaymentBatchs(xs), 1+3; got != want {
		t.Errorf("SumPaymentBatchs() = %d, want %d", got, want)
	}
	SortPaymentBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentBatchs() put %q first", xs[0].ID)
	}
	if len(IndexPaymentBatchs(xs)) != 3 {
		t.Errorf("IndexPaymentBatchs() = %v", IndexPaymentBatchs(xs))
	}
	if err := ValidatePaymentBatchs(ActivePaymentBatchs(xs)); err != nil {
		t.Errorf("ValidatePaymentBatchs() = %v", err)
	}
}

func TestPaymentBatchNormalizeClamps(t *testing.T) {
	x := SamplePaymentBatch("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentBatch("r1", when)
	rev, err := r.PutPaymentBatch(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentBatch() = %v", err)
	}
	got, gotRev, err := r.GetPaymentBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentBatch(PaymentBatch{}, 0); err == nil {
		t.Error("PutPaymentBatch() accepted an empty entity")
	}
}

func TestPaymentRequestRoundTrip(t *testing.T) {
	x := SamplePaymentRequest("paymentrequest-1", when)
	got, err := ParsePaymentRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentRequest() = %v", err)
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

func TestPaymentRequestParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentRequest(f); err == nil {
		t.Error("ParsePaymentRequest() accepted a record with no id")
	}
	f = SamplePaymentRequest("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParsePaymentRequest(f); err == nil {
		t.Error("ParsePaymentRequest() accepted a non-numeric offset")
	}
}

func TestPaymentRequestCollections(t *testing.T) {
	xs := []PaymentRequest{SamplePaymentRequest("a", when), SamplePaymentRequest("bb", when), SamplePaymentRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentRequests(xs)); got != 2 {
		t.Errorf("ActivePaymentRequests() kept %d, want 2", got)
	}
	if got, want := SumPaymentRequests(xs), 1+3; got != want {
		t.Errorf("SumPaymentRequests() = %d, want %d", got, want)
	}
	SortPaymentRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentRequests() put %q first", xs[0].ID)
	}
	if len(IndexPaymentRequests(xs)) != 3 {
		t.Errorf("IndexPaymentRequests() = %v", IndexPaymentRequests(xs))
	}
	if err := ValidatePaymentRequests(ActivePaymentRequests(xs)); err != nil {
		t.Errorf("ValidatePaymentRequests() = %v", err)
	}
}

func TestPaymentRequestNormalizeClamps(t *testing.T) {
	x := SamplePaymentRequest("a", when)
	x.Source = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentRequest("r1", when)
	rev, err := r.PutPaymentRequest(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentRequest() = %v", err)
	}
	got, gotRev, err := r.GetPaymentRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentRequests() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentRequest(PaymentRequest{}, 0); err == nil {
		t.Error("PutPaymentRequest() accepted an empty entity")
	}
}

func TestPaymentSummaryRoundTrip(t *testing.T) {
	x := SamplePaymentSummary("paymentsummary-1", when)
	got, err := ParsePaymentSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentSummary() = %v", err)
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

func TestPaymentSummaryParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentSummary(f); err == nil {
		t.Error("ParsePaymentSummary() accepted a record with no id")
	}
	f = SamplePaymentSummary("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePaymentSummary(f); err == nil {
		t.Error("ParsePaymentSummary() accepted a non-numeric retained")
	}
}

func TestPaymentSummaryCollections(t *testing.T) {
	xs := []PaymentSummary{SamplePaymentSummary("a", when), SamplePaymentSummary("bb", when), SamplePaymentSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentSummarys(xs)); got != 2 {
		t.Errorf("ActivePaymentSummarys() kept %d, want 2", got)
	}
	if got, want := SumPaymentSummarys(xs), 1+3; got != want {
		t.Errorf("SumPaymentSummarys() = %d, want %d", got, want)
	}
	SortPaymentSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentSummarys() put %q first", xs[0].ID)
	}
	if len(IndexPaymentSummarys(xs)) != 3 {
		t.Errorf("IndexPaymentSummarys() = %v", IndexPaymentSummarys(xs))
	}
	if err := ValidatePaymentSummarys(ActivePaymentSummarys(xs)); err != nil {
		t.Errorf("ValidatePaymentSummarys() = %v", err)
	}
}

func TestPaymentSummaryNormalizeClamps(t *testing.T) {
	x := SamplePaymentSummary("a", when)
	x.Owner = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentSummary("r1", when)
	rev, err := r.PutPaymentSummary(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentSummary() = %v", err)
	}
	got, gotRev, err := r.GetPaymentSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentSummary(PaymentSummary{}, 0); err == nil {
		t.Error("PutPaymentSummary() accepted an empty entity")
	}
}

func TestPaymentPolicyRoundTrip(t *testing.T) {
	x := SamplePaymentPolicy("paymentpolicy-1", when)
	got, err := ParsePaymentPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentPolicy() = %v", err)
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

func TestPaymentPolicyParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentPolicy(f); err == nil {
		t.Error("ParsePaymentPolicy() accepted a record with no id")
	}
	f = SamplePaymentPolicy("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParsePaymentPolicy(f); err == nil {
		t.Error("ParsePaymentPolicy() accepted a non-numeric weight")
	}
}

func TestPaymentPolicyCollections(t *testing.T) {
	xs := []PaymentPolicy{SamplePaymentPolicy("a", when), SamplePaymentPolicy("bb", when), SamplePaymentPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentPolicys(xs)); got != 2 {
		t.Errorf("ActivePaymentPolicys() kept %d, want 2", got)
	}
	if got, want := SumPaymentPolicys(xs), 1+3; got != want {
		t.Errorf("SumPaymentPolicys() = %d, want %d", got, want)
	}
	SortPaymentPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentPolicys() put %q first", xs[0].ID)
	}
	if len(IndexPaymentPolicys(xs)) != 3 {
		t.Errorf("IndexPaymentPolicys() = %v", IndexPaymentPolicys(xs))
	}
	if err := ValidatePaymentPolicys(ActivePaymentPolicys(xs)); err != nil {
		t.Errorf("ValidatePaymentPolicys() = %v", err)
	}
}

func TestPaymentPolicyNormalizeClamps(t *testing.T) {
	x := SamplePaymentPolicy("a", when)
	x.Comment = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentPolicy("r1", when)
	rev, err := r.PutPaymentPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentPolicy() = %v", err)
	}
	got, gotRev, err := r.GetPaymentPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentPolicy(PaymentPolicy{}, 0); err == nil {
		t.Error("PutPaymentPolicy() accepted an empty entity")
	}
}

func TestPaymentLinkRoundTrip(t *testing.T) {
	x := SamplePaymentLink("paymentlink-1", when)
	got, err := ParsePaymentLink(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentLink() = %v", err)
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

func TestPaymentLinkParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentLink(f); err == nil {
		t.Error("ParsePaymentLink() accepted a record with no id")
	}
	f = SamplePaymentLink("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParsePaymentLink(f); err == nil {
		t.Error("ParsePaymentLink() accepted a non-numeric weight")
	}
}

func TestPaymentLinkCollections(t *testing.T) {
	xs := []PaymentLink{SamplePaymentLink("a", when), SamplePaymentLink("bb", when), SamplePaymentLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentLinks(xs)); got != 2 {
		t.Errorf("ActivePaymentLinks() kept %d, want 2", got)
	}
	if got, want := SumPaymentLinks(xs), 1+3; got != want {
		t.Errorf("SumPaymentLinks() = %d, want %d", got, want)
	}
	SortPaymentLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentLinks() put %q first", xs[0].ID)
	}
	if len(IndexPaymentLinks(xs)) != 3 {
		t.Errorf("IndexPaymentLinks() = %v", IndexPaymentLinks(xs))
	}
	if err := ValidatePaymentLinks(ActivePaymentLinks(xs)); err != nil {
		t.Errorf("ValidatePaymentLinks() = %v", err)
	}
}

func TestPaymentLinkNormalizeClamps(t *testing.T) {
	x := SamplePaymentLink("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentLink("r1", when)
	rev, err := r.PutPaymentLink(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentLink() = %v", err)
	}
	got, gotRev, err := r.GetPaymentLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentLinks() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentLink(PaymentLink{}, 0); err == nil {
		t.Error("PutPaymentLink() accepted an empty entity")
	}
}

func TestPaymentNoteRoundTrip(t *testing.T) {
	x := SamplePaymentNote("paymentnote-1", when)
	got, err := ParsePaymentNote(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentNote() = %v", err)
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

func TestPaymentNoteParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentNote(f); err == nil {
		t.Error("ParsePaymentNote() accepted a record with no id")
	}
	f = SamplePaymentNote("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePaymentNote(f); err == nil {
		t.Error("ParsePaymentNote() accepted a non-numeric count")
	}
}

func TestPaymentNoteCollections(t *testing.T) {
	xs := []PaymentNote{SamplePaymentNote("a", when), SamplePaymentNote("bb", when), SamplePaymentNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentNotes(xs)); got != 2 {
		t.Errorf("ActivePaymentNotes() kept %d, want 2", got)
	}
	if got, want := SumPaymentNotes(xs), 1+3; got != want {
		t.Errorf("SumPaymentNotes() = %d, want %d", got, want)
	}
	SortPaymentNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentNotes() put %q first", xs[0].ID)
	}
	if len(IndexPaymentNotes(xs)) != 3 {
		t.Errorf("IndexPaymentNotes() = %v", IndexPaymentNotes(xs))
	}
	if err := ValidatePaymentNotes(ActivePaymentNotes(xs)); err != nil {
		t.Errorf("ValidatePaymentNotes() = %v", err)
	}
}

func TestPaymentNoteNormalizeClamps(t *testing.T) {
	x := SamplePaymentNote("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentNote("r1", when)
	rev, err := r.PutPaymentNote(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentNote() = %v", err)
	}
	got, gotRev, err := r.GetPaymentNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentNotes() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentNote(PaymentNote{}, 0); err == nil {
		t.Error("PutPaymentNote() accepted an empty entity")
	}
}

func TestPaymentWindowRoundTrip(t *testing.T) {
	x := SamplePaymentWindow("paymentwindow-1", when)
	got, err := ParsePaymentWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentWindow() = %v", err)
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

func TestPaymentWindowParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentWindow(f); err == nil {
		t.Error("ParsePaymentWindow() accepted a record with no id")
	}
	f = SamplePaymentWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParsePaymentWindow(f); err == nil {
		t.Error("ParsePaymentWindow() accepted a non-numeric priority")
	}
}

func TestPaymentWindowCollections(t *testing.T) {
	xs := []PaymentWindow{SamplePaymentWindow("a", when), SamplePaymentWindow("bb", when), SamplePaymentWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentWindows(xs)); got != 2 {
		t.Errorf("ActivePaymentWindows() kept %d, want 2", got)
	}
	if got, want := SumPaymentWindows(xs), 1+3; got != want {
		t.Errorf("SumPaymentWindows() = %d, want %d", got, want)
	}
	SortPaymentWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentWindows() put %q first", xs[0].ID)
	}
	if len(IndexPaymentWindows(xs)) != 3 {
		t.Errorf("IndexPaymentWindows() = %v", IndexPaymentWindows(xs))
	}
	if err := ValidatePaymentWindows(ActivePaymentWindows(xs)); err != nil {
		t.Errorf("ValidatePaymentWindows() = %v", err)
	}
}

func TestPaymentWindowNormalizeClamps(t *testing.T) {
	x := SamplePaymentWindow("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentWindow("r1", when)
	rev, err := r.PutPaymentWindow(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentWindow() = %v", err)
	}
	got, gotRev, err := r.GetPaymentWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentWindows() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentWindow(PaymentWindow{}, 0); err == nil {
		t.Error("PutPaymentWindow() accepted an empty entity")
	}
}

func TestPaymentQuotaRoundTrip(t *testing.T) {
	x := SamplePaymentQuota("paymentquota-1", when)
	got, err := ParsePaymentQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentQuota() = %v", err)
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

func TestPaymentQuotaParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentQuota(f); err == nil {
		t.Error("ParsePaymentQuota() accepted a record with no id")
	}
	f = SamplePaymentQuota("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParsePaymentQuota(f); err == nil {
		t.Error("ParsePaymentQuota() accepted a non-numeric count")
	}
}

func TestPaymentQuotaCollections(t *testing.T) {
	xs := []PaymentQuota{SamplePaymentQuota("a", when), SamplePaymentQuota("bb", when), SamplePaymentQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentQuotas(xs)); got != 2 {
		t.Errorf("ActivePaymentQuotas() kept %d, want 2", got)
	}
	if got, want := SumPaymentQuotas(xs), 1+3; got != want {
		t.Errorf("SumPaymentQuotas() = %d, want %d", got, want)
	}
	SortPaymentQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentQuotas() put %q first", xs[0].ID)
	}
	if len(IndexPaymentQuotas(xs)) != 3 {
		t.Errorf("IndexPaymentQuotas() = %v", IndexPaymentQuotas(xs))
	}
	if err := ValidatePaymentQuotas(ActivePaymentQuotas(xs)); err != nil {
		t.Errorf("ValidatePaymentQuotas() = %v", err)
	}
}

func TestPaymentQuotaNormalizeClamps(t *testing.T) {
	x := SamplePaymentQuota("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentQuota("r1", when)
	rev, err := r.PutPaymentQuota(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentQuota() = %v", err)
	}
	got, gotRev, err := r.GetPaymentQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentQuota(PaymentQuota{}, 0); err == nil {
		t.Error("PutPaymentQuota() accepted an empty entity")
	}
}

func TestPaymentBucketRoundTrip(t *testing.T) {
	x := SamplePaymentBucket("paymentbucket-1", when)
	got, err := ParsePaymentBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentBucket() = %v", err)
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

func TestPaymentBucketParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentBucket(f); err == nil {
		t.Error("ParsePaymentBucket() accepted a record with no id")
	}
	f = SamplePaymentBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParsePaymentBucket(f); err == nil {
		t.Error("ParsePaymentBucket() accepted a non-numeric retained")
	}
}

func TestPaymentBucketCollections(t *testing.T) {
	xs := []PaymentBucket{SamplePaymentBucket("a", when), SamplePaymentBucket("bb", when), SamplePaymentBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentBuckets(xs)); got != 2 {
		t.Errorf("ActivePaymentBuckets() kept %d, want 2", got)
	}
	if got, want := SumPaymentBuckets(xs), 1+3; got != want {
		t.Errorf("SumPaymentBuckets() = %d, want %d", got, want)
	}
	SortPaymentBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentBuckets() put %q first", xs[0].ID)
	}
	if len(IndexPaymentBuckets(xs)) != 3 {
		t.Errorf("IndexPaymentBuckets() = %v", IndexPaymentBuckets(xs))
	}
	if err := ValidatePaymentBuckets(ActivePaymentBuckets(xs)); err != nil {
		t.Errorf("ValidatePaymentBuckets() = %v", err)
	}
}

func TestPaymentBucketNormalizeClamps(t *testing.T) {
	x := SamplePaymentBucket("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentBucket("r1", when)
	rev, err := r.PutPaymentBucket(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentBucket() = %v", err)
	}
	got, gotRev, err := r.GetPaymentBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentBucket(PaymentBucket{}, 0); err == nil {
		t.Error("PutPaymentBucket() accepted an empty entity")
	}
}

func TestPaymentLaneRoundTrip(t *testing.T) {
	x := SamplePaymentLane("paymentlane-1", when)
	got, err := ParsePaymentLane(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentLane() = %v", err)
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

func TestPaymentLaneParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentLane(f); err == nil {
		t.Error("ParsePaymentLane() accepted a record with no id")
	}
	f = SamplePaymentLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParsePaymentLane(f); err == nil {
		t.Error("ParsePaymentLane() accepted a non-numeric priority")
	}
}

func TestPaymentLaneCollections(t *testing.T) {
	xs := []PaymentLane{SamplePaymentLane("a", when), SamplePaymentLane("bb", when), SamplePaymentLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentLanes(xs)); got != 2 {
		t.Errorf("ActivePaymentLanes() kept %d, want 2", got)
	}
	if got, want := SumPaymentLanes(xs), 1+3; got != want {
		t.Errorf("SumPaymentLanes() = %d, want %d", got, want)
	}
	SortPaymentLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentLanes() put %q first", xs[0].ID)
	}
	if len(IndexPaymentLanes(xs)) != 3 {
		t.Errorf("IndexPaymentLanes() = %v", IndexPaymentLanes(xs))
	}
	if err := ValidatePaymentLanes(ActivePaymentLanes(xs)); err != nil {
		t.Errorf("ValidatePaymentLanes() = %v", err)
	}
}

func TestPaymentLaneNormalizeClamps(t *testing.T) {
	x := SamplePaymentLane("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentLane("r1", when)
	rev, err := r.PutPaymentLane(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentLane() = %v", err)
	}
	got, gotRev, err := r.GetPaymentLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentLanes() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentLane(PaymentLane{}, 0); err == nil {
		t.Error("PutPaymentLane() accepted an empty entity")
	}
}

func TestPaymentSlotRoundTrip(t *testing.T) {
	x := SamplePaymentSlot("paymentslot-1", when)
	got, err := ParsePaymentSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentSlot() = %v", err)
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

func TestPaymentSlotParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentSlot(f); err == nil {
		t.Error("ParsePaymentSlot() accepted a record with no id")
	}
	f = SamplePaymentSlot("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePaymentSlot(f); err == nil {
		t.Error("ParsePaymentSlot() accepted a non-numeric capacity")
	}
}

func TestPaymentSlotCollections(t *testing.T) {
	xs := []PaymentSlot{SamplePaymentSlot("a", when), SamplePaymentSlot("bb", when), SamplePaymentSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentSlots(xs)); got != 2 {
		t.Errorf("ActivePaymentSlots() kept %d, want 2", got)
	}
	if got, want := SumPaymentSlots(xs), 1+3; got != want {
		t.Errorf("SumPaymentSlots() = %d, want %d", got, want)
	}
	SortPaymentSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentSlots() put %q first", xs[0].ID)
	}
	if len(IndexPaymentSlots(xs)) != 3 {
		t.Errorf("IndexPaymentSlots() = %v", IndexPaymentSlots(xs))
	}
	if err := ValidatePaymentSlots(ActivePaymentSlots(xs)); err != nil {
		t.Errorf("ValidatePaymentSlots() = %v", err)
	}
}

func TestPaymentSlotNormalizeClamps(t *testing.T) {
	x := SamplePaymentSlot("a", when)
	x.Comment = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentSlot("r1", when)
	rev, err := r.PutPaymentSlot(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentSlot() = %v", err)
	}
	got, gotRev, err := r.GetPaymentSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentSlots() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentSlot(PaymentSlot{}, 0); err == nil {
		t.Error("PutPaymentSlot() accepted an empty entity")
	}
}

func TestPaymentTagRoundTrip(t *testing.T) {
	x := SamplePaymentTag("paymenttag-1", when)
	got, err := ParsePaymentTag(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentTag() = %v", err)
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

func TestPaymentTagParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentTag(f); err == nil {
		t.Error("ParsePaymentTag() accepted a record with no id")
	}
	f = SamplePaymentTag("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePaymentTag(f); err == nil {
		t.Error("ParsePaymentTag() accepted a non-numeric version")
	}
}

func TestPaymentTagCollections(t *testing.T) {
	xs := []PaymentTag{SamplePaymentTag("a", when), SamplePaymentTag("bb", when), SamplePaymentTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentTags(xs)); got != 2 {
		t.Errorf("ActivePaymentTags() kept %d, want 2", got)
	}
	if got, want := SumPaymentTags(xs), 1+3; got != want {
		t.Errorf("SumPaymentTags() = %d, want %d", got, want)
	}
	SortPaymentTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentTags() put %q first", xs[0].ID)
	}
	if len(IndexPaymentTags(xs)) != 3 {
		t.Errorf("IndexPaymentTags() = %v", IndexPaymentTags(xs))
	}
	if err := ValidatePaymentTags(ActivePaymentTags(xs)); err != nil {
		t.Errorf("ValidatePaymentTags() = %v", err)
	}
}

func TestPaymentTagNormalizeClamps(t *testing.T) {
	x := SamplePaymentTag("a", when)
	x.Channel = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentTag("r1", when)
	rev, err := r.PutPaymentTag(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentTag() = %v", err)
	}
	got, gotRev, err := r.GetPaymentTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentTags() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentTag(PaymentTag{}, 0); err == nil {
		t.Error("PutPaymentTag() accepted an empty entity")
	}
}

func TestPaymentRuleRoundTrip(t *testing.T) {
	x := SamplePaymentRule("paymentrule-1", when)
	got, err := ParsePaymentRule(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentRule() = %v", err)
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

func TestPaymentRuleParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentRule(f); err == nil {
		t.Error("ParsePaymentRule() accepted a record with no id")
	}
	f = SamplePaymentRule("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParsePaymentRule(f); err == nil {
		t.Error("ParsePaymentRule() accepted a non-numeric weight")
	}
}

func TestPaymentRuleCollections(t *testing.T) {
	xs := []PaymentRule{SamplePaymentRule("a", when), SamplePaymentRule("bb", when), SamplePaymentRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentRules(xs)); got != 2 {
		t.Errorf("ActivePaymentRules() kept %d, want 2", got)
	}
	if got, want := SumPaymentRules(xs), 1+3; got != want {
		t.Errorf("SumPaymentRules() = %d, want %d", got, want)
	}
	SortPaymentRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentRules() put %q first", xs[0].ID)
	}
	if len(IndexPaymentRules(xs)) != 3 {
		t.Errorf("IndexPaymentRules() = %v", IndexPaymentRules(xs))
	}
	if err := ValidatePaymentRules(ActivePaymentRules(xs)); err != nil {
		t.Errorf("ValidatePaymentRules() = %v", err)
	}
}

func TestPaymentRuleNormalizeClamps(t *testing.T) {
	x := SamplePaymentRule("a", when)
	x.Group = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentRule("r1", when)
	rev, err := r.PutPaymentRule(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentRule() = %v", err)
	}
	got, gotRev, err := r.GetPaymentRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentRules() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentRule(PaymentRule{}, 0); err == nil {
		t.Error("PutPaymentRule() accepted an empty entity")
	}
}

func TestPaymentHoldRoundTrip(t *testing.T) {
	x := SamplePaymentHold("paymenthold-1", when)
	got, err := ParsePaymentHold(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentHold() = %v", err)
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

func TestPaymentHoldParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentHold(f); err == nil {
		t.Error("ParsePaymentHold() accepted a record with no id")
	}
	f = SamplePaymentHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParsePaymentHold(f); err == nil {
		t.Error("ParsePaymentHold() accepted a non-numeric weight")
	}
}

func TestPaymentHoldCollections(t *testing.T) {
	xs := []PaymentHold{SamplePaymentHold("a", when), SamplePaymentHold("bb", when), SamplePaymentHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentHolds(xs)); got != 2 {
		t.Errorf("ActivePaymentHolds() kept %d, want 2", got)
	}
	if got, want := SumPaymentHolds(xs), 1+3; got != want {
		t.Errorf("SumPaymentHolds() = %d, want %d", got, want)
	}
	SortPaymentHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentHolds() put %q first", xs[0].ID)
	}
	if len(IndexPaymentHolds(xs)) != 3 {
		t.Errorf("IndexPaymentHolds() = %v", IndexPaymentHolds(xs))
	}
	if err := ValidatePaymentHolds(ActivePaymentHolds(xs)); err != nil {
		t.Errorf("ValidatePaymentHolds() = %v", err)
	}
}

func TestPaymentHoldNormalizeClamps(t *testing.T) {
	x := SamplePaymentHold("a", when)
	x.Comment = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentHold("r1", when)
	rev, err := r.PutPaymentHold(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentHold() = %v", err)
	}
	got, gotRev, err := r.GetPaymentHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentHolds() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentHold(PaymentHold{}, 0); err == nil {
		t.Error("PutPaymentHold() accepted an empty entity")
	}
}

func TestPaymentClaimRoundTrip(t *testing.T) {
	x := SamplePaymentClaim("paymentclaim-1", when)
	got, err := ParsePaymentClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentClaim() = %v", err)
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

func TestPaymentClaimParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentClaim(f); err == nil {
		t.Error("ParsePaymentClaim() accepted a record with no id")
	}
	f = SamplePaymentClaim("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParsePaymentClaim(f); err == nil {
		t.Error("ParsePaymentClaim() accepted a non-numeric capacity")
	}
}

func TestPaymentClaimCollections(t *testing.T) {
	xs := []PaymentClaim{SamplePaymentClaim("a", when), SamplePaymentClaim("bb", when), SamplePaymentClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentClaims(xs)); got != 2 {
		t.Errorf("ActivePaymentClaims() kept %d, want 2", got)
	}
	if got, want := SumPaymentClaims(xs), 1+3; got != want {
		t.Errorf("SumPaymentClaims() = %d, want %d", got, want)
	}
	SortPaymentClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentClaims() put %q first", xs[0].ID)
	}
	if len(IndexPaymentClaims(xs)) != 3 {
		t.Errorf("IndexPaymentClaims() = %v", IndexPaymentClaims(xs))
	}
	if err := ValidatePaymentClaims(ActivePaymentClaims(xs)); err != nil {
		t.Errorf("ValidatePaymentClaims() = %v", err)
	}
}

func TestPaymentClaimNormalizeClamps(t *testing.T) {
	x := SamplePaymentClaim("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentClaim("r1", when)
	rev, err := r.PutPaymentClaim(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentClaim() = %v", err)
	}
	got, gotRev, err := r.GetPaymentClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentClaims() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentClaim(PaymentClaim{}, 0); err == nil {
		t.Error("PutPaymentClaim() accepted an empty entity")
	}
}

func TestPaymentCursorRoundTrip(t *testing.T) {
	x := SamplePaymentCursor("paymentcursor-1", when)
	got, err := ParsePaymentCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParsePaymentCursor() = %v", err)
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

func TestPaymentCursorParseRejectsBadInput(t *testing.T) {
	f := SamplePaymentCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParsePaymentCursor(f); err == nil {
		t.Error("ParsePaymentCursor() accepted a record with no id")
	}
	f = SamplePaymentCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParsePaymentCursor(f); err == nil {
		t.Error("ParsePaymentCursor() accepted a non-numeric version")
	}
}

func TestPaymentCursorCollections(t *testing.T) {
	xs := []PaymentCursor{SamplePaymentCursor("a", when), SamplePaymentCursor("bb", when), SamplePaymentCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActivePaymentCursors(xs)); got != 2 {
		t.Errorf("ActivePaymentCursors() kept %d, want 2", got)
	}
	if got, want := SumPaymentCursors(xs), 1+3; got != want {
		t.Errorf("SumPaymentCursors() = %d, want %d", got, want)
	}
	SortPaymentCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortPaymentCursors() put %q first", xs[0].ID)
	}
	if len(IndexPaymentCursors(xs)) != 3 {
		t.Errorf("IndexPaymentCursors() = %v", IndexPaymentCursors(xs))
	}
	if err := ValidatePaymentCursors(ActivePaymentCursors(xs)); err != nil {
		t.Errorf("ValidatePaymentCursors() = %v", err)
	}
}

func TestPaymentCursorNormalizeClamps(t *testing.T) {
	x := SamplePaymentCursor("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestPaymentCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SamplePaymentCursor("r1", when)
	rev, err := r.PutPaymentCursor(x, 0)
	if err != nil {
		t.Fatalf("PutPaymentCursor() = %v", err)
	}
	got, gotRev, err := r.GetPaymentCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetPaymentCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListPaymentCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListPaymentCursors() = %v, %v", all, err)
	}
	if _, err := r.PutPaymentCursor(PaymentCursor{}, 0); err == nil {
		t.Error("PutPaymentCursor() accepted an empty entity")
	}
}
