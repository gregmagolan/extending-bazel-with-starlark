package template

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
		if !strings.HasPrefix(Prefixes[i], "template/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleTemplateRecord("ok", when).Fields()); err != nil {
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

func TestTemplateRecordRoundTrip(t *testing.T) {
	x := SampleTemplateRecord("templaterecord-1", when)
	got, err := ParseTemplateRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateRecord() = %v", err)
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

func TestTemplateRecordParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateRecord(f); err == nil {
		t.Error("ParseTemplateRecord() accepted a record with no id")
	}
	f = SampleTemplateRecord("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTemplateRecord(f); err == nil {
		t.Error("ParseTemplateRecord() accepted a non-numeric version")
	}
}

func TestTemplateRecordCollections(t *testing.T) {
	xs := []TemplateRecord{SampleTemplateRecord("a", when), SampleTemplateRecord("bb", when), SampleTemplateRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateRecords(xs)); got != 2 {
		t.Errorf("ActiveTemplateRecords() kept %d, want 2", got)
	}
	if got, want := SumTemplateRecords(xs), 1+3; got != want {
		t.Errorf("SumTemplateRecords() = %d, want %d", got, want)
	}
	SortTemplateRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateRecords() put %q first", xs[0].ID)
	}
	if len(IndexTemplateRecords(xs)) != 3 {
		t.Errorf("IndexTemplateRecords() = %v", IndexTemplateRecords(xs))
	}
	if err := ValidateTemplateRecords(ActiveTemplateRecords(xs)); err != nil {
		t.Errorf("ValidateTemplateRecords() = %v", err)
	}
}

func TestTemplateRecordNormalizeClamps(t *testing.T) {
	x := SampleTemplateRecord("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateRecord("r1", when)
	rev, err := r.PutTemplateRecord(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateRecord() = %v", err)
	}
	got, gotRev, err := r.GetTemplateRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateRecords() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateRecord(TemplateRecord{}, 0); err == nil {
		t.Error("PutTemplateRecord() accepted an empty entity")
	}
}

func TestTemplateEntryRoundTrip(t *testing.T) {
	x := SampleTemplateEntry("templateentry-1", when)
	got, err := ParseTemplateEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateEntry() = %v", err)
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

func TestTemplateEntryParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateEntry(f); err == nil {
		t.Error("ParseTemplateEntry() accepted a record with no id")
	}
	f = SampleTemplateEntry("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTemplateEntry(f); err == nil {
		t.Error("ParseTemplateEntry() accepted a non-numeric priority")
	}
}

func TestTemplateEntryCollections(t *testing.T) {
	xs := []TemplateEntry{SampleTemplateEntry("a", when), SampleTemplateEntry("bb", when), SampleTemplateEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateEntrys(xs)); got != 2 {
		t.Errorf("ActiveTemplateEntrys() kept %d, want 2", got)
	}
	if got, want := SumTemplateEntrys(xs), 1+3; got != want {
		t.Errorf("SumTemplateEntrys() = %d, want %d", got, want)
	}
	SortTemplateEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateEntrys() put %q first", xs[0].ID)
	}
	if len(IndexTemplateEntrys(xs)) != 3 {
		t.Errorf("IndexTemplateEntrys() = %v", IndexTemplateEntrys(xs))
	}
	if err := ValidateTemplateEntrys(ActiveTemplateEntrys(xs)); err != nil {
		t.Errorf("ValidateTemplateEntrys() = %v", err)
	}
}

func TestTemplateEntryNormalizeClamps(t *testing.T) {
	x := SampleTemplateEntry("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateEntry("r1", when)
	rev, err := r.PutTemplateEntry(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateEntry() = %v", err)
	}
	got, gotRev, err := r.GetTemplateEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateEntry(TemplateEntry{}, 0); err == nil {
		t.Error("PutTemplateEntry() accepted an empty entity")
	}
}

func TestTemplateItemRoundTrip(t *testing.T) {
	x := SampleTemplateItem("templateitem-1", when)
	got, err := ParseTemplateItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateItem() = %v", err)
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

func TestTemplateItemParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateItem(f); err == nil {
		t.Error("ParseTemplateItem() accepted a record with no id")
	}
	f = SampleTemplateItem("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTemplateItem(f); err == nil {
		t.Error("ParseTemplateItem() accepted a non-numeric attempts")
	}
}

func TestTemplateItemCollections(t *testing.T) {
	xs := []TemplateItem{SampleTemplateItem("a", when), SampleTemplateItem("bb", when), SampleTemplateItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateItems(xs)); got != 2 {
		t.Errorf("ActiveTemplateItems() kept %d, want 2", got)
	}
	if got, want := SumTemplateItems(xs), 1+3; got != want {
		t.Errorf("SumTemplateItems() = %d, want %d", got, want)
	}
	SortTemplateItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateItems() put %q first", xs[0].ID)
	}
	if len(IndexTemplateItems(xs)) != 3 {
		t.Errorf("IndexTemplateItems() = %v", IndexTemplateItems(xs))
	}
	if err := ValidateTemplateItems(ActiveTemplateItems(xs)); err != nil {
		t.Errorf("ValidateTemplateItems() = %v", err)
	}
}

