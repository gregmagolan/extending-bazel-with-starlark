// Package warehouse holds the warehouse domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package warehouse

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// WarehouseRecordSchema is the stored shape of a WarehouseRecord.
var WarehouseRecordSchema = schema.New("warehouse.warehouserecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseRecord is one warehouse record.
type WarehouseRecord struct {
	ID       string
	Category string
	Group    string
	Channel  string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseRecord) Key() string { return "warehouse/warehouserecord/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"group":    x.Group,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseRecord decodes a stored record.
func ParseWarehouseRecord(f codec.Fields) (WarehouseRecord, error) {
	var x WarehouseRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouserecord: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserecord %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouserecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouserecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseRecord) String() string {
	return fmt.Sprintf("WarehouseRecord(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseRecord) Normalize() WarehouseRecord {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortWarehouseRecords orders by capacity descending, then by ID.
func SortWarehouseRecords(xs []WarehouseRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseRecords drops the archived entries, preserving order.
func ActiveWarehouseRecords(xs []WarehouseRecord) []WarehouseRecord {
	out := make([]WarehouseRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseRecords adds up the capacity of every active entry.
func SumWarehouseRecords(xs []WarehouseRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexWarehouseRecords keys the entries by ID, last write winning.
func IndexWarehouseRecords(xs []WarehouseRecord) map[string]WarehouseRecord {
	out := make(map[string]WarehouseRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseRecord builds a valid WarehouseRecord, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseRecord(id string, at time.Time) WarehouseRecord {
	return WarehouseRecord{
		ID:       id,
		Category: "category-" + id,
		Group:    "group-" + id,
		Channel:  "seeded by warehouse.SampleWarehouseRecord",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseEntrySchema is the stored shape of a WarehouseEntry.
var WarehouseEntrySchema = schema.New("warehouse.warehouseentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseEntry is one warehouse entry.
type WarehouseEntry struct {
	ID       string
	Comment  string
	Owner    string
	Source   string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseEntry) Key() string { return "warehouse/warehouseentry/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseEntry decodes a stored record.
func ParseWarehouseEntry(f codec.Fields) (WarehouseEntry, error) {
	var x WarehouseEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouseentry: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouseentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouseentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseEntry) String() string {
	return fmt.Sprintf("WarehouseEntry(%s, comment=%q, weight=%d, archived=%t)", x.ID, x.Comment, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseEntry) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseEntry) Normalize() WarehouseEntry {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWarehouseEntrys orders by weight descending, then by ID.
func SortWarehouseEntrys(xs []WarehouseEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseEntrys drops the archived entries, preserving order.
func ActiveWarehouseEntrys(xs []WarehouseEntry) []WarehouseEntry {
	out := make([]WarehouseEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseEntrys adds up the weight of every active entry.
func SumWarehouseEntrys(xs []WarehouseEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexWarehouseEntrys keys the entries by ID, last write winning.
func IndexWarehouseEntrys(xs []WarehouseEntry) map[string]WarehouseEntry {
	out := make(map[string]WarehouseEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseEntry builds a valid WarehouseEntry, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseEntry(id string, at time.Time) WarehouseEntry {
	return WarehouseEntry{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by warehouse.SampleWarehouseEntry",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseItemSchema is the stored shape of a WarehouseItem.
var WarehouseItemSchema = schema.New("warehouse.warehouseitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseItem is one warehouse item.
type WarehouseItem struct {
	ID        string
	Owner     string
	Reference string
	Category  string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseItem) Key() string { return "warehouse/warehouseitem/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"reference": x.Reference,
		"category":  x.Category,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseItem decodes a stored record.
func ParseWarehouseItem(f codec.Fields) (WarehouseItem, error) {
	var x WarehouseItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouseitem: missing id")
	}
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseitem %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouseitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouseitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseItem) String() string {
	return fmt.Sprintf("WarehouseItem(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseItem) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseItem) Normalize() WarehouseItem {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortWarehouseItems orders by offset descending, then by ID.
func SortWarehouseItems(xs []WarehouseItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseItems drops the archived entries, preserving order.
func ActiveWarehouseItems(xs []WarehouseItem) []WarehouseItem {
	out := make([]WarehouseItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseItems adds up the offset of every active entry.
func SumWarehouseItems(xs []WarehouseItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexWarehouseItems keys the entries by ID, last write winning.
func IndexWarehouseItems(xs []WarehouseItem) map[string]WarehouseItem {
	out := make(map[string]WarehouseItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseItem builds a valid WarehouseItem, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseItem(id string, at time.Time) WarehouseItem {
	return WarehouseItem{
		ID:        id,
		Owner:     "owner-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by warehouse.SampleWarehouseItem",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseDraftSchema is the stored shape of a WarehouseDraft.
var WarehouseDraftSchema = schema.New("warehouse.warehousedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseDraft is one warehouse draft.
type WarehouseDraft struct {
	ID        string
	Category  string
	Name      string
	Reference string
	Priority  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseDraft) Key() string { return "warehouse/warehousedraft/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"name":      x.Name,
		"reference": x.Reference,
		"priority":  strconv.Itoa(x.Priority),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseDraft decodes a stored record.
func ParseWarehouseDraft(f codec.Fields) (WarehouseDraft, error) {
	var x WarehouseDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousedraft: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousedraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseDraft) String() string {
	return fmt.Sprintf("WarehouseDraft(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseDraft) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseDraft) Normalize() WarehouseDraft {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortWarehouseDrafts orders by priority descending, then by ID.
func SortWarehouseDrafts(xs []WarehouseDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseDrafts drops the archived entries, preserving order.
func ActiveWarehouseDrafts(xs []WarehouseDraft) []WarehouseDraft {
	out := make([]WarehouseDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseDrafts adds up the priority of every active entry.
func SumWarehouseDrafts(xs []WarehouseDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWarehouseDrafts keys the entries by ID, last write winning.
func IndexWarehouseDrafts(xs []WarehouseDraft) map[string]WarehouseDraft {
	out := make(map[string]WarehouseDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseDraft builds a valid WarehouseDraft, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseDraft(id string, at time.Time) WarehouseDraft {
	return WarehouseDraft{
		ID:        id,
		Category:  "category-" + id,
		Name:      "name-" + id,
		Reference: "seeded by warehouse.SampleWarehouseDraft",
		Priority:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseRevisionSchema is the stored shape of a WarehouseRevision.
var WarehouseRevisionSchema = schema.New("warehouse.warehouserevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseRevision is one warehouse revision.
type WarehouseRevision struct {
	ID       string
	Source   string
	Owner    string
	Region   string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseRevision) Key() string { return "warehouse/warehouserevision/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseRevision decodes a stored record.
func ParseWarehouseRevision(f codec.Fields) (WarehouseRevision, error) {
	var x WarehouseRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouserevision: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserevision %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouserevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouserevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseRevision) String() string {
	return fmt.Sprintf("WarehouseRevision(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseRevision) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseRevision) Normalize() WarehouseRevision {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWarehouseRevisions orders by weight descending, then by ID.
func SortWarehouseRevisions(xs []WarehouseRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseRevisions drops the archived entries, preserving order.
func ActiveWarehouseRevisions(xs []WarehouseRevision) []WarehouseRevision {
	out := make([]WarehouseRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseRevisions adds up the weight of every active entry.
func SumWarehouseRevisions(xs []WarehouseRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexWarehouseRevisions keys the entries by ID, last write winning.
func IndexWarehouseRevisions(xs []WarehouseRevision) map[string]WarehouseRevision {
	out := make(map[string]WarehouseRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseRevision builds a valid WarehouseRevision, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseRevision(id string, at time.Time) WarehouseRevision {
	return WarehouseRevision{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by warehouse.SampleWarehouseRevision",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseSnapshotSchema is the stored shape of a WarehouseSnapshot.
var WarehouseSnapshotSchema = schema.New("warehouse.warehousesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseSnapshot is one warehouse snapshot.
type WarehouseSnapshot struct {
	ID       string
	Owner    string
	Category string
	Comment  string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseSnapshot) Key() string { return "warehouse/warehousesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"category": x.Category,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseSnapshot decodes a stored record.
func ParseWarehouseSnapshot(f codec.Fields) (WarehouseSnapshot, error) {
	var x WarehouseSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousesnapshot: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousesnapshot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseSnapshot) String() string {
	return fmt.Sprintf("WarehouseSnapshot(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseSnapshot) Normalize() WarehouseSnapshot {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortWarehouseSnapshots orders by priority descending, then by ID.
func SortWarehouseSnapshots(xs []WarehouseSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseSnapshots drops the archived entries, preserving order.
func ActiveWarehouseSnapshots(xs []WarehouseSnapshot) []WarehouseSnapshot {
	out := make([]WarehouseSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseSnapshots adds up the priority of every active entry.
func SumWarehouseSnapshots(xs []WarehouseSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWarehouseSnapshots keys the entries by ID, last write winning.
func IndexWarehouseSnapshots(xs []WarehouseSnapshot) map[string]WarehouseSnapshot {
	out := make(map[string]WarehouseSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseSnapshot builds a valid WarehouseSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseSnapshot(id string, at time.Time) WarehouseSnapshot {
	return WarehouseSnapshot{
		ID:       id,
		Owner:    "owner-" + id,
		Category: "category-" + id,
		Comment:  "seeded by warehouse.SampleWarehouseSnapshot",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseBatchSchema is the stored shape of a WarehouseBatch.
var WarehouseBatchSchema = schema.New("warehouse.warehousebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseBatch is one warehouse batch.
type WarehouseBatch struct {
	ID       string
	Region   string
	Category string
	Name     string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseBatch) Key() string { return "warehouse/warehousebatch/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseBatch decodes a stored record.
func ParseWarehouseBatch(f codec.Fields) (WarehouseBatch, error) {
	var x WarehouseBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousebatch: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousebatch %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseBatch) String() string {
	return fmt.Sprintf("WarehouseBatch(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseBatch) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseBatch) Normalize() WarehouseBatch {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortWarehouseBatchs orders by count descending, then by ID.
func SortWarehouseBatchs(xs []WarehouseBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseBatchs drops the archived entries, preserving order.
func ActiveWarehouseBatchs(xs []WarehouseBatch) []WarehouseBatch {
	out := make([]WarehouseBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseBatchs adds up the count of every active entry.
func SumWarehouseBatchs(xs []WarehouseBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexWarehouseBatchs keys the entries by ID, last write winning.
func IndexWarehouseBatchs(xs []WarehouseBatch) map[string]WarehouseBatch {
	out := make(map[string]WarehouseBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseBatch builds a valid WarehouseBatch, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseBatch(id string, at time.Time) WarehouseBatch {
	return WarehouseBatch{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Name:     "seeded by warehouse.SampleWarehouseBatch",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseRequestSchema is the stored shape of a WarehouseRequest.
var WarehouseRequestSchema = schema.New("warehouse.warehouserequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseRequest is one warehouse request.
type WarehouseRequest struct {
	ID        string
	Owner     string
	Region    string
	Reference string
	Attempts  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseRequest) Key() string { return "warehouse/warehouserequest/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"region":    x.Region,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseRequest decodes a stored record.
func ParseWarehouseRequest(f codec.Fields) (WarehouseRequest, error) {
	var x WarehouseRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouserequest: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserequest %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouserequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouserequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseRequest) String() string {
	return fmt.Sprintf("WarehouseRequest(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseRequest) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseRequest) Normalize() WarehouseRequest {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWarehouseRequests orders by attempts descending, then by ID.
func SortWarehouseRequests(xs []WarehouseRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseRequests drops the archived entries, preserving order.
func ActiveWarehouseRequests(xs []WarehouseRequest) []WarehouseRequest {
	out := make([]WarehouseRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseRequests adds up the attempts of every active entry.
func SumWarehouseRequests(xs []WarehouseRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexWarehouseRequests keys the entries by ID, last write winning.
func IndexWarehouseRequests(xs []WarehouseRequest) map[string]WarehouseRequest {
	out := make(map[string]WarehouseRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseRequest builds a valid WarehouseRequest, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseRequest(id string, at time.Time) WarehouseRequest {
	return WarehouseRequest{
		ID:        id,
		Owner:     "owner-" + id,
		Region:    "region-" + id,
		Reference: "seeded by warehouse.SampleWarehouseRequest",
		Attempts:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseSummarySchema is the stored shape of a WarehouseSummary.
var WarehouseSummarySchema = schema.New("warehouse.warehousesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseSummary is one warehouse summary.
type WarehouseSummary struct {
	ID       string
	Source   string
	Name     string
	Label    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseSummary) Key() string { return "warehouse/warehousesummary/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseSummary decodes a stored record.
func ParseWarehouseSummary(f codec.Fields) (WarehouseSummary, error) {
	var x WarehouseSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousesummary: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousesummary %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseSummary) String() string {
	return fmt.Sprintf("WarehouseSummary(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseSummary) Normalize() WarehouseSummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortWarehouseSummarys orders by capacity descending, then by ID.
func SortWarehouseSummarys(xs []WarehouseSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseSummarys drops the archived entries, preserving order.
func ActiveWarehouseSummarys(xs []WarehouseSummary) []WarehouseSummary {
	out := make([]WarehouseSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseSummarys adds up the capacity of every active entry.
func SumWarehouseSummarys(xs []WarehouseSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexWarehouseSummarys keys the entries by ID, last write winning.
func IndexWarehouseSummarys(xs []WarehouseSummary) map[string]WarehouseSummary {
	out := make(map[string]WarehouseSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseSummary builds a valid WarehouseSummary, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseSummary(id string, at time.Time) WarehouseSummary {
	return WarehouseSummary{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Label:    "seeded by warehouse.SampleWarehouseSummary",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehousePolicySchema is the stored shape of a WarehousePolicy.
var WarehousePolicySchema = schema.New("warehouse.warehousepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehousePolicy is one warehouse policy.
type WarehousePolicy struct {
	ID       string
	Comment  string
	Channel  string
	Category string
	Version  int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehousePolicy) Key() string { return "warehouse/warehousepolicy/" + x.ID }

// Fields encodes x for storage.
func (x WarehousePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehousePolicy decodes a stored record.
func ParseWarehousePolicy(f codec.Fields) (WarehousePolicy, error) {
	var x WarehousePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousepolicy: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousepolicy %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehousePolicy) String() string {
	return fmt.Sprintf("WarehousePolicy(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehousePolicy) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehousePolicy) Normalize() WarehousePolicy {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortWarehousePolicys orders by version descending, then by ID.
func SortWarehousePolicys(xs []WarehousePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehousePolicys drops the archived entries, preserving order.
func ActiveWarehousePolicys(xs []WarehousePolicy) []WarehousePolicy {
	out := make([]WarehousePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehousePolicys adds up the version of every active entry.
func SumWarehousePolicys(xs []WarehousePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexWarehousePolicys keys the entries by ID, last write winning.
func IndexWarehousePolicys(xs []WarehousePolicy) map[string]WarehousePolicy {
	out := make(map[string]WarehousePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehousePolicy builds a valid WarehousePolicy, for tests and for the seed data the cron
// binary writes.
func SampleWarehousePolicy(id string, at time.Time) WarehousePolicy {
	return WarehousePolicy{
		ID:       id,
		Comment:  "comment-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by warehouse.SampleWarehousePolicy",
		Version:  len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseLinkSchema is the stored shape of a WarehouseLink.
var WarehouseLinkSchema = schema.New("warehouse.warehouselink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseLink is one warehouse link.
type WarehouseLink struct {
	ID       string
	Owner    string
	Label    string
	Comment  string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseLink) Key() string { return "warehouse/warehouselink/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"label":    x.Label,
		"comment":  x.Comment,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseLink decodes a stored record.
func ParseWarehouseLink(f codec.Fields) (WarehouseLink, error) {
	var x WarehouseLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouselink: missing id")
	}
	x.Owner = f["owner"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouselink %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouselink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouselink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouselink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseLink) String() string {
	return fmt.Sprintf("WarehouseLink(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseLink) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseLink) Normalize() WarehouseLink {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortWarehouseLinks orders by attempts descending, then by ID.
func SortWarehouseLinks(xs []WarehouseLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseLinks drops the archived entries, preserving order.
func ActiveWarehouseLinks(xs []WarehouseLink) []WarehouseLink {
	out := make([]WarehouseLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseLinks adds up the attempts of every active entry.
func SumWarehouseLinks(xs []WarehouseLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexWarehouseLinks keys the entries by ID, last write winning.
func IndexWarehouseLinks(xs []WarehouseLink) map[string]WarehouseLink {
	out := make(map[string]WarehouseLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseLink builds a valid WarehouseLink, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseLink(id string, at time.Time) WarehouseLink {
	return WarehouseLink{
		ID:       id,
		Owner:    "owner-" + id,
		Label:    "label-" + id,
		Comment:  "seeded by warehouse.SampleWarehouseLink",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseNoteSchema is the stored shape of a WarehouseNote.
var WarehouseNoteSchema = schema.New("warehouse.warehousenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseNote is one warehouse note.
type WarehouseNote struct {
	ID       string
	Label    string
	Source   string
	Owner    string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseNote) Key() string { return "warehouse/warehousenote/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"owner":    x.Owner,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseNote decodes a stored record.
func ParseWarehouseNote(f codec.Fields) (WarehouseNote, error) {
	var x WarehouseNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousenote: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousenote %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseNote) String() string {
	return fmt.Sprintf("WarehouseNote(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseNote) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseNote) Normalize() WarehouseNote {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWarehouseNotes orders by weight descending, then by ID.
func SortWarehouseNotes(xs []WarehouseNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseNotes drops the archived entries, preserving order.
func ActiveWarehouseNotes(xs []WarehouseNote) []WarehouseNote {
	out := make([]WarehouseNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseNotes adds up the weight of every active entry.
func SumWarehouseNotes(xs []WarehouseNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexWarehouseNotes keys the entries by ID, last write winning.
func IndexWarehouseNotes(xs []WarehouseNote) map[string]WarehouseNote {
	out := make(map[string]WarehouseNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseNote builds a valid WarehouseNote, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseNote(id string, at time.Time) WarehouseNote {
	return WarehouseNote{
		ID:       id,
		Label:    "label-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by warehouse.SampleWarehouseNote",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseWindowSchema is the stored shape of a WarehouseWindow.
var WarehouseWindowSchema = schema.New("warehouse.warehousewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseWindow is one warehouse window.
type WarehouseWindow struct {
	ID        string
	Reference string
	Group     string
	Label     string
	Attempts  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseWindow) Key() string { return "warehouse/warehousewindow/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"label":     x.Label,
		"attempts":  strconv.Itoa(x.Attempts),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseWindow decodes a stored record.
func ParseWarehouseWindow(f codec.Fields) (WarehouseWindow, error) {
	var x WarehouseWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousewindow: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousewindow %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseWindow) String() string {
	return fmt.Sprintf("WarehouseWindow(%s, reference=%q, attempts=%d, archived=%t)", x.ID, x.Reference, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseWindow) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseWindow) Normalize() WarehouseWindow {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortWarehouseWindows orders by attempts descending, then by ID.
func SortWarehouseWindows(xs []WarehouseWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseWindows drops the archived entries, preserving order.
func ActiveWarehouseWindows(xs []WarehouseWindow) []WarehouseWindow {
	out := make([]WarehouseWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseWindows adds up the attempts of every active entry.
func SumWarehouseWindows(xs []WarehouseWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexWarehouseWindows keys the entries by ID, last write winning.
func IndexWarehouseWindows(xs []WarehouseWindow) map[string]WarehouseWindow {
	out := make(map[string]WarehouseWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseWindow builds a valid WarehouseWindow, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseWindow(id string, at time.Time) WarehouseWindow {
	return WarehouseWindow{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Label:     "seeded by warehouse.SampleWarehouseWindow",
		Attempts:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseQuotaSchema is the stored shape of a WarehouseQuota.
var WarehouseQuotaSchema = schema.New("warehouse.warehousequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseQuota is one warehouse quota.
type WarehouseQuota struct {
	ID       string
	Group    string
	Name     string
	Channel  string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseQuota) Key() string { return "warehouse/warehousequota/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseQuota decodes a stored record.
func ParseWarehouseQuota(f codec.Fields) (WarehouseQuota, error) {
	var x WarehouseQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousequota: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousequota %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseQuota) String() string {
	return fmt.Sprintf("WarehouseQuota(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseQuota) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseQuota) Normalize() WarehouseQuota {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortWarehouseQuotas orders by count descending, then by ID.
func SortWarehouseQuotas(xs []WarehouseQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseQuotas drops the archived entries, preserving order.
func ActiveWarehouseQuotas(xs []WarehouseQuota) []WarehouseQuota {
	out := make([]WarehouseQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseQuotas adds up the count of every active entry.
func SumWarehouseQuotas(xs []WarehouseQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexWarehouseQuotas keys the entries by ID, last write winning.
func IndexWarehouseQuotas(xs []WarehouseQuota) map[string]WarehouseQuota {
	out := make(map[string]WarehouseQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseQuota builds a valid WarehouseQuota, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseQuota(id string, at time.Time) WarehouseQuota {
	return WarehouseQuota{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by warehouse.SampleWarehouseQuota",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseBucketSchema is the stored shape of a WarehouseBucket.
var WarehouseBucketSchema = schema.New("warehouse.warehousebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseBucket is one warehouse bucket.
type WarehouseBucket struct {
	ID       string
	Group    string
	Label    string
	Name     string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseBucket) Key() string { return "warehouse/warehousebucket/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"label":    x.Label,
		"name":     x.Name,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseBucket decodes a stored record.
func ParseWarehouseBucket(f codec.Fields) (WarehouseBucket, error) {
	var x WarehouseBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousebucket: missing id")
	}
	x.Group = f["group"]
	x.Label = f["label"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousebucket %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseBucket) String() string {
	return fmt.Sprintf("WarehouseBucket(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseBucket) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseBucket) Normalize() WarehouseBucket {
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortWarehouseBuckets orders by capacity descending, then by ID.
func SortWarehouseBuckets(xs []WarehouseBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseBuckets drops the archived entries, preserving order.
func ActiveWarehouseBuckets(xs []WarehouseBucket) []WarehouseBucket {
	out := make([]WarehouseBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseBuckets adds up the capacity of every active entry.
func SumWarehouseBuckets(xs []WarehouseBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexWarehouseBuckets keys the entries by ID, last write winning.
func IndexWarehouseBuckets(xs []WarehouseBucket) map[string]WarehouseBucket {
	out := make(map[string]WarehouseBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseBucket builds a valid WarehouseBucket, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseBucket(id string, at time.Time) WarehouseBucket {
	return WarehouseBucket{
		ID:       id,
		Group:    "group-" + id,
		Label:    "label-" + id,
		Name:     "seeded by warehouse.SampleWarehouseBucket",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseLaneSchema is the stored shape of a WarehouseLane.
var WarehouseLaneSchema = schema.New("warehouse.warehouselane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseLane is one warehouse lane.
type WarehouseLane struct {
	ID        string
	Channel   string
	Name      string
	Reference string
	Attempts  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseLane) Key() string { return "warehouse/warehouselane/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"name":      x.Name,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseLane decodes a stored record.
func ParseWarehouseLane(f codec.Fields) (WarehouseLane, error) {
	var x WarehouseLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouselane: missing id")
	}
	x.Channel = f["channel"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouselane %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouselane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouselane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouselane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseLane) String() string {
	return fmt.Sprintf("WarehouseLane(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseLane) Normalize() WarehouseLane {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWarehouseLanes orders by attempts descending, then by ID.
func SortWarehouseLanes(xs []WarehouseLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseLanes drops the archived entries, preserving order.
func ActiveWarehouseLanes(xs []WarehouseLane) []WarehouseLane {
	out := make([]WarehouseLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseLanes adds up the attempts of every active entry.
func SumWarehouseLanes(xs []WarehouseLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexWarehouseLanes keys the entries by ID, last write winning.
func IndexWarehouseLanes(xs []WarehouseLane) map[string]WarehouseLane {
	out := make(map[string]WarehouseLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseLane builds a valid WarehouseLane, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseLane(id string, at time.Time) WarehouseLane {
	return WarehouseLane{
		ID:        id,
		Channel:   "channel-" + id,
		Name:      "name-" + id,
		Reference: "seeded by warehouse.SampleWarehouseLane",
		Attempts:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseSlotSchema is the stored shape of a WarehouseSlot.
var WarehouseSlotSchema = schema.New("warehouse.warehouseslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseSlot is one warehouse slot.
type WarehouseSlot struct {
	ID        string
	Channel   string
	Comment   string
	Reference string
	Attempts  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseSlot) Key() string { return "warehouse/warehouseslot/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"comment":   x.Comment,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseSlot decodes a stored record.
func ParseWarehouseSlot(f codec.Fields) (WarehouseSlot, error) {
	var x WarehouseSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouseslot: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseslot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouseslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouseslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseSlot) String() string {
	return fmt.Sprintf("WarehouseSlot(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseSlot) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseSlot) Normalize() WarehouseSlot {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortWarehouseSlots orders by attempts descending, then by ID.
func SortWarehouseSlots(xs []WarehouseSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseSlots drops the archived entries, preserving order.
func ActiveWarehouseSlots(xs []WarehouseSlot) []WarehouseSlot {
	out := make([]WarehouseSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseSlots adds up the attempts of every active entry.
func SumWarehouseSlots(xs []WarehouseSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexWarehouseSlots keys the entries by ID, last write winning.
func IndexWarehouseSlots(xs []WarehouseSlot) map[string]WarehouseSlot {
	out := make(map[string]WarehouseSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseSlot builds a valid WarehouseSlot, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseSlot(id string, at time.Time) WarehouseSlot {
	return WarehouseSlot{
		ID:        id,
		Channel:   "channel-" + id,
		Comment:   "comment-" + id,
		Reference: "seeded by warehouse.SampleWarehouseSlot",
		Attempts:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseTagSchema is the stored shape of a WarehouseTag.
var WarehouseTagSchema = schema.New("warehouse.warehousetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseTag is one warehouse tag.
type WarehouseTag struct {
	ID       string
	Comment  string
	Category string
	Name     string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseTag) Key() string { return "warehouse/warehousetag/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"name":     x.Name,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseTag decodes a stored record.
func ParseWarehouseTag(f codec.Fields) (WarehouseTag, error) {
	var x WarehouseTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousetag: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousetag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseTag) String() string {
	return fmt.Sprintf("WarehouseTag(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseTag) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseTag) Normalize() WarehouseTag {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortWarehouseTags orders by version descending, then by ID.
func SortWarehouseTags(xs []WarehouseTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseTags drops the archived entries, preserving order.
func ActiveWarehouseTags(xs []WarehouseTag) []WarehouseTag {
	out := make([]WarehouseTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseTags adds up the version of every active entry.
func SumWarehouseTags(xs []WarehouseTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexWarehouseTags keys the entries by ID, last write winning.
func IndexWarehouseTags(xs []WarehouseTag) map[string]WarehouseTag {
	out := make(map[string]WarehouseTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseTag builds a valid WarehouseTag, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseTag(id string, at time.Time) WarehouseTag {
	return WarehouseTag{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Name:     "seeded by warehouse.SampleWarehouseTag",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseRuleSchema is the stored shape of a WarehouseRule.
var WarehouseRuleSchema = schema.New("warehouse.warehouserule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseRule is one warehouse rule.
type WarehouseRule struct {
	ID        string
	Owner     string
	Channel   string
	Reference string
	Priority  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseRule) Key() string { return "warehouse/warehouserule/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"channel":   x.Channel,
		"reference": x.Reference,
		"priority":  strconv.Itoa(x.Priority),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseRule decodes a stored record.
func ParseWarehouseRule(f codec.Fields) (WarehouseRule, error) {
	var x WarehouseRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouserule: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserule %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouserule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouserule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouserule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseRule) String() string {
	return fmt.Sprintf("WarehouseRule(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseRule) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseRule) Normalize() WarehouseRule {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortWarehouseRules orders by priority descending, then by ID.
func SortWarehouseRules(xs []WarehouseRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseRules drops the archived entries, preserving order.
func ActiveWarehouseRules(xs []WarehouseRule) []WarehouseRule {
	out := make([]WarehouseRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseRules adds up the priority of every active entry.
func SumWarehouseRules(xs []WarehouseRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWarehouseRules keys the entries by ID, last write winning.
func IndexWarehouseRules(xs []WarehouseRule) map[string]WarehouseRule {
	out := make(map[string]WarehouseRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseRule builds a valid WarehouseRule, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseRule(id string, at time.Time) WarehouseRule {
	return WarehouseRule{
		ID:        id,
		Owner:     "owner-" + id,
		Channel:   "channel-" + id,
		Reference: "seeded by warehouse.SampleWarehouseRule",
		Priority:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseHoldSchema is the stored shape of a WarehouseHold.
var WarehouseHoldSchema = schema.New("warehouse.warehousehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseHold is one warehouse hold.
type WarehouseHold struct {
	ID       string
	Name     string
	Category string
	Owner    string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseHold) Key() string { return "warehouse/warehousehold/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseHold decodes a stored record.
func ParseWarehouseHold(f codec.Fields) (WarehouseHold, error) {
	var x WarehouseHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousehold: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousehold %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseHold) String() string {
	return fmt.Sprintf("WarehouseHold(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseHold) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseHold) Normalize() WarehouseHold {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortWarehouseHolds orders by priority descending, then by ID.
func SortWarehouseHolds(xs []WarehouseHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseHolds drops the archived entries, preserving order.
func ActiveWarehouseHolds(xs []WarehouseHold) []WarehouseHold {
	out := make([]WarehouseHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseHolds adds up the priority of every active entry.
func SumWarehouseHolds(xs []WarehouseHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWarehouseHolds keys the entries by ID, last write winning.
func IndexWarehouseHolds(xs []WarehouseHold) map[string]WarehouseHold {
	out := make(map[string]WarehouseHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseHold builds a valid WarehouseHold, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseHold(id string, at time.Time) WarehouseHold {
	return WarehouseHold{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Owner:    "seeded by warehouse.SampleWarehouseHold",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WarehouseClaimSchema is the stored shape of a WarehouseClaim.
var WarehouseClaimSchema = schema.New("warehouse.warehouseclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseClaim is one warehouse claim.
type WarehouseClaim struct {
	ID        string
	Category  string
	Reference string
	Name      string
	Offset    int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WarehouseClaim) Key() string { return "warehouse/warehouseclaim/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"name":      x.Name,
		"offset":    strconv.Itoa(x.Offset),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseClaim decodes a stored record.
func ParseWarehouseClaim(f codec.Fields) (WarehouseClaim, error) {
	var x WarehouseClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehouseclaim: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehouseclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehouseclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehouseclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseClaim) String() string {
	return fmt.Sprintf("WarehouseClaim(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseClaim) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseClaim) Normalize() WarehouseClaim {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortWarehouseClaims orders by offset descending, then by ID.
func SortWarehouseClaims(xs []WarehouseClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseClaims drops the archived entries, preserving order.
func ActiveWarehouseClaims(xs []WarehouseClaim) []WarehouseClaim {
	out := make([]WarehouseClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseClaims adds up the offset of every active entry.
func SumWarehouseClaims(xs []WarehouseClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexWarehouseClaims keys the entries by ID, last write winning.
func IndexWarehouseClaims(xs []WarehouseClaim) map[string]WarehouseClaim {
	out := make(map[string]WarehouseClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseClaim builds a valid WarehouseClaim, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseClaim(id string, at time.Time) WarehouseClaim {
	return WarehouseClaim{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by warehouse.SampleWarehouseClaim",
		Offset:    len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WarehouseCursorSchema is the stored shape of a WarehouseCursor.
var WarehouseCursorSchema = schema.New("warehouse.warehousecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WarehouseCursor is one warehouse cursor.
type WarehouseCursor struct {
	ID       string
	Region   string
	Channel  string
	Owner    string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WarehouseCursor) Key() string { return "warehouse/warehousecursor/" + x.ID }

// Fields encodes x for storage.
func (x WarehouseCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWarehouseCursor decodes a stored record.
func ParseWarehouseCursor(f codec.Fields) (WarehouseCursor, error) {
	var x WarehouseCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("warehouse: warehousecursor: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousecursor %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("warehouse: warehousecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("warehouse: warehousecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("warehouse: warehousecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WarehouseCursor) String() string {
	return fmt.Sprintf("WarehouseCursor(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WarehouseCursor) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WarehouseCursor) Normalize() WarehouseCursor {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortWarehouseCursors orders by priority descending, then by ID.
func SortWarehouseCursors(xs []WarehouseCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWarehouseCursors drops the archived entries, preserving order.
func ActiveWarehouseCursors(xs []WarehouseCursor) []WarehouseCursor {
	out := make([]WarehouseCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWarehouseCursors adds up the priority of every active entry.
func SumWarehouseCursors(xs []WarehouseCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWarehouseCursors keys the entries by ID, last write winning.
func IndexWarehouseCursors(xs []WarehouseCursor) map[string]WarehouseCursor {
	out := make(map[string]WarehouseCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWarehouseCursor builds a valid WarehouseCursor, for tests and for the seed data the cron
// binary writes.
func SampleWarehouseCursor(id string, at time.Time) WarehouseCursor {
	return WarehouseCursor{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by warehouse.SampleWarehouseCursor",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	WarehouseRecordSchema,
	WarehouseEntrySchema,
	WarehouseItemSchema,
	WarehouseDraftSchema,
	WarehouseRevisionSchema,
	WarehouseSnapshotSchema,
	WarehouseBatchSchema,
	WarehouseRequestSchema,
	WarehouseSummarySchema,
	WarehousePolicySchema,
	WarehouseLinkSchema,
	WarehouseNoteSchema,
	WarehouseWindowSchema,
	WarehouseQuotaSchema,
	WarehouseBucketSchema,
	WarehouseLaneSchema,
	WarehouseSlotSchema,
	WarehouseTagSchema,
	WarehouseRuleSchema,
	WarehouseHoldSchema,
	WarehouseClaimSchema,
	WarehouseCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"warehouse/warehouserecord/",
	"warehouse/warehouseentry/",
	"warehouse/warehouseitem/",
	"warehouse/warehousedraft/",
	"warehouse/warehouserevision/",
	"warehouse/warehousesnapshot/",
	"warehouse/warehousebatch/",
	"warehouse/warehouserequest/",
	"warehouse/warehousesummary/",
	"warehouse/warehousepolicy/",
	"warehouse/warehouselink/",
	"warehouse/warehousenote/",
	"warehouse/warehousewindow/",
	"warehouse/warehousequota/",
	"warehouse/warehousebucket/",
	"warehouse/warehouselane/",
	"warehouse/warehouseslot/",
	"warehouse/warehousetag/",
	"warehouse/warehouserule/",
	"warehouse/warehousehold/",
	"warehouse/warehouseclaim/",
	"warehouse/warehousecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "warehouse: " + strings.Join(names, ", ")
}
