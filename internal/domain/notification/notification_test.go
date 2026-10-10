package notification

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
		if !strings.HasPrefix(Prefixes[i], "notification/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleNotificationRecord("ok", when).Fields()); err != nil {
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

func TestNotificationRecordRoundTrip(t *testing.T) {
	x := SampleNotificationRecord("notificationrecord-1", when)
	got, err := ParseNotificationRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationRecord() = %v", err)
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

func TestNotificationRecordParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationRecord(f); err == nil {
		t.Error("ParseNotificationRecord() accepted a record with no id")
	}
	f = SampleNotificationRecord("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseNotificationRecord(f); err == nil {
		t.Error("ParseNotificationRecord() accepted a non-numeric version")
	}
}

func TestNotificationRecordCollections(t *testing.T) {
	xs := []NotificationRecord{SampleNotificationRecord("a", when), SampleNotificationRecord("bb", when), SampleNotificationRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationRecords(xs)); got != 2 {
		t.Errorf("ActiveNotificationRecords() kept %d, want 2", got)
	}
	if got, want := SumNotificationRecords(xs), 1+3; got != want {
		t.Errorf("SumNotificationRecords() = %d, want %d", got, want)
	}
	SortNotificationRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationRecords() put %q first", xs[0].ID)
	}
	if len(IndexNotificationRecords(xs)) != 3 {
		t.Errorf("IndexNotificationRecords() = %v", IndexNotificationRecords(xs))
	}
	if err := ValidateNotificationRecords(ActiveNotificationRecords(xs)); err != nil {
		t.Errorf("ValidateNotificationRecords() = %v", err)
	}
}

func TestNotificationRecordNormalizeClamps(t *testing.T) {
	x := SampleNotificationRecord("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationRecord("r1", when)
	rev, err := r.PutNotificationRecord(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationRecord() = %v", err)
	}
	got, gotRev, err := r.GetNotificationRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationRecords() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationRecord(NotificationRecord{}, 0); err == nil {
		t.Error("PutNotificationRecord() accepted an empty entity")
	}
}

func TestNotificationEntryRoundTrip(t *testing.T) {
	x := SampleNotificationEntry("notificationentry-1", when)
	got, err := ParseNotificationEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationEntry() = %v", err)
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

func TestNotificationEntryParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationEntry(f); err == nil {
		t.Error("ParseNotificationEntry() accepted a record with no id")
	}
	f = SampleNotificationEntry("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseNotificationEntry(f); err == nil {
		t.Error("ParseNotificationEntry() accepted a non-numeric retained")
	}
}

func TestNotificationEntryCollections(t *testing.T) {
	xs := []NotificationEntry{SampleNotificationEntry("a", when), SampleNotificationEntry("bb", when), SampleNotificationEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationEntrys(xs)); got != 2 {
		t.Errorf("ActiveNotificationEntrys() kept %d, want 2", got)
	}
	if got, want := SumNotificationEntrys(xs), 1+3; got != want {
		t.Errorf("SumNotificationEntrys() = %d, want %d", got, want)
	}
	SortNotificationEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationEntrys() put %q first", xs[0].ID)
	}
	if len(IndexNotificationEntrys(xs)) != 3 {
		t.Errorf("IndexNotificationEntrys() = %v", IndexNotificationEntrys(xs))
	}
	if err := ValidateNotificationEntrys(ActiveNotificationEntrys(xs)); err != nil {
		t.Errorf("ValidateNotificationEntrys() = %v", err)
	}
}

func TestNotificationEntryNormalizeClamps(t *testing.T) {
	x := SampleNotificationEntry("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationEntry("r1", when)
	rev, err := r.PutNotificationEntry(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationEntry() = %v", err)
	}
	got, gotRev, err := r.GetNotificationEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationEntry(NotificationEntry{}, 0); err == nil {
		t.Error("PutNotificationEntry() accepted an empty entity")
	}
}

func TestNotificationItemRoundTrip(t *testing.T) {
	x := SampleNotificationItem("notificationitem-1", when)
	got, err := ParseNotificationItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationItem() = %v", err)
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

func TestNotificationItemParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationItem(f); err == nil {
		t.Error("ParseNotificationItem() accepted a record with no id")
	}
	f = SampleNotificationItem("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseNotificationItem(f); err == nil {
		t.Error("ParseNotificationItem() accepted a non-numeric attempts")
	}
}

func TestNotificationItemCollections(t *testing.T) {
	xs := []NotificationItem{SampleNotificationItem("a", when), SampleNotificationItem("bb", when), SampleNotificationItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationItems(xs)); got != 2 {
		t.Errorf("ActiveNotificationItems() kept %d, want 2", got)
	}
	if got, want := SumNotificationItems(xs), 1+3; got != want {
		t.Errorf("SumNotificationItems() = %d, want %d", got, want)
	}
	SortNotificationItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationItems() put %q first", xs[0].ID)
	}
	if len(IndexNotificationItems(xs)) != 3 {
		t.Errorf("IndexNotificationItems() = %v", IndexNotificationItems(xs))
	}
	if err := ValidateNotificationItems(ActiveNotificationItems(xs)); err != nil {
		t.Errorf("ValidateNotificationItems() = %v", err)
	}
}

