// Package delivery holds the delivery domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package delivery

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// DeliveryRecordSchema is the stored shape of a DeliveryRecord.
var DeliveryRecordSchema = schema.New("delivery.deliveryrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryRecord is one delivery record.
type DeliveryRecord struct {
	ID       string
	Source   string
	Comment  string
	Owner    string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryRecord) Key() string { return "delivery/deliveryrecord/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryRecord decodes a stored record.
func ParseDeliveryRecord(f codec.Fields) (DeliveryRecord, error) {
	var x DeliveryRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryrecord: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrecord %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryRecord) String() string {
	return fmt.Sprintf("DeliveryRecord(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryRecord) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryRecord) Normalize() DeliveryRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDeliveryRecords orders by priority descending, then by ID.
func SortDeliveryRecords(xs []DeliveryRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryRecords drops the archived entries, preserving order.
func ActiveDeliveryRecords(xs []DeliveryRecord) []DeliveryRecord {
	out := make([]DeliveryRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryRecords adds up the priority of every active entry.
func SumDeliveryRecords(xs []DeliveryRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeliveryRecords keys the entries by ID, last write winning.
func IndexDeliveryRecords(xs []DeliveryRecord) map[string]DeliveryRecord {
	out := make(map[string]DeliveryRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryRecord builds a valid DeliveryRecord, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryRecord(id string, at time.Time) DeliveryRecord {
	return DeliveryRecord{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by delivery.SampleDeliveryRecord",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryEntrySchema is the stored shape of a DeliveryEntry.
var DeliveryEntrySchema = schema.New("delivery.deliveryentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryEntry is one delivery entry.
type DeliveryEntry struct {
	ID       string
	Owner    string
	Source   string
	Comment  string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryEntry) Key() string { return "delivery/deliveryentry/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"comment":  x.Comment,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryEntry decodes a stored record.
func ParseDeliveryEntry(f codec.Fields) (DeliveryEntry, error) {
	var x DeliveryEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryentry: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryentry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryEntry) String() string {
	return fmt.Sprintf("DeliveryEntry(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryEntry) Normalize() DeliveryEntry {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDeliveryEntrys orders by offset descending, then by ID.
func SortDeliveryEntrys(xs []DeliveryEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryEntrys drops the archived entries, preserving order.
func ActiveDeliveryEntrys(xs []DeliveryEntry) []DeliveryEntry {
	out := make([]DeliveryEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryEntrys adds up the offset of every active entry.
func SumDeliveryEntrys(xs []DeliveryEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDeliveryEntrys keys the entries by ID, last write winning.
func IndexDeliveryEntrys(xs []DeliveryEntry) map[string]DeliveryEntry {
	out := make(map[string]DeliveryEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryEntry builds a valid DeliveryEntry, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryEntry(id string, at time.Time) DeliveryEntry {
	return DeliveryEntry{
		ID:      id,
		Owner:   "owner-" + id,
		Source:  "source-" + id,
		Comment: "seeded by delivery.SampleDeliveryEntry",
		Offset:  len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// DeliveryItemSchema is the stored shape of a DeliveryItem.
var DeliveryItemSchema = schema.New("delivery.deliveryitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryItem is one delivery item.
type DeliveryItem struct {
	ID       string
	Owner    string
	Region   string
	Group    string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryItem) Key() string { return "delivery/deliveryitem/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryItem decodes a stored record.
func ParseDeliveryItem(f codec.Fields) (DeliveryItem, error) {
	var x DeliveryItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryitem: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryitem %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryItem) String() string {
	return fmt.Sprintf("DeliveryItem(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryItem) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryItem) Normalize() DeliveryItem {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortDeliveryItems orders by attempts descending, then by ID.
func SortDeliveryItems(xs []DeliveryItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryItems drops the archived entries, preserving order.
func ActiveDeliveryItems(xs []DeliveryItem) []DeliveryItem {
	out := make([]DeliveryItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryItems adds up the attempts of every active entry.
func SumDeliveryItems(xs []DeliveryItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDeliveryItems keys the entries by ID, last write winning.
func IndexDeliveryItems(xs []DeliveryItem) map[string]DeliveryItem {
	out := make(map[string]DeliveryItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryItem builds a valid DeliveryItem, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryItem(id string, at time.Time) DeliveryItem {
	return DeliveryItem{
		ID:       id,
		Owner:    "owner-" + id,
		Region:   "region-" + id,
		Group:    "seeded by delivery.SampleDeliveryItem",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryDraftSchema is the stored shape of a DeliveryDraft.
var DeliveryDraftSchema = schema.New("delivery.deliverydraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryDraft is one delivery draft.
type DeliveryDraft struct {
	ID       string
	Source   string
	Owner    string
	Label    string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryDraft) Key() string { return "delivery/deliverydraft/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryDraft decodes a stored record.
func ParseDeliveryDraft(f codec.Fields) (DeliveryDraft, error) {
	var x DeliveryDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverydraft: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverydraft %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverydraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverydraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverydraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryDraft) String() string {
	return fmt.Sprintf("DeliveryDraft(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryDraft) Normalize() DeliveryDraft {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeliveryDrafts orders by attempts descending, then by ID.
func SortDeliveryDrafts(xs []DeliveryDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryDrafts drops the archived entries, preserving order.
func ActiveDeliveryDrafts(xs []DeliveryDraft) []DeliveryDraft {
	out := make([]DeliveryDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryDrafts adds up the attempts of every active entry.
func SumDeliveryDrafts(xs []DeliveryDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDeliveryDrafts keys the entries by ID, last write winning.
func IndexDeliveryDrafts(xs []DeliveryDraft) map[string]DeliveryDraft {
	out := make(map[string]DeliveryDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryDraft builds a valid DeliveryDraft, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryDraft(id string, at time.Time) DeliveryDraft {
	return DeliveryDraft{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by delivery.SampleDeliveryDraft",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryRevisionSchema is the stored shape of a DeliveryRevision.
var DeliveryRevisionSchema = schema.New("delivery.deliveryrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryRevision is one delivery revision.
type DeliveryRevision struct {
	ID       string
	Category string
	Label    string
	Channel  string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryRevision) Key() string { return "delivery/deliveryrevision/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"channel":  x.Channel,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryRevision decodes a stored record.
func ParseDeliveryRevision(f codec.Fields) (DeliveryRevision, error) {
	var x DeliveryRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryrevision: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryRevision) String() string {
	return fmt.Sprintf("DeliveryRevision(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryRevision) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryRevision) Normalize() DeliveryRevision {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDeliveryRevisions orders by offset descending, then by ID.
func SortDeliveryRevisions(xs []DeliveryRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryRevisions drops the archived entries, preserving order.
func ActiveDeliveryRevisions(xs []DeliveryRevision) []DeliveryRevision {
	out := make([]DeliveryRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryRevisions adds up the offset of every active entry.
func SumDeliveryRevisions(xs []DeliveryRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDeliveryRevisions keys the entries by ID, last write winning.
func IndexDeliveryRevisions(xs []DeliveryRevision) map[string]DeliveryRevision {
	out := make(map[string]DeliveryRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryRevision builds a valid DeliveryRevision, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryRevision(id string, at time.Time) DeliveryRevision {
	return DeliveryRevision{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by delivery.SampleDeliveryRevision",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliverySnapshotSchema is the stored shape of a DeliverySnapshot.
var DeliverySnapshotSchema = schema.New("delivery.deliverysnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliverySnapshot is one delivery snapshot.
type DeliverySnapshot struct {
	ID       string
	Name     string
	Owner    string
	Group    string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliverySnapshot) Key() string { return "delivery/deliverysnapshot/" + x.ID }

// Fields encodes x for storage.
func (x DeliverySnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliverySnapshot decodes a stored record.
func ParseDeliverySnapshot(f codec.Fields) (DeliverySnapshot, error) {
	var x DeliverySnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverysnapshot: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverysnapshot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverysnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverysnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverysnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliverySnapshot) String() string {
	return fmt.Sprintf("DeliverySnapshot(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliverySnapshot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliverySnapshot) Normalize() DeliverySnapshot {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDeliverySnapshots orders by offset descending, then by ID.
func SortDeliverySnapshots(xs []DeliverySnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliverySnapshots drops the archived entries, preserving order.
func ActiveDeliverySnapshots(xs []DeliverySnapshot) []DeliverySnapshot {
	out := make([]DeliverySnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliverySnapshots adds up the offset of every active entry.
func SumDeliverySnapshots(xs []DeliverySnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDeliverySnapshots keys the entries by ID, last write winning.
func IndexDeliverySnapshots(xs []DeliverySnapshot) map[string]DeliverySnapshot {
	out := make(map[string]DeliverySnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliverySnapshot builds a valid DeliverySnapshot, for tests and for the seed data the cron
// binary writes.
func SampleDeliverySnapshot(id string, at time.Time) DeliverySnapshot {
	return DeliverySnapshot{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by delivery.SampleDeliverySnapshot",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryBatchSchema is the stored shape of a DeliveryBatch.
var DeliveryBatchSchema = schema.New("delivery.deliverybatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryBatch is one delivery batch.
type DeliveryBatch struct {
	ID       string
	Owner    string
	Group    string
	Source   string
	Priority int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryBatch) Key() string { return "delivery/deliverybatch/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryBatch decodes a stored record.
func ParseDeliveryBatch(f codec.Fields) (DeliveryBatch, error) {
	var x DeliveryBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverybatch: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverybatch %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverybatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverybatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverybatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryBatch) String() string {
	return fmt.Sprintf("DeliveryBatch(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryBatch) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryBatch) Normalize() DeliveryBatch {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortDeliveryBatchs orders by priority descending, then by ID.
func SortDeliveryBatchs(xs []DeliveryBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryBatchs drops the archived entries, preserving order.
func ActiveDeliveryBatchs(xs []DeliveryBatch) []DeliveryBatch {
	out := make([]DeliveryBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryBatchs adds up the priority of every active entry.
func SumDeliveryBatchs(xs []DeliveryBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeliveryBatchs keys the entries by ID, last write winning.
func IndexDeliveryBatchs(xs []DeliveryBatch) map[string]DeliveryBatch {
	out := make(map[string]DeliveryBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryBatch builds a valid DeliveryBatch, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryBatch(id string, at time.Time) DeliveryBatch {
	return DeliveryBatch{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Source:   "seeded by delivery.SampleDeliveryBatch",
		Priority: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryRequestSchema is the stored shape of a DeliveryRequest.
var DeliveryRequestSchema = schema.New("delivery.deliveryrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryRequest is one delivery request.
type DeliveryRequest struct {
	ID       string
	Region   string
	Comment  string
	Group    string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryRequest) Key() string { return "delivery/deliveryrequest/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryRequest decodes a stored record.
func ParseDeliveryRequest(f codec.Fields) (DeliveryRequest, error) {
	var x DeliveryRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryrequest: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryRequest) String() string {
	return fmt.Sprintf("DeliveryRequest(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryRequest) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryRequest) Normalize() DeliveryRequest {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDeliveryRequests orders by priority descending, then by ID.
func SortDeliveryRequests(xs []DeliveryRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryRequests drops the archived entries, preserving order.
func ActiveDeliveryRequests(xs []DeliveryRequest) []DeliveryRequest {
	out := make([]DeliveryRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryRequests adds up the priority of every active entry.
func SumDeliveryRequests(xs []DeliveryRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeliveryRequests keys the entries by ID, last write winning.
func IndexDeliveryRequests(xs []DeliveryRequest) map[string]DeliveryRequest {
	out := make(map[string]DeliveryRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryRequest builds a valid DeliveryRequest, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryRequest(id string, at time.Time) DeliveryRequest {
	return DeliveryRequest{
		ID:       id,
		Region:   "region-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by delivery.SampleDeliveryRequest",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliverySummarySchema is the stored shape of a DeliverySummary.
var DeliverySummarySchema = schema.New("delivery.deliverysummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliverySummary is one delivery summary.
type DeliverySummary struct {
	ID       string
	Channel  string
	Category string
	Label    string
	Weight   int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliverySummary) Key() string { return "delivery/deliverysummary/" + x.ID }

// Fields encodes x for storage.
func (x DeliverySummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliverySummary decodes a stored record.
func ParseDeliverySummary(f codec.Fields) (DeliverySummary, error) {
	var x DeliverySummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverysummary: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverysummary %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverysummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverysummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverysummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliverySummary) String() string {
	return fmt.Sprintf("DeliverySummary(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliverySummary) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliverySummary) Normalize() DeliverySummary {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDeliverySummarys orders by weight descending, then by ID.
func SortDeliverySummarys(xs []DeliverySummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliverySummarys drops the archived entries, preserving order.
func ActiveDeliverySummarys(xs []DeliverySummary) []DeliverySummary {
	out := make([]DeliverySummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliverySummarys adds up the weight of every active entry.
func SumDeliverySummarys(xs []DeliverySummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDeliverySummarys keys the entries by ID, last write winning.
func IndexDeliverySummarys(xs []DeliverySummary) map[string]DeliverySummary {
	out := make(map[string]DeliverySummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliverySummary builds a valid DeliverySummary, for tests and for the seed data the cron
// binary writes.
func SampleDeliverySummary(id string, at time.Time) DeliverySummary {
	return DeliverySummary{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Label:    "seeded by delivery.SampleDeliverySummary",
		Weight:   len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryPolicySchema is the stored shape of a DeliveryPolicy.
var DeliveryPolicySchema = schema.New("delivery.deliverypolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryPolicy is one delivery policy.
type DeliveryPolicy struct {
	ID        string
	Name      string
	Group     string
	Reference string
	Retained  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeliveryPolicy) Key() string { return "delivery/deliverypolicy/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"group":     x.Group,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryPolicy decodes a stored record.
func ParseDeliveryPolicy(f codec.Fields) (DeliveryPolicy, error) {
	var x DeliveryPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverypolicy: missing id")
	}
	x.Name = f["name"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverypolicy %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverypolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverypolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverypolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryPolicy) String() string {
	return fmt.Sprintf("DeliveryPolicy(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryPolicy) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryPolicy) Normalize() DeliveryPolicy {
	x.Name = strings.TrimSpace(x.Name)
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

// SortDeliveryPolicys orders by retained descending, then by ID.
func SortDeliveryPolicys(xs []DeliveryPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryPolicys drops the archived entries, preserving order.
func ActiveDeliveryPolicys(xs []DeliveryPolicy) []DeliveryPolicy {
	out := make([]DeliveryPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryPolicys adds up the retained of every active entry.
func SumDeliveryPolicys(xs []DeliveryPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDeliveryPolicys keys the entries by ID, last write winning.
func IndexDeliveryPolicys(xs []DeliveryPolicy) map[string]DeliveryPolicy {
	out := make(map[string]DeliveryPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryPolicy builds a valid DeliveryPolicy, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryPolicy(id string, at time.Time) DeliveryPolicy {
	return DeliveryPolicy{
		ID:        id,
		Name:      "name-" + id,
		Group:     "group-" + id,
		Reference: "seeded by delivery.SampleDeliveryPolicy",
		Retained:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeliveryLinkSchema is the stored shape of a DeliveryLink.
var DeliveryLinkSchema = schema.New("delivery.deliverylink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryLink is one delivery link.
type DeliveryLink struct {
	ID       string
	Owner    string
	Group    string
	Region   string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryLink) Key() string { return "delivery/deliverylink/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryLink decodes a stored record.
func ParseDeliveryLink(f codec.Fields) (DeliveryLink, error) {
	var x DeliveryLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverylink: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverylink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverylink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverylink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverylink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryLink) String() string {
	return fmt.Sprintf("DeliveryLink(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryLink) Normalize() DeliveryLink {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDeliveryLinks orders by retained descending, then by ID.
func SortDeliveryLinks(xs []DeliveryLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryLinks drops the archived entries, preserving order.
func ActiveDeliveryLinks(xs []DeliveryLink) []DeliveryLink {
	out := make([]DeliveryLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryLinks adds up the retained of every active entry.
func SumDeliveryLinks(xs []DeliveryLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDeliveryLinks keys the entries by ID, last write winning.
func IndexDeliveryLinks(xs []DeliveryLink) map[string]DeliveryLink {
	out := make(map[string]DeliveryLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryLink builds a valid DeliveryLink, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryLink(id string, at time.Time) DeliveryLink {
	return DeliveryLink{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Region:   "seeded by delivery.SampleDeliveryLink",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryNoteSchema is the stored shape of a DeliveryNote.
var DeliveryNoteSchema = schema.New("delivery.deliverynote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryNote is one delivery note.
type DeliveryNote struct {
	ID       string
	Group    string
	Owner    string
	Comment  string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryNote) Key() string { return "delivery/deliverynote/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryNote decodes a stored record.
func ParseDeliveryNote(f codec.Fields) (DeliveryNote, error) {
	var x DeliveryNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverynote: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverynote %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverynote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverynote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverynote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryNote) String() string {
	return fmt.Sprintf("DeliveryNote(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryNote) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryNote) Normalize() DeliveryNote {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeliveryNotes orders by version descending, then by ID.
func SortDeliveryNotes(xs []DeliveryNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryNotes drops the archived entries, preserving order.
func ActiveDeliveryNotes(xs []DeliveryNote) []DeliveryNote {
	out := make([]DeliveryNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryNotes adds up the version of every active entry.
func SumDeliveryNotes(xs []DeliveryNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeliveryNotes keys the entries by ID, last write winning.
func IndexDeliveryNotes(xs []DeliveryNote) map[string]DeliveryNote {
	out := make(map[string]DeliveryNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryNote builds a valid DeliveryNote, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryNote(id string, at time.Time) DeliveryNote {
	return DeliveryNote{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Comment:  "seeded by delivery.SampleDeliveryNote",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryWindowSchema is the stored shape of a DeliveryWindow.
var DeliveryWindowSchema = schema.New("delivery.deliverywindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryWindow is one delivery window.
type DeliveryWindow struct {
	ID        string
	Source    string
	Owner     string
	Reference string
	Priority  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeliveryWindow) Key() string { return "delivery/deliverywindow/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"owner":     x.Owner,
		"reference": x.Reference,
		"priority":  strconv.Itoa(x.Priority),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryWindow decodes a stored record.
func ParseDeliveryWindow(f codec.Fields) (DeliveryWindow, error) {
	var x DeliveryWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverywindow: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverywindow %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverywindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverywindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverywindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryWindow) String() string {
	return fmt.Sprintf("DeliveryWindow(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryWindow) Normalize() DeliveryWindow {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDeliveryWindows orders by priority descending, then by ID.
func SortDeliveryWindows(xs []DeliveryWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryWindows drops the archived entries, preserving order.
func ActiveDeliveryWindows(xs []DeliveryWindow) []DeliveryWindow {
	out := make([]DeliveryWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryWindows adds up the priority of every active entry.
func SumDeliveryWindows(xs []DeliveryWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeliveryWindows keys the entries by ID, last write winning.
func IndexDeliveryWindows(xs []DeliveryWindow) map[string]DeliveryWindow {
	out := make(map[string]DeliveryWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryWindow builds a valid DeliveryWindow, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryWindow(id string, at time.Time) DeliveryWindow {
	return DeliveryWindow{
		ID:        id,
		Source:    "source-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by delivery.SampleDeliveryWindow",
		Priority:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeliveryQuotaSchema is the stored shape of a DeliveryQuota.
var DeliveryQuotaSchema = schema.New("delivery.deliveryquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryQuota is one delivery quota.
type DeliveryQuota struct {
	ID       string
	Label    string
	Name     string
	Source   string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryQuota) Key() string { return "delivery/deliveryquota/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"name":     x.Name,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryQuota decodes a stored record.
func ParseDeliveryQuota(f codec.Fields) (DeliveryQuota, error) {
	var x DeliveryQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryquota: missing id")
	}
	x.Label = f["label"]
	x.Name = f["name"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryquota %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryQuota) String() string {
	return fmt.Sprintf("DeliveryQuota(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryQuota) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryQuota) Normalize() DeliveryQuota {
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeliveryQuotas orders by priority descending, then by ID.
func SortDeliveryQuotas(xs []DeliveryQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryQuotas drops the archived entries, preserving order.
func ActiveDeliveryQuotas(xs []DeliveryQuota) []DeliveryQuota {
	out := make([]DeliveryQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryQuotas adds up the priority of every active entry.
func SumDeliveryQuotas(xs []DeliveryQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeliveryQuotas keys the entries by ID, last write winning.
func IndexDeliveryQuotas(xs []DeliveryQuota) map[string]DeliveryQuota {
	out := make(map[string]DeliveryQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryQuota builds a valid DeliveryQuota, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryQuota(id string, at time.Time) DeliveryQuota {
	return DeliveryQuota{
		ID:       id,
		Label:    "label-" + id,
		Name:     "name-" + id,
		Source:   "seeded by delivery.SampleDeliveryQuota",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryBucketSchema is the stored shape of a DeliveryBucket.
var DeliveryBucketSchema = schema.New("delivery.deliverybucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryBucket is one delivery bucket.
type DeliveryBucket struct {
	ID        string
	Reference string
	Category  string
	Region    string
	Count     int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeliveryBucket) Key() string { return "delivery/deliverybucket/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"category":  x.Category,
		"region":    x.Region,
		"count":     strconv.Itoa(x.Count),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryBucket decodes a stored record.
func ParseDeliveryBucket(f codec.Fields) (DeliveryBucket, error) {
	var x DeliveryBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverybucket: missing id")
	}
	x.Reference = f["reference"]
	x.Category = f["category"]
	x.Region = f["region"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverybucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverybucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverybucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverybucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryBucket) String() string {
	return fmt.Sprintf("DeliveryBucket(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryBucket) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryBucket) Normalize() DeliveryBucket {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDeliveryBuckets orders by count descending, then by ID.
func SortDeliveryBuckets(xs []DeliveryBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryBuckets drops the archived entries, preserving order.
func ActiveDeliveryBuckets(xs []DeliveryBucket) []DeliveryBucket {
	out := make([]DeliveryBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryBuckets adds up the count of every active entry.
func SumDeliveryBuckets(xs []DeliveryBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDeliveryBuckets keys the entries by ID, last write winning.
func IndexDeliveryBuckets(xs []DeliveryBucket) map[string]DeliveryBucket {
	out := make(map[string]DeliveryBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryBucket builds a valid DeliveryBucket, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryBucket(id string, at time.Time) DeliveryBucket {
	return DeliveryBucket{
		ID:        id,
		Reference: "reference-" + id,
		Category:  "category-" + id,
		Region:    "seeded by delivery.SampleDeliveryBucket",
		Count:     len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeliveryLaneSchema is the stored shape of a DeliveryLane.
var DeliveryLaneSchema = schema.New("delivery.deliverylane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryLane is one delivery lane.
type DeliveryLane struct {
	ID       string
	Group    string
	Name     string
	Category string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryLane) Key() string { return "delivery/deliverylane/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryLane decodes a stored record.
func ParseDeliveryLane(f codec.Fields) (DeliveryLane, error) {
	var x DeliveryLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverylane: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverylane %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverylane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverylane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverylane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryLane) String() string {
	return fmt.Sprintf("DeliveryLane(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryLane) Normalize() DeliveryLane {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDeliveryLanes orders by capacity descending, then by ID.
func SortDeliveryLanes(xs []DeliveryLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryLanes drops the archived entries, preserving order.
func ActiveDeliveryLanes(xs []DeliveryLane) []DeliveryLane {
	out := make([]DeliveryLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryLanes adds up the capacity of every active entry.
func SumDeliveryLanes(xs []DeliveryLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDeliveryLanes keys the entries by ID, last write winning.
func IndexDeliveryLanes(xs []DeliveryLane) map[string]DeliveryLane {
	out := make(map[string]DeliveryLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryLane builds a valid DeliveryLane, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryLane(id string, at time.Time) DeliveryLane {
	return DeliveryLane{
		ID:       id,
		Group:    "group-" + id,
		Name:     "name-" + id,
		Category: "seeded by delivery.SampleDeliveryLane",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliverySlotSchema is the stored shape of a DeliverySlot.
var DeliverySlotSchema = schema.New("delivery.deliveryslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliverySlot is one delivery slot.
type DeliverySlot struct {
	ID       string
	Category string
	Region   string
	Channel  string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliverySlot) Key() string { return "delivery/deliveryslot/" + x.ID }

// Fields encodes x for storage.
func (x DeliverySlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"channel":  x.Channel,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliverySlot decodes a stored record.
func ParseDeliverySlot(f codec.Fields) (DeliverySlot, error) {
	var x DeliverySlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryslot: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliverySlot) String() string {
	return fmt.Sprintf("DeliverySlot(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliverySlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliverySlot) Normalize() DeliverySlot {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortDeliverySlots orders by count descending, then by ID.
func SortDeliverySlots(xs []DeliverySlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliverySlots drops the archived entries, preserving order.
func ActiveDeliverySlots(xs []DeliverySlot) []DeliverySlot {
	out := make([]DeliverySlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliverySlots adds up the count of every active entry.
func SumDeliverySlots(xs []DeliverySlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDeliverySlots keys the entries by ID, last write winning.
func IndexDeliverySlots(xs []DeliverySlot) map[string]DeliverySlot {
	out := make(map[string]DeliverySlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliverySlot builds a valid DeliverySlot, for tests and for the seed data the cron
// binary writes.
func SampleDeliverySlot(id string, at time.Time) DeliverySlot {
	return DeliverySlot{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Channel:  "seeded by delivery.SampleDeliverySlot",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryTagSchema is the stored shape of a DeliveryTag.
var DeliveryTagSchema = schema.New("delivery.deliverytag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryTag is one delivery tag.
type DeliveryTag struct {
	ID       string
	Comment  string
	Source   string
	Category string
	Version  int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryTag) Key() string { return "delivery/deliverytag/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryTag decodes a stored record.
func ParseDeliveryTag(f codec.Fields) (DeliveryTag, error) {
	var x DeliveryTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverytag: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverytag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverytag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverytag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverytag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryTag) String() string {
	return fmt.Sprintf("DeliveryTag(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryTag) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryTag) Normalize() DeliveryTag {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeliveryTags orders by version descending, then by ID.
func SortDeliveryTags(xs []DeliveryTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryTags drops the archived entries, preserving order.
func ActiveDeliveryTags(xs []DeliveryTag) []DeliveryTag {
	out := make([]DeliveryTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryTags adds up the version of every active entry.
func SumDeliveryTags(xs []DeliveryTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeliveryTags keys the entries by ID, last write winning.
func IndexDeliveryTags(xs []DeliveryTag) map[string]DeliveryTag {
	out := make(map[string]DeliveryTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryTag builds a valid DeliveryTag, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryTag(id string, at time.Time) DeliveryTag {
	return DeliveryTag{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Category: "seeded by delivery.SampleDeliveryTag",
		Version:  len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DeliveryRuleSchema is the stored shape of a DeliveryRule.
var DeliveryRuleSchema = schema.New("delivery.deliveryrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryRule is one delivery rule.
type DeliveryRule struct {
	ID        string
	Owner     string
	Category  string
	Reference string
	Version   int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeliveryRule) Key() string { return "delivery/deliveryrule/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"category":  x.Category,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryRule decodes a stored record.
func ParseDeliveryRule(f codec.Fields) (DeliveryRule, error) {
	var x DeliveryRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryrule: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrule %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryRule) String() string {
	return fmt.Sprintf("DeliveryRule(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryRule) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryRule) Normalize() DeliveryRule {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDeliveryRules orders by version descending, then by ID.
func SortDeliveryRules(xs []DeliveryRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryRules drops the archived entries, preserving order.
func ActiveDeliveryRules(xs []DeliveryRule) []DeliveryRule {
	out := make([]DeliveryRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryRules adds up the version of every active entry.
func SumDeliveryRules(xs []DeliveryRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeliveryRules keys the entries by ID, last write winning.
func IndexDeliveryRules(xs []DeliveryRule) map[string]DeliveryRule {
	out := make(map[string]DeliveryRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryRule builds a valid DeliveryRule, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryRule(id string, at time.Time) DeliveryRule {
	return DeliveryRule{
		ID:        id,
		Owner:     "owner-" + id,
		Category:  "category-" + id,
		Reference: "seeded by delivery.SampleDeliveryRule",
		Version:   len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeliveryHoldSchema is the stored shape of a DeliveryHold.
var DeliveryHoldSchema = schema.New("delivery.deliveryhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryHold is one delivery hold.
type DeliveryHold struct {
	ID        string
	Source    string
	Reference string
	Category  string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeliveryHold) Key() string { return "delivery/deliveryhold/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"reference": x.Reference,
		"category":  x.Category,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryHold decodes a stored record.
func ParseDeliveryHold(f codec.Fields) (DeliveryHold, error) {
	var x DeliveryHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryhold: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryhold %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryHold) String() string {
	return fmt.Sprintf("DeliveryHold(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryHold) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryHold) Normalize() DeliveryHold {
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeliveryHolds orders by version descending, then by ID.
func SortDeliveryHolds(xs []DeliveryHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryHolds drops the archived entries, preserving order.
func ActiveDeliveryHolds(xs []DeliveryHold) []DeliveryHold {
	out := make([]DeliveryHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryHolds adds up the version of every active entry.
func SumDeliveryHolds(xs []DeliveryHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDeliveryHolds keys the entries by ID, last write winning.
func IndexDeliveryHolds(xs []DeliveryHold) map[string]DeliveryHold {
	out := make(map[string]DeliveryHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryHold builds a valid DeliveryHold, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryHold(id string, at time.Time) DeliveryHold {
	return DeliveryHold{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by delivery.SampleDeliveryHold",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeliveryClaimSchema is the stored shape of a DeliveryClaim.
var DeliveryClaimSchema = schema.New("delivery.deliveryclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryClaim is one delivery claim.
type DeliveryClaim struct {
	ID        string
	Reference string
	Channel   string
	Comment   string
	Priority  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DeliveryClaim) Key() string { return "delivery/deliveryclaim/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"comment":   x.Comment,
		"priority":  strconv.Itoa(x.Priority),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryClaim decodes a stored record.
func ParseDeliveryClaim(f codec.Fields) (DeliveryClaim, error) {
	var x DeliveryClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliveryclaim: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryclaim %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("delivery: deliveryclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliveryclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliveryclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryClaim) String() string {
	return fmt.Sprintf("DeliveryClaim(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryClaim) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryClaim) Normalize() DeliveryClaim {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDeliveryClaims orders by priority descending, then by ID.
func SortDeliveryClaims(xs []DeliveryClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryClaims drops the archived entries, preserving order.
func ActiveDeliveryClaims(xs []DeliveryClaim) []DeliveryClaim {
	out := make([]DeliveryClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryClaims adds up the priority of every active entry.
func SumDeliveryClaims(xs []DeliveryClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDeliveryClaims keys the entries by ID, last write winning.
func IndexDeliveryClaims(xs []DeliveryClaim) map[string]DeliveryClaim {
	out := make(map[string]DeliveryClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryClaim builds a valid DeliveryClaim, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryClaim(id string, at time.Time) DeliveryClaim {
	return DeliveryClaim{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Comment:   "seeded by delivery.SampleDeliveryClaim",
		Priority:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DeliveryCursorSchema is the stored shape of a DeliveryCursor.
var DeliveryCursorSchema = schema.New("delivery.deliverycursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DeliveryCursor is one delivery cursor.
type DeliveryCursor struct {
	ID       string
	Owner    string
	Name     string
	Comment  string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DeliveryCursor) Key() string { return "delivery/deliverycursor/" + x.ID }

// Fields encodes x for storage.
func (x DeliveryCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"name":     x.Name,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDeliveryCursor decodes a stored record.
func ParseDeliveryCursor(f codec.Fields) (DeliveryCursor, error) {
	var x DeliveryCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("delivery: deliverycursor: missing id")
	}
	x.Owner = f["owner"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverycursor %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("delivery: deliverycursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("delivery: deliverycursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("delivery: deliverycursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DeliveryCursor) String() string {
	return fmt.Sprintf("DeliveryCursor(%s, owner=%q, count=%d, archived=%t)", x.ID, x.Owner, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DeliveryCursor) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DeliveryCursor) Normalize() DeliveryCursor {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDeliveryCursors orders by count descending, then by ID.
func SortDeliveryCursors(xs []DeliveryCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDeliveryCursors drops the archived entries, preserving order.
func ActiveDeliveryCursors(xs []DeliveryCursor) []DeliveryCursor {
	out := make([]DeliveryCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDeliveryCursors adds up the count of every active entry.
func SumDeliveryCursors(xs []DeliveryCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDeliveryCursors keys the entries by ID, last write winning.
func IndexDeliveryCursors(xs []DeliveryCursor) map[string]DeliveryCursor {
	out := make(map[string]DeliveryCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDeliveryCursor builds a valid DeliveryCursor, for tests and for the seed data the cron
// binary writes.
func SampleDeliveryCursor(id string, at time.Time) DeliveryCursor {
	return DeliveryCursor{
		ID:      id,
		Owner:   "owner-" + id,
		Name:    "name-" + id,
		Comment: "seeded by delivery.SampleDeliveryCursor",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	DeliveryRecordSchema,
	DeliveryEntrySchema,
	DeliveryItemSchema,
	DeliveryDraftSchema,
	DeliveryRevisionSchema,
	DeliverySnapshotSchema,
	DeliveryBatchSchema,
	DeliveryRequestSchema,
	DeliverySummarySchema,
	DeliveryPolicySchema,
	DeliveryLinkSchema,
	DeliveryNoteSchema,
	DeliveryWindowSchema,
	DeliveryQuotaSchema,
	DeliveryBucketSchema,
	DeliveryLaneSchema,
	DeliverySlotSchema,
	DeliveryTagSchema,
	DeliveryRuleSchema,
	DeliveryHoldSchema,
	DeliveryClaimSchema,
	DeliveryCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"delivery/deliveryrecord/",
	"delivery/deliveryentry/",
	"delivery/deliveryitem/",
	"delivery/deliverydraft/",
	"delivery/deliveryrevision/",
	"delivery/deliverysnapshot/",
	"delivery/deliverybatch/",
	"delivery/deliveryrequest/",
	"delivery/deliverysummary/",
	"delivery/deliverypolicy/",
	"delivery/deliverylink/",
	"delivery/deliverynote/",
	"delivery/deliverywindow/",
	"delivery/deliveryquota/",
	"delivery/deliverybucket/",
	"delivery/deliverylane/",
	"delivery/deliveryslot/",
	"delivery/deliverytag/",
	"delivery/deliveryrule/",
	"delivery/deliveryhold/",
	"delivery/deliveryclaim/",
	"delivery/deliverycursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "delivery: " + strings.Join(names, ", ")
}