func TestTemplateItemNormalizeClamps(t *testing.T) {
	x := SampleTemplateItem("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateItem("r1", when)
	rev, err := r.PutTemplateItem(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateItem() = %v", err)
	}
	got, gotRev, err := r.GetTemplateItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateItems() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateItem(TemplateItem{}, 0); err == nil {
		t.Error("PutTemplateItem() accepted an empty entity")
	}
}

func TestTemplateDraftRoundTrip(t *testing.T) {
	x := SampleTemplateDraft("templatedraft-1", when)
	got, err := ParseTemplateDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateDraft() = %v", err)
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

func TestTemplateDraftParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateDraft(f); err == nil {
		t.Error("ParseTemplateDraft() accepted a record with no id")
	}
	f = SampleTemplateDraft("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTemplateDraft(f); err == nil {
		t.Error("ParseTemplateDraft() accepted a non-numeric attempts")
	}
}

func TestTemplateDraftCollections(t *testing.T) {
	xs := []TemplateDraft{SampleTemplateDraft("a", when), SampleTemplateDraft("bb", when), SampleTemplateDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateDrafts(xs)); got != 2 {
		t.Errorf("ActiveTemplateDrafts() kept %d, want 2", got)
	}
	if got, want := SumTemplateDrafts(xs), 1+3; got != want {
		t.Errorf("SumTemplateDrafts() = %d, want %d", got, want)
	}
	SortTemplateDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateDrafts() put %q first", xs[0].ID)
	}
	if len(IndexTemplateDrafts(xs)) != 3 {
		t.Errorf("IndexTemplateDrafts() = %v", IndexTemplateDrafts(xs))
	}
	if err := ValidateTemplateDrafts(ActiveTemplateDrafts(xs)); err != nil {
		t.Errorf("ValidateTemplateDrafts() = %v", err)
	}
}

func TestTemplateDraftNormalizeClamps(t *testing.T) {
	x := SampleTemplateDraft("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateDraft("r1", when)
	rev, err := r.PutTemplateDraft(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateDraft() = %v", err)
	}
	got, gotRev, err := r.GetTemplateDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateDraft(TemplateDraft{}, 0); err == nil {
		t.Error("PutTemplateDraft() accepted an empty entity")
	}
}

func TestTemplateRevisionRoundTrip(t *testing.T) {
	x := SampleTemplateRevision("templaterevision-1", when)
	got, err := ParseTemplateRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateRevision() = %v", err)
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

func TestTemplateRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateRevision(f); err == nil {
		t.Error("ParseTemplateRevision() accepted a record with no id")
	}
	f = SampleTemplateRevision("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTemplateRevision(f); err == nil {
		t.Error("ParseTemplateRevision() accepted a non-numeric count")
	}
}

func TestTemplateRevisionCollections(t *testing.T) {
	xs := []TemplateRevision{SampleTemplateRevision("a", when), SampleTemplateRevision("bb", when), SampleTemplateRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateRevisions(xs)); got != 2 {
		t.Errorf("ActiveTemplateRevisions() kept %d, want 2", got)
	}
	if got, want := SumTemplateRevisions(xs), 1+3; got != want {
		t.Errorf("SumTemplateRevisions() = %d, want %d", got, want)
	}
	SortTemplateRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateRevisions() put %q first", xs[0].ID)
	}
	if len(IndexTemplateRevisions(xs)) != 3 {
		t.Errorf("IndexTemplateRevisions() = %v", IndexTemplateRevisions(xs))
	}
	if err := ValidateTemplateRevisions(ActiveTemplateRevisions(xs)); err != nil {
		t.Errorf("ValidateTemplateRevisions() = %v", err)
	}
}

func TestTemplateRevisionNormalizeClamps(t *testing.T) {
	x := SampleTemplateRevision("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateRevision("r1", when)
	rev, err := r.PutTemplateRevision(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateRevision() = %v", err)
	}
	got, gotRev, err := r.GetTemplateRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateRevision(TemplateRevision{}, 0); err == nil {
		t.Error("PutTemplateRevision() accepted an empty entity")
	}
}

func TestTemplateSnapshotRoundTrip(t *testing.T) {
	x := SampleTemplateSnapshot("templatesnapshot-1", when)
	got, err := ParseTemplateSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateSnapshot() = %v", err)
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

func TestTemplateSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateSnapshot(f); err == nil {
		t.Error("ParseTemplateSnapshot() accepted a record with no id")
	}
	f = SampleTemplateSnapshot("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTemplateSnapshot(f); err == nil {
		t.Error("ParseTemplateSnapshot() accepted a non-numeric version")
	}
}

func TestTemplateSnapshotCollections(t *testing.T) {
	xs := []TemplateSnapshot{SampleTemplateSnapshot("a", when), SampleTemplateSnapshot("bb", when), SampleTemplateSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateSnapshots(xs)); got != 2 {
		t.Errorf("ActiveTemplateSnapshots() kept %d, want 2", got)
	}
	if got, want := SumTemplateSnapshots(xs), 1+3; got != want {
		t.Errorf("SumTemplateSnapshots() = %d, want %d", got, want)
	}
	SortTemplateSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexTemplateSnapshots(xs)) != 3 {
		t.Errorf("IndexTemplateSnapshots() = %v", IndexTemplateSnapshots(xs))
	}
	if err := ValidateTemplateSnapshots(ActiveTemplateSnapshots(xs)); err != nil {
		t.Errorf("ValidateTemplateSnapshots() = %v", err)
	}
}

