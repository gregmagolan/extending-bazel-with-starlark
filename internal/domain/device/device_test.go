package device

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
		if !strings.HasPrefix(Prefixes[i], "device/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleDeviceRecord("ok", when).Fields()); err != nil {
		t.Errorf("Check(%q) = %v", Schemas[0].Name, err)
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

func TestDeviceRecordRoundTrip(t *testing.T) {
	x := SampleDeviceRecord("devicerecord-1", when)
	got, err := ParseDeviceRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceRecord() = %v", err)
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

func TestDeviceRecordParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceRecord(f); err == nil {
		t.Error("ParseDeviceRecord() accepted a record with no id")
	}
	f = SampleDeviceRecord("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeviceRecord(f); err == nil {
		t.Error("ParseDeviceRecord() accepted a non-numeric priority")
	}
}

func TestDeviceRecordCollections(t *testing.T) {
	xs := []DeviceRecord{SampleDeviceRecord("a", when), SampleDeviceRecord("bb", when), SampleDeviceRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceRecords(xs)); got != 2 {
		t.Errorf("ActiveDeviceRecords() kept %d, want 2", got)
	}
	if got, want := SumDeviceRecords(xs), 1+3; got != want {
		t.Errorf("SumDeviceRecords() = %d, want %d", got, want)
	}
	SortDeviceRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceRecords() put %q first", xs[0].ID)
	}
	if len(IndexDeviceRecords(xs)) != 3 {
		t.Errorf("IndexDeviceRecords() = %v", IndexDeviceRecords(xs))
	}
	if err := ValidateDeviceRecords(ActiveDeviceRecords(xs)); err != nil {
		t.Errorf("ValidateDeviceRecords() = %v", err)
	}
}

func TestDeviceRecordNormalizeClamps(t *testing.T) {
	x := SampleDeviceRecord("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceRecord("r1", when)
	rev, err := r.PutDeviceRecord(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceRecord() = %v", err)
	}
	got, gotRev, err := r.GetDeviceRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceRecords() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceRecord(DeviceRecord{}, 0); err == nil {
		t.Error("PutDeviceRecord() accepted an empty entity")
	}
}

func TestDeviceEntryRoundTrip(t *testing.T) {
	x := SampleDeviceEntry("deviceentry-1", when)
	got, err := ParseDeviceEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceEntry() = %v", err)
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

func TestDeviceEntryParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceEntry(f); err == nil {
		t.Error("ParseDeviceEntry() accepted a record with no id")
	}
	f = SampleDeviceEntry("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeviceEntry(f); err == nil {
		t.Error("ParseDeviceEntry() accepted a non-numeric version")
	}
}

func TestDeviceEntryCollections(t *testing.T) {
	xs := []DeviceEntry{SampleDeviceEntry("a", when), SampleDeviceEntry("bb", when), SampleDeviceEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceEntrys(xs)); got != 2 {
		t.Errorf("ActiveDeviceEntrys() kept %d, want 2", got)
	}
	if got, want := SumDeviceEntrys(xs), 1+3; got != want {
		t.Errorf("SumDeviceEntrys() = %d, want %d", got, want)
	}
	SortDeviceEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceEntrys() put %q first", xs[0].ID)
	}
	if len(IndexDeviceEntrys(xs)) != 3 {
		t.Errorf("IndexDeviceEntrys() = %v", IndexDeviceEntrys(xs))
	}
	if err := ValidateDeviceEntrys(ActiveDeviceEntrys(xs)); err != nil {
		t.Errorf("ValidateDeviceEntrys() = %v", err)
	}
}

func TestDeviceEntryNormalizeClamps(t *testing.T) {
	x := SampleDeviceEntry("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceEntry("r1", when)
	rev, err := r.PutDeviceEntry(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceEntry() = %v", err)
	}
	got, gotRev, err := r.GetDeviceEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceEntry(DeviceEntry{}, 0); err == nil {
		t.Error("PutDeviceEntry() accepted an empty entity")
	}
}

func TestDeviceItemRoundTrip(t *testing.T) {
	x := SampleDeviceItem("deviceitem-1", when)
	got, err := ParseDeviceItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceItem() = %v", err)
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

func TestDeviceItemParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceItem(f); err == nil {
		t.Error("ParseDeviceItem() accepted a record with no id")
	}
	f = SampleDeviceItem("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDeviceItem(f); err == nil {
		t.Error("ParseDeviceItem() accepted a non-numeric retained")
	}
}

func TestDeviceItemCollections(t *testing.T) {
	xs := []DeviceItem{SampleDeviceItem("a", when), SampleDeviceItem("bb", when), SampleDeviceItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceItems(xs)); got != 2 {
		t.Errorf("ActiveDeviceItems() kept %d, want 2", got)
	}
	if got, want := SumDeviceItems(xs), 1+3; got != want {
		t.Errorf("SumDeviceItems() = %d, want %d", got, want)
	}
	SortDeviceItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceItems() put %q first", xs[0].ID)
	}
	if len(IndexDeviceItems(xs)) != 3 {
		t.Errorf("IndexDeviceItems() = %v", IndexDeviceItems(xs))
	}
	if err := ValidateDeviceItems(ActiveDeviceItems(xs)); err != nil {
		t.Errorf("ValidateDeviceItems() = %v", err)
	}
}

