package session

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
		if !strings.HasPrefix(Prefixes[i], "session/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleSessionRecord("ok", when).Fields()); err != nil {
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

func TestSessionRecordRoundTrip(t *testing.T) {
	x := SampleSessionRecord("sessionrecord-1", when)
	got, err := ParseSessionRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionRecord() = %v", err)
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

func TestSessionRecordParseRejectsBadInput(t *testing.T) {
	f := SampleSessionRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionRecord(f); err == nil {
		t.Error("ParseSessionRecord() accepted a record with no id")
	}
	f = SampleSessionRecord("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSessionRecord(f); err == nil {
		t.Error("ParseSessionRecord() accepted a non-numeric priority")
	}
}

func TestSessionRecordCollections(t *testing.T) {
	xs := []SessionRecord{SampleSessionRecord("a", when), SampleSessionRecord("bb", when), SampleSessionRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionRecords(xs)); got != 2 {
		t.Errorf("ActiveSessionRecords() kept %d, want 2", got)
	}
	if got, want := SumSessionRecords(xs), 1+3; got != want {
		t.Errorf("SumSessionRecords() = %d, want %d", got, want)
	}
	SortSessionRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionRecords() put %q first", xs[0].ID)
	}
	if len(IndexSessionRecords(xs)) != 3 {
		t.Errorf("IndexSessionRecords() = %v", IndexSessionRecords(xs))
	}
	if err := ValidateSessionRecords(ActiveSessionRecords(xs)); err != nil {
		t.Errorf("ValidateSessionRecords() = %v", err)
	}
}

func TestSessionRecordNormalizeClamps(t *testing.T) {
	x := SampleSessionRecord("a", when)
	x.Source = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionRecord("r1", when)
	rev, err := r.PutSessionRecord(x, 0)
	if err != nil {
		t.Fatalf("PutSessionRecord() = %v", err)
	}
	got, gotRev, err := r.GetSessionRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionRecords() = %v, %v", all, err)
	}
	if _, err := r.PutSessionRecord(SessionRecord{}, 0); err == nil {
		t.Error("PutSessionRecord() accepted an empty entity")
	}
}

func TestSessionEntryRoundTrip(t *testing.T) {
	x := SampleSessionEntry("sessionentry-1", when)
	got, err := ParseSessionEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionEntry() = %v", err)
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

func TestSessionEntryParseRejectsBadInput(t *testing.T) {
	f := SampleSessionEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionEntry(f); err == nil {
		t.Error("ParseSessionEntry() accepted a record with no id")
	}
	f = SampleSessionEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSessionEntry(f); err == nil {
		t.Error("ParseSessionEntry() accepted a non-numeric offset")
	}
}

func TestSessionEntryCollections(t *testing.T) {
	xs := []SessionEntry{SampleSessionEntry("a", when), SampleSessionEntry("bb", when), SampleSessionEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionEntrys(xs)); got != 2 {
		t.Errorf("ActiveSessionEntrys() kept %d, want 2", got)
	}
	if got, want := SumSessionEntrys(xs), 1+3; got != want {
		t.Errorf("SumSessionEntrys() = %d, want %d", got, want)
	}
	SortSessionEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionEntrys() put %q first", xs[0].ID)
	}
	if len(IndexSessionEntrys(xs)) != 3 {
		t.Errorf("IndexSessionEntrys() = %v", IndexSessionEntrys(xs))
	}
	if err := ValidateSessionEntrys(ActiveSessionEntrys(xs)); err != nil {
		t.Errorf("ValidateSessionEntrys() = %v", err)
	}
}

func TestSessionEntryNormalizeClamps(t *testing.T) {
	x := SampleSessionEntry("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionEntry("r1", when)
	rev, err := r.PutSessionEntry(x, 0)
	if err != nil {
		t.Fatalf("PutSessionEntry() = %v", err)
	}
	got, gotRev, err := r.GetSessionEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutSessionEntry(SessionEntry{}, 0); err == nil {
		t.Error("PutSessionEntry() accepted an empty entity")
	}
}

func TestSessionItemRoundTrip(t *testing.T) {
	x := SampleSessionItem("sessionitem-1", when)
	got, err := ParseSessionItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionItem() = %v", err)
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

func TestSessionItemParseRejectsBadInput(t *testing.T) {
	f := SampleSessionItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionItem(f); err == nil {
		t.Error("ParseSessionItem() accepted a record with no id")
	}
	f = SampleSessionItem("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSessionItem(f); err == nil {
		t.Error("ParseSessionItem() accepted a non-numeric count")
	}
}

func TestSessionItemCollections(t *testing.T) {
	xs := []SessionItem{SampleSessionItem("a", when), SampleSessionItem("bb", when), SampleSessionItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionItems(xs)); got != 2 {
		t.Errorf("ActiveSessionItems() kept %d, want 2", got)
	}
	if got, want := SumSessionItems(xs), 1+3; got != want {
		t.Errorf("SumSessionItems() = %d, want %d", got, want)
	}
	SortSessionItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionItems() put %q first", xs[0].ID)
	}
	if len(IndexSessionItems(xs)) != 3 {
		t.Errorf("IndexSessionItems() = %v", IndexSessionItems(xs))
	}
	if err := ValidateSessionItems(ActiveSessionItems(xs)); err != nil {
		t.Errorf("ValidateSessionItems() = %v", err)
	}
}

func TestSessionItemNormalizeClamps(t *testing.T) {
	x := SampleSessionItem("a", when)
	x.Name = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionItem("r1", when)
	rev, err := r.PutSessionItem(x, 0)
	if err != nil {
		t.Fatalf("PutSessionItem() = %v", err)
	}
	got, gotRev, err := r.GetSessionItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionItems() = %v, %v", all, err)
	}
	if _, err := r.PutSessionItem(SessionItem{}, 0); err == nil {
		t.Error("PutSessionItem() accepted an empty entity")
	}
}