func TestNotificationItemNormalizeClamps(t *testing.T) {
	x := SampleNotificationItem("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationItem("r1", when)
	rev, err := r.PutNotificationItem(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationItem() = %v", err)
	}
	got, gotRev, err := r.GetNotificationItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationItems() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationItem(NotificationItem{}, 0); err == nil {
		t.Error("PutNotificationItem() accepted an empty entity")
	}
}

func TestNotificationDraftRoundTrip(t *testing.T) {
	x := SampleNotificationDraft("notificationdraft-1", when)
	got, err := ParseNotificationDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationDraft() = %v", err)
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

func TestNotificationDraftParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationDraft(f); err == nil {
		t.Error("ParseNotificationDraft() accepted a record with no id")
	}
	f = SampleNotificationDraft("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseNotificationDraft(f); err == nil {
		t.Error("ParseNotificationDraft() accepted a non-numeric retained")
	}
}

func TestNotificationDraftCollections(t *testing.T) {
	xs := []NotificationDraft{SampleNotificationDraft("a", when), SampleNotificationDraft("bb", when), SampleNotificationDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationDrafts(xs)); got != 2 {
		t.Errorf("ActiveNotificationDrafts() kept %d, want 2", got)
	}
	if got, want := SumNotificationDrafts(xs), 1+3; got != want {
		t.Errorf("SumNotificationDrafts() = %d, want %d", got, want)
	}
	SortNotificationDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationDrafts() put %q first", xs[0].ID)
	}
	if len(IndexNotificationDrafts(xs)) != 3 {
		t.Errorf("IndexNotificationDrafts() = %v", IndexNotificationDrafts(xs))
	}
	if err := ValidateNotificationDrafts(ActiveNotificationDrafts(xs)); err != nil {
		t.Errorf("ValidateNotificationDrafts() = %v", err)
	}
}

func TestNotificationDraftNormalizeClamps(t *testing.T) {
	x := SampleNotificationDraft("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationDraft("r1", when)
	rev, err := r.PutNotificationDraft(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationDraft() = %v", err)
	}
	got, gotRev, err := r.GetNotificationDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationDraft(NotificationDraft{}, 0); err == nil {
		t.Error("PutNotificationDraft() accepted an empty entity")
	}
}

func TestNotificationRevisionRoundTrip(t *testing.T) {
	x := SampleNotificationRevision("notificationrevision-1", when)
	got, err := ParseNotificationRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationRevision() = %v", err)
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

func TestNotificationRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationRevision(f); err == nil {
		t.Error("ParseNotificationRevision() accepted a record with no id")
	}
	f = SampleNotificationRevision("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseNotificationRevision(f); err == nil {
		t.Error("ParseNotificationRevision() accepted a non-numeric count")
	}
}

func TestNotificationRevisionCollections(t *testing.T) {
	xs := []NotificationRevision{SampleNotificationRevision("a", when), SampleNotificationRevision("bb", when), SampleNotificationRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationRevisions(xs)); got != 2 {
		t.Errorf("ActiveNotificationRevisions() kept %d, want 2", got)
	}
	if got, want := SumNotificationRevisions(xs), 1+3; got != want {
		t.Errorf("SumNotificationRevisions() = %d, want %d", got, want)
	}
	SortNotificationRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationRevisions() put %q first", xs[0].ID)
	}
	if len(IndexNotificationRevisions(xs)) != 3 {
		t.Errorf("IndexNotificationRevisions() = %v", IndexNotificationRevisions(xs))
	}
	if err := ValidateNotificationRevisions(ActiveNotificationRevisions(xs)); err != nil {
		t.Errorf("ValidateNotificationRevisions() = %v", err)
	}
}

func TestNotificationRevisionNormalizeClamps(t *testing.T) {
	x := SampleNotificationRevision("a", when)
	x.Region = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationRevision("r1", when)
	rev, err := r.PutNotificationRevision(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationRevision() = %v", err)
	}
	got, gotRev, err := r.GetNotificationRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationRevision(NotificationRevision{}, 0); err == nil {
		t.Error("PutNotificationRevision() accepted an empty entity")
	}
}

func TestNotificationSnapshotRoundTrip(t *testing.T) {
	x := SampleNotificationSnapshot("notificationsnapshot-1", when)
	got, err := ParseNotificationSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationSnapshot() = %v", err)
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

func TestNotificationSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationSnapshot(f); err == nil {
		t.Error("ParseNotificationSnapshot() accepted a record with no id")
	}
	f = SampleNotificationSnapshot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseNotificationSnapshot(f); err == nil {
		t.Error("ParseNotificationSnapshot() accepted a non-numeric offset")
	}
}