func TestDeviceItemNormalizeClamps(t *testing.T) {
	x := SampleDeviceItem("a", when)
	x.Reference = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceItem("r1", when)
	rev, err := r.PutDeviceItem(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceItem() = %v", err)
	}
	got, gotRev, err := r.GetDeviceItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceItems() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceItem(DeviceItem{}, 0); err == nil {
		t.Error("PutDeviceItem() accepted an empty entity")
	}
}

func TestDeviceDraftRoundTrip(t *testing.T) {
	x := SampleDeviceDraft("devicedraft-1", when)
	got, err := ParseDeviceDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceDraft() = %v", err)
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

func TestDeviceDraftParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceDraft(f); err == nil {
		t.Error("ParseDeviceDraft() accepted a record with no id")
	}
	f = SampleDeviceDraft("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDeviceDraft(f); err == nil {
		t.Error("ParseDeviceDraft() accepted a non-numeric retained")
	}
}

func TestDeviceDraftCollections(t *testing.T) {
	xs := []DeviceDraft{SampleDeviceDraft("a", when), SampleDeviceDraft("bb", when), SampleDeviceDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceDrafts(xs)); got != 2 {
		t.Errorf("ActiveDeviceDrafts() kept %d, want 2", got)
	}
	if got, want := SumDeviceDrafts(xs), 1+3; got != want {
		t.Errorf("SumDeviceDrafts() = %d, want %d", got, want)
	}
	SortDeviceDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceDrafts() put %q first", xs[0].ID)
	}
	if len(IndexDeviceDrafts(xs)) != 3 {
		t.Errorf("IndexDeviceDrafts() = %v", IndexDeviceDrafts(xs))
	}
	if err := ValidateDeviceDrafts(ActiveDeviceDrafts(xs)); err != nil {
		t.Errorf("ValidateDeviceDrafts() = %v", err)
	}
}

func TestDeviceDraftNormalizeClamps(t *testing.T) {
	x := SampleDeviceDraft("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceDraft("r1", when)
	rev, err := r.PutDeviceDraft(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceDraft() = %v", err)
	}
	got, gotRev, err := r.GetDeviceDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceDraft(DeviceDraft{}, 0); err == nil {
		t.Error("PutDeviceDraft() accepted an empty entity")
	}
}

func TestDeviceRevisionRoundTrip(t *testing.T) {
	x := SampleDeviceRevision("devicerevision-1", when)
	got, err := ParseDeviceRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceRevision() = %v", err)
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

func TestDeviceRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceRevision(f); err == nil {
		t.Error("ParseDeviceRevision() accepted a record with no id")
	}
	f = SampleDeviceRevision("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDeviceRevision(f); err == nil {
		t.Error("ParseDeviceRevision() accepted a non-numeric capacity")
	}
}

func TestDeviceRevisionCollections(t *testing.T) {
	xs := []DeviceRevision{SampleDeviceRevision("a", when), SampleDeviceRevision("bb", when), SampleDeviceRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceRevisions(xs)); got != 2 {
		t.Errorf("ActiveDeviceRevisions() kept %d, want 2", got)
	}
	if got, want := SumDeviceRevisions(xs), 1+3; got != want {
		t.Errorf("SumDeviceRevisions() = %d, want %d", got, want)
	}
	SortDeviceRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceRevisions() put %q first", xs[0].ID)
	}
	if len(IndexDeviceRevisions(xs)) != 3 {
		t.Errorf("IndexDeviceRevisions() = %v", IndexDeviceRevisions(xs))
	}
	if err := ValidateDeviceRevisions(ActiveDeviceRevisions(xs)); err != nil {
		t.Errorf("ValidateDeviceRevisions() = %v", err)
	}
}

func TestDeviceRevisionNormalizeClamps(t *testing.T) {
	x := SampleDeviceRevision("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceRevision("r1", when)
	rev, err := r.PutDeviceRevision(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceRevision() = %v", err)
	}
	got, gotRev, err := r.GetDeviceRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceRevision(DeviceRevision{}, 0); err == nil {
		t.Error("PutDeviceRevision() accepted an empty entity")
	}
}

func TestDeviceSnapshotRoundTrip(t *testing.T) {
	x := SampleDeviceSnapshot("devicesnapshot-1", when)
	got, err := ParseDeviceSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceSnapshot() = %v", err)
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

func TestDeviceSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceSnapshot(f); err == nil {
		t.Error("ParseDeviceSnapshot() accepted a record with no id")
	}
	f = SampleDeviceSnapshot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDeviceSnapshot(f); err == nil {
		t.Error("ParseDeviceSnapshot() accepted a non-numeric count")
	}
}

func TestDeviceSnapshotCollections(t *testing.T) {
	xs := []DeviceSnapshot{SampleDeviceSnapshot("a", when), SampleDeviceSnapshot("bb", when), SampleDeviceSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceSnapshots(xs)); got != 2 {
		t.Errorf("ActiveDeviceSnapshots() kept %d, want 2", got)
	}
	if got, want := SumDeviceSnapshots(xs), 1+3; got != want {
		t.Errorf("SumDeviceSnapshots() = %d, want %d", got, want)
	}
	SortDeviceSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexDeviceSnapshots(xs)) != 3 {
		t.Errorf("IndexDeviceSnapshots() = %v", IndexDeviceSnapshots(xs))
	}
	if err := ValidateDeviceSnapshots(ActiveDeviceSnapshots(xs)); err != nil {
		t.Errorf("ValidateDeviceSnapshots() = %v", err)
	}
}