func TestTemplateSnapshotNormalizeClamps(t *testing.T) {
	x := SampleTemplateSnapshot("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateSnapshot("r1", when)
	rev, err := r.PutTemplateSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetTemplateSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateSnapshot(TemplateSnapshot{}, 0); err == nil {
		t.Error("PutTemplateSnapshot() accepted an empty entity")
	}
}

func TestTemplateBatchRoundTrip(t *testing.T) {
	x := SampleTemplateBatch("templatebatch-1", when)
	got, err := ParseTemplateBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateBatch() = %v", err)
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

func TestTemplateBatchParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateBatch(f); err == nil {
		t.Error("ParseTemplateBatch() accepted a record with no id")
	}
	f = SampleTemplateBatch("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTemplateBatch(f); err == nil {
		t.Error("ParseTemplateBatch() accepted a non-numeric retained")
	}
}

func TestTemplateBatchCollections(t *testing.T) {
	xs := []TemplateBatch{SampleTemplateBatch("a", when), SampleTemplateBatch("bb", when), SampleTemplateBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateBatchs(xs)); got != 2 {
		t.Errorf("ActiveTemplateBatchs() kept %d, want 2", got)
	}
	if got, want := SumTemplateBatchs(xs), 1+3; got != want {
		t.Errorf("SumTemplateBatchs() = %d, want %d", got, want)
	}
	SortTemplateBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateBatchs() put %q first", xs[0].ID)
	}
	if len(IndexTemplateBatchs(xs)) != 3 {
		t.Errorf("IndexTemplateBatchs() = %v", IndexTemplateBatchs(xs))
	}
	if err := ValidateTemplateBatchs(ActiveTemplateBatchs(xs)); err != nil {
		t.Errorf("ValidateTemplateBatchs() = %v", err)
	}
}

func TestTemplateBatchNormalizeClamps(t *testing.T) {
	x := SampleTemplateBatch("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateBatch("r1", when)
	rev, err := r.PutTemplateBatch(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateBatch() = %v", err)
	}
	got, gotRev, err := r.GetTemplateBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateBatch(TemplateBatch{}, 0); err == nil {
		t.Error("PutTemplateBatch() accepted an empty entity")
	}
}

func TestTemplateRequestRoundTrip(t *testing.T) {
	x := SampleTemplateRequest("templaterequest-1", when)
	got, err := ParseTemplateRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateRequest() = %v", err)
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

func TestTemplateRequestParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateRequest(f); err == nil {
		t.Error("ParseTemplateRequest() accepted a record with no id")
	}
	f = SampleTemplateRequest("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTemplateRequest(f); err == nil {
		t.Error("ParseTemplateRequest() accepted a non-numeric capacity")
	}
}

func TestTemplateRequestCollections(t *testing.T) {
	xs := []TemplateRequest{SampleTemplateRequest("a", when), SampleTemplateRequest("bb", when), SampleTemplateRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateRequests(xs)); got != 2 {
		t.Errorf("ActiveTemplateRequests() kept %d, want 2", got)
	}
	if got, want := SumTemplateRequests(xs), 1+3; got != want {
		t.Errorf("SumTemplateRequests() = %d, want %d", got, want)
	}
	SortTemplateRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateRequests() put %q first", xs[0].ID)
	}
	if len(IndexTemplateRequests(xs)) != 3 {
		t.Errorf("IndexTemplateRequests() = %v", IndexTemplateRequests(xs))
	}
	if err := ValidateTemplateRequests(ActiveTemplateRequests(xs)); err != nil {
		t.Errorf("ValidateTemplateRequests() = %v", err)
	}
}

func TestTemplateRequestNormalizeClamps(t *testing.T) {
	x := SampleTemplateRequest("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateRequest("r1", when)
	rev, err := r.PutTemplateRequest(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateRequest() = %v", err)
	}
	got, gotRev, err := r.GetTemplateRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateRequests() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateRequest(TemplateRequest{}, 0); err == nil {
		t.Error("PutTemplateRequest() accepted an empty entity")
	}
}

func TestTemplateSummaryRoundTrip(t *testing.T) {
	x := SampleTemplateSummary("templatesummary-1", when)
	got, err := ParseTemplateSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateSummary() = %v", err)
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

func TestTemplateSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateSummary(f); err == nil {
		t.Error("ParseTemplateSummary() accepted a record with no id")
	}
	f = SampleTemplateSummary("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTemplateSummary(f); err == nil {
		t.Error("ParseTemplateSummary() accepted a non-numeric retained")
	}
}

func TestTemplateSummaryCollections(t *testing.T) {
	xs := []TemplateSummary{SampleTemplateSummary("a", when), SampleTemplateSummary("bb", when), SampleTemplateSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateSummarys(xs)); got != 2 {
		t.Errorf("ActiveTemplateSummarys() kept %d, want 2", got)
	}
	if got, want := SumTemplateSummarys(xs), 1+3; got != want {
		t.Errorf("SumTemplateSummarys() = %d, want %d", got, want)
	}
	SortTemplateSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateSummarys() put %q first", xs[0].ID)
	}
	if len(IndexTemplateSummarys(xs)) != 3 {
		t.Errorf("IndexTemplateSummarys() = %v", IndexTemplateSummarys(xs))
	}
	if err := ValidateTemplateSummarys(ActiveTemplateSummarys(xs)); err != nil {
		t.Errorf("ValidateTemplateSummarys() = %v", err)
	}
}