func TestNotificationSnapshotCollections(t *testing.T) {
	xs := []NotificationSnapshot{SampleNotificationSnapshot("a", when), SampleNotificationSnapshot("bb", when), SampleNotificationSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationSnapshots(xs)); got != 2 {
		t.Errorf("ActiveNotificationSnapshots() kept %d, want 2", got)
	}
	if got, want := SumNotificationSnapshots(xs), 1+3; got != want {
		t.Errorf("SumNotificationSnapshots() = %d, want %d", got, want)
	}
	SortNotificationSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexNotificationSnapshots(xs)) != 3 {
		t.Errorf("IndexNotificationSnapshots() = %v", IndexNotificationSnapshots(xs))
	}
	if err := ValidateNotificationSnapshots(ActiveNotificationSnapshots(xs)); err != nil {
		t.Errorf("ValidateNotificationSnapshots() = %v", err)
	}
}

func TestNotificationSnapshotNormalizeClamps(t *testing.T) {
	x := SampleNotificationSnapshot("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationSnapshot("r1", when)
	rev, err := r.PutNotificationSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetNotificationSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationSnapshot(NotificationSnapshot{}, 0); err == nil {
		t.Error("PutNotificationSnapshot() accepted an empty entity")
	}
}

func TestNotificationBatchRoundTrip(t *testing.T) {
	x := SampleNotificationBatch("notificationbatch-1", when)
	got, err := ParseNotificationBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationBatch() = %v", err)
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

func TestNotificationBatchParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationBatch(f); err == nil {
		t.Error("ParseNotificationBatch() accepted a record with no id")
	}
	f = SampleNotificationBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseNotificationBatch(f); err == nil {
		t.Error("ParseNotificationBatch() accepted a non-numeric version")
	}
}

func TestNotificationBatchCollections(t *testing.T) {
	xs := []NotificationBatch{SampleNotificationBatch("a", when), SampleNotificationBatch("bb", when), SampleNotificationBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationBatchs(xs)); got != 2 {
		t.Errorf("ActiveNotificationBatchs() kept %d, want 2", got)
	}
	if got, want := SumNotificationBatchs(xs), 1+3; got != want {
		t.Errorf("SumNotificationBatchs() = %d, want %d", got, want)
	}
	SortNotificationBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationBatchs() put %q first", xs[0].ID)
	}
	if len(IndexNotificationBatchs(xs)) != 3 {
		t.Errorf("IndexNotificationBatchs() = %v", IndexNotificationBatchs(xs))
	}
	if err := ValidateNotificationBatchs(ActiveNotificationBatchs(xs)); err != nil {
		t.Errorf("ValidateNotificationBatchs() = %v", err)
	}
}

func TestNotificationBatchNormalizeClamps(t *testing.T) {
	x := SampleNotificationBatch("a", when)
	x.Label = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationBatch("r1", when)
	rev, err := r.PutNotificationBatch(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationBatch() = %v", err)
	}
	got, gotRev, err := r.GetNotificationBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationBatch(NotificationBatch{}, 0); err == nil {
		t.Error("PutNotificationBatch() accepted an empty entity")
	}
}

func TestNotificationRequestRoundTrip(t *testing.T) {
	x := SampleNotificationRequest("notificationrequest-1", when)
	got, err := ParseNotificationRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationRequest() = %v", err)
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

func TestNotificationRequestParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationRequest(f); err == nil {
		t.Error("ParseNotificationRequest() accepted a record with no id")
	}
	f = SampleNotificationRequest("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseNotificationRequest(f); err == nil {
		t.Error("ParseNotificationRequest() accepted a non-numeric count")
	}
}

func TestNotificationRequestCollections(t *testing.T) {
	xs := []NotificationRequest{SampleNotificationRequest("a", when), SampleNotificationRequest("bb", when), SampleNotificationRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationRequests(xs)); got != 2 {
		t.Errorf("ActiveNotificationRequests() kept %d, want 2", got)
	}
	if got, want := SumNotificationRequests(xs), 1+3; got != want {
		t.Errorf("SumNotificationRequests() = %d, want %d", got, want)
	}
	SortNotificationRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationRequests() put %q first", xs[0].ID)
	}
	if len(IndexNotificationRequests(xs)) != 3 {
		t.Errorf("IndexNotificationRequests() = %v", IndexNotificationRequests(xs))
	}
	if err := ValidateNotificationRequests(ActiveNotificationRequests(xs)); err != nil {
		t.Errorf("ValidateNotificationRequests() = %v", err)
	}
}

