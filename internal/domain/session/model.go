// Package session holds the session domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package session

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// SessionRecordSchema is the stored shape of a SessionRecord.
var SessionRecordSchema = schema.New("session.sessionrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionRecord is one session record.
type SessionRecord struct {
	ID       string
	Source   string
	Name     string
	Label    string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionRecord) Key() string { return "session/sessionrecord/" + x.ID }

// Fields encodes x for storage.
func (x SessionRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionRecord decodes a stored record.
func ParseSessionRecord(f codec.Fields) (SessionRecord, error) {
	var x SessionRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionrecord: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("session: sessionrecord %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionRecord) String() string {
	return fmt.Sprintf("SessionRecord(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionRecord) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionRecord) Normalize() SessionRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSessionRecords orders by priority descending, then by ID.
func SortSessionRecords(xs []SessionRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionRecords drops the archived entries, preserving order.
func ActiveSessionRecords(xs []SessionRecord) []SessionRecord {
	out := make([]SessionRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionRecords adds up the priority of every active entry.
func SumSessionRecords(xs []SessionRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSessionRecords keys the entries by ID, last write winning.
func IndexSessionRecords(xs []SessionRecord) map[string]SessionRecord {
	out := make(map[string]SessionRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionRecord builds a valid SessionRecord, for tests and for the seed data the cron
// binary writes.
func SampleSessionRecord(id string, at time.Time) SessionRecord {
	return SessionRecord{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Label:    "seeded by session.SampleSessionRecord",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionEntrySchema is the stored shape of a SessionEntry.
var SessionEntrySchema = schema.New("session.sessionentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionEntry is one session entry.
type SessionEntry struct {
	ID       string
	Name     string
	Channel  string
	Label    string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionEntry) Key() string { return "session/sessionentry/" + x.ID }

// Fields encodes x for storage.
func (x SessionEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionEntry decodes a stored record.
func ParseSessionEntry(f codec.Fields) (SessionEntry, error) {
	var x SessionEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionentry: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("session: sessionentry %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionEntry) String() string {
	return fmt.Sprintf("SessionEntry(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionEntry) Normalize() SessionEntry {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSessionEntrys orders by offset descending, then by ID.
func SortSessionEntrys(xs []SessionEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionEntrys drops the archived entries, preserving order.
func ActiveSessionEntrys(xs []SessionEntry) []SessionEntry {
	out := make([]SessionEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionEntrys adds up the offset of every active entry.
func SumSessionEntrys(xs []SessionEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSessionEntrys keys the entries by ID, last write winning.
func IndexSessionEntrys(xs []SessionEntry) map[string]SessionEntry {
	out := make(map[string]SessionEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionEntry builds a valid SessionEntry, for tests and for the seed data the cron
// binary writes.
func SampleSessionEntry(id string, at time.Time) SessionEntry {
	return SessionEntry{
		ID:       id,
		Name:     "name-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by session.SampleSessionEntry",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionItemSchema is the stored shape of a SessionItem.
var SessionItemSchema = schema.New("session.sessionitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionItem is one session item.
type SessionItem struct {
	ID       string
	Name     string
	Group    string
	Label    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionItem) Key() string { return "session/sessionitem/" + x.ID }

// Fields encodes x for storage.
func (x SessionItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"group":    x.Group,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionItem decodes a stored record.
func ParseSessionItem(f codec.Fields) (SessionItem, error) {
	var x SessionItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionitem: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("session: sessionitem %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("session: sessionitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionItem) String() string {
	return fmt.Sprintf("SessionItem(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionItem) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionItem) Normalize() SessionItem {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSessionItems orders by count descending, then by ID.
func SortSessionItems(xs []SessionItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionItems drops the archived entries, preserving order.
func ActiveSessionItems(xs []SessionItem) []SessionItem {
	out := make([]SessionItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionItems adds up the count of every active entry.
func SumSessionItems(xs []SessionItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSessionItems keys the entries by ID, last write winning.
func IndexSessionItems(xs []SessionItem) map[string]SessionItem {
	out := make(map[string]SessionItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionItem builds a valid SessionItem, for tests and for the seed data the cron
// binary writes.
func SampleSessionItem(id string, at time.Time) SessionItem {
	return SessionItem{
		ID:      id,
		Name:    "name-" + id,
		Group:   "group-" + id,
		Label:   "seeded by session.SampleSessionItem",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SessionDraftSchema is the stored shape of a SessionDraft.
var SessionDraftSchema = schema.New("session.sessiondraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionDraft is one session draft.
type SessionDraft struct {
	ID       string
	Group    string
	Label    string
	Comment  string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionDraft) Key() string { return "session/sessiondraft/" + x.ID }

// Fields encodes x for storage.
func (x SessionDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"label":    x.Label,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionDraft decodes a stored record.
func ParseSessionDraft(f codec.Fields) (SessionDraft, error) {
	var x SessionDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessiondraft: missing id")
	}
	x.Group = f["group"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("session: sessiondraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("session: sessiondraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessiondraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessiondraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionDraft) String() string {
	return fmt.Sprintf("SessionDraft(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionDraft) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionDraft) Normalize() SessionDraft {
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSessionDrafts orders by count descending, then by ID.
func SortSessionDrafts(xs []SessionDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionDrafts drops the archived entries, preserving order.
func ActiveSessionDrafts(xs []SessionDraft) []SessionDraft {
	out := make([]SessionDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionDrafts adds up the count of every active entry.
func SumSessionDrafts(xs []SessionDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSessionDrafts keys the entries by ID, last write winning.
func IndexSessionDrafts(xs []SessionDraft) map[string]SessionDraft {
	out := make(map[string]SessionDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionDraft builds a valid SessionDraft, for tests and for the seed data the cron
// binary writes.
func SampleSessionDraft(id string, at time.Time) SessionDraft {
	return SessionDraft{
		ID:      id,
		Group:   "group-" + id,
		Label:   "label-" + id,
		Comment: "seeded by session.SampleSessionDraft",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SessionRevisionSchema is the stored shape of a SessionRevision.
var SessionRevisionSchema = schema.New("session.sessionrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionRevision is one session revision.
type SessionRevision struct {
	ID        string
	Name      string
	Reference string
	Channel   string
	Priority  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SessionRevision) Key() string { return "session/sessionrevision/" + x.ID }

// Fields encodes x for storage.
func (x SessionRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"channel":   x.Channel,
		"priority":  strconv.Itoa(x.Priority),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionRevision decodes a stored record.
func ParseSessionRevision(f codec.Fields) (SessionRevision, error) {
	var x SessionRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionrevision: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("session: sessionrevision %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionRevision) String() string {
	return fmt.Sprintf("SessionRevision(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionRevision) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionRevision) Normalize() SessionRevision {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSessionRevisions orders by priority descending, then by ID.
func SortSessionRevisions(xs []SessionRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionRevisions drops the archived entries, preserving order.
func ActiveSessionRevisions(xs []SessionRevision) []SessionRevision {
	out := make([]SessionRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionRevisions adds up the priority of every active entry.
func SumSessionRevisions(xs []SessionRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSessionRevisions keys the entries by ID, last write winning.
func IndexSessionRevisions(xs []SessionRevision) map[string]SessionRevision {
	out := make(map[string]SessionRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionRevision builds a valid SessionRevision, for tests and for the seed data the cron
// binary writes.
func SampleSessionRevision(id string, at time.Time) SessionRevision {
	return SessionRevision{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by session.SampleSessionRevision",
		Priority:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SessionSnapshotSchema is the stored shape of a SessionSnapshot.
var SessionSnapshotSchema = schema.New("session.sessionsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionSnapshot is one session snapshot.
type SessionSnapshot struct {
	ID       string
	Group    string
	Owner    string
	Category string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionSnapshot) Key() string { return "session/sessionsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x SessionSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"category": x.Category,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionSnapshot decodes a stored record.
func ParseSessionSnapshot(f codec.Fields) (SessionSnapshot, error) {
	var x SessionSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionsnapshot: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("session: sessionsnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionSnapshot) String() string {
	return fmt.Sprintf("SessionSnapshot(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionSnapshot) Normalize() SessionSnapshot {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSessionSnapshots orders by priority descending, then by ID.
func SortSessionSnapshots(xs []SessionSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionSnapshots drops the archived entries, preserving order.
func ActiveSessionSnapshots(xs []SessionSnapshot) []SessionSnapshot {
	out := make([]SessionSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionSnapshots adds up the priority of every active entry.
func SumSessionSnapshots(xs []SessionSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSessionSnapshots keys the entries by ID, last write winning.
func IndexSessionSnapshots(xs []SessionSnapshot) map[string]SessionSnapshot {
	out := make(map[string]SessionSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionSnapshot builds a valid SessionSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleSessionSnapshot(id string, at time.Time) SessionSnapshot {
	return SessionSnapshot{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by session.SampleSessionSnapshot",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionBatchSchema is the stored shape of a SessionBatch.
var SessionBatchSchema = schema.New("session.sessionbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionBatch is one session batch.
type SessionBatch struct {
	ID        string
	Reference string
	Label     string
	Channel   string
	Capacity  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SessionBatch) Key() string { return "session/sessionbatch/" + x.ID }

// Fields encodes x for storage.
func (x SessionBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"channel":   x.Channel,
		"capacity":  strconv.Itoa(x.Capacity),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionBatch decodes a stored record.
func ParseSessionBatch(f codec.Fields) (SessionBatch, error) {
	var x SessionBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionbatch: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionbatch %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("session: sessionbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionBatch) String() string {
	return fmt.Sprintf("SessionBatch(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionBatch) Normalize() SessionBatch {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortSessionBatchs orders by capacity descending, then by ID.
func SortSessionBatchs(xs []SessionBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionBatchs drops the archived entries, preserving order.
func ActiveSessionBatchs(xs []SessionBatch) []SessionBatch {
	out := make([]SessionBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionBatchs adds up the capacity of every active entry.
func SumSessionBatchs(xs []SessionBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSessionBatchs keys the entries by ID, last write winning.
func IndexSessionBatchs(xs []SessionBatch) map[string]SessionBatch {
	out := make(map[string]SessionBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionBatch builds a valid SessionBatch, for tests and for the seed data the cron
// binary writes.
func SampleSessionBatch(id string, at time.Time) SessionBatch {
	return SessionBatch{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Channel:   "seeded by session.SampleSessionBatch",
		Capacity:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SessionRequestSchema is the stored shape of a SessionRequest.
var SessionRequestSchema = schema.New("session.sessionrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionRequest is one session request.
type SessionRequest struct {
	ID        string
	Source    string
	Owner     string
	Reference string
	Version   int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SessionRequest) Key() string { return "session/sessionrequest/" + x.ID }

// Fields encodes x for storage.
func (x SessionRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"owner":     x.Owner,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionRequest decodes a stored record.
func ParseSessionRequest(f codec.Fields) (SessionRequest, error) {
	var x SessionRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionrequest: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("session: sessionrequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("session: sessionrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionRequest) String() string {
	return fmt.Sprintf("SessionRequest(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionRequest) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionRequest) Normalize() SessionRequest {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortSessionRequests orders by version descending, then by ID.
func SortSessionRequests(xs []SessionRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionRequests drops the archived entries, preserving order.
func ActiveSessionRequests(xs []SessionRequest) []SessionRequest {
	out := make([]SessionRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionRequests adds up the version of every active entry.
func SumSessionRequests(xs []SessionRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSessionRequests keys the entries by ID, last write winning.
func IndexSessionRequests(xs []SessionRequest) map[string]SessionRequest {
	out := make(map[string]SessionRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionRequest builds a valid SessionRequest, for tests and for the seed data the cron
// binary writes.
func SampleSessionRequest(id string, at time.Time) SessionRequest {
	return SessionRequest{
		ID:        id,
		Source:    "source-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by session.SampleSessionRequest",
		Version:   len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SessionSummarySchema is the stored shape of a SessionSummary.
var SessionSummarySchema = schema.New("session.sessionsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionSummary is one session summary.
type SessionSummary struct {
	ID       string
	Owner    string
	Source   string
	Comment  string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionSummary) Key() string { return "session/sessionsummary/" + x.ID }

// Fields encodes x for storage.
func (x SessionSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionSummary decodes a stored record.
func ParseSessionSummary(f codec.Fields) (SessionSummary, error) {
	var x SessionSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionsummary: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionsummary %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("session: sessionsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionSummary) String() string {
	return fmt.Sprintf("SessionSummary(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionSummary) Normalize() SessionSummary {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortSessionSummarys orders by capacity descending, then by ID.
func SortSessionSummarys(xs []SessionSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionSummarys drops the archived entries, preserving order.
func ActiveSessionSummarys(xs []SessionSummary) []SessionSummary {
	out := make([]SessionSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionSummarys adds up the capacity of every active entry.
func SumSessionSummarys(xs []SessionSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSessionSummarys keys the entries by ID, last write winning.
func IndexSessionSummarys(xs []SessionSummary) map[string]SessionSummary {
	out := make(map[string]SessionSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionSummary builds a valid SessionSummary, for tests and for the seed data the cron
// binary writes.
func SampleSessionSummary(id string, at time.Time) SessionSummary {
	return SessionSummary{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Comment:  "seeded by session.SampleSessionSummary",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionPolicySchema is the stored shape of a SessionPolicy.
var SessionPolicySchema = schema.New("session.sessionpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionPolicy is one session policy.
type SessionPolicy struct {
	ID        string
	Channel   string
	Reference string
	Group     string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SessionPolicy) Key() string { return "session/sessionpolicy/" + x.ID }

// Fields encodes x for storage.
func (x SessionPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"group":     x.Group,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionPolicy decodes a stored record.
func ParseSessionPolicy(f codec.Fields) (SessionPolicy, error) {
	var x SessionPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionpolicy: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("session: sessionpolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("session: sessionpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionPolicy) String() string {
	return fmt.Sprintf("SessionPolicy(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionPolicy) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionPolicy) Normalize() SessionPolicy {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortSessionPolicys orders by weight descending, then by ID.
func SortSessionPolicys(xs []SessionPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionPolicys drops the archived entries, preserving order.
func ActiveSessionPolicys(xs []SessionPolicy) []SessionPolicy {
	out := make([]SessionPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionPolicys adds up the weight of every active entry.
func SumSessionPolicys(xs []SessionPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexSessionPolicys keys the entries by ID, last write winning.
func IndexSessionPolicys(xs []SessionPolicy) map[string]SessionPolicy {
	out := make(map[string]SessionPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionPolicy builds a valid SessionPolicy, for tests and for the seed data the cron
// binary writes.
func SampleSessionPolicy(id string, at time.Time) SessionPolicy {
	return SessionPolicy{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Group:     "seeded by session.SampleSessionPolicy",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SessionLinkSchema is the stored shape of a SessionLink.
var SessionLinkSchema = schema.New("session.sessionlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionLink is one session link.
type SessionLink struct {
	ID       string
	Region   string
	Name     string
	Source   string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionLink) Key() string { return "session/sessionlink/" + x.ID }

// Fields encodes x for storage.
func (x SessionLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionLink decodes a stored record.
func ParseSessionLink(f codec.Fields) (SessionLink, error) {
	var x SessionLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionlink: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("session: sessionlink %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("session: sessionlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionLink) String() string {
	return fmt.Sprintf("SessionLink(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionLink) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionLink) Normalize() SessionLink {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSessionLinks orders by attempts descending, then by ID.
func SortSessionLinks(xs []SessionLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionLinks drops the archived entries, preserving order.
func ActiveSessionLinks(xs []SessionLink) []SessionLink {
	out := make([]SessionLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionLinks adds up the attempts of every active entry.
func SumSessionLinks(xs []SessionLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexSessionLinks keys the entries by ID, last write winning.
func IndexSessionLinks(xs []SessionLink) map[string]SessionLink {
	out := make(map[string]SessionLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionLink builds a valid SessionLink, for tests and for the seed data the cron
// binary writes.
func SampleSessionLink(id string, at time.Time) SessionLink {
	return SessionLink{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Source:   "seeded by session.SampleSessionLink",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionNoteSchema is the stored shape of a SessionNote.
var SessionNoteSchema = schema.New("session.sessionnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionNote is one session note.
type SessionNote struct {
	ID       string
	Name     string
	Source   string
	Owner    string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionNote) Key() string { return "session/sessionnote/" + x.ID }

// Fields encodes x for storage.
func (x SessionNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionNote decodes a stored record.
func ParseSessionNote(f codec.Fields) (SessionNote, error) {
	var x SessionNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionnote: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionnote %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("session: sessionnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionNote) String() string {
	return fmt.Sprintf("SessionNote(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionNote) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionNote) Normalize() SessionNote {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSessionNotes orders by capacity descending, then by ID.
func SortSessionNotes(xs []SessionNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionNotes drops the archived entries, preserving order.
func ActiveSessionNotes(xs []SessionNote) []SessionNote {
	out := make([]SessionNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionNotes adds up the capacity of every active entry.
func SumSessionNotes(xs []SessionNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSessionNotes keys the entries by ID, last write winning.
func IndexSessionNotes(xs []SessionNote) map[string]SessionNote {
	out := make(map[string]SessionNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionNote builds a valid SessionNote, for tests and for the seed data the cron
// binary writes.
func SampleSessionNote(id string, at time.Time) SessionNote {
	return SessionNote{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by session.SampleSessionNote",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionWindowSchema is the stored shape of a SessionWindow.
var SessionWindowSchema = schema.New("session.sessionwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionWindow is one session window.
type SessionWindow struct {
	ID       string
	Name     string
	Category string
	Owner    string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionWindow) Key() string { return "session/sessionwindow/" + x.ID }

// Fields encodes x for storage.
func (x SessionWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionWindow decodes a stored record.
func ParseSessionWindow(f codec.Fields) (SessionWindow, error) {
	var x SessionWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionwindow: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("session: sessionwindow %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("session: sessionwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionWindow) String() string {
	return fmt.Sprintf("SessionWindow(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionWindow) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionWindow) Normalize() SessionWindow {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSessionWindows orders by offset descending, then by ID.
func SortSessionWindows(xs []SessionWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionWindows drops the archived entries, preserving order.
func ActiveSessionWindows(xs []SessionWindow) []SessionWindow {
	out := make([]SessionWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionWindows adds up the offset of every active entry.
func SumSessionWindows(xs []SessionWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSessionWindows keys the entries by ID, last write winning.
func IndexSessionWindows(xs []SessionWindow) map[string]SessionWindow {
	out := make(map[string]SessionWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionWindow builds a valid SessionWindow, for tests and for the seed data the cron
// binary writes.
func SampleSessionWindow(id string, at time.Time) SessionWindow {
	return SessionWindow{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Owner:    "seeded by session.SampleSessionWindow",
		Offset:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionQuotaSchema is the stored shape of a SessionQuota.
var SessionQuotaSchema = schema.New("session.sessionquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionQuota is one session quota.
type SessionQuota struct {
	ID       string
	Source   string
	Channel  string
	Comment  string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionQuota) Key() string { return "session/sessionquota/" + x.ID }

// Fields encodes x for storage.
func (x SessionQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionQuota decodes a stored record.
func ParseSessionQuota(f codec.Fields) (SessionQuota, error) {
	var x SessionQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionquota: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionquota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("session: sessionquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionQuota) String() string {
	return fmt.Sprintf("SessionQuota(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionQuota) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionQuota) Normalize() SessionQuota {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSessionQuotas orders by capacity descending, then by ID.
func SortSessionQuotas(xs []SessionQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionQuotas drops the archived entries, preserving order.
func ActiveSessionQuotas(xs []SessionQuota) []SessionQuota {
	out := make([]SessionQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionQuotas adds up the capacity of every active entry.
func SumSessionQuotas(xs []SessionQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSessionQuotas keys the entries by ID, last write winning.
func IndexSessionQuotas(xs []SessionQuota) map[string]SessionQuota {
	out := make(map[string]SessionQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionQuota builds a valid SessionQuota, for tests and for the seed data the cron
// binary writes.
func SampleSessionQuota(id string, at time.Time) SessionQuota {
	return SessionQuota{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by session.SampleSessionQuota",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionBucketSchema is the stored shape of a SessionBucket.
var SessionBucketSchema = schema.New("session.sessionbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionBucket is one session bucket.
type SessionBucket struct {
	ID        string
	Group     string
	Reference string
	Category  string
	Retained  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SessionBucket) Key() string { return "session/sessionbucket/" + x.ID }

// Fields encodes x for storage.
func (x SessionBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"category":  x.Category,
		"retained":  strconv.Itoa(x.Retained),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionBucket decodes a stored record.
func ParseSessionBucket(f codec.Fields) (SessionBucket, error) {
	var x SessionBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionbucket: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("session: sessionbucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("session: sessionbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionBucket) String() string {
	return fmt.Sprintf("SessionBucket(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionBucket) Normalize() SessionBucket {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSessionBuckets orders by retained descending, then by ID.
func SortSessionBuckets(xs []SessionBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionBuckets drops the archived entries, preserving order.
func ActiveSessionBuckets(xs []SessionBucket) []SessionBucket {
	out := make([]SessionBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionBuckets adds up the retained of every active entry.
func SumSessionBuckets(xs []SessionBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexSessionBuckets keys the entries by ID, last write winning.
func IndexSessionBuckets(xs []SessionBucket) map[string]SessionBucket {
	out := make(map[string]SessionBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionBucket builds a valid SessionBucket, for tests and for the seed data the cron
// binary writes.
func SampleSessionBucket(id string, at time.Time) SessionBucket {
	return SessionBucket{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by session.SampleSessionBucket",
		Retained:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SessionLaneSchema is the stored shape of a SessionLane.
var SessionLaneSchema = schema.New("session.sessionlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionLane is one session lane.
type SessionLane struct {
	ID       string
	Region   string
	Comment  string
	Owner    string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionLane) Key() string { return "session/sessionlane/" + x.ID }

// Fields encodes x for storage.
func (x SessionLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionLane decodes a stored record.
func ParseSessionLane(f codec.Fields) (SessionLane, error) {
	var x SessionLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionlane: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("session: sessionlane %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("session: sessionlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionLane) String() string {
	return fmt.Sprintf("SessionLane(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionLane) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionLane) Normalize() SessionLane {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSessionLanes orders by offset descending, then by ID.
func SortSessionLanes(xs []SessionLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionLanes drops the archived entries, preserving order.
func ActiveSessionLanes(xs []SessionLane) []SessionLane {
	out := make([]SessionLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionLanes adds up the offset of every active entry.
func SumSessionLanes(xs []SessionLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSessionLanes keys the entries by ID, last write winning.
func IndexSessionLanes(xs []SessionLane) map[string]SessionLane {
	out := make(map[string]SessionLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionLane builds a valid SessionLane, for tests and for the seed data the cron
// binary writes.
func SampleSessionLane(id string, at time.Time) SessionLane {
	return SessionLane{
		ID:       id,
		Region:   "region-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by session.SampleSessionLane",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionSlotSchema is the stored shape of a SessionSlot.
var SessionSlotSchema = schema.New("session.sessionslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionSlot is one session slot.
type SessionSlot struct {
	ID       string
	Category string
	Label    string
	Channel  string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionSlot) Key() string { return "session/sessionslot/" + x.ID }

// Fields encodes x for storage.
func (x SessionSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionSlot decodes a stored record.
func ParseSessionSlot(f codec.Fields) (SessionSlot, error) {
	var x SessionSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionslot: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("session: sessionslot %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("session: sessionslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionSlot) String() string {
	return fmt.Sprintf("SessionSlot(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionSlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionSlot) Normalize() SessionSlot {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSessionSlots orders by count descending, then by ID.
func SortSessionSlots(xs []SessionSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionSlots drops the archived entries, preserving order.
func ActiveSessionSlots(xs []SessionSlot) []SessionSlot {
	out := make([]SessionSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionSlots adds up the count of every active entry.
func SumSessionSlots(xs []SessionSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSessionSlots keys the entries by ID, last write winning.
func IndexSessionSlots(xs []SessionSlot) map[string]SessionSlot {
	out := make(map[string]SessionSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionSlot builds a valid SessionSlot, for tests and for the seed data the cron
// binary writes.
func SampleSessionSlot(id string, at time.Time) SessionSlot {
	return SessionSlot{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by session.SampleSessionSlot",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionTagSchema is the stored shape of a SessionTag.
var SessionTagSchema = schema.New("session.sessiontag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionTag is one session tag.
type SessionTag struct {
	ID       string
	Region   string
	Name     string
	Category string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionTag) Key() string { return "session/sessiontag/" + x.ID }

// Fields encodes x for storage.
func (x SessionTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionTag decodes a stored record.
func ParseSessionTag(f codec.Fields) (SessionTag, error) {
	var x SessionTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessiontag: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("session: sessiontag %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("session: sessiontag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessiontag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessiontag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionTag) String() string {
	return fmt.Sprintf("SessionTag(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionTag) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionTag) Normalize() SessionTag {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSessionTags orders by version descending, then by ID.
func SortSessionTags(xs []SessionTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionTags drops the archived entries, preserving order.
func ActiveSessionTags(xs []SessionTag) []SessionTag {
	out := make([]SessionTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionTags adds up the version of every active entry.
func SumSessionTags(xs []SessionTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSessionTags keys the entries by ID, last write winning.
func IndexSessionTags(xs []SessionTag) map[string]SessionTag {
	out := make(map[string]SessionTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionTag builds a valid SessionTag, for tests and for the seed data the cron
// binary writes.
func SampleSessionTag(id string, at time.Time) SessionTag {
	return SessionTag{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Category: "seeded by session.SampleSessionTag",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionRuleSchema is the stored shape of a SessionRule.
var SessionRuleSchema = schema.New("session.sessionrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionRule is one session rule.
type SessionRule struct {
	ID       string
	Region   string
	Name     string
	Label    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionRule) Key() string { return "session/sessionrule/" + x.ID }

// Fields encodes x for storage.
func (x SessionRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionRule decodes a stored record.
func ParseSessionRule(f codec.Fields) (SessionRule, error) {
	var x SessionRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionrule: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("session: sessionrule %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("session: sessionrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionRule) String() string {
	return fmt.Sprintf("SessionRule(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionRule) Normalize() SessionRule {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSessionRules orders by offset descending, then by ID.
func SortSessionRules(xs []SessionRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionRules drops the archived entries, preserving order.
func ActiveSessionRules(xs []SessionRule) []SessionRule {
	out := make([]SessionRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionRules adds up the offset of every active entry.
func SumSessionRules(xs []SessionRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSessionRules keys the entries by ID, last write winning.
func IndexSessionRules(xs []SessionRule) map[string]SessionRule {
	out := make(map[string]SessionRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionRule builds a valid SessionRule, for tests and for the seed data the cron
// binary writes.
func SampleSessionRule(id string, at time.Time) SessionRule {
	return SessionRule{
		ID:      id,
		Region:  "region-" + id,
		Name:    "name-" + id,
		Label:   "seeded by session.SampleSessionRule",
		Offset:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SessionHoldSchema is the stored shape of a SessionHold.
var SessionHoldSchema = schema.New("session.sessionhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionHold is one session hold.
type SessionHold struct {
	ID       string
	Label    string
	Name     string
	Source   string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionHold) Key() string { return "session/sessionhold/" + x.ID }

// Fields encodes x for storage.
func (x SessionHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"name":     x.Name,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionHold decodes a stored record.
func ParseSessionHold(f codec.Fields) (SessionHold, error) {
	var x SessionHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionhold: missing id")
	}
	x.Label = f["label"]
	x.Name = f["name"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("session: sessionhold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("session: sessionhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionHold) String() string {
	return fmt.Sprintf("SessionHold(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionHold) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionHold) Normalize() SessionHold {
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSessionHolds orders by offset descending, then by ID.
func SortSessionHolds(xs []SessionHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionHolds drops the archived entries, preserving order.
func ActiveSessionHolds(xs []SessionHold) []SessionHold {
	out := make([]SessionHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionHolds adds up the offset of every active entry.
func SumSessionHolds(xs []SessionHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSessionHolds keys the entries by ID, last write winning.
func IndexSessionHolds(xs []SessionHold) map[string]SessionHold {
	out := make(map[string]SessionHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionHold builds a valid SessionHold, for tests and for the seed data the cron
// binary writes.
func SampleSessionHold(id string, at time.Time) SessionHold {
	return SessionHold{
		ID:      id,
		Label:   "label-" + id,
		Name:    "name-" + id,
		Source:  "seeded by session.SampleSessionHold",
		Offset:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SessionClaimSchema is the stored shape of a SessionClaim.
var SessionClaimSchema = schema.New("session.sessionclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionClaim is one session claim.
type SessionClaim struct {
	ID       string
	Channel  string
	Region   string
	Name     string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionClaim) Key() string { return "session/sessionclaim/" + x.ID }

// Fields encodes x for storage.
func (x SessionClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"name":     x.Name,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionClaim decodes a stored record.
func ParseSessionClaim(f codec.Fields) (SessionClaim, error) {
	var x SessionClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessionclaim: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("session: sessionclaim %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("session: sessionclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessionclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessionclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionClaim) String() string {
	return fmt.Sprintf("SessionClaim(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionClaim) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionClaim) Normalize() SessionClaim {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSessionClaims orders by offset descending, then by ID.
func SortSessionClaims(xs []SessionClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionClaims drops the archived entries, preserving order.
func ActiveSessionClaims(xs []SessionClaim) []SessionClaim {
	out := make([]SessionClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionClaims adds up the offset of every active entry.
func SumSessionClaims(xs []SessionClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSessionClaims keys the entries by ID, last write winning.
func IndexSessionClaims(xs []SessionClaim) map[string]SessionClaim {
	out := make(map[string]SessionClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionClaim builds a valid SessionClaim, for tests and for the seed data the cron
// binary writes.
func SampleSessionClaim(id string, at time.Time) SessionClaim {
	return SessionClaim{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Name:     "seeded by session.SampleSessionClaim",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SessionCursorSchema is the stored shape of a SessionCursor.
var SessionCursorSchema = schema.New("session.sessioncursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SessionCursor is one session cursor.
type SessionCursor struct {
	ID       string
	Comment  string
	Label    string
	Owner    string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SessionCursor) Key() string { return "session/sessioncursor/" + x.ID }

// Fields encodes x for storage.
func (x SessionCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"label":    x.Label,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSessionCursor decodes a stored record.
func ParseSessionCursor(f codec.Fields) (SessionCursor, error) {
	var x SessionCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("session: sessioncursor: missing id")
	}
	x.Comment = f["comment"]
	x.Label = f["label"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("session: sessioncursor %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("session: sessioncursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("session: sessioncursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("session: sessioncursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SessionCursor) String() string {
	return fmt.Sprintf("SessionCursor(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SessionCursor) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SessionCursor) Normalize() SessionCursor {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortSessionCursors orders by offset descending, then by ID.
func SortSessionCursors(xs []SessionCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSessionCursors drops the archived entries, preserving order.
func ActiveSessionCursors(xs []SessionCursor) []SessionCursor {
	out := make([]SessionCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSessionCursors adds up the offset of every active entry.
func SumSessionCursors(xs []SessionCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSessionCursors keys the entries by ID, last write winning.
func IndexSessionCursors(xs []SessionCursor) map[string]SessionCursor {
	out := make(map[string]SessionCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSessionCursor builds a valid SessionCursor, for tests and for the seed data the cron
// binary writes.
func SampleSessionCursor(id string, at time.Time) SessionCursor {
	return SessionCursor{
		ID:       id,
		Comment:  "comment-" + id,
		Label:    "label-" + id,
		Owner:    "seeded by session.SampleSessionCursor",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	SessionRecordSchema,
	SessionEntrySchema,
	SessionItemSchema,
	SessionDraftSchema,
	SessionRevisionSchema,
	SessionSnapshotSchema,
	SessionBatchSchema,
	SessionRequestSchema,
	SessionSummarySchema,
	SessionPolicySchema,
	SessionLinkSchema,
	SessionNoteSchema,
	SessionWindowSchema,
	SessionQuotaSchema,
	SessionBucketSchema,
	SessionLaneSchema,
	SessionSlotSchema,
	SessionTagSchema,
	SessionRuleSchema,
	SessionHoldSchema,
	SessionClaimSchema,
	SessionCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"session/sessionrecord/",
	"session/sessionentry/",
	"session/sessionitem/",
	"session/sessiondraft/",
	"session/sessionrevision/",
	"session/sessionsnapshot/",
	"session/sessionbatch/",
	"session/sessionrequest/",
	"session/sessionsummary/",
	"session/sessionpolicy/",
	"session/sessionlink/",
	"session/sessionnote/",
	"session/sessionwindow/",
	"session/sessionquota/",
	"session/sessionbucket/",
	"session/sessionlane/",
	"session/sessionslot/",
	"session/sessiontag/",
	"session/sessionrule/",
	"session/sessionhold/",
	"session/sessionclaim/",
	"session/sessioncursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "session: " + strings.Join(names, ", ")
}
