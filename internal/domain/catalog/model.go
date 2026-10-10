// Package catalog holds the catalog domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package catalog

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// CatalogRecordSchema is the stored shape of a CatalogRecord.
var CatalogRecordSchema = schema.New("catalog.catalogrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogRecord is one catalog record.
type CatalogRecord struct {
	ID       string
	Name     string
	Channel  string
	Owner    string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogRecord) Key() string { return "catalog/catalogrecord/" + x.ID }

// Fields encodes x for storage.
func (x CatalogRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogRecord decodes a stored record.
func ParseCatalogRecord(f codec.Fields) (CatalogRecord, error) {
	var x CatalogRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogrecord: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrecord %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogRecord) String() string {
	return fmt.Sprintf("CatalogRecord(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogRecord) Normalize() CatalogRecord {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortCatalogRecords orders by capacity descending, then by ID.
func SortCatalogRecords(xs []CatalogRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogRecords drops the archived entries, preserving order.
func ActiveCatalogRecords(xs []CatalogRecord) []CatalogRecord {
	out := make([]CatalogRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogRecords adds up the capacity of every active entry.
func SumCatalogRecords(xs []CatalogRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCatalogRecords keys the entries by ID, last write winning.
func IndexCatalogRecords(xs []CatalogRecord) map[string]CatalogRecord {
	out := make(map[string]CatalogRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogRecord builds a valid CatalogRecord, for tests and for the seed data the cron
// binary writes.
func SampleCatalogRecord(id string, at time.Time) CatalogRecord {
	return CatalogRecord{
		ID:       id,
		Name:     "name-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by catalog.SampleCatalogRecord",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogEntrySchema is the stored shape of a CatalogEntry.
var CatalogEntrySchema = schema.New("catalog.catalogentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogEntry is one catalog entry.
type CatalogEntry struct {
	ID       string
	Label    string
	Region   string
	Category string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogEntry) Key() string { return "catalog/catalogentry/" + x.ID }

// Fields encodes x for storage.
func (x CatalogEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogEntry decodes a stored record.
func ParseCatalogEntry(f codec.Fields) (CatalogEntry, error) {
	var x CatalogEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogentry: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogEntry) String() string {
	return fmt.Sprintf("CatalogEntry(%s, label=%q, version=%d, archived=%t)", x.ID, x.Label, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogEntry) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogEntry) Normalize() CatalogEntry {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortCatalogEntrys orders by version descending, then by ID.
func SortCatalogEntrys(xs []CatalogEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogEntrys drops the archived entries, preserving order.
func ActiveCatalogEntrys(xs []CatalogEntry) []CatalogEntry {
	out := make([]CatalogEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogEntrys adds up the version of every active entry.
func SumCatalogEntrys(xs []CatalogEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexCatalogEntrys keys the entries by ID, last write winning.
func IndexCatalogEntrys(xs []CatalogEntry) map[string]CatalogEntry {
	out := make(map[string]CatalogEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogEntry builds a valid CatalogEntry, for tests and for the seed data the cron
// binary writes.
func SampleCatalogEntry(id string, at time.Time) CatalogEntry {
	return CatalogEntry{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Category: "seeded by catalog.SampleCatalogEntry",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogItemSchema is the stored shape of a CatalogItem.
var CatalogItemSchema = schema.New("catalog.catalogitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogItem is one catalog item.
type CatalogItem struct {
	ID       string
	Comment  string
	Group    string
	Owner    string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogItem) Key() string { return "catalog/catalogitem/" + x.ID }

// Fields encodes x for storage.
func (x CatalogItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogItem decodes a stored record.
func ParseCatalogItem(f codec.Fields) (CatalogItem, error) {
	var x CatalogItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogitem: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogItem) String() string {
	return fmt.Sprintf("CatalogItem(%s, comment=%q, capacity=%d, archived=%t)", x.ID, x.Comment, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogItem) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogItem) Normalize() CatalogItem {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortCatalogItems orders by capacity descending, then by ID.
func SortCatalogItems(xs []CatalogItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogItems drops the archived entries, preserving order.
func ActiveCatalogItems(xs []CatalogItem) []CatalogItem {
	out := make([]CatalogItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogItems adds up the capacity of every active entry.
func SumCatalogItems(xs []CatalogItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCatalogItems keys the entries by ID, last write winning.
func IndexCatalogItems(xs []CatalogItem) map[string]CatalogItem {
	out := make(map[string]CatalogItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogItem builds a valid CatalogItem, for tests and for the seed data the cron
// binary writes.
func SampleCatalogItem(id string, at time.Time) CatalogItem {
	return CatalogItem{
		ID:       id,
		Comment:  "comment-" + id,
		Group:    "group-" + id,
		Owner:    "seeded by catalog.SampleCatalogItem",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogDraftSchema is the stored shape of a CatalogDraft.
var CatalogDraftSchema = schema.New("catalog.catalogdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogDraft is one catalog draft.
type CatalogDraft struct {
	ID       string
	Source   string
	Comment  string
	Group    string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogDraft) Key() string { return "catalog/catalogdraft/" + x.ID }

// Fields encodes x for storage.
func (x CatalogDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogDraft decodes a stored record.
func ParseCatalogDraft(f codec.Fields) (CatalogDraft, error) {
	var x CatalogDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogdraft: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogdraft %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogDraft) String() string {
	return fmt.Sprintf("CatalogDraft(%s, source=%q, count=%d, archived=%t)", x.ID, x.Source, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogDraft) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogDraft) Normalize() CatalogDraft {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortCatalogDrafts orders by count descending, then by ID.
func SortCatalogDrafts(xs []CatalogDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogDrafts drops the archived entries, preserving order.
func ActiveCatalogDrafts(xs []CatalogDraft) []CatalogDraft {
	out := make([]CatalogDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogDrafts adds up the count of every active entry.
func SumCatalogDrafts(xs []CatalogDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCatalogDrafts keys the entries by ID, last write winning.
func IndexCatalogDrafts(xs []CatalogDraft) map[string]CatalogDraft {
	out := make(map[string]CatalogDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogDraft builds a valid CatalogDraft, for tests and for the seed data the cron
// binary writes.
func SampleCatalogDraft(id string, at time.Time) CatalogDraft {
	return CatalogDraft{
		ID:      id,
		Source:  "source-" + id,
		Comment: "comment-" + id,
		Group:   "seeded by catalog.SampleCatalogDraft",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// CatalogRevisionSchema is the stored shape of a CatalogRevision.
var CatalogRevisionSchema = schema.New("catalog.catalogrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogRevision is one catalog revision.
type CatalogRevision struct {
	ID       string
	Channel  string
	Region   string
	Name     string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogRevision) Key() string { return "catalog/catalogrevision/" + x.ID }

// Fields encodes x for storage.
func (x CatalogRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"name":     x.Name,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogRevision decodes a stored record.
func ParseCatalogRevision(f codec.Fields) (CatalogRevision, error) {
	var x CatalogRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogrevision: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrevision %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogRevision) String() string {
	return fmt.Sprintf("CatalogRevision(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogRevision) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogRevision) Normalize() CatalogRevision {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortCatalogRevisions orders by capacity descending, then by ID.
func SortCatalogRevisions(xs []CatalogRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogRevisions drops the archived entries, preserving order.
func ActiveCatalogRevisions(xs []CatalogRevision) []CatalogRevision {
	out := make([]CatalogRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogRevisions adds up the capacity of every active entry.
func SumCatalogRevisions(xs []CatalogRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCatalogRevisions keys the entries by ID, last write winning.
func IndexCatalogRevisions(xs []CatalogRevision) map[string]CatalogRevision {
	out := make(map[string]CatalogRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogRevision builds a valid CatalogRevision, for tests and for the seed data the cron
// binary writes.
func SampleCatalogRevision(id string, at time.Time) CatalogRevision {
	return CatalogRevision{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Name:     "seeded by catalog.SampleCatalogRevision",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogSnapshotSchema is the stored shape of a CatalogSnapshot.
var CatalogSnapshotSchema = schema.New("catalog.catalogsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogSnapshot is one catalog snapshot.
type CatalogSnapshot struct {
	ID       string
	Category string
	Owner    string
	Source   string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogSnapshot) Key() string { return "catalog/catalogsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x CatalogSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"source":   x.Source,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogSnapshot decodes a stored record.
func ParseCatalogSnapshot(f codec.Fields) (CatalogSnapshot, error) {
	var x CatalogSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogsnapshot: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogsnapshot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogSnapshot) String() string {
	return fmt.Sprintf("CatalogSnapshot(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogSnapshot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogSnapshot) Normalize() CatalogSnapshot {
	x.Category = strings.TrimSpace(x.Category)
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

// SortCatalogSnapshots orders by count descending, then by ID.
func SortCatalogSnapshots(xs []CatalogSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogSnapshots drops the archived entries, preserving order.
func ActiveCatalogSnapshots(xs []CatalogSnapshot) []CatalogSnapshot {
	out := make([]CatalogSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogSnapshots adds up the count of every active entry.
func SumCatalogSnapshots(xs []CatalogSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCatalogSnapshots keys the entries by ID, last write winning.
func IndexCatalogSnapshots(xs []CatalogSnapshot) map[string]CatalogSnapshot {
	out := make(map[string]CatalogSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogSnapshot builds a valid CatalogSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleCatalogSnapshot(id string, at time.Time) CatalogSnapshot {
	return CatalogSnapshot{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by catalog.SampleCatalogSnapshot",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogBatchSchema is the stored shape of a CatalogBatch.
var CatalogBatchSchema = schema.New("catalog.catalogbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogBatch is one catalog batch.
type CatalogBatch struct {
	ID       string
	Source   string
	Name     string
	Label    string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogBatch) Key() string { return "catalog/catalogbatch/" + x.ID }

// Fields encodes x for storage.
func (x CatalogBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogBatch decodes a stored record.
func ParseCatalogBatch(f codec.Fields) (CatalogBatch, error) {
	var x CatalogBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogbatch: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogbatch %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogBatch) String() string {
	return fmt.Sprintf("CatalogBatch(%s, source=%q, offset=%d, archived=%t)", x.ID, x.Source, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogBatch) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogBatch) Normalize() CatalogBatch {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortCatalogBatchs orders by offset descending, then by ID.
func SortCatalogBatchs(xs []CatalogBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogBatchs drops the archived entries, preserving order.
func ActiveCatalogBatchs(xs []CatalogBatch) []CatalogBatch {
	out := make([]CatalogBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogBatchs adds up the offset of every active entry.
func SumCatalogBatchs(xs []CatalogBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexCatalogBatchs keys the entries by ID, last write winning.
func IndexCatalogBatchs(xs []CatalogBatch) map[string]CatalogBatch {
	out := make(map[string]CatalogBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogBatch builds a valid CatalogBatch, for tests and for the seed data the cron
// binary writes.
func SampleCatalogBatch(id string, at time.Time) CatalogBatch {
	return CatalogBatch{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Label:    "seeded by catalog.SampleCatalogBatch",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogRequestSchema is the stored shape of a CatalogRequest.
var CatalogRequestSchema = schema.New("catalog.catalogrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogRequest is one catalog request.
type CatalogRequest struct {
	ID       string
	Name     string
	Label    string
	Source   string
	Version  int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogRequest) Key() string { return "catalog/catalogrequest/" + x.ID }

// Fields encodes x for storage.
func (x CatalogRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogRequest decodes a stored record.
func ParseCatalogRequest(f codec.Fields) (CatalogRequest, error) {
	var x CatalogRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogrequest: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrequest %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogRequest) String() string {
	return fmt.Sprintf("CatalogRequest(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogRequest) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogRequest) Normalize() CatalogRequest {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortCatalogRequests orders by version descending, then by ID.
func SortCatalogRequests(xs []CatalogRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogRequests drops the archived entries, preserving order.
func ActiveCatalogRequests(xs []CatalogRequest) []CatalogRequest {
	out := make([]CatalogRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogRequests adds up the version of every active entry.
func SumCatalogRequests(xs []CatalogRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexCatalogRequests keys the entries by ID, last write winning.
func IndexCatalogRequests(xs []CatalogRequest) map[string]CatalogRequest {
	out := make(map[string]CatalogRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogRequest builds a valid CatalogRequest, for tests and for the seed data the cron
// binary writes.
func SampleCatalogRequest(id string, at time.Time) CatalogRequest {
	return CatalogRequest{
		ID:      id,
		Name:    "name-" + id,
		Label:   "label-" + id,
		Source:  "seeded by catalog.SampleCatalogRequest",
		Version: len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// CatalogSummarySchema is the stored shape of a CatalogSummary.
var CatalogSummarySchema = schema.New("catalog.catalogsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogSummary is one catalog summary.
type CatalogSummary struct {
	ID       string
	Source   string
	Category string
	Name     string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogSummary) Key() string { return "catalog/catalogsummary/" + x.ID }

// Fields encodes x for storage.
func (x CatalogSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"category": x.Category,
		"name":     x.Name,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogSummary decodes a stored record.
func ParseCatalogSummary(f codec.Fields) (CatalogSummary, error) {
	var x CatalogSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogsummary: missing id")
	}
	x.Source = f["source"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogsummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogSummary) String() string {
	return fmt.Sprintf("CatalogSummary(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogSummary) Normalize() CatalogSummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortCatalogSummarys orders by capacity descending, then by ID.
func SortCatalogSummarys(xs []CatalogSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogSummarys drops the archived entries, preserving order.
func ActiveCatalogSummarys(xs []CatalogSummary) []CatalogSummary {
	out := make([]CatalogSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogSummarys adds up the capacity of every active entry.
func SumCatalogSummarys(xs []CatalogSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCatalogSummarys keys the entries by ID, last write winning.
func IndexCatalogSummarys(xs []CatalogSummary) map[string]CatalogSummary {
	out := make(map[string]CatalogSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogSummary builds a valid CatalogSummary, for tests and for the seed data the cron
// binary writes.
func SampleCatalogSummary(id string, at time.Time) CatalogSummary {
	return CatalogSummary{
		ID:       id,
		Source:   "source-" + id,
		Category: "category-" + id,
		Name:     "seeded by catalog.SampleCatalogSummary",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogPolicySchema is the stored shape of a CatalogPolicy.
var CatalogPolicySchema = schema.New("catalog.catalogpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogPolicy is one catalog policy.
type CatalogPolicy struct {
	ID       string
	Comment  string
	Owner    string
	Category string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogPolicy) Key() string { return "catalog/catalogpolicy/" + x.ID }

// Fields encodes x for storage.
func (x CatalogPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogPolicy decodes a stored record.
func ParseCatalogPolicy(f codec.Fields) (CatalogPolicy, error) {
	var x CatalogPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogpolicy: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogpolicy %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogPolicy) String() string {
	return fmt.Sprintf("CatalogPolicy(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogPolicy) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogPolicy) Normalize() CatalogPolicy {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortCatalogPolicys orders by attempts descending, then by ID.
func SortCatalogPolicys(xs []CatalogPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogPolicys drops the archived entries, preserving order.
func ActiveCatalogPolicys(xs []CatalogPolicy) []CatalogPolicy {
	out := make([]CatalogPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogPolicys adds up the attempts of every active entry.
func SumCatalogPolicys(xs []CatalogPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexCatalogPolicys keys the entries by ID, last write winning.
func IndexCatalogPolicys(xs []CatalogPolicy) map[string]CatalogPolicy {
	out := make(map[string]CatalogPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogPolicy builds a valid CatalogPolicy, for tests and for the seed data the cron
// binary writes.
func SampleCatalogPolicy(id string, at time.Time) CatalogPolicy {
	return CatalogPolicy{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by catalog.SampleCatalogPolicy",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogLinkSchema is the stored shape of a CatalogLink.
var CatalogLinkSchema = schema.New("catalog.cataloglink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogLink is one catalog link.
type CatalogLink struct {
	ID       string
	Name     string
	Label    string
	Channel  string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogLink) Key() string { return "catalog/cataloglink/" + x.ID }

// Fields encodes x for storage.
func (x CatalogLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogLink decodes a stored record.
func ParseCatalogLink(f codec.Fields) (CatalogLink, error) {
	var x CatalogLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: cataloglink: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("catalog: cataloglink %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("catalog: cataloglink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: cataloglink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: cataloglink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogLink) String() string {
	return fmt.Sprintf("CatalogLink(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogLink) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogLink) Normalize() CatalogLink {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCatalogLinks orders by version descending, then by ID.
func SortCatalogLinks(xs []CatalogLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogLinks drops the archived entries, preserving order.
func ActiveCatalogLinks(xs []CatalogLink) []CatalogLink {
	out := make([]CatalogLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogLinks adds up the version of every active entry.
func SumCatalogLinks(xs []CatalogLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexCatalogLinks keys the entries by ID, last write winning.
func IndexCatalogLinks(xs []CatalogLink) map[string]CatalogLink {
	out := make(map[string]CatalogLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogLink builds a valid CatalogLink, for tests and for the seed data the cron
// binary writes.
func SampleCatalogLink(id string, at time.Time) CatalogLink {
	return CatalogLink{
		ID:       id,
		Name:     "name-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by catalog.SampleCatalogLink",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogNoteSchema is the stored shape of a CatalogNote.
var CatalogNoteSchema = schema.New("catalog.catalognote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogNote is one catalog note.
type CatalogNote struct {
	ID       string
	Region   string
	Channel  string
	Source   string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogNote) Key() string { return "catalog/catalognote/" + x.ID }

// Fields encodes x for storage.
func (x CatalogNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogNote decodes a stored record.
func ParseCatalogNote(f codec.Fields) (CatalogNote, error) {
	var x CatalogNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalognote: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("catalog: catalognote %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("catalog: catalognote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalognote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalognote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogNote) String() string {
	return fmt.Sprintf("CatalogNote(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogNote) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogNote) Normalize() CatalogNote {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortCatalogNotes orders by weight descending, then by ID.
func SortCatalogNotes(xs []CatalogNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogNotes drops the archived entries, preserving order.
func ActiveCatalogNotes(xs []CatalogNote) []CatalogNote {
	out := make([]CatalogNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogNotes adds up the weight of every active entry.
func SumCatalogNotes(xs []CatalogNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexCatalogNotes keys the entries by ID, last write winning.
func IndexCatalogNotes(xs []CatalogNote) map[string]CatalogNote {
	out := make(map[string]CatalogNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogNote builds a valid CatalogNote, for tests and for the seed data the cron
// binary writes.
func SampleCatalogNote(id string, at time.Time) CatalogNote {
	return CatalogNote{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by catalog.SampleCatalogNote",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogWindowSchema is the stored shape of a CatalogWindow.
var CatalogWindowSchema = schema.New("catalog.catalogwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogWindow is one catalog window.
type CatalogWindow struct {
	ID        string
	Region    string
	Comment   string
	Reference string
	Offset    int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CatalogWindow) Key() string { return "catalog/catalogwindow/" + x.ID }

// Fields encodes x for storage.
func (x CatalogWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"comment":   x.Comment,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogWindow decodes a stored record.
func ParseCatalogWindow(f codec.Fields) (CatalogWindow, error) {
	var x CatalogWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogwindow: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogwindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogWindow) String() string {
	return fmt.Sprintf("CatalogWindow(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogWindow) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogWindow) Normalize() CatalogWindow {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortCatalogWindows orders by offset descending, then by ID.
func SortCatalogWindows(xs []CatalogWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogWindows drops the archived entries, preserving order.
func ActiveCatalogWindows(xs []CatalogWindow) []CatalogWindow {
	out := make([]CatalogWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogWindows adds up the offset of every active entry.
func SumCatalogWindows(xs []CatalogWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexCatalogWindows keys the entries by ID, last write winning.
func IndexCatalogWindows(xs []CatalogWindow) map[string]CatalogWindow {
	out := make(map[string]CatalogWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogWindow builds a valid CatalogWindow, for tests and for the seed data the cron
// binary writes.
func SampleCatalogWindow(id string, at time.Time) CatalogWindow {
	return CatalogWindow{
		ID:        id,
		Region:    "region-" + id,
		Comment:   "comment-" + id,
		Reference: "seeded by catalog.SampleCatalogWindow",
		Offset:    len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CatalogQuotaSchema is the stored shape of a CatalogQuota.
var CatalogQuotaSchema = schema.New("catalog.catalogquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogQuota is one catalog quota.
type CatalogQuota struct {
	ID        string
	Source    string
	Name      string
	Reference string
	Attempts  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CatalogQuota) Key() string { return "catalog/catalogquota/" + x.ID }

// Fields encodes x for storage.
func (x CatalogQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"name":      x.Name,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogQuota decodes a stored record.
func ParseCatalogQuota(f codec.Fields) (CatalogQuota, error) {
	var x CatalogQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogquota: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogquota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogQuota) String() string {
	return fmt.Sprintf("CatalogQuota(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogQuota) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogQuota) Normalize() CatalogQuota {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortCatalogQuotas orders by attempts descending, then by ID.
func SortCatalogQuotas(xs []CatalogQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogQuotas drops the archived entries, preserving order.
func ActiveCatalogQuotas(xs []CatalogQuota) []CatalogQuota {
	out := make([]CatalogQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogQuotas adds up the attempts of every active entry.
func SumCatalogQuotas(xs []CatalogQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexCatalogQuotas keys the entries by ID, last write winning.
func IndexCatalogQuotas(xs []CatalogQuota) map[string]CatalogQuota {
	out := make(map[string]CatalogQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogQuota builds a valid CatalogQuota, for tests and for the seed data the cron
// binary writes.
func SampleCatalogQuota(id string, at time.Time) CatalogQuota {
	return CatalogQuota{
		ID:        id,
		Source:    "source-" + id,
		Name:      "name-" + id,
		Reference: "seeded by catalog.SampleCatalogQuota",
		Attempts:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CatalogBucketSchema is the stored shape of a CatalogBucket.
var CatalogBucketSchema = schema.New("catalog.catalogbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogBucket is one catalog bucket.
type CatalogBucket struct {
	ID        string
	Name      string
	Reference string
	Channel   string
	Retained  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CatalogBucket) Key() string { return "catalog/catalogbucket/" + x.ID }

// Fields encodes x for storage.
func (x CatalogBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"channel":   x.Channel,
		"retained":  strconv.Itoa(x.Retained),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogBucket decodes a stored record.
func ParseCatalogBucket(f codec.Fields) (CatalogBucket, error) {
	var x CatalogBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogbucket: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogbucket %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogBucket) String() string {
	return fmt.Sprintf("CatalogBucket(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogBucket) Normalize() CatalogBucket {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortCatalogBuckets orders by retained descending, then by ID.
func SortCatalogBuckets(xs []CatalogBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogBuckets drops the archived entries, preserving order.
func ActiveCatalogBuckets(xs []CatalogBucket) []CatalogBucket {
	out := make([]CatalogBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogBuckets adds up the retained of every active entry.
func SumCatalogBuckets(xs []CatalogBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexCatalogBuckets keys the entries by ID, last write winning.
func IndexCatalogBuckets(xs []CatalogBucket) map[string]CatalogBucket {
	out := make(map[string]CatalogBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogBucket builds a valid CatalogBucket, for tests and for the seed data the cron
// binary writes.
func SampleCatalogBucket(id string, at time.Time) CatalogBucket {
	return CatalogBucket{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by catalog.SampleCatalogBucket",
		Retained:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CatalogLaneSchema is the stored shape of a CatalogLane.
var CatalogLaneSchema = schema.New("catalog.cataloglane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogLane is one catalog lane.
type CatalogLane struct {
	ID        string
	Reference string
	Channel   string
	Category  string
	Capacity  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CatalogLane) Key() string { return "catalog/cataloglane/" + x.ID }

// Fields encodes x for storage.
func (x CatalogLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"category":  x.Category,
		"capacity":  strconv.Itoa(x.Capacity),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogLane decodes a stored record.
func ParseCatalogLane(f codec.Fields) (CatalogLane, error) {
	var x CatalogLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: cataloglane: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("catalog: cataloglane %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("catalog: cataloglane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: cataloglane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: cataloglane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogLane) String() string {
	return fmt.Sprintf("CatalogLane(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogLane) Normalize() CatalogLane {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortCatalogLanes orders by capacity descending, then by ID.
func SortCatalogLanes(xs []CatalogLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogLanes drops the archived entries, preserving order.
func ActiveCatalogLanes(xs []CatalogLane) []CatalogLane {
	out := make([]CatalogLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogLanes adds up the capacity of every active entry.
func SumCatalogLanes(xs []CatalogLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCatalogLanes keys the entries by ID, last write winning.
func IndexCatalogLanes(xs []CatalogLane) map[string]CatalogLane {
	out := make(map[string]CatalogLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogLane builds a valid CatalogLane, for tests and for the seed data the cron
// binary writes.
func SampleCatalogLane(id string, at time.Time) CatalogLane {
	return CatalogLane{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Category:  "seeded by catalog.SampleCatalogLane",
		Capacity:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CatalogSlotSchema is the stored shape of a CatalogSlot.
var CatalogSlotSchema = schema.New("catalog.catalogslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogSlot is one catalog slot.
type CatalogSlot struct {
	ID        string
	Label     string
	Source    string
	Reference string
	Retained  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CatalogSlot) Key() string { return "catalog/catalogslot/" + x.ID }

// Fields encodes x for storage.
func (x CatalogSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"source":    x.Source,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogSlot decodes a stored record.
func ParseCatalogSlot(f codec.Fields) (CatalogSlot, error) {
	var x CatalogSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogslot: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogslot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogSlot) String() string {
	return fmt.Sprintf("CatalogSlot(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogSlot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogSlot) Normalize() CatalogSlot {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCatalogSlots orders by retained descending, then by ID.
func SortCatalogSlots(xs []CatalogSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogSlots drops the archived entries, preserving order.
func ActiveCatalogSlots(xs []CatalogSlot) []CatalogSlot {
	out := make([]CatalogSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogSlots adds up the retained of every active entry.
func SumCatalogSlots(xs []CatalogSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexCatalogSlots keys the entries by ID, last write winning.
func IndexCatalogSlots(xs []CatalogSlot) map[string]CatalogSlot {
	out := make(map[string]CatalogSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogSlot builds a valid CatalogSlot, for tests and for the seed data the cron
// binary writes.
func SampleCatalogSlot(id string, at time.Time) CatalogSlot {
	return CatalogSlot{
		ID:        id,
		Label:     "label-" + id,
		Source:    "source-" + id,
		Reference: "seeded by catalog.SampleCatalogSlot",
		Retained:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CatalogTagSchema is the stored shape of a CatalogTag.
var CatalogTagSchema = schema.New("catalog.catalogtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogTag is one catalog tag.
type CatalogTag struct {
	ID        string
	Group     string
	Reference string
	Source    string
	Count     int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CatalogTag) Key() string { return "catalog/catalogtag/" + x.ID }

// Fields encodes x for storage.
func (x CatalogTag) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"source":    x.Source,
		"count":     strconv.Itoa(x.Count),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogTag decodes a stored record.
func ParseCatalogTag(f codec.Fields) (CatalogTag, error) {
	var x CatalogTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogtag: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogtag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogTag) String() string {
	return fmt.Sprintf("CatalogTag(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogTag) Normalize() CatalogTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortCatalogTags orders by count descending, then by ID.
func SortCatalogTags(xs []CatalogTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogTags drops the archived entries, preserving order.
func ActiveCatalogTags(xs []CatalogTag) []CatalogTag {
	out := make([]CatalogTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogTags adds up the count of every active entry.
func SumCatalogTags(xs []CatalogTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCatalogTags keys the entries by ID, last write winning.
func IndexCatalogTags(xs []CatalogTag) map[string]CatalogTag {
	out := make(map[string]CatalogTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogTag builds a valid CatalogTag, for tests and for the seed data the cron
// binary writes.
func SampleCatalogTag(id string, at time.Time) CatalogTag {
	return CatalogTag{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by catalog.SampleCatalogTag",
		Count:     len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CatalogRuleSchema is the stored shape of a CatalogRule.
var CatalogRuleSchema = schema.New("catalog.catalogrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogRule is one catalog rule.
type CatalogRule struct {
	ID       string
	Region   string
	Category string
	Owner    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogRule) Key() string { return "catalog/catalogrule/" + x.ID }

// Fields encodes x for storage.
func (x CatalogRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"owner":    x.Owner,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogRule decodes a stored record.
func ParseCatalogRule(f codec.Fields) (CatalogRule, error) {
	var x CatalogRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogrule: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrule %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogRule) String() string {
	return fmt.Sprintf("CatalogRule(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogRule) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogRule) Normalize() CatalogRule {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortCatalogRules orders by count descending, then by ID.
func SortCatalogRules(xs []CatalogRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogRules drops the archived entries, preserving order.
func ActiveCatalogRules(xs []CatalogRule) []CatalogRule {
	out := make([]CatalogRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogRules adds up the count of every active entry.
func SumCatalogRules(xs []CatalogRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCatalogRules keys the entries by ID, last write winning.
func IndexCatalogRules(xs []CatalogRule) map[string]CatalogRule {
	out := make(map[string]CatalogRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogRule builds a valid CatalogRule, for tests and for the seed data the cron
// binary writes.
func SampleCatalogRule(id string, at time.Time) CatalogRule {
	return CatalogRule{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Owner:    "seeded by catalog.SampleCatalogRule",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogHoldSchema is the stored shape of a CatalogHold.
var CatalogHoldSchema = schema.New("catalog.cataloghold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogHold is one catalog hold.
type CatalogHold struct {
	ID       string
	Source   string
	Channel  string
	Category string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogHold) Key() string { return "catalog/cataloghold/" + x.ID }

// Fields encodes x for storage.
func (x CatalogHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"category": x.Category,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogHold decodes a stored record.
func ParseCatalogHold(f codec.Fields) (CatalogHold, error) {
	var x CatalogHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: cataloghold: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("catalog: cataloghold %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("catalog: cataloghold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: cataloghold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: cataloghold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogHold) String() string {
	return fmt.Sprintf("CatalogHold(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogHold) Normalize() CatalogHold {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortCatalogHolds orders by weight descending, then by ID.
func SortCatalogHolds(xs []CatalogHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogHolds drops the archived entries, preserving order.
func ActiveCatalogHolds(xs []CatalogHold) []CatalogHold {
	out := make([]CatalogHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogHolds adds up the weight of every active entry.
func SumCatalogHolds(xs []CatalogHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexCatalogHolds keys the entries by ID, last write winning.
func IndexCatalogHolds(xs []CatalogHold) map[string]CatalogHold {
	out := make(map[string]CatalogHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogHold builds a valid CatalogHold, for tests and for the seed data the cron
// binary writes.
func SampleCatalogHold(id string, at time.Time) CatalogHold {
	return CatalogHold{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by catalog.SampleCatalogHold",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogClaimSchema is the stored shape of a CatalogClaim.
var CatalogClaimSchema = schema.New("catalog.catalogclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogClaim is one catalog claim.
type CatalogClaim struct {
	ID       string
	Name     string
	Owner    string
	Group    string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogClaim) Key() string { return "catalog/catalogclaim/" + x.ID }

// Fields encodes x for storage.
func (x CatalogClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogClaim decodes a stored record.
func ParseCatalogClaim(f codec.Fields) (CatalogClaim, error) {
	var x CatalogClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogclaim: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogClaim) String() string {
	return fmt.Sprintf("CatalogClaim(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogClaim) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogClaim) Normalize() CatalogClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortCatalogClaims orders by capacity descending, then by ID.
func SortCatalogClaims(xs []CatalogClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogClaims drops the archived entries, preserving order.
func ActiveCatalogClaims(xs []CatalogClaim) []CatalogClaim {
	out := make([]CatalogClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogClaims adds up the capacity of every active entry.
func SumCatalogClaims(xs []CatalogClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCatalogClaims keys the entries by ID, last write winning.
func IndexCatalogClaims(xs []CatalogClaim) map[string]CatalogClaim {
	out := make(map[string]CatalogClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogClaim builds a valid CatalogClaim, for tests and for the seed data the cron
// binary writes.
func SampleCatalogClaim(id string, at time.Time) CatalogClaim {
	return CatalogClaim{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by catalog.SampleCatalogClaim",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CatalogCursorSchema is the stored shape of a CatalogCursor.
var CatalogCursorSchema = schema.New("catalog.catalogcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CatalogCursor is one catalog cursor.
type CatalogCursor struct {
	ID       string
	Label    string
	Owner    string
	Region   string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CatalogCursor) Key() string { return "catalog/catalogcursor/" + x.ID }

// Fields encodes x for storage.
func (x CatalogCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"owner":    x.Owner,
		"region":   x.Region,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCatalogCursor decodes a stored record.
func ParseCatalogCursor(f codec.Fields) (CatalogCursor, error) {
	var x CatalogCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("catalog: catalogcursor: missing id")
	}
	x.Label = f["label"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogcursor %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("catalog: catalogcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("catalog: catalogcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("catalog: catalogcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CatalogCursor) String() string {
	return fmt.Sprintf("CatalogCursor(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CatalogCursor) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CatalogCursor) Normalize() CatalogCursor {
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCatalogCursors orders by count descending, then by ID.
func SortCatalogCursors(xs []CatalogCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCatalogCursors drops the archived entries, preserving order.
func ActiveCatalogCursors(xs []CatalogCursor) []CatalogCursor {
	out := make([]CatalogCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCatalogCursors adds up the count of every active entry.
func SumCatalogCursors(xs []CatalogCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCatalogCursors keys the entries by ID, last write winning.
func IndexCatalogCursors(xs []CatalogCursor) map[string]CatalogCursor {
	out := make(map[string]CatalogCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCatalogCursor builds a valid CatalogCursor, for tests and for the seed data the cron
// binary writes.
func SampleCatalogCursor(id string, at time.Time) CatalogCursor {
	return CatalogCursor{
		ID:       id,
		Label:    "label-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by catalog.SampleCatalogCursor",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	CatalogRecordSchema,
	CatalogEntrySchema,
	CatalogItemSchema,
	CatalogDraftSchema,
	CatalogRevisionSchema,
	CatalogSnapshotSchema,
	CatalogBatchSchema,
	CatalogRequestSchema,
	CatalogSummarySchema,
	CatalogPolicySchema,
	CatalogLinkSchema,
	CatalogNoteSchema,
	CatalogWindowSchema,
	CatalogQuotaSchema,
	CatalogBucketSchema,
	CatalogLaneSchema,
	CatalogSlotSchema,
	CatalogTagSchema,
	CatalogRuleSchema,
	CatalogHoldSchema,
	CatalogClaimSchema,
	CatalogCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"catalog/catalogrecord/",
	"catalog/catalogentry/",
	"catalog/catalogitem/",
	"catalog/catalogdraft/",
	"catalog/catalogrevision/",
	"catalog/catalogsnapshot/",
	"catalog/catalogbatch/",
	"catalog/catalogrequest/",
	"catalog/catalogsummary/",
	"catalog/catalogpolicy/",
	"catalog/cataloglink/",
	"catalog/catalognote/",
	"catalog/catalogwindow/",
	"catalog/catalogquota/",
	"catalog/catalogbucket/",
	"catalog/cataloglane/",
	"catalog/catalogslot/",
	"catalog/catalogtag/",
	"catalog/catalogrule/",
	"catalog/cataloghold/",
	"catalog/catalogclaim/",
	"catalog/catalogcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "catalog: " + strings.Join(names, ", ")
}