func TestNotificationRequestNormalizeClamps(t *testing.T) {
	x := SampleNotificationRequest("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationRequest("r1", when)
	rev, err := r.PutNotificationRequest(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationRequest() = %v", err)
	}
	got, gotRev, err := r.GetNotificationRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationRequests() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationRequest(NotificationRequest{}, 0); err == nil {
		t.Error("PutNotificationRequest() accepted an empty entity")
	}
}

func TestNotificationSummaryRoundTrip(t *testing.T) {
	x := SampleNotificationSummary("notificationsummary-1", when)
	got, err := ParseNotificationSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationSummary() = %v", err)
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

func TestNotificationSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationSummary(f); err == nil {
		t.Error("ParseNotificationSummary() accepted a record with no id")
	}
	f = SampleNotificationSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseNotificationSummary(f); err == nil {
		t.Error("ParseNotificationSummary() accepted a non-numeric capacity")
	}
}

func TestNotificationSummaryCollections(t *testing.T) {
	xs := []NotificationSummary{SampleNotificationSummary("a", when), SampleNotificationSummary("bb", when), SampleNotificationSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationSummarys(xs)); got != 2 {
		t.Errorf("ActiveNotificationSummarys() kept %d, want 2", got)
	}
	if got, want := SumNotificationSummarys(xs), 1+3; got != want {
		t.Errorf("SumNotificationSummarys() = %d, want %d", got, want)
	}
	SortNotificationSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationSummarys() put %q first", xs[0].ID)
	}
	if len(IndexNotificationSummarys(xs)) != 3 {
		t.Errorf("IndexNotificationSummarys() = %v", IndexNotificationSummarys(xs))
	}
	if err := ValidateNotificationSummarys(ActiveNotificationSummarys(xs)); err != nil {
		t.Errorf("ValidateNotificationSummarys() = %v", err)
	}
}

func TestNotificationSummaryNormalizeClamps(t *testing.T) {
	x := SampleNotificationSummary("a", when)
	x.Comment = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationSummary("r1", when)
	rev, err := r.PutNotificationSummary(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationSummary() = %v", err)
	}
	got, gotRev, err := r.GetNotificationSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationSummary(NotificationSummary{}, 0); err == nil {
		t.Error("PutNotificationSummary() accepted an empty entity")
	}
}

func TestNotificationPolicyRoundTrip(t *testing.T) {
	x := SampleNotificationPolicy("notificationpolicy-1", when)
	got, err := ParseNotificationPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationPolicy() = %v", err)
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

func TestNotificationPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationPolicy(f); err == nil {
		t.Error("ParseNotificationPolicy() accepted a record with no id")
	}
	f = SampleNotificationPolicy("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseNotificationPolicy(f); err == nil {
		t.Error("ParseNotificationPolicy() accepted a non-numeric attempts")
	}
}

func TestNotificationPolicyCollections(t *testing.T) {
	xs := []NotificationPolicy{SampleNotificationPolicy("a", when), SampleNotificationPolicy("bb", when), SampleNotificationPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationPolicys(xs)); got != 2 {
		t.Errorf("ActiveNotificationPolicys() kept %d, want 2", got)
	}
	if got, want := SumNotificationPolicys(xs), 1+3; got != want {
		t.Errorf("SumNotificationPolicys() = %d, want %d", got, want)
	}
	SortNotificationPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationPolicys() put %q first", xs[0].ID)
	}
	if len(IndexNotificationPolicys(xs)) != 3 {
		t.Errorf("IndexNotificationPolicys() = %v", IndexNotificationPolicys(xs))
	}
	if err := ValidateNotificationPolicys(ActiveNotificationPolicys(xs)); err != nil {
		t.Errorf("ValidateNotificationPolicys() = %v", err)
	}
}

func TestNotificationPolicyNormalizeClamps(t *testing.T) {
	x := SampleNotificationPolicy("a", when)
	x.Group = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationPolicy("r1", when)
	rev, err := r.PutNotificationPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationPolicy() = %v", err)
	}
	got, gotRev, err := r.GetNotificationPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationPolicy(NotificationPolicy{}, 0); err == nil {
		t.Error("PutNotificationPolicy() accepted an empty entity")
	}
}

func TestNotificationLinkRoundTrip(t *testing.T) {
	x := SampleNotificationLink("notificationlink-1", when)
	got, err := ParseNotificationLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationLink() = %v", err)
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

func TestNotificationLinkParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationLink(f); err == nil {
		t.Error("ParseNotificationLink() accepted a record with no id")
	}
	f = SampleNotificationLink("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseNotificationLink(f); err == nil {
		t.Error("ParseNotificationLink() accepted a non-numeric retained")
	}
}

func TestNotificationLinkCollections(t *testing.T) {
	xs := []NotificationLink{SampleNotificationLink("a", when), SampleNotificationLink("bb", when), SampleNotificationLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationLinks(xs)); got != 2 {
		t.Errorf("ActiveNotificationLinks() kept %d, want 2", got)
	}
	if got, want := SumNotificationLinks(xs), 1+3; got != want {
		t.Errorf("SumNotificationLinks() = %d, want %d", got, want)
	}
	SortNotificationLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationLinks() put %q first", xs[0].ID)
	}
	if len(IndexNotificationLinks(xs)) != 3 {
		t.Errorf("IndexNotificationLinks() = %v", IndexNotificationLinks(xs))
	}
	if err := ValidateNotificationLinks(ActiveNotificationLinks(xs)); err != nil {
		t.Errorf("ValidateNotificationLinks() = %v", err)
	}
}

func TestNotificationLinkNormalizeClamps(t *testing.T) {
	x := SampleNotificationLink("a", when)
	x.Channel = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationLink("r1", when)
	rev, err := r.PutNotificationLink(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationLink() = %v", err)
	}
	got, gotRev, err := r.GetNotificationLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationLinks() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationLink(NotificationLink{}, 0); err == nil {
		t.Error("PutNotificationLink() accepted an empty entity")
	}
}

