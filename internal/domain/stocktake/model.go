// Package stocktake holds the stocktake domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package stocktake

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// StocktakeRecordSchema is the stored shape of a StocktakeRecord.
var StocktakeRecordSchema = schema.New("stocktake.stocktakerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeRecord is one stocktake record.
type StocktakeRecord struct {
	ID        string
	Source    string
	Reference string
	Channel   string
	Version   int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x StocktakeRecord) Key() string { return "stocktake/stocktakerecord/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"reference": x.Reference,
		"channel":   x.Channel,
		"version":   strconv.Itoa(x.Version),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeRecord decodes a stored record.
func ParseStocktakeRecord(f codec.Fields) (StocktakeRecord, error) {
	var x StocktakeRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakerecord: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeRecord) String() string {
	return fmt.Sprintf("StocktakeRecord(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeRecord) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeRecord) Normalize() StocktakeRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortStocktakeRecords orders by version descending, then by ID.
func SortStocktakeRecords(xs []StocktakeRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeRecords drops the archived entries, preserving order.
func ActiveStocktakeRecords(xs []StocktakeRecord) []StocktakeRecord {
	out := make([]StocktakeRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeRecords adds up the version of every active entry.
func SumStocktakeRecords(xs []StocktakeRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexStocktakeRecords keys the entries by ID, last write winning.
func IndexStocktakeRecords(xs []StocktakeRecord) map[string]StocktakeRecord {
	out := make(map[string]StocktakeRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeRecord builds a valid StocktakeRecord, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeRecord(id string, at time.Time) StocktakeRecord {
	return StocktakeRecord{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by stocktake.SampleStocktakeRecord",
		Version:   len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// StocktakeEntrySchema is the stored shape of a StocktakeEntry.
var StocktakeEntrySchema = schema.New("stocktake.stocktakeentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeEntry is one stocktake entry.
type StocktakeEntry struct {
	ID       string
	Name     string
	Category string
	Group    string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeEntry) Key() string { return "stocktake/stocktakeentry/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeEntry decodes a stored record.
func ParseStocktakeEntry(f codec.Fields) (StocktakeEntry, error) {
	var x StocktakeEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakeentry: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeentry %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakeentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeEntry) String() string {
	return fmt.Sprintf("StocktakeEntry(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeEntry) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeEntry) Normalize() StocktakeEntry {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortStocktakeEntrys orders by count descending, then by ID.
func SortStocktakeEntrys(xs []StocktakeEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeEntrys drops the archived entries, preserving order.
func ActiveStocktakeEntrys(xs []StocktakeEntry) []StocktakeEntry {
	out := make([]StocktakeEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeEntrys adds up the count of every active entry.
func SumStocktakeEntrys(xs []StocktakeEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexStocktakeEntrys keys the entries by ID, last write winning.
func IndexStocktakeEntrys(xs []StocktakeEntry) map[string]StocktakeEntry {
	out := make(map[string]StocktakeEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeEntry builds a valid StocktakeEntry, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeEntry(id string, at time.Time) StocktakeEntry {
	return StocktakeEntry{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Group:    "seeded by stocktake.SampleStocktakeEntry",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeItemSchema is the stored shape of a StocktakeItem.
var StocktakeItemSchema = schema.New("stocktake.stocktakeitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeItem is one stocktake item.
type StocktakeItem struct {
	ID       string
	Label    string
	Channel  string
	Owner    string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeItem) Key() string { return "stocktake/stocktakeitem/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeItem decodes a stored record.
func ParseStocktakeItem(f codec.Fields) (StocktakeItem, error) {
	var x StocktakeItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakeitem: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeitem %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakeitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeItem) String() string {
	return fmt.Sprintf("StocktakeItem(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeItem) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeItem) Normalize() StocktakeItem {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortStocktakeItems orders by retained descending, then by ID.
func SortStocktakeItems(xs []StocktakeItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeItems drops the archived entries, preserving order.
func ActiveStocktakeItems(xs []StocktakeItem) []StocktakeItem {
	out := make([]StocktakeItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeItems adds up the retained of every active entry.
func SumStocktakeItems(xs []StocktakeItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexStocktakeItems keys the entries by ID, last write winning.
func IndexStocktakeItems(xs []StocktakeItem) map[string]StocktakeItem {
	out := make(map[string]StocktakeItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeItem builds a valid StocktakeItem, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeItem(id string, at time.Time) StocktakeItem {
	return StocktakeItem{
		ID:       id,
		Label:    "label-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by stocktake.SampleStocktakeItem",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeDraftSchema is the stored shape of a StocktakeDraft.
var StocktakeDraftSchema = schema.New("stocktake.stocktakedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeDraft is one stocktake draft.
type StocktakeDraft struct {
	ID       string
	Owner    string
	Group    string
	Channel  string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeDraft) Key() string { return "stocktake/stocktakedraft/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"channel":  x.Channel,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeDraft decodes a stored record.
func ParseStocktakeDraft(f codec.Fields) (StocktakeDraft, error) {
	var x StocktakeDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakedraft: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakedraft %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeDraft) String() string {
	return fmt.Sprintf("StocktakeDraft(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeDraft) Normalize() StocktakeDraft {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortStocktakeDrafts orders by attempts descending, then by ID.
func SortStocktakeDrafts(xs []StocktakeDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeDrafts drops the archived entries, preserving order.
func ActiveStocktakeDrafts(xs []StocktakeDraft) []StocktakeDraft {
	out := make([]StocktakeDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeDrafts adds up the attempts of every active entry.
func SumStocktakeDrafts(xs []StocktakeDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexStocktakeDrafts keys the entries by ID, last write winning.
func IndexStocktakeDrafts(xs []StocktakeDraft) map[string]StocktakeDraft {
	out := make(map[string]StocktakeDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeDraft builds a valid StocktakeDraft, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeDraft(id string, at time.Time) StocktakeDraft {
	return StocktakeDraft{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Channel:  "seeded by stocktake.SampleStocktakeDraft",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeRevisionSchema is the stored shape of a StocktakeRevision.
var StocktakeRevisionSchema = schema.New("stocktake.stocktakerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeRevision is one stocktake revision.
type StocktakeRevision struct {
	ID        string
	Group     string
	Label     string
	Reference string
	Priority  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x StocktakeRevision) Key() string { return "stocktake/stocktakerevision/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"label":     x.Label,
		"reference": x.Reference,
		"priority":  strconv.Itoa(x.Priority),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeRevision decodes a stored record.
func ParseStocktakeRevision(f codec.Fields) (StocktakeRevision, error) {
	var x StocktakeRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakerevision: missing id")
	}
	x.Group = f["group"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerevision %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeRevision) String() string {
	return fmt.Sprintf("StocktakeRevision(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeRevision) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeRevision) Normalize() StocktakeRevision {
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortStocktakeRevisions orders by priority descending, then by ID.
func SortStocktakeRevisions(xs []StocktakeRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeRevisions drops the archived entries, preserving order.
func ActiveStocktakeRevisions(xs []StocktakeRevision) []StocktakeRevision {
	out := make([]StocktakeRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeRevisions adds up the priority of every active entry.
func SumStocktakeRevisions(xs []StocktakeRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexStocktakeRevisions keys the entries by ID, last write winning.
func IndexStocktakeRevisions(xs []StocktakeRevision) map[string]StocktakeRevision {
	out := make(map[string]StocktakeRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeRevision builds a valid StocktakeRevision, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeRevision(id string, at time.Time) StocktakeRevision {
	return StocktakeRevision{
		ID:        id,
		Group:     "group-" + id,
		Label:     "label-" + id,
		Reference: "seeded by stocktake.SampleStocktakeRevision",
		Priority:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// StocktakeSnapshotSchema is the stored shape of a StocktakeSnapshot.
var StocktakeSnapshotSchema = schema.New("stocktake.stocktakesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeSnapshot is one stocktake snapshot.
type StocktakeSnapshot struct {
	ID       string
	Channel  string
	Category string
	Source   string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeSnapshot) Key() string { return "stocktake/stocktakesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeSnapshot decodes a stored record.
func ParseStocktakeSnapshot(f codec.Fields) (StocktakeSnapshot, error) {
	var x StocktakeSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakesnapshot: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakesnapshot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeSnapshot) String() string {
	return fmt.Sprintf("StocktakeSnapshot(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeSnapshot) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeSnapshot) Normalize() StocktakeSnapshot {
	x.Channel = strings.TrimSpace(x.Channel)
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

// SortStocktakeSnapshots orders by capacity descending, then by ID.
func SortStocktakeSnapshots(xs []StocktakeSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeSnapshots drops the archived entries, preserving order.
func ActiveStocktakeSnapshots(xs []StocktakeSnapshot) []StocktakeSnapshot {
	out := make([]StocktakeSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeSnapshots adds up the capacity of every active entry.
func SumStocktakeSnapshots(xs []StocktakeSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexStocktakeSnapshots keys the entries by ID, last write winning.
func IndexStocktakeSnapshots(xs []StocktakeSnapshot) map[string]StocktakeSnapshot {
	out := make(map[string]StocktakeSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeSnapshot builds a valid StocktakeSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeSnapshot(id string, at time.Time) StocktakeSnapshot {
	return StocktakeSnapshot{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Source:   "seeded by stocktake.SampleStocktakeSnapshot",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeBatchSchema is the stored shape of a StocktakeBatch.
var StocktakeBatchSchema = schema.New("stocktake.stocktakebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeBatch is one stocktake batch.
type StocktakeBatch struct {
	ID       string
	Source   string
	Channel  string
	Label    string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeBatch) Key() string { return "stocktake/stocktakebatch/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeBatch decodes a stored record.
func ParseStocktakeBatch(f codec.Fields) (StocktakeBatch, error) {
	var x StocktakeBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakebatch: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakebatch %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeBatch) String() string {
	return fmt.Sprintf("StocktakeBatch(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeBatch) Normalize() StocktakeBatch {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortStocktakeBatchs orders by capacity descending, then by ID.
func SortStocktakeBatchs(xs []StocktakeBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeBatchs drops the archived entries, preserving order.
func ActiveStocktakeBatchs(xs []StocktakeBatch) []StocktakeBatch {
	out := make([]StocktakeBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeBatchs adds up the capacity of every active entry.
func SumStocktakeBatchs(xs []StocktakeBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexStocktakeBatchs keys the entries by ID, last write winning.
func IndexStocktakeBatchs(xs []StocktakeBatch) map[string]StocktakeBatch {
	out := make(map[string]StocktakeBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeBatch builds a valid StocktakeBatch, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeBatch(id string, at time.Time) StocktakeBatch {
	return StocktakeBatch{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by stocktake.SampleStocktakeBatch",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeRequestSchema is the stored shape of a StocktakeRequest.
var StocktakeRequestSchema = schema.New("stocktake.stocktakerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeRequest is one stocktake request.
type StocktakeRequest struct {
	ID       string
	Label    string
	Region   string
	Source   string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeRequest) Key() string { return "stocktake/stocktakerequest/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeRequest decodes a stored record.
func ParseStocktakeRequest(f codec.Fields) (StocktakeRequest, error) {
	var x StocktakeRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakerequest: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerequest %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeRequest) String() string {
	return fmt.Sprintf("StocktakeRequest(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeRequest) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeRequest) Normalize() StocktakeRequest {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortStocktakeRequests orders by attempts descending, then by ID.
func SortStocktakeRequests(xs []StocktakeRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeRequests drops the archived entries, preserving order.
func ActiveStocktakeRequests(xs []StocktakeRequest) []StocktakeRequest {
	out := make([]StocktakeRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeRequests adds up the attempts of every active entry.
func SumStocktakeRequests(xs []StocktakeRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexStocktakeRequests keys the entries by ID, last write winning.
func IndexStocktakeRequests(xs []StocktakeRequest) map[string]StocktakeRequest {
	out := make(map[string]StocktakeRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeRequest builds a valid StocktakeRequest, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeRequest(id string, at time.Time) StocktakeRequest {
	return StocktakeRequest{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Source:   "seeded by stocktake.SampleStocktakeRequest",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeSummarySchema is the stored shape of a StocktakeSummary.
var StocktakeSummarySchema = schema.New("stocktake.stocktakesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeSummary is one stocktake summary.
type StocktakeSummary struct {
	ID        string
	Reference string
	Comment   string
	Group     string
	Priority  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x StocktakeSummary) Key() string { return "stocktake/stocktakesummary/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"comment":   x.Comment,
		"group":     x.Group,
		"priority":  strconv.Itoa(x.Priority),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeSummary decodes a stored record.
func ParseStocktakeSummary(f codec.Fields) (StocktakeSummary, error) {
	var x StocktakeSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakesummary: missing id")
	}
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakesummary %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeSummary) String() string {
	return fmt.Sprintf("StocktakeSummary(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeSummary) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeSummary) Normalize() StocktakeSummary {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortStocktakeSummarys orders by priority descending, then by ID.
func SortStocktakeSummarys(xs []StocktakeSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeSummarys drops the archived entries, preserving order.
func ActiveStocktakeSummarys(xs []StocktakeSummary) []StocktakeSummary {
	out := make([]StocktakeSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeSummarys adds up the priority of every active entry.
func SumStocktakeSummarys(xs []StocktakeSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexStocktakeSummarys keys the entries by ID, last write winning.
func IndexStocktakeSummarys(xs []StocktakeSummary) map[string]StocktakeSummary {
	out := make(map[string]StocktakeSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeSummary builds a valid StocktakeSummary, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeSummary(id string, at time.Time) StocktakeSummary {
	return StocktakeSummary{
		ID:        id,
		Reference: "reference-" + id,
		Comment:   "comment-" + id,
		Group:     "seeded by stocktake.SampleStocktakeSummary",
		Priority:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// StocktakePolicySchema is the stored shape of a StocktakePolicy.
var StocktakePolicySchema = schema.New("stocktake.stocktakepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakePolicy is one stocktake policy.
type StocktakePolicy struct {
	ID       string
	Region   string
	Label    string
	Name     string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakePolicy) Key() string { return "stocktake/stocktakepolicy/" + x.ID }

// Fields encodes x for storage.
func (x StocktakePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakePolicy decodes a stored record.
func ParseStocktakePolicy(f codec.Fields) (StocktakePolicy, error) {
	var x StocktakePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakepolicy: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakepolicy %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakePolicy) String() string {
	return fmt.Sprintf("StocktakePolicy(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakePolicy) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakePolicy) Normalize() StocktakePolicy {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortStocktakePolicys orders by count descending, then by ID.
func SortStocktakePolicys(xs []StocktakePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakePolicys drops the archived entries, preserving order.
func ActiveStocktakePolicys(xs []StocktakePolicy) []StocktakePolicy {
	out := make([]StocktakePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakePolicys adds up the count of every active entry.
func SumStocktakePolicys(xs []StocktakePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexStocktakePolicys keys the entries by ID, last write winning.
func IndexStocktakePolicys(xs []StocktakePolicy) map[string]StocktakePolicy {
	out := make(map[string]StocktakePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakePolicy builds a valid StocktakePolicy, for tests and for the seed data the cron
// binary writes.
func SampleStocktakePolicy(id string, at time.Time) StocktakePolicy {
	return StocktakePolicy{
		ID:      id,
		Region:  "region-" + id,
		Label:   "label-" + id,
		Name:    "seeded by stocktake.SampleStocktakePolicy",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// StocktakeLinkSchema is the stored shape of a StocktakeLink.
var StocktakeLinkSchema = schema.New("stocktake.stocktakelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeLink is one stocktake link.
type StocktakeLink struct {
	ID       string
	Group    string
	Category string
	Region   string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeLink) Key() string { return "stocktake/stocktakelink/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeLink decodes a stored record.
func ParseStocktakeLink(f codec.Fields) (StocktakeLink, error) {
	var x StocktakeLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakelink: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakelink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeLink) String() string {
	return fmt.Sprintf("StocktakeLink(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeLink) Normalize() StocktakeLink {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortStocktakeLinks orders by retained descending, then by ID.
func SortStocktakeLinks(xs []StocktakeLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeLinks drops the archived entries, preserving order.
func ActiveStocktakeLinks(xs []StocktakeLink) []StocktakeLink {
	out := make([]StocktakeLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeLinks adds up the retained of every active entry.
func SumStocktakeLinks(xs []StocktakeLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexStocktakeLinks keys the entries by ID, last write winning.
func IndexStocktakeLinks(xs []StocktakeLink) map[string]StocktakeLink {
	out := make(map[string]StocktakeLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeLink builds a valid StocktakeLink, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeLink(id string, at time.Time) StocktakeLink {
	return StocktakeLink{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Region:   "seeded by stocktake.SampleStocktakeLink",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeNoteSchema is the stored shape of a StocktakeNote.
var StocktakeNoteSchema = schema.New("stocktake.stocktakenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeNote is one stocktake note.
type StocktakeNote struct {
	ID       string
	Source   string
	Comment  string
	Group    string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeNote) Key() string { return "stocktake/stocktakenote/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeNote decodes a stored record.
func ParseStocktakeNote(f codec.Fields) (StocktakeNote, error) {
	var x StocktakeNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakenote: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakenote %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeNote) String() string {
	return fmt.Sprintf("StocktakeNote(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeNote) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeNote) Normalize() StocktakeNote {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortStocktakeNotes orders by version descending, then by ID.
func SortStocktakeNotes(xs []StocktakeNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeNotes drops the archived entries, preserving order.
func ActiveStocktakeNotes(xs []StocktakeNote) []StocktakeNote {
	out := make([]StocktakeNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeNotes adds up the version of every active entry.
func SumStocktakeNotes(xs []StocktakeNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexStocktakeNotes keys the entries by ID, last write winning.
func IndexStocktakeNotes(xs []StocktakeNote) map[string]StocktakeNote {
	out := make(map[string]StocktakeNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeNote builds a valid StocktakeNote, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeNote(id string, at time.Time) StocktakeNote {
	return StocktakeNote{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by stocktake.SampleStocktakeNote",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeWindowSchema is the stored shape of a StocktakeWindow.
var StocktakeWindowSchema = schema.New("stocktake.stocktakewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeWindow is one stocktake window.
type StocktakeWindow struct {
	ID       string
	Channel  string
	Region   string
	Category string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeWindow) Key() string { return "stocktake/stocktakewindow/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeWindow decodes a stored record.
func ParseStocktakeWindow(f codec.Fields) (StocktakeWindow, error) {
	var x StocktakeWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakewindow: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakewindow %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeWindow) String() string {
	return fmt.Sprintf("StocktakeWindow(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeWindow) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeWindow) Normalize() StocktakeWindow {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortStocktakeWindows orders by capacity descending, then by ID.
func SortStocktakeWindows(xs []StocktakeWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeWindows drops the archived entries, preserving order.
func ActiveStocktakeWindows(xs []StocktakeWindow) []StocktakeWindow {
	out := make([]StocktakeWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeWindows adds up the capacity of every active entry.
func SumStocktakeWindows(xs []StocktakeWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexStocktakeWindows keys the entries by ID, last write winning.
func IndexStocktakeWindows(xs []StocktakeWindow) map[string]StocktakeWindow {
	out := make(map[string]StocktakeWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeWindow builds a valid StocktakeWindow, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeWindow(id string, at time.Time) StocktakeWindow {
	return StocktakeWindow{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Category: "seeded by stocktake.SampleStocktakeWindow",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeQuotaSchema is the stored shape of a StocktakeQuota.
var StocktakeQuotaSchema = schema.New("stocktake.stocktakequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeQuota is one stocktake quota.
type StocktakeQuota struct {
	ID       string
	Name     string
	Label    string
	Region   string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeQuota) Key() string { return "stocktake/stocktakequota/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeQuota decodes a stored record.
func ParseStocktakeQuota(f codec.Fields) (StocktakeQuota, error) {
	var x StocktakeQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakequota: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakequota %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeQuota) String() string {
	return fmt.Sprintf("StocktakeQuota(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeQuota) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeQuota) Normalize() StocktakeQuota {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortStocktakeQuotas orders by weight descending, then by ID.
func SortStocktakeQuotas(xs []StocktakeQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeQuotas drops the archived entries, preserving order.
func ActiveStocktakeQuotas(xs []StocktakeQuota) []StocktakeQuota {
	out := make([]StocktakeQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeQuotas adds up the weight of every active entry.
func SumStocktakeQuotas(xs []StocktakeQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexStocktakeQuotas keys the entries by ID, last write winning.
func IndexStocktakeQuotas(xs []StocktakeQuota) map[string]StocktakeQuota {
	out := make(map[string]StocktakeQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeQuota builds a valid StocktakeQuota, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeQuota(id string, at time.Time) StocktakeQuota {
	return StocktakeQuota{
		ID:       id,
		Name:     "name-" + id,
		Label:    "label-" + id,
		Region:   "seeded by stocktake.SampleStocktakeQuota",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeBucketSchema is the stored shape of a StocktakeBucket.
var StocktakeBucketSchema = schema.New("stocktake.stocktakebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeBucket is one stocktake bucket.
type StocktakeBucket struct {
	ID       string
	Region   string
	Owner    string
	Label    string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeBucket) Key() string { return "stocktake/stocktakebucket/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"owner":    x.Owner,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeBucket decodes a stored record.
func ParseStocktakeBucket(f codec.Fields) (StocktakeBucket, error) {
	var x StocktakeBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakebucket: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakebucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeBucket) String() string {
	return fmt.Sprintf("StocktakeBucket(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeBucket) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeBucket) Normalize() StocktakeBucket {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortStocktakeBuckets orders by priority descending, then by ID.
func SortStocktakeBuckets(xs []StocktakeBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeBuckets drops the archived entries, preserving order.
func ActiveStocktakeBuckets(xs []StocktakeBucket) []StocktakeBucket {
	out := make([]StocktakeBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeBuckets adds up the priority of every active entry.
func SumStocktakeBuckets(xs []StocktakeBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexStocktakeBuckets keys the entries by ID, last write winning.
func IndexStocktakeBuckets(xs []StocktakeBucket) map[string]StocktakeBucket {
	out := make(map[string]StocktakeBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeBucket builds a valid StocktakeBucket, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeBucket(id string, at time.Time) StocktakeBucket {
	return StocktakeBucket{
		ID:       id,
		Region:   "region-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by stocktake.SampleStocktakeBucket",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeLaneSchema is the stored shape of a StocktakeLane.
var StocktakeLaneSchema = schema.New("stocktake.stocktakelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeLane is one stocktake lane.
type StocktakeLane struct {
	ID       string
	Category string
	Owner    string
	Channel  string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeLane) Key() string { return "stocktake/stocktakelane/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeLane decodes a stored record.
func ParseStocktakeLane(f codec.Fields) (StocktakeLane, error) {
	var x StocktakeLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakelane: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakelane %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeLane) String() string {
	return fmt.Sprintf("StocktakeLane(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeLane) Normalize() StocktakeLane {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortStocktakeLanes orders by capacity descending, then by ID.
func SortStocktakeLanes(xs []StocktakeLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeLanes drops the archived entries, preserving order.
func ActiveStocktakeLanes(xs []StocktakeLane) []StocktakeLane {
	out := make([]StocktakeLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeLanes adds up the capacity of every active entry.
func SumStocktakeLanes(xs []StocktakeLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexStocktakeLanes keys the entries by ID, last write winning.
func IndexStocktakeLanes(xs []StocktakeLane) map[string]StocktakeLane {
	out := make(map[string]StocktakeLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeLane builds a valid StocktakeLane, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeLane(id string, at time.Time) StocktakeLane {
	return StocktakeLane{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by stocktake.SampleStocktakeLane",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeSlotSchema is the stored shape of a StocktakeSlot.
var StocktakeSlotSchema = schema.New("stocktake.stocktakeslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeSlot is one stocktake slot.
type StocktakeSlot struct {
	ID        string
	Reference string
	Source    string
	Channel   string
	Capacity  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x StocktakeSlot) Key() string { return "stocktake/stocktakeslot/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"source":    x.Source,
		"channel":   x.Channel,
		"capacity":  strconv.Itoa(x.Capacity),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeSlot decodes a stored record.
func ParseStocktakeSlot(f codec.Fields) (StocktakeSlot, error) {
	var x StocktakeSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakeslot: missing id")
	}
	x.Reference = f["reference"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeslot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakeslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeSlot) String() string {
	return fmt.Sprintf("StocktakeSlot(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeSlot) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeSlot) Normalize() StocktakeSlot {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortStocktakeSlots orders by capacity descending, then by ID.
func SortStocktakeSlots(xs []StocktakeSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeSlots drops the archived entries, preserving order.
func ActiveStocktakeSlots(xs []StocktakeSlot) []StocktakeSlot {
	out := make([]StocktakeSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeSlots adds up the capacity of every active entry.
func SumStocktakeSlots(xs []StocktakeSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexStocktakeSlots keys the entries by ID, last write winning.
func IndexStocktakeSlots(xs []StocktakeSlot) map[string]StocktakeSlot {
	out := make(map[string]StocktakeSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeSlot builds a valid StocktakeSlot, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeSlot(id string, at time.Time) StocktakeSlot {
	return StocktakeSlot{
		ID:        id,
		Reference: "reference-" + id,
		Source:    "source-" + id,
		Channel:   "seeded by stocktake.SampleStocktakeSlot",
		Capacity:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// StocktakeTagSchema is the stored shape of a StocktakeTag.
var StocktakeTagSchema = schema.New("stocktake.stocktaketag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeTag is one stocktake tag.
type StocktakeTag struct {
	ID       string
	Channel  string
	Comment  string
	Category string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeTag) Key() string { return "stocktake/stocktaketag/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeTag decodes a stored record.
func ParseStocktakeTag(f codec.Fields) (StocktakeTag, error) {
	var x StocktakeTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktaketag: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktaketag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktaketag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktaketag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktaketag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeTag) String() string {
	return fmt.Sprintf("StocktakeTag(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeTag) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeTag) Normalize() StocktakeTag {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortStocktakeTags orders by capacity descending, then by ID.
func SortStocktakeTags(xs []StocktakeTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeTags drops the archived entries, preserving order.
func ActiveStocktakeTags(xs []StocktakeTag) []StocktakeTag {
	out := make([]StocktakeTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeTags adds up the capacity of every active entry.
func SumStocktakeTags(xs []StocktakeTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexStocktakeTags keys the entries by ID, last write winning.
func IndexStocktakeTags(xs []StocktakeTag) map[string]StocktakeTag {
	out := make(map[string]StocktakeTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeTag builds a valid StocktakeTag, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeTag(id string, at time.Time) StocktakeTag {
	return StocktakeTag{
		ID:       id,
		Channel:  "channel-" + id,
		Comment:  "comment-" + id,
		Category: "seeded by stocktake.SampleStocktakeTag",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeRuleSchema is the stored shape of a StocktakeRule.
var StocktakeRuleSchema = schema.New("stocktake.stocktakerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeRule is one stocktake rule.
type StocktakeRule struct {
	ID       string
	Region   string
	Category string
	Comment  string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeRule) Key() string { return "stocktake/stocktakerule/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"comment":  x.Comment,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeRule decodes a stored record.
func ParseStocktakeRule(f codec.Fields) (StocktakeRule, error) {
	var x StocktakeRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakerule: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerule %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeRule) String() string {
	return fmt.Sprintf("StocktakeRule(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeRule) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeRule) Normalize() StocktakeRule {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortStocktakeRules orders by attempts descending, then by ID.
func SortStocktakeRules(xs []StocktakeRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeRules drops the archived entries, preserving order.
func ActiveStocktakeRules(xs []StocktakeRule) []StocktakeRule {
	out := make([]StocktakeRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeRules adds up the attempts of every active entry.
func SumStocktakeRules(xs []StocktakeRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexStocktakeRules keys the entries by ID, last write winning.
func IndexStocktakeRules(xs []StocktakeRule) map[string]StocktakeRule {
	out := make(map[string]StocktakeRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeRule builds a valid StocktakeRule, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeRule(id string, at time.Time) StocktakeRule {
	return StocktakeRule{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Comment:  "seeded by stocktake.SampleStocktakeRule",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeHoldSchema is the stored shape of a StocktakeHold.
var StocktakeHoldSchema = schema.New("stocktake.stocktakehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeHold is one stocktake hold.
type StocktakeHold struct {
	ID       string
	Label    string
	Group    string
	Region   string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeHold) Key() string { return "stocktake/stocktakehold/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"region":   x.Region,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeHold decodes a stored record.
func ParseStocktakeHold(f codec.Fields) (StocktakeHold, error) {
	var x StocktakeHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakehold: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Region = f["region"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakehold %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeHold) String() string {
	return fmt.Sprintf("StocktakeHold(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeHold) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeHold) Normalize() StocktakeHold {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortStocktakeHolds orders by attempts descending, then by ID.
func SortStocktakeHolds(xs []StocktakeHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeHolds drops the archived entries, preserving order.
func ActiveStocktakeHolds(xs []StocktakeHold) []StocktakeHold {
	out := make([]StocktakeHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeHolds adds up the attempts of every active entry.
func SumStocktakeHolds(xs []StocktakeHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexStocktakeHolds keys the entries by ID, last write winning.
func IndexStocktakeHolds(xs []StocktakeHold) map[string]StocktakeHold {
	out := make(map[string]StocktakeHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeHold builds a valid StocktakeHold, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeHold(id string, at time.Time) StocktakeHold {
	return StocktakeHold{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Region:   "seeded by stocktake.SampleStocktakeHold",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeClaimSchema is the stored shape of a StocktakeClaim.
var StocktakeClaimSchema = schema.New("stocktake.stocktakeclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeClaim is one stocktake claim.
type StocktakeClaim struct {
	ID       string
	Owner    string
	Source   string
	Label    string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeClaim) Key() string { return "stocktake/stocktakeclaim/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeClaim decodes a stored record.
func ParseStocktakeClaim(f codec.Fields) (StocktakeClaim, error) {
	var x StocktakeClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakeclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeclaim %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakeclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakeclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeClaim) String() string {
	return fmt.Sprintf("StocktakeClaim(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeClaim) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeClaim) Normalize() StocktakeClaim {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortStocktakeClaims orders by attempts descending, then by ID.
func SortStocktakeClaims(xs []StocktakeClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeClaims drops the archived entries, preserving order.
func ActiveStocktakeClaims(xs []StocktakeClaim) []StocktakeClaim {
	out := make([]StocktakeClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeClaims adds up the attempts of every active entry.
func SumStocktakeClaims(xs []StocktakeClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexStocktakeClaims keys the entries by ID, last write winning.
func IndexStocktakeClaims(xs []StocktakeClaim) map[string]StocktakeClaim {
	out := make(map[string]StocktakeClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeClaim builds a valid StocktakeClaim, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeClaim(id string, at time.Time) StocktakeClaim {
	return StocktakeClaim{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Label:    "seeded by stocktake.SampleStocktakeClaim",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// StocktakeCursorSchema is the stored shape of a StocktakeCursor.
var StocktakeCursorSchema = schema.New("stocktake.stocktakecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// StocktakeCursor is one stocktake cursor.
type StocktakeCursor struct {
	ID       string
	Group    string
	Name     string
	Comment  string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x StocktakeCursor) Key() string { return "stocktake/stocktakecursor/" + x.ID }

// Fields encodes x for storage.
func (x StocktakeCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseStocktakeCursor decodes a stored record.
func ParseStocktakeCursor(f codec.Fields) (StocktakeCursor, error) {
	var x StocktakeCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("stocktake: stocktakecursor: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakecursor %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("stocktake: stocktakecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("stocktake: stocktakecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("stocktake: stocktakecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x StocktakeCursor) String() string {
	return fmt.Sprintf("StocktakeCursor(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x StocktakeCursor) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x StocktakeCursor) Normalize() StocktakeCursor {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortStocktakeCursors orders by retained descending, then by ID.
func SortStocktakeCursors(xs []StocktakeCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveStocktakeCursors drops the archived entries, preserving order.
func ActiveStocktakeCursors(xs []StocktakeCursor) []StocktakeCursor {
	out := make([]StocktakeCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumStocktakeCursors adds up the retained of every active entry.
func SumStocktakeCursors(xs []StocktakeCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexStocktakeCursors keys the entries by ID, last write winning.
func IndexStocktakeCursors(xs []StocktakeCursor) map[string]StocktakeCursor {
	out := make(map[string]StocktakeCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleStocktakeCursor builds a valid StocktakeCursor, for tests and for the seed data the cron
// binary writes.
func SampleStocktakeCursor(id string, at time.Time) StocktakeCursor {
	return StocktakeCursor{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by stocktake.SampleStocktakeCursor",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	StocktakeRecordSchema,
	StocktakeEntrySchema,
	StocktakeItemSchema,
	StocktakeDraftSchema,
	StocktakeRevisionSchema,
	StocktakeSnapshotSchema,
	StocktakeBatchSchema,
	StocktakeRequestSchema,
	StocktakeSummarySchema,
	StocktakePolicySchema,
	StocktakeLinkSchema,
	StocktakeNoteSchema,
	StocktakeWindowSchema,
	StocktakeQuotaSchema,
	StocktakeBucketSchema,
	StocktakeLaneSchema,
	StocktakeSlotSchema,
	StocktakeTagSchema,
	StocktakeRuleSchema,
	StocktakeHoldSchema,
	StocktakeClaimSchema,
	StocktakeCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"stocktake/stocktakerecord/",
	"stocktake/stocktakeentry/",
	"stocktake/stocktakeitem/",
	"stocktake/stocktakedraft/",
	"stocktake/stocktakerevision/",
	"stocktake/stocktakesnapshot/",
	"stocktake/stocktakebatch/",
	"stocktake/stocktakerequest/",
	"stocktake/stocktakesummary/",
	"stocktake/stocktakepolicy/",
	"stocktake/stocktakelink/",
	"stocktake/stocktakenote/",
	"stocktake/stocktakewindow/",
	"stocktake/stocktakequota/",
	"stocktake/stocktakebucket/",
	"stocktake/stocktakelane/",
	"stocktake/stocktakeslot/",
	"stocktake/stocktaketag/",
	"stocktake/stocktakerule/",
	"stocktake/stocktakehold/",
	"stocktake/stocktakeclaim/",
	"stocktake/stocktakecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "stocktake: " + strings.Join(names, ", ")
}
