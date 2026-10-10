package user

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
		if !strings.HasPrefix(Prefixes[i], "user/") {
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
	if err := NewValidator(true).Check(Schemas[0].Name, SampleUserRecord("ok", when).Fields()); err != nil {
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

func TestUserRecordRoundTrip(t *testing.T) {
	x := SampleUserRecord("userrecord-1", when)
	got, err := ParseUserRecord(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserRecord() = %v", err)
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

func TestUserRecordParseRejectsBadInput(t *testing.T) {
	f := SampleUserRecord("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserRecord(f); err == nil {
		t.Error("ParseUserRecord() accepted a record with no id")
	}
	f = SampleUserRecord("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseUserRecord(f); err == nil {
		t.Error("ParseUserRecord() accepted a non-numeric capacity")
	}
}

func TestUserRecordCollections(t *testing.T) {
	xs := []UserRecord{SampleUserRecord("a", when), SampleUserRecord("bb", when), SampleUserRecord("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserRecords(xs)); got != 2 {
		t.Errorf("ActiveUserRecords() kept %d, want 2", got)
	}
	if got, want := SumUserRecords(xs), 1+3; got != want {
		t.Errorf("SumUserRecords() = %d, want %d", got, want)
	}
	SortUserRecords(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserRecords() put %q first", xs[0].ID)
	}
	if len(IndexUserRecords(xs)) != 3 {
		t.Errorf("IndexUserRecords() = %v", IndexUserRecords(xs))
	}
	if err := ValidateUserRecords(ActiveUserRecords(xs)); err != nil {
		t.Errorf("ValidateUserRecords() = %v", err)
	}
}

func TestUserRecordNormalizeClamps(t *testing.T) {
	x := SampleUserRecord("a", when)
	x.Channel = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserEntryRoundTrip(t *testing.T) {
	x := SampleUserEntry("userentry-1", when)
	got, err := ParseUserEntry(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserEntry() = %v", err)
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

func TestUserEntryParseRejectsBadInput(t *testing.T) {
	f := SampleUserEntry("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserEntry(f); err == nil {
		t.Error("ParseUserEntry() accepted a record with no id")
	}
	f = SampleUserEntry("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseUserEntry(f); err == nil {
		t.Error("ParseUserEntry() accepted a non-numeric version")
	}
}

func TestUserEntryCollections(t *testing.T) {
	xs := []UserEntry{SampleUserEntry("a", when), SampleUserEntry("bb", when), SampleUserEntry("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserEntrys(xs)); got != 2 {
		t.Errorf("ActiveUserEntrys() kept %d, want 2", got)
	}
	if got, want := SumUserEntrys(xs), 1+3; got != want {
		t.Errorf("SumUserEntrys() = %d, want %d", got, want)
	}
	SortUserEntrys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserEntrys() put %q first", xs[0].ID)
	}
	if len(IndexUserEntrys(xs)) != 3 {
		t.Errorf("IndexUserEntrys() = %v", IndexUserEntrys(xs))
	}
	if err := ValidateUserEntrys(ActiveUserEntrys(xs)); err != nil {
		t.Errorf("ValidateUserEntrys() = %v", err)
	}
}

func TestUserEntryNormalizeClamps(t *testing.T) {
	x := SampleUserEntry("a", when)
	x.Region = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserItemRoundTrip(t *testing.T) {
	x := SampleUserItem("useritem-1", when)
	got, err := ParseUserItem(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserItem() = %v", err)
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

func TestUserItemParseRejectsBadInput(t *testing.T) {
	f := SampleUserItem("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserItem(f); err == nil {
		t.Error("ParseUserItem() accepted a record with no id")
	}
	f = SampleUserItem("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseUserItem(f); err == nil {
		t.Error("ParseUserItem() accepted a non-numeric version")
	}
}

func TestUserItemCollections(t *testing.T) {
	xs := []UserItem{SampleUserItem("a", when), SampleUserItem("bb", when), SampleUserItem("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserItems(xs)); got != 2 {
		t.Errorf("ActiveUserItems() kept %d, want 2", got)
	}
	if got, want := SumUserItems(xs), 1+3; got != want {
		t.Errorf("SumUserItems() = %d, want %d", got, want)
	}
	SortUserItems(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserItems() put %q first", xs[0].ID)
	}
	if len(IndexUserItems(xs)) != 3 {
		t.Errorf("IndexUserItems() = %v", IndexUserItems(xs))
	}
	if err := ValidateUserItems(ActiveUserItems(xs)); err != nil {
		t.Errorf("ValidateUserItems() = %v", err)
	}
}

func TestUserItemNormalizeClamps(t *testing.T) {
	x := SampleUserItem("a", when)
	x.Reference = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserDraftRoundTrip(t *testing.T) {
	x := SampleUserDraft("userdraft-1", when)
	got, err := ParseUserDraft(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserDraft() = %v", err)
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

func TestUserDraftParseRejectsBadInput(t *testing.T) {
	f := SampleUserDraft("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserDraft(f); err == nil {
		t.Error("ParseUserDraft() accepted a record with no id")
	}
	f = SampleUserDraft("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseUserDraft(f); err == nil {
		t.Error("ParseUserDraft() accepted a non-numeric priority")
	}
}

func TestUserDraftCollections(t *testing.T) {
	xs := []UserDraft{SampleUserDraft("a", when), SampleUserDraft("bb", when), SampleUserDraft("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserDrafts(xs)); got != 2 {
		t.Errorf("ActiveUserDrafts() kept %d, want 2", got)
	}
	if got, want := SumUserDrafts(xs), 1+3; got != want {
		t.Errorf("SumUserDrafts() = %d, want %d", got, want)
	}
	SortUserDrafts(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserDrafts() put %q first", xs[0].ID)
	}
	if len(IndexUserDrafts(xs)) != 3 {
		t.Errorf("IndexUserDrafts() = %v", IndexUserDrafts(xs))
	}
	if err := ValidateUserDrafts(ActiveUserDrafts(xs)); err != nil {
		t.Errorf("ValidateUserDrafts() = %v", err)
	}
}

func TestUserDraftNormalizeClamps(t *testing.T) {
	x := SampleUserDraft("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserRevisionRoundTrip(t *testing.T) {
	x := SampleUserRevision("userrevision-1", when)
	got, err := ParseUserRevision(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserRevision() = %v", err)
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

func TestUserRevisionParseRejectsBadInput(t *testing.T) {
	f := SampleUserRevision("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserRevision(f); err == nil {
		t.Error("ParseUserRevision() accepted a record with no id")
	}
	f = SampleUserRevision("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseUserRevision(f); err == nil {
		t.Error("ParseUserRevision() accepted a non-numeric count")
	}
}

func TestUserRevisionCollections(t *testing.T) {
	xs := []UserRevision{SampleUserRevision("a", when), SampleUserRevision("bb", when), SampleUserRevision("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserRevisions(xs)); got != 2 {
		t.Errorf("ActiveUserRevisions() kept %d, want 2", got)
	}
	if got, want := SumUserRevisions(xs), 1+3; got != want {
		t.Errorf("SumUserRevisions() = %d, want %d", got, want)
	}
	SortUserRevisions(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserRevisions() put %q first", xs[0].ID)
	}
	if len(IndexUserRevisions(xs)) != 3 {
		t.Errorf("IndexUserRevisions() = %v", IndexUserRevisions(xs))
	}
	if err := ValidateUserRevisions(ActiveUserRevisions(xs)); err != nil {
		t.Errorf("ValidateUserRevisions() = %v", err)
	}
}

func TestUserRevisionNormalizeClamps(t *testing.T) {
	x := SampleUserRevision("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserSnapshotRoundTrip(t *testing.T) {
	x := SampleUserSnapshot("usersnapshot-1", when)
	got, err := ParseUserSnapshot(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserSnapshot() = %v", err)
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

func TestUserSnapshotParseRejectsBadInput(t *testing.T) {
	f := SampleUserSnapshot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserSnapshot(f); err == nil {
		t.Error("ParseUserSnapshot() accepted a record with no id")
	}
	f = SampleUserSnapshot("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseUserSnapshot(f); err == nil {
		t.Error("ParseUserSnapshot() accepted a non-numeric attempts")
	}
}

func TestUserSnapshotCollections(t *testing.T) {
	xs := []UserSnapshot{SampleUserSnapshot("a", when), SampleUserSnapshot("bb", when), SampleUserSnapshot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserSnapshots(xs)); got != 2 {
		t.Errorf("ActiveUserSnapshots() kept %d, want 2", got)
	}
	if got, want := SumUserSnapshots(xs), 1+3; got != want {
		t.Errorf("SumUserSnapshots() = %d, want %d", got, want)
	}
	SortUserSnapshots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserSnapshots() put %q first", xs[0].ID)
	}
	if len(IndexUserSnapshots(xs)) != 3 {
		t.Errorf("IndexUserSnapshots() = %v", IndexUserSnapshots(xs))
	}
	if err := ValidateUserSnapshots(ActiveUserSnapshots(xs)); err != nil {
		t.Errorf("ValidateUserSnapshots() = %v", err)
	}
}

func TestUserSnapshotNormalizeClamps(t *testing.T) {
	x := SampleUserSnapshot("a", when)
	x.Owner = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserBatchRoundTrip(t *testing.T) {
	x := SampleUserBatch("userbatch-1", when)
	got, err := ParseUserBatch(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserBatch() = %v", err)
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

func TestUserBatchParseRejectsBadInput(t *testing.T) {
	f := SampleUserBatch("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserBatch(f); err == nil {
		t.Error("ParseUserBatch() accepted a record with no id")
	}
	f = SampleUserBatch("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUserBatch(f); err == nil {
		t.Error("ParseUserBatch() accepted a non-numeric offset")
	}
}

func TestUserBatchCollections(t *testing.T) {
	xs := []UserBatch{SampleUserBatch("a", when), SampleUserBatch("bb", when), SampleUserBatch("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserBatchs(xs)); got != 2 {
		t.Errorf("ActiveUserBatchs() kept %d, want 2", got)
	}
	if got, want := SumUserBatchs(xs), 1+3; got != want {
		t.Errorf("SumUserBatchs() = %d, want %d", got, want)
	}
	SortUserBatchs(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserBatchs() put %q first", xs[0].ID)
	}
	if len(IndexUserBatchs(xs)) != 3 {
		t.Errorf("IndexUserBatchs() = %v", IndexUserBatchs(xs))
	}
	if err := ValidateUserBatchs(ActiveUserBatchs(xs)); err != nil {
		t.Errorf("ValidateUserBatchs() = %v", err)
	}
}

func TestUserBatchNormalizeClamps(t *testing.T) {
	x := SampleUserBatch("a", when)
	x.Channel = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserRequestRoundTrip(t *testing.T) {
	x := SampleUserRequest("userrequest-1", when)
	got, err := ParseUserRequest(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserRequest() = %v", err)
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

func TestUserRequestParseRejectsBadInput(t *testing.T) {
	f := SampleUserRequest("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserRequest(f); err == nil {
		t.Error("ParseUserRequest() accepted a record with no id")
	}
	f = SampleUserRequest("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUserRequest(f); err == nil {
		t.Error("ParseUserRequest() accepted a non-numeric offset")
	}
}

func TestUserRequestCollections(t *testing.T) {
	xs := []UserRequest{SampleUserRequest("a", when), SampleUserRequest("bb", when), SampleUserRequest("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserRequests(xs)); got != 2 {
		t.Errorf("ActiveUserRequests() kept %d, want 2", got)
	}
	if got, want := SumUserRequests(xs), 1+3; got != want {
		t.Errorf("SumUserRequests() = %d, want %d", got, want)
	}
	SortUserRequests(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserRequests() put %q first", xs[0].ID)
	}
	if len(IndexUserRequests(xs)) != 3 {
		t.Errorf("IndexUserRequests() = %v", IndexUserRequests(xs))
	}
	if err := ValidateUserRequests(ActiveUserRequests(xs)); err != nil {
		t.Errorf("ValidateUserRequests() = %v", err)
	}
}

func TestUserRequestNormalizeClamps(t *testing.T) {
	x := SampleUserRequest("a", when)
	x.Reference = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Reference != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserSummaryRoundTrip(t *testing.T) {
	x := SampleUserSummary("usersummary-1", when)
	got, err := ParseUserSummary(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserSummary() = %v", err)
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

func TestUserSummaryParseRejectsBadInput(t *testing.T) {
	f := SampleUserSummary("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserSummary(f); err == nil {
		t.Error("ParseUserSummary() accepted a record with no id")
	}
	f = SampleUserSummary("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseUserSummary(f); err == nil {
		t.Error("ParseUserSummary() accepted a non-numeric priority")
	}
}

func TestUserSummaryCollections(t *testing.T) {
	xs := []UserSummary{SampleUserSummary("a", when), SampleUserSummary("bb", when), SampleUserSummary("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserSummarys(xs)); got != 2 {
		t.Errorf("ActiveUserSummarys() kept %d, want 2", got)
	}
	if got, want := SumUserSummarys(xs), 1+3; got != want {
		t.Errorf("SumUserSummarys() = %d, want %d", got, want)
	}
	SortUserSummarys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserSummarys() put %q first", xs[0].ID)
	}
	if len(IndexUserSummarys(xs)) != 3 {
		t.Errorf("IndexUserSummarys() = %v", IndexUserSummarys(xs))
	}
	if err := ValidateUserSummarys(ActiveUserSummarys(xs)); err != nil {
		t.Errorf("ValidateUserSummarys() = %v", err)
	}
}

func TestUserSummaryNormalizeClamps(t *testing.T) {
	x := SampleUserSummary("a", when)
	x.Channel = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Channel != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserPolicyRoundTrip(t *testing.T) {
	x := SampleUserPolicy("userpolicy-1", when)
	got, err := ParseUserPolicy(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserPolicy() = %v", err)
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

func TestUserPolicyParseRejectsBadInput(t *testing.T) {
	f := SampleUserPolicy("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserPolicy(f); err == nil {
		t.Error("ParseUserPolicy() accepted a record with no id")
	}
	f = SampleUserPolicy("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUserPolicy(f); err == nil {
		t.Error("ParseUserPolicy() accepted a non-numeric offset")
	}
}

func TestUserPolicyCollections(t *testing.T) {
	xs := []UserPolicy{SampleUserPolicy("a", when), SampleUserPolicy("bb", when), SampleUserPolicy("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserPolicys(xs)); got != 2 {
		t.Errorf("ActiveUserPolicys() kept %d, want 2", got)
	}
	if got, want := SumUserPolicys(xs), 1+3; got != want {
		t.Errorf("SumUserPolicys() = %d, want %d", got, want)
	}
	SortUserPolicys(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserPolicys() put %q first", xs[0].ID)
	}
	if len(IndexUserPolicys(xs)) != 3 {
		t.Errorf("IndexUserPolicys() = %v", IndexUserPolicys(xs))
	}
	if err := ValidateUserPolicys(ActiveUserPolicys(xs)); err != nil {
		t.Errorf("ValidateUserPolicys() = %v", err)
	}
}

func TestUserPolicyNormalizeClamps(t *testing.T) {
	x := SampleUserPolicy("a", when)
	x.Label = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserLinkRoundTrip(t *testing.T) {
	x := SampleUserLink("userlink-1", when)
	got, err := ParseUserLink(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserLink() = %v", err)
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

func TestUserLinkParseRejectsBadInput(t *testing.T) {
	f := SampleUserLink("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserLink(f); err == nil {
		t.Error("ParseUserLink() accepted a record with no id")
	}
	f = SampleUserLink("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseUserLink(f); err == nil {
		t.Error("ParseUserLink() accepted a non-numeric attempts")
	}
}

func TestUserLinkCollections(t *testing.T) {
	xs := []UserLink{SampleUserLink("a", when), SampleUserLink("bb", when), SampleUserLink("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserLinks(xs)); got != 2 {
		t.Errorf("ActiveUserLinks() kept %d, want 2", got)
	}
	if got, want := SumUserLinks(xs), 1+3; got != want {
		t.Errorf("SumUserLinks() = %d, want %d", got, want)
	}
	SortUserLinks(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserLinks() put %q first", xs[0].ID)
	}
	if len(IndexUserLinks(xs)) != 3 {
		t.Errorf("IndexUserLinks() = %v", IndexUserLinks(xs))
	}
	if err := ValidateUserLinks(ActiveUserLinks(xs)); err != nil {
		t.Errorf("ValidateUserLinks() = %v", err)
	}
}

func TestUserLinkNormalizeClamps(t *testing.T) {
	x := SampleUserLink("a", when)
	x.Comment = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Comment != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserNoteRoundTrip(t *testing.T) {
	x := SampleUserNote("usernote-1", when)
	got, err := ParseUserNote(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserNote() = %v", err)
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

func TestUserNoteParseRejectsBadInput(t *testing.T) {
	f := SampleUserNote("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserNote(f); err == nil {
		t.Error("ParseUserNote() accepted a record with no id")
	}
	f = SampleUserNote("x", when).Fields()
	f["capacity"] = "not-a-number"
	if _, err := ParseUserNote(f); err == nil {
		t.Error("ParseUserNote() accepted a non-numeric capacity")
	}
}

func TestUserNoteCollections(t *testing.T) {
	xs := []UserNote{SampleUserNote("a", when), SampleUserNote("bb", when), SampleUserNote("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserNotes(xs)); got != 2 {
		t.Errorf("ActiveUserNotes() kept %d, want 2", got)
	}
	if got, want := SumUserNotes(xs), 1+3; got != want {
		t.Errorf("SumUserNotes() = %d, want %d", got, want)
	}
	SortUserNotes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserNotes() put %q first", xs[0].ID)
	}
	if len(IndexUserNotes(xs)) != 3 {
		t.Errorf("IndexUserNotes() = %v", IndexUserNotes(xs))
	}
	if err := ValidateUserNotes(ActiveUserNotes(xs)); err != nil {
		t.Errorf("ValidateUserNotes() = %v", err)
	}
}

func TestUserNoteNormalizeClamps(t *testing.T) {
	x := SampleUserNote("a", when)
	x.Region = "  padded  "
	x.Capacity = -5
	got := x.Normalize()
	if got.Region != "padded" || got.Capacity != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserWindowRoundTrip(t *testing.T) {
	x := SampleUserWindow("userwindow-1", when)
	got, err := ParseUserWindow(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserWindow() = %v", err)
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

func TestUserWindowParseRejectsBadInput(t *testing.T) {
	f := SampleUserWindow("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserWindow(f); err == nil {
		t.Error("ParseUserWindow() accepted a record with no id")
	}
	f = SampleUserWindow("x", when).Fields()
	f["priority"] = "not-a-number"
	if _, err := ParseUserWindow(f); err == nil {
		t.Error("ParseUserWindow() accepted a non-numeric priority")
	}
}

func TestUserWindowCollections(t *testing.T) {
	xs := []UserWindow{SampleUserWindow("a", when), SampleUserWindow("bb", when), SampleUserWindow("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserWindows(xs)); got != 2 {
		t.Errorf("ActiveUserWindows() kept %d, want 2", got)
	}
	if got, want := SumUserWindows(xs), 1+3; got != want {
		t.Errorf("SumUserWindows() = %d, want %d", got, want)
	}
	SortUserWindows(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserWindows() put %q first", xs[0].ID)
	}
	if len(IndexUserWindows(xs)) != 3 {
		t.Errorf("IndexUserWindows() = %v", IndexUserWindows(xs))
	}
	if err := ValidateUserWindows(ActiveUserWindows(xs)); err != nil {
		t.Errorf("ValidateUserWindows() = %v", err)
	}
}

func TestUserWindowNormalizeClamps(t *testing.T) {
	x := SampleUserWindow("a", when)
	x.Label = "  padded  "
	x.Priority = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Priority != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserQuotaRoundTrip(t *testing.T) {
	x := SampleUserQuota("userquota-1", when)
	got, err := ParseUserQuota(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserQuota() = %v", err)
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

func TestUserQuotaParseRejectsBadInput(t *testing.T) {
	f := SampleUserQuota("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserQuota(f); err == nil {
		t.Error("ParseUserQuota() accepted a record with no id")
	}
	f = SampleUserQuota("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUserQuota(f); err == nil {
		t.Error("ParseUserQuota() accepted a non-numeric offset")
	}
}

func TestUserQuotaCollections(t *testing.T) {
	xs := []UserQuota{SampleUserQuota("a", when), SampleUserQuota("bb", when), SampleUserQuota("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserQuotas(xs)); got != 2 {
		t.Errorf("ActiveUserQuotas() kept %d, want 2", got)
	}
	if got, want := SumUserQuotas(xs), 1+3; got != want {
		t.Errorf("SumUserQuotas() = %d, want %d", got, want)
	}
	SortUserQuotas(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserQuotas() put %q first", xs[0].ID)
	}
	if len(IndexUserQuotas(xs)) != 3 {
		t.Errorf("IndexUserQuotas() = %v", IndexUserQuotas(xs))
	}
	if err := ValidateUserQuotas(ActiveUserQuotas(xs)); err != nil {
		t.Errorf("ValidateUserQuotas() = %v", err)
	}
}

func TestUserQuotaNormalizeClamps(t *testing.T) {
	x := SampleUserQuota("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserBucketRoundTrip(t *testing.T) {
	x := SampleUserBucket("userbucket-1", when)
	got, err := ParseUserBucket(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserBucket() = %v", err)
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

func TestUserBucketParseRejectsBadInput(t *testing.T) {
	f := SampleUserBucket("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserBucket(f); err == nil {
		t.Error("ParseUserBucket() accepted a record with no id")
	}
	f = SampleUserBucket("x", when).Fields()
	f["attempts"] = "not-a-number"
	if _, err := ParseUserBucket(f); err == nil {
		t.Error("ParseUserBucket() accepted a non-numeric attempts")
	}
}

func TestUserBucketCollections(t *testing.T) {
	xs := []UserBucket{SampleUserBucket("a", when), SampleUserBucket("bb", when), SampleUserBucket("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserBuckets(xs)); got != 2 {
		t.Errorf("ActiveUserBuckets() kept %d, want 2", got)
	}
	if got, want := SumUserBuckets(xs), 1+3; got != want {
		t.Errorf("SumUserBuckets() = %d, want %d", got, want)
	}
	SortUserBuckets(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserBuckets() put %q first", xs[0].ID)
	}
	if len(IndexUserBuckets(xs)) != 3 {
		t.Errorf("IndexUserBuckets() = %v", IndexUserBuckets(xs))
	}
	if err := ValidateUserBuckets(ActiveUserBuckets(xs)); err != nil {
		t.Errorf("ValidateUserBuckets() = %v", err)
	}
}

func TestUserBucketNormalizeClamps(t *testing.T) {
	x := SampleUserBucket("a", when)
	x.Label = "  padded  "
	x.Attempts = -5
	got := x.Normalize()
	if got.Label != "padded" || got.Attempts != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserLaneRoundTrip(t *testing.T) {
	x := SampleUserLane("userlane-1", when)
	got, err := ParseUserLane(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserLane() = %v", err)
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

func TestUserLaneParseRejectsBadInput(t *testing.T) {
	f := SampleUserLane("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserLane(f); err == nil {
		t.Error("ParseUserLane() accepted a record with no id")
	}
	f = SampleUserLane("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseUserLane(f); err == nil {
		t.Error("ParseUserLane() accepted a non-numeric count")
	}
}

func TestUserLaneCollections(t *testing.T) {
	xs := []UserLane{SampleUserLane("a", when), SampleUserLane("bb", when), SampleUserLane("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserLanes(xs)); got != 2 {
		t.Errorf("ActiveUserLanes() kept %d, want 2", got)
	}
	if got, want := SumUserLanes(xs), 1+3; got != want {
		t.Errorf("SumUserLanes() = %d, want %d", got, want)
	}
	SortUserLanes(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserLanes() put %q first", xs[0].ID)
	}
	if len(IndexUserLanes(xs)) != 3 {
		t.Errorf("IndexUserLanes() = %v", IndexUserLanes(xs))
	}
	if err := ValidateUserLanes(ActiveUserLanes(xs)); err != nil {
		t.Errorf("ValidateUserLanes() = %v", err)
	}
}

func TestUserLaneNormalizeClamps(t *testing.T) {
	x := SampleUserLane("a", when)
	x.Owner = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserSlotRoundTrip(t *testing.T) {
	x := SampleUserSlot("userslot-1", when)
	got, err := ParseUserSlot(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserSlot() = %v", err)
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

func TestUserSlotParseRejectsBadInput(t *testing.T) {
	f := SampleUserSlot("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserSlot(f); err == nil {
		t.Error("ParseUserSlot() accepted a record with no id")
	}
	f = SampleUserSlot("x", when).Fields()
	f["count"] = "not-a-number"
	if _, err := ParseUserSlot(f); err == nil {
		t.Error("ParseUserSlot() accepted a non-numeric count")
	}
}

func TestUserSlotCollections(t *testing.T) {
	xs := []UserSlot{SampleUserSlot("a", when), SampleUserSlot("bb", when), SampleUserSlot("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserSlots(xs)); got != 2 {
		t.Errorf("ActiveUserSlots() kept %d, want 2", got)
	}
	if got, want := SumUserSlots(xs), 1+3; got != want {
		t.Errorf("SumUserSlots() = %d, want %d", got, want)
	}
	SortUserSlots(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserSlots() put %q first", xs[0].ID)
	}
	if len(IndexUserSlots(xs)) != 3 {
		t.Errorf("IndexUserSlots() = %v", IndexUserSlots(xs))
	}
	if err := ValidateUserSlots(ActiveUserSlots(xs)); err != nil {
		t.Errorf("ValidateUserSlots() = %v", err)
	}
}

func TestUserSlotNormalizeClamps(t *testing.T) {
	x := SampleUserSlot("a", when)
	x.Group = "  padded  "
	x.Count = -5
	got := x.Normalize()
	if got.Group != "padded" || got.Count != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserTagRoundTrip(t *testing.T) {
	x := SampleUserTag("usertag-1", when)
	got, err := ParseUserTag(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserTag() = %v", err)
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

func TestUserTagParseRejectsBadInput(t *testing.T) {
	f := SampleUserTag("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserTag(f); err == nil {
		t.Error("ParseUserTag() accepted a record with no id")
	}
	f = SampleUserTag("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUserTag(f); err == nil {
		t.Error("ParseUserTag() accepted a non-numeric offset")
	}
}

func TestUserTagCollections(t *testing.T) {
	xs := []UserTag{SampleUserTag("a", when), SampleUserTag("bb", when), SampleUserTag("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserTags(xs)); got != 2 {
		t.Errorf("ActiveUserTags() kept %d, want 2", got)
	}
	if got, want := SumUserTags(xs), 1+3; got != want {
		t.Errorf("SumUserTags() = %d, want %d", got, want)
	}
	SortUserTags(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserTags() put %q first", xs[0].ID)
	}
	if len(IndexUserTags(xs)) != 3 {
		t.Errorf("IndexUserTags() = %v", IndexUserTags(xs))
	}
	if err := ValidateUserTags(ActiveUserTags(xs)); err != nil {
		t.Errorf("ValidateUserTags() = %v", err)
	}
}

func TestUserTagNormalizeClamps(t *testing.T) {
	x := SampleUserTag("a", when)
	x.Owner = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserRuleRoundTrip(t *testing.T) {
	x := SampleUserRule("userrule-1", when)
	got, err := ParseUserRule(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserRule() = %v", err)
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

func TestUserRuleParseRejectsBadInput(t *testing.T) {
	f := SampleUserRule("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserRule(f); err == nil {
		t.Error("ParseUserRule() accepted a record with no id")
	}
	f = SampleUserRule("x", when).Fields()
	f["offset"] = "not-a-number"
	if _, err := ParseUserRule(f); err == nil {
		t.Error("ParseUserRule() accepted a non-numeric offset")
	}
}

func TestUserRuleCollections(t *testing.T) {
	xs := []UserRule{SampleUserRule("a", when), SampleUserRule("bb", when), SampleUserRule("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserRules(xs)); got != 2 {
		t.Errorf("ActiveUserRules() kept %d, want 2", got)
	}
	if got, want := SumUserRules(xs), 1+3; got != want {
		t.Errorf("SumUserRules() = %d, want %d", got, want)
	}
	SortUserRules(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserRules() put %q first", xs[0].ID)
	}
	if len(IndexUserRules(xs)) != 3 {
		t.Errorf("IndexUserRules() = %v", IndexUserRules(xs))
	}
	if err := ValidateUserRules(ActiveUserRules(xs)); err != nil {
		t.Errorf("ValidateUserRules() = %v", err)
	}
}

func TestUserRuleNormalizeClamps(t *testing.T) {
	x := SampleUserRule("a", when)
	x.Name = "  padded  "
	x.Offset = -5
	got := x.Normalize()
	if got.Name != "padded" || got.Offset != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserHoldRoundTrip(t *testing.T) {
	x := SampleUserHold("userhold-1", when)
	got, err := ParseUserHold(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserHold() = %v", err)
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

func TestUserHoldParseRejectsBadInput(t *testing.T) {
	f := SampleUserHold("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserHold(f); err == nil {
		t.Error("ParseUserHold() accepted a record with no id")
	}
	f = SampleUserHold("x", when).Fields()
	f["version"] = "not-a-number"
	if _, err := ParseUserHold(f); err == nil {
		t.Error("ParseUserHold() accepted a non-numeric version")
	}
}

func TestUserHoldCollections(t *testing.T) {
	xs := []UserHold{SampleUserHold("a", when), SampleUserHold("bb", when), SampleUserHold("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserHolds(xs)); got != 2 {
		t.Errorf("ActiveUserHolds() kept %d, want 2", got)
	}
	if got, want := SumUserHolds(xs), 1+3; got != want {
		t.Errorf("SumUserHolds() = %d, want %d", got, want)
	}
	SortUserHolds(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserHolds() put %q first", xs[0].ID)
	}
	if len(IndexUserHolds(xs)) != 3 {
		t.Errorf("IndexUserHolds() = %v", IndexUserHolds(xs))
	}
	if err := ValidateUserHolds(ActiveUserHolds(xs)); err != nil {
		t.Errorf("ValidateUserHolds() = %v", err)
	}
}

func TestUserHoldNormalizeClamps(t *testing.T) {
	x := SampleUserHold("a", when)
	x.Category = "  padded  "
	x.Version = -5
	got := x.Normalize()
	if got.Category != "padded" || got.Version != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserClaimRoundTrip(t *testing.T) {
	x := SampleUserClaim("userclaim-1", when)
	got, err := ParseUserClaim(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserClaim() = %v", err)
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

func TestUserClaimParseRejectsBadInput(t *testing.T) {
	f := SampleUserClaim("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserClaim(f); err == nil {
		t.Error("ParseUserClaim() accepted a record with no id")
	}
	f = SampleUserClaim("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseUserClaim(f); err == nil {
		t.Error("ParseUserClaim() accepted a non-numeric weight")
	}
}

func TestUserClaimCollections(t *testing.T) {
	xs := []UserClaim{SampleUserClaim("a", when), SampleUserClaim("bb", when), SampleUserClaim("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserClaims(xs)); got != 2 {
		t.Errorf("ActiveUserClaims() kept %d, want 2", got)
	}
	if got, want := SumUserClaims(xs), 1+3; got != want {
		t.Errorf("SumUserClaims() = %d, want %d", got, want)
	}
	SortUserClaims(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserClaims() put %q first", xs[0].ID)
	}
	if len(IndexUserClaims(xs)) != 3 {
		t.Errorf("IndexUserClaims() = %v", IndexUserClaims(xs))
	}
	if err := ValidateUserClaims(ActiveUserClaims(xs)); err != nil {
		t.Errorf("ValidateUserClaims() = %v", err)
	}
}

func TestUserClaimNormalizeClamps(t *testing.T) {
	x := SampleUserClaim("a", when)
	x.Owner = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}

func TestUserCursorRoundTrip(t *testing.T) {
	x := SampleUserCursor("usercursor-1", when)
	got, err := ParseUserCursor(x.Fields())
	if err != nil {
		t.Fatalf("ParseUserCursor() = %v", err)
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

func TestUserCursorParseRejectsBadInput(t *testing.T) {
	f := SampleUserCursor("x", when).Fields()
	delete(f, "id")
	if _, err := ParseUserCursor(f); err == nil {
		t.Error("ParseUserCursor() accepted a record with no id")
	}
	f = SampleUserCursor("x", when).Fields()
	f["weight"] = "not-a-number"
	if _, err := ParseUserCursor(f); err == nil {
		t.Error("ParseUserCursor() accepted a non-numeric weight")
	}
}

func TestUserCursorCollections(t *testing.T) {
	xs := []UserCursor{SampleUserCursor("a", when), SampleUserCursor("bb", when), SampleUserCursor("ccc", when)}
	xs[1].Archived = true
	if got := len(ActiveUserCursors(xs)); got != 2 {
		t.Errorf("ActiveUserCursors() kept %d, want 2", got)
	}
	if got, want := SumUserCursors(xs), 1+3; got != want {
		t.Errorf("SumUserCursors() = %d, want %d", got, want)
	}
	SortUserCursors(xs)
	if xs[0].ID != "ccc" {
		t.Errorf("SortUserCursors() put %q first", xs[0].ID)
	}
	if len(IndexUserCursors(xs)) != 3 {
		t.Errorf("IndexUserCursors() = %v", IndexUserCursors(xs))
	}
	if err := ValidateUserCursors(ActiveUserCursors(xs)); err != nil {
		t.Errorf("ValidateUserCursors() = %v", err)
	}
}

func TestUserCursorNormalizeClamps(t *testing.T) {
	x := SampleUserCursor("a", when)
	x.Owner = "  padded  "
	x.Weight = -5
	got := x.Normalize()
	if got.Owner != "padded" || got.Weight != 0 {
		t.Errorf("Normalize() = %+v", got)
	}
}
