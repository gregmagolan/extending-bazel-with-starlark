// Package user holds the user domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package user

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// UserRecordSchema is the stored shape of a UserRecord.
var UserRecordSchema = schema.New("user.userrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserRecord is one user record.
type UserRecord struct {
	ID       string
	Channel  string
	Category string
	Source   string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserRecord) Key() string { return "user/userrecord/" + x.ID }

// Fields encodes x for storage.
func (x UserRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserRecord decodes a stored record.
func ParseUserRecord(f codec.Fields) (UserRecord, error) {
	var x UserRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userrecord: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("user: userrecord %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: userrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserRecord) String() string {
	return fmt.Sprintf("UserRecord(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserRecord) Normalize() UserRecord {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortUserRecords orders by capacity descending, then by ID.
func SortUserRecords(xs []UserRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserRecords drops the archived entries, preserving order.
func ActiveUserRecords(xs []UserRecord) []UserRecord {
	out := make([]UserRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserRecords adds up the capacity of every active entry.
func SumUserRecords(xs []UserRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexUserRecords keys the entries by ID, last write winning.
func IndexUserRecords(xs []UserRecord) map[string]UserRecord {
	out := make(map[string]UserRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserRecord builds a valid UserRecord, for tests and for the seed data the cron
// binary writes.
func SampleUserRecord(id string, at time.Time) UserRecord {
	return UserRecord{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Source:   "seeded by user.SampleUserRecord",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserEntrySchema is the stored shape of a UserEntry.
var UserEntrySchema = schema.New("user.userentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserEntry is one user entry.
type UserEntry struct {
	ID       string
	Region   string
	Label    string
	Group    string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserEntry) Key() string { return "user/userentry/" + x.ID }

// Fields encodes x for storage.
func (x UserEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserEntry decodes a stored record.
func ParseUserEntry(f codec.Fields) (UserEntry, error) {
	var x UserEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userentry: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("user: userentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("user: userentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserEntry) String() string {
	return fmt.Sprintf("UserEntry(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserEntry) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserEntry) Normalize() UserEntry {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortUserEntrys orders by version descending, then by ID.
func SortUserEntrys(xs []UserEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserEntrys drops the archived entries, preserving order.
func ActiveUserEntrys(xs []UserEntry) []UserEntry {
	out := make([]UserEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserEntrys adds up the version of every active entry.
func SumUserEntrys(xs []UserEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexUserEntrys keys the entries by ID, last write winning.
func IndexUserEntrys(xs []UserEntry) map[string]UserEntry {
	out := make(map[string]UserEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserEntry builds a valid UserEntry, for tests and for the seed data the cron
// binary writes.
func SampleUserEntry(id string, at time.Time) UserEntry {
	return UserEntry{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Group:    "seeded by user.SampleUserEntry",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserItemSchema is the stored shape of a UserItem.
var UserItemSchema = schema.New("user.useritem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserItem is one user item.
type UserItem struct {
	ID        string
	Reference string
	Group     string
	Owner     string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UserItem) Key() string { return "user/useritem/" + x.ID }

// Fields encodes x for storage.
func (x UserItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"owner":     x.Owner,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserItem decodes a stored record.
func ParseUserItem(f codec.Fields) (UserItem, error) {
	var x UserItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: useritem: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("user: useritem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("user: useritem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: useritem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: useritem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserItem) String() string {
	return fmt.Sprintf("UserItem(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserItem) Normalize() UserItem {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortUserItems orders by version descending, then by ID.
func SortUserItems(xs []UserItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserItems drops the archived entries, preserving order.
func ActiveUserItems(xs []UserItem) []UserItem {
	out := make([]UserItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserItems adds up the version of every active entry.
func SumUserItems(xs []UserItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexUserItems keys the entries by ID, last write winning.
func IndexUserItems(xs []UserItem) map[string]UserItem {
	out := make(map[string]UserItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserItem builds a valid UserItem, for tests and for the seed data the cron
// binary writes.
func SampleUserItem(id string, at time.Time) UserItem {
	return UserItem{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Owner:     "seeded by user.SampleUserItem",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UserDraftSchema is the stored shape of a UserDraft.
var UserDraftSchema = schema.New("user.userdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserDraft is one user draft.
type UserDraft struct {
	ID        string
	Channel   string
	Reference string
	Owner     string
	Priority  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UserDraft) Key() string { return "user/userdraft/" + x.ID }

// Fields encodes x for storage.
func (x UserDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"owner":     x.Owner,
		"priority":  strconv.Itoa(x.Priority),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserDraft decodes a stored record.
func ParseUserDraft(f codec.Fields) (UserDraft, error) {
	var x UserDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userdraft: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("user: userdraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("user: userdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserDraft) String() string {
	return fmt.Sprintf("UserDraft(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserDraft) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserDraft) Normalize() UserDraft {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortUserDrafts orders by priority descending, then by ID.
func SortUserDrafts(xs []UserDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserDrafts drops the archived entries, preserving order.
func ActiveUserDrafts(xs []UserDraft) []UserDraft {
	out := make([]UserDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserDrafts adds up the priority of every active entry.
func SumUserDrafts(xs []UserDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexUserDrafts keys the entries by ID, last write winning.
func IndexUserDrafts(xs []UserDraft) map[string]UserDraft {
	out := make(map[string]UserDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserDraft builds a valid UserDraft, for tests and for the seed data the cron
// binary writes.
func SampleUserDraft(id string, at time.Time) UserDraft {
	return UserDraft{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by user.SampleUserDraft",
		Priority:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UserRevisionSchema is the stored shape of a UserRevision.
var UserRevisionSchema = schema.New("user.userrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserRevision is one user revision.
type UserRevision struct {
	ID        string
	Owner     string
	Reference string
	Comment   string
	Count     int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UserRevision) Key() string { return "user/userrevision/" + x.ID }

// Fields encodes x for storage.
func (x UserRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"reference": x.Reference,
		"comment":   x.Comment,
		"count":     strconv.Itoa(x.Count),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserRevision decodes a stored record.
func ParseUserRevision(f codec.Fields) (UserRevision, error) {
	var x UserRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userrevision: missing id")
	}
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("user: userrevision %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("user: userrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserRevision) String() string {
	return fmt.Sprintf("UserRevision(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserRevision) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserRevision) Normalize() UserRevision {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortUserRevisions orders by count descending, then by ID.
func SortUserRevisions(xs []UserRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserRevisions drops the archived entries, preserving order.
func ActiveUserRevisions(xs []UserRevision) []UserRevision {
	out := make([]UserRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserRevisions adds up the count of every active entry.
func SumUserRevisions(xs []UserRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexUserRevisions keys the entries by ID, last write winning.
func IndexUserRevisions(xs []UserRevision) map[string]UserRevision {
	out := make(map[string]UserRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserRevision builds a valid UserRevision, for tests and for the seed data the cron
// binary writes.
func SampleUserRevision(id string, at time.Time) UserRevision {
	return UserRevision{
		ID:        id,
		Owner:     "owner-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by user.SampleUserRevision",
		Count:     len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UserSnapshotSchema is the stored shape of a UserSnapshot.
var UserSnapshotSchema = schema.New("user.usersnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserSnapshot is one user snapshot.
type UserSnapshot struct {
	ID       string
	Owner    string
	Region   string
	Source   string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserSnapshot) Key() string { return "user/usersnapshot/" + x.ID }

// Fields encodes x for storage.
func (x UserSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserSnapshot decodes a stored record.
func ParseUserSnapshot(f codec.Fields) (UserSnapshot, error) {
	var x UserSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: usersnapshot: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("user: usersnapshot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: usersnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: usersnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: usersnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserSnapshot) String() string {
	return fmt.Sprintf("UserSnapshot(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserSnapshot) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserSnapshot) Normalize() UserSnapshot {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortUserSnapshots orders by attempts descending, then by ID.
func SortUserSnapshots(xs []UserSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserSnapshots drops the archived entries, preserving order.
func ActiveUserSnapshots(xs []UserSnapshot) []UserSnapshot {
	out := make([]UserSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserSnapshots adds up the attempts of every active entry.
func SumUserSnapshots(xs []UserSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexUserSnapshots keys the entries by ID, last write winning.
func IndexUserSnapshots(xs []UserSnapshot) map[string]UserSnapshot {
	out := make(map[string]UserSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserSnapshot builds a valid UserSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleUserSnapshot(id string, at time.Time) UserSnapshot {
	return UserSnapshot{
		ID:       id,
		Owner:    "owner-" + id,
		Region:   "region-" + id,
		Source:   "seeded by user.SampleUserSnapshot",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserBatchSchema is the stored shape of a UserBatch.
var UserBatchSchema = schema.New("user.userbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserBatch is one user batch.
type UserBatch struct {
	ID       string
	Channel  string
	Group    string
	Source   string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserBatch) Key() string { return "user/userbatch/" + x.ID }

// Fields encodes x for storage.
func (x UserBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserBatch decodes a stored record.
func ParseUserBatch(f codec.Fields) (UserBatch, error) {
	var x UserBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userbatch: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: userbatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("user: userbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserBatch) String() string {
	return fmt.Sprintf("UserBatch(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserBatch) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserBatch) Normalize() UserBatch {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortUserBatchs orders by offset descending, then by ID.
func SortUserBatchs(xs []UserBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserBatchs drops the archived entries, preserving order.
func ActiveUserBatchs(xs []UserBatch) []UserBatch {
	out := make([]UserBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserBatchs adds up the offset of every active entry.
func SumUserBatchs(xs []UserBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUserBatchs keys the entries by ID, last write winning.
func IndexUserBatchs(xs []UserBatch) map[string]UserBatch {
	out := make(map[string]UserBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserBatch builds a valid UserBatch, for tests and for the seed data the cron
// binary writes.
func SampleUserBatch(id string, at time.Time) UserBatch {
	return UserBatch{
		ID:      id,
		Channel: "channel-" + id,
		Group:   "group-" + id,
		Source:  "seeded by user.SampleUserBatch",
		Offset:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// UserRequestSchema is the stored shape of a UserRequest.
var UserRequestSchema = schema.New("user.userrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserRequest is one user request.
type UserRequest struct {
	ID        string
	Reference string
	Label     string
	Owner     string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UserRequest) Key() string { return "user/userrequest/" + x.ID }

// Fields encodes x for storage.
func (x UserRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"owner":     x.Owner,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserRequest decodes a stored record.
func ParseUserRequest(f codec.Fields) (UserRequest, error) {
	var x UserRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userrequest: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: userrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("user: userrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserRequest) String() string {
	return fmt.Sprintf("UserRequest(%s, reference=%q, offset=%d, archived=%t)", x.ID, x.Reference, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserRequest) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserRequest) Normalize() UserRequest {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortUserRequests orders by offset descending, then by ID.
func SortUserRequests(xs []UserRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserRequests drops the archived entries, preserving order.
func ActiveUserRequests(xs []UserRequest) []UserRequest {
	out := make([]UserRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserRequests adds up the offset of every active entry.
func SumUserRequests(xs []UserRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUserRequests keys the entries by ID, last write winning.
func IndexUserRequests(xs []UserRequest) map[string]UserRequest {
	out := make(map[string]UserRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserRequest builds a valid UserRequest, for tests and for the seed data the cron
// binary writes.
func SampleUserRequest(id string, at time.Time) UserRequest {
	return UserRequest{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Owner:     "seeded by user.SampleUserRequest",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UserSummarySchema is the stored shape of a UserSummary.
var UserSummarySchema = schema.New("user.usersummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserSummary is one user summary.
type UserSummary struct {
	ID       string
	Channel  string
	Category string
	Label    string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserSummary) Key() string { return "user/usersummary/" + x.ID }

// Fields encodes x for storage.
func (x UserSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserSummary decodes a stored record.
func ParseUserSummary(f codec.Fields) (UserSummary, error) {
	var x UserSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: usersummary: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("user: usersummary %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: usersummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: usersummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: usersummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserSummary) String() string {
	return fmt.Sprintf("UserSummary(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserSummary) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserSummary) Normalize() UserSummary {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortUserSummarys orders by priority descending, then by ID.
func SortUserSummarys(xs []UserSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserSummarys drops the archived entries, preserving order.
func ActiveUserSummarys(xs []UserSummary) []UserSummary {
	out := make([]UserSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserSummarys adds up the priority of every active entry.
func SumUserSummarys(xs []UserSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexUserSummarys keys the entries by ID, last write winning.
func IndexUserSummarys(xs []UserSummary) map[string]UserSummary {
	out := make(map[string]UserSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserSummary builds a valid UserSummary, for tests and for the seed data the cron
// binary writes.
func SampleUserSummary(id string, at time.Time) UserSummary {
	return UserSummary{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Label:    "seeded by user.SampleUserSummary",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserPolicySchema is the stored shape of a UserPolicy.
var UserPolicySchema = schema.New("user.userpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserPolicy is one user policy.
type UserPolicy struct {
	ID       string
	Label    string
	Comment  string
	Owner    string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserPolicy) Key() string { return "user/userpolicy/" + x.ID }

// Fields encodes x for storage.
func (x UserPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserPolicy decodes a stored record.
func ParseUserPolicy(f codec.Fields) (UserPolicy, error) {
	var x UserPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userpolicy: missing id")
	}
	x.Label = f["label"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: userpolicy %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("user: userpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserPolicy) String() string {
	return fmt.Sprintf("UserPolicy(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserPolicy) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserPolicy) Normalize() UserPolicy {
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortUserPolicys orders by offset descending, then by ID.
func SortUserPolicys(xs []UserPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserPolicys drops the archived entries, preserving order.
func ActiveUserPolicys(xs []UserPolicy) []UserPolicy {
	out := make([]UserPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserPolicys adds up the offset of every active entry.
func SumUserPolicys(xs []UserPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUserPolicys keys the entries by ID, last write winning.
func IndexUserPolicys(xs []UserPolicy) map[string]UserPolicy {
	out := make(map[string]UserPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserPolicy builds a valid UserPolicy, for tests and for the seed data the cron
// binary writes.
func SampleUserPolicy(id string, at time.Time) UserPolicy {
	return UserPolicy{
		ID:       id,
		Label:    "label-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by user.SampleUserPolicy",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserLinkSchema is the stored shape of a UserLink.
var UserLinkSchema = schema.New("user.userlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserLink is one user link.
type UserLink struct {
	ID       string
	Comment  string
	Source   string
	Label    string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserLink) Key() string { return "user/userlink/" + x.ID }

// Fields encodes x for storage.
func (x UserLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserLink decodes a stored record.
func ParseUserLink(f codec.Fields) (UserLink, error) {
	var x UserLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userlink: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("user: userlink %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("user: userlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserLink) String() string {
	return fmt.Sprintf("UserLink(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserLink) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserLink) Normalize() UserLink {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortUserLinks orders by attempts descending, then by ID.
func SortUserLinks(xs []UserLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserLinks drops the archived entries, preserving order.
func ActiveUserLinks(xs []UserLink) []UserLink {
	out := make([]UserLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserLinks adds up the attempts of every active entry.
func SumUserLinks(xs []UserLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexUserLinks keys the entries by ID, last write winning.
func IndexUserLinks(xs []UserLink) map[string]UserLink {
	out := make(map[string]UserLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserLink builds a valid UserLink, for tests and for the seed data the cron
// binary writes.
func SampleUserLink(id string, at time.Time) UserLink {
	return UserLink{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Label:    "seeded by user.SampleUserLink",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserNoteSchema is the stored shape of a UserNote.
var UserNoteSchema = schema.New("user.usernote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserNote is one user note.
type UserNote struct {
	ID       string
	Region   string
	Category string
	Source   string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserNote) Key() string { return "user/usernote/" + x.ID }

// Fields encodes x for storage.
func (x UserNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserNote decodes a stored record.
func ParseUserNote(f codec.Fields) (UserNote, error) {
	var x UserNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: usernote: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("user: usernote %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("user: usernote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: usernote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: usernote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserNote) String() string {
	return fmt.Sprintf("UserNote(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserNote) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserNote) Normalize() UserNote {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortUserNotes orders by capacity descending, then by ID.
func SortUserNotes(xs []UserNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserNotes drops the archived entries, preserving order.
func ActiveUserNotes(xs []UserNote) []UserNote {
	out := make([]UserNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserNotes adds up the capacity of every active entry.
func SumUserNotes(xs []UserNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexUserNotes keys the entries by ID, last write winning.
func IndexUserNotes(xs []UserNote) map[string]UserNote {
	out := make(map[string]UserNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserNote builds a valid UserNote, for tests and for the seed data the cron
// binary writes.
func SampleUserNote(id string, at time.Time) UserNote {
	return UserNote{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Source:   "seeded by user.SampleUserNote",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserWindowSchema is the stored shape of a UserWindow.
var UserWindowSchema = schema.New("user.userwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserWindow is one user window.
type UserWindow struct {
	ID       string
	Label    string
	Group    string
	Category string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserWindow) Key() string { return "user/userwindow/" + x.ID }

// Fields encodes x for storage.
func (x UserWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"category": x.Category,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserWindow decodes a stored record.
func ParseUserWindow(f codec.Fields) (UserWindow, error) {
	var x UserWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userwindow: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("user: userwindow %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("user: userwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserWindow) String() string {
	return fmt.Sprintf("UserWindow(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserWindow) Normalize() UserWindow {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortUserWindows orders by priority descending, then by ID.
func SortUserWindows(xs []UserWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserWindows drops the archived entries, preserving order.
func ActiveUserWindows(xs []UserWindow) []UserWindow {
	out := make([]UserWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserWindows adds up the priority of every active entry.
func SumUserWindows(xs []UserWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexUserWindows keys the entries by ID, last write winning.
func IndexUserWindows(xs []UserWindow) map[string]UserWindow {
	out := make(map[string]UserWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserWindow builds a valid UserWindow, for tests and for the seed data the cron
// binary writes.
func SampleUserWindow(id string, at time.Time) UserWindow {
	return UserWindow{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Category: "seeded by user.SampleUserWindow",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserQuotaSchema is the stored shape of a UserQuota.
var UserQuotaSchema = schema.New("user.userquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserQuota is one user quota.
type UserQuota struct {
	ID       string
	Name     string
	Category string
	Owner    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserQuota) Key() string { return "user/userquota/" + x.ID }

// Fields encodes x for storage.
func (x UserQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserQuota decodes a stored record.
func ParseUserQuota(f codec.Fields) (UserQuota, error) {
	var x UserQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userquota: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: userquota %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("user: userquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserQuota) String() string {
	return fmt.Sprintf("UserQuota(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserQuota) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserQuota) Normalize() UserQuota {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortUserQuotas orders by offset descending, then by ID.
func SortUserQuotas(xs []UserQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserQuotas drops the archived entries, preserving order.
func ActiveUserQuotas(xs []UserQuota) []UserQuota {
	out := make([]UserQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserQuotas adds up the offset of every active entry.
func SumUserQuotas(xs []UserQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUserQuotas keys the entries by ID, last write winning.
func IndexUserQuotas(xs []UserQuota) map[string]UserQuota {
	out := make(map[string]UserQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserQuota builds a valid UserQuota, for tests and for the seed data the cron
// binary writes.
func SampleUserQuota(id string, at time.Time) UserQuota {
	return UserQuota{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Owner:    "seeded by user.SampleUserQuota",
		Offset:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserBucketSchema is the stored shape of a UserBucket.
var UserBucketSchema = schema.New("user.userbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserBucket is one user bucket.
type UserBucket struct {
	ID       string
	Label    string
	Category string
	Source   string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserBucket) Key() string { return "user/userbucket/" + x.ID }

// Fields encodes x for storage.
func (x UserBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserBucket decodes a stored record.
func ParseUserBucket(f codec.Fields) (UserBucket, error) {
	var x UserBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userbucket: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("user: userbucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("user: userbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserBucket) String() string {
	return fmt.Sprintf("UserBucket(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserBucket) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserBucket) Normalize() UserBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortUserBuckets orders by attempts descending, then by ID.
func SortUserBuckets(xs []UserBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserBuckets drops the archived entries, preserving order.
func ActiveUserBuckets(xs []UserBucket) []UserBucket {
	out := make([]UserBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserBuckets adds up the attempts of every active entry.
func SumUserBuckets(xs []UserBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexUserBuckets keys the entries by ID, last write winning.
func IndexUserBuckets(xs []UserBucket) map[string]UserBucket {
	out := make(map[string]UserBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserBucket builds a valid UserBucket, for tests and for the seed data the cron
// binary writes.
func SampleUserBucket(id string, at time.Time) UserBucket {
	return UserBucket{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Source:   "seeded by user.SampleUserBucket",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserLaneSchema is the stored shape of a UserLane.
var UserLaneSchema = schema.New("user.userlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserLane is one user lane.
type UserLane struct {
	ID        string
	Owner     string
	Name      string
	Reference string
	Count     int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UserLane) Key() string { return "user/userlane/" + x.ID }

// Fields encodes x for storage.
func (x UserLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"name":      x.Name,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserLane decodes a stored record.
func ParseUserLane(f codec.Fields) (UserLane, error) {
	var x UserLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userlane: missing id")
	}
	x.Owner = f["owner"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("user: userlane %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: userlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserLane) String() string {
	return fmt.Sprintf("UserLane(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserLane) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserLane) Normalize() UserLane {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortUserLanes orders by count descending, then by ID.
func SortUserLanes(xs []UserLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserLanes drops the archived entries, preserving order.
func ActiveUserLanes(xs []UserLane) []UserLane {
	out := make([]UserLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserLanes adds up the count of every active entry.
func SumUserLanes(xs []UserLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexUserLanes keys the entries by ID, last write winning.
func IndexUserLanes(xs []UserLane) map[string]UserLane {
	out := make(map[string]UserLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserLane builds a valid UserLane, for tests and for the seed data the cron
// binary writes.
func SampleUserLane(id string, at time.Time) UserLane {
	return UserLane{
		ID:        id,
		Owner:     "owner-" + id,
		Name:      "name-" + id,
		Reference: "seeded by user.SampleUserLane",
		Count:     len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UserSlotSchema is the stored shape of a UserSlot.
var UserSlotSchema = schema.New("user.userslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserSlot is one user slot.
type UserSlot struct {
	ID       string
	Group    string
	Source   string
	Category string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserSlot) Key() string { return "user/userslot/" + x.ID }

// Fields encodes x for storage.
func (x UserSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserSlot decodes a stored record.
func ParseUserSlot(f codec.Fields) (UserSlot, error) {
	var x UserSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userslot: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("user: userslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("user: userslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserSlot) String() string {
	return fmt.Sprintf("UserSlot(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserSlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserSlot) Normalize() UserSlot {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortUserSlots orders by count descending, then by ID.
func SortUserSlots(xs []UserSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserSlots drops the archived entries, preserving order.
func ActiveUserSlots(xs []UserSlot) []UserSlot {
	out := make([]UserSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserSlots adds up the count of every active entry.
func SumUserSlots(xs []UserSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexUserSlots keys the entries by ID, last write winning.
func IndexUserSlots(xs []UserSlot) map[string]UserSlot {
	out := make(map[string]UserSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserSlot builds a valid UserSlot, for tests and for the seed data the cron
// binary writes.
func SampleUserSlot(id string, at time.Time) UserSlot {
	return UserSlot{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Category: "seeded by user.SampleUserSlot",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserTagSchema is the stored shape of a UserTag.
var UserTagSchema = schema.New("user.usertag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserTag is one user tag.
type UserTag struct {
	ID        string
	Owner     string
	Group     string
	Reference string
	Offset    int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UserTag) Key() string { return "user/usertag/" + x.ID }

// Fields encodes x for storage.
func (x UserTag) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"group":     x.Group,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserTag decodes a stored record.
func ParseUserTag(f codec.Fields) (UserTag, error) {
	var x UserTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: usertag: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: usertag %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("user: usertag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: usertag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: usertag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserTag) String() string {
	return fmt.Sprintf("UserTag(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserTag) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserTag) Normalize() UserTag {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortUserTags orders by offset descending, then by ID.
func SortUserTags(xs []UserTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserTags drops the archived entries, preserving order.
func ActiveUserTags(xs []UserTag) []UserTag {
	out := make([]UserTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserTags adds up the offset of every active entry.
func SumUserTags(xs []UserTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUserTags keys the entries by ID, last write winning.
func IndexUserTags(xs []UserTag) map[string]UserTag {
	out := make(map[string]UserTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserTag builds a valid UserTag, for tests and for the seed data the cron
// binary writes.
func SampleUserTag(id string, at time.Time) UserTag {
	return UserTag{
		ID:        id,
		Owner:     "owner-" + id,
		Group:     "group-" + id,
		Reference: "seeded by user.SampleUserTag",
		Offset:    len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UserRuleSchema is the stored shape of a UserRule.
var UserRuleSchema = schema.New("user.userrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserRule is one user rule.
type UserRule struct {
	ID        string
	Name      string
	Reference string
	Category  string
	Offset    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UserRule) Key() string { return "user/userrule/" + x.ID }

// Fields encodes x for storage.
func (x UserRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"category":  x.Category,
		"offset":    strconv.Itoa(x.Offset),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserRule decodes a stored record.
func ParseUserRule(f codec.Fields) (UserRule, error) {
	var x UserRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userrule: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("user: userrule %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("user: userrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserRule) String() string {
	return fmt.Sprintf("UserRule(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserRule) Normalize() UserRule {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortUserRules orders by offset descending, then by ID.
func SortUserRules(xs []UserRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserRules drops the archived entries, preserving order.
func ActiveUserRules(xs []UserRule) []UserRule {
	out := make([]UserRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserRules adds up the offset of every active entry.
func SumUserRules(xs []UserRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUserRules keys the entries by ID, last write winning.
func IndexUserRules(xs []UserRule) map[string]UserRule {
	out := make(map[string]UserRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserRule builds a valid UserRule, for tests and for the seed data the cron
// binary writes.
func SampleUserRule(id string, at time.Time) UserRule {
	return UserRule{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by user.SampleUserRule",
		Offset:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UserHoldSchema is the stored shape of a UserHold.
var UserHoldSchema = schema.New("user.userhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserHold is one user hold.
type UserHold struct {
	ID       string
	Category string
	Label    string
	Source   string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserHold) Key() string { return "user/userhold/" + x.ID }

// Fields encodes x for storage.
func (x UserHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserHold decodes a stored record.
func ParseUserHold(f codec.Fields) (UserHold, error) {
	var x UserHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userhold: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("user: userhold %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("user: userhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserHold) String() string {
	return fmt.Sprintf("UserHold(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserHold) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserHold) Normalize() UserHold {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortUserHolds orders by version descending, then by ID.
func SortUserHolds(xs []UserHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserHolds drops the archived entries, preserving order.
func ActiveUserHolds(xs []UserHold) []UserHold {
	out := make([]UserHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserHolds adds up the version of every active entry.
func SumUserHolds(xs []UserHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexUserHolds keys the entries by ID, last write winning.
func IndexUserHolds(xs []UserHold) map[string]UserHold {
	out := make(map[string]UserHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserHold builds a valid UserHold, for tests and for the seed data the cron
// binary writes.
func SampleUserHold(id string, at time.Time) UserHold {
	return UserHold{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Source:   "seeded by user.SampleUserHold",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserClaimSchema is the stored shape of a UserClaim.
var UserClaimSchema = schema.New("user.userclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserClaim is one user claim.
type UserClaim struct {
	ID       string
	Owner    string
	Comment  string
	Category string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserClaim) Key() string { return "user/userclaim/" + x.ID }

// Fields encodes x for storage.
func (x UserClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"category": x.Category,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserClaim decodes a stored record.
func ParseUserClaim(f codec.Fields) (UserClaim, error) {
	var x UserClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: userclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Category = f["category"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("user: userclaim %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("user: userclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: userclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: userclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserClaim) String() string {
	return fmt.Sprintf("UserClaim(%s, owner=%q, weight=%d, archived=%t)", x.ID, x.Owner, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserClaim) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserClaim) Normalize() UserClaim {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortUserClaims orders by weight descending, then by ID.
func SortUserClaims(xs []UserClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserClaims drops the archived entries, preserving order.
func ActiveUserClaims(xs []UserClaim) []UserClaim {
	out := make([]UserClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserClaims adds up the weight of every active entry.
func SumUserClaims(xs []UserClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexUserClaims keys the entries by ID, last write winning.
func IndexUserClaims(xs []UserClaim) map[string]UserClaim {
	out := make(map[string]UserClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserClaim builds a valid UserClaim, for tests and for the seed data the cron
// binary writes.
func SampleUserClaim(id string, at time.Time) UserClaim {
	return UserClaim{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Category: "seeded by user.SampleUserClaim",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UserCursorSchema is the stored shape of a UserCursor.
var UserCursorSchema = schema.New("user.usercursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UserCursor is one user cursor.
type UserCursor struct {
	ID       string
	Owner    string
	Comment  string
	Region   string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UserCursor) Key() string { return "user/usercursor/" + x.ID }

// Fields encodes x for storage.
func (x UserCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUserCursor decodes a stored record.
func ParseUserCursor(f codec.Fields) (UserCursor, error) {
	var x UserCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("user: usercursor: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("user: usercursor %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("user: usercursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("user: usercursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("user: usercursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UserCursor) String() string {
	return fmt.Sprintf("UserCursor(%s, owner=%q, weight=%d, archived=%t)", x.ID, x.Owner, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UserCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UserCursor) Normalize() UserCursor {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortUserCursors orders by weight descending, then by ID.
func SortUserCursors(xs []UserCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUserCursors drops the archived entries, preserving order.
func ActiveUserCursors(xs []UserCursor) []UserCursor {
	out := make([]UserCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUserCursors adds up the weight of every active entry.
func SumUserCursors(xs []UserCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexUserCursors keys the entries by ID, last write winning.
func IndexUserCursors(xs []UserCursor) map[string]UserCursor {
	out := make(map[string]UserCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUserCursor builds a valid UserCursor, for tests and for the seed data the cron
// binary writes.
func SampleUserCursor(id string, at time.Time) UserCursor {
	return UserCursor{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by user.SampleUserCursor",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	UserRecordSchema,
	UserEntrySchema,
	UserItemSchema,
	UserDraftSchema,
	UserRevisionSchema,
	UserSnapshotSchema,
	UserBatchSchema,
	UserRequestSchema,
	UserSummarySchema,
	UserPolicySchema,
	UserLinkSchema,
	UserNoteSchema,
	UserWindowSchema,
	UserQuotaSchema,
	UserBucketSchema,
	UserLaneSchema,
	UserSlotSchema,
	UserTagSchema,
	UserRuleSchema,
	UserHoldSchema,
	UserClaimSchema,
	UserCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"user/userrecord/",
	"user/userentry/",
	"user/useritem/",
	"user/userdraft/",
	"user/userrevision/",
	"user/usersnapshot/",
	"user/userbatch/",
	"user/userrequest/",
	"user/usersummary/",
	"user/userpolicy/",
	"user/userlink/",
	"user/usernote/",
	"user/userwindow/",
	"user/userquota/",
	"user/userbucket/",
	"user/userlane/",
	"user/userslot/",
	"user/usertag/",
	"user/userrule/",
	"user/userhold/",
	"user/userclaim/",
	"user/usercursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "user: " + strings.Join(names, ", ")
}
