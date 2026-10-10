// Package notification holds the notification domain model.
//
// Every type in here follows the same template: a struct, the schema it is
// stored under, and the pure helpers the service layer needs. There is no
// behaviour worth reading closely; the package exists to give the build graph
// realistic width.
package notification

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gregmagolan/extending-bazel-with-starlark/lib/codec"
	"github.com/gregmagolan/extending-bazel-with-starlark/lib/schema"
)

// NotificationRecordSchema is the stored shape of a NotificationRecord.
var NotificationRecordSchema = schema.New("notification.notificationrecord",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationRecord is one notification record.
type NotificationRecord struct {
	ID       string
	Region   string
	Source   string
	Category string
	Version  int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationRecord) Key() string { return "notification/notificationrecord/" + x.ID }

// Fields encodes x for storage.
func (x NotificationRecord) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"source":   x.Source,
		"category": x.Category,
		"version":  strconv.Itoa(x.Version),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationRecord decodes a stored record.
func ParseNotificationRecord(f codec.Fields) (NotificationRecord, error) {
	var x NotificationRecord
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationrecord: missing id")
	}
	x.Region = f["region"]
	x.Source = f["source"]
	x.Category = f["category"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrecord %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrecord %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationrecord %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationrecord %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationRecord) String() string {
	return fmt.Sprintf("NotificationRecord(%s, region=%q, version=%d, archived=%t)", x.ID, x.Region, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationRecord) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationRecord) Normalize() NotificationRecord {
	x.Region = strings.TrimSpace(x.Region)
	x.Source = strings.TrimSpace(x.Source)
	x.Category = strings.TrimSpace(x.Category)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortNotificationRecords orders by version descending, then by ID.
func SortNotificationRecords(xs []NotificationRecord) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationRecords drops the archived entries, preserving order.
func ActiveNotificationRecords(xs []NotificationRecord) []NotificationRecord {
	out := make([]NotificationRecord, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationRecords adds up the version of every active entry.
func SumNotificationRecords(xs []NotificationRecord) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexNotificationRecords keys the entries by ID, last write winning.
func IndexNotificationRecords(xs []NotificationRecord) map[string]NotificationRecord {
	out := make(map[string]NotificationRecord, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationRecord builds a valid NotificationRecord, for tests and for the seed data the cron
// binary writes.
func SampleNotificationRecord(id string, at time.Time) NotificationRecord {
	return NotificationRecord{
		ID:       id,
		Region:   "region-" + id,
		Source:   "source-" + id,
		Category: "seeded by notification.SampleNotificationRecord",
		Version:  len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationEntrySchema is the stored shape of a NotificationEntry.
var NotificationEntrySchema = schema.New("notification.notificationentry",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationEntry is one notification entry.
type NotificationEntry struct {
	ID       string
	Name     string
	Channel  string
	Region   string
	Retained int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationEntry) Key() string { return "notification/notificationentry/" + x.ID }

// Fields encodes x for storage.
func (x NotificationEntry) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationEntry decodes a stored record.
func ParseNotificationEntry(f codec.Fields) (NotificationEntry, error) {
	var x NotificationEntry
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationentry: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("notification: notificationentry %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("notification: notificationentry %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationentry %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationentry %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationEntry) String() string {
	return fmt.Sprintf("NotificationEntry(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationEntry) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationEntry) Normalize() NotificationEntry {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortNotificationEntrys orders by retained descending, then by ID.
func SortNotificationEntrys(xs []NotificationEntry) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationEntrys drops the archived entries, preserving order.
func ActiveNotificationEntrys(xs []NotificationEntry) []NotificationEntry {
	out := make([]NotificationEntry, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationEntrys adds up the retained of every active entry.
func SumNotificationEntrys(xs []NotificationEntry) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexNotificationEntrys keys the entries by ID, last write winning.
func IndexNotificationEntrys(xs []NotificationEntry) map[string]NotificationEntry {
	out := make(map[string]NotificationEntry, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationEntry builds a valid NotificationEntry, for tests and for the seed data the cron
// binary writes.
func SampleNotificationEntry(id string, at time.Time) NotificationEntry {
	return NotificationEntry{
		ID:       id,
		Name:     "name-" + id,
		Channel:  "channel-" + id,
		Region:   "seeded by notification.SampleNotificationEntry",
		Retained: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationItemSchema is the stored shape of a NotificationItem.
var NotificationItemSchema = schema.New("notification.notificationitem",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "owner", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationItem is one notification item.
type NotificationItem struct {
	ID       string
	Owner    string
	Channel  string
	Name     string
	Attempts int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationItem) Key() string { return "notification/notificationitem/" + x.ID }

// Fields encodes x for storage.
func (x NotificationItem) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationItem decodes a stored record.
func ParseNotificationItem(f codec.Fields) (NotificationItem, error) {
	var x NotificationItem
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationitem: missing id")
	}
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("notification: notificationitem %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("notification: notificationitem %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationitem %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationitem %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationItem) String() string {
	return fmt.Sprintf("NotificationItem(%s, owner=%q, attempts=%d, archived=%t)", x.ID, x.Owner, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationItem) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationItem) Normalize() NotificationItem {
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortNotificationItems orders by attempts descending, then by ID.
func SortNotificationItems(xs []NotificationItem) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationItems drops the archived entries, preserving order.
func ActiveNotificationItems(xs []NotificationItem) []NotificationItem {
	out := make([]NotificationItem, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationItems adds up the attempts of every active entry.
func SumNotificationItems(xs []NotificationItem) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexNotificationItems keys the entries by ID, last write winning.
func IndexNotificationItems(xs []NotificationItem) map[string]NotificationItem {
	out := make(map[string]NotificationItem, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationItem builds a valid NotificationItem, for tests and for the seed data the cron
// binary writes.
func SampleNotificationItem(id string, at time.Time) NotificationItem {
	return NotificationItem{
		ID:       id,
		Owner:    "owner-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by notification.SampleNotificationItem",
		Attempts: len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationDraftSchema is the stored shape of a NotificationDraft.
var NotificationDraftSchema = schema.New("notification.notificationdraft",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationDraft is one notification draft.
type NotificationDraft struct {
	ID       string
	Name     string
	Label    string
	Group    string
	Retained int
	Offset   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationDraft) Key() string { return "notification/notificationdraft/" + x.ID }

// Fields encodes x for storage.
func (x NotificationDraft) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"label":    x.Label,
		"group":    x.Group,
		"retained": strconv.Itoa(x.Retained),
		"offset":   strconv.Itoa(x.Offset),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationDraft decodes a stored record.
func ParseNotificationDraft(f codec.Fields) (NotificationDraft, error) {
	var x NotificationDraft
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationdraft: missing id")
	}
	x.Name = f["name"]
	x.Label = f["label"]
	x.Group = f["group"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("notification: notificationdraft %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationdraft %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationdraft %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationdraft %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationDraft) String() string {
	return fmt.Sprintf("NotificationDraft(%s, name=%q, retained=%d, archived=%t)", x.ID, x.Name, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationDraft) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationDraft) Normalize() NotificationDraft {
	x.Name = strings.TrimSpace(x.Name)
	x.Label = strings.TrimSpace(x.Label)
	x.Group = strings.TrimSpace(x.Group)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortNotificationDrafts orders by retained descending, then by ID.
func SortNotificationDrafts(xs []NotificationDraft) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationDrafts drops the archived entries, preserving order.
func ActiveNotificationDrafts(xs []NotificationDraft) []NotificationDraft {
	out := make([]NotificationDraft, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationDrafts adds up the retained of every active entry.
func SumNotificationDrafts(xs []NotificationDraft) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexNotificationDrafts keys the entries by ID, last write winning.
func IndexNotificationDrafts(xs []NotificationDraft) map[string]NotificationDraft {
	out := make(map[string]NotificationDraft, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationDraft builds a valid NotificationDraft, for tests and for the seed data the cron
// binary writes.
func SampleNotificationDraft(id string, at time.Time) NotificationDraft {
	return NotificationDraft{
		ID:       id,
		Name:     "name-" + id,
		Label:    "label-" + id,
		Group:    "seeded by notification.SampleNotificationDraft",
		Retained: len(id),
		Offset:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationRevisionSchema is the stored shape of a NotificationRevision.
var NotificationRevisionSchema = schema.New("notification.notificationrevision",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationRevision is one notification revision.
type NotificationRevision struct {
	ID       string
	Region   string
	Name     string
	Category string
	Count    int
	Retained int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationRevision) Key() string { return "notification/notificationrevision/" + x.ID }

// Fields encodes x for storage.
func (x NotificationRevision) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"name":     x.Name,
		"category": x.Category,
		"count":    strconv.Itoa(x.Count),
		"retained": strconv.Itoa(x.Retained),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationRevision decodes a stored record.
func ParseNotificationRevision(f codec.Fields) (NotificationRevision, error) {
	var x NotificationRevision
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationrevision: missing id")
	}
	x.Region = f["region"]
	x.Name = f["name"]
	x.Category = f["category"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrevision %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrevision %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationrevision %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationrevision %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationRevision) String() string {
	return fmt.Sprintf("NotificationRevision(%s, region=%q, count=%d, archived=%t)", x.ID, x.Region, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationRevision) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationRevision) Normalize() NotificationRevision {
	x.Region = strings.TrimSpace(x.Region)
	x.Name = strings.TrimSpace(x.Name)
	x.Category = strings.TrimSpace(x.Category)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortNotificationRevisions orders by count descending, then by ID.
func SortNotificationRevisions(xs []NotificationRevision) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationRevisions drops the archived entries, preserving order.
func ActiveNotificationRevisions(xs []NotificationRevision) []NotificationRevision {
	out := make([]NotificationRevision, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationRevisions adds up the count of every active entry.
func SumNotificationRevisions(xs []NotificationRevision) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexNotificationRevisions keys the entries by ID, last write winning.
func IndexNotificationRevisions(xs []NotificationRevision) map[string]NotificationRevision {
	out := make(map[string]NotificationRevision, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationRevision builds a valid NotificationRevision, for tests and for the seed data the cron
// binary writes.
func SampleNotificationRevision(id string, at time.Time) NotificationRevision {
	return NotificationRevision{
		ID:       id,
		Region:   "region-" + id,
		Name:     "name-" + id,
		Category: "seeded by notification.SampleNotificationRevision",
		Count:    len(id),
		Retained: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationSnapshotSchema is the stored shape of a NotificationSnapshot.
var NotificationSnapshotSchema = schema.New("notification.notificationsnapshot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "category", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationSnapshot is one notification snapshot.
type NotificationSnapshot struct {
	ID       string
	Category string
	Channel  string
	Name     string
	Offset   int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationSnapshot) Key() string { return "notification/notificationsnapshot/" + x.ID }

// Fields encodes x for storage.
func (x NotificationSnapshot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"category": x.Category,
		"channel":  x.Channel,
		"name":     x.Name,
		"offset":   strconv.Itoa(x.Offset),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationSnapshot decodes a stored record.
func ParseNotificationSnapshot(f codec.Fields) (NotificationSnapshot, error) {
	var x NotificationSnapshot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationsnapshot: missing id")
	}
	x.Category = f["category"]
	x.Channel = f["channel"]
	x.Name = f["name"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationsnapshot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("notification: notificationsnapshot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationsnapshot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationsnapshot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationSnapshot) String() string {
	return fmt.Sprintf("NotificationSnapshot(%s, category=%q, offset=%d, archived=%t)", x.ID, x.Category, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationSnapshot) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationSnapshot) Normalize() NotificationSnapshot {
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Name = strings.TrimSpace(x.Name)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortNotificationSnapshots orders by offset descending, then by ID.
func SortNotificationSnapshots(xs []NotificationSnapshot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationSnapshots drops the archived entries, preserving order.
func ActiveNotificationSnapshots(xs []NotificationSnapshot) []NotificationSnapshot {
	out := make([]NotificationSnapshot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationSnapshots adds up the offset of every active entry.
func SumNotificationSnapshots(xs []NotificationSnapshot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexNotificationSnapshots keys the entries by ID, last write winning.
func IndexNotificationSnapshots(xs []NotificationSnapshot) map[string]NotificationSnapshot {
	out := make(map[string]NotificationSnapshot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationSnapshot builds a valid NotificationSnapshot, for tests and for the seed data the cron
// binary writes.
func SampleNotificationSnapshot(id string, at time.Time) NotificationSnapshot {
	return NotificationSnapshot{
		ID:       id,
		Category: "category-" + id,
		Channel:  "channel-" + id,
		Name:     "seeded by notification.SampleNotificationSnapshot",
		Offset:   len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationBatchSchema is the stored shape of a NotificationBatch.
var NotificationBatchSchema = schema.New("notification.notificationbatch",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationBatch is one notification batch.
type NotificationBatch struct {
	ID        string
	Label     string
	Channel   string
	Reference string
	Version   int
	Attempts  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationBatch) Key() string { return "notification/notificationbatch/" + x.ID }

// Fields encodes x for storage.
func (x NotificationBatch) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"label":     x.Label,
		"channel":   x.Channel,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"attempts":  strconv.Itoa(x.Attempts),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationBatch decodes a stored record.
func ParseNotificationBatch(f codec.Fields) (NotificationBatch, error) {
	var x NotificationBatch
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationbatch: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationbatch %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("notification: notificationbatch %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationbatch %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationbatch %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationBatch) String() string {
	return fmt.Sprintf("NotificationBatch(%s, label=%q, version=%d, archived=%t)", x.ID, x.Label, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationBatch) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationBatch) Normalize() NotificationBatch {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortNotificationBatchs orders by version descending, then by ID.
func SortNotificationBatchs(xs []NotificationBatch) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationBatchs drops the archived entries, preserving order.
func ActiveNotificationBatchs(xs []NotificationBatch) []NotificationBatch {
	out := make([]NotificationBatch, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationBatchs adds up the version of every active entry.
func SumNotificationBatchs(xs []NotificationBatch) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexNotificationBatchs keys the entries by ID, last write winning.
func IndexNotificationBatchs(xs []NotificationBatch) map[string]NotificationBatch {
	out := make(map[string]NotificationBatch, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationBatch builds a valid NotificationBatch, for tests and for the seed data the cron
// binary writes.
func SampleNotificationBatch(id string, at time.Time) NotificationBatch {
	return NotificationBatch{
		ID:        id,
		Label:     "label-" + id,
		Channel:   "channel-" + id,
		Reference: "seeded by notification.SampleNotificationBatch",
		Version:   len(id),
		Attempts:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// NotificationRequestSchema is the stored shape of a NotificationRequest.
var NotificationRequestSchema = schema.New("notification.notificationrequest",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationRequest is one notification request.
type NotificationRequest struct {
	ID        string
	Group     string
	Category  string
	Reference string
	Count     int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationRequest) Key() string { return "notification/notificationrequest/" + x.ID }

// Fields encodes x for storage.
func (x NotificationRequest) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"category":  x.Category,
		"reference": x.Reference,
		"count":     strconv.Itoa(x.Count),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationRequest decodes a stored record.
func ParseNotificationRequest(f codec.Fields) (NotificationRequest, error) {
	var x NotificationRequest
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationrequest: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Reference = f["reference"]
	var err error
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrequest %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrequest %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationrequest %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationrequest %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationRequest) String() string {
	return fmt.Sprintf("NotificationRequest(%s, group=%q, count=%d, archived=%t)", x.ID, x.Group, x.Count, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationRequest) Active() bool { return !x.Archived && x.Count >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationRequest) Normalize() NotificationRequest {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Count < 0 {
		x.Count = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortNotificationRequests orders by count descending, then by ID.
func SortNotificationRequests(xs []NotificationRequest) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Count != xs[j].Count {
			return xs[i].Count > xs[j].Count
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationRequests drops the archived entries, preserving order.
func ActiveNotificationRequests(xs []NotificationRequest) []NotificationRequest {
	out := make([]NotificationRequest, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationRequests adds up the count of every active entry.
func SumNotificationRequests(xs []NotificationRequest) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Count
		}
	}
	return total
}

// IndexNotificationRequests keys the entries by ID, last write winning.
func IndexNotificationRequests(xs []NotificationRequest) map[string]NotificationRequest {
	out := make(map[string]NotificationRequest, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationRequest builds a valid NotificationRequest, for tests and for the seed data the cron
// binary writes.
func SampleNotificationRequest(id string, at time.Time) NotificationRequest {
	return NotificationRequest{
		ID:        id,
		Group:     "group-" + id,
		Category:  "category-" + id,
		Reference: "seeded by notification.SampleNotificationRequest",
		Count:     len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// NotificationSummarySchema is the stored shape of a NotificationSummary.
var NotificationSummarySchema = schema.New("notification.notificationsummary",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "comment", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationSummary is one notification summary.
type NotificationSummary struct {
	ID        string
	Comment   string
	Reference string
	Channel   string
	Capacity  int
	Retained  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationSummary) Key() string { return "notification/notificationsummary/" + x.ID }

// Fields encodes x for storage.
func (x NotificationSummary) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"comment":   x.Comment,
		"reference": x.Reference,
		"channel":   x.Channel,
		"capacity":  strconv.Itoa(x.Capacity),
		"retained":  strconv.Itoa(x.Retained),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationSummary decodes a stored record.
func ParseNotificationSummary(f codec.Fields) (NotificationSummary, error) {
	var x NotificationSummary
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationsummary: missing id")
	}
	x.Comment = f["comment"]
	x.Reference = f["reference"]
	x.Channel = f["channel"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("notification: notificationsummary %s: %w", x.ID, err)
	}
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("notification: notificationsummary %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationsummary %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationsummary %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationSummary) String() string {
	return fmt.Sprintf("NotificationSummary(%s, comment=%q, capacity=%d, archived=%t)", x.ID, x.Comment, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationSummary) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationSummary) Normalize() NotificationSummary {
	x.Comment = strings.TrimSpace(x.Comment)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Retained < 0 {
		x.Retained = 0
	}
	return x
}

// SortNotificationSummarys orders by capacity descending, then by ID.
func SortNotificationSummarys(xs []NotificationSummary) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationSummarys drops the archived entries, preserving order.
func ActiveNotificationSummarys(xs []NotificationSummary) []NotificationSummary {
	out := make([]NotificationSummary, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationSummarys adds up the capacity of every active entry.
func SumNotificationSummarys(xs []NotificationSummary) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexNotificationSummarys keys the entries by ID, last write winning.
func IndexNotificationSummarys(xs []NotificationSummary) map[string]NotificationSummary {
	out := make(map[string]NotificationSummary, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationSummary builds a valid NotificationSummary, for tests and for the seed data the cron
// binary writes.
func SampleNotificationSummary(id string, at time.Time) NotificationSummary {
	return NotificationSummary{
		ID:        id,
		Comment:   "comment-" + id,
		Reference: "reference-" + id,
		Channel:   "seeded by notification.SampleNotificationSummary",
		Capacity:  len(id),
		Retained:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// NotificationPolicySchema is the stored shape of a NotificationPolicy.
var NotificationPolicySchema = schema.New("notification.notificationpolicy",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "source", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationPolicy is one notification policy.
type NotificationPolicy struct {
	ID       string
	Group    string
	Source   string
	Name     string
	Attempts int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationPolicy) Key() string { return "notification/notificationpolicy/" + x.ID }

// Fields encodes x for storage.
func (x NotificationPolicy) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"source":   x.Source,
		"name":     x.Name,
		"attempts": strconv.Itoa(x.Attempts),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationPolicy decodes a stored record.
func ParseNotificationPolicy(f codec.Fields) (NotificationPolicy, error) {
	var x NotificationPolicy
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationpolicy: missing id")
	}
	x.Group = f["group"]
	x.Source = f["source"]
	x.Name = f["name"]
	var err error
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("notification: notificationpolicy %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("notification: notificationpolicy %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationpolicy %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationpolicy %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationPolicy) String() string {
	return fmt.Sprintf("NotificationPolicy(%s, group=%q, attempts=%d, archived=%t)", x.ID, x.Group, x.Attempts, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationPolicy) Active() bool { return !x.Archived && x.Attempts >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationPolicy) Normalize() NotificationPolicy {
	x.Group = strings.TrimSpace(x.Group)
	x.Source = strings.TrimSpace(x.Source)
	x.Name = strings.TrimSpace(x.Name)
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortNotificationPolicys orders by attempts descending, then by ID.
func SortNotificationPolicys(xs []NotificationPolicy) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Attempts != xs[j].Attempts {
			return xs[i].Attempts > xs[j].Attempts
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationPolicys drops the archived entries, preserving order.
func ActiveNotificationPolicys(xs []NotificationPolicy) []NotificationPolicy {
	out := make([]NotificationPolicy, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationPolicys adds up the attempts of every active entry.
func SumNotificationPolicys(xs []NotificationPolicy) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Attempts
		}
	}
	return total
}

// IndexNotificationPolicys keys the entries by ID, last write winning.
func IndexNotificationPolicys(xs []NotificationPolicy) map[string]NotificationPolicy {
	out := make(map[string]NotificationPolicy, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationPolicy builds a valid NotificationPolicy, for tests and for the seed data the cron
// binary writes.
func SampleNotificationPolicy(id string, at time.Time) NotificationPolicy {
	return NotificationPolicy{
		ID:       id,
		Group:    "group-" + id,
		Source:   "source-" + id,
		Name:     "seeded by notification.SampleNotificationPolicy",
		Attempts: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationLinkSchema is the stored shape of a NotificationLink.
var NotificationLinkSchema = schema.New("notification.notificationlink",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "retained", Kind: schema.KindInt},
	schema.Field{Name: "count", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationLink is one notification link.
type NotificationLink struct {
	ID       string
	Channel  string
	Comment  string
	Region   string
	Retained int
	Count    int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationLink) Key() string { return "notification/notificationlink/" + x.ID }

// Fields encodes x for storage.
func (x NotificationLink) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"region":   x.Region,
		"retained": strconv.Itoa(x.Retained),
		"count":    strconv.Itoa(x.Count),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationLink decodes a stored record.
func ParseNotificationLink(f codec.Fields) (NotificationLink, error) {
	var x NotificationLink
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationlink: missing id")
	}
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	x.Region = f["region"]
	var err error
	if x.Retained, err = f.Int("retained", 0); err != nil {
		return x, fmt.Errorf("notification: notificationlink %s: %w", x.ID, err)
	}
	if x.Count, err = f.Int("count", 0); err != nil {
		return x, fmt.Errorf("notification: notificationlink %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationlink %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationlink %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationLink) String() string {
	return fmt.Sprintf("NotificationLink(%s, channel=%q, retained=%d, archived=%t)", x.ID, x.Channel, x.Retained, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationLink) Active() bool { return !x.Archived && x.Retained >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationLink) Normalize() NotificationLink {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	x.Region = strings.TrimSpace(x.Region)
	if x.Retained < 0 {
		x.Retained = 0
	}
	if x.Count < 0 {
		x.Count = 0
	}
	return x
}

// SortNotificationLinks orders by retained descending, then by ID.
func SortNotificationLinks(xs []NotificationLink) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Retained != xs[j].Retained {
			return xs[i].Retained > xs[j].Retained
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationLinks drops the archived entries, preserving order.
func ActiveNotificationLinks(xs []NotificationLink) []NotificationLink {
	out := make([]NotificationLink, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationLinks adds up the retained of every active entry.
func SumNotificationLinks(xs []NotificationLink) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Retained
		}
	}
	return total
}

// IndexNotificationLinks keys the entries by ID, last write winning.
func IndexNotificationLinks(xs []NotificationLink) map[string]NotificationLink {
	out := make(map[string]NotificationLink, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationLink builds a valid NotificationLink, for tests and for the seed data the cron
// binary writes.
func SampleNotificationLink(id string, at time.Time) NotificationLink {
	return NotificationLink{
		ID:       id,
		Channel:  "channel-" + id,
		Comment:  "comment-" + id,
		Region:   "seeded by notification.SampleNotificationLink",
		Retained: len(id),
		Count:    len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationNoteSchema is the stored shape of a NotificationNote.
var NotificationNoteSchema = schema.New("notification.notificationnote",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationNote is one notification note.
type NotificationNote struct {
	ID       string
	Group    string
	Category string
	Channel  string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationNote) Key() string { return "notification/notificationnote/" + x.ID }

// Fields encodes x for storage.
func (x NotificationNote) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"group":    x.Group,
		"category": x.Category,
		"channel":  x.Channel,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationNote decodes a stored record.
func ParseNotificationNote(f codec.Fields) (NotificationNote, error) {
	var x NotificationNote
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationnote: missing id")
	}
	x.Group = f["group"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationnote %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("notification: notificationnote %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationnote %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationnote %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationNote) String() string {
	return fmt.Sprintf("NotificationNote(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationNote) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationNote) Normalize() NotificationNote {
	x.Group = strings.TrimSpace(x.Group)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortNotificationNotes orders by offset descending, then by ID.
func SortNotificationNotes(xs []NotificationNote) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationNotes drops the archived entries, preserving order.
func ActiveNotificationNotes(xs []NotificationNote) []NotificationNote {
	out := make([]NotificationNote, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationNotes adds up the offset of every active entry.
func SumNotificationNotes(xs []NotificationNote) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexNotificationNotes keys the entries by ID, last write winning.
func IndexNotificationNotes(xs []NotificationNote) map[string]NotificationNote {
	out := make(map[string]NotificationNote, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationNote builds a valid NotificationNote, for tests and for the seed data the cron
// binary writes.
func SampleNotificationNote(id string, at time.Time) NotificationNote {
	return NotificationNote{
		ID:       id,
		Group:    "group-" + id,
		Category: "category-" + id,
		Channel:  "seeded by notification.SampleNotificationNote",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationWindowSchema is the stored shape of a NotificationWindow.
var NotificationWindowSchema = schema.New("notification.notificationwindow",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationWindow is one notification window.
type NotificationWindow struct {
	ID        string
	Region    string
	Reference string
	Category  string
	Capacity  int
	Version   int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationWindow) Key() string { return "notification/notificationwindow/" + x.ID }

// Fields encodes x for storage.
func (x NotificationWindow) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"region":    x.Region,
		"reference": x.Reference,
		"category":  x.Category,
		"capacity":  strconv.Itoa(x.Capacity),
		"version":   strconv.Itoa(x.Version),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationWindow decodes a stored record.
func ParseNotificationWindow(f codec.Fields) (NotificationWindow, error) {
	var x NotificationWindow
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationwindow: missing id")
	}
	x.Region = f["region"]
	x.Reference = f["reference"]
	x.Category = f["category"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("notification: notificationwindow %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationwindow %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationwindow %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationwindow %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationWindow) String() string {
	return fmt.Sprintf("NotificationWindow(%s, region=%q, capacity=%d, archived=%t)", x.ID, x.Region, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationWindow) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationWindow) Normalize() NotificationWindow {
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	x.Category = strings.TrimSpace(x.Category)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortNotificationWindows orders by capacity descending, then by ID.
func SortNotificationWindows(xs []NotificationWindow) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationWindows drops the archived entries, preserving order.
func ActiveNotificationWindows(xs []NotificationWindow) []NotificationWindow {
	out := make([]NotificationWindow, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationWindows adds up the capacity of every active entry.
func SumNotificationWindows(xs []NotificationWindow) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexNotificationWindows keys the entries by ID, last write winning.
func IndexNotificationWindows(xs []NotificationWindow) map[string]NotificationWindow {
	out := make(map[string]NotificationWindow, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationWindow builds a valid NotificationWindow, for tests and for the seed data the cron
// binary writes.
func SampleNotificationWindow(id string, at time.Time) NotificationWindow {
	return NotificationWindow{
		ID:        id,
		Region:    "region-" + id,
		Reference: "reference-" + id,
		Category:  "seeded by notification.SampleNotificationWindow",
		Capacity:  len(id),
		Version:   len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// NotificationQuotaSchema is the stored shape of a NotificationQuota.
var NotificationQuotaSchema = schema.New("notification.notificationquota",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationQuota is one notification quota.
type NotificationQuota struct {
	ID        string
	Source    string
	Region    string
	Reference string
	Version   int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationQuota) Key() string { return "notification/notificationquota/" + x.ID }

// Fields encodes x for storage.
func (x NotificationQuota) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"region":    x.Region,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationQuota decodes a stored record.
func ParseNotificationQuota(f codec.Fields) (NotificationQuota, error) {
	var x NotificationQuota
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationquota: missing id")
	}
	x.Source = f["source"]
	x.Region = f["region"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationquota %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationquota %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationquota %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationquota %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationQuota) String() string {
	return fmt.Sprintf("NotificationQuota(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationQuota) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationQuota) Normalize() NotificationQuota {
	x.Source = strings.TrimSpace(x.Source)
	x.Region = strings.TrimSpace(x.Region)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortNotificationQuotas orders by version descending, then by ID.
func SortNotificationQuotas(xs []NotificationQuota) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationQuotas drops the archived entries, preserving order.
func ActiveNotificationQuotas(xs []NotificationQuota) []NotificationQuota {
	out := make([]NotificationQuota, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationQuotas adds up the version of every active entry.
func SumNotificationQuotas(xs []NotificationQuota) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexNotificationQuotas keys the entries by ID, last write winning.
func IndexNotificationQuotas(xs []NotificationQuota) map[string]NotificationQuota {
	out := make(map[string]NotificationQuota, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationQuota builds a valid NotificationQuota, for tests and for the seed data the cron
// binary writes.
func SampleNotificationQuota(id string, at time.Time) NotificationQuota {
	return NotificationQuota{
		ID:        id,
		Source:    "source-" + id,
		Region:    "region-" + id,
		Reference: "seeded by notification.SampleNotificationQuota",
		Version:   len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// NotificationBucketSchema is the stored shape of a NotificationBucket.
var NotificationBucketSchema = schema.New("notification.notificationbucket",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "reference", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "group", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationBucket is one notification bucket.
type NotificationBucket struct {
	ID        string
	Reference string
	Region    string
	Group     string
	Version   int
	Weight    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationBucket) Key() string { return "notification/notificationbucket/" + x.ID }

// Fields encodes x for storage.
func (x NotificationBucket) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"reference": x.Reference,
		"region":    x.Region,
		"group":     x.Group,
		"version":   strconv.Itoa(x.Version),
		"weight":    strconv.Itoa(x.Weight),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationBucket decodes a stored record.
func ParseNotificationBucket(f codec.Fields) (NotificationBucket, error) {
	var x NotificationBucket
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationbucket: missing id")
	}
	x.Reference = f["reference"]
	x.Region = f["region"]
	x.Group = f["group"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationbucket %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("notification: notificationbucket %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationbucket %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationbucket %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationBucket) String() string {
	return fmt.Sprintf("NotificationBucket(%s, reference=%q, version=%d, archived=%t)", x.ID, x.Reference, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationBucket) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationBucket) Normalize() NotificationBucket {
	x.Reference = strings.TrimSpace(x.Reference)
	x.Region = strings.TrimSpace(x.Region)
	x.Group = strings.TrimSpace(x.Group)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortNotificationBuckets orders by version descending, then by ID.
func SortNotificationBuckets(xs []NotificationBucket) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationBuckets drops the archived entries, preserving order.
func ActiveNotificationBuckets(xs []NotificationBucket) []NotificationBucket {
	out := make([]NotificationBucket, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationBuckets adds up the version of every active entry.
func SumNotificationBuckets(xs []NotificationBucket) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexNotificationBuckets keys the entries by ID, last write winning.
func IndexNotificationBuckets(xs []NotificationBucket) map[string]NotificationBucket {
	out := make(map[string]NotificationBucket, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationBucket builds a valid NotificationBucket, for tests and for the seed data the cron
// binary writes.
func SampleNotificationBucket(id string, at time.Time) NotificationBucket {
	return NotificationBucket{
		ID:        id,
		Reference: "reference-" + id,
		Region:    "region-" + id,
		Group:     "seeded by notification.SampleNotificationBucket",
		Version:   len(id),
		Weight:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// NotificationLaneSchema is the stored shape of a NotificationLane.
var NotificationLaneSchema = schema.New("notification.notificationlane",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "owner", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationLane is one notification lane.
type NotificationLane struct {
	ID       string
	Source   string
	Owner    string
	Channel  string
	Version  int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationLane) Key() string { return "notification/notificationlane/" + x.ID }

// Fields encodes x for storage.
func (x NotificationLane) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"source":   x.Source,
		"owner":    x.Owner,
		"channel":  x.Channel,
		"version":  strconv.Itoa(x.Version),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationLane decodes a stored record.
func ParseNotificationLane(f codec.Fields) (NotificationLane, error) {
	var x NotificationLane
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationlane: missing id")
	}
	x.Source = f["source"]
	x.Owner = f["owner"]
	x.Channel = f["channel"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationlane %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("notification: notificationlane %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationlane %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationlane %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationLane) String() string {
	return fmt.Sprintf("NotificationLane(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationLane) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationLane) Normalize() NotificationLane {
	x.Source = strings.TrimSpace(x.Source)
	x.Owner = strings.TrimSpace(x.Owner)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortNotificationLanes orders by version descending, then by ID.
func SortNotificationLanes(xs []NotificationLane) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationLanes drops the archived entries, preserving order.
func ActiveNotificationLanes(xs []NotificationLane) []NotificationLane {
	out := make([]NotificationLane, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationLanes adds up the version of every active entry.
func SumNotificationLanes(xs []NotificationLane) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexNotificationLanes keys the entries by ID, last write winning.
func IndexNotificationLanes(xs []NotificationLane) map[string]NotificationLane {
	out := make(map[string]NotificationLane, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationLane builds a valid NotificationLane, for tests and for the seed data the cron
// binary writes.
func SampleNotificationLane(id string, at time.Time) NotificationLane {
	return NotificationLane{
		ID:       id,
		Source:   "source-" + id,
		Owner:    "owner-" + id,
		Channel:  "seeded by notification.SampleNotificationLane",
		Version:  len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationSlotSchema is the stored shape of a NotificationSlot.
var NotificationSlotSchema = schema.New("notification.notificationslot",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "channel", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "region", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "attempts", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationSlot is one notification slot.
type NotificationSlot struct {
	ID       string
	Channel  string
	Region   string
	Comment  string
	Priority int
	Attempts int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationSlot) Key() string { return "notification/notificationslot/" + x.ID }

// Fields encodes x for storage.
func (x NotificationSlot) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"channel":  x.Channel,
		"region":   x.Region,
		"comment":  x.Comment,
		"priority": strconv.Itoa(x.Priority),
		"attempts": strconv.Itoa(x.Attempts),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationSlot decodes a stored record.
func ParseNotificationSlot(f codec.Fields) (NotificationSlot, error) {
	var x NotificationSlot
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationslot: missing id")
	}
	x.Channel = f["channel"]
	x.Region = f["region"]
	x.Comment = f["comment"]
	var err error
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("notification: notificationslot %s: %w", x.ID, err)
	}
	if x.Attempts, err = f.Int("attempts", 0); err != nil {
		return x, fmt.Errorf("notification: notificationslot %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationslot %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationslot %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationSlot) String() string {
	return fmt.Sprintf("NotificationSlot(%s, channel=%q, priority=%d, archived=%t)", x.ID, x.Channel, x.Priority, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationSlot) Active() bool { return !x.Archived && x.Priority >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationSlot) Normalize() NotificationSlot {
	x.Channel = strings.TrimSpace(x.Channel)
	x.Region = strings.TrimSpace(x.Region)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Priority < 0 {
		x.Priority = 0
	}
	if x.Attempts < 0 {
		x.Attempts = 0
	}
	return x
}

// SortNotificationSlots orders by priority descending, then by ID.
func SortNotificationSlots(xs []NotificationSlot) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Priority != xs[j].Priority {
			return xs[i].Priority > xs[j].Priority
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationSlots drops the archived entries, preserving order.
func ActiveNotificationSlots(xs []NotificationSlot) []NotificationSlot {
	out := make([]NotificationSlot, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationSlots adds up the priority of every active entry.
func SumNotificationSlots(xs []NotificationSlot) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Priority
		}
	}
	return total
}

// IndexNotificationSlots keys the entries by ID, last write winning.
func IndexNotificationSlots(xs []NotificationSlot) map[string]NotificationSlot {
	out := make(map[string]NotificationSlot, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationSlot builds a valid NotificationSlot, for tests and for the seed data the cron
// binary writes.
func SampleNotificationSlot(id string, at time.Time) NotificationSlot {
	return NotificationSlot{
		ID:       id,
		Channel:  "channel-" + id,
		Region:   "region-" + id,
		Comment:  "seeded by notification.SampleNotificationSlot",
		Priority: len(id),
		Attempts: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationTagSchema is the stored shape of a NotificationTag.
var NotificationTagSchema = schema.New("notification.notificationtag",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "label", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "weight", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationTag is one notification tag.
type NotificationTag struct {
	ID       string
	Label    string
	Channel  string
	Category string
	Offset   int
	Weight   int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationTag) Key() string { return "notification/notificationtag/" + x.ID }

// Fields encodes x for storage.
func (x NotificationTag) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"label":    x.Label,
		"channel":  x.Channel,
		"category": x.Category,
		"offset":   strconv.Itoa(x.Offset),
		"weight":   strconv.Itoa(x.Weight),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationTag decodes a stored record.
func ParseNotificationTag(f codec.Fields) (NotificationTag, error) {
	var x NotificationTag
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationtag: missing id")
	}
	x.Label = f["label"]
	x.Channel = f["channel"]
	x.Category = f["category"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationtag %s: %w", x.ID, err)
	}
	if x.Weight, err = f.Int("weight", 0); err != nil {
		return x, fmt.Errorf("notification: notificationtag %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationtag %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationtag %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationTag) String() string {
	return fmt.Sprintf("NotificationTag(%s, label=%q, offset=%d, archived=%t)", x.ID, x.Label, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationTag) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationTag) Normalize() NotificationTag {
	x.Label = strings.TrimSpace(x.Label)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Category = strings.TrimSpace(x.Category)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Weight < 0 {
		x.Weight = 0
	}
	return x
}

// SortNotificationTags orders by offset descending, then by ID.
func SortNotificationTags(xs []NotificationTag) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationTags drops the archived entries, preserving order.
func ActiveNotificationTags(xs []NotificationTag) []NotificationTag {
	out := make([]NotificationTag, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationTags adds up the offset of every active entry.
func SumNotificationTags(xs []NotificationTag) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexNotificationTags keys the entries by ID, last write winning.
func IndexNotificationTags(xs []NotificationTag) map[string]NotificationTag {
	out := make(map[string]NotificationTag, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationTag builds a valid NotificationTag, for tests and for the seed data the cron
// binary writes.
func SampleNotificationTag(id string, at time.Time) NotificationTag {
	return NotificationTag{
		ID:       id,
		Label:    "label-" + id,
		Channel:  "channel-" + id,
		Category: "seeded by notification.SampleNotificationTag",
		Offset:   len(id),
		Weight:   len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationRuleSchema is the stored shape of a NotificationRule.
var NotificationRuleSchema = schema.New("notification.notificationrule",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "group", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "name", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationRule is one notification rule.
type NotificationRule struct {
	ID        string
	Group     string
	Name      string
	Reference string
	Offset    int
	Priority  int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationRule) Key() string { return "notification/notificationrule/" + x.ID }

// Fields encodes x for storage.
func (x NotificationRule) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"group":     x.Group,
		"name":      x.Name,
		"reference": x.Reference,
		"offset":    strconv.Itoa(x.Offset),
		"priority":  strconv.Itoa(x.Priority),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationRule decodes a stored record.
func ParseNotificationRule(f codec.Fields) (NotificationRule, error) {
	var x NotificationRule
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationrule: missing id")
	}
	x.Group = f["group"]
	x.Name = f["name"]
	x.Reference = f["reference"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrule %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("notification: notificationrule %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationrule %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationrule %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationRule) String() string {
	return fmt.Sprintf("NotificationRule(%s, group=%q, offset=%d, archived=%t)", x.ID, x.Group, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationRule) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationRule) Normalize() NotificationRule {
	x.Group = strings.TrimSpace(x.Group)
	x.Name = strings.TrimSpace(x.Name)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortNotificationRules orders by offset descending, then by ID.
func SortNotificationRules(xs []NotificationRule) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationRules drops the archived entries, preserving order.
func ActiveNotificationRules(xs []NotificationRule) []NotificationRule {
	out := make([]NotificationRule, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationRules adds up the offset of every active entry.
func SumNotificationRules(xs []NotificationRule) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexNotificationRules keys the entries by ID, last write winning.
func IndexNotificationRules(xs []NotificationRule) map[string]NotificationRule {
	out := make(map[string]NotificationRule, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationRule builds a valid NotificationRule, for tests and for the seed data the cron
// binary writes.
func SampleNotificationRule(id string, at time.Time) NotificationRule {
	return NotificationRule{
		ID:        id,
		Group:     "group-" + id,
		Name:      "name-" + id,
		Reference: "seeded by notification.SampleNotificationRule",
		Offset:    len(id),
		Priority:  len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// NotificationHoldSchema is the stored shape of a NotificationHold.
var NotificationHoldSchema = schema.New("notification.notificationhold",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "region", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "category", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationHold is one notification hold.
type NotificationHold struct {
	ID       string
	Region   string
	Category string
	Channel  string
	Offset   int
	Version  int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationHold) Key() string { return "notification/notificationhold/" + x.ID }

// Fields encodes x for storage.
func (x NotificationHold) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"region":   x.Region,
		"category": x.Category,
		"channel":  x.Channel,
		"offset":   strconv.Itoa(x.Offset),
		"version":  strconv.Itoa(x.Version),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationHold decodes a stored record.
func ParseNotificationHold(f codec.Fields) (NotificationHold, error) {
	var x NotificationHold
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationhold: missing id")
	}
	x.Region = f["region"]
	x.Category = f["category"]
	x.Channel = f["channel"]
	var err error
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationhold %s: %w", x.ID, err)
	}
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationhold %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationhold %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationhold %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationHold) String() string {
	return fmt.Sprintf("NotificationHold(%s, region=%q, offset=%d, archived=%t)", x.ID, x.Region, x.Offset, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationHold) Active() bool { return !x.Archived && x.Offset >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationHold) Normalize() NotificationHold {
	x.Region = strings.TrimSpace(x.Region)
	x.Category = strings.TrimSpace(x.Category)
	x.Channel = strings.TrimSpace(x.Channel)
	if x.Offset < 0 {
		x.Offset = 0
	}
	if x.Version < 0 {
		x.Version = 0
	}
	return x
}

// SortNotificationHolds orders by offset descending, then by ID.
func SortNotificationHolds(xs []NotificationHold) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Offset != xs[j].Offset {
			return xs[i].Offset > xs[j].Offset
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationHolds drops the archived entries, preserving order.
func ActiveNotificationHolds(xs []NotificationHold) []NotificationHold {
	out := make([]NotificationHold, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationHolds adds up the offset of every active entry.
func SumNotificationHolds(xs []NotificationHold) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Offset
		}
	}
	return total
}

// IndexNotificationHolds keys the entries by ID, last write winning.
func IndexNotificationHolds(xs []NotificationHold) map[string]NotificationHold {
	out := make(map[string]NotificationHold, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationHold builds a valid NotificationHold, for tests and for the seed data the cron
// binary writes.
func SampleNotificationHold(id string, at time.Time) NotificationHold {
	return NotificationHold{
		ID:       id,
		Region:   "region-" + id,
		Category: "category-" + id,
		Channel:  "seeded by notification.SampleNotificationHold",
		Offset:   len(id),
		Version:  len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationClaimSchema is the stored shape of a NotificationClaim.
var NotificationClaimSchema = schema.New("notification.notificationclaim",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "name", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "channel", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "comment", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "capacity", Kind: schema.KindInt},
	schema.Field{Name: "priority", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationClaim is one notification claim.
type NotificationClaim struct {
	ID       string
	Name     string
	Channel  string
	Comment  string
	Capacity int
	Priority int
	Archived bool
	Updated  time.Time
}

// Key returns the store key for x.
func (x NotificationClaim) Key() string { return "notification/notificationclaim/" + x.ID }

// Fields encodes x for storage.
func (x NotificationClaim) Fields() codec.Fields {
	return codec.Fields{
		"id":       x.ID,
		"name":     x.Name,
		"channel":  x.Channel,
		"comment":  x.Comment,
		"capacity": strconv.Itoa(x.Capacity),
		"priority": strconv.Itoa(x.Priority),
		"archived": strconv.FormatBool(x.Archived),
		"updated":  x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationClaim decodes a stored record.
func ParseNotificationClaim(f codec.Fields) (NotificationClaim, error) {
	var x NotificationClaim
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationclaim: missing id")
	}
	x.Name = f["name"]
	x.Channel = f["channel"]
	x.Comment = f["comment"]
	var err error
	if x.Capacity, err = f.Int("capacity", 0); err != nil {
		return x, fmt.Errorf("notification: notificationclaim %s: %w", x.ID, err)
	}
	if x.Priority, err = f.Int("priority", 0); err != nil {
		return x, fmt.Errorf("notification: notificationclaim %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationclaim %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationclaim %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationClaim) String() string {
	return fmt.Sprintf("NotificationClaim(%s, name=%q, capacity=%d, archived=%t)", x.ID, x.Name, x.Capacity, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationClaim) Active() bool { return !x.Archived && x.Capacity >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationClaim) Normalize() NotificationClaim {
	x.Name = strings.TrimSpace(x.Name)
	x.Channel = strings.TrimSpace(x.Channel)
	x.Comment = strings.TrimSpace(x.Comment)
	if x.Capacity < 0 {
		x.Capacity = 0
	}
	if x.Priority < 0 {
		x.Priority = 0
	}
	return x
}

// SortNotificationClaims orders by capacity descending, then by ID.
func SortNotificationClaims(xs []NotificationClaim) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Capacity != xs[j].Capacity {
			return xs[i].Capacity > xs[j].Capacity
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationClaims drops the archived entries, preserving order.
func ActiveNotificationClaims(xs []NotificationClaim) []NotificationClaim {
	out := make([]NotificationClaim, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationClaims adds up the capacity of every active entry.
func SumNotificationClaims(xs []NotificationClaim) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Capacity
		}
	}
	return total
}

// IndexNotificationClaims keys the entries by ID, last write winning.
func IndexNotificationClaims(xs []NotificationClaim) map[string]NotificationClaim {
	out := make(map[string]NotificationClaim, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationClaim builds a valid NotificationClaim, for tests and for the seed data the cron
// binary writes.
func SampleNotificationClaim(id string, at time.Time) NotificationClaim {
	return NotificationClaim{
		ID:       id,
		Name:     "name-" + id,
		Channel:  "channel-" + id,
		Comment:  "seeded by notification.SampleNotificationClaim",
		Capacity: len(id),
		Priority: len(id) * 2,
		Updated:  at.UTC().Truncate(time.Second),
	}
}

// NotificationCursorSchema is the stored shape of a NotificationCursor.
var NotificationCursorSchema = schema.New("notification.notificationcursor",
	schema.Field{Name: "id", Kind: schema.KindString, Required: true, MaxLen: 64},
	schema.Field{Name: "source", Kind: schema.KindString, Required: true, MaxLen: 128},
	schema.Field{Name: "label", Kind: schema.KindString, MaxLen: 128},
	schema.Field{Name: "reference", Kind: schema.KindString, MaxLen: 256},
	schema.Field{Name: "version", Kind: schema.KindInt},
	schema.Field{Name: "offset", Kind: schema.KindInt},
	schema.Field{Name: "archived", Kind: schema.KindBool},
	schema.Field{Name: "updated", Kind: schema.KindTimestamp},
)

// NotificationCursor is one notification cursor.
type NotificationCursor struct {
	ID        string
	Source    string
	Label     string
	Reference string
	Version   int
	Offset    int
	Archived  bool
	Updated   time.Time
}

// Key returns the store key for x.
func (x NotificationCursor) Key() string { return "notification/notificationcursor/" + x.ID }

// Fields encodes x for storage.
func (x NotificationCursor) Fields() codec.Fields {
	return codec.Fields{
		"id":        x.ID,
		"source":    x.Source,
		"label":     x.Label,
		"reference": x.Reference,
		"version":   strconv.Itoa(x.Version),
		"offset":    strconv.Itoa(x.Offset),
		"archived":  strconv.FormatBool(x.Archived),
		"updated":   x.Updated.UTC().Format(time.RFC3339),
	}
}

// ParseNotificationCursor decodes a stored record.
func ParseNotificationCursor(f codec.Fields) (NotificationCursor, error) {
	var x NotificationCursor
	x.ID = f["id"]
	if x.ID == "" {
		return x, fmt.Errorf("notification: notificationcursor: missing id")
	}
	x.Source = f["source"]
	x.Label = f["label"]
	x.Reference = f["reference"]
	var err error
	if x.Version, err = f.Int("version", 0); err != nil {
		return x, fmt.Errorf("notification: notificationcursor %s: %w", x.ID, err)
	}
	if x.Offset, err = f.Int("offset", 0); err != nil {
		return x, fmt.Errorf("notification: notificationcursor %s: %w", x.ID, err)
	}
	if x.Archived, err = f.Bool("archived", false); err != nil {
		return x, fmt.Errorf("notification: notificationcursor %s: %w", x.ID, err)
	}
	if raw := f["updated"]; raw != "" {
		if x.Updated, err = time.Parse(time.RFC3339, raw); err != nil {
			return x, fmt.Errorf("notification: notificationcursor %s: updated: %w", x.ID, err)
		}
	}
	return x, nil
}

// String implements fmt.Stringer.
func (x NotificationCursor) String() string {
	return fmt.Sprintf("NotificationCursor(%s, source=%q, version=%d, archived=%t)", x.ID, x.Source, x.Version, x.Archived)
}

// Active reports whether the service layer should consider x.
func (x NotificationCursor) Active() bool { return !x.Archived && x.Version >= 0 }

// Normalize trims the free-text fields and clamps the counters.
func (x NotificationCursor) Normalize() NotificationCursor {
	x.Source = strings.TrimSpace(x.Source)
	x.Label = strings.TrimSpace(x.Label)
	x.Reference = strings.TrimSpace(x.Reference)
	if x.Version < 0 {
		x.Version = 0
	}
	if x.Offset < 0 {
		x.Offset = 0
	}
	return x
}

// SortNotificationCursors orders by version descending, then by ID.
func SortNotificationCursors(xs []NotificationCursor) {
	sort.Slice(xs, func(i, j int) bool {
		if xs[i].Version != xs[j].Version {
			return xs[i].Version > xs[j].Version
		}
		return xs[i].ID < xs[j].ID
	})
}

// ActiveNotificationCursors drops the archived entries, preserving order.
func ActiveNotificationCursors(xs []NotificationCursor) []NotificationCursor {
	out := make([]NotificationCursor, 0, len(xs))
	for _, x := range xs {
		if x.Active() {
			out = append(out, x)
		}
	}
	return out
}

// SumNotificationCursors adds up the version of every active entry.
func SumNotificationCursors(xs []NotificationCursor) int {
	total := 0
	for _, x := range xs {
		if x.Active() {
			total += x.Version
		}
	}
	return total
}

// IndexNotificationCursors keys the entries by ID, last write winning.
func IndexNotificationCursors(xs []NotificationCursor) map[string]NotificationCursor {
	out := make(map[string]NotificationCursor, len(xs))
	for _, x := range xs {
		out[x.ID] = x
	}
	return out
}

// SampleNotificationCursor builds a valid NotificationCursor, for tests and for the seed data the cron
// binary writes.
func SampleNotificationCursor(id string, at time.Time) NotificationCursor {
	return NotificationCursor{
		ID:        id,
		Source:    "source-" + id,
		Label:     "label-" + id,
		Reference: "seeded by notification.SampleNotificationCursor",
		Version:   len(id),
		Offset:    len(id) * 2,
		Updated:   at.UTC().Truncate(time.Second),
	}
}

// Schemas lists every schema this package declares, in declaration order.
var Schemas = []schema.Schema{
	NotificationRecordSchema,
	NotificationEntrySchema,
	NotificationItemSchema,
	NotificationDraftSchema,
	NotificationRevisionSchema,
	NotificationSnapshotSchema,
	NotificationBatchSchema,
	NotificationRequestSchema,
	NotificationSummarySchema,
	NotificationPolicySchema,
	NotificationLinkSchema,
	NotificationNoteSchema,
	NotificationWindowSchema,
	NotificationQuotaSchema,
	NotificationBucketSchema,
	NotificationLaneSchema,
	NotificationSlotSchema,
	NotificationTagSchema,
	NotificationRuleSchema,
	NotificationHoldSchema,
	NotificationClaimSchema,
	NotificationCursorSchema,
}

// Prefixes lists the store key prefix for each schema, positionally matching
// Schemas.
var Prefixes = []string{
	"notification/notificationrecord/",
	"notification/notificationentry/",
	"notification/notificationitem/",
	"notification/notificationdraft/",
	"notification/notificationrevision/",
	"notification/notificationsnapshot/",
	"notification/notificationbatch/",
	"notification/notificationrequest/",
	"notification/notificationsummary/",
	"notification/notificationpolicy/",
	"notification/notificationlink/",
	"notification/notificationnote/",
	"notification/notificationwindow/",
	"notification/notificationquota/",
	"notification/notificationbucket/",
	"notification/notificationlane/",
	"notification/notificationslot/",
	"notification/notificationtag/",
	"notification/notificationrule/",
	"notification/notificationhold/",
	"notification/notificationclaim/",
	"notification/notificationcursor/",
}

// Summary renders a one-line description of what this package manages.
func Summary() string {
	names := make([]string, 0, len(Schemas))
	for _, s := range Schemas {
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return "notification: " + strings.Join(names, ", ")
}
