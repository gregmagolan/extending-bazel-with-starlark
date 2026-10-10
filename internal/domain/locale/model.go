// Package locale holds the locale domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package locale

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// LocaleRecordSchema is the stored shape of a LocaleRecord.
var LocaleRecordSchema = schema.New("locale.localerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleRecord is one locale record.
type LocaleRecord struct {
	ID       string
	Source   string
	Owner    string
	Name     string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleRecord) Key() string { return "locale/localerecord/" + x.ID }

// Fields encodes x for storage.
func (x LocaleRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleRecord decodes a stored record.
func ParseLocaleRecord(f codec.Fields) (LocaleRecord, error) {
	var x LocaleRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localerecord: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localerecord %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("locale: localerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleRecord) String() string {
	return fmt.Sprintf("LocaleRecord(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleRecord) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleRecord) Normalize() LocaleRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortLocaleRecords orders by attempts descending, then by ID.
func SortLocaleRecords(xs []LocaleRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleRecords drops the archived entries, preserving order.
func ActiveLocaleRecords(xs []LocaleRecord) []LocaleRecord {
	out := make([]LocaleRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleRecords adds up the attempts of every active entry.
func SumLocaleRecords(xs []LocaleRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLocaleRecords keys the entries by ID, last write winning.
func IndexLocaleRecords(xs []LocaleRecord) map[string]LocaleRecord {
	out := make(map[string]LocaleRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleRecord builds a valid LocaleRecord, for tests and for the seed data the cron
// binary writes.
func SampleLocaleRecord(id string, at time.Time) LocaleRecord {
	return LocaleRecord{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by locale.SampleLocaleRecord",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleEntrySchema is the stored shape of a LocaleEntry.
var LocaleEntrySchema = schema.New("locale.localeentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleEntry is one locale entry.
type LocaleEntry struct {
	ID        string
	Group     string
	Category  string
	Reference string
	Capacity  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LocaleEntry) Key() string { return "locale/localeentry/" + x.ID }

// Fields encodes x for storage.
func (x LocaleEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"category":  x.Category,
		"reference": x.Reference,
		"capacity":  strconv.Itoa(x.Capacity),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleEntry decodes a stored record.
func ParseLocaleEntry(f codec.Fields) (LocaleEntry, error) {
	var x LocaleEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localeentry: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("locale: localeentry %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localeentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localeentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localeentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleEntry) String() string {
	return fmt.Sprintf("LocaleEntry(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleEntry) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleEntry) Normalize() LocaleEntry {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortLocaleEntrys orders by capacity descending, then by ID.
func SortLocaleEntrys(xs []LocaleEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleEntrys drops the archived entries, preserving order.
func ActiveLocaleEntrys(xs []LocaleEntry) []LocaleEntry {
	out := make([]LocaleEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleEntrys adds up the capacity of every active entry.
func SumLocaleEntrys(xs []LocaleEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexLocaleEntrys keys the entries by ID, last write winning.
func IndexLocaleEntrys(xs []LocaleEntry) map[string]LocaleEntry {
	out := make(map[string]LocaleEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleEntry builds a valid LocaleEntry, for tests and for the seed data the cron
// binary writes.
func SampleLocaleEntry(id string, at time.Time) LocaleEntry {
	return LocaleEntry{
		ID:        id,
		Group:     "group-" + id,
		Category:  "category-" + id,
		Reference: "seeded by locale.SampleLocaleEntry",
		Capacity:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LocaleItemSchema is the stored shape of a LocaleItem.
var LocaleItemSchema = schema.New("locale.localeitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleItem is one locale item.
type LocaleItem struct {
	ID       string
	Region   string
	Owner    string
	Category string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleItem) Key() string { return "locale/localeitem/" + x.ID }

// Fields encodes x for storage.
func (x LocaleItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"owner":    x.Owner,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleItem decodes a stored record.
func ParseLocaleItem(f codec.Fields) (LocaleItem, error) {
	var x LocaleItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localeitem: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("locale: localeitem %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("locale: localeitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localeitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localeitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleItem) String() string {
	return fmt.Sprintf("LocaleItem(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleItem) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleItem) Normalize() LocaleItem {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortLocaleItems orders by capacity descending, then by ID.
func SortLocaleItems(xs []LocaleItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleItems drops the archived entries, preserving order.
func ActiveLocaleItems(xs []LocaleItem) []LocaleItem {
	out := make([]LocaleItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleItems adds up the capacity of every active entry.
func SumLocaleItems(xs []LocaleItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexLocaleItems keys the entries by ID, last write winning.
func IndexLocaleItems(xs []LocaleItem) map[string]LocaleItem {
	out := make(map[string]LocaleItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleItem builds a valid LocaleItem, for tests and for the seed data the cron
// binary writes.
func SampleLocaleItem(id string, at time.Time) LocaleItem {
	return LocaleItem{
		ID:       id,
		Region:   "region-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by locale.SampleLocaleItem",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleDraftSchema is the stored shape of a LocaleDraft.
var LocaleDraftSchema = schema.New("locale.localedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleDraft is one locale draft.
type LocaleDraft struct {
	ID       string
	Owner    string
	Channel  string
	Name     string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleDraft) Key() string { return "locale/localedraft/" + x.ID }

// Fields encodes x for storage.
func (x LocaleDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"name":     x.Name,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleDraft decodes a stored record.
func ParseLocaleDraft(f codec.Fields) (LocaleDraft, error) {
	var x LocaleDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localedraft: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localedraft %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("locale: localedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleDraft) String() string {
	return fmt.Sprintf("LocaleDraft(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleDraft) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleDraft) Normalize() LocaleDraft {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortLocaleDrafts orders by offset descending, then by ID.
func SortLocaleDrafts(xs []LocaleDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleDrafts drops the archived entries, preserving order.
func ActiveLocaleDrafts(xs []LocaleDraft) []LocaleDraft {
	out := make([]LocaleDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleDrafts adds up the offset of every active entry.
func SumLocaleDrafts(xs []LocaleDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexLocaleDrafts keys the entries by ID, last write winning.
func IndexLocaleDrafts(xs []LocaleDraft) map[string]LocaleDraft {
	out := make(map[string]LocaleDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleDraft builds a valid LocaleDraft, for tests and for the seed data the cron
// binary writes.
func SampleLocaleDraft(id string, at time.Time) LocaleDraft {
	return LocaleDraft{
		ID:       id,
		Owner:    "owner-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by locale.SampleLocaleDraft",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleRevisionSchema is the stored shape of a LocaleRevision.
var LocaleRevisionSchema = schema.New("locale.localerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleRevision is one locale revision.
type LocaleRevision struct {
	ID       string
	Source   string
	Name     string
	Label    string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleRevision) Key() string { return "locale/localerevision/" + x.ID }

// Fields encodes x for storage.
func (x LocaleRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleRevision decodes a stored record.
func ParseLocaleRevision(f codec.Fields) (LocaleRevision, error) {
	var x LocaleRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localerevision: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localerevision %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleRevision) String() string {
	return fmt.Sprintf("LocaleRevision(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleRevision) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleRevision) Normalize() LocaleRevision {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortLocaleRevisions orders by attempts descending, then by ID.
func SortLocaleRevisions(xs []LocaleRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleRevisions drops the archived entries, preserving order.
func ActiveLocaleRevisions(xs []LocaleRevision) []LocaleRevision {
	out := make([]LocaleRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleRevisions adds up the attempts of every active entry.
func SumLocaleRevisions(xs []LocaleRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLocaleRevisions keys the entries by ID, last write winning.
func IndexLocaleRevisions(xs []LocaleRevision) map[string]LocaleRevision {
	out := make(map[string]LocaleRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleRevision builds a valid LocaleRevision, for tests and for the seed data the cron
// binary writes.
func SampleLocaleRevision(id string, at time.Time) LocaleRevision {
	return LocaleRevision{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Label:    "seeded by locale.SampleLocaleRevision",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleSnapshotSchema is the stored shape of a LocaleSnapshot.
var LocaleSnapshotSchema = schema.New("locale.localesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleSnapshot is one locale snapshot.
type LocaleSnapshot struct {
	ID        string
	Category  string
	Reference string
	Name      string
	Retained  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LocaleSnapshot) Key() string { return "locale/localesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x LocaleSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"name":      x.Name,
		"retained":  strconv.Itoa(x.Retained),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleSnapshot decodes a stored record.
func ParseLocaleSnapshot(f codec.Fields) (LocaleSnapshot, error) {
	var x LocaleSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localesnapshot: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("locale: localesnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("locale: localesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleSnapshot) String() string {
	return fmt.Sprintf("LocaleSnapshot(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleSnapshot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleSnapshot) Normalize() LocaleSnapshot {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortLocaleSnapshots orders by retained descending, then by ID.
func SortLocaleSnapshots(xs []LocaleSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleSnapshots drops the archived entries, preserving order.
func ActiveLocaleSnapshots(xs []LocaleSnapshot) []LocaleSnapshot {
	out := make([]LocaleSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleSnapshots adds up the retained of every active entry.
func SumLocaleSnapshots(xs []LocaleSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLocaleSnapshots keys the entries by ID, last write winning.
func IndexLocaleSnapshots(xs []LocaleSnapshot) map[string]LocaleSnapshot {
	out := make(map[string]LocaleSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleSnapshot builds a valid LocaleSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleLocaleSnapshot(id string, at time.Time) LocaleSnapshot {
	return LocaleSnapshot{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by locale.SampleLocaleSnapshot",
		Retained:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LocaleBatchSchema is the stored shape of a LocaleBatch.
var LocaleBatchSchema = schema.New("locale.localebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleBatch is one locale batch.
type LocaleBatch struct {
	ID       string
	Group    string
	Channel  string
	Owner    string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleBatch) Key() string { return "locale/localebatch/" + x.ID }

// Fields encodes x for storage.
func (x LocaleBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleBatch decodes a stored record.
func ParseLocaleBatch(f codec.Fields) (LocaleBatch, error) {
	var x LocaleBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localebatch: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("locale: localebatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("locale: localebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleBatch) String() string {
	return fmt.Sprintf("LocaleBatch(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleBatch) Normalize() LocaleBatch {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortLocaleBatchs orders by capacity descending, then by ID.
func SortLocaleBatchs(xs []LocaleBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleBatchs drops the archived entries, preserving order.
func ActiveLocaleBatchs(xs []LocaleBatch) []LocaleBatch {
	out := make([]LocaleBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleBatchs adds up the capacity of every active entry.
func SumLocaleBatchs(xs []LocaleBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexLocaleBatchs keys the entries by ID, last write winning.
func IndexLocaleBatchs(xs []LocaleBatch) map[string]LocaleBatch {
	out := make(map[string]LocaleBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleBatch builds a valid LocaleBatch, for tests and for the seed data the cron
// binary writes.
func SampleLocaleBatch(id string, at time.Time) LocaleBatch {
	return LocaleBatch{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by locale.SampleLocaleBatch",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleRequestSchema is the stored shape of a LocaleRequest.
var LocaleRequestSchema = schema.New("locale.localerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleRequest is one locale request.
type LocaleRequest struct {
	ID       string
	Source   string
	Name     string
	Region   string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleRequest) Key() string { return "locale/localerequest/" + x.ID }

// Fields encodes x for storage.
func (x LocaleRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleRequest decodes a stored record.
func ParseLocaleRequest(f codec.Fields) (LocaleRequest, error) {
	var x LocaleRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localerequest: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localerequest %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("locale: localerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleRequest) String() string {
	return fmt.Sprintf("LocaleRequest(%s, source=%q, offset=%d, archived=%t)", x.ID, x.Source, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleRequest) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleRequest) Normalize() LocaleRequest {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortLocaleRequests orders by offset descending, then by ID.
func SortLocaleRequests(xs []LocaleRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleRequests drops the archived entries, preserving order.
func ActiveLocaleRequests(xs []LocaleRequest) []LocaleRequest {
	out := make([]LocaleRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleRequests adds up the offset of every active entry.
func SumLocaleRequests(xs []LocaleRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexLocaleRequests keys the entries by ID, last write winning.
func IndexLocaleRequests(xs []LocaleRequest) map[string]LocaleRequest {
	out := make(map[string]LocaleRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleRequest builds a valid LocaleRequest, for tests and for the seed data the cron
// binary writes.
func SampleLocaleRequest(id string, at time.Time) LocaleRequest {
	return LocaleRequest{
		ID:      id,
		Source:  "source-" + id,
		Name:    "name-" + id,
		Region:  "seeded by locale.SampleLocaleRequest",
		Offset:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// LocaleSummarySchema is the stored shape of a LocaleSummary.
var LocaleSummarySchema = schema.New("locale.localesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleSummary is one locale summary.
type LocaleSummary struct {
	ID       string
	Source   string
	Comment  string
	Group    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleSummary) Key() string { return "locale/localesummary/" + x.ID }

// Fields encodes x for storage.
func (x LocaleSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleSummary decodes a stored record.
func ParseLocaleSummary(f codec.Fields) (LocaleSummary, error) {
	var x LocaleSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localesummary: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("locale: localesummary %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleSummary) String() string {
	return fmt.Sprintf("LocaleSummary(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleSummary) Normalize() LocaleSummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLocaleSummarys orders by capacity descending, then by ID.
func SortLocaleSummarys(xs []LocaleSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleSummarys drops the archived entries, preserving order.
func ActiveLocaleSummarys(xs []LocaleSummary) []LocaleSummary {
	out := make([]LocaleSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleSummarys adds up the capacity of every active entry.
func SumLocaleSummarys(xs []LocaleSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexLocaleSummarys keys the entries by ID, last write winning.
func IndexLocaleSummarys(xs []LocaleSummary) map[string]LocaleSummary {
	out := make(map[string]LocaleSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleSummary builds a valid LocaleSummary, for tests and for the seed data the cron
// binary writes.
func SampleLocaleSummary(id string, at time.Time) LocaleSummary {
	return LocaleSummary{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by locale.SampleLocaleSummary",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocalePolicySchema is the stored shape of a LocalePolicy.
var LocalePolicySchema = schema.New("locale.localepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocalePolicy is one locale policy.
type LocalePolicy struct {
	ID       string
	Category string
	Name     string
	Comment  string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocalePolicy) Key() string { return "locale/localepolicy/" + x.ID }

// Fields encodes x for storage.
func (x LocalePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"comment":  x.Comment,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocalePolicy decodes a stored record.
func ParseLocalePolicy(f codec.Fields) (LocalePolicy, error) {
	var x LocalePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localepolicy: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localepolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocalePolicy) String() string {
	return fmt.Sprintf("LocalePolicy(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocalePolicy) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocalePolicy) Normalize() LocalePolicy {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLocalePolicys orders by offset descending, then by ID.
func SortLocalePolicys(xs []LocalePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocalePolicys drops the archived entries, preserving order.
func ActiveLocalePolicys(xs []LocalePolicy) []LocalePolicy {
	out := make([]LocalePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocalePolicys adds up the offset of every active entry.
func SumLocalePolicys(xs []LocalePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexLocalePolicys keys the entries by ID, last write winning.
func IndexLocalePolicys(xs []LocalePolicy) map[string]LocalePolicy {
	out := make(map[string]LocalePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocalePolicy builds a valid LocalePolicy, for tests and for the seed data the cron
// binary writes.
func SampleLocalePolicy(id string, at time.Time) LocalePolicy {
	return LocalePolicy{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by locale.SampleLocalePolicy",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleLinkSchema is the stored shape of a LocaleLink.
var LocaleLinkSchema = schema.New("locale.localelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleLink is one locale link.
type LocaleLink struct {
	ID        string
	Owner     string
	Source    string
	Reference string
	Offset    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LocaleLink) Key() string { return "locale/localelink/" + x.ID }

// Fields encodes x for storage.
func (x LocaleLink) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"source":    x.Source,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleLink decodes a stored record.
func ParseLocaleLink(f codec.Fields) (LocaleLink, error) {
	var x LocaleLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localelink: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localelink %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("locale: localelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleLink) String() string {
	return fmt.Sprintf("LocaleLink(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleLink) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleLink) Normalize() LocaleLink {
	x.Owner = strings.TrimSpace(x.Owner)
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

// SortLocaleLinks orders by offset descending, then by ID.
func SortLocaleLinks(xs []LocaleLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleLinks drops the archived entries, preserving order.
func ActiveLocaleLinks(xs []LocaleLink) []LocaleLink {
	out := make([]LocaleLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleLinks adds up the offset of every active entry.
func SumLocaleLinks(xs []LocaleLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexLocaleLinks keys the entries by ID, last write winning.
func IndexLocaleLinks(xs []LocaleLink) map[string]LocaleLink {
	out := make(map[string]LocaleLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleLink builds a valid LocaleLink, for tests and for the seed data the cron
// binary writes.
func SampleLocaleLink(id string, at time.Time) LocaleLink {
	return LocaleLink{
		ID:        id,
		Owner:     "owner-" + id,
		Source:    "source-" + id,
		Reference: "seeded by locale.SampleLocaleLink",
		Offset:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LocaleNoteSchema is the stored shape of a LocaleNote.
var LocaleNoteSchema = schema.New("locale.localenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleNote is one locale note.
type LocaleNote struct {
	ID       string
	Comment  string
	Category string
	Region   string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleNote) Key() string { return "locale/localenote/" + x.ID }

// Fields encodes x for storage.
func (x LocaleNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"region":   x.Region,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleNote decodes a stored record.
func ParseLocaleNote(f codec.Fields) (LocaleNote, error) {
	var x LocaleNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localenote: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("locale: localenote %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleNote) String() string {
	return fmt.Sprintf("LocaleNote(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleNote) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleNote) Normalize() LocaleNote {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLocaleNotes orders by priority descending, then by ID.
func SortLocaleNotes(xs []LocaleNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleNotes drops the archived entries, preserving order.
func ActiveLocaleNotes(xs []LocaleNote) []LocaleNote {
	out := make([]LocaleNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleNotes adds up the priority of every active entry.
func SumLocaleNotes(xs []LocaleNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexLocaleNotes keys the entries by ID, last write winning.
func IndexLocaleNotes(xs []LocaleNote) map[string]LocaleNote {
	out := make(map[string]LocaleNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleNote builds a valid LocaleNote, for tests and for the seed data the cron
// binary writes.
func SampleLocaleNote(id string, at time.Time) LocaleNote {
	return LocaleNote{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Region:   "seeded by locale.SampleLocaleNote",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleWindowSchema is the stored shape of a LocaleWindow.
var LocaleWindowSchema = schema.New("locale.localewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleWindow is one locale window.
type LocaleWindow struct {
	ID       string
	Source   string
	Label    string
	Comment  string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleWindow) Key() string { return "locale/localewindow/" + x.ID }

// Fields encodes x for storage.
func (x LocaleWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"label":    x.Label,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleWindow decodes a stored record.
func ParseLocaleWindow(f codec.Fields) (LocaleWindow, error) {
	var x LocaleWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localewindow: missing id")
	}
	x.Source = f["source"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("locale: localewindow %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("locale: localewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleWindow) String() string {
	return fmt.Sprintf("LocaleWindow(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleWindow) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleWindow) Normalize() LocaleWindow {
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortLocaleWindows orders by retained descending, then by ID.
func SortLocaleWindows(xs []LocaleWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleWindows drops the archived entries, preserving order.
func ActiveLocaleWindows(xs []LocaleWindow) []LocaleWindow {
	out := make([]LocaleWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleWindows adds up the retained of every active entry.
func SumLocaleWindows(xs []LocaleWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLocaleWindows keys the entries by ID, last write winning.
func IndexLocaleWindows(xs []LocaleWindow) map[string]LocaleWindow {
	out := make(map[string]LocaleWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleWindow builds a valid LocaleWindow, for tests and for the seed data the cron
// binary writes.
func SampleLocaleWindow(id string, at time.Time) LocaleWindow {
	return LocaleWindow{
		ID:       id,
		Source:   "source-" + id,
		Label:    "label-" + id,
		Comment:  "seeded by locale.SampleLocaleWindow",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleQuotaSchema is the stored shape of a LocaleQuota.
var LocaleQuotaSchema = schema.New("locale.localequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleQuota is one locale quota.
type LocaleQuota struct {
	ID       string
	Owner    string
	Label    string
	Channel  string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleQuota) Key() string { return "locale/localequota/" + x.ID }

// Fields encodes x for storage.
func (x LocaleQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"label":    x.Label,
		"channel":  x.Channel,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleQuota decodes a stored record.
func ParseLocaleQuota(f codec.Fields) (LocaleQuota, error) {
	var x LocaleQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localequota: missing id")
	}
	x.Owner = f["owner"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localequota %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleQuota) String() string {
	return fmt.Sprintf("LocaleQuota(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleQuota) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleQuota) Normalize() LocaleQuota {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortLocaleQuotas orders by attempts descending, then by ID.
func SortLocaleQuotas(xs []LocaleQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleQuotas drops the archived entries, preserving order.
func ActiveLocaleQuotas(xs []LocaleQuota) []LocaleQuota {
	out := make([]LocaleQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleQuotas adds up the attempts of every active entry.
func SumLocaleQuotas(xs []LocaleQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLocaleQuotas keys the entries by ID, last write winning.
func IndexLocaleQuotas(xs []LocaleQuota) map[string]LocaleQuota {
	out := make(map[string]LocaleQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleQuota builds a valid LocaleQuota, for tests and for the seed data the cron
// binary writes.
func SampleLocaleQuota(id string, at time.Time) LocaleQuota {
	return LocaleQuota{
		ID:       id,
		Owner:    "owner-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by locale.SampleLocaleQuota",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleBucketSchema is the stored shape of a LocaleBucket.
var LocaleBucketSchema = schema.New("locale.localebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleBucket is one locale bucket.
type LocaleBucket struct {
	ID       string
	Name     string
	Group    string
	Source   string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleBucket) Key() string { return "locale/localebucket/" + x.ID }

// Fields encodes x for storage.
func (x LocaleBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"group":    x.Group,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleBucket decodes a stored record.
func ParseLocaleBucket(f codec.Fields) (LocaleBucket, error) {
	var x LocaleBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localebucket: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localebucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("locale: localebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleBucket) String() string {
	return fmt.Sprintf("LocaleBucket(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleBucket) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleBucket) Normalize() LocaleBucket {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortLocaleBuckets orders by attempts descending, then by ID.
func SortLocaleBuckets(xs []LocaleBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleBuckets drops the archived entries, preserving order.
func ActiveLocaleBuckets(xs []LocaleBucket) []LocaleBucket {
	out := make([]LocaleBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleBuckets adds up the attempts of every active entry.
func SumLocaleBuckets(xs []LocaleBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLocaleBuckets keys the entries by ID, last write winning.
func IndexLocaleBuckets(xs []LocaleBucket) map[string]LocaleBucket {
	out := make(map[string]LocaleBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleBucket builds a valid LocaleBucket, for tests and for the seed data the cron
// binary writes.
func SampleLocaleBucket(id string, at time.Time) LocaleBucket {
	return LocaleBucket{
		ID:       id,
		Name:     "name-" + id,
		Group:    "group-" + id,
		Source:   "seeded by locale.SampleLocaleBucket",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleLaneSchema is the stored shape of a LocaleLane.
var LocaleLaneSchema = schema.New("locale.localelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleLane is one locale lane.
type LocaleLane struct {
	ID       string
	Owner    string
	Comment  string
	Channel  string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleLane) Key() string { return "locale/localelane/" + x.ID }

// Fields encodes x for storage.
func (x LocaleLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleLane decodes a stored record.
func ParseLocaleLane(f codec.Fields) (LocaleLane, error) {
	var x LocaleLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localelane: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localelane %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("locale: localelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleLane) String() string {
	return fmt.Sprintf("LocaleLane(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleLane) Normalize() LocaleLane {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortLocaleLanes orders by attempts descending, then by ID.
func SortLocaleLanes(xs []LocaleLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleLanes drops the archived entries, preserving order.
func ActiveLocaleLanes(xs []LocaleLane) []LocaleLane {
	out := make([]LocaleLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleLanes adds up the attempts of every active entry.
func SumLocaleLanes(xs []LocaleLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLocaleLanes keys the entries by ID, last write winning.
func IndexLocaleLanes(xs []LocaleLane) map[string]LocaleLane {
	out := make(map[string]LocaleLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleLane builds a valid LocaleLane, for tests and for the seed data the cron
// binary writes.
func SampleLocaleLane(id string, at time.Time) LocaleLane {
	return LocaleLane{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Channel:  "seeded by locale.SampleLocaleLane",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleSlotSchema is the stored shape of a LocaleSlot.
var LocaleSlotSchema = schema.New("locale.localeslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleSlot is one locale slot.
type LocaleSlot struct {
	ID       string
	Group    string
	Channel  string
	Region   string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleSlot) Key() string { return "locale/localeslot/" + x.ID }

// Fields encodes x for storage.
func (x LocaleSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"region":   x.Region,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleSlot decodes a stored record.
func ParseLocaleSlot(f codec.Fields) (LocaleSlot, error) {
	var x LocaleSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localeslot: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("locale: localeslot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("locale: localeslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localeslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localeslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleSlot) String() string {
	return fmt.Sprintf("LocaleSlot(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleSlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleSlot) Normalize() LocaleSlot {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLocaleSlots orders by count descending, then by ID.
func SortLocaleSlots(xs []LocaleSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleSlots drops the archived entries, preserving order.
func ActiveLocaleSlots(xs []LocaleSlot) []LocaleSlot {
	out := make([]LocaleSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleSlots adds up the count of every active entry.
func SumLocaleSlots(xs []LocaleSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexLocaleSlots keys the entries by ID, last write winning.
func IndexLocaleSlots(xs []LocaleSlot) map[string]LocaleSlot {
	out := make(map[string]LocaleSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleSlot builds a valid LocaleSlot, for tests and for the seed data the cron
// binary writes.
func SampleLocaleSlot(id string, at time.Time) LocaleSlot {
	return LocaleSlot{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by locale.SampleLocaleSlot",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleTagSchema is the stored shape of a LocaleTag.
var LocaleTagSchema = schema.New("locale.localetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleTag is one locale tag.
type LocaleTag struct {
	ID       string
	Group    string
	Owner    string
	Label    string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleTag) Key() string { return "locale/localetag/" + x.ID }

// Fields encodes x for storage.
func (x LocaleTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleTag decodes a stored record.
func ParseLocaleTag(f codec.Fields) (LocaleTag, error) {
	var x LocaleTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localetag: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("locale: localetag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("locale: localetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleTag) String() string {
	return fmt.Sprintf("LocaleTag(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleTag) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleTag) Normalize() LocaleTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortLocaleTags orders by priority descending, then by ID.
func SortLocaleTags(xs []LocaleTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleTags drops the archived entries, preserving order.
func ActiveLocaleTags(xs []LocaleTag) []LocaleTag {
	out := make([]LocaleTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleTags adds up the priority of every active entry.
func SumLocaleTags(xs []LocaleTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexLocaleTags keys the entries by ID, last write winning.
func IndexLocaleTags(xs []LocaleTag) map[string]LocaleTag {
	out := make(map[string]LocaleTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleTag builds a valid LocaleTag, for tests and for the seed data the cron
// binary writes.
func SampleLocaleTag(id string, at time.Time) LocaleTag {
	return LocaleTag{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by locale.SampleLocaleTag",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleRuleSchema is the stored shape of a LocaleRule.
var LocaleRuleSchema = schema.New("locale.localerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleRule is one locale rule.
type LocaleRule struct {
	ID        string
	Name      string
	Owner     string
	Reference string
	Version   int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LocaleRule) Key() string { return "locale/localerule/" + x.ID }

// Fields encodes x for storage.
func (x LocaleRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"owner":     x.Owner,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleRule decodes a stored record.
func ParseLocaleRule(f codec.Fields) (LocaleRule, error) {
	var x LocaleRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localerule: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("locale: localerule %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("locale: localerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleRule) String() string {
	return fmt.Sprintf("LocaleRule(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleRule) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleRule) Normalize() LocaleRule {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortLocaleRules orders by version descending, then by ID.
func SortLocaleRules(xs []LocaleRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleRules drops the archived entries, preserving order.
func ActiveLocaleRules(xs []LocaleRule) []LocaleRule {
	out := make([]LocaleRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleRules adds up the version of every active entry.
func SumLocaleRules(xs []LocaleRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexLocaleRules keys the entries by ID, last write winning.
func IndexLocaleRules(xs []LocaleRule) map[string]LocaleRule {
	out := make(map[string]LocaleRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleRule builds a valid LocaleRule, for tests and for the seed data the cron
// binary writes.
func SampleLocaleRule(id string, at time.Time) LocaleRule {
	return LocaleRule{
		ID:        id,
		Name:      "name-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by locale.SampleLocaleRule",
		Version:   len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LocaleHoldSchema is the stored shape of a LocaleHold.
var LocaleHoldSchema = schema.New("locale.localehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleHold is one locale hold.
type LocaleHold struct {
	ID       string
	Category string
	Label    string
	Comment  string
	Weight   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleHold) Key() string { return "locale/localehold/" + x.ID }

// Fields encodes x for storage.
func (x LocaleHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"comment":  x.Comment,
		"weight":   strconv.Itoa(x.Weight),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleHold decodes a stored record.
func ParseLocaleHold(f codec.Fields) (LocaleHold, error) {
	var x LocaleHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localehold: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("locale: localehold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("locale: localehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleHold) String() string {
	return fmt.Sprintf("LocaleHold(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleHold) Normalize() LocaleHold {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortLocaleHolds orders by weight descending, then by ID.
func SortLocaleHolds(xs []LocaleHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleHolds drops the archived entries, preserving order.
func ActiveLocaleHolds(xs []LocaleHold) []LocaleHold {
	out := make([]LocaleHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleHolds adds up the weight of every active entry.
func SumLocaleHolds(xs []LocaleHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexLocaleHolds keys the entries by ID, last write winning.
func IndexLocaleHolds(xs []LocaleHold) map[string]LocaleHold {
	out := make(map[string]LocaleHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleHold builds a valid LocaleHold, for tests and for the seed data the cron
// binary writes.
func SampleLocaleHold(id string, at time.Time) LocaleHold {
	return LocaleHold{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Comment:  "seeded by locale.SampleLocaleHold",
		Weight:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleClaimSchema is the stored shape of a LocaleClaim.
var LocaleClaimSchema = schema.New("locale.localeclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleClaim is one locale claim.
type LocaleClaim struct {
	ID       string
	Comment  string
	Name     string
	Owner    string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleClaim) Key() string { return "locale/localeclaim/" + x.ID }

// Fields encodes x for storage.
func (x LocaleClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"name":     x.Name,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleClaim decodes a stored record.
func ParseLocaleClaim(f codec.Fields) (LocaleClaim, error) {
	var x LocaleClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localeclaim: missing id")
	}
	x.Comment = f["comment"]
	x.Name = f["name"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("locale: localeclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("locale: localeclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localeclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localeclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleClaim) String() string {
	return fmt.Sprintf("LocaleClaim(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleClaim) Normalize() LocaleClaim {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortLocaleClaims orders by retained descending, then by ID.
func SortLocaleClaims(xs []LocaleClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleClaims drops the archived entries, preserving order.
func ActiveLocaleClaims(xs []LocaleClaim) []LocaleClaim {
	out := make([]LocaleClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleClaims adds up the retained of every active entry.
func SumLocaleClaims(xs []LocaleClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLocaleClaims keys the entries by ID, last write winning.
func IndexLocaleClaims(xs []LocaleClaim) map[string]LocaleClaim {
	out := make(map[string]LocaleClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleClaim builds a valid LocaleClaim, for tests and for the seed data the cron
// binary writes.
func SampleLocaleClaim(id string, at time.Time) LocaleClaim {
	return LocaleClaim{
		ID:       id,
		Comment:  "comment-" + id,
		Name:     "name-" + id,
		Owner:    "seeded by locale.SampleLocaleClaim",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LocaleCursorSchema is the stored shape of a LocaleCursor.
var LocaleCursorSchema = schema.New("locale.localecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LocaleCursor is one locale cursor.
type LocaleCursor struct {
	ID       string
	Channel  string
	Owner    string
	Comment  string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LocaleCursor) Key() string { return "locale/localecursor/" + x.ID }

// Fields encodes x for storage.
func (x LocaleCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLocaleCursor decodes a stored record.
func ParseLocaleCursor(f codec.Fields) (LocaleCursor, error) {
	var x LocaleCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("locale: localecursor: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("locale: localecursor %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("locale: localecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("locale: localecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("locale: localecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LocaleCursor) String() string {
	return fmt.Sprintf("LocaleCursor(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LocaleCursor) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LocaleCursor) Normalize() LocaleCursor {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortLocaleCursors orders by offset descending, then by ID.
func SortLocaleCursors(xs []LocaleCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLocaleCursors drops the archived entries, preserving order.
func ActiveLocaleCursors(xs []LocaleCursor) []LocaleCursor {
	out := make([]LocaleCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLocaleCursors adds up the offset of every active entry.
func SumLocaleCursors(xs []LocaleCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexLocaleCursors keys the entries by ID, last write winning.
func IndexLocaleCursors(xs []LocaleCursor) map[string]LocaleCursor {
	out := make(map[string]LocaleCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLocaleCursor builds a valid LocaleCursor, for tests and for the seed data the cron
// binary writes.
func SampleLocaleCursor(id string, at time.Time) LocaleCursor {
	return LocaleCursor{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Comment:  "seeded by locale.SampleLocaleCursor",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	LocaleRecordSchema,
	LocaleEntrySchema,
	LocaleItemSchema,
	LocaleDraftSchema,
	LocaleRevisionSchema,
	LocaleSnapshotSchema,
	LocaleBatchSchema,
	LocaleRequestSchema,
	LocaleSummarySchema,
	LocalePolicySchema,
	LocaleLinkSchema,
	LocaleNoteSchema,
	LocaleWindowSchema,
	LocaleQuotaSchema,
	LocaleBucketSchema,
	LocaleLaneSchema,
	LocaleSlotSchema,
	LocaleTagSchema,
	LocaleRuleSchema,
	LocaleHoldSchema,
	LocaleClaimSchema,
	LocaleCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"locale/localerecord/",
	"locale/localeentry/",
	"locale/localeitem/",
	"locale/localedraft/",
	"locale/localerevision/",
	"locale/localesnapshot/",
	"locale/localebatch/",
	"locale/localerequest/",
	"locale/localesummary/",
	"locale/localepolicy/",
	"locale/localelink/",
	"locale/localenote/",
	"locale/localewindow/",
	"locale/localequota/",
	"locale/localebucket/",
	"locale/localelane/",
	"locale/localeslot/",
	"locale/localetag/",
	"locale/localerule/",
	"locale/localehold/",
	"locale/localeclaim/",
	"locale/localecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "locale: " + strings.Join(names, ", ")
}