func TestTemplateSummaryNormalizeClamps(t *testing.T) {
	x := SampleTemplateSummary("a", when)
	x.Name = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateSummary("r1", when)
	rev, err := r.PutTemplateSummary(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateSummary() = %v", err)
	}
	got, gotRev, err := r.GetTemplateSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateSummary(TemplateSummary{}, 0); err == nil {
		t.Error("PutTemplateSummary() accepted an empty entity")
	}
}

func TestTemplatePolicyRoundTrip(t *testing.T) {
	x := SampleTemplatePolicy("templatepolicy-1", when)
	got, err := ParseTemplatePolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplatePolicy() = %v", err)
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

func TestTemplatePolicyParseRejectsBadInput(t *testing.T) {
	f := SampleTemplatePolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplatePolicy(f); err == nil {
		t.Error("ParseTemplatePolicy() accepted a record with no id")
	}
	f = SampleTemplatePolicy("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseTemplatePolicy(f); err == nil {
		t.Error("ParseTemplatePolicy() accepted a non-numeric capacity")
	}
}

func TestTemplatePolicyCollections(t *testing.T) {
	xs := []TemplatePolicy{SampleTemplatePolicy("a", when), SampleTemplatePolicy("bb", when), SampleTemplatePolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplatePolicys(xs)); got != 2 {
		t.Errorf("ActiveTemplatePolicys() kept %d, want 2", got)
	}
	if got, want := SumTemplatePolicys(xs), 1+3; got != want {
		t.Errorf("SumTemplatePolicys() = %d, want %d", got, want)
	}
	SortTemplatePolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplatePolicys() put %q first", xs[0].ID)
	}
	if len(IndexTemplatePolicys(xs)) != 3 {
		t.Errorf("IndexTemplatePolicys() = %v", IndexTemplatePolicys(xs))
	}
	if err := ValidateTemplatePolicys(ActiveTemplatePolicys(xs)); err != nil {
		t.Errorf("ValidateTemplatePolicys() = %v", err)
	}
}

func TestTemplatePolicyNormalizeClamps(t *testing.T) {
	x := SampleTemplatePolicy("a", when)
	x.Label = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplatePolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplatePolicy("r1", when)
	rev, err := r.PutTemplatePolicy(x, 0)
	if err != nil {
		t.Fatalf("PutTemplatePolicy() = %v", err)
	}
	got, gotRev, err := r.GetTemplatePolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplatePolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplatePolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplatePolicys() = %v, %v", all, err)
	}
	if _, err := r.PutTemplatePolicy(TemplatePolicy{}, 0); err == nil {
		t.Error("PutTemplatePolicy() accepted an empty entity")
	}
}

func TestTemplateLinkRoundTrip(t *testing.T) {
	x := SampleTemplateLink("templatelink-1", when)
	got, err := ParseTemplateLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateLink() = %v", err)
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

func TestTemplateLinkParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateLink(f); err == nil {
		t.Error("ParseTemplateLink() accepted a record with no id")
	}
	f = SampleTemplateLink("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTemplateLink(f); err == nil {
		t.Error("ParseTemplateLink() accepted a non-numeric attempts")
	}
}

func TestTemplateLinkCollections(t *testing.T) {
	xs := []TemplateLink{SampleTemplateLink("a", when), SampleTemplateLink("bb", when), SampleTemplateLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateLinks(xs)); got != 2 {
		t.Errorf("ActiveTemplateLinks() kept %d, want 2", got)
	}
	if got, want := SumTemplateLinks(xs), 1+3; got != want {
		t.Errorf("SumTemplateLinks() = %d, want %d", got, want)
	}
	SortTemplateLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateLinks() put %q first", xs[0].ID)
	}
	if len(IndexTemplateLinks(xs)) != 3 {
		t.Errorf("IndexTemplateLinks() = %v", IndexTemplateLinks(xs))
	}
	if err := ValidateTemplateLinks(ActiveTemplateLinks(xs)); err != nil {
		t.Errorf("ValidateTemplateLinks() = %v", err)
	}
}

func TestTemplateLinkNormalizeClamps(t *testing.T) {
	x := SampleTemplateLink("a", when)
	x.Region = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateLink("r1", when)
	rev, err := r.PutTemplateLink(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateLink() = %v", err)
	}
	got, gotRev, err := r.GetTemplateLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateLinks() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateLink(TemplateLink{}, 0); err == nil {
		t.Error("PutTemplateLink() accepted an empty entity")
	}
}

