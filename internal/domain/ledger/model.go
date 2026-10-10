// Package ledger holds the ledger domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package ledger

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// LedgerRecordSchema is the stored shape of a LedgerRecord.
var LedgerRecordSchema = schema.New("ledger.ledgerrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerRecord is one ledger record.
type LedgerRecord struct {
	ID       string
	Source   string
	Category string
	Name     string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerRecord) Key() string { return "ledger/ledgerrecord/" + x.ID }

// Fields encodes x for storage.
func (x LedgerRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"category": x.Category,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerRecord decodes a stored record.
func ParseLedgerRecord(f codec.Fields) (LedgerRecord, error) {
	var x LedgerRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerrecord: missing id")
	}
	x.Source = f["source"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerRecord) String() string {
	return fmt.Sprintf("LedgerRecord(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerRecord) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerRecord) Normalize() LedgerRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLedgerRecords orders by retained descending, then by ID.
func SortLedgerRecords(xs []LedgerRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerRecords drops the archived entries, preserving order.
func ActiveLedgerRecords(xs []LedgerRecord) []LedgerRecord {
	out := make([]LedgerRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerRecords adds up the retained of every active entry.
func SumLedgerRecords(xs []LedgerRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLedgerRecords keys the entries by ID, last write winning.
func IndexLedgerRecords(xs []LedgerRecord) map[string]LedgerRecord {
	out := make(map[string]LedgerRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerRecord builds a valid LedgerRecord, for tests and for the seed data the cron
// binary writes.
func SampleLedgerRecord(id string, at time.Time) LedgerRecord {
	return LedgerRecord{
		ID:       id,
		Source:   "source-" + id,
		Category: "category-" + id,
		Name:     "seeded by ledger.SampleLedgerRecord",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerEntrySchema is the stored shape of a LedgerEntry.
var LedgerEntrySchema = schema.New("ledger.ledgerentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerEntry is one ledger entry.
type LedgerEntry struct {
	ID       string
	Label    string
	Owner    string
	Region   string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerEntry) Key() string { return "ledger/ledgerentry/" + x.ID }

// Fields encodes x for storage.
func (x LedgerEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"owner":    x.Owner,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerEntry decodes a stored record.
func ParseLedgerEntry(f codec.Fields) (LedgerEntry, error) {
	var x LedgerEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerentry: missing id")
	}
	x.Label = f["label"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerentry %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerEntry) String() string {
	return fmt.Sprintf("LedgerEntry(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerEntry) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerEntry) Normalize() LedgerEntry {
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortLedgerEntrys orders by weight descending, then by ID.
func SortLedgerEntrys(xs []LedgerEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerEntrys drops the archived entries, preserving order.
func ActiveLedgerEntrys(xs []LedgerEntry) []LedgerEntry {
	out := make([]LedgerEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerEntrys adds up the weight of every active entry.
func SumLedgerEntrys(xs []LedgerEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexLedgerEntrys keys the entries by ID, last write winning.
func IndexLedgerEntrys(xs []LedgerEntry) map[string]LedgerEntry {
	out := make(map[string]LedgerEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerEntry builds a valid LedgerEntry, for tests and for the seed data the cron
// binary writes.
func SampleLedgerEntry(id string, at time.Time) LedgerEntry {
	return LedgerEntry{
		ID:       id,
		Label:    "label-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by ledger.SampleLedgerEntry",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerItemSchema is the stored shape of a LedgerItem.
var LedgerItemSchema = schema.New("ledger.ledgeritem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerItem is one ledger item.
type LedgerItem struct {
	ID        string
	Channel   string
	Group     string
	Reference string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LedgerItem) Key() string { return "ledger/ledgeritem/" + x.ID }

// Fields encodes x for storage.
func (x LedgerItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"group":     x.Group,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerItem decodes a stored record.
func ParseLedgerItem(f codec.Fields) (LedgerItem, error) {
	var x LedgerItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgeritem: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgeritem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgeritem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgeritem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgeritem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerItem) String() string {
	return fmt.Sprintf("LedgerItem(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerItem) Normalize() LedgerItem {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortLedgerItems orders by version descending, then by ID.
func SortLedgerItems(xs []LedgerItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerItems drops the archived entries, preserving order.
func ActiveLedgerItems(xs []LedgerItem) []LedgerItem {
	out := make([]LedgerItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerItems adds up the version of every active entry.
func SumLedgerItems(xs []LedgerItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexLedgerItems keys the entries by ID, last write winning.
func IndexLedgerItems(xs []LedgerItem) map[string]LedgerItem {
	out := make(map[string]LedgerItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerItem builds a valid LedgerItem, for tests and for the seed data the cron
// binary writes.
func SampleLedgerItem(id string, at time.Time) LedgerItem {
	return LedgerItem{
		ID:        id,
		Channel:   "channel-" + id,
		Group:     "group-" + id,
		Reference: "seeded by ledger.SampleLedgerItem",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LedgerDraftSchema is the stored shape of a LedgerDraft.
var LedgerDraftSchema = schema.New("ledger.ledgerdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerDraft is one ledger draft.
type LedgerDraft struct {
	ID       string
	Category string
	Group    string
	Owner    string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerDraft) Key() string { return "ledger/ledgerdraft/" + x.ID }

// Fields encodes x for storage.
func (x LedgerDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"group":    x.Group,
		"owner":    x.Owner,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerDraft decodes a stored record.
func ParseLedgerDraft(f codec.Fields) (LedgerDraft, error) {
	var x LedgerDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerdraft: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerdraft %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerDraft) String() string {
	return fmt.Sprintf("LedgerDraft(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerDraft) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerDraft) Normalize() LedgerDraft {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortLedgerDrafts orders by weight descending, then by ID.
func SortLedgerDrafts(xs []LedgerDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerDrafts drops the archived entries, preserving order.
func ActiveLedgerDrafts(xs []LedgerDraft) []LedgerDraft {
	out := make([]LedgerDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerDrafts adds up the weight of every active entry.
func SumLedgerDrafts(xs []LedgerDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexLedgerDrafts keys the entries by ID, last write winning.
func IndexLedgerDrafts(xs []LedgerDraft) map[string]LedgerDraft {
	out := make(map[string]LedgerDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerDraft builds a valid LedgerDraft, for tests and for the seed data the cron
// binary writes.
func SampleLedgerDraft(id string, at time.Time) LedgerDraft {
	return LedgerDraft{
		ID:       id,
		Category: "category-" + id,
		Group:    "group-" + id,
		Owner:    "seeded by ledger.SampleLedgerDraft",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerRevisionSchema is the stored shape of a LedgerRevision.
var LedgerRevisionSchema = schema.New("ledger.ledgerrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerRevision is one ledger revision.
type LedgerRevision struct {
	ID       string
	Category string
	Comment  string
	Group    string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerRevision) Key() string { return "ledger/ledgerrevision/" + x.ID }

// Fields encodes x for storage.
func (x LedgerRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"comment":  x.Comment,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerRevision decodes a stored record.
func ParseLedgerRevision(f codec.Fields) (LedgerRevision, error) {
	var x LedgerRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerrevision: missing id")
	}
	x.Category = f["category"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerRevision) String() string {
	return fmt.Sprintf("LedgerRevision(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerRevision) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerRevision) Normalize() LedgerRevision {
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortLedgerRevisions orders by priority descending, then by ID.
func SortLedgerRevisions(xs []LedgerRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerRevisions drops the archived entries, preserving order.
func ActiveLedgerRevisions(xs []LedgerRevision) []LedgerRevision {
	out := make([]LedgerRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerRevisions adds up the priority of every active entry.
func SumLedgerRevisions(xs []LedgerRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexLedgerRevisions keys the entries by ID, last write winning.
func IndexLedgerRevisions(xs []LedgerRevision) map[string]LedgerRevision {
	out := make(map[string]LedgerRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerRevision builds a valid LedgerRevision, for tests and for the seed data the cron
// binary writes.
func SampleLedgerRevision(id string, at time.Time) LedgerRevision {
	return LedgerRevision{
		ID:       id,
		Category: "category-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by ledger.SampleLedgerRevision",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerSnapshotSchema is the stored shape of a LedgerSnapshot.
var LedgerSnapshotSchema = schema.New("ledger.ledgersnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerSnapshot is one ledger snapshot.
type LedgerSnapshot struct {
	ID        string
	Region    string
	Reference string
	Comment   string
	Retained  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LedgerSnapshot) Key() string { return "ledger/ledgersnapshot/" + x.ID }

// Fields encodes x for storage.
func (x LedgerSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"comment":   x.Comment,
		"retained":  strconv.Itoa(x.Retained),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerSnapshot decodes a stored record.
func ParseLedgerSnapshot(f codec.Fields) (LedgerSnapshot, error) {
	var x LedgerSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgersnapshot: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgersnapshot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgersnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgersnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgersnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerSnapshot) String() string {
	return fmt.Sprintf("LedgerSnapshot(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerSnapshot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerSnapshot) Normalize() LedgerSnapshot {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortLedgerSnapshots orders by retained descending, then by ID.
func SortLedgerSnapshots(xs []LedgerSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerSnapshots drops the archived entries, preserving order.
func ActiveLedgerSnapshots(xs []LedgerSnapshot) []LedgerSnapshot {
	out := make([]LedgerSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerSnapshots adds up the retained of every active entry.
func SumLedgerSnapshots(xs []LedgerSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLedgerSnapshots keys the entries by ID, last write winning.
func IndexLedgerSnapshots(xs []LedgerSnapshot) map[string]LedgerSnapshot {
	out := make(map[string]LedgerSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerSnapshot builds a valid LedgerSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleLedgerSnapshot(id string, at time.Time) LedgerSnapshot {
	return LedgerSnapshot{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by ledger.SampleLedgerSnapshot",
		Retained:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LedgerBatchSchema is the stored shape of a LedgerBatch.
var LedgerBatchSchema = schema.New("ledger.ledgerbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerBatch is one ledger batch.
type LedgerBatch struct {
	ID        string
	Reference string
	Category  string
	Group     string
	Retained  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LedgerBatch) Key() string { return "ledger/ledgerbatch/" + x.ID }

// Fields encodes x for storage.
func (x LedgerBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"category":  x.Category,
		"group":     x.Group,
		"retained":  strconv.Itoa(x.Retained),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerBatch decodes a stored record.
func ParseLedgerBatch(f codec.Fields) (LedgerBatch, error) {
	var x LedgerBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerbatch: missing id")
	}
	x.Reference = f["reference"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerbatch %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerBatch) String() string {
	return fmt.Sprintf("LedgerBatch(%s, reference=%q, retained=%d, archived=%t)", x.ID, x.Reference, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerBatch) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerBatch) Normalize() LedgerBatch {
	x.Reference = strings.TrimSpace(x.Reference)
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

// SortLedgerBatchs orders by retained descending, then by ID.
func SortLedgerBatchs(xs []LedgerBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerBatchs drops the archived entries, preserving order.
func ActiveLedgerBatchs(xs []LedgerBatch) []LedgerBatch {
	out := make([]LedgerBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerBatchs adds up the retained of every active entry.
func SumLedgerBatchs(xs []LedgerBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLedgerBatchs keys the entries by ID, last write winning.
func IndexLedgerBatchs(xs []LedgerBatch) map[string]LedgerBatch {
	out := make(map[string]LedgerBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerBatch builds a valid LedgerBatch, for tests and for the seed data the cron
// binary writes.
func SampleLedgerBatch(id string, at time.Time) LedgerBatch {
	return LedgerBatch{
		ID:        id,
		Reference: "reference-" + id,
		Category:  "category-" + id,
		Group:     "seeded by ledger.SampleLedgerBatch",
		Retained:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LedgerRequestSchema is the stored shape of a LedgerRequest.
var LedgerRequestSchema = schema.New("ledger.ledgerrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerRequest is one ledger request.
type LedgerRequest struct {
	ID       string
	Category string
	Comment  string
	Label    string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerRequest) Key() string { return "ledger/ledgerrequest/" + x.ID }

// Fields encodes x for storage.
func (x LedgerRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"comment":  x.Comment,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerRequest decodes a stored record.
func ParseLedgerRequest(f codec.Fields) (LedgerRequest, error) {
	var x LedgerRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerrequest: missing id")
	}
	x.Category = f["category"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerRequest) String() string {
	return fmt.Sprintf("LedgerRequest(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerRequest) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerRequest) Normalize() LedgerRequest {
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLedgerRequests orders by count descending, then by ID.
func SortLedgerRequests(xs []LedgerRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerRequests drops the archived entries, preserving order.
func ActiveLedgerRequests(xs []LedgerRequest) []LedgerRequest {
	out := make([]LedgerRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerRequests adds up the count of every active entry.
func SumLedgerRequests(xs []LedgerRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexLedgerRequests keys the entries by ID, last write winning.
func IndexLedgerRequests(xs []LedgerRequest) map[string]LedgerRequest {
	out := make(map[string]LedgerRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerRequest builds a valid LedgerRequest, for tests and for the seed data the cron
// binary writes.
func SampleLedgerRequest(id string, at time.Time) LedgerRequest {
	return LedgerRequest{
		ID:       id,
		Category: "category-" + id,
		Comment:  "comment-" + id,
		Label:    "seeded by ledger.SampleLedgerRequest",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerSummarySchema is the stored shape of a LedgerSummary.
var LedgerSummarySchema = schema.New("ledger.ledgersummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerSummary is one ledger summary.
type LedgerSummary struct {
	ID       string
	Owner    string
	Group    string
	Category string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerSummary) Key() string { return "ledger/ledgersummary/" + x.ID }

// Fields encodes x for storage.
func (x LedgerSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"category": x.Category,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerSummary decodes a stored record.
func ParseLedgerSummary(f codec.Fields) (LedgerSummary, error) {
	var x LedgerSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgersummary: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgersummary %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgersummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgersummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgersummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerSummary) String() string {
	return fmt.Sprintf("LedgerSummary(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerSummary) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerSummary) Normalize() LedgerSummary {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortLedgerSummarys orders by offset descending, then by ID.
func SortLedgerSummarys(xs []LedgerSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerSummarys drops the archived entries, preserving order.
func ActiveLedgerSummarys(xs []LedgerSummary) []LedgerSummary {
	out := make([]LedgerSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerSummarys adds up the offset of every active entry.
func SumLedgerSummarys(xs []LedgerSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexLedgerSummarys keys the entries by ID, last write winning.
func IndexLedgerSummarys(xs []LedgerSummary) map[string]LedgerSummary {
	out := make(map[string]LedgerSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerSummary builds a valid LedgerSummary, for tests and for the seed data the cron
// binary writes.
func SampleLedgerSummary(id string, at time.Time) LedgerSummary {
	return LedgerSummary{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Category: "seeded by ledger.SampleLedgerSummary",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerPolicySchema is the stored shape of a LedgerPolicy.
var LedgerPolicySchema = schema.New("ledger.ledgerpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerPolicy is one ledger policy.
type LedgerPolicy struct {
	ID       string
	Channel  string
	Group    string
	Source   string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerPolicy) Key() string { return "ledger/ledgerpolicy/" + x.ID }

// Fields encodes x for storage.
func (x LedgerPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerPolicy decodes a stored record.
func ParseLedgerPolicy(f codec.Fields) (LedgerPolicy, error) {
	var x LedgerPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerpolicy: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerpolicy %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerPolicy) String() string {
	return fmt.Sprintf("LedgerPolicy(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerPolicy) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerPolicy) Normalize() LedgerPolicy {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortLedgerPolicys orders by attempts descending, then by ID.
func SortLedgerPolicys(xs []LedgerPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerPolicys drops the archived entries, preserving order.
func ActiveLedgerPolicys(xs []LedgerPolicy) []LedgerPolicy {
	out := make([]LedgerPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerPolicys adds up the attempts of every active entry.
func SumLedgerPolicys(xs []LedgerPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLedgerPolicys keys the entries by ID, last write winning.
func IndexLedgerPolicys(xs []LedgerPolicy) map[string]LedgerPolicy {
	out := make(map[string]LedgerPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerPolicy builds a valid LedgerPolicy, for tests and for the seed data the cron
// binary writes.
func SampleLedgerPolicy(id string, at time.Time) LedgerPolicy {
	return LedgerPolicy{
		ID:       id,
		Channel:  "channel-" + id,
		Group:    "group-" + id,
		Source:   "seeded by ledger.SampleLedgerPolicy",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerLinkSchema is the stored shape of a LedgerLink.
var LedgerLinkSchema = schema.New("ledger.ledgerlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerLink is one ledger link.
type LedgerLink struct {
	ID       string
	Label    string
	Owner    string
	Name     string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerLink) Key() string { return "ledger/ledgerlink/" + x.ID }

// Fields encodes x for storage.
func (x LedgerLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"owner":    x.Owner,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerLink decodes a stored record.
func ParseLedgerLink(f codec.Fields) (LedgerLink, error) {
	var x LedgerLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerlink: missing id")
	}
	x.Label = f["label"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerlink %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerLink) String() string {
	return fmt.Sprintf("LedgerLink(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerLink) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerLink) Normalize() LedgerLink {
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortLedgerLinks orders by count descending, then by ID.
func SortLedgerLinks(xs []LedgerLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerLinks drops the archived entries, preserving order.
func ActiveLedgerLinks(xs []LedgerLink) []LedgerLink {
	out := make([]LedgerLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerLinks adds up the count of every active entry.
func SumLedgerLinks(xs []LedgerLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexLedgerLinks keys the entries by ID, last write winning.
func IndexLedgerLinks(xs []LedgerLink) map[string]LedgerLink {
	out := make(map[string]LedgerLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerLink builds a valid LedgerLink, for tests and for the seed data the cron
// binary writes.
func SampleLedgerLink(id string, at time.Time) LedgerLink {
	return LedgerLink{
		ID:      id,
		Label:   "label-" + id,
		Owner:   "owner-" + id,
		Name:    "seeded by ledger.SampleLedgerLink",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// LedgerNoteSchema is the stored shape of a LedgerNote.
var LedgerNoteSchema = schema.New("ledger.ledgernote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerNote is one ledger note.
type LedgerNote struct {
	ID       string
	Group    string
	Comment  string
	Owner    string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerNote) Key() string { return "ledger/ledgernote/" + x.ID }

// Fields encodes x for storage.
func (x LedgerNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerNote decodes a stored record.
func ParseLedgerNote(f codec.Fields) (LedgerNote, error) {
	var x LedgerNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgernote: missing id")
	}
	x.Group = f["group"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgernote %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgernote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgernote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgernote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerNote) String() string {
	return fmt.Sprintf("LedgerNote(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerNote) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerNote) Normalize() LedgerNote {
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortLedgerNotes orders by attempts descending, then by ID.
func SortLedgerNotes(xs []LedgerNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerNotes drops the archived entries, preserving order.
func ActiveLedgerNotes(xs []LedgerNote) []LedgerNote {
	out := make([]LedgerNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerNotes adds up the attempts of every active entry.
func SumLedgerNotes(xs []LedgerNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLedgerNotes keys the entries by ID, last write winning.
func IndexLedgerNotes(xs []LedgerNote) map[string]LedgerNote {
	out := make(map[string]LedgerNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerNote builds a valid LedgerNote, for tests and for the seed data the cron
// binary writes.
func SampleLedgerNote(id string, at time.Time) LedgerNote {
	return LedgerNote{
		ID:       id,
		Group:    "group-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by ledger.SampleLedgerNote",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerWindowSchema is the stored shape of a LedgerWindow.
var LedgerWindowSchema = schema.New("ledger.ledgerwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerWindow is one ledger window.
type LedgerWindow struct {
	ID        string
	Reference string
	Label     string
	Source    string
	Count     int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LedgerWindow) Key() string { return "ledger/ledgerwindow/" + x.ID }

// Fields encodes x for storage.
func (x LedgerWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"source":    x.Source,
		"count":     strconv.Itoa(x.Count),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerWindow decodes a stored record.
func ParseLedgerWindow(f codec.Fields) (LedgerWindow, error) {
	var x LedgerWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerwindow: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerwindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerWindow) String() string {
	return fmt.Sprintf("LedgerWindow(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerWindow) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerWindow) Normalize() LedgerWindow {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortLedgerWindows orders by count descending, then by ID.
func SortLedgerWindows(xs []LedgerWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerWindows drops the archived entries, preserving order.
func ActiveLedgerWindows(xs []LedgerWindow) []LedgerWindow {
	out := make([]LedgerWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerWindows adds up the count of every active entry.
func SumLedgerWindows(xs []LedgerWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexLedgerWindows keys the entries by ID, last write winning.
func IndexLedgerWindows(xs []LedgerWindow) map[string]LedgerWindow {
	out := make(map[string]LedgerWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerWindow builds a valid LedgerWindow, for tests and for the seed data the cron
// binary writes.
func SampleLedgerWindow(id string, at time.Time) LedgerWindow {
	return LedgerWindow{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Source:    "seeded by ledger.SampleLedgerWindow",
		Count:     len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LedgerQuotaSchema is the stored shape of a LedgerQuota.
var LedgerQuotaSchema = schema.New("ledger.ledgerquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerQuota is one ledger quota.
type LedgerQuota struct {
	ID       string
	Owner    string
	Group    string
	Source   string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerQuota) Key() string { return "ledger/ledgerquota/" + x.ID }

// Fields encodes x for storage.
func (x LedgerQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerQuota decodes a stored record.
func ParseLedgerQuota(f codec.Fields) (LedgerQuota, error) {
	var x LedgerQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerquota: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerquota %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerQuota) String() string {
	return fmt.Sprintf("LedgerQuota(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerQuota) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerQuota) Normalize() LedgerQuota {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortLedgerQuotas orders by capacity descending, then by ID.
func SortLedgerQuotas(xs []LedgerQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerQuotas drops the archived entries, preserving order.
func ActiveLedgerQuotas(xs []LedgerQuota) []LedgerQuota {
	out := make([]LedgerQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerQuotas adds up the capacity of every active entry.
func SumLedgerQuotas(xs []LedgerQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexLedgerQuotas keys the entries by ID, last write winning.
func IndexLedgerQuotas(xs []LedgerQuota) map[string]LedgerQuota {
	out := make(map[string]LedgerQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerQuota builds a valid LedgerQuota, for tests and for the seed data the cron
// binary writes.
func SampleLedgerQuota(id string, at time.Time) LedgerQuota {
	return LedgerQuota{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Source:   "seeded by ledger.SampleLedgerQuota",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerBucketSchema is the stored shape of a LedgerBucket.
var LedgerBucketSchema = schema.New("ledger.ledgerbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerBucket is one ledger bucket.
type LedgerBucket struct {
	ID        string
	Reference string
	Region    string
	Label     string
	Retained  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LedgerBucket) Key() string { return "ledger/ledgerbucket/" + x.ID }

// Fields encodes x for storage.
func (x LedgerBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"label":     x.Label,
		"retained":  strconv.Itoa(x.Retained),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerBucket decodes a stored record.
func ParseLedgerBucket(f codec.Fields) (LedgerBucket, error) {
	var x LedgerBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerbucket: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerbucket %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerBucket) String() string {
	return fmt.Sprintf("LedgerBucket(%s, reference=%q, retained=%d, archived=%t)", x.ID, x.Reference, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerBucket) Normalize() LedgerBucket {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortLedgerBuckets orders by retained descending, then by ID.
func SortLedgerBuckets(xs []LedgerBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerBuckets drops the archived entries, preserving order.
func ActiveLedgerBuckets(xs []LedgerBucket) []LedgerBucket {
	out := make([]LedgerBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerBuckets adds up the retained of every active entry.
func SumLedgerBuckets(xs []LedgerBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLedgerBuckets keys the entries by ID, last write winning.
func IndexLedgerBuckets(xs []LedgerBucket) map[string]LedgerBucket {
	out := make(map[string]LedgerBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerBucket builds a valid LedgerBucket, for tests and for the seed data the cron
// binary writes.
func SampleLedgerBucket(id string, at time.Time) LedgerBucket {
	return LedgerBucket{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Label:     "seeded by ledger.SampleLedgerBucket",
		Retained:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LedgerLaneSchema is the stored shape of a LedgerLane.
var LedgerLaneSchema = schema.New("ledger.ledgerlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerLane is one ledger lane.
type LedgerLane struct {
	ID       string
	Channel  string
	Group    string
	Label    string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerLane) Key() string { return "ledger/ledgerlane/" + x.ID }

// Fields encodes x for storage.
func (x LedgerLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerLane decodes a stored record.
func ParseLedgerLane(f codec.Fields) (LedgerLane, error) {
	var x LedgerLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerlane: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerlane %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerLane) String() string {
	return fmt.Sprintf("LedgerLane(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerLane) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerLane) Normalize() LedgerLane {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLedgerLanes orders by retained descending, then by ID.
func SortLedgerLanes(xs []LedgerLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerLanes drops the archived entries, preserving order.
func ActiveLedgerLanes(xs []LedgerLane) []LedgerLane {
	out := make([]LedgerLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerLanes adds up the retained of every active entry.
func SumLedgerLanes(xs []LedgerLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexLedgerLanes keys the entries by ID, last write winning.
func IndexLedgerLanes(xs []LedgerLane) map[string]LedgerLane {
	out := make(map[string]LedgerLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerLane builds a valid LedgerLane, for tests and for the seed data the cron
// binary writes.
func SampleLedgerLane(id string, at time.Time) LedgerLane {
	return LedgerLane{
		ID:       id,
		Channel:  "channel-" + id,
		Group:    "group-" + id,
		Label:    "seeded by ledger.SampleLedgerLane",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerSlotSchema is the stored shape of a LedgerSlot.
var LedgerSlotSchema = schema.New("ledger.ledgerslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerSlot is one ledger slot.
type LedgerSlot struct {
	ID       string
	Channel  string
	Comment  string
	Group    string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerSlot) Key() string { return "ledger/ledgerslot/" + x.ID }

// Fields encodes x for storage.
func (x LedgerSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerSlot decodes a stored record.
func ParseLedgerSlot(f codec.Fields) (LedgerSlot, error) {
	var x LedgerSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerslot: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerslot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerSlot) String() string {
	return fmt.Sprintf("LedgerSlot(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerSlot) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerSlot) Normalize() LedgerSlot {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortLedgerSlots orders by version descending, then by ID.
func SortLedgerSlots(xs []LedgerSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerSlots drops the archived entries, preserving order.
func ActiveLedgerSlots(xs []LedgerSlot) []LedgerSlot {
	out := make([]LedgerSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerSlots adds up the version of every active entry.
func SumLedgerSlots(xs []LedgerSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexLedgerSlots keys the entries by ID, last write winning.
func IndexLedgerSlots(xs []LedgerSlot) map[string]LedgerSlot {
	out := make(map[string]LedgerSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerSlot builds a valid LedgerSlot, for tests and for the seed data the cron
// binary writes.
func SampleLedgerSlot(id string, at time.Time) LedgerSlot {
	return LedgerSlot{
		ID:       id,
		Channel:  "channel-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by ledger.SampleLedgerSlot",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerTagSchema is the stored shape of a LedgerTag.
var LedgerTagSchema = schema.New("ledger.ledgertag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerTag is one ledger tag.
type LedgerTag struct {
	ID       string
	Region   string
	Name     string
	Category string
	Count    int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerTag) Key() string { return "ledger/ledgertag/" + x.ID }

// Fields encodes x for storage.
func (x LedgerTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerTag decodes a stored record.
func ParseLedgerTag(f codec.Fields) (LedgerTag, error) {
	var x LedgerTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgertag: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgertag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgertag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgertag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgertag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerTag) String() string {
	return fmt.Sprintf("LedgerTag(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerTag) Normalize() LedgerTag {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortLedgerTags orders by count descending, then by ID.
func SortLedgerTags(xs []LedgerTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerTags drops the archived entries, preserving order.
func ActiveLedgerTags(xs []LedgerTag) []LedgerTag {
	out := make([]LedgerTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerTags adds up the count of every active entry.
func SumLedgerTags(xs []LedgerTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexLedgerTags keys the entries by ID, last write winning.
func IndexLedgerTags(xs []LedgerTag) map[string]LedgerTag {
	out := make(map[string]LedgerTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerTag builds a valid LedgerTag, for tests and for the seed data the cron
// binary writes.
func SampleLedgerTag(id string, at time.Time) LedgerTag {
	return LedgerTag{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Category: "seeded by ledger.SampleLedgerTag",
		Count:    len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerRuleSchema is the stored shape of a LedgerRule.
var LedgerRuleSchema = schema.New("ledger.ledgerrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerRule is one ledger rule.
type LedgerRule struct {
	ID       string
	Label    string
	Name     string
	Group    string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerRule) Key() string { return "ledger/ledgerrule/" + x.ID }

// Fields encodes x for storage.
func (x LedgerRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"name":     x.Name,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerRule decodes a stored record.
func ParseLedgerRule(f codec.Fields) (LedgerRule, error) {
	var x LedgerRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerrule: missing id")
	}
	x.Label = f["label"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrule %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerRule) String() string {
	return fmt.Sprintf("LedgerRule(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerRule) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerRule) Normalize() LedgerRule {
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortLedgerRules orders by attempts descending, then by ID.
func SortLedgerRules(xs []LedgerRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerRules drops the archived entries, preserving order.
func ActiveLedgerRules(xs []LedgerRule) []LedgerRule {
	out := make([]LedgerRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerRules adds up the attempts of every active entry.
func SumLedgerRules(xs []LedgerRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexLedgerRules keys the entries by ID, last write winning.
func IndexLedgerRules(xs []LedgerRule) map[string]LedgerRule {
	out := make(map[string]LedgerRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerRule builds a valid LedgerRule, for tests and for the seed data the cron
// binary writes.
func SampleLedgerRule(id string, at time.Time) LedgerRule {
	return LedgerRule{
		ID:       id,
		Label:    "label-" + id,
		Name:     "name-" + id,
		Group:    "seeded by ledger.SampleLedgerRule",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// LedgerHoldSchema is the stored shape of a LedgerHold.
var LedgerHoldSchema = schema.New("ledger.ledgerhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerHold is one ledger hold.
type LedgerHold struct {
	ID       string
	Channel  string
	Group    string
	Label    string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerHold) Key() string { return "ledger/ledgerhold/" + x.ID }

// Fields encodes x for storage.
func (x LedgerHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerHold decodes a stored record.
func ParseLedgerHold(f codec.Fields) (LedgerHold, error) {
	var x LedgerHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerhold: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerhold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerHold) String() string {
	return fmt.Sprintf("LedgerHold(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerHold) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerHold) Normalize() LedgerHold {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortLedgerHolds orders by count descending, then by ID.
func SortLedgerHolds(xs []LedgerHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerHolds drops the archived entries, preserving order.
func ActiveLedgerHolds(xs []LedgerHold) []LedgerHold {
	out := make([]LedgerHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerHolds adds up the count of every active entry.
func SumLedgerHolds(xs []LedgerHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexLedgerHolds keys the entries by ID, last write winning.
func IndexLedgerHolds(xs []LedgerHold) map[string]LedgerHold {
	out := make(map[string]LedgerHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerHold builds a valid LedgerHold, for tests and for the seed data the cron
// binary writes.
func SampleLedgerHold(id string, at time.Time) LedgerHold {
	return LedgerHold{
		ID:      id,
		Channel: "channel-" + id,
		Group:   "group-" + id,
		Label:   "seeded by ledger.SampleLedgerHold",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// LedgerClaimSchema is the stored shape of a LedgerClaim.
var LedgerClaimSchema = schema.New("ledger.ledgerclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerClaim is one ledger claim.
type LedgerClaim struct {
	ID        string
	Region    string
	Reference string
	Source    string
	Weight    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x LedgerClaim) Key() string { return "ledger/ledgerclaim/" + x.ID }

// Fields encodes x for storage.
func (x LedgerClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"source":    x.Source,
		"weight":    strconv.Itoa(x.Weight),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerClaim decodes a stored record.
func ParseLedgerClaim(f codec.Fields) (LedgerClaim, error) {
	var x LedgerClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgerclaim: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerclaim %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgerclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgerclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgerclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerClaim) String() string {
	return fmt.Sprintf("LedgerClaim(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerClaim) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerClaim) Normalize() LedgerClaim {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortLedgerClaims orders by weight descending, then by ID.
func SortLedgerClaims(xs []LedgerClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerClaims drops the archived entries, preserving order.
func ActiveLedgerClaims(xs []LedgerClaim) []LedgerClaim {
	out := make([]LedgerClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerClaims adds up the weight of every active entry.
func SumLedgerClaims(xs []LedgerClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexLedgerClaims keys the entries by ID, last write winning.
func IndexLedgerClaims(xs []LedgerClaim) map[string]LedgerClaim {
	out := make(map[string]LedgerClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerClaim builds a valid LedgerClaim, for tests and for the seed data the cron
// binary writes.
func SampleLedgerClaim(id string, at time.Time) LedgerClaim {
	return LedgerClaim{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by ledger.SampleLedgerClaim",
		Weight:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// LedgerCursorSchema is the stored shape of a LedgerCursor.
var LedgerCursorSchema = schema.New("ledger.ledgercursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// LedgerCursor is one ledger cursor.
type LedgerCursor struct {
	ID       string
	Region   string
	Group    string
	Comment  string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x LedgerCursor) Key() string { return "ledger/ledgercursor/" + x.ID }

// Fields encodes x for storage.
func (x LedgerCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"group":    x.Group,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseLedgerCursor decodes a stored record.
func ParseLedgerCursor(f codec.Fields) (LedgerCursor, error) {
	var x LedgerCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("ledger: ledgercursor: missing id")
	}
	x.Region = f["region"]
	x.Group = f["group"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgercursor %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("ledger: ledgercursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("ledger: ledgercursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("ledger: ledgercursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x LedgerCursor) String() string {
	return fmt.Sprintf("LedgerCursor(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x LedgerCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x LedgerCursor) Normalize() LedgerCursor {
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortLedgerCursors orders by version descending, then by ID.
func SortLedgerCursors(xs []LedgerCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveLedgerCursors drops the archived entries, preserving order.
func ActiveLedgerCursors(xs []LedgerCursor) []LedgerCursor {
	out := make([]LedgerCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumLedgerCursors adds up the version of every active entry.
func SumLedgerCursors(xs []LedgerCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexLedgerCursors keys the entries by ID, last write winning.
func IndexLedgerCursors(xs []LedgerCursor) map[string]LedgerCursor {
	out := make(map[string]LedgerCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleLedgerCursor builds a valid LedgerCursor, for tests and for the seed data the cron
// binary writes.
func SampleLedgerCursor(id string, at time.Time) LedgerCursor {
	return LedgerCursor{
		ID:       id,
		Region:   "region-" + id,
		Group:    "group-" + id,
		Comment:  "seeded by ledger.SampleLedgerCursor",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	LedgerRecordSchema,
	LedgerEntrySchema,
	LedgerItemSchema,
	LedgerDraftSchema,
	LedgerRevisionSchema,
	LedgerSnapshotSchema,
	LedgerBatchSchema,
	LedgerRequestSchema,
	LedgerSummarySchema,
	LedgerPolicySchema,
	LedgerLinkSchema,
	LedgerNoteSchema,
	LedgerWindowSchema,
	LedgerQuotaSchema,
	LedgerBucketSchema,
	LedgerLaneSchema,
	LedgerSlotSchema,
	LedgerTagSchema,
	LedgerRuleSchema,
	LedgerHoldSchema,
	LedgerClaimSchema,
	LedgerCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"ledger/ledgerrecord/",
	"ledger/ledgerentry/",
	"ledger/ledgeritem/",
	"ledger/ledgerdraft/",
	"ledger/ledgerrevision/",
	"ledger/ledgersnapshot/",
	"ledger/ledgerbatch/",
	"ledger/ledgerrequest/",
	"ledger/ledgersummary/",
	"ledger/ledgerpolicy/",
	"ledger/ledgerlink/",
	"ledger/ledgernote/",
	"ledger/ledgerwindow/",
	"ledger/ledgerquota/",
	"ledger/ledgerbucket/",
	"ledger/ledgerlane/",
	"ledger/ledgerslot/",
	"ledger/ledgertag/",
	"ledger/ledgerrule/",
	"ledger/ledgerhold/",
	"ledger/ledgerclaim/",
	"ledger/ledgercursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "ledger: " + strings.Join(names, ", ")
}
