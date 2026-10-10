package audit

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
		if !strings.HasPrefix(Prefixes[i], "audit/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleAuditRecord("ok", when).Fields()); err != nil {
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

func TestAuditRecordRoundTrip(t *testing.T) {
	x := SampleAuditRecord("auditrecord-1", when)
	got, err := ParseAuditRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditRecord() = %v", err)
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

func TestAuditRecordParseRejectsBadInput(t *testing.T) {
	f := SampleAuditRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditRecord(f); err == nil {
		t.Error("ParseAuditRecord() accepted a record with no id")
	}
	f = SampleAuditRecord("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseAuditRecord(f); err == nil {
		t.Error("ParseAuditRecord() accepted a non-numeric weight")
	}
}

func TestAuditRecordCollections(t *testing.T) {
	xs := []AuditRecord{SampleAuditRecord("a", when), SampleAuditRecord("bb", when), SampleAuditRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditRecords(xs)); got != 2 {
		t.Errorf("ActiveAuditRecords() kept %d, want 2", got)
	}
	if got, want := SumAuditRecords(xs), 1+3; got != want {
		t.Errorf("SumAuditRecords() = %d, want %d", got, want)
	}
	SortAuditRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditRecords() put %q first", xs[0].ID)
	}
	if len(IndexAuditRecords(xs)) != 3 {
		t.Errorf("IndexAuditRecords() = %v", IndexAuditRecords(xs))
	}
	if err := ValidateAuditRecords(ActiveAuditRecords(xs)); err != nil {
		t.Errorf("ValidateAuditRecords() = %v", err)
	}
}

func TestAuditRecordNormalizeClamps(t *testing.T) {
	x := SampleAuditRecord("a", when)
	x.Owner = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditRecord("r1", when)
	rev, err := r.PutAuditRecord(x, 0)
	if err != nil {
		t.Fatalf("PutAuditRecord() = %v", err)
	}
	got, gotRev, err := r.GetAuditRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditRecords() = %v, %v", all, err)
	}
	if _, err := r.PutAuditRecord(AuditRecord{}, 0); err == nil {
		t.Error("PutAuditRecord() accepted an empty entity")
	}
}

func TestAuditEntryRoundTrip(t *testing.T) {
	x := SampleAuditEntry("auditentry-1", when)
	got, err := ParseAuditEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditEntry() = %v", err)
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

func TestAuditEntryParseRejectsBadInput(t *testing.T) {
	f := SampleAuditEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditEntry(f); err == nil {
		t.Error("ParseAuditEntry() accepted a record with no id")
	}
	f = SampleAuditEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseAuditEntry(f); err == nil {
		t.Error("ParseAuditEntry() accepted a non-numeric offset")
	}
}

func TestAuditEntryCollections(t *testing.T) {
	xs := []AuditEntry{SampleAuditEntry("a", when), SampleAuditEntry("bb", when), SampleAuditEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditEntrys(xs)); got != 2 {
		t.Errorf("ActiveAuditEntrys() kept %d, want 2", got)
	}
	if got, want := SumAuditEntrys(xs), 1+3; got != want {
		t.Errorf("SumAuditEntrys() = %d, want %d", got, want)
	}
	SortAuditEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditEntrys() put %q first", xs[0].ID)
	}
	if len(IndexAuditEntrys(xs)) != 3 {
		t.Errorf("IndexAuditEntrys() = %v", IndexAuditEntrys(xs))
	}
	if err := ValidateAuditEntrys(ActiveAuditEntrys(xs)); err != nil {
		t.Errorf("ValidateAuditEntrys() = %v", err)
	}
}