func TestTemplateNoteRoundTrip(t *testing.T) {
	x := SampleTemplateNote("templatenote-1", when)
	got, err := ParseTemplateNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateNote() = %v", err)
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

func TestTemplateNoteParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateNote(f); err == nil {
		t.Error("ParseTemplateNote() accepted a record with no id")
	}
	f = SampleTemplateNote("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTemplateNote(f); err == nil {
		t.Error("ParseTemplateNote() accepted a non-numeric version")
	}
}

func TestTemplateNoteCollections(t *testing.T) {
	xs := []TemplateNote{SampleTemplateNote("a", when), SampleTemplateNote("bb", when), SampleTemplateNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateNotes(xs)); got != 2 {
		t.Errorf("ActiveTemplateNotes() kept %d, want 2", got)
	}
	if got, want := SumTemplateNotes(xs), 1+3; got != want {
		t.Errorf("SumTemplateNotes() = %d, want %d", got, want)
	}
	SortTemplateNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateNotes() put %q first", xs[0].ID)
	}
	if len(IndexTemplateNotes(xs)) != 3 {
		t.Errorf("IndexTemplateNotes() = %v", IndexTemplateNotes(xs))
	}
	if err := ValidateTemplateNotes(ActiveTemplateNotes(xs)); err != nil {
		t.Errorf("ValidateTemplateNotes() = %v", err)
	}
}

func TestTemplateNoteNormalizeClamps(t *testing.T) {
	x := SampleTemplateNote("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateNote("r1", when)
	rev, err := r.PutTemplateNote(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateNote() = %v", err)
	}
	got, gotRev, err := r.GetTemplateNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateNotes() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateNote(TemplateNote{}, 0); err == nil {
		t.Error("PutTemplateNote() accepted an empty entity")
	}
}

func TestTemplateWindowRoundTrip(t *testing.T) {
	x := SampleTemplateWindow("templatewindow-1", when)
	got, err := ParseTemplateWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateWindow() = %v", err)
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

func TestTemplateWindowParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateWindow(f); err == nil {
		t.Error("ParseTemplateWindow() accepted a record with no id")
	}
	f = SampleTemplateWindow("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTemplateWindow(f); err == nil {
		t.Error("ParseTemplateWindow() accepted a non-numeric retained")
	}
}

func TestTemplateWindowCollections(t *testing.T) {
	xs := []TemplateWindow{SampleTemplateWindow("a", when), SampleTemplateWindow("bb", when), SampleTemplateWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateWindows(xs)); got != 2 {
		t.Errorf("ActiveTemplateWindows() kept %d, want 2", got)
	}
	if got, want := SumTemplateWindows(xs), 1+3; got != want {
		t.Errorf("SumTemplateWindows() = %d, want %d", got, want)
	}
	SortTemplateWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateWindows() put %q first", xs[0].ID)
	}
	if len(IndexTemplateWindows(xs)) != 3 {
		t.Errorf("IndexTemplateWindows() = %v", IndexTemplateWindows(xs))
	}
	if err := ValidateTemplateWindows(ActiveTemplateWindows(xs)); err != nil {
		t.Errorf("ValidateTemplateWindows() = %v", err)
	}
}

func TestTemplateWindowNormalizeClamps(t *testing.T) {
	x := SampleTemplateWindow("a", when)
	x.Comment = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateWindow("r1", when)
	rev, err := r.PutTemplateWindow(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateWindow() = %v", err)
	}
	got, gotRev, err := r.GetTemplateWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateWindows() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateWindow(TemplateWindow{}, 0); err == nil {
		t.Error("PutTemplateWindow() accepted an empty entity")
	}
}

func TestTemplateQuotaRoundTrip(t *testing.T) {
	x := SampleTemplateQuota("templatequota-1", when)
	got, err := ParseTemplateQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateQuota() = %v", err)
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

func TestTemplateQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateQuota(f); err == nil {
		t.Error("ParseTemplateQuota() accepted a record with no id")
	}
	f = SampleTemplateQuota("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseTemplateQuota(f); err == nil {
		t.Error("ParseTemplateQuota() accepted a non-numeric offset")
	}
}

func TestTemplateQuotaCollections(t *testing.T) {
	xs := []TemplateQuota{SampleTemplateQuota("a", when), SampleTemplateQuota("bb", when), SampleTemplateQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateQuotas(xs)); got != 2 {
		t.Errorf("ActiveTemplateQuotas() kept %d, want 2", got)
	}
	if got, want := SumTemplateQuotas(xs), 1+3; got != want {
		t.Errorf("SumTemplateQuotas() = %d, want %d", got, want)
	}
	SortTemplateQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateQuotas() put %q first", xs[0].ID)
	}
	if len(IndexTemplateQuotas(xs)) != 3 {
		t.Errorf("IndexTemplateQuotas() = %v", IndexTemplateQuotas(xs))
	}
	if err := ValidateTemplateQuotas(ActiveTemplateQuotas(xs)); err != nil {
		t.Errorf("ValidateTemplateQuotas() = %v", err)
	}
}

