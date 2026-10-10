// Package payment holds the payment domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package payment

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// PaymentRecordSchema is the stored shape of a PaymentRecord.
var PaymentRecordSchema = schema.New("payment.paymentrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentRecord is one payment record.
type PaymentRecord struct {
	ID       string
	Channel  string
	Owner    string
	Name     string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentRecord) Key() string { return "payment/paymentrecord/" + x.ID }

// Fields encodes x for storage.
func (x PaymentRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"name":     x.Name,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentRecord decodes a stored record.
func ParsePaymentRecord(f codec.Fields) (PaymentRecord, error) {
	var x PaymentRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentrecord: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentRecord) String() string {
	return fmt.Sprintf("PaymentRecord(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentRecord) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentRecord) Normalize() PaymentRecord {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPaymentRecords orders by version descending, then by ID.
func SortPaymentRecords(xs []PaymentRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentRecords drops the archived entries, preserving order.
func ActivePaymentRecords(xs []PaymentRecord) []PaymentRecord {
	out := make([]PaymentRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentRecords adds up the version of every active entry.
func SumPaymentRecords(xs []PaymentRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPaymentRecords keys the entries by ID, last write winning.
func IndexPaymentRecords(xs []PaymentRecord) map[string]PaymentRecord {
	out := make(map[string]PaymentRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentRecord builds a valid PaymentRecord, for tests and for the seed data the cron
// binary writes.
func SamplePaymentRecord(id string, at time.Time) PaymentRecord {
	return PaymentRecord{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by payment.SamplePaymentRecord",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentEntrySchema is the stored shape of a PaymentEntry.
var PaymentEntrySchema = schema.New("payment.paymententry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentEntry is one payment entry.
type PaymentEntry struct {
	ID       string
	Channel  string
	Label    string
	Source   string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentEntry) Key() string { return "payment/paymententry/" + x.ID }

// Fields encodes x for storage.
func (x PaymentEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentEntry decodes a stored record.
func ParsePaymentEntry(f codec.Fields) (PaymentEntry, error) {
	var x PaymentEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymententry: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("payment: paymententry %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("payment: paymententry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymententry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymententry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentEntry) String() string {
	return fmt.Sprintf("PaymentEntry(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentEntry) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentEntry) Normalize() PaymentEntry {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortPaymentEntrys orders by capacity descending, then by ID.
func SortPaymentEntrys(xs []PaymentEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentEntrys drops the archived entries, preserving order.
func ActivePaymentEntrys(xs []PaymentEntry) []PaymentEntry {
	out := make([]PaymentEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentEntrys adds up the capacity of every active entry.
func SumPaymentEntrys(xs []PaymentEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPaymentEntrys keys the entries by ID, last write winning.
func IndexPaymentEntrys(xs []PaymentEntry) map[string]PaymentEntry {
	out := make(map[string]PaymentEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentEntry builds a valid PaymentEntry, for tests and for the seed data the cron
// binary writes.
func SamplePaymentEntry(id string, at time.Time) PaymentEntry {
	return PaymentEntry{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Source:   "seeded by payment.SamplePaymentEntry",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentItemSchema is the stored shape of a PaymentItem.
var PaymentItemSchema = schema.New("payment.paymentitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentItem is one payment item.
type PaymentItem struct {
	ID       string
	Category string
	Name     string
	Group    string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentItem) Key() string { return "payment/paymentitem/" + x.ID }

// Fields encodes x for storage.
func (x PaymentItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentItem decodes a stored record.
func ParsePaymentItem(f codec.Fields) (PaymentItem, error) {
	var x PaymentItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentitem: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("payment: paymentitem %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentItem) String() string {
	return fmt.Sprintf("PaymentItem(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentItem) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentItem) Normalize() PaymentItem {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPaymentItems orders by version descending, then by ID.
func SortPaymentItems(xs []PaymentItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentItems drops the archived entries, preserving order.
func ActivePaymentItems(xs []PaymentItem) []PaymentItem {
	out := make([]PaymentItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentItems adds up the version of every active entry.
func SumPaymentItems(xs []PaymentItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPaymentItems keys the entries by ID, last write winning.
func IndexPaymentItems(xs []PaymentItem) map[string]PaymentItem {
	out := make(map[string]PaymentItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentItem builds a valid PaymentItem, for tests and for the seed data the cron
// binary writes.
func SamplePaymentItem(id string, at time.Time) PaymentItem {
	return PaymentItem{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Group:    "seeded by payment.SamplePaymentItem",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentDraftSchema is the stored shape of a PaymentDraft.
var PaymentDraftSchema = schema.New("payment.paymentdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentDraft is one payment draft.
type PaymentDraft struct {
	ID        string
	Region    string
	Source    string
	Reference string
	Offset    int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PaymentDraft) Key() string { return "payment/paymentdraft/" + x.ID }

// Fields encodes x for storage.
func (x PaymentDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"source":    x.Source,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentDraft decodes a stored record.
func ParsePaymentDraft(f codec.Fields) (PaymentDraft, error) {
	var x PaymentDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentdraft: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("payment: paymentdraft %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("payment: paymentdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentDraft) String() string {
	return fmt.Sprintf("PaymentDraft(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentDraft) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentDraft) Normalize() PaymentDraft {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortPaymentDrafts orders by offset descending, then by ID.
func SortPaymentDrafts(xs []PaymentDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentDrafts drops the archived entries, preserving order.
func ActivePaymentDrafts(xs []PaymentDraft) []PaymentDraft {
	out := make([]PaymentDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentDrafts adds up the offset of every active entry.
func SumPaymentDrafts(xs []PaymentDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPaymentDrafts keys the entries by ID, last write winning.
func IndexPaymentDrafts(xs []PaymentDraft) map[string]PaymentDraft {
	out := make(map[string]PaymentDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentDraft builds a valid PaymentDraft, for tests and for the seed data the cron
// binary writes.
func SamplePaymentDraft(id string, at time.Time) PaymentDraft {
	return PaymentDraft{
		ID:        id,
		Region:    "region-" + id,
		Source:    "source-" + id,
		Reference: "seeded by payment.SamplePaymentDraft",
		Offset:    len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PaymentRevisionSchema is the stored shape of a PaymentRevision.
var PaymentRevisionSchema = schema.New("payment.paymentrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentRevision is one payment revision.
type PaymentRevision struct {
	ID       string
	Label    string
	Region   string
	Category string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentRevision) Key() string { return "payment/paymentrevision/" + x.ID }

// Fields encodes x for storage.
func (x PaymentRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentRevision decodes a stored record.
func ParsePaymentRevision(f codec.Fields) (PaymentRevision, error) {
	var x PaymentRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentrevision: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentRevision) String() string {
	return fmt.Sprintf("PaymentRevision(%s, label=%q, attempts=%d, archived=%t)", x.ID, x.Label, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentRevision) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentRevision) Normalize() PaymentRevision {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortPaymentRevisions orders by attempts descending, then by ID.
func SortPaymentRevisions(xs []PaymentRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentRevisions drops the archived entries, preserving order.
func ActivePaymentRevisions(xs []PaymentRevision) []PaymentRevision {
	out := make([]PaymentRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentRevisions adds up the attempts of every active entry.
func SumPaymentRevisions(xs []PaymentRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexPaymentRevisions keys the entries by ID, last write winning.
func IndexPaymentRevisions(xs []PaymentRevision) map[string]PaymentRevision {
	out := make(map[string]PaymentRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentRevision builds a valid PaymentRevision, for tests and for the seed data the cron
// binary writes.
func SamplePaymentRevision(id string, at time.Time) PaymentRevision {
	return PaymentRevision{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Category: "seeded by payment.SamplePaymentRevision",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentSnapshotSchema is the stored shape of a PaymentSnapshot.
var PaymentSnapshotSchema = schema.New("payment.paymentsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentSnapshot is one payment snapshot.
type PaymentSnapshot struct {
	ID       string
	Channel  string
	Owner    string
	Source   string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentSnapshot) Key() string { return "payment/paymentsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x PaymentSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentSnapshot decodes a stored record.
func ParsePaymentSnapshot(f codec.Fields) (PaymentSnapshot, error) {
	var x PaymentSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentsnapshot: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentsnapshot %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("payment: paymentsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentSnapshot) String() string {
	return fmt.Sprintf("PaymentSnapshot(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentSnapshot) Normalize() PaymentSnapshot {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortPaymentSnapshots orders by priority descending, then by ID.
func SortPaymentSnapshots(xs []PaymentSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentSnapshots drops the archived entries, preserving order.
func ActivePaymentSnapshots(xs []PaymentSnapshot) []PaymentSnapshot {
	out := make([]PaymentSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentSnapshots adds up the priority of every active entry.
func SumPaymentSnapshots(xs []PaymentSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexPaymentSnapshots keys the entries by ID, last write winning.
func IndexPaymentSnapshots(xs []PaymentSnapshot) map[string]PaymentSnapshot {
	out := make(map[string]PaymentSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentSnapshot builds a valid PaymentSnapshot, for tests and for the seed data the cron
// binary writes.
func SamplePaymentSnapshot(id string, at time.Time) PaymentSnapshot {
	return PaymentSnapshot{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by payment.SamplePaymentSnapshot",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentBatchSchema is the stored shape of a PaymentBatch.
var PaymentBatchSchema = schema.New("payment.paymentbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentBatch is one payment batch.
type PaymentBatch struct {
	ID        string
	Reference string
	Channel   string
	Name      string
	Capacity  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PaymentBatch) Key() string { return "payment/paymentbatch/" + x.ID }

// Fields encodes x for storage.
func (x PaymentBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"name":      x.Name,
		"capacity":  strconv.Itoa(x.Capacity),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentBatch decodes a stored record.
func ParsePaymentBatch(f codec.Fields) (PaymentBatch, error) {
	var x PaymentBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentbatch: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("payment: paymentbatch %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("payment: paymentbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentBatch) String() string {
	return fmt.Sprintf("PaymentBatch(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentBatch) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentBatch) Normalize() PaymentBatch {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortPaymentBatchs orders by capacity descending, then by ID.
func SortPaymentBatchs(xs []PaymentBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentBatchs drops the archived entries, preserving order.
func ActivePaymentBatchs(xs []PaymentBatch) []PaymentBatch {
	out := make([]PaymentBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentBatchs adds up the capacity of every active entry.
func SumPaymentBatchs(xs []PaymentBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPaymentBatchs keys the entries by ID, last write winning.
func IndexPaymentBatchs(xs []PaymentBatch) map[string]PaymentBatch {
	out := make(map[string]PaymentBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentBatch builds a valid PaymentBatch, for tests and for the seed data the cron
// binary writes.
func SamplePaymentBatch(id string, at time.Time) PaymentBatch {
	return PaymentBatch{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Name:      "seeded by payment.SamplePaymentBatch",
		Capacity:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PaymentRequestSchema is the stored shape of a PaymentRequest.
var PaymentRequestSchema = schema.New("payment.paymentrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentRequest is one payment request.
type PaymentRequest struct {
	ID       string
	Source   string
	Comment  string
	Name     string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentRequest) Key() string { return "payment/paymentrequest/" + x.ID }

// Fields encodes x for storage.
func (x PaymentRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"name":     x.Name,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentRequest decodes a stored record.
func ParsePaymentRequest(f codec.Fields) (PaymentRequest, error) {
	var x PaymentRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentrequest: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentRequest) String() string {
	return fmt.Sprintf("PaymentRequest(%s, source=%q, offset=%d, archived=%t)", x.ID, x.Source, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentRequest) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentRequest) Normalize() PaymentRequest {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPaymentRequests orders by offset descending, then by ID.
func SortPaymentRequests(xs []PaymentRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentRequests drops the archived entries, preserving order.
func ActivePaymentRequests(xs []PaymentRequest) []PaymentRequest {
	out := make([]PaymentRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentRequests adds up the offset of every active entry.
func SumPaymentRequests(xs []PaymentRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexPaymentRequests keys the entries by ID, last write winning.
func IndexPaymentRequests(xs []PaymentRequest) map[string]PaymentRequest {
	out := make(map[string]PaymentRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentRequest builds a valid PaymentRequest, for tests and for the seed data the cron
// binary writes.
func SamplePaymentRequest(id string, at time.Time) PaymentRequest {
	return PaymentRequest{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Name:     "seeded by payment.SamplePaymentRequest",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentSummarySchema is the stored shape of a PaymentSummary.
var PaymentSummarySchema = schema.New("payment.paymentsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentSummary is one payment summary.
type PaymentSummary struct {
	ID        string
	Owner     string
	Source    string
	Reference string
	Retained  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PaymentSummary) Key() string { return "payment/paymentsummary/" + x.ID }

// Fields encodes x for storage.
func (x PaymentSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"source":    x.Source,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentSummary decodes a stored record.
func ParsePaymentSummary(f codec.Fields) (PaymentSummary, error) {
	var x PaymentSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentsummary: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("payment: paymentsummary %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentSummary) String() string {
	return fmt.Sprintf("PaymentSummary(%s, owner=%q, retained=%d, archived=%t)", x.ID, x.Owner, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentSummary) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentSummary) Normalize() PaymentSummary {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPaymentSummarys orders by retained descending, then by ID.
func SortPaymentSummarys(xs []PaymentSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentSummarys drops the archived entries, preserving order.
func ActivePaymentSummarys(xs []PaymentSummary) []PaymentSummary {
	out := make([]PaymentSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentSummarys adds up the retained of every active entry.
func SumPaymentSummarys(xs []PaymentSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPaymentSummarys keys the entries by ID, last write winning.
func IndexPaymentSummarys(xs []PaymentSummary) map[string]PaymentSummary {
	out := make(map[string]PaymentSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentSummary builds a valid PaymentSummary, for tests and for the seed data the cron
// binary writes.
func SamplePaymentSummary(id string, at time.Time) PaymentSummary {
	return PaymentSummary{
		ID:        id,
		Owner:     "owner-" + id,
		Source:    "source-" + id,
		Reference: "seeded by payment.SamplePaymentSummary",
		Retained:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PaymentPolicySchema is the stored shape of a PaymentPolicy.
var PaymentPolicySchema = schema.New("payment.paymentpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentPolicy is one payment policy.
type PaymentPolicy struct {
	ID        string
	Comment   string
	Name      string
	Reference string
	Weight    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PaymentPolicy) Key() string { return "payment/paymentpolicy/" + x.ID }

// Fields encodes x for storage.
func (x PaymentPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"name":      x.Name,
		"reference": x.Reference,
		"weight":    strconv.Itoa(x.Weight),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentPolicy decodes a stored record.
func ParsePaymentPolicy(f codec.Fields) (PaymentPolicy, error) {
	var x PaymentPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentpolicy: missing id")
	}
	x.Comment = f["comment"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("payment: paymentpolicy %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("payment: paymentpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentPolicy) String() string {
	return fmt.Sprintf("PaymentPolicy(%s, comment=%q, weight=%d, archived=%t)", x.ID, x.Comment, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentPolicy) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentPolicy) Normalize() PaymentPolicy {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortPaymentPolicys orders by weight descending, then by ID.
func SortPaymentPolicys(xs []PaymentPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentPolicys drops the archived entries, preserving order.
func ActivePaymentPolicys(xs []PaymentPolicy) []PaymentPolicy {
	out := make([]PaymentPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentPolicys adds up the weight of every active entry.
func SumPaymentPolicys(xs []PaymentPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexPaymentPolicys keys the entries by ID, last write winning.
func IndexPaymentPolicys(xs []PaymentPolicy) map[string]PaymentPolicy {
	out := make(map[string]PaymentPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentPolicy builds a valid PaymentPolicy, for tests and for the seed data the cron
// binary writes.
func SamplePaymentPolicy(id string, at time.Time) PaymentPolicy {
	return PaymentPolicy{
		ID:        id,
		Comment:   "comment-" + id,
		Name:      "name-" + id,
		Reference: "seeded by payment.SamplePaymentPolicy",
		Weight:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PaymentLinkSchema is the stored shape of a PaymentLink.
var PaymentLinkSchema = schema.New("payment.paymentlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentLink is one payment link.
type PaymentLink struct {
	ID        string
	Label     string
	Name      string
	Reference string
	Weight    int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x PaymentLink) Key() string { return "payment/paymentlink/" + x.ID }

// Fields encodes x for storage.
func (x PaymentLink) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"name":      x.Name,
		"reference": x.Reference,
		"weight":    strconv.Itoa(x.Weight),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentLink decodes a stored record.
func ParsePaymentLink(f codec.Fields) (PaymentLink, error) {
	var x PaymentLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentlink: missing id")
	}
	x.Label = f["label"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("payment: paymentlink %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("payment: paymentlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentLink) String() string {
	return fmt.Sprintf("PaymentLink(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentLink) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentLink) Normalize() PaymentLink {
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPaymentLinks orders by weight descending, then by ID.
func SortPaymentLinks(xs []PaymentLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentLinks drops the archived entries, preserving order.
func ActivePaymentLinks(xs []PaymentLink) []PaymentLink {
	out := make([]PaymentLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentLinks adds up the weight of every active entry.
func SumPaymentLinks(xs []PaymentLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexPaymentLinks keys the entries by ID, last write winning.
func IndexPaymentLinks(xs []PaymentLink) map[string]PaymentLink {
	out := make(map[string]PaymentLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentLink builds a valid PaymentLink, for tests and for the seed data the cron
// binary writes.
func SamplePaymentLink(id string, at time.Time) PaymentLink {
	return PaymentLink{
		ID:        id,
		Label:     "label-" + id,
		Name:      "name-" + id,
		Reference: "seeded by payment.SamplePaymentLink",
		Weight:    len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// PaymentNoteSchema is the stored shape of a PaymentNote.
var PaymentNoteSchema = schema.New("payment.paymentnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentNote is one payment note.
type PaymentNote struct {
	ID       string
	Comment  string
	Group    string
	Owner    string
	Count    int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentNote) Key() string { return "payment/paymentnote/" + x.ID }

// Fields encodes x for storage.
func (x PaymentNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"owner":    x.Owner,
		"count":    strconv.Itoa(x.Count),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentNote decodes a stored record.
func ParsePaymentNote(f codec.Fields) (PaymentNote, error) {
	var x PaymentNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentnote: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("payment: paymentnote %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("payment: paymentnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentNote) String() string {
	return fmt.Sprintf("PaymentNote(%s, comment=%q, count=%d, archived=%t)", x.ID, x.Comment, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentNote) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentNote) Normalize() PaymentNote {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortPaymentNotes orders by count descending, then by ID.
func SortPaymentNotes(xs []PaymentNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentNotes drops the archived entries, preserving order.
func ActivePaymentNotes(xs []PaymentNote) []PaymentNote {
	out := make([]PaymentNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentNotes adds up the count of every active entry.
func SumPaymentNotes(xs []PaymentNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPaymentNotes keys the entries by ID, last write winning.
func IndexPaymentNotes(xs []PaymentNote) map[string]PaymentNote {
	out := make(map[string]PaymentNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentNote builds a valid PaymentNote, for tests and for the seed data the cron
// binary writes.
func SamplePaymentNote(id string, at time.Time) PaymentNote {
	return PaymentNote{
		ID:      id,
		Comment: "comment-" + id,
		Group:   "group-" + id,
		Owner:   "seeded by payment.SamplePaymentNote",
		Count:   len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// PaymentWindowSchema is the stored shape of a PaymentWindow.
var PaymentWindowSchema = schema.New("payment.paymentwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentWindow is one payment window.
type PaymentWindow struct {
	ID       string
	Name     string
	Channel  string
	Owner    string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentWindow) Key() string { return "payment/paymentwindow/" + x.ID }

// Fields encodes x for storage.
func (x PaymentWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentWindow decodes a stored record.
func ParsePaymentWindow(f codec.Fields) (PaymentWindow, error) {
	var x PaymentWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentwindow: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentwindow %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("payment: paymentwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentWindow) String() string {
	return fmt.Sprintf("PaymentWindow(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentWindow) Normalize() PaymentWindow {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortPaymentWindows orders by priority descending, then by ID.
func SortPaymentWindows(xs []PaymentWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentWindows drops the archived entries, preserving order.
func ActivePaymentWindows(xs []PaymentWindow) []PaymentWindow {
	out := make([]PaymentWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentWindows adds up the priority of every active entry.
func SumPaymentWindows(xs []PaymentWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexPaymentWindows keys the entries by ID, last write winning.
func IndexPaymentWindows(xs []PaymentWindow) map[string]PaymentWindow {
	out := make(map[string]PaymentWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentWindow builds a valid PaymentWindow, for tests and for the seed data the cron
// binary writes.
func SamplePaymentWindow(id string, at time.Time) PaymentWindow {
	return PaymentWindow{
		ID:       id,
		Name:     "name-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by payment.SamplePaymentWindow",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentQuotaSchema is the stored shape of a PaymentQuota.
var PaymentQuotaSchema = schema.New("payment.paymentquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentQuota is one payment quota.
type PaymentQuota struct {
	ID       string
	Name     string
	Owner    string
	Category string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentQuota) Key() string { return "payment/paymentquota/" + x.ID }

// Fields encodes x for storage.
func (x PaymentQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentQuota decodes a stored record.
func ParsePaymentQuota(f codec.Fields) (PaymentQuota, error) {
	var x PaymentQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentquota: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("payment: paymentquota %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentQuota) String() string {
	return fmt.Sprintf("PaymentQuota(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentQuota) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentQuota) Normalize() PaymentQuota {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPaymentQuotas orders by count descending, then by ID.
func SortPaymentQuotas(xs []PaymentQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentQuotas drops the archived entries, preserving order.
func ActivePaymentQuotas(xs []PaymentQuota) []PaymentQuota {
	out := make([]PaymentQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentQuotas adds up the count of every active entry.
func SumPaymentQuotas(xs []PaymentQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexPaymentQuotas keys the entries by ID, last write winning.
func IndexPaymentQuotas(xs []PaymentQuota) map[string]PaymentQuota {
	out := make(map[string]PaymentQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentQuota builds a valid PaymentQuota, for tests and for the seed data the cron
// binary writes.
func SamplePaymentQuota(id string, at time.Time) PaymentQuota {
	return PaymentQuota{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by payment.SamplePaymentQuota",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentBucketSchema is the stored shape of a PaymentBucket.
var PaymentBucketSchema = schema.New("payment.paymentbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentBucket is one payment bucket.
type PaymentBucket struct {
	ID       string
	Label    string
	Channel  string
	Group    string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentBucket) Key() string { return "payment/paymentbucket/" + x.ID }

// Fields encodes x for storage.
func (x PaymentBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentBucket decodes a stored record.
func ParsePaymentBucket(f codec.Fields) (PaymentBucket, error) {
	var x PaymentBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentbucket: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("payment: paymentbucket %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentBucket) String() string {
	return fmt.Sprintf("PaymentBucket(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentBucket) Normalize() PaymentBucket {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPaymentBuckets orders by retained descending, then by ID.
func SortPaymentBuckets(xs []PaymentBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentBuckets drops the archived entries, preserving order.
func ActivePaymentBuckets(xs []PaymentBucket) []PaymentBucket {
	out := make([]PaymentBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentBuckets adds up the retained of every active entry.
func SumPaymentBuckets(xs []PaymentBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexPaymentBuckets keys the entries by ID, last write winning.
func IndexPaymentBuckets(xs []PaymentBucket) map[string]PaymentBucket {
	out := make(map[string]PaymentBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentBucket builds a valid PaymentBucket, for tests and for the seed data the cron
// binary writes.
func SamplePaymentBucket(id string, at time.Time) PaymentBucket {
	return PaymentBucket{
		ID:       id,
		Label:    "label-" + id,
		Channel:  "channel-" + id,
		Group:    "seeded by payment.SamplePaymentBucket",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentLaneSchema is the stored shape of a PaymentLane.
var PaymentLaneSchema = schema.New("payment.paymentlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentLane is one payment lane.
type PaymentLane struct {
	ID       string
	Group    string
	Source   string
	Label    string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentLane) Key() string { return "payment/paymentlane/" + x.ID }

// Fields encodes x for storage.
func (x PaymentLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentLane decodes a stored record.
func ParsePaymentLane(f codec.Fields) (PaymentLane, error) {
	var x PaymentLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentlane: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentlane %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("payment: paymentlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentLane) String() string {
	return fmt.Sprintf("PaymentLane(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentLane) Normalize() PaymentLane {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortPaymentLanes orders by priority descending, then by ID.
func SortPaymentLanes(xs []PaymentLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentLanes drops the archived entries, preserving order.
func ActivePaymentLanes(xs []PaymentLane) []PaymentLane {
	out := make([]PaymentLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentLanes adds up the priority of every active entry.
func SumPaymentLanes(xs []PaymentLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexPaymentLanes keys the entries by ID, last write winning.
func IndexPaymentLanes(xs []PaymentLane) map[string]PaymentLane {
	out := make(map[string]PaymentLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentLane builds a valid PaymentLane, for tests and for the seed data the cron
// binary writes.
func SamplePaymentLane(id string, at time.Time) PaymentLane {
	return PaymentLane{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Label:    "seeded by payment.SamplePaymentLane",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentSlotSchema is the stored shape of a PaymentSlot.
var PaymentSlotSchema = schema.New("payment.paymentslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentSlot is one payment slot.
type PaymentSlot struct {
	ID       string
	Comment  string
	Name     string
	Category string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentSlot) Key() string { return "payment/paymentslot/" + x.ID }

// Fields encodes x for storage.
func (x PaymentSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"name":     x.Name,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentSlot decodes a stored record.
func ParsePaymentSlot(f codec.Fields) (PaymentSlot, error) {
	var x PaymentSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentslot: missing id")
	}
	x.Comment = f["comment"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("payment: paymentslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("payment: paymentslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentSlot) String() string {
	return fmt.Sprintf("PaymentSlot(%s, comment=%q, capacity=%d, archived=%t)", x.ID, x.Comment, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentSlot) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentSlot) Normalize() PaymentSlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortPaymentSlots orders by capacity descending, then by ID.
func SortPaymentSlots(xs []PaymentSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentSlots drops the archived entries, preserving order.
func ActivePaymentSlots(xs []PaymentSlot) []PaymentSlot {
	out := make([]PaymentSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentSlots adds up the capacity of every active entry.
func SumPaymentSlots(xs []PaymentSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPaymentSlots keys the entries by ID, last write winning.
func IndexPaymentSlots(xs []PaymentSlot) map[string]PaymentSlot {
	out := make(map[string]PaymentSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentSlot builds a valid PaymentSlot, for tests and for the seed data the cron
// binary writes.
func SamplePaymentSlot(id string, at time.Time) PaymentSlot {
	return PaymentSlot{
		ID:       id,
		Comment:  "comment-" + id,
		Name:     "name-" + id,
		Category: "seeded by payment.SamplePaymentSlot",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentTagSchema is the stored shape of a PaymentTag.
var PaymentTagSchema = schema.New("payment.paymenttag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentTag is one payment tag.
type PaymentTag struct {
	ID       string
	Channel  string
	Owner    string
	Category string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentTag) Key() string { return "payment/paymenttag/" + x.ID }

// Fields encodes x for storage.
func (x PaymentTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentTag decodes a stored record.
func ParsePaymentTag(f codec.Fields) (PaymentTag, error) {
	var x PaymentTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymenttag: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("payment: paymenttag %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("payment: paymenttag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymenttag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymenttag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentTag) String() string {
	return fmt.Sprintf("PaymentTag(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentTag) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentTag) Normalize() PaymentTag {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortPaymentTags orders by version descending, then by ID.
func SortPaymentTags(xs []PaymentTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentTags drops the archived entries, preserving order.
func ActivePaymentTags(xs []PaymentTag) []PaymentTag {
	out := make([]PaymentTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentTags adds up the version of every active entry.
func SumPaymentTags(xs []PaymentTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPaymentTags keys the entries by ID, last write winning.
func IndexPaymentTags(xs []PaymentTag) map[string]PaymentTag {
	out := make(map[string]PaymentTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentTag builds a valid PaymentTag, for tests and for the seed data the cron
// binary writes.
func SamplePaymentTag(id string, at time.Time) PaymentTag {
	return PaymentTag{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by payment.SamplePaymentTag",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentRuleSchema is the stored shape of a PaymentRule.
var PaymentRuleSchema = schema.New("payment.paymentrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentRule is one payment rule.
type PaymentRule struct {
	ID       string
	Group    string
	Category string
	Name     string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentRule) Key() string { return "payment/paymentrule/" + x.ID }

// Fields encodes x for storage.
func (x PaymentRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"name":     x.Name,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentRule decodes a stored record.
func ParsePaymentRule(f codec.Fields) (PaymentRule, error) {
	var x PaymentRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentrule: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrule %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("payment: paymentrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentRule) String() string {
	return fmt.Sprintf("PaymentRule(%s, group=%q, weight=%d, archived=%t)", x.ID, x.Group, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentRule) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentRule) Normalize() PaymentRule {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortPaymentRules orders by weight descending, then by ID.
func SortPaymentRules(xs []PaymentRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentRules drops the archived entries, preserving order.
func ActivePaymentRules(xs []PaymentRule) []PaymentRule {
	out := make([]PaymentRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentRules adds up the weight of every active entry.
func SumPaymentRules(xs []PaymentRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexPaymentRules keys the entries by ID, last write winning.
func IndexPaymentRules(xs []PaymentRule) map[string]PaymentRule {
	out := make(map[string]PaymentRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentRule builds a valid PaymentRule, for tests and for the seed data the cron
// binary writes.
func SamplePaymentRule(id string, at time.Time) PaymentRule {
	return PaymentRule{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Name:     "seeded by payment.SamplePaymentRule",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentHoldSchema is the stored shape of a PaymentHold.
var PaymentHoldSchema = schema.New("payment.paymenthold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentHold is one payment hold.
type PaymentHold struct {
	ID       string
	Comment  string
	Group    string
	Channel  string
	Weight   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentHold) Key() string { return "payment/paymenthold/" + x.ID }

// Fields encodes x for storage.
func (x PaymentHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentHold decodes a stored record.
func ParsePaymentHold(f codec.Fields) (PaymentHold, error) {
	var x PaymentHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymenthold: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("payment: paymenthold %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("payment: paymenthold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymenthold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymenthold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentHold) String() string {
	return fmt.Sprintf("PaymentHold(%s, comment=%q, weight=%d, archived=%t)", x.ID, x.Comment, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentHold) Normalize() PaymentHold {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortPaymentHolds orders by weight descending, then by ID.
func SortPaymentHolds(xs []PaymentHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentHolds drops the archived entries, preserving order.
func ActivePaymentHolds(xs []PaymentHold) []PaymentHold {
	out := make([]PaymentHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentHolds adds up the weight of every active entry.
func SumPaymentHolds(xs []PaymentHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexPaymentHolds keys the entries by ID, last write winning.
func IndexPaymentHolds(xs []PaymentHold) map[string]PaymentHold {
	out := make(map[string]PaymentHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentHold builds a valid PaymentHold, for tests and for the seed data the cron
// binary writes.
func SamplePaymentHold(id string, at time.Time) PaymentHold {
	return PaymentHold{
		ID:       id,
		Comment:  "comment-" + id,
		Group:    "group-" + id,
		Channel:  "seeded by payment.SamplePaymentHold",
		Weight:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentClaimSchema is the stored shape of a PaymentClaim.
var PaymentClaimSchema = schema.New("payment.paymentclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentClaim is one payment claim.
type PaymentClaim struct {
	ID       string
	Owner    string
	Group    string
	Label    string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentClaim) Key() string { return "payment/paymentclaim/" + x.ID }

// Fields encodes x for storage.
func (x PaymentClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentClaim decodes a stored record.
func ParsePaymentClaim(f codec.Fields) (PaymentClaim, error) {
	var x PaymentClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("payment: paymentclaim %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("payment: paymentclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentClaim) String() string {
	return fmt.Sprintf("PaymentClaim(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentClaim) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentClaim) Normalize() PaymentClaim {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortPaymentClaims orders by capacity descending, then by ID.
func SortPaymentClaims(xs []PaymentClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentClaims drops the archived entries, preserving order.
func ActivePaymentClaims(xs []PaymentClaim) []PaymentClaim {
	out := make([]PaymentClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentClaims adds up the capacity of every active entry.
func SumPaymentClaims(xs []PaymentClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexPaymentClaims keys the entries by ID, last write winning.
func IndexPaymentClaims(xs []PaymentClaim) map[string]PaymentClaim {
	out := make(map[string]PaymentClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentClaim builds a valid PaymentClaim, for tests and for the seed data the cron
// binary writes.
func SamplePaymentClaim(id string, at time.Time) PaymentClaim {
	return PaymentClaim{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Label:    "seeded by payment.SamplePaymentClaim",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// PaymentCursorSchema is the stored shape of a PaymentCursor.
var PaymentCursorSchema = schema.New("payment.paymentcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// PaymentCursor is one payment cursor.
type PaymentCursor struct {
	ID       string
	Comment  string
	Channel  string
	Source   string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x PaymentCursor) Key() string { return "payment/paymentcursor/" + x.ID }

// Fields encodes x for storage.
func (x PaymentCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParsePaymentCursor decodes a stored record.
func ParsePaymentCursor(f codec.Fields) (PaymentCursor, error) {
	var x PaymentCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("payment: paymentcursor: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("payment: paymentcursor %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("payment: paymentcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("payment: paymentcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("payment: paymentcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x PaymentCursor) String() string {
	return fmt.Sprintf("PaymentCursor(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x PaymentCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x PaymentCursor) Normalize() PaymentCursor {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortPaymentCursors orders by version descending, then by ID.
func SortPaymentCursors(xs []PaymentCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActivePaymentCursors drops the archived entries, preserving order.
func ActivePaymentCursors(xs []PaymentCursor) []PaymentCursor {
	out := make([]PaymentCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumPaymentCursors adds up the version of every active entry.
func SumPaymentCursors(xs []PaymentCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexPaymentCursors keys the entries by ID, last write winning.
func IndexPaymentCursors(xs []PaymentCursor) map[string]PaymentCursor {
	out := make(map[string]PaymentCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SamplePaymentCursor builds a valid PaymentCursor, for tests and for the seed data the cron
// binary writes.
func SamplePaymentCursor(id string, at time.Time) PaymentCursor {
	return PaymentCursor{
		ID:      id,
		Comment: "comment-" + id,
		Channel: "channel-" + id,
		Source:  "seeded by payment.SamplePaymentCursor",
		Version: len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	PaymentRecordSchema,
	PaymentEntrySchema,
	PaymentItemSchema,
	PaymentDraftSchema,
	PaymentRevisionSchema,
	PaymentSnapshotSchema,
	PaymentBatchSchema,
	PaymentRequestSchema,
	PaymentSummarySchema,
	PaymentPolicySchema,
	PaymentLinkSchema,
	PaymentNoteSchema,
	PaymentWindowSchema,
	PaymentQuotaSchema,
	PaymentBucketSchema,
	PaymentLaneSchema,
	PaymentSlotSchema,
	PaymentTagSchema,
	PaymentRuleSchema,
	PaymentHoldSchema,
	PaymentClaimSchema,
	PaymentCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"payment/paymentrecord/",
	"payment/paymententry/",
	"payment/paymentitem/",
	"payment/paymentdraft/",
	"payment/paymentrevision/",
	"payment/paymentsnapshot/",
	"payment/paymentbatch/",
	"payment/paymentrequest/",
	"payment/paymentsummary/",
	"payment/paymentpolicy/",
	"payment/paymentlink/",
	"payment/paymentnote/",
	"payment/paymentwindow/",
	"payment/paymentquota/",
	"payment/paymentbucket/",
	"payment/paymentlane/",
	"payment/paymentslot/",
	"payment/paymenttag/",
	"payment/paymentrule/",
	"payment/paymenthold/",
	"payment/paymentclaim/",
	"payment/paymentcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "payment: " + strings.Join(names, ", ")
}
