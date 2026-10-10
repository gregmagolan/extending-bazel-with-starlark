// Package invoice holds the invoice domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package invoice

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// InvoiceRecordSchema is the stored shape of a InvoiceRecord.
var InvoiceRecordSchema = schema.New("invoice.invoicerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceRecord is one invoice record.
type InvoiceRecord struct {
	ID        string
	Reference string
	Name      string
	Category  string
	Capacity  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceRecord) Key() string { return "invoice/invoicerecord/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"name":      x.Name,
		"category":  x.Category,
		"capacity":  strconv.Itoa(x.Capacity),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceRecord decodes a stored record.
func ParseInvoiceRecord(f codec.Fields) (InvoiceRecord, error) {
	var x InvoiceRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicerecord: missing id")
	}
	x.Reference = f["reference"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceRecord) String() string {
	return fmt.Sprintf("InvoiceRecord(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceRecord) Normalize() InvoiceRecord {
	x.Reference = strings.TrimSpace(x.Reference)
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

// SortInvoiceRecords orders by capacity descending, then by ID.
func SortInvoiceRecords(xs []InvoiceRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceRecords drops the archived entries, preserving order.
func ActiveInvoiceRecords(xs []InvoiceRecord) []InvoiceRecord {
	out := make([]InvoiceRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceRecords adds up the capacity of every active entry.
func SumInvoiceRecords(xs []InvoiceRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexInvoiceRecords keys the entries by ID, last write winning.
func IndexInvoiceRecords(xs []InvoiceRecord) map[string]InvoiceRecord {
	out := make(map[string]InvoiceRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceRecord builds a valid InvoiceRecord, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceRecord(id string, at time.Time) InvoiceRecord {
	return InvoiceRecord{
		ID:        id,
		Reference: "reference-" + id,
		Name:      "name-" + id,
		Category:  "seeded by invoice.SampleInvoiceRecord",
		Capacity:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceEntrySchema is the stored shape of a InvoiceEntry.
var InvoiceEntrySchema = schema.New("invoice.invoiceentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceEntry is one invoice entry.
type InvoiceEntry struct {
	ID       string
	Name     string
	Category string
	Group    string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceEntry) Key() string { return "invoice/invoiceentry/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"category": x.Category,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceEntry decodes a stored record.
func ParseInvoiceEntry(f codec.Fields) (InvoiceEntry, error) {
	var x InvoiceEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoiceentry: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoiceentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoiceentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceEntry) String() string {
	return fmt.Sprintf("InvoiceEntry(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceEntry) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceEntry) Normalize() InvoiceEntry {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortInvoiceEntrys orders by attempts descending, then by ID.
func SortInvoiceEntrys(xs []InvoiceEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceEntrys drops the archived entries, preserving order.
func ActiveInvoiceEntrys(xs []InvoiceEntry) []InvoiceEntry {
	out := make([]InvoiceEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceEntrys adds up the attempts of every active entry.
func SumInvoiceEntrys(xs []InvoiceEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexInvoiceEntrys keys the entries by ID, last write winning.
func IndexInvoiceEntrys(xs []InvoiceEntry) map[string]InvoiceEntry {
	out := make(map[string]InvoiceEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceEntry builds a valid InvoiceEntry, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceEntry(id string, at time.Time) InvoiceEntry {
	return InvoiceEntry{
		ID:       id,
		Name:     "name-" + id,
		Category: "category-" + id,
		Group:    "seeded by invoice.SampleInvoiceEntry",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceItemSchema is the stored shape of a InvoiceItem.
var InvoiceItemSchema = schema.New("invoice.invoiceitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceItem is one invoice item.
type InvoiceItem struct {
	ID       string
	Comment  string
	Source   string
	Name     string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceItem) Key() string { return "invoice/invoiceitem/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceItem decodes a stored record.
func ParseInvoiceItem(f codec.Fields) (InvoiceItem, error) {
	var x InvoiceItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoiceitem: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceitem %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoiceitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoiceitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceItem) String() string {
	return fmt.Sprintf("InvoiceItem(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceItem) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceItem) Normalize() InvoiceItem {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortInvoiceItems orders by retained descending, then by ID.
func SortInvoiceItems(xs []InvoiceItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceItems drops the archived entries, preserving order.
func ActiveInvoiceItems(xs []InvoiceItem) []InvoiceItem {
	out := make([]InvoiceItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceItems adds up the retained of every active entry.
func SumInvoiceItems(xs []InvoiceItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexInvoiceItems keys the entries by ID, last write winning.
func IndexInvoiceItems(xs []InvoiceItem) map[string]InvoiceItem {
	out := make(map[string]InvoiceItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceItem builds a valid InvoiceItem, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceItem(id string, at time.Time) InvoiceItem {
	return InvoiceItem{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Name:     "seeded by invoice.SampleInvoiceItem",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceDraftSchema is the stored shape of a InvoiceDraft.
var InvoiceDraftSchema = schema.New("invoice.invoicedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceDraft is one invoice draft.
type InvoiceDraft struct {
	ID       string
	Label    string
	Group    string
	Owner    string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceDraft) Key() string { return "invoice/invoicedraft/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceDraft decodes a stored record.
func ParseInvoiceDraft(f codec.Fields) (InvoiceDraft, error) {
	var x InvoiceDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicedraft: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicedraft %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceDraft) String() string {
	return fmt.Sprintf("InvoiceDraft(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceDraft) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceDraft) Normalize() InvoiceDraft {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortInvoiceDrafts orders by offset descending, then by ID.
func SortInvoiceDrafts(xs []InvoiceDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceDrafts drops the archived entries, preserving order.
func ActiveInvoiceDrafts(xs []InvoiceDraft) []InvoiceDraft {
	out := make([]InvoiceDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceDrafts adds up the offset of every active entry.
func SumInvoiceDrafts(xs []InvoiceDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexInvoiceDrafts keys the entries by ID, last write winning.
func IndexInvoiceDrafts(xs []InvoiceDraft) map[string]InvoiceDraft {
	out := make(map[string]InvoiceDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceDraft builds a valid InvoiceDraft, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceDraft(id string, at time.Time) InvoiceDraft {
	return InvoiceDraft{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Owner:    "seeded by invoice.SampleInvoiceDraft",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceRevisionSchema is the stored shape of a InvoiceRevision.
var InvoiceRevisionSchema = schema.New("invoice.invoicerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceRevision is one invoice revision.
type InvoiceRevision struct {
	ID        string
	Name      string
	Reference string
	Group     string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceRevision) Key() string { return "invoice/invoicerevision/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"group":     x.Group,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceRevision decodes a stored record.
func ParseInvoiceRevision(f codec.Fields) (InvoiceRevision, error) {
	var x InvoiceRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicerevision: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerevision %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceRevision) String() string {
	return fmt.Sprintf("InvoiceRevision(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceRevision) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceRevision) Normalize() InvoiceRevision {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortInvoiceRevisions orders by weight descending, then by ID.
func SortInvoiceRevisions(xs []InvoiceRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceRevisions drops the archived entries, preserving order.
func ActiveInvoiceRevisions(xs []InvoiceRevision) []InvoiceRevision {
	out := make([]InvoiceRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceRevisions adds up the weight of every active entry.
func SumInvoiceRevisions(xs []InvoiceRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexInvoiceRevisions keys the entries by ID, last write winning.
func IndexInvoiceRevisions(xs []InvoiceRevision) map[string]InvoiceRevision {
	out := make(map[string]InvoiceRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceRevision builds a valid InvoiceRevision, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceRevision(id string, at time.Time) InvoiceRevision {
	return InvoiceRevision{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Group:     "seeded by invoice.SampleInvoiceRevision",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceSnapshotSchema is the stored shape of a InvoiceSnapshot.
var InvoiceSnapshotSchema = schema.New("invoice.invoicesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceSnapshot is one invoice snapshot.
type InvoiceSnapshot struct {
	ID        string
	Reference string
	Channel   string
	Region    string
	Weight    int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceSnapshot) Key() string { return "invoice/invoicesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"region":    x.Region,
		"weight":    strconv.Itoa(x.Weight),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceSnapshot decodes a stored record.
func ParseInvoiceSnapshot(f codec.Fields) (InvoiceSnapshot, error) {
	var x InvoiceSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicesnapshot: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicesnapshot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceSnapshot) String() string {
	return fmt.Sprintf("InvoiceSnapshot(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceSnapshot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceSnapshot) Normalize() InvoiceSnapshot {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortInvoiceSnapshots orders by weight descending, then by ID.
func SortInvoiceSnapshots(xs []InvoiceSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceSnapshots drops the archived entries, preserving order.
func ActiveInvoiceSnapshots(xs []InvoiceSnapshot) []InvoiceSnapshot {
	out := make([]InvoiceSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceSnapshots adds up the weight of every active entry.
func SumInvoiceSnapshots(xs []InvoiceSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexInvoiceSnapshots keys the entries by ID, last write winning.
func IndexInvoiceSnapshots(xs []InvoiceSnapshot) map[string]InvoiceSnapshot {
	out := make(map[string]InvoiceSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceSnapshot builds a valid InvoiceSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceSnapshot(id string, at time.Time) InvoiceSnapshot {
	return InvoiceSnapshot{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Region:    "seeded by invoice.SampleInvoiceSnapshot",
		Weight:    len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceBatchSchema is the stored shape of a InvoiceBatch.
var InvoiceBatchSchema = schema.New("invoice.invoicebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceBatch is one invoice batch.
type InvoiceBatch struct {
	ID        string
	Channel   string
	Name      string
	Reference string
	Attempts  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceBatch) Key() string { return "invoice/invoicebatch/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"name":      x.Name,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceBatch decodes a stored record.
func ParseInvoiceBatch(f codec.Fields) (InvoiceBatch, error) {
	var x InvoiceBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicebatch: missing id")
	}
	x.Channel = f["channel"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicebatch %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceBatch) String() string {
	return fmt.Sprintf("InvoiceBatch(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceBatch) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceBatch) Normalize() InvoiceBatch {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortInvoiceBatchs orders by attempts descending, then by ID.
func SortInvoiceBatchs(xs []InvoiceBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceBatchs drops the archived entries, preserving order.
func ActiveInvoiceBatchs(xs []InvoiceBatch) []InvoiceBatch {
	out := make([]InvoiceBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceBatchs adds up the attempts of every active entry.
func SumInvoiceBatchs(xs []InvoiceBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexInvoiceBatchs keys the entries by ID, last write winning.
func IndexInvoiceBatchs(xs []InvoiceBatch) map[string]InvoiceBatch {
	out := make(map[string]InvoiceBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceBatch builds a valid InvoiceBatch, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceBatch(id string, at time.Time) InvoiceBatch {
	return InvoiceBatch{
		ID:        id,
		Channel:   "channel-" + id,
		Name:      "name-" + id,
		Reference: "seeded by invoice.SampleInvoiceBatch",
		Attempts:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceRequestSchema is the stored shape of a InvoiceRequest.
var InvoiceRequestSchema = schema.New("invoice.invoicerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceRequest is one invoice request.
type InvoiceRequest struct {
	ID        string
	Source    string
	Group     string
	Reference string
	Priority  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceRequest) Key() string { return "invoice/invoicerequest/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"group":     x.Group,
		"reference": x.Reference,
		"priority":  strconv.Itoa(x.Priority),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceRequest decodes a stored record.
func ParseInvoiceRequest(f codec.Fields) (InvoiceRequest, error) {
	var x InvoiceRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicerequest: missing id")
	}
	x.Source = f["source"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerequest %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceRequest) String() string {
	return fmt.Sprintf("InvoiceRequest(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceRequest) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceRequest) Normalize() InvoiceRequest {
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInvoiceRequests orders by priority descending, then by ID.
func SortInvoiceRequests(xs []InvoiceRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceRequests drops the archived entries, preserving order.
func ActiveInvoiceRequests(xs []InvoiceRequest) []InvoiceRequest {
	out := make([]InvoiceRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceRequests adds up the priority of every active entry.
func SumInvoiceRequests(xs []InvoiceRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInvoiceRequests keys the entries by ID, last write winning.
func IndexInvoiceRequests(xs []InvoiceRequest) map[string]InvoiceRequest {
	out := make(map[string]InvoiceRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceRequest builds a valid InvoiceRequest, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceRequest(id string, at time.Time) InvoiceRequest {
	return InvoiceRequest{
		ID:        id,
		Source:    "source-" + id,
		Group:     "group-" + id,
		Reference: "seeded by invoice.SampleInvoiceRequest",
		Priority:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceSummarySchema is the stored shape of a InvoiceSummary.
var InvoiceSummarySchema = schema.New("invoice.invoicesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceSummary is one invoice summary.
type InvoiceSummary struct {
	ID       string
	Label    string
	Source   string
	Region   string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceSummary) Key() string { return "invoice/invoicesummary/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceSummary decodes a stored record.
func ParseInvoiceSummary(f codec.Fields) (InvoiceSummary, error) {
	var x InvoiceSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicesummary: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicesummary %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceSummary) String() string {
	return fmt.Sprintf("InvoiceSummary(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceSummary) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceSummary) Normalize() InvoiceSummary {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortInvoiceSummarys orders by weight descending, then by ID.
func SortInvoiceSummarys(xs []InvoiceSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceSummarys drops the archived entries, preserving order.
func ActiveInvoiceSummarys(xs []InvoiceSummary) []InvoiceSummary {
	out := make([]InvoiceSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceSummarys adds up the weight of every active entry.
func SumInvoiceSummarys(xs []InvoiceSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexInvoiceSummarys keys the entries by ID, last write winning.
func IndexInvoiceSummarys(xs []InvoiceSummary) map[string]InvoiceSummary {
	out := make(map[string]InvoiceSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceSummary builds a valid InvoiceSummary, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceSummary(id string, at time.Time) InvoiceSummary {
	return InvoiceSummary{
		ID:       id,
		Label:    "label-" + id,
		Source:   "source-" + id,
		Region:   "seeded by invoice.SampleInvoiceSummary",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoicePolicySchema is the stored shape of a InvoicePolicy.
var InvoicePolicySchema = schema.New("invoice.invoicepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoicePolicy is one invoice policy.
type InvoicePolicy struct {
	ID       string
	Comment  string
	Owner    string
	Label    string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoicePolicy) Key() string { return "invoice/invoicepolicy/" + x.ID }

// Fields encodes x for storage.
func (x InvoicePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoicePolicy decodes a stored record.
func ParseInvoicePolicy(f codec.Fields) (InvoicePolicy, error) {
	var x InvoicePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicepolicy: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicepolicy %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoicePolicy) String() string {
	return fmt.Sprintf("InvoicePolicy(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoicePolicy) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoicePolicy) Normalize() InvoicePolicy {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortInvoicePolicys orders by priority descending, then by ID.
func SortInvoicePolicys(xs []InvoicePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoicePolicys drops the archived entries, preserving order.
func ActiveInvoicePolicys(xs []InvoicePolicy) []InvoicePolicy {
	out := make([]InvoicePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoicePolicys adds up the priority of every active entry.
func SumInvoicePolicys(xs []InvoicePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInvoicePolicys keys the entries by ID, last write winning.
func IndexInvoicePolicys(xs []InvoicePolicy) map[string]InvoicePolicy {
	out := make(map[string]InvoicePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoicePolicy builds a valid InvoicePolicy, for tests and for the seed data the cron
// binary writes.
func SampleInvoicePolicy(id string, at time.Time) InvoicePolicy {
	return InvoicePolicy{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by invoice.SampleInvoicePolicy",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceLinkSchema is the stored shape of a InvoiceLink.
var InvoiceLinkSchema = schema.New("invoice.invoicelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceLink is one invoice link.
type InvoiceLink struct {
	ID       string
	Source   string
	Owner    string
	Comment  string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceLink) Key() string { return "invoice/invoicelink/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceLink decodes a stored record.
func ParseInvoiceLink(f codec.Fields) (InvoiceLink, error) {
	var x InvoiceLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicelink: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicelink %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceLink) String() string {
	return fmt.Sprintf("InvoiceLink(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceLink) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceLink) Normalize() InvoiceLink {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInvoiceLinks orders by priority descending, then by ID.
func SortInvoiceLinks(xs []InvoiceLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceLinks drops the archived entries, preserving order.
func ActiveInvoiceLinks(xs []InvoiceLink) []InvoiceLink {
	out := make([]InvoiceLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceLinks adds up the priority of every active entry.
func SumInvoiceLinks(xs []InvoiceLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInvoiceLinks keys the entries by ID, last write winning.
func IndexInvoiceLinks(xs []InvoiceLink) map[string]InvoiceLink {
	out := make(map[string]InvoiceLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceLink builds a valid InvoiceLink, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceLink(id string, at time.Time) InvoiceLink {
	return InvoiceLink{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Comment:  "seeded by invoice.SampleInvoiceLink",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceNoteSchema is the stored shape of a InvoiceNote.
var InvoiceNoteSchema = schema.New("invoice.invoicenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceNote is one invoice note.
type InvoiceNote struct {
	ID       string
	Category string
	Channel  string
	Source   string
	Retained int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceNote) Key() string { return "invoice/invoicenote/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceNote decodes a stored record.
func ParseInvoiceNote(f codec.Fields) (InvoiceNote, error) {
	var x InvoiceNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicenote: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicenote %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceNote) String() string {
	return fmt.Sprintf("InvoiceNote(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceNote) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceNote) Normalize() InvoiceNote {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortInvoiceNotes orders by retained descending, then by ID.
func SortInvoiceNotes(xs []InvoiceNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceNotes drops the archived entries, preserving order.
func ActiveInvoiceNotes(xs []InvoiceNote) []InvoiceNote {
	out := make([]InvoiceNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceNotes adds up the retained of every active entry.
func SumInvoiceNotes(xs []InvoiceNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexInvoiceNotes keys the entries by ID, last write winning.
func IndexInvoiceNotes(xs []InvoiceNote) map[string]InvoiceNote {
	out := make(map[string]InvoiceNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceNote builds a valid InvoiceNote, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceNote(id string, at time.Time) InvoiceNote {
	return InvoiceNote{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by invoice.SampleInvoiceNote",
		Retained: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceWindowSchema is the stored shape of a InvoiceWindow.
var InvoiceWindowSchema = schema.New("invoice.invoicewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceWindow is one invoice window.
type InvoiceWindow struct {
	ID        string
	Source    string
	Reference string
	Label     string
	Priority  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceWindow) Key() string { return "invoice/invoicewindow/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"reference": x.Reference,
		"label":     x.Label,
		"priority":  strconv.Itoa(x.Priority),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceWindow decodes a stored record.
func ParseInvoiceWindow(f codec.Fields) (InvoiceWindow, error) {
	var x InvoiceWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicewindow: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicewindow %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceWindow) String() string {
	return fmt.Sprintf("InvoiceWindow(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceWindow) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceWindow) Normalize() InvoiceWindow {
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortInvoiceWindows orders by priority descending, then by ID.
func SortInvoiceWindows(xs []InvoiceWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceWindows drops the archived entries, preserving order.
func ActiveInvoiceWindows(xs []InvoiceWindow) []InvoiceWindow {
	out := make([]InvoiceWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceWindows adds up the priority of every active entry.
func SumInvoiceWindows(xs []InvoiceWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInvoiceWindows keys the entries by ID, last write winning.
func IndexInvoiceWindows(xs []InvoiceWindow) map[string]InvoiceWindow {
	out := make(map[string]InvoiceWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceWindow builds a valid InvoiceWindow, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceWindow(id string, at time.Time) InvoiceWindow {
	return InvoiceWindow{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Label:     "seeded by invoice.SampleInvoiceWindow",
		Priority:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceQuotaSchema is the stored shape of a InvoiceQuota.
var InvoiceQuotaSchema = schema.New("invoice.invoicequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceQuota is one invoice quota.
type InvoiceQuota struct {
	ID        string
	Category  string
	Group     string
	Reference string
	Weight    int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceQuota) Key() string { return "invoice/invoicequota/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"group":     x.Group,
		"reference": x.Reference,
		"weight":    strconv.Itoa(x.Weight),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceQuota decodes a stored record.
func ParseInvoiceQuota(f codec.Fields) (InvoiceQuota, error) {
	var x InvoiceQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicequota: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicequota %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceQuota) String() string {
	return fmt.Sprintf("InvoiceQuota(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceQuota) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceQuota) Normalize() InvoiceQuota {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortInvoiceQuotas orders by weight descending, then by ID.
func SortInvoiceQuotas(xs []InvoiceQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceQuotas drops the archived entries, preserving order.
func ActiveInvoiceQuotas(xs []InvoiceQuota) []InvoiceQuota {
	out := make([]InvoiceQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceQuotas adds up the weight of every active entry.
func SumInvoiceQuotas(xs []InvoiceQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexInvoiceQuotas keys the entries by ID, last write winning.
func IndexInvoiceQuotas(xs []InvoiceQuota) map[string]InvoiceQuota {
	out := make(map[string]InvoiceQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceQuota builds a valid InvoiceQuota, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceQuota(id string, at time.Time) InvoiceQuota {
	return InvoiceQuota{
		ID:        id,
		Category:  "category-" + id,
		Group:     "group-" + id,
		Reference: "seeded by invoice.SampleInvoiceQuota",
		Weight:    len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceBucketSchema is the stored shape of a InvoiceBucket.
var InvoiceBucketSchema = schema.New("invoice.invoicebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceBucket is one invoice bucket.
type InvoiceBucket struct {
	ID       string
	Region   string
	Name     string
	Comment  string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceBucket) Key() string { return "invoice/invoicebucket/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceBucket decodes a stored record.
func ParseInvoiceBucket(f codec.Fields) (InvoiceBucket, error) {
	var x InvoiceBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicebucket: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicebucket %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceBucket) String() string {
	return fmt.Sprintf("InvoiceBucket(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceBucket) Normalize() InvoiceBucket {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortInvoiceBuckets orders by retained descending, then by ID.
func SortInvoiceBuckets(xs []InvoiceBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceBuckets drops the archived entries, preserving order.
func ActiveInvoiceBuckets(xs []InvoiceBucket) []InvoiceBucket {
	out := make([]InvoiceBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceBuckets adds up the retained of every active entry.
func SumInvoiceBuckets(xs []InvoiceBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexInvoiceBuckets keys the entries by ID, last write winning.
func IndexInvoiceBuckets(xs []InvoiceBucket) map[string]InvoiceBucket {
	out := make(map[string]InvoiceBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceBucket builds a valid InvoiceBucket, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceBucket(id string, at time.Time) InvoiceBucket {
	return InvoiceBucket{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by invoice.SampleInvoiceBucket",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceLaneSchema is the stored shape of a InvoiceLane.
var InvoiceLaneSchema = schema.New("invoice.invoicelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceLane is one invoice lane.
type InvoiceLane struct {
	ID        string
	Name      string
	Owner     string
	Reference string
	Priority  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceLane) Key() string { return "invoice/invoicelane/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"owner":     x.Owner,
		"reference": x.Reference,
		"priority":  strconv.Itoa(x.Priority),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceLane decodes a stored record.
func ParseInvoiceLane(f codec.Fields) (InvoiceLane, error) {
	var x InvoiceLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicelane: missing id")
	}
	x.Name = f["name"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicelane %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceLane) String() string {
	return fmt.Sprintf("InvoiceLane(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceLane) Normalize() InvoiceLane {
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortInvoiceLanes orders by priority descending, then by ID.
func SortInvoiceLanes(xs []InvoiceLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceLanes drops the archived entries, preserving order.
func ActiveInvoiceLanes(xs []InvoiceLane) []InvoiceLane {
	out := make([]InvoiceLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceLanes adds up the priority of every active entry.
func SumInvoiceLanes(xs []InvoiceLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInvoiceLanes keys the entries by ID, last write winning.
func IndexInvoiceLanes(xs []InvoiceLane) map[string]InvoiceLane {
	out := make(map[string]InvoiceLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceLane builds a valid InvoiceLane, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceLane(id string, at time.Time) InvoiceLane {
	return InvoiceLane{
		ID:        id,
		Name:      "name-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by invoice.SampleInvoiceLane",
		Priority:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceSlotSchema is the stored shape of a InvoiceSlot.
var InvoiceSlotSchema = schema.New("invoice.invoiceslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceSlot is one invoice slot.
type InvoiceSlot struct {
	ID        string
	Group     string
	Reference string
	Category  string
	Retained  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceSlot) Key() string { return "invoice/invoiceslot/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"category":  x.Category,
		"retained":  strconv.Itoa(x.Retained),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceSlot decodes a stored record.
func ParseInvoiceSlot(f codec.Fields) (InvoiceSlot, error) {
	var x InvoiceSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoiceslot: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoiceslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoiceslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceSlot) String() string {
	return fmt.Sprintf("InvoiceSlot(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceSlot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceSlot) Normalize() InvoiceSlot {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortInvoiceSlots orders by retained descending, then by ID.
func SortInvoiceSlots(xs []InvoiceSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceSlots drops the archived entries, preserving order.
func ActiveInvoiceSlots(xs []InvoiceSlot) []InvoiceSlot {
	out := make([]InvoiceSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceSlots adds up the retained of every active entry.
func SumInvoiceSlots(xs []InvoiceSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexInvoiceSlots keys the entries by ID, last write winning.
func IndexInvoiceSlots(xs []InvoiceSlot) map[string]InvoiceSlot {
	out := make(map[string]InvoiceSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceSlot builds a valid InvoiceSlot, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceSlot(id string, at time.Time) InvoiceSlot {
	return InvoiceSlot{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by invoice.SampleInvoiceSlot",
		Retained:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceTagSchema is the stored shape of a InvoiceTag.
var InvoiceTagSchema = schema.New("invoice.invoicetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceTag is one invoice tag.
type InvoiceTag struct {
	ID       string
	Channel  string
	Owner    string
	Region   string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceTag) Key() string { return "invoice/invoicetag/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceTag decodes a stored record.
func ParseInvoiceTag(f codec.Fields) (InvoiceTag, error) {
	var x InvoiceTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicetag: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicetag %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceTag) String() string {
	return fmt.Sprintf("InvoiceTag(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceTag) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceTag) Normalize() InvoiceTag {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortInvoiceTags orders by capacity descending, then by ID.
func SortInvoiceTags(xs []InvoiceTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceTags drops the archived entries, preserving order.
func ActiveInvoiceTags(xs []InvoiceTag) []InvoiceTag {
	out := make([]InvoiceTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceTags adds up the capacity of every active entry.
func SumInvoiceTags(xs []InvoiceTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexInvoiceTags keys the entries by ID, last write winning.
func IndexInvoiceTags(xs []InvoiceTag) map[string]InvoiceTag {
	out := make(map[string]InvoiceTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceTag builds a valid InvoiceTag, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceTag(id string, at time.Time) InvoiceTag {
	return InvoiceTag{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by invoice.SampleInvoiceTag",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// InvoiceRuleSchema is the stored shape of a InvoiceRule.
var InvoiceRuleSchema = schema.New("invoice.invoicerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceRule is one invoice rule.
type InvoiceRule struct {
	ID        string
	Category  string
	Group     string
	Reference string
	Offset    int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceRule) Key() string { return "invoice/invoicerule/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"group":     x.Group,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceRule decodes a stored record.
func ParseInvoiceRule(f codec.Fields) (InvoiceRule, error) {
	var x InvoiceRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicerule: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerule %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceRule) String() string {
	return fmt.Sprintf("InvoiceRule(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceRule) Normalize() InvoiceRule {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInvoiceRules orders by offset descending, then by ID.
func SortInvoiceRules(xs []InvoiceRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceRules drops the archived entries, preserving order.
func ActiveInvoiceRules(xs []InvoiceRule) []InvoiceRule {
	out := make([]InvoiceRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceRules adds up the offset of every active entry.
func SumInvoiceRules(xs []InvoiceRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexInvoiceRules keys the entries by ID, last write winning.
func IndexInvoiceRules(xs []InvoiceRule) map[string]InvoiceRule {
	out := make(map[string]InvoiceRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceRule builds a valid InvoiceRule, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceRule(id string, at time.Time) InvoiceRule {
	return InvoiceRule{
		ID:        id,
		Category:  "category-" + id,
		Group:     "group-" + id,
		Reference: "seeded by invoice.SampleInvoiceRule",
		Offset:    len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceHoldSchema is the stored shape of a InvoiceHold.
var InvoiceHoldSchema = schema.New("invoice.invoicehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceHold is one invoice hold.
type InvoiceHold struct {
	ID       string
	Source   string
	Owner    string
	Group    string
	Weight   int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceHold) Key() string { return "invoice/invoicehold/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"group":    x.Group,
		"weight":   strconv.Itoa(x.Weight),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceHold decodes a stored record.
func ParseInvoiceHold(f codec.Fields) (InvoiceHold, error) {
	var x InvoiceHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicehold: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicehold %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceHold) String() string {
	return fmt.Sprintf("InvoiceHold(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceHold) Normalize() InvoiceHold {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortInvoiceHolds orders by weight descending, then by ID.
func SortInvoiceHolds(xs []InvoiceHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceHolds drops the archived entries, preserving order.
func ActiveInvoiceHolds(xs []InvoiceHold) []InvoiceHold {
	out := make([]InvoiceHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceHolds adds up the weight of every active entry.
func SumInvoiceHolds(xs []InvoiceHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexInvoiceHolds keys the entries by ID, last write winning.
func IndexInvoiceHolds(xs []InvoiceHold) map[string]InvoiceHold {
	out := make(map[string]InvoiceHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceHold builds a valid InvoiceHold, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceHold(id string, at time.Time) InvoiceHold {
	return InvoiceHold{
		ID:      id,
		Source:  "source-" + id,
		Owner:   "owner-" + id,
		Group:   "seeded by invoice.SampleInvoiceHold",
		Weight:  len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// InvoiceClaimSchema is the stored shape of a InvoiceClaim.
var InvoiceClaimSchema = schema.New("invoice.invoiceclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceClaim is one invoice claim.
type InvoiceClaim struct {
	ID        string
	Region    string
	Category  string
	Reference string
	Attempts  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x InvoiceClaim) Key() string { return "invoice/invoiceclaim/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"category":  x.Category,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceClaim decodes a stored record.
func ParseInvoiceClaim(f codec.Fields) (InvoiceClaim, error) {
	var x InvoiceClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoiceclaim: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceclaim %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("invoice: invoiceclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoiceclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoiceclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceClaim) String() string {
	return fmt.Sprintf("InvoiceClaim(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceClaim) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceClaim) Normalize() InvoiceClaim {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInvoiceClaims orders by attempts descending, then by ID.
func SortInvoiceClaims(xs []InvoiceClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceClaims drops the archived entries, preserving order.
func ActiveInvoiceClaims(xs []InvoiceClaim) []InvoiceClaim {
	out := make([]InvoiceClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceClaims adds up the attempts of every active entry.
func SumInvoiceClaims(xs []InvoiceClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexInvoiceClaims keys the entries by ID, last write winning.
func IndexInvoiceClaims(xs []InvoiceClaim) map[string]InvoiceClaim {
	out := make(map[string]InvoiceClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceClaim builds a valid InvoiceClaim, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceClaim(id string, at time.Time) InvoiceClaim {
	return InvoiceClaim{
		ID:        id,
		Region:    "region-" + id,
		Category:  "category-" + id,
		Reference: "seeded by invoice.SampleInvoiceClaim",
		Attempts:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// InvoiceCursorSchema is the stored shape of a InvoiceCursor.
var InvoiceCursorSchema = schema.New("invoice.invoicecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// InvoiceCursor is one invoice cursor.
type InvoiceCursor struct {
	ID       string
	Owner    string
	Source   string
	Category string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x InvoiceCursor) Key() string { return "invoice/invoicecursor/" + x.ID }

// Fields encodes x for storage.
func (x InvoiceCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"category": x.Category,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseInvoiceCursor decodes a stored record.
func ParseInvoiceCursor(f codec.Fields) (InvoiceCursor, error) {
	var x InvoiceCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("invoice: invoicecursor: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicecursor %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("invoice: invoicecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("invoice: invoicecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("invoice: invoicecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x InvoiceCursor) String() string {
	return fmt.Sprintf("InvoiceCursor(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x InvoiceCursor) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x InvoiceCursor) Normalize() InvoiceCursor {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortInvoiceCursors orders by priority descending, then by ID.
func SortInvoiceCursors(xs []InvoiceCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveInvoiceCursors drops the archived entries, preserving order.
func ActiveInvoiceCursors(xs []InvoiceCursor) []InvoiceCursor {
	out := make([]InvoiceCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumInvoiceCursors adds up the priority of every active entry.
func SumInvoiceCursors(xs []InvoiceCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexInvoiceCursors keys the entries by ID, last write winning.
func IndexInvoiceCursors(xs []InvoiceCursor) map[string]InvoiceCursor {
	out := make(map[string]InvoiceCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleInvoiceCursor builds a valid InvoiceCursor, for tests and for the seed data the cron
// binary writes.
func SampleInvoiceCursor(id string, at time.Time) InvoiceCursor {
	return InvoiceCursor{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Category: "seeded by invoice.SampleInvoiceCursor",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	InvoiceRecordSchema,
	InvoiceEntrySchema,
	InvoiceItemSchema,
	InvoiceDraftSchema,
	InvoiceRevisionSchema,
	InvoiceSnapshotSchema,
	InvoiceBatchSchema,
	InvoiceRequestSchema,
	InvoiceSummarySchema,
	InvoicePolicySchema,
	InvoiceLinkSchema,
	InvoiceNoteSchema,
	InvoiceWindowSchema,
	InvoiceQuotaSchema,
	InvoiceBucketSchema,
	InvoiceLaneSchema,
	InvoiceSlotSchema,
	InvoiceTagSchema,
	InvoiceRuleSchema,
	InvoiceHoldSchema,
	InvoiceClaimSchema,
	InvoiceCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"invoice/invoicerecord/",
	"invoice/invoiceentry/",
	"invoice/invoiceitem/",
	"invoice/invoicedraft/",
	"invoice/invoicerevision/",
	"invoice/invoicesnapshot/",
	"invoice/invoicebatch/",
	"invoice/invoicerequest/",
	"invoice/invoicesummary/",
	"invoice/invoicepolicy/",
	"invoice/invoicelink/",
	"invoice/invoicenote/",
	"invoice/invoicewindow/",
	"invoice/invoicequota/",
	"invoice/invoicebucket/",
	"invoice/invoicelane/",
	"invoice/invoiceslot/",
	"invoice/invoicetag/",
	"invoice/invoicerule/",
	"invoice/invoicehold/",
	"invoice/invoiceclaim/",
	"invoice/invoicecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "invoice: " + strings.Join(names, ", ")
}
