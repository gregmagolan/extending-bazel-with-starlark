// Package settlement holds the settlement domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package settlement

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// SettlementRecordSchema is the stored shape of a SettlementRecord.
var SettlementRecordSchema = schema.New("settlement.settlementrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementRecord is one settlement record.
type SettlementRecord struct {
	ID       string
	Name     string
	Owner    string
	Channel  string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementRecord) Key() string { return "settlement/settlementrecord/" + x.ID }

// Fields encodes x for storage.
func (x SettlementRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementRecord decodes a stored record.
func ParseSettlementRecord(f codec.Fields) (SettlementRecord, error) {
	var x SettlementRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementrecord: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrecord %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementRecord) String() string {
	return fmt.Sprintf("SettlementRecord(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementRecord) Normalize() SettlementRecord {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSettlementRecords orders by capacity descending, then by ID.
func SortSettlementRecords(xs []SettlementRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementRecords drops the archived entries, preserving order.
func ActiveSettlementRecords(xs []SettlementRecord) []SettlementRecord {
	out := make([]SettlementRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementRecords adds up the capacity of every active entry.
func SumSettlementRecords(xs []SettlementRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSettlementRecords keys the entries by ID, last write winning.
func IndexSettlementRecords(xs []SettlementRecord) map[string]SettlementRecord {
	out := make(map[string]SettlementRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementRecord builds a valid SettlementRecord, for tests and for the seed data the cron
// binary writes.
func SampleSettlementRecord(id string, at time.Time) SettlementRecord {
	return SettlementRecord{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by settlement.SampleSettlementRecord",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementEntrySchema is the stored shape of a SettlementEntry.
var SettlementEntrySchema = schema.New("settlement.settlemententry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementEntry is one settlement entry.
type SettlementEntry struct {
	ID       string
	Name     string
	Owner    string
	Comment  string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementEntry) Key() string { return "settlement/settlemententry/" + x.ID }

// Fields encodes x for storage.
func (x SettlementEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementEntry decodes a stored record.
func ParseSettlementEntry(f codec.Fields) (SettlementEntry, error) {
	var x SettlementEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlemententry: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlemententry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("settlement: settlemententry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlemententry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlemententry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementEntry) String() string {
	return fmt.Sprintf("SettlementEntry(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementEntry) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementEntry) Normalize() SettlementEntry {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSettlementEntrys orders by count descending, then by ID.
func SortSettlementEntrys(xs []SettlementEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementEntrys drops the archived entries, preserving order.
func ActiveSettlementEntrys(xs []SettlementEntry) []SettlementEntry {
	out := make([]SettlementEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementEntrys adds up the count of every active entry.
func SumSettlementEntrys(xs []SettlementEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSettlementEntrys keys the entries by ID, last write winning.
func IndexSettlementEntrys(xs []SettlementEntry) map[string]SettlementEntry {
	out := make(map[string]SettlementEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementEntry builds a valid SettlementEntry, for tests and for the seed data the cron
// binary writes.
func SampleSettlementEntry(id string, at time.Time) SettlementEntry {
	return SettlementEntry{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Comment:  "seeded by settlement.SampleSettlementEntry",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementItemSchema is the stored shape of a SettlementItem.
var SettlementItemSchema = schema.New("settlement.settlementitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementItem is one settlement item.
type SettlementItem struct {
	ID       string
	Label    string
	Group    string
	Channel  string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementItem) Key() string { return "settlement/settlementitem/" + x.ID }

// Fields encodes x for storage.
func (x SettlementItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementItem decodes a stored record.
func ParseSettlementItem(f codec.Fields) (SettlementItem, error) {
	var x SettlementItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementitem: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementItem) String() string {
	return fmt.Sprintf("SettlementItem(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementItem) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementItem) Normalize() SettlementItem {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortSettlementItems orders by priority descending, then by ID.
func SortSettlementItems(xs []SettlementItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementItems drops the archived entries, preserving order.
func ActiveSettlementItems(xs []SettlementItem) []SettlementItem {
	out := make([]SettlementItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementItems adds up the priority of every active entry.
func SumSettlementItems(xs []SettlementItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSettlementItems keys the entries by ID, last write winning.
func IndexSettlementItems(xs []SettlementItem) map[string]SettlementItem {
	out := make(map[string]SettlementItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementItem builds a valid SettlementItem, for tests and for the seed data the cron
// binary writes.
func SampleSettlementItem(id string, at time.Time) SettlementItem {
	return SettlementItem{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Channel:  "seeded by settlement.SampleSettlementItem",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementDraftSchema is the stored shape of a SettlementDraft.
var SettlementDraftSchema = schema.New("settlement.settlementdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementDraft is one settlement draft.
type SettlementDraft struct {
	ID       string
	Owner    string
	Label    string
	Source   string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementDraft) Key() string { return "settlement/settlementdraft/" + x.ID }

// Fields encodes x for storage.
func (x SettlementDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"label":    x.Label,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementDraft decodes a stored record.
func ParseSettlementDraft(f codec.Fields) (SettlementDraft, error) {
	var x SettlementDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementdraft: missing id")
	}
	x.Owner = f["owner"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementdraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementDraft) String() string {
	return fmt.Sprintf("SettlementDraft(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementDraft) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementDraft) Normalize() SettlementDraft {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSettlementDrafts orders by retained descending, then by ID.
func SortSettlementDrafts(xs []SettlementDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementDrafts drops the archived entries, preserving order.
func ActiveSettlementDrafts(xs []SettlementDraft) []SettlementDraft {
	out := make([]SettlementDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementDrafts adds up the retained of every active entry.
func SumSettlementDrafts(xs []SettlementDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexSettlementDrafts keys the entries by ID, last write winning.
func IndexSettlementDrafts(xs []SettlementDraft) map[string]SettlementDraft {
	out := make(map[string]SettlementDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementDraft builds a valid SettlementDraft, for tests and for the seed data the cron
// binary writes.
func SampleSettlementDraft(id string, at time.Time) SettlementDraft {
	return SettlementDraft{
		ID:       id,
		Owner:    "owner-" + id,
		Label:    "label-" + id,
		Source:   "seeded by settlement.SampleSettlementDraft",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementRevisionSchema is the stored shape of a SettlementRevision.
var SettlementRevisionSchema = schema.New("settlement.settlementrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementRevision is one settlement revision.
type SettlementRevision struct {
	ID       string
	Channel  string
	Region   string
	Source   string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementRevision) Key() string { return "settlement/settlementrevision/" + x.ID }

// Fields encodes x for storage.
func (x SettlementRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementRevision decodes a stored record.
func ParseSettlementRevision(f codec.Fields) (SettlementRevision, error) {
	var x SettlementRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementrevision: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementRevision) String() string {
	return fmt.Sprintf("SettlementRevision(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementRevision) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementRevision) Normalize() SettlementRevision {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSettlementRevisions orders by version descending, then by ID.
func SortSettlementRevisions(xs []SettlementRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementRevisions drops the archived entries, preserving order.
func ActiveSettlementRevisions(xs []SettlementRevision) []SettlementRevision {
	out := make([]SettlementRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementRevisions adds up the version of every active entry.
func SumSettlementRevisions(xs []SettlementRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSettlementRevisions keys the entries by ID, last write winning.
func IndexSettlementRevisions(xs []SettlementRevision) map[string]SettlementRevision {
	out := make(map[string]SettlementRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementRevision builds a valid SettlementRevision, for tests and for the seed data the cron
// binary writes.
func SampleSettlementRevision(id string, at time.Time) SettlementRevision {
	return SettlementRevision{
		ID:      id,
		Channel: "channel-" + id,
		Region:  "region-" + id,
		Source:  "seeded by settlement.SampleSettlementRevision",
		Version: len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SettlementSnapshotSchema is the stored shape of a SettlementSnapshot.
var SettlementSnapshotSchema = schema.New("settlement.settlementsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementSnapshot is one settlement snapshot.
type SettlementSnapshot struct {
	ID       string
	Source   string
	Label    string
	Group    string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementSnapshot) Key() string { return "settlement/settlementsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x SettlementSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"label":    x.Label,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementSnapshot decodes a stored record.
func ParseSettlementSnapshot(f codec.Fields) (SettlementSnapshot, error) {
	var x SettlementSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementsnapshot: missing id")
	}
	x.Source = f["source"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementsnapshot %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementSnapshot) String() string {
	return fmt.Sprintf("SettlementSnapshot(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementSnapshot) Normalize() SettlementSnapshot {
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortSettlementSnapshots orders by priority descending, then by ID.
func SortSettlementSnapshots(xs []SettlementSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementSnapshots drops the archived entries, preserving order.
func ActiveSettlementSnapshots(xs []SettlementSnapshot) []SettlementSnapshot {
	out := make([]SettlementSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementSnapshots adds up the priority of every active entry.
func SumSettlementSnapshots(xs []SettlementSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexSettlementSnapshots keys the entries by ID, last write winning.
func IndexSettlementSnapshots(xs []SettlementSnapshot) map[string]SettlementSnapshot {
	out := make(map[string]SettlementSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementSnapshot builds a valid SettlementSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleSettlementSnapshot(id string, at time.Time) SettlementSnapshot {
	return SettlementSnapshot{
		ID:       id,
		Source:   "source-" + id,
		Label:    "label-" + id,
		Group:    "seeded by settlement.SampleSettlementSnapshot",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementBatchSchema is the stored shape of a SettlementBatch.
var SettlementBatchSchema = schema.New("settlement.settlementbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementBatch is one settlement batch.
type SettlementBatch struct {
	ID       string
	Channel  string
	Name     string
	Label    string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementBatch) Key() string { return "settlement/settlementbatch/" + x.ID }

// Fields encodes x for storage.
func (x SettlementBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"name":     x.Name,
		"label":    x.Label,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementBatch decodes a stored record.
func ParseSettlementBatch(f codec.Fields) (SettlementBatch, error) {
	var x SettlementBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementbatch: missing id")
	}
	x.Channel = f["channel"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementbatch %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementBatch) String() string {
	return fmt.Sprintf("SettlementBatch(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementBatch) Normalize() SettlementBatch {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSettlementBatchs orders by version descending, then by ID.
func SortSettlementBatchs(xs []SettlementBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementBatchs drops the archived entries, preserving order.
func ActiveSettlementBatchs(xs []SettlementBatch) []SettlementBatch {
	out := make([]SettlementBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementBatchs adds up the version of every active entry.
func SumSettlementBatchs(xs []SettlementBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSettlementBatchs keys the entries by ID, last write winning.
func IndexSettlementBatchs(xs []SettlementBatch) map[string]SettlementBatch {
	out := make(map[string]SettlementBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementBatch builds a valid SettlementBatch, for tests and for the seed data the cron
// binary writes.
func SampleSettlementBatch(id string, at time.Time) SettlementBatch {
	return SettlementBatch{
		ID:       id,
		Channel:  "channel-" + id,
		Name:     "name-" + id,
		Label:    "seeded by settlement.SampleSettlementBatch",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementRequestSchema is the stored shape of a SettlementRequest.
var SettlementRequestSchema = schema.New("settlement.settlementrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementRequest is one settlement request.
type SettlementRequest struct {
	ID       string
	Comment  string
	Label    string
	Owner    string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementRequest) Key() string { return "settlement/settlementrequest/" + x.ID }

// Fields encodes x for storage.
func (x SettlementRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"label":    x.Label,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementRequest decodes a stored record.
func ParseSettlementRequest(f codec.Fields) (SettlementRequest, error) {
	var x SettlementRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementrequest: missing id")
	}
	x.Comment = f["comment"]
	x.Label = f["label"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrequest %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementRequest) String() string {
	return fmt.Sprintf("SettlementRequest(%s, comment=%q, attempts=%d, archived=%t)", x.ID, x.Comment, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementRequest) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementRequest) Normalize() SettlementRequest {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSettlementRequests orders by attempts descending, then by ID.
func SortSettlementRequests(xs []SettlementRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementRequests drops the archived entries, preserving order.
func ActiveSettlementRequests(xs []SettlementRequest) []SettlementRequest {
	out := make([]SettlementRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementRequests adds up the attempts of every active entry.
func SumSettlementRequests(xs []SettlementRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexSettlementRequests keys the entries by ID, last write winning.
func IndexSettlementRequests(xs []SettlementRequest) map[string]SettlementRequest {
	out := make(map[string]SettlementRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementRequest builds a valid SettlementRequest, for tests and for the seed data the cron
// binary writes.
func SampleSettlementRequest(id string, at time.Time) SettlementRequest {
	return SettlementRequest{
		ID:       id,
		Comment:  "comment-" + id,
		Label:    "label-" + id,
		Owner:    "seeded by settlement.SampleSettlementRequest",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementSummarySchema is the stored shape of a SettlementSummary.
var SettlementSummarySchema = schema.New("settlement.settlementsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementSummary is one settlement summary.
type SettlementSummary struct {
	ID       string
	Category string
	Comment  string
	Owner    string
	Version  int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementSummary) Key() string { return "settlement/settlementsummary/" + x.ID }

// Fields encodes x for storage.
func (x SettlementSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"version":  strconv.Itoa(x.Version),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementSummary decodes a stored record.
func ParseSettlementSummary(f codec.Fields) (SettlementSummary, error) {
	var x SettlementSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementsummary: missing id")
	}
	x.Category = f["category"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementsummary %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementSummary) String() string {
	return fmt.Sprintf("SettlementSummary(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementSummary) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementSummary) Normalize() SettlementSummary {
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortSettlementSummarys orders by version descending, then by ID.
func SortSettlementSummarys(xs []SettlementSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementSummarys drops the archived entries, preserving order.
func ActiveSettlementSummarys(xs []SettlementSummary) []SettlementSummary {
	out := make([]SettlementSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementSummarys adds up the version of every active entry.
func SumSettlementSummarys(xs []SettlementSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSettlementSummarys keys the entries by ID, last write winning.
func IndexSettlementSummarys(xs []SettlementSummary) map[string]SettlementSummary {
	out := make(map[string]SettlementSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementSummary builds a valid SettlementSummary, for tests and for the seed data the cron
// binary writes.
func SampleSettlementSummary(id string, at time.Time) SettlementSummary {
	return SettlementSummary{
		ID:       id,
		Category: "category-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by settlement.SampleSettlementSummary",
		Version:  len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementPolicySchema is the stored shape of a SettlementPolicy.
var SettlementPolicySchema = schema.New("settlement.settlementpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementPolicy is one settlement policy.
type SettlementPolicy struct {
	ID       string
	Category string
	Source   string
	Comment  string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementPolicy) Key() string { return "settlement/settlementpolicy/" + x.ID }

// Fields encodes x for storage.
func (x SettlementPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"source":   x.Source,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementPolicy decodes a stored record.
func ParseSettlementPolicy(f codec.Fields) (SettlementPolicy, error) {
	var x SettlementPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementpolicy: missing id")
	}
	x.Category = f["category"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementpolicy %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementPolicy) String() string {
	return fmt.Sprintf("SettlementPolicy(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementPolicy) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementPolicy) Normalize() SettlementPolicy {
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortSettlementPolicys orders by retained descending, then by ID.
func SortSettlementPolicys(xs []SettlementPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementPolicys drops the archived entries, preserving order.
func ActiveSettlementPolicys(xs []SettlementPolicy) []SettlementPolicy {
	out := make([]SettlementPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementPolicys adds up the retained of every active entry.
func SumSettlementPolicys(xs []SettlementPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexSettlementPolicys keys the entries by ID, last write winning.
func IndexSettlementPolicys(xs []SettlementPolicy) map[string]SettlementPolicy {
	out := make(map[string]SettlementPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementPolicy builds a valid SettlementPolicy, for tests and for the seed data the cron
// binary writes.
func SampleSettlementPolicy(id string, at time.Time) SettlementPolicy {
	return SettlementPolicy{
		ID:       id,
		Category: "category-" + id,
		Source:   "source-" + id,
		Comment:  "seeded by settlement.SampleSettlementPolicy",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementLinkSchema is the stored shape of a SettlementLink.
var SettlementLinkSchema = schema.New("settlement.settlementlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementLink is one settlement link.
type SettlementLink struct {
	ID       string
	Category string
	Channel  string
	Region   string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementLink) Key() string { return "settlement/settlementlink/" + x.ID }

// Fields encodes x for storage.
func (x SettlementLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementLink decodes a stored record.
func ParseSettlementLink(f codec.Fields) (SettlementLink, error) {
	var x SettlementLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementlink: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementlink %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementLink) String() string {
	return fmt.Sprintf("SettlementLink(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementLink) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementLink) Normalize() SettlementLink {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSettlementLinks orders by offset descending, then by ID.
func SortSettlementLinks(xs []SettlementLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementLinks drops the archived entries, preserving order.
func ActiveSettlementLinks(xs []SettlementLink) []SettlementLink {
	out := make([]SettlementLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementLinks adds up the offset of every active entry.
func SumSettlementLinks(xs []SettlementLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSettlementLinks keys the entries by ID, last write winning.
func IndexSettlementLinks(xs []SettlementLink) map[string]SettlementLink {
	out := make(map[string]SettlementLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementLink builds a valid SettlementLink, for tests and for the seed data the cron
// binary writes.
func SampleSettlementLink(id string, at time.Time) SettlementLink {
	return SettlementLink{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by settlement.SampleSettlementLink",
		Offset:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementNoteSchema is the stored shape of a SettlementNote.
var SettlementNoteSchema = schema.New("settlement.settlementnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementNote is one settlement note.
type SettlementNote struct {
	ID       string
	Comment  string
	Owner    string
	Category string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementNote) Key() string { return "settlement/settlementnote/" + x.ID }

// Fields encodes x for storage.
func (x SettlementNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementNote decodes a stored record.
func ParseSettlementNote(f codec.Fields) (SettlementNote, error) {
	var x SettlementNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementnote: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementnote %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementNote) String() string {
	return fmt.Sprintf("SettlementNote(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementNote) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementNote) Normalize() SettlementNote {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSettlementNotes orders by count descending, then by ID.
func SortSettlementNotes(xs []SettlementNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementNotes drops the archived entries, preserving order.
func ActiveSettlementNotes(xs []SettlementNote) []SettlementNote {
	out := make([]SettlementNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementNotes adds up the count of every active entry.
func SumSettlementNotes(xs []SettlementNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSettlementNotes keys the entries by ID, last write winning.
func IndexSettlementNotes(xs []SettlementNote) map[string]SettlementNote {
	out := make(map[string]SettlementNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementNote builds a valid SettlementNote, for tests and for the seed data the cron
// binary writes.
func SampleSettlementNote(id string, at time.Time) SettlementNote {
	return SettlementNote{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by settlement.SampleSettlementNote",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementWindowSchema is the stored shape of a SettlementWindow.
var SettlementWindowSchema = schema.New("settlement.settlementwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementWindow is one settlement window.
type SettlementWindow struct {
	ID        string
	Reference string
	Category  string
	Label     string
	Attempts  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SettlementWindow) Key() string { return "settlement/settlementwindow/" + x.ID }

// Fields encodes x for storage.
func (x SettlementWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"category":  x.Category,
		"label":     x.Label,
		"attempts":  strconv.Itoa(x.Attempts),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementWindow decodes a stored record.
func ParseSettlementWindow(f codec.Fields) (SettlementWindow, error) {
	var x SettlementWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementwindow: missing id")
	}
	x.Reference = f["reference"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementwindow %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementWindow) String() string {
	return fmt.Sprintf("SettlementWindow(%s, reference=%q, attempts=%d, archived=%t)", x.ID, x.Reference, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementWindow) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementWindow) Normalize() SettlementWindow {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortSettlementWindows orders by attempts descending, then by ID.
func SortSettlementWindows(xs []SettlementWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementWindows drops the archived entries, preserving order.
func ActiveSettlementWindows(xs []SettlementWindow) []SettlementWindow {
	out := make([]SettlementWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementWindows adds up the attempts of every active entry.
func SumSettlementWindows(xs []SettlementWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexSettlementWindows keys the entries by ID, last write winning.
func IndexSettlementWindows(xs []SettlementWindow) map[string]SettlementWindow {
	out := make(map[string]SettlementWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementWindow builds a valid SettlementWindow, for tests and for the seed data the cron
// binary writes.
func SampleSettlementWindow(id string, at time.Time) SettlementWindow {
	return SettlementWindow{
		ID:        id,
		Reference: "reference-" + id,
		Category:  "category-" + id,
		Label:     "seeded by settlement.SampleSettlementWindow",
		Attempts:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SettlementQuotaSchema is the stored shape of a SettlementQuota.
var SettlementQuotaSchema = schema.New("settlement.settlementquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementQuota is one settlement quota.
type SettlementQuota struct {
	ID       string
	Comment  string
	Source   string
	Name     string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementQuota) Key() string { return "settlement/settlementquota/" + x.ID }

// Fields encodes x for storage.
func (x SettlementQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"name":     x.Name,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementQuota decodes a stored record.
func ParseSettlementQuota(f codec.Fields) (SettlementQuota, error) {
	var x SettlementQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementquota: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementquota %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementQuota) String() string {
	return fmt.Sprintf("SettlementQuota(%s, comment=%q, weight=%d, archived=%t)", x.ID, x.Comment, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementQuota) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementQuota) Normalize() SettlementQuota {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortSettlementQuotas orders by weight descending, then by ID.
func SortSettlementQuotas(xs []SettlementQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementQuotas drops the archived entries, preserving order.
func ActiveSettlementQuotas(xs []SettlementQuota) []SettlementQuota {
	out := make([]SettlementQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementQuotas adds up the weight of every active entry.
func SumSettlementQuotas(xs []SettlementQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexSettlementQuotas keys the entries by ID, last write winning.
func IndexSettlementQuotas(xs []SettlementQuota) map[string]SettlementQuota {
	out := make(map[string]SettlementQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementQuota builds a valid SettlementQuota, for tests and for the seed data the cron
// binary writes.
func SampleSettlementQuota(id string, at time.Time) SettlementQuota {
	return SettlementQuota{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Name:     "seeded by settlement.SampleSettlementQuota",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementBucketSchema is the stored shape of a SettlementBucket.
var SettlementBucketSchema = schema.New("settlement.settlementbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementBucket is one settlement bucket.
type SettlementBucket struct {
	ID       string
	Region   string
	Channel  string
	Group    string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementBucket) Key() string { return "settlement/settlementbucket/" + x.ID }

// Fields encodes x for storage.
func (x SettlementBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementBucket decodes a stored record.
func ParseSettlementBucket(f codec.Fields) (SettlementBucket, error) {
	var x SettlementBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementbucket: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementbucket %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementBucket) String() string {
	return fmt.Sprintf("SettlementBucket(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementBucket) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementBucket) Normalize() SettlementBucket {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSettlementBuckets orders by offset descending, then by ID.
func SortSettlementBuckets(xs []SettlementBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementBuckets drops the archived entries, preserving order.
func ActiveSettlementBuckets(xs []SettlementBucket) []SettlementBucket {
	out := make([]SettlementBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementBuckets adds up the offset of every active entry.
func SumSettlementBuckets(xs []SettlementBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexSettlementBuckets keys the entries by ID, last write winning.
func IndexSettlementBuckets(xs []SettlementBucket) map[string]SettlementBucket {
	out := make(map[string]SettlementBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementBucket builds a valid SettlementBucket, for tests and for the seed data the cron
// binary writes.
func SampleSettlementBucket(id string, at time.Time) SettlementBucket {
	return SettlementBucket{
		ID:      id,
		Region:  "region-" + id,
		Channel: "channel-" + id,
		Group:   "seeded by settlement.SampleSettlementBucket",
		Offset:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SettlementLaneSchema is the stored shape of a SettlementLane.
var SettlementLaneSchema = schema.New("settlement.settlementlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementLane is one settlement lane.
type SettlementLane struct {
	ID       string
	Label    string
	Name     string
	Channel  string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementLane) Key() string { return "settlement/settlementlane/" + x.ID }

// Fields encodes x for storage.
func (x SettlementLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"name":     x.Name,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementLane decodes a stored record.
func ParseSettlementLane(f codec.Fields) (SettlementLane, error) {
	var x SettlementLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementlane: missing id")
	}
	x.Label = f["label"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementlane %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementLane) String() string {
	return fmt.Sprintf("SettlementLane(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementLane) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementLane) Normalize() SettlementLane {
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSettlementLanes orders by weight descending, then by ID.
func SortSettlementLanes(xs []SettlementLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementLanes drops the archived entries, preserving order.
func ActiveSettlementLanes(xs []SettlementLane) []SettlementLane {
	out := make([]SettlementLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementLanes adds up the weight of every active entry.
func SumSettlementLanes(xs []SettlementLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexSettlementLanes keys the entries by ID, last write winning.
func IndexSettlementLanes(xs []SettlementLane) map[string]SettlementLane {
	out := make(map[string]SettlementLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementLane builds a valid SettlementLane, for tests and for the seed data the cron
// binary writes.
func SampleSettlementLane(id string, at time.Time) SettlementLane {
	return SettlementLane{
		ID:      id,
		Label:   "label-" + id,
		Name:    "name-" + id,
		Channel: "seeded by settlement.SampleSettlementLane",
		Weight:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SettlementSlotSchema is the stored shape of a SettlementSlot.
var SettlementSlotSchema = schema.New("settlement.settlementslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementSlot is one settlement slot.
type SettlementSlot struct {
	ID       string
	Region   string
	Channel  string
	Source   string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementSlot) Key() string { return "settlement/settlementslot/" + x.ID }

// Fields encodes x for storage.
func (x SettlementSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"source":   x.Source,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementSlot decodes a stored record.
func ParseSettlementSlot(f codec.Fields) (SettlementSlot, error) {
	var x SettlementSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementslot: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementSlot) String() string {
	return fmt.Sprintf("SettlementSlot(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementSlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementSlot) Normalize() SettlementSlot {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSettlementSlots orders by count descending, then by ID.
func SortSettlementSlots(xs []SettlementSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementSlots drops the archived entries, preserving order.
func ActiveSettlementSlots(xs []SettlementSlot) []SettlementSlot {
	out := make([]SettlementSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementSlots adds up the count of every active entry.
func SumSettlementSlots(xs []SettlementSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSettlementSlots keys the entries by ID, last write winning.
func IndexSettlementSlots(xs []SettlementSlot) map[string]SettlementSlot {
	out := make(map[string]SettlementSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementSlot builds a valid SettlementSlot, for tests and for the seed data the cron
// binary writes.
func SampleSettlementSlot(id string, at time.Time) SettlementSlot {
	return SettlementSlot{
		ID:      id,
		Region:  "region-" + id,
		Channel: "channel-" + id,
		Source:  "seeded by settlement.SampleSettlementSlot",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SettlementTagSchema is the stored shape of a SettlementTag.
var SettlementTagSchema = schema.New("settlement.settlementtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementTag is one settlement tag.
type SettlementTag struct {
	ID       string
	Group    string
	Channel  string
	Category string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementTag) Key() string { return "settlement/settlementtag/" + x.ID }

// Fields encodes x for storage.
func (x SettlementTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementTag decodes a stored record.
func ParseSettlementTag(f codec.Fields) (SettlementTag, error) {
	var x SettlementTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementtag: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementtag %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementTag) String() string {
	return fmt.Sprintf("SettlementTag(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementTag) Normalize() SettlementTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortSettlementTags orders by count descending, then by ID.
func SortSettlementTags(xs []SettlementTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementTags drops the archived entries, preserving order.
func ActiveSettlementTags(xs []SettlementTag) []SettlementTag {
	out := make([]SettlementTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementTags adds up the count of every active entry.
func SumSettlementTags(xs []SettlementTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSettlementTags keys the entries by ID, last write winning.
func IndexSettlementTags(xs []SettlementTag) map[string]SettlementTag {
	out := make(map[string]SettlementTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementTag builds a valid SettlementTag, for tests and for the seed data the cron
// binary writes.
func SampleSettlementTag(id string, at time.Time) SettlementTag {
	return SettlementTag{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by settlement.SampleSettlementTag",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementRuleSchema is the stored shape of a SettlementRule.
var SettlementRuleSchema = schema.New("settlement.settlementrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementRule is one settlement rule.
type SettlementRule struct {
	ID       string
	Region   string
	Source   string
	Category string
	Count    int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementRule) Key() string { return "settlement/settlementrule/" + x.ID }

// Fields encodes x for storage.
func (x SettlementRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"source":   x.Source,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementRule decodes a stored record.
func ParseSettlementRule(f codec.Fields) (SettlementRule, error) {
	var x SettlementRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementrule: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrule %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementRule) String() string {
	return fmt.Sprintf("SettlementRule(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementRule) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementRule) Normalize() SettlementRule {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortSettlementRules orders by count descending, then by ID.
func SortSettlementRules(xs []SettlementRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementRules drops the archived entries, preserving order.
func ActiveSettlementRules(xs []SettlementRule) []SettlementRule {
	out := make([]SettlementRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementRules adds up the count of every active entry.
func SumSettlementRules(xs []SettlementRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSettlementRules keys the entries by ID, last write winning.
func IndexSettlementRules(xs []SettlementRule) map[string]SettlementRule {
	out := make(map[string]SettlementRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementRule builds a valid SettlementRule, for tests and for the seed data the cron
// binary writes.
func SampleSettlementRule(id string, at time.Time) SettlementRule {
	return SettlementRule{
		ID:       id,
		Region:   "region-" + id,
		Source:   "source-" + id,
		Category: "seeded by settlement.SampleSettlementRule",
		Count:    len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// SettlementHoldSchema is the stored shape of a SettlementHold.
var SettlementHoldSchema = schema.New("settlement.settlementhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementHold is one settlement hold.
type SettlementHold struct {
	ID        string
	Reference string
	Source    string
	Comment   string
	Capacity  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x SettlementHold) Key() string { return "settlement/settlementhold/" + x.ID }

// Fields encodes x for storage.
func (x SettlementHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"source":    x.Source,
		"comment":   x.Comment,
		"capacity":  strconv.Itoa(x.Capacity),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementHold decodes a stored record.
func ParseSettlementHold(f codec.Fields) (SettlementHold, error) {
	var x SettlementHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementhold: missing id")
	}
	x.Reference = f["reference"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementhold %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementHold) String() string {
	return fmt.Sprintf("SettlementHold(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementHold) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementHold) Normalize() SettlementHold {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortSettlementHolds orders by capacity descending, then by ID.
func SortSettlementHolds(xs []SettlementHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementHolds drops the archived entries, preserving order.
func ActiveSettlementHolds(xs []SettlementHold) []SettlementHold {
	out := make([]SettlementHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementHolds adds up the capacity of every active entry.
func SumSettlementHolds(xs []SettlementHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexSettlementHolds keys the entries by ID, last write winning.
func IndexSettlementHolds(xs []SettlementHold) map[string]SettlementHold {
	out := make(map[string]SettlementHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementHold builds a valid SettlementHold, for tests and for the seed data the cron
// binary writes.
func SampleSettlementHold(id string, at time.Time) SettlementHold {
	return SettlementHold{
		ID:        id,
		Reference: "reference-" + id,
		Source:    "source-" + id,
		Comment:   "seeded by settlement.SampleSettlementHold",
		Capacity:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// SettlementClaimSchema is the stored shape of a SettlementClaim.
var SettlementClaimSchema = schema.New("settlement.settlementclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementClaim is one settlement claim.
type SettlementClaim struct {
	ID       string
	Owner    string
	Region   string
	Channel  string
	Version  int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementClaim) Key() string { return "settlement/settlementclaim/" + x.ID }

// Fields encodes x for storage.
func (x SettlementClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementClaim decodes a stored record.
func ParseSettlementClaim(f codec.Fields) (SettlementClaim, error) {
	var x SettlementClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementclaim %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementClaim) String() string {
	return fmt.Sprintf("SettlementClaim(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementClaim) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementClaim) Normalize() SettlementClaim {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortSettlementClaims orders by version descending, then by ID.
func SortSettlementClaims(xs []SettlementClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementClaims drops the archived entries, preserving order.
func ActiveSettlementClaims(xs []SettlementClaim) []SettlementClaim {
	out := make([]SettlementClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementClaims adds up the version of every active entry.
func SumSettlementClaims(xs []SettlementClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexSettlementClaims keys the entries by ID, last write winning.
func IndexSettlementClaims(xs []SettlementClaim) map[string]SettlementClaim {
	out := make(map[string]SettlementClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementClaim builds a valid SettlementClaim, for tests and for the seed data the cron
// binary writes.
func SampleSettlementClaim(id string, at time.Time) SettlementClaim {
	return SettlementClaim{
		ID:      id,
		Owner:   "owner-" + id,
		Region:  "region-" + id,
		Channel: "seeded by settlement.SampleSettlementClaim",
		Version: len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// SettlementCursorSchema is the stored shape of a SettlementCursor.
var SettlementCursorSchema = schema.New("settlement.settlementcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// SettlementCursor is one settlement cursor.
type SettlementCursor struct {
	ID       string
	Comment  string
	Channel  string
	Label    string
	Count    int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x SettlementCursor) Key() string { return "settlement/settlementcursor/" + x.ID }

// Fields encodes x for storage.
func (x SettlementCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseSettlementCursor decodes a stored record.
func ParseSettlementCursor(f codec.Fields) (SettlementCursor, error) {
	var x SettlementCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("settlement: settlementcursor: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementcursor %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("settlement: settlementcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("settlement: settlementcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("settlement: settlementcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x SettlementCursor) String() string {
	return fmt.Sprintf("SettlementCursor(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x SettlementCursor) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x SettlementCursor) Normalize() SettlementCursor {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortSettlementCursors orders by count descending, then by ID.
func SortSettlementCursors(xs []SettlementCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveSettlementCursors drops the archived entries, preserving order.
func ActiveSettlementCursors(xs []SettlementCursor) []SettlementCursor {
	out := make([]SettlementCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumSettlementCursors adds up the count of every active entry.
func SumSettlementCursors(xs []SettlementCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexSettlementCursors keys the entries by ID, last write winning.
func IndexSettlementCursors(xs []SettlementCursor) map[string]SettlementCursor {
	out := make(map[string]SettlementCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleSettlementCursor builds a valid SettlementCursor, for tests and for the seed data the cron
// binary writes.
func SampleSettlementCursor(id string, at time.Time) SettlementCursor {
	return SettlementCursor{
		ID:      id,
		Comment: "comment-" + id,
		Channel: "channel-" + id,
		Label:   "seeded by settlement.SampleSettlementCursor",
		Count:   len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	SettlementRecordSchema,
	SettlementEntrySchema,
	SettlementItemSchema,
	SettlementDraftSchema,
	SettlementRevisionSchema,
	SettlementSnapshotSchema,
	SettlementBatchSchema,
	SettlementRequestSchema,
	SettlementSummarySchema,
	SettlementPolicySchema,
	SettlementLinkSchema,
	SettlementNoteSchema,
	SettlementWindowSchema,
	SettlementQuotaSchema,
	SettlementBucketSchema,
	SettlementLaneSchema,
	SettlementSlotSchema,
	SettlementTagSchema,
	SettlementRuleSchema,
	SettlementHoldSchema,
	SettlementClaimSchema,
	SettlementCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"settlement/settlementrecord/",
	"settlement/settlemententry/",
	"settlement/settlementitem/",
	"settlement/settlementdraft/",
	"settlement/settlementrevision/",
	"settlement/settlementsnapshot/",
	"settlement/settlementbatch/",
	"settlement/settlementrequest/",
	"settlement/settlementsummary/",
	"settlement/settlementpolicy/",
	"settlement/settlementlink/",
	"settlement/settlementnote/",
	"settlement/settlementwindow/",
	"settlement/settlementquota/",
	"settlement/settlementbucket/",
	"settlement/settlementlane/",
	"settlement/settlementslot/",
	"settlement/settlementtag/",
	"settlement/settlementrule/",
	"settlement/settlementhold/",
	"settlement/settlementclaim/",
	"settlement/settlementcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "settlement: " + strings.Join(names, ", ")
}