func TestTemplateQuotaNormalizeClamps(t *testing.T) {
	x := SampleTemplateQuota("a", when)
	x.Category = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateQuota("r1", when)
	rev, err := r.PutTemplateQuota(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateQuota() = %v", err)
	}
	got, gotRev, err := r.GetTemplateQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateQuota(TemplateQuota{}, 0); err == nil {
		t.Error("PutTemplateQuota() accepted an empty entity")
	}
}

func TestTemplateBucketRoundTrip(t *testing.T) {
	x := SampleTemplateBucket("templatebucket-1", when)
	got, err := ParseTemplateBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateBucket() = %v", err)
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

func TestTemplateBucketParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateBucket(f); err == nil {
		t.Error("ParseTemplateBucket() accepted a record with no id")
	}
	f = SampleTemplateBucket("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseTemplateBucket(f); err == nil {
		t.Error("ParseTemplateBucket() accepted a non-numeric priority")
	}
}

func TestTemplateBucketCollections(t *testing.T) {
	xs := []TemplateBucket{SampleTemplateBucket("a", when), SampleTemplateBucket("bb", when), SampleTemplateBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateBuckets(xs)); got != 2 {
		t.Errorf("ActiveTemplateBuckets() kept %d, want 2", got)
	}
	if got, want := SumTemplateBuckets(xs), 1+3; got != want {
		t.Errorf("SumTemplateBuckets() = %d, want %d", got, want)
	}
	SortTemplateBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateBuckets() put %q first", xs[0].ID)
	}
	if len(IndexTemplateBuckets(xs)) != 3 {
		t.Errorf("IndexTemplateBuckets() = %v", IndexTemplateBuckets(xs))
	}
	if err := ValidateTemplateBuckets(ActiveTemplateBuckets(xs)); err != nil {
		t.Errorf("ValidateTemplateBuckets() = %v", err)
	}
}

func TestTemplateBucketNormalizeClamps(t *testing.T) {
	x := SampleTemplateBucket("a", when)
	x.Reference = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateBucket("r1", when)
	rev, err := r.PutTemplateBucket(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateBucket() = %v", err)
	}
	got, gotRev, err := r.GetTemplateBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateBucket(TemplateBucket{}, 0); err == nil {
		t.Error("PutTemplateBucket() accepted an empty entity")
	}
}

func TestTemplateLaneRoundTrip(t *testing.T) {
	x := SampleTemplateLane("templatelane-1", when)
	got, err := ParseTemplateLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateLane() = %v", err)
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

func TestTemplateLaneParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateLane(f); err == nil {
		t.Error("ParseTemplateLane() accepted a record with no id")
	}
	f = SampleTemplateLane("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTemplateLane(f); err == nil {
		t.Error("ParseTemplateLane() accepted a non-numeric weight")
	}
}

func TestTemplateLaneCollections(t *testing.T) {
	xs := []TemplateLane{SampleTemplateLane("a", when), SampleTemplateLane("bb", when), SampleTemplateLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateLanes(xs)); got != 2 {
		t.Errorf("ActiveTemplateLanes() kept %d, want 2", got)
	}
	if got, want := SumTemplateLanes(xs), 1+3; got != want {
		t.Errorf("SumTemplateLanes() = %d, want %d", got, want)
	}
	SortTemplateLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateLanes() put %q first", xs[0].ID)
	}
	if len(IndexTemplateLanes(xs)) != 3 {
		t.Errorf("IndexTemplateLanes() = %v", IndexTemplateLanes(xs))
	}
	if err := ValidateTemplateLanes(ActiveTemplateLanes(xs)); err != nil {
		t.Errorf("ValidateTemplateLanes() = %v", err)
	}
}

func TestTemplateLaneNormalizeClamps(t *testing.T) {
	x := SampleTemplateLane("a", when)
	x.Reference = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateLane("r1", when)
	rev, err := r.PutTemplateLane(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateLane() = %v", err)
	}
	got, gotRev, err := r.GetTemplateLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateLanes() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateLane(TemplateLane{}, 0); err == nil {
		t.Error("PutTemplateLane() accepted an empty entity")
	}
}

func TestTemplateSlotRoundTrip(t *testing.T) {
	x := SampleTemplateSlot("templateslot-1", when)
	got, err := ParseTemplateSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateSlot() = %v", err)
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

func TestTemplateSlotParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateSlot(f); err == nil {
		t.Error("ParseTemplateSlot() accepted a record with no id")
	}
	f = SampleTemplateSlot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTemplateSlot(f); err == nil {
		t.Error("ParseTemplateSlot() accepted a non-numeric weight")
	}
}

func TestTemplateSlotCollections(t *testing.T) {
	xs := []TemplateSlot{SampleTemplateSlot("a", when), SampleTemplateSlot("bb", when), SampleTemplateSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateSlots(xs)); got != 2 {
		t.Errorf("ActiveTemplateSlots() kept %d, want 2", got)
	}
	if got, want := SumTemplateSlots(xs), 1+3; got != want {
		t.Errorf("SumTemplateSlots() = %d, want %d", got, want)
	}
	SortTemplateSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateSlots() put %q first", xs[0].ID)
	}
	if len(IndexTemplateSlots(xs)) != 3 {
		t.Errorf("IndexTemplateSlots() = %v", IndexTemplateSlots(xs))
	}
	if err := ValidateTemplateSlots(ActiveTemplateSlots(xs)); err != nil {
		t.Errorf("ValidateTemplateSlots() = %v", err)
	}
}

