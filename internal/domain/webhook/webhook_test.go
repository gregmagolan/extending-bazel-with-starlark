package webhook

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
		if !strings.HasPrefix(Prefixes[i], "webhook/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleWebhookRecord("ok", when).Fields()); err != nil {
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

func TestWebhookRecordRoundTrip(t *testing.T) {
	x := SampleWebhookRecord("webhookrecord-1", when)
	got, err := ParseWebhookRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookRecord() = %v", err)
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

func TestWebhookRecordParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookRecord(f); err == nil {
		t.Error("ParseWebhookRecord() accepted a record with no id")
	}
	f = SampleWebhookRecord("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseWebhookRecord(f); err == nil {
		t.Error("ParseWebhookRecord() accepted a non-numeric count")
	}
}

func TestWebhookRecordCollections(t *testing.T) {
	xs := []WebhookRecord{SampleWebhookRecord("a", when), SampleWebhookRecord("bb", when), SampleWebhookRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookRecords(xs)); got != 2 {
		t.Errorf("ActiveWebhookRecords() kept %d, want 2", got)
	}
	if got, want := SumWebhookRecords(xs), 1+3; got != want {
		t.Errorf("SumWebhookRecords() = %d, want %d", got, want)
	}
	SortWebhookRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookRecords() put %q first", xs[0].ID)
	}
	if len(IndexWebhookRecords(xs)) != 3 {
		t.Errorf("IndexWebhookRecords() = %v", IndexWebhookRecords(xs))
	}
	if err := ValidateWebhookRecords(ActiveWebhookRecords(xs)); err != nil {
		t.Errorf("ValidateWebhookRecords() = %v", err)
	}
}

