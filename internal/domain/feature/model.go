// Package feature holds the feature domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package feature

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// FeatureRecordSchema is the stored shape of a FeatureRecord.
var FeatureRecordSchema = schema.New("feature.featurerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureRecord is one feature record.
type FeatureRecord struct {
	ID       string
	Source   string
	Comment  string
	Region   string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureRecord) Key() string { return "feature/featurerecord/" + x.ID }

// Fields encodes x for storage.
func (x FeatureRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureRecord decodes a stored record.
func ParseFeatureRecord(f codec.Fields) (FeatureRecord, error) {
	var x FeatureRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurerecord: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featurerecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("feature: featurerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureRecord) String() string {
	return fmt.Sprintf("FeatureRecord(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureRecord) Normalize() FeatureRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortFeatureRecords orders by capacity descending, then by ID.
func SortFeatureRecords(xs []FeatureRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureRecords drops the archived entries, preserving order.
func ActiveFeatureRecords(xs []FeatureRecord) []FeatureRecord {
	out := make([]FeatureRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureRecords adds up the capacity of every active entry.
func SumFeatureRecords(xs []FeatureRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexFeatureRecords keys the entries by ID, last write winning.
func IndexFeatureRecords(xs []FeatureRecord) map[string]FeatureRecord {
	out := make(map[string]FeatureRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureRecord builds a valid FeatureRecord, for tests and for the seed data the cron
// binary writes.
func SampleFeatureRecord(id string, at time.Time) FeatureRecord {
	return FeatureRecord{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by feature.SampleFeatureRecord",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureEntrySchema is the stored shape of a FeatureEntry.
var FeatureEntrySchema = schema.New("feature.featureentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureEntry is one feature entry.
type FeatureEntry struct {
	ID       string
	Group    string
	Region   string
	Name     string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureEntry) Key() string { return "feature/featureentry/" + x.ID }

// Fields encodes x for storage.
func (x FeatureEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"region":   x.Region,
		"name":     x.Name,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureEntry decodes a stored record.
func ParseFeatureEntry(f codec.Fields) (FeatureEntry, error) {
	var x FeatureEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featureentry: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("feature: featureentry %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("feature: featureentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featureentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featureentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureEntry) String() string {
	return fmt.Sprintf("FeatureEntry(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureEntry) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureEntry) Normalize() FeatureEntry {
	x.Group = strings.TrimSpace(x.Group)
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

// SortFeatureEntrys orders by priority descending, then by ID.
func SortFeatureEntrys(xs []FeatureEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureEntrys drops the archived entries, preserving order.
func ActiveFeatureEntrys(xs []FeatureEntry) []FeatureEntry {
	out := make([]FeatureEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureEntrys adds up the priority of every active entry.
func SumFeatureEntrys(xs []FeatureEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexFeatureEntrys keys the entries by ID, last write winning.
func IndexFeatureEntrys(xs []FeatureEntry) map[string]FeatureEntry {
	out := make(map[string]FeatureEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureEntry builds a valid FeatureEntry, for tests and for the seed data the cron
// binary writes.
func SampleFeatureEntry(id string, at time.Time) FeatureEntry {
	return FeatureEntry{
		ID:       id,
		Group:    "group-" + id,
		Region:   "region-" + id,
		Name:     "seeded by feature.SampleFeatureEntry",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureItemSchema is the stored shape of a FeatureItem.
var FeatureItemSchema = schema.New("feature.featureitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureItem is one feature item.
type FeatureItem struct {
	ID       string
	Source   string
	Comment  string
	Label    string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureItem) Key() string { return "feature/featureitem/" + x.ID }

// Fields encodes x for storage.
func (x FeatureItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureItem decodes a stored record.
func ParseFeatureItem(f codec.Fields) (FeatureItem, error) {
	var x FeatureItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featureitem: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("feature: featureitem %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featureitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featureitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featureitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureItem) String() string {
	return fmt.Sprintf("FeatureItem(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureItem) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureItem) Normalize() FeatureItem {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortFeatureItems orders by retained descending, then by ID.
func SortFeatureItems(xs []FeatureItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureItems drops the archived entries, preserving order.
func ActiveFeatureItems(xs []FeatureItem) []FeatureItem {
	out := make([]FeatureItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureItems adds up the retained of every active entry.
func SumFeatureItems(xs []FeatureItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexFeatureItems keys the entries by ID, last write winning.
func IndexFeatureItems(xs []FeatureItem) map[string]FeatureItem {
	out := make(map[string]FeatureItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureItem builds a valid FeatureItem, for tests and for the seed data the cron
// binary writes.
func SampleFeatureItem(id string, at time.Time) FeatureItem {
	return FeatureItem{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Label:    "seeded by feature.SampleFeatureItem",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureDraftSchema is the stored shape of a FeatureDraft.
var FeatureDraftSchema = schema.New("feature.featuredraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureDraft is one feature draft.
type FeatureDraft struct {
	ID       string
	Source   string
	Label    string
	Category string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureDraft) Key() string { return "feature/featuredraft/" + x.ID }

// Fields encodes x for storage.
func (x FeatureDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"label":    x.Label,
		"category": x.Category,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureDraft decodes a stored record.
func ParseFeatureDraft(f codec.Fields) (FeatureDraft, error) {
	var x FeatureDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featuredraft: missing id")
	}
	x.Source = f["source"]
	x.Label = f["label"]
	x.Category = f["category"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("feature: featuredraft %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("feature: featuredraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featuredraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featuredraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureDraft) String() string {
	return fmt.Sprintf("FeatureDraft(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureDraft) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureDraft) Normalize() FeatureDraft {
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortFeatureDrafts orders by retained descending, then by ID.
func SortFeatureDrafts(xs []FeatureDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureDrafts drops the archived entries, preserving order.
func ActiveFeatureDrafts(xs []FeatureDraft) []FeatureDraft {
	out := make([]FeatureDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureDrafts adds up the retained of every active entry.
func SumFeatureDrafts(xs []FeatureDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexFeatureDrafts keys the entries by ID, last write winning.
func IndexFeatureDrafts(xs []FeatureDraft) map[string]FeatureDraft {
	out := make(map[string]FeatureDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureDraft builds a valid FeatureDraft, for tests and for the seed data the cron
// binary writes.
func SampleFeatureDraft(id string, at time.Time) FeatureDraft {
	return FeatureDraft{
		ID:       id,
		Source:   "source-" + id,
		Label:    "label-" + id,
		Category: "seeded by feature.SampleFeatureDraft",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureRevisionSchema is the stored shape of a FeatureRevision.
var FeatureRevisionSchema = schema.New("feature.featurerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureRevision is one feature revision.
type FeatureRevision struct {
	ID        string
	Source    string
	Reference string
	Name      string
	Capacity  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x FeatureRevision) Key() string { return "feature/featurerevision/" + x.ID }

// Fields encodes x for storage.
func (x FeatureRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"reference": x.Reference,
		"name":      x.Name,
		"capacity":  strconv.Itoa(x.Capacity),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureRevision decodes a stored record.
func ParseFeatureRevision(f codec.Fields) (FeatureRevision, error) {
	var x FeatureRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurerevision: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featurerevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("feature: featurerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureRevision) String() string {
	return fmt.Sprintf("FeatureRevision(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureRevision) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureRevision) Normalize() FeatureRevision {
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortFeatureRevisions orders by capacity descending, then by ID.
func SortFeatureRevisions(xs []FeatureRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureRevisions drops the archived entries, preserving order.
func ActiveFeatureRevisions(xs []FeatureRevision) []FeatureRevision {
	out := make([]FeatureRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureRevisions adds up the capacity of every active entry.
func SumFeatureRevisions(xs []FeatureRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexFeatureRevisions keys the entries by ID, last write winning.
func IndexFeatureRevisions(xs []FeatureRevision) map[string]FeatureRevision {
	out := make(map[string]FeatureRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureRevision builds a valid FeatureRevision, for tests and for the seed data the cron
// binary writes.
func SampleFeatureRevision(id string, at time.Time) FeatureRevision {
	return FeatureRevision{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by feature.SampleFeatureRevision",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// FeatureSnapshotSchema is the stored shape of a FeatureSnapshot.
var FeatureSnapshotSchema = schema.New("feature.featuresnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureSnapshot is one feature snapshot.
type FeatureSnapshot struct {
	ID        string
	Comment   string
	Reference string
	Source    string
	Retained  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x FeatureSnapshot) Key() string { return "feature/featuresnapshot/" + x.ID }

// Fields encodes x for storage.
func (x FeatureSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"source":    x.Source,
		"retained":  strconv.Itoa(x.Retained),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureSnapshot decodes a stored record.
func ParseFeatureSnapshot(f codec.Fields) (FeatureSnapshot, error) {
	var x FeatureSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featuresnapshot: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("feature: featuresnapshot %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("feature: featuresnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featuresnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featuresnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureSnapshot) String() string {
	return fmt.Sprintf("FeatureSnapshot(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureSnapshot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureSnapshot) Normalize() FeatureSnapshot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortFeatureSnapshots orders by retained descending, then by ID.
func SortFeatureSnapshots(xs []FeatureSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureSnapshots drops the archived entries, preserving order.
func ActiveFeatureSnapshots(xs []FeatureSnapshot) []FeatureSnapshot {
	out := make([]FeatureSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureSnapshots adds up the retained of every active entry.
func SumFeatureSnapshots(xs []FeatureSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexFeatureSnapshots keys the entries by ID, last write winning.
func IndexFeatureSnapshots(xs []FeatureSnapshot) map[string]FeatureSnapshot {
	out := make(map[string]FeatureSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureSnapshot builds a valid FeatureSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleFeatureSnapshot(id string, at time.Time) FeatureSnapshot {
	return FeatureSnapshot{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by feature.SampleFeatureSnapshot",
		Retained:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// FeatureBatchSchema is the stored shape of a FeatureBatch.
var FeatureBatchSchema = schema.New("feature.featurebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureBatch is one feature batch.
type FeatureBatch struct {
	ID       string
	Category string
	Owner    string
	Name     string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureBatch) Key() string { return "feature/featurebatch/" + x.ID }

// Fields encodes x for storage.
func (x FeatureBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"name":     x.Name,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureBatch decodes a stored record.
func ParseFeatureBatch(f codec.Fields) (FeatureBatch, error) {
	var x FeatureBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurebatch: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("feature: featurebatch %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("feature: featurebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureBatch) String() string {
	return fmt.Sprintf("FeatureBatch(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureBatch) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureBatch) Normalize() FeatureBatch {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortFeatureBatchs orders by offset descending, then by ID.
func SortFeatureBatchs(xs []FeatureBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureBatchs drops the archived entries, preserving order.
func ActiveFeatureBatchs(xs []FeatureBatch) []FeatureBatch {
	out := make([]FeatureBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureBatchs adds up the offset of every active entry.
func SumFeatureBatchs(xs []FeatureBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexFeatureBatchs keys the entries by ID, last write winning.
func IndexFeatureBatchs(xs []FeatureBatch) map[string]FeatureBatch {
	out := make(map[string]FeatureBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureBatch builds a valid FeatureBatch, for tests and for the seed data the cron
// binary writes.
func SampleFeatureBatch(id string, at time.Time) FeatureBatch {
	return FeatureBatch{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by feature.SampleFeatureBatch",
		Offset:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureRequestSchema is the stored shape of a FeatureRequest.
var FeatureRequestSchema = schema.New("feature.featurerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureRequest is one feature request.
type FeatureRequest struct {
	ID       string
	Region   string
	Channel  string
	Comment  string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureRequest) Key() string { return "feature/featurerequest/" + x.ID }

// Fields encodes x for storage.
func (x FeatureRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureRequest decodes a stored record.
func ParseFeatureRequest(f codec.Fields) (FeatureRequest, error) {
	var x FeatureRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurerequest: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("feature: featurerequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featurerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureRequest) String() string {
	return fmt.Sprintf("FeatureRequest(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureRequest) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureRequest) Normalize() FeatureRequest {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortFeatureRequests orders by offset descending, then by ID.
func SortFeatureRequests(xs []FeatureRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureRequests drops the archived entries, preserving order.
func ActiveFeatureRequests(xs []FeatureRequest) []FeatureRequest {
	out := make([]FeatureRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureRequests adds up the offset of every active entry.
func SumFeatureRequests(xs []FeatureRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexFeatureRequests keys the entries by ID, last write winning.
func IndexFeatureRequests(xs []FeatureRequest) map[string]FeatureRequest {
	out := make(map[string]FeatureRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureRequest builds a valid FeatureRequest, for tests and for the seed data the cron
// binary writes.
func SampleFeatureRequest(id string, at time.Time) FeatureRequest {
	return FeatureRequest{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by feature.SampleFeatureRequest",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureSummarySchema is the stored shape of a FeatureSummary.
var FeatureSummarySchema = schema.New("feature.featuresummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureSummary is one feature summary.
type FeatureSummary struct {
	ID       string
	Source   string
	Region   string
	Group    string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureSummary) Key() string { return "feature/featuresummary/" + x.ID }

// Fields encodes x for storage.
func (x FeatureSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"region":   x.Region,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureSummary decodes a stored record.
func ParseFeatureSummary(f codec.Fields) (FeatureSummary, error) {
	var x FeatureSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featuresummary: missing id")
	}
	x.Source = f["source"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("feature: featuresummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("feature: featuresummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featuresummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featuresummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureSummary) String() string {
	return fmt.Sprintf("FeatureSummary(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureSummary) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureSummary) Normalize() FeatureSummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortFeatureSummarys orders by version descending, then by ID.
func SortFeatureSummarys(xs []FeatureSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureSummarys drops the archived entries, preserving order.
func ActiveFeatureSummarys(xs []FeatureSummary) []FeatureSummary {
	out := make([]FeatureSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureSummarys adds up the version of every active entry.
func SumFeatureSummarys(xs []FeatureSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexFeatureSummarys keys the entries by ID, last write winning.
func IndexFeatureSummarys(xs []FeatureSummary) map[string]FeatureSummary {
	out := make(map[string]FeatureSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureSummary builds a valid FeatureSummary, for tests and for the seed data the cron
// binary writes.
func SampleFeatureSummary(id string, at time.Time) FeatureSummary {
	return FeatureSummary{
		ID:       id,
		Source:   "source-" + id,
		Region:   "region-" + id,
		Group:    "seeded by feature.SampleFeatureSummary",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeaturePolicySchema is the stored shape of a FeaturePolicy.
var FeaturePolicySchema = schema.New("feature.featurepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeaturePolicy is one feature policy.
type FeaturePolicy struct {
	ID       string
	Owner    string
	Label    string
	Category string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeaturePolicy) Key() string { return "feature/featurepolicy/" + x.ID }

// Fields encodes x for storage.
func (x FeaturePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"label":    x.Label,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeaturePolicy decodes a stored record.
func ParseFeaturePolicy(f codec.Fields) (FeaturePolicy, error) {
	var x FeaturePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurepolicy: missing id")
	}
	x.Owner = f["owner"]
	x.Label = f["label"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featurepolicy %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featurepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeaturePolicy) String() string {
	return fmt.Sprintf("FeaturePolicy(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeaturePolicy) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeaturePolicy) Normalize() FeaturePolicy {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortFeaturePolicys orders by attempts descending, then by ID.
func SortFeaturePolicys(xs []FeaturePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeaturePolicys drops the archived entries, preserving order.
func ActiveFeaturePolicys(xs []FeaturePolicy) []FeaturePolicy {
	out := make([]FeaturePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeaturePolicys adds up the attempts of every active entry.
func SumFeaturePolicys(xs []FeaturePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexFeaturePolicys keys the entries by ID, last write winning.
func IndexFeaturePolicys(xs []FeaturePolicy) map[string]FeaturePolicy {
	out := make(map[string]FeaturePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeaturePolicy builds a valid FeaturePolicy, for tests and for the seed data the cron
// binary writes.
func SampleFeaturePolicy(id string, at time.Time) FeaturePolicy {
	return FeaturePolicy{
		ID:       id,
		Owner:    "owner-" + id,
		Label:    "label-" + id,
		Category: "seeded by feature.SampleFeaturePolicy",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureLinkSchema is the stored shape of a FeatureLink.
var FeatureLinkSchema = schema.New("feature.featurelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureLink is one feature link.
type FeatureLink struct {
	ID        string
	Category  string
	Region    string
	Reference string
	Count     int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x FeatureLink) Key() string { return "feature/featurelink/" + x.ID }

// Fields encodes x for storage.
func (x FeatureLink) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"region":    x.Region,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureLink decodes a stored record.
func ParseFeatureLink(f codec.Fields) (FeatureLink, error) {
	var x FeatureLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurelink: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("feature: featurelink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("feature: featurelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureLink) String() string {
	return fmt.Sprintf("FeatureLink(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureLink) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureLink) Normalize() FeatureLink {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortFeatureLinks orders by count descending, then by ID.
func SortFeatureLinks(xs []FeatureLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureLinks drops the archived entries, preserving order.
func ActiveFeatureLinks(xs []FeatureLink) []FeatureLink {
	out := make([]FeatureLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureLinks adds up the count of every active entry.
func SumFeatureLinks(xs []FeatureLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexFeatureLinks keys the entries by ID, last write winning.
func IndexFeatureLinks(xs []FeatureLink) map[string]FeatureLink {
	out := make(map[string]FeatureLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureLink builds a valid FeatureLink, for tests and for the seed data the cron
// binary writes.
func SampleFeatureLink(id string, at time.Time) FeatureLink {
	return FeatureLink{
		ID:        id,
		Category:  "category-" + id,
		Region:    "region-" + id,
		Reference: "seeded by feature.SampleFeatureLink",
		Count:     len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// FeatureNoteSchema is the stored shape of a FeatureNote.
var FeatureNoteSchema = schema.New("feature.featurenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureNote is one feature note.
type FeatureNote struct {
	ID       string
	Source   string
	Group    string
	Category string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureNote) Key() string { return "feature/featurenote/" + x.ID }

// Fields encodes x for storage.
func (x FeatureNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"group":    x.Group,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureNote decodes a stored record.
func ParseFeatureNote(f codec.Fields) (FeatureNote, error) {
	var x FeatureNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurenote: missing id")
	}
	x.Source = f["source"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("feature: featurenote %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featurenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureNote) String() string {
	return fmt.Sprintf("FeatureNote(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureNote) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureNote) Normalize() FeatureNote {
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortFeatureNotes orders by version descending, then by ID.
func SortFeatureNotes(xs []FeatureNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureNotes drops the archived entries, preserving order.
func ActiveFeatureNotes(xs []FeatureNote) []FeatureNote {
	out := make([]FeatureNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureNotes adds up the version of every active entry.
func SumFeatureNotes(xs []FeatureNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexFeatureNotes keys the entries by ID, last write winning.
func IndexFeatureNotes(xs []FeatureNote) map[string]FeatureNote {
	out := make(map[string]FeatureNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureNote builds a valid FeatureNote, for tests and for the seed data the cron
// binary writes.
func SampleFeatureNote(id string, at time.Time) FeatureNote {
	return FeatureNote{
		ID:       id,
		Source:   "source-" + id,
		Group:    "group-" + id,
		Category: "seeded by feature.SampleFeatureNote",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureWindowSchema is the stored shape of a FeatureWindow.
var FeatureWindowSchema = schema.New("feature.featurewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureWindow is one feature window.
type FeatureWindow struct {
	ID        string
	Region    string
	Reference string
	Name      string
	Capacity  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x FeatureWindow) Key() string { return "feature/featurewindow/" + x.ID }

// Fields encodes x for storage.
func (x FeatureWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"name":      x.Name,
		"capacity":  strconv.Itoa(x.Capacity),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureWindow decodes a stored record.
func ParseFeatureWindow(f codec.Fields) (FeatureWindow, error) {
	var x FeatureWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurewindow: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featurewindow %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("feature: featurewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureWindow) String() string {
	return fmt.Sprintf("FeatureWindow(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureWindow) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureWindow) Normalize() FeatureWindow {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortFeatureWindows orders by capacity descending, then by ID.
func SortFeatureWindows(xs []FeatureWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureWindows drops the archived entries, preserving order.
func ActiveFeatureWindows(xs []FeatureWindow) []FeatureWindow {
	out := make([]FeatureWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureWindows adds up the capacity of every active entry.
func SumFeatureWindows(xs []FeatureWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexFeatureWindows keys the entries by ID, last write winning.
func IndexFeatureWindows(xs []FeatureWindow) map[string]FeatureWindow {
	out := make(map[string]FeatureWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureWindow builds a valid FeatureWindow, for tests and for the seed data the cron
// binary writes.
func SampleFeatureWindow(id string, at time.Time) FeatureWindow {
	return FeatureWindow{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by feature.SampleFeatureWindow",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// FeatureQuotaSchema is the stored shape of a FeatureQuota.
var FeatureQuotaSchema = schema.New("feature.featurequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureQuota is one feature quota.
type FeatureQuota struct {
	ID       string
	Channel  string
	Owner    string
	Source   string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureQuota) Key() string { return "feature/featurequota/" + x.ID }

// Fields encodes x for storage.
func (x FeatureQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"source":   x.Source,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureQuota decodes a stored record.
func ParseFeatureQuota(f codec.Fields) (FeatureQuota, error) {
	var x FeatureQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurequota: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("feature: featurequota %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featurequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureQuota) String() string {
	return fmt.Sprintf("FeatureQuota(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureQuota) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureQuota) Normalize() FeatureQuota {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortFeatureQuotas orders by count descending, then by ID.
func SortFeatureQuotas(xs []FeatureQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureQuotas drops the archived entries, preserving order.
func ActiveFeatureQuotas(xs []FeatureQuota) []FeatureQuota {
	out := make([]FeatureQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureQuotas adds up the count of every active entry.
func SumFeatureQuotas(xs []FeatureQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexFeatureQuotas keys the entries by ID, last write winning.
func IndexFeatureQuotas(xs []FeatureQuota) map[string]FeatureQuota {
	out := make(map[string]FeatureQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureQuota builds a valid FeatureQuota, for tests and for the seed data the cron
// binary writes.
func SampleFeatureQuota(id string, at time.Time) FeatureQuota {
	return FeatureQuota{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by feature.SampleFeatureQuota",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureBucketSchema is the stored shape of a FeatureBucket.
var FeatureBucketSchema = schema.New("feature.featurebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureBucket is one feature bucket.
type FeatureBucket struct {
	ID       string
	Group    string
	Label    string
	Channel  string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureBucket) Key() string { return "feature/featurebucket/" + x.ID }

// Fields encodes x for storage.
func (x FeatureBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"label":    x.Label,
		"channel":  x.Channel,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureBucket decodes a stored record.
func ParseFeatureBucket(f codec.Fields) (FeatureBucket, error) {
	var x FeatureBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurebucket: missing id")
	}
	x.Group = f["group"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featurebucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featurebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureBucket) String() string {
	return fmt.Sprintf("FeatureBucket(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureBucket) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureBucket) Normalize() FeatureBucket {
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortFeatureBuckets orders by attempts descending, then by ID.
func SortFeatureBuckets(xs []FeatureBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureBuckets drops the archived entries, preserving order.
func ActiveFeatureBuckets(xs []FeatureBucket) []FeatureBucket {
	out := make([]FeatureBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureBuckets adds up the attempts of every active entry.
func SumFeatureBuckets(xs []FeatureBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexFeatureBuckets keys the entries by ID, last write winning.
func IndexFeatureBuckets(xs []FeatureBucket) map[string]FeatureBucket {
	out := make(map[string]FeatureBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureBucket builds a valid FeatureBucket, for tests and for the seed data the cron
// binary writes.
func SampleFeatureBucket(id string, at time.Time) FeatureBucket {
	return FeatureBucket{
		ID:       id,
		Group:    "group-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by feature.SampleFeatureBucket",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureLaneSchema is the stored shape of a FeatureLane.
var FeatureLaneSchema = schema.New("feature.featurelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureLane is one feature lane.
type FeatureLane struct {
	ID       string
	Comment  string
	Group    string
	Region   string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureLane) Key() string { return "feature/featurelane/" + x.ID }

// Fields encodes x for storage.
func (x FeatureLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"region":   x.Region,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureLane decodes a stored record.
func ParseFeatureLane(f codec.Fields) (FeatureLane, error) {
	var x FeatureLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurelane: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Region = f["region"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featurelane %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featurelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureLane) String() string {
	return fmt.Sprintf("FeatureLane(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureLane) Normalize() FeatureLane {
	x.Comment = strings.TrimSpace(x.Comment)
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

// SortFeatureLanes orders by attempts descending, then by ID.
func SortFeatureLanes(xs []FeatureLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureLanes drops the archived entries, preserving order.
func ActiveFeatureLanes(xs []FeatureLane) []FeatureLane {
	out := make([]FeatureLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureLanes adds up the attempts of every active entry.
func SumFeatureLanes(xs []FeatureLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexFeatureLanes keys the entries by ID, last write winning.
func IndexFeatureLanes(xs []FeatureLane) map[string]FeatureLane {
	out := make(map[string]FeatureLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureLane builds a valid FeatureLane, for tests and for the seed data the cron
// binary writes.
func SampleFeatureLane(id string, at time.Time) FeatureLane {
	return FeatureLane{
		ID:       id,
		Comment:  "comment-" + id,
		Group:    "group-" + id,
		Region:   "seeded by feature.SampleFeatureLane",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureSlotSchema is the stored shape of a FeatureSlot.
var FeatureSlotSchema = schema.New("feature.featureslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureSlot is one feature slot.
type FeatureSlot struct {
	ID       string
	Comment  string
	Owner    string
	Group    string
	Priority int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureSlot) Key() string { return "feature/featureslot/" + x.ID }

// Fields encodes x for storage.
func (x FeatureSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureSlot decodes a stored record.
func ParseFeatureSlot(f codec.Fields) (FeatureSlot, error) {
	var x FeatureSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featureslot: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("feature: featureslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("feature: featureslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featureslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featureslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureSlot) String() string {
	return fmt.Sprintf("FeatureSlot(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureSlot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureSlot) Normalize() FeatureSlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortFeatureSlots orders by priority descending, then by ID.
func SortFeatureSlots(xs []FeatureSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureSlots drops the archived entries, preserving order.
func ActiveFeatureSlots(xs []FeatureSlot) []FeatureSlot {
	out := make([]FeatureSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureSlots adds up the priority of every active entry.
func SumFeatureSlots(xs []FeatureSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexFeatureSlots keys the entries by ID, last write winning.
func IndexFeatureSlots(xs []FeatureSlot) map[string]FeatureSlot {
	out := make(map[string]FeatureSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureSlot builds a valid FeatureSlot, for tests and for the seed data the cron
// binary writes.
func SampleFeatureSlot(id string, at time.Time) FeatureSlot {
	return FeatureSlot{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by feature.SampleFeatureSlot",
		Priority: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureTagSchema is the stored shape of a FeatureTag.
var FeatureTagSchema = schema.New("feature.featuretag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureTag is one feature tag.
type FeatureTag struct {
	ID        string
	Reference string
	Name      string
	Category  string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x FeatureTag) Key() string { return "feature/featuretag/" + x.ID }

// Fields encodes x for storage.
func (x FeatureTag) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"name":      x.Name,
		"category":  x.Category,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureTag decodes a stored record.
func ParseFeatureTag(f codec.Fields) (FeatureTag, error) {
	var x FeatureTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featuretag: missing id")
	}
	x.Reference = f["reference"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("feature: featuretag %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("feature: featuretag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featuretag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featuretag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureTag) String() string {
	return fmt.Sprintf("FeatureTag(%s, reference=%q, offset=%d, archived=%t)", x.ID, x.Reference, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureTag) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureTag) Normalize() FeatureTag {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortFeatureTags orders by offset descending, then by ID.
func SortFeatureTags(xs []FeatureTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureTags drops the archived entries, preserving order.
func ActiveFeatureTags(xs []FeatureTag) []FeatureTag {
	out := make([]FeatureTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureTags adds up the offset of every active entry.
func SumFeatureTags(xs []FeatureTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexFeatureTags keys the entries by ID, last write winning.
func IndexFeatureTags(xs []FeatureTag) map[string]FeatureTag {
	out := make(map[string]FeatureTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureTag builds a valid FeatureTag, for tests and for the seed data the cron
// binary writes.
func SampleFeatureTag(id string, at time.Time) FeatureTag {
	return FeatureTag{
		ID:        id,
		Reference: "reference-" + id,
		Name:      "name-" + id,
		Category:  "seeded by feature.SampleFeatureTag",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// FeatureRuleSchema is the stored shape of a FeatureRule.
var FeatureRuleSchema = schema.New("feature.featurerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureRule is one feature rule.
type FeatureRule struct {
	ID        string
	Label     string
	Reference string
	Owner     string
	Count     int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x FeatureRule) Key() string { return "feature/featurerule/" + x.ID }

// Fields encodes x for storage.
func (x FeatureRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"reference": x.Reference,
		"owner":     x.Owner,
		"count":     strconv.Itoa(x.Count),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureRule decodes a stored record.
func ParseFeatureRule(f codec.Fields) (FeatureRule, error) {
	var x FeatureRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurerule: missing id")
	}
	x.Label = f["label"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("feature: featurerule %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("feature: featurerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureRule) String() string {
	return fmt.Sprintf("FeatureRule(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureRule) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureRule) Normalize() FeatureRule {
	x.Label = strings.TrimSpace(x.Label)
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

// SortFeatureRules orders by count descending, then by ID.
func SortFeatureRules(xs []FeatureRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureRules drops the archived entries, preserving order.
func ActiveFeatureRules(xs []FeatureRule) []FeatureRule {
	out := make([]FeatureRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureRules adds up the count of every active entry.
func SumFeatureRules(xs []FeatureRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexFeatureRules keys the entries by ID, last write winning.
func IndexFeatureRules(xs []FeatureRule) map[string]FeatureRule {
	out := make(map[string]FeatureRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureRule builds a valid FeatureRule, for tests and for the seed data the cron
// binary writes.
func SampleFeatureRule(id string, at time.Time) FeatureRule {
	return FeatureRule{
		ID:        id,
		Label:     "label-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by feature.SampleFeatureRule",
		Count:     len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// FeatureHoldSchema is the stored shape of a FeatureHold.
var FeatureHoldSchema = schema.New("feature.featurehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureHold is one feature hold.
type FeatureHold struct {
	ID       string
	Channel  string
	Owner    string
	Comment  string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureHold) Key() string { return "feature/featurehold/" + x.ID }

// Fields encodes x for storage.
func (x FeatureHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureHold decodes a stored record.
func ParseFeatureHold(f codec.Fields) (FeatureHold, error) {
	var x FeatureHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurehold: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("feature: featurehold %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featurehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureHold) String() string {
	return fmt.Sprintf("FeatureHold(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureHold) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureHold) Normalize() FeatureHold {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortFeatureHolds orders by priority descending, then by ID.
func SortFeatureHolds(xs []FeatureHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureHolds drops the archived entries, preserving order.
func ActiveFeatureHolds(xs []FeatureHold) []FeatureHold {
	out := make([]FeatureHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureHolds adds up the priority of every active entry.
func SumFeatureHolds(xs []FeatureHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexFeatureHolds keys the entries by ID, last write winning.
func IndexFeatureHolds(xs []FeatureHold) map[string]FeatureHold {
	out := make(map[string]FeatureHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureHold builds a valid FeatureHold, for tests and for the seed data the cron
// binary writes.
func SampleFeatureHold(id string, at time.Time) FeatureHold {
	return FeatureHold{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Comment:  "seeded by feature.SampleFeatureHold",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// FeatureClaimSchema is the stored shape of a FeatureClaim.
var FeatureClaimSchema = schema.New("feature.featureclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureClaim is one feature claim.
type FeatureClaim struct {
	ID        string
	Name      string
	Reference string
	Comment   string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x FeatureClaim) Key() string { return "feature/featureclaim/" + x.ID }

// Fields encodes x for storage.
func (x FeatureClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"comment":   x.Comment,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureClaim decodes a stored record.
func ParseFeatureClaim(f codec.Fields) (FeatureClaim, error) {
	var x FeatureClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featureclaim: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("feature: featureclaim %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("feature: featureclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featureclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featureclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureClaim) String() string {
	return fmt.Sprintf("FeatureClaim(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureClaim) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureClaim) Normalize() FeatureClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortFeatureClaims orders by weight descending, then by ID.
func SortFeatureClaims(xs []FeatureClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureClaims drops the archived entries, preserving order.
func ActiveFeatureClaims(xs []FeatureClaim) []FeatureClaim {
	out := make([]FeatureClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureClaims adds up the weight of every active entry.
func SumFeatureClaims(xs []FeatureClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexFeatureClaims keys the entries by ID, last write winning.
func IndexFeatureClaims(xs []FeatureClaim) map[string]FeatureClaim {
	out := make(map[string]FeatureClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureClaim builds a valid FeatureClaim, for tests and for the seed data the cron
// binary writes.
func SampleFeatureClaim(id string, at time.Time) FeatureClaim {
	return FeatureClaim{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by feature.SampleFeatureClaim",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// FeatureCursorSchema is the stored shape of a FeatureCursor.
var FeatureCursorSchema = schema.New("feature.featurecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// FeatureCursor is one feature cursor.
type FeatureCursor struct {
	ID       string
	Comment  string
	Owner    string
	Name     string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x FeatureCursor) Key() string { return "feature/featurecursor/" + x.ID }

// Fields encodes x for storage.
func (x FeatureCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseFeatureCursor decodes a stored record.
func ParseFeatureCursor(f codec.Fields) (FeatureCursor, error) {
	var x FeatureCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("feature: featurecursor: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("feature: featurecursor %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("feature: featurecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("feature: featurecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("feature: featurecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x FeatureCursor) String() string {
	return fmt.Sprintf("FeatureCursor(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x FeatureCursor) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x FeatureCursor) Normalize() FeatureCursor {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortFeatureCursors orders by count descending, then by ID.
func SortFeatureCursors(xs []FeatureCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveFeatureCursors drops the archived entries, preserving order.
func ActiveFeatureCursors(xs []FeatureCursor) []FeatureCursor {
	out := make([]FeatureCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumFeatureCursors adds up the count of every active entry.
func SumFeatureCursors(xs []FeatureCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexFeatureCursors keys the entries by ID, last write winning.
func IndexFeatureCursors(xs []FeatureCursor) map[string]FeatureCursor {
	out := make(map[string]FeatureCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleFeatureCursor builds a valid FeatureCursor, for tests and for the seed data the cron
// binary writes.
func SampleFeatureCursor(id string, at time.Time) FeatureCursor {
	return FeatureCursor{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by feature.SampleFeatureCursor",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	FeatureRecordSchema,
	FeatureEntrySchema,
	FeatureItemSchema,
	FeatureDraftSchema,
	FeatureRevisionSchema,
	FeatureSnapshotSchema,
	FeatureBatchSchema,
	FeatureRequestSchema,
	FeatureSummarySchema,
	FeaturePolicySchema,
	FeatureLinkSchema,
	FeatureNoteSchema,
	FeatureWindowSchema,
	FeatureQuotaSchema,
	FeatureBucketSchema,
	FeatureLaneSchema,
	FeatureSlotSchema,
	FeatureTagSchema,
	FeatureRuleSchema,
	FeatureHoldSchema,
	FeatureClaimSchema,
	FeatureCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"feature/featurerecord/",
	"feature/featureentry/",
	"feature/featureitem/",
	"feature/featuredraft/",
	"feature/featurerevision/",
	"feature/featuresnapshot/",
	"feature/featurebatch/",
	"feature/featurerequest/",
	"feature/featuresummary/",
	"feature/featurepolicy/",
	"feature/featurelink/",
	"feature/featurenote/",
	"feature/featurewindow/",
	"feature/featurequota/",
	"feature/featurebucket/",
	"feature/featurelane/",
	"feature/featureslot/",
	"feature/featuretag/",
	"feature/featurerule/",
	"feature/featurehold/",
	"feature/featureclaim/",
	"feature/featurecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "feature: " + strings.Join(names, ", ")
}