func TestSessionDraftRoundTrip(t *testing.T) {
	x := SampleSessionDraft("sessiondraft-1", when)
	got, err := ParseSessionDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionDraft() = %v", err)
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

func TestSessionDraftParseRejectsBadInput(t *testing.T) {
	f := SampleSessionDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionDraft(f); err == nil {
		t.Error("ParseSessionDraft() accepted a record with no id")
	}
	f = SampleSessionDraft("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSessionDraft(f); err == nil {
		t.Error("ParseSessionDraft() accepted a non-numeric count")
	}
}

func TestSessionDraftCollections(t *testing.T) {
	xs := []SessionDraft{SampleSessionDraft("a", when), SampleSessionDraft("bb", when), SampleSessionDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionDrafts(xs)); got != 2 {
		t.Errorf("ActiveSessionDrafts() kept %d, want 2", got)
	}
	if got, want := SumSessionDrafts(xs), 1+3; got != want {
		t.Errorf("SumSessionDrafts() = %d, want %d", got, want)
	}
	SortSessionDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionDrafts() put %q first", xs[0].ID)
	}
	if len(IndexSessionDrafts(xs)) != 3 {
		t.Errorf("IndexSessionDrafts() = %v", IndexSessionDrafts(xs))
	}
	if err := ValidateSessionDrafts(ActiveSessionDrafts(xs)); err != nil {
		t.Errorf("ValidateSessionDrafts() = %v", err)
	}
}

func TestSessionDraftNormalizeClamps(t *testing.T) {
	x := SampleSessionDraft("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionDraft("r1", when)
	rev, err := r.PutSessionDraft(x, 0)
	if err != nil {
		t.Fatalf("PutSessionDraft() = %v", err)
	}
	got, gotRev, err := r.GetSessionDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutSessionDraft(SessionDraft{}, 0); err == nil {
		t.Error("PutSessionDraft() accepted an empty entity")
	}
}

func TestSessionRevisionRoundTrip(t *testing.T) {
	x := SampleSessionRevision("sessionrevision-1", when)
	got, err := ParseSessionRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionRevision() = %v", err)
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

func TestSessionRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleSessionRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionRevision(f); err == nil {
		t.Error("ParseSessionRevision() accepted a record with no id")
	}
	f = SampleSessionRevision("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSessionRevision(f); err == nil {
		t.Error("ParseSessionRevision() accepted a non-numeric priority")
	}
}

func TestSessionRevisionCollections(t *testing.T) {
	xs := []SessionRevision{SampleSessionRevision("a", when), SampleSessionRevision("bb", when), SampleSessionRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionRevisions(xs)); got != 2 {
		t.Errorf("ActiveSessionRevisions() kept %d, want 2", got)
	}
	if got, want := SumSessionRevisions(xs), 1+3; got != want {
		t.Errorf("SumSessionRevisions() = %d, want %d", got, want)
	}
	SortSessionRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionRevisions() put %q first", xs[0].ID)
	}
	if len(IndexSessionRevisions(xs)) != 3 {
		t.Errorf("IndexSessionRevisions() = %v", IndexSessionRevisions(xs))
	}
	if err := ValidateSessionRevisions(ActiveSessionRevisions(xs)); err != nil {
		t.Errorf("ValidateSessionRevisions() = %v", err)
	}
}

func TestSessionRevisionNormalizeClamps(t *testing.T) {
	x := SampleSessionRevision("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionRevision("r1", when)
	rev, err := r.PutSessionRevision(x, 0)
	if err != nil {
		t.Fatalf("PutSessionRevision() = %v", err)
	}
	got, gotRev, err := r.GetSessionRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutSessionRevision(SessionRevision{}, 0); err == nil {
		t.Error("PutSessionRevision() accepted an empty entity")
	}
}

func TestSessionSnapshotRoundTrip(t *testing.T) {
	x := SampleSessionSnapshot("sessionsnapshot-1", when)
	got, err := ParseSessionSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionSnapshot() = %v", err)
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

func TestSessionSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleSessionSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionSnapshot(f); err == nil {
		t.Error("ParseSessionSnapshot() accepted a record with no id")
	}
	f = SampleSessionSnapshot("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseSessionSnapshot(f); err == nil {
		t.Error("ParseSessionSnapshot() accepted a non-numeric priority")
	}
}

func TestSessionSnapshotCollections(t *testing.T) {
	xs := []SessionSnapshot{SampleSessionSnapshot("a", when), SampleSessionSnapshot("bb", when), SampleSessionSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionSnapshots(xs)); got != 2 {
		t.Errorf("ActiveSessionSnapshots() kept %d, want 2", got)
	}
	if got, want := SumSessionSnapshots(xs), 1+3; got != want {
		t.Errorf("SumSessionSnapshots() = %d, want %d", got, want)
	}
	SortSessionSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexSessionSnapshots(xs)) != 3 {
		t.Errorf("IndexSessionSnapshots() = %v", IndexSessionSnapshots(xs))
	}
	if err := ValidateSessionSnapshots(ActiveSessionSnapshots(xs)); err != nil {
		t.Errorf("ValidateSessionSnapshots() = %v", err)
	}
}

