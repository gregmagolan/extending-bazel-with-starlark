// Package order holds the order domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package order

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// OrderRecordSchema is the stored shape of a OrderRecord.
var OrderRecordSchema = schema.New("order.orderrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderRecord is one order record.
type OrderRecord struct {
	ID       string
	Channel  string
	Category string
	Label    string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderRecord) Key() string { return "order/orderrecord/" + x.ID }

// Fields encodes x for storage.
func (x OrderRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"label":    x.Label,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderRecord decodes a stored record.
func ParseOrderRecord(f codec.Fields) (OrderRecord, error) {
	var x OrderRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderrecord: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("order: orderrecord %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: orderrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderRecord) String() string {
	return fmt.Sprintf("OrderRecord(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderRecord) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderRecord) Normalize() OrderRecord {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortOrderRecords orders by attempts descending, then by ID.
func SortOrderRecords(xs []OrderRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderRecords drops the archived entries, preserving order.
func ActiveOrderRecords(xs []OrderRecord) []OrderRecord {
	out := make([]OrderRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderRecords adds up the attempts of every active entry.
func SumOrderRecords(xs []OrderRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexOrderRecords keys the entries by ID, last write winning.
func IndexOrderRecords(xs []OrderRecord) map[string]OrderRecord {
	out := make(map[string]OrderRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderRecord builds a valid OrderRecord, for tests and for the seed data the cron
// binary writes.
func SampleOrderRecord(id string, at time.Time) OrderRecord {
	return OrderRecord{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Label:    "seeded by order.SampleOrderRecord",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderEntrySchema is the stored shape of a OrderEntry.
var OrderEntrySchema = schema.New("order.orderentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderEntry is one order entry.
type OrderEntry struct {
	ID       string
	Comment  string
	Region   string
	Channel  string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderEntry) Key() string { return "order/orderentry/" + x.ID }

// Fields encodes x for storage.
func (x OrderEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"region":   x.Region,
		"channel":  x.Channel,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderEntry decodes a stored record.
func ParseOrderEntry(f codec.Fields) (OrderEntry, error) {
	var x OrderEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderentry: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: orderentry %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("order: orderentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderEntry) String() string {
	return fmt.Sprintf("OrderEntry(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderEntry) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderEntry) Normalize() OrderEntry {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortOrderEntrys orders by retained descending, then by ID.
func SortOrderEntrys(xs []OrderEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderEntrys drops the archived entries, preserving order.
func ActiveOrderEntrys(xs []OrderEntry) []OrderEntry {
	out := make([]OrderEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderEntrys adds up the retained of every active entry.
func SumOrderEntrys(xs []OrderEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexOrderEntrys keys the entries by ID, last write winning.
func IndexOrderEntrys(xs []OrderEntry) map[string]OrderEntry {
	out := make(map[string]OrderEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderEntry builds a valid OrderEntry, for tests and for the seed data the cron
// binary writes.
func SampleOrderEntry(id string, at time.Time) OrderEntry {
	return OrderEntry{
		ID:       id,
		Comment:  "comment-" + id,
		Region:   "region-" + id,
		Channel:  "seeded by order.SampleOrderEntry",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderItemSchema is the stored shape of a OrderItem.
var OrderItemSchema = schema.New("order.orderitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderItem is one order item.
type OrderItem struct {
	ID       string
	Label    string
	Region   string
	Owner    string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderItem) Key() string { return "order/orderitem/" + x.ID }

// Fields encodes x for storage.
func (x OrderItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"owner":    x.Owner,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderItem decodes a stored record.
func ParseOrderItem(f codec.Fields) (OrderItem, error) {
	var x OrderItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderitem: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("order: orderitem %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("order: orderitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderItem) String() string {
	return fmt.Sprintf("OrderItem(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderItem) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderItem) Normalize() OrderItem {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortOrderItems orders by priority descending, then by ID.
func SortOrderItems(xs []OrderItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderItems drops the archived entries, preserving order.
func ActiveOrderItems(xs []OrderItem) []OrderItem {
	out := make([]OrderItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderItems adds up the priority of every active entry.
func SumOrderItems(xs []OrderItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexOrderItems keys the entries by ID, last write winning.
func IndexOrderItems(xs []OrderItem) map[string]OrderItem {
	out := make(map[string]OrderItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderItem builds a valid OrderItem, for tests and for the seed data the cron
// binary writes.
func SampleOrderItem(id string, at time.Time) OrderItem {
	return OrderItem{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Owner:    "seeded by order.SampleOrderItem",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderDraftSchema is the stored shape of a OrderDraft.
var OrderDraftSchema = schema.New("order.orderdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderDraft is one order draft.
type OrderDraft struct {
	ID       string
	Channel  string
	Category string
	Comment  string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderDraft) Key() string { return "order/orderdraft/" + x.ID }

// Fields encodes x for storage.
func (x OrderDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"category": x.Category,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderDraft decodes a stored record.
func ParseOrderDraft(f codec.Fields) (OrderDraft, error) {
	var x OrderDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderdraft: missing id")
	}
	x.Channel = f["channel"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("order: orderdraft %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: orderdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderDraft) String() string {
	return fmt.Sprintf("OrderDraft(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderDraft) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderDraft) Normalize() OrderDraft {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortOrderDrafts orders by priority descending, then by ID.
func SortOrderDrafts(xs []OrderDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderDrafts drops the archived entries, preserving order.
func ActiveOrderDrafts(xs []OrderDraft) []OrderDraft {
	out := make([]OrderDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderDrafts adds up the priority of every active entry.
func SumOrderDrafts(xs []OrderDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexOrderDrafts keys the entries by ID, last write winning.
func IndexOrderDrafts(xs []OrderDraft) map[string]OrderDraft {
	out := make(map[string]OrderDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderDraft builds a valid OrderDraft, for tests and for the seed data the cron
// binary writes.
func SampleOrderDraft(id string, at time.Time) OrderDraft {
	return OrderDraft{
		ID:       id,
		Channel:  "channel-" + id,
		Category: "category-" + id,
		Comment:  "seeded by order.SampleOrderDraft",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderRevisionSchema is the stored shape of a OrderRevision.
var OrderRevisionSchema = schema.New("order.orderrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderRevision is one order revision.
type OrderRevision struct {
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
func (x OrderRevision) Key() string { return "order/orderrevision/" + x.ID }

// Fields encodes x for storage.
func (x OrderRevision) Fields() codec.Fields {
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

// ParseOrderRevision decodes a stored record.
func ParseOrderRevision(f codec.Fields) (OrderRevision, error) {
	var x OrderRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderrevision: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("order: orderrevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("order: orderrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderRevision) String() string {
	return fmt.Sprintf("OrderRevision(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderRevision) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderRevision) Normalize() OrderRevision {
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

// SortOrderRevisions orders by count descending, then by ID.
func SortOrderRevisions(xs []OrderRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderRevisions drops the archived entries, preserving order.
func ActiveOrderRevisions(xs []OrderRevision) []OrderRevision {
	out := make([]OrderRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderRevisions adds up the count of every active entry.
func SumOrderRevisions(xs []OrderRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexOrderRevisions keys the entries by ID, last write winning.
func IndexOrderRevisions(xs []OrderRevision) map[string]OrderRevision {
	out := make(map[string]OrderRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderRevision builds a valid OrderRevision, for tests and for the seed data the cron
// binary writes.
func SampleOrderRevision(id string, at time.Time) OrderRevision {
	return OrderRevision{
		ID:        id,
		Category:  "category-" + id,
		Group:     "group-" + id,
		Reference: "seeded by order.SampleOrderRevision",
		Count:     len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// OrderSnapshotSchema is the stored shape of a OrderSnapshot.
var OrderSnapshotSchema = schema.New("order.ordersnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderSnapshot is one order snapshot.
type OrderSnapshot struct {
	ID       string
	Region   string
	Channel  string
	Name     string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderSnapshot) Key() string { return "order/ordersnapshot/" + x.ID }

// Fields encodes x for storage.
func (x OrderSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderSnapshot decodes a stored record.
func ParseOrderSnapshot(f codec.Fields) (OrderSnapshot, error) {
	var x OrderSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: ordersnapshot: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: ordersnapshot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("order: ordersnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: ordersnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: ordersnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderSnapshot) String() string {
	return fmt.Sprintf("OrderSnapshot(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderSnapshot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderSnapshot) Normalize() OrderSnapshot {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortOrderSnapshots orders by retained descending, then by ID.
func SortOrderSnapshots(xs []OrderSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderSnapshots drops the archived entries, preserving order.
func ActiveOrderSnapshots(xs []OrderSnapshot) []OrderSnapshot {
	out := make([]OrderSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderSnapshots adds up the retained of every active entry.
func SumOrderSnapshots(xs []OrderSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexOrderSnapshots keys the entries by ID, last write winning.
func IndexOrderSnapshots(xs []OrderSnapshot) map[string]OrderSnapshot {
	out := make(map[string]OrderSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderSnapshot builds a valid OrderSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleOrderSnapshot(id string, at time.Time) OrderSnapshot {
	return OrderSnapshot{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by order.SampleOrderSnapshot",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderBatchSchema is the stored shape of a OrderBatch.
var OrderBatchSchema = schema.New("order.orderbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderBatch is one order batch.
type OrderBatch struct {
	ID       string
	Label    string
	Region   string
	Category string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderBatch) Key() string { return "order/orderbatch/" + x.ID }

// Fields encodes x for storage.
func (x OrderBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"region":   x.Region,
		"category": x.Category,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderBatch decodes a stored record.
func ParseOrderBatch(f codec.Fields) (OrderBatch, error) {
	var x OrderBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderbatch: missing id")
	}
	x.Label = f["label"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("order: orderbatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("order: orderbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderBatch) String() string {
	return fmt.Sprintf("OrderBatch(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderBatch) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderBatch) Normalize() OrderBatch {
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortOrderBatchs orders by offset descending, then by ID.
func SortOrderBatchs(xs []OrderBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderBatchs drops the archived entries, preserving order.
func ActiveOrderBatchs(xs []OrderBatch) []OrderBatch {
	out := make([]OrderBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderBatchs adds up the offset of every active entry.
func SumOrderBatchs(xs []OrderBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexOrderBatchs keys the entries by ID, last write winning.
func IndexOrderBatchs(xs []OrderBatch) map[string]OrderBatch {
	out := make(map[string]OrderBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderBatch builds a valid OrderBatch, for tests and for the seed data the cron
// binary writes.
func SampleOrderBatch(id string, at time.Time) OrderBatch {
	return OrderBatch{
		ID:       id,
		Label:    "label-" + id,
		Region:   "region-" + id,
		Category: "seeded by order.SampleOrderBatch",
		Offset:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderRequestSchema is the stored shape of a OrderRequest.
var OrderRequestSchema = schema.New("order.orderrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderRequest is one order request.
type OrderRequest struct {
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
func (x OrderRequest) Key() string { return "order/orderrequest/" + x.ID }

// Fields encodes x for storage.
func (x OrderRequest) Fields() codec.Fields {
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

// ParseOrderRequest decodes a stored record.
func ParseOrderRequest(f codec.Fields) (OrderRequest, error) {
	var x OrderRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderrequest: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("order: orderrequest %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: orderrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderRequest) String() string {
	return fmt.Sprintf("OrderRequest(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderRequest) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderRequest) Normalize() OrderRequest {
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

// SortOrderRequests orders by version descending, then by ID.
func SortOrderRequests(xs []OrderRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderRequests drops the archived entries, preserving order.
func ActiveOrderRequests(xs []OrderRequest) []OrderRequest {
	out := make([]OrderRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderRequests adds up the version of every active entry.
func SumOrderRequests(xs []OrderRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexOrderRequests keys the entries by ID, last write winning.
func IndexOrderRequests(xs []OrderRequest) map[string]OrderRequest {
	out := make(map[string]OrderRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderRequest builds a valid OrderRequest, for tests and for the seed data the cron
// binary writes.
func SampleOrderRequest(id string, at time.Time) OrderRequest {
	return OrderRequest{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by order.SampleOrderRequest",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// OrderSummarySchema is the stored shape of a OrderSummary.
var OrderSummarySchema = schema.New("order.ordersummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderSummary is one order summary.
type OrderSummary struct {
	ID        string
	Name      string
	Source    string
	Reference string
	Count     int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x OrderSummary) Key() string { return "order/ordersummary/" + x.ID }

// Fields encodes x for storage.
func (x OrderSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"source":    x.Source,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderSummary decodes a stored record.
func ParseOrderSummary(f codec.Fields) (OrderSummary, error) {
	var x OrderSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: ordersummary: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("order: ordersummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: ordersummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: ordersummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: ordersummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderSummary) String() string {
	return fmt.Sprintf("OrderSummary(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderSummary) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderSummary) Normalize() OrderSummary {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortOrderSummarys orders by count descending, then by ID.
func SortOrderSummarys(xs []OrderSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderSummarys drops the archived entries, preserving order.
func ActiveOrderSummarys(xs []OrderSummary) []OrderSummary {
	out := make([]OrderSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderSummarys adds up the count of every active entry.
func SumOrderSummarys(xs []OrderSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexOrderSummarys keys the entries by ID, last write winning.
func IndexOrderSummarys(xs []OrderSummary) map[string]OrderSummary {
	out := make(map[string]OrderSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderSummary builds a valid OrderSummary, for tests and for the seed data the cron
// binary writes.
func SampleOrderSummary(id string, at time.Time) OrderSummary {
	return OrderSummary{
		ID:        id,
		Name:      "name-" + id,
		Source:    "source-" + id,
		Reference: "seeded by order.SampleOrderSummary",
		Count:     len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// OrderPolicySchema is the stored shape of a OrderPolicy.
var OrderPolicySchema = schema.New("order.orderpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderPolicy is one order policy.
type OrderPolicy struct {
	ID       string
	Name     string
	Label    string
	Region   string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderPolicy) Key() string { return "order/orderpolicy/" + x.ID }

// Fields encodes x for storage.
func (x OrderPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderPolicy decodes a stored record.
func ParseOrderPolicy(f codec.Fields) (OrderPolicy, error) {
	var x OrderPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderpolicy: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("order: orderpolicy %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("order: orderpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderPolicy) String() string {
	return fmt.Sprintf("OrderPolicy(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderPolicy) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderPolicy) Normalize() OrderPolicy {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortOrderPolicys orders by offset descending, then by ID.
func SortOrderPolicys(xs []OrderPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderPolicys drops the archived entries, preserving order.
func ActiveOrderPolicys(xs []OrderPolicy) []OrderPolicy {
	out := make([]OrderPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderPolicys adds up the offset of every active entry.
func SumOrderPolicys(xs []OrderPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexOrderPolicys keys the entries by ID, last write winning.
func IndexOrderPolicys(xs []OrderPolicy) map[string]OrderPolicy {
	out := make(map[string]OrderPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderPolicy builds a valid OrderPolicy, for tests and for the seed data the cron
// binary writes.
func SampleOrderPolicy(id string, at time.Time) OrderPolicy {
	return OrderPolicy{
		ID:       id,
		Name:     "name-" + id,
		Label:    "label-" + id,
		Region:   "seeded by order.SampleOrderPolicy",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderLinkSchema is the stored shape of a OrderLink.
var OrderLinkSchema = schema.New("order.orderlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderLink is one order link.
type OrderLink struct {
	ID       string
	Category string
	Comment  string
	Owner    string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderLink) Key() string { return "order/orderlink/" + x.ID }

// Fields encodes x for storage.
func (x OrderLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderLink decodes a stored record.
func ParseOrderLink(f codec.Fields) (OrderLink, error) {
	var x OrderLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderlink: missing id")
	}
	x.Category = f["category"]
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: orderlink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("order: orderlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderLink) String() string {
	return fmt.Sprintf("OrderLink(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderLink) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderLink) Normalize() OrderLink {
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortOrderLinks orders by capacity descending, then by ID.
func SortOrderLinks(xs []OrderLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderLinks drops the archived entries, preserving order.
func ActiveOrderLinks(xs []OrderLink) []OrderLink {
	out := make([]OrderLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderLinks adds up the capacity of every active entry.
func SumOrderLinks(xs []OrderLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexOrderLinks keys the entries by ID, last write winning.
func IndexOrderLinks(xs []OrderLink) map[string]OrderLink {
	out := make(map[string]OrderLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderLink builds a valid OrderLink, for tests and for the seed data the cron
// binary writes.
func SampleOrderLink(id string, at time.Time) OrderLink {
	return OrderLink{
		ID:       id,
		Category: "category-" + id,
		Comment:  "comment-" + id,
		Owner:    "seeded by order.SampleOrderLink",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderNoteSchema is the stored shape of a OrderNote.
var OrderNoteSchema = schema.New("order.ordernote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderNote is one order note.
type OrderNote struct {
	ID       string
	Owner    string
	Source   string
	Region   string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderNote) Key() string { return "order/ordernote/" + x.ID }

// Fields encodes x for storage.
func (x OrderNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderNote decodes a stored record.
func ParseOrderNote(f codec.Fields) (OrderNote, error) {
	var x OrderNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: ordernote: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: ordernote %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: ordernote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: ordernote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: ordernote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderNote) String() string {
	return fmt.Sprintf("OrderNote(%s, owner=%q, capacity=%d, archived=%t)", x.ID, x.Owner, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderNote) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderNote) Normalize() OrderNote {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortOrderNotes orders by capacity descending, then by ID.
func SortOrderNotes(xs []OrderNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderNotes drops the archived entries, preserving order.
func ActiveOrderNotes(xs []OrderNote) []OrderNote {
	out := make([]OrderNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderNotes adds up the capacity of every active entry.
func SumOrderNotes(xs []OrderNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexOrderNotes keys the entries by ID, last write winning.
func IndexOrderNotes(xs []OrderNote) map[string]OrderNote {
	out := make(map[string]OrderNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderNote builds a valid OrderNote, for tests and for the seed data the cron
// binary writes.
func SampleOrderNote(id string, at time.Time) OrderNote {
	return OrderNote{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Region:   "seeded by order.SampleOrderNote",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderWindowSchema is the stored shape of a OrderWindow.
var OrderWindowSchema = schema.New("order.orderwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderWindow is one order window.
type OrderWindow struct {
	ID       string
	Owner    string
	Group    string
	Category string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderWindow) Key() string { return "order/orderwindow/" + x.ID }

// Fields encodes x for storage.
func (x OrderWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderWindow decodes a stored record.
func ParseOrderWindow(f codec.Fields) (OrderWindow, error) {
	var x OrderWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderwindow: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("order: orderwindow %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: orderwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderWindow) String() string {
	return fmt.Sprintf("OrderWindow(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderWindow) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderWindow) Normalize() OrderWindow {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortOrderWindows orders by attempts descending, then by ID.
func SortOrderWindows(xs []OrderWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderWindows drops the archived entries, preserving order.
func ActiveOrderWindows(xs []OrderWindow) []OrderWindow {
	out := make([]OrderWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderWindows adds up the attempts of every active entry.
func SumOrderWindows(xs []OrderWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexOrderWindows keys the entries by ID, last write winning.
func IndexOrderWindows(xs []OrderWindow) map[string]OrderWindow {
	out := make(map[string]OrderWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderWindow builds a valid OrderWindow, for tests and for the seed data the cron
// binary writes.
func SampleOrderWindow(id string, at time.Time) OrderWindow {
	return OrderWindow{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Category: "seeded by order.SampleOrderWindow",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderQuotaSchema is the stored shape of a OrderQuota.
var OrderQuotaSchema = schema.New("order.orderquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderQuota is one order quota.
type OrderQuota struct {
	ID        string
	Reference string
	Label     string
	Comment   string
	Priority  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x OrderQuota) Key() string { return "order/orderquota/" + x.ID }

// Fields encodes x for storage.
func (x OrderQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"comment":   x.Comment,
		"priority":  strconv.Itoa(x.Priority),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderQuota decodes a stored record.
func ParseOrderQuota(f codec.Fields) (OrderQuota, error) {
	var x OrderQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderquota: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("order: orderquota %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("order: orderquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderQuota) String() string {
	return fmt.Sprintf("OrderQuota(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderQuota) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderQuota) Normalize() OrderQuota {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortOrderQuotas orders by priority descending, then by ID.
func SortOrderQuotas(xs []OrderQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderQuotas drops the archived entries, preserving order.
func ActiveOrderQuotas(xs []OrderQuota) []OrderQuota {
	out := make([]OrderQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderQuotas adds up the priority of every active entry.
func SumOrderQuotas(xs []OrderQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexOrderQuotas keys the entries by ID, last write winning.
func IndexOrderQuotas(xs []OrderQuota) map[string]OrderQuota {
	out := make(map[string]OrderQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderQuota builds a valid OrderQuota, for tests and for the seed data the cron
// binary writes.
func SampleOrderQuota(id string, at time.Time) OrderQuota {
	return OrderQuota{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Comment:   "seeded by order.SampleOrderQuota",
		Priority:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// OrderBucketSchema is the stored shape of a OrderBucket.
var OrderBucketSchema = schema.New("order.orderbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderBucket is one order bucket.
type OrderBucket struct {
	ID       string
	Name     string
	Region   string
	Group    string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderBucket) Key() string { return "order/orderbucket/" + x.ID }

// Fields encodes x for storage.
func (x OrderBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"region":   x.Region,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderBucket decodes a stored record.
func ParseOrderBucket(f codec.Fields) (OrderBucket, error) {
	var x OrderBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderbucket: missing id")
	}
	x.Name = f["name"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("order: orderbucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: orderbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderBucket) String() string {
	return fmt.Sprintf("OrderBucket(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderBucket) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderBucket) Normalize() OrderBucket {
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortOrderBuckets orders by offset descending, then by ID.
func SortOrderBuckets(xs []OrderBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderBuckets drops the archived entries, preserving order.
func ActiveOrderBuckets(xs []OrderBucket) []OrderBucket {
	out := make([]OrderBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderBuckets adds up the offset of every active entry.
func SumOrderBuckets(xs []OrderBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexOrderBuckets keys the entries by ID, last write winning.
func IndexOrderBuckets(xs []OrderBucket) map[string]OrderBucket {
	out := make(map[string]OrderBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderBucket builds a valid OrderBucket, for tests and for the seed data the cron
// binary writes.
func SampleOrderBucket(id string, at time.Time) OrderBucket {
	return OrderBucket{
		ID:       id,
		Name:     "name-" + id,
		Region:   "region-" + id,
		Group:    "seeded by order.SampleOrderBucket",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderLaneSchema is the stored shape of a OrderLane.
var OrderLaneSchema = schema.New("order.orderlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderLane is one order lane.
type OrderLane struct {
	ID        string
	Group     string
	Label     string
	Reference string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x OrderLane) Key() string { return "order/orderlane/" + x.ID }

// Fields encodes x for storage.
func (x OrderLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"label":     x.Label,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderLane decodes a stored record.
func ParseOrderLane(f codec.Fields) (OrderLane, error) {
	var x OrderLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderlane: missing id")
	}
	x.Group = f["group"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("order: orderlane %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: orderlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderLane) String() string {
	return fmt.Sprintf("OrderLane(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderLane) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderLane) Normalize() OrderLane {
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortOrderLanes orders by offset descending, then by ID.
func SortOrderLanes(xs []OrderLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderLanes drops the archived entries, preserving order.
func ActiveOrderLanes(xs []OrderLane) []OrderLane {
	out := make([]OrderLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderLanes adds up the offset of every active entry.
func SumOrderLanes(xs []OrderLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexOrderLanes keys the entries by ID, last write winning.
func IndexOrderLanes(xs []OrderLane) map[string]OrderLane {
	out := make(map[string]OrderLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderLane builds a valid OrderLane, for tests and for the seed data the cron
// binary writes.
func SampleOrderLane(id string, at time.Time) OrderLane {
	return OrderLane{
		ID:        id,
		Group:     "group-" + id,
		Label:     "label-" + id,
		Reference: "seeded by order.SampleOrderLane",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// OrderSlotSchema is the stored shape of a OrderSlot.
var OrderSlotSchema = schema.New("order.orderslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderSlot is one order slot.
type OrderSlot struct {
	ID       string
	Region   string
	Owner    string
	Source   string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderSlot) Key() string { return "order/orderslot/" + x.ID }

// Fields encodes x for storage.
func (x OrderSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"owner":    x.Owner,
		"source":   x.Source,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderSlot decodes a stored record.
func ParseOrderSlot(f codec.Fields) (OrderSlot, error) {
	var x OrderSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderslot: missing id")
	}
	x.Region = f["region"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("order: orderslot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: orderslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderSlot) String() string {
	return fmt.Sprintf("OrderSlot(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderSlot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderSlot) Normalize() OrderSlot {
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortOrderSlots orders by count descending, then by ID.
func SortOrderSlots(xs []OrderSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderSlots drops the archived entries, preserving order.
func ActiveOrderSlots(xs []OrderSlot) []OrderSlot {
	out := make([]OrderSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderSlots adds up the count of every active entry.
func SumOrderSlots(xs []OrderSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexOrderSlots keys the entries by ID, last write winning.
func IndexOrderSlots(xs []OrderSlot) map[string]OrderSlot {
	out := make(map[string]OrderSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderSlot builds a valid OrderSlot, for tests and for the seed data the cron
// binary writes.
func SampleOrderSlot(id string, at time.Time) OrderSlot {
	return OrderSlot{
		ID:       id,
		Region:   "region-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by order.SampleOrderSlot",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderTagSchema is the stored shape of a OrderTag.
var OrderTagSchema = schema.New("order.ordertag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderTag is one order tag.
type OrderTag struct {
	ID       string
	Group    string
	Channel  string
	Category string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderTag) Key() string { return "order/ordertag/" + x.ID }

// Fields encodes x for storage.
func (x OrderTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderTag decodes a stored record.
func ParseOrderTag(f codec.Fields) (OrderTag, error) {
	var x OrderTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: ordertag: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: ordertag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: ordertag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: ordertag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: ordertag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderTag) String() string {
	return fmt.Sprintf("OrderTag(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderTag) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderTag) Normalize() OrderTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortOrderTags orders by capacity descending, then by ID.
func SortOrderTags(xs []OrderTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderTags drops the archived entries, preserving order.
func ActiveOrderTags(xs []OrderTag) []OrderTag {
	out := make([]OrderTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderTags adds up the capacity of every active entry.
func SumOrderTags(xs []OrderTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexOrderTags keys the entries by ID, last write winning.
func IndexOrderTags(xs []OrderTag) map[string]OrderTag {
	out := make(map[string]OrderTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderTag builds a valid OrderTag, for tests and for the seed data the cron
// binary writes.
func SampleOrderTag(id string, at time.Time) OrderTag {
	return OrderTag{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by order.SampleOrderTag",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderRuleSchema is the stored shape of a OrderRule.
var OrderRuleSchema = schema.New("order.orderrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderRule is one order rule.
type OrderRule struct {
	ID        string
	Reference string
	Source    string
	Name      string
	Weight    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x OrderRule) Key() string { return "order/orderrule/" + x.ID }

// Fields encodes x for storage.
func (x OrderRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"source":    x.Source,
		"name":      x.Name,
		"weight":    strconv.Itoa(x.Weight),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderRule decodes a stored record.
func ParseOrderRule(f codec.Fields) (OrderRule, error) {
	var x OrderRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderrule: missing id")
	}
	x.Reference = f["reference"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("order: orderrule %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("order: orderrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderRule) String() string {
	return fmt.Sprintf("OrderRule(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderRule) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderRule) Normalize() OrderRule {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortOrderRules orders by weight descending, then by ID.
func SortOrderRules(xs []OrderRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderRules drops the archived entries, preserving order.
func ActiveOrderRules(xs []OrderRule) []OrderRule {
	out := make([]OrderRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderRules adds up the weight of every active entry.
func SumOrderRules(xs []OrderRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexOrderRules keys the entries by ID, last write winning.
func IndexOrderRules(xs []OrderRule) map[string]OrderRule {
	out := make(map[string]OrderRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderRule builds a valid OrderRule, for tests and for the seed data the cron
// binary writes.
func SampleOrderRule(id string, at time.Time) OrderRule {
	return OrderRule{
		ID:        id,
		Reference: "reference-" + id,
		Source:    "source-" + id,
		Name:      "seeded by order.SampleOrderRule",
		Weight:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// OrderHoldSchema is the stored shape of a OrderHold.
var OrderHoldSchema = schema.New("order.orderhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderHold is one order hold.
type OrderHold struct {
	ID       string
	Name     string
	Owner    string
	Category string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderHold) Key() string { return "order/orderhold/" + x.ID }

// Fields encodes x for storage.
func (x OrderHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"owner":    x.Owner,
		"category": x.Category,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderHold decodes a stored record.
func ParseOrderHold(f codec.Fields) (OrderHold, error) {
	var x OrderHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderhold: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("order: orderhold %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("order: orderhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderHold) String() string {
	return fmt.Sprintf("OrderHold(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderHold) Normalize() OrderHold {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortOrderHolds orders by weight descending, then by ID.
func SortOrderHolds(xs []OrderHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderHolds drops the archived entries, preserving order.
func ActiveOrderHolds(xs []OrderHold) []OrderHold {
	out := make([]OrderHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderHolds adds up the weight of every active entry.
func SumOrderHolds(xs []OrderHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexOrderHolds keys the entries by ID, last write winning.
func IndexOrderHolds(xs []OrderHold) map[string]OrderHold {
	out := make(map[string]OrderHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderHold builds a valid OrderHold, for tests and for the seed data the cron
// binary writes.
func SampleOrderHold(id string, at time.Time) OrderHold {
	return OrderHold{
		ID:       id,
		Name:     "name-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by order.SampleOrderHold",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// OrderClaimSchema is the stored shape of a OrderClaim.
var OrderClaimSchema = schema.New("order.orderclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderClaim is one order claim.
type OrderClaim struct {
	ID        string
	Reference string
	Group     string
	Channel   string
	Capacity  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x OrderClaim) Key() string { return "order/orderclaim/" + x.ID }

// Fields encodes x for storage.
func (x OrderClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"channel":   x.Channel,
		"capacity":  strconv.Itoa(x.Capacity),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderClaim decodes a stored record.
func ParseOrderClaim(f codec.Fields) (OrderClaim, error) {
	var x OrderClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: orderclaim: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: orderclaim %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("order: orderclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: orderclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: orderclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderClaim) String() string {
	return fmt.Sprintf("OrderClaim(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderClaim) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderClaim) Normalize() OrderClaim {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortOrderClaims orders by capacity descending, then by ID.
func SortOrderClaims(xs []OrderClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderClaims drops the archived entries, preserving order.
func ActiveOrderClaims(xs []OrderClaim) []OrderClaim {
	out := make([]OrderClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderClaims adds up the capacity of every active entry.
func SumOrderClaims(xs []OrderClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexOrderClaims keys the entries by ID, last write winning.
func IndexOrderClaims(xs []OrderClaim) map[string]OrderClaim {
	out := make(map[string]OrderClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderClaim builds a valid OrderClaim, for tests and for the seed data the cron
// binary writes.
func SampleOrderClaim(id string, at time.Time) OrderClaim {
	return OrderClaim{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Channel:   "seeded by order.SampleOrderClaim",
		Capacity:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// OrderCursorSchema is the stored shape of a OrderCursor.
var OrderCursorSchema = schema.New("order.ordercursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// OrderCursor is one order cursor.
type OrderCursor struct {
	ID       string
	Category string
	Comment  string
	Label    string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x OrderCursor) Key() string { return "order/ordercursor/" + x.ID }

// Fields encodes x for storage.
func (x OrderCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"comment":  x.Comment,
		"label":    x.Label,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseOrderCursor decodes a stored record.
func ParseOrderCursor(f codec.Fields) (OrderCursor, error) {
	var x OrderCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("order: ordercursor: missing id")
	}
	x.Category = f["category"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("order: ordercursor %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("order: ordercursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("order: ordercursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("order: ordercursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x OrderCursor) String() string {
	return fmt.Sprintf("OrderCursor(%s, category=%q, version=%d, archived=%t)", x.ID, x.Category, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x OrderCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x OrderCursor) Normalize() OrderCursor {
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortOrderCursors orders by version descending, then by ID.
func SortOrderCursors(xs []OrderCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveOrderCursors drops the archived entries, preserving order.
func ActiveOrderCursors(xs []OrderCursor) []OrderCursor {
	out := make([]OrderCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumOrderCursors adds up the version of every active entry.
func SumOrderCursors(xs []OrderCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexOrderCursors keys the entries by ID, last write winning.
func IndexOrderCursors(xs []OrderCursor) map[string]OrderCursor {
	out := make(map[string]OrderCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleOrderCursor builds a valid OrderCursor, for tests and for the seed data the cron
// binary writes.
func SampleOrderCursor(id string, at time.Time) OrderCursor {
	return OrderCursor{
		ID:       id,
		Category: "category-" + id,
		Comment:  "comment-" + id,
		Label:    "seeded by order.SampleOrderCursor",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	OrderRecordSchema,
	OrderEntrySchema,
	OrderItemSchema,
	OrderDraftSchema,
	OrderRevisionSchema,
	OrderSnapshotSchema,
	OrderBatchSchema,
	OrderRequestSchema,
	OrderSummarySchema,
	OrderPolicySchema,
	OrderLinkSchema,
	OrderNoteSchema,
	OrderWindowSchema,
	OrderQuotaSchema,
	OrderBucketSchema,
	OrderLaneSchema,
	OrderSlotSchema,
	OrderTagSchema,
	OrderRuleSchema,
	OrderHoldSchema,
	OrderClaimSchema,
	OrderCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"order/orderrecord/",
	"order/orderentry/",
	"order/orderitem/",
	"order/orderdraft/",
	"order/orderrevision/",
	"order/ordersnapshot/",
	"order/orderbatch/",
	"order/orderrequest/",
	"order/ordersummary/",
	"order/orderpolicy/",
	"order/orderlink/",
	"order/ordernote/",
	"order/orderwindow/",
	"order/orderquota/",
	"order/orderbucket/",
	"order/orderlane/",
	"order/orderslot/",
	"order/ordertag/",
	"order/orderrule/",
	"order/orderhold/",
	"order/orderclaim/",
	"order/ordercursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "order: " + strings.Join(names, ", ")
}
