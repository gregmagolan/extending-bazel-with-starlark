// Package geo holds the geo domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package geo

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// GeoRecordSchema is the stored shape of a GeoRecord.
var GeoRecordSchema = schema.New("geo.georecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoRecord is one geo record.
type GeoRecord struct {
	ID        string
	Region    string
	Source    string
	Reference string
	Capacity  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x GeoRecord) Key() string { return "geo/georecord/" + x.ID }

// Fields encodes x for storage.
func (x GeoRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"source":    x.Source,
		"reference": x.Reference,
		"capacity":  strconv.Itoa(x.Capacity),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoRecord decodes a stored record.
func ParseGeoRecord(f codec.Fields) (GeoRecord, error) {
	var x GeoRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: georecord: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Reference = f["reference"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: georecord %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: georecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: georecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: georecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoRecord) String() string {
	return fmt.Sprintf("GeoRecord(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoRecord) Normalize() GeoRecord {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortGeoRecords orders by capacity descending, then by ID.
func SortGeoRecords(xs []GeoRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoRecords drops the archived entries, preserving order.
func ActiveGeoRecords(xs []GeoRecord) []GeoRecord {
	out := make([]GeoRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoRecords adds up the capacity of every active entry.
func SumGeoRecords(xs []GeoRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexGeoRecords keys the entries by ID, last write winning.
func IndexGeoRecords(xs []GeoRecord) map[string]GeoRecord {
	out := make(map[string]GeoRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoRecord builds a valid GeoRecord, for tests and for the seed data the cron
// binary writes.
func SampleGeoRecord(id string, at time.Time) GeoRecord {
	return GeoRecord{
		ID:        id,
		Region:    "region-" + id,
		Source:    "source-" + id,
		Reference: "seeded by geo.SampleGeoRecord",
		Capacity:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// GeoEntrySchema is the stored shape of a GeoEntry.
var GeoEntrySchema = schema.New("geo.geoentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoEntry is one geo entry.
type GeoEntry struct {
	ID        string
	Category  string
	Name      string
	Reference string
	Capacity  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x GeoEntry) Key() string { return "geo/geoentry/" + x.ID }

// Fields encodes x for storage.
func (x GeoEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"name":      x.Name,
		"reference": x.Reference,
		"capacity":  strconv.Itoa(x.Capacity),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoEntry decodes a stored record.
func ParseGeoEntry(f codec.Fields) (GeoEntry, error) {
	var x GeoEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geoentry: missing id")
	}
	x.Category = f["category"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: geoentry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("geo: geoentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geoentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geoentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoEntry) String() string {
	return fmt.Sprintf("GeoEntry(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoEntry) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoEntry) Normalize() GeoEntry {
	x.Category = strings.TrimSpace(x.Category)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortGeoEntrys orders by capacity descending, then by ID.
func SortGeoEntrys(xs []GeoEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoEntrys drops the archived entries, preserving order.
func ActiveGeoEntrys(xs []GeoEntry) []GeoEntry {
	out := make([]GeoEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoEntrys adds up the capacity of every active entry.
func SumGeoEntrys(xs []GeoEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexGeoEntrys keys the entries by ID, last write winning.
func IndexGeoEntrys(xs []GeoEntry) map[string]GeoEntry {
	out := make(map[string]GeoEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoEntry builds a valid GeoEntry, for tests and for the seed data the cron
// binary writes.
func SampleGeoEntry(id string, at time.Time) GeoEntry {
	return GeoEntry{
		ID:        id,
		Category:  "category-" + id,
		Name:      "name-" + id,
		Reference: "seeded by geo.SampleGeoEntry",
		Capacity:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// GeoItemSchema is the stored shape of a GeoItem.
var GeoItemSchema = schema.New("geo.geoitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoItem is one geo item.
type GeoItem struct {
	ID       string
	Label    string
	Source   string
	Name     string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoItem) Key() string { return "geo/geoitem/" + x.ID }

// Fields encodes x for storage.
func (x GeoItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoItem decodes a stored record.
func ParseGeoItem(f codec.Fields) (GeoItem, error) {
	var x GeoItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geoitem: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("geo: geoitem %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("geo: geoitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geoitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geoitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoItem) String() string {
	return fmt.Sprintf("GeoItem(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoItem) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoItem) Normalize() GeoItem {
	x.Label = strings.TrimSpace(x.Label)
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

// SortGeoItems orders by retained descending, then by ID.
func SortGeoItems(xs []GeoItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoItems drops the archived entries, preserving order.
func ActiveGeoItems(xs []GeoItem) []GeoItem {
	out := make([]GeoItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoItems adds up the retained of every active entry.
func SumGeoItems(xs []GeoItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexGeoItems keys the entries by ID, last write winning.
func IndexGeoItems(xs []GeoItem) map[string]GeoItem {
	out := make(map[string]GeoItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoItem builds a valid GeoItem, for tests and for the seed data the cron
// binary writes.
func SampleGeoItem(id string, at time.Time) GeoItem {
	return GeoItem{
		ID:       id,
		Label:    "label-" + id,
		Source:   "source-" + id,
		Name:     "seeded by geo.SampleGeoItem",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoDraftSchema is the stored shape of a GeoDraft.
var GeoDraftSchema = schema.New("geo.geodraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoDraft is one geo draft.
type GeoDraft struct {
	ID       string
	Category string
	Label    string
	Comment  string
	Count    int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoDraft) Key() string { return "geo/geodraft/" + x.ID }

// Fields encodes x for storage.
func (x GeoDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoDraft decodes a stored record.
func ParseGeoDraft(f codec.Fields) (GeoDraft, error) {
	var x GeoDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geodraft: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("geo: geodraft %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: geodraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geodraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geodraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoDraft) String() string {
	return fmt.Sprintf("GeoDraft(%s, category=%q, count=%d, archived=%t)", x.ID, x.Category, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoDraft) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoDraft) Normalize() GeoDraft {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortGeoDrafts orders by count descending, then by ID.
func SortGeoDrafts(xs []GeoDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoDrafts drops the archived entries, preserving order.
func ActiveGeoDrafts(xs []GeoDraft) []GeoDraft {
	out := make([]GeoDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoDrafts adds up the count of every active entry.
func SumGeoDrafts(xs []GeoDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexGeoDrafts keys the entries by ID, last write winning.
func IndexGeoDrafts(xs []GeoDraft) map[string]GeoDraft {
	out := make(map[string]GeoDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoDraft builds a valid GeoDraft, for tests and for the seed data the cron
// binary writes.
func SampleGeoDraft(id string, at time.Time) GeoDraft {
	return GeoDraft{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Comment:  "seeded by geo.SampleGeoDraft",
		Count:    len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoRevisionSchema is the stored shape of a GeoRevision.
var GeoRevisionSchema = schema.New("geo.georevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoRevision is one geo revision.
type GeoRevision struct {
	ID       string
	Group    string
	Owner    string
	Label    string
	Capacity int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoRevision) Key() string { return "geo/georevision/" + x.ID }

// Fields encodes x for storage.
func (x GeoRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"owner":    x.Owner,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoRevision decodes a stored record.
func ParseGeoRevision(f codec.Fields) (GeoRevision, error) {
	var x GeoRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: georevision: missing id")
	}
	x.Group = f["group"]
	x.Owner = f["owner"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: georevision %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: georevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: georevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: georevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoRevision) String() string {
	return fmt.Sprintf("GeoRevision(%s, group=%q, capacity=%d, archived=%t)", x.ID, x.Group, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoRevision) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoRevision) Normalize() GeoRevision {
	x.Group = strings.TrimSpace(x.Group)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortGeoRevisions orders by capacity descending, then by ID.
func SortGeoRevisions(xs []GeoRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoRevisions drops the archived entries, preserving order.
func ActiveGeoRevisions(xs []GeoRevision) []GeoRevision {
	out := make([]GeoRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoRevisions adds up the capacity of every active entry.
func SumGeoRevisions(xs []GeoRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexGeoRevisions keys the entries by ID, last write winning.
func IndexGeoRevisions(xs []GeoRevision) map[string]GeoRevision {
	out := make(map[string]GeoRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoRevision builds a valid GeoRevision, for tests and for the seed data the cron
// binary writes.
func SampleGeoRevision(id string, at time.Time) GeoRevision {
	return GeoRevision{
		ID:       id,
		Group:    "group-" + id,
		Owner:    "owner-" + id,
		Label:    "seeded by geo.SampleGeoRevision",
		Capacity: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoSnapshotSchema is the stored shape of a GeoSnapshot.
var GeoSnapshotSchema = schema.New("geo.geosnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoSnapshot is one geo snapshot.
type GeoSnapshot struct {
	ID       string
	Source   string
	Comment  string
	Label    string
	Weight   int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoSnapshot) Key() string { return "geo/geosnapshot/" + x.ID }

// Fields encodes x for storage.
func (x GeoSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoSnapshot decodes a stored record.
func ParseGeoSnapshot(f codec.Fields) (GeoSnapshot, error) {
	var x GeoSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geosnapshot: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("geo: geosnapshot %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("geo: geosnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geosnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geosnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoSnapshot) String() string {
	return fmt.Sprintf("GeoSnapshot(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoSnapshot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoSnapshot) Normalize() GeoSnapshot {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortGeoSnapshots orders by weight descending, then by ID.
func SortGeoSnapshots(xs []GeoSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoSnapshots drops the archived entries, preserving order.
func ActiveGeoSnapshots(xs []GeoSnapshot) []GeoSnapshot {
	out := make([]GeoSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoSnapshots adds up the weight of every active entry.
func SumGeoSnapshots(xs []GeoSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexGeoSnapshots keys the entries by ID, last write winning.
func IndexGeoSnapshots(xs []GeoSnapshot) map[string]GeoSnapshot {
	out := make(map[string]GeoSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoSnapshot builds a valid GeoSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleGeoSnapshot(id string, at time.Time) GeoSnapshot {
	return GeoSnapshot{
		ID:      id,
		Source:  "source-" + id,
		Comment: "comment-" + id,
		Label:   "seeded by geo.SampleGeoSnapshot",
		Weight:  len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// GeoBatchSchema is the stored shape of a GeoBatch.
var GeoBatchSchema = schema.New("geo.geobatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoBatch is one geo batch.
type GeoBatch struct {
	ID       string
	Name     string
	Source   string
	Owner    string
	Count    int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoBatch) Key() string { return "geo/geobatch/" + x.ID }

// Fields encodes x for storage.
func (x GeoBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"owner":    x.Owner,
		"count":    strconv.Itoa(x.Count),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoBatch decodes a stored record.
func ParseGeoBatch(f codec.Fields) (GeoBatch, error) {
	var x GeoBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geobatch: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Owner = f["owner"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("geo: geobatch %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("geo: geobatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geobatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geobatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoBatch) String() string {
	return fmt.Sprintf("GeoBatch(%s, name=%q, count=%d, archived=%t)", x.ID, x.Name, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoBatch) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoBatch) Normalize() GeoBatch {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortGeoBatchs orders by count descending, then by ID.
func SortGeoBatchs(xs []GeoBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoBatchs drops the archived entries, preserving order.
func ActiveGeoBatchs(xs []GeoBatch) []GeoBatch {
	out := make([]GeoBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoBatchs adds up the count of every active entry.
func SumGeoBatchs(xs []GeoBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexGeoBatchs keys the entries by ID, last write winning.
func IndexGeoBatchs(xs []GeoBatch) map[string]GeoBatch {
	out := make(map[string]GeoBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoBatch builds a valid GeoBatch, for tests and for the seed data the cron
// binary writes.
func SampleGeoBatch(id string, at time.Time) GeoBatch {
	return GeoBatch{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Owner:    "seeded by geo.SampleGeoBatch",
		Count:    len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoRequestSchema is the stored shape of a GeoRequest.
var GeoRequestSchema = schema.New("geo.georequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoRequest is one geo request.
type GeoRequest struct {
	ID        string
	Name      string
	Reference string
	Channel   string
	Retained  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x GeoRequest) Key() string { return "geo/georequest/" + x.ID }

// Fields encodes x for storage.
func (x GeoRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"channel":   x.Channel,
		"retained":  strconv.Itoa(x.Retained),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoRequest decodes a stored record.
func ParseGeoRequest(f codec.Fields) (GeoRequest, error) {
	var x GeoRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: georequest: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("geo: georequest %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: georequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: georequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: georequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoRequest) String() string {
	return fmt.Sprintf("GeoRequest(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoRequest) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoRequest) Normalize() GeoRequest {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortGeoRequests orders by retained descending, then by ID.
func SortGeoRequests(xs []GeoRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoRequests drops the archived entries, preserving order.
func ActiveGeoRequests(xs []GeoRequest) []GeoRequest {
	out := make([]GeoRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoRequests adds up the retained of every active entry.
func SumGeoRequests(xs []GeoRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexGeoRequests keys the entries by ID, last write winning.
func IndexGeoRequests(xs []GeoRequest) map[string]GeoRequest {
	out := make(map[string]GeoRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoRequest builds a valid GeoRequest, for tests and for the seed data the cron
// binary writes.
func SampleGeoRequest(id string, at time.Time) GeoRequest {
	return GeoRequest{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by geo.SampleGeoRequest",
		Retained:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// GeoSummarySchema is the stored shape of a GeoSummary.
var GeoSummarySchema = schema.New("geo.geosummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoSummary is one geo summary.
type GeoSummary struct {
	ID        string
	Category  string
	Reference string
	Region    string
	Retained  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x GeoSummary) Key() string { return "geo/geosummary/" + x.ID }

// Fields encodes x for storage.
func (x GeoSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"category":  x.Category,
		"reference": x.Reference,
		"region":    x.Region,
		"retained":  strconv.Itoa(x.Retained),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoSummary decodes a stored record.
func ParseGeoSummary(f codec.Fields) (GeoSummary, error) {
	var x GeoSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geosummary: missing id")
	}
	x.Category = f["category"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("geo: geosummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: geosummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geosummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geosummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoSummary) String() string {
	return fmt.Sprintf("GeoSummary(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoSummary) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoSummary) Normalize() GeoSummary {
	x.Category = strings.TrimSpace(x.Category)
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

// SortGeoSummarys orders by retained descending, then by ID.
func SortGeoSummarys(xs []GeoSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoSummarys drops the archived entries, preserving order.
func ActiveGeoSummarys(xs []GeoSummary) []GeoSummary {
	out := make([]GeoSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoSummarys adds up the retained of every active entry.
func SumGeoSummarys(xs []GeoSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexGeoSummarys keys the entries by ID, last write winning.
func IndexGeoSummarys(xs []GeoSummary) map[string]GeoSummary {
	out := make(map[string]GeoSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoSummary builds a valid GeoSummary, for tests and for the seed data the cron
// binary writes.
func SampleGeoSummary(id string, at time.Time) GeoSummary {
	return GeoSummary{
		ID:        id,
		Category:  "category-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by geo.SampleGeoSummary",
		Retained:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// GeoPolicySchema is the stored shape of a GeoPolicy.
var GeoPolicySchema = schema.New("geo.geopolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoPolicy is one geo policy.
type GeoPolicy struct {
	ID       string
	Region   string
	Name     string
	Owner    string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoPolicy) Key() string { return "geo/geopolicy/" + x.ID }

// Fields encodes x for storage.
func (x GeoPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"owner":    x.Owner,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoPolicy decodes a stored record.
func ParseGeoPolicy(f codec.Fields) (GeoPolicy, error) {
	var x GeoPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geopolicy: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Owner = f["owner"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: geopolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("geo: geopolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geopolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geopolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoPolicy) String() string {
	return fmt.Sprintf("GeoPolicy(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoPolicy) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoPolicy) Normalize() GeoPolicy {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortGeoPolicys orders by version descending, then by ID.
func SortGeoPolicys(xs []GeoPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoPolicys drops the archived entries, preserving order.
func ActiveGeoPolicys(xs []GeoPolicy) []GeoPolicy {
	out := make([]GeoPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoPolicys adds up the version of every active entry.
func SumGeoPolicys(xs []GeoPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexGeoPolicys keys the entries by ID, last write winning.
func IndexGeoPolicys(xs []GeoPolicy) map[string]GeoPolicy {
	out := make(map[string]GeoPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoPolicy builds a valid GeoPolicy, for tests and for the seed data the cron
// binary writes.
func SampleGeoPolicy(id string, at time.Time) GeoPolicy {
	return GeoPolicy{
		ID:      id,
		Region:  "region-" + id,
		Name:    "name-" + id,
		Owner:   "seeded by geo.SampleGeoPolicy",
		Version: len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// GeoLinkSchema is the stored shape of a GeoLink.
var GeoLinkSchema = schema.New("geo.geolink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoLink is one geo link.
type GeoLink struct {
	ID        string
	Reference string
	Region    string
	Label     string
	Offset    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x GeoLink) Key() string { return "geo/geolink/" + x.ID }

// Fields encodes x for storage.
func (x GeoLink) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"label":     x.Label,
		"offset":    strconv.Itoa(x.Offset),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoLink decodes a stored record.
func ParseGeoLink(f codec.Fields) (GeoLink, error) {
	var x GeoLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geolink: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("geo: geolink %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("geo: geolink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geolink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geolink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoLink) String() string {
	return fmt.Sprintf("GeoLink(%s, reference=%q, offset=%d, archived=%t)", x.ID, x.Reference, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoLink) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoLink) Normalize() GeoLink {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortGeoLinks orders by offset descending, then by ID.
func SortGeoLinks(xs []GeoLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoLinks drops the archived entries, preserving order.
func ActiveGeoLinks(xs []GeoLink) []GeoLink {
	out := make([]GeoLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoLinks adds up the offset of every active entry.
func SumGeoLinks(xs []GeoLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexGeoLinks keys the entries by ID, last write winning.
func IndexGeoLinks(xs []GeoLink) map[string]GeoLink {
	out := make(map[string]GeoLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoLink builds a valid GeoLink, for tests and for the seed data the cron
// binary writes.
func SampleGeoLink(id string, at time.Time) GeoLink {
	return GeoLink{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Label:     "seeded by geo.SampleGeoLink",
		Offset:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// GeoNoteSchema is the stored shape of a GeoNote.
var GeoNoteSchema = schema.New("geo.geonote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoNote is one geo note.
type GeoNote struct {
	ID        string
	Group     string
	Reference string
	Region    string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x GeoNote) Key() string { return "geo/geonote/" + x.ID }

// Fields encodes x for storage.
func (x GeoNote) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"region":    x.Region,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoNote decodes a stored record.
func ParseGeoNote(f codec.Fields) (GeoNote, error) {
	var x GeoNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geonote: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: geonote %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("geo: geonote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geonote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geonote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoNote) String() string {
	return fmt.Sprintf("GeoNote(%s, group=%q, version=%d, archived=%t)", x.ID, x.Group, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoNote) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoNote) Normalize() GeoNote {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortGeoNotes orders by version descending, then by ID.
func SortGeoNotes(xs []GeoNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoNotes drops the archived entries, preserving order.
func ActiveGeoNotes(xs []GeoNote) []GeoNote {
	out := make([]GeoNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoNotes adds up the version of every active entry.
func SumGeoNotes(xs []GeoNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexGeoNotes keys the entries by ID, last write winning.
func IndexGeoNotes(xs []GeoNote) map[string]GeoNote {
	out := make(map[string]GeoNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoNote builds a valid GeoNote, for tests and for the seed data the cron
// binary writes.
func SampleGeoNote(id string, at time.Time) GeoNote {
	return GeoNote{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Region:    "seeded by geo.SampleGeoNote",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// GeoWindowSchema is the stored shape of a GeoWindow.
var GeoWindowSchema = schema.New("geo.geowindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoWindow is one geo window.
type GeoWindow struct {
	ID       string
	Category string
	Channel  string
	Source   string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoWindow) Key() string { return "geo/geowindow/" + x.ID }

// Fields encodes x for storage.
func (x GeoWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoWindow decodes a stored record.
func ParseGeoWindow(f codec.Fields) (GeoWindow, error) {
	var x GeoWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geowindow: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: geowindow %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("geo: geowindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geowindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geowindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoWindow) String() string {
	return fmt.Sprintf("GeoWindow(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoWindow) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoWindow) Normalize() GeoWindow {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortGeoWindows orders by capacity descending, then by ID.
func SortGeoWindows(xs []GeoWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoWindows drops the archived entries, preserving order.
func ActiveGeoWindows(xs []GeoWindow) []GeoWindow {
	out := make([]GeoWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoWindows adds up the capacity of every active entry.
func SumGeoWindows(xs []GeoWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexGeoWindows keys the entries by ID, last write winning.
func IndexGeoWindows(xs []GeoWindow) map[string]GeoWindow {
	out := make(map[string]GeoWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoWindow builds a valid GeoWindow, for tests and for the seed data the cron
// binary writes.
func SampleGeoWindow(id string, at time.Time) GeoWindow {
	return GeoWindow{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Source:   "seeded by geo.SampleGeoWindow",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoQuotaSchema is the stored shape of a GeoQuota.
var GeoQuotaSchema = schema.New("geo.geoquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoQuota is one geo quota.
type GeoQuota struct {
	ID        string
	Reference string
	Label     string
	Region    string
	Version   int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x GeoQuota) Key() string { return "geo/geoquota/" + x.ID }

// Fields encodes x for storage.
func (x GeoQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"label":     x.Label,
		"region":    x.Region,
		"version":   strconv.Itoa(x.Version),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoQuota decodes a stored record.
func ParseGeoQuota(f codec.Fields) (GeoQuota, error) {
	var x GeoQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geoquota: missing id")
	}
	x.Reference = f["reference"]
	x.Label = f["label"]
	x.Region = f["region"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: geoquota %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("geo: geoquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geoquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geoquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoQuota) String() string {
	return fmt.Sprintf("GeoQuota(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoQuota) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoQuota) Normalize() GeoQuota {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortGeoQuotas orders by version descending, then by ID.
func SortGeoQuotas(xs []GeoQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoQuotas drops the archived entries, preserving order.
func ActiveGeoQuotas(xs []GeoQuota) []GeoQuota {
	out := make([]GeoQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoQuotas adds up the version of every active entry.
func SumGeoQuotas(xs []GeoQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexGeoQuotas keys the entries by ID, last write winning.
func IndexGeoQuotas(xs []GeoQuota) map[string]GeoQuota {
	out := make(map[string]GeoQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoQuota builds a valid GeoQuota, for tests and for the seed data the cron
// binary writes.
func SampleGeoQuota(id string, at time.Time) GeoQuota {
	return GeoQuota{
		ID:        id,
		Reference: "reference-" + id,
		Label:     "label-" + id,
		Region:    "seeded by geo.SampleGeoQuota",
		Version:   len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// GeoBucketSchema is the stored shape of a GeoBucket.
var GeoBucketSchema = schema.New("geo.geobucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoBucket is one geo bucket.
type GeoBucket struct {
	ID       string
	Channel  string
	Region   string
	Name     string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoBucket) Key() string { return "geo/geobucket/" + x.ID }

// Fields encodes x for storage.
func (x GeoBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"name":     x.Name,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoBucket decodes a stored record.
func ParseGeoBucket(f codec.Fields) (GeoBucket, error) {
	var x GeoBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geobucket: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Name = f["name"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("geo: geobucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: geobucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geobucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geobucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoBucket) String() string {
	return fmt.Sprintf("GeoBucket(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoBucket) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoBucket) Normalize() GeoBucket {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortGeoBuckets orders by count descending, then by ID.
func SortGeoBuckets(xs []GeoBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoBuckets drops the archived entries, preserving order.
func ActiveGeoBuckets(xs []GeoBucket) []GeoBucket {
	out := make([]GeoBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoBuckets adds up the count of every active entry.
func SumGeoBuckets(xs []GeoBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexGeoBuckets keys the entries by ID, last write winning.
func IndexGeoBuckets(xs []GeoBucket) map[string]GeoBucket {
	out := make(map[string]GeoBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoBucket builds a valid GeoBucket, for tests and for the seed data the cron
// binary writes.
func SampleGeoBucket(id string, at time.Time) GeoBucket {
	return GeoBucket{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Name:     "seeded by geo.SampleGeoBucket",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoLaneSchema is the stored shape of a GeoLane.
var GeoLaneSchema = schema.New("geo.geolane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoLane is one geo lane.
type GeoLane struct {
	ID       string
	Owner    string
	Channel  string
	Group    string
	Attempts int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoLane) Key() string { return "geo/geolane/" + x.ID }

// Fields encodes x for storage.
func (x GeoLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoLane decodes a stored record.
func ParseGeoLane(f codec.Fields) (GeoLane, error) {
	var x GeoLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geolane: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("geo: geolane %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: geolane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geolane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geolane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoLane) String() string {
	return fmt.Sprintf("GeoLane(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoLane) Normalize() GeoLane {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortGeoLanes orders by attempts descending, then by ID.
func SortGeoLanes(xs []GeoLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoLanes drops the archived entries, preserving order.
func ActiveGeoLanes(xs []GeoLane) []GeoLane {
	out := make([]GeoLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoLanes adds up the attempts of every active entry.
func SumGeoLanes(xs []GeoLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexGeoLanes keys the entries by ID, last write winning.
func IndexGeoLanes(xs []GeoLane) map[string]GeoLane {
	out := make(map[string]GeoLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoLane builds a valid GeoLane, for tests and for the seed data the cron
// binary writes.
func SampleGeoLane(id string, at time.Time) GeoLane {
	return GeoLane{
		ID:       id,
		Owner:    "owner-" + id,
		Channel:  "channel-" + id,
		Group:    "seeded by geo.SampleGeoLane",
		Attempts: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoSlotSchema is the stored shape of a GeoSlot.
var GeoSlotSchema = schema.New("geo.geoslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoSlot is one geo slot.
type GeoSlot struct {
	ID       string
	Comment  string
	Region   string
	Group    string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoSlot) Key() string { return "geo/geoslot/" + x.ID }

// Fields encodes x for storage.
func (x GeoSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"region":   x.Region,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoSlot decodes a stored record.
func ParseGeoSlot(f codec.Fields) (GeoSlot, error) {
	var x GeoSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geoslot: missing id")
	}
	x.Comment = f["comment"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("geo: geoslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: geoslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geoslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geoslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoSlot) String() string {
	return fmt.Sprintf("GeoSlot(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoSlot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoSlot) Normalize() GeoSlot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortGeoSlots orders by offset descending, then by ID.
func SortGeoSlots(xs []GeoSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoSlots drops the archived entries, preserving order.
func ActiveGeoSlots(xs []GeoSlot) []GeoSlot {
	out := make([]GeoSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoSlots adds up the offset of every active entry.
func SumGeoSlots(xs []GeoSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexGeoSlots keys the entries by ID, last write winning.
func IndexGeoSlots(xs []GeoSlot) map[string]GeoSlot {
	out := make(map[string]GeoSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoSlot builds a valid GeoSlot, for tests and for the seed data the cron
// binary writes.
func SampleGeoSlot(id string, at time.Time) GeoSlot {
	return GeoSlot{
		ID:      id,
		Comment: "comment-" + id,
		Region:  "region-" + id,
		Group:   "seeded by geo.SampleGeoSlot",
		Offset:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// GeoTagSchema is the stored shape of a GeoTag.
var GeoTagSchema = schema.New("geo.geotag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoTag is one geo tag.
type GeoTag struct {
	ID       string
	Channel  string
	Source   string
	Comment  string
	Weight   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoTag) Key() string { return "geo/geotag/" + x.ID }

// Fields encodes x for storage.
func (x GeoTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"source":   x.Source,
		"comment":  x.Comment,
		"weight":   strconv.Itoa(x.Weight),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoTag decodes a stored record.
func ParseGeoTag(f codec.Fields) (GeoTag, error) {
	var x GeoTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geotag: missing id")
	}
	x.Channel = f["channel"]
	x.Source = f["source"]
	x.Comment = f["comment"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("geo: geotag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("geo: geotag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geotag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geotag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoTag) String() string {
	return fmt.Sprintf("GeoTag(%s, channel=%q, weight=%d, archived=%t)", x.ID, x.Channel, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoTag) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoTag) Normalize() GeoTag {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortGeoTags orders by weight descending, then by ID.
func SortGeoTags(xs []GeoTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoTags drops the archived entries, preserving order.
func ActiveGeoTags(xs []GeoTag) []GeoTag {
	out := make([]GeoTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoTags adds up the weight of every active entry.
func SumGeoTags(xs []GeoTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexGeoTags keys the entries by ID, last write winning.
func IndexGeoTags(xs []GeoTag) map[string]GeoTag {
	out := make(map[string]GeoTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoTag builds a valid GeoTag, for tests and for the seed data the cron
// binary writes.
func SampleGeoTag(id string, at time.Time) GeoTag {
	return GeoTag{
		ID:       id,
		Channel:  "channel-" + id,
		Source:   "source-" + id,
		Comment:  "seeded by geo.SampleGeoTag",
		Weight:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoRuleSchema is the stored shape of a GeoRule.
var GeoRuleSchema = schema.New("geo.georule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoRule is one geo rule.
type GeoRule struct {
	ID       string
	Owner    string
	Region   string
	Source   string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoRule) Key() string { return "geo/georule/" + x.ID }

// Fields encodes x for storage.
func (x GeoRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"source":   x.Source,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoRule decodes a stored record.
func ParseGeoRule(f codec.Fields) (GeoRule, error) {
	var x GeoRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: georule: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("geo: georule %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("geo: georule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: georule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: georule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoRule) String() string {
	return fmt.Sprintf("GeoRule(%s, owner=%q, offset=%d, archived=%t)", x.ID, x.Owner, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoRule) Normalize() GeoRule {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortGeoRules orders by offset descending, then by ID.
func SortGeoRules(xs []GeoRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoRules drops the archived entries, preserving order.
func ActiveGeoRules(xs []GeoRule) []GeoRule {
	out := make([]GeoRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoRules adds up the offset of every active entry.
func SumGeoRules(xs []GeoRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexGeoRules keys the entries by ID, last write winning.
func IndexGeoRules(xs []GeoRule) map[string]GeoRule {
	out := make(map[string]GeoRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoRule builds a valid GeoRule, for tests and for the seed data the cron
// binary writes.
func SampleGeoRule(id string, at time.Time) GeoRule {
	return GeoRule{
		ID:      id,
		Owner:   "owner-" + id,
		Region:  "region-" + id,
		Source:  "seeded by geo.SampleGeoRule",
		Offset:  len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// GeoHoldSchema is the stored shape of a GeoHold.
var GeoHoldSchema = schema.New("geo.geohold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoHold is one geo hold.
type GeoHold struct {
	ID       string
	Category string
	Owner    string
	Region   string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoHold) Key() string { return "geo/geohold/" + x.ID }

// Fields encodes x for storage.
func (x GeoHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoHold decodes a stored record.
func ParseGeoHold(f codec.Fields) (GeoHold, error) {
	var x GeoHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geohold: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: geohold %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("geo: geohold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geohold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geohold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoHold) String() string {
	return fmt.Sprintf("GeoHold(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoHold) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoHold) Normalize() GeoHold {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortGeoHolds orders by capacity descending, then by ID.
func SortGeoHolds(xs []GeoHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoHolds drops the archived entries, preserving order.
func ActiveGeoHolds(xs []GeoHold) []GeoHold {
	out := make([]GeoHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoHolds adds up the capacity of every active entry.
func SumGeoHolds(xs []GeoHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexGeoHolds keys the entries by ID, last write winning.
func IndexGeoHolds(xs []GeoHold) map[string]GeoHold {
	out := make(map[string]GeoHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoHold builds a valid GeoHold, for tests and for the seed data the cron
// binary writes.
func SampleGeoHold(id string, at time.Time) GeoHold {
	return GeoHold{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Region:   "seeded by geo.SampleGeoHold",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// GeoClaimSchema is the stored shape of a GeoClaim.
var GeoClaimSchema = schema.New("geo.geoclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoClaim is one geo claim.
type GeoClaim struct {
	ID       string
	Owner    string
	Region   string
	Label    string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoClaim) Key() string { return "geo/geoclaim/" + x.ID }

// Fields encodes x for storage.
func (x GeoClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"region":   x.Region,
		"label":    x.Label,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoClaim decodes a stored record.
func ParseGeoClaim(f codec.Fields) (GeoClaim, error) {
	var x GeoClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geoclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("geo: geoclaim %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("geo: geoclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geoclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geoclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoClaim) String() string {
	return fmt.Sprintf("GeoClaim(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoClaim) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoClaim) Normalize() GeoClaim {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortGeoClaims orders by version descending, then by ID.
func SortGeoClaims(xs []GeoClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoClaims drops the archived entries, preserving order.
func ActiveGeoClaims(xs []GeoClaim) []GeoClaim {
	out := make([]GeoClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoClaims adds up the version of every active entry.
func SumGeoClaims(xs []GeoClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexGeoClaims keys the entries by ID, last write winning.
func IndexGeoClaims(xs []GeoClaim) map[string]GeoClaim {
	out := make(map[string]GeoClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoClaim builds a valid GeoClaim, for tests and for the seed data the cron
// binary writes.
func SampleGeoClaim(id string, at time.Time) GeoClaim {
	return GeoClaim{
		ID:      id,
		Owner:   "owner-" + id,
		Region:  "region-" + id,
		Label:   "seeded by geo.SampleGeoClaim",
		Version: len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// GeoCursorSchema is the stored shape of a GeoCursor.
var GeoCursorSchema = schema.New("geo.geocursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// GeoCursor is one geo cursor.
type GeoCursor struct {
	ID       string
	Region   string
	Comment  string
	Group    string
	Count    int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x GeoCursor) Key() string { return "geo/geocursor/" + x.ID }

// Fields encodes x for storage.
func (x GeoCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"comment":  x.Comment,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseGeoCursor decodes a stored record.
func ParseGeoCursor(f codec.Fields) (GeoCursor, error) {
	var x GeoCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("geo: geocursor: missing id")
	}
	x.Region = f["region"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("geo: geocursor %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("geo: geocursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("geo: geocursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("geo: geocursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x GeoCursor) String() string {
	return fmt.Sprintf("GeoCursor(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x GeoCursor) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x GeoCursor) Normalize() GeoCursor {
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortGeoCursors orders by count descending, then by ID.
func SortGeoCursors(xs []GeoCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveGeoCursors drops the archived entries, preserving order.
func ActiveGeoCursors(xs []GeoCursor) []GeoCursor {
	out := make([]GeoCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumGeoCursors adds up the count of every active entry.
func SumGeoCursors(xs []GeoCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexGeoCursors keys the entries by ID, last write winning.
func IndexGeoCursors(xs []GeoCursor) map[string]GeoCursor {
	out := make(map[string]GeoCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleGeoCursor builds a valid GeoCursor, for tests and for the seed data the cron
// binary writes.
func SampleGeoCursor(id string, at time.Time) GeoCursor {
	return GeoCursor{
		ID:       id,
		Region:   "region-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by geo.SampleGeoCursor",
		Count:    len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	GeoRecordSchema,
	GeoEntrySchema,
	GeoItemSchema,
	GeoDraftSchema,
	GeoRevisionSchema,
	GeoSnapshotSchema,
	GeoBatchSchema,
	GeoRequestSchema,
	GeoSummarySchema,
	GeoPolicySchema,
	GeoLinkSchema,
	GeoNoteSchema,
	GeoWindowSchema,
	GeoQuotaSchema,
	GeoBucketSchema,
	GeoLaneSchema,
	GeoSlotSchema,
	GeoTagSchema,
	GeoRuleSchema,
	GeoHoldSchema,
	GeoClaimSchema,
	GeoCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"geo/georecord/",
	"geo/geoentry/",
	"geo/geoitem/",
	"geo/geodraft/",
	"geo/georevision/",
	"geo/geosnapshot/",
	"geo/geobatch/",
	"geo/georequest/",
	"geo/geosummary/",
	"geo/geopolicy/",
	"geo/geolink/",
	"geo/geonote/",
	"geo/geowindow/",
	"geo/geoquota/",
	"geo/geobucket/",
	"geo/geolane/",
	"geo/geoslot/",
	"geo/geotag/",
	"geo/georule/",
	"geo/geohold/",
	"geo/geoclaim/",
	"geo/geocursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "geo: " + strings.Join(names, ", ")
}
