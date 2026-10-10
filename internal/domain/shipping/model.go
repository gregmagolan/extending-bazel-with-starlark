// Package shipping holds the shipping domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package shipping

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// ShippingRecordSchema is the stored shape of a ShippingRecord.
var ShippingRecordSchema = schema.New("shipping.shippingrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingRecord is one shipping record.
type ShippingRecord struct {
	ID       string
	Channel  string
	Label    string
	Group    string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingRecord) Key() string { return "shipping/shippingrecord/" + x.ID }

// Fields encodes x for storage.
func (x ShippingRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingRecord decodes a stored record.
func ParseShippingRecord(f codec.Fields) (ShippingRecord, error) {
	var x ShippingRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingrecord: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingRecord) String() string {
	return fmt.Sprintf("ShippingRecord(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingRecord) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingRecord) Normalize() ShippingRecord {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortShippingRecords orders by attempts descending, then by ID.
func SortShippingRecords(xs []ShippingRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingRecords drops the archived entries, preserving order.
func ActiveShippingRecords(xs []ShippingRecord) []ShippingRecord {
	out := make([]ShippingRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingRecords adds up the attempts of every active entry.
func SumShippingRecords(xs []ShippingRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexShippingRecords keys the entries by ID, last write winning.
func IndexShippingRecords(xs []ShippingRecord) map[string]ShippingRecord {
	out := make(map[string]ShippingRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingRecord builds a valid ShippingRecord, for tests and for the seed data the cron
// binary writes.
func SampleShippingRecord(id string, at time.Time) ShippingRecord {
	return ShippingRecord{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Group:    "seeded by shipping.SampleShippingRecord",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingEntrySchema is the stored shape of a ShippingEntry.
var ShippingEntrySchema = schema.New("shipping.shippingentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingEntry is one shipping entry.
type ShippingEntry struct {
	ID       string
	Source   string
	Channel  string
	Label    string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingEntry) Key() string { return "shipping/shippingentry/" + x.ID }

// Fields encodes x for storage.
func (x ShippingEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingEntry decodes a stored record.
func ParseShippingEntry(f codec.Fields) (ShippingEntry, error) {
	var x ShippingEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingentry: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingEntry) String() string {
	return fmt.Sprintf("ShippingEntry(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingEntry) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingEntry) Normalize() ShippingEntry {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortShippingEntrys orders by weight descending, then by ID.
func SortShippingEntrys(xs []ShippingEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingEntrys drops the archived entries, preserving order.
func ActiveShippingEntrys(xs []ShippingEntry) []ShippingEntry {
	out := make([]ShippingEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingEntrys adds up the weight of every active entry.
func SumShippingEntrys(xs []ShippingEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexShippingEntrys keys the entries by ID, last write winning.
func IndexShippingEntrys(xs []ShippingEntry) map[string]ShippingEntry {
	out := make(map[string]ShippingEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingEntry builds a valid ShippingEntry, for tests and for the seed data the cron
// binary writes.
func SampleShippingEntry(id string, at time.Time) ShippingEntry {
	return ShippingEntry{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by shipping.SampleShippingEntry",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingItemSchema is the stored shape of a ShippingItem.
var ShippingItemSchema = schema.New("shipping.shippingitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingItem is one shipping item.
type ShippingItem struct {
	ID        string
	Channel   string
	Group     string
	Reference string
	Capacity  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ShippingItem) Key() string { return "shipping/shippingitem/" + x.ID }

// Fields encodes x for storage.
func (x ShippingItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"group":     x.Group,
		"reference": x.Reference,
		"capacity":  strconv.Itoa(x.Capacity),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingItem decodes a stored record.
func ParseShippingItem(f codec.Fields) (ShippingItem, error) {
	var x ShippingItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingitem: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingitem %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingItem) String() string {
	return fmt.Sprintf("ShippingItem(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingItem) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingItem) Normalize() ShippingItem {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortShippingItems orders by capacity descending, then by ID.
func SortShippingItems(xs []ShippingItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingItems drops the archived entries, preserving order.
func ActiveShippingItems(xs []ShippingItem) []ShippingItem {
	out := make([]ShippingItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingItems adds up the capacity of every active entry.
func SumShippingItems(xs []ShippingItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexShippingItems keys the entries by ID, last write winning.
func IndexShippingItems(xs []ShippingItem) map[string]ShippingItem {
	out := make(map[string]ShippingItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingItem builds a valid ShippingItem, for tests and for the seed data the cron
// binary writes.
func SampleShippingItem(id string, at time.Time) ShippingItem {
	return ShippingItem{
		ID:        id,
		Channel:   "channel-" + id,
		Group:     "group-" + id,
		Reference: "seeded by shipping.SampleShippingItem",
		Capacity:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ShippingDraftSchema is the stored shape of a ShippingDraft.
var ShippingDraftSchema = schema.New("shipping.shippingdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingDraft is one shipping draft.
type ShippingDraft struct {
	ID        string
	Source    string
	Category  string
	Reference string
	Capacity  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ShippingDraft) Key() string { return "shipping/shippingdraft/" + x.ID }

// Fields encodes x for storage.
func (x ShippingDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"category":  x.Category,
		"reference": x.Reference,
		"capacity":  strconv.Itoa(x.Capacity),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingDraft decodes a stored record.
func ParseShippingDraft(f codec.Fields) (ShippingDraft, error) {
	var x ShippingDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingdraft: missing id")
	}
	x.Source = f["source"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingdraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingDraft) String() string {
	return fmt.Sprintf("ShippingDraft(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingDraft) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingDraft) Normalize() ShippingDraft {
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortShippingDrafts orders by capacity descending, then by ID.
func SortShippingDrafts(xs []ShippingDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingDrafts drops the archived entries, preserving order.
func ActiveShippingDrafts(xs []ShippingDraft) []ShippingDraft {
	out := make([]ShippingDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingDrafts adds up the capacity of every active entry.
func SumShippingDrafts(xs []ShippingDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexShippingDrafts keys the entries by ID, last write winning.
func IndexShippingDrafts(xs []ShippingDraft) map[string]ShippingDraft {
	out := make(map[string]ShippingDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingDraft builds a valid ShippingDraft, for tests and for the seed data the cron
// binary writes.
func SampleShippingDraft(id string, at time.Time) ShippingDraft {
	return ShippingDraft{
		ID:        id,
		Source:    "source-" + id,
		Category:  "category-" + id,
		Reference: "seeded by shipping.SampleShippingDraft",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ShippingRevisionSchema is the stored shape of a ShippingRevision.
var ShippingRevisionSchema = schema.New("shipping.shippingrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingRevision is one shipping revision.
type ShippingRevision struct {
	ID       string
	Region   string
	Label    string
	Channel  string
	Weight   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingRevision) Key() string { return "shipping/shippingrevision/" + x.ID }

// Fields encodes x for storage.
func (x ShippingRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingRevision decodes a stored record.
func ParseShippingRevision(f codec.Fields) (ShippingRevision, error) {
	var x ShippingRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingrevision: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrevision %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingRevision) String() string {
	return fmt.Sprintf("ShippingRevision(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingRevision) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingRevision) Normalize() ShippingRevision {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortShippingRevisions orders by weight descending, then by ID.
func SortShippingRevisions(xs []ShippingRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingRevisions drops the archived entries, preserving order.
func ActiveShippingRevisions(xs []ShippingRevision) []ShippingRevision {
	out := make([]ShippingRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingRevisions adds up the weight of every active entry.
func SumShippingRevisions(xs []ShippingRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexShippingRevisions keys the entries by ID, last write winning.
func IndexShippingRevisions(xs []ShippingRevision) map[string]ShippingRevision {
	out := make(map[string]ShippingRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingRevision builds a valid ShippingRevision, for tests and for the seed data the cron
// binary writes.
func SampleShippingRevision(id string, at time.Time) ShippingRevision {
	return ShippingRevision{
		ID:      id,
		Region:  "region-" + id,
		Label:   "label-" + id,
		Channel: "seeded by shipping.SampleShippingRevision",
		Weight:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// ShippingSnapshotSchema is the stored shape of a ShippingSnapshot.
var ShippingSnapshotSchema = schema.New("shipping.shippingsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingSnapshot is one shipping snapshot.
type ShippingSnapshot struct {
	ID       string
	Comment  string
	Channel  string
	Group    string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingSnapshot) Key() string { return "shipping/shippingsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x ShippingSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingSnapshot decodes a stored record.
func ParseShippingSnapshot(f codec.Fields) (ShippingSnapshot, error) {
	var x ShippingSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingsnapshot: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingsnapshot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingSnapshot) String() string {
	return fmt.Sprintf("ShippingSnapshot(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingSnapshot) Normalize() ShippingSnapshot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortShippingSnapshots orders by priority descending, then by ID.
func SortShippingSnapshots(xs []ShippingSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingSnapshots drops the archived entries, preserving order.
func ActiveShippingSnapshots(xs []ShippingSnapshot) []ShippingSnapshot {
	out := make([]ShippingSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingSnapshots adds up the priority of every active entry.
func SumShippingSnapshots(xs []ShippingSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexShippingSnapshots keys the entries by ID, last write winning.
func IndexShippingSnapshots(xs []ShippingSnapshot) map[string]ShippingSnapshot {
	out := make(map[string]ShippingSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingSnapshot builds a valid ShippingSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleShippingSnapshot(id string, at time.Time) ShippingSnapshot {
	return ShippingSnapshot{
		ID:       id,
		Comment:  "comment-" + id,
		Channel:  "channel-" + id,
		Group:    "seeded by shipping.SampleShippingSnapshot",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingBatchSchema is the stored shape of a ShippingBatch.
var ShippingBatchSchema = schema.New("shipping.shippingbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingBatch is one shipping batch.
type ShippingBatch struct {
	ID       string
	Owner    string
	Comment  string
	Channel  string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingBatch) Key() string { return "shipping/shippingbatch/" + x.ID }

// Fields encodes x for storage.
func (x ShippingBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingBatch decodes a stored record.
func ParseShippingBatch(f codec.Fields) (ShippingBatch, error) {
	var x ShippingBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingbatch: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingbatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingBatch) String() string {
	return fmt.Sprintf("ShippingBatch(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingBatch) Normalize() ShippingBatch {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortShippingBatchs orders by version descending, then by ID.
func SortShippingBatchs(xs []ShippingBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingBatchs drops the archived entries, preserving order.
func ActiveShippingBatchs(xs []ShippingBatch) []ShippingBatch {
	out := make([]ShippingBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingBatchs adds up the version of every active entry.
func SumShippingBatchs(xs []ShippingBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexShippingBatchs keys the entries by ID, last write winning.
func IndexShippingBatchs(xs []ShippingBatch) map[string]ShippingBatch {
	out := make(map[string]ShippingBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingBatch builds a valid ShippingBatch, for tests and for the seed data the cron
// binary writes.
func SampleShippingBatch(id string, at time.Time) ShippingBatch {
	return ShippingBatch{
		ID:      id,
		Owner:   "owner-" + id,
		Comment: "comment-" + id,
		Channel: "seeded by shipping.SampleShippingBatch",
		Version: len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// ShippingRequestSchema is the stored shape of a ShippingRequest.
var ShippingRequestSchema = schema.New("shipping.shippingrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingRequest is one shipping request.
type ShippingRequest struct {
	ID       string
	Region   string
	Channel  string
	Comment  string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingRequest) Key() string { return "shipping/shippingrequest/" + x.ID }

// Fields encodes x for storage.
func (x ShippingRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingRequest decodes a stored record.
func ParseShippingRequest(f codec.Fields) (ShippingRequest, error) {
	var x ShippingRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingrequest: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingRequest) String() string {
	return fmt.Sprintf("ShippingRequest(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingRequest) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingRequest) Normalize() ShippingRequest {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortShippingRequests orders by version descending, then by ID.
func SortShippingRequests(xs []ShippingRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingRequests drops the archived entries, preserving order.
func ActiveShippingRequests(xs []ShippingRequest) []ShippingRequest {
	out := make([]ShippingRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingRequests adds up the version of every active entry.
func SumShippingRequests(xs []ShippingRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexShippingRequests keys the entries by ID, last write winning.
func IndexShippingRequests(xs []ShippingRequest) map[string]ShippingRequest {
	out := make(map[string]ShippingRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingRequest builds a valid ShippingRequest, for tests and for the seed data the cron
// binary writes.
func SampleShippingRequest(id string, at time.Time) ShippingRequest {
	return ShippingRequest{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by shipping.SampleShippingRequest",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingSummarySchema is the stored shape of a ShippingSummary.
var ShippingSummarySchema = schema.New("shipping.shippingsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingSummary is one shipping summary.
type ShippingSummary struct {
	ID       string
	Source   string
	Name     string
	Channel  string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingSummary) Key() string { return "shipping/shippingsummary/" + x.ID }

// Fields encodes x for storage.
func (x ShippingSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"channel":  x.Channel,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingSummary decodes a stored record.
func ParseShippingSummary(f codec.Fields) (ShippingSummary, error) {
	var x ShippingSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingsummary: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingsummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingSummary) String() string {
	return fmt.Sprintf("ShippingSummary(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingSummary) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingSummary) Normalize() ShippingSummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortShippingSummarys orders by retained descending, then by ID.
func SortShippingSummarys(xs []ShippingSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingSummarys drops the archived entries, preserving order.
func ActiveShippingSummarys(xs []ShippingSummary) []ShippingSummary {
	out := make([]ShippingSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingSummarys adds up the retained of every active entry.
func SumShippingSummarys(xs []ShippingSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexShippingSummarys keys the entries by ID, last write winning.
func IndexShippingSummarys(xs []ShippingSummary) map[string]ShippingSummary {
	out := make(map[string]ShippingSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingSummary builds a valid ShippingSummary, for tests and for the seed data the cron
// binary writes.
func SampleShippingSummary(id string, at time.Time) ShippingSummary {
	return ShippingSummary{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by shipping.SampleShippingSummary",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingPolicySchema is the stored shape of a ShippingPolicy.
var ShippingPolicySchema = schema.New("shipping.shippingpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingPolicy is one shipping policy.
type ShippingPolicy struct {
	ID       string
	Name     string
	Group    string
	Channel  string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingPolicy) Key() string { return "shipping/shippingpolicy/" + x.ID }

// Fields encodes x for storage.
func (x ShippingPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"group":    x.Group,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingPolicy decodes a stored record.
func ParseShippingPolicy(f codec.Fields) (ShippingPolicy, error) {
	var x ShippingPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingpolicy: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingpolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingPolicy) String() string {
	return fmt.Sprintf("ShippingPolicy(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingPolicy) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingPolicy) Normalize() ShippingPolicy {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortShippingPolicys orders by weight descending, then by ID.
func SortShippingPolicys(xs []ShippingPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingPolicys drops the archived entries, preserving order.
func ActiveShippingPolicys(xs []ShippingPolicy) []ShippingPolicy {
	out := make([]ShippingPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingPolicys adds up the weight of every active entry.
func SumShippingPolicys(xs []ShippingPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexShippingPolicys keys the entries by ID, last write winning.
func IndexShippingPolicys(xs []ShippingPolicy) map[string]ShippingPolicy {
	out := make(map[string]ShippingPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingPolicy builds a valid ShippingPolicy, for tests and for the seed data the cron
// binary writes.
func SampleShippingPolicy(id string, at time.Time) ShippingPolicy {
	return ShippingPolicy{
		ID:       id,
		Name:     "name-" + id,
		Group:    "group-" + id,
		Channel:  "seeded by shipping.SampleShippingPolicy",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingLinkSchema is the stored shape of a ShippingLink.
var ShippingLinkSchema = schema.New("shipping.shippinglink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingLink is one shipping link.
type ShippingLink struct {
	ID       string
	Owner    string
	Name     string
	Comment  string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingLink) Key() string { return "shipping/shippinglink/" + x.ID }

// Fields encodes x for storage.
func (x ShippingLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"name":     x.Name,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingLink decodes a stored record.
func ParseShippingLink(f codec.Fields) (ShippingLink, error) {
	var x ShippingLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippinglink: missing id")
	}
	x.Owner = f["owner"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("shipping: shippinglink %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("shipping: shippinglink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippinglink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippinglink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingLink) String() string {
	return fmt.Sprintf("ShippingLink(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingLink) Normalize() ShippingLink {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortShippingLinks orders by retained descending, then by ID.
func SortShippingLinks(xs []ShippingLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingLinks drops the archived entries, preserving order.
func ActiveShippingLinks(xs []ShippingLink) []ShippingLink {
	out := make([]ShippingLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingLinks adds up the retained of every active entry.
func SumShippingLinks(xs []ShippingLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexShippingLinks keys the entries by ID, last write winning.
func IndexShippingLinks(xs []ShippingLink) map[string]ShippingLink {
	out := make(map[string]ShippingLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingLink builds a valid ShippingLink, for tests and for the seed data the cron
// binary writes.
func SampleShippingLink(id string, at time.Time) ShippingLink {
	return ShippingLink{
		ID:       id,
		Owner:    "owner-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by shipping.SampleShippingLink",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingNoteSchema is the stored shape of a ShippingNote.
var ShippingNoteSchema = schema.New("shipping.shippingnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingNote is one shipping note.
type ShippingNote struct {
	ID        string
	Channel   string
	Reference string
	Label     string
	Count     int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ShippingNote) Key() string { return "shipping/shippingnote/" + x.ID }

// Fields encodes x for storage.
func (x ShippingNote) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"label":     x.Label,
		"count":     strconv.Itoa(x.Count),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingNote decodes a stored record.
func ParseShippingNote(f codec.Fields) (ShippingNote, error) {
	var x ShippingNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingnote: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingnote %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingNote) String() string {
	return fmt.Sprintf("ShippingNote(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingNote) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingNote) Normalize() ShippingNote {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortShippingNotes orders by count descending, then by ID.
func SortShippingNotes(xs []ShippingNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingNotes drops the archived entries, preserving order.
func ActiveShippingNotes(xs []ShippingNote) []ShippingNote {
	out := make([]ShippingNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingNotes adds up the count of every active entry.
func SumShippingNotes(xs []ShippingNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexShippingNotes keys the entries by ID, last write winning.
func IndexShippingNotes(xs []ShippingNote) map[string]ShippingNote {
	out := make(map[string]ShippingNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingNote builds a valid ShippingNote, for tests and for the seed data the cron
// binary writes.
func SampleShippingNote(id string, at time.Time) ShippingNote {
	return ShippingNote{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Label:     "seeded by shipping.SampleShippingNote",
		Count:     len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ShippingWindowSchema is the stored shape of a ShippingWindow.
var ShippingWindowSchema = schema.New("shipping.shippingwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingWindow is one shipping window.
type ShippingWindow struct {
	ID       string
	Group    string
	Owner    string
	Channel  string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingWindow) Key() string { return "shipping/shippingwindow/" + x.ID }

// Fields encodes x for storage.
func (x ShippingWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingWindow decodes a stored record.
func ParseShippingWindow(f codec.Fields) (ShippingWindow, error) {
	var x ShippingWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingwindow: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingwindow %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingWindow) String() string {
	return fmt.Sprintf("ShippingWindow(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingWindow) Normalize() ShippingWindow {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortShippingWindows orders by priority descending, then by ID.
func SortShippingWindows(xs []ShippingWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingWindows drops the archived entries, preserving order.
func ActiveShippingWindows(xs []ShippingWindow) []ShippingWindow {
	out := make([]ShippingWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingWindows adds up the priority of every active entry.
func SumShippingWindows(xs []ShippingWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexShippingWindows keys the entries by ID, last write winning.
func IndexShippingWindows(xs []ShippingWindow) map[string]ShippingWindow {
	out := make(map[string]ShippingWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingWindow builds a valid ShippingWindow, for tests and for the seed data the cron
// binary writes.
func SampleShippingWindow(id string, at time.Time) ShippingWindow {
	return ShippingWindow{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by shipping.SampleShippingWindow",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingQuotaSchema is the stored shape of a ShippingQuota.
var ShippingQuotaSchema = schema.New("shipping.shippingquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingQuota is one shipping quota.
type ShippingQuota struct {
	ID       string
	Category string
	Label    string
	Channel  string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingQuota) Key() string { return "shipping/shippingquota/" + x.ID }

// Fields encodes x for storage.
func (x ShippingQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingQuota decodes a stored record.
func ParseShippingQuota(f codec.Fields) (ShippingQuota, error) {
	var x ShippingQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingquota: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingquota %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingQuota) String() string {
	return fmt.Sprintf("ShippingQuota(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingQuota) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingQuota) Normalize() ShippingQuota {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortShippingQuotas orders by weight descending, then by ID.
func SortShippingQuotas(xs []ShippingQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingQuotas drops the archived entries, preserving order.
func ActiveShippingQuotas(xs []ShippingQuota) []ShippingQuota {
	out := make([]ShippingQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingQuotas adds up the weight of every active entry.
func SumShippingQuotas(xs []ShippingQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexShippingQuotas keys the entries by ID, last write winning.
func IndexShippingQuotas(xs []ShippingQuota) map[string]ShippingQuota {
	out := make(map[string]ShippingQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingQuota builds a valid ShippingQuota, for tests and for the seed data the cron
// binary writes.
func SampleShippingQuota(id string, at time.Time) ShippingQuota {
	return ShippingQuota{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by shipping.SampleShippingQuota",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingBucketSchema is the stored shape of a ShippingBucket.
var ShippingBucketSchema = schema.New("shipping.shippingbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingBucket is one shipping bucket.
type ShippingBucket struct {
	ID       string
	Label    string
	Source   string
	Channel  string
	Priority int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingBucket) Key() string { return "shipping/shippingbucket/" + x.ID }

// Fields encodes x for storage.
func (x ShippingBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingBucket decodes a stored record.
func ParseShippingBucket(f codec.Fields) (ShippingBucket, error) {
	var x ShippingBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingbucket: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingbucket %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingBucket) String() string {
	return fmt.Sprintf("ShippingBucket(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingBucket) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingBucket) Normalize() ShippingBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortShippingBuckets orders by priority descending, then by ID.
func SortShippingBuckets(xs []ShippingBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingBuckets drops the archived entries, preserving order.
func ActiveShippingBuckets(xs []ShippingBucket) []ShippingBucket {
	out := make([]ShippingBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingBuckets adds up the priority of every active entry.
func SumShippingBuckets(xs []ShippingBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexShippingBuckets keys the entries by ID, last write winning.
func IndexShippingBuckets(xs []ShippingBucket) map[string]ShippingBucket {
	out := make(map[string]ShippingBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingBucket builds a valid ShippingBucket, for tests and for the seed data the cron
// binary writes.
func SampleShippingBucket(id string, at time.Time) ShippingBucket {
	return ShippingBucket{
		ID:       id,
		Label:    "label-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by shipping.SampleShippingBucket",
		Priority: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingLaneSchema is the stored shape of a ShippingLane.
var ShippingLaneSchema = schema.New("shipping.shippinglane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingLane is one shipping lane.
type ShippingLane struct {
	ID        string
	Reference string
	Region    string
	Name      string
	Priority  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ShippingLane) Key() string { return "shipping/shippinglane/" + x.ID }

// Fields encodes x for storage.
func (x ShippingLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"name":      x.Name,
		"priority":  strconv.Itoa(x.Priority),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingLane decodes a stored record.
func ParseShippingLane(f codec.Fields) (ShippingLane, error) {
	var x ShippingLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippinglane: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippinglane %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("shipping: shippinglane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippinglane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippinglane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingLane) String() string {
	return fmt.Sprintf("ShippingLane(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingLane) Normalize() ShippingLane {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortShippingLanes orders by priority descending, then by ID.
func SortShippingLanes(xs []ShippingLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingLanes drops the archived entries, preserving order.
func ActiveShippingLanes(xs []ShippingLane) []ShippingLane {
	out := make([]ShippingLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingLanes adds up the priority of every active entry.
func SumShippingLanes(xs []ShippingLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexShippingLanes keys the entries by ID, last write winning.
func IndexShippingLanes(xs []ShippingLane) map[string]ShippingLane {
	out := make(map[string]ShippingLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingLane builds a valid ShippingLane, for tests and for the seed data the cron
// binary writes.
func SampleShippingLane(id string, at time.Time) ShippingLane {
	return ShippingLane{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Name:      "seeded by shipping.SampleShippingLane",
		Priority:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ShippingSlotSchema is the stored shape of a ShippingSlot.
var ShippingSlotSchema = schema.New("shipping.shippingslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingSlot is one shipping slot.
type ShippingSlot struct {
	ID        string
	Region    string
	Group     string
	Reference string
	Offset    int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ShippingSlot) Key() string { return "shipping/shippingslot/" + x.ID }

// Fields encodes x for storage.
func (x ShippingSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"group":     x.Group,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingSlot decodes a stored record.
func ParseShippingSlot(f codec.Fields) (ShippingSlot, error) {
	var x ShippingSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingslot: missing id")
	}
	x.Region = f["region"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingslot %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingSlot) String() string {
	return fmt.Sprintf("ShippingSlot(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingSlot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingSlot) Normalize() ShippingSlot {
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortShippingSlots orders by offset descending, then by ID.
func SortShippingSlots(xs []ShippingSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingSlots drops the archived entries, preserving order.
func ActiveShippingSlots(xs []ShippingSlot) []ShippingSlot {
	out := make([]ShippingSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingSlots adds up the offset of every active entry.
func SumShippingSlots(xs []ShippingSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexShippingSlots keys the entries by ID, last write winning.
func IndexShippingSlots(xs []ShippingSlot) map[string]ShippingSlot {
	out := make(map[string]ShippingSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingSlot builds a valid ShippingSlot, for tests and for the seed data the cron
// binary writes.
func SampleShippingSlot(id string, at time.Time) ShippingSlot {
	return ShippingSlot{
		ID:        id,
		Region:    "region-" + id,
		Group:     "group-" + id,
		Reference: "seeded by shipping.SampleShippingSlot",
		Offset:    len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ShippingTagSchema is the stored shape of a ShippingTag.
var ShippingTagSchema = schema.New("shipping.shippingtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingTag is one shipping tag.
type ShippingTag struct {
	ID       string
	Group    string
	Source   string
	Region   string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingTag) Key() string { return "shipping/shippingtag/" + x.ID }

// Fields encodes x for storage.
func (x ShippingTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"region":   x.Region,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingTag decodes a stored record.
func ParseShippingTag(f codec.Fields) (ShippingTag, error) {
	var x ShippingTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingtag: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingtag %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingTag) String() string {
	return fmt.Sprintf("ShippingTag(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingTag) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingTag) Normalize() ShippingTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortShippingTags orders by priority descending, then by ID.
func SortShippingTags(xs []ShippingTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingTags drops the archived entries, preserving order.
func ActiveShippingTags(xs []ShippingTag) []ShippingTag {
	out := make([]ShippingTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingTags adds up the priority of every active entry.
func SumShippingTags(xs []ShippingTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexShippingTags keys the entries by ID, last write winning.
func IndexShippingTags(xs []ShippingTag) map[string]ShippingTag {
	out := make(map[string]ShippingTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingTag builds a valid ShippingTag, for tests and for the seed data the cron
// binary writes.
func SampleShippingTag(id string, at time.Time) ShippingTag {
	return ShippingTag{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Region:   "seeded by shipping.SampleShippingTag",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingRuleSchema is the stored shape of a ShippingRule.
var ShippingRuleSchema = schema.New("shipping.shippingrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingRule is one shipping rule.
type ShippingRule struct {
	ID       string
	Channel  string
	Region   string
	Category string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingRule) Key() string { return "shipping/shippingrule/" + x.ID }

// Fields encodes x for storage.
func (x ShippingRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingRule decodes a stored record.
func ParseShippingRule(f codec.Fields) (ShippingRule, error) {
	var x ShippingRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingrule: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrule %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingRule) String() string {
	return fmt.Sprintf("ShippingRule(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingRule) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingRule) Normalize() ShippingRule {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortShippingRules orders by capacity descending, then by ID.
func SortShippingRules(xs []ShippingRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingRules drops the archived entries, preserving order.
func ActiveShippingRules(xs []ShippingRule) []ShippingRule {
	out := make([]ShippingRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingRules adds up the capacity of every active entry.
func SumShippingRules(xs []ShippingRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexShippingRules keys the entries by ID, last write winning.
func IndexShippingRules(xs []ShippingRule) map[string]ShippingRule {
	out := make(map[string]ShippingRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingRule builds a valid ShippingRule, for tests and for the seed data the cron
// binary writes.
func SampleShippingRule(id string, at time.Time) ShippingRule {
	return ShippingRule{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Category: "seeded by shipping.SampleShippingRule",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingHoldSchema is the stored shape of a ShippingHold.
var ShippingHoldSchema = schema.New("shipping.shippinghold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingHold is one shipping hold.
type ShippingHold struct {
	ID        string
	Group     string
	Reference string
	Comment   string
	Priority  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ShippingHold) Key() string { return "shipping/shippinghold/" + x.ID }

// Fields encodes x for storage.
func (x ShippingHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"comment":   x.Comment,
		"priority":  strconv.Itoa(x.Priority),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingHold decodes a stored record.
func ParseShippingHold(f codec.Fields) (ShippingHold, error) {
	var x ShippingHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippinghold: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("shipping: shippinghold %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("shipping: shippinghold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippinghold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippinghold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingHold) String() string {
	return fmt.Sprintf("ShippingHold(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingHold) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingHold) Normalize() ShippingHold {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortShippingHolds orders by priority descending, then by ID.
func SortShippingHolds(xs []ShippingHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingHolds drops the archived entries, preserving order.
func ActiveShippingHolds(xs []ShippingHold) []ShippingHold {
	out := make([]ShippingHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingHolds adds up the priority of every active entry.
func SumShippingHolds(xs []ShippingHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexShippingHolds keys the entries by ID, last write winning.
func IndexShippingHolds(xs []ShippingHold) map[string]ShippingHold {
	out := make(map[string]ShippingHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingHold builds a valid ShippingHold, for tests and for the seed data the cron
// binary writes.
func SampleShippingHold(id string, at time.Time) ShippingHold {
	return ShippingHold{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by shipping.SampleShippingHold",
		Priority:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ShippingClaimSchema is the stored shape of a ShippingClaim.
var ShippingClaimSchema = schema.New("shipping.shippingclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingClaim is one shipping claim.
type ShippingClaim struct {
	ID       string
	Channel  string
	Name     string
	Group    string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ShippingClaim) Key() string { return "shipping/shippingclaim/" + x.ID }

// Fields encodes x for storage.
func (x ShippingClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"name":     x.Name,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingClaim decodes a stored record.
func ParseShippingClaim(f codec.Fields) (ShippingClaim, error) {
	var x ShippingClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingclaim: missing id")
	}
	x.Channel = f["channel"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingclaim %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingClaim) String() string {
	return fmt.Sprintf("ShippingClaim(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingClaim) Normalize() ShippingClaim {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortShippingClaims orders by retained descending, then by ID.
func SortShippingClaims(xs []ShippingClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingClaims drops the archived entries, preserving order.
func ActiveShippingClaims(xs []ShippingClaim) []ShippingClaim {
	out := make([]ShippingClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingClaims adds up the retained of every active entry.
func SumShippingClaims(xs []ShippingClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexShippingClaims keys the entries by ID, last write winning.
func IndexShippingClaims(xs []ShippingClaim) map[string]ShippingClaim {
	out := make(map[string]ShippingClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingClaim builds a valid ShippingClaim, for tests and for the seed data the cron
// binary writes.
func SampleShippingClaim(id string, at time.Time) ShippingClaim {
	return ShippingClaim{
		ID:       id,
		Channel:  "channel-" + id,
		Name:     "name-" + id,
		Group:    "seeded by shipping.SampleShippingClaim",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ShippingCursorSchema is the stored shape of a ShippingCursor.
var ShippingCursorSchema = schema.New("shipping.shippingcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ShippingCursor is one shipping cursor.
type ShippingCursor struct {
	ID        string
	Channel   string
	Reference string
	Category  string
	Count     int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ShippingCursor) Key() string { return "shipping/shippingcursor/" + x.ID }

// Fields encodes x for storage.
func (x ShippingCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"category":  x.Category,
		"count":     strconv.Itoa(x.Count),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseShippingCursor decodes a stored record.
func ParseShippingCursor(f codec.Fields) (ShippingCursor, error) {
	var x ShippingCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("shipping: shippingcursor: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingcursor %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("shipping: shippingcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("shipping: shippingcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("shipping: shippingcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ShippingCursor) String() string {
	return fmt.Sprintf("ShippingCursor(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ShippingCursor) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ShippingCursor) Normalize() ShippingCursor {
	x.Channel = strings.TrimSpace(x.Channel)
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

// SortShippingCursors orders by count descending, then by ID.
func SortShippingCursors(xs []ShippingCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveShippingCursors drops the archived entries, preserving order.
func ActiveShippingCursors(xs []ShippingCursor) []ShippingCursor {
	out := make([]ShippingCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumShippingCursors adds up the count of every active entry.
func SumShippingCursors(xs []ShippingCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexShippingCursors keys the entries by ID, last write winning.
func IndexShippingCursors(xs []ShippingCursor) map[string]ShippingCursor {
	out := make(map[string]ShippingCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleShippingCursor builds a valid ShippingCursor, for tests and for the seed data the cron
// binary writes.
func SampleShippingCursor(id string, at time.Time) ShippingCursor {
	return ShippingCursor{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by shipping.SampleShippingCursor",
		Count:     len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	ShippingRecordSchema,
	ShippingEntrySchema,
	ShippingItemSchema,
	ShippingDraftSchema,
	ShippingRevisionSchema,
	ShippingSnapshotSchema,
	ShippingBatchSchema,
	ShippingRequestSchema,
	ShippingSummarySchema,
	ShippingPolicySchema,
	ShippingLinkSchema,
	ShippingNoteSchema,
	ShippingWindowSchema,
	ShippingQuotaSchema,
	ShippingBucketSchema,
	ShippingLaneSchema,
	ShippingSlotSchema,
	ShippingTagSchema,
	ShippingRuleSchema,
	ShippingHoldSchema,
	ShippingClaimSchema,
	ShippingCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"shipping/shippingrecord/",
	"shipping/shippingentry/",
	"shipping/shippingitem/",
	"shipping/shippingdraft/",
	"shipping/shippingrevision/",
	"shipping/shippingsnapshot/",
	"shipping/shippingbatch/",
	"shipping/shippingrequest/",
	"shipping/shippingsummary/",
	"shipping/shippingpolicy/",
	"shipping/shippinglink/",
	"shipping/shippingnote/",
	"shipping/shippingwindow/",
	"shipping/shippingquota/",
	"shipping/shippingbucket/",
	"shipping/shippinglane/",
	"shipping/shippingslot/",
	"shipping/shippingtag/",
	"shipping/shippingrule/",
	"shipping/shippinghold/",
	"shipping/shippingclaim/",
	"shipping/shippingcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "shipping: " + strings.Join(names, ", ")
}
