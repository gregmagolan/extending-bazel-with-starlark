// Package tenant holds the tenant domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package tenant

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// TenantRecordSchema is the stored shape of a TenantRecord.
var TenantRecordSchema = schema.New("tenant.tenantrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantRecord is one tenant record.
type TenantRecord struct {
	ID       string
	Owner    string
	Group    string
	Region   string
	Offset   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantRecord) Key() string { return "tenant/tenantrecord/" + x.ID }

// Fields encodes x for storage.
func (x TenantRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantRecord decodes a stored record.
func ParseTenantRecord(f codec.Fields) (TenantRecord, error) {
	var x TenantRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantrecord: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantRecord) String() string {
	return fmt.Sprintf("TenantRecord(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantRecord) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantRecord) Normalize() TenantRecord {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTenantRecords orders by offset descending, then by ID.
func SortTenantRecords(xs []TenantRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantRecords drops the archived entries, preserving order.
func ActiveTenantRecords(xs []TenantRecord) []TenantRecord {
	out := make([]TenantRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantRecords adds up the offset of every active entry.
func SumTenantRecords(xs []TenantRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTenantRecords keys the entries by ID, last write winning.
func IndexTenantRecords(xs []TenantRecord) map[string]TenantRecord {
	out := make(map[string]TenantRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantRecord builds a valid TenantRecord, for tests and for the seed data the cron
// binary writes.
func SampleTenantRecord(id string, at time.Time) TenantRecord {
	return TenantRecord{
		ID:      id,
		Owner:   "owner-" + id,
		Group:   "group-" + id,
		Region:  "seeded by tenant.SampleTenantRecord",
		Offset:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TenantEntrySchema is the stored shape of a TenantEntry.
var TenantEntrySchema = schema.New("tenant.tenantentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantEntry is one tenant entry.
type TenantEntry struct {
	ID        string
	Channel   string
	Owner     string
	Reference string
	Retained  int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantEntry) Key() string { return "tenant/tenantentry/" + x.ID }

// Fields encodes x for storage.
func (x TenantEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"channel":   x.Channel,
		"owner":     x.Owner,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantEntry decodes a stored record.
func ParseTenantEntry(f codec.Fields) (TenantEntry, error) {
	var x TenantEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantentry: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantEntry) String() string {
	return fmt.Sprintf("TenantEntry(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantEntry) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantEntry) Normalize() TenantEntry {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTenantEntrys orders by retained descending, then by ID.
func SortTenantEntrys(xs []TenantEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantEntrys drops the archived entries, preserving order.
func ActiveTenantEntrys(xs []TenantEntry) []TenantEntry {
	out := make([]TenantEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantEntrys adds up the retained of every active entry.
func SumTenantEntrys(xs []TenantEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTenantEntrys keys the entries by ID, last write winning.
func IndexTenantEntrys(xs []TenantEntry) map[string]TenantEntry {
	out := make(map[string]TenantEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantEntry builds a valid TenantEntry, for tests and for the seed data the cron
// binary writes.
func SampleTenantEntry(id string, at time.Time) TenantEntry {
	return TenantEntry{
		ID:        id,
		Channel:   "channel-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by tenant.SampleTenantEntry",
		Retained:  len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantItemSchema is the stored shape of a TenantItem.
var TenantItemSchema = schema.New("tenant.tenantitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantItem is one tenant item.
type TenantItem struct {
	ID       string
	Source   string
	Owner    string
	Channel  string
	Priority int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantItem) Key() string { return "tenant/tenantitem/" + x.ID }

// Fields encodes x for storage.
func (x TenantItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"priority": strconv.Itoa(x.Priority),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantItem decodes a stored record.
func ParseTenantItem(f codec.Fields) (TenantItem, error) {
	var x TenantItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantitem: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantitem %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantItem) String() string {
	return fmt.Sprintf("TenantItem(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantItem) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantItem) Normalize() TenantItem {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTenantItems orders by priority descending, then by ID.
func SortTenantItems(xs []TenantItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantItems drops the archived entries, preserving order.
func ActiveTenantItems(xs []TenantItem) []TenantItem {
	out := make([]TenantItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantItems adds up the priority of every active entry.
func SumTenantItems(xs []TenantItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTenantItems keys the entries by ID, last write winning.
func IndexTenantItems(xs []TenantItem) map[string]TenantItem {
	out := make(map[string]TenantItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantItem builds a valid TenantItem, for tests and for the seed data the cron
// binary writes.
func SampleTenantItem(id string, at time.Time) TenantItem {
	return TenantItem{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by tenant.SampleTenantItem",
		Priority: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantDraftSchema is the stored shape of a TenantDraft.
var TenantDraftSchema = schema.New("tenant.tenantdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantDraft is one tenant draft.
type TenantDraft struct {
	ID       string
	Name     string
	Label    string
	Group    string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantDraft) Key() string { return "tenant/tenantdraft/" + x.ID }

// Fields encodes x for storage.
func (x TenantDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantDraft decodes a stored record.
func ParseTenantDraft(f codec.Fields) (TenantDraft, error) {
	var x TenantDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantdraft: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantdraft %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantDraft) String() string {
	return fmt.Sprintf("TenantDraft(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantDraft) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantDraft) Normalize() TenantDraft {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortTenantDrafts orders by capacity descending, then by ID.
func SortTenantDrafts(xs []TenantDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantDrafts drops the archived entries, preserving order.
func ActiveTenantDrafts(xs []TenantDraft) []TenantDraft {
	out := make([]TenantDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantDrafts adds up the capacity of every active entry.
func SumTenantDrafts(xs []TenantDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTenantDrafts keys the entries by ID, last write winning.
func IndexTenantDrafts(xs []TenantDraft) map[string]TenantDraft {
	out := make(map[string]TenantDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantDraft builds a valid TenantDraft, for tests and for the seed data the cron
// binary writes.
func SampleTenantDraft(id string, at time.Time) TenantDraft {
	return TenantDraft{
		ID:       id,
		Name:     "name-" + id,
		Label:    "label-" + id,
		Group:    "seeded by tenant.SampleTenantDraft",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantRevisionSchema is the stored shape of a TenantRevision.
var TenantRevisionSchema = schema.New("tenant.tenantrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantRevision is one tenant revision.
type TenantRevision struct {
	ID        string
	Name      string
	Reference string
	Owner     string
	Offset    int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantRevision) Key() string { return "tenant/tenantrevision/" + x.ID }

// Fields encodes x for storage.
func (x TenantRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"owner":     x.Owner,
		"offset":    strconv.Itoa(x.Offset),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantRevision decodes a stored record.
func ParseTenantRevision(f codec.Fields) (TenantRevision, error) {
	var x TenantRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantrevision: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrevision %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantRevision) String() string {
	return fmt.Sprintf("TenantRevision(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantRevision) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantRevision) Normalize() TenantRevision {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantRevisions orders by offset descending, then by ID.
func SortTenantRevisions(xs []TenantRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantRevisions drops the archived entries, preserving order.
func ActiveTenantRevisions(xs []TenantRevision) []TenantRevision {
	out := make([]TenantRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantRevisions adds up the offset of every active entry.
func SumTenantRevisions(xs []TenantRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTenantRevisions keys the entries by ID, last write winning.
func IndexTenantRevisions(xs []TenantRevision) map[string]TenantRevision {
	out := make(map[string]TenantRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantRevision builds a valid TenantRevision, for tests and for the seed data the cron
// binary writes.
func SampleTenantRevision(id string, at time.Time) TenantRevision {
	return TenantRevision{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by tenant.SampleTenantRevision",
		Offset:    len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantSnapshotSchema is the stored shape of a TenantSnapshot.
var TenantSnapshotSchema = schema.New("tenant.tenantsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantSnapshot is one tenant snapshot.
type TenantSnapshot struct {
	ID       string
	Owner    string
	Category string
	Comment  string
	Priority int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantSnapshot) Key() string { return "tenant/tenantsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x TenantSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"category": x.Category,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantSnapshot decodes a stored record.
func ParseTenantSnapshot(f codec.Fields) (TenantSnapshot, error) {
	var x TenantSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantsnapshot: missing id")
	}
	x.Owner = f["owner"]
	x.Category = f["category"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantsnapshot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantSnapshot) String() string {
	return fmt.Sprintf("TenantSnapshot(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantSnapshot) Normalize() TenantSnapshot {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantSnapshots orders by priority descending, then by ID.
func SortTenantSnapshots(xs []TenantSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantSnapshots drops the archived entries, preserving order.
func ActiveTenantSnapshots(xs []TenantSnapshot) []TenantSnapshot {
	out := make([]TenantSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantSnapshots adds up the priority of every active entry.
func SumTenantSnapshots(xs []TenantSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTenantSnapshots keys the entries by ID, last write winning.
func IndexTenantSnapshots(xs []TenantSnapshot) map[string]TenantSnapshot {
	out := make(map[string]TenantSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantSnapshot builds a valid TenantSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleTenantSnapshot(id string, at time.Time) TenantSnapshot {
	return TenantSnapshot{
		ID:       id,
		Owner:    "owner-" + id,
		Category: "category-" + id,
		Comment:  "seeded by tenant.SampleTenantSnapshot",
		Priority: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantBatchSchema is the stored shape of a TenantBatch.
var TenantBatchSchema = schema.New("tenant.tenantbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantBatch is one tenant batch.
type TenantBatch struct {
	ID        string
	Reference string
	Name      string
	Group     string
	Version   int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantBatch) Key() string { return "tenant/tenantbatch/" + x.ID }

// Fields encodes x for storage.
func (x TenantBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"name":      x.Name,
		"group":     x.Group,
		"version":   strconv.Itoa(x.Version),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantBatch decodes a stored record.
func ParseTenantBatch(f codec.Fields) (TenantBatch, error) {
	var x TenantBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantbatch: missing id")
	}
	x.Reference = f["reference"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantbatch %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantBatch) String() string {
	return fmt.Sprintf("TenantBatch(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantBatch) Normalize() TenantBatch {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTenantBatchs orders by version descending, then by ID.
func SortTenantBatchs(xs []TenantBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantBatchs drops the archived entries, preserving order.
func ActiveTenantBatchs(xs []TenantBatch) []TenantBatch {
	out := make([]TenantBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantBatchs adds up the version of every active entry.
func SumTenantBatchs(xs []TenantBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTenantBatchs keys the entries by ID, last write winning.
func IndexTenantBatchs(xs []TenantBatch) map[string]TenantBatch {
	out := make(map[string]TenantBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantBatch builds a valid TenantBatch, for tests and for the seed data the cron
// binary writes.
func SampleTenantBatch(id string, at time.Time) TenantBatch {
	return TenantBatch{
		ID:        id,
		Reference: "reference-" + id,
		Name:      "name-" + id,
		Group:     "seeded by tenant.SampleTenantBatch",
		Version:   len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantRequestSchema is the stored shape of a TenantRequest.
var TenantRequestSchema = schema.New("tenant.tenantrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantRequest is one tenant request.
type TenantRequest struct {
	ID        string
	Reference string
	Source    string
	Region    string
	Attempts  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantRequest) Key() string { return "tenant/tenantrequest/" + x.ID }

// Fields encodes x for storage.
func (x TenantRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"source":    x.Source,
		"region":    x.Region,
		"attempts":  strconv.Itoa(x.Attempts),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantRequest decodes a stored record.
func ParseTenantRequest(f codec.Fields) (TenantRequest, error) {
	var x TenantRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantrequest: missing id")
	}
	x.Reference = f["reference"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrequest %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantRequest) String() string {
	return fmt.Sprintf("TenantRequest(%s, reference=%q, attempts=%d, archived=%t)", x.ID, x.Reference, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantRequest) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantRequest) Normalize() TenantRequest {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortTenantRequests orders by attempts descending, then by ID.
func SortTenantRequests(xs []TenantRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantRequests drops the archived entries, preserving order.
func ActiveTenantRequests(xs []TenantRequest) []TenantRequest {
	out := make([]TenantRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantRequests adds up the attempts of every active entry.
func SumTenantRequests(xs []TenantRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTenantRequests keys the entries by ID, last write winning.
func IndexTenantRequests(xs []TenantRequest) map[string]TenantRequest {
	out := make(map[string]TenantRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantRequest builds a valid TenantRequest, for tests and for the seed data the cron
// binary writes.
func SampleTenantRequest(id string, at time.Time) TenantRequest {
	return TenantRequest{
		ID:        id,
		Reference: "reference-" + id,
		Source:    "source-" + id,
		Region:    "seeded by tenant.SampleTenantRequest",
		Attempts:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantSummarySchema is the stored shape of a TenantSummary.
var TenantSummarySchema = schema.New("tenant.tenantsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantSummary is one tenant summary.
type TenantSummary struct {
	ID       string
	Label    string
	Category string
	Owner    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantSummary) Key() string { return "tenant/tenantsummary/" + x.ID }

// Fields encodes x for storage.
func (x TenantSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"owner":    x.Owner,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantSummary decodes a stored record.
func ParseTenantSummary(f codec.Fields) (TenantSummary, error) {
	var x TenantSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantsummary: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantsummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantSummary) String() string {
	return fmt.Sprintf("TenantSummary(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantSummary) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantSummary) Normalize() TenantSummary {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantSummarys orders by offset descending, then by ID.
func SortTenantSummarys(xs []TenantSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantSummarys drops the archived entries, preserving order.
func ActiveTenantSummarys(xs []TenantSummary) []TenantSummary {
	out := make([]TenantSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantSummarys adds up the offset of every active entry.
func SumTenantSummarys(xs []TenantSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTenantSummarys keys the entries by ID, last write winning.
func IndexTenantSummarys(xs []TenantSummary) map[string]TenantSummary {
	out := make(map[string]TenantSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantSummary builds a valid TenantSummary, for tests and for the seed data the cron
// binary writes.
func SampleTenantSummary(id string, at time.Time) TenantSummary {
	return TenantSummary{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Owner:    "seeded by tenant.SampleTenantSummary",
		Offset:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantPolicySchema is the stored shape of a TenantPolicy.
var TenantPolicySchema = schema.New("tenant.tenantpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantPolicy is one tenant policy.
type TenantPolicy struct {
	ID       string
	Label    string
	Owner    string
	Category string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantPolicy) Key() string { return "tenant/tenantpolicy/" + x.ID }

// Fields encodes x for storage.
func (x TenantPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"owner":    x.Owner,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantPolicy decodes a stored record.
func ParseTenantPolicy(f codec.Fields) (TenantPolicy, error) {
	var x TenantPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantpolicy: missing id")
	}
	x.Label = f["label"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantpolicy %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantPolicy) String() string {
	return fmt.Sprintf("TenantPolicy(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantPolicy) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantPolicy) Normalize() TenantPolicy {
	x.Label = strings.TrimSpace(x.Label)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantPolicys orders by count descending, then by ID.
func SortTenantPolicys(xs []TenantPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantPolicys drops the archived entries, preserving order.
func ActiveTenantPolicys(xs []TenantPolicy) []TenantPolicy {
	out := make([]TenantPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantPolicys adds up the count of every active entry.
func SumTenantPolicys(xs []TenantPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTenantPolicys keys the entries by ID, last write winning.
func IndexTenantPolicys(xs []TenantPolicy) map[string]TenantPolicy {
	out := make(map[string]TenantPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantPolicy builds a valid TenantPolicy, for tests and for the seed data the cron
// binary writes.
func SampleTenantPolicy(id string, at time.Time) TenantPolicy {
	return TenantPolicy{
		ID:       id,
		Label:    "label-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by tenant.SampleTenantPolicy",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantLinkSchema is the stored shape of a TenantLink.
var TenantLinkSchema = schema.New("tenant.tenantlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantLink is one tenant link.
type TenantLink struct {
	ID        string
	Region    string
	Reference string
	Source    string
	Priority  int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantLink) Key() string { return "tenant/tenantlink/" + x.ID }

// Fields encodes x for storage.
func (x TenantLink) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"source":    x.Source,
		"priority":  strconv.Itoa(x.Priority),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantLink decodes a stored record.
func ParseTenantLink(f codec.Fields) (TenantLink, error) {
	var x TenantLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantlink: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantlink %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantLink) String() string {
	return fmt.Sprintf("TenantLink(%s, region=%q, priority=%d, archived=%t)", x.ID, x.Region, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantLink) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantLink) Normalize() TenantLink {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTenantLinks orders by priority descending, then by ID.
func SortTenantLinks(xs []TenantLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantLinks drops the archived entries, preserving order.
func ActiveTenantLinks(xs []TenantLink) []TenantLink {
	out := make([]TenantLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantLinks adds up the priority of every active entry.
func SumTenantLinks(xs []TenantLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTenantLinks keys the entries by ID, last write winning.
func IndexTenantLinks(xs []TenantLink) map[string]TenantLink {
	out := make(map[string]TenantLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantLink builds a valid TenantLink, for tests and for the seed data the cron
// binary writes.
func SampleTenantLink(id string, at time.Time) TenantLink {
	return TenantLink{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by tenant.SampleTenantLink",
		Priority:  len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantNoteSchema is the stored shape of a TenantNote.
var TenantNoteSchema = schema.New("tenant.tenantnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantNote is one tenant note.
type TenantNote struct {
	ID       string
	Category string
	Name     string
	Comment  string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantNote) Key() string { return "tenant/tenantnote/" + x.ID }

// Fields encodes x for storage.
func (x TenantNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantNote decodes a stored record.
func ParseTenantNote(f codec.Fields) (TenantNote, error) {
	var x TenantNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantnote: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantnote %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantNote) String() string {
	return fmt.Sprintf("TenantNote(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantNote) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantNote) Normalize() TenantNote {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantNotes orders by count descending, then by ID.
func SortTenantNotes(xs []TenantNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantNotes drops the archived entries, preserving order.
func ActiveTenantNotes(xs []TenantNote) []TenantNote {
	out := make([]TenantNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantNotes adds up the count of every active entry.
func SumTenantNotes(xs []TenantNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTenantNotes keys the entries by ID, last write winning.
func IndexTenantNotes(xs []TenantNote) map[string]TenantNote {
	out := make(map[string]TenantNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantNote builds a valid TenantNote, for tests and for the seed data the cron
// binary writes.
func SampleTenantNote(id string, at time.Time) TenantNote {
	return TenantNote{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by tenant.SampleTenantNote",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantWindowSchema is the stored shape of a TenantWindow.
var TenantWindowSchema = schema.New("tenant.tenantwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantWindow is one tenant window.
type TenantWindow struct {
	ID        string
	Category  string
	Label     string
	Reference string
	Attempts  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantWindow) Key() string { return "tenant/tenantwindow/" + x.ID }

// Fields encodes x for storage.
func (x TenantWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"label":     x.Label,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantWindow decodes a stored record.
func ParseTenantWindow(f codec.Fields) (TenantWindow, error) {
	var x TenantWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantwindow: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantwindow %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantWindow) String() string {
	return fmt.Sprintf("TenantWindow(%s, category=%q, attempts=%d, archived=%t)", x.ID, x.Category, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantWindow) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantWindow) Normalize() TenantWindow {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTenantWindows orders by attempts descending, then by ID.
func SortTenantWindows(xs []TenantWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantWindows drops the archived entries, preserving order.
func ActiveTenantWindows(xs []TenantWindow) []TenantWindow {
	out := make([]TenantWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantWindows adds up the attempts of every active entry.
func SumTenantWindows(xs []TenantWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTenantWindows keys the entries by ID, last write winning.
func IndexTenantWindows(xs []TenantWindow) map[string]TenantWindow {
	out := make(map[string]TenantWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantWindow builds a valid TenantWindow, for tests and for the seed data the cron
// binary writes.
func SampleTenantWindow(id string, at time.Time) TenantWindow {
	return TenantWindow{
		ID:        id,
		Category:  "category-" + id,
		Label:     "label-" + id,
		Reference: "seeded by tenant.SampleTenantWindow",
		Attempts:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantQuotaSchema is the stored shape of a TenantQuota.
var TenantQuotaSchema = schema.New("tenant.tenantquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantQuota is one tenant quota.
type TenantQuota struct {
	ID       string
	Channel  string
	Region   string
	Name     string
	Version  int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantQuota) Key() string { return "tenant/tenantquota/" + x.ID }

// Fields encodes x for storage.
func (x TenantQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"name":     x.Name,
		"version":  strconv.Itoa(x.Version),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantQuota decodes a stored record.
func ParseTenantQuota(f codec.Fields) (TenantQuota, error) {
	var x TenantQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantquota: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantquota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantQuota) String() string {
	return fmt.Sprintf("TenantQuota(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantQuota) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantQuota) Normalize() TenantQuota {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTenantQuotas orders by version descending, then by ID.
func SortTenantQuotas(xs []TenantQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantQuotas drops the archived entries, preserving order.
func ActiveTenantQuotas(xs []TenantQuota) []TenantQuota {
	out := make([]TenantQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantQuotas adds up the version of every active entry.
func SumTenantQuotas(xs []TenantQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTenantQuotas keys the entries by ID, last write winning.
func IndexTenantQuotas(xs []TenantQuota) map[string]TenantQuota {
	out := make(map[string]TenantQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantQuota builds a valid TenantQuota, for tests and for the seed data the cron
// binary writes.
func SampleTenantQuota(id string, at time.Time) TenantQuota {
	return TenantQuota{
		ID:      id,
		Channel: "channel-" + id,
		Region:  "region-" + id,
		Name:    "seeded by tenant.SampleTenantQuota",
		Version: len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TenantBucketSchema is the stored shape of a TenantBucket.
var TenantBucketSchema = schema.New("tenant.tenantbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantBucket is one tenant bucket.
type TenantBucket struct {
	ID       string
	Group    string
	Owner    string
	Label    string
	Capacity int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantBucket) Key() string { return "tenant/tenantbucket/" + x.ID }

// Fields encodes x for storage.
func (x TenantBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantBucket decodes a stored record.
func ParseTenantBucket(f codec.Fields) (TenantBucket, error) {
	var x TenantBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantbucket: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantbucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantBucket) String() string {
	return fmt.Sprintf("TenantBucket(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantBucket) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantBucket) Normalize() TenantBucket {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTenantBuckets orders by capacity descending, then by ID.
func SortTenantBuckets(xs []TenantBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantBuckets drops the archived entries, preserving order.
func ActiveTenantBuckets(xs []TenantBucket) []TenantBucket {
	out := make([]TenantBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantBuckets adds up the capacity of every active entry.
func SumTenantBuckets(xs []TenantBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTenantBuckets keys the entries by ID, last write winning.
func IndexTenantBuckets(xs []TenantBucket) map[string]TenantBucket {
	out := make(map[string]TenantBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantBucket builds a valid TenantBucket, for tests and for the seed data the cron
// binary writes.
func SampleTenantBucket(id string, at time.Time) TenantBucket {
	return TenantBucket{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by tenant.SampleTenantBucket",
		Capacity: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantLaneSchema is the stored shape of a TenantLane.
var TenantLaneSchema = schema.New("tenant.tenantlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantLane is one tenant lane.
type TenantLane struct {
	ID       string
	Label    string
	Category string
	Name     string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantLane) Key() string { return "tenant/tenantlane/" + x.ID }

// Fields encodes x for storage.
func (x TenantLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"category": x.Category,
		"name":     x.Name,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantLane decodes a stored record.
func ParseTenantLane(f codec.Fields) (TenantLane, error) {
	var x TenantLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantlane: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Name = f["name"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantlane %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantLane) String() string {
	return fmt.Sprintf("TenantLane(%s, label=%q, capacity=%d, archived=%t)", x.ID, x.Label, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantLane) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantLane) Normalize() TenantLane {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantLanes orders by capacity descending, then by ID.
func SortTenantLanes(xs []TenantLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantLanes drops the archived entries, preserving order.
func ActiveTenantLanes(xs []TenantLane) []TenantLane {
	out := make([]TenantLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantLanes adds up the capacity of every active entry.
func SumTenantLanes(xs []TenantLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTenantLanes keys the entries by ID, last write winning.
func IndexTenantLanes(xs []TenantLane) map[string]TenantLane {
	out := make(map[string]TenantLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantLane builds a valid TenantLane, for tests and for the seed data the cron
// binary writes.
func SampleTenantLane(id string, at time.Time) TenantLane {
	return TenantLane{
		ID:       id,
		Label:    "label-" + id,
		Category: "category-" + id,
		Name:     "seeded by tenant.SampleTenantLane",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantSlotSchema is the stored shape of a TenantSlot.
var TenantSlotSchema = schema.New("tenant.tenantslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantSlot is one tenant slot.
type TenantSlot struct {
	ID       string
	Channel  string
	Group    string
	Category string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantSlot) Key() string { return "tenant/tenantslot/" + x.ID }

// Fields encodes x for storage.
func (x TenantSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantSlot decodes a stored record.
func ParseTenantSlot(f codec.Fields) (TenantSlot, error) {
	var x TenantSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantslot: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantslot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantSlot) String() string {
	return fmt.Sprintf("TenantSlot(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantSlot) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantSlot) Normalize() TenantSlot {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTenantSlots orders by attempts descending, then by ID.
func SortTenantSlots(xs []TenantSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantSlots drops the archived entries, preserving order.
func ActiveTenantSlots(xs []TenantSlot) []TenantSlot {
	out := make([]TenantSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantSlots adds up the attempts of every active entry.
func SumTenantSlots(xs []TenantSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTenantSlots keys the entries by ID, last write winning.
func IndexTenantSlots(xs []TenantSlot) map[string]TenantSlot {
	out := make(map[string]TenantSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantSlot builds a valid TenantSlot, for tests and for the seed data the cron
// binary writes.
func SampleTenantSlot(id string, at time.Time) TenantSlot {
	return TenantSlot{
		ID:       id,
		Channel:  "channel-" + id,
		Group:    "group-" + id,
		Category: "seeded by tenant.SampleTenantSlot",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantTagSchema is the stored shape of a TenantTag.
var TenantTagSchema = schema.New("tenant.tenanttag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantTag is one tenant tag.
type TenantTag struct {
	ID       string
	Category string
	Owner    string
	Source   string
	Weight   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantTag) Key() string { return "tenant/tenanttag/" + x.ID }

// Fields encodes x for storage.
func (x TenantTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantTag decodes a stored record.
func ParseTenantTag(f codec.Fields) (TenantTag, error) {
	var x TenantTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenanttag: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("tenant: tenanttag %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenanttag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenanttag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenanttag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantTag) String() string {
	return fmt.Sprintf("TenantTag(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantTag) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantTag) Normalize() TenantTag {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantTags orders by weight descending, then by ID.
func SortTenantTags(xs []TenantTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantTags drops the archived entries, preserving order.
func ActiveTenantTags(xs []TenantTag) []TenantTag {
	out := make([]TenantTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantTags adds up the weight of every active entry.
func SumTenantTags(xs []TenantTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTenantTags keys the entries by ID, last write winning.
func IndexTenantTags(xs []TenantTag) map[string]TenantTag {
	out := make(map[string]TenantTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantTag builds a valid TenantTag, for tests and for the seed data the cron
// binary writes.
func SampleTenantTag(id string, at time.Time) TenantTag {
	return TenantTag{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by tenant.SampleTenantTag",
		Weight:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantRuleSchema is the stored shape of a TenantRule.
var TenantRuleSchema = schema.New("tenant.tenantrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantRule is one tenant rule.
type TenantRule struct {
	ID        string
	Reference string
	Channel   string
	Comment   string
	Attempts  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantRule) Key() string { return "tenant/tenantrule/" + x.ID }

// Fields encodes x for storage.
func (x TenantRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"comment":   x.Comment,
		"attempts":  strconv.Itoa(x.Attempts),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantRule decodes a stored record.
func ParseTenantRule(f codec.Fields) (TenantRule, error) {
	var x TenantRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantrule: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrule %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantRule) String() string {
	return fmt.Sprintf("TenantRule(%s, reference=%q, attempts=%d, archived=%t)", x.ID, x.Reference, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantRule) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantRule) Normalize() TenantRule {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTenantRules orders by attempts descending, then by ID.
func SortTenantRules(xs []TenantRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantRules drops the archived entries, preserving order.
func ActiveTenantRules(xs []TenantRule) []TenantRule {
	out := make([]TenantRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantRules adds up the attempts of every active entry.
func SumTenantRules(xs []TenantRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTenantRules keys the entries by ID, last write winning.
func IndexTenantRules(xs []TenantRule) map[string]TenantRule {
	out := make(map[string]TenantRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantRule builds a valid TenantRule, for tests and for the seed data the cron
// binary writes.
func SampleTenantRule(id string, at time.Time) TenantRule {
	return TenantRule{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Comment:   "seeded by tenant.SampleTenantRule",
		Attempts:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantHoldSchema is the stored shape of a TenantHold.
var TenantHoldSchema = schema.New("tenant.tenanthold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantHold is one tenant hold.
type TenantHold struct {
	ID       string
	Category string
	Label    string
	Comment  string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TenantHold) Key() string { return "tenant/tenanthold/" + x.ID }

// Fields encodes x for storage.
func (x TenantHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"comment":  x.Comment,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantHold decodes a stored record.
func ParseTenantHold(f codec.Fields) (TenantHold, error) {
	var x TenantHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenanthold: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("tenant: tenanthold %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("tenant: tenanthold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenanthold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenanthold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantHold) String() string {
	return fmt.Sprintf("TenantHold(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantHold) Normalize() TenantHold {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTenantHolds orders by weight descending, then by ID.
func SortTenantHolds(xs []TenantHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantHolds drops the archived entries, preserving order.
func ActiveTenantHolds(xs []TenantHold) []TenantHold {
	out := make([]TenantHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantHolds adds up the weight of every active entry.
func SumTenantHolds(xs []TenantHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTenantHolds keys the entries by ID, last write winning.
func IndexTenantHolds(xs []TenantHold) map[string]TenantHold {
	out := make(map[string]TenantHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantHold builds a valid TenantHold, for tests and for the seed data the cron
// binary writes.
func SampleTenantHold(id string, at time.Time) TenantHold {
	return TenantHold{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Comment:  "seeded by tenant.SampleTenantHold",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TenantClaimSchema is the stored shape of a TenantClaim.
var TenantClaimSchema = schema.New("tenant.tenantclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantClaim is one tenant claim.
type TenantClaim struct {
	ID        string
	Name      string
	Reference string
	Owner     string
	Offset    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantClaim) Key() string { return "tenant/tenantclaim/" + x.ID }

// Fields encodes x for storage.
func (x TenantClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"owner":     x.Owner,
		"offset":    strconv.Itoa(x.Offset),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantClaim decodes a stored record.
func ParseTenantClaim(f codec.Fields) (TenantClaim, error) {
	var x TenantClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantclaim: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantclaim %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantClaim) String() string {
	return fmt.Sprintf("TenantClaim(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantClaim) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantClaim) Normalize() TenantClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTenantClaims orders by offset descending, then by ID.
func SortTenantClaims(xs []TenantClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantClaims drops the archived entries, preserving order.
func ActiveTenantClaims(xs []TenantClaim) []TenantClaim {
	out := make([]TenantClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantClaims adds up the offset of every active entry.
func SumTenantClaims(xs []TenantClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTenantClaims keys the entries by ID, last write winning.
func IndexTenantClaims(xs []TenantClaim) map[string]TenantClaim {
	out := make(map[string]TenantClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantClaim builds a valid TenantClaim, for tests and for the seed data the cron
// binary writes.
func SampleTenantClaim(id string, at time.Time) TenantClaim {
	return TenantClaim{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by tenant.SampleTenantClaim",
		Offset:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TenantCursorSchema is the stored shape of a TenantCursor.
var TenantCursorSchema = schema.New("tenant.tenantcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TenantCursor is one tenant cursor.
type TenantCursor struct {
	ID        string
	Category  string
	Reference string
	Group     string
	Priority  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TenantCursor) Key() string { return "tenant/tenantcursor/" + x.ID }

// Fields encodes x for storage.
func (x TenantCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"group":     x.Group,
		"priority":  strconv.Itoa(x.Priority),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTenantCursor decodes a stored record.
func ParseTenantCursor(f codec.Fields) (TenantCursor, error) {
	var x TenantCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("tenant: tenantcursor: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantcursor %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("tenant: tenantcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("tenant: tenantcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("tenant: tenantcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TenantCursor) String() string {
	return fmt.Sprintf("TenantCursor(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TenantCursor) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TenantCursor) Normalize() TenantCursor {
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTenantCursors orders by priority descending, then by ID.
func SortTenantCursors(xs []TenantCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTenantCursors drops the archived entries, preserving order.
func ActiveTenantCursors(xs []TenantCursor) []TenantCursor {
	out := make([]TenantCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTenantCursors adds up the priority of every active entry.
func SumTenantCursors(xs []TenantCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTenantCursors keys the entries by ID, last write winning.
func IndexTenantCursors(xs []TenantCursor) map[string]TenantCursor {
	out := make(map[string]TenantCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTenantCursor builds a valid TenantCursor, for tests and for the seed data the cron
// binary writes.
func SampleTenantCursor(id string, at time.Time) TenantCursor {
	return TenantCursor{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Group:     "seeded by tenant.SampleTenantCursor",
		Priority:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	TenantRecordSchema,
	TenantEntrySchema,
	TenantItemSchema,
	TenantDraftSchema,
	TenantRevisionSchema,
	TenantSnapshotSchema,
	TenantBatchSchema,
	TenantRequestSchema,
	TenantSummarySchema,
	TenantPolicySchema,
	TenantLinkSchema,
	TenantNoteSchema,
	TenantWindowSchema,
	TenantQuotaSchema,
	TenantBucketSchema,
	TenantLaneSchema,
	TenantSlotSchema,
	TenantTagSchema,
	TenantRuleSchema,
	TenantHoldSchema,
	TenantClaimSchema,
	TenantCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"tenant/tenantrecord/",
	"tenant/tenantentry/",
	"tenant/tenantitem/",
	"tenant/tenantdraft/",
	"tenant/tenantrevision/",
	"tenant/tenantsnapshot/",
	"tenant/tenantbatch/",
	"tenant/tenantrequest/",
	"tenant/tenantsummary/",
	"tenant/tenantpolicy/",
	"tenant/tenantlink/",
	"tenant/tenantnote/",
	"tenant/tenantwindow/",
	"tenant/tenantquota/",
	"tenant/tenantbucket/",
	"tenant/tenantlane/",
	"tenant/tenantslot/",
	"tenant/tenanttag/",
	"tenant/tenantrule/",
	"tenant/tenanthold/",
	"tenant/tenantclaim/",
	"tenant/tenantcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "tenant: " + strings.Join(names, ", ")
}