func TestSessionSnapshotNormalizeClamps(t *testing.T) {
	x := SampleSessionSnapshot("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionSnapshot("r1", when)
	rev, err := r.PutSessionSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutSessionSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetSessionSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutSessionSnapshot(SessionSnapshot{}, 0); err == nil {
		t.Error("PutSessionSnapshot() accepted an empty entity")
	}
}

func TestSessionBatchRoundTrip(t *testing.T) {
	x := SampleSessionBatch("sessionbatch-1", when)
	got, err := ParseSessionBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionBatch() = %v", err)
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

func TestSessionBatchParseRejectsBadInput(t *testing.T) {
	f := SampleSessionBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionBatch(f); err == nil {
		t.Error("ParseSessionBatch() accepted a record with no id")
	}
	f = SampleSessionBatch("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSessionBatch(f); err == nil {
		t.Error("ParseSessionBatch() accepted a non-numeric capacity")
	}
}

func TestSessionBatchCollections(t *testing.T) {
	xs := []SessionBatch{SampleSessionBatch("a", when), SampleSessionBatch("bb", when), SampleSessionBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionBatchs(xs)); got != 2 {
		t.Errorf("ActiveSessionBatchs() kept %d, want 2", got)
	}
	if got, want := SumSessionBatchs(xs), 1+3; got != want {
		t.Errorf("SumSessionBatchs() = %d, want %d", got, want)
	}
	SortSessionBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionBatchs() put %q first", xs[0].ID)
	}
	if len(IndexSessionBatchs(xs)) != 3 {
		t.Errorf("IndexSessionBatchs() = %v", IndexSessionBatchs(xs))
	}
	if err := ValidateSessionBatchs(ActiveSessionBatchs(xs)); err != nil {
		t.Errorf("ValidateSessionBatchs() = %v", err)
	}
}

func TestSessionBatchNormalizeClamps(t *testing.T) {
	x := SampleSessionBatch("a", when)
	x.Reference = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionBatch("r1", when)
	rev, err := r.PutSessionBatch(x, 0)
	if err != nil {
		t.Fatalf("PutSessionBatch() = %v", err)
	}
	got, gotRev, err := r.GetSessionBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutSessionBatch(SessionBatch{}, 0); err == nil {
		t.Error("PutSessionBatch() accepted an empty entity")
	}
}

func TestSessionRequestRoundTrip(t *testing.T) {
	x := SampleSessionRequest("sessionrequest-1", when)
	got, err := ParseSessionRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionRequest() = %v", err)
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

func TestSessionRequestParseRejectsBadInput(t *testing.T) {
	f := SampleSessionRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionRequest(f); err == nil {
		t.Error("ParseSessionRequest() accepted a record with no id")
	}
	f = SampleSessionRequest("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSessionRequest(f); err == nil {
		t.Error("ParseSessionRequest() accepted a non-numeric version")
	}
}

func TestSessionRequestCollections(t *testing.T) {
	xs := []SessionRequest{SampleSessionRequest("a", when), SampleSessionRequest("bb", when), SampleSessionRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionRequests(xs)); got != 2 {
		t.Errorf("ActiveSessionRequests() kept %d, want 2", got)
	}
	if got, want := SumSessionRequests(xs), 1+3; got != want {
		t.Errorf("SumSessionRequests() = %d, want %d", got, want)
	}
	SortSessionRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionRequests() put %q first", xs[0].ID)
	}
	if len(IndexSessionRequests(xs)) != 3 {
		t.Errorf("IndexSessionRequests() = %v", IndexSessionRequests(xs))
	}
	if err := ValidateSessionRequests(ActiveSessionRequests(xs)); err != nil {
		t.Errorf("ValidateSessionRequests() = %v", err)
	}
}

func TestSessionRequestNormalizeClamps(t *testing.T) {
	x := SampleSessionRequest("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionRequest("r1", when)
	rev, err := r.PutSessionRequest(x, 0)
	if err != nil {
		t.Fatalf("PutSessionRequest() = %v", err)
	}
	got, gotRev, err := r.GetSessionRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionRequests() = %v, %v", all, err)
	}
	if _, err := r.PutSessionRequest(SessionRequest{}, 0); err == nil {
		t.Error("PutSessionRequest() accepted an empty entity")
	}
}

func TestSessionSummaryRoundTrip(t *testing.T) {
	x := SampleSessionSummary("sessionsummary-1", when)
	got, err := ParseSessionSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionSummary() = %v", err)
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

func TestSessionSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleSessionSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionSummary(f); err == nil {
		t.Error("ParseSessionSummary() accepted a record with no id")
	}
	f = SampleSessionSummary("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSessionSummary(f); err == nil {
		t.Error("ParseSessionSummary() accepted a non-numeric capacity")
	}
}

func TestSessionSummaryCollections(t *testing.T) {
	xs := []SessionSummary{SampleSessionSummary("a", when), SampleSessionSummary("bb", when), SampleSessionSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionSummarys(xs)); got != 2 {
		t.Errorf("ActiveSessionSummarys() kept %d, want 2", got)
	}
	if got, want := SumSessionSummarys(xs), 1+3; got != want {
		t.Errorf("SumSessionSummarys() = %d, want %d", got, want)
	}
	SortSessionSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionSummarys() put %q first", xs[0].ID)
	}
	if len(IndexSessionSummarys(xs)) != 3 {
		t.Errorf("IndexSessionSummarys() = %v", IndexSessionSummarys(xs))
	}
	if err := ValidateSessionSummarys(ActiveSessionSummarys(xs)); err != nil {
		t.Errorf("ValidateSessionSummarys() = %v", err)
	}
}