func TestNotificationNoteRoundTrip(t *testing.T) {
	x := SampleNotificationNote("notificationnote-1", when)
	got, err := ParseNotificationNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationNote() = %v", err)
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

func TestNotificationNoteParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationNote(f); err == nil {
		t.Error("ParseNotificationNote() accepted a record with no id")
	}
	f = SampleNotificationNote("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseNotificationNote(f); err == nil {
		t.Error("ParseNotificationNote() accepted a non-numeric offset")
	}
}

func TestNotificationNoteCollections(t *testing.T) {
	xs := []NotificationNote{SampleNotificationNote("a", when), SampleNotificationNote("bb", when), SampleNotificationNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationNotes(xs)); got != 2 {
		t.Errorf("ActiveNotificationNotes() kept %d, want 2", got)
	}
	if got, want := SumNotificationNotes(xs), 1+3; got != want {
		t.Errorf("SumNotificationNotes() = %d, want %d", got, want)
	}
	SortNotificationNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationNotes() put %q first", xs[0].ID)
	}
	if len(IndexNotificationNotes(xs)) != 3 {
		t.Errorf("IndexNotificationNotes() = %v", IndexNotificationNotes(xs))
	}
	if err := ValidateNotificationNotes(ActiveNotificationNotes(xs)); err != nil {
		t.Errorf("ValidateNotificationNotes() = %v", err)
	}
}

func TestNotificationNoteNormalizeClamps(t *testing.T) {
	x := SampleNotificationNote("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationNote("r1", when)
	rev, err := r.PutNotificationNote(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationNote() = %v", err)
	}
	got, gotRev, err := r.GetNotificationNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationNotes() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationNote(NotificationNote{}, 0); err == nil {
		t.Error("PutNotificationNote() accepted an empty entity")
	}
}

func TestNotificationWindowRoundTrip(t *testing.T) {
	x := SampleNotificationWindow("notificationwindow-1", when)
	got, err := ParseNotificationWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationWindow() = %v", err)
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

func TestNotificationWindowParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationWindow(f); err == nil {
		t.Error("ParseNotificationWindow() accepted a record with no id")
	}
	f = SampleNotificationWindow("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseNotificationWindow(f); err == nil {
		t.Error("ParseNotificationWindow() accepted a non-numeric capacity")
	}
}

func TestNotificationWindowCollections(t *testing.T) {
	xs := []NotificationWindow{SampleNotificationWindow("a", when), SampleNotificationWindow("bb", when), SampleNotificationWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationWindows(xs)); got != 2 {
		t.Errorf("ActiveNotificationWindows() kept %d, want 2", got)
	}
	if got, want := SumNotificationWindows(xs), 1+3; got != want {
		t.Errorf("SumNotificationWindows() = %d, want %d", got, want)
	}
	SortNotificationWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationWindows() put %q first", xs[0].ID)
	}
	if len(IndexNotificationWindows(xs)) != 3 {
		t.Errorf("IndexNotificationWindows() = %v", IndexNotificationWindows(xs))
	}
	if err := ValidateNotificationWindows(ActiveNotificationWindows(xs)); err != nil {
		t.Errorf("ValidateNotificationWindows() = %v", err)
	}
}

func TestNotificationWindowNormalizeClamps(t *testing.T) {
	x := SampleNotificationWindow("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationWindow("r1", when)
	rev, err := r.PutNotificationWindow(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationWindow() = %v", err)
	}
	got, gotRev, err := r.GetNotificationWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationWindows() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationWindow(NotificationWindow{}, 0); err == nil {
		t.Error("PutNotificationWindow() accepted an empty entity")
	}
}

func TestNotificationQuotaRoundTrip(t *testing.T) {
	x := SampleNotificationQuota("notificationquota-1", when)
	got, err := ParseNotificationQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationQuota() = %v", err)
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

func TestNotificationQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationQuota(f); err == nil {
		t.Error("ParseNotificationQuota() accepted a record with no id")
	}
	f = SampleNotificationQuota("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseNotificationQuota(f); err == nil {
		t.Error("ParseNotificationQuota() accepted a non-numeric version")
	}
}

func TestNotificationQuotaCollections(t *testing.T) {
	xs := []NotificationQuota{SampleNotificationQuota("a", when), SampleNotificationQuota("bb", when), SampleNotificationQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationQuotas(xs)); got != 2 {
		t.Errorf("ActiveNotificationQuotas() kept %d, want 2", got)
	}
	if got, want := SumNotificationQuotas(xs), 1+3; got != want {
		t.Errorf("SumNotificationQuotas() = %d, want %d", got, want)
	}
	SortNotificationQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationQuotas() put %q first", xs[0].ID)
	}
	if len(IndexNotificationQuotas(xs)) != 3 {
		t.Errorf("IndexNotificationQuotas() = %v", IndexNotificationQuotas(xs))
	}
	if err := ValidateNotificationQuotas(ActiveNotificationQuotas(xs)); err != nil {
		t.Errorf("ValidateNotificationQuotas() = %v", err)
	}
}