func TestWebhookRecordNormalizeClamps(t *testing.T) {
	x := SampleWebhookRecord("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookRecordRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookRecord("r1", when)
	rev, err := r.PutWebhookRecord(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookRecord() = %v", err)
	}
	got, gotRev, err := r.GetWebhookRecord(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookRecord() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookRecords()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookRecords() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookRecord(WebhookRecord{}, 0); err == nil {
		t.Error("PutWebhookRecord() accepted an empty entity")
	}
}

func TestWebhookEntryRoundTrip(t *testing.T) {
	x := SampleWebhookEntry("webhookentry-1", when)
	got, err := ParseWebhookEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookEntry() = %v", err)
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

func TestWebhookEntryParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookEntry(f); err == nil {
		t.Error("ParseWebhookEntry() accepted a record with no id")
	}
	f = SampleWebhookEntry("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseWebhookEntry(f); err == nil {
		t.Error("ParseWebhookEntry() accepted a non-numeric offset")
	}
}

func TestWebhookEntryCollections(t *testing.T) {
	xs := []WebhookEntry{SampleWebhookEntry("a", when), SampleWebhookEntry("bb", when), SampleWebhookEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookEntrys(xs)); got != 2 {
		t.Errorf("ActiveWebhookEntrys() kept %d, want 2", got)
	}
	if got, want := SumWebhookEntrys(xs), 1+3; got != want {
		t.Errorf("SumWebhookEntrys() = %d, want %d", got, want)
	}
	SortWebhookEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookEntrys() put %q first", xs[0].ID)
	}
	if len(IndexWebhookEntrys(xs)) != 3 {
		t.Errorf("IndexWebhookEntrys() = %v", IndexWebhookEntrys(xs))
	}
	if err := ValidateWebhookEntrys(ActiveWebhookEntrys(xs)); err != nil {
		t.Errorf("ValidateWebhookEntrys() = %v", err)
	}
}

func TestWebhookEntryNormalizeClamps(t *testing.T) {
	x := SampleWebhookEntry("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookEntryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookEntry("r1", when)
	rev, err := r.PutWebhookEntry(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookEntry() = %v", err)
	}
	got, gotRev, err := r.GetWebhookEntry(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookEntry() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookEntrys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookEntrys() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookEntry(WebhookEntry{}, 0); err == nil {
		t.Error("PutWebhookEntry() accepted an empty entity")
	}
}

func TestWebhookItemRoundTrip(t *testing.T) {
	x := SampleWebhookItem("webhookitem-1", when)
	got, err := ParseWebhookItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookItem() = %v", err)
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

func TestWebhookItemParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookItem(f); err == nil {
		t.Error("ParseWebhookItem() accepted a record with no id")
	}
	f = SampleWebhookItem("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseWebhookItem(f); err == nil {
		t.Error("ParseWebhookItem() accepted a non-numeric weight")
	}
}

func TestWebhookItemCollections(t *testing.T) {
	xs := []WebhookItem{SampleWebhookItem("a", when), SampleWebhookItem("bb", when), SampleWebhookItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookItems(xs)); got != 2 {
		t.Errorf("ActiveWebhookItems() kept %d, want 2", got)
	}
	if got, want := SumWebhookItems(xs), 1+3; got != want {
		t.Errorf("SumWebhookItems() = %d, want %d", got, want)
	}
	SortWebhookItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookItems() put %q first", xs[0].ID)
	}
	if len(IndexWebhookItems(xs)) != 3 {
		t.Errorf("IndexWebhookItems() = %v", IndexWebhookItems(xs))
	}
	if err := ValidateWebhookItems(ActiveWebhookItems(xs)); err != nil {
		t.Errorf("ValidateWebhookItems() = %v", err)
	}
}

func TestWebhookItemNormalizeClamps(t *testing.T) {
	x := SampleWebhookItem("a", when)
	x.Source = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookItemRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookItem("r1", when)
	rev, err := r.PutWebhookItem(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookItem() = %v", err)
	}
	got, gotRev, err := r.GetWebhookItem(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookItem() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookItems()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookItems() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookItem(WebhookItem{}, 0); err == nil {
		t.Error("PutWebhookItem() accepted an empty entity")
	}
}

func TestWebhookDraftRoundTrip(t *testing.T) {
	x := SampleWebhookDraft("webhookdraft-1", when)
	got, err := ParseWebhookDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookDraft() = %v", err)
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

func TestWebhookDraftParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookDraft(f); err == nil {
		t.Error("ParseWebhookDraft() accepted a record with no id")
	}
	f = SampleWebhookDraft("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseWebhookDraft(f); err == nil {
		t.Error("ParseWebhookDraft() accepted a non-numeric capacity")
	}
}

func TestWebhookDraftCollections(t *testing.T) {
	xs := []WebhookDraft{SampleWebhookDraft("a", when), SampleWebhookDraft("bb", when), SampleWebhookDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookDrafts(xs)); got != 2 {
		t.Errorf("ActiveWebhookDrafts() kept %d, want 2", got)
	}
	if got, want := SumWebhookDrafts(xs), 1+3; got != want {
		t.Errorf("SumWebhookDrafts() = %d, want %d", got, want)
	}
	SortWebhookDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookDrafts() put %q first", xs[0].ID)
	}
	if len(IndexWebhookDrafts(xs)) != 3 {
		t.Errorf("IndexWebhookDrafts() = %v", IndexWebhookDrafts(xs))
	}
	if err := ValidateWebhookDrafts(ActiveWebhookDrafts(xs)); err != nil {
		t.Errorf("ValidateWebhookDrafts() = %v", err)
	}
}

func TestWebhookDraftNormalizeClamps(t *testing.T) {
	x := SampleWebhookDraft("a", when)
	x.Source = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Source != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookDraftRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookDraft("r1", when)
	rev, err := r.PutWebhookDraft(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookDraft() = %v", err)
	}
	got, gotRev, err := r.GetWebhookDraft(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookDraft() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookDrafts()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookDrafts() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookDraft(WebhookDraft{}, 0); err == nil {
		t.Error("PutWebhookDraft() accepted an empty entity")
	}
}

func TestWebhookRevisionRoundTrip(t *testing.T) {
	x := SampleWebhookRevision("webhookrevision-1", when)
	got, err := ParseWebhookRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookRevision() = %v", err)
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

func TestWebhookRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookRevision(f); err == nil {
		t.Error("ParseWebhookRevision() accepted a record with no id")
	}
	f = SampleWebhookRevision("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseWebhookRevision(f); err == nil {
		t.Error("ParseWebhookRevision() accepted a non-numeric offset")
	}
}

func TestWebhookRevisionCollections(t *testing.T) {
	xs := []WebhookRevision{SampleWebhookRevision("a", when), SampleWebhookRevision("bb", when), SampleWebhookRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookRevisions(xs)); got != 2 {
		t.Errorf("ActiveWebhookRevisions() kept %d, want 2", got)
	}
	if got, want := SumWebhookRevisions(xs), 1+3; got != want {
		t.Errorf("SumWebhookRevisions() = %d, want %d", got, want)
	}
	SortWebhookRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookRevisions() put %q first", xs[0].ID)
	}
	if len(IndexWebhookRevisions(xs)) != 3 {
		t.Errorf("IndexWebhookRevisions() = %v", IndexWebhookRevisions(xs))
	}
	if err := ValidateWebhookRevisions(ActiveWebhookRevisions(xs)); err != nil {
		t.Errorf("ValidateWebhookRevisions() = %v", err)
	}
}

func TestWebhookRevisionNormalizeClamps(t *testing.T) {
	x := SampleWebhookRevision("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookRevisionRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookRevision("r1", when)
	rev, err := r.PutWebhookRevision(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookRevision() = %v", err)
	}
	got, gotRev, err := r.GetWebhookRevision(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookRevision() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookRevisions()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookRevisions() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookRevision(WebhookRevision{}, 0); err == nil {
		t.Error("PutWebhookRevision() accepted an empty entity")
	}
}

func TestWebhookSnapshotRoundTrip(t *testing.T) {
	x := SampleWebhookSnapshot("webhooksnapshot-1", when)
	got, err := ParseWebhookSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookSnapshot() = %v", err)
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

func TestWebhookSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookSnapshot(f); err == nil {
		t.Error("ParseWebhookSnapshot() accepted a record with no id")
	}
	f = SampleWebhookSnapshot("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseWebhookSnapshot(f); err == nil {
		t.Error("ParseWebhookSnapshot() accepted a non-numeric weight")
	}
}

func TestWebhookSnapshotCollections(t *testing.T) {
	xs := []WebhookSnapshot{SampleWebhookSnapshot("a", when), SampleWebhookSnapshot("bb", when), SampleWebhookSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookSnapshots(xs)); got != 2 {
		t.Errorf("ActiveWebhookSnapshots() kept %d, want 2", got)
	}
	if got, want := SumWebhookSnapshots(xs), 1+3; got != want {
		t.Errorf("SumWebhookSnapshots() = %d, want %d", got, want)
	}
	SortWebhookSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexWebhookSnapshots(xs)) != 3 {
		t.Errorf("IndexWebhookSnapshots() = %v", IndexWebhookSnapshots(xs))
	}
	if err := ValidateWebhookSnapshots(ActiveWebhookSnapshots(xs)); err != nil {
		t.Errorf("ValidateWebhookSnapshots() = %v", err)
	}
}

func TestWebhookSnapshotNormalizeClamps(t *testing.T) {
	x := SampleWebhookSnapshot("a", when)
	x.Name = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookSnapshotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookSnapshot("r1", when)
	rev, err := r.PutWebhookSnapshot(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookSnapshot() = %v", err)
	}
	got, gotRev, err := r.GetWebhookSnapshot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookSnapshot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookSnapshots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookSnapshots() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookSnapshot(WebhookSnapshot{}, 0); err == nil {
		t.Error("PutWebhookSnapshot() accepted an empty entity")
	}
}

func TestWebhookBatchRoundTrip(t *testing.T) {
	x := SampleWebhookBatch("webhookbatch-1", when)
	got, err := ParseWebhookBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookBatch() = %v", err)
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

func TestWebhookBatchParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookBatch(f); err == nil {
		t.Error("ParseWebhookBatch() accepted a record with no id")
	}
	f = SampleWebhookBatch("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseWebhookBatch(f); err == nil {
		t.Error("ParseWebhookBatch() accepted a non-numeric version")
	}
}

func TestWebhookBatchCollections(t *testing.T) {
	xs := []WebhookBatch{SampleWebhookBatch("a", when), SampleWebhookBatch("bb", when), SampleWebhookBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookBatchs(xs)); got != 2 {
		t.Errorf("ActiveWebhookBatchs() kept %d, want 2", got)
	}
	if got, want := SumWebhookBatchs(xs), 1+3; got != want {
		t.Errorf("SumWebhookBatchs() = %d, want %d", got, want)
	}
	SortWebhookBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookBatchs() put %q first", xs[0].ID)
	}
	if len(IndexWebhookBatchs(xs)) != 3 {
		t.Errorf("IndexWebhookBatchs() = %v", IndexWebhookBatchs(xs))
	}
	if err := ValidateWebhookBatchs(ActiveWebhookBatchs(xs)); err != nil {
		t.Errorf("ValidateWebhookBatchs() = %v", err)
	}
}

func TestWebhookBatchNormalizeClamps(t *testing.T) {
	x := SampleWebhookBatch("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookBatchRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookBatch("r1", when)
	rev, err := r.PutWebhookBatch(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookBatch() = %v", err)
	}
	got, gotRev, err := r.GetWebhookBatch(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookBatch() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookBatchs()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookBatchs() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookBatch(WebhookBatch{}, 0); err == nil {
		t.Error("PutWebhookBatch() accepted an empty entity")
	}
}

func TestWebhookRequestRoundTrip(t *testing.T) {
	x := SampleWebhookRequest("webhookrequest-1", when)
	got, err := ParseWebhookRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookRequest() = %v", err)
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

func TestWebhookRequestParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookRequest(f); err == nil {
		t.Error("ParseWebhookRequest() accepted a record with no id")
	}
	f = SampleWebhookRequest("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseWebhookRequest(f); err == nil {
		t.Error("ParseWebhookRequest() accepted a non-numeric offset")
	}
}

func TestWebhookRequestCollections(t *testing.T) {
	xs := []WebhookRequest{SampleWebhookRequest("a", when), SampleWebhookRequest("bb", when), SampleWebhookRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookRequests(xs)); got != 2 {
		t.Errorf("ActiveWebhookRequests() kept %d, want 2", got)
	}
	if got, want := SumWebhookRequests(xs), 1+3; got != want {
		t.Errorf("SumWebhookRequests() = %d, want %d", got, want)
	}
	SortWebhookRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookRequests() put %q first", xs[0].ID)
	}
	if len(IndexWebhookRequests(xs)) != 3 {
		t.Errorf("IndexWebhookRequests() = %v", IndexWebhookRequests(xs))
	}
	if err := ValidateWebhookRequests(ActiveWebhookRequests(xs)); err != nil {
		t.Errorf("ValidateWebhookRequests() = %v", err)
	}
}

func TestWebhookRequestNormalizeClamps(t *testing.T) {
	x := SampleWebhookRequest("a", when)
	x.Group = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookRequestRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookRequest("r1", when)
	rev, err := r.PutWebhookRequest(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookRequest() = %v", err)
	}
	got, gotRev, err := r.GetWebhookRequest(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookRequest() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookRequests()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookRequests() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookRequest(WebhookRequest{}, 0); err == nil {
		t.Error("PutWebhookRequest() accepted an empty entity")
	}
}

func TestWebhookSummaryRoundTrip(t *testing.T) {
	x := SampleWebhookSummary("webhooksummary-1", when)
	got, err := ParseWebhookSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookSummary() = %v", err)
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

func TestWebhookSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookSummary(f); err == nil {
		t.Error("ParseWebhookSummary() accepted a record with no id")
	}
	f = SampleWebhookSummary("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseWebhookSummary(f); err == nil {
		t.Error("ParseWebhookSummary() accepted a non-numeric count")
	}
}

func TestWebhookSummaryCollections(t *testing.T) {
	xs := []WebhookSummary{SampleWebhookSummary("a", when), SampleWebhookSummary("bb", when), SampleWebhookSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookSummarys(xs)); got != 2 {
		t.Errorf("ActiveWebhookSummarys() kept %d, want 2", got)
	}
	if got, want := SumWebhookSummarys(xs), 1+3; got != want {
		t.Errorf("SumWebhookSummarys() = %d, want %d", got, want)
	}
	SortWebhookSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookSummarys() put %q first", xs[0].ID)
	}
	if len(IndexWebhookSummarys(xs)) != 3 {
		t.Errorf("IndexWebhookSummarys() = %v", IndexWebhookSummarys(xs))
	}
	if err := ValidateWebhookSummarys(ActiveWebhookSummarys(xs)); err != nil {
		t.Errorf("ValidateWebhookSummarys() = %v", err)
	}
}

func TestWebhookSummaryNormalizeClamps(t *testing.T) {
	x := SampleWebhookSummary("a", when)
	x.Category = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookSummaryRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookSummary("r1", when)
	rev, err := r.PutWebhookSummary(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookSummary() = %v", err)
	}
	got, gotRev, err := r.GetWebhookSummary(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookSummary() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookSummarys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookSummarys() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookSummary(WebhookSummary{}, 0); err == nil {
		t.Error("PutWebhookSummary() accepted an empty entity")
	}
}

func TestWebhookPolicyRoundTrip(t *testing.T) {
	x := SampleWebhookPolicy("webhookpolicy-1", when)
	got, err := ParseWebhookPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookPolicy() = %v", err)
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

func TestWebhookPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookPolicy(f); err == nil {
		t.Error("ParseWebhookPolicy() accepted a record with no id")
	}
	f = SampleWebhookPolicy("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseWebhookPolicy(f); err == nil {
		t.Error("ParseWebhookPolicy() accepted a non-numeric retained")
	}
}

func TestWebhookPolicyCollections(t *testing.T) {
	xs := []WebhookPolicy{SampleWebhookPolicy("a", when), SampleWebhookPolicy("bb", when), SampleWebhookPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookPolicys(xs)); got != 2 {
		t.Errorf("ActiveWebhookPolicys() kept %d, want 2", got)
	}
	if got, want := SumWebhookPolicys(xs), 1+3; got != want {
		t.Errorf("SumWebhookPolicys() = %d, want %d", got, want)
	}
	SortWebhookPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookPolicys() put %q first", xs[0].ID)
	}
	if len(IndexWebhookPolicys(xs)) != 3 {
		t.Errorf("IndexWebhookPolicys() = %v", IndexWebhookPolicys(xs))
	}
	if err := ValidateWebhookPolicys(ActiveWebhookPolicys(xs)); err != nil {
		t.Errorf("ValidateWebhookPolicys() = %v", err)
	}
}

func TestWebhookPolicyNormalizeClamps(t *testing.T) {
	x := SampleWebhookPolicy("a", when)
	x.Group = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookPolicyRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookPolicy("r1", when)
	rev, err := r.PutWebhookPolicy(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookPolicy() = %v", err)
	}
	got, gotRev, err := r.GetWebhookPolicy(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookPolicy() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookPolicys()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookPolicys() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookPolicy(WebhookPolicy{}, 0); err == nil {
		t.Error("PutWebhookPolicy() accepted an empty entity")
	}
}

func TestWebhookLinkRoundTrip(t *testing.T) {
	x := SampleWebhookLink("webhooklink-1", when)
	got, err := ParseWebhookLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookLink() = %v", err)
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

func TestWebhookLinkParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookLink(f); err == nil {
		t.Error("ParseWebhookLink() accepted a record with no id")
	}
	f = SampleWebhookLink("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseWebhookLink(f); err == nil {
		t.Error("ParseWebhookLink() accepted a non-numeric capacity")
	}
}

func TestWebhookLinkCollections(t *testing.T) {
	xs := []WebhookLink{SampleWebhookLink("a", when), SampleWebhookLink("bb", when), SampleWebhookLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookLinks(xs)); got != 2 {
		t.Errorf("ActiveWebhookLinks() kept %d, want 2", got)
	}
	if got, want := SumWebhookLinks(xs), 1+3; got != want {
		t.Errorf("SumWebhookLinks() = %d, want %d", got, want)
	}
	SortWebhookLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookLinks() put %q first", xs[0].ID)
	}
	if len(IndexWebhookLinks(xs)) != 3 {
		t.Errorf("IndexWebhookLinks() = %v", IndexWebhookLinks(xs))
	}
	if err := ValidateWebhookLinks(ActiveWebhookLinks(xs)); err != nil {
		t.Errorf("ValidateWebhookLinks() = %v", err)
	}
}

func TestWebhookLinkNormalizeClamps(t *testing.T) {
	x := SampleWebhookLink("a", when)
	x.Group = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookLinkRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookLink("r1", when)
	rev, err := r.PutWebhookLink(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookLink() = %v", err)
	}
	got, gotRev, err := r.GetWebhookLink(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookLink() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookLinks()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookLinks() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookLink(WebhookLink{}, 0); err == nil {
		t.Error("PutWebhookLink() accepted an empty entity")
	}
}

func TestWebhookNoteRoundTrip(t *testing.T) {
	x := SampleWebhookNote("webhooknote-1", when)
	got, err := ParseWebhookNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookNote() = %v", err)
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

func TestWebhookNoteParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookNote(f); err == nil {
		t.Error("ParseWebhookNote() accepted a record with no id")
	}
	f = SampleWebhookNote("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseWebhookNote(f); err == nil {
		t.Error("ParseWebhookNote() accepted a non-numeric offset")
	}
}

func TestWebhookNoteCollections(t *testing.T) {
	xs := []WebhookNote{SampleWebhookNote("a", when), SampleWebhookNote("bb", when), SampleWebhookNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookNotes(xs)); got != 2 {
		t.Errorf("ActiveWebhookNotes() kept %d, want 2", got)
	}
	if got, want := SumWebhookNotes(xs), 1+3; got != want {
		t.Errorf("SumWebhookNotes() = %d, want %d", got, want)
	}
	SortWebhookNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookNotes() put %q first", xs[0].ID)
	}
	if len(IndexWebhookNotes(xs)) != 3 {
		t.Errorf("IndexWebhookNotes() = %v", IndexWebhookNotes(xs))
	}
	if err := ValidateWebhookNotes(ActiveWebhookNotes(xs)); err != nil {
		t.Errorf("ValidateWebhookNotes() = %v", err)
	}
}

func TestWebhookNoteNormalizeClamps(t *testing.T) {
	x := SampleWebhookNote("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookNoteRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookNote("r1", when)
	rev, err := r.PutWebhookNote(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookNote() = %v", err)
	}
	got, gotRev, err := r.GetWebhookNote(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookNote() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookNotes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookNotes() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookNote(WebhookNote{}, 0); err == nil {
		t.Error("PutWebhookNote() accepted an empty entity")
	}
}

func TestWebhookWindowRoundTrip(t *testing.T) {
	x := SampleWebhookWindow("webhookwindow-1", when)
	got, err := ParseWebhookWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookWindow() = %v", err)
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

func TestWebhookWindowParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookWindow(f); err == nil {
		t.Error("ParseWebhookWindow() accepted a record with no id")
	}
	f = SampleWebhookWindow("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseWebhookWindow(f); err == nil {
		t.Error("ParseWebhookWindow() accepted a non-numeric count")
	}
}

func TestWebhookWindowCollections(t *testing.T) {
	xs := []WebhookWindow{SampleWebhookWindow("a", when), SampleWebhookWindow("bb", when), SampleWebhookWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookWindows(xs)); got != 2 {
		t.Errorf("ActiveWebhookWindows() kept %d, want 2", got)
	}
	if got, want := SumWebhookWindows(xs), 1+3; got != want {
		t.Errorf("SumWebhookWindows() = %d, want %d", got, want)
	}
	SortWebhookWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookWindows() put %q first", xs[0].ID)
	}
	if len(IndexWebhookWindows(xs)) != 3 {
		t.Errorf("IndexWebhookWindows() = %v", IndexWebhookWindows(xs))
	}
	if err := ValidateWebhookWindows(ActiveWebhookWindows(xs)); err != nil {
		t.Errorf("ValidateWebhookWindows() = %v", err)
	}
}

func TestWebhookWindowNormalizeClamps(t *testing.T) {
	x := SampleWebhookWindow("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookWindowRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookWindow("r1", when)
	rev, err := r.PutWebhookWindow(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookWindow() = %v", err)
	}
	got, gotRev, err := r.GetWebhookWindow(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookWindow() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookWindows()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookWindows() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookWindow(WebhookWindow{}, 0); err == nil {
		t.Error("PutWebhookWindow() accepted an empty entity")
	}
}

func TestWebhookQuotaRoundTrip(t *testing.T) {
	x := SampleWebhookQuota("webhookquota-1", when)
	got, err := ParseWebhookQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookQuota() = %v", err)
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

func TestWebhookQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookQuota(f); err == nil {
		t.Error("ParseWebhookQuota() accepted a record with no id")
	}
	f = SampleWebhookQuota("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseWebhookQuota(f); err == nil {
		t.Error("ParseWebhookQuota() accepted a non-numeric capacity")
	}
}

func TestWebhookQuotaCollections(t *testing.T) {
	xs := []WebhookQuota{SampleWebhookQuota("a", when), SampleWebhookQuota("bb", when), SampleWebhookQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookQuotas(xs)); got != 2 {
		t.Errorf("ActiveWebhookQuotas() kept %d, want 2", got)
	}
	if got, want := SumWebhookQuotas(xs), 1+3; got != want {
		t.Errorf("SumWebhookQuotas() = %d, want %d", got, want)
	}
	SortWebhookQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookQuotas() put %q first", xs[0].ID)
	}
	if len(IndexWebhookQuotas(xs)) != 3 {
		t.Errorf("IndexWebhookQuotas() = %v", IndexWebhookQuotas(xs))
	}
	if err := ValidateWebhookQuotas(ActiveWebhookQuotas(xs)); err != nil {
		t.Errorf("ValidateWebhookQuotas() = %v", err)
	}
}

func TestWebhookQuotaNormalizeClamps(t *testing.T) {
	x := SampleWebhookQuota("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookQuotaRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookQuota("r1", when)
	rev, err := r.PutWebhookQuota(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookQuota() = %v", err)
	}
	got, gotRev, err := r.GetWebhookQuota(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookQuota() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookQuotas()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookQuotas() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookQuota(WebhookQuota{}, 0); err == nil {
		t.Error("PutWebhookQuota() accepted an empty entity")
	}
}

func TestWebhookBucketRoundTrip(t *testing.T) {
	x := SampleWebhookBucket("webhookbucket-1", when)
	got, err := ParseWebhookBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookBucket() = %v", err)
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

func TestWebhookBucketParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookBucket(f); err == nil {
		t.Error("ParseWebhookBucket() accepted a record with no id")
	}
	f = SampleWebhookBucket("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWebhookBucket(f); err == nil {
		t.Error("ParseWebhookBucket() accepted a non-numeric priority")
	}
}

func TestWebhookBucketCollections(t *testing.T) {
	xs := []WebhookBucket{SampleWebhookBucket("a", when), SampleWebhookBucket("bb", when), SampleWebhookBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookBuckets(xs)); got != 2 {
		t.Errorf("ActiveWebhookBuckets() kept %d, want 2", got)
	}
	if got, want := SumWebhookBuckets(xs), 1+3; got != want {
		t.Errorf("SumWebhookBuckets() = %d, want %d", got, want)
	}
	SortWebhookBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookBuckets() put %q first", xs[0].ID)
	}
	if len(IndexWebhookBuckets(xs)) != 3 {
		t.Errorf("IndexWebhookBuckets() = %v", IndexWebhookBuckets(xs))
	}
	if err := ValidateWebhookBuckets(ActiveWebhookBuckets(xs)); err != nil {
		t.Errorf("ValidateWebhookBuckets() = %v", err)
	}
}

func TestWebhookBucketNormalizeClamps(t *testing.T) {
	x := SampleWebhookBucket("a", when)
	x.Name = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookBucketRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookBucket("r1", when)
	rev, err := r.PutWebhookBucket(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookBucket() = %v", err)
	}
	got, gotRev, err := r.GetWebhookBucket(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookBucket() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookBuckets()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookBuckets() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookBucket(WebhookBucket{}, 0); err == nil {
		t.Error("PutWebhookBucket() accepted an empty entity")
	}
}

func TestWebhookLaneRoundTrip(t *testing.T) {
	x := SampleWebhookLane("webhooklane-1", when)
	got, err := ParseWebhookLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookLane() = %v", err)
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

func TestWebhookLaneParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookLane(f); err == nil {
		t.Error("ParseWebhookLane() accepted a record with no id")
	}
	f = SampleWebhookLane("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseWebhookLane(f); err == nil {
		t.Error("ParseWebhookLane() accepted a non-numeric weight")
	}
}

func TestWebhookLaneCollections(t *testing.T) {
	xs := []WebhookLane{SampleWebhookLane("a", when), SampleWebhookLane("bb", when), SampleWebhookLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookLanes(xs)); got != 2 {
		t.Errorf("ActiveWebhookLanes() kept %d, want 2", got)
	}
	if got, want := SumWebhookLanes(xs), 1+3; got != want {
		t.Errorf("SumWebhookLanes() = %d, want %d", got, want)
	}
	SortWebhookLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookLanes() put %q first", xs[0].ID)
	}
	if len(IndexWebhookLanes(xs)) != 3 {
		t.Errorf("IndexWebhookLanes() = %v", IndexWebhookLanes(xs))
	}
	if err := ValidateWebhookLanes(ActiveWebhookLanes(xs)); err != nil {
		t.Errorf("ValidateWebhookLanes() = %v", err)
	}
}

func TestWebhookLaneNormalizeClamps(t *testing.T) {
	x := SampleWebhookLane("a", when)
	x.Group = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookLaneRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookLane("r1", when)
	rev, err := r.PutWebhookLane(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookLane() = %v", err)
	}
	got, gotRev, err := r.GetWebhookLane(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookLane() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookLanes()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookLanes() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookLane(WebhookLane{}, 0); err == nil {
		t.Error("PutWebhookLane() accepted an empty entity")
	}
}

func TestWebhookSlotRoundTrip(t *testing.T) {
	x := SampleWebhookSlot("webhookslot-1", when)
	got, err := ParseWebhookSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookSlot() = %v", err)
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

func TestWebhookSlotParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookSlot(f); err == nil {
		t.Error("ParseWebhookSlot() accepted a record with no id")
	}
	f = SampleWebhookSlot("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseWebhookSlot(f); err == nil {
		t.Error("ParseWebhookSlot() accepted a non-numeric version")
	}
}

func TestWebhookSlotCollections(t *testing.T) {
	xs := []WebhookSlot{SampleWebhookSlot("a", when), SampleWebhookSlot("bb", when), SampleWebhookSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookSlots(xs)); got != 2 {
		t.Errorf("ActiveWebhookSlots() kept %d, want 2", got)
	}
	if got, want := SumWebhookSlots(xs), 1+3; got != want {
		t.Errorf("SumWebhookSlots() = %d, want %d", got, want)
	}
	SortWebhookSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookSlots() put %q first", xs[0].ID)
	}
	if len(IndexWebhookSlots(xs)) != 3 {
		t.Errorf("IndexWebhookSlots() = %v", IndexWebhookSlots(xs))
	}
	if err := ValidateWebhookSlots(ActiveWebhookSlots(xs)); err != nil {
		t.Errorf("ValidateWebhookSlots() = %v", err)
	}
}

func TestWebhookSlotNormalizeClamps(t *testing.T) {
	x := SampleWebhookSlot("a", when)
	x.Comment = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookSlotRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookSlot("r1", when)
	rev, err := r.PutWebhookSlot(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookSlot() = %v", err)
	}
	got, gotRev, err := r.GetWebhookSlot(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookSlot() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookSlots()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookSlots() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookSlot(WebhookSlot{}, 0); err == nil {
		t.Error("PutWebhookSlot() accepted an empty entity")
	}
}

func TestWebhookTagRoundTrip(t *testing.T) {
	x := SampleWebhookTag("webhooktag-1", when)
	got, err := ParseWebhookTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookTag() = %v", err)
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

func TestWebhookTagParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookTag(f); err == nil {
		t.Error("ParseWebhookTag() accepted a record with no id")
	}
	f = SampleWebhookTag("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseWebhookTag(f); err == nil {
		t.Error("ParseWebhookTag() accepted a non-numeric retained")
	}
}

func TestWebhookTagCollections(t *testing.T) {
	xs := []WebhookTag{SampleWebhookTag("a", when), SampleWebhookTag("bb", when), SampleWebhookTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookTags(xs)); got != 2 {
		t.Errorf("ActiveWebhookTags() kept %d, want 2", got)
	}
	if got, want := SumWebhookTags(xs), 1+3; got != want {
		t.Errorf("SumWebhookTags() = %d, want %d", got, want)
	}
	SortWebhookTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookTags() put %q first", xs[0].ID)
	}
	if len(IndexWebhookTags(xs)) != 3 {
		t.Errorf("IndexWebhookTags() = %v", IndexWebhookTags(xs))
	}
	if err := ValidateWebhookTags(ActiveWebhookTags(xs)); err != nil {
		t.Errorf("ValidateWebhookTags() = %v", err)
	}
}

func TestWebhookTagNormalizeClamps(t *testing.T) {
	x := SampleWebhookTag("a", when)
	x.Category = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookTagRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookTag("r1", when)
	rev, err := r.PutWebhookTag(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookTag() = %v", err)
	}
	got, gotRev, err := r.GetWebhookTag(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookTag() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookTags()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookTags() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookTag(WebhookTag{}, 0); err == nil {
		t.Error("PutWebhookTag() accepted an empty entity")
	}
}

func TestWebhookRuleRoundTrip(t *testing.T) {
	x := SampleWebhookRule("webhookrule-1", when)
	got, err := ParseWebhookRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookRule() = %v", err)
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

func TestWebhookRuleParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookRule(f); err == nil {
		t.Error("ParseWebhookRule() accepted a record with no id")
	}
	f = SampleWebhookRule("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWebhookRule(f); err == nil {
		t.Error("ParseWebhookRule() accepted a non-numeric priority")
	}
}

func TestWebhookRuleCollections(t *testing.T) {
	xs := []WebhookRule{SampleWebhookRule("a", when), SampleWebhookRule("bb", when), SampleWebhookRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookRules(xs)); got != 2 {
		t.Errorf("ActiveWebhookRules() kept %d, want 2", got)
	}
	if got, want := SumWebhookRules(xs), 1+3; got != want {
		t.Errorf("SumWebhookRules() = %d, want %d", got, want)
	}
	SortWebhookRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookRules() put %q first", xs[0].ID)
	}
	if len(IndexWebhookRules(xs)) != 3 {
		t.Errorf("IndexWebhookRules() = %v", IndexWebhookRules(xs))
	}
	if err := ValidateWebhookRules(ActiveWebhookRules(xs)); err != nil {
		t.Errorf("ValidateWebhookRules() = %v", err)
	}
}

func TestWebhookRuleNormalizeClamps(t *testing.T) {
	x := SampleWebhookRule("a", when)
	x.Group = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookRuleRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookRule("r1", when)
	rev, err := r.PutWebhookRule(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookRule() = %v", err)
	}
	got, gotRev, err := r.GetWebhookRule(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookRule() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookRules()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookRules() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookRule(WebhookRule{}, 0); err == nil {
		t.Error("PutWebhookRule() accepted an empty entity")
	}
}

func TestWebhookHoldRoundTrip(t *testing.T) {
	x := SampleWebhookHold("webhookhold-1", when)
	got, err := ParseWebhookHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookHold() = %v", err)
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

func TestWebhookHoldParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookHold(f); err == nil {
		t.Error("ParseWebhookHold() accepted a record with no id")
	}
	f = SampleWebhookHold("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseWebhookHold(f); err == nil {
		t.Error("ParseWebhookHold() accepted a non-numeric priority")
	}
}

func TestWebhookHoldCollections(t *testing.T) {
	xs := []WebhookHold{SampleWebhookHold("a", when), SampleWebhookHold("bb", when), SampleWebhookHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookHolds(xs)); got != 2 {
		t.Errorf("ActiveWebhookHolds() kept %d, want 2", got)
	}
	if got, want := SumWebhookHolds(xs), 1+3; got != want {
		t.Errorf("SumWebhookHolds() = %d, want %d", got, want)
	}
	SortWebhookHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookHolds() put %q first", xs[0].ID)
	}
	if len(IndexWebhookHolds(xs)) != 3 {
		t.Errorf("IndexWebhookHolds() = %v", IndexWebhookHolds(xs))
	}
	if err := ValidateWebhookHolds(ActiveWebhookHolds(xs)); err != nil {
		t.Errorf("ValidateWebhookHolds() = %v", err)
	}
}

func TestWebhookHoldNormalizeClamps(t *testing.T) {
	x := SampleWebhookHold("a", when)
	x.Comment = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookHoldRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookHold("r1", when)
	rev, err := r.PutWebhookHold(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookHold() = %v", err)
	}
	got, gotRev, err := r.GetWebhookHold(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookHold() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookHolds()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookHolds() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookHold(WebhookHold{}, 0); err == nil {
		t.Error("PutWebhookHold() accepted an empty entity")
	}
}

func TestWebhookClaimRoundTrip(t *testing.T) {
	x := SampleWebhookClaim("webhookclaim-1", when)
	got, err := ParseWebhookClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookClaim() = %v", err)
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

func TestWebhookClaimParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookClaim(f); err == nil {
		t.Error("ParseWebhookClaim() accepted a record with no id")
	}
	f = SampleWebhookClaim("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseWebhookClaim(f); err == nil {
		t.Error("ParseWebhookClaim() accepted a non-numeric retained")
	}
}

func TestWebhookClaimCollections(t *testing.T) {
	xs := []WebhookClaim{SampleWebhookClaim("a", when), SampleWebhookClaim("bb", when), SampleWebhookClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookClaims(xs)); got != 2 {
		t.Errorf("ActiveWebhookClaims() kept %d, want 2", got)
	}
	if got, want := SumWebhookClaims(xs), 1+3; got != want {
		t.Errorf("SumWebhookClaims() = %d, want %d", got, want)
	}
	SortWebhookClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookClaims() put %q first", xs[0].ID)
	}
	if len(IndexWebhookClaims(xs)) != 3 {
		t.Errorf("IndexWebhookClaims() = %v", IndexWebhookClaims(xs))
	}
	if err := ValidateWebhookClaims(ActiveWebhookClaims(xs)); err != nil {
		t.Errorf("ValidateWebhookClaims() = %v", err)
	}
}

func TestWebhookClaimNormalizeClamps(t *testing.T) {
	x := SampleWebhookClaim("a", when)
	x.Reference = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookClaimRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookClaim("r1", when)
	rev, err := r.PutWebhookClaim(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookClaim() = %v", err)
	}
	got, gotRev, err := r.GetWebhookClaim(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookClaim() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookClaims()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookClaims() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookClaim(WebhookClaim{}, 0); err == nil {
		t.Error("PutWebhookClaim() accepted an empty entity")
	}
}

func TestWebhookCursorRoundTrip(t *testing.T) {
	x := SampleWebhookCursor("webhookcursor-1", when)
	got, err := ParseWebhookCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseWebhookCursor() = %v", err)
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

func TestWebhookCursorParseRejectsBadInput(t *testing.T) {
	f := SampleWebhookCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseWebhookCursor(f); err == nil {
		t.Error("ParseWebhookCursor() accepted a record with no id")
	}
	f = SampleWebhookCursor("x", when).Fields()
	f["retained"] = "not-a-number"
	if _, err := ParseWebhookCursor(f); err == nil {
		t.Error("ParseWebhookCursor() accepted a non-numeric retained")
	}
}

func TestWebhookCursorCollections(t *testing.T) {
	xs := []WebhookCursor{SampleWebhookCursor("a", when), SampleWebhookCursor("bb", when), SampleWebhookCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveWebhookCursors(xs)); got != 2 {
		t.Errorf("ActiveWebhookCursors() kept %d, want 2", got)
	}
	if got, want := SumWebhookCursors(xs), 1+3; got != want {
		t.Errorf("SumWebhookCursors() = %d, want %d", got, want)
	}
	SortWebhookCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortWebhookCursors() put %q first", xs[0].ID)
	}
	if len(IndexWebhookCursors(xs)) != 3 {
		t.Errorf("IndexWebhookCursors() = %v", IndexWebhookCursors(xs))
	}
	if err := ValidateWebhookCursors(ActiveWebhookCursors(xs)); err != nil {
		t.Errorf("ValidateWebhookCursors() = %v", err)
	}
}

func TestWebhookCursorNormalizeClamps(t *testing.T) {
	x := SampleWebhookCursor("a", when)
	x.Region = "  padded  "
	x.Retained = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Retained != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestWebhookCursorRepo(t *testing.T) {
	r := newRepo(t)
	x := SampleWebhookCursor("r1", when)
	rev, err := r.PutWebhookCursor(x, 0)
	if err != nil {
		t.Fatalf("PutWebhookCursor() = %v", err)
	}
	got, gotRev, err := r.GetWebhookCursor(x.ID)
	if err != nil || gotRev != rev || got.ID != x.ID {
		t.Fatalf("GetWebhookCursor() = %+v, %d, %v", got, gotRev, err)
	}
	all, err := r.ListWebhookCursors()
	if err != nil || len(all) != 1 {
		t.Fatalf("ListWebhookCursors() = %v, %v", all, err)
	}
	if _, err := r.PutWebhookCursor(WebhookCursor{}, 0); err == nil {
		t.Error("PutWebhookCursor() accepted an empty entity")
	}
}
