// Package billing holds the billing domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package billing

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// BillingRecordSchema is the stored shape of a BillingRecord.
var BillingRecordSchema = schema.New("billing.billingrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingRecord is one billing record.
type BillingRecord struct {
	ID       string
	Channel  string
	Category string
	Region   string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingRecord) Key() string { return "billing/billingrecord/" + x.ID }

// Fields encodes x for storage.
func (x BillingRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"region":   x.Region,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingRecord decodes a stored record.
func ParseBillingRecord(f codec.Fields) (BillingRecord, error) {
	var x BillingRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingrecord: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("billing: billingrecord %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingRecord) String() string {
	return fmt.Sprintf("BillingRecord(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingRecord) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingRecord) Normalize() BillingRecord {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortBillingRecords orders by count descending, then by ID.
func SortBillingRecords(xs []BillingRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingRecords drops the archived entries, preserving order.
func ActiveBillingRecords(xs []BillingRecord) []BillingRecord {
	out := make([]BillingRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingRecords adds up the count of every active entry.
func SumBillingRecords(xs []BillingRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexBillingRecords keys the entries by ID, last write winning.
func IndexBillingRecords(xs []BillingRecord) map[string]BillingRecord {
	out := make(map[string]BillingRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingRecord builds a valid BillingRecord, for tests and for the seed data the cron
// binary writes.
func SampleBillingRecord(id string, at time.Time) BillingRecord {
	return BillingRecord{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Region:   "seeded by billing.SampleBillingRecord",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingEntrySchema is the stored shape of a BillingEntry.
var BillingEntrySchema = schema.New("billing.billingentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingEntry is one billing entry.
type BillingEntry struct {
	ID       string
	Group    string
	Name     string
	Channel  string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingEntry) Key() string { return "billing/billingentry/" + x.ID }

// Fields encodes x for storage.
func (x BillingEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingEntry decodes a stored record.
func ParseBillingEntry(f codec.Fields) (BillingEntry, error) {
	var x BillingEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingentry: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingentry %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("billing: billingentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingEntry) String() string {
	return fmt.Sprintf("BillingEntry(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingEntry) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingEntry) Normalize() BillingEntry {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortBillingEntrys orders by version descending, then by ID.
func SortBillingEntrys(xs []BillingEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingEntrys drops the archived entries, preserving order.
func ActiveBillingEntrys(xs []BillingEntry) []BillingEntry {
	out := make([]BillingEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingEntrys adds up the version of every active entry.
func SumBillingEntrys(xs []BillingEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexBillingEntrys keys the entries by ID, last write winning.
func IndexBillingEntrys(xs []BillingEntry) map[string]BillingEntry {
	out := make(map[string]BillingEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingEntry builds a valid BillingEntry, for tests and for the seed data the cron
// binary writes.
func SampleBillingEntry(id string, at time.Time) BillingEntry {
	return BillingEntry{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by billing.SampleBillingEntry",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingItemSchema is the stored shape of a BillingItem.
var BillingItemSchema = schema.New("billing.billingitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingItem is one billing item.
type BillingItem struct {
	ID       string
	Category string
	Region   string
	Comment  string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingItem) Key() string { return "billing/billingitem/" + x.ID }

// Fields encodes x for storage.
func (x BillingItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingItem decodes a stored record.
func ParseBillingItem(f codec.Fields) (BillingItem, error) {
	var x BillingItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingitem: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("billing: billingitem %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingItem) String() string {
	return fmt.Sprintf("BillingItem(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingItem) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingItem) Normalize() BillingItem {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortBillingItems orders by retained descending, then by ID.
func SortBillingItems(xs []BillingItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingItems drops the archived entries, preserving order.
func ActiveBillingItems(xs []BillingItem) []BillingItem {
	out := make([]BillingItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingItems adds up the retained of every active entry.
func SumBillingItems(xs []BillingItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexBillingItems keys the entries by ID, last write winning.
func IndexBillingItems(xs []BillingItem) map[string]BillingItem {
	out := make(map[string]BillingItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingItem builds a valid BillingItem, for tests and for the seed data the cron
// binary writes.
func SampleBillingItem(id string, at time.Time) BillingItem {
	return BillingItem{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Comment:  "seeded by billing.SampleBillingItem",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingDraftSchema is the stored shape of a BillingDraft.
var BillingDraftSchema = schema.New("billing.billingdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingDraft is one billing draft.
type BillingDraft struct {
	ID       string
	Owner    string
	Category string
	Group    string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingDraft) Key() string { return "billing/billingdraft/" + x.ID }

// Fields encodes x for storage.
func (x BillingDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"category": x.Category,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingDraft decodes a stored record.
func ParseBillingDraft(f codec.Fields) (BillingDraft, error) {
	var x BillingDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingdraft: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingdraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingDraft) String() string {
	return fmt.Sprintf("BillingDraft(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingDraft) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingDraft) Normalize() BillingDraft {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortBillingDrafts orders by version descending, then by ID.
func SortBillingDrafts(xs []BillingDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingDrafts drops the archived entries, preserving order.
func ActiveBillingDrafts(xs []BillingDraft) []BillingDraft {
	out := make([]BillingDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingDrafts adds up the version of every active entry.
func SumBillingDrafts(xs []BillingDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexBillingDrafts keys the entries by ID, last write winning.
func IndexBillingDrafts(xs []BillingDraft) map[string]BillingDraft {
	out := make(map[string]BillingDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingDraft builds a valid BillingDraft, for tests and for the seed data the cron
// binary writes.
func SampleBillingDraft(id string, at time.Time) BillingDraft {
	return BillingDraft{
		ID:       id,
		Owner:    "owner-" + id,
		Category: "category-" + id,
		Group:    "seeded by billing.SampleBillingDraft",
		Version:  len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingRevisionSchema is the stored shape of a BillingRevision.
var BillingRevisionSchema = schema.New("billing.billingrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingRevision is one billing revision.
type BillingRevision struct {
	ID       string
	Category string
	Source   string
	Comment  string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingRevision) Key() string { return "billing/billingrevision/" + x.ID }

// Fields encodes x for storage.
func (x BillingRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"source":   x.Source,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingRevision decodes a stored record.
func ParseBillingRevision(f codec.Fields) (BillingRevision, error) {
	var x BillingRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingrevision: missing id")
	}
	x.Category = f["category"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("billing: billingrevision %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingRevision) String() string {
	return fmt.Sprintf("BillingRevision(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingRevision) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingRevision) Normalize() BillingRevision {
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortBillingRevisions orders by capacity descending, then by ID.
func SortBillingRevisions(xs []BillingRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingRevisions drops the archived entries, preserving order.
func ActiveBillingRevisions(xs []BillingRevision) []BillingRevision {
	out := make([]BillingRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingRevisions adds up the capacity of every active entry.
func SumBillingRevisions(xs []BillingRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexBillingRevisions keys the entries by ID, last write winning.
func IndexBillingRevisions(xs []BillingRevision) map[string]BillingRevision {
	out := make(map[string]BillingRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingRevision builds a valid BillingRevision, for tests and for the seed data the cron
// binary writes.
func SampleBillingRevision(id string, at time.Time) BillingRevision {
	return BillingRevision{
		ID:       id,
		Category: "category-" + id,
		Source:   "source-" + id,
		Comment:  "seeded by billing.SampleBillingRevision",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingSnapshotSchema is the stored shape of a BillingSnapshot.
var BillingSnapshotSchema = schema.New("billing.billingsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingSnapshot is one billing snapshot.
type BillingSnapshot struct {
	ID        string
	Region    string
	Comment   string
	Reference string
	Retained  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x BillingSnapshot) Key() string { return "billing/billingsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x BillingSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"comment":   x.Comment,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingSnapshot decodes a stored record.
func ParseBillingSnapshot(f codec.Fields) (BillingSnapshot, error) {
	var x BillingSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingsnapshot: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("billing: billingsnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("billing: billingsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingSnapshot) String() string {
	return fmt.Sprintf("BillingSnapshot(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingSnapshot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingSnapshot) Normalize() BillingSnapshot {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortBillingSnapshots orders by retained descending, then by ID.
func SortBillingSnapshots(xs []BillingSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingSnapshots drops the archived entries, preserving order.
func ActiveBillingSnapshots(xs []BillingSnapshot) []BillingSnapshot {
	out := make([]BillingSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingSnapshots adds up the retained of every active entry.
func SumBillingSnapshots(xs []BillingSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexBillingSnapshots keys the entries by ID, last write winning.
func IndexBillingSnapshots(xs []BillingSnapshot) map[string]BillingSnapshot {
	out := make(map[string]BillingSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingSnapshot builds a valid BillingSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleBillingSnapshot(id string, at time.Time) BillingSnapshot {
	return BillingSnapshot{
		ID:        id,
		Region:    "region-" + id,
		Comment:   "comment-" + id,
		Reference: "seeded by billing.SampleBillingSnapshot",
		Retained:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// BillingBatchSchema is the stored shape of a BillingBatch.
var BillingBatchSchema = schema.New("billing.billingbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingBatch is one billing batch.
type BillingBatch struct {
	ID        string
	Reference string
	Comment   string
	Group     string
	Version   int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x BillingBatch) Key() string { return "billing/billingbatch/" + x.ID }

// Fields encodes x for storage.
func (x BillingBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"comment":   x.Comment,
		"group":     x.Group,
		"version":   strconv.Itoa(x.Version),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingBatch decodes a stored record.
func ParseBillingBatch(f codec.Fields) (BillingBatch, error) {
	var x BillingBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingbatch: missing id")
	}
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingbatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("billing: billingbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingBatch) String() string {
	return fmt.Sprintf("BillingBatch(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingBatch) Normalize() BillingBatch {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortBillingBatchs orders by version descending, then by ID.
func SortBillingBatchs(xs []BillingBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingBatchs drops the archived entries, preserving order.
func ActiveBillingBatchs(xs []BillingBatch) []BillingBatch {
	out := make([]BillingBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingBatchs adds up the version of every active entry.
func SumBillingBatchs(xs []BillingBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexBillingBatchs keys the entries by ID, last write winning.
func IndexBillingBatchs(xs []BillingBatch) map[string]BillingBatch {
	out := make(map[string]BillingBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingBatch builds a valid BillingBatch, for tests and for the seed data the cron
// binary writes.
func SampleBillingBatch(id string, at time.Time) BillingBatch {
	return BillingBatch{
		ID:        id,
		Reference: "reference-" + id,
		Comment:   "comment-" + id,
		Group:     "seeded by billing.SampleBillingBatch",
		Version:   len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// BillingRequestSchema is the stored shape of a BillingRequest.
var BillingRequestSchema = schema.New("billing.billingrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingRequest is one billing request.
type BillingRequest struct {
	ID       string
	Group    string
	Name     string
	Region   string
	Weight   int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingRequest) Key() string { return "billing/billingrequest/" + x.ID }

// Fields encodes x for storage.
func (x BillingRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingRequest decodes a stored record.
func ParseBillingRequest(f codec.Fields) (BillingRequest, error) {
	var x BillingRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingrequest: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingrequest %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("billing: billingrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingRequest) String() string {
	return fmt.Sprintf("BillingRequest(%s, group=%q, weight=%d, archived=%t)", x.ID, x.Group, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingRequest) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingRequest) Normalize() BillingRequest {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortBillingRequests orders by weight descending, then by ID.
func SortBillingRequests(xs []BillingRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingRequests drops the archived entries, preserving order.
func ActiveBillingRequests(xs []BillingRequest) []BillingRequest {
	out := make([]BillingRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingRequests adds up the weight of every active entry.
func SumBillingRequests(xs []BillingRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexBillingRequests keys the entries by ID, last write winning.
func IndexBillingRequests(xs []BillingRequest) map[string]BillingRequest {
	out := make(map[string]BillingRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingRequest builds a valid BillingRequest, for tests and for the seed data the cron
// binary writes.
func SampleBillingRequest(id string, at time.Time) BillingRequest {
	return BillingRequest{
		ID:      id,
		Group:   "group-" + id,
		Name:    "name-" + id,
		Region:  "seeded by billing.SampleBillingRequest",
		Weight:  len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// BillingSummarySchema is the stored shape of a BillingSummary.
var BillingSummarySchema = schema.New("billing.billingsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingSummary is one billing summary.
type BillingSummary struct {
	ID        string
	Group     string
	Reference string
	Region    string
	Retained  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x BillingSummary) Key() string { return "billing/billingsummary/" + x.ID }

// Fields encodes x for storage.
func (x BillingSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"region":    x.Region,
		"retained":  strconv.Itoa(x.Retained),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingSummary decodes a stored record.
func ParseBillingSummary(f codec.Fields) (BillingSummary, error) {
	var x BillingSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingsummary: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("billing: billingsummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingSummary) String() string {
	return fmt.Sprintf("BillingSummary(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingSummary) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingSummary) Normalize() BillingSummary {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortBillingSummarys orders by retained descending, then by ID.
func SortBillingSummarys(xs []BillingSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingSummarys drops the archived entries, preserving order.
func ActiveBillingSummarys(xs []BillingSummary) []BillingSummary {
	out := make([]BillingSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingSummarys adds up the retained of every active entry.
func SumBillingSummarys(xs []BillingSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexBillingSummarys keys the entries by ID, last write winning.
func IndexBillingSummarys(xs []BillingSummary) map[string]BillingSummary {
	out := make(map[string]BillingSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingSummary builds a valid BillingSummary, for tests and for the seed data the cron
// binary writes.
func SampleBillingSummary(id string, at time.Time) BillingSummary {
	return BillingSummary{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by billing.SampleBillingSummary",
		Retained:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// BillingPolicySchema is the stored shape of a BillingPolicy.
var BillingPolicySchema = schema.New("billing.billingpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingPolicy is one billing policy.
type BillingPolicy struct {
	ID       string
	Region   string
	Label    string
	Category string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingPolicy) Key() string { return "billing/billingpolicy/" + x.ID }

// Fields encodes x for storage.
func (x BillingPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"category": x.Category,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingPolicy decodes a stored record.
func ParseBillingPolicy(f codec.Fields) (BillingPolicy, error) {
	var x BillingPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingpolicy: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Category = f["category"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("billing: billingpolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("billing: billingpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingPolicy) String() string {
	return fmt.Sprintf("BillingPolicy(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingPolicy) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingPolicy) Normalize() BillingPolicy {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortBillingPolicys orders by priority descending, then by ID.
func SortBillingPolicys(xs []BillingPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingPolicys drops the archived entries, preserving order.
func ActiveBillingPolicys(xs []BillingPolicy) []BillingPolicy {
	out := make([]BillingPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingPolicys adds up the priority of every active entry.
func SumBillingPolicys(xs []BillingPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexBillingPolicys keys the entries by ID, last write winning.
func IndexBillingPolicys(xs []BillingPolicy) map[string]BillingPolicy {
	out := make(map[string]BillingPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingPolicy builds a valid BillingPolicy, for tests and for the seed data the cron
// binary writes.
func SampleBillingPolicy(id string, at time.Time) BillingPolicy {
	return BillingPolicy{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Category: "seeded by billing.SampleBillingPolicy",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingLinkSchema is the stored shape of a BillingLink.
var BillingLinkSchema = schema.New("billing.billinglink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingLink is one billing link.
type BillingLink struct {
	ID       string
	Comment  string
	Source   string
	Name     string
	Retained int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingLink) Key() string { return "billing/billinglink/" + x.ID }

// Fields encodes x for storage.
func (x BillingLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingLink decodes a stored record.
func ParseBillingLink(f codec.Fields) (BillingLink, error) {
	var x BillingLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billinglink: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("billing: billinglink %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("billing: billinglink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billinglink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billinglink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingLink) String() string {
	return fmt.Sprintf("BillingLink(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingLink) Normalize() BillingLink {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortBillingLinks orders by retained descending, then by ID.
func SortBillingLinks(xs []BillingLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingLinks drops the archived entries, preserving order.
func ActiveBillingLinks(xs []BillingLink) []BillingLink {
	out := make([]BillingLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingLinks adds up the retained of every active entry.
func SumBillingLinks(xs []BillingLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexBillingLinks keys the entries by ID, last write winning.
func IndexBillingLinks(xs []BillingLink) map[string]BillingLink {
	out := make(map[string]BillingLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingLink builds a valid BillingLink, for tests and for the seed data the cron
// binary writes.
func SampleBillingLink(id string, at time.Time) BillingLink {
	return BillingLink{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Name:     "seeded by billing.SampleBillingLink",
		Retained: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingNoteSchema is the stored shape of a BillingNote.
var BillingNoteSchema = schema.New("billing.billingnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingNote is one billing note.
type BillingNote struct {
	ID       string
	Region   string
	Comment  string
	Channel  string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingNote) Key() string { return "billing/billingnote/" + x.ID }

// Fields encodes x for storage.
func (x BillingNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingNote decodes a stored record.
func ParseBillingNote(f codec.Fields) (BillingNote, error) {
	var x BillingNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingnote: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("billing: billingnote %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingNote) String() string {
	return fmt.Sprintf("BillingNote(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingNote) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingNote) Normalize() BillingNote {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortBillingNotes orders by offset descending, then by ID.
func SortBillingNotes(xs []BillingNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingNotes drops the archived entries, preserving order.
func ActiveBillingNotes(xs []BillingNote) []BillingNote {
	out := make([]BillingNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingNotes adds up the offset of every active entry.
func SumBillingNotes(xs []BillingNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexBillingNotes keys the entries by ID, last write winning.
func IndexBillingNotes(xs []BillingNote) map[string]BillingNote {
	out := make(map[string]BillingNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingNote builds a valid BillingNote, for tests and for the seed data the cron
// binary writes.
func SampleBillingNote(id string, at time.Time) BillingNote {
	return BillingNote{
		ID:      id,
		Region:  "region-" + id,
		Comment: "comment-" + id,
		Channel: "seeded by billing.SampleBillingNote",
		Offset:  len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// BillingWindowSchema is the stored shape of a BillingWindow.
var BillingWindowSchema = schema.New("billing.billingwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingWindow is one billing window.
type BillingWindow struct {
	ID       string
	Channel  string
	Group    string
	Label    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingWindow) Key() string { return "billing/billingwindow/" + x.ID }

// Fields encodes x for storage.
func (x BillingWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingWindow decodes a stored record.
func ParseBillingWindow(f codec.Fields) (BillingWindow, error) {
	var x BillingWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingwindow: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("billing: billingwindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingWindow) String() string {
	return fmt.Sprintf("BillingWindow(%s, channel=%q, offset=%d, archived=%t)", x.ID, x.Channel, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingWindow) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingWindow) Normalize() BillingWindow {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortBillingWindows orders by offset descending, then by ID.
func SortBillingWindows(xs []BillingWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingWindows drops the archived entries, preserving order.
func ActiveBillingWindows(xs []BillingWindow) []BillingWindow {
	out := make([]BillingWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingWindows adds up the offset of every active entry.
func SumBillingWindows(xs []BillingWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexBillingWindows keys the entries by ID, last write winning.
func IndexBillingWindows(xs []BillingWindow) map[string]BillingWindow {
	out := make(map[string]BillingWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingWindow builds a valid BillingWindow, for tests and for the seed data the cron
// binary writes.
func SampleBillingWindow(id string, at time.Time) BillingWindow {
	return BillingWindow{
		ID:      id,
		Channel: "channel-" + id,
		Group:   "group-" + id,
		Label:   "seeded by billing.SampleBillingWindow",
		Offset:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// BillingQuotaSchema is the stored shape of a BillingQuota.
var BillingQuotaSchema = schema.New("billing.billingquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingQuota is one billing quota.
type BillingQuota struct {
	ID       string
	Category string
	Owner    string
	Source   string
	Version  int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingQuota) Key() string { return "billing/billingquota/" + x.ID }

// Fields encodes x for storage.
func (x BillingQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingQuota decodes a stored record.
func ParseBillingQuota(f codec.Fields) (BillingQuota, error) {
	var x BillingQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingquota: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingquota %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("billing: billingquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingQuota) String() string {
	return fmt.Sprintf("BillingQuota(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingQuota) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingQuota) Normalize() BillingQuota {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortBillingQuotas orders by version descending, then by ID.
func SortBillingQuotas(xs []BillingQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingQuotas drops the archived entries, preserving order.
func ActiveBillingQuotas(xs []BillingQuota) []BillingQuota {
	out := make([]BillingQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingQuotas adds up the version of every active entry.
func SumBillingQuotas(xs []BillingQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexBillingQuotas keys the entries by ID, last write winning.
func IndexBillingQuotas(xs []BillingQuota) map[string]BillingQuota {
	out := make(map[string]BillingQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingQuota builds a valid BillingQuota, for tests and for the seed data the cron
// binary writes.
func SampleBillingQuota(id string, at time.Time) BillingQuota {
	return BillingQuota{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by billing.SampleBillingQuota",
		Version:  len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingBucketSchema is the stored shape of a BillingBucket.
var BillingBucketSchema = schema.New("billing.billingbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingBucket is one billing bucket.
type BillingBucket struct {
	ID       string
	Label    string
	Group    string
	Owner    string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingBucket) Key() string { return "billing/billingbucket/" + x.ID }

// Fields encodes x for storage.
func (x BillingBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingBucket decodes a stored record.
func ParseBillingBucket(f codec.Fields) (BillingBucket, error) {
	var x BillingBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingbucket: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("billing: billingbucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("billing: billingbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingBucket) String() string {
	return fmt.Sprintf("BillingBucket(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingBucket) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingBucket) Normalize() BillingBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortBillingBuckets orders by attempts descending, then by ID.
func SortBillingBuckets(xs []BillingBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingBuckets drops the archived entries, preserving order.
func ActiveBillingBuckets(xs []BillingBucket) []BillingBucket {
	out := make([]BillingBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingBuckets adds up the attempts of every active entry.
func SumBillingBuckets(xs []BillingBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexBillingBuckets keys the entries by ID, last write winning.
func IndexBillingBuckets(xs []BillingBucket) map[string]BillingBucket {
	out := make(map[string]BillingBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingBucket builds a valid BillingBucket, for tests and for the seed data the cron
// binary writes.
func SampleBillingBucket(id string, at time.Time) BillingBucket {
	return BillingBucket{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Owner:    "seeded by billing.SampleBillingBucket",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingLaneSchema is the stored shape of a BillingLane.
var BillingLaneSchema = schema.New("billing.billinglane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingLane is one billing lane.
type BillingLane struct {
	ID        string
	Reference string
	Comment   string
	Source    string
	Priority  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x BillingLane) Key() string { return "billing/billinglane/" + x.ID }

// Fields encodes x for storage.
func (x BillingLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"comment":   x.Comment,
		"source":    x.Source,
		"priority":  strconv.Itoa(x.Priority),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingLane decodes a stored record.
func ParseBillingLane(f codec.Fields) (BillingLane, error) {
	var x BillingLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billinglane: missing id")
	}
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("billing: billinglane %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("billing: billinglane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billinglane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billinglane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingLane) String() string {
	return fmt.Sprintf("BillingLane(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingLane) Normalize() BillingLane {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortBillingLanes orders by priority descending, then by ID.
func SortBillingLanes(xs []BillingLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingLanes drops the archived entries, preserving order.
func ActiveBillingLanes(xs []BillingLane) []BillingLane {
	out := make([]BillingLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingLanes adds up the priority of every active entry.
func SumBillingLanes(xs []BillingLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexBillingLanes keys the entries by ID, last write winning.
func IndexBillingLanes(xs []BillingLane) map[string]BillingLane {
	out := make(map[string]BillingLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingLane builds a valid BillingLane, for tests and for the seed data the cron
// binary writes.
func SampleBillingLane(id string, at time.Time) BillingLane {
	return BillingLane{
		ID:        id,
		Reference: "reference-" + id,
		Comment:   "comment-" + id,
		Source:    "seeded by billing.SampleBillingLane",
		Priority:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// BillingSlotSchema is the stored shape of a BillingSlot.
var BillingSlotSchema = schema.New("billing.billingslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingSlot is one billing slot.
type BillingSlot struct {
	ID       string
	Group    string
	Label    string
	Owner    string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingSlot) Key() string { return "billing/billingslot/" + x.ID }

// Fields encodes x for storage.
func (x BillingSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"label":    x.Label,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingSlot decodes a stored record.
func ParseBillingSlot(f codec.Fields) (BillingSlot, error) {
	var x BillingSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingslot: missing id")
	}
	x.Group = f["group"]
	x.Label = f["label"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("billing: billingslot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("billing: billingslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingSlot) String() string {
	return fmt.Sprintf("BillingSlot(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingSlot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingSlot) Normalize() BillingSlot {
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortBillingSlots orders by priority descending, then by ID.
func SortBillingSlots(xs []BillingSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingSlots drops the archived entries, preserving order.
func ActiveBillingSlots(xs []BillingSlot) []BillingSlot {
	out := make([]BillingSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingSlots adds up the priority of every active entry.
func SumBillingSlots(xs []BillingSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexBillingSlots keys the entries by ID, last write winning.
func IndexBillingSlots(xs []BillingSlot) map[string]BillingSlot {
	out := make(map[string]BillingSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingSlot builds a valid BillingSlot, for tests and for the seed data the cron
// binary writes.
func SampleBillingSlot(id string, at time.Time) BillingSlot {
	return BillingSlot{
		ID:       id,
		Group:    "group-" + id,
		Label:    "label-" + id,
		Owner:    "seeded by billing.SampleBillingSlot",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingTagSchema is the stored shape of a BillingTag.
var BillingTagSchema = schema.New("billing.billingtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingTag is one billing tag.
type BillingTag struct {
	ID       string
	Category string
	Channel  string
	Owner    string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingTag) Key() string { return "billing/billingtag/" + x.ID }

// Fields encodes x for storage.
func (x BillingTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingTag decodes a stored record.
func ParseBillingTag(f codec.Fields) (BillingTag, error) {
	var x BillingTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingtag: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingtag %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("billing: billingtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingTag) String() string {
	return fmt.Sprintf("BillingTag(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingTag) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingTag) Normalize() BillingTag {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortBillingTags orders by weight descending, then by ID.
func SortBillingTags(xs []BillingTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingTags drops the archived entries, preserving order.
func ActiveBillingTags(xs []BillingTag) []BillingTag {
	out := make([]BillingTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingTags adds up the weight of every active entry.
func SumBillingTags(xs []BillingTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexBillingTags keys the entries by ID, last write winning.
func IndexBillingTags(xs []BillingTag) map[string]BillingTag {
	out := make(map[string]BillingTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingTag builds a valid BillingTag, for tests and for the seed data the cron
// binary writes.
func SampleBillingTag(id string, at time.Time) BillingTag {
	return BillingTag{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by billing.SampleBillingTag",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingRuleSchema is the stored shape of a BillingRule.
var BillingRuleSchema = schema.New("billing.billingrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingRule is one billing rule.
type BillingRule struct {
	ID        string
	Label     string
	Owner     string
	Reference string
	Version   int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x BillingRule) Key() string { return "billing/billingrule/" + x.ID }

// Fields encodes x for storage.
func (x BillingRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"owner":     x.Owner,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingRule decodes a stored record.
func ParseBillingRule(f codec.Fields) (BillingRule, error) {
	var x BillingRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingrule: missing id")
	}
	x.Label = f["label"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("billing: billingrule %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingRule) String() string {
	return fmt.Sprintf("BillingRule(%s, label=%q, version=%d, archived=%t)", x.ID, x.Label, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingRule) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingRule) Normalize() BillingRule {
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortBillingRules orders by version descending, then by ID.
func SortBillingRules(xs []BillingRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingRules drops the archived entries, preserving order.
func ActiveBillingRules(xs []BillingRule) []BillingRule {
	out := make([]BillingRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingRules adds up the version of every active entry.
func SumBillingRules(xs []BillingRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexBillingRules keys the entries by ID, last write winning.
func IndexBillingRules(xs []BillingRule) map[string]BillingRule {
	out := make(map[string]BillingRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingRule builds a valid BillingRule, for tests and for the seed data the cron
// binary writes.
func SampleBillingRule(id string, at time.Time) BillingRule {
	return BillingRule{
		ID:        id,
		Label:     "label-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by billing.SampleBillingRule",
		Version:   len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// BillingHoldSchema is the stored shape of a BillingHold.
var BillingHoldSchema = schema.New("billing.billinghold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingHold is one billing hold.
type BillingHold struct {
	ID       string
	Name     string
	Category string
	Owner    string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingHold) Key() string { return "billing/billinghold/" + x.ID }

// Fields encodes x for storage.
func (x BillingHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"owner":    x.Owner,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingHold decodes a stored record.
func ParseBillingHold(f codec.Fields) (BillingHold, error) {
	var x BillingHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billinghold: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("billing: billinghold %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billinghold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billinghold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billinghold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingHold) String() string {
	return fmt.Sprintf("BillingHold(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingHold) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingHold) Normalize() BillingHold {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortBillingHolds orders by attempts descending, then by ID.
func SortBillingHolds(xs []BillingHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingHolds drops the archived entries, preserving order.
func ActiveBillingHolds(xs []BillingHold) []BillingHold {
	out := make([]BillingHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingHolds adds up the attempts of every active entry.
func SumBillingHolds(xs []BillingHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexBillingHolds keys the entries by ID, last write winning.
func IndexBillingHolds(xs []BillingHold) map[string]BillingHold {
	out := make(map[string]BillingHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingHold builds a valid BillingHold, for tests and for the seed data the cron
// binary writes.
func SampleBillingHold(id string, at time.Time) BillingHold {
	return BillingHold{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Owner:    "seeded by billing.SampleBillingHold",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// BillingClaimSchema is the stored shape of a BillingClaim.
var BillingClaimSchema = schema.New("billing.billingclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingClaim is one billing claim.
type BillingClaim struct {
	ID        string
	Category  string
	Group     string
	Reference string
	Count     int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x BillingClaim) Key() string { return "billing/billingclaim/" + x.ID }

// Fields encodes x for storage.
func (x BillingClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"group":     x.Group,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingClaim decodes a stored record.
func ParseBillingClaim(f codec.Fields) (BillingClaim, error) {
	var x BillingClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingclaim: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("billing: billingclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingClaim) String() string {
	return fmt.Sprintf("BillingClaim(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingClaim) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingClaim) Normalize() BillingClaim {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortBillingClaims orders by count descending, then by ID.
func SortBillingClaims(xs []BillingClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingClaims drops the archived entries, preserving order.
func ActiveBillingClaims(xs []BillingClaim) []BillingClaim {
	out := make([]BillingClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingClaims adds up the count of every active entry.
func SumBillingClaims(xs []BillingClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexBillingClaims keys the entries by ID, last write winning.
func IndexBillingClaims(xs []BillingClaim) map[string]BillingClaim {
	out := make(map[string]BillingClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingClaim builds a valid BillingClaim, for tests and for the seed data the cron
// binary writes.
func SampleBillingClaim(id string, at time.Time) BillingClaim {
	return BillingClaim{
		ID:        id,
		Category:  "category-" + id,
		Group:     "group-" + id,
		Reference: "seeded by billing.SampleBillingClaim",
		Count:     len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// BillingCursorSchema is the stored shape of a BillingCursor.
var BillingCursorSchema = schema.New("billing.billingcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// BillingCursor is one billing cursor.
type BillingCursor struct {
	ID       string
	Category string
	Name     string
	Channel  string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x BillingCursor) Key() string { return "billing/billingcursor/" + x.ID }

// Fields encodes x for storage.
func (x BillingCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseBillingCursor decodes a stored record.
func ParseBillingCursor(f codec.Fields) (BillingCursor, error) {
	var x BillingCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("billing: billingcursor: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("billing: billingcursor %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("billing: billingcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("billing: billingcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("billing: billingcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x BillingCursor) String() string {
	return fmt.Sprintf("BillingCursor(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x BillingCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x BillingCursor) Normalize() BillingCursor {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortBillingCursors orders by weight descending, then by ID.
func SortBillingCursors(xs []BillingCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveBillingCursors drops the archived entries, preserving order.
func ActiveBillingCursors(xs []BillingCursor) []BillingCursor {
	out := make([]BillingCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumBillingCursors adds up the weight of every active entry.
func SumBillingCursors(xs []BillingCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexBillingCursors keys the entries by ID, last write winning.
func IndexBillingCursors(xs []BillingCursor) map[string]BillingCursor {
	out := make(map[string]BillingCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleBillingCursor builds a valid BillingCursor, for tests and for the seed data the cron
// binary writes.
func SampleBillingCursor(id string, at time.Time) BillingCursor {
	return BillingCursor{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by billing.SampleBillingCursor",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	BillingRecordSchema,
	BillingEntrySchema,
	BillingItemSchema,
	BillingDraftSchema,
	BillingRevisionSchema,
	BillingSnapshotSchema,
	BillingBatchSchema,
	BillingRequestSchema,
	BillingSummarySchema,
	BillingPolicySchema,
	BillingLinkSchema,
	BillingNoteSchema,
	BillingWindowSchema,
	BillingQuotaSchema,
	BillingBucketSchema,
	BillingLaneSchema,
	BillingSlotSchema,
	BillingTagSchema,
	BillingRuleSchema,
	BillingHoldSchema,
	BillingClaimSchema,
	BillingCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"billing/billingrecord/",
	"billing/billingentry/",
	"billing/billingitem/",
	"billing/billingdraft/",
	"billing/billingrevision/",
	"billing/billingsnapshot/",
	"billing/billingbatch/",
	"billing/billingrequest/",
	"billing/billingsummary/",
	"billing/billingpolicy/",
	"billing/billinglink/",
	"billing/billingnote/",
	"billing/billingwindow/",
	"billing/billingquota/",
	"billing/billingbucket/",
	"billing/billinglane/",
	"billing/billingslot/",
	"billing/billingtag/",
	"billing/billingrule/",
	"billing/billinghold/",
	"billing/billingclaim/",
	"billing/billingcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "billing: " + strings.Join(names, ", ")
}