func TestNotificationQuotaNormalizeClamps(t *testing.T) {
	x := SampleNotificationQuota("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationQuota("r1", when)
	rev, err := r.PutNotificationQuota(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationQuota() = %v", err)
	}
	got, gotRev, err := r.GetNotificationQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationQuota(NotificationQuota{}, 0); err == nil {
		t.Error("PutNotificationQuota() accepted an empty entity")
	}
}

func TestNotificationBucketRoundTrip(t *testing.T) {
	x := SampleNotificationBucket("notificationbucket-1", when)
	got, err := ParseNotificationBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationBucket() = %v", err)
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

func TestNotificationBucketParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationBucket(f); err == nil {
		t.Error("ParseNotificationBucket() accepted a record with no id")
	}
	f = SampleNotificationBucket("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseNotificationBucket(f); err == nil {
		t.Error("ParseNotificationBucket() accepted a non-numeric version")
	}
}

func TestNotificationBucketCollections(t *testing.T) {
	xs := []NotificationBucket{SampleNotificationBucket("a", when), SampleNotificationBucket("bb", when), SampleNotificationBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationBuckets(xs)); got != 2 {
		t.Errorf("ActiveNotificationBuckets() kept %d, want 2", got)
	}
	if got, want := SumNotificationBuckets(xs), 1+3; got != want {
		t.Errorf("SumNotificationBuckets() = %d, want %d", got, want)
	}
	SortNotificationBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationBuckets() put %q first", xs[0].ID)
	}
	if len(IndexNotificationBuckets(xs)) != 3 {
		t.Errorf("IndexNotificationBuckets() = %v", IndexNotificationBuckets(xs))
	}
	if err := ValidateNotificationBuckets(ActiveNotificationBuckets(xs)); err != nil {
		t.Errorf("ValidateNotificationBuckets() = %v", err)
	}
}

func TestNotificationBucketNormalizeClamps(t *testing.T) {
	x := SampleNotificationBucket("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationBucket("r1", when)
	rev, err := r.PutNotificationBucket(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationBucket() = %v", err)
	}
	got, gotRev, err := r.GetNotificationBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationBucket(NotificationBucket{}, 0); err == nil {
		t.Error("PutNotificationBucket() accepted an empty entity")
	}
}

func TestNotificationLaneRoundTrip(t *testing.T) {
	x := SampleNotificationLane("notificationlane-1", when)
	got, err := ParseNotificationLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationLane() = %v", err)
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

func TestNotificationLaneParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationLane(f); err == nil {
		t.Error("ParseNotificationLane() accepted a record with no id")
	}
	f = SampleNotificationLane("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseNotificationLane(f); err == nil {
		t.Error("ParseNotificationLane() accepted a non-numeric version")
	}
}

func TestNotificationLaneCollections(t *testing.T) {
	xs := []NotificationLane{SampleNotificationLane("a", when), SampleNotificationLane("bb", when), SampleNotificationLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationLanes(xs)); got != 2 {
		t.Errorf("ActiveNotificationLanes() kept %d, want 2", got)
	}
	if got, want := SumNotificationLanes(xs), 1+3; got != want {
		t.Errorf("SumNotificationLanes() = %d, want %d", got, want)
	}
	SortNotificationLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationLanes() put %q first", xs[0].ID)
	}
	if len(IndexNotificationLanes(xs)) != 3 {
		t.Errorf("IndexNotificationLanes() = %v", IndexNotificationLanes(xs))
	}
	if err := ValidateNotificationLanes(ActiveNotificationLanes(xs)); err != nil {
		t.Errorf("ValidateNotificationLanes() = %v", err)
	}
}

func TestNotificationLaneNormalizeClamps(t *testing.T) {
	x := SampleNotificationLane("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationLane("r1", when)
	rev, err := r.PutNotificationLane(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationLane() = %v", err)
	}
	got, gotRev, err := r.GetNotificationLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationLanes() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationLane(NotificationLane{}, 0); err == nil {
		t.Error("PutNotificationLane() accepted an empty entity")
	}
}

func TestNotificationSlotRoundTrip(t *testing.T) {
	x := SampleNotificationSlot("notificationslot-1", when)
	got, err := ParseNotificationSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationSlot() = %v", err)
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

func TestNotificationSlotParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationSlot(f); err == nil {
		t.Error("ParseNotificationSlot() accepted a record with no id")
	}
	f = SampleNotificationSlot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseNotificationSlot(f); err == nil {
		t.Error("ParseNotificationSlot() accepted a non-numeric priority")
	}
}

func TestNotificationSlotCollections(t *testing.T) {
	xs := []NotificationSlot{SampleNotificationSlot("a", when), SampleNotificationSlot("bb", when), SampleNotificationSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationSlots(xs)); got != 2 {
		t.Errorf("ActiveNotificationSlots() kept %d, want 2", got)
	}
	if got, want := SumNotificationSlots(xs), 1+3; got != want {
		t.Errorf("SumNotificationSlots() = %d, want %d", got, want)
	}
	SortNotificationSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationSlots() put %q first", xs[0].ID)
	}
	if len(IndexNotificationSlots(xs)) != 3 {
		t.Errorf("IndexNotificationSlots() = %v", IndexNotificationSlots(xs))
	}
	if err := ValidateNotificationSlots(ActiveNotificationSlots(xs)); err != nil {
		t.Errorf("ValidateNotificationSlots() = %v", err)
	}
}

