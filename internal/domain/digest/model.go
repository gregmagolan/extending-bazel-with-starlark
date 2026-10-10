// Package digest holds the digest domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package digest

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// DigestRecordSchema is the stored shape of a DigestRecord.
var DigestRecordSchema = schema.New("digest.digestrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestRecord is one digest record.
type DigestRecord struct {
	ID        string
	Reference string
	Channel   string
	Category  string
	Capacity  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestRecord) Key() string { return "digest/digestrecord/" + x.ID }

// Fields encodes x for storage.
func (x DigestRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"category":  x.Category,
		"capacity":  strconv.Itoa(x.Capacity),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestRecord decodes a stored record.
func ParseDigestRecord(f codec.Fields) (DigestRecord, error) {
	var x DigestRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestrecord: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("digest: digestrecord %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("digest: digestrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestRecord) String() string {
	return fmt.Sprintf("DigestRecord(%s, reference=%q, capacity=%d, archived=%t)", x.ID, x.Reference, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestRecord) Normalize() DigestRecord {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortDigestRecords orders by capacity descending, then by ID.
func SortDigestRecords(xs []DigestRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestRecords drops the archived entries, preserving order.
func ActiveDigestRecords(xs []DigestRecord) []DigestRecord {
	out := make([]DigestRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestRecords adds up the capacity of every active entry.
func SumDigestRecords(xs []DigestRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDigestRecords keys the entries by ID, last write winning.
func IndexDigestRecords(xs []DigestRecord) map[string]DigestRecord {
	out := make(map[string]DigestRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestRecord builds a valid DigestRecord, for tests and for the seed data the cron
// binary writes.
func SampleDigestRecord(id string, at time.Time) DigestRecord {
	return DigestRecord{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Category:  "seeded by digest.SampleDigestRecord",
		Capacity:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestEntrySchema is the stored shape of a DigestEntry.
var DigestEntrySchema = schema.New("digest.digestentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestEntry is one digest entry.
type DigestEntry struct {
	ID       string
	Source   string
	Group    string
	Region   string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestEntry) Key() string { return "digest/digestentry/" + x.ID }

// Fields encodes x for storage.
func (x DigestEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"group":    x.Group,
		"region":   x.Region,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestEntry decodes a stored record.
func ParseDigestEntry(f codec.Fields) (DigestEntry, error) {
	var x DigestEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestentry: missing id")
	}
	x.Source = f["source"]
	x.Group = f["group"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("digest: digestentry %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("digest: digestentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestEntry) String() string {
	return fmt.Sprintf("DigestEntry(%s, source=%q, priority=%d, archived=%t)", x.ID, x.Source, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestEntry) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestEntry) Normalize() DigestEntry {
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDigestEntrys orders by priority descending, then by ID.
func SortDigestEntrys(xs []DigestEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestEntrys drops the archived entries, preserving order.
func ActiveDigestEntrys(xs []DigestEntry) []DigestEntry {
	out := make([]DigestEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestEntrys adds up the priority of every active entry.
func SumDigestEntrys(xs []DigestEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDigestEntrys keys the entries by ID, last write winning.
func IndexDigestEntrys(xs []DigestEntry) map[string]DigestEntry {
	out := make(map[string]DigestEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestEntry builds a valid DigestEntry, for tests and for the seed data the cron
// binary writes.
func SampleDigestEntry(id string, at time.Time) DigestEntry {
	return DigestEntry{
		ID:       id,
		Source:   "source-" + id,
		Group:    "group-" + id,
		Region:   "seeded by digest.SampleDigestEntry",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestItemSchema is the stored shape of a DigestItem.
var DigestItemSchema = schema.New("digest.digestitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestItem is one digest item.
type DigestItem struct {
	ID        string
	Reference string
	Channel   string
	Label     string
	Weight    int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestItem) Key() string { return "digest/digestitem/" + x.ID }

// Fields encodes x for storage.
func (x DigestItem) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"label":     x.Label,
		"weight":    strconv.Itoa(x.Weight),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestItem decodes a stored record.
func ParseDigestItem(f codec.Fields) (DigestItem, error) {
	var x DigestItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestitem: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestitem %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("digest: digestitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestItem) String() string {
	return fmt.Sprintf("DigestItem(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestItem) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestItem) Normalize() DigestItem {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDigestItems orders by weight descending, then by ID.
func SortDigestItems(xs []DigestItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestItems drops the archived entries, preserving order.
func ActiveDigestItems(xs []DigestItem) []DigestItem {
	out := make([]DigestItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestItems adds up the weight of every active entry.
func SumDigestItems(xs []DigestItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDigestItems keys the entries by ID, last write winning.
func IndexDigestItems(xs []DigestItem) map[string]DigestItem {
	out := make(map[string]DigestItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestItem builds a valid DigestItem, for tests and for the seed data the cron
// binary writes.
func SampleDigestItem(id string, at time.Time) DigestItem {
	return DigestItem{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Label:     "seeded by digest.SampleDigestItem",
		Weight:    len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestDraftSchema is the stored shape of a DigestDraft.
var DigestDraftSchema = schema.New("digest.digestdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestDraft is one digest draft.
type DigestDraft struct {
	ID        string
	Name      string
	Category  string
	Reference string
	Weight    int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestDraft) Key() string { return "digest/digestdraft/" + x.ID }

// Fields encodes x for storage.
func (x DigestDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"category":  x.Category,
		"reference": x.Reference,
		"weight":    strconv.Itoa(x.Weight),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestDraft decodes a stored record.
func ParseDigestDraft(f codec.Fields) (DigestDraft, error) {
	var x DigestDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestdraft: missing id")
	}
	x.Name = f["name"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestdraft %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("digest: digestdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestDraft) String() string {
	return fmt.Sprintf("DigestDraft(%s, name=%q, weight=%d, archived=%t)", x.ID, x.Name, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestDraft) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestDraft) Normalize() DigestDraft {
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortDigestDrafts orders by weight descending, then by ID.
func SortDigestDrafts(xs []DigestDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestDrafts drops the archived entries, preserving order.
func ActiveDigestDrafts(xs []DigestDraft) []DigestDraft {
	out := make([]DigestDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestDrafts adds up the weight of every active entry.
func SumDigestDrafts(xs []DigestDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDigestDrafts keys the entries by ID, last write winning.
func IndexDigestDrafts(xs []DigestDraft) map[string]DigestDraft {
	out := make(map[string]DigestDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestDraft builds a valid DigestDraft, for tests and for the seed data the cron
// binary writes.
func SampleDigestDraft(id string, at time.Time) DigestDraft {
	return DigestDraft{
		ID:        id,
		Name:      "name-" + id,
		Category:  "category-" + id,
		Reference: "seeded by digest.SampleDigestDraft",
		Weight:    len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestRevisionSchema is the stored shape of a DigestRevision.
var DigestRevisionSchema = schema.New("digest.digestrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestRevision is one digest revision.
type DigestRevision struct {
	ID       string
	Label    string
	Channel  string
	Group    string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestRevision) Key() string { return "digest/digestrevision/" + x.ID }

// Fields encodes x for storage.
func (x DigestRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"group":    x.Group,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestRevision decodes a stored record.
func ParseDigestRevision(f codec.Fields) (DigestRevision, error) {
	var x DigestRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestrevision: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestrevision %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("digest: digestrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestRevision) String() string {
	return fmt.Sprintf("DigestRevision(%s, label=%q, weight=%d, archived=%t)", x.ID, x.Label, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestRevision) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestRevision) Normalize() DigestRevision {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDigestRevisions orders by weight descending, then by ID.
func SortDigestRevisions(xs []DigestRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestRevisions drops the archived entries, preserving order.
func ActiveDigestRevisions(xs []DigestRevision) []DigestRevision {
	out := make([]DigestRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestRevisions adds up the weight of every active entry.
func SumDigestRevisions(xs []DigestRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDigestRevisions keys the entries by ID, last write winning.
func IndexDigestRevisions(xs []DigestRevision) map[string]DigestRevision {
	out := make(map[string]DigestRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestRevision builds a valid DigestRevision, for tests and for the seed data the cron
// binary writes.
func SampleDigestRevision(id string, at time.Time) DigestRevision {
	return DigestRevision{
		ID:      id,
		Label:   "label-" + id,
		Channel: "channel-" + id,
		Group:   "seeded by digest.SampleDigestRevision",
		Weight:  len(id),
		Count:   len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// DigestSnapshotSchema is the stored shape of a DigestSnapshot.
var DigestSnapshotSchema = schema.New("digest.digestsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestSnapshot is one digest snapshot.
type DigestSnapshot struct {
	ID       string
	Region   string
	Source   string
	Label    string
	Weight   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestSnapshot) Key() string { return "digest/digestsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x DigestSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"source":   x.Source,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestSnapshot decodes a stored record.
func ParseDigestSnapshot(f codec.Fields) (DigestSnapshot, error) {
	var x DigestSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestsnapshot: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestsnapshot %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("digest: digestsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestSnapshot) String() string {
	return fmt.Sprintf("DigestSnapshot(%s, region=%q, weight=%d, archived=%t)", x.ID, x.Region, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestSnapshot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestSnapshot) Normalize() DigestSnapshot {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDigestSnapshots orders by weight descending, then by ID.
func SortDigestSnapshots(xs []DigestSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestSnapshots drops the archived entries, preserving order.
func ActiveDigestSnapshots(xs []DigestSnapshot) []DigestSnapshot {
	out := make([]DigestSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestSnapshots adds up the weight of every active entry.
func SumDigestSnapshots(xs []DigestSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDigestSnapshots keys the entries by ID, last write winning.
func IndexDigestSnapshots(xs []DigestSnapshot) map[string]DigestSnapshot {
	out := make(map[string]DigestSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestSnapshot builds a valid DigestSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleDigestSnapshot(id string, at time.Time) DigestSnapshot {
	return DigestSnapshot{
		ID:       id,
		Region:   "region-" + id,
		Source:   "source-" + id,
		Label:    "seeded by digest.SampleDigestSnapshot",
		Weight:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestBatchSchema is the stored shape of a DigestBatch.
var DigestBatchSchema = schema.New("digest.digestbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestBatch is one digest batch.
type DigestBatch struct {
	ID       string
	Comment  string
	Source   string
	Name     string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestBatch) Key() string { return "digest/digestbatch/" + x.ID }

// Fields encodes x for storage.
func (x DigestBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"name":     x.Name,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestBatch decodes a stored record.
func ParseDigestBatch(f codec.Fields) (DigestBatch, error) {
	var x DigestBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestbatch: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("digest: digestbatch %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("digest: digestbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestBatch) String() string {
	return fmt.Sprintf("DigestBatch(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestBatch) Normalize() DigestBatch {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDigestBatchs orders by version descending, then by ID.
func SortDigestBatchs(xs []DigestBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestBatchs drops the archived entries, preserving order.
func ActiveDigestBatchs(xs []DigestBatch) []DigestBatch {
	out := make([]DigestBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestBatchs adds up the version of every active entry.
func SumDigestBatchs(xs []DigestBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDigestBatchs keys the entries by ID, last write winning.
func IndexDigestBatchs(xs []DigestBatch) map[string]DigestBatch {
	out := make(map[string]DigestBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestBatch builds a valid DigestBatch, for tests and for the seed data the cron
// binary writes.
func SampleDigestBatch(id string, at time.Time) DigestBatch {
	return DigestBatch{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Name:     "seeded by digest.SampleDigestBatch",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestRequestSchema is the stored shape of a DigestRequest.
var DigestRequestSchema = schema.New("digest.digestrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestRequest is one digest request.
type DigestRequest struct {
	ID       string
	Channel  string
	Source   string
	Label    string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestRequest) Key() string { return "digest/digestrequest/" + x.ID }

// Fields encodes x for storage.
func (x DigestRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"source":   x.Source,
		"label":    x.Label,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestRequest decodes a stored record.
func ParseDigestRequest(f codec.Fields) (DigestRequest, error) {
	var x DigestRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestrequest: missing id")
	}
	x.Channel = f["channel"]
	x.Source = f["source"]
	x.Label = f["label"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("digest: digestrequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("digest: digestrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestRequest) String() string {
	return fmt.Sprintf("DigestRequest(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestRequest) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestRequest) Normalize() DigestRequest {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortDigestRequests orders by capacity descending, then by ID.
func SortDigestRequests(xs []DigestRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestRequests drops the archived entries, preserving order.
func ActiveDigestRequests(xs []DigestRequest) []DigestRequest {
	out := make([]DigestRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestRequests adds up the capacity of every active entry.
func SumDigestRequests(xs []DigestRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDigestRequests keys the entries by ID, last write winning.
func IndexDigestRequests(xs []DigestRequest) map[string]DigestRequest {
	out := make(map[string]DigestRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestRequest builds a valid DigestRequest, for tests and for the seed data the cron
// binary writes.
func SampleDigestRequest(id string, at time.Time) DigestRequest {
	return DigestRequest{
		ID:       id,
		Channel:  "channel-" + id,
		Source:   "source-" + id,
		Label:    "seeded by digest.SampleDigestRequest",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestSummarySchema is the stored shape of a DigestSummary.
var DigestSummarySchema = schema.New("digest.digestsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestSummary is one digest summary.
type DigestSummary struct {
	ID       string
	Name     string
	Region   string
	Category string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestSummary) Key() string { return "digest/digestsummary/" + x.ID }

// Fields encodes x for storage.
func (x DigestSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"region":   x.Region,
		"category": x.Category,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestSummary decodes a stored record.
func ParseDigestSummary(f codec.Fields) (DigestSummary, error) {
	var x DigestSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestsummary: missing id")
	}
	x.Name = f["name"]
	x.Region = f["region"]
	x.Category = f["category"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("digest: digestsummary %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("digest: digestsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestSummary) String() string {
	return fmt.Sprintf("DigestSummary(%s, name=%q, attempts=%d, archived=%t)", x.ID, x.Name, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestSummary) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestSummary) Normalize() DigestSummary {
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortDigestSummarys orders by attempts descending, then by ID.
func SortDigestSummarys(xs []DigestSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestSummarys drops the archived entries, preserving order.
func ActiveDigestSummarys(xs []DigestSummary) []DigestSummary {
	out := make([]DigestSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestSummarys adds up the attempts of every active entry.
func SumDigestSummarys(xs []DigestSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDigestSummarys keys the entries by ID, last write winning.
func IndexDigestSummarys(xs []DigestSummary) map[string]DigestSummary {
	out := make(map[string]DigestSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestSummary builds a valid DigestSummary, for tests and for the seed data the cron
// binary writes.
func SampleDigestSummary(id string, at time.Time) DigestSummary {
	return DigestSummary{
		ID:       id,
		Name:     "name-" + id,
		Region:   "region-" + id,
		Category: "seeded by digest.SampleDigestSummary",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestPolicySchema is the stored shape of a DigestPolicy.
var DigestPolicySchema = schema.New("digest.digestpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestPolicy is one digest policy.
type DigestPolicy struct {
	ID        string
	Owner     string
	Group     string
	Reference string
	Weight    int
	Count     int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestPolicy) Key() string { return "digest/digestpolicy/" + x.ID }

// Fields encodes x for storage.
func (x DigestPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"owner":     x.Owner,
		"group":     x.Group,
		"reference": x.Reference,
		"weight":    strconv.Itoa(x.Weight),
		"count":     strconv.Itoa(x.Count),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestPolicy decodes a stored record.
func ParseDigestPolicy(f codec.Fields) (DigestPolicy, error) {
	var x DigestPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestpolicy: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Reference = f["reference"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestpolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("digest: digestpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestPolicy) String() string {
	return fmt.Sprintf("DigestPolicy(%s, owner=%q, weight=%d, archived=%t)", x.ID, x.Owner, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestPolicy) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestPolicy) Normalize() DigestPolicy {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortDigestPolicys orders by weight descending, then by ID.
func SortDigestPolicys(xs []DigestPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestPolicys drops the archived entries, preserving order.
func ActiveDigestPolicys(xs []DigestPolicy) []DigestPolicy {
	out := make([]DigestPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestPolicys adds up the weight of every active entry.
func SumDigestPolicys(xs []DigestPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexDigestPolicys keys the entries by ID, last write winning.
func IndexDigestPolicys(xs []DigestPolicy) map[string]DigestPolicy {
	out := make(map[string]DigestPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestPolicy builds a valid DigestPolicy, for tests and for the seed data the cron
// binary writes.
func SampleDigestPolicy(id string, at time.Time) DigestPolicy {
	return DigestPolicy{
		ID:        id,
		Owner:     "owner-" + id,
		Group:     "group-" + id,
		Reference: "seeded by digest.SampleDigestPolicy",
		Weight:    len(id),
		Count:     len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestLinkSchema is the stored shape of a DigestLink.
var DigestLinkSchema = schema.New("digest.digestlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestLink is one digest link.
type DigestLink struct {
	ID        string
	Reference string
	Group     string
	Source    string
	Retained  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestLink) Key() string { return "digest/digestlink/" + x.ID }

// Fields encodes x for storage.
func (x DigestLink) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"group":     x.Group,
		"source":    x.Source,
		"retained":  strconv.Itoa(x.Retained),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestLink decodes a stored record.
func ParseDigestLink(f codec.Fields) (DigestLink, error) {
	var x DigestLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestlink: missing id")
	}
	x.Reference = f["reference"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digestlink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestLink) String() string {
	return fmt.Sprintf("DigestLink(%s, reference=%q, retained=%d, archived=%t)", x.ID, x.Reference, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestLink) Normalize() DigestLink {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDigestLinks orders by retained descending, then by ID.
func SortDigestLinks(xs []DigestLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestLinks drops the archived entries, preserving order.
func ActiveDigestLinks(xs []DigestLink) []DigestLink {
	out := make([]DigestLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestLinks adds up the retained of every active entry.
func SumDigestLinks(xs []DigestLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDigestLinks keys the entries by ID, last write winning.
func IndexDigestLinks(xs []DigestLink) map[string]DigestLink {
	out := make(map[string]DigestLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestLink builds a valid DigestLink, for tests and for the seed data the cron
// binary writes.
func SampleDigestLink(id string, at time.Time) DigestLink {
	return DigestLink{
		ID:        id,
		Reference: "reference-" + id,
		Group:     "group-" + id,
		Source:    "seeded by digest.SampleDigestLink",
		Retained:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestNoteSchema is the stored shape of a DigestNote.
var DigestNoteSchema = schema.New("digest.digestnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestNote is one digest note.
type DigestNote struct {
	ID       string
	Channel  string
	Name     string
	Source   string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestNote) Key() string { return "digest/digestnote/" + x.ID }

// Fields encodes x for storage.
func (x DigestNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"name":     x.Name,
		"source":   x.Source,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestNote decodes a stored record.
func ParseDigestNote(f codec.Fields) (DigestNote, error) {
	var x DigestNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestnote: missing id")
	}
	x.Channel = f["channel"]
	x.Name = f["name"]
	x.Source = f["source"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("digest: digestnote %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digestnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestNote) String() string {
	return fmt.Sprintf("DigestNote(%s, channel=%q, capacity=%d, archived=%t)", x.ID, x.Channel, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestNote) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestNote) Normalize() DigestNote {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDigestNotes orders by capacity descending, then by ID.
func SortDigestNotes(xs []DigestNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestNotes drops the archived entries, preserving order.
func ActiveDigestNotes(xs []DigestNote) []DigestNote {
	out := make([]DigestNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestNotes adds up the capacity of every active entry.
func SumDigestNotes(xs []DigestNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexDigestNotes keys the entries by ID, last write winning.
func IndexDigestNotes(xs []DigestNote) map[string]DigestNote {
	out := make(map[string]DigestNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestNote builds a valid DigestNote, for tests and for the seed data the cron
// binary writes.
func SampleDigestNote(id string, at time.Time) DigestNote {
	return DigestNote{
		ID:       id,
		Channel:  "channel-" + id,
		Name:     "name-" + id,
		Source:   "seeded by digest.SampleDigestNote",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestWindowSchema is the stored shape of a DigestWindow.
var DigestWindowSchema = schema.New("digest.digestwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestWindow is one digest window.
type DigestWindow struct {
	ID       string
	Name     string
	Source   string
	Channel  string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestWindow) Key() string { return "digest/digestwindow/" + x.ID }

// Fields encodes x for storage.
func (x DigestWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"channel":  x.Channel,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestWindow decodes a stored record.
func ParseDigestWindow(f codec.Fields) (DigestWindow, error) {
	var x DigestWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestwindow: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digestwindow %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestWindow) String() string {
	return fmt.Sprintf("DigestWindow(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestWindow) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestWindow) Normalize() DigestWindow {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDigestWindows orders by retained descending, then by ID.
func SortDigestWindows(xs []DigestWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestWindows drops the archived entries, preserving order.
func ActiveDigestWindows(xs []DigestWindow) []DigestWindow {
	out := make([]DigestWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestWindows adds up the retained of every active entry.
func SumDigestWindows(xs []DigestWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDigestWindows keys the entries by ID, last write winning.
func IndexDigestWindows(xs []DigestWindow) map[string]DigestWindow {
	out := make(map[string]DigestWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestWindow builds a valid DigestWindow, for tests and for the seed data the cron
// binary writes.
func SampleDigestWindow(id string, at time.Time) DigestWindow {
	return DigestWindow{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by digest.SampleDigestWindow",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestQuotaSchema is the stored shape of a DigestQuota.
var DigestQuotaSchema = schema.New("digest.digestquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestQuota is one digest quota.
type DigestQuota struct {
	ID       string
	Label    string
	Source   string
	Group    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestQuota) Key() string { return "digest/digestquota/" + x.ID }

// Fields encodes x for storage.
func (x DigestQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestQuota decodes a stored record.
func ParseDigestQuota(f codec.Fields) (DigestQuota, error) {
	var x DigestQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestquota: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("digest: digestquota %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestQuota) String() string {
	return fmt.Sprintf("DigestQuota(%s, label=%q, count=%d, archived=%t)", x.ID, x.Label, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestQuota) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestQuota) Normalize() DigestQuota {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDigestQuotas orders by count descending, then by ID.
func SortDigestQuotas(xs []DigestQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestQuotas drops the archived entries, preserving order.
func ActiveDigestQuotas(xs []DigestQuota) []DigestQuota {
	out := make([]DigestQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestQuotas adds up the count of every active entry.
func SumDigestQuotas(xs []DigestQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexDigestQuotas keys the entries by ID, last write winning.
func IndexDigestQuotas(xs []DigestQuota) map[string]DigestQuota {
	out := make(map[string]DigestQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestQuota builds a valid DigestQuota, for tests and for the seed data the cron
// binary writes.
func SampleDigestQuota(id string, at time.Time) DigestQuota {
	return DigestQuota{
		ID:      id,
		Label:   "label-" + id,
		Source:  "source-" + id,
		Group:   "seeded by digest.SampleDigestQuota",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// DigestBucketSchema is the stored shape of a DigestBucket.
var DigestBucketSchema = schema.New("digest.digestbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestBucket is one digest bucket.
type DigestBucket struct {
	ID       string
	Owner    string
	Label    string
	Group    string
	Priority int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestBucket) Key() string { return "digest/digestbucket/" + x.ID }

// Fields encodes x for storage.
func (x DigestBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"label":    x.Label,
		"group":    x.Group,
		"priority": strconv.Itoa(x.Priority),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestBucket decodes a stored record.
func ParseDigestBucket(f codec.Fields) (DigestBucket, error) {
	var x DigestBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestbucket: missing id")
	}
	x.Owner = f["owner"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("digest: digestbucket %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("digest: digestbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestBucket) String() string {
	return fmt.Sprintf("DigestBucket(%s, owner=%q, priority=%d, archived=%t)", x.ID, x.Owner, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestBucket) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestBucket) Normalize() DigestBucket {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDigestBuckets orders by priority descending, then by ID.
func SortDigestBuckets(xs []DigestBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestBuckets drops the archived entries, preserving order.
func ActiveDigestBuckets(xs []DigestBucket) []DigestBucket {
	out := make([]DigestBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestBuckets adds up the priority of every active entry.
func SumDigestBuckets(xs []DigestBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDigestBuckets keys the entries by ID, last write winning.
func IndexDigestBuckets(xs []DigestBucket) map[string]DigestBucket {
	out := make(map[string]DigestBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestBucket builds a valid DigestBucket, for tests and for the seed data the cron
// binary writes.
func SampleDigestBucket(id string, at time.Time) DigestBucket {
	return DigestBucket{
		ID:       id,
		Owner:    "owner-" + id,
		Label:    "label-" + id,
		Group:    "seeded by digest.SampleDigestBucket",
		Priority: len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestLaneSchema is the stored shape of a DigestLane.
var DigestLaneSchema = schema.New("digest.digestlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestLane is one digest lane.
type DigestLane struct {
	ID       string
	Label    string
	Group    string
	Source   string
	Priority int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestLane) Key() string { return "digest/digestlane/" + x.ID }

// Fields encodes x for storage.
func (x DigestLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"source":   x.Source,
		"priority": strconv.Itoa(x.Priority),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestLane decodes a stored record.
func ParseDigestLane(f codec.Fields) (DigestLane, error) {
	var x DigestLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestlane: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Source = f["source"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("digest: digestlane %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digestlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestLane) String() string {
	return fmt.Sprintf("DigestLane(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestLane) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestLane) Normalize() DigestLane {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDigestLanes orders by priority descending, then by ID.
func SortDigestLanes(xs []DigestLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestLanes drops the archived entries, preserving order.
func ActiveDigestLanes(xs []DigestLane) []DigestLane {
	out := make([]DigestLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestLanes adds up the priority of every active entry.
func SumDigestLanes(xs []DigestLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexDigestLanes keys the entries by ID, last write winning.
func IndexDigestLanes(xs []DigestLane) map[string]DigestLane {
	out := make(map[string]DigestLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestLane builds a valid DigestLane, for tests and for the seed data the cron
// binary writes.
func SampleDigestLane(id string, at time.Time) DigestLane {
	return DigestLane{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Source:   "seeded by digest.SampleDigestLane",
		Priority: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestSlotSchema is the stored shape of a DigestSlot.
var DigestSlotSchema = schema.New("digest.digestslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestSlot is one digest slot.
type DigestSlot struct {
	ID        string
	Group     string
	Reference string
	Source    string
	Attempts  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestSlot) Key() string { return "digest/digestslot/" + x.ID }

// Fields encodes x for storage.
func (x DigestSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"source":    x.Source,
		"attempts":  strconv.Itoa(x.Attempts),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestSlot decodes a stored record.
func ParseDigestSlot(f codec.Fields) (DigestSlot, error) {
	var x DigestSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestslot: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("digest: digestslot %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("digest: digestslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestSlot) String() string {
	return fmt.Sprintf("DigestSlot(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestSlot) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestSlot) Normalize() DigestSlot {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortDigestSlots orders by attempts descending, then by ID.
func SortDigestSlots(xs []DigestSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestSlots drops the archived entries, preserving order.
func ActiveDigestSlots(xs []DigestSlot) []DigestSlot {
	out := make([]DigestSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestSlots adds up the attempts of every active entry.
func SumDigestSlots(xs []DigestSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDigestSlots keys the entries by ID, last write winning.
func IndexDigestSlots(xs []DigestSlot) map[string]DigestSlot {
	out := make(map[string]DigestSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestSlot builds a valid DigestSlot, for tests and for the seed data the cron
// binary writes.
func SampleDigestSlot(id string, at time.Time) DigestSlot {
	return DigestSlot{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by digest.SampleDigestSlot",
		Attempts:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestTagSchema is the stored shape of a DigestTag.
var DigestTagSchema = schema.New("digest.digesttag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestTag is one digest tag.
type DigestTag struct {
	ID       string
	Category string
	Group    string
	Label    string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestTag) Key() string { return "digest/digesttag/" + x.ID }

// Fields encodes x for storage.
func (x DigestTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"group":    x.Group,
		"label":    x.Label,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestTag decodes a stored record.
func ParseDigestTag(f codec.Fields) (DigestTag, error) {
	var x DigestTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digesttag: missing id")
	}
	x.Category = f["category"]
	x.Group = f["group"]
	x.Label = f["label"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("digest: digesttag %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digesttag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digesttag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digesttag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestTag) String() string {
	return fmt.Sprintf("DigestTag(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestTag) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestTag) Normalize() DigestTag {
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	x.Label = strings.TrimSpace(x.Label)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDigestTags orders by offset descending, then by ID.
func SortDigestTags(xs []DigestTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestTags drops the archived entries, preserving order.
func ActiveDigestTags(xs []DigestTag) []DigestTag {
	out := make([]DigestTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestTags adds up the offset of every active entry.
func SumDigestTags(xs []DigestTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDigestTags keys the entries by ID, last write winning.
func IndexDigestTags(xs []DigestTag) map[string]DigestTag {
	out := make(map[string]DigestTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestTag builds a valid DigestTag, for tests and for the seed data the cron
// binary writes.
func SampleDigestTag(id string, at time.Time) DigestTag {
	return DigestTag{
		ID:       id,
		Category: "category-" + id,
		Group:    "group-" + id,
		Label:    "seeded by digest.SampleDigestTag",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestRuleSchema is the stored shape of a DigestRule.
var DigestRuleSchema = schema.New("digest.digestrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestRule is one digest rule.
type DigestRule struct {
	ID        string
	Comment   string
	Owner     string
	Reference string
	Offset    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestRule) Key() string { return "digest/digestrule/" + x.ID }

// Fields encodes x for storage.
func (x DigestRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"owner":     x.Owner,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestRule decodes a stored record.
func ParseDigestRule(f codec.Fields) (DigestRule, error) {
	var x DigestRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestrule: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("digest: digestrule %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digestrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestRule) String() string {
	return fmt.Sprintf("DigestRule(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestRule) Normalize() DigestRule {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDigestRules orders by offset descending, then by ID.
func SortDigestRules(xs []DigestRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestRules drops the archived entries, preserving order.
func ActiveDigestRules(xs []DigestRule) []DigestRule {
	out := make([]DigestRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestRules adds up the offset of every active entry.
func SumDigestRules(xs []DigestRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexDigestRules keys the entries by ID, last write winning.
func IndexDigestRules(xs []DigestRule) map[string]DigestRule {
	out := make(map[string]DigestRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestRule builds a valid DigestRule, for tests and for the seed data the cron
// binary writes.
func SampleDigestRule(id string, at time.Time) DigestRule {
	return DigestRule{
		ID:        id,
		Comment:   "comment-" + id,
		Owner:     "owner-" + id,
		Reference: "seeded by digest.SampleDigestRule",
		Offset:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestHoldSchema is the stored shape of a DigestHold.
var DigestHoldSchema = schema.New("digest.digesthold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestHold is one digest hold.
type DigestHold struct {
	ID        string
	Region    string
	Category  string
	Reference string
	Attempts  int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestHold) Key() string { return "digest/digesthold/" + x.ID }

// Fields encodes x for storage.
func (x DigestHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"category":  x.Category,
		"reference": x.Reference,
		"attempts":  strconv.Itoa(x.Attempts),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestHold decodes a stored record.
func ParseDigestHold(f codec.Fields) (DigestHold, error) {
	var x DigestHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digesthold: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("digest: digesthold %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("digest: digesthold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digesthold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digesthold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestHold) String() string {
	return fmt.Sprintf("DigestHold(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestHold) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestHold) Normalize() DigestHold {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortDigestHolds orders by attempts descending, then by ID.
func SortDigestHolds(xs []DigestHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestHolds drops the archived entries, preserving order.
func ActiveDigestHolds(xs []DigestHold) []DigestHold {
	out := make([]DigestHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestHolds adds up the attempts of every active entry.
func SumDigestHolds(xs []DigestHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexDigestHolds keys the entries by ID, last write winning.
func IndexDigestHolds(xs []DigestHold) map[string]DigestHold {
	out := make(map[string]DigestHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestHold builds a valid DigestHold, for tests and for the seed data the cron
// binary writes.
func SampleDigestHold(id string, at time.Time) DigestHold {
	return DigestHold{
		ID:        id,
		Region:    "region-" + id,
		Category:  "category-" + id,
		Reference: "seeded by digest.SampleDigestHold",
		Attempts:  len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// DigestClaimSchema is the stored shape of a DigestClaim.
var DigestClaimSchema = schema.New("digest.digestclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestClaim is one digest claim.
type DigestClaim struct {
	ID       string
	Name     string
	Region   string
	Comment  string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x DigestClaim) Key() string { return "digest/digestclaim/" + x.ID }

// Fields encodes x for storage.
func (x DigestClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"region":   x.Region,
		"comment":  x.Comment,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestClaim decodes a stored record.
func ParseDigestClaim(f codec.Fields) (DigestClaim, error) {
	var x DigestClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestclaim: missing id")
	}
	x.Name = f["name"]
	x.Region = f["region"]
	x.Comment = f["comment"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digestclaim %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("digest: digestclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestClaim) String() string {
	return fmt.Sprintf("DigestClaim(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestClaim) Normalize() DigestClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortDigestClaims orders by retained descending, then by ID.
func SortDigestClaims(xs []DigestClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestClaims drops the archived entries, preserving order.
func ActiveDigestClaims(xs []DigestClaim) []DigestClaim {
	out := make([]DigestClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestClaims adds up the retained of every active entry.
func SumDigestClaims(xs []DigestClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexDigestClaims keys the entries by ID, last write winning.
func IndexDigestClaims(xs []DigestClaim) map[string]DigestClaim {
	out := make(map[string]DigestClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestClaim builds a valid DigestClaim, for tests and for the seed data the cron
// binary writes.
func SampleDigestClaim(id string, at time.Time) DigestClaim {
	return DigestClaim{
		ID:       id,
		Name:     "name-" + id,
		Region:   "region-" + id,
		Comment:  "seeded by digest.SampleDigestClaim",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// DigestCursorSchema is the stored shape of a DigestCursor.
var DigestCursorSchema = schema.New("digest.digestcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// DigestCursor is one digest cursor.
type DigestCursor struct {
	ID        string
	Reference string
	Name      string
	Group     string
	Version   int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x DigestCursor) Key() string { return "digest/digestcursor/" + x.ID }

// Fields encodes x for storage.
func (x DigestCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"name":      x.Name,
		"group":     x.Group,
		"version":   strconv.Itoa(x.Version),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseDigestCursor decodes a stored record.
func ParseDigestCursor(f codec.Fields) (DigestCursor, error) {
	var x DigestCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("digest: digestcursor: missing id")
	}
	x.Reference = f["reference"]
	x.Name = f["name"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("digest: digestcursor %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("digest: digestcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("digest: digestcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("digest: digestcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x DigestCursor) String() string {
	return fmt.Sprintf("DigestCursor(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x DigestCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x DigestCursor) Normalize() DigestCursor {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Name = strings.TrimSpace(x.Name)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortDigestCursors orders by version descending, then by ID.
func SortDigestCursors(xs []DigestCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveDigestCursors drops the archived entries, preserving order.
func ActiveDigestCursors(xs []DigestCursor) []DigestCursor {
	out := make([]DigestCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumDigestCursors adds up the version of every active entry.
func SumDigestCursors(xs []DigestCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexDigestCursors keys the entries by ID, last write winning.
func IndexDigestCursors(xs []DigestCursor) map[string]DigestCursor {
	out := make(map[string]DigestCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleDigestCursor builds a valid DigestCursor, for tests and for the seed data the cron
// binary writes.
func SampleDigestCursor(id string, at time.Time) DigestCursor {
	return DigestCursor{
		ID:        id,
		Reference: "reference-" + id,
		Name:      "name-" + id,
		Group:     "seeded by digest.SampleDigestCursor",
		Version:   len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	DigestRecordSchema,
	DigestEntrySchema,
	DigestItemSchema,
	DigestDraftSchema,
	DigestRevisionSchema,
	DigestSnapshotSchema,
	DigestBatchSchema,
	DigestRequestSchema,
	DigestSummarySchema,
	DigestPolicySchema,
	DigestLinkSchema,
	DigestNoteSchema,
	DigestWindowSchema,
	DigestQuotaSchema,
	DigestBucketSchema,
	DigestLaneSchema,
	DigestSlotSchema,
	DigestTagSchema,
	DigestRuleSchema,
	DigestHoldSchema,
	DigestClaimSchema,
	DigestCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"digest/digestrecord/",
	"digest/digestentry/",
	"digest/digestitem/",
	"digest/digestdraft/",
	"digest/digestrevision/",
	"digest/digestsnapshot/",
	"digest/digestbatch/",
	"digest/digestrequest/",
	"digest/digestsummary/",
	"digest/digestpolicy/",
	"digest/digestlink/",
	"digest/digestnote/",
	"digest/digestwindow/",
	"digest/digestquota/",
	"digest/digestbucket/",
	"digest/digestlane/",
	"digest/digestslot/",
	"digest/digesttag/",
	"digest/digestrule/",
	"digest/digesthold/",
	"digest/digestclaim/",
	"digest/digestcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "digest: " + strings.Join(names, ", ")
}