func TestAuditEntryNormalizeClamps(t *testing.T) {
	x := SampleAuditEntry("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditEntry("r1", when)
	rev, err := r.PutAuditEntry(x, 0)
	if err != nil {
		t.Fatalf("PutAuditEntry() = %v", err)
	}
	got, gotRev, err := r.GetAuditEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutAuditEntry(AuditEntry{}, 0); err == nil {
		t.Error("PutAuditEntry() accepted an empty entity")
	}
}

func TestAuditItemRoundTrip(t *testing.T) {
	x := SampleAuditItem("audititem-1", when)
	got, err := ParseAuditItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditItem() = %v", err)
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

func TestAuditItemParseRejectsBadInput(t *testing.T) {
	f := SampleAuditItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditItem(f); err == nil {
		t.Error("ParseAuditItem() accepted a record with no id")
	}
	f = SampleAuditItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseAuditItem(f); err == nil {
		t.Error("ParseAuditItem() accepted a non-numeric version")
	}
}

func TestAuditItemCollections(t *testing.T) {
	xs := []AuditItem{SampleAuditItem("a", when), SampleAuditItem("bb", when), SampleAuditItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditItems(xs)); got != 2 {
		t.Errorf("ActiveAuditItems() kept %d, want 2", got)
	}
	if got, want := SumAuditItems(xs), 1+3; got != want {
		t.Errorf("SumAuditItems() = %d, want %d", got, want)
	}
	SortAuditItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditItems() put %q first", xs[0].ID)
	}
	if len(IndexAuditItems(xs)) != 3 {
		t.Errorf("IndexAuditItems() = %v", IndexAuditItems(xs))
	}
	if err := ValidateAuditItems(ActiveAuditItems(xs)); err != nil {
		t.Errorf("ValidateAuditItems() = %v", err)
	}
}

func TestAuditItemNormalizeClamps(t *testing.T) {
	x := SampleAuditItem("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditItem("r1", when)
	rev, err := r.PutAuditItem(x, 0)
	if err != nil {
		t.Fatalf("PutAuditItem() = %v", err)
	}
	got, gotRev, err := r.GetAuditItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditItems() = %v, %v", all, err)
	}
	if _, err := r.PutAuditItem(AuditItem{}, 0); err == nil {
		t.Error("PutAuditItem() accepted an empty entity")
	}
}

func TestAuditDraftRoundTrip(t *testing.T) {
	x := SampleAuditDraft("auditdraft-1", when)
	got, err := ParseAuditDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditDraft() = %v", err)
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

func TestAuditDraftParseRejectsBadInput(t *testing.T) {
	f := SampleAuditDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditDraft(f); err == nil {
		t.Error("ParseAuditDraft() accepted a record with no id")
	}
	f = SampleAuditDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseAuditDraft(f); err == nil {
		t.Error("ParseAuditDraft() accepted a non-numeric attempts")
	}
}

func TestAuditDraftCollections(t *testing.T) {
	xs := []AuditDraft{SampleAuditDraft("a", when), SampleAuditDraft("bb", when), SampleAuditDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditDrafts(xs)); got != 2 {
		t.Errorf("ActiveAuditDrafts() kept %d, want 2", got)
	}
	if got, want := SumAuditDrafts(xs), 1+3; got != want {
		t.Errorf("SumAuditDrafts() = %d, want %d", got, want)
	}
	SortAuditDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditDrafts() put %q first", xs[0].ID)
	}
	if len(IndexAuditDrafts(xs)) != 3 {
		t.Errorf("IndexAuditDrafts() = %v", IndexAuditDrafts(xs))
	}
	if err := ValidateAuditDrafts(ActiveAuditDrafts(xs)); err != nil {
		t.Errorf("ValidateAuditDrafts() = %v", err)
	}
}

func TestAuditDraftNormalizeClamps(t *testing.T) {
	x := SampleAuditDraft("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditDraft("r1", when)
	rev, err := r.PutAuditDraft(x, 0)
	if err != nil {
		t.Fatalf("PutAuditDraft() = %v", err)
	}
	got, gotRev, err := r.GetAuditDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutAuditDraft(AuditDraft{}, 0); err == nil {
		t.Error("PutAuditDraft() accepted an empty entity")
	}
}

func TestAuditRevisionRoundTrip(t *testing.T) {
	x := SampleAuditRevision("auditrevision-1", when)
	got, err := ParseAuditRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditRevision() = %v", err)
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

func TestAuditRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleAuditRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditRevision(f); err == nil {
		t.Error("ParseAuditRevision() accepted a record with no id")
	}
	f = SampleAuditRevision("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseAuditRevision(f); err == nil {
		t.Error("ParseAuditRevision() accepted a non-numeric attempts")
	}
}

func TestAuditRevisionCollections(t *testing.T) {
	xs := []AuditRevision{SampleAuditRevision("a", when), SampleAuditRevision("bb", when), SampleAuditRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditRevisions(xs)); got != 2 {
		t.Errorf("ActiveAuditRevisions() kept %d, want 2", got)
	}
	if got, want := SumAuditRevisions(xs), 1+3; got != want {
		t.Errorf("SumAuditRevisions() = %d, want %d", got, want)
	}
	SortAuditRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditRevisions() put %q first", xs[0].ID)
	}
	if len(IndexAuditRevisions(xs)) != 3 {
		t.Errorf("IndexAuditRevisions() = %v", IndexAuditRevisions(xs))
	}
	if err := ValidateAuditRevisions(ActiveAuditRevisions(xs)); err != nil {
		t.Errorf("ValidateAuditRevisions() = %v", err)
	}
}

func TestAuditRevisionNormalizeClamps(t *testing.T) {
	x := SampleAuditRevision("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditRevision("r1", when)
	rev, err := r.PutAuditRevision(x, 0)
	if err != nil {
		t.Fatalf("PutAuditRevision() = %v", err)
	}
	got, gotRev, err := r.GetAuditRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutAuditRevision(AuditRevision{}, 0); err == nil {
		t.Error("PutAuditRevision() accepted an empty entity")
	}
}

func TestAuditSnapshotRoundTrip(t *testing.T) {
	x := SampleAuditSnapshot("auditsnapshot-1", when)
	got, err := ParseAuditSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditSnapshot() = %v", err)
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

func TestAuditSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleAuditSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditSnapshot(f); err == nil {
		t.Error("ParseAuditSnapshot() accepted a record with no id")
	}
	f = SampleAuditSnapshot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseAuditSnapshot(f); err == nil {
		t.Error("ParseAuditSnapshot() accepted a non-numeric weight")
	}
}

func TestAuditSnapshotCollections(t *testing.T) {
	xs := []AuditSnapshot{SampleAuditSnapshot("a", when), SampleAuditSnapshot("bb", when), SampleAuditSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditSnapshots(xs)); got != 2 {
		t.Errorf("ActiveAuditSnapshots() kept %d, want 2", got)
	}
	if got, want := SumAuditSnapshots(xs), 1+3; got != want {
		t.Errorf("SumAuditSnapshots() = %d, want %d", got, want)
	}
	SortAuditSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexAuditSnapshots(xs)) != 3 {
		t.Errorf("IndexAuditSnapshots() = %v", IndexAuditSnapshots(xs))
	}
	if err := ValidateAuditSnapshots(ActiveAuditSnapshots(xs)); err != nil {
		t.Errorf("ValidateAuditSnapshots() = %v", err)
	}
}

func TestAuditSnapshotNormalizeClamps(t *testing.T) {
	x := SampleAuditSnapshot("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditSnapshot("r1", when)
	rev, err := r.PutAuditSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutAuditSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetAuditSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutAuditSnapshot(AuditSnapshot{}, 0); err == nil {
		t.Error("PutAuditSnapshot() accepted an empty entity")
	}
}

func TestAuditBatchRoundTrip(t *testing.T) {
	x := SampleAuditBatch("auditbatch-1", when)
	got, err := ParseAuditBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditBatch() = %v", err)
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

func TestAuditBatchParseRejectsBadInput(t *testing.T) {
	f := SampleAuditBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditBatch(f); err == nil {
		t.Error("ParseAuditBatch() accepted a record with no id")
	}
	f = SampleAuditBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseAuditBatch(f); err == nil {
		t.Error("ParseAuditBatch() accepted a non-numeric version")
	}
}

func TestAuditBatchCollections(t *testing.T) {
	xs := []AuditBatch{SampleAuditBatch("a", when), SampleAuditBatch("bb", when), SampleAuditBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditBatchs(xs)); got != 2 {
		t.Errorf("ActiveAuditBatchs() kept %d, want 2", got)
	}
	if got, want := SumAuditBatchs(xs), 1+3; got != want {
		t.Errorf("SumAuditBatchs() = %d, want %d", got, want)
	}
	SortAuditBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditBatchs() put %q first", xs[0].ID)
	}
	if len(IndexAuditBatchs(xs)) != 3 {
		t.Errorf("IndexAuditBatchs() = %v", IndexAuditBatchs(xs))
	}
	if err := ValidateAuditBatchs(ActiveAuditBatchs(xs)); err != nil {
		t.Errorf("ValidateAuditBatchs() = %v", err)
	}
}

func TestAuditBatchNormalizeClamps(t *testing.T) {
	x := SampleAuditBatch("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditBatch("r1", when)
	rev, err := r.PutAuditBatch(x, 0)
	if err != nil {
		t.Fatalf("PutAuditBatch() = %v", err)
	}
	got, gotRev, err := r.GetAuditBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutAuditBatch(AuditBatch{}, 0); err == nil {
		t.Error("PutAuditBatch() accepted an empty entity")
	}
}

func TestAuditRequestRoundTrip(t *testing.T) {
	x := SampleAuditRequest("auditrequest-1", when)
	got, err := ParseAuditRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditRequest() = %v", err)
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

func TestAuditRequestParseRejectsBadInput(t *testing.T) {
	f := SampleAuditRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditRequest(f); err == nil {
		t.Error("ParseAuditRequest() accepted a record with no id")
	}
	f = SampleAuditRequest("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseAuditRequest(f); err == nil {
		t.Error("ParseAuditRequest() accepted a non-numeric version")
	}
}

func TestAuditRequestCollections(t *testing.T) {
	xs := []AuditRequest{SampleAuditRequest("a", when), SampleAuditRequest("bb", when), SampleAuditRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditRequests(xs)); got != 2 {
		t.Errorf("ActiveAuditRequests() kept %d, want 2", got)
	}
	if got, want := SumAuditRequests(xs), 1+3; got != want {
		t.Errorf("SumAuditRequests() = %d, want %d", got, want)
	}
	SortAuditRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditRequests() put %q first", xs[0].ID)
	}
	if len(IndexAuditRequests(xs)) != 3 {
		t.Errorf("IndexAuditRequests() = %v", IndexAuditRequests(xs))
	}
	if err := ValidateAuditRequests(ActiveAuditRequests(xs)); err != nil {
		t.Errorf("ValidateAuditRequests() = %v", err)
	}
}

func TestAuditRequestNormalizeClamps(t *testing.T) {
	x := SampleAuditRequest("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditRequest("r1", when)
	rev, err := r.PutAuditRequest(x, 0)
	if err != nil {
		t.Fatalf("PutAuditRequest() = %v", err)
	}
	got, gotRev, err := r.GetAuditRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditRequests() = %v, %v", all, err)
	}
	if _, err := r.PutAuditRequest(AuditRequest{}, 0); err == nil {
		t.Error("PutAuditRequest() accepted an empty entity")
	}
}

func TestAuditSummaryRoundTrip(t *testing.T) {
	x := SampleAuditSummary("auditsummary-1", when)
	got, err := ParseAuditSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditSummary() = %v", err)
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

func TestAuditSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleAuditSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditSummary(f); err == nil {
		t.Error("ParseAuditSummary() accepted a record with no id")
	}
	f = SampleAuditSummary("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseAuditSummary(f); err == nil {
		t.Error("ParseAuditSummary() accepted a non-numeric priority")
	}
}

func TestAuditSummaryCollections(t *testing.T) {
	xs := []AuditSummary{SampleAuditSummary("a", when), SampleAuditSummary("bb", when), SampleAuditSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditSummarys(xs)); got != 2 {
		t.Errorf("ActiveAuditSummarys() kept %d, want 2", got)
	}
	if got, want := SumAuditSummarys(xs), 1+3; got != want {
		t.Errorf("SumAuditSummarys() = %d, want %d", got, want)
	}
	SortAuditSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditSummarys() put %q first", xs[0].ID)
	}
	if len(IndexAuditSummarys(xs)) != 3 {
		t.Errorf("IndexAuditSummarys() = %v", IndexAuditSummarys(xs))
	}
	if err := ValidateAuditSummarys(ActiveAuditSummarys(xs)); err != nil {
		t.Errorf("ValidateAuditSummarys() = %v", err)
	}
}

func TestAuditSummaryNormalizeClamps(t *testing.T) {
	x := SampleAuditSummary("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditSummary("r1", when)
	rev, err := r.PutAuditSummary(x, 0)
	if err != nil {
		t.Fatalf("PutAuditSummary() = %v", err)
	}
	got, gotRev, err := r.GetAuditSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutAuditSummary(AuditSummary{}, 0); err == nil {
		t.Error("PutAuditSummary() accepted an empty entity")
	}
}

func TestAuditPolicyRoundTrip(t *testing.T) {
	x := SampleAuditPolicy("auditpolicy-1", when)
	got, err := ParseAuditPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditPolicy() = %v", err)
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

func TestAuditPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleAuditPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditPolicy(f); err == nil {
		t.Error("ParseAuditPolicy() accepted a record with no id")
	}
	f = SampleAuditPolicy("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseAuditPolicy(f); err == nil {
		t.Error("ParseAuditPolicy() accepted a non-numeric capacity")
	}
}

func TestAuditPolicyCollections(t *testing.T) {
	xs := []AuditPolicy{SampleAuditPolicy("a", when), SampleAuditPolicy("bb", when), SampleAuditPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditPolicys(xs)); got != 2 {
		t.Errorf("ActiveAuditPolicys() kept %d, want 2", got)
	}
	if got, want := SumAuditPolicys(xs), 1+3; got != want {
		t.Errorf("SumAuditPolicys() = %d, want %d", got, want)
	}
	SortAuditPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditPolicys() put %q first", xs[0].ID)
	}
	if len(IndexAuditPolicys(xs)) != 3 {
		t.Errorf("IndexAuditPolicys() = %v", IndexAuditPolicys(xs))
	}
	if err := ValidateAuditPolicys(ActiveAuditPolicys(xs)); err != nil {
		t.Errorf("ValidateAuditPolicys() = %v", err)
	}
}

func TestAuditPolicyNormalizeClamps(t *testing.T) {
	x := SampleAuditPolicy("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditPolicy("r1", when)
	rev, err := r.PutAuditPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutAuditPolicy() = %v", err)
	}
	got, gotRev, err := r.GetAuditPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutAuditPolicy(AuditPolicy{}, 0); err == nil {
		t.Error("PutAuditPolicy() accepted an empty entity")
	}
}

func TestAuditLinkRoundTrip(t *testing.T) {
	x := SampleAuditLink("auditlink-1", when)
	got, err := ParseAuditLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditLink() = %v", err)
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

func TestAuditLinkParseRejectsBadInput(t *testing.T) {
	f := SampleAuditLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditLink(f); err == nil {
		t.Error("ParseAuditLink() accepted a record with no id")
	}
	f = SampleAuditLink("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseAuditLink(f); err == nil {
		t.Error("ParseAuditLink() accepted a non-numeric capacity")
	}
}

func TestAuditLinkCollections(t *testing.T) {
	xs := []AuditLink{SampleAuditLink("a", when), SampleAuditLink("bb", when), SampleAuditLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditLinks(xs)); got != 2 {
		t.Errorf("ActiveAuditLinks() kept %d, want 2", got)
	}
	if got, want := SumAuditLinks(xs), 1+3; got != want {
		t.Errorf("SumAuditLinks() = %d, want %d", got, want)
	}
	SortAuditLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditLinks() put %q first", xs[0].ID)
	}
	if len(IndexAuditLinks(xs)) != 3 {
		t.Errorf("IndexAuditLinks() = %v", IndexAuditLinks(xs))
	}
	if err := ValidateAuditLinks(ActiveAuditLinks(xs)); err != nil {
		t.Errorf("ValidateAuditLinks() = %v", err)
	}
}

func TestAuditLinkNormalizeClamps(t *testing.T) {
	x := SampleAuditLink("a", when)
	x.Category = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditLink("r1", when)
	rev, err := r.PutAuditLink(x, 0)
	if err != nil {
		t.Fatalf("PutAuditLink() = %v", err)
	}
	got, gotRev, err := r.GetAuditLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditLinks() = %v, %v", all, err)
	}
	if _, err := r.PutAuditLink(AuditLink{}, 0); err == nil {
		t.Error("PutAuditLink() accepted an empty entity")
	}
}

func TestAuditNoteRoundTrip(t *testing.T) {
	x := SampleAuditNote("auditnote-1", when)
	got, err := ParseAuditNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditNote() = %v", err)
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

func TestAuditNoteParseRejectsBadInput(t *testing.T) {
	f := SampleAuditNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditNote(f); err == nil {
		t.Error("ParseAuditNote() accepted a record with no id")
	}
	f = SampleAuditNote("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseAuditNote(f); err == nil {
		t.Error("ParseAuditNote() accepted a non-numeric retained")
	}
}

func TestAuditNoteCollections(t *testing.T) {
	xs := []AuditNote{SampleAuditNote("a", when), SampleAuditNote("bb", when), SampleAuditNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditNotes(xs)); got != 2 {
		t.Errorf("ActiveAuditNotes() kept %d, want 2", got)
	}
	if got, want := SumAuditNotes(xs), 1+3; got != want {
		t.Errorf("SumAuditNotes() = %d, want %d", got, want)
	}
	SortAuditNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditNotes() put %q first", xs[0].ID)
	}
	if len(IndexAuditNotes(xs)) != 3 {
		t.Errorf("IndexAuditNotes() = %v", IndexAuditNotes(xs))
	}
	if err := ValidateAuditNotes(ActiveAuditNotes(xs)); err != nil {
		t.Errorf("ValidateAuditNotes() = %v", err)
	}
}

func TestAuditNoteNormalizeClamps(t *testing.T) {
	x := SampleAuditNote("a", when)
	x.Label = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditNote("r1", when)
	rev, err := r.PutAuditNote(x, 0)
	if err != nil {
		t.Fatalf("PutAuditNote() = %v", err)
	}
	got, gotRev, err := r.GetAuditNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditNotes() = %v, %v", all, err)
	}
	if _, err := r.PutAuditNote(AuditNote{}, 0); err == nil {
		t.Error("PutAuditNote() accepted an empty entity")
	}
}

