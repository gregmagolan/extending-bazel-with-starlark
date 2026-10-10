// Package refund holds the refund domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package refund

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// RefundRecordSchema is the stored shape of a RefundRecord.
var RefundRecordSchema = schema.New("refund.refundrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundRecord is one refund record.
type RefundRecord struct {
	ID       string
	Category string
	Label    string
	Group    string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundRecord) Key() string { return "refund/refundrecord/" + x.ID }

// Fields encodes x for storage.
func (x RefundRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundRecord decodes a stored record.
func ParseRefundRecord(f codec.Fields) (RefundRecord, error) {
	var x RefundRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundrecord: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("refund: refundrecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("refund: refundrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundRecord) String() string {
	return fmt.Sprintf("RefundRecord(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundRecord) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundRecord) Normalize() RefundRecord {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRefundRecords orders by version descending, then by ID.
func SortRefundRecords(xs []RefundRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundRecords drops the archived entries, preserving order.
func ActiveRefundRecords(xs []RefundRecord) []RefundRecord {
	out := make([]RefundRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundRecords adds up the version of every active entry.
func SumRefundRecords(xs []RefundRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexRefundRecords keys the entries by ID, last write winning.
func IndexRefundRecords(xs []RefundRecord) map[string]RefundRecord {
	out := make(map[string]RefundRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundRecord builds a valid RefundRecord, for tests and for the seed data the cron
// binary writes.
func SampleRefundRecord(id string, at time.Time) RefundRecord {
	return RefundRecord{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Group:    "seeded by refund.SampleRefundRecord",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundEntrySchema is the stored shape of a RefundEntry.
var RefundEntrySchema = schema.New("refund.refundentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundEntry is one refund entry.
type RefundEntry struct {
	ID       string
	Owner    string
	Source   string
	Label    string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundEntry) Key() string { return "refund/refundentry/" + x.ID }

// Fields encodes x for storage.
func (x RefundEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundEntry decodes a stored record.
func ParseRefundEntry(f codec.Fields) (RefundEntry, error) {
	var x RefundEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundentry: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("refund: refundentry %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("refund: refundentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundEntry) String() string {
	return fmt.Sprintf("RefundEntry(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundEntry) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundEntry) Normalize() RefundEntry {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRefundEntrys orders by priority descending, then by ID.
func SortRefundEntrys(xs []RefundEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundEntrys drops the archived entries, preserving order.
func ActiveRefundEntrys(xs []RefundEntry) []RefundEntry {
	out := make([]RefundEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundEntrys adds up the priority of every active entry.
func SumRefundEntrys(xs []RefundEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRefundEntrys keys the entries by ID, last write winning.
func IndexRefundEntrys(xs []RefundEntry) map[string]RefundEntry {
	out := make(map[string]RefundEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundEntry builds a valid RefundEntry, for tests and for the seed data the cron
// binary writes.
func SampleRefundEntry(id string, at time.Time) RefundEntry {
	return RefundEntry{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Label:    "seeded by refund.SampleRefundEntry",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundItemSchema is the stored shape of a RefundItem.
var RefundItemSchema = schema.New("refund.refunditem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundItem is one refund item.
type RefundItem struct {
	ID       string
	Category string
	Channel  string
	Comment  string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundItem) Key() string { return "refund/refunditem/" + x.ID }

// Fields encodes x for storage.
func (x RefundItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundItem decodes a stored record.
func ParseRefundItem(f codec.Fields) (RefundItem, error) {
	var x RefundItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refunditem: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("refund: refunditem %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refunditem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refunditem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refunditem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundItem) String() string {
	return fmt.Sprintf("RefundItem(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundItem) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundItem) Normalize() RefundItem {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRefundItems orders by offset descending, then by ID.
func SortRefundItems(xs []RefundItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundItems drops the archived entries, preserving order.
func ActiveRefundItems(xs []RefundItem) []RefundItem {
	out := make([]RefundItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundItems adds up the offset of every active entry.
func SumRefundItems(xs []RefundItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRefundItems keys the entries by ID, last write winning.
func IndexRefundItems(xs []RefundItem) map[string]RefundItem {
	out := make(map[string]RefundItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundItem builds a valid RefundItem, for tests and for the seed data the cron
// binary writes.
func SampleRefundItem(id string, at time.Time) RefundItem {
	return RefundItem{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by refund.SampleRefundItem",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundDraftSchema is the stored shape of a RefundDraft.
var RefundDraftSchema = schema.New("refund.refunddraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundDraft is one refund draft.
type RefundDraft struct {
	ID       string
	Owner    string
	Channel  string
	Region   string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundDraft) Key() string { return "refund/refunddraft/" + x.ID }

// Fields encodes x for storage.
func (x RefundDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundDraft decodes a stored record.
func ParseRefundDraft(f codec.Fields) (RefundDraft, error) {
	var x RefundDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refunddraft: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refunddraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refunddraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refunddraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refunddraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundDraft) String() string {
	return fmt.Sprintf("RefundDraft(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundDraft) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundDraft) Normalize() RefundDraft {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRefundDrafts orders by capacity descending, then by ID.
func SortRefundDrafts(xs []RefundDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundDrafts drops the archived entries, preserving order.
func ActiveRefundDrafts(xs []RefundDraft) []RefundDraft {
	out := make([]RefundDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundDrafts adds up the capacity of every active entry.
func SumRefundDrafts(xs []RefundDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRefundDrafts keys the entries by ID, last write winning.
func IndexRefundDrafts(xs []RefundDraft) map[string]RefundDraft {
	out := make(map[string]RefundDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundDraft builds a valid RefundDraft, for tests and for the seed data the cron
// binary writes.
func SampleRefundDraft(id string, at time.Time) RefundDraft {
	return RefundDraft{
		ID:       id,
		Owner:    "owner-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by refund.SampleRefundDraft",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundRevisionSchema is the stored shape of a RefundRevision.
var RefundRevisionSchema = schema.New("refund.refundrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundRevision is one refund revision.
type RefundRevision struct {
	ID       string
	Region   string
	Name     string
	Owner    string
	Retained int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundRevision) Key() string { return "refund/refundrevision/" + x.ID }

// Fields encodes x for storage.
func (x RefundRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundRevision decodes a stored record.
func ParseRefundRevision(f codec.Fields) (RefundRevision, error) {
	var x RefundRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundrevision: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("refund: refundrevision %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refundrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundRevision) String() string {
	return fmt.Sprintf("RefundRevision(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundRevision) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundRevision) Normalize() RefundRevision {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRefundRevisions orders by retained descending, then by ID.
func SortRefundRevisions(xs []RefundRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundRevisions drops the archived entries, preserving order.
func ActiveRefundRevisions(xs []RefundRevision) []RefundRevision {
	out := make([]RefundRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundRevisions adds up the retained of every active entry.
func SumRefundRevisions(xs []RefundRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRefundRevisions keys the entries by ID, last write winning.
func IndexRefundRevisions(xs []RefundRevision) map[string]RefundRevision {
	out := make(map[string]RefundRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundRevision builds a valid RefundRevision, for tests and for the seed data the cron
// binary writes.
func SampleRefundRevision(id string, at time.Time) RefundRevision {
	return RefundRevision{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Owner:    "seeded by refund.SampleRefundRevision",
		Retained: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundSnapshotSchema is the stored shape of a RefundSnapshot.
var RefundSnapshotSchema = schema.New("refund.refundsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundSnapshot is one refund snapshot.
type RefundSnapshot struct {
	ID        string
	Reference string
	Group     string
	Channel   string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RefundSnapshot) Key() string { return "refund/refundsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x RefundSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"channel":   x.Channel,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundSnapshot decodes a stored record.
func ParseRefundSnapshot(f codec.Fields) (RefundSnapshot, error) {
	var x RefundSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundsnapshot: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("refund: refundsnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refundsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundSnapshot) String() string {
	return fmt.Sprintf("RefundSnapshot(%s, reference=%q, offset=%d, archived=%t)", x.ID, x.Reference, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundSnapshot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundSnapshot) Normalize() RefundSnapshot {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRefundSnapshots orders by offset descending, then by ID.
func SortRefundSnapshots(xs []RefundSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundSnapshots drops the archived entries, preserving order.
func ActiveRefundSnapshots(xs []RefundSnapshot) []RefundSnapshot {
	out := make([]RefundSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundSnapshots adds up the offset of every active entry.
func SumRefundSnapshots(xs []RefundSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRefundSnapshots keys the entries by ID, last write winning.
func IndexRefundSnapshots(xs []RefundSnapshot) map[string]RefundSnapshot {
	out := make(map[string]RefundSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundSnapshot builds a valid RefundSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleRefundSnapshot(id string, at time.Time) RefundSnapshot {
	return RefundSnapshot{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Channel:   "seeded by refund.SampleRefundSnapshot",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RefundBatchSchema is the stored shape of a RefundBatch.
var RefundBatchSchema = schema.New("refund.refundbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundBatch is one refund batch.
type RefundBatch struct {
	ID       string
	Group    string
	Region   string
	Name     string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundBatch) Key() string { return "refund/refundbatch/" + x.ID }

// Fields encodes x for storage.
func (x RefundBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"region":   x.Region,
		"name":     x.Name,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundBatch decodes a stored record.
func ParseRefundBatch(f codec.Fields) (RefundBatch, error) {
	var x RefundBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundbatch: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("refund: refundbatch %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("refund: refundbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundBatch) String() string {
	return fmt.Sprintf("RefundBatch(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundBatch) Normalize() RefundBatch {
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRefundBatchs orders by version descending, then by ID.
func SortRefundBatchs(xs []RefundBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundBatchs drops the archived entries, preserving order.
func ActiveRefundBatchs(xs []RefundBatch) []RefundBatch {
	out := make([]RefundBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundBatchs adds up the version of every active entry.
func SumRefundBatchs(xs []RefundBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexRefundBatchs keys the entries by ID, last write winning.
func IndexRefundBatchs(xs []RefundBatch) map[string]RefundBatch {
	out := make(map[string]RefundBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundBatch builds a valid RefundBatch, for tests and for the seed data the cron
// binary writes.
func SampleRefundBatch(id string, at time.Time) RefundBatch {
	return RefundBatch{
		ID:       id,
		Group:    "group-" + id,
		Region:   "region-" + id,
		Name:     "seeded by refund.SampleRefundBatch",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundRequestSchema is the stored shape of a RefundRequest.
var RefundRequestSchema = schema.New("refund.refundrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundRequest is one refund request.
type RefundRequest struct {
	ID       string
	Region   string
	Comment  string
	Label    string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundRequest) Key() string { return "refund/refundrequest/" + x.ID }

// Fields encodes x for storage.
func (x RefundRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundRequest decodes a stored record.
func ParseRefundRequest(f codec.Fields) (RefundRequest, error) {
	var x RefundRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundrequest: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refundrequest %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("refund: refundrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundRequest) String() string {
	return fmt.Sprintf("RefundRequest(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundRequest) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundRequest) Normalize() RefundRequest {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRefundRequests orders by capacity descending, then by ID.
func SortRefundRequests(xs []RefundRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundRequests drops the archived entries, preserving order.
func ActiveRefundRequests(xs []RefundRequest) []RefundRequest {
	out := make([]RefundRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundRequests adds up the capacity of every active entry.
func SumRefundRequests(xs []RefundRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRefundRequests keys the entries by ID, last write winning.
func IndexRefundRequests(xs []RefundRequest) map[string]RefundRequest {
	out := make(map[string]RefundRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundRequest builds a valid RefundRequest, for tests and for the seed data the cron
// binary writes.
func SampleRefundRequest(id string, at time.Time) RefundRequest {
	return RefundRequest{
		ID:       id,
		Region:   "region-" + id,
		Comment:  "comment-" + id,
		Label:    "seeded by refund.SampleRefundRequest",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundSummarySchema is the stored shape of a RefundSummary.
var RefundSummarySchema = schema.New("refund.refundsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundSummary is one refund summary.
type RefundSummary struct {
	ID        string
	Source    string
	Label     string
	Reference string
	Retained  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RefundSummary) Key() string { return "refund/refundsummary/" + x.ID }

// Fields encodes x for storage.
func (x RefundSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"label":     x.Label,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundSummary decodes a stored record.
func ParseRefundSummary(f codec.Fields) (RefundSummary, error) {
	var x RefundSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundsummary: missing id")
	}
	x.Source = f["source"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("refund: refundsummary %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("refund: refundsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundSummary) String() string {
	return fmt.Sprintf("RefundSummary(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundSummary) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundSummary) Normalize() RefundSummary {
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRefundSummarys orders by retained descending, then by ID.
func SortRefundSummarys(xs []RefundSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundSummarys drops the archived entries, preserving order.
func ActiveRefundSummarys(xs []RefundSummary) []RefundSummary {
	out := make([]RefundSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundSummarys adds up the retained of every active entry.
func SumRefundSummarys(xs []RefundSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRefundSummarys keys the entries by ID, last write winning.
func IndexRefundSummarys(xs []RefundSummary) map[string]RefundSummary {
	out := make(map[string]RefundSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundSummary builds a valid RefundSummary, for tests and for the seed data the cron
// binary writes.
func SampleRefundSummary(id string, at time.Time) RefundSummary {
	return RefundSummary{
		ID:        id,
		Source:    "source-" + id,
		Label:     "label-" + id,
		Reference: "seeded by refund.SampleRefundSummary",
		Retained:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RefundPolicySchema is the stored shape of a RefundPolicy.
var RefundPolicySchema = schema.New("refund.refundpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundPolicy is one refund policy.
type RefundPolicy struct {
	ID       string
	Channel  string
	Region   string
	Category string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundPolicy) Key() string { return "refund/refundpolicy/" + x.ID }

// Fields encodes x for storage.
func (x RefundPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundPolicy decodes a stored record.
func ParseRefundPolicy(f codec.Fields) (RefundPolicy, error) {
	var x RefundPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundpolicy: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("refund: refundpolicy %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refundpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundPolicy) String() string {
	return fmt.Sprintf("RefundPolicy(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundPolicy) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundPolicy) Normalize() RefundPolicy {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRefundPolicys orders by count descending, then by ID.
func SortRefundPolicys(xs []RefundPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundPolicys drops the archived entries, preserving order.
func ActiveRefundPolicys(xs []RefundPolicy) []RefundPolicy {
	out := make([]RefundPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundPolicys adds up the count of every active entry.
func SumRefundPolicys(xs []RefundPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRefundPolicys keys the entries by ID, last write winning.
func IndexRefundPolicys(xs []RefundPolicy) map[string]RefundPolicy {
	out := make(map[string]RefundPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundPolicy builds a valid RefundPolicy, for tests and for the seed data the cron
// binary writes.
func SampleRefundPolicy(id string, at time.Time) RefundPolicy {
	return RefundPolicy{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Category: "seeded by refund.SampleRefundPolicy",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundLinkSchema is the stored shape of a RefundLink.
var RefundLinkSchema = schema.New("refund.refundlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundLink is one refund link.
type RefundLink struct {
	ID       string
	Name     string
	Category string
	Owner    string
	Weight   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundLink) Key() string { return "refund/refundlink/" + x.ID }

// Fields encodes x for storage.
func (x RefundLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"owner":    x.Owner,
		"weight":   strconv.Itoa(x.Weight),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundLink decodes a stored record.
func ParseRefundLink(f codec.Fields) (RefundLink, error) {
	var x RefundLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundlink: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refundlink %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("refund: refundlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundLink) String() string {
	return fmt.Sprintf("RefundLink(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundLink) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundLink) Normalize() RefundLink {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRefundLinks orders by weight descending, then by ID.
func SortRefundLinks(xs []RefundLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundLinks drops the archived entries, preserving order.
func ActiveRefundLinks(xs []RefundLink) []RefundLink {
	out := make([]RefundLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundLinks adds up the weight of every active entry.
func SumRefundLinks(xs []RefundLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRefundLinks keys the entries by ID, last write winning.
func IndexRefundLinks(xs []RefundLink) map[string]RefundLink {
	out := make(map[string]RefundLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundLink builds a valid RefundLink, for tests and for the seed data the cron
// binary writes.
func SampleRefundLink(id string, at time.Time) RefundLink {
	return RefundLink{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Owner:    "seeded by refund.SampleRefundLink",
		Weight:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundNoteSchema is the stored shape of a RefundNote.
var RefundNoteSchema = schema.New("refund.refundnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundNote is one refund note.
type RefundNote struct {
	ID       string
	Label    string
	Group    string
	Comment  string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundNote) Key() string { return "refund/refundnote/" + x.ID }

// Fields encodes x for storage.
func (x RefundNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"comment":  x.Comment,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundNote decodes a stored record.
func ParseRefundNote(f codec.Fields) (RefundNote, error) {
	var x RefundNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundnote: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("refund: refundnote %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("refund: refundnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundNote) String() string {
	return fmt.Sprintf("RefundNote(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundNote) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundNote) Normalize() RefundNote {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRefundNotes orders by attempts descending, then by ID.
func SortRefundNotes(xs []RefundNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundNotes drops the archived entries, preserving order.
func ActiveRefundNotes(xs []RefundNote) []RefundNote {
	out := make([]RefundNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundNotes adds up the attempts of every active entry.
func SumRefundNotes(xs []RefundNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexRefundNotes keys the entries by ID, last write winning.
func IndexRefundNotes(xs []RefundNote) map[string]RefundNote {
	out := make(map[string]RefundNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundNote builds a valid RefundNote, for tests and for the seed data the cron
// binary writes.
func SampleRefundNote(id string, at time.Time) RefundNote {
	return RefundNote{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Comment:  "seeded by refund.SampleRefundNote",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundWindowSchema is the stored shape of a RefundWindow.
var RefundWindowSchema = schema.New("refund.refundwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundWindow is one refund window.
type RefundWindow struct {
	ID        string
	Label     string
	Reference string
	Source    string
	Priority  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RefundWindow) Key() string { return "refund/refundwindow/" + x.ID }

// Fields encodes x for storage.
func (x RefundWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"reference": x.Reference,
		"source":    x.Source,
		"priority":  strconv.Itoa(x.Priority),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundWindow decodes a stored record.
func ParseRefundWindow(f codec.Fields) (RefundWindow, error) {
	var x RefundWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundwindow: missing id")
	}
	x.Label = f["label"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("refund: refundwindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("refund: refundwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundWindow) String() string {
	return fmt.Sprintf("RefundWindow(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundWindow) Normalize() RefundWindow {
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRefundWindows orders by priority descending, then by ID.
func SortRefundWindows(xs []RefundWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundWindows drops the archived entries, preserving order.
func ActiveRefundWindows(xs []RefundWindow) []RefundWindow {
	out := make([]RefundWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundWindows adds up the priority of every active entry.
func SumRefundWindows(xs []RefundWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRefundWindows keys the entries by ID, last write winning.
func IndexRefundWindows(xs []RefundWindow) map[string]RefundWindow {
	out := make(map[string]RefundWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundWindow builds a valid RefundWindow, for tests and for the seed data the cron
// binary writes.
func SampleRefundWindow(id string, at time.Time) RefundWindow {
	return RefundWindow{
		ID:        id,
		Label:     "label-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by refund.SampleRefundWindow",
		Priority:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RefundQuotaSchema is the stored shape of a RefundQuota.
var RefundQuotaSchema = schema.New("refund.refundquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundQuota is one refund quota.
type RefundQuota struct {
	ID        string
	Label     string
	Reference string
	Comment   string
	Capacity  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RefundQuota) Key() string { return "refund/refundquota/" + x.ID }

// Fields encodes x for storage.
func (x RefundQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"reference": x.Reference,
		"comment":   x.Comment,
		"capacity":  strconv.Itoa(x.Capacity),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundQuota decodes a stored record.
func ParseRefundQuota(f codec.Fields) (RefundQuota, error) {
	var x RefundQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundquota: missing id")
	}
	x.Label = f["label"]
	x.Reference = f["reference"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refundquota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refundquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundQuota) String() string {
	return fmt.Sprintf("RefundQuota(%s, label=%q, capacity=%d, archived=%t)", x.ID, x.Label, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundQuota) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundQuota) Normalize() RefundQuota {
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRefundQuotas orders by capacity descending, then by ID.
func SortRefundQuotas(xs []RefundQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundQuotas drops the archived entries, preserving order.
func ActiveRefundQuotas(xs []RefundQuota) []RefundQuota {
	out := make([]RefundQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundQuotas adds up the capacity of every active entry.
func SumRefundQuotas(xs []RefundQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRefundQuotas keys the entries by ID, last write winning.
func IndexRefundQuotas(xs []RefundQuota) map[string]RefundQuota {
	out := make(map[string]RefundQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundQuota builds a valid RefundQuota, for tests and for the seed data the cron
// binary writes.
func SampleRefundQuota(id string, at time.Time) RefundQuota {
	return RefundQuota{
		ID:        id,
		Label:     "label-" + id,
		Reference: "reference-" + id,
		Comment:   "seeded by refund.SampleRefundQuota",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RefundBucketSchema is the stored shape of a RefundBucket.
var RefundBucketSchema = schema.New("refund.refundbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundBucket is one refund bucket.
type RefundBucket struct {
	ID       string
	Label    string
	Channel  string
	Group    string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundBucket) Key() string { return "refund/refundbucket/" + x.ID }

// Fields encodes x for storage.
func (x RefundBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundBucket decodes a stored record.
func ParseRefundBucket(f codec.Fields) (RefundBucket, error) {
	var x RefundBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundbucket: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("refund: refundbucket %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("refund: refundbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundBucket) String() string {
	return fmt.Sprintf("RefundBucket(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundBucket) Normalize() RefundBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortRefundBuckets orders by retained descending, then by ID.
func SortRefundBuckets(xs []RefundBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundBuckets drops the archived entries, preserving order.
func ActiveRefundBuckets(xs []RefundBucket) []RefundBucket {
	out := make([]RefundBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundBuckets adds up the retained of every active entry.
func SumRefundBuckets(xs []RefundBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRefundBuckets keys the entries by ID, last write winning.
func IndexRefundBuckets(xs []RefundBucket) map[string]RefundBucket {
	out := make(map[string]RefundBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundBucket builds a valid RefundBucket, for tests and for the seed data the cron
// binary writes.
func SampleRefundBucket(id string, at time.Time) RefundBucket {
	return RefundBucket{
		ID:       id,
		Label:    "label-" + id,
		Channel:  "channel-" + id,
		Group:    "seeded by refund.SampleRefundBucket",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundLaneSchema is the stored shape of a RefundLane.
var RefundLaneSchema = schema.New("refund.refundlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundLane is one refund lane.
type RefundLane struct {
	ID        string
	Source    string
	Group     string
	Reference string
	Retained  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RefundLane) Key() string { return "refund/refundlane/" + x.ID }

// Fields encodes x for storage.
func (x RefundLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"group":     x.Group,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundLane decodes a stored record.
func ParseRefundLane(f codec.Fields) (RefundLane, error) {
	var x RefundLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundlane: missing id")
	}
	x.Source = f["source"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("refund: refundlane %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refundlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundLane) String() string {
	return fmt.Sprintf("RefundLane(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundLane) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundLane) Normalize() RefundLane {
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRefundLanes orders by retained descending, then by ID.
func SortRefundLanes(xs []RefundLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundLanes drops the archived entries, preserving order.
func ActiveRefundLanes(xs []RefundLane) []RefundLane {
	out := make([]RefundLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundLanes adds up the retained of every active entry.
func SumRefundLanes(xs []RefundLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRefundLanes keys the entries by ID, last write winning.
func IndexRefundLanes(xs []RefundLane) map[string]RefundLane {
	out := make(map[string]RefundLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundLane builds a valid RefundLane, for tests and for the seed data the cron
// binary writes.
func SampleRefundLane(id string, at time.Time) RefundLane {
	return RefundLane{
		ID:        id,
		Source:    "source-" + id,
		Group:     "group-" + id,
		Reference: "seeded by refund.SampleRefundLane",
		Retained:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RefundSlotSchema is the stored shape of a RefundSlot.
var RefundSlotSchema = schema.New("refund.refundslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundSlot is one refund slot.
type RefundSlot struct {
	ID       string
	Category string
	Channel  string
	Region   string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundSlot) Key() string { return "refund/refundslot/" + x.ID }

// Fields encodes x for storage.
func (x RefundSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundSlot decodes a stored record.
func ParseRefundSlot(f codec.Fields) (RefundSlot, error) {
	var x RefundSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundslot: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("refund: refundslot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refundslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundSlot) String() string {
	return fmt.Sprintf("RefundSlot(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundSlot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundSlot) Normalize() RefundSlot {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRefundSlots orders by offset descending, then by ID.
func SortRefundSlots(xs []RefundSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundSlots drops the archived entries, preserving order.
func ActiveRefundSlots(xs []RefundSlot) []RefundSlot {
	out := make([]RefundSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundSlots adds up the offset of every active entry.
func SumRefundSlots(xs []RefundSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRefundSlots keys the entries by ID, last write winning.
func IndexRefundSlots(xs []RefundSlot) map[string]RefundSlot {
	out := make(map[string]RefundSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundSlot builds a valid RefundSlot, for tests and for the seed data the cron
// binary writes.
func SampleRefundSlot(id string, at time.Time) RefundSlot {
	return RefundSlot{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by refund.SampleRefundSlot",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundTagSchema is the stored shape of a RefundTag.
var RefundTagSchema = schema.New("refund.refundtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundTag is one refund tag.
type RefundTag struct {
	ID        string
	Comment   string
	Reference string
	Category  string
	Count     int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RefundTag) Key() string { return "refund/refundtag/" + x.ID }

// Fields encodes x for storage.
func (x RefundTag) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"category":  x.Category,
		"count":     strconv.Itoa(x.Count),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundTag decodes a stored record.
func ParseRefundTag(f codec.Fields) (RefundTag, error) {
	var x RefundTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundtag: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("refund: refundtag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("refund: refundtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundTag) String() string {
	return fmt.Sprintf("RefundTag(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundTag) Normalize() RefundTag {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortRefundTags orders by count descending, then by ID.
func SortRefundTags(xs []RefundTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundTags drops the archived entries, preserving order.
func ActiveRefundTags(xs []RefundTag) []RefundTag {
	out := make([]RefundTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundTags adds up the count of every active entry.
func SumRefundTags(xs []RefundTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRefundTags keys the entries by ID, last write winning.
func IndexRefundTags(xs []RefundTag) map[string]RefundTag {
	out := make(map[string]RefundTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundTag builds a valid RefundTag, for tests and for the seed data the cron
// binary writes.
func SampleRefundTag(id string, at time.Time) RefundTag {
	return RefundTag{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by refund.SampleRefundTag",
		Count:     len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RefundRuleSchema is the stored shape of a RefundRule.
var RefundRuleSchema = schema.New("refund.refundrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundRule is one refund rule.
type RefundRule struct {
	ID       string
	Group    string
	Region   string
	Comment  string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundRule) Key() string { return "refund/refundrule/" + x.ID }

// Fields encodes x for storage.
func (x RefundRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"region":   x.Region,
		"comment":  x.Comment,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundRule decodes a stored record.
func ParseRefundRule(f codec.Fields) (RefundRule, error) {
	var x RefundRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundrule: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refundrule %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("refund: refundrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundRule) String() string {
	return fmt.Sprintf("RefundRule(%s, group=%q, weight=%d, archived=%t)", x.ID, x.Group, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundRule) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundRule) Normalize() RefundRule {
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRefundRules orders by weight descending, then by ID.
func SortRefundRules(xs []RefundRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundRules drops the archived entries, preserving order.
func ActiveRefundRules(xs []RefundRule) []RefundRule {
	out := make([]RefundRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundRules adds up the weight of every active entry.
func SumRefundRules(xs []RefundRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRefundRules keys the entries by ID, last write winning.
func IndexRefundRules(xs []RefundRule) map[string]RefundRule {
	out := make(map[string]RefundRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundRule builds a valid RefundRule, for tests and for the seed data the cron
// binary writes.
func SampleRefundRule(id string, at time.Time) RefundRule {
	return RefundRule{
		ID:       id,
		Group:    "group-" + id,
		Region:   "region-" + id,
		Comment:  "seeded by refund.SampleRefundRule",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundHoldSchema is the stored shape of a RefundHold.
var RefundHoldSchema = schema.New("refund.refundhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundHold is one refund hold.
type RefundHold struct {
	ID       string
	Region   string
	Comment  string
	Source   string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundHold) Key() string { return "refund/refundhold/" + x.ID }

// Fields encodes x for storage.
func (x RefundHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundHold decodes a stored record.
func ParseRefundHold(f codec.Fields) (RefundHold, error) {
	var x RefundHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundhold: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refundhold %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("refund: refundhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundHold) String() string {
	return fmt.Sprintf("RefundHold(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundHold) Normalize() RefundHold {
	x.Region = strings.TrimSpace(x.Region)
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

// SortRefundHolds orders by weight descending, then by ID.
func SortRefundHolds(xs []RefundHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundHolds drops the archived entries, preserving order.
func ActiveRefundHolds(xs []RefundHold) []RefundHold {
	out := make([]RefundHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundHolds adds up the weight of every active entry.
func SumRefundHolds(xs []RefundHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRefundHolds keys the entries by ID, last write winning.
func IndexRefundHolds(xs []RefundHold) map[string]RefundHold {
	out := make(map[string]RefundHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundHold builds a valid RefundHold, for tests and for the seed data the cron
// binary writes.
func SampleRefundHold(id string, at time.Time) RefundHold {
	return RefundHold{
		ID:       id,
		Region:   "region-" + id,
		Comment:  "comment-" + id,
		Source:   "seeded by refund.SampleRefundHold",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundClaimSchema is the stored shape of a RefundClaim.
var RefundClaimSchema = schema.New("refund.refundclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundClaim is one refund claim.
type RefundClaim struct {
	ID       string
	Label    string
	Region   string
	Owner    string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RefundClaim) Key() string { return "refund/refundclaim/" + x.ID }

// Fields encodes x for storage.
func (x RefundClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"owner":    x.Owner,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundClaim decodes a stored record.
func ParseRefundClaim(f codec.Fields) (RefundClaim, error) {
	var x RefundClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundclaim: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("refund: refundclaim %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("refund: refundclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundClaim) String() string {
	return fmt.Sprintf("RefundClaim(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundClaim) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundClaim) Normalize() RefundClaim {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortRefundClaims orders by count descending, then by ID.
func SortRefundClaims(xs []RefundClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundClaims drops the archived entries, preserving order.
func ActiveRefundClaims(xs []RefundClaim) []RefundClaim {
	out := make([]RefundClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundClaims adds up the count of every active entry.
func SumRefundClaims(xs []RefundClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRefundClaims keys the entries by ID, last write winning.
func IndexRefundClaims(xs []RefundClaim) map[string]RefundClaim {
	out := make(map[string]RefundClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundClaim builds a valid RefundClaim, for tests and for the seed data the cron
// binary writes.
func SampleRefundClaim(id string, at time.Time) RefundClaim {
	return RefundClaim{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Owner:    "seeded by refund.SampleRefundClaim",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RefundCursorSchema is the stored shape of a RefundCursor.
var RefundCursorSchema = schema.New("refund.refundcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RefundCursor is one refund cursor.
type RefundCursor struct {
	ID        string
	Channel   string
	Reference string
	Category  string
	Weight    int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RefundCursor) Key() string { return "refund/refundcursor/" + x.ID }

// Fields encodes x for storage.
func (x RefundCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"reference": x.Reference,
		"category":  x.Category,
		"weight":    strconv.Itoa(x.Weight),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRefundCursor decodes a stored record.
func ParseRefundCursor(f codec.Fields) (RefundCursor, error) {
	var x RefundCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("refund: refundcursor: missing id")
	}
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("refund: refundcursor %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("refund: refundcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("refund: refundcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("refund: refundcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RefundCursor) String() string {
	return fmt.Sprintf("RefundCursor(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RefundCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RefundCursor) Normalize() RefundCursor {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortRefundCursors orders by weight descending, then by ID.
func SortRefundCursors(xs []RefundCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRefundCursors drops the archived entries, preserving order.
func ActiveRefundCursors(xs []RefundCursor) []RefundCursor {
	out := make([]RefundCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRefundCursors adds up the weight of every active entry.
func SumRefundCursors(xs []RefundCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRefundCursors keys the entries by ID, last write winning.
func IndexRefundCursors(xs []RefundCursor) map[string]RefundCursor {
	out := make(map[string]RefundCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRefundCursor builds a valid RefundCursor, for tests and for the seed data the cron
// binary writes.
func SampleRefundCursor(id string, at time.Time) RefundCursor {
	return RefundCursor{
		ID:        id,
		Channel:   "channel-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by refund.SampleRefundCursor",
		Weight:    len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	RefundRecordSchema,
	RefundEntrySchema,
	RefundItemSchema,
	RefundDraftSchema,
	RefundRevisionSchema,
	RefundSnapshotSchema,
	RefundBatchSchema,
	RefundRequestSchema,
	RefundSummarySchema,
	RefundPolicySchema,
	RefundLinkSchema,
	RefundNoteSchema,
	RefundWindowSchema,
	RefundQuotaSchema,
	RefundBucketSchema,
	RefundLaneSchema,
	RefundSlotSchema,
	RefundTagSchema,
	RefundRuleSchema,
	RefundHoldSchema,
	RefundClaimSchema,
	RefundCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"refund/refundrecord/",
	"refund/refundentry/",
	"refund/refunditem/",
	"refund/refunddraft/",
	"refund/refundrevision/",
	"refund/refundsnapshot/",
	"refund/refundbatch/",
	"refund/refundrequest/",
	"refund/refundsummary/",
	"refund/refundpolicy/",
	"refund/refundlink/",
	"refund/refundnote/",
	"refund/refundwindow/",
	"refund/refundquota/",
	"refund/refundbucket/",
	"refund/refundlane/",
	"refund/refundslot/",
	"refund/refundtag/",
	"refund/refundrule/",
	"refund/refundhold/",
	"refund/refundclaim/",
	"refund/refundcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "refund: " + strings.Join(names, ", ")
}