func TestDeviceSnapshotNormalizeClamps(t *testing.T) {
	x := SampleDeviceSnapshot("a", when)
	x.Comment = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceSnapshot("r1", when)
	rev, err := r.PutDeviceSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetDeviceSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceSnapshot(DeviceSnapshot{}, 0); err == nil {
		t.Error("PutDeviceSnapshot() accepted an empty entity")
	}
}

func TestDeviceBatchRoundTrip(t *testing.T) {
	x := SampleDeviceBatch("devicebatch-1", when)
	got, err := ParseDeviceBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceBatch() = %v", err)
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

func TestDeviceBatchParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceBatch(f); err == nil {
		t.Error("ParseDeviceBatch() accepted a record with no id")
	}
	f = SampleDeviceBatch("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseDeviceBatch(f); err == nil {
		t.Error("ParseDeviceBatch() accepted a non-numeric count")
	}
}

func TestDeviceBatchCollections(t *testing.T) {
	xs := []DeviceBatch{SampleDeviceBatch("a", when), SampleDeviceBatch("bb", when), SampleDeviceBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceBatchs(xs)); got != 2 {
		t.Errorf("ActiveDeviceBatchs() kept %d, want 2", got)
	}
	if got, want := SumDeviceBatchs(xs), 1+3; got != want {
		t.Errorf("SumDeviceBatchs() = %d, want %d", got, want)
	}
	SortDeviceBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceBatchs() put %q first", xs[0].ID)
	}
	if len(IndexDeviceBatchs(xs)) != 3 {
		t.Errorf("IndexDeviceBatchs() = %v", IndexDeviceBatchs(xs))
	}
	if err := ValidateDeviceBatchs(ActiveDeviceBatchs(xs)); err != nil {
		t.Errorf("ValidateDeviceBatchs() = %v", err)
	}
}

func TestDeviceBatchNormalizeClamps(t *testing.T) {
	x := SampleDeviceBatch("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceBatch("r1", when)
	rev, err := r.PutDeviceBatch(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceBatch() = %v", err)
	}
	got, gotRev, err := r.GetDeviceBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceBatch(DeviceBatch{}, 0); err == nil {
		t.Error("PutDeviceBatch() accepted an empty entity")
	}
}

func TestDeviceRequestRoundTrip(t *testing.T) {
	x := SampleDeviceRequest("devicerequest-1", when)
	got, err := ParseDeviceRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceRequest() = %v", err)
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

func TestDeviceRequestParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceRequest(f); err == nil {
		t.Error("ParseDeviceRequest() accepted a record with no id")
	}
	f = SampleDeviceRequest("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseDeviceRequest(f); err == nil {
		t.Error("ParseDeviceRequest() accepted a non-numeric offset")
	}
}

func TestDeviceRequestCollections(t *testing.T) {
	xs := []DeviceRequest{SampleDeviceRequest("a", when), SampleDeviceRequest("bb", when), SampleDeviceRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceRequests(xs)); got != 2 {
		t.Errorf("ActiveDeviceRequests() kept %d, want 2", got)
	}
	if got, want := SumDeviceRequests(xs), 1+3; got != want {
		t.Errorf("SumDeviceRequests() = %d, want %d", got, want)
	}
	SortDeviceRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceRequests() put %q first", xs[0].ID)
	}
	if len(IndexDeviceRequests(xs)) != 3 {
		t.Errorf("IndexDeviceRequests() = %v", IndexDeviceRequests(xs))
	}
	if err := ValidateDeviceRequests(ActiveDeviceRequests(xs)); err != nil {
		t.Errorf("ValidateDeviceRequests() = %v", err)
	}
}