func TestNotificationSlotNormalizeClamps(t *testing.T) {
	x := SampleNotificationSlot("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationSlot("r1", when)
	rev, err := r.PutNotificationSlot(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationSlot() = %v", err)
	}
	got, gotRev, err := r.GetNotificationSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationSlots() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationSlot(NotificationSlot{}, 0); err == nil {
		t.Error("PutNotificationSlot() accepted an empty entity")
	}
}

func TestNotificationTagRoundTrip(t *testing.T) {
	x := SampleNotificationTag("notificationtag-1", when)
	got, err := ParseNotificationTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationTag() = %v", err)
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

func TestNotificationTagParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationTag(f); err == nil {
		t.Error("ParseNotificationTag() accepted a record with no id")
	}
	f = SampleNotificationTag("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseNotificationTag(f); err == nil {
		t.Error("ParseNotificationTag() accepted a non-numeric offset")
	}
}

func TestNotificationTagCollections(t *testing.T) {
	xs := []NotificationTag{SampleNotificationTag("a", when), SampleNotificationTag("bb", when), SampleNotificationTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationTags(xs)); got != 2 {
		t.Errorf("ActiveNotificationTags() kept %d, want 2", got)
	}
	if got, want := SumNotificationTags(xs), 1+3; got != want {
		t.Errorf("SumNotificationTags() = %d, want %d", got, want)
	}
	SortNotificationTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationTags() put %q first", xs[0].ID)
	}
	if len(IndexNotificationTags(xs)) != 3 {
		t.Errorf("IndexNotificationTags() = %v", IndexNotificationTags(xs))
	}
	if err := ValidateNotificationTags(ActiveNotificationTags(xs)); err != nil {
		t.Errorf("ValidateNotificationTags() = %v", err)
	}
}

func TestNotificationTagNormalizeClamps(t *testing.T) {
	x := SampleNotificationTag("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationTag("r1", when)
	rev, err := r.PutNotificationTag(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationTag() = %v", err)
	}
	got, gotRev, err := r.GetNotificationTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationTags() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationTag(NotificationTag{}, 0); err == nil {
		t.Error("PutNotificationTag() accepted an empty entity")
	}
}

func TestNotificationRuleRoundTrip(t *testing.T) {
	x := SampleNotificationRule("notificationrule-1", when)
	got, err := ParseNotificationRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationRule() = %v", err)
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

func TestNotificationRuleParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationRule(f); err == nil {
		t.Error("ParseNotificationRule() accepted a record with no id")
	}
	f = SampleNotificationRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseNotificationRule(f); err == nil {
		t.Error("ParseNotificationRule() accepted a non-numeric offset")
	}
}

func TestNotificationRuleCollections(t *testing.T) {
	xs := []NotificationRule{SampleNotificationRule("a", when), SampleNotificationRule("bb", when), SampleNotificationRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationRules(xs)); got != 2 {
		t.Errorf("ActiveNotificationRules() kept %d, want 2", got)
	}
	if got, want := SumNotificationRules(xs), 1+3; got != want {
		t.Errorf("SumNotificationRules() = %d, want %d", got, want)
	}
	SortNotificationRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationRules() put %q first", xs[0].ID)
	}
	if len(IndexNotificationRules(xs)) != 3 {
		t.Errorf("IndexNotificationRules() = %v", IndexNotificationRules(xs))
	}
	if err := ValidateNotificationRules(ActiveNotificationRules(xs)); err != nil {
		t.Errorf("ValidateNotificationRules() = %v", err)
	}
}

func TestNotificationRuleNormalizeClamps(t *testing.T) {
	x := SampleNotificationRule("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationRule("r1", when)
	rev, err := r.PutNotificationRule(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationRule() = %v", err)
	}
	got, gotRev, err := r.GetNotificationRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationRules() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationRule(NotificationRule{}, 0); err == nil {
		t.Error("PutNotificationRule() accepted an empty entity")
	}
}

func TestNotificationHoldRoundTrip(t *testing.T) {
	x := SampleNotificationHold("notificationhold-1", when)
	got, err := ParseNotificationHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationHold() = %v", err)
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

func TestNotificationHoldParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationHold(f); err == nil {
		t.Error("ParseNotificationHold() accepted a record with no id")
	}
	f = SampleNotificationHold("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseNotificationHold(f); err == nil {
		t.Error("ParseNotificationHold() accepted a non-numeric offset")
	}
}

func TestNotificationHoldCollections(t *testing.T) {
	xs := []NotificationHold{SampleNotificationHold("a", when), SampleNotificationHold("bb", when), SampleNotificationHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationHolds(xs)); got != 2 {
		t.Errorf("ActiveNotificationHolds() kept %d, want 2", got)
	}
	if got, want := SumNotificationHolds(xs), 1+3; got != want {
		t.Errorf("SumNotificationHolds() = %d, want %d", got, want)
	}
	SortNotificationHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationHolds() put %q first", xs[0].ID)
	}
	if len(IndexNotificationHolds(xs)) != 3 {
		t.Errorf("IndexNotificationHolds() = %v", IndexNotificationHolds(xs))
	}
	if err := ValidateNotificationHolds(ActiveNotificationHolds(xs)); err != nil {
		t.Errorf("ValidateNotificationHolds() = %v", err)
	}
}

