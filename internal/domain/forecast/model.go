// Package forecast holds the forecast domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package forecast

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// ForecastRecordSchema is the stored shape of a ForecastRecord.
var ForecastRecordSchema = schema.New("forecast.forecastrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastRecord is one forecast record.
type ForecastRecord struct {
	ID        string
	Category  string
	Owner     string
	Reference string
	Offset    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ForecastRecord) Key() string { return "forecast/forecastrecord/" + x.ID }

// Fields encodes x for storage.
func (x ForecastRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"owner":     x.Owner,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastRecord decodes a stored record.
func ParseForecastRecord(f codec.Fields) (ForecastRecord, error) {
	var x ForecastRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastrecord: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastRecord) String() string {
	return fmt.Sprintf("ForecastRecord(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastRecord) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastRecord) Normalize() ForecastRecord {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortForecastRecords orders by offset descending, then by ID.
func SortForecastRecords(xs []ForecastRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastRecords drops the archived entries, preserving order.
func ActiveForecastRecords(xs []ForecastRecord) []ForecastRecord {
	out := make([]ForecastRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastRecords adds up the offset of every active entry.
func SumForecastRecords(xs []ForecastRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexForecastRecords keys the entries by ID, last write winning.
func IndexForecastRecords(xs []ForecastRecord) map[string]ForecastRecord {
	out := make(map[string]ForecastRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastRecord builds a valid ForecastRecord, for tests and for the seed data the cron
// binary writes.
func SampleForecastRecord(id string, at time.Time) ForecastRecord {
	return ForecastRecord{
		ID:        id,
		Category:  "category-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by forecast.SampleForecastRecord",
		Offset:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ForecastEntrySchema is the stored shape of a ForecastEntry.
var ForecastEntrySchema = schema.New("forecast.forecastentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastEntry is one forecast entry.
type ForecastEntry struct {
	ID        string
	Category  string
	Reference string
	Region    string
	Capacity  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ForecastEntry) Key() string { return "forecast/forecastentry/" + x.ID }

// Fields encodes x for storage.
func (x ForecastEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"region":    x.Region,
		"capacity":  strconv.Itoa(x.Capacity),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastEntry decodes a stored record.
func ParseForecastEntry(f codec.Fields) (ForecastEntry, error) {
	var x ForecastEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastentry: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastEntry) String() string {
	return fmt.Sprintf("ForecastEntry(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastEntry) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastEntry) Normalize() ForecastEntry {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortForecastEntrys orders by capacity descending, then by ID.
func SortForecastEntrys(xs []ForecastEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastEntrys drops the archived entries, preserving order.
func ActiveForecastEntrys(xs []ForecastEntry) []ForecastEntry {
	out := make([]ForecastEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastEntrys adds up the capacity of every active entry.
func SumForecastEntrys(xs []ForecastEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexForecastEntrys keys the entries by ID, last write winning.
func IndexForecastEntrys(xs []ForecastEntry) map[string]ForecastEntry {
	out := make(map[string]ForecastEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastEntry builds a valid ForecastEntry, for tests and for the seed data the cron
// binary writes.
func SampleForecastEntry(id string, at time.Time) ForecastEntry {
	return ForecastEntry{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by forecast.SampleForecastEntry",
		Capacity:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ForecastItemSchema is the stored shape of a ForecastItem.
var ForecastItemSchema = schema.New("forecast.forecastitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastItem is one forecast item.
type ForecastItem struct {
	ID       string
	Name     string
	Category string
	Label    string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastItem) Key() string { return "forecast/forecastitem/" + x.ID }

// Fields encodes x for storage.
func (x ForecastItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastItem decodes a stored record.
func ParseForecastItem(f codec.Fields) (ForecastItem, error) {
	var x ForecastItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastitem: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastitem %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastItem) String() string {
	return fmt.Sprintf("ForecastItem(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastItem) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastItem) Normalize() ForecastItem {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortForecastItems orders by weight descending, then by ID.
func SortForecastItems(xs []ForecastItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastItems drops the archived entries, preserving order.
func ActiveForecastItems(xs []ForecastItem) []ForecastItem {
	out := make([]ForecastItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastItems adds up the weight of every active entry.
func SumForecastItems(xs []ForecastItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexForecastItems keys the entries by ID, last write winning.
func IndexForecastItems(xs []ForecastItem) map[string]ForecastItem {
	out := make(map[string]ForecastItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastItem builds a valid ForecastItem, for tests and for the seed data the cron
// binary writes.
func SampleForecastItem(id string, at time.Time) ForecastItem {
	return ForecastItem{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Label:    "seeded by forecast.SampleForecastItem",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastDraftSchema is the stored shape of a ForecastDraft.
var ForecastDraftSchema = schema.New("forecast.forecastdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastDraft is one forecast draft.
type ForecastDraft struct {
	ID       string
	Group    string
	Category string
	Region   string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastDraft) Key() string { return "forecast/forecastdraft/" + x.ID }

// Fields encodes x for storage.
func (x ForecastDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastDraft decodes a stored record.
func ParseForecastDraft(f codec.Fields) (ForecastDraft, error) {
	var x ForecastDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastdraft: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastdraft %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastDraft) String() string {
	return fmt.Sprintf("ForecastDraft(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastDraft) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastDraft) Normalize() ForecastDraft {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortForecastDrafts orders by retained descending, then by ID.
func SortForecastDrafts(xs []ForecastDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastDrafts drops the archived entries, preserving order.
func ActiveForecastDrafts(xs []ForecastDraft) []ForecastDraft {
	out := make([]ForecastDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastDrafts adds up the retained of every active entry.
func SumForecastDrafts(xs []ForecastDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexForecastDrafts keys the entries by ID, last write winning.
func IndexForecastDrafts(xs []ForecastDraft) map[string]ForecastDraft {
	out := make(map[string]ForecastDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastDraft builds a valid ForecastDraft, for tests and for the seed data the cron
// binary writes.
func SampleForecastDraft(id string, at time.Time) ForecastDraft {
	return ForecastDraft{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Region:   "seeded by forecast.SampleForecastDraft",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastRevisionSchema is the stored shape of a ForecastRevision.
var ForecastRevisionSchema = schema.New("forecast.forecastrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastRevision is one forecast revision.
type ForecastRevision struct {
	ID       string
	Region   string
	Category string
	Channel  string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastRevision) Key() string { return "forecast/forecastrevision/" + x.ID }

// Fields encodes x for storage.
func (x ForecastRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"channel":  x.Channel,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastRevision decodes a stored record.
func ParseForecastRevision(f codec.Fields) (ForecastRevision, error) {
	var x ForecastRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastrevision: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrevision %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastRevision) String() string {
	return fmt.Sprintf("ForecastRevision(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastRevision) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastRevision) Normalize() ForecastRevision {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortForecastRevisions orders by attempts descending, then by ID.
func SortForecastRevisions(xs []ForecastRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastRevisions drops the archived entries, preserving order.
func ActiveForecastRevisions(xs []ForecastRevision) []ForecastRevision {
	out := make([]ForecastRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastRevisions adds up the attempts of every active entry.
func SumForecastRevisions(xs []ForecastRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexForecastRevisions keys the entries by ID, last write winning.
func IndexForecastRevisions(xs []ForecastRevision) map[string]ForecastRevision {
	out := make(map[string]ForecastRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastRevision builds a valid ForecastRevision, for tests and for the seed data the cron
// binary writes.
func SampleForecastRevision(id string, at time.Time) ForecastRevision {
	return ForecastRevision{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Channel:  "seeded by forecast.SampleForecastRevision",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastSnapshotSchema is the stored shape of a ForecastSnapshot.
var ForecastSnapshotSchema = schema.New("forecast.forecastsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastSnapshot is one forecast snapshot.
type ForecastSnapshot struct {
	ID       string
	Name     string
	Owner    string
	Group    string
	Weight   int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastSnapshot) Key() string { return "forecast/forecastsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x ForecastSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"group":    x.Group,
		"weight":   strconv.Itoa(x.Weight),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastSnapshot decodes a stored record.
func ParseForecastSnapshot(f codec.Fields) (ForecastSnapshot, error) {
	var x ForecastSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastsnapshot: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastsnapshot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastSnapshot) String() string {
	return fmt.Sprintf("ForecastSnapshot(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastSnapshot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastSnapshot) Normalize() ForecastSnapshot {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortForecastSnapshots orders by weight descending, then by ID.
func SortForecastSnapshots(xs []ForecastSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastSnapshots drops the archived entries, preserving order.
func ActiveForecastSnapshots(xs []ForecastSnapshot) []ForecastSnapshot {
	out := make([]ForecastSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastSnapshots adds up the weight of every active entry.
func SumForecastSnapshots(xs []ForecastSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexForecastSnapshots keys the entries by ID, last write winning.
func IndexForecastSnapshots(xs []ForecastSnapshot) map[string]ForecastSnapshot {
	out := make(map[string]ForecastSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastSnapshot builds a valid ForecastSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleForecastSnapshot(id string, at time.Time) ForecastSnapshot {
	return ForecastSnapshot{
		ID:      id,
		Name:    "name-" + id,
		Owner:   "owner-" + id,
		Group:   "seeded by forecast.SampleForecastSnapshot",
		Weight:  len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// ForecastBatchSchema is the stored shape of a ForecastBatch.
var ForecastBatchSchema = schema.New("forecast.forecastbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastBatch is one forecast batch.
type ForecastBatch struct {
	ID       string
	Region   string
	Group    string
	Owner    string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastBatch) Key() string { return "forecast/forecastbatch/" + x.ID }

// Fields encodes x for storage.
func (x ForecastBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"group":    x.Group,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastBatch decodes a stored record.
func ParseForecastBatch(f codec.Fields) (ForecastBatch, error) {
	var x ForecastBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastbatch: missing id")
	}
	x.Region = f["region"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastbatch %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastBatch) String() string {
	return fmt.Sprintf("ForecastBatch(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastBatch) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastBatch) Normalize() ForecastBatch {
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortForecastBatchs orders by priority descending, then by ID.
func SortForecastBatchs(xs []ForecastBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastBatchs drops the archived entries, preserving order.
func ActiveForecastBatchs(xs []ForecastBatch) []ForecastBatch {
	out := make([]ForecastBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastBatchs adds up the priority of every active entry.
func SumForecastBatchs(xs []ForecastBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexForecastBatchs keys the entries by ID, last write winning.
func IndexForecastBatchs(xs []ForecastBatch) map[string]ForecastBatch {
	out := make(map[string]ForecastBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastBatch builds a valid ForecastBatch, for tests and for the seed data the cron
// binary writes.
func SampleForecastBatch(id string, at time.Time) ForecastBatch {
	return ForecastBatch{
		ID:       id,
		Region:   "region-" + id,
		Group:    "group-" + id,
		Owner:    "seeded by forecast.SampleForecastBatch",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastRequestSchema is the stored shape of a ForecastRequest.
var ForecastRequestSchema = schema.New("forecast.forecastrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastRequest is one forecast request.
type ForecastRequest struct {
	ID       string
	Channel  string
	Owner    string
	Group    string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastRequest) Key() string { return "forecast/forecastrequest/" + x.ID }

// Fields encodes x for storage.
func (x ForecastRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"group":    x.Group,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastRequest decodes a stored record.
func ParseForecastRequest(f codec.Fields) (ForecastRequest, error) {
	var x ForecastRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastrequest: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastRequest) String() string {
	return fmt.Sprintf("ForecastRequest(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastRequest) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastRequest) Normalize() ForecastRequest {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortForecastRequests orders by weight descending, then by ID.
func SortForecastRequests(xs []ForecastRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastRequests drops the archived entries, preserving order.
func ActiveForecastRequests(xs []ForecastRequest) []ForecastRequest {
	out := make([]ForecastRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastRequests adds up the weight of every active entry.
func SumForecastRequests(xs []ForecastRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexForecastRequests keys the entries by ID, last write winning.
func IndexForecastRequests(xs []ForecastRequest) map[string]ForecastRequest {
	out := make(map[string]ForecastRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastRequest builds a valid ForecastRequest, for tests and for the seed data the cron
// binary writes.
func SampleForecastRequest(id string, at time.Time) ForecastRequest {
	return ForecastRequest{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by forecast.SampleForecastRequest",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastSummarySchema is the stored shape of a ForecastSummary.
var ForecastSummarySchema = schema.New("forecast.forecastsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastSummary is one forecast summary.
type ForecastSummary struct {
	ID        string
	Category  string
	Reference string
	Channel   string
	Offset    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ForecastSummary) Key() string { return "forecast/forecastsummary/" + x.ID }

// Fields encodes x for storage.
func (x ForecastSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"channel":   x.Channel,
		"offset":    strconv.Itoa(x.Offset),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastSummary decodes a stored record.
func ParseForecastSummary(f codec.Fields) (ForecastSummary, error) {
	var x ForecastSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastsummary: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastsummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastSummary) String() string {
	return fmt.Sprintf("ForecastSummary(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastSummary) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastSummary) Normalize() ForecastSummary {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortForecastSummarys orders by offset descending, then by ID.
func SortForecastSummarys(xs []ForecastSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastSummarys drops the archived entries, preserving order.
func ActiveForecastSummarys(xs []ForecastSummary) []ForecastSummary {
	out := make([]ForecastSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastSummarys adds up the offset of every active entry.
func SumForecastSummarys(xs []ForecastSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexForecastSummarys keys the entries by ID, last write winning.
func IndexForecastSummarys(xs []ForecastSummary) map[string]ForecastSummary {
	out := make(map[string]ForecastSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastSummary builds a valid ForecastSummary, for tests and for the seed data the cron
// binary writes.
func SampleForecastSummary(id string, at time.Time) ForecastSummary {
	return ForecastSummary{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by forecast.SampleForecastSummary",
		Offset:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ForecastPolicySchema is the stored shape of a ForecastPolicy.
var ForecastPolicySchema = schema.New("forecast.forecastpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastPolicy is one forecast policy.
type ForecastPolicy struct {
	ID       string
	Source   string
	Category string
	Group    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastPolicy) Key() string { return "forecast/forecastpolicy/" + x.ID }

// Fields encodes x for storage.
func (x ForecastPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"category": x.Category,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastPolicy decodes a stored record.
func ParseForecastPolicy(f codec.Fields) (ForecastPolicy, error) {
	var x ForecastPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastpolicy: missing id")
	}
	x.Source = f["source"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastpolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastPolicy) String() string {
	return fmt.Sprintf("ForecastPolicy(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastPolicy) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastPolicy) Normalize() ForecastPolicy {
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortForecastPolicys orders by capacity descending, then by ID.
func SortForecastPolicys(xs []ForecastPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastPolicys drops the archived entries, preserving order.
func ActiveForecastPolicys(xs []ForecastPolicy) []ForecastPolicy {
	out := make([]ForecastPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastPolicys adds up the capacity of every active entry.
func SumForecastPolicys(xs []ForecastPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexForecastPolicys keys the entries by ID, last write winning.
func IndexForecastPolicys(xs []ForecastPolicy) map[string]ForecastPolicy {
	out := make(map[string]ForecastPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastPolicy builds a valid ForecastPolicy, for tests and for the seed data the cron
// binary writes.
func SampleForecastPolicy(id string, at time.Time) ForecastPolicy {
	return ForecastPolicy{
		ID:       id,
		Source:   "source-" + id,
		Category: "category-" + id,
		Group:    "seeded by forecast.SampleForecastPolicy",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastLinkSchema is the stored shape of a ForecastLink.
var ForecastLinkSchema = schema.New("forecast.forecastlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastLink is one forecast link.
type ForecastLink struct {
	ID       string
	Category string
	Region   string
	Comment  string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastLink) Key() string { return "forecast/forecastlink/" + x.ID }

// Fields encodes x for storage.
func (x ForecastLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastLink decodes a stored record.
func ParseForecastLink(f codec.Fields) (ForecastLink, error) {
	var x ForecastLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastlink: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastlink %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastLink) String() string {
	return fmt.Sprintf("ForecastLink(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastLink) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastLink) Normalize() ForecastLink {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortForecastLinks orders by priority descending, then by ID.
func SortForecastLinks(xs []ForecastLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastLinks drops the archived entries, preserving order.
func ActiveForecastLinks(xs []ForecastLink) []ForecastLink {
	out := make([]ForecastLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastLinks adds up the priority of every active entry.
func SumForecastLinks(xs []ForecastLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexForecastLinks keys the entries by ID, last write winning.
func IndexForecastLinks(xs []ForecastLink) map[string]ForecastLink {
	out := make(map[string]ForecastLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastLink builds a valid ForecastLink, for tests and for the seed data the cron
// binary writes.
func SampleForecastLink(id string, at time.Time) ForecastLink {
	return ForecastLink{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Comment:  "seeded by forecast.SampleForecastLink",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastNoteSchema is the stored shape of a ForecastNote.
var ForecastNoteSchema = schema.New("forecast.forecastnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastNote is one forecast note.
type ForecastNote struct {
	ID       string
	Owner    string
	Region   string
	Source   string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastNote) Key() string { return "forecast/forecastnote/" + x.ID }

// Fields encodes x for storage.
func (x ForecastNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastNote decodes a stored record.
func ParseForecastNote(f codec.Fields) (ForecastNote, error) {
	var x ForecastNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastnote: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastnote %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastNote) String() string {
	return fmt.Sprintf("ForecastNote(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastNote) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastNote) Normalize() ForecastNote {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortForecastNotes orders by offset descending, then by ID.
func SortForecastNotes(xs []ForecastNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastNotes drops the archived entries, preserving order.
func ActiveForecastNotes(xs []ForecastNote) []ForecastNote {
	out := make([]ForecastNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastNotes adds up the offset of every active entry.
func SumForecastNotes(xs []ForecastNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexForecastNotes keys the entries by ID, last write winning.
func IndexForecastNotes(xs []ForecastNote) map[string]ForecastNote {
	out := make(map[string]ForecastNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastNote builds a valid ForecastNote, for tests and for the seed data the cron
// binary writes.
func SampleForecastNote(id string, at time.Time) ForecastNote {
	return ForecastNote{
		ID:      id,
		Owner:   "owner-" + id,
		Region:  "region-" + id,
		Source:  "seeded by forecast.SampleForecastNote",
		Offset:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// ForecastWindowSchema is the stored shape of a ForecastWindow.
var ForecastWindowSchema = schema.New("forecast.forecastwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastWindow is one forecast window.
type ForecastWindow struct {
	ID       string
	Channel  string
	Label    string
	Owner    string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastWindow) Key() string { return "forecast/forecastwindow/" + x.ID }

// Fields encodes x for storage.
func (x ForecastWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastWindow decodes a stored record.
func ParseForecastWindow(f codec.Fields) (ForecastWindow, error) {
	var x ForecastWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastwindow: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastwindow %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastWindow) String() string {
	return fmt.Sprintf("ForecastWindow(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastWindow) Normalize() ForecastWindow {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortForecastWindows orders by priority descending, then by ID.
func SortForecastWindows(xs []ForecastWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastWindows drops the archived entries, preserving order.
func ActiveForecastWindows(xs []ForecastWindow) []ForecastWindow {
	out := make([]ForecastWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastWindows adds up the priority of every active entry.
func SumForecastWindows(xs []ForecastWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexForecastWindows keys the entries by ID, last write winning.
func IndexForecastWindows(xs []ForecastWindow) map[string]ForecastWindow {
	out := make(map[string]ForecastWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastWindow builds a valid ForecastWindow, for tests and for the seed data the cron
// binary writes.
func SampleForecastWindow(id string, at time.Time) ForecastWindow {
	return ForecastWindow{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Owner:    "seeded by forecast.SampleForecastWindow",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastQuotaSchema is the stored shape of a ForecastQuota.
var ForecastQuotaSchema = schema.New("forecast.forecastquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastQuota is one forecast quota.
type ForecastQuota struct {
	ID       string
	Owner    string
	Label    string
	Group    string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastQuota) Key() string { return "forecast/forecastquota/" + x.ID }

// Fields encodes x for storage.
func (x ForecastQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"label":    x.Label,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastQuota decodes a stored record.
func ParseForecastQuota(f codec.Fields) (ForecastQuota, error) {
	var x ForecastQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastquota: missing id")
	}
	x.Owner = f["owner"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastquota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastQuota) String() string {
	return fmt.Sprintf("ForecastQuota(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastQuota) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastQuota) Normalize() ForecastQuota {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortForecastQuotas orders by retained descending, then by ID.
func SortForecastQuotas(xs []ForecastQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastQuotas drops the archived entries, preserving order.
func ActiveForecastQuotas(xs []ForecastQuota) []ForecastQuota {
	out := make([]ForecastQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastQuotas adds up the retained of every active entry.
func SumForecastQuotas(xs []ForecastQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexForecastQuotas keys the entries by ID, last write winning.
func IndexForecastQuotas(xs []ForecastQuota) map[string]ForecastQuota {
	out := make(map[string]ForecastQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastQuota builds a valid ForecastQuota, for tests and for the seed data the cron
// binary writes.
func SampleForecastQuota(id string, at time.Time) ForecastQuota {
	return ForecastQuota{
		ID:       id,
		Owner:    "owner-" + id,
		Label:    "label-" + id,
		Group:    "seeded by forecast.SampleForecastQuota",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastBucketSchema is the stored shape of a ForecastBucket.
var ForecastBucketSchema = schema.New("forecast.forecastbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastBucket is one forecast bucket.
type ForecastBucket struct {
	ID       string
	Source   string
	Label    string
	Channel  string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastBucket) Key() string { return "forecast/forecastbucket/" + x.ID }

// Fields encodes x for storage.
func (x ForecastBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"label":    x.Label,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastBucket decodes a stored record.
func ParseForecastBucket(f codec.Fields) (ForecastBucket, error) {
	var x ForecastBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastbucket: missing id")
	}
	x.Source = f["source"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastbucket %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastBucket) String() string {
	return fmt.Sprintf("ForecastBucket(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastBucket) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastBucket) Normalize() ForecastBucket {
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortForecastBuckets orders by capacity descending, then by ID.
func SortForecastBuckets(xs []ForecastBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastBuckets drops the archived entries, preserving order.
func ActiveForecastBuckets(xs []ForecastBucket) []ForecastBucket {
	out := make([]ForecastBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastBuckets adds up the capacity of every active entry.
func SumForecastBuckets(xs []ForecastBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexForecastBuckets keys the entries by ID, last write winning.
func IndexForecastBuckets(xs []ForecastBucket) map[string]ForecastBucket {
	out := make(map[string]ForecastBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastBucket builds a valid ForecastBucket, for tests and for the seed data the cron
// binary writes.
func SampleForecastBucket(id string, at time.Time) ForecastBucket {
	return ForecastBucket{
		ID:       id,
		Source:   "source-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by forecast.SampleForecastBucket",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastLaneSchema is the stored shape of a ForecastLane.
var ForecastLaneSchema = schema.New("forecast.forecastlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastLane is one forecast lane.
type ForecastLane struct {
	ID       string
	Owner    string
	Group    string
	Comment  string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastLane) Key() string { return "forecast/forecastlane/" + x.ID }

// Fields encodes x for storage.
func (x ForecastLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastLane decodes a stored record.
func ParseForecastLane(f codec.Fields) (ForecastLane, error) {
	var x ForecastLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastlane: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastlane %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastLane) String() string {
	return fmt.Sprintf("ForecastLane(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastLane) Normalize() ForecastLane {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortForecastLanes orders by capacity descending, then by ID.
func SortForecastLanes(xs []ForecastLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastLanes drops the archived entries, preserving order.
func ActiveForecastLanes(xs []ForecastLane) []ForecastLane {
	out := make([]ForecastLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastLanes adds up the capacity of every active entry.
func SumForecastLanes(xs []ForecastLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexForecastLanes keys the entries by ID, last write winning.
func IndexForecastLanes(xs []ForecastLane) map[string]ForecastLane {
	out := make(map[string]ForecastLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastLane builds a valid ForecastLane, for tests and for the seed data the cron
// binary writes.
func SampleForecastLane(id string, at time.Time) ForecastLane {
	return ForecastLane{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Comment:  "seeded by forecast.SampleForecastLane",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastSlotSchema is the stored shape of a ForecastSlot.
var ForecastSlotSchema = schema.New("forecast.forecastslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastSlot is one forecast slot.
type ForecastSlot struct {
	ID       string
	Label    string
	Group    string
	Category string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastSlot) Key() string { return "forecast/forecastslot/" + x.ID }

// Fields encodes x for storage.
func (x ForecastSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"category": x.Category,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastSlot decodes a stored record.
func ParseForecastSlot(f codec.Fields) (ForecastSlot, error) {
	var x ForecastSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastslot: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastslot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastSlot) String() string {
	return fmt.Sprintf("ForecastSlot(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastSlot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastSlot) Normalize() ForecastSlot {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortForecastSlots orders by retained descending, then by ID.
func SortForecastSlots(xs []ForecastSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastSlots drops the archived entries, preserving order.
func ActiveForecastSlots(xs []ForecastSlot) []ForecastSlot {
	out := make([]ForecastSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastSlots adds up the retained of every active entry.
func SumForecastSlots(xs []ForecastSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexForecastSlots keys the entries by ID, last write winning.
func IndexForecastSlots(xs []ForecastSlot) map[string]ForecastSlot {
	out := make(map[string]ForecastSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastSlot builds a valid ForecastSlot, for tests and for the seed data the cron
// binary writes.
func SampleForecastSlot(id string, at time.Time) ForecastSlot {
	return ForecastSlot{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Category: "seeded by forecast.SampleForecastSlot",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastTagSchema is the stored shape of a ForecastTag.
var ForecastTagSchema = schema.New("forecast.forecasttag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastTag is one forecast tag.
type ForecastTag struct {
	ID       string
	Group    string
	Source   string
	Region   string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastTag) Key() string { return "forecast/forecasttag/" + x.ID }

// Fields encodes x for storage.
func (x ForecastTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"region":   x.Region,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastTag decodes a stored record.
func ParseForecastTag(f codec.Fields) (ForecastTag, error) {
	var x ForecastTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecasttag: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecasttag %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecasttag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecasttag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecasttag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastTag) String() string {
	return fmt.Sprintf("ForecastTag(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastTag) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastTag) Normalize() ForecastTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortForecastTags orders by priority descending, then by ID.
func SortForecastTags(xs []ForecastTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastTags drops the archived entries, preserving order.
func ActiveForecastTags(xs []ForecastTag) []ForecastTag {
	out := make([]ForecastTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastTags adds up the priority of every active entry.
func SumForecastTags(xs []ForecastTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexForecastTags keys the entries by ID, last write winning.
func IndexForecastTags(xs []ForecastTag) map[string]ForecastTag {
	out := make(map[string]ForecastTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastTag builds a valid ForecastTag, for tests and for the seed data the cron
// binary writes.
func SampleForecastTag(id string, at time.Time) ForecastTag {
	return ForecastTag{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Region:   "seeded by forecast.SampleForecastTag",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastRuleSchema is the stored shape of a ForecastRule.
var ForecastRuleSchema = schema.New("forecast.forecastrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastRule is one forecast rule.
type ForecastRule struct {
	ID        string
	Reference string
	Region    string
	Category  string
	Priority  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ForecastRule) Key() string { return "forecast/forecastrule/" + x.ID }

// Fields encodes x for storage.
func (x ForecastRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"category":  x.Category,
		"priority":  strconv.Itoa(x.Priority),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastRule decodes a stored record.
func ParseForecastRule(f codec.Fields) (ForecastRule, error) {
	var x ForecastRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastrule: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrule %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastRule) String() string {
	return fmt.Sprintf("ForecastRule(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastRule) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastRule) Normalize() ForecastRule {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortForecastRules orders by priority descending, then by ID.
func SortForecastRules(xs []ForecastRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastRules drops the archived entries, preserving order.
func ActiveForecastRules(xs []ForecastRule) []ForecastRule {
	out := make([]ForecastRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastRules adds up the priority of every active entry.
func SumForecastRules(xs []ForecastRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexForecastRules keys the entries by ID, last write winning.
func IndexForecastRules(xs []ForecastRule) map[string]ForecastRule {
	out := make(map[string]ForecastRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastRule builds a valid ForecastRule, for tests and for the seed data the cron
// binary writes.
func SampleForecastRule(id string, at time.Time) ForecastRule {
	return ForecastRule{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Category:  "seeded by forecast.SampleForecastRule",
		Priority:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ForecastHoldSchema is the stored shape of a ForecastHold.
var ForecastHoldSchema = schema.New("forecast.forecasthold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastHold is one forecast hold.
type ForecastHold struct {
	ID       string
	Group    string
	Region   string
	Source   string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastHold) Key() string { return "forecast/forecasthold/" + x.ID }

// Fields encodes x for storage.
func (x ForecastHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"region":   x.Region,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastHold decodes a stored record.
func ParseForecastHold(f codec.Fields) (ForecastHold, error) {
	var x ForecastHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecasthold: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("forecast: forecasthold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("forecast: forecasthold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecasthold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecasthold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastHold) String() string {
	return fmt.Sprintf("ForecastHold(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastHold) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastHold) Normalize() ForecastHold {
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortForecastHolds orders by retained descending, then by ID.
func SortForecastHolds(xs []ForecastHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastHolds drops the archived entries, preserving order.
func ActiveForecastHolds(xs []ForecastHold) []ForecastHold {
	out := make([]ForecastHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastHolds adds up the retained of every active entry.
func SumForecastHolds(xs []ForecastHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexForecastHolds keys the entries by ID, last write winning.
func IndexForecastHolds(xs []ForecastHold) map[string]ForecastHold {
	out := make(map[string]ForecastHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastHold builds a valid ForecastHold, for tests and for the seed data the cron
// binary writes.
func SampleForecastHold(id string, at time.Time) ForecastHold {
	return ForecastHold{
		ID:       id,
		Group:    "group-" + id,
		Region:   "region-" + id,
		Source:   "seeded by forecast.SampleForecastHold",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastClaimSchema is the stored shape of a ForecastClaim.
var ForecastClaimSchema = schema.New("forecast.forecastclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastClaim is one forecast claim.
type ForecastClaim struct {
	ID       string
	Name     string
	Comment  string
	Group    string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ForecastClaim) Key() string { return "forecast/forecastclaim/" + x.ID }

// Fields encodes x for storage.
func (x ForecastClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"comment":  x.Comment,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastClaim decodes a stored record.
func ParseForecastClaim(f codec.Fields) (ForecastClaim, error) {
	var x ForecastClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastclaim: missing id")
	}
	x.Name = f["name"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastclaim %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastClaim) String() string {
	return fmt.Sprintf("ForecastClaim(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastClaim) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastClaim) Normalize() ForecastClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortForecastClaims orders by offset descending, then by ID.
func SortForecastClaims(xs []ForecastClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastClaims drops the archived entries, preserving order.
func ActiveForecastClaims(xs []ForecastClaim) []ForecastClaim {
	out := make([]ForecastClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastClaims adds up the offset of every active entry.
func SumForecastClaims(xs []ForecastClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexForecastClaims keys the entries by ID, last write winning.
func IndexForecastClaims(xs []ForecastClaim) map[string]ForecastClaim {
	out := make(map[string]ForecastClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastClaim builds a valid ForecastClaim, for tests and for the seed data the cron
// binary writes.
func SampleForecastClaim(id string, at time.Time) ForecastClaim {
	return ForecastClaim{
		ID:       id,
		Name:     "name-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by forecast.SampleForecastClaim",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ForecastCursorSchema is the stored shape of a ForecastCursor.
var ForecastCursorSchema = schema.New("forecast.forecastcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ForecastCursor is one forecast cursor.
type ForecastCursor struct {
	ID        string
	Category  string
	Reference string
	Source    string
	Attempts  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ForecastCursor) Key() string { return "forecast/forecastcursor/" + x.ID }

// Fields encodes x for storage.
func (x ForecastCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"source":    x.Source,
		"attempts":  strconv.Itoa(x.Attempts),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseForecastCursor decodes a stored record.
func ParseForecastCursor(f codec.Fields) (ForecastCursor, error) {
	var x ForecastCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("forecast: forecastcursor: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastcursor %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("forecast: forecastcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("forecast: forecastcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("forecast: forecastcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ForecastCursor) String() string {
	return fmt.Sprintf("ForecastCursor(%s, category=%q, attempts=%d, archived=%t)", x.ID, x.Category, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ForecastCursor) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ForecastCursor) Normalize() ForecastCursor {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortForecastCursors orders by attempts descending, then by ID.
func SortForecastCursors(xs []ForecastCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveForecastCursors drops the archived entries, preserving order.
func ActiveForecastCursors(xs []ForecastCursor) []ForecastCursor {
	out := make([]ForecastCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumForecastCursors adds up the attempts of every active entry.
func SumForecastCursors(xs []ForecastCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexForecastCursors keys the entries by ID, last write winning.
func IndexForecastCursors(xs []ForecastCursor) map[string]ForecastCursor {
	out := make(map[string]ForecastCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleForecastCursor builds a valid ForecastCursor, for tests and for the seed data the cron
// binary writes.
func SampleForecastCursor(id string, at time.Time) ForecastCursor {
	return ForecastCursor{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by forecast.SampleForecastCursor",
		Attempts:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	ForecastRecordSchema,
	ForecastEntrySchema,
	ForecastItemSchema,
	ForecastDraftSchema,
	ForecastRevisionSchema,
	ForecastSnapshotSchema,
	ForecastBatchSchema,
	ForecastRequestSchema,
	ForecastSummarySchema,
	ForecastPolicySchema,
	ForecastLinkSchema,
	ForecastNoteSchema,
	ForecastWindowSchema,
	ForecastQuotaSchema,
	ForecastBucketSchema,
	ForecastLaneSchema,
	ForecastSlotSchema,
	ForecastTagSchema,
	ForecastRuleSchema,
	ForecastHoldSchema,
	ForecastClaimSchema,
	ForecastCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"forecast/forecastrecord/",
	"forecast/forecastentry/",
	"forecast/forecastitem/",
	"forecast/forecastdraft/",
	"forecast/forecastrevision/",
	"forecast/forecastsnapshot/",
	"forecast/forecastbatch/",
	"forecast/forecastrequest/",
	"forecast/forecastsummary/",
	"forecast/forecastpolicy/",
	"forecast/forecastlink/",
	"forecast/forecastnote/",
	"forecast/forecastwindow/",
	"forecast/forecastquota/",
	"forecast/forecastbucket/",
	"forecast/forecastlane/",
	"forecast/forecastslot/",
	"forecast/forecasttag/",
	"forecast/forecastrule/",
	"forecast/forecasthold/",
	"forecast/forecastclaim/",
	"forecast/forecastcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "forecast: " + strings.Join(names, ", ")
}
