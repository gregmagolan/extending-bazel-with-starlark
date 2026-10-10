// Package rollout holds the rollout domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package rollout

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// RolloutRecordSchema is the stored shape of a RolloutRecord.
var RolloutRecordSchema = schema.New("rollout.rolloutrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutRecord is one rollout record.
type RolloutRecord struct {
	ID       string
	Name     string
	Category string
	Channel  string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutRecord) Key() string { return "rollout/rolloutrecord/" + x.ID }

// Fields encodes x for storage.
func (x RolloutRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutRecord decodes a stored record.
func ParseRolloutRecord(f codec.Fields) (RolloutRecord, error) {
	var x RolloutRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutrecord: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrecord %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutRecord) String() string {
	return fmt.Sprintf("RolloutRecord(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutRecord) Normalize() RolloutRecord {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortRolloutRecords orders by capacity descending, then by ID.
func SortRolloutRecords(xs []RolloutRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutRecords drops the archived entries, preserving order.
func ActiveRolloutRecords(xs []RolloutRecord) []RolloutRecord {
	out := make([]RolloutRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutRecords adds up the capacity of every active entry.
func SumRolloutRecords(xs []RolloutRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRolloutRecords keys the entries by ID, last write winning.
func IndexRolloutRecords(xs []RolloutRecord) map[string]RolloutRecord {
	out := make(map[string]RolloutRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutRecord builds a valid RolloutRecord, for tests and for the seed data the cron
// binary writes.
func SampleRolloutRecord(id string, at time.Time) RolloutRecord {
	return RolloutRecord{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Channel:  "seeded by rollout.SampleRolloutRecord",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutEntrySchema is the stored shape of a RolloutEntry.
var RolloutEntrySchema = schema.New("rollout.rolloutentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutEntry is one rollout entry.
type RolloutEntry struct {
	ID       string
	Name     string
	Owner    string
	Group    string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutEntry) Key() string { return "rollout/rolloutentry/" + x.ID }

// Fields encodes x for storage.
func (x RolloutEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutEntry decodes a stored record.
func ParseRolloutEntry(f codec.Fields) (RolloutEntry, error) {
	var x RolloutEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutentry: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutentry %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutEntry) String() string {
	return fmt.Sprintf("RolloutEntry(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutEntry) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutEntry) Normalize() RolloutEntry {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortRolloutEntrys orders by attempts descending, then by ID.
func SortRolloutEntrys(xs []RolloutEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutEntrys drops the archived entries, preserving order.
func ActiveRolloutEntrys(xs []RolloutEntry) []RolloutEntry {
	out := make([]RolloutEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutEntrys adds up the attempts of every active entry.
func SumRolloutEntrys(xs []RolloutEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexRolloutEntrys keys the entries by ID, last write winning.
func IndexRolloutEntrys(xs []RolloutEntry) map[string]RolloutEntry {
	out := make(map[string]RolloutEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutEntry builds a valid RolloutEntry, for tests and for the seed data the cron
// binary writes.
func SampleRolloutEntry(id string, at time.Time) RolloutEntry {
	return RolloutEntry{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by rollout.SampleRolloutEntry",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutItemSchema is the stored shape of a RolloutItem.
var RolloutItemSchema = schema.New("rollout.rolloutitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutItem is one rollout item.
type RolloutItem struct {
	ID       string
	Label    string
	Comment  string
	Source   string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutItem) Key() string { return "rollout/rolloutitem/" + x.ID }

// Fields encodes x for storage.
func (x RolloutItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"comment":  x.Comment,
		"source":   x.Source,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutItem decodes a stored record.
func ParseRolloutItem(f codec.Fields) (RolloutItem, error) {
	var x RolloutItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutitem: missing id")
	}
	x.Label = f["label"]
	x.Comment = f["comment"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutItem) String() string {
	return fmt.Sprintf("RolloutItem(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutItem) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutItem) Normalize() RolloutItem {
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRolloutItems orders by count descending, then by ID.
func SortRolloutItems(xs []RolloutItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutItems drops the archived entries, preserving order.
func ActiveRolloutItems(xs []RolloutItem) []RolloutItem {
	out := make([]RolloutItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutItems adds up the count of every active entry.
func SumRolloutItems(xs []RolloutItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRolloutItems keys the entries by ID, last write winning.
func IndexRolloutItems(xs []RolloutItem) map[string]RolloutItem {
	out := make(map[string]RolloutItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutItem builds a valid RolloutItem, for tests and for the seed data the cron
// binary writes.
func SampleRolloutItem(id string, at time.Time) RolloutItem {
	return RolloutItem{
		ID:       id,
		Label:    "label-" + id,
		Comment:  "comment-" + id,
		Source:   "seeded by rollout.SampleRolloutItem",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutDraftSchema is the stored shape of a RolloutDraft.
var RolloutDraftSchema = schema.New("rollout.rolloutdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutDraft is one rollout draft.
type RolloutDraft struct {
	ID       string
	Comment  string
	Source   string
	Category string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutDraft) Key() string { return "rollout/rolloutdraft/" + x.ID }

// Fields encodes x for storage.
func (x RolloutDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"category": x.Category,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutDraft decodes a stored record.
func ParseRolloutDraft(f codec.Fields) (RolloutDraft, error) {
	var x RolloutDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutdraft: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutdraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutDraft) String() string {
	return fmt.Sprintf("RolloutDraft(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutDraft) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutDraft) Normalize() RolloutDraft {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRolloutDrafts orders by offset descending, then by ID.
func SortRolloutDrafts(xs []RolloutDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutDrafts drops the archived entries, preserving order.
func ActiveRolloutDrafts(xs []RolloutDraft) []RolloutDraft {
	out := make([]RolloutDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutDrafts adds up the offset of every active entry.
func SumRolloutDrafts(xs []RolloutDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRolloutDrafts keys the entries by ID, last write winning.
func IndexRolloutDrafts(xs []RolloutDraft) map[string]RolloutDraft {
	out := make(map[string]RolloutDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutDraft builds a valid RolloutDraft, for tests and for the seed data the cron
// binary writes.
func SampleRolloutDraft(id string, at time.Time) RolloutDraft {
	return RolloutDraft{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Category: "seeded by rollout.SampleRolloutDraft",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutRevisionSchema is the stored shape of a RolloutRevision.
var RolloutRevisionSchema = schema.New("rollout.rolloutrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutRevision is one rollout revision.
type RolloutRevision struct {
	ID        string
	Category  string
	Reference string
	Region    string
	Priority  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutRevision) Key() string { return "rollout/rolloutrevision/" + x.ID }

// Fields encodes x for storage.
func (x RolloutRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"region":    x.Region,
		"priority":  strconv.Itoa(x.Priority),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutRevision decodes a stored record.
func ParseRolloutRevision(f codec.Fields) (RolloutRevision, error) {
	var x RolloutRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutrevision: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrevision %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutRevision) String() string {
	return fmt.Sprintf("RolloutRevision(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutRevision) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutRevision) Normalize() RolloutRevision {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRolloutRevisions orders by priority descending, then by ID.
func SortRolloutRevisions(xs []RolloutRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutRevisions drops the archived entries, preserving order.
func ActiveRolloutRevisions(xs []RolloutRevision) []RolloutRevision {
	out := make([]RolloutRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutRevisions adds up the priority of every active entry.
func SumRolloutRevisions(xs []RolloutRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRolloutRevisions keys the entries by ID, last write winning.
func IndexRolloutRevisions(xs []RolloutRevision) map[string]RolloutRevision {
	out := make(map[string]RolloutRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutRevision builds a valid RolloutRevision, for tests and for the seed data the cron
// binary writes.
func SampleRolloutRevision(id string, at time.Time) RolloutRevision {
	return RolloutRevision{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by rollout.SampleRolloutRevision",
		Priority:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutSnapshotSchema is the stored shape of a RolloutSnapshot.
var RolloutSnapshotSchema = schema.New("rollout.rolloutsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutSnapshot is one rollout snapshot.
type RolloutSnapshot struct {
	ID       string
	Label    string
	Group    string
	Owner    string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutSnapshot) Key() string { return "rollout/rolloutsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x RolloutSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutSnapshot decodes a stored record.
func ParseRolloutSnapshot(f codec.Fields) (RolloutSnapshot, error) {
	var x RolloutSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutsnapshot: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutsnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutSnapshot) String() string {
	return fmt.Sprintf("RolloutSnapshot(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutSnapshot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutSnapshot) Normalize() RolloutSnapshot {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRolloutSnapshots orders by offset descending, then by ID.
func SortRolloutSnapshots(xs []RolloutSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutSnapshots drops the archived entries, preserving order.
func ActiveRolloutSnapshots(xs []RolloutSnapshot) []RolloutSnapshot {
	out := make([]RolloutSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutSnapshots adds up the offset of every active entry.
func SumRolloutSnapshots(xs []RolloutSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRolloutSnapshots keys the entries by ID, last write winning.
func IndexRolloutSnapshots(xs []RolloutSnapshot) map[string]RolloutSnapshot {
	out := make(map[string]RolloutSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutSnapshot builds a valid RolloutSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleRolloutSnapshot(id string, at time.Time) RolloutSnapshot {
	return RolloutSnapshot{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Owner:    "seeded by rollout.SampleRolloutSnapshot",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutBatchSchema is the stored shape of a RolloutBatch.
var RolloutBatchSchema = schema.New("rollout.rolloutbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutBatch is one rollout batch.
type RolloutBatch struct {
	ID        string
	Category  string
	Reference string
	Region    string
	Capacity  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutBatch) Key() string { return "rollout/rolloutbatch/" + x.ID }

// Fields encodes x for storage.
func (x RolloutBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"region":    x.Region,
		"capacity":  strconv.Itoa(x.Capacity),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutBatch decodes a stored record.
func ParseRolloutBatch(f codec.Fields) (RolloutBatch, error) {
	var x RolloutBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutbatch: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutbatch %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutBatch) String() string {
	return fmt.Sprintf("RolloutBatch(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutBatch) Normalize() RolloutBatch {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRolloutBatchs orders by capacity descending, then by ID.
func SortRolloutBatchs(xs []RolloutBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutBatchs drops the archived entries, preserving order.
func ActiveRolloutBatchs(xs []RolloutBatch) []RolloutBatch {
	out := make([]RolloutBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutBatchs adds up the capacity of every active entry.
func SumRolloutBatchs(xs []RolloutBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRolloutBatchs keys the entries by ID, last write winning.
func IndexRolloutBatchs(xs []RolloutBatch) map[string]RolloutBatch {
	out := make(map[string]RolloutBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutBatch builds a valid RolloutBatch, for tests and for the seed data the cron
// binary writes.
func SampleRolloutBatch(id string, at time.Time) RolloutBatch {
	return RolloutBatch{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by rollout.SampleRolloutBatch",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutRequestSchema is the stored shape of a RolloutRequest.
var RolloutRequestSchema = schema.New("rollout.rolloutrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutRequest is one rollout request.
type RolloutRequest struct {
	ID        string
	Owner     string
	Group     string
	Reference string
	Retained  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutRequest) Key() string { return "rollout/rolloutrequest/" + x.ID }

// Fields encodes x for storage.
func (x RolloutRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"group":     x.Group,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutRequest decodes a stored record.
func ParseRolloutRequest(f codec.Fields) (RolloutRequest, error) {
	var x RolloutRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutrequest: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutRequest) String() string {
	return fmt.Sprintf("RolloutRequest(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutRequest) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutRequest) Normalize() RolloutRequest {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRolloutRequests orders by retained descending, then by ID.
func SortRolloutRequests(xs []RolloutRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutRequests drops the archived entries, preserving order.
func ActiveRolloutRequests(xs []RolloutRequest) []RolloutRequest {
	out := make([]RolloutRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutRequests adds up the retained of every active entry.
func SumRolloutRequests(xs []RolloutRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRolloutRequests keys the entries by ID, last write winning.
func IndexRolloutRequests(xs []RolloutRequest) map[string]RolloutRequest {
	out := make(map[string]RolloutRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutRequest builds a valid RolloutRequest, for tests and for the seed data the cron
// binary writes.
func SampleRolloutRequest(id string, at time.Time) RolloutRequest {
	return RolloutRequest{
		ID:        id,
		Owner:     "owner-" + id,
		Group:     "group-" + id,
		Reference: "seeded by rollout.SampleRolloutRequest",
		Retained:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutSummarySchema is the stored shape of a RolloutSummary.
var RolloutSummarySchema = schema.New("rollout.rolloutsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutSummary is one rollout summary.
type RolloutSummary struct {
	ID        string
	Channel   string
	Reference string
	Name      string
	Retained  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutSummary) Key() string { return "rollout/rolloutsummary/" + x.ID }

// Fields encodes x for storage.
func (x RolloutSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"name":      x.Name,
		"retained":  strconv.Itoa(x.Retained),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutSummary decodes a stored record.
func ParseRolloutSummary(f codec.Fields) (RolloutSummary, error) {
	var x RolloutSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutsummary: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutsummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutSummary) String() string {
	return fmt.Sprintf("RolloutSummary(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutSummary) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutSummary) Normalize() RolloutSummary {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRolloutSummarys orders by retained descending, then by ID.
func SortRolloutSummarys(xs []RolloutSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutSummarys drops the archived entries, preserving order.
func ActiveRolloutSummarys(xs []RolloutSummary) []RolloutSummary {
	out := make([]RolloutSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutSummarys adds up the retained of every active entry.
func SumRolloutSummarys(xs []RolloutSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRolloutSummarys keys the entries by ID, last write winning.
func IndexRolloutSummarys(xs []RolloutSummary) map[string]RolloutSummary {
	out := make(map[string]RolloutSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutSummary builds a valid RolloutSummary, for tests and for the seed data the cron
// binary writes.
func SampleRolloutSummary(id string, at time.Time) RolloutSummary {
	return RolloutSummary{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by rollout.SampleRolloutSummary",
		Retained:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutPolicySchema is the stored shape of a RolloutPolicy.
var RolloutPolicySchema = schema.New("rollout.rolloutpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutPolicy is one rollout policy.
type RolloutPolicy struct {
	ID       string
	Label    string
	Category string
	Source   string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutPolicy) Key() string { return "rollout/rolloutpolicy/" + x.ID }

// Fields encodes x for storage.
func (x RolloutPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutPolicy decodes a stored record.
func ParseRolloutPolicy(f codec.Fields) (RolloutPolicy, error) {
	var x RolloutPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutpolicy: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutpolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutPolicy) String() string {
	return fmt.Sprintf("RolloutPolicy(%s, label=%q, capacity=%d, archived=%t)", x.ID, x.Label, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutPolicy) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutPolicy) Normalize() RolloutPolicy {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortRolloutPolicys orders by capacity descending, then by ID.
func SortRolloutPolicys(xs []RolloutPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutPolicys drops the archived entries, preserving order.
func ActiveRolloutPolicys(xs []RolloutPolicy) []RolloutPolicy {
	out := make([]RolloutPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutPolicys adds up the capacity of every active entry.
func SumRolloutPolicys(xs []RolloutPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRolloutPolicys keys the entries by ID, last write winning.
func IndexRolloutPolicys(xs []RolloutPolicy) map[string]RolloutPolicy {
	out := make(map[string]RolloutPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutPolicy builds a valid RolloutPolicy, for tests and for the seed data the cron
// binary writes.
func SampleRolloutPolicy(id string, at time.Time) RolloutPolicy {
	return RolloutPolicy{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Source:   "seeded by rollout.SampleRolloutPolicy",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutLinkSchema is the stored shape of a RolloutLink.
var RolloutLinkSchema = schema.New("rollout.rolloutlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutLink is one rollout link.
type RolloutLink struct {
	ID       string
	Owner    string
	Source   string
	Channel  string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutLink) Key() string { return "rollout/rolloutlink/" + x.ID }

// Fields encodes x for storage.
func (x RolloutLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"channel":  x.Channel,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutLink decodes a stored record.
func ParseRolloutLink(f codec.Fields) (RolloutLink, error) {
	var x RolloutLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutlink: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutlink %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutLink) String() string {
	return fmt.Sprintf("RolloutLink(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutLink) Normalize() RolloutLink {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRolloutLinks orders by retained descending, then by ID.
func SortRolloutLinks(xs []RolloutLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutLinks drops the archived entries, preserving order.
func ActiveRolloutLinks(xs []RolloutLink) []RolloutLink {
	out := make([]RolloutLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutLinks adds up the retained of every active entry.
func SumRolloutLinks(xs []RolloutLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRolloutLinks keys the entries by ID, last write winning.
func IndexRolloutLinks(xs []RolloutLink) map[string]RolloutLink {
	out := make(map[string]RolloutLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutLink builds a valid RolloutLink, for tests and for the seed data the cron
// binary writes.
func SampleRolloutLink(id string, at time.Time) RolloutLink {
	return RolloutLink{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by rollout.SampleRolloutLink",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutNoteSchema is the stored shape of a RolloutNote.
var RolloutNoteSchema = schema.New("rollout.rolloutnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutNote is one rollout note.
type RolloutNote struct {
	ID       string
	Group    string
	Category string
	Owner    string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutNote) Key() string { return "rollout/rolloutnote/" + x.ID }

// Fields encodes x for storage.
func (x RolloutNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutNote decodes a stored record.
func ParseRolloutNote(f codec.Fields) (RolloutNote, error) {
	var x RolloutNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutnote: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutnote %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutNote) String() string {
	return fmt.Sprintf("RolloutNote(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutNote) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutNote) Normalize() RolloutNote {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRolloutNotes orders by capacity descending, then by ID.
func SortRolloutNotes(xs []RolloutNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutNotes drops the archived entries, preserving order.
func ActiveRolloutNotes(xs []RolloutNote) []RolloutNote {
	out := make([]RolloutNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutNotes adds up the capacity of every active entry.
func SumRolloutNotes(xs []RolloutNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRolloutNotes keys the entries by ID, last write winning.
func IndexRolloutNotes(xs []RolloutNote) map[string]RolloutNote {
	out := make(map[string]RolloutNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutNote builds a valid RolloutNote, for tests and for the seed data the cron
// binary writes.
func SampleRolloutNote(id string, at time.Time) RolloutNote {
	return RolloutNote{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Owner:    "seeded by rollout.SampleRolloutNote",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutWindowSchema is the stored shape of a RolloutWindow.
var RolloutWindowSchema = schema.New("rollout.rolloutwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutWindow is one rollout window.
type RolloutWindow struct {
	ID       string
	Comment  string
	Group    string
	Channel  string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutWindow) Key() string { return "rollout/rolloutwindow/" + x.ID }

// Fields encodes x for storage.
func (x RolloutWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutWindow decodes a stored record.
func ParseRolloutWindow(f codec.Fields) (RolloutWindow, error) {
	var x RolloutWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutwindow: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutwindow %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutWindow) String() string {
	return fmt.Sprintf("RolloutWindow(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutWindow) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutWindow) Normalize() RolloutWindow {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRolloutWindows orders by count descending, then by ID.
func SortRolloutWindows(xs []RolloutWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutWindows drops the archived entries, preserving order.
func ActiveRolloutWindows(xs []RolloutWindow) []RolloutWindow {
	out := make([]RolloutWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutWindows adds up the count of every active entry.
func SumRolloutWindows(xs []RolloutWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRolloutWindows keys the entries by ID, last write winning.
func IndexRolloutWindows(xs []RolloutWindow) map[string]RolloutWindow {
	out := make(map[string]RolloutWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutWindow builds a valid RolloutWindow, for tests and for the seed data the cron
// binary writes.
func SampleRolloutWindow(id string, at time.Time) RolloutWindow {
	return RolloutWindow{
		ID:       id,
		Comment:  "comment-" + id,
		Group:    "group-" + id,
		Channel:  "seeded by rollout.SampleRolloutWindow",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutQuotaSchema is the stored shape of a RolloutQuota.
var RolloutQuotaSchema = schema.New("rollout.rolloutquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutQuota is one rollout quota.
type RolloutQuota struct {
	ID       string
	Label    string
	Channel  string
	Source   string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutQuota) Key() string { return "rollout/rolloutquota/" + x.ID }

// Fields encodes x for storage.
func (x RolloutQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutQuota decodes a stored record.
func ParseRolloutQuota(f codec.Fields) (RolloutQuota, error) {
	var x RolloutQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutquota: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutquota %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutQuota) String() string {
	return fmt.Sprintf("RolloutQuota(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutQuota) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutQuota) Normalize() RolloutQuota {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRolloutQuotas orders by retained descending, then by ID.
func SortRolloutQuotas(xs []RolloutQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutQuotas drops the archived entries, preserving order.
func ActiveRolloutQuotas(xs []RolloutQuota) []RolloutQuota {
	out := make([]RolloutQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutQuotas adds up the retained of every active entry.
func SumRolloutQuotas(xs []RolloutQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRolloutQuotas keys the entries by ID, last write winning.
func IndexRolloutQuotas(xs []RolloutQuota) map[string]RolloutQuota {
	out := make(map[string]RolloutQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutQuota builds a valid RolloutQuota, for tests and for the seed data the cron
// binary writes.
func SampleRolloutQuota(id string, at time.Time) RolloutQuota {
	return RolloutQuota{
		ID:       id,
		Label:    "label-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by rollout.SampleRolloutQuota",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutBucketSchema is the stored shape of a RolloutBucket.
var RolloutBucketSchema = schema.New("rollout.rolloutbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutBucket is one rollout bucket.
type RolloutBucket struct {
	ID        string
	Reference string
	Region    string
	Label     string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutBucket) Key() string { return "rollout/rolloutbucket/" + x.ID }

// Fields encodes x for storage.
func (x RolloutBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"label":     x.Label,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutBucket decodes a stored record.
func ParseRolloutBucket(f codec.Fields) (RolloutBucket, error) {
	var x RolloutBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutbucket: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutbucket %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutBucket) String() string {
	return fmt.Sprintf("RolloutBucket(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutBucket) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutBucket) Normalize() RolloutBucket {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRolloutBuckets orders by weight descending, then by ID.
func SortRolloutBuckets(xs []RolloutBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutBuckets drops the archived entries, preserving order.
func ActiveRolloutBuckets(xs []RolloutBucket) []RolloutBucket {
	out := make([]RolloutBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutBuckets adds up the weight of every active entry.
func SumRolloutBuckets(xs []RolloutBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRolloutBuckets keys the entries by ID, last write winning.
func IndexRolloutBuckets(xs []RolloutBucket) map[string]RolloutBucket {
	out := make(map[string]RolloutBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutBucket builds a valid RolloutBucket, for tests and for the seed data the cron
// binary writes.
func SampleRolloutBucket(id string, at time.Time) RolloutBucket {
	return RolloutBucket{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Label:     "seeded by rollout.SampleRolloutBucket",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutLaneSchema is the stored shape of a RolloutLane.
var RolloutLaneSchema = schema.New("rollout.rolloutlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutLane is one rollout lane.
type RolloutLane struct {
	ID       string
	Region   string
	Source   string
	Comment  string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutLane) Key() string { return "rollout/rolloutlane/" + x.ID }

// Fields encodes x for storage.
func (x RolloutLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"source":   x.Source,
		"comment":  x.Comment,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutLane decodes a stored record.
func ParseRolloutLane(f codec.Fields) (RolloutLane, error) {
	var x RolloutLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutlane: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutlane %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutLane) String() string {
	return fmt.Sprintf("RolloutLane(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutLane) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutLane) Normalize() RolloutLane {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortRolloutLanes orders by weight descending, then by ID.
func SortRolloutLanes(xs []RolloutLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutLanes drops the archived entries, preserving order.
func ActiveRolloutLanes(xs []RolloutLane) []RolloutLane {
	out := make([]RolloutLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutLanes adds up the weight of every active entry.
func SumRolloutLanes(xs []RolloutLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRolloutLanes keys the entries by ID, last write winning.
func IndexRolloutLanes(xs []RolloutLane) map[string]RolloutLane {
	out := make(map[string]RolloutLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutLane builds a valid RolloutLane, for tests and for the seed data the cron
// binary writes.
func SampleRolloutLane(id string, at time.Time) RolloutLane {
	return RolloutLane{
		ID:       id,
		Region:   "region-" + id,
		Source:   "source-" + id,
		Comment:  "seeded by rollout.SampleRolloutLane",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutSlotSchema is the stored shape of a RolloutSlot.
var RolloutSlotSchema = schema.New("rollout.rolloutslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutSlot is one rollout slot.
type RolloutSlot struct {
	ID        string
	Comment   string
	Group     string
	Reference string
	Version   int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutSlot) Key() string { return "rollout/rolloutslot/" + x.ID }

// Fields encodes x for storage.
func (x RolloutSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"group":     x.Group,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutSlot decodes a stored record.
func ParseRolloutSlot(f codec.Fields) (RolloutSlot, error) {
	var x RolloutSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutslot: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutslot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutSlot) String() string {
	return fmt.Sprintf("RolloutSlot(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutSlot) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutSlot) Normalize() RolloutSlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortRolloutSlots orders by version descending, then by ID.
func SortRolloutSlots(xs []RolloutSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutSlots drops the archived entries, preserving order.
func ActiveRolloutSlots(xs []RolloutSlot) []RolloutSlot {
	out := make([]RolloutSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutSlots adds up the version of every active entry.
func SumRolloutSlots(xs []RolloutSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexRolloutSlots keys the entries by ID, last write winning.
func IndexRolloutSlots(xs []RolloutSlot) map[string]RolloutSlot {
	out := make(map[string]RolloutSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutSlot builds a valid RolloutSlot, for tests and for the seed data the cron
// binary writes.
func SampleRolloutSlot(id string, at time.Time) RolloutSlot {
	return RolloutSlot{
		ID:        id,
		Comment:   "comment-" + id,
		Group:     "group-" + id,
		Reference: "seeded by rollout.SampleRolloutSlot",
		Version:   len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutTagSchema is the stored shape of a RolloutTag.
var RolloutTagSchema = schema.New("rollout.rollouttag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutTag is one rollout tag.
type RolloutTag struct {
	ID        string
	Source    string
	Reference string
	Label     string
	Offset    int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutTag) Key() string { return "rollout/rollouttag/" + x.ID }

// Fields encodes x for storage.
func (x RolloutTag) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"reference": x.Reference,
		"label":     x.Label,
		"offset":    strconv.Itoa(x.Offset),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutTag decodes a stored record.
func ParseRolloutTag(f codec.Fields) (RolloutTag, error) {
	var x RolloutTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rollouttag: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("rollout: rollouttag %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("rollout: rollouttag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rollouttag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rollouttag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutTag) String() string {
	return fmt.Sprintf("RolloutTag(%s, source=%q, offset=%d, archived=%t)", x.ID, x.Source, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutTag) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutTag) Normalize() RolloutTag {
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRolloutTags orders by offset descending, then by ID.
func SortRolloutTags(xs []RolloutTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutTags drops the archived entries, preserving order.
func ActiveRolloutTags(xs []RolloutTag) []RolloutTag {
	out := make([]RolloutTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutTags adds up the offset of every active entry.
func SumRolloutTags(xs []RolloutTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRolloutTags keys the entries by ID, last write winning.
func IndexRolloutTags(xs []RolloutTag) map[string]RolloutTag {
	out := make(map[string]RolloutTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutTag builds a valid RolloutTag, for tests and for the seed data the cron
// binary writes.
func SampleRolloutTag(id string, at time.Time) RolloutTag {
	return RolloutTag{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Label:     "seeded by rollout.SampleRolloutTag",
		Offset:    len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutRuleSchema is the stored shape of a RolloutRule.
var RolloutRuleSchema = schema.New("rollout.rolloutrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutRule is one rollout rule.
type RolloutRule struct {
	ID       string
	Comment  string
	Group    string
	Name     string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutRule) Key() string { return "rollout/rolloutrule/" + x.ID }

// Fields encodes x for storage.
func (x RolloutRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"name":     x.Name,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutRule decodes a stored record.
func ParseRolloutRule(f codec.Fields) (RolloutRule, error) {
	var x RolloutRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutrule: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Name = f["name"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrule %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutRule) String() string {
	return fmt.Sprintf("RolloutRule(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutRule) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutRule) Normalize() RolloutRule {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRolloutRules orders by priority descending, then by ID.
func SortRolloutRules(xs []RolloutRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutRules drops the archived entries, preserving order.
func ActiveRolloutRules(xs []RolloutRule) []RolloutRule {
	out := make([]RolloutRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutRules adds up the priority of every active entry.
func SumRolloutRules(xs []RolloutRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRolloutRules keys the entries by ID, last write winning.
func IndexRolloutRules(xs []RolloutRule) map[string]RolloutRule {
	out := make(map[string]RolloutRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutRule builds a valid RolloutRule, for tests and for the seed data the cron
// binary writes.
func SampleRolloutRule(id string, at time.Time) RolloutRule {
	return RolloutRule{
		ID:       id,
		Comment:  "comment-" + id,
		Group:    "group-" + id,
		Name:     "seeded by rollout.SampleRolloutRule",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutHoldSchema is the stored shape of a RolloutHold.
var RolloutHoldSchema = schema.New("rollout.rollouthold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutHold is one rollout hold.
type RolloutHold struct {
	ID       string
	Region   string
	Name     string
	Group    string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutHold) Key() string { return "rollout/rollouthold/" + x.ID }

// Fields encodes x for storage.
func (x RolloutHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutHold decodes a stored record.
func ParseRolloutHold(f codec.Fields) (RolloutHold, error) {
	var x RolloutHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rollouthold: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("rollout: rollouthold %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rollouthold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rollouthold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rollouthold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutHold) String() string {
	return fmt.Sprintf("RolloutHold(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutHold) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutHold) Normalize() RolloutHold {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRolloutHolds orders by attempts descending, then by ID.
func SortRolloutHolds(xs []RolloutHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutHolds drops the archived entries, preserving order.
func ActiveRolloutHolds(xs []RolloutHold) []RolloutHold {
	out := make([]RolloutHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutHolds adds up the attempts of every active entry.
func SumRolloutHolds(xs []RolloutHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexRolloutHolds keys the entries by ID, last write winning.
func IndexRolloutHolds(xs []RolloutHold) map[string]RolloutHold {
	out := make(map[string]RolloutHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutHold builds a valid RolloutHold, for tests and for the seed data the cron
// binary writes.
func SampleRolloutHold(id string, at time.Time) RolloutHold {
	return RolloutHold{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Group:    "seeded by rollout.SampleRolloutHold",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RolloutClaimSchema is the stored shape of a RolloutClaim.
var RolloutClaimSchema = schema.New("rollout.rolloutclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutClaim is one rollout claim.
type RolloutClaim struct {
	ID        string
	Group     string
	Reference string
	Comment   string
	Offset    int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RolloutClaim) Key() string { return "rollout/rolloutclaim/" + x.ID }

// Fields encodes x for storage.
func (x RolloutClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"comment":   x.Comment,
		"offset":    strconv.Itoa(x.Offset),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutClaim decodes a stored record.
func ParseRolloutClaim(f codec.Fields) (RolloutClaim, error) {
	var x RolloutClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutclaim: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutclaim %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutClaim) String() string {
	return fmt.Sprintf("RolloutClaim(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutClaim) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutClaim) Normalize() RolloutClaim {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRolloutClaims orders by offset descending, then by ID.
func SortRolloutClaims(xs []RolloutClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutClaims drops the archived entries, preserving order.
func ActiveRolloutClaims(xs []RolloutClaim) []RolloutClaim {
	out := make([]RolloutClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutClaims adds up the offset of every active entry.
func SumRolloutClaims(xs []RolloutClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRolloutClaims keys the entries by ID, last write winning.
func IndexRolloutClaims(xs []RolloutClaim) map[string]RolloutClaim {
	out := make(map[string]RolloutClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutClaim builds a valid RolloutClaim, for tests and for the seed data the cron
// binary writes.
func SampleRolloutClaim(id string, at time.Time) RolloutClaim {
	return RolloutClaim{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by rollout.SampleRolloutClaim",
		Offset:    len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RolloutCursorSchema is the stored shape of a RolloutCursor.
var RolloutCursorSchema = schema.New("rollout.rolloutcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RolloutCursor is one rollout cursor.
type RolloutCursor struct {
	ID       string
	Channel  string
	Region   string
	Owner    string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RolloutCursor) Key() string { return "rollout/rolloutcursor/" + x.ID }

// Fields encodes x for storage.
func (x RolloutCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRolloutCursor decodes a stored record.
func ParseRolloutCursor(f codec.Fields) (RolloutCursor, error) {
	var x RolloutCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("rollout: rolloutcursor: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutcursor %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("rollout: rolloutcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("rollout: rolloutcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("rollout: rolloutcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RolloutCursor) String() string {
	return fmt.Sprintf("RolloutCursor(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RolloutCursor) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RolloutCursor) Normalize() RolloutCursor {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRolloutCursors orders by retained descending, then by ID.
func SortRolloutCursors(xs []RolloutCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRolloutCursors drops the archived entries, preserving order.
func ActiveRolloutCursors(xs []RolloutCursor) []RolloutCursor {
	out := make([]RolloutCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRolloutCursors adds up the retained of every active entry.
func SumRolloutCursors(xs []RolloutCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRolloutCursors keys the entries by ID, last write winning.
func IndexRolloutCursors(xs []RolloutCursor) map[string]RolloutCursor {
	out := make(map[string]RolloutCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRolloutCursor builds a valid RolloutCursor, for tests and for the seed data the cron
// binary writes.
func SampleRolloutCursor(id string, at time.Time) RolloutCursor {
	return RolloutCursor{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Owner:    "seeded by rollout.SampleRolloutCursor",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	RolloutRecordSchema,
	RolloutEntrySchema,
	RolloutItemSchema,
	RolloutDraftSchema,
	RolloutRevisionSchema,
	RolloutSnapshotSchema,
	RolloutBatchSchema,
	RolloutRequestSchema,
	RolloutSummarySchema,
	RolloutPolicySchema,
	RolloutLinkSchema,
	RolloutNoteSchema,
	RolloutWindowSchema,
	RolloutQuotaSchema,
	RolloutBucketSchema,
	RolloutLaneSchema,
	RolloutSlotSchema,
	RolloutTagSchema,
	RolloutRuleSchema,
	RolloutHoldSchema,
	RolloutClaimSchema,
	RolloutCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"rollout/rolloutrecord/",
	"rollout/rolloutentry/",
	"rollout/rolloutitem/",
	"rollout/rolloutdraft/",
	"rollout/rolloutrevision/",
	"rollout/rolloutsnapshot/",
	"rollout/rolloutbatch/",
	"rollout/rolloutrequest/",
	"rollout/rolloutsummary/",
	"rollout/rolloutpolicy/",
	"rollout/rolloutlink/",
	"rollout/rolloutnote/",
	"rollout/rolloutwindow/",
	"rollout/rolloutquota/",
	"rollout/rolloutbucket/",
	"rollout/rolloutlane/",
	"rollout/rolloutslot/",
	"rollout/rollouttag/",
	"rollout/rolloutrule/",
	"rollout/rollouthold/",
	"rollout/rolloutclaim/",
	"rollout/rolloutcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "rollout: " + strings.Join(names, ", ")
}