func TestDeviceRequestNormalizeClamps(t *testing.T) {
	x := SampleDeviceRequest("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceRequest("r1", when)
	rev, err := r.PutDeviceRequest(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceRequest() = %v", err)
	}
	got, gotRev, err := r.GetDeviceRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceRequests() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceRequest(DeviceRequest{}, 0); err == nil {
		t.Error("PutDeviceRequest() accepted an empty entity")
	}
}

func TestDeviceSummaryRoundTrip(t *testing.T) {
	x := SampleDeviceSummary("devicesummary-1", when)
	got, err := ParseDeviceSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceSummary() = %v", err)
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

func TestDeviceSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceSummary(f); err == nil {
		t.Error("ParseDeviceSummary() accepted a record with no id")
	}
	f = SampleDeviceSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDeviceSummary(f); err == nil {
		t.Error("ParseDeviceSummary() accepted a non-numeric capacity")
	}
}

func TestDeviceSummaryCollections(t *testing.T) {
	xs := []DeviceSummary{SampleDeviceSummary("a", when), SampleDeviceSummary("bb", when), SampleDeviceSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceSummarys(xs)); got != 2 {
		t.Errorf("ActiveDeviceSummarys() kept %d, want 2", got)
	}
	if got, want := SumDeviceSummarys(xs), 1+3; got != want {
		t.Errorf("SumDeviceSummarys() = %d, want %d", got, want)
	}
	SortDeviceSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceSummarys() put %q first", xs[0].ID)
	}
	if len(IndexDeviceSummarys(xs)) != 3 {
		t.Errorf("IndexDeviceSummarys() = %v", IndexDeviceSummarys(xs))
	}
	if err := ValidateDeviceSummarys(ActiveDeviceSummarys(xs)); err != nil {
		t.Errorf("ValidateDeviceSummarys() = %v", err)
	}
}

func TestDeviceSummaryNormalizeClamps(t *testing.T) {
	x := SampleDeviceSummary("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceSummary("r1", when)
	rev, err := r.PutDeviceSummary(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceSummary() = %v", err)
	}
	got, gotRev, err := r.GetDeviceSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceSummary(DeviceSummary{}, 0); err == nil {
		t.Error("PutDeviceSummary() accepted an empty entity")
	}
}

func TestDevicePolicyRoundTrip(t *testing.T) {
	x := SampleDevicePolicy("devicepolicy-1", when)
	got, err := ParseDevicePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseDevicePolicy() = %v", err)
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

func TestDevicePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleDevicePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDevicePolicy(f); err == nil {
		t.Error("ParseDevicePolicy() accepted a record with no id")
	}
	f = SampleDevicePolicy("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDevicePolicy(f); err == nil {
		t.Error("ParseDevicePolicy() accepted a non-numeric weight")
	}
}

func TestDevicePolicyCollections(t *testing.T) {
	xs := []DevicePolicy{SampleDevicePolicy("a", when), SampleDevicePolicy("bb", when), SampleDevicePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDevicePolicys(xs)); got != 2 {
		t.Errorf("ActiveDevicePolicys() kept %d, want 2", got)
	}
	if got, want := SumDevicePolicys(xs), 1+3; got != want {
		t.Errorf("SumDevicePolicys() = %d, want %d", got, want)
	}
	SortDevicePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDevicePolicys() put %q first", xs[0].ID)
	}
	if len(IndexDevicePolicys(xs)) != 3 {
		t.Errorf("IndexDevicePolicys() = %v", IndexDevicePolicys(xs))
	}
	if err := ValidateDevicePolicys(ActiveDevicePolicys(xs)); err != nil {
		t.Errorf("ValidateDevicePolicys() = %v", err)
	}
}

func TestDevicePolicyNormalizeClamps(t *testing.T) {
	x := SampleDevicePolicy("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDevicePolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDevicePolicy("r1", when)
	rev, err := r.PutDevicePolicy(x, 0)
	if err != nil {
		t.Fatalf("PutDevicePolicy() = %v", err)
	}
	got, gotRev, err := r.GetDevicePolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDevicePolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDevicePolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDevicePolicys() = %v, %v", all, err)
	}
	if _, err := r.PutDevicePolicy(DevicePolicy{}, 0); err == nil {
		t.Error("PutDevicePolicy() accepted an empty entity")
	}
}

func TestDeviceLinkRoundTrip(t *testing.T) {
	x := SampleDeviceLink("devicelink-1", when)
	got, err := ParseDeviceLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceLink() = %v", err)
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

func TestDeviceLinkParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceLink(f); err == nil {
		t.Error("ParseDeviceLink() accepted a record with no id")
	}
	f = SampleDeviceLink("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDeviceLink(f); err == nil {
		t.Error("ParseDeviceLink() accepted a non-numeric capacity")
	}
}

func TestDeviceLinkCollections(t *testing.T) {
	xs := []DeviceLink{SampleDeviceLink("a", when), SampleDeviceLink("bb", when), SampleDeviceLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceLinks(xs)); got != 2 {
		t.Errorf("ActiveDeviceLinks() kept %d, want 2", got)
	}
	if got, want := SumDeviceLinks(xs), 1+3; got != want {
		t.Errorf("SumDeviceLinks() = %d, want %d", got, want)
	}
	SortDeviceLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceLinks() put %q first", xs[0].ID)
	}
	if len(IndexDeviceLinks(xs)) != 3 {
		t.Errorf("IndexDeviceLinks() = %v", IndexDeviceLinks(xs))
	}
	if err := ValidateDeviceLinks(ActiveDeviceLinks(xs)); err != nil {
		t.Errorf("ValidateDeviceLinks() = %v", err)
	}
}