func TestAuditWindowRoundTrip(t *testing.T) {
	x := SampleAuditWindow("auditwindow-1", when)
	got, err := ParseAuditWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditWindow() = %v", err)
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

func TestAuditWindowParseRejectsBadInput(t *testing.T) {
	f := SampleAuditWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditWindow(f); err == nil {
		t.Error("ParseAuditWindow() accepted a record with no id")
	}
	f = SampleAuditWindow("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseAuditWindow(f); err == nil {
		t.Error("ParseAuditWindow() accepted a non-numeric capacity")
	}
}

func TestAuditWindowCollections(t *testing.T) {
	xs := []AuditWindow{SampleAuditWindow("a", when), SampleAuditWindow("bb", when), SampleAuditWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditWindows(xs)); got != 2 {
		t.Errorf("ActiveAuditWindows() kept %d, want 2", got)
	}
	if got, want := SumAuditWindows(xs), 1+3; got != want {
		t.Errorf("SumAuditWindows() = %d, want %d", got, want)
	}
	SortAuditWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditWindows() put %q first", xs[0].ID)
	}
	if len(IndexAuditWindows(xs)) != 3 {
		t.Errorf("IndexAuditWindows() = %v", IndexAuditWindows(xs))
	}
	if err := ValidateAuditWindows(ActiveAuditWindows(xs)); err != nil {
		t.Errorf("ValidateAuditWindows() = %v", err)
	}
}

