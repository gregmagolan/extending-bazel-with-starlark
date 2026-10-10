// Package report holds the report domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package report

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// ReportRecordSchema is the stored shape of a ReportRecord.
var ReportRecordSchema = schema.New("report.reportrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportRecord is one report record.
type ReportRecord struct {
	ID       string
	Comment  string
	Group    string
	Source   string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportRecord) Key() string { return "report/reportrecord/" + x.ID }

// Fields encodes x for storage.
func (x ReportRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportRecord decodes a stored record.
func ParseReportRecord(f codec.Fields) (ReportRecord, error) {
	var x ReportRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportrecord: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("report: reportrecord %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("report: reportrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportRecord) String() string {
	return fmt.Sprintf("ReportRecord(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportRecord) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportRecord) Normalize() ReportRecord {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortReportRecords orders by offset descending, then by ID.
func SortReportRecords(xs []ReportRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportRecords drops the archived entries, preserving order.
func ActiveReportRecords(xs []ReportRecord) []ReportRecord {
	out := make([]ReportRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportRecords adds up the offset of every active entry.
func SumReportRecords(xs []ReportRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexReportRecords keys the entries by ID, last write winning.
func IndexReportRecords(xs []ReportRecord) map[string]ReportRecord {
	out := make(map[string]ReportRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportRecord builds a valid ReportRecord, for tests and for the seed data the cron
// binary writes.
func SampleReportRecord(id string, at time.Time) ReportRecord {
	return ReportRecord{
		ID:       id,
		Comment:  "comment-" + id,
		Group:    "group-" + id,
		Source:   "seeded by report.SampleReportRecord",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportEntrySchema is the stored shape of a ReportEntry.
var ReportEntrySchema = schema.New("report.reportentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportEntry is one report entry.
type ReportEntry struct {
	ID       string
	Category string
	Group    string
	Region   string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportEntry) Key() string { return "report/reportentry/" + x.ID }

// Fields encodes x for storage.
func (x ReportEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"group":    x.Group,
		"region":   x.Region,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportEntry decodes a stored record.
func ParseReportEntry(f codec.Fields) (ReportEntry, error) {
	var x ReportEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportentry: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("report: reportentry %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("report: reportentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportEntry) String() string {
	return fmt.Sprintf("ReportEntry(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportEntry) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportEntry) Normalize() ReportEntry {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortReportEntrys orders by version descending, then by ID.
func SortReportEntrys(xs []ReportEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportEntrys drops the archived entries, preserving order.
func ActiveReportEntrys(xs []ReportEntry) []ReportEntry {
	out := make([]ReportEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportEntrys adds up the version of every active entry.
func SumReportEntrys(xs []ReportEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexReportEntrys keys the entries by ID, last write winning.
func IndexReportEntrys(xs []ReportEntry) map[string]ReportEntry {
	out := make(map[string]ReportEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportEntry builds a valid ReportEntry, for tests and for the seed data the cron
// binary writes.
func SampleReportEntry(id string, at time.Time) ReportEntry {
	return ReportEntry{
		ID:       id,
		Category: "category-" + id,
		Group:    "group-" + id,
		Region:   "seeded by report.SampleReportEntry",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportItemSchema is the stored shape of a ReportItem.
var ReportItemSchema = schema.New("report.reportitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportItem is one report item.
type ReportItem struct {
	ID       string
	Source   string
	Channel  string
	Owner    string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportItem) Key() string { return "report/reportitem/" + x.ID }

// Fields encodes x for storage.
func (x ReportItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportItem decodes a stored record.
func ParseReportItem(f codec.Fields) (ReportItem, error) {
	var x ReportItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportitem: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("report: reportitem %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("report: reportitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportItem) String() string {
	return fmt.Sprintf("ReportItem(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportItem) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportItem) Normalize() ReportItem {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortReportItems orders by attempts descending, then by ID.
func SortReportItems(xs []ReportItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportItems drops the archived entries, preserving order.
func ActiveReportItems(xs []ReportItem) []ReportItem {
	out := make([]ReportItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportItems adds up the attempts of every active entry.
func SumReportItems(xs []ReportItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexReportItems keys the entries by ID, last write winning.
func IndexReportItems(xs []ReportItem) map[string]ReportItem {
	out := make(map[string]ReportItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportItem builds a valid ReportItem, for tests and for the seed data the cron
// binary writes.
func SampleReportItem(id string, at time.Time) ReportItem {
	return ReportItem{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by report.SampleReportItem",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportDraftSchema is the stored shape of a ReportDraft.
var ReportDraftSchema = schema.New("report.reportdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportDraft is one report draft.
type ReportDraft struct {
	ID        string
	Label     string
	Region    string
	Reference string
	Offset    int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ReportDraft) Key() string { return "report/reportdraft/" + x.ID }

// Fields encodes x for storage.
func (x ReportDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"region":    x.Region,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportDraft decodes a stored record.
func ParseReportDraft(f codec.Fields) (ReportDraft, error) {
	var x ReportDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportdraft: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("report: reportdraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reportdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportDraft) String() string {
	return fmt.Sprintf("ReportDraft(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportDraft) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportDraft) Normalize() ReportDraft {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortReportDrafts orders by offset descending, then by ID.
func SortReportDrafts(xs []ReportDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportDrafts drops the archived entries, preserving order.
func ActiveReportDrafts(xs []ReportDraft) []ReportDraft {
	out := make([]ReportDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportDrafts adds up the offset of every active entry.
func SumReportDrafts(xs []ReportDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexReportDrafts keys the entries by ID, last write winning.
func IndexReportDrafts(xs []ReportDraft) map[string]ReportDraft {
	out := make(map[string]ReportDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportDraft builds a valid ReportDraft, for tests and for the seed data the cron
// binary writes.
func SampleReportDraft(id string, at time.Time) ReportDraft {
	return ReportDraft{
		ID:        id,
		Label:     "label-" + id,
		Region:    "region-" + id,
		Reference: "seeded by report.SampleReportDraft",
		Offset:    len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ReportRevisionSchema is the stored shape of a ReportRevision.
var ReportRevisionSchema = schema.New("report.reportrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportRevision is one report revision.
type ReportRevision struct {
	ID       string
	Comment  string
	Label    string
	Region   string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportRevision) Key() string { return "report/reportrevision/" + x.ID }

// Fields encodes x for storage.
func (x ReportRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"label":    x.Label,
		"region":   x.Region,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportRevision decodes a stored record.
func ParseReportRevision(f codec.Fields) (ReportRevision, error) {
	var x ReportRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportrevision: missing id")
	}
	x.Comment = f["comment"]
	x.Label = f["label"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("report: reportrevision %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("report: reportrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportRevision) String() string {
	return fmt.Sprintf("ReportRevision(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportRevision) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportRevision) Normalize() ReportRevision {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortReportRevisions orders by count descending, then by ID.
func SortReportRevisions(xs []ReportRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportRevisions drops the archived entries, preserving order.
func ActiveReportRevisions(xs []ReportRevision) []ReportRevision {
	out := make([]ReportRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportRevisions adds up the count of every active entry.
func SumReportRevisions(xs []ReportRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexReportRevisions keys the entries by ID, last write winning.
func IndexReportRevisions(xs []ReportRevision) map[string]ReportRevision {
	out := make(map[string]ReportRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportRevision builds a valid ReportRevision, for tests and for the seed data the cron
// binary writes.
func SampleReportRevision(id string, at time.Time) ReportRevision {
	return ReportRevision{
		ID:       id,
		Comment:  "comment-" + id,
		Label:    "label-" + id,
		Region:   "seeded by report.SampleReportRevision",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportSnapshotSchema is the stored shape of a ReportSnapshot.
var ReportSnapshotSchema = schema.New("report.reportsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportSnapshot is one report snapshot.
type ReportSnapshot struct {
	ID        string
	Name      string
	Reference string
	Category  string
	Retained  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ReportSnapshot) Key() string { return "report/reportsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x ReportSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"category":  x.Category,
		"retained":  strconv.Itoa(x.Retained),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportSnapshot decodes a stored record.
func ParseReportSnapshot(f codec.Fields) (ReportSnapshot, error) {
	var x ReportSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportsnapshot: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("report: reportsnapshot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("report: reportsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportSnapshot) String() string {
	return fmt.Sprintf("ReportSnapshot(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportSnapshot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportSnapshot) Normalize() ReportSnapshot {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortReportSnapshots orders by retained descending, then by ID.
func SortReportSnapshots(xs []ReportSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportSnapshots drops the archived entries, preserving order.
func ActiveReportSnapshots(xs []ReportSnapshot) []ReportSnapshot {
	out := make([]ReportSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportSnapshots adds up the retained of every active entry.
func SumReportSnapshots(xs []ReportSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexReportSnapshots keys the entries by ID, last write winning.
func IndexReportSnapshots(xs []ReportSnapshot) map[string]ReportSnapshot {
	out := make(map[string]ReportSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportSnapshot builds a valid ReportSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleReportSnapshot(id string, at time.Time) ReportSnapshot {
	return ReportSnapshot{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by report.SampleReportSnapshot",
		Retained:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ReportBatchSchema is the stored shape of a ReportBatch.
var ReportBatchSchema = schema.New("report.reportbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportBatch is one report batch.
type ReportBatch struct {
	ID        string
	Reference string
	Region    string
	Category  string
	Count     int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ReportBatch) Key() string { return "report/reportbatch/" + x.ID }

// Fields encodes x for storage.
func (x ReportBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"category":  x.Category,
		"count":     strconv.Itoa(x.Count),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportBatch decodes a stored record.
func ParseReportBatch(f codec.Fields) (ReportBatch, error) {
	var x ReportBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportbatch: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("report: reportbatch %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("report: reportbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportBatch) String() string {
	return fmt.Sprintf("ReportBatch(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportBatch) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportBatch) Normalize() ReportBatch {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortReportBatchs orders by count descending, then by ID.
func SortReportBatchs(xs []ReportBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportBatchs drops the archived entries, preserving order.
func ActiveReportBatchs(xs []ReportBatch) []ReportBatch {
	out := make([]ReportBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportBatchs adds up the count of every active entry.
func SumReportBatchs(xs []ReportBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexReportBatchs keys the entries by ID, last write winning.
func IndexReportBatchs(xs []ReportBatch) map[string]ReportBatch {
	out := make(map[string]ReportBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportBatch builds a valid ReportBatch, for tests and for the seed data the cron
// binary writes.
func SampleReportBatch(id string, at time.Time) ReportBatch {
	return ReportBatch{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Category:  "seeded by report.SampleReportBatch",
		Count:     len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ReportRequestSchema is the stored shape of a ReportRequest.
var ReportRequestSchema = schema.New("report.reportrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportRequest is one report request.
type ReportRequest struct {
	ID       string
	Channel  string
	Owner    string
	Label    string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportRequest) Key() string { return "report/reportrequest/" + x.ID }

// Fields encodes x for storage.
func (x ReportRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportRequest decodes a stored record.
func ParseReportRequest(f codec.Fields) (ReportRequest, error) {
	var x ReportRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportrequest: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("report: reportrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("report: reportrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportRequest) String() string {
	return fmt.Sprintf("ReportRequest(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportRequest) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportRequest) Normalize() ReportRequest {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortReportRequests orders by attempts descending, then by ID.
func SortReportRequests(xs []ReportRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportRequests drops the archived entries, preserving order.
func ActiveReportRequests(xs []ReportRequest) []ReportRequest {
	out := make([]ReportRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportRequests adds up the attempts of every active entry.
func SumReportRequests(xs []ReportRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexReportRequests keys the entries by ID, last write winning.
func IndexReportRequests(xs []ReportRequest) map[string]ReportRequest {
	out := make(map[string]ReportRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportRequest builds a valid ReportRequest, for tests and for the seed data the cron
// binary writes.
func SampleReportRequest(id string, at time.Time) ReportRequest {
	return ReportRequest{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by report.SampleReportRequest",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportSummarySchema is the stored shape of a ReportSummary.
var ReportSummarySchema = schema.New("report.reportsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportSummary is one report summary.
type ReportSummary struct {
	ID       string
	Owner    string
	Name     string
	Region   string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportSummary) Key() string { return "report/reportsummary/" + x.ID }

// Fields encodes x for storage.
func (x ReportSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"name":     x.Name,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportSummary decodes a stored record.
func ParseReportSummary(f codec.Fields) (ReportSummary, error) {
	var x ReportSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportsummary: missing id")
	}
	x.Owner = f["owner"]
	x.Name = f["name"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("report: reportsummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("report: reportsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportSummary) String() string {
	return fmt.Sprintf("ReportSummary(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportSummary) Normalize() ReportSummary {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortReportSummarys orders by capacity descending, then by ID.
func SortReportSummarys(xs []ReportSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportSummarys drops the archived entries, preserving order.
func ActiveReportSummarys(xs []ReportSummary) []ReportSummary {
	out := make([]ReportSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportSummarys adds up the capacity of every active entry.
func SumReportSummarys(xs []ReportSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexReportSummarys keys the entries by ID, last write winning.
func IndexReportSummarys(xs []ReportSummary) map[string]ReportSummary {
	out := make(map[string]ReportSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportSummary builds a valid ReportSummary, for tests and for the seed data the cron
// binary writes.
func SampleReportSummary(id string, at time.Time) ReportSummary {
	return ReportSummary{
		ID:       id,
		Owner:    "owner-" + id,
		Name:     "name-" + id,
		Region:   "seeded by report.SampleReportSummary",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportPolicySchema is the stored shape of a ReportPolicy.
var ReportPolicySchema = schema.New("report.reportpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportPolicy is one report policy.
type ReportPolicy struct {
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
func (x ReportPolicy) Key() string { return "report/reportpolicy/" + x.ID }

// Fields encodes x for storage.
func (x ReportPolicy) Fields() codec.Fields {
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

// ParseReportPolicy decodes a stored record.
func ParseReportPolicy(f codec.Fields) (ReportPolicy, error) {
	var x ReportPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportpolicy: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("report: reportpolicy %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reportpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportPolicy) String() string {
	return fmt.Sprintf("ReportPolicy(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportPolicy) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportPolicy) Normalize() ReportPolicy {
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

// SortReportPolicys orders by priority descending, then by ID.
func SortReportPolicys(xs []ReportPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportPolicys drops the archived entries, preserving order.
func ActiveReportPolicys(xs []ReportPolicy) []ReportPolicy {
	out := make([]ReportPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportPolicys adds up the priority of every active entry.
func SumReportPolicys(xs []ReportPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexReportPolicys keys the entries by ID, last write winning.
func IndexReportPolicys(xs []ReportPolicy) map[string]ReportPolicy {
	out := make(map[string]ReportPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportPolicy builds a valid ReportPolicy, for tests and for the seed data the cron
// binary writes.
func SampleReportPolicy(id string, at time.Time) ReportPolicy {
	return ReportPolicy{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by report.SampleReportPolicy",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportLinkSchema is the stored shape of a ReportLink.
var ReportLinkSchema = schema.New("report.reportlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportLink is one report link.
type ReportLink struct {
	ID       string
	Group    string
	Channel  string
	Source   string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportLink) Key() string { return "report/reportlink/" + x.ID }

// Fields encodes x for storage.
func (x ReportLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportLink decodes a stored record.
func ParseReportLink(f codec.Fields) (ReportLink, error) {
	var x ReportLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportlink: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("report: reportlink %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("report: reportlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportLink) String() string {
	return fmt.Sprintf("ReportLink(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportLink) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportLink) Normalize() ReportLink {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortReportLinks orders by offset descending, then by ID.
func SortReportLinks(xs []ReportLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportLinks drops the archived entries, preserving order.
func ActiveReportLinks(xs []ReportLink) []ReportLink {
	out := make([]ReportLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportLinks adds up the offset of every active entry.
func SumReportLinks(xs []ReportLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexReportLinks keys the entries by ID, last write winning.
func IndexReportLinks(xs []ReportLink) map[string]ReportLink {
	out := make(map[string]ReportLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportLink builds a valid ReportLink, for tests and for the seed data the cron
// binary writes.
func SampleReportLink(id string, at time.Time) ReportLink {
	return ReportLink{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by report.SampleReportLink",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportNoteSchema is the stored shape of a ReportNote.
var ReportNoteSchema = schema.New("report.reportnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportNote is one report note.
type ReportNote struct {
	ID       string
	Name     string
	Region   string
	Group    string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportNote) Key() string { return "report/reportnote/" + x.ID }

// Fields encodes x for storage.
func (x ReportNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"region":   x.Region,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportNote decodes a stored record.
func ParseReportNote(f codec.Fields) (ReportNote, error) {
	var x ReportNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportnote: missing id")
	}
	x.Name = f["name"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("report: reportnote %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("report: reportnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportNote) String() string {
	return fmt.Sprintf("ReportNote(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportNote) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportNote) Normalize() ReportNote {
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortReportNotes orders by count descending, then by ID.
func SortReportNotes(xs []ReportNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportNotes drops the archived entries, preserving order.
func ActiveReportNotes(xs []ReportNote) []ReportNote {
	out := make([]ReportNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportNotes adds up the count of every active entry.
func SumReportNotes(xs []ReportNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexReportNotes keys the entries by ID, last write winning.
func IndexReportNotes(xs []ReportNote) map[string]ReportNote {
	out := make(map[string]ReportNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportNote builds a valid ReportNote, for tests and for the seed data the cron
// binary writes.
func SampleReportNote(id string, at time.Time) ReportNote {
	return ReportNote{
		ID:       id,
		Name:     "name-" + id,
		Region:   "region-" + id,
		Group:    "seeded by report.SampleReportNote",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportWindowSchema is the stored shape of a ReportWindow.
var ReportWindowSchema = schema.New("report.reportwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportWindow is one report window.
type ReportWindow struct {
	ID        string
	Owner     string
	Reference string
	Group     string
	Attempts  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ReportWindow) Key() string { return "report/reportwindow/" + x.ID }

// Fields encodes x for storage.
func (x ReportWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"reference": x.Reference,
		"group":     x.Group,
		"attempts":  strconv.Itoa(x.Attempts),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportWindow decodes a stored record.
func ParseReportWindow(f codec.Fields) (ReportWindow, error) {
	var x ReportWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportwindow: missing id")
	}
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("report: reportwindow %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reportwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportWindow) String() string {
	return fmt.Sprintf("ReportWindow(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportWindow) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportWindow) Normalize() ReportWindow {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortReportWindows orders by attempts descending, then by ID.
func SortReportWindows(xs []ReportWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportWindows drops the archived entries, preserving order.
func ActiveReportWindows(xs []ReportWindow) []ReportWindow {
	out := make([]ReportWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportWindows adds up the attempts of every active entry.
func SumReportWindows(xs []ReportWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexReportWindows keys the entries by ID, last write winning.
func IndexReportWindows(xs []ReportWindow) map[string]ReportWindow {
	out := make(map[string]ReportWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportWindow builds a valid ReportWindow, for tests and for the seed data the cron
// binary writes.
func SampleReportWindow(id string, at time.Time) ReportWindow {
	return ReportWindow{
		ID:        id,
		Owner:     "owner-" + id,
		Reference: "reference-" + id,
		Group:     "seeded by report.SampleReportWindow",
		Attempts:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ReportQuotaSchema is the stored shape of a ReportQuota.
var ReportQuotaSchema = schema.New("report.reportquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportQuota is one report quota.
type ReportQuota struct {
	ID        string
	Comment   string
	Channel   string
	Reference string
	Count     int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ReportQuota) Key() string { return "report/reportquota/" + x.ID }

// Fields encodes x for storage.
func (x ReportQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"channel":   x.Channel,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportQuota decodes a stored record.
func ParseReportQuota(f codec.Fields) (ReportQuota, error) {
	var x ReportQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportquota: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("report: reportquota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reportquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportQuota) String() string {
	return fmt.Sprintf("ReportQuota(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportQuota) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportQuota) Normalize() ReportQuota {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortReportQuotas orders by count descending, then by ID.
func SortReportQuotas(xs []ReportQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportQuotas drops the archived entries, preserving order.
func ActiveReportQuotas(xs []ReportQuota) []ReportQuota {
	out := make([]ReportQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportQuotas adds up the count of every active entry.
func SumReportQuotas(xs []ReportQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexReportQuotas keys the entries by ID, last write winning.
func IndexReportQuotas(xs []ReportQuota) map[string]ReportQuota {
	out := make(map[string]ReportQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportQuota builds a valid ReportQuota, for tests and for the seed data the cron
// binary writes.
func SampleReportQuota(id string, at time.Time) ReportQuota {
	return ReportQuota{
		ID:        id,
		Comment:   "comment-" + id,
		Channel:   "channel-" + id,
		Reference: "seeded by report.SampleReportQuota",
		Count:     len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ReportBucketSchema is the stored shape of a ReportBucket.
var ReportBucketSchema = schema.New("report.reportbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportBucket is one report bucket.
type ReportBucket struct {
	ID       string
	Owner    string
	Comment  string
	Region   string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportBucket) Key() string { return "report/reportbucket/" + x.ID }

// Fields encodes x for storage.
func (x ReportBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportBucket decodes a stored record.
func ParseReportBucket(f codec.Fields) (ReportBucket, error) {
	var x ReportBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportbucket: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("report: reportbucket %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("report: reportbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportBucket) String() string {
	return fmt.Sprintf("ReportBucket(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportBucket) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportBucket) Normalize() ReportBucket {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortReportBuckets orders by capacity descending, then by ID.
func SortReportBuckets(xs []ReportBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportBuckets drops the archived entries, preserving order.
func ActiveReportBuckets(xs []ReportBucket) []ReportBucket {
	out := make([]ReportBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportBuckets adds up the capacity of every active entry.
func SumReportBuckets(xs []ReportBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexReportBuckets keys the entries by ID, last write winning.
func IndexReportBuckets(xs []ReportBucket) map[string]ReportBucket {
	out := make(map[string]ReportBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportBucket builds a valid ReportBucket, for tests and for the seed data the cron
// binary writes.
func SampleReportBucket(id string, at time.Time) ReportBucket {
	return ReportBucket{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by report.SampleReportBucket",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportLaneSchema is the stored shape of a ReportLane.
var ReportLaneSchema = schema.New("report.reportlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportLane is one report lane.
type ReportLane struct {
	ID       string
	Category string
	Region   string
	Owner    string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportLane) Key() string { return "report/reportlane/" + x.ID }

// Fields encodes x for storage.
func (x ReportLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportLane decodes a stored record.
func ParseReportLane(f codec.Fields) (ReportLane, error) {
	var x ReportLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportlane: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("report: reportlane %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("report: reportlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportLane) String() string {
	return fmt.Sprintf("ReportLane(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportLane) Normalize() ReportLane {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortReportLanes orders by priority descending, then by ID.
func SortReportLanes(xs []ReportLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportLanes drops the archived entries, preserving order.
func ActiveReportLanes(xs []ReportLane) []ReportLane {
	out := make([]ReportLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportLanes adds up the priority of every active entry.
func SumReportLanes(xs []ReportLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexReportLanes keys the entries by ID, last write winning.
func IndexReportLanes(xs []ReportLane) map[string]ReportLane {
	out := make(map[string]ReportLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportLane builds a valid ReportLane, for tests and for the seed data the cron
// binary writes.
func SampleReportLane(id string, at time.Time) ReportLane {
	return ReportLane{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Owner:    "seeded by report.SampleReportLane",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportSlotSchema is the stored shape of a ReportSlot.
var ReportSlotSchema = schema.New("report.reportslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportSlot is one report slot.
type ReportSlot struct {
	ID       string
	Comment  string
	Category string
	Name     string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportSlot) Key() string { return "report/reportslot/" + x.ID }

// Fields encodes x for storage.
func (x ReportSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"name":     x.Name,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportSlot decodes a stored record.
func ParseReportSlot(f codec.Fields) (ReportSlot, error) {
	var x ReportSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportslot: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("report: reportslot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("report: reportslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportSlot) String() string {
	return fmt.Sprintf("ReportSlot(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportSlot) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportSlot) Normalize() ReportSlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortReportSlots orders by version descending, then by ID.
func SortReportSlots(xs []ReportSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportSlots drops the archived entries, preserving order.
func ActiveReportSlots(xs []ReportSlot) []ReportSlot {
	out := make([]ReportSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportSlots adds up the version of every active entry.
func SumReportSlots(xs []ReportSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexReportSlots keys the entries by ID, last write winning.
func IndexReportSlots(xs []ReportSlot) map[string]ReportSlot {
	out := make(map[string]ReportSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportSlot builds a valid ReportSlot, for tests and for the seed data the cron
// binary writes.
func SampleReportSlot(id string, at time.Time) ReportSlot {
	return ReportSlot{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Name:     "seeded by report.SampleReportSlot",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportTagSchema is the stored shape of a ReportTag.
var ReportTagSchema = schema.New("report.reporttag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportTag is one report tag.
type ReportTag struct {
	ID       string
	Group    string
	Channel  string
	Comment  string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportTag) Key() string { return "report/reporttag/" + x.ID }

// Fields encodes x for storage.
func (x ReportTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportTag decodes a stored record.
func ParseReportTag(f codec.Fields) (ReportTag, error) {
	var x ReportTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reporttag: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("report: reporttag %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reporttag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reporttag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reporttag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportTag) String() string {
	return fmt.Sprintf("ReportTag(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportTag) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportTag) Normalize() ReportTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortReportTags orders by attempts descending, then by ID.
func SortReportTags(xs []ReportTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportTags drops the archived entries, preserving order.
func ActiveReportTags(xs []ReportTag) []ReportTag {
	out := make([]ReportTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportTags adds up the attempts of every active entry.
func SumReportTags(xs []ReportTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexReportTags keys the entries by ID, last write winning.
func IndexReportTags(xs []ReportTag) map[string]ReportTag {
	out := make(map[string]ReportTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportTag builds a valid ReportTag, for tests and for the seed data the cron
// binary writes.
func SampleReportTag(id string, at time.Time) ReportTag {
	return ReportTag{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by report.SampleReportTag",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportRuleSchema is the stored shape of a ReportRule.
var ReportRuleSchema = schema.New("report.reportrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportRule is one report rule.
type ReportRule struct {
	ID        string
	Group     string
	Reference string
	Channel   string
	Capacity  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ReportRule) Key() string { return "report/reportrule/" + x.ID }

// Fields encodes x for storage.
func (x ReportRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"channel":   x.Channel,
		"capacity":  strconv.Itoa(x.Capacity),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportRule decodes a stored record.
func ParseReportRule(f codec.Fields) (ReportRule, error) {
	var x ReportRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportrule: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("report: reportrule %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("report: reportrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportRule) String() string {
	return fmt.Sprintf("ReportRule(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportRule) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportRule) Normalize() ReportRule {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortReportRules orders by capacity descending, then by ID.
func SortReportRules(xs []ReportRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportRules drops the archived entries, preserving order.
func ActiveReportRules(xs []ReportRule) []ReportRule {
	out := make([]ReportRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportRules adds up the capacity of every active entry.
func SumReportRules(xs []ReportRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexReportRules keys the entries by ID, last write winning.
func IndexReportRules(xs []ReportRule) map[string]ReportRule {
	out := make(map[string]ReportRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportRule builds a valid ReportRule, for tests and for the seed data the cron
// binary writes.
func SampleReportRule(id string, at time.Time) ReportRule {
	return ReportRule{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by report.SampleReportRule",
		Capacity:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ReportHoldSchema is the stored shape of a ReportHold.
var ReportHoldSchema = schema.New("report.reporthold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportHold is one report hold.
type ReportHold struct {
	ID       string
	Region   string
	Source   string
	Category string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportHold) Key() string { return "report/reporthold/" + x.ID }

// Fields encodes x for storage.
func (x ReportHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"source":   x.Source,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportHold decodes a stored record.
func ParseReportHold(f codec.Fields) (ReportHold, error) {
	var x ReportHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reporthold: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("report: reporthold %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reporthold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reporthold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reporthold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportHold) String() string {
	return fmt.Sprintf("ReportHold(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportHold) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportHold) Normalize() ReportHold {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortReportHolds orders by count descending, then by ID.
func SortReportHolds(xs []ReportHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportHolds drops the archived entries, preserving order.
func ActiveReportHolds(xs []ReportHold) []ReportHold {
	out := make([]ReportHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportHolds adds up the count of every active entry.
func SumReportHolds(xs []ReportHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexReportHolds keys the entries by ID, last write winning.
func IndexReportHolds(xs []ReportHold) map[string]ReportHold {
	out := make(map[string]ReportHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportHold builds a valid ReportHold, for tests and for the seed data the cron
// binary writes.
func SampleReportHold(id string, at time.Time) ReportHold {
	return ReportHold{
		ID:       id,
		Region:   "region-" + id,
		Source:   "source-" + id,
		Category: "seeded by report.SampleReportHold",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ReportClaimSchema is the stored shape of a ReportClaim.
var ReportClaimSchema = schema.New("report.reportclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportClaim is one report claim.
type ReportClaim struct {
	ID        string
	Name      string
	Reference string
	Region    string
	Weight    int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ReportClaim) Key() string { return "report/reportclaim/" + x.ID }

// Fields encodes x for storage.
func (x ReportClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"region":    x.Region,
		"weight":    strconv.Itoa(x.Weight),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportClaim decodes a stored record.
func ParseReportClaim(f codec.Fields) (ReportClaim, error) {
	var x ReportClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportclaim: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reportclaim %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("report: reportclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportClaim) String() string {
	return fmt.Sprintf("ReportClaim(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportClaim) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportClaim) Normalize() ReportClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortReportClaims orders by weight descending, then by ID.
func SortReportClaims(xs []ReportClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportClaims drops the archived entries, preserving order.
func ActiveReportClaims(xs []ReportClaim) []ReportClaim {
	out := make([]ReportClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportClaims adds up the weight of every active entry.
func SumReportClaims(xs []ReportClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexReportClaims keys the entries by ID, last write winning.
func IndexReportClaims(xs []ReportClaim) map[string]ReportClaim {
	out := make(map[string]ReportClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportClaim builds a valid ReportClaim, for tests and for the seed data the cron
// binary writes.
func SampleReportClaim(id string, at time.Time) ReportClaim {
	return ReportClaim{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by report.SampleReportClaim",
		Weight:    len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ReportCursorSchema is the stored shape of a ReportCursor.
var ReportCursorSchema = schema.New("report.reportcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ReportCursor is one report cursor.
type ReportCursor struct {
	ID       string
	Name     string
	Group    string
	Category string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ReportCursor) Key() string { return "report/reportcursor/" + x.ID }

// Fields encodes x for storage.
func (x ReportCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"group":    x.Group,
		"category": x.Category,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseReportCursor decodes a stored record.
func ParseReportCursor(f codec.Fields) (ReportCursor, error) {
	var x ReportCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("report: reportcursor: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("report: reportcursor %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("report: reportcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("report: reportcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("report: reportcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ReportCursor) String() string {
	return fmt.Sprintf("ReportCursor(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ReportCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ReportCursor) Normalize() ReportCursor {
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortReportCursors orders by weight descending, then by ID.
func SortReportCursors(xs []ReportCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveReportCursors drops the archived entries, preserving order.
func ActiveReportCursors(xs []ReportCursor) []ReportCursor {
	out := make([]ReportCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumReportCursors adds up the weight of every active entry.
func SumReportCursors(xs []ReportCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexReportCursors keys the entries by ID, last write winning.
func IndexReportCursors(xs []ReportCursor) map[string]ReportCursor {
	out := make(map[string]ReportCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleReportCursor builds a valid ReportCursor, for tests and for the seed data the cron
// binary writes.
func SampleReportCursor(id string, at time.Time) ReportCursor {
	return ReportCursor{
		ID:       id,
		Name:     "name-" + id,
		Group:    "group-" + id,
		Category: "seeded by report.SampleReportCursor",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	ReportRecordSchema,
	ReportEntrySchema,
	ReportItemSchema,
	ReportDraftSchema,
	ReportRevisionSchema,
	ReportSnapshotSchema,
	ReportBatchSchema,
	ReportRequestSchema,
	ReportSummarySchema,
	ReportPolicySchema,
	ReportLinkSchema,
	ReportNoteSchema,
	ReportWindowSchema,
	ReportQuotaSchema,
	ReportBucketSchema,
	ReportLaneSchema,
	ReportSlotSchema,
	ReportTagSchema,
	ReportRuleSchema,
	ReportHoldSchema,
	ReportClaimSchema,
	ReportCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"report/reportrecord/",
	"report/reportentry/",
	"report/reportitem/",
	"report/reportdraft/",
	"report/reportrevision/",
	"report/reportsnapshot/",
	"report/reportbatch/",
	"report/reportrequest/",
	"report/reportsummary/",
	"report/reportpolicy/",
	"report/reportlink/",
	"report/reportnote/",
	"report/reportwindow/",
	"report/reportquota/",
	"report/reportbucket/",
	"report/reportlane/",
	"report/reportslot/",
	"report/reporttag/",
	"report/reportrule/",
	"report/reporthold/",
	"report/reportclaim/",
	"report/reportcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "report: " + strings.Join(names, ", ")
}
