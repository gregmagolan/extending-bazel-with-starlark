// Package quota holds the quota domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package quota

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// QuotaRecordSchema is the stored shape of a QuotaRecord.
var QuotaRecordSchema = schema.New("quota.quotarecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaRecord is one quota record.
type QuotaRecord struct {
	ID        string
	Label     string
	Reference string
	Region    string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x QuotaRecord) Key() string { return "quota/quotarecord/" + x.ID }

// Fields encodes x for storage.
func (x QuotaRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"reference": x.Reference,
		"region":    x.Region,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaRecord decodes a stored record.
func ParseQuotaRecord(f codec.Fields) (QuotaRecord, error) {
	var x QuotaRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotarecord: missing id")
	}
	x.Label = f["label"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("quota: quotarecord %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("quota: quotarecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotarecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotarecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaRecord) String() string {
	return fmt.Sprintf("QuotaRecord(%s, label=%q, version=%d, archived=%t)", x.ID, x.Label, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaRecord) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaRecord) Normalize() QuotaRecord {
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortQuotaRecords orders by version descending, then by ID.
func SortQuotaRecords(xs []QuotaRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaRecords drops the archived entries, preserving order.
func ActiveQuotaRecords(xs []QuotaRecord) []QuotaRecord {
	out := make([]QuotaRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaRecords adds up the version of every active entry.
func SumQuotaRecords(xs []QuotaRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexQuotaRecords keys the entries by ID, last write winning.
func IndexQuotaRecords(xs []QuotaRecord) map[string]QuotaRecord {
	out := make(map[string]QuotaRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaRecord builds a valid QuotaRecord, for tests and for the seed data the cron
// binary writes.
func SampleQuotaRecord(id string, at time.Time) QuotaRecord {
	return QuotaRecord{
		ID:        id,
		Label:     "label-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by quota.SampleQuotaRecord",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// QuotaEntrySchema is the stored shape of a QuotaEntry.
var QuotaEntrySchema = schema.New("quota.quotaentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaEntry is one quota entry.
type QuotaEntry struct {
	ID       string
	Region   string
	Label    string
	Comment  string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaEntry) Key() string { return "quota/quotaentry/" + x.ID }

// Fields encodes x for storage.
func (x QuotaEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"comment":  x.Comment,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaEntry decodes a stored record.
func ParseQuotaEntry(f codec.Fields) (QuotaEntry, error) {
	var x QuotaEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotaentry: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotaentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("quota: quotaentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotaentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotaentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaEntry) String() string {
	return fmt.Sprintf("QuotaEntry(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaEntry) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaEntry) Normalize() QuotaEntry {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortQuotaEntrys orders by attempts descending, then by ID.
func SortQuotaEntrys(xs []QuotaEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaEntrys drops the archived entries, preserving order.
func ActiveQuotaEntrys(xs []QuotaEntry) []QuotaEntry {
	out := make([]QuotaEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaEntrys adds up the attempts of every active entry.
func SumQuotaEntrys(xs []QuotaEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexQuotaEntrys keys the entries by ID, last write winning.
func IndexQuotaEntrys(xs []QuotaEntry) map[string]QuotaEntry {
	out := make(map[string]QuotaEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaEntry builds a valid QuotaEntry, for tests and for the seed data the cron
// binary writes.
func SampleQuotaEntry(id string, at time.Time) QuotaEntry {
	return QuotaEntry{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Comment:  "seeded by quota.SampleQuotaEntry",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaItemSchema is the stored shape of a QuotaItem.
var QuotaItemSchema = schema.New("quota.quotaitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaItem is one quota item.
type QuotaItem struct {
	ID       string
	Comment  string
	Owner    string
	Source   string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaItem) Key() string { return "quota/quotaitem/" + x.ID }

// Fields encodes x for storage.
func (x QuotaItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaItem decodes a stored record.
func ParseQuotaItem(f codec.Fields) (QuotaItem, error) {
	var x QuotaItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotaitem: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotaitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("quota: quotaitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotaitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotaitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaItem) String() string {
	return fmt.Sprintf("QuotaItem(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaItem) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaItem) Normalize() QuotaItem {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortQuotaItems orders by attempts descending, then by ID.
func SortQuotaItems(xs []QuotaItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaItems drops the archived entries, preserving order.
func ActiveQuotaItems(xs []QuotaItem) []QuotaItem {
	out := make([]QuotaItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaItems adds up the attempts of every active entry.
func SumQuotaItems(xs []QuotaItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexQuotaItems keys the entries by ID, last write winning.
func IndexQuotaItems(xs []QuotaItem) map[string]QuotaItem {
	out := make(map[string]QuotaItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaItem builds a valid QuotaItem, for tests and for the seed data the cron
// binary writes.
func SampleQuotaItem(id string, at time.Time) QuotaItem {
	return QuotaItem{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by quota.SampleQuotaItem",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaDraftSchema is the stored shape of a QuotaDraft.
var QuotaDraftSchema = schema.New("quota.quotadraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaDraft is one quota draft.
type QuotaDraft struct {
	ID       string
	Group    string
	Region   string
	Category string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaDraft) Key() string { return "quota/quotadraft/" + x.ID }

// Fields encodes x for storage.
func (x QuotaDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"region":   x.Region,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaDraft decodes a stored record.
func ParseQuotaDraft(f codec.Fields) (QuotaDraft, error) {
	var x QuotaDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotadraft: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("quota: quotadraft %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("quota: quotadraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotadraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotadraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaDraft) String() string {
	return fmt.Sprintf("QuotaDraft(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaDraft) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaDraft) Normalize() QuotaDraft {
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortQuotaDrafts orders by version descending, then by ID.
func SortQuotaDrafts(xs []QuotaDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaDrafts drops the archived entries, preserving order.
func ActiveQuotaDrafts(xs []QuotaDraft) []QuotaDraft {
	out := make([]QuotaDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaDrafts adds up the version of every active entry.
func SumQuotaDrafts(xs []QuotaDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexQuotaDrafts keys the entries by ID, last write winning.
func IndexQuotaDrafts(xs []QuotaDraft) map[string]QuotaDraft {
	out := make(map[string]QuotaDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaDraft builds a valid QuotaDraft, for tests and for the seed data the cron
// binary writes.
func SampleQuotaDraft(id string, at time.Time) QuotaDraft {
	return QuotaDraft{
		ID:       id,
		Group:    "group-" + id,
		Region:   "region-" + id,
		Category: "seeded by quota.SampleQuotaDraft",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaRevisionSchema is the stored shape of a QuotaRevision.
var QuotaRevisionSchema = schema.New("quota.quotarevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaRevision is one quota revision.
type QuotaRevision struct {
	ID        string
	Reference string
	Category  string
	Owner     string
	Priority  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x QuotaRevision) Key() string { return "quota/quotarevision/" + x.ID }

// Fields encodes x for storage.
func (x QuotaRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"category":  x.Category,
		"owner":     x.Owner,
		"priority":  strconv.Itoa(x.Priority),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaRevision decodes a stored record.
func ParseQuotaRevision(f codec.Fields) (QuotaRevision, error) {
	var x QuotaRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotarevision: missing id")
	}
	x.Reference = f["reference"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("quota: quotarevision %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("quota: quotarevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotarevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotarevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaRevision) String() string {
	return fmt.Sprintf("QuotaRevision(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaRevision) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaRevision) Normalize() QuotaRevision {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortQuotaRevisions orders by priority descending, then by ID.
func SortQuotaRevisions(xs []QuotaRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaRevisions drops the archived entries, preserving order.
func ActiveQuotaRevisions(xs []QuotaRevision) []QuotaRevision {
	out := make([]QuotaRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaRevisions adds up the priority of every active entry.
func SumQuotaRevisions(xs []QuotaRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexQuotaRevisions keys the entries by ID, last write winning.
func IndexQuotaRevisions(xs []QuotaRevision) map[string]QuotaRevision {
	out := make(map[string]QuotaRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaRevision builds a valid QuotaRevision, for tests and for the seed data the cron
// binary writes.
func SampleQuotaRevision(id string, at time.Time) QuotaRevision {
	return QuotaRevision{
		ID:        id,
		Reference: "reference-" + id,
		Category:  "category-" + id,
		Owner:     "seeded by quota.SampleQuotaRevision",
		Priority:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// QuotaSnapshotSchema is the stored shape of a QuotaSnapshot.
var QuotaSnapshotSchema = schema.New("quota.quotasnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaSnapshot is one quota snapshot.
type QuotaSnapshot struct {
	ID       string
	Label    string
	Name     string
	Group    string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaSnapshot) Key() string { return "quota/quotasnapshot/" + x.ID }

// Fields encodes x for storage.
func (x QuotaSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"name":     x.Name,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaSnapshot decodes a stored record.
func ParseQuotaSnapshot(f codec.Fields) (QuotaSnapshot, error) {
	var x QuotaSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotasnapshot: missing id")
	}
	x.Label = f["label"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("quota: quotasnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("quota: quotasnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotasnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotasnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaSnapshot) String() string {
	return fmt.Sprintf("QuotaSnapshot(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaSnapshot) Normalize() QuotaSnapshot {
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortQuotaSnapshots orders by priority descending, then by ID.
func SortQuotaSnapshots(xs []QuotaSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaSnapshots drops the archived entries, preserving order.
func ActiveQuotaSnapshots(xs []QuotaSnapshot) []QuotaSnapshot {
	out := make([]QuotaSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaSnapshots adds up the priority of every active entry.
func SumQuotaSnapshots(xs []QuotaSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexQuotaSnapshots keys the entries by ID, last write winning.
func IndexQuotaSnapshots(xs []QuotaSnapshot) map[string]QuotaSnapshot {
	out := make(map[string]QuotaSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaSnapshot builds a valid QuotaSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleQuotaSnapshot(id string, at time.Time) QuotaSnapshot {
	return QuotaSnapshot{
		ID:       id,
		Label:    "label-" + id,
		Name:     "name-" + id,
		Group:    "seeded by quota.SampleQuotaSnapshot",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaBatchSchema is the stored shape of a QuotaBatch.
var QuotaBatchSchema = schema.New("quota.quotabatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaBatch is one quota batch.
type QuotaBatch struct {
	ID        string
	Category  string
	Reference string
	Source    string
	Capacity  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x QuotaBatch) Key() string { return "quota/quotabatch/" + x.ID }

// Fields encodes x for storage.
func (x QuotaBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"source":    x.Source,
		"capacity":  strconv.Itoa(x.Capacity),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaBatch decodes a stored record.
func ParseQuotaBatch(f codec.Fields) (QuotaBatch, error) {
	var x QuotaBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotabatch: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("quota: quotabatch %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("quota: quotabatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotabatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotabatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaBatch) String() string {
	return fmt.Sprintf("QuotaBatch(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaBatch) Normalize() QuotaBatch {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortQuotaBatchs orders by capacity descending, then by ID.
func SortQuotaBatchs(xs []QuotaBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaBatchs drops the archived entries, preserving order.
func ActiveQuotaBatchs(xs []QuotaBatch) []QuotaBatch {
	out := make([]QuotaBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaBatchs adds up the capacity of every active entry.
func SumQuotaBatchs(xs []QuotaBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexQuotaBatchs keys the entries by ID, last write winning.
func IndexQuotaBatchs(xs []QuotaBatch) map[string]QuotaBatch {
	out := make(map[string]QuotaBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaBatch builds a valid QuotaBatch, for tests and for the seed data the cron
// binary writes.
func SampleQuotaBatch(id string, at time.Time) QuotaBatch {
	return QuotaBatch{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by quota.SampleQuotaBatch",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// QuotaRequestSchema is the stored shape of a QuotaRequest.
var QuotaRequestSchema = schema.New("quota.quotarequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaRequest is one quota request.
type QuotaRequest struct {
	ID       string
	Region   string
	Name     string
	Label    string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaRequest) Key() string { return "quota/quotarequest/" + x.ID }

// Fields encodes x for storage.
func (x QuotaRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaRequest decodes a stored record.
func ParseQuotaRequest(f codec.Fields) (QuotaRequest, error) {
	var x QuotaRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotarequest: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("quota: quotarequest %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("quota: quotarequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotarequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotarequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaRequest) String() string {
	return fmt.Sprintf("QuotaRequest(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaRequest) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaRequest) Normalize() QuotaRequest {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortQuotaRequests orders by capacity descending, then by ID.
func SortQuotaRequests(xs []QuotaRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaRequests drops the archived entries, preserving order.
func ActiveQuotaRequests(xs []QuotaRequest) []QuotaRequest {
	out := make([]QuotaRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaRequests adds up the capacity of every active entry.
func SumQuotaRequests(xs []QuotaRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexQuotaRequests keys the entries by ID, last write winning.
func IndexQuotaRequests(xs []QuotaRequest) map[string]QuotaRequest {
	out := make(map[string]QuotaRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaRequest builds a valid QuotaRequest, for tests and for the seed data the cron
// binary writes.
func SampleQuotaRequest(id string, at time.Time) QuotaRequest {
	return QuotaRequest{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Label:    "seeded by quota.SampleQuotaRequest",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaSummarySchema is the stored shape of a QuotaSummary.
var QuotaSummarySchema = schema.New("quota.quotasummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaSummary is one quota summary.
type QuotaSummary struct {
	ID       string
	Region   string
	Channel  string
	Label    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaSummary) Key() string { return "quota/quotasummary/" + x.ID }

// Fields encodes x for storage.
func (x QuotaSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaSummary decodes a stored record.
func ParseQuotaSummary(f codec.Fields) (QuotaSummary, error) {
	var x QuotaSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotasummary: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("quota: quotasummary %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotasummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotasummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotasummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaSummary) String() string {
	return fmt.Sprintf("QuotaSummary(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaSummary) Normalize() QuotaSummary {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortQuotaSummarys orders by capacity descending, then by ID.
func SortQuotaSummarys(xs []QuotaSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaSummarys drops the archived entries, preserving order.
func ActiveQuotaSummarys(xs []QuotaSummary) []QuotaSummary {
	out := make([]QuotaSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaSummarys adds up the capacity of every active entry.
func SumQuotaSummarys(xs []QuotaSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexQuotaSummarys keys the entries by ID, last write winning.
func IndexQuotaSummarys(xs []QuotaSummary) map[string]QuotaSummary {
	out := make(map[string]QuotaSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaSummary builds a valid QuotaSummary, for tests and for the seed data the cron
// binary writes.
func SampleQuotaSummary(id string, at time.Time) QuotaSummary {
	return QuotaSummary{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by quota.SampleQuotaSummary",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaPolicySchema is the stored shape of a QuotaPolicy.
var QuotaPolicySchema = schema.New("quota.quotapolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaPolicy is one quota policy.
type QuotaPolicy struct {
	ID        string
	Channel   string
	Label     string
	Reference string
	Offset    int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x QuotaPolicy) Key() string { return "quota/quotapolicy/" + x.ID }

// Fields encodes x for storage.
func (x QuotaPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"label":     x.Label,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaPolicy decodes a stored record.
func ParseQuotaPolicy(f codec.Fields) (QuotaPolicy, error) {
	var x QuotaPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotapolicy: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("quota: quotapolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("quota: quotapolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotapolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotapolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaPolicy) String() string {
	return fmt.Sprintf("QuotaPolicy(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaPolicy) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaPolicy) Normalize() QuotaPolicy {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortQuotaPolicys orders by offset descending, then by ID.
func SortQuotaPolicys(xs []QuotaPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaPolicys drops the archived entries, preserving order.
func ActiveQuotaPolicys(xs []QuotaPolicy) []QuotaPolicy {
	out := make([]QuotaPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaPolicys adds up the offset of every active entry.
func SumQuotaPolicys(xs []QuotaPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexQuotaPolicys keys the entries by ID, last write winning.
func IndexQuotaPolicys(xs []QuotaPolicy) map[string]QuotaPolicy {
	out := make(map[string]QuotaPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaPolicy builds a valid QuotaPolicy, for tests and for the seed data the cron
// binary writes.
func SampleQuotaPolicy(id string, at time.Time) QuotaPolicy {
	return QuotaPolicy{
		ID:        id,
		Channel:   "channel-" + id,
		Label:     "label-" + id,
		Reference: "seeded by quota.SampleQuotaPolicy",
		Offset:    len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// QuotaLinkSchema is the stored shape of a QuotaLink.
var QuotaLinkSchema = schema.New("quota.quotalink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaLink is one quota link.
type QuotaLink struct {
	ID       string
	Region   string
	Category string
	Comment  string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaLink) Key() string { return "quota/quotalink/" + x.ID }

// Fields encodes x for storage.
func (x QuotaLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"comment":  x.Comment,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaLink decodes a stored record.
func ParseQuotaLink(f codec.Fields) (QuotaLink, error) {
	var x QuotaLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotalink: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotalink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("quota: quotalink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotalink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotalink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaLink) String() string {
	return fmt.Sprintf("QuotaLink(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaLink) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaLink) Normalize() QuotaLink {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortQuotaLinks orders by attempts descending, then by ID.
func SortQuotaLinks(xs []QuotaLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaLinks drops the archived entries, preserving order.
func ActiveQuotaLinks(xs []QuotaLink) []QuotaLink {
	out := make([]QuotaLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaLinks adds up the attempts of every active entry.
func SumQuotaLinks(xs []QuotaLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexQuotaLinks keys the entries by ID, last write winning.
func IndexQuotaLinks(xs []QuotaLink) map[string]QuotaLink {
	out := make(map[string]QuotaLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaLink builds a valid QuotaLink, for tests and for the seed data the cron
// binary writes.
func SampleQuotaLink(id string, at time.Time) QuotaLink {
	return QuotaLink{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Comment:  "seeded by quota.SampleQuotaLink",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaNoteSchema is the stored shape of a QuotaNote.
var QuotaNoteSchema = schema.New("quota.quotanote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaNote is one quota note.
type QuotaNote struct {
	ID       string
	Source   string
	Category string
	Owner    string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaNote) Key() string { return "quota/quotanote/" + x.ID }

// Fields encodes x for storage.
func (x QuotaNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"category": x.Category,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaNote decodes a stored record.
func ParseQuotaNote(f codec.Fields) (QuotaNote, error) {
	var x QuotaNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotanote: missing id")
	}
	x.Source = f["source"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotanote %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("quota: quotanote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotanote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotanote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaNote) String() string {
	return fmt.Sprintf("QuotaNote(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaNote) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaNote) Normalize() QuotaNote {
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortQuotaNotes orders by attempts descending, then by ID.
func SortQuotaNotes(xs []QuotaNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaNotes drops the archived entries, preserving order.
func ActiveQuotaNotes(xs []QuotaNote) []QuotaNote {
	out := make([]QuotaNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaNotes adds up the attempts of every active entry.
func SumQuotaNotes(xs []QuotaNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexQuotaNotes keys the entries by ID, last write winning.
func IndexQuotaNotes(xs []QuotaNote) map[string]QuotaNote {
	out := make(map[string]QuotaNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaNote builds a valid QuotaNote, for tests and for the seed data the cron
// binary writes.
func SampleQuotaNote(id string, at time.Time) QuotaNote {
	return QuotaNote{
		ID:       id,
		Source:   "source-" + id,
		Category: "category-" + id,
		Owner:    "seeded by quota.SampleQuotaNote",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaWindowSchema is the stored shape of a QuotaWindow.
var QuotaWindowSchema = schema.New("quota.quotawindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaWindow is one quota window.
type QuotaWindow struct {
	ID       string
	Channel  string
	Comment  string
	Category string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaWindow) Key() string { return "quota/quotawindow/" + x.ID }

// Fields encodes x for storage.
func (x QuotaWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaWindow decodes a stored record.
func ParseQuotaWindow(f codec.Fields) (QuotaWindow, error) {
	var x QuotaWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotawindow: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotawindow %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("quota: quotawindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotawindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotawindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaWindow) String() string {
	return fmt.Sprintf("QuotaWindow(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaWindow) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaWindow) Normalize() QuotaWindow {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortQuotaWindows orders by attempts descending, then by ID.
func SortQuotaWindows(xs []QuotaWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaWindows drops the archived entries, preserving order.
func ActiveQuotaWindows(xs []QuotaWindow) []QuotaWindow {
	out := make([]QuotaWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaWindows adds up the attempts of every active entry.
func SumQuotaWindows(xs []QuotaWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexQuotaWindows keys the entries by ID, last write winning.
func IndexQuotaWindows(xs []QuotaWindow) map[string]QuotaWindow {
	out := make(map[string]QuotaWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaWindow builds a valid QuotaWindow, for tests and for the seed data the cron
// binary writes.
func SampleQuotaWindow(id string, at time.Time) QuotaWindow {
	return QuotaWindow{
		ID:       id,
		Channel:  "channel-" + id,
		Comment:  "comment-" + id,
		Category: "seeded by quota.SampleQuotaWindow",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaQuotaSchema is the stored shape of a QuotaQuota.
var QuotaQuotaSchema = schema.New("quota.quotaquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaQuota is one quota quota.
type QuotaQuota struct {
	ID        string
	Name      string
	Reference string
	Comment   string
	Attempts  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x QuotaQuota) Key() string { return "quota/quotaquota/" + x.ID }

// Fields encodes x for storage.
func (x QuotaQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"comment":   x.Comment,
		"attempts":  strconv.Itoa(x.Attempts),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaQuota decodes a stored record.
func ParseQuotaQuota(f codec.Fields) (QuotaQuota, error) {
	var x QuotaQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotaquota: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotaquota %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("quota: quotaquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotaquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotaquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaQuota) String() string {
	return fmt.Sprintf("QuotaQuota(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaQuota) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaQuota) Normalize() QuotaQuota {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortQuotaQuotas orders by attempts descending, then by ID.
func SortQuotaQuotas(xs []QuotaQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaQuotas drops the archived entries, preserving order.
func ActiveQuotaQuotas(xs []QuotaQuota) []QuotaQuota {
	out := make([]QuotaQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaQuotas adds up the attempts of every active entry.
func SumQuotaQuotas(xs []QuotaQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexQuotaQuotas keys the entries by ID, last write winning.
func IndexQuotaQuotas(xs []QuotaQuota) map[string]QuotaQuota {
	out := make(map[string]QuotaQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaQuota builds a valid QuotaQuota, for tests and for the seed data the cron
// binary writes.
func SampleQuotaQuota(id string, at time.Time) QuotaQuota {
	return QuotaQuota{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by quota.SampleQuotaQuota",
		Attempts:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// QuotaBucketSchema is the stored shape of a QuotaBucket.
var QuotaBucketSchema = schema.New("quota.quotabucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaBucket is one quota bucket.
type QuotaBucket struct {
	ID       string
	Label    string
	Channel  string
	Comment  string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaBucket) Key() string { return "quota/quotabucket/" + x.ID }

// Fields encodes x for storage.
func (x QuotaBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaBucket decodes a stored record.
func ParseQuotaBucket(f codec.Fields) (QuotaBucket, error) {
	var x QuotaBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotabucket: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("quota: quotabucket %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("quota: quotabucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotabucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotabucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaBucket) String() string {
	return fmt.Sprintf("QuotaBucket(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaBucket) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaBucket) Normalize() QuotaBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortQuotaBuckets orders by weight descending, then by ID.
func SortQuotaBuckets(xs []QuotaBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaBuckets drops the archived entries, preserving order.
func ActiveQuotaBuckets(xs []QuotaBucket) []QuotaBucket {
	out := make([]QuotaBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaBuckets adds up the weight of every active entry.
func SumQuotaBuckets(xs []QuotaBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexQuotaBuckets keys the entries by ID, last write winning.
func IndexQuotaBuckets(xs []QuotaBucket) map[string]QuotaBucket {
	out := make(map[string]QuotaBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaBucket builds a valid QuotaBucket, for tests and for the seed data the cron
// binary writes.
func SampleQuotaBucket(id string, at time.Time) QuotaBucket {
	return QuotaBucket{
		ID:       id,
		Label:    "label-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by quota.SampleQuotaBucket",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaLaneSchema is the stored shape of a QuotaLane.
var QuotaLaneSchema = schema.New("quota.quotalane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaLane is one quota lane.
type QuotaLane struct {
	ID       string
	Channel  string
	Comment  string
	Source   string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaLane) Key() string { return "quota/quotalane/" + x.ID }

// Fields encodes x for storage.
func (x QuotaLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaLane decodes a stored record.
func ParseQuotaLane(f codec.Fields) (QuotaLane, error) {
	var x QuotaLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotalane: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("quota: quotalane %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("quota: quotalane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotalane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotalane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaLane) String() string {
	return fmt.Sprintf("QuotaLane(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaLane) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaLane) Normalize() QuotaLane {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortQuotaLanes orders by weight descending, then by ID.
func SortQuotaLanes(xs []QuotaLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaLanes drops the archived entries, preserving order.
func ActiveQuotaLanes(xs []QuotaLane) []QuotaLane {
	out := make([]QuotaLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaLanes adds up the weight of every active entry.
func SumQuotaLanes(xs []QuotaLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexQuotaLanes keys the entries by ID, last write winning.
func IndexQuotaLanes(xs []QuotaLane) map[string]QuotaLane {
	out := make(map[string]QuotaLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaLane builds a valid QuotaLane, for tests and for the seed data the cron
// binary writes.
func SampleQuotaLane(id string, at time.Time) QuotaLane {
	return QuotaLane{
		ID:       id,
		Channel:  "channel-" + id,
		Comment:  "comment-" + id,
		Source:   "seeded by quota.SampleQuotaLane",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaSlotSchema is the stored shape of a QuotaSlot.
var QuotaSlotSchema = schema.New("quota.quotaslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaSlot is one quota slot.
type QuotaSlot struct {
	ID       string
	Channel  string
	Group    string
	Comment  string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaSlot) Key() string { return "quota/quotaslot/" + x.ID }

// Fields encodes x for storage.
func (x QuotaSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaSlot decodes a stored record.
func ParseQuotaSlot(f codec.Fields) (QuotaSlot, error) {
	var x QuotaSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotaslot: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("quota: quotaslot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("quota: quotaslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotaslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotaslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaSlot) String() string {
	return fmt.Sprintf("QuotaSlot(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaSlot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaSlot) Normalize() QuotaSlot {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortQuotaSlots orders by retained descending, then by ID.
func SortQuotaSlots(xs []QuotaSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaSlots drops the archived entries, preserving order.
func ActiveQuotaSlots(xs []QuotaSlot) []QuotaSlot {
	out := make([]QuotaSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaSlots adds up the retained of every active entry.
func SumQuotaSlots(xs []QuotaSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexQuotaSlots keys the entries by ID, last write winning.
func IndexQuotaSlots(xs []QuotaSlot) map[string]QuotaSlot {
	out := make(map[string]QuotaSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaSlot builds a valid QuotaSlot, for tests and for the seed data the cron
// binary writes.
func SampleQuotaSlot(id string, at time.Time) QuotaSlot {
	return QuotaSlot{
		ID:       id,
		Channel:  "channel-" + id,
		Group:    "group-" + id,
		Comment:  "seeded by quota.SampleQuotaSlot",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaTagSchema is the stored shape of a QuotaTag.
var QuotaTagSchema = schema.New("quota.quotatag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaTag is one quota tag.
type QuotaTag struct {
	ID       string
	Owner    string
	Name     string
	Channel  string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaTag) Key() string { return "quota/quotatag/" + x.ID }

// Fields encodes x for storage.
func (x QuotaTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"name":     x.Name,
		"channel":  x.Channel,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaTag decodes a stored record.
func ParseQuotaTag(f codec.Fields) (QuotaTag, error) {
	var x QuotaTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotatag: missing id")
	}
	x.Owner = f["owner"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("quota: quotatag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("quota: quotatag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotatag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotatag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaTag) String() string {
	return fmt.Sprintf("QuotaTag(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaTag) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaTag) Normalize() QuotaTag {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortQuotaTags orders by offset descending, then by ID.
func SortQuotaTags(xs []QuotaTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaTags drops the archived entries, preserving order.
func ActiveQuotaTags(xs []QuotaTag) []QuotaTag {
	out := make([]QuotaTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaTags adds up the offset of every active entry.
func SumQuotaTags(xs []QuotaTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexQuotaTags keys the entries by ID, last write winning.
func IndexQuotaTags(xs []QuotaTag) map[string]QuotaTag {
	out := make(map[string]QuotaTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaTag builds a valid QuotaTag, for tests and for the seed data the cron
// binary writes.
func SampleQuotaTag(id string, at time.Time) QuotaTag {
	return QuotaTag{
		ID:       id,
		Owner:    "owner-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by quota.SampleQuotaTag",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaRuleSchema is the stored shape of a QuotaRule.
var QuotaRuleSchema = schema.New("quota.quotarule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaRule is one quota rule.
type QuotaRule struct {
	ID       string
	Category string
	Comment  string
	Group    string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaRule) Key() string { return "quota/quotarule/" + x.ID }

// Fields encodes x for storage.
func (x QuotaRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"comment":  x.Comment,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaRule decodes a stored record.
func ParseQuotaRule(f codec.Fields) (QuotaRule, error) {
	var x QuotaRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotarule: missing id")
	}
	x.Category = f["category"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("quota: quotarule %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("quota: quotarule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotarule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotarule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaRule) String() string {
	return fmt.Sprintf("QuotaRule(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaRule) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaRule) Normalize() QuotaRule {
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortQuotaRules orders by version descending, then by ID.
func SortQuotaRules(xs []QuotaRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaRules drops the archived entries, preserving order.
func ActiveQuotaRules(xs []QuotaRule) []QuotaRule {
	out := make([]QuotaRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaRules adds up the version of every active entry.
func SumQuotaRules(xs []QuotaRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexQuotaRules keys the entries by ID, last write winning.
func IndexQuotaRules(xs []QuotaRule) map[string]QuotaRule {
	out := make(map[string]QuotaRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaRule builds a valid QuotaRule, for tests and for the seed data the cron
// binary writes.
func SampleQuotaRule(id string, at time.Time) QuotaRule {
	return QuotaRule{
		ID:       id,
		Category: "category-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by quota.SampleQuotaRule",
		Version:  len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaHoldSchema is the stored shape of a QuotaHold.
var QuotaHoldSchema = schema.New("quota.quotahold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaHold is one quota hold.
type QuotaHold struct {
	ID       string
	Owner    string
	Name     string
	Group    string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaHold) Key() string { return "quota/quotahold/" + x.ID }

// Fields encodes x for storage.
func (x QuotaHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"name":     x.Name,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaHold decodes a stored record.
func ParseQuotaHold(f codec.Fields) (QuotaHold, error) {
	var x QuotaHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotahold: missing id")
	}
	x.Owner = f["owner"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("quota: quotahold %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("quota: quotahold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotahold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotahold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaHold) String() string {
	return fmt.Sprintf("QuotaHold(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaHold) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaHold) Normalize() QuotaHold {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortQuotaHolds orders by capacity descending, then by ID.
func SortQuotaHolds(xs []QuotaHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaHolds drops the archived entries, preserving order.
func ActiveQuotaHolds(xs []QuotaHold) []QuotaHold {
	out := make([]QuotaHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaHolds adds up the capacity of every active entry.
func SumQuotaHolds(xs []QuotaHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexQuotaHolds keys the entries by ID, last write winning.
func IndexQuotaHolds(xs []QuotaHold) map[string]QuotaHold {
	out := make(map[string]QuotaHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaHold builds a valid QuotaHold, for tests and for the seed data the cron
// binary writes.
func SampleQuotaHold(id string, at time.Time) QuotaHold {
	return QuotaHold{
		ID:       id,
		Owner:    "owner-" + id,
		Name:     "name-" + id,
		Group:    "seeded by quota.SampleQuotaHold",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaClaimSchema is the stored shape of a QuotaClaim.
var QuotaClaimSchema = schema.New("quota.quotaclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaClaim is one quota claim.
type QuotaClaim struct {
	ID       string
	Source   string
	Comment  string
	Label    string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaClaim) Key() string { return "quota/quotaclaim/" + x.ID }

// Fields encodes x for storage.
func (x QuotaClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaClaim decodes a stored record.
func ParseQuotaClaim(f codec.Fields) (QuotaClaim, error) {
	var x QuotaClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotaclaim: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("quota: quotaclaim %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("quota: quotaclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotaclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotaclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaClaim) String() string {
	return fmt.Sprintf("QuotaClaim(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaClaim) Normalize() QuotaClaim {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortQuotaClaims orders by retained descending, then by ID.
func SortQuotaClaims(xs []QuotaClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaClaims drops the archived entries, preserving order.
func ActiveQuotaClaims(xs []QuotaClaim) []QuotaClaim {
	out := make([]QuotaClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaClaims adds up the retained of every active entry.
func SumQuotaClaims(xs []QuotaClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexQuotaClaims keys the entries by ID, last write winning.
func IndexQuotaClaims(xs []QuotaClaim) map[string]QuotaClaim {
	out := make(map[string]QuotaClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaClaim builds a valid QuotaClaim, for tests and for the seed data the cron
// binary writes.
func SampleQuotaClaim(id string, at time.Time) QuotaClaim {
	return QuotaClaim{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Label:    "seeded by quota.SampleQuotaClaim",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// QuotaCursorSchema is the stored shape of a QuotaCursor.
var QuotaCursorSchema = schema.New("quota.quotacursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// QuotaCursor is one quota cursor.
type QuotaCursor struct {
	ID       string
	Group    string
	Region   string
	Owner    string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x QuotaCursor) Key() string { return "quota/quotacursor/" + x.ID }

// Fields encodes x for storage.
func (x QuotaCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"region":   x.Region,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseQuotaCursor decodes a stored record.
func ParseQuotaCursor(f codec.Fields) (QuotaCursor, error) {
	var x QuotaCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("quota: quotacursor: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("quota: quotacursor %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("quota: quotacursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("quota: quotacursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("quota: quotacursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x QuotaCursor) String() string {
	return fmt.Sprintf("QuotaCursor(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x QuotaCursor) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x QuotaCursor) Normalize() QuotaCursor {
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortQuotaCursors orders by priority descending, then by ID.
func SortQuotaCursors(xs []QuotaCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveQuotaCursors drops the archived entries, preserving order.
func ActiveQuotaCursors(xs []QuotaCursor) []QuotaCursor {
	out := make([]QuotaCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumQuotaCursors adds up the priority of every active entry.
func SumQuotaCursors(xs []QuotaCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexQuotaCursors keys the entries by ID, last write winning.
func IndexQuotaCursors(xs []QuotaCursor) map[string]QuotaCursor {
	out := make(map[string]QuotaCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleQuotaCursor builds a valid QuotaCursor, for tests and for the seed data the cron
// binary writes.
func SampleQuotaCursor(id string, at time.Time) QuotaCursor {
	return QuotaCursor{
		ID:       id,
		Group:    "group-" + id,
		Region:   "region-" + id,
		Owner:    "seeded by quota.SampleQuotaCursor",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	QuotaRecordSchema,
	QuotaEntrySchema,
	QuotaItemSchema,
	QuotaDraftSchema,
	QuotaRevisionSchema,
	QuotaSnapshotSchema,
	QuotaBatchSchema,
	QuotaRequestSchema,
	QuotaSummarySchema,
	QuotaPolicySchema,
	QuotaLinkSchema,
	QuotaNoteSchema,
	QuotaWindowSchema,
	QuotaQuotaSchema,
	QuotaBucketSchema,
	QuotaLaneSchema,
	QuotaSlotSchema,
	QuotaTagSchema,
	QuotaRuleSchema,
	QuotaHoldSchema,
	QuotaClaimSchema,
	QuotaCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"quota/quotarecord/",
	"quota/quotaentry/",
	"quota/quotaitem/",
	"quota/quotadraft/",
	"quota/quotarevision/",
	"quota/quotasnapshot/",
	"quota/quotabatch/",
	"quota/quotarequest/",
	"quota/quotasummary/",
	"quota/quotapolicy/",
	"quota/quotalink/",
	"quota/quotanote/",
	"quota/quotawindow/",
	"quota/quotaquota/",
	"quota/quotabucket/",
	"quota/quotalane/",
	"quota/quotaslot/",
	"quota/quotatag/",
	"quota/quotarule/",
	"quota/quotahold/",
	"quota/quotaclaim/",
	"quota/quotacursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "quota: " + strings.Join(names, ", ")
}