func TestAuditWindowNormalizeClamps(t *testing.T) {
	x := SampleAuditWindow("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditWindow("r1", when)
	rev, err := r.PutAuditWindow(x, 0)
	if err != nil {
		t.Fatalf("PutAuditWindow() = %v", err)
	}
	got, gotRev, err := r.GetAuditWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditWindows() = %v, %v", all, err)
	}
	if _, err := r.PutAuditWindow(AuditWindow{}, 0); err == nil {
		t.Error("PutAuditWindow() accepted an empty entity")
	}
}

func TestAuditQuotaRoundTrip(t *testing.T) {
	x := SampleAuditQuota("auditquota-1", when)
	got, err := ParseAuditQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditQuota() = %v", err)
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

func TestAuditQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleAuditQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditQuota(f); err == nil {
		t.Error("ParseAuditQuota() accepted a record with no id")
	}
	f = SampleAuditQuota("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseAuditQuota(f); err == nil {
		t.Error("ParseAuditQuota() accepted a non-numeric weight")
	}
}

func TestAuditQuotaCollections(t *testing.T) {
	xs := []AuditQuota{SampleAuditQuota("a", when), SampleAuditQuota("bb", when), SampleAuditQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditQuotas(xs)); got != 2 {
		t.Errorf("ActiveAuditQuotas() kept %d, want 2", got)
	}
	if got, want := SumAuditQuotas(xs), 1+3; got != want {
		t.Errorf("SumAuditQuotas() = %d, want %d", got, want)
	}
	SortAuditQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditQuotas() put %q first", xs[0].ID)
	}
	if len(IndexAuditQuotas(xs)) != 3 {
		t.Errorf("IndexAuditQuotas() = %v", IndexAuditQuotas(xs))
	}
	if err := ValidateAuditQuotas(ActiveAuditQuotas(xs)); err != nil {
		t.Errorf("ValidateAuditQuotas() = %v", err)
	}
}