func TestTemplateSlotNormalizeClamps(t *testing.T) {
	x := SampleTemplateSlot("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateSlot("r1", when)
	rev, err := r.PutTemplateSlot(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateSlot() = %v", err)
	}
	got, gotRev, err := r.GetTemplateSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateSlots() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateSlot(TemplateSlot{}, 0); err == nil {
		t.Error("PutTemplateSlot() accepted an empty entity")
	}
}

func TestTemplateTagRoundTrip(t *testing.T) {
	x := SampleTemplateTag("templatetag-1", when)
	got, err := ParseTemplateTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateTag() = %v", err)
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

func TestTemplateTagParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateTag(f); err == nil {
		t.Error("ParseTemplateTag() accepted a record with no id")
	}
	f = SampleTemplateTag("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseTemplateTag(f); err == nil {
		t.Error("ParseTemplateTag() accepted a non-numeric count")
	}
}

func TestTemplateTagCollections(t *testing.T) {
	xs := []TemplateTag{SampleTemplateTag("a", when), SampleTemplateTag("bb", when), SampleTemplateTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateTags(xs)); got != 2 {
		t.Errorf("ActiveTemplateTags() kept %d, want 2", got)
	}
	if got, want := SumTemplateTags(xs), 1+3; got != want {
		t.Errorf("SumTemplateTags() = %d, want %d", got, want)
	}
	SortTemplateTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateTags() put %q first", xs[0].ID)
	}
	if len(IndexTemplateTags(xs)) != 3 {
		t.Errorf("IndexTemplateTags() = %v", IndexTemplateTags(xs))
	}
	if err := ValidateTemplateTags(ActiveTemplateTags(xs)); err != nil {
		t.Errorf("ValidateTemplateTags() = %v", err)
	}
}

func TestTemplateTagNormalizeClamps(t *testing.T) {
	x := SampleTemplateTag("a", when)
	x.Channel = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateTag("r1", when)
	rev, err := r.PutTemplateTag(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateTag() = %v", err)
	}
	got, gotRev, err := r.GetTemplateTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateTags() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateTag(TemplateTag{}, 0); err == nil {
		t.Error("PutTemplateTag() accepted an empty entity")
	}
}

func TestTemplateRuleRoundTrip(t *testing.T) {
	x := SampleTemplateRule("templaterule-1", when)
	got, err := ParseTemplateRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateRule() = %v", err)
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

func TestTemplateRuleParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateRule(f); err == nil {
		t.Error("ParseTemplateRule() accepted a record with no id")
	}
	f = SampleTemplateRule("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseTemplateRule(f); err == nil {
		t.Error("ParseTemplateRule() accepted a non-numeric attempts")
	}
}

func TestTemplateRuleCollections(t *testing.T) {
	xs := []TemplateRule{SampleTemplateRule("a", when), SampleTemplateRule("bb", when), SampleTemplateRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateRules(xs)); got != 2 {
		t.Errorf("ActiveTemplateRules() kept %d, want 2", got)
	}
	if got, want := SumTemplateRules(xs), 1+3; got != want {
		t.Errorf("SumTemplateRules() = %d, want %d", got, want)
	}
	SortTemplateRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateRules() put %q first", xs[0].ID)
	}
	if len(IndexTemplateRules(xs)) != 3 {
		t.Errorf("IndexTemplateRules() = %v", IndexTemplateRules(xs))
	}
	if err := ValidateTemplateRules(ActiveTemplateRules(xs)); err != nil {
		t.Errorf("ValidateTemplateRules() = %v", err)
	}
}

func TestTemplateRuleNormalizeClamps(t *testing.T) {
	x := SampleTemplateRule("a", when)
	x.Channel = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateRule("r1", when)
	rev, err := r.PutTemplateRule(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateRule() = %v", err)
	}
	got, gotRev, err := r.GetTemplateRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateRules() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateRule(TemplateRule{}, 0); err == nil {
		t.Error("PutTemplateRule() accepted an empty entity")
	}
}

func TestTemplateHoldRoundTrip(t *testing.T) {
	x := SampleTemplateHold("templatehold-1", when)
	got, err := ParseTemplateHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateHold() = %v", err)
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

func TestTemplateHoldParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateHold(f); err == nil {
		t.Error("ParseTemplateHold() accepted a record with no id")
	}
	f = SampleTemplateHold("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseTemplateHold(f); err == nil {
		t.Error("ParseTemplateHold() accepted a non-numeric weight")
	}
}

