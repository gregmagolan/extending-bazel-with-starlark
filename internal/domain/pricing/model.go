// Package pricing holds the pricing domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package pricing

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// PricingRecordSchema is the stored shape of a PricingRecord.
var PricingRecordSchema = schema.New("pricing.pricingrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingRecord is one pricing record.
type PricingRecord struct {
	ID       string
	Source   string
	Channel  string
	Label    string
	Retained int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingRecord) Key() string { return "pricing/pricingrecord/" + x.ID }

// Fields encodes x for storage.
func (x PricingRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingRecord decodes a stored record.
func ParsePricingRecord(f codec.Fields) (PricingRecord, error) {
	var x PricingRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingrecord: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrecord %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingRecord) String() string {
	return fmt.Sprintf("PricingRecord(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingRecord) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingRecord) Normalize() PricingRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPricingRecords orders by retained descending, then by ID.
func SortPricingRecords(xs []PricingRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingRecords drops the archived entries, preserving order.
func ActivePricingRecords(xs []PricingRecord) []PricingRecord {
	out := make([]PricingRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingRecords adds up the retained of every active entry.
func SumPricingRecords(xs []PricingRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPricingRecords keys the entries by ID, last write winning.
func IndexPricingRecords(xs []PricingRecord) map[string]PricingRecord {
	out := make(map[string]PricingRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingRecord builds a valid PricingRecord, for tests and for the seed data the cron
// binary writes.
func SamplePricingRecord(id string, at time.Time) PricingRecord {
	return PricingRecord{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by pricing.SamplePricingRecord",
		Retained: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingEntrySchema is the stored shape of a PricingEntry.
var PricingEntrySchema = schema.New("pricing.pricingentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingEntry is one pricing entry.
type PricingEntry struct {
	ID        string
	Comment   string
	Group     string
	Reference string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PricingEntry) Key() string { return "pricing/pricingentry/" + x.ID }

// Fields encodes x for storage.
func (x PricingEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"group":     x.Group,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingEntry decodes a stored record.
func ParsePricingEntry(f codec.Fields) (PricingEntry, error) {
	var x PricingEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingentry: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingentry %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingEntry) String() string {
	return fmt.Sprintf("PricingEntry(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingEntry) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingEntry) Normalize() PricingEntry {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortPricingEntrys orders by version descending, then by ID.
func SortPricingEntrys(xs []PricingEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingEntrys drops the archived entries, preserving order.
func ActivePricingEntrys(xs []PricingEntry) []PricingEntry {
	out := make([]PricingEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingEntrys adds up the version of every active entry.
func SumPricingEntrys(xs []PricingEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPricingEntrys keys the entries by ID, last write winning.
func IndexPricingEntrys(xs []PricingEntry) map[string]PricingEntry {
	out := make(map[string]PricingEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingEntry builds a valid PricingEntry, for tests and for the seed data the cron
// binary writes.
func SamplePricingEntry(id string, at time.Time) PricingEntry {
	return PricingEntry{
		ID:        id,
		Comment:   "comment-" + id,
		Group:     "group-" + id,
		Reference: "seeded by pricing.SamplePricingEntry",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PricingItemSchema is the stored shape of a PricingItem.
var PricingItemSchema = schema.New("pricing.pricingitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingItem is one pricing item.
type PricingItem struct {
	ID       string
	Comment  string
	Source   string
	Owner    string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingItem) Key() string { return "pricing/pricingitem/" + x.ID }

// Fields encodes x for storage.
func (x PricingItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingItem decodes a stored record.
func ParsePricingItem(f codec.Fields) (PricingItem, error) {
	var x PricingItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingitem: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingitem %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingItem) String() string {
	return fmt.Sprintf("PricingItem(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingItem) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingItem) Normalize() PricingItem {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPricingItems orders by retained descending, then by ID.
func SortPricingItems(xs []PricingItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingItems drops the archived entries, preserving order.
func ActivePricingItems(xs []PricingItem) []PricingItem {
	out := make([]PricingItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingItems adds up the retained of every active entry.
func SumPricingItems(xs []PricingItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPricingItems keys the entries by ID, last write winning.
func IndexPricingItems(xs []PricingItem) map[string]PricingItem {
	out := make(map[string]PricingItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingItem builds a valid PricingItem, for tests and for the seed data the cron
// binary writes.
func SamplePricingItem(id string, at time.Time) PricingItem {
	return PricingItem{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by pricing.SamplePricingItem",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingDraftSchema is the stored shape of a PricingDraft.
var PricingDraftSchema = schema.New("pricing.pricingdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingDraft is one pricing draft.
type PricingDraft struct {
	ID       string
	Category string
	Name     string
	Group    string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingDraft) Key() string { return "pricing/pricingdraft/" + x.ID }

// Fields encodes x for storage.
func (x PricingDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingDraft decodes a stored record.
func ParsePricingDraft(f codec.Fields) (PricingDraft, error) {
	var x PricingDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingdraft: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingdraft %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingDraft) String() string {
	return fmt.Sprintf("PricingDraft(%s, category=%q, attempts=%d, archived=%t)", x.ID, x.Category, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingDraft) Normalize() PricingDraft {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPricingDrafts orders by attempts descending, then by ID.
func SortPricingDrafts(xs []PricingDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingDrafts drops the archived entries, preserving order.
func ActivePricingDrafts(xs []PricingDraft) []PricingDraft {
	out := make([]PricingDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingDrafts adds up the attempts of every active entry.
func SumPricingDrafts(xs []PricingDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPricingDrafts keys the entries by ID, last write winning.
func IndexPricingDrafts(xs []PricingDraft) map[string]PricingDraft {
	out := make(map[string]PricingDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingDraft builds a valid PricingDraft, for tests and for the seed data the cron
// binary writes.
func SamplePricingDraft(id string, at time.Time) PricingDraft {
	return PricingDraft{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Group:    "seeded by pricing.SamplePricingDraft",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingRevisionSchema is the stored shape of a PricingRevision.
var PricingRevisionSchema = schema.New("pricing.pricingrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingRevision is one pricing revision.
type PricingRevision struct {
	ID        string
	Reference string
	Group     string
	Label     string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PricingRevision) Key() string { return "pricing/pricingrevision/" + x.ID }

// Fields encodes x for storage.
func (x PricingRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"label":     x.Label,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingRevision decodes a stored record.
func ParsePricingRevision(f codec.Fields) (PricingRevision, error) {
	var x PricingRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingrevision: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrevision %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingRevision) String() string {
	return fmt.Sprintf("PricingRevision(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingRevision) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingRevision) Normalize() PricingRevision {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPricingRevisions orders by weight descending, then by ID.
func SortPricingRevisions(xs []PricingRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingRevisions drops the archived entries, preserving order.
func ActivePricingRevisions(xs []PricingRevision) []PricingRevision {
	out := make([]PricingRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingRevisions adds up the weight of every active entry.
func SumPricingRevisions(xs []PricingRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexPricingRevisions keys the entries by ID, last write winning.
func IndexPricingRevisions(xs []PricingRevision) map[string]PricingRevision {
	out := make(map[string]PricingRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingRevision builds a valid PricingRevision, for tests and for the seed data the cron
// binary writes.
func SamplePricingRevision(id string, at time.Time) PricingRevision {
	return PricingRevision{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Label:     "seeded by pricing.SamplePricingRevision",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PricingSnapshotSchema is the stored shape of a PricingSnapshot.
var PricingSnapshotSchema = schema.New("pricing.pricingsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingSnapshot is one pricing snapshot.
type PricingSnapshot struct {
	ID        string
	Comment   string
	Reference string
	Category  string
	Offset    int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PricingSnapshot) Key() string { return "pricing/pricingsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x PricingSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"category":  x.Category,
		"offset":    strconv.Itoa(x.Offset),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingSnapshot decodes a stored record.
func ParsePricingSnapshot(f codec.Fields) (PricingSnapshot, error) {
	var x PricingSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingsnapshot: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingsnapshot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingSnapshot) String() string {
	return fmt.Sprintf("PricingSnapshot(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingSnapshot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingSnapshot) Normalize() PricingSnapshot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPricingSnapshots orders by offset descending, then by ID.
func SortPricingSnapshots(xs []PricingSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingSnapshots drops the archived entries, preserving order.
func ActivePricingSnapshots(xs []PricingSnapshot) []PricingSnapshot {
	out := make([]PricingSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingSnapshots adds up the offset of every active entry.
func SumPricingSnapshots(xs []PricingSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPricingSnapshots keys the entries by ID, last write winning.
func IndexPricingSnapshots(xs []PricingSnapshot) map[string]PricingSnapshot {
	out := make(map[string]PricingSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingSnapshot builds a valid PricingSnapshot, for tests and for the seed data the cron
// binary writes.
func SamplePricingSnapshot(id string, at time.Time) PricingSnapshot {
	return PricingSnapshot{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by pricing.SamplePricingSnapshot",
		Offset:    len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PricingBatchSchema is the stored shape of a PricingBatch.
var PricingBatchSchema = schema.New("pricing.pricingbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingBatch is one pricing batch.
type PricingBatch struct {
	ID       string
	Channel  string
	Group    string
	Label    string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingBatch) Key() string { return "pricing/pricingbatch/" + x.ID }

// Fields encodes x for storage.
func (x PricingBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"label":    x.Label,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingBatch decodes a stored record.
func ParsePricingBatch(f codec.Fields) (PricingBatch, error) {
	var x PricingBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingbatch: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingbatch %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingBatch) String() string {
	return fmt.Sprintf("PricingBatch(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingBatch) Normalize() PricingBatch {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPricingBatchs orders by version descending, then by ID.
func SortPricingBatchs(xs []PricingBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingBatchs drops the archived entries, preserving order.
func ActivePricingBatchs(xs []PricingBatch) []PricingBatch {
	out := make([]PricingBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingBatchs adds up the version of every active entry.
func SumPricingBatchs(xs []PricingBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPricingBatchs keys the entries by ID, last write winning.
func IndexPricingBatchs(xs []PricingBatch) map[string]PricingBatch {
	out := make(map[string]PricingBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingBatch builds a valid PricingBatch, for tests and for the seed data the cron
// binary writes.
func SamplePricingBatch(id string, at time.Time) PricingBatch {
	return PricingBatch{
		ID:       id,
		Channel:  "channel-" + id,
		Group:    "group-" + id,
		Label:    "seeded by pricing.SamplePricingBatch",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingRequestSchema is the stored shape of a PricingRequest.
var PricingRequestSchema = schema.New("pricing.pricingrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingRequest is one pricing request.
type PricingRequest struct {
	ID       string
	Owner    string
	Group    string
	Name     string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingRequest) Key() string { return "pricing/pricingrequest/" + x.ID }

// Fields encodes x for storage.
func (x PricingRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingRequest decodes a stored record.
func ParsePricingRequest(f codec.Fields) (PricingRequest, error) {
	var x PricingRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingrequest: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrequest %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingRequest) String() string {
	return fmt.Sprintf("PricingRequest(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingRequest) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingRequest) Normalize() PricingRequest {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortPricingRequests orders by retained descending, then by ID.
func SortPricingRequests(xs []PricingRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingRequests drops the archived entries, preserving order.
func ActivePricingRequests(xs []PricingRequest) []PricingRequest {
	out := make([]PricingRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingRequests adds up the retained of every active entry.
func SumPricingRequests(xs []PricingRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPricingRequests keys the entries by ID, last write winning.
func IndexPricingRequests(xs []PricingRequest) map[string]PricingRequest {
	out := make(map[string]PricingRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingRequest builds a valid PricingRequest, for tests and for the seed data the cron
// binary writes.
func SamplePricingRequest(id string, at time.Time) PricingRequest {
	return PricingRequest{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Name:     "seeded by pricing.SamplePricingRequest",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingSummarySchema is the stored shape of a PricingSummary.
var PricingSummarySchema = schema.New("pricing.pricingsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingSummary is one pricing summary.
type PricingSummary struct {
	ID       string
	Channel  string
	Name     string
	Group    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingSummary) Key() string { return "pricing/pricingsummary/" + x.ID }

// Fields encodes x for storage.
func (x PricingSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"name":     x.Name,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingSummary decodes a stored record.
func ParsePricingSummary(f codec.Fields) (PricingSummary, error) {
	var x PricingSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingsummary: missing id")
	}
	x.Channel = f["channel"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingsummary %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingSummary) String() string {
	return fmt.Sprintf("PricingSummary(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingSummary) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingSummary) Normalize() PricingSummary {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortPricingSummarys orders by count descending, then by ID.
func SortPricingSummarys(xs []PricingSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingSummarys drops the archived entries, preserving order.
func ActivePricingSummarys(xs []PricingSummary) []PricingSummary {
	out := make([]PricingSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingSummarys adds up the count of every active entry.
func SumPricingSummarys(xs []PricingSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPricingSummarys keys the entries by ID, last write winning.
func IndexPricingSummarys(xs []PricingSummary) map[string]PricingSummary {
	out := make(map[string]PricingSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingSummary builds a valid PricingSummary, for tests and for the seed data the cron
// binary writes.
func SamplePricingSummary(id string, at time.Time) PricingSummary {
	return PricingSummary{
		ID:      id,
		Channel: "channel-" + id,
		Name:    "name-" + id,
		Group:   "seeded by pricing.SamplePricingSummary",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// PricingPolicySchema is the stored shape of a PricingPolicy.
var PricingPolicySchema = schema.New("pricing.pricingpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingPolicy is one pricing policy.
type PricingPolicy struct {
	ID       string
	Channel  string
	Label    string
	Region   string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingPolicy) Key() string { return "pricing/pricingpolicy/" + x.ID }

// Fields encodes x for storage.
func (x PricingPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingPolicy decodes a stored record.
func ParsePricingPolicy(f codec.Fields) (PricingPolicy, error) {
	var x PricingPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingpolicy: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingpolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingPolicy) String() string {
	return fmt.Sprintf("PricingPolicy(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingPolicy) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingPolicy) Normalize() PricingPolicy {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPricingPolicys orders by capacity descending, then by ID.
func SortPricingPolicys(xs []PricingPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingPolicys drops the archived entries, preserving order.
func ActivePricingPolicys(xs []PricingPolicy) []PricingPolicy {
	out := make([]PricingPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingPolicys adds up the capacity of every active entry.
func SumPricingPolicys(xs []PricingPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPricingPolicys keys the entries by ID, last write winning.
func IndexPricingPolicys(xs []PricingPolicy) map[string]PricingPolicy {
	out := make(map[string]PricingPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingPolicy builds a valid PricingPolicy, for tests and for the seed data the cron
// binary writes.
func SamplePricingPolicy(id string, at time.Time) PricingPolicy {
	return PricingPolicy{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Region:   "seeded by pricing.SamplePricingPolicy",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingLinkSchema is the stored shape of a PricingLink.
var PricingLinkSchema = schema.New("pricing.pricinglink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingLink is one pricing link.
type PricingLink struct {
	ID       string
	Label    string
	Comment  string
	Group    string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingLink) Key() string { return "pricing/pricinglink/" + x.ID }

// Fields encodes x for storage.
func (x PricingLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"comment":  x.Comment,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingLink decodes a stored record.
func ParsePricingLink(f codec.Fields) (PricingLink, error) {
	var x PricingLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricinglink: missing id")
	}
	x.Label = f["label"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricinglink %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricinglink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricinglink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricinglink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingLink) String() string {
	return fmt.Sprintf("PricingLink(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingLink) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingLink) Normalize() PricingLink {
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPricingLinks orders by attempts descending, then by ID.
func SortPricingLinks(xs []PricingLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingLinks drops the archived entries, preserving order.
func ActivePricingLinks(xs []PricingLink) []PricingLink {
	out := make([]PricingLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingLinks adds up the attempts of every active entry.
func SumPricingLinks(xs []PricingLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPricingLinks keys the entries by ID, last write winning.
func IndexPricingLinks(xs []PricingLink) map[string]PricingLink {
	out := make(map[string]PricingLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingLink builds a valid PricingLink, for tests and for the seed data the cron
// binary writes.
func SamplePricingLink(id string, at time.Time) PricingLink {
	return PricingLink{
		ID:       id,
		Label:    "label-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by pricing.SamplePricingLink",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingNoteSchema is the stored shape of a PricingNote.
var PricingNoteSchema = schema.New("pricing.pricingnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingNote is one pricing note.
type PricingNote struct {
	ID       string
	Name     string
	Source   string
	Group    string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingNote) Key() string { return "pricing/pricingnote/" + x.ID }

// Fields encodes x for storage.
func (x PricingNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingNote decodes a stored record.
func ParsePricingNote(f codec.Fields) (PricingNote, error) {
	var x PricingNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingnote: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingnote %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingNote) String() string {
	return fmt.Sprintf("PricingNote(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingNote) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingNote) Normalize() PricingNote {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortPricingNotes orders by capacity descending, then by ID.
func SortPricingNotes(xs []PricingNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingNotes drops the archived entries, preserving order.
func ActivePricingNotes(xs []PricingNote) []PricingNote {
	out := make([]PricingNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingNotes adds up the capacity of every active entry.
func SumPricingNotes(xs []PricingNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPricingNotes keys the entries by ID, last write winning.
func IndexPricingNotes(xs []PricingNote) map[string]PricingNote {
	out := make(map[string]PricingNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingNote builds a valid PricingNote, for tests and for the seed data the cron
// binary writes.
func SamplePricingNote(id string, at time.Time) PricingNote {
	return PricingNote{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Group:    "seeded by pricing.SamplePricingNote",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingWindowSchema is the stored shape of a PricingWindow.
var PricingWindowSchema = schema.New("pricing.pricingwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingWindow is one pricing window.
type PricingWindow struct {
	ID       string
	Channel  string
	Category string
	Comment  string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingWindow) Key() string { return "pricing/pricingwindow/" + x.ID }

// Fields encodes x for storage.
func (x PricingWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"comment":  x.Comment,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingWindow decodes a stored record.
func ParsePricingWindow(f codec.Fields) (PricingWindow, error) {
	var x PricingWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingwindow: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingwindow %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingWindow) String() string {
	return fmt.Sprintf("PricingWindow(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingWindow) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingWindow) Normalize() PricingWindow {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPricingWindows orders by offset descending, then by ID.
func SortPricingWindows(xs []PricingWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingWindows drops the archived entries, preserving order.
func ActivePricingWindows(xs []PricingWindow) []PricingWindow {
	out := make([]PricingWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingWindows adds up the offset of every active entry.
func SumPricingWindows(xs []PricingWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPricingWindows keys the entries by ID, last write winning.
func IndexPricingWindows(xs []PricingWindow) map[string]PricingWindow {
	out := make(map[string]PricingWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingWindow builds a valid PricingWindow, for tests and for the seed data the cron
// binary writes.
func SamplePricingWindow(id string, at time.Time) PricingWindow {
	return PricingWindow{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Comment:  "seeded by pricing.SamplePricingWindow",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingQuotaSchema is the stored shape of a PricingQuota.
var PricingQuotaSchema = schema.New("pricing.pricingquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingQuota is one pricing quota.
type PricingQuota struct {
	ID       string
	Group    string
	Category string
	Name     string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingQuota) Key() string { return "pricing/pricingquota/" + x.ID }

// Fields encodes x for storage.
func (x PricingQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingQuota decodes a stored record.
func ParsePricingQuota(f codec.Fields) (PricingQuota, error) {
	var x PricingQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingquota: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingquota %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingQuota) String() string {
	return fmt.Sprintf("PricingQuota(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingQuota) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingQuota) Normalize() PricingQuota {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortPricingQuotas orders by attempts descending, then by ID.
func SortPricingQuotas(xs []PricingQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingQuotas drops the archived entries, preserving order.
func ActivePricingQuotas(xs []PricingQuota) []PricingQuota {
	out := make([]PricingQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingQuotas adds up the attempts of every active entry.
func SumPricingQuotas(xs []PricingQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPricingQuotas keys the entries by ID, last write winning.
func IndexPricingQuotas(xs []PricingQuota) map[string]PricingQuota {
	out := make(map[string]PricingQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingQuota builds a valid PricingQuota, for tests and for the seed data the cron
// binary writes.
func SamplePricingQuota(id string, at time.Time) PricingQuota {
	return PricingQuota{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Name:     "seeded by pricing.SamplePricingQuota",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingBucketSchema is the stored shape of a PricingBucket.
var PricingBucketSchema = schema.New("pricing.pricingbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingBucket is one pricing bucket.
type PricingBucket struct {
	ID       string
	Group    string
	Source   string
	Category string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingBucket) Key() string { return "pricing/pricingbucket/" + x.ID }

// Fields encodes x for storage.
func (x PricingBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingBucket decodes a stored record.
func ParsePricingBucket(f codec.Fields) (PricingBucket, error) {
	var x PricingBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingbucket: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingbucket %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingBucket) String() string {
	return fmt.Sprintf("PricingBucket(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingBucket) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingBucket) Normalize() PricingBucket {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPricingBuckets orders by capacity descending, then by ID.
func SortPricingBuckets(xs []PricingBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingBuckets drops the archived entries, preserving order.
func ActivePricingBuckets(xs []PricingBucket) []PricingBucket {
	out := make([]PricingBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingBuckets adds up the capacity of every active entry.
func SumPricingBuckets(xs []PricingBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPricingBuckets keys the entries by ID, last write winning.
func IndexPricingBuckets(xs []PricingBucket) map[string]PricingBucket {
	out := make(map[string]PricingBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingBucket builds a valid PricingBucket, for tests and for the seed data the cron
// binary writes.
func SamplePricingBucket(id string, at time.Time) PricingBucket {
	return PricingBucket{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Category: "seeded by pricing.SamplePricingBucket",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingLaneSchema is the stored shape of a PricingLane.
var PricingLaneSchema = schema.New("pricing.pricinglane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingLane is one pricing lane.
type PricingLane struct {
	ID        string
	Channel   string
	Reference string
	Comment   string
	Offset    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PricingLane) Key() string { return "pricing/pricinglane/" + x.ID }

// Fields encodes x for storage.
func (x PricingLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"comment":   x.Comment,
		"offset":    strconv.Itoa(x.Offset),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingLane decodes a stored record.
func ParsePricingLane(f codec.Fields) (PricingLane, error) {
	var x PricingLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricinglane: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("pricing: pricinglane %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricinglane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricinglane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricinglane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingLane) String() string {
	return fmt.Sprintf("PricingLane(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingLane) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingLane) Normalize() PricingLane {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPricingLanes orders by offset descending, then by ID.
func SortPricingLanes(xs []PricingLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingLanes drops the archived entries, preserving order.
func ActivePricingLanes(xs []PricingLane) []PricingLane {
	out := make([]PricingLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingLanes adds up the offset of every active entry.
func SumPricingLanes(xs []PricingLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPricingLanes keys the entries by ID, last write winning.
func IndexPricingLanes(xs []PricingLane) map[string]PricingLane {
	out := make(map[string]PricingLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingLane builds a valid PricingLane, for tests and for the seed data the cron
// binary writes.
func SamplePricingLane(id string, at time.Time) PricingLane {
	return PricingLane{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by pricing.SamplePricingLane",
		Offset:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PricingSlotSchema is the stored shape of a PricingSlot.
var PricingSlotSchema = schema.New("pricing.pricingslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingSlot is one pricing slot.
type PricingSlot struct {
	ID       string
	Category string
	Label    string
	Channel  string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingSlot) Key() string { return "pricing/pricingslot/" + x.ID }

// Fields encodes x for storage.
func (x PricingSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingSlot decodes a stored record.
func ParsePricingSlot(f codec.Fields) (PricingSlot, error) {
	var x PricingSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingslot: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingslot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingSlot) String() string {
	return fmt.Sprintf("PricingSlot(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingSlot) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingSlot) Normalize() PricingSlot {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPricingSlots orders by capacity descending, then by ID.
func SortPricingSlots(xs []PricingSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingSlots drops the archived entries, preserving order.
func ActivePricingSlots(xs []PricingSlot) []PricingSlot {
	out := make([]PricingSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingSlots adds up the capacity of every active entry.
func SumPricingSlots(xs []PricingSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPricingSlots keys the entries by ID, last write winning.
func IndexPricingSlots(xs []PricingSlot) map[string]PricingSlot {
	out := make(map[string]PricingSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingSlot builds a valid PricingSlot, for tests and for the seed data the cron
// binary writes.
func SamplePricingSlot(id string, at time.Time) PricingSlot {
	return PricingSlot{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by pricing.SamplePricingSlot",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingTagSchema is the stored shape of a PricingTag.
var PricingTagSchema = schema.New("pricing.pricingtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingTag is one pricing tag.
type PricingTag struct {
	ID       string
	Owner    string
	Category string
	Comment  string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingTag) Key() string { return "pricing/pricingtag/" + x.ID }

// Fields encodes x for storage.
func (x PricingTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"category": x.Category,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingTag decodes a stored record.
func ParsePricingTag(f codec.Fields) (PricingTag, error) {
	var x PricingTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingtag: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingtag %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingTag) String() string {
	return fmt.Sprintf("PricingTag(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingTag) Normalize() PricingTag {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPricingTags orders by count descending, then by ID.
func SortPricingTags(xs []PricingTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingTags drops the archived entries, preserving order.
func ActivePricingTags(xs []PricingTag) []PricingTag {
	out := make([]PricingTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingTags adds up the count of every active entry.
func SumPricingTags(xs []PricingTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPricingTags keys the entries by ID, last write winning.
func IndexPricingTags(xs []PricingTag) map[string]PricingTag {
	out := make(map[string]PricingTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingTag builds a valid PricingTag, for tests and for the seed data the cron
// binary writes.
func SamplePricingTag(id string, at time.Time) PricingTag {
	return PricingTag{
		ID:       id,
		Owner:    "owner-" + id,
		Category: "category-" + id,
		Comment:  "seeded by pricing.SamplePricingTag",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingRuleSchema is the stored shape of a PricingRule.
var PricingRuleSchema = schema.New("pricing.pricingrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingRule is one pricing rule.
type PricingRule struct {
	ID       string
	Comment  string
	Channel  string
	Owner    string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingRule) Key() string { return "pricing/pricingrule/" + x.ID }

// Fields encodes x for storage.
func (x PricingRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingRule decodes a stored record.
func ParsePricingRule(f codec.Fields) (PricingRule, error) {
	var x PricingRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingrule: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrule %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingRule) String() string {
	return fmt.Sprintf("PricingRule(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingRule) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingRule) Normalize() PricingRule {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPricingRules orders by attempts descending, then by ID.
func SortPricingRules(xs []PricingRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingRules drops the archived entries, preserving order.
func ActivePricingRules(xs []PricingRule) []PricingRule {
	out := make([]PricingRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingRules adds up the attempts of every active entry.
func SumPricingRules(xs []PricingRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPricingRules keys the entries by ID, last write winning.
func IndexPricingRules(xs []PricingRule) map[string]PricingRule {
	out := make(map[string]PricingRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingRule builds a valid PricingRule, for tests and for the seed data the cron
// binary writes.
func SamplePricingRule(id string, at time.Time) PricingRule {
	return PricingRule{
		ID:       id,
		Comment:  "comment-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by pricing.SamplePricingRule",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingHoldSchema is the stored shape of a PricingHold.
var PricingHoldSchema = schema.New("pricing.pricinghold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingHold is one pricing hold.
type PricingHold struct {
	ID       string
	Group    string
	Owner    string
	Region   string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingHold) Key() string { return "pricing/pricinghold/" + x.ID }

// Fields encodes x for storage.
func (x PricingHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"region":   x.Region,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingHold decodes a stored record.
func ParsePricingHold(f codec.Fields) (PricingHold, error) {
	var x PricingHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricinghold: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricinghold %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("pricing: pricinghold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricinghold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricinghold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingHold) String() string {
	return fmt.Sprintf("PricingHold(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingHold) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingHold) Normalize() PricingHold {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPricingHolds orders by count descending, then by ID.
func SortPricingHolds(xs []PricingHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingHolds drops the archived entries, preserving order.
func ActivePricingHolds(xs []PricingHold) []PricingHold {
	out := make([]PricingHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingHolds adds up the count of every active entry.
func SumPricingHolds(xs []PricingHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPricingHolds keys the entries by ID, last write winning.
func IndexPricingHolds(xs []PricingHold) map[string]PricingHold {
	out := make(map[string]PricingHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingHold builds a valid PricingHold, for tests and for the seed data the cron
// binary writes.
func SamplePricingHold(id string, at time.Time) PricingHold {
	return PricingHold{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by pricing.SamplePricingHold",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingClaimSchema is the stored shape of a PricingClaim.
var PricingClaimSchema = schema.New("pricing.pricingclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingClaim is one pricing claim.
type PricingClaim struct {
	ID       string
	Name     string
	Group    string
	Label    string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingClaim) Key() string { return "pricing/pricingclaim/" + x.ID }

// Fields encodes x for storage.
func (x PricingClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"group":    x.Group,
		"label":    x.Label,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingClaim decodes a stored record.
func ParsePricingClaim(f codec.Fields) (PricingClaim, error) {
	var x PricingClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingclaim: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingclaim %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingClaim) String() string {
	return fmt.Sprintf("PricingClaim(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingClaim) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingClaim) Normalize() PricingClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPricingClaims orders by version descending, then by ID.
func SortPricingClaims(xs []PricingClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingClaims drops the archived entries, preserving order.
func ActivePricingClaims(xs []PricingClaim) []PricingClaim {
	out := make([]PricingClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingClaims adds up the version of every active entry.
func SumPricingClaims(xs []PricingClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPricingClaims keys the entries by ID, last write winning.
func IndexPricingClaims(xs []PricingClaim) map[string]PricingClaim {
	out := make(map[string]PricingClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingClaim builds a valid PricingClaim, for tests and for the seed data the cron
// binary writes.
func SamplePricingClaim(id string, at time.Time) PricingClaim {
	return PricingClaim{
		ID:       id,
		Name:     "name-" + id,
		Group:    "group-" + id,
		Label:    "seeded by pricing.SamplePricingClaim",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PricingCursorSchema is the stored shape of a PricingCursor.
var PricingCursorSchema = schema.New("pricing.pricingcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PricingCursor is one pricing cursor.
type PricingCursor struct {
	ID       string
	Channel  string
	Label    string
	Name     string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PricingCursor) Key() string { return "pricing/pricingcursor/" + x.ID }

// Fields encodes x for storage.
func (x PricingCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePricingCursor decodes a stored record.
func ParsePricingCursor(f codec.Fields) (PricingCursor, error) {
	var x PricingCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("pricing: pricingcursor: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingcursor %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("pricing: pricingcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("pricing: pricingcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("pricing: pricingcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PricingCursor) String() string {
	return fmt.Sprintf("PricingCursor(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PricingCursor) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PricingCursor) Normalize() PricingCursor {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPricingCursors orders by count descending, then by ID.
func SortPricingCursors(xs []PricingCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePricingCursors drops the archived entries, preserving order.
func ActivePricingCursors(xs []PricingCursor) []PricingCursor {
	out := make([]PricingCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPricingCursors adds up the count of every active entry.
func SumPricingCursors(xs []PricingCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPricingCursors keys the entries by ID, last write winning.
func IndexPricingCursors(xs []PricingCursor) map[string]PricingCursor {
	out := make(map[string]PricingCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePricingCursor builds a valid PricingCursor, for tests and for the seed data the cron
// binary writes.
func SamplePricingCursor(id string, at time.Time) PricingCursor {
	return PricingCursor{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Name:     "seeded by pricing.SamplePricingCursor",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	PricingRecordSchema,
	PricingEntrySchema,
	PricingItemSchema,
	PricingDraftSchema,
	PricingRevisionSchema,
	PricingSnapshotSchema,
	PricingBatchSchema,
	PricingRequestSchema,
	PricingSummarySchema,
	PricingPolicySchema,
	PricingLinkSchema,
	PricingNoteSchema,
	PricingWindowSchema,
	PricingQuotaSchema,
	PricingBucketSchema,
	PricingLaneSchema,
	PricingSlotSchema,
	PricingTagSchema,
	PricingRuleSchema,
	PricingHoldSchema,
	PricingClaimSchema,
	PricingCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"pricing/pricingrecord/",
	"pricing/pricingentry/",
	"pricing/pricingitem/",
	"pricing/pricingdraft/",
	"pricing/pricingrevision/",
	"pricing/pricingsnapshot/",
	"pricing/pricingbatch/",
	"pricing/pricingrequest/",
	"pricing/pricingsummary/",
	"pricing/pricingpolicy/",
	"pricing/pricinglink/",
	"pricing/pricingnote/",
	"pricing/pricingwindow/",
	"pricing/pricingquota/",
	"pricing/pricingbucket/",
	"pricing/pricinglane/",
	"pricing/pricingslot/",
	"pricing/pricingtag/",
	"pricing/pricingrule/",
	"pricing/pricinghold/",
	"pricing/pricingclaim/",
	"pricing/pricingcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "pricing: " + strings.Join(names, ", ")
}
