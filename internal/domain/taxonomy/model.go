// Package taxonomy holds the taxonomy domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package taxonomy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// TaxonomyRecordSchema is the stored shape of a TaxonomyRecord.
var TaxonomyRecordSchema = schema.New("taxonomy.taxonomyrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyRecord is one taxonomy record.
type TaxonomyRecord struct {
	ID       string
	Channel  string
	Region   string
	Label    string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyRecord) Key() string { return "taxonomy/taxonomyrecord/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyRecord decodes a stored record.
func ParseTaxonomyRecord(f codec.Fields) (TaxonomyRecord, error) {
	var x TaxonomyRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyrecord: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrecord %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyRecord) String() string {
	return fmt.Sprintf("TaxonomyRecord(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyRecord) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyRecord) Normalize() TaxonomyRecord {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortTaxonomyRecords orders by attempts descending, then by ID.
func SortTaxonomyRecords(xs []TaxonomyRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyRecords drops the archived entries, preserving order.
func ActiveTaxonomyRecords(xs []TaxonomyRecord) []TaxonomyRecord {
	out := make([]TaxonomyRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyRecords adds up the attempts of every active entry.
func SumTaxonomyRecords(xs []TaxonomyRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTaxonomyRecords keys the entries by ID, last write winning.
func IndexTaxonomyRecords(xs []TaxonomyRecord) map[string]TaxonomyRecord {
	out := make(map[string]TaxonomyRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyRecord builds a valid TaxonomyRecord, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyRecord(id string, at time.Time) TaxonomyRecord {
	return TaxonomyRecord{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Label:    "seeded by taxonomy.SampleTaxonomyRecord",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyEntrySchema is the stored shape of a TaxonomyEntry.
var TaxonomyEntrySchema = schema.New("taxonomy.taxonomyentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyEntry is one taxonomy entry.
type TaxonomyEntry struct {
	ID        string
	Name      string
	Comment   string
	Reference string
	Offset    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TaxonomyEntry) Key() string { return "taxonomy/taxonomyentry/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"comment":   x.Comment,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyEntry decodes a stored record.
func ParseTaxonomyEntry(f codec.Fields) (TaxonomyEntry, error) {
	var x TaxonomyEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyentry: missing id")
	}
	x.Name = f["name"]
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyentry %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyEntry) String() string {
	return fmt.Sprintf("TaxonomyEntry(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyEntry) Normalize() TaxonomyEntry {
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortTaxonomyEntrys orders by offset descending, then by ID.
func SortTaxonomyEntrys(xs []TaxonomyEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyEntrys drops the archived entries, preserving order.
func ActiveTaxonomyEntrys(xs []TaxonomyEntry) []TaxonomyEntry {
	out := make([]TaxonomyEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyEntrys adds up the offset of every active entry.
func SumTaxonomyEntrys(xs []TaxonomyEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTaxonomyEntrys keys the entries by ID, last write winning.
func IndexTaxonomyEntrys(xs []TaxonomyEntry) map[string]TaxonomyEntry {
	out := make(map[string]TaxonomyEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyEntry builds a valid TaxonomyEntry, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyEntry(id string, at time.Time) TaxonomyEntry {
	return TaxonomyEntry{
		ID:        id,
		Name:      "name-" + id,
		Comment:   "comment-" + id,
		Reference: "seeded by taxonomy.SampleTaxonomyEntry",
		Offset:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TaxonomyItemSchema is the stored shape of a TaxonomyItem.
var TaxonomyItemSchema = schema.New("taxonomy.taxonomyitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyItem is one taxonomy item.
type TaxonomyItem struct {
	ID       string
	Owner    string
	Source   string
	Comment  string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyItem) Key() string { return "taxonomy/taxonomyitem/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyItem decodes a stored record.
func ParseTaxonomyItem(f codec.Fields) (TaxonomyItem, error) {
	var x TaxonomyItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyitem: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyItem) String() string {
	return fmt.Sprintf("TaxonomyItem(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyItem) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyItem) Normalize() TaxonomyItem {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTaxonomyItems orders by priority descending, then by ID.
func SortTaxonomyItems(xs []TaxonomyItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyItems drops the archived entries, preserving order.
func ActiveTaxonomyItems(xs []TaxonomyItem) []TaxonomyItem {
	out := make([]TaxonomyItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyItems adds up the priority of every active entry.
func SumTaxonomyItems(xs []TaxonomyItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTaxonomyItems keys the entries by ID, last write winning.
func IndexTaxonomyItems(xs []TaxonomyItem) map[string]TaxonomyItem {
	out := make(map[string]TaxonomyItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyItem builds a valid TaxonomyItem, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyItem(id string, at time.Time) TaxonomyItem {
	return TaxonomyItem{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Comment:  "seeded by taxonomy.SampleTaxonomyItem",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyDraftSchema is the stored shape of a TaxonomyDraft.
var TaxonomyDraftSchema = schema.New("taxonomy.taxonomydraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyDraft is one taxonomy draft.
type TaxonomyDraft struct {
	ID       string
	Name     string
	Owner    string
	Group    string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyDraft) Key() string { return "taxonomy/taxonomydraft/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyDraft decodes a stored record.
func ParseTaxonomyDraft(f codec.Fields) (TaxonomyDraft, error) {
	var x TaxonomyDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomydraft: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomydraft %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomydraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomydraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomydraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyDraft) String() string {
	return fmt.Sprintf("TaxonomyDraft(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyDraft) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyDraft) Normalize() TaxonomyDraft {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTaxonomyDrafts orders by offset descending, then by ID.
func SortTaxonomyDrafts(xs []TaxonomyDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyDrafts drops the archived entries, preserving order.
func ActiveTaxonomyDrafts(xs []TaxonomyDraft) []TaxonomyDraft {
	out := make([]TaxonomyDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyDrafts adds up the offset of every active entry.
func SumTaxonomyDrafts(xs []TaxonomyDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTaxonomyDrafts keys the entries by ID, last write winning.
func IndexTaxonomyDrafts(xs []TaxonomyDraft) map[string]TaxonomyDraft {
	out := make(map[string]TaxonomyDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyDraft builds a valid TaxonomyDraft, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyDraft(id string, at time.Time) TaxonomyDraft {
	return TaxonomyDraft{
		ID:      id,
		Name:    "name-" + id,
		Owner:   "owner-" + id,
		Group:   "seeded by taxonomy.SampleTaxonomyDraft",
		Offset:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TaxonomyRevisionSchema is the stored shape of a TaxonomyRevision.
var TaxonomyRevisionSchema = schema.New("taxonomy.taxonomyrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyRevision is one taxonomy revision.
type TaxonomyRevision struct {
	ID       string
	Region   string
	Comment  string
	Source   string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyRevision) Key() string { return "taxonomy/taxonomyrevision/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyRevision decodes a stored record.
func ParseTaxonomyRevision(f codec.Fields) (TaxonomyRevision, error) {
	var x TaxonomyRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyrevision: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrevision %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyRevision) String() string {
	return fmt.Sprintf("TaxonomyRevision(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyRevision) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyRevision) Normalize() TaxonomyRevision {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortTaxonomyRevisions orders by retained descending, then by ID.
func SortTaxonomyRevisions(xs []TaxonomyRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyRevisions drops the archived entries, preserving order.
func ActiveTaxonomyRevisions(xs []TaxonomyRevision) []TaxonomyRevision {
	out := make([]TaxonomyRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyRevisions adds up the retained of every active entry.
func SumTaxonomyRevisions(xs []TaxonomyRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTaxonomyRevisions keys the entries by ID, last write winning.
func IndexTaxonomyRevisions(xs []TaxonomyRevision) map[string]TaxonomyRevision {
	out := make(map[string]TaxonomyRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyRevision builds a valid TaxonomyRevision, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyRevision(id string, at time.Time) TaxonomyRevision {
	return TaxonomyRevision{
		ID:       id,
		Region:   "region-" + id,
		Comment:  "comment-" + id,
		Source:   "seeded by taxonomy.SampleTaxonomyRevision",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomySnapshotSchema is the stored shape of a TaxonomySnapshot.
var TaxonomySnapshotSchema = schema.New("taxonomy.taxonomysnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomySnapshot is one taxonomy snapshot.
type TaxonomySnapshot struct {
	ID        string
	Comment   string
	Reference string
	Channel   string
	Priority  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TaxonomySnapshot) Key() string { return "taxonomy/taxonomysnapshot/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomySnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"channel":   x.Channel,
		"priority":  strconv.Itoa(x.Priority),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomySnapshot decodes a stored record.
func ParseTaxonomySnapshot(f codec.Fields) (TaxonomySnapshot, error) {
	var x TaxonomySnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomysnapshot: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomysnapshot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomysnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomysnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomysnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomySnapshot) String() string {
	return fmt.Sprintf("TaxonomySnapshot(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomySnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomySnapshot) Normalize() TaxonomySnapshot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTaxonomySnapshots orders by priority descending, then by ID.
func SortTaxonomySnapshots(xs []TaxonomySnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomySnapshots drops the archived entries, preserving order.
func ActiveTaxonomySnapshots(xs []TaxonomySnapshot) []TaxonomySnapshot {
	out := make([]TaxonomySnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomySnapshots adds up the priority of every active entry.
func SumTaxonomySnapshots(xs []TaxonomySnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTaxonomySnapshots keys the entries by ID, last write winning.
func IndexTaxonomySnapshots(xs []TaxonomySnapshot) map[string]TaxonomySnapshot {
	out := make(map[string]TaxonomySnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomySnapshot builds a valid TaxonomySnapshot, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomySnapshot(id string, at time.Time) TaxonomySnapshot {
	return TaxonomySnapshot{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by taxonomy.SampleTaxonomySnapshot",
		Priority:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TaxonomyBatchSchema is the stored shape of a TaxonomyBatch.
var TaxonomyBatchSchema = schema.New("taxonomy.taxonomybatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyBatch is one taxonomy batch.
type TaxonomyBatch struct {
	ID       string
	Category string
	Owner    string
	Channel  string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyBatch) Key() string { return "taxonomy/taxonomybatch/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyBatch decodes a stored record.
func ParseTaxonomyBatch(f codec.Fields) (TaxonomyBatch, error) {
	var x TaxonomyBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomybatch: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomybatch %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomybatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomybatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomybatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyBatch) String() string {
	return fmt.Sprintf("TaxonomyBatch(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyBatch) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyBatch) Normalize() TaxonomyBatch {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortTaxonomyBatchs orders by priority descending, then by ID.
func SortTaxonomyBatchs(xs []TaxonomyBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyBatchs drops the archived entries, preserving order.
func ActiveTaxonomyBatchs(xs []TaxonomyBatch) []TaxonomyBatch {
	out := make([]TaxonomyBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyBatchs adds up the priority of every active entry.
func SumTaxonomyBatchs(xs []TaxonomyBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTaxonomyBatchs keys the entries by ID, last write winning.
func IndexTaxonomyBatchs(xs []TaxonomyBatch) map[string]TaxonomyBatch {
	out := make(map[string]TaxonomyBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyBatch builds a valid TaxonomyBatch, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyBatch(id string, at time.Time) TaxonomyBatch {
	return TaxonomyBatch{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by taxonomy.SampleTaxonomyBatch",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyRequestSchema is the stored shape of a TaxonomyRequest.
var TaxonomyRequestSchema = schema.New("taxonomy.taxonomyrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyRequest is one taxonomy request.
type TaxonomyRequest struct {
	ID       string
	Name     string
	Group    string
	Region   string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyRequest) Key() string { return "taxonomy/taxonomyrequest/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"group":    x.Group,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyRequest decodes a stored record.
func ParseTaxonomyRequest(f codec.Fields) (TaxonomyRequest, error) {
	var x TaxonomyRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyrequest: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrequest %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyRequest) String() string {
	return fmt.Sprintf("TaxonomyRequest(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyRequest) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyRequest) Normalize() TaxonomyRequest {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTaxonomyRequests orders by capacity descending, then by ID.
func SortTaxonomyRequests(xs []TaxonomyRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyRequests drops the archived entries, preserving order.
func ActiveTaxonomyRequests(xs []TaxonomyRequest) []TaxonomyRequest {
	out := make([]TaxonomyRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyRequests adds up the capacity of every active entry.
func SumTaxonomyRequests(xs []TaxonomyRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTaxonomyRequests keys the entries by ID, last write winning.
func IndexTaxonomyRequests(xs []TaxonomyRequest) map[string]TaxonomyRequest {
	out := make(map[string]TaxonomyRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyRequest builds a valid TaxonomyRequest, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyRequest(id string, at time.Time) TaxonomyRequest {
	return TaxonomyRequest{
		ID:       id,
		Name:     "name-" + id,
		Group:    "group-" + id,
		Region:   "seeded by taxonomy.SampleTaxonomyRequest",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomySummarySchema is the stored shape of a TaxonomySummary.
var TaxonomySummarySchema = schema.New("taxonomy.taxonomysummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomySummary is one taxonomy summary.
type TaxonomySummary struct {
	ID       string
	Group    string
	Name     string
	Region   string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomySummary) Key() string { return "taxonomy/taxonomysummary/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomySummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomySummary decodes a stored record.
func ParseTaxonomySummary(f codec.Fields) (TaxonomySummary, error) {
	var x TaxonomySummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomysummary: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomysummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomysummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomysummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomysummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomySummary) String() string {
	return fmt.Sprintf("TaxonomySummary(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomySummary) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomySummary) Normalize() TaxonomySummary {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTaxonomySummarys orders by offset descending, then by ID.
func SortTaxonomySummarys(xs []TaxonomySummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomySummarys drops the archived entries, preserving order.
func ActiveTaxonomySummarys(xs []TaxonomySummary) []TaxonomySummary {
	out := make([]TaxonomySummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomySummarys adds up the offset of every active entry.
func SumTaxonomySummarys(xs []TaxonomySummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTaxonomySummarys keys the entries by ID, last write winning.
func IndexTaxonomySummarys(xs []TaxonomySummary) map[string]TaxonomySummary {
	out := make(map[string]TaxonomySummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomySummary builds a valid TaxonomySummary, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomySummary(id string, at time.Time) TaxonomySummary {
	return TaxonomySummary{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Region:   "seeded by taxonomy.SampleTaxonomySummary",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyPolicySchema is the stored shape of a TaxonomyPolicy.
var TaxonomyPolicySchema = schema.New("taxonomy.taxonomypolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyPolicy is one taxonomy policy.
type TaxonomyPolicy struct {
	ID        string
	Reference string
	Source    string
	Category  string
	Attempts  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TaxonomyPolicy) Key() string { return "taxonomy/taxonomypolicy/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"source":    x.Source,
		"category":  x.Category,
		"attempts":  strconv.Itoa(x.Attempts),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyPolicy decodes a stored record.
func ParseTaxonomyPolicy(f codec.Fields) (TaxonomyPolicy, error) {
	var x TaxonomyPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomypolicy: missing id")
	}
	x.Reference = f["reference"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomypolicy %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomypolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomypolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomypolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyPolicy) String() string {
	return fmt.Sprintf("TaxonomyPolicy(%s, reference=%q, attempts=%d, archived=%t)", x.ID, x.Reference, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyPolicy) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyPolicy) Normalize() TaxonomyPolicy {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTaxonomyPolicys orders by attempts descending, then by ID.
func SortTaxonomyPolicys(xs []TaxonomyPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyPolicys drops the archived entries, preserving order.
func ActiveTaxonomyPolicys(xs []TaxonomyPolicy) []TaxonomyPolicy {
	out := make([]TaxonomyPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyPolicys adds up the attempts of every active entry.
func SumTaxonomyPolicys(xs []TaxonomyPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTaxonomyPolicys keys the entries by ID, last write winning.
func IndexTaxonomyPolicys(xs []TaxonomyPolicy) map[string]TaxonomyPolicy {
	out := make(map[string]TaxonomyPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyPolicy builds a valid TaxonomyPolicy, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyPolicy(id string, at time.Time) TaxonomyPolicy {
	return TaxonomyPolicy{
		ID:        id,
		Reference: "reference-" + id,
		Source:    "source-" + id,
		Category:  "seeded by taxonomy.SampleTaxonomyPolicy",
		Attempts:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TaxonomyLinkSchema is the stored shape of a TaxonomyLink.
var TaxonomyLinkSchema = schema.New("taxonomy.taxonomylink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyLink is one taxonomy link.
type TaxonomyLink struct {
	ID       string
	Group    string
	Comment  string
	Name     string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyLink) Key() string { return "taxonomy/taxonomylink/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"comment":  x.Comment,
		"name":     x.Name,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyLink decodes a stored record.
func ParseTaxonomyLink(f codec.Fields) (TaxonomyLink, error) {
	var x TaxonomyLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomylink: missing id")
	}
	x.Group = f["group"]
	x.Comment = f["comment"]
	x.Name = f["name"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomylink %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomylink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomylink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomylink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyLink) String() string {
	return fmt.Sprintf("TaxonomyLink(%s, group=%q, weight=%d, archived=%t)", x.ID, x.Group, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyLink) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyLink) Normalize() TaxonomyLink {
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTaxonomyLinks orders by weight descending, then by ID.
func SortTaxonomyLinks(xs []TaxonomyLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyLinks drops the archived entries, preserving order.
func ActiveTaxonomyLinks(xs []TaxonomyLink) []TaxonomyLink {
	out := make([]TaxonomyLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyLinks adds up the weight of every active entry.
func SumTaxonomyLinks(xs []TaxonomyLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTaxonomyLinks keys the entries by ID, last write winning.
func IndexTaxonomyLinks(xs []TaxonomyLink) map[string]TaxonomyLink {
	out := make(map[string]TaxonomyLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyLink builds a valid TaxonomyLink, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyLink(id string, at time.Time) TaxonomyLink {
	return TaxonomyLink{
		ID:       id,
		Group:    "group-" + id,
		Comment:  "comment-" + id,
		Name:     "seeded by taxonomy.SampleTaxonomyLink",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyNoteSchema is the stored shape of a TaxonomyNote.
var TaxonomyNoteSchema = schema.New("taxonomy.taxonomynote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyNote is one taxonomy note.
type TaxonomyNote struct {
	ID        string
	Channel   string
	Reference string
	Region    string
	Count     int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TaxonomyNote) Key() string { return "taxonomy/taxonomynote/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyNote) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"region":    x.Region,
		"count":     strconv.Itoa(x.Count),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyNote decodes a stored record.
func ParseTaxonomyNote(f codec.Fields) (TaxonomyNote, error) {
	var x TaxonomyNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomynote: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomynote %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomynote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomynote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomynote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyNote) String() string {
	return fmt.Sprintf("TaxonomyNote(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyNote) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyNote) Normalize() TaxonomyNote {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTaxonomyNotes orders by count descending, then by ID.
func SortTaxonomyNotes(xs []TaxonomyNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyNotes drops the archived entries, preserving order.
func ActiveTaxonomyNotes(xs []TaxonomyNote) []TaxonomyNote {
	out := make([]TaxonomyNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyNotes adds up the count of every active entry.
func SumTaxonomyNotes(xs []TaxonomyNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTaxonomyNotes keys the entries by ID, last write winning.
func IndexTaxonomyNotes(xs []TaxonomyNote) map[string]TaxonomyNote {
	out := make(map[string]TaxonomyNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyNote builds a valid TaxonomyNote, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyNote(id string, at time.Time) TaxonomyNote {
	return TaxonomyNote{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by taxonomy.SampleTaxonomyNote",
		Count:     len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TaxonomyWindowSchema is the stored shape of a TaxonomyWindow.
var TaxonomyWindowSchema = schema.New("taxonomy.taxonomywindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyWindow is one taxonomy window.
type TaxonomyWindow struct {
	ID       string
	Label    string
	Region   string
	Category string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyWindow) Key() string { return "taxonomy/taxonomywindow/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyWindow decodes a stored record.
func ParseTaxonomyWindow(f codec.Fields) (TaxonomyWindow, error) {
	var x TaxonomyWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomywindow: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomywindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomywindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomywindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomywindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyWindow) String() string {
	return fmt.Sprintf("TaxonomyWindow(%s, label=%q, capacity=%d, archived=%t)", x.ID, x.Label, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyWindow) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyWindow) Normalize() TaxonomyWindow {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTaxonomyWindows orders by capacity descending, then by ID.
func SortTaxonomyWindows(xs []TaxonomyWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyWindows drops the archived entries, preserving order.
func ActiveTaxonomyWindows(xs []TaxonomyWindow) []TaxonomyWindow {
	out := make([]TaxonomyWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyWindows adds up the capacity of every active entry.
func SumTaxonomyWindows(xs []TaxonomyWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTaxonomyWindows keys the entries by ID, last write winning.
func IndexTaxonomyWindows(xs []TaxonomyWindow) map[string]TaxonomyWindow {
	out := make(map[string]TaxonomyWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyWindow builds a valid TaxonomyWindow, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyWindow(id string, at time.Time) TaxonomyWindow {
	return TaxonomyWindow{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Category: "seeded by taxonomy.SampleTaxonomyWindow",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyQuotaSchema is the stored shape of a TaxonomyQuota.
var TaxonomyQuotaSchema = schema.New("taxonomy.taxonomyquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyQuota is one taxonomy quota.
type TaxonomyQuota struct {
	ID       string
	Owner    string
	Channel  string
	Comment  string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyQuota) Key() string { return "taxonomy/taxonomyquota/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyQuota decodes a stored record.
func ParseTaxonomyQuota(f codec.Fields) (TaxonomyQuota, error) {
	var x TaxonomyQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyquota: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyquota %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyQuota) String() string {
	return fmt.Sprintf("TaxonomyQuota(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyQuota) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyQuota) Normalize() TaxonomyQuota {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortTaxonomyQuotas orders by count descending, then by ID.
func SortTaxonomyQuotas(xs []TaxonomyQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyQuotas drops the archived entries, preserving order.
func ActiveTaxonomyQuotas(xs []TaxonomyQuota) []TaxonomyQuota {
	out := make([]TaxonomyQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyQuotas adds up the count of every active entry.
func SumTaxonomyQuotas(xs []TaxonomyQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTaxonomyQuotas keys the entries by ID, last write winning.
func IndexTaxonomyQuotas(xs []TaxonomyQuota) map[string]TaxonomyQuota {
	out := make(map[string]TaxonomyQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyQuota builds a valid TaxonomyQuota, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyQuota(id string, at time.Time) TaxonomyQuota {
	return TaxonomyQuota{
		ID:       id,
		Owner:    "owner-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by taxonomy.SampleTaxonomyQuota",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyBucketSchema is the stored shape of a TaxonomyBucket.
var TaxonomyBucketSchema = schema.New("taxonomy.taxonomybucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyBucket is one taxonomy bucket.
type TaxonomyBucket struct {
	ID       string
	Group    string
	Owner    string
	Source   string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyBucket) Key() string { return "taxonomy/taxonomybucket/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyBucket decodes a stored record.
func ParseTaxonomyBucket(f codec.Fields) (TaxonomyBucket, error) {
	var x TaxonomyBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomybucket: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomybucket %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomybucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomybucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomybucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyBucket) String() string {
	return fmt.Sprintf("TaxonomyBucket(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyBucket) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyBucket) Normalize() TaxonomyBucket {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTaxonomyBuckets orders by priority descending, then by ID.
func SortTaxonomyBuckets(xs []TaxonomyBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyBuckets drops the archived entries, preserving order.
func ActiveTaxonomyBuckets(xs []TaxonomyBucket) []TaxonomyBucket {
	out := make([]TaxonomyBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyBuckets adds up the priority of every active entry.
func SumTaxonomyBuckets(xs []TaxonomyBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTaxonomyBuckets keys the entries by ID, last write winning.
func IndexTaxonomyBuckets(xs []TaxonomyBucket) map[string]TaxonomyBucket {
	out := make(map[string]TaxonomyBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyBucket builds a valid TaxonomyBucket, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyBucket(id string, at time.Time) TaxonomyBucket {
	return TaxonomyBucket{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by taxonomy.SampleTaxonomyBucket",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyLaneSchema is the stored shape of a TaxonomyLane.
var TaxonomyLaneSchema = schema.New("taxonomy.taxonomylane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyLane is one taxonomy lane.
type TaxonomyLane struct {
	ID       string
	Name     string
	Category string
	Region   string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyLane) Key() string { return "taxonomy/taxonomylane/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"region":   x.Region,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyLane decodes a stored record.
func ParseTaxonomyLane(f codec.Fields) (TaxonomyLane, error) {
	var x TaxonomyLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomylane: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomylane %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomylane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomylane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomylane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyLane) String() string {
	return fmt.Sprintf("TaxonomyLane(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyLane) Normalize() TaxonomyLane {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTaxonomyLanes orders by priority descending, then by ID.
func SortTaxonomyLanes(xs []TaxonomyLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyLanes drops the archived entries, preserving order.
func ActiveTaxonomyLanes(xs []TaxonomyLane) []TaxonomyLane {
	out := make([]TaxonomyLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyLanes adds up the priority of every active entry.
func SumTaxonomyLanes(xs []TaxonomyLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTaxonomyLanes keys the entries by ID, last write winning.
func IndexTaxonomyLanes(xs []TaxonomyLane) map[string]TaxonomyLane {
	out := make(map[string]TaxonomyLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyLane builds a valid TaxonomyLane, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyLane(id string, at time.Time) TaxonomyLane {
	return TaxonomyLane{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Region:   "seeded by taxonomy.SampleTaxonomyLane",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomySlotSchema is the stored shape of a TaxonomySlot.
var TaxonomySlotSchema = schema.New("taxonomy.taxonomyslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomySlot is one taxonomy slot.
type TaxonomySlot struct {
	ID       string
	Comment  string
	Channel  string
	Label    string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomySlot) Key() string { return "taxonomy/taxonomyslot/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomySlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomySlot decodes a stored record.
func ParseTaxonomySlot(f codec.Fields) (TaxonomySlot, error) {
	var x TaxonomySlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyslot: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyslot %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomySlot) String() string {
	return fmt.Sprintf("TaxonomySlot(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomySlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomySlot) Normalize() TaxonomySlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTaxonomySlots orders by count descending, then by ID.
func SortTaxonomySlots(xs []TaxonomySlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomySlots drops the archived entries, preserving order.
func ActiveTaxonomySlots(xs []TaxonomySlot) []TaxonomySlot {
	out := make([]TaxonomySlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomySlots adds up the count of every active entry.
func SumTaxonomySlots(xs []TaxonomySlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTaxonomySlots keys the entries by ID, last write winning.
func IndexTaxonomySlots(xs []TaxonomySlot) map[string]TaxonomySlot {
	out := make(map[string]TaxonomySlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomySlot builds a valid TaxonomySlot, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomySlot(id string, at time.Time) TaxonomySlot {
	return TaxonomySlot{
		ID:       id,
		Comment:  "comment-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by taxonomy.SampleTaxonomySlot",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyTagSchema is the stored shape of a TaxonomyTag.
var TaxonomyTagSchema = schema.New("taxonomy.taxonomytag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyTag is one taxonomy tag.
type TaxonomyTag struct {
	ID       string
	Group    string
	Comment  string
	Name     string
	Retained int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyTag) Key() string { return "taxonomy/taxonomytag/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"comment":  x.Comment,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyTag decodes a stored record.
func ParseTaxonomyTag(f codec.Fields) (TaxonomyTag, error) {
	var x TaxonomyTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomytag: missing id")
	}
	x.Group = f["group"]
	x.Comment = f["comment"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomytag %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomytag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomytag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomytag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyTag) String() string {
	return fmt.Sprintf("TaxonomyTag(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyTag) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyTag) Normalize() TaxonomyTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortTaxonomyTags orders by retained descending, then by ID.
func SortTaxonomyTags(xs []TaxonomyTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyTags drops the archived entries, preserving order.
func ActiveTaxonomyTags(xs []TaxonomyTag) []TaxonomyTag {
	out := make([]TaxonomyTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyTags adds up the retained of every active entry.
func SumTaxonomyTags(xs []TaxonomyTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTaxonomyTags keys the entries by ID, last write winning.
func IndexTaxonomyTags(xs []TaxonomyTag) map[string]TaxonomyTag {
	out := make(map[string]TaxonomyTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyTag builds a valid TaxonomyTag, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyTag(id string, at time.Time) TaxonomyTag {
	return TaxonomyTag{
		ID:       id,
		Group:    "group-" + id,
		Comment:  "comment-" + id,
		Name:     "seeded by taxonomy.SampleTaxonomyTag",
		Retained: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyRuleSchema is the stored shape of a TaxonomyRule.
var TaxonomyRuleSchema = schema.New("taxonomy.taxonomyrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyRule is one taxonomy rule.
type TaxonomyRule struct {
	ID        string
	Region    string
	Reference string
	Source    string
	Offset    int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TaxonomyRule) Key() string { return "taxonomy/taxonomyrule/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"source":    x.Source,
		"offset":    strconv.Itoa(x.Offset),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyRule decodes a stored record.
func ParseTaxonomyRule(f codec.Fields) (TaxonomyRule, error) {
	var x TaxonomyRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyrule: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrule %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyRule) String() string {
	return fmt.Sprintf("TaxonomyRule(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyRule) Normalize() TaxonomyRule {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTaxonomyRules orders by offset descending, then by ID.
func SortTaxonomyRules(xs []TaxonomyRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyRules drops the archived entries, preserving order.
func ActiveTaxonomyRules(xs []TaxonomyRule) []TaxonomyRule {
	out := make([]TaxonomyRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyRules adds up the offset of every active entry.
func SumTaxonomyRules(xs []TaxonomyRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTaxonomyRules keys the entries by ID, last write winning.
func IndexTaxonomyRules(xs []TaxonomyRule) map[string]TaxonomyRule {
	out := make(map[string]TaxonomyRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyRule builds a valid TaxonomyRule, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyRule(id string, at time.Time) TaxonomyRule {
	return TaxonomyRule{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by taxonomy.SampleTaxonomyRule",
		Offset:    len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TaxonomyHoldSchema is the stored shape of a TaxonomyHold.
var TaxonomyHoldSchema = schema.New("taxonomy.taxonomyhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyHold is one taxonomy hold.
type TaxonomyHold struct {
	ID       string
	Name     string
	Channel  string
	Label    string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyHold) Key() string { return "taxonomy/taxonomyhold/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyHold decodes a stored record.
func ParseTaxonomyHold(f codec.Fields) (TaxonomyHold, error) {
	var x TaxonomyHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyhold: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyhold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyHold) String() string {
	return fmt.Sprintf("TaxonomyHold(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyHold) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyHold) Normalize() TaxonomyHold {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTaxonomyHolds orders by count descending, then by ID.
func SortTaxonomyHolds(xs []TaxonomyHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyHolds drops the archived entries, preserving order.
func ActiveTaxonomyHolds(xs []TaxonomyHold) []TaxonomyHold {
	out := make([]TaxonomyHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyHolds adds up the count of every active entry.
func SumTaxonomyHolds(xs []TaxonomyHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTaxonomyHolds keys the entries by ID, last write winning.
func IndexTaxonomyHolds(xs []TaxonomyHold) map[string]TaxonomyHold {
	out := make(map[string]TaxonomyHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyHold builds a valid TaxonomyHold, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyHold(id string, at time.Time) TaxonomyHold {
	return TaxonomyHold{
		ID:      id,
		Name:    "name-" + id,
		Channel: "channel-" + id,
		Label:   "seeded by taxonomy.SampleTaxonomyHold",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TaxonomyClaimSchema is the stored shape of a TaxonomyClaim.
var TaxonomyClaimSchema = schema.New("taxonomy.taxonomyclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyClaim is one taxonomy claim.
type TaxonomyClaim struct {
	ID       string
	Source   string
	Owner    string
	Channel  string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyClaim) Key() string { return "taxonomy/taxonomyclaim/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyClaim decodes a stored record.
func ParseTaxonomyClaim(f codec.Fields) (TaxonomyClaim, error) {
	var x TaxonomyClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomyclaim: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomyclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomyclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyClaim) String() string {
	return fmt.Sprintf("TaxonomyClaim(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyClaim) Normalize() TaxonomyClaim {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTaxonomyClaims orders by retained descending, then by ID.
func SortTaxonomyClaims(xs []TaxonomyClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyClaims drops the archived entries, preserving order.
func ActiveTaxonomyClaims(xs []TaxonomyClaim) []TaxonomyClaim {
	out := make([]TaxonomyClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyClaims adds up the retained of every active entry.
func SumTaxonomyClaims(xs []TaxonomyClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTaxonomyClaims keys the entries by ID, last write winning.
func IndexTaxonomyClaims(xs []TaxonomyClaim) map[string]TaxonomyClaim {
	out := make(map[string]TaxonomyClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyClaim builds a valid TaxonomyClaim, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyClaim(id string, at time.Time) TaxonomyClaim {
	return TaxonomyClaim{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by taxonomy.SampleTaxonomyClaim",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TaxonomyCursorSchema is the stored shape of a TaxonomyCursor.
var TaxonomyCursorSchema = schema.New("taxonomy.taxonomycursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TaxonomyCursor is one taxonomy cursor.
type TaxonomyCursor struct {
	ID       string
	Source   string
	Group    string
	Channel  string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TaxonomyCursor) Key() string { return "taxonomy/taxonomycursor/" + x.ID }

// Fields encodes x for storage.
func (x TaxonomyCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"group":    x.Group,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTaxonomyCursor decodes a stored record.
func ParseTaxonomyCursor(f codec.Fields) (TaxonomyCursor, error) {
	var x TaxonomyCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("taxonomy: taxonomycursor: missing id")
	}
	x.Source = f["source"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomycursor %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomycursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("taxonomy: taxonomycursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("taxonomy: taxonomycursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TaxonomyCursor) String() string {
	return fmt.Sprintf("TaxonomyCursor(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TaxonomyCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TaxonomyCursor) Normalize() TaxonomyCursor {
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTaxonomyCursors orders by weight descending, then by ID.
func SortTaxonomyCursors(xs []TaxonomyCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTaxonomyCursors drops the archived entries, preserving order.
func ActiveTaxonomyCursors(xs []TaxonomyCursor) []TaxonomyCursor {
	out := make([]TaxonomyCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTaxonomyCursors adds up the weight of every active entry.
func SumTaxonomyCursors(xs []TaxonomyCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTaxonomyCursors keys the entries by ID, last write winning.
func IndexTaxonomyCursors(xs []TaxonomyCursor) map[string]TaxonomyCursor {
	out := make(map[string]TaxonomyCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTaxonomyCursor builds a valid TaxonomyCursor, for tests and for the seed data the cron
// binary writes.
func SampleTaxonomyCursor(id string, at time.Time) TaxonomyCursor {
	return TaxonomyCursor{
		ID:      id,
		Source:  "source-" + id,
		Group:   "group-" + id,
		Channel: "seeded by taxonomy.SampleTaxonomyCursor",
		Weight:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	TaxonomyRecordSchema,
	TaxonomyEntrySchema,
	TaxonomyItemSchema,
	TaxonomyDraftSchema,
	TaxonomyRevisionSchema,
	TaxonomySnapshotSchema,
	TaxonomyBatchSchema,
	TaxonomyRequestSchema,
	TaxonomySummarySchema,
	TaxonomyPolicySchema,
	TaxonomyLinkSchema,
	TaxonomyNoteSchema,
	TaxonomyWindowSchema,
	TaxonomyQuotaSchema,
	TaxonomyBucketSchema,
	TaxonomyLaneSchema,
	TaxonomySlotSchema,
	TaxonomyTagSchema,
	TaxonomyRuleSchema,
	TaxonomyHoldSchema,
	TaxonomyClaimSchema,
	TaxonomyCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"taxonomy/taxonomyrecord/",
	"taxonomy/taxonomyentry/",
	"taxonomy/taxonomyitem/",
	"taxonomy/taxonomydraft/",
	"taxonomy/taxonomyrevision/",
	"taxonomy/taxonomysnapshot/",
	"taxonomy/taxonomybatch/",
	"taxonomy/taxonomyrequest/",
	"taxonomy/taxonomysummary/",
	"taxonomy/taxonomypolicy/",
	"taxonomy/taxonomylink/",
	"taxonomy/taxonomynote/",
	"taxonomy/taxonomywindow/",
	"taxonomy/taxonomyquota/",
	"taxonomy/taxonomybucket/",
	"taxonomy/taxonomylane/",
	"taxonomy/taxonomyslot/",
	"taxonomy/taxonomytag/",
	"taxonomy/taxonomyrule/",
	"taxonomy/taxonomyhold/",
	"taxonomy/taxonomyclaim/",
	"taxonomy/taxonomycursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "taxonomy: " + strings.Join(names, ", ")
}
