// Package carrier holds the carrier domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package carrier

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// CarrierRecordSchema is the stored shape of a CarrierRecord.
var CarrierRecordSchema = schema.New("carrier.carrierrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierRecord is one carrier record.
type CarrierRecord struct {
	ID       string
	Comment  string
	Label    string
	Group    string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierRecord) Key() string { return "carrier/carrierrecord/" + x.ID }

// Fields encodes x for storage.
func (x CarrierRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"label":    x.Label,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierRecord decodes a stored record.
func ParseCarrierRecord(f codec.Fields) (CarrierRecord, error) {
	var x CarrierRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierrecord: missing id")
	}
	x.Comment = f["comment"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierRecord) String() string {
	return fmt.Sprintf("CarrierRecord(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierRecord) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierRecord) Normalize() CarrierRecord {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCarrierRecords orders by priority descending, then by ID.
func SortCarrierRecords(xs []CarrierRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierRecords drops the archived entries, preserving order.
func ActiveCarrierRecords(xs []CarrierRecord) []CarrierRecord {
	out := make([]CarrierRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierRecords adds up the priority of every active entry.
func SumCarrierRecords(xs []CarrierRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexCarrierRecords keys the entries by ID, last write winning.
func IndexCarrierRecords(xs []CarrierRecord) map[string]CarrierRecord {
	out := make(map[string]CarrierRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierRecord builds a valid CarrierRecord, for tests and for the seed data the cron
// binary writes.
func SampleCarrierRecord(id string, at time.Time) CarrierRecord {
	return CarrierRecord{
		ID:       id,
		Comment:  "comment-" + id,
		Label:    "label-" + id,
		Group:    "seeded by carrier.SampleCarrierRecord",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierEntrySchema is the stored shape of a CarrierEntry.
var CarrierEntrySchema = schema.New("carrier.carrierentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierEntry is one carrier entry.
type CarrierEntry struct {
	ID       string
	Channel  string
	Source   string
	Owner    string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierEntry) Key() string { return "carrier/carrierentry/" + x.ID }

// Fields encodes x for storage.
func (x CarrierEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"source":   x.Source,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierEntry decodes a stored record.
func ParseCarrierEntry(f codec.Fields) (CarrierEntry, error) {
	var x CarrierEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierentry: missing id")
	}
	x.Channel = f["channel"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierentry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierEntry) String() string {
	return fmt.Sprintf("CarrierEntry(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierEntry) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierEntry) Normalize() CarrierEntry {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortCarrierEntrys orders by capacity descending, then by ID.
func SortCarrierEntrys(xs []CarrierEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierEntrys drops the archived entries, preserving order.
func ActiveCarrierEntrys(xs []CarrierEntry) []CarrierEntry {
	out := make([]CarrierEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierEntrys adds up the capacity of every active entry.
func SumCarrierEntrys(xs []CarrierEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCarrierEntrys keys the entries by ID, last write winning.
func IndexCarrierEntrys(xs []CarrierEntry) map[string]CarrierEntry {
	out := make(map[string]CarrierEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierEntry builds a valid CarrierEntry, for tests and for the seed data the cron
// binary writes.
func SampleCarrierEntry(id string, at time.Time) CarrierEntry {
	return CarrierEntry{
		ID:       id,
		Channel:  "channel-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by carrier.SampleCarrierEntry",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierItemSchema is the stored shape of a CarrierItem.
var CarrierItemSchema = schema.New("carrier.carrieritem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierItem is one carrier item.
type CarrierItem struct {
	ID       string
	Owner    string
	Comment  string
	Region   string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierItem) Key() string { return "carrier/carrieritem/" + x.ID }

// Fields encodes x for storage.
func (x CarrierItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"region":   x.Region,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierItem decodes a stored record.
func ParseCarrierItem(f codec.Fields) (CarrierItem, error) {
	var x CarrierItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrieritem: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("carrier: carrieritem %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("carrier: carrieritem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrieritem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrieritem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierItem) String() string {
	return fmt.Sprintf("CarrierItem(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierItem) Normalize() CarrierItem {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortCarrierItems orders by version descending, then by ID.
func SortCarrierItems(xs []CarrierItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierItems drops the archived entries, preserving order.
func ActiveCarrierItems(xs []CarrierItem) []CarrierItem {
	out := make([]CarrierItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierItems adds up the version of every active entry.
func SumCarrierItems(xs []CarrierItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexCarrierItems keys the entries by ID, last write winning.
func IndexCarrierItems(xs []CarrierItem) map[string]CarrierItem {
	out := make(map[string]CarrierItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierItem builds a valid CarrierItem, for tests and for the seed data the cron
// binary writes.
func SampleCarrierItem(id string, at time.Time) CarrierItem {
	return CarrierItem{
		ID:      id,
		Owner:   "owner-" + id,
		Comment: "comment-" + id,
		Region:  "seeded by carrier.SampleCarrierItem",
		Version: len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// CarrierDraftSchema is the stored shape of a CarrierDraft.
var CarrierDraftSchema = schema.New("carrier.carrierdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierDraft is one carrier draft.
type CarrierDraft struct {
	ID       string
	Group    string
	Category string
	Source   string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierDraft) Key() string { return "carrier/carrierdraft/" + x.ID }

// Fields encodes x for storage.
func (x CarrierDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierDraft decodes a stored record.
func ParseCarrierDraft(f codec.Fields) (CarrierDraft, error) {
	var x CarrierDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierdraft: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierdraft %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierDraft) String() string {
	return fmt.Sprintf("CarrierDraft(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierDraft) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierDraft) Normalize() CarrierDraft {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortCarrierDrafts orders by version descending, then by ID.
func SortCarrierDrafts(xs []CarrierDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierDrafts drops the archived entries, preserving order.
func ActiveCarrierDrafts(xs []CarrierDraft) []CarrierDraft {
	out := make([]CarrierDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierDrafts adds up the version of every active entry.
func SumCarrierDrafts(xs []CarrierDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexCarrierDrafts keys the entries by ID, last write winning.
func IndexCarrierDrafts(xs []CarrierDraft) map[string]CarrierDraft {
	out := make(map[string]CarrierDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierDraft builds a valid CarrierDraft, for tests and for the seed data the cron
// binary writes.
func SampleCarrierDraft(id string, at time.Time) CarrierDraft {
	return CarrierDraft{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Source:   "seeded by carrier.SampleCarrierDraft",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierRevisionSchema is the stored shape of a CarrierRevision.
var CarrierRevisionSchema = schema.New("carrier.carrierrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierRevision is one carrier revision.
type CarrierRevision struct {
	ID       string
	Region   string
	Channel  string
	Group    string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierRevision) Key() string { return "carrier/carrierrevision/" + x.ID }

// Fields encodes x for storage.
func (x CarrierRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierRevision decodes a stored record.
func ParseCarrierRevision(f codec.Fields) (CarrierRevision, error) {
	var x CarrierRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierrevision: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrevision %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierRevision) String() string {
	return fmt.Sprintf("CarrierRevision(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierRevision) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierRevision) Normalize() CarrierRevision {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCarrierRevisions orders by count descending, then by ID.
func SortCarrierRevisions(xs []CarrierRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierRevisions drops the archived entries, preserving order.
func ActiveCarrierRevisions(xs []CarrierRevision) []CarrierRevision {
	out := make([]CarrierRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierRevisions adds up the count of every active entry.
func SumCarrierRevisions(xs []CarrierRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCarrierRevisions keys the entries by ID, last write winning.
func IndexCarrierRevisions(xs []CarrierRevision) map[string]CarrierRevision {
	out := make(map[string]CarrierRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierRevision builds a valid CarrierRevision, for tests and for the seed data the cron
// binary writes.
func SampleCarrierRevision(id string, at time.Time) CarrierRevision {
	return CarrierRevision{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Group:    "seeded by carrier.SampleCarrierRevision",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierSnapshotSchema is the stored shape of a CarrierSnapshot.
var CarrierSnapshotSchema = schema.New("carrier.carriersnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierSnapshot is one carrier snapshot.
type CarrierSnapshot struct {
	ID        string
	Reference string
	Label     string
	Name      string
	Count     int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CarrierSnapshot) Key() string { return "carrier/carriersnapshot/" + x.ID }

// Fields encodes x for storage.
func (x CarrierSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"name":      x.Name,
		"count":     strconv.Itoa(x.Count),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierSnapshot decodes a stored record.
func ParseCarrierSnapshot(f codec.Fields) (CarrierSnapshot, error) {
	var x CarrierSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carriersnapshot: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("carrier: carriersnapshot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("carrier: carriersnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carriersnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carriersnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierSnapshot) String() string {
	return fmt.Sprintf("CarrierSnapshot(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierSnapshot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierSnapshot) Normalize() CarrierSnapshot {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCarrierSnapshots orders by count descending, then by ID.
func SortCarrierSnapshots(xs []CarrierSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierSnapshots drops the archived entries, preserving order.
func ActiveCarrierSnapshots(xs []CarrierSnapshot) []CarrierSnapshot {
	out := make([]CarrierSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierSnapshots adds up the count of every active entry.
func SumCarrierSnapshots(xs []CarrierSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCarrierSnapshots keys the entries by ID, last write winning.
func IndexCarrierSnapshots(xs []CarrierSnapshot) map[string]CarrierSnapshot {
	out := make(map[string]CarrierSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierSnapshot builds a valid CarrierSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleCarrierSnapshot(id string, at time.Time) CarrierSnapshot {
	return CarrierSnapshot{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Name:      "seeded by carrier.SampleCarrierSnapshot",
		Count:     len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CarrierBatchSchema is the stored shape of a CarrierBatch.
var CarrierBatchSchema = schema.New("carrier.carrierbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierBatch is one carrier batch.
type CarrierBatch struct {
	ID       string
	Region   string
	Name     string
	Comment  string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierBatch) Key() string { return "carrier/carrierbatch/" + x.ID }

// Fields encodes x for storage.
func (x CarrierBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierBatch decodes a stored record.
func ParseCarrierBatch(f codec.Fields) (CarrierBatch, error) {
	var x CarrierBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierbatch: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierbatch %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierBatch) String() string {
	return fmt.Sprintf("CarrierBatch(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierBatch) Normalize() CarrierBatch {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortCarrierBatchs orders by capacity descending, then by ID.
func SortCarrierBatchs(xs []CarrierBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierBatchs drops the archived entries, preserving order.
func ActiveCarrierBatchs(xs []CarrierBatch) []CarrierBatch {
	out := make([]CarrierBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierBatchs adds up the capacity of every active entry.
func SumCarrierBatchs(xs []CarrierBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCarrierBatchs keys the entries by ID, last write winning.
func IndexCarrierBatchs(xs []CarrierBatch) map[string]CarrierBatch {
	out := make(map[string]CarrierBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierBatch builds a valid CarrierBatch, for tests and for the seed data the cron
// binary writes.
func SampleCarrierBatch(id string, at time.Time) CarrierBatch {
	return CarrierBatch{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by carrier.SampleCarrierBatch",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierRequestSchema is the stored shape of a CarrierRequest.
var CarrierRequestSchema = schema.New("carrier.carrierrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierRequest is one carrier request.
type CarrierRequest struct {
	ID       string
	Group    string
	Channel  string
	Owner    string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierRequest) Key() string { return "carrier/carrierrequest/" + x.ID }

// Fields encodes x for storage.
func (x CarrierRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierRequest decodes a stored record.
func ParseCarrierRequest(f codec.Fields) (CarrierRequest, error) {
	var x CarrierRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierrequest: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierRequest) String() string {
	return fmt.Sprintf("CarrierRequest(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierRequest) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierRequest) Normalize() CarrierRequest {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortCarrierRequests orders by version descending, then by ID.
func SortCarrierRequests(xs []CarrierRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierRequests drops the archived entries, preserving order.
func ActiveCarrierRequests(xs []CarrierRequest) []CarrierRequest {
	out := make([]CarrierRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierRequests adds up the version of every active entry.
func SumCarrierRequests(xs []CarrierRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexCarrierRequests keys the entries by ID, last write winning.
func IndexCarrierRequests(xs []CarrierRequest) map[string]CarrierRequest {
	out := make(map[string]CarrierRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierRequest builds a valid CarrierRequest, for tests and for the seed data the cron
// binary writes.
func SampleCarrierRequest(id string, at time.Time) CarrierRequest {
	return CarrierRequest{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by carrier.SampleCarrierRequest",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierSummarySchema is the stored shape of a CarrierSummary.
var CarrierSummarySchema = schema.New("carrier.carriersummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierSummary is one carrier summary.
type CarrierSummary struct {
	ID       string
	Group    string
	Comment  string
	Name     string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierSummary) Key() string { return "carrier/carriersummary/" + x.ID }

// Fields encodes x for storage.
func (x CarrierSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"comment":  x.Comment,
		"name":     x.Name,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierSummary decodes a stored record.
func ParseCarrierSummary(f codec.Fields) (CarrierSummary, error) {
	var x CarrierSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carriersummary: missing id")
	}
	x.Group = f["group"]
	x.Comment = f["comment"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("carrier: carriersummary %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("carrier: carriersummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carriersummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carriersummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierSummary) String() string {
	return fmt.Sprintf("CarrierSummary(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierSummary) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierSummary) Normalize() CarrierSummary {
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortCarrierSummarys orders by offset descending, then by ID.
func SortCarrierSummarys(xs []CarrierSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierSummarys drops the archived entries, preserving order.
func ActiveCarrierSummarys(xs []CarrierSummary) []CarrierSummary {
	out := make([]CarrierSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierSummarys adds up the offset of every active entry.
func SumCarrierSummarys(xs []CarrierSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexCarrierSummarys keys the entries by ID, last write winning.
func IndexCarrierSummarys(xs []CarrierSummary) map[string]CarrierSummary {
	out := make(map[string]CarrierSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierSummary builds a valid CarrierSummary, for tests and for the seed data the cron
// binary writes.
func SampleCarrierSummary(id string, at time.Time) CarrierSummary {
	return CarrierSummary{
		ID:      id,
		Group:   "group-" + id,
		Comment: "comment-" + id,
		Name:    "seeded by carrier.SampleCarrierSummary",
		Offset:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// CarrierPolicySchema is the stored shape of a CarrierPolicy.
var CarrierPolicySchema = schema.New("carrier.carrierpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierPolicy is one carrier policy.
type CarrierPolicy struct {
	ID       string
	Category string
	Name     string
	Region   string
	Weight   int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierPolicy) Key() string { return "carrier/carrierpolicy/" + x.ID }

// Fields encodes x for storage.
func (x CarrierPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierPolicy decodes a stored record.
func ParseCarrierPolicy(f codec.Fields) (CarrierPolicy, error) {
	var x CarrierPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierpolicy: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierpolicy %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierPolicy) String() string {
	return fmt.Sprintf("CarrierPolicy(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierPolicy) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierPolicy) Normalize() CarrierPolicy {
	x.Category = strings.TrimSpace(x.Category)
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

// SortCarrierPolicys orders by weight descending, then by ID.
func SortCarrierPolicys(xs []CarrierPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierPolicys drops the archived entries, preserving order.
func ActiveCarrierPolicys(xs []CarrierPolicy) []CarrierPolicy {
	out := make([]CarrierPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierPolicys adds up the weight of every active entry.
func SumCarrierPolicys(xs []CarrierPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexCarrierPolicys keys the entries by ID, last write winning.
func IndexCarrierPolicys(xs []CarrierPolicy) map[string]CarrierPolicy {
	out := make(map[string]CarrierPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierPolicy builds a valid CarrierPolicy, for tests and for the seed data the cron
// binary writes.
func SampleCarrierPolicy(id string, at time.Time) CarrierPolicy {
	return CarrierPolicy{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Region:   "seeded by carrier.SampleCarrierPolicy",
		Weight:   len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierLinkSchema is the stored shape of a CarrierLink.
var CarrierLinkSchema = schema.New("carrier.carrierlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierLink is one carrier link.
type CarrierLink struct {
	ID       string
	Channel  string
	Label    string
	Group    string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierLink) Key() string { return "carrier/carrierlink/" + x.ID }

// Fields encodes x for storage.
func (x CarrierLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierLink decodes a stored record.
func ParseCarrierLink(f codec.Fields) (CarrierLink, error) {
	var x CarrierLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierlink: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierlink %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierLink) String() string {
	return fmt.Sprintf("CarrierLink(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierLink) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierLink) Normalize() CarrierLink {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCarrierLinks orders by priority descending, then by ID.
func SortCarrierLinks(xs []CarrierLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierLinks drops the archived entries, preserving order.
func ActiveCarrierLinks(xs []CarrierLink) []CarrierLink {
	out := make([]CarrierLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierLinks adds up the priority of every active entry.
func SumCarrierLinks(xs []CarrierLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexCarrierLinks keys the entries by ID, last write winning.
func IndexCarrierLinks(xs []CarrierLink) map[string]CarrierLink {
	out := make(map[string]CarrierLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierLink builds a valid CarrierLink, for tests and for the seed data the cron
// binary writes.
func SampleCarrierLink(id string, at time.Time) CarrierLink {
	return CarrierLink{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Group:    "seeded by carrier.SampleCarrierLink",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierNoteSchema is the stored shape of a CarrierNote.
var CarrierNoteSchema = schema.New("carrier.carriernote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierNote is one carrier note.
type CarrierNote struct {
	ID        string
	Region    string
	Reference string
	Label     string
	Priority  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CarrierNote) Key() string { return "carrier/carriernote/" + x.ID }

// Fields encodes x for storage.
func (x CarrierNote) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"label":     x.Label,
		"priority":  strconv.Itoa(x.Priority),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierNote decodes a stored record.
func ParseCarrierNote(f codec.Fields) (CarrierNote, error) {
	var x CarrierNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carriernote: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carriernote %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("carrier: carriernote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carriernote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carriernote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierNote) String() string {
	return fmt.Sprintf("CarrierNote(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierNote) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierNote) Normalize() CarrierNote {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortCarrierNotes orders by priority descending, then by ID.
func SortCarrierNotes(xs []CarrierNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierNotes drops the archived entries, preserving order.
func ActiveCarrierNotes(xs []CarrierNote) []CarrierNote {
	out := make([]CarrierNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierNotes adds up the priority of every active entry.
func SumCarrierNotes(xs []CarrierNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexCarrierNotes keys the entries by ID, last write winning.
func IndexCarrierNotes(xs []CarrierNote) map[string]CarrierNote {
	out := make(map[string]CarrierNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierNote builds a valid CarrierNote, for tests and for the seed data the cron
// binary writes.
func SampleCarrierNote(id string, at time.Time) CarrierNote {
	return CarrierNote{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Label:     "seeded by carrier.SampleCarrierNote",
		Priority:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CarrierWindowSchema is the stored shape of a CarrierWindow.
var CarrierWindowSchema = schema.New("carrier.carrierwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierWindow is one carrier window.
type CarrierWindow struct {
	ID       string
	Source   string
	Name     string
	Comment  string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierWindow) Key() string { return "carrier/carrierwindow/" + x.ID }

// Fields encodes x for storage.
func (x CarrierWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierWindow decodes a stored record.
func ParseCarrierWindow(f codec.Fields) (CarrierWindow, error) {
	var x CarrierWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierwindow: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierwindow %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierWindow) String() string {
	return fmt.Sprintf("CarrierWindow(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierWindow) Normalize() CarrierWindow {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortCarrierWindows orders by priority descending, then by ID.
func SortCarrierWindows(xs []CarrierWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierWindows drops the archived entries, preserving order.
func ActiveCarrierWindows(xs []CarrierWindow) []CarrierWindow {
	out := make([]CarrierWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierWindows adds up the priority of every active entry.
func SumCarrierWindows(xs []CarrierWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexCarrierWindows keys the entries by ID, last write winning.
func IndexCarrierWindows(xs []CarrierWindow) map[string]CarrierWindow {
	out := make(map[string]CarrierWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierWindow builds a valid CarrierWindow, for tests and for the seed data the cron
// binary writes.
func SampleCarrierWindow(id string, at time.Time) CarrierWindow {
	return CarrierWindow{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by carrier.SampleCarrierWindow",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierQuotaSchema is the stored shape of a CarrierQuota.
var CarrierQuotaSchema = schema.New("carrier.carrierquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierQuota is one carrier quota.
type CarrierQuota struct {
	ID       string
	Owner    string
	Source   string
	Channel  string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierQuota) Key() string { return "carrier/carrierquota/" + x.ID }

// Fields encodes x for storage.
func (x CarrierQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"channel":  x.Channel,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierQuota decodes a stored record.
func ParseCarrierQuota(f codec.Fields) (CarrierQuota, error) {
	var x CarrierQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierquota: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierquota %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierQuota) String() string {
	return fmt.Sprintf("CarrierQuota(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierQuota) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierQuota) Normalize() CarrierQuota {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortCarrierQuotas orders by offset descending, then by ID.
func SortCarrierQuotas(xs []CarrierQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierQuotas drops the archived entries, preserving order.
func ActiveCarrierQuotas(xs []CarrierQuota) []CarrierQuota {
	out := make([]CarrierQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierQuotas adds up the offset of every active entry.
func SumCarrierQuotas(xs []CarrierQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexCarrierQuotas keys the entries by ID, last write winning.
func IndexCarrierQuotas(xs []CarrierQuota) map[string]CarrierQuota {
	out := make(map[string]CarrierQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierQuota builds a valid CarrierQuota, for tests and for the seed data the cron
// binary writes.
func SampleCarrierQuota(id string, at time.Time) CarrierQuota {
	return CarrierQuota{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by carrier.SampleCarrierQuota",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierBucketSchema is the stored shape of a CarrierBucket.
var CarrierBucketSchema = schema.New("carrier.carrierbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierBucket is one carrier bucket.
type CarrierBucket struct {
	ID       string
	Channel  string
	Source   string
	Region   string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierBucket) Key() string { return "carrier/carrierbucket/" + x.ID }

// Fields encodes x for storage.
func (x CarrierBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"source":   x.Source,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierBucket decodes a stored record.
func ParseCarrierBucket(f codec.Fields) (CarrierBucket, error) {
	var x CarrierBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierbucket: missing id")
	}
	x.Channel = f["channel"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierbucket %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierBucket) String() string {
	return fmt.Sprintf("CarrierBucket(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierBucket) Normalize() CarrierBucket {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortCarrierBuckets orders by retained descending, then by ID.
func SortCarrierBuckets(xs []CarrierBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierBuckets drops the archived entries, preserving order.
func ActiveCarrierBuckets(xs []CarrierBucket) []CarrierBucket {
	out := make([]CarrierBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierBuckets adds up the retained of every active entry.
func SumCarrierBuckets(xs []CarrierBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexCarrierBuckets keys the entries by ID, last write winning.
func IndexCarrierBuckets(xs []CarrierBucket) map[string]CarrierBucket {
	out := make(map[string]CarrierBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierBucket builds a valid CarrierBucket, for tests and for the seed data the cron
// binary writes.
func SampleCarrierBucket(id string, at time.Time) CarrierBucket {
	return CarrierBucket{
		ID:       id,
		Channel:  "channel-" + id,
		Source:   "source-" + id,
		Region:   "seeded by carrier.SampleCarrierBucket",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierLaneSchema is the stored shape of a CarrierLane.
var CarrierLaneSchema = schema.New("carrier.carrierlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierLane is one carrier lane.
type CarrierLane struct {
	ID        string
	Comment   string
	Region    string
	Reference string
	Capacity  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x CarrierLane) Key() string { return "carrier/carrierlane/" + x.ID }

// Fields encodes x for storage.
func (x CarrierLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"region":    x.Region,
		"reference": x.Reference,
		"capacity":  strconv.Itoa(x.Capacity),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierLane decodes a stored record.
func ParseCarrierLane(f codec.Fields) (CarrierLane, error) {
	var x CarrierLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierlane: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Reference = f["reference"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierlane %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierLane) String() string {
	return fmt.Sprintf("CarrierLane(%s, comment=%q, capacity=%d, archived=%t)", x.ID, x.Comment, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierLane) Normalize() CarrierLane {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortCarrierLanes orders by capacity descending, then by ID.
func SortCarrierLanes(xs []CarrierLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierLanes drops the archived entries, preserving order.
func ActiveCarrierLanes(xs []CarrierLane) []CarrierLane {
	out := make([]CarrierLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierLanes adds up the capacity of every active entry.
func SumCarrierLanes(xs []CarrierLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexCarrierLanes keys the entries by ID, last write winning.
func IndexCarrierLanes(xs []CarrierLane) map[string]CarrierLane {
	out := make(map[string]CarrierLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierLane builds a valid CarrierLane, for tests and for the seed data the cron
// binary writes.
func SampleCarrierLane(id string, at time.Time) CarrierLane {
	return CarrierLane{
		ID:        id,
		Comment:   "comment-" + id,
		Region:    "region-" + id,
		Reference: "seeded by carrier.SampleCarrierLane",
		Capacity:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// CarrierSlotSchema is the stored shape of a CarrierSlot.
var CarrierSlotSchema = schema.New("carrier.carrierslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierSlot is one carrier slot.
type CarrierSlot struct {
	ID       string
	Source   string
	Name     string
	Owner    string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierSlot) Key() string { return "carrier/carrierslot/" + x.ID }

// Fields encodes x for storage.
func (x CarrierSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"name":     x.Name,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierSlot decodes a stored record.
func ParseCarrierSlot(f codec.Fields) (CarrierSlot, error) {
	var x CarrierSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierslot: missing id")
	}
	x.Source = f["source"]
	x.Name = f["name"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierslot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierSlot) String() string {
	return fmt.Sprintf("CarrierSlot(%s, source=%q, offset=%d, archived=%t)", x.ID, x.Source, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierSlot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierSlot) Normalize() CarrierSlot {
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortCarrierSlots orders by offset descending, then by ID.
func SortCarrierSlots(xs []CarrierSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierSlots drops the archived entries, preserving order.
func ActiveCarrierSlots(xs []CarrierSlot) []CarrierSlot {
	out := make([]CarrierSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierSlots adds up the offset of every active entry.
func SumCarrierSlots(xs []CarrierSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexCarrierSlots keys the entries by ID, last write winning.
func IndexCarrierSlots(xs []CarrierSlot) map[string]CarrierSlot {
	out := make(map[string]CarrierSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierSlot builds a valid CarrierSlot, for tests and for the seed data the cron
// binary writes.
func SampleCarrierSlot(id string, at time.Time) CarrierSlot {
	return CarrierSlot{
		ID:       id,
		Source:   "source-" + id,
		Name:     "name-" + id,
		Owner:    "seeded by carrier.SampleCarrierSlot",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierTagSchema is the stored shape of a CarrierTag.
var CarrierTagSchema = schema.New("carrier.carriertag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierTag is one carrier tag.
type CarrierTag struct {
	ID       string
	Name     string
	Region   string
	Group    string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierTag) Key() string { return "carrier/carriertag/" + x.ID }

// Fields encodes x for storage.
func (x CarrierTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"region":   x.Region,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierTag decodes a stored record.
func ParseCarrierTag(f codec.Fields) (CarrierTag, error) {
	var x CarrierTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carriertag: missing id")
	}
	x.Name = f["name"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carriertag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("carrier: carriertag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carriertag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carriertag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierTag) String() string {
	return fmt.Sprintf("CarrierTag(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierTag) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierTag) Normalize() CarrierTag {
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortCarrierTags orders by priority descending, then by ID.
func SortCarrierTags(xs []CarrierTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierTags drops the archived entries, preserving order.
func ActiveCarrierTags(xs []CarrierTag) []CarrierTag {
	out := make([]CarrierTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierTags adds up the priority of every active entry.
func SumCarrierTags(xs []CarrierTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexCarrierTags keys the entries by ID, last write winning.
func IndexCarrierTags(xs []CarrierTag) map[string]CarrierTag {
	out := make(map[string]CarrierTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierTag builds a valid CarrierTag, for tests and for the seed data the cron
// binary writes.
func SampleCarrierTag(id string, at time.Time) CarrierTag {
	return CarrierTag{
		ID:       id,
		Name:     "name-" + id,
		Region:   "region-" + id,
		Group:    "seeded by carrier.SampleCarrierTag",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierRuleSchema is the stored shape of a CarrierRule.
var CarrierRuleSchema = schema.New("carrier.carrierrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierRule is one carrier rule.
type CarrierRule struct {
	ID       string
	Region   string
	Owner    string
	Label    string
	Weight   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierRule) Key() string { return "carrier/carrierrule/" + x.ID }

// Fields encodes x for storage.
func (x CarrierRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"owner":    x.Owner,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierRule decodes a stored record.
func ParseCarrierRule(f codec.Fields) (CarrierRule, error) {
	var x CarrierRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierrule: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrule %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierRule) String() string {
	return fmt.Sprintf("CarrierRule(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierRule) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierRule) Normalize() CarrierRule {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortCarrierRules orders by weight descending, then by ID.
func SortCarrierRules(xs []CarrierRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierRules drops the archived entries, preserving order.
func ActiveCarrierRules(xs []CarrierRule) []CarrierRule {
	out := make([]CarrierRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierRules adds up the weight of every active entry.
func SumCarrierRules(xs []CarrierRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexCarrierRules keys the entries by ID, last write winning.
func IndexCarrierRules(xs []CarrierRule) map[string]CarrierRule {
	out := make(map[string]CarrierRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierRule builds a valid CarrierRule, for tests and for the seed data the cron
// binary writes.
func SampleCarrierRule(id string, at time.Time) CarrierRule {
	return CarrierRule{
		ID:      id,
		Region:  "region-" + id,
		Owner:   "owner-" + id,
		Label:   "seeded by carrier.SampleCarrierRule",
		Weight:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// CarrierHoldSchema is the stored shape of a CarrierHold.
var CarrierHoldSchema = schema.New("carrier.carrierhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierHold is one carrier hold.
type CarrierHold struct {
	ID       string
	Category string
	Owner    string
	Source   string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierHold) Key() string { return "carrier/carrierhold/" + x.ID }

// Fields encodes x for storage.
func (x CarrierHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierHold decodes a stored record.
func ParseCarrierHold(f codec.Fields) (CarrierHold, error) {
	var x CarrierHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierhold: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierhold %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierHold) String() string {
	return fmt.Sprintf("CarrierHold(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierHold) Normalize() CarrierHold {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortCarrierHolds orders by weight descending, then by ID.
func SortCarrierHolds(xs []CarrierHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierHolds drops the archived entries, preserving order.
func ActiveCarrierHolds(xs []CarrierHold) []CarrierHold {
	out := make([]CarrierHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierHolds adds up the weight of every active entry.
func SumCarrierHolds(xs []CarrierHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexCarrierHolds keys the entries by ID, last write winning.
func IndexCarrierHolds(xs []CarrierHold) map[string]CarrierHold {
	out := make(map[string]CarrierHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierHold builds a valid CarrierHold, for tests and for the seed data the cron
// binary writes.
func SampleCarrierHold(id string, at time.Time) CarrierHold {
	return CarrierHold{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by carrier.SampleCarrierHold",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// CarrierClaimSchema is the stored shape of a CarrierClaim.
var CarrierClaimSchema = schema.New("carrier.carrierclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierClaim is one carrier claim.
type CarrierClaim struct {
	ID       string
	Owner    string
	Source   string
	Comment  string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierClaim) Key() string { return "carrier/carrierclaim/" + x.ID }

// Fields encodes x for storage.
func (x CarrierClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierClaim decodes a stored record.
func ParseCarrierClaim(f codec.Fields) (CarrierClaim, error) {
	var x CarrierClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carrierclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierclaim %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("carrier: carrierclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carrierclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carrierclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierClaim) String() string {
	return fmt.Sprintf("CarrierClaim(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierClaim) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierClaim) Normalize() CarrierClaim {
	x.Owner = strings.TrimSpace(x.Owner)
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

// SortCarrierClaims orders by count descending, then by ID.
func SortCarrierClaims(xs []CarrierClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierClaims drops the archived entries, preserving order.
func ActiveCarrierClaims(xs []CarrierClaim) []CarrierClaim {
	out := make([]CarrierClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierClaims adds up the count of every active entry.
func SumCarrierClaims(xs []CarrierClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexCarrierClaims keys the entries by ID, last write winning.
func IndexCarrierClaims(xs []CarrierClaim) map[string]CarrierClaim {
	out := make(map[string]CarrierClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierClaim builds a valid CarrierClaim, for tests and for the seed data the cron
// binary writes.
func SampleCarrierClaim(id string, at time.Time) CarrierClaim {
	return CarrierClaim{
		ID:      id,
		Owner:   "owner-" + id,
		Source:  "source-" + id,
		Comment: "seeded by carrier.SampleCarrierClaim",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// CarrierCursorSchema is the stored shape of a CarrierCursor.
var CarrierCursorSchema = schema.New("carrier.carriercursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// CarrierCursor is one carrier cursor.
type CarrierCursor struct {
	ID       string
	Channel  string
	Label    string
	Group    string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x CarrierCursor) Key() string { return "carrier/carriercursor/" + x.ID }

// Fields encodes x for storage.
func (x CarrierCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"group":    x.Group,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseCarrierCursor decodes a stored record.
func ParseCarrierCursor(f codec.Fields) (CarrierCursor, error) {
	var x CarrierCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("carrier: carriercursor: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("carrier: carriercursor %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("carrier: carriercursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("carrier: carriercursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("carrier: carriercursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x CarrierCursor) String() string {
	return fmt.Sprintf("CarrierCursor(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x CarrierCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x CarrierCursor) Normalize() CarrierCursor {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortCarrierCursors orders by weight descending, then by ID.
func SortCarrierCursors(xs []CarrierCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveCarrierCursors drops the archived entries, preserving order.
func ActiveCarrierCursors(xs []CarrierCursor) []CarrierCursor {
	out := make([]CarrierCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumCarrierCursors adds up the weight of every active entry.
func SumCarrierCursors(xs []CarrierCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexCarrierCursors keys the entries by ID, last write winning.
func IndexCarrierCursors(xs []CarrierCursor) map[string]CarrierCursor {
	out := make(map[string]CarrierCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleCarrierCursor builds a valid CarrierCursor, for tests and for the seed data the cron
// binary writes.
func SampleCarrierCursor(id string, at time.Time) CarrierCursor {
	return CarrierCursor{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Group:    "seeded by carrier.SampleCarrierCursor",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	CarrierRecordSchema,
	CarrierEntrySchema,
	CarrierItemSchema,
	CarrierDraftSchema,
	CarrierRevisionSchema,
	CarrierSnapshotSchema,
	CarrierBatchSchema,
	CarrierRequestSchema,
	CarrierSummarySchema,
	CarrierPolicySchema,
	CarrierLinkSchema,
	CarrierNoteSchema,
	CarrierWindowSchema,
	CarrierQuotaSchema,
	CarrierBucketSchema,
	CarrierLaneSchema,
	CarrierSlotSchema,
	CarrierTagSchema,
	CarrierRuleSchema,
	CarrierHoldSchema,
	CarrierClaimSchema,
	CarrierCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"carrier/carrierrecord/",
	"carrier/carrierentry/",
	"carrier/carrieritem/",
	"carrier/carrierdraft/",
	"carrier/carrierrevision/",
	"carrier/carriersnapshot/",
	"carrier/carrierbatch/",
	"carrier/carrierrequest/",
	"carrier/carriersummary/",
	"carrier/carrierpolicy/",
	"carrier/carrierlink/",
	"carrier/carriernote/",
	"carrier/carrierwindow/",
	"carrier/carrierquota/",
	"carrier/carrierbucket/",
	"carrier/carrierlane/",
	"carrier/carrierslot/",
	"carrier/carriertag/",
	"carrier/carrierrule/",
	"carrier/carrierhold/",
	"carrier/carrierclaim/",
	"carrier/carriercursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "carrier: " + strings.Join(names, ", ")
}
