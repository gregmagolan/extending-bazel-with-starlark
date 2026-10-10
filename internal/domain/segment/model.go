// Package segment holds the segment domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package segment

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// SegmentRecordSchema is the stored shape of a SegmentRecord.
var SegmentRecordSchema = schema.New("segment.segmentrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentRecord is one segment record.
type SegmentRecord struct {
	ID       string
	Owner    string
	Source   string
	Category string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentRecord) Key() string { return "segment/segmentrecord/" + x.ID }

// Fields encodes x for storage.
func (x SegmentRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentRecord decodes a stored record.
func ParseSegmentRecord(f codec.Fields) (SegmentRecord, error) {
	var x SegmentRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentrecord: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrecord %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentRecord) String() string {
	return fmt.Sprintf("SegmentRecord(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentRecord) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentRecord) Normalize() SegmentRecord {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortSegmentRecords orders by count descending, then by ID.
func SortSegmentRecords(xs []SegmentRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentRecords drops the archived entries, preserving order.
func ActiveSegmentRecords(xs []SegmentRecord) []SegmentRecord {
	out := make([]SegmentRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentRecords adds up the count of every active entry.
func SumSegmentRecords(xs []SegmentRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSegmentRecords keys the entries by ID, last write winning.
func IndexSegmentRecords(xs []SegmentRecord) map[string]SegmentRecord {
	out := make(map[string]SegmentRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentRecord builds a valid SegmentRecord, for tests and for the seed data the cron
// binary writes.
func SampleSegmentRecord(id string, at time.Time) SegmentRecord {
	return SegmentRecord{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Category: "seeded by segment.SampleSegmentRecord",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentEntrySchema is the stored shape of a SegmentEntry.
var SegmentEntrySchema = schema.New("segment.segmententry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentEntry is one segment entry.
type SegmentEntry struct {
	ID       string
	Label    string
	Category string
	Owner    string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentEntry) Key() string { return "segment/segmententry/" + x.ID }

// Fields encodes x for storage.
func (x SegmentEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentEntry decodes a stored record.
func ParseSegmentEntry(f codec.Fields) (SegmentEntry, error) {
	var x SegmentEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmententry: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("segment: segmententry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("segment: segmententry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmententry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmententry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentEntry) String() string {
	return fmt.Sprintf("SegmentEntry(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentEntry) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentEntry) Normalize() SegmentEntry {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSegmentEntrys orders by attempts descending, then by ID.
func SortSegmentEntrys(xs []SegmentEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentEntrys drops the archived entries, preserving order.
func ActiveSegmentEntrys(xs []SegmentEntry) []SegmentEntry {
	out := make([]SegmentEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentEntrys adds up the attempts of every active entry.
func SumSegmentEntrys(xs []SegmentEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexSegmentEntrys keys the entries by ID, last write winning.
func IndexSegmentEntrys(xs []SegmentEntry) map[string]SegmentEntry {
	out := make(map[string]SegmentEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentEntry builds a valid SegmentEntry, for tests and for the seed data the cron
// binary writes.
func SampleSegmentEntry(id string, at time.Time) SegmentEntry {
	return SegmentEntry{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Owner:    "seeded by segment.SampleSegmentEntry",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentItemSchema is the stored shape of a SegmentItem.
var SegmentItemSchema = schema.New("segment.segmentitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentItem is one segment item.
type SegmentItem struct {
	ID       string
	Label    string
	Region   string
	Name     string
	Priority int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentItem) Key() string { return "segment/segmentitem/" + x.ID }

// Fields encodes x for storage.
func (x SegmentItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"name":     x.Name,
		"priority": strconv.Itoa(x.Priority),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentItem decodes a stored record.
func ParseSegmentItem(f codec.Fields) (SegmentItem, error) {
	var x SegmentItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentitem: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmentitem %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("segment: segmentitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentItem) String() string {
	return fmt.Sprintf("SegmentItem(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentItem) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentItem) Normalize() SegmentItem {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSegmentItems orders by priority descending, then by ID.
func SortSegmentItems(xs []SegmentItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentItems drops the archived entries, preserving order.
func ActiveSegmentItems(xs []SegmentItem) []SegmentItem {
	out := make([]SegmentItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentItems adds up the priority of every active entry.
func SumSegmentItems(xs []SegmentItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSegmentItems keys the entries by ID, last write winning.
func IndexSegmentItems(xs []SegmentItem) map[string]SegmentItem {
	out := make(map[string]SegmentItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentItem builds a valid SegmentItem, for tests and for the seed data the cron
// binary writes.
func SampleSegmentItem(id string, at time.Time) SegmentItem {
	return SegmentItem{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Name:     "seeded by segment.SampleSegmentItem",
		Priority: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentDraftSchema is the stored shape of a SegmentDraft.
var SegmentDraftSchema = schema.New("segment.segmentdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentDraft is one segment draft.
type SegmentDraft struct {
	ID        string
	Region    string
	Comment   string
	Reference string
	Count     int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentDraft) Key() string { return "segment/segmentdraft/" + x.ID }

// Fields encodes x for storage.
func (x SegmentDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"comment":   x.Comment,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentDraft decodes a stored record.
func ParseSegmentDraft(f codec.Fields) (SegmentDraft, error) {
	var x SegmentDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentdraft: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("segment: segmentdraft %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("segment: segmentdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentDraft) String() string {
	return fmt.Sprintf("SegmentDraft(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentDraft) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentDraft) Normalize() SegmentDraft {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortSegmentDrafts orders by count descending, then by ID.
func SortSegmentDrafts(xs []SegmentDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentDrafts drops the archived entries, preserving order.
func ActiveSegmentDrafts(xs []SegmentDraft) []SegmentDraft {
	out := make([]SegmentDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentDrafts adds up the count of every active entry.
func SumSegmentDrafts(xs []SegmentDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSegmentDrafts keys the entries by ID, last write winning.
func IndexSegmentDrafts(xs []SegmentDraft) map[string]SegmentDraft {
	out := make(map[string]SegmentDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentDraft builds a valid SegmentDraft, for tests and for the seed data the cron
// binary writes.
func SampleSegmentDraft(id string, at time.Time) SegmentDraft {
	return SegmentDraft{
		ID:        id,
		Region:    "region-" + id,
		Comment:   "comment-" + id,
		Reference: "seeded by segment.SampleSegmentDraft",
		Count:     len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentRevisionSchema is the stored shape of a SegmentRevision.
var SegmentRevisionSchema = schema.New("segment.segmentrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentRevision is one segment revision.
type SegmentRevision struct {
	ID        string
	Name      string
	Group     string
	Reference string
	Version   int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentRevision) Key() string { return "segment/segmentrevision/" + x.ID }

// Fields encodes x for storage.
func (x SegmentRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"group":     x.Group,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentRevision decodes a stored record.
func ParseSegmentRevision(f codec.Fields) (SegmentRevision, error) {
	var x SegmentRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentrevision: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentRevision) String() string {
	return fmt.Sprintf("SegmentRevision(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentRevision) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentRevision) Normalize() SegmentRevision {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSegmentRevisions orders by version descending, then by ID.
func SortSegmentRevisions(xs []SegmentRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentRevisions drops the archived entries, preserving order.
func ActiveSegmentRevisions(xs []SegmentRevision) []SegmentRevision {
	out := make([]SegmentRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentRevisions adds up the version of every active entry.
func SumSegmentRevisions(xs []SegmentRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSegmentRevisions keys the entries by ID, last write winning.
func IndexSegmentRevisions(xs []SegmentRevision) map[string]SegmentRevision {
	out := make(map[string]SegmentRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentRevision builds a valid SegmentRevision, for tests and for the seed data the cron
// binary writes.
func SampleSegmentRevision(id string, at time.Time) SegmentRevision {
	return SegmentRevision{
		ID:        id,
		Name:      "name-" + id,
		Group:     "group-" + id,
		Reference: "seeded by segment.SampleSegmentRevision",
		Version:   len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentSnapshotSchema is the stored shape of a SegmentSnapshot.
var SegmentSnapshotSchema = schema.New("segment.segmentsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentSnapshot is one segment snapshot.
type SegmentSnapshot struct {
	ID        string
	Name      string
	Reference string
	Comment   string
	Count     int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentSnapshot) Key() string { return "segment/segmentsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x SegmentSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"comment":   x.Comment,
		"count":     strconv.Itoa(x.Count),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentSnapshot decodes a stored record.
func ParseSegmentSnapshot(f codec.Fields) (SegmentSnapshot, error) {
	var x SegmentSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentsnapshot: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("segment: segmentsnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("segment: segmentsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentSnapshot) String() string {
	return fmt.Sprintf("SegmentSnapshot(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentSnapshot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentSnapshot) Normalize() SegmentSnapshot {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSegmentSnapshots orders by count descending, then by ID.
func SortSegmentSnapshots(xs []SegmentSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentSnapshots drops the archived entries, preserving order.
func ActiveSegmentSnapshots(xs []SegmentSnapshot) []SegmentSnapshot {
	out := make([]SegmentSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentSnapshots adds up the count of every active entry.
func SumSegmentSnapshots(xs []SegmentSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSegmentSnapshots keys the entries by ID, last write winning.
func IndexSegmentSnapshots(xs []SegmentSnapshot) map[string]SegmentSnapshot {
	out := make(map[string]SegmentSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentSnapshot builds a valid SegmentSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleSegmentSnapshot(id string, at time.Time) SegmentSnapshot {
	return SegmentSnapshot{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by segment.SampleSegmentSnapshot",
		Count:     len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentBatchSchema is the stored shape of a SegmentBatch.
var SegmentBatchSchema = schema.New("segment.segmentbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentBatch is one segment batch.
type SegmentBatch struct {
	ID       string
	Source   string
	Category string
	Region   string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentBatch) Key() string { return "segment/segmentbatch/" + x.ID }

// Fields encodes x for storage.
func (x SegmentBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"category": x.Category,
		"region":   x.Region,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentBatch decodes a stored record.
func ParseSegmentBatch(f codec.Fields) (SegmentBatch, error) {
	var x SegmentBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentbatch: missing id")
	}
	x.Source = f["source"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("segment: segmentbatch %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("segment: segmentbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentBatch) String() string {
	return fmt.Sprintf("SegmentBatch(%s, source=%q, count=%d, archived=%t)", x.ID, x.Source, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentBatch) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentBatch) Normalize() SegmentBatch {
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSegmentBatchs orders by count descending, then by ID.
func SortSegmentBatchs(xs []SegmentBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentBatchs drops the archived entries, preserving order.
func ActiveSegmentBatchs(xs []SegmentBatch) []SegmentBatch {
	out := make([]SegmentBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentBatchs adds up the count of every active entry.
func SumSegmentBatchs(xs []SegmentBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSegmentBatchs keys the entries by ID, last write winning.
func IndexSegmentBatchs(xs []SegmentBatch) map[string]SegmentBatch {
	out := make(map[string]SegmentBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentBatch builds a valid SegmentBatch, for tests and for the seed data the cron
// binary writes.
func SampleSegmentBatch(id string, at time.Time) SegmentBatch {
	return SegmentBatch{
		ID:       id,
		Source:   "source-" + id,
		Category: "category-" + id,
		Region:   "seeded by segment.SampleSegmentBatch",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentRequestSchema is the stored shape of a SegmentRequest.
var SegmentRequestSchema = schema.New("segment.segmentrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentRequest is one segment request.
type SegmentRequest struct {
	ID        string
	Reference string
	Owner     string
	Group     string
	Count     int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentRequest) Key() string { return "segment/segmentrequest/" + x.ID }

// Fields encodes x for storage.
func (x SegmentRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"owner":     x.Owner,
		"group":     x.Group,
		"count":     strconv.Itoa(x.Count),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentRequest decodes a stored record.
func ParseSegmentRequest(f codec.Fields) (SegmentRequest, error) {
	var x SegmentRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentrequest: missing id")
	}
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentRequest) String() string {
	return fmt.Sprintf("SegmentRequest(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentRequest) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentRequest) Normalize() SegmentRequest {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSegmentRequests orders by count descending, then by ID.
func SortSegmentRequests(xs []SegmentRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentRequests drops the archived entries, preserving order.
func ActiveSegmentRequests(xs []SegmentRequest) []SegmentRequest {
	out := make([]SegmentRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentRequests adds up the count of every active entry.
func SumSegmentRequests(xs []SegmentRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSegmentRequests keys the entries by ID, last write winning.
func IndexSegmentRequests(xs []SegmentRequest) map[string]SegmentRequest {
	out := make(map[string]SegmentRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentRequest builds a valid SegmentRequest, for tests and for the seed data the cron
// binary writes.
func SampleSegmentRequest(id string, at time.Time) SegmentRequest {
	return SegmentRequest{
		ID:        id,
		Reference: "reference-" + id,
		Owner:     "owner-" + id,
		Group:     "seeded by segment.SampleSegmentRequest",
		Count:     len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentSummarySchema is the stored shape of a SegmentSummary.
var SegmentSummarySchema = schema.New("segment.segmentsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentSummary is one segment summary.
type SegmentSummary struct {
	ID        string
	Reference string
	Group     string
	Source    string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentSummary) Key() string { return "segment/segmentsummary/" + x.ID }

// Fields encodes x for storage.
func (x SegmentSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"source":    x.Source,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentSummary decodes a stored record.
func ParseSegmentSummary(f codec.Fields) (SegmentSummary, error) {
	var x SegmentSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentsummary: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("segment: segmentsummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("segment: segmentsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentSummary) String() string {
	return fmt.Sprintf("SegmentSummary(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentSummary) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentSummary) Normalize() SegmentSummary {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortSegmentSummarys orders by version descending, then by ID.
func SortSegmentSummarys(xs []SegmentSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentSummarys drops the archived entries, preserving order.
func ActiveSegmentSummarys(xs []SegmentSummary) []SegmentSummary {
	out := make([]SegmentSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentSummarys adds up the version of every active entry.
func SumSegmentSummarys(xs []SegmentSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSegmentSummarys keys the entries by ID, last write winning.
func IndexSegmentSummarys(xs []SegmentSummary) map[string]SegmentSummary {
	out := make(map[string]SegmentSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentSummary builds a valid SegmentSummary, for tests and for the seed data the cron
// binary writes.
func SampleSegmentSummary(id string, at time.Time) SegmentSummary {
	return SegmentSummary{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Source:    "seeded by segment.SampleSegmentSummary",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentPolicySchema is the stored shape of a SegmentPolicy.
var SegmentPolicySchema = schema.New("segment.segmentpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentPolicy is one segment policy.
type SegmentPolicy struct {
	ID       string
	Group    string
	Name     string
	Source   string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentPolicy) Key() string { return "segment/segmentpolicy/" + x.ID }

// Fields encodes x for storage.
func (x SegmentPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentPolicy decodes a stored record.
func ParseSegmentPolicy(f codec.Fields) (SegmentPolicy, error) {
	var x SegmentPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentpolicy: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmentpolicy %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("segment: segmentpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentPolicy) String() string {
	return fmt.Sprintf("SegmentPolicy(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentPolicy) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentPolicy) Normalize() SegmentPolicy {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortSegmentPolicys orders by priority descending, then by ID.
func SortSegmentPolicys(xs []SegmentPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentPolicys drops the archived entries, preserving order.
func ActiveSegmentPolicys(xs []SegmentPolicy) []SegmentPolicy {
	out := make([]SegmentPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentPolicys adds up the priority of every active entry.
func SumSegmentPolicys(xs []SegmentPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSegmentPolicys keys the entries by ID, last write winning.
func IndexSegmentPolicys(xs []SegmentPolicy) map[string]SegmentPolicy {
	out := make(map[string]SegmentPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentPolicy builds a valid SegmentPolicy, for tests and for the seed data the cron
// binary writes.
func SampleSegmentPolicy(id string, at time.Time) SegmentPolicy {
	return SegmentPolicy{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Source:   "seeded by segment.SampleSegmentPolicy",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentLinkSchema is the stored shape of a SegmentLink.
var SegmentLinkSchema = schema.New("segment.segmentlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentLink is one segment link.
type SegmentLink struct {
	ID       string
	Region   string
	Group    string
	Source   string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentLink) Key() string { return "segment/segmentlink/" + x.ID }

// Fields encodes x for storage.
func (x SegmentLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"group":    x.Group,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentLink decodes a stored record.
func ParseSegmentLink(f codec.Fields) (SegmentLink, error) {
	var x SegmentLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentlink: missing id")
	}
	x.Region = f["region"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmentlink %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("segment: segmentlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentLink) String() string {
	return fmt.Sprintf("SegmentLink(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentLink) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentLink) Normalize() SegmentLink {
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortSegmentLinks orders by priority descending, then by ID.
func SortSegmentLinks(xs []SegmentLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentLinks drops the archived entries, preserving order.
func ActiveSegmentLinks(xs []SegmentLink) []SegmentLink {
	out := make([]SegmentLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentLinks adds up the priority of every active entry.
func SumSegmentLinks(xs []SegmentLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSegmentLinks keys the entries by ID, last write winning.
func IndexSegmentLinks(xs []SegmentLink) map[string]SegmentLink {
	out := make(map[string]SegmentLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentLink builds a valid SegmentLink, for tests and for the seed data the cron
// binary writes.
func SampleSegmentLink(id string, at time.Time) SegmentLink {
	return SegmentLink{
		ID:       id,
		Region:   "region-" + id,
		Group:    "group-" + id,
		Source:   "seeded by segment.SampleSegmentLink",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentNoteSchema is the stored shape of a SegmentNote.
var SegmentNoteSchema = schema.New("segment.segmentnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentNote is one segment note.
type SegmentNote struct {
	ID       string
	Source   string
	Name     string
	Comment  string
	Priority int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentNote) Key() string { return "segment/segmentnote/" + x.ID }

// Fields encodes x for storage.
func (x SegmentNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentNote decodes a stored record.
func ParseSegmentNote(f codec.Fields) (SegmentNote, error) {
	var x SegmentNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentnote: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmentnote %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("segment: segmentnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentNote) String() string {
	return fmt.Sprintf("SegmentNote(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentNote) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentNote) Normalize() SegmentNote {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSegmentNotes orders by priority descending, then by ID.
func SortSegmentNotes(xs []SegmentNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentNotes drops the archived entries, preserving order.
func ActiveSegmentNotes(xs []SegmentNote) []SegmentNote {
	out := make([]SegmentNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentNotes adds up the priority of every active entry.
func SumSegmentNotes(xs []SegmentNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSegmentNotes keys the entries by ID, last write winning.
func IndexSegmentNotes(xs []SegmentNote) map[string]SegmentNote {
	out := make(map[string]SegmentNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentNote builds a valid SegmentNote, for tests and for the seed data the cron
// binary writes.
func SampleSegmentNote(id string, at time.Time) SegmentNote {
	return SegmentNote{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by segment.SampleSegmentNote",
		Priority: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentWindowSchema is the stored shape of a SegmentWindow.
var SegmentWindowSchema = schema.New("segment.segmentwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentWindow is one segment window.
type SegmentWindow struct {
	ID       string
	Owner    string
	Category string
	Source   string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentWindow) Key() string { return "segment/segmentwindow/" + x.ID }

// Fields encodes x for storage.
func (x SegmentWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"category": x.Category,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentWindow decodes a stored record.
func ParseSegmentWindow(f codec.Fields) (SegmentWindow, error) {
	var x SegmentWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentwindow: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("segment: segmentwindow %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("segment: segmentwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentWindow) String() string {
	return fmt.Sprintf("SegmentWindow(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentWindow) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentWindow) Normalize() SegmentWindow {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSegmentWindows orders by offset descending, then by ID.
func SortSegmentWindows(xs []SegmentWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentWindows drops the archived entries, preserving order.
func ActiveSegmentWindows(xs []SegmentWindow) []SegmentWindow {
	out := make([]SegmentWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentWindows adds up the offset of every active entry.
func SumSegmentWindows(xs []SegmentWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSegmentWindows keys the entries by ID, last write winning.
func IndexSegmentWindows(xs []SegmentWindow) map[string]SegmentWindow {
	out := make(map[string]SegmentWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentWindow builds a valid SegmentWindow, for tests and for the seed data the cron
// binary writes.
func SampleSegmentWindow(id string, at time.Time) SegmentWindow {
	return SegmentWindow{
		ID:       id,
		Owner:    "owner-" + id,
		Category: "category-" + id,
		Source:   "seeded by segment.SampleSegmentWindow",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentQuotaSchema is the stored shape of a SegmentQuota.
var SegmentQuotaSchema = schema.New("segment.segmentquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentQuota is one segment quota.
type SegmentQuota struct {
	ID       string
	Category string
	Channel  string
	Group    string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentQuota) Key() string { return "segment/segmentquota/" + x.ID }

// Fields encodes x for storage.
func (x SegmentQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentQuota decodes a stored record.
func ParseSegmentQuota(f codec.Fields) (SegmentQuota, error) {
	var x SegmentQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentquota: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("segment: segmentquota %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("segment: segmentquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentQuota) String() string {
	return fmt.Sprintf("SegmentQuota(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentQuota) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentQuota) Normalize() SegmentQuota {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSegmentQuotas orders by capacity descending, then by ID.
func SortSegmentQuotas(xs []SegmentQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentQuotas drops the archived entries, preserving order.
func ActiveSegmentQuotas(xs []SegmentQuota) []SegmentQuota {
	out := make([]SegmentQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentQuotas adds up the capacity of every active entry.
func SumSegmentQuotas(xs []SegmentQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSegmentQuotas keys the entries by ID, last write winning.
func IndexSegmentQuotas(xs []SegmentQuota) map[string]SegmentQuota {
	out := make(map[string]SegmentQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentQuota builds a valid SegmentQuota, for tests and for the seed data the cron
// binary writes.
func SampleSegmentQuota(id string, at time.Time) SegmentQuota {
	return SegmentQuota{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Group:    "seeded by segment.SampleSegmentQuota",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentBucketSchema is the stored shape of a SegmentBucket.
var SegmentBucketSchema = schema.New("segment.segmentbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentBucket is one segment bucket.
type SegmentBucket struct {
	ID       string
	Category string
	Group    string
	Source   string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentBucket) Key() string { return "segment/segmentbucket/" + x.ID }

// Fields encodes x for storage.
func (x SegmentBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"group":    x.Group,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentBucket decodes a stored record.
func ParseSegmentBucket(f codec.Fields) (SegmentBucket, error) {
	var x SegmentBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentbucket: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmentbucket %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("segment: segmentbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentBucket) String() string {
	return fmt.Sprintf("SegmentBucket(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentBucket) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentBucket) Normalize() SegmentBucket {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortSegmentBuckets orders by priority descending, then by ID.
func SortSegmentBuckets(xs []SegmentBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentBuckets drops the archived entries, preserving order.
func ActiveSegmentBuckets(xs []SegmentBucket) []SegmentBucket {
	out := make([]SegmentBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentBuckets adds up the priority of every active entry.
func SumSegmentBuckets(xs []SegmentBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSegmentBuckets keys the entries by ID, last write winning.
func IndexSegmentBuckets(xs []SegmentBucket) map[string]SegmentBucket {
	out := make(map[string]SegmentBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentBucket builds a valid SegmentBucket, for tests and for the seed data the cron
// binary writes.
func SampleSegmentBucket(id string, at time.Time) SegmentBucket {
	return SegmentBucket{
		ID:       id,
		Category: "category-" + id,
		Group:    "group-" + id,
		Source:   "seeded by segment.SampleSegmentBucket",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentLaneSchema is the stored shape of a SegmentLane.
var SegmentLaneSchema = schema.New("segment.segmentlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentLane is one segment lane.
type SegmentLane struct {
	ID        string
	Reference string
	Comment   string
	Name      string
	Attempts  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentLane) Key() string { return "segment/segmentlane/" + x.ID }

// Fields encodes x for storage.
func (x SegmentLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"comment":   x.Comment,
		"name":      x.Name,
		"attempts":  strconv.Itoa(x.Attempts),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentLane decodes a stored record.
func ParseSegmentLane(f codec.Fields) (SegmentLane, error) {
	var x SegmentLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentlane: missing id")
	}
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("segment: segmentlane %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("segment: segmentlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentLane) String() string {
	return fmt.Sprintf("SegmentLane(%s, reference=%q, attempts=%d, archived=%t)", x.ID, x.Reference, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentLane) Normalize() SegmentLane {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSegmentLanes orders by attempts descending, then by ID.
func SortSegmentLanes(xs []SegmentLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentLanes drops the archived entries, preserving order.
func ActiveSegmentLanes(xs []SegmentLane) []SegmentLane {
	out := make([]SegmentLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentLanes adds up the attempts of every active entry.
func SumSegmentLanes(xs []SegmentLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexSegmentLanes keys the entries by ID, last write winning.
func IndexSegmentLanes(xs []SegmentLane) map[string]SegmentLane {
	out := make(map[string]SegmentLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentLane builds a valid SegmentLane, for tests and for the seed data the cron
// binary writes.
func SampleSegmentLane(id string, at time.Time) SegmentLane {
	return SegmentLane{
		ID:        id,
		Reference: "reference-" + id,
		Comment:   "comment-" + id,
		Name:      "seeded by segment.SampleSegmentLane",
		Attempts:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentSlotSchema is the stored shape of a SegmentSlot.
var SegmentSlotSchema = schema.New("segment.segmentslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentSlot is one segment slot.
type SegmentSlot struct {
	ID       string
	Category string
	Region   string
	Channel  string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentSlot) Key() string { return "segment/segmentslot/" + x.ID }

// Fields encodes x for storage.
func (x SegmentSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentSlot decodes a stored record.
func ParseSegmentSlot(f codec.Fields) (SegmentSlot, error) {
	var x SegmentSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentslot: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("segment: segmentslot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmentslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentSlot) String() string {
	return fmt.Sprintf("SegmentSlot(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentSlot) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentSlot) Normalize() SegmentSlot {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSegmentSlots orders by capacity descending, then by ID.
func SortSegmentSlots(xs []SegmentSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentSlots drops the archived entries, preserving order.
func ActiveSegmentSlots(xs []SegmentSlot) []SegmentSlot {
	out := make([]SegmentSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentSlots adds up the capacity of every active entry.
func SumSegmentSlots(xs []SegmentSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSegmentSlots keys the entries by ID, last write winning.
func IndexSegmentSlots(xs []SegmentSlot) map[string]SegmentSlot {
	out := make(map[string]SegmentSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentSlot builds a valid SegmentSlot, for tests and for the seed data the cron
// binary writes.
func SampleSegmentSlot(id string, at time.Time) SegmentSlot {
	return SegmentSlot{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Channel:  "seeded by segment.SampleSegmentSlot",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentTagSchema is the stored shape of a SegmentTag.
var SegmentTagSchema = schema.New("segment.segmenttag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentTag is one segment tag.
type SegmentTag struct {
	ID       string
	Category string
	Region   string
	Source   string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentTag) Key() string { return "segment/segmenttag/" + x.ID }

// Fields encodes x for storage.
func (x SegmentTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentTag decodes a stored record.
func ParseSegmentTag(f codec.Fields) (SegmentTag, error) {
	var x SegmentTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmenttag: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("segment: segmenttag %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("segment: segmenttag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmenttag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmenttag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentTag) String() string {
	return fmt.Sprintf("SegmentTag(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentTag) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentTag) Normalize() SegmentTag {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSegmentTags orders by retained descending, then by ID.
func SortSegmentTags(xs []SegmentTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentTags drops the archived entries, preserving order.
func ActiveSegmentTags(xs []SegmentTag) []SegmentTag {
	out := make([]SegmentTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentTags adds up the retained of every active entry.
func SumSegmentTags(xs []SegmentTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexSegmentTags keys the entries by ID, last write winning.
func IndexSegmentTags(xs []SegmentTag) map[string]SegmentTag {
	out := make(map[string]SegmentTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentTag builds a valid SegmentTag, for tests and for the seed data the cron
// binary writes.
func SampleSegmentTag(id string, at time.Time) SegmentTag {
	return SegmentTag{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Source:   "seeded by segment.SampleSegmentTag",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentRuleSchema is the stored shape of a SegmentRule.
var SegmentRuleSchema = schema.New("segment.segmentrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentRule is one segment rule.
type SegmentRule struct {
	ID        string
	Source    string
	Comment   string
	Reference string
	Attempts  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentRule) Key() string { return "segment/segmentrule/" + x.ID }

// Fields encodes x for storage.
func (x SegmentRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"comment":   x.Comment,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentRule decodes a stored record.
func ParseSegmentRule(f codec.Fields) (SegmentRule, error) {
	var x SegmentRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentrule: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrule %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("segment: segmentrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentRule) String() string {
	return fmt.Sprintf("SegmentRule(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentRule) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentRule) Normalize() SegmentRule {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSegmentRules orders by attempts descending, then by ID.
func SortSegmentRules(xs []SegmentRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentRules drops the archived entries, preserving order.
func ActiveSegmentRules(xs []SegmentRule) []SegmentRule {
	out := make([]SegmentRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentRules adds up the attempts of every active entry.
func SumSegmentRules(xs []SegmentRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexSegmentRules keys the entries by ID, last write winning.
func IndexSegmentRules(xs []SegmentRule) map[string]SegmentRule {
	out := make(map[string]SegmentRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentRule builds a valid SegmentRule, for tests and for the seed data the cron
// binary writes.
func SampleSegmentRule(id string, at time.Time) SegmentRule {
	return SegmentRule{
		ID:        id,
		Source:    "source-" + id,
		Comment:   "comment-" + id,
		Reference: "seeded by segment.SampleSegmentRule",
		Attempts:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentHoldSchema is the stored shape of a SegmentHold.
var SegmentHoldSchema = schema.New("segment.segmenthold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentHold is one segment hold.
type SegmentHold struct {
	ID        string
	Name      string
	Label     string
	Reference string
	Version   int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentHold) Key() string { return "segment/segmenthold/" + x.ID }

// Fields encodes x for storage.
func (x SegmentHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"label":     x.Label,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentHold decodes a stored record.
func ParseSegmentHold(f codec.Fields) (SegmentHold, error) {
	var x SegmentHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmenthold: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("segment: segmenthold %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmenthold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmenthold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmenthold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentHold) String() string {
	return fmt.Sprintf("SegmentHold(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentHold) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentHold) Normalize() SegmentHold {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSegmentHolds orders by version descending, then by ID.
func SortSegmentHolds(xs []SegmentHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentHolds drops the archived entries, preserving order.
func ActiveSegmentHolds(xs []SegmentHold) []SegmentHold {
	out := make([]SegmentHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentHolds adds up the version of every active entry.
func SumSegmentHolds(xs []SegmentHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSegmentHolds keys the entries by ID, last write winning.
func IndexSegmentHolds(xs []SegmentHold) map[string]SegmentHold {
	out := make(map[string]SegmentHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentHold builds a valid SegmentHold, for tests and for the seed data the cron
// binary writes.
func SampleSegmentHold(id string, at time.Time) SegmentHold {
	return SegmentHold{
		ID:        id,
		Name:      "name-" + id,
		Label:     "label-" + id,
		Reference: "seeded by segment.SampleSegmentHold",
		Version:   len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SegmentClaimSchema is the stored shape of a SegmentClaim.
var SegmentClaimSchema = schema.New("segment.segmentclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentClaim is one segment claim.
type SegmentClaim struct {
	ID       string
	Label    string
	Channel  string
	Region   string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SegmentClaim) Key() string { return "segment/segmentclaim/" + x.ID }

// Fields encodes x for storage.
func (x SegmentClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentClaim decodes a stored record.
func ParseSegmentClaim(f codec.Fields) (SegmentClaim, error) {
	var x SegmentClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentclaim: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("segment: segmentclaim %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("segment: segmentclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentClaim) String() string {
	return fmt.Sprintf("SegmentClaim(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentClaim) Normalize() SegmentClaim {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSegmentClaims orders by retained descending, then by ID.
func SortSegmentClaims(xs []SegmentClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentClaims drops the archived entries, preserving order.
func ActiveSegmentClaims(xs []SegmentClaim) []SegmentClaim {
	out := make([]SegmentClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentClaims adds up the retained of every active entry.
func SumSegmentClaims(xs []SegmentClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexSegmentClaims keys the entries by ID, last write winning.
func IndexSegmentClaims(xs []SegmentClaim) map[string]SegmentClaim {
	out := make(map[string]SegmentClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentClaim builds a valid SegmentClaim, for tests and for the seed data the cron
// binary writes.
func SampleSegmentClaim(id string, at time.Time) SegmentClaim {
	return SegmentClaim{
		ID:       id,
		Label:    "label-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by segment.SampleSegmentClaim",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SegmentCursorSchema is the stored shape of a SegmentCursor.
var SegmentCursorSchema = schema.New("segment.segmentcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SegmentCursor is one segment cursor.
type SegmentCursor struct {
	ID        string
	Reference string
	Region    string
	Owner     string
	Priority  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SegmentCursor) Key() string { return "segment/segmentcursor/" + x.ID }

// Fields encodes x for storage.
func (x SegmentCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"owner":     x.Owner,
		"priority":  strconv.Itoa(x.Priority),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSegmentCursor decodes a stored record.
func ParseSegmentCursor(f codec.Fields) (SegmentCursor, error) {
	var x SegmentCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("segment: segmentcursor: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("segment: segmentcursor %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("segment: segmentcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("segment: segmentcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("segment: segmentcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SegmentCursor) String() string {
	return fmt.Sprintf("SegmentCursor(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SegmentCursor) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SegmentCursor) Normalize() SegmentCursor {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSegmentCursors orders by priority descending, then by ID.
func SortSegmentCursors(xs []SegmentCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSegmentCursors drops the archived entries, preserving order.
func ActiveSegmentCursors(xs []SegmentCursor) []SegmentCursor {
	out := make([]SegmentCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSegmentCursors adds up the priority of every active entry.
func SumSegmentCursors(xs []SegmentCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSegmentCursors keys the entries by ID, last write winning.
func IndexSegmentCursors(xs []SegmentCursor) map[string]SegmentCursor {
	out := make(map[string]SegmentCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSegmentCursor builds a valid SegmentCursor, for tests and for the seed data the cron
// binary writes.
func SampleSegmentCursor(id string, at time.Time) SegmentCursor {
	return SegmentCursor{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Owner:     "seeded by segment.SampleSegmentCursor",
		Priority:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	SegmentRecordSchema,
	SegmentEntrySchema,
	SegmentItemSchema,
	SegmentDraftSchema,
	SegmentRevisionSchema,
	SegmentSnapshotSchema,
	SegmentBatchSchema,
	SegmentRequestSchema,
	SegmentSummarySchema,
	SegmentPolicySchema,
	SegmentLinkSchema,
	SegmentNoteSchema,
	SegmentWindowSchema,
	SegmentQuotaSchema,
	SegmentBucketSchema,
	SegmentLaneSchema,
	SegmentSlotSchema,
	SegmentTagSchema,
	SegmentRuleSchema,
	SegmentHoldSchema,
	SegmentClaimSchema,
	SegmentCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"segment/segmentrecord/",
	"segment/segmententry/",
	"segment/segmentitem/",
	"segment/segmentdraft/",
	"segment/segmentrevision/",
	"segment/segmentsnapshot/",
	"segment/segmentbatch/",
	"segment/segmentrequest/",
	"segment/segmentsummary/",
	"segment/segmentpolicy/",
	"segment/segmentlink/",
	"segment/segmentnote/",
	"segment/segmentwindow/",
	"segment/segmentquota/",
	"segment/segmentbucket/",
	"segment/segmentlane/",
	"segment/segmentslot/",
	"segment/segmenttag/",
	"segment/segmentrule/",
	"segment/segmenthold/",
	"segment/segmentclaim/",
	"segment/segmentcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "segment: " + strings.Join(names, ", ")
}