func TestAuditQuotaNormalizeClamps(t *testing.T) {
	x := SampleAuditQuota("a", when)
	x.Category = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditQuota("r1", when)
	rev, err := r.PutAuditQuota(x, 0)
	if err != nil {
		t.Fatalf("PutAuditQuota() = %v", err)
	}
	got, gotRev, err := r.GetAuditQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutAuditQuota(AuditQuota{}, 0); err == nil {
		t.Error("PutAuditQuota() accepted an empty entity")
	}
}

func TestAuditBucketRoundTrip(t *testing.T) {
	x := SampleAuditBucket("auditbucket-1", when)
	got, err := ParseAuditBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditBucket() = %v", err)
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

func TestAuditBucketParseRejectsBadInput(t *testing.T) {
	f := SampleAuditBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditBucket(f); err == nil {
		t.Error("ParseAuditBucket() accepted a record with no id")
	}
	f = SampleAuditBucket("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseAuditBucket(f); err == nil {
		t.Error("ParseAuditBucket() accepted a non-numeric count")
	}
}

func TestAuditBucketCollections(t *testing.T) {
	xs := []AuditBucket{SampleAuditBucket("a", when), SampleAuditBucket("bb", when), SampleAuditBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditBuckets(xs)); got != 2 {
		t.Errorf("ActiveAuditBuckets() kept %d, want 2", got)
	}
	if got, want := SumAuditBuckets(xs), 1+3; got != want {
		t.Errorf("SumAuditBuckets() = %d, want %d", got, want)
	}
	SortAuditBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditBuckets() put %q first", xs[0].ID)
	}
	if len(IndexAuditBuckets(xs)) != 3 {
		t.Errorf("IndexAuditBuckets() = %v", IndexAuditBuckets(xs))
	}
	if err := ValidateAuditBuckets(ActiveAuditBuckets(xs)); err != nil {
		t.Errorf("ValidateAuditBuckets() = %v", err)
	}
}

