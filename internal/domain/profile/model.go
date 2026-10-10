// Package profile holds the profile domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package profile

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// ProfileRecordSchema is the stored shape of a ProfileRecord.
var ProfileRecordSchema = schema.New("profile.profilerecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileRecord is one profile record.
type ProfileRecord struct {
	ID       string
	Source   string
	Comment  string
	Group    string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileRecord) Key() string { return "profile/profilerecord/" + x.ID }

// Fields encodes x for storage.
func (x ProfileRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"comment":  x.Comment,
		"group":    x.Group,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileRecord decodes a stored record.
func ParseProfileRecord(f codec.Fields) (ProfileRecord, error) {
	var x ProfileRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilerecord: missing id")
	}
	x.Source = f["source"]
	x.Comment = f["comment"]
	x.Group = f["group"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("profile: profilerecord %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilerecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilerecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilerecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileRecord) String() string {
	return fmt.Sprintf("ProfileRecord(%s, source=%q, capacity=%d, archived=%t)", x.ID, x.Source, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileRecord) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileRecord) Normalize() ProfileRecord {
	x.Source = strings.TrimSpace(x.Source)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Group = strings.TrimSpace(x.Group)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortProfileRecords orders by capacity descending, then by ID.
func SortProfileRecords(xs []ProfileRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileRecords drops the archived entries, preserving order.
func ActiveProfileRecords(xs []ProfileRecord) []ProfileRecord {
	out := make([]ProfileRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileRecords adds up the capacity of every active entry.
func SumProfileRecords(xs []ProfileRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexProfileRecords keys the entries by ID, last write winning.
func IndexProfileRecords(xs []ProfileRecord) map[string]ProfileRecord {
	out := make(map[string]ProfileRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileRecord builds a valid ProfileRecord, for tests and for the seed data the cron
// binary writes.
func SampleProfileRecord(id string, at time.Time) ProfileRecord {
	return ProfileRecord{
		ID:       id,
		Source:   "source-" + id,
		Comment:  "comment-" + id,
		Group:    "seeded by profile.SampleProfileRecord",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileEntrySchema is the stored shape of a ProfileEntry.
var ProfileEntrySchema = schema.New("profile.profileentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileEntry is one profile entry.
type ProfileEntry struct {
	ID       string
	Category string
	Label    string
	Region   string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileEntry) Key() string { return "profile/profileentry/" + x.ID }

// Fields encodes x for storage.
func (x ProfileEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"label":    x.Label,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileEntry decodes a stored record.
func ParseProfileEntry(f codec.Fields) (ProfileEntry, error) {
	var x ProfileEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profileentry: missing id")
	}
	x.Category = f["category"]
	x.Label = f["label"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profileentry %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("profile: profileentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profileentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profileentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileEntry) String() string {
	return fmt.Sprintf("ProfileEntry(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileEntry) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileEntry) Normalize() ProfileEntry {
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortProfileEntrys orders by offset descending, then by ID.
func SortProfileEntrys(xs []ProfileEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileEntrys drops the archived entries, preserving order.
func ActiveProfileEntrys(xs []ProfileEntry) []ProfileEntry {
	out := make([]ProfileEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileEntrys adds up the offset of every active entry.
func SumProfileEntrys(xs []ProfileEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexProfileEntrys keys the entries by ID, last write winning.
func IndexProfileEntrys(xs []ProfileEntry) map[string]ProfileEntry {
	out := make(map[string]ProfileEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileEntry builds a valid ProfileEntry, for tests and for the seed data the cron
// binary writes.
func SampleProfileEntry(id string, at time.Time) ProfileEntry {
	return ProfileEntry{
		ID:       id,
		Category: "category-" + id,
		Label:    "label-" + id,
		Region:   "seeded by profile.SampleProfileEntry",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileItemSchema is the stored shape of a ProfileItem.
var ProfileItemSchema = schema.New("profile.profileitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileItem is one profile item.
type ProfileItem struct {
	ID       string
	Name     string
	Source   string
	Region   string
	Offset   int
	Capacity int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileItem) Key() string { return "profile/profileitem/" + x.ID }

// Fields encodes x for storage.
func (x ProfileItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"source":   x.Source,
		"region":   x.Region,
		"offset":   strconv.Itoa(x.Offset),
		"capacity": strconv.Itoa(x.Capacity),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileItem decodes a stored record.
func ParseProfileItem(f codec.Fields) (ProfileItem, error) {
	var x ProfileItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profileitem: missing id")
	}
	x.Name = f["name"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profileitem %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("profile: profileitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profileitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profileitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileItem) String() string {
	return fmt.Sprintf("ProfileItem(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileItem) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileItem) Normalize() ProfileItem {
	x.Name = strings.TrimSpace(x.Name)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortProfileItems orders by offset descending, then by ID.
func SortProfileItems(xs []ProfileItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileItems drops the archived entries, preserving order.
func ActiveProfileItems(xs []ProfileItem) []ProfileItem {
	out := make([]ProfileItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileItems adds up the offset of every active entry.
func SumProfileItems(xs []ProfileItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexProfileItems keys the entries by ID, last write winning.
func IndexProfileItems(xs []ProfileItem) map[string]ProfileItem {
	out := make(map[string]ProfileItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileItem builds a valid ProfileItem, for tests and for the seed data the cron
// binary writes.
func SampleProfileItem(id string, at time.Time) ProfileItem {
	return ProfileItem{
		ID:       id,
		Name:     "name-" + id,
		Source:   "source-" + id,
		Region:   "seeded by profile.SampleProfileItem",
		Offset:   len(id),
		Capacity: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileDraftSchema is the stored shape of a ProfileDraft.
var ProfileDraftSchema = schema.New("profile.profiledraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileDraft is one profile draft.
type ProfileDraft struct {
	ID       string
	Owner    string
	Comment  string
	Region   string
	Attempts int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileDraft) Key() string { return "profile/profiledraft/" + x.ID }

// Fields encodes x for storage.
func (x ProfileDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"comment":  x.Comment,
		"region":   x.Region,
		"attempts": strconv.Itoa(x.Attempts),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileDraft decodes a stored record.
func ParseProfileDraft(f codec.Fields) (ProfileDraft, error) {
	var x ProfileDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profiledraft: missing id")
	}
	x.Owner = f["owner"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("profile: profiledraft %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profiledraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profiledraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profiledraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileDraft) String() string {
	return fmt.Sprintf("ProfileDraft(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileDraft) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileDraft) Normalize() ProfileDraft {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortProfileDrafts orders by attempts descending, then by ID.
func SortProfileDrafts(xs []ProfileDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileDrafts drops the archived entries, preserving order.
func ActiveProfileDrafts(xs []ProfileDraft) []ProfileDraft {
	out := make([]ProfileDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileDrafts adds up the attempts of every active entry.
func SumProfileDrafts(xs []ProfileDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexProfileDrafts keys the entries by ID, last write winning.
func IndexProfileDrafts(xs []ProfileDraft) map[string]ProfileDraft {
	out := make(map[string]ProfileDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileDraft builds a valid ProfileDraft, for tests and for the seed data the cron
// binary writes.
func SampleProfileDraft(id string, at time.Time) ProfileDraft {
	return ProfileDraft{
		ID:       id,
		Owner:    "owner-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by profile.SampleProfileDraft",
		Attempts: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileRevisionSchema is the stored shape of a ProfileRevision.
var ProfileRevisionSchema = schema.New("profile.profilerevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileRevision is one profile revision.
type ProfileRevision struct {
	ID       string
	Label    string
	Source   string
	Region   string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileRevision) Key() string { return "profile/profilerevision/" + x.ID }

// Fields encodes x for storage.
func (x ProfileRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"source":   x.Source,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileRevision decodes a stored record.
func ParseProfileRevision(f codec.Fields) (ProfileRevision, error) {
	var x ProfileRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilerevision: missing id")
	}
	x.Label = f["label"]
	x.Source = f["source"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilerevision %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("profile: profilerevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilerevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilerevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileRevision) String() string {
	return fmt.Sprintf("ProfileRevision(%s, label=%q, retained=%d, archived=%t)", x.ID, x.Label, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileRevision) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileRevision) Normalize() ProfileRevision {
	x.Label = strings.TrimSpace(x.Label)
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortProfileRevisions orders by retained descending, then by ID.
func SortProfileRevisions(xs []ProfileRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileRevisions drops the archived entries, preserving order.
func ActiveProfileRevisions(xs []ProfileRevision) []ProfileRevision {
	out := make([]ProfileRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileRevisions adds up the retained of every active entry.
func SumProfileRevisions(xs []ProfileRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexProfileRevisions keys the entries by ID, last write winning.
func IndexProfileRevisions(xs []ProfileRevision) map[string]ProfileRevision {
	out := make(map[string]ProfileRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileRevision builds a valid ProfileRevision, for tests and for the seed data the cron
// binary writes.
func SampleProfileRevision(id string, at time.Time) ProfileRevision {
	return ProfileRevision{
		ID:       id,
		Label:    "label-" + id,
		Source:   "source-" + id,
		Region:   "seeded by profile.SampleProfileRevision",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileSnapshotSchema is the stored shape of a ProfileSnapshot.
var ProfileSnapshotSchema = schema.New("profile.profilesnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileSnapshot is one profile snapshot.
type ProfileSnapshot struct {
	ID        string
	Label     string
	Reference string
	Group     string
	Priority  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ProfileSnapshot) Key() string { return "profile/profilesnapshot/" + x.ID }

// Fields encodes x for storage.
func (x ProfileSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"reference": x.Reference,
		"group":     x.Group,
		"priority":  strconv.Itoa(x.Priority),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileSnapshot decodes a stored record.
func ParseProfileSnapshot(f codec.Fields) (ProfileSnapshot, error) {
	var x ProfileSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilesnapshot: missing id")
	}
	x.Label = f["label"]
	x.Reference = f["reference"]
	x.Group = f["group"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("profile: profilesnapshot %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilesnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilesnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilesnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileSnapshot) String() string {
	return fmt.Sprintf("ProfileSnapshot(%s, label=%q, priority=%d, archived=%t)", x.ID, x.Label, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileSnapshot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileSnapshot) Normalize() ProfileSnapshot {
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Group = strings.TrimSpace(x.Group)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortProfileSnapshots orders by priority descending, then by ID.
func SortProfileSnapshots(xs []ProfileSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileSnapshots drops the archived entries, preserving order.
func ActiveProfileSnapshots(xs []ProfileSnapshot) []ProfileSnapshot {
	out := make([]ProfileSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileSnapshots adds up the priority of every active entry.
func SumProfileSnapshots(xs []ProfileSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexProfileSnapshots keys the entries by ID, last write winning.
func IndexProfileSnapshots(xs []ProfileSnapshot) map[string]ProfileSnapshot {
	out := make(map[string]ProfileSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileSnapshot builds a valid ProfileSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleProfileSnapshot(id string, at time.Time) ProfileSnapshot {
	return ProfileSnapshot{
		ID:        id,
		Label:     "label-" + id,
		Reference: "reference-" + id,
		Group:     "seeded by profile.SampleProfileSnapshot",
		Priority:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ProfileBatchSchema is the stored shape of a ProfileBatch.
var ProfileBatchSchema = schema.New("profile.profilebatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileBatch is one profile batch.
type ProfileBatch struct {
	ID       string
	Category string
	Channel  string
	Region   string
	Weight   int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileBatch) Key() string { return "profile/profilebatch/" + x.ID }

// Fields encodes x for storage.
func (x ProfileBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"region":   x.Region,
		"weight":   strconv.Itoa(x.Weight),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileBatch decodes a stored record.
func ParseProfileBatch(f codec.Fields) (ProfileBatch, error) {
	var x ProfileBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilebatch: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("profile: profilebatch %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("profile: profilebatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilebatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilebatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileBatch) String() string {
	return fmt.Sprintf("ProfileBatch(%s, category=%q, weight=%d, archived=%t)", x.ID, x.Category, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileBatch) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileBatch) Normalize() ProfileBatch {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortProfileBatchs orders by weight descending, then by ID.
func SortProfileBatchs(xs []ProfileBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileBatchs drops the archived entries, preserving order.
func ActiveProfileBatchs(xs []ProfileBatch) []ProfileBatch {
	out := make([]ProfileBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileBatchs adds up the weight of every active entry.
func SumProfileBatchs(xs []ProfileBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexProfileBatchs keys the entries by ID, last write winning.
func IndexProfileBatchs(xs []ProfileBatch) map[string]ProfileBatch {
	out := make(map[string]ProfileBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileBatch builds a valid ProfileBatch, for tests and for the seed data the cron
// binary writes.
func SampleProfileBatch(id string, at time.Time) ProfileBatch {
	return ProfileBatch{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by profile.SampleProfileBatch",
		Weight:   len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileRequestSchema is the stored shape of a ProfileRequest.
var ProfileRequestSchema = schema.New("profile.profilerequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileRequest is one profile request.
type ProfileRequest struct {
	ID        string
	Group     string
	Region    string
	Reference string
	Offset    int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ProfileRequest) Key() string { return "profile/profilerequest/" + x.ID }

// Fields encodes x for storage.
func (x ProfileRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"region":    x.Region,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileRequest decodes a stored record.
func ParseProfileRequest(f codec.Fields) (ProfileRequest, error) {
	var x ProfileRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilerequest: missing id")
	}
	x.Group = f["group"]
	x.Region = f["region"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profilerequest %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilerequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilerequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilerequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileRequest) String() string {
	return fmt.Sprintf("ProfileRequest(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileRequest) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileRequest) Normalize() ProfileRequest {
	x.Group = strings.TrimSpace(x.Group)
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortProfileRequests orders by offset descending, then by ID.
func SortProfileRequests(xs []ProfileRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileRequests drops the archived entries, preserving order.
func ActiveProfileRequests(xs []ProfileRequest) []ProfileRequest {
	out := make([]ProfileRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileRequests adds up the offset of every active entry.
func SumProfileRequests(xs []ProfileRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexProfileRequests keys the entries by ID, last write winning.
func IndexProfileRequests(xs []ProfileRequest) map[string]ProfileRequest {
	out := make(map[string]ProfileRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileRequest builds a valid ProfileRequest, for tests and for the seed data the cron
// binary writes.
func SampleProfileRequest(id string, at time.Time) ProfileRequest {
	return ProfileRequest{
		ID:        id,
		Group:     "group-" + id,
		Region:    "region-" + id,
		Reference: "seeded by profile.SampleProfileRequest",
		Offset:    len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ProfileSummarySchema is the stored shape of a ProfileSummary.
var ProfileSummarySchema = schema.New("profile.profilesummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileSummary is one profile summary.
type ProfileSummary struct {
	ID       string
	Category string
	Channel  string
	Region   string
	Capacity int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileSummary) Key() string { return "profile/profilesummary/" + x.ID }

// Fields encodes x for storage.
func (x ProfileSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"region":   x.Region,
		"capacity": strconv.Itoa(x.Capacity),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileSummary decodes a stored record.
func ParseProfileSummary(f codec.Fields) (ProfileSummary, error) {
	var x ProfileSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilesummary: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("profile: profilesummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilesummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilesummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilesummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileSummary) String() string {
	return fmt.Sprintf("ProfileSummary(%s, category=%q, capacity=%d, archived=%t)", x.ID, x.Category, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileSummary) Normalize() ProfileSummary {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortProfileSummarys orders by capacity descending, then by ID.
func SortProfileSummarys(xs []ProfileSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileSummarys drops the archived entries, preserving order.
func ActiveProfileSummarys(xs []ProfileSummary) []ProfileSummary {
	out := make([]ProfileSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileSummarys adds up the capacity of every active entry.
func SumProfileSummarys(xs []ProfileSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexProfileSummarys keys the entries by ID, last write winning.
func IndexProfileSummarys(xs []ProfileSummary) map[string]ProfileSummary {
	out := make(map[string]ProfileSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileSummary builds a valid ProfileSummary, for tests and for the seed data the cron
// binary writes.
func SampleProfileSummary(id string, at time.Time) ProfileSummary {
	return ProfileSummary{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by profile.SampleProfileSummary",
		Capacity: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfilePolicySchema is the stored shape of a ProfilePolicy.
var ProfilePolicySchema = schema.New("profile.profilepolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfilePolicy is one profile policy.
type ProfilePolicy struct {
	ID        string
	Name      string
	Reference string
	Source    string
	Offset    int
	Capacity  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x ProfilePolicy) Key() string { return "profile/profilepolicy/" + x.ID }

// Fields encodes x for storage.
func (x ProfilePolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"name":      x.Name,
		"reference": x.Reference,
		"source":    x.Source,
		"offset":    strconv.Itoa(x.Offset),
		"capacity":  strconv.Itoa(x.Capacity),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfilePolicy decodes a stored record.
func ParseProfilePolicy(f codec.Fields) (ProfilePolicy, error) {
	var x ProfilePolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilepolicy: missing id")
	}
	x.Name = f["name"]
	x.Reference = f["reference"]
	x.Source = f["source"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profilepolicy %s: %w", x.ID, err)
	}
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("profile: profilepolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilepolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilepolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfilePolicy) String() string {
	return fmt.Sprintf("ProfilePolicy(%s, name=%q, offset=%d, archived=%t)", x.ID, x.Name, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfilePolicy) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfilePolicy) Normalize() ProfilePolicy {
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Source = strings.TrimSpace(x.Source)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	return x
}

// SortProfilePolicys orders by offset descending, then by ID.
func SortProfilePolicys(xs []ProfilePolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfilePolicys drops the archived entries, preserving order.
func ActiveProfilePolicys(xs []ProfilePolicy) []ProfilePolicy {
	out := make([]ProfilePolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfilePolicys adds up the offset of every active entry.
func SumProfilePolicys(xs []ProfilePolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexProfilePolicys keys the entries by ID, last write winning.
func IndexProfilePolicys(xs []ProfilePolicy) map[string]ProfilePolicy {
	out := make(map[string]ProfilePolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfilePolicy builds a valid ProfilePolicy, for tests and for the seed data the cron
// binary writes.
func SampleProfilePolicy(id string, at time.Time) ProfilePolicy {
	return ProfilePolicy{
		ID:        id,
		Name:      "name-" + id,
		Reference: "reference-" + id,
		Source:    "seeded by profile.SampleProfilePolicy",
		Offset:    len(id),
		Capacity:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// ProfileLinkSchema is the stored shape of a ProfileLink.
var ProfileLinkSchema = schema.New("profile.profilelink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileLink is one profile link.
type ProfileLink struct {
	ID       string
	Comment  string
	Category string
	Label    string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileLink) Key() string { return "profile/profilelink/" + x.ID }

// Fields encodes x for storage.
func (x ProfileLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"label":    x.Label,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileLink decodes a stored record.
func ParseProfileLink(f codec.Fields) (ProfileLink, error) {
	var x ProfileLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilelink: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Label = f["label"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilelink %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profilelink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilelink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilelink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileLink) String() string {
	return fmt.Sprintf("ProfileLink(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileLink) Normalize() ProfileLink {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Label = strings.TrimSpace(x.Label)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortProfileLinks orders by retained descending, then by ID.
func SortProfileLinks(xs []ProfileLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileLinks drops the archived entries, preserving order.
func ActiveProfileLinks(xs []ProfileLink) []ProfileLink {
	out := make([]ProfileLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileLinks adds up the retained of every active entry.
func SumProfileLinks(xs []ProfileLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexProfileLinks keys the entries by ID, last write winning.
func IndexProfileLinks(xs []ProfileLink) map[string]ProfileLink {
	out := make(map[string]ProfileLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileLink builds a valid ProfileLink, for tests and for the seed data the cron
// binary writes.
func SampleProfileLink(id string, at time.Time) ProfileLink {
	return ProfileLink{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Label:    "seeded by profile.SampleProfileLink",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileNoteSchema is the stored shape of a ProfileNote.
var ProfileNoteSchema = schema.New("profile.profilenote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileNote is one profile note.
type ProfileNote struct {
	ID       string
	Source   string
	Channel  string
	Region   string
	Attempts int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileNote) Key() string { return "profile/profilenote/" + x.ID }

// Fields encodes x for storage.
func (x ProfileNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"channel":  x.Channel,
		"region":   x.Region,
		"attempts": strconv.Itoa(x.Attempts),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileNote decodes a stored record.
func ParseProfileNote(f codec.Fields) (ProfileNote, error) {
	var x ProfileNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilenote: missing id")
	}
	x.Source = f["source"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("profile: profilenote %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("profile: profilenote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilenote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilenote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileNote) String() string {
	return fmt.Sprintf("ProfileNote(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileNote) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileNote) Normalize() ProfileNote {
	x.Source = strings.TrimSpace(x.Source)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortProfileNotes orders by attempts descending, then by ID.
func SortProfileNotes(xs []ProfileNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileNotes drops the archived entries, preserving order.
func ActiveProfileNotes(xs []ProfileNote) []ProfileNote {
	out := make([]ProfileNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileNotes adds up the attempts of every active entry.
func SumProfileNotes(xs []ProfileNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexProfileNotes keys the entries by ID, last write winning.
func IndexProfileNotes(xs []ProfileNote) map[string]ProfileNote {
	out := make(map[string]ProfileNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileNote builds a valid ProfileNote, for tests and for the seed data the cron
// binary writes.
func SampleProfileNote(id string, at time.Time) ProfileNote {
	return ProfileNote{
		ID:       id,
		Source:   "source-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by profile.SampleProfileNote",
		Attempts: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileWindowSchema is the stored shape of a ProfileWindow.
var ProfileWindowSchema = schema.New("profile.profilewindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileWindow is one profile window.
type ProfileWindow struct {
	ID       string
	Comment  string
	Channel  string
	Name     string
	Offset   int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileWindow) Key() string { return "profile/profilewindow/" + x.ID }

// Fields encodes x for storage.
func (x ProfileWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"channel":  x.Channel,
		"name":     x.Name,
		"offset":   strconv.Itoa(x.Offset),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileWindow decodes a stored record.
func ParseProfileWindow(f codec.Fields) (ProfileWindow, error) {
	var x ProfileWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilewindow: missing id")
	}
	x.Comment = f["comment"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profilewindow %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("profile: profilewindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilewindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilewindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileWindow) String() string {
	return fmt.Sprintf("ProfileWindow(%s, comment=%q, offset=%d, archived=%t)", x.ID, x.Comment, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileWindow) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileWindow) Normalize() ProfileWindow {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortProfileWindows orders by offset descending, then by ID.
func SortProfileWindows(xs []ProfileWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileWindows drops the archived entries, preserving order.
func ActiveProfileWindows(xs []ProfileWindow) []ProfileWindow {
	out := make([]ProfileWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileWindows adds up the offset of every active entry.
func SumProfileWindows(xs []ProfileWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexProfileWindows keys the entries by ID, last write winning.
func IndexProfileWindows(xs []ProfileWindow) map[string]ProfileWindow {
	out := make(map[string]ProfileWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileWindow builds a valid ProfileWindow, for tests and for the seed data the cron
// binary writes.
func SampleProfileWindow(id string, at time.Time) ProfileWindow {
	return ProfileWindow{
		ID:       id,
		Comment:  "comment-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by profile.SampleProfileWindow",
		Offset:   len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileQuotaSchema is the stored shape of a ProfileQuota.
var ProfileQuotaSchema = schema.New("profile.profilequota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileQuota is one profile quota.
type ProfileQuota struct {
	ID       string
	Comment  string
	Category string
	Group    string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileQuota) Key() string { return "profile/profilequota/" + x.ID }

// Fields encodes x for storage.
func (x ProfileQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"category": x.Category,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileQuota decodes a stored record.
func ParseProfileQuota(f codec.Fields) (ProfileQuota, error) {
	var x ProfileQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilequota: missing id")
	}
	x.Comment = f["comment"]
	x.Category = f["category"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilequota %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("profile: profilequota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilequota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilequota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileQuota) String() string {
	return fmt.Sprintf("ProfileQuota(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileQuota) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileQuota) Normalize() ProfileQuota {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Category = strings.TrimSpace(x.Category)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortProfileQuotas orders by retained descending, then by ID.
func SortProfileQuotas(xs []ProfileQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileQuotas drops the archived entries, preserving order.
func ActiveProfileQuotas(xs []ProfileQuota) []ProfileQuota {
	out := make([]ProfileQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileQuotas adds up the retained of every active entry.
func SumProfileQuotas(xs []ProfileQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexProfileQuotas keys the entries by ID, last write winning.
func IndexProfileQuotas(xs []ProfileQuota) map[string]ProfileQuota {
	out := make(map[string]ProfileQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileQuota builds a valid ProfileQuota, for tests and for the seed data the cron
// binary writes.
func SampleProfileQuota(id string, at time.Time) ProfileQuota {
	return ProfileQuota{
		ID:       id,
		Comment:  "comment-" + id,
		Category: "category-" + id,
		Group:    "seeded by profile.SampleProfileQuota",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileBucketSchema is the stored shape of a ProfileBucket.
var ProfileBucketSchema = schema.New("profile.profilebucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileBucket is one profile bucket.
type ProfileBucket struct {
	ID       string
	Comment  string
	Owner    string
	Source   string
	Retained int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileBucket) Key() string { return "profile/profilebucket/" + x.ID }

// Fields encodes x for storage.
func (x ProfileBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"comment":  x.Comment,
		"owner":    x.Owner,
		"source":   x.Source,
		"retained": strconv.Itoa(x.Retained),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileBucket decodes a stored record.
func ParseProfileBucket(f codec.Fields) (ProfileBucket, error) {
	var x ProfileBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilebucket: missing id")
	}
	x.Comment = f["comment"]
	x.Owner = f["owner"]
	x.Source = f["source"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilebucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("profile: profilebucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilebucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilebucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileBucket) String() string {
	return fmt.Sprintf("ProfileBucket(%s, comment=%q, retained=%d, archived=%t)", x.ID, x.Comment, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileBucket) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileBucket) Normalize() ProfileBucket {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Source = strings.TrimSpace(x.Source)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortProfileBuckets orders by retained descending, then by ID.
func SortProfileBuckets(xs []ProfileBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileBuckets drops the archived entries, preserving order.
func ActiveProfileBuckets(xs []ProfileBucket) []ProfileBucket {
	out := make([]ProfileBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileBuckets adds up the retained of every active entry.
func SumProfileBuckets(xs []ProfileBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexProfileBuckets keys the entries by ID, last write winning.
func IndexProfileBuckets(xs []ProfileBucket) map[string]ProfileBucket {
	out := make(map[string]ProfileBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileBucket builds a valid ProfileBucket, for tests and for the seed data the cron
// binary writes.
func SampleProfileBucket(id string, at time.Time) ProfileBucket {
	return ProfileBucket{
		ID:       id,
		Comment:  "comment-" + id,
		Owner:    "owner-" + id,
		Source:   "seeded by profile.SampleProfileBucket",
		Retained: len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileLaneSchema is the stored shape of a ProfileLane.
var ProfileLaneSchema = schema.New("profile.profilelane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileLane is one profile lane.
type ProfileLane struct {
	ID       string
	Source   string
	Label    string
	Name     string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileLane) Key() string { return "profile/profilelane/" + x.ID }

// Fields encodes x for storage.
func (x ProfileLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"label":    x.Label,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileLane decodes a stored record.
func ParseProfileLane(f codec.Fields) (ProfileLane, error) {
	var x ProfileLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilelane: missing id")
	}
	x.Source = f["source"]
	x.Label = f["label"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("profile: profilelane %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("profile: profilelane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilelane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilelane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileLane) String() string {
	return fmt.Sprintf("ProfileLane(%s, source=%q, attempts=%d, archived=%t)", x.ID, x.Source, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileLane) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileLane) Normalize() ProfileLane {
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortProfileLanes orders by attempts descending, then by ID.
func SortProfileLanes(xs []ProfileLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileLanes drops the archived entries, preserving order.
func ActiveProfileLanes(xs []ProfileLane) []ProfileLane {
	out := make([]ProfileLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileLanes adds up the attempts of every active entry.
func SumProfileLanes(xs []ProfileLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexProfileLanes keys the entries by ID, last write winning.
func IndexProfileLanes(xs []ProfileLane) map[string]ProfileLane {
	out := make(map[string]ProfileLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileLane builds a valid ProfileLane, for tests and for the seed data the cron
// binary writes.
func SampleProfileLane(id string, at time.Time) ProfileLane {
	return ProfileLane{
		ID:       id,
		Source:   "source-" + id,
		Label:    "label-" + id,
		Name:     "seeded by profile.SampleProfileLane",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileSlotSchema is the stored shape of a ProfileSlot.
var ProfileSlotSchema = schema.New("profile.profileslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileSlot is one profile slot.
type ProfileSlot struct {
	ID       string
	Source   string
	Region   string
	Owner    string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileSlot) Key() string { return "profile/profileslot/" + x.ID }

// Fields encodes x for storage.
func (x ProfileSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"region":   x.Region,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileSlot decodes a stored record.
func ParseProfileSlot(f codec.Fields) (ProfileSlot, error) {
	var x ProfileSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profileslot: missing id")
	}
	x.Source = f["source"]
	x.Region = f["region"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profileslot %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("profile: profileslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profileslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profileslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileSlot) String() string {
	return fmt.Sprintf("ProfileSlot(%s, source=%q, retained=%d, archived=%t)", x.ID, x.Source, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileSlot) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileSlot) Normalize() ProfileSlot {
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortProfileSlots orders by retained descending, then by ID.
func SortProfileSlots(xs []ProfileSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileSlots drops the archived entries, preserving order.
func ActiveProfileSlots(xs []ProfileSlot) []ProfileSlot {
	out := make([]ProfileSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileSlots adds up the retained of every active entry.
func SumProfileSlots(xs []ProfileSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexProfileSlots keys the entries by ID, last write winning.
func IndexProfileSlots(xs []ProfileSlot) map[string]ProfileSlot {
	out := make(map[string]ProfileSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileSlot builds a valid ProfileSlot, for tests and for the seed data the cron
// binary writes.
func SampleProfileSlot(id string, at time.Time) ProfileSlot {
	return ProfileSlot{
		ID:       id,
		Source:   "source-" + id,
		Region:   "region-" + id,
		Owner:    "seeded by profile.SampleProfileSlot",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileTagSchema is the stored shape of a ProfileTag.
var ProfileTagSchema = schema.New("profile.profiletag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileTag is one profile tag.
type ProfileTag struct {
	ID       string
	Source   string
	Region   string
	Label    string
	Count    int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileTag) Key() string { return "profile/profiletag/" + x.ID }

// Fields encodes x for storage.
func (x ProfileTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"region":   x.Region,
		"label":    x.Label,
		"count":    strconv.Itoa(x.Count),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileTag decodes a stored record.
func ParseProfileTag(f codec.Fields) (ProfileTag, error) {
	var x ProfileTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profiletag: missing id")
	}
	x.Source = f["source"]
	x.Region = f["region"]
	x.Label = f["label"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("profile: profiletag %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("profile: profiletag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profiletag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profiletag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileTag) String() string {
	return fmt.Sprintf("ProfileTag(%s, source=%q, count=%d, archived=%t)", x.ID, x.Source, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileTag) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileTag) Normalize() ProfileTag {
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	x.Label = strings.TrimSpace(x.Label)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortProfileTags orders by count descending, then by ID.
func SortProfileTags(xs []ProfileTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileTags drops the archived entries, preserving order.
func ActiveProfileTags(xs []ProfileTag) []ProfileTag {
	out := make([]ProfileTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileTags adds up the count of every active entry.
func SumProfileTags(xs []ProfileTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexProfileTags keys the entries by ID, last write winning.
func IndexProfileTags(xs []ProfileTag) map[string]ProfileTag {
	out := make(map[string]ProfileTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileTag builds a valid ProfileTag, for tests and for the seed data the cron
// binary writes.
func SampleProfileTag(id string, at time.Time) ProfileTag {
	return ProfileTag{
		ID:       id,
		Source:   "source-" + id,
		Region:   "region-" + id,
		Label:    "seeded by profile.SampleProfileTag",
		Count:    len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileRuleSchema is the stored shape of a ProfileRule.
var ProfileRuleSchema = schema.New("profile.profilerule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileRule is one profile rule.
type ProfileRule struct {
	ID       string
	Group    string
	Channel  string
	Label    string
	Weight   int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileRule) Key() string { return "profile/profilerule/" + x.ID }

// Fields encodes x for storage.
func (x ProfileRule) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"label":    x.Label,
		"weight":   strconv.Itoa(x.Weight),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileRule decodes a stored record.
func ParseProfileRule(f codec.Fields) (ProfileRule, error) {
	var x ProfileRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilerule: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Label = f["label"]
	var err error
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("profile: profilerule %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profilerule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilerule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilerule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileRule) String() string {
	return fmt.Sprintf("ProfileRule(%s, group=%q, weight=%d, archived=%t)", x.ID, x.Group, x.Weight, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileRule) Active() bool { return !x.Archived && x.Weight >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileRule) Normalize() ProfileRule {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Label = strings.TrimSpace(x.Label)
	if x.Weight < 0 {
		x.Weight = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortProfileRules orders by weight descending, then by ID.
func SortProfileRules(xs []ProfileRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Weight != xs[j].Weight {
			return xs[i].Weight > xs[j].Weight
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileRules drops the archived entries, preserving order.
func ActiveProfileRules(xs []ProfileRule) []ProfileRule {
	out := make([]ProfileRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileRules adds up the weight of every active entry.
func SumProfileRules(xs []ProfileRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Weight
		}
	}
	return total
}

// IndexProfileRules keys the entries by ID, last write winning.
func IndexProfileRules(xs []ProfileRule) map[string]ProfileRule {
	out := make(map[string]ProfileRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileRule builds a valid ProfileRule, for tests and for the seed data the cron
// binary writes.
func SampleProfileRule(id string, at time.Time) ProfileRule {
	return ProfileRule{
		ID:      id,
		Group:   "group-" + id,
		Channel: "channel-" + id,
		Label:   "seeded by profile.SampleProfileRule",
		Weight:  len(id),
		Offset:  len(id) * 2,
		Updated: at.UTC().Truncate(time.Second),
	}
}

// ProfileHoldSchema is the stored shape of a ProfileHold.
var ProfileHoldSchema = schema.New("profile.profilehold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileHold is one profile hold.
type ProfileHold struct {
	ID       string
	Group    string
	Channel  string
	Owner    string
	Retained int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileHold) Key() string { return "profile/profilehold/" + x.ID }

// Fields encodes x for storage.
func (x ProfileHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"channel":  x.Channel,
		"owner":    x.Owner,
		"retained": strconv.Itoa(x.Retained),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileHold decodes a stored record.
func ParseProfileHold(f codec.Fields) (ProfileHold, error) {
	var x ProfileHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilehold: missing id")
	}
	x.Group = f["group"]
	x.Channel = f["channel"]
	x.Owner = f["owner"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("profile: profilehold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("profile: profilehold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilehold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilehold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileHold) String() string {
	return fmt.Sprintf("ProfileHold(%s, group=%q, retained=%d, archived=%t)", x.ID, x.Group, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileHold) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileHold) Normalize() ProfileHold {
	x.Group = strings.TrimSpace(x.Group)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Owner = strings.TrimSpace(x.Owner)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortProfileHolds orders by retained descending, then by ID.
func SortProfileHolds(xs []ProfileHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileHolds drops the archived entries, preserving order.
func ActiveProfileHolds(xs []ProfileHold) []ProfileHold {
	out := make([]ProfileHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileHolds adds up the retained of every active entry.
func SumProfileHolds(xs []ProfileHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexProfileHolds keys the entries by ID, last write winning.
func IndexProfileHolds(xs []ProfileHold) map[string]ProfileHold {
	out := make(map[string]ProfileHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileHold builds a valid ProfileHold, for tests and for the seed data the cron
// binary writes.
func SampleProfileHold(id string, at time.Time) ProfileHold {
	return ProfileHold{
		ID:       id,
		Group:    "group-" + id,
		Channel:  "channel-" + id,
		Owner:    "seeded by profile.SampleProfileHold",
		Retained: len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileClaimSchema is the stored shape of a ProfileClaim.
var ProfileClaimSchema = schema.New("profile.profileclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileClaim is one profile claim.
type ProfileClaim struct {
	ID       string
	Owner    string
	Group    string
	Comment  string
	Version  int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileClaim) Key() string { return "profile/profileclaim/" + x.ID }

// Fields encodes x for storage.
func (x ProfileClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"group":    x.Group,
		"comment":  x.Comment,
		"version":  strconv.Itoa(x.Version),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileClaim decodes a stored record.
func ParseProfileClaim(f codec.Fields) (ProfileClaim, error) {
	var x ProfileClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profileclaim: missing id")
	}
	x.Owner = f["owner"]
	x.Group = f["group"]
	x.Comment = f["comment"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("profile: profileclaim %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("profile: profileclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profileclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profileclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileClaim) String() string {
	return fmt.Sprintf("ProfileClaim(%s, owner=%q, version=%d, archived=%t)", x.ID, x.Owner, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileClaim) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileClaim) Normalize() ProfileClaim {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Group = strings.TrimSpace(x.Group)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortProfileClaims orders by version descending, then by ID.
func SortProfileClaims(xs []ProfileClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileClaims drops the archived entries, preserving order.
func ActiveProfileClaims(xs []ProfileClaim) []ProfileClaim {
	out := make([]ProfileClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileClaims adds up the version of every active entry.
func SumProfileClaims(xs []ProfileClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexProfileClaims keys the entries by ID, last write winning.
func IndexProfileClaims(xs []ProfileClaim) map[string]ProfileClaim {
	out := make(map[string]ProfileClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileClaim builds a valid ProfileClaim, for tests and for the seed data the cron
// binary writes.
func SampleProfileClaim(id string, at time.Time) ProfileClaim {
	return ProfileClaim{
		ID:       id,
		Owner:    "owner-" + id,
		Group:    "group-" + id,
		Comment:  "seeded by profile.SampleProfileClaim",
		Version:  len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// ProfileCursorSchema is the stored shape of a ProfileCursor.
var ProfileCursorSchema = schema.New("profile.profilecursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// ProfileCursor is one profile cursor.
type ProfileCursor struct {
	ID       string
	Region   string
	Channel  string
	Comment  string
	Capacity int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x ProfileCursor) Key() string { return "profile/profilecursor/" + x.ID }

// Fields encodes x for storage.
func (x ProfileCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseProfileCursor decodes a stored record.
func ParseProfileCursor(f codec.Fields) (ProfileCursor, error) {
	var x ProfileCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("profile: profilecursor: missing id")
	}
	x.Region = f["region"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("profile: profilecursor %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("profile: profilecursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("profile: profilecursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("profile: profilecursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x ProfileCursor) String() string {
	return fmt.Sprintf("ProfileCursor(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x ProfileCursor) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x ProfileCursor) Normalize() ProfileCursor {
	x.Region = strings.TrimSpace(x.Region)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortProfileCursors orders by capacity descending, then by ID.
func SortProfileCursors(xs []ProfileCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveProfileCursors drops the archived entries, preserving order.
func ActiveProfileCursors(xs []ProfileCursor) []ProfileCursor {
	out := make([]ProfileCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumProfileCursors adds up the capacity of every active entry.
func SumProfileCursors(xs []ProfileCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexProfileCursors keys the entries by ID, last write winning.
func IndexProfileCursors(xs []ProfileCursor) map[string]ProfileCursor {
	out := make(map[string]ProfileCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleProfileCursor builds a valid ProfileCursor, for tests and for the seed data the cron
// binary writes.
func SampleProfileCursor(id string, at time.Time) ProfileCursor {
	return ProfileCursor{
		ID:       id,
		Region:   "region-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by profile.SampleProfileCursor",
		Capacity: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	ProfileRecordSchema,
	ProfileEntrySchema,
	ProfileItemSchema,
	ProfileDraftSchema,
	ProfileRevisionSchema,
	ProfileSnapshotSchema,
	ProfileBatchSchema,
	ProfileRequestSchema,
	ProfileSummarySchema,
	ProfilePolicySchema,
	ProfileLinkSchema,
	ProfileNoteSchema,
	ProfileWindowSchema,
	ProfileQuotaSchema,
	ProfileBucketSchema,
	ProfileLaneSchema,
	ProfileSlotSchema,
	ProfileTagSchema,
	ProfileRuleSchema,
	ProfileHoldSchema,
	ProfileClaimSchema,
	ProfileCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"profile/profilerecord/",
	"profile/profileentry/",
	"profile/profileitem/",
	"profile/profiledraft/",
	"profile/profilerevision/",
	"profile/profilesnapshot/",
	"profile/profilebatch/",
	"profile/profilerequest/",
	"profile/profilesummary/",
	"profile/profilepolicy/",
	"profile/profilelink/",
	"profile/profilenote/",
	"profile/profilewindow/",
	"profile/profilequota/",
	"profile/profilebucket/",
	"profile/profilelane/",
	"profile/profileslot/",
	"profile/profiletag/",
	"profile/profilerule/",
	"profile/profilehold/",
	"profile/profileclaim/",
	"profile/profilecursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "profile: " + strings.Join(names, ", ")
}