func TestDeviceLinkNormalizeClamps(t *testing.T) {
	x := SampleDeviceLink("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceLink("r1", when)
	rev, err := r.PutDeviceLink(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceLink() = %v", err)
	}
	got, gotRev, err := r.GetDeviceLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceLinks() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceLink(DeviceLink{}, 0); err == nil {
		t.Error("PutDeviceLink() accepted an empty entity")
	}
}

func TestDeviceNoteRoundTrip(t *testing.T) {
	x := SampleDeviceNote("devicenote-1", when)
	got, err := ParseDeviceNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceNote() = %v", err)
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

func TestDeviceNoteParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceNote(f); err == nil {
		t.Error("ParseDeviceNote() accepted a record with no id")
	}
	f = SampleDeviceNote("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseDeviceNote(f); err == nil {
		t.Error("ParseDeviceNote() accepted a non-numeric retained")
	}
}

func TestDeviceNoteCollections(t *testing.T) {
	xs := []DeviceNote{SampleDeviceNote("a", when), SampleDeviceNote("bb", when), SampleDeviceNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceNotes(xs)); got != 2 {
		t.Errorf("ActiveDeviceNotes() kept %d, want 2", got)
	}
	if got, want := SumDeviceNotes(xs), 1+3; got != want {
		t.Errorf("SumDeviceNotes() = %d, want %d", got, want)
	}
	SortDeviceNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceNotes() put %q first", xs[0].ID)
	}
	if len(IndexDeviceNotes(xs)) != 3 {
		t.Errorf("IndexDeviceNotes() = %v", IndexDeviceNotes(xs))
	}
	if err := ValidateDeviceNotes(ActiveDeviceNotes(xs)); err != nil {
		t.Errorf("ValidateDeviceNotes() = %v", err)
	}
}

func TestDeviceNoteNormalizeClamps(t *testing.T) {
	x := SampleDeviceNote("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceNote("r1", when)
	rev, err := r.PutDeviceNote(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceNote() = %v", err)
	}
	got, gotRev, err := r.GetDeviceNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceNotes() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceNote(DeviceNote{}, 0); err == nil {
		t.Error("PutDeviceNote() accepted an empty entity")
	}
}

func TestDeviceWindowRoundTrip(t *testing.T) {
	x := SampleDeviceWindow("devicewindow-1", when)
	got, err := ParseDeviceWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceWindow() = %v", err)
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

func TestDeviceWindowParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceWindow(f); err == nil {
		t.Error("ParseDeviceWindow() accepted a record with no id")
	}
	f = SampleDeviceWindow("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeviceWindow(f); err == nil {
		t.Error("ParseDeviceWindow() accepted a non-numeric version")
	}
}

func TestDeviceWindowCollections(t *testing.T) {
	xs := []DeviceWindow{SampleDeviceWindow("a", when), SampleDeviceWindow("bb", when), SampleDeviceWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceWindows(xs)); got != 2 {
		t.Errorf("ActiveDeviceWindows() kept %d, want 2", got)
	}
	if got, want := SumDeviceWindows(xs), 1+3; got != want {
		t.Errorf("SumDeviceWindows() = %d, want %d", got, want)
	}
	SortDeviceWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceWindows() put %q first", xs[0].ID)
	}
	if len(IndexDeviceWindows(xs)) != 3 {
		t.Errorf("IndexDeviceWindows() = %v", IndexDeviceWindows(xs))
	}
	if err := ValidateDeviceWindows(ActiveDeviceWindows(xs)); err != nil {
		t.Errorf("ValidateDeviceWindows() = %v", err)
	}
}

func TestDeviceWindowNormalizeClamps(t *testing.T) {
	x := SampleDeviceWindow("a", when)
	x.Group = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceWindow("r1", when)
	rev, err := r.PutDeviceWindow(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceWindow() = %v", err)
	}
	got, gotRev, err := r.GetDeviceWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceWindows() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceWindow(DeviceWindow{}, 0); err == nil {
		t.Error("PutDeviceWindow() accepted an empty entity")
	}
}

func TestDeviceQuotaRoundTrip(t *testing.T) {
	x := SampleDeviceQuota("devicequota-1", when)
	got, err := ParseDeviceQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceQuota() = %v", err)
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

func TestDeviceQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceQuota(f); err == nil {
		t.Error("ParseDeviceQuota() accepted a record with no id")
	}
	f = SampleDeviceQuota("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDeviceQuota(f); err == nil {
		t.Error("ParseDeviceQuota() accepted a non-numeric capacity")
	}
}

func TestDeviceQuotaCollections(t *testing.T) {
	xs := []DeviceQuota{SampleDeviceQuota("a", when), SampleDeviceQuota("bb", when), SampleDeviceQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceQuotas(xs)); got != 2 {
		t.Errorf("ActiveDeviceQuotas() kept %d, want 2", got)
	}
	if got, want := SumDeviceQuotas(xs), 1+3; got != want {
		t.Errorf("SumDeviceQuotas() = %d, want %d", got, want)
	}
	SortDeviceQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceQuotas() put %q first", xs[0].ID)
	}
	if len(IndexDeviceQuotas(xs)) != 3 {
		t.Errorf("IndexDeviceQuotas() = %v", IndexDeviceQuotas(xs))
	}
	if err := ValidateDeviceQuotas(ActiveDeviceQuotas(xs)); err != nil {
		t.Errorf("ValidateDeviceQuotas() = %v", err)
	}
}