func TestAuditBucketNormalizeClamps(t *testing.T) {
	x := SampleAuditBucket("a", when)
	x.Label = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditBucket("r1", when)
	rev, err := r.PutAuditBucket(x, 0)
	if err != nil {
		t.Fatalf("PutAuditBucket() = %v", err)
	}
	got, gotRev, err := r.GetAuditBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutAuditBucket(AuditBucket{}, 0); err == nil {
		t.Error("PutAuditBucket() accepted an empty entity")
	}
}

func TestAuditLaneRoundTrip(t *testing.T) {
	x := SampleAuditLane("auditlane-1", when)
	got, err := ParseAuditLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditLane() = %v", err)
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

func TestAuditLaneParseRejectsBadInput(t *testing.T) {
	f := SampleAuditLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditLane(f); err == nil {
		t.Error("ParseAuditLane() accepted a record with no id")
	}
	f = SampleAuditLane("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseAuditLane(f); err == nil {
		t.Error("ParseAuditLane() accepted a non-numeric attempts")
	}
}

func TestAuditLaneCollections(t *testing.T) {
	xs := []AuditLane{SampleAuditLane("a", when), SampleAuditLane("bb", when), SampleAuditLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditLanes(xs)); got != 2 {
		t.Errorf("ActiveAuditLanes() kept %d, want 2", got)
	}
	if got, want := SumAuditLanes(xs), 1+3; got != want {
		t.Errorf("SumAuditLanes() = %d, want %d", got, want)
	}
	SortAuditLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditLanes() put %q first", xs[0].ID)
	}
	if len(IndexAuditLanes(xs)) != 3 {
		t.Errorf("IndexAuditLanes() = %v", IndexAuditLanes(xs))
	}
	if err := ValidateAuditLanes(ActiveAuditLanes(xs)); err != nil {
		t.Errorf("ValidateAuditLanes() = %v", err)
	}
}

func TestAuditLaneNormalizeClamps(t *testing.T) {
	x := SampleAuditLane("a", when)
	x.Source = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditLane("r1", when)
	rev, err := r.PutAuditLane(x, 0)
	if err != nil {
		t.Fatalf("PutAuditLane() = %v", err)
	}
	got, gotRev, err := r.GetAuditLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditLanes() = %v, %v", all, err)
	}
	if _, err := r.PutAuditLane(AuditLane{}, 0); err == nil {
		t.Error("PutAuditLane() accepted an empty entity")
	}
}

func TestAuditSlotRoundTrip(t *testing.T) {
	x := SampleAuditSlot("auditslot-1", when)
	got, err := ParseAuditSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditSlot() = %v", err)
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

func TestAuditSlotParseRejectsBadInput(t *testing.T) {
	f := SampleAuditSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditSlot(f); err == nil {
		t.Error("ParseAuditSlot() accepted a record with no id")
	}
	f = SampleAuditSlot("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseAuditSlot(f); err == nil {
		t.Error("ParseAuditSlot() accepted a non-numeric offset")
	}
}

func TestAuditSlotCollections(t *testing.T) {
	xs := []AuditSlot{SampleAuditSlot("a", when), SampleAuditSlot("bb", when), SampleAuditSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditSlots(xs)); got != 2 {
		t.Errorf("ActiveAuditSlots() kept %d, want 2", got)
	}
	if got, want := SumAuditSlots(xs), 1+3; got != want {
		t.Errorf("SumAuditSlots() = %d, want %d", got, want)
	}
	SortAuditSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditSlots() put %q first", xs[0].ID)
	}
	if len(IndexAuditSlots(xs)) != 3 {
		t.Errorf("IndexAuditSlots() = %v", IndexAuditSlots(xs))
	}
	if err := ValidateAuditSlots(ActiveAuditSlots(xs)); err != nil {
		t.Errorf("ValidateAuditSlots() = %v", err)
	}
}