func TestSessionSummaryNormalizeClamps(t *testing.T) {
	x := SampleSessionSummary("a", when)
	x.Owner = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionSummary("r1", when)
	rev, err := r.PutSessionSummary(x, 0)
	if err != nil {
		t.Fatalf("PutSessionSummary() = %v", err)
	}
	got, gotRev, err := r.GetSessionSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutSessionSummary(SessionSummary{}, 0); err == nil {
		t.Error("PutSessionSummary() accepted an empty entity")
	}
}

func TestSessionPolicyRoundTrip(t *testing.T) {
	x := SampleSessionPolicy("sessionpolicy-1", when)
	got, err := ParseSessionPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionPolicy() = %v", err)
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

func TestSessionPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleSessionPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionPolicy(f); err == nil {
		t.Error("ParseSessionPolicy() accepted a record with no id")
	}
	f = SampleSessionPolicy("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseSessionPolicy(f); err == nil {
		t.Error("ParseSessionPolicy() accepted a non-numeric weight")
	}
}

func TestSessionPolicyCollections(t *testing.T) {
	xs := []SessionPolicy{SampleSessionPolicy("a", when), SampleSessionPolicy("bb", when), SampleSessionPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionPolicys(xs)); got != 2 {
		t.Errorf("ActiveSessionPolicys() kept %d, want 2", got)
	}
	if got, want := SumSessionPolicys(xs), 1+3; got != want {
		t.Errorf("SumSessionPolicys() = %d, want %d", got, want)
	}
	SortSessionPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionPolicys() put %q first", xs[0].ID)
	}
	if len(IndexSessionPolicys(xs)) != 3 {
		t.Errorf("IndexSessionPolicys() = %v", IndexSessionPolicys(xs))
	}
	if err := ValidateSessionPolicys(ActiveSessionPolicys(xs)); err != nil {
		t.Errorf("ValidateSessionPolicys() = %v", err)
	}
}

func TestSessionPolicyNormalizeClamps(t *testing.T) {
	x := SampleSessionPolicy("a", when)
	x.Channel = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionPolicy("r1", when)
	rev, err := r.PutSessionPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutSessionPolicy() = %v", err)
	}
	got, gotRev, err := r.GetSessionPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutSessionPolicy(SessionPolicy{}, 0); err == nil {
		t.Error("PutSessionPolicy() accepted an empty entity")
	}
}

func TestSessionLinkRoundTrip(t *testing.T) {
	x := SampleSessionLink("sessionlink-1", when)
	got, err := ParseSessionLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionLink() = %v", err)
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

func TestSessionLinkParseRejectsBadInput(t *testing.T) {
	f := SampleSessionLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionLink(f); err == nil {
		t.Error("ParseSessionLink() accepted a record with no id")
	}
	f = SampleSessionLink("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseSessionLink(f); err == nil {
		t.Error("ParseSessionLink() accepted a non-numeric attempts")
	}
}

func TestSessionLinkCollections(t *testing.T) {
	xs := []SessionLink{SampleSessionLink("a", when), SampleSessionLink("bb", when), SampleSessionLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionLinks(xs)); got != 2 {
		t.Errorf("ActiveSessionLinks() kept %d, want 2", got)
	}
	if got, want := SumSessionLinks(xs), 1+3; got != want {
		t.Errorf("SumSessionLinks() = %d, want %d", got, want)
	}
	SortSessionLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionLinks() put %q first", xs[0].ID)
	}
	if len(IndexSessionLinks(xs)) != 3 {
		t.Errorf("IndexSessionLinks() = %v", IndexSessionLinks(xs))
	}
	if err := ValidateSessionLinks(ActiveSessionLinks(xs)); err != nil {
		t.Errorf("ValidateSessionLinks() = %v", err)
	}
}

func TestSessionLinkNormalizeClamps(t *testing.T) {
	x := SampleSessionLink("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionLink("r1", when)
	rev, err := r.PutSessionLink(x, 0)
	if err != nil {
		t.Fatalf("PutSessionLink() = %v", err)
	}
	got, gotRev, err := r.GetSessionLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionLinks() = %v, %v", all, err)
	}
	if _, err := r.PutSessionLink(SessionLink{}, 0); err == nil {
		t.Error("PutSessionLink() accepted an empty entity")
	}
}