func TestDeviceQuotaNormalizeClamps(t *testing.T) {
	x := SampleDeviceQuota("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceQuota("r1", when)
	rev, err := r.PutDeviceQuota(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceQuota() = %v", err)
	}
	got, gotRev, err := r.GetDeviceQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceQuota(DeviceQuota{}, 0); err == nil {
		t.Error("PutDeviceQuota() accepted an empty entity")
	}
}

func TestDeviceBucketRoundTrip(t *testing.T) {
	x := SampleDeviceBucket("devicebucket-1", when)
	got, err := ParseDeviceBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceBucket() = %v", err)
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

func TestDeviceBucketParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceBucket(f); err == nil {
		t.Error("ParseDeviceBucket() accepted a record with no id")
	}
	f = SampleDeviceBucket("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDeviceBucket(f); err == nil {
		t.Error("ParseDeviceBucket() accepted a non-numeric weight")
	}
}

func TestDeviceBucketCollections(t *testing.T) {
	xs := []DeviceBucket{SampleDeviceBucket("a", when), SampleDeviceBucket("bb", when), SampleDeviceBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceBuckets(xs)); got != 2 {
		t.Errorf("ActiveDeviceBuckets() kept %d, want 2", got)
	}
	if got, want := SumDeviceBuckets(xs), 1+3; got != want {
		t.Errorf("SumDeviceBuckets() = %d, want %d", got, want)
	}
	SortDeviceBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceBuckets() put %q first", xs[0].ID)
	}
	if len(IndexDeviceBuckets(xs)) != 3 {
		t.Errorf("IndexDeviceBuckets() = %v", IndexDeviceBuckets(xs))
	}
	if err := ValidateDeviceBuckets(ActiveDeviceBuckets(xs)); err != nil {
		t.Errorf("ValidateDeviceBuckets() = %v", err)
	}
}

func TestDeviceBucketNormalizeClamps(t *testing.T) {
	x := SampleDeviceBucket("a", when)
	x.Label = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceBucket("r1", when)
	rev, err := r.PutDeviceBucket(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceBucket() = %v", err)
	}
	got, gotRev, err := r.GetDeviceBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceBucket(DeviceBucket{}, 0); err == nil {
		t.Error("PutDeviceBucket() accepted an empty entity")
	}
}

func TestDeviceLaneRoundTrip(t *testing.T) {
	x := SampleDeviceLane("devicelane-1", when)
	got, err := ParseDeviceLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceLane() = %v", err)
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

func TestDeviceLaneParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceLane(f); err == nil {
		t.Error("ParseDeviceLane() accepted a record with no id")
	}
	f = SampleDeviceLane("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseDeviceLane(f); err == nil {
		t.Error("ParseDeviceLane() accepted a non-numeric priority")
	}
}

func TestDeviceLaneCollections(t *testing.T) {
	xs := []DeviceLane{SampleDeviceLane("a", when), SampleDeviceLane("bb", when), SampleDeviceLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceLanes(xs)); got != 2 {
		t.Errorf("ActiveDeviceLanes() kept %d, want 2", got)
	}
	if got, want := SumDeviceLanes(xs), 1+3; got != want {
		t.Errorf("SumDeviceLanes() = %d, want %d", got, want)
	}
	SortDeviceLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceLanes() put %q first", xs[0].ID)
	}
	if len(IndexDeviceLanes(xs)) != 3 {
		t.Errorf("IndexDeviceLanes() = %v", IndexDeviceLanes(xs))
	}
	if err := ValidateDeviceLanes(ActiveDeviceLanes(xs)); err != nil {
		t.Errorf("ValidateDeviceLanes() = %v", err)
	}
}

func TestDeviceLaneNormalizeClamps(t *testing.T) {
	x := SampleDeviceLane("a", when)
	x.Owner = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceLane("r1", when)
	rev, err := r.PutDeviceLane(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceLane() = %v", err)
	}
	got, gotRev, err := r.GetDeviceLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceLanes() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceLane(DeviceLane{}, 0); err == nil {
		t.Error("PutDeviceLane() accepted an empty entity")
	}
}

func TestDeviceSlotRoundTrip(t *testing.T) {
	x := SampleDeviceSlot("deviceslot-1", when)
	got, err := ParseDeviceSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceSlot() = %v", err)
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

func TestDeviceSlotParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceSlot(f); err == nil {
		t.Error("ParseDeviceSlot() accepted a record with no id")
	}
	f = SampleDeviceSlot("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDeviceSlot(f); err == nil {
		t.Error("ParseDeviceSlot() accepted a non-numeric attempts")
	}
}

