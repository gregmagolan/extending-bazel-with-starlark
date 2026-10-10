// Package preference holds the preference domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package preference

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// PreferenceRecordSchema is the stored shape of a PreferenceRecord.
var PreferenceRecordSchema = schema.New("preference.preferencerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceRecord is one preference record.
type PreferenceRecord struct {
	ID       string
	Category string
	Region   string
	Group    string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceRecord) Key() string { return "preference/preferencerecord/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceRecord decodes a stored record.
func ParsePreferenceRecord(f codec.Fields) (PreferenceRecord, error) {
	var x PreferenceRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencerecord: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceRecord) String() string {
	return fmt.Sprintf("PreferenceRecord(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceRecord) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceRecord) Normalize() PreferenceRecord {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPreferenceRecords orders by offset descending, then by ID.
func SortPreferenceRecords(xs []PreferenceRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceRecords drops the archived entries, preserving order.
func ActivePreferenceRecords(xs []PreferenceRecord) []PreferenceRecord {
	out := make([]PreferenceRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceRecords adds up the offset of every active entry.
func SumPreferenceRecords(xs []PreferenceRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPreferenceRecords keys the entries by ID, last write winning.
func IndexPreferenceRecords(xs []PreferenceRecord) map[string]PreferenceRecord {
	out := make(map[string]PreferenceRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceRecord builds a valid PreferenceRecord, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceRecord(id string, at time.Time) PreferenceRecord {
	return PreferenceRecord{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Group:    "seeded by preference.SamplePreferenceRecord",
		Offset:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceEntrySchema is the stored shape of a PreferenceEntry.
var PreferenceEntrySchema = schema.New("preference.preferenceentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceEntry is one preference entry.
type PreferenceEntry struct {
	ID       string
	Owner    string
	Comment  string
	Group    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceEntry) Key() string { return "preference/preferenceentry/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceEntry decodes a stored record.
func ParsePreferenceEntry(f codec.Fields) (PreferenceEntry, error) {
	var x PreferenceEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferenceentry: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceentry %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferenceentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferenceentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceEntry) String() string {
	return fmt.Sprintf("PreferenceEntry(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceEntry) Normalize() PreferenceEntry {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortPreferenceEntrys orders by offset descending, then by ID.
func SortPreferenceEntrys(xs []PreferenceEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceEntrys drops the archived entries, preserving order.
func ActivePreferenceEntrys(xs []PreferenceEntry) []PreferenceEntry {
	out := make([]PreferenceEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceEntrys adds up the offset of every active entry.
func SumPreferenceEntrys(xs []PreferenceEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPreferenceEntrys keys the entries by ID, last write winning.
func IndexPreferenceEntrys(xs []PreferenceEntry) map[string]PreferenceEntry {
	out := make(map[string]PreferenceEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceEntry builds a valid PreferenceEntry, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceEntry(id string, at time.Time) PreferenceEntry {
	return PreferenceEntry{
		ID:      id,
		Owner:   "owner-" + id,
		Comment: "comment-" + id,
		Group:   "seeded by preference.SamplePreferenceEntry",
		Offset:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// PreferenceItemSchema is the stored shape of a PreferenceItem.
var PreferenceItemSchema = schema.New("preference.preferenceitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceItem is one preference item.
type PreferenceItem struct {
	ID       string
	Category string
	Region   string
	Source   string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceItem) Key() string { return "preference/preferenceitem/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceItem decodes a stored record.
func ParsePreferenceItem(f codec.Fields) (PreferenceItem, error) {
	var x PreferenceItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferenceitem: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceitem %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferenceitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferenceitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceItem) String() string {
	return fmt.Sprintf("PreferenceItem(%s, category=%q, attempts=%d, archived=%t)", x.ID, x.Category, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceItem) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceItem) Normalize() PreferenceItem {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPreferenceItems orders by attempts descending, then by ID.
func SortPreferenceItems(xs []PreferenceItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceItems drops the archived entries, preserving order.
func ActivePreferenceItems(xs []PreferenceItem) []PreferenceItem {
	out := make([]PreferenceItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceItems adds up the attempts of every active entry.
func SumPreferenceItems(xs []PreferenceItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPreferenceItems keys the entries by ID, last write winning.
func IndexPreferenceItems(xs []PreferenceItem) map[string]PreferenceItem {
	out := make(map[string]PreferenceItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceItem builds a valid PreferenceItem, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceItem(id string, at time.Time) PreferenceItem {
	return PreferenceItem{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Source:   "seeded by preference.SamplePreferenceItem",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceDraftSchema is the stored shape of a PreferenceDraft.
var PreferenceDraftSchema = schema.New("preference.preferencedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceDraft is one preference draft.
type PreferenceDraft struct {
	ID       string
	Label    string
	Category string
	Channel  string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceDraft) Key() string { return "preference/preferencedraft/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceDraft decodes a stored record.
func ParsePreferenceDraft(f codec.Fields) (PreferenceDraft, error) {
	var x PreferenceDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencedraft: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("preference: preferencedraft %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferencedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceDraft) String() string {
	return fmt.Sprintf("PreferenceDraft(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceDraft) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceDraft) Normalize() PreferenceDraft {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPreferenceDrafts orders by weight descending, then by ID.
func SortPreferenceDrafts(xs []PreferenceDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceDrafts drops the archived entries, preserving order.
func ActivePreferenceDrafts(xs []PreferenceDraft) []PreferenceDraft {
	out := make([]PreferenceDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceDrafts adds up the weight of every active entry.
func SumPreferenceDrafts(xs []PreferenceDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexPreferenceDrafts keys the entries by ID, last write winning.
func IndexPreferenceDrafts(xs []PreferenceDraft) map[string]PreferenceDraft {
	out := make(map[string]PreferenceDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceDraft builds a valid PreferenceDraft, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceDraft(id string, at time.Time) PreferenceDraft {
	return PreferenceDraft{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Channel:  "seeded by preference.SamplePreferenceDraft",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceRevisionSchema is the stored shape of a PreferenceRevision.
var PreferenceRevisionSchema = schema.New("preference.preferencerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceRevision is one preference revision.
type PreferenceRevision struct {
	ID       string
	Label    string
	Name     string
	Group    string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceRevision) Key() string { return "preference/preferencerevision/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"name":     x.Name,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceRevision decodes a stored record.
func ParsePreferenceRevision(f codec.Fields) (PreferenceRevision, error) {
	var x PreferenceRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencerevision: missing id")
	}
	x.Label = f["label"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerevision %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceRevision) String() string {
	return fmt.Sprintf("PreferenceRevision(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceRevision) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceRevision) Normalize() PreferenceRevision {
	x.Label = strings.TrimSpace(x.Label)
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

// SortPreferenceRevisions orders by attempts descending, then by ID.
func SortPreferenceRevisions(xs []PreferenceRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceRevisions drops the archived entries, preserving order.
func ActivePreferenceRevisions(xs []PreferenceRevision) []PreferenceRevision {
	out := make([]PreferenceRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceRevisions adds up the attempts of every active entry.
func SumPreferenceRevisions(xs []PreferenceRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPreferenceRevisions keys the entries by ID, last write winning.
func IndexPreferenceRevisions(xs []PreferenceRevision) map[string]PreferenceRevision {
	out := make(map[string]PreferenceRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceRevision builds a valid PreferenceRevision, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceRevision(id string, at time.Time) PreferenceRevision {
	return PreferenceRevision{
		ID:       id,
		Label:    "label-" + id,
		Name:     "name-" + id,
		Group:    "seeded by preference.SamplePreferenceRevision",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceSnapshotSchema is the stored shape of a PreferenceSnapshot.
var PreferenceSnapshotSchema = schema.New("preference.preferencesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceSnapshot is one preference snapshot.
type PreferenceSnapshot struct {
	ID       string
	Group    string
	Channel  string
	Category string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceSnapshot) Key() string { return "preference/preferencesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceSnapshot decodes a stored record.
func ParsePreferenceSnapshot(f codec.Fields) (PreferenceSnapshot, error) {
	var x PreferenceSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencesnapshot: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("preference: preferencesnapshot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("preference: preferencesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceSnapshot) String() string {
	return fmt.Sprintf("PreferenceSnapshot(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceSnapshot) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceSnapshot) Normalize() PreferenceSnapshot {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPreferenceSnapshots orders by capacity descending, then by ID.
func SortPreferenceSnapshots(xs []PreferenceSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceSnapshots drops the archived entries, preserving order.
func ActivePreferenceSnapshots(xs []PreferenceSnapshot) []PreferenceSnapshot {
	out := make([]PreferenceSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceSnapshots adds up the capacity of every active entry.
func SumPreferenceSnapshots(xs []PreferenceSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPreferenceSnapshots keys the entries by ID, last write winning.
func IndexPreferenceSnapshots(xs []PreferenceSnapshot) map[string]PreferenceSnapshot {
	out := make(map[string]PreferenceSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceSnapshot builds a valid PreferenceSnapshot, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceSnapshot(id string, at time.Time) PreferenceSnapshot {
	return PreferenceSnapshot{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by preference.SamplePreferenceSnapshot",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceBatchSchema is the stored shape of a PreferenceBatch.
var PreferenceBatchSchema = schema.New("preference.preferencebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceBatch is one preference batch.
type PreferenceBatch struct {
	ID       string
	Source   string
	Channel  string
	Owner    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceBatch) Key() string { return "preference/preferencebatch/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceBatch decodes a stored record.
func ParsePreferenceBatch(f codec.Fields) (PreferenceBatch, error) {
	var x PreferenceBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencebatch: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("preference: preferencebatch %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceBatch) String() string {
	return fmt.Sprintf("PreferenceBatch(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceBatch) Normalize() PreferenceBatch {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPreferenceBatchs orders by capacity descending, then by ID.
func SortPreferenceBatchs(xs []PreferenceBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceBatchs drops the archived entries, preserving order.
func ActivePreferenceBatchs(xs []PreferenceBatch) []PreferenceBatch {
	out := make([]PreferenceBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceBatchs adds up the capacity of every active entry.
func SumPreferenceBatchs(xs []PreferenceBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPreferenceBatchs keys the entries by ID, last write winning.
func IndexPreferenceBatchs(xs []PreferenceBatch) map[string]PreferenceBatch {
	out := make(map[string]PreferenceBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceBatch builds a valid PreferenceBatch, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceBatch(id string, at time.Time) PreferenceBatch {
	return PreferenceBatch{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by preference.SamplePreferenceBatch",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceRequestSchema is the stored shape of a PreferenceRequest.
var PreferenceRequestSchema = schema.New("preference.preferencerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceRequest is one preference request.
type PreferenceRequest struct {
	ID        string
	Group     string
	Reference string
	Region    string
	Retained  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PreferenceRequest) Key() string { return "preference/preferencerequest/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"region":    x.Region,
		"retained":  strconv.Itoa(x.Retained),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceRequest decodes a stored record.
func ParsePreferenceRequest(f codec.Fields) (PreferenceRequest, error) {
	var x PreferenceRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencerequest: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerequest %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceRequest) String() string {
	return fmt.Sprintf("PreferenceRequest(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceRequest) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceRequest) Normalize() PreferenceRequest {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPreferenceRequests orders by retained descending, then by ID.
func SortPreferenceRequests(xs []PreferenceRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceRequests drops the archived entries, preserving order.
func ActivePreferenceRequests(xs []PreferenceRequest) []PreferenceRequest {
	out := make([]PreferenceRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceRequests adds up the retained of every active entry.
func SumPreferenceRequests(xs []PreferenceRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPreferenceRequests keys the entries by ID, last write winning.
func IndexPreferenceRequests(xs []PreferenceRequest) map[string]PreferenceRequest {
	out := make(map[string]PreferenceRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceRequest builds a valid PreferenceRequest, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceRequest(id string, at time.Time) PreferenceRequest {
	return PreferenceRequest{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by preference.SamplePreferenceRequest",
		Retained:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PreferenceSummarySchema is the stored shape of a PreferenceSummary.
var PreferenceSummarySchema = schema.New("preference.preferencesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceSummary is one preference summary.
type PreferenceSummary struct {
	ID        string
	Owner     string
	Reference string
	Channel   string
	Count     int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PreferenceSummary) Key() string { return "preference/preferencesummary/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"reference": x.Reference,
		"channel":   x.Channel,
		"count":     strconv.Itoa(x.Count),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceSummary decodes a stored record.
func ParsePreferenceSummary(f codec.Fields) (PreferenceSummary, error) {
	var x PreferenceSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencesummary: missing id")
	}
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferencesummary %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceSummary) String() string {
	return fmt.Sprintf("PreferenceSummary(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceSummary) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceSummary) Normalize() PreferenceSummary {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPreferenceSummarys orders by count descending, then by ID.
func SortPreferenceSummarys(xs []PreferenceSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceSummarys drops the archived entries, preserving order.
func ActivePreferenceSummarys(xs []PreferenceSummary) []PreferenceSummary {
	out := make([]PreferenceSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceSummarys adds up the count of every active entry.
func SumPreferenceSummarys(xs []PreferenceSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPreferenceSummarys keys the entries by ID, last write winning.
func IndexPreferenceSummarys(xs []PreferenceSummary) map[string]PreferenceSummary {
	out := make(map[string]PreferenceSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceSummary builds a valid PreferenceSummary, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceSummary(id string, at time.Time) PreferenceSummary {
	return PreferenceSummary{
		ID:        id,
		Owner:     "owner-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by preference.SamplePreferenceSummary",
		Count:     len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PreferencePolicySchema is the stored shape of a PreferencePolicy.
var PreferencePolicySchema = schema.New("preference.preferencepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferencePolicy is one preference policy.
type PreferencePolicy struct {
	ID        string
	Group     string
	Channel   string
	Reference string
	Retained  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PreferencePolicy) Key() string { return "preference/preferencepolicy/" + x.ID }

// Fields encodes x for storage.
func (x PreferencePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"channel":   x.Channel,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferencePolicy decodes a stored record.
func ParsePreferencePolicy(f codec.Fields) (PreferencePolicy, error) {
	var x PreferencePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencepolicy: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("preference: preferencepolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferencepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferencePolicy) String() string {
	return fmt.Sprintf("PreferencePolicy(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferencePolicy) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferencePolicy) Normalize() PreferencePolicy {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPreferencePolicys orders by retained descending, then by ID.
func SortPreferencePolicys(xs []PreferencePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferencePolicys drops the archived entries, preserving order.
func ActivePreferencePolicys(xs []PreferencePolicy) []PreferencePolicy {
	out := make([]PreferencePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferencePolicys adds up the retained of every active entry.
func SumPreferencePolicys(xs []PreferencePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPreferencePolicys keys the entries by ID, last write winning.
func IndexPreferencePolicys(xs []PreferencePolicy) map[string]PreferencePolicy {
	out := make(map[string]PreferencePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferencePolicy builds a valid PreferencePolicy, for tests and for the seed data the cron
// binary writes.
func SamplePreferencePolicy(id string, at time.Time) PreferencePolicy {
	return PreferencePolicy{
		ID:        id,
		Group:     "group-" + id,
		Channel:   "channel-" + id,
		Reference: "seeded by preference.SamplePreferencePolicy",
		Retained:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PreferenceLinkSchema is the stored shape of a PreferenceLink.
var PreferenceLinkSchema = schema.New("preference.preferencelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceLink is one preference link.
type PreferenceLink struct {
	ID       string
	Name     string
	Source   string
	Owner    string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceLink) Key() string { return "preference/preferencelink/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceLink decodes a stored record.
func ParsePreferenceLink(f codec.Fields) (PreferenceLink, error) {
	var x PreferenceLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencelink: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("preference: preferencelink %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceLink) String() string {
	return fmt.Sprintf("PreferenceLink(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceLink) Normalize() PreferenceLink {
	x.Name = strings.TrimSpace(x.Name)
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

// SortPreferenceLinks orders by retained descending, then by ID.
func SortPreferenceLinks(xs []PreferenceLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceLinks drops the archived entries, preserving order.
func ActivePreferenceLinks(xs []PreferenceLink) []PreferenceLink {
	out := make([]PreferenceLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceLinks adds up the retained of every active entry.
func SumPreferenceLinks(xs []PreferenceLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPreferenceLinks keys the entries by ID, last write winning.
func IndexPreferenceLinks(xs []PreferenceLink) map[string]PreferenceLink {
	out := make(map[string]PreferenceLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceLink builds a valid PreferenceLink, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceLink(id string, at time.Time) PreferenceLink {
	return PreferenceLink{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by preference.SamplePreferenceLink",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceNoteSchema is the stored shape of a PreferenceNote.
var PreferenceNoteSchema = schema.New("preference.preferencenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceNote is one preference note.
type PreferenceNote struct {
	ID       string
	Category string
	Owner    string
	Label    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceNote) Key() string { return "preference/preferencenote/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceNote decodes a stored record.
func ParsePreferenceNote(f codec.Fields) (PreferenceNote, error) {
	var x PreferenceNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencenote: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("preference: preferencenote %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("preference: preferencenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceNote) String() string {
	return fmt.Sprintf("PreferenceNote(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceNote) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceNote) Normalize() PreferenceNote {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortPreferenceNotes orders by offset descending, then by ID.
func SortPreferenceNotes(xs []PreferenceNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceNotes drops the archived entries, preserving order.
func ActivePreferenceNotes(xs []PreferenceNote) []PreferenceNote {
	out := make([]PreferenceNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceNotes adds up the offset of every active entry.
func SumPreferenceNotes(xs []PreferenceNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPreferenceNotes keys the entries by ID, last write winning.
func IndexPreferenceNotes(xs []PreferenceNote) map[string]PreferenceNote {
	out := make(map[string]PreferenceNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceNote builds a valid PreferenceNote, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceNote(id string, at time.Time) PreferenceNote {
	return PreferenceNote{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by preference.SamplePreferenceNote",
		Offset:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceWindowSchema is the stored shape of a PreferenceWindow.
var PreferenceWindowSchema = schema.New("preference.preferencewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceWindow is one preference window.
type PreferenceWindow struct {
	ID       string
	Comment  string
	Region   string
	Label    string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceWindow) Key() string { return "preference/preferencewindow/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"region":   x.Region,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceWindow decodes a stored record.
func ParsePreferenceWindow(f codec.Fields) (PreferenceWindow, error) {
	var x PreferenceWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencewindow: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("preference: preferencewindow %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceWindow) String() string {
	return fmt.Sprintf("PreferenceWindow(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceWindow) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceWindow) Normalize() PreferenceWindow {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPreferenceWindows orders by offset descending, then by ID.
func SortPreferenceWindows(xs []PreferenceWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceWindows drops the archived entries, preserving order.
func ActivePreferenceWindows(xs []PreferenceWindow) []PreferenceWindow {
	out := make([]PreferenceWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceWindows adds up the offset of every active entry.
func SumPreferenceWindows(xs []PreferenceWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPreferenceWindows keys the entries by ID, last write winning.
func IndexPreferenceWindows(xs []PreferenceWindow) map[string]PreferenceWindow {
	out := make(map[string]PreferenceWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceWindow builds a valid PreferenceWindow, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceWindow(id string, at time.Time) PreferenceWindow {
	return PreferenceWindow{
		ID:       id,
		Comment:  "comment-" + id,
		Region:   "region-" + id,
		Label:    "seeded by preference.SamplePreferenceWindow",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceQuotaSchema is the stored shape of a PreferenceQuota.
var PreferenceQuotaSchema = schema.New("preference.preferencequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceQuota is one preference quota.
type PreferenceQuota struct {
	ID       string
	Name     string
	Category string
	Source   string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceQuota) Key() string { return "preference/preferencequota/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceQuota decodes a stored record.
func ParsePreferenceQuota(f codec.Fields) (PreferenceQuota, error) {
	var x PreferenceQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencequota: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencequota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("preference: preferencequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceQuota) String() string {
	return fmt.Sprintf("PreferenceQuota(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceQuota) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceQuota) Normalize() PreferenceQuota {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortPreferenceQuotas orders by attempts descending, then by ID.
func SortPreferenceQuotas(xs []PreferenceQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceQuotas drops the archived entries, preserving order.
func ActivePreferenceQuotas(xs []PreferenceQuota) []PreferenceQuota {
	out := make([]PreferenceQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceQuotas adds up the attempts of every active entry.
func SumPreferenceQuotas(xs []PreferenceQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPreferenceQuotas keys the entries by ID, last write winning.
func IndexPreferenceQuotas(xs []PreferenceQuota) map[string]PreferenceQuota {
	out := make(map[string]PreferenceQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceQuota builds a valid PreferenceQuota, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceQuota(id string, at time.Time) PreferenceQuota {
	return PreferenceQuota{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Source:   "seeded by preference.SamplePreferenceQuota",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceBucketSchema is the stored shape of a PreferenceBucket.
var PreferenceBucketSchema = schema.New("preference.preferencebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceBucket is one preference bucket.
type PreferenceBucket struct {
	ID       string
	Region   string
	Label    string
	Owner    string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceBucket) Key() string { return "preference/preferencebucket/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceBucket decodes a stored record.
func ParsePreferenceBucket(f codec.Fields) (PreferenceBucket, error) {
	var x PreferenceBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencebucket: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("preference: preferencebucket %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceBucket) String() string {
	return fmt.Sprintf("PreferenceBucket(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceBucket) Normalize() PreferenceBucket {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPreferenceBuckets orders by retained descending, then by ID.
func SortPreferenceBuckets(xs []PreferenceBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceBuckets drops the archived entries, preserving order.
func ActivePreferenceBuckets(xs []PreferenceBucket) []PreferenceBucket {
	out := make([]PreferenceBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceBuckets adds up the retained of every active entry.
func SumPreferenceBuckets(xs []PreferenceBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPreferenceBuckets keys the entries by ID, last write winning.
func IndexPreferenceBuckets(xs []PreferenceBucket) map[string]PreferenceBucket {
	out := make(map[string]PreferenceBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceBucket builds a valid PreferenceBucket, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceBucket(id string, at time.Time) PreferenceBucket {
	return PreferenceBucket{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Owner:    "seeded by preference.SamplePreferenceBucket",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceLaneSchema is the stored shape of a PreferenceLane.
var PreferenceLaneSchema = schema.New("preference.preferencelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceLane is one preference lane.
type PreferenceLane struct {
	ID       string
	Name     string
	Channel  string
	Label    string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceLane) Key() string { return "preference/preferencelane/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceLane decodes a stored record.
func ParsePreferenceLane(f codec.Fields) (PreferenceLane, error) {
	var x PreferenceLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencelane: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencelane %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("preference: preferencelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceLane) String() string {
	return fmt.Sprintf("PreferenceLane(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceLane) Normalize() PreferenceLane {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortPreferenceLanes orders by attempts descending, then by ID.
func SortPreferenceLanes(xs []PreferenceLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceLanes drops the archived entries, preserving order.
func ActivePreferenceLanes(xs []PreferenceLane) []PreferenceLane {
	out := make([]PreferenceLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceLanes adds up the attempts of every active entry.
func SumPreferenceLanes(xs []PreferenceLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPreferenceLanes keys the entries by ID, last write winning.
func IndexPreferenceLanes(xs []PreferenceLane) map[string]PreferenceLane {
	out := make(map[string]PreferenceLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceLane builds a valid PreferenceLane, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceLane(id string, at time.Time) PreferenceLane {
	return PreferenceLane{
		ID:       id,
		Name:     "name-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by preference.SamplePreferenceLane",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceSlotSchema is the stored shape of a PreferenceSlot.
var PreferenceSlotSchema = schema.New("preference.preferenceslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceSlot is one preference slot.
type PreferenceSlot struct {
	ID       string
	Group    string
	Name     string
	Channel  string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceSlot) Key() string { return "preference/preferenceslot/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceSlot decodes a stored record.
func ParsePreferenceSlot(f codec.Fields) (PreferenceSlot, error) {
	var x PreferenceSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferenceslot: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceslot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferenceslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferenceslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceSlot) String() string {
	return fmt.Sprintf("PreferenceSlot(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceSlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceSlot) Normalize() PreferenceSlot {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPreferenceSlots orders by count descending, then by ID.
func SortPreferenceSlots(xs []PreferenceSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceSlots drops the archived entries, preserving order.
func ActivePreferenceSlots(xs []PreferenceSlot) []PreferenceSlot {
	out := make([]PreferenceSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceSlots adds up the count of every active entry.
func SumPreferenceSlots(xs []PreferenceSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPreferenceSlots keys the entries by ID, last write winning.
func IndexPreferenceSlots(xs []PreferenceSlot) map[string]PreferenceSlot {
	out := make(map[string]PreferenceSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceSlot builds a valid PreferenceSlot, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceSlot(id string, at time.Time) PreferenceSlot {
	return PreferenceSlot{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by preference.SamplePreferenceSlot",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceTagSchema is the stored shape of a PreferenceTag.
var PreferenceTagSchema = schema.New("preference.preferencetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceTag is one preference tag.
type PreferenceTag struct {
	ID       string
	Owner    string
	Channel  string
	Label    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceTag) Key() string { return "preference/preferencetag/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceTag decodes a stored record.
func ParsePreferenceTag(f codec.Fields) (PreferenceTag, error) {
	var x PreferenceTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencetag: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("preference: preferencetag %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("preference: preferencetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceTag) String() string {
	return fmt.Sprintf("PreferenceTag(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceTag) Normalize() PreferenceTag {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortPreferenceTags orders by count descending, then by ID.
func SortPreferenceTags(xs []PreferenceTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceTags drops the archived entries, preserving order.
func ActivePreferenceTags(xs []PreferenceTag) []PreferenceTag {
	out := make([]PreferenceTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceTags adds up the count of every active entry.
func SumPreferenceTags(xs []PreferenceTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPreferenceTags keys the entries by ID, last write winning.
func IndexPreferenceTags(xs []PreferenceTag) map[string]PreferenceTag {
	out := make(map[string]PreferenceTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceTag builds a valid PreferenceTag, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceTag(id string, at time.Time) PreferenceTag {
	return PreferenceTag{
		ID:      id,
		Owner:   "owner-" + id,
		Channel: "channel-" + id,
		Label:   "seeded by preference.SamplePreferenceTag",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// PreferenceRuleSchema is the stored shape of a PreferenceRule.
var PreferenceRuleSchema = schema.New("preference.preferencerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceRule is one preference rule.
type PreferenceRule struct {
	ID       string
	Owner    string
	Group    string
	Category string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceRule) Key() string { return "preference/preferencerule/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"category": x.Category,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceRule decodes a stored record.
func ParsePreferenceRule(f codec.Fields) (PreferenceRule, error) {
	var x PreferenceRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencerule: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerule %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("preference: preferencerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceRule) String() string {
	return fmt.Sprintf("PreferenceRule(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceRule) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceRule) Normalize() PreferenceRule {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortPreferenceRules orders by priority descending, then by ID.
func SortPreferenceRules(xs []PreferenceRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceRules drops the archived entries, preserving order.
func ActivePreferenceRules(xs []PreferenceRule) []PreferenceRule {
	out := make([]PreferenceRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceRules adds up the priority of every active entry.
func SumPreferenceRules(xs []PreferenceRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexPreferenceRules keys the entries by ID, last write winning.
func IndexPreferenceRules(xs []PreferenceRule) map[string]PreferenceRule {
	out := make(map[string]PreferenceRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceRule builds a valid PreferenceRule, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceRule(id string, at time.Time) PreferenceRule {
	return PreferenceRule{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Category: "seeded by preference.SamplePreferenceRule",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceHoldSchema is the stored shape of a PreferenceHold.
var PreferenceHoldSchema = schema.New("preference.preferencehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceHold is one preference hold.
type PreferenceHold struct {
	ID       string
	Source   string
	Region   string
	Owner    string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceHold) Key() string { return "preference/preferencehold/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"region":   x.Region,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceHold decodes a stored record.
func ParsePreferenceHold(f codec.Fields) (PreferenceHold, error) {
	var x PreferenceHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencehold: missing id")
	}
	x.Source = f["source"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("preference: preferencehold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("preference: preferencehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceHold) String() string {
	return fmt.Sprintf("PreferenceHold(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceHold) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceHold) Normalize() PreferenceHold {
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortPreferenceHolds orders by attempts descending, then by ID.
func SortPreferenceHolds(xs []PreferenceHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceHolds drops the archived entries, preserving order.
func ActivePreferenceHolds(xs []PreferenceHold) []PreferenceHold {
	out := make([]PreferenceHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceHolds adds up the attempts of every active entry.
func SumPreferenceHolds(xs []PreferenceHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPreferenceHolds keys the entries by ID, last write winning.
func IndexPreferenceHolds(xs []PreferenceHold) map[string]PreferenceHold {
	out := make(map[string]PreferenceHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceHold builds a valid PreferenceHold, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceHold(id string, at time.Time) PreferenceHold {
	return PreferenceHold{
		ID:       id,
		Source:   "source-" + id,
		Region:   "region-" + id,
		Owner:    "seeded by preference.SamplePreferenceHold",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceClaimSchema is the stored shape of a PreferenceClaim.
var PreferenceClaimSchema = schema.New("preference.preferenceclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceClaim is one preference claim.
type PreferenceClaim struct {
	ID       string
	Channel  string
	Owner    string
	Category string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceClaim) Key() string { return "preference/preferenceclaim/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceClaim decodes a stored record.
func ParsePreferenceClaim(f codec.Fields) (PreferenceClaim, error) {
	var x PreferenceClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferenceclaim: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceclaim %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("preference: preferenceclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferenceclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferenceclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceClaim) String() string {
	return fmt.Sprintf("PreferenceClaim(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceClaim) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceClaim) Normalize() PreferenceClaim {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPreferenceClaims orders by version descending, then by ID.
func SortPreferenceClaims(xs []PreferenceClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceClaims drops the archived entries, preserving order.
func ActivePreferenceClaims(xs []PreferenceClaim) []PreferenceClaim {
	out := make([]PreferenceClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceClaims adds up the version of every active entry.
func SumPreferenceClaims(xs []PreferenceClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPreferenceClaims keys the entries by ID, last write winning.
func IndexPreferenceClaims(xs []PreferenceClaim) map[string]PreferenceClaim {
	out := make(map[string]PreferenceClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceClaim builds a valid PreferenceClaim, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceClaim(id string, at time.Time) PreferenceClaim {
	return PreferenceClaim{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by preference.SamplePreferenceClaim",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PreferenceCursorSchema is the stored shape of a PreferenceCursor.
var PreferenceCursorSchema = schema.New("preference.preferencecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PreferenceCursor is one preference cursor.
type PreferenceCursor struct {
	ID       string
	Region   string
	Category string
	Comment  string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PreferenceCursor) Key() string { return "preference/preferencecursor/" + x.ID }

// Fields encodes x for storage.
func (x PreferenceCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"comment":  x.Comment,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePreferenceCursor decodes a stored record.
func ParsePreferenceCursor(f codec.Fields) (PreferenceCursor, error) {
	var x PreferenceCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("preference: preferencecursor: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("preference: preferencecursor %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("preference: preferencecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("preference: preferencecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("preference: preferencecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PreferenceCursor) String() string {
	return fmt.Sprintf("PreferenceCursor(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PreferenceCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PreferenceCursor) Normalize() PreferenceCursor {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPreferenceCursors orders by weight descending, then by ID.
func SortPreferenceCursors(xs []PreferenceCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePreferenceCursors drops the archived entries, preserving order.
func ActivePreferenceCursors(xs []PreferenceCursor) []PreferenceCursor {
	out := make([]PreferenceCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPreferenceCursors adds up the weight of every active entry.
func SumPreferenceCursors(xs []PreferenceCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexPreferenceCursors keys the entries by ID, last write winning.
func IndexPreferenceCursors(xs []PreferenceCursor) map[string]PreferenceCursor {
	out := make(map[string]PreferenceCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePreferenceCursor builds a valid PreferenceCursor, for tests and for the seed data the cron
// binary writes.
func SamplePreferenceCursor(id string, at time.Time) PreferenceCursor {
	return PreferenceCursor{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Comment:  "seeded by preference.SamplePreferenceCursor",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	PreferenceRecordSchema,
	PreferenceEntrySchema,
	PreferenceItemSchema,
	PreferenceDraftSchema,
	PreferenceRevisionSchema,
	PreferenceSnapshotSchema,
	PreferenceBatchSchema,
	PreferenceRequestSchema,
	PreferenceSummarySchema,
	PreferencePolicySchema,
	PreferenceLinkSchema,
	PreferenceNoteSchema,
	PreferenceWindowSchema,
	PreferenceQuotaSchema,
	PreferenceBucketSchema,
	PreferenceLaneSchema,
	PreferenceSlotSchema,
	PreferenceTagSchema,
	PreferenceRuleSchema,
	PreferenceHoldSchema,
	PreferenceClaimSchema,
	PreferenceCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"preference/preferencerecord/",
	"preference/preferenceentry/",
	"preference/preferenceitem/",
	"preference/preferencedraft/",
	"preference/preferencerevision/",
	"preference/preferencesnapshot/",
	"preference/preferencebatch/",
	"preference/preferencerequest/",
	"preference/preferencesummary/",
	"preference/preferencepolicy/",
	"preference/preferencelink/",
	"preference/preferencenote/",
	"preference/preferencewindow/",
	"preference/preferencequota/",
	"preference/preferencebucket/",
	"preference/preferencelane/",
	"preference/preferenceslot/",
	"preference/preferencetag/",
	"preference/preferencerule/",
	"preference/preferencehold/",
	"preference/preferenceclaim/",
	"preference/preferencecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "preference: " + strings.Join(names, ", ")
}