func TestNotificationHoldNormalizeClamps(t *testing.T) {
	x := SampleNotificationHold("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationHold("r1", when)
	rev, err := r.PutNotificationHold(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationHold() = %v", err)
	}
	got, gotRev, err := r.GetNotificationHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationHolds() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationHold(NotificationHold{}, 0); err == nil {
		t.Error("PutNotificationHold() accepted an empty entity")
	}
}

func TestNotificationClaimRoundTrip(t *testing.T) {
	x := SampleNotificationClaim("notificationclaim-1", when)
	got, err := ParseNotificationClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationClaim() = %v", err)
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

func TestNotificationClaimParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationClaim(f); err == nil {
		t.Error("ParseNotificationClaim() accepted a record with no id")
	}
	f = SampleNotificationClaim("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseNotificationClaim(f); err == nil {
		t.Error("ParseNotificationClaim() accepted a non-numeric capacity")
	}
}

func TestNotificationClaimCollections(t *testing.T) {
	xs := []NotificationClaim{SampleNotificationClaim("a", when), SampleNotificationClaim("bb", when), SampleNotificationClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationClaims(xs)); got != 2 {
		t.Errorf("ActiveNotificationClaims() kept %d, want 2", got)
	}
	if got, want := SumNotificationClaims(xs), 1+3; got != want {
		t.Errorf("SumNotificationClaims() = %d, want %d", got, want)
	}
	SortNotificationClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationClaims() put %q first", xs[0].ID)
	}
	if len(IndexNotificationClaims(xs)) != 3 {
		t.Errorf("IndexNotificationClaims() = %v", IndexNotificationClaims(xs))
	}
	if err := ValidateNotificationClaims(ActiveNotificationClaims(xs)); err != nil {
		t.Errorf("ValidateNotificationClaims() = %v", err)
	}
}

func TestNotificationClaimNormalizeClamps(t *testing.T) {
	x := SampleNotificationClaim("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationClaim("r1", when)
	rev, err := r.PutNotificationClaim(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationClaim() = %v", err)
	}
	got, gotRev, err := r.GetNotificationClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationClaims() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationClaim(NotificationClaim{}, 0); err == nil {
		t.Error("PutNotificationClaim() accepted an empty entity")
	}
}

func TestNotificationCursorRoundTrip(t *testing.T) {
	x := SampleNotificationCursor("notificationcursor-1", when)
	got, err := ParseNotificationCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseNotificationCursor() = %v", err)
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

func TestNotificationCursorParseRejectsBadInput(t *testing.T) {
	f := SampleNotificationCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseNotificationCursor(f); err == nil {
		t.Error("ParseNotificationCursor() accepted a record with no id")
	}
	f = SampleNotificationCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseNotificationCursor(f); err == nil {
		t.Error("ParseNotificationCursor() accepted a non-numeric version")
	}
}

func TestNotificationCursorCollections(t *testing.T) {
	xs := []NotificationCursor{SampleNotificationCursor("a", when), SampleNotificationCursor("bb", when), SampleNotificationCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveNotificationCursors(xs)); got != 2 {
		t.Errorf("ActiveNotificationCursors() kept %d, want 2", got)
	}
	if got, want := SumNotificationCursors(xs), 1+3; got != want {
		t.Errorf("SumNotificationCursors() = %d, want %d", got, want)
	}
	SortNotificationCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortNotificationCursors() put %q first", xs[0].ID)
	}
	if len(IndexNotificationCursors(xs)) != 3 {
		t.Errorf("IndexNotificationCursors() = %v", IndexNotificationCursors(xs))
	}
	if err := ValidateNotificationCursors(ActiveNotificationCursors(xs)); err != nil {
		t.Errorf("ValidateNotificationCursors() = %v", err)
	}
}

func TestNotificationCursorNormalizeClamps(t *testing.T) {
	x := SampleNotificationCursor("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestNotificationCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleNotificationCursor("r1", when)
	rev, err := r.PutNotificationCursor(x, 0)
	if err != nil {
		t.Fatalf("PutNotificationCursor() = %v", err)
	}
	got, gotRev, err := r.GetNotificationCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetNotificationCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListNotificationCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListNotificationCursors() = %v, %v", all, err)
	}
	if _, err := r.PutNotificationCursor(NotificationCursor{}, 0); err == nil {
		t.Error("PutNotificationCursor() accepted an empty entity")
	}
}