func TestTemplateHoldCollections(t *testing.T) {
	xs := []TemplateHold{SampleTemplateHold("a", when), SampleTemplateHold("bb", when), SampleTemplateHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateHolds(xs)); got != 2 {
		t.Errorf("ActiveTemplateHolds() kept %d, want 2", got)
	}
	if got, want := SumTemplateHolds(xs), 1+3; got != want {
		t.Errorf("SumTemplateHolds() = %d, want %d", got, want)
	}
	SortTemplateHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateHolds() put %q first", xs[0].ID)
	}
	if len(IndexTemplateHolds(xs)) != 3 {
		t.Errorf("IndexTemplateHolds() = %v", IndexTemplateHolds(xs))
	}
	if err := ValidateTemplateHolds(ActiveTemplateHolds(xs)); err != nil {
		t.Errorf("ValidateTemplateHolds() = %v", err)
	}
}

func TestTemplateHoldNormalizeClamps(t *testing.T) {
	x := SampleTemplateHold("a", when)
	x.Group = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateHold("r1", when)
	rev, err := r.PutTemplateHold(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateHold() = %v", err)
	}
	got, gotRev, err := r.GetTemplateHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateHolds() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateHold(TemplateHold{}, 0); err == nil {
		t.Error("PutTemplateHold() accepted an empty entity")
	}
}

func TestTemplateClaimRoundTrip(t *testing.T) {
	x := SampleTemplateClaim("templateclaim-1", when)
	got, err := ParseTemplateClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateClaim() = %v", err)
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

func TestTemplateClaimParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateClaim(f); err == nil {
		t.Error("ParseTemplateClaim() accepted a record with no id")
	}
	f = SampleTemplateClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseTemplateClaim(f); err == nil {
		t.Error("ParseTemplateClaim() accepted a non-numeric retained")
	}
}

func TestTemplateClaimCollections(t *testing.T) {
	xs := []TemplateClaim{SampleTemplateClaim("a", when), SampleTemplateClaim("bb", when), SampleTemplateClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateClaims(xs)); got != 2 {
		t.Errorf("ActiveTemplateClaims() kept %d, want 2", got)
	}
	if got, want := SumTemplateClaims(xs), 1+3; got != want {
		t.Errorf("SumTemplateClaims() = %d, want %d", got, want)
	}
	SortTemplateClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateClaims() put %q first", xs[0].ID)
	}
	if len(IndexTemplateClaims(xs)) != 3 {
		t.Errorf("IndexTemplateClaims() = %v", IndexTemplateClaims(xs))
	}
	if err := ValidateTemplateClaims(ActiveTemplateClaims(xs)); err != nil {
		t.Errorf("ValidateTemplateClaims() = %v", err)
	}
}

func TestTemplateClaimNormalizeClamps(t *testing.T) {
	x := SampleTemplateClaim("a", when)
	x.Source = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateClaim("r1", when)
	rev, err := r.PutTemplateClaim(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateClaim() = %v", err)
	}
	got, gotRev, err := r.GetTemplateClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateClaims() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateClaim(TemplateClaim{}, 0); err == nil {
		t.Error("PutTemplateClaim() accepted an empty entity")
	}
}

func TestTemplateCursorRoundTrip(t *testing.T) {
	x := SampleTemplateCursor("templatecursor-1", when)
	got, err := ParseTemplateCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseTemplateCursor() = %v", err)
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

func TestTemplateCursorParseRejectsBadInput(t *testing.T) {
	f := SampleTemplateCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseTemplateCursor(f); err == nil {
		t.Error("ParseTemplateCursor() accepted a record with no id")
	}
	f = SampleTemplateCursor("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseTemplateCursor(f); err == nil {
		t.Error("ParseTemplateCursor() accepted a non-numeric version")
	}
}

func TestTemplateCursorCollections(t *testing.T) {
	xs := []TemplateCursor{SampleTemplateCursor("a", when), SampleTemplateCursor("bb", when), SampleTemplateCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveTemplateCursors(xs)); got != 2 {
		t.Errorf("ActiveTemplateCursors() kept %d, want 2", got)
	}
	if got, want := SumTemplateCursors(xs), 1+3; got != want {
		t.Errorf("SumTemplateCursors() = %d, want %d", got, want)
	}
	SortTemplateCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortTemplateCursors() put %q first", xs[0].ID)
	}
	if len(IndexTemplateCursors(xs)) != 3 {
		t.Errorf("IndexTemplateCursors() = %v", IndexTemplateCursors(xs))
	}
	if err := ValidateTemplateCursors(ActiveTemplateCursors(xs)); err != nil {
		t.Errorf("ValidateTemplateCursors() = %v", err)
	}
}

func TestTemplateCursorNormalizeClamps(t *testing.T) {
	x := SampleTemplateCursor("a", when)
	x.Source = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestTemplateCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleTemplateCursor("r1", when)
	rev, err := r.PutTemplateCursor(x, 0)
	if err != nil {
		t.Fatalf("PutTemplateCursor() = %v", err)
	}
	got, gotRev, err := r.GetTemplateCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetTemplateCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListTemplateCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListTemplateCursors() = %v, %v", all, err)
	}
	if _, err := r.PutTemplateCursor(TemplateCursor{}, 0); err == nil {
		t.Error("PutTemplateCursor() accepted an empty entity")
	}
}