func TestSessionNoteRoundTrip(t *testing.T) {
	x := SampleSessionNote("sessionnote-1", when)
	got, err := ParseSessionNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionNote() = %v", err)
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

func TestSessionNoteParseRejectsBadInput(t *testing.T) {
	f := SampleSessionNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionNote(f); err == nil {
		t.Error("ParseSessionNote() accepted a record with no id")
	}
	f = SampleSessionNote("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSessionNote(f); err == nil {
		t.Error("ParseSessionNote() accepted a non-numeric capacity")
	}
}

func TestSessionNoteCollections(t *testing.T) {
	xs := []SessionNote{SampleSessionNote("a", when), SampleSessionNote("bb", when), SampleSessionNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionNotes(xs)); got != 2 {
		t.Errorf("ActiveSessionNotes() kept %d, want 2", got)
	}
	if got, want := SumSessionNotes(xs), 1+3; got != want {
		t.Errorf("SumSessionNotes() = %d, want %d", got, want)
	}
	SortSessionNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionNotes() put %q first", xs[0].ID)
	}
	if len(IndexSessionNotes(xs)) != 3 {
		t.Errorf("IndexSessionNotes() = %v", IndexSessionNotes(xs))
	}
	if err := ValidateSessionNotes(ActiveSessionNotes(xs)); err != nil {
		t.Errorf("ValidateSessionNotes() = %v", err)
	}
}

func TestSessionNoteNormalizeClamps(t *testing.T) {
	x := SampleSessionNote("a", when)
	x.Name = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionNote("r1", when)
	rev, err := r.PutSessionNote(x, 0)
	if err != nil {
		t.Fatalf("PutSessionNote() = %v", err)
	}
	got, gotRev, err := r.GetSessionNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionNotes() = %v, %v", all, err)
	}
	if _, err := r.PutSessionNote(SessionNote{}, 0); err == nil {
		t.Error("PutSessionNote() accepted an empty entity")
	}
}

func TestSessionWindowRoundTrip(t *testing.T) {
	x := SampleSessionWindow("sessionwindow-1", when)
	got, err := ParseSessionWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionWindow() = %v", err)
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

func TestSessionWindowParseRejectsBadInput(t *testing.T) {
	f := SampleSessionWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionWindow(f); err == nil {
		t.Error("ParseSessionWindow() accepted a record with no id")
	}
	f = SampleSessionWindow("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSessionWindow(f); err == nil {
		t.Error("ParseSessionWindow() accepted a non-numeric offset")
	}
}

func TestSessionWindowCollections(t *testing.T) {
	xs := []SessionWindow{SampleSessionWindow("a", when), SampleSessionWindow("bb", when), SampleSessionWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionWindows(xs)); got != 2 {
		t.Errorf("ActiveSessionWindows() kept %d, want 2", got)
	}
	if got, want := SumSessionWindows(xs), 1+3; got != want {
		t.Errorf("SumSessionWindows() = %d, want %d", got, want)
	}
	SortSessionWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionWindows() put %q first", xs[0].ID)
	}
	if len(IndexSessionWindows(xs)) != 3 {
		t.Errorf("IndexSessionWindows() = %v", IndexSessionWindows(xs))
	}
	if err := ValidateSessionWindows(ActiveSessionWindows(xs)); err != nil {
		t.Errorf("ValidateSessionWindows() = %v", err)
	}
}

func TestSessionWindowNormalizeClamps(t *testing.T) {
	x := SampleSessionWindow("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionWindow("r1", when)
	rev, err := r.PutSessionWindow(x, 0)
	if err != nil {
		t.Fatalf("PutSessionWindow() = %v", err)
	}
	got, gotRev, err := r.GetSessionWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionWindows() = %v, %v", all, err)
	}
	if _, err := r.PutSessionWindow(SessionWindow{}, 0); err == nil {
		t.Error("PutSessionWindow() accepted an empty entity")
	}
}

func TestSessionQuotaRoundTrip(t *testing.T) {
	x := SampleSessionQuota("sessionquota-1", when)
	got, err := ParseSessionQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionQuota() = %v", err)
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

func TestSessionQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleSessionQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionQuota(f); err == nil {
		t.Error("ParseSessionQuota() accepted a record with no id")
	}
	f = SampleSessionQuota("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseSessionQuota(f); err == nil {
		t.Error("ParseSessionQuota() accepted a non-numeric capacity")
	}
}

func TestSessionQuotaCollections(t *testing.T) {
	xs := []SessionQuota{SampleSessionQuota("a", when), SampleSessionQuota("bb", when), SampleSessionQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionQuotas(xs)); got != 2 {
		t.Errorf("ActiveSessionQuotas() kept %d, want 2", got)
	}
	if got, want := SumSessionQuotas(xs), 1+3; got != want {
		t.Errorf("SumSessionQuotas() = %d, want %d", got, want)
	}
	SortSessionQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionQuotas() put %q first", xs[0].ID)
	}
	if len(IndexSessionQuotas(xs)) != 3 {
		t.Errorf("IndexSessionQuotas() = %v", IndexSessionQuotas(xs))
	}
	if err := ValidateSessionQuotas(ActiveSessionQuotas(xs)); err != nil {
		t.Errorf("ValidateSessionQuotas() = %v", err)
	}
}

