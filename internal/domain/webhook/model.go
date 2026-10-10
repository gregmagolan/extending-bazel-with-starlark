// Package webhook holds the webhook domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package webhook

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// WebhookRecordSchema is the stored shape of a WebhookRecord.
var WebhookRecordSchema = schema.New("webhook.webhookrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookRecord is one webhook record.
type WebhookRecord struct {
	ID       string
	Category string
	Owner    string
	Name     string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookRecord) Key() string { return "webhook/webhookrecord/" + x.ID }

// Fields encodes x for storage.
func (x WebhookRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookRecord decodes a stored record.
func ParseWebhookRecord(f codec.Fields) (WebhookRecord, error) {
	var x WebhookRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookrecord: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrecord %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookRecord) String() string {
	return fmt.Sprintf("WebhookRecord(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookRecord) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookRecord) Normalize() WebhookRecord {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortWebhookRecords orders by count descending, then by ID.
func SortWebhookRecords(xs []WebhookRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookRecords drops the archived entries, preserving order.
func ActiveWebhookRecords(xs []WebhookRecord) []WebhookRecord {
	out := make([]WebhookRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookRecords adds up the count of every active entry.
func SumWebhookRecords(xs []WebhookRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexWebhookRecords keys the entries by ID, last write winning.
func IndexWebhookRecords(xs []WebhookRecord) map[string]WebhookRecord {
	out := make(map[string]WebhookRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookRecord builds a valid WebhookRecord, for tests and for the seed data the cron
// binary writes.
func SampleWebhookRecord(id string, at time.Time) WebhookRecord {
	return WebhookRecord{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by webhook.SampleWebhookRecord",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookEntrySchema is the stored shape of a WebhookEntry.
var WebhookEntrySchema = schema.New("webhook.webhookentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookEntry is one webhook entry.
type WebhookEntry struct {
	ID        string
	Owner     string
	Region    string
	Reference string
	Offset    int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WebhookEntry) Key() string { return "webhook/webhookentry/" + x.ID }

// Fields encodes x for storage.
func (x WebhookEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"region":    x.Region,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookEntry decodes a stored record.
func ParseWebhookEntry(f codec.Fields) (WebhookEntry, error) {
	var x WebhookEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookentry: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookentry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookEntry) String() string {
	return fmt.Sprintf("WebhookEntry(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookEntry) Normalize() WebhookEntry {
	x.Owner = strings.TrimSpace(x.Owner)
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

// SortWebhookEntrys orders by offset descending, then by ID.
func SortWebhookEntrys(xs []WebhookEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookEntrys drops the archived entries, preserving order.
func ActiveWebhookEntrys(xs []WebhookEntry) []WebhookEntry {
	out := make([]WebhookEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookEntrys adds up the offset of every active entry.
func SumWebhookEntrys(xs []WebhookEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexWebhookEntrys keys the entries by ID, last write winning.
func IndexWebhookEntrys(xs []WebhookEntry) map[string]WebhookEntry {
	out := make(map[string]WebhookEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookEntry builds a valid WebhookEntry, for tests and for the seed data the cron
// binary writes.
func SampleWebhookEntry(id string, at time.Time) WebhookEntry {
	return WebhookEntry{
		ID:        id,
		Owner:     "owner-" + id,
		Region:    "region-" + id,
		Reference: "seeded by webhook.SampleWebhookEntry",
		Offset:    len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WebhookItemSchema is the stored shape of a WebhookItem.
var WebhookItemSchema = schema.New("webhook.webhookitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookItem is one webhook item.
type WebhookItem struct {
	ID       string
	Source   string
	Group    string
	Channel  string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookItem) Key() string { return "webhook/webhookitem/" + x.ID }

// Fields encodes x for storage.
func (x WebhookItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"group":    x.Group,
		"channel":  x.Channel,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookItem decodes a stored record.
func ParseWebhookItem(f codec.Fields) (WebhookItem, error) {
	var x WebhookItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookitem: missing id")
	}
	x.Source = f["source"]
	x.Group = f["group"]
	x.Channel = f["channel"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookitem %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookItem) String() string {
	return fmt.Sprintf("WebhookItem(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookItem) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookItem) Normalize() WebhookItem {
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWebhookItems orders by weight descending, then by ID.
func SortWebhookItems(xs []WebhookItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookItems drops the archived entries, preserving order.
func ActiveWebhookItems(xs []WebhookItem) []WebhookItem {
	out := make([]WebhookItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookItems adds up the weight of every active entry.
func SumWebhookItems(xs []WebhookItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexWebhookItems keys the entries by ID, last write winning.
func IndexWebhookItems(xs []WebhookItem) map[string]WebhookItem {
	out := make(map[string]WebhookItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookItem builds a valid WebhookItem, for tests and for the seed data the cron
// binary writes.
func SampleWebhookItem(id string, at time.Time) WebhookItem {
	return WebhookItem{
		ID:       id,
		Source:   "source-" + id,
		Group:    "group-" + id,
		Channel:  "seeded by webhook.SampleWebhookItem",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookDraftSchema is the stored shape of a WebhookDraft.
var WebhookDraftSchema = schema.New("webhook.webhookdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookDraft is one webhook draft.
type WebhookDraft struct {
	ID       string
	Source   string
	Region   string
	Group    string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookDraft) Key() string { return "webhook/webhookdraft/" + x.ID }

// Fields encodes x for storage.
func (x WebhookDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"region":   x.Region,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookDraft decodes a stored record.
func ParseWebhookDraft(f codec.Fields) (WebhookDraft, error) {
	var x WebhookDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookdraft: missing id")
	}
	x.Source = f["source"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookdraft %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookDraft) String() string {
	return fmt.Sprintf("WebhookDraft(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookDraft) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookDraft) Normalize() WebhookDraft {
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortWebhookDrafts orders by capacity descending, then by ID.
func SortWebhookDrafts(xs []WebhookDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookDrafts drops the archived entries, preserving order.
func ActiveWebhookDrafts(xs []WebhookDraft) []WebhookDraft {
	out := make([]WebhookDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookDrafts adds up the capacity of every active entry.
func SumWebhookDrafts(xs []WebhookDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexWebhookDrafts keys the entries by ID, last write winning.
func IndexWebhookDrafts(xs []WebhookDraft) map[string]WebhookDraft {
	out := make(map[string]WebhookDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookDraft builds a valid WebhookDraft, for tests and for the seed data the cron
// binary writes.
func SampleWebhookDraft(id string, at time.Time) WebhookDraft {
	return WebhookDraft{
		ID:       id,
		Source:   "source-" + id,
		Region:   "region-" + id,
		Group:    "seeded by webhook.SampleWebhookDraft",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookRevisionSchema is the stored shape of a WebhookRevision.
var WebhookRevisionSchema = schema.New("webhook.webhookrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookRevision is one webhook revision.
type WebhookRevision struct {
	ID       string
	Name     string
	Region   string
	Group    string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookRevision) Key() string { return "webhook/webhookrevision/" + x.ID }

// Fields encodes x for storage.
func (x WebhookRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"region":   x.Region,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookRevision decodes a stored record.
func ParseWebhookRevision(f codec.Fields) (WebhookRevision, error) {
	var x WebhookRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookrevision: missing id")
	}
	x.Name = f["name"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookRevision) String() string {
	return fmt.Sprintf("WebhookRevision(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookRevision) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookRevision) Normalize() WebhookRevision {
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortWebhookRevisions orders by offset descending, then by ID.
func SortWebhookRevisions(xs []WebhookRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookRevisions drops the archived entries, preserving order.
func ActiveWebhookRevisions(xs []WebhookRevision) []WebhookRevision {
	out := make([]WebhookRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookRevisions adds up the offset of every active entry.
func SumWebhookRevisions(xs []WebhookRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexWebhookRevisions keys the entries by ID, last write winning.
func IndexWebhookRevisions(xs []WebhookRevision) map[string]WebhookRevision {
	out := make(map[string]WebhookRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookRevision builds a valid WebhookRevision, for tests and for the seed data the cron
// binary writes.
func SampleWebhookRevision(id string, at time.Time) WebhookRevision {
	return WebhookRevision{
		ID:      id,
		Name:    "name-" + id,
		Region:  "region-" + id,
		Group:   "seeded by webhook.SampleWebhookRevision",
		Offset:  len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// WebhookSnapshotSchema is the stored shape of a WebhookSnapshot.
var WebhookSnapshotSchema = schema.New("webhook.webhooksnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookSnapshot is one webhook snapshot.
type WebhookSnapshot struct {
	ID       string
	Name     string
	Comment  string
	Label    string
	Weight   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookSnapshot) Key() string { return "webhook/webhooksnapshot/" + x.ID }

// Fields encodes x for storage.
func (x WebhookSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"comment":  x.Comment,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookSnapshot decodes a stored record.
func ParseWebhookSnapshot(f codec.Fields) (WebhookSnapshot, error) {
	var x WebhookSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhooksnapshot: missing id")
	}
	x.Name = f["name"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooksnapshot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooksnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhooksnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhooksnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookSnapshot) String() string {
	return fmt.Sprintf("WebhookSnapshot(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookSnapshot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookSnapshot) Normalize() WebhookSnapshot {
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWebhookSnapshots orders by weight descending, then by ID.
func SortWebhookSnapshots(xs []WebhookSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookSnapshots drops the archived entries, preserving order.
func ActiveWebhookSnapshots(xs []WebhookSnapshot) []WebhookSnapshot {
	out := make([]WebhookSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookSnapshots adds up the weight of every active entry.
func SumWebhookSnapshots(xs []WebhookSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexWebhookSnapshots keys the entries by ID, last write winning.
func IndexWebhookSnapshots(xs []WebhookSnapshot) map[string]WebhookSnapshot {
	out := make(map[string]WebhookSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookSnapshot builds a valid WebhookSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleWebhookSnapshot(id string, at time.Time) WebhookSnapshot {
	return WebhookSnapshot{
		ID:       id,
		Name:     "name-" + id,
		Comment:  "comment-" + id,
		Label:    "seeded by webhook.SampleWebhookSnapshot",
		Weight:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookBatchSchema is the stored shape of a WebhookBatch.
var WebhookBatchSchema = schema.New("webhook.webhookbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookBatch is one webhook batch.
type WebhookBatch struct {
	ID        string
	Reference string
	Label     string
	Comment   string
	Version   int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WebhookBatch) Key() string { return "webhook/webhookbatch/" + x.ID }

// Fields encodes x for storage.
func (x WebhookBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"comment":   x.Comment,
		"version":   strconv.Itoa(x.Version),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookBatch decodes a stored record.
func ParseWebhookBatch(f codec.Fields) (WebhookBatch, error) {
	var x WebhookBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookbatch: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookbatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookBatch) String() string {
	return fmt.Sprintf("WebhookBatch(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookBatch) Normalize() WebhookBatch {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortWebhookBatchs orders by version descending, then by ID.
func SortWebhookBatchs(xs []WebhookBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookBatchs drops the archived entries, preserving order.
func ActiveWebhookBatchs(xs []WebhookBatch) []WebhookBatch {
	out := make([]WebhookBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookBatchs adds up the version of every active entry.
func SumWebhookBatchs(xs []WebhookBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexWebhookBatchs keys the entries by ID, last write winning.
func IndexWebhookBatchs(xs []WebhookBatch) map[string]WebhookBatch {
	out := make(map[string]WebhookBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookBatch builds a valid WebhookBatch, for tests and for the seed data the cron
// binary writes.
func SampleWebhookBatch(id string, at time.Time) WebhookBatch {
	return WebhookBatch{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Comment:   "seeded by webhook.SampleWebhookBatch",
		Version:   len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WebhookRequestSchema is the stored shape of a WebhookRequest.
var WebhookRequestSchema = schema.New("webhook.webhookrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookRequest is one webhook request.
type WebhookRequest struct {
	ID       string
	Group    string
	Source   string
	Channel  string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookRequest) Key() string { return "webhook/webhookrequest/" + x.ID }

// Fields encodes x for storage.
func (x WebhookRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"channel":  x.Channel,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookRequest decodes a stored record.
func ParseWebhookRequest(f codec.Fields) (WebhookRequest, error) {
	var x WebhookRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookrequest: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookRequest) String() string {
	return fmt.Sprintf("WebhookRequest(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookRequest) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookRequest) Normalize() WebhookRequest {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortWebhookRequests orders by offset descending, then by ID.
func SortWebhookRequests(xs []WebhookRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookRequests drops the archived entries, preserving order.
func ActiveWebhookRequests(xs []WebhookRequest) []WebhookRequest {
	out := make([]WebhookRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookRequests adds up the offset of every active entry.
func SumWebhookRequests(xs []WebhookRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexWebhookRequests keys the entries by ID, last write winning.
func IndexWebhookRequests(xs []WebhookRequest) map[string]WebhookRequest {
	out := make(map[string]WebhookRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookRequest builds a valid WebhookRequest, for tests and for the seed data the cron
// binary writes.
func SampleWebhookRequest(id string, at time.Time) WebhookRequest {
	return WebhookRequest{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by webhook.SampleWebhookRequest",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookSummarySchema is the stored shape of a WebhookSummary.
var WebhookSummarySchema = schema.New("webhook.webhooksummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookSummary is one webhook summary.
type WebhookSummary struct {
	ID       string
	Category string
	Region   string
	Label    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookSummary) Key() string { return "webhook/webhooksummary/" + x.ID }

// Fields encodes x for storage.
func (x WebhookSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookSummary decodes a stored record.
func ParseWebhookSummary(f codec.Fields) (WebhookSummary, error) {
	var x WebhookSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhooksummary: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooksummary %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooksummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhooksummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhooksummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookSummary) String() string {
	return fmt.Sprintf("WebhookSummary(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookSummary) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookSummary) Normalize() WebhookSummary {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortWebhookSummarys orders by count descending, then by ID.
func SortWebhookSummarys(xs []WebhookSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookSummarys drops the archived entries, preserving order.
func ActiveWebhookSummarys(xs []WebhookSummary) []WebhookSummary {
	out := make([]WebhookSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookSummarys adds up the count of every active entry.
func SumWebhookSummarys(xs []WebhookSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexWebhookSummarys keys the entries by ID, last write winning.
func IndexWebhookSummarys(xs []WebhookSummary) map[string]WebhookSummary {
	out := make(map[string]WebhookSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookSummary builds a valid WebhookSummary, for tests and for the seed data the cron
// binary writes.
func SampleWebhookSummary(id string, at time.Time) WebhookSummary {
	return WebhookSummary{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Label:    "seeded by webhook.SampleWebhookSummary",
		Count:    len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookPolicySchema is the stored shape of a WebhookPolicy.
var WebhookPolicySchema = schema.New("webhook.webhookpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookPolicy is one webhook policy.
type WebhookPolicy struct {
	ID       string
	Group    string
	Owner    string
	Name     string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookPolicy) Key() string { return "webhook/webhookpolicy/" + x.ID }

// Fields encodes x for storage.
func (x WebhookPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookPolicy decodes a stored record.
func ParseWebhookPolicy(f codec.Fields) (WebhookPolicy, error) {
	var x WebhookPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookpolicy: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookpolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookPolicy) String() string {
	return fmt.Sprintf("WebhookPolicy(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookPolicy) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookPolicy) Normalize() WebhookPolicy {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortWebhookPolicys orders by retained descending, then by ID.
func SortWebhookPolicys(xs []WebhookPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookPolicys drops the archived entries, preserving order.
func ActiveWebhookPolicys(xs []WebhookPolicy) []WebhookPolicy {
	out := make([]WebhookPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookPolicys adds up the retained of every active entry.
func SumWebhookPolicys(xs []WebhookPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexWebhookPolicys keys the entries by ID, last write winning.
func IndexWebhookPolicys(xs []WebhookPolicy) map[string]WebhookPolicy {
	out := make(map[string]WebhookPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookPolicy builds a valid WebhookPolicy, for tests and for the seed data the cron
// binary writes.
func SampleWebhookPolicy(id string, at time.Time) WebhookPolicy {
	return WebhookPolicy{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by webhook.SampleWebhookPolicy",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookLinkSchema is the stored shape of a WebhookLink.
var WebhookLinkSchema = schema.New("webhook.webhooklink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookLink is one webhook link.
type WebhookLink struct {
	ID       string
	Group    string
	Region   string
	Name     string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookLink) Key() string { return "webhook/webhooklink/" + x.ID }

// Fields encodes x for storage.
func (x WebhookLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"region":   x.Region,
		"name":     x.Name,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookLink decodes a stored record.
func ParseWebhookLink(f codec.Fields) (WebhookLink, error) {
	var x WebhookLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhooklink: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooklink %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooklink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhooklink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhooklink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookLink) String() string {
	return fmt.Sprintf("WebhookLink(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookLink) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookLink) Normalize() WebhookLink {
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortWebhookLinks orders by capacity descending, then by ID.
func SortWebhookLinks(xs []WebhookLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookLinks drops the archived entries, preserving order.
func ActiveWebhookLinks(xs []WebhookLink) []WebhookLink {
	out := make([]WebhookLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookLinks adds up the capacity of every active entry.
func SumWebhookLinks(xs []WebhookLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexWebhookLinks keys the entries by ID, last write winning.
func IndexWebhookLinks(xs []WebhookLink) map[string]WebhookLink {
	out := make(map[string]WebhookLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookLink builds a valid WebhookLink, for tests and for the seed data the cron
// binary writes.
func SampleWebhookLink(id string, at time.Time) WebhookLink {
	return WebhookLink{
		ID:       id,
		Group:    "group-" + id,
		Region:   "region-" + id,
		Name:     "seeded by webhook.SampleWebhookLink",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookNoteSchema is the stored shape of a WebhookNote.
var WebhookNoteSchema = schema.New("webhook.webhooknote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookNote is one webhook note.
type WebhookNote struct {
	ID        string
	Owner     string
	Category  string
	Reference string
	Offset    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WebhookNote) Key() string { return "webhook/webhooknote/" + x.ID }

// Fields encodes x for storage.
func (x WebhookNote) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"category":  x.Category,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookNote decodes a stored record.
func ParseWebhookNote(f codec.Fields) (WebhookNote, error) {
	var x WebhookNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhooknote: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooknote %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooknote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhooknote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhooknote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookNote) String() string {
	return fmt.Sprintf("WebhookNote(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookNote) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookNote) Normalize() WebhookNote {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortWebhookNotes orders by offset descending, then by ID.
func SortWebhookNotes(xs []WebhookNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookNotes drops the archived entries, preserving order.
func ActiveWebhookNotes(xs []WebhookNote) []WebhookNote {
	out := make([]WebhookNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookNotes adds up the offset of every active entry.
func SumWebhookNotes(xs []WebhookNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexWebhookNotes keys the entries by ID, last write winning.
func IndexWebhookNotes(xs []WebhookNote) map[string]WebhookNote {
	out := make(map[string]WebhookNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookNote builds a valid WebhookNote, for tests and for the seed data the cron
// binary writes.
func SampleWebhookNote(id string, at time.Time) WebhookNote {
	return WebhookNote{
		ID:        id,
		Owner:     "owner-" + id,
		Category:  "category-" + id,
		Reference: "seeded by webhook.SampleWebhookNote",
		Offset:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WebhookWindowSchema is the stored shape of a WebhookWindow.
var WebhookWindowSchema = schema.New("webhook.webhookwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookWindow is one webhook window.
type WebhookWindow struct {
	ID       string
	Group    string
	Name     string
	Source   string
	Count    int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookWindow) Key() string { return "webhook/webhookwindow/" + x.ID }

// Fields encodes x for storage.
func (x WebhookWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"name":     x.Name,
		"source":   x.Source,
		"count":    strconv.Itoa(x.Count),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookWindow decodes a stored record.
func ParseWebhookWindow(f codec.Fields) (WebhookWindow, error) {
	var x WebhookWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookwindow: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookwindow %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookWindow) String() string {
	return fmt.Sprintf("WebhookWindow(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookWindow) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookWindow) Normalize() WebhookWindow {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortWebhookWindows orders by count descending, then by ID.
func SortWebhookWindows(xs []WebhookWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookWindows drops the archived entries, preserving order.
func ActiveWebhookWindows(xs []WebhookWindow) []WebhookWindow {
	out := make([]WebhookWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookWindows adds up the count of every active entry.
func SumWebhookWindows(xs []WebhookWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexWebhookWindows keys the entries by ID, last write winning.
func IndexWebhookWindows(xs []WebhookWindow) map[string]WebhookWindow {
	out := make(map[string]WebhookWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookWindow builds a valid WebhookWindow, for tests and for the seed data the cron
// binary writes.
func SampleWebhookWindow(id string, at time.Time) WebhookWindow {
	return WebhookWindow{
		ID:      id,
		Group:   "group-" + id,
		Name:    "name-" + id,
		Source:  "seeded by webhook.SampleWebhookWindow",
		Count:   len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// WebhookQuotaSchema is the stored shape of a WebhookQuota.
var WebhookQuotaSchema = schema.New("webhook.webhookquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookQuota is one webhook quota.
type WebhookQuota struct {
	ID       string
	Region   string
	Label    string
	Channel  string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookQuota) Key() string { return "webhook/webhookquota/" + x.ID }

// Fields encodes x for storage.
func (x WebhookQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookQuota decodes a stored record.
func ParseWebhookQuota(f codec.Fields) (WebhookQuota, error) {
	var x WebhookQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookquota: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookquota %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookQuota) String() string {
	return fmt.Sprintf("WebhookQuota(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookQuota) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookQuota) Normalize() WebhookQuota {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWebhookQuotas orders by capacity descending, then by ID.
func SortWebhookQuotas(xs []WebhookQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookQuotas drops the archived entries, preserving order.
func ActiveWebhookQuotas(xs []WebhookQuota) []WebhookQuota {
	out := make([]WebhookQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookQuotas adds up the capacity of every active entry.
func SumWebhookQuotas(xs []WebhookQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexWebhookQuotas keys the entries by ID, last write winning.
func IndexWebhookQuotas(xs []WebhookQuota) map[string]WebhookQuota {
	out := make(map[string]WebhookQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookQuota builds a valid WebhookQuota, for tests and for the seed data the cron
// binary writes.
func SampleWebhookQuota(id string, at time.Time) WebhookQuota {
	return WebhookQuota{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by webhook.SampleWebhookQuota",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookBucketSchema is the stored shape of a WebhookBucket.
var WebhookBucketSchema = schema.New("webhook.webhookbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookBucket is one webhook bucket.
type WebhookBucket struct {
	ID        string
	Name      string
	Reference string
	Region    string
	Priority  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WebhookBucket) Key() string { return "webhook/webhookbucket/" + x.ID }

// Fields encodes x for storage.
func (x WebhookBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"region":    x.Region,
		"priority":  strconv.Itoa(x.Priority),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookBucket decodes a stored record.
func ParseWebhookBucket(f codec.Fields) (WebhookBucket, error) {
	var x WebhookBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookbucket: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookbucket %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookBucket) String() string {
	return fmt.Sprintf("WebhookBucket(%s, name=%q, priority=%d, archived=%t)", x.ID, x.Name, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookBucket) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookBucket) Normalize() WebhookBucket {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortWebhookBuckets orders by priority descending, then by ID.
func SortWebhookBuckets(xs []WebhookBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookBuckets drops the archived entries, preserving order.
func ActiveWebhookBuckets(xs []WebhookBucket) []WebhookBucket {
	out := make([]WebhookBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookBuckets adds up the priority of every active entry.
func SumWebhookBuckets(xs []WebhookBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWebhookBuckets keys the entries by ID, last write winning.
func IndexWebhookBuckets(xs []WebhookBucket) map[string]WebhookBucket {
	out := make(map[string]WebhookBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookBucket builds a valid WebhookBucket, for tests and for the seed data the cron
// binary writes.
func SampleWebhookBucket(id string, at time.Time) WebhookBucket {
	return WebhookBucket{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by webhook.SampleWebhookBucket",
		Priority:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WebhookLaneSchema is the stored shape of a WebhookLane.
var WebhookLaneSchema = schema.New("webhook.webhooklane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookLane is one webhook lane.
type WebhookLane struct {
	ID       string
	Group    string
	Channel  string
	Name     string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookLane) Key() string { return "webhook/webhooklane/" + x.ID }

// Fields encodes x for storage.
func (x WebhookLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"name":     x.Name,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookLane decodes a stored record.
func ParseWebhookLane(f codec.Fields) (WebhookLane, error) {
	var x WebhookLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhooklane: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooklane %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooklane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhooklane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhooklane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookLane) String() string {
	return fmt.Sprintf("WebhookLane(%s, group=%q, weight=%d, archived=%t)", x.ID, x.Group, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookLane) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookLane) Normalize() WebhookLane {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortWebhookLanes orders by weight descending, then by ID.
func SortWebhookLanes(xs []WebhookLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookLanes drops the archived entries, preserving order.
func ActiveWebhookLanes(xs []WebhookLane) []WebhookLane {
	out := make([]WebhookLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookLanes adds up the weight of every active entry.
func SumWebhookLanes(xs []WebhookLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexWebhookLanes keys the entries by ID, last write winning.
func IndexWebhookLanes(xs []WebhookLane) map[string]WebhookLane {
	out := make(map[string]WebhookLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookLane builds a valid WebhookLane, for tests and for the seed data the cron
// binary writes.
func SampleWebhookLane(id string, at time.Time) WebhookLane {
	return WebhookLane{
		ID:      id,
		Group:   "group-" + id,
		Channel: "channel-" + id,
		Name:    "seeded by webhook.SampleWebhookLane",
		Weight:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// WebhookSlotSchema is the stored shape of a WebhookSlot.
var WebhookSlotSchema = schema.New("webhook.webhookslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookSlot is one webhook slot.
type WebhookSlot struct {
	ID       string
	Comment  string
	Category string
	Group    string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookSlot) Key() string { return "webhook/webhookslot/" + x.ID }

// Fields encodes x for storage.
func (x WebhookSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"group":    x.Group,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookSlot decodes a stored record.
func ParseWebhookSlot(f codec.Fields) (WebhookSlot, error) {
	var x WebhookSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookslot: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookslot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookSlot) String() string {
	return fmt.Sprintf("WebhookSlot(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookSlot) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookSlot) Normalize() WebhookSlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortWebhookSlots orders by version descending, then by ID.
func SortWebhookSlots(xs []WebhookSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookSlots drops the archived entries, preserving order.
func ActiveWebhookSlots(xs []WebhookSlot) []WebhookSlot {
	out := make([]WebhookSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookSlots adds up the version of every active entry.
func SumWebhookSlots(xs []WebhookSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexWebhookSlots keys the entries by ID, last write winning.
func IndexWebhookSlots(xs []WebhookSlot) map[string]WebhookSlot {
	out := make(map[string]WebhookSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookSlot builds a valid WebhookSlot, for tests and for the seed data the cron
// binary writes.
func SampleWebhookSlot(id string, at time.Time) WebhookSlot {
	return WebhookSlot{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Group:    "seeded by webhook.SampleWebhookSlot",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookTagSchema is the stored shape of a WebhookTag.
var WebhookTagSchema = schema.New("webhook.webhooktag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookTag is one webhook tag.
type WebhookTag struct {
	ID       string
	Category string
	Owner    string
	Channel  string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookTag) Key() string { return "webhook/webhooktag/" + x.ID }

// Fields encodes x for storage.
func (x WebhookTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookTag decodes a stored record.
func ParseWebhookTag(f codec.Fields) (WebhookTag, error) {
	var x WebhookTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhooktag: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooktag %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("webhook: webhooktag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhooktag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhooktag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookTag) String() string {
	return fmt.Sprintf("WebhookTag(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookTag) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookTag) Normalize() WebhookTag {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortWebhookTags orders by retained descending, then by ID.
func SortWebhookTags(xs []WebhookTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookTags drops the archived entries, preserving order.
func ActiveWebhookTags(xs []WebhookTag) []WebhookTag {
	out := make([]WebhookTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookTags adds up the retained of every active entry.
func SumWebhookTags(xs []WebhookTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexWebhookTags keys the entries by ID, last write winning.
func IndexWebhookTags(xs []WebhookTag) map[string]WebhookTag {
	out := make(map[string]WebhookTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookTag builds a valid WebhookTag, for tests and for the seed data the cron
// binary writes.
func SampleWebhookTag(id string, at time.Time) WebhookTag {
	return WebhookTag{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by webhook.SampleWebhookTag",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookRuleSchema is the stored shape of a WebhookRule.
var WebhookRuleSchema = schema.New("webhook.webhookrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookRule is one webhook rule.
type WebhookRule struct {
	ID       string
	Group    string
	Category string
	Source   string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookRule) Key() string { return "webhook/webhookrule/" + x.ID }

// Fields encodes x for storage.
func (x WebhookRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookRule decodes a stored record.
func ParseWebhookRule(f codec.Fields) (WebhookRule, error) {
	var x WebhookRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookrule: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrule %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookRule) String() string {
	return fmt.Sprintf("WebhookRule(%s, group=%q, priority=%d, archived=%t)", x.ID, x.Group, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookRule) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookRule) Normalize() WebhookRule {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortWebhookRules orders by priority descending, then by ID.
func SortWebhookRules(xs []WebhookRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookRules drops the archived entries, preserving order.
func ActiveWebhookRules(xs []WebhookRule) []WebhookRule {
	out := make([]WebhookRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookRules adds up the priority of every active entry.
func SumWebhookRules(xs []WebhookRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWebhookRules keys the entries by ID, last write winning.
func IndexWebhookRules(xs []WebhookRule) map[string]WebhookRule {
	out := make(map[string]WebhookRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookRule builds a valid WebhookRule, for tests and for the seed data the cron
// binary writes.
func SampleWebhookRule(id string, at time.Time) WebhookRule {
	return WebhookRule{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Source:   "seeded by webhook.SampleWebhookRule",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookHoldSchema is the stored shape of a WebhookHold.
var WebhookHoldSchema = schema.New("webhook.webhookhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookHold is one webhook hold.
type WebhookHold struct {
	ID       string
	Comment  string
	Group    string
	Source   string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookHold) Key() string { return "webhook/webhookhold/" + x.ID }

// Fields encodes x for storage.
func (x WebhookHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"group":    x.Group,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookHold decodes a stored record.
func ParseWebhookHold(f codec.Fields) (WebhookHold, error) {
	var x WebhookHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookhold: missing id")
	}
	x.Comment = f["comment"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookhold %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookHold) String() string {
	return fmt.Sprintf("WebhookHold(%s, comment=%q, priority=%d, archived=%t)", x.ID, x.Comment, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookHold) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookHold) Normalize() WebhookHold {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortWebhookHolds orders by priority descending, then by ID.
func SortWebhookHolds(xs []WebhookHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookHolds drops the archived entries, preserving order.
func ActiveWebhookHolds(xs []WebhookHold) []WebhookHold {
	out := make([]WebhookHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookHolds adds up the priority of every active entry.
func SumWebhookHolds(xs []WebhookHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexWebhookHolds keys the entries by ID, last write winning.
func IndexWebhookHolds(xs []WebhookHold) map[string]WebhookHold {
	out := make(map[string]WebhookHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookHold builds a valid WebhookHold, for tests and for the seed data the cron
// binary writes.
func SampleWebhookHold(id string, at time.Time) WebhookHold {
	return WebhookHold{
		ID:       id,
		Comment:  "comment-" + id,
		Group:    "group-" + id,
		Source:   "seeded by webhook.SampleWebhookHold",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// WebhookClaimSchema is the stored shape of a WebhookClaim.
var WebhookClaimSchema = schema.New("webhook.webhookclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookClaim is one webhook claim.
type WebhookClaim struct {
	ID        string
	Reference string
	Category  string
	Channel   string
	Retained  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x WebhookClaim) Key() string { return "webhook/webhookclaim/" + x.ID }

// Fields encodes x for storage.
func (x WebhookClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"category":  x.Category,
		"channel":   x.Channel,
		"retained":  strconv.Itoa(x.Retained),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookClaim decodes a stored record.
func ParseWebhookClaim(f codec.Fields) (WebhookClaim, error) {
	var x WebhookClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookclaim: missing id")
	}
	x.Reference = f["reference"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookClaim) String() string {
	return fmt.Sprintf("WebhookClaim(%s, reference=%q, retained=%d, archived=%t)", x.ID, x.Reference, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookClaim) Normalize() WebhookClaim {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortWebhookClaims orders by retained descending, then by ID.
func SortWebhookClaims(xs []WebhookClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookClaims drops the archived entries, preserving order.
func ActiveWebhookClaims(xs []WebhookClaim) []WebhookClaim {
	out := make([]WebhookClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookClaims adds up the retained of every active entry.
func SumWebhookClaims(xs []WebhookClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexWebhookClaims keys the entries by ID, last write winning.
func IndexWebhookClaims(xs []WebhookClaim) map[string]WebhookClaim {
	out := make(map[string]WebhookClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookClaim builds a valid WebhookClaim, for tests and for the seed data the cron
// binary writes.
func SampleWebhookClaim(id string, at time.Time) WebhookClaim {
	return WebhookClaim{
		ID:        id,
		Reference: "reference-" + id,
		Category:  "category-" + id,
		Channel:   "seeded by webhook.SampleWebhookClaim",
		Retained:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// WebhookCursorSchema is the stored shape of a WebhookCursor.
var WebhookCursorSchema = schema.New("webhook.webhookcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// WebhookCursor is one webhook cursor.
type WebhookCursor struct {
	ID       string
	Region   string
	Label    string
	Group    string
	Retained int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x WebhookCursor) Key() string { return "webhook/webhookcursor/" + x.ID }

// Fields encodes x for storage.
func (x WebhookCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseWebhookCursor decodes a stored record.
func ParseWebhookCursor(f codec.Fields) (WebhookCursor, error) {
	var x WebhookCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("webhook: webhookcursor: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookcursor %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("webhook: webhookcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("webhook: webhookcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("webhook: webhookcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x WebhookCursor) String() string {
	return fmt.Sprintf("WebhookCursor(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x WebhookCursor) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x WebhookCursor) Normalize() WebhookCursor {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortWebhookCursors orders by retained descending, then by ID.
func SortWebhookCursors(xs []WebhookCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveWebhookCursors drops the archived entries, preserving order.
func ActiveWebhookCursors(xs []WebhookCursor) []WebhookCursor {
	out := make([]WebhookCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumWebhookCursors adds up the retained of every active entry.
func SumWebhookCursors(xs []WebhookCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexWebhookCursors keys the entries by ID, last write winning.
func IndexWebhookCursors(xs []WebhookCursor) map[string]WebhookCursor {
	out := make(map[string]WebhookCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleWebhookCursor builds a valid WebhookCursor, for tests and for the seed data the cron
// binary writes.
func SampleWebhookCursor(id string, at time.Time) WebhookCursor {
	return WebhookCursor{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Group:    "seeded by webhook.SampleWebhookCursor",
		Retained: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	WebhookRecordSchema,
	WebhookEntrySchema,
	WebhookItemSchema,
	WebhookDraftSchema,
	WebhookRevisionSchema,
	WebhookSnapshotSchema,
	WebhookBatchSchema,
	WebhookRequestSchema,
	WebhookSummarySchema,
	WebhookPolicySchema,
	WebhookLinkSchema,
	WebhookNoteSchema,
	WebhookWindowSchema,
	WebhookQuotaSchema,
	WebhookBucketSchema,
	WebhookLaneSchema,
	WebhookSlotSchema,
	WebhookTagSchema,
	WebhookRuleSchema,
	WebhookHoldSchema,
	WebhookClaimSchema,
	WebhookCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"webhook/webhookrecord/",
	"webhook/webhookentry/",
	"webhook/webhookitem/",
	"webhook/webhookdraft/",
	"webhook/webhookrevision/",
	"webhook/webhooksnapshot/",
	"webhook/webhookbatch/",
	"webhook/webhookrequest/",
	"webhook/webhooksummary/",
	"webhook/webhookpolicy/",
	"webhook/webhooklink/",
	"webhook/webhooknote/",
	"webhook/webhookwindow/",
	"webhook/webhookquota/",
	"webhook/webhookbucket/",
	"webhook/webhooklane/",
	"webhook/webhookslot/",
	"webhook/webhooktag/",
	"webhook/webhookrule/",
	"webhook/webhookhold/",
	"webhook/webhookclaim/",
	"webhook/webhookcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "webhook: " + strings.Join(names, ", ")
}