func TestAuditSlotNormalizeClamps(t *testing.T) {
	x := SampleAuditSlot("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditSlot("r1", when)
	rev, err := r.PutAuditSlot(x, 0)
	if err != nil {
		t.Fatalf("PutAuditSlot() = %v", err)
	}
	got, gotRev, err := r.GetAuditSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditSlots() = %v, %v", all, err)
	}
	if _, err := r.PutAuditSlot(AuditSlot{}, 0); err == nil {
		t.Error("PutAuditSlot() accepted an empty entity")
	}
}

func TestAuditTagRoundTrip(t *testing.T) {
	x := SampleAuditTag("audittag-1", when)
	got, err := ParseAuditTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditTag() = %v", err)
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

func TestAuditTagParseRejectsBadInput(t *testing.T) {
	f := SampleAuditTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditTag(f); err == nil {
		t.Error("ParseAuditTag() accepted a record with no id")
	}
	f = SampleAuditTag("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseAuditTag(f); err == nil {
		t.Error("ParseAuditTag() accepted a non-numeric offset")
	}
}

func TestAuditTagCollections(t *testing.T) {
	xs := []AuditTag{SampleAuditTag("a", when), SampleAuditTag("bb", when), SampleAuditTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditTags(xs)); got != 2 {
		t.Errorf("ActiveAuditTags() kept %d, want 2", got)
	}
	if got, want := SumAuditTags(xs), 1+3; got != want {
		t.Errorf("SumAuditTags() = %d, want %d", got, want)
	}
	SortAuditTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditTags() put %q first", xs[0].ID)
	}
	if len(IndexAuditTags(xs)) != 3 {
		t.Errorf("IndexAuditTags() = %v", IndexAuditTags(xs))
	}
	if err := ValidateAuditTags(ActiveAuditTags(xs)); err != nil {
		t.Errorf("ValidateAuditTags() = %v", err)
	}
}

func TestAuditTagNormalizeClamps(t *testing.T) {
	x := SampleAuditTag("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditTag("r1", when)
	rev, err := r.PutAuditTag(x, 0)
	if err != nil {
		t.Fatalf("PutAuditTag() = %v", err)
	}
	got, gotRev, err := r.GetAuditTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditTags() = %v, %v", all, err)
	}
	if _, err := r.PutAuditTag(AuditTag{}, 0); err == nil {
		t.Error("PutAuditTag() accepted an empty entity")
	}
}

func TestAuditRuleRoundTrip(t *testing.T) {
	x := SampleAuditRule("auditrule-1", when)
	got, err := ParseAuditRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditRule() = %v", err)
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

func TestAuditRuleParseRejectsBadInput(t *testing.T) {
	f := SampleAuditRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditRule(f); err == nil {
		t.Error("ParseAuditRule() accepted a record with no id")
	}
	f = SampleAuditRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseAuditRule(f); err == nil {
		t.Error("ParseAuditRule() accepted a non-numeric offset")
	}
}

func TestAuditRuleCollections(t *testing.T) {
	xs := []AuditRule{SampleAuditRule("a", when), SampleAuditRule("bb", when), SampleAuditRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditRules(xs)); got != 2 {
		t.Errorf("ActiveAuditRules() kept %d, want 2", got)
	}
	if got, want := SumAuditRules(xs), 1+3; got != want {
		t.Errorf("SumAuditRules() = %d, want %d", got, want)
	}
	SortAuditRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditRules() put %q first", xs[0].ID)
	}
	if len(IndexAuditRules(xs)) != 3 {
		t.Errorf("IndexAuditRules() = %v", IndexAuditRules(xs))
	}
	if err := ValidateAuditRules(ActiveAuditRules(xs)); err != nil {
		t.Errorf("ValidateAuditRules() = %v", err)
	}
}

func TestAuditRuleNormalizeClamps(t *testing.T) {
	x := SampleAuditRule("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditRule("r1", when)
	rev, err := r.PutAuditRule(x, 0)
	if err != nil {
		t.Fatalf("PutAuditRule() = %v", err)
	}
	got, gotRev, err := r.GetAuditRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditRules() = %v, %v", all, err)
	}
	if _, err := r.PutAuditRule(AuditRule{}, 0); err == nil {
		t.Error("PutAuditRule() accepted an empty entity")
	}
}

func TestAuditHoldRoundTrip(t *testing.T) {
	x := SampleAuditHold("audithold-1", when)
	got, err := ParseAuditHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditHold() = %v", err)
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

func TestAuditHoldParseRejectsBadInput(t *testing.T) {
	f := SampleAuditHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditHold(f); err == nil {
		t.Error("ParseAuditHold() accepted a record with no id")
	}
	f = SampleAuditHold("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseAuditHold(f); err == nil {
		t.Error("ParseAuditHold() accepted a non-numeric count")
	}
}

func TestAuditHoldCollections(t *testing.T) {
	xs := []AuditHold{SampleAuditHold("a", when), SampleAuditHold("bb", when), SampleAuditHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditHolds(xs)); got != 2 {
		t.Errorf("ActiveAuditHolds() kept %d, want 2", got)
	}
	if got, want := SumAuditHolds(xs), 1+3; got != want {
		t.Errorf("SumAuditHolds() = %d, want %d", got, want)
	}
	SortAuditHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditHolds() put %q first", xs[0].ID)
	}
	if len(IndexAuditHolds(xs)) != 3 {
		t.Errorf("IndexAuditHolds() = %v", IndexAuditHolds(xs))
	}
	if err := ValidateAuditHolds(ActiveAuditHolds(xs)); err != nil {
		t.Errorf("ValidateAuditHolds() = %v", err)
	}
}