func TestSessionQuotaNormalizeClamps(t *testing.T) {
	x := SampleSessionQuota("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionQuota("r1", when)
	rev, err := r.PutSessionQuota(x, 0)
	if err != nil {
		t.Fatalf("PutSessionQuota() = %v", err)
	}
	got, gotRev, err := r.GetSessionQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutSessionQuota(SessionQuota{}, 0); err == nil {
		t.Error("PutSessionQuota() accepted an empty entity")
	}
}

func TestSessionBucketRoundTrip(t *testing.T) {
	x := SampleSessionBucket("sessionbucket-1", when)
	got, err := ParseSessionBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionBucket() = %v", err)
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

func TestSessionBucketParseRejectsBadInput(t *testing.T) {
	f := SampleSessionBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionBucket(f); err == nil {
		t.Error("ParseSessionBucket() accepted a record with no id")
	}
	f = SampleSessionBucket("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseSessionBucket(f); err == nil {
		t.Error("ParseSessionBucket() accepted a non-numeric retained")
	}
}

func TestSessionBucketCollections(t *testing.T) {
	xs := []SessionBucket{SampleSessionBucket("a", when), SampleSessionBucket("bb", when), SampleSessionBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionBuckets(xs)); got != 2 {
		t.Errorf("ActiveSessionBuckets() kept %d, want 2", got)
	}
	if got, want := SumSessionBuckets(xs), 1+3; got != want {
		t.Errorf("SumSessionBuckets() = %d, want %d", got, want)
	}
	SortSessionBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionBuckets() put %q first", xs[0].ID)
	}
	if len(IndexSessionBuckets(xs)) != 3 {
		t.Errorf("IndexSessionBuckets() = %v", IndexSessionBuckets(xs))
	}
	if err := ValidateSessionBuckets(ActiveSessionBuckets(xs)); err != nil {
		t.Errorf("ValidateSessionBuckets() = %v", err)
	}
}

func TestSessionBucketNormalizeClamps(t *testing.T) {
	x := SampleSessionBucket("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionBucket("r1", when)
	rev, err := r.PutSessionBucket(x, 0)
	if err != nil {
		t.Fatalf("PutSessionBucket() = %v", err)
	}
	got, gotRev, err := r.GetSessionBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutSessionBucket(SessionBucket{}, 0); err == nil {
		t.Error("PutSessionBucket() accepted an empty entity")
	}
}

func TestSessionLaneRoundTrip(t *testing.T) {
	x := SampleSessionLane("sessionlane-1", when)
	got, err := ParseSessionLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionLane() = %v", err)
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

func TestSessionLaneParseRejectsBadInput(t *testing.T) {
	f := SampleSessionLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionLane(f); err == nil {
		t.Error("ParseSessionLane() accepted a record with no id")
	}
	f = SampleSessionLane("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSessionLane(f); err == nil {
		t.Error("ParseSessionLane() accepted a non-numeric offset")
	}
}

func TestSessionLaneCollections(t *testing.T) {
	xs := []SessionLane{SampleSessionLane("a", when), SampleSessionLane("bb", when), SampleSessionLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionLanes(xs)); got != 2 {
		t.Errorf("ActiveSessionLanes() kept %d, want 2", got)
	}
	if got, want := SumSessionLanes(xs), 1+3; got != want {
		t.Errorf("SumSessionLanes() = %d, want %d", got, want)
	}
	SortSessionLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionLanes() put %q first", xs[0].ID)
	}
	if len(IndexSessionLanes(xs)) != 3 {
		t.Errorf("IndexSessionLanes() = %v", IndexSessionLanes(xs))
	}
	if err := ValidateSessionLanes(ActiveSessionLanes(xs)); err != nil {
		t.Errorf("ValidateSessionLanes() = %v", err)
	}
}

func TestSessionLaneNormalizeClamps(t *testing.T) {
	x := SampleSessionLane("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionLane("r1", when)
	rev, err := r.PutSessionLane(x, 0)
	if err != nil {
		t.Fatalf("PutSessionLane() = %v", err)
	}
	got, gotRev, err := r.GetSessionLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionLanes() = %v, %v", all, err)
	}
	if _, err := r.PutSessionLane(SessionLane{}, 0); err == nil {
		t.Error("PutSessionLane() accepted an empty entity")
	}
}

func TestSessionSlotRoundTrip(t *testing.T) {
	x := SampleSessionSlot("sessionslot-1", when)
	got, err := ParseSessionSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionSlot() = %v", err)
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

func TestSessionSlotParseRejectsBadInput(t *testing.T) {
	f := SampleSessionSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionSlot(f); err == nil {
		t.Error("ParseSessionSlot() accepted a record with no id")
	}
	f = SampleSessionSlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseSessionSlot(f); err == nil {
		t.Error("ParseSessionSlot() accepted a non-numeric count")
	}
}

func TestSessionSlotCollections(t *testing.T) {
	xs := []SessionSlot{SampleSessionSlot("a", when), SampleSessionSlot("bb", when), SampleSessionSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionSlots(xs)); got != 2 {
		t.Errorf("ActiveSessionSlots() kept %d, want 2", got)
	}
	if got, want := SumSessionSlots(xs), 1+3; got != want {
		t.Errorf("SumSessionSlots() = %d, want %d", got, want)
	}
	SortSessionSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionSlots() put %q first", xs[0].ID)
	}
	if len(IndexSessionSlots(xs)) != 3 {
		t.Errorf("IndexSessionSlots() = %v", IndexSessionSlots(xs))
	}
	if err := ValidateSessionSlots(ActiveSessionSlots(xs)); err != nil {
		t.Errorf("ValidateSessionSlots() = %v", err)
	}
}

func TestSessionSlotNormalizeClamps(t *testing.T) {
	x := SampleSessionSlot("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionSlot("r1", when)
	rev, err := r.PutSessionSlot(x, 0)
	if err != nil {
		t.Fatalf("PutSessionSlot() = %v", err)
	}
	got, gotRev, err := r.GetSessionSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionSlots() = %v, %v", all, err)
	}
	if _, err := r.PutSessionSlot(SessionSlot{}, 0); err == nil {
		t.Error("PutSessionSlot() accepted an empty entity")
	}
}

func TestSessionTagRoundTrip(t *testing.T) {
	x := SampleSessionTag("sessiontag-1", when)
	got, err := ParseSessionTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionTag() = %v", err)
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

func TestSessionTagParseRejectsBadInput(t *testing.T) {
	f := SampleSessionTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionTag(f); err == nil {
		t.Error("ParseSessionTag() accepted a record with no id")
	}
	f = SampleSessionTag("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseSessionTag(f); err == nil {
		t.Error("ParseSessionTag() accepted a non-numeric version")
	}
}

func TestSessionTagCollections(t *testing.T) {
	xs := []SessionTag{SampleSessionTag("a", when), SampleSessionTag("bb", when), SampleSessionTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionTags(xs)); got != 2 {
		t.Errorf("ActiveSessionTags() kept %d, want 2", got)
	}
	if got, want := SumSessionTags(xs), 1+3; got != want {
		t.Errorf("SumSessionTags() = %d, want %d", got, want)
	}
	SortSessionTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionTags() put %q first", xs[0].ID)
	}
	if len(IndexSessionTags(xs)) != 3 {
		t.Errorf("IndexSessionTags() = %v", IndexSessionTags(xs))
	}
	if err := ValidateSessionTags(ActiveSessionTags(xs)); err != nil {
		t.Errorf("ValidateSessionTags() = %v", err)
	}
}

func TestSessionTagNormalizeClamps(t *testing.T) {
	x := SampleSessionTag("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionTag("r1", when)
	rev, err := r.PutSessionTag(x, 0)
	if err != nil {
		t.Fatalf("PutSessionTag() = %v", err)
	}
	got, gotRev, err := r.GetSessionTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionTags() = %v, %v", all, err)
	}
	if _, err := r.PutSessionTag(SessionTag{}, 0); err == nil {
		t.Error("PutSessionTag() accepted an empty entity")
	}
}

func TestSessionRuleRoundTrip(t *testing.T) {
	x := SampleSessionRule("sessionrule-1", when)
	got, err := ParseSessionRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionRule() = %v", err)
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

func TestSessionRuleParseRejectsBadInput(t *testing.T) {
	f := SampleSessionRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionRule(f); err == nil {
		t.Error("ParseSessionRule() accepted a record with no id")
	}
	f = SampleSessionRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSessionRule(f); err == nil {
		t.Error("ParseSessionRule() accepted a non-numeric offset")
	}
}

func TestSessionRuleCollections(t *testing.T) {
	xs := []SessionRule{SampleSessionRule("a", when), SampleSessionRule("bb", when), SampleSessionRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionRules(xs)); got != 2 {
		t.Errorf("ActiveSessionRules() kept %d, want 2", got)
	}
	if got, want := SumSessionRules(xs), 1+3; got != want {
		t.Errorf("SumSessionRules() = %d, want %d", got, want)
	}
	SortSessionRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionRules() put %q first", xs[0].ID)
	}
	if len(IndexSessionRules(xs)) != 3 {
		t.Errorf("IndexSessionRules() = %v", IndexSessionRules(xs))
	}
	if err := ValidateSessionRules(ActiveSessionRules(xs)); err != nil {
		t.Errorf("ValidateSessionRules() = %v", err)
	}
}

func TestSessionRuleNormalizeClamps(t *testing.T) {
	x := SampleSessionRule("a", when)
	x.Region = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionRule("r1", when)
	rev, err := r.PutSessionRule(x, 0)
	if err != nil {
		t.Fatalf("PutSessionRule() = %v", err)
	}
	got, gotRev, err := r.GetSessionRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionRules() = %v, %v", all, err)
	}
	if _, err := r.PutSessionRule(SessionRule{}, 0); err == nil {
		t.Error("PutSessionRule() accepted an empty entity")
	}
}

func TestSessionHoldRoundTrip(t *testing.T) {
	x := SampleSessionHold("sessionhold-1", when)
	got, err := ParseSessionHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionHold() = %v", err)
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

func TestSessionHoldParseRejectsBadInput(t *testing.T) {
	f := SampleSessionHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionHold(f); err == nil {
		t.Error("ParseSessionHold() accepted a record with no id")
	}
	f = SampleSessionHold("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSessionHold(f); err == nil {
		t.Error("ParseSessionHold() accepted a non-numeric offset")
	}
}

func TestSessionHoldCollections(t *testing.T) {
	xs := []SessionHold{SampleSessionHold("a", when), SampleSessionHold("bb", when), SampleSessionHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionHolds(xs)); got != 2 {
		t.Errorf("ActiveSessionHolds() kept %d, want 2", got)
	}
	if got, want := SumSessionHolds(xs), 1+3; got != want {
		t.Errorf("SumSessionHolds() = %d, want %d", got, want)
	}
	SortSessionHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionHolds() put %q first", xs[0].ID)
	}
	if len(IndexSessionHolds(xs)) != 3 {
		t.Errorf("IndexSessionHolds() = %v", IndexSessionHolds(xs))
	}
	if err := ValidateSessionHolds(ActiveSessionHolds(xs)); err != nil {
		t.Errorf("ValidateSessionHolds() = %v", err)
	}
}

func TestSessionHoldNormalizeClamps(t *testing.T) {
	x := SampleSessionHold("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionHold("r1", when)
	rev, err := r.PutSessionHold(x, 0)
	if err != nil {
		t.Fatalf("PutSessionHold() = %v", err)
	}
	got, gotRev, err := r.GetSessionHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionHolds() = %v, %v", all, err)
	}
	if _, err := r.PutSessionHold(SessionHold{}, 0); err == nil {
		t.Error("PutSessionHold() accepted an empty entity")
	}
}

func TestSessionClaimRoundTrip(t *testing.T) {
	x := SampleSessionClaim("sessionclaim-1", when)
	got, err := ParseSessionClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionClaim() = %v", err)
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

func TestSessionClaimParseRejectsBadInput(t *testing.T) {
	f := SampleSessionClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionClaim(f); err == nil {
		t.Error("ParseSessionClaim() accepted a record with no id")
	}
	f = SampleSessionClaim("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSessionClaim(f); err == nil {
		t.Error("ParseSessionClaim() accepted a non-numeric offset")
	}
}

func TestSessionClaimCollections(t *testing.T) {
	xs := []SessionClaim{SampleSessionClaim("a", when), SampleSessionClaim("bb", when), SampleSessionClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionClaims(xs)); got != 2 {
		t.Errorf("ActiveSessionClaims() kept %d, want 2", got)
	}
	if got, want := SumSessionClaims(xs), 1+3; got != want {
		t.Errorf("SumSessionClaims() = %d, want %d", got, want)
	}
	SortSessionClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionClaims() put %q first", xs[0].ID)
	}
	if len(IndexSessionClaims(xs)) != 3 {
		t.Errorf("IndexSessionClaims() = %v", IndexSessionClaims(xs))
	}
	if err := ValidateSessionClaims(ActiveSessionClaims(xs)); err != nil {
		t.Errorf("ValidateSessionClaims() = %v", err)
	}
}

func TestSessionClaimNormalizeClamps(t *testing.T) {
	x := SampleSessionClaim("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionClaim("r1", when)
	rev, err := r.PutSessionClaim(x, 0)
	if err != nil {
		t.Fatalf("PutSessionClaim() = %v", err)
	}
	got, gotRev, err := r.GetSessionClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionClaims() = %v, %v", all, err)
	}
	if _, err := r.PutSessionClaim(SessionClaim{}, 0); err == nil {
		t.Error("PutSessionClaim() accepted an empty entity")
	}
}

func TestSessionCursorRoundTrip(t *testing.T) {
	x := SampleSessionCursor("sessioncursor-1", when)
	got, err := ParseSessionCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseSessionCursor() = %v", err)
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

func TestSessionCursorParseRejectsBadInput(t *testing.T) {
	f := SampleSessionCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseSessionCursor(f); err == nil {
		t.Error("ParseSessionCursor() accepted a record with no id")
	}
	f = SampleSessionCursor("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseSessionCursor(f); err == nil {
		t.Error("ParseSessionCursor() accepted a non-numeric offset")
	}
}

func TestSessionCursorCollections(t *testing.T) {
	xs := []SessionCursor{SampleSessionCursor("a", when), SampleSessionCursor("bb", when), SampleSessionCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveSessionCursors(xs)); got != 2 {
		t.Errorf("ActiveSessionCursors() kept %d, want 2", got)
	}
	if got, want := SumSessionCursors(xs), 1+3; got != want {
		t.Errorf("SumSessionCursors() = %d, want %d", got, want)
	}
	SortSessionCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortSessionCursors() put %q first", xs[0].ID)
	}
	if len(IndexSessionCursors(xs)) != 3 {
		t.Errorf("IndexSessionCursors() = %v", IndexSessionCursors(xs))
	}
	if err := ValidateSessionCursors(ActiveSessionCursors(xs)); err != nil {
		t.Errorf("ValidateSessionCursors() = %v", err)
	}
}

func TestSessionCursorNormalizeClamps(t *testing.T) {
	x := SampleSessionCursor("a", when)
	x.Comment = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestSessionCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleSessionCursor("r1", when)
	rev, err := r.PutSessionCursor(x, 0)
	if err != nil {
		t.Fatalf("PutSessionCursor() = %v", err)
	}
	got, gotRev, err := r.GetSessionCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetSessionCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListSessionCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListSessionCursors() = %v, %v", all, err)
	}
	if _, err := r.PutSessionCursor(SessionCursor{}, 0); err == nil {
		t.Error("PutSessionCursor() accepted an empty entity")
	}
}