func TestDeviceSlotCollections(t *testing.T) {
	xs := []DeviceSlot{SampleDeviceSlot("a", when), SampleDeviceSlot("bb", when), SampleDeviceSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceSlots(xs)); got != 2 {
		t.Errorf("ActiveDeviceSlots() kept %d, want 2", got)
	}
	if got, want := SumDeviceSlots(xs), 1+3; got != want {
		t.Errorf("SumDeviceSlots() = %d, want %d", got, want)
	}
	SortDeviceSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceSlots() put %q first", xs[0].ID)
	}
	if len(IndexDeviceSlots(xs)) != 3 {
		t.Errorf("IndexDeviceSlots() = %v", IndexDeviceSlots(xs))
	}
	if err := ValidateDeviceSlots(ActiveDeviceSlots(xs)); err != nil {
		t.Errorf("ValidateDeviceSlots() = %v", err)
	}
}

func TestDeviceSlotNormalizeClamps(t *testing.T) {
	x := SampleDeviceSlot("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceSlot("r1", when)
	rev, err := r.PutDeviceSlot(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceSlot() = %v", err)
	}
	got, gotRev, err := r.GetDeviceSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceSlots() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceSlot(DeviceSlot{}, 0); err == nil {
		t.Error("PutDeviceSlot() accepted an empty entity")
	}
}

func TestDeviceTagRoundTrip(t *testing.T) {
	x := SampleDeviceTag("devicetag-1", when)
	got, err := ParseDeviceTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceTag() = %v", err)
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

func TestDeviceTagParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceTag(f); err == nil {
		t.Error("ParseDeviceTag() accepted a record with no id")
	}
	f = SampleDeviceTag("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseDeviceTag(f); err == nil {
		t.Error("ParseDeviceTag() accepted a non-numeric attempts")
	}
}

func TestDeviceTagCollections(t *testing.T) {
	xs := []DeviceTag{SampleDeviceTag("a", when), SampleDeviceTag("bb", when), SampleDeviceTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceTags(xs)); got != 2 {
		t.Errorf("ActiveDeviceTags() kept %d, want 2", got)
	}
	if got, want := SumDeviceTags(xs), 1+3; got != want {
		t.Errorf("SumDeviceTags() = %d, want %d", got, want)
	}
	SortDeviceTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceTags() put %q first", xs[0].ID)
	}
	if len(IndexDeviceTags(xs)) != 3 {
		t.Errorf("IndexDeviceTags() = %v", IndexDeviceTags(xs))
	}
	if err := ValidateDeviceTags(ActiveDeviceTags(xs)); err != nil {
		t.Errorf("ValidateDeviceTags() = %v", err)
	}
}

func TestDeviceTagNormalizeClamps(t *testing.T) {
	x := SampleDeviceTag("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceTag("r1", when)
	rev, err := r.PutDeviceTag(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceTag() = %v", err)
	}
	got, gotRev, err := r.GetDeviceTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceTags() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceTag(DeviceTag{}, 0); err == nil {
		t.Error("PutDeviceTag() accepted an empty entity")
	}
}

func TestDeviceRuleRoundTrip(t *testing.T) {
	x := SampleDeviceRule("devicerule-1", when)
	got, err := ParseDeviceRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceRule() = %v", err)
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

func TestDeviceRuleParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceRule(f); err == nil {
		t.Error("ParseDeviceRule() accepted a record with no id")
	}
	f = SampleDeviceRule("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseDeviceRule(f); err == nil {
		t.Error("ParseDeviceRule() accepted a non-numeric weight")
	}
}

func TestDeviceRuleCollections(t *testing.T) {
	xs := []DeviceRule{SampleDeviceRule("a", when), SampleDeviceRule("bb", when), SampleDeviceRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceRules(xs)); got != 2 {
		t.Errorf("ActiveDeviceRules() kept %d, want 2", got)
	}
	if got, want := SumDeviceRules(xs), 1+3; got != want {
		t.Errorf("SumDeviceRules() = %d, want %d", got, want)
	}
	SortDeviceRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceRules() put %q first", xs[0].ID)
	}
	if len(IndexDeviceRules(xs)) != 3 {
		t.Errorf("IndexDeviceRules() = %v", IndexDeviceRules(xs))
	}
	if err := ValidateDeviceRules(ActiveDeviceRules(xs)); err != nil {
		t.Errorf("ValidateDeviceRules() = %v", err)
	}
}

func TestDeviceRuleNormalizeClamps(t *testing.T) {
	x := SampleDeviceRule("a", when)
	x.Owner = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceRule("r1", when)
	rev, err := r.PutDeviceRule(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceRule() = %v", err)
	}
	got, gotRev, err := r.GetDeviceRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceRules() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceRule(DeviceRule{}, 0); err == nil {
		t.Error("PutDeviceRule() accepted an empty entity")
	}
}

func TestDeviceHoldRoundTrip(t *testing.T) {
	x := SampleDeviceHold("devicehold-1", when)
	got, err := ParseDeviceHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceHold() = %v", err)
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

func TestDeviceHoldParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceHold(f); err == nil {
		t.Error("ParseDeviceHold() accepted a record with no id")
	}
	f = SampleDeviceHold("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeviceHold(f); err == nil {
		t.Error("ParseDeviceHold() accepted a non-numeric version")
	}
}