func TestAuditHoldNormalizeClamps(t *testing.T) {
	x := SampleAuditHold("a", when)
	x.Reference = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditHold("r1", when)
	rev, err := r.PutAuditHold(x, 0)
	if err != nil {
		t.Fatalf("PutAuditHold() = %v", err)
	}
	got, gotRev, err := r.GetAuditHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditHolds() = %v, %v", all, err)
	}
	if _, err := r.PutAuditHold(AuditHold{}, 0); err == nil {
		t.Error("PutAuditHold() accepted an empty entity")
	}
}

func TestAuditClaimRoundTrip(t *testing.T) {
	x := SampleAuditClaim("auditclaim-1", when)
	got, err := ParseAuditClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditClaim() = %v", err)
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

func TestAuditClaimParseRejectsBadInput(t *testing.T) {
	f := SampleAuditClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditClaim(f); err == nil {
		t.Error("ParseAuditClaim() accepted a record with no id")
	}
	f = SampleAuditClaim("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseAuditClaim(f); err == nil {
		t.Error("ParseAuditClaim() accepted a non-numeric attempts")
	}
}

func TestAuditClaimCollections(t *testing.T) {
	xs := []AuditClaim{SampleAuditClaim("a", when), SampleAuditClaim("bb", when), SampleAuditClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditClaims(xs)); got != 2 {
		t.Errorf("ActiveAuditClaims() kept %d, want 2", got)
	}
	if got, want := SumAuditClaims(xs), 1+3; got != want {
		t.Errorf("SumAuditClaims() = %d, want %d", got, want)
	}
	SortAuditClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditClaims() put %q first", xs[0].ID)
	}
	if len(IndexAuditClaims(xs)) != 3 {
		t.Errorf("IndexAuditClaims() = %v", IndexAuditClaims(xs))
	}
	if err := ValidateAuditClaims(ActiveAuditClaims(xs)); err != nil {
		t.Errorf("ValidateAuditClaims() = %v", err)
	}
}

func TestAuditClaimNormalizeClamps(t *testing.T) {
	x := SampleAuditClaim("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditClaim("r1", when)
	rev, err := r.PutAuditClaim(x, 0)
	if err != nil {
		t.Fatalf("PutAuditClaim() = %v", err)
	}
	got, gotRev, err := r.GetAuditClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditClaims() = %v, %v", all, err)
	}
	if _, err := r.PutAuditClaim(AuditClaim{}, 0); err == nil {
		t.Error("PutAuditClaim() accepted an empty entity")
	}
}

func TestAuditCursorRoundTrip(t *testing.T) {
	x := SampleAuditCursor("auditcursor-1", when)
	got, err := ParseAuditCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseAuditCursor() = %v", err)
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

func TestAuditCursorParseRejectsBadInput(t *testing.T) {
	f := SampleAuditCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseAuditCursor(f); err == nil {
		t.Error("ParseAuditCursor() accepted a record with no id")
	}
	f = SampleAuditCursor("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseAuditCursor(f); err == nil {
		t.Error("ParseAuditCursor() accepted a non-numeric attempts")
	}
}

func TestAuditCursorCollections(t *testing.T) {
	xs := []AuditCursor{SampleAuditCursor("a", when), SampleAuditCursor("bb", when), SampleAuditCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveAuditCursors(xs)); got != 2 {
		t.Errorf("ActiveAuditCursors() kept %d, want 2", got)
	}
	if got, want := SumAuditCursors(xs), 1+3; got != want {
		t.Errorf("SumAuditCursors() = %d, want %d", got, want)
	}
	SortAuditCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortAuditCursors() put %q first", xs[0].ID)
	}
	if len(IndexAuditCursors(xs)) != 3 {
		t.Errorf("IndexAuditCursors() = %v", IndexAuditCursors(xs))
	}
	if err := ValidateAuditCursors(ActiveAuditCursors(xs)); err != nil {
		t.Errorf("ValidateAuditCursors() = %v", err)
	}
}

func TestAuditCursorNormalizeClamps(t *testing.T) {
	x := SampleAuditCursor("a", when)
	x.Name = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestAuditCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleAuditCursor("r1", when)
	rev, err := r.PutAuditCursor(x, 0)
	if err != nil {
		t.Fatalf("PutAuditCursor() = %v", err)
	}
	got, gotRev, err := r.GetAuditCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetAuditCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListAuditCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListAuditCursors() = %v, %v", all, err)
	}
	if _, err := r.PutAuditCursor(AuditCursor{}, 0); err == nil {
		t.Error("PutAuditCursor() accepted an empty entity")
	}
}
