// Package inventory holds the inventory domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package inventory

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// InventoryRecordSchema is the stored shape of a InventoryRecord.
var InventoryRecordSchema = schema.New("inventory.inventoryrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryRecord is one inventory record.
type InventoryRecord struct {
	ID       string
	Region   string
	Owner    string
	Category string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryRecord) Key() string { return "inventory/inventoryrecord/" + x.ID }

// Fields encodes x for storage.
func (x InventoryRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"owner":    x.Owner,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryRecord decodes a stored record.
func ParseInventoryRecord(f codec.Fields) (InventoryRecord, error) {
	var x InventoryRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryrecord: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrecord %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryRecord) String() string {
	return fmt.Sprintf("InventoryRecord(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryRecord) Normalize() InventoryRecord {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortInventoryRecords orders by capacity descending, then by ID.
func SortInventoryRecords(xs []InventoryRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryRecords drops the archived entries, preserving order.
func ActiveInventoryRecords(xs []InventoryRecord) []InventoryRecord {
	out := make([]InventoryRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryRecords adds up the capacity of every active entry.
func SumInventoryRecords(xs []InventoryRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexInventoryRecords keys the entries by ID, last write winning.
func IndexInventoryRecords(xs []InventoryRecord) map[string]InventoryRecord {
	out := make(map[string]InventoryRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryRecord builds a valid InventoryRecord, for tests and for the seed data the cron
// binary writes.
func SampleInventoryRecord(id string, at time.Time) InventoryRecord {
	return InventoryRecord{
		ID:       id,
		Region:   "region-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by inventory.SampleInventoryRecord",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryEntrySchema is the stored shape of a InventoryEntry.
var InventoryEntrySchema = schema.New("inventory.inventoryentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryEntry is one inventory entry.
type InventoryEntry struct {
	ID        string
	Reference string
	Comment   string
	Owner     string
	Priority  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryEntry) Key() string { return "inventory/inventoryentry/" + x.ID }

// Fields encodes x for storage.
func (x InventoryEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"comment":   x.Comment,
		"owner":     x.Owner,
		"priority":  strconv.Itoa(x.Priority),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryEntry decodes a stored record.
func ParseInventoryEntry(f codec.Fields) (InventoryEntry, error) {
	var x InventoryEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryentry: missing id")
	}
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryentry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryEntry) String() string {
	return fmt.Sprintf("InventoryEntry(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryEntry) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryEntry) Normalize() InventoryEntry {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortInventoryEntrys orders by priority descending, then by ID.
func SortInventoryEntrys(xs []InventoryEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryEntrys drops the archived entries, preserving order.
func ActiveInventoryEntrys(xs []InventoryEntry) []InventoryEntry {
	out := make([]InventoryEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryEntrys adds up the priority of every active entry.
func SumInventoryEntrys(xs []InventoryEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInventoryEntrys keys the entries by ID, last write winning.
func IndexInventoryEntrys(xs []InventoryEntry) map[string]InventoryEntry {
	out := make(map[string]InventoryEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryEntry builds a valid InventoryEntry, for tests and for the seed data the cron
// binary writes.
func SampleInventoryEntry(id string, at time.Time) InventoryEntry {
	return InventoryEntry{
		ID:        id,
		Reference: "reference-" + id,
		Comment:   "comment-" + id,
		Owner:     "seeded by inventory.SampleInventoryEntry",
		Priority:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventoryItemSchema is the stored shape of a InventoryItem.
var InventoryItemSchema = schema.New("inventory.inventoryitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryItem is one inventory item.
type InventoryItem struct {
	ID        string
	Reference string
	Source    string
	Channel   string
	Weight    int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryItem) Key() string { return "inventory/inventoryitem/" + x.ID }

// Fields encodes x for storage.
func (x InventoryItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"source":    x.Source,
		"channel":   x.Channel,
		"weight":    strconv.Itoa(x.Weight),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryItem decodes a stored record.
func ParseInventoryItem(f codec.Fields) (InventoryItem, error) {
	var x InventoryItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryitem: missing id")
	}
	x.Reference = f["reference"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryitem %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryItem) String() string {
	return fmt.Sprintf("InventoryItem(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryItem) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryItem) Normalize() InventoryItem {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortInventoryItems orders by weight descending, then by ID.
func SortInventoryItems(xs []InventoryItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryItems drops the archived entries, preserving order.
func ActiveInventoryItems(xs []InventoryItem) []InventoryItem {
	out := make([]InventoryItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryItems adds up the weight of every active entry.
func SumInventoryItems(xs []InventoryItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexInventoryItems keys the entries by ID, last write winning.
func IndexInventoryItems(xs []InventoryItem) map[string]InventoryItem {
	out := make(map[string]InventoryItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryItem builds a valid InventoryItem, for tests and for the seed data the cron
// binary writes.
func SampleInventoryItem(id string, at time.Time) InventoryItem {
	return InventoryItem{
		ID:        id,
		Reference: "reference-" + id,
		Source:    "source-" + id,
		Channel:   "seeded by inventory.SampleInventoryItem",
		Weight:    len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventoryDraftSchema is the stored shape of a InventoryDraft.
var InventoryDraftSchema = schema.New("inventory.inventorydraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryDraft is one inventory draft.
type InventoryDraft struct {
	ID       string
	Group    string
	Comment  string
	Source   string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryDraft) Key() string { return "inventory/inventorydraft/" + x.ID }

// Fields encodes x for storage.
func (x InventoryDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"comment":  x.Comment,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryDraft decodes a stored record.
func ParseInventoryDraft(f codec.Fields) (InventoryDraft, error) {
	var x InventoryDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorydraft: missing id")
	}
	x.Group = f["group"]
	x.Comment = f["comment"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorydraft %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorydraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorydraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorydraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryDraft) String() string {
	return fmt.Sprintf("InventoryDraft(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryDraft) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryDraft) Normalize() InventoryDraft {
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortInventoryDrafts orders by offset descending, then by ID.
func SortInventoryDrafts(xs []InventoryDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryDrafts drops the archived entries, preserving order.
func ActiveInventoryDrafts(xs []InventoryDraft) []InventoryDraft {
	out := make([]InventoryDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryDrafts adds up the offset of every active entry.
func SumInventoryDrafts(xs []InventoryDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexInventoryDrafts keys the entries by ID, last write winning.
func IndexInventoryDrafts(xs []InventoryDraft) map[string]InventoryDraft {
	out := make(map[string]InventoryDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryDraft builds a valid InventoryDraft, for tests and for the seed data the cron
// binary writes.
func SampleInventoryDraft(id string, at time.Time) InventoryDraft {
	return InventoryDraft{
		ID:       id,
		Group:    "group-" + id,
		Comment:  "comment-" + id,
		Source:   "seeded by inventory.SampleInventoryDraft",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryRevisionSchema is the stored shape of a InventoryRevision.
var InventoryRevisionSchema = schema.New("inventory.inventoryrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryRevision is one inventory revision.
type InventoryRevision struct {
	ID        string
	Name      string
	Source    string
	Reference string
	Offset    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryRevision) Key() string { return "inventory/inventoryrevision/" + x.ID }

// Fields encodes x for storage.
func (x InventoryRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"source":    x.Source,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryRevision decodes a stored record.
func ParseInventoryRevision(f codec.Fields) (InventoryRevision, error) {
	var x InventoryRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryrevision: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrevision %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryRevision) String() string {
	return fmt.Sprintf("InventoryRevision(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryRevision) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryRevision) Normalize() InventoryRevision {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortInventoryRevisions orders by offset descending, then by ID.
func SortInventoryRevisions(xs []InventoryRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryRevisions drops the archived entries, preserving order.
func ActiveInventoryRevisions(xs []InventoryRevision) []InventoryRevision {
	out := make([]InventoryRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryRevisions adds up the offset of every active entry.
func SumInventoryRevisions(xs []InventoryRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexInventoryRevisions keys the entries by ID, last write winning.
func IndexInventoryRevisions(xs []InventoryRevision) map[string]InventoryRevision {
	out := make(map[string]InventoryRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryRevision builds a valid InventoryRevision, for tests and for the seed data the cron
// binary writes.
func SampleInventoryRevision(id string, at time.Time) InventoryRevision {
	return InventoryRevision{
		ID:        id,
		Name:      "name-" + id,
		Source:    "source-" + id,
		Reference: "seeded by inventory.SampleInventoryRevision",
		Offset:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventorySnapshotSchema is the stored shape of a InventorySnapshot.
var InventorySnapshotSchema = schema.New("inventory.inventorysnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventorySnapshot is one inventory snapshot.
type InventorySnapshot struct {
	ID       string
	Name     string
	Label    string
	Source   string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventorySnapshot) Key() string { return "inventory/inventorysnapshot/" + x.ID }

// Fields encodes x for storage.
func (x InventorySnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventorySnapshot decodes a stored record.
func ParseInventorySnapshot(f codec.Fields) (InventorySnapshot, error) {
	var x InventorySnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorysnapshot: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorysnapshot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorysnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorysnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorysnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventorySnapshot) String() string {
	return fmt.Sprintf("InventorySnapshot(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventorySnapshot) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventorySnapshot) Normalize() InventorySnapshot {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInventorySnapshots orders by attempts descending, then by ID.
func SortInventorySnapshots(xs []InventorySnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventorySnapshots drops the archived entries, preserving order.
func ActiveInventorySnapshots(xs []InventorySnapshot) []InventorySnapshot {
	out := make([]InventorySnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventorySnapshots adds up the attempts of every active entry.
func SumInventorySnapshots(xs []InventorySnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexInventorySnapshots keys the entries by ID, last write winning.
func IndexInventorySnapshots(xs []InventorySnapshot) map[string]InventorySnapshot {
	out := make(map[string]InventorySnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventorySnapshot builds a valid InventorySnapshot, for tests and for the seed data the cron
// binary writes.
func SampleInventorySnapshot(id string, at time.Time) InventorySnapshot {
	return InventorySnapshot{
		ID:       id,
		Name:     "name-" + id,
		Label:    "label-" + id,
		Source:   "seeded by inventory.SampleInventorySnapshot",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryBatchSchema is the stored shape of a InventoryBatch.
var InventoryBatchSchema = schema.New("inventory.inventorybatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryBatch is one inventory batch.
type InventoryBatch struct {
	ID       string
	Owner    string
	Name     string
	Channel  string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryBatch) Key() string { return "inventory/inventorybatch/" + x.ID }

// Fields encodes x for storage.
func (x InventoryBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"name":     x.Name,
		"channel":  x.Channel,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryBatch decodes a stored record.
func ParseInventoryBatch(f codec.Fields) (InventoryBatch, error) {
	var x InventoryBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorybatch: missing id")
	}
	x.Owner = f["owner"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorybatch %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorybatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorybatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorybatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryBatch) String() string {
	return fmt.Sprintf("InventoryBatch(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryBatch) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryBatch) Normalize() InventoryBatch {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortInventoryBatchs orders by attempts descending, then by ID.
func SortInventoryBatchs(xs []InventoryBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryBatchs drops the archived entries, preserving order.
func ActiveInventoryBatchs(xs []InventoryBatch) []InventoryBatch {
	out := make([]InventoryBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryBatchs adds up the attempts of every active entry.
func SumInventoryBatchs(xs []InventoryBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexInventoryBatchs keys the entries by ID, last write winning.
func IndexInventoryBatchs(xs []InventoryBatch) map[string]InventoryBatch {
	out := make(map[string]InventoryBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryBatch builds a valid InventoryBatch, for tests and for the seed data the cron
// binary writes.
func SampleInventoryBatch(id string, at time.Time) InventoryBatch {
	return InventoryBatch{
		ID:       id,
		Owner:    "owner-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by inventory.SampleInventoryBatch",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryRequestSchema is the stored shape of a InventoryRequest.
var InventoryRequestSchema = schema.New("inventory.inventoryrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryRequest is one inventory request.
type InventoryRequest struct {
	ID        string
	Region    string
	Reference string
	Owner     string
	Count     int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryRequest) Key() string { return "inventory/inventoryrequest/" + x.ID }

// Fields encodes x for storage.
func (x InventoryRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"owner":     x.Owner,
		"count":     strconv.Itoa(x.Count),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryRequest decodes a stored record.
func ParseInventoryRequest(f codec.Fields) (InventoryRequest, error) {
	var x InventoryRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryrequest: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrequest %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryRequest) String() string {
	return fmt.Sprintf("InventoryRequest(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryRequest) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryRequest) Normalize() InventoryRequest {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortInventoryRequests orders by count descending, then by ID.
func SortInventoryRequests(xs []InventoryRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryRequests drops the archived entries, preserving order.
func ActiveInventoryRequests(xs []InventoryRequest) []InventoryRequest {
	out := make([]InventoryRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryRequests adds up the count of every active entry.
func SumInventoryRequests(xs []InventoryRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexInventoryRequests keys the entries by ID, last write winning.
func IndexInventoryRequests(xs []InventoryRequest) map[string]InventoryRequest {
	out := make(map[string]InventoryRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryRequest builds a valid InventoryRequest, for tests and for the seed data the cron
// binary writes.
func SampleInventoryRequest(id string, at time.Time) InventoryRequest {
	return InventoryRequest{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by inventory.SampleInventoryRequest",
		Count:     len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventorySummarySchema is the stored shape of a InventorySummary.
var InventorySummarySchema = schema.New("inventory.inventorysummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventorySummary is one inventory summary.
type InventorySummary struct {
	ID       string
	Source   string
	Channel  string
	Group    string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventorySummary) Key() string { return "inventory/inventorysummary/" + x.ID }

// Fields encodes x for storage.
func (x InventorySummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventorySummary decodes a stored record.
func ParseInventorySummary(f codec.Fields) (InventorySummary, error) {
	var x InventorySummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorysummary: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorysummary %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorysummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorysummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorysummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventorySummary) String() string {
	return fmt.Sprintf("InventorySummary(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventorySummary) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventorySummary) Normalize() InventorySummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortInventorySummarys orders by attempts descending, then by ID.
func SortInventorySummarys(xs []InventorySummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventorySummarys drops the archived entries, preserving order.
func ActiveInventorySummarys(xs []InventorySummary) []InventorySummary {
	out := make([]InventorySummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventorySummarys adds up the attempts of every active entry.
func SumInventorySummarys(xs []InventorySummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexInventorySummarys keys the entries by ID, last write winning.
func IndexInventorySummarys(xs []InventorySummary) map[string]InventorySummary {
	out := make(map[string]InventorySummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventorySummary builds a valid InventorySummary, for tests and for the seed data the cron
// binary writes.
func SampleInventorySummary(id string, at time.Time) InventorySummary {
	return InventorySummary{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Group:    "seeded by inventory.SampleInventorySummary",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryPolicySchema is the stored shape of a InventoryPolicy.
var InventoryPolicySchema = schema.New("inventory.inventorypolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryPolicy is one inventory policy.
type InventoryPolicy struct {
	ID       string
	Channel  string
	Comment  string
	Owner    string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryPolicy) Key() string { return "inventory/inventorypolicy/" + x.ID }

// Fields encodes x for storage.
func (x InventoryPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryPolicy decodes a stored record.
func ParseInventoryPolicy(f codec.Fields) (InventoryPolicy, error) {
	var x InventoryPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorypolicy: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorypolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorypolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorypolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorypolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryPolicy) String() string {
	return fmt.Sprintf("InventoryPolicy(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryPolicy) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryPolicy) Normalize() InventoryPolicy {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortInventoryPolicys orders by weight descending, then by ID.
func SortInventoryPolicys(xs []InventoryPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryPolicys drops the archived entries, preserving order.
func ActiveInventoryPolicys(xs []InventoryPolicy) []InventoryPolicy {
	out := make([]InventoryPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryPolicys adds up the weight of every active entry.
func SumInventoryPolicys(xs []InventoryPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexInventoryPolicys keys the entries by ID, last write winning.
func IndexInventoryPolicys(xs []InventoryPolicy) map[string]InventoryPolicy {
	out := make(map[string]InventoryPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryPolicy builds a valid InventoryPolicy, for tests and for the seed data the cron
// binary writes.
func SampleInventoryPolicy(id string, at time.Time) InventoryPolicy {
	return InventoryPolicy{
		ID:       id,
		Channel:  "channel-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by inventory.SampleInventoryPolicy",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryLinkSchema is the stored shape of a InventoryLink.
var InventoryLinkSchema = schema.New("inventory.inventorylink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryLink is one inventory link.
type InventoryLink struct {
	ID       string
	Region   string
	Group    string
	Comment  string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryLink) Key() string { return "inventory/inventorylink/" + x.ID }

// Fields encodes x for storage.
func (x InventoryLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"group":    x.Group,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryLink decodes a stored record.
func ParseInventoryLink(f codec.Fields) (InventoryLink, error) {
	var x InventoryLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorylink: missing id")
	}
	x.Region = f["region"]
	x.Group = f["group"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorylink %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorylink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorylink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorylink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryLink) String() string {
	return fmt.Sprintf("InventoryLink(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryLink) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryLink) Normalize() InventoryLink {
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortInventoryLinks orders by version descending, then by ID.
func SortInventoryLinks(xs []InventoryLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryLinks drops the archived entries, preserving order.
func ActiveInventoryLinks(xs []InventoryLink) []InventoryLink {
	out := make([]InventoryLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryLinks adds up the version of every active entry.
func SumInventoryLinks(xs []InventoryLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexInventoryLinks keys the entries by ID, last write winning.
func IndexInventoryLinks(xs []InventoryLink) map[string]InventoryLink {
	out := make(map[string]InventoryLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryLink builds a valid InventoryLink, for tests and for the seed data the cron
// binary writes.
func SampleInventoryLink(id string, at time.Time) InventoryLink {
	return InventoryLink{
		ID:       id,
		Region:   "region-" + id,
		Group:    "group-" + id,
		Comment:  "seeded by inventory.SampleInventoryLink",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryNoteSchema is the stored shape of a InventoryNote.
var InventoryNoteSchema = schema.New("inventory.inventorynote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryNote is one inventory note.
type InventoryNote struct {
	ID       string
	Comment  string
	Region   string
	Source   string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryNote) Key() string { return "inventory/inventorynote/" + x.ID }

// Fields encodes x for storage.
func (x InventoryNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"region":   x.Region,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryNote decodes a stored record.
func ParseInventoryNote(f codec.Fields) (InventoryNote, error) {
	var x InventoryNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorynote: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorynote %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorynote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorynote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorynote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryNote) String() string {
	return fmt.Sprintf("InventoryNote(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryNote) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryNote) Normalize() InventoryNote {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortInventoryNotes orders by version descending, then by ID.
func SortInventoryNotes(xs []InventoryNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryNotes drops the archived entries, preserving order.
func ActiveInventoryNotes(xs []InventoryNote) []InventoryNote {
	out := make([]InventoryNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryNotes adds up the version of every active entry.
func SumInventoryNotes(xs []InventoryNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexInventoryNotes keys the entries by ID, last write winning.
func IndexInventoryNotes(xs []InventoryNote) map[string]InventoryNote {
	out := make(map[string]InventoryNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryNote builds a valid InventoryNote, for tests and for the seed data the cron
// binary writes.
func SampleInventoryNote(id string, at time.Time) InventoryNote {
	return InventoryNote{
		ID:       id,
		Comment:  "comment-" + id,
		Region:   "region-" + id,
		Source:   "seeded by inventory.SampleInventoryNote",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryWindowSchema is the stored shape of a InventoryWindow.
var InventoryWindowSchema = schema.New("inventory.inventorywindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryWindow is one inventory window.
type InventoryWindow struct {
	ID       string
	Category string
	Name     string
	Label    string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryWindow) Key() string { return "inventory/inventorywindow/" + x.ID }

// Fields encodes x for storage.
func (x InventoryWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryWindow decodes a stored record.
func ParseInventoryWindow(f codec.Fields) (InventoryWindow, error) {
	var x InventoryWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorywindow: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorywindow %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorywindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorywindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorywindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryWindow) String() string {
	return fmt.Sprintf("InventoryWindow(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryWindow) Normalize() InventoryWindow {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInventoryWindows orders by priority descending, then by ID.
func SortInventoryWindows(xs []InventoryWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryWindows drops the archived entries, preserving order.
func ActiveInventoryWindows(xs []InventoryWindow) []InventoryWindow {
	out := make([]InventoryWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryWindows adds up the priority of every active entry.
func SumInventoryWindows(xs []InventoryWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInventoryWindows keys the entries by ID, last write winning.
func IndexInventoryWindows(xs []InventoryWindow) map[string]InventoryWindow {
	out := make(map[string]InventoryWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryWindow builds a valid InventoryWindow, for tests and for the seed data the cron
// binary writes.
func SampleInventoryWindow(id string, at time.Time) InventoryWindow {
	return InventoryWindow{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Label:    "seeded by inventory.SampleInventoryWindow",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryQuotaSchema is the stored shape of a InventoryQuota.
var InventoryQuotaSchema = schema.New("inventory.inventoryquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryQuota is one inventory quota.
type InventoryQuota struct {
	ID       string
	Owner    string
	Region   string
	Channel  string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryQuota) Key() string { return "inventory/inventoryquota/" + x.ID }

// Fields encodes x for storage.
func (x InventoryQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryQuota decodes a stored record.
func ParseInventoryQuota(f codec.Fields) (InventoryQuota, error) {
	var x InventoryQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryquota: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryquota %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryQuota) String() string {
	return fmt.Sprintf("InventoryQuota(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryQuota) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryQuota) Normalize() InventoryQuota {
	x.Owner = strings.TrimSpace(x.Owner)
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

// SortInventoryQuotas orders by count descending, then by ID.
func SortInventoryQuotas(xs []InventoryQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryQuotas drops the archived entries, preserving order.
func ActiveInventoryQuotas(xs []InventoryQuota) []InventoryQuota {
	out := make([]InventoryQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryQuotas adds up the count of every active entry.
func SumInventoryQuotas(xs []InventoryQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexInventoryQuotas keys the entries by ID, last write winning.
func IndexInventoryQuotas(xs []InventoryQuota) map[string]InventoryQuota {
	out := make(map[string]InventoryQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryQuota builds a valid InventoryQuota, for tests and for the seed data the cron
// binary writes.
func SampleInventoryQuota(id string, at time.Time) InventoryQuota {
	return InventoryQuota{
		ID:       id,
		Owner:    "owner-" + id,
		Region:   "region-" + id,
		Channel:  "seeded by inventory.SampleInventoryQuota",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryBucketSchema is the stored shape of a InventoryBucket.
var InventoryBucketSchema = schema.New("inventory.inventorybucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryBucket is one inventory bucket.
type InventoryBucket struct {
	ID        string
	Reference string
	Comment   string
	Group     string
	Retained  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryBucket) Key() string { return "inventory/inventorybucket/" + x.ID }

// Fields encodes x for storage.
func (x InventoryBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"comment":   x.Comment,
		"group":     x.Group,
		"retained":  strconv.Itoa(x.Retained),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryBucket decodes a stored record.
func ParseInventoryBucket(f codec.Fields) (InventoryBucket, error) {
	var x InventoryBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorybucket: missing id")
	}
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorybucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorybucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorybucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorybucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryBucket) String() string {
	return fmt.Sprintf("InventoryBucket(%s, reference=%q, retained=%d, archived=%t)", x.ID, x.Reference, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryBucket) Normalize() InventoryBucket {
	x.Reference = strings.TrimSpace(x.Reference)
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

// SortInventoryBuckets orders by retained descending, then by ID.
func SortInventoryBuckets(xs []InventoryBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryBuckets drops the archived entries, preserving order.
func ActiveInventoryBuckets(xs []InventoryBucket) []InventoryBucket {
	out := make([]InventoryBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryBuckets adds up the retained of every active entry.
func SumInventoryBuckets(xs []InventoryBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexInventoryBuckets keys the entries by ID, last write winning.
func IndexInventoryBuckets(xs []InventoryBucket) map[string]InventoryBucket {
	out := make(map[string]InventoryBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryBucket builds a valid InventoryBucket, for tests and for the seed data the cron
// binary writes.
func SampleInventoryBucket(id string, at time.Time) InventoryBucket {
	return InventoryBucket{
		ID:        id,
		Reference: "reference-" + id,
		Comment:   "comment-" + id,
		Group:     "seeded by inventory.SampleInventoryBucket",
		Retained:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventoryLaneSchema is the stored shape of a InventoryLane.
var InventoryLaneSchema = schema.New("inventory.inventorylane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryLane is one inventory lane.
type InventoryLane struct {
	ID       string
	Owner    string
	Category string
	Channel  string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryLane) Key() string { return "inventory/inventorylane/" + x.ID }

// Fields encodes x for storage.
func (x InventoryLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"category": x.Category,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryLane decodes a stored record.
func ParseInventoryLane(f codec.Fields) (InventoryLane, error) {
	var x InventoryLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorylane: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorylane %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorylane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorylane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorylane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryLane) String() string {
	return fmt.Sprintf("InventoryLane(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryLane) Normalize() InventoryLane {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortInventoryLanes orders by capacity descending, then by ID.
func SortInventoryLanes(xs []InventoryLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryLanes drops the archived entries, preserving order.
func ActiveInventoryLanes(xs []InventoryLane) []InventoryLane {
	out := make([]InventoryLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryLanes adds up the capacity of every active entry.
func SumInventoryLanes(xs []InventoryLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexInventoryLanes keys the entries by ID, last write winning.
func IndexInventoryLanes(xs []InventoryLane) map[string]InventoryLane {
	out := make(map[string]InventoryLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryLane builds a valid InventoryLane, for tests and for the seed data the cron
// binary writes.
func SampleInventoryLane(id string, at time.Time) InventoryLane {
	return InventoryLane{
		ID:       id,
		Owner:    "owner-" + id,
		Category: "category-" + id,
		Channel:  "seeded by inventory.SampleInventoryLane",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventorySlotSchema is the stored shape of a InventorySlot.
var InventorySlotSchema = schema.New("inventory.inventoryslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventorySlot is one inventory slot.
type InventorySlot struct {
	ID        string
	Name      string
	Reference string
	Category  string
	Count     int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventorySlot) Key() string { return "inventory/inventoryslot/" + x.ID }

// Fields encodes x for storage.
func (x InventorySlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"category":  x.Category,
		"count":     strconv.Itoa(x.Count),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventorySlot decodes a stored record.
func ParseInventorySlot(f codec.Fields) (InventorySlot, error) {
	var x InventorySlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryslot: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryslot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventorySlot) String() string {
	return fmt.Sprintf("InventorySlot(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventorySlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventorySlot) Normalize() InventorySlot {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortInventorySlots orders by count descending, then by ID.
func SortInventorySlots(xs []InventorySlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventorySlots drops the archived entries, preserving order.
func ActiveInventorySlots(xs []InventorySlot) []InventorySlot {
	out := make([]InventorySlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventorySlots adds up the count of every active entry.
func SumInventorySlots(xs []InventorySlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexInventorySlots keys the entries by ID, last write winning.
func IndexInventorySlots(xs []InventorySlot) map[string]InventorySlot {
	out := make(map[string]InventorySlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventorySlot builds a valid InventorySlot, for tests and for the seed data the cron
// binary writes.
func SampleInventorySlot(id string, at time.Time) InventorySlot {
	return InventorySlot{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by inventory.SampleInventorySlot",
		Count:     len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventoryTagSchema is the stored shape of a InventoryTag.
var InventoryTagSchema = schema.New("inventory.inventorytag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryTag is one inventory tag.
type InventoryTag struct {
	ID       string
	Group    string
	Name     string
	Channel  string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryTag) Key() string { return "inventory/inventorytag/" + x.ID }

// Fields encodes x for storage.
func (x InventoryTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryTag decodes a stored record.
func ParseInventoryTag(f codec.Fields) (InventoryTag, error) {
	var x InventoryTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorytag: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorytag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorytag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorytag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorytag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryTag) String() string {
	return fmt.Sprintf("InventoryTag(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryTag) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryTag) Normalize() InventoryTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortInventoryTags orders by priority descending, then by ID.
func SortInventoryTags(xs []InventoryTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryTags drops the archived entries, preserving order.
func ActiveInventoryTags(xs []InventoryTag) []InventoryTag {
	out := make([]InventoryTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryTags adds up the priority of every active entry.
func SumInventoryTags(xs []InventoryTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInventoryTags keys the entries by ID, last write winning.
func IndexInventoryTags(xs []InventoryTag) map[string]InventoryTag {
	out := make(map[string]InventoryTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryTag builds a valid InventoryTag, for tests and for the seed data the cron
// binary writes.
func SampleInventoryTag(id string, at time.Time) InventoryTag {
	return InventoryTag{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by inventory.SampleInventoryTag",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryRuleSchema is the stored shape of a InventoryRule.
var InventoryRuleSchema = schema.New("inventory.inventoryrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryRule is one inventory rule.
type InventoryRule struct {
	ID       string
	Channel  string
	Group    string
	Name     string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InventoryRule) Key() string { return "inventory/inventoryrule/" + x.ID }

// Fields encodes x for storage.
func (x InventoryRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"name":     x.Name,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryRule decodes a stored record.
func ParseInventoryRule(f codec.Fields) (InventoryRule, error) {
	var x InventoryRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryrule: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Name = f["name"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrule %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryRule) String() string {
	return fmt.Sprintf("InventoryRule(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryRule) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryRule) Normalize() InventoryRule {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInventoryRules orders by priority descending, then by ID.
func SortInventoryRules(xs []InventoryRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryRules drops the archived entries, preserving order.
func ActiveInventoryRules(xs []InventoryRule) []InventoryRule {
	out := make([]InventoryRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryRules adds up the priority of every active entry.
func SumInventoryRules(xs []InventoryRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInventoryRules keys the entries by ID, last write winning.
func IndexInventoryRules(xs []InventoryRule) map[string]InventoryRule {
	out := make(map[string]InventoryRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryRule builds a valid InventoryRule, for tests and for the seed data the cron
// binary writes.
func SampleInventoryRule(id string, at time.Time) InventoryRule {
	return InventoryRule{
		ID:       id,
		Channel:  "channel-" + id,
		Group:    "group-" + id,
		Name:     "seeded by inventory.SampleInventoryRule",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InventoryHoldSchema is the stored shape of a InventoryHold.
var InventoryHoldSchema = schema.New("inventory.inventoryhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryHold is one inventory hold.
type InventoryHold struct {
	ID        string
	Reference string
	Comment   string
	Group     string
	Count     int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryHold) Key() string { return "inventory/inventoryhold/" + x.ID }

// Fields encodes x for storage.
func (x InventoryHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"comment":   x.Comment,
		"group":     x.Group,
		"count":     strconv.Itoa(x.Count),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryHold decodes a stored record.
func ParseInventoryHold(f codec.Fields) (InventoryHold, error) {
	var x InventoryHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryhold: missing id")
	}
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryhold %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryHold) String() string {
	return fmt.Sprintf("InventoryHold(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryHold) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryHold) Normalize() InventoryHold {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortInventoryHolds orders by count descending, then by ID.
func SortInventoryHolds(xs []InventoryHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryHolds drops the archived entries, preserving order.
func ActiveInventoryHolds(xs []InventoryHold) []InventoryHold {
	out := make([]InventoryHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryHolds adds up the count of every active entry.
func SumInventoryHolds(xs []InventoryHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexInventoryHolds keys the entries by ID, last write winning.
func IndexInventoryHolds(xs []InventoryHold) map[string]InventoryHold {
	out := make(map[string]InventoryHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryHold builds a valid InventoryHold, for tests and for the seed data the cron
// binary writes.
func SampleInventoryHold(id string, at time.Time) InventoryHold {
	return InventoryHold{
		ID:        id,
		Reference: "reference-" + id,
		Comment:   "comment-" + id,
		Group:     "seeded by inventory.SampleInventoryHold",
		Count:     len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventoryClaimSchema is the stored shape of a InventoryClaim.
var InventoryClaimSchema = schema.New("inventory.inventoryclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryClaim is one inventory claim.
type InventoryClaim struct {
	ID        string
	Region    string
	Reference string
	Channel   string
	Version   int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryClaim) Key() string { return "inventory/inventoryclaim/" + x.ID }

// Fields encodes x for storage.
func (x InventoryClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"channel":   x.Channel,
		"version":   strconv.Itoa(x.Version),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryClaim decodes a stored record.
func ParseInventoryClaim(f codec.Fields) (InventoryClaim, error) {
	var x InventoryClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventoryclaim: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("inventory: inventoryclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventoryclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventoryclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryClaim) String() string {
	return fmt.Sprintf("InventoryClaim(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryClaim) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryClaim) Normalize() InventoryClaim {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortInventoryClaims orders by version descending, then by ID.
func SortInventoryClaims(xs []InventoryClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryClaims drops the archived entries, preserving order.
func ActiveInventoryClaims(xs []InventoryClaim) []InventoryClaim {
	out := make([]InventoryClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryClaims adds up the version of every active entry.
func SumInventoryClaims(xs []InventoryClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexInventoryClaims keys the entries by ID, last write winning.
func IndexInventoryClaims(xs []InventoryClaim) map[string]InventoryClaim {
	out := make(map[string]InventoryClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryClaim builds a valid InventoryClaim, for tests and for the seed data the cron
// binary writes.
func SampleInventoryClaim(id string, at time.Time) InventoryClaim {
	return InventoryClaim{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by inventory.SampleInventoryClaim",
		Version:   len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InventoryCursorSchema is the stored shape of a InventoryCursor.
var InventoryCursorSchema = schema.New("inventory.inventorycursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InventoryCursor is one inventory cursor.
type InventoryCursor struct {
	ID        string
	Reference string
	Group     string
	Source    string
	Version   int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InventoryCursor) Key() string { return "inventory/inventorycursor/" + x.ID }

// Fields encodes x for storage.
func (x InventoryCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"source":    x.Source,
		"version":   strconv.Itoa(x.Version),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInventoryCursor decodes a stored record.
func ParseInventoryCursor(f codec.Fields) (InventoryCursor, error) {
	var x InventoryCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("inventory: inventorycursor: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorycursor %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("inventory: inventorycursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("inventory: inventorycursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("inventory: inventorycursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InventoryCursor) String() string {
	return fmt.Sprintf("InventoryCursor(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InventoryCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InventoryCursor) Normalize() InventoryCursor {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortInventoryCursors orders by version descending, then by ID.
func SortInventoryCursors(xs []InventoryCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInventoryCursors drops the archived entries, preserving order.
func ActiveInventoryCursors(xs []InventoryCursor) []InventoryCursor {
	out := make([]InventoryCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInventoryCursors adds up the version of every active entry.
func SumInventoryCursors(xs []InventoryCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexInventoryCursors keys the entries by ID, last write winning.
func IndexInventoryCursors(xs []InventoryCursor) map[string]InventoryCursor {
	out := make(map[string]InventoryCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInventoryCursor builds a valid InventoryCursor, for tests and for the seed data the cron
// binary writes.
func SampleInventoryCursor(id string, at time.Time) InventoryCursor {
	return InventoryCursor{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Source:    "seeded by inventory.SampleInventoryCursor",
		Version:   len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	InventoryRecordSchema,
	InventoryEntrySchema,
	InventoryItemSchema,
	InventoryDraftSchema,
	InventoryRevisionSchema,
	InventorySnapshotSchema,
	InventoryBatchSchema,
	InventoryRequestSchema,
	InventorySummarySchema,
	InventoryPolicySchema,
	InventoryLinkSchema,
	InventoryNoteSchema,
	InventoryWindowSchema,
	InventoryQuotaSchema,
	InventoryBucketSchema,
	InventoryLaneSchema,
	InventorySlotSchema,
	InventoryTagSchema,
	InventoryRuleSchema,
	InventoryHoldSchema,
	InventoryClaimSchema,
	InventoryCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"inventory/inventoryrecord/",
	"inventory/inventoryentry/",
	"inventory/inventoryitem/",
	"inventory/inventorydraft/",
	"inventory/inventoryrevision/",
	"inventory/inventorysnapshot/",
	"inventory/inventorybatch/",
	"inventory/inventoryrequest/",
	"inventory/inventorysummary/",
	"inventory/inventorypolicy/",
	"inventory/inventorylink/",
	"inventory/inventorynote/",
	"inventory/inventorywindow/",
	"inventory/inventoryquota/",
	"inventory/inventorybucket/",
	"inventory/inventorylane/",
	"inventory/inventoryslot/",
	"inventory/inventorytag/",
	"inventory/inventoryrule/",
	"inventory/inventoryhold/",
	"inventory/inventoryclaim/",
	"inventory/inventorycursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "inventory: " + strings.Join(names, ", ")
}
