// Package route holds the route domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package route

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// RouteRecordSchema is the stored shape of a RouteRecord.
var RouteRecordSchema = schema.New("route.routerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteRecord is one route record.
type RouteRecord struct {
	ID       string
	Category string
	Channel  string
	Label    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteRecord) Key() string { return "route/routerecord/" + x.ID }

// Fields encodes x for storage.
func (x RouteRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteRecord decodes a stored record.
func ParseRouteRecord(f codec.Fields) (RouteRecord, error) {
	var x RouteRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routerecord: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("route: routerecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("route: routerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteRecord) String() string {
	return fmt.Sprintf("RouteRecord(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteRecord) Normalize() RouteRecord {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRouteRecords orders by capacity descending, then by ID.
func SortRouteRecords(xs []RouteRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteRecords drops the archived entries, preserving order.
func ActiveRouteRecords(xs []RouteRecord) []RouteRecord {
	out := make([]RouteRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteRecords adds up the capacity of every active entry.
func SumRouteRecords(xs []RouteRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRouteRecords keys the entries by ID, last write winning.
func IndexRouteRecords(xs []RouteRecord) map[string]RouteRecord {
	out := make(map[string]RouteRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteRecord builds a valid RouteRecord, for tests and for the seed data the cron
// binary writes.
func SampleRouteRecord(id string, at time.Time) RouteRecord {
	return RouteRecord{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Label:    "seeded by route.SampleRouteRecord",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteEntrySchema is the stored shape of a RouteEntry.
var RouteEntrySchema = schema.New("route.routeentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteEntry is one route entry.
type RouteEntry struct {
	ID        string
	Region    string
	Reference string
	Owner     string
	Retained  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RouteEntry) Key() string { return "route/routeentry/" + x.ID }

// Fields encodes x for storage.
func (x RouteEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"owner":     x.Owner,
		"retained":  strconv.Itoa(x.Retained),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteEntry decodes a stored record.
func ParseRouteEntry(f codec.Fields) (RouteEntry, error) {
	var x RouteEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routeentry: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routeentry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("route: routeentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routeentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routeentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteEntry) String() string {
	return fmt.Sprintf("RouteEntry(%s, region=%q, retained=%d, archived=%t)", x.ID, x.Region, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteEntry) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteEntry) Normalize() RouteEntry {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRouteEntrys orders by retained descending, then by ID.
func SortRouteEntrys(xs []RouteEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteEntrys drops the archived entries, preserving order.
func ActiveRouteEntrys(xs []RouteEntry) []RouteEntry {
	out := make([]RouteEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteEntrys adds up the retained of every active entry.
func SumRouteEntrys(xs []RouteEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRouteEntrys keys the entries by ID, last write winning.
func IndexRouteEntrys(xs []RouteEntry) map[string]RouteEntry {
	out := make(map[string]RouteEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteEntry builds a valid RouteEntry, for tests and for the seed data the cron
// binary writes.
func SampleRouteEntry(id string, at time.Time) RouteEntry {
	return RouteEntry{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by route.SampleRouteEntry",
		Retained:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RouteItemSchema is the stored shape of a RouteItem.
var RouteItemSchema = schema.New("route.routeitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteItem is one route item.
type RouteItem struct {
	ID       string
	Owner    string
	Source   string
	Name     string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteItem) Key() string { return "route/routeitem/" + x.ID }

// Fields encodes x for storage.
func (x RouteItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"source":   x.Source,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteItem decodes a stored record.
func ParseRouteItem(f codec.Fields) (RouteItem, error) {
	var x RouteItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routeitem: missing id")
	}
	x.Owner = f["owner"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("route: routeitem %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("route: routeitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routeitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routeitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteItem) String() string {
	return fmt.Sprintf("RouteItem(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteItem) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteItem) Normalize() RouteItem {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortRouteItems orders by attempts descending, then by ID.
func SortRouteItems(xs []RouteItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteItems drops the archived entries, preserving order.
func ActiveRouteItems(xs []RouteItem) []RouteItem {
	out := make([]RouteItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteItems adds up the attempts of every active entry.
func SumRouteItems(xs []RouteItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexRouteItems keys the entries by ID, last write winning.
func IndexRouteItems(xs []RouteItem) map[string]RouteItem {
	out := make(map[string]RouteItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteItem builds a valid RouteItem, for tests and for the seed data the cron
// binary writes.
func SampleRouteItem(id string, at time.Time) RouteItem {
	return RouteItem{
		ID:       id,
		Owner:    "owner-" + id,
		Source:   "source-" + id,
		Name:     "seeded by route.SampleRouteItem",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteDraftSchema is the stored shape of a RouteDraft.
var RouteDraftSchema = schema.New("route.routedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteDraft is one route draft.
type RouteDraft struct {
	ID        string
	Reference string
	Owner     string
	Category  string
	Attempts  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RouteDraft) Key() string { return "route/routedraft/" + x.ID }

// Fields encodes x for storage.
func (x RouteDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"owner":     x.Owner,
		"category":  x.Category,
		"attempts":  strconv.Itoa(x.Attempts),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteDraft decodes a stored record.
func ParseRouteDraft(f codec.Fields) (RouteDraft, error) {
	var x RouteDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routedraft: missing id")
	}
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("route: routedraft %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteDraft) String() string {
	return fmt.Sprintf("RouteDraft(%s, reference=%q, attempts=%d, archived=%t)", x.ID, x.Reference, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteDraft) Normalize() RouteDraft {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRouteDrafts orders by attempts descending, then by ID.
func SortRouteDrafts(xs []RouteDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteDrafts drops the archived entries, preserving order.
func ActiveRouteDrafts(xs []RouteDraft) []RouteDraft {
	out := make([]RouteDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteDrafts adds up the attempts of every active entry.
func SumRouteDrafts(xs []RouteDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexRouteDrafts keys the entries by ID, last write winning.
func IndexRouteDrafts(xs []RouteDraft) map[string]RouteDraft {
	out := make(map[string]RouteDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteDraft builds a valid RouteDraft, for tests and for the seed data the cron
// binary writes.
func SampleRouteDraft(id string, at time.Time) RouteDraft {
	return RouteDraft{
		ID:        id,
		Reference: "reference-" + id,
		Owner:     "owner-" + id,
		Category:  "seeded by route.SampleRouteDraft",
		Attempts:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RouteRevisionSchema is the stored shape of a RouteRevision.
var RouteRevisionSchema = schema.New("route.routerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteRevision is one route revision.
type RouteRevision struct {
	ID        string
	Reference string
	Group     string
	Owner     string
	Count     int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RouteRevision) Key() string { return "route/routerevision/" + x.ID }

// Fields encodes x for storage.
func (x RouteRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"owner":     x.Owner,
		"count":     strconv.Itoa(x.Count),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteRevision decodes a stored record.
func ParseRouteRevision(f codec.Fields) (RouteRevision, error) {
	var x RouteRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routerevision: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("route: routerevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("route: routerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteRevision) String() string {
	return fmt.Sprintf("RouteRevision(%s, reference=%q, count=%d, archived=%t)", x.ID, x.Reference, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteRevision) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteRevision) Normalize() RouteRevision {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRouteRevisions orders by count descending, then by ID.
func SortRouteRevisions(xs []RouteRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteRevisions drops the archived entries, preserving order.
func ActiveRouteRevisions(xs []RouteRevision) []RouteRevision {
	out := make([]RouteRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteRevisions adds up the count of every active entry.
func SumRouteRevisions(xs []RouteRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRouteRevisions keys the entries by ID, last write winning.
func IndexRouteRevisions(xs []RouteRevision) map[string]RouteRevision {
	out := make(map[string]RouteRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteRevision builds a valid RouteRevision, for tests and for the seed data the cron
// binary writes.
func SampleRouteRevision(id string, at time.Time) RouteRevision {
	return RouteRevision{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Owner:     "seeded by route.SampleRouteRevision",
		Count:     len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RouteSnapshotSchema is the stored shape of a RouteSnapshot.
var RouteSnapshotSchema = schema.New("route.routesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteSnapshot is one route snapshot.
type RouteSnapshot struct {
	ID       string
	Category string
	Owner    string
	Source   string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteSnapshot) Key() string { return "route/routesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x RouteSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"source":   x.Source,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteSnapshot decodes a stored record.
func ParseRouteSnapshot(f codec.Fields) (RouteSnapshot, error) {
	var x RouteSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routesnapshot: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("route: routesnapshot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("route: routesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteSnapshot) String() string {
	return fmt.Sprintf("RouteSnapshot(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteSnapshot) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteSnapshot) Normalize() RouteSnapshot {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRouteSnapshots orders by count descending, then by ID.
func SortRouteSnapshots(xs []RouteSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteSnapshots drops the archived entries, preserving order.
func ActiveRouteSnapshots(xs []RouteSnapshot) []RouteSnapshot {
	out := make([]RouteSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteSnapshots adds up the count of every active entry.
func SumRouteSnapshots(xs []RouteSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRouteSnapshots keys the entries by ID, last write winning.
func IndexRouteSnapshots(xs []RouteSnapshot) map[string]RouteSnapshot {
	out := make(map[string]RouteSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteSnapshot builds a valid RouteSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleRouteSnapshot(id string, at time.Time) RouteSnapshot {
	return RouteSnapshot{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by route.SampleRouteSnapshot",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteBatchSchema is the stored shape of a RouteBatch.
var RouteBatchSchema = schema.New("route.routebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteBatch is one route batch.
type RouteBatch struct {
	ID       string
	Group    string
	Comment  string
	Category string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteBatch) Key() string { return "route/routebatch/" + x.ID }

// Fields encodes x for storage.
func (x RouteBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"comment":  x.Comment,
		"category": x.Category,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteBatch decodes a stored record.
func ParseRouteBatch(f codec.Fields) (RouteBatch, error) {
	var x RouteBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routebatch: missing id")
	}
	x.Group = f["group"]
	x.Comment = f["comment"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("route: routebatch %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("route: routebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteBatch) String() string {
	return fmt.Sprintf("RouteBatch(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteBatch) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteBatch) Normalize() RouteBatch {
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRouteBatchs orders by offset descending, then by ID.
func SortRouteBatchs(xs []RouteBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteBatchs drops the archived entries, preserving order.
func ActiveRouteBatchs(xs []RouteBatch) []RouteBatch {
	out := make([]RouteBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteBatchs adds up the offset of every active entry.
func SumRouteBatchs(xs []RouteBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRouteBatchs keys the entries by ID, last write winning.
func IndexRouteBatchs(xs []RouteBatch) map[string]RouteBatch {
	out := make(map[string]RouteBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteBatch builds a valid RouteBatch, for tests and for the seed data the cron
// binary writes.
func SampleRouteBatch(id string, at time.Time) RouteBatch {
	return RouteBatch{
		ID:       id,
		Group:    "group-" + id,
		Comment:  "comment-" + id,
		Category: "seeded by route.SampleRouteBatch",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteRequestSchema is the stored shape of a RouteRequest.
var RouteRequestSchema = schema.New("route.routerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteRequest is one route request.
type RouteRequest struct {
	ID        string
	Name      string
	Reference string
	Owner     string
	Retained  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RouteRequest) Key() string { return "route/routerequest/" + x.ID }

// Fields encodes x for storage.
func (x RouteRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"owner":     x.Owner,
		"retained":  strconv.Itoa(x.Retained),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteRequest decodes a stored record.
func ParseRouteRequest(f codec.Fields) (RouteRequest, error) {
	var x RouteRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routerequest: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routerequest %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("route: routerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteRequest) String() string {
	return fmt.Sprintf("RouteRequest(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteRequest) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteRequest) Normalize() RouteRequest {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRouteRequests orders by retained descending, then by ID.
func SortRouteRequests(xs []RouteRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteRequests drops the archived entries, preserving order.
func ActiveRouteRequests(xs []RouteRequest) []RouteRequest {
	out := make([]RouteRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteRequests adds up the retained of every active entry.
func SumRouteRequests(xs []RouteRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRouteRequests keys the entries by ID, last write winning.
func IndexRouteRequests(xs []RouteRequest) map[string]RouteRequest {
	out := make(map[string]RouteRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteRequest builds a valid RouteRequest, for tests and for the seed data the cron
// binary writes.
func SampleRouteRequest(id string, at time.Time) RouteRequest {
	return RouteRequest{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by route.SampleRouteRequest",
		Retained:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RouteSummarySchema is the stored shape of a RouteSummary.
var RouteSummarySchema = schema.New("route.routesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteSummary is one route summary.
type RouteSummary struct {
	ID       string
	Category string
	Channel  string
	Region   string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteSummary) Key() string { return "route/routesummary/" + x.ID }

// Fields encodes x for storage.
func (x RouteSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteSummary decodes a stored record.
func ParseRouteSummary(f codec.Fields) (RouteSummary, error) {
	var x RouteSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routesummary: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("route: routesummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("route: routesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteSummary) String() string {
	return fmt.Sprintf("RouteSummary(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteSummary) Normalize() RouteSummary {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRouteSummarys orders by capacity descending, then by ID.
func SortRouteSummarys(xs []RouteSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteSummarys drops the archived entries, preserving order.
func ActiveRouteSummarys(xs []RouteSummary) []RouteSummary {
	out := make([]RouteSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteSummarys adds up the capacity of every active entry.
func SumRouteSummarys(xs []RouteSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexRouteSummarys keys the entries by ID, last write winning.
func IndexRouteSummarys(xs []RouteSummary) map[string]RouteSummary {
	out := make(map[string]RouteSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteSummary builds a valid RouteSummary, for tests and for the seed data the cron
// binary writes.
func SampleRouteSummary(id string, at time.Time) RouteSummary {
	return RouteSummary{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by route.SampleRouteSummary",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RoutePolicySchema is the stored shape of a RoutePolicy.
var RoutePolicySchema = schema.New("route.routepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RoutePolicy is one route policy.
type RoutePolicy struct {
	ID       string
	Channel  string
	Group    string
	Source   string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RoutePolicy) Key() string { return "route/routepolicy/" + x.ID }

// Fields encodes x for storage.
func (x RoutePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"group":    x.Group,
		"source":   x.Source,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRoutePolicy decodes a stored record.
func ParseRoutePolicy(f codec.Fields) (RoutePolicy, error) {
	var x RoutePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routepolicy: missing id")
	}
	x.Channel = f["channel"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("route: routepolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("route: routepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RoutePolicy) String() string {
	return fmt.Sprintf("RoutePolicy(%s, channel=%q, version=%d, archived=%t)", x.ID, x.Channel, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RoutePolicy) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RoutePolicy) Normalize() RoutePolicy {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortRoutePolicys orders by version descending, then by ID.
func SortRoutePolicys(xs []RoutePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRoutePolicys drops the archived entries, preserving order.
func ActiveRoutePolicys(xs []RoutePolicy) []RoutePolicy {
	out := make([]RoutePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRoutePolicys adds up the version of every active entry.
func SumRoutePolicys(xs []RoutePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexRoutePolicys keys the entries by ID, last write winning.
func IndexRoutePolicys(xs []RoutePolicy) map[string]RoutePolicy {
	out := make(map[string]RoutePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRoutePolicy builds a valid RoutePolicy, for tests and for the seed data the cron
// binary writes.
func SampleRoutePolicy(id string, at time.Time) RoutePolicy {
	return RoutePolicy{
		ID:      id,
		Channel: "channel-" + id,
		Group:   "group-" + id,
		Source:  "seeded by route.SampleRoutePolicy",
		Version: len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// RouteLinkSchema is the stored shape of a RouteLink.
var RouteLinkSchema = schema.New("route.routelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteLink is one route link.
type RouteLink struct {
	ID       string
	Region   string
	Comment  string
	Source   string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteLink) Key() string { return "route/routelink/" + x.ID }

// Fields encodes x for storage.
func (x RouteLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteLink decodes a stored record.
func ParseRouteLink(f codec.Fields) (RouteLink, error) {
	var x RouteLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routelink: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("route: routelink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("route: routelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteLink) String() string {
	return fmt.Sprintf("RouteLink(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteLink) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteLink) Normalize() RouteLink {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortRouteLinks orders by offset descending, then by ID.
func SortRouteLinks(xs []RouteLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteLinks drops the archived entries, preserving order.
func ActiveRouteLinks(xs []RouteLink) []RouteLink {
	out := make([]RouteLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteLinks adds up the offset of every active entry.
func SumRouteLinks(xs []RouteLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRouteLinks keys the entries by ID, last write winning.
func IndexRouteLinks(xs []RouteLink) map[string]RouteLink {
	out := make(map[string]RouteLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteLink builds a valid RouteLink, for tests and for the seed data the cron
// binary writes.
func SampleRouteLink(id string, at time.Time) RouteLink {
	return RouteLink{
		ID:      id,
		Region:  "region-" + id,
		Comment: "comment-" + id,
		Source:  "seeded by route.SampleRouteLink",
		Offset:  len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// RouteNoteSchema is the stored shape of a RouteNote.
var RouteNoteSchema = schema.New("route.routenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteNote is one route note.
type RouteNote struct {
	ID       string
	Channel  string
	Owner    string
	Source   string
	Priority int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteNote) Key() string { return "route/routenote/" + x.ID }

// Fields encodes x for storage.
func (x RouteNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteNote decodes a stored record.
func ParseRouteNote(f codec.Fields) (RouteNote, error) {
	var x RouteNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routenote: missing id")
	}
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("route: routenote %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("route: routenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteNote) String() string {
	return fmt.Sprintf("RouteNote(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteNote) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteNote) Normalize() RouteNote {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRouteNotes orders by priority descending, then by ID.
func SortRouteNotes(xs []RouteNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteNotes drops the archived entries, preserving order.
func ActiveRouteNotes(xs []RouteNote) []RouteNote {
	out := make([]RouteNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteNotes adds up the priority of every active entry.
func SumRouteNotes(xs []RouteNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRouteNotes keys the entries by ID, last write winning.
func IndexRouteNotes(xs []RouteNote) map[string]RouteNote {
	out := make(map[string]RouteNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteNote builds a valid RouteNote, for tests and for the seed data the cron
// binary writes.
func SampleRouteNote(id string, at time.Time) RouteNote {
	return RouteNote{
		ID:       id,
		Channel:  "channel-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by route.SampleRouteNote",
		Priority: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteWindowSchema is the stored shape of a RouteWindow.
var RouteWindowSchema = schema.New("route.routewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteWindow is one route window.
type RouteWindow struct {
	ID       string
	Channel  string
	Comment  string
	Group    string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteWindow) Key() string { return "route/routewindow/" + x.ID }

// Fields encodes x for storage.
func (x RouteWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteWindow decodes a stored record.
func ParseRouteWindow(f codec.Fields) (RouteWindow, error) {
	var x RouteWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routewindow: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("route: routewindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("route: routewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteWindow) String() string {
	return fmt.Sprintf("RouteWindow(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteWindow) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteWindow) Normalize() RouteWindow {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortRouteWindows orders by count descending, then by ID.
func SortRouteWindows(xs []RouteWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteWindows drops the archived entries, preserving order.
func ActiveRouteWindows(xs []RouteWindow) []RouteWindow {
	out := make([]RouteWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteWindows adds up the count of every active entry.
func SumRouteWindows(xs []RouteWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexRouteWindows keys the entries by ID, last write winning.
func IndexRouteWindows(xs []RouteWindow) map[string]RouteWindow {
	out := make(map[string]RouteWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteWindow builds a valid RouteWindow, for tests and for the seed data the cron
// binary writes.
func SampleRouteWindow(id string, at time.Time) RouteWindow {
	return RouteWindow{
		ID:      id,
		Channel: "channel-" + id,
		Comment: "comment-" + id,
		Group:   "seeded by route.SampleRouteWindow",
		Count:   len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// RouteQuotaSchema is the stored shape of a RouteQuota.
var RouteQuotaSchema = schema.New("route.routequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteQuota is one route quota.
type RouteQuota struct {
	ID       string
	Category string
	Name     string
	Label    string
	Priority int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteQuota) Key() string { return "route/routequota/" + x.ID }

// Fields encodes x for storage.
func (x RouteQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"label":    x.Label,
		"priority": strconv.Itoa(x.Priority),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteQuota decodes a stored record.
func ParseRouteQuota(f codec.Fields) (RouteQuota, error) {
	var x RouteQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routequota: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Label = f["label"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("route: routequota %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("route: routequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteQuota) String() string {
	return fmt.Sprintf("RouteQuota(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteQuota) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteQuota) Normalize() RouteQuota {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortRouteQuotas orders by priority descending, then by ID.
func SortRouteQuotas(xs []RouteQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteQuotas drops the archived entries, preserving order.
func ActiveRouteQuotas(xs []RouteQuota) []RouteQuota {
	out := make([]RouteQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteQuotas adds up the priority of every active entry.
func SumRouteQuotas(xs []RouteQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRouteQuotas keys the entries by ID, last write winning.
func IndexRouteQuotas(xs []RouteQuota) map[string]RouteQuota {
	out := make(map[string]RouteQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteQuota builds a valid RouteQuota, for tests and for the seed data the cron
// binary writes.
func SampleRouteQuota(id string, at time.Time) RouteQuota {
	return RouteQuota{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Label:    "seeded by route.SampleRouteQuota",
		Priority: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteBucketSchema is the stored shape of a RouteBucket.
var RouteBucketSchema = schema.New("route.routebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteBucket is one route bucket.
type RouteBucket struct {
	ID       string
	Source   string
	Owner    string
	Category string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteBucket) Key() string { return "route/routebucket/" + x.ID }

// Fields encodes x for storage.
func (x RouteBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteBucket decodes a stored record.
func ParseRouteBucket(f codec.Fields) (RouteBucket, error) {
	var x RouteBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routebucket: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("route: routebucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("route: routebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteBucket) String() string {
	return fmt.Sprintf("RouteBucket(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteBucket) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteBucket) Normalize() RouteBucket {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRouteBuckets orders by attempts descending, then by ID.
func SortRouteBuckets(xs []RouteBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteBuckets drops the archived entries, preserving order.
func ActiveRouteBuckets(xs []RouteBucket) []RouteBucket {
	out := make([]RouteBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteBuckets adds up the attempts of every active entry.
func SumRouteBuckets(xs []RouteBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexRouteBuckets keys the entries by ID, last write winning.
func IndexRouteBuckets(xs []RouteBucket) map[string]RouteBucket {
	out := make(map[string]RouteBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteBucket builds a valid RouteBucket, for tests and for the seed data the cron
// binary writes.
func SampleRouteBucket(id string, at time.Time) RouteBucket {
	return RouteBucket{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Category: "seeded by route.SampleRouteBucket",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteLaneSchema is the stored shape of a RouteLane.
var RouteLaneSchema = schema.New("route.routelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteLane is one route lane.
type RouteLane struct {
	ID       string
	Region   string
	Source   string
	Name     string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteLane) Key() string { return "route/routelane/" + x.ID }

// Fields encodes x for storage.
func (x RouteLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"source":   x.Source,
		"name":     x.Name,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteLane decodes a stored record.
func ParseRouteLane(f codec.Fields) (RouteLane, error) {
	var x RouteLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routelane: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("route: routelane %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("route: routelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteLane) String() string {
	return fmt.Sprintf("RouteLane(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteLane) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteLane) Normalize() RouteLane {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortRouteLanes orders by weight descending, then by ID.
func SortRouteLanes(xs []RouteLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteLanes drops the archived entries, preserving order.
func ActiveRouteLanes(xs []RouteLane) []RouteLane {
	out := make([]RouteLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteLanes adds up the weight of every active entry.
func SumRouteLanes(xs []RouteLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRouteLanes keys the entries by ID, last write winning.
func IndexRouteLanes(xs []RouteLane) map[string]RouteLane {
	out := make(map[string]RouteLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteLane builds a valid RouteLane, for tests and for the seed data the cron
// binary writes.
func SampleRouteLane(id string, at time.Time) RouteLane {
	return RouteLane{
		ID:       id,
		Region:   "region-" + id,
		Source:   "source-" + id,
		Name:     "seeded by route.SampleRouteLane",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteSlotSchema is the stored shape of a RouteSlot.
var RouteSlotSchema = schema.New("route.routeslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteSlot is one route slot.
type RouteSlot struct {
	ID       string
	Category string
	Group    string
	Source   string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteSlot) Key() string { return "route/routeslot/" + x.ID }

// Fields encodes x for storage.
func (x RouteSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"group":    x.Group,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteSlot decodes a stored record.
func ParseRouteSlot(f codec.Fields) (RouteSlot, error) {
	var x RouteSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routeslot: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("route: routeslot %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routeslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routeslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routeslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteSlot) String() string {
	return fmt.Sprintf("RouteSlot(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteSlot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteSlot) Normalize() RouteSlot {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRouteSlots orders by offset descending, then by ID.
func SortRouteSlots(xs []RouteSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteSlots drops the archived entries, preserving order.
func ActiveRouteSlots(xs []RouteSlot) []RouteSlot {
	out := make([]RouteSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteSlots adds up the offset of every active entry.
func SumRouteSlots(xs []RouteSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexRouteSlots keys the entries by ID, last write winning.
func IndexRouteSlots(xs []RouteSlot) map[string]RouteSlot {
	out := make(map[string]RouteSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteSlot builds a valid RouteSlot, for tests and for the seed data the cron
// binary writes.
func SampleRouteSlot(id string, at time.Time) RouteSlot {
	return RouteSlot{
		ID:       id,
		Category: "category-" + id,
		Group:    "group-" + id,
		Source:   "seeded by route.SampleRouteSlot",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteTagSchema is the stored shape of a RouteTag.
var RouteTagSchema = schema.New("route.routetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteTag is one route tag.
type RouteTag struct {
	ID        string
	Group     string
	Owner     string
	Reference string
	Retained  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x RouteTag) Key() string { return "route/routetag/" + x.ID }

// Fields encodes x for storage.
func (x RouteTag) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"owner":     x.Owner,
		"reference": x.Reference,
		"retained":  strconv.Itoa(x.Retained),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteTag decodes a stored record.
func ParseRouteTag(f codec.Fields) (RouteTag, error) {
	var x RouteTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routetag: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routetag %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("route: routetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteTag) String() string {
	return fmt.Sprintf("RouteTag(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteTag) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteTag) Normalize() RouteTag {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortRouteTags orders by retained descending, then by ID.
func SortRouteTags(xs []RouteTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteTags drops the archived entries, preserving order.
func ActiveRouteTags(xs []RouteTag) []RouteTag {
	out := make([]RouteTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteTags adds up the retained of every active entry.
func SumRouteTags(xs []RouteTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRouteTags keys the entries by ID, last write winning.
func IndexRouteTags(xs []RouteTag) map[string]RouteTag {
	out := make(map[string]RouteTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteTag builds a valid RouteTag, for tests and for the seed data the cron
// binary writes.
func SampleRouteTag(id string, at time.Time) RouteTag {
	return RouteTag{
		ID:        id,
		Group:     "group-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by route.SampleRouteTag",
		Retained:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// RouteRuleSchema is the stored shape of a RouteRule.
var RouteRuleSchema = schema.New("route.routerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteRule is one route rule.
type RouteRule struct {
	ID       string
	Source   string
	Comment  string
	Group    string
	Priority int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteRule) Key() string { return "route/routerule/" + x.ID }

// Fields encodes x for storage.
func (x RouteRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteRule decodes a stored record.
func ParseRouteRule(f codec.Fields) (RouteRule, error) {
	var x RouteRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routerule: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("route: routerule %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("route: routerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteRule) String() string {
	return fmt.Sprintf("RouteRule(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteRule) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteRule) Normalize() RouteRule {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortRouteRules orders by priority descending, then by ID.
func SortRouteRules(xs []RouteRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteRules drops the archived entries, preserving order.
func ActiveRouteRules(xs []RouteRule) []RouteRule {
	out := make([]RouteRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteRules adds up the priority of every active entry.
func SumRouteRules(xs []RouteRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRouteRules keys the entries by ID, last write winning.
func IndexRouteRules(xs []RouteRule) map[string]RouteRule {
	out := make(map[string]RouteRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteRule builds a valid RouteRule, for tests and for the seed data the cron
// binary writes.
func SampleRouteRule(id string, at time.Time) RouteRule {
	return RouteRule{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by route.SampleRouteRule",
		Priority: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteHoldSchema is the stored shape of a RouteHold.
var RouteHoldSchema = schema.New("route.routehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteHold is one route hold.
type RouteHold struct {
	ID       string
	Name     string
	Comment  string
	Region   string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteHold) Key() string { return "route/routehold/" + x.ID }

// Fields encodes x for storage.
func (x RouteHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"comment":  x.Comment,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteHold decodes a stored record.
func ParseRouteHold(f codec.Fields) (RouteHold, error) {
	var x RouteHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routehold: missing id")
	}
	x.Name = f["name"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routehold %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("route: routehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteHold) String() string {
	return fmt.Sprintf("RouteHold(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteHold) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteHold) Normalize() RouteHold {
	x.Name = strings.TrimSpace(x.Name)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortRouteHolds orders by retained descending, then by ID.
func SortRouteHolds(xs []RouteHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteHolds drops the archived entries, preserving order.
func ActiveRouteHolds(xs []RouteHold) []RouteHold {
	out := make([]RouteHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteHolds adds up the retained of every active entry.
func SumRouteHolds(xs []RouteHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexRouteHolds keys the entries by ID, last write winning.
func IndexRouteHolds(xs []RouteHold) map[string]RouteHold {
	out := make(map[string]RouteHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteHold builds a valid RouteHold, for tests and for the seed data the cron
// binary writes.
func SampleRouteHold(id string, at time.Time) RouteHold {
	return RouteHold{
		ID:       id,
		Name:     "name-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by route.SampleRouteHold",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteClaimSchema is the stored shape of a RouteClaim.
var RouteClaimSchema = schema.New("route.routeclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteClaim is one route claim.
type RouteClaim struct {
	ID       string
	Category string
	Name     string
	Comment  string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteClaim) Key() string { return "route/routeclaim/" + x.ID }

// Fields encodes x for storage.
func (x RouteClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"name":     x.Name,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteClaim decodes a stored record.
func ParseRouteClaim(f codec.Fields) (RouteClaim, error) {
	var x RouteClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routeclaim: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("route: routeclaim %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routeclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routeclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routeclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteClaim) String() string {
	return fmt.Sprintf("RouteClaim(%s, category=%q, priority=%d, archived=%t)", x.ID, x.Category, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteClaim) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteClaim) Normalize() RouteClaim {
	x.Category = strings.TrimSpace(x.Category)
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

// SortRouteClaims orders by priority descending, then by ID.
func SortRouteClaims(xs []RouteClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteClaims drops the archived entries, preserving order.
func ActiveRouteClaims(xs []RouteClaim) []RouteClaim {
	out := make([]RouteClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteClaims adds up the priority of every active entry.
func SumRouteClaims(xs []RouteClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexRouteClaims keys the entries by ID, last write winning.
func IndexRouteClaims(xs []RouteClaim) map[string]RouteClaim {
	out := make(map[string]RouteClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteClaim builds a valid RouteClaim, for tests and for the seed data the cron
// binary writes.
func SampleRouteClaim(id string, at time.Time) RouteClaim {
	return RouteClaim{
		ID:       id,
		Category: "category-" + id,
		Name:     "name-" + id,
		Comment:  "seeded by route.SampleRouteClaim",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// RouteCursorSchema is the stored shape of a RouteCursor.
var RouteCursorSchema = schema.New("route.routecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// RouteCursor is one route cursor.
type RouteCursor struct {
	ID       string
	Comment  string
	Label    string
	Source   string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x RouteCursor) Key() string { return "route/routecursor/" + x.ID }

// Fields encodes x for storage.
func (x RouteCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"label":    x.Label,
		"source":   x.Source,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseRouteCursor decodes a stored record.
func ParseRouteCursor(f codec.Fields) (RouteCursor, error) {
	var x RouteCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("route: routecursor: missing id")
	}
	x.Comment = f["comment"]
	x.Label = f["label"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("route: routecursor %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("route: routecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("route: routecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("route: routecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x RouteCursor) String() string {
	return fmt.Sprintf("RouteCursor(%s, comment=%q, weight=%d, archived=%t)", x.ID, x.Comment, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x RouteCursor) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x RouteCursor) Normalize() RouteCursor {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortRouteCursors orders by weight descending, then by ID.
func SortRouteCursors(xs []RouteCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveRouteCursors drops the archived entries, preserving order.
func ActiveRouteCursors(xs []RouteCursor) []RouteCursor {
	out := make([]RouteCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumRouteCursors adds up the weight of every active entry.
func SumRouteCursors(xs []RouteCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexRouteCursors keys the entries by ID, last write winning.
func IndexRouteCursors(xs []RouteCursor) map[string]RouteCursor {
	out := make(map[string]RouteCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleRouteCursor builds a valid RouteCursor, for tests and for the seed data the cron
// binary writes.
func SampleRouteCursor(id string, at time.Time) RouteCursor {
	return RouteCursor{
		ID:       id,
		Comment:  "comment-" + id,
		Label:    "label-" + id,
		Source:   "seeded by route.SampleRouteCursor",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	RouteRecordSchema,
	RouteEntrySchema,
	RouteItemSchema,
	RouteDraftSchema,
	RouteRevisionSchema,
	RouteSnapshotSchema,
	RouteBatchSchema,
	RouteRequestSchema,
	RouteSummarySchema,
	RoutePolicySchema,
	RouteLinkSchema,
	RouteNoteSchema,
	RouteWindowSchema,
	RouteQuotaSchema,
	RouteBucketSchema,
	RouteLaneSchema,
	RouteSlotSchema,
	RouteTagSchema,
	RouteRuleSchema,
	RouteHoldSchema,
	RouteClaimSchema,
	RouteCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"route/routerecord/",
	"route/routeentry/",
	"route/routeitem/",
	"route/routedraft/",
	"route/routerevision/",
	"route/routesnapshot/",
	"route/routebatch/",
	"route/routerequest/",
	"route/routesummary/",
	"route/routepolicy/",
	"route/routelink/",
	"route/routenote/",
	"route/routewindow/",
	"route/routequota/",
	"route/routebucket/",
	"route/routelane/",
	"route/routeslot/",
	"route/routetag/",
	"route/routerule/",
	"route/routehold/",
	"route/routeclaim/",
	"route/routecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "route: " + strings.Join(names, ", ")
}