func TestDeviceHoldCollections(t *testing.T) {
	xs := []DeviceHold{SampleDeviceHold("a", when), SampleDeviceHold("bb", when), SampleDeviceHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceHolds(xs)); got != 2 {
		t.Errorf("ActiveDeviceHolds() kept %d, want 2", got)
	}
	if got, want := SumDeviceHolds(xs), 1+3; got != want {
		t.Errorf("SumDeviceHolds() = %d, want %d", got, want)
	}
	SortDeviceHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceHolds() put %q first", xs[0].ID)
	}
	if len(IndexDeviceHolds(xs)) != 3 {
		t.Errorf("IndexDeviceHolds() = %v", IndexDeviceHolds(xs))
	}
	if err := ValidateDeviceHolds(ActiveDeviceHolds(xs)); err != nil {
		t.Errorf("ValidateDeviceHolds() = %v", err)
	}
}

func TestDeviceHoldNormalizeClamps(t *testing.T) {
	x := SampleDeviceHold("a", when)
	x.Owner = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceHold("r1", when)
	rev, err := r.PutDeviceHold(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceHold() = %v", err)
	}
	got, gotRev, err := r.GetDeviceHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceHolds() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceHold(DeviceHold{}, 0); err == nil {
		t.Error("PutDeviceHold() accepted an empty entity")
	}
}

func TestDeviceClaimRoundTrip(t *testing.T) {
	x := SampleDeviceClaim("deviceclaim-1", when)
	got, err := ParseDeviceClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceClaim() = %v", err)
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

func TestDeviceClaimParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceClaim(f); err == nil {
		t.Error("ParseDeviceClaim() accepted a record with no id")
	}
	f = SampleDeviceClaim("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseDeviceClaim(f); err == nil {
		t.Error("ParseDeviceClaim() accepted a non-numeric capacity")
	}
}

func TestDeviceClaimCollections(t *testing.T) {
	xs := []DeviceClaim{SampleDeviceClaim("a", when), SampleDeviceClaim("bb", when), SampleDeviceClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceClaims(xs)); got != 2 {
		t.Errorf("ActiveDeviceClaims() kept %d, want 2", got)
	}
	if got, want := SumDeviceClaims(xs), 1+3; got != want {
		t.Errorf("SumDeviceClaims() = %d, want %d", got, want)
	}
	SortDeviceClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceClaims() put %q first", xs[0].ID)
	}
	if len(IndexDeviceClaims(xs)) != 3 {
		t.Errorf("IndexDeviceClaims() = %v", IndexDeviceClaims(xs))
	}
	if err := ValidateDeviceClaims(ActiveDeviceClaims(xs)); err != nil {
		t.Errorf("ValidateDeviceClaims() = %v", err)
	}
}

func TestDeviceClaimNormalizeClamps(t *testing.T) {
	x := SampleDeviceClaim("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceClaim("r1", when)
	rev, err := r.PutDeviceClaim(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceClaim() = %v", err)
	}
	got, gotRev, err := r.GetDeviceClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceClaims() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceClaim(DeviceClaim{}, 0); err == nil {
		t.Error("PutDeviceClaim() accepted an empty entity")
	}
}

func TestDeviceCursorRoundTrip(t *testing.T) {
	x := SampleDeviceCursor("devicecursor-1", when)
	got, err := ParseDeviceCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseDeviceCursor() = %v", err)
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

func TestDeviceCursorParseRejectsBadInput(t *testing.T) {
	f := SampleDeviceCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseDeviceCursor(f); err == nil {
		t.Error("ParseDeviceCursor() accepted a record with no id")
	}
	f = SampleDeviceCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseDeviceCursor(f); err == nil {
		t.Error("ParseDeviceCursor() accepted a non-numeric version")
	}
}

func TestDeviceCursorCollections(t *testing.T) {
	xs := []DeviceCursor{SampleDeviceCursor("a", when), SampleDeviceCursor("bb", when), SampleDeviceCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveDeviceCursors(xs)); got != 2 {
		t.Errorf("ActiveDeviceCursors() kept %d, want 2", got)
	}
	if got, want := SumDeviceCursors(xs), 1+3; got != want {
		t.Errorf("SumDeviceCursors() = %d, want %d", got, want)
	}
	SortDeviceCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortDeviceCursors() put %q first", xs[0].ID)
	}
	if len(IndexDeviceCursors(xs)) != 3 {
		t.Errorf("IndexDeviceCursors() = %v", IndexDeviceCursors(xs))
	}
	if err := ValidateDeviceCursors(ActiveDeviceCursors(xs)); err != nil {
		t.Errorf("ValidateDeviceCursors() = %v", err)
	}
}

func TestDeviceCursorNormalizeClamps(t *testing.T) {
	x := SampleDeviceCursor("a", when)
	x.Name = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestDeviceCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleDeviceCursor("r1", when)
	rev, err := r.PutDeviceCursor(x, 0)
	if err != nil {
		t.Fatalf("PutDeviceCursor() = %v", err)
	}
	got, gotRev, err := r.GetDeviceCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetDeviceCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListDeviceCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListDeviceCursors() = %v, %v", all, err)
	}
	if _, err := r.PutDeviceCursor(DeviceCursor{}, 0); err == nil {
		t.Error("PutDeviceCursor() accepted an empty entity")
	}
}
