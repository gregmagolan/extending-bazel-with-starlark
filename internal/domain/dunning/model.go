// Package dunning holds the dunning domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package dunning

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// DunningRecordSchema is the stored shape of a DunningRecord.
var DunningRecordSchema = schema.New("dunning.dunningrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningRecord is one dunning record.
type DunningRecord struct {
	ID       string
	Channel  string
	Owner    string
	Name     string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningRecord) Key() string { return "dunning/dunningrecord/" + x.ID }

// Fields encodes x for storage.
func (x DunningRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningRecord decodes a stored record.
func ParseDunningRecord(f codec.Fields) (DunningRecord, error) {
	var x DunningRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningrecord: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrecord %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningRecord) String() string {
	return fmt.Sprintf("DunningRecord(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningRecord) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningRecord) Normalize() DunningRecord {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortDunningRecords orders by count descending, then by ID.
func SortDunningRecords(xs []DunningRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningRecords drops the archived entries, preserving order.
func ActiveDunningRecords(xs []DunningRecord) []DunningRecord {
	out := make([]DunningRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningRecords adds up the count of every active entry.
func SumDunningRecords(xs []DunningRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDunningRecords keys the entries by ID, last write winning.
func IndexDunningRecords(xs []DunningRecord) map[string]DunningRecord {
	out := make(map[string]DunningRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningRecord builds a valid DunningRecord, for tests and for the seed data the cron
// binary writes.
func SampleDunningRecord(id string, at time.Time) DunningRecord {
	return DunningRecord{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by dunning.SampleDunningRecord",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningEntrySchema is the stored shape of a DunningEntry.
var DunningEntrySchema = schema.New("dunning.dunningentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningEntry is one dunning entry.
type DunningEntry struct {
	ID        string
	Reference string
	Region    string
	Channel   string
	Retained  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DunningEntry) Key() string { return "dunning/dunningentry/" + x.ID }

// Fields encodes x for storage.
func (x DunningEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"channel":   x.Channel,
		"retained":  strconv.Itoa(x.Retained),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningEntry decodes a stored record.
func ParseDunningEntry(f codec.Fields) (DunningEntry, error) {
	var x DunningEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningentry: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningentry %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningEntry) String() string {
	return fmt.Sprintf("DunningEntry(%s, reference=%q, retained=%d, archived=%t)", x.ID, x.Reference, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningEntry) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningEntry) Normalize() DunningEntry {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDunningEntrys orders by retained descending, then by ID.
func SortDunningEntrys(xs []DunningEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningEntrys drops the archived entries, preserving order.
func ActiveDunningEntrys(xs []DunningEntry) []DunningEntry {
	out := make([]DunningEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningEntrys adds up the retained of every active entry.
func SumDunningEntrys(xs []DunningEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDunningEntrys keys the entries by ID, last write winning.
func IndexDunningEntrys(xs []DunningEntry) map[string]DunningEntry {
	out := make(map[string]DunningEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningEntry builds a valid DunningEntry, for tests and for the seed data the cron
// binary writes.
func SampleDunningEntry(id string, at time.Time) DunningEntry {
	return DunningEntry{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Channel:   "seeded by dunning.SampleDunningEntry",
		Retained:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DunningItemSchema is the stored shape of a DunningItem.
var DunningItemSchema = schema.New("dunning.dunningitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningItem is one dunning item.
type DunningItem struct {
	ID       string
	Owner    string
	Group    string
	Comment  string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningItem) Key() string { return "dunning/dunningitem/" + x.ID }

// Fields encodes x for storage.
func (x DunningItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningItem decodes a stored record.
func ParseDunningItem(f codec.Fields) (DunningItem, error) {
	var x DunningItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningitem: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningItem) String() string {
	return fmt.Sprintf("DunningItem(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningItem) Normalize() DunningItem {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDunningItems orders by version descending, then by ID.
func SortDunningItems(xs []DunningItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningItems drops the archived entries, preserving order.
func ActiveDunningItems(xs []DunningItem) []DunningItem {
	out := make([]DunningItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningItems adds up the version of every active entry.
func SumDunningItems(xs []DunningItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDunningItems keys the entries by ID, last write winning.
func IndexDunningItems(xs []DunningItem) map[string]DunningItem {
	out := make(map[string]DunningItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningItem builds a valid DunningItem, for tests and for the seed data the cron
// binary writes.
func SampleDunningItem(id string, at time.Time) DunningItem {
	return DunningItem{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Comment:  "seeded by dunning.SampleDunningItem",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningDraftSchema is the stored shape of a DunningDraft.
var DunningDraftSchema = schema.New("dunning.dunningdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningDraft is one dunning draft.
type DunningDraft struct {
	ID       string
	Source   string
	Channel  string
	Category string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningDraft) Key() string { return "dunning/dunningdraft/" + x.ID }

// Fields encodes x for storage.
func (x DunningDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningDraft decodes a stored record.
func ParseDunningDraft(f codec.Fields) (DunningDraft, error) {
	var x DunningDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningdraft: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningdraft %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningDraft) String() string {
	return fmt.Sprintf("DunningDraft(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningDraft) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningDraft) Normalize() DunningDraft {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDunningDrafts orders by capacity descending, then by ID.
func SortDunningDrafts(xs []DunningDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningDrafts drops the archived entries, preserving order.
func ActiveDunningDrafts(xs []DunningDraft) []DunningDraft {
	out := make([]DunningDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningDrafts adds up the capacity of every active entry.
func SumDunningDrafts(xs []DunningDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDunningDrafts keys the entries by ID, last write winning.
func IndexDunningDrafts(xs []DunningDraft) map[string]DunningDraft {
	out := make(map[string]DunningDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningDraft builds a valid DunningDraft, for tests and for the seed data the cron
// binary writes.
func SampleDunningDraft(id string, at time.Time) DunningDraft {
	return DunningDraft{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by dunning.SampleDunningDraft",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningRevisionSchema is the stored shape of a DunningRevision.
var DunningRevisionSchema = schema.New("dunning.dunningrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningRevision is one dunning revision.
type DunningRevision struct {
	ID       string
	Owner    string
	Comment  string
	Label    string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningRevision) Key() string { return "dunning/dunningrevision/" + x.ID }

// Fields encodes x for storage.
func (x DunningRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningRevision decodes a stored record.
func ParseDunningRevision(f codec.Fields) (DunningRevision, error) {
	var x DunningRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningrevision: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrevision %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningRevision) String() string {
	return fmt.Sprintf("DunningRevision(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningRevision) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningRevision) Normalize() DunningRevision {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDunningRevisions orders by capacity descending, then by ID.
func SortDunningRevisions(xs []DunningRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningRevisions drops the archived entries, preserving order.
func ActiveDunningRevisions(xs []DunningRevision) []DunningRevision {
	out := make([]DunningRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningRevisions adds up the capacity of every active entry.
func SumDunningRevisions(xs []DunningRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDunningRevisions keys the entries by ID, last write winning.
func IndexDunningRevisions(xs []DunningRevision) map[string]DunningRevision {
	out := make(map[string]DunningRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningRevision builds a valid DunningRevision, for tests and for the seed data the cron
// binary writes.
func SampleDunningRevision(id string, at time.Time) DunningRevision {
	return DunningRevision{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Label:    "seeded by dunning.SampleDunningRevision",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningSnapshotSchema is the stored shape of a DunningSnapshot.
var DunningSnapshotSchema = schema.New("dunning.dunningsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningSnapshot is one dunning snapshot.
type DunningSnapshot struct {
	ID       string
	Name     string
	Category string
	Group    string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningSnapshot) Key() string { return "dunning/dunningsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x DunningSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningSnapshot decodes a stored record.
func ParseDunningSnapshot(f codec.Fields) (DunningSnapshot, error) {
	var x DunningSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningsnapshot: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningsnapshot %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningSnapshot) String() string {
	return fmt.Sprintf("DunningSnapshot(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningSnapshot) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningSnapshot) Normalize() DunningSnapshot {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDunningSnapshots orders by version descending, then by ID.
func SortDunningSnapshots(xs []DunningSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningSnapshots drops the archived entries, preserving order.
func ActiveDunningSnapshots(xs []DunningSnapshot) []DunningSnapshot {
	out := make([]DunningSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningSnapshots adds up the version of every active entry.
func SumDunningSnapshots(xs []DunningSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDunningSnapshots keys the entries by ID, last write winning.
func IndexDunningSnapshots(xs []DunningSnapshot) map[string]DunningSnapshot {
	out := make(map[string]DunningSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningSnapshot builds a valid DunningSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleDunningSnapshot(id string, at time.Time) DunningSnapshot {
	return DunningSnapshot{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Group:    "seeded by dunning.SampleDunningSnapshot",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningBatchSchema is the stored shape of a DunningBatch.
var DunningBatchSchema = schema.New("dunning.dunningbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningBatch is one dunning batch.
type DunningBatch struct {
	ID       string
	Group    string
	Category string
	Owner    string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningBatch) Key() string { return "dunning/dunningbatch/" + x.ID }

// Fields encodes x for storage.
func (x DunningBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningBatch decodes a stored record.
func ParseDunningBatch(f codec.Fields) (DunningBatch, error) {
	var x DunningBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningbatch: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningbatch %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningBatch) String() string {
	return fmt.Sprintf("DunningBatch(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningBatch) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningBatch) Normalize() DunningBatch {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDunningBatchs orders by offset descending, then by ID.
func SortDunningBatchs(xs []DunningBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningBatchs drops the archived entries, preserving order.
func ActiveDunningBatchs(xs []DunningBatch) []DunningBatch {
	out := make([]DunningBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningBatchs adds up the offset of every active entry.
func SumDunningBatchs(xs []DunningBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDunningBatchs keys the entries by ID, last write winning.
func IndexDunningBatchs(xs []DunningBatch) map[string]DunningBatch {
	out := make(map[string]DunningBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningBatch builds a valid DunningBatch, for tests and for the seed data the cron
// binary writes.
func SampleDunningBatch(id string, at time.Time) DunningBatch {
	return DunningBatch{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Owner:    "seeded by dunning.SampleDunningBatch",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningRequestSchema is the stored shape of a DunningRequest.
var DunningRequestSchema = schema.New("dunning.dunningrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningRequest is one dunning request.
type DunningRequest struct {
	ID       string
	Channel  string
	Comment  string
	Region   string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningRequest) Key() string { return "dunning/dunningrequest/" + x.ID }

// Fields encodes x for storage.
func (x DunningRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"region":   x.Region,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningRequest decodes a stored record.
func ParseDunningRequest(f codec.Fields) (DunningRequest, error) {
	var x DunningRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningrequest: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningRequest) String() string {
	return fmt.Sprintf("DunningRequest(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningRequest) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningRequest) Normalize() DunningRequest {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDunningRequests orders by version descending, then by ID.
func SortDunningRequests(xs []DunningRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningRequests drops the archived entries, preserving order.
func ActiveDunningRequests(xs []DunningRequest) []DunningRequest {
	out := make([]DunningRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningRequests adds up the version of every active entry.
func SumDunningRequests(xs []DunningRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDunningRequests keys the entries by ID, last write winning.
func IndexDunningRequests(xs []DunningRequest) map[string]DunningRequest {
	out := make(map[string]DunningRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningRequest builds a valid DunningRequest, for tests and for the seed data the cron
// binary writes.
func SampleDunningRequest(id string, at time.Time) DunningRequest {
	return DunningRequest{
		ID:       id,
		Channel:  "channel-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by dunning.SampleDunningRequest",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningSummarySchema is the stored shape of a DunningSummary.
var DunningSummarySchema = schema.New("dunning.dunningsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningSummary is one dunning summary.
type DunningSummary struct {
	ID       string
	Owner    string
	Source   string
	Group    string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningSummary) Key() string { return "dunning/dunningsummary/" + x.ID }

// Fields encodes x for storage.
func (x DunningSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"group":    x.Group,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningSummary decodes a stored record.
func ParseDunningSummary(f codec.Fields) (DunningSummary, error) {
	var x DunningSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningsummary: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningsummary %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningSummary) String() string {
	return fmt.Sprintf("DunningSummary(%s, owner=%q, weight=%d, archived=%t)", x.ID, x.Owner, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningSummary) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningSummary) Normalize() DunningSummary {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortDunningSummarys orders by weight descending, then by ID.
func SortDunningSummarys(xs []DunningSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningSummarys drops the archived entries, preserving order.
func ActiveDunningSummarys(xs []DunningSummary) []DunningSummary {
	out := make([]DunningSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningSummarys adds up the weight of every active entry.
func SumDunningSummarys(xs []DunningSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDunningSummarys keys the entries by ID, last write winning.
func IndexDunningSummarys(xs []DunningSummary) map[string]DunningSummary {
	out := make(map[string]DunningSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningSummary builds a valid DunningSummary, for tests and for the seed data the cron
// binary writes.
func SampleDunningSummary(id string, at time.Time) DunningSummary {
	return DunningSummary{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Group:    "seeded by dunning.SampleDunningSummary",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningPolicySchema is the stored shape of a DunningPolicy.
var DunningPolicySchema = schema.New("dunning.dunningpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningPolicy is one dunning policy.
type DunningPolicy struct {
	ID        string
	Region    string
	Owner     string
	Reference string
	Attempts  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DunningPolicy) Key() string { return "dunning/dunningpolicy/" + x.ID }

// Fields encodes x for storage.
func (x DunningPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"owner":     x.Owner,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningPolicy decodes a stored record.
func ParseDunningPolicy(f codec.Fields) (DunningPolicy, error) {
	var x DunningPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningpolicy: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningpolicy %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningPolicy) String() string {
	return fmt.Sprintf("DunningPolicy(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningPolicy) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningPolicy) Normalize() DunningPolicy {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDunningPolicys orders by attempts descending, then by ID.
func SortDunningPolicys(xs []DunningPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningPolicys drops the archived entries, preserving order.
func ActiveDunningPolicys(xs []DunningPolicy) []DunningPolicy {
	out := make([]DunningPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningPolicys adds up the attempts of every active entry.
func SumDunningPolicys(xs []DunningPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDunningPolicys keys the entries by ID, last write winning.
func IndexDunningPolicys(xs []DunningPolicy) map[string]DunningPolicy {
	out := make(map[string]DunningPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningPolicy builds a valid DunningPolicy, for tests and for the seed data the cron
// binary writes.
func SampleDunningPolicy(id string, at time.Time) DunningPolicy {
	return DunningPolicy{
		ID:        id,
		Region:    "region-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by dunning.SampleDunningPolicy",
		Attempts:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DunningLinkSchema is the stored shape of a DunningLink.
var DunningLinkSchema = schema.New("dunning.dunninglink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningLink is one dunning link.
type DunningLink struct {
	ID        string
	Name      string
	Reference string
	Owner     string
	Priority  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DunningLink) Key() string { return "dunning/dunninglink/" + x.ID }

// Fields encodes x for storage.
func (x DunningLink) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"owner":     x.Owner,
		"priority":  strconv.Itoa(x.Priority),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningLink decodes a stored record.
func ParseDunningLink(f codec.Fields) (DunningLink, error) {
	var x DunningLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunninglink: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("dunning: dunninglink %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunninglink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunninglink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunninglink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningLink) String() string {
	return fmt.Sprintf("DunningLink(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningLink) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningLink) Normalize() DunningLink {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDunningLinks orders by priority descending, then by ID.
func SortDunningLinks(xs []DunningLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningLinks drops the archived entries, preserving order.
func ActiveDunningLinks(xs []DunningLink) []DunningLink {
	out := make([]DunningLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningLinks adds up the priority of every active entry.
func SumDunningLinks(xs []DunningLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDunningLinks keys the entries by ID, last write winning.
func IndexDunningLinks(xs []DunningLink) map[string]DunningLink {
	out := make(map[string]DunningLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningLink builds a valid DunningLink, for tests and for the seed data the cron
// binary writes.
func SampleDunningLink(id string, at time.Time) DunningLink {
	return DunningLink{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by dunning.SampleDunningLink",
		Priority:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DunningNoteSchema is the stored shape of a DunningNote.
var DunningNoteSchema = schema.New("dunning.dunningnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningNote is one dunning note.
type DunningNote struct {
	ID       string
	Comment  string
	Owner    string
	Category string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningNote) Key() string { return "dunning/dunningnote/" + x.ID }

// Fields encodes x for storage.
func (x DunningNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"category": x.Category,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningNote decodes a stored record.
func ParseDunningNote(f codec.Fields) (DunningNote, error) {
	var x DunningNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningnote: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningnote %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningNote) String() string {
	return fmt.Sprintf("DunningNote(%s, comment=%q, weight=%d, archived=%t)", x.ID, x.Comment, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningNote) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningNote) Normalize() DunningNote {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortDunningNotes orders by weight descending, then by ID.
func SortDunningNotes(xs []DunningNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningNotes drops the archived entries, preserving order.
func ActiveDunningNotes(xs []DunningNote) []DunningNote {
	out := make([]DunningNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningNotes adds up the weight of every active entry.
func SumDunningNotes(xs []DunningNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDunningNotes keys the entries by ID, last write winning.
func IndexDunningNotes(xs []DunningNote) map[string]DunningNote {
	out := make(map[string]DunningNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningNote builds a valid DunningNote, for tests and for the seed data the cron
// binary writes.
func SampleDunningNote(id string, at time.Time) DunningNote {
	return DunningNote{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by dunning.SampleDunningNote",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningWindowSchema is the stored shape of a DunningWindow.
var DunningWindowSchema = schema.New("dunning.dunningwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningWindow is one dunning window.
type DunningWindow struct {
	ID        string
	Source    string
	Group     string
	Reference string
	Count     int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DunningWindow) Key() string { return "dunning/dunningwindow/" + x.ID }

// Fields encodes x for storage.
func (x DunningWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"group":     x.Group,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningWindow decodes a stored record.
func ParseDunningWindow(f codec.Fields) (DunningWindow, error) {
	var x DunningWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningwindow: missing id")
	}
	x.Source = f["source"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningwindow %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningWindow) String() string {
	return fmt.Sprintf("DunningWindow(%s, source=%q, count=%d, archived=%t)", x.ID, x.Source, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningWindow) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningWindow) Normalize() DunningWindow {
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortDunningWindows orders by count descending, then by ID.
func SortDunningWindows(xs []DunningWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningWindows drops the archived entries, preserving order.
func ActiveDunningWindows(xs []DunningWindow) []DunningWindow {
	out := make([]DunningWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningWindows adds up the count of every active entry.
func SumDunningWindows(xs []DunningWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDunningWindows keys the entries by ID, last write winning.
func IndexDunningWindows(xs []DunningWindow) map[string]DunningWindow {
	out := make(map[string]DunningWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningWindow builds a valid DunningWindow, for tests and for the seed data the cron
// binary writes.
func SampleDunningWindow(id string, at time.Time) DunningWindow {
	return DunningWindow{
		ID:        id,
		Source:    "source-" + id,
		Group:     "group-" + id,
		Reference: "seeded by dunning.SampleDunningWindow",
		Count:     len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DunningQuotaSchema is the stored shape of a DunningQuota.
var DunningQuotaSchema = schema.New("dunning.dunningquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningQuota is one dunning quota.
type DunningQuota struct {
	ID       string
	Name     string
	Region   string
	Group    string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningQuota) Key() string { return "dunning/dunningquota/" + x.ID }

// Fields encodes x for storage.
func (x DunningQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"region":   x.Region,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningQuota decodes a stored record.
func ParseDunningQuota(f codec.Fields) (DunningQuota, error) {
	var x DunningQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningquota: missing id")
	}
	x.Name = f["name"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningquota %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningQuota) String() string {
	return fmt.Sprintf("DunningQuota(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningQuota) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningQuota) Normalize() DunningQuota {
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDunningQuotas orders by offset descending, then by ID.
func SortDunningQuotas(xs []DunningQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningQuotas drops the archived entries, preserving order.
func ActiveDunningQuotas(xs []DunningQuota) []DunningQuota {
	out := make([]DunningQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningQuotas adds up the offset of every active entry.
func SumDunningQuotas(xs []DunningQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDunningQuotas keys the entries by ID, last write winning.
func IndexDunningQuotas(xs []DunningQuota) map[string]DunningQuota {
	out := make(map[string]DunningQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningQuota builds a valid DunningQuota, for tests and for the seed data the cron
// binary writes.
func SampleDunningQuota(id string, at time.Time) DunningQuota {
	return DunningQuota{
		ID:       id,
		Name:     "name-" + id,
		Region:   "region-" + id,
		Group:    "seeded by dunning.SampleDunningQuota",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningBucketSchema is the stored shape of a DunningBucket.
var DunningBucketSchema = schema.New("dunning.dunningbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningBucket is one dunning bucket.
type DunningBucket struct {
	ID       string
	Group    string
	Name     string
	Region   string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningBucket) Key() string { return "dunning/dunningbucket/" + x.ID }

// Fields encodes x for storage.
func (x DunningBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningBucket decodes a stored record.
func ParseDunningBucket(f codec.Fields) (DunningBucket, error) {
	var x DunningBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningbucket: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningbucket %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningBucket) String() string {
	return fmt.Sprintf("DunningBucket(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningBucket) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningBucket) Normalize() DunningBucket {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortDunningBuckets orders by offset descending, then by ID.
func SortDunningBuckets(xs []DunningBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningBuckets drops the archived entries, preserving order.
func ActiveDunningBuckets(xs []DunningBucket) []DunningBucket {
	out := make([]DunningBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningBuckets adds up the offset of every active entry.
func SumDunningBuckets(xs []DunningBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDunningBuckets keys the entries by ID, last write winning.
func IndexDunningBuckets(xs []DunningBucket) map[string]DunningBucket {
	out := make(map[string]DunningBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningBucket builds a valid DunningBucket, for tests and for the seed data the cron
// binary writes.
func SampleDunningBucket(id string, at time.Time) DunningBucket {
	return DunningBucket{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Region:   "seeded by dunning.SampleDunningBucket",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningLaneSchema is the stored shape of a DunningLane.
var DunningLaneSchema = schema.New("dunning.dunninglane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningLane is one dunning lane.
type DunningLane struct {
	ID       string
	Category string
	Label    string
	Owner    string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningLane) Key() string { return "dunning/dunninglane/" + x.ID }

// Fields encodes x for storage.
func (x DunningLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"owner":    x.Owner,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningLane decodes a stored record.
func ParseDunningLane(f codec.Fields) (DunningLane, error) {
	var x DunningLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunninglane: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Owner = f["owner"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("dunning: dunninglane %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunninglane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunninglane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunninglane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningLane) String() string {
	return fmt.Sprintf("DunningLane(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningLane) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningLane) Normalize() DunningLane {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDunningLanes orders by version descending, then by ID.
func SortDunningLanes(xs []DunningLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningLanes drops the archived entries, preserving order.
func ActiveDunningLanes(xs []DunningLane) []DunningLane {
	out := make([]DunningLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningLanes adds up the version of every active entry.
func SumDunningLanes(xs []DunningLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDunningLanes keys the entries by ID, last write winning.
func IndexDunningLanes(xs []DunningLane) map[string]DunningLane {
	out := make(map[string]DunningLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningLane builds a valid DunningLane, for tests and for the seed data the cron
// binary writes.
func SampleDunningLane(id string, at time.Time) DunningLane {
	return DunningLane{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Owner:    "seeded by dunning.SampleDunningLane",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningSlotSchema is the stored shape of a DunningSlot.
var DunningSlotSchema = schema.New("dunning.dunningslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningSlot is one dunning slot.
type DunningSlot struct {
	ID       string
	Region   string
	Group    string
	Label    string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningSlot) Key() string { return "dunning/dunningslot/" + x.ID }

// Fields encodes x for storage.
func (x DunningSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"group":    x.Group,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningSlot decodes a stored record.
func ParseDunningSlot(f codec.Fields) (DunningSlot, error) {
	var x DunningSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningslot: missing id")
	}
	x.Region = f["region"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningslot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningSlot) String() string {
	return fmt.Sprintf("DunningSlot(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningSlot) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningSlot) Normalize() DunningSlot {
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDunningSlots orders by attempts descending, then by ID.
func SortDunningSlots(xs []DunningSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningSlots drops the archived entries, preserving order.
func ActiveDunningSlots(xs []DunningSlot) []DunningSlot {
	out := make([]DunningSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningSlots adds up the attempts of every active entry.
func SumDunningSlots(xs []DunningSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDunningSlots keys the entries by ID, last write winning.
func IndexDunningSlots(xs []DunningSlot) map[string]DunningSlot {
	out := make(map[string]DunningSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningSlot builds a valid DunningSlot, for tests and for the seed data the cron
// binary writes.
func SampleDunningSlot(id string, at time.Time) DunningSlot {
	return DunningSlot{
		ID:       id,
		Region:   "region-" + id,
		Group:    "group-" + id,
		Label:    "seeded by dunning.SampleDunningSlot",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningTagSchema is the stored shape of a DunningTag.
var DunningTagSchema = schema.New("dunning.dunningtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningTag is one dunning tag.
type DunningTag struct {
	ID       string
	Comment  string
	Name     string
	Channel  string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningTag) Key() string { return "dunning/dunningtag/" + x.ID }

// Fields encodes x for storage.
func (x DunningTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"name":     x.Name,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningTag decodes a stored record.
func ParseDunningTag(f codec.Fields) (DunningTag, error) {
	var x DunningTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningtag: missing id")
	}
	x.Comment = f["comment"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningtag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningTag) String() string {
	return fmt.Sprintf("DunningTag(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningTag) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningTag) Normalize() DunningTag {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDunningTags orders by priority descending, then by ID.
func SortDunningTags(xs []DunningTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningTags drops the archived entries, preserving order.
func ActiveDunningTags(xs []DunningTag) []DunningTag {
	out := make([]DunningTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningTags adds up the priority of every active entry.
func SumDunningTags(xs []DunningTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDunningTags keys the entries by ID, last write winning.
func IndexDunningTags(xs []DunningTag) map[string]DunningTag {
	out := make(map[string]DunningTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningTag builds a valid DunningTag, for tests and for the seed data the cron
// binary writes.
func SampleDunningTag(id string, at time.Time) DunningTag {
	return DunningTag{
		ID:       id,
		Comment:  "comment-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by dunning.SampleDunningTag",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningRuleSchema is the stored shape of a DunningRule.
var DunningRuleSchema = schema.New("dunning.dunningrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningRule is one dunning rule.
type DunningRule struct {
	ID       string
	Channel  string
	Region   string
	Category string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningRule) Key() string { return "dunning/dunningrule/" + x.ID }

// Fields encodes x for storage.
func (x DunningRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningRule decodes a stored record.
func ParseDunningRule(f codec.Fields) (DunningRule, error) {
	var x DunningRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningrule: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrule %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningRule) String() string {
	return fmt.Sprintf("DunningRule(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningRule) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningRule) Normalize() DunningRule {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDunningRules orders by count descending, then by ID.
func SortDunningRules(xs []DunningRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningRules drops the archived entries, preserving order.
func ActiveDunningRules(xs []DunningRule) []DunningRule {
	out := make([]DunningRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningRules adds up the count of every active entry.
func SumDunningRules(xs []DunningRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDunningRules keys the entries by ID, last write winning.
func IndexDunningRules(xs []DunningRule) map[string]DunningRule {
	out := make(map[string]DunningRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningRule builds a valid DunningRule, for tests and for the seed data the cron
// binary writes.
func SampleDunningRule(id string, at time.Time) DunningRule {
	return DunningRule{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Category: "seeded by dunning.SampleDunningRule",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningHoldSchema is the stored shape of a DunningHold.
var DunningHoldSchema = schema.New("dunning.dunninghold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningHold is one dunning hold.
type DunningHold struct {
	ID        string
	Channel   string
	Reference string
	Region    string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DunningHold) Key() string { return "dunning/dunninghold/" + x.ID }

// Fields encodes x for storage.
func (x DunningHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"region":    x.Region,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningHold decodes a stored record.
func ParseDunningHold(f codec.Fields) (DunningHold, error) {
	var x DunningHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunninghold: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("dunning: dunninghold %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunninghold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunninghold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunninghold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningHold) String() string {
	return fmt.Sprintf("DunningHold(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningHold) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningHold) Normalize() DunningHold {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDunningHolds orders by offset descending, then by ID.
func SortDunningHolds(xs []DunningHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningHolds drops the archived entries, preserving order.
func ActiveDunningHolds(xs []DunningHold) []DunningHold {
	out := make([]DunningHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningHolds adds up the offset of every active entry.
func SumDunningHolds(xs []DunningHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDunningHolds keys the entries by ID, last write winning.
func IndexDunningHolds(xs []DunningHold) map[string]DunningHold {
	out := make(map[string]DunningHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningHold builds a valid DunningHold, for tests and for the seed data the cron
// binary writes.
func SampleDunningHold(id string, at time.Time) DunningHold {
	return DunningHold{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by dunning.SampleDunningHold",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DunningClaimSchema is the stored shape of a DunningClaim.
var DunningClaimSchema = schema.New("dunning.dunningclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningClaim is one dunning claim.
type DunningClaim struct {
	ID       string
	Group    string
	Source   string
	Label    string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningClaim) Key() string { return "dunning/dunningclaim/" + x.ID }

// Fields encodes x for storage.
func (x DunningClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningClaim decodes a stored record.
func ParseDunningClaim(f codec.Fields) (DunningClaim, error) {
	var x DunningClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningclaim: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningClaim) String() string {
	return fmt.Sprintf("DunningClaim(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningClaim) Normalize() DunningClaim {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDunningClaims orders by retained descending, then by ID.
func SortDunningClaims(xs []DunningClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningClaims drops the archived entries, preserving order.
func ActiveDunningClaims(xs []DunningClaim) []DunningClaim {
	out := make([]DunningClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningClaims adds up the retained of every active entry.
func SumDunningClaims(xs []DunningClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDunningClaims keys the entries by ID, last write winning.
func IndexDunningClaims(xs []DunningClaim) map[string]DunningClaim {
	out := make(map[string]DunningClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningClaim builds a valid DunningClaim, for tests and for the seed data the cron
// binary writes.
func SampleDunningClaim(id string, at time.Time) DunningClaim {
	return DunningClaim{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Label:    "seeded by dunning.SampleDunningClaim",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DunningCursorSchema is the stored shape of a DunningCursor.
var DunningCursorSchema = schema.New("dunning.dunningcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DunningCursor is one dunning cursor.
type DunningCursor struct {
	ID       string
	Label    string
	Comment  string
	Group    string
	Retained int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DunningCursor) Key() string { return "dunning/dunningcursor/" + x.ID }

// Fields encodes x for storage.
func (x DunningCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"comment":  x.Comment,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDunningCursor decodes a stored record.
func ParseDunningCursor(f codec.Fields) (DunningCursor, error) {
	var x DunningCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("dunning: dunningcursor: missing id")
	}
	x.Label = f["label"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningcursor %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("dunning: dunningcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("dunning: dunningcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("dunning: dunningcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DunningCursor) String() string {
	return fmt.Sprintf("DunningCursor(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DunningCursor) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DunningCursor) Normalize() DunningCursor {
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDunningCursors orders by retained descending, then by ID.
func SortDunningCursors(xs []DunningCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDunningCursors drops the archived entries, preserving order.
func ActiveDunningCursors(xs []DunningCursor) []DunningCursor {
	out := make([]DunningCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDunningCursors adds up the retained of every active entry.
func SumDunningCursors(xs []DunningCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDunningCursors keys the entries by ID, last write winning.
func IndexDunningCursors(xs []DunningCursor) map[string]DunningCursor {
	out := make(map[string]DunningCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDunningCursor builds a valid DunningCursor, for tests and for the seed data the cron
// binary writes.
func SampleDunningCursor(id string, at time.Time) DunningCursor {
	return DunningCursor{
		ID:       id,
		Label:    "label-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by dunning.SampleDunningCursor",
		Retained: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	DunningRecordSchema,
	DunningEntrySchema,
	DunningItemSchema,
	DunningDraftSchema,
	DunningRevisionSchema,
	DunningSnapshotSchema,
	DunningBatchSchema,
	DunningRequestSchema,
	DunningSummarySchema,
	DunningPolicySchema,
	DunningLinkSchema,
	DunningNoteSchema,
	DunningWindowSchema,
	DunningQuotaSchema,
	DunningBucketSchema,
	DunningLaneSchema,
	DunningSlotSchema,
	DunningTagSchema,
	DunningRuleSchema,
	DunningHoldSchema,
	DunningClaimSchema,
	DunningCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"dunning/dunningrecord/",
	"dunning/dunningentry/",
	"dunning/dunningitem/",
	"dunning/dunningdraft/",
	"dunning/dunningrevision/",
	"dunning/dunningsnapshot/",
	"dunning/dunningbatch/",
	"dunning/dunningrequest/",
	"dunning/dunningsummary/",
	"dunning/dunningpolicy/",
	"dunning/dunninglink/",
	"dunning/dunningnote/",
	"dunning/dunningwindow/",
	"dunning/dunningquota/",
	"dunning/dunningbucket/",
	"dunning/dunninglane/",
	"dunning/dunningslot/",
	"dunning/dunningtag/",
	"dunning/dunningrule/",
	"dunning/dunninghold/",
	"dunning/dunningclaim/",
	"dunning/dunningcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "dunning: " + strings.Join(names, ", ")
}
