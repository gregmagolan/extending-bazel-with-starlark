// Package usage holds the usage domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package usage

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// UsageRecordSchema is the stored shape of a UsageRecord.
var UsageRecordSchema = schema.New("usage.usagerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageRecord is one usage record.
type UsageRecord struct {
	ID       string
	Region   string
	Label    string
	Category string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageRecord) Key() string { return "usage/usagerecord/" + x.ID }

// Fields encodes x for storage.
func (x UsageRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageRecord decodes a stored record.
func ParseUsageRecord(f codec.Fields) (UsageRecord, error) {
	var x UsageRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagerecord: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("usage: usagerecord %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("usage: usagerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageRecord) String() string {
	return fmt.Sprintf("UsageRecord(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageRecord) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageRecord) Normalize() UsageRecord {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortUsageRecords orders by count descending, then by ID.
func SortUsageRecords(xs []UsageRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageRecords drops the archived entries, preserving order.
func ActiveUsageRecords(xs []UsageRecord) []UsageRecord {
	out := make([]UsageRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageRecords adds up the count of every active entry.
func SumUsageRecords(xs []UsageRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexUsageRecords keys the entries by ID, last write winning.
func IndexUsageRecords(xs []UsageRecord) map[string]UsageRecord {
	out := make(map[string]UsageRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageRecord builds a valid UsageRecord, for tests and for the seed data the cron
// binary writes.
func SampleUsageRecord(id string, at time.Time) UsageRecord {
	return UsageRecord{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Category: "seeded by usage.SampleUsageRecord",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageEntrySchema is the stored shape of a UsageEntry.
var UsageEntrySchema = schema.New("usage.usageentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageEntry is one usage entry.
type UsageEntry struct {
	ID        string
	Group     string
	Label     string
	Reference string
	Capacity  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UsageEntry) Key() string { return "usage/usageentry/" + x.ID }

// Fields encodes x for storage.
func (x UsageEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"label":     x.Label,
		"reference": x.Reference,
		"capacity":  strconv.Itoa(x.Capacity),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageEntry decodes a stored record.
func ParseUsageEntry(f codec.Fields) (UsageEntry, error) {
	var x UsageEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usageentry: missing id")
	}
	x.Group = f["group"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("usage: usageentry %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("usage: usageentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usageentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usageentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageEntry) String() string {
	return fmt.Sprintf("UsageEntry(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageEntry) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageEntry) Normalize() UsageEntry {
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortUsageEntrys orders by capacity descending, then by ID.
func SortUsageEntrys(xs []UsageEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageEntrys drops the archived entries, preserving order.
func ActiveUsageEntrys(xs []UsageEntry) []UsageEntry {
	out := make([]UsageEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageEntrys adds up the capacity of every active entry.
func SumUsageEntrys(xs []UsageEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexUsageEntrys keys the entries by ID, last write winning.
func IndexUsageEntrys(xs []UsageEntry) map[string]UsageEntry {
	out := make(map[string]UsageEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageEntry builds a valid UsageEntry, for tests and for the seed data the cron
// binary writes.
func SampleUsageEntry(id string, at time.Time) UsageEntry {
	return UsageEntry{
		ID:        id,
		Group:     "group-" + id,
		Label:     "label-" + id,
		Reference: "seeded by usage.SampleUsageEntry",
		Capacity:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UsageItemSchema is the stored shape of a UsageItem.
var UsageItemSchema = schema.New("usage.usageitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageItem is one usage item.
type UsageItem struct {
	ID       string
	Source   string
	Name     string
	Owner    string
	Version  int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageItem) Key() string { return "usage/usageitem/" + x.ID }

// Fields encodes x for storage.
func (x UsageItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"owner":    x.Owner,
		"version":  strconv.Itoa(x.Version),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageItem decodes a stored record.
func ParseUsageItem(f codec.Fields) (UsageItem, error) {
	var x UsageItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usageitem: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Owner = f["owner"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("usage: usageitem %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("usage: usageitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usageitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usageitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageItem) String() string {
	return fmt.Sprintf("UsageItem(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageItem) Normalize() UsageItem {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortUsageItems orders by version descending, then by ID.
func SortUsageItems(xs []UsageItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageItems drops the archived entries, preserving order.
func ActiveUsageItems(xs []UsageItem) []UsageItem {
	out := make([]UsageItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageItems adds up the version of every active entry.
func SumUsageItems(xs []UsageItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexUsageItems keys the entries by ID, last write winning.
func IndexUsageItems(xs []UsageItem) map[string]UsageItem {
	out := make(map[string]UsageItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageItem builds a valid UsageItem, for tests and for the seed data the cron
// binary writes.
func SampleUsageItem(id string, at time.Time) UsageItem {
	return UsageItem{
		ID:      id,
		Source:  "source-" + id,
		Name:    "name-" + id,
		Owner:   "seeded by usage.SampleUsageItem",
		Version: len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// UsageDraftSchema is the stored shape of a UsageDraft.
var UsageDraftSchema = schema.New("usage.usagedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageDraft is one usage draft.
type UsageDraft struct {
	ID       string
	Group    string
	Channel  string
	Name     string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageDraft) Key() string { return "usage/usagedraft/" + x.ID }

// Fields encodes x for storage.
func (x UsageDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageDraft decodes a stored record.
func ParseUsageDraft(f codec.Fields) (UsageDraft, error) {
	var x UsageDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagedraft: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("usage: usagedraft %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("usage: usagedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageDraft) String() string {
	return fmt.Sprintf("UsageDraft(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageDraft) Normalize() UsageDraft {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortUsageDrafts orders by attempts descending, then by ID.
func SortUsageDrafts(xs []UsageDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageDrafts drops the archived entries, preserving order.
func ActiveUsageDrafts(xs []UsageDraft) []UsageDraft {
	out := make([]UsageDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageDrafts adds up the attempts of every active entry.
func SumUsageDrafts(xs []UsageDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexUsageDrafts keys the entries by ID, last write winning.
func IndexUsageDrafts(xs []UsageDraft) map[string]UsageDraft {
	out := make(map[string]UsageDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageDraft builds a valid UsageDraft, for tests and for the seed data the cron
// binary writes.
func SampleUsageDraft(id string, at time.Time) UsageDraft {
	return UsageDraft{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by usage.SampleUsageDraft",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageRevisionSchema is the stored shape of a UsageRevision.
var UsageRevisionSchema = schema.New("usage.usagerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageRevision is one usage revision.
type UsageRevision struct {
	ID       string
	Name     string
	Channel  string
	Comment  string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageRevision) Key() string { return "usage/usagerevision/" + x.ID }

// Fields encodes x for storage.
func (x UsageRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageRevision decodes a stored record.
func ParseUsageRevision(f codec.Fields) (UsageRevision, error) {
	var x UsageRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagerevision: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("usage: usagerevision %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("usage: usagerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageRevision) String() string {
	return fmt.Sprintf("UsageRevision(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageRevision) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageRevision) Normalize() UsageRevision {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortUsageRevisions orders by offset descending, then by ID.
func SortUsageRevisions(xs []UsageRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageRevisions drops the archived entries, preserving order.
func ActiveUsageRevisions(xs []UsageRevision) []UsageRevision {
	out := make([]UsageRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageRevisions adds up the offset of every active entry.
func SumUsageRevisions(xs []UsageRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUsageRevisions keys the entries by ID, last write winning.
func IndexUsageRevisions(xs []UsageRevision) map[string]UsageRevision {
	out := make(map[string]UsageRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageRevision builds a valid UsageRevision, for tests and for the seed data the cron
// binary writes.
func SampleUsageRevision(id string, at time.Time) UsageRevision {
	return UsageRevision{
		ID:       id,
		Name:     "name-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by usage.SampleUsageRevision",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageSnapshotSchema is the stored shape of a UsageSnapshot.
var UsageSnapshotSchema = schema.New("usage.usagesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageSnapshot is one usage snapshot.
type UsageSnapshot struct {
	ID       string
	Category string
	Name     string
	Group    string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageSnapshot) Key() string { return "usage/usagesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x UsageSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageSnapshot decodes a stored record.
func ParseUsageSnapshot(f codec.Fields) (UsageSnapshot, error) {
	var x UsageSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagesnapshot: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("usage: usagesnapshot %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("usage: usagesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageSnapshot) String() string {
	return fmt.Sprintf("UsageSnapshot(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageSnapshot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageSnapshot) Normalize() UsageSnapshot {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortUsageSnapshots orders by retained descending, then by ID.
func SortUsageSnapshots(xs []UsageSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageSnapshots drops the archived entries, preserving order.
func ActiveUsageSnapshots(xs []UsageSnapshot) []UsageSnapshot {
	out := make([]UsageSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageSnapshots adds up the retained of every active entry.
func SumUsageSnapshots(xs []UsageSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexUsageSnapshots keys the entries by ID, last write winning.
func IndexUsageSnapshots(xs []UsageSnapshot) map[string]UsageSnapshot {
	out := make(map[string]UsageSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageSnapshot builds a valid UsageSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleUsageSnapshot(id string, at time.Time) UsageSnapshot {
	return UsageSnapshot{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Group:    "seeded by usage.SampleUsageSnapshot",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageBatchSchema is the stored shape of a UsageBatch.
var UsageBatchSchema = schema.New("usage.usagebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageBatch is one usage batch.
type UsageBatch struct {
	ID       string
	Label    string
	Category string
	Group    string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageBatch) Key() string { return "usage/usagebatch/" + x.ID }

// Fields encodes x for storage.
func (x UsageBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageBatch decodes a stored record.
func ParseUsageBatch(f codec.Fields) (UsageBatch, error) {
	var x UsageBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagebatch: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("usage: usagebatch %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("usage: usagebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageBatch) String() string {
	return fmt.Sprintf("UsageBatch(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageBatch) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageBatch) Normalize() UsageBatch {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortUsageBatchs orders by retained descending, then by ID.
func SortUsageBatchs(xs []UsageBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageBatchs drops the archived entries, preserving order.
func ActiveUsageBatchs(xs []UsageBatch) []UsageBatch {
	out := make([]UsageBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageBatchs adds up the retained of every active entry.
func SumUsageBatchs(xs []UsageBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexUsageBatchs keys the entries by ID, last write winning.
func IndexUsageBatchs(xs []UsageBatch) map[string]UsageBatch {
	out := make(map[string]UsageBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageBatch builds a valid UsageBatch, for tests and for the seed data the cron
// binary writes.
func SampleUsageBatch(id string, at time.Time) UsageBatch {
	return UsageBatch{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Group:    "seeded by usage.SampleUsageBatch",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageRequestSchema is the stored shape of a UsageRequest.
var UsageRequestSchema = schema.New("usage.usagerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageRequest is one usage request.
type UsageRequest struct {
	ID       string
	Label    string
	Source   string
	Owner    string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageRequest) Key() string { return "usage/usagerequest/" + x.ID }

// Fields encodes x for storage.
func (x UsageRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageRequest decodes a stored record.
func ParseUsageRequest(f codec.Fields) (UsageRequest, error) {
	var x UsageRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagerequest: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("usage: usagerequest %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("usage: usagerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageRequest) String() string {
	return fmt.Sprintf("UsageRequest(%s, label=%q, capacity=%d, archived=%t)", x.ID, x.Label, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageRequest) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageRequest) Normalize() UsageRequest {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortUsageRequests orders by capacity descending, then by ID.
func SortUsageRequests(xs []UsageRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageRequests drops the archived entries, preserving order.
func ActiveUsageRequests(xs []UsageRequest) []UsageRequest {
	out := make([]UsageRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageRequests adds up the capacity of every active entry.
func SumUsageRequests(xs []UsageRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexUsageRequests keys the entries by ID, last write winning.
func IndexUsageRequests(xs []UsageRequest) map[string]UsageRequest {
	out := make(map[string]UsageRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageRequest builds a valid UsageRequest, for tests and for the seed data the cron
// binary writes.
func SampleUsageRequest(id string, at time.Time) UsageRequest {
	return UsageRequest{
		ID:       id,
		Label:    "label-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by usage.SampleUsageRequest",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageSummarySchema is the stored shape of a UsageSummary.
var UsageSummarySchema = schema.New("usage.usagesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageSummary is one usage summary.
type UsageSummary struct {
	ID        string
	Reference string
	Group     string
	Source    string
	Capacity  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UsageSummary) Key() string { return "usage/usagesummary/" + x.ID }

// Fields encodes x for storage.
func (x UsageSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"source":    x.Source,
		"capacity":  strconv.Itoa(x.Capacity),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageSummary decodes a stored record.
func ParseUsageSummary(f codec.Fields) (UsageSummary, error) {
	var x UsageSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagesummary: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("usage: usagesummary %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("usage: usagesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageSummary) String() string {
	return fmt.Sprintf("UsageSummary(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageSummary) Normalize() UsageSummary {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortUsageSummarys orders by capacity descending, then by ID.
func SortUsageSummarys(xs []UsageSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageSummarys drops the archived entries, preserving order.
func ActiveUsageSummarys(xs []UsageSummary) []UsageSummary {
	out := make([]UsageSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageSummarys adds up the capacity of every active entry.
func SumUsageSummarys(xs []UsageSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexUsageSummarys keys the entries by ID, last write winning.
func IndexUsageSummarys(xs []UsageSummary) map[string]UsageSummary {
	out := make(map[string]UsageSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageSummary builds a valid UsageSummary, for tests and for the seed data the cron
// binary writes.
func SampleUsageSummary(id string, at time.Time) UsageSummary {
	return UsageSummary{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Source:    "seeded by usage.SampleUsageSummary",
		Capacity:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UsagePolicySchema is the stored shape of a UsagePolicy.
var UsagePolicySchema = schema.New("usage.usagepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsagePolicy is one usage policy.
type UsagePolicy struct {
	ID       string
	Source   string
	Channel  string
	Label    string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsagePolicy) Key() string { return "usage/usagepolicy/" + x.ID }

// Fields encodes x for storage.
func (x UsagePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"label":    x.Label,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsagePolicy decodes a stored record.
func ParseUsagePolicy(f codec.Fields) (UsagePolicy, error) {
	var x UsagePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagepolicy: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("usage: usagepolicy %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("usage: usagepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsagePolicy) String() string {
	return fmt.Sprintf("UsagePolicy(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsagePolicy) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsagePolicy) Normalize() UsagePolicy {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortUsagePolicys orders by version descending, then by ID.
func SortUsagePolicys(xs []UsagePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsagePolicys drops the archived entries, preserving order.
func ActiveUsagePolicys(xs []UsagePolicy) []UsagePolicy {
	out := make([]UsagePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsagePolicys adds up the version of every active entry.
func SumUsagePolicys(xs []UsagePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexUsagePolicys keys the entries by ID, last write winning.
func IndexUsagePolicys(xs []UsagePolicy) map[string]UsagePolicy {
	out := make(map[string]UsagePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsagePolicy builds a valid UsagePolicy, for tests and for the seed data the cron
// binary writes.
func SampleUsagePolicy(id string, at time.Time) UsagePolicy {
	return UsagePolicy{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by usage.SampleUsagePolicy",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageLinkSchema is the stored shape of a UsageLink.
var UsageLinkSchema = schema.New("usage.usagelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageLink is one usage link.
type UsageLink struct {
	ID       string
	Group    string
	Source   string
	Comment  string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageLink) Key() string { return "usage/usagelink/" + x.ID }

// Fields encodes x for storage.
func (x UsageLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageLink decodes a stored record.
func ParseUsageLink(f codec.Fields) (UsageLink, error) {
	var x UsageLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagelink: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("usage: usagelink %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("usage: usagelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageLink) String() string {
	return fmt.Sprintf("UsageLink(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageLink) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageLink) Normalize() UsageLink {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortUsageLinks orders by count descending, then by ID.
func SortUsageLinks(xs []UsageLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageLinks drops the archived entries, preserving order.
func ActiveUsageLinks(xs []UsageLink) []UsageLink {
	out := make([]UsageLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageLinks adds up the count of every active entry.
func SumUsageLinks(xs []UsageLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexUsageLinks keys the entries by ID, last write winning.
func IndexUsageLinks(xs []UsageLink) map[string]UsageLink {
	out := make(map[string]UsageLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageLink builds a valid UsageLink, for tests and for the seed data the cron
// binary writes.
func SampleUsageLink(id string, at time.Time) UsageLink {
	return UsageLink{
		ID:      id,
		Group:   "group-" + id,
		Source:  "source-" + id,
		Comment: "seeded by usage.SampleUsageLink",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// UsageNoteSchema is the stored shape of a UsageNote.
var UsageNoteSchema = schema.New("usage.usagenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageNote is one usage note.
type UsageNote struct {
	ID        string
	Comment   string
	Reference string
	Name      string
	Offset    int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UsageNote) Key() string { return "usage/usagenote/" + x.ID }

// Fields encodes x for storage.
func (x UsageNote) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"name":      x.Name,
		"offset":    strconv.Itoa(x.Offset),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageNote decodes a stored record.
func ParseUsageNote(f codec.Fields) (UsageNote, error) {
	var x UsageNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagenote: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("usage: usagenote %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("usage: usagenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageNote) String() string {
	return fmt.Sprintf("UsageNote(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageNote) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageNote) Normalize() UsageNote {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortUsageNotes orders by offset descending, then by ID.
func SortUsageNotes(xs []UsageNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageNotes drops the archived entries, preserving order.
func ActiveUsageNotes(xs []UsageNote) []UsageNote {
	out := make([]UsageNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageNotes adds up the offset of every active entry.
func SumUsageNotes(xs []UsageNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexUsageNotes keys the entries by ID, last write winning.
func IndexUsageNotes(xs []UsageNote) map[string]UsageNote {
	out := make(map[string]UsageNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageNote builds a valid UsageNote, for tests and for the seed data the cron
// binary writes.
func SampleUsageNote(id string, at time.Time) UsageNote {
	return UsageNote{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by usage.SampleUsageNote",
		Offset:    len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UsageWindowSchema is the stored shape of a UsageWindow.
var UsageWindowSchema = schema.New("usage.usagewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageWindow is one usage window.
type UsageWindow struct {
	ID       string
	Group    string
	Source   string
	Channel  string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageWindow) Key() string { return "usage/usagewindow/" + x.ID }

// Fields encodes x for storage.
func (x UsageWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageWindow decodes a stored record.
func ParseUsageWindow(f codec.Fields) (UsageWindow, error) {
	var x UsageWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagewindow: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("usage: usagewindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("usage: usagewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageWindow) String() string {
	return fmt.Sprintf("UsageWindow(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageWindow) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageWindow) Normalize() UsageWindow {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortUsageWindows orders by count descending, then by ID.
func SortUsageWindows(xs []UsageWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageWindows drops the archived entries, preserving order.
func ActiveUsageWindows(xs []UsageWindow) []UsageWindow {
	out := make([]UsageWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageWindows adds up the count of every active entry.
func SumUsageWindows(xs []UsageWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexUsageWindows keys the entries by ID, last write winning.
func IndexUsageWindows(xs []UsageWindow) map[string]UsageWindow {
	out := make(map[string]UsageWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageWindow builds a valid UsageWindow, for tests and for the seed data the cron
// binary writes.
func SampleUsageWindow(id string, at time.Time) UsageWindow {
	return UsageWindow{
		ID:      id,
		Group:   "group-" + id,
		Source:  "source-" + id,
		Channel: "seeded by usage.SampleUsageWindow",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// UsageQuotaSchema is the stored shape of a UsageQuota.
var UsageQuotaSchema = schema.New("usage.usagequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageQuota is one usage quota.
type UsageQuota struct {
	ID       string
	Name     string
	Category string
	Comment  string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageQuota) Key() string { return "usage/usagequota/" + x.ID }

// Fields encodes x for storage.
func (x UsageQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageQuota decodes a stored record.
func ParseUsageQuota(f codec.Fields) (UsageQuota, error) {
	var x UsageQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagequota: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("usage: usagequota %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("usage: usagequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageQuota) String() string {
	return fmt.Sprintf("UsageQuota(%s, name=%q, version=%d, archived=%t)", x.ID, x.Name, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageQuota) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageQuota) Normalize() UsageQuota {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortUsageQuotas orders by version descending, then by ID.
func SortUsageQuotas(xs []UsageQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageQuotas drops the archived entries, preserving order.
func ActiveUsageQuotas(xs []UsageQuota) []UsageQuota {
	out := make([]UsageQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageQuotas adds up the version of every active entry.
func SumUsageQuotas(xs []UsageQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexUsageQuotas keys the entries by ID, last write winning.
func IndexUsageQuotas(xs []UsageQuota) map[string]UsageQuota {
	out := make(map[string]UsageQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageQuota builds a valid UsageQuota, for tests and for the seed data the cron
// binary writes.
func SampleUsageQuota(id string, at time.Time) UsageQuota {
	return UsageQuota{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Comment:  "seeded by usage.SampleUsageQuota",
		Version:  len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageBucketSchema is the stored shape of a UsageBucket.
var UsageBucketSchema = schema.New("usage.usagebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageBucket is one usage bucket.
type UsageBucket struct {
	ID       string
	Channel  string
	Category string
	Label    string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageBucket) Key() string { return "usage/usagebucket/" + x.ID }

// Fields encodes x for storage.
func (x UsageBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageBucket decodes a stored record.
func ParseUsageBucket(f codec.Fields) (UsageBucket, error) {
	var x UsageBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagebucket: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("usage: usagebucket %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("usage: usagebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageBucket) String() string {
	return fmt.Sprintf("UsageBucket(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageBucket) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageBucket) Normalize() UsageBucket {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortUsageBuckets orders by attempts descending, then by ID.
func SortUsageBuckets(xs []UsageBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageBuckets drops the archived entries, preserving order.
func ActiveUsageBuckets(xs []UsageBucket) []UsageBucket {
	out := make([]UsageBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageBuckets adds up the attempts of every active entry.
func SumUsageBuckets(xs []UsageBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexUsageBuckets keys the entries by ID, last write winning.
func IndexUsageBuckets(xs []UsageBucket) map[string]UsageBucket {
	out := make(map[string]UsageBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageBucket builds a valid UsageBucket, for tests and for the seed data the cron
// binary writes.
func SampleUsageBucket(id string, at time.Time) UsageBucket {
	return UsageBucket{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Label:    "seeded by usage.SampleUsageBucket",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageLaneSchema is the stored shape of a UsageLane.
var UsageLaneSchema = schema.New("usage.usagelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageLane is one usage lane.
type UsageLane struct {
	ID       string
	Label    string
	Owner    string
	Region   string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageLane) Key() string { return "usage/usagelane/" + x.ID }

// Fields encodes x for storage.
func (x UsageLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"owner":    x.Owner,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageLane decodes a stored record.
func ParseUsageLane(f codec.Fields) (UsageLane, error) {
	var x UsageLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagelane: missing id")
	}
	x.Label = f["label"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("usage: usagelane %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("usage: usagelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageLane) String() string {
	return fmt.Sprintf("UsageLane(%s, label=%q, capacity=%d, archived=%t)", x.ID, x.Label, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageLane) Normalize() UsageLane {
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortUsageLanes orders by capacity descending, then by ID.
func SortUsageLanes(xs []UsageLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageLanes drops the archived entries, preserving order.
func ActiveUsageLanes(xs []UsageLane) []UsageLane {
	out := make([]UsageLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageLanes adds up the capacity of every active entry.
func SumUsageLanes(xs []UsageLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexUsageLanes keys the entries by ID, last write winning.
func IndexUsageLanes(xs []UsageLane) map[string]UsageLane {
	out := make(map[string]UsageLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageLane builds a valid UsageLane, for tests and for the seed data the cron
// binary writes.
func SampleUsageLane(id string, at time.Time) UsageLane {
	return UsageLane{
		ID:       id,
		Label:    "label-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by usage.SampleUsageLane",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageSlotSchema is the stored shape of a UsageSlot.
var UsageSlotSchema = schema.New("usage.usageslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageSlot is one usage slot.
type UsageSlot struct {
	ID        string
	Owner     string
	Reference string
	Group     string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UsageSlot) Key() string { return "usage/usageslot/" + x.ID }

// Fields encodes x for storage.
func (x UsageSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"reference": x.Reference,
		"group":     x.Group,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageSlot decodes a stored record.
func ParseUsageSlot(f codec.Fields) (UsageSlot, error) {
	var x UsageSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usageslot: missing id")
	}
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("usage: usageslot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("usage: usageslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usageslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usageslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageSlot) String() string {
	return fmt.Sprintf("UsageSlot(%s, owner=%q, weight=%d, archived=%t)", x.ID, x.Owner, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageSlot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageSlot) Normalize() UsageSlot {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortUsageSlots orders by weight descending, then by ID.
func SortUsageSlots(xs []UsageSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageSlots drops the archived entries, preserving order.
func ActiveUsageSlots(xs []UsageSlot) []UsageSlot {
	out := make([]UsageSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageSlots adds up the weight of every active entry.
func SumUsageSlots(xs []UsageSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexUsageSlots keys the entries by ID, last write winning.
func IndexUsageSlots(xs []UsageSlot) map[string]UsageSlot {
	out := make(map[string]UsageSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageSlot builds a valid UsageSlot, for tests and for the seed data the cron
// binary writes.
func SampleUsageSlot(id string, at time.Time) UsageSlot {
	return UsageSlot{
		ID:        id,
		Owner:     "owner-" + id,
		Reference: "reference-" + id,
		Group:     "seeded by usage.SampleUsageSlot",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UsageTagSchema is the stored shape of a UsageTag.
var UsageTagSchema = schema.New("usage.usagetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageTag is one usage tag.
type UsageTag struct {
	ID       string
	Owner    string
	Region   string
	Comment  string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageTag) Key() string { return "usage/usagetag/" + x.ID }

// Fields encodes x for storage.
func (x UsageTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageTag decodes a stored record.
func ParseUsageTag(f codec.Fields) (UsageTag, error) {
	var x UsageTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagetag: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("usage: usagetag %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("usage: usagetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageTag) String() string {
	return fmt.Sprintf("UsageTag(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageTag) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageTag) Normalize() UsageTag {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortUsageTags orders by capacity descending, then by ID.
func SortUsageTags(xs []UsageTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageTags drops the archived entries, preserving order.
func ActiveUsageTags(xs []UsageTag) []UsageTag {
	out := make([]UsageTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageTags adds up the capacity of every active entry.
func SumUsageTags(xs []UsageTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexUsageTags keys the entries by ID, last write winning.
func IndexUsageTags(xs []UsageTag) map[string]UsageTag {
	out := make(map[string]UsageTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageTag builds a valid UsageTag, for tests and for the seed data the cron
// binary writes.
func SampleUsageTag(id string, at time.Time) UsageTag {
	return UsageTag{
		ID:       id,
		Owner:    "owner-" + id,
		Region:   "region-" + id,
		Comment:  "seeded by usage.SampleUsageTag",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageRuleSchema is the stored shape of a UsageRule.
var UsageRuleSchema = schema.New("usage.usagerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageRule is one usage rule.
type UsageRule struct {
	ID       string
	Label    string
	Category string
	Source   string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageRule) Key() string { return "usage/usagerule/" + x.ID }

// Fields encodes x for storage.
func (x UsageRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageRule decodes a stored record.
func ParseUsageRule(f codec.Fields) (UsageRule, error) {
	var x UsageRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagerule: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("usage: usagerule %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("usage: usagerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageRule) String() string {
	return fmt.Sprintf("UsageRule(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageRule) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageRule) Normalize() UsageRule {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortUsageRules orders by attempts descending, then by ID.
func SortUsageRules(xs []UsageRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageRules drops the archived entries, preserving order.
func ActiveUsageRules(xs []UsageRule) []UsageRule {
	out := make([]UsageRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageRules adds up the attempts of every active entry.
func SumUsageRules(xs []UsageRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexUsageRules keys the entries by ID, last write winning.
func IndexUsageRules(xs []UsageRule) map[string]UsageRule {
	out := make(map[string]UsageRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageRule builds a valid UsageRule, for tests and for the seed data the cron
// binary writes.
func SampleUsageRule(id string, at time.Time) UsageRule {
	return UsageRule{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Source:   "seeded by usage.SampleUsageRule",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageHoldSchema is the stored shape of a UsageHold.
var UsageHoldSchema = schema.New("usage.usagehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageHold is one usage hold.
type UsageHold struct {
	ID        string
	Source    string
	Category  string
	Reference string
	Count     int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UsageHold) Key() string { return "usage/usagehold/" + x.ID }

// Fields encodes x for storage.
func (x UsageHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"category":  x.Category,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageHold decodes a stored record.
func ParseUsageHold(f codec.Fields) (UsageHold, error) {
	var x UsageHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagehold: missing id")
	}
	x.Source = f["source"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("usage: usagehold %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("usage: usagehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageHold) String() string {
	return fmt.Sprintf("UsageHold(%s, source=%q, count=%d, archived=%t)", x.ID, x.Source, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageHold) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageHold) Normalize() UsageHold {
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortUsageHolds orders by count descending, then by ID.
func SortUsageHolds(xs []UsageHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageHolds drops the archived entries, preserving order.
func ActiveUsageHolds(xs []UsageHold) []UsageHold {
	out := make([]UsageHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageHolds adds up the count of every active entry.
func SumUsageHolds(xs []UsageHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexUsageHolds keys the entries by ID, last write winning.
func IndexUsageHolds(xs []UsageHold) map[string]UsageHold {
	out := make(map[string]UsageHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageHold builds a valid UsageHold, for tests and for the seed data the cron
// binary writes.
func SampleUsageHold(id string, at time.Time) UsageHold {
	return UsageHold{
		ID:        id,
		Source:    "source-" + id,
		Category:  "category-" + id,
		Reference: "seeded by usage.SampleUsageHold",
		Count:     len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// UsageClaimSchema is the stored shape of a UsageClaim.
var UsageClaimSchema = schema.New("usage.usageclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageClaim is one usage claim.
type UsageClaim struct {
	ID       string
	Category string
	Region   string
	Comment  string
	Retained int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x UsageClaim) Key() string { return "usage/usageclaim/" + x.ID }

// Fields encodes x for storage.
func (x UsageClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageClaim decodes a stored record.
func ParseUsageClaim(f codec.Fields) (UsageClaim, error) {
	var x UsageClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usageclaim: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("usage: usageclaim %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("usage: usageclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usageclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usageclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageClaim) String() string {
	return fmt.Sprintf("UsageClaim(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageClaim) Normalize() UsageClaim {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortUsageClaims orders by retained descending, then by ID.
func SortUsageClaims(xs []UsageClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageClaims drops the archived entries, preserving order.
func ActiveUsageClaims(xs []UsageClaim) []UsageClaim {
	out := make([]UsageClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageClaims adds up the retained of every active entry.
func SumUsageClaims(xs []UsageClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexUsageClaims keys the entries by ID, last write winning.
func IndexUsageClaims(xs []UsageClaim) map[string]UsageClaim {
	out := make(map[string]UsageClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageClaim builds a valid UsageClaim, for tests and for the seed data the cron
// binary writes.
func SampleUsageClaim(id string, at time.Time) UsageClaim {
	return UsageClaim{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Comment:  "seeded by usage.SampleUsageClaim",
		Retained: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// UsageCursorSchema is the stored shape of a UsageCursor.
var UsageCursorSchema = schema.New("usage.usagecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// UsageCursor is one usage cursor.
type UsageCursor struct {
	ID        string
	Comment   string
	Reference string
	Source    string
	Retained  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x UsageCursor) Key() string { return "usage/usagecursor/" + x.ID }

// Fields encodes x for storage.
func (x UsageCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"source":    x.Source,
		"retained":  strconv.Itoa(x.Retained),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseUsageCursor decodes a stored record.
func ParseUsageCursor(f codec.Fields) (UsageCursor, error) {
	var x UsageCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("usage: usagecursor: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("usage: usagecursor %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("usage: usagecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("usage: usagecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("usage: usagecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x UsageCursor) String() string {
	return fmt.Sprintf("UsageCursor(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x UsageCursor) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x UsageCursor) Normalize() UsageCursor {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortUsageCursors orders by retained descending, then by ID.
func SortUsageCursors(xs []UsageCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveUsageCursors drops the archived entries, preserving order.
func ActiveUsageCursors(xs []UsageCursor) []UsageCursor {
	out := make([]UsageCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumUsageCursors adds up the retained of every active entry.
func SumUsageCursors(xs []UsageCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexUsageCursors keys the entries by ID, last write winning.
func IndexUsageCursors(xs []UsageCursor) map[string]UsageCursor {
	out := make(map[string]UsageCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleUsageCursor builds a valid UsageCursor, for tests and for the seed data the cron
// binary writes.
func SampleUsageCursor(id string, at time.Time) UsageCursor {
	return UsageCursor{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by usage.SampleUsageCursor",
		Retained:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	UsageRecordSchema,
	UsageEntrySchema,
	UsageItemSchema,
	UsageDraftSchema,
	UsageRevisionSchema,
	UsageSnapshotSchema,
	UsageBatchSchema,
	UsageRequestSchema,
	UsageSummarySchema,
	UsagePolicySchema,
	UsageLinkSchema,
	UsageNoteSchema,
	UsageWindowSchema,
	UsageQuotaSchema,
	UsageBucketSchema,
	UsageLaneSchema,
	UsageSlotSchema,
	UsageTagSchema,
	UsageRuleSchema,
	UsageHoldSchema,
	UsageClaimSchema,
	UsageCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"usage/usagerecord/",
	"usage/usageentry/",
	"usage/usageitem/",
	"usage/usagedraft/",
	"usage/usagerevision/",
	"usage/usagesnapshot/",
	"usage/usagebatch/",
	"usage/usagerequest/",
	"usage/usagesummary/",
	"usage/usagepolicy/",
	"usage/usagelink/",
	"usage/usagenote/",
	"usage/usagewindow/",
	"usage/usagequota/",
	"usage/usagebucket/",
	"usage/usagelane/",
	"usage/usageslot/",
	"usage/usagetag/",
	"usage/usagerule/",
	"usage/usagehold/",
	"usage/usageclaim/",
	"usage/usagecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "usage: " + strings.Join(names, ", ")
}
