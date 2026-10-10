// Package audit holds the audit domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package audit

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// AuditRecordSchema is the stored shape of a AuditRecord.
var AuditRecordSchema = schema.New("audit.auditrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditRecord is one audit record.
type AuditRecord struct {
	ID       string
	Owner    string
	Group    string
	Category string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditRecord) Key() string { return "audit/auditrecord/" + x.ID }

// Fields encodes x for storage.
func (x AuditRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"category": x.Category,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditRecord decodes a stored record.
func ParseAuditRecord(f codec.Fields) (AuditRecord, error) {
	var x AuditRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditrecord: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: auditrecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("audit: auditrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditRecord) String() string {
	return fmt.Sprintf("AuditRecord(%s, owner=%q, weight=%d, archived=%t)", x.ID, x.Owner, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditRecord) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditRecord) Normalize() AuditRecord {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortAuditRecords orders by weight descending, then by ID.
func SortAuditRecords(xs []AuditRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditRecords drops the archived entries, preserving order.
func ActiveAuditRecords(xs []AuditRecord) []AuditRecord {
	out := make([]AuditRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditRecords adds up the weight of every active entry.
func SumAuditRecords(xs []AuditRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexAuditRecords keys the entries by ID, last write winning.
func IndexAuditRecords(xs []AuditRecord) map[string]AuditRecord {
	out := make(map[string]AuditRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditRecord builds a valid AuditRecord, for tests and for the seed data the cron
// binary writes.
func SampleAuditRecord(id string, at time.Time) AuditRecord {
	return AuditRecord{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Category: "seeded by audit.SampleAuditRecord",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditEntrySchema is the stored shape of a AuditEntry.
var AuditEntrySchema = schema.New("audit.auditentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditEntry is one audit entry.
type AuditEntry struct {
	ID       string
	Label    string
	Region   string
	Category string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditEntry) Key() string { return "audit/auditentry/" + x.ID }

// Fields encodes x for storage.
func (x AuditEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"category": x.Category,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditEntry decodes a stored record.
func ParseAuditEntry(f codec.Fields) (AuditEntry, error) {
	var x AuditEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditentry: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("audit: auditentry %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("audit: auditentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditEntry) String() string {
	return fmt.Sprintf("AuditEntry(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditEntry) Normalize() AuditEntry {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortAuditEntrys orders by offset descending, then by ID.
func SortAuditEntrys(xs []AuditEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditEntrys drops the archived entries, preserving order.
func ActiveAuditEntrys(xs []AuditEntry) []AuditEntry {
	out := make([]AuditEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditEntrys adds up the offset of every active entry.
func SumAuditEntrys(xs []AuditEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexAuditEntrys keys the entries by ID, last write winning.
func IndexAuditEntrys(xs []AuditEntry) map[string]AuditEntry {
	out := make(map[string]AuditEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditEntry builds a valid AuditEntry, for tests and for the seed data the cron
// binary writes.
func SampleAuditEntry(id string, at time.Time) AuditEntry {
	return AuditEntry{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Category: "seeded by audit.SampleAuditEntry",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditItemSchema is the stored shape of a AuditItem.
var AuditItemSchema = schema.New("audit.audititem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditItem is one audit item.
type AuditItem struct {
	ID        string
	Category  string
	Label     string
	Reference string
	Version   int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x AuditItem) Key() string { return "audit/audititem/" + x.ID }

// Fields encodes x for storage.
func (x AuditItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"label":     x.Label,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditItem decodes a stored record.
func ParseAuditItem(f codec.Fields) (AuditItem, error) {
	var x AuditItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: audititem: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("audit: audititem %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: audititem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: audititem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: audititem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditItem) String() string {
	return fmt.Sprintf("AuditItem(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditItem) Normalize() AuditItem {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortAuditItems orders by version descending, then by ID.
func SortAuditItems(xs []AuditItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditItems drops the archived entries, preserving order.
func ActiveAuditItems(xs []AuditItem) []AuditItem {
	out := make([]AuditItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditItems adds up the version of every active entry.
func SumAuditItems(xs []AuditItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexAuditItems keys the entries by ID, last write winning.
func IndexAuditItems(xs []AuditItem) map[string]AuditItem {
	out := make(map[string]AuditItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditItem builds a valid AuditItem, for tests and for the seed data the cron
// binary writes.
func SampleAuditItem(id string, at time.Time) AuditItem {
	return AuditItem{
		ID:        id,
		Category:  "category-" + id,
		Label:     "label-" + id,
		Reference: "seeded by audit.SampleAuditItem",
		Version:   len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditDraftSchema is the stored shape of a AuditDraft.
var AuditDraftSchema = schema.New("audit.auditdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditDraft is one audit draft.
type AuditDraft struct {
	ID       string
	Source   string
	Name     string
	Comment  string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditDraft) Key() string { return "audit/auditdraft/" + x.ID }

// Fields encodes x for storage.
func (x AuditDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"comment":  x.Comment,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditDraft decodes a stored record.
func ParseAuditDraft(f codec.Fields) (AuditDraft, error) {
	var x AuditDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditdraft: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("audit: auditdraft %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("audit: auditdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditDraft) String() string {
	return fmt.Sprintf("AuditDraft(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditDraft) Normalize() AuditDraft {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortAuditDrafts orders by attempts descending, then by ID.
func SortAuditDrafts(xs []AuditDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditDrafts drops the archived entries, preserving order.
func ActiveAuditDrafts(xs []AuditDraft) []AuditDraft {
	out := make([]AuditDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditDrafts adds up the attempts of every active entry.
func SumAuditDrafts(xs []AuditDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexAuditDrafts keys the entries by ID, last write winning.
func IndexAuditDrafts(xs []AuditDraft) map[string]AuditDraft {
	out := make(map[string]AuditDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditDraft builds a valid AuditDraft, for tests and for the seed data the cron
// binary writes.
func SampleAuditDraft(id string, at time.Time) AuditDraft {
	return AuditDraft{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by audit.SampleAuditDraft",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditRevisionSchema is the stored shape of a AuditRevision.
var AuditRevisionSchema = schema.New("audit.auditrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditRevision is one audit revision.
type AuditRevision struct {
	ID       string
	Comment  string
	Source   string
	Group    string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditRevision) Key() string { return "audit/auditrevision/" + x.ID }

// Fields encodes x for storage.
func (x AuditRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditRevision decodes a stored record.
func ParseAuditRevision(f codec.Fields) (AuditRevision, error) {
	var x AuditRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditrevision: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("audit: auditrevision %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("audit: auditrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditRevision) String() string {
	return fmt.Sprintf("AuditRevision(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditRevision) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditRevision) Normalize() AuditRevision {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortAuditRevisions orders by attempts descending, then by ID.
func SortAuditRevisions(xs []AuditRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditRevisions drops the archived entries, preserving order.
func ActiveAuditRevisions(xs []AuditRevision) []AuditRevision {
	out := make([]AuditRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditRevisions adds up the attempts of every active entry.
func SumAuditRevisions(xs []AuditRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexAuditRevisions keys the entries by ID, last write winning.
func IndexAuditRevisions(xs []AuditRevision) map[string]AuditRevision {
	out := make(map[string]AuditRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditRevision builds a valid AuditRevision, for tests and for the seed data the cron
// binary writes.
func SampleAuditRevision(id string, at time.Time) AuditRevision {
	return AuditRevision{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Group:    "seeded by audit.SampleAuditRevision",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditSnapshotSchema is the stored shape of a AuditSnapshot.
var AuditSnapshotSchema = schema.New("audit.auditsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditSnapshot is one audit snapshot.
type AuditSnapshot struct {
	ID        string
	Reference string
	Group     string
	Label     string
	Weight    int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x AuditSnapshot) Key() string { return "audit/auditsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x AuditSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"label":     x.Label,
		"weight":    strconv.Itoa(x.Weight),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditSnapshot decodes a stored record.
func ParseAuditSnapshot(f codec.Fields) (AuditSnapshot, error) {
	var x AuditSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditsnapshot: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: auditsnapshot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("audit: auditsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditSnapshot) String() string {
	return fmt.Sprintf("AuditSnapshot(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditSnapshot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditSnapshot) Normalize() AuditSnapshot {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortAuditSnapshots orders by weight descending, then by ID.
func SortAuditSnapshots(xs []AuditSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditSnapshots drops the archived entries, preserving order.
func ActiveAuditSnapshots(xs []AuditSnapshot) []AuditSnapshot {
	out := make([]AuditSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditSnapshots adds up the weight of every active entry.
func SumAuditSnapshots(xs []AuditSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexAuditSnapshots keys the entries by ID, last write winning.
func IndexAuditSnapshots(xs []AuditSnapshot) map[string]AuditSnapshot {
	out := make(map[string]AuditSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditSnapshot builds a valid AuditSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleAuditSnapshot(id string, at time.Time) AuditSnapshot {
	return AuditSnapshot{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Label:     "seeded by audit.SampleAuditSnapshot",
		Weight:    len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditBatchSchema is the stored shape of a AuditBatch.
var AuditBatchSchema = schema.New("audit.auditbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditBatch is one audit batch.
type AuditBatch struct {
	ID       string
	Region   string
	Label    string
	Source   string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditBatch) Key() string { return "audit/auditbatch/" + x.ID }

// Fields encodes x for storage.
func (x AuditBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditBatch decodes a stored record.
func ParseAuditBatch(f codec.Fields) (AuditBatch, error) {
	var x AuditBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditbatch: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("audit: auditbatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("audit: auditbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditBatch) String() string {
	return fmt.Sprintf("AuditBatch(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditBatch) Normalize() AuditBatch {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortAuditBatchs orders by version descending, then by ID.
func SortAuditBatchs(xs []AuditBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditBatchs drops the archived entries, preserving order.
func ActiveAuditBatchs(xs []AuditBatch) []AuditBatch {
	out := make([]AuditBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditBatchs adds up the version of every active entry.
func SumAuditBatchs(xs []AuditBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexAuditBatchs keys the entries by ID, last write winning.
func IndexAuditBatchs(xs []AuditBatch) map[string]AuditBatch {
	out := make(map[string]AuditBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditBatch builds a valid AuditBatch, for tests and for the seed data the cron
// binary writes.
func SampleAuditBatch(id string, at time.Time) AuditBatch {
	return AuditBatch{
		ID:      id,
		Region:  "region-" + id,
		Label:   "label-" + id,
		Source:  "seeded by audit.SampleAuditBatch",
		Version: len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// AuditRequestSchema is the stored shape of a AuditRequest.
var AuditRequestSchema = schema.New("audit.auditrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditRequest is one audit request.
type AuditRequest struct {
	ID       string
	Comment  string
	Region   string
	Category string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditRequest) Key() string { return "audit/auditrequest/" + x.ID }

// Fields encodes x for storage.
func (x AuditRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"region":   x.Region,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditRequest decodes a stored record.
func ParseAuditRequest(f codec.Fields) (AuditRequest, error) {
	var x AuditRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditrequest: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("audit: auditrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("audit: auditrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditRequest) String() string {
	return fmt.Sprintf("AuditRequest(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditRequest) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditRequest) Normalize() AuditRequest {
	x.Comment = strings.TrimSpace(x.Comment)
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

// SortAuditRequests orders by version descending, then by ID.
func SortAuditRequests(xs []AuditRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditRequests drops the archived entries, preserving order.
func ActiveAuditRequests(xs []AuditRequest) []AuditRequest {
	out := make([]AuditRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditRequests adds up the version of every active entry.
func SumAuditRequests(xs []AuditRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexAuditRequests keys the entries by ID, last write winning.
func IndexAuditRequests(xs []AuditRequest) map[string]AuditRequest {
	out := make(map[string]AuditRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditRequest builds a valid AuditRequest, for tests and for the seed data the cron
// binary writes.
func SampleAuditRequest(id string, at time.Time) AuditRequest {
	return AuditRequest{
		ID:       id,
		Comment:  "comment-" + id,
		Region:   "region-" + id,
		Category: "seeded by audit.SampleAuditRequest",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditSummarySchema is the stored shape of a AuditSummary.
var AuditSummarySchema = schema.New("audit.auditsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditSummary is one audit summary.
type AuditSummary struct {
	ID       string
	Group    string
	Channel  string
	Source   string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditSummary) Key() string { return "audit/auditsummary/" + x.ID }

// Fields encodes x for storage.
func (x AuditSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditSummary decodes a stored record.
func ParseAuditSummary(f codec.Fields) (AuditSummary, error) {
	var x AuditSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditsummary: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("audit: auditsummary %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: auditsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditSummary) String() string {
	return fmt.Sprintf("AuditSummary(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditSummary) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditSummary) Normalize() AuditSummary {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortAuditSummarys orders by priority descending, then by ID.
func SortAuditSummarys(xs []AuditSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditSummarys drops the archived entries, preserving order.
func ActiveAuditSummarys(xs []AuditSummary) []AuditSummary {
	out := make([]AuditSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditSummarys adds up the priority of every active entry.
func SumAuditSummarys(xs []AuditSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexAuditSummarys keys the entries by ID, last write winning.
func IndexAuditSummarys(xs []AuditSummary) map[string]AuditSummary {
	out := make(map[string]AuditSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditSummary builds a valid AuditSummary, for tests and for the seed data the cron
// binary writes.
func SampleAuditSummary(id string, at time.Time) AuditSummary {
	return AuditSummary{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by audit.SampleAuditSummary",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditPolicySchema is the stored shape of a AuditPolicy.
var AuditPolicySchema = schema.New("audit.auditpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditPolicy is one audit policy.
type AuditPolicy struct {
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
func (x AuditPolicy) Key() string { return "audit/auditpolicy/" + x.ID }

// Fields encodes x for storage.
func (x AuditPolicy) Fields() codec.Fields {
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

// ParseAuditPolicy decodes a stored record.
func ParseAuditPolicy(f codec.Fields) (AuditPolicy, error) {
	var x AuditPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditpolicy: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("audit: auditpolicy %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: auditpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditPolicy) String() string {
	return fmt.Sprintf("AuditPolicy(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditPolicy) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditPolicy) Normalize() AuditPolicy {
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

// SortAuditPolicys orders by capacity descending, then by ID.
func SortAuditPolicys(xs []AuditPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditPolicys drops the archived entries, preserving order.
func ActiveAuditPolicys(xs []AuditPolicy) []AuditPolicy {
	out := make([]AuditPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditPolicys adds up the capacity of every active entry.
func SumAuditPolicys(xs []AuditPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexAuditPolicys keys the entries by ID, last write winning.
func IndexAuditPolicys(xs []AuditPolicy) map[string]AuditPolicy {
	out := make(map[string]AuditPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditPolicy builds a valid AuditPolicy, for tests and for the seed data the cron
// binary writes.
func SampleAuditPolicy(id string, at time.Time) AuditPolicy {
	return AuditPolicy{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Name:      "seeded by audit.SampleAuditPolicy",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditLinkSchema is the stored shape of a AuditLink.
var AuditLinkSchema = schema.New("audit.auditlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditLink is one audit link.
type AuditLink struct {
	ID       string
	Category string
	Region   string
	Source   string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditLink) Key() string { return "audit/auditlink/" + x.ID }

// Fields encodes x for storage.
func (x AuditLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditLink decodes a stored record.
func ParseAuditLink(f codec.Fields) (AuditLink, error) {
	var x AuditLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditlink: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("audit: auditlink %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("audit: auditlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditLink) String() string {
	return fmt.Sprintf("AuditLink(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditLink) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditLink) Normalize() AuditLink {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortAuditLinks orders by capacity descending, then by ID.
func SortAuditLinks(xs []AuditLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditLinks drops the archived entries, preserving order.
func ActiveAuditLinks(xs []AuditLink) []AuditLink {
	out := make([]AuditLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditLinks adds up the capacity of every active entry.
func SumAuditLinks(xs []AuditLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexAuditLinks keys the entries by ID, last write winning.
func IndexAuditLinks(xs []AuditLink) map[string]AuditLink {
	out := make(map[string]AuditLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditLink builds a valid AuditLink, for tests and for the seed data the cron
// binary writes.
func SampleAuditLink(id string, at time.Time) AuditLink {
	return AuditLink{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Source:   "seeded by audit.SampleAuditLink",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditNoteSchema is the stored shape of a AuditNote.
var AuditNoteSchema = schema.New("audit.auditnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditNote is one audit note.
type AuditNote struct {
	ID       string
	Label    string
	Comment  string
	Category string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditNote) Key() string { return "audit/auditnote/" + x.ID }

// Fields encodes x for storage.
func (x AuditNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"comment":  x.Comment,
		"category": x.Category,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditNote decodes a stored record.
func ParseAuditNote(f codec.Fields) (AuditNote, error) {
	var x AuditNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditnote: missing id")
	}
	x.Label = f["label"]
	x.Comment = f["comment"]
	x.Category = f["category"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("audit: auditnote %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: auditnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditNote) String() string {
	return fmt.Sprintf("AuditNote(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditNote) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditNote) Normalize() AuditNote {
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortAuditNotes orders by retained descending, then by ID.
func SortAuditNotes(xs []AuditNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditNotes drops the archived entries, preserving order.
func ActiveAuditNotes(xs []AuditNote) []AuditNote {
	out := make([]AuditNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditNotes adds up the retained of every active entry.
func SumAuditNotes(xs []AuditNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexAuditNotes keys the entries by ID, last write winning.
func IndexAuditNotes(xs []AuditNote) map[string]AuditNote {
	out := make(map[string]AuditNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditNote builds a valid AuditNote, for tests and for the seed data the cron
// binary writes.
func SampleAuditNote(id string, at time.Time) AuditNote {
	return AuditNote{
		ID:       id,
		Label:    "label-" + id,
		Comment:  "comment-" + id,
		Category: "seeded by audit.SampleAuditNote",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditWindowSchema is the stored shape of a AuditWindow.
var AuditWindowSchema = schema.New("audit.auditwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditWindow is one audit window.
type AuditWindow struct {
	ID       string
	Channel  string
	Category string
	Region   string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditWindow) Key() string { return "audit/auditwindow/" + x.ID }

// Fields encodes x for storage.
func (x AuditWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditWindow decodes a stored record.
func ParseAuditWindow(f codec.Fields) (AuditWindow, error) {
	var x AuditWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditwindow: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("audit: auditwindow %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("audit: auditwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditWindow) String() string {
	return fmt.Sprintf("AuditWindow(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditWindow) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditWindow) Normalize() AuditWindow {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortAuditWindows orders by capacity descending, then by ID.
func SortAuditWindows(xs []AuditWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditWindows drops the archived entries, preserving order.
func ActiveAuditWindows(xs []AuditWindow) []AuditWindow {
	out := make([]AuditWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditWindows adds up the capacity of every active entry.
func SumAuditWindows(xs []AuditWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexAuditWindows keys the entries by ID, last write winning.
func IndexAuditWindows(xs []AuditWindow) map[string]AuditWindow {
	out := make(map[string]AuditWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditWindow builds a valid AuditWindow, for tests and for the seed data the cron
// binary writes.
func SampleAuditWindow(id string, at time.Time) AuditWindow {
	return AuditWindow{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Region:   "seeded by audit.SampleAuditWindow",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditQuotaSchema is the stored shape of a AuditQuota.
var AuditQuotaSchema = schema.New("audit.auditquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditQuota is one audit quota.
type AuditQuota struct {
	ID       string
	Category string
	Comment  string
	Owner    string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditQuota) Key() string { return "audit/auditquota/" + x.ID }

// Fields encodes x for storage.
func (x AuditQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditQuota decodes a stored record.
func ParseAuditQuota(f codec.Fields) (AuditQuota, error) {
	var x AuditQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditquota: missing id")
	}
	x.Category = f["category"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: auditquota %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("audit: auditquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditQuota) String() string {
	return fmt.Sprintf("AuditQuota(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditQuota) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditQuota) Normalize() AuditQuota {
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortAuditQuotas orders by weight descending, then by ID.
func SortAuditQuotas(xs []AuditQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditQuotas drops the archived entries, preserving order.
func ActiveAuditQuotas(xs []AuditQuota) []AuditQuota {
	out := make([]AuditQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditQuotas adds up the weight of every active entry.
func SumAuditQuotas(xs []AuditQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexAuditQuotas keys the entries by ID, last write winning.
func IndexAuditQuotas(xs []AuditQuota) map[string]AuditQuota {
	out := make(map[string]AuditQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditQuota builds a valid AuditQuota, for tests and for the seed data the cron
// binary writes.
func SampleAuditQuota(id string, at time.Time) AuditQuota {
	return AuditQuota{
		ID:       id,
		Category: "category-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by audit.SampleAuditQuota",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditBucketSchema is the stored shape of a AuditBucket.
var AuditBucketSchema = schema.New("audit.auditbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditBucket is one audit bucket.
type AuditBucket struct {
	ID       string
	Label    string
	Group    string
	Owner    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditBucket) Key() string { return "audit/auditbucket/" + x.ID }

// Fields encodes x for storage.
func (x AuditBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"owner":    x.Owner,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditBucket decodes a stored record.
func ParseAuditBucket(f codec.Fields) (AuditBucket, error) {
	var x AuditBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditbucket: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("audit: auditbucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("audit: auditbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditBucket) String() string {
	return fmt.Sprintf("AuditBucket(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditBucket) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditBucket) Normalize() AuditBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortAuditBuckets orders by count descending, then by ID.
func SortAuditBuckets(xs []AuditBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditBuckets drops the archived entries, preserving order.
func ActiveAuditBuckets(xs []AuditBucket) []AuditBucket {
	out := make([]AuditBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditBuckets adds up the count of every active entry.
func SumAuditBuckets(xs []AuditBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexAuditBuckets keys the entries by ID, last write winning.
func IndexAuditBuckets(xs []AuditBucket) map[string]AuditBucket {
	out := make(map[string]AuditBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditBucket builds a valid AuditBucket, for tests and for the seed data the cron
// binary writes.
func SampleAuditBucket(id string, at time.Time) AuditBucket {
	return AuditBucket{
		ID:      id,
		Label:   "label-" + id,
		Group:   "group-" + id,
		Owner:   "seeded by audit.SampleAuditBucket",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// AuditLaneSchema is the stored shape of a AuditLane.
var AuditLaneSchema = schema.New("audit.auditlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditLane is one audit lane.
type AuditLane struct {
	ID        string
	Source    string
	Owner     string
	Reference string
	Attempts  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x AuditLane) Key() string { return "audit/auditlane/" + x.ID }

// Fields encodes x for storage.
func (x AuditLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"owner":     x.Owner,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditLane decodes a stored record.
func ParseAuditLane(f codec.Fields) (AuditLane, error) {
	var x AuditLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditlane: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("audit: auditlane %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("audit: auditlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditLane) String() string {
	return fmt.Sprintf("AuditLane(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditLane) Normalize() AuditLane {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortAuditLanes orders by attempts descending, then by ID.
func SortAuditLanes(xs []AuditLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditLanes drops the archived entries, preserving order.
func ActiveAuditLanes(xs []AuditLane) []AuditLane {
	out := make([]AuditLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditLanes adds up the attempts of every active entry.
func SumAuditLanes(xs []AuditLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexAuditLanes keys the entries by ID, last write winning.
func IndexAuditLanes(xs []AuditLane) map[string]AuditLane {
	out := make(map[string]AuditLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditLane builds a valid AuditLane, for tests and for the seed data the cron
// binary writes.
func SampleAuditLane(id string, at time.Time) AuditLane {
	return AuditLane{
		ID:        id,
		Source:    "source-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by audit.SampleAuditLane",
		Attempts:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditSlotSchema is the stored shape of a AuditSlot.
var AuditSlotSchema = schema.New("audit.auditslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditSlot is one audit slot.
type AuditSlot struct {
	ID        string
	Label     string
	Reference string
	Source    string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x AuditSlot) Key() string { return "audit/auditslot/" + x.ID }

// Fields encodes x for storage.
func (x AuditSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"reference": x.Reference,
		"source":    x.Source,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditSlot decodes a stored record.
func ParseAuditSlot(f codec.Fields) (AuditSlot, error) {
	var x AuditSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditslot: missing id")
	}
	x.Label = f["label"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("audit: auditslot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("audit: auditslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditSlot) String() string {
	return fmt.Sprintf("AuditSlot(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditSlot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditSlot) Normalize() AuditSlot {
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortAuditSlots orders by offset descending, then by ID.
func SortAuditSlots(xs []AuditSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditSlots drops the archived entries, preserving order.
func ActiveAuditSlots(xs []AuditSlot) []AuditSlot {
	out := make([]AuditSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditSlots adds up the offset of every active entry.
func SumAuditSlots(xs []AuditSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexAuditSlots keys the entries by ID, last write winning.
func IndexAuditSlots(xs []AuditSlot) map[string]AuditSlot {
	out := make(map[string]AuditSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditSlot builds a valid AuditSlot, for tests and for the seed data the cron
// binary writes.
func SampleAuditSlot(id string, at time.Time) AuditSlot {
	return AuditSlot{
		ID:        id,
		Label:     "label-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by audit.SampleAuditSlot",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditTagSchema is the stored shape of a AuditTag.
var AuditTagSchema = schema.New("audit.audittag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditTag is one audit tag.
type AuditTag struct {
	ID        string
	Name      string
	Source    string
	Reference string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x AuditTag) Key() string { return "audit/audittag/" + x.ID }

// Fields encodes x for storage.
func (x AuditTag) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"source":    x.Source,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditTag decodes a stored record.
func ParseAuditTag(f codec.Fields) (AuditTag, error) {
	var x AuditTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: audittag: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("audit: audittag %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("audit: audittag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: audittag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: audittag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditTag) String() string {
	return fmt.Sprintf("AuditTag(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditTag) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditTag) Normalize() AuditTag {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortAuditTags orders by offset descending, then by ID.
func SortAuditTags(xs []AuditTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditTags drops the archived entries, preserving order.
func ActiveAuditTags(xs []AuditTag) []AuditTag {
	out := make([]AuditTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditTags adds up the offset of every active entry.
func SumAuditTags(xs []AuditTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexAuditTags keys the entries by ID, last write winning.
func IndexAuditTags(xs []AuditTag) map[string]AuditTag {
	out := make(map[string]AuditTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditTag builds a valid AuditTag, for tests and for the seed data the cron
// binary writes.
func SampleAuditTag(id string, at time.Time) AuditTag {
	return AuditTag{
		ID:        id,
		Name:      "name-" + id,
		Source:    "source-" + id,
		Reference: "seeded by audit.SampleAuditTag",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditRuleSchema is the stored shape of a AuditRule.
var AuditRuleSchema = schema.New("audit.auditrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditRule is one audit rule.
type AuditRule struct {
	ID        string
	Name      string
	Reference string
	Source    string
	Offset    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x AuditRule) Key() string { return "audit/auditrule/" + x.ID }

// Fields encodes x for storage.
func (x AuditRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"source":    x.Source,
		"offset":    strconv.Itoa(x.Offset),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditRule decodes a stored record.
func ParseAuditRule(f codec.Fields) (AuditRule, error) {
	var x AuditRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditrule: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("audit: auditrule %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("audit: auditrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditRule) String() string {
	return fmt.Sprintf("AuditRule(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditRule) Normalize() AuditRule {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortAuditRules orders by offset descending, then by ID.
func SortAuditRules(xs []AuditRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditRules drops the archived entries, preserving order.
func ActiveAuditRules(xs []AuditRule) []AuditRule {
	out := make([]AuditRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditRules adds up the offset of every active entry.
func SumAuditRules(xs []AuditRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexAuditRules keys the entries by ID, last write winning.
func IndexAuditRules(xs []AuditRule) map[string]AuditRule {
	out := make(map[string]AuditRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditRule builds a valid AuditRule, for tests and for the seed data the cron
// binary writes.
func SampleAuditRule(id string, at time.Time) AuditRule {
	return AuditRule{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by audit.SampleAuditRule",
		Offset:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditHoldSchema is the stored shape of a AuditHold.
var AuditHoldSchema = schema.New("audit.audithold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditHold is one audit hold.
type AuditHold struct {
	ID        string
	Reference string
	Group     string
	Label     string
	Count     int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x AuditHold) Key() string { return "audit/audithold/" + x.ID }

// Fields encodes x for storage.
func (x AuditHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"label":     x.Label,
		"count":     strconv.Itoa(x.Count),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditHold decodes a stored record.
func ParseAuditHold(f codec.Fields) (AuditHold, error) {
	var x AuditHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: audithold: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("audit: audithold %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("audit: audithold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: audithold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: audithold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditHold) String() string {
	return fmt.Sprintf("AuditHold(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditHold) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditHold) Normalize() AuditHold {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortAuditHolds orders by count descending, then by ID.
func SortAuditHolds(xs []AuditHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditHolds drops the archived entries, preserving order.
func ActiveAuditHolds(xs []AuditHold) []AuditHold {
	out := make([]AuditHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditHolds adds up the count of every active entry.
func SumAuditHolds(xs []AuditHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexAuditHolds keys the entries by ID, last write winning.
func IndexAuditHolds(xs []AuditHold) map[string]AuditHold {
	out := make(map[string]AuditHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditHold builds a valid AuditHold, for tests and for the seed data the cron
// binary writes.
func SampleAuditHold(id string, at time.Time) AuditHold {
	return AuditHold{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Label:     "seeded by audit.SampleAuditHold",
		Count:     len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// AuditClaimSchema is the stored shape of a AuditClaim.
var AuditClaimSchema = schema.New("audit.auditclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditClaim is one audit claim.
type AuditClaim struct {
	ID       string
	Label    string
	Category string
	Name     string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditClaim) Key() string { return "audit/auditclaim/" + x.ID }

// Fields encodes x for storage.
func (x AuditClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditClaim decodes a stored record.
func ParseAuditClaim(f codec.Fields) (AuditClaim, error) {
	var x AuditClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditclaim: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("audit: auditclaim %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("audit: auditclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditClaim) String() string {
	return fmt.Sprintf("AuditClaim(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditClaim) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditClaim) Normalize() AuditClaim {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortAuditClaims orders by attempts descending, then by ID.
func SortAuditClaims(xs []AuditClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditClaims drops the archived entries, preserving order.
func ActiveAuditClaims(xs []AuditClaim) []AuditClaim {
	out := make([]AuditClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditClaims adds up the attempts of every active entry.
func SumAuditClaims(xs []AuditClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexAuditClaims keys the entries by ID, last write winning.
func IndexAuditClaims(xs []AuditClaim) map[string]AuditClaim {
	out := make(map[string]AuditClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditClaim builds a valid AuditClaim, for tests and for the seed data the cron
// binary writes.
func SampleAuditClaim(id string, at time.Time) AuditClaim {
	return AuditClaim{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Name:     "seeded by audit.SampleAuditClaim",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// AuditCursorSchema is the stored shape of a AuditCursor.
var AuditCursorSchema = schema.New("audit.auditcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// AuditCursor is one audit cursor.
type AuditCursor struct {
	ID       string
	Name     string
	Group    string
	Category string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x AuditCursor) Key() string { return "audit/auditcursor/" + x.ID }

// Fields encodes x for storage.
func (x AuditCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"group":    x.Group,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseAuditCursor decodes a stored record.
func ParseAuditCursor(f codec.Fields) (AuditCursor, error) {
	var x AuditCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("audit: auditcursor: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("audit: auditcursor %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("audit: auditcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("audit: auditcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("audit: auditcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x AuditCursor) String() string {
	return fmt.Sprintf("AuditCursor(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x AuditCursor) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x AuditCursor) Normalize() AuditCursor {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortAuditCursors orders by attempts descending, then by ID.
func SortAuditCursors(xs []AuditCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveAuditCursors drops the archived entries, preserving order.
func ActiveAuditCursors(xs []AuditCursor) []AuditCursor {
	out := make([]AuditCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumAuditCursors adds up the attempts of every active entry.
func SumAuditCursors(xs []AuditCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexAuditCursors keys the entries by ID, last write winning.
func IndexAuditCursors(xs []AuditCursor) map[string]AuditCursor {
	out := make(map[string]AuditCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleAuditCursor builds a valid AuditCursor, for tests and for the seed data the cron
// binary writes.
func SampleAuditCursor(id string, at time.Time) AuditCursor {
	return AuditCursor{
		ID:       id,
		Name:     "name-" + id,
		Group:    "group-" + id,
		Category: "seeded by audit.SampleAuditCursor",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	AuditRecordSchema,
	AuditEntrySchema,
	AuditItemSchema,
	AuditDraftSchema,
	AuditRevisionSchema,
	AuditSnapshotSchema,
	AuditBatchSchema,
	AuditRequestSchema,
	AuditSummarySchema,
	AuditPolicySchema,
	AuditLinkSchema,
	AuditNoteSchema,
	AuditWindowSchema,
	AuditQuotaSchema,
	AuditBucketSchema,
	AuditLaneSchema,
	AuditSlotSchema,
	AuditTagSchema,
	AuditRuleSchema,
	AuditHoldSchema,
	AuditClaimSchema,
	AuditCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"audit/auditrecord/",
	"audit/auditentry/",
	"audit/audititem/",
	"audit/auditdraft/",
	"audit/auditrevision/",
	"audit/auditsnapshot/",
	"audit/auditbatch/",
	"audit/auditrequest/",
	"audit/auditsummary/",
	"audit/auditpolicy/",
	"audit/auditlink/",
	"audit/auditnote/",
	"audit/auditwindow/",
	"audit/auditquota/",
	"audit/auditbucket/",
	"audit/auditlane/",
	"audit/auditslot/",
	"audit/audittag/",
	"audit/auditrule/",
	"audit/audithold/",
	"audit/auditclaim/",
	"audit/auditcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "audit: " + strings.Join(names, ", ")
}
