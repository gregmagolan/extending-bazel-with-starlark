// Package trail holds the trail domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package trail

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// TrailRecordSchema is the stored shape of a TrailRecord.
var TrailRecordSchema = schema.New("trail.trailrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailRecord is one trail record.
type TrailRecord struct {
	ID       string
	Region   string
	Category string
	Source   string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailRecord) Key() string { return "trail/trailrecord/" + x.ID }

// Fields encodes x for storage.
func (x TrailRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailRecord decodes a stored record.
func ParseTrailRecord(f codec.Fields) (TrailRecord, error) {
	var x TrailRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailrecord: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("trail: trailrecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("trail: trailrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailRecord) String() string {
	return fmt.Sprintf("TrailRecord(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailRecord) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailRecord) Normalize() TrailRecord {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTrailRecords orders by offset descending, then by ID.
func SortTrailRecords(xs []TrailRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailRecords drops the archived entries, preserving order.
func ActiveTrailRecords(xs []TrailRecord) []TrailRecord {
	out := make([]TrailRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailRecords adds up the offset of every active entry.
func SumTrailRecords(xs []TrailRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTrailRecords keys the entries by ID, last write winning.
func IndexTrailRecords(xs []TrailRecord) map[string]TrailRecord {
	out := make(map[string]TrailRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailRecord builds a valid TrailRecord, for tests and for the seed data the cron
// binary writes.
func SampleTrailRecord(id string, at time.Time) TrailRecord {
	return TrailRecord{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Source:   "seeded by trail.SampleTrailRecord",
		Offset:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailEntrySchema is the stored shape of a TrailEntry.
var TrailEntrySchema = schema.New("trail.trailentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailEntry is one trail entry.
type TrailEntry struct {
	ID        string
	Reference string
	Group     string
	Source    string
	Offset    int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TrailEntry) Key() string { return "trail/trailentry/" + x.ID }

// Fields encodes x for storage.
func (x TrailEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"source":    x.Source,
		"offset":    strconv.Itoa(x.Offset),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailEntry decodes a stored record.
func ParseTrailEntry(f codec.Fields) (TrailEntry, error) {
	var x TrailEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailentry: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("trail: trailentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("trail: trailentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailEntry) String() string {
	return fmt.Sprintf("TrailEntry(%s, reference=%q, offset=%d, archived=%t)", x.ID, x.Reference, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailEntry) Normalize() TrailEntry {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTrailEntrys orders by offset descending, then by ID.
func SortTrailEntrys(xs []TrailEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailEntrys drops the archived entries, preserving order.
func ActiveTrailEntrys(xs []TrailEntry) []TrailEntry {
	out := make([]TrailEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailEntrys adds up the offset of every active entry.
func SumTrailEntrys(xs []TrailEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTrailEntrys keys the entries by ID, last write winning.
func IndexTrailEntrys(xs []TrailEntry) map[string]TrailEntry {
	out := make(map[string]TrailEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailEntry builds a valid TrailEntry, for tests and for the seed data the cron
// binary writes.
func SampleTrailEntry(id string, at time.Time) TrailEntry {
	return TrailEntry{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Source:    "seeded by trail.SampleTrailEntry",
		Offset:    len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TrailItemSchema is the stored shape of a TrailItem.
var TrailItemSchema = schema.New("trail.trailitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailItem is one trail item.
type TrailItem struct {
	ID       string
	Category string
	Owner    string
	Group    string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailItem) Key() string { return "trail/trailitem/" + x.ID }

// Fields encodes x for storage.
func (x TrailItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailItem decodes a stored record.
func ParseTrailItem(f codec.Fields) (TrailItem, error) {
	var x TrailItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailitem: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: trailitem %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("trail: trailitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailItem) String() string {
	return fmt.Sprintf("TrailItem(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailItem) Normalize() TrailItem {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTrailItems orders by version descending, then by ID.
func SortTrailItems(xs []TrailItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailItems drops the archived entries, preserving order.
func ActiveTrailItems(xs []TrailItem) []TrailItem {
	out := make([]TrailItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailItems adds up the version of every active entry.
func SumTrailItems(xs []TrailItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTrailItems keys the entries by ID, last write winning.
func IndexTrailItems(xs []TrailItem) map[string]TrailItem {
	out := make(map[string]TrailItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailItem builds a valid TrailItem, for tests and for the seed data the cron
// binary writes.
func SampleTrailItem(id string, at time.Time) TrailItem {
	return TrailItem{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by trail.SampleTrailItem",
		Version:  len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailDraftSchema is the stored shape of a TrailDraft.
var TrailDraftSchema = schema.New("trail.traildraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailDraft is one trail draft.
type TrailDraft struct {
	ID       string
	Category string
	Owner    string
	Region   string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailDraft) Key() string { return "trail/traildraft/" + x.ID }

// Fields encodes x for storage.
func (x TrailDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailDraft decodes a stored record.
func ParseTrailDraft(f codec.Fields) (TrailDraft, error) {
	var x TrailDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: traildraft: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("trail: traildraft %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("trail: traildraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: traildraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: traildraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailDraft) String() string {
	return fmt.Sprintf("TrailDraft(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailDraft) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailDraft) Normalize() TrailDraft {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTrailDrafts orders by capacity descending, then by ID.
func SortTrailDrafts(xs []TrailDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailDrafts drops the archived entries, preserving order.
func ActiveTrailDrafts(xs []TrailDraft) []TrailDraft {
	out := make([]TrailDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailDrafts adds up the capacity of every active entry.
func SumTrailDrafts(xs []TrailDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTrailDrafts keys the entries by ID, last write winning.
func IndexTrailDrafts(xs []TrailDraft) map[string]TrailDraft {
	out := make(map[string]TrailDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailDraft builds a valid TrailDraft, for tests and for the seed data the cron
// binary writes.
func SampleTrailDraft(id string, at time.Time) TrailDraft {
	return TrailDraft{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by trail.SampleTrailDraft",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailRevisionSchema is the stored shape of a TrailRevision.
var TrailRevisionSchema = schema.New("trail.trailrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailRevision is one trail revision.
type TrailRevision struct {
	ID       string
	Category string
	Source   string
	Channel  string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailRevision) Key() string { return "trail/trailrevision/" + x.ID }

// Fields encodes x for storage.
func (x TrailRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"source":   x.Source,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailRevision decodes a stored record.
func ParseTrailRevision(f codec.Fields) (TrailRevision, error) {
	var x TrailRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailrevision: missing id")
	}
	x.Category = f["category"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: trailrevision %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("trail: trailrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailRevision) String() string {
	return fmt.Sprintf("TrailRevision(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailRevision) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailRevision) Normalize() TrailRevision {
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTrailRevisions orders by version descending, then by ID.
func SortTrailRevisions(xs []TrailRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailRevisions drops the archived entries, preserving order.
func ActiveTrailRevisions(xs []TrailRevision) []TrailRevision {
	out := make([]TrailRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailRevisions adds up the version of every active entry.
func SumTrailRevisions(xs []TrailRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTrailRevisions keys the entries by ID, last write winning.
func IndexTrailRevisions(xs []TrailRevision) map[string]TrailRevision {
	out := make(map[string]TrailRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailRevision builds a valid TrailRevision, for tests and for the seed data the cron
// binary writes.
func SampleTrailRevision(id string, at time.Time) TrailRevision {
	return TrailRevision{
		ID:       id,
		Category: "category-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by trail.SampleTrailRevision",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailSnapshotSchema is the stored shape of a TrailSnapshot.
var TrailSnapshotSchema = schema.New("trail.trailsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailSnapshot is one trail snapshot.
type TrailSnapshot struct {
	ID       string
	Comment  string
	Source   string
	Owner    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailSnapshot) Key() string { return "trail/trailsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x TrailSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailSnapshot decodes a stored record.
func ParseTrailSnapshot(f codec.Fields) (TrailSnapshot, error) {
	var x TrailSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailsnapshot: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("trail: trailsnapshot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: trailsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailSnapshot) String() string {
	return fmt.Sprintf("TrailSnapshot(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailSnapshot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailSnapshot) Normalize() TrailSnapshot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTrailSnapshots orders by offset descending, then by ID.
func SortTrailSnapshots(xs []TrailSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailSnapshots drops the archived entries, preserving order.
func ActiveTrailSnapshots(xs []TrailSnapshot) []TrailSnapshot {
	out := make([]TrailSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailSnapshots adds up the offset of every active entry.
func SumTrailSnapshots(xs []TrailSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTrailSnapshots keys the entries by ID, last write winning.
func IndexTrailSnapshots(xs []TrailSnapshot) map[string]TrailSnapshot {
	out := make(map[string]TrailSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailSnapshot builds a valid TrailSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleTrailSnapshot(id string, at time.Time) TrailSnapshot {
	return TrailSnapshot{
		ID:      id,
		Comment: "comment-" + id,
		Source:  "source-" + id,
		Owner:   "seeded by trail.SampleTrailSnapshot",
		Offset:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TrailBatchSchema is the stored shape of a TrailBatch.
var TrailBatchSchema = schema.New("trail.trailbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailBatch is one trail batch.
type TrailBatch struct {
	ID       string
	Label    string
	Source   string
	Group    string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailBatch) Key() string { return "trail/trailbatch/" + x.ID }

// Fields encodes x for storage.
func (x TrailBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailBatch decodes a stored record.
func ParseTrailBatch(f codec.Fields) (TrailBatch, error) {
	var x TrailBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailbatch: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("trail: trailbatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("trail: trailbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailBatch) String() string {
	return fmt.Sprintf("TrailBatch(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailBatch) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailBatch) Normalize() TrailBatch {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTrailBatchs orders by retained descending, then by ID.
func SortTrailBatchs(xs []TrailBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailBatchs drops the archived entries, preserving order.
func ActiveTrailBatchs(xs []TrailBatch) []TrailBatch {
	out := make([]TrailBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailBatchs adds up the retained of every active entry.
func SumTrailBatchs(xs []TrailBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTrailBatchs keys the entries by ID, last write winning.
func IndexTrailBatchs(xs []TrailBatch) map[string]TrailBatch {
	out := make(map[string]TrailBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailBatch builds a valid TrailBatch, for tests and for the seed data the cron
// binary writes.
func SampleTrailBatch(id string, at time.Time) TrailBatch {
	return TrailBatch{
		ID:       id,
		Label:    "label-" + id,
		Source:   "source-" + id,
		Group:    "seeded by trail.SampleTrailBatch",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailRequestSchema is the stored shape of a TrailRequest.
var TrailRequestSchema = schema.New("trail.trailrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailRequest is one trail request.
type TrailRequest struct {
	ID       string
	Label    string
	Owner    string
	Region   string
	Weight   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailRequest) Key() string { return "trail/trailrequest/" + x.ID }

// Fields encodes x for storage.
func (x TrailRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"owner":    x.Owner,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailRequest decodes a stored record.
func ParseTrailRequest(f codec.Fields) (TrailRequest, error) {
	var x TrailRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailrequest: missing id")
	}
	x.Label = f["label"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("trail: trailrequest %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: trailrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailRequest) String() string {
	return fmt.Sprintf("TrailRequest(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailRequest) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailRequest) Normalize() TrailRequest {
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTrailRequests orders by weight descending, then by ID.
func SortTrailRequests(xs []TrailRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailRequests drops the archived entries, preserving order.
func ActiveTrailRequests(xs []TrailRequest) []TrailRequest {
	out := make([]TrailRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailRequests adds up the weight of every active entry.
func SumTrailRequests(xs []TrailRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTrailRequests keys the entries by ID, last write winning.
func IndexTrailRequests(xs []TrailRequest) map[string]TrailRequest {
	out := make(map[string]TrailRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailRequest builds a valid TrailRequest, for tests and for the seed data the cron
// binary writes.
func SampleTrailRequest(id string, at time.Time) TrailRequest {
	return TrailRequest{
		ID:      id,
		Label:   "label-" + id,
		Owner:   "owner-" + id,
		Region:  "seeded by trail.SampleTrailRequest",
		Weight:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TrailSummarySchema is the stored shape of a TrailSummary.
var TrailSummarySchema = schema.New("trail.trailsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailSummary is one trail summary.
type TrailSummary struct {
	ID       string
	Source   string
	Owner    string
	Label    string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailSummary) Key() string { return "trail/trailsummary/" + x.ID }

// Fields encodes x for storage.
func (x TrailSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"label":    x.Label,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailSummary decodes a stored record.
func ParseTrailSummary(f codec.Fields) (TrailSummary, error) {
	var x TrailSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailsummary: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: trailsummary %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("trail: trailsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailSummary) String() string {
	return fmt.Sprintf("TrailSummary(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailSummary) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailSummary) Normalize() TrailSummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTrailSummarys orders by version descending, then by ID.
func SortTrailSummarys(xs []TrailSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailSummarys drops the archived entries, preserving order.
func ActiveTrailSummarys(xs []TrailSummary) []TrailSummary {
	out := make([]TrailSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailSummarys adds up the version of every active entry.
func SumTrailSummarys(xs []TrailSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTrailSummarys keys the entries by ID, last write winning.
func IndexTrailSummarys(xs []TrailSummary) map[string]TrailSummary {
	out := make(map[string]TrailSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailSummary builds a valid TrailSummary, for tests and for the seed data the cron
// binary writes.
func SampleTrailSummary(id string, at time.Time) TrailSummary {
	return TrailSummary{
		ID:      id,
		Source:  "source-" + id,
		Owner:   "owner-" + id,
		Label:   "seeded by trail.SampleTrailSummary",
		Version: len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TrailPolicySchema is the stored shape of a TrailPolicy.
var TrailPolicySchema = schema.New("trail.trailpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailPolicy is one trail policy.
type TrailPolicy struct {
	ID        string
	Reference string
	Region    string
	Group     string
	Offset    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TrailPolicy) Key() string { return "trail/trailpolicy/" + x.ID }

// Fields encodes x for storage.
func (x TrailPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"group":     x.Group,
		"offset":    strconv.Itoa(x.Offset),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailPolicy decodes a stored record.
func ParseTrailPolicy(f codec.Fields) (TrailPolicy, error) {
	var x TrailPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailpolicy: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("trail: trailpolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("trail: trailpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailPolicy) String() string {
	return fmt.Sprintf("TrailPolicy(%s, reference=%q, offset=%d, archived=%t)", x.ID, x.Reference, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailPolicy) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailPolicy) Normalize() TrailPolicy {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortTrailPolicys orders by offset descending, then by ID.
func SortTrailPolicys(xs []TrailPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailPolicys drops the archived entries, preserving order.
func ActiveTrailPolicys(xs []TrailPolicy) []TrailPolicy {
	out := make([]TrailPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailPolicys adds up the offset of every active entry.
func SumTrailPolicys(xs []TrailPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTrailPolicys keys the entries by ID, last write winning.
func IndexTrailPolicys(xs []TrailPolicy) map[string]TrailPolicy {
	out := make(map[string]TrailPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailPolicy builds a valid TrailPolicy, for tests and for the seed data the cron
// binary writes.
func SampleTrailPolicy(id string, at time.Time) TrailPolicy {
	return TrailPolicy{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Group:     "seeded by trail.SampleTrailPolicy",
		Offset:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TrailLinkSchema is the stored shape of a TrailLink.
var TrailLinkSchema = schema.New("trail.traillink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailLink is one trail link.
type TrailLink struct {
	ID       string
	Name     string
	Source   string
	Owner    string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailLink) Key() string { return "trail/traillink/" + x.ID }

// Fields encodes x for storage.
func (x TrailLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailLink decodes a stored record.
func ParseTrailLink(f codec.Fields) (TrailLink, error) {
	var x TrailLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: traillink: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("trail: traillink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("trail: traillink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: traillink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: traillink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailLink) String() string {
	return fmt.Sprintf("TrailLink(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailLink) Normalize() TrailLink {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTrailLinks orders by retained descending, then by ID.
func SortTrailLinks(xs []TrailLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailLinks drops the archived entries, preserving order.
func ActiveTrailLinks(xs []TrailLink) []TrailLink {
	out := make([]TrailLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailLinks adds up the retained of every active entry.
func SumTrailLinks(xs []TrailLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTrailLinks keys the entries by ID, last write winning.
func IndexTrailLinks(xs []TrailLink) map[string]TrailLink {
	out := make(map[string]TrailLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailLink builds a valid TrailLink, for tests and for the seed data the cron
// binary writes.
func SampleTrailLink(id string, at time.Time) TrailLink {
	return TrailLink{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by trail.SampleTrailLink",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailNoteSchema is the stored shape of a TrailNote.
var TrailNoteSchema = schema.New("trail.trailnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailNote is one trail note.
type TrailNote struct {
	ID       string
	Owner    string
	Channel  string
	Comment  string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailNote) Key() string { return "trail/trailnote/" + x.ID }

// Fields encodes x for storage.
func (x TrailNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailNote decodes a stored record.
func ParseTrailNote(f codec.Fields) (TrailNote, error) {
	var x TrailNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailnote: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("trail: trailnote %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("trail: trailnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailNote) String() string {
	return fmt.Sprintf("TrailNote(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailNote) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailNote) Normalize() TrailNote {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTrailNotes orders by count descending, then by ID.
func SortTrailNotes(xs []TrailNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailNotes drops the archived entries, preserving order.
func ActiveTrailNotes(xs []TrailNote) []TrailNote {
	out := make([]TrailNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailNotes adds up the count of every active entry.
func SumTrailNotes(xs []TrailNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTrailNotes keys the entries by ID, last write winning.
func IndexTrailNotes(xs []TrailNote) map[string]TrailNote {
	out := make(map[string]TrailNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailNote builds a valid TrailNote, for tests and for the seed data the cron
// binary writes.
func SampleTrailNote(id string, at time.Time) TrailNote {
	return TrailNote{
		ID:      id,
		Owner:   "owner-" + id,
		Channel: "channel-" + id,
		Comment: "seeded by trail.SampleTrailNote",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TrailWindowSchema is the stored shape of a TrailWindow.
var TrailWindowSchema = schema.New("trail.trailwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailWindow is one trail window.
type TrailWindow struct {
	ID       string
	Group    string
	Channel  string
	Name     string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailWindow) Key() string { return "trail/trailwindow/" + x.ID }

// Fields encodes x for storage.
func (x TrailWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailWindow decodes a stored record.
func ParseTrailWindow(f codec.Fields) (TrailWindow, error) {
	var x TrailWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailwindow: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("trail: trailwindow %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("trail: trailwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailWindow) String() string {
	return fmt.Sprintf("TrailWindow(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailWindow) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailWindow) Normalize() TrailWindow {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTrailWindows orders by attempts descending, then by ID.
func SortTrailWindows(xs []TrailWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailWindows drops the archived entries, preserving order.
func ActiveTrailWindows(xs []TrailWindow) []TrailWindow {
	out := make([]TrailWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailWindows adds up the attempts of every active entry.
func SumTrailWindows(xs []TrailWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTrailWindows keys the entries by ID, last write winning.
func IndexTrailWindows(xs []TrailWindow) map[string]TrailWindow {
	out := make(map[string]TrailWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailWindow builds a valid TrailWindow, for tests and for the seed data the cron
// binary writes.
func SampleTrailWindow(id string, at time.Time) TrailWindow {
	return TrailWindow{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by trail.SampleTrailWindow",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailQuotaSchema is the stored shape of a TrailQuota.
var TrailQuotaSchema = schema.New("trail.trailquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailQuota is one trail quota.
type TrailQuota struct {
	ID       string
	Region   string
	Channel  string
	Name     string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailQuota) Key() string { return "trail/trailquota/" + x.ID }

// Fields encodes x for storage.
func (x TrailQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"name":     x.Name,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailQuota decodes a stored record.
func ParseTrailQuota(f codec.Fields) (TrailQuota, error) {
	var x TrailQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailquota: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("trail: trailquota %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("trail: trailquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailQuota) String() string {
	return fmt.Sprintf("TrailQuota(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailQuota) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailQuota) Normalize() TrailQuota {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTrailQuotas orders by capacity descending, then by ID.
func SortTrailQuotas(xs []TrailQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailQuotas drops the archived entries, preserving order.
func ActiveTrailQuotas(xs []TrailQuota) []TrailQuota {
	out := make([]TrailQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailQuotas adds up the capacity of every active entry.
func SumTrailQuotas(xs []TrailQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTrailQuotas keys the entries by ID, last write winning.
func IndexTrailQuotas(xs []TrailQuota) map[string]TrailQuota {
	out := make(map[string]TrailQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailQuota builds a valid TrailQuota, for tests and for the seed data the cron
// binary writes.
func SampleTrailQuota(id string, at time.Time) TrailQuota {
	return TrailQuota{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by trail.SampleTrailQuota",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailBucketSchema is the stored shape of a TrailBucket.
var TrailBucketSchema = schema.New("trail.trailbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailBucket is one trail bucket.
type TrailBucket struct {
	ID       string
	Comment  string
	Region   string
	Channel  string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailBucket) Key() string { return "trail/trailbucket/" + x.ID }

// Fields encodes x for storage.
func (x TrailBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"region":   x.Region,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailBucket decodes a stored record.
func ParseTrailBucket(f codec.Fields) (TrailBucket, error) {
	var x TrailBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailbucket: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("trail: trailbucket %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("trail: trailbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailBucket) String() string {
	return fmt.Sprintf("TrailBucket(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailBucket) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailBucket) Normalize() TrailBucket {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTrailBuckets orders by count descending, then by ID.
func SortTrailBuckets(xs []TrailBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailBuckets drops the archived entries, preserving order.
func ActiveTrailBuckets(xs []TrailBucket) []TrailBucket {
	out := make([]TrailBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailBuckets adds up the count of every active entry.
func SumTrailBuckets(xs []TrailBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTrailBuckets keys the entries by ID, last write winning.
func IndexTrailBuckets(xs []TrailBucket) map[string]TrailBucket {
	out := make(map[string]TrailBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailBucket builds a valid TrailBucket, for tests and for the seed data the cron
// binary writes.
func SampleTrailBucket(id string, at time.Time) TrailBucket {
	return TrailBucket{
		ID:       id,
		Comment:  "comment-" + id,
		Region:   "region-" + id,
		Channel:  "seeded by trail.SampleTrailBucket",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailLaneSchema is the stored shape of a TrailLane.
var TrailLaneSchema = schema.New("trail.traillane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailLane is one trail lane.
type TrailLane struct {
	ID       string
	Category string
	Label    string
	Group    string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailLane) Key() string { return "trail/traillane/" + x.ID }

// Fields encodes x for storage.
func (x TrailLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailLane decodes a stored record.
func ParseTrailLane(f codec.Fields) (TrailLane, error) {
	var x TrailLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: traillane: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("trail: traillane %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: traillane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: traillane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: traillane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailLane) String() string {
	return fmt.Sprintf("TrailLane(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailLane) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailLane) Normalize() TrailLane {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTrailLanes orders by retained descending, then by ID.
func SortTrailLanes(xs []TrailLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailLanes drops the archived entries, preserving order.
func ActiveTrailLanes(xs []TrailLane) []TrailLane {
	out := make([]TrailLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailLanes adds up the retained of every active entry.
func SumTrailLanes(xs []TrailLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTrailLanes keys the entries by ID, last write winning.
func IndexTrailLanes(xs []TrailLane) map[string]TrailLane {
	out := make(map[string]TrailLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailLane builds a valid TrailLane, for tests and for the seed data the cron
// binary writes.
func SampleTrailLane(id string, at time.Time) TrailLane {
	return TrailLane{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Group:    "seeded by trail.SampleTrailLane",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailSlotSchema is the stored shape of a TrailSlot.
var TrailSlotSchema = schema.New("trail.trailslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailSlot is one trail slot.
type TrailSlot struct {
	ID       string
	Name     string
	Source   string
	Channel  string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailSlot) Key() string { return "trail/trailslot/" + x.ID }

// Fields encodes x for storage.
func (x TrailSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailSlot decodes a stored record.
func ParseTrailSlot(f codec.Fields) (TrailSlot, error) {
	var x TrailSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailslot: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("trail: trailslot %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("trail: trailslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailSlot) String() string {
	return fmt.Sprintf("TrailSlot(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailSlot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailSlot) Normalize() TrailSlot {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTrailSlots orders by weight descending, then by ID.
func SortTrailSlots(xs []TrailSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailSlots drops the archived entries, preserving order.
func ActiveTrailSlots(xs []TrailSlot) []TrailSlot {
	out := make([]TrailSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailSlots adds up the weight of every active entry.
func SumTrailSlots(xs []TrailSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTrailSlots keys the entries by ID, last write winning.
func IndexTrailSlots(xs []TrailSlot) map[string]TrailSlot {
	out := make(map[string]TrailSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailSlot builds a valid TrailSlot, for tests and for the seed data the cron
// binary writes.
func SampleTrailSlot(id string, at time.Time) TrailSlot {
	return TrailSlot{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by trail.SampleTrailSlot",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailTagSchema is the stored shape of a TrailTag.
var TrailTagSchema = schema.New("trail.trailtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailTag is one trail tag.
type TrailTag struct {
	ID       string
	Region   string
	Owner    string
	Group    string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailTag) Key() string { return "trail/trailtag/" + x.ID }

// Fields encodes x for storage.
func (x TrailTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"owner":    x.Owner,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailTag decodes a stored record.
func ParseTrailTag(f codec.Fields) (TrailTag, error) {
	var x TrailTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailtag: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("trail: trailtag %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("trail: trailtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailTag) String() string {
	return fmt.Sprintf("TrailTag(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailTag) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailTag) Normalize() TrailTag {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTrailTags orders by retained descending, then by ID.
func SortTrailTags(xs []TrailTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailTags drops the archived entries, preserving order.
func ActiveTrailTags(xs []TrailTag) []TrailTag {
	out := make([]TrailTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailTags adds up the retained of every active entry.
func SumTrailTags(xs []TrailTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTrailTags keys the entries by ID, last write winning.
func IndexTrailTags(xs []TrailTag) map[string]TrailTag {
	out := make(map[string]TrailTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailTag builds a valid TrailTag, for tests and for the seed data the cron
// binary writes.
func SampleTrailTag(id string, at time.Time) TrailTag {
	return TrailTag{
		ID:       id,
		Region:   "region-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by trail.SampleTrailTag",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailRuleSchema is the stored shape of a TrailRule.
var TrailRuleSchema = schema.New("trail.trailrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailRule is one trail rule.
type TrailRule struct {
	ID       string
	Comment  string
	Category string
	Label    string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailRule) Key() string { return "trail/trailrule/" + x.ID }

// Fields encodes x for storage.
func (x TrailRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailRule decodes a stored record.
func ParseTrailRule(f codec.Fields) (TrailRule, error) {
	var x TrailRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailrule: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("trail: trailrule %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: trailrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailRule) String() string {
	return fmt.Sprintf("TrailRule(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailRule) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailRule) Normalize() TrailRule {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTrailRules orders by retained descending, then by ID.
func SortTrailRules(xs []TrailRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailRules drops the archived entries, preserving order.
func ActiveTrailRules(xs []TrailRule) []TrailRule {
	out := make([]TrailRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailRules adds up the retained of every active entry.
func SumTrailRules(xs []TrailRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTrailRules keys the entries by ID, last write winning.
func IndexTrailRules(xs []TrailRule) map[string]TrailRule {
	out := make(map[string]TrailRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailRule builds a valid TrailRule, for tests and for the seed data the cron
// binary writes.
func SampleTrailRule(id string, at time.Time) TrailRule {
	return TrailRule{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Label:    "seeded by trail.SampleTrailRule",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailHoldSchema is the stored shape of a TrailHold.
var TrailHoldSchema = schema.New("trail.trailhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailHold is one trail hold.
type TrailHold struct {
	ID       string
	Region   string
	Owner    string
	Channel  string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailHold) Key() string { return "trail/trailhold/" + x.ID }

// Fields encodes x for storage.
func (x TrailHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailHold decodes a stored record.
func ParseTrailHold(f codec.Fields) (TrailHold, error) {
	var x TrailHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailhold: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("trail: trailhold %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("trail: trailhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailHold) String() string {
	return fmt.Sprintf("TrailHold(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailHold) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailHold) Normalize() TrailHold {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTrailHolds orders by priority descending, then by ID.
func SortTrailHolds(xs []TrailHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailHolds drops the archived entries, preserving order.
func ActiveTrailHolds(xs []TrailHold) []TrailHold {
	out := make([]TrailHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailHolds adds up the priority of every active entry.
func SumTrailHolds(xs []TrailHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTrailHolds keys the entries by ID, last write winning.
func IndexTrailHolds(xs []TrailHold) map[string]TrailHold {
	out := make(map[string]TrailHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailHold builds a valid TrailHold, for tests and for the seed data the cron
// binary writes.
func SampleTrailHold(id string, at time.Time) TrailHold {
	return TrailHold{
		ID:       id,
		Region:   "region-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by trail.SampleTrailHold",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailClaimSchema is the stored shape of a TrailClaim.
var TrailClaimSchema = schema.New("trail.trailclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailClaim is one trail claim.
type TrailClaim struct {
	ID       string
	Owner    string
	Category string
	Label    string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TrailClaim) Key() string { return "trail/trailclaim/" + x.ID }

// Fields encodes x for storage.
func (x TrailClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"category": x.Category,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailClaim decodes a stored record.
func ParseTrailClaim(f codec.Fields) (TrailClaim, error) {
	var x TrailClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("trail: trailclaim %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("trail: trailclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailClaim) String() string {
	return fmt.Sprintf("TrailClaim(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailClaim) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailClaim) Normalize() TrailClaim {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTrailClaims orders by count descending, then by ID.
func SortTrailClaims(xs []TrailClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailClaims drops the archived entries, preserving order.
func ActiveTrailClaims(xs []TrailClaim) []TrailClaim {
	out := make([]TrailClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailClaims adds up the count of every active entry.
func SumTrailClaims(xs []TrailClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTrailClaims keys the entries by ID, last write winning.
func IndexTrailClaims(xs []TrailClaim) map[string]TrailClaim {
	out := make(map[string]TrailClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailClaim builds a valid TrailClaim, for tests and for the seed data the cron
// binary writes.
func SampleTrailClaim(id string, at time.Time) TrailClaim {
	return TrailClaim{
		ID:       id,
		Owner:    "owner-" + id,
		Category: "category-" + id,
		Label:    "seeded by trail.SampleTrailClaim",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TrailCursorSchema is the stored shape of a TrailCursor.
var TrailCursorSchema = schema.New("trail.trailcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TrailCursor is one trail cursor.
type TrailCursor struct {
	ID        string
	Reference string
	Name      string
	Category  string
	Version   int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TrailCursor) Key() string { return "trail/trailcursor/" + x.ID }

// Fields encodes x for storage.
func (x TrailCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"name":      x.Name,
		"category":  x.Category,
		"version":   strconv.Itoa(x.Version),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTrailCursor decodes a stored record.
func ParseTrailCursor(f codec.Fields) (TrailCursor, error) {
	var x TrailCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("trail: trailcursor: missing id")
	}
	x.Reference = f["reference"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("trail: trailcursor %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("trail: trailcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("trail: trailcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("trail: trailcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TrailCursor) String() string {
	return fmt.Sprintf("TrailCursor(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TrailCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TrailCursor) Normalize() TrailCursor {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortTrailCursors orders by version descending, then by ID.
func SortTrailCursors(xs []TrailCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTrailCursors drops the archived entries, preserving order.
func ActiveTrailCursors(xs []TrailCursor) []TrailCursor {
	out := make([]TrailCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTrailCursors adds up the version of every active entry.
func SumTrailCursors(xs []TrailCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTrailCursors keys the entries by ID, last write winning.
func IndexTrailCursors(xs []TrailCursor) map[string]TrailCursor {
	out := make(map[string]TrailCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTrailCursor builds a valid TrailCursor, for tests and for the seed data the cron
// binary writes.
func SampleTrailCursor(id string, at time.Time) TrailCursor {
	return TrailCursor{
		ID:        id,
		Reference: "reference-" + id,
		Name:      "name-" + id,
		Category:  "seeded by trail.SampleTrailCursor",
		Version:   len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	TrailRecordSchema,
	TrailEntrySchema,
	TrailItemSchema,
	TrailDraftSchema,
	TrailRevisionSchema,
	TrailSnapshotSchema,
	TrailBatchSchema,
	TrailRequestSchema,
	TrailSummarySchema,
	TrailPolicySchema,
	TrailLinkSchema,
	TrailNoteSchema,
	TrailWindowSchema,
	TrailQuotaSchema,
	TrailBucketSchema,
	TrailLaneSchema,
	TrailSlotSchema,
	TrailTagSchema,
	TrailRuleSchema,
	TrailHoldSchema,
	TrailClaimSchema,
	TrailCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"trail/trailrecord/",
	"trail/trailentry/",
	"trail/trailitem/",
	"trail/traildraft/",
	"trail/trailrevision/",
	"trail/trailsnapshot/",
	"trail/trailbatch/",
	"trail/trailrequest/",
	"trail/trailsummary/",
	"trail/trailpolicy/",
	"trail/traillink/",
	"trail/trailnote/",
	"trail/trailwindow/",
	"trail/trailquota/",
	"trail/trailbucket/",
	"trail/traillane/",
	"trail/trailslot/",
	"trail/trailtag/",
	"trail/trailrule/",
	"trail/trailhold/",
	"trail/trailclaim/",
	"trail/trailcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "trail: " + strings.Join(names, ", ")
}
