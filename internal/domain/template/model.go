// Package template holds the template domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package template

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// TemplateRecordSchema is the stored shape of a TemplateRecord.
var TemplateRecordSchema = schema.New("template.templaterecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateRecord is one template record.
type TemplateRecord struct {
	ID        string
	Reference string
	Channel   string
	Label     string
	Version   int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TemplateRecord) Key() string { return "template/templaterecord/" + x.ID }

// Fields encodes x for storage.
func (x TemplateRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"label":     x.Label,
		"version":   strconv.Itoa(x.Version),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateRecord decodes a stored record.
func ParseTemplateRecord(f codec.Fields) (TemplateRecord, error) {
	var x TemplateRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templaterecord: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("template: templaterecord %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templaterecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templaterecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templaterecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateRecord) String() string {
	return fmt.Sprintf("TemplateRecord(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateRecord) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateRecord) Normalize() TemplateRecord {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortTemplateRecords orders by version descending, then by ID.
func SortTemplateRecords(xs []TemplateRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateRecords drops the archived entries, preserving order.
func ActiveTemplateRecords(xs []TemplateRecord) []TemplateRecord {
	out := make([]TemplateRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateRecords adds up the version of every active entry.
func SumTemplateRecords(xs []TemplateRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTemplateRecords keys the entries by ID, last write winning.
func IndexTemplateRecords(xs []TemplateRecord) map[string]TemplateRecord {
	out := make(map[string]TemplateRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateRecord builds a valid TemplateRecord, for tests and for the seed data the cron
// binary writes.
func SampleTemplateRecord(id string, at time.Time) TemplateRecord {
	return TemplateRecord{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Label:     "seeded by template.SampleTemplateRecord",
		Version:   len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TemplateEntrySchema is the stored shape of a TemplateEntry.
var TemplateEntrySchema = schema.New("template.templateentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateEntry is one template entry.
type TemplateEntry struct {
	ID        string
	Label     string
	Category  string
	Reference string
	Priority  int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TemplateEntry) Key() string { return "template/templateentry/" + x.ID }

// Fields encodes x for storage.
func (x TemplateEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"category":  x.Category,
		"reference": x.Reference,
		"priority":  strconv.Itoa(x.Priority),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateEntry decodes a stored record.
func ParseTemplateEntry(f codec.Fields) (TemplateEntry, error) {
	var x TemplateEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templateentry: missing id")
	}
	x.Label = f["label"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("template: templateentry %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templateentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templateentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templateentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateEntry) String() string {
	return fmt.Sprintf("TemplateEntry(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateEntry) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateEntry) Normalize() TemplateEntry {
	x.Label = strings.TrimSpace(x.Label)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortTemplateEntrys orders by priority descending, then by ID.
func SortTemplateEntrys(xs []TemplateEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateEntrys drops the archived entries, preserving order.
func ActiveTemplateEntrys(xs []TemplateEntry) []TemplateEntry {
	out := make([]TemplateEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateEntrys adds up the priority of every active entry.
func SumTemplateEntrys(xs []TemplateEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTemplateEntrys keys the entries by ID, last write winning.
func IndexTemplateEntrys(xs []TemplateEntry) map[string]TemplateEntry {
	out := make(map[string]TemplateEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateEntry builds a valid TemplateEntry, for tests and for the seed data the cron
// binary writes.
func SampleTemplateEntry(id string, at time.Time) TemplateEntry {
	return TemplateEntry{
		ID:        id,
		Label:     "label-" + id,
		Category:  "category-" + id,
		Reference: "seeded by template.SampleTemplateEntry",
		Priority:  len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TemplateItemSchema is the stored shape of a TemplateItem.
var TemplateItemSchema = schema.New("template.templateitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateItem is one template item.
type TemplateItem struct {
	ID       string
	Channel  string
	Region   string
	Source   string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateItem) Key() string { return "template/templateitem/" + x.ID }

// Fields encodes x for storage.
func (x TemplateItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"source":   x.Source,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateItem decodes a stored record.
func ParseTemplateItem(f codec.Fields) (TemplateItem, error) {
	var x TemplateItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templateitem: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templateitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("template: templateitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templateitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templateitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateItem) String() string {
	return fmt.Sprintf("TemplateItem(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateItem) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateItem) Normalize() TemplateItem {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTemplateItems orders by attempts descending, then by ID.
func SortTemplateItems(xs []TemplateItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateItems drops the archived entries, preserving order.
func ActiveTemplateItems(xs []TemplateItem) []TemplateItem {
	out := make([]TemplateItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateItems adds up the attempts of every active entry.
func SumTemplateItems(xs []TemplateItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTemplateItems keys the entries by ID, last write winning.
func IndexTemplateItems(xs []TemplateItem) map[string]TemplateItem {
	out := make(map[string]TemplateItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateItem builds a valid TemplateItem, for tests and for the seed data the cron
// binary writes.
func SampleTemplateItem(id string, at time.Time) TemplateItem {
	return TemplateItem{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Source:   "seeded by template.SampleTemplateItem",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateDraftSchema is the stored shape of a TemplateDraft.
var TemplateDraftSchema = schema.New("template.templatedraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateDraft is one template draft.
type TemplateDraft struct {
	ID       string
	Owner    string
	Comment  string
	Group    string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateDraft) Key() string { return "template/templatedraft/" + x.ID }

// Fields encodes x for storage.
func (x TemplateDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateDraft decodes a stored record.
func ParseTemplateDraft(f codec.Fields) (TemplateDraft, error) {
	var x TemplateDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatedraft: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templatedraft %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("template: templatedraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatedraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatedraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateDraft) String() string {
	return fmt.Sprintf("TemplateDraft(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateDraft) Normalize() TemplateDraft {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTemplateDrafts orders by attempts descending, then by ID.
func SortTemplateDrafts(xs []TemplateDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateDrafts drops the archived entries, preserving order.
func ActiveTemplateDrafts(xs []TemplateDraft) []TemplateDraft {
	out := make([]TemplateDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateDrafts adds up the attempts of every active entry.
func SumTemplateDrafts(xs []TemplateDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTemplateDrafts keys the entries by ID, last write winning.
func IndexTemplateDrafts(xs []TemplateDraft) map[string]TemplateDraft {
	out := make(map[string]TemplateDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateDraft builds a valid TemplateDraft, for tests and for the seed data the cron
// binary writes.
func SampleTemplateDraft(id string, at time.Time) TemplateDraft {
	return TemplateDraft{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by template.SampleTemplateDraft",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateRevisionSchema is the stored shape of a TemplateRevision.
var TemplateRevisionSchema = schema.New("template.templaterevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateRevision is one template revision.
type TemplateRevision struct {
	ID       string
	Channel  string
	Name     string
	Comment  string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateRevision) Key() string { return "template/templaterevision/" + x.ID }

// Fields encodes x for storage.
func (x TemplateRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"name":     x.Name,
		"comment":  x.Comment,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateRevision decodes a stored record.
func ParseTemplateRevision(f codec.Fields) (TemplateRevision, error) {
	var x TemplateRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templaterevision: missing id")
	}
	x.Channel = f["channel"]
	x.Name = f["name"]
	x.Comment = f["comment"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("template: templaterevision %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templaterevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templaterevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templaterevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateRevision) String() string {
	return fmt.Sprintf("TemplateRevision(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateRevision) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateRevision) Normalize() TemplateRevision {
	x.Channel = strings.TrimSpace(x.Channel)
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

// SortTemplateRevisions orders by count descending, then by ID.
func SortTemplateRevisions(xs []TemplateRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateRevisions drops the archived entries, preserving order.
func ActiveTemplateRevisions(xs []TemplateRevision) []TemplateRevision {
	out := make([]TemplateRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateRevisions adds up the count of every active entry.
func SumTemplateRevisions(xs []TemplateRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTemplateRevisions keys the entries by ID, last write winning.
func IndexTemplateRevisions(xs []TemplateRevision) map[string]TemplateRevision {
	out := make(map[string]TemplateRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateRevision builds a valid TemplateRevision, for tests and for the seed data the cron
// binary writes.
func SampleTemplateRevision(id string, at time.Time) TemplateRevision {
	return TemplateRevision{
		ID:      id,
		Channel: "channel-" + id,
		Name:    "name-" + id,
		Comment: "seeded by template.SampleTemplateRevision",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TemplateSnapshotSchema is the stored shape of a TemplateSnapshot.
var TemplateSnapshotSchema = schema.New("template.templatesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateSnapshot is one template snapshot.
type TemplateSnapshot struct {
	ID       string
	Comment  string
	Source   string
	Channel  string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateSnapshot) Key() string { return "template/templatesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x TemplateSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"source":   x.Source,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateSnapshot decodes a stored record.
func ParseTemplateSnapshot(f codec.Fields) (TemplateSnapshot, error) {
	var x TemplateSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatesnapshot: missing id")
	}
	x.Comment = f["comment"]
	x.Source = f["source"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("template: templatesnapshot %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("template: templatesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateSnapshot) String() string {
	return fmt.Sprintf("TemplateSnapshot(%s, comment=%q, version=%d, archived=%t)", x.ID, x.Comment, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateSnapshot) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateSnapshot) Normalize() TemplateSnapshot {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTemplateSnapshots orders by version descending, then by ID.
func SortTemplateSnapshots(xs []TemplateSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateSnapshots drops the archived entries, preserving order.
func ActiveTemplateSnapshots(xs []TemplateSnapshot) []TemplateSnapshot {
	out := make([]TemplateSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateSnapshots adds up the version of every active entry.
func SumTemplateSnapshots(xs []TemplateSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTemplateSnapshots keys the entries by ID, last write winning.
func IndexTemplateSnapshots(xs []TemplateSnapshot) map[string]TemplateSnapshot {
	out := make(map[string]TemplateSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateSnapshot builds a valid TemplateSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleTemplateSnapshot(id string, at time.Time) TemplateSnapshot {
	return TemplateSnapshot{
		ID:       id,
		Comment:  "comment-" + id,
		Source:   "source-" + id,
		Channel:  "seeded by template.SampleTemplateSnapshot",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateBatchSchema is the stored shape of a TemplateBatch.
var TemplateBatchSchema = schema.New("template.templatebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateBatch is one template batch.
type TemplateBatch struct {
	ID       string
	Category string
	Region   string
	Source   string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateBatch) Key() string { return "template/templatebatch/" + x.ID }

// Fields encodes x for storage.
func (x TemplateBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"region":   x.Region,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateBatch decodes a stored record.
func ParseTemplateBatch(f codec.Fields) (TemplateBatch, error) {
	var x TemplateBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatebatch: missing id")
	}
	x.Category = f["category"]
	x.Region = f["region"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("template: templatebatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("template: templatebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateBatch) String() string {
	return fmt.Sprintf("TemplateBatch(%s, category=%q, retained=%d, archived=%t)", x.ID, x.Category, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateBatch) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateBatch) Normalize() TemplateBatch {
	x.Category = strings.TrimSpace(x.Category)
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTemplateBatchs orders by retained descending, then by ID.
func SortTemplateBatchs(xs []TemplateBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateBatchs drops the archived entries, preserving order.
func ActiveTemplateBatchs(xs []TemplateBatch) []TemplateBatch {
	out := make([]TemplateBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateBatchs adds up the retained of every active entry.
func SumTemplateBatchs(xs []TemplateBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTemplateBatchs keys the entries by ID, last write winning.
func IndexTemplateBatchs(xs []TemplateBatch) map[string]TemplateBatch {
	out := make(map[string]TemplateBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateBatch builds a valid TemplateBatch, for tests and for the seed data the cron
// binary writes.
func SampleTemplateBatch(id string, at time.Time) TemplateBatch {
	return TemplateBatch{
		ID:       id,
		Category: "category-" + id,
		Region:   "region-" + id,
		Source:   "seeded by template.SampleTemplateBatch",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateRequestSchema is the stored shape of a TemplateRequest.
var TemplateRequestSchema = schema.New("template.templaterequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateRequest is one template request.
type TemplateRequest struct {
	ID       string
	Region   string
	Name     string
	Channel  string
	Capacity int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateRequest) Key() string { return "template/templaterequest/" + x.ID }

// Fields encodes x for storage.
func (x TemplateRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"channel":  x.Channel,
		"capacity": strconv.Itoa(x.Capacity),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateRequest decodes a stored record.
func ParseTemplateRequest(f codec.Fields) (TemplateRequest, error) {
	var x TemplateRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templaterequest: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("template: templaterequest %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templaterequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templaterequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templaterequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateRequest) String() string {
	return fmt.Sprintf("TemplateRequest(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateRequest) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateRequest) Normalize() TemplateRequest {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortTemplateRequests orders by capacity descending, then by ID.
func SortTemplateRequests(xs []TemplateRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateRequests drops the archived entries, preserving order.
func ActiveTemplateRequests(xs []TemplateRequest) []TemplateRequest {
	out := make([]TemplateRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateRequests adds up the capacity of every active entry.
func SumTemplateRequests(xs []TemplateRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTemplateRequests keys the entries by ID, last write winning.
func IndexTemplateRequests(xs []TemplateRequest) map[string]TemplateRequest {
	out := make(map[string]TemplateRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateRequest builds a valid TemplateRequest, for tests and for the seed data the cron
// binary writes.
func SampleTemplateRequest(id string, at time.Time) TemplateRequest {
	return TemplateRequest{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Channel:  "seeded by template.SampleTemplateRequest",
		Capacity: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateSummarySchema is the stored shape of a TemplateSummary.
var TemplateSummarySchema = schema.New("template.templatesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateSummary is one template summary.
type TemplateSummary struct {
	ID        string
	Name      string
	Reference string
	Source    string
	Retained  int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TemplateSummary) Key() string { return "template/templatesummary/" + x.ID }

// Fields encodes x for storage.
func (x TemplateSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"source":    x.Source,
		"retained":  strconv.Itoa(x.Retained),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateSummary decodes a stored record.
func ParseTemplateSummary(f codec.Fields) (TemplateSummary, error) {
	var x TemplateSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatesummary: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("template: templatesummary %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("template: templatesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateSummary) String() string {
	return fmt.Sprintf("TemplateSummary(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateSummary) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateSummary) Normalize() TemplateSummary {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortTemplateSummarys orders by retained descending, then by ID.
func SortTemplateSummarys(xs []TemplateSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateSummarys drops the archived entries, preserving order.
func ActiveTemplateSummarys(xs []TemplateSummary) []TemplateSummary {
	out := make([]TemplateSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateSummarys adds up the retained of every active entry.
func SumTemplateSummarys(xs []TemplateSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTemplateSummarys keys the entries by ID, last write winning.
func IndexTemplateSummarys(xs []TemplateSummary) map[string]TemplateSummary {
	out := make(map[string]TemplateSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateSummary builds a valid TemplateSummary, for tests and for the seed data the cron
// binary writes.
func SampleTemplateSummary(id string, at time.Time) TemplateSummary {
	return TemplateSummary{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by template.SampleTemplateSummary",
		Retained:  len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TemplatePolicySchema is the stored shape of a TemplatePolicy.
var TemplatePolicySchema = schema.New("template.templatepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplatePolicy is one template policy.
type TemplatePolicy struct {
	ID       string
	Label    string
	Group    string
	Category string
	Capacity int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplatePolicy) Key() string { return "template/templatepolicy/" + x.ID }

// Fields encodes x for storage.
func (x TemplatePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"group":    x.Group,
		"category": x.Category,
		"capacity": strconv.Itoa(x.Capacity),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplatePolicy decodes a stored record.
func ParseTemplatePolicy(f codec.Fields) (TemplatePolicy, error) {
	var x TemplatePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatepolicy: missing id")
	}
	x.Label = f["label"]
	x.Group = f["group"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("template: templatepolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("template: templatepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplatePolicy) String() string {
	return fmt.Sprintf("TemplatePolicy(%s, label=%q, capacity=%d, archived=%t)", x.ID, x.Label, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplatePolicy) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplatePolicy) Normalize() TemplatePolicy {
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortTemplatePolicys orders by capacity descending, then by ID.
func SortTemplatePolicys(xs []TemplatePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplatePolicys drops the archived entries, preserving order.
func ActiveTemplatePolicys(xs []TemplatePolicy) []TemplatePolicy {
	out := make([]TemplatePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplatePolicys adds up the capacity of every active entry.
func SumTemplatePolicys(xs []TemplatePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexTemplatePolicys keys the entries by ID, last write winning.
func IndexTemplatePolicys(xs []TemplatePolicy) map[string]TemplatePolicy {
	out := make(map[string]TemplatePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplatePolicy builds a valid TemplatePolicy, for tests and for the seed data the cron
// binary writes.
func SampleTemplatePolicy(id string, at time.Time) TemplatePolicy {
	return TemplatePolicy{
		ID:       id,
		Label:    "label-" + id,
		Group:    "group-" + id,
		Category: "seeded by template.SampleTemplatePolicy",
		Capacity: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateLinkSchema is the stored shape of a TemplateLink.
var TemplateLinkSchema = schema.New("template.templatelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateLink is one template link.
type TemplateLink struct {
	ID       string
	Region   string
	Label    string
	Name     string
	Attempts int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateLink) Key() string { return "template/templatelink/" + x.ID }

// Fields encodes x for storage.
func (x TemplateLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateLink decodes a stored record.
func ParseTemplateLink(f codec.Fields) (TemplateLink, error) {
	var x TemplateLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatelink: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templatelink %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templatelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateLink) String() string {
	return fmt.Sprintf("TemplateLink(%s, region=%q, attempts=%d, archived=%t)", x.ID, x.Region, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateLink) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateLink) Normalize() TemplateLink {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTemplateLinks orders by attempts descending, then by ID.
func SortTemplateLinks(xs []TemplateLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateLinks drops the archived entries, preserving order.
func ActiveTemplateLinks(xs []TemplateLink) []TemplateLink {
	out := make([]TemplateLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateLinks adds up the attempts of every active entry.
func SumTemplateLinks(xs []TemplateLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTemplateLinks keys the entries by ID, last write winning.
func IndexTemplateLinks(xs []TemplateLink) map[string]TemplateLink {
	out := make(map[string]TemplateLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateLink builds a valid TemplateLink, for tests and for the seed data the cron
// binary writes.
func SampleTemplateLink(id string, at time.Time) TemplateLink {
	return TemplateLink{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Name:     "seeded by template.SampleTemplateLink",
		Attempts: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateNoteSchema is the stored shape of a TemplateNote.
var TemplateNoteSchema = schema.New("template.templatenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateNote is one template note.
type TemplateNote struct {
	ID       string
	Region   string
	Label    string
	Channel  string
	Version  int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateNote) Key() string { return "template/templatenote/" + x.ID }

// Fields encodes x for storage.
func (x TemplateNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"label":    x.Label,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateNote decodes a stored record.
func ParseTemplateNote(f codec.Fields) (TemplateNote, error) {
	var x TemplateNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatenote: missing id")
	}
	x.Region = f["region"]
	x.Label = f["label"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("template: templatenote %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("template: templatenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateNote) String() string {
	return fmt.Sprintf("TemplateNote(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateNote) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateNote) Normalize() TemplateNote {
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortTemplateNotes orders by version descending, then by ID.
func SortTemplateNotes(xs []TemplateNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateNotes drops the archived entries, preserving order.
func ActiveTemplateNotes(xs []TemplateNote) []TemplateNote {
	out := make([]TemplateNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateNotes adds up the version of every active entry.
func SumTemplateNotes(xs []TemplateNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTemplateNotes keys the entries by ID, last write winning.
func IndexTemplateNotes(xs []TemplateNote) map[string]TemplateNote {
	out := make(map[string]TemplateNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateNote builds a valid TemplateNote, for tests and for the seed data the cron
// binary writes.
func SampleTemplateNote(id string, at time.Time) TemplateNote {
	return TemplateNote{
		ID:       id,
		Region:   "region-" + id,
		Label:    "label-" + id,
		Channel:  "seeded by template.SampleTemplateNote",
		Version:  len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateWindowSchema is the stored shape of a TemplateWindow.
var TemplateWindowSchema = schema.New("template.templatewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateWindow is one template window.
type TemplateWindow struct {
	ID       string
	Comment  string
	Owner    string
	Name     string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateWindow) Key() string { return "template/templatewindow/" + x.ID }

// Fields encodes x for storage.
func (x TemplateWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"name":     x.Name,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateWindow decodes a stored record.
func ParseTemplateWindow(f codec.Fields) (TemplateWindow, error) {
	var x TemplateWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatewindow: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Name = f["name"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("template: templatewindow %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("template: templatewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateWindow) String() string {
	return fmt.Sprintf("TemplateWindow(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateWindow) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateWindow) Normalize() TemplateWindow {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Name = strings.TrimSpace(x.Name)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortTemplateWindows orders by retained descending, then by ID.
func SortTemplateWindows(xs []TemplateWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateWindows drops the archived entries, preserving order.
func ActiveTemplateWindows(xs []TemplateWindow) []TemplateWindow {
	out := make([]TemplateWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateWindows adds up the retained of every active entry.
func SumTemplateWindows(xs []TemplateWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTemplateWindows keys the entries by ID, last write winning.
func IndexTemplateWindows(xs []TemplateWindow) map[string]TemplateWindow {
	out := make(map[string]TemplateWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateWindow builds a valid TemplateWindow, for tests and for the seed data the cron
// binary writes.
func SampleTemplateWindow(id string, at time.Time) TemplateWindow {
	return TemplateWindow{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Name:     "seeded by template.SampleTemplateWindow",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateQuotaSchema is the stored shape of a TemplateQuota.
var TemplateQuotaSchema = schema.New("template.templatequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateQuota is one template quota.
type TemplateQuota struct {
	ID       string
	Category string
	Owner    string
	Group    string
	Offset   int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateQuota) Key() string { return "template/templatequota/" + x.ID }

// Fields encodes x for storage.
func (x TemplateQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"owner":    x.Owner,
		"group":    x.Group,
		"offset":   strconv.Itoa(x.Offset),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateQuota decodes a stored record.
func ParseTemplateQuota(f codec.Fields) (TemplateQuota, error) {
	var x TemplateQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatequota: missing id")
	}
	x.Category = f["category"]
	x.Owner = f["owner"]
	x.Group = f["group"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("template: templatequota %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("template: templatequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateQuota) String() string {
	return fmt.Sprintf("TemplateQuota(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateQuota) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateQuota) Normalize() TemplateQuota {
	x.Category = strings.TrimSpace(x.Category)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortTemplateQuotas orders by offset descending, then by ID.
func SortTemplateQuotas(xs []TemplateQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateQuotas drops the archived entries, preserving order.
func ActiveTemplateQuotas(xs []TemplateQuota) []TemplateQuota {
	out := make([]TemplateQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateQuotas adds up the offset of every active entry.
func SumTemplateQuotas(xs []TemplateQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexTemplateQuotas keys the entries by ID, last write winning.
func IndexTemplateQuotas(xs []TemplateQuota) map[string]TemplateQuota {
	out := make(map[string]TemplateQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateQuota builds a valid TemplateQuota, for tests and for the seed data the cron
// binary writes.
func SampleTemplateQuota(id string, at time.Time) TemplateQuota {
	return TemplateQuota{
		ID:       id,
		Category: "category-" + id,
		Owner:    "owner-" + id,
		Group:    "seeded by template.SampleTemplateQuota",
		Offset:   len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateBucketSchema is the stored shape of a TemplateBucket.
var TemplateBucketSchema = schema.New("template.templatebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateBucket is one template bucket.
type TemplateBucket struct {
	ID        string
	Reference string
	Channel   string
	Region    string
	Priority  int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TemplateBucket) Key() string { return "template/templatebucket/" + x.ID }

// Fields encodes x for storage.
func (x TemplateBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"channel":   x.Channel,
		"region":    x.Region,
		"priority":  strconv.Itoa(x.Priority),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateBucket decodes a stored record.
func ParseTemplateBucket(f codec.Fields) (TemplateBucket, error) {
	var x TemplateBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatebucket: missing id")
	}
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("template: templatebucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templatebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateBucket) String() string {
	return fmt.Sprintf("TemplateBucket(%s, reference=%q, priority=%d, archived=%t)", x.ID, x.Reference, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateBucket) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateBucket) Normalize() TemplateBucket {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTemplateBuckets orders by priority descending, then by ID.
func SortTemplateBuckets(xs []TemplateBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateBuckets drops the archived entries, preserving order.
func ActiveTemplateBuckets(xs []TemplateBucket) []TemplateBucket {
	out := make([]TemplateBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateBuckets adds up the priority of every active entry.
func SumTemplateBuckets(xs []TemplateBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexTemplateBuckets keys the entries by ID, last write winning.
func IndexTemplateBuckets(xs []TemplateBucket) map[string]TemplateBucket {
	out := make(map[string]TemplateBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateBucket builds a valid TemplateBucket, for tests and for the seed data the cron
// binary writes.
func SampleTemplateBucket(id string, at time.Time) TemplateBucket {
	return TemplateBucket{
		ID:        id,
		Reference: "reference-" + id,
		Channel:   "channel-" + id,
		Region:    "seeded by template.SampleTemplateBucket",
		Priority:  len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TemplateLaneSchema is the stored shape of a TemplateLane.
var TemplateLaneSchema = schema.New("template.templatelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateLane is one template lane.
type TemplateLane struct {
	ID        string
	Reference string
	Region    string
	Group     string
	Weight    int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TemplateLane) Key() string { return "template/templatelane/" + x.ID }

// Fields encodes x for storage.
func (x TemplateLane) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"group":     x.Group,
		"weight":    strconv.Itoa(x.Weight),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateLane decodes a stored record.
func ParseTemplateLane(f codec.Fields) (TemplateLane, error) {
	var x TemplateLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatelane: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templatelane %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templatelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateLane) String() string {
	return fmt.Sprintf("TemplateLane(%s, reference=%q, weight=%d, archived=%t)", x.ID, x.Reference, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateLane) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateLane) Normalize() TemplateLane {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortTemplateLanes orders by weight descending, then by ID.
func SortTemplateLanes(xs []TemplateLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateLanes drops the archived entries, preserving order.
func ActiveTemplateLanes(xs []TemplateLane) []TemplateLane {
	out := make([]TemplateLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateLanes adds up the weight of every active entry.
func SumTemplateLanes(xs []TemplateLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTemplateLanes keys the entries by ID, last write winning.
func IndexTemplateLanes(xs []TemplateLane) map[string]TemplateLane {
	out := make(map[string]TemplateLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateLane builds a valid TemplateLane, for tests and for the seed data the cron
// binary writes.
func SampleTemplateLane(id string, at time.Time) TemplateLane {
	return TemplateLane{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Group:     "seeded by template.SampleTemplateLane",
		Weight:    len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TemplateSlotSchema is the stored shape of a TemplateSlot.
var TemplateSlotSchema = schema.New("template.templateslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateSlot is one template slot.
type TemplateSlot struct {
	ID       string
	Source   string
	Region   string
	Label    string
	Weight   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateSlot) Key() string { return "template/templateslot/" + x.ID }

// Fields encodes x for storage.
func (x TemplateSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"region":   x.Region,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateSlot decodes a stored record.
func ParseTemplateSlot(f codec.Fields) (TemplateSlot, error) {
	var x TemplateSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templateslot: missing id")
	}
	x.Source = f["source"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templateslot %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("template: templateslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templateslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templateslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateSlot) String() string {
	return fmt.Sprintf("TemplateSlot(%s, source=%q, weight=%d, archived=%t)", x.ID, x.Source, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateSlot) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateSlot) Normalize() TemplateSlot {
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTemplateSlots orders by weight descending, then by ID.
func SortTemplateSlots(xs []TemplateSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateSlots drops the archived entries, preserving order.
func ActiveTemplateSlots(xs []TemplateSlot) []TemplateSlot {
	out := make([]TemplateSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateSlots adds up the weight of every active entry.
func SumTemplateSlots(xs []TemplateSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTemplateSlots keys the entries by ID, last write winning.
func IndexTemplateSlots(xs []TemplateSlot) map[string]TemplateSlot {
	out := make(map[string]TemplateSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateSlot builds a valid TemplateSlot, for tests and for the seed data the cron
// binary writes.
func SampleTemplateSlot(id string, at time.Time) TemplateSlot {
	return TemplateSlot{
		ID:      id,
		Source:  "source-" + id,
		Region:  "region-" + id,
		Label:   "seeded by template.SampleTemplateSlot",
		Weight:  len(id),
		Version: len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TemplateTagSchema is the stored shape of a TemplateTag.
var TemplateTagSchema = schema.New("template.templatetag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateTag is one template tag.
type TemplateTag struct {
	ID       string
	Channel  string
	Label    string
	Group    string
	Count    int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateTag) Key() string { return "template/templatetag/" + x.ID }

// Fields encodes x for storage.
func (x TemplateTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"group":    x.Group,
		"count":    strconv.Itoa(x.Count),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateTag decodes a stored record.
func ParseTemplateTag(f codec.Fields) (TemplateTag, error) {
	var x TemplateTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatetag: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("template: templatetag %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templatetag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatetag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatetag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateTag) String() string {
	return fmt.Sprintf("TemplateTag(%s, channel=%q, count=%d, archived=%t)", x.ID, x.Channel, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateTag) Normalize() TemplateTag {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTemplateTags orders by count descending, then by ID.
func SortTemplateTags(xs []TemplateTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateTags drops the archived entries, preserving order.
func ActiveTemplateTags(xs []TemplateTag) []TemplateTag {
	out := make([]TemplateTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateTags adds up the count of every active entry.
func SumTemplateTags(xs []TemplateTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexTemplateTags keys the entries by ID, last write winning.
func IndexTemplateTags(xs []TemplateTag) map[string]TemplateTag {
	out := make(map[string]TemplateTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateTag builds a valid TemplateTag, for tests and for the seed data the cron
// binary writes.
func SampleTemplateTag(id string, at time.Time) TemplateTag {
	return TemplateTag{
		ID:      id,
		Channel: "channel-" + id,
		Label:   "label-" + id,
		Group:   "seeded by template.SampleTemplateTag",
		Count:   len(id),
		Weight:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// TemplateRuleSchema is the stored shape of a TemplateRule.
var TemplateRuleSchema = schema.New("template.templaterule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateRule is one template rule.
type TemplateRule struct {
	ID       string
	Channel  string
	Label    string
	Group    string
	Attempts int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateRule) Key() string { return "template/templaterule/" + x.ID }

// Fields encodes x for storage.
func (x TemplateRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"label":    x.Label,
		"group":    x.Group,
		"attempts": strconv.Itoa(x.Attempts),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateRule decodes a stored record.
func ParseTemplateRule(f codec.Fields) (TemplateRule, error) {
	var x TemplateRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templaterule: missing id")
	}
	x.Channel = f["channel"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templaterule %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("template: templaterule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templaterule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templaterule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateRule) String() string {
	return fmt.Sprintf("TemplateRule(%s, channel=%q, attempts=%d, archived=%t)", x.ID, x.Channel, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateRule) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateRule) Normalize() TemplateRule {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortTemplateRules orders by attempts descending, then by ID.
func SortTemplateRules(xs []TemplateRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateRules drops the archived entries, preserving order.
func ActiveTemplateRules(xs []TemplateRule) []TemplateRule {
	out := make([]TemplateRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateRules adds up the attempts of every active entry.
func SumTemplateRules(xs []TemplateRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexTemplateRules keys the entries by ID, last write winning.
func IndexTemplateRules(xs []TemplateRule) map[string]TemplateRule {
	out := make(map[string]TemplateRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateRule builds a valid TemplateRule, for tests and for the seed data the cron
// binary writes.
func SampleTemplateRule(id string, at time.Time) TemplateRule {
	return TemplateRule{
		ID:       id,
		Channel:  "channel-" + id,
		Label:    "label-" + id,
		Group:    "seeded by template.SampleTemplateRule",
		Attempts: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateHoldSchema is the stored shape of a TemplateHold.
var TemplateHoldSchema = schema.New("template.templatehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateHold is one template hold.
type TemplateHold struct {
	ID        string
	Group     string
	Reference string
	Source    string
	Weight    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TemplateHold) Key() string { return "template/templatehold/" + x.ID }

// Fields encodes x for storage.
func (x TemplateHold) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"reference": x.Reference,
		"source":    x.Source,
		"weight":    strconv.Itoa(x.Weight),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateHold decodes a stored record.
func ParseTemplateHold(f codec.Fields) (TemplateHold, error) {
	var x TemplateHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatehold: missing id")
	}
	x.Group = f["group"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templatehold %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("template: templatehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateHold) String() string {
	return fmt.Sprintf("TemplateHold(%s, group=%q, weight=%d, archived=%t)", x.ID, x.Group, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateHold) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateHold) Normalize() TemplateHold {
	x.Group = strings.TrimSpace(x.Group)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortTemplateHolds orders by weight descending, then by ID.
func SortTemplateHolds(xs []TemplateHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateHolds drops the archived entries, preserving order.
func ActiveTemplateHolds(xs []TemplateHold) []TemplateHold {
	out := make([]TemplateHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateHolds adds up the weight of every active entry.
func SumTemplateHolds(xs []TemplateHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexTemplateHolds keys the entries by ID, last write winning.
func IndexTemplateHolds(xs []TemplateHold) map[string]TemplateHold {
	out := make(map[string]TemplateHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateHold builds a valid TemplateHold, for tests and for the seed data the cron
// binary writes.
func SampleTemplateHold(id string, at time.Time) TemplateHold {
	return TemplateHold{
		ID:        id,
		Group:     "group-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by template.SampleTemplateHold",
		Weight:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// TemplateClaimSchema is the stored shape of a TemplateClaim.
var TemplateClaimSchema = schema.New("template.templateclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateClaim is one template claim.
type TemplateClaim struct {
	ID       string
	Source   string
	Comment  string
	Region   string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x TemplateClaim) Key() string { return "template/templateclaim/" + x.ID }

// Fields encodes x for storage.
func (x TemplateClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateClaim decodes a stored record.
func ParseTemplateClaim(f codec.Fields) (TemplateClaim, error) {
	var x TemplateClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templateclaim: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("template: templateclaim %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("template: templateclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templateclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templateclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateClaim) String() string {
	return fmt.Sprintf("TemplateClaim(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateClaim) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateClaim) Normalize() TemplateClaim {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortTemplateClaims orders by retained descending, then by ID.
func SortTemplateClaims(xs []TemplateClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateClaims drops the archived entries, preserving order.
func ActiveTemplateClaims(xs []TemplateClaim) []TemplateClaim {
	out := make([]TemplateClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateClaims adds up the retained of every active entry.
func SumTemplateClaims(xs []TemplateClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexTemplateClaims keys the entries by ID, last write winning.
func IndexTemplateClaims(xs []TemplateClaim) map[string]TemplateClaim {
	out := make(map[string]TemplateClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateClaim builds a valid TemplateClaim, for tests and for the seed data the cron
// binary writes.
func SampleTemplateClaim(id string, at time.Time) TemplateClaim {
	return TemplateClaim{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by template.SampleTemplateClaim",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// TemplateCursorSchema is the stored shape of a TemplateCursor.
var TemplateCursorSchema = schema.New("template.templatecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// TemplateCursor is one template cursor.
type TemplateCursor struct {
	ID        string
	Source    string
	Reference string
	Owner     string
	Version   int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x TemplateCursor) Key() string { return "template/templatecursor/" + x.ID }

// Fields encodes x for storage.
func (x TemplateCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"reference": x.Reference,
		"owner":     x.Owner,
		"version":   strconv.Itoa(x.Version),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseTemplateCursor decodes a stored record.
func ParseTemplateCursor(f codec.Fields) (TemplateCursor, error) {
	var x TemplateCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("template: templatecursor: missing id")
	}
	x.Source = f["source"]
	x.Reference = f["reference"]
	x.Owner = f["owner"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("template: templatecursor %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("template: templatecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("template: templatecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("template: templatecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x TemplateCursor) String() string {
	return fmt.Sprintf("TemplateCursor(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x TemplateCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x TemplateCursor) Normalize() TemplateCursor {
	x.Source = strings.TrimSpace(x.Source)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortTemplateCursors orders by version descending, then by ID.
func SortTemplateCursors(xs []TemplateCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveTemplateCursors drops the archived entries, preserving order.
func ActiveTemplateCursors(xs []TemplateCursor) []TemplateCursor {
	out := make([]TemplateCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumTemplateCursors adds up the version of every active entry.
func SumTemplateCursors(xs []TemplateCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexTemplateCursors keys the entries by ID, last write winning.
func IndexTemplateCursors(xs []TemplateCursor) map[string]TemplateCursor {
	out := make(map[string]TemplateCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleTemplateCursor builds a valid TemplateCursor, for tests and for the seed data the cron
// binary writes.
func SampleTemplateCursor(id string, at time.Time) TemplateCursor {
	return TemplateCursor{
		ID:        id,
		Source:    "source-" + id,
		Reference: "reference-" + id,
		Owner:     "seeded by template.SampleTemplateCursor",
		Version:   len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	TemplateRecordSchema,
	TemplateEntrySchema,
	TemplateItemSchema,
	TemplateDraftSchema,
	TemplateRevisionSchema,
	TemplateSnapshotSchema,
	TemplateBatchSchema,
	TemplateRequestSchema,
	TemplateSummarySchema,
	TemplatePolicySchema,
	TemplateLinkSchema,
	TemplateNoteSchema,
	TemplateWindowSchema,
	TemplateQuotaSchema,
	TemplateBucketSchema,
	TemplateLaneSchema,
	TemplateSlotSchema,
	TemplateTagSchema,
	TemplateRuleSchema,
	TemplateHoldSchema,
	TemplateClaimSchema,
	TemplateCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"template/templaterecord/",
	"template/templateentry/",
	"template/templateitem/",
	"template/templatedraft/",
	"template/templaterevision/",
	"template/templatesnapshot/",
	"template/templatebatch/",
	"template/templaterequest/",
	"template/templatesummary/",
	"template/templatepolicy/",
	"template/templatelink/",
	"template/templatenote/",
	"template/templatewindow/",
	"template/templatequota/",
	"template/templatebucket/",
	"template/templatelane/",
	"template/templateslot/",
	"template/templatetag/",
	"template/templaterule/",
	"template/templatehold/",
	"template/templateclaim/",
	"template/templatecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "template: " + strings.Join(names, ", ")
}
